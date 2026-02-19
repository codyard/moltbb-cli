package localweb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"moltbb-cli/internal/utils"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS prompts (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
  builtin INTEGER NOT NULL DEFAULT 0 CHECK (builtin IN (0,1)),
  active INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0,1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS diary_entries (
  id TEXT PRIMARY KEY,
  rel_path TEXT NOT NULL UNIQUE,
  filename TEXT NOT NULL,
  date TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  preview TEXT NOT NULL,
  size INTEGER NOT NULL,
  modified_at TEXT NOT NULL,
  indexed_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_diary_entries_date ON diary_entries(date);
CREATE INDEX IF NOT EXISTS idx_diary_entries_modified_at ON diary_entries(modified_at);
`

var promptIDRe = regexp.MustCompile(`[^a-z0-9-]+`)

type Prompt struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Enabled     bool   `json:"enabled"`
	Builtin     bool   `json:"builtin"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type PromptPatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Content     *string `json:"content,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type PromptMeta struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
	Builtin       bool   `json:"builtin"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	ContentLength int    `json:"contentLength"`
	Active        bool   `json:"active"`
}

type PromptStore struct {
	mu         sync.Mutex
	db         *sql.DB
	legacyPath string
}

type legacyPromptCatalog struct {
	ActivePromptID string   `json:"activePromptId"`
	Prompts        []Prompt `json:"prompts"`
}

func OpenDB(dbPath string) (*sql.DB, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, errors.New("db path is required")
	}
	if err := utils.EnsureDir(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set sqlite pragma journal_mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set sqlite pragma busy_timeout: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize sqlite schema: %w", err)
	}

	return db, nil
}

func NewPromptStore(db *sql.DB, legacyPath string, defaultContent string) (*PromptStore, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	store := &PromptStore{db: db, legacyPath: legacyPath}
	if err := store.bootstrap(defaultContent); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PromptStore) bootstrap(defaultContent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	count, err := s.countPromptsLocked()
	if err != nil {
		return err
	}

	if count == 0 {
		migrated, err := s.migrateLegacyLocked()
		if err != nil {
			return err
		}
		if !migrated {
			if err := s.insertDefaultPromptLocked(strings.TrimSpace(defaultContent)); err != nil {
				return err
			}
		}
	}

	return s.ensurePromptConsistencyLocked()
}

func (s *PromptStore) List() []PromptMeta {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
SELECT id, name, description, content, enabled, builtin, active, created_at, updated_at
FROM prompts
`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	metas := make([]PromptMeta, 0, 8)
	for rows.Next() {
		var (
			p       Prompt
			enabled int
			builtin int
			active  int
		)
		if scanErr := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Content, &enabled, &builtin, &active, &p.CreatedAt, &p.UpdatedAt); scanErr != nil {
			continue
		}
		metas = append(metas, PromptMeta{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			Enabled:       enabled == 1,
			Builtin:       builtin == 1,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
			ContentLength: len(p.Content),
			Active:        active == 1,
		})
	}

	sort.Slice(metas, func(i, j int) bool {
		if metas[i].Active != metas[j].Active {
			return metas[i].Active
		}
		return metas[i].UpdatedAt > metas[j].UpdatedAt
	})

	return metas
}

func (s *PromptStore) Get(id string) (Prompt, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt, active, ok := s.getPromptByIDLocked(id)
	if !ok {
		return Prompt{}, false
	}
	_ = active
	return prompt, true
}

func (s *PromptStore) GetActive() (Prompt, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt, _, ok := s.getActivePromptLocked()
	if !ok {
		return Prompt{}, false
	}
	return prompt, true
}

func (s *PromptStore) Create(input Prompt) (Prompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Prompt{}, errors.New("name is required")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return Prompt{}, errors.New("content is required")
	}

	id := normalizePromptID(input.ID)
	if id == "" {
		id = normalizePromptID(name)
	}
	if id == "" {
		id = "prompt"
	}
	id, err := s.ensureUniqueIDLocked(id)
	if err != nil {
		return Prompt{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	enabled := input.Enabled
	active := false

	if _, _, hasActive := s.getActivePromptLocked(); !hasActive {
		enabled = true
		active = true
	}

	_, err = s.db.Exec(`
INSERT INTO prompts(id, name, description, content, enabled, builtin, active, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, 0, ?, ?, ?)
`, id, name, strings.TrimSpace(input.Description), content, boolToInt(enabled), boolToInt(active), now, now)
	if err != nil {
		return Prompt{}, fmt.Errorf("create prompt: %w", err)
	}

	if err := s.ensurePromptConsistencyLocked(); err != nil {
		return Prompt{}, err
	}
	prompt, _, ok := s.getPromptByIDLocked(id)
	if !ok {
		return Prompt{}, errors.New("created prompt not found")
	}
	return prompt, nil
}

func (s *PromptStore) Patch(id string, patch PromptPatch) (Prompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt, active, ok := s.getPromptByIDLocked(id)
	if !ok {
		return Prompt{}, os.ErrNotExist
	}

	if patch.Name != nil {
		name := strings.TrimSpace(*patch.Name)
		if name == "" {
			return Prompt{}, errors.New("name cannot be empty")
		}
		prompt.Name = name
	}
	if patch.Description != nil {
		prompt.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.Content != nil {
		content := strings.TrimSpace(*patch.Content)
		if content == "" {
			return Prompt{}, errors.New("content cannot be empty")
		}
		prompt.Content = content
	}
	if patch.Enabled != nil {
		prompt.Enabled = *patch.Enabled
	}

	if !prompt.Enabled {
		enabledCount, err := s.enabledPromptCountLocked(id)
		if err != nil {
			return Prompt{}, err
		}
		if enabledCount == 0 {
			return Prompt{}, errors.New("at least one prompt must remain enabled")
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
UPDATE prompts
SET name = ?, description = ?, content = ?, enabled = ?, updated_at = ?
WHERE id = ?
`, prompt.Name, prompt.Description, prompt.Content, boolToInt(prompt.Enabled), now, id)
	if err != nil {
		return Prompt{}, fmt.Errorf("update prompt: %w", err)
	}

	if active && !prompt.Enabled {
		if err := s.promoteAnyEnabledPromptLocked(id); err != nil {
			return Prompt{}, err
		}
	}
	if err := s.ensurePromptConsistencyLocked(); err != nil {
		return Prompt{}, err
	}

	updated, _, ok := s.getPromptByIDLocked(id)
	if !ok {
		return Prompt{}, os.ErrNotExist
	}
	return updated, nil
}

func (s *PromptStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt, active, ok := s.getPromptByIDLocked(id)
	if !ok {
		return os.ErrNotExist
	}
	if prompt.Builtin {
		return errors.New("builtin prompt cannot be deleted")
	}

	count, err := s.countPromptsLocked()
	if err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("at least one prompt must exist")
	}

	_, err = s.db.Exec(`DELETE FROM prompts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete prompt: %w", err)
	}

	if active {
		if err := s.promoteAnyEnabledPromptLocked(""); err != nil {
			return err
		}
	}
	return s.ensurePromptConsistencyLocked()
}

func (s *PromptStore) Activate(id string) (Prompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt, _, ok := s.getPromptByIDLocked(id)
	if !ok {
		return Prompt{}, os.ErrNotExist
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.Begin()
	if err != nil {
		return Prompt{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE prompts SET active = 0`); err != nil {
		return Prompt{}, fmt.Errorf("reset active prompt: %w", err)
	}
	if _, err := tx.Exec(`
UPDATE prompts
SET active = 1, enabled = 1, updated_at = ?
WHERE id = ?
`, now, id); err != nil {
		return Prompt{}, fmt.Errorf("activate prompt: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Prompt{}, fmt.Errorf("commit tx: %w", err)
	}

	prompt.Enabled = true
	prompt.UpdatedAt = now
	return prompt, nil
}

func (s *PromptStore) ActivePromptID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, id, ok := s.getActivePromptLocked()
	if !ok {
		return ""
	}
	return id
}

func (s *PromptStore) countPromptsLocked() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(1) FROM prompts`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count prompts: %w", err)
	}
	return count, nil
}

func (s *PromptStore) enabledPromptCountLocked(excludeID string) (int, error) {
	if strings.TrimSpace(excludeID) == "" {
		row := s.db.QueryRow(`SELECT COUNT(1) FROM prompts WHERE enabled = 1`)
		var count int
		if err := row.Scan(&count); err != nil {
			return 0, fmt.Errorf("count enabled prompts: %w", err)
		}
		return count, nil
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM prompts WHERE enabled = 1 AND id <> ?`, excludeID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count enabled prompts: %w", err)
	}
	return count, nil
}

func (s *PromptStore) getPromptByIDLocked(id string) (Prompt, bool, bool) {
	row := s.db.QueryRow(`
SELECT id, name, description, content, enabled, builtin, active, created_at, updated_at
FROM prompts
WHERE id = ?
`, id)

	var (
		prompt  Prompt
		enabled int
		builtin int
		active  int
	)
	if err := row.Scan(&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Content, &enabled, &builtin, &active, &prompt.CreatedAt, &prompt.UpdatedAt); err != nil {
		return Prompt{}, false, false
	}
	prompt.Enabled = enabled == 1
	prompt.Builtin = builtin == 1
	return prompt, active == 1, true
}

func (s *PromptStore) getActivePromptLocked() (Prompt, string, bool) {
	row := s.db.QueryRow(`
SELECT id, name, description, content, enabled, builtin, active, created_at, updated_at
FROM prompts
WHERE active = 1
LIMIT 1
`)

	var (
		prompt  Prompt
		enabled int
		builtin int
		active  int
	)
	if err := row.Scan(&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Content, &enabled, &builtin, &active, &prompt.CreatedAt, &prompt.UpdatedAt); err != nil {
		return Prompt{}, "", false
	}
	prompt.Enabled = enabled == 1
	prompt.Builtin = builtin == 1
	return prompt, prompt.ID, true
}

func (s *PromptStore) ensureUniqueIDLocked(base string) (string, error) {
	candidate := base
	suffix := 2
	for {
		row := s.db.QueryRow(`SELECT COUNT(1) FROM prompts WHERE id = ?`, candidate)
		var count int
		if err := row.Scan(&count); err != nil {
			return "", fmt.Errorf("check prompt id uniqueness: %w", err)
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, suffix)
		suffix++
	}
}

func (s *PromptStore) promoteAnyEnabledPromptLocked(skipID string) error {
	query := `SELECT id FROM prompts WHERE enabled = 1`
	args := []any{}
	if strings.TrimSpace(skipID) != "" {
		query += ` AND id <> ?`
		args = append(args, skipID)
	}
	query += ` ORDER BY updated_at DESC LIMIT 1`

	row := s.db.QueryRow(query, args...)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no enabled prompt available")
		}
		return fmt.Errorf("find replacement active prompt: %w", err)
	}

	_, err := s.db.Exec(`UPDATE prompts SET active = CASE WHEN id = ? THEN 1 ELSE 0 END`, id)
	if err != nil {
		return fmt.Errorf("promote active prompt: %w", err)
	}
	return nil
}

func (s *PromptStore) insertDefaultPromptLocked(content string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(content) == "" {
		content = "[TODAY_STRUCTURED_SUMMARY]\n[OPTIONAL: RECENT MEMORY EXCERPT]\n[ROLE_DEFINITION]"
	}
	_, err := s.db.Exec(`
INSERT INTO prompts(id, name, description, content, enabled, builtin, active, created_at, updated_at)
VALUES('default', 'Default Diary Prompt', 'Bundled prompt template for diary generation.', ?, 1, 1, 1, ?, ?)
`, strings.TrimSpace(content), now, now)
	if err != nil {
		return fmt.Errorf("insert default prompt: %w", err)
	}
	return nil
}

func (s *PromptStore) ensurePromptConsistencyLocked() error {
	count, err := s.countPromptsLocked()
	if err != nil {
		return err
	}
	if count == 0 {
		return s.insertDefaultPromptLocked("")
	}

	if _, _, ok := s.getActivePromptLocked(); ok {
		return nil
	}

	row := s.db.QueryRow(`SELECT id FROM prompts WHERE enabled = 1 ORDER BY updated_at DESC LIMIT 1`)
	var id string
	if err := row.Scan(&id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("select active candidate: %w", err)
		}
		if _, err := s.db.Exec(`UPDATE prompts SET enabled = 1 WHERE id = (SELECT id FROM prompts ORDER BY updated_at DESC LIMIT 1)`); err != nil {
			return fmt.Errorf("enable fallback prompt: %w", err)
		}
		row = s.db.QueryRow(`SELECT id FROM prompts WHERE enabled = 1 ORDER BY updated_at DESC LIMIT 1`)
		if err := row.Scan(&id); err != nil {
			return fmt.Errorf("select fallback prompt: %w", err)
		}
	}

	if _, err := s.db.Exec(`UPDATE prompts SET active = CASE WHEN id = ? THEN 1 ELSE 0 END`, id); err != nil {
		return fmt.Errorf("set active prompt: %w", err)
	}
	return nil
}

func (s *PromptStore) migrateLegacyLocked() (bool, error) {
	legacyPath := strings.TrimSpace(s.legacyPath)
	if legacyPath == "" {
		return false, nil
	}
	if _, err := os.Stat(legacyPath); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("stat legacy prompts file: %w", err)
	}

	data, err := os.ReadFile(legacyPath)
	if err != nil {
		return false, fmt.Errorf("read legacy prompts file: %w", err)
	}

	var legacy legacyPromptCatalog
	if err := json.Unmarshal(data, &legacy); err != nil {
		return false, fmt.Errorf("parse legacy prompts file: %w", err)
	}
	if len(legacy.Prompts) == 0 {
		return false, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

	seenIDs := make(map[string]int)
	activeID := strings.TrimSpace(legacy.ActivePromptID)
	now := time.Now().UTC().Format(time.RFC3339)

	for i, prompt := range legacy.Prompts {
		id := normalizePromptID(prompt.ID)
		if id == "" {
			id = normalizePromptID(prompt.Name)
		}
		if id == "" {
			id = fmt.Sprintf("prompt-%d", i+1)
		}
		if n, exists := seenIDs[id]; exists {
			n++
			seenIDs[id] = n
			id = fmt.Sprintf("%s-%d", id, n)
		} else {
			seenIDs[id] = 1
		}

		name := strings.TrimSpace(prompt.Name)
		if name == "" {
			name = id
		}
		content := strings.TrimSpace(prompt.Content)
		if content == "" {
			continue
		}
		createdAt := normalizeTimestamp(prompt.CreatedAt, now)
		updatedAt := normalizeTimestamp(prompt.UpdatedAt, now)
		isActive := activeID != "" && prompt.ID == activeID

		_, err := tx.Exec(`
INSERT INTO prompts(id, name, description, content, enabled, builtin, active, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
			id,
			name,
			strings.TrimSpace(prompt.Description),
			content,
			boolToInt(prompt.Enabled),
			boolToInt(prompt.Builtin),
			boolToInt(isActive),
			createdAt,
			updatedAt,
		)
		if err != nil {
			return false, fmt.Errorf("insert migrated prompt: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit migration tx: %w", err)
	}

	archivePath := legacyPath + ".migrated-" + time.Now().UTC().Format("20060102150405")
	if renameErr := os.Rename(legacyPath, archivePath); renameErr != nil {
		// keep non-fatal: data is already migrated into sqlite.
	}

	return true, nil
}

func normalizePromptID(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = promptIDRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 48 {
		s = s[:48]
		s = strings.Trim(s, "-")
	}
	return s
}

func normalizeTimestamp(input string, fallback string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return fallback
	}
	if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
		return fallback
	}
	return trimmed
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

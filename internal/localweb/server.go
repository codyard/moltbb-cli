package localweb

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/diary"
	"moltbb-cli/internal/utils"
)

//go:embed static/*
var staticFS embed.FS

var diaryDateRe = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
var diaryLabelDateRe = regexp.MustCompile(`(?im)^\s*(?:[-*]\s*)?(?:date|日期)\s*:\s*(\d{4}-\d{2}-\d{2})\s*$`)

const settingKeyCloudSyncEnabled = "cloud_sync_enabled"

type Options struct {
	DiaryDir   string
	DataDir    string
	APIBaseURL string
	InputPaths []string
}

type Server struct {
	diaryDir   string
	dataDir    string
	dbPath     string
	apiBaseURL string
	inputPaths []string
	db         *sql.DB
	prompts    *PromptStore
	mux        *http.ServeMux
}

type diarySummary struct {
	ID         string `json:"id"`
	Date       string `json:"date,omitempty"`
	Title      string `json:"title"`
	Preview    string `json:"preview"`
	Filename   string `json:"filename"`
	RelPath    string `json:"relPath"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
}

type diaryDetail struct {
	diarySummary
	Content string `json:"content"`
}

type diariesResponse struct {
	Items  []diarySummary `json:"items"`
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

type stateResponse struct {
	DiaryDir      string `json:"diaryDir"`
	DataDir       string `json:"dataDir"`
	DatabasePath  string `json:"databasePath"`
	PromptCount   int    `json:"promptCount"`
	ActivePrompt  string `json:"activePrompt"`
	DiaryCount    int    `json:"diaryCount"`
	APIBaseURL    string `json:"apiBaseUrl"`
	DefaultOutput string `json:"defaultOutput"`
}

type settingsResponse struct {
	CloudSyncEnabled bool   `json:"cloudSyncEnabled"`
	APIKeyConfigured bool   `json:"apiKeyConfigured"`
	APIKeyMasked     string `json:"apiKeyMasked,omitempty"`
	APIKeySource     string `json:"apiKeySource,omitempty"`
}

type settingsUpdateRequest struct {
	CloudSyncEnabled *bool   `json:"cloudSyncEnabled,omitempty"`
	APIKey           *string `json:"apiKey,omitempty"`
}

type generatePacketRequest struct {
	Date           string   `json:"date"`
	Hostname       string   `json:"hostname"`
	PromptID       string   `json:"promptId"`
	OutputDir      string   `json:"outputDir"`
	LogSourceHints []string `json:"logSourceHints"`
}

type generatePacketResponse struct {
	Success    bool     `json:"success"`
	PacketPath string   `json:"packetPath"`
	Date       string   `json:"date"`
	Hostname   string   `json:"hostname"`
	PromptID   string   `json:"promptId"`
	Hints      []string `json:"hints"`
	Summary    string   `json:"summary"`
}

type promptCreateRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func New(options Options) (*Server, error) {
	diaryDir := strings.TrimSpace(options.DiaryDir)
	if diaryDir == "" {
		return nil, errors.New("diary dir is required")
	}

	expandedDiaryDir, err := utils.ExpandPath(diaryDir)
	if err != nil {
		return nil, fmt.Errorf("expand diary dir: %w", err)
	}
	if err := utils.EnsureDir(expandedDiaryDir, 0o700); err != nil {
		return nil, err
	}

	dataDir := strings.TrimSpace(options.DataDir)
	if dataDir == "" {
		moltbbDir, err := utils.MoltbbDir()
		if err != nil {
			return nil, err
		}
		dataDir = filepath.Join(moltbbDir, "local-web")
	}
	expandedDataDir, err := utils.ExpandPath(dataDir)
	if err != nil {
		return nil, fmt.Errorf("expand data dir: %w", err)
	}
	if err := utils.EnsureDir(expandedDataDir, 0o700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(expandedDataDir, "local.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		return nil, err
	}

	defaultPrompt := loadDefaultPromptTemplate()
	promptStore, err := NewPromptStore(db, filepath.Join(expandedDataDir, "prompts.json"), defaultPrompt)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &Server{
		diaryDir:   expandedDiaryDir,
		dataDir:    expandedDataDir,
		dbPath:     dbPath,
		apiBaseURL: strings.TrimSpace(options.APIBaseURL),
		inputPaths: filterNonEmpty(options.InputPaths),
		db:         db,
		prompts:    promptStore,
		mux:        http.NewServeMux(),
	}
	if _, err := s.reindexDiaries(); err != nil {
		_ = db.Close()
		return nil, err
	}
	s.registerRoutes()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rewrittenPath, ok := rewritePrefixedPath(r.URL.Path); ok {
		cloned := r.Clone(r.Context())
		clonedURL := *r.URL
		clonedURL.Path = rewrittenPath
		cloned.URL = &clonedURL
		s.mux.ServeHTTP(w, cloned)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/state", s.handleState)
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/diaries", s.handleDiaries)
	s.mux.HandleFunc("/api/diaries/reindex", s.handleReindex)
	s.mux.HandleFunc("/api/diaries/", s.handleDiaryByID)
	s.mux.HandleFunc("/api/prompts", s.handlePrompts)
	s.mux.HandleFunc("/api/prompts/", s.handlePromptByID)
	s.mux.HandleFunc("/api/generate-packet", s.handleGeneratePacket)

	assetFS, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(assetFS))
	indexHTML, indexErr := fs.ReadFile(assetFS, "index.html")
	serveIndex := func(w http.ResponseWriter) {
		if indexErr != nil {
			http.Error(w, "index.html missing", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexHTML)
	}

	s.mux.Handle("/styles.css", fileServer)
	s.mux.Handle("/app.js", fileServer)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			serveIndex(w)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		serveIndex(w)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	diaryCount, err := s.countDiaries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	prompts := s.prompts.List()

	writeJSON(w, http.StatusOK, stateResponse{
		DiaryDir:      s.diaryDir,
		DataDir:       s.dataDir,
		DatabasePath:  s.dbPath,
		PromptCount:   len(prompts),
		ActivePrompt:  s.prompts.ActivePromptID(),
		DiaryCount:    diaryCount,
		APIBaseURL:    s.apiBaseURL,
		DefaultOutput: s.diaryDir,
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.readSettings()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	case http.MethodPatch:
		var req settingsUpdateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if req.CloudSyncEnabled == nil && req.APIKey == nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "at least one setting is required"})
			return
		}

		if req.CloudSyncEnabled != nil {
			if err := s.setCloudSyncEnabled(*req.CloudSyncEnabled); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		}
		if req.APIKey != nil {
			apiKey := strings.TrimSpace(*req.APIKey)
			if apiKey == "" {
				if err := auth.Clear(); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			} else {
				token := ""
				if existing, err := auth.Load(); err == nil {
					token = existing.Token
				}
				if err := auth.Save(apiKey, token); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			}
		}

		settings, err := s.readSettings()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (s *Server) handleReindex(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	count, err := s.reindexDiaries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"diaryCount":  count,
		"reindexedAt": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleDiaries(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	limit := parseInt(r.URL.Query().Get("limit"), 50, 1, 500)
	offset := parseInt(r.URL.Query().Get("offset"), 0, 0, 1_000_000)

	items, total, err := s.listDiaries(q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, diariesResponse{
		Items:  items,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	})
}

func (s *Server) handleDiaryByID(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/diaries/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "diary id is required"})
		return
	}
	decoded, err := url.PathUnescape(id)
	if err == nil {
		id = decoded
	}

	detail, found, err := s.loadDiaryDetail(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "diary not found"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handlePrompts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"items":          s.prompts.List(),
			"activePromptId": s.prompts.ActivePromptID(),
		})
	case http.MethodPost:
		var req promptCreateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		enabled := true
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		created, err := s.prompts.Create(Prompt{
			ID:          req.ID,
			Name:        req.Name,
			Description: req.Description,
			Content:     req.Content,
			Enabled:     enabled,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (s *Server) handlePromptByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/prompts/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "prompt id is required"})
		return
	}

	parts := strings.Split(path, "/")
	id := parts[0]
	decoded, err := url.PathUnescape(id)
	if err == nil {
		id = decoded
	}
	if strings.TrimSpace(id) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "prompt id is required"})
		return
	}

	if len(parts) == 2 && parts[1] == "activate" {
		if !allowMethod(w, r, http.MethodPost) {
			return
		}
		prompt, err := s.prompts.Activate(id)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, os.ErrNotExist) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, prompt)
		return
	}

	switch r.Method {
	case http.MethodGet:
		prompt, ok := s.prompts.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "prompt not found"})
			return
		}
		writeJSON(w, http.StatusOK, prompt)
	case http.MethodPatch:
		var patch PromptPatch
		if err := decodeJSON(r, &patch); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		updated, err := s.prompts.Patch(id, patch)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, os.ErrNotExist) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		err := s.prompts.Delete(id)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, os.ErrNotExist) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (s *Server) handleGeneratePacket(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	var req generatePacketRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	date := strings.TrimSpace(req.Date)
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "date must use YYYY-MM-DD"})
		return
	}

	hostname := strings.TrimSpace(req.Hostname)
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	if hostname == "" {
		hostname = "local-agent"
	}

	outputDir := strings.TrimSpace(req.OutputDir)
	if outputDir == "" {
		outputDir = s.diaryDir
	}
	expandedOutputDir, err := utils.ExpandPath(outputDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if err := utils.EnsureDir(expandedOutputDir, 0o700); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	promptID := strings.TrimSpace(req.PromptID)
	var prompt Prompt
	var ok bool
	if promptID == "" {
		prompt, ok = s.prompts.GetActive()
	} else {
		prompt, ok = s.prompts.Get(promptID)
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "prompt not found"})
		return
	}

	hints := filterNonEmpty(req.LogSourceHints)
	if len(hints) == 0 {
		hints = append([]string{}, s.inputPaths...)
	}

	tmpFile, err := os.CreateTemp(s.dataDir, "prompt-template-*.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := os.WriteFile(tmpPath, []byte(prompt.Content), 0o600); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	packetPath, err := diary.WritePromptPacket(date, hostname, s.apiBaseURL, expandedOutputDir, tmpPath, hints)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, generatePacketResponse{
		Success:    true,
		PacketPath: packetPath,
		Date:       date,
		Hostname:   hostname,
		PromptID:   prompt.ID,
		Hints:      hints,
		Summary:    diary.AgentManagedSummary(len(hints)),
	})
}

func (s *Server) readSettings() (settingsResponse, error) {
	cloudSyncEnabled, err := s.getCloudSyncEnabled()
	if err != nil {
		return settingsResponse{}, err
	}

	configured, masked, source, err := s.resolveAPIKeyState()
	if err != nil {
		return settingsResponse{}, err
	}

	return settingsResponse{
		CloudSyncEnabled: cloudSyncEnabled,
		APIKeyConfigured: configured,
		APIKeyMasked:     masked,
		APIKeySource:     source,
	}, nil
}

func (s *Server) getCloudSyncEnabled() (bool, error) {
	row := s.db.QueryRow(`SELECT value FROM app_settings WHERE key = ?`, settingKeyCloudSyncEnabled)

	var raw string
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query cloud sync setting: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true, nil
	default:
		return false, nil
	}
}

func (s *Server) setCloudSyncEnabled(enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
INSERT INTO app_settings(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
`, settingKeyCloudSyncEnabled, value, now)
	if err != nil {
		return fmt.Errorf("save cloud sync setting: %w", err)
	}
	return nil
}

func (s *Server) resolveAPIKeyState() (configured bool, masked string, source string, err error) {
	if envKey := strings.TrimSpace(os.Getenv("MOLTBB_API_KEY")); envKey != "" {
		return true, maskAPIKey(envKey), "env", nil
	}

	credentialsPath, err := utils.CredentialsPath()
	if err != nil {
		return false, "", "", err
	}
	if !utils.FileExists(credentialsPath) {
		return false, "", "", nil
	}

	credentials, err := auth.Load()
	if err != nil {
		return false, "", "", fmt.Errorf("load credentials: %w", err)
	}
	apiKey := strings.TrimSpace(credentials.APIKey)
	if apiKey == "" {
		return false, "", "", nil
	}
	return true, maskAPIKey(apiKey), "credentials", nil
}

func maskAPIKey(input string) string {
	key := strings.TrimSpace(input)
	if key == "" {
		return ""
	}
	runes := []rune(key)
	if len(runes) <= 6 {
		return strings.Repeat("*", len(runes))
	}
	head := string(runes[:3])
	tail := string(runes[len(runes)-3:])
	return head + strings.Repeat("*", len(runes)-6) + tail
}

func (s *Server) loadDiaryDetail(id string) (diaryDetail, bool, error) {
	item, found, err := s.getDiaryByID(id)
	if err != nil {
		return diaryDetail{}, false, err
	}
	if !found {
		return diaryDetail{}, false, nil
	}

	path := filepath.Join(s.diaryDir, item.RelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return diaryDetail{}, false, fmt.Errorf("read diary file: %w", err)
	}

	return diaryDetail{diarySummary: item, Content: string(data)}, true, nil
}

func (s *Server) listDiaries(q string, limit, offset int) ([]diarySummary, int, error) {
	whereSQL := ""
	args := make([]any, 0, 8)
	if strings.TrimSpace(q) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		whereSQL = ` WHERE lower(title) LIKE ? OR lower(preview) LIKE ? OR lower(filename) LIKE ? OR lower(date) LIKE ?`
		args = append(args, like, like, like, like)
	}

	countQuery := `SELECT COUNT(1) FROM diary_entries` + whereSQL
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count diaries: %w", err)
	}

	itemsQuery := `
SELECT id, date, title, preview, filename, rel_path, size, modified_at
FROM diary_entries` + whereSQL + `
ORDER BY CASE WHEN date = '' THEN 1 ELSE 0 END, date DESC, modified_at DESC
LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(itemsQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query diaries: %w", err)
	}
	defer rows.Close()

	items := make([]diarySummary, 0, limit)
	for rows.Next() {
		item, scanErr := scanDiarySummary(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read diaries rows: %w", err)
	}

	return items, total, nil
}

func (s *Server) getDiaryByID(id string) (diarySummary, bool, error) {
	row := s.db.QueryRow(`
SELECT id, date, title, preview, filename, rel_path, size, modified_at
FROM diary_entries
WHERE id = ?
`, id)

	var item diarySummary
	var size int64
	if err := row.Scan(&item.ID, &item.Date, &item.Title, &item.Preview, &item.Filename, &item.RelPath, &size, &item.ModifiedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return diarySummary{}, false, nil
		}
		return diarySummary{}, false, fmt.Errorf("query diary by id: %w", err)
	}
	item.Size = size
	return item, true, nil
}

func (s *Server) countDiaries() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(1) FROM diary_entries`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count diary entries: %w", err)
	}
	return count, nil
}

func (s *Server) reindexDiaries() (int, error) {
	items, err := s.scanDiaryFiles()
	if err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin reindex tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM diary_entries`); err != nil {
		return 0, fmt.Errorf("clear diary index: %w", err)
	}

	stmt, err := tx.Prepare(`
INSERT INTO diary_entries(id, rel_path, filename, date, title, preview, size, modified_at, indexed_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return 0, fmt.Errorf("prepare diary insert: %w", err)
	}
	defer stmt.Close()

	indexedAt := time.Now().UTC().Format(time.RFC3339)
	for _, item := range items {
		_, err := stmt.Exec(item.ID, item.RelPath, item.Filename, item.Date, item.Title, item.Preview, item.Size, item.ModifiedAt, indexedAt)
		if err != nil {
			return 0, fmt.Errorf("insert diary index row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit reindex tx: %w", err)
	}
	return len(items), nil
}

func (s *Server) scanDiaryFiles() ([]diarySummary, error) {
	items := make([]diarySummary, 0, 32)
	root := s.diaryDir

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".prompt.md") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = name
		}
		base := strings.TrimSuffix(name, ".md")
		id := filepath.ToSlash(strings.TrimSuffix(rel, ".md"))
		title, preview := extractTitleAndPreview(data)
		items = append(items, diarySummary{
			ID:         id,
			Date:       detectDiaryDate(base, data),
			Title:      title,
			Preview:    preview,
			Filename:   name,
			RelPath:    filepath.ToSlash(rel),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		di := items[i].Date
		dj := items[j].Date
		if di != "" || dj != "" {
			if di == dj {
				return items[i].ModifiedAt > items[j].ModifiedAt
			}
			return di > dj
		}
		return items[i].ModifiedAt > items[j].ModifiedAt
	})

	return items, nil
}

func loadDefaultPromptTemplate() string {
	candidates := []string{
		"prompts/bot-diary-prompt.md",
		"cli/moltbb-cli/prompts/bot-diary-prompt.md",
	}
	for _, candidate := range candidates {
		expanded, err := utils.ExpandPath(candidate)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(expanded)
		if err == nil {
			return string(data)
		}
	}

	return diary.DefaultPromptTemplate()
}

func detectDiaryDate(base string, content []byte) string {
	if matches := diaryLabelDateRe.FindStringSubmatch(string(content)); len(matches) == 2 {
		labelDate := strings.TrimSpace(matches[1])
		if _, err := time.Parse("2006-01-02", labelDate); err == nil {
			return labelDate
		}
	}

	if len(base) >= 10 {
		prefix := base[:10]
		if _, err := time.Parse("2006-01-02", prefix); err == nil {
			return prefix
		}
	}
	match := diaryDateRe.FindString(string(content))
	if match != "" {
		if _, err := time.Parse("2006-01-02", match); err == nil {
			return match
		}
	}
	return ""
}

func extractTitleAndPreview(content []byte) (string, string) {
	lines := strings.Split(string(content), "\n")
	title := "Untitled Diary"
	previewLines := make([]string, 0, 6)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if title == "Untitled Diary" {
				title = strings.TrimSpace(strings.TrimLeft(line, "#"))
			}
			continue
		}
		clean := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if clean == "" {
			continue
		}
		if title == "Untitled Diary" {
			title = truncate(clean, 80)
		}
		previewLines = append(previewLines, clean)
		if len(previewLines) >= 4 {
			break
		}
	}

	preview := strings.Join(previewLines, " ")
	if preview == "" {
		preview = "(empty diary content)"
	}
	return truncate(title, 90), truncate(preview, 220)
}

func truncate(input string, limit int) string {
	runes := []rune(strings.TrimSpace(input))
	if len(runes) <= limit {
		return string(runes)
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}

func scanDiarySummary(rows *sql.Rows) (diarySummary, error) {
	var item diarySummary
	var size int64
	if err := rows.Scan(&item.ID, &item.Date, &item.Title, &item.Preview, &item.Filename, &item.RelPath, &size, &item.ModifiedAt); err != nil {
		return diarySummary{}, fmt.Errorf("scan diary row: %w", err)
	}
	item.Size = size
	return item, nil
}

func parseInt(raw string, fallback, min, max int) int {
	value := fallback
	if strings.TrimSpace(raw) != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			value = parsed
		}
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func filterNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func rewritePrefixedPath(path string) (string, bool) {
	switch path {
	case "", "/", "/index.html", "/styles.css", "/app.js":
		return "", false
	}
	if strings.HasPrefix(path, "/api/") {
		return "", false
	}
	if strings.HasSuffix(path, "/styles.css") {
		return "/styles.css", true
	}
	if strings.HasSuffix(path, "/app.js") {
		return "/app.js", true
	}
	if strings.HasSuffix(path, "/index.html") {
		return "/index.html", true
	}
	if idx := strings.Index(path, "/api/"); idx >= 0 {
		return path[idx:], true
	}
	return "", false
}

func allowMethod(w http.ResponseWriter, r *http.Request, allowed string) bool {
	if r.Method == allowed {
		return true
	}
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	return false
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

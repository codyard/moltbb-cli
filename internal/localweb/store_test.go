package localweb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptStoreMigratesLegacyJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "local.db")
	legacyPath := filepath.Join(dir, "prompts.json")

	legacy := `{
  "activePromptId": "legacy-custom",
  "prompts": [
    {
      "id": "default",
      "name": "Default Diary Prompt",
      "description": "legacy default",
      "content": "[TODAY_STRUCTURED_SUMMARY]",
      "enabled": true,
      "builtin": true,
      "createdAt": "2026-02-01T00:00:00Z",
      "updatedAt": "2026-02-01T00:00:00Z"
    },
    {
      "id": "legacy-custom",
      "name": "Legacy Custom",
      "description": "legacy custom",
      "content": "custom prompt content",
      "enabled": true,
      "builtin": false,
      "createdAt": "2026-02-02T00:00:00Z",
      "updatedAt": "2026-02-03T00:00:00Z"
    }
  ]
}`
	if err := os.WriteFile(legacyPath, []byte(legacy), 0o600); err != nil {
		t.Fatalf("write legacy prompts: %v", err)
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	store, err := NewPromptStore(db, legacyPath, "fallback")
	if err != nil {
		t.Fatalf("new prompt store: %v", err)
	}

	metas := store.List()
	if len(metas) != 2 {
		t.Fatalf("expected 2 prompts after migration, got %d", len(metas))
	}

	active := store.ActivePromptID()
	if active == "" {
		t.Fatal("expected active prompt after migration")
	}
	if active != "legacy-custom" {
		t.Fatalf("expected active prompt legacy-custom, got %s", active)
	}

	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy prompts file moved, stat err=%v", err)
	}
}

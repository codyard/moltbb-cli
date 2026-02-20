package diary

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildRuntimeUpsertPayload_InferDateAndSummary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "2026-02-20.md")
	content := "# Daily Note\n\n- finished task A\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	payload, err := BuildRuntimeUpsertPayload(path, "", 3, time.Date(2026, 2, 21, 8, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}

	if payload.DiaryDate != "2026-02-20" {
		t.Fatalf("unexpected diary date: %s", payload.DiaryDate)
	}
	if payload.Summary != "Daily Note" {
		t.Fatalf("unexpected summary: %s", payload.Summary)
	}
	if payload.ExecutionLevel != 3 {
		t.Fatalf("unexpected execution level: %d", payload.ExecutionLevel)
	}
}

func TestBuildRuntimeUpsertPayload_ClampExecutionLevel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "daily.md")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	payload, err := BuildRuntimeUpsertPayload(path, "2026-02-20", 99, time.Now())
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if payload.ExecutionLevel != 4 {
		t.Fatalf("unexpected clamped level: %d", payload.ExecutionLevel)
	}
}

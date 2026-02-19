package localweb

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiariesAndPromptsAPI(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(diaryDir, "2026-02-19.md"), []byte("# Demo\n\nDiary content here."), 0o600); err != nil {
		t.Fatalf("write diary: %v", err)
	}

	srv, err := New(Options{
		DiaryDir:   diaryDir,
		DataDir:    dataDir,
		APIBaseURL: "https://api.moltbb.com",
		InputPaths: []string{"~/.openclaw/logs/work.log"},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/diaries", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list diaries status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var diaries diariesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &diaries); err != nil {
		t.Fatalf("decode diaries response: %v", err)
	}
	if diaries.Total != 1 {
		t.Fatalf("expected one diary, got %d", diaries.Total)
	}
	if diaries.Items[0].ID != "2026-02-19" {
		t.Fatalf("unexpected diary id: %s", diaries.Items[0].ID)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/prompts", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list prompts status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var prompts struct {
		Items []PromptMeta `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &prompts); err != nil {
		t.Fatalf("decode prompts response: %v", err)
	}
	if len(prompts.Items) == 0 {
		t.Fatal("expected default prompt")
	}
}

func TestGeneratePacketAPI(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	srv, err := New(Options{
		DiaryDir:   diaryDir,
		DataDir:    dataDir,
		APIBaseURL: "https://api.moltbb.com",
		InputPaths: []string{"/tmp/a.log", "/tmp/b.log"},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	payload := []byte(`{"date":"2026-02-19","promptId":"default"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/generate-packet", bytes.NewReader(payload))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("generate packet status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var result generatePacketResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode generate response: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success=true")
	}
	if result.PacketPath == "" {
		t.Fatal("packet path is empty")
	}

	data, err := os.ReadFile(result.PacketPath)
	if err != nil {
		t.Fatalf("read packet file: %v", err)
	}
	if !strings.Contains(string(data), "TODAY_STRUCTURED_SUMMARY") {
		t.Fatal("packet missing injected structured summary token")
	}
}

func TestPrefixedReverseProxyPaths(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(diaryDir, "2026-02-19.md"), []byte("# Demo\n\nDiary content here."), 0o600); err != nil {
		t.Fatalf("write diary: %v", err)
	}

	srv, err := New(Options{
		DiaryDir:   diaryDir,
		DataDir:    dataDir,
		APIBaseURL: "https://api.moltbb.com",
		InputPaths: []string{"/tmp/work.log"},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moltbb-local/styles.css", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("prefixed styles.css status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Fatalf("expected css content-type, got %q", ct)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/moltbb-local/api/state", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("prefixed /api/state status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var state stateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode state response: %v", err)
	}
	if state.DatabasePath == "" {
		t.Fatal("expected databasePath in state response")
	}
}

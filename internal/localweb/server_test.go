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
	"time"
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

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/moltbb-local/icon.png", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("prefixed icon.png status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("expected png content-type, got %q", ct)
	}
}

func TestDiariesOrderedByDiaryDateDesc(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	olderPath := filepath.Join(diaryDir, "alpha.md")
	newerPath := filepath.Join(diaryDir, "beta.md")

	olderContent := "# Old Entry\n\n- Date: 2026-02-10\n\nold"
	newerContent := "# New Entry\n\n- Date: 2026-02-19\n\nnew"

	if err := os.WriteFile(olderPath, []byte(olderContent), 0o600); err != nil {
		t.Fatalf("write older diary: %v", err)
	}
	if err := os.WriteFile(newerPath, []byte(newerContent), 0o600); err != nil {
		t.Fatalf("write newer diary: %v", err)
	}

	// Force mtime so that newer-date diary has older mtime; ordering should still follow diary date.
	oldMTime := time.Now().Add(-1 * time.Hour)
	newMTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(olderPath, oldMTime, oldMTime); err != nil {
		t.Fatalf("set older diary mtime: %v", err)
	}
	if err := os.Chtimes(newerPath, newMTime, newMTime); err != nil {
		t.Fatalf("set newer diary mtime: %v", err)
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
	req := httptest.NewRequest(http.MethodGet, "/api/diaries?limit=10", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list diaries status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var response diariesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode diaries response: %v", err)
	}
	if len(response.Items) < 2 {
		t.Fatalf("expected at least 2 diaries, got %d", len(response.Items))
	}
	if response.Items[0].Date != "2026-02-19" {
		t.Fatalf("expected first diary date 2026-02-19, got %s", response.Items[0].Date)
	}
	if response.Items[1].Date != "2026-02-10" {
		t.Fatalf("expected second diary date 2026-02-10, got %s", response.Items[1].Date)
	}
}

func TestSettingsCloudSyncToggle(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

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
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list settings status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var initial settingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &initial); err != nil {
		t.Fatalf("decode initial settings: %v", err)
	}
	if initial.CloudSyncEnabled {
		t.Fatal("expected cloud sync disabled by default")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/settings", strings.NewReader(`{"cloudSyncEnabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch settings status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var updated settingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated settings: %v", err)
	}
	if !updated.CloudSyncEnabled {
		t.Fatal("expected cloud sync enabled after patch")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list settings status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var final settingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &final); err != nil {
		t.Fatalf("decode final settings: %v", err)
	}
	if !final.CloudSyncEnabled {
		t.Fatal("expected cloud sync enabled after reload")
	}
}

func TestSettingsTestConnectionSuccess(t *testing.T) {
	expectedAPIKey := "sk-test-123456"
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/validate" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("X-API-Key"); got != expectedAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"success":false,"message":"invalid key"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"valid":true}}`))
	}))
	defer remote.Close()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	srv, err := New(Options{
		DiaryDir:   diaryDir,
		DataDir:    dataDir,
		APIBaseURL: remote.URL,
		InputPaths: []string{"/tmp/work.log"},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settings/test-connection", strings.NewReader(`{"apiKey":"`+expectedAPIKey+`"}`))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("test connection status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var result settingsConnectionTestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode test connection response: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success=true, message=%q", result.Message)
	}
	if !result.Connected {
		t.Fatal("expected connected=true")
	}
	if !result.Authenticated {
		t.Fatal("expected authenticated=true")
	}
	if result.KeySource != "request" {
		t.Fatalf("expected keySource=request, got %q", result.KeySource)
	}
}

func TestDiarySearchMatchesFullContent(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(diaryDir, "2026-02-20.md"), []byte("# Notes\n\nAlpha signal from deep memory"), 0o600); err != nil {
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
	req := httptest.NewRequest(http.MethodGet, "/api/diaries?q=deep%20memory", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search diaries status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var resp diariesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected 1 result, got %d", resp.Total)
	}
}

func TestPatchDiaryContentPersistsAndIsSearchable(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()
	diaryPath := filepath.Join(diaryDir, "2026-02-20.md")
	if err := os.WriteFile(diaryPath, []byte("# Before\n\nold content"), 0o600); err != nil {
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

	payload := `{"content":"# After\n\nneedle phrase in body"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/diaries/2026-02-20", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch diary status = %d, body=%s", rec.Code, rec.Body.String())
	}

	updatedBytes, err := os.ReadFile(diaryPath)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if !strings.Contains(string(updatedBytes), "needle phrase in body") {
		t.Fatalf("patched file missing updated content: %s", string(updatedBytes))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/diaries?q=needle%20phrase", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search diaries status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var resp diariesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected 1 result after patch, got %d", resp.Total)
	}
}

package localweb

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/moltbb-local/pure-logo.png", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("prefixed pure-logo.png status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("expected pure-logo png content-type, got %q", ct)
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

func TestInsightsAPIRequiresAPIKey(t *testing.T) {
	t.Setenv("MOLTBB_API_KEY", "")

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
	req := httptest.NewRequest(http.MethodGet, "/api/insights", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("list insights status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "API key is not configured") {
		t.Fatalf("expected api key configured message, got=%s", rec.Body.String())
	}
}

func TestInsightsAPIProxyCRUD(t *testing.T) {
	expectedAPIKey := "sk-insight-123456"
	t.Setenv("MOLTBB_API_KEY", expectedAPIKey)

	insightID := "11111111-1111-1111-1111-111111111111"
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != expectedAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"success":false,"message":"invalid key"}`))
			return
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/runtime/insights":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "success": true,
  "data": [{
    "id": "` + insightID + `",
    "botId": "bot-1",
    "title": "Alpha Insight",
    "content": "alpha content",
    "tags": ["alpha"],
    "catalogs": ["engineering"],
    "visibilityLevel": 0,
    "likes": 3,
    "createdAt": "2026-02-20T10:00:00Z",
    "updatedAt": "2026-02-20T10:30:00Z"
  }],
  "pagination": {"page": 1, "pageSize": 100, "totalCount": 1, "totalPages": 1}
}`))
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/runtime/insights":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "success": true,
  "data": {
    "id": "22222222-2222-2222-2222-222222222222",
    "botId": "bot-1",
    "title": "Created Insight",
    "content": "created content",
    "tags": ["new"],
    "catalogs": ["product"],
    "visibilityLevel": 1,
    "likes": 0,
    "createdAt": "2026-02-20T11:00:00Z",
    "updatedAt": "2026-02-20T11:00:00Z"
  }
}`))
			return
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/v1/runtime/insights/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "success": true,
  "data": {
    "id": "` + insightID + `",
    "botId": "bot-1",
    "title": "Updated Insight",
    "content": "updated content",
    "tags": ["alpha", "beta"],
    "catalogs": ["engineering"],
    "visibilityLevel": 1,
    "likes": 5,
    "createdAt": "2026-02-20T10:00:00Z",
    "updatedAt": "2026-02-20T12:00:00Z"
  }
}`))
			return
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/runtime/insights/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true}`))
			return
		default:
			http.NotFound(w, r)
		}
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
	req := httptest.NewRequest(http.MethodGet, "/api/insights?q=alpha", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list insights status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var listResp insightsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Items) != 1 {
		t.Fatalf("expected 1 insight in list response, got total=%d len=%d", listResp.Total, len(listResp.Items))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/insights/"+url.PathEscape(insightID), nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get insight status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/insights", strings.NewReader(`{
  "title":"Created Insight",
  "content":"created content",
  "tags":["new"],
  "catalogs":["product"],
  "visibilityLevel":1
}`))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create insight status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/insights/"+url.PathEscape(insightID), strings.NewReader(`{
  "title":"Updated Insight",
  "content":"updated content",
  "tags":["alpha","beta"],
  "visibilityLevel":1
}`))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update insight status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/insights/"+url.PathEscape(insightID), nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete insight status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestInsightsListReturnsUnsupportedWhenRuntimeEndpointMissing(t *testing.T) {
	expectedAPIKey := "sk-insight-unsupported"
	t.Setenv("MOLTBB_API_KEY", expectedAPIKey)

	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != expectedAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"success":false,"message":"invalid key"}`))
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/runtime/insights" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":"not found"}`))
			return
		}
		http.NotFound(w, r)
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
	req := httptest.NewRequest(http.MethodGet, "/api/insights", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list insights status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var listResp insightsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if !listResp.Unsupported {
		t.Fatalf("expected unsupported=true, got %+v", listResp)
	}
	if listResp.Total != 0 || len(listResp.Items) != 0 {
		t.Fatalf("expected empty insights when unsupported, got total=%d len=%d", listResp.Total, len(listResp.Items))
	}
	if !strings.Contains(strings.ToLower(listResp.Notice), "runtime insights endpoint is unavailable") {
		t.Fatalf("expected unsupported notice, got %q", listResp.Notice)
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

func TestDiaryDefaultSelectionAndSetDefault(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	aPath := filepath.Join(diaryDir, "2026-02-20-a.md")
	bPath := filepath.Join(diaryDir, "2026-02-20-b.md")
	if err := os.WriteFile(aPath, []byte("# A\n\n- Date: 2026-02-20"), 0o600); err != nil {
		t.Fatalf("write diary a: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("# B\n\n- Date: 2026-02-20"), 0o600); err != nil {
		t.Fatalf("write diary b: %v", err)
	}

	oldTime := time.Date(2026, 2, 20, 1, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 2, 20, 2, 0, 0, 0, time.UTC)
	if err := os.Chtimes(aPath, oldTime, oldTime); err != nil {
		t.Fatalf("set diary a mtime: %v", err)
	}
	if err := os.Chtimes(bPath, newTime, newTime); err != nil {
		t.Fatalf("set diary b mtime: %v", err)
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

	var resp diariesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode diaries response: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 diaries, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != "2026-02-20-b" {
		t.Fatalf("expected latest diary first, got %s", resp.Items[0].ID)
	}
	if !resp.Items[0].IsDefault {
		t.Fatal("expected latest diary to be default")
	}
	if resp.Items[1].IsDefault {
		t.Fatal("expected older diary not default before manual set")
	}

	targetID := resp.Items[1].ID
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/diaries/"+url.PathEscape(targetID)+"/set-default", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("set-default status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/diaries?limit=10", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list diaries after set-default status = %d, body=%s", rec.Code, rec.Body.String())
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode diaries response after set-default: %v", err)
	}

	defaultCount := 0
	for _, item := range resp.Items {
		if item.IsDefault {
			defaultCount++
			if item.ID != targetID {
				t.Fatalf("expected %s as default, got %s", targetID, item.ID)
			}
		}
	}
	if defaultCount != 1 {
		t.Fatalf("expected exactly one default diary, got %d", defaultCount)
	}
}

func TestDiaryHistoryAPI_ReturnsPerDayStatus(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()

	aPath := filepath.Join(diaryDir, "2026-02-18-a.md")
	bPath := filepath.Join(diaryDir, "2026-02-18-b.md")
	cPath := filepath.Join(diaryDir, "2026-02-17.md")
	if err := os.WriteFile(aPath, []byte("# A\n\n- Date: 2026-02-18"), 0o600); err != nil {
		t.Fatalf("write diary a: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("# B\n\n- Date: 2026-02-18"), 0o600); err != nil {
		t.Fatalf("write diary b: %v", err)
	}
	if err := os.WriteFile(cPath, []byte("# C\n\n- Date: 2026-02-17"), 0o600); err != nil {
		t.Fatalf("write diary c: %v", err)
	}

	oldTime := time.Date(2026, 2, 18, 1, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 2, 18, 2, 0, 0, 0, time.UTC)
	if err := os.Chtimes(aPath, oldTime, oldTime); err != nil {
		t.Fatalf("set diary a mtime: %v", err)
	}
	if err := os.Chtimes(bPath, newTime, newTime); err != nil {
		t.Fatalf("set diary b mtime: %v", err)
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
	req := httptest.NewRequest(http.MethodGet, "/api/diaries/history", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("history status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var history diaryHistoryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &history); err != nil {
		t.Fatalf("decode history response: %v", err)
	}
	if history.Total != 2 {
		t.Fatalf("expected 2 diary dates, got %d", history.Total)
	}
	if len(history.Items) != 2 {
		t.Fatalf("expected 2 history items, got %d", len(history.Items))
	}
	if history.Items[0].Date != "2026-02-18" {
		t.Fatalf("expected first history date 2026-02-18, got %s", history.Items[0].Date)
	}
	if history.Items[0].DiaryCount != 2 {
		t.Fatalf("expected 2 diaries on 2026-02-18, got %d", history.Items[0].DiaryCount)
	}
	if !history.Items[0].HasDefault {
		t.Fatal("expected hasDefault=true for 2026-02-18")
	}
	if history.Items[0].DefaultDiaryID != "2026-02-18-b" {
		t.Fatalf("expected default diary 2026-02-18-b, got %s", history.Items[0].DefaultDiaryID)
	}
	if history.Items[0].DefaultIsManual {
		t.Fatal("expected defaultIsManual=false before manual override")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/diaries/2026-02-18-a/set-default", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("set default status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/diaries/history", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("history after set-default status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &history); err != nil {
		t.Fatalf("decode history response after set-default: %v", err)
	}
	if history.Items[0].DefaultDiaryID != "2026-02-18-a" {
		t.Fatalf("expected manual default diary 2026-02-18-a, got %s", history.Items[0].DefaultDiaryID)
	}
	if !history.Items[0].DefaultIsManual {
		t.Fatal("expected defaultIsManual=true after manual set-default")
	}
}

func TestSyncDiary_WithCloudSyncDisabled_ReturnsExplicitReason(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(diaryDir, "2026-02-20.md"), []byte("# A\n\n- Date: 2026-02-20"), 0o600); err != nil {
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
	req := httptest.NewRequest(http.MethodPost, "/api/diaries/2026-02-20/sync", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("sync status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cloud sync is disabled") {
		t.Fatalf("expected cloud sync disabled reason, got body=%s", rec.Body.String())
	}

	logData, err := os.ReadFile(filepath.Join(dataDir, syncDiagnosticsLogFileName))
	if err != nil {
		t.Fatalf("read sync diagnostics log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(logData)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Fatalf("expected sync diagnostics log line, got=%q", string(logData))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("decode sync diagnostics log line: %v", err)
	}
	if got, ok := entry["event"].(string); !ok || got != "diary_sync_blocked" {
		t.Fatalf("expected blocked event, got=%v", entry["event"])
	}
	if got, ok := entry["stage"].(string); !ok || got != "precheck_cloud_sync" {
		t.Fatalf("expected precheck_cloud_sync stage, got=%v", entry["stage"])
	}
	if got, ok := entry["diaryId"].(string); !ok || got != "2026-02-20" {
		t.Fatalf("expected diaryId=2026-02-20, got=%v", entry["diaryId"])
	}
	if got, ok := entry["cloudSyncEnabled"].(bool); !ok || got {
		t.Fatalf("expected cloudSyncEnabled=false, got=%v", entry["cloudSyncEnabled"])
	}
	if got, ok := entry["error"].(string); !ok || !strings.Contains(got, "cloud sync is disabled") {
		t.Fatalf("expected cloud sync error in log, got=%v", entry["error"])
	}
	if got, ok := entry["timestamp"].(string); !ok || strings.TrimSpace(got) == "" {
		t.Fatalf("expected timestamp in sync diagnostics log, got=%v", entry["timestamp"])
	}
}

func TestSyncDiary_WithCloudSyncEnabledAndNoAPIKey_ReturnsExplicitReason(t *testing.T) {
	t.Parallel()

	diaryDir := t.TempDir()
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(diaryDir, "2026-02-20.md"), []byte("# A\n\n- Date: 2026-02-20"), 0o600); err != nil {
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
	req := httptest.NewRequest(http.MethodPatch, "/api/settings", strings.NewReader(`{"cloudSyncEnabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable cloud sync status = %d, body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/diaries/2026-02-20/sync", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("sync status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "API key is not configured") {
		t.Fatalf("expected api key not configured reason, got body=%s", rec.Body.String())
	}

	logData, err := os.ReadFile(filepath.Join(dataDir, syncDiagnosticsLogFileName))
	if err != nil {
		t.Fatalf("read sync diagnostics log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(logData)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Fatalf("expected sync diagnostics log line, got=%q", string(logData))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("decode sync diagnostics log line: %v", err)
	}
	if got, ok := entry["event"].(string); !ok || got != "diary_sync_blocked" {
		t.Fatalf("expected blocked event, got=%v", entry["event"])
	}
	if got, ok := entry["stage"].(string); !ok || got != "precheck_api_key" {
		t.Fatalf("expected precheck_api_key stage, got=%v", entry["stage"])
	}
	if got, ok := entry["cloudSyncEnabled"].(bool); !ok || !got {
		t.Fatalf("expected cloudSyncEnabled=true, got=%v", entry["cloudSyncEnabled"])
	}
	if got, ok := entry["apiKeyConfigured"].(bool); !ok || got {
		t.Fatalf("expected apiKeyConfigured=false, got=%v", entry["apiKeyConfigured"])
	}
	if got, ok := entry["error"].(string); !ok || !strings.Contains(got, "API key is not configured") {
		t.Fatalf("expected api key error in log, got=%v", entry["error"])
	}
}

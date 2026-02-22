package localweb

import (
	"context"
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

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/diary"
	"moltbb-cli/internal/utils"
)

//go:embed static/*
var staticFS embed.FS

var diaryDateRe = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
var diaryLabelDateRe = regexp.MustCompile(`(?im)^\s*(?:[-*]\s*)?(?:date|日期)\s*:\s*(\d{4}-\d{2}-\d{2})\s*$`)

const settingKeyCloudSyncEnabled = "cloud_sync_enabled"
const syncDiagnosticsLogFileName = "sync.log"

type Options struct {
	DiaryDir   string
	DataDir    string
	APIBaseURL string
	InputPaths []string
	Version    string
}

type Server struct {
	diaryDir   string
	dataDir    string
	dbPath     string
	apiBaseURL string
	inputPaths []string
	version    string
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
	IsDefault  bool   `json:"isDefault"`
	CanSync    bool   `json:"canSync"`
	SearchText string `json:"-"`
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

type diaryHistoryItem struct {
	Date             string `json:"date"`
	DiaryCount       int    `json:"diaryCount"`
	DefaultDiaryID   string `json:"defaultDiaryId,omitempty"`
	DefaultIsManual  bool   `json:"defaultIsManual"`
	HasDefault       bool   `json:"hasDefault"`
	LatestModifiedAt string `json:"latestModifiedAt,omitempty"`
}

type diaryHistoryResponse struct {
	Items []diaryHistoryItem `json:"items"`
	Total int                `json:"total"`
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
	Version       string `json:"version"`
}

type settingsResponse struct {
	CloudSyncEnabled bool   `json:"cloudSyncEnabled"`
	APIKeyConfigured bool   `json:"apiKeyConfigured"`
	APIKeyMasked     string `json:"apiKeyMasked,omitempty"`
	APIKeySource     string `json:"apiKeySource,omitempty"`
	Bound            bool   `json:"bound"`
	BotID            string `json:"botId,omitempty"`
	OwnerID          string `json:"ownerId,omitempty"`
	OwnerNickname    string `json:"ownerNickname,omitempty"`
	SetupComplete    bool   `json:"setupComplete"`
}

type settingsUpdateRequest struct {
	CloudSyncEnabled *bool   `json:"cloudSyncEnabled,omitempty"`
	APIKey           *string `json:"apiKey,omitempty"`
}

type settingsConnectionTestRequest struct {
	APIKey string `json:"apiKey,omitempty"`
}

type settingsConnectionTestResponse struct {
	Success       bool   `json:"success"`
	Connected     bool   `json:"connected"`
	Authenticated bool   `json:"authenticated"`
	APIBaseURL    string `json:"apiBaseUrl"`
	KeySource     string `json:"keySource,omitempty"`
	Message       string `json:"message"`
	CheckedAt     string `json:"checkedAt"`
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

type diaryUpdateRequest struct {
	Content *string `json:"content"`
}

type diarySyncResponse struct {
	Success    bool   `json:"success"`
	DiaryID    string `json:"diaryId,omitempty"`
	Action     string `json:"action"`
	StatusCode int    `json:"statusCode"`
}

type insightSummary struct {
	ID              string   `json:"id"`
	BotID           string   `json:"botId"`
	DiaryID         string   `json:"diaryId,omitempty"`
	Title           string   `json:"title"`
	Catalogs        []string `json:"catalogs,omitempty"`
	Content         string   `json:"content"`
	Tags            []string `json:"tags,omitempty"`
	VisibilityLevel int      `json:"visibilityLevel"`
	Likes           int      `json:"likes"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
	SearchText      string   `json:"-"`
}

type insightsResponse struct {
	Items       []insightSummary `json:"items"`
	Total       int              `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	TotalPages  int              `json:"totalPages"`
	Unsupported bool             `json:"unsupported,omitempty"`
	Notice      string           `json:"notice,omitempty"`
}

type insightCreateRequest struct {
	Title           string   `json:"title"`
	DiaryID         string   `json:"diaryId,omitempty"`
	Catalogs        []string `json:"catalogs,omitempty"`
	Content         string   `json:"content"`
	Tags            []string `json:"tags,omitempty"`
	VisibilityLevel int      `json:"visibilityLevel,omitempty"`
}

type insightUpdateRequest struct {
	Title           *string  `json:"title,omitempty"`
	Catalogs        []string `json:"catalogs,omitempty"`
	Content         *string  `json:"content,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	VisibilityLevel *int     `json:"visibilityLevel,omitempty"`
}

type dayDefaultRecord struct {
	DiaryID  string
	IsManual bool
}

type syncDiagContext struct {
	DiaryID          string
	DiaryDate        string
	DiaryTitle       string
	DiaryFilename    string
	DiaryRelPath     string
	DiaryPath        string
	IsDefault        *bool
	CloudSyncEnabled *bool
	APIKeyConfigured *bool
	APIKeySource     string
	APIBaseURL       string
}

type syncLogEntry struct {
	Timestamp        string `json:"timestamp"`
	Level            string `json:"level"`
	Event            string `json:"event"`
	Stage            string `json:"stage,omitempty"`
	DiaryID          string `json:"diaryId,omitempty"`
	DiaryDate        string `json:"diaryDate,omitempty"`
	DiaryTitle       string `json:"diaryTitle,omitempty"`
	DiaryFilename    string `json:"diaryFilename,omitempty"`
	DiaryRelPath     string `json:"diaryRelPath,omitempty"`
	DiaryPath        string `json:"diaryPath,omitempty"`
	IsDefault        *bool  `json:"isDefault,omitempty"`
	CloudSyncEnabled *bool  `json:"cloudSyncEnabled,omitempty"`
	APIKeyConfigured *bool  `json:"apiKeyConfigured,omitempty"`
	APIKeySource     string `json:"apiKeySource,omitempty"`
	APIBaseURL       string `json:"apiBaseUrl,omitempty"`
	Action           string `json:"action,omitempty"`
	StatusCode       int    `json:"statusCode,omitempty"`
	RemoteDiaryID    string `json:"remoteDiaryId,omitempty"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
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
		version:    strings.TrimSpace(options.Version),
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
	s.mux.HandleFunc("/api/settings/test-connection", s.handleSettingsConnectionTest)
	s.mux.HandleFunc("/api/diaries", s.handleDiaries)
	s.mux.HandleFunc("/api/diaries/history", s.handleDiaryHistory)
	s.mux.HandleFunc("/api/diaries/reindex", s.handleReindex)
	s.mux.HandleFunc("/api/diaries/", s.handleDiaryByID)
	s.mux.HandleFunc("/api/insights", s.handleInsights)
	s.mux.HandleFunc("/api/insights/", s.handleInsightByID)
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
	s.mux.Handle("/icon.png", fileServer)
	s.mux.Handle("/pure-logo.png", fileServer)
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
		Version:       s.version,
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

func (s *Server) handleSettingsConnectionTest(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	var req settingsConnectionTestRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	result, err := s.testSettingsConnection(strings.TrimSpace(req.APIKey))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
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

func (s *Server) handleDiaryHistory(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	items, err := s.listDiaryHistory()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, diaryHistoryResponse{
		Items: items,
		Total: len(items),
	})
}

func (s *Server) handleDiaryByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/diaries/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "diary id is required"})
		return
	}

	parts := strings.Split(path, "/")
	id := strings.TrimSpace(parts[0])
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "diary id is required"})
		return
	}
	decoded, decodeErr := url.PathUnescape(id)
	if decodeErr == nil {
		id = decoded
	}

	if len(parts) == 2 && parts[1] == "set-default" {
		if !allowMethod(w, r, http.MethodPost) {
			return
		}
		detail, found, err := s.setDiaryAsDefault(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "diary not found"})
			return
		}
		writeJSON(w, http.StatusOK, detail)
		return
	}

	if len(parts) == 2 && parts[1] == "sync" {
		if !allowMethod(w, r, http.MethodPost) {
			return
		}
		result, found, err := s.syncDiaryByID(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "diary not found"})
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	if len(parts) > 1 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "route not found"})
		return
	}

	switch r.Method {
	case http.MethodGet:
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
	case http.MethodPatch:
		var req diaryUpdateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if req.Content == nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "content is required"})
			return
		}
		detail, found, err := s.saveDiaryContent(id, *req.Content)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "diary not found"})
			return
		}
		writeJSON(w, http.StatusOK, detail)
	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		page := parseInt(r.URL.Query().Get("page"), 1, 1, 1_000_000)
		pageSize := parseInt(r.URL.Query().Get("pageSize"), 100, 1, 100)
		tags := normalizeStringListValues(splitCommaList(r.URL.Query()["tags"]))
		diaryID := strings.TrimSpace(r.URL.Query().Get("diaryId"))

		client, apiKey, cfg, err := s.runtimeClientWithAPIKey()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
		defer cancel()

		resp, err := s.listInsights(ctx, client, apiKey, page, pageSize, tags, diaryID, q)
		if err != nil {
			writeError(w, http.StatusBadRequest, normalizeRuntimeInsightsError(err))
			return
		}
		writeJSON(w, http.StatusOK, resp)
	case http.MethodPost:
		var req insightCreateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "title is required"})
			return
		}
		if strings.TrimSpace(req.Content) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "content is required"})
			return
		}
		if req.VisibilityLevel < 0 || req.VisibilityLevel > 1 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "visibilityLevel must be 0 or 1"})
			return
		}

		client, apiKey, cfg, err := s.runtimeClientWithAPIKey()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
		defer cancel()

		created, err := client.CreateRuntimeInsight(ctx, apiKey, api.RuntimeInsightCreatePayload{
			Title:           strings.TrimSpace(req.Title),
			DiaryID:         strings.TrimSpace(req.DiaryID),
			Catalogs:        normalizeStringListValues(req.Catalogs),
			Content:         strings.TrimSpace(req.Content),
			Tags:            normalizeStringListValues(req.Tags),
			VisibilityLevel: req.VisibilityLevel,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, normalizeRuntimeInsightsError(err))
			return
		}
		writeJSON(w, http.StatusCreated, mapRuntimeInsight(created))
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (s *Server) handleInsightByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/insights/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "insight id is required"})
		return
	}

	parts := strings.Split(path, "/")
	id := strings.TrimSpace(parts[0])
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "insight id is required"})
		return
	}
	decoded, decodeErr := url.PathUnescape(id)
	if decodeErr == nil {
		id = decoded
	}
	if len(parts) > 1 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "route not found"})
		return
	}

	client, apiKey, cfg, err := s.runtimeClientWithAPIKey()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		item, found, err := s.getInsightByID(ctx, client, apiKey, id)
		if err != nil {
			writeError(w, http.StatusBadRequest, normalizeRuntimeInsightsError(err))
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "insight not found"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var req insightUpdateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		title := trimOptionalString(req.Title)
		content := trimOptionalString(req.Content)
		tags := normalizeStringListValues(req.Tags)
		catalogs := normalizeStringListValues(req.Catalogs)
		if req.VisibilityLevel != nil {
			if *req.VisibilityLevel < 0 || *req.VisibilityLevel > 1 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "visibilityLevel must be 0 or 1"})
				return
			}
		}
		if title == nil && content == nil && req.VisibilityLevel == nil && len(tags) == 0 && len(catalogs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "at least one field is required for insight update"})
			return
		}

		updated, err := client.UpdateRuntimeInsight(ctx, apiKey, id, api.RuntimeInsightUpdatePayload{
			Title:           title,
			Catalogs:        catalogs,
			Content:         content,
			Tags:            tags,
			VisibilityLevel: req.VisibilityLevel,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, normalizeRuntimeInsightsError(err))
			return
		}
		writeJSON(w, http.StatusOK, mapRuntimeInsight(updated))
	case http.MethodDelete:
		if err := client.DeleteRuntimeInsight(ctx, apiKey, id); err != nil {
			writeError(w, http.StatusBadRequest, normalizeRuntimeInsightsError(err))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
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

	// 读取 binding 状态：通过 API 验证而不是只看本地文件
	bound, botID, ownerID, ownerNickname := s.resolveBindingStateWithAPI()

	// 判断设置是否完成：需要同时有 API key 和绑定（owner ID 存在）
	setupComplete := configured && bound

	return settingsResponse{
		CloudSyncEnabled: cloudSyncEnabled,
		APIKeyConfigured: configured,
		APIKeyMasked:     masked,
		APIKeySource:     source,
		Bound:            bound,
		BotID:            botID,
		OwnerID:          ownerID,
		OwnerNickname:    ownerNickname,
		SetupComplete:    setupComplete,
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

func (s *Server) resolveBindingStateWithAPI() (bound bool, botID string, ownerID string, ownerNickname string) {
	// 从本地文件读取 bot ID
	state, err := binding.Load()
	if err == nil && state.Bound {
		botID = strings.TrimSpace(state.BotID)
	}

	// 通过 API 验证来确定是否真正绑定（需要有 owner ID）
	apiKey, keySource := s.resolveAPIKeyForConnectionTest("")
	if apiKey == "" {
		// 没有 API key，无法验证绑定
		return false, botID, "", ""
	}

	// 调用 ValidateAPIKey 获取 owner ID 和 nickname
	baseURL := strings.TrimSpace(s.apiBaseURL)
	if baseURL == "" {
		baseURL = config.DefaultAPIBaseURL
	}

	cfg := config.Default()
	cfg.APIBaseURL = baseURL
	if strings.HasPrefix(baseURL, "http://") {
		cfg.AllowInsecureHTTP = true
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		// 无法创建 client，返回未绑定
		return false, botID, "", ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()

	validateResp, err := client.ValidateAPIKey(ctx, apiKey)
	if err != nil || !validateResp.Valid {
		// API key 无效，返回未绑定
		return false, botID, "", ""
	}

	ownerID = strings.TrimSpace(validateResp.OwnerID)
	ownerNickname = strings.TrimSpace(validateResp.OwnerNickname)
	if ownerID == "" {
		// API key 有效但没有 owner ID，说明还未绑定 owner
		return false, botID, "", ""
	}

	// 有 owner ID，说明已绑定
	_ = keySource // 避免 unused 警告
	return true, botID, ownerID, ownerNickname
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

func (s *Server) testSettingsConnection(apiKeyOverride string) (settingsConnectionTestResponse, error) {
	baseURL := strings.TrimSpace(s.apiBaseURL)
	if baseURL == "" {
		baseURL = config.DefaultAPIBaseURL
	}

	cfg := config.Default()
	cfg.APIBaseURL = baseURL
	if strings.HasPrefix(baseURL, "http://") {
		cfg.AllowInsecureHTTP = true
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return settingsConnectionTestResponse{}, fmt.Errorf("create api client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()

	apiKey, keySource := s.resolveAPIKeyForConnectionTest(apiKeyOverride)
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	if apiKey == "" {
		if err := client.Ping(ctx); err != nil {
			return settingsConnectionTestResponse{
				Success:       false,
				Connected:     false,
				Authenticated: false,
				APIBaseURL:    baseURL,
				KeySource:     keySource,
				Message:       fmt.Sprintf("Failed to connect to API: %v", err),
				CheckedAt:     checkedAt,
			}, nil
		}
		return settingsConnectionTestResponse{
			Success:       true,
			Connected:     true,
			Authenticated: false,
			APIBaseURL:    baseURL,
			KeySource:     keySource,
			Message:       "Connected to API, but API key is not configured.",
			CheckedAt:     checkedAt,
		}, nil
	}

	validateResp, validateErr := client.ValidateAPIKey(ctx, apiKey)
	if validateErr == nil && validateResp.Valid {
		return settingsConnectionTestResponse{
			Success:       true,
			Connected:     true,
			Authenticated: true,
			APIBaseURL:    baseURL,
			KeySource:     keySource,
			Message:       "Connection successful and API key is valid.",
			CheckedAt:     checkedAt,
		}, nil
	}

	if pingErr := client.Ping(ctx); pingErr == nil {
		msg := "API reachable, but API key validation failed."
		if validateErr != nil {
			msg = fmt.Sprintf("%s %v", msg, validateErr)
		}
		return settingsConnectionTestResponse{
			Success:       false,
			Connected:     true,
			Authenticated: false,
			APIBaseURL:    baseURL,
			KeySource:     keySource,
			Message:       msg,
			CheckedAt:     checkedAt,
		}, nil
	}

	msg := "Failed to connect to API."
	if validateErr != nil {
		msg = fmt.Sprintf("%s %v", msg, validateErr)
	}
	return settingsConnectionTestResponse{
		Success:       false,
		Connected:     false,
		Authenticated: false,
		APIBaseURL:    baseURL,
		KeySource:     keySource,
		Message:       msg,
		CheckedAt:     checkedAt,
	}, nil
}

func (s *Server) resolveAPIKeyForConnectionTest(override string) (apiKey string, source string) {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed, "request"
	}
	if envKey := strings.TrimSpace(os.Getenv("MOLTBB_API_KEY")); envKey != "" {
		return envKey, "env"
	}
	creds, err := auth.Load()
	if err != nil {
		return "", ""
	}
	apiKey = strings.TrimSpace(creds.APIKey)
	if apiKey == "" {
		return "", ""
	}
	return apiKey, "credentials"
}

func (s *Server) runtimeClientWithAPIKey() (*api.Client, string, config.Config, error) {
	cfg := config.Default()
	base := strings.TrimSpace(s.apiBaseURL)
	if base == "" {
		base = config.DefaultAPIBaseURL
	}
	cfg.APIBaseURL = base
	if strings.HasPrefix(base, "http://") {
		cfg.AllowInsecureHTTP = true
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, "", cfg, fmt.Errorf("create api client: %w", err)
	}
	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		return nil, "", cfg, normalizeResolveAPIKeyError(err)
	}
	return client, strings.TrimSpace(apiKey), cfg, nil
}

func normalizeResolveAPIKeyError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if msg == "" {
		return errors.New("resolve api key failed")
	}
	if strings.Contains(msg, "credentials file not found") || strings.Contains(msg, "credentials missing api_key") {
		return errors.New("API key is not configured. Set it in Settings or run `moltbb login --apikey <key>`")
	}
	return fmt.Errorf("resolve api key: %w", err)
}

func (s *Server) listInsights(
	ctx context.Context,
	client *api.Client,
	apiKey string,
	page int,
	pageSize int,
	tags []string,
	diaryID string,
	q string,
) (insightsResponse, error) {
	result, err := client.ListRuntimeInsights(ctx, apiKey, page, pageSize, tags, diaryID)
	if err != nil {
		if isRuntimeInsightsNotFoundError(err) {
			return insightsResponse{
				Items:       []insightSummary{},
				Total:       0,
				Page:        page,
				PageSize:    pageSize,
				TotalPages:  1,
				Unsupported: true,
				Notice:      runtimeInsightsUnsupportedMessage(),
			}, nil
		}
		return insightsResponse{}, err
	}

	items := make([]insightSummary, 0, len(result.Items))
	for _, item := range result.Items {
		mapped := mapRuntimeInsight(item)
		if q != "" && !strings.Contains(mapped.SearchText, q) {
			continue
		}
		items = append(items, mapped)
	}

	total := result.TotalCount
	totalPages := result.TotalPages
	if q != "" {
		total = len(items)
		totalPages = 1
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	return insightsResponse{
		Items:      items,
		Total:      total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *Server) getInsightByID(ctx context.Context, client *api.Client, apiKey, id string) (insightSummary, bool, error) {
	const pageSize = 100
	page := 1
	for {
		result, err := client.ListRuntimeInsights(ctx, apiKey, page, pageSize, nil, "")
		if err != nil {
			return insightSummary{}, false, err
		}
		for _, item := range result.Items {
			if strings.TrimSpace(item.ID) == id {
				return mapRuntimeInsight(item), true, nil
			}
		}
		if result.TotalPages <= 0 || page >= result.TotalPages || len(result.Items) == 0 {
			return insightSummary{}, false, nil
		}
		page++
	}
}

func mapRuntimeInsight(input api.RuntimeInsight) insightSummary {
	item := insightSummary{
		ID:              strings.TrimSpace(input.ID),
		BotID:           strings.TrimSpace(input.BotID),
		DiaryID:         strings.TrimSpace(input.DiaryID),
		Title:           strings.TrimSpace(input.Title),
		Catalogs:        normalizeStringListValues(input.Catalogs),
		Content:         strings.TrimSpace(input.Content),
		Tags:            normalizeStringListValues(input.Tags),
		VisibilityLevel: input.VisibilityLevel,
		Likes:           input.Likes,
		CreatedAt:       strings.TrimSpace(input.CreatedAt),
		UpdatedAt:       strings.TrimSpace(input.UpdatedAt),
	}
	item.SearchText = buildInsightSearchText(item)
	return item
}

func buildInsightSearchText(item insightSummary) string {
	fields := make([]string, 0, 8+len(item.Tags)+len(item.Catalogs))
	fields = append(fields,
		item.ID,
		item.BotID,
		item.DiaryID,
		item.Title,
		item.Content,
		strconv.Itoa(item.VisibilityLevel),
		strconv.Itoa(item.Likes),
		item.CreatedAt,
		item.UpdatedAt,
	)
	fields = append(fields, item.Tags...)
	fields = append(fields, item.Catalogs...)
	return normalizeSearchText(strings.Join(fields, " "))
}

func trimOptionalString(input *string) *string {
	if input == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func runtimeInsightsUnsupportedMessage() string {
	return "runtime insights endpoint is unavailable on current server (404). Upgrade backend to version supporting /api/v1/runtime/insights."
}

func isRuntimeInsightsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if msg == "" {
		return false
	}
	if !strings.Contains(msg, "insight") {
		return false
	}
	return strings.Contains(msg, "status 404")
}

func normalizeRuntimeInsightsError(err error) error {
	if isRuntimeInsightsNotFoundError(err) {
		return errors.New(runtimeInsightsUnsupportedMessage())
	}
	return err
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

	if err := s.enrichDiaryItems([]*diarySummary{&item}); err != nil {
		return diaryDetail{}, false, err
	}

	return diaryDetail{diarySummary: item, Content: string(data)}, true, nil
}

func (s *Server) saveDiaryContent(id, content string) (diaryDetail, bool, error) {
	item, found, err := s.getDiaryByID(id)
	if err != nil {
		return diaryDetail{}, false, err
	}
	if !found {
		return diaryDetail{}, false, nil
	}

	path := filepath.Join(s.diaryDir, item.RelPath)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return diaryDetail{}, false, fmt.Errorf("write diary file: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return diaryDetail{}, false, fmt.Errorf("stat diary file: %w", err)
	}

	base := strings.TrimSuffix(item.Filename, ".md")
	contentBytes := []byte(content)
	updated := item
	updated.Date = detectDiaryDate(base, contentBytes)
	updated.Title, updated.Preview = extractTitleAndPreview(contentBytes)
	updated.Size = info.Size()
	updated.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	updated.SearchText = normalizeSearchText(content)

	_, err = s.db.Exec(`
UPDATE diary_entries
SET date = ?, title = ?, preview = ?, content_text = ?, size = ?, modified_at = ?, indexed_at = ?
WHERE id = ?
`, updated.Date, updated.Title, updated.Preview, updated.SearchText, updated.Size, updated.ModifiedAt, time.Now().UTC().Format(time.RFC3339), updated.ID)
	if err != nil {
		return diaryDetail{}, false, fmt.Errorf("update diary index row: %w", err)
	}

	if err := s.reconcileDayDefaults(); err != nil {
		return diaryDetail{}, false, err
	}
	if err := s.enrichDiaryItems([]*diarySummary{&updated}); err != nil {
		return diaryDetail{}, false, err
	}

	return diaryDetail{diarySummary: updated, Content: content}, true, nil
}

func (s *Server) listDiaries(q string, limit, offset int) ([]diarySummary, int, error) {
	whereSQL := ""
	args := make([]any, 0, 8)
	if strings.TrimSpace(q) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		whereSQL = ` WHERE lower(title) LIKE ? OR lower(preview) LIKE ? OR lower(filename) LIKE ? OR lower(date) LIKE ? OR content_text LIKE ?`
		args = append(args, like, like, like, like, like)
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

	ptrs := make([]*diarySummary, 0, len(items))
	for i := range items {
		ptrs = append(ptrs, &items[i])
	}
	if err := s.enrichDiaryItems(ptrs); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Server) listDiaryHistory() ([]diaryHistoryItem, error) {
	rows, err := s.db.Query(`
SELECT e.date, COUNT(1) AS diary_count, MAX(e.modified_at) AS latest_modified_at,
       COALESCE(d.diary_id, '') AS default_diary_id,
       COALESCE(d.is_manual, 0) AS default_is_manual
FROM diary_entries e
LEFT JOIN diary_day_defaults d ON d.diary_date = e.date
WHERE e.date <> ''
GROUP BY e.date, d.diary_id, d.is_manual
ORDER BY e.date DESC
`)
	if err != nil {
		return nil, fmt.Errorf("query diary history: %w", err)
	}
	defer rows.Close()

	items := make([]diaryHistoryItem, 0, 128)
	for rows.Next() {
		var item diaryHistoryItem
		var defaultIsManualInt int
		if err := rows.Scan(&item.Date, &item.DiaryCount, &item.LatestModifiedAt, &item.DefaultDiaryID, &defaultIsManualInt); err != nil {
			return nil, fmt.Errorf("scan diary history row: %w", err)
		}
		item.HasDefault = strings.TrimSpace(item.DefaultDiaryID) != ""
		item.DefaultIsManual = defaultIsManualInt == 1
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read diary history rows: %w", err)
	}

	return items, nil
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
INSERT INTO diary_entries(id, rel_path, filename, date, title, preview, content_text, size, modified_at, indexed_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return 0, fmt.Errorf("prepare diary insert: %w", err)
	}
	defer stmt.Close()

	indexedAt := time.Now().UTC().Format(time.RFC3339)
	for _, item := range items {
		_, err := stmt.Exec(item.ID, item.RelPath, item.Filename, item.Date, item.Title, item.Preview, item.SearchText, item.Size, item.ModifiedAt, indexedAt)
		if err != nil {
			return 0, fmt.Errorf("insert diary index row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit reindex tx: %w", err)
	}
	if err := s.reconcileDayDefaults(); err != nil {
		return 0, err
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
			SearchText: normalizeSearchText(string(data)),
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

func (s *Server) setDiaryAsDefault(id string) (diaryDetail, bool, error) {
	item, found, err := s.getDiaryByID(id)
	if err != nil {
		return diaryDetail{}, false, err
	}
	if !found {
		return diaryDetail{}, false, nil
	}
	if strings.TrimSpace(item.Date) == "" {
		return diaryDetail{}, true, errors.New("diary date is empty, cannot set default")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`
INSERT INTO diary_day_defaults(diary_date, diary_id, is_manual, updated_at)
VALUES(?, ?, 1, ?)
ON CONFLICT(diary_date) DO UPDATE SET diary_id = excluded.diary_id, is_manual = excluded.is_manual, updated_at = excluded.updated_at
`, item.Date, item.ID, now)
	if err != nil {
		return diaryDetail{}, true, fmt.Errorf("set day default diary: %w", err)
	}

	return s.loadDiaryDetail(id)
}

func (s *Server) syncDiaryByID(id string) (diarySyncResponse, bool, error) {
	diag := s.newSyncDiagContext(id)

	detail, found, err := s.loadDiaryDetail(id)
	if err != nil {
		s.logSyncFailure("load_diary_detail", diag, err)
		return diarySyncResponse{}, false, err
	}
	if !found {
		return diarySyncResponse{}, false, nil
	}
	diag.DiaryDate = detail.Date
	diag.DiaryTitle = detail.Title
	diag.DiaryFilename = detail.Filename
	diag.DiaryRelPath = detail.RelPath
	diag.DiaryPath = filepath.Join(s.diaryDir, detail.RelPath)
	isDefault := detail.IsDefault
	diag.IsDefault = &isDefault

	if strings.TrimSpace(detail.Date) == "" {
		err = errors.New("diary date is required for sync")
		s.logSyncBlocked("validate_diary_date", diag, err)
		return diarySyncResponse{}, true, err
	}
	if !detail.IsDefault {
		err = errors.New("sync blocked: this diary is not the day default, use 'Set Default' first")
		s.logSyncBlocked("validate_day_default", diag, err)
		return diarySyncResponse{}, true, err
	}

	cloudSyncEnabled, err := s.getCloudSyncEnabled()
	if err != nil {
		s.logSyncFailure("read_cloud_sync_setting", diag, err)
		return diarySyncResponse{}, true, err
	}
	diag.CloudSyncEnabled = &cloudSyncEnabled
	if !cloudSyncEnabled {
		err = errors.New("sync blocked: cloud sync is disabled in Settings")
		s.logSyncBlocked("precheck_cloud_sync", diag, err)
		return diarySyncResponse{}, true, err
	}

	apiKeyConfigured, _, keySource, err := s.resolveAPIKeyState()
	if err != nil {
		s.logSyncFailure("resolve_api_key_state", diag, err)
		return diarySyncResponse{}, true, err
	}
	diag.APIKeyConfigured = &apiKeyConfigured
	diag.APIKeySource = keySource
	if !apiKeyConfigured {
		err = errors.New("sync blocked: API key is not configured. Set it in Settings or run `moltbb login --apikey <key>`")
		s.logSyncBlocked("precheck_api_key", diag, err)
		return diarySyncResponse{}, true, err
	}

	filePath := filepath.Join(s.diaryDir, detail.RelPath)
	diag.DiaryPath = filePath
	payload, err := diary.BuildRuntimeUpsertPayload(filePath, detail.Date, 0, time.Now().UTC())
	if err != nil {
		s.logSyncFailure("build_runtime_payload", diag, err)
		return diarySyncResponse{}, true, err
	}

	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		err = fmt.Errorf("resolve api key from %s: %w", keySource, err)
		s.logSyncFailure("resolve_api_key", diag, err)
		return diarySyncResponse{}, true, err
	}

	cfg := config.Default()
	base := strings.TrimSpace(s.apiBaseURL)
	if base == "" {
		base = config.DefaultAPIBaseURL
	}
	diag.APIBaseURL = base
	cfg.APIBaseURL = base
	if strings.HasPrefix(base, "http://") {
		cfg.AllowInsecureHTTP = true
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		s.logSyncFailure("create_api_client", diag, err)
		return diarySyncResponse{}, true, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	result, err := client.UpsertRuntimeDiary(ctx, apiKey, api.RuntimeDiaryUpsertPayload{
		Summary:        payload.Summary,
		PersonaText:    payload.PersonaText,
		ExecutionLevel: payload.ExecutionLevel,
		DiaryDate:      payload.DiaryDate,
	})
	if err != nil {
		s.logSyncFailure("upsert_runtime_diary", diag, err)
		return diarySyncResponse{}, true, err
	}

	s.logSyncSuccess("upsert_runtime_diary", diag, result)

	return diarySyncResponse{
		Success:    true,
		DiaryID:    result.DiaryID,
		Action:     result.Action,
		StatusCode: result.StatusCode,
	}, true, nil
}

func (s *Server) newSyncDiagContext(id string) syncDiagContext {
	baseURL := strings.TrimSpace(s.apiBaseURL)
	if baseURL == "" {
		baseURL = config.DefaultAPIBaseURL
	}
	return syncDiagContext{
		DiaryID:    strings.TrimSpace(id),
		APIBaseURL: baseURL,
	}
}

func (s *Server) logSyncBlocked(stage string, diag syncDiagContext, err error) {
	s.appendSyncLog(syncLogEntry{
		Level:            "warn",
		Event:            "diary_sync_blocked",
		Stage:            stage,
		DiaryID:          diag.DiaryID,
		DiaryDate:        diag.DiaryDate,
		DiaryTitle:       diag.DiaryTitle,
		DiaryFilename:    diag.DiaryFilename,
		DiaryRelPath:     diag.DiaryRelPath,
		DiaryPath:        diag.DiaryPath,
		IsDefault:        diag.IsDefault,
		CloudSyncEnabled: diag.CloudSyncEnabled,
		APIKeyConfigured: diag.APIKeyConfigured,
		APIKeySource:     diag.APIKeySource,
		APIBaseURL:       diag.APIBaseURL,
		Message:          "sync blocked by local precondition",
		Error:            err.Error(),
	})
}

func (s *Server) logSyncFailure(stage string, diag syncDiagContext, err error) {
	s.appendSyncLog(syncLogEntry{
		Level:            "error",
		Event:            "diary_sync_failed",
		Stage:            stage,
		DiaryID:          diag.DiaryID,
		DiaryDate:        diag.DiaryDate,
		DiaryTitle:       diag.DiaryTitle,
		DiaryFilename:    diag.DiaryFilename,
		DiaryRelPath:     diag.DiaryRelPath,
		DiaryPath:        diag.DiaryPath,
		IsDefault:        diag.IsDefault,
		CloudSyncEnabled: diag.CloudSyncEnabled,
		APIKeyConfigured: diag.APIKeyConfigured,
		APIKeySource:     diag.APIKeySource,
		APIBaseURL:       diag.APIBaseURL,
		Message:          "sync request failed",
		Error:            err.Error(),
	})
}

func (s *Server) logSyncSuccess(stage string, diag syncDiagContext, result api.RuntimeDiaryUpsertResult) {
	s.appendSyncLog(syncLogEntry{
		Level:            "info",
		Event:            "diary_sync_succeeded",
		Stage:            stage,
		DiaryID:          diag.DiaryID,
		DiaryDate:        diag.DiaryDate,
		DiaryTitle:       diag.DiaryTitle,
		DiaryFilename:    diag.DiaryFilename,
		DiaryRelPath:     diag.DiaryRelPath,
		DiaryPath:        diag.DiaryPath,
		IsDefault:        diag.IsDefault,
		CloudSyncEnabled: diag.CloudSyncEnabled,
		APIKeyConfigured: diag.APIKeyConfigured,
		APIKeySource:     diag.APIKeySource,
		APIBaseURL:       diag.APIBaseURL,
		Action:           result.Action,
		StatusCode:       result.StatusCode,
		RemoteDiaryID:    result.DiaryID,
		Message:          "sync request completed",
	})
}

func (s *Server) appendSyncLog(entry syncLogEntry) {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	if strings.TrimSpace(entry.Level) == "" {
		entry.Level = "info"
	}
	if strings.TrimSpace(entry.Event) == "" {
		entry.Event = "diary_sync_event"
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: encode sync diagnostics log failed: %v\n", err)
		return
	}

	logPath := filepath.Join(s.dataDir, syncDiagnosticsLogFileName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: open sync diagnostics log failed: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "warning: write sync diagnostics log failed: %v\n", err)
	}
}

func (s *Server) reconcileDayDefaults() error {
	rows, err := s.db.Query(`
SELECT id, date
FROM diary_entries
WHERE date <> ''
ORDER BY date ASC, modified_at DESC
`)
	if err != nil {
		return fmt.Errorf("query diary entries for day defaults: %w", err)
	}
	defer rows.Close()

	dayItems := make(map[string][]string)
	daySet := make(map[string]map[string]struct{})
	for rows.Next() {
		var id, date string
		if err := rows.Scan(&id, &date); err != nil {
			return fmt.Errorf("scan diary entry for day defaults: %w", err)
		}
		if strings.TrimSpace(date) == "" {
			continue
		}
		dayItems[date] = append(dayItems[date], id)
		set, ok := daySet[date]
		if !ok {
			set = make(map[string]struct{})
			daySet[date] = set
		}
		set[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read diary entries for day defaults: %w", err)
	}

	existing, err := s.loadDayDefaultRecords()
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin day defaults reconcile tx: %w", err)
	}
	defer tx.Rollback()

	for date, rec := range existing {
		ids, ok := daySet[date]
		if !ok {
			if _, err := tx.Exec(`DELETE FROM diary_day_defaults WHERE diary_date = ?`, date); err != nil {
				return fmt.Errorf("delete stale day default: %w", err)
			}
			continue
		}
		if _, ok := ids[rec.DiaryID]; !ok {
			// stale pointer, will be reset by upsert below.
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for date, ids := range dayItems {
		selectedID := ""
		isManual := 0
		if rec, ok := existing[date]; ok {
			if rec.IsManual {
				if _, exists := daySet[date][rec.DiaryID]; exists {
					selectedID = rec.DiaryID
					isManual = 1
				}
			}
		}
		if selectedID == "" && len(ids) > 0 {
			// ids are already sorted by modified_at desc within same day.
			selectedID = ids[0]
			isManual = 0
		}
		if selectedID == "" {
			continue
		}
		if _, err := tx.Exec(`
INSERT INTO diary_day_defaults(diary_date, diary_id, is_manual, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(diary_date) DO UPDATE SET diary_id = excluded.diary_id, is_manual = excluded.is_manual, updated_at = excluded.updated_at
`, date, selectedID, isManual, now); err != nil {
			return fmt.Errorf("upsert day default: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit day defaults reconcile tx: %w", err)
	}
	return nil
}

func (s *Server) loadDayDefaultRecords() (map[string]dayDefaultRecord, error) {
	rows, err := s.db.Query(`SELECT diary_date, diary_id, is_manual FROM diary_day_defaults`)
	if err != nil {
		return nil, fmt.Errorf("query day defaults: %w", err)
	}
	defer rows.Close()

	out := make(map[string]dayDefaultRecord)
	for rows.Next() {
		var (
			date     string
			diaryID  string
			isManual int
		)
		if err := rows.Scan(&date, &diaryID, &isManual); err != nil {
			return nil, fmt.Errorf("scan day default row: %w", err)
		}
		out[date] = dayDefaultRecord{
			DiaryID:  diaryID,
			IsManual: isManual == 1,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read day default rows: %w", err)
	}
	return out, nil
}

func (s *Server) enrichDiaryItems(items []*diarySummary) error {
	if len(items) == 0 {
		return nil
	}
	if err := s.reconcileDayDefaults(); err != nil {
		return err
	}
	dayDefaults, err := s.loadDayDefaultRecords()
	if err != nil {
		return err
	}

	cloudSyncEnabled, err := s.getCloudSyncEnabled()
	if err != nil {
		return err
	}
	apiKeyConfigured, _, _, err := s.resolveAPIKeyState()
	if err != nil {
		return err
	}
	canSync := cloudSyncEnabled && apiKeyConfigured

	for _, item := range items {
		if item == nil {
			continue
		}
		item.IsDefault = false
		item.CanSync = false
		date := strings.TrimSpace(item.Date)
		if date == "" {
			continue
		}
		if rec, ok := dayDefaults[date]; ok && strings.TrimSpace(rec.DiaryID) == item.ID {
			item.IsDefault = true
			item.CanSync = canSync
		}
	}
	return nil
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

func normalizeSearchText(content string) string {
	fields := strings.Fields(strings.ToLower(content))
	if len(fields) == 0 {
		return ""
	}
	text := strings.Join(fields, " ")
	if len(text) > 120000 {
		return text[:120000]
	}
	return text
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

func splitCommaList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		for _, part := range strings.Split(item, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeStringListValues(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, item := range in {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func rewritePrefixedPath(path string) (string, bool) {
	switch path {
	case "", "/", "/index.html", "/styles.css", "/app.js", "/icon.png", "/pure-logo.png":
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
	if strings.HasSuffix(path, "/icon.png") {
		return "/icon.png", true
	}
	if strings.HasSuffix(path, "/pure-logo.png") {
		return "/pure-logo.png", true
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

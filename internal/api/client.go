package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moltbb-cli/internal/config"

	_ "modernc.org/sqlite"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	retryCount int
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type ValidateResponse struct {
	Valid         bool   `json:"valid"`
	Token         string `json:"token,omitempty"`
	OwnerID       string `json:"owner_id,omitempty"`
	OwnerNickname string `json:"owner_nickname,omitempty"`
}

type BindRequest struct {
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	Version     string `json:"version"`
	Fingerprint string `json:"fingerprint"`
}

type BindResponse struct {
	BotID            string `json:"bot_id"`
	ActivationStatus string `json:"activation_status"`
}

type RuntimeDiaryUpsertPayload struct {
	Summary        string `json:"summary"`
	PersonaText    string `json:"personaText,omitempty"`
	ExecutionLevel int    `json:"executionLevel"`
	DiaryDate      string `json:"diaryDate"`
}

type RuntimeDiary struct {
	ID              string `json:"id"`
	DiaryDate       string `json:"diaryDate,omitempty"`
	Date            string `json:"date,omitempty"`
	Summary         string `json:"summary,omitempty"`
	PersonaText     string `json:"personaText,omitempty"`
	ExecutionLevel  int    `json:"executionLevel,omitempty"`
	VisibilityLevel int    `json:"visibilityLevel,omitempty"`
}

type RuntimeDiaryListResult struct {
	Items      []RuntimeDiary
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

type RuntimeDiaryUpsertResult struct {
	Action     string `json:"action"`
	DiaryID    string `json:"diaryId,omitempty"`
	StatusCode int    `json:"statusCode"`
}

type RuntimeDiaryPatchPayload struct {
	Summary *string `json:"summary,omitempty"`
	Content *string `json:"content,omitempty"`
}

type RuntimeInsightCreatePayload struct {
	Title           string   `json:"title"`
	DiaryID         string   `json:"diaryId,omitempty"`
	Catalogs        []string `json:"catalogs,omitempty"`
	Content         string   `json:"content"`
	Tags            []string `json:"tags,omitempty"`
	VisibilityLevel int      `json:"visibilityLevel,omitempty"`
}

type RuntimeInsightUpdatePayload struct {
	Title           *string  `json:"title,omitempty"`
	Catalogs        []string `json:"catalogs,omitempty"`
	Content         *string  `json:"content,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	VisibilityLevel *int     `json:"visibilityLevel,omitempty"`
}

type RuntimeInsight struct {
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
}

type RuntimeInsightListResult struct {
	Items      []RuntimeInsight
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

func NewClient(cfg config.Config) (*Client, error) {
	if strings.HasPrefix(cfg.APIBaseURL, "http://") && !cfg.AllowInsecureHTTP {
		return nil, fmt.Errorf("api endpoint must use https unless allow_insecure_http is enabled: %s", cfg.APIBaseURL)
	}
	if !strings.HasPrefix(cfg.APIBaseURL, "https://") && !strings.HasPrefix(cfg.APIBaseURL, "http://") {
		return nil, fmt.Errorf("api endpoint must be http(s): %s", cfg.APIBaseURL)
	}
	_, err := url.ParseRequestURI(cfg.APIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid api endpoint: %w", err)
	}

	return &Client{
		baseURL: strings.TrimRight(cfg.APIBaseURL, "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeoutSeconds) * time.Second,
		},
		retryCount: cfg.RetryCount,
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	paths := []string{"/health", "/api/v1/runtime/capabilities"}
	for _, p := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+p, nil)
		if err != nil {
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 500 {
			return nil
		}
	}
	return errors.New("unable to reach MoltBB API")
}

func (c *Client) ValidateAPIKey(ctx context.Context, apiKey string) (ValidateResponse, error) {
	payload := map[string]string{"api_key": apiKey}
	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/auth/validate", apiKey, payload)
	if err != nil {
		return ValidateResponse{}, err
	}

	// Some deployments removed /auth/validate. Fallback to other auth-protected endpoints.
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		// 1) MoltBB runtime insights (primary for moltbb backend)
		body, status, err = c.doJSONWithAPIKey(ctx, http.MethodGet, "/api/v1/runtime/insights", apiKey, nil)
		if err == nil && status >= 200 && status < 300 {
			return ValidateResponse{Valid: true}, nil
		}
		// 2) Moltbook agents/me (legacy for moltbook backend)
		body, status, err = c.doJSONWithAPIKey(ctx, http.MethodGet, "/api/v1/agents/me", apiKey, nil)
		if err != nil {
			return ValidateResponse{}, err
		}
		if status >= 200 && status < 300 {
			return ValidateResponse{Valid: true}, nil
		}
		return ValidateResponse{}, fmt.Errorf("validate failed with status %d: %s", status, string(body))
	}

	if status < 200 || status >= 300 {
		return ValidateResponse{}, fmt.Errorf("validate failed with status %d: %s", status, string(body))
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		var resp ValidateResponse
		if err := json.Unmarshal(env.Data, &resp); err == nil {
			if !resp.Valid && env.Success {
				resp.Valid = true
			}
			return resp, nil
		}
	}

	// Fallback for non-enveloped APIs.
	var resp ValidateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ValidateResponse{}, fmt.Errorf("parse validate response: %w", err)
	}
	if !resp.Valid {
		resp.Valid = true
	}
	return resp, nil
}

func (c *Client) BindBot(ctx context.Context, apiKey string, req BindRequest) (BindResponse, error) {
	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/bot/bind", apiKey, req)
	if err != nil {
		return BindResponse{}, err
	}

	if status >= 200 && status < 300 {
		return decodeBindResponse(body)
	}

	// Some deployments accept only API key on this endpoint.
	bodyNoPayload, statusNoPayload, errNoPayload := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/bot/bind", apiKey, map[string]any{})
	if errNoPayload == nil && statusNoPayload >= 200 && statusNoPayload < 300 {
		return decodeBindResponse(bodyNoPayload)
	}

	legacyFallback := strings.EqualFold(strings.TrimSpace(os.Getenv("MOLTBB_LEGACY_RUNTIME_BIND")), "1") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("MOLTBB_LEGACY_RUNTIME_BIND")), "true")

	if legacyFallback && (status == http.StatusNotFound || statusNoPayload == http.StatusNotFound) {
		// Compatibility with runtime API currently available on backend.
		body, status, err = c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/runtime/activate", apiKey, req)
		if err != nil {
			return BindResponse{}, err
		}
	}
	if status < 200 || status >= 300 {
		if errNoPayload != nil {
			return BindResponse{}, fmt.Errorf("bind failed with status %d: %s (no-payload retry error: %v)", status, string(body), errNoPayload)
		}
		if statusNoPayload >= 200 && statusNoPayload < 300 {
			return decodeBindResponse(bodyNoPayload)
		}
		if statusNoPayload >= 300 {
			return BindResponse{}, fmt.Errorf("bind failed with status %d: %s (no-payload retry status %d: %s)", status, string(body), statusNoPayload, string(bodyNoPayload))
		}
		return BindResponse{}, fmt.Errorf("bind failed with status %d: %s", status, string(body))
	}

	return decodeBindResponse(body)
}

func decodeBindResponse(body []byte) (BindResponse, error) {
	parse := func(raw map[string]any) BindResponse {
		var resp BindResponse
		if v, ok := raw["bot_id"].(string); ok {
			resp.BotID = v
		}
		if resp.BotID == "" {
			if v, ok := raw["botId"].(string); ok {
				resp.BotID = v
			}
		}
		if v, ok := raw["activation_status"].(string); ok {
			resp.ActivationStatus = v
		}
		if resp.ActivationStatus == "" {
			if v, ok := raw["activationStatus"].(string); ok {
				resp.ActivationStatus = v
			}
		}
		return resp
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		var mapped map[string]any
		if err := json.Unmarshal(env.Data, &mapped); err == nil {
			resp := parse(mapped)
			if resp.ActivationStatus == "" {
				resp.ActivationStatus = "active"
			}
			if resp.BotID != "" {
				return resp, nil
			}
		}
	}

	// runtime/activate currently returns bot information.
	var mapped map[string]any
	if err := json.Unmarshal(body, &mapped); err == nil {
		resp := parse(mapped)
		if resp.BotID != "" {
			if resp.ActivationStatus == "" {
				resp.ActivationStatus = "active"
			}
			return resp, nil
		}
	}

	return BindResponse{}, fmt.Errorf("unable to parse bind response")
}

func (c *Client) UpsertRuntimeDiary(ctx context.Context, apiKey string, payload RuntimeDiaryUpsertPayload) (RuntimeDiaryUpsertResult, error) {
	diaryDate := strings.TrimSpace(payload.DiaryDate)
	if diaryDate == "" {
		return RuntimeDiaryUpsertResult{}, errors.New("diaryDate is required")
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return RuntimeDiaryUpsertResult{}, errors.New("summary is required")
	}

	// Save to local database first
	if err := saveToLocalDB(diaryDate, payload.Summary); err != nil {
		fmt.Printf("⚠️  Warning: failed to save to local DB: %v\n", err)
	}

	existingID, err := c.findRuntimeDiaryIDByDate(ctx, apiKey, diaryDate)
	if err != nil {
		return RuntimeDiaryUpsertResult{}, err
	}
	if existingID != "" {
		status, err := c.patchRuntimeDiary(ctx, apiKey, existingID, payload.Summary, payload.PersonaText)
		if err != nil {
			return RuntimeDiaryUpsertResult{}, err
		}
		return RuntimeDiaryUpsertResult{
			Action:     "PATCH",
			DiaryID:    existingID,
			StatusCode: status,
		}, nil
	}

	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/runtime/diaries", apiKey, payload)
	if err != nil {
		return RuntimeDiaryUpsertResult{}, err
	}
	if status < 200 || status >= 300 {
		return RuntimeDiaryUpsertResult{}, fmt.Errorf("upload diary failed with status %d: %s", status, string(body))
	}

	if diaryID, ok := parseCreatedDiaryID(body); ok {
		return RuntimeDiaryUpsertResult{
			Action:     "POST",
			DiaryID:    diaryID,
			StatusCode: status,
		}, nil
	}

	if diaryID, conflict := parseDuplicateDiaryID(body); conflict {
		if diaryID == "" {
			diaryID, err = c.findRuntimeDiaryIDByDate(ctx, apiKey, diaryDate)
			if err != nil {
				return RuntimeDiaryUpsertResult{}, err
			}
		}
		if diaryID == "" {
			return RuntimeDiaryUpsertResult{}, errors.New("diary already exists but diary id is missing")
		}
		patchStatus, err := c.patchRuntimeDiary(ctx, apiKey, diaryID, payload.Summary, payload.PersonaText)
		if err != nil {
			return RuntimeDiaryUpsertResult{}, err
		}
		return RuntimeDiaryUpsertResult{
			Action:     "PATCH_AFTER_CONFLICT",
			DiaryID:    diaryID,
			StatusCode: patchStatus,
		}, nil
	}

	return RuntimeDiaryUpsertResult{
		Action:     "POST",
		StatusCode: status,
	}, nil
}

func (c *Client) PatchRuntimeDiary(ctx context.Context, apiKey, diaryID string, payload RuntimeDiaryPatchPayload) error {
	id := strings.TrimSpace(diaryID)
	if id == "" {
		return errors.New("diary id is required")
	}
	if payload.Summary == nil && payload.Content == nil {
		return errors.New("at least one field is required for diary patch")
	}

	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPatch, "/api/v1/runtime/diaries/"+id, apiKey, payload)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("patch diary failed with status %d: %s", status, string(body))
	}
	return nil
}

func (c *Client) CreateRuntimeInsight(ctx context.Context, apiKey string, payload RuntimeInsightCreatePayload) (RuntimeInsight, error) {
	if strings.TrimSpace(payload.Title) == "" {
		return RuntimeInsight{}, errors.New("title is required")
	}
	if strings.TrimSpace(payload.Content) == "" {
		return RuntimeInsight{}, errors.New("content is required")
	}
	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/runtime/insights", apiKey, payload)
	if err != nil {
		return RuntimeInsight{}, err
	}
	if status < 200 || status >= 300 {
		return RuntimeInsight{}, fmt.Errorf("upload insight failed with status %d: %s", status, string(body))
	}
	var insight RuntimeInsight
	if err := decodeEnvelopeData(body, &insight); err != nil {
		return RuntimeInsight{}, fmt.Errorf("parse insight response: %w", err)
	}
	return insight, nil
}

func (c *Client) UpdateRuntimeInsight(ctx context.Context, apiKey, insightID string, payload RuntimeInsightUpdatePayload) (RuntimeInsight, error) {
	id := strings.TrimSpace(insightID)
	if id == "" {
		return RuntimeInsight{}, errors.New("insight id is required")
	}
	hasAnyField := payload.Title != nil || payload.Content != nil || payload.VisibilityLevel != nil ||
		len(payload.Catalogs) > 0 || len(payload.Tags) > 0
	if !hasAnyField {
		return RuntimeInsight{}, errors.New("at least one field is required for insight update")
	}

	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPatch, "/api/v1/runtime/insights/"+id, apiKey, payload)
	if err != nil {
		return RuntimeInsight{}, err
	}
	if status < 200 || status >= 300 {
		return RuntimeInsight{}, fmt.Errorf("update insight failed with status %d: %s", status, string(body))
	}
	var insight RuntimeInsight
	if err := decodeEnvelopeData(body, &insight); err != nil {
		return RuntimeInsight{}, fmt.Errorf("parse insight response: %w", err)
	}
	return insight, nil
}

func (c *Client) DeleteRuntimeInsight(ctx context.Context, apiKey, insightID string) error {
	id := strings.TrimSpace(insightID)
	if id == "" {
		return errors.New("insight id is required")
	}
	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodDelete, "/api/v1/runtime/insights/"+id, apiKey, nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("delete insight failed with status %d: %s", status, string(body))
	}
	return nil
}

func (c *Client) ListRuntimeInsights(
	ctx context.Context,
	apiKey string,
	page, pageSize int,
	tags []string,
	diaryID string,
) (RuntimeInsightListResult, error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			query.Add("tags", trimmed)
		}
	}
	if trimmedDiaryID := strings.TrimSpace(diaryID); trimmedDiaryID != "" {
		query.Set("diaryId", trimmedDiaryID)
	}

	path := "/api/v1/runtime/insights"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, path, apiKey, nil)
	if err != nil {
		return RuntimeInsightListResult{}, err
	}
	if status < 200 || status >= 300 {
		return RuntimeInsightListResult{}, fmt.Errorf("list insights failed with status %d: %s", status, string(body))
	}

	var raw struct {
		Success    bool             `json:"success"`
		Data       []RuntimeInsight `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			TotalCount int `json:"totalCount"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return RuntimeInsightListResult{}, fmt.Errorf("parse list insights response: %w", err)
	}

	return RuntimeInsightListResult{
		Items:      raw.Data,
		Page:       raw.Pagination.Page,
		PageSize:   raw.Pagination.PageSize,
		TotalCount: raw.Pagination.TotalCount,
		TotalPages: raw.Pagination.TotalPages,
	}, nil
}

func (c *Client) ListRuntimeDiaries(ctx context.Context, apiKey, startDate, endDate string, page, pageSize int) (RuntimeDiaryListResult, error) {
	query := url.Values{}
	if strings.TrimSpace(startDate) != "" {
		query.Set("startDate", startDate)
	}
	if strings.TrimSpace(endDate) != "" {
		query.Set("endDate", endDate)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))

	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/runtime/diaries?"+query.Encode(), apiKey, nil)
	if err != nil {
		return RuntimeDiaryListResult{}, err
	}
	if status < 200 || status >= 300 {
		return RuntimeDiaryListResult{}, fmt.Errorf("list runtime diaries failed with status %d: %s", status, string(body))
	}

	var raw struct {
		Success    bool           `json:"success"`
		Data       []RuntimeDiary `json:"data"`
		Items      []RuntimeDiary `json:"items"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			TotalCount int `json:"totalCount"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		// Some deployments return envelope or bare array.
		if err2 := decodeEnvelopeData(body, &raw); err2 != nil {
			var items []RuntimeDiary
			if err3 := json.Unmarshal(body, &items); err3 != nil {
				return RuntimeDiaryListResult{}, fmt.Errorf("parse list runtime diaries response: %w", err)
			}
			return RuntimeDiaryListResult{Items: items}, nil
		}
	}

	items := raw.Data
	if len(items) == 0 {
		items = raw.Items
	}

	return RuntimeDiaryListResult{
		Items:      items,
		Page:       raw.Pagination.Page,
		PageSize:   raw.Pagination.PageSize,
		TotalCount: raw.Pagination.TotalCount,
		TotalPages: raw.Pagination.TotalPages,
	}, nil
}

func (c *Client) findRuntimeDiaryIDByDate(ctx context.Context, apiKey, diaryDate string) (string, error) {
	query := url.Values{}
	query.Set("startDate", diaryDate)
	query.Set("endDate", diaryDate)
	query.Set("page", "1")
	query.Set("pageSize", "1")

	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/runtime/diaries?"+query.Encode(), apiKey, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("query runtime diaries failed with status %d: %s", status, string(body))
	}

	return parseFirstDiaryID(body), nil
}

func (c *Client) patchRuntimeDiary(ctx context.Context, apiKey, diaryID, summary, personaText string) (int, error) {
	payload := map[string]string{
		"summary":     summary,
		"personaText": personaText,
	}
	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPatch, "/api/v1/runtime/diaries/"+diaryID, apiKey, payload)
	if err != nil {
		return 0, err
	}
	if status < 200 || status >= 300 {
		return status, fmt.Errorf("patch diary failed with status %d: %s", status, string(body))
	}
	return status, nil
}

func parseCreatedDiaryID(body []byte) (string, bool) {
	if id := parseFirstDiaryID(body); id != "" {
		return id, true
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", false
	}
	success, ok := raw["success"].(bool)
	return "", ok && success
}

func parseDuplicateDiaryID(body []byte) (string, bool) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", false
	}
	success, _ := raw["success"].(bool)
	if success {
		return "", false
	}
	code, _ := raw["code"].(string)
	if code != "DIARY_ALREADY_EXISTS_USE_PATCH" {
		return "", false
	}
	details, _ := raw["details"].(map[string]any)
	if details == nil {
		return "", true
	}
	id, _ := details["diaryId"].(string)
	return strings.TrimSpace(id), true
}

func parseFirstDiaryID(body []byte) string {
	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		var data any
		if err := json.Unmarshal(env.Data, &data); err == nil {
			if id := extractDiaryID(data); id != "" {
				return id
			}
		}
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}
	if id := extractDiaryID(raw); id != "" {
		return id
	}
	if data, ok := raw["data"]; ok {
		return extractDiaryID(data)
	}
	return ""
}

func extractDiaryID(value any) string {
	switch v := value.(type) {
	case map[string]any:
		if id, ok := v["id"].(string); ok {
			return strings.TrimSpace(id)
		}
		if data, ok := v["data"]; ok {
			if id := extractDiaryID(data); id != "" {
				return id
			}
		}
		if items, ok := v["items"]; ok {
			if id := extractDiaryID(items); id != "" {
				return id
			}
		}
		return ""
	case []any:
		if len(v) == 0 {
			return ""
		}
		return extractDiaryID(v[0])
	default:
		return ""
	}
}

func decodeEnvelopeData(body []byte, dest any) error {
	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		return json.Unmarshal(env.Data, dest)
	}
	return json.Unmarshal(body, dest)
}

func (c *Client) doJSONWithAPIKey(ctx context.Context, method, path, apiKey string, payload any) ([]byte, int, error) {
	return c.doRequestWithAPIKey(ctx, method, path, apiKey, payload)
}

func (c *Client) doRequestWithAPIKey(ctx context.Context, method, path, apiKey string, payload any) ([]byte, int, error) {
	var data []byte
	var err error
	if payload != nil {
		data, err = json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
	}

	var lastErr error
	maxAttempts := c.retryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if strings.TrimSpace(apiKey) != "" {
			req.Header.Set("X-API-Key", apiKey)
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
			} else if resp.StatusCode >= 500 && attempt < maxAttempts {
				lastErr = fmt.Errorf("server returned %d", resp.StatusCode)
			} else {
				return body, resp.StatusCode, nil
			}
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
		}
	}

	return nil, 0, fmt.Errorf("request failed after retries: %w", lastErr)
}

// Tower API types
type TowerCheckinResponse struct {
	Code        string `json:"code"`
	GlobalIndex int    `json:"globalIndex"`
	Floor       int    `json:"floor"`
	RoomNumber  int    `json:"roomNumber"`
	JoinTime    *int64 `json:"joinTime,omitempty"` // Unix timestamp, nullable
}

type TowerHeartbeatResponse struct {
	Success   bool  `json:"success"`
	Timestamp int64 `json:"timestamp"`
}

type TowerRoomState struct {
	Code          string  `json:"code"`
	GlobalIndex   int     `json:"globalIndex"`
	BotId         string  `json:"botId,omitempty"`
	BotName       string  `json:"botName,omitempty"`
	Status        int     `json:"status"`
	LastHeartbeat *int64  `json:"lastHeartbeat,omitempty"`
	StatusMessage *string `json:"statusMessage,omitempty"`
}

type TowerRoomDetail struct {
	Code            string  `json:"code"`
	Floor           int     `json:"floor"`
	RoomNumber      int     `json:"roomNumber"`
	GlobalIndex     int     `json:"globalIndex"`
	BotId           string  `json:"botId,omitempty"`
	BotName         string  `json:"botName,omitempty"`
	Status          int     `json:"status"`
	LastHeartbeat   *int64  `json:"lastHeartbeat,omitempty"`
	JoinTime        *int64  `json:"joinTime,omitempty"`
	TotalHeartbeats int     `json:"totalHeartbeats"`
	StatusMessage   *string `json:"statusMessage,omitempty"`
}

type TowerStatistics struct {
	TotalRooms       int     `json:"totalRooms"`
	OccupiedRooms    int     `json:"occupiedRooms"`
	OnlineNodes      int     `json:"onlineNodes"`
	OnlineRooms      int     `json:"onlineRooms"`
	Stable7DNodes    int     `json:"stable7DNodes"`
	Stable30DNodes   int     `json:"stable30DNodes"`
	OccupancyRate    float64 `json:"occupancyRate"`
	RoomsJoinedToday int     `json:"roomsJoinedToday"`
	FullFloors       int     `json:"fullFloors"`
	IsFullTower      bool    `json:"isFullTower"`
}

// TowerCheckin assigns an available room to the authenticated bot
func (c *Client) TowerCheckin(ctx context.Context, apiKey string, roomCode string) (TowerCheckinResponse, error) {
	var payload interface{}
	if strings.TrimSpace(roomCode) != "" {
		payload = map[string]string{"roomCode": roomCode}
	}

	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/tower/checkin", apiKey, payload)
	if err != nil {
		return TowerCheckinResponse{}, err
	}
	if status < 200 || status >= 300 {
		return TowerCheckinResponse{}, fmt.Errorf("tower checkin failed with status %d: %s", status, string(body))
	}
	var resp TowerCheckinResponse
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return TowerCheckinResponse{}, fmt.Errorf("parse tower checkin response: %w", err)
	}
	return resp, nil
}

// TowerSendHeartbeat sends a heartbeat for the specified room
func (c *Client) TowerSendHeartbeat(ctx context.Context, apiKey, roomCode string, statusMessage *string) (TowerHeartbeatResponse, error) {
	payload := map[string]interface{}{"roomCode": roomCode}
	if statusMessage != nil {
		payload["statusMessage"] = *statusMessage
	}
	body, status, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/tower/heartbeat", apiKey, payload)
	if err != nil {
		return TowerHeartbeatResponse{}, err
	}
	if status < 200 || status >= 300 {
		return TowerHeartbeatResponse{}, fmt.Errorf("tower heartbeat failed with status %d: %s", status, string(body))
	}
	var resp TowerHeartbeatResponse
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return TowerHeartbeatResponse{}, fmt.Errorf("parse tower heartbeat response: %w", err)
	}
	return resp, nil
}

// TowerGetMyRoom returns the authenticated bot's current room assignment
func (c *Client) TowerGetMyRoom(ctx context.Context, apiKey string) (TowerRoomState, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/tower/my-room", apiKey, nil)
	if err != nil {
		return TowerRoomState{}, err
	}
	if status < 200 || status >= 300 {
		return TowerRoomState{}, fmt.Errorf("tower get my room failed with status %d: %s", status, string(body))
	}
	var resp TowerRoomState
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return TowerRoomState{}, fmt.Errorf("parse tower my room response: %w", err)
	}
	return resp, nil
}

// TowerGetAllRooms returns all tower rooms with their current state
func (c *Client) TowerGetAllRooms(ctx context.Context) ([]TowerRoomState, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/tower", "", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("tower get all rooms failed with status %d: %s", status, string(body))
	}

	// Handle nested structure: { success: true, data: { rooms: [...] } }
	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		var wrapper struct {
			Rooms []TowerRoomState `json:"rooms"`
		}
		if err := json.Unmarshal(env.Data, &wrapper); err == nil {
			return wrapper.Rooms, nil
		}
	}

	// Fallback: try direct array parsing
	var resp []TowerRoomState
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse tower rooms response: %w", err)
	}
	return resp, nil
}

// TowerGetStatistics returns tower-wide statistics
func (c *Client) TowerGetStatistics(ctx context.Context) (TowerStatistics, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/tower/stats", "", nil)
	if err != nil {
		return TowerStatistics{}, err
	}
	if status < 200 || status >= 300 {
		return TowerStatistics{}, fmt.Errorf("tower get statistics failed with status %d: %s", status, string(body))
	}
	var resp TowerStatistics
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return TowerStatistics{}, fmt.Errorf("parse tower statistics response: %w", err)
	}
	return resp, nil
}

// TowerGetRoomDetail returns detailed information about a specific room
func (c *Client) TowerGetRoomDetail(ctx context.Context, roomCode string) (TowerRoomDetail, error) {
	path := "/api/v1/tower/room/" + strings.TrimSpace(roomCode)
	body, status, err := c.doRequestWithAPIKey(ctx, http.MethodGet, path, "", nil)
	if err != nil {
		return TowerRoomDetail{}, err
	}
	if status < 200 || status >= 300 {
		return TowerRoomDetail{}, fmt.Errorf("tower get room detail failed with status %d: %s", status, string(body))
	}
	var resp TowerRoomDetail
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return TowerRoomDetail{}, fmt.Errorf("parse tower room detail response: %w", err)
	}
	return resp, nil
}

// ──────────────────────────────────────────────────────────────
// Bot Messages API
// ──────────────────────────────────────────────────────────────

// BotMessage represents a single inbox message for a bot.
type BotMessage struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	SenderID   string  `json:"senderId"`
	SenderType int     `json:"senderType"`
	SenderName *string `json:"senderName"`
	SendTime   string  `json:"sendTime"`
	ReadTime   *string `json:"readTime"`
	Status     int     `json:"status"` // 0=deleted 1=unread 2=read
}

type BotMessageListResult struct {
	Items      []BotMessage
	Page       int
	PageSize   int
	TotalCount int
}

type BotMessageSendResult struct {
	ID          string `json:"id"`
	ToBotID     string `json:"toBotId"`
	ToBotName   string `json:"toBotName"`
	FromBotID   string `json:"fromBotId"`
	FromBotName string `json:"fromBotName"`
	SendTime    string `json:"sendTime"`
}

// SendMessageByBotName sends a bot-to-bot internal message; target is resolved only by bot_name.
func (c *Client) SendMessageByBotName(ctx context.Context, apiKey, toBotName, title, content string) (BotMessageSendResult, error) {
	toBotName = strings.TrimSpace(toBotName)
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if toBotName == "" {
		return BotMessageSendResult{}, fmt.Errorf("to bot name is required")
	}
	if title == "" {
		return BotMessageSendResult{}, fmt.Errorf("title is required")
	}
	if content == "" {
		return BotMessageSendResult{}, fmt.Errorf("content is required")
	}

	payload := map[string]string{
		"toBotName": toBotName,
		"title":     title,
		"content":   content,
	}
	body, httpStatus, err := c.doJSONWithAPIKey(ctx, http.MethodPost, "/api/v1/messages/send", apiKey, payload)
	if err != nil {
		return BotMessageSendResult{}, err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return BotMessageSendResult{}, fmt.Errorf("send message failed with status %d: %s", httpStatus, string(body))
	}

	var resp BotMessageSendResult
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return BotMessageSendResult{}, fmt.Errorf("parse send message response: %w", err)
	}
	return resp, nil
}

// ListMessages fetches the bot's messages. status: 0=deleted,1=unread,2=read; -1=all (no filter).
func (c *Client) ListMessages(ctx context.Context, apiKey string, status, page, pageSize int) (BotMessageListResult, error) {
	query := url.Values{}
	if status >= 0 {
		query.Set("status", fmt.Sprintf("%d", status))
	}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	path := "/api/v1/messages"
	if enc := query.Encode(); enc != "" {
		path += "?" + enc
	}
	body, httpStatus, err := c.doRequestWithAPIKey(ctx, http.MethodGet, path, apiKey, nil)
	if err != nil {
		return BotMessageListResult{}, err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return BotMessageListResult{}, fmt.Errorf("list messages failed with status %d: %s", httpStatus, string(body))
	}

	var raw struct {
		Success    bool         `json:"success"`
		Data       []BotMessage `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			TotalCount int `json:"totalCount"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return BotMessageListResult{}, fmt.Errorf("parse messages response: %w", err)
	}
	return BotMessageListResult{
		Items:      raw.Data,
		Page:       raw.Pagination.Page,
		PageSize:   raw.Pagination.PageSize,
		TotalCount: raw.Pagination.TotalCount,
	}, nil
}

// GetMessage fetches a single message and auto-marks it as read.
func (c *Client) GetMessage(ctx context.Context, apiKey, messageID string) (BotMessage, error) {
	id := strings.TrimSpace(messageID)
	if id == "" {
		return BotMessage{}, fmt.Errorf("message id is required")
	}
	body, httpStatus, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/messages/"+id, apiKey, nil)
	if err != nil {
		return BotMessage{}, err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return BotMessage{}, fmt.Errorf("get message failed with status %d: %s", httpStatus, string(body))
	}
	var msg BotMessage
	if err := decodeEnvelopeData(body, &msg); err != nil {
		return BotMessage{}, fmt.Errorf("parse message response: %w", err)
	}
	return msg, nil
}

// DeleteMessage soft-deletes a message (status → 0).
func (c *Client) DeleteMessage(ctx context.Context, apiKey, messageID string) error {
	id := strings.TrimSpace(messageID)
	if id == "" {
		return fmt.Errorf("message id is required")
	}
	body, httpStatus, err := c.doRequestWithAPIKey(ctx, http.MethodDelete, "/api/v1/messages/"+id, apiKey, nil)
	if err != nil {
		return err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return fmt.Errorf("delete message failed with status %d: %s", httpStatus, string(body))
	}
	return nil
}

// GetUnreadCount returns the number of unread messages for the bot.
func (c *Client) GetUnreadCount(ctx context.Context, apiKey string) (int, error) {
	body, httpStatus, err := c.doRequestWithAPIKey(ctx, http.MethodGet, "/api/v1/messages/unread-count", apiKey, nil)
	if err != nil {
		return 0, err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return 0, fmt.Errorf("unread count failed with status %d: %s", httpStatus, string(body))
	}
	var raw struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, fmt.Errorf("parse unread count response: %w", err)
	}
	return raw.Data.Count, nil
}

// ──────────────────────────────────────────────────────────────
// Pipeline API types
// ──────────────────────────────────────────────────────────────

// PipelineSessionInvitationResponse is returned when a bot creates a session invitation.
type PipelineSessionInvitationResponse struct {
	SessionToken               string `json:"sessionToken"`
	InitiatorBotId             string `json:"initiatorBotId"`
	ResponderBotId             string `json:"responderBotId"`
	InitiatorMonitoringEnabled bool   `json:"initiatorMonitoringEnabled"`
	ResponderMonitoringEnabled bool   `json:"responderMonitoringEnabled"`
	CreatedAt                  string `json:"createdAt"`
	ExpiresAt                  string `json:"expiresAt"`
}

// PipelineSessionResponse is returned for session lifecycle actions.
type PipelineSessionResponse struct {
	SessionId                  string  `json:"sessionId"`
	SessionToken               string  `json:"sessionToken"`
	InitiatorBotId             string  `json:"initiatorBotId"`
	InitiatorBotName           string  `json:"initiatorBotName"`
	ResponderBotId             string  `json:"responderBotId"`
	ResponderBotName           string  `json:"responderBotName"`
	Status                     string  `json:"status"`
	CreatedAt                  string  `json:"createdAt"`
	ActivatedAt                *string `json:"activatedAt,omitempty"`
	CompletedAt                *string `json:"completedAt,omitempty"`
	ExpiresAt                  string  `json:"expiresAt"`
	MessageCount               int     `json:"messageCount"`
	DurationSeconds            *int    `json:"durationSeconds,omitempty"`
	RejectionReason            *string `json:"rejectionReason,omitempty"`
	InitiatorMonitoringEnabled bool    `json:"initiatorMonitoringEnabled"`
	ResponderMonitoringEnabled bool    `json:"responderMonitoringEnabled"`
}

// PipelineMessageResponse is returned when a message is sent.
type PipelineMessageResponse struct {
	SessionToken   string                  `json:"sessionToken"`
	SenderBotId    string                  `json:"senderBotId"`
	RecipientBotId string                  `json:"recipientBotId"`
	Content        string                  `json:"content"`
	SentAt         string                  `json:"sentAt"`
	Encryption     *PipelineEncryptionMeta `json:"encryption,omitempty"`
	Delivered      bool                    `json:"delivered"`
	Queued         bool                    `json:"queued"`
}

// PipelineEncryptionMeta carries optional encryption metadata for a message.
type PipelineEncryptionMeta struct {
	Algorithm string `json:"algorithm"`
	KeyId     string `json:"keyId"`
}

// PipelineSessionMetadata is a lightweight session summary (used in history).
type PipelineSessionMetadata struct {
	SessionToken     string  `json:"sessionToken"`
	InitiatorBotId   string  `json:"initiatorBotId"`
	InitiatorBotName string  `json:"initiatorBotName"`
	ResponderBotId   string  `json:"responderBotId"`
	ResponderBotName string  `json:"responderBotName"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"createdAt"`
	ActivatedAt      *string `json:"activatedAt,omitempty"`
	CompletedAt      *string `json:"completedAt,omitempty"`
	MessageCount     int     `json:"messageCount"`
	DurationSeconds  *int    `json:"durationSeconds,omitempty"`
}

// PipelineSessionListResult holds a paginated list of session metadata.
type PipelineSessionListResult struct {
	Items      []PipelineSessionMetadata
	Page       int
	PageSize   int
	TotalCount int
}

// PipelineConnectionStatus describes a bot's current pipeline connection.
type PipelineConnectionStatus struct {
	BotId               string  `json:"botId"`
	IsOnline            bool    `json:"isOnline"`
	LastHeartbeat       *string `json:"lastHeartbeat,omitempty"`
	ActiveSessionsCount int     `json:"activeSessionsCount"`
	QueuedMessagesCount int     `json:"queuedMessagesCount"`
}

// ──────────────────────────────────────────────────────────────
// Pipeline REST API methods
// ──────────────────────────────────────────────────────────────

// BotTokenResponse is returned by POST /api/v1/pipeline/token.
type BotTokenResponse struct {
	Success   bool      `json:"success"`
	Token     string    `json:"token"`
	BotID     string    `json:"botId"`
	BotName   string    `json:"botName"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PipelineGetBotToken exchanges the plain API key for a signed bot JWT.
// The API key is sent as X-API-Key; no existing JWT is required.
func (c *Client) PipelineGetBotToken(ctx context.Context, apiKey string) (BotTokenResponse, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, "POST", "/api/v1/pipeline/token", apiKey, map[string]any{})
	if err != nil {
		return BotTokenResponse{}, err
	}
	if status < 200 || status >= 300 {
		return BotTokenResponse{}, fmt.Errorf("get bot token failed (%d): %s", status, string(body))
	}
	var resp BotTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return BotTokenResponse{}, fmt.Errorf("parse token response: %w", err)
	}
	return resp, nil
}

// PipelineGetSessionHistory returns the authenticated bot's session history.
func (c *Client) PipelineGetSessionHistory(ctx context.Context, apiKey string, page, pageSize int) (PipelineSessionListResult, error) {
	q := url.Values{}
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	path := "/api/v1/pipeline/sessions/history?" + q.Encode()

	body, status, err := c.doRequestWithAPIKey(ctx, "GET", path, apiKey, nil)
	if err != nil {
		return PipelineSessionListResult{}, err
	}
	if status < 200 || status >= 300 {
		return PipelineSessionListResult{}, fmt.Errorf("pipeline history failed (%d): %s", status, string(body))
	}

	var raw struct {
		Success    bool                      `json:"success"`
		Data       []PipelineSessionMetadata `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			TotalCount int `json:"totalCount"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return PipelineSessionListResult{}, fmt.Errorf("parse history response: %w", err)
	}
	return PipelineSessionListResult{
		Items:      raw.Data,
		Page:       raw.Pagination.Page,
		PageSize:   raw.Pagination.PageSize,
		TotalCount: raw.Pagination.TotalCount,
	}, nil
}

// PipelineGetSession returns a single session by its token.
func (c *Client) PipelineGetSession(ctx context.Context, apiKey string, sessionToken string) (*PipelineSessionResponse, error) {
	path := "/api/v1/pipeline/sessions/" + strings.TrimSpace(sessionToken)
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", path, apiKey, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, fmt.Errorf("session not found")
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get session failed (%d): %s", status, string(body))
	}
	var resp PipelineSessionResponse
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return nil, fmt.Errorf("parse session response: %w", err)
	}
	return &resp, nil
}

// PipelineGetConnectionStatus returns a bot's current pipeline connection status.
func (c *Client) PipelineGetConnectionStatus(ctx context.Context, apiKey string, botId string) (*PipelineConnectionStatus, error) {
	path := "/api/v1/pipeline/connections/" + strings.TrimSpace(botId) + "/status"
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", path, apiKey, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get connection status failed (%d): %s", status, string(body))
	}
	var resp PipelineConnectionStatus
	if err := decodeEnvelopeData(body, &resp); err != nil {
		return nil, fmt.Errorf("parse connection status response: %w", err)
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────
// Pipeline SignalR helper methods
// Each opens a one-shot connection, invokes the method, and closes.
// ──────────────────────────────────────────────────────────────

// PipelineSendInvitation sends a session invitation via TowerHub SignalR.
func (c *Client) PipelineSendInvitation(ctx context.Context, apiKey, responderBotId string) (*PipelineSessionInvitationResponse, error) {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		return nil, fmt.Errorf("join pipeline: %w", err)
	}

	raw, err := sc.Invoke(ctx, "SendInvitation", responderBotId)
	if err != nil {
		return nil, fmt.Errorf("send invitation: %w", err)
	}
	var resp PipelineSessionInvitationResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse invitation response: %w", err)
	}
	return &resp, nil
}

// PipelineAcceptSession accepts an incoming session invitation via TowerHub SignalR.
func (c *Client) PipelineAcceptSession(ctx context.Context, apiKey, sessionToken string) (*PipelineSessionResponse, error) {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		return nil, fmt.Errorf("join pipeline: %w", err)
	}

	raw, err := sc.Invoke(ctx, "AcceptSession", sessionToken)
	if err != nil {
		return nil, fmt.Errorf("accept session: %w", err)
	}
	var resp PipelineSessionResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse accept response: %w", err)
	}
	return &resp, nil
}

// PipelineRejectSession rejects a session invitation via TowerHub SignalR.
func (c *Client) PipelineRejectSession(ctx context.Context, apiKey, sessionToken, reason string) error {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return err
	}
	defer sc.Close()

	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		return fmt.Errorf("join pipeline: %w", err)
	}

	_, err = sc.Invoke(ctx, "RejectSession", sessionToken, reason)
	return err
}

// PipelineSendMessage sends a message in an active session via TowerHub SignalR.
func (c *Client) PipelineSendMessage(ctx context.Context, apiKey, sessionToken, content string, encryption *PipelineEncryptionMeta) (*PipelineMessageResponse, error) {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		return nil, fmt.Errorf("join pipeline: %w", err)
	}

	raw, err := sc.Invoke(ctx, "SendMessage", sessionToken, content, encryption)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	var resp PipelineMessageResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse message response: %w", err)
	}
	return &resp, nil
}

// PipelineEndSession ends an active session via TowerHub SignalR.
func (c *Client) PipelineEndSession(ctx context.Context, apiKey, sessionToken string) (*PipelineSessionMetadata, error) {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		return nil, fmt.Errorf("join pipeline: %w", err)
	}

	raw, err := sc.Invoke(ctx, "EndSession", sessionToken)
	if err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}
	var resp PipelineSessionMetadata
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse end session response: %w", err)
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────
// Room Mode API types
// ──────────────────────────────────────────────────────────────

// RoomCreatedResponse is returned when a bot creates a room.
type RoomCreatedResponse struct {
	RoomCode     string `json:"roomCode"`
	CreatorBotId string `json:"creatorBotId"`
	Capacity     int    `json:"capacity"`
	ExpiresAt    string `json:"expiresAt"`
	HasPassword  bool   `json:"hasPassword"`
}

// RoomParticipantDto describes a bot in a room.
type RoomParticipantDto struct {
	BotId     string `json:"botId"`
	BotName   string `json:"botName"`
	BotAvatar string `json:"botAvatar"`
	IsCreator bool   `json:"isCreator"`
	IsOnline  bool   `json:"isOnline"`
	JoinedAt  string `json:"joinedAt"`
}

// RoomJoinedResponse is returned when a bot joins a room.
type RoomJoinedResponse struct {
	RoomCode     string               `json:"roomCode"`
	Participants []RoomParticipantDto `json:"participants"`
	JoinedAt     string               `json:"joinedAt"`
	Rejoined     bool                 `json:"rejoined"`
}

// RoomInfoDto describes a room's current state.
type RoomInfoDto struct {
	RoomCode         string `json:"roomCode"`
	CreatorBotId     string `json:"creatorBotId"`
	CreatorBotName   string `json:"creatorBotName"`
	Capacity         int    `json:"capacity"`
	ParticipantCount int    `json:"participantCount"`
	HasPassword      bool   `json:"hasPassword"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
	ExpiresAt        string `json:"expiresAt"`
	MessageCount     int    `json:"messageCount"`
}

// RoomMessageDto describes one cached room message.
type RoomMessageDto struct {
	RoomCode       string `json:"roomCode"`
	SenderBotId    string `json:"senderBotId"`
	SenderBotName  string `json:"senderBotName"`
	Content        string `json:"content"`
	EncryptionMeta any    `json:"encryptionMetadata"`
	SentAt         string `json:"sentAt"`
}

// RoomStatsDto holds platform-wide room statistics.
type RoomStatsDto struct {
	ActiveRoomCount    int `json:"activeRoomCount"`
	TotalRoomsToday    int `json:"totalRoomsToday"`
	TotalRoomsThisWeek int `json:"totalRoomsThisWeek"`
}

// ──────────────────────────────────────────────────────────────
// Room Mode SignalR helper methods (one-shot connection each)
// ──────────────────────────────────────────────────────────────

func (c *Client) roomConnect(ctx context.Context, apiKey string) (*SignalRConn, error) {
	sc, err := c.ConnectToHub(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if err := sc.InvokeVoid(ctx, "JoinPipeline"); err != nil {
		sc.Close()
		return nil, fmt.Errorf("join pipeline: %w", err)
	}
	return sc, nil
}

// RoomCreate creates a new room and returns its code.
func (c *Client) RoomCreate(ctx context.Context, apiKey string, capacity int, password string, ttlMinutes int) (*RoomCreatedResponse, error) {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	// Build args map — server accepts nullable fields
	args := map[string]interface{}{}
	if capacity > 0 {
		args["capacity"] = capacity
	}
	if password != "" {
		args["password"] = password
	}
	if ttlMinutes > 0 {
		args["ttlMinutes"] = ttlMinutes
	}

	raw, err := sc.Invoke(ctx, "CreateRoom", capacity, password, ttlMinutes)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	// Server returns { success, roomCode, expiresAt } — parse into RoomCreatedResponse
	var wrapper struct {
		RoomCode  string `json:"roomCode"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("parse create room response: %w", err)
	}
	return &RoomCreatedResponse{
		RoomCode:  wrapper.RoomCode,
		ExpiresAt: wrapper.ExpiresAt,
		Capacity:  capacity,
	}, nil
}

// RoomJoin joins a room by code and returns the participant list.
func (c *Client) RoomJoin(ctx context.Context, apiKey, roomCode, password string) (*RoomJoinedResponse, error) {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	raw, err := sc.Invoke(ctx, "JoinRoom", roomCode, password)
	if err != nil {
		return nil, fmt.Errorf("join room: %w", err)
	}
	var wrapper struct {
		RoomCode     string               `json:"roomCode"`
		Participants []RoomParticipantDto `json:"participants"`
		JoinedAt     string               `json:"joinedAt"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("parse join room response: %w", err)
	}
	return &RoomJoinedResponse{
		RoomCode:     wrapper.RoomCode,
		Participants: wrapper.Participants,
		JoinedAt:     wrapper.JoinedAt,
	}, nil
}

// RoomLeave leaves a room.
func (c *Client) RoomLeave(ctx context.Context, apiKey, roomCode string) error {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return err
	}
	defer sc.Close()

	if _, err := sc.Invoke(ctx, "LeaveRoom", roomCode); err != nil {
		return fmt.Errorf("leave room: %w", err)
	}
	return nil
}

// RoomClose closes a room (creator only).
func (c *Client) RoomClose(ctx context.Context, apiKey, roomCode, reason string) error {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return err
	}
	defer sc.Close()

	if _, err := sc.Invoke(ctx, "CloseRoom", roomCode, reason); err != nil {
		return fmt.Errorf("close room: %w", err)
	}
	return nil
}

// RoomSendMessage sends a message to all room participants.
func (c *Client) RoomSendMessage(ctx context.Context, apiKey, roomCode, content string) error {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return err
	}
	defer sc.Close()

	if _, err := sc.Invoke(ctx, "SendRoomMessage", roomCode, content, nil); err != nil {
		return fmt.Errorf("send room message: %w", err)
	}
	return nil
}

// RoomGetInfo returns current info about a room via REST API.
func (c *Client) RoomGetInfo(ctx context.Context, apiKey, roomCode string) (*RoomInfoDto, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", "/api/v1/rooms/"+roomCode, apiKey, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, fmt.Errorf("room %s not found", roomCode)
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get room info failed (%d): %s", status, string(body))
	}
	var info RoomInfoDto
	if err := decodeEnvelopeData(body, &info); err != nil {
		return nil, fmt.Errorf("parse room info: %w", err)
	}
	return &info, nil
}

// RoomGetParticipants returns the participant list for a room via REST API.
func (c *Client) RoomGetParticipants(ctx context.Context, apiKey, roomCode string) ([]RoomParticipantDto, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", "/api/v1/rooms/"+roomCode+"/participants", apiKey, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get participants failed (%d): %s", status, string(body))
	}
	var participants []RoomParticipantDto
	if err := decodeEnvelopeData(body, &participants); err != nil {
		return nil, fmt.Errorf("parse participants: %w", err)
	}
	return participants, nil
}

// RoomGetMessages returns recent cached room messages via REST API.
func (c *Client) RoomGetMessages(ctx context.Context, apiKey, roomCode string, limit int) ([]RoomMessageDto, error) {
	path := fmt.Sprintf("/api/v1/rooms/%s/messages?limit=%d", roomCode, limit)
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", path, apiKey, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get messages failed (%d): %s", status, string(body))
	}
	var messages []RoomMessageDto
	if err := decodeEnvelopeData(body, &messages); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}
	return messages, nil
}

// RoomExtendTtl extends a room's TTL (creator only).
func (c *Client) RoomExtendTtl(ctx context.Context, apiKey, roomCode string, additionalMinutes int) error {
	sc, err := c.roomConnect(ctx, apiKey)
	if err != nil {
		return err
	}
	defer sc.Close()

	if _, err := sc.Invoke(ctx, "ExtendRoomTtl", roomCode, additionalMinutes); err != nil {
		return fmt.Errorf("extend room TTL: %w", err)
	}
	return nil
}

// RoomGetPublicStats returns platform-wide room statistics (no auth required).
func (c *Client) RoomGetPublicStats(ctx context.Context) (*RoomStatsDto, error) {
	body, status, err := c.doRequestWithAPIKey(ctx, "GET", "/api/v1/rooms/public/stats", "", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get public stats failed (%d): %s", status, string(body))
	}
	var stats RoomStatsDto
	if err := decodeEnvelopeData(body, &stats); err != nil {
		return nil, fmt.Errorf("parse public stats: %w", err)
	}
	return &stats, nil
}

// saveToLocalDB saves diary to local SQLite database
func saveToLocalDB(date, summary string) error {
	// Build local db path
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".moltbb", "local-web", "local.db")

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// Check if entry exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM diary_entries WHERE date = ?)", date).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check exists: %w", err)
	}

	// Extract title from summary (first line)
	title := summary
	if lines := strings.Split(summary, "\n"); len(lines) > 0 {
		title = strings.TrimSpace(lines[0])
		if len(title) > 100 {
			title = title[:100]
		}
	}

	if exists {
		_, err = db.Exec(`
			UPDATE diary_entries 
			SET title = ?, preview = ?, content_text = ?, modified_at = datetime('now')
			WHERE date = ?`,
			title, summary, summary, date)
	} else {
		// Generate unique id
		uniqueID := date + "-" + fmt.Sprintf("%d", time.Now().Unix())
		relPath := date + ".md"
		_, err = db.Exec(`
			INSERT INTO diary_entries (id, rel_path, filename, date, title, preview, content_text, size, modified_at, indexed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
			uniqueID, relPath, relPath, date, title, summary, summary, len(summary))
	}

	if err != nil {
		return fmt.Errorf("save diary: %w", err)
	}

	return nil
}

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"moltbb-cli/internal/config"
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

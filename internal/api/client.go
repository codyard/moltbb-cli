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
	Valid   bool   `json:"valid"`
	Token   string `json:"token,omitempty"`
	OwnerID string `json:"owner_id,omitempty"`
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

func (c *Client) doJSONWithAPIKey(ctx context.Context, method, path, apiKey string, payload any) ([]byte, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	maxAttempts := c.retryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(data))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
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

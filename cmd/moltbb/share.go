package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
)

func newShareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share <file>",
		Short: "Upload a temporary shared file (max 50 MB, expires in 1 day)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			filePath := strings.TrimSpace(args[0])
			info, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("file not found: %w", err)
			}
			if info.IsDir() {
				return fmt.Errorf("path is a directory, not a file")
			}
			if info.Size() > 50*1024*1024 {
				return fmt.Errorf("file size %.1f MB exceeds 50 MB limit", float64(info.Size())/1024/1024)
			}

			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve api key: %w", err)
			}

			result, err := uploadSharedFile(cfg, apiKey, filePath)
			if err != nil {
				return err
			}

			fmt.Println("File shared successfully")
			fmt.Println("URL:     ", result.URL)
			fmt.Println("Code:    ", result.FileCode)
			fmt.Println("Expires: ", result.ExpiresAt.Format("2006-01-02 15:04 UTC"))
			fmt.Printf("Size:     %.1f KB\n", float64(result.FileSize)/1024)
			return nil
		},
	}
	return cmd
}

type shareUploadResult struct {
	FileCode     string    `json:"fileCode"`
	URL          string    `json:"url"`
	OriginalName string    `json:"originalFileName"`
	FileSize     int64     `json:"fileSize"`
	ContentType  string    `json:"contentType"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type shareEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func uploadSharedFile(cfg config.Config, apiKey, filePath string) (shareUploadResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return shareUploadResult{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	fileName := filepath.Base(filePath)
	contentType := detectContentType(filePath)

	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return shareUploadResult{}, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return shareUploadResult{}, fmt.Errorf("copy file: %w", err)
	}
	w.Close()

	url := strings.TrimRight(cfg.APIBaseURL, "/") + "/api/v1/files"
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		&body,
	)
	if err != nil {
		return shareUploadResult{}, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)
	_ = contentType // server detects from file content

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return shareUploadResult{}, fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return shareUploadResult{}, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var env shareEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return shareUploadResult{}, fmt.Errorf("parse response: %w", err)
	}
	if !env.Success {
		return shareUploadResult{}, fmt.Errorf("upload failed: %s", env.Message)
	}

	var result shareUploadResult
	if err := json.Unmarshal(env.Data, &result); err != nil {
		return shareUploadResult{}, fmt.Errorf("parse result: %w", err)
	}
	return result, nil
}

func detectContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/localweb"
	"moltbb-cli/internal/utils"
)

type statusCard struct {
	Version         string
	APIBaseURL      string
	APIKeyMasked    string
	APIKeyOK        bool
	BotID           string
	Activation      string
	LocalWebURL     string
	LocalDBPath     string
	LocalDiaryCount int
	LocalLastUpdate string
	CloudDiaryCount *int
	CloudInsightCnt *int
}

func buildStatusCard() statusCard {
	card := statusCard{
		Version:     version,
		LocalWebURL: "http://127.0.0.1:3789",
	}

	cfg, cfgErr := config.Load()
	if cfgErr == nil {
		card.APIBaseURL = cfg.APIBaseURL
	}

	if key, err := auth.ResolveAPIKey(); err == nil {
		card.APIKeyOK = true
		card.APIKeyMasked = maskAPIKey(key)
	}

	if state, err := binding.Load(); err == nil && state.Bound {
		card.BotID = state.BotID
		card.Activation = state.ActivationStatus
	}

	if dbPath, err := resolveLocalDBPath(); err == nil {
		card.LocalDBPath = dbPath
		populateLocalDiaryStats(&card, dbPath)
	}

	if cfgErr == nil && card.APIKeyOK {
		populateCloudStats(&card, cfg)
	}

	return card
}

func resolveLocalDBPath() (string, error) {
	home, err := utils.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".moltbb", "local-web", "local.db"), nil
}

func populateLocalDiaryStats(card *statusCard, dbPath string) {
	if strings.TrimSpace(dbPath) == "" {
		return
	}
	if _, err := os.Stat(dbPath); err != nil {
		return
	}
	adb, err := localweb.OpenDB(dbPath)
	if err != nil {
		return
	}
	defer adb.Close()

	var count int
	if err := adb.QueryRow(`SELECT COUNT(1) FROM diary_entries`).Scan(&count); err == nil {
		card.LocalDiaryCount = count
	}

	var last sql.NullString
	if err := adb.QueryRow(`SELECT MAX(modified_at) FROM diary_entries`).Scan(&last); err == nil {
		if last.Valid {
			card.LocalLastUpdate = last.String
		}
	}
}

func populateCloudStats(card *statusCard, cfg config.Config) {
	client, err := api.NewClient(cfg)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()

	if card.APIKeyMasked == "" {
		return
	}
	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		return
	}

	if diaries, err := client.ListRuntimeDiaries(ctx, apiKey, "", "", 1, 1); err == nil {
		cnt := diaries.TotalCount
		card.CloudDiaryCount = &cnt
	}

	if insights, err := client.ListRuntimeInsights(ctx, apiKey, "", 1, 1); err == nil {
		cnt := insights.TotalCount
		card.CloudInsightCnt = &cnt
	}
}

func (c statusCard) render() string {
	apiLine := "API: not configured"
	if strings.TrimSpace(c.APIBaseURL) != "" {
		apiLine = fmt.Sprintf("API: %s", c.APIBaseURL)
	}
	keyLine := "🔑 key: missing"
	if c.APIKeyOK {
		keyLine = fmt.Sprintf("🔑 key: %s", c.APIKeyMasked)
	}

	botLine := "🤖 Bot: not bound"
	if strings.TrimSpace(c.BotID) != "" {
		activation := strings.TrimSpace(c.Activation)
		if activation == "" {
			activation = "unknown"
		}
		botLine = fmt.Sprintf("🤖 Bot: %s · %s", c.BotID, activation)
	}

	localLine := "🏠 Local: not initialized"
	if strings.TrimSpace(c.LocalDBPath) != "" {
		localLine = fmt.Sprintf("🏠 Local: %s · DB: %s", c.LocalWebURL, c.LocalDBPath)
	}

	diaryLine := fmt.Sprintf("📚 Diaries: local %d", c.LocalDiaryCount)
	if c.CloudDiaryCount != nil {
		diaryLine = fmt.Sprintf("📚 Diaries: local %d / cloud %d", c.LocalDiaryCount, *c.CloudDiaryCount)
	}

	insightLine := "💡 Insights: cloud n/a"
	if c.CloudInsightCnt != nil {
		insightLine = fmt.Sprintf("💡 Insights: cloud %d", *c.CloudInsightCnt)
	}

	lastLine := "🕒 Last local update: n/a"
	if strings.TrimSpace(c.LocalLastUpdate) != "" {
		lastLine = fmt.Sprintf("🕒 Last local update: %s", c.LocalLastUpdate)
	}

	return strings.Join([]string{
		fmt.Sprintf("🦞 MoltBB %s", c.Version),
		fmt.Sprintf("%s · %s", apiLine, keyLine),
		botLine,
		localLine,
		diaryLine + " · " + insightLine,
		lastLine,
	}, "\n")
}

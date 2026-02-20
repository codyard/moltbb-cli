package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/utils"
)

func newInsightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insight",
		Short: "Manage runtime insights",
	}
	cmd.AddCommand(newInsightUploadCmd())
	cmd.AddCommand(newInsightListCmd())
	cmd.AddCommand(newInsightUpdateCmd())
	cmd.AddCommand(newInsightDeleteCmd())
	return cmd
}

func newInsightUploadCmd() *cobra.Command {
	var title string
	var diaryID string
	var tags []string
	var catalogs []string
	var visibilityLevel int

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload one insight from local markdown file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if visibilityLevel < 0 || visibilityLevel > 1 {
				return errors.New("visibility-level must be 0 or 1")
			}
			resolvedFile, content, err := readInsightFile(args[0])
			if err != nil {
				return err
			}

			resolvedTitle := strings.TrimSpace(title)
			if resolvedTitle == "" {
				resolvedTitle = inferInsightTitle(content, resolvedFile)
			}

			payload := api.RuntimeInsightCreatePayload{
				Title:           resolvedTitle,
				DiaryID:         strings.TrimSpace(diaryID),
				Catalogs:        normalizeStringList(catalogs),
				Content:         strings.TrimSpace(content),
				Tags:            normalizeStringList(tags),
				VisibilityLevel: visibilityLevel,
			}

			resp, err := createRuntimeInsight(cfg, payload)
			if err != nil {
				return err
			}

			fmt.Println("Insight upload success")
			fmt.Println("File:", resolvedFile)
			fmt.Println("Insight ID:", resp.ID)
			fmt.Println("Title:", resp.Title)
			fmt.Println("Visibility level:", resp.VisibilityLevel)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Insight title (default: first # heading or filename)")
	cmd.Flags().StringVar(&diaryID, "diary-id", "", "Related diary ID (optional)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Insight tags, repeat or use comma-separated values")
	cmd.Flags().StringSliceVar(&catalogs, "catalogs", nil, "Insight catalogs, repeat or use comma-separated values")
	cmd.Flags().IntVar(&visibilityLevel, "visibility-level", 0, "Visibility level: 0=public, 1=private")
	return cmd
}

func newInsightListCmd() *cobra.Command {
	var page int
	var pageSize int
	var tags []string
	var diaryID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List insights for current bound bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve api key: %w", err)
			}
			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			result, err := client.ListRuntimeInsights(ctx, apiKey, page, pageSize, normalizeStringList(tags), strings.TrimSpace(diaryID))
			if err != nil {
				return err
			}

			fmt.Printf("Insights page %d/%d, total %d\n", result.Page, result.TotalPages, result.TotalCount)
			if len(result.Items) == 0 {
				fmt.Println("No insights found")
				return nil
			}
			for _, item := range result.Items {
				fmt.Printf("- %s | %s | visibility=%d | likes=%d\n", item.ID, item.Title, item.VisibilityLevel, item.Likes)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Filter tags, repeat or use comma-separated values")
	cmd.Flags().StringVar(&diaryID, "diary-id", "", "Filter by related diary ID")
	return cmd
}

func newInsightUpdateCmd() *cobra.Command {
	var title string
	var tags []string
	var catalogs []string
	var visibilityLevel int
	var setVisibility bool

	cmd := &cobra.Command{
		Use:   "update <insight-id> <file>",
		Short: "Update one insight from local markdown file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			insightID := strings.TrimSpace(args[0])
			if insightID == "" {
				return errors.New("insight-id is required")
			}

			_, content, err := readInsightFile(args[1])
			if err != nil {
				return err
			}
			content = strings.TrimSpace(content)
			if content == "" {
				return errors.New("insight file is empty")
			}

			var titlePtr *string
			if trimmed := strings.TrimSpace(title); trimmed != "" {
				titlePtr = &trimmed
			}
			var contentPtr *string
			contentPtr = &content

			var visibilityPtr *int
			if setVisibility {
				if visibilityLevel < 0 || visibilityLevel > 1 {
					return errors.New("visibility-level must be 0 or 1")
				}
				visibilityPtr = &visibilityLevel
			}

			payload := api.RuntimeInsightUpdatePayload{
				Title:           titlePtr,
				Catalogs:        normalizeStringList(catalogs),
				Content:         contentPtr,
				Tags:            normalizeStringList(tags),
				VisibilityLevel: visibilityPtr,
			}

			resp, err := updateRuntimeInsight(cfg, insightID, payload)
			if err != nil {
				return err
			}

			fmt.Println("Insight update success")
			fmt.Println("Insight ID:", resp.ID)
			fmt.Println("Title:", resp.Title)
			fmt.Println("Visibility level:", resp.VisibilityLevel)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Insight title (optional)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Insight tags, repeat or use comma-separated values")
	cmd.Flags().StringSliceVar(&catalogs, "catalogs", nil, "Insight catalogs, repeat or use comma-separated values")
	cmd.Flags().IntVar(&visibilityLevel, "visibility-level", 0, "Visibility level: 0=public, 1=private")
	cmd.Flags().BoolVar(&setVisibility, "set-visibility", false, "Whether to apply visibility-level on update")
	return cmd
}

func newInsightDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <insight-id>",
		Short: "Delete one insight",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			insightID := strings.TrimSpace(args[0])
			if insightID == "" {
				return errors.New("insight-id is required")
			}

			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve api key: %w", err)
			}
			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()
			if err := client.DeleteRuntimeInsight(ctx, apiKey, insightID); err != nil {
				return err
			}

			fmt.Println("Insight delete success")
			fmt.Println("Insight ID:", insightID)
			return nil
		},
	}
	return cmd
}

func createRuntimeInsight(cfg config.Config, payload api.RuntimeInsightCreatePayload) (api.RuntimeInsight, error) {
	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		return api.RuntimeInsight{}, fmt.Errorf("resolve api key: %w", err)
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return api.RuntimeInsight{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	return client.CreateRuntimeInsight(ctx, apiKey, payload)
}

func updateRuntimeInsight(cfg config.Config, insightID string, payload api.RuntimeInsightUpdatePayload) (api.RuntimeInsight, error) {
	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		return api.RuntimeInsight{}, fmt.Errorf("resolve api key: %w", err)
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return api.RuntimeInsight{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	return client.UpdateRuntimeInsight(ctx, apiKey, insightID, payload)
}

func readInsightFile(inputPath string) (string, string, error) {
	expanded, err := utils.ExpandPath(strings.TrimSpace(inputPath))
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(expanded)
	if err != nil {
		return "", "", fmt.Errorf("stat insight file: %w", err)
	}
	if info.IsDir() {
		return "", "", errors.New("insight file path cannot be a directory")
	}
	raw, err := os.ReadFile(expanded)
	if err != nil {
		return "", "", fmt.Errorf("read insight file: %w", err)
	}
	content := strings.TrimSpace(string(raw))
	if content == "" {
		return "", "", errors.New("insight file is empty")
	}
	return expanded, content, nil
}

func inferInsightTitle(content, filePath string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			if title != "" {
				return title
			}
		}
	}
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	if strings.TrimSpace(base) == "" {
		return "Untitled Insight"
	}
	return strings.TrimSpace(base)
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

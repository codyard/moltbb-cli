package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/diary"
	"moltbb-cli/internal/utils"
)

func newDiaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diary",
		Short: "Manage runtime diary upload workflow",
	}
	cmd.AddCommand(newDiaryUploadCmd())
	cmd.AddCommand(newDiaryPatchCmd())
	return cmd
}

func newDiaryUploadCmd() *cobra.Command {
	var diaryDate string
	var executionLevel int

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload or update a runtime diary from local markdown file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			result, resolvedFile, payload, err := upsertDiaryFromFile(cfg, args[0], diaryDate, executionLevel)
			if err != nil {
				return err
			}

			fmt.Println("Diary sync success")
			fmt.Println("File:", resolvedFile)
			fmt.Println("Diary date (UTC):", payload.DiaryDate)
			fmt.Println("Execution level:", payload.ExecutionLevel)
			fmt.Println("Action:", result.Action)
			if result.DiaryID != "" {
				fmt.Println("Diary ID:", result.DiaryID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&diaryDate, "date", "", "Diary date (YYYY-MM-DD), defaults to date parsed from filename or UTC today")
	cmd.Flags().IntVar(&executionLevel, "execution-level", 0, "Execution level to upload (0-4)")
	return cmd
}

func newDiaryPatchCmd() *cobra.Command {
	var summary string
	var content string

	cmd := &cobra.Command{
		Use:   "patch <diary-id>",
		Short: "Patch runtime diary summary/content by diary ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			diaryID := strings.TrimSpace(args[0])
			if diaryID == "" {
				return errors.New("diary-id is required")
			}

			var summaryPtr *string
			if cmd.Flags().Changed("summary") {
				v := summary
				summaryPtr = &v
			}

			var contentPtr *string
			if cmd.Flags().Changed("content") {
				v := content
				contentPtr = &v
			}

			if summaryPtr == nil && contentPtr == nil {
				return errors.New("at least one of --summary or --content is required")
			}

			if err := patchRuntimeDiary(cfg, diaryID, api.RuntimeDiaryPatchPayload{
				Summary: summaryPtr,
				Content: contentPtr,
			}); err != nil {
				return err
			}

			fmt.Println("Diary patch success")
			fmt.Println("Diary ID:", diaryID)
			if summaryPtr != nil {
				fmt.Println("Summary: updated")
			}
			if contentPtr != nil {
				fmt.Println("Content: updated")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "Patch diary summary")
	cmd.Flags().StringVar(&content, "content", "", "Patch diary content")
	return cmd
}

func patchRuntimeDiary(cfg config.Config, diaryID string, payload api.RuntimeDiaryPatchPayload) error {
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

	return client.PatchRuntimeDiary(ctx, apiKey, diaryID, payload)
}

func upsertDiaryFromFile(cfg config.Config, filePath, diaryDate string, executionLevel int) (api.RuntimeDiaryUpsertResult, string, diary.RuntimeUpsertPayload, error) {
	expandedFile, err := utils.ExpandPath(strings.TrimSpace(filePath))
	if err != nil {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, err
	}
	info, err := os.Stat(expandedFile)
	if err != nil {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, fmt.Errorf("stat diary file: %w", err)
	}
	if info.IsDir() {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, errors.New("diary file path cannot be a directory")
	}

	payload, err := diary.BuildRuntimeUpsertPayload(expandedFile, diaryDate, executionLevel, time.Now().UTC())
	if err != nil {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, err
	}

	apiKey, err := auth.ResolveAPIKey()
	if err != nil {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, fmt.Errorf("resolve api key: %w", err)
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, err
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
		return api.RuntimeDiaryUpsertResult{}, "", diary.RuntimeUpsertPayload{}, err
	}
	return result, expandedFile, payload, nil
}

func resolveMemoryDiaryFile(memoryDir, explicitFile, date string) (string, bool, error) {
	if trimmed := strings.TrimSpace(explicitFile); trimmed != "" {
		expanded, err := utils.ExpandPath(trimmed)
		if err != nil {
			return "", false, err
		}
		info, err := os.Stat(expanded)
		if err != nil {
			return "", false, err
		}
		if info.IsDir() {
			return "", false, errors.New("memory diary file points to a directory")
		}
		return expanded, true, nil
	}

	dir := strings.TrimSpace(memoryDir)
	if dir == "" {
		dir = "memory/daily"
	}
	expandedDir, err := utils.ExpandPath(dir)
	if err != nil {
		return "", false, err
	}
	info, err := os.Stat(expandedDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if !info.IsDir() {
		return "", false, errors.New("memory diary directory is not a directory")
	}

	exact := filepath.Join(expandedDir, date+".md")
	if _, err := os.Stat(exact); err == nil {
		return exact, true, nil
	}

	pattern := filepath.Join(expandedDir, date+"*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", false, err
	}
	if len(matches) == 0 {
		return "", false, nil
	}
	sort.Strings(matches)
	return matches[0], true, nil
}

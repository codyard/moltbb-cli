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
	cmd.AddCommand(newDiaryPublishCmd())
	cmd.AddCommand(newDiaryPullCmd())
	cmd.AddCommand(newDiaryPatchCmd())
	cmd.AddCommand(newDiaryDeleteCmd())
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

func newDiaryPublishCmd() *cobra.Command {
	var diaryDate string
	var executionLevel int
	var doLocalSync bool
	var forceSync bool

	cmd := &cobra.Command{
		Use:   "publish <file>",
		Short: "Sync local DB then upload runtime diary from local markdown file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			filePath := strings.TrimSpace(args[0])
			if filePath == "" {
				return errors.New("file path is required")
			}
			expandedFile, err := utils.ExpandPath(filePath)
			if err != nil {
				return err
			}
			if _, err := os.Stat(expandedFile); err != nil {
				return err
			}

			if doLocalSync {
				diaryDir := filepath.Dir(expandedFile)
				_, _ = syncDiaryFiles(diaryDir, forceSync)
			}

			result, resolvedFile, payload, err := upsertDiaryFromFile(cfg, expandedFile, diaryDate, executionLevel)
			if err != nil {
				return err
			}

			fmt.Println("Diary publish success")
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
	cmd.Flags().BoolVar(&doLocalSync, "local-sync", true, "Sync local database before upload")
	cmd.Flags().BoolVar(&forceSync, "force-sync", false, "Force overwrite existing local entries")
	return cmd
}

func newDiaryPullCmd() *cobra.Command {
	var startDate string
	var endDate string
	var outputDir string
	var overwrite bool
	var doLocalSync bool
	var forceSync bool
	var pageSize int

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Download runtime diaries from cloud and backfill local files/DB",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			startDate = strings.TrimSpace(startDate)
			endDate = strings.TrimSpace(endDate)
			if startDate == "" || endDate == "" {
				return errors.New("--start and --end are required (YYYY-MM-DD)")
			}
			if _, err := time.Parse("2006-01-02", startDate); err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			if _, err := time.Parse("2006-01-02", endDate); err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}

			if strings.TrimSpace(outputDir) == "" {
				outputDir = cfg.OutputDir
			}
			expandedDir, err := utils.ExpandPath(outputDir)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(expandedDir, 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
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

			page := 1
			written := 0
			for {
				result, err := client.ListRuntimeDiaries(ctx, apiKey, startDate, endDate, page, pageSize)
				if err != nil {
					return err
				}
				if len(result.Items) == 0 {
					break
				}

				for _, item := range result.Items {
					date := strings.TrimSpace(item.DiaryDate)
					if date == "" {
						date = strings.TrimSpace(item.Date)
					}
					if date == "" {
						continue
					}
					filePath := filepath.Join(expandedDir, date+".md")
					if !overwrite {
						if _, err := os.Stat(filePath); err == nil {
							continue
						}
					}

					parts := []string{}
					if strings.TrimSpace(item.Summary) != "" {
						parts = append(parts, strings.TrimSpace(item.Summary))
					}
					if strings.TrimSpace(item.PersonaText) != "" {
						parts = append(parts, strings.TrimSpace(item.PersonaText))
					}
					content := strings.Join(parts, "\n\n")
					if strings.TrimSpace(content) == "" {
						content = "# " + date
					}
					if err := os.WriteFile(filePath, []byte(content+"\n"), 0o644); err != nil {
						return fmt.Errorf("write %s: %w", filePath, err)
					}
					written++
				}

				page++
				if result.TotalPages > 0 && page > result.TotalPages {
					break
				}
			}

			if doLocalSync {
				_, _ = syncDiaryFiles(expandedDir, forceSync)
			}

			fmt.Printf("Pulled %d diaries into %s\n", written, expandedDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to write diary files (default: config output_dir)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing diary files")
	cmd.Flags().BoolVar(&doLocalSync, "local-sync", true, "Sync local database after download")
	cmd.Flags().BoolVar(&forceSync, "force-sync", false, "Force overwrite existing local entries")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Page size for API list")
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

func newDiaryDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <diary-id>",
		Short: "Delete a runtime diary by diary ID (only deletes your own diaries)",
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

			if err := deleteRuntimeDiary(cfg, diaryID); err != nil {
				return err
			}

			fmt.Println("Diary deleted successfully")
			fmt.Println("Diary ID:", diaryID)
			return nil
		},
	}

	return cmd
}

func deleteRuntimeDiary(cfg config.Config, diaryID string) error {
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

	return client.DeleteRuntimeDiary(ctx, apiKey, diaryID)
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

	payload, err := diary.BuildRuntimeUpsertPayload(expandedFile, diaryDate, executionLevel, time.Now())
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

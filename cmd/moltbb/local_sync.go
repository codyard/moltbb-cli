package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
	_ "modernc.org/sqlite"
)

func newLocalSyncCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "local-sync",
		Short: "Sync local diary files to local database",
		Long:  `Read local diary files and save to local SQLite database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Find diary files - use diaries_dir or output_dir
			diaryPaths := []string{}
			diaryDir := cfg.DiariesDir
			if diaryDir == "" {
				diaryDir = cfg.OutputDir
			}
			// Resolve relative path
			if diaryDir != "" && !filepath.IsAbs(diaryDir) {
				homeDir, _ := os.UserHomeDir()
				diaryDir = filepath.Join(homeDir, "moltbb", diaryDir)
			}
			if diaryDir != "" {
				diaryPaths = []string{diaryDir}
			} else {
				homeDir, _ := os.UserHomeDir()
				diaryPaths = []string{filepath.Join(homeDir, "moltbb", "diaries")}
			}

			output.PrintInfo(fmt.Sprintf("Scanning: %v", diaryPaths))

			var synced int
			for _, diaryPath := range diaryPaths {
				count, err := syncDiaryFiles(diaryPath, force)
				if err != nil {
					output.PrintError(fmt.Sprintf("Error scanning %s: %v", diaryPath, err))
					continue
				}
				synced += count
			}

			output.Success(fmt.Sprintf("Synced %d diary entries to local database", synced))
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing entries")

	return cmd
}

func syncDiaryFiles(diaryPath string, force bool) (int, error) {
	// Open local database
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".moltbb", "local-web", "local.db")
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	var count int

	err = filepath.Walk(diaryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		// Extract date from filename (e.g., 2026-03-03.md)
		filename := filepath.Base(path)
		date := strings.TrimSuffix(filename, filepath.Ext(filename))

		// Validate date format
		if len(date) != 10 || date[4] != '-' || date[7] != '-' {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		text := string(content)
		
		// Extract title (first line)
		title := text
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			title = strings.TrimSpace(lines[0])
			title = strings.TrimPrefix(title, "# ")
			if len(title) > 100 {
				title = title[:100]
			}
		}

		// Preview (first 200 chars)
		preview := text
		if len(preview) > 200 {
			preview = preview[:200]
		}

		// Check if entry exists
		if !force {
			var exists bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM diary_entries WHERE date = ?)", date).Scan(&exists)
			if err == nil && exists {
				return nil // Skip existing
			}
		}

		// Insert or update
		if force {
			_, err = db.Exec(`
				UPDATE diary_entries 
				SET title = ?, preview = ?, content_text = ?, modified_at = datetime('now')
				WHERE date = ?`,
				title, preview, text, date)
		}
		
		uniqueID := date + "-" + fmt.Sprintf("%d", time.Now().Unix())
		relPath := date + ".md"
		_, err = db.Exec(`
			INSERT OR REPLACE INTO diary_entries (id, rel_path, filename, date, title, preview, content_text, size, modified_at, indexed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
			uniqueID, relPath, relPath, date, title, preview, text, len(text))

		if err != nil {
			// Ignore duplicate rel_path errors to avoid crashing local service.
			if strings.Contains(err.Error(), "UNIQUE constraint failed: diary_entries.rel_path") {
				return nil
			}
			return nil
		}

		count++
		fmt.Printf("Synced: %s\n", date)

		return nil
	})

	return count, err
}

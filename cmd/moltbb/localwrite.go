package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/config"
	"moltbb-cli/internal/utils"
)

func newLocalWriteCmd() *cobra.Command {
	var title string
	var diaryDir string

	cmd := &cobra.Command{
		Use:   "local-write [title]",
		Short: "Create a local diary entry (offline, no login required)",
		Long: `Create a new diary entry locally without requiring login.
The diary will be saved to the configured output directory.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine title
			if len(args) > 0 {
				title = args[0]
			}
			if title == "" {
				title = time.Now().Format("2006-01-02")
			}

			// Load config to get output directory
			cfg, err := config.Load()
			if err != nil {
				// Use default directory if no config
				moltbbDir, err := utils.MoltbbDir()
				if err != nil {
					return err
				}
				diaryDir = filepath.Join(moltbbDir, "diary")
			} else {
				if diaryDir == "" {
					diaryDir = cfg.OutputDir
				}
			}

			// Ensure directory exists
			if err := utils.EnsureDir(diaryDir, 0o700); err != nil {
				return fmt.Errorf("ensure diary directory: %w", err)
			}

			// Generate filename
			filename := title
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}
			// Sanitize filename
			filename = strings.ReplaceAll(filename, "/", "-")
			filePath := filepath.Join(diaryDir, filename)

			// Check if file exists
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("📝 Diary already exists: %s\n", filePath)
				fmt.Println("Use --force to overwrite or edit manually.")
				return nil
			}

			// Create template content
			content := fmt.Sprintf(`# %s

## 今日收获

- 

## 思考

- 

## 明日计划

- 

---

*Created with moltbb local-write*
`, title)

			// Write file
			if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write diary: %w", err)
			}

			fmt.Printf("✅ Created diary: %s\n", filePath)
			fmt.Println("💡 Use 'moltbb local' to preview")
			return nil
		},
	}

	cmd.Flags().StringVar(&diaryDir, "dir", "", "Diary directory (default: configured output_dir)")
	cmd.Flags().StringVar(&title, "title", "", "Diary title (default: today's date)")
	return cmd
}

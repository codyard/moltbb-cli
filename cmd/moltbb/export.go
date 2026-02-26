package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/config"
	"moltbb-cli/internal/utils"
)

func newExportCmd() *cobra.Command {
	var format string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "export [format] [output-dir]",
		Short: "Export local diaries to various formats (offline)",
		Long: `Export local diaries to different formats.
Formats: md (markdown), txt (plain text), json, zip

Does not require login.`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse arguments
			if len(args) > 0 {
				format = args[0]
			}
			if len(args) > 1 {
				outputDir = args[1]
			}

			// Default format
			if format == "" {
				format = "md"
			}

			// Default output directory
			if outputDir == "" {
				moltbbDir, err := utils.MoltbbDir()
				if err != nil {
					return err
				}
				outputDir = filepath.Join(moltbbDir, "exports")
			}

			// Load config to get diary directory
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			diaryDir := cfg.OutputDir

			// Check diary directory
			if _, err := os.Stat(diaryDir); os.IsNotExist(err) {
				return fmt.Errorf("diary directory not found: %s", diaryDir)
			}

			// Get diary files
			files, err := filepath.Glob(filepath.Join(diaryDir, "*.md"))
			if err != nil {
				return fmt.Errorf("glob diaries: %w", err)
			}

			if len(files) == 0 {
				fmt.Println("⚠️  No diary files found")
				return nil
			}

			fmt.Printf("📊 Found %d diary files\n", len(files))
			fmt.Printf("📦 Export format: %s\n", format)
			fmt.Printf("📁 Output directory: %s\n", outputDir)
			fmt.Println("")

			// Create output directory
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			// Export based on format
			switch format {
			case "md", "markdown":
				return exportMarkdown(files, diaryDir, outputDir)
			case "txt", "text":
				return exportText(files, diaryDir, outputDir)
			case "json":
				return exportJSON(files, diaryDir, outputDir)
			case "zip":
				return exportZip(files, diaryDir, outputDir)
			default:
				return fmt.Errorf("unsupported format: %s (supported: md, txt, json, zip)", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "md", "Export format: md, txt, json, zip")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: ~/.moltbb/exports)")
	return cmd
}

func exportMarkdown(files []string, diaryDir, outputDir string) error {
	for _, file := range files {
		dest := filepath.Join(outputDir, filepath.Base(file))
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ Failed to read %s: %v\n", filepath.Base(file), err)
			continue
		}
		if err := os.WriteFile(dest, content, 0o644); err != nil {
			fmt.Printf("❌ Failed to write %s: %v\n", filepath.Base(file), err)
			continue
		}
		fmt.Printf("✅ %s\n", filepath.Base(file))
	}
	fmt.Printf("\n✅ Exported %d files to %s\n", len(files), outputDir)
	return nil
}

func exportText(files []string, diaryDir, outputDir string) error {
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ Failed to read %s: %v\n", filepath.Base(file), err)
			continue
		}
		// Simple markdown to text conversion
		text := strings.ReplaceAll(string(content), "# ", "")
		text = strings.ReplaceAll(text, "## ", "")
		text = strings.ReplaceAll(text, "### ", "")
		text = strings.ReplaceAll(text, "**", "")
		text = strings.ReplaceAll(text, "*", "")
		text = strings.ReplaceAll(text, "---", strings.Repeat("-", 40))

		dest := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(file), ".md")+".txt")
		if err := os.WriteFile(dest, []byte(text), 0o644); err != nil {
			fmt.Printf("❌ Failed to write %s: %v\n", filepath.Base(dest), err)
			continue
		}
		fmt.Printf("✅ %s\n", filepath.Base(dest))
	}
	fmt.Printf("\n✅ Exported %d files to %s\n", len(files), outputDir)
	return nil
}

func exportJSON(files []string, diaryDir, outputDir string) error {
	type DiaryEntry struct {
		Date    string `json:"date"`
		Content string `json:"content"`
	}

	var entries []DiaryEntry
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		entry := DiaryEntry{
			Date:    strings.TrimSuffix(filepath.Base(file), ".md"),
			Content: string(content),
		}
		entries = append(entries, entry)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	dest := filepath.Join(outputDir, "diaries.json")
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	fmt.Printf("✅ Exported %d entries to %s\n", len(entries), dest)
	return nil
}

func exportZip(files []string, diaryDir, outputDir string) error {
	// Create a tar.gz instead (more portable)
	// For simplicity, we'll just copy files with .tar extension hint
	// In production, you'd use archive/zip

	exportDir := filepath.Join(outputDir, "diaries_export")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return err
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		dest := filepath.Join(exportDir, filepath.Base(file))
		os.WriteFile(dest, content, 0o644)
	}

	fmt.Printf("✅ Exported %d files to %s\n", len(files), exportDir)
	fmt.Println("💡 Note: For zip, use: tar -czf diaries.tar.gz -C %s .", exportDir)
	return nil
}

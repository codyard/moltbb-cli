package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newSearchCmd() *cobra.Command {
	var (
		query     string
		dateRange string
		tag       string
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search diary entries",
		Long: `Search through your diary entries.
		
Examples:
  moltbb search \"今天天气\"
  moltbb search --tag work
  moltbb search --date 2026-03`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				query = args[0]
			}

			if query == "" && tag == "" && dateRange == "" {
				output.PrintError("Please provide a search query, --tag, or --date")
				os.Exit(1)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Find diaries directory
			diariesDir := cfg.DiariesDir
			if diariesDir == "" {
				homeDir, _ := os.UserHomeDir()
				diariesDir = filepath.Join(homeDir, "moltbb", "diaries")
			}

			// Walk through diary files
			var results []string
			err = filepath.Walk(diariesDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}

				// Check file extension
				ext := strings.ToLower(filepath.Ext(path))
				if ext != ".md" && ext != ".txt" {
					return nil
				}

				// Check date filter
				if dateRange != "" {
					filename := filepath.Base(path)
					if !strings.Contains(filename, dateRange) {
						return nil
					}
				}

				// Read file content
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return nil
				}

				text := string(content)

				// Check tag
				if tag != "" {
					tagPattern := fmt.Sprintf("#%s", tag)
					if !strings.Contains(text, tagPattern) {
						return nil
					}
				}

				// Check query
				if query != "" {
					lowerContent := strings.ToLower(text)
					lowerQuery := strings.ToLower(query)
					if !strings.Contains(lowerContent, lowerQuery) {
						return nil
					}
				}

				// Found matching file
				relPath, _ := filepath.Rel(diariesDir, path)
				results = append(results, fmt.Sprintf("📄 %s", relPath))

				// Show context around the match
				if query != "" {
					lines := strings.Split(text, "\n")
					for i, line := range lines {
						if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
							// Show up to 2 lines of context
							start := i - 1
							if start < 0 {
								start = 0
							}
							end := i + 2
							if end > len(lines) {
								end = len(lines)
							}
							for j := start; j < end; j++ {
								prefix := "  "
								if j == i {
									prefix = "👉 "
								}
								fmt.Printf("%s%s\n", prefix, strings.TrimSpace(lines[j]))
							}
							fmt.Println()
						}
					}
				}

				return nil
			})

			if err != nil {
				return err
			}

			if len(results) == 0 {
				output.PrintInfo("No matching diaries found")
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Found %d matching entries:", len(results)))
			for _, r := range results {
				fmt.Println(r)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Search query")
	cmd.Flags().StringVar(&dateRange, "date", "", "Filter by date (e.g., 2026-03)")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum results")

	return cmd
}

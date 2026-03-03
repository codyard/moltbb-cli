package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newStatsCmd() *cobra.Command {
	var (
		year       int
		month      string
		all        bool
	)

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show diary statistics",
		Long:  `Display statistics about your diary entries.
		
Examples:
  moltbb stats
  moltbb stats --year 2026
  moltbb stats --month 2026-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if _, err := os.Stat(diariesDir); os.IsNotExist(err) {
				output.PrintInfo("No diaries found. Diary files are stored in the cloud.")
				output.PrintInfo("Use 'moltbb local' to sync locally.")
				return nil
			}

			// Collect stats
			var totalEntries int
			var totalWords int
			var totalChars int
			var entriesByMonth map[string]int = make(map[string]int)
			var entriesByTag map[string]int = make(map[string]int)
			var dates []string

			now := time.Now()
			if year == 0 {
				year = now.Year()
			}

			err = filepath.Walk(diariesDir, func(path string, info os.FileInfo, err error) error {
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

				filename := filepath.Base(path)
				
				// Check year filter
				yearStr := fmt.Sprintf("%d", year)
				if !strings.Contains(filename, yearStr) {
					return nil
				}

				// Check month filter
				if month != "" {
					if !strings.Contains(filename, month) {
						return nil
					}
				}

				// Extract date from filename
				dates = append(dates, filename)

				// Read content
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return nil
				}

				text := string(content)
				totalEntries++
				totalWords += len(strings.Fields(text))
				totalChars += len(text)

				// Extract month
				if len(filename) >= 7 {
					monthKey := filename[:7]
					entriesByMonth[monthKey]++
				}

				// Extract tags
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					words := strings.Fields(line)
					for _, word := range words {
						if strings.HasPrefix(word, "#") {
							tag := strings.TrimPrefix(word, "#")
							tag = strings.Trim(tag, ".,!?;:")
							if tag != "" {
								entriesByTag[tag]++
							}
						}
					}
				}

				return nil
			})

			if err != nil {
				return err
			}

			// Calculate streak
			streak := calculateStreak(dates)

			// Print stats
			output.PrintSection("📊 Diary Statistics")
			
			fmt.Printf("Total Entries:     %d\n", totalEntries)
			fmt.Printf("Total Words:       %d\n", totalWords)
			fmt.Printf("Total Characters: %d\n", totalChars)
			if totalEntries > 0 {
				fmt.Printf("Avg Words/Entry:   %d\n", totalWords/totalEntries)
			}
			fmt.Printf("Current Streak:   %d days\n", streak)
			
			if len(entriesByMonth) > 0 {
				output.PrintSection("📅 Entries by Month")
				for month, count := range entriesByMonth {
					fmt.Printf("  %s: %d entries\n", month, count)
				}
			}

			if len(entriesByTag) > 0 {
				output.PrintSection("🏷️ Top Tags")
				// Sort tags by count
				for i := 0; i < len(entriesByTag) && i < 5; i++ {
					maxTag := ""
					maxCount := 0
					for tag, count := range entriesByTag {
						if count > maxCount {
							maxCount = count
							maxTag = tag
						}
					}
					if maxTag != "" {
						fmt.Printf("  #%s: %d\n", maxTag, maxCount)
						delete(entriesByTag, maxTag)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&year, "year", 0, "Year to show stats for")
	cmd.Flags().StringVar(&month, "month", "", "Month to show stats for (e.g., 2026-03)")
	cmd.Flags().BoolVar(&all, "all", false, "Show all time stats")

	return cmd
}

func calculateStreak(dates []string) int {
	if len(dates) == 0 {
		return 0
	}

	// Parse dates
	dateSet := make(map[string]bool)
	for _, d := range dates {
		// Extract date part from filename
		dateStr := strings.Split(d, ".")[0]
		dateSet[dateStr] = true
	}

	// Count consecutive days
	streak := 0
	current := time.Now()
	
	for {
		dateStr := current.Format("2006-01-02")
		if dateSet[dateStr] {
			streak++
			current = current.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak
}

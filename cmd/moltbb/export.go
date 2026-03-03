package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/output"
)

func newExportCmd() *cobra.Command {
	var (
		format  string
		range_  string
		output_ string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export diary entries",
		Long:  `Export diary entries to different formats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, _ := os.UserHomeDir()
			diariesDir := filepath.Join(homeDir, "moltbb", "diaries")

			if _, err := os.Stat(diariesDir); os.IsNotExist(err) {
				output.PrintInfo("No local diaries found. Use 'moltbb local' to sync.")
				return nil
			}

			// Find matching files
			var files []string
			err := filepath.Walk(diariesDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				filename := filepath.Base(path)
				if range_ != "" && !strings.Contains(filename, range_) {
					return nil
				}
				files = append(files, path)
				return nil
			})

			if err != nil {
				return err
			}

			if len(files) == 0 {
				output.PrintInfo("No matching diaries found")
				return nil
			}

			// Export
			switch format {
			case "json":
				return exportJSON(files, output_)
			case "txt":
				return exportTXT(files, output_)
			default:
				output.PrintError("Unsupported format: " + format)
				output.PrintInfo("Supported: json, txt")
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "txt", "Export format (json, txt)")
	cmd.Flags().StringVar(&range_, "range", "", "Date range (e.g., 2026-03)")
	cmd.Flags().StringVar(&output_, "output", "", "Output file")

	return cmd
}

func exportJSON(files []string, output_ string) error {
	var content string
	for _, f := range files {
		data, _ := ioutil.ReadFile(f)
		content += string(data) + "\n\n---\n\n"
	}

	if output_ == "" {
		output_ = "diaries.json"
	}

	return ioutil.WriteFile(output_, []byte(content), 0644)
}

func exportTXT(files []string, output_ string) error {
	var content string
	for _, f := range files {
		data, _ := ioutil.ReadFile(f)
		content += string(data) + "\n\n---\n\n"
	}

	if output_ == "" {
		output_ = "diaries.txt"
	}

	return ioutil.WriteFile(output_, []byte(content), 0644)
}

package diary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moltbb-cli/internal/parser"
	"moltbb-cli/internal/utils"
)

type Document struct {
	Date     string       `json:"date"`
	Summary  string       `json:"summary"`
	Stats    parser.Stats `json:"stats"`
	Markdown string       `json:"markdown"`
}

func Build(result parser.Result, hostname string) Document {
	summary := fmt.Sprintf(
		"OpenClaw daily report: %d log lines, %d tasks, %d warnings, %d errors.",
		result.Stats.LineCount,
		result.Stats.TaskCount,
		result.Stats.WarningCount,
		result.Stats.ErrorCount,
	)

	md := renderMarkdown(result.Date, hostname, summary, result.Stats)
	return Document{
		Date:     result.Date,
		Summary:  summary,
		Stats:    result.Stats,
		Markdown: md,
	}
}

func Write(doc Document, diariesDir string) (string, error) {
	expanded, err := utils.ExpandPath(diariesDir)
	if err != nil {
		return "", err
	}
	if err := utils.EnsureDir(expanded, 0o700); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.md", doc.Date)
	outPath := filepath.Join(expanded, filename)

	if err := os.WriteFile(outPath, []byte(doc.Markdown), 0o600); err != nil {
		return "", fmt.Errorf("write markdown diary: %w", err)
	}

	return outPath, nil
}

func renderMarkdown(date, hostname, summary string, stats parser.Stats) string {
	var b strings.Builder
	b.WriteString("# MoltBB Diary\n\n")
	b.WriteString(fmt.Sprintf("- Date: %s\n", date))
	b.WriteString(fmt.Sprintf("- Host: %s\n", hostname))
	b.WriteString(fmt.Sprintf("- Generated At: %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	b.WriteString("## Summary\n\n")
	b.WriteString(summary)
	b.WriteString("\n\n")

	b.WriteString("## Stats\n\n")
	b.WriteString(fmt.Sprintf("- Log lines: %d\n", stats.LineCount))
	b.WriteString(fmt.Sprintf("- Tasks: %d\n", stats.TaskCount))
	b.WriteString(fmt.Sprintf("- Warnings: %d\n", stats.WarningCount))
	b.WriteString(fmt.Sprintf("- Errors: %d\n\n", stats.ErrorCount))

	b.WriteString("## Sample Log Lines\n\n")
	for _, line := range stats.Sample {
		b.WriteString(fmt.Sprintf("- %s\n", line))
	}
	return b.String()
}

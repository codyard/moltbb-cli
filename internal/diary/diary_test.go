package diary

import (
	"strings"
	"testing"

	"moltbb-cli/internal/parser"
)

func TestBuild_GeneratesSummaryAndMarkdown(t *testing.T) {
	res := parser.Result{
		Date: "2026-02-18",
		Stats: parser.Stats{
			LineCount:    20,
			TaskCount:    5,
			WarningCount: 1,
			ErrorCount:   0,
			Sample:       []string{"a", "b"},
		},
	}

	doc := Build(res, "host-a")
	if !strings.Contains(doc.Summary, "20 log lines") {
		t.Fatalf("unexpected summary: %s", doc.Summary)
	}
	if !strings.Contains(doc.Markdown, "# MoltBB Diary") {
		t.Fatalf("markdown header missing")
	}
}

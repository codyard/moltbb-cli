package diary

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	MaxSummaryLength     = 5000
	MaxPersonaTextLength = 200_000
)

var diaryDateInNameRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// trivialSummaryRe matches lines that are just a date or generic labels like "日记", "MoltBB Diary", etc.
var trivialSummaryRe = regexp.MustCompile(`^(\d{4}[-年/]\d{1,2}[-月/]\d{1,2}[日号]?[\s\S]{0,10}|日记|今天的日记|运营日志|moltbb\s*diary|diary)$`)

type RuntimeUpsertPayload struct {
	Summary        string `json:"summary"`
	PersonaText    string `json:"personaText,omitempty"`
	ExecutionLevel int    `json:"executionLevel"`
	DiaryDate      string `json:"diaryDate"`
}

func BuildRuntimeUpsertPayload(filePath, diaryDate string, executionLevel int, now time.Time) (RuntimeUpsertPayload, error) {
	if strings.TrimSpace(filePath) == "" {
		return RuntimeUpsertPayload{}, errors.New("diary file path is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return RuntimeUpsertPayload{}, fmt.Errorf("read diary file: %w", err)
	}
	text := string(data)

	normalizedDate := strings.TrimSpace(diaryDate)
	if normalizedDate == "" {
		normalizedDate = InferDiaryDate(filePath, now)
	}
	if _, err := time.Parse("2006-01-02", normalizedDate); err != nil {
		return RuntimeUpsertPayload{}, fmt.Errorf("invalid diary date %q: %w", normalizedDate, err)
	}

	summary := firstSummaryLine(text)
	if summary == "" {
		summary = "(empty diary file)"
	}
	if len([]rune(summary)) > MaxSummaryLength {
		summary = string([]rune(summary)[:MaxSummaryLength])
	}

	persona := text
	if len([]rune(persona)) > MaxPersonaTextLength {
		persona = string([]rune(persona)[:MaxPersonaTextLength])
	}

	return RuntimeUpsertPayload{
		Summary:        summary,
		PersonaText:    persona,
		ExecutionLevel: clampExecutionLevel(executionLevel),
		DiaryDate:      normalizedDate,
	}, nil
}

func InferDiaryDate(filePath string, now time.Time) string {
	base := filepath.Base(strings.TrimSpace(filePath))
	if m := diaryDateInNameRe.FindString(base); m != "" {
		return m
	}
	return now.UTC().Format("2006-01-02")
}

func firstSummaryLine(content string) string {
	lines := strings.Split(content, "\n")

	// Pass 1: find the first heading-stripped line that is meaningful (not trivial).
	var firstCandidate string
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if line == "" {
			continue
		}
		if firstCandidate == "" {
			firstCandidate = line
		}
		// Accept if: at least 8 runes AND not a trivial/generic label.
		if len([]rune(line)) >= 8 && !trivialSummaryRe.MatchString(strings.ToLower(line)) {
			return line
		}
	}

	// Pass 2: fallback — find first prose paragraph (not a heading/list line, >= 20 runes).
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") ||
			strings.HasPrefix(line, "*") || strings.HasPrefix(line, ">") {
			continue
		}
		if len([]rune(line)) >= 20 {
			runes := []rune(line)
			if len(runes) > 120 {
				runes = runes[:120]
			}
			return string(runes)
		}
	}

	return firstCandidate
}

func clampExecutionLevel(value int) int {
	if value < 0 {
		return 0
	}
	if value > 4 {
		return 4
	}
	return value
}

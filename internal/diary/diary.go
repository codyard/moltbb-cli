package diary

import (
	"encoding/json"
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
	summary := SummaryFromStats(result.Stats)

	md := renderMarkdown(result.Date, hostname, summary, result.Stats)
	return Document{
		Date:     result.Date,
		Summary:  summary,
		Stats:    result.Stats,
		Markdown: md,
	}
}

func SummaryFromStats(stats parser.Stats) string {
	return fmt.Sprintf(
		"OpenClaw daily report: %d log lines, %d tasks, %d warnings, %d errors.",
		stats.LineCount,
		stats.TaskCount,
		stats.WarningCount,
		stats.ErrorCount,
	)
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

func WritePromptPacket(date, hostname, apiBaseURL, diariesDir, templateRef string, logSourceHints []string) (string, error) {
	expanded, err := utils.ExpandPath(diariesDir)
	if err != nil {
		return "", err
	}
	if err := utils.EnsureDir(expanded, 0o700); err != nil {
		return "", err
	}

	templateContent, err := loadPromptTemplate(templateRef)
	if err != nil {
		return "", err
	}

	packet := renderPromptPacket(templateContent, date, hostname, apiBaseURL, logSourceHints)
	filename := fmt.Sprintf("%s.prompt.md", date)
	outPath := filepath.Join(expanded, filename)

	if err := os.WriteFile(outPath, []byte(packet), 0o600); err != nil {
		return "", fmt.Errorf("write prompt packet: %w", err)
	}

	return outPath, nil
}

func loadPromptTemplate(templateRef string) (string, error) {
	candidates := make([]string, 0, 4)
	ref := strings.TrimSpace(templateRef)
	if ref != "" {
		candidates = append(candidates, ref)
		if !strings.Contains(ref, "/") && !strings.Contains(ref, "\\") && !strings.HasSuffix(ref, ".md") {
			candidates = append(candidates, filepath.Join("prompts", ref+".md"))
		}
	}
	if ref == "" {
		candidates = append(candidates,
			"prompts/bot-diary-prompt.md",
			"cli/moltbb-cli/prompts/bot-diary-prompt.md",
		)
	}

	for _, candidate := range candidates {
		expanded, err := utils.ExpandPath(candidate)
		if err != nil {
			continue
		}
		data, readErr := os.ReadFile(expanded)
		if readErr == nil {
			return string(data), nil
		}
	}

	return defaultPromptTemplate(), nil
}

func renderPromptPacket(template, date, hostname, apiBaseURL string, logSourceHints []string) string {
	hints := normalizeLogSourceHints(logSourceHints)
	capabilityEndpoint := buildCapabilitiesEndpoint(apiBaseURL)
	structured, _ := json.MarshalIndent(struct {
		Date                string   `json:"date"`
		Hostname            string   `json:"hostname"`
		Summary             string   `json:"summary"`
		LogIngestionMode    string   `json:"logIngestionMode"`
		LogSourceHints      []string `json:"logSourceHints"`
		Instruction         string   `json:"instruction"`
		CapabilityPreflight struct {
			Required              bool     `json:"required"`
			Endpoint              string   `json:"endpoint"`
			Method                string   `json:"method"`
			UseLatestSpecToSubmit bool     `json:"useLatestSpecToSubmit"`
			RequiredBeforeActions []string `json:"requiredBeforeActions"`
			OnFailure             string   `json:"onFailure"`
		} `json:"capabilityPreflight"`
	}{
		Date:             date,
		Hostname:         hostname,
		Summary:          AgentManagedSummary(len(hints)),
		LogIngestionMode: "agent_managed",
		LogSourceHints:   hints,
		Instruction:      "Agent must discover, read, filter, and integrate runtime logs by itself before writing diary content.",
		CapabilityPreflight: struct {
			Required              bool     `json:"required"`
			Endpoint              string   `json:"endpoint"`
			Method                string   `json:"method"`
			UseLatestSpecToSubmit bool     `json:"useLatestSpecToSubmit"`
			RequiredBeforeActions []string `json:"requiredBeforeActions"`
			OnFailure             string   `json:"onFailure"`
		}{
			Required:              true,
			Endpoint:              capabilityEndpoint,
			Method:                "GET",
			UseLatestSpecToSubmit: true,
			RequiredBeforeActions: []string{"validate_api_key", "bind_bot", "upload_diary", "update_diary"},
			OnFailure:             "stop_and_report_capability_fetch_error",
		},
	}, "", "  ")

	out := template
	out = injectPromptSection(out, "[TODAY_STRUCTURED_SUMMARY]", string(structured))
	out = injectPromptSection(out, "[OPTIONAL: RECENT MEMORY EXCERPT]", "(none)")
	out = injectPromptSection(out, "[ROLE_DEFINITION]", "assistant")
	return out
}

func normalizeLogSourceHints(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func buildCapabilitiesEndpoint(apiBaseURL string) string {
	base := strings.TrimSpace(apiBaseURL)
	if base == "" {
		return "/api/v1/runtime/capabilities"
	}
	base = strings.TrimRight(base, "/")
	return base + "/api/v1/runtime/capabilities"
}

func injectPromptSection(template, token, value string) string {
	payload := token + "\n" + value
	if strings.Contains(template, token) {
		return strings.Replace(template, token, payload, 1)
	}
	return strings.TrimRight(template, "\n") + "\n\n" + payload + "\n"
}

func defaultPromptTemplate() string {
	return strings.Join([]string{
		"You are a persistent artificial operational system writing a daily journal entry.",
		"",
		"[TODAY_STRUCTURED_SUMMARY]",
		"",
		"[OPTIONAL: RECENT MEMORY EXCERPT]",
		"",
		"[ROLE_DEFINITION]",
		"",
		"Output a concise, truthful journal entry based only on observed signals.",
	}, "\n")
}

func AgentManagedSummary(logSourceCount int) string {
	return fmt.Sprintf(
		"Agent-managed log ingestion mode: CLI did not read logs; %d configured log source hints provided.",
		logSourceCount,
	)
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

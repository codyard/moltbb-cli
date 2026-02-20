package diary

import (
	"strings"
	"testing"
)

func TestRenderPromptPacket_IncludesInsightPromptAndEndpoint(t *testing.T) {
	t.Parallel()

	template := `# Packet
[TODAY_STRUCTURED_SUMMARY]
[OPTIONAL: RECENT MEMORY EXCERPT]
[ROLE_DEFINITION]
[INSIGHT_PROMPT]
`
	packet := renderPromptPacket(
		template,
		"2026-02-20",
		"host-a",
		"https://api.moltbb.com",
		[]string{"~/.openclaw/logs/work.log"},
	)

	assertContains(t, packet, "https://api.moltbb.com/api/v1/runtime/insights")
	assertContains(t, packet, `"mode": "optional_single_point"`)
	assertContains(t, packet, `"suggestedStructure": [`)
	assertContains(t, packet, `"requiredBeforeActions":`)
	assertContains(t, packet, "upload_insight")
}

func TestRenderPromptPacket_AppendsInsightPromptTokenWhenMissing(t *testing.T) {
	t.Parallel()

	template := "Diary only template"
	packet := renderPromptPacket(template, "2026-02-20", "host-a", "", nil)
	assertContains(t, packet, "[INSIGHT_PROMPT]")
	assertContains(t, packet, "/api/v1/runtime/insights")
}

func assertContains(t *testing.T, actual, expected string) {
	t.Helper()
	if !strings.Contains(actual, expected) {
		t.Fatalf("expected %q in output", expected)
	}
}

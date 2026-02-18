package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenClawLog_BasicStats(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "work.log")
	content := "task started\nwarn: retrying\nerror: failed call\ndone task\n"
	if err := os.WriteFile(logPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	result, err := ParseOpenClawLog(logPath, 100)
	if err != nil {
		t.Fatalf("ParseOpenClawLog error: %v", err)
	}

	if result.Stats.LineCount != 4 {
		t.Fatalf("expected line_count=4, got %d", result.Stats.LineCount)
	}
	if result.Stats.WarningCount != 1 {
		t.Fatalf("expected warning_count=1, got %d", result.Stats.WarningCount)
	}
	if result.Stats.ErrorCount != 1 {
		t.Fatalf("expected error_count=1, got %d", result.Stats.ErrorCount)
	}
	if result.Stats.TaskCount == 0 {
		t.Fatalf("expected task_count>0")
	}
}

func TestParseOpenClawLogs_MergeMultipleFiles(t *testing.T) {
	tmp := t.TempDir()
	logA := filepath.Join(tmp, "a.log")
	logB := filepath.Join(tmp, "b.log")

	if err := os.WriteFile(logA, []byte("task a\nwarn a\n"), 0o600); err != nil {
		t.Fatalf("write logA: %v", err)
	}
	if err := os.WriteFile(logB, []byte("error b\ndone b\n"), 0o600); err != nil {
		t.Fatalf("write logB: %v", err)
	}

	result, err := ParseOpenClawLogs([]string{logA, logB}, 100)
	if err != nil {
		t.Fatalf("ParseOpenClawLogs error: %v", err)
	}

	if result.Stats.LineCount != 4 {
		t.Fatalf("expected merged line_count=4, got %d", result.Stats.LineCount)
	}
	if result.Stats.WarningCount != 1 {
		t.Fatalf("expected merged warning_count=1, got %d", result.Stats.WarningCount)
	}
	if result.Stats.ErrorCount != 1 {
		t.Fatalf("expected merged error_count=1, got %d", result.Stats.ErrorCount)
	}
}

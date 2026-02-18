package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Stats struct {
	LineCount    int      `json:"line_count"`
	ErrorCount   int      `json:"error_count"`
	WarningCount int      `json:"warning_count"`
	TaskCount    int      `json:"task_count"`
	Sample       []string `json:"sample"`
}

type Result struct {
	Date  string `json:"date"`
	Stats Stats  `json:"stats"`
}

func ParseOpenClawLog(path string, maxLines int) (Result, error) {
	if maxLines <= 0 {
		maxLines = 2000
	}

	f, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	res := Result{Date: time.Now().UTC().Format("2006-01-02")}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		res.Stats.LineCount++
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "error"), strings.Contains(lower, "failed"):
			res.Stats.ErrorCount++
		case strings.Contains(lower, "warn"):
			res.Stats.WarningCount++
		}

		if strings.Contains(lower, "task") || strings.Contains(lower, "done") || strings.Contains(lower, "complete") {
			res.Stats.TaskCount++
		}

		if len(res.Stats.Sample) < 8 {
			res.Stats.Sample = append(res.Stats.Sample, line)
		}

		if res.Stats.LineCount >= maxLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("scan log file: %w", err)
	}

	return res, nil
}

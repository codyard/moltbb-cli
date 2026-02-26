package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Color output utilities

var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()

	// Check if terminal supports color
	supportsColor = color.NoColor == false
)

// PrintSuccess prints success message
func PrintSuccess(msg string) {
	if supportsColor {
		fmt.Fprintln(os.Stdout, Success("✅")+" "+msg)
	} else {
		fmt.Fprintln(os.Stdout, "✓ "+msg)
	}
}

// PrintError prints error message
func PrintError(msg string) {
	if supportsColor {
		fmt.Fprintln(os.Stderr, Error("❌")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "✗ "+msg)
	}
}

// PrintWarning prints warning message
func PrintWarning(msg string) {
	if supportsColor {
		fmt.Fprintln(os.Stdout, Warning("⚠️")+" "+msg)
	} else {
		fmt.Fprintln(os.Stdout, "⚠ "+msg)
	}
}

// PrintInfo prints info message
func PrintInfo(msg string) {
	if supportsColor {
		fmt.Fprintln(os.Stdout, Info("ℹ️")+" "+msg)
	} else {
		fmt.Fprintln(os.Stdout, "ℹ "+msg)
	}
}

// PrintSection prints a section header
func PrintSection(title string) {
	if supportsColor {
		fmt.Fprintln(os.Stdout, Bold("\n━━ "+title+" ━━\n"))
	} else {
		fmt.Fprintln(os.Stdout, "\n== "+title+" ==\n")
	}
}

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptString(reader *bufio.Reader, label, defaultValue string) (string, error) {
	display := label
	if strings.TrimSpace(defaultValue) != "" {
		display = fmt.Sprintf("%s [%s]", label, defaultValue)
	}
	fmt.Printf("%s: ", display)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

func PromptYesNo(reader *bufio.Reader, label string, defaultYes bool) (bool, error) {
	defaultText := "y/N"
	if defaultYes {
		defaultText = "Y/n"
	}
	fmt.Printf("%s [%s]: ", label, defaultText)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultYes, nil
	}
	if line == "y" || line == "yes" {
		return true, nil
	}
	if line == "n" || line == "no" {
		return false, nil
	}
	return false, fmt.Errorf("invalid input: %s", line)
}

func PromptSecret(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}

	fmt.Println("(warning: hidden input unavailable on this terminal)")
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Prompt reads a single line from stdin, trimmed.
func Prompt(label string) (string, error) {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// PromptPassword reads a password from stdin without echoing.
// If stdin is not a terminal, it falls back to a line read (useful in CI).
func PromptPassword(label string) (string, error) {
	fmt.Print(label)
	fd := int(syscall.Stdin)
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return Prompt("")
}

// Confirm asks a yes/no question. Returns true on y/yes (case-insensitive).
func Confirm(label string) bool {
	ans, err := Prompt(label)
	if err != nil {
		return false
	}
	switch strings.ToLower(ans) {
	case "y", "yes":
		return true
	}
	return false
}

// IsTerminal reports whether stdin is a tty, so non-interactive mode can skip prompts.
func IsStdinTerminal() bool {
	return term.IsTerminal(int(syscall.Stdin))
}

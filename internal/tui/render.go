// Package tui provides terminal presentation helpers for the GoLingua CLI.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// renderShell composes exactly height rows of one full-screen shell frame:
// header, tab bar, runtime info, body block, separator, prompt, footer.
// The caller is responsible for sanitizing text and pre-wrapping body.
func renderShell(width, height, screen int, focus, runtimePath, body, prompt, footer string) string {
	if width < 40 || height < 12 {
		if width < 1 {
			return ""
		}
		notice := fmt.Sprintf("Terminal too small (%dx%d); resize to at least 40x12.", width, height)
		return ansi.Truncate(notice, width, "")
	}

	rows := make([]string, 0, height)

	rows = append(rows, ansi.Truncate(fmt.Sprintf("GoLingua Focus: %s", focus), width, ""))

	labels := [...]string{"F1 Translate", "F2 Forms", "F3 Concepts", "F4 Languages", "F5 Settings"}
	parts := make([]string, 0, len(labels))
	for i, label := range labels {
		if i == screen {
			parts = append(parts, "["+label+"]")
		} else {
			parts = append(parts, label)
		}
	}
	rows = append(rows, ansi.Truncate(strings.Join(parts, "  "), width, ""))

	rows = append(rows, ansi.Truncate(fmt.Sprintf("Runtime: %s", runtimePath), width, ""))

	bodyRows := height - 6
	bodyLines := strings.Split(body, "\n")
	for i := 0; i < bodyRows; i++ {
		line := ""
		if i < len(bodyLines) {
			line = bodyLines[i]
		}
		rows = append(rows, ansi.Truncate(line, width, ""))
	}

	rows = append(rows, ansi.Truncate(strings.Repeat("─", width), width, ""))

	rows = append(rows, ansi.Truncate(fmt.Sprintf("> %s", prompt), width, ""))

	rows = append(rows, ansi.Truncate(footer, width, ""))

	return strings.Join(rows, "\n")
}

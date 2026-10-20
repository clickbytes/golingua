// Package cli provides a deliberately small, fail-closed bootstrap CLI.
// No language plugin, translation, generation, or inference is available yet.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	Version         = "0.0.0-bootstrap"
	ExitOK          = 0
	ExitUsage       = 2
	ExitUnsupported = 3
)

type capabilities struct {
	Version   string   `json:"version"`
	Status    string   `json:"status"`
	Languages []string `json:"languages"`
	Features  []string `json:"features"`
	Ready     bool     `json:"ready"`
}

func writeJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	return enc.Encode(value)
}

// Run returns a process exit code without mutating global state.
func Run(args []string, stdout, stderr io.Writer) int {
	if stdout == nil || stderr == nil {
		return ExitUsage
	}
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help")) {
		_, _ = fmt.Fprint(stdout, "golingua bootstrap\nUsage: golingua --help | --version | capabilities --json\nTranslation and language plugins: NOT IMPLEMENTED\n")
		return ExitOK
	}
	if len(args) == 1 && args[0] == "--version" {
		_, _ = fmt.Fprintln(stdout, Version)
		return ExitOK
	}
	if len(args) == 2 && args[0] == "capabilities" && args[1] == "--json" {
		if err := writeJSON(stdout, capabilities{Version: Version, Status: "unsupported", Languages: []string{}, Features: []string{}, Ready: false}); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize capabilities: %v\n", err)
			return ExitUsage
		}
		return ExitOK
	}
	if len(args) > 0 && args[0] == "translate" {
		_, _ = fmt.Fprintln(stderr, "LANGUAGE_RUNTIME_UNAVAILABLE: sentence translation is not implemented in the bootstrap")
		return ExitUnsupported
	}
	_, _ = fmt.Fprintf(stderr, "INVALID_COMMAND: %s\n", strings.Join(args, " "))
	return ExitUsage
}

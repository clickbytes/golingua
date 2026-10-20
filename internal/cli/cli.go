// Package cli provides a deliberately small, fail-closed bootstrap CLI.
// No language plugin, translation, generation, or inference is available yet.
package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/status"
)

const (
	Version           = "0.0.0-bootstrap"
	ExitError         = 1
	ExitOK            = 0
	ExitUsage         = 2
	ExitUnsupported   = 3
	ExitResourceLimit = 4
	ExitUnknown       = 5
	ExitAmbiguous     = 6
	ExitIncomplete    = 7
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

func exitCode(kind status.OutcomeKind) int {
	switch kind {
	case status.OutcomeOK:
		return ExitOK
	case status.OutcomeInvalid:
		return ExitUsage
	case status.OutcomeUnsupported:
		return ExitUnsupported
	case status.OutcomeResourceLimit:
		return ExitResourceLimit
	case status.OutcomeUnknown:
		return ExitUnknown
	case status.OutcomeAmbiguous:
		return ExitAmbiguous
	case status.OutcomeIncomplete:
		return ExitIncomplete
	default:
		return ExitError
	}
}

func writeResult(stdout, stderr io.Writer, result status.Result) int {
	if err := result.Validate(); err != nil {
		_, _ = fmt.Fprintf(stderr, "INTERNAL_RESULT_ERROR: %v\n", err)
		return ExitError
	}
	if err := writeJSON(stdout, result); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize result: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

func diagnosticResult(kind status.OutcomeKind, code, message, path string) status.Result {
	result, err := status.NewResult(kind, "", []status.Diagnostic{{
		Code:     code,
		Message:  message,
		Path:     path,
		Severity: status.SeverityError,
	}}, nil)
	if err != nil {
		return status.Result{
			Status: status.OutcomeError,
			Diagnostics: []status.Diagnostic{{
				Code:     "INTERNAL_RESULT_ERROR",
				Message:  "could not construct command result",
				Severity: status.SeverityError,
			}},
		}
	}
	return result
}

// Run returns a process exit code without mutating global state.
func Run(args []string, stdout, stderr io.Writer) int {
	if stdout == nil {
		if stderr != nil {
			_, _ = fmt.Fprintln(stderr, "INFRASTRUCTURE_ERROR: stdout writer is nil")
		}
		return ExitError
	}
	if stderr == nil {
		stderr = io.Discard
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
			return ExitError
		}
		return ExitOK
	}
	if len(args) > 0 && args[0] == "translate" {
		result := diagnosticResult(status.OutcomeUnsupported, "LANGUAGE_RUNTIME_UNAVAILABLE", "sentence translation is not implemented in the bootstrap", "translate")
		return writeResult(stdout, stderr, result)
	}
	result := diagnosticResult(status.OutcomeInvalid, "INVALID_COMMAND", "unknown or invalid command", "command")
	return writeResult(stdout, stderr, result)
}

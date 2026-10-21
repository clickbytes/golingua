// Package cli provides a deliberately small, fail-closed bootstrap CLI.
// Translation is available only when verified language runtimes are configured.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
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

type translateOptions struct {
	from, to, text, previewID string
	runtimeConfig             string
	json                      bool
}

// EngineLoader loads one explicitly selected runtime configuration. The
// executable supplies trusted, compiled providers and stage adapters through
// this boundary; JSON data cannot install code.
type EngineLoader func(string) (*engine.Engine, error)

func parseTranslate(args []string) (translateOptions, string) {
	var options translateOptions
	seen := make(map[string]bool, 5)
	for i := 1; i < len(args); i++ {
		flag := args[i]
		if flag == "--json" {
			if seen[flag] {
				return options, "duplicate --json flag"
			}
			seen[flag] = true
			options.json = true
			continue
		}
		if flag != "--from" && flag != "--to" && flag != "--text" && flag != "--preview-id" && flag != "--runtime-config" {
			return options, fmt.Sprintf("unknown or malformed translate argument %q", flag)
		}
		if seen[flag] {
			return options, fmt.Sprintf("duplicate %s flag", flag)
		}
		seen[flag] = true
		if i+1 >= len(args) || args[i+1] == "--json" || args[i+1] == "--from" || args[i+1] == "--to" || args[i+1] == "--text" || args[i+1] == "--preview-id" || args[i+1] == "--runtime-config" {
			return options, fmt.Sprintf("missing value for %s", flag)
		}
		i++
		value := args[i]
		switch flag {
		case "--from":
			options.from = value
		case "--to":
			options.to = value
		case "--text":
			options.text = value
		case "--preview-id":
			options.previewID = value
		case "--runtime-config":
			options.runtimeConfig = value
		}
	}
	if !options.json {
		return options, "translate requires --json"
	}
	for _, flag := range []string{"--from", "--to", "--text"} {
		if !seen[flag] {
			return options, fmt.Sprintf("missing required %s flag", flag)
		}
	}
	if options.from == "" || options.to == "" || options.text == "" {
		return options, "--from, --to, and --text values must not be empty"
	}
	if seen["--preview-id"] && options.previewID == "" {
		return options, "--preview-id value must not be empty"
	}
	if seen["--runtime-config"] && options.runtimeConfig == "" {
		return options, "--runtime-config value must not be empty"
	}
	return options, ""
}

func runTranslate(args []string, stdout, stderr io.Writer, loadEngine EngineLoader) int {
	options, parseError := parseTranslate(args)
	if parseError != "" {
		result := diagnosticResult(status.OutcomeInvalid, "INVALID_TRANSLATE_ARGUMENTS", parseError, "translate")
		return writeResult(stdout, stderr, result)
	}
	var translator *engine.Engine
	var err error
	if options.runtimeConfig != "" {
		if loadEngine == nil {
			return writeResult(stdout, stderr, diagnosticResult(status.OutcomeUnsupported, "RUNTIME_LOADER_UNAVAILABLE", "this executable cannot load configured language runtimes", "runtime_config"))
		}
		translator, err = loadEngine(options.runtimeConfig)
	} else {
		translator, err = engine.New(engine.Config{})
	}
	if err != nil {
		var resultError interface{ Result() status.Result }
		if errors.As(err, &resultError) {
			return writeResult(stdout, stderr, resultError.Result())
		}
		failure := diagnosticResult(status.OutcomeError, "ENGINE_SETUP_ERROR", "translation engine setup failed unexpectedly", "translate")
		return writeResult(stdout, stderr, failure)
	}
	result, err := translator.Translate(context.Background(), engine.Request{
		From: features.LanguageID(options.from), To: features.LanguageID(options.to), Text: options.text, PreviewID: options.previewID,
	})
	if err != nil {
		var resultError interface{ Result() status.Result }
		if errors.As(err, &resultError) {
			return writeResult(stdout, stderr, resultError.Result())
		}
		failure := diagnosticResult(status.OutcomeError, "TRANSLATION_RUNTIME_ERROR", "translation failed unexpectedly", "translate")
		return writeResult(stdout, stderr, failure)
	}
	if err := result.Validate(); err != nil {
		_, _ = fmt.Fprintf(stderr, "INTERNAL_RESULT_ERROR: %v\n", err)
		return ExitError
	}
	if err := writeJSON(stdout, result); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize translation result: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

// Run returns a process exit code without mutating global state.
func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithInput(args, bytes.NewReader(nil), stdout, stderr)
}

// RunWithInput is Run with an explicit stdin reader for commands that consume
// bounded JSON documents.
func RunWithInput(args []string, input io.Reader, stdout, stderr io.Writer) int {
	return RunWithInputAndEngineLoader(args, input, stdout, stderr, nil)
}

// RunWithInputAndEngineLoader is RunWithInput with an optional trusted loader
// for explicitly configured language runtimes.
func RunWithInputAndEngineLoader(args []string, input io.Reader, stdout, stderr io.Writer, loadEngine EngineLoader) int {
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
		_, _ = fmt.Fprint(stdout, "golingua bootstrap\nUsage: golingua --help | --version | capabilities --json | meaning canonicalize --json < meaning.json | meaning project --json < meaning.json | translate --from <lang> --to <lang> --text <text> --json [--preview-id <id>] [--runtime-config <path>]\nNo language providers are configured by default; translate returns unsupported until runtimes are configured.\n")
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
	if len(args) == 3 && args[0] == "meaning" && args[1] == "canonicalize" && args[2] == "--json" {
		if input == nil {
			result := diagnosticResult(status.OutcomeError, "INPUT_UNAVAILABLE", "stdin reader is unavailable", "meaning")
			return writeResult(stdout, stderr, result)
		}
		data, err := io.ReadAll(io.LimitReader(input, meaning.MaxMeaningBytes+1))
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read meaning JSON: %v\n", err)
			return ExitError
		}
		parsed, err := meaning.DecodeMeaning(data, meaning.DefaultBudget())
		if err != nil {
			var typed *meaning.StatusError
			if errors.As(err, &typed) {
				return writeResult(stdout, stderr, typed.Result())
			}
			result := diagnosticResult(status.OutcomeError, "MEANING_RUNTIME_ERROR", "meaning validation failed unexpectedly", "meaning")
			return writeResult(stdout, stderr, result)
		}
		canonical, err := meaning.CanonicalJSON(parsed, meaning.DefaultBudget())
		if err != nil {
			var typed *meaning.StatusError
			if errors.As(err, &typed) {
				return writeResult(stdout, stderr, typed.Result())
			}
			result := diagnosticResult(status.OutcomeError, "MEANING_RUNTIME_ERROR", "meaning canonicalization failed unexpectedly", "meaning")
			return writeResult(stdout, stderr, result)
		}
		if _, err := fmt.Fprintln(stdout, string(canonical)); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write canonical meaning: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	if len(args) == 3 && args[0] == "meaning" && args[1] == "project" && args[2] == "--json" {
		if input == nil {
			result := diagnosticResult(status.OutcomeError, "INPUT_UNAVAILABLE", "stdin reader is unavailable", "meaning")
			return writeResult(stdout, stderr, result)
		}
		data, err := io.ReadAll(io.LimitReader(input, meaning.MaxMeaningBytes+1))
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read meaning JSON: %v\n", err)
			return ExitError
		}
		parsed, err := meaning.DecodeMeaning(data, meaning.DefaultBudget())
		if err != nil {
			var typed *meaning.StatusError
			if errors.As(err, &typed) {
				return writeResult(stdout, stderr, typed.Result())
			}
			result := diagnosticResult(status.OutcomeError, "MEANING_RUNTIME_ERROR", "meaning validation failed unexpectedly", "meaning")
			return writeResult(stdout, stderr, result)
		}
		projection, err := meaning.ProjectAsserted(parsed, meaning.DefaultBudget())
		if err != nil {
			var typed *meaning.StatusError
			if errors.As(err, &typed) {
				return writeResult(stdout, stderr, typed.Result())
			}
			result := diagnosticResult(status.OutcomeError, "MEANING_PROJECTION_ERROR", "asserted projection failed unexpectedly", "meaning")
			return writeResult(stdout, stderr, result)
		}
		if err := writeJSON(stdout, projection); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize meaning projection: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	if len(args) > 0 && args[0] == "translate" {
		return runTranslate(args, stdout, stderr, loadEngine)
	}
	result := diagnosticResult(status.OutcomeInvalid, "INVALID_COMMAND", "unknown or invalid command", "command")
	return writeResult(stdout, stderr, result)
}

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

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	Version           = "0.1.0-alpha.1"
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
type EngineLoader = app.EngineLoader

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

func runTranslate(args []string, stdout, stderr io.Writer, service *app.Service) int {
	options, parseError := parseTranslate(args)
	if parseError != "" {
		result := diagnosticResult(status.OutcomeInvalid, "INVALID_TRANSLATE_ARGUMENTS", parseError, "translate")
		return writeResult(stdout, stderr, result)
	}
	response := service.Translate(context.Background(), app.TranslateRequest{
		From: options.from, To: options.to, Text: options.text,
		PreviewID: options.previewID, RuntimeConfig: options.runtimeConfig,
	})
	if response.Failure != nil {
		return writeResult(stdout, stderr, *response.Failure)
	}
	if response.Result == nil {
		return writeResult(stdout, stderr, diagnosticResult(status.OutcomeError, "ENGINE_SETUP_ERROR", "translation engine setup failed unexpectedly", "translate"))
	}
	result := *response.Result
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
	return RunWithInputAndLoaders(args, input, stdout, stderr, loadEngine, nil)
}

// RunWithInputAndLoaders adds an optional CORE loader for typed question
// execution while keeping language runtime loading behind EngineLoader.
func RunWithInputAndLoaders(args []string, input io.Reader, stdout, stderr io.Writer, loadEngine EngineLoader, loadCore CoreLoader) int {
	service := app.NewService(app.Config{LoadEngine: loadEngine})
	return RunWithInputAndService(args, input, stdout, stderr, service, loadEngine, loadCore)
}

// RunWithInputAndService lets an executable share one application service
// across its CLI command handlers. loadEngine/loadCore remain available for
// the legacy question command while it is outside the service boundary.
func RunWithInputAndService(args []string, input io.Reader, stdout, stderr io.Writer, service *app.Service, loadEngine EngineLoader, loadCore CoreLoader) int {
	if service == nil {
		service = app.NewService(app.Config{LoadEngine: loadEngine})
	}
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
		_, _ = fmt.Fprint(stdout, "golingua v0 preview\nUsage: golingua --help | --version | capabilities --json | meaning canonicalize --json < meaning.json | meaning project --json < meaning.json | question --runtime-config <path> [--language pl|en] --json < question.json | translate --from <lang> --to <lang> --text <text> --json [--preview-id <id>] [--runtime-config <path>] | forms --language pl|en (--lexeme <id> --feature name=value... | --surface <form>) --json | catalog --summary|--details --json | snapshot --preview|--apply --store <path> --json\nAlgebra: golingua algebra check --input <file> [--metadata <package.json> ...] --json (or stdin with --input -); algebra build --metadata <package.json> [--metadata ...] [--input <construction.json>] --json (stdin if omitted or --input -); algebra synthesize --mode construct --metadata <package.json> [--metadata ...] --goal <goal.json> --json (stdin with --goal -); algebra synthesize --mode search --metadata <package.json> [--metadata ...] --goal <goal.json> [--search-limit name=value ...] --json (stdin with --goal -); algebra verify-goal --goal <goal.json> --candidate <candidate.json> --certificate <certificate.json> --metadata <package.json> [--metadata ...] [--verify-limit name=value ...] --json (all three inputs are named files); algebra catalog export --json (compiled inert catalog inventory); algebra facts validate --metadata <package.json> [--metadata ...] [--input <facts.json>] --json (stdin if omitted or --input -); algebra evaluate --metadata <package.json> [--metadata ...] --expression <expression.json> --facts <facts.json> --select <selection.json> --context-id <id> --profile positive-facts-v1 [--at <UTC>] [--evaluation-limit name=value ...] --json (finite positive support; no negative/world-truth inference); algebra export --metadata <package.json> [--metadata ...] [--input <request.json>] --json [--no-cache] (stdin if omitted or --input -); algebra inspect [--input <artifact.json>] --json (structural only, no metadata verification; stdin if omitted or --input -); algebra replay --metadata <package.json> [--metadata ...] [--input <request.json>] --json (stdin if omitted or --input -); algebra normalize --metadata <package.json> [--metadata ...] [--input <request.json>] --json (stdin if omitted or --input -); algebra verify-rewrite --metadata <package.json> [--metadata ...] [--input <verification.json>] --json (stdin if omitted or --input -); algebra metadata inspect|validate --metadata <package.json> [--metadata <package.json> ...] --json. Export and replay use explicitly selected metadata; replay checks artifact claims against independently supplied metadata and request context. Inspect checks structural integrity only. Normalize applies only the rewrite rules explicitly listed in a finite profile; verify-rewrite checks a submitted certificate against selected metadata. Verify-goal checks producer claims against selected metadata and candidate input; it does not evaluate facts or prove world truth. Synthesize construct results carry independent verification before export; search candidates are independently verified and still do not establish world truth. Candidate retention bytes and the separate 16 MiB whole-response cap are distinct. Metadata files are explicit inputs; validation is structural and does not prove world truth.\nquestion executes typed MeaningGraph requests against caller-supplied FactSets; it does not parse natural language. With --language, supported answers use verified PL/EN sentence routes and retain proof evidence.\nCompiled PL/EN sentence adapters cover limited seeing, hearing, looking, and reading (present 3sg/3pl and past 3sg/3pl), bounded present negation for seeing/hearing/looking in 3sg/3pl (EN do/does not; PL nie), and receiving (past 3sg/3pl); Polish plural past requires known subject number and agreement class. Reading supports English simple past read and Polish czytać past forms in third-person singular and plural; plural English read remains tense-ambiguous. Identification classifies a compatible known referent (for example, cat or dog) and attributes black to that referent in the tested PL/EN routes. English realization retains this/that/it alternatives for unresolved reference and this/that for demonstrative reference, with explicit reference assumptions; Polish to/ten may carry a declared deixis loss when source proximity is not preserved. A bounded positive composition route supports a present seeing event with a black subject and an object, both singular, including compatible known nouns. Other modifier combinations, tenses, numbers, and unsupported or missing lexical forms are refused. These routes do not resolve entity identity or animal sex. translate requires --runtime-config with an active snapshot.\n")
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
		return runTranslate(args, stdout, stderr, service)
	}
	if len(args) > 0 && args[0] == "algebra" {
		return runAlgebra(args, input, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "question" {
		return runQuestion(args, input, stdout, stderr, loadEngine, loadCore)
	}
	result := diagnosticResult(status.OutcomeInvalid, "INVALID_COMMAND", "unknown or invalid command", "command")
	return writeResult(stdout, stderr, result)
}

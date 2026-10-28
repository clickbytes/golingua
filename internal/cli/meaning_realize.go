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
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const maxMeaningRealizeOutputBytes = 8 << 20

type meaningRealizeRequest struct {
	Meaning meaning.Meaning        `json:"meaning"`
	Context meaning.RequestContext `json:"context"`
}

type meaningRealizeOptions struct {
	language, runtimeConfig string
}

type meaningRealizeOutput struct {
	Status       status.OutcomeKind          `json:"status"`
	Language     string                      `json:"language"`
	Alternatives []engine.SurfaceAlternative `json:"alternatives"`
	Diagnostics  []status.Diagnostic         `json:"diagnostics"`
}

func parseMeaningRealizeArgs(args []string) (meaningRealizeOptions, string) {
	if len(args) < 2 || args[0] != "meaning" || args[1] != "realize" {
		return meaningRealizeOptions{}, "usage: meaning realize --language pl|en --runtime-config <path> --json"
	}
	var options meaningRealizeOptions
	seenLanguage, seenConfig, seenJSON := false, false, false
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--language":
			if seenLanguage || i+1 >= len(args) || len(args[i+1]) >= 2 && args[i+1][:2] == "--" {
				return options, "--language requires one language ID: pl or en"
			}
			seenLanguage = true
			i++
			options.language = args[i]
			if options.language != "pl" && options.language != "en" {
				return options, "--language must be pl or en"
			}
		case "--runtime-config":
			if seenConfig || i+1 >= len(args) || len(args[i+1]) >= 2 && args[i+1][:2] == "--" {
				return options, "--runtime-config requires one path and may appear only once"
			}
			seenConfig = true
			i++
			options.runtimeConfig = args[i]
			if options.runtimeConfig == "" {
				return options, "--runtime-config value must not be empty"
			}
		case "--json":
			if seenJSON {
				return options, "duplicate --json flag"
			}
			seenJSON = true
		default:
			return options, fmt.Sprintf("unknown or malformed meaning realize argument %q", args[i])
		}
	}
	if !seenLanguage || !seenConfig || !seenJSON {
		return options, "meaning realize requires --language pl|en, --runtime-config <path>, and --json"
	}
	return options, ""
}

func runMeaningRealize(args []string, input io.Reader, stdout, stderr io.Writer, loader EngineLoader) int {
	options, parseErr := parseMeaningRealizeArgs(args)
	if parseErr != "" {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeInvalid}, parseErr)
	}
	if input == nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError}, "stdin reader is unavailable")
	}
	data, err := io.ReadAll(io.LimitReader(input, meaning.MaxMeaningBytes+1))
	if err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError}, "request could not be read")
	}
	if len(data) == 0 || len(data) > meaning.MaxMeaningBytes {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeResourceLimit}, "request must be non-empty and within the Meaning input limit")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeInvalid}, "request has invalid or duplicate JSON fields")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var request meaningRealizeRequest
	if err := decoder.Decode(&request); err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeInvalid}, "request does not match the supported schema")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeInvalid}, "request must contain exactly one JSON value")
	}
	if loader == nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeUnsupported}, "this executable cannot load configured language runtimes")
	}
	configured, err := loader(options.runtimeConfig)
	if err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError}, "configured language runtime could not be loaded")
	}
	if configured == nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeUnknown}, "configured language runtime returned no engine")
	}
	graph := request.Meaning
	if err := meaning.ValidateMeaning(graph, configured.CoreGraph(), meaning.DefaultBudget()); err != nil {
		return writeMeaningRealizeError(stdout, stderr, err, "meaning graph failed validation")
	}
	if err := meaning.ValidateReferences(graph, request.Context, meaning.DefaultBudget()); err != nil {
		return writeMeaningRealizeError(stdout, stderr, err, "request context failed validation")
	}
	semanticID, err := meaning.CanonicalHash(graph, meaning.DefaultBudget())
	if err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError}, "meaning identity could not be computed")
	}
	candidate := semantic.MeaningCandidate{
		ID: "provided-meaning", SemanticID: semanticID, Meaning: &graph,
		Contexts: []semantic.MeaningContext{{Meaning: graph, Context: request.Context}},
		Status:   status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: "provided-meaning", Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{}},
	}
	target, ok := configured.TargetRealizerFor(features.LanguageID(options.language))
	if !ok || target == nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeUnsupported, Language: options.language}, "no verified target realizer is configured for the requested language")
	}
	requestCtx, cancel := context.WithTimeout(context.Background(), engine.DefaultBudget().RequestTimeout)
	defer cancel()
	surface, err := target.Realize(requestCtx, candidate)
	if err != nil {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError, Language: options.language}, "target realization failed")
	}
	if !surface.Outcome.Valid() {
		return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeError, Language: options.language}, "target returned an invalid outcome")
	}
	return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: surface.Outcome, Language: options.language, Alternatives: surface.Alternatives, Diagnostics: surface.Diagnostics}, "")
}

func writeMeaningRealizeError(stdout, stderr io.Writer, err error, fallback string) int {
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		result := typed.Result()
		if result.Validate() == nil {
			return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: result.Status, Diagnostics: result.Diagnostics}, "")
		}
	}
	return writeMeaningRealize(stdout, stderr, meaningRealizeOutput{Status: status.OutcomeInvalid}, fallback)
}

func writeMeaningRealize(stdout, stderr io.Writer, result meaningRealizeOutput, message string) int {
	if message != "" {
		code := "MEANING_REALIZATION_ERROR"
		if result.Status == status.OutcomeInvalid {
			code = "INVALID_MEANING_REALIZATION_REQUEST"
		}
		if result.Status == status.OutcomeResourceLimit {
			code = "MEANING_REALIZATION_SIZE"
		}
		if result.Status == status.OutcomeUnsupported {
			code = "MEANING_REALIZATION_UNSUPPORTED"
		}
		result.Diagnostics = []status.Diagnostic{{Code: code, Path: "meaning", Message: message, Severity: status.SeverityError}}
	}
	if result.Alternatives == nil {
		result.Alternatives = []engine.SurfaceAlternative{}
	}
	if result.Diagnostics == nil {
		result.Diagnostics = []status.Diagnostic{}
	}
	data, err := json.Marshal(result)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize meaning realization: %v\n", err)
		return ExitError
	}
	if len(data) > maxMeaningRealizeOutputBytes {
		result = meaningRealizeOutput{Status: status.OutcomeResourceLimit, Alternatives: []engine.SurfaceAlternative{}, Diagnostics: []status.Diagnostic{{Code: "MEANING_REALIZATION_OUTPUT_SIZE", Path: "alternatives", Message: "target response exceeds the output limit", Severity: status.SeverityError}}}
		data, err = json.Marshal(result)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize meaning realization: %v\n", err)
		return ExitError
	}
	if _, err := fmt.Fprintln(stdout, string(data)); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write meaning realization: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

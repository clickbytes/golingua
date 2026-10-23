package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
)

type formsOptions struct {
	language, lexeme, surface string
	features                  []features.FeatureValue
	json                      bool
}

type formsResponse = app.FormsResult

func parseForms(args []string) (formsOptions, error) {
	var options formsOptions
	if len(args) > 80 {
		return options, fmt.Errorf("forms argument count exceeds limit")
	}
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if flag == "--json" {
			if seen[flag] {
				return options, fmt.Errorf("duplicate --json flag")
			}
			seen[flag], options.json = true, true
			continue
		}
		if flag != "--language" && flag != "--lexeme" && flag != "--surface" && flag != "--feature" {
			return options, fmt.Errorf("unknown forms flag")
		}
		if flag != "--feature" && seen[flag] {
			return options, fmt.Errorf("duplicate %s flag", flag)
		}
		seen[flag] = true
		if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
			return options, fmt.Errorf("missing value for %s", flag)
		}
		index++
		value := args[index]
		if len(value) > 1024 || value == "" {
			return options, fmt.Errorf("invalid value for %s", flag)
		}
		switch flag {
		case "--language":
			options.language = value
		case "--lexeme":
			options.lexeme = value
		case "--surface":
			options.surface = value
		case "--feature":
			name, featureValue, found := strings.Cut(value, "=")
			if !found || name == "" || featureValue == "" || len(name) > 64 || len(featureValue) > 64 {
				return options, fmt.Errorf("feature must be name=value with bounded names")
			}
			options.features = append(options.features, features.FeatureValue{Feature: features.FeatureID(name), State: features.StateKnown, Value: features.FeatureValueID(featureValue)})
		}
	}
	if !options.json || options.language != "pl" && options.language != "en" {
		return options, fmt.Errorf("forms requires --json and --language pl|en")
	}
	if (options.lexeme == "") == (options.surface == "") {
		return options, fmt.Errorf("forms requires exactly one of --lexeme or --surface")
	}
	if options.surface != "" && len(options.features) > 0 {
		return options, fmt.Errorf("--feature is only valid with --lexeme")
	}
	if options.lexeme != "" {
		if len(options.features) == 0 {
			return options, fmt.Errorf("generation requires at least one --feature")
		}
		parsedID, err := ids.ParseLexemeID(options.lexeme)
		if err != nil || parsedID.String() != options.lexeme {
			return options, fmt.Errorf("invalid lexeme ID")
		}
	}
	return options, nil
}

func runForms(args []string, stdout, stderr io.Writer, service *app.Service) int {
	options, err := parseForms(args)
	if err != nil {
		return writeForms(stdout, stderr, formsResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{formsDiagnostic("INVALID_FORMS_ARGUMENTS", err.Error(), "forms")}})
	}
	response := service.Forms(context.Background(), app.FormsRequest{
		Language: options.language, LexemeID: options.lexeme, Surface: options.surface,
		Features: options.features,
	})
	return writeForms(stdout, stderr, response)
}

func formsDiagnostic(code, message, path string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityError}
}

func writeForms(stdout, stderr io.Writer, response formsResponse) int {
	if len(response.Forms) > morph.MaxGeneratedForms || len(response.Readings) > 4096 || len(response.Diagnostics) > status.MaxDiagnostics {
		response = formsResponse{Status: status.OutcomeResourceLimit, Diagnostics: []status.Diagnostic{formsDiagnostic("RESOURCE_LIMIT", "morphology output exceeds limit", "forms")}}
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(response); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize forms: %v\n", err)
		return 1
	}
	switch response.Status {
	case status.OutcomeOK:
		return 0
	case status.OutcomeInvalid:
		return 2
	case status.OutcomeUnsupported:
		return 3
	case status.OutcomeResourceLimit:
		return 4
	case status.OutcomeUnknown:
		return 5
	case status.OutcomeAmbiguous:
		return 6
	case status.OutcomeIncomplete:
		return 7
	default:
		return 1
	}
}

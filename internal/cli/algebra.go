package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

type algebraOptions struct {
	input string
	json  bool
}

type algebraOutput struct {
	Status            status.OutcomeKind        `json:"status"`
	Checked           algebra.CheckedExpression `json:"checked"`
	CanonicalDocument json.RawMessage           `json:"canonical_document"`
	Diagnostics       []status.Diagnostic       `json:"diagnostics,omitempty"`
}

func parseAlgebraArgs(args []string) (algebraOptions, string) {
	var options algebraOptions
	seenJSON, seenInput := false, false
	if len(args) < 2 || args[0] != "algebra" || args[1] != "check" {
		return options, "expected algebra check"
	}
	for i := 2; i < len(args); i++ {
		flag := args[i]
		switch flag {
		case "--json":
			if seenJSON {
				return options, "duplicate --json flag"
			}
			seenJSON = true
			options.json = true
		case "--input":
			if seenInput {
				return options, "duplicate --input flag"
			}
			seenInput = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --input"
			}
			i++
			options.input = args[i]
			if options.input == "" {
				return options, "--input value must not be empty"
			}
		default:
			return options, fmt.Sprintf("unknown algebra check argument %q", flag)
		}
	}
	if !seenJSON {
		return options, "algebra check requires --json"
	}
	return options, ""
}

func runAlgebra(args []string, input io.Reader, stdout, stderr io.Writer) int {
	options, parseError := parseAlgebraArgs(args)
	if parseError != "" {
		return writeResult(stdout, stderr, algebraDiagnostic(status.OutcomeInvalid, "INVALID_ALGEBRA_ARGUMENTS", parseError, "algebra"))
	}
	var reader io.Reader
	if options.input == "" || options.input == "-" {
		if input == nil {
			_, _ = fmt.Fprintln(stderr, "INFRASTRUCTURE_ERROR: stdin reader is unavailable")
			return ExitError
		}
		reader = input
	} else {
		file, err := os.Open(options.input)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not open algebra input: %v\n", err)
			return ExitError
		}
		defer file.Close()
		reader = file
	}
	data, err := io.ReadAll(io.LimitReader(reader, algebra.MaxDocumentBytes+1))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read algebra input: %v\n", err)
		return ExitError
	}
	document, err := algebra.Decode(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	checked, err := algebra.Check(document)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	canonical, err := algebra.Encode(document)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraOutput{
		Status: checked.Status, Checked: checked,
		CanonicalDocument: json.RawMessage(canonical),
	}
	if checked.Status == status.OutcomeIncomplete {
		output.Diagnostics = []status.Diagnostic{{
			Code: "EXPRESSION_INCOMPLETE", Message: "typed holes remain unresolved",
			Path: "expression", Severity: status.SeverityWarning,
		}}
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra result: %v\n", err)
		return ExitError
	}
	return exitCode(checked.Status)
}

func writeAlgebraError(stdout, stderr io.Writer, err error) int {
	var typed *algebra.Error
	if errors.As(err, &typed) {
		return writeResult(stdout, stderr, algebraDiagnostic(typed.Status, typed.Code, typed.Message, typed.Path))
	}
	_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: algebra check failed unexpectedly: %v\n", err)
	return ExitError
}

func algebraDiagnostic(kind status.OutcomeKind, code, message, path string) status.Result {
	message = strings.ToValidUTF8(message, "�")
	path = strings.ToValidUTF8(path, "�")
	if len(path) > status.MaxDiagnosticPathBytes {
		path = truncateAlgebraText(path, status.MaxDiagnosticPathBytes)
		const truncationNotice = " (diagnostic path was truncated)"
		message = truncateAlgebraText(message, status.MaxDiagnosticMessageBytes-len(truncationNotice)) + truncationNotice
	} else {
		message = truncateAlgebraText(message, status.MaxDiagnosticMessageBytes)
	}
	return diagnosticResult(kind, code, message, path)
}

func truncateAlgebraText(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	const marker = "..."
	limit := maxBytes - len(marker)
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit] + marker
}

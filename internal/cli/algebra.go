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
	"github.com/clickbytes/golingua/internal/metamodel"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

type algebraOptions struct {
	command  string
	input    string
	metadata []string
	json     bool
}

type algebraOutput struct {
	Status            status.OutcomeKind        `json:"status"`
	Checked           algebra.CheckedExpression `json:"checked"`
	CanonicalDocument json.RawMessage           `json:"canonical_document"`
	Metadata          []metadataReference       `json:"metadata,omitempty"`
	Diagnostics       []status.Diagnostic       `json:"diagnostics,omitempty"`
}

type metadataReference struct {
	ID                  string `json:"id"`
	Version             string `json:"version"`
	ResourceFingerprint string `json:"resource_fingerprint"`
}

type algebraBuildOutput struct {
	algebra.BuildResult
	ResultKind   string                `json:"result_kind"`
	Substitution *algebra.Substitution `json:"substitution,omitempty"`
	Metadata     []metadataReference   `json:"metadata,omitempty"`
	Diagnostics  []status.Diagnostic   `json:"diagnostics,omitempty"`
}

type metadataPackageOutput struct {
	ID                  string                 `json:"id"`
	Version             string                 `json:"version"`
	Profile             string                 `json:"profile"`
	ResourceFingerprint string                 `json:"resource_fingerprint"`
	FormalDenotationID  string                 `json:"formal_denotation_id"`
	SortCount           int                    `json:"sort_count"`
	SymbolCount         int                    `json:"symbol_count"`
	PrimitiveCount      int                    `json:"primitive_count"`
	AliasCount          int                    `json:"alias_count"`
	OpaqueCount         int                    `json:"opaque_unformalized_count"`
	PendingCount        int                    `json:"pending_count"`
	ApprovedCount       int                    `json:"approved_count"`
	Provenance          []provenance.SourceRef `json:"provenance,omitempty"`
	Symbols             []metadataSymbolOutput `json:"symbols,omitempty"`
}

type metadataSymbolOutput struct {
	ID           string                 `json:"id"`
	Kind         metamodel.SymbolKind   `json:"kind"`
	Label        string                 `json:"label,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ReviewStatus metamodel.ReviewStatus `json:"review_status"`
	Provenance   []provenance.SourceRef `json:"provenance"`
}

type metadataCommandOutput struct {
	Status          status.OutcomeKind      `json:"status"`
	Operation       string                  `json:"operation"`
	ValidationScope string                  `json:"validation_scope,omitempty"`
	Packages        []metadataPackageOutput `json:"packages"`
}

type algebraInfrastructureError struct{ message string }

func (failure *algebraInfrastructureError) Error() string { return failure.message }

func parseAlgebraArgs(args []string) (algebraOptions, string) {
	var options algebraOptions
	if len(args) < 2 || args[0] != "algebra" {
		return options, "expected algebra command"
	}
	index := 2
	if args[1] == "check" {
		options.command = "check"
	} else if args[1] == "build" {
		options.command = "build"
	} else if args[1] == "metadata" {
		if len(args) < 3 || (args[2] != "inspect" && args[2] != "validate") {
			return options, "expected algebra metadata inspect or validate"
		}
		options.command = "metadata_" + args[2]
		index = 3
	} else {
		return options, "expected algebra check, build, or metadata"
	}
	seenJSON, seenInput := false, false
	for i := index; i < len(args); i++ {
		flag := args[i]
		switch flag {
		case "--json":
			if seenJSON {
				return options, "duplicate --json flag"
			}
			seenJSON = true
			options.json = true
		case "--input":
			if options.command != "check" && options.command != "build" {
				return options, "--input is only valid for algebra check or build"
			}
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
		case "--metadata":
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --metadata"
			}
			i++
			if args[i] == "" {
				return options, "--metadata value must not be empty"
			}
			options.metadata = append(options.metadata, args[i])
		default:
			return options, fmt.Sprintf("unknown algebra check argument %q", flag)
		}
	}
	if !seenJSON {
		return options, "algebra command requires --json"
	}
	if strings.HasPrefix(options.command, "metadata_") && len(options.metadata) == 0 {
		return options, "algebra metadata requires at least one --metadata file"
	}
	if len(options.metadata) > 16 {
		return options, "algebra metadata accepts at most 16 files"
	}
	if options.command == "build" {
		if len(options.metadata) == 0 {
			return options, "algebra build requires at least one --metadata file"
		}
	}
	return options, ""
}

func runAlgebra(args []string, input io.Reader, stdout, stderr io.Writer) int {
	options, parseError := parseAlgebraArgs(args)
	if parseError != "" {
		return writeResult(stdout, stderr, algebraDiagnostic(status.OutcomeInvalid, "INVALID_ALGEBRA_ARGUMENTS", parseError, "algebra"))
	}
	if options.command == "metadata_inspect" || options.command == "metadata_validate" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		operation := strings.TrimPrefix(options.command, "metadata_")
		output := metadataCommandOutput{Status: status.OutcomeOK, Operation: operation, Packages: make([]metadataPackageOutput, 0, len(registries))}
		if operation == "validate" {
			output.ValidationScope = "positive-facts-v1 profile, manifest structure, dependency closure, declarations and types, constraints, provenance, and alias definitions; no theorem consistency, world-truth, or owner-acceptance claim"
		}
		for _, registry := range registries {
			output.Packages = append(output.Packages, metadataPackageSummary(registry, operation == "inspect"))
		}
		if err := writeJSON(stdout, output); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize metadata result: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	var registries []*metamodel.Registry
	if len(options.metadata) != 0 {
		var err error
		registries, err = readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	if options.command == "build" {
		return runAlgebraBuild(options, input, stdout, stderr, registries)
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
	var document algebra.Document
	var checked algebra.CheckedExpression
	var canonical []byte
	if len(options.metadata) == 0 {
		document, err = algebra.Decode(data)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		checked, err = algebra.Check(document)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		canonical, err = algebra.Encode(document)
	} else {
		document, err = algebra.DecodeDocumentJSON(data)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		checked, document, err = metamodel.CheckDocument(document, registries)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		canonical, err = algebra.Encode(document)
	}
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraOutput{
		Status: checked.Status, Checked: checked,
		CanonicalDocument: json.RawMessage(canonical),
	}
	for _, registry := range registries {
		output.Metadata = append(output.Metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
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

func runAlgebraBuild(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	if len(registries) == 0 {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "MISSING_METADATA", Path: "metadata", Message: "algebra build requires at least one metadata package"})
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
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not open algebra build input: %v\n", err)
			return ExitError
		}
		defer file.Close()
		reader = file
	}
	data, err := io.ReadAll(io.LimitReader(reader, algebra.MaxDocumentBytes+1))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read algebra build input: %v\n", err)
		return ExitError
	}
	request, err := algebra.DecodeBuildRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	context, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.Build(context, request)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if result.CanonicalDocument != nil && result.Expression != nil {
		admission := *result.CanonicalDocument
		admission.Sorts = nil
		admission.Signatures = nil
		checked, hydrated, err := metamodel.CheckDocument(admission, registries)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		result.Status = checked.Status
		result.Checked = &checked
		for i := range hydrated.Signatures {
			if hydrated.Signatures[i].Ports == nil {
				hydrated.Signatures[i].Ports = []algebra.Port{}
			}
		}
		result.CanonicalDocument = &hydrated
	}
	resultKind := ""
	for _, record := range result.Trace {
		if record.ID == result.ResultID {
			resultKind = record.ResultKind
			break
		}
	}
	output := algebraBuildOutput{BuildResult: result, ResultKind: resultKind}
	if resultKind == "substitution" {
		substitution := result.Substitution
		if substitution == nil {
			substitution = algebra.Substitution{}
		}
		output.Substitution = &substitution
	}
	for _, registry := range registries {
		output.Metadata = append(output.Metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
	}
	if result.Status == status.OutcomeIncomplete && result.Checked != nil && len(result.Checked.Holes) > 0 {
		output.Diagnostics = []status.Diagnostic{{Code: "EXPRESSION_INCOMPLETE", Message: "typed holes remain unresolved", Path: "result.expression", Severity: status.SeverityWarning}}
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra build result: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

func readMetadataRegistries(paths []string) ([]*metamodel.Registry, error) {
	if len(paths) == 0 {
		return nil, &algebra.Error{Status: status.OutcomeInvalid, Code: "MISSING_METADATA", Path: "metadata", Message: "at least one metadata file is required"}
	}
	if len(paths) > 16 {
		return nil, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: "metadata", Message: "metadata file count exceeds limit 16"}
	}
	inputs := make([][]byte, 0, len(paths))
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not open metadata file %q: %v", path, err)}
		}
		data, readErr := io.ReadAll(io.LimitReader(file, algebra.MaxDocumentBytes+1))
		closeErr := file.Close()
		if readErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not read metadata file %q: %v", path, readErr)}
		}
		if closeErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not close metadata file %q: %v", path, closeErr)}
		}
		inputs = append(inputs, data)
	}
	registries, err := metamodel.LoadPackages(inputs)
	if err != nil {
		return nil, err
	}
	return registries, nil
}

func metadataPackageSummary(registry *metamodel.Registry, includeSymbols bool) metadataPackageOutput {
	document := registry.Document()
	result := metadataPackageOutput{
		ID: registry.ID(), Version: registry.Version(), Profile: document.Profile,
		ResourceFingerprint: registry.ResourceFingerprint(), FormalDenotationID: registry.FormalDenotationID(),
		SortCount: len(document.Sorts), SymbolCount: len(document.Symbols),
		Provenance: append([]provenance.SourceRef(nil), document.Provenance...),
	}
	for _, symbol := range document.Symbols {
		switch symbol.Kind {
		case metamodel.SymbolPrimitive:
			result.PrimitiveCount++
		case metamodel.SymbolAlias:
			result.AliasCount++
		case metamodel.SymbolOpaqueUnformalized:
			result.OpaqueCount++
		}
		if symbol.ReviewStatus == metamodel.ReviewPending {
			result.PendingCount++
		} else if symbol.ReviewStatus == metamodel.ReviewApproved {
			result.ApprovedCount++
		}
		if includeSymbols {
			result.Symbols = append(result.Symbols, metadataSymbolOutput{ID: symbol.ID, Kind: symbol.Kind, Label: symbol.Label, Description: symbol.Description, ReviewStatus: symbol.ReviewStatus, Provenance: append([]provenance.SourceRef(nil), symbol.Provenance...)})
		}
	}
	return result
}

func writeAlgebraError(stdout, stderr io.Writer, err error) int {
	var infrastructure *algebraInfrastructureError
	if errors.As(err, &infrastructure) {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: %s\n", infrastructure.message)
		return ExitError
	}
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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
)

const maxManifestInput = importer.MaxManifestBytes
const maxCLIPathBytes = 4096

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "unimorph" {
		return runUniMorph(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "constructions" {
		return runConstructions(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "snapshot" {
		return runSnapshot(args[1:], stdout, stderr)
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Fprintln(stdout, "golingua-import --manifest PATH --data-dir DIR\n        golingua-import snapshot preview --manifest PATH --data-dir DIR --store DIR\n        golingua-import snapshot apply --manifest PATH --data-dir DIR --store DIR\n        golingua-import unimorph --input PATH --mapping PATH --features PATH --source-id ID --source-version VERSION --license LABEL (--source-path PATH | --source-uri URI)\n        golingua-import constructions --input PATH --features PATH\nValidates offline data and emits non-activated proposals; snapshot activation is local and explicit.")
		return 0
	}
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintln(stdout, "golingua-import 0.1.0")
		return 0
	}
	if len(args) != 4 || args[0] != "--manifest" || args[2] != "--data-dir" {
		fmt.Fprintln(stderr, "usage: golingua-import --manifest PATH --data-dir DIR")
		return 2
	}
	manifestPath, err := filepath.Abs(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "INVALID manifest path: %v\n", err)
		return 2
	}
	manifestBytes, err := readFileBounded(manifestPath, maxManifestInput)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read manifest: %v\n", err)
		return 1
	}
	manifest, err := importer.DecodeManifest(manifestBytes)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	data, err := importer.LoadDataset(args[3], manifest)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	conceptCount, lexemeCount, senseCount, bindingCount := 0, 0, 0, 0
	for _, file := range data.Concepts {
		conceptCount += len(file.Concepts)
	}
	for _, file := range data.Lexicons {
		lexemeCount += len(file.Lexemes)
		senseCount += len(file.Senses)
		bindingCount += len(file.Bindings)
	}
	_ = json.NewEncoder(stdout).Encode(struct {
		Status   string `json:"status"`
		Dataset  string `json:"dataset_id"`
		Concepts int    `json:"concepts"`
		Lexemes  int    `json:"lexemes"`
		Senses   int    `json:"senses"`
		Bindings int    `json:"bindings"`
	}{"valid", manifest.DatasetID, conceptCount, lexemeCount, senseCount, bindingCount})
	return 0
}

func runSnapshot(args []string, stdout, stderr io.Writer) int {
	if (len(args) == 1 && (args[0] == "--help" || args[0] == "-h")) ||
		(len(args) == 2 && (args[1] == "--help" || args[1] == "-h") && (args[0] == "preview" || args[0] == "apply")) {
		fmt.Fprintln(stdout, "Usage: golingua-import snapshot preview|apply --manifest PATH --data-dir DIR --store DIR")
		return 0
	}
	if len(args) == 0 || (args[0] != "preview" && args[0] != "apply") {
		fmt.Fprintln(stderr, "usage: golingua-import snapshot preview|apply --manifest PATH --data-dir DIR --store DIR")
		return 2
	}
	action := args[0]
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		fmt.Fprintln(stdout, "Usage: golingua-import snapshot "+action+" --manifest PATH --data-dir DIR --store DIR")
		return 0
	}
	flags := flag.NewFlagSet("snapshot "+action, flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "bounded dataset manifest JSON")
	dataRoot := flags.String("data-dir", "", "root directory containing manifest files")
	storeRoot := flags.String("store", "", "local active snapshot store directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || !validCLIPath(*manifestPath) || !validCLIPath(*dataRoot) || !validCLIPath(*storeRoot) {
		fmt.Fprintln(stderr, "usage: golingua-import snapshot preview|apply --manifest PATH --data-dir DIR --store DIR")
		return 2
	}
	manifestBytes, err := readFileBounded(*manifestPath, maxManifestInput)
	if err != nil {
		printFailure(stdout, stderr, inputReadError("manifest", maxManifestInput, err))
		return 1
	}
	manifest, err := importer.DecodeManifest(manifestBytes)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	candidate, err := staging.BuildCandidate(*dataRoot, manifest)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	store, err := staging.OpenStore(*storeRoot)
	if err != nil {
		printFailure(stdout, stderr, storeInputError(*storeRoot, err))
		return 1
	}
	var result staging.Preview
	if action == "preview" {
		result, err = store.Preview(candidate)
	} else {
		result, err = store.Apply(candidate)
	}
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	activation := "not_applied"
	resultStatus := "previewed"
	if result.Applied {
		activation = "local_only"
		resultStatus = "applied"
		if !result.DurabilityConfirmed {
			resultStatus = "applied_durability_unconfirmed"
		}
	}
	report := struct {
		Status              string `json:"status"`
		Operation           string `json:"operation"`
		DatasetID           string `json:"dataset_id"`
		Version             string `json:"version"`
		Digest              string `json:"digest"`
		ActiveDigest        string `json:"active_digest,omitempty"`
		Changed             bool   `json:"changed"`
		Applied             bool   `json:"applied"`
		Activation          string `json:"activation"`
		DurabilityConfirmed bool   `json:"durability_confirmed"`
		Warning             string `json:"warning,omitempty"`
		Summary             struct {
			Files    int `json:"files"`
			Concepts int `json:"concepts"`
			Lexemes  int `json:"lexemes"`
			Senses   int `json:"senses"`
			Bindings int `json:"bindings"`
		} `json:"summary"`
	}{
		Status: resultStatus, Operation: action, DatasetID: result.DatasetID, Version: result.Version,
		Digest: result.Digest, ActiveDigest: result.ActiveDigest, Changed: result.Changed,
		Applied: result.Applied, Activation: activation,
		DurabilityConfirmed: result.DurabilityConfirmed, Warning: result.Warning,
		Summary: struct {
			Files    int `json:"files"`
			Concepts int `json:"concepts"`
			Lexemes  int `json:"lexemes"`
			Senses   int `json:"senses"`
			Bindings int `json:"bindings"`
		}{result.Summary.Files, result.Summary.Concepts, result.Summary.Lexemes, result.Summary.Senses, result.Summary.Bindings},
	}
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		fmt.Fprintf(stderr, "cannot write snapshot report: %v\n", err)
		return 1
	}
	return 0
}

func validCLIPath(value string) bool {
	return value != "" && len(value) <= maxCLIPathBytes && !strings.ContainsRune(value, '\x00')
}

func runConstructions(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(stdout, "Usage: golingua-import constructions --input PATH --features PATH")
		return 0
	}
	flags := flag.NewFlagSet("constructions", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "versioned construction and valency JSON dataset")
	featuresPath := flags.String("features", "", "matching language feature profile JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *inputPath == "" || *featuresPath == "" {
		fmt.Fprintln(stderr, "usage: golingua-import constructions --input PATH --features PATH")
		return 2
	}
	data, err := readFileBounded(*inputPath, importer.MaxConstructionDatasetBytes)
	if err != nil {
		printFailure(stdout, stderr, constructionInputError("input", importer.MaxConstructionDatasetBytes, err))
		return 1
	}
	profileData, err := readFileBounded(*featuresPath, features.MaxProfileBytes)
	if err != nil {
		printFailure(stdout, stderr, constructionInputError("features", features.MaxProfileBytes, err))
		return 1
	}
	profile, err := features.DecodeProfile(profileData)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	compiled, err := importer.DecodeConstructionDataset(data, profile)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	plans := compiled.Plans.Plans()
	planIDs := make([]string, 0, 16)
	for index, plan := range plans {
		if index == 16 {
			break
		}
		planIDs = append(planIDs, string(plan.ID))
	}
	result := struct {
		Status            string   `json:"status"`
		Language          string   `json:"language"`
		Families          int      `json:"families"`
		ValencyFrames     int      `json:"valency_frames"`
		Operators         int      `json:"operators"`
		PlanCount         int      `json:"compiled_plan_count"`
		SamplePlanIDs     []string `json:"sample_plan_ids"`
		RuntimeActivation string   `json:"runtime_activation"`
	}{"compiled_not_activated", string(compiled.Dataset.Language), len(compiled.Dataset.Packages), len(compiled.Dataset.Frames), len(compiled.Dataset.Operators), len(plans), planIDs, "not_activated"}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		fmt.Fprintf(stderr, "cannot write compile report: %v\n", err)
		return 1
	}
	return 0
}

func runUniMorph(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(stdout, "Usage: golingua-import unimorph --input PATH --mapping PATH --features PATH --source-id ID --source-version VERSION --license LABEL (--source-path PATH | --source-uri URI)")
		return 0
	}
	flags := flag.NewFlagSet("unimorph", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "local three-column UniMorph TSV")
	mappingPath := flags.String("mapping", "", "reviewed language UniMorph mapping JSON")
	featuresPath := flags.String("features", "", "language feature profile JSON")
	sourceID := flags.String("source-id", "", "stable source identifier")
	sourceVersion := flags.String("source-version", "", "source version")
	sourceLicense := flags.String("license", "", "source license label")
	sourcePath := flags.String("source-path", "", "relative source path")
	sourceURI := flags.String("source-uri", "", "absolute source URI")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *inputPath == "" || *mappingPath == "" || *featuresPath == "" || *sourceID == "" || *sourceVersion == "" || *sourceLicense == "" || (*sourcePath == "") == (*sourceURI == "") {
		fmt.Fprintln(stderr, "usage: golingua-import unimorph --input PATH --mapping PATH --features PATH --source-id ID --source-version VERSION --license LABEL (--source-path PATH | --source-uri URI)")
		return 2
	}
	input, err := os.Open(*inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "cannot open TSV: %v\n", err)
		return 1
	}
	defer input.Close()
	mappingBytes, err := readFileBounded(*mappingPath, importer.MaxUniMorphProfileBytes)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read mapping: %v\n", err)
		return 1
	}
	featureBytes, err := readFileBounded(*featuresPath, features.MaxProfileBytes)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read feature profile: %v\n", err)
		return 1
	}
	featureProfile, err := features.DecodeProfile(featureBytes)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	mapping, err := importer.DecodeUniMorphProfile(mappingBytes, featureProfile)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	source := provenance.SourceRef{ID: *sourceID, Path: *sourcePath, URI: *sourceURI, Version: *sourceVersion, License: *sourceLicense}
	result, err := importer.ReadUniMorph(input, mapping, featureProfile, source)
	if err != nil {
		printFailure(stdout, stderr, err)
		return 1
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(stderr, "cannot encode proposals: %v\n", err)
		return 1
	}
	if len(encoded) > importer.MaxUniMorphOutputBytes {
		fmt.Fprintln(stderr, "proposal output exceeds configured limit")
		return 1
	}
	if _, err := fmt.Fprintln(stdout, string(encoded)); err != nil {
		fmt.Fprintf(stderr, "cannot write proposals: %v\n", err)
		return 1
	}
	return 0
}

func readFileBounded(path string, maximum int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, errInputFileTooLarge
	}
	return data, nil
}

var errInputFileTooLarge = errors.New("input file exceeds configured limit")

type cliFailure struct{ result status.Result }

func (failure *cliFailure) Error() string         { return failure.result.Diagnostics[0].Message }
func (failure *cliFailure) Result() status.Result { return failure.result }

func constructionInputError(label string, maximum int64, cause error) error {
	kind := status.OutcomeError
	code, message := "INPUT_READ_FAILED", "cannot read construction input file"
	if errors.Is(cause, errInputFileTooLarge) {
		kind, code = status.OutcomeResourceLimit, "RESOURCE_LIMIT"
		message = fmt.Sprintf("construction input exceeds configured byte limit of %d", maximum)
	}
	diagnostic := status.Diagnostic{Code: code, Path: label, Message: message, Severity: status.SeverityError}
	result, err := status.NewResult(kind, "", []status.Diagnostic{diagnostic}, nil)
	if err != nil {
		return &cliFailure{result: status.Result{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{{Code: "INPUT_READ_FAILED", Path: label, Message: "cannot read construction input file", Severity: status.SeverityError}}}}
	}
	return &cliFailure{result: result}
}

func inputReadError(label string, maximum int64, cause error) error {
	kind := status.OutcomeError
	code, message := "INPUT_READ_FAILED", "cannot read input file"
	if errors.Is(cause, errInputFileTooLarge) {
		kind, code = status.OutcomeResourceLimit, "RESOURCE_LIMIT"
		message = fmt.Sprintf("input exceeds configured byte limit of %d", maximum)
	}
	return newCLIFailure(kind, code, label, message)
}

func storeInputError(_ string, cause error) error {
	var typed interface{ Result() status.Result }
	if errors.As(cause, &typed) && typed.Result().Validate() == nil {
		return cause
	}
	return newCLIFailure(status.OutcomeError, "SNAPSHOT_STORE_OPEN_FAILED", "store", "cannot open local snapshot store")
}

func newCLIFailure(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	result, err := status.NewResult(kind, "", []status.Diagnostic{diagnostic}, nil)
	if err != nil {
		return &cliFailure{result: status.Result{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{{Code: "CLI_ERROR", Path: path, Message: "command failed", Severity: status.SeverityError}}}}
	}
	return &cliFailure{result: result}
}

func printFailure(stdout, stderr io.Writer, err error) {
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		if typed.Result().Validate() == nil {
			if encodeErr := json.NewEncoder(stdout).Encode(typed.Result()); encodeErr == nil {
				return
			}
		}
	}
	fmt.Fprintf(stderr, "ERROR: %v\n", err)
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

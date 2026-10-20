package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const maxManifestInput = importer.MaxManifestBytes

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "unimorph" {
		return runUniMorph(args[1:], stdout, stderr)
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Fprintln(stdout, "golingua-import --manifest PATH --data-dir DIR\n        golingua-import unimorph --input PATH --mapping PATH --features PATH --source-id ID --source-version VERSION --license LABEL (--source-path PATH | --source-uri URI)\nValidates offline data and emits non-activated morphology import proposals.")
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
		return nil, fmt.Errorf("manifest exceeds %d bytes", maximum)
	}
	return data, nil
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

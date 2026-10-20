package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/status"
)

const maxManifestInput = importer.MaxManifestBytes

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Fprintln(stdout, "golingua-import --manifest PATH --data-dir DIR\nValidates offline concept and lexicon definition files.")
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

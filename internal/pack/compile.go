package pack

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/core"
	conceptgraph "github.com/clickbytes/golingua/internal/graph"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
)

// Compile verifies the exact candidate bytes and builds immutable CORE and
// lexical indexes. Candidate data is bounded by the importer contracts.
func Compile(candidate staging.Candidate) (*Pack, error) {
	if err := candidate.Manifest.Validate(); err != nil {
		return nil, preserveOutcome(err, status.OutcomeInvalid, "INVALID_PACK_SOURCE", "manifest")
	}
	digest, err := staging.ComputeDigest(candidate.Manifest, candidate.Files)
	if err != nil {
		return nil, preserveOutcome(err, status.OutcomeInvalid, "INVALID_PACK_SOURCE", "candidate")
	}
	if candidate.Digest == "" || candidate.Digest != digest {
		return nil, packError(status.OutcomeInvalid, "PACK_DIGEST_MISMATCH", "candidate.digest", "candidate bytes do not match their declared snapshot digest")
	}

	var conceptBytes, lexiconBytes []byte
	for index, file := range candidate.Files {
		declared := candidate.Manifest.Files[index]
		if file.ID != declared.ID || file.Path != declared.Path || file.SourceID != declared.SourceID {
			return nil, packError(status.OutcomeInvalid, "INVALID_PACK_SOURCE", fmt.Sprintf("files[%d]", index), "candidate file metadata differs from its manifest")
		}
		switch {
		case file.ID == "concepts" || strings.HasSuffix(filepath.ToSlash(file.Path), ".concepts.json"):
			if conceptBytes != nil {
				return nil, packError(status.OutcomeInvalid, "DUPLICATE_CONCEPT_FILE", file.ID, "pack source must contain one concepts file")
			}
			conceptBytes = file.Contents
		case file.ID == "lexicon" || strings.HasSuffix(filepath.ToSlash(file.Path), ".lexicon.json"):
			if lexiconBytes != nil {
				return nil, packError(status.OutcomeInvalid, "DUPLICATE_LEXICON_FILE", file.ID, "pack source may contain at most one lexicon file")
			}
			lexiconBytes = file.Contents
		default:
			return nil, packError(status.OutcomeInvalid, "UNKNOWN_DATA_FILE_KIND", file.ID, "pack source contains an unsupported file kind")
		}
	}
	if conceptBytes == nil {
		return nil, packError(status.OutcomeInvalid, "MISSING_CONCEPT_FILE", "files", "pack source must contain a concepts file")
	}
	conceptFile, err := importer.DecodeConceptFile(conceptBytes)
	if err != nil {
		return nil, preserveOutcome(err, status.OutcomeInvalid, "INVALID_CONCEPT_FILE", "concepts")
	}
	graph, err := core.CompileConceptGraph(conceptFile.Concepts, conceptFile.Relations)
	if err != nil {
		return nil, packError(status.OutcomeInvalid, "INVALID_CONCEPT_FILE", "concepts", err.Error())
	}
	adjacency, err := conceptgraph.Compile(graph)
	if err != nil {
		return nil, preserveOutcome(err, status.OutcomeInvalid, "INVALID_CONCEPT_GRAPH", "concepts")
	}

	lexicalDocument := lexicon.Document{SchemaVersion: lexicon.SchemaVersion}
	if lexiconBytes != nil {
		lexiconFile, decodeErr := importer.DecodeLexiconFile(lexiconBytes, graph)
		if decodeErr != nil {
			return nil, preserveOutcome(decodeErr, status.OutcomeInvalid, "INVALID_LEXICON_FILE", "lexicon")
		}
		lexicalDocument = lexiconFile.Document()
	}
	lexicalIndex, err := lexicon.Build(lexicalDocument, graph, lexicon.DefaultBudget())
	if err != nil {
		return nil, preserveOutcome(err, status.OutcomeInvalid, "INVALID_LEXICON_FILE", "lexicon")
	}

	symbolBuilder := ids.NewSymbolTableBuilder()
	for _, concept := range graph.Concepts() {
		if err := symbolBuilder.Add(concept.ID.Symbol()); err != nil {
			return nil, packError(status.OutcomeInvalid, "INVALID_PACK_SYMBOL", "concepts", err.Error())
		}
	}
	for _, lexeme := range lexicalDocument.Lexemes {
		if err := symbolBuilder.Add(lexeme.ID.Symbol()); err != nil {
			return nil, packError(status.OutcomeInvalid, "INVALID_PACK_SYMBOL", "lexicon.lexemes", err.Error())
		}
	}
	for _, sense := range lexicalDocument.Senses {
		if err := symbolBuilder.Add(sense.ID.Symbol()); err != nil {
			return nil, packError(status.OutcomeInvalid, "INVALID_PACK_SYMBOL", "lexicon.senses", err.Error())
		}
	}
	symbols, err := symbolBuilder.Freeze()
	if err != nil {
		return nil, packError(status.OutcomeResourceLimit, "PACK_SYMBOL_LIMIT", "symbols", err.Error())
	}

	return &Pack{
		identity: Identity{FormatVersion: FormatVersion, DatasetID: candidate.Manifest.DatasetID,
			DatasetVersion: candidate.Manifest.Version, Digest: digest},
		graph: graph, adjacency: adjacency, lexicon: lexicalIndex, symbols: symbols,
	}, nil
}

func preserveOutcome(err error, fallback status.OutcomeKind, code, path string) error {
	var source interface{ Result() status.Result }
	if errors.As(err, &source) {
		result := source.Result()
		if result.Validate() == nil && result.Status != status.OutcomeOK {
			return &PackError{result: result}
		}
	}
	return packError(fallback, code, path, err.Error())
}

func packError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{
		Code:     boundedText(code, status.MaxDiagnosticCodeBytes),
		Path:     boundedText(path, status.MaxDiagnosticPathBytes),
		Message:  boundedText(message, status.MaxDiagnosticMessageBytes),
		Severity: status.SeverityError,
	}
	result, err := status.NewResult(kind, "", []status.Diagnostic{diagnostic}, nil)
	if err != nil {
		return &PackError{result: status.Result{Status: status.OutcomeError}}
	}
	return &PackError{result: result}
}

func boundedText(value string, maximum int) string {
	value = strings.ToValidUTF8(value, "\uFFFD")
	if len(value) <= maximum {
		return value
	}
	value = value[:maximum]
	for len(value) > 0 && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

// PackError exposes a bounded typed result for CLI serialization.
type PackError struct{ result status.Result }

func (err *PackError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "pack compilation failed"
	}
	return err.result.Diagnostics[0].Message
}

func (err *PackError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

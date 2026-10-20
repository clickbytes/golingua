package provenance

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxIndexedRecords = 100_000
	MaxRecordIDBytes  = 256
	MaxIndexedBytes   = 64 << 20
)

type RecordKind string

const (
	RecordLexeme       RecordKind = "lexeme"
	RecordParadigm     RecordKind = "paradigm"
	RecordConstruction RecordKind = "construction"
	RecordProcedure    RecordKind = "procedure"
	RecordException    RecordKind = "exception"
)

func (kind RecordKind) valid() bool {
	switch kind {
	case RecordLexeme, RecordParadigm, RecordConstruction, RecordProcedure, RecordException:
		return true
	default:
		return false
	}
}

// RecordRef links one bounded domain record to the source metadata that
// describes its origin. It does not grant redistribution or transfer rights.
type RecordRef struct {
	Kind   RecordKind `json:"kind"`
	ID     string     `json:"id"`
	Source SourceRef  `json:"source"`
}

// Index is an immutable source-to-record index. Accessors return copies.
type Index struct {
	bySource map[string][]RecordRef
	byRecord map[string][]RecordRef
}

// BuildIndex validates and freezes source references in deterministic order.
func BuildIndex(records []RecordRef) (*Index, error) {
	if len(records) > MaxIndexedRecords {
		return nil, provenanceError(status.OutcomeResourceLimit, "PROVENANCE_RECORD_LIMIT", "records", fmt.Sprintf("provenance records exceed limit %d", MaxIndexedRecords))
	}
	index := &Index{
		bySource: make(map[string][]RecordRef),
		byRecord: make(map[string][]RecordRef),
	}
	seen := make(map[string]struct{}, len(records))
	var totalBytes int64
	for position, record := range records {
		if !record.Kind.valid() || !validRecordID(record.Kind, record.ID) {
			return nil, provenanceError(status.OutcomeInvalid, "INVALID_PROVENANCE_RECORD", fmt.Sprintf("records[%d]", position), "record kind and bounded stable ID are required")
		}
		if err := record.Source.Validate(); err != nil {
			return nil, provenanceError(status.OutcomeInvalid, "INVALID_SOURCE_REF", fmt.Sprintf("records[%d].source", position), err.Error())
		}
		totalBytes += int64(len(record.Kind) + len(record.ID) + len(record.Source.ID) + len(record.Source.Path) + len(record.Source.URI) + len(record.Source.Version) + len(record.Source.License) + len(record.Source.Author) + len(record.Source.ChecksumSHA256))
		if totalBytes > MaxIndexedBytes {
			return nil, provenanceError(status.OutcomeResourceLimit, "PROVENANCE_BYTE_LIMIT", "records", fmt.Sprintf("provenance record bytes exceed limit %d", MaxIndexedBytes))
		}
		key := strings.Join([]string{record.Source.ID, string(record.Kind), record.ID, record.Source.Path, record.Source.URI, record.Source.Version, record.Source.License, record.Source.Author, record.Source.ChecksumSHA256}, "\x00")
		if _, duplicate := seen[key]; duplicate {
			return nil, provenanceError(status.OutcomeInvalid, "DUPLICATE_PROVENANCE_RECORD", fmt.Sprintf("records[%d]", position), "identical source-to-record reference is duplicated")
		}
		seen[key] = struct{}{}
		index.bySource[record.Source.ID] = append(index.bySource[record.Source.ID], record)
		index.byRecord[recordKey(record.Kind, record.ID)] = append(index.byRecord[recordKey(record.Kind, record.ID)], record)
	}
	for sourceID := range index.bySource {
		sort.Slice(index.bySource[sourceID], func(i, j int) bool { return recordLess(index.bySource[sourceID][i], index.bySource[sourceID][j]) })
	}
	for key := range index.byRecord {
		sort.Slice(index.byRecord[key], func(i, j int) bool { return recordLess(index.byRecord[key][i], index.byRecord[key][j]) })
	}
	return index, nil
}

func (index *Index) BySource(sourceID string) []RecordRef {
	if index == nil {
		return nil
	}
	return append([]RecordRef(nil), index.bySource[sourceID]...)
}

func (index *Index) ByRecord(kind RecordKind, id string) []RecordRef {
	if index == nil {
		return nil
	}
	return append([]RecordRef(nil), index.byRecord[recordKey(kind, id)]...)
}

func (index *Index) Len() int {
	if index == nil {
		return 0
	}
	total := 0
	for _, records := range index.bySource {
		total += len(records)
	}
	return total
}

func recordKey(kind RecordKind, id string) string { return string(kind) + "\x00" + id }

func validRecordID(kind RecordKind, id string) bool {
	if id == "" || len(id) > MaxRecordIDBytes || !utf8.ValidString(id) || strings.TrimSpace(id) != id || strings.ContainsRune(id, '\x00') {
		return false
	}
	switch kind {
	case RecordLexeme:
		parsed, err := ids.ParseLexemeID(id)
		return err == nil && parsed.String() == id
	case RecordConstruction:
		parsed, err := ids.ParseConstructionID(id)
		return err == nil && parsed.String() == id
	case RecordProcedure:
		parsed, err := ids.ParseProcedureID(id)
		return err == nil && parsed.String() == id
	case RecordParadigm, RecordException:
		if id[0] < 'a' || id[0] > 'z' {
			return false
		}
		for _, character := range id {
			if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func recordLess(left, right RecordRef) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.ID != right.ID {
		return left.ID < right.ID
	}
	if left.Source.ID != right.Source.ID {
		return left.Source.ID < right.Source.ID
	}
	if left.Source.Version != right.Source.Version {
		return left.Source.Version < right.Source.Version
	}
	if left.Source.ChecksumSHA256 != right.Source.ChecksumSHA256 {
		return left.Source.ChecksumSHA256 < right.Source.ChecksumSHA256
	}
	if left.Source.License != right.Source.License {
		return left.Source.License < right.Source.License
	}
	if left.Source.Path != right.Source.Path {
		return left.Source.Path < right.Source.Path
	}
	if left.Source.URI != right.Source.URI {
		return left.Source.URI < right.Source.URI
	}
	return left.Source.Author < right.Source.Author
}

type IndexError struct{ result status.Result }

func (err *IndexError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "provenance index build failed"
	}
	return err.result.Diagnostics[0].Message
}

func (err *IndexError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

func provenanceError(kind status.OutcomeKind, code, path, message string) error {
	limit := func(value string, maxBytes int) string {
		value = strings.ToValidUTF8(value, "�")
		if len(value) <= maxBytes {
			return value
		}
		value = value[:maxBytes]
		for len(value) > 0 && !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
		return value
	}
	diagnostic := status.Diagnostic{
		Code: limit(code, status.MaxDiagnosticCodeBytes), Path: limit(path, status.MaxDiagnosticPathBytes),
		Message: limit(message, status.MaxDiagnosticMessageBytes), Severity: status.SeverityError,
	}
	result, err := status.NewResult(kind, "", []status.Diagnostic{diagnostic}, nil)
	if err != nil {
		return &IndexError{result: status.Result{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{{Code: "PROVENANCE_INDEX_ERROR", Message: "provenance index build failed", Severity: status.SeverityError}}}}
	}
	return &IndexError{result: result}
}

package morph

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"golang.org/x/text/unicode/norm"
)

const (
	FiniteTableSchemaVersion = 1
	MaxFiniteTableBytes      = 32 << 20
	MaxFiniteTableReadings   = 100_000
	MaxSurfaceBytes          = 4_096
	MaxSurfaceCodepoints     = 2_048
)

type NormalizationForm string

const NormalizationNFC NormalizationForm = "nfc"

type NormalizationBudget struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxCodepoints  int
}

func DefaultNormalizationBudget() NormalizationBudget {
	return NormalizationBudget{MaxInputBytes: MaxSurfaceBytes, MaxOutputBytes: MaxSurfaceBytes, MaxCodepoints: MaxSurfaceCodepoints}
}

type FiniteTableBudget struct {
	MaxReadings int
	MaxBytes    int
	MaxSurface  int
}

func DefaultFiniteTableBudget() FiniteTableBudget {
	return FiniteTableBudget{MaxReadings: MaxFiniteTableReadings, MaxBytes: MaxFiniteTableBytes, MaxSurface: MaxSurfaceBytes}
}

type Reading struct {
	ID          string                 `json:"id"`
	Surface     string                 `json:"surface"`
	LexemeID    string                 `json:"lexeme_id"`
	POS         string                 `json:"pos"`
	ParadigmID  ParadigmID             `json:"paradigm_id"`
	CellID      CellID                 `json:"cell_id"`
	Features    features.FeatureBundle `json:"features"`
	Source      string                 `json:"source"`
	ExceptionID ExceptionID            `json:"exception_id,omitempty"`
	Provenance  *provenance.SourceRef  `json:"provenance,omitempty"`
	Trace       []GenerationTraceStep  `json:"trace"`
}

type finiteTablePayload struct {
	SchemaVersion  int                     `json:"schema_version"`
	Normalization  NormalizationForm       `json:"normalization"`
	Language       features.LanguageID     `json:"language"`
	ProfileVersion features.ProfileVersion `json:"profile_version"`
	Readings       []Reading               `json:"readings"`
}

type finiteTableDocument struct {
	SchemaVersion  int                     `json:"schema_version"`
	Normalization  NormalizationForm       `json:"normalization"`
	Language       features.LanguageID     `json:"language"`
	ProfileVersion features.ProfileVersion `json:"profile_version"`
	Readings       []Reading               `json:"readings"`
	SHA256         string                  `json:"sha256"`
}

type FiniteRelationTable struct {
	payload   finiteTablePayload
	profile   *features.LanguageFeatureProfile
	canonical []byte
	digest    string
	maxBytes  int
}

func NormalizeSurface(surface string, budget NormalizationBudget) (string, error) {
	if budget.MaxInputBytes <= 0 || budget.MaxInputBytes > MaxSurfaceBytes || budget.MaxOutputBytes <= 0 || budget.MaxOutputBytes > MaxSurfaceBytes || budget.MaxCodepoints <= 0 || budget.MaxCodepoints > MaxSurfaceCodepoints {
		return "", morphError(status.OutcomeInvalid, "INVALID_NORMALIZATION_BUDGET", "budget.normalization", "normalization limits must be positive and cannot exceed package ceilings")
	}
	if len(surface) == 0 || len(surface) > budget.MaxInputBytes || !utf8.ValidString(surface) {
		return "", morphError(status.OutcomeInvalid, "INVALID_SURFACE", "surface", "surface is empty, invalid UTF-8 or exceeds its input bound")
	}
	if utf8.RuneCountInString(surface) > budget.MaxCodepoints {
		return "", morphLimit("surface_codepoints", utf8.RuneCountInString(surface), budget.MaxCodepoints)
	}
	normalized := norm.NFC.String(surface)
	if len(normalized) > budget.MaxOutputBytes {
		return "", morphLimit("normalized_surface_bytes", len(normalized), budget.MaxOutputBytes)
	}
	if utf8.RuneCountInString(normalized) > budget.MaxCodepoints {
		return "", morphLimit("normalized_surface_codepoints", utf8.RuneCountInString(normalized), budget.MaxCodepoints)
	}
	return normalized, nil
}

func (budget FiniteTableBudget) Validate() error {
	if budget.MaxReadings <= 0 || budget.MaxReadings > MaxFiniteTableReadings || budget.MaxBytes <= 0 || budget.MaxBytes > MaxFiniteTableBytes || budget.MaxSurface <= 0 || budget.MaxSurface > MaxSurfaceBytes {
		return morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_BUDGET", "budget", "finite table limits must be positive and cannot exceed package ceilings")
	}
	return nil
}

func BuildFiniteRelationTable(input []Reading, profile *features.LanguageFeatureProfile, budget FiniteTableBudget) (*FiniteRelationTable, error) {
	if profile == nil {
		return nil, morphError(status.OutcomeUnknown, "FEATURE_PROFILE_UNAVAILABLE", "profile", "finite table compilation requires a feature profile")
	}
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	if len(input) > budget.MaxReadings {
		return nil, morphLimit("finite_table_readings", len(input), budget.MaxReadings)
	}
	readings := make([]Reading, len(input))
	seen := make(map[string]struct{}, len(input))
	encodedRows := 0
	for index, source := range input {
		reading := cloneReading(source)
		path := fmt.Sprintf("readings[%d]", index)
		if reading.Features.Language != profile.Language() || reading.Features.ProfileVersion != profile.Version() {
			return nil, morphError(status.OutcomeIncomplete, "READING_PROFILE_MISMATCH", path+".features", "reading and table profile language/version differ")
		}
		if err := features.ValidateBundle(profile, reading.Features); err != nil {
			return nil, wrapError(err)
		}
		lexemeID, err := ids.ParseLexemeID(reading.LexemeID)
		if err != nil || lexemeID.String() != reading.LexemeID {
			return nil, morphError(status.OutcomeInvalid, "INVALID_READING_LEXEME_ID", path+".lexeme_id", "reading lexeme ID is invalid")
		}
		if !validMorphToken(reading.POS) || !validMorphToken(string(reading.ParadigmID)) || !validMorphToken(string(reading.CellID)) {
			return nil, morphError(status.OutcomeInvalid, "INVALID_READING_TARGET", path, "reading POS, paradigm and cell IDs must be bounded tokens")
		}
		if len(reading.Trace) > MaxGenerationTrace {
			return nil, morphLimit("reading_trace", len(reading.Trace), MaxGenerationTrace)
		}
		if err := validateReadingTrace(reading.Trace); err != nil {
			return nil, morphError(status.OutcomeInvalid, "INVALID_READING_TRACE", path+".trace", err.Error())
		}
		if err := validateReadingSource(reading); err != nil {
			return nil, morphError(status.OutcomeInvalid, "INVALID_READING_SOURCE", path+".source", err.Error())
		}
		reading.Features.Values = canonicalFeatureValues(reading.Features.Values)
		reading.Trace = cloneGenerationTrace(reading.Trace)
		normalizedSurface, err := NormalizeSurface(reading.Surface, NormalizationBudget{MaxInputBytes: budget.MaxSurface, MaxOutputBytes: budget.MaxSurface, MaxCodepoints: MaxSurfaceCodepoints})
		if err != nil {
			return nil, err
		}
		reading.Surface = normalizedSurface
		providedID := reading.ID
		reading.ID = ""
		row, err := json.Marshal(reading)
		if err != nil {
			return nil, morphError(status.OutcomeInvalid, "INVALID_READING_JSON", path, "reading cannot be serialized")
		}
		encodedRows += len(row) + 64 + 96
		if encodedRows > budget.MaxBytes {
			return nil, morphLimit("finite_table_bytes", encodedRows, budget.MaxBytes)
		}
		reading.ID = readingDigest(reading)
		if providedID != "" && providedID != reading.ID {
			return nil, morphError(status.OutcomeInvalid, "READING_ID_MISMATCH", path+".id", "reading ID does not match its canonical content")
		}
		if _, exists := seen[reading.ID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_READING", path, "finite table contains an identical reading more than once")
		}
		seen[reading.ID] = struct{}{}
		readings[index] = reading
	}
	sort.Slice(readings, func(i, j int) bool { return readingLess(readings[i], readings[j]) })
	payload := finiteTablePayload{SchemaVersion: FiniteTableSchemaVersion, Normalization: NormalizationNFC, Language: profile.Language(), ProfileVersion: profile.Version(), Readings: readings}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "table", "finite table cannot be serialized")
	}
	if len(canonical) > budget.MaxBytes {
		return nil, morphLimit("finite_table_bytes", len(canonical), budget.MaxBytes)
	}
	digest := sha256.Sum256(canonical)
	return &FiniteRelationTable{payload: payload, profile: profile, canonical: append([]byte(nil), canonical...), digest: hex.EncodeToString(digest[:]), maxBytes: budget.MaxBytes}, nil
}

func DecodeFiniteRelationTable(data []byte, profile *features.LanguageFeatureProfile, budget FiniteTableBudget) (*FiniteRelationTable, error) {
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	if len(data) > budget.MaxBytes {
		return nil, morphLimit("finite_table_bytes", len(data), budget.MaxBytes)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "document", "finite table must be a JSON object")
	}
	allowed := map[string]bool{"schema_version": true, "normalization": true, "language": true, "profile_version": true, "readings": true, "sha256": true}
	for field := range fields {
		if !allowed[field] {
			return nil, morphError(status.OutcomeInvalid, "UNKNOWN_FINITE_TABLE_FIELD", "document."+field, "finite table contains an unknown or mis-cased field")
		}
	}
	if len(fields) != 6 {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "document", "finite table must contain exactly six top-level fields")
	}
	for _, field := range []string{"schema_version", "normalization", "language", "profile_version", "readings", "sha256"} {
		if _, exists := fields[field]; !exists {
			return nil, morphError(status.OutcomeInvalid, "MISSING_FINITE_TABLE_FIELD", "document."+field, "finite table field is required")
		}
	}
	if bytes.Equal(bytes.TrimSpace(fields["readings"]), []byte("null")) {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "document.readings", "readings cannot be null")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document finiteTableDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "document", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, morphError(status.OutcomeInvalid, "TRAILING_JSON", "document", "expected exactly one finite table JSON value")
	}
	if document.SchemaVersion != FiniteTableSchemaVersion || document.Normalization != NormalizationNFC {
		return nil, morphError(status.OutcomeUnsupported, "UNSUPPORTED_FINITE_TABLE_SCHEMA", "document", "finite table schema or normalization form is unsupported")
	}
	table, err := BuildFiniteRelationTable(document.Readings, profile, budget)
	if err != nil {
		return nil, err
	}
	if table.payload.Language != document.Language || table.payload.ProfileVersion != document.ProfileVersion || table.digest != document.SHA256 {
		return nil, morphError(status.OutcomeInvalid, "FINITE_TABLE_DIGEST_MISMATCH", "document.sha256", "finite table metadata or digest does not match its canonical readings")
	}
	return table, nil
}

func (table *FiniteRelationTable) CanonicalBytes() []byte {
	if table == nil {
		return nil
	}
	return append([]byte(nil), table.canonical...)
}

func (table *FiniteRelationTable) SHA256() string {
	if table == nil {
		return ""
	}
	return table.digest
}

func (table *FiniteRelationTable) Readings() []Reading {
	if table == nil {
		return nil
	}
	result := make([]Reading, len(table.payload.Readings))
	for index, reading := range table.payload.Readings {
		result[index] = cloneReading(reading)
	}
	return result
}

func (table *FiniteRelationTable) Encode() ([]byte, error) {
	if table == nil {
		return nil, morphError(status.OutcomeUnknown, "FINITE_TABLE_UNAVAILABLE", "table", "finite table is unavailable")
	}
	document := finiteTableDocument{SchemaVersion: table.payload.SchemaVersion, Normalization: table.payload.Normalization, Language: table.payload.Language, ProfileVersion: table.payload.ProfileVersion, Readings: table.Readings(), SHA256: table.digest}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, morphError(status.OutcomeInvalid, "INVALID_FINITE_TABLE_JSON", "table", "finite table cannot be encoded")
	}
	if len(encoded) > table.maxBytes {
		return nil, morphLimit("finite_table_bytes", len(encoded), table.maxBytes)
	}
	return encoded, nil
}

func cloneReading(source Reading) Reading {
	result := source
	result.Features = features.FeatureBundle{Language: source.Features.Language, ProfileVersion: source.Features.ProfileVersion, Values: cloneFeatureValues(source.Features.Values)}
	result.Trace = append([]GenerationTraceStep(nil), source.Trace...)
	for index, step := range source.Trace {
		if step.Applied != nil {
			applied := *step.Applied
			result.Trace[index].Applied = &applied
		}
	}
	if source.Provenance != nil {
		provenanceCopy := *source.Provenance
		result.Provenance = &provenanceCopy
	}
	return result
}

func canonicalFeatureValues(values []features.FeatureValue) []features.FeatureValue {
	result := cloneFeatureValues(values)
	sort.Slice(result, func(i, j int) bool { return result[i].Feature < result[j].Feature })
	return result
}

func readingDigest(reading Reading) string {
	reading.ID = ""
	encoded, _ := json.Marshal(reading)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func readingLess(left, right Reading) bool {
	if left.Surface != right.Surface {
		return left.Surface < right.Surface
	}
	if left.LexemeID != right.LexemeID {
		return left.LexemeID < right.LexemeID
	}
	if left.CellID != right.CellID {
		return left.CellID < right.CellID
	}
	leftFeatures, rightFeatures := readingFeatureSignature(left.Features.Values), readingFeatureSignature(right.Features.Values)
	if leftFeatures != rightFeatures {
		return leftFeatures < rightFeatures
	}
	if left.Source != right.Source {
		return left.Source < right.Source
	}
	if left.ExceptionID != right.ExceptionID {
		return left.ExceptionID < right.ExceptionID
	}
	return left.ID < right.ID
}

func readingFeatureSignature(values []features.FeatureValue) string {
	var builder strings.Builder
	for _, value := range values {
		builder.WriteString(string(value.Feature))
		builder.WriteByte('=')
		builder.WriteString(string(value.State))
		builder.WriteByte(':')
		builder.WriteString(string(value.Value))
		builder.WriteByte(';')
	}
	return builder.String()
}

func validateReadingSource(reading Reading) error {
	switch reading.Source {
	case ExceptionSourceRegular:
		if reading.ExceptionID != "" {
			return fmt.Errorf("regular reading cannot carry an exception identity")
		}
		if reading.Provenance != nil {
			if err := reading.Provenance.Validate(); err != nil {
				return err
			}
			foundTrace := false
			for _, step := range reading.Trace {
				if step.Operation != "" && step.SourceID == reading.Provenance.ID {
					foundTrace = true
					break
				}
			}
			if !foundTrace {
				return fmt.Errorf("regular reading provenance must be retained in a procedure trace")
			}
		}
	case ExceptionSource:
		if !validMorphToken(string(reading.ExceptionID)) || reading.Provenance == nil {
			return fmt.Errorf("exception reading requires its rule ID and provenance")
		}
		if err := reading.Provenance.Validate(); err != nil {
			return err
		}
		foundTrace := false
		for _, step := range reading.Trace {
			if step.Exception == reading.ExceptionID && step.SourceID == reading.Provenance.ID {
				foundTrace = true
				break
			}
		}
		if !foundTrace {
			return fmt.Errorf("exception reading must retain a trace to its rule and source")
		}
	default:
		return fmt.Errorf("reading source must be regular or exception")
	}
	return nil
}

func validateReadingTrace(trace []GenerationTraceStep) error {
	for index, step := range trace {
		if err := validateTraceContext(step); err != nil {
			return fmt.Errorf("trace step %d: %w", index, err)
		}
		if step.Operation != "" {
			if !step.Operation.Valid() || step.Applied == nil || step.OutputSize < 0 || step.OutputSize > procedure.MaxOutputBytes || step.Exception != "" || step.Effect != "" {
				return fmt.Errorf("trace operation %d is malformed", index)
			}
			continue
		}
		if step.Exception == "" || !validMorphToken(string(step.Exception)) || len(step.Exception) > MaxExceptionIDBytes || step.Applied != nil || step.OutputSize != 0 || !validExceptionEffectKind(step.Effect) || len(step.SourceID) == 0 || len(step.SourceID) > provenance.MaxSourceIDBytes {
			return fmt.Errorf("trace exception %d is malformed", index)
		}
		if strings.TrimSpace(step.SourceID) != step.SourceID {
			return fmt.Errorf("trace exception %d source ID is malformed", index)
		}
	}
	return nil
}

func validateTraceContext(step GenerationTraceStep) error {
	if step.ProcedureID != "" {
		if len(step.ProcedureID) > status.MaxTraceIDBytes {
			return fmt.Errorf("procedure ID exceeds trace limit")
		}
		procedureID, err := ids.ParseProcedureID(step.ProcedureID)
		if err != nil || procedureID.String() != step.ProcedureID {
			return fmt.Errorf("procedure ID is malformed")
		}
	}
	if step.ParadigmID != "" && !validMorphToken(string(step.ParadigmID)) {
		return fmt.Errorf("paradigm ID is malformed")
	}
	if step.CellID != "" && !validMorphToken(string(step.CellID)) {
		return fmt.Errorf("cell ID is malformed")
	}
	if (step.ParadigmID == "") != (step.CellID == "") {
		return fmt.Errorf("paradigm and cell IDs must be provided together")
	}
	if step.SourceID != "" && (len(step.SourceID) > provenance.MaxSourceIDBytes || strings.TrimSpace(step.SourceID) != step.SourceID) {
		return fmt.Errorf("source ID is malformed")
	}
	if step.Operation != "" && step.ProcedureID != "" && step.ParadigmID == "" {
		return fmt.Errorf("procedure trace requires paradigm and cell IDs")
	}
	return nil
}

func validExceptionEffectKind(effect ExceptionEffectKind) bool {
	switch effect {
	case ExceptionEffectBlock, ExceptionEffectReplace, ExceptionEffectAddVariant, ExceptionEffectStemOverride, ExceptionEffectSuppletion:
		return true
	default:
		return false
	}
}

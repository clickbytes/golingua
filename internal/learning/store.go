// Package learning stores explicit, local morphology corrections. It does not
// decide whether a correction is linguistically valid or compile its rule.
package learning

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	SchemaVersion          = 1
	MaxStoreBytes          = 4 << 20
	MaxRecords             = 4_096
	MaxIDBytes             = 128
	MaxDatasetVersionBytes = 128
	MaxActorBytes          = 256
	MaxJSONDepth           = 64
)

// Store is a versioned collection of local correction records. Records retain
// revoked entries so the correction history remains auditable.
type Store struct {
	SchemaVersion int      `json:"schema_version"`
	Records       []Record `json:"records"`
}

// Record pins one explicit replacement to one lexeme-cell and source digest.
type Record struct {
	ID                   string                  `json:"id"`
	DatasetID            string                  `json:"dataset_id"`
	DatasetVersion       string                  `json:"dataset_version"`
	Language             features.LanguageID     `json:"language"`
	ProfileVersion       features.ProfileVersion `json:"profile_version"`
	BaselineDigest       string                  `json:"baseline_digest"`
	Features             []features.FeatureValue `json:"features"`
	BaselineForm         morph.GeneratedForm     `json:"baseline_form"`
	Rule                 morph.ExceptionRule     `json:"rule"`
	OriginalForm         string                  `json:"original_form"`
	CorrectedForm        string                  `json:"corrected_form"`
	InitiatedBy          string                  `json:"initiated_by"`
	PreviousCorrectionID string                  `json:"previous_correction_id,omitempty"`
	Author               provenance.SourceRef    `json:"author_provenance"`
	RecordedAt           string                  `json:"recorded_at"`
	Active               bool                    `json:"active"`
	Revocation           *Revocation             `json:"revocation,omitempty"`
}

// Revocation records who withdrew a correction and when.
type Revocation struct {
	Actor string `json:"actor"`
	Time  string `json:"time"`
}

// mutationMu is an additional in-process guard. Store updates also acquire an
// OS lock on a stable sibling lock file for cross-process read-modify-write.
var mutationMu sync.Mutex

// Load reads and validates an existing local correction store. A missing file
// is an error; only Add treats a missing file as a new empty store.
func Load(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("learning store path is empty")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open learning store: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, MaxStoreBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read learning store: %w", err)
	}
	if len(data) > MaxStoreBytes {
		return nil, fmt.Errorf("learning store exceeds %d bytes", MaxStoreBytes)
	}
	if err := validateStrictJSON(data); err != nil {
		return nil, fmt.Errorf("learning store JSON: %w", err)
	}
	if err := validateDocumentShape(data); err != nil {
		return nil, fmt.Errorf("learning store shape: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var store Store
	if err := decoder.Decode(&store); err != nil {
		return nil, fmt.Errorf("decode learning store: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return nil, fmt.Errorf("learning store trailing data: %w", err)
	}
	if err := validateStore(store); err != nil {
		return nil, err
	}
	return &store, nil
}

// Add appends an active one-cell correction. An identical record is
// idempotent. Different active corrections for the same language, lexeme,
// cell, and baseline digest are rejected.
func Add(path string, record Record) (resultErr error) {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if path == "" {
		return errors.New("learning store path is empty")
	}
	release, err := acquireStoreLock(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := release(); resultErr == nil && err != nil {
			resultErr = err
		}
	}()
	if err := validateRecord(record); err != nil {
		return fmt.Errorf("invalid learning record: %w", err)
	}
	if !record.Active || record.Revocation != nil {
		return errors.New("new learning records must be active and have no revocation")
	}

	store, err := Load(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		store = &Store{SchemaVersion: SchemaVersion, Records: []Record{}}
	}

	for _, existing := range store.Records {
		if existing.ID == record.ID {
			if reflect.DeepEqual(existing, record) {
				return syncPersistedStore(path)
			}
			return fmt.Errorf("learning record ID %q already exists with different contents", record.ID)
		}
		if existing.Active && sameTarget(existing, record) {
			return fmt.Errorf("active correction already exists for language %q, lexeme %q, cell %q, and baseline %q", record.Language, record.Rule.Selector.LexemeID, record.Rule.Selector.CellID, record.BaselineDigest)
		}
	}
	if err := validatePreviousCorrection(store.Records, record); err != nil {
		return err
	}
	if len(store.Records) >= MaxRecords {
		return fmt.Errorf("learning record count exceeds %d", MaxRecords)
	}
	activeCount := 0
	for _, existing := range store.Records {
		if existing.Active {
			activeCount++
		}
	}
	if activeCount >= morph.MaxExceptionRules {
		return fmt.Errorf("active learning record count exceeds %d", morph.MaxExceptionRules)
	}
	store.Records = append(store.Records, record)
	return save(path, *store)
}

// Revoke deactivates one active record and preserves an audit entry. It errors
// when the store or record is absent, or when the record was already revoked.
func Revoke(path, id, actor, at string) (resultErr error) {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if path == "" {
		return errors.New("learning store path is empty")
	}
	release, err := acquireStoreLock(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := release(); resultErr == nil && err != nil {
			resultErr = err
		}
	}()
	if !validIdentifier(id, morph.MaxExceptionIDBytes) {
		return errors.New("revocation record ID is invalid")
	}
	if !validText(actor, MaxActorBytes) {
		return errors.New("revocation actor is empty, invalid, or too long")
	}
	if err := validateTime(at); err != nil {
		return fmt.Errorf("revocation time: %w", err)
	}
	store, err := Load(path)
	if err != nil {
		return err
	}
	for index := range store.Records {
		record := &store.Records[index]
		if record.ID != id {
			continue
		}
		if !record.Active || record.Revocation != nil {
			return fmt.Errorf("learning record %q is already revoked", id)
		}
		record.Active = false
		record.Revocation = &Revocation{Actor: actor, Time: at}
		return save(path, *store)
	}
	return fmt.Errorf("learning record %q does not exist", id)
}

func save(path string, store Store) error {
	if err := validateStore(store); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encode learning store: %w", err)
	}
	data = append(data, '\n')
	if len(data) > MaxStoreBytes {
		return fmt.Errorf("learning store exceeds %d bytes", MaxStoreBytes)
	}

	directory := filepath.Dir(path)
	base := filepath.Base(path)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return errors.New("learning store path must name a file")
	}
	temporary, err := os.CreateTemp(directory, "."+base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary learning store: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set learning store permissions: %w", err)
	}
	written, err := temporary.Write(data)
	if err == nil && written != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary learning store: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary learning store: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary learning store: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace learning store: %w", err)
	}

	directoryFile, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open learning store directory for sync: %w", err)
	}
	if err := directoryFile.Sync(); err != nil {
		_ = directoryFile.Close()
		return fmt.Errorf("sync learning store directory: %w", err)
	}
	if err := directoryFile.Close(); err != nil {
		return fmt.Errorf("close learning store directory: %w", err)
	}
	return nil
}

func syncPersistedStore(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open learning store for retry sync: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync learning store after idempotent retry: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close learning store after idempotent retry: %w", err)
	}
	directoryFile, err := os.Open(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("open learning store directory after idempotent retry: %w", err)
	}
	if err := directoryFile.Sync(); err != nil {
		_ = directoryFile.Close()
		return fmt.Errorf("sync learning store directory after idempotent retry: %w", err)
	}
	if err := directoryFile.Close(); err != nil {
		return fmt.Errorf("close learning store directory after idempotent retry: %w", err)
	}
	return nil
}

func acquireStoreLock(path string) (func() error, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve learning store path for lock: %w", err)
	}
	directory := filepath.Dir(absolutePath)
	canonicalDirectory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve learning store directory for lock: %w", err)
	}
	lockPath := filepath.Join(canonicalDirectory, filepath.Base(absolutePath)+".lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open learning store lock: %w", err)
	}
	release, err := lockPlatformFile(file)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("acquire learning store lock: %w", err)
	}
	return func() error {
		unlockErr := release()
		closeErr := file.Close()
		if unlockErr != nil {
			return fmt.Errorf("release learning store lock: %w", unlockErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close learning store lock: %w", closeErr)
		}
		return nil
	}, nil
}

func validateStore(store Store) error {
	if store.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported learning store schema version %d", store.SchemaVersion)
	}
	if store.Records == nil {
		return errors.New("learning store records must be an array")
	}
	if len(store.Records) > MaxRecords {
		return fmt.Errorf("learning record count exceeds %d", MaxRecords)
	}
	seenIDs := make(map[string]struct{}, len(store.Records))
	activeTargets := make(map[string]string)
	latestByTarget := make(map[string]Record)
	activeCount := 0
	for index, record := range store.Records {
		if err := validateRecord(record); err != nil {
			return fmt.Errorf("records[%d]: %w", index, err)
		}
		if _, exists := seenIDs[record.ID]; exists {
			return fmt.Errorf("records[%d]: duplicate record ID %q", index, record.ID)
		}
		seenIDs[record.ID] = struct{}{}
		if record.Active {
			activeCount++
			key := targetKey(record)
			if previous, exists := activeTargets[key]; exists {
				return fmt.Errorf("records[%d]: active correction conflicts with record %q", index, previous)
			}
			activeTargets[key] = record.ID
		}
		key := targetKey(record)
		if previous, exists := latestByTarget[key]; exists && !previous.Active {
			if record.PreviousCorrectionID != previous.ID {
				return fmt.Errorf("records[%d]: correction must identify revoked predecessor %q", index, previous.ID)
			}
		} else if record.PreviousCorrectionID != "" {
			return fmt.Errorf("records[%d]: previous correction ID does not identify the latest revoked correction for this target", index)
		}
		latestByTarget[key] = record
	}
	if activeCount > morph.MaxExceptionRules {
		return fmt.Errorf("active learning record count exceeds %d", morph.MaxExceptionRules)
	}
	return nil
}

func validatePreviousCorrection(records []Record, record Record) error {
	var latest *Record
	for index := range records {
		candidate := &records[index]
		if sameTarget(*candidate, record) {
			latest = candidate
		}
	}
	if latest == nil {
		if record.PreviousCorrectionID != "" {
			return errors.New("previous correction ID was supplied without an earlier correction for this target")
		}
		return nil
	}
	if latest.Active {
		return errors.New("target already has an active correction")
	}
	if record.PreviousCorrectionID != latest.ID {
		return fmt.Errorf("new correction must identify revoked predecessor %q", latest.ID)
	}
	return nil
}

func validateRecord(record Record) error {
	if !validIdentifier(record.ID, morph.MaxExceptionIDBytes) || string(record.Rule.ID) != record.ID {
		return errors.New("record ID must be valid and equal its exception rule ID")
	}
	if !validIdentifier(record.DatasetID, MaxIDBytes) || !validDatasetVersion(record.DatasetVersion) {
		return errors.New("dataset ID and version are required bounded identifiers")
	}
	if !validIdentifier(string(record.Language), MaxIDBytes) || record.Rule.Selector.Language != record.Language {
		return errors.New("record language must match its exception selector")
	}
	if !validIdentifier(string(record.ProfileVersion), features.MaxTokenBytes) {
		return errors.New("profile version is required and must be a bounded token")
	}
	if !validDigest(record.BaselineDigest) {
		return errors.New("baseline digest must be a SHA-256 digest")
	}
	if len(record.Features) == 0 || len(record.Features) > features.MaxBundleValues || !reflect.DeepEqual(record.Features, record.BaselineForm.Features) {
		return errors.New("feature values must be a complete bounded set identical to the baseline form")
	}
	featureValues := make(map[features.FeatureID]features.FeatureValue, len(record.Features))
	for _, feature := range record.Features {
		if !validIdentifier(string(feature.Feature), features.MaxTokenBytes) {
			return errors.New("feature set contains an invalid feature ID")
		}
		if _, exists := featureValues[feature.Feature]; exists {
			return errors.New("feature set repeats a feature ID")
		}
		if feature.Variable != nil || (feature.State != features.StateKnown && feature.State != features.StateAbsent) {
			return errors.New("baseline feature values must be resolved known or absent values")
		}
		if feature.State == features.StateKnown {
			if !validIdentifier(string(feature.Value), features.MaxTokenBytes) {
				return errors.New("known feature value is invalid")
			}
		} else if feature.Value != "" {
			return errors.New("absent feature value cannot carry a value")
		}
		featureValues[feature.Feature] = feature
	}
	if !validText(record.OriginalForm, morph.MaxPrincipalPartBytes) || !validText(record.CorrectedForm, morph.MaxPrincipalPartBytes) {
		return errors.New("original and corrected forms must be nonempty, valid, and bounded")
	}
	if record.OriginalForm == record.CorrectedForm {
		return errors.New("corrected form must differ from the baseline form")
	}
	if !validText(record.InitiatedBy, MaxActorBytes) {
		return errors.New("record initiator is required, valid, and bounded")
	}
	baseline := record.BaselineForm
	if baseline.Surface != record.OriginalForm || baseline.Language != record.Language || baseline.ProfileVersion != record.ProfileVersion ||
		baseline.LexemeID != record.Rule.Selector.LexemeID || baseline.POS != record.Rule.Selector.POS ||
		string(baseline.ParadigmID) != string(record.Rule.Selector.ParadigmID) || baseline.CellID != record.Rule.Selector.CellID ||
		len(baseline.Trace) == 0 || len(baseline.Trace) > morph.MaxGenerationTrace {
		return errors.New("baseline generated form does not match the correction selector and original form")
	}
	if baseline.Source != morph.ExceptionSourceRegular && baseline.Source != morph.ExceptionSource {
		return errors.New("baseline generated form has an invalid source kind")
	}
	if baseline.ExceptionID != "" && !validIdentifier(string(baseline.ExceptionID), morph.MaxExceptionIDBytes) {
		return errors.New("baseline generated form has an invalid exception ID")
	}
	if baseline.Provenance != nil {
		if err := baseline.Provenance.Validate(); err != nil {
			return fmt.Errorf("baseline generated form provenance: %w", err)
		}
	}
	for index, step := range baseline.Trace {
		if step.OutputSize < 0 || step.OutputSize > morph.MaxPrincipalPartBytes ||
			(step.SourceID != "" && !validSourceID(step.SourceID)) ||
			(step.Exception != "" && !validIdentifier(string(step.Exception), morph.MaxExceptionIDBytes)) {
			return fmt.Errorf("baseline trace step %d is invalid", index)
		}
	}
	selectorFeatures := make(map[features.FeatureID]features.FeatureValue, len(record.Rule.Selector.Features))
	for _, selector := range record.Rule.Selector.Features {
		selectorFeatures[selector.Feature] = features.FeatureValue{Feature: selector.Feature, State: features.StateKnown, Value: selector.Value}
	}
	for featureID, selector := range selectorFeatures {
		if baselineValue, exists := featureValues[featureID]; !exists || baselineValue != selector {
			return fmt.Errorf("exception selector feature %q conflicts with the baseline feature set", featureID)
		}
	}
	if record.Rule.Effect.Kind != morph.ExceptionEffectReplace || len(record.Rule.Effect.Forms) != 1 || record.Rule.Effect.Forms[0] != record.CorrectedForm {
		return errors.New("rule must replace exactly one cell with the corrected form")
	}
	if record.Rule.Scope != morph.ExceptionScopeLexemeCell || record.Rule.Selector.POS == "" || record.Rule.Selector.ParadigmID == "" || record.Rule.Selector.LexemeID == "" || record.Rule.Selector.CellID == "" {
		return errors.New("rule must target one exact lexeme_cell")
	}
	if record.PreviousCorrectionID != "" && (!validIdentifier(record.PreviousCorrectionID, morph.MaxExceptionIDBytes) || record.PreviousCorrectionID == record.ID) {
		return errors.New("previous correction ID is invalid or self-referential")
	}
	lexemeID, lexemeErr := ids.ParseLexemeID(record.Rule.Selector.LexemeID)
	if !validIdentifier(record.Rule.Selector.POS, MaxIDBytes) || !validIdentifier(string(record.Rule.Selector.ParadigmID), MaxIDBytes) || !validIdentifier(string(record.Rule.Selector.CellID), MaxIDBytes) || lexemeErr != nil || lexemeID.String() != record.Rule.Selector.LexemeID {
		return errors.New("rule lexeme-cell selector contains an invalid ID")
	}
	if record.Rule.SchemaVersion != morph.ExceptionSchemaVersion || record.Rule.Priority < 0 || record.Rule.Priority > morph.MaxExceptionPriority {
		return errors.New("exception rule version or priority is invalid")
	}
	if !validIdentifier(record.Rule.ReviewedBy, MaxActorBytes) {
		return errors.New("exception rule reviewer is required")
	}
	if err := validateTime(record.Rule.ReviewedAt); err != nil {
		return fmt.Errorf("exception rule review time: %w", err)
	}
	if err := record.Rule.Provenance.Validate(); err != nil {
		return fmt.Errorf("exception rule provenance: %w", err)
	}
	if len(record.Rule.Selector.Features) > features.MaxDimensions {
		return errors.New("exception rule has too many feature selectors")
	}
	seenFeatures := make(map[features.FeatureID]struct{}, len(record.Rule.Selector.Features))
	for _, feature := range record.Rule.Selector.Features {
		if !validIdentifier(string(feature.Feature), MaxIDBytes) || !validIdentifier(string(feature.Value), MaxIDBytes) {
			return errors.New("exception rule has an invalid feature selector")
		}
		if _, exists := seenFeatures[feature.Feature]; exists {
			return errors.New("exception rule repeats a feature selector")
		}
		seenFeatures[feature.Feature] = struct{}{}
	}
	if len(record.Rule.OverrideOf) > morph.MaxExceptionOverrides {
		return errors.New("exception rule has too many override references")
	}
	seenOverrides := make(map[morph.ExceptionID]struct{}, len(record.Rule.OverrideOf))
	for _, id := range record.Rule.OverrideOf {
		if !validIdentifier(string(id), morph.MaxExceptionIDBytes) || id == record.Rule.ID {
			return errors.New("exception rule has an invalid or self-referential override")
		}
		if _, exists := seenOverrides[id]; exists {
			return errors.New("exception rule repeats an override reference")
		}
		seenOverrides[id] = struct{}{}
	}
	if err := record.Author.Validate(); err != nil {
		return fmt.Errorf("author provenance: %w", err)
	}
	if record.Author.Author == "" {
		return errors.New("author provenance must name an author")
	}
	if err := validateTime(record.RecordedAt); err != nil {
		return fmt.Errorf("recorded time: %w", err)
	}
	if record.Active {
		if record.Revocation != nil {
			return errors.New("active record cannot contain a revocation")
		}
	} else {
		if record.Revocation == nil || !validText(record.Revocation.Actor, MaxActorBytes) {
			return errors.New("inactive record requires revocation actor and time")
		}
		if err := validateTime(record.Revocation.Time); err != nil {
			return fmt.Errorf("revocation time: %w", err)
		}
	}
	return nil
}

func validateDocumentShape(data []byte) error {
	fields, err := objectFields(data, []string{"schema_version", "records"}, []string{"schema_version", "records"}, []string{"schema_version", "records"})
	if err != nil {
		return err
	}
	var records []json.RawMessage
	if err := json.Unmarshal(fields["records"], &records); err != nil || records == nil {
		return errors.New("records must be a JSON array")
	}
	if len(records) > MaxRecords {
		return fmt.Errorf("record count exceeds %d", MaxRecords)
	}
	for index, raw := range records {
		path := fmt.Sprintf("records[%d]", index)
		recordFields, err := objectFields(raw,
			[]string{"id", "dataset_id", "dataset_version", "language", "profile_version", "baseline_digest", "features", "baseline_form", "rule", "original_form", "corrected_form", "initiated_by", "previous_correction_id", "author_provenance", "recorded_at", "active", "revocation"},
			[]string{"id", "dataset_id", "dataset_version", "language", "profile_version", "baseline_digest", "features", "baseline_form", "rule", "original_form", "corrected_form", "initiated_by", "author_provenance", "recorded_at", "active"},
			[]string{"id", "dataset_id", "dataset_version", "language", "profile_version", "baseline_digest", "features", "baseline_form", "rule", "original_form", "corrected_form", "initiated_by", "author_provenance", "recorded_at", "active", "revocation"})
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := validateRuleShape(recordFields["rule"]); err != nil {
			return fmt.Errorf("%s.rule: %w", path, err)
		}
		if err := validateFeatureValuesShape(recordFields["features"]); err != nil {
			return fmt.Errorf("%s.features: %w", path, err)
		}
		if err := validateGeneratedFormShape(recordFields["baseline_form"]); err != nil {
			return fmt.Errorf("%s.baseline_form: %w", path, err)
		}
		if _, err := objectFields(recordFields["author_provenance"],
			[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"},
			[]string{"id", "version", "license"},
			[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}); err != nil {
			return fmt.Errorf("%s.author_provenance: %w", path, err)
		}
		if _, exists := recordFields["revocation"]; exists {
			if _, err := objectFields(recordFields["revocation"], []string{"actor", "time"}, []string{"actor", "time"}, []string{"actor", "time"}); err != nil {
				return fmt.Errorf("%s.revocation: %w", path, err)
			}
		}
	}
	return nil
}

func validateFeatureValuesShape(data []byte) error {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil || values == nil {
		return errors.New("feature values must be a JSON array")
	}
	if len(values) == 0 || len(values) > features.MaxBundleValues {
		return errors.New("feature values count is empty or exceeds its limit")
	}
	for index, raw := range values {
		fields, err := objectFields(raw,
			[]string{"feature", "state", "value", "variable"},
			[]string{"feature", "state"},
			[]string{"feature", "state"})
		if err != nil {
			return fmt.Errorf("values[%d]: %w", index, err)
		}
		if variable, exists := fields["variable"]; exists {
			if _, err := objectFields(variable, []string{"id", "scope"}, []string{"id", "scope"}, []string{"id", "scope"}); err != nil {
				return fmt.Errorf("values[%d].variable: %w", index, err)
			}
		}
	}
	return nil
}

func validateGeneratedFormShape(data []byte) error {
	fields, err := objectFields(data,
		[]string{"surface", "lexeme_id", "language", "profile_version", "pos", "paradigm_id", "cell_id", "features", "source", "exception_id", "provenance", "trace"},
		[]string{"surface", "lexeme_id", "language", "profile_version", "pos", "paradigm_id", "cell_id", "features", "source", "trace"},
		[]string{"surface", "lexeme_id", "language", "profile_version", "pos", "paradigm_id", "cell_id", "features", "source", "trace"})
	if err != nil {
		return err
	}
	if err := validateFeatureValuesShape(fields["features"]); err != nil {
		return fmt.Errorf("features: %w", err)
	}
	var trace []json.RawMessage
	if err := json.Unmarshal(fields["trace"], &trace); err != nil || trace == nil || len(trace) == 0 || len(trace) > morph.MaxGenerationTrace {
		return errors.New("trace must be a nonempty bounded array")
	}
	for index, raw := range trace {
		if _, err := objectFields(raw,
			[]string{"operation", "applied", "output_size", "procedure_id", "paradigm_id", "cell_id", "exception", "effect", "source_id"},
			[]string{},
			[]string{}); err != nil {
			return fmt.Errorf("trace[%d]: %w", index, err)
		}
	}
	if _, exists := fields["provenance"]; exists {
		if _, err := objectFields(fields["provenance"],
			[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"},
			[]string{"id", "version", "license"},
			[]string{"id", "version", "license"}); err != nil {
			return fmt.Errorf("provenance: %w", err)
		}
	}
	return nil
}

func validateRuleShape(data []byte) error {
	fields, err := objectFields(data,
		[]string{"schema_version", "id", "scope", "selector", "effect", "priority", "override_of", "provenance", "reviewed_by", "reviewed_at"},
		[]string{"schema_version", "id", "scope", "selector", "effect", "priority", "provenance", "reviewed_by", "reviewed_at"},
		[]string{"schema_version", "id", "scope", "selector", "effect", "priority", "provenance", "reviewed_by", "reviewed_at", "override_of"})
	if err != nil {
		return err
	}
	selector, err := objectFields(fields["selector"],
		[]string{"language", "pos", "paradigm_id", "lexeme_id", "cell_id", "features"},
		[]string{"language"},
		[]string{"language", "pos", "paradigm_id", "lexeme_id", "cell_id", "features"})
	if err != nil {
		return fmt.Errorf("selector: %w", err)
	}
	if rawFeatures, exists := selector["features"]; exists {
		var featureMatches []json.RawMessage
		if err := json.Unmarshal(rawFeatures, &featureMatches); err != nil || featureMatches == nil {
			return errors.New("selector.features must be an array")
		}
		if len(featureMatches) > features.MaxDimensions {
			return errors.New("selector.features exceeds its limit")
		}
		for index, rawFeature := range featureMatches {
			if _, err := objectFields(rawFeature, []string{"feature", "value"}, []string{"feature", "value"}, []string{"feature", "value"}); err != nil {
				return fmt.Errorf("selector.features[%d]: %w", index, err)
			}
		}
	}
	effect, err := objectFields(fields["effect"],
		[]string{"kind", "forms", "stem_id", "stem_form"},
		[]string{"kind"},
		[]string{"kind", "forms", "stem_id", "stem_form"})
	if err != nil {
		return fmt.Errorf("effect: %w", err)
	}
	if rawForms, exists := effect["forms"]; exists {
		var forms []json.RawMessage
		if err := json.Unmarshal(rawForms, &forms); err != nil || forms == nil {
			return errors.New("effect.forms must be an array")
		}
	}
	provenanceFields, err := objectFields(fields["provenance"],
		[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"},
		[]string{"id", "version", "license"},
		[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"})
	if err != nil {
		return fmt.Errorf("provenance: %w", err)
	}
	if _, pathExists := provenanceFields["path"]; !pathExists {
		if _, uriExists := provenanceFields["uri"]; !uriExists {
			return errors.New("provenance requires path or uri")
		}
	}
	return nil
}

func objectFields(data []byte, allowed, required, nonNull []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, errors.New("expected a JSON object")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	for name := range fields {
		if _, exists := allowedSet[name]; !exists {
			return nil, fmt.Errorf("unknown field %q", name)
		}
	}
	for _, name := range required {
		if _, exists := fields[name]; !exists {
			return nil, fmt.Errorf("required field %q is missing", name)
		}
	}
	for _, name := range nonNull {
		if raw, exists := fields[name]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, fmt.Errorf("field %q cannot be null", name)
		}
	}
	return fields, nil
}

func validateStrictJSON(data []byte) error {
	if !utf8.Valid(data) {
		return errors.New("input is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, 0); err != nil {
		return err
	}
	return requireEOF(decoder)
}

func scanJSONValue(decoder *json.Decoder, depth int) error {
	if depth > MaxJSONDepth {
		return fmt.Errorf("JSON nesting exceeds %d", MaxJSONDepth)
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder, depth+1); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder, depth+1); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func requireEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("expected exactly one JSON value")
		}
		return err
	}
	return nil
}

func validIdentifier(value string, maxBytes int) bool {
	if len(value) == 0 || len(value) > maxBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for index, character := range value {
		if character > unicode.MaxASCII {
			return false
		}
		if index == 0 && (character < 'a' || character > 'z') {
			return false
		}
		if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

func validText(value string, maxBytes int) bool {
	if len(value) == 0 || len(value) > maxBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value || utf8.RuneCountInString(value) > morph.MaxPrincipalPartRunes {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validateTime(value string) error {
	if value == "" || len(value) > 64 {
		return errors.New("RFC3339 timestamp is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return errors.New("timestamp must use RFC3339")
	}
	return nil
}

func validDigest(value string) bool {
	value = strings.TrimPrefix(value, "sha256:")
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validSourceID(value string) bool {
	if len(value) == 0 || len(value) > provenance.MaxSourceIDBytes {
		return false
	}
	for index, character := range value {
		alphaNumeric := character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9'
		if index == 0 && !alphaNumeric {
			return false
		}
		if !alphaNumeric && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func validDatasetVersion(value string) bool {
	if len(value) == 0 || len(value) > MaxDatasetVersionBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func sameTarget(left, right Record) bool {
	return targetKey(left) == targetKey(right)
}

func targetKey(record Record) string {
	featuresCopy := cloneFeatureValues(record.Features)
	sort.Slice(featuresCopy, func(left, right int) bool {
		return featuresCopy[left].Feature < featuresCopy[right].Feature
	})
	canonicalFeatures, _ := json.Marshal(featuresCopy)
	return strings.Join([]string{
		record.DatasetID,
		record.DatasetVersion,
		string(record.Language),
		string(record.ProfileVersion),
		record.Rule.Selector.LexemeID,
		string(record.Rule.Selector.CellID),
		digestValue(record.BaselineDigest),
		string(canonicalFeatures),
	}, "\x00")
}

func digestValue(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

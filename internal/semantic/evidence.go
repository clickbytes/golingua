// Package semantic contains bounded semantic alternatives and their evidence.
package semantic

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultEvidenceMaxRecords = 4_096
	maxEvidenceRecords        = 65_536
	defaultEvidenceMaxText    = 1 << 20
	maxEvidenceTextBytes      = 4 << 20
)

// EvidenceBudget bounds retained evidence records and their total text data.
// It is independent of the parse and Meaning graph budgets.
type EvidenceBudget struct {
	MaxRecords   int
	MaxTextBytes int
}

// DefaultEvidenceBudget returns finite limits for one semantic alternative.
func DefaultEvidenceBudget() EvidenceBudget {
	return EvidenceBudget{MaxRecords: defaultEvidenceMaxRecords, MaxTextBytes: defaultEvidenceMaxText}
}

// Validate rejects disabled or above-ceiling evidence limits.
func (budget EvidenceBudget) Validate() error {
	if budget.MaxRecords <= 0 || budget.MaxRecords > maxEvidenceRecords {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_BUDGET", "budget.max_records", fmt.Sprintf("must be between 1 and %d", maxEvidenceRecords))
	}
	if budget.MaxTextBytes <= 0 || budget.MaxTextBytes > maxEvidenceTextBytes {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_BUDGET", "budget.max_text_bytes", fmt.Sprintf("must be between 1 and %d", maxEvidenceTextBytes))
	}
	return nil
}

// EvidenceTrace links one semantic alternative to the lexical, syntactic and
// licensing records that support it. Evidence is request-scoped, not stored in
// the language-neutral Meaning graph.
type EvidenceTrace struct {
	AlternativeID string              `json:"alternative_id"`
	Outcome       status.OutcomeKind  `json:"outcome"`
	Records       []EvidenceRecord    `json:"records"`
	Diagnostics   []status.Diagnostic `json:"diagnostics,omitempty"`
}

// EvidenceOrigin distinguishes evidence aligned to an analyzed source span
// from evidence produced while realizing a Meaning that has no source text.
// The empty value is source-aligned for backwards compatibility.
type EvidenceOrigin string

const EvidenceOriginGenerated EvidenceOrigin = "generated"

// EvidenceRecord ties one semantic or unresolved claim to source and typed
// lexical/construction evidence. Empty semantic IDs represent unavailable
// evidence; they never license a guessed mapping. MeaningNodeID addresses the
// exact node in this candidate graph. PredicateNodeID additionally identifies
// the owning predicate for argument evidence; it is empty for predicate-local
// lexical evidence. These are neither CORE IDs nor persistent entity IDs.
type EvidenceRecord struct {
	Origin          EvidenceOrigin                  `json:"origin,omitempty"`
	SourceSpan      text.Span                       `json:"source_span"`
	CandidateID     string                          `json:"candidate_id,omitempty"`
	PhraseNodeID    parser.NodeID                   `json:"phrase_node_id,omitempty"`
	MeaningNodeID   meaning.NodeID                  `json:"meaning_node_id,omitempty"`
	PredicateNodeID meaning.NodeID                  `json:"predicate_node_id,omitempty"`
	ReadingID       string                          `json:"reading_id,omitempty"`
	SenseID         ids.SenseID                     `json:"sense_id,omitempty"`
	CoreConceptID   ids.ConceptID                   `json:"core_concept_id,omitempty"`
	BindingKey      string                          `json:"binding_key,omitempty"`
	Role            meaning.Role                    `json:"role,omitempty"`
	Features        []features.FeatureValue         `json:"features,omitempty"`
	PlanID          construction.ConstructionPlanID `json:"plan_id,omitempty"`
	NPPhraseID      string                          `json:"np_phrase_id,omitempty"`
	NPPhrasePlanID  string                          `json:"np_phrase_plan_id,omitempty"`
	FamilyID        ids.ConstructionID              `json:"family_id,omitempty"`
	VariantID       construction.VariantID          `json:"variant_id,omitempty"`
	FrameID         valency.FrameID                 `json:"frame_id,omitempty"`
	Outcome         status.OutcomeKind              `json:"outcome"`
	Reason          string                          `json:"reason,omitempty"`
	Provenance      provenance.SourceRef            `json:"provenance,omitempty"`
}

// MarshalJSON writes typed stable IDs in their external string form. The ID
// structs intentionally keep their representation private, so encoding them
// directly would otherwise produce empty JSON objects.
func (record EvidenceRecord) MarshalJSON() ([]byte, error) {
	type evidenceRecordJSON struct {
		Origin          EvidenceOrigin                  `json:"origin,omitempty"`
		SourceSpan      text.Span                       `json:"source_span"`
		CandidateID     string                          `json:"candidate_id,omitempty"`
		PhraseNodeID    parser.NodeID                   `json:"phrase_node_id,omitempty"`
		MeaningNodeID   meaning.NodeID                  `json:"meaning_node_id,omitempty"`
		PredicateNodeID meaning.NodeID                  `json:"predicate_node_id,omitempty"`
		ReadingID       string                          `json:"reading_id,omitempty"`
		SenseID         string                          `json:"sense_id,omitempty"`
		CoreConceptID   string                          `json:"core_concept_id,omitempty"`
		BindingKey      string                          `json:"binding_key,omitempty"`
		Role            meaning.Role                    `json:"role,omitempty"`
		Features        []features.FeatureValue         `json:"features,omitempty"`
		PlanID          construction.ConstructionPlanID `json:"plan_id,omitempty"`
		NPPhraseID      string                          `json:"np_phrase_id,omitempty"`
		NPPhrasePlanID  string                          `json:"np_phrase_plan_id,omitempty"`
		FamilyID        string                          `json:"family_id,omitempty"`
		VariantID       construction.VariantID          `json:"variant_id,omitempty"`
		FrameID         valency.FrameID                 `json:"frame_id,omitempty"`
		Outcome         status.OutcomeKind              `json:"outcome"`
		Reason          string                          `json:"reason,omitempty"`
		Provenance      provenance.SourceRef            `json:"provenance,omitempty"`
	}
	return json.Marshal(evidenceRecordJSON{
		Origin: record.Origin, SourceSpan: record.SourceSpan, CandidateID: record.CandidateID,
		PhraseNodeID: record.PhraseNodeID, MeaningNodeID: record.MeaningNodeID, PredicateNodeID: record.PredicateNodeID, ReadingID: record.ReadingID,
		SenseID: record.SenseID.String(), CoreConceptID: record.CoreConceptID.String(),
		BindingKey: record.BindingKey, Role: record.Role, Features: record.Features,
		PlanID: record.PlanID, NPPhraseID: record.NPPhraseID, NPPhrasePlanID: record.NPPhrasePlanID, FamilyID: record.FamilyID.String(), VariantID: record.VariantID,
		FrameID: record.FrameID, Outcome: record.Outcome, Reason: record.Reason,
		Provenance: record.Provenance,
	})
}

// BindingKeyFor returns the stable composite key for an exact sense-to-CORE
// binding. It is a key, not a separately allocated identity.
func BindingKeyFor(senseID ids.SenseID, coreConceptID ids.ConceptID) string {
	return senseID.String() + "->" + coreConceptID.String()
}

// Clone returns a defensive copy of all trace-owned slices.
func (trace EvidenceTrace) Clone() EvidenceTrace {
	trace.Records = append([]EvidenceRecord(nil), trace.Records...)
	for index := range trace.Records {
		trace.Records[index].Features = cloneEvidenceFeatures(trace.Records[index].Features)
	}
	trace.Diagnostics = append([]status.Diagnostic(nil), trace.Diagnostics...)
	return trace
}

func cloneEvidenceFeatures(source []features.FeatureValue) []features.FeatureValue {
	result := append([]features.FeatureValue(nil), source...)
	for index := range result {
		if result[index].Variable != nil {
			variable := *result[index].Variable
			result[index].Variable = &variable
		}
	}
	return result
}

// Validate checks the trace against the package's default evidence budget.
func (trace EvidenceTrace) Validate() error {
	return trace.ValidateWithBudget(DefaultEvidenceBudget())
}

// ValidateWithBudget checks IDs, spans, provenance and aggregate text bounds.
func (trace EvidenceTrace) ValidateWithBudget(budget EvidenceBudget) error {
	if err := budget.Validate(); err != nil {
		return err
	}
	if !validEvidenceText(trace.AlternativeID, 256) {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_ALTERNATIVE_ID", "alternative_id", "alternative ID must be non-empty bounded UTF-8 without control characters")
	}
	if len(trace.AlternativeID) > budget.MaxTextBytes {
		return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", fmt.Sprintf("evidence text exceeds %d bytes", budget.MaxTextBytes))
	}
	if !trace.Outcome.Valid() {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_OUTCOME", "outcome", "evidence outcome is unsupported")
	}
	if len(trace.Records) > budget.MaxRecords {
		return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.records", fmt.Sprintf("evidence record count %d exceeds %d", len(trace.Records), budget.MaxRecords))
	}
	if len(trace.Diagnostics) > status.MaxDiagnostics {
		return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.diagnostics", fmt.Sprintf("diagnostic count %d exceeds %d", len(trace.Diagnostics), status.MaxDiagnostics))
	}
	if trace.Outcome != status.OutcomeOK && len(trace.Diagnostics) == 0 {
		return evidenceError(status.OutcomeInvalid, "MISSING_EVIDENCE_DIAGNOSTIC", "diagnostics", "non-ok evidence trace requires a diagnostic")
	}
	for i, diagnostic := range trace.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_DIAGNOSTIC", fmt.Sprintf("diagnostics[%d]", i), err.Error())
		}
	}
	textBytes := len(trace.AlternativeID)
	for i, record := range trace.Records {
		path := fmt.Sprintf("records[%d]", i)
		if err := validateEvidenceRecord(record, path); err != nil {
			return err
		}
		for _, value := range evidenceRecordStrings(record) {
			if len(value) > budget.MaxTextBytes-textBytes {
				return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", fmt.Sprintf("evidence text exceeds %d bytes", budget.MaxTextBytes))
			}
			textBytes += len(value)
		}
		for featureIndex, value := range record.Features {
			if !validEvidenceFeatureToken(string(value.Feature)) || value.State != features.StateKnown || !validEvidenceFeatureToken(string(value.Value)) || value.Variable != nil {
				return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_FEATURE", fmt.Sprintf("records[%d].features[%d]", i, featureIndex), "source feature evidence must contain a bounded known feature value")
			}
			if featureIndex > 0 && record.Features[featureIndex-1].Feature >= value.Feature {
				return evidenceError(status.OutcomeInvalid, "NONCANONICAL_EVIDENCE_FEATURES", fmt.Sprintf("records[%d].features", i), "source features must be unique and sorted by feature ID")
			}
		}
	}
	for _, diagnostic := range trace.Diagnostics {
		for _, value := range []string{diagnostic.Code, diagnostic.Message, diagnostic.Path} {
			if len(value) > budget.MaxTextBytes-textBytes {
				return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", fmt.Sprintf("evidence text exceeds %d bytes", budget.MaxTextBytes))
			}
			textBytes += len(value)
		}
	}
	return nil
}

func validEvidenceFeatureToken(value string) bool {
	if value == "" || len(value) > features.MaxTokenBytes {
		return false
	}
	for index, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && !(index > 0 && character == '_') {
			return false
		}
	}
	return true
}

func validateEvidenceRecord(record EvidenceRecord, path string) error {
	switch record.Origin {
	case "":
		if record.SourceSpan.Start < 0 || record.SourceSpan.End <= record.SourceSpan.Start || record.SourceSpan.End > status.DefaultBounds().MaxInputBytes {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_SPAN", path+".source_span", "source span must be a non-empty half-open byte range within the input limit")
		}
	case EvidenceOriginGenerated:
		if record.SourceSpan != (text.Span{}) || record.Outcome != status.OutcomeOK || record.CandidateID != "" || record.PhraseNodeID != "" || record.ReadingID != "" || len(record.Features) != 0 || record.Role != "" || record.MeaningNodeID != "" || record.PredicateNodeID != "" {
			return evidenceError(status.OutcomeInvalid, "INVALID_GENERATED_EVIDENCE", path, "generated evidence cannot claim a source span, parse identity, source features, or non-success outcome")
		}
	default:
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_ORIGIN", path+".origin", "evidence origin is unsupported")
	}
	if !record.Outcome.Valid() {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_OUTCOME", path+".outcome", "evidence outcome is unsupported")
	}
	if len(record.Features) > features.MaxBundleValues {
		return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", path+".features", "source feature evidence exceeds the feature bundle limit")
	}
	for _, value := range []struct {
		text string
		name string
		max  int
	}{
		{record.CandidateID, "candidate_id", 256},
		{record.NPPhraseID, "np_phrase_id", 256},
		{record.NPPhrasePlanID, "np_phrase_plan_id", 256},
		{string(record.PhraseNodeID), "phrase_node_id", 256},
		{string(record.MeaningNodeID), "meaning_node_id", 256},
		{string(record.PredicateNodeID), "predicate_node_id", 256},
		{record.ReadingID, "reading_id", 256},
		{record.BindingKey, "binding_key", 2*ids.MaxSymbolBytes + 2},
		{record.Reason, "reason", status.MaxDiagnosticMessageBytes},
	} {
		if value.text != "" && !validEvidenceText(value.text, value.max) {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_TEXT", path+"."+value.name, "evidence text is invalid, empty, or exceeds its byte limit")
		}
	}
	if record.PredicateNodeID != "" && (record.MeaningNodeID == "" || record.Role == "") {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_ADDRESS", path, "participant evidence requires a Meaning node, an owning predicate and a role")
	}
	if record.Reason == "" && record.Outcome != status.OutcomeOK {
		return evidenceError(status.OutcomeInvalid, "MISSING_EVIDENCE_REASON", path+".reason", "non-ok evidence record requires a reason")
	}
	if record.SenseID.String() != "" {
		if _, err := ids.ParseSenseID(record.SenseID.String()); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_SENSE_ID", path+".sense_id", "sense ID is malformed")
		}
	}
	if record.CoreConceptID.String() != "" {
		if _, err := ids.ParseConceptID(record.CoreConceptID.String()); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_CONCEPT_ID", path+".core_concept_id", "CORE concept ID is malformed")
		}
		if record.SenseID.String() == "" || record.BindingKey != BindingKeyFor(record.SenseID, record.CoreConceptID) {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_BINDING_KEY", path+".binding_key", "CORE binding key must identify the exact sense and concept pair")
		}
	} else if record.BindingKey != "" {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_BINDING_KEY", path+".binding_key", "binding key requires a CORE concept ID")
	}
	if record.Role != "" && !validEvidenceRole(string(record.Role)) {
		return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_ROLE", path+".role", "semantic role must be a bounded lowercase identifier")
	}
	planPresent := record.PlanID != "" || record.FamilyID.String() != "" || record.VariantID.String() != "" || record.FrameID != ""
	if record.Origin == EvidenceOriginGenerated && (!planPresent || record.CoreConceptID.String() == "") {
		return evidenceError(status.OutcomeInvalid, "INCOMPLETE_GENERATED_EVIDENCE", path, "generated evidence requires an exact CORE binding and construction plan")
	}
	if planPresent {
		if !record.PlanID.Valid() {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_PLAN_ID", path+".plan_id", "plan ID must be canonical")
		}
		if _, err := ids.ParseConstructionID(record.FamilyID.String()); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_FAMILY_ID", path+".family_id", "family ID is malformed")
		}
		if _, err := ids.ParseConstructionID(record.VariantID.String()); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_VARIANT_ID", path+".variant_id", "variant ID is malformed")
		}
		if !validFrameID(string(record.FrameID)) {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_FRAME_ID", path+".frame_id", "frame ID is malformed")
		}
	}
	if hasSourceRef(record.Provenance) {
		if err := record.Provenance.Validate(); err != nil {
			return evidenceError(status.OutcomeInvalid, "INVALID_EVIDENCE_PROVENANCE", path+".provenance", err.Error())
		}
	} else if record.CoreConceptID.String() != "" {
		return evidenceError(status.OutcomeInvalid, "MISSING_EVIDENCE_PROVENANCE", path+".provenance", "CORE binding evidence requires provenance")
	}
	return nil
}

func evidenceRecordStrings(record EvidenceRecord) []string {
	ref := record.Provenance
	result := []string{
		record.CandidateID, string(record.PhraseNodeID), string(record.MeaningNodeID), string(record.PredicateNodeID), record.ReadingID, record.SenseID.String(),
		record.NPPhraseID, record.NPPhrasePlanID,
		record.CoreConceptID.String(), record.BindingKey, string(record.Role), string(record.PlanID),
		record.FamilyID.String(), record.VariantID.String(), string(record.FrameID), record.Reason,
		ref.ID, ref.Path, ref.URI, ref.Version, ref.License, ref.Author, ref.ChecksumSHA256,
	}
	for _, feature := range record.Features {
		result = append(result, string(feature.Feature), string(feature.Value))
	}
	return result
}

func hasSourceRef(ref provenance.SourceRef) bool {
	return ref.ID != "" || ref.Path != "" || ref.URI != "" || ref.Version != "" || ref.License != "" || ref.Author != "" || ref.ChecksumSHA256 != ""
}

func validEvidenceText(value string, maximum int) bool {
	if value == "" || len(value) > maximum || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func validEvidenceRole(value string) bool {
	if !validEvidenceText(value, 256) {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func validFrameID(value string) bool {
	if !validEvidenceText(value, valency.MaxFrameTextBytes) {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func evidenceError(kind status.OutcomeKind, code, path, message string) error {
	return &EvidenceError{result: status.Result{
		Status: kind,
		Diagnostics: []status.Diagnostic{{
			Code: code, Path: path, Message: message, Severity: status.SeverityError,
		}},
	}}
}

// EvidenceError carries a typed validation or resource-limit result.
type EvidenceError struct{ result status.Result }

func (err *EvidenceError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "semantic evidence error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed evidence result.
func (err *EvidenceError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

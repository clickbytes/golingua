package realizer

import (
	"fmt"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	defaultSegmentMaxTextBytes = 64 * 1024
	maxSegmentTextBytes        = 1 << 20
	defaultSegmentMaxParts     = 4096
	maxSegmentParts            = 16384
	maxSegmentSourceIDBytes    = status.MaxTraceIDBytes
)

// SegmentKind classifies a surface slice without claiming that the slice can
// be reverse-parsed or that it has a particular grammatical interpretation.
type SegmentKind string

const (
	SegmentLexical    SegmentKind = "lexical"
	SegmentTerminal   SegmentKind = "terminal"
	SegmentSeparator  SegmentKind = "separator"
	SegmentMorphology SegmentKind = "morphology"
)

// LinearizedPart is one exact UTF-8 string emitted by a surface linearizer.
// Every emitted byte, including spaces and punctuation, must occur in a part.
// The identity fields retain the source trace responsible for that slice.
type LinearizedPart struct {
	Text           string                      `json:"text"`
	Kind           SegmentKind                 `json:"kind"`
	SourceID       string                      `json:"source_id,omitempty"`
	NodeID         string                      `json:"node_id,omitempty"`
	LexemeID       string                      `json:"lexeme_id,omitempty"`
	SenseID        string                      `json:"sense_id,omitempty"`
	ConstructionID string                      `json:"construction_id,omitempty"`
	ProcedureID    string                      `json:"procedure_id,omitempty"`
	MorphTrace     []morph.GenerationTraceStep `json:"morph_trace,omitempty"`
	NPPlanID       string                      `json:"np_plan_id,omitempty"`
	NPRole         string                      `json:"np_role,omitempty"`
}

// Segment is a half-open UTF-8 byte span into the final linearized text.
// Identity fields copy the exact provenance supplied by the corresponding
// LinearizedPart; no identity or meaning is inferred from surface spelling.
type Segment struct {
	Start          int                         `json:"start"`
	End            int                         `json:"end"`
	Kind           SegmentKind                 `json:"kind"`
	SourceID       string                      `json:"source_id,omitempty"`
	NodeID         string                      `json:"node_id,omitempty"`
	LexemeID       string                      `json:"lexeme_id,omitempty"`
	SenseID        string                      `json:"sense_id,omitempty"`
	ConstructionID string                      `json:"construction_id,omitempty"`
	ProcedureID    string                      `json:"procedure_id,omitempty"`
	MorphTrace     []morph.GenerationTraceStep `json:"morph_trace,omitempty"`
	NPPlanID       string                      `json:"np_plan_id,omitempty"`
	NPRole         string                      `json:"np_role,omitempty"`
}

// SegmentBudget bounds final UTF-8 text size and the number of source-traced
// parts retained as spans.
type SegmentBudget struct {
	MaxTextBytes int
	MaxSegments  int
}

// DefaultSegmentBudget returns finite limits for one linearized result.
func DefaultSegmentBudget() SegmentBudget {
	return SegmentBudget{MaxTextBytes: defaultSegmentMaxTextBytes, MaxSegments: defaultSegmentMaxParts}
}

// Validate requires positive limits within package ceilings.
func (budget SegmentBudget) Validate() error {
	if budget.MaxTextBytes <= 0 || budget.MaxTextBytes > maxSegmentTextBytes || budget.MaxSegments <= 0 || budget.MaxSegments > maxSegmentParts {
		return segmentError(status.OutcomeInvalid, "INVALID_SEGMENT_BUDGET", "budget", "segment limits must be positive and within package ceilings")
	}
	return nil
}

// SegmentError carries a typed outcome and bounded diagnostic for invalid
// surface provenance or exhausted segmentation limits.
type SegmentError struct{ result status.Result }

func (err *SegmentError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "surface segment error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed error result.
func (err *SegmentError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

// BuildSegments derives exact half-open UTF-8 byte spans from parts emitted by
// the surface linearizer. It accepts only a complete, byte-for-byte partition
// of text; it never searches the output for a repeated token or reconstructs
// missing provenance.
func BuildSegments(text string, parts []LinearizedPart, budget SegmentBudget) ([]Segment, error) {
	if budget == (SegmentBudget{}) {
		budget = DefaultSegmentBudget()
	}
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	if !utf8.ValidString(text) {
		return nil, segmentError(status.OutcomeInvalid, "INVALID_LINEARIZED_UTF8", "text", "linearized output must be valid UTF-8")
	}
	if len(text) > budget.MaxTextBytes {
		return nil, segmentError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", "linearized output exceeds the segment text-byte limit")
	}
	if len(parts) > budget.MaxSegments {
		return nil, segmentError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.segments", "linearized output exceeds the segment count limit")
	}
	if len(parts) == 0 {
		if text != "" {
			return nil, segmentError(status.OutcomeInvalid, "SEGMENT_COVERAGE_GAP", "parts", "non-empty output requires source-traced parts covering every byte")
		}
		return []Segment{}, nil
	}

	segments := make([]Segment, 0, len(parts))
	position := 0
	for index, part := range parts {
		path := fmt.Sprintf("parts[%d]", index)
		if part.Text == "" || !utf8.ValidString(part.Text) {
			return nil, segmentError(status.OutcomeInvalid, "INVALID_LINEARIZED_PART", path+".text", "each linearized part must contain non-empty valid UTF-8")
		}
		if !validSegmentKind(part.Kind) {
			return nil, segmentError(status.OutcomeInvalid, "INVALID_SEGMENT_KIND", path+".kind", "linearized part kind is not supported")
		}
		if err := validatePartTrace(part); err != nil {
			return nil, segmentError(status.OutcomeInvalid, "INVALID_SEGMENT_TRACE", path, err.Error())
		}
		if len(part.Text) > len(text)-position || text[position:position+len(part.Text)] != part.Text {
			return nil, segmentError(status.OutcomeInvalid, "SEGMENT_TEXT_MISMATCH", path+".text", "parts must match the final output exactly and in order")
		}
		end := position + len(part.Text)
		segments = append(segments, Segment{
			Start: position, End: end, Kind: part.Kind, SourceID: part.SourceID,
			NodeID: part.NodeID, LexemeID: part.LexemeID, SenseID: part.SenseID,
			ConstructionID: part.ConstructionID, ProcedureID: part.ProcedureID,
			MorphTrace: append([]morph.GenerationTraceStep(nil), part.MorphTrace...),
			NPPlanID:   part.NPPlanID, NPRole: part.NPRole,
		})
		position = end
	}
	if position != len(text) {
		return nil, segmentError(status.OutcomeInvalid, "SEGMENT_COVERAGE_GAP", "parts", "source-traced parts do not cover the complete output")
	}
	return segments, nil
}

func validSegmentKind(kind SegmentKind) bool {
	switch kind {
	case SegmentLexical, SegmentTerminal, SegmentSeparator, SegmentMorphology:
		return true
	default:
		return false
	}
}

func validatePartTrace(part LinearizedPart) error {
	values := [...]string{part.SourceID, part.NodeID, part.LexemeID, part.SenseID, part.ConstructionID, part.ProcedureID, part.NPPlanID, part.NPRole}
	hasIdentity := false
	for _, value := range values {
		if value == "" {
			continue
		}
		hasIdentity = true
		if len(value) > maxSegmentSourceIDBytes || !utf8.ValidString(value) {
			return fmt.Errorf("source trace identity is invalid UTF-8 or exceeds %d bytes", maxSegmentSourceIDBytes)
		}
	}
	if !hasIdentity {
		return fmt.Errorf("each part requires an explicit source trace identity")
	}
	return nil
}

func segmentError(kind status.OutcomeKind, code, path, message string) error {
	return &SegmentError{result: status.Result{
		Status: kind,
		Diagnostics: []status.Diagnostic{{
			Code: code, Message: message, Path: path, Severity: status.SeverityError,
		}},
	}}
}

// Package engine contains the bounded request and result contract for text
// translation. It does not assign semantic interpretations itself.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxTranslationAlternatives          = 256
	MaxTranslationIDBytes               = 256
	MaxTranslationTextBytes             = 64 * 1024
	MaxTranslationLosses                = 64
	MaxTranslationAssumptions           = 64
	MaxTranslationAssumptionReasonBytes = 1024
	MaxTranslationEvidence              = 32
	MaxTranslationPayloadBytes          = 4 << 20
	MaxTranslationMeaningNodes          = 65_536
	MaxTranslationEvidenceRows          = 65_536
)

var ErrInvalidTranslationResult = errors.New("invalid translation result")

// TranslationAlternative is one explicitly retained interpretation and its
// target realization. A candidate's outcome remains distinct from the outer
// request outcome.
type TranslationAlternative struct {
	ID                string                     `json:"id"`
	SourceCandidateID string                     `json:"source_candidate_id"`
	SemanticID        string                     `json:"semantic_id"`
	Outcome           status.OutcomeKind         `json:"outcome"`
	Text              string                     `json:"text,omitempty"`
	Meaning           *meaning.Meaning           `json:"meaning,omitempty"`
	Evidence          []semantic.EvidenceTrace   `json:"evidence,omitempty"`
	Assumptions       []realizer.FieldAssumption `json:"assumptions,omitempty"`
	Losses            []realizer.FieldLoss       `json:"losses,omitempty"`
	Diagnostics       []status.Diagnostic        `json:"diagnostics,omitempty"`
}

// TranslationResult is the bounded outward result of one translation request.
// Text is present for a completed result, or for a preview explicitly selected
// by PreviewID while the overall result remains ambiguous or incomplete.
type TranslationResult struct {
	Status      status.OutcomeKind       `json:"status"`
	From        features.LanguageID      `json:"from"`
	To          features.LanguageID      `json:"to"`
	Text        string                   `json:"text,omitempty"`
	PreviewID   string                   `json:"preview_id,omitempty"`
	Candidates  []TranslationAlternative `json:"candidates,omitempty"`
	Diagnostics []status.Diagnostic      `json:"diagnostics,omitempty"`
}

// Clone returns a defensive copy of slices and nested Meaning/evidence data.
func (candidate TranslationAlternative) Clone() TranslationAlternative {
	candidate.Evidence = append([]semantic.EvidenceTrace(nil), candidate.Evidence...)
	for i := range candidate.Evidence {
		candidate.Evidence[i] = candidate.Evidence[i].Clone()
	}
	candidate.Assumptions = cloneFieldAssumptions(candidate.Assumptions)
	candidate.Losses = append([]realizer.FieldLoss(nil), candidate.Losses...)
	for i := range candidate.Losses {
		candidate.Losses[i].Path = append([]meaning.NodeID(nil), candidate.Losses[i].Path...)
	}
	candidate.Diagnostics = append([]status.Diagnostic(nil), candidate.Diagnostics...)
	if candidate.Meaning != nil {
		m := *candidate.Meaning
		m.Roots = append([]meaning.NodeID(nil), m.Roots...)
		m.Nodes = append([]meaning.Node(nil), m.Nodes...)
		for i := range m.Nodes {
			m.Nodes[i].Participants = append([]meaning.Participant(nil), m.Nodes[i].Participants...)
			m.Nodes[i].Content = append([]meaning.NodeID(nil), m.Nodes[i].Content...)
			if m.Nodes[i].Quantifier != nil {
				q := *m.Nodes[i].Quantifier
				if q.Count != nil {
					count := *q.Count
					q.Count = &count
				}
				m.Nodes[i].Quantifier = &q
			}
		}
		candidate.Meaning = &m
	}
	return candidate
}

// Clone returns a defensive copy of all result-owned slices and nested data.
func (result TranslationResult) Clone() TranslationResult {
	result.Candidates = append([]TranslationAlternative(nil), result.Candidates...)
	for i := range result.Candidates {
		result.Candidates[i] = result.Candidates[i].Clone()
	}
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}

// Validate checks the finite public result contract. It validates nested
// evidence and Meaning structure without claiming any semantic truth.
func (result TranslationResult) Validate() error {
	if !result.Status.Valid() {
		return invalid("status", "unknown outcome")
	}
	if result.Status == status.OutcomeInvalid {
		if !validRawLanguage(string(result.From)) || !validRawLanguage(string(result.To)) {
			return invalid("language", "invalid requests may retain only bounded UTF-8 language text")
		}
	} else if !validLanguage(string(result.From)) || !validLanguage(string(result.To)) || result.From == result.To {
		return invalid("language", "from and to must be distinct bounded language identifiers")
	}
	if err := validateText(result.Text, MaxTranslationTextBytes, true); err != nil {
		return invalid("text", err.Error())
	}
	if len(result.Candidates) > MaxTranslationAlternatives {
		return invalid("candidates", fmt.Sprintf("candidate count exceeds %d", MaxTranslationAlternatives))
	}
	if len(result.Diagnostics) > status.MaxDiagnostics {
		return invalid("diagnostics", "diagnostic count exceeds limit")
	}
	for i, d := range result.Diagnostics {
		if err := d.Validate(); err != nil {
			return invalid(fmt.Sprintf("diagnostics[%d]", i), err.Error())
		}
	}
	if result.Status != status.OutcomeOK && len(result.Diagnostics) == 0 {
		return invalid("diagnostics", "non-ok result requires an explanatory diagnostic")
	}
	seen := make(map[string]struct{}, len(result.Candidates))
	var selected *TranslationAlternative
	var payloadBytes, meaningNodes, evidenceRows int
	payloadBytes += len(result.Text) + resultTextBytes(result.Diagnostics)
	for i := range result.Candidates {
		candidate := &result.Candidates[i]
		path := fmt.Sprintf("candidates[%d]", i)
		if !validID(candidate.ID) || !validID(candidate.SourceCandidateID) || !validID(candidate.SemanticID) {
			return invalid(path, "candidate identifiers must be non-empty bounded UTF-8 identifiers")
		}
		if _, ok := seen[candidate.ID]; ok {
			return invalid(path+".id", "candidate IDs must be unique")
		}
		seen[candidate.ID] = struct{}{}
		if !candidate.Outcome.Valid() {
			return invalid(path+".outcome", "unknown candidate outcome")
		}
		if err := validateText(candidate.Text, MaxTranslationTextBytes, true); err != nil {
			return invalid(path+".text", err.Error())
		}
		payloadBytes += len(candidate.ID) + len(candidate.SourceCandidateID) + len(candidate.SemanticID) + len(candidate.Text)
		if len(candidate.Evidence) > MaxTranslationEvidence || len(candidate.Assumptions) > MaxTranslationAssumptions || len(candidate.Losses) > MaxTranslationLosses || len(candidate.Diagnostics) > status.MaxDiagnostics {
			return invalid(path, "nested evidence, assumptions, losses, or diagnostics exceed limits")
		}
		for j, trace := range candidate.Evidence {
			if err := trace.Validate(); err != nil {
				return invalid(fmt.Sprintf("%s.evidence[%d]", path, j), err.Error())
			}
			evidenceRows += len(trace.Records)
			payloadBytes += evidenceTextBytes(trace)
		}
		for j, assumption := range candidate.Assumptions {
			if err := validateAssumption(assumption); err != nil {
				return invalid(fmt.Sprintf("%s.assumptions[%d]", path, j), err.Error())
			}
			payloadBytes += len(assumption.Field) + len(assumption.Value) + len(assumption.Reason)
			for _, id := range assumption.Path {
				payloadBytes += len(string(id))
			}
		}
		for j, loss := range candidate.Losses {
			if err := validateLoss(loss); err != nil {
				return invalid(fmt.Sprintf("%s.losses[%d]", path, j), err.Error())
			}
			if candidate.Outcome == status.OutcomeOK && loss.Outcome != status.OutcomeOK {
				return invalid(fmt.Sprintf("%s.losses[%d].outcome", path, j), "a complete candidate cannot contain a non-ok loss")
			}
			payloadBytes += len(loss.Field) + len(loss.Reason)
		}
		for j, diagnostic := range candidate.Diagnostics {
			if err := diagnostic.Validate(); err != nil {
				return invalid(fmt.Sprintf("%s.diagnostics[%d]", path, j), err.Error())
			}
			payloadBytes += diagnosticTextBytes(diagnostic)
		}
		if candidate.Outcome != status.OutcomeOK && len(candidate.Diagnostics) == 0 {
			return invalid(path+".diagnostics", "non-ok candidate requires an explanatory diagnostic")
		}
		if candidate.Outcome == status.OutcomeOK && candidate.Text == "" {
			return invalid(path+".text", "a complete candidate requires realized text")
		}
		if candidate.Meaning != nil {
			if err := meaning.ValidateStructure(*candidate.Meaning, meaning.DefaultBudget()); err != nil {
				return invalid(path+".meaning", err.Error())
			}
			meaningNodes += len(candidate.Meaning.Nodes)
			payloadBytes += meaningTextBytes(*candidate.Meaning)
		}
		if candidate.ID == result.PreviewID {
			selected = candidate
		}
	}
	if payloadBytes > MaxTranslationPayloadBytes || meaningNodes > MaxTranslationMeaningNodes || evidenceRows > MaxTranslationEvidenceRows {
		return invalid("payload", "aggregate Meaning, evidence, assumptions, and text payload exceeds the translation result limits")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return invalid("payload", "translation result cannot be encoded as JSON")
	}
	if len(encoded) > MaxTranslationPayloadBytes {
		return invalid("payload", "serialized translation result exceeds the translation result byte limit")
	}
	if result.Status == status.OutcomeOK {
		if result.Text == "" || result.PreviewID != "" || len(result.Candidates) != 1 || result.Candidates[0].Outcome != status.OutcomeOK || result.Candidates[0].Text != result.Text {
			return invalid("text", "ok result requires exactly one complete candidate whose text matches the result")
		}
	} else {
		if result.PreviewID == "" {
			if result.Text != "" {
				return invalid("text", "non-ok result text requires an explicitly selected preview")
			}
		} else {
			if result.Status != status.OutcomeAmbiguous && result.Status != status.OutcomeIncomplete {
				return invalid("preview_id", "preview is allowed only for ambiguous or incomplete results")
			}
			if !validID(result.PreviewID) || selected == nil || selected.Outcome != status.OutcomeOK || selected.Text == "" || result.Text != selected.Text {
				return invalid("preview_id", "preview must identify a candidate whose text matches the result")
			}
		}
	}
	return nil
}

func validRawLanguage(value string) bool {
	if len(value) > 35 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func evidenceTextBytes(trace semantic.EvidenceTrace) int {
	total := len(trace.AlternativeID) + resultTextBytes(trace.Diagnostics)
	for _, record := range trace.Records {
		total += len(string(record.CandidateID)) + len(string(record.PhraseNodeID)) + len(record.ReadingID) +
			len(record.SenseID.String()) + len(record.CoreConceptID.String()) + len(record.BindingKey) +
			len(string(record.Role)) + len(record.PlanID) + len(record.FamilyID.String()) +
			len(string(record.VariantID)) + len(string(record.FrameID)) + len(record.Reason)
		total += len(record.Provenance.ID) + len(record.Provenance.Path) + len(record.Provenance.URI) +
			len(record.Provenance.Version) + len(record.Provenance.License) + len(record.Provenance.Author) + len(record.Provenance.ChecksumSHA256)
	}
	return total
}

func resultTextBytes(diagnostics []status.Diagnostic) int {
	total := 0
	for _, diagnostic := range diagnostics {
		total += diagnosticTextBytes(diagnostic)
	}
	return total
}

func diagnosticTextBytes(diagnostic status.Diagnostic) int {
	return len(diagnostic.Code) + len(diagnostic.Path) + len(diagnostic.Message)
}

func meaningTextBytes(graph meaning.Meaning) int {
	total := 0
	for _, root := range graph.Roots {
		total += len(string(root))
	}
	for _, node := range graph.Nodes {
		total += len(string(node.ID)) + len(node.Kind) + len(string(node.Concept)) + len(string(node.Target)) + len(node.Operator) + len(node.Modality) + len(node.Act)
		for _, participant := range node.Participants {
			total += len(string(participant.Role)) + len(string(participant.Target))
		}
		for _, child := range node.Content {
			total += len(string(child))
		}
	}
	return total
}

func validateLoss(loss realizer.FieldLoss) error {
	if len(loss.Path) == 0 || len(loss.Path) > meaning.MaxMeaningDepth || !validID(loss.Field) || !loss.Outcome.Valid() || !validID(loss.Reason) {
		return errors.New("loss requires a bounded path, field, valid outcome, and reason")
	}
	for _, id := range loss.Path {
		if !validID(string(id)) {
			return errors.New("loss path contains an invalid node ID")
		}
	}
	return nil
}

func validateAssumption(assumption realizer.FieldAssumption) error {
	if len(assumption.Path) == 0 || len(assumption.Path) > meaning.MaxMeaningDepth || !validID(assumption.Field) || !validID(assumption.Value) {
		return errors.New("assumption requires a bounded path, field, and value")
	}
	if assumption.Reason == "" || len(assumption.Reason) > MaxTranslationAssumptionReasonBytes || !utf8.ValidString(assumption.Reason) || strings.TrimSpace(assumption.Reason) != assumption.Reason {
		return errors.New("assumption requires a bounded non-empty reason")
	}
	for _, r := range assumption.Reason {
		if r < 0x20 || r == 0x7f {
			return errors.New("assumption reason contains a control character")
		}
	}
	for _, id := range assumption.Path {
		if !validID(string(id)) {
			return errors.New("assumption path contains an invalid node ID")
		}
	}
	return nil
}

func cloneFieldAssumptions(values []realizer.FieldAssumption) []realizer.FieldAssumption {
	result := append([]realizer.FieldAssumption(nil), values...)
	for index := range result {
		result[index].Path = append([]meaning.NodeID(nil), result[index].Path...)
	}
	return result
}

func validLanguage(value string) bool {
	if len(value) < 2 || len(value) > 35 || strings.ToLower(value) != value || !utf8.ValidString(value) {
		return false
	}
	parts := strings.Split(value, "-")
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z' || index > 0 && r >= '0' && r <= '9') {
				return false
			}
		}
	}
	return len(parts[0]) >= 2
}

func validID(value string) bool {
	if value == "" || len(value) > MaxTranslationIDBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func validateText(value string, maximum int, allowEmpty bool) error {
	if (!allowEmpty && value == "") || len(value) > maximum || !utf8.ValidString(value) {
		return errors.New("text is empty, invalid UTF-8, or exceeds its byte limit")
	}
	return nil
}

func invalid(path, message string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidTranslationResult, path, message)
}

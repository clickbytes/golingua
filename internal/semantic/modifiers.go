package semantic

import (
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
)

type modifierComposition struct {
	state meaning.Node
}

func familyModifier(family construction.ConstructionFamily) (construction.SlotSpec, bool) {
	for _, slot := range family.SlotSpecs {
		if slot.ModifierBinding != nil {
			return slot, true
		}
	}
	return construction.SlotSpec{}, false
}

func modifierCompositionFor(
	input CompositionInput,
	alternative parser.ParseAlternative,
	family construction.ConstructionFamily,
	candidates map[string]parser.PhraseCandidate,
	phraseNodes map[string]parser.NodeID,
	referentNodeByCandidate map[string]meaning.NodeID,
	meaningNodes []meaning.Node,
	attributes map[string]NodeAttributesEvidence,
) (modifierComposition, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	modifierSlot, declared := familyModifier(family)
	if !declared {
		return modifierComposition{}, status.OutcomeOK, nil, status.Diagnostic{}
	}
	modifierBinding := modifierSlot.ModifierBinding

	modifierSlotBinding, found, duplicate := alternativeSlotBinding(alternative, modifierSlot.ID)
	if duplicate {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_SLOT_AMBIGUOUS", "parse alternative binds the modifier slot more than once", "modifier.slot")
	}
	if !found || len(modifierSlotBinding.CandidateIDs) == 0 {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_CANDIDATE_UNRESOLVED", "required modifier slot has no selected AP candidate", "modifier.slot")
	}
	if len(modifierSlotBinding.CandidateIDs) != 1 {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_CANDIDATE_AMBIGUOUS", "modifier slot has multiple AP candidates", "modifier.slot")
	}
	modifierCandidate, exists := candidates[modifierSlotBinding.CandidateIDs[0]]
	if !exists || modifierCandidate.Category != construction.CategoryAP {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_CANDIDATE_UNSUPPORTED", "modifier slot does not resolve to one AP candidate", "modifier.slot")
	}

	targetSlotBinding, found, duplicate := alternativeSlotBinding(alternative, modifierBinding.TargetSlot)
	if duplicate {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_TARGET_AMBIGUOUS", "modifier target slot is bound more than once", "modifier.target")
	}
	if !found || len(targetSlotBinding.CandidateIDs) == 0 {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_TARGET_UNRESOLVED", "modifier target slot has no selected NP candidate", "modifier.target")
	}
	if len(targetSlotBinding.CandidateIDs) != 1 {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_TARGET_AMBIGUOUS", "modifier target slot has multiple NP candidates", "modifier.target")
	}
	targetCandidate, exists := candidates[targetSlotBinding.CandidateIDs[0]]
	if !exists || targetCandidate.Category != construction.CategoryNP {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target slot does not resolve to one NP candidate", "modifier.target")
	}
	targetNodeID, exists := referentNodeByCandidate[targetCandidate.ID]
	if !exists {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_TARGET_UNRESOLVED", "selected NP candidate was not materialized as a referent", "modifier.target")
	}
	targetNode, exists := meaningNodeByID(meaningNodes, targetNodeID)
	if !exists || targetNode.Kind != meaning.KindReferent {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target is not the exact materialized referent for its NP candidate", "modifier.target")
	}
	targetConceptID, err := ids.ParseConceptID(string(targetNode.Concept))
	if err != nil {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target referent has no valid CORE concept", "modifier.target")
	}
	targetConcept, exists := input.Core.Lookup(targetConceptID)
	if !exists {
		return modifierCompositionFailure(status.OutcomeUnknown, "MODIFIER_TARGET_UNKNOWN", "modifier target concept is absent from CORE", "modifier.target")
	}

	mapping, outcome, records, reason := uniqueModifierTransfer(input.Lexicon, modifierCandidate, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID])
	if outcome != status.OutcomeOK {
		code := "MODIFIER_TRANSFER_UNAVAILABLE"
		if outcome == status.OutcomeAmbiguous {
			code = "MODIFIER_TRANSFER_AMBIGUOUS"
		} else if outcome == status.OutcomeIncomplete {
			code = "MODIFIER_TRANSFER_UNRESOLVED"
		}
		return modifierComposition{}, outcome, records, semanticDiagnostic(code, reason, "modifier.transfer")
	}
	property, exists := input.Core.Lookup(mapping.binding.CoreRef)
	if !exists {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnknown, "approved modifier binding points to a property absent from CORE")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnknown, "MODIFIER_PROPERTY_UNKNOWN", record.Reason, "modifier.transfer", []EvidenceRecord{record})
	}
	if property.Kind != core.KindProperty {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "modifier binding does not point to a CORE property")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_PROPERTY_UNSUPPORTED", record.Reason, "modifier.transfer", []EvidenceRecord{record})
	}
	if !propertyRoleAllowsTarget(property, modifierBinding.SemanticRole, targetConcept.Kind) {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "CORE property role does not license the selected target concept kind")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_ROLE_TARGET_UNSUPPORTED", record.Reason, "modifier.target", []EvidenceRecord{record})
	}
	state, err := meaning.NewState("modifier0", mapping.binding.CoreRef, []meaning.Participant{{Role: modifierBinding.SemanticRole, Target: targetNodeID}})
	if err != nil {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "modifier state cannot be represented by Meaning")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_STATE_UNSUPPORTED", record.Reason, "modifier", []EvidenceRecord{record})
	}
	if attribute, supplied := attributes[modifierCandidate.ID]; supplied {
		state.TemporalFrame = attribute.TemporalFrame
	}
	record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeOK, "")
	return modifierComposition{state: state}, status.OutcomeOK, []EvidenceRecord{record}, status.Diagnostic{}
}

func uniqueModifierTransfer(index *lexicon.Index, candidate parser.PhraseCandidate, role meaning.Role, phraseNode parser.NodeID) (transfer, status.OutcomeKind, []EvidenceRecord, string) {
	if index == nil || len(candidate.SenseIDs) == 0 {
		return transfer{}, status.OutcomeIncomplete, nil, "AP candidate has no lexical senses to bind as a modifier"
	}
	seenSense := make(map[ids.SenseID]struct{}, len(candidate.SenseIDs))
	selected := []transfer{}
	records := []EvidenceRecord{}
	problem := status.OutcomeOK
	reason := ""
	for _, sense := range candidate.SenseIDs {
		if _, seen := seenSense[sense]; seen {
			continue
		}
		seenSense[sense] = struct{}{}
		all := index.BindingsForSense(sense)
		authorized := index.TransferBindings(sense)
		if len(all) == 0 {
			problem = mergeSemanticOutcome(problem, status.OutcomeUnsupported)
			reason = "an AP candidate sense has no semantic transfer binding"
			records = append(records, modifierEvidenceRecord(candidate, transfer{sense: sense}, role, phraseNode, status.OutcomeUnsupported, reason))
			continue
		}
		if len(authorized) != 1 || len(all) != 1 {
			if len(authorized) > 1 || len(all) > 1 && len(authorized) > 0 {
				problem = mergeSemanticOutcome(problem, status.OutcomeAmbiguous)
				reason = "AP sense has competing approved or unapproved CORE mappings"
			} else if problem == status.OutcomeOK || problem == status.OutcomeUnsupported {
				problem = mergeSemanticOutcome(problem, status.OutcomeIncomplete)
				reason = "AP sense has no approved equivalent CORE binding"
			}
			for _, binding := range all {
				outcome := status.OutcomeIncomplete
				bindingReason := "AP sense mapping is not an approved equivalent transfer"
				if binding.TransferAuthorized() {
					outcome = status.OutcomeAmbiguous
					bindingReason = "AP sense has competing modifier mappings"
				}
				records = append(records, modifierEvidenceRecord(candidate, transfer{binding: binding, sense: sense}, role, phraseNode, outcome, bindingReason))
			}
			continue
		}
		mapping := transfer{binding: authorized[0], sense: sense}
		selected = append(selected, mapping)
		records = append(records, modifierEvidenceRecord(candidate, mapping, role, phraseNode, status.OutcomeOK, ""))
	}
	if problem != status.OutcomeOK {
		if problem == status.OutcomeAmbiguous || len(selected) > 0 && problem != status.OutcomeOK {
			problem = status.OutcomeAmbiguous
		}
		for index := range records {
			if records[index].Outcome == status.OutcomeOK {
				records[index].Outcome = problem
				records[index].Reason = "modifier sense competes with an unresolved or unsupported candidate sense"
			}
		}
		return transfer{}, problem, records, reason
	}
	if len(selected) != 1 {
		if len(selected) == 0 {
			return transfer{}, status.OutcomeUnsupported, records, "AP candidate has no unique approved equivalent CORE mapping"
		}
		for index := range records {
			records[index].Outcome = status.OutcomeAmbiguous
			records[index].Reason = "AP candidate has multiple lexical senses with modifier mappings"
		}
		return transfer{}, status.OutcomeAmbiguous, records, "AP candidate has multiple lexical senses with modifier mappings"
	}
	return selected[0], status.OutcomeOK, records, ""
}

func modifierEvidenceRecord(candidate parser.PhraseCandidate, mapping transfer, role meaning.Role, phraseNode parser.NodeID, outcome status.OutcomeKind, reason string) EvidenceRecord {
	record := EvidenceRecord{
		SourceSpan: candidate.Span, CandidateID: candidate.ID, PhraseNodeID: phraseNode,
		ReadingID: firstString(candidate.ReadingIDs), SenseID: mapping.sense,
		Role: role, Outcome: outcome, Reason: reason,
		Features: evidenceFeatures(candidate.Features),
	}
	if mapping.binding.CoreRef.String() != "" {
		record.CoreConceptID = mapping.binding.CoreRef
		record.BindingKey = BindingKeyFor(mapping.sense, mapping.binding.CoreRef)
		record.Provenance = mapping.binding.Provenance
	}
	return record
}

func modifierCompositionFailure(outcome status.OutcomeKind, code, message, path string) (modifierComposition, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	return modifierComposition{}, outcome, nil, semanticDiagnostic(code, message, path)
}

func modifierCompositionFailureWithEvidence(outcome status.OutcomeKind, code, message, path string, records []EvidenceRecord) (modifierComposition, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	return modifierComposition{}, outcome, records, semanticDiagnostic(code, message, path)
}

func alternativeSlotBinding(alternative parser.ParseAlternative, slotID string) (parser.SlotBinding, bool, bool) {
	var selected parser.SlotBinding
	found := false
	for _, slot := range alternative.Slots {
		if slot.SlotID != slotID {
			continue
		}
		if found {
			return parser.SlotBinding{}, false, true
		}
		selected, found = slot, true
	}
	return selected, found, false
}

func propertyRoleAllowsTarget(property core.Concept, role meaning.Role, targetKind core.ConceptKind) bool {
	for _, propertyRole := range property.RoleTypes {
		if propertyRole.Name != string(role) {
			continue
		}
		if len(propertyRole.AllowedKinds) == 0 {
			return true
		}
		for _, allowed := range propertyRole.AllowedKinds {
			if allowed == targetKind {
				return true
			}
		}
		return false
	}
	return false
}

func meaningNodeByID(nodes []meaning.Node, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

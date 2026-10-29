package realizer

import (
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

// FindModifier resolves one declared modifier state through its exact NP slot
// and the selected main predicate's role participant. Concept equality never
// substitutes for a graph-node link.
func FindModifier(graph meaning.Meaning, predicateID meaning.NodeID, family construction.ConstructionFamily, slotID string) (meaning.Node, bool) {
	node, outcome := FindModifierOutcome(graph, predicateID, family, slotID)
	return node, outcome == status.OutcomeOK
}

// FindModifierOutcome resolves one modifier slot without using node IDs or
// graph order as a selection rule. Missing links remain incomplete; competing
// states with the same semantic role and exact referent remain ambiguous.
func FindModifierOutcome(graph meaning.Meaning, predicateID meaning.NodeID, family construction.ConstructionFamily, slotID string) (meaning.Node, status.OutcomeKind) {
	predicate, exists := meaningNodeByID(graph, predicateID)
	if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return meaning.Node{}, status.OutcomeIncomplete
	}

	var modifierSlot, targetSlot construction.SlotSpec
	modifierCount, targetCount := 0, 0
	for _, slot := range family.SlotSpecs {
		if slot.ID == slotID && slot.ModifierBinding != nil {
			modifierSlot = slot
			modifierCount++
		}
		if slot.ID == familyTargetSlot(family, slotID) {
			targetSlot = slot
			targetCount++
		}
	}
	if modifierCount != 1 || targetCount != 1 || modifierSlot.ModifierBinding == nil || targetSlot.Category != construction.CategoryNP || targetSlot.ArgumentBinding == nil || targetSlot.Optionality != construction.SlotRequired || targetSlot.Cardinality.Minimum != 1 || targetSlot.Cardinality.Maximum != 1 || targetSlot.ID == modifierSlot.ID ||
		modifierSlot.Category != construction.CategoryAP || modifierSlot.Optionality != construction.SlotRequired || modifierSlot.Cardinality.Minimum != 1 || modifierSlot.Cardinality.Maximum != 1 {
		return meaning.Node{}, status.OutcomeUnsupported
	}
	targetParticipant, participantCount := participantByRole(predicate, targetSlot.ArgumentBinding.SemanticRole)
	if participantCount == 0 {
		return meaning.Node{}, status.OutcomeIncomplete
	}
	if participantCount > 1 {
		return meaning.Node{}, status.OutcomeAmbiguous
	}
	target, exists := meaningNodeByID(graph, targetParticipant.Target)
	if !exists || target.Kind != meaning.KindReferent {
		return meaning.Node{}, status.OutcomeUnsupported
	}

	stateCount := 0
	var selected meaning.Node
	for _, rootID := range graph.Roots {
		if rootID == predicateID {
			continue
		}
		root, found := meaningNodeByID(graph, rootID)
		if !found || root.Kind != meaning.KindState || len(root.Participants) != 1 {
			continue
		}
		participant := root.Participants[0]
		if participant.Role == modifierSlot.ModifierBinding.SemanticRole && participant.Target == targetParticipant.Target {
			selected = root
			stateCount++
		}
	}
	if stateCount == 0 {
		return meaning.Node{}, status.OutcomeIncomplete
	}
	if stateCount > 1 {
		return meaning.Node{}, status.OutcomeAmbiguous
	}
	return selected, status.OutcomeOK
}

// CoversLexicalMeaning reports whether a selected construction covers every
// lexical node in the Meaning graph through its predicate, arguments, and
// declared modifier. Operator and discourse wrappers remain the caller's
// responsibility.
func CoversLexicalMeaning(graph meaning.Meaning, predicateID meaning.NodeID, family construction.ConstructionFamily) bool {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	lexical := make(map[meaning.NodeID]struct{})
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
		if isLexicalMeaningNode(node.Kind) {
			lexical[node.ID] = struct{}{}
		}
	}
	selected, exists := nodes[predicateID]
	if !exists {
		return false
	}
	if selected.Kind == meaning.KindReferent {
		return len(lexical) == 1
	}
	if selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState {
		return false
	}

	mainPredicates := 0
	boundRoles := make(map[meaning.Role]struct{}, len(family.SlotSpecs))
	modifierSlots := make([]construction.SlotSpec, 0, 1)
	for _, slot := range family.SlotSpecs {
		if slot.IsMainPredicate() {
			mainPredicates++
		}
		if slot.ArgumentBinding != nil {
			boundRoles[slot.ArgumentBinding.SemanticRole] = struct{}{}
		}
		if slot.ModifierBinding != nil {
			modifierSlots = append(modifierSlots, slot)
		}
	}
	if mainPredicates != 1 {
		return false
	}

	covered := map[meaning.NodeID]struct{}{predicateID: {}}
	seenRoles := make(map[meaning.Role]struct{}, len(selected.Participants))
	for _, participant := range selected.Participants {
		if _, duplicate := seenRoles[participant.Role]; duplicate {
			return false
		}
		seenRoles[participant.Role] = struct{}{}
		if _, bound := boundRoles[participant.Role]; !bound {
			return false
		}
		target, found := nodes[participant.Target]
		if !found || target.Kind != meaning.KindReferent {
			return false
		}
		covered[target.ID] = struct{}{}
	}
	sort.Slice(modifierSlots, func(i, j int) bool { return modifierSlots[i].Position < modifierSlots[j].Position })
	selectedModifierIDs := make(map[meaning.NodeID]struct{}, len(modifierSlots))
	for _, modifierSlot := range modifierSlots {
		modifier, outcome := FindModifierOutcome(graph, predicateID, family, modifierSlot.ID)
		if outcome != status.OutcomeOK {
			return false
		}
		if _, duplicate := selectedModifierIDs[modifier.ID]; duplicate {
			return false
		}
		selectedModifierIDs[modifier.ID] = struct{}{}
		covered[modifier.ID] = struct{}{}
		if len(modifier.Participants) != 1 || modifier.Participants[0].Target == "" {
			return false
		}
		if target, found := nodes[modifier.Participants[0].Target]; !found || target.Kind != meaning.KindReferent {
			return false
		} else {
			covered[target.ID] = struct{}{}
		}
	}
	if len(covered) != len(lexical) {
		return false
	}
	for id := range lexical {
		if _, found := covered[id]; !found {
			return false
		}
	}
	return true
}

func familyTargetSlot(family construction.ConstructionFamily, modifierSlotID string) string {
	for _, slot := range family.SlotSpecs {
		if slot.ID == modifierSlotID && slot.ModifierBinding != nil {
			return slot.ModifierBinding.TargetSlot
		}
	}
	return ""
}

func isLexicalMeaningNode(kind meaning.NodeKind) bool {
	return kind == meaning.KindEvent || kind == meaning.KindState || kind == meaning.KindReferent
}

func meaningNodeByID(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

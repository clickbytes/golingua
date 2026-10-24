package realizer

import (
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
)

// FindModifier resolves one declared modifier state through its exact NP slot
// and the selected main predicate's role participant. Concept equality never
// substitutes for a graph-node link.
func FindModifier(graph meaning.Meaning, predicateID meaning.NodeID, family construction.ConstructionFamily, slotID string) (meaning.Node, bool) {
	predicate, exists := meaningNodeByID(graph, predicateID)
	if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return meaning.Node{}, false
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
	if modifierCount != 1 || targetCount != 1 || targetSlot.Category != construction.CategoryNP || targetSlot.ArgumentBinding == nil || targetSlot.ID == modifierSlot.ID {
		return meaning.Node{}, false
	}
	targetParticipant, participantCount := participantByRole(predicate, targetSlot.ArgumentBinding.SemanticRole)
	if participantCount != 1 {
		return meaning.Node{}, false
	}
	target, exists := meaningNodeByID(graph, targetParticipant.Target)
	if !exists || target.Kind != meaning.KindReferent {
		return meaning.Node{}, false
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
	if stateCount != 1 {
		return meaning.Node{}, false
	}
	return selected, true
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
	modifiers := 0
	modifierSlotID := ""
	for _, slot := range family.SlotSpecs {
		if slot.IsMainPredicate() {
			mainPredicates++
		}
		if slot.ArgumentBinding != nil {
			boundRoles[slot.ArgumentBinding.SemanticRole] = struct{}{}
		}
		if slot.ModifierBinding != nil {
			modifiers++
			modifierSlotID = slot.ID
		}
	}
	if mainPredicates != 1 || modifiers > 1 {
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
	if modifiers == 1 {
		modifier, found := FindModifier(graph, predicateID, family, modifierSlotID)
		if !found {
			return false
		}
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

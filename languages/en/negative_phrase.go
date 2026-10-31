package en

import (
	"reflect"
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/realizer"
)

// englishNegativePhraseRecipe opts an explicitly licensed family into local
// NP determiners and VP do-support. The sentence recipe contains only its
// declared slots, adpositions and punctuation, never duplicated function words.
func englishNegativePhraseRecipe(family construction.ConstructionFamily, variant construction.Variant) bool {
	if family.Language != "en" || !familyHasCapability(family, "vp-auxiliary-negation") ||
		!familyHasCapability(family, "negation-negated") || !familyHasOperator(family, "negate") ||
		variant.FamilyID != family.ID || variant.DirectionPolicy != construction.DirectionRealize ||
		!variantHasRequiredSlots(family, variant) {
		return false
	}
	main, ok := englishFamilyMainPredicateSlot(family)
	if !ok || !sequenceHasSlot(variant.Sequence, main.ID) {
		return false
	}
	seen := map[string]bool{}
	for i, token := range variant.Sequence {
		switch token.Kind {
		case construction.TokenSlot:
			if seen[token.SlotID] || token.Optional {
				return false
			}
			seen[token.SlotID] = true
		case construction.TokenTerminal:
			if token.Optional || len(token.AcceptedInputs) != 0 {
				return false
			}
			switch strings.ToLower(token.Terminal) {
			case ".":
				if i != len(variant.Sequence)-1 {
					return false
				}
			case "to": // The approved transfer variant retains this adposition.
				if i+1 >= len(variant.Sequence) || variant.Sequence[i+1].Kind != construction.TokenSlot {
					return false
				}
			default:
				return false
			}
		default:
			return false
		}
	}
	return true
}

func englishVPAuxiliaryVariant(candidate realizer.TargetCandidate) bool {
	return englishNegativePhraseRecipe(candidate.Family.Family, candidate.CompiledVariant.Variant)
}

// The frame, not a fixed subject/object list, owns the semantic arguments.
// Preserve family order for deterministic choices without changing role links.
func englishNegativeNPSlots(candidate realizer.TargetCandidate) []string {
	roles, ok := englishRequiredFrameRoleSlots(candidate)
	if !ok || len(roles) < 2 || len(roles) > 3 {
		return nil
	}
	var slots []string
	for _, spec := range candidate.Family.Family.SlotSpecs {
		for _, slot := range roles {
			if slot == spec.ID {
				slots = append(slots, slot)
			}
		}
	}
	if len(slots) != len(roles) {
		return nil
	}
	return slots
}

// A compact plan supersedes only the same argument/modifier/adposition order.
// A different construction is not discarded merely because its text is similar.
func englishNegativeSameSurfaceOrder(legacy, compact realizer.TargetCandidate) bool {
	if legacy.Family.Family.ID != compact.Family.Family.ID || legacy.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze ||
		!sequenceHasTerminal(legacy.CompiledVariant.Variant.Sequence, "not") {
		return false
	}
	strip := func(variant construction.Variant) []string {
		var parts []string
		for _, token := range variant.Sequence {
			if token.Kind == construction.TokenSlot {
				parts = append(parts, "slot:"+token.SlotID)
				continue
			}
			switch strings.ToLower(token.Terminal) {
			case "a", "an", "the", "do", "does", "did", "not":
			default:
				parts = append(parts, "terminal:"+token.Terminal)
			}
		}
		return parts
	}
	return reflect.DeepEqual(strip(legacy.CompiledVariant.Variant), strip(compact.CompiledVariant.Variant))
}

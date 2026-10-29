package en

import (
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/realizer"
)

func englishDeterminersForCandidate(candidate realizer.TargetCandidate) (englishDeterminers, bool) {
	familyID := candidate.Family.Family.ID.String()
	switch familyID {
	case "construction:en.transitive", "construction:en.transitive_past", "construction:en.transitive_personal", "construction:en.transitive_negative":
		result := englishDeterminers{}
		found := false
		for _, slot := range candidate.Family.Family.SlotSpecs {
			if slot.ID != "subject" && slot.ID != "object" {
				continue
			}
			form := "bare"
			for index, token := range candidate.CompiledVariant.Variant.Sequence {
				if token.Kind != construction.TokenSlot || token.SlotID != slot.ID {
					continue
				}
				if index > 0 {
					previous := candidate.CompiledVariant.Variant.Sequence[index-1]
					if previous.Kind == construction.TokenTerminal && isEnglishDeterminerToken(previous.Terminal) {
						form = previous.Terminal
					}
				}
				break
			}
			if slot.ID == "subject" {
				result.subject = form
			} else {
				result.object = form
			}
			found = true
		}
		return result, found
	default:
		return englishDeterminerVariant(candidate.CompiledVariant.Variant.ID.String())
	}
}

func isEnglishDeterminerToken(value string) bool {
	return value == "the" || value == "a" || value == "an" || value == "this" || value == "that" || value == "these" || value == "those"
}

func englishVariantRequiresNPNumber(candidate realizer.TargetCandidate, slotID, number string) bool {
	for _, condition := range candidate.CompiledVariant.Variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == slotID &&
			(condition.Feature == "noun_number" || condition.Feature == "number") {
			if string(condition.Value) != number {
				return false
			}
		}
	}
	return true
}

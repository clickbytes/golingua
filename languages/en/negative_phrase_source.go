package en

import (
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/parser"
)

// englishParsedNegativeAgreement checks the consumed subject and auxiliary,
// not the canonical spelling of an accepted-input template. A family/frame
// license is checked by the caller; this check never creates a new license.
func englishParsedNegativeAgreement(lattice analyzer.LexicalLattice, alternative parser.ParseAlternative, family construction.ConstructionFamily, variant construction.Variant, byID map[string]parser.PhraseCandidate) (parser.PhraseCandidate, bool) {
	var empty parser.PhraseCandidate
	if variant.DirectionPolicy == construction.DirectionRealize || variant.FamilyID != family.ID ||
		!familyHasOperator(family, "negate") || !variantHasRequiredSlots(family, variant) ||
		!englishAlternativeUnresolvedSlotsAreBound(alternative, family) ||
		!englishAlternativeHasRequiredSlots(alternative, family, byID) {
		return empty, false
	}
	main, found := englishFamilyMainPredicateSlot(family)
	if !found {
		return empty, false
	}
	var subject, predicate parser.PhraseCandidate
	for _, slot := range alternative.Slots {
		if slot.SlotID != "subject" && slot.SlotID != main.ID {
			continue
		}
		if len(slot.CandidateIDs) != 1 || slot.WHGap != nil {
			return empty, false
		}
		candidate, exists := byID[slot.CandidateIDs[0]]
		if !exists {
			return empty, false
		}
		if slot.SlotID == "subject" {
			subject = candidate
		} else {
			predicate = candidate
		}
	}
	if subject.Category != construction.CategoryNP || predicate.Category != construction.CategoryVP ||
		predicate.Features["verb_form"] != "base" || predicate.TokenStart < 2 || predicate.TokenStart >= len(lattice.Tokens) ||
		subject.TokenEnd != predicate.TokenStart-2 || lattice.Tokens[predicate.TokenStart-1].Token.Folded != "not" {
		return empty, false
	}
	person, number := englishCandidateAgreement(subject)
	if person == "" || number != "singular" && number != "plural" {
		return empty, false
	}
	auxiliary := lattice.Tokens[predicate.TokenStart-2].Token.Folded
	tense := "present"
	if auxiliary == "did" {
		tense = "past"
	} else if !englishDoAuxiliaryMatches(auxiliary, person, number) {
		return empty, false
	}
	if string(predicate.Features["tense"]) != tense || string(predicate.Features["person"]) != person ||
		string(predicate.Features["number"]) != number || variantTenseConflicts(variant, main.ID, tense) {
		return empty, false
	}
	// The variant itself must license this exact operator sequence. Accepted
	// spellings do not license a different tense or relax subject agreement.
	for i, token := range variant.Sequence {
		if token.Kind != construction.TokenTerminal || (token.Terminal != "do" && token.Terminal != "does" && token.Terminal != "did") {
			continue
		}
		accepted := token.Terminal == auxiliary
		for _, spelling := range token.AcceptedInputs {
			accepted = accepted || spelling == auxiliary
		}
		if !accepted || i+2 >= len(variant.Sequence) || variant.Sequence[i+1].Kind != construction.TokenTerminal ||
			variant.Sequence[i+1].Terminal != "not" || variant.Sequence[i+2].Kind != construction.TokenSlot || variant.Sequence[i+2].SlotID != main.ID {
			return empty, false
		}
		return predicate, true
	}
	return empty, false
}

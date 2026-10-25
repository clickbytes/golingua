package pl

import (
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/valency"
)

func polishPersonalPredicateFeatures(includePast bool) []map[features.FeatureID]features.FeatureValueID {
	result := make([]map[features.FeatureID]features.FeatureValueID, 0, 4+8)
	for _, person := range []string{"first", "second"} {
		for _, number := range []string{"singular", "plural"} {
			result = append(result, map[features.FeatureID]features.FeatureValueID{
				"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": features.FeatureValueID(person), "number": features.FeatureValueID(number),
			})
		}
	}
	if !includePast {
		return result
	}
	for _, person := range []string{"first", "second"} {
		for _, number := range []string{"singular", "plural"} {
			for _, gender := range personalPastGenders(number) {
				result = append(result, map[features.FeatureID]features.FeatureValueID{
					"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": features.FeatureValueID(person), "number": features.FeatureValueID(number), "gender": features.FeatureValueID(gender),
				})
			}
		}
	}
	return result
}

func isPolishPersonalPronoun(id ids.LexemeID) bool {
	switch id.String() {
	case "lexeme:pl.ja.pron", "lexeme:pl.my.pron", "lexeme:pl.ty.pron", "lexeme:pl.wy.pron":
		return true
	default:
		return false
	}
}

func polishPersonalAnchor(person, number string) (meaning.DeicticAnchor, meaning.Multiplicity, bool) {
	var anchor string
	var multiplicity meaning.Multiplicity
	switch number {
	case "singular":
		multiplicity = meaning.MultiplicityOne
	case "plural":
		multiplicity = meaning.MultiplicityMany
	default:
		return meaning.DeicticAnchor{}, "", false
	}
	switch person {
	case "first":
		if number == "singular" {
			anchor = "speaker"
		} else {
			anchor = "speaker_inclusive_group"
		}
	case "second":
		anchor = "addressee"
	default:
		return meaning.DeicticAnchor{}, "", false
	}
	return meaning.DeicticAnchor{Kind: "participant", Value: anchor}, multiplicity, true
}

func polishPersonalEvidenceMatches(index *valency.Index, alternative parser.ParseAlternative, family string) (valency.ValencyFrame, bool) {
	if index == nil || alternative.FamilyID.String() != family || !strings.HasPrefix(alternative.VariantID.String(), family+".") {
		return valency.ValencyFrame{}, false
	}
	for _, binding := range index.BindingsByFamilyID(alternative.FamilyID) {
		if binding.SenseID != alternative.Permission.SenseID || binding.VariantID.String() != alternative.VariantID.String() {
			continue
		}
		for _, frame := range index.FramesBySenseID(binding.SenseID) {
			if frame.ID == binding.ValencyFrameID {
				return frame, true
			}
		}
	}
	return valency.ValencyFrame{}, false
}

func polishPersonalSubjectRole(frame valency.ValencyFrame) (meaning.Role, bool) {
	for _, argument := range frame.Arguments {
		if argument.ID == "subject" && argument.GrammaticalFunction == "subject" {
			return argument.Role, true
		}
	}
	return "", false
}

func addPolishPersonalEvidence(evidence runtime.CompositionEvidence, candidates []parser.PhraseCandidate, forest parser.Forest, index *valency.Index, registry *construction.Registry) runtime.CompositionEvidence {
	if !evidence.OperatorCoverageComplete || index == nil || registry == nil {
		return evidence
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	implicitFamilyID, err := ids.NewConstructionID("pl.transitive_personal")
	if err != nil {
		return evidence
	}
	explicitFamilyID, err := ids.NewConstructionID("pl.transitive_personal_explicit")
	if err != nil {
		return evidence
	}
	personID, err := ids.NewConceptID("person")
	if err != nil {
		return evidence
	}
	seenReferences := map[string]bool{}
	seenImplicit := map[string]bool{}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			familyText := alternative.FamilyID.String()
			isImplicitFamily := familyText == implicitFamilyID.String()
			isExplicitFamily := familyText == explicitFamilyID.String()
			if !isImplicitFamily && !isExplicitFamily {
				continue
			}
			frame, bindingFound := polishPersonalEvidenceMatches(index, alternative, familyText)
			if !bindingFound {
				continue
			}
			role, roleFound := polishPersonalSubjectRole(frame)
			if !roleFound {
				continue
			}
			if isExplicitFamily {
				for _, slot := range alternative.Slots {
					if slot.SlotID != "subject" {
						continue
					}
					for _, candidateID := range slot.CandidateIDs {
						candidate, exists := byID[candidateID]
						if !exists {
							continue
						}
						anchor, _, personal := polishAnchorForLexeme(candidate.LexemeID)
						if !personal || candidate.Category != construction.CategoryNP || seenReferences[candidate.ID] {
							continue
						}
						seenReferences[candidate.ID] = true
						evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: candidate.ID, Anchor: &anchor})
					}
				}
			}
			if !isImplicitFamily || !strings.Contains(alternative.VariantID.String(), ".prodrop.") || len(alternative.Unresolved) != 0 {
				continue
			}
			familyRecord, familyFound := registry.ByID(implicitFamilyID)
			if !familyFound {
				continue
			}
			familyBound, bindErr := construction.BindFrame(familyRecord, frame)
			if bindErr != nil {
				continue
			}
			if _, eligible := construction.ImplicitParticipantFamilySlot(familyBound.Family, "subject"); !eligible {
				continue
			}
			selectedPredicate := ""
			for _, slot := range alternative.Slots {
				if slot.SlotID != "verb" || len(slot.CandidateIDs) != 1 {
					continue
				}
				selectedPredicate = slot.CandidateIDs[0]
			}
			predicate, exists := byID[selectedPredicate]
			if !exists || predicate.Category != construction.CategoryVP || !candidateHasPersonalSense(predicate, alternative.Permission.SenseID) {
				continue
			}
			person, number := string(predicate.Features["person"]), string(predicate.Features["number"])
			anchor, multiplicity, known := polishPersonalAnchor(person, number)
			if !known {
				continue
			}
			key := alternative.DerivationID + "\x00subject"
			if seenImplicit[key] {
				continue
			}
			seenImplicit[key] = true
			evidence.ImplicitRoles = append(evidence.ImplicitRoles, semantic.ImplicitRoleEvidence{
				DerivationID: alternative.DerivationID, SlotID: "subject", PredicateCandidateID: predicate.ID,
				Role: role, CoreConceptID: personID, Multiplicity: multiplicity,
				ParticipantAnchor: anchor, SourceSpan: predicate.Span,
			})
		}
	}
	return evidence
}

func candidateHasPersonalSense(candidate parser.PhraseCandidate, sense ids.SenseID) bool {
	for _, candidateSense := range candidate.SenseIDs {
		if candidateSense == sense {
			return true
		}
	}
	return false
}

func polishAnchorForLexeme(lexeme ids.LexemeID) (meaning.DeicticAnchor, meaning.Multiplicity, bool) {
	for _, pronoun := range polishPersonalPronouns {
		if lexeme.String() == "lexeme:"+pronoun.lexeme {
			return polishPersonalAnchor(pronoun.person, pronoun.number)
		}
	}
	return meaning.DeicticAnchor{}, "", false
}

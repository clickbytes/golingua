package en

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

func englishNodeAttributes(supported map[ids.SenseID]bool, index *lexicon.Index, candidates []parser.PhraseCandidate, forest parser.Forest, evidence runtime.CompositionEvidence) ([]semantic.NodeAttributesEvidence, error) {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	operatorTargets := make(map[string]bool, len(evidence.Operators))
	for _, operator := range evidence.Operators {
		if operator.Operator.Kind == construction.OperatorNegation {
			operatorTargets[operator.TargetCandidateID] = true
		}
	}
	attributes := make(map[string]semantic.NodeAttributesEvidence)
	add := func(candidateID string, frame meaning.TemporalFrame, multiplicity meaning.Multiplicity) error {
		current := attributes[candidateID]
		if current.CandidateID != "" && current.TemporalFrame != frame && current.TemporalFrame != "" && frame != "" {
			return fmt.Errorf("candidate %q has conflicting licensed temporal frames", candidateID)
		}
		if current.CandidateID != "" && current.Multiplicity != multiplicity && current.Multiplicity != "" && multiplicity != "" {
			return fmt.Errorf("candidate %q has conflicting licensed multiplicity", candidateID)
		}
		current.CandidateID = candidateID
		if frame != "" {
			current.TemporalFrame = frame
		}
		if multiplicity != "" {
			current.Multiplicity = multiplicity
		}
		attributes[candidateID] = current
		return nil
	}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			sense := alternative.Permission.SenseID
			if !supported[sense] {
				continue
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || !containsEnglishSense(candidate.SenseIDs, sense) {
						continue
					}
					if candidate.Category == construction.CategoryVP {
						frame := englishTemporalFrame(string(candidate.Features["tense"]), string(candidate.Features["verb_form"]), operatorTargets[candidate.ID])
						if frame != "" {
							if err := add(candidate.ID, frame, ""); err != nil {
								return nil, err
							}
						}
					}
					_, attributionSense := englishAttributionForSense(sense.String())
					if candidate.Category == construction.CategoryAP && attributionSense && alternative.FamilyID.String() == "construction:en.attribution" &&
						(alternative.VariantID.String() == "construction:en.attribution.this" || alternative.VariantID.String() == "construction:en.attribution.that" || alternative.VariantID.String() == "construction:en.attribution.bare" || alternative.VariantID.String() == "construction:en.attribution.the" || alternative.VariantID.String() == "construction:en.attribution.the_plural" || alternative.VariantID.String() == "construction:en.attribution.a" || alternative.VariantID.String() == "construction:en.attribution.an" ||
							alternative.VariantID.String() == "construction:en.attribution.these" || alternative.VariantID.String() == "construction:en.attribution.those") {
						if err := add(candidate.ID, meaning.TemporalFramePresent, ""); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryNP {
			switch candidate.LexemeID.String() {
			case "lexeme:en.i.pron":
				if err := add(candidate.ID, "", meaning.MultiplicityOne); err != nil {
					return nil, err
				}
				continue
			case "lexeme:en.we.pron":
				if err := add(candidate.ID, "", meaning.MultiplicityMany); err != nil {
					return nil, err
				}
				continue
			case "lexeme:en.you.pron":
				continue
			}
		}
		if candidate.Category != construction.CategoryNP || !hasNumberBehavior(index, candidate.LexemeID, "count") {
			continue
		}
		number := knownEnglishNounNumber(candidate.Features)
		var multiplicity meaning.Multiplicity
		switch number {
		case "singular":
			multiplicity = meaning.MultiplicityOne
		case "plural":
			multiplicity = meaning.MultiplicityMany
		}
		if multiplicity != "" {
			if err := add(candidate.ID, "", multiplicity); err != nil {
				return nil, err
			}
		}
	}
	result := make([]semantic.NodeAttributesEvidence, 0, len(attributes))
	for _, attribute := range attributes {
		result = append(result, attribute)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CandidateID < result[j].CandidateID })
	return result, nil
}

func englishTemporalFrame(tense, form string, negatedBase bool) meaning.TemporalFrame {
	if form != "finite" && !(form == "base" && negatedBase) {
		return ""
	}
	switch tense {
	case "past":
		return meaning.TemporalFramePast
	case "present":
		return meaning.TemporalFramePresent
	default:
		return ""
	}
}

func knownEnglishNounNumber(values map[features.FeatureID]features.FeatureValueID) string {
	number, nounNumber := string(values["number"]), string(values["noun_number"])
	if number != "" && nounNumber != "" && number != nounNumber {
		return ""
	}
	if nounNumber != "" {
		return nounNumber
	}
	return number
}

func hasNumberBehavior(index *lexicon.Index, id ids.LexemeID, expected string) bool {
	if index == nil {
		return false
	}
	lexeme, exists := index.ByLexemeID(id)
	if !exists {
		return false
	}
	for _, property := range lexeme.Properties {
		if property.Key == "number_behavior" && property.Value == expected {
			return true
		}
	}
	return false
}

func containsEnglishSense(senses []ids.SenseID, expected ids.SenseID) bool {
	for _, sense := range senses {
		if sense == expected {
			return true
		}
	}
	return false
}

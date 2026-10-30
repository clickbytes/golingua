package pl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
)

func polishNodeAttributes(supported map[ids.SenseID]bool, index *lexicon.Index, candidates []parser.PhraseCandidate, forest parser.Forest) ([]semantic.NodeAttributesEvidence, error) {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
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
					if !exists || !containsPolishSense(candidate.SenseIDs, sense) {
						continue
					}
					if candidate.Category == construction.CategoryVP {
						frame := polishTemporalFrame(string(candidate.Features["tense"]), string(candidate.Features["verb_form"]))
						if frame != "" {
							if err := add(candidate.ID, frame, ""); err != nil {
								return nil, err
							}
						}
					}
					_, attributionSense := polishAttributionForSense(sense.String())
					variant := string(alternative.VariantID)
					if candidate.Category == construction.CategoryAP && attributionSense && (alternative.FamilyID.String() == "construction:pl.attribution" || alternative.FamilyID.String() == "construction:pl.attribution_plural" || alternative.FamilyID.String() == "construction:pl.attribution_negative") {
						frame := meaning.TemporalFramePresent
						if strings.Contains(variant, ".past.") {
							frame = meaning.TemporalFramePast
						}
						if err := add(candidate.ID, frame, ""); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}
	for _, candidate := range candidates {
		if candidate.Category != construction.CategoryNP || (!hasNumberBehavior(index, candidate.LexemeID, "count") && !isPolishPersonalPronoun(candidate.LexemeID)) || polishLexicalPluralDoesNotAssertCardinality(index, candidate.LexemeID) {
			continue
		}
		number := knownPolishNounNumber(candidate.Features)
		if isPolishPersonalPronoun(candidate.LexemeID) {
			number = string(candidate.Features["number"])
		}
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
	return sortedNodeAttributes(attributes), nil
}

func polishTemporalFrame(tense, form string) meaning.TemporalFrame {
	if form != "finite" {
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

func knownPolishNounNumber(values map[features.FeatureID]features.FeatureValueID) string {
	number, nounNumber := string(values["number"]), string(values["noun_number"])
	if number != "" && nounNumber != "" && number != nounNumber {
		return ""
	}
	if number != "" {
		return number
	}
	return nounNumber
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

func polishLexicalPluralDoesNotAssertCardinality(index *lexicon.Index, id ids.LexemeID) bool {
	if index == nil {
		return false
	}
	lexeme, exists := index.ByLexemeID(id)
	if !exists {
		return false
	}
	for _, property := range lexeme.Properties {
		if property.Key == "multiplicity_behavior" && property.Value == "lexical_plural_not_cardinality" {
			return true
		}
	}
	return false
}

func containsPolishSense(senses []ids.SenseID, expected ids.SenseID) bool {
	for _, sense := range senses {
		if sense == expected {
			return true
		}
	}
	return false
}

func sortedNodeAttributes(attributes map[string]semantic.NodeAttributesEvidence) []semantic.NodeAttributesEvidence {
	result := make([]semantic.NodeAttributesEvidence, 0, len(attributes))
	for _, attribute := range attributes {
		result = append(result, attribute)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CandidateID < result[j].CandidateID })
	return result
}

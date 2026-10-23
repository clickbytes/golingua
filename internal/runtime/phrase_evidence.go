package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
)

const MaxSimplePhraseCandidates = 4096

// POSPhraseConfig contains explicit language-local mappings from morphology
// labels to parser categories. The builder has no built-in POS assumptions.
type POSPhraseConfig struct {
	Categories  map[lexicon.PartOfSpeech][]construction.SlotCategory
	CaseFeature features.FeatureID
	CaseValues  map[features.FeatureValueID]string
}

// SimplePhraseEvidenceBuilder creates one-token phrase candidates from every
// licensed morphology reading and sense. It preserves ambiguity and only
// copies known feature values. Multiword expressions require a dedicated
// language-specific PhraseEvidenceBuilder.
type SimplePhraseEvidenceBuilder struct {
	Config POSPhraseConfig
}

func (builder SimplePhraseEvidenceBuilder) Build(_ context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	if len(lattice.MWEs) != 0 {
		return nil, parser.ParseContext{}, runtimeError(status.OutcomeUnsupported, "MWE_PHRASE_BUILDER_REQUIRED", "lattice.mwes", "multiword expression edges require a language-specific phrase evidence builder")
	}
	candidates := make([]parser.PhraseCandidate, 0)
	for _, token := range lattice.Tokens {
		for _, reading := range token.Readings {
			if !reading.HasLexeme {
				continue
			}
			categories := builder.Config.Categories[reading.Lexeme.POS]
			if len(categories) == 0 {
				continue
			}
			features := knownReadingFeatures(reading)
			senseIDs := make([]ids.SenseID, 0, len(reading.Senses))
			for _, sense := range reading.Senses {
				senseIDs = append(senseIDs, sense.Sense.ID)
			}
			caseValue, caseKnown := knownCase(reading, builder.Config.CaseFeature, builder.Config.CaseValues)
			if _, exists := features["case"]; exists && !caseKnown {
				// Parser case evidence must have a matching dedicated label.
				// The original morphology reading still retains its exact case.
				delete(features, "case")
			}
			seenCategories := make(map[construction.SlotCategory]struct{}, len(categories))
			for _, category := range categories {
				if !validPhraseCategory(category) {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeInvalid, "INVALID_POS_PHRASE_CATEGORY", "phrase_categories."+string(reading.Lexeme.POS), "POS category mapping contains an unsupported phrase category")
				}
				if _, duplicate := seenCategories[category]; duplicate {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeInvalid, "DUPLICATE_POS_PHRASE_CATEGORY", "phrase_categories."+string(reading.Lexeme.POS), "POS category mapping contains a duplicate category")
				}
				seenCategories[category] = struct{}{}
				if len(candidates) >= MaxSimplePhraseCandidates {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeResourceLimit, "PHRASE_CANDIDATE_LIMIT", "phrase_candidates", "simple phrase candidate limit was exhausted")
				}
				headTokenIndex := 0
				if category == construction.CategoryVP {
					headTokenIndex = token.Index
				}
				candidates = append(candidates, parser.PhraseCandidate{
					ID:       simplePhraseID(token.Index, reading.Reading.ID, category),
					Category: category, TokenStart: token.Index, TokenEnd: token.Index + 1,
					Span: token.Token.Span, HeadTokenIndex: headTokenIndex,
					ReadingIDs: []string{reading.Reading.ID}, LexemeID: reading.Lexeme.ID,
					SenseIDs: append([]ids.SenseID(nil), senseIDs...), Features: clonePhraseFeatures(features), Case: caseValue, CaseKnown: caseKnown,
				})
			}
		}
	}
	return candidates, parser.ParseContext{}, nil
}

func clonePhraseFeatures(source map[features.FeatureID]features.FeatureValueID) map[features.FeatureID]features.FeatureValueID {
	result := make(map[features.FeatureID]features.FeatureValueID, len(source))
	for feature, value := range source {
		result[feature] = value
	}
	return result
}

func knownReadingFeatures(candidate analyzer.ReadingCandidate) map[features.FeatureID]features.FeatureValueID {
	values := make(map[features.FeatureID]features.FeatureValueID)
	for _, value := range candidate.Reading.Features.Values {
		if value.State == features.StateKnown {
			values[value.Feature] = value.Value
		}
	}
	return values
}

func knownCase(candidate analyzer.ReadingCandidate, feature features.FeatureID, labels map[features.FeatureValueID]string) (string, bool) {
	if feature == "" {
		return "", false
	}
	for _, value := range candidate.Reading.Features.Values {
		if value.Feature == feature && value.State == features.StateKnown {
			label, exists := labels[value.Value]
			return label, exists && label != ""
		}
	}
	return "", false
}

func validPhraseCategory(category construction.SlotCategory) bool {
	return category == construction.CategoryNP || category == construction.CategoryVP || category == construction.CategoryPP
}

func simplePhraseID(tokenIndex int, readingID string, category construction.SlotCategory) string {
	identity := fmt.Sprintf("%d\x00%s\x00%s", tokenIndex, readingID, category)
	digest := sha256.Sum256([]byte(identity))
	return "phrase:" + hex.EncodeToString(digest[:])
}

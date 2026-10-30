package en

import (
	"context"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type englishDemonstrative struct {
	distance string
	number   string
}

func englishDemonstrativeAt(lattice analyzer.LexicalLattice, index int) (englishDemonstrative, bool) {
	if index < 0 || index >= len(lattice.Tokens) {
		return englishDemonstrative{}, false
	}
	token := lattice.Tokens[index].Token
	surface := token.Folded
	if surface == "" {
		surface = token.Normalized
	}
	switch surface {
	case "this":
		return englishDemonstrative{distance: "proximal", number: "singular"}, true
	case "these":
		return englishDemonstrative{distance: "proximal", number: "plural"}, true
	case "that":
		return englishDemonstrative{distance: "distal", number: "singular"}, true
	case "those":
		return englishDemonstrative{distance: "distal", number: "plural"}, true
	default:
		return englishDemonstrative{}, false
	}
}

func englishExactDemonstrativeNP(candidate parser.PhraseCandidate, alternative parser.ParseAlternative, modifierSlot, headSlot string, byID map[string]parser.PhraseCandidate, lattice analyzer.LexicalLattice) bool {
	var group *parser.NPPhraseGroupBinding
	for i := range alternative.NPPhraseGroups {
		item := &alternative.NPPhraseGroups[i]
		if item.HeadSlotID == headSlot && item.HeadCandidateID == candidate.ID && item.ModifierSlotID == modifierSlot {
			group = item
			break
		}
	}
	if group == nil || group.DeterminerCandidateID == "" || group.ModifierCandidateID == "" || candidate.Features["number"] == "" {
		return false
	}
	modifier, modifierFound := byID[group.ModifierCandidateID]
	if !modifierFound || modifier.Category != construction.CategoryAP {
		return false
	}
	for _, composite := range candidate.CompositeNPs {
		if composite.ID != group.CompositeCandidateID || composite.HeadCandidateID != candidate.ID || composite.ModifierCandidateID != modifier.ID ||
			composite.DeterminerCandidateID != group.DeterminerCandidateID || composite.Determiner == nil || composite.Outcome != status.OutcomeOK ||
			composite.Plan.Plan.TemplateID != "np-attributive" ||
			composite.TokenStart != group.TokenStart || composite.TokenEnd != group.TokenEnd || composite.Determiner.TokenIndex != group.TokenStart {
			continue
		}
		demonstrative, ok := englishDemonstrativeAt(lattice, composite.Determiner.TokenIndex)
		if !ok || composite.Determiner.Surface != lattice.Tokens[group.TokenStart].Token.Surface || composite.Determiner.Span != lattice.Tokens[group.TokenStart].Token.Span ||
			composite.Determiner.Features["determiner"] != features.FeatureValueID("demonstrative_"+demonstrative.distance) ||
			composite.Determiner.Features["number"] != features.FeatureValueID(demonstrative.number) || candidate.Features["number"] != features.FeatureValueID(demonstrative.number) {
			continue
		}
		return true
	}
	return false
}

type englishIdentificationPhraseBuilder struct {
	inner          englishDeterminerNPBuilder
	lexicon        *lexicon.Index
	phraseProfile  construction.PhraseProfile
	featureProfile *features.LanguageFeatureProfile
}

func (builder englishIdentificationPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	allowedAttributionLexemes := make(map[string]bool, len(englishAttributionSpecs))
	for _, spec := range englishAttributionSpecs {
		allowedAttributionLexemes[spec.lexeme] = true
	}
	const beLexeme = "lexeme:en.be.v"
	attributionAt := make(map[int]bool)
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && allowedAttributionLexemes[candidate.LexemeID.String()] {
			attributionAt[candidate.TokenStart] = true
		}
	}
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && !allowedAttributionLexemes[candidate.LexemeID.String()] {
			continue
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == "lexeme:en.opening_object.v" &&
			containsEnglishSenseName(candidate.SenseIDs, "en.opening_object.event") && (attributionAt[candidate.TokenStart] &&
			candidate.TokenStart > 0 && (lattice.Tokens[candidate.TokenStart-1].Token.Folded == "is" || lattice.Tokens[candidate.TokenStart-1].Token.Folded == "are") ||
			englishOpenAdjectiveHeadsNP(candidate, candidates, lattice, builder.lexicon)) {
			// The same open token has an exact licensed adjective + NP analysis
			// (or the established be + adjective frame); its unmatched transitive
			// verb reading cannot head a competing clause candidate.
			continue
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == beLexeme && attributionAt[candidate.TokenEnd] {
			continue
		}
		if candidate.Category == construction.CategoryNP {
			article, hasArticle := englishIdentificationPrefixArticle(lattice, candidate.TokenStart)
			if hasArticle {
				if candidate.Features["noun_number"] != "singular" {
					continue
				}
				lexeme, found := builder.lexicon.ByLexemeID(candidate.LexemeID)
				if !found {
					continue
				}
				expected, known := englishIndefiniteArticle(lexeme.Properties)
				if !known || expected != article || !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") {
					continue
				}
			} else if hasEnglishPluralIdentificationPrefix(lattice, candidate.TokenStart) {
				if candidate.Features["noun_number"] != "plural" || hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass") {
					continue
				}
			} else if hasEnglishThisNPPrefix(lattice, candidate.TokenStart) {
				if !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass") && !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") {
					continue
				}
			}
		}
		filtered = append(filtered, candidate)
	}
	composed, err := runtime.AddAttributiveNPEvidence(lattice, filtered, builder.phraseProfile, builder.featureProfile, func(lattice analyzer.LexicalLattice, head, modifier parser.PhraseCandidate) *parser.NPPhraseTokenEvidence {
		return englishNPPhraseDeterminer(lattice, head, modifier, builder.lexicon)
	})
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	return composed, parseContext, nil
}

func englishNPPhraseDeterminer(lattice analyzer.LexicalLattice, _ parser.PhraseCandidate, modifier parser.PhraseCandidate, lexiconIndex *lexicon.Index) *parser.NPPhraseTokenEvidence {
	tokenIndex := modifier.TokenStart - 1
	if tokenIndex < 0 || tokenIndex >= len(lattice.Tokens) {
		return nil
	}
	token := lattice.Tokens[tokenIndex].Token
	values := map[features.FeatureID]features.FeatureValueID{}
	if article, ok := englishArticleAt(lattice, tokenIndex); ok {
		if article == "a" || article == "an" {
			values["determiner"] = "indefinite"
			values["number"] = "singular"
		} else {
			values["determiner"] = "definite"
		}
	} else if deixis, ok := englishDemonstrativeAt(lattice, tokenIndex); ok {
		values["determiner"] = features.FeatureValueID("demonstrative_" + deixis.distance)
		values["number"] = features.FeatureValueID(deixis.number)
	} else {
		return nil
	}
	var onset string
	if selected, ok := lexiconIndex.ByLexemeID(modifier.LexemeID); ok {
		for _, property := range selected.Properties {
			if property.Key != "indefinite_article" {
				continue
			}
			switch property.Value {
			case "a":
				onset = "consonant"
			case "an":
				onset = "vowel"
			}
		}
	}
	return runtime.NewNPPhraseTokenEvidenceWithOnset("en", tokenIndex, token.Surface, token.Span, values, onset)
}

func englishOpenAdjectiveHeadsNP(verb parser.PhraseCandidate, candidates []parser.PhraseCandidate, lattice analyzer.LexicalLattice, index *lexicon.Index) bool {
	for _, adjective := range candidates {
		if adjective.Category != construction.CategoryAP || adjective.LexemeID.String() != "lexeme:en.open.adj" ||
			!containsEnglishSenseName(adjective.SenseIDs, "en.open.state") || adjective.TokenStart != verb.TokenStart || adjective.TokenEnd != verb.TokenEnd || adjective.TokenStart == 0 {
			continue
		}
		if _, allowed := englishAttributionArticleForPhrase(adjective, index); !allowed {
			continue
		}
		article := lattice.Tokens[adjective.TokenStart-1].Token.Folded
		if article != "the" && article != "a" && article != "an" {
			continue
		}
		for _, nounPhrase := range candidates {
			if nounPhrase.Category == construction.CategoryNP && nounPhrase.TokenStart == adjective.TokenEnd {
				return true
			}
		}
	}
	return false
}

func hasEnglishPluralIdentificationPrefix(lattice analyzer.LexicalLattice, start int) bool {
	return start >= 2 && lattice.Tokens[start-1].Token.Folded == "are" &&
		(lattice.Tokens[start-2].Token.Folded == "these" || lattice.Tokens[start-2].Token.Folded == "those")
}

func englishIdentificationPrefixArticle(lattice analyzer.LexicalLattice, start int) (string, bool) {
	if start >= 3 && (lattice.Tokens[start-1].Token.Folded == "a" || lattice.Tokens[start-1].Token.Folded == "an") &&
		lattice.Tokens[start-2].Token.Folded == "is" && (lattice.Tokens[start-3].Token.Folded == "this" || lattice.Tokens[start-3].Token.Folded == "that" || lattice.Tokens[start-3].Token.Folded == "it") {
		return lattice.Tokens[start-1].Token.Folded, true
	}
	if start >= 1 {
		switch lattice.Tokens[start-1].Token.Folded {
		case "a":
			return "a", true
		case "an":
			return "an", true
		}
	}
	return "", false
}

func hasEnglishThisNPPrefix(lattice analyzer.LexicalLattice, start int) bool {
	if start >= 1 && (lattice.Tokens[start-1].Token.Folded == "this" || lattice.Tokens[start-1].Token.Folded == "that") {
		return true
	}
	if start >= 2 && (lattice.Tokens[start-2].Token.Folded == "this" || lattice.Tokens[start-2].Token.Folded == "that") && lattice.Tokens[start-1].Token.Folded == "is" {
		return true
	}
	return start >= 3 && (lattice.Tokens[start-3].Token.Folded == "this" || lattice.Tokens[start-3].Token.Folded == "that") &&
		lattice.Tokens[start-2].Token.Folded == "is" && lattice.Tokens[start-1].Token.Folded == "a"
}

func addEnglishIdentificationReferences(evidence runtime.CompositionEvidence, candidates []parser.PhraseCandidate, forest parser.Forest, lattice analyzer.LexicalLattice) runtime.CompositionEvidence {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	seen := make(map[string]bool, len(evidence.References))
	for _, reference := range evidence.References {
		if reference.Anchor != nil {
			seen[reference.CandidateID+"\x00"+reference.Anchor.Kind+"\x00"+reference.Anchor.Value] = true
		}
	}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			sense := alternative.Permission.SenseID.Symbol().Name
			senseID := alternative.Permission.SenseID.String()
			frame := string(alternative.Permission.FrameID)
			family := alternative.FamilyID.String()
			variant := strings.TrimSuffix(string(alternative.VariantID), ".past")
			var slotID string
			modifierSlotID := "modifier"
			exactSeeingDemonstrative := false
			switch {
			case sense == "en.hear.auditory" && frame == "en.hear.auditory.attributed_subject" && family == "construction:en.attributed_seeing" &&
				(variant == "construction:en.attributed_seeing.svo.demonstrative_this" || variant == "construction:en.attributed_seeing.svo.demonstrative_that" ||
					variant == "construction:en.attributed_seeing.svo.demonstrative_these" || variant == "construction:en.attributed_seeing.svo.demonstrative_those"):
				slotID, exactSeeingDemonstrative = "subject", true
			case sense == "en.see.visual" && frame == "en.see.visual.transitive" && family == "construction:en.attributed_seeing" &&
				(variant == "construction:en.attributed_seeing.svo.demonstrative_this" || variant == "construction:en.attributed_seeing.svo.demonstrative_that" ||
					variant == "construction:en.attributed_seeing.svo.demonstrative_these" || variant == "construction:en.attributed_seeing.svo.demonstrative_those"):
				slotID, exactSeeingDemonstrative = "subject", true
			case sense == "en.see.visual" && frame == "en.see.visual.transitive" && family == "construction:en.attributed_seeing_both" &&
				strings.HasPrefix(variant, "construction:en.attributed_seeing_both.svo.demonstrative_"):
				slotID, modifierSlotID, exactSeeingDemonstrative = "subject", "modifier_subject", true
			case sense == "en.see.visual" && frame == "en.see.visual.transitive" && family == "construction:en.attributed_seeing_object" &&
				(variant == "construction:en.attributed_seeing_object.svo.this_object" || variant == "construction:en.attributed_seeing_object.svo.that_object" ||
					variant == "construction:en.attributed_seeing_object.svo.these_object" || variant == "construction:en.attributed_seeing_object.svo.those_object"):
				slotID, modifierSlotID, exactSeeingDemonstrative = "object", "modifier_object", true
			case sense == "en.be.class_membership" && frame == "en.be.class_membership.frame" && family == "construction:en.identification_negative" && variant == "construction:en.identification_negative.this":
				slotID = "member"
			case sense == "en.be.class_membership" && frame == "en.be.class_membership.frame" && family == "construction:en.identification" && (variant == "construction:en.identification.this" || variant == "construction:en.identification.this.an" || variant == "construction:en.identification.this.mass" || variant == "construction:en.identification.that" || variant == "construction:en.identification.that.an" || variant == "construction:en.identification.that.mass" || variant == "construction:en.identification.these" || variant == "construction:en.identification.those"):
				slotID = "member"
			case func() bool {
				_, allowedSense := englishAttributionForSense(senseID)
				return allowedSense && frame == sense+".frame" && family == "construction:en.attribution" &&
					(variant == "construction:en.attribution.this" || variant == "construction:en.attribution.that" ||
						variant == "construction:en.attribution.these" || variant == "construction:en.attribution.those" || variant == "construction:en.attribution.bare" || variant == "construction:en.attribution.bare_plural" || variant == "construction:en.attribution.the_plural" ||
						variant == "construction:en.attribution.the" || variant == "construction:en.attribution.a" || variant == "construction:en.attribution.an")
			}():
				slotID = "bearer"
			default:
				continue
			}
			for _, slot := range alternative.Slots {
				if slot.SlotID != slotID {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || candidate.Category != construction.CategoryNP {
						continue
					}
					if exactSeeingDemonstrative && !englishExactDemonstrativeNP(candidate, alternative, modifierSlotID, slotID, byID, lattice) {
						continue
					}
					anchorValue := "proximal"
					if strings.Contains(variant, "demonstrative_that") || strings.Contains(variant, "demonstrative_those") ||
						strings.HasSuffix(variant, ".that") || strings.HasSuffix(variant, ".those") ||
						strings.HasSuffix(variant, ".that_object") || strings.HasSuffix(variant, ".those_object") {
						anchorValue = "distal"
					}
					if strings.HasSuffix(variant, ".bare") || strings.HasSuffix(variant, ".bare_plural") || strings.HasSuffix(variant, ".the") || strings.HasSuffix(variant, ".the_plural") || strings.HasSuffix(variant, ".a") || strings.HasSuffix(variant, ".an") {
						evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: candidateID, Unanchored: true})
						continue
					}
					key := candidateID + "\x00deixis\x00" + anchorValue
					if seen[key] {
						continue
					}
					anchor := &meaning.DeicticAnchor{Kind: "deixis", Value: anchorValue}
					evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: candidateID, Anchor: anchor})
					seen[key] = true
				}
			}
		}
	}
	return evidence
}

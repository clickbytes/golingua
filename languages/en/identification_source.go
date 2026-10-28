package en

import (
	"context"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

type englishIdentificationPhraseBuilder struct {
	inner   englishDeterminerNPBuilder
	lexicon *lexicon.Index
}

func (builder englishIdentificationPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	const blackLexeme = "lexeme:en.black.adj"
	const beLexeme = "lexeme:en.be.v"
	blackAt := make(map[int]bool)
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && candidate.LexemeID.String() == blackLexeme {
			blackAt[candidate.TokenStart] = true
		}
	}
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && candidate.LexemeID.String() != blackLexeme {
			continue
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == beLexeme && blackAt[candidate.TokenEnd] {
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
				if !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass") {
					continue
				}
			}
		}
		filtered = append(filtered, candidate)
	}
	return filtered, parseContext, nil
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

func addEnglishIdentificationReferences(evidence runtime.CompositionEvidence, candidates []parser.PhraseCandidate, forest parser.Forest) runtime.CompositionEvidence {
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
			frame := string(alternative.Permission.FrameID)
			family := alternative.FamilyID.String()
			variant := string(alternative.VariantID)
			var slotID string
			switch {
			case sense == "en.be.class_membership" && frame == "en.be.class_membership.frame" && family == "construction:en.identification_negative" && variant == "construction:en.identification_negative.this":
				slotID = "member"
			case sense == "en.be.class_membership" && frame == "en.be.class_membership.frame" && family == "construction:en.identification" && (variant == "construction:en.identification.this" || variant == "construction:en.identification.this.an" || variant == "construction:en.identification.this.mass" || variant == "construction:en.identification.that" || variant == "construction:en.identification.that.an" || variant == "construction:en.identification.that.mass" || variant == "construction:en.identification.these" || variant == "construction:en.identification.those"):
				slotID = "member"
			case sense == "en.black.color" && frame == "en.black.color.frame" && family == "construction:en.attribution" && variant == "construction:en.attribution.this":
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
					anchorValue := "proximal"
					if strings.HasPrefix(variant, "construction:en.identification.that") || variant == "construction:en.identification.those" {
						anchorValue = "distal"
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

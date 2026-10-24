package pl

import (
	"context"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

type polishIdentificationPhraseBuilder struct {
	inner runtime.SimplePhraseEvidenceBuilder
}

func (builder polishIdentificationPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	const adjectiveLexeme = "lexeme:pl.czarny.adj"
	const copulaLexeme = "lexeme:pl.byc.v"
	adjectiveAt := make(map[int]bool)
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && candidate.LexemeID.String() == adjectiveLexeme {
			adjectiveAt[candidate.TokenStart] = true
		}
	}
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && candidate.LexemeID.String() != adjectiveLexeme {
			continue
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == copulaLexeme && adjectiveAt[candidate.TokenEnd] {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered, parseContext, nil
}

func addPolishIdentificationReferences(evidence runtime.CompositionEvidence, candidates []parser.PhraseCandidate, forest parser.Forest) runtime.CompositionEvidence {
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
			frame := string(alternative.Permission.FrameID)
			family := alternative.FamilyID.String()
			variant := string(alternative.VariantID)
			sense := alternative.Permission.SenseID.Symbol().Name
			var slotID, anchorValue string
			switch {
			case sense == "pl.byc.class_membership" && frame == "pl.byc.class_membership.frame" && family == "construction:pl.identification" && variant == "construction:pl.identification.to":
				slotID, anchorValue = "member", "unresolved"
			case sense == "pl.czarny.black" && frame == "pl.czarny.black.frame" && family == "construction:pl.attribution" && variant == "construction:pl.attribution.ten":
				slotID, anchorValue = "bearer", "demonstrative"
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

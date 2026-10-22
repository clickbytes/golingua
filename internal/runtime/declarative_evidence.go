package runtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// BareDeclarativeEvidenceBuilder attests absence of separate semantic
// operators only for a complete parse whose predicate has an explicitly
// supported sense and finite-form feature signature. Other input keeps
// operator coverage unknown and cannot become an OK semantic candidate.
type BareDeclarativeEvidenceBuilder struct {
	SupportedSenses           map[ids.SenseID]bool
	RequiredPredicateFeatures map[features.FeatureID]features.FeatureValueID
}

func (builder BareDeclarativeEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (CompositionEvidence, error) {
	if err := ctx.Err(); err != nil {
		return CompositionEvidence{}, err
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return CompositionEvidence{}, fmt.Errorf("request context identity: %w", err)
	}
	evidence := CompositionEvidence{RequestContextID: "request-" + hex.EncodeToString(random[:])}
	if forest.Outcome != status.OutcomeOK || len(forest.Roots) == 0 || len(forest.Unmatched) != 0 || len(lattice.MWEs) != 0 {
		return evidence, nil
	}
	for _, token := range lattice.Tokens {
		if token.Token.Kind != text.TokenWord || len(token.Readings) == 0 {
			return evidence, nil
		}
	}
	byCandidate := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byCandidate[candidate.ID] = candidate
	}
	rootCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		rootCount++
		for _, alternative := range node.Alternatives {
			if len(alternative.Unresolved) != 0 || !builder.SupportedSenses[alternative.Permission.SenseID] {
				return evidence, nil
			}
			predicateFound := false
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, id := range slot.CandidateIDs {
					candidate, exists := byCandidate[id]
					if !exists || candidate.Category != construction.CategoryVP || !predicateFeaturesMatch(candidate.Features, builder.RequiredPredicateFeatures) {
						return evidence, nil
					}
					predicateFound = true
				}
			}
			if !predicateFound {
				return evidence, nil
			}
		}
	}
	if rootCount == 0 {
		return evidence, nil
	}
	evidence.OperatorCoverageComplete = true
	return evidence, nil
}

func predicateFeaturesMatch(actual, required map[features.FeatureID]features.FeatureValueID) bool {
	if len(required) == 0 {
		return false
	}
	for feature, value := range required {
		if actual[feature] != value {
			return false
		}
	}
	return true
}

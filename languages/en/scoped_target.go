package en

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
)

type englishScopedFeaturePlanner struct{}

func (englishScopedFeaturePlanner) Features(_ context.Context, req engine.ScopedTargetRequest, part construction.ScopedClausePart, _, _ lexicon.MorphLexeme) ([]features.FeatureValue, error) {
	switch part.Slot {
	case "x", "y":
		return []features.FeatureValue{
			{Feature: "noun_number", State: features.StateKnown, Value: "singular"},
			{Feature: "number", State: features.StateKnown, Value: "singular"},
		}, nil
	case "event":
		if req.EventTemporalFrame != "present" {
			return nil, fmt.Errorf("English scoped event requires explicit present temporal frame")
		}
		return englishFiniteVerbFeatures("present", "third", "singular"), nil
	default:
		return nil, fmt.Errorf("unsupported English scoped slot %q", part.Slot)
	}
}

package pl

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
)

type polishScopedFeaturePlanner struct{}

func (polishScopedFeaturePlanner) Features(_ context.Context, req engine.ScopedTargetRequest, part construction.ScopedClausePart, lexeme, _ lexicon.MorphLexeme) ([]features.FeatureValue, error) {
	switch part.Slot {
	case "x", "y":
		if part.Case == "" {
			return nil, fmt.Errorf("Polish scoped noun slot requires an explicit case")
		}
		gender := ""
		for _, property := range lexeme.Properties {
			if property.Key == "gender" {
				if gender != "" || property.Value == "" {
					return nil, fmt.Errorf("Polish scoped noun requires one unambiguous seed gender")
				}
				gender = property.Value
			}
		}
		if gender == "" {
			return nil, fmt.Errorf("Polish scoped noun seed has no gender")
		}
		return knownPolishFeatures([2]string{"case", part.Case}, [2]string{"number", "singular"}, [2]string{"gender", gender}), nil
	case "event":
		if req.EventTemporalFrame != "present" {
			return nil, fmt.Errorf("Polish scoped event requires explicit present temporal frame")
		}
		return knownPolishFeatures(
			[2]string{"verb_form", "finite"}, [2]string{"tense", "present"},
			[2]string{"aspect", "imperfective"}, [2]string{"person", "third"},
			[2]string{"number", "singular"}, [2]string{"mood", "indicative"},
			[2]string{"voice", "active"},
		), nil
	default:
		return nil, fmt.Errorf("unsupported Polish scoped slot %q", part.Slot)
	}
}

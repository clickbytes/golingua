package pl

import (
	"context"
	"embed"
	"fmt"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
)

//go:embed constructions/relation_path.json
var relationPathPlanFS embed.FS

func SentenceRelationPathPlan() (engine.RelationPathPlan, error) {
	data, err := relationPathPlanFS.ReadFile("constructions/relation_path.json")
	if err != nil {
		return engine.RelationPathPlan{}, err
	}
	return engine.DecodeRelationPathPlan(data, "pl")
}

type polishRelationPathFeaturePlanner struct{}

func (polishRelationPathFeaturePlanner) Features(ctx context.Context, request engine.RelationPathRequest, _ engine.RelationPathVariant, part engine.RelationPathPart, lexeme lexicon.MorphLexeme) ([]features.FeatureValue, error) {
	if ctx == nil {
		return nil, fmt.Errorf("Polish relation path features require a request context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if part.Slot != "verb" || part.Step != "first_atom" && part.Step != "second_atom" || lexeme.ID != request.Selection.Lexeme.ID || lexeme.Language != "pl" || lexeme.POS != "verb" {
		return nil, fmt.Errorf("Polish relation path verb slot does not match the exact selected Polish verb")
	}
	return []features.FeatureValue{
		{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
		{Feature: "tense", State: features.StateKnown, Value: "present"},
		{Feature: "aspect", State: features.StateKnown, Value: "imperfective"},
		{Feature: "mood", State: features.StateKnown, Value: "indicative"},
		{Feature: "voice", State: features.StateKnown, Value: "active"},
		{Feature: "person", State: features.StateKnown, Value: "third"},
		{Feature: "number", State: features.StateKnown, Value: "singular"},
	}, nil
}

// RelationSourceReadingPolicy accepts known Polish verb readings that match
// the finite present third-person singular active imperfective plans.
func RelationSourceReadingPolicy(ctx context.Context, variant engine.RelationPathVariant, part engine.RelationPathPart, candidate analyzer.ReadingCandidate, sense analyzer.SenseCandidate) (bool, error) {
	if ctx == nil {
		return false, fmt.Errorf("Polish relation source policy requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if part.Slot != "verb" || part.Step != "first_atom" && part.Step != "second_atom" || candidate.Lexeme.Language != "pl" || string(candidate.Lexeme.POS) != "verb" || candidate.Reading.POS != "verb" {
		return false, nil
	}
	scopedVariant := construction.ScopedClauseVariant{
		ID:          variant.ID,
		Constraints: []construction.ScopedClauseVariantConstraint{{Binder: "event", Feature: "temporal_frame", Values: []string{"present"}}},
	}
	return ScopedSourceReadingPolicy(ctx, scopedVariant, construction.ScopedClausePart{Slot: "event"}, candidate, sense)
}

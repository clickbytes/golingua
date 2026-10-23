package en

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// NewSentenceExecutionBuilder compiles the English morphology inputs used by
// the bounded sentence realization slice.
func NewSentenceExecutionBuilder() (runtime.MorphologyExecutionBuilder, error) {
	seed, err := Load()
	if err != nil {
		return runtime.MorphologyExecutionBuilder{}, err
	}
	profile, _, paradigms, exceptions := seed.RuntimeMorphology()
	return runtime.MorphologyExecutionBuilder{
		Profile: profile, Paradigms: paradigms, Exceptions: exceptions,
		Parts: seed.PrincipalParts, FeaturePlan: englishSentenceFeatures{},
	}, nil
}

// NewSentenceTarget composes English target selection, planning, morphology
// and linearization over one validated snapshot and registry.
func NewSentenceTarget(snapshot *pack.Pack, registry *language.Registry) (*runtime.TargetPipeline, error) {
	if snapshot == nil || registry == nil {
		return nil, fmt.Errorf("English target pipeline requires a snapshot and verified language registry")
	}
	seed, err := Load()
	if err != nil {
		return nil, err
	}
	profile, _, _, _ := seed.RuntimeMorphology()
	valencyIndex, err := CompileValency()
	if err != nil {
		return nil, err
	}
	packages, constructions, err := SentenceConstructions(valencyIndex)
	if err != nil {
		return nil, err
	}
	_, variants, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
	if err != nil {
		return nil, err
	}
	execution, err := NewSentenceExecutionBuilder()
	if err != nil {
		return nil, err
	}
	return &runtime.TargetPipeline{
		Language: "en", LanguageRegistry: registry, Lexicon: snapshot.Lexicon(), Valency: valencyIndex,
		Constructions: constructions, Variants: variants, Core: snapshot.CoreGraph(),
		OperatorBuilder: runtime.NoOperatorEvidenceBuilder{}, ExecutionBuilder: execution,
		PlanBuilder: englishSentencePlanAnnotations{},
	}, nil
}

type englishSentenceFeatures struct{}

func (englishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, _ []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.Category == construction.CategoryVP {
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		if !found {
			return nil, nil
		}
		if values["verb_form"] == "past" || values["tense"] == "past" {
			return knownFeature("verb_form", "past"), nil
		}
		if values["verb_form"] == "third_person_singular" ||
			values["verb_form"] == "finite" && values["tense"] == "present" && values["person"] == "third" && values["number"] == "singular" {
			return knownFeature("verb_form", "third_person_singular"), nil
		}
		return nil, nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil {
		return nil, nil
	}
	values, found := runtime.SourceFeatures(source, slot.ArgumentBinding.SemanticRole, candidate.TransferBinding.CoreRef)
	if !found {
		return nil, nil
	}
	number := values["number"]
	if number == "" {
		number = values["noun_number"]
	}
	for _, property := range candidate.Lexeme.Properties {
		if property.Key == "number_behavior" && property.Value == "mass" {
			// English mass nouns use the singular attested form even when the
			// source language lexicalizes the same substance as plural.
			number = "singular"
			break
		}
	}
	if number != "singular" && number != "plural" {
		return nil, nil
	}
	return knownFeature("noun_number", string(number)), nil
}

type englishSentencePlanAnnotations struct{}

func (englishSentencePlanAnnotations) Annotations(_ context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error) {
	if selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState || source.Meaning == nil {
		return nil, nil, nil, nil
	}
	predicate := findMeaningNode(*source.Meaning, selected.NodeID)
	if predicate.ID == "" {
		return nil, nil, nil, nil
	}
	assumptions := make([]realizer.FieldAssumption, 0, len(predicate.Participants))
	losses := make([]realizer.FieldLoss, 0, len(predicate.Participants))
	for _, participant := range predicate.Participants {
		referent := findMeaningNode(*source.Meaning, participant.Target)
		if referent.ID == "" {
			continue
		}
		values, found := runtime.SourceFeatures(source, participant.Role, coreConcept(referent))
		if found && values["determiner"] == "definite" {
			continue
		}
		if found && values["determiner"] == "indefinite" {
			losses = append(losses, realizer.FieldLoss{
				Path: []meaning.NodeID{selected.NodeID, participant.Target}, Field: "determiner", Outcome: status.OutcomeUnsupported,
				Reason: "English target construction currently realizes only a definite article; an explicit indefinite source determiner cannot be preserved",
			})
			continue
		}
		assumptions = append(assumptions, realizer.FieldAssumption{
			Path: []meaning.NodeID{selected.NodeID, participant.Target}, Field: "determiner", Value: "definite",
			Reason: "English transitive template realizes the definite determiner; source Meaning and morphology provide no explicit definiteness value",
		})
	}
	return nil, assumptions, losses, nil
}

func knownFeature(feature features.FeatureID, value string) []features.FeatureValue {
	return []features.FeatureValue{{Feature: feature, State: features.StateKnown, Value: features.FeatureValueID(value)}}
}

func findMeaningNode(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func coreConcept(node meaning.Node) ids.ConceptID {
	id, _ := ids.ParseConceptID(string(node.Concept))
	return id
}

package pl

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// NewSentenceExecutionBuilder compiles the Polish morphology inputs used by
// the bounded sentence realization slice.
func NewSentenceExecutionBuilder() (runtime.MorphologyExecutionBuilder, error) {
	seed, err := Load()
	if err != nil {
		return runtime.MorphologyExecutionBuilder{}, err
	}
	profile, _, paradigms, exceptions := seed.RuntimeMorphology()
	return runtime.MorphologyExecutionBuilder{
		Profile: profile, Paradigms: paradigms, Exceptions: exceptions,
		Parts: seed.PrincipalParts, FeaturePlan: polishSentenceFeatures{}, CaseFeature: "case",
	}, nil
}

// NewSentenceTarget composes Polish target selection, planning, morphology
// and linearization over one validated snapshot and registry.
func NewSentenceTarget(snapshot *pack.Pack, registry *language.Registry) (*runtime.TargetPipeline, error) {
	if snapshot == nil || registry == nil {
		return nil, fmt.Errorf("Polish target pipeline requires a snapshot and verified language registry")
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
		Language: "pl", LanguageRegistry: registry, Lexicon: snapshot.Lexicon(), Valency: valencyIndex,
		Constructions: constructions, Variants: variants, Core: snapshot.CoreGraph(),
		OperatorBuilder: runtime.NoOperatorEvidenceBuilder{}, ExecutionBuilder: execution,
	}, nil
}

type polishSentenceFeatures struct{}

func (polishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.Category == construction.CategoryVP {
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		if !found {
			return nil, nil
		}
		if values["tense"] == "past" || values["verb_form"] == "past" {
			gender, known := polishSubjectGender(source, predicate, candidates)
			if !known {
				return nil, nil
			}
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "past"},
				[2]string{"person", "third"}, [2]string{"number", "singular"},
				[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
				[2]string{"gender", gender},
			), nil
		}
		presentSingular := values["verb_form"] == "third_person_singular" ||
			values["verb_form"] == "finite" && values["tense"] == "present" && values["person"] == "third" && values["number"] == "singular"
		if presentSingular {
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "present"},
				[2]string{"person", "third"}, [2]string{"number", "singular"},
				[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
			), nil
		}
		return nil, nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil || slot.Case == "" {
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
	pluralOnly := candidate.Lexeme.ParadigmID == "pl_noun_plural_only_attested"
	if pluralOnly {
		number = "plural"
	}
	if number != "singular" && number != "plural" {
		return nil, nil
	}
	gender := ""
	for _, property := range candidate.Lexeme.Properties {
		if property.Key == "gender" {
			gender = property.Value
			break
		}
	}
	if gender == "" && !pluralOnly {
		return nil, nil
	}
	featureValues := [][2]string{{"case", slot.Case}, {"number", string(number)}}
	if gender != "" {
		featureValues = append(featureValues, [2]string{"gender", gender})
	}
	return knownPolishFeatures(featureValues...), nil
}

func polishSubjectGender(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, bool) {
	if source.Meaning == nil {
		return "", false
	}
	var subjectRole string
	for _, argument := range predicate.Frame.Arguments {
		if argument.GrammaticalFunction != "subject" {
			continue
		}
		if subjectRole != "" {
			return "", false
		}
		subjectRole = string(argument.Role)
	}
	if subjectRole == "" {
		return "", false
	}
	meaningPredicate := findSentencePredicate(*source.Meaning, predicate.NodeID)
	if meaningPredicate.ID == "" {
		return "", false
	}
	var targetNode meaning.NodeID
	for _, participant := range meaningPredicate.Participants {
		if string(participant.Role) != subjectRole {
			continue
		}
		if targetNode != "" {
			return "", false
		}
		targetNode = participant.Target
	}
	if targetNode == "" {
		return "", false
	}
	gender := ""
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != targetNode || candidate.Outcome != status.OutcomeOK {
			continue
		}
		for _, property := range candidate.Lexeme.Properties {
			if property.Key != "gender" {
				continue
			}
			if gender != "" && gender != property.Value {
				return "", false
			}
			gender = property.Value
		}
	}
	return gender, gender != ""
}

func findSentencePredicate(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id && (node.Kind == meaning.KindEvent || node.Kind == meaning.KindState) {
			return node
		}
	}
	return meaning.Node{}
}

func knownPolishFeatures(values ...[2]string) []features.FeatureValue {
	result := make([]features.FeatureValue, 0, len(values))
	for _, value := range values {
		result = append(result, features.FeatureValue{Feature: features.FeatureID(value[0]), State: features.StateKnown, Value: features.FeatureValueID(value[1])})
	}
	return result
}

var _ runtime.SlotFeaturePlanner = polishSentenceFeatures{}

package pl

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
		OperatorBuilder: polishSentenceOperatorBuilder{}, ExecutionBuilder: execution,
		CandidatePolicy: polishSentenceCandidatePolicy{},
		PlanBuilder:     polishSentencePlanAnnotations{},
	}, nil
}

type polishSentencePlanAnnotations struct{}

func (polishSentencePlanAnnotations) Annotations(_ context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error) {
	if len(source.Evidence.Records) != 0 || source.Meaning == nil || selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState {
		return nil, nil, nil, nil
	}
	predicatePath, found := runtime.MeaningPath(*source.Meaning, selected.NodeID)
	if !found {
		return nil, nil, nil, nil
	}
	assumptions := []realizer.FieldAssumption{
		{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning v1 carries no tense; this response uses the approved present-tense route"},
		{Path: predicatePath, Field: "person", Value: "third", Reason: "Meaning v1 carries no grammatical person; this response uses the approved third-person route"},
		{Path: predicatePath, Field: "number", Value: "singular", Reason: "Meaning v1 carries no grammatical number; this response uses the approved singular route"},
	}
	for _, participant := range findSentencePredicate(*source.Meaning, selected.NodeID).Participants {
		path, ok := runtime.MeaningPath(*source.Meaning, participant.Target)
		if ok {
			assumptions = append(assumptions, realizer.FieldAssumption{
				Path: path, Field: "number", Value: "lexeme_default",
				Reason: "Meaning v1 carries no noun number feature; this response uses the target lexeme's attested number form",
			})
		}
	}
	return nil, assumptions, nil, nil
}

type polishSentenceFeatures struct{}

func (polishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.Category == construction.CategoryVP {
		if _, event, found := polishNegatedSeeing(source); found && event.ID == predicate.NodeID && predicate.Sense.ID.String() == "sense:pl.widziec.visual" {
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "present"}, [2]string{"aspect", "imperfective"},
				[2]string{"person", "third"}, [2]string{"number", "singular"}, [2]string{"mood", "indicative"}, [2]string{"voice", "active"},
			), nil
		}
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		if !found {
			if len(source.Evidence.Records) == 0 {
				return knownPolishFeatures(
					[2]string{"verb_form", "finite"}, [2]string{"tense", "present"}, [2]string{"aspect", "imperfective"},
					[2]string{"person", "third"}, [2]string{"number", "singular"},
					[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
				), nil
			}
			return nil, nil
		}
		if values["tense"] == "past" || values["verb_form"] == "past" {
			number := string(values["number"])
			if number == "" {
				number = string(values["verb_number"])
			}
			subjectNumber, subjectNumberKnown := polishSubjectNumber(source, predicate, candidates)
			if number == "" && subjectNumberKnown {
				number = subjectNumber
			} else if number != "" && subjectNumberKnown && number != subjectNumber {
				return nil, nil
			}
			if number != "singular" && number != "plural" {
				return nil, nil
			}
			gender, known := polishSubjectPastGender(source, predicate, candidates, number)
			if !known {
				return nil, nil
			}
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "past"},
				[2]string{"aspect", "perfective"},
				[2]string{"person", "third"}, [2]string{"number", number},
				[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
				[2]string{"gender", gender},
			), nil
		}
		number := values["number"]
		if number == "" {
			number = values["verb_number"]
		}
		presentSingular := values["verb_form"] == "third_person_singular" && (number == "" || number == "singular") ||
			values["verb_form"] == "finite" && values["tense"] == "present" && values["person"] == "third" && number == "singular"
		presentPlural := values["verb_form"] == "finite" && values["tense"] == "present" && values["person"] == "third" && number == "plural"
		if presentSingular || presentPlural {
			finiteNumber := number
			if presentSingular && finiteNumber == "" {
				finiteNumber = "singular"
			}
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "present"},
				[2]string{"aspect", "imperfective"},
				[2]string{"person", "third"}, [2]string{"number", string(finiteNumber)},
				[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
			), nil
		}
		return nil, nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil || slot.Case == "" {
		return nil, nil
	}
	values, found := runtime.SourceFeatures(source, slot.ArgumentBinding.SemanticRole, candidate.TransferBinding.CoreRef)
	if !found && len(source.Evidence.Records) != 0 {
		return nil, nil
	}
	number := values["number"]
	if number == "" {
		number = values["noun_number"]
	}
	if !found {
		number = "singular"
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

type polishSentenceOperatorBuilder struct{}

func (polishSentenceOperatorBuilder) Build(ctx context.Context, source semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
	if ctx == nil {
		return nil, false, fmt.Errorf("Polish operator evidence requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	scope, event, found := polishNegatedSeeing(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyID, err := ids.NewConstructionID("pl.transitive_negative")
	if err != nil {
		return nil, false, err
	}
	return []construction.Operator{{
		ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
		InputFamilyID: familyID, OutputFamilyID: familyID,
		SemanticScope: construction.SemanticScope{TargetNodeID: event.ID, ScopeNodeID: scope.ID},
		Capabilities:  []construction.Capability{"negation-negated"}, Provenance: polishNegationConstructionSource,
	}}, true, nil
}

func polishNegatedSeeing(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, meaning.Node{}, false
	}
	var scope meaning.Node
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindScope {
			continue
		}
		if scope.ID != "" || node.Operator != meaning.OperatorNegation {
			return meaning.Node{}, meaning.Node{}, false
		}
		scope = node
	}
	if scope.ID == "" {
		return meaning.Node{}, meaning.Node{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID == scope.Target && node.Kind == meaning.KindEvent && node.Concept == "concept:seeing" {
			return scope, node, true
		}
	}
	return meaning.Node{}, meaning.Node{}, false
}

type polishSentenceCandidatePolicy struct{}

func (polishSentenceCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, _ []realizer.TargetCandidate) bool {
	if candidate.Family.Family.ID.String() != "construction:pl.transitive_negative" {
		return true
	}
	_, _, negated := polishNegatedSeeing(source)
	return negated
}

func polishSubjectGender(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, bool) {
	_, targetNode, found := polishSubjectTarget(source, predicate)
	if !found {
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

func polishSubjectNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, bool) {
	role, targetNode, found := polishSubjectTarget(source, predicate)
	if !found {
		return "", false
	}
	number := ""
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != targetNode || candidate.Outcome != status.OutcomeOK {
			continue
		}
		values, found := runtime.SourceFeatures(source, role, candidate.TransferBinding.CoreRef)
		if !found {
			continue
		}
		candidateNumber := string(values["number"])
		if candidateNumber == "" {
			candidateNumber = string(values["noun_number"])
		}
		if candidateNumber != "singular" && candidateNumber != "plural" {
			continue
		}
		if number != "" && number != candidateNumber {
			return "", false
		}
		number = candidateNumber
	}
	return number, number != ""
}

func polishSubjectPastGender(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate, number string) (string, bool) {
	if number == "plural" {
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		if found {
			genderClass := string(values["gender"])
			if genderClass == "masculine_personal" || genderClass == "non_masculine_personal" {
				return genderClass, true
			}
		}
	}
	gender, found := polishSubjectGender(source, predicate, candidates)
	if !found {
		return "", false
	}
	if number == "singular" {
		switch gender {
		case "masculine_personal", "feminine", "neuter":
			return gender, true
		default:
			return "", false
		}
	}
	if number == "plural" {
		switch gender {
		case "masculine_personal":
			return "masculine_personal", true
		case "masculine_animal", "masculine_inanimate", "feminine", "neuter":
			return "non_masculine_personal", true
		default:
			return "", false
		}
	}
	return "", false
}

func polishSubjectTarget(source semantic.MeaningCandidate, predicate realizer.TargetCandidate) (meaning.Role, meaning.NodeID, bool) {
	if source.Meaning == nil {
		return "", "", false
	}
	var subjectRole meaning.Role
	for _, argument := range predicate.Frame.Arguments {
		if argument.GrammaticalFunction != "subject" {
			continue
		}
		if subjectRole != "" {
			return "", "", false
		}
		subjectRole = argument.Role
	}
	if subjectRole == "" {
		return "", "", false
	}
	meaningPredicate := findSentencePredicate(*source.Meaning, predicate.NodeID)
	if meaningPredicate.ID == "" {
		return "", "", false
	}
	var targetNode meaning.NodeID
	for _, participant := range meaningPredicate.Participants {
		if participant.Role != subjectRole {
			continue
		}
		if targetNode != "" {
			return "", "", false
		}
		targetNode = participant.Target
	}
	return subjectRole, targetNode, targetNode != ""
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

package en

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
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
		PlanBuilder: englishSentencePlanAnnotations{}, CandidatePolicy: englishReceivingCandidatePolicy{},
	}, nil
}

type englishSentenceFeatures struct{}

func (englishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, _ []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.Category == construction.CategoryVP {
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		if !found {
			if len(source.Evidence.Records) == 0 {
				return englishFiniteVerbFeatures("present", "third", "singular"), nil
			}
			return nil, nil
		}
		if values["verb_form"] == "past" || values["tense"] == "past" {
			return englishFiniteVerbFeatures("past", "", ""), nil
		}
		if values["verb_form"] == "third_person_singular" {
			return englishFiniteVerbFeatures("present", "third", "singular"), nil
		}
		if values["verb_form"] == "finite" && values["tense"] == "present" {
			number := values["verb_number"]
			if number == "" {
				number = values["number"]
			}
			if (values["person"] == "first" || values["person"] == "second" || values["person"] == "third") &&
				(number == "singular" || number == "plural") {
				return englishFiniteVerbFeatures("present", string(values["person"]), string(number)), nil
			}
		}
		return nil, nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil {
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

func englishFiniteVerbFeatures(tense, person, number string) []features.FeatureValue {
	values := []features.FeatureValue{
		{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
		{Feature: "tense", State: features.StateKnown, Value: features.FeatureValueID(tense)},
		{Feature: "mood", State: features.StateKnown, Value: "indicative"},
		{Feature: "voice", State: features.StateKnown, Value: "active"},
	}
	if tense == "present" {
		values = append(values,
			features.FeatureValue{Feature: "person", State: features.StateKnown, Value: features.FeatureValueID(person)},
			features.FeatureValue{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
		)
	}
	return values
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
	assumptions := make([]realizer.FieldAssumption, 0, 3+2*len(predicate.Participants))
	losses := make([]realizer.FieldLoss, 0, len(predicate.Participants))
	predicatePath, pathFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
	if !pathFound {
		return nil, nil, nil, nil
	}
	if len(source.Evidence.Records) == 0 {
		assumptions = append(assumptions,
			realizer.FieldAssumption{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning v1 carries no tense; this response uses the approved present-tense route"},
			realizer.FieldAssumption{Path: predicatePath, Field: "person", Value: "third", Reason: "Meaning v1 carries no grammatical person; this response uses the approved third-person route"},
			realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "singular", Reason: "Meaning v1 carries no grammatical number; this response uses the approved singular route"},
		)
	}
	for _, participant := range predicate.Participants {
		referent := findMeaningNode(*source.Meaning, participant.Target)
		if referent.ID == "" {
			continue
		}
		referentPath, pathFound := runtime.MeaningPath(*source.Meaning, participant.Target)
		if !pathFound {
			continue
		}
		values, found := runtime.SourceFeatures(source, participant.Role, coreConcept(referent))
		if found && values["determiner"] == "definite" {
			continue
		}
		if found && values["determiner"] == "indefinite" {
			if englishIndefiniteObjectRole(selected, participant.Role) {
				continue
			}
			losses = append(losses, realizer.FieldLoss{
				Path: referentPath, Field: "determiner", Outcome: status.OutcomeUnsupported,
				Reason: "the selected English target construction cannot preserve this explicit indefinite source determiner",
			})
			continue
		}
		if englishMassObjectRole(selected, participant.Role) {
			assumptions = append(assumptions, realizer.FieldAssumption{
				Path: referentPath, Field: "determiner", Value: "bare",
				Reason: "The selected English construction realizes an explicitly mass target noun without an article; source Meaning does not encode definiteness",
			})
			if len(source.Evidence.Records) == 0 {
				assumptions = append(assumptions, realizer.FieldAssumption{
					Path: referentPath, Field: "number", Value: "singular",
					Reason: "Meaning v1 identifies individual referents but carries no noun number feature; this response uses the approved singular form",
				})
			}
			continue
		}
		determiner, reason := "definite", "English transitive template realizes the definite determiner; source Meaning and morphology provide no explicit definiteness value"
		if englishIndefiniteObjectRole(selected, participant.Role) {
			determiner, reason = "indefinite", "English receiving construction offers an indefinite count-object alternative because source Meaning does not encode definiteness"
		}
		assumptions = append(assumptions, realizer.FieldAssumption{
			Path: referentPath, Field: "determiner", Value: determiner, Reason: reason,
		})
		if len(source.Evidence.Records) == 0 {
			assumptions = append(assumptions, realizer.FieldAssumption{
				Path: referentPath, Field: "number", Value: "singular",
				Reason: "Meaning v1 identifies individual referents but carries no noun number feature; this response uses the approved singular form",
			})
		}
	}
	return nil, assumptions, losses, nil
}

type englishReceivingCandidatePolicy struct{}

func (englishReceivingCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	const (
		receivingSense          = "sense:en.get.receive"
		transitiveFamily        = "construction:en.transitive"
		standardVariant         = "construction:en.transitive.svo"
		massObjectVariant       = "construction:en.transitive.svo.mass_object"
		indefiniteObjectVariant = "construction:en.transitive.svo.indefinite_object"
	)
	if candidate.Kind != meaning.KindEvent && candidate.Kind != meaning.KindState || candidate.Sense.ID.String() != receivingSense ||
		candidate.Family.Family.ID.String() != transitiveFamily {
		return true
	}
	variantID := candidate.CompiledVariant.Variant.ID.String()
	if variantID != standardVariant && variantID != massObjectVariant && variantID != indefiniteObjectVariant {
		return true
	}
	if source.Meaning == nil {
		return variantID == standardVariant
	}
	predicate := findMeaningNode(*source.Meaning, candidate.NodeID)
	if predicate.ID == "" {
		return variantID == standardVariant
	}
	var objectTarget meaning.NodeID
	var objectRole meaning.Role
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.ID != "object" || slot.ArgumentBinding == nil {
			continue
		}
		objectRole = slot.ArgumentBinding.SemanticRole
		for _, participant := range predicate.Participants {
			if participant.Role == objectRole {
				objectTarget = participant.Target
				break
			}
		}
		break
	}
	if objectTarget == "" {
		return variantID == standardVariant
	}
	var objectCandidate realizer.TargetCandidate
	count := 0
	for _, available := range candidates {
		if available.SourceCandidateID == candidate.SourceCandidateID && available.NodeID == objectTarget &&
			available.Kind == meaning.KindReferent && available.Outcome == status.OutcomeOK {
			objectCandidate = available
			count++
		}
	}
	if count != 1 {
		return variantID == standardVariant
	}
	isMassObject := count == 1 && hasLexicalProperty(objectCandidate.Lexeme.Properties, "number_behavior", "mass")
	if isMassObject {
		return variantID == massObjectVariant
	}
	values, found := runtime.SourceFeatures(source, objectRole, objectCandidate.TransferBinding.CoreRef)
	determiner := values["determiner"]
	number := values["number"]
	if number == "" {
		number = values["noun_number"]
	}
	if determiner == "definite" {
		return variantID == standardVariant
	}
	if determiner == "indefinite" {
		return number == "singular" && variantID == indefiniteObjectVariant
	}
	if number == "singular" || !found && len(source.Evidence.Records) == 0 {
		return variantID == standardVariant || variantID == indefiniteObjectVariant
	}
	return variantID == standardVariant
}

func englishMassObjectRole(candidate realizer.TargetCandidate, role meaning.Role) bool {
	if candidate.CompiledVariant.Variant.ID.String() != "construction:en.transitive.svo.mass_object" ||
		candidate.Sense.ID.String() != "sense:en.get.receive" {
		return false
	}
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.ID == "object" && slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == role {
			return true
		}
	}
	return false
}

func englishIndefiniteObjectRole(candidate realizer.TargetCandidate, role meaning.Role) bool {
	if candidate.CompiledVariant.Variant.ID.String() != "construction:en.transitive.svo.indefinite_object" ||
		candidate.Sense.ID.String() != "sense:en.get.receive" {
		return false
	}
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.ID == "object" && slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == role {
			return true
		}
	}
	return false
}

func hasLexicalProperty(properties []lexicon.Property, key, value string) bool {
	for _, property := range properties {
		if property.Key == key && property.Value == value {
			return true
		}
	}
	return false
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

package en

import (
	"context"
	"fmt"
	"strings"

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
		OperatorBuilder: englishSentenceOperatorBuilder{}, ExecutionBuilder: execution,
		PlanBuilder: englishSentencePlanAnnotations{}, CandidatePolicy: englishDeterminerCandidatePolicy{},
	}, nil
}

type englishSentenceFeatures struct{}

func (englishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, _ []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.Category == construction.CategoryVP {
		if _, event, route, found := englishNegatedPresent(source); found && event.ID == predicate.NodeID && predicate.Sense.ID.String() == route.sense {
			number, known := englishPresentNegationNumber(source, predicate, route)
			if !known {
				return nil, nil
			}
			return []features.FeatureValue{
				{Feature: "verb_form", State: features.StateKnown, Value: "imperative"},
				{Feature: "mood", State: features.StateKnown, Value: "imperative"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
				{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
			}, nil
		}
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
	return []features.FeatureValue{
		{Feature: "noun_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
		{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
	}, nil
}

func englishPresentNegationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route englishPresentNegationRoute) (string, bool) {
	values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
	if found {
		if person := values["person"]; person != "" && person != "third" {
			return "", false
		}
		if tense := values["tense"]; tense != "" && tense != "present" {
			return "", false
		}
		number := string(values["number"])
		if number == "" {
			number = string(values["verb_number"])
		}
		return number, number == "singular" || number == "plural"
	}
	if source.Meaning == nil {
		return "", false
	}
	event := findMeaningNode(*source.Meaning, predicate.NodeID)
	for _, participant := range event.Participants {
		if participant.Role != route.subjectRole {
			continue
		}
		referent := findMeaningNode(*source.Meaning, participant.Target)
		if referent.ID == "" {
			continue
		}
		roleValues, roleFound := runtime.SourceFeatures(source, route.subjectRole, coreConcept(referent))
		if !roleFound {
			continue
		}
		number := string(roleValues["number"])
		if number == "" {
			number = string(roleValues["noun_number"])
		}
		return number, number == "singular" || number == "plural"
	}
	if len(source.Evidence.Records) == 0 {
		return "singular", true
	}
	return "", false
}

type englishSentenceOperatorBuilder struct{}

func (englishSentenceOperatorBuilder) Build(ctx context.Context, source semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
	if ctx == nil {
		return nil, false, fmt.Errorf("English operator evidence requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	scope, event, route, found := englishNegatedPresent(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(route.family, "construction:"))
	if err != nil {
		return nil, false, err
	}
	return []construction.Operator{{
		ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
		InputFamilyID: familyID, OutputFamilyID: familyID,
		SemanticScope: construction.SemanticScope{TargetNodeID: event.ID, ScopeNodeID: scope.ID},
		Capabilities:  []construction.Capability{"negation-negated"}, Provenance: englishNegationConstructionSource,
	}}, true, nil
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
			features.FeatureValue{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
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
		article, targetDeterminerKnown := englishDeterminerForRole(selected, participant.Role)
		if !targetDeterminerKnown {
			continue
		}
		determiner := englishDeterminerFeature(article)
		sourceDeterminer := ""
		if found {
			sourceDeterminer = string(values["determiner"])
		}
		if sourceDeterminer != "" {
			if sourceDeterminer != determiner {
				losses = append(losses, realizer.FieldLoss{
					Path: referentPath, Field: "determiner", Outcome: status.OutcomeUnsupported,
					Reason: "the selected English target construction does not preserve the explicit source determiner",
				})
			}
			continue
		}
		reason := "the source does not encode English definiteness; this candidate selects the definite article"
		switch determiner {
		case "indefinite":
			reason = "the source does not encode English definiteness; this candidate selects the indefinite article " + article
		case "bare":
			reason = "the target noun is explicitly mass and this candidate realizes it without an article"
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

// englishDeterminerCandidatePolicy retains only variants supported by source
// determiner/number evidence and explicit target-lexeme article properties.
// Missing source definiteness keeps all compatible options visible.
type englishDeterminerCandidatePolicy struct{}

func (englishDeterminerCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	const (
		transitiveFamily     = "construction:en.transitive"
		transitivePastFamily = "construction:en.transitive_past"
		obliqueFamily        = "construction:en.oblique"
	)
	familyID := candidate.Family.Family.ID.String()
	negativeRoute, isNegativeFamily := englishPresentNegationRouteForFamily(candidate.Family.Family.ID)
	if isNegativeFamily {
		_, event, sourceRoute, negated := englishNegatedPresent(source)
		if !negated || event.ID != candidate.NodeID || sourceRoute.family != negativeRoute.family || sourceRoute.sense != candidate.Sense.ID.String() {
			return false
		}
		number, known := englishPresentNegationNumber(source, candidate, sourceRoute)
		if !known || candidate.CompiledVariant.Variant.ID.String() != sourceRoute.variantForNumber(number) {
			return false
		}
	}
	if candidate.Kind != meaning.KindEvent && candidate.Kind != meaning.KindState || familyID != transitiveFamily && familyID != transitivePastFamily && familyID != obliqueFamily && !isNegativeFamily {
		return true
	}
	variant, knownVariant := englishDeterminerVariant(candidate.CompiledVariant.Variant.ID.String())
	if !knownVariant {
		return true
	}
	if source.Meaning == nil {
		return candidate.CompiledVariant.Variant.ID.String() == englishStandardDeterminerVariant(familyID)
	}
	predicate := findMeaningNode(*source.Meaning, candidate.NodeID)
	if predicate.ID == "" {
		return candidate.CompiledVariant.Variant.ID.String() == englishStandardDeterminerVariant(familyID)
	}
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if (slot.ID != "subject" && slot.ID != "object" && slot.ID != "target") || slot.ArgumentBinding == nil {
			continue
		}
		role := slot.ArgumentBinding.SemanticRole
		var referent meaning.Node
		foundReferent := false
		for _, participant := range predicate.Participants {
			if participant.Role == role {
				referent = findMeaningNode(*source.Meaning, participant.Target)
				foundReferent = referent.ID != ""
				break
			}
		}
		if !foundReferent {
			if englishDeterminerFeature(englishDeterminerForSlot(variant, slot.ID)) != "definite" {
				return false
			}
			continue
		}
		lexicalCandidates := make([]realizer.TargetCandidate, 0, 1)
		for _, available := range candidates {
			if available.SourceCandidateID == candidate.SourceCandidateID && available.NodeID == referent.ID &&
				available.Kind == meaning.KindReferent && (available.Outcome == status.OutcomeOK || available.Outcome == status.OutcomeIncomplete) &&
				available.Lexeme.ID.String() != "" {
				lexicalCandidates = append(lexicalCandidates, available)
			}
		}
		if len(lexicalCandidates) == 0 {
			if englishDeterminerFeature(englishDeterminerForSlot(variant, slot.ID)) != "definite" {
				return false
			}
			continue
		}
		values, found := runtime.SourceFeatures(source, role, coreConcept(referent))
		determiner := ""
		if found {
			determiner = string(values["determiner"])
		}
		number := values["number"]
		if number == "" {
			number = values["noun_number"]
		}
		allMass := true
		for _, lexicalCandidate := range lexicalCandidates {
			allMass = allMass && hasLexicalProperty(lexicalCandidate.Lexeme.Properties, "number_behavior", "mass")
		}
		indefiniteArticle, hasCommonIndefiniteArticle := commonEnglishIndefiniteArticle(lexicalCandidates)
		mass := allMass
		if mass {
			number = "singular"
		}
		if number == "" && !found && len(source.Evidence.Records) == 0 {
			number = "singular"
		}
		actualArticle := englishDeterminerForSlot(variant, slot.ID)
		actualDeterminer := englishDeterminerFeature(actualArticle)
		if slot.ID == "object" && mass {
			if determiner == "definite" {
				if actualDeterminer != "definite" {
					return false
				}
			} else if actualDeterminer != "bare" {
				return false
			}
			continue
		}
		if actualDeterminer == "indefinite" && (!hasCommonIndefiniteArticle || actualArticle != indefiniteArticle) {
			return false
		}
		if determiner != "" {
			if actualDeterminer != determiner {
				return false
			}
			continue
		}
		if number == "plural" {
			if actualDeterminer != "definite" {
				return false
			}
			continue
		}
		if number != "singular" {
			if actualDeterminer != "definite" {
				return false
			}
			continue
		}
		if actualDeterminer != "definite" && actualDeterminer != "indefinite" {
			return false
		}
	}
	return true
}

type englishDeterminers struct {
	subject string
	object  string
	target  string
}

func englishDeterminerVariant(id string) (englishDeterminers, bool) {
	variant, exists := englishDeterminerVariants[id]
	const pastPrefix = "construction:en.transitive_past."
	if !exists && len(id) > len(pastPrefix) && id[:len(pastPrefix)] == pastPrefix {
		variant, exists = englishDeterminerVariants["construction:en.transitive."+id[len(pastPrefix):]]
	}
	return variant, exists
}

var englishDeterminerVariants = map[string]englishDeterminers{
	"construction:en.transitive.svo":                                   {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo":                          {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.plural":                   {subject: "the", object: "the"},
	"construction:en.transitive.svo.indefinite_object":                 {subject: "the", object: "a"},
	"construction:en.transitive.svo.indefinite_subject":                {subject: "a", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_object":         {subject: "a", object: "a"},
	"construction:en.transitive.svo.mass_object":                       {subject: "the", object: "bare"},
	"construction:en.transitive.svo.indefinite_subject_mass_object":    {subject: "a", object: "bare"},
	"construction:en.transitive.svo.indefinite_object_an":              {subject: "the", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an":             {subject: "an", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_a_object_an":    {subject: "a", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_object_a":    {subject: "an", object: "a"},
	"construction:en.transitive.svo.indefinite_subject_an_object_an":   {subject: "an", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_mass_object": {subject: "an", object: "bare"},
	"construction:en.oblique.svo":                                      {subject: "the", target: "the"},
	"construction:en.oblique.svo.indefinite_target":                    {subject: "the", target: "a"},
	"construction:en.oblique.svo.indefinite_subject":                   {subject: "a", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_target":            {subject: "a", target: "a"},
	"construction:en.oblique.svo.indefinite_target_an":                 {subject: "the", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an":                {subject: "an", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_a_target_an":       {subject: "a", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an_target_a":       {subject: "an", target: "a"},
	"construction:en.oblique.svo.indefinite_subject_an_target_an":      {subject: "an", target: "an"},
	"construction:en.oblique_negative.svo":                             {subject: "the", target: "the"},
	"construction:en.oblique_negative.svo.plural":                      {subject: "the", target: "the"},
}

func englishStandardDeterminerVariant(familyID string) string {
	switch familyID {
	case "construction:en.transitive":
		return "construction:en.transitive.svo"
	case "construction:en.transitive_past":
		return "construction:en.transitive_past.svo"
	case "construction:en.transitive_negative":
		return "construction:en.transitive_negative.svo"
	case "construction:en.oblique":
		return "construction:en.oblique.svo"
	case "construction:en.oblique_negative":
		return "construction:en.oblique_negative.svo"
	default:
		return ""
	}
}

func englishDeterminerForRole(candidate realizer.TargetCandidate, role meaning.Role) (string, bool) {
	familyID := candidate.Family.Family.ID.String()
	if familyID != "construction:en.transitive" && familyID != "construction:en.transitive_past" && familyID != "construction:en.transitive_negative" && familyID != "construction:en.oblique" && familyID != "construction:en.oblique_negative" {
		return "", false
	}
	variant, exists := englishDeterminerVariant(candidate.CompiledVariant.Variant.ID.String())
	if !exists {
		return "", false
	}
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.ArgumentBinding == nil || slot.ArgumentBinding.SemanticRole != role {
			continue
		}
		return englishDeterminerForSlot(variant, slot.ID), true
	}
	return "", false
}

func englishDeterminerForSlot(variant englishDeterminers, slotID string) string {
	switch slotID {
	case "subject":
		return variant.subject
	case "object":
		return variant.object
	case "target":
		return variant.target
	default:
		return ""
	}
}

func englishDeterminerFeature(article string) string {
	switch article {
	case "the":
		return "definite"
	case "a", "an":
		return "indefinite"
	case "bare":
		return "bare"
	default:
		return ""
	}
}

func commonEnglishIndefiniteArticle(candidates []realizer.TargetCandidate) (string, bool) {
	article := ""
	for _, candidate := range candidates {
		candidateArticle, exists := englishIndefiniteArticle(candidate.Lexeme.Properties)
		if !exists {
			return "", false
		}
		if article != "" && article != candidateArticle {
			return "", false
		}
		article = candidateArticle
	}
	return article, article != ""
}

func englishIndefiniteArticle(properties []lexicon.Property) (string, bool) {
	for _, property := range properties {
		if property.Key == "indefinite_article" && (property.Value == "a" || property.Value == "an") {
			return property.Value, true
		}
	}
	return "", false
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

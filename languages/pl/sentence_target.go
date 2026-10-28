package pl

import (
	"context"
	"fmt"
	"strings"

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
	return NewSentenceExecutionBuilderWithSeed(seed)
}

// NewSentenceExecutionBuilderWithSeed compiles target morphology from the
// caller's exact Polish seed.
func NewSentenceExecutionBuilderWithSeed(seed *Seed) (runtime.MorphologyExecutionBuilder, error) {
	if seed == nil {
		return runtime.MorphologyExecutionBuilder{}, fmt.Errorf("Polish execution builder requires a Polish morphology seed")
	}
	profile, _, paradigms, exceptions := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "pl" || paradigms == nil || exceptions == nil {
		return runtime.MorphologyExecutionBuilder{}, fmt.Errorf("Polish execution builder requires a Polish morphology seed")
	}
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
	return NewSentenceTargetWithSeed(snapshot, registry, seed)
}

// NewSentenceTargetWithSeed builds all Polish target generators from one
// caller-supplied morphology seed.
func NewSentenceTargetWithSeed(snapshot *pack.Pack, registry *language.Registry, seed *Seed) (*runtime.TargetPipeline, error) {
	if snapshot == nil || registry == nil {
		return nil, fmt.Errorf("Polish target pipeline requires a snapshot and verified language registry")
	}
	if seed == nil {
		return nil, fmt.Errorf("Polish target pipeline requires a Polish morphology seed")
	}
	profile, _, _, _ := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "pl" {
		return nil, fmt.Errorf("Polish target pipeline requires a Polish morphology seed")
	}
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
	execution, err := NewSentenceExecutionBuilderWithSeed(seed)
	if err != nil {
		return nil, err
	}
	coordination, err := SentenceCoordination()
	if err != nil {
		return nil, err
	}
	scopePlan, err := SentenceScopedClauses()
	if err != nil {
		return nil, err
	}
	relationPlan, err := SentenceRelationPathPlan()
	if err != nil {
		return nil, err
	}
	return &runtime.TargetPipeline{
		Language: "pl", LanguageRegistry: registry, Lexicon: snapshot.Lexicon(), Valency: valencyIndex,
		Constructions: constructions, Variants: variants, Core: snapshot.CoreGraph(),
		OperatorBuilder: polishSentenceOperatorBuilder{}, ExecutionBuilder: execution,
		CandidatePolicy:      polishSentenceCandidatePolicy{},
		PlanBuilder:          polishSentencePlanAnnotations{},
		ClauseCoordination:   &coordination,
		ScopedRealizer:       &runtime.ScopedClauseRealizer{Language: "pl", Plan: scopePlan, Valency: valencyIndex, Morphology: seed, Features: polishScopedFeaturePlanner{}},
		RelationPathRealizer: &runtime.RelationPathRealizer{Language: "pl", Plan: relationPlan, Morphology: seed, Features: polishRelationPathFeaturePlanner{}},
	}, nil
}

type polishSentencePlanAnnotations struct{}

func (polishSentencePlanAnnotations) Annotations(_ context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error) {
	if selected.Family.Family.ID.String() == polishPersonalTargetFamily {
		target, ok := polishPersonalTargetForPredicate(source, selected.NodeID)
		if !ok || !polishPersonalVariantMatches(source, selected, target) || !polishPredicateEvidenceConsistent(source, selected) {
			return nil, nil, nil, fmt.Errorf("Polish personal target requires a valid anchored subject and matching licensed variant")
		}
		return nil, polishPersonalVariantAssumptions(source, selected), nil, nil
	}
	if _, referentID, negated := polishNegatedClassificationState(source, selected.NodeID); negated {
		predicatePath, predicateFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
		referentPath, referentFound := runtime.MeaningPath(*source.Meaning, referentID)
		mode, contextOK := polishNegatedClassificationReferenceMode(source, selected.NodeID)
		predicate := findSentencePredicate(*source.Meaning, selected.NodeID)
		if !predicateFound || !referentFound || !contextOK ||
			!polishIdentificationAttributesAllowed(source, selected.NodeID, "member", referentID, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("Polish negative identification realization requires a valid unresolved or proximal member reference context")
		}
		assumptions := polishIdentificationDefaults(predicatePath, referentPath, predicate, source.Meaning.Nodes, referentID, "negative identification")
		var losses []realizer.FieldLoss
		if mode == "proximal" || mode == "distal" {
			losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "deixis", Outcome: status.OutcomeIncomplete, Reason: "Polish to nie preserves demonstrative reference but does not encode the source's explicit distance"})
		}
		return nil, assumptions, losses, nil
	}
	if role, referentID, _, ok := polishIdentificationState(source, selected.NodeID); ok {
		predicatePath, predicateFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
		referentPath, referentFound := runtime.MeaningPath(*source.Meaning, referentID)
		mode, contextOK := polishIdentificationReferenceMode(source, selected.NodeID)
		predicate := findSentencePredicate(*source.Meaning, selected.NodeID)
		if !predicateFound || !referentFound || !contextOK ||
			!polishIdentificationAttributesAllowed(source, selected.NodeID, role, referentID, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("Polish identification realization requires a valid unresolved %s reference context", role)
		}
		assumptions := polishIdentificationDefaults(predicatePath, referentPath, predicate, source.Meaning.Nodes, referentID, "identification")
		var losses []realizer.FieldLoss
		if mode == "proximal" || mode == "distal" {
			losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "deixis", Outcome: status.OutcomeIncomplete, Reason: "Polish to/ten preserves demonstrative reference but does not encode the source's explicit distance"})
		}
		return nil, assumptions, losses, nil
	}
	if source.Meaning == nil || selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState {
		return nil, nil, nil, nil
	}
	predicate, found := polishMeaningNode(source, selected.NodeID)
	predicatePath, pathFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
	if !found || !pathFound {
		return nil, nil, nil, nil
	}
	if !polishPredicateEvidenceConsistent(source, selected) {
		return nil, nil, nil, fmt.Errorf("Polish target realization has Meaning attributes that conflict with explicit source evidence")
	}
	assumptions := []realizer.FieldAssumption{}
	if selected.TransferBinding.SenseID.String() == "sense:pl.isc.move" {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "manner", Value: "on_foot", Reason: "Polish iść specifies movement on foot; the source meaning does not specify manner"})
	}
	if predicate.TemporalFrame == "" {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning carries no temporal frame; this response uses the approved present-tense route"})
	}
	_, subjectID, hasSubject := polishSubjectTarget(source, selected)
	if hasSubject {
		if subject, exists := polishMeaningNode(source, subjectID); exists && subject.Multiplicity == "" {
			assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "lexeme_default", Reason: "Meaning carries no subject multiplicity; predicate agreement follows the target subject's local lexical class"})
		}
	}
	for _, participant := range predicate.Participants {
		referent, exists := polishMeaningNode(source, participant.Target)
		if !exists || referent.Kind != meaning.KindReferent || referent.Multiplicity != "" {
			continue
		}
		path, ok := runtime.MeaningPath(*source.Meaning, participant.Target)
		if ok {
			assumptions = append(assumptions, realizer.FieldAssumption{
				Path: path, Field: "number", Value: "lexeme_default",
				Reason: "Meaning carries no referent multiplicity; this response uses the target lexeme's local number class",
			})
		}
	}
	return nil, assumptions, nil, nil
}

func polishIdentificationDefaults(predicatePath, referentPath []meaning.NodeID, predicate meaning.Node, nodes []meaning.Node, referentID meaning.NodeID, route string) []realizer.FieldAssumption {
	assumptions := []realizer.FieldAssumption{}
	if predicate.TemporalFrame == "" {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning carries no temporal frame; this " + route + " route uses the approved present form"})
	}
	for _, node := range nodes {
		if node.ID == referentID && node.Multiplicity == "" {
			assumptions = append(assumptions,
				realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "singular", Reason: "Meaning carries no referent multiplicity; this licensed identification route uses the singular count-noun form"},
				realizer.FieldAssumption{Path: referentPath, Field: "number", Value: "singular", Reason: "Meaning carries no referent multiplicity; this licensed identification route uses the singular count-noun form"},
			)
			break
		}
	}
	return assumptions
}

type polishSentenceFeatures struct{}

func (polishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if answer := predicate.QuestionAnswer; answer != nil && answer.TargetNodeID == predicate.NodeID && slot.Category == construction.CategoryVP {
		if !polishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("Polish WH target conflicts with explicit event frame evidence")
		}
		predicateMeaning, exists := polishMeaningNode(source, predicate.NodeID)
		if !exists || predicateMeaning.TemporalFrame != "" && predicateMeaning.TemporalFrame != meaning.TemporalFramePresent {
			return nil, fmt.Errorf("Polish WH target supports only the present temporal frame")
		}
		number := "singular"
		if answer.Role == "stimulus" {
			var known bool
			number, known = polishPredicateNumber(source, predicate, candidates)
			if !known {
				return nil, fmt.Errorf("Polish WH object question requires explicit or lexeme-grounded subject number for agreement")
			}
		} else if answer.Role != "experiencer" {
			return nil, fmt.Errorf("Polish WH target does not support answer role %q", answer.Role)
		}
		return knownPolishFeatures(
			[2]string{"verb_form", "finite"}, [2]string{"tense", "present"},
			[2]string{"aspect", "imperfective"}, [2]string{"person", "third"},
			[2]string{"number", number}, [2]string{"mood", "indicative"},
			[2]string{"voice", "active"},
		), nil
	}
	if slot.ModifierBinding != nil && polishAttributedSeeingFamily(predicate) {
		shape, valid := polishAttributedSeeingShape(source)
		if !valid || source.Meaning == nil {
			return nil, fmt.Errorf("Polish attributed seeing adjective requires the licensed two-root Meaning")
		}
		modifier, modifierFound := realizer.FindModifier(*source.Meaning, predicate.NodeID, predicate.Family.Family, slot.ID)
		_, subjectID, subjectFound := polishSubjectTarget(source, predicate)
		if !polishAttributedSeeingContextsValid(source) || !modifierFound || modifier.ID != shape.property.ID ||
			!subjectFound || subjectID != shape.subject.ID || candidate.NodeID != modifier.ID ||
			candidate.Sense.ID.String() != "sense:pl.czarny.black" || candidate.TransferBinding.CoreRef.String() != "concept:black" || candidate.Lexeme.ID.String() != "lexeme:pl.czarny.adj" {
			return nil, fmt.Errorf("Polish attributed seeing adjective is not linked to the selected subject property")
		}
		gender, number, found := polishIdentificationReferentFeatures(source, subjectID, candidates)
		if !found {
			return nil, fmt.Errorf("Polish attributed seeing adjective has no unique compatible subject lexical class")
		}
		return knownPolishFeatures(
			[2]string{"degree", "positive"}, [2]string{"number", number},
			[2]string{"gender", gender}, [2]string{"case", "nominative"},
		), nil
	}
	if slot.Category == construction.CategoryAP {
		if predicate.Sense.ID.String() != "sense:pl.czarny.black" || candidate.Lexeme.ID.String() != "lexeme:pl.czarny.adj" {
			return nil, nil
		}
		_, referentID, _, found := polishIdentificationState(source, predicate.NodeID)
		if !found {
			return nil, nil
		}
		gender, number, found := polishIdentificationReferentFeatures(source, referentID, candidates)
		if !found {
			return nil, nil
		}
		return knownPolishFeatures(
			[2]string{"degree", "positive"}, [2]string{"number", number},
			[2]string{"gender", gender}, [2]string{"case", "nominative"},
		), nil
	}
	if slot.Category == construction.CategoryVP {
		if predicate.Family.Family.ID.String() == polishPersonalTargetFamily {
			target, ok := polishPersonalTargetForPredicate(source, predicate.NodeID)
			if !ok || !polishPersonalVariantMatches(source, predicate, target) ||
				!polishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
				return nil, fmt.Errorf("Polish personal predicate variant conflicts with participant, temporal, or valency evidence")
			}
			values := polishPersonalVariantFeatures(predicate.CompiledVariant.Variant)
			featureValues := [][2]string{
				{"verb_form", "finite"}, {"tense", values["tense"]}, {"aspect", "imperfective"},
				{"person", values["person"]}, {"number", values["number"]}, {"mood", "indicative"}, {"voice", "active"},
			}
			if values["tense"] == "past" {
				featureValues = append(featureValues, [2]string{"gender", values["gender"]})
			}
			return knownPolishFeatures(featureValues...), nil
		}
		if _, event, route, found := polishNegatedPresent(source); found && event.ID == predicate.NodeID && predicate.Sense.ID.String() == route.sense {
			number, known := polishPresentNegationNumber(source, predicate, route, candidates)
			if !known {
				return nil, fmt.Errorf("Polish negative predicate cannot realize its Meaning number or temporal frame")
			}
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "present"}, [2]string{"aspect", "imperfective"},
				[2]string{"person", "third"}, [2]string{"number", number}, [2]string{"mood", "indicative"}, [2]string{"voice", "active"},
			), nil
		}
		predicateMeaning, exists := polishMeaningNode(source, predicate.NodeID)
		if !exists {
			return nil, nil
		}
		if !polishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("Polish predicate temporal frame conflicts with explicit source evidence")
		}
		tense := predicateMeaning.TemporalFrame
		if tense == "" {
			tense = meaning.TemporalFramePresent
		}
		if tense == meaning.TemporalFramePast {
			number, known := polishPredicateNumber(source, predicate, candidates)
			if !known {
				return nil, fmt.Errorf("Polish past predicate cannot realize subject multiplicity")
			}
			gender, known := polishSubjectPastGender(source, predicate, candidates, number)
			if !known {
				return nil, fmt.Errorf("Polish past predicate cannot realize the selected subject's lexical agreement class")
			}
			aspect := "perfective"
			if predicate.Sense.ID.String() == "sense:pl.czytac.interpret" || predicate.Sense.ID.String() == "sense:pl.widziec.visual" {
				// The approved reading lexeme is imperfective; do not assign the
				// receiving route's perfective aspect to this distinct sense.
				aspect = "imperfective"
			}
			return knownPolishFeatures(
				[2]string{"verb_form", "finite"}, [2]string{"tense", "past"},
				[2]string{"aspect", aspect},
				[2]string{"person", "third"}, [2]string{"number", number},
				[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
				[2]string{"gender", gender},
			), nil
		}
		number, known := polishPredicateNumber(source, predicate, candidates)
		if !known {
			return nil, fmt.Errorf("Polish predicate cannot realize subject multiplicity")
		}
		return knownPolishFeatures(
			[2]string{"verb_form", "finite"}, [2]string{"tense", "present"},
			[2]string{"aspect", "imperfective"},
			[2]string{"person", "third"}, [2]string{"number", number},
			[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
		), nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil || slot.Case == "" {
		return nil, nil
	}
	number, meaningNumber, numberOK := polishTargetNumber(source, candidate.NodeID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
	if !numberOK {
		return nil, fmt.Errorf("Polish lexical number class cannot realize the Meaning multiplicity")
	}
	if meaningNumber {
		referent, exists := polishMeaningNode(source, candidate.NodeID)
		checkPredicateNumber := polishPredicateSubjectRole(predicate, slot.ArgumentBinding.SemanticRole)
		if !exists || !polishMultiplicityEvidenceConsistent(source, predicate.NodeID, slot.ArgumentBinding.SemanticRole, referent, predicate.TransferBinding.CoreRef, checkPredicateNumber) {
			return nil, fmt.Errorf("Polish referent multiplicity conflicts with explicit source evidence")
		}
	}
	pluralOnly := candidate.Lexeme.ParadigmID == "pl_noun_plural_only_attested"
	for _, property := range candidate.Lexeme.Properties {
		if property.Key == "grammatical_number" && property.Value == "plural_only" {
			pluralOnly = true
		}
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
	if intent, valid := polishPersonalContext(source); intent && !valid {
		return nil, false, polishParticipantContextError{}
	}
	if source.QuestionAnswer != nil {
		if source.Meaning == nil {
			return nil, false, fmt.Errorf("Polish WH operator requires its typed Meaning question")
		}
		var question meaning.Node
		questionCount := 0
		for _, node := range source.Meaning.Nodes {
			if node.Kind != meaning.KindQuestion {
				continue
			}
			question = node
			questionCount++
		}
		if questionCount != 1 || question.Target != source.QuestionAnswer.TargetNodeID {
			return nil, false, fmt.Errorf("Polish WH operator requires one question wrapper over the typed answer target")
		}
		familyID, err := ids.NewConstructionID("pl.question.wh")
		if err != nil {
			return nil, false, err
		}
		return []construction.Operator{{
			ID: "question-wh", Kind: construction.OperatorQuestion, Value: construction.ValueWH,
			InputFamilyID: familyID, OutputFamilyID: familyID,
			SemanticScope: construction.SemanticScope{TargetNodeID: source.QuestionAnswer.TargetNodeID, ScopeNodeID: question.ID},
			Capabilities:  []construction.Capability{"question-wh"}, Provenance: polishWHQuestionSource,
		}}, true, nil
	}
	scope, event, route, found := polishNegatedPresent(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(route.family, "construction:"))
	if err != nil {
		return nil, false, err
	}
	operatorProvenance := polishNegationConstructionSource
	if route.family == "construction:pl.identification_negative" {
		operatorProvenance = polishIdentificationNegationSource
	}
	return []construction.Operator{{
		ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
		InputFamilyID: familyID, OutputFamilyID: familyID,
		SemanticScope: construction.SemanticScope{TargetNodeID: event.ID, ScopeNodeID: scope.ID},
		Capabilities:  []construction.Capability{"negation-negated"}, Provenance: operatorProvenance,
	}}, true, nil
}

type polishSentenceCandidatePolicy struct{}

func (polishSentenceCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	anchorIntent, contextsValid := polishPersonalContext(source)
	if anchorIntent && !contextsValid {
		return false
	}
	if source.QuestionAnswer != nil && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) {
		return candidate.NodeID == source.QuestionAnswer.TargetNodeID && polishWHQuestionTargetCandidate(source, candidate)
	}
	familyID := candidate.Family.Family.ID.String()
	if familyID == "construction:pl.transitive_personal_explicit" {
		return false
	}
	if familyID == polishPersonalTargetFamily && !anchorIntent {
		return false
	}
	if !anchorIntent && isPolishPersonalPronoun(candidate.Lexeme.ID) {
		return false
	}
	if anchorIntent && candidate.Lexeme.POS == "pronoun" {
		return false // Polish personal subjects are pro-drop in this target construction.
	}
	if anchorIntent && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) {
		target, personalSubject := polishPersonalTargetForPredicate(source, candidate.NodeID)
		return personalSubject && candidate.Family.Family.ID.String() == polishPersonalTargetFamily &&
			polishPersonalVariantMatches(source, candidate, target) && polishPredicateEvidenceConsistent(source, candidate)
	}
	if anchorIntent && candidate.Kind == meaning.KindReferent {
		for _, predicate := range source.Meaning.Nodes {
			if predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
				continue
			}
			target, ok := polishPersonalTargetForPredicate(source, predicate.ID)
			if ok && target.subject.ID == candidate.NodeID {
				return false // never fall back to a generic person noun for an anchored, omitted subject.
			}
		}
	}
	if (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) && !polishPredicateEvidenceConsistent(source, candidate) {
		return false
	}
	if polishAttributedSeeingFamily(candidate) {
		return polishAttributedSeeingMainCandidate(source, candidate)
	}
	if shape, valid := polishAttributedSeeingShape(source); valid {
		if candidate.NodeID == shape.property.ID {
			return polishAttributedSeeingModifierCandidate(source, candidate, all)
		}
		if candidate.NodeID == shape.event.ID {
			return false
		}
	}
	if concept := polishIdentificationPredicateConcept(source, candidate.NodeID); concept != "" {
		if concept == "concept:class_membership" {
			if sourceHasScope(source) {
				scopeID, referentID, validGraph := polishNegatedClassificationState(source, candidate.NodeID)
				mode, validContext := polishNegatedClassificationReferenceMode(source, candidate.NodeID)
				return validGraph && validContext && (mode == "unresolved" || mode == "proximal" || mode == "distal") && scopeID != "" &&
					polishIdentificationAttributesAllowed(source, candidate.NodeID, "member", referentID, candidate.TransferBinding.CoreRef) &&
					candidate.Sense.ID.String() == "sense:pl.byc.class_membership" &&
					string(candidate.Frame.ID) == "pl.byc.class_membership.frame" &&
					candidate.Family.Family.ID.String() == "construction:pl.identification_negative" &&
					candidate.CompiledVariant.Variant.ID.String() == "construction:pl.identification_negative.to"
			}
			role, referentID, _, valid := polishIdentificationState(source, candidate.NodeID)
			if !valid || !polishIdentificationAttributesAllowed(source, candidate.NodeID, role, referentID, candidate.TransferBinding.CoreRef) {
				return false
			}
			mode, valid := polishIdentificationReferenceMode(source, candidate.NodeID)
			return valid && (mode == "proximal" || mode == "distal" || mode == "unresolved") &&
				candidate.Sense.ID.String() == "sense:pl.byc.class_membership" &&
				string(candidate.Frame.ID) == "pl.byc.class_membership.frame" &&
				candidate.Family.Family.ID.String() == "construction:pl.identification" &&
				candidate.CompiledVariant.Variant.ID.String() == "construction:pl.identification.to"
		}
		mode, valid := polishIdentificationReferenceMode(source, candidate.NodeID)
		role, referentID, _, validState := polishIdentificationState(source, candidate.NodeID)
		if !valid || !validState || !polishIdentificationAttributesAllowed(source, candidate.NodeID, role, referentID, candidate.TransferBinding.CoreRef) {
			return false
		}
		switch concept {
		case "concept:class_membership":
			return mode == "proximal" || mode == "distal" || mode == "unresolved"
		case "concept:black":
			gender, _, knownFeatures := polishIdentificationReferentFeatures(source, referentID, all)
			if !knownFeatures {
				return false
			}
			variant := candidate.CompiledVariant.Variant.ID.String()
			switch gender {
			case "masculine_animal":
				return variant == "construction:pl.attribution.ten" && (mode == "proximal" || mode == "demonstrative")
			case "neuter":
				return variant == "construction:pl.attribution.to" && (mode == "proximal" || mode == "demonstrative")
			default:
				return false
			}
		}
		return false
	}
	route, negativeFamily := polishPresentNegationRouteForFamily(candidate.Family.Family.ID)
	if !negativeFamily {
		if polishNegatedPredicateTarget(source, candidate.NodeID) {
			return false
		}
		return true
	}
	_, event, sourceRoute, negated := polishNegatedPresent(source)
	if !negated || event.ID != candidate.NodeID || sourceRoute.sense != candidate.Sense.ID.String() || sourceRoute.family != route.family ||
		event.TemporalFrame == meaning.TemporalFramePast || !polishFrameEvidenceConsistent(source, event.ID, candidate.TransferBinding.CoreRef) {
		return false
	}
	referent, exists := polishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || route.concept == "concept:class_membership" && referent.Multiplicity == meaning.MultiplicityMany {
		return false
	}
	return polishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, candidate.TransferBinding.CoreRef, true)
}

func polishWHQuestionTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	answer := source.QuestionAnswer
	if answer == nil || candidate.QuestionAnswer == nil || *candidate.QuestionAnswer != *answer ||
		candidate.Family.Family.ID.String() != "construction:pl.question.wh" ||
		candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze || answer.TargetNodeID != candidate.NodeID {
		return false
	}
	roleSlots := make(map[meaning.Role]construction.SlotSpec, len(candidate.Family.Family.SlotSpecs))
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.ArgumentBinding != nil {
			roleSlots[slot.ArgumentBinding.SemanticRole] = slot
		}
	}
	answerSlot, exists := roleSlots[answer.Role]
	if !exists || answerSlot.Category != construction.CategoryNP {
		return false
	}
	sequenceSlots := make(map[string]struct{}, len(candidate.CompiledVariant.Variant.Sequence))
	terminals := make(map[string]struct{})
	for _, token := range candidate.CompiledVariant.Variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		} else if token.Kind == construction.TokenTerminal {
			terminals[token.Terminal] = struct{}{}
		}
	}
	if _, included := sequenceSlots[answerSlot.ID]; included {
		return false
	}
	for _, argument := range candidate.Frame.Arguments {
		if !argument.Required || argument.Role == answer.Role {
			continue
		}
		slot, bound := roleSlots[argument.Role]
		if !bound {
			return false
		}
		if _, included := sequenceSlots[slot.ID]; !included {
			return false
		}
	}
	if _, hasVerb := sequenceSlots["verb"]; !hasVerb {
		return false
	}
	if _, hasQuestionMark := terminals["?"]; !hasQuestionMark {
		return false
	}
	switch answer.Role {
	case "experiencer":
		_, kto := terminals["kto"]
		_, kogo := terminals["kogo"]
		return kto && !kogo
	case "stimulus":
		_, kogo := terminals["kogo"]
		_, kto := terminals["kto"]
		return kogo && !kto
	default:
		return false
	}
}

func polishNegatedPredicateTarget(source semantic.MeaningCandidate, predicateID meaning.NodeID) bool {
	if source.Meaning == nil {
		return false
	}
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindScope && node.Operator == meaning.OperatorNegation && node.Target == predicateID {
			return true
		}
	}
	return false
}

func sourceHasScope(source semantic.MeaningCandidate) bool {
	if source.Meaning == nil {
		return false
	}
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindScope {
			return true
		}
	}
	return false
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
	return polishNumberForReferentCandidates(source, targetNode, role, predicate.NodeID, predicate.TransferBinding.CoreRef, candidates)
}

func polishPredicateNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, bool) {
	if role, referentID, _, valid := polishIdentificationState(source, predicate.NodeID); valid {
		return polishNumberForReferentCandidates(source, referentID, role, predicate.NodeID, predicate.TransferBinding.CoreRef, candidates)
	}
	return polishSubjectNumber(source, predicate, candidates)
}

func polishSubjectPastGender(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate, number string) (string, bool) {
	gender, found := polishSubjectGender(source, predicate, candidates)
	if !found {
		return "", false
	}
	if number == "singular" {
		switch gender {
		case "masculine_personal", "feminine", "neuter":
			return gender, true
		case "masculine_animal", "masculine_inanimate":
			if predicate.Sense.ID.String() == "sense:pl.widziec.visual" {
				return "masculine_personal", true
			}
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

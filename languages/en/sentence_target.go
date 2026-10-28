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
	return NewSentenceExecutionBuilderWithSeed(seed)
}

// NewSentenceExecutionBuilderWithSeed compiles target morphology from the
// caller's exact English seed.
func NewSentenceExecutionBuilderWithSeed(seed *Seed) (runtime.MorphologyExecutionBuilder, error) {
	if seed == nil {
		return runtime.MorphologyExecutionBuilder{}, fmt.Errorf("English execution builder requires an English morphology seed")
	}
	profile, _, paradigms, exceptions := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "en" || paradigms == nil || exceptions == nil {
		return runtime.MorphologyExecutionBuilder{}, fmt.Errorf("English execution builder requires an English morphology seed")
	}
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
	return NewSentenceTargetWithSeed(snapshot, registry, seed)
}

// NewSentenceTargetWithSeed builds all English target generators from one
// caller-supplied morphology seed.
func NewSentenceTargetWithSeed(snapshot *pack.Pack, registry *language.Registry, seed *Seed) (*runtime.TargetPipeline, error) {
	if snapshot == nil || registry == nil {
		return nil, fmt.Errorf("English target pipeline requires a snapshot and verified language registry")
	}
	if seed == nil {
		return nil, fmt.Errorf("English target pipeline requires an English morphology seed")
	}
	profile, _, _, _ := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "en" {
		return nil, fmt.Errorf("English target pipeline requires an English morphology seed")
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
		Language: "en", LanguageRegistry: registry, Lexicon: snapshot.Lexicon(), Valency: valencyIndex,
		Constructions: constructions, Variants: variants, Core: snapshot.CoreGraph(),
		OperatorBuilder: englishSentenceOperatorBuilder{}, ExecutionBuilder: execution,
		PlanBuilder: englishSentencePlanAnnotations{}, CandidatePolicy: englishDeterminerCandidatePolicy{},
		ClauseCoordination:   &coordination,
		ScopedRealizer:       &runtime.ScopedClauseRealizer{Language: "en", Plan: scopePlan, Valency: valencyIndex, Morphology: seed, Features: englishScopedFeaturePlanner{}},
		RelationPathRealizer: &runtime.RelationPathRealizer{Language: "en", Plan: relationPlan, Morphology: seed, Features: englishRelationPathFeaturePlanner{}},
	}, nil
}

type englishSentenceFeatures struct{}

func (englishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if has, valid := englishParticipantContextIntent(source); has && !valid {
		return nil, fmt.Errorf("English participant anchor context does not match the supplied Meaning")
	}
	if (predicate.Kind == meaning.KindEvent || predicate.Kind == meaning.KindState) && !englishPredicateEvidenceConsistent(source, predicate) {
		return nil, fmt.Errorf("English target Meaning attributes conflict with explicit source evidence")
	}
	if predicate.QuestionAnswer != nil && predicate.QuestionAnswer.TargetNodeID == predicate.NodeID && slot.Category == construction.CategoryVP {
		switch predicate.QuestionAnswer.Role {
		case "experiencer":
			person, number, _, _, ok := englishWHSubjectAgreement(source, predicate, candidates)
			if !ok {
				return nil, fmt.Errorf("English subject WH cannot preserve the supplied person and number evidence")
			}
			return englishFiniteVerbFeatures("present", person, number), nil
		case "stimulus":
			person, number, _, _, ok := englishWHSubjectAgreement(source, predicate, candidates)
			if !ok {
				return nil, fmt.Errorf("English object WH cannot preserve the supplied subject agreement evidence")
			}
			values := []features.FeatureValue{{Feature: "verb_form", State: features.StateKnown, Value: "base"}}
			if person != "" {
				values = append(values, features.FeatureValue{Feature: "person", State: features.StateKnown, Value: features.FeatureValueID(person)})
			}
			if number != "" {
				values = append(values,
					features.FeatureValue{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					features.FeatureValue{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
				)
			}
			return values, nil
		default:
			return nil, fmt.Errorf("English WH target does not support answer role %q", predicate.QuestionAnswer.Role)
		}
	}
	if slot.ModifierBinding != nil && englishAttributedSeeingFamily(predicate) {
		shape, valid := englishAttributedSeeingShape(source)
		if !valid || source.Meaning == nil {
			return nil, fmt.Errorf("English attributed seeing adjective requires the licensed two-root Meaning")
		}
		modifier, modifierFound := realizer.FindModifier(*source.Meaning, predicate.NodeID, predicate.Family.Family, slot.ID)
		if !englishAttributedSeeingContextsValid(source) || !modifierFound || modifier.ID != shape.property.ID || candidate.NodeID != modifier.ID ||
			candidate.Sense.ID.String() != "sense:en.black.color" || candidate.Lexeme.ID.String() != "lexeme:en.black.adj" || candidate.TransferBinding.CoreRef.String() != "concept:black" {
			return nil, fmt.Errorf("English attributed seeing adjective is not linked to the selected subject property")
		}
		return knownFeature("adjective_degree", "positive"), nil
	}
	if identification, recognized := englishNegativeIdentificationState(source); recognized && identification.stateID == predicate.NodeID {
		if !identification.valid || !englishIdentificationAttributesAllowed(source, predicate.NodeID, identification.role, identification.referent, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("English negative identification cannot realize the supplied Meaning attributes")
		}
		if slot.Category == construction.CategoryVP && predicate.Sense.ID.String() == identification.senseID {
			return englishFiniteVerbFeatures("present", "third", "singular"), nil
		}
		if slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && string(slot.ArgumentBinding.SemanticRole) == identification.role {
			if !englishIdentificationMemberCountSupported(source, identification, []realizer.TargetCandidate{candidate}) {
				return nil, fmt.Errorf("English class membership requires an explicit count-noun target for its member referent")
			}
			return []features.FeatureValue{{Feature: "noun_number", State: features.StateKnown, Value: "singular"}, {Feature: "number", State: features.StateKnown, Value: "singular"}}, nil
		}
	}
	if identification, recognized := englishIdentificationState(source); recognized && identification.stateID == predicate.NodeID {
		if !identification.valid || !englishIdentificationAttributesAllowed(source, predicate.NodeID, identification.role, identification.referent, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("English identification cannot realize the supplied Meaning attributes")
		}
		if slot.Category == construction.CategoryVP && predicate.Sense.ID.String() == identification.senseID {
			return englishFiniteVerbFeatures("present", "third", "singular"), nil
		}
		if slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && string(slot.ArgumentBinding.SemanticRole) == identification.role {
			if !englishIdentificationMemberCountSupported(source, identification, []realizer.TargetCandidate{candidate}) {
				return nil, fmt.Errorf("English class membership requires an explicit count-noun target for its member referent")
			}
			return []features.FeatureValue{{Feature: "noun_number", State: features.StateKnown, Value: "singular"}, {Feature: "number", State: features.StateKnown, Value: "singular"}}, nil
		}
	}
	if slot.Category == construction.CategoryAP && predicate.Sense.ID.String() == "sense:en.black.color" && candidate.Sense.ID.String() == "sense:en.black.color" && candidate.Lexeme.ID.String() == "lexeme:en.black.adj" {
		return knownFeature("adjective_degree", "positive"), nil
	}
	if slot.Category == construction.CategoryVP {
		if _, event, route, found := englishNegatedPresent(source); found && event.ID == predicate.NodeID && predicate.Sense.ID.String() == route.sense {
			if route.statePredicate {
				if event.TemporalFrame == meaning.TemporalFramePast {
					return nil, fmt.Errorf("English negative identification does not support a past temporal frame")
				}
				return englishFiniteVerbFeatures("present", "third", "singular"), nil
			}
			number, known := englishPresentNegationNumber(source, predicate, route, candidates)
			if !known {
				return nil, fmt.Errorf("English present negation cannot realize the supplied subject multiplicity")
			}
			return []features.FeatureValue{
				{Feature: "verb_form", State: features.StateKnown, Value: "imperative"},
				{Feature: "mood", State: features.StateKnown, Value: "imperative"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
				{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
			}, nil
		}
		predicateMeaning, exists := englishMeaningNode(source, predicate.NodeID)
		if !exists {
			return nil, nil
		}
		if !englishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("English temporal frame conflicts with explicit source evidence")
		}
		tense := predicateMeaning.TemporalFrame
		if tense == "" {
			tense = meaning.TemporalFramePresent
		}
		if tense == meaning.TemporalFramePast {
			return englishFiniteVerbFeatures("past", "", ""), nil
		}
		subjectID := englishSubjectNode(source, predicate.NodeID)
		if englishParticipantAnchorIntent(source, subjectID) {
			if anchor, anchored := englishParticipantAnchor(source, subjectID); !anchored || anchor.Kind != "participant" {
				return nil, fmt.Errorf("English participant anchor evidence is inconsistent")
			} else {
				person, grammaticalNumber := "", "singular"
				switch anchor.Value {
				case "speaker":
					person = "first"
				case "speaker_inclusive_group":
					person = "first"
					grammaticalNumber = "plural"
				case "addressee":
					person = "second"
				}
				if person != "" {
					return englishFiniteVerbFeatures("present", person, grammaticalNumber), nil
				}
			}
		}
		number, _, known := englishPredicateNumber(source, predicate, candidates)
		if !known {
			return nil, fmt.Errorf("English present predicate cannot realize subject multiplicity")
		}
		person := "third"
		if values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef); found {
			if explicit := string(values["person"]); explicit == "first" || explicit == "second" || explicit == "third" {
				person = explicit
			}
		}
		return englishFiniteVerbFeatures("present", person, number), nil
	}
	if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil {
		return nil, nil
	}
	if candidate.Lexeme.POS == "pronoun" {
		if !englishParticipantPronoun(source, candidate.NodeID, candidate.Lexeme.ID.String()) {
			return nil, fmt.Errorf("English participant pronoun requires its matching explicit anchor")
		}
		person, number := "second", "singular"
		switch candidate.Lexeme.ID.String() {
		case "lexeme:en.i.pron":
			person = "first"
		case "lexeme:en.we.pron":
			person = "first"
			number = "plural"
		}
		values := []features.FeatureValue{{Feature: "pronoun_case", State: features.StateKnown, Value: "subject"}, {Feature: "pronoun_person", State: features.StateKnown, Value: features.FeatureValueID(person)}, {Feature: "person", State: features.StateKnown, Value: features.FeatureValueID(person)}}
		if candidate.Lexeme.ID.String() != "lexeme:en.you.pron" {
			values = append(values, features.FeatureValue{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)}, features.FeatureValue{Feature: "pronoun_number", State: features.StateKnown, Value: features.FeatureValueID(number)})
		}
		return values, nil
	}
	number, semanticNumber, ok := englishTargetNumber(source, candidate.NodeID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
	if !ok {
		return nil, fmt.Errorf("English target lexical number class cannot realize the Meaning multiplicity")
	}
	if semanticNumber {
		referent, exists := englishMeaningNode(source, candidate.NodeID)
		checkPredicateNumber := englishPredicateNumberApplies(predicate, slot.ArgumentBinding.SemanticRole)
		if !exists || !englishMultiplicityEvidenceConsistent(source, predicate.NodeID, slot.ArgumentBinding.SemanticRole, referent, predicate.TransferBinding.CoreRef, checkPredicateNumber) {
			return nil, fmt.Errorf("English referent multiplicity conflicts with explicit source evidence")
		}
	}
	return []features.FeatureValue{
		{Feature: "noun_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
		{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
	}, nil
}

type englishSentenceOperatorBuilder struct{}

func (englishSentenceOperatorBuilder) Build(ctx context.Context, source semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
	if ctx == nil {
		return nil, false, fmt.Errorf("English operator evidence requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if intent, valid := englishParticipantContextIntent(source); intent && !valid {
		return nil, false, englishParticipantContextError{}
	}
	if source.QuestionAnswer != nil {
		if source.Meaning == nil {
			return nil, false, fmt.Errorf("English WH operator requires its typed Meaning question")
		}
		if _, _, _, negated := englishNegatedPresent(source); negated {
			return nil, false, fmt.Errorf("English WH target does not combine a typed answer with present negation")
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
			return nil, false, fmt.Errorf("English WH operator requires one question wrapper over the typed answer target")
		}
		familyID, err := ids.NewConstructionID("en.question.wh")
		if err != nil {
			return nil, false, err
		}
		return []construction.Operator{{
			ID: "question-wh", Kind: construction.OperatorQuestion, Value: construction.ValueWH,
			InputFamilyID: familyID, OutputFamilyID: familyID,
			SemanticScope: construction.SemanticScope{TargetNodeID: source.QuestionAnswer.TargetNodeID, ScopeNodeID: question.ID},
			Capabilities:  []construction.Capability{"question-wh"}, Provenance: englishWHQuestionSource,
		}}, true, nil
	}
	scope, event, route, found := englishNegatedPresent(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(route.family, "construction:"))
	if err != nil {
		return nil, false, err
	}
	provenanceSource := englishNegationConstructionSource
	if route.statePredicate {
		provenanceSource = englishIdentificationNegationSource
	}
	return []construction.Operator{{
		ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
		InputFamilyID: familyID, OutputFamilyID: familyID,
		SemanticScope: construction.SemanticScope{TargetNodeID: event.ID, ScopeNodeID: scope.ID},
		Capabilities:  []construction.Capability{"negation-negated"}, Provenance: provenanceSource,
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

// englishWHSubjectAgreement keeps question agreement tied to the explicit
// subject participant and its selected target number. A target-local default
// may guide the licensed surface only when the realization plan records it as
// an assumption.
func englishWHSubjectAgreement(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (person, number string, personKnown, numberKnown, ok bool) {
	referent, exists := englishReferentForRole(source, predicate.NodeID, "experiencer")
	if !exists {
		if source.QuestionAnswer == nil || source.QuestionAnswer.TargetNodeID != predicate.NodeID || source.QuestionAnswer.Role != "experiencer" {
			return "", "", false, false, false
		}
		values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
		person, number := "", ""
		if found {
			person = string(values["person"])
			number = string(values["verb_number"])
			if number == "" {
				number = string(values["number"])
			}
		}
		if person != "" && person != "first" && person != "second" && person != "third" ||
			number != "" && number != "singular" && number != "plural" {
			return "", "", false, false, false
		}
		if person != "" && person != "third" {
			return "", "", false, false, false
		}
		personKnown, numberKnown = person != "", number != ""
		person = "third"
		if number == "" {
			number = "singular"
		}
		return person, number, personKnown, numberKnown, true
	}
	personFeature, numberFeature := "", ""
	mergeFeatures := func(values map[features.FeatureID]features.FeatureValueID, found bool) bool {
		if !found {
			return true
		}
		if value := string(values[features.FeatureID("person")]); value != "" {
			if value != "first" && value != "second" && value != "third" || personFeature != "" && personFeature != value {
				return false
			}
			personFeature = value
		}
		for _, key := range []features.FeatureID{"number", "verb_number"} {
			value := string(values[key])
			if value == "" {
				continue
			}
			if value != "singular" && value != "plural" || numberFeature != "" && numberFeature != value {
				return false
			}
			numberFeature = value
		}
		return true
	}
	if !mergeFeatures(runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)) ||
		!mergeFeatures(runtime.SourceFeatures(source, "experiencer", coreConcept(referent))) {
		return "", "", false, false, false
	}

	localNumber, semanticNumber, numberOK := englishPredicateNumber(source, predicate, candidates)
	if !numberOK {
		return "", "", false, false, false
	}
	if multiplicity := englishMultiplicityNumber(referent.Multiplicity); multiplicity != "" && numberFeature != "" && multiplicity != numberFeature {
		return "", "", false, false, false
	}
	if numberFeature != "" {
		if localNumber != numberFeature {
			return "", "", false, false, false
		}
		localNumber = numberFeature
	}
	person, personKnown = personFeature, personFeature != ""
	number, numberKnown = localNumber, semanticNumber || numberFeature != ""

	if englishParticipantAnchorIntent(source, referent.ID) {
		anchor, anchored := englishParticipantAnchor(source, referent.ID)
		if !anchored || anchor.Kind != "participant" {
			return "", "", false, false, false
		}
		anchorPerson, anchorNumber := "", ""
		switch anchor.Value {
		case "speaker":
			anchorPerson, anchorNumber = "first", "singular"
		case "speaker_inclusive_group":
			anchorPerson, anchorNumber = "first", "plural"
		case "addressee":
			anchorPerson = "second"
		default:
			return "", "", false, false, false
		}
		if personFeature != "" && personFeature != anchorPerson ||
			anchorNumber != "" && numberKnown && number != anchorNumber {
			return "", "", false, false, false
		}
		person, personKnown = anchorPerson, true
		if anchorNumber != "" {
			number, numberKnown = anchorNumber, true
		}
	} else if person != "" && person != "third" {
		// Non-deictic noun phrases take third-person agreement. First- and
		// second-person subjects require their explicit participant anchor.
		return "", "", false, false, false
	}
	if person == "" {
		// A non-deictic nominal subject has third-person agreement. The target
		// lexical class does not determine whether an unknown referent is a
		// speaker or addressee.
		person, personKnown = "third", referent.Concept != ""
	}
	if number == "" {
		number = "singular"
	}
	if source.QuestionAnswer != nil && source.QuestionAnswer.Role == "stimulus" &&
		!numberKnown && englishParticipantAnchorIntent(source, referent.ID) {
		anchor, anchored := englishParticipantAnchor(source, referent.ID)
		if anchored && anchor.Value == "addressee" {
			// Do-support is invariant across the addressee's unresolved number.
			number = ""
		}
	}
	return person, number, personKnown, numberKnown, true
}

func englishWHQuestionNumberKnown(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, role meaning.Role) bool {
	values, found := runtime.SourceFeatures(source, "", predicate.TransferBinding.CoreRef)
	if found && (string(values["number"]) == "singular" || string(values["number"]) == "plural" ||
		string(values["verb_number"]) == "singular" || string(values["verb_number"]) == "plural") {
		return true
	}
	referent, exists := englishReferentForRole(source, predicate.NodeID, role)
	if !exists {
		return false
	}
	if englishMultiplicityNumber(referent.Multiplicity) != "" {
		return true
	}
	for _, request := range []struct {
		role    meaning.Role
		concept ids.ConceptID
	}{{"", predicate.TransferBinding.CoreRef}, {role, coreConcept(referent)}} {
		values, found := runtime.SourceFeatures(source, request.role, request.concept)
		if found && (string(values["number"]) == "singular" || string(values["number"]) == "plural" ||
			string(values["verb_number"]) == "singular" || string(values["verb_number"]) == "plural") {
			return true
		}
	}
	if englishParticipantAnchorIntent(source, referent.ID) {
		anchor, anchored := englishParticipantAnchor(source, referent.ID)
		return anchored && anchor.Kind == "participant" && (anchor.Value == "speaker" || anchor.Value == "speaker_inclusive_group")
	}
	return false
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
	if identification, recognized := englishNegativeIdentificationState(source); recognized && identification.stateID == predicate.ID {
		if !identification.valid || !englishIdentificationAttributesAllowed(source, predicate.ID, identification.role, identification.referent, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("English negative identification cannot realize the supplied Meaning attributes")
		}
		predicatePath, found := runtime.MeaningPath(*source.Meaning, predicate.ID)
		referentPath, referentPathFound := runtime.MeaningPath(*source.Meaning, identification.referent)
		referent, referentFound := englishMeaningNode(source, identification.referent)
		if !found || !referentPathFound || !referentFound {
			return nil, nil, nil, fmt.Errorf("English negative identification is missing its Meaning paths")
		}
		assumptions := englishIdentificationDefaults(source, predicatePath, referentPath, predicate, referent, selected.TransferBinding.CoreRef)
		if identification.mode == "unresolved" {
			value, ok := englishIdentificationReferenceChoice(selected.CompiledVariant.Variant.ID.String(), identification.family)
			if !ok {
				return nil, nil, nil, nil
			}
			assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "reference", Value: value, Reason: "the source preserves unresolved reference; this candidate selects the named English deictic form"})
		}
		return nil, assumptions, nil, nil
	}
	if identification, recognized := englishIdentificationState(source); recognized && identification.stateID == predicate.ID {
		if !identification.valid || !englishIdentificationAttributesAllowed(source, predicate.ID, identification.role, identification.referent, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("English identification cannot realize the supplied Meaning attributes")
		}
		predicatePath, found := runtime.MeaningPath(*source.Meaning, predicate.ID)
		referentPath, referentPathFound := runtime.MeaningPath(*source.Meaning, identification.referent)
		referent, referentFound := englishMeaningNode(source, identification.referent)
		if !found || !referentPathFound || !referentFound {
			return nil, nil, nil, fmt.Errorf("English identification is missing its Meaning paths")
		}
		assumptions := englishIdentificationDefaults(source, predicatePath, referentPath, predicate, referent, selected.TransferBinding.CoreRef)
		mode := identification.mode
		if mode == "unresolved" || mode == "demonstrative" {
			value, ok := englishIdentificationReferenceChoice(selected.CompiledVariant.Variant.ID.String(), identification.family)
			if !ok {
				return nil, nil, nil, nil
			}
			assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "reference", Value: value, Reason: "the source preserves unresolved reference; this candidate selects the named English deictic form"})
		}
		return nil, assumptions, nil, nil
	}
	if !englishPredicateEvidenceConsistent(source, selected) {
		return nil, nil, nil, fmt.Errorf("English target Meaning attributes conflict with explicit source evidence")
	}
	assumptions := make([]realizer.FieldAssumption, 0, 3+2*len(predicate.Participants))
	losses := make([]realizer.FieldLoss, 0, len(predicate.Participants))
	predicatePath, pathFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
	if !pathFound {
		return nil, nil, nil, nil
	}
	whSubjectAnswer := source.QuestionAnswer != nil && source.QuestionAnswer.TargetNodeID == selected.NodeID && source.QuestionAnswer.Role == "experiencer"
	whSubjectNumberKnown := whSubjectAnswer && englishWHQuestionNumberKnown(source, selected, "experiencer")
	whSubjectNumberAssumptionAdded := false
	if whSubjectAnswer && !whSubjectNumberKnown {
		if _, hasSubjectReferent := englishReferentForRole(source, selected.NodeID, "experiencer"); !hasSubjectReferent {
			assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "singular", Reason: "WH subject answer carries no number evidence; this target candidate uses singular agreement"})
			whSubjectNumberAssumptionAdded = true
		}
	}
	if predicate.TemporalFrame == "" {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning carries no temporal frame; this response uses the approved present-tense route"})
	}
	subjectID := englishSubjectNode(source, selected.NodeID)
	_, participantPersonKnown := englishParticipantAnchor(source, subjectID)
	subject, subjectExists := englishMeaningNode(source, subjectID)
	nominalSubject := subjectExists && subject.Kind == meaning.KindReferent && !englishParticipantAnchorIntent(source, subjectID)
	if values, found := runtime.SourceFeatures(source, "", selected.TransferBinding.CoreRef); (!found || values["person"] == "") && !participantPersonKnown && !nominalSubject {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "person", Value: "third", Reason: "Meaning carries no grammatical person; this response uses the approved third-person form"})
	}
	if role := englishPredicateSubjectRoleForTarget(selected); role != "" {
		if referent, found := englishReferentForRole(source, selected.NodeID, role); found && referent.Multiplicity == "" && !whSubjectAnswer {
			nodeID := englishSubjectNode(source, selected.NodeID)
			anchor, anchored := englishParticipantAnchor(source, nodeID)
			if !anchored || anchor.Value != "addressee" {
				assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "lexeme_default", Reason: "Meaning carries no subject multiplicity; agreement follows the target subject's local lexical class"})
			}
		}
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
		participantAnchor, participantAnchored := englishParticipantAnchor(source, referent.ID)
		if referent.Multiplicity == "" && !(participantAnchored && participantAnchor.Value == "addressee") {
			value, reason := "lexeme_default", "Meaning carries no referent multiplicity; the target uses its local number class"
			if whSubjectAnswer && participant.Role == "experiencer" && !whSubjectNumberKnown && !whSubjectNumberAssumptionAdded {
				value, reason = "singular", "WH subject answer carries no number evidence; this target candidate uses singular agreement"
				whSubjectNumberAssumptionAdded = true
			}
			assumptions = append(assumptions, realizer.FieldAssumption{Path: referentPath, Field: "number", Value: value, Reason: reason})
		}
		if participantAnchored && participantAnchor.Value == "addressee" && referent.Multiplicity != "" {
			losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "number", Outcome: status.OutcomeIncomplete, Reason: "English you does not encode the source addressee number"})
		}
		if selected.Family.Family.ID.String() == "construction:en.transitive_personal" && (participant.Role == "experiencer" || participant.Role == "reader") {
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
		assumptionValue := determiner
		switch determiner {
		case "indefinite":
			reason = "the source does not encode English definiteness; this candidate selects the indefinite article " + article
		case "bare":
			reason = "the selected English target construction realizes the " + string(participant.Role) + " referent with a zero article"
		}
		assumptions = append(assumptions, realizer.FieldAssumption{
			Path: referentPath, Field: "determiner", Value: assumptionValue, Reason: reason,
		})
	}
	return nil, assumptions, losses, nil
}

// englishDeterminerCandidatePolicy retains only variants supported by source
// determiner/number evidence and explicit target-lexeme article properties.
// Missing source definiteness keeps all compatible options visible.
type englishDeterminerCandidatePolicy struct{}

func (englishDeterminerCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	// The kinship reading of child is licensed for English source analysis, but
	// offspring remains the canonical English target lexeme for child_of_person.
	if candidate.Kind == meaning.KindReferent && englishChildKinshipTargetCandidate(candidate) {
		return false
	}
	if has, valid := englishParticipantContextIntent(source); has && !valid {
		return false
	}
	if source.QuestionAnswer != nil && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) && candidate.NodeID == source.QuestionAnswer.TargetNodeID {
		return englishWHQuestionTargetCandidate(source, candidate, candidates)
	}
	const (
		transitiveFamily     = "construction:en.transitive"
		transitivePastFamily = "construction:en.transitive_past"
		obliqueFamily        = "construction:en.oblique"
		attributedFamily     = "construction:en.attributed_seeing"
	)
	familyID := candidate.Family.Family.ID.String()
	if (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) && !englishPredicateEvidenceConsistent(source, candidate) {
		return false
	}
	if familyID == attributedFamily && !englishAttributedSeeingMainCandidate(source, candidate) {
		return false
	}
	if shape, valid := englishAttributedSeeingShape(source); valid {
		if candidate.NodeID == shape.property.ID {
			return englishAttributedSeeingModifierCandidate(source, candidate, candidates)
		}
		if candidate.NodeID == shape.event.ID && familyID != attributedFamily {
			return false
		}
	}
	if identification, recognized := englishNegativeIdentificationState(source); recognized {
		if candidate.Kind == meaning.KindState && candidate.NodeID == identification.stateID {
			if !identification.valid || !englishIdentificationAttributesAllowed(source, candidate.NodeID, identification.role, identification.referent, candidate.TransferBinding.CoreRef) ||
				!englishIdentificationMemberCountSupported(source, identification, candidates) || familyID != identification.family || candidate.Sense.ID.String() != identification.senseID || candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze {
				return false
			}
			variant := candidate.CompiledVariant.Variant.ID.String()
			if identification.mode == "proximal" {
				return variant == identification.family+".this"
			}
			_, allowed := englishIdentificationReferenceChoice(variant, identification.family)
			return allowed && (variant == identification.family+".this" || variant == identification.family+".that" || variant == identification.family+".it")
		}
		if candidate.Family.Family.ID.String() == "construction:en.identification" || candidate.Family.Family.ID.String() == "construction:en.attribution" {
			return false
		}
	}
	if identification, recognized := englishIdentificationState(source); recognized {
		if candidate.Kind == meaning.KindState && candidate.NodeID == identification.stateID {
			if !identification.valid || !englishIdentificationAttributesAllowed(source, candidate.NodeID, identification.role, identification.referent, candidate.TransferBinding.CoreRef) ||
				!englishIdentificationMemberCountSupported(source, identification, candidates) {
				return false
			}
			variant := candidate.CompiledVariant.Variant.ID.String()
			_, allowed := englishIdentificationReferenceChoice(variant, identification.family)
			if identification.mode == "proximal" {
				allowed = variant == identification.family+".this"
			}
			if identification.mode == "demonstrative" {
				allowed = identification.family == "construction:en.attribution" && (variant == identification.family+".this" || variant == identification.family+".that")
			}
			if identification.mode == "unresolved" {
				allowed = identification.family == "construction:en.identification" && (variant == identification.family+".this" || variant == identification.family+".that" || variant == identification.family+".it")
			}
			return allowed && familyID == identification.family && candidate.Sense.ID.String() == identification.senseID
		}
	}
	if englishParticipantAnchorIntent(source, candidate.NodeID) {
		anchor, anchored := englishParticipantAnchor(source, candidate.NodeID)
		if !anchored || anchor.Kind != "participant" || !englishParticipantPronoun(source, candidate.NodeID, candidate.Lexeme.ID.String()) {
			return false
		}
	} else if candidate.Lexeme.POS == "pronoun" {
		return false
	}
	subjectNode := englishSubjectNode(source, candidate.NodeID)
	_, hasParticipantAnchor := englishParticipantAnchor(source, subjectNode)
	if hasParticipantAnchor && familyID != "construction:en.transitive_personal" && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) {
		return false
	}
	if candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze {
		return false
	}
	if familyID == "construction:en.transitive_personal" && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) {
		subject := englishSubjectNode(source, candidate.NodeID)
		anchor, anchored := englishParticipantAnchor(source, subject)
		if !anchored || anchor.Kind != "participant" {
			return false
		}
		node, ok := englishMeaningNode(source, candidate.NodeID)
		if !ok {
			return false
		}
		tense := node.TemporalFrame
		if tense == "" {
			tense = meaning.TemporalFramePresent
		}
		expected := ""
		if tense == meaning.TemporalFramePast {
			if candidate.Sense.ID.String() == "sense:en.read.text" {
				if anchor.Value == "speaker" || anchor.Value == "speaker_inclusive_group" {
					expected = "construction:en.transitive_personal.svo.reading.first"
				} else if anchor.Value == "addressee" {
					expected = "construction:en.transitive_personal.svo.reading.second"
				}
			}
		} else if anchor.Value == "speaker" || anchor.Value == "speaker_inclusive_group" {
			expected = "construction:en.transitive_personal.svo.first"
		} else if anchor.Value == "addressee" {
			expected = "construction:en.transitive_personal.svo.second"
		}
		variantID := candidate.CompiledVariant.Variant.ID.String()
		if expected == "" || (variantID != expected && !strings.HasPrefix(variantID, expected+".")) {
			return false
		}
		determinerVariant, ok := englishDeterminersForCandidate(candidate)
		if !ok {
			return false
		}
		role := meaning.Role("stimulus")
		if candidate.Sense.ID.String() == "sense:en.read.text" {
			role = "material"
		}
		var ref meaning.Node
		for _, part := range node.Participants {
			if part.Role == role {
				ref = findMeaningNode(*source.Meaning, part.Target)
				break
			}
		}
		if ref.ID == "" {
			return false
		}
		lexical := []realizer.TargetCandidate{}
		for _, item := range candidates {
			if item.NodeID == ref.ID && item.Kind == meaning.KindReferent && item.Lexeme.ID.String() != "" && !englishChildKinshipTargetCandidate(item) {
				lexical = append(lexical, item)
			}
		}
		article, articleKnown := commonEnglishIndefiniteArticle(lexical)
		number, numberKnown := "", false
		allMass, allPluralOnly := true, true
		for _, item := range lexical {
			candidateNumber, _, numberOK := englishTargetNumber(source, ref.ID, item.Lexeme.Properties, item.Lexeme.ParadigmID)
			if !numberOK || numberKnown && candidateNumber != number {
				return false
			}
			number, numberKnown = candidateNumber, true
			mass, pluralOnly := englishNumberClass(item.Lexeme.Properties, item.Lexeme.ParadigmID)
			allMass = allMass && mass
			allPluralOnly = allPluralOnly && pluralOnly
		}
		if !numberKnown || !englishVariantRequiresNPNumber(candidate, "object", number) {
			return false
		}
		if (determinerVariant.object == "a" || determinerVariant.object == "an") &&
			(allMass || allPluralOnly || number != "singular" || !articleKnown || article != determinerVariant.object) {
			return false
		}
		if determinerVariant.object == "bare" && !allMass && number != "plural" {
			return false
		}
		if values, found := runtime.SourceFeatures(source, role, coreConcept(ref)); found && values["determiner"] != "" && string(values["determiner"]) != englishDeterminerFeature(determinerVariant.object) {
			return false
		}
		return true
	}

	if (candidate.Sense.ID.String() == "sense:en.read.text" || candidate.Sense.ID.String() == "sense:en.see.visual") && !englishPastTargetFamilyMatches(source, candidate, familyID, transitiveFamily, transitivePastFamily) {
		return false
	}
	negativeRoute, isNegativeFamily := englishPresentNegationRouteForFamily(candidate.Family.Family.ID)
	if isNegativeFamily {
		_, event, sourceRoute, negated := englishNegatedPresent(source)
		if !negated || event.ID != candidate.NodeID || sourceRoute.family != negativeRoute.family || sourceRoute.sense != candidate.Sense.ID.String() {
			return false
		}
		number, known := englishPresentNegationNumber(source, candidate, sourceRoute, candidates)
		if !known || !sourceRoute.variantMatchesNumber(candidate.CompiledVariant.Variant.ID.String(), number) {
			return false
		}
	}
	if candidate.Kind != meaning.KindEvent && candidate.Kind != meaning.KindState || familyID != transitiveFamily && familyID != transitivePastFamily && familyID != obliqueFamily && familyID != attributedFamily && familyID != "construction:en.transitive_personal" && !isNegativeFamily {
		return true
	}
	variant, knownVariant := englishDeterminersForCandidate(candidate)
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
		if familyID == "construction:en.transitive_personal" && slot.ID == "subject" {
			continue
		}
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
				available.Lexeme.ID.String() != "" && !englishChildKinshipTargetCandidate(available) {
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
		allMass := true
		allPluralOnly := true
		number := ""
		numberSet, numberConflict := false, false
		for _, lexicalCandidate := range lexicalCandidates {
			mass, pluralOnly := englishNumberClass(lexicalCandidate.Lexeme.Properties, lexicalCandidate.Lexeme.ParadigmID)
			allMass = allMass && mass
			allPluralOnly = allPluralOnly && pluralOnly
			candidateNumber, _, numberOK := englishTargetNumber(source, referent.ID, lexicalCandidate.Lexeme.Properties, lexicalCandidate.Lexeme.ParadigmID)
			if !numberOK {
				return false
			}
			if !numberSet {
				number = candidateNumber
				numberSet = true
			} else if number != candidateNumber {
				numberConflict = true
			}
		}
		if numberConflict {
			number = ""
		}
		indefiniteArticle, hasCommonIndefiniteArticle := commonEnglishIndefiniteArticle(lexicalCandidates)
		if familyID == attributedFamily && slot.ID == "subject" {
			// The first pronounced word of this licensed NP is the modifier
			// black, even when the noun itself takes an (for example offspring).
			indefiniteArticle, hasCommonIndefiniteArticle = "a", true
		}
		mass := allMass
		if mass {
			number = "singular"
		}
		if allPluralOnly {
			number = "plural"
		}
		if !englishVariantRequiresNPNumber(candidate, slot.ID, number) {
			return false
		}
		actualArticle := englishDeterminerForSlot(variant, slot.ID)
		actualDeterminer := englishDeterminerFeature(actualArticle)
		if actualDeterminer == "indefinite" &&
			(mass || allPluralOnly || number != "singular" || !hasCommonIndefiniteArticle || actualArticle != indefiniteArticle) {
			return false
		}
		if determiner != "" {
			if actualDeterminer != determiner {
				return false
			}
			continue
		}
		if slot.ID == "object" && mass {
			if actualDeterminer != "bare" {
				return false
			}
			continue
		}
		if actualDeterminer == "bare" {
			if number != "plural" {
				return false
			}
			continue
		}
		if number == "plural" {
			if actualDeterminer != "definite" && actualDeterminer != "bare" {
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

// englishWHQuestionTargetCandidate selects a neutral target variant by its
// actual missing role slot and licensed WH terminals. Variant IDs remain data
// identifiers; they do not decide the answer role.
func englishWHQuestionTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	answer := source.QuestionAnswer
	if answer == nil || candidate.QuestionAnswer == nil || *candidate.QuestionAnswer != *answer ||
		candidate.Family.Family.ID.String() != "construction:en.question.wh" ||
		candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze ||
		candidate.CompiledVariant.Variant.Markedness != construction.MarkednessNeutral {
		return false
	}
	if answer.TargetNodeID != candidate.NodeID {
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
	switch answer.Role {
	case "experiencer":
		_, who := terminals["who"]
		_, whom := terminals["whom"]
		_, does := terminals["does"]
		_, do := terminals["do"]
		if !who || whom || does || do {
			return false
		}
		person, number, _, _, agreementKnown := englishWHSubjectAgreement(source, candidate, candidates)
		if !agreementKnown || person != "third" {
			return false
		}
		variantNumber := ""
		for _, condition := range candidate.CompiledVariant.Variant.Conditions {
			if condition.Kind != construction.ConditionFeatureEquals || condition.SlotID != "verb" || condition.Feature != "verb_number" {
				continue
			}
			value := string(condition.Value)
			if value != "singular" && value != "plural" || variantNumber != "" && variantNumber != value {
				return false
			}
			variantNumber = value
		}
		return variantNumber != "" && variantNumber == number
	case "stimulus":
		_, whom := terminals["whom"]
		_, does := terminals["does"]
		_, do := terminals["do"]
		_, who := terminals["who"]
		if !whom || who || does == do {
			return false
		}
		person, number, _, _, agreementKnown := englishWHSubjectAgreement(source, candidate, candidates)
		if !agreementKnown {
			return false
		}
		if does {
			return person == "third" && number == "singular"
		}
		return person != "third" || number != "singular"
	default:
		return false
	}
}

func englishChildKinshipTargetCandidate(candidate realizer.TargetCandidate) bool {
	return candidate.Lexeme.ID.String() == "lexeme:en.child.n" && candidate.Sense.ID.String() == "sense:en.child.kinship"
}

func englishPastTargetFamilyMatches(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, familyID, presentFamily, pastFamily string) bool {
	if familyID != presentFamily && familyID != pastFamily {
		return true
	}
	predicate, found := englishMeaningNode(source, candidate.NodeID)
	if !found {
		return false
	}
	if !englishFrameEvidenceConsistent(source, candidate.NodeID, candidate.TransferBinding.CoreRef) {
		return false
	}
	frame := predicate.TemporalFrame
	if frame == "" {
		frame = meaning.TemporalFramePresent
	}
	if familyID == pastFamily {
		return frame == meaning.TemporalFramePast
	}
	return frame == meaning.TemporalFramePresent
}

type englishDeterminers struct {
	subject string
	object  string
	target  string
}

func englishDeterminerVariant(id string) (englishDeterminers, bool) {
	variant, exists := englishDeterminerVariants[id]
	if !exists && strings.HasPrefix(id, "construction:en.transitive_personal.") {
		switch {
		case strings.HasSuffix(id, ".indefinite_object"):
			return englishDeterminers{object: "a"}, true
		case strings.HasSuffix(id, ".indefinite_object_an"):
			return englishDeterminers{object: "an"}, true
		case strings.HasSuffix(id, ".reading.first") || strings.HasSuffix(id, ".reading.second") || strings.HasSuffix(id, ".svo.first") || strings.HasSuffix(id, ".svo.second"):
			return englishDeterminers{object: "the"}, true
		}
	}
	if !exists {
		if id == "construction:en.transitive.svo.reading" || id == "construction:en.transitive_past.svo.reading.realize" {
			variant, exists = englishDeterminerVariants["construction:en.transitive.svo"]
			return variant, exists
		}
		if variant, exists = englishNegativeDeterminerVariant(id, "construction:en.transitive_negative.svo", englishNegativeTransitiveDeterminerVariants); exists {
			return variant, true
		}
		if variant, exists = englishNegativeDeterminerVariant(id, "construction:en.oblique_negative.svo", englishNegativeObliqueDeterminerVariants); exists {
			return variant, true
		}
	}
	const pastPrefix = "construction:en.transitive_past."
	if !exists && len(id) > len(pastPrefix) && id[:len(pastPrefix)] == pastPrefix {
		variant, exists = englishDeterminerVariants["construction:en.transitive."+id[len(pastPrefix):]]
	}
	return variant, exists
}

func englishNegativeDeterminerVariant(id, prefix string, variants map[string]englishDeterminers) (englishDeterminers, bool) {
	rest, found := strings.CutPrefix(id, prefix)
	if !found {
		return englishDeterminers{}, false
	}
	rest = strings.TrimPrefix(rest, ".plural")
	if rest == "" {
		return englishDeterminers{subject: "the", object: "the", target: "the"}, true
	}
	variant, exists := variants[rest]
	return variant, exists
}

var englishDeterminerVariants = map[string]englishDeterminers{
	"construction:en.attributed_seeing.svo":                                {subject: "the", object: "the"},
	"construction:en.attributed_seeing.svo.indefinite_object":              {subject: "the", object: "a"},
	"construction:en.attributed_seeing.svo.indefinite_object_an":           {subject: "the", object: "an"},
	"construction:en.attributed_seeing.svo.indefinite_subject":             {subject: "a", object: "the"},
	"construction:en.attributed_seeing.svo.indefinite_subject_object":      {subject: "a", object: "a"},
	"construction:en.attributed_seeing.svo.indefinite_subject_a_object_an": {subject: "a", object: "an"},
	"construction:en.transitive.svo":                                       {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo":                              {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.plural":                       {subject: "the", object: "the"},
	"construction:en.transitive.svo.indefinite_object":                     {subject: "the", object: "a"},
	"construction:en.transitive.svo.indefinite_subject":                    {subject: "a", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_object":             {subject: "a", object: "a"},
	"construction:en.transitive.svo.mass_object":                           {subject: "the", object: "bare"},
	"construction:en.transitive.svo.indefinite_subject_mass_object":        {subject: "a", object: "bare"},
	"construction:en.transitive.svo.indefinite_object_an":                  {subject: "the", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an":                 {subject: "an", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_a_object_an":        {subject: "a", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_object_a":        {subject: "an", object: "a"},
	"construction:en.transitive.svo.indefinite_subject_an_object_an":       {subject: "an", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_mass_object":     {subject: "an", object: "bare"},
	"construction:en.oblique.svo":                                          {subject: "the", target: "the"},
	"construction:en.oblique.svo.indefinite_target":                        {subject: "the", target: "a"},
	"construction:en.oblique.svo.indefinite_subject":                       {subject: "a", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_target":                {subject: "a", target: "a"},
	"construction:en.oblique.svo.indefinite_target_an":                     {subject: "the", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an":                    {subject: "an", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_a_target_an":           {subject: "a", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an_target_a":           {subject: "an", target: "a"},
	"construction:en.oblique.svo.indefinite_subject_an_target_an":          {subject: "an", target: "an"},
	"construction:en.oblique_negative.svo":                                 {subject: "the", target: "the"},
	"construction:en.oblique_negative.svo.plural":                          {subject: "the", target: "the"},
}

var englishNegativeTransitiveDeterminerVariants = map[string]englishDeterminers{
	".indefinite_object":               {subject: "the", object: "a"},
	".indefinite_subject":              {subject: "a", object: "the"},
	".indefinite_subject_object":       {subject: "a", object: "a"},
	".indefinite_object_an":            {subject: "the", object: "an"},
	".indefinite_subject_an":           {subject: "an", object: "the"},
	".indefinite_subject_a_object_an":  {subject: "a", object: "an"},
	".indefinite_subject_an_object_a":  {subject: "an", object: "a"},
	".indefinite_subject_an_object_an": {subject: "an", object: "an"},
}

var englishNegativeObliqueDeterminerVariants = map[string]englishDeterminers{
	".indefinite_target":               {subject: "the", target: "a"},
	".indefinite_subject":              {subject: "a", target: "the"},
	".indefinite_subject_target":       {subject: "a", target: "a"},
	".indefinite_target_an":            {subject: "the", target: "an"},
	".indefinite_subject_an":           {subject: "an", target: "the"},
	".indefinite_subject_a_target_an":  {subject: "a", target: "an"},
	".indefinite_subject_an_target_a":  {subject: "an", target: "a"},
	".indefinite_subject_an_target_an": {subject: "an", target: "an"},
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
	case "construction:en.attributed_seeing":
		return "construction:en.attributed_seeing.svo"
	default:
		return ""
	}
}

func englishDeterminerForRole(candidate realizer.TargetCandidate, role meaning.Role) (string, bool) {
	familyID := candidate.Family.Family.ID.String()
	if familyID == "construction:en.question.wh" {
		for _, slot := range candidate.Family.Family.SlotSpecs {
			if slot.ArgumentBinding == nil || slot.ArgumentBinding.SemanticRole != role {
				continue
			}
			for index, token := range candidate.CompiledVariant.Variant.Sequence {
				if token.Kind != construction.TokenSlot || token.SlotID != slot.ID {
					continue
				}
				if index > 0 {
					previous := candidate.CompiledVariant.Variant.Sequence[index-1]
					if previous.Kind == construction.TokenTerminal && isEnglishDeterminerToken(previous.Terminal) {
						return previous.Terminal, true
					}
				}
				return "bare", true
			}
		}
		return "", false
	}
	if familyID != "construction:en.transitive" && familyID != "construction:en.transitive_past" && familyID != "construction:en.transitive_negative" && familyID != "construction:en.oblique" && familyID != "construction:en.oblique_negative" && familyID != "construction:en.attributed_seeing" && familyID != "construction:en.transitive_personal" {
		return "", false
	}
	variant, exists := englishDeterminersForCandidate(candidate)
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

func englishSubjectNode(source semantic.MeaningCandidate, predicateID meaning.NodeID) meaning.NodeID {
	if source.Meaning == nil {
		return ""
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID == predicateID {
			for _, participant := range node.Participants {
				if participant.Role == "experiencer" || participant.Role == "reader" {
					return participant.Target
				}
			}
		}
	}
	return ""
}

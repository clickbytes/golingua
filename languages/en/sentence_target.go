package en

import (
	"context"
	"fmt"
	"reflect"
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
	"github.com/clickbytes/golingua/internal/valency"
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
	phraseProfile, err := SentencePhraseProfile(profile)
	if err != nil {
		return runtime.MorphologyExecutionBuilder{}, err
	}
	return runtime.MorphologyExecutionBuilder{
		Profile: profile, PhraseProfile: phraseProfile, Paradigms: paradigms, Exceptions: exceptions,
		Parts: seed.PrincipalParts, FeaturePlan: englishSentenceFeatures{profile: profile},
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
	negationResolver := englishNegationResolver{registry: constructions, valency: valencyIndex, lexicon: snapshot.Lexicon(), variants: variants}
	execution, err := NewSentenceExecutionBuilderWithSeed(seed)
	if err != nil {
		return nil, err
	}
	execution.FeaturePlan = englishSentenceFeatures{resolver: negationResolver, profile: profile}
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
		OperatorBuilder: englishSentenceOperatorBuilder{resolver: negationResolver}, ExecutionBuilder: execution,
		PlanBuilder: englishSentencePlanAnnotations{resolver: negationResolver}, CandidatePolicy: englishDeterminerCandidatePolicy{resolver: negationResolver},
		ClauseCoordination:   &coordination,
		ScopedRealizer:       &runtime.ScopedClauseRealizer{Language: "en", Plan: scopePlan, Valency: valencyIndex, Morphology: seed, Features: englishScopedFeaturePlanner{}},
		RelationPathRealizer: &runtime.RelationPathRealizer{Language: "en", Plan: relationPlan, Morphology: seed, Features: englishRelationPathFeaturePlanner{}},
	}, nil
}

type englishSentenceFeatures struct {
	resolver englishNegationResolver
	profile  *features.LanguageFeatureProfile
}

func (builder englishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
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
		_, valid := englishAttributedSeeingShape(source, builder.resolver)
		if !valid || source.Meaning == nil {
			return nil, fmt.Errorf("English attributed seeing adjective requires the licensed two-root Meaning")
		}
		modifier, modifierFound := realizer.FindModifier(*source.Meaning, predicate.NodeID, predicate.Family.Family, slot.ID)
		if !englishAttributedSeeingContextsValid(source, builder.resolver) || !modifierFound || candidate.NodeID != modifier.ID ||
			!englishAttributedSeeingModifierCandidate(source, candidate, candidates, builder.resolver) {
			return nil, fmt.Errorf("English attributed seeing adjective is not linked to its selected bearer property")
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
	if identification, recognized := englishTargetIdentificationState(source); recognized && identification.stateID == predicate.NodeID {
		if !identification.valid || !englishIdentificationAttributesAllowed(source, predicate.NodeID, identification.role, identification.referent, predicate.TransferBinding.CoreRef) {
			return nil, fmt.Errorf("English identification cannot realize the supplied Meaning attributes")
		}
		if slot.Category == construction.CategoryVP && predicate.Sense.ID.String() == identification.senseID {
			number := englishMultiplicityNumber(identification.multiplicity)
			if number == "" {
				number = "singular"
			}
			if selectedNumber, selected := englishAttributionDoorNumberChoice(source, predicate, "", ""); selected {
				number = selectedNumber
			}
			return englishFiniteVerbFeatures("present", "third", number), nil
		}
		if slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && string(slot.ArgumentBinding.SemanticRole) == identification.role {
			if !englishIdentificationMemberCountSupported(source, identification, []realizer.TargetCandidate{candidate}) {
				return nil, fmt.Errorf("English class membership requires an explicit count-noun target for its member referent")
			}
			number := ""
			mass, pluralOnly := englishNumberClass(candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
			if mass && !pluralOnly {
				number = "singular"
			} else {
				var ok bool
				number, _, ok = englishTargetNumber(source, candidate.NodeID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
				if !ok {
					return nil, fmt.Errorf("English class membership member number is incompatible with its lexical class")
				}
			}
			if selectedNumber, selected := englishAttributionDoorNumberChoice(source, predicate, slot.ID, string(slot.ArgumentBinding.SemanticRole)); selected {
				number = selectedNumber
			}
			return []features.FeatureValue{{Feature: "noun_number", State: features.StateKnown, Value: features.FeatureValueID(number)}, {Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)}}, nil
		}
	}
	if slot.Category == construction.CategoryAP {
		attribution, allowed := englishAttributionForSense(predicate.Sense.ID.String())
		if allowed && candidate.Sense.ID.String() == predicate.Sense.ID.String() && candidate.Lexeme.ID.String() == attribution.lexeme && candidate.TransferBinding.CoreRef.String() == attribution.concept {
			return knownFeature("adjective_degree", "positive"), nil
		}
	}
	if slot.Category == construction.CategoryVP {
		if _, _, _, found := englishTransitiveNegativeVPAuxContext(builder.resolver, source, predicate, candidates); found && englishVPAuxiliaryVariant(predicate) {
			_, event, _, _ := builder.resolver.negatedPresent(source)
			number, tense, ok := englishTransitiveNegativeVPFeatures(builder.resolver, source, predicate, candidates)
			if !ok || event.ID != predicate.NodeID {
				return nil, fmt.Errorf("English auxiliary negation requires exact event tense and subject number evidence")
			}
			return englishVPBaseFeatures(tense, number), nil
		}
		if _, event, route, found := builder.resolver.negatedPresent(source); found && event.ID == predicate.NodeID && predicate.Sense.ID.String() == route.sense {
			if route.statePredicate {
				if event.TemporalFrame == meaning.TemporalFramePast {
					return nil, fmt.Errorf("English negative identification does not support a past temporal frame")
				}
				return englishFiniteVerbFeatures("present", "third", "singular"), nil
			}
			number, known := builder.resolver.negationNumber(source, predicate, route, candidates)
			if !known {
				return nil, fmt.Errorf("English present negation cannot realize the supplied subject multiplicity")
			}
			verbFeatures := []features.FeatureValue{
				{Feature: "verb_form", State: features.StateKnown, Value: "imperative"},
				{Feature: "mood", State: features.StateKnown, Value: "imperative"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
				{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
			}
			if event.TemporalFrame == meaning.TemporalFramePast {
				verbFeatures = append(verbFeatures, features.FeatureValue{Feature: "tense", State: features.StateKnown, Value: "past"})
			}
			return verbFeatures, nil
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
			if predicate.Sense.ID.String() == "sense:en.give.transfer" && predicate.Family.Family.ID.String() == "construction:en.give.transfer.past" && predicate.Frame.ID == "en.give.transfer.past" {
				// Simple past gave is number-invariant; the separate past family
				// deliberately carries no subject-verb number agreement.
				return englishFiniteVerbFeatures("past", "", ""), nil
			}
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
	if englishVPAuxiliaryVariant(predicate) && slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && candidate.NodeID != predicate.NodeID {
		referent, exists := englishMeaningNode(source, candidate.NodeID)
		if exists && referent.Kind == meaning.KindReferent {
			switch englishLocalNPClass(candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID) {
			case "plural_only":
				number, semanticNumber, ok = "plural", false, true
			case "mass":
				if referent.Multiplicity == "" || referent.Multiplicity == meaning.MultiplicityOne {
					number, semanticNumber, ok = "singular", false, true
				}
			case "count":
				if referent.Multiplicity == "" {
					if values, found := runtime.SourceFeatures(source, slot.ArgumentBinding.SemanticRole, coreConcept(referent)); found {
						sourceNumber := string(values["number"])
						if sourceNumber == "" {
							sourceNumber = string(values["noun_number"])
						}
						if sourceNumber == "singular" || sourceNumber == "plural" {
							number, semanticNumber, ok = sourceNumber, false, true
						}
					}
				}
			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("English target lexical number class cannot realize the Meaning multiplicity")
	}
	if selectedNumber, selected := englishAttributedSeeingDoorNumberChoice(source, predicate, slot.ID, string(slot.ArgumentBinding.SemanticRole), builder.resolver); selected && candidate.NodeID == englishAttributedSeeingObjectID(source, builder.resolver) {
		number, semanticNumber = selectedNumber, false
	}
	if selectedNumber, selected := englishDoorActionObjectNumberChoice(source, predicate, slot.ArgumentBinding.SemanticRole); selected && candidate.NodeID == selectedDoorActionObjectReferent(source, predicate.NodeID) {
		number, semanticNumber = selectedNumber, false
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

func englishVPAuxiliaryVariant(candidate realizer.TargetCandidate) bool {
	return candidate.Family.Family.ID.String() == "construction:en.transitive_negative" &&
		candidate.CompiledVariant.Variant.ID.String() == "construction:en.transitive_negative.svo.vp_auxiliary" &&
		candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionRealize
}

func englishLegacyNegativeVPVariant(candidate realizer.TargetCandidate) bool {
	sequence := candidate.CompiledVariant.Variant.Sequence
	if candidate.Family.Family.ID.String() != "construction:en.transitive_negative" || len(sequence) < 6 {
		return false
	}
	seen := map[string]bool{}
	position := map[string]int{}
	for index, token := range sequence {
		if token.Kind == construction.TokenSlot {
			if token.SlotID != "subject" && token.SlotID != "verb" && token.SlotID != "object" || seen[token.SlotID] {
				return false
			}
			seen[token.SlotID], position[token.SlotID] = true, index
			continue
		}
		if token.Kind != construction.TokenTerminal {
			return false
		}
		terminal := strings.ToLower(token.Terminal)
		switch terminal {
		case "a", "an", "the":
		case "do", "does", "did":
			if seen["aux"] {
				return false
			}
			seen["aux"], position["aux"] = true, index
		case "not":
			if seen["not"] {
				return false
			}
			seen["not"], position["not"] = true, index
		case ".":
			if index != len(sequence)-1 {
				return false
			}
		default:
			return false
		}
	}
	return seen["subject"] && seen["verb"] && seen["object"] && seen["aux"] && seen["not"] &&
		position["subject"] < position["aux"] && position["aux"] < position["not"] && position["not"] < position["verb"] && position["verb"] < position["object"]
}

func englishTransitiveNegativeVPAuxContextForSource(resolver englishNegationResolver, source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	if candidate.Family.Family.ID.String() != "construction:en.transitive_negative" || candidate.Kind != meaning.KindEvent {
		return false
	}
	scope, event, route, found := resolver.negatedPresent(source)
	if !found || scope.ID == "" || event.ID != candidate.NodeID || route.family != candidate.Family.Family.ID.String() || route.sense != candidate.Sense.ID.String() ||
		(event.TemporalFrame != meaning.TemporalFramePresent && event.TemporalFrame != meaning.TemporalFramePast) {
		return false
	}
	referent, exists := englishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || referent.Multiplicity == "" || englishParticipantAnchorIntent(source, referent.ID) {
		return false
	}
	number := englishMultiplicityNumber(referent.Multiplicity)
	if number != "singular" && number != "plural" {
		return false
	}
	if _, _, ok := englishPredicateNumber(source, candidate, candidates); !ok {
		return false
	}
	return true
}

func englishTransitiveNegativeVPFeatures(resolver englishNegationResolver, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, string, bool) {
	scope, event, route, ok := englishTransitiveNegativeVPAuxContext(resolver, source, predicate, candidates)
	if !ok || scope.ID == "" || event.TemporalFrame != meaning.TemporalFramePresent && event.TemporalFrame != meaning.TemporalFramePast {
		return "", "", false
	}
	referent, exists := englishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || referent.Multiplicity != "" && !englishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, predicate.TransferBinding.CoreRef, true) {
		return "", "", false
	}
	meaningNumber := englishMultiplicityNumber(referent.Multiplicity)
	featureNumber := ""
	if values, found := runtime.SourceFeatures(source, route.subjectRole, coreConcept(referent)); found {
		if person := string(values["person"]); person != "" && person != "third" {
			return "", "", false
		}
		featureNumber = string(values["number"])
		if featureNumber == "" {
			featureNumber = string(values["verb_number"])
		}
	}
	class := ""
	for _, candidate := range candidates {
		if candidate.NodeID != referent.ID || candidate.Kind != meaning.KindReferent || candidate.Outcome != status.OutcomeOK || candidate.Lexeme.POS != "noun" {
			continue
		}
		candidateClass := englishLocalNPClass(candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if candidateClass == "" || class != "" && class != candidateClass {
			return "", "", false
		}
		class = candidateClass
	}
	switch class {
	case "plural_only":
		if featureNumber != "" && featureNumber != "plural" {
			return "", "", false
		}
		meaningNumber = "plural"
	case "mass":
		if referent.Multiplicity == meaning.MultiplicityMany || featureNumber != "" && featureNumber != "singular" {
			return "", "", false
		}
		meaningNumber = "singular"
	}
	if meaningNumber != "" && featureNumber != "" && meaningNumber != featureNumber || meaningNumber == "" && featureNumber == "" {
		return "", "", false
	}
	number := meaningNumber
	if number == "" {
		number = featureNumber
	}
	if number != "singular" && number != "plural" {
		return "", "", false
	}
	if englishParticipantAnchorIntent(source, referent.ID) {
		return "", "", false
	}
	for _, candidate := range candidates {
		if candidate.NodeID != referent.ID || candidate.Kind != meaning.KindReferent || candidate.Outcome != status.OutcomeOK || candidate.Lexeme.POS != "noun" {
			continue
		}
		candidateNumber, _, known := englishTargetNumber(source, referent.ID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if englishLocalNPClass(candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID) == "plural_only" {
			candidateNumber, known = "plural", true
		}
		if known && candidateNumber == number {
			if event.TemporalFrame == meaning.TemporalFramePresent {
				return number, "present", true
			}
			return number, "past", true
		}
	}
	return "", "", false
}

func englishTransitiveNegativeVPAuxContext(resolver englishNegationResolver, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (meaning.Node, meaning.Node, englishPresentNegationRoute, bool) {
	if !englishVPAuxiliaryVariant(predicate) || predicate.Kind != meaning.KindEvent || predicate.Frame.ID == "" ||
		!englishPredicateEvidenceConsistent(source, predicate) {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	scope, event, route, found := resolver.negatedPresent(source)
	if !found || event.ID != predicate.NodeID || route.family != "construction:en.transitive_negative" || route.sense != predicate.Sense.ID.String() || route.frame != string(predicate.Frame.ID) {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	if resolver.registry == nil {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	approved := false
	for _, binding := range resolver.registry.BindingsByFamilyID(predicate.Family.Family.ID) {
		if binding.Approved && binding.SenseID == predicate.Sense.ID && binding.ValencyFrameID == predicate.Frame.ID && binding.VariantID.String() == predicate.CompiledVariant.Variant.ID.String() {
			approved = true
			break
		}
	}
	if !approved || !variantHasRequiredSlots(predicate.Family.Family, predicate.CompiledVariant.Variant) ||
		!sequenceHasSlot(predicate.CompiledVariant.Variant.Sequence, "subject") || !sequenceHasSlot(predicate.CompiledVariant.Variant.Sequence, "verb") || !sequenceHasSlot(predicate.CompiledVariant.Variant.Sequence, "object") {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	if _, _, ok := englishTransitiveNegativeVPFeaturesNoCandidates(source, event, route); !ok {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	return scope, event, route, true
}

func englishTransitiveNegativeVPFeaturesNoCandidates(source semantic.MeaningCandidate, event meaning.Node, route englishPresentNegationRoute) (string, string, bool) {
	if event.TemporalFrame != meaning.TemporalFramePresent && event.TemporalFrame != meaning.TemporalFramePast {
		return "", "", false
	}
	referent, exists := englishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || referent.Multiplicity == "" || englishParticipantAnchorIntent(source, referent.ID) {
		return "", "", false
	}
	if values, found := runtime.SourceFeatures(source, route.subjectRole, coreConcept(referent)); found {
		if person := string(values["person"]); person != "" && person != "third" {
			return "", "", false
		}
	}
	number := englishMultiplicityNumber(referent.Multiplicity)
	if number != "singular" && number != "plural" {
		return "", "", false
	}
	if event.TemporalFrame == meaning.TemporalFramePresent {
		return number, "present", true
	}
	return number, "past", true
}

func englishVPBaseFeatures(tense, number string) []features.FeatureValue {
	return []features.FeatureValue{
		{Feature: "verb_form", State: features.StateKnown, Value: "base"},
		{Feature: "tense", State: features.StateKnown, Value: features.FeatureValueID(tense)},
		{Feature: "mood", State: features.StateKnown, Value: "indicative"},
		{Feature: "voice", State: features.StateKnown, Value: "active"},
		{Feature: "person", State: features.StateKnown, Value: "third"},
		{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
		{Feature: "verb_number", State: features.StateKnown, Value: features.FeatureValueID(number)},
	}
}

func (builder englishSentenceFeatures) BuildVPAuxiliary(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, _ realizer.TargetCandidate, candidates []realizer.TargetCandidate) (*realizer.VPAuxiliaryExecutionRequest, error) {
	if !slot.IsMainPredicate() || !englishVPAuxiliaryVariant(predicate) {
		return nil, nil
	}
	_, event, _, ok := englishTransitiveNegativeVPAuxContext(builder.resolver, source, predicate, candidates)
	if !ok || builder.profile == nil {
		return nil, fmt.Errorf("English auxiliary negation requires an approved exact binding, scoped Not, and explicit event features")
	}
	number, tense, ok := englishTransitiveNegativeVPFeatures(builder.resolver, source, predicate, candidates)
	if !ok {
		return nil, fmt.Errorf("English auxiliary negation requires known subject number and event tense")
	}
	auxID, err := ids.ParseLexemeID("lexeme:en.do.aux")
	if err != nil {
		return nil, err
	}
	auxLexeme, exists := builder.resolver.lexicon.LookupLexeme(auxID)
	if !exists {
		return nil, fmt.Errorf("English nonsemantic do auxiliary is unavailable")
	}
	finiteValues := englishFiniteVerbFeatures(tense, "third", number)
	if tense == "past" {
		finiteValues = englishFiniteVerbFeatures("past", "", "")
	}
	auxBundle, err := features.NewBundle(builder.profile, finiteValues)
	if err != nil {
		return nil, err
	}
	baseValues := englishVPBaseFeatures(tense, number)
	baseBundle, err := features.NewBundle(builder.profile, baseValues)
	if err != nil {
		return nil, err
	}
	return &realizer.VPAuxiliaryExecutionRequest{
		EventNodeID: event.ID, Lexeme: auxLexeme, Features: auxBundle, BaseFeatures: baseBundle,
		NegationParticle: &realizer.VPAuxiliaryParticleRequest{OperatorID: "negate", Surface: "not", Provenance: englishNegationConstructionSource},
	}, nil
}

func englishAttributedSeeingObjectID(source semantic.MeaningCandidate, resolver englishNegationResolver) meaning.NodeID {
	shape, valid := englishAttributedSeeingShape(source, resolver)
	if !valid {
		return ""
	}
	return shape.object.ID
}

type englishSentenceOperatorBuilder struct{ resolver englishNegationResolver }

func (builder englishSentenceOperatorBuilder) Build(ctx context.Context, source semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
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
		if _, _, _, negated := builder.resolver.negatedPresent(source); negated {
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
	scope, event, route, found := builder.resolver.negatedPresent(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyName := route.family
	if route.concept == "concept:seeing" {
		if shape, valid := englishAttributedSeeingShape(source, builder.resolver); valid && shape.scope.ID == scope.ID {
			familyName = "construction:en.attributed_seeing_object_negative"
		}
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(familyName, "construction:"))
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

type englishSentencePlanAnnotations struct{ resolver englishNegationResolver }

const englishGivingLocalNPVariant = "construction:en.give.transfer.past.double_np.np_local_subject_recipient_theme"
const englishGivingLocalToPPNPVariant = "construction:en.give.transfer.past.to_pp.np_local_subject_theme_recipient"
const englishGivingNegativeLocalToPPNPVariant = "construction:en.ditransitive_negative.past.to_pp.np_local_subject_theme_recipient"

func englishGivingLocalNPVariantID(id string) bool {
	return id == englishGivingLocalNPVariant || id == englishGivingLocalToPPNPVariant || id == englishGivingNegativeLocalToPPNPVariant
}

// Independent Meaning input may provide a structurally valid graph with no
// discourse mentions. For a plain attribution, that licenses the bare route;
// it does not invent demonstrative deixis.
func englishTargetIdentificationState(source semantic.MeaningCandidate) (englishIdentification, bool) {
	if negative, ok := englishNegativeAttributionState(source); ok {
		return negative, true
	}
	identification, recognized := englishIdentificationState(source)
	if recognized && identification.valid {
		return identification, true
	}
	if source.Meaning == nil || len(source.Contexts) == 0 {
		return identification, recognized
	}
	for _, item := range source.Contexts {
		if meaning.ValidateReferences(item.Meaning, item.Context, meaning.DefaultBudget()) != nil || len(item.Context.Referents) != 0 || len(item.Context.Mentions) != 0 {
			return identification, recognized
		}
		equal, err := meaning.CanonicalEqual(item.Meaning, *source.Meaning, meaning.DefaultBudget())
		if err != nil || !equal {
			return identification, recognized
		}
	}
	graph := *source.Meaning
	if meaning.ValidateStructure(graph, meaning.DefaultBudget()) != nil || len(graph.Roots) != 1 || len(graph.Nodes) != 2 {
		return identification, recognized
	}
	for _, node := range graph.Nodes {
		if node.Kind != meaning.KindState || node.ID != graph.Roots[0] || len(node.Participants) != 1 || node.Participants[0].Role != "bearer" {
			continue
		}
		if _, ok := englishAttributionForConcept(string(node.Concept)); !ok {
			continue
		}
		referent := findMeaningNode(graph, node.Participants[0].Target)
		if referent.Kind != meaning.KindReferent || referent.Concept == "" {
			continue
		}
		attribution, _ := englishAttributionForConcept(string(node.Concept))
		return englishIdentification{
			stateID: node.ID, referent: referent.ID, role: "bearer", concept: referent.Concept,
			senseID: attribution.sense, family: "construction:en.attribution", mode: "none",
			temporalFrame: node.TemporalFrame, multiplicity: referent.Multiplicity, valid: true,
		}, true
	}
	return identification, recognized
}

func (builder englishSentencePlanAnnotations) Annotations(_ context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error) {
	if selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState || source.Meaning == nil {
		return nil, nil, nil, nil
	}
	predicate := findMeaningNode(*source.Meaning, selected.NodeID)
	if predicate.ID == "" {
		return nil, nil, nil, nil
	}
	if selected.TransferBinding.SenseID.String() == "sense:en.go.spatial_movement" {
		path, found := runtime.MeaningPath(*source.Meaning, predicate.ID)
		if !found {
			return nil, nil, nil, fmt.Errorf("English go target is missing its Meaning path")
		}
		// If the source was Polish iść, the walking sense contributes manner
		// not encoded by English go. Preserve that information loss explicitly.
		for _, record := range source.Evidence.Records {
			if (record.SenseID.String() == "sense:pl.isc.go" || record.SenseID.String() == "sense:pl.isc.move") && record.Origin != semantic.EvidenceOriginGenerated {
				loss := realizer.FieldLoss{Path: path, Field: "manner", Outcome: status.OutcomeIncomplete, Reason: "Polish iść specifies on-foot movement; English go does not encode that manner"}
				return nil, nil, []realizer.FieldLoss{loss}, nil
			}
		}
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
	if identification, recognized := englishTargetIdentificationState(source); recognized && identification.stateID == predicate.ID {
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
		if identification.mode == "none" && identification.family == "construction:en.attribution_negative" {
			variant := selected.CompiledVariant.Variant.ID.String()
			determiner := ""
			bearerFeatures, bearerFeaturesFound := runtime.SourceFeatures(source, meaning.Role(identification.role), coreConcept(referent))
			sourceDeterminer := ""
			if bearerFeaturesFound {
				sourceDeterminer = string(bearerFeatures["determiner"])
			}
			switch {
			case strings.HasSuffix(variant, ".the"):
				determiner = "definite"
			case strings.HasSuffix(variant, ".a") || strings.HasSuffix(variant, ".an"):
				determiner = "indefinite"
			case strings.HasSuffix(variant, ".bare") && sourceDeterminer == "":
				determiner = "bare"
			}
			if determiner != "" {
				reason := "Meaning has no determiner evidence; this English negative attribution candidate selects an article"
				if determiner == "bare" {
					reason = "Meaning has no determiner evidence; this English negative attribution candidate selects a zero article"
				}
				assumptions = append(assumptions, realizer.FieldAssumption{Path: referentPath, Field: "determiner", Value: determiner, Reason: reason})
			}
		}
		if identification.mode == "none" && identification.family == "construction:en.attribution" {
			variant := strings.TrimSuffix(selected.CompiledVariant.Variant.ID.String(), ".past")
			determiner := ""
			switch variant {
			case identification.family + ".the", identification.family + ".the_plural":
				determiner = "definite"
			case identification.family + ".bare_plural":
				determiner = "bare"
			case identification.family + ".a", identification.family + ".an":
				determiner = "indefinite"
			}
			if determiner != "" {
				assumptions = append(assumptions, realizer.FieldAssumption{Path: referentPath, Field: "determiner", Value: determiner, Reason: "Polish source has no determiner evidence; this English candidate selects a determiner value"})
			}
		}
		if number, choice := englishAttributionDoorNumberChoice(source, selected, "", ""); choice && number == "plural" && referent.Multiplicity == "" {
			filtered := assumptions[:0]
			for _, assumption := range assumptions {
				if assumption.Field == "number" && (sameMeaningNodePath(assumption.Path, predicatePath) || sameMeaningNodePath(assumption.Path, referentPath)) {
					continue
				}
				filtered = append(filtered, assumption)
			}
			assumptions = append(filtered,
				realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "plural", Reason: "Polish drzwi does not specify cardinality; this English candidate explicitly selects plural agreement"},
				realizer.FieldAssumption{Path: referentPath, Field: "number", Value: "plural", Reason: "Polish drzwi does not specify cardinality; this English candidate explicitly selects plural number"},
			)
		}
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
			invariantGivingPast := englishInvariantGivingPastTarget(selected, predicate)
			if (!anchored || anchor.Value != "addressee") && !invariantGivingPast {
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
			if selectedNumber, openingChoice := englishDoorActionObjectNumberChoice(source, selected, participant.Role); openingChoice && participant.Role == "object" && referent.Concept == "concept:door" {
				value = selectedNumber
				reason = "Meaning leaves door cardinality unspecified; this English opening candidate selects " + selectedNumber + " number"
			}
			if selectedNumber, doorChoice := englishAttributedSeeingDoorNumberChoice(source, selected, "object", string(participant.Role), builder.resolver); doorChoice {
				value = selectedNumber
				reason = "Meaning leaves door number unspecified; this English candidate explicitly selects " + selectedNumber + " number"
			}
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
		if participant.Role == "experiencer" && (selected.Family.Family.ID.String() == "construction:en.attributed_seeing" || selected.Family.Family.ID.String() == "construction:en.attributed_seeing_both") {
			anchors := englishAttributedSeeingSubjectAnchors(source, referent.ID)
			if len(anchors) != 0 {
				value := "distal"
				if article == "this" || article == "these" {
					value = "proximal"
				}
				_, isProximal := anchors["proximal"]
				_, isDistal := anchors["distal"]
				if anchors["demonstrative"] || isProximal && isDistal {
					assumptions = append(assumptions, realizer.FieldAssumption{Path: referentPath, Field: "reference", Value: value, Reason: "source context leaves demonstrative distance unresolved; this candidate selects one target deixis"})
				}
				continue
			}
		}
		if participant.Role == "stimulus" && selected.Family.Family.ID.String() == "construction:en.attributed_seeing_object" {
			deixis := englishAttributedSeeingDeixis(source, builder.resolver)
			if deixis != "" {
				if deixis == "demonstrative" || deixis == "ambiguous" {
					value := "distal"
					if article == "this" || article == "these" {
						value = "proximal"
					}
					assumptions = append(assumptions, realizer.FieldAssumption{Path: referentPath, Field: "reference", Value: value, Reason: "source context leaves demonstrative distance unresolved; this candidate selects one target deixis"})
				}
				continue
			}
		}
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
	if englishSimplePastEventTarget(builder.resolver, source, selected, predicate) {
		if _, known := englishSourcePredicateAspect(source, predicate); known {
			losses = append(losses, realizer.FieldLoss{Path: predicatePath, Field: "aspect", Outcome: status.OutcomeIncomplete, Reason: "Polish lexical aspect is explicit in the source, while the selected English simple-past form does not distinguish it"})
		}
	}
	return nil, assumptions, losses, nil
}

func (builder englishSentencePlanAnnotations) AnnotationAlternatives(ctx context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]runtime.TargetPlanAnnotations, error) {
	constraints, assumptions, losses, err := builder.Annotations(ctx, source, selected)
	if err != nil {
		return nil, err
	}
	base := runtime.TargetPlanAnnotations{FeatureConstraints: constraints, Assumptions: assumptions, Losses: losses}
	if englishVPAuxiliaryVariant(selected) {
		return englishTransitiveNegativeLocalArticleAlternatives(builder.resolver, source, selected, base)
	}
	if !englishGivingLocalNPVariantID(selected.CompiledVariant.Variant.ID.String()) {
		return []runtime.TargetPlanAnnotations{base}, nil
	}
	negativeRoute := selected.Frame.ID == "en.give.transfer.negative" && selected.Family.Family.ID.String() == "construction:en.ditransitive_negative"
	positiveRoute := selected.Frame.ID == "en.give.transfer.past" && selected.Family.Family.ID.String() == "construction:en.give.transfer.past"
	writingMessagePastRoute := selected.Sense.ID.String() == "sense:en.writing_message.event" && selected.Frame.ID == "en.writing_message.full" &&
		selected.Family.Family.ID.String() == "construction:en.give.transfer.past" && selected.CompiledVariant.Variant.ID.String() == englishGivingLocalToPPNPVariant
	if source.Meaning == nil || selected.Kind != meaning.KindEvent ||
		(!positiveRoute && !negativeRoute && !writingMessagePastRoute) ||
		selected.Sense.ID.String() != "sense:en.give.transfer" && !writingMessagePastRoute ||
		selected.TransferBinding.CoreRef.String() != "concept:giving" && !writingMessagePastRoute {
		return nil, fmt.Errorf("NP-local determiner alternatives require the exact approved past giving candidate")
	}
	event, exists := englishMeaningNode(source, selected.NodeID)
	wantConcept := meaning.ConceptRef("concept:giving")
	if writingMessagePastRoute {
		wantConcept = "concept:writing_message"
	}
	if !exists || event.Kind != meaning.KindEvent || event.Concept != wantConcept || event.TemporalFrame != meaning.TemporalFramePast {
		return nil, fmt.Errorf("NP-local determiner alternatives require a compatible past Meaning event")
	}
	if negativeRoute {
		_, negatedEvent, route, negated := builder.resolver.negatedPresent(source)
		if !negated || negatedEvent.ID != event.ID || route.sense != selected.Sense.ID.String() || !englishSimplePastGivingNegativeVariant(selected.CompiledVariant.Variant.ID.String()) {
			return nil, fmt.Errorf("NP-local negative giving determiner alternatives require the exact negated past event")
		}
	}
	roleSlots, ok := englishRequiredFrameRoleSlots(selected)
	if writingMessagePastRoute {
		roleSlots, ok = englishWritingMessagePastRoleSlots(selected)
	}
	if !ok {
		return nil, fmt.Errorf("NP-local determiner alternatives require complete frame slot bindings")
	}
	type localRoleChoice struct {
		path   []meaning.NodeID
		values []string
		assume bool
	}
	localChoices := make([]localRoleChoice, 0, 3)
	for _, localSlot := range []string{"subject", "recipient", "theme"} {
		role := meaning.Role("")
		for candidateRole, slotID := range roleSlots {
			if slotID == localSlot {
				if role != "" {
					return nil, fmt.Errorf("NP-local determiner alternatives have multiple roles for %s", localSlot)
				}
				role = candidateRole
			}
		}
		if role == "" {
			return nil, fmt.Errorf("NP-local determiner alternatives have no frame role for %s", localSlot)
		}
		var referent meaning.Node
		for _, participant := range event.Participants {
			if participant.Role == role {
				if referent.ID != "" {
					return nil, fmt.Errorf("NP-local determiner alternatives require one %s referent", localSlot)
				}
				referent, _ = englishMeaningNode(source, participant.Target)
			}
		}
		if referent.ID == "" {
			return nil, fmt.Errorf("NP-local determiner alternatives have no exact %s referent", localSlot)
		}
		path, found := runtime.MeaningPath(*source.Meaning, referent.ID)
		if !found {
			return nil, fmt.Errorf("NP-local determiner alternatives have no %s Meaning path", localSlot)
		}
		sourceDeterminer := ""
		if values, has := runtime.SourceFeatures(source, role, coreConcept(referent)); has {
			sourceDeterminer = string(values["determiner"])
		}
		choice := localRoleChoice{path: path, assume: sourceDeterminer == ""}
		switch sourceDeterminer {
		case "":
			choice.values = []string{"indefinite", "definite"}
		case "bare":
			if !writingMessagePastRoute {
				return nil, fmt.Errorf("NP-local English %s article cannot preserve source determiner value %q", localSlot, sourceDeterminer)
			}
			choice.values = []string{"indefinite", "definite"}
			choice.assume = true
		case "indefinite", "definite":
			choice.values = []string{sourceDeterminer}
		default:
			return nil, fmt.Errorf("NP-local English %s article cannot preserve source determiner value %q", localSlot, sourceDeterminer)
		}
		localChoices = append(localChoices, choice)
	}
	result := make([]runtime.TargetPlanAnnotations, 0, len(localChoices[0].values)*len(localChoices[1].values)*len(localChoices[2].values))
	for _, first := range localChoices[0].values {
		for _, second := range localChoices[1].values {
			for _, third := range localChoices[2].values {
				selected := runtime.TargetPlanAnnotations{
					FeatureConstraints: append([]realizer.FeatureConstraint(nil), base.FeatureConstraints...),
					Assumptions:        cloneEnglishAssumptions(base.Assumptions),
					Losses:             append([]realizer.FieldLoss(nil), base.Losses...),
				}
				for index, choice := range []string{first, second, third} {
					local := localChoices[index]
					if local.assume {
						reason := "source Meaning has no determiner evidence; this English target candidate selects the " + choice + " determiner"
						selected.Assumptions = append(selected.Assumptions, realizer.FieldAssumption{Path: append([]meaning.NodeID(nil), local.path...), Field: "determiner", Value: choice, Reason: reason})
					}
				}
				result = append(result, selected)
			}
		}
	}
	return result, nil
}

func englishTransitiveNegativeLocalArticleAlternatives(resolver englishNegationResolver, source semantic.MeaningCandidate, selected realizer.TargetCandidate, base runtime.TargetPlanAnnotations) ([]runtime.TargetPlanAnnotations, error) {
	if source.Meaning == nil || selected.Kind != meaning.KindEvent || selected.Family.Family.ID.String() != "construction:en.transitive_negative" ||
		!englishVPAuxiliaryVariant(selected) {
		return nil, fmt.Errorf("NP-local negative articles require the exact approved transitive-negative variant")
	}
	event, exists := englishMeaningNode(source, selected.NodeID)
	if !exists || event.Kind != meaning.KindEvent || (event.TemporalFrame != meaning.TemporalFramePresent && event.TemporalFrame != meaning.TemporalFramePast) {
		return nil, fmt.Errorf("NP-local negative articles require an explicitly present or past event")
	}
	roleSlots, ok := englishRequiredFrameRoleSlots(selected)
	if !ok {
		return nil, fmt.Errorf("NP-local negative articles require exact frame slot roles")
	}
	choices := make([]englishLocalDeterminerChoice, 0, 2)
	for _, slotID := range []string{"subject", "object"} {
		role := meaning.Role("")
		for candidateRole, candidateSlot := range roleSlots {
			if candidateSlot == slotID {
				if role != "" {
					return nil, fmt.Errorf("NP-local negative articles have multiple roles for slot %s", slotID)
				}
				role = candidateRole
			}
		}
		if role == "" {
			return nil, fmt.Errorf("NP-local negative articles lack required role for slot %s", slotID)
		}
		var referent meaning.Node
		for _, participant := range event.Participants {
			if participant.Role == role {
				if referent.ID != "" {
					return nil, fmt.Errorf("NP-local negative articles require one referent for role %s", role)
				}
				referent, _ = englishMeaningNode(source, participant.Target)
			}
		}
		if referent.ID == "" || referent.Kind != meaning.KindReferent || englishParticipantAnchorIntent(source, referent.ID) {
			return nil, fmt.Errorf("NP-local negative articles require an unanchored referent for role %s", role)
		}
		conceptID, conceptErr := ids.ParseConceptID(string(referent.Concept))
		if conceptErr != nil {
			return nil, fmt.Errorf("NP-local negative articles require a valid target concept for role %s", role)
		}
		class, classOK := englishLocalNPClassForConcept(resolver.lexicon, conceptID)
		if !classOK {
			return nil, fmt.Errorf("NP-local negative articles require a consistent approved English noun number class for role %s", role)
		}
		path, found := runtime.MeaningPath(*source.Meaning, referent.ID)
		if !found {
			return nil, fmt.Errorf("NP-local negative articles require an exact Meaning path for role %s", role)
		}
		determiner := ""
		if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found {
			determiner = string(values["determiner"])
		}
		number := ""
		switch class {
		case "count":
			number = englishMultiplicityNumber(referent.Multiplicity)
			if number == "" {
				if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found {
					number = string(values["number"])
					if number == "" {
						number = string(values["noun_number"])
					}
				}
			}
			if number != "singular" && number != "plural" {
				return nil, fmt.Errorf("NP-local negative articles require explicit singular or plural evidence for count role %s", role)
			}
		case "mass":
			if referent.Multiplicity != "" && referent.Multiplicity != meaning.MultiplicityOne {
				return nil, fmt.Errorf("NP-local mass noun route cannot realize explicit plural cardinality for role %s", role)
			}
			number = "singular"
		case "plural_only":
			number = "plural"
		default:
			return nil, fmt.Errorf("NP-local negative articles lack a usable number class for role %s", role)
		}
		item := englishLocalDeterminerChoice{path: path, role: role, assume: determiner == ""}
		choicesForClass := []string{"bare", "definite"}
		if class == "count" && number == "singular" {
			choicesForClass = []string{"indefinite", "definite"}
		}
		switch determiner {
		case "":
			item.values = choicesForClass
		case "indefinite":
			if class != "count" || number != "singular" {
				return nil, fmt.Errorf("NP-local English indefinite determiner is incompatible with role %s lexical number class", role)
			}
			item.values = []string{determiner}
		case "definite":
			item.values = []string{determiner}
		case "bare":
			if class == "count" && number != "plural" {
				return nil, fmt.Errorf("NP-local English bare determiner is incompatible with singular count role %s", role)
			}
			item.values = []string{determiner}
		default:
			return nil, fmt.Errorf("NP-local English determiner cannot preserve source value %q for role %s", determiner, role)
		}
		if class == "plural_only" && referent.Multiplicity != "" {
			base.Losses = append(base.Losses, realizer.FieldLoss{Path: path, Field: "number", Outcome: status.OutcomeIncomplete, Reason: "the selected English plural-only noun form does not distinguish the referent's explicit semantic cardinality"})
		}
		choices = append(choices, item)
	}
	if len(choices) != 2 || len(choices[0].values)*len(choices[1].values) > 4 {
		return nil, fmt.Errorf("NP-local negative article choices exceed the two-role bound")
	}
	base.Assumptions = filterDeterminerAnnotations(base.Assumptions, choices)
	base.Losses = filterDeterminerLosses(base.Losses, choices)
	for _, choice := range choices {
		base.Assumptions = filterRoleNumberAssumption(base.Assumptions, choice.path)
	}
	result := make([]runtime.TargetPlanAnnotations, 0, len(choices[0].values)*len(choices[1].values))
	for _, subjectChoice := range choices[0].values {
		for _, objectChoice := range choices[1].values {
			selectedAnnotations := runtime.TargetPlanAnnotations{
				FeatureConstraints: append([]realizer.FeatureConstraint(nil), base.FeatureConstraints...),
				Assumptions:        cloneEnglishAssumptions(base.Assumptions),
				Losses:             append([]realizer.FieldLoss(nil), base.Losses...),
			}
			for index, value := range []string{subjectChoice, objectChoice} {
				if choices[index].assume {
					selectedAnnotations.Assumptions = append(selectedAnnotations.Assumptions, realizer.FieldAssumption{
						Path: append([]meaning.NodeID(nil), choices[index].path...), Field: "determiner", Value: value,
						Reason: "source Meaning does not specify English definiteness; this target alternative selects the " + value + " determiner",
					})
				}
			}
			result = append(result, selectedAnnotations)
		}
	}
	return result, nil
}

func englishLocalNPClass(properties []lexicon.Property, paradigmID string) string {
	mass, pluralOnly := englishNumberClass(properties, paradigmID)
	if mass && pluralOnly {
		return ""
	}
	if mass {
		return "mass"
	}
	if pluralOnly {
		return "plural_only"
	}
	return "count"
}

func englishLocalNPClassForConcept(index *lexicon.Index, concept ids.ConceptID) (string, bool) {
	if index == nil || concept.String() == "" {
		return "", false
	}
	class := ""
	found := false
	for _, binding := range index.ByCoreConcept(concept) {
		if !binding.TransferAuthorized() {
			continue
		}
		sense, exists := index.BySenseID(binding.SenseID)
		if !exists {
			continue
		}
		lexeme, exists := index.LookupLexeme(sense.LexemeID)
		if !exists || lexeme.Language != "en" || lexeme.POS != "noun" {
			continue
		}
		candidateClass := englishLocalNPClass(lexeme.Properties, lexeme.ParadigmID)
		if candidateClass == "" || found && candidateClass != class {
			return "", false
		}
		class, found = candidateClass, true
	}
	return class, found
}

type englishLocalDeterminerChoice struct {
	path   []meaning.NodeID
	role   meaning.Role
	values []string
	assume bool
}

func englishTransitiveNegativeLocalNPCandidateAllowed(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	if source.Meaning == nil || candidate.Kind != meaning.KindEvent || candidate.Outcome != status.OutcomeOK ||
		candidate.Family.Family.ID.String() != "construction:en.transitive_negative" || !englishVPAuxiliaryVariant(candidate) ||
		candidate.Frame.ID == "" || !englishPredicateEvidenceConsistent(source, candidate) {
		return false
	}
	event, exists := englishMeaningNode(source, candidate.NodeID)
	if !exists || event.Kind != meaning.KindEvent || (event.TemporalFrame != meaning.TemporalFramePresent && event.TemporalFrame != meaning.TemporalFramePast) {
		return false
	}
	roleSlots, ok := englishRequiredFrameRoleSlots(candidate)
	if !ok {
		return false
	}
	for _, slotID := range []string{"subject", "object"} {
		role := meaning.Role("")
		for candidateRole, candidateSlot := range roleSlots {
			if candidateSlot == slotID {
				if role != "" {
					return false
				}
				role = candidateRole
			}
		}
		if role == "" {
			return false
		}
		var referent meaning.Node
		for _, participant := range event.Participants {
			if participant.Role == role {
				if referent.ID != "" {
					return false
				}
				referent, _ = englishMeaningNode(source, participant.Target)
			}
		}
		if referent.Kind != meaning.KindReferent || englishParticipantAnchorIntent(source, referent.ID) {
			return false
		}
		class, number := "", ""
		lexicalCount, articleAvailable := 0, true
		for _, available := range candidates {
			if available.SourceCandidateID != candidate.SourceCandidateID || available.SemanticID != candidate.SemanticID ||
				available.NodeID != referent.ID || available.Kind != meaning.KindReferent || available.Outcome != status.OutcomeOK ||
				available.TransferBinding.CoreRef.String() != string(referent.Concept) || available.Lexeme.POS != "noun" || !available.TransferBinding.TransferAuthorized() {
				continue
			}
			lexicalCount++
			candidateClass := englishLocalNPClass(available.Lexeme.Properties, available.Lexeme.ParadigmID)
			if candidateClass == "" || class != "" && class != candidateClass {
				return false
			}
			class = candidateClass
			if candidateClass == "count" {
				if _, known := englishIndefiniteArticle(available.Lexeme.Properties); !known {
					articleAvailable = false
				}
			}
		}
		if lexicalCount == 0 || class == "" {
			return false
		}
		switch class {
		case "count":
			number = englishMultiplicityNumber(referent.Multiplicity)
			if number == "" {
				if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found {
					number = string(values["number"])
					if number == "" {
						number = string(values["noun_number"])
					}
				}
			}
			if number != "singular" && number != "plural" || number == "singular" && !articleAvailable {
				return false
			}
		case "mass":
			if referent.Multiplicity != "" && referent.Multiplicity != meaning.MultiplicityOne {
				return false
			}
			number = "singular"
		case "plural_only":
			number = "plural"
		default:
			return false
		}
		if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found {
			determiner := string(values["determiner"])
			if determiner != "" && determiner != "definite" && determiner != "indefinite" && determiner != "bare" {
				return false
			}
			if determiner == "indefinite" && (class != "count" || number != "singular" || !articleAvailable) || determiner == "bare" && class == "count" && number != "plural" {
				return false
			}
		}
	}
	return true
}

func filterDeterminerAnnotations(values []realizer.FieldAssumption, choices []englishLocalDeterminerChoice) []realizer.FieldAssumption {
	result := make([]realizer.FieldAssumption, 0, len(values))
	for _, value := range values {
		remove := false
		if value.Field == "determiner" {
			for _, choice := range choices {
				if sameMeaningNodePath(value.Path, choice.path) {
					remove = true
					break
				}
			}
		}
		if !remove {
			result = append(result, value)
		}
	}
	return result
}

func filterRoleNumberAssumption(values []realizer.FieldAssumption, path []meaning.NodeID) []realizer.FieldAssumption {
	result := make([]realizer.FieldAssumption, 0, len(values))
	for _, value := range values {
		if value.Field == "number" && sameMeaningNodePath(value.Path, path) {
			continue
		}
		result = append(result, value)
	}
	return result
}

func filterDeterminerLosses(values []realizer.FieldLoss, choices []englishLocalDeterminerChoice) []realizer.FieldLoss {
	result := make([]realizer.FieldLoss, 0, len(values))
	for _, value := range values {
		remove := false
		if value.Field == "determiner" {
			for _, choice := range choices {
				if sameMeaningNodePath(value.Path, choice.path) {
					remove = true
					break
				}
			}
		}
		if !remove {
			result = append(result, value)
		}
	}
	return result
}

func cloneEnglishAssumptions(values []realizer.FieldAssumption) []realizer.FieldAssumption {
	result := make([]realizer.FieldAssumption, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Path = append([]meaning.NodeID(nil), value.Path...)
	}
	return result
}

func englishSimplePastEventTarget(resolver englishNegationResolver, source semantic.MeaningCandidate, selected realizer.TargetCandidate, predicate meaning.Node) bool {
	if selected.Kind != meaning.KindEvent || predicate.Kind != meaning.KindEvent || selected.NodeID != predicate.ID || predicate.TemporalFrame != meaning.TemporalFramePast ||
		!englishApprovedPastEventCandidate(resolver, selected, predicate) {
		return false
	}
	family := selected.Family.Family.ID.String()
	variant := selected.CompiledVariant.Variant.ID.String()
	if family == "construction:en.intransitive" {
		return englishPastNegativeScopeMatches(resolver, source, selected, predicate) ||
			!englishMeaningHasScope(source, predicate.ID) && !sequenceHasTerminal(selected.CompiledVariant.Variant.Sequence, "not")
	}
	if family == "construction:en.give.transfer.past" && selected.Sense.ID.String() == "sense:en.give.transfer" && selected.Frame.ID == "en.give.transfer.past" && englishSimplePastGivingVariant(variant) {
		return !englishMeaningHasScope(source, predicate.ID)
	}
	if family == "construction:en.give.transfer.past" && selected.Sense.ID.String() == "sense:en.writing_message.event" &&
		selected.Frame.ID == "en.writing_message.full" && variant == englishGivingLocalToPPNPVariant {
		_, valid := englishWritingMessagePastRoleSlots(selected)
		return valid && !englishMeaningHasScope(source, predicate.ID)
	}
	if family == "construction:en.transitive_past" {
		return !englishMeaningHasScope(source, predicate.ID)
	}
	if family == "construction:en.ditransitive_negative" && selected.Sense.ID.String() == "sense:en.give.transfer" && selected.Frame.ID == "en.give.transfer.negative" && englishSimplePastGivingNegativeVariant(variant) {
		_, event, route, negated := resolver.negatedPresent(source)
		return negated && event.ID == predicate.ID && event.TemporalFrame == meaning.TemporalFramePast &&
			route.concept == "concept:giving" && route.sense == selected.Sense.ID.String() && route.pastVariantAllowed(variant)
	}
	if family == "construction:en.transitive_negative" || family == "construction:en.oblique_negative" {
		return englishPastNegativeScopeMatches(resolver, source, selected, predicate)
	}
	return false
}

// englishApprovedPastEventCandidate derives simple-past eligibility from the
// exact target binding, compiled family variant, predicate slot, and selected
// frame. Dedicated past families may encode tense in the family; any explicit
// verb-tense condition must agree with the Meaning.
func englishApprovedPastEventCandidate(resolver englishNegationResolver, selected realizer.TargetCandidate, predicate meaning.Node) bool {
	licensedBuyerOmission := selected.Outcome == status.OutcomeIncomplete && selected.Frame.ID == "en.selling.event.omit_buyer.past" &&
		len(selected.MissingOmittedRoles) == 1 && selected.MissingOmittedRoles[0] == "buyer"
	if resolver.registry == nil || resolver.valency == nil || resolver.variants == nil || selected.Outcome != status.OutcomeOK && !licensedBuyerOmission ||
		selected.TransferBinding.CoreRef.String() != string(predicate.Concept) || selected.Sense.ID.String() == "" ||
		selected.Frame.ID == "" || selected.Frame.SenseID != selected.Sense.ID || selected.ConstructionBinding.SenseID != selected.Sense.ID ||
		selected.ConstructionBinding.ValencyFrameID != selected.Frame.ID || selected.ConstructionBinding.FamilyID != selected.Family.Family.ID ||
		selected.ConstructionBinding.VariantID.String() != selected.CompiledVariant.Variant.ID.String() || !selected.ConstructionBinding.Approved ||
		selected.CompiledVariant.FamilyID != selected.Family.Family.ID || selected.CompiledVariant.Variant.FamilyID != selected.Family.Family.ID {
		return false
	}
	familyID := selected.Family.Family.ID
	if selected.Family.Family.Language != "en" || selected.Family.Family.Validate() != nil ||
		selected.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze ||
		selected.CompiledVariant.Variant.Validate(selected.Family.Family) != nil || !variantHasRequiredSlots(selected.Family.Family, selected.CompiledVariant.Variant) ||
		!valencyFrameAllowsFamily(selected.Frame, familyID) {
		return false
	}
	variantID, err := construction.ParseVariantID(selected.CompiledVariant.Variant.ID.String())
	if err != nil {
		return false
	}
	variantIndex := resolver.variants[familyID]
	if variantIndex == nil {
		return false
	}
	compiled, exists := variantIndex.ByID(variantID)
	if !exists || compiled.FamilyID != familyID || !reflect.DeepEqual(compiled.Variant, selected.CompiledVariant.Variant) {
		return false
	}
	familyRecord, exists := resolver.registry.ByID(familyID)
	if !exists || familyRecord.Family.ID != familyID {
		return false
	}
	approved := false
	for _, binding := range resolver.registry.BindingsByFamilyID(familyID) {
		if binding.Approved && binding.SenseID == selected.Sense.ID && binding.ValencyFrameID == selected.Frame.ID &&
			binding.VariantID.String() == selected.CompiledVariant.Variant.ID.String() {
			approved = true
			break
		}
	}
	if !approved {
		return false
	}
	mainSlot, found := englishFamilyMainPredicateSlot(selected.Family.Family)
	if !found || !sequenceHasSlot(selected.CompiledVariant.Variant.Sequence, mainSlot.ID) || !englishPastFrameRolesMatch(selected.Frame, predicate) {
		return false
	}
	for _, condition := range selected.CompiledVariant.Variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == mainSlot.ID && condition.Feature == "tense" && string(condition.Value) != "past" {
			return false
		}
	}
	return true
}

func englishFamilyMainPredicateSlot(family construction.ConstructionFamily) (construction.SlotSpec, bool) {
	var result construction.SlotSpec
	for _, slot := range family.SlotSpecs {
		if slot.Category != construction.CategoryVP || !slot.IsMainPredicate() {
			continue
		}
		if result.ID != "" || slot.Optionality != construction.SlotRequired || slot.Cardinality.Minimum != 1 || slot.Cardinality.Maximum != 1 {
			return construction.SlotSpec{}, false
		}
		result = slot
	}
	return result, result.ID != ""
}

func englishPastFrameRolesMatch(frame valency.ValencyFrame, predicate meaning.Node) bool {
	if predicate.Kind != meaning.KindEvent {
		return false
	}
	participants := make(map[meaning.Role]bool, len(predicate.Participants))
	for _, participant := range predicate.Participants {
		if participant.Role == "" || participants[participant.Role] {
			return false
		}
		participants[participant.Role] = true
	}
	arguments := make(map[meaning.Role]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		if argument.Role == "" {
			return false
		}
		if _, duplicate := arguments[argument.Role]; duplicate {
			return false
		}
		arguments[argument.Role] = argument
		if argument.Required && !participants[argument.Role] {
			return false
		}
	}
	for role := range participants {
		if _, licensed := arguments[role]; !licensed {
			return false
		}
	}
	return true
}

func englishMeaningHasScope(source semantic.MeaningCandidate, target meaning.NodeID) bool {
	if source.Meaning == nil {
		return false
	}
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindScope && node.Target == target {
			return true
		}
	}
	return false
}

func englishPastNegativeScopeMatches(resolver englishNegationResolver, source semantic.MeaningCandidate, selected realizer.TargetCandidate, predicate meaning.Node) bool {
	_, event, route, negated := resolver.negatedPresent(source)
	variant := selected.CompiledVariant.Variant
	return negated && event.ID == predicate.ID && event.TemporalFrame == meaning.TemporalFramePast && route.sense == selected.Sense.ID.String() &&
		route.family == selected.Family.Family.ID.String() && route.frame == string(selected.Frame.ID) &&
		route.pastVariantAllowed(variant.ID.String()) && sequenceHasTerminal(variant.Sequence, "did") && sequenceHasTerminal(variant.Sequence, "not")
}

func englishSimplePastGivingVariant(id string) bool {
	return strings.HasPrefix(id, "construction:en.give.transfer.past.double_np.") || strings.HasPrefix(id, "construction:en.give.transfer.past.to_pp.")
}

func englishSimplePastGivingNegativeVariant(id string) bool {
	if id == englishGivingNegativeLocalToPPNPVariant {
		return true
	}
	switch id {
	case "construction:en.ditransitive_negative.past.the_the_the", "construction:en.ditransitive_negative.past.the_the_a", "construction:en.ditransitive_negative.past.the_a_the", "construction:en.ditransitive_negative.past.the_a_a",
		"construction:en.ditransitive_negative.past.a_the_the", "construction:en.ditransitive_negative.past.a_the_a", "construction:en.ditransitive_negative.past.a_a_the", "construction:en.ditransitive_negative.past.a_a_a":
		return true
	default:
		return false
	}
}

func englishInvariantGivingPastTarget(selected realizer.TargetCandidate, predicate meaning.Node) bool {
	return selected.Sense.ID.String() == "sense:en.give.transfer" && predicate.Kind == meaning.KindEvent && predicate.TemporalFrame == meaning.TemporalFramePast &&
		selected.Family.Family.ID.String() == "construction:en.give.transfer.past" && selected.Frame.ID == "en.give.transfer.past"
}

func englishSourcePredicateAspect(source semantic.MeaningCandidate, predicate meaning.Node) (string, bool) {
	if source.Meaning == nil || predicate.Kind != meaning.KindEvent {
		return "", false
	}
	type evidence struct {
		start, end int
		sense      string
		aspect     string
	}
	seen := map[evidence]struct{}{}
	occurrences := map[[2]int]map[evidence]struct{}{}
	for _, record := range source.Evidence.Records {
		if record.Origin == semantic.EvidenceOriginGenerated || record.Outcome != status.OutcomeOK ||
			record.Role != "" || record.CoreConceptID.String() != string(predicate.Concept) || record.SenseID.String() == "" || record.FamilyID.String() == "" || record.SourceSpan.End <= record.SourceSpan.Start {
			continue
		}
		aspect := ""
		conflict := false
		for _, feature := range record.Features {
			if feature.Feature != "aspect" {
				continue
			}
			if feature.State != features.StateKnown || (feature.Value != "perfective" && feature.Value != "imperfective") {
				conflict = true
				break
			}
			value := string(feature.Value)
			if aspect != "" && aspect != value {
				conflict = true
				break
			}
			aspect = value
		}
		if conflict || aspect == "" {
			return "", false
		}
		item := evidence{start: record.SourceSpan.Start, end: record.SourceSpan.End, sense: record.SenseID.String(), aspect: aspect}
		seen[item] = struct{}{}
		key := [2]int{record.SourceSpan.Start, record.SourceSpan.End}
		if occurrences[key] == nil {
			occurrences[key] = map[evidence]struct{}{}
		}
		occurrences[key][item] = struct{}{}
	}
	if len(occurrences) != 1 || len(seen) == 0 {
		return "", false
	}
	var selected evidence
	for _, values := range occurrences {
		if len(values) != 1 {
			return "", false
		}
		for item := range values {
			selected = item
		}
	}
	return selected.aspect, true
}

func sameMeaningNodePath(left, right []meaning.NodeID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// englishDeterminerCandidatePolicy retains only variants supported by source
// determiner/number evidence and explicit target-lexeme article properties.
// Missing source definiteness keeps all compatible options visible.
// englishApprovedAttributedNegativeSeeingCandidate keeps the dedicated
// object-attribute family licensed by its own approved binding and compiled
// variant. It is separate from the ordinary transitive-negative route.
func englishApprovedAttributedNegativeSeeingCandidate(resolver englishNegationResolver, source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	const familyText = "construction:en.attributed_seeing_object_negative"
	if candidate.Family.Family.ID.String() != familyText || candidate.Kind != meaning.KindEvent ||
		candidate.Sense.ID.String() != "sense:en.see.visual" || candidate.TransferBinding.CoreRef.String() != "concept:seeing" ||
		string(candidate.Frame.ID) != "en.see.visual.attributed_negative" || candidate.Outcome != status.OutcomeOK ||
		candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze {
		return false
	}
	shape, ok := englishAttributedSeeingShape(source, resolver)
	if !ok || shape.scope.ID == "" || shape.event.ID != candidate.NodeID || shape.subject.Multiplicity != meaning.MultiplicityOne ||
		shape.object.Multiplicity != meaning.MultiplicityOne || !englishAttributedSeeingContextsValid(source, resolver) ||
		!englishAttributedSeeingVariant(candidate) || !englishAttributedSeeingFamilyMatches(source, candidate, shape) ||
		!englishAttributedSeeingSubjectVariantAllowed(source, candidate, shape, resolver) || !englishPredicateEvidenceConsistent(source, candidate) || source.Meaning == nil ||
		!realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
		return false
	}
	pastVariant := strings.Contains(candidate.CompiledVariant.Variant.ID.String(), ".svo.past")
	if pastVariant != (shape.event.TemporalFrame == meaning.TemporalFramePast) {
		return false
	}
	familyID := candidate.Family.Family.ID
	if resolver.registry == nil || resolver.valency == nil || resolver.variants == nil {
		return false
	}
	record, exists := resolver.registry.ByID(familyID)
	if !exists || !familyHasOperator(record.Family, construction.OperatorID("negate")) ||
		!familyHasCapability(record.Family, construction.Capability("negation-negated")) || !variantHasRequiredSlots(record.Family, candidate.CompiledVariant.Variant) {
		return false
	}
	variantIndex := resolver.variants[familyID]
	if variantIndex == nil {
		return false
	}
	variant, variantErr := construction.ParseVariantID(candidate.CompiledVariant.Variant.ID.String())
	if variantErr != nil {
		return false
	}
	if _, compiled := variantIndex.ByID(variant); !compiled {
		return false
	}
	frameExists, rolesValid := false, false
	for _, candidateFrame := range resolver.valency.Frames() {
		if string(candidateFrame.ID) == string(candidate.Frame.ID) {
			frameExists = true
			if candidateFrame.SenseID != candidate.Sense.ID || !valencyFrameAllowsFamily(candidateFrame, familyID) {
				break
			}
			subject, object := false, false
			for _, argument := range candidateFrame.Arguments {
				if argument.ID == "subject" {
					subject = argument.Required && argument.Role == "experiencer" && argument.GrammaticalFunction == "subject"
				} else if argument.ID == "object" {
					object = argument.Required && argument.Role == "stimulus" && argument.GrammaticalFunction == "object"
				} else if argument.Required {
					return false
				}
			}
			rolesValid = subject && object
			break
		}
	}
	if !frameExists || !rolesValid {
		return false
	}
	for _, binding := range resolver.registry.BindingsByFamilyID(familyID) {
		if binding.Approved && binding.SenseID == candidate.Sense.ID && string(binding.ValencyFrameID) == string(candidate.Frame.ID) && binding.VariantID.String() == candidate.CompiledVariant.Variant.ID.String() {
			return true
		}
	}
	return false
}

type englishDeterminerCandidatePolicy struct{ resolver englishNegationResolver }

func englishDoorActionObjectNumberChoice(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, role meaning.Role) (string, bool) {
	familyID := candidate.Family.Family.ID.String()
	if source.Meaning == nil || role != "object" || familyID != "construction:en.transitive_past" && familyID != "construction:en.transitive" && familyID != "construction:en.transitive_negative" {
		return "", false
	}
	variantID := candidate.CompiledVariant.Variant.ID.String()
	number := "singular"
	if strings.HasSuffix(variantID, ".opening_object_past.the_plural") || strings.HasSuffix(variantID, ".opening_object_past.indefinite_subject.the_plural") || strings.HasSuffix(variantID, ".closing_object_past.the_plural") || strings.HasSuffix(variantID, ".closing_object_past.indefinite_subject.the_plural") || strings.HasSuffix(variantID, ".bare_object.subject_the") ||
		strings.HasSuffix(variantID, ".bare_object.subject_a") || strings.HasSuffix(variantID, ".bare_object.subject_an") ||
		strings.HasSuffix(variantID, ".opening_object_present.indefinite_subject.the_plural") || strings.HasSuffix(variantID, ".opening_object_present.indefinite_subject.bare_plural") ||
		strings.HasSuffix(variantID, ".opening_object_present.definite_subject.the_plural") || strings.HasSuffix(variantID, ".opening_object_present.definite_subject.bare_plural") ||
		strings.Contains(variantID, ".opening_object.") && strings.Contains(variantID, ".object_the_plural") || strings.Contains(variantID, ".closing_object.") && strings.Contains(variantID, ".object_the_plural") {
		number = "plural"
	}
	var event, object meaning.Node
	eventConcept := "concept:opening_object"
	if candidate.Kind == meaning.KindEvent && candidate.Sense.ID.String() == "sense:en.closing_object.event" {
		eventConcept = "concept:closing_object"
	}
	if candidate.Kind == meaning.KindEvent && (candidate.Sense.ID.String() == "sense:en.opening_object.event" || candidate.Sense.ID.String() == "sense:en.closing_object.event") {
		event, _ = englishMeaningNode(source, candidate.NodeID)
	} else {
		for _, node := range source.Meaning.Nodes {
			if node.Kind != meaning.KindEvent || (string(node.Concept) != "concept:opening_object" && string(node.Concept) != "concept:closing_object") {
				continue
			}
			for _, participant := range node.Participants {
				if participant.Role == "object" && participant.Target == candidate.NodeID {
					event = node
					eventConcept = string(node.Concept)
					break
				}
			}
		}
	}
	expectedFrame := meaning.TemporalFramePresent
	if familyID == "construction:en.transitive_past" || familyID == "construction:en.transitive_negative" && (strings.HasSuffix(variantID, ".past") || strings.Contains(variantID, ".past.")) {
		expectedFrame = meaning.TemporalFramePast
	}
	if event.ID == "" || event.TemporalFrame != expectedFrame || string(event.Concept) != eventConcept {
		return "", false
	}
	for _, participant := range event.Participants {
		if participant.Role == "object" {
			object, _ = englishMeaningNode(source, participant.Target)
			break
		}
	}
	if object.ID == "" || string(object.Concept) != "concept:door" ||
		object.Multiplicity == meaning.MultiplicityOne && number != "singular" || object.Multiplicity == meaning.MultiplicityMany && number != "plural" {
		return "", false
	}
	return number, true
}

func selectedDoorActionObjectReferent(source semantic.MeaningCandidate, eventID meaning.NodeID) meaning.NodeID {
	event, exists := englishMeaningNode(source, eventID)
	if !exists || event.Kind != meaning.KindEvent || (string(event.Concept) != "concept:opening_object" && string(event.Concept) != "concept:closing_object") {
		return ""
	}
	for _, participant := range event.Participants {
		if participant.Role == "object" {
			return participant.Target
		}
	}
	return ""
}

func (policy englishDeterminerCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
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
		intransitiveFamily   = "construction:en.intransitive"
		transitiveFamily     = "construction:en.transitive"
		transitivePastFamily = "construction:en.transitive_past"
		obliqueFamily        = "construction:en.oblique"
		attributedFamily     = "construction:en.attributed_seeing"
	)
	familyID := candidate.Family.Family.ID.String()
	if (familyID == "construction:en.give.transfer" || familyID == "construction:en.give.transfer.past") && candidate.Kind == meaning.KindEvent {
		return englishGivingTargetCandidateAllowed(source, candidate, candidates)
	}
	if (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState) && !englishPredicateEvidenceConsistent(source, candidate) {
		return false
	}
	attributedNegativeSeeing := candidate.Family.Family.ID.String() == "construction:en.attributed_seeing_object_negative"
	if englishAttributedSeeingFamily(candidate) {
		if attributedNegativeSeeing {
			if !englishApprovedAttributedNegativeSeeingCandidate(policy.resolver, source, candidate) {
				return false
			}
		} else if !englishAttributedSeeingMainCandidate(source, candidate, policy.resolver) {
			return false
		}
	}
	if shape, valid := englishAttributedSeeingShape(source, policy.resolver); valid {
		if _, property := shape.properties[candidate.NodeID]; property {
			return englishAttributedSeeingModifierCandidate(source, candidate, candidates, policy.resolver)
		}
		if candidate.NodeID == shape.event.ID && !englishAttributedSeeingFamily(candidate) {
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
				return englishIdentificationVariantArticleAllowed(identification, variant, candidates, identification.family+".this")
			}
			_, allowed := englishIdentificationReferenceChoice(variant, identification.family)
			return allowed && englishIdentificationVariantArticleAllowed(identification, variant, candidates, identification.family+".this", identification.family+".that", identification.family+".it")
		}
		if candidate.Family.Family.ID.String() == "construction:en.identification" || candidate.Family.Family.ID.String() == "construction:en.attribution" {
			return false
		}
	}
	if identification, recognized := englishTargetIdentificationState(source); recognized {
		if candidate.Kind == meaning.KindState && candidate.NodeID == identification.stateID {
			if !identification.valid || !englishIdentificationAttributesAllowed(source, candidate.NodeID, identification.role, identification.referent, candidate.TransferBinding.CoreRef) ||
				!englishIdentificationMemberCountSupported(source, identification, candidates) {
				return false
			}
			variant := candidate.CompiledVariant.Variant.ID.String()
			if identification.family == "construction:en.attribution_negative" {
				if familyID != identification.family || candidate.Sense.ID.String() != identification.senseID || identification.mode != "none" {
					return false
				}
				number := englishMultiplicityNumber(identification.multiplicity)
				if number == "" {
					number = "singular"
				}
				tense := "present"
				if identification.temporalFrame == meaning.TemporalFramePast {
					tense = "past"
				} else if identification.temporalFrame != "" && identification.temporalFrame != meaning.TemporalFramePresent {
					return false
				}
				prefix := identification.family + "." + number + "." + tense + "."
				if !strings.HasPrefix(variant, prefix) {
					return false
				}
				article := strings.TrimPrefix(variant, prefix)
				switch article {
				case "the":
					return true
				case "a", "an":
					if number != "singular" {
						return false
					}
					expectedArticle, known := englishIdentificationIndefiniteArticle(identification, candidates)
					return known && article == expectedArticle
				case "bare":
					return number == "plural"
				default:
					return false
				}
			}
			pastVariant := strings.HasSuffix(variant, ".past")
			if (identification.temporalFrame == meaning.TemporalFramePast) != pastVariant {
				return false
			}
			variant = strings.TrimSuffix(variant, ".past")
			if identification.multiplicity == meaning.MultiplicityMany && identification.role == "bearer" {
				if identification.mode == "none" && identification.family == "construction:en.attribution" &&
					(variant == identification.family+".the_plural" || variant == identification.family+".bare_plural") {
					return true
				}
				switch identification.mode {
				case "proximal":
					return variant == identification.family+".these"
				case "distal":
					return variant == identification.family+".those"
				default:
					return false
				}
			}
			if identification.role == "member" && identification.multiplicity == meaning.MultiplicityMany {
				switch identification.mode {
				case "proximal":
					return variant == identification.family+".these"
				case "distal":
					return variant == identification.family+".those"
				case "unresolved", "demonstrative":
					return variant == identification.family+".these" || variant == identification.family+".those"
				}
			}
			if identification.mode == "none" && identification.role == "bearer" {
				if familyID != identification.family || candidate.Sense.ID.String() != identification.senseID {
					return false
				}
				if variant == identification.family+".the" {
					return identification.multiplicity != meaning.MultiplicityMany
				}
				if variant == identification.family+".the_plural" {
					return identification.concept == "concept:door" && identification.multiplicity == ""
				}
				if variant != identification.family+".a" && variant != identification.family+".an" {
					return false
				}
				article, known := englishIdentificationIndefiniteArticle(identification, candidates)
				return known && variant == identification.family+"."+article
			}
			_, allowed := englishIdentificationReferenceChoice(variant, identification.family)
			if identification.mode == "proximal" {
				if identification.role == "member" {
					allowed = englishIdentificationVariantArticleAllowed(identification, variant, candidates, identification.family+".this")
				} else {
					allowed = variant == identification.family+".this"
				}
			}
			if identification.mode == "distal" {
				if identification.role == "member" {
					allowed = englishIdentificationVariantArticleAllowed(identification, variant, candidates, identification.family+".that")
				} else {
					allowed = variant == identification.family+".that"
				}
			}
			if identification.mode == "demonstrative" {
				allowed = identification.family == "construction:en.attribution" && (variant == identification.family+".this" || variant == identification.family+".that")
			}
			if identification.mode == "unresolved" {
				allowed = identification.family == "construction:en.identification" && englishIdentificationVariantArticleAllowed(identification, variant, candidates, identification.family+".this", identification.family+".that", identification.family+".it")
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
	if familyID == "construction:en.attributed_seeing_object" && candidate.Kind == meaning.KindEvent {
		if _, valid := englishAttributedSeeingShape(source, policy.resolver); valid {
			if deixis := englishAttributedSeeingDeixis(source, policy.resolver); deixis != "" {
				determiners, known := englishDeterminersForCandidate(candidate)
				if !known {
					return false
				}
				objectDeterminer := determiners.object
				switch deixis {
				case "distal", "proximal", "demonstrative", "ambiguous":
					if !englishAttributedSeeingDeixisAllows(source, objectDeterminer, policy.resolver) {
						return false
					}
				default:
					return false
				}
				singular := englishVariantRequiresNPNumber(candidate, "object", "singular")
				plural := englishVariantRequiresNPNumber(candidate, "object", "plural")
				if singular == plural || singular && objectDeterminer != "this" && objectDeterminer != "that" || plural && objectDeterminer != "these" && objectDeterminer != "those" {
					return false
				}
			}
		}
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
		if values, found := runtime.SourceFeatures(source, role, coreConcept(ref)); found && values["determiner"] != "" && string(values["determiner"]) != englishDeterminerFeature(determinerVariant.object) && englishAttributedSeeingDeixis(source, policy.resolver) == "" {
			return false
		}
		return true
	}

	if !englishPastTargetFamilyMatches(source, candidate, familyID, transitiveFamily, transitivePastFamily) {
		return false
	}
	negativeRoute, isNegativeSense := policy.resolver.routeForSense(candidate.Sense.ID)
	if _, event, route, negated := policy.resolver.negatedPresent(source); negated && event.ID == candidate.NodeID && route.family == "construction:en.intransitive" &&
		candidate.Family.Family.ID.String() == route.family && !strings.Contains(candidate.CompiledVariant.Variant.ID.String(), ".negative.") {
		return false
	}
	isNegativeFamily := attributedNegativeSeeing || candidate.Family.Family.ID.String() == "construction:en.transitive_negative" || candidate.Family.Family.ID.String() == "construction:en.oblique_negative" ||
		isNegativeSense && negativeRoute.concept == "concept:giving" && candidate.Family.Family.ID.String() == negativeRoute.family ||
		isNegativeSense && negativeRoute.family == "construction:en.intransitive" && candidate.Family.Family.ID.String() == negativeRoute.family && strings.Contains(candidate.CompiledVariant.Variant.ID.String(), ".negative.")
	if familyID == "construction:en.transitive_negative" && candidate.Kind == meaning.KindEvent {
		_, event, route, negated := policy.resolver.negatedPresent(source)
		if negated && event.ID == candidate.NodeID && route.family == "construction:en.transitive_negative" {
			if englishVPAuxiliaryVariant(candidate) {
				if _, _, ok := englishTransitiveNegativeVPFeatures(policy.resolver, source, candidate, candidates); !ok {
					return false
				}
			} else if englishLegacyNegativeVPVariant(candidate) {
				for _, auxiliaryCandidate := range candidates {
					if auxiliaryCandidate.NodeID == candidate.NodeID && auxiliaryCandidate.Sense.ID == candidate.Sense.ID && englishVPAuxiliaryVariant(auxiliaryCandidate) &&
						englishTransitiveNegativeLocalNPCandidateAllowed(source, auxiliaryCandidate, candidates) {
						return false
					}
				}
			}
		}
	}
	if isNegativeFamily {
		_, event, sourceRoute, negated := policy.resolver.negatedPresent(source)
		if attributedNegativeSeeing {
			if !negated || event.ID != candidate.NodeID || sourceRoute.concept != "concept:seeing" || sourceRoute.sense != candidate.Sense.ID.String() {
				return false
			}
		} else if (!isNegativeSense && !englishVPAuxiliaryVariant(candidate)) || !negated || event.ID != candidate.NodeID || sourceRoute.family != candidate.Family.Family.ID.String() || sourceRoute.sense != candidate.Sense.ID.String() {
			return false
		}
		number, known := "singular", attributedNegativeSeeing
		if englishVPAuxiliaryVariant(candidate) {
			number, _, known = englishTransitiveNegativeVPFeatures(policy.resolver, source, candidate, candidates)
		} else if !attributedNegativeSeeing {
			number, known = policy.resolver.negationNumber(source, candidate, sourceRoute, candidates)
		}
		variantID := candidate.CompiledVariant.Variant.ID.String()
		variantAllowed := attributedNegativeSeeing && englishApprovedAttributedNegativeSeeingCandidate(policy.resolver, source, candidate)
		if !attributedNegativeSeeing && !englishVPAuxiliaryVariant(candidate) {
			variantAllowed = sourceRoute.variantMatchesNumber(variantID, number)
		} else if englishVPAuxiliaryVariant(candidate) {
			_, _, variantAllowed = englishTransitiveNegativeVPFeatures(policy.resolver, source, candidate, candidates)
		}
		if event.TemporalFrame == meaning.TemporalFramePast && !englishVPAuxiliaryVariant(candidate) && !attributedNegativeSeeing {
			variantAllowed = sourceRoute.pastVariantAllowed(variantID)
		}
		if !known || !variantAllowed {
			return false
		}
	}
	if englishVPAuxiliaryVariant(candidate) {
		return englishTransitiveNegativeLocalNPCandidateAllowed(source, candidate, candidates)
	}
	if candidate.Kind != meaning.KindEvent && candidate.Kind != meaning.KindState || familyID != intransitiveFamily && familyID != transitiveFamily && familyID != transitivePastFamily && familyID != obliqueFamily && familyID != "construction:en.to_oblique" && familyID != "construction:en.with_oblique" && familyID != "construction:en.transitive_with_oblique" && familyID != attributedFamily && familyID != "construction:en.attribution_negative" && familyID != "construction:en.attributed_seeing_object" && familyID != "construction:en.attributed_seeing_both" && familyID != "construction:en.transitive_personal" && !isNegativeFamily {
		return true
	}
	variant, knownVariant := englishDeterminersForCandidate(candidate)
	if !knownVariant {
		return true
	}
	if familyID == "construction:en.attribution_negative" && candidate.Kind == meaning.KindState {
		if source.Meaning == nil {
			return false
		}
		state := findMeaningNode(*source.Meaning, candidate.NodeID)
		if state.Kind != meaning.KindState || len(state.Participants) != 1 || state.Participants[0].Role != "bearer" {
			return false
		}
		referent := findMeaningNode(*source.Meaning, state.Participants[0].Target)
		if referent.Kind != meaning.KindReferent {
			return false
		}
		lexical := make([]realizer.TargetCandidate, 0, 2)
		for _, available := range candidates {
			if available.SourceCandidateID == candidate.SourceCandidateID && available.NodeID == referent.ID && available.Kind == meaning.KindReferent &&
				(available.Outcome == status.OutcomeOK || available.Outcome == status.OutcomeIncomplete) && available.Lexeme.ID.String() != "" {
				lexical = append(lexical, available)
			}
		}
		if len(lexical) == 0 {
			return variant.subject == "the" || variant.subject == "bare"
		}
		article, articleKnown := commonEnglishIndefiniteArticle(lexical)
		number := ""
		for _, item := range lexical {
			itemNumber, _, ok := englishTargetNumber(source, referent.ID, item.Lexeme.Properties, item.Lexeme.ParadigmID)
			if !ok || number != "" && number != itemNumber {
				return false
			}
			number = itemNumber
		}
		if variant.subject == "bare" {
			return number == "plural"
		}
		if variant.subject == "a" || variant.subject == "an" {
			if number != "singular" || !articleKnown || article != variant.subject {
				return false
			}
		}
		if values, found := runtime.SourceFeatures(source, "bearer", coreConcept(referent)); found && values["determiner"] != "" && string(values["determiner"]) != englishDeterminerFeature(variant.subject) {
			return false
		}
		return true
	}
	if source.Meaning == nil {
		return candidate.CompiledVariant.Variant.ID.String() == englishStandardDeterminerVariant(familyID)
	}
	predicate := findMeaningNode(*source.Meaning, candidate.NodeID)
	if predicate.ID == "" {
		return candidate.CompiledVariant.Variant.ID.String() == englishStandardDeterminerVariant(familyID)
	}
	switch string(predicate.Concept) {
	case "concept:waking_up":
		if !strings.HasPrefix(candidate.CompiledVariant.Variant.ID.String(), "construction:en.intransitive.sv.wake_up") {
			return false
		}
	case "concept:falling_asleep":
		if !strings.HasPrefix(candidate.CompiledVariant.Variant.ID.String(), "construction:en.intransitive.sv.fall_asleep") {
			return false
		}
	}
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if familyID == "construction:en.transitive_personal" && slot.ID == "subject" {
			continue
		}
		if (slot.ID != "subject" && slot.ID != "bearer" && slot.ID != "object" && slot.ID != "target" && slot.ID != "recipient" && slot.ID != "theme") || slot.ArgumentBinding == nil {
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
		if selectedNumber, selected := englishAttributedSeeingDoorNumberChoice(source, candidate, slot.ID, string(role), policy.resolver); selected && referent.ID == englishAttributedSeeingObjectID(source, policy.resolver) {
			number, numberSet, numberConflict = selectedNumber, true, false
		}
		if selectedNumber, selected := englishDoorActionObjectNumberChoice(source, candidate, role); selected {
			number, numberSet, numberConflict = selectedNumber, true, false
		}
		indefiniteArticle, hasCommonIndefiniteArticle := commonEnglishIndefiniteArticle(lexicalCandidates)
		if (familyID == attributedFamily || familyID == "construction:en.attributed_seeing_both") && slot.ID == "subject" ||
			(familyID == "construction:en.attributed_seeing_object" || familyID == "construction:en.attributed_seeing_both") && slot.ID == "object" {
			// The first pronounced word of this licensed NP is the modifier
			// adjective, so select the article from the adjective's onset.
			if article, licensed := englishAttributedSeeingIndefiniteArticle(source, candidate, slot.ID, candidates, policy.resolver); licensed {
				indefiniteArticle, hasCommonIndefiniteArticle = article, true
			} else {
				return false
			}
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
		if familyID == "construction:en.attributed_seeing_object" && slot.ID == "object" && englishAttributedSeeingDeixis(source, policy.resolver) != "" &&
			(actualArticle == "this" || actualArticle == "that" || actualArticle == "these" || actualArticle == "those") {
			if !englishAttributedSeeingDeixisAllows(source, actualArticle, policy.resolver) ||
				(number == "singular" && actualArticle != "this" && actualArticle != "that") ||
				(number == "plural" && actualArticle != "these" && actualArticle != "those") ||
				(actualArticle != "this" && actualArticle != "that" && actualArticle != "these" && actualArticle != "those") {
				return false
			}
			continue
		}
		if (familyID == attributedFamily || familyID == "construction:en.attributed_seeing_both") && slot.ID == "subject" &&
			len(englishAttributedSeeingSubjectAnchors(source, referent.ID)) != 0 &&
			(actualArticle == "this" || actualArticle == "that" || actualArticle == "these" || actualArticle == "those") {
			if number == "singular" && actualArticle != "this" && actualArticle != "that" ||
				number == "plural" && actualArticle != "these" && actualArticle != "those" {
				return false
			}
			continue
		}
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
			if actualDeterminer == "bare" {
				continue
			}
			cuttingMassDefinite := candidate.Sense.ID.String() == "sense:en.cutting_object.event" &&
				(candidate.Frame.ID == "en.cutting_object.transitive" && familyID == "construction:en.transitive" ||
					candidate.Frame.ID == "en.cutting_object.transitive.past" && familyID == "construction:en.transitive_past" ||
					candidate.Frame.ID == "en.cutting_object.transitive.negative" && familyID == "construction:en.transitive_negative")
			tearingMassDefinite := candidate.Sense.ID.String() == "sense:en.tearing_object.event" &&
				(candidate.Frame.ID == "en.tearing_object.transitive" && familyID == "construction:en.transitive" ||
					candidate.Frame.ID == "en.tearing_object.transitive.past" && familyID == "construction:en.transitive_past" ||
					candidate.Frame.ID == "en.tearing_object.transitive.negative" && familyID == "construction:en.transitive_negative")
			if !cuttingMassDefinite && !tearingMassDefinite || actualDeterminer != "definite" {
				return false
			}
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

func englishAttributedSeeingIndefiniteArticle(source semantic.MeaningCandidate, scope realizer.TargetCandidate, slotID string, candidates []realizer.TargetCandidate, resolver englishNegationResolver) (string, bool) {
	shape, valid := englishAttributedSeeingShape(source, resolver)
	if !valid {
		return "", false
	}
	var bearer meaning.NodeID
	switch slotID {
	case "subject", "bearer":
		bearer = shape.subject.ID
	case "object":
		bearer = shape.object.ID
	default:
		return "", false
	}
	property, exists := shape.propertyByBearer[bearer]
	if !exists {
		return "", false
	}
	article := ""
	found := false
	for _, candidate := range candidates {
		if candidate.SourceCandidateID != scope.SourceCandidateID || candidate.SemanticID != scope.SemanticID ||
			candidate.NodeID != property.ID || candidate.Kind != meaning.KindState || candidate.Outcome != status.OutcomeOK && candidate.Outcome != status.OutcomeIncomplete || candidate.TransferBinding.CoreRef.String() != string(property.Concept) {
			continue
		}
		value, licensed := englishIndefiniteArticle(candidate.Lexeme.Properties)
		if !licensed {
			return "", false
		}
		if found && article != value {
			return "", false
		}
		article, found = value, true
	}
	return article, found
}

func englishIdentificationVariantArticleAllowed(identification englishIdentification, variant string, candidates []realizer.TargetCandidate, stems ...string) bool {
	article := ""
	mass := false
	for _, target := range candidates {
		if target.NodeID != identification.referent || target.Kind != meaning.KindReferent || target.Outcome != status.OutcomeOK || target.TransferBinding.CoreRef.String() != string(identification.concept) {
			continue
		}
		mass, pluralOnly := englishNumberClass(target.Lexeme.Properties, target.Lexeme.ParadigmID)
		if mass && !pluralOnly {
			for _, stem := range stems {
				if variant == stem+".mass" {
					return true
				}
			}
			continue
		}
		if value, known := englishIndefiniteArticle(target.Lexeme.Properties); known {
			if article != "" && article != value {
				return false
			}
			article = value
		}
		isMass, _ := englishNumberClass(target.Lexeme.Properties, target.Lexeme.ParadigmID)
		mass = mass || isMass
	}
	if article == "" && !mass {
		return false
	}
	for _, stem := range stems {
		if mass && variant == stem+".mass" || variant == stem && article == "a" || variant == stem+".an" && article == "an" {
			return true
		}
	}
	return false
}

// englishIdentificationIndefiniteArticle returns an article only when every
// compatible target lexeme for the exact bearer agrees on its explicitly
// declared onset article. Unknown or conflicting candidates remain unresolved.
func englishIdentificationIndefiniteArticle(identification englishIdentification, candidates []realizer.TargetCandidate) (string, bool) {
	article := ""
	count := 0
	for _, target := range candidates {
		if target.NodeID != identification.referent || target.Kind != meaning.KindReferent || target.Outcome != status.OutcomeOK || target.TransferBinding.CoreRef.String() != string(identification.concept) {
			continue
		}
		count++
		value, known := englishIndefiniteArticle(target.Lexeme.Properties)
		if !known || value != "a" && value != "an" || article != "" && article != value {
			return "", false
		}
		article = value
	}
	return article, count > 0 && article != ""
}

func englishGivingTargetCandidateAllowed(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	if candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze || !englishPredicateEvidenceConsistent(source, candidate) {
		return false
	}
	if englishGivingLocalNPVariantID(candidate.CompiledVariant.Variant.ID.String()) && !englishGivingLocalNPCandidateAllowed(source, candidate) {
		return false
	}
	var roleSlots map[meaning.Role]string
	switch {
	case candidate.Sense.ID.String() == "sense:en.give.transfer":
		predicate, ok := englishMeaningNode(source, candidate.NodeID)
		if !ok || predicate.Kind != meaning.KindEvent {
			return false
		}
		if predicate.TemporalFrame == meaning.TemporalFramePast {
			if candidate.Family.Family.ID.String() == "construction:en.ditransitive_negative" && candidate.Frame.ID == "en.give.transfer.negative" {
				_, negatedEvent, route, negated := (englishNegationResolver{}).negatedPresent(source)
				if !negated || negatedEvent.ID != predicate.ID || route.concept != "concept:giving" || !englishSimplePastGivingNegativeVariant(candidate.CompiledVariant.Variant.ID.String()) {
					return false
				}
			} else if candidate.Family.Family.ID.String() != "construction:en.give.transfer.past" || candidate.Frame.ID != "en.give.transfer.past" {
				return false
			}
		} else if candidate.Family.Family.ID.String() != "construction:en.give.transfer" || candidate.Frame.ID != "en.give.transfer" {
			return false
		}
		roleSlots = map[meaning.Role]string{"giver": "subject", "recipient": "recipient", "theme": "theme"}
	case candidate.Sense.ID.String() == "sense:en.buying.event" && candidate.Frame.ID == "en.buying.event.commerce" && strings.HasPrefix(candidate.CompiledVariant.Variant.ID.String(), "construction:en.give.transfer.from_pp."):
		roleSlots = map[meaning.Role]string{"buyer": "subject", "seller": "recipient", "goods": "theme"}
	case candidate.Sense.ID.String() == "sense:en.selling.event" && candidate.Frame.ID == "en.selling.event.commerce" && strings.HasPrefix(candidate.CompiledVariant.Variant.ID.String(), "construction:en.give.transfer.to_pp."):
		roleSlots = map[meaning.Role]string{"seller": "subject", "buyer": "recipient", "goods": "theme"}
	case candidate.Sense.ID.String() == "sense:en.writing_message.event" && candidate.Frame.ID == "en.writing_message.full" && strings.HasPrefix(candidate.CompiledVariant.Variant.ID.String(), "construction:en.give.transfer.to_pp."):
		predicate, ok := englishMeaningNode(source, candidate.NodeID)
		if !ok || predicate.Kind != meaning.KindEvent || predicate.Concept != "concept:writing_message" || predicate.TemporalFrame == meaning.TemporalFramePast ||
			candidate.Family.Family.ID.String() != "construction:en.give.transfer" {
			return false
		}
		roleSlots = map[meaning.Role]string{"writer": "subject", "message": "theme", "recipient": "recipient"}
	case candidate.Sense.ID.String() == "sense:en.writing_message.event" && candidate.Frame.ID == "en.writing_message.full" &&
		candidate.Family.Family.ID.String() == "construction:en.give.transfer.past" && candidate.CompiledVariant.Variant.ID.String() == englishGivingLocalToPPNPVariant:
		predicate, ok := englishMeaningNode(source, candidate.NodeID)
		if !ok || predicate.Kind != meaning.KindEvent || predicate.Concept != "concept:writing_message" || predicate.TemporalFrame != meaning.TemporalFramePast {
			return false
		}
		var valid bool
		roleSlots, valid = englishWritingMessagePastRoleSlots(candidate)
		if !valid {
			return false
		}
	default:
		return false
	}
	if englishGivingLocalNPVariantID(candidate.CompiledVariant.Variant.ID.String()) {
		if candidate.Sense.ID.String() == "sense:en.writing_message.event" {
			var valid bool
			roleSlots, valid = englishWritingMessagePastRoleSlots(candidate)
			if !valid {
				return false
			}
		} else {
			var valid bool
			roleSlots, valid = englishRequiredFrameRoleSlots(candidate)
			if !valid {
				return false
			}
		}
	}
	predicate, ok := englishMeaningNode(source, candidate.NodeID)
	if !ok || predicate.Kind != meaning.KindEvent {
		return false
	}
	for role, slotID := range roleSlots {
		var referent meaning.Node
		for _, participant := range predicate.Participants {
			if participant.Role == role {
				referent = findMeaningNode(*source.Meaning, participant.Target)
				break
			}
		}
		if referent.ID == "" {
			return false
		}
		localNP := englishGivingLocalNPVariantID(candidate.CompiledVariant.Variant.ID.String()) && (slotID == "subject" || slotID == "recipient" || slotID == "theme")
		article, foundArticle := englishGivingVariantArticle(candidate, slotID)
		if localNP {
			article, foundArticle = "", true
		}
		if !foundArticle {
			return false
		}
		lexical := make([]realizer.TargetCandidate, 0, 2)
		for _, available := range candidates {
			if available.NodeID == referent.ID && available.Kind == meaning.KindReferent && available.Lexeme.ID.String() != "" &&
				available.TransferBinding.CoreRef.String() == string(referent.Concept) && !englishChildKinshipTargetCandidate(available) {
				lexical = append(lexical, available)
			}
		}
		if len(lexical) == 0 {
			return false
		}
		number, numberSet := "", false
		mass, pluralOnly, articleForm := true, true, ""
		for _, item := range lexical {
			itemNumber, _, known := englishTargetNumber(source, referent.ID, item.Lexeme.Properties, item.Lexeme.ParadigmID)
			if !known || numberSet && itemNumber != number {
				return false
			}
			number, numberSet = itemNumber, true
			itemMass, itemPluralOnly := englishNumberClass(item.Lexeme.Properties, item.Lexeme.ParadigmID)
			mass, pluralOnly = mass && itemMass, pluralOnly && itemPluralOnly
			itemArticle, articleKnown := englishIndefiniteArticle(item.Lexeme.Properties)
			if localNP && (!articleKnown || itemMass || itemPluralOnly || itemNumber != "singular") {
				return false
			}
			if articleKnown {
				if articleForm != "" && articleForm != itemArticle {
					return false
				}
				articleForm = itemArticle
			} else if article == "a" || article == "an" {
				return false
			}
		}
		if !numberSet || !englishVariantRequiresNPNumber(candidate, slotID, number) {
			return false
		}
		if localNP {
			if number != "singular" || mass || pluralOnly || articleForm == "" {
				return false
			}
			if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found {
				sourceDeterminer := string(values["determiner"])
				if sourceDeterminer != "" && sourceDeterminer != "definite" && sourceDeterminer != "indefinite" &&
					!(candidate.Sense.ID.String() == "sense:en.writing_message.event" && sourceDeterminer == "bare") {
					return false
				}
				if sourceDeterminer == "indefinite" && (articleForm != "a" && articleForm != "an") {
					return false
				}
			}
			continue
		}
		actual := englishDeterminerFeature(article)
		if actual == "indefinite" && (mass || pluralOnly || number != "singular" || articleForm != article) {
			return false
		}
		if actual == "bare" && !mass && !pluralOnly && number != "plural" {
			return false
		}
		if values, found := runtime.SourceFeatures(source, role, coreConcept(referent)); found && values["determiner"] != "" && string(values["determiner"]) != actual {
			return false
		}
	}
	return true
}

func englishWritingMessagePastRoleSlots(candidate realizer.TargetCandidate) (map[meaning.Role]string, bool) {
	if candidate.Sense.ID.String() != "sense:en.writing_message.event" || candidate.Frame.ID != "en.writing_message.full" ||
		candidate.Frame.SenseID != candidate.Sense.ID || candidate.Family.Family.ID.String() != "construction:en.give.transfer.past" ||
		candidate.Family.Family.Language != "en" || candidate.TransferBinding.SenseID != candidate.Sense.ID ||
		candidate.TransferBinding.CoreRef.String() != "concept:writing_message" || !candidate.TransferBinding.TransferAuthorized() ||
		candidate.Outcome != status.OutcomeOK || len(candidate.MissingOmittedRoles) != 0 || len(candidate.Frame.SurfaceOmittedRoles) != 0 ||
		candidate.CompiledVariant.FamilyID != candidate.Family.Family.ID || candidate.CompiledVariant.Variant.FamilyID != candidate.Family.Family.ID ||
		candidate.CompiledVariant.Variant.ID.String() != englishGivingLocalToPPNPVariant ||
		candidate.CompiledVariant.Variant.DirectionPolicy != construction.DirectionRealize ||
		candidate.ConstructionBinding.SenseID != candidate.Sense.ID || candidate.ConstructionBinding.ValencyFrameID != candidate.Frame.ID ||
		candidate.ConstructionBinding.FamilyID != candidate.Family.Family.ID || candidate.ConstructionBinding.VariantID.String() != candidate.CompiledVariant.Variant.ID.String() ||
		!candidate.ConstructionBinding.Approved || candidate.Frame.Validate() != nil || candidate.Family.Family.Validate() != nil ||
		candidate.CompiledVariant.Variant.Validate(candidate.Family.Family) != nil || !valencyFrameAllowsFamily(candidate.Frame, candidate.Family.Family.ID) {
		return nil, false
	}
	wantRoles := map[meaning.Role]bool{"writer": true, "message": true, "recipient": true}
	if len(candidate.Frame.SemanticRoles) != len(wantRoles) || len(candidate.Frame.Arguments) != len(wantRoles) {
		return nil, false
	}
	seenRoles := make(map[meaning.Role]bool, len(candidate.Frame.SemanticRoles))
	for _, role := range candidate.Frame.SemanticRoles {
		if !wantRoles[role] || seenRoles[role] {
			return nil, false
		}
		seenRoles[role] = true
	}
	wantArguments := map[string]struct {
		role       meaning.Role
		function   string
		adposition string
	}{
		"subject":   {role: "writer", function: "subject"},
		"theme":     {role: "message", function: "object"},
		"recipient": {role: "recipient", function: "oblique", adposition: "to"},
	}
	roleSlots := make(map[meaning.Role]string, len(wantArguments))
	for _, argument := range candidate.Frame.Arguments {
		want, exists := wantArguments[argument.ID]
		if !exists || roleSlots[argument.Role] != "" || argument.Role != want.role || !argument.Required ||
			argument.GrammaticalFunction != want.function || argument.Adposition != want.adposition {
			return nil, false
		}
		roleSlots[argument.Role] = argument.ID
	}
	if len(roleSlots) != len(wantArguments) {
		return nil, false
	}
	return roleSlots, true
}

func englishGivingLocalNPCandidateAllowed(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	negativeRoute := candidate.Family.Family.ID.String() == "construction:en.ditransitive_negative" && candidate.Frame.ID == "en.give.transfer.negative"
	positiveRoute := candidate.Family.Family.ID.String() == "construction:en.give.transfer.past" && candidate.Frame.ID == "en.give.transfer.past"
	writingMessagePastRoute := candidate.Sense.ID.String() == "sense:en.writing_message.event" && candidate.Frame.ID == "en.writing_message.full" &&
		candidate.Family.Family.ID.String() == "construction:en.give.transfer.past" && candidate.CompiledVariant.Variant.ID.String() == englishGivingLocalToPPNPVariant
	if source.Meaning == nil || candidate.Kind != meaning.KindEvent || (!positiveRoute && !negativeRoute && !writingMessagePastRoute) ||
		candidate.Sense.ID.String() != "sense:en.give.transfer" && !writingMessagePastRoute ||
		candidate.TransferBinding.CoreRef.String() != "concept:giving" && !writingMessagePastRoute || candidate.Outcome != status.OutcomeOK ||
		candidate.CompiledVariant.Variant.DirectionPolicy != construction.DirectionRealize ||
		!englishGivingLocalNPVariantID(candidate.CompiledVariant.Variant.ID.String()) {
		return false
	}
	event, exists := englishMeaningNode(source, candidate.NodeID)
	wantConcept := meaning.ConceptRef("concept:giving")
	if writingMessagePastRoute {
		wantConcept = "concept:writing_message"
	}
	if !exists || event.Kind != meaning.KindEvent || event.Concept != wantConcept || event.TemporalFrame != meaning.TemporalFramePast {
		return false
	}
	if negativeRoute {
		_, negatedEvent, route, negated := (englishNegationResolver{}).negatedPresent(source)
		if !negated || negatedEvent.ID != event.ID || route.concept != "concept:giving" || !englishSimplePastGivingNegativeVariant(candidate.CompiledVariant.Variant.ID.String()) {
			return false
		}
	}
	roleSlots, ok := englishRequiredFrameRoleSlots(candidate)
	if writingMessagePastRoute {
		roleSlots, ok = englishWritingMessagePastRoleSlots(candidate)
	}
	if !ok {
		return false
	}
	frameArguments := make(map[meaning.Role]valency.ArgumentSlot, len(candidate.Frame.Arguments))
	for _, argument := range candidate.Frame.Arguments {
		if _, duplicate := frameArguments[argument.Role]; duplicate {
			return false
		}
		frameArguments[argument.Role] = argument
	}
	seen := make(map[meaning.Role]bool, len(event.Participants))
	for _, participant := range event.Participants {
		if _, licensed := roleSlots[participant.Role]; !licensed || seen[participant.Role] {
			return false
		}
		seen[participant.Role] = true
		referent, found := englishMeaningNode(source, participant.Target)
		if !found || referent.Kind != meaning.KindReferent || referent.Concept == "" || referent.Multiplicity != meaning.MultiplicityOne {
			return false
		}
	}
	for role, argument := range frameArguments {
		if argument.Required && !seen[role] {
			return false
		}
	}
	localRoles := make(map[meaning.Role]bool, 2)
	for role, slotID := range roleSlots {
		if slotID == "subject" || slotID == "recipient" || slotID == "theme" {
			localRoles[role] = true
		}
	}
	if len(localRoles) != 3 {
		return false
	}
	for _, participant := range event.Participants {
		if !localRoles[participant.Role] {
			continue
		}
		referent, _ := englishMeaningNode(source, participant.Target)
		if values, found := runtime.SourceFeatures(source, participant.Role, coreConcept(referent)); found {
			determiner := string(values["determiner"])
			if determiner != "" && determiner != "definite" && determiner != "indefinite" {
				return false
			}
		}
	}
	return true
}

func englishRequiredFrameRoleSlots(candidate realizer.TargetCandidate) (map[meaning.Role]string, bool) {
	if candidate.Frame.SenseID != candidate.Sense.ID || len(candidate.Frame.Arguments) == 0 {
		return nil, false
	}
	slots := make(map[meaning.Role]string, len(candidate.Frame.Arguments))
	for _, argument := range candidate.Frame.Arguments {
		if _, duplicate := slots[argument.Role]; duplicate {
			return nil, false
		}
		var found *construction.SlotSpec
		for index := range candidate.Family.Family.SlotSpecs {
			slot := &candidate.Family.Family.SlotSpecs[index]
			if slot.ID == argument.ID {
				if found != nil {
					return nil, false
				}
				found = slot
			}
		}
		if found == nil || found.Category != construction.CategoryNP || found.ArgumentBinding == nil || found.ArgumentBinding.SemanticRole != argument.Role || found.Optionality != construction.SlotRequired || found.Cardinality.Minimum != 1 {
			return nil, false
		}
		slots[argument.Role] = found.ID
	}
	return slots, len(slots) == len(candidate.Frame.Arguments)
}

func englishGivingVariantArticle(candidate realizer.TargetCandidate, slotID string) (string, bool) {
	for index, token := range candidate.CompiledVariant.Variant.Sequence {
		if token.Kind != construction.TokenSlot || token.SlotID != slotID {
			continue
		}
		for prior := index - 1; prior >= 0; prior-- {
			previous := candidate.CompiledVariant.Variant.Sequence[prior]
			if previous.Kind == construction.TokenSlot {
				break
			}
			if previous.Kind == construction.TokenTerminal && isEnglishDeterminerToken(previous.Terminal) {
				return previous.Terminal, true
			}
		}
		return "", false
	}
	return "", false
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
	if strings.HasPrefix(id, "construction:en.give.transfer.") {
		rest := strings.TrimPrefix(id, "construction:en.give.transfer.")
		if strings.HasPrefix(rest, "past.") {
			rest = strings.TrimPrefix(rest, "past.")
		}
		parts := strings.Split(rest, ".")
		if len(parts) == 2 && (parts[0] == "double_np" || parts[0] == "to_pp") {
			articles := strings.Split(parts[1], "_")
			isArticle := func(article string) bool { return article == "the" || article == "a" || article == "an" }
			if len(articles) == 3 && isArticle(articles[0]) && isArticle(articles[1]) && isArticle(articles[2]) {
				return englishDeterminers{subject: articles[0], target: articles[1], object: articles[2]}, true
			}
		}
	}
	if strings.HasPrefix(id, "construction:en.attribution_negative.") {
		parts := strings.Split(strings.TrimPrefix(id, "construction:en.attribution_negative."), ".")
		if len(parts) == 3 && (parts[0] == "singular" || parts[0] == "plural") && (parts[1] == "present" || parts[1] == "past") &&
			(parts[2] == "the" || parts[2] == "a" || parts[2] == "an" || parts[2] == "bare") {
			return englishDeterminers{subject: parts[2]}, true
		}
	}
	if strings.HasPrefix(id, "construction:en.ditransitive_negative.") {
		rest := strings.TrimPrefix(id, "construction:en.ditransitive_negative.")
		parts := strings.Split(rest, ".")
		if len(parts) == 2 && (parts[0] == "present" || parts[0] == "present_plural" || parts[0] == "past") {
			articles := strings.Split(parts[1], "_")
			if len(articles) == 3 && (articles[0] == "a" || articles[0] == "the") && (articles[1] == "a" || articles[1] == "the") && (articles[2] == "a" || articles[2] == "the") {
				return englishDeterminers{subject: articles[0], target: articles[1], object: articles[2]}, true
			}
		}
	}
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
	if strings.HasPrefix(rest, ".past") {
		rest = strings.TrimPrefix(rest, ".past")
		if rest == "" {
			return englishDeterminers{subject: "the", object: "the", target: "the"}, true
		}
		variant, exists := variants[rest]
		return variant, exists
	}
	rest = strings.TrimPrefix(rest, ".plural")
	if rest == "" {
		return englishDeterminers{subject: "the", object: "the", target: "the"}, true
	}
	variant, exists := variants[rest]
	return variant, exists
}

var englishDeterminerVariants = map[string]englishDeterminers{
	"construction:en.transitive_past.svo.bare_object.subject_the":                                         {subject: "the", object: "bare"},
	"construction:en.transitive_past.svo.bare_object.subject_a":                                           {subject: "a", object: "bare"},
	"construction:en.transitive_past.svo.bare_object.subject_an":                                          {subject: "an", object: "bare"},
	"construction:en.transitive_past.svo.bare_subject.object_the":                                         {subject: "bare", object: "the"},
	"construction:en.transitive_past.svo.bare_subject.object_a":                                           {subject: "bare", object: "a"},
	"construction:en.transitive_past.svo.bare_subject.object_an":                                          {subject: "bare", object: "an"},
	"construction:en.transitive_past.svo.bare_subject.object_bare":                                        {subject: "bare", object: "bare"},
	"construction:en.transitive_past.svo.opening_object_past.the_plural":                                  {subject: "the", object: "the"},
	"construction:en.transitive_past.svo.opening_object_past.indefinite_subject.the_plural":               {subject: "a", object: "the"},
	"construction:en.transitive_past.svo.closing_object_past.the_plural":                                  {subject: "the", object: "the"},
	"construction:en.transitive_past.svo.closing_object_past.indefinite_subject.the_plural":               {subject: "a", object: "the"},
	"construction:en.transitive.svo.opening_object_present.indefinite_subject.the_plural":                 {subject: "a", object: "the"},
	"construction:en.transitive.svo.opening_object_present.indefinite_subject.bare_plural":                {subject: "a", object: "bare"},
	"construction:en.transitive.svo.opening_object_present.definite_subject.the_plural":                   {subject: "the", object: "the"},
	"construction:en.transitive.svo.opening_object_present.definite_subject.bare_plural":                  {subject: "the", object: "bare"},
	"construction:en.transitive_negative.svo.opening_object.present.subject_the.object_the_plural":        {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.opening_object.present.subject_a.object_the_plural":          {subject: "a", object: "the"},
	"construction:en.transitive_negative.svo.past.opening_object.past.subject_the.object_the_plural":      {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.past.opening_object.past.subject_a.object_the_plural":        {subject: "a", object: "the"},
	"construction:en.transitive_negative.svo.closing_object.present.subject_the.object_the_plural":        {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.closing_object.present.subject_a.object_the_plural":          {subject: "a", object: "the"},
	"construction:en.transitive_negative.svo.closing_object.present.plural.subject_the.object_the_plural": {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.past.closing_object.subject_the.object_the_plural":           {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.past.closing_object.subject_a.object_the_plural":             {subject: "a", object: "the"},
	"construction:en.transitive_negative.svo.past.closing_object.plural.subject_the.object_the_plural":    {subject: "the", object: "the"},
	"construction:en.attribution.bare_plural.past":                                                        {subject: "bare"},
	"construction:en.intransitive.sv.wake_up":                                                             {subject: "the"},
	"construction:en.intransitive.sv.wake_up.indefinite_subject":                                          {subject: "a"},
	"construction:en.intransitive.sv.wake_up.indefinite_subject_an":                                       {subject: "an"},
	"construction:en.intransitive.sv.wake_up.bare_subject_plural":                                         {subject: "bare"},
	"construction:en.intransitive.sv.fall_asleep":                                                         {subject: "the"},
	"construction:en.intransitive.sv.fall_asleep.indefinite_subject":                                      {subject: "a"},
	"construction:en.intransitive.sv.fall_asleep.indefinite_subject_an":                                   {subject: "an"},
	"construction:en.intransitive.sv.fall_asleep.bare_subject_plural":                                     {subject: "bare"},
	"construction:en.intransitive.sv.negative.singular":                                                   {subject: "the"},
	"construction:en.intransitive.sv.negative.singular.indefinite_subject":                                {subject: "a"},
	"construction:en.intransitive.sv.negative.singular.indefinite_subject_an":                             {subject: "an"},
	"construction:en.intransitive.sv.negative.plural":                                                     {subject: "bare"},
	"construction:en.intransitive.sv.negative.plural.definite":                                            {subject: "the"},
	"construction:en.intransitive.sv.negative.past.singular":                                              {subject: "the"},
	"construction:en.intransitive.sv.negative.past.singular.indefinite_subject":                           {subject: "a"},
	"construction:en.intransitive.sv.negative.past.singular.indefinite_subject_an":                        {subject: "an"},
	"construction:en.intransitive.sv.negative.past.plural":                                                {subject: "bare"},
	"construction:en.intransitive.sv.negative.past.plural.definite":                                       {subject: "the"},
	"construction:en.intransitive.sv":                                                                     {subject: "the"},
	"construction:en.intransitive.sv.indefinite_subject":                                                  {subject: "a"},
	"construction:en.intransitive.sv.indefinite_subject_an":                                               {subject: "an"},
	"construction:en.intransitive.sv.bare_subject_plural":                                                 {subject: "bare"},
	"construction:en.attributed_seeing.svo":                                                               {subject: "the", object: "the"},
	"construction:en.attributed_seeing.svo.demonstrative_this":                                            {subject: "this", object: "the"},
	"construction:en.attributed_seeing.svo.demonstrative_that":                                            {subject: "that", object: "the"},
	"construction:en.attributed_seeing.svo.demonstrative_these":                                           {subject: "these", object: "the"},
	"construction:en.attributed_seeing.svo.demonstrative_those":                                           {subject: "those", object: "the"},
	"construction:en.attributed_seeing.svo.indefinite_object":                                             {subject: "the", object: "a"},
	"construction:en.attributed_seeing.svo.indefinite_object_an":                                          {subject: "the", object: "an"},
	"construction:en.attributed_seeing.svo.indefinite_subject":                                            {subject: "a", object: "the"},
	"construction:en.attributed_seeing.svo.indefinite_subject_an":                                         {subject: "an", object: "the"},
	"construction:en.attributed_seeing.svo.indefinite_subject_object":                                     {subject: "a", object: "a"},
	"construction:en.attributed_seeing.svo.indefinite_subject_a_object_an":                                {subject: "a", object: "an"},
	"construction:en.attributed_seeing.svo.plural_subject":                                                {subject: "the", object: "the"},
	"construction:en.attributed_seeing.svo.plural_subject.indefinite_object":                              {subject: "the", object: "a"},
	"construction:en.attributed_seeing.svo.plural_subject.indefinite_object_an":                           {subject: "the", object: "an"},
	"construction:en.attributed_seeing_object.svo":                                                        {subject: "the", object: "the"},
	"construction:en.attributed_seeing_object_negative.svo":                                               {subject: "the", object: "the"},
	"construction:en.attributed_seeing_object_negative.svo.indefinite_subject":                            {subject: "a", object: "the"},
	"construction:en.attributed_seeing_object_negative.svo.indefinite_object":                             {subject: "the", object: "a"},
	"construction:en.attributed_seeing_object_negative.svo.indefinite_both":                               {subject: "a", object: "a"},
	"construction:en.attributed_seeing_object_negative.svo.past":                                          {subject: "the", object: "the"},
	"construction:en.attributed_seeing_object_negative.svo.past.indefinite_subject":                       {subject: "a", object: "the"},
	"construction:en.attributed_seeing_object_negative.svo.past.indefinite_object":                        {subject: "the", object: "a"},
	"construction:en.attributed_seeing_object_negative.svo.past.indefinite_both":                          {subject: "a", object: "a"},
	"construction:en.attributed_seeing_object.svo.this_object":                                            {subject: "the", object: "this"},
	"construction:en.attributed_seeing_object.svo.that_object":                                            {subject: "the", object: "that"},
	"construction:en.attributed_seeing_object.svo.these_object":                                           {subject: "the", object: "these"},
	"construction:en.attributed_seeing_object.svo.those_object":                                           {subject: "the", object: "those"},
	"construction:en.attributed_seeing_object.svo.indefinite_object":                                      {subject: "the", object: "a"},
	"construction:en.attributed_seeing_object.svo.indefinite_object_an":                                   {subject: "the", object: "an"},
	"construction:en.attributed_seeing_object.svo.indefinite_subject":                                     {subject: "a", object: "the"},
	"construction:en.attributed_seeing_object.svo.indefinite_subject_an":                                  {subject: "an", object: "the"},
	"construction:en.attributed_seeing_object.svo.indefinite_subject_object":                              {subject: "a", object: "a"},
	"construction:en.attributed_seeing_object.svo.indefinite_subject_a_object_an":                         {subject: "a", object: "an"},
	"construction:en.attributed_seeing_object.svo.plural_object":                                          {subject: "the", object: "the"},
	"construction:en.attributed_seeing_object.svo.plural_object.indefinite_subject":                       {subject: "a", object: "the"},
	"construction:en.attributed_seeing_object.svo.plural_object.indefinite_subject_an":                    {subject: "an", object: "the"},
	"construction:en.attributed_seeing_both.svo":                                                          {subject: "the", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_this":                                       {subject: "this", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_that":                                       {subject: "that", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_these.plural_subject":                       {subject: "these", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_those.plural_subject":                       {subject: "those", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_these.plural_both":                          {subject: "these", object: "the"},
	"construction:en.attributed_seeing_both.svo.demonstrative_those.plural_both":                          {subject: "those", object: "the"},
	"construction:en.attributed_seeing_both.svo.indefinite_object":                                        {subject: "the", object: "a"},
	"construction:en.attributed_seeing_both.svo.indefinite_object_an":                                     {subject: "the", object: "an"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject":                                       {subject: "a", object: "the"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject_an":                                    {subject: "an", object: "the"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject_object":                                {subject: "a", object: "a"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject_an_object":                             {subject: "an", object: "a"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject_a_object_an":                           {subject: "a", object: "an"},
	"construction:en.attributed_seeing_both.svo.indefinite_subject_an_object_an":                          {subject: "an", object: "an"},
	"construction:en.attributed_seeing_both.svo.plural_both":                                              {subject: "the", object: "the"},
	"construction:en.attributed_seeing_both.svo.plural_subject":                                           {subject: "the", object: "the"},
	"construction:en.attributed_seeing_both.svo.plural_object":                                            {subject: "the", object: "the"},
	"construction:en.transitive.svo":                                                                      {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo":                                                             {subject: "the", object: "the"},
	"construction:en.transitive_negative.svo.plural":                                                      {subject: "the", object: "the"},
	"construction:en.transitive.svo.indefinite_object":                                                    {subject: "the", object: "a"},
	"construction:en.transitive.svo.indefinite_subject":                                                   {subject: "a", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_object":                                            {subject: "a", object: "a"},
	"construction:en.transitive.svo.mass_object":                                                          {subject: "the", object: "bare"},
	"construction:en.transitive.svo.indefinite_subject_mass_object":                                       {subject: "a", object: "bare"},
	"construction:en.transitive.svo.indefinite_object_an":                                                 {subject: "the", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an":                                                {subject: "an", object: "the"},
	"construction:en.transitive.svo.indefinite_subject_a_object_an":                                       {subject: "a", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_object_a":                                       {subject: "an", object: "a"},
	"construction:en.transitive.svo.indefinite_subject_an_object_an":                                      {subject: "an", object: "an"},
	"construction:en.transitive.svo.indefinite_subject_an_mass_object":                                    {subject: "an", object: "bare"},
	"construction:en.oblique.svo":                                                                         {subject: "the", target: "the"},
	"construction:en.oblique.svo.indefinite_target":                                                       {subject: "the", target: "a"},
	"construction:en.oblique.svo.indefinite_subject":                                                      {subject: "a", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_target":                                               {subject: "a", target: "a"},
	"construction:en.oblique.svo.indefinite_target_an":                                                    {subject: "the", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an":                                                   {subject: "an", target: "the"},
	"construction:en.oblique.svo.indefinite_subject_a_target_an":                                          {subject: "a", target: "an"},
	"construction:en.oblique.svo.indefinite_subject_an_target_a":                                          {subject: "an", target: "a"},
	"construction:en.oblique.svo.indefinite_subject_an_target_an":                                         {subject: "an", target: "an"},
	"construction:en.to_oblique.svo":                                                                      {subject: "the", target: "the"},
	"construction:en.to_oblique.svo.indefinite_target":                                                    {subject: "the", target: "a"},
	"construction:en.to_oblique.svo.indefinite_target_an":                                                 {subject: "the", target: "an"},
	"construction:en.to_oblique.svo.indefinite_subject":                                                   {subject: "a", target: "the"},
	"construction:en.to_oblique.svo.indefinite_subject_an":                                                {subject: "an", target: "the"},
	"construction:en.to_oblique.svo.indefinite_subject_target":                                            {subject: "a", target: "a"},
	"construction:en.to_oblique.svo.indefinite_subject_a_target_an":                                       {subject: "a", target: "an"},
	"construction:en.to_oblique.svo.indefinite_subject_an_target_a":                                       {subject: "an", target: "a"},
	"construction:en.to_oblique.svo.indefinite_subject_an_target_an":                                      {subject: "an", target: "an"},
	"construction:en.with_oblique.svo":                                                                    {subject: "the", target: "the"},
	"construction:en.with_oblique.svo.indefinite_target":                                                  {subject: "the", target: "a"},
	"construction:en.with_oblique.svo.indefinite_target_an":                                               {subject: "the", target: "an"},
	"construction:en.with_oblique.svo.indefinite_subject":                                                 {subject: "a", target: "the"},
	"construction:en.with_oblique.svo.indefinite_subject_an":                                              {subject: "an", target: "the"},
	"construction:en.with_oblique.svo.indefinite_subject_target":                                          {subject: "a", target: "a"},
	"construction:en.with_oblique.svo.indefinite_subject_a_target_an":                                     {subject: "a", target: "an"},
	"construction:en.with_oblique.svo.indefinite_subject_an_target_a":                                     {subject: "an", target: "a"},
	"construction:en.with_oblique.svo.indefinite_subject_an_target_an":                                    {subject: "an", target: "an"},
	"construction:en.transitive_with_oblique.svo":                                                         {subject: "the", object: "bare", target: "the"},
	"construction:en.transitive_with_oblique.svo.indefinite_target":                                       {subject: "the", object: "bare", target: "a"},
	"construction:en.transitive_with_oblique.svo.indefinite_subject":                                      {subject: "a", object: "bare", target: "the"},
	"construction:en.transitive_with_oblique.svo.indefinite_subject_target":                               {subject: "a", object: "bare", target: "a"},
	"construction:en.oblique_negative.svo":                                                                {subject: "the", target: "the"},
	"construction:en.oblique_negative.svo.plural":                                                         {subject: "the", target: "the"},
}

var englishNegativeTransitiveDeterminerVariants = map[string]englishDeterminers{
	".mass_object":                                                 {subject: "the", object: "bare"},
	".mass_object.indefinite_subject":                              {subject: "a", object: "bare"},
	".bare_object.subject_the":                                     {subject: "the", object: "bare"},
	".bare_object.subject_a":                                       {subject: "a", object: "bare"},
	".bare_object.subject_an":                                      {subject: "an", object: "bare"},
	".plural.bare_object.subject_the":                              {subject: "the", object: "bare"},
	".indefinite_object":                                           {subject: "the", object: "a"},
	".indefinite_subject":                                          {subject: "a", object: "the"},
	".indefinite_subject_object":                                   {subject: "a", object: "a"},
	".indefinite_object_an":                                        {subject: "the", object: "an"},
	".indefinite_subject_an":                                       {subject: "an", object: "the"},
	".indefinite_subject_a_object_an":                              {subject: "a", object: "an"},
	".indefinite_subject_an_object_a":                              {subject: "an", object: "a"},
	".indefinite_subject_an_object_an":                             {subject: "an", object: "an"},
	".closing_object.present.subject_the.object_the_plural":        {subject: "the", object: "the"},
	".closing_object.present.subject_a.object_the_plural":          {subject: "a", object: "the"},
	".closing_object.present.plural.subject_the.object_the_plural": {subject: "the", object: "the"},
	".closing_object.subject_the.object_the_plural":                {subject: "the", object: "the"},
	".closing_object.subject_a.object_the_plural":                  {subject: "a", object: "the"},
	".closing_object.plural.subject_the.object_the_plural":         {subject: "the", object: "the"},
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
	case "construction:en.intransitive":
		return "construction:en.intransitive.sv"
	case "construction:en.transitive":
		return "construction:en.transitive.svo"
	case "construction:en.transitive_past":
		return "construction:en.transitive_past.svo"
	case "construction:en.transitive_negative":
		return "construction:en.transitive_negative.svo"
	case "construction:en.ditransitive_negative":
		return "construction:en.ditransitive_negative.present.the_the_a"
	case "construction:en.oblique":
		return "construction:en.oblique.svo"
	case "construction:en.to_oblique":
		return "construction:en.to_oblique.svo"
	case "construction:en.with_oblique":
		return "construction:en.with_oblique.svo"
	case "construction:en.transitive_with_oblique":
		return "construction:en.transitive_with_oblique.svo"
	case "construction:en.oblique_negative":
		return "construction:en.oblique_negative.svo"
	case "construction:en.attributed_seeing":
		return "construction:en.attributed_seeing.svo"
	case "construction:en.attributed_seeing_object":
		return "construction:en.attributed_seeing_object.svo"
	case "construction:en.attributed_seeing_both":
		return "construction:en.attributed_seeing_both.svo"
	default:
		return ""
	}
}

func englishDeterminerForRole(candidate realizer.TargetCandidate, role meaning.Role) (string, bool) {
	familyID := candidate.Family.Family.ID.String()
	if familyID == "construction:en.give.transfer" || familyID == "construction:en.give.transfer.past" {
		if candidate.Sense.ID.String() != "sense:en.give.transfer" {
			return "", false
		}
		if familyID == "construction:en.give.transfer.past" {
			if candidate.Frame.ID != "en.give.transfer.past" {
				return "", false
			}
		} else if candidate.Frame.ID != "en.give.transfer" {
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
					if previous.Kind == construction.TokenSlot && index > 1 {
						for _, modifierSlot := range candidate.Family.Family.SlotSpecs {
							if modifierSlot.ID != previous.SlotID || modifierSlot.ModifierBinding == nil || modifierSlot.ModifierBinding.TargetSlot != slot.ID {
								continue
							}
							beforeModifier := candidate.CompiledVariant.Variant.Sequence[index-2]
							if beforeModifier.Kind == construction.TokenTerminal && isEnglishDeterminerToken(beforeModifier.Terminal) {
								return beforeModifier.Terminal, true
							}
						}
					}
				}
				return "bare", true
			}
		}
		return "", false
	}
	if familyID != "construction:en.intransitive" && familyID != "construction:en.transitive" && familyID != "construction:en.transitive_past" && familyID != "construction:en.transitive_negative" && familyID != "construction:en.ditransitive_negative" && familyID != "construction:en.oblique" && familyID != "construction:en.to_oblique" && familyID != "construction:en.with_oblique" && familyID != "construction:en.transitive_with_oblique" && familyID != "construction:en.oblique_negative" && familyID != "construction:en.attribution_negative" && familyID != "construction:en.attributed_seeing" && familyID != "construction:en.attributed_seeing_object" && familyID != "construction:en.attributed_seeing_object_negative" && familyID != "construction:en.attributed_seeing_both" && familyID != "construction:en.transitive_personal" {
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
	case "subject", "bearer":
		return variant.subject
	case "object", "theme":
		return variant.object
	case "target", "recipient":
		return variant.target
	default:
		return ""
	}
}

func englishDeterminerFeature(article string) string {
	switch article {
	case "this", "that", "these", "those":
		return "demonstrative"
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

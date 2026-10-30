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
	"github.com/clickbytes/golingua/internal/morph"
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
	phraseProfile, err := SentencePhraseProfile(profile)
	if err != nil {
		return runtime.MorphologyExecutionBuilder{}, err
	}
	return runtime.MorphologyExecutionBuilder{
		Profile: profile, PhraseProfile: phraseProfile, Paradigms: paradigms, Exceptions: exceptions,
		Parts: seed.PrincipalParts, FeaturePlan: polishSentenceFeatures{paradigms: paradigms}, CaseFeature: "case",
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
	profile, _, paradigms, _ := seed.RuntimeMorphology()
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
	negationRoutes := newPolishNegationRouteResolver(snapshot.Lexicon(), valencyIndex, constructions, packages)
	_, variants, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
	if err != nil {
		return nil, err
	}
	execution, err := NewSentenceExecutionBuilderWithSeed(seed)
	if err != nil {
		return nil, err
	}
	execution.FeaturePlan = polishSentenceFeatures{negationRoutes: negationRoutes, paradigms: paradigms}
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
		OperatorBuilder: polishSentenceOperatorBuilder{negationRoutes: negationRoutes}, ExecutionBuilder: execution,
		CandidatePolicy:      polishSentenceCandidatePolicy{negationRoutes: negationRoutes},
		PlanBuilder:          polishSentencePlanAnnotations{negationRoutes: negationRoutes, paradigms: paradigms},
		ClauseCoordination:   &coordination,
		ScopedRealizer:       &runtime.ScopedClauseRealizer{Language: "pl", Plan: scopePlan, Valency: valencyIndex, Morphology: seed, Features: polishScopedFeaturePlanner{}},
		RelationPathRealizer: &runtime.RelationPathRealizer{Language: "pl", Plan: relationPlan, Morphology: seed, Features: polishRelationPathFeaturePlanner{}},
	}, nil
}

type polishSentencePlanAnnotations struct {
	negationRoutes polishNegationRouteResolver
	paradigms      *morph.CompiledParadigms
}

func (plan polishSentencePlanAnnotations) Annotations(_ context.Context, source semantic.MeaningCandidate, selected realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error) {
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
		_, contextOK := polishTargetIdentificationReferenceMode(source, selected.NodeID)
		predicate := findSentencePredicate(*source.Meaning, selected.NodeID)
		if !predicateFound || !referentFound || !contextOK ||
			!polishIdentificationAttributesAllowed(source, selected.NodeID, role, referentID, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("Polish identification realization requires a valid unresolved %s reference context", role)
		}
		assumptions := polishIdentificationDefaults(predicatePath, referentPath, predicate, source.Meaning.Nodes, referentID, "identification")
		var losses []realizer.FieldLoss
		if role == "bearer" {
			referent := findSentenceReferent(*source.Meaning, referentID)
			conceptID, err := ids.ParseConceptID(string(referent.Concept))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("Polish identification source concept %q is invalid: %w", referent.Concept, err)
			}
			if values, found := runtime.SourceFeatures(source, "bearer", conceptID); found {
				determiner := string(values["determiner"])
				if determiner == "definite" || determiner == "indefinite" {
					losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "determiner", Outcome: status.OutcomeIncomplete, Reason: "Polish noun phrase does not encode the source English article distinction"})
				}
			}
		}
		return nil, assumptions, losses, nil
	}
	if source.Meaning == nil || selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState {
		return nil, nil, nil, nil
	}
	if role, referentID, _, attribution := polishIdentificationStateShape(source, selected.NodeID); attribution && role == "bearer" {
		predicatePath, predicateFound := runtime.MeaningPath(*source.Meaning, selected.NodeID)
		referentPath, referentFound := runtime.MeaningPath(*source.Meaning, referentID)
		predicate := findSentencePredicate(*source.Meaning, selected.NodeID)
		referent := findSentenceReferent(*source.Meaning, referentID)
		if !predicateFound || !referentFound || referent.ID == "" || !polishPredicateEvidenceConsistent(source, selected) ||
			!polishIdentificationAttributesAllowed(source, selected.NodeID, role, referentID, selected.TransferBinding.CoreRef) {
			return nil, nil, nil, fmt.Errorf("Polish attribution realization requires a valid anchored bearer")
		}
		assumptions := polishIdentificationDefaults(predicatePath, referentPath, predicate, source.Meaning.Nodes, referentID, "attribution")
		var losses []realizer.FieldLoss
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Polish attribution source concept %q is invalid: %w", referent.Concept, err)
		}
		if values, found := runtime.SourceFeatures(source, role, conceptID); found {
			determiner := string(values["determiner"])
			if determiner == "definite" || determiner == "indefinite" {
				losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "determiner", Outcome: status.OutcomeIncomplete, Reason: "Polish noun phrase does not encode the source English article distinction"})
			}
		}
		return nil, assumptions, losses, nil
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
	if predicate.TemporalFrame == meaning.TemporalFramePast && (selected.TransferBinding.CoreRef.String() == "concept:giving" || selected.TransferBinding.CoreRef.String() == "concept:buying") {
		aspect := ""
		switch selected.TransferBinding.SenseID.String() {
		case "sense:pl.dac.transfer":
			aspect = "perfective"
		case "sense:pl.dawac.transfer":
			aspect = "imperfective"
		case "sense:pl.kupic.buy":
			aspect = "perfective"
		case "sense:pl.kupowac.buy":
			aspect = "imperfective"
		}
		if aspect != "" {
			values, known := runtime.SourceFeatures(source, "", selected.TransferBinding.CoreRef)
			if !known || string(values["aspect"]) == "" {
				reason := "Meaning does not specify Polish aspect; this candidate selects the lexical aspect of the Polish buying predicate"
				if selected.TransferBinding.CoreRef.String() == "concept:giving" {
					reason = "Meaning does not specify Polish aspect; this candidate selects the lexical aspect of dać or dawać"
				}
				assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "aspect", Value: aspect, Reason: reason})
			}
		}
	}
	if predicate.TemporalFrame == meaning.TemporalFramePast && selected.Kind == meaning.KindEvent &&
		polishPastGenericTargetFamily(selected.Family.Family.ID.String()) && !polishDedicatedPastTransfer(selected) &&
		polishApprovedPastTargetCandidate(source, selected) {
		// Replace any older per-sense annotation with one based on the selected
		// candidate's actual past paradigm. This also covers negative past plans.
		filtered := assumptions[:0]
		for _, assumption := range assumptions {
			if assumption.Field == "aspect" && sameMeaningPath(assumption.Path, predicatePath) {
				continue
			}
			filtered = append(filtered, assumption)
		}
		assumptions = filtered
		values, known := runtime.SourceFeatures(source, "", selected.TransferBinding.CoreRef)
		hasAspectEvidence := known && values["aspect"] != ""
		if !hasAspectEvidence && !polishHasKnownAspectEvidence(source, selected.TransferBinding.CoreRef) {
			aspect, aspectKnown := polishPastLexicalAspect(selected, plan.paradigms)
			if aspectKnown {
				assumptions = append(assumptions, realizer.FieldAssumption{
					Path: predicatePath, Field: "aspect", Value: aspect,
					Reason: "Meaning and source evidence do not specify Polish aspect; this candidate selects the lexical aspect of its approved past paradigm",
				})
			}
		}
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
	var losses []realizer.FieldLoss
	familyID := selected.Family.Family.ID.String()
	losses = append(losses, polishEventDeterminerLosses(source, selected)...)
	if selected.Kind == meaning.KindEvent && (familyID == "construction:pl.transitive" || familyID == "construction:pl.transitive_negative") &&
		((selected.TransferBinding.CoreRef.String() == "concept:opening_object" && selected.TransferBinding.SenseID.String() == "sense:pl.otwierac_obiekt.open") ||
			(selected.TransferBinding.CoreRef.String() == "concept:closing_object" && selected.TransferBinding.SenseID.String() == "sense:pl.zamykac_obiekt.close")) {
		for _, participant := range predicate.Participants {
			if participant.Role != "object" {
				continue
			}
			referent, exists := polishMeaningNode(source, participant.Target)
			if !exists || referent.Concept != "concept:door" || referent.Multiplicity == "" {
				continue
			}
			path, pathFound := runtime.MeaningPath(*source.Meaning, participant.Target)
			if pathFound && selected.Lexeme.ID.String() == "lexeme:pl.drzwi.n" {
				losses = append(losses, realizer.FieldLoss{Path: path, Field: "number", Outcome: status.OutcomeIncomplete, Reason: "Polish drzwi is lexically plural but does not distinguish one door from multiple doors"})
			}
		}
	}
	// Polish does not mark English article definiteness on the noun phrase.
	// Keep the source distinction visible on the exact participant referent.
	if role, referentID, _, attribution := polishIdentificationStateShape(source, predicate.ID); attribution && role == "bearer" {
		referent := findSentenceReferent(*source.Meaning, referentID)
		if referentPath, pathFound := runtime.MeaningPath(*source.Meaning, referentID); pathFound && referent.ID != "" {
			conceptID, err := ids.ParseConceptID(string(referent.Concept))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("Polish participant source concept %q is invalid: %w", referent.Concept, err)
			}
			if values, found := runtime.SourceFeatures(source, role, conceptID); found {
				determiner := string(values["determiner"])
				if determiner == "definite" || determiner == "indefinite" {
					losses = append(losses, realizer.FieldLoss{Path: referentPath, Field: "determiner", Outcome: status.OutcomeIncomplete, Reason: "Polish noun phrase does not encode the source English article distinction"})
				}
			}
		}
	}
	return nil, assumptions, losses, nil
}

func polishEventDeterminerLosses(source semantic.MeaningCandidate, selected realizer.TargetCandidate) []realizer.FieldLoss {
	if source.Meaning == nil || selected.Kind != meaning.KindEvent ||
		selected.Family.Family.Language != "pl" ||
		!selected.TransferBinding.TransferAuthorized() ||
		selected.TransferBinding.SenseID != selected.Sense.ID ||
		selected.TransferBinding.CoreRef.String() != string(findSentencePredicate(*source.Meaning, selected.NodeID).Concept) ||
		!selected.ConstructionBinding.Approved || selected.ConstructionBinding.SenseID != selected.Sense.ID ||
		selected.ConstructionBinding.ValencyFrameID != selected.Frame.ID || selected.ConstructionBinding.FamilyID != selected.Family.Family.ID ||
		selected.ConstructionBinding.VariantID.String() != selected.CompiledVariant.Variant.ID.String() ||
		selected.Frame.SenseID != selected.Sense.ID || !polishFrameAllowsFamily(selected.Frame, selected.Family.Family.ID) ||
		selected.CompiledVariant.FamilyID != selected.Family.Family.ID || selected.CompiledVariant.Variant.FamilyID != selected.Family.Family.ID ||
		selected.CompiledVariant.Variant.ID.String() != selected.ConstructionBinding.VariantID.String() ||
		selected.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze ||
		selected.Family.Family.Validate() != nil || selected.Frame.Validate() != nil ||
		selected.CompiledVariant.Variant.Validate(selected.Family.Family) != nil {
		return nil
	}
	event := findSentencePredicate(*source.Meaning, selected.NodeID)
	if event.ID == "" || event.Kind != meaning.KindEvent || string(event.Concept) != selected.TransferBinding.CoreRef.String() {
		return nil
	}
	compiled, err := construction.ValidateVariantForFrame(selected.CompiledVariant.Variant, selected.Family.Family, selected.Frame)
	if err != nil || compiled.FamilyID != selected.Family.Family.ID || compiled.Variant.ID.String() != selected.ConstructionBinding.VariantID.String() ||
		compiled.PlanID != selected.Family.PlanID || selected.CompiledVariant.PlanID != selected.Family.PlanID {
		return nil
	}

	arguments := make(map[meaning.Role]struct{}, len(selected.Frame.Arguments))
	for _, argument := range selected.Frame.Arguments {
		if argument.Role == "" {
			return nil
		}
		if _, duplicate := arguments[argument.Role]; duplicate {
			return nil
		}
		arguments[argument.Role] = struct{}{}
	}
	participants := make(map[meaning.Role]meaning.Node, len(event.Participants))
	for _, participant := range event.Participants {
		if participant.Role == "" || participant.Target == "" {
			return nil
		}
		if _, duplicate := participants[participant.Role]; duplicate {
			return nil
		}
		if _, licensed := arguments[participant.Role]; !licensed {
			return nil
		}
		referent := findSentenceReferent(*source.Meaning, participant.Target)
		if referent.ID == "" || referent.Kind != meaning.KindReferent {
			return nil
		}
		participants[participant.Role] = referent
	}

	var losses []realizer.FieldLoss
	seen := make(map[meaning.NodeID]struct{}, len(participants))
	for _, argument := range selected.Frame.Arguments {
		role := argument.Role
		referent, present := participants[role]
		if !present {
			continue
		}
		npSlots := 0
		for _, slot := range selected.Family.Family.SlotSpecs {
			if slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == role && slot.Category == construction.CategoryNP {
				npSlots++
			}
		}
		if npSlots != 1 {
			return nil
		}
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil {
			continue
		}
		values, known := runtime.SourceFeatures(source, role, conceptID)
		if !known || (values["determiner"] != "definite" && values["determiner"] != "indefinite") {
			continue
		}
		if _, duplicate := seen[referent.ID]; duplicate {
			continue
		}
		path, found := runtime.MeaningPath(*source.Meaning, referent.ID)
		if !found {
			continue
		}
		duplicate := false
		for _, prior := range losses {
			if prior.Field == "determiner" && sameMeaningPath(prior.Path, path) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			losses = append(losses, realizer.FieldLoss{
				Path: path, Field: "determiner", Outcome: status.OutcomeIncomplete,
				Reason: "Polish noun phrase does not encode the explicit English article distinction",
			})
			seen[referent.ID] = struct{}{}
		}
	}
	return losses
}

func polishIdentificationStateShape(source semantic.MeaningCandidate, predicateID meaning.NodeID) (meaning.Role, meaning.NodeID, string, bool) {
	role, referent, concept, ok := polishIdentificationState(source, predicateID)
	if !ok || role != "bearer" {
		return "", "", "", false
	}
	if _, attribution := polishAttributionForConcept(concept); !attribution {
		return "", "", "", false
	}
	return role, referent, concept, true
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

type polishSentenceFeatures struct {
	negationRoutes polishNegationRouteResolver
	paradigms      *morph.CompiledParadigms
}

func (featurePlan polishSentenceFeatures) Features(_ context.Context, source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slot construction.SlotSpec, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) ([]features.FeatureValue, error) {
	if slot.NominalPredicate {
		state, stateExists := polishMeaningNode(source, predicate.NodeID)
		if slot.Category != construction.CategoryNP || predicate.Kind != meaning.KindState || source.Meaning == nil || candidate.NodeID != predicate.NodeID || candidate.Kind != meaning.KindState ||
			!predicate.TransferBinding.TransferAuthorized() || predicate.TransferBinding.CoreRef.String() != "concept:legal_ownership" || !stateExists || !state.RelationPredicate || state.TemporalFrame != "" || state.Concept != "concept:legal_ownership" ||
			predicate.Family.Family.ID.String() != "construction:pl.legal_ownership_nominal" || candidate.Lexeme.POS != "noun" || slot.Case != "instrumental" {
			return nil, fmt.Errorf("Polish nominal predicate requires the approved legal ownership noun route")
		}
		return knownPolishFeatures([2]string{"case", "instrumental"}, [2]string{"number", "singular"}, [2]string{"gender", "masculine_personal"}), nil
	}
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
		_, valid := polishAttributedSeeingShape(source)
		if !valid || source.Meaning == nil {
			return nil, fmt.Errorf("Polish attributed seeing adjective requires the licensed two-root Meaning")
		}
		modifier, modifierFound := realizer.FindModifier(*source.Meaning, predicate.NodeID, predicate.Family.Family, slot.ID)
		if !polishAttributedSeeingContextsValid(source) || !modifierFound || candidate.NodeID != modifier.ID ||
			!polishAttributedSeeingModifierCandidate(source, candidate, candidates) {
			return nil, fmt.Errorf("Polish attributed seeing adjective is not linked to its selected bearer property")
		}
		targetReferent, targetCase, found := polishAttributedSeeingModifierReferent(source, predicate, slot)
		if !found {
			return nil, fmt.Errorf("Polish attributed seeing modifier has no unique target argument")
		}
		gender, number, found := polishIdentificationReferentFeatures(source, targetReferent, candidates)
		if !found {
			return nil, fmt.Errorf("Polish attributed seeing adjective has no unique compatible bearer lexical class")
		}
		return knownPolishFeatures(
			[2]string{"degree", "positive"}, [2]string{"number", number},
			[2]string{"gender", gender}, [2]string{"agreement_gender", gender}, [2]string{"case", targetCase},
		), nil
	}
	if slot.Category == construction.CategoryAP {
		attribution, allowed := polishAttributionForSense(predicate.Sense.ID.String())
		if !allowed || candidate.Sense.ID.String() != predicate.Sense.ID.String() || candidate.Lexeme.ID.String() != attribution.lexeme || candidate.TransferBinding.CoreRef.String() != attribution.concept {
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
			[2]string{"gender", gender}, [2]string{"agreement_gender", gender}, [2]string{"case", "nominative"},
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
		if _, event, route, found := featurePlan.negationRoutes.negatedPresent(source); found && event.ID == predicate.NodeID && (predicate.Sense.ID.String() == route.sense || route.concept == "concept:giving" && polishGivingNegationSenseMatches(route, predicate.Sense.ID.String()) || route.concept == "concept:buying" && polishBuyingNegationSenseMatches(route, predicate.Sense.ID.String())) {
			if route.concept == "concept:giving" && predicate.Sense.ID.String() == "sense:pl.dac.transfer" && event.TemporalFrame != meaning.TemporalFramePast {
				return nil, fmt.Errorf("perfective Polish giving target requires a past temporal frame")
			}
			number, known := polishNegationNumber(source, predicate, route, candidates, featurePlan.negationRoutes)
			if !known {
				return nil, fmt.Errorf("Polish negative predicate cannot realize its Meaning number or temporal frame")
			}
			dynamicPast := event.TemporalFrame == meaning.TemporalFramePast && (route.family == "construction:pl.intransitive" || route.derived)
			dedicatedPast := event.TemporalFrame == meaning.TemporalFramePast && polishPastNegativeConceptAllowed(route.concept)
			if dynamicPast || dedicatedPast {
				if dynamicPast && !polishApprovedPastTargetCandidate(source, predicate) {
					return nil, fmt.Errorf("Polish past negative candidate lacks its exact approved past construction route")
				}
				gender, known := polishSubjectPastGender(source, predicate, candidates, number)
				if !known {
					return nil, fmt.Errorf("Polish past negative predicate cannot realize subject gender agreement")
				}
				aspect := "imperfective"
				if dedicatedPast {
					aspect, known = polishDedicatedPastAspect(predicate)
					if !known {
						return nil, fmt.Errorf("Polish giving or buying past route has no dedicated aspect")
					}
				} else {
					aspect, known = polishPastLexicalAspect(predicate, featurePlan.paradigms)
					if !known {
						return nil, fmt.Errorf("Polish past negative candidate has no past aspect in its selected paradigm")
					}
				}
				return knownPolishFeatures(
					[2]string{"verb_form", "finite"}, [2]string{"tense", "past"}, [2]string{"aspect", aspect},
					[2]string{"person", "third"}, [2]string{"number", number}, [2]string{"gender", gender},
					[2]string{"mood", "indicative"}, [2]string{"voice", "active"},
				), nil
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
			if polishPastGenericTargetFamily(predicate.Family.Family.ID.String()) && !polishDedicatedPastTransfer(predicate) {
				if !polishApprovedPastTargetCandidate(source, predicate) {
					return nil, fmt.Errorf("Polish past target candidate lacks its exact approved construction route")
				}
				var aspectKnown bool
				aspect, aspectKnown = polishPastLexicalAspect(predicate, featurePlan.paradigms)
				if !aspectKnown {
					return nil, fmt.Errorf("Polish past candidate has no aspect in its selected paradigm")
				}
			} else if polishDedicatedPastTransfer(predicate) {
				var aspectKnown bool
				aspect, aspectKnown = polishDedicatedPastAspect(predicate)
				if !aspectKnown {
					return nil, fmt.Errorf("Polish giving or buying past route has no dedicated aspect")
				}
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
	agreementGender := ""
	for _, property := range candidate.Lexeme.Properties {
		if property.Key == "gender" {
			gender = property.Value
		} else if property.Key == "agreement_gender" {
			agreementGender = property.Value
		}
	}
	if gender == "" && !pluralOnly {
		return nil, nil
	}
	featureValues := [][2]string{{"case", slot.Case}, {"number", string(number)}}
	if gender != "" {
		featureValues = append(featureValues, [2]string{"gender", gender})
	}
	if agreementGender == "" && gender != "" {
		agreementGender = gender
		if string(number) == "plural" {
			agreementGender = polishPluralAgreementGender(gender)
		}
	}
	if agreementGender != "" {
		featureValues = append(featureValues, [2]string{"agreement_gender", agreementGender})
	}
	return knownPolishFeatures(featureValues...), nil
}

func polishPastNegativeConceptAllowed(concept string) bool {
	switch concept {
	case "concept:buying", "concept:giving":
		return true
	default:
		return false
	}
}

func polishPastGenericTargetFamily(family string) bool {
	switch family {
	case "construction:pl.intransitive", "construction:pl.transitive", "construction:pl.transitive_to", "construction:pl.transitive_negative", "construction:pl.attributed_seeing_object_negative":
		return true
	default:
		return false
	}
}

func polishDedicatedPastTransfer(candidate realizer.TargetCandidate) bool {
	concept := candidate.TransferBinding.CoreRef.String()
	return candidate.Kind == meaning.KindEvent && (concept == "concept:giving" || concept == "concept:buying")
}

func polishDedicatedPastAspect(candidate realizer.TargetCandidate) (string, bool) {
	switch candidate.TransferBinding.SenseID.String() {
	case "sense:pl.dac.transfer", "sense:pl.kupic.buy":
		return "perfective", candidate.TransferBinding.CoreRef.String() == "concept:giving" || candidate.TransferBinding.CoreRef.String() == "concept:buying"
	case "sense:pl.dawac.transfer", "sense:pl.kupowac.buy":
		return "imperfective", candidate.TransferBinding.CoreRef.String() == "concept:giving" || candidate.TransferBinding.CoreRef.String() == "concept:buying"
	default:
		return "", false
	}
}

// polishApprovedPastTargetCandidate checks only the selected candidate's
// already-approved construction path. The selected family and variant remain
// the source of truth; the paradigm supplies the lexical aspect separately.
func polishApprovedPastTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	licensedSurfaceOmission := len(candidate.MissingOmittedRoles) > 0
	declaredOmittedRoles := make(map[meaning.Role]bool, len(candidate.Frame.SurfaceOmittedRoles))
	for _, role := range candidate.Frame.SurfaceOmittedRoles {
		if role == "" || declaredOmittedRoles[role] {
			licensedSurfaceOmission = false
			break
		}
		declaredOmittedRoles[role] = true
	}
	seenOmittedRoles := make(map[meaning.Role]bool, len(candidate.MissingOmittedRoles))
	for _, role := range candidate.MissingOmittedRoles {
		if role == "" || !declaredOmittedRoles[role] || seenOmittedRoles[role] {
			licensedSurfaceOmission = false
			break
		}
		seenOmittedRoles[role] = true
	}
	outcomeAllowed := candidate.Outcome == status.OutcomeOK || candidate.Outcome == status.OutcomeIncomplete && licensedSurfaceOmission
	if source.Meaning == nil || candidate.Kind != meaning.KindEvent || !outcomeAllowed ||
		len(candidate.MissingOmittedRoles) != 0 && !licensedSurfaceOmission || candidate.Family.Family.Language != "pl" ||
		!polishPastGenericTargetFamily(candidate.Family.Family.ID.String()) || polishDedicatedPastTransfer(candidate) ||
		!candidate.TransferBinding.TransferAuthorized() || candidate.TransferBinding.SenseID != candidate.Sense.ID ||
		candidate.TransferBinding.CoreRef.String() != string(findSentencePredicate(*source.Meaning, candidate.NodeID).Concept) ||
		!candidate.ConstructionBinding.Approved || candidate.ConstructionBinding.SenseID != candidate.Sense.ID ||
		candidate.ConstructionBinding.ValencyFrameID != candidate.Frame.ID || candidate.ConstructionBinding.FamilyID != candidate.Family.Family.ID ||
		candidate.ConstructionBinding.VariantID.String() != candidate.CompiledVariant.Variant.ID.String() ||
		candidate.Frame.SenseID != candidate.Sense.ID || !polishFrameAllowsFamily(candidate.Frame, candidate.Family.Family.ID) ||
		candidate.CompiledVariant.FamilyID != candidate.Family.Family.ID || candidate.CompiledVariant.Variant.FamilyID != candidate.Family.Family.ID ||
		candidate.CompiledVariant.Variant.ID.String() != candidate.ConstructionBinding.VariantID.String() ||
		candidate.Family.Family.Validate() != nil || candidate.Frame.Validate() != nil ||
		candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze ||
		candidate.CompiledVariant.Variant.Validate(candidate.Family.Family) != nil {
		return false
	}
	event := findSentencePredicate(*source.Meaning, candidate.NodeID)
	if event.ID == "" || event.TemporalFrame != meaning.TemporalFramePast {
		return false
	}
	negative, scoped := false, false
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindScope || node.Target != event.ID {
			continue
		}
		if scoped || node.Operator != meaning.OperatorNegation {
			return false
		}
		scoped, negative = true, true
	}
	familyID := candidate.Family.Family.ID.String()
	hasNie := false
	var mainPredicate construction.SlotSpec
	for _, slot := range candidate.Family.Family.SlotSpecs {
		if slot.Category == construction.CategoryVP && slot.IsMainPredicate() {
			if mainPredicate.ID != "" || slot.Optionality != construction.SlotRequired || slot.Cardinality.Minimum != 1 || slot.Cardinality.Maximum != 1 {
				return false
			}
			mainPredicate = slot
		}
	}
	if mainPredicate.ID == "" {
		return false
	}
	mainCount := 0
	for _, token := range candidate.CompiledVariant.Variant.Sequence {
		if token.Kind == construction.TokenSlot && token.SlotID == mainPredicate.ID {
			mainCount++
		}
		if token.Kind == construction.TokenTerminal && token.Terminal == "nie" {
			hasNie = true
		}
	}
	if mainCount != 1 {
		return false
	}
	switch familyID {
	case "construction:pl.intransitive":
		if hasNie != negative {
			return false
		}
	case "construction:pl.transitive", "construction:pl.transitive_to":
		if negative || hasNie {
			return false
		}
	case "construction:pl.transitive_negative", "construction:pl.attributed_seeing_object_negative":
		if !negative || !hasNie {
			return false
		}
		if familyID == "construction:pl.attributed_seeing_object_negative" {
			shape, valid := polishAttributedSeeingShape(source)
			if !valid || shape.scope.ID == "" || shape.scope.Target != event.ID || shape.event.Concept != "concept:seeing" {
				return false
			}
		}
	}
	participants := make(map[meaning.Role]meaning.NodeID, len(event.Participants))
	for _, participant := range event.Participants {
		if participant.Role == "" || participant.Target == "" || participants[participant.Role] != "" {
			return false
		}
		participants[participant.Role] = participant.Target
	}
	subjects := 0
	for _, argument := range candidate.Frame.Arguments {
		if argument.Required && participants[argument.Role] == "" {
			return false
		}
		if _, present := participants[argument.Role]; present {
			matchedSlot := false
			for _, slot := range candidate.Family.Family.SlotSpecs {
				if slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == argument.Role && slot.Category == construction.CategoryNP {
					matchedSlot = true
					break
				}
			}
			if !matchedSlot {
				return false
			}
		}
		if argument.GrammaticalFunction == "subject" {
			subjects++
			if !argument.Required || argument.Case != "nominative" || participants[argument.Role] == "" {
				return false
			}
		}
	}
	if subjects != 1 {
		return false
	}
	compiled, err := construction.ValidateVariantForFrame(candidate.CompiledVariant.Variant, candidate.Family.Family, candidate.Frame)
	if err != nil || compiled.FamilyID != candidate.Family.Family.ID || compiled.Variant.ID != candidate.CompiledVariant.Variant.ID ||
		compiled.PlanID != candidate.Family.PlanID || candidate.CompiledVariant.PlanID != candidate.Family.PlanID {
		return false
	}
	for _, condition := range candidate.CompiledVariant.Variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == mainPredicate.ID && condition.Feature == "tense" && string(condition.Value) != "past" {
			return false
		}
	}
	return polishFrameEvidenceConsistent(source, candidate.NodeID, candidate.TransferBinding.CoreRef)
}

func polishPastLexicalAspect(candidate realizer.TargetCandidate, paradigms *morph.CompiledParadigms) (string, bool) {
	if paradigms == nil || candidate.Lexeme.ParadigmID == "" {
		return "", false
	}
	paradigm, found := paradigms.Lookup(morph.ParadigmID(candidate.Lexeme.ParadigmID))
	if !found {
		return "", false
	}
	aspect := ""
	for _, cell := range paradigm.Cells {
		var tense, cellAspect string
		for _, feature := range cell.Features {
			if feature.State != features.StateKnown {
				continue
			}
			switch feature.Feature {
			case "tense":
				tense = string(feature.Value)
			case "aspect":
				cellAspect = string(feature.Value)
			}
		}
		if tense != "past" {
			continue
		}
		if cellAspect != "imperfective" && cellAspect != "perfective" || aspect != "" && aspect != cellAspect {
			return "", false
		}
		aspect = cellAspect
	}
	return aspect, aspect != ""
}

func polishHasKnownAspectEvidence(source semantic.MeaningCandidate, concept ids.ConceptID) bool {
	for _, record := range source.Evidence.Records {
		if record.Outcome != status.OutcomeOK || record.Role != "" || record.CoreConceptID != concept {
			continue
		}
		for _, feature := range record.Features {
			if feature.Feature == "aspect" && feature.State == features.StateKnown && feature.Value != "" {
				return true
			}
		}
	}
	return false
}

func sameMeaningPath(left, right []meaning.NodeID) bool {
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

type polishSentenceOperatorBuilder struct{ negationRoutes polishNegationRouteResolver }

func (builder polishSentenceOperatorBuilder) Build(ctx context.Context, source semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
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
	scope, event, route, found := builder.negationRoutes.negatedPresent(source)
	if !found {
		return runtime.NoOperatorEvidenceBuilder{}.Build(ctx, source)
	}
	familyName := route.family
	if route.concept == "concept:seeing" {
		if shape, valid := polishAttributedSeeingShape(source); valid && shape.scope.ID == scope.ID {
			familyName = "construction:pl.attributed_seeing_object_negative"
		}
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(familyName, "construction:"))
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

type polishSentenceCandidatePolicy struct{ negationRoutes polishNegationRouteResolver }

// Independent Meaning input can omit mention context entirely. In that case
// an attribution's bare route is licensed without inventing deixis.
func polishTargetIdentificationReferenceMode(source semantic.MeaningCandidate, predicateID meaning.NodeID) (string, bool) {
	if mode, ok := polishIdentificationReferenceMode(source, predicateID); ok {
		return mode, true
	}
	if role, _, concept, ok := polishIdentificationState(source, predicateID); ok && role == "bearer" && sourceHasScope(source) {
		if _, attributed := polishAttributionForConcept(concept); attributed && len(source.Contexts) == 0 {
			return "none", true
		}
	}
	if source.Meaning == nil || len(source.Contexts) == 0 {
		return "", false
	}
	for _, item := range source.Contexts {
		if meaning.ValidateReferences(item.Meaning, item.Context, meaning.DefaultBudget()) != nil || len(item.Context.Referents) != 0 || len(item.Context.Mentions) != 0 {
			return "", false
		}
		equal, err := meaning.CanonicalEqual(item.Meaning, *source.Meaning, meaning.DefaultBudget())
		if err != nil || !equal {
			return "", false
		}
	}
	return "none", true
}

func (policy polishSentenceCandidatePolicy) KeepTargetCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	if candidate.Kind == meaning.KindEvent && candidate.TransferBinding.CoreRef.String() == "concept:giving" && candidate.TransferBinding.SenseID.String() == "sense:pl.dac.transfer" {
		predicate, exists := polishMeaningNode(source, candidate.NodeID)
		if !exists || predicate.TemporalFrame != meaning.TemporalFramePast {
			return false
		}
	}
	if candidate.Kind == meaning.KindEvent && candidate.TransferBinding.CoreRef.String() == "concept:buying" && candidate.TransferBinding.SenseID.String() == "sense:pl.kupic.buy" {
		predicate, exists := polishMeaningNode(source, candidate.NodeID)
		if !exists || predicate.TemporalFrame != meaning.TemporalFramePast {
			return false
		}
	}
	if event, exists := polishMeaningNode(source, candidate.NodeID); exists && event.Kind == meaning.KindEvent && event.TemporalFrame == meaning.TemporalFramePast &&
		(candidate.Family.Family.ID.String() == "construction:pl.intransitive" || candidate.Family.Family.ID.String() == "construction:pl.transitive" || candidate.Family.Family.ID.String() == "construction:pl.transitive_to") &&
		!polishDedicatedPastTransfer(candidate) && !polishApprovedPastTargetCandidate(source, candidate) {
		return false
	}
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
		return polishAttributedSeeingMainCandidate(source, candidate, all)
	}
	if shape, valid := polishAttributedSeeingShape(source); valid {
		if _, property := shape.properties[candidate.NodeID]; property {
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
			mode, valid := polishTargetIdentificationReferenceMode(source, candidate.NodeID)
			return valid && (mode == "proximal" || mode == "distal" || mode == "unresolved") &&
				candidate.Sense.ID.String() == "sense:pl.byc.class_membership" &&
				string(candidate.Frame.ID) == "pl.byc.class_membership.frame" &&
				candidate.Family.Family.ID.String() == "construction:pl.identification" &&
				candidate.CompiledVariant.Variant.ID.String() == "construction:pl.identification.to"
		}
		mode, valid := polishTargetIdentificationReferenceMode(source, candidate.NodeID)
		role, referentID, _, validState := polishIdentificationState(source, candidate.NodeID)
		if !valid || !validState || !polishIdentificationAttributesAllowed(source, candidate.NodeID, role, referentID, candidate.TransferBinding.CoreRef) {
			return false
		}
		switch concept {
		case "concept:class_membership":
			return mode == "proximal" || mode == "distal" || mode == "unresolved"
		case "concept:black", "concept:red", "concept:white", "concept:blue", "concept:green", "concept:open_state", "concept:sad":
			attribution, licensed := polishAttributionForConcept(concept)
			if !licensed {
				return false
			}
			gender, number, knownFeatures := polishIdentificationReferentFeatures(source, referentID, all)
			if !knownFeatures {
				return false
			}
			variant := candidate.CompiledVariant.Variant.ID.String()
			predicate := findSentencePredicate(*source.Meaning, candidate.NodeID)
			if predicate.ID == "" {
				return false
			}
			if sourceHasScope(source) {
				if mode != "none" || candidate.Family.Family.ID.String() != "construction:pl.attribution_negative" || candidate.Sense.ID.String() != attribution.sense {
					return false
				}
				tense := "present"
				if predicate.TemporalFrame == meaning.TemporalFramePast {
					tense = "past"
				} else if predicate.TemporalFrame != "" && predicate.TemporalFrame != meaning.TemporalFramePresent {
					return false
				}
				if number != "singular" && number != "plural" {
					return false
				}
				expectedVariant := "construction:pl.attribution_negative." + number + "." + tense + "." + gender
				return string(candidate.Frame.ID) == strings.TrimPrefix(attribution.sense, "sense:")+".negative.frame" && variant == expectedVariant
			}
			pastGender := ""
			if marker := strings.Index(variant, ".past."); marker >= 0 {
				pastGender = variant[marker+len(".past."):]
				variant = variant[:marker]
			}
			if (predicate.TemporalFrame == meaning.TemporalFramePast) != (pastGender != "") {
				return false
			}
			if number == "plural" {
				if pastGender != "" && pastGender != gender {
					return false
				}
				if candidate.Family.Family.ID.String() != "construction:pl.attribution_plural" || string(candidate.Frame.ID) != strings.TrimPrefix(attribution.sense, "sense:")+".plural.frame" {
					return false
				}
				if mode == "none" {
					return variant == "construction:pl.attribution_plural.bare"
				}
				if gender == "masculine_personal" {
					switch mode {
					case "proximal":
						return variant == "construction:pl.attribution_plural.ci"
					case "distal":
						return variant == "construction:pl.attribution_plural.tamci"
					default:
						return false
					}
				}
				switch mode {
				case "proximal":
					return variant == "construction:pl.attribution_plural.te"
				case "distal":
					return variant == "construction:pl.attribution_plural.tamte"
				default:
					return false
				}
			}
			if pastGender != "" && pastGender != gender {
				return false
			}
			if candidate.Family.Family.ID.String() != "construction:pl.attribution" {
				return false
			}
			if mode == "none" {
				return number == "singular" && variant == "construction:pl.attribution.bare"
			}
			switch gender {
			case "masculine_animal":
				return polishAttributionVariantMatchesDeixis(variant, "ten", "tamten", mode)
			case "masculine_inanimate":
				return polishAttributionVariantMatchesDeixis(variant, "ten.inanimate", "tamten.inanimate", mode)
			case "neuter":
				return polishAttributionVariantMatchesDeixis(variant, "to", "tamto", mode)
			case "feminine":
				return polishAttributionVariantMatchesDeixis(variant, "ta", "tamta", mode)
			default:
				return false
			}
		}
		return false
	}
	route, negativeSense := policy.negationRoutes.routeForSense(candidate.Sense.ID)
	negativeFamily := candidate.Family.Family.ID.String() == "construction:pl.transitive_negative" || candidate.Family.Family.ID.String() == "construction:pl.ditransitive_negative" || candidate.Family.Family.ID.String() == "construction:pl.oblique_negative" ||
		negativeSense && route.family == "construction:pl.intransitive" && candidate.Family.Family.ID.String() == route.family && candidate.CompiledVariant.Variant.ID.String() == route.variant ||
		negativeSense && route.derived && route.variantsByFamily[candidate.Family.Family.ID.String()][candidate.CompiledVariant.Variant.ID.String()]
	if !negativeFamily {
		if polishNegatedPredicateTarget(source, candidate.NodeID) {
			return false
		}
		return true
	}
	if route.derived && !route.variantsByFamily[candidate.Family.Family.ID.String()][candidate.CompiledVariant.Variant.ID.String()] {
		return false
	}
	_, event, sourceRoute, negated := policy.negationRoutes.negatedPresent(source)
	dynamicPast := negated && event.TemporalFrame == meaning.TemporalFramePast && (route.family == "construction:pl.intransitive" || route.derived)
	if dynamicPast && !polishApprovedPastTargetCandidate(source, candidate) {
		return false
	}
	if !negativeSense || !negated || event.ID != candidate.NodeID || sourceRoute.sense != candidate.Sense.ID.String() && !(route.concept == "concept:giving" && polishGivingNegationSenseMatches(route, candidate.Sense.ID.String())) && !(route.concept == "concept:buying" && polishBuyingNegationSenseMatches(route, candidate.Sense.ID.String())) || sourceRoute.family != route.family ||
		(route.concept == "concept:giving" && candidate.Sense.ID.String() == "sense:pl.dac.transfer" && event.TemporalFrame != meaning.TemporalFramePast) ||
		(event.TemporalFrame == meaning.TemporalFramePast && route.family != "construction:pl.intransitive" && !route.derived && !polishPastNegativeConceptAllowed(route.concept)) ||
		!polishFrameEvidenceConsistent(source, event.ID, candidate.TransferBinding.CoreRef) {
		return false
	}
	referent, exists := polishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || route.concept == "concept:class_membership" && referent.Multiplicity == meaning.MultiplicityMany {
		return false
	}
	return polishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, candidate.TransferBinding.CoreRef, true)
}

func polishAttributionVariantMatchesDeixis(variant, proximal, distal, mode string) bool {
	proximalID := "construction:pl.attribution." + proximal
	distalID := "construction:pl.attribution." + distal
	switch mode {
	case "proximal":
		return variant == proximalID
	case "distal":
		return variant == distalID
	case "demonstrative", "unresolved":
		return variant == proximalID || variant == distalID
	default:
		return false
	}
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
		case "masculine_personal":
			return gender, true
		case "masculine_animal", "masculine_inanimate":
			return "masculine_personal", true
		case "feminine", "neuter":
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

func findSentenceReferent(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id && node.Kind == meaning.KindReferent {
			return node
		}
	}
	return meaning.Node{}
}

func polishAttributedSeeingModifierReferent(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, modifierSlot construction.SlotSpec) (meaning.NodeID, string, bool) {
	if source.Meaning == nil || modifierSlot.ModifierBinding == nil || modifierSlot.ModifierBinding.SemanticRole != "bearer" {
		return "", "", false
	}
	var target construction.SlotSpec
	found := false
	for _, slot := range predicate.Family.Family.SlotSpecs {
		if slot.ID == modifierSlot.ModifierBinding.TargetSlot {
			target, found = slot, true
			break
		}
	}
	if !found || target.Category != construction.CategoryNP || target.ArgumentBinding == nil || target.Case == "" {
		return "", "", false
	}
	shape, valid := polishAttributedSeeingShape(source)
	if !valid {
		return "", "", false
	}
	switch target.ArgumentBinding.SemanticRole {
	case "experiencer":
		return shape.subject.ID, target.Case, true
	case "stimulus":
		return shape.object.ID, target.Case, true
	default:
		return "", "", false
	}
}

func knownPolishFeatures(values ...[2]string) []features.FeatureValue {
	result := make([]features.FeatureValue, 0, len(values))
	for _, value := range values {
		result = append(result, features.FeatureValue{Feature: features.FeatureID(value[0]), State: features.StateKnown, Value: features.FeatureValueID(value[1])})
	}
	return result
}

var _ runtime.SlotFeaturePlanner = polishSentenceFeatures{}

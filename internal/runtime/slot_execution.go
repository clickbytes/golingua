package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// SlotFeaturePlanner supplies language-local morphology values for one exact
// selected lexical candidate and family slot. It must return only feature
// values licensed by the compiled language profile and its own policy.
type SlotFeaturePlanner interface {
	Features(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate, construction.SlotSpec, realizer.TargetCandidate, []realizer.TargetCandidate) ([]features.FeatureValue, error)
}

// VPAuxiliaryExecutionPlanner derives an auxiliary request only after the
// exact selected predicate, family variant, and event Meaning are available.
type VPAuxiliaryExecutionPlanner interface {
	BuildVPAuxiliary(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate, construction.SlotSpec, realizer.TargetCandidate, []realizer.TargetCandidate) (*realizer.VPAuxiliaryExecutionRequest, error)
}

// MorphologyExecutionBuilder binds exact selector candidates to the
// role-bearing frame slots, then requests morphology from the compiled seed.
// It never selects among competing participant senses or invents role links.
type MorphologyExecutionBuilder struct {
	Profile       *features.LanguageFeatureProfile
	PhraseProfile construction.PhraseProfile
	Paradigms     *morph.CompiledParadigms
	Exceptions    *morph.CompiledExceptions
	Parts         func(string) (morph.PrincipalParts, bool)
	FeaturePlan   SlotFeaturePlanner
	CaseFeature   features.FeatureID
	// VPAuxiliary is an opt-in language-local request copied only to the exact
	// main predicate slot. The realizer revalidates its event and operator scope.
	VPAuxiliary *realizer.VPAuxiliaryExecutionRequest
}

var _ TargetExecutionBuilder = MorphologyExecutionBuilder{}

func (builder MorphologyExecutionBuilder) Build(ctx context.Context, source semantic.MeaningCandidate, candidates []realizer.TargetCandidate, selected realizer.TargetCandidate, plan realizer.RealizationPlan) (realizer.ExecuteInput, error) {
	if ctx == nil {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "slot execution requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
	}
	if source.Meaning == nil || builder.Profile == nil || builder.Paradigms == nil || builder.Parts == nil || builder.FeaturePlan == nil {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "SLOT_EXECUTION_UNAVAILABLE", "target.slot_execution", "slot execution requires a Meaning graph, compiled morphology and explicit language feature planner")
	}
	if builder.PhraseProfile.Language != builder.Profile.Language() || builder.PhraseProfile.Version == "" {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "NP_PHRASE_PROFILE_UNAVAILABLE", "target.np.profile", "target NP execution requires the exact language-local phrase profile")
	}
	if plan.Language != builder.Profile.Language() || string(selected.Lexeme.Language) != string(builder.Profile.Language()) || selected.NodeID != plan.NodeID {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "SLOT_EXECUTION_IDENTITY_MISMATCH", "target.slot_execution", "selected predicate, realization plan and morphology profile must identify the same target language and Meaning node")
	}
	predicate, found := meaningNode(*source.Meaning, plan.NodeID)
	if !found {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "PREDICATE_NODE_UNAVAILABLE", "source.meaning", "selected plan must identify an event or state node in the source Meaning")
	}
	if predicate.Kind == meaning.KindReferent {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "STANDALONE_NP_REALIZATION_UNAVAILABLE", "target.np", "this runtime has no registered standalone noun phrase construction")
	}
	if predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "PREDICATE_NODE_UNAVAILABLE", "source.meaning", "selected plan must identify an event or state node in the source Meaning")
	}
	input := realizer.ExecuteInput{
		Plan: plan, SourceMeaning: *source.Meaning,
		Family: selected.Family.Family, Frame: selected.Frame,
		FeatureProfile: builder.Profile, PhraseProfile: builder.PhraseProfile,
		Paradigms: builder.Paradigms, Exceptions: builder.Exceptions,
		Slots:                make(map[string]realizer.SlotRequest, len(selected.Family.Family.SlotSpecs)),
		InformationStructure: map[meaning.Role]construction.InformationValue{},
	}
	input.SourceContexts = make([]meaning.RequestContext, 0, len(source.Contexts))
	for _, candidateContext := range source.Contexts {
		input.SourceContexts = append(input.SourceContexts, candidateContext.Context)
	}
	for _, slot := range selected.Family.Family.SlotSpecs {
		if err := ctx.Err(); err != nil {
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
		var candidate realizer.TargetCandidate
		var nodeID meaning.NodeID
		if slot.ModifierBinding != nil {
			modifier, outcome := realizer.FindModifierOutcome(*source.Meaning, plan.NodeID, selected.Family.Family, slot.ID)
			if outcome != status.OutcomeOK {
				code := "MODIFIER_STATE_UNAVAILABLE"
				if outcome == status.OutcomeAmbiguous {
					code = "MODIFIER_STATE_AMBIGUOUS"
				}
				return realizer.ExecuteInput{}, runtimeError(outcome, code, "source.meaning.modifiers."+slot.ID, "declared modifier slot does not have exactly one state linked to its NP referent")
			}
			choices := make(map[string]realizer.TargetCandidate)
			for _, available := range candidates {
				if available.SourceCandidateID != source.ID || available.SemanticID != source.SemanticID ||
					available.NodeID != modifier.ID || available.Kind != meaning.KindState || available.Outcome != status.OutcomeOK ||
					available.TransferBinding.CoreRef.String() != string(modifier.Concept) || available.Lexeme.POS != lexicon.PartOfSpeech("adjective") {
					continue
				}
				key := available.Lexeme.ID.String() + "\x00" + available.Sense.ID.String() + "\x00" + available.TransferBinding.CoreRef.String()
				if _, exists := choices[key]; !exists {
					choices[key] = available
				}
			}
			if len(choices) == 0 {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "MODIFIER_LEXICALIZATION_UNAVAILABLE", "target.slots."+slot.ID, "no licensed adjective candidate is available for the exact modifier state")
			}
			if len(choices) > 1 {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeAmbiguous, "MODIFIER_LEXICALIZATION_AMBIGUOUS", "target.slots."+slot.ID, "multiple lexical identities remain for the exact modifier state")
			}
			for _, selectedModifier := range choices {
				candidate = selectedModifier
			}
			nodeID = modifier.ID
		} else if slot.IsMainPredicate() {
			candidate = selected
			nodeID = plan.NodeID
		} else if slot.ArgumentBinding != nil {
			participant, count := participantForRole(predicate, slot.ArgumentBinding.SemanticRole)
			if count == 0 {
				if slot.Optionality == construction.SlotOptional && slot.Cardinality.Minimum == 0 {
					continue
				}
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeIncomplete, "MEANING_ROLE_UNAVAILABLE", "source.meaning."+string(plan.NodeID)+".participants", "required construction role is absent from source Meaning")
			}
			if count > 1 {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeAmbiguous, "MEANING_ROLE_AMBIGUOUS", "source.meaning."+string(plan.NodeID)+".participants", "source Meaning has multiple targets for one construction role")
			}
			nodeID = participant.Target
			if _, licensed := construction.ImplicitParticipantSlot(selected.Family.Family, plan.CompiledVariant.Variant, slot.ID); licensed {
				for _, paired := range source.Contexts {
					same, err := meaning.StructuralEquivalence(*source.Meaning, paired.Meaning, meaning.DefaultBudget())
					if err != nil || !same {
						return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "IMPLICIT_PARTICIPANT_CONTEXT_MISMATCH", "source.contexts."+string(nodeID), "an omitted participant requires contexts paired with the selected Meaning graph")
					}
				}
				if _, anchored := realizer.ImplicitParticipantAnchor(*source.Meaning, input.SourceContexts, nodeID); anchored {
					continue
				}
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "IMPLICIT_PARTICIPANT_CONTEXT_UNAVAILABLE", "source.contexts."+string(nodeID), "an omitted optional subject role requires a person referent with one consistent participant anchor")
			}
			choices := make([]realizer.TargetCandidate, 0, 2)
			for _, available := range candidates {
				if available.SourceCandidateID == source.ID && available.SemanticID == source.SemanticID &&
					available.NodeID == nodeID && available.Kind == meaning.KindReferent && available.Outcome == status.OutcomeOK {
					choices = append(choices, available)
				}
			}
			if len(choices) == 0 {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "ROLE_LEXICALIZATION_UNAVAILABLE", "target.slots."+slot.ID, "no licensed target candidate is available for this exact semantic role referent")
			}
			if len(choices) > 1 {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeAmbiguous, "ROLE_LEXICALIZATION_AMBIGUOUS", "target.slots."+slot.ID, "multiple target senses remain for this exact semantic role referent")
			}
			candidate = choices[0]
		} else {
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "UNBOUND_SLOT_UNSUPPORTED", "target.slots."+slot.ID, "only the main predicate, declared modifiers, and explicitly role-bound argument slots are supported")
		}
		if slot.Category != construction.CategoryNP && !slot.Category.IsPredicate() && slot.Category != construction.CategoryPP {
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "SLOT_CATEGORY_UNSUPPORTED", "target.slots."+slot.ID, "slot execution does not yet realize PP slots")
		}
		if slot.Category == construction.CategoryPP || slot.Adposition != "" {
			if slot.Adposition == "" || !construction.VariantEmitsAdposition(plan.CompiledVariant.Variant, selected.Family.Family, slot.ID, slot.Adposition) {
				return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "ADPOSITION_SEQUENCE_UNAVAILABLE", "target.slots."+slot.ID, "an adpositional argument requires its exact adposition terminal before the slot in the selected construction")
			}
		}
		parts, exists := builder.Parts(candidate.Lexeme.ID.String())
		if !exists {
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeIncomplete, "PRINCIPAL_PARTS_UNAVAILABLE", "target.slots."+slot.ID, "selected target lexeme has no attested principal parts")
		}
		values, err := builder.FeaturePlan.Features(ctx, source, selected, slot, candidate, candidates)
		if err != nil {
			return realizer.ExecuteInput{}, err
		}
		bundle, err := features.NewBundle(builder.Profile, values)
		if err != nil {
			return realizer.ExecuteInput{}, fmt.Errorf("target slot %s feature bundle: %w", slot.ID, err)
		}
		request := realizer.SlotRequest{
			NodeID: nodeID, Candidate: candidate, PrincipalParts: parts,
			Features: bundle, CaseFeature: builder.CaseFeature,
		}
		if builder.VPAuxiliary != nil {
			if slot.IsMainPredicate() && slot.Category == construction.CategoryVP {
				if builder.VPAuxiliary.EventNodeID != nodeID {
					return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "VP_AUXILIARY_EVENT_MISMATCH", "target.slot_execution.vp_auxiliary", "opt-in VP auxiliary request must target the exact selected main predicate")
				}
				auxiliary := *builder.VPAuxiliary
				request.VPAuxiliary = &auxiliary
			}
		}
		if slot.IsMainPredicate() && slot.Category == construction.CategoryVP {
			if planner, ok := builder.FeaturePlan.(VPAuxiliaryExecutionPlanner); ok {
				auxiliary, planErr := planner.BuildVPAuxiliary(ctx, source, selected, slot, candidate, candidates)
				if planErr != nil {
					return realizer.ExecuteInput{}, planErr
				}
				if auxiliary != nil {
					if request.VPAuxiliary != nil || auxiliary.EventNodeID != nodeID {
						return realizer.ExecuteInput{}, runtimeError(status.OutcomeInvalid, "VP_AUXILIARY_EVENT_MISMATCH", "target.slot_execution.vp_auxiliary", "planned auxiliary must target the exact main predicate and cannot conflict with a static request")
					}
					if parts, ok := builder.Parts(auxiliary.Lexeme.ID.String()); !ok {
						return realizer.ExecuteInput{}, runtimeError(status.OutcomeIncomplete, "VP_AUXILIARY_PRINCIPAL_PARTS_UNAVAILABLE", "target.slot_execution.vp_auxiliary", "the selected auxiliary lexeme has no principal parts")
					} else {
						auxiliary.PrincipalParts = parts
					}
					request.VPAuxiliary = auxiliary
				}
			}
		}
		input.Slots[slot.ID] = request
	}
	if err := attachNPExecutionRequests(input.Slots, source, selected, plan, builder.Profile, builder.PhraseProfile); err != nil {
		return realizer.ExecuteInput{}, err
	}
	return input, nil
}

func attachNPExecutionRequests(slots map[string]realizer.SlotRequest, source semantic.MeaningCandidate, selected realizer.TargetCandidate, realizationPlan realizer.RealizationPlan, profile *features.LanguageFeatureProfile, phraseProfile construction.PhraseProfile) error {
	if profile == nil || phraseProfile.Language != profile.Language() {
		return runtimeError(status.OutcomeUnsupported, "NP_PHRASE_PROFILE_UNAVAILABLE", "target.np.profile", "NP plan compilation requires matching language phrase and feature profiles")
	}
	template, ok := runtimeNPTemplate(phraseProfile, "np-attributive")
	if !ok {
		return runtimeError(status.OutcomeUnknown, "ATTRIBUTIVE_NP_TEMPLATE_UNAVAILABLE", "target.np.template", "language phrase profile has no compiled np-attributive template")
	}
	for _, headSpec := range selected.Family.Family.SlotSpecs {
		if headSpec.Category != construction.CategoryNP {
			continue
		}
		head, exists := slots[headSpec.ID]
		if !exists {
			continue
		}
		var modifier *realizer.SlotRequest
		modifierSlotID := ""
		for _, spec := range selected.Family.Family.SlotSpecs {
			if spec.ModifierBinding != nil && spec.ModifierBinding.TargetSlot == headSpec.ID {
				if modifierSlotID != "" {
					return runtimeError(status.OutcomeAmbiguous, "NP_MODIFIER_SLOT_AMBIGUOUS", "target.np."+headSpec.ID, "multiple modifier slots target one NP head")
				}
				modifierSlotID = spec.ID
			}
		}
		if modifierSlotID != "" {
			value, found := slots[modifierSlotID]
			if !found {
				return runtimeError(status.OutcomeIncomplete, "NP_MODIFIER_REQUEST_MISSING", "target.np."+headSpec.ID+".modifier", "the licensed modifier slot has no exact execution request")
			}
			modifier = &value
		}
		var determiner *realizer.NPDeterminerExecutionRequest
		sequence := selected.CompiledVariant.Variant.Sequence
		firstSlot := headSpec.ID
		headIndex := variantSlotSequenceIndex(sequence, headSpec.ID)
		if headIndex < 0 {
			return runtimeError(status.OutcomeIncomplete, "NP_HEAD_OUTER_SLOT_UNAVAILABLE", "target.np."+headSpec.ID, "selected NP head does not have one exact outer sequence position")
		}
		modifierIndex := -1
		if modifierSlotID != "" {
			modifierIndex = variantSlotSequenceIndex(sequence, modifierSlotID)
			if modifierIndex < 0 {
				return runtimeError(status.OutcomeIncomplete, "NP_MODIFIER_OUTER_SLOT_UNAVAILABLE", "target.np."+headSpec.ID+".modifier", "selected NP modifier does not have one exact outer sequence position")
			}
			firstSlot = modifierSlotID
		}
		firstIndex := -1
		for i, token := range sequence {
			if token.Kind == construction.TokenSlot && token.SlotID == firstSlot {
				firstIndex = i
				break
			}
		}
		if firstIndex > 0 {
			determinerIndex := firstIndex - 1
			if modifierIndex >= 0 && headIndex > modifierIndex {
				determinerIndex = modifierIndex - 1
			}
			previous := sequence[determinerIndex]
			if previous.Kind == construction.TokenTerminal && hasNPDeterminerRule(phraseProfile, previous.Terminal) {
				firstPronounced := head
				if modifier != nil && modifierIndex >= 0 && modifierIndex < headIndex {
					firstPronounced = *modifier
				}
				selectedDeterminer, err := targetNPDeterminer(profile, phraseProfile, template, previous.Terminal, head.Features, firstPronounced.Candidate.Lexeme.Properties)
				if err != nil {
					return err
				}
				selectedDeterminer.InputID = "determiner-input"
				selectedDeterminer.VariantID = selected.CompiledVariant.Variant.ID
				selectedDeterminer.SequenceIndex = determinerIndex
				determiner = selectedDeterminer
			}
		}
		if determiner == nil {
			firstModifier := modifier != nil && modifierIndex >= 0 && modifierIndex < headIndex
			role := meaning.Role("")
			if headSpec.ArgumentBinding != nil {
				role = headSpec.ArgumentBinding.SemanticRole
			}
			selectedDeterminer, err := targetNPEmbeddedDeterminer(profile, phraseProfile, template, realizationPlan.Assumptions, source, role, head, modifier, firstModifier)
			if err != nil {
				return err
			}
			if selectedDeterminer != nil {
				selectedDeterminer.VariantID = selected.CompiledVariant.Variant.ID
			}
			determiner = selectedDeterminer
		}
		phrasePlan, headChild, modifierChild, err := compileTargetNP(profile, phraseProfile, template, head, modifier, determiner, selected)
		if err != nil {
			return err
		}
		headChild.Assumptions = assumptionsForNode(realizationPlan.Assumptions, head.NodeID)
		if modifierChild != nil {
			modifierChild.Assumptions = assumptionsForNode(realizationPlan.Assumptions, modifierChild.NodeID)
		}
		request := &realizer.NPExecutionRequest{
			Plan: phrasePlan, ProfileVersion: phraseProfile.Version,
			VariantID: selected.CompiledVariant.Variant.ID, HeadSlotID: headSpec.ID,
			HeadSequenceIndex: headIndex, Head: headChild, ModifierSlotID: modifierSlotID,
			ModifierSequenceIndex: modifierIndex, Modifier: modifierChild, Determiner: determiner,
			Bare: determiner == nil,
		}
		head.NP = request
		slots[headSpec.ID] = head
	}
	return nil
}

// targetNPEmbeddedDeterminer represents a caller-selected determiner as an NP
// part when the selected outer variant has no determiner terminal. It only
// acts on explicit determiner assumptions and obtains surfaces/forms from the
// language phrase profile.
func targetNPEmbeddedDeterminer(profile *features.LanguageFeatureProfile, phraseProfile construction.PhraseProfile, template construction.PhraseTemplate, assumptions []realizer.FieldAssumption, source semantic.MeaningCandidate, role meaning.Role, head realizer.SlotRequest, modifier *realizer.SlotRequest, modifierFirst bool) (*realizer.NPDeterminerExecutionRequest, error) {
	value := ""
	bare := false
	for _, assumption := range assumptions {
		if assumption.Field != "determiner" || len(assumption.Path) == 0 || assumption.Path[len(assumption.Path)-1] != head.NodeID {
			continue
		}
		if assumption.Value == "bare" {
			bare = true
			continue
		}
		if value != "" && value != assumption.Value {
			return nil, runtimeError(status.OutcomeAmbiguous, "NP_EMBEDDED_DETERMINER_ASSUMPTION_CONFLICT", "target.np.determiner", "multiple determiner assumptions select different phrase-profile values")
		}
		value = assumption.Value
	}
	if bare && value != "" {
		return nil, runtimeError(status.OutcomeAmbiguous, "NP_EMBEDDED_DETERMINER_ASSUMPTION_CONFLICT", "target.np.determiner", "bare and overt determiner assumptions conflict")
	}
	if bare {
		return nil, nil
	}
	sourceDeterminer := ""
	if values, found := SourceFeatures(source, role, head.Candidate.TransferBinding.CoreRef); found {
		sourceDeterminer = string(values["determiner"])
	}
	if sourceDeterminer != "" {
		if sourceDeterminer == "bare" {
			if value != "" {
				return nil, runtimeError(status.OutcomeUnsupported, "NP_EMBEDDED_DETERMINER_SOURCE_MISMATCH", "target.np.determiner", "source zero determiner conflicts with an overt target determiner choice")
			}
			return nil, nil
		}
		if sourceDeterminer != "definite" && sourceDeterminer != "indefinite" {
			return nil, runtimeError(status.OutcomeUnsupported, "NP_EMBEDDED_SOURCE_DETERMINER_UNSUPPORTED", "target.np.determiner", "source determiner cannot be represented by this embedded determiner choice")
		}
		if value != "" && value != sourceDeterminer {
			return nil, runtimeError(status.OutcomeUnsupported, "NP_EMBEDDED_DETERMINER_SOURCE_MISMATCH", "target.np.determiner", "target determiner assumption conflicts with explicit source determiner evidence")
		}
		value = sourceDeterminer
	}
	if value == "" {
		return nil, nil
	}
	profileHasChoice := false
	for _, rule := range phraseProfile.DeterminerForms {
		for _, feature := range rule.DeterminerFeatures {
			if feature.Feature == "determiner" && string(feature.Value) == value {
				profileHasChoice = true
				break
			}
		}
		if profileHasChoice {
			break
		}
	}
	if !profileHasChoice {
		return nil, nil
	}
	headBundle, err := projectNPBundle(profile, template, "head", head.Features)
	if err != nil {
		return nil, err
	}
	if profile.Language() == "en" && !containsBundleFeature(headBundle, "number") {
		if nounNumber, known := knownNPFeature(headBundle, "noun_number"); known {
			values := append(append([]features.FeatureValue(nil), headBundle.Values...), features.FeatureValue{Feature: "number", State: features.StateKnown, Value: nounNumber})
			headBundle, err = features.NewBundle(profile, values)
			if err != nil {
				return nil, runtimeError(status.OutcomeIncomplete, "NP_HEAD_NUMBER_FEATURE_INVALID", "target.np.head.number", err.Error())
			}
		}
	}
	firstChild := head
	if modifier != nil && modifierFirst {
		firstChild = *modifier
	}
	onset, known := lexicalDeterminerOnset(firstChild.Candidate.Lexeme.Properties)
	if !known {
		onset = ""
	}
	unique := map[string]realizer.NPDeterminerExecutionRequest{}
	for _, rule := range phraseProfile.DeterminerForms {
		matchesValue := false
		for _, feature := range rule.DeterminerFeatures {
			if feature.Feature == "determiner" && string(feature.Value) == value {
				matchesValue = true
				break
			}
		}
		if !matchesValue {
			continue
		}
		bundle, bundleErr := features.NewBundle(profile, rule.DeterminerFeatures)
		if bundleErr != nil {
			return nil, runtimeError(status.OutcomeIncomplete, "NP_DETERMINER_FEATURE_INVALID", "target.np.determiner", bundleErr.Error())
		}
		match := construction.EvaluateDeterminerForms(phraseProfile, rule.Surface, headBundle, bundle, onset)
		if match.Status != construction.DeterminerOK {
			continue
		}
		key := rule.Surface + "\x00" + fmt.Sprint(rule.DeterminerFeatures)
		unique[key] = realizer.NPDeterminerExecutionRequest{InputID: "determiner-input", Surface: rule.Surface, Embedded: true, SequenceIndex: -1, Features: bundle}
	}
	if len(unique) > 1 {
		return nil, runtimeError(status.OutcomeAmbiguous, "NP_EMBEDDED_DETERMINER_AMBIGUOUS", "target.np.determiner", "multiple phrase-profile forms satisfy the explicit determiner assumption")
	}
	for _, request := range unique {
		return &request, nil
	}
	return nil, runtimeError(status.OutcomeIncomplete, "NP_EMBEDDED_DETERMINER_UNAVAILABLE", "target.np.determiner", "explicit determiner assumption has no matching phrase-profile form for this head")
}

func variantSlotSequenceIndex(sequence []construction.SequenceToken, slotID string) int {
	found := -1
	for i, token := range sequence {
		if token.Kind == construction.TokenSlot && token.SlotID == slotID {
			if found >= 0 {
				return -1
			}
			found = i
		}
	}
	return found
}

func assumptionsForNode(assumptions []realizer.FieldAssumption, nodeID meaning.NodeID) []realizer.FieldAssumption {
	var selected []realizer.FieldAssumption
	for _, assumption := range assumptions {
		if len(assumption.Path) > 0 && assumption.Path[len(assumption.Path)-1] == nodeID {
			copy := assumption
			copy.Path = append([]meaning.NodeID(nil), assumption.Path...)
			selected = append(selected, copy)
		}
	}
	return selected
}

func runtimeNPTemplate(profile construction.PhraseProfile, id string) (construction.PhraseTemplate, bool) {
	for _, template := range profile.Templates {
		if template.ID == id && template.Kind == construction.PhraseNP {
			return template, true
		}
	}
	return construction.PhraseTemplate{}, false
}

func compileTargetNP(profile *features.LanguageFeatureProfile, phraseProfile construction.PhraseProfile, template construction.PhraseTemplate, head realizer.SlotRequest, modifier *realizer.SlotRequest, determiner *realizer.NPDeterminerExecutionRequest, selected realizer.TargetCandidate) (construction.CompiledPhrasePlan, realizer.NPChildExecutionRequest, *realizer.NPChildExecutionRequest, error) {
	headBundle, err := projectNPBundle(profile, template, "head", head.Features)
	if err != nil {
		return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, err
	}
	if requiredNPFeature(profile.Language(), headBundle, "number") == "" {
		return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeIncomplete, "NP_HEAD_NUMBER_UNKNOWN", "target.np.head.number", "an NP plan requires explicit number evidence for its selected head")
	}
	headChild := realizer.NPChildExecutionRequest{InputID: "head-input", Role: "head", NodeID: head.NodeID, Candidate: head.Candidate, PrincipalParts: head.PrincipalParts, Features: head.Features, CaseFeature: head.CaseFeature}
	inputs := []construction.PhraseInput{{ID: "head-input", Role: "head", Kind: construction.PhraseInputLexical, Required: true, Features: headBundle}}
	var modifierChild *realizer.NPChildExecutionRequest
	agreements := make([]construction.FeatureAgreement, 0, 4)
	if modifier != nil {
		modifierBundle, bundleErr := projectNPBundle(profile, template, "modifier", modifier.Features)
		if bundleErr != nil {
			return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, bundleErr
		}
		for _, declaration := range template.Agreements {
			if declaration.ControllerRole != "head" || declaration.TargetRole != "modifier" {
				continue
			}
			id := declaration.Feature
			headValue, headKnown := knownNPFeature(headBundle, id)
			modifierValue, modifierKnown := knownNPFeature(modifierBundle, id)
			if !headKnown || !modifierKnown {
				return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeIncomplete, "NP_AGREEMENT_FEATURE_UNKNOWN", "target.np.modifier."+string(id), "selected head and modifier lack known shared agreement evidence")
			}
			if headValue != modifierValue {
				return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeUnsupported, "NP_AGREEMENT_FEATURE_MISMATCH", "target.np.modifier."+string(id), "selected head and modifier have conflicting agreement evidence")
			}
			agreements = append(agreements, construction.FeatureAgreement{ControllerInput: "head-input", TargetInput: "modifier-input", Feature: id})
		}
		modifierChild = &realizer.NPChildExecutionRequest{InputID: "modifier-input", Role: "modifier", NodeID: modifier.NodeID, Candidate: modifier.Candidate, PrincipalParts: modifier.PrincipalParts, Features: modifier.Features, CaseFeature: modifier.CaseFeature}
		inputs = append(inputs, construction.PhraseInput{ID: "modifier-input", Role: "modifier", Kind: construction.PhraseInputAP, Required: false, Features: modifierBundle})
	}
	if determiner != nil {
		inputs = append(inputs, construction.PhraseInput{ID: "determiner-input", Role: "determiner", Kind: construction.PhraseInputLexical, Required: false, Features: determiner.Features})
		for _, declaration := range template.Agreements {
			if declaration.ControllerRole != "head" || declaration.TargetRole != "determiner" {
				continue
			}
			id := declaration.Feature
			headValue, headKnown := knownNPFeature(headBundle, id)
			determinerValue, determinerKnown := knownNPFeature(determiner.Features, id)
			if !headKnown || !determinerKnown {
				return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeIncomplete, "NP_DETERMINER_AGREEMENT_UNKNOWN", "target.np.determiner."+string(id), "selected head and determiner lack known shared agreement evidence")
			}
			if headValue != determinerValue {
				return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeUnsupported, "NP_DETERMINER_AGREEMENT_MISMATCH", "target.np.determiner."+string(id), "selected head and determiner have conflicting agreement evidence")
			}
			agreements = append(agreements, construction.FeatureAgreement{ControllerInput: "head-input", TargetInput: "determiner-input", Feature: id})
		}
	}
	order := make([]string, 0, len(inputs))
	present := map[string]bool{}
	for _, input := range inputs {
		present[input.Role] = true
	}
	for _, role := range template.Order {
		if present[role] {
			order = append(order, role)
		}
	}
	if len(order) != len(inputs) {
		return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeInvalid, "NP_TEMPLATE_ORDER_INVALID", "target.np.order", "selected NP inputs cannot be placed by the language template")
	}
	seed := string(profile.Language()) + "\x00" + string(head.NodeID) + "\x00" + head.Candidate.Lexeme.ID.String() + "\x00" + selected.CompiledVariant.Variant.ID.String()
	if modifier != nil {
		seed += "\x00" + string(modifier.NodeID) + "\x00" + modifier.Candidate.Lexeme.ID.String()
	}
	if determiner != nil {
		seed += "\x00" + determiner.Surface + "\x00" + fmt.Sprint(determiner.SequenceIndex) + "\x00" + fmt.Sprint(determiner.Embedded)
	}
	digest := sha256.Sum256([]byte(seed))
	phrasePlan := construction.PhrasePlan{ID: construction.PhraseID("target-np-" + hex.EncodeToString(digest[:])), TemplateID: template.ID, Language: profile.Language(), Kind: construction.PhraseNP, Inputs: inputs, Order: order, Agreements: agreements}
	compiled, err := construction.CompileNP(phrasePlan, phraseProfile, profile)
	if err != nil {
		return construction.CompiledPhrasePlan{}, realizer.NPChildExecutionRequest{}, nil, runtimeError(status.OutcomeIncomplete, "NP_PLAN_COMPILATION_FAILED", "target.np.plan", err.Error())
	}
	return compiled, headChild, modifierChild, nil
}

func projectNPBundle(profile *features.LanguageFeatureProfile, template construction.PhraseTemplate, role string, source features.FeatureBundle) (features.FeatureBundle, error) {
	allowed := map[features.FeatureID]bool{}
	for _, spec := range template.Inputs {
		if spec.Role == role {
			for _, id := range spec.FeatureIDs {
				allowed[id] = true
			}
		}
	}
	values := make([]features.FeatureValue, 0, len(source.Values))
	for _, value := range source.Values {
		if allowed[value.Feature] {
			values = append(values, value)
		}
	}
	bundle, err := features.NewBundle(profile, values)
	if err != nil {
		return features.FeatureBundle{}, runtimeError(status.OutcomeIncomplete, "NP_FEATURE_PROJECTION_FAILED", "target.np."+role, err.Error())
	}
	return bundle, nil
}

func containsNPFeature(template construction.PhraseTemplate, role string, id features.FeatureID) bool {
	for _, input := range template.Inputs {
		if input.Role == role {
			for _, candidate := range input.FeatureIDs {
				if candidate == id {
					return true
				}
			}
		}
	}
	return false
}

func knownNPFeature(bundle features.FeatureBundle, id features.FeatureID) (features.FeatureValueID, bool) {
	for _, value := range bundle.Values {
		if value.Feature == id && value.State == features.StateKnown {
			return value.Value, true
		}
	}
	return "", false
}

func knownNPBundleFeature(left, right features.FeatureBundle, id features.FeatureID) bool {
	_, a := knownNPFeature(left, id)
	_, b := knownNPFeature(right, id)
	return a && b
}

func requiredNPFeature(language features.LanguageID, bundle features.FeatureBundle, id features.FeatureID) features.FeatureValueID {
	if language == "en" && id == "number" {
		if number, ok := knownNPFeature(bundle, "noun_number"); ok {
			return number
		}
	}
	value, _ := knownNPFeature(bundle, features.FeatureID(id))
	return value
}

func hasNPDeterminerRule(profile construction.PhraseProfile, surface string) bool {
	for _, rule := range profile.DeterminerForms {
		if rule.Surface == surface {
			return true
		}
	}
	return false
}

func targetNPDeterminer(profile *features.LanguageFeatureProfile, phraseProfile construction.PhraseProfile, template construction.PhraseTemplate, surface string, head features.FeatureBundle, firstLexemeProperties []lexicon.Property) (*realizer.NPDeterminerExecutionRequest, error) {
	headBundle, err := projectNPBundle(profile, template, "head", head)
	if err != nil {
		return nil, err
	}
	if profile.Language() == "en" {
		if !containsBundleFeature(headBundle, "number") {
			if nounNumber, known := knownNPFeature(headBundle, "noun_number"); known {
				values := append(append([]features.FeatureValue(nil), headBundle.Values...), features.FeatureValue{Feature: "number", State: features.StateKnown, Value: nounNumber})
				headBundle, err = features.NewBundle(profile, values)
				if err != nil {
					return nil, runtimeError(status.OutcomeIncomplete, "NP_HEAD_NUMBER_FEATURE_INVALID", "target.np.head.number", err.Error())
				}
			}
		}
	}
	onset, onsetKnown := lexicalDeterminerOnset(firstLexemeProperties)
	if !onsetKnown {
		onset = ""
	}
	var incomplete bool
	var ambiguous bool
	var ambiguousRules []construction.DeterminerFormRule
	var candidates []construction.DeterminerFormRule
	for _, rule := range phraseProfile.DeterminerForms {
		if rule.Surface != surface {
			continue
		}
		determinerBundle, bundleErr := features.NewBundle(profile, rule.DeterminerFeatures)
		if bundleErr != nil {
			return nil, runtimeError(status.OutcomeIncomplete, "NP_DETERMINER_FEATURE_INVALID", "target.np.determiner", bundleErr.Error())
		}
		match := construction.EvaluateDeterminerForms(phraseProfile, surface, headBundle, determinerBundle, onset)
		switch match.Status {
		case construction.DeterminerOK:
			candidates = append(candidates, match.Candidates...)
		case construction.DeterminerAmbiguous:
			ambiguous = true
			ambiguousRules = append(ambiguousRules, match.Candidates...)
		case construction.DeterminerIncomplete:
			incomplete = true
		}
	}
	if ambiguous {
		return nil, runtimeError(status.OutcomeAmbiguous, "NP_DETERMINER_FORM_AMBIGUOUS", "target.np.determiner", fmt.Sprintf("%d non-identical exact determiner rules remain", len(ambiguousRules)))
	}
	if len(candidates) == 0 {
		if incomplete {
			return nil, runtimeError(status.OutcomeIncomplete, "NP_DETERMINER_EVIDENCE_INCOMPLETE", "target.np.determiner", "determiner form requires explicit head or onset evidence")
		}
		return nil, runtimeError(status.OutcomeUnsupported, "NP_DETERMINER_FORM_UNSUPPORTED", "target.np.determiner", "no exact phrase-profile rule licenses this determiner for the selected head")
	}
	unique := map[string]construction.DeterminerFormRule{}
	for _, rule := range candidates {
		key := rule.Surface + "\x00" + rule.Onset + "\x00" + fmt.Sprint(rule.HeadFeatures) + "\x00" + fmt.Sprint(rule.DeterminerFeatures)
		unique[key] = rule
	}
	if len(unique) > 1 {
		return nil, runtimeError(status.OutcomeAmbiguous, "NP_DETERMINER_FORM_AMBIGUOUS", "target.np.determiner", "multiple non-identical exact determiner rules remain")
	}
	var rule construction.DeterminerFormRule
	for _, candidate := range unique {
		rule = candidate
	}
	bundle, err := features.NewBundle(profile, rule.DeterminerFeatures)
	if err != nil {
		return nil, runtimeError(status.OutcomeIncomplete, "NP_DETERMINER_FEATURE_INVALID", "target.np.determiner", err.Error())
	}
	return &realizer.NPDeterminerExecutionRequest{Surface: surface, Features: bundle}, nil
}

func containsBundleFeature(bundle features.FeatureBundle, id features.FeatureID) bool {
	for _, value := range bundle.Values {
		if value.Feature == id {
			return true
		}
	}
	return false
}

func lexicalDeterminerOnset(properties []lexicon.Property) (string, bool) {
	article := ""
	for _, property := range properties {
		if property.Key != "indefinite_article" || (property.Value != "a" && property.Value != "an") {
			continue
		}
		if article != "" && article != property.Value {
			return "", false
		}
		article = property.Value
	}
	switch article {
	case "a":
		return "consonant", true
	case "an":
		return "vowel", true
	default:
		return "", false
	}
}

func meaningNode(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func participantForRole(predicate meaning.Node, role meaning.Role) (meaning.Participant, int) {
	var selected meaning.Participant
	count := 0
	for _, participant := range predicate.Participants {
		if participant.Role == role {
			selected = participant
			count++
		}
	}
	return selected, count
}

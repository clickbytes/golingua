package runtime

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
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

// MorphologyExecutionBuilder binds exact selector candidates to the
// role-bearing frame slots, then requests morphology from the compiled seed.
// It never selects among competing participant senses or invents role links.
type MorphologyExecutionBuilder struct {
	Profile     *features.LanguageFeatureProfile
	Paradigms   *morph.CompiledParadigms
	Exceptions  *morph.CompiledExceptions
	Parts       func(string) (morph.PrincipalParts, bool)
	FeaturePlan SlotFeaturePlanner
	CaseFeature features.FeatureID
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
		Paradigms: builder.Paradigms, Exceptions: builder.Exceptions,
		Slots:                make(map[string]realizer.SlotRequest, len(selected.Family.Family.SlotSpecs)),
		InformationStructure: map[meaning.Role]construction.InformationValue{},
	}
	for _, slot := range selected.Family.Family.SlotSpecs {
		if err := ctx.Err(); err != nil {
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
		var candidate realizer.TargetCandidate
		var nodeID meaning.NodeID
		if slot.Category.IsPredicate() && slot.ArgumentBinding == nil {
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
			choices := make([]realizer.TargetCandidate, 0, 2)
			for _, available := range candidates {
				if available.NodeID == nodeID && available.Kind == meaning.KindReferent && available.Outcome == status.OutcomeOK {
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
			return realizer.ExecuteInput{}, runtimeError(status.OutcomeUnsupported, "UNBOUND_SLOT_UNSUPPORTED", "target.slots."+slot.ID, "only predicate VP/AP slots and explicitly role-bound argument slots are supported")
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
		input.Slots[slot.ID] = realizer.SlotRequest{
			NodeID: nodeID, Candidate: candidate, PrincipalParts: parts,
			Features: bundle, CaseFeature: builder.CaseFeature,
		}
	}
	return input, nil
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

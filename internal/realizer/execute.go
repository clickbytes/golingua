package realizer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultExecuteMaxSlots      = construction.MaxSlotsPerFamily
	maxExecuteMaxSlots          = construction.MaxSlotsPerFamily
	defaultExecuteMaxTraceSteps = 4096
	maxExecuteMaxTraceSteps     = morph.MaxGenerationTrace * construction.MaxSlotsPerFamily
	defaultExecuteMaxTextBytes  = 1 << 20
	maxExecuteMaxTextBytes      = 4 << 20
)

// ExecuteBudget bounds generated slot count and aggregate retained surface and
// morphology trace data. Morphology has its own bounded procedure budget.
type ExecuteBudget struct {
	MaxSlots      int
	MaxTraceSteps int
	MaxTextBytes  int
	Procedure     procedure.Budget
}

// DefaultExecuteBudget returns finite limits for one surface execution.
func DefaultExecuteBudget() ExecuteBudget {
	return ExecuteBudget{
		MaxSlots: defaultExecuteMaxSlots, MaxTraceSteps: defaultExecuteMaxTraceSteps,
		MaxTextBytes: defaultExecuteMaxTextBytes, Procedure: procedure.DefaultBudget(),
	}
}

func (budget ExecuteBudget) Validate() error {
	if budget.MaxSlots <= 0 || budget.MaxSlots > maxExecuteMaxSlots ||
		budget.MaxTraceSteps <= 0 || budget.MaxTraceSteps > maxExecuteMaxTraceSteps ||
		budget.MaxTextBytes <= 0 || budget.MaxTextBytes > maxExecuteMaxTextBytes {
		return executeError(status.OutcomeInvalid, "INVALID_EXECUTE_BUDGET", "budget", "surface execution limits must be positive and within package ceilings")
	}
	if err := budget.Procedure.Validate(); err != nil {
		return err
	}
	return nil
}

// SlotRequest supplies an explicit target lexical choice and morphology input
// for one family slot. Feature values are caller supplied; execution does not
// infer case, agreement, or defaults from semantic roles. Adposition-bearing
// slots require a matching terminal in the selected construction sequence.
type SlotRequest struct {
	NodeID         meaning.NodeID
	Candidate      TargetCandidate
	PrincipalParts morph.PrincipalParts
	Features       features.FeatureBundle
	// CaseFeature names the explicit morphology feature used to realize a
	// licensed case label, if the family or frame declares one.
	CaseFeature features.FeatureID
	// NP carries a typed phrase plan and exact child execution requests when
	// this slot is a licensed noun phrase. The current linearizer still uses
	// the outer construction sequence; this plan is retained for the next NP
	// execution stage and does not alter surface output.
	NP *NPExecutionRequest
	// VPAuxiliary is an opt-in language-local realization attached to this
	// exact main predicate slot. It never introduces a Meaning node.
	VPAuxiliary *VPAuxiliaryExecutionRequest `json:"vp_auxiliary,omitempty"`
}

// VPAuxiliaryExecutionRequest supplies an auxiliary lexeme and its finite
// morphology, plus the selected predicate's lexical base-form feature bundle.
// A negation particle is allowed only when the exact plan operator scopes over
// EventNodeID.
type VPAuxiliaryExecutionRequest struct {
	EventNodeID      meaning.NodeID              `json:"event_node_id"`
	Lexeme           lexicon.MorphLexeme         `json:"lexeme"`
	PrincipalParts   morph.PrincipalParts        `json:"principal_parts"`
	Features         features.FeatureBundle      `json:"features"`
	BaseFeatures     features.FeatureBundle      `json:"base_features"`
	NegationParticle *VPAuxiliaryParticleRequest `json:"negation_particle,omitempty"`
}

// VPAuxiliaryParticleRequest is one supplied target-language terminal tied to
// an exact scoped negation operator.
type VPAuxiliaryParticleRequest struct {
	OperatorID construction.OperatorID `json:"operator_id"`
	Surface    string                  `json:"surface"`
	Provenance provenance.SourceRef    `json:"provenance"`
}

// NPExecutionRequest is a compiled, language-local NP plan tied to the exact
// lexical head and optional modifier selected for one Meaning referent.
type NPExecutionRequest struct {
	Plan                  construction.CompiledPhrasePlan `json:"plan"`
	ProfileVersion        string                          `json:"profile_version"`
	VariantID             construction.VariantID          `json:"variant_id"`
	HeadSlotID            string                          `json:"head_slot_id"`
	HeadSequenceIndex     int                             `json:"head_sequence_index"`
	Head                  NPChildExecutionRequest         `json:"head"`
	ModifierSlotID        string                          `json:"modifier_slot_id,omitempty"`
	ModifierSequenceIndex int                             `json:"modifier_sequence_index"`
	Modifier              *NPChildExecutionRequest        `json:"modifier,omitempty"`
	Determiner            *NPDeterminerExecutionRequest   `json:"determiner,omitempty"`
	// Bare records that this NP has no selected determiner. It is a construction
	// choice, not an inferred semantic feature.
	Bare bool `json:"bare"`
}

// NPChildExecutionRequest preserves the exact target lexical and morphology
// choice for one phrase-plan input.
type NPChildExecutionRequest struct {
	InputID        string                 `json:"input_id"`
	Role           string                 `json:"role"`
	NodeID         meaning.NodeID         `json:"node_id"`
	Candidate      TargetCandidate        `json:"candidate"`
	PrincipalParts morph.PrincipalParts   `json:"principal_parts"`
	Features       features.FeatureBundle `json:"features"`
	CaseFeature    features.FeatureID     `json:"case_feature,omitempty"`
	Assumptions    []FieldAssumption      `json:"assumptions,omitempty"`
}

// NPDeterminerExecutionRequest links a determiner licensed by the phrase
// profile to the local NP plan. Embedded determiners are emitted as NP parts
// and do not claim an outer sequence position.
type NPDeterminerExecutionRequest struct {
	InputID       string                 `json:"input_id"`
	Surface       string                 `json:"surface"`
	VariantID     construction.VariantID `json:"variant_id"`
	SequenceIndex int                    `json:"sequence_index,omitempty"`
	Embedded      bool                   `json:"embedded,omitempty"`
	Features      features.FeatureBundle `json:"features"`
}

// ExecuteInput contains one validated plan, its source Meaning, the exact
// family/frame license, and caller supplied morphology inputs for each slot.
type ExecuteInput struct {
	Plan           RealizationPlan
	SourceMeaning  meaning.Meaning
	SourceContexts []meaning.RequestContext
	Core           *core.ConceptGraph
	Family         construction.ConstructionFamily
	Frame          valency.ValencyFrame
	FeatureProfile *features.LanguageFeatureProfile
	PhraseProfile  construction.PhraseProfile
	Slots          map[string]SlotRequest
	// InformationStructure carries explicit caller evidence for variant
	// information-structure conditions, keyed by semantic role.
	InformationStructure map[meaning.Role]construction.InformationValue
	Paradigms            *morph.CompiledParadigms
	Exceptions           *morph.CompiledExceptions
	ProcedureBudget      procedure.Budget
	Budget               ExecuteBudget
}

// SlotForm is a generated lexical form associated with its semantic source.
// Trace is the morphology trace for this exact lexeme and feature request.
type SlotForm struct {
	SlotID      string                      `json:"slot_id"`
	Surface     string                      `json:"surface"`
	NodeID      meaning.NodeID              `json:"node_id"`
	Role        meaning.Role                `json:"role,omitempty"`
	MeaningPath []meaning.NodeID            `json:"meaning_path"`
	LexemeID    ids.LexemeID                `json:"lexeme_id"`
	SenseID     ids.SenseID                 `json:"sense_id"`
	Features    []features.FeatureValue     `json:"features"`
	Trace       []morph.GenerationTraceStep `json:"trace"`
	Provenance  provenance.SourceRef        `json:"provenance"`
	VPAuxiliary *VPAuxiliaryForm            `json:"vp_auxiliary,omitempty"`
}

// VPAuxiliaryForm preserves the auxiliary and lexical base forms emitted at
// one exact predicate slot. EventNodeID is the existing Meaning event ID.
type VPAuxiliaryForm struct {
	SlotID      string                   `json:"slot_id"`
	EventNodeID meaning.NodeID           `json:"event_node_id"`
	Auxiliary   VPAuxiliaryMorphForm     `json:"auxiliary"`
	Base        VPAuxiliaryMorphForm     `json:"base"`
	Negation    *VPAuxiliaryParticleForm `json:"negation,omitempty"`
}

// VPAuxiliaryMorphForm is a generated lexical form without a semantic sense.
type VPAuxiliaryMorphForm struct {
	Surface    string                      `json:"surface"`
	LexemeID   ids.LexemeID                `json:"lexeme_id"`
	Features   []features.FeatureValue     `json:"features"`
	Trace      []morph.GenerationTraceStep `json:"trace"`
	Provenance provenance.SourceRef        `json:"provenance"`
}

// VPAuxiliaryParticleForm is one exact supplied language-local operator word.
type VPAuxiliaryParticleForm struct {
	Surface     string                  `json:"surface"`
	OperatorID  construction.OperatorID `json:"operator_id"`
	ScopeNodeID meaning.NodeID          `json:"scope_node_id"`
	Provenance  provenance.SourceRef    `json:"provenance"`
}

// ExecuteResult preserves generated forms in family slot order. Optional
// omitted slots have no SlotForm; the family remains available to linearize.
type ExecuteResult struct {
	Outcome     status.OutcomeKind              `json:"outcome"`
	Family      construction.ConstructionFamily `json:"family"`
	Slots       []SlotForm                      `json:"slots"`
	NPForms     []NPForm                        `json:"np_forms,omitempty"`
	Diagnostics []status.Diagnostic             `json:"diagnostics,omitempty"`
}

// NPForm is a prepared phrase in template order. It retains every outer
// sequence position claimed by the NP for a future linearizer integration.
type NPForm struct {
	SlotID            string                          `json:"slot_id"`
	HeadNodeID        meaning.NodeID                  `json:"head_node_id"`
	VariantID         construction.VariantID          `json:"variant_id"`
	VariantProvenance provenance.SourceRef            `json:"variant_provenance"`
	Plan              construction.CompiledPhrasePlan `json:"plan"`
	// DeterminerOnset is evidence from the exact first lexical child selected
	// during execution. Empty means that onset is unknown.
	DeterminerOnset      string         `json:"determiner_onset,omitempty"`
	EmbeddedDeterminer   bool           `json:"embedded_determiner,omitempty"`
	Parts                []NPFormPart   `json:"parts"`
	OuterSequenceIndexes []int          `json:"outer_sequence_indexes"`
	OuterSlotIDs         []string       `json:"outer_slot_ids"`
	OuterTokens          []NPOuterToken `json:"outer_tokens"`
}

// NPOuterToken identifies one exact outer sequence token that belongs to this
// phrase. SlotIDs are empty for terminals.
type NPOuterToken struct {
	SequenceIndex int    `json:"sequence_index"`
	SlotID        string `json:"slot_id,omitempty"`
	Terminal      string `json:"terminal,omitempty"`
}

// NPFormPart is one ordered child form or one variant terminal.
type NPFormPart struct {
	Role               string                       `json:"role"`
	InputID            string                       `json:"input_id"`
	Kind               construction.PhraseInputKind `json:"kind"`
	Surface            string                       `json:"surface"`
	SlotForm           *SlotForm                    `json:"slot_form,omitempty"`
	TerminalProvenance provenance.SourceRef         `json:"terminal_provenance,omitempty"`
}

// ExecuteError reports invalid setup or a controlled refusal with a typed,
// bounded result.
type ExecuteError struct{ result status.Result }

func (failure *ExecuteError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "surface execution failed"
	}
	return failure.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed execution failure.
func (failure *ExecuteError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

// Execute generates each requested slot through the morphology engine. It
// validates the selected plan against the supplied family, variant, frame,
// and source Meaning before it accepts any generated text.
func Execute(input ExecuteInput) (ExecuteResult, error) {
	budget := input.Budget
	if budget == (ExecuteBudget{}) {
		budget = DefaultExecuteBudget()
	}
	if err := budget.Validate(); err != nil {
		return ExecuteResult{}, err
	}
	if input.ProcedureBudget != (procedure.Budget{}) {
		if err := input.ProcedureBudget.Validate(); err != nil {
			return ExecuteResult{}, err
		}
		budget.Procedure = input.ProcedureBudget
	}
	if input.Core == nil || input.Paradigms == nil {
		return ExecuteResult{}, executeError(status.OutcomeUnknown, "MORPHOLOGY_OR_CORE_UNAVAILABLE", "indexes", "CORE graph and compiled target paradigms are required")
	}
	plan := input.Plan
	if plan.ID == "" || plan.ID != realizationPlanID(plan) || !plan.Outcome.Valid() || len(plan.Diagnostics) > status.MaxDiagnostics {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "INVALID_REALIZATION_PLAN", "plan", "execution requires an unmodified canonical realization plan")
	}
	for index, diagnostic := range plan.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "INVALID_PLAN_DIAGNOSTIC", fmt.Sprintf("plan.diagnostics[%d]", index), err.Error())
		}
	}
	if plan.Outcome == status.OutcomeInvalid || plan.Outcome == status.OutcomeUnsupported || plan.Outcome == status.OutcomeUnknown || plan.Outcome == status.OutcomeAmbiguous || plan.Outcome == status.OutcomeResourceLimit || plan.Outcome == status.OutcomeError {
		return ExecuteResult{}, executeError(plan.Outcome, "PLAN_NOT_EXECUTABLE", "plan.outcome", "the selected plan outcome does not authorize surface generation")
	}
	if err := meaning.ValidateMeaning(input.SourceMeaning, input.Core, meaning.DefaultBudget()); err != nil {
		return ExecuteResult{}, executeWrapError(err)
	}
	if plan.Kind != meaning.KindEvent && plan.Kind != meaning.KindState {
		return ExecuteResult{}, executeError(status.OutcomeUnsupported, "PLAN_HAS_NO_PREDICATE_FRAME", "plan.kind", "surface execution requires an event or state construction plan")
	}
	path, predicate, found := pathToNode(input.SourceMeaning, plan.NodeID, meaning.DefaultBudget().MaxDepth)
	if !found || predicate.Kind != plan.Kind || !equalNodePath(path, plan.MeaningPath) || string(predicate.Concept) != plan.CoreConceptID.String() {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_MEANING_MISMATCH", "plan.meaning_path", "plan predicate identity or path differs from the supplied Meaning graph")
	}
	if err := input.Family.Validate(); err != nil {
		return ExecuteResult{}, executeWrapError(err)
	}
	for _, slot := range input.Family.SlotSpecs {
		if slot.ModifierBinding == nil {
			continue
		}
		_, outcome := FindModifierOutcome(input.SourceMeaning, plan.NodeID, input.Family, slot.ID)
		if outcome == status.OutcomeOK {
			continue
		}
		code := "MODIFIER_STATE_UNAVAILABLE"
		if outcome == status.OutcomeAmbiguous {
			code = "MODIFIER_STATE_AMBIGUOUS"
		}
		return ExecuteResult{}, executeError(outcome, code, "source.meaning.modifiers."+slot.ID, "modifier slot does not have exactly one state linked to its exact NP referent")
	}
	if !CoversLexicalMeaning(input.SourceMeaning, plan.NodeID, input.Family) {
		return ExecuteResult{}, executeError(status.OutcomeUnsupported, "MEANING_NOT_COVERED_BY_FAMILY", "source.meaning", "selected construction does not cover every lexical node in the source Meaning")
	}
	if input.Family.ID != plan.FamilyID || plan.CompiledVariant.FamilyID != plan.FamilyID || plan.CompiledVariant.PlanID != plan.FamilyPlanID {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_FAMILY_MISMATCH", "family", "supplied family identity or plan version differs from the selected plan")
	}
	var compiled construction.CompiledVariant
	var err error
	if plan.QuestionAnswer != nil {
		answer := *plan.QuestionAnswer
		if plan.NodeID != answer.TargetNodeID {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "QUESTION_ANSWER_TARGET_MISMATCH", "plan.question_answer.target_node_id", "typed WH answer must target the selected realization node")
		}
		if err := answer.Validate(input.SourceMeaning, input.Core); err != nil {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "INVALID_QUESTION_ANSWER", "plan.question_answer", err.Error())
		}
		if err := validateQuestionDiscourse(input.SourceMeaning, answer); err != nil || !hasExactQuestionOperator(input.SourceMeaning, plan.Construction.Steps, answer) {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "QUESTION_OPERATOR_MISMATCH", "plan.construction.steps", "typed WH answer requires the exact question discourse and operator")
		}
		compiled, err = construction.ValidateVariantForQuestionGap(plan.CompiledVariant.Variant, input.Family, input.Frame, answer.Role)
		if err == nil && !questionGapVariantHasOnlyOmittedRequiredRole(plan.CompiledVariant.Variant, input.Family, input.Frame, answer.Role) {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "QUESTION_GAP_ADDITIONAL_ROLE_OMITTED", "plan.compiled_variant", "all required roles except the typed WH answer role must have explicit target slots")
		}
	} else {
		if hasQuestionAct(input.SourceMeaning.Nodes) {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "QUESTION_ANSWER_UNAVAILABLE", "plan.question_answer", "question realization requires an explicit typed answer role")
		}
		compiled, err = construction.ValidateVariantForFrame(plan.CompiledVariant.Variant, input.Family, input.Frame)
	}
	if err != nil {
		return ExecuteResult{}, executeWrapError(err)
	}
	if compiled.PlanID != plan.FamilyPlanID || compiled.FamilyID != plan.FamilyID || compiled.Variant.ID != plan.VariantID || !reflect.DeepEqual(compiled, plan.CompiledVariant) {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_VARIANT_MISMATCH", "plan.compiled_variant", "compiled variant does not match the supplied family and plan")
	}
	if input.Frame.ID != plan.FrameID || input.Frame.SenseID != plan.SenseID {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_FRAME_MISMATCH", "frame", "supplied valency frame differs from the selected plan sense or frame")
	}
	omitted := make(map[meaning.Role]bool, len(plan.MissingOmittedRoles))
	for _, role := range plan.MissingOmittedRoles {
		if omitted[role] || !containsRole(input.Frame.SurfaceOmittedRoles, role) {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "OMITTED_ROLE_FRAME_MISMATCH", "plan.missing_omitted_roles", "only unique roles explicitly licensed by the exact selected frame may be omitted")
		}
		omitted[role] = true
	}
	if err := validateConstructionPlanForSurface(plan, input.Family); err != nil {
		return ExecuteResult{}, err
	}
	if len(input.Slots) > budget.MaxSlots {
		return ExecuteResult{}, executeLimit("slots", len(input.Slots), budget.MaxSlots)
	}
	if len(input.InformationStructure) > valency.MaxSemanticRoles {
		return ExecuteResult{}, executeLimit("information_structure", len(input.InformationStructure), valency.MaxSemanticRoles)
	}
	for role, value := range input.InformationStructure {
		if !validPlanField(string(role)) || !containsRole(plan.Roles, role) || !validInformationValue(value) {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "INVALID_INFORMATION_STRUCTURE_EVIDENCE", "information_structure."+string(role), "information-structure evidence must use a plan role and a supported explicit value")
		}
	}
	if len(input.Family.SlotSpecs) > budget.MaxSlots {
		return ExecuteResult{}, executeLimit("family_slots", len(input.Family.SlotSpecs), budget.MaxSlots)
	}

	slots := append([]construction.SlotSpec(nil), input.Family.SlotSpecs...)
	for left := 0; left < len(slots); left++ {
		for right := left + 1; right < len(slots); right++ {
			if slots[right].Position < slots[left].Position {
				slots[left], slots[right] = slots[right], slots[left]
			}
		}
	}
	slotByID := make(map[string]construction.SlotSpec, len(slots))
	predicateSlots := 0
	for _, slot := range slots {
		slotByID[slot.ID] = slot
		if slot.IsMainPredicate() {
			predicateSlots++
		}
	}
	if predicateSlots != 1 {
		return ExecuteResult{}, executeError(status.OutcomeUnsupported, "PREDICATE_SLOT_CARDINALITY", "family.slot_specs", "surface execution requires exactly one unbound predicate VP/AP slot")
	}
	if err := validateMeaningRoleCoverage(predicate, slots); err != nil {
		return ExecuteResult{}, err
	}
	for slotID := range input.Slots {
		if _, exists := slotByID[slotID]; !exists {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "UNKNOWN_SLOT_REQUEST", "slots."+slotID, "slot request is absent from the selected construction family")
		}
	}
	requests := make(map[string]SlotRequest, len(input.Slots))
	for _, slot := range slots {
		req, supplied := input.Slots[slot.ID]
		if !supplied {
			if slot.IsMainPredicate() {
				return ExecuteResult{}, executeError(status.OutcomeIncomplete, "PREDICATE_SLOT_UNAVAILABLE", "slots."+slot.ID, "the selected event/state predicate requires an explicit lexicalization request")
			}
			if slot.ArgumentBinding != nil {
				participant, count := participantByRole(predicate, slot.ArgumentBinding.SemanticRole)
				if omitted[slot.ArgumentBinding.SemanticRole] && count == 0 {
					continue
				}
				if count > 1 {
					return ExecuteResult{}, executeError(status.OutcomeAmbiguous, "MEANING_ROLE_AMBIGUOUS", "meaning."+string(plan.NodeID)+".participants", "source Meaning supplies multiple targets for one semantic role")
				}
				if count == 1 {
					if _, licensed := construction.ImplicitParticipantSlot(input.Family, plan.CompiledVariant.Variant, slot.ID); licensed {
						if _, anchored := ImplicitParticipantAnchor(input.SourceMeaning, input.SourceContexts, participant.Target); anchored {
							continue
						}
					}
					return ExecuteResult{}, executeError(status.OutcomeIncomplete, "PRESENT_ROLE_SLOT_UNAVAILABLE", "slots."+slot.ID, "a semantic role present in Meaning requires an explicit slot lexicalization request")
				}
			}
			if slot.Optionality == construction.SlotRequired || slot.Cardinality.Minimum > 0 {
				return ExecuteResult{}, executeError(status.OutcomeIncomplete, "REQUIRED_SLOT_UNAVAILABLE", "slots."+slot.ID, "required construction slot has no explicit lexicalization request")
			}
			continue
		}
		if slot.Cardinality.Minimum > 1 || slot.Cardinality.Maximum < 1 {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "SLOT_CARDINALITY_UNSUPPORTED", "family.slot_specs."+slot.ID, "one supplied lexical form cannot satisfy this slot cardinality")
		}
		requests[slot.ID] = req
	}
	sequenceSlots := make(map[string]struct{}, len(plan.CompiledVariant.Variant.Sequence))
	for _, token := range plan.CompiledVariant.Variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	for slotID := range requests {
		if _, exists := sequenceSlots[slotID]; !exists {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "UNSEQUENCED_SLOT_REQUEST", "slots."+slotID, "slot request is not present in the selected variant sequence")
		}
	}
	for slotID, request := range requests {
		if request.VPAuxiliary != nil && !slotByID[slotID].IsMainPredicate() {
			return ExecuteResult{}, executeError(status.OutcomeUnsupported, "VP_AUXILIARY_NON_PREDICATE_SLOT", "slots."+slotID+".vp_auxiliary", "VP auxiliary execution can only attach to the exact main predicate slot")
		}
	}
	if err := validateSlotCasePolicies(input.Frame, slots, requests, input.Family, plan.CompiledVariant.Variant); err != nil {
		return ExecuteResult{}, err
	}
	if err := validateVariantConditions(plan.CompiledVariant.Variant.Conditions, requests, input.InformationStructure); err != nil {
		return ExecuteResult{}, err
	}
	if err := validateHardAgreement(input.Family, slotByID, requests); err != nil {
		return ExecuteResult{}, err
	}

	outcome := plan.Outcome
	diagnostics := append([]status.Diagnostic(nil), plan.Diagnostics...)
	result := ExecuteResult{Outcome: outcome, Family: cloneExecutionFamily(input.Family), Slots: []SlotForm{}, Diagnostics: diagnostics}
	if len(omitted) > 0 {
		result.Outcome = mergePlanOutcome(result.Outcome, status.OutcomeIncomplete)
		for role := range omitted {
			result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "TARGET_ROLE_ABSENT_BY_FRAME_LICENSE", Path: "meaning." + string(plan.NodeID) + ".participants." + string(role), Message: "licensed surface omission leaves this semantic role unresolved", Severity: status.SeverityWarning})
		}
	}
	traceSteps, textBytes := 0, 0
	for _, slot := range slots {
		req, supplied := requests[slot.ID]
		if !supplied {
			continue
		}
		if slot.Category == construction.CategoryPP || slot.Adposition != "" {
			if slot.Adposition == "" || !construction.VariantEmitsAdposition(plan.CompiledVariant.Variant, input.Family, slot.ID, slot.Adposition) {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "ADPOSITION_SEQUENCE_UNAVAILABLE", "family.slot_specs."+slot.ID, "an adpositional argument requires its exact adposition terminal before the slot in the selected construction")
			}
		}
		role := meaning.Role("")
		var sourcePath []meaning.NodeID
		var sourceNode meaning.Node
		if slot.ModifierBinding != nil {
			modifier, outcome := FindModifierOutcome(input.SourceMeaning, plan.NodeID, input.Family, slot.ID)
			if outcome != status.OutcomeOK {
				code := "MODIFIER_STATE_UNAVAILABLE"
				if outcome == status.OutcomeAmbiguous {
					code = "MODIFIER_STATE_AMBIGUOUS"
				}
				return ExecuteResult{}, executeError(outcome, code, "source.meaning.modifiers."+slot.ID, "declared modifier slot has no unique linked root state")
			}
			modifierConceptID, parseErr := ids.ParseConceptID(string(modifier.Concept))
			if parseErr != nil {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "MODIFIER_PROPERTY_UNSUPPORTED", "meaning."+string(modifier.ID)+".concept", "modifier state does not identify a CORE property")
			}
			modifierConcept, exists := input.Core.Lookup(modifierConceptID)
			if !exists || modifierConcept.Kind != core.KindProperty {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "MODIFIER_PROPERTY_UNSUPPORTED", "meaning."+string(modifier.ID)+".concept", "modifier binding requires a CORE property concept")
			}
			if req.NodeID != modifier.ID {
				return ExecuteResult{}, executeError(status.OutcomeInvalid, "SLOT_NODE_MODIFIER_MISMATCH", "slots."+slot.ID+".node_id", "slot node does not match the exact modifier state linked to its NP target")
			}
			sourcePath, sourceNode, found = pathToNode(input.SourceMeaning, modifier.ID, meaning.DefaultBudget().MaxDepth)
			if !found {
				return ExecuteResult{}, executeError(status.OutcomeUnknown, "MODIFIER_NODE_UNREACHABLE", "slots."+slot.ID+".node_id", "modifier state is absent or unreachable in source Meaning")
			}
			role = slot.ModifierBinding.SemanticRole
		} else if slot.ArgumentBinding != nil {
			role = slot.ArgumentBinding.SemanticRole
			participant, count := participantByRole(predicate, role)
			if count == 0 {
				return ExecuteResult{}, executeError(status.OutcomeIncomplete, "MEANING_ROLE_UNAVAILABLE", "meaning."+string(plan.NodeID)+".participants", "construction slot role is not present in source Meaning")
			}
			if count > 1 {
				return ExecuteResult{}, executeError(status.OutcomeAmbiguous, "MEANING_ROLE_AMBIGUOUS", "meaning."+string(plan.NodeID)+".participants", "source Meaning supplies multiple targets for one semantic role")
			}
			if req.NodeID != participant.Target {
				return ExecuteResult{}, executeError(status.OutcomeInvalid, "SLOT_NODE_ROLE_MISMATCH", "slots."+slot.ID+".node_id", "slot node does not match its exact semantic-role participant")
			}
			sourcePath, sourceNode, found = pathToNode(input.SourceMeaning, participant.Target, meaning.DefaultBudget().MaxDepth)
			if !found {
				return ExecuteResult{}, executeError(status.OutcomeUnknown, "SLOT_NODE_UNREACHABLE", "slots."+slot.ID+".node_id", "argument node is absent or unreachable in source Meaning")
			}
		} else {
			if !slot.IsMainPredicate() || req.NodeID != plan.NodeID {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "UNBOUND_NON_PREDICATE_SLOT", "family.slot_specs."+slot.ID, "only the predicate VP/AP slot may be unbound to a semantic argument role")
			}
			sourceNode, sourcePath = predicate, path
		}
		if err := validateSlotCandidate(req.Candidate, req.NodeID, sourceNode, plan.Language, slot.ID); err != nil {
			return ExecuteResult{}, err
		}
		if sourceNode.ID == plan.NodeID && (req.Candidate.Lexeme.ID != plan.LexemeID || req.Candidate.Sense.ID != plan.SenseID || req.Candidate.TransferBinding.CoreRef != plan.CoreConceptID) {
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "PREDICATE_CANDIDATE_MISMATCH", "slots."+slot.ID+".candidate", "predicate slot lexical choice differs from the selected realization plan")
		}
		if err := requirePlanFeatureConstraints(plan.FeatureConstraints, sourcePath, req.Features); err != nil {
			return ExecuteResult{}, err
		}
		morphLexeme := req.Candidate.Lexeme
		if req.Features.Language != features.LanguageID(morphLexeme.Language) {
			return ExecuteResult{}, executeError(status.OutcomeIncomplete, "FEATURE_LANGUAGE_MISMATCH", "slots."+slot.ID+".features.language", "slot feature bundle must explicitly match the target lexeme language")
		}
		generation, generationErr := morph.Generate(morphLexeme, req.PrincipalParts, req.Features, input.Paradigms, input.Exceptions, budget.Procedure)
		if generationErr != nil {
			return ExecuteResult{}, executeWrapError(generationErr)
		}
		switch generation.State {
		case morph.GenerationGenerated:
			if len(generation.Forms) != 1 {
				if len(generation.Forms) > 1 {
					return ExecuteResult{}, executeError(status.OutcomeAmbiguous, "SLOT_FORM_AMBIGUOUS", "slots."+slot.ID, "morphology produced multiple licensed surface forms; no form was selected")
				}
				return ExecuteResult{}, executeError(status.OutcomeIncomplete, "SLOT_FORM_MISSING", "slots."+slot.ID, "morphology reported generated without a surface form")
			}
		case morph.GenerationBlocked:
			return ExecuteResult{}, executeFromDiagnostics(status.OutcomeUnsupported, "MORPHOLOGY_BLOCKED", "slots."+slot.ID, "morphology explicitly blocks this requested cell", generation.Diagnostics)
		case morph.GenerationIncomplete:
			return ExecuteResult{}, executeFromDiagnostics(status.OutcomeIncomplete, "MORPHOLOGY_INCOMPLETE", "slots."+slot.ID, "morphology lacks required evidence for this requested cell", generation.Diagnostics)
		case morph.GenerationUnavailable, morph.GenerationUnknown:
			return ExecuteResult{}, executeFromDiagnostics(status.OutcomeUnknown, "MORPHOLOGY_UNKNOWN", "slots."+slot.ID, "morphology could not resolve this requested cell", generation.Diagnostics)
		default:
			return ExecuteResult{}, executeError(status.OutcomeInvalid, "INVALID_MORPHOLOGY_STATE", "slots."+slot.ID, "morphology returned an invalid generation state")
		}
		form := generation.Forms[0]
		traceSteps += len(generation.Trace)
		textBytes += len(form.Surface)
		var vpAuxiliary *VPAuxiliaryForm
		if req.VPAuxiliary != nil {
			if !slot.IsMainPredicate() || req.NodeID != plan.NodeID || sourceNode.Kind != meaning.KindEvent {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "VP_AUXILIARY_EVENT_MISMATCH", "slots."+slot.ID+".vp_auxiliary", "VP auxiliary execution requires the exact main event predicate slot")
			}
			generated, auxErr := generateVPAuxiliary(input, slot, req, budget.Procedure)
			if auxErr != nil {
				return ExecuteResult{}, auxErr
			}
			vpAuxiliary = &generated
			traceSteps += len(generated.Auxiliary.Trace) + len(generated.Base.Trace)
			textBytes += len(generated.Auxiliary.Surface) + len(generated.Base.Surface)
			if generated.Negation != nil {
				textBytes += len(generated.Negation.Surface)
			}
		}
		if traceSteps > budget.MaxTraceSteps {
			return ExecuteResult{}, executeLimit("trace_steps", traceSteps, budget.MaxTraceSteps)
		}
		if textBytes > budget.MaxTextBytes {
			return ExecuteResult{}, executeLimit("text_bytes", textBytes, budget.MaxTextBytes)
		}
		result.Slots = append(result.Slots, SlotForm{
			SlotID: slot.ID, Surface: form.Surface, NodeID: req.NodeID, Role: role,
			MeaningPath: append([]meaning.NodeID(nil), sourcePath...), LexemeID: req.Candidate.Lexeme.ID,
			SenseID: req.Candidate.Sense.ID, Features: append([]features.FeatureValue(nil), req.Features.Values...),
			Trace: append([]morph.GenerationTraceStep(nil), generation.Trace...), Provenance: req.Candidate.TransferBinding.Provenance,
			VPAuxiliary: vpAuxiliary,
		})
	}
	npForms, err := prepareNPForms(input, result, requests, slotByID, budget)
	if err != nil {
		return ExecuteResult{}, err
	}
	result.NPForms = npForms
	if len(result.Slots) == 0 {
		return ExecuteResult{}, executeError(status.OutcomeIncomplete, "NO_GENERATED_SLOTS", "slots", "no required or explicitly supplied optional slot produced a surface form")
	}
	return result, nil
}

func prepareNPForms(input ExecuteInput, result ExecuteResult, requests map[string]SlotRequest, familySlots map[string]construction.SlotSpec, budget ExecuteBudget) ([]NPForm, error) {
	ids := make([]string, 0)
	for slotID, request := range requests {
		if request.NP != nil {
			ids = append(ids, slotID)
		}
	}
	if len(ids) > budget.MaxSlots {
		return nil, executeLimit("np_forms", len(ids), budget.MaxSlots)
	}
	sort.Slice(ids, func(i, j int) bool {
		left, right := requestNPHeadIndex(requests[ids[i]]), requestNPHeadIndex(requests[ids[j]])
		if left != right {
			return left < right
		}
		return ids[i] < ids[j]
	})
	formsBySlot := make(map[string]SlotForm, len(result.Slots))
	for _, form := range result.Slots {
		if _, exists := formsBySlot[form.SlotID]; exists {
			return nil, executeError(status.OutcomeInvalid, "DUPLICATE_GENERATED_SLOT_FORM", "slots."+form.SlotID, "one outer slot generated more than one form")
		}
		formsBySlot[form.SlotID] = form
	}
	claimedIndexes := make(map[int]string)
	claimedSlots := make(map[string]string)
	forms := make([]NPForm, 0, len(ids))
	retainedTrace, retainedText := 0, 0
	for _, headSlotID := range ids {
		request := requests[headSlotID].NP
		if err := validateNPExecutionRequest(input, headSlotID, request, requests, familySlots); err != nil {
			return nil, err
		}
		form, err := buildNPForm(input, headSlotID, request, formsBySlot, requests)
		if err != nil {
			return nil, err
		}
		for _, part := range form.Parts {
			retainedText += len(part.Surface)
			if part.SlotForm != nil {
				retainedTrace += len(part.SlotForm.Trace)
			}
			if retainedText > budget.MaxTextBytes {
				return nil, executeLimit("np_form_text_bytes", retainedText, budget.MaxTextBytes)
			}
			if retainedTrace > budget.MaxTraceSteps {
				return nil, executeLimit("np_form_trace_steps", retainedTrace, budget.MaxTraceSteps)
			}
		}
		for i, index := range form.OuterSequenceIndexes {
			if previous, exists := claimedIndexes[index]; exists {
				return nil, executeError(status.OutcomeInvalid, "NP_OUTER_SEQUENCE_POSITION_REUSED", "np_forms."+headSlotID, "an outer sequence position is claimed by multiple NPs: "+previous)
			}
			claimedIndexes[index] = headSlotID
			if i < len(form.OuterSlotIDs) && form.OuterSlotIDs[i] != "" {
				slotID := form.OuterSlotIDs[i]
				if previous, exists := claimedSlots[slotID]; exists {
					return nil, executeError(status.OutcomeInvalid, "NP_OUTER_SLOT_REUSED", "np_forms."+headSlotID, "an outer slot is claimed by multiple NPs: "+previous)
				}
				claimedSlots[slotID] = headSlotID
			}
		}
		forms = append(forms, form)
	}
	return forms, nil
}

func requestNPHeadIndex(request SlotRequest) int {
	if request.NP == nil {
		return -1
	}
	return request.NP.HeadSequenceIndex
}

func validateNPExecutionRequest(input ExecuteInput, headSlotID string, request *NPExecutionRequest, requests map[string]SlotRequest, familySlots map[string]construction.SlotSpec) error {
	path := "slots." + headSlotID + ".np"
	if request == nil || input.FeatureProfile == nil || input.PhraseProfile.Language != input.Plan.Language || input.FeatureProfile.Language() != input.Plan.Language || input.PhraseProfile.Version == "" || request.ProfileVersion != input.PhraseProfile.Version {
		return executeError(status.OutcomeIncomplete, "NP_PROFILE_UNAVAILABLE", path+".profile", "NP execution requires the exact target feature and phrase profiles")
	}
	if err := construction.ValidatePhraseProfile(input.PhraseProfile, input.FeatureProfile); err != nil {
		return executeWrapError(err)
	}
	headSlot, exists := familySlots[headSlotID]
	if !exists || headSlot.Category != construction.CategoryNP || request.HeadSlotID != headSlotID || request.Head.Role != "head" || request.Head.InputID != "head-input" {
		return executeError(status.OutcomeInvalid, "NP_HEAD_SLOT_MISMATCH", path+".head", "NP request head identity must match its exact outer NP slot")
	}
	boundModifierCount := 0
	for _, spec := range input.Family.SlotSpecs {
		if spec.ModifierBinding != nil && spec.ModifierBinding.TargetSlot == headSlotID {
			boundModifierCount++
		}
	}
	if (request.Modifier == nil && boundModifierCount != 0) || (request.Modifier != nil && boundModifierCount != 1) {
		return executeError(status.OutcomeIncomplete, "NP_MODIFIER_BINDING_COVERAGE_MISMATCH", path+".modifier", "NP request must represent exactly the modifier slots bound to this head")
	}
	if request.VariantID != input.Plan.VariantID || request.VariantID != input.Plan.CompiledVariant.Variant.ID || request.Plan.Plan.Kind != construction.PhraseNP || request.Plan.Plan.TemplateID != "np-attributive" || request.Plan.Plan.Language != input.Plan.Language {
		return executeError(status.OutcomeInvalid, "NP_PLAN_CONTEXT_MISMATCH", path+".plan", "compiled NP plan must match the target language and exact licensed outer variant")
	}
	if request.HeadSequenceIndex != variantSlotIndex(input.Plan.CompiledVariant.Variant.Sequence, headSlotID) {
		return executeError(status.OutcomeInvalid, "NP_HEAD_SEQUENCE_INDEX_MISMATCH", path+".head_sequence_index", "head sequence index does not point to its exact outer slot")
	}
	compiled, err := construction.CompileNP(request.Plan.Plan, input.PhraseProfile, input.FeatureProfile)
	if err != nil {
		return executeWrapError(err)
	}
	if !reflect.DeepEqual(compiled, request.Plan) {
		return executeError(status.OutcomeInvalid, "NP_PLAN_NOT_CANONICAL", path+".plan", "compiled NP plan differs from a fresh compilation under the selected target profiles")
	}
	headRequest, supplied := requests[headSlotID]
	if !supplied || request.Head.NodeID != headRequest.NodeID || !reflect.DeepEqual(request.Head.Candidate, headRequest.Candidate) || !reflect.DeepEqual(request.Head.PrincipalParts, headRequest.PrincipalParts) || !reflect.DeepEqual(request.Head.Features, headRequest.Features) || request.Head.CaseFeature != headRequest.CaseFeature {
		return executeError(status.OutcomeInvalid, "NP_HEAD_CHILD_MISMATCH", path+".head", "NP head child must preserve the exact outer slot candidate and morphology request")
	}
	if err := features.ValidateBundle(input.FeatureProfile, request.Head.Features); err != nil {
		return executeWrapError(err)
	}
	headNode, found := meaningNodeForExecute(input.SourceMeaning, request.Head.NodeID)
	if !found || request.Head.Candidate.NodeID != request.Head.NodeID || request.Head.Candidate.Kind != meaning.KindReferent || request.Head.Candidate.TransferBinding.CoreRef.String() != string(headNode.Concept) {
		return executeError(status.OutcomeInvalid, "NP_HEAD_MEANING_MISMATCH", path+".head", "NP head candidate does not transfer the exact Meaning referent")
	}
	if request.Modifier != nil {
		if request.ModifierSlotID == "" || request.ModifierSequenceIndex < 0 || request.Modifier.InputID != "modifier-input" || request.Modifier.Role != "modifier" {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_SLOT_MISMATCH", path+".modifier", "modifier child requires one exact outer modifier slot and sequence index")
		}
		modifierSlot, exists := familySlots[request.ModifierSlotID]
		if !exists || modifierSlot.ModifierBinding == nil || modifierSlot.ModifierBinding.TargetSlot != headSlotID || request.Modifier.NodeID == "" {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_BINDING_MISMATCH", path+".modifier", "modifier slot must bind its state to this exact NP head")
		}
		if request.ModifierSequenceIndex != variantSlotIndex(input.Plan.CompiledVariant.Variant.Sequence, request.ModifierSlotID) {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_SEQUENCE_INDEX_MISMATCH", path+".modifier_sequence_index", "modifier sequence index does not point to its exact outer slot")
		}
		outerModifier, supplied := requests[request.ModifierSlotID]
		if !supplied || !reflect.DeepEqual(request.Modifier.Candidate, outerModifier.Candidate) || !reflect.DeepEqual(request.Modifier.PrincipalParts, outerModifier.PrincipalParts) || !reflect.DeepEqual(request.Modifier.Features, outerModifier.Features) || request.Modifier.CaseFeature != outerModifier.CaseFeature || request.Modifier.NodeID != outerModifier.NodeID {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_CHILD_MISMATCH", path+".modifier", "NP modifier child must preserve the exact selected outer modifier request")
		}
		modifierNode, outcome := FindModifierOutcome(input.SourceMeaning, input.Plan.NodeID, input.Family, request.ModifierSlotID)
		if outcome != status.OutcomeOK || modifierNode.ID != request.Modifier.NodeID || string(modifierNode.Concept) != request.Modifier.Candidate.TransferBinding.CoreRef.String() || request.Modifier.Candidate.Kind != meaning.KindState || request.Modifier.Candidate.Lexeme.POS != "adjective" {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_MEANING_MISMATCH", path+".modifier", "modifier candidate must realize the exact state whose bearer is this NP head")
		}
		participant, count := participantByRole(modifierNode, modifierSlot.ModifierBinding.SemanticRole)
		if count != 1 || participant.Target != request.Head.NodeID {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_BEARER_MISMATCH", path+".modifier", "modifier state bearer does not match the exact NP head referent")
		}
	} else if request.ModifierSlotID != "" || request.ModifierSequenceIndex >= 0 {
		return executeError(status.OutcomeInvalid, "NP_UNEXPECTED_MODIFIER_SLOT", path+".modifier", "modifier outer slot metadata is present without a modifier child")
	}
	if request.Determiner == nil {
		if !request.Bare {
			return executeError(status.OutcomeInvalid, "NP_DETERMINER_CHOICE_MISSING", path+".determiner", "determiner absence requires an explicit Bare choice")
		}
		firstIndex := request.HeadSequenceIndex
		if request.Modifier != nil && request.ModifierSequenceIndex < firstIndex {
			firstIndex = request.ModifierSequenceIndex
		}
		sequence := input.Plan.CompiledVariant.Variant.Sequence
		if firstIndex > 0 && sequence[firstIndex-1].Kind == construction.TokenTerminal && hasNPDeterminerRule(input.PhraseProfile, sequence[firstIndex-1].Terminal) {
			return executeError(status.OutcomeInvalid, "NP_BARE_TERMINAL_MISMATCH", path+".determiner", "Bare cannot omit a determiner terminal selected immediately before the NP")
		}
	} else {
		if request.Bare || request.Determiner.InputID != "determiner-input" || request.Determiner.VariantID != request.VariantID || !hasNPDeterminerRule(input.PhraseProfile, request.Determiner.Surface) {
			return executeError(status.OutcomeInvalid, "NP_DETERMINER_IDENTITY_MISMATCH", path+".determiner", "determiner must be tied to the selected outer variant and licensed phrase profile")
		}
		if request.Determiner.Embedded {
			if request.Determiner.SequenceIndex != -1 {
				return executeError(status.OutcomeInvalid, "NP_EMBEDDED_DETERMINER_OUTER_INDEX", path+".determiner.sequence_index", "embedded determiner must not claim an outer sequence index")
			}
			firstIndex := request.HeadSequenceIndex
			if request.Modifier != nil && request.ModifierSequenceIndex < firstIndex {
				firstIndex = request.ModifierSequenceIndex
			}
			sequence := input.Plan.CompiledVariant.Variant.Sequence
			if firstIndex > 0 && sequence[firstIndex-1].Kind == construction.TokenTerminal && hasNPDeterminerRule(input.PhraseProfile, sequence[firstIndex-1].Terminal) {
				return executeError(status.OutcomeInvalid, "NP_EMBEDDED_DETERMINER_OUTER_DUPLICATE", path+".determiner", "embedded determiner cannot duplicate a selected outer determiner terminal")
			}
		} else {
			if request.Determiner.SequenceIndex < 0 || request.Determiner.SequenceIndex >= len(input.Plan.CompiledVariant.Variant.Sequence) {
				return executeError(status.OutcomeInvalid, "NP_DETERMINER_SEQUENCE_INDEX_INVALID", path+".determiner.sequence_index", "outer determiner requires one exact selected variant sequence index")
			}
			token := input.Plan.CompiledVariant.Variant.Sequence[request.Determiner.SequenceIndex]
			if token.Kind != construction.TokenTerminal || token.Terminal != request.Determiner.Surface {
				return executeError(status.OutcomeInvalid, "NP_DETERMINER_TERMINAL_MISMATCH", path+".determiner", "determiner request does not match one licensed outer terminal")
			}
			firstIndex := request.HeadSequenceIndex
			if request.Modifier != nil && request.ModifierSequenceIndex < firstIndex {
				firstIndex = request.ModifierSequenceIndex
			}
			if request.Determiner.SequenceIndex != firstIndex-1 {
				return executeError(status.OutcomeInvalid, "NP_DETERMINER_SEQUENCE_POSITION_MISMATCH", path+".determiner.sequence_index", "determiner terminal must immediately precede the NP's first selected child")
			}
		}
	}
	if err := validateNPPlanInputs(input, path, request, headRequest, requests); err != nil {
		return err
	}
	return nil
}

func validateNPPlanInputs(input ExecuteInput, path string, request *NPExecutionRequest, head SlotRequest, requests map[string]SlotRequest) error {
	plan := request.Plan.Plan
	template, exists := phraseNPSpec(input.PhraseProfile, plan.TemplateID)
	if !exists {
		return executeError(status.OutcomeUnknown, "NP_TEMPLATE_UNAVAILABLE", path+".plan.template_id", "NP plan template is absent from the selected phrase profile")
	}
	expectedCount := 1
	if request.Modifier != nil {
		expectedCount++
	}
	if request.Determiner != nil {
		expectedCount++
	}
	if len(plan.Inputs) != expectedCount || len(plan.Children) != 0 || !reflect.DeepEqual(plan.Features, features.FeatureBundle{}) || plan.FrameID != "" || plan.GovernedCase != (features.FeatureValue{}) {
		return executeError(status.OutcomeInvalid, "NP_PLAN_SHAPE_MISMATCH", path+".plan", "NP plan inputs do not correspond exactly to selected head, modifier, and determiner")
	}
	inputs := make(map[string]construction.PhraseInput, len(plan.Inputs))
	for _, phraseInput := range plan.Inputs {
		if _, duplicate := inputs[phraseInput.Role]; duplicate {
			return executeError(status.OutcomeInvalid, "NP_PLAN_DUPLICATE_INPUT", path+".plan.inputs", "NP plan contains duplicate input roles")
		}
		inputs[phraseInput.Role] = phraseInput
	}
	for role, want := range map[string]struct {
		id       string
		kind     construction.PhraseInputKind
		features features.FeatureBundle
	}{
		"head": {"head-input", construction.PhraseInputLexical, projectedPhraseFeatures(input.FeatureProfile, template, "head", head.Features)},
	} {
		got, ok := inputs[role]
		if !ok || got.ID != want.id || got.Kind != want.kind || !got.Required || !reflect.DeepEqual(got.Features, want.features) {
			return executeError(status.OutcomeInvalid, "NP_HEAD_INPUT_EVIDENCE_MISMATCH", path+".plan.inputs.head", "head plan input does not match exact selected head features")
		}
	}
	if request.Modifier != nil {
		modifier := requests[request.ModifierSlotID]
		got, ok := inputs["modifier"]
		want := projectedPhraseFeatures(input.FeatureProfile, template, "modifier", modifier.Features)
		if !ok || got.ID != "modifier-input" || got.Kind != construction.PhraseInputAP || got.Required || !reflect.DeepEqual(got.Features, want) {
			return executeError(status.OutcomeInvalid, "NP_MODIFIER_INPUT_EVIDENCE_MISMATCH", path+".plan.inputs.modifier", "modifier plan input does not match exact selected AP features")
		}
	}
	if request.Determiner != nil {
		got, ok := inputs["determiner"]
		if !ok || got.ID != "determiner-input" || got.Kind != construction.PhraseInputLexical || got.Required || !reflect.DeepEqual(got.Features, request.Determiner.Features) {
			return executeError(status.OutcomeInvalid, "NP_DETERMINER_INPUT_EVIDENCE_MISMATCH", path+".plan.inputs.determiner", "determiner plan input does not match the selected outer terminal evidence")
		}
		if err := validateNPDeterminerFeatures(input, request, inputs["head"].Features); err != nil {
			return err
		}
	}
	wantOrder := make([]string, 0, expectedCount)
	present := map[string]bool{"head": true, "modifier": request.Modifier != nil, "determiner": request.Determiner != nil}
	for _, role := range template.Order {
		if present[role] {
			wantOrder = append(wantOrder, role)
		}
	}
	if !reflect.DeepEqual(plan.Order, wantOrder) {
		return executeError(status.OutcomeInvalid, "NP_PLAN_ORDER_MISMATCH", path+".plan.order", "NP input order differs from the selected language phrase profile")
	}
	if err := validateNPAgreements(input, path, request, inputs); err != nil {
		return err
	}
	return nil
}

func validateNPAgreements(input ExecuteInput, path string, request *NPExecutionRequest, inputs map[string]construction.PhraseInput) error {
	template, ok := phraseNPSpec(input.PhraseProfile, request.Plan.Plan.TemplateID)
	if !ok {
		return executeError(status.OutcomeIncomplete, "NP_AGREEMENT_TEMPLATE_UNAVAILABLE", path+".plan.template_id", "selected NP phrase template is unavailable for agreement validation")
	}
	want := construction.ApplicablePhraseAgreements(template, request.Plan.Plan.Inputs)
	if !sameNPAgreementSet(request.Plan.Plan.Agreements, want) {
		return executeError(status.OutcomeInvalid, "NP_PLAN_AGREEMENT_MISMATCH", path+".plan.agreements", "compiled NP agreements do not match the exact known feature evidence")
	}
	for _, agreement := range want {
		controller, controllerOK := inputs[roleForNPInputID(request.Plan.Plan.Inputs, agreement.ControllerInput)]
		target, targetOK := inputs[roleForNPInputID(request.Plan.Plan.Inputs, agreement.TargetInput)]
		if !controllerOK || !targetOK {
			return executeError(status.OutcomeIncomplete, "NP_AGREEMENT_INPUT_INCOMPLETE", path+".plan.agreements", "declared agreement input evidence is unavailable")
		}
		left, leftOK := knownFeature(controller.Features.Values, agreement.Feature)
		right, rightOK := knownFeature(target.Features.Values, agreement.Feature)
		if !leftOK || !rightOK {
			return executeError(status.OutcomeIncomplete, "NP_AGREEMENT_EVIDENCE_INCOMPLETE", path+".plan.agreements."+string(agreement.Feature), "declared NP agreement requires known values on both exact phrase inputs")
		}
		if left != right {
			return executeError(status.OutcomeUnsupported, "NP_AGREEMENT_FEATURE_MISMATCH", path+".plan.agreements."+string(agreement.Feature), "known NP input values do not satisfy the selected template agreement")
		}
	}
	return nil
}

func sameNPAgreementSet(left, right []construction.FeatureAgreement) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[construction.FeatureAgreement]struct{}, len(left))
	for _, agreement := range left {
		if _, duplicate := seen[agreement]; duplicate {
			return false
		}
		seen[agreement] = struct{}{}
	}
	for _, agreement := range right {
		if _, exists := seen[agreement]; !exists {
			return false
		}
	}
	return true
}

func roleForNPInputID(inputs []construction.PhraseInput, id string) string {
	for _, input := range inputs {
		if input.ID == id {
			return input.Role
		}
	}
	return ""
}

func buildNPForm(input ExecuteInput, headSlotID string, request *NPExecutionRequest, formsBySlot map[string]SlotForm, requests map[string]SlotRequest) (NPForm, error) {
	form := NPForm{SlotID: headSlotID, HeadNodeID: request.Head.NodeID, VariantID: request.VariantID, VariantProvenance: input.Plan.CompiledVariant.Variant.Provenance, Plan: request.Plan, Parts: []NPFormPart{}, OuterSequenceIndexes: []int{}, OuterSlotIDs: []string{}, OuterTokens: []NPOuterToken{}}
	// Preserve onset only when it comes from the exact first pronounced lexical
	// child chosen for this NP. The linearizer never guesses onset from spelling.
	for _, role := range request.Plan.Plan.Order {
		var properties []lexicon.Property
		switch role {
		case "head":
			properties = request.Head.Candidate.Lexeme.Properties
		case "modifier":
			if request.Modifier != nil {
				properties = request.Modifier.Candidate.Lexeme.Properties
			}
		case "determiner":
			continue
		}
		if onset, known := lexicalDeterminerOnset(properties); known {
			form.DeterminerOnset = onset
		}
		break
	}
	sequence := input.Plan.CompiledVariant.Variant.Sequence
	claimed := map[int]bool{}
	claim := func(index int) error {
		if index < 0 || index >= len(sequence) || claimed[index] {
			return executeError(status.OutcomeInvalid, "NP_OUTER_SEQUENCE_POSITION_INVALID", "np_forms."+headSlotID, "outer NP sequence position is missing or claimed twice")
		}
		claimed[index] = true
		return nil
	}
	if err := claim(request.HeadSequenceIndex); err != nil {
		return NPForm{}, err
	}
	outerSlots := map[string]bool{headSlotID: true}
	if request.Modifier != nil {
		if err := claim(request.ModifierSequenceIndex); err != nil {
			return NPForm{}, err
		}
		outerSlots[request.ModifierSlotID] = true
	}
	firstNPIndex := request.HeadSequenceIndex
	if request.Modifier != nil && request.ModifierSequenceIndex < firstNPIndex {
		firstNPIndex = request.ModifierSequenceIndex
	}
	form.EmbeddedDeterminer = request.Determiner != nil && request.Determiner.Embedded
	if request.Determiner != nil && !request.Determiner.Embedded && request.Determiner.SequenceIndex < firstNPIndex {
		firstNPIndex = request.Determiner.SequenceIndex
	}
	for i := firstNPIndex; i <= request.HeadSequenceIndex; i++ {
		if claimed[i] {
			continue
		}
		if request.Determiner == nil || request.Determiner.Embedded || i != request.Determiner.SequenceIndex || sequence[i].Kind != construction.TokenTerminal || !hasNPDeterminerRule(input.PhraseProfile, sequence[i].Terminal) {
			return NPForm{}, executeError(status.OutcomeInvalid, "NP_OUTER_UNEXPECTED_SEQUENCE_TOKEN", "np_forms."+headSlotID, fmt.Sprintf("NP span token %d (%s:%s), determiner=%v", i, sequence[i].Kind, sequence[i].Terminal, request.Determiner))
		}
		if err := claim(i); err != nil {
			return NPForm{}, err
		}
	}
	if request.Determiner != nil {
		// The exact determiner position was included in the contiguous NP span above.
	}
	for i := range sequence {
		if !claimed[i] {
			continue
		}
		token := sequence[i]
		npToken := NPOuterToken{SequenceIndex: i}
		if token.Kind == construction.TokenSlot {
			if !outerSlots[token.SlotID] {
				return NPForm{}, executeError(status.OutcomeInvalid, "NP_OUTER_SLOT_IDENTITY_MISMATCH", "np_forms."+headSlotID, "claimed slot index does not point to the exact NP child slot")
			}
			npToken.SlotID = token.SlotID
			form.OuterSlotIDs = append(form.OuterSlotIDs, token.SlotID)
		} else {
			if request.Determiner == nil || token.Terminal != request.Determiner.Surface || i != request.Determiner.SequenceIndex {
				return NPForm{}, executeError(status.OutcomeInvalid, "NP_OUTER_TERMINAL_IDENTITY_MISMATCH", "np_forms."+headSlotID, "claimed terminal is not the exact selected determiner")
			}
			npToken.Terminal = token.Terminal
			form.OuterSlotIDs = append(form.OuterSlotIDs, "")
		}
		form.OuterTokens = append(form.OuterTokens, npToken)
		form.OuterSequenceIndexes = append(form.OuterSequenceIndexes, i)
	}
	for _, role := range request.Plan.Plan.Order {
		switch role {
		case "head":
			slotForm, ok := formsBySlot[headSlotID]
			if !ok || slotForm.NodeID != request.Head.NodeID || slotForm.LexemeID != request.Head.Candidate.Lexeme.ID || slotForm.SenseID != request.Head.Candidate.Sense.ID || !reflect.DeepEqual(slotForm.Features, request.Head.Features.Values) || !reflect.DeepEqual(slotForm.Provenance, request.Head.Candidate.TransferBinding.Provenance) {
				return NPForm{}, executeError(status.OutcomeIncomplete, "NP_HEAD_FORM_UNAVAILABLE", "np_forms."+headSlotID+".head", "generated head form does not match the exact NP head request")
			}
			copy := cloneSlotForm(slotForm)
			form.Parts = append(form.Parts, NPFormPart{Role: "head", InputID: request.Head.InputID, Kind: construction.PhraseInputLexical, Surface: slotForm.Surface, SlotForm: &copy})
		case "modifier":
			if request.Modifier == nil {
				return NPForm{}, executeError(status.OutcomeInvalid, "NP_PLAN_MODIFIER_MISSING", "np_forms."+headSlotID+".modifier", "NP plan order names an absent modifier")
			}
			slotForm, ok := formsBySlot[request.ModifierSlotID]
			modifierRequest := requests[request.ModifierSlotID]
			modifierSpec, specFound := construction.SlotSpec{}, false
			for _, candidateSpec := range input.Family.SlotSpecs {
				if candidateSpec.ID == request.ModifierSlotID {
					modifierSpec, specFound = candidateSpec, true
					break
				}
			}
			if !ok || !specFound || slotForm.NodeID != request.Modifier.NodeID || slotForm.Role != modifierSpec.ModifierBinding.SemanticRole || slotForm.LexemeID != request.Modifier.Candidate.Lexeme.ID || slotForm.SenseID != request.Modifier.Candidate.Sense.ID || !reflect.DeepEqual(slotForm.Features, request.Modifier.Features.Values) || !reflect.DeepEqual(slotForm.Provenance, modifierRequest.Candidate.TransferBinding.Provenance) || request.Modifier.Candidate.SourceCandidateID != request.Head.Candidate.SourceCandidateID || request.Modifier.Candidate.SemanticID != request.Head.Candidate.SemanticID {
				return NPForm{}, executeError(status.OutcomeIncomplete, "NP_MODIFIER_FORM_UNAVAILABLE", "np_forms."+headSlotID+".modifier", "generated AP form does not match the exact modifier state request")
			}
			copy := cloneSlotForm(slotForm)
			form.Parts = append(form.Parts, NPFormPart{Role: "modifier", InputID: request.Modifier.InputID, Kind: construction.PhraseInputAP, Surface: slotForm.Surface, SlotForm: &copy})
		case "determiner":
			if request.Determiner == nil {
				return NPForm{}, executeError(status.OutcomeInvalid, "NP_PLAN_DETERMINER_MISSING", "np_forms."+headSlotID+".determiner", "NP plan order names an absent determiner")
			}
			part := NPFormPart{Role: "determiner", InputID: request.Determiner.InputID, Kind: construction.PhraseInputLexical, Surface: request.Determiner.Surface}
			if !request.Determiner.Embedded {
				part.TerminalProvenance = input.Plan.CompiledVariant.Variant.Provenance
			}
			form.Parts = append(form.Parts, part)
		default:
			return NPForm{}, executeError(status.OutcomeUnsupported, "NP_PLAN_ROLE_UNSUPPORTED", "np_forms."+headSlotID+".order", "NP template contains an unsupported role")
		}
	}
	return form, nil
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

func cloneSlotForm(form SlotForm) SlotForm {
	form.MeaningPath = append([]meaning.NodeID(nil), form.MeaningPath...)
	form.Features = append([]features.FeatureValue(nil), form.Features...)
	form.Trace = cloneAuxiliaryTrace(form.Trace)
	if form.VPAuxiliary != nil {
		auxiliary := *form.VPAuxiliary
		auxiliary.Auxiliary.Features = append([]features.FeatureValue(nil), auxiliary.Auxiliary.Features...)
		auxiliary.Auxiliary.Trace = cloneAuxiliaryTrace(auxiliary.Auxiliary.Trace)
		auxiliary.Base.Features = append([]features.FeatureValue(nil), auxiliary.Base.Features...)
		auxiliary.Base.Trace = cloneAuxiliaryTrace(auxiliary.Base.Trace)
		if auxiliary.Negation != nil {
			particle := *auxiliary.Negation
			auxiliary.Negation = &particle
		}
		form.VPAuxiliary = &auxiliary
	}
	return form
}

func cloneAuxiliaryTrace(trace []morph.GenerationTraceStep) []morph.GenerationTraceStep {
	cloned := append([]morph.GenerationTraceStep(nil), trace...)
	for index := range cloned {
		if cloned[index].Applied == nil {
			continue
		}
		value := *cloned[index].Applied
		cloned[index].Applied = &value
	}
	return cloned
}

func generateVPAuxiliary(input ExecuteInput, slot construction.SlotSpec, request SlotRequest, budget procedure.Budget) (VPAuxiliaryForm, error) {
	auxiliary := request.VPAuxiliary
	if auxiliary == nil || auxiliary.EventNodeID == "" || auxiliary.EventNodeID != input.Plan.NodeID || auxiliary.Lexeme.ID.String() != "lexeme:en.do.aux" || auxiliary.Lexeme.Language != lexicon.LanguageID(input.Plan.Language) || auxiliary.Lexeme.POS != lexicon.PartOfSpeech("verb") {
		return VPAuxiliaryForm{}, executeError(status.OutcomeInvalid, "VP_AUXILIARY_REQUEST_INVALID", "slots."+slot.ID+".vp_auxiliary", "VP auxiliary requires the declared nonsemantic auxiliary lexeme, target language, and exact event identity")
	}
	if slot.Category != construction.CategoryVP || !slot.IsMainPredicate() || request.NodeID != auxiliary.EventNodeID {
		return VPAuxiliaryForm{}, executeError(status.OutcomeUnsupported, "VP_AUXILIARY_SLOT_UNSUPPORTED", "slots."+slot.ID+".vp_auxiliary", "VP auxiliary requires a VP-category main predicate slot bound to its event")
	}
	if auxiliary.Features.Language != input.Plan.Language || auxiliary.BaseFeatures.Language != input.Plan.Language || request.Candidate.Lexeme.Language != lexicon.LanguageID(input.Plan.Language) {
		return VPAuxiliaryForm{}, executeError(status.OutcomeIncomplete, "VP_AUXILIARY_FEATURE_LANGUAGE_MISMATCH", "slots."+slot.ID+".vp_auxiliary", "auxiliary, lexical base, and selected predicate must use the plan language")
	}
	if err := validateVPAuxiliaryNegation(input, auxiliary); err != nil {
		return VPAuxiliaryForm{}, err
	}
	auxGenerated, err := morph.Generate(auxiliary.Lexeme, auxiliary.PrincipalParts, auxiliary.Features, input.Paradigms, input.Exceptions, budget)
	if err != nil {
		return VPAuxiliaryForm{}, executeWrapError(err)
	}
	auxForm, err := exactlyOneVPAuxiliaryForm(auxGenerated, "auxiliary")
	if err != nil {
		return VPAuxiliaryForm{}, err
	}
	if auxForm.LexemeID != auxiliary.Lexeme.ID.String() {
		return VPAuxiliaryForm{}, executeError(status.OutcomeInvalid, "VP_AUXILIARY_LEXEME_MISMATCH", "slots."+slot.ID+".vp_auxiliary.auxiliary", "morphology output does not identify the requested auxiliary lexeme")
	}
	baseGenerated, err := morph.Generate(request.Candidate.Lexeme, request.PrincipalParts, auxiliary.BaseFeatures, input.Paradigms, input.Exceptions, budget)
	if err != nil {
		return VPAuxiliaryForm{}, executeWrapError(err)
	}
	baseForm, err := exactlyOneVPAuxiliaryForm(baseGenerated, "base")
	if err != nil {
		return VPAuxiliaryForm{}, err
	}
	if baseForm.LexemeID != request.Candidate.Lexeme.ID.String() {
		return VPAuxiliaryForm{}, executeError(status.OutcomeInvalid, "VP_BASE_LEXEME_MISMATCH", "slots."+slot.ID+".vp_auxiliary.base", "morphology output does not identify the selected predicate lexeme")
	}
	for _, token := range input.Plan.CompiledVariant.Variant.Sequence {
		if token.Kind != construction.TokenTerminal {
			continue
		}
		if strings.EqualFold(token.Terminal, auxForm.Surface) {
			return VPAuxiliaryForm{}, executeError(status.OutcomeUnsupported, "VP_AUXILIARY_OUTER_DUPLICATE", "plan.compiled_variant.sequence", "selected variant already emits the generated auxiliary as an outer terminal")
		}
		if auxiliary.NegationParticle != nil && strings.EqualFold(token.Terminal, auxiliary.NegationParticle.Surface) {
			return VPAuxiliaryForm{}, executeError(status.OutcomeUnsupported, "VP_NEGATION_OUTER_DUPLICATE", "plan.compiled_variant.sequence", "selected variant already emits the scoped negation particle as an outer terminal")
		}
	}
	result := VPAuxiliaryForm{
		SlotID: slot.ID, EventNodeID: auxiliary.EventNodeID,
		Auxiliary: VPAuxiliaryMorphForm{
			Surface: auxForm.Surface, LexemeID: auxiliary.Lexeme.ID,
			Features: append([]features.FeatureValue(nil), auxForm.Features...), Trace: append([]morph.GenerationTraceStep(nil), auxGenerated.Trace...),
			Provenance: *auxForm.Provenance,
		},
		Base: VPAuxiliaryMorphForm{
			Surface: baseForm.Surface, LexemeID: request.Candidate.Lexeme.ID,
			Features: append([]features.FeatureValue(nil), baseForm.Features...), Trace: append([]morph.GenerationTraceStep(nil), baseGenerated.Trace...),
			Provenance: *baseForm.Provenance,
		},
	}
	if auxiliary.NegationParticle != nil {
		particle := auxiliary.NegationParticle
		result.Negation = &VPAuxiliaryParticleForm{
			Surface: particle.Surface, OperatorID: particle.OperatorID,
			ScopeNodeID: input.Plan.Construction.Steps[0].SemanticScope.ScopeNodeID, Provenance: particle.Provenance,
		}
	}
	return result, nil
}

func exactlyOneVPAuxiliaryForm(generation morph.GenerationResult, role string) (morph.GeneratedForm, error) {
	switch generation.State {
	case morph.GenerationGenerated:
		if len(generation.Forms) > 1 {
			return morph.GeneratedForm{}, executeError(status.OutcomeAmbiguous, "VP_AUXILIARY_FORM_AMBIGUOUS", "vp_auxiliary."+role, "morphology produced more than one auxiliary or base form")
		}
		if len(generation.Forms) != 1 || generation.Forms[0].Surface == "" || generation.Forms[0].Provenance == nil || len(generation.Trace) == 0 {
			return morph.GeneratedForm{}, executeError(status.OutcomeIncomplete, "VP_AUXILIARY_FORM_INCOMPLETE", "vp_auxiliary."+role, "auxiliary and base forms require one generated surface with morphology provenance")
		}
		if err := generation.Forms[0].Provenance.Validate(); err != nil {
			return morph.GeneratedForm{}, executeError(status.OutcomeIncomplete, "VP_AUXILIARY_PROVENANCE_INVALID", "vp_auxiliary."+role, "generated auxiliary or base form has invalid morphology provenance")
		}
		return generation.Forms[0], nil
	case morph.GenerationBlocked:
		return morph.GeneratedForm{}, executeFromDiagnostics(status.OutcomeUnsupported, "VP_AUXILIARY_MORPHOLOGY_BLOCKED", "vp_auxiliary."+role, "morphology blocks the requested auxiliary or base form", generation.Diagnostics)
	case morph.GenerationIncomplete:
		return morph.GeneratedForm{}, executeFromDiagnostics(status.OutcomeIncomplete, "VP_AUXILIARY_MORPHOLOGY_INCOMPLETE", "vp_auxiliary."+role, "morphology lacks evidence for the requested auxiliary or base form", generation.Diagnostics)
	case morph.GenerationUnavailable, morph.GenerationUnknown:
		return morph.GeneratedForm{}, executeFromDiagnostics(status.OutcomeUnknown, "VP_AUXILIARY_MORPHOLOGY_UNKNOWN", "vp_auxiliary."+role, "morphology cannot resolve the requested auxiliary or base form", generation.Diagnostics)
	default:
		return morph.GeneratedForm{}, executeError(status.OutcomeInvalid, "VP_AUXILIARY_MORPHOLOGY_STATE_INVALID", "vp_auxiliary."+role, "morphology returned an invalid auxiliary or base state")
	}
}

func validateVPAuxiliaryNegation(input ExecuteInput, auxiliary *VPAuxiliaryExecutionRequest) error {
	if len(input.Plan.Construction.Steps) == 0 {
		if auxiliary.NegationParticle != nil {
			return executeError(status.OutcomeInvalid, "VP_NEGATION_OPERATOR_UNAVAILABLE", "slots.vp_auxiliary.negation", "a negation particle requires one exact plan operator scoped to the event")
		}
		return nil
	}
	if len(input.Plan.Construction.Steps) != 1 || !explicitSurfaceNegation(input.Plan, input.Family) {
		return executeError(status.OutcomeUnsupported, "VP_AUXILIARY_OPERATOR_CHAIN_UNSUPPORTED", "plan.construction.steps", "VP auxiliary realization accepts only one exact identity negation operator")
	}
	operator := input.Plan.Construction.Steps[0]
	if operator.SemanticScope.TargetNodeID != auxiliary.EventNodeID || operator.SemanticScope.ScopeNodeID == "" || auxiliary.NegationParticle == nil || auxiliary.NegationParticle.OperatorID != operator.ID || auxiliary.NegationParticle.Provenance != operator.Provenance {
		return executeError(status.OutcomeUnsupported, "VP_NEGATION_SCOPE_MISMATCH", "slots.vp_auxiliary.negation", "negation particle must match the exact operator, event target, scope node, and provenance")
	}
	if !validSlotSurface(auxiliary.NegationParticle.Surface) || auxiliary.NegationParticle.Provenance.Validate() != nil {
		return executeError(status.OutcomeInvalid, "VP_NEGATION_PARTICLE_INVALID", "slots.vp_auxiliary.negation", "negation particle must have a valid surface and operator provenance")
	}
	var scope *meaning.Node
	for index := range input.SourceMeaning.Nodes {
		if input.SourceMeaning.Nodes[index].ID == operator.SemanticScope.ScopeNodeID {
			scope = &input.SourceMeaning.Nodes[index]
			break
		}
	}
	if scope == nil || scope.Kind != meaning.KindScope || scope.Operator != meaning.OperatorNegation || scope.Target != auxiliary.EventNodeID {
		return executeError(status.OutcomeUnsupported, "VP_NEGATION_MEANING_SCOPE_MISMATCH", "source.meaning."+string(operator.SemanticScope.ScopeNodeID), "negation Meaning scope must target the exact auxiliary event")
	}
	return nil
}

func phraseNPSpec(profile construction.PhraseProfile, templateID string) (construction.PhraseTemplate, bool) {
	for _, template := range profile.Templates {
		if template.ID == templateID && template.Kind == construction.PhraseNP {
			return template, true
		}
	}
	return construction.PhraseTemplate{}, false
}

func hasNPDeterminerRule(profile construction.PhraseProfile, surface string) bool {
	for _, rule := range profile.DeterminerForms {
		if rule.Surface == surface {
			return true
		}
	}
	return false
}

func projectedPhraseFeatures(profile *features.LanguageFeatureProfile, template construction.PhraseTemplate, role string, source features.FeatureBundle) features.FeatureBundle {
	allowed := map[features.FeatureID]bool{}
	for _, input := range template.Inputs {
		if input.Role == role {
			for _, id := range input.FeatureIDs {
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
	bundle, _ := features.NewBundle(profile, values)
	return bundle
}

func validateNPDeterminerFeatures(input ExecuteInput, request *NPExecutionRequest, head features.FeatureBundle) error {
	path := "slots." + request.HeadSlotID + ".np.determiner.features"
	if err := features.ValidateBundle(input.FeatureProfile, request.Determiner.Features); err != nil {
		return executeWrapError(err)
	}
	if err := features.ValidateBundle(input.FeatureProfile, head); err != nil {
		return executeWrapError(err)
	}
	if !hasFeatureValue(head, "number") {
		if number, known := knownFeature(head.Values, "noun_number"); known {
			values := append(append([]features.FeatureValue(nil), head.Values...), features.FeatureValue{Feature: "number", State: features.StateKnown, Value: number})
			var err error
			head, err = features.NewBundle(input.FeatureProfile, values)
			if err != nil {
				return executeWrapError(err)
			}
		}
	}
	onset := ""
	firstRole := "head"
	for _, role := range request.Plan.Plan.Order {
		if role != "determiner" {
			firstRole = role
			break
		}
	}
	properties := request.Head.Candidate.Lexeme.Properties
	if firstRole == "modifier" && request.Modifier != nil {
		properties = request.Modifier.Candidate.Lexeme.Properties
	}
	if hasOnsetConstraint(input.PhraseProfile, request.Determiner.Surface) {
		var known bool
		onset, known = lexicalDeterminerOnset(properties)
		if !known {
			return executeError(status.OutcomeIncomplete, "NP_DETERMINER_ONSET_UNKNOWN", path+".onset", "phrase profile onset rule requires evidence from the exact first pronounced lexical candidate")
		}
	}
	match := construction.EvaluateDeterminerForms(input.PhraseProfile, request.Determiner.Surface, head, request.Determiner.Features, onset)
	switch match.Status {
	case construction.DeterminerOK:
		return nil
	case construction.DeterminerIncomplete:
		return executeError(status.OutcomeIncomplete, "NP_DETERMINER_EVIDENCE_INCOMPLETE", path, "determiner form requires known exact head and determiner features")
	case construction.DeterminerAmbiguous:
		return executeError(status.OutcomeAmbiguous, "NP_DETERMINER_FORM_AMBIGUOUS", path, "multiple non-identical phrase profile rules match the exact determiner features")
	default:
		return executeError(status.OutcomeUnsupported, "NP_DETERMINER_FORM_UNSUPPORTED", path, "no phrase profile rule licenses the exact determiner surface and features")
	}
}

func hasFeatureValue(bundle features.FeatureBundle, id features.FeatureID) bool {
	for _, value := range bundle.Values {
		if value.Feature == id {
			return true
		}
	}
	return false
}

func hasOnsetConstraint(profile construction.PhraseProfile, surface string) bool {
	for _, rule := range profile.DeterminerForms {
		if rule.Surface == surface && rule.Onset != "" {
			return true
		}
	}
	return false
}

func meaningNodeForExecute(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func variantSlotIndex(sequence []construction.SequenceToken, slotID string) int {
	found := -1
	for index, token := range sequence {
		if token.Kind == construction.TokenSlot && token.SlotID == slotID {
			if found >= 0 {
				return -1
			}
			found = index
		}
	}
	return found
}

func validateConstructionPlanForSurface(plan RealizationPlan, family construction.ConstructionFamily) error {
	composition := plan.Construction
	if composition.Language != family.Language || composition.SenseID != plan.SenseID || composition.ValencyFrameID != plan.FrameID || composition.InputFamilyID != plan.FamilyID || composition.OutputFamilyID != plan.FamilyID {
		return executeError(status.OutcomeInvalid, "CONSTRUCTION_PLAN_CONTEXT_MISMATCH", "plan.construction", "construction plan language, sense, frame, or family differs from the selected realization")
	}
	if len(composition.Steps) > 0 {
		if !composition.ID.Valid() {
			return executeError(status.OutcomeInvalid, "CONSTRUCTION_PLAN_ID_MISMATCH", "plan.construction.id", "operator construction plan requires a canonical identity")
		}
		if explicitSurfaceNegation(plan, family) || explicitSurfaceQuestion(plan, family) {
			return nil
		}
		return executeError(status.OutcomeUnsupported, "CONSTRUCTION_OPERATOR_NOT_EXECUTABLE", "plan.construction.steps", "only an explicitly declared identity negation has a compiled surface realization")
	}
	if !composition.ID.Valid() || composition.ID != noOpConstructionPlanID(plan.FamilyPlanID, plan.SenseID, plan.FrameID) {
		return executeError(status.OutcomeInvalid, "CONSTRUCTION_PLAN_ID_MISMATCH", "plan.construction.id", "no-op construction plan identity does not match its selected family, sense, and frame")
	}
	return nil
}

func noOpConstructionPlanID(familyPlanID construction.ConstructionPlanID, senseID ids.SenseID, frameID valency.FrameID) construction.ConstructionPlanID {
	var builder strings.Builder
	writeCompositionPart(&builder, "compose-v1")
	writeCompositionPart(&builder, string(familyPlanID))
	writeCompositionPart(&builder, string(familyPlanID))
	writeCompositionPart(&builder, senseID.String())
	writeCompositionPart(&builder, string(frameID))
	writeCompositionPart(&builder, "0")
	digest := sha256.Sum256([]byte(builder.String()))
	return construction.ConstructionPlanID("plan:" + hex.EncodeToString(digest[:]))
}

func writeCompositionPart(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func validateSlotCandidate(candidate TargetCandidate, nodeID meaning.NodeID, node meaning.Node, language features.LanguageID, slotID string) error {
	path := "slots." + slotID + ".candidate"
	if candidate.Outcome != status.OutcomeOK {
		if candidate.Outcome.Valid() && candidate.Outcome != status.OutcomeOK && len(candidate.Diagnostics) > 0 {
			return executeFromDiagnostics(candidate.Outcome, "SLOT_CANDIDATE_NOT_OK", path, "slot lexical candidate is not an OK selection", candidate.Diagnostics)
		}
		return executeError(status.OutcomeInvalid, "INVALID_SLOT_CANDIDATE_OUTCOME", path, "slot lexical candidate must have an OK outcome")
	}
	if candidate.NodeID != nodeID || candidate.Kind != node.Kind || candidate.Lexeme.Language == "" || features.LanguageID(candidate.Lexeme.Language) != language || candidate.Sense.LexemeID != candidate.Lexeme.ID || candidate.Sense.ID != candidate.TransferBinding.SenseID || candidate.TransferBinding.CoreRef.String() != string(node.Concept) || !candidate.TransferBinding.TransferAuthorized() {
		return executeError(status.OutcomeUnsupported, "SLOT_TRANSFER_NOT_LICENSED", path, "slot candidate does not provide an approved equivalent transfer for this exact Meaning node")
	}
	if len(candidate.Diagnostics) > status.MaxDiagnostics {
		return executeError(status.OutcomeResourceLimit, "SLOT_DIAGNOSTIC_LIMIT", path+".diagnostics", "slot candidate diagnostics exceed the supported bound")
	}
	for index, diagnostic := range candidate.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return executeError(status.OutcomeInvalid, "INVALID_SLOT_CANDIDATE_DIAGNOSTIC", fmt.Sprintf("%s.diagnostics[%d]", path, index), err.Error())
		}
	}
	return nil
}

func validateHardAgreement(family construction.ConstructionFamily, slots map[string]construction.SlotSpec, requests map[string]SlotRequest) error {
	for _, constraint := range family.Constraints {
		if constraint.Kind != construction.ConstraintAgreement || constraint.Strength != construction.ConstraintHard {
			continue
		}
		controller, controllerPresent := requests[constraint.ControllerSlot]
		target, targetPresent := requests[constraint.TargetSlot]
		if !controllerPresent || !targetPresent {
			return executeError(status.OutcomeIncomplete, "AGREEMENT_SLOT_UNAVAILABLE", "family.constraints."+constraint.ID, "hard agreement requires explicit controller and target slot feature bundles")
		}
		if _, exists := slots[constraint.ControllerSlot]; !exists {
			return executeError(status.OutcomeInvalid, "AGREEMENT_CONTROLLER_UNKNOWN", "family.constraints."+constraint.ID, "hard agreement controller slot is absent")
		}
		if _, exists := slots[constraint.TargetSlot]; !exists {
			return executeError(status.OutcomeInvalid, "AGREEMENT_TARGET_UNKNOWN", "family.constraints."+constraint.ID, "hard agreement target slot is absent")
		}
		for _, feature := range constraint.Features {
			left, leftOK := knownFeature(controller.Features.Values, feature)
			right, rightOK := knownFeature(target.Features.Values, feature)
			if !leftOK || !rightOK {
				return executeError(status.OutcomeIncomplete, "AGREEMENT_FEATURE_UNKNOWN", "family.constraints."+constraint.ID+".features."+string(feature), "hard agreement feature values must be explicitly known on both slots")
			}
			if left != right {
				return executeError(status.OutcomeUnsupported, "AGREEMENT_FEATURE_MISMATCH", "family.constraints."+constraint.ID+".features."+string(feature), "explicit slot feature values violate hard agreement")
			}
		}
	}
	return nil
}

func validateSlotCasePolicies(frame valency.ValencyFrame, orderedSlots []construction.SlotSpec, requests map[string]SlotRequest, family construction.ConstructionFamily, variant construction.Variant) error {
	frameArguments := make(map[meaning.Role]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		frameArguments[argument.Role] = argument
	}
	for _, slot := range orderedSlots {
		if slot.ArgumentBinding == nil {
			continue
		}
		role := slot.ArgumentBinding.SemanticRole
		argument, exists := frameArguments[role]
		if !exists {
			return executeError(status.OutcomeInvalid, "FAMILY_ROLE_NOT_IN_FRAME", "family.slot_specs."+slot.ID, "slot semantic role is absent from the selected valency frame")
		}
		request, supplied := requests[slot.ID]
		if supplied && (slot.Adposition != "" || argument.Adposition != "") {
			if slot.Adposition != "" && argument.Adposition != "" && slot.Adposition != argument.Adposition {
				return executeError(status.OutcomeUnsupported, "ADPOSITION_POLICY_CONFLICT", "family.slot_specs."+slot.ID+".adposition", "family slot adposition conflicts with the selected frame argument adposition")
			}
			adposition := slot.Adposition
			if adposition == "" {
				adposition = argument.Adposition
			}
			if !construction.VariantEmitsAdposition(variant, family, slot.ID, adposition) {
				return executeError(status.OutcomeUnsupported, "ADPOSITION_SEQUENCE_UNAVAILABLE", "family.slot_specs."+slot.ID, "an adpositional argument requires its exact adposition terminal before the slot in the selected construction")
			}
		}
		if slot.Case != "" && argument.Case != "" && slot.Case != argument.Case {
			return executeError(status.OutcomeUnsupported, "CASE_POLICY_CONFLICT", "family.slot_specs."+slot.ID+".case", "family slot case conflicts with the selected frame argument case")
		}
		expectedCase := slot.Case
		if expectedCase == "" {
			expectedCase = argument.Case
		}
		if expectedCase == "" {
			continue
		}
		if !supplied {
			continue
		}
		if request.CaseFeature == "" || !validPlanField(string(request.CaseFeature)) {
			return executeError(status.OutcomeIncomplete, "CASE_FEATURE_MAPPING_UNAVAILABLE", "slots."+slot.ID+".case_feature", "case policy requires an explicit mapping to a language-local morphological case feature")
		}
		value, known := knownFeature(request.Features.Values, request.CaseFeature)
		if !known {
			return executeError(status.OutcomeIncomplete, "CASE_FEATURE_UNKNOWN", "slots."+slot.ID+".features."+string(request.CaseFeature), "case feature must be explicitly known in the slot feature bundle")
		}
		if string(value) != expectedCase {
			return executeError(status.OutcomeUnsupported, "CASE_FEATURE_MISMATCH", "slots."+slot.ID+".features."+string(request.CaseFeature), "explicit morphological case differs from the licensed frame/family case")
		}
	}
	return nil
}

func validateMeaningRoleCoverage(predicate meaning.Node, slots []construction.SlotSpec) error {
	bound := make(map[meaning.Role]struct{}, len(slots))
	for _, slot := range slots {
		if slot.ArgumentBinding != nil {
			bound[slot.ArgumentBinding.SemanticRole] = struct{}{}
		}
	}
	seen := make(map[meaning.Role]struct{}, len(predicate.Participants))
	for _, participant := range predicate.Participants {
		if _, duplicate := seen[participant.Role]; duplicate {
			return executeError(status.OutcomeAmbiguous, "MEANING_ROLE_AMBIGUOUS", "meaning."+string(predicate.ID)+".participants", "source Meaning supplies multiple targets for one semantic role")
		}
		seen[participant.Role] = struct{}{}
		if _, exists := bound[participant.Role]; !exists {
			return executeError(status.OutcomeUnsupported, "MEANING_ROLE_UNBOUND_BY_FAMILY", "family.slot_specs", "source Meaning contains a semantic role with no licensed family slot")
		}
	}
	return nil
}

func knownFeature(values []features.FeatureValue, feature features.FeatureID) (features.FeatureValueID, bool) {
	for _, value := range values {
		if value.Feature == feature && value.State == features.StateKnown && value.Value != "" {
			return value.Value, true
		}
	}
	return "", false
}

func validateVariantConditions(conditions []construction.Condition, requests map[string]SlotRequest, information map[meaning.Role]construction.InformationValue) error {
	for index, condition := range conditions {
		path := fmt.Sprintf("variant.conditions[%d]", index)
		switch condition.Kind {
		case construction.ConditionFeatureEquals:
			request, exists := requests[condition.SlotID]
			if !exists {
				return executeError(status.OutcomeIncomplete, "VARIANT_FEATURE_SLOT_UNAVAILABLE", path+".slot_id", "variant condition requires an explicitly realized slot feature")
			}
			value, known := knownFeature(request.Features.Values, condition.Feature)
			if !known {
				return executeError(status.OutcomeIncomplete, "VARIANT_FEATURE_UNKNOWN", path+".feature", "variant condition feature is absent or unknown in the explicit slot feature bundle")
			}
			if value != condition.Value {
				return executeError(status.OutcomeUnsupported, "VARIANT_FEATURE_CONDITION_MISMATCH", path+".value", "explicit slot feature does not satisfy the selected variant condition")
			}
		case construction.ConditionInformationStructure:
			value, exists := information[condition.Role]
			if !exists {
				return executeError(status.OutcomeIncomplete, "VARIANT_INFORMATION_STRUCTURE_UNKNOWN", path+".role", "variant condition requires explicit information-structure evidence for this semantic role")
			}
			if !validInformationValue(value) {
				return executeError(status.OutcomeInvalid, "INVALID_INFORMATION_STRUCTURE_EVIDENCE", "information_structure."+string(condition.Role), "information-structure evidence has an unsupported value")
			}
			if value != condition.Info {
				return executeError(status.OutcomeUnsupported, "VARIANT_INFORMATION_STRUCTURE_MISMATCH", path+".information_structure", "explicit information-structure evidence does not satisfy the selected variant condition")
			}
		default:
			return executeError(status.OutcomeUnsupported, "VARIANT_CONDITION_UNSUPPORTED", path+".kind", "surface execution cannot evaluate this variant condition kind")
		}
	}
	return nil
}

func validInformationValue(value construction.InformationValue) bool {
	switch value {
	case construction.InformationTopic, construction.InformationFocus, construction.InformationGiven, construction.InformationNew:
		return true
	default:
		return false
	}
}

func containsRole(roles []meaning.Role, target meaning.Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func requirePlanFeatureConstraints(constraints []FeatureConstraint, path []meaning.NodeID, bundle features.FeatureBundle) error {
	for _, constraint := range constraints {
		if !equalNodePath(constraint.Path, path) {
			continue
		}
		value, exists := knownFeature(bundle.Values, constraint.Feature)
		if !exists {
			return executeError(status.OutcomeIncomplete, "PLAN_FEATURE_UNAVAILABLE", "feature_constraints."+string(constraint.Feature), "explicit plan feature constraint is absent or unknown in the slot feature bundle")
		}
		if value != constraint.Value {
			return executeError(status.OutcomeUnsupported, "PLAN_FEATURE_CONFLICT", "feature_constraints."+string(constraint.Feature), "slot feature bundle conflicts with the explicit plan feature constraint")
		}
	}
	return nil
}

func participantByRole(node meaning.Node, role meaning.Role) (meaning.Participant, int) {
	var found meaning.Participant
	count := 0
	for _, participant := range node.Participants {
		if participant.Role == role {
			found = participant
			count++
		}
	}
	return found, count
}

func cloneExecutionFamily(family construction.ConstructionFamily) construction.ConstructionFamily {
	result := family
	result.SlotSpecs = append([]construction.SlotSpec(nil), family.SlotSpecs...)
	for index, slot := range result.SlotSpecs {
		if slot.ArgumentBinding != nil {
			binding := *slot.ArgumentBinding
			result.SlotSpecs[index].ArgumentBinding = &binding
		}
		if slot.ModifierBinding != nil {
			binding := *slot.ModifierBinding
			result.SlotSpecs[index].ModifierBinding = &binding
		}
	}
	result.Constraints = make([]construction.FamilyConstraint, len(family.Constraints))
	for index, constraint := range family.Constraints {
		result.Constraints[index] = constraint
		result.Constraints[index].Features = append([]features.FeatureID(nil), constraint.Features...)
	}
	result.Operators = append([]construction.OperatorID(nil), family.Operators...)
	result.Capabilities = append([]construction.Capability(nil), family.Capabilities...)
	return result
}

func executeError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &ExecuteError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

func executeLimit(resource string, value, maximum int) error {
	return executeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("surface execution %s limit exceeded: %d > %d", resource, value, maximum))
}

func executeWrapError(err error) error {
	if err == nil {
		return nil
	}
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		result := typed.Result()
		if result.Status.Valid() && len(result.Diagnostics) > 0 {
			return &ExecuteError{result: status.Result{Status: result.Status, Diagnostics: append([]status.Diagnostic(nil), result.Diagnostics...)}}
		}
	}
	return executeError(status.OutcomeError, "SURFACE_EXECUTION_ERROR", "execution", err.Error())
}

func executeFromDiagnostics(kind status.OutcomeKind, code, path, message string, diagnostics []status.Diagnostic) error {
	result := status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}
	for _, diagnostic := range diagnostics {
		if len(result.Diagnostics) >= status.MaxDiagnostics {
			break
		}
		if diagnostic.Validate() == nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic)
		}
	}
	return &ExecuteError{result: result}
}

func equalNodePath(left, right []meaning.NodeID) bool {
	return reflect.DeepEqual(left, right)
}

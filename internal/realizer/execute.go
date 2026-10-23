package realizer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
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
}

// ExecuteInput contains one validated plan, its source Meaning, the exact
// family/frame license, and caller supplied morphology inputs for each slot.
type ExecuteInput struct {
	Plan          RealizationPlan
	SourceMeaning meaning.Meaning
	Core          *core.ConceptGraph
	Family        construction.ConstructionFamily
	Frame         valency.ValencyFrame
	Slots         map[string]SlotRequest
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
}

// ExecuteResult preserves generated forms in family slot order. Optional
// omitted slots have no SlotForm; the family remains available to linearize.
type ExecuteResult struct {
	Outcome     status.OutcomeKind              `json:"outcome"`
	Family      construction.ConstructionFamily `json:"family"`
	Slots       []SlotForm                      `json:"slots"`
	Diagnostics []status.Diagnostic             `json:"diagnostics,omitempty"`
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
	if len(plan.Construction.Steps) > 0 {
		return ExecuteResult{}, executeError(status.OutcomeUnsupported, "SURFACE_OPERATOR_UNSUPPORTED", "plan.construction.steps", "surface execution cannot yet emit composed operator morphology or segments")
	}
	if err := input.Family.Validate(); err != nil {
		return ExecuteResult{}, executeWrapError(err)
	}
	if input.Family.ID != plan.FamilyID || plan.CompiledVariant.FamilyID != plan.FamilyID || plan.CompiledVariant.PlanID != plan.FamilyPlanID {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_FAMILY_MISMATCH", "family", "supplied family identity or plan version differs from the selected plan")
	}
	compiled, err := construction.ValidateVariantForFrame(plan.CompiledVariant.Variant, input.Family, input.Frame)
	if err != nil {
		return ExecuteResult{}, executeWrapError(err)
	}
	if compiled.PlanID != plan.FamilyPlanID || compiled.FamilyID != plan.FamilyID || compiled.Variant.ID != plan.VariantID || !reflect.DeepEqual(compiled, plan.CompiledVariant) {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_VARIANT_MISMATCH", "plan.compiled_variant", "compiled variant does not match the supplied family and plan")
	}
	if input.Frame.ID != plan.FrameID || input.Frame.SenseID != plan.SenseID {
		return ExecuteResult{}, executeError(status.OutcomeInvalid, "PLAN_FRAME_MISMATCH", "frame", "supplied valency frame differs from the selected plan sense or frame")
	}
	if err := validateNoOpConstructionPlan(plan, input.Family); err != nil {
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
		if slot.Category == construction.CategoryVP && slot.ArgumentBinding == nil {
			predicateSlots++
		}
	}
	if predicateSlots != 1 {
		return ExecuteResult{}, executeError(status.OutcomeUnsupported, "PREDICATE_SLOT_CARDINALITY", "family.slot_specs", "surface execution requires exactly one unbound predicate VP slot")
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
			if slot.Category == construction.CategoryVP && slot.ArgumentBinding == nil {
				return ExecuteResult{}, executeError(status.OutcomeIncomplete, "PREDICATE_SLOT_UNAVAILABLE", "slots."+slot.ID, "the selected event/state predicate requires an explicit lexicalization request")
			}
			if slot.ArgumentBinding != nil {
				_, count := participantByRole(predicate, slot.ArgumentBinding.SemanticRole)
				if count > 1 {
					return ExecuteResult{}, executeError(status.OutcomeAmbiguous, "MEANING_ROLE_AMBIGUOUS", "meaning."+string(plan.NodeID)+".participants", "source Meaning supplies multiple targets for one semantic role")
				}
				if count == 1 {
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
		if slot.ArgumentBinding != nil {
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
			if slot.Category != construction.CategoryVP || req.NodeID != plan.NodeID {
				return ExecuteResult{}, executeError(status.OutcomeUnsupported, "UNBOUND_NON_PREDICATE_SLOT", "family.slot_specs."+slot.ID, "only the predicate VP slot may be unbound to a semantic argument role")
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
		})
	}
	if len(result.Slots) == 0 {
		return ExecuteResult{}, executeError(status.OutcomeIncomplete, "NO_GENERATED_SLOTS", "slots", "no required or explicitly supplied optional slot produced a surface form")
	}
	return result, nil
}

func validateNoOpConstructionPlan(plan RealizationPlan, family construction.ConstructionFamily) error {
	composition := plan.Construction
	if composition.Language != family.Language || composition.SenseID != plan.SenseID || composition.ValencyFrameID != plan.FrameID || composition.InputFamilyID != plan.FamilyID || composition.OutputFamilyID != plan.FamilyID {
		return executeError(status.OutcomeInvalid, "CONSTRUCTION_PLAN_CONTEXT_MISMATCH", "plan.construction", "no-op construction plan language, sense, frame, or family differs from the selected realization")
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

package realizer

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	defaultLinearizeMaxTextBytes = 64 * 1024
	maxLinearizeTextBytes        = 1 << 20
	defaultLinearizeMaxParts     = 4096
	maxLinearizeParts            = 16384
	maxLinearizedNPInputs        = 3
)

// LinearizeBudget bounds the final surface string and its source-traced parts.
type LinearizeBudget struct {
	MaxTextBytes int
	MaxParts     int
}

// DefaultLinearizeBudget returns finite limits for one surface composition.
func DefaultLinearizeBudget() LinearizeBudget {
	return LinearizeBudget{MaxTextBytes: defaultLinearizeMaxTextBytes, MaxParts: defaultLinearizeMaxParts}
}

func (budget LinearizeBudget) Validate() error {
	if budget.MaxTextBytes <= 0 || budget.MaxTextBytes > maxLinearizeTextBytes || budget.MaxParts <= 0 || budget.MaxParts > maxLinearizeParts {
		return linearizeError(status.OutcomeInvalid, "INVALID_LINEARIZE_BUDGET", "budget", "linearization limits must be positive and within package ceilings")
	}
	return nil
}

// LinearizedResult contains final text and an exact ordered partition into
// source-traced slot, terminal, and separator parts.
type LinearizedResult struct {
	Text        string              `json:"text"`
	Parts       []LinearizedPart    `json:"parts"`
	Segments    []Segment           `json:"segments"`
	Outcome     status.OutcomeKind  `json:"outcome"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// LinearizationError carries a typed outcome and bounded diagnostic.
type LinearizationError struct{ result status.Result }

func (err *LinearizationError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "surface linearization error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed error result.
func (err *LinearizationError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

// Linearize places generated slot forms in the exact compiled variant order.
// It handles only declared slots, fixed literal terminals, and explicit
// separators; it does not interpret or repair an unrecognized marker.
func Linearize(plan RealizationPlan, family construction.ConstructionFamily, slots []SlotForm, information map[meaning.Role]construction.InformationValue, budget LinearizeBudget) (LinearizedResult, error) {
	return LinearizeWithNPForms(plan, family, slots, nil, information, budget)
}

// LinearizeWithNPForms emits prepared NP parts through their nested phrase
// order while preserving the legacy flat path when no NP forms are supplied.
func LinearizeWithNPForms(plan RealizationPlan, family construction.ConstructionFamily, slots []SlotForm, npForms []NPForm, information map[meaning.Role]construction.InformationValue, budget LinearizeBudget) (LinearizedResult, error) {
	return linearizeWithNPProfiles(plan, family, slots, npForms, construction.PhraseProfile{}, nil, information, budget)
}

// LinearizeWithNPProfiles validates prepared NP forms against the exact
// language phrase and feature profiles used to compile and execute them.
func LinearizeWithNPProfiles(plan RealizationPlan, family construction.ConstructionFamily, slots []SlotForm, npForms []NPForm, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, information map[meaning.Role]construction.InformationValue, budget LinearizeBudget) (LinearizedResult, error) {
	return linearizeWithNPProfiles(plan, family, slots, npForms, phraseProfile, featureProfile, information, budget)
}

func linearizeWithNPProfiles(plan RealizationPlan, family construction.ConstructionFamily, slots []SlotForm, npForms []NPForm, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, information map[meaning.Role]construction.InformationValue, budget LinearizeBudget) (LinearizedResult, error) {
	if budget == (LinearizeBudget{}) {
		budget = DefaultLinearizeBudget()
	}
	if err := budget.Validate(); err != nil {
		return LinearizedResult{}, err
	}
	if plan.Outcome == "" || !plan.Outcome.Valid() || len(plan.Diagnostics) > status.MaxDiagnostics {
		return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "INVALID_REALIZATION_PLAN", "plan", "plan outcome or diagnostics are invalid")
	}
	if plan.ID == "" || plan.ID != realizationPlanID(plan) {
		return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "MUTATED_REALIZATION_PLAN", "plan.id", "realization plan identity does not match its content")
	}
	for index, diagnostic := range plan.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "INVALID_PLAN_DIAGNOSTIC", fmt.Sprintf("plan.diagnostics[%d]", index), err.Error())
		}
	}
	if plan.Outcome != status.OutcomeOK && plan.Outcome != status.OutcomeIncomplete {
		return LinearizedResult{}, linearizeError(plan.Outcome, "PLAN_OUTCOME_NOT_REALIZABLE", "plan.outcome", "only ok or explicitly incomplete plans may be linearized")
	}
	if len(plan.Construction.Steps) != 0 && !explicitSurfaceNegation(plan, family) && !explicitSurfaceQuestion(plan, family) {
		return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "CONSTRUCTION_OPERATORS_NOT_LINEARIZED", "plan.construction.steps", "the selected family has no explicit surface realization for this operator chain")
	}
	compiled := plan.CompiledVariant
	if family.ID != plan.FamilyID || family.ID != compiled.FamilyID || family.ID != plan.Construction.OutputFamilyID || compiled.PlanID != plan.FamilyPlanID || compiled.Variant.ID != plan.VariantID {
		return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "LINEARIZATION_FAMILY_MISMATCH", "family", "family does not match the selected plan and compiled variant")
	}
	if compiled.Variant.DirectionPolicy != construction.DirectionBoth && compiled.Variant.DirectionPolicy != construction.DirectionRealize {
		return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "VARIANT_NOT_LICENSED_FOR_REALIZATION", "plan.compiled_variant", "selected variant is not licensed for realization")
	}
	verified, err := construction.CompileVariant(compiled.Variant, family)
	if err != nil {
		return LinearizedResult{}, wrapLinearizeConstructionError(err)
	}
	if verified.PlanID != compiled.PlanID || verified.FamilyID != compiled.FamilyID || !sameSequence(verified.FirstAllowed, compiled.FirstAllowed) || !sameSequence(verified.LastAllowed, compiled.LastAllowed) {
		return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "STALE_COMPILED_VARIANT", "plan.compiled_variant", "compiled variant boundaries or family plan do not match the current family")
	}
	if len(slots) > budget.MaxParts || len(slots) > maxLinearizeParts {
		return LinearizedResult{}, linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "slots", "slot forms exceed the linearization part limit")
	}
	if len(information) > len(plan.Roles) {
		return LinearizedResult{}, linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "information_structure", "information-structure evidence exceeds the selected plan role limit")
	}
	for role, value := range information {
		if !validLinearizationRole(role) || !containsRole(plan.Roles, role) {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "UNLICENSED_INFORMATION_STRUCTURE_ROLE", "information_structure."+string(role), "information-structure evidence must identify a valid role in the selected plan")
		}
		if !validLinearizationInformationValue(value) {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "INVALID_INFORMATION_STRUCTURE_EVIDENCE", "information_structure."+string(role), "information-structure evidence has an unsupported value")
		}
	}

	familySlots := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	familyRoles := make(map[meaning.Role]struct{}, len(family.SlotSpecs))
	predicateSlot := ""
	predicateSlotCount := 0
	for _, slot := range family.SlotSpecs {
		familySlots[slot.ID] = slot
		if slot.ArgumentBinding != nil {
			familyRoles[slot.ArgumentBinding.SemanticRole] = struct{}{}
		}
		if slot.ArgumentBinding == nil && slot.ModifierBinding == nil {
			if !slot.IsMainPredicate() {
				return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "UNBOUND_NON_PREDICATE_SLOT", "family.slot_specs."+slot.ID, "only an unbound predicate VP/AP slot may be linearized")
			}
			predicateSlot = slot.ID
			predicateSlotCount++
		}
	}
	if predicateSlotCount != 1 {
		return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "PREDICATE_SLOT_CARDINALITY", "family.slot_specs", "surface linearization requires exactly one unbound predicate VP/AP slot")
	}
	seenRoles := make(map[meaning.Role]struct{}, len(plan.Roles))
	for _, role := range plan.Roles {
		if !validLinearizationRole(role) {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "INVALID_PLAN_ROLE", "plan.roles", "selected plan contains an invalid semantic role")
		}
		if _, duplicate := seenRoles[role]; duplicate {
			return LinearizedResult{}, linearizeError(status.OutcomeAmbiguous, "DUPLICATE_PLAN_ROLE", "plan.roles", "selected plan contains multiple targets for one semantic role")
		}
		seenRoles[role] = struct{}{}
		if _, bound := familyRoles[role]; !bound {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "MEANING_ROLE_UNBOUND", "family.slot_specs", "selected Meaning contains a role with no licensed family slot")
		}
	}
	sequenceSlots := make(map[string]struct{}, len(compiled.Variant.Sequence))
	for _, token := range compiled.Variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	forms := make(map[string]SlotForm, len(slots))
	textBytes := 0
	for index, form := range slots {
		path := fmt.Sprintf("slots[%d]", index)
		if _, exists := familySlots[form.SlotID]; !exists {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "UNKNOWN_SLOT_FORM", path+".slot_id", "slot form does not belong to the licensed construction family")
		}
		if _, exists := sequenceSlots[form.SlotID]; !exists {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "UNSEQUENCED_SLOT_FORM", path+".slot_id", "slot form is not present in the licensed variant sequence")
		}
		if _, exists := forms[form.SlotID]; exists {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "DUPLICATE_SLOT_FORM", path+".slot_id", "each construction slot may have at most one generated form")
		}
		if !validSlotSurface(form.Surface) {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "INVALID_SLOT_SURFACE", path+".surface", "slot surface must be non-empty valid UTF-8 without boundary whitespace or control characters")
		}
		spec := familySlots[form.SlotID]
		if spec.Category == construction.CategoryPP || spec.Adposition != "" {
			if spec.Adposition == "" || !construction.VariantEmitsAdposition(compiled.Variant, family, spec.ID, spec.Adposition) {
				return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "ADPOSITION_SEQUENCE_UNAVAILABLE", path+".slot_id", "an adpositional argument requires its exact adposition terminal before the slot in the selected construction")
			}
		}
		expectedRole := meaning.Role("")
		if spec.ArgumentBinding != nil {
			expectedRole = spec.ArgumentBinding.SemanticRole
		} else if spec.ModifierBinding != nil {
			expectedRole = spec.ModifierBinding.SemanticRole
		}
		if form.Role != expectedRole {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "SLOT_ROLE_LICENSE_MISMATCH", path+".role", "slot form role does not match the family slot binding")
		}
		if spec.ArgumentBinding != nil && !containsRole(plan.Roles, spec.ArgumentBinding.SemanticRole) {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "SLOT_ROLE_ABSENT_FROM_MEANING", path+".role", "slot form supplies a role absent from the selected Meaning")
		}
		if spec.IsMainPredicate() && (form.NodeID != plan.NodeID || form.LexemeID != plan.LexemeID || form.SenseID != plan.SenseID) {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "PREDICATE_SLOT_PLAN_MISMATCH", path, "predicate slot form must match the selected plan node, lexeme, and sense")
		}
		if spec.ModifierBinding != nil && form.NodeID == plan.NodeID {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "MODIFIER_SLOT_PLAN_MISMATCH", path, "modifier form must identify its own state rather than the main predicate")
		}
		if !containsNode(form.MeaningPath, form.NodeID) {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "SLOT_MEANING_PATH_MISMATCH", path+".meaning_path", "slot source node must occur in its declared Meaning path")
		}
		if err := form.Provenance.Validate(); err != nil || form.LexemeID.String() == "" || form.SenseID.String() == "" {
			return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "MISSING_SLOT_TRACE", path, "slot form requires source, lexeme, and sense identity")
		}
		textBytes += len(form.Surface)
		if textBytes > budget.MaxTextBytes {
			return LinearizedResult{}, linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", "slot surfaces exceed the linearization text-byte limit")
		}
		forms[form.SlotID] = form
	}
	for _, slot := range family.SlotSpecs {
		_, inVariant := sequenceSlots[slot.ID]
		_, hasForm := forms[slot.ID]
		activeRole := slot.ArgumentBinding != nil && containsRole(plan.Roles, slot.ArgumentBinding.SemanticRole)
		required := slot.ID == predicateSlot || slot.Optionality == construction.SlotRequired || slot.Cardinality.Minimum > 0 || activeRole
		if !required {
			continue
		}
		if !inVariant {
			if activeRole {
				if _, licensed := construction.ImplicitParticipantSlot(family, compiled.Variant, slot.ID); licensed {
					continue
				}
			}
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "ACTIVE_SLOT_OMITTED_BY_VARIANT", "variant.sequence."+slot.ID, "licensed variant omits a slot required by the selected Meaning")
		}
		if !hasForm {
			return LinearizedResult{}, linearizeError(status.OutcomeIncomplete, "REQUIRED_SLOT_FORM_MISSING", "slots."+slot.ID, "source Meaning requires a generated form for this construction slot")
		}
	}
	if err := validateLinearizationConditions(compiled.Variant.Conditions, forms, information); err != nil {
		return LinearizedResult{}, err
	}
	if len(npForms) > 0 && (featureProfile == nil || phraseProfile.Language != plan.Language || featureProfile.Language() != plan.Language) {
		return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "NP_LINEARIZATION_PROFILE_UNAVAILABLE", "np_forms.profile", "NP forms require the exact target phrase and feature profiles")
	}
	npByStart, npClaimed, err := validateLinearizedNPForms(plan, family, compiled.Variant, forms, npForms, phraseProfile, featureProfile, budget)
	if err != nil {
		return LinearizedResult{}, err
	}

	result := LinearizedResult{Text: "", Parts: []LinearizedPart{}, Outcome: plan.Outcome, Diagnostics: append([]status.Diagnostic(nil), plan.Diagnostics...)}
	lastWasOpening := false
	var output strings.Builder
	appendPart := func(part LinearizedPart) error {
		if len(result.Parts) >= budget.MaxParts {
			return linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.parts", "linearized output exceeds the part limit")
		}
		if len(part.Text) > budget.MaxTextBytes-output.Len() {
			return linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.text_bytes", "linearized output exceeds the text-byte limit")
		}
		output.WriteString(part.Text)
		result.Parts = append(result.Parts, part)
		return nil
	}
	appendSeparator := func() error {
		return appendPart(LinearizedPart{Text: " ", Kind: SegmentSeparator, ConstructionID: plan.FamilyID.String()})
	}
	for index, token := range compiled.Variant.Sequence {
		path := fmt.Sprintf("variant.sequence[%d]", index)
		if npForm, startsNP := npByStart[index]; startsNP {
			for _, part := range npForm.Parts {
				if output.Len() > 0 && !lastWasOpening {
					if err := appendSeparator(); err != nil {
						return LinearizedResult{}, err
					}
				}
				linearPart, err := linearizedNPPart(plan, npForm, part)
				if err != nil {
					return LinearizedResult{}, err
				}
				if err := appendPart(linearPart); err != nil {
					return LinearizedResult{}, err
				}
				lastWasOpening = false
			}
			continue
		}
		if npClaimed[index] {
			continue
		}
		if token.Kind == construction.TokenSlot {
			form, exists := forms[token.SlotID]
			if !exists { // Only a declared optional slot may be absent.
				continue
			}
			if output.Len() > 0 && !lastWasOpening {
				if err := appendSeparator(); err != nil {
					return LinearizedResult{}, err
				}
			}
			part := LinearizedPart{
				Text: form.Surface, Kind: SegmentLexical, SourceID: form.Provenance.ID,
				NodeID: string(form.NodeID), LexemeID: form.LexemeID.String(), SenseID: form.SenseID.String(),
				ConstructionID: plan.FamilyID.String(),
			}
			if len(form.Trace) > 0 {
				part.ProcedureID = form.Trace[len(form.Trace)-1].ProcedureID
			}
			if err := appendPart(part); err != nil {
				return LinearizedResult{}, err
			}
			lastWasOpening = false
			continue
		}
		if token.Kind != construction.TokenTerminal {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "UNSUPPORTED_SEQUENCE_MARKER", path+".kind", "compiled variant contains an unsupported sequence marker")
		}
		class, literalErr := classifyTerminal(token.Terminal)
		if literalErr != nil {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "UNSUPPORTED_TERMINAL_MARKER", path+".terminal", literalErr.Error())
		}
		if class == terminalOpeningPunctuation {
			if output.Len() > 0 && !lastWasOpening {
				if err := appendSeparator(); err != nil {
					return LinearizedResult{}, err
				}
			}
		} else if class == terminalWord {
			if output.Len() > 0 && !lastWasOpening {
				if err := appendSeparator(); err != nil {
					return LinearizedResult{}, err
				}
			}
		} else if output.Len() == 0 {
			return LinearizedResult{}, linearizeError(status.OutcomeUnsupported, "PUNCTUATION_WITHOUT_LEFT_CONTEXT", path+".terminal", "closing punctuation requires preceding realized text")
		}
		if err := appendPart(LinearizedPart{Text: token.Terminal, Kind: SegmentTerminal, SourceID: compiled.Variant.Provenance.ID, ConstructionID: plan.FamilyID.String()}); err != nil {
			return LinearizedResult{}, err
		}
		lastWasOpening = class == terminalOpeningPunctuation
	}
	for index := range result.Parts {
		capitalized, hasLetter := capitalizeInitialLetter(result.Parts[index].Text)
		if hasLetter {
			result.Parts[index].Text = capitalized
			break
		}
	}
	output.Reset()
	for _, part := range result.Parts {
		output.WriteString(part.Text)
	}
	result.Text = output.String()
	if !utf8.ValidString(result.Text) {
		return LinearizedResult{}, linearizeError(status.OutcomeInvalid, "INVALID_LINEARIZED_UTF8", "text", "linearized output is not valid UTF-8")
	}
	segments, err := BuildSegments(result.Text, result.Parts, SegmentBudget{
		MaxTextBytes: budget.MaxTextBytes,
		MaxSegments:  budget.MaxParts,
	})
	if err != nil {
		return LinearizedResult{}, wrapLinearizeConstructionError(err)
	}
	result.Segments = segments
	return result, nil
}

func validateLinearizedNPForms(plan RealizationPlan, family construction.ConstructionFamily, variant construction.Variant, forms map[string]SlotForm, npForms []NPForm, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, budget LinearizeBudget) (map[int]NPForm, map[int]bool, error) {
	npSlotCount := 0
	for _, slot := range family.SlotSpecs {
		if slot.Category == construction.CategoryNP {
			npSlotCount++
		}
	}
	if len(npForms) > budget.MaxParts || len(npForms) > npSlotCount {
		return nil, nil, linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "np_forms", "NP forms exceed the selected family NP slots or linearization part limit")
	}
	byStart := make(map[int]NPForm, len(npForms))
	claimedIndexes := make(map[int]bool)
	claimedSlots := make(map[string]bool)
	byHead := make(map[string]bool, len(npForms))
	for index, np := range npForms {
		path := fmt.Sprintf("np_forms[%d]", index)
		if np.SlotID == "" || np.HeadNodeID == "" || np.VariantID != plan.VariantID || np.VariantProvenance != variant.Provenance || np.Plan.Plan.Kind != construction.PhraseNP || np.Plan.Plan.Language != plan.Language || np.Plan.Plan.TemplateID != "np-attributive" || np.Plan.Plan.ID == "" {
			return nil, nil, linearizeError(status.OutcomeInvalid, "INVALID_NP_FORM_CONTEXT", path, "NP form must identify one exact head, target language, NP plan and selected outer variant")
		}
		if _, duplicate := byHead[np.SlotID]; duplicate {
			return nil, nil, linearizeError(status.OutcomeInvalid, "DUPLICATE_NP_FORM_HEAD", path+".slot_id", "one outer NP head has multiple NP forms")
		}
		byHead[np.SlotID] = true
		head, exists := forms[np.SlotID]
		if !exists || head.NodeID != np.HeadNodeID {
			return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_HEAD_SLOT_MISMATCH", path+".head_node_id", "NP form does not match the generated exact head slot form")
		}
		if len(np.OuterSequenceIndexes) == 0 || len(np.OuterSequenceIndexes) != len(np.OuterTokens) || len(np.OuterSlotIDs) != len(np.OuterTokens) || len(np.OuterSequenceIndexes) > 3 {
			return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_CLAIM_SHAPE", path+".outer_sequence_indexes", "NP form requires a bounded exact list of outer slots and optional determiner")
		}
		start := np.OuterSequenceIndexes[0]
		if _, duplicate := byStart[start]; duplicate {
			return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_START_OVERLAP", path+".outer_sequence_indexes", "multiple NP forms start at the same outer sequence position")
		}
		if start < 0 || start+len(np.OuterSequenceIndexes) > len(variant.Sequence) {
			return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_RANGE_INVALID", path+".outer_sequence_indexes", "NP form outer positions exceed the selected variant")
		}
		for offset, seqIndex := range np.OuterSequenceIndexes {
			if seqIndex != start+offset || np.OuterTokens[offset].SequenceIndex != seqIndex || claimedIndexes[seqIndex] {
				return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_OVERLAP_OR_GAP", path+".outer_sequence_indexes", "NP outer claims must be contiguous and disjoint")
			}
			token := variant.Sequence[seqIndex]
			outer := np.OuterTokens[offset]
			switch token.Kind {
			case construction.TokenSlot:
				if outer.SlotID != token.SlotID || outer.Terminal != "" || np.OuterSlotIDs[offset] != token.SlotID || claimedSlots[token.SlotID] {
					return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_SLOT_MISMATCH", path+".outer_tokens", "NP slot claim does not match one unused exact outer slot")
				}
				claimedSlots[token.SlotID] = true
			case construction.TokenTerminal:
				if outer.Terminal != token.Terminal || outer.SlotID != "" || np.OuterSlotIDs[offset] != "" {
					return nil, nil, linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_TERMINAL_MISMATCH", path+".outer_tokens", "NP terminal claim does not match the exact outer variant terminal")
				}
			default:
				return nil, nil, linearizeError(status.OutcomeUnsupported, "NP_FORM_OUTER_TOKEN_UNSUPPORTED", path+".outer_tokens", "NP form claims an unsupported outer token")
			}
			claimedIndexes[seqIndex] = true
		}
		if err := validateNPFormParts(family, variant, forms, np, phraseProfile, featureProfile, path); err != nil {
			return nil, nil, err
		}
		byStart[start] = np
	}
	if len(npForms) > 0 {
		for slotID, form := range forms {
			spec := findLinearizerSlot(family, slotID)
			if spec != nil && spec.Category == construction.CategoryNP && !byHead[slotID] {
				return nil, nil, linearizeError(status.OutcomeIncomplete, "NP_FORM_MISSING", "np_forms."+slotID, "an NP SlotForm has no prepared NPForm")
			}
			_ = form
		}
	}
	return byStart, claimedIndexes, nil
}

func validateNPFormParts(family construction.ConstructionFamily, variant construction.Variant, forms map[string]SlotForm, np NPForm, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, path string) error {
	if len(np.Plan.Plan.Inputs) > maxLinearizedNPInputs || len(np.Plan.Plan.Order) > maxLinearizedNPInputs || len(np.Parts) > maxLinearizedNPInputs {
		return linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", path+".plan", "NP plans are limited to the supported head, modifier, and determiner inputs")
	}
	if len(np.Parts) != len(np.Plan.Plan.Order) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_PART_COUNT_MISMATCH", path+".parts", "NP form parts must cover every plan role exactly once")
	}
	inputByRole := make(map[string]construction.PhraseInput, len(np.Plan.Plan.Inputs))
	profileVersion := ""
	for _, input := range np.Plan.Plan.Inputs {
		if _, exists := inputByRole[input.Role]; exists {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_PLAN_DUPLICATE_INPUT", path+".plan.inputs", "NP plan has duplicate input roles")
		}
		inputByRole[input.Role] = input
		if profileVersion == "" {
			profileVersion = string(input.Features.ProfileVersion)
		} else if profileVersion != string(input.Features.ProfileVersion) {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_PROFILE_VERSION_MISMATCH", path+".plan.inputs", "NP inputs must use one exact language feature profile version")
		}
	}
	if len(inputByRole) < 1 || len(inputByRole) > 3 || len(np.Plan.Plan.Order) != len(inputByRole) || len(np.Plan.Plan.Children) != 0 || len(np.Plan.Plan.Features.Values) != 0 || inputByRole["head"].ID == "" || !inputByRole["head"].Required {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_PLAN_SHAPE_INVALID", path+".plan", "NP plan must contain only head, optional modifier, and optional determiner inputs")
	}
	if err := construction.ValidatePhraseProfile(phraseProfile, featureProfile); err != nil {
		return linearizeError(status.OutcomeInvalid, "NP_PHRASE_PROFILE_INVALID", path+".profile", err.Error())
	}
	if np.Plan.Plan.Language != phraseProfile.Language || featureProfile.Language() != np.Plan.Plan.Language || profileVersion != string(featureProfile.Version()) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_PROFILE_MISMATCH", path+".plan.inputs", "NP plan inputs must use the exact supplied language feature profile")
	}
	compiled, err := construction.CompileNP(np.Plan.Plan, phraseProfile, featureProfile)
	if err != nil || !reflect.DeepEqual(compiled, np.Plan) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_COMPILED_PLAN_MISMATCH", path+".plan", "NP plan agreements and structure must match recompilation against the exact phrase profile")
	}
	if err := validateLinearizedNPAgreements(np, phraseProfile, path); err != nil {
		return err
	}
	if err := validateLinearizedNPDeterminer(np, phraseProfile, featureProfile, path); err != nil {
		return err
	}
	for _, input := range np.Plan.Plan.Inputs {
		if input.Features.Language != np.Plan.Plan.Language || input.Features.ProfileVersion == "" {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_PROFILE_MISMATCH", path+".plan.inputs."+input.Role, "NP input features must carry the selected target language profile")
		}
	}
	outerSlotSet := make(map[string]bool)
	terminalCount := 0
	for _, token := range np.OuterTokens {
		if token.SlotID != "" {
			outerSlotSet[token.SlotID] = true
		} else if token.Terminal != "" {
			terminalCount++
		}
	}
	partRoles := map[string]bool{}
	expectedOuterSlots := map[string]bool{np.SlotID: true}
	partByRole := make(map[string]NPFormPart, len(np.Parts))
	for i, role := range np.Plan.Plan.Order {
		part := np.Parts[i]
		input, exists := inputByRole[role]
		if !exists || part.Role != role || part.InputID != input.ID || part.Kind != input.Kind || partRoles[role] {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_PART_PLAN_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "NP part identity/order must match its compiled plan input")
		}
		partRoles[role] = true
		partByRole[role] = part
		if part.SlotForm != nil && len(part.SlotForm.Trace) > morph.MaxGenerationTrace {
			return linearizeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("%s.parts[%d].slot_form.trace", path, i), "nested NP morphology trace exceeds the generation trace limit")
		}
		switch role {
		case "head":
			if input.ID != "head-input" || input.Kind != construction.PhraseInputLexical || !input.Required || part.SlotForm == nil || part.SlotForm.SlotID != np.SlotID || part.Surface != part.SlotForm.Surface || !reflect.DeepEqual(*part.SlotForm, forms[np.SlotID]) || !outerSlotSet[np.SlotID] {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_HEAD_PART_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "head part must preserve the exact generated outer NP SlotForm")
			}
			if !slotFeaturesCover(input.Features, part.SlotForm.Features) {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_HEAD_FEATURE_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "head input features are not preserved by its generated slot form")
			}
		case "modifier":
			if input.ID != "modifier-input" || input.Kind != construction.PhraseInputAP || input.Required || part.SlotForm == nil || part.Surface != part.SlotForm.Surface || !outerSlotSet[part.SlotForm.SlotID] {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_MODIFIER_PART_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "modifier part must preserve an exact generated outer AP SlotForm")
			}
			spec := findLinearizerSlot(family, part.SlotForm.SlotID)
			if spec == nil || spec.ModifierBinding == nil || spec.ModifierBinding.TargetSlot != np.SlotID || !reflect.DeepEqual(*part.SlotForm, forms[part.SlotForm.SlotID]) {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_MODIFIER_BINDING_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "modifier part is not the exact generated state bound to this NP head")
			}
			if !slotFeaturesCover(input.Features, part.SlotForm.Features) {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_MODIFIER_FEATURE_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "modifier input features are not preserved by its generated slot form")
			}
			expectedOuterSlots[part.SlotForm.SlotID] = true
		case "determiner":
			if input.ID != "determiner-input" || input.Kind != construction.PhraseInputLexical || input.Required || part.SlotForm != nil || part.Kind != construction.PhraseInputLexical || terminalCount != 1 {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_DETERMINER_PART_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "determiner part must match one exact terminal claim")
			}
			if part.TerminalProvenance != variant.Provenance {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_DETERMINER_PROVENANCE_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "determiner terminal must retain the selected variant provenance")
			}
			found := false
			for _, token := range np.OuterTokens {
				if token.Terminal == part.Surface && token.Terminal != "" {
					found = true
				}
			}
			if !found {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_DETERMINER_TERMINAL_MISMATCH", fmt.Sprintf("%s.parts[%d]", path, i), "determiner part has no exact outer terminal claim")
			}
		default:
			return linearizeError(status.OutcomeUnsupported, "NP_FORM_ROLE_UNSUPPORTED", fmt.Sprintf("%s.parts[%d]", path, i), "NP form contains an unsupported plan role")
		}
	}
	if len(expectedOuterSlots) != len(outerSlotSet) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_SLOT_COVERAGE_MISMATCH", path+".outer_tokens", "outer slot claims must equal the exact head and optional modifier parts")
	}
	for slotID := range expectedOuterSlots {
		if !outerSlotSet[slotID] {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_SLOT_COVERAGE_MISMATCH", path+".outer_tokens", "outer slot claims omit a selected NP child slot")
		}
	}
	if terminalCount > 1 || (terminalCount == 0 && partRoles["determiner"]) || (terminalCount == 1 && !partRoles["determiner"]) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_DETERMINER_COVERAGE_MISMATCH", path+".outer_tokens", "determiner terminal and determiner part must be consumed exactly once")
	}
	outerIndex := 0
	for _, role := range np.Plan.Plan.Order {
		if role == "determiner" {
			part := partByRole[role]
			if outerIndex >= len(np.OuterTokens) || np.OuterTokens[outerIndex].Terminal != part.Surface || np.OuterTokens[outerIndex].SlotID != "" {
				return linearizeError(status.OutcomeInvalid, "NP_FORM_DETERMINER_ORDER_MISMATCH", path+".outer_tokens", "determiner must occupy the first phrase position licensed by the NP plan")
			}
			outerIndex++
			continue
		}
		part := partByRole[role]
		if part.SlotForm == nil || outerIndex >= len(np.OuterTokens) || np.OuterTokens[outerIndex].SlotID != part.SlotForm.SlotID || np.OuterTokens[outerIndex].Terminal != "" {
			return linearizeError(status.OutcomeInvalid, "NP_FORM_CHILD_ORDER_MISMATCH", path+".outer_tokens", "outer slot order must match exact NP plan child order")
		}
		outerIndex++
	}
	if outerIndex != len(np.OuterTokens) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_OUTER_CLAIM_EXTRA", path+".outer_tokens", "outer claims contain tokens absent from the compiled NP plan")
	}
	return nil
}

func validateLinearizedNPAgreements(np NPForm, phraseProfile construction.PhraseProfile, path string) error {
	inputs := make(map[string]construction.PhraseInput, len(np.Plan.Plan.Inputs))
	for _, input := range np.Plan.Plan.Inputs {
		inputs[input.Role] = input
	}
	var template *construction.PhraseTemplate
	for i := range phraseProfile.Templates {
		if phraseProfile.Templates[i].ID == np.Plan.Plan.TemplateID && phraseProfile.Templates[i].Kind == construction.PhraseNP {
			template = &phraseProfile.Templates[i]
			break
		}
	}
	if template == nil {
		return linearizeError(status.OutcomeIncomplete, "NP_FORM_AGREEMENT_TEMPLATE_UNAVAILABLE", path+".plan.template_id", "selected NP phrase template is unavailable for agreement validation")
	}
	expected := construction.ApplicablePhraseAgreements(*template, np.Plan.Plan.Inputs)
	if !sameNPAgreementSet(np.Plan.Plan.Agreements, expected) {
		return linearizeError(status.OutcomeInvalid, "NP_FORM_AGREEMENT_LIST_MISMATCH", path+".plan.agreements", "NP plan agreements do not match the selected template declarations")
	}
	for _, agreement := range expected {
		controllerRole := ""
		targetRole := ""
		for role, input := range inputs {
			if input.ID == agreement.ControllerInput {
				controllerRole = role
			}
			if input.ID == agreement.TargetInput {
				targetRole = role
			}
		}
		controller, controllerOK := inputs[controllerRole]
		target, targetOK := inputs[targetRole]
		if !controllerOK || !targetOK {
			return linearizeError(status.OutcomeIncomplete, "NP_FORM_AGREEMENT_INPUT_MISSING", path+".plan.agreements", "declared agreement requires both exact phrase inputs")
		}
		left, leftOK := knownFeature(controller.Features.Values, agreement.Feature)
		right, rightOK := knownFeature(target.Features.Values, agreement.Feature)
		if !leftOK || !rightOK {
			return linearizeError(status.OutcomeIncomplete, "NP_FORM_AGREEMENT_UNKNOWN", path+".plan.agreements."+string(agreement.Feature), "declared agreement requires known values on both exact phrase inputs")
		}
		if left != right {
			return linearizeError(status.OutcomeUnsupported, "NP_FORM_AGREEMENT_MISMATCH", path+".plan.agreements."+string(agreement.Feature), "known input values do not satisfy the selected template agreement")
		}
	}
	return nil
}

func validateLinearizedNPDeterminer(np NPForm, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, path string) error {
	for _, part := range np.Parts {
		if part.Role != "determiner" {
			continue
		}
		var head, determiner features.FeatureBundle
		for _, input := range np.Plan.Plan.Inputs {
			switch input.Role {
			case "head":
				head = input.Features
			case "determiner":
				determiner = input.Features
			}
		}
		if err := features.ValidateBundle(featureProfile, head); err != nil {
			return linearizeError(status.OutcomeInvalid, "NP_HEAD_FEATURES_INVALID", path+".plan.inputs.head", err.Error())
		}
		if err := features.ValidateBundle(featureProfile, determiner); err != nil {
			return linearizeError(status.OutcomeInvalid, "NP_DETERMINER_FEATURES_INVALID", path+".plan.inputs.determiner", err.Error())
		}
		match := construction.EvaluateDeterminerForms(phraseProfile, part.Surface, head, determiner, np.DeterminerOnset)
		switch match.Status {
		case construction.DeterminerOK:
			return nil
		case construction.DeterminerIncomplete:
			return linearizeError(status.OutcomeIncomplete, "NP_DETERMINER_EVIDENCE_INCOMPLETE", path+".determiner", "determiner licensing requires known feature or onset evidence")
		case construction.DeterminerAmbiguous:
			return linearizeError(status.OutcomeAmbiguous, "NP_DETERMINER_FORM_AMBIGUOUS", path+".determiner", "multiple non-identical phrase profile rules license this determiner")
		default:
			return linearizeError(status.OutcomeUnsupported, "NP_DETERMINER_FORM_UNSUPPORTED", path+".determiner", "no phrase profile rule licenses the determiner for the selected NP features and onset")
		}
	}
	return nil
}

func findLinearizerSlot(family construction.ConstructionFamily, id string) *construction.SlotSpec {
	for i := range family.SlotSpecs {
		if family.SlotSpecs[i].ID == id {
			return &family.SlotSpecs[i]
		}
	}
	return nil
}

func slotFeaturesCover(bundle features.FeatureBundle, values []features.FeatureValue) bool {
	for _, expected := range bundle.Values {
		found := false
		for _, value := range values {
			if value.Feature == expected.Feature && value.State == expected.State && value.Value == expected.Value && reflect.DeepEqual(value.Variable, expected.Variable) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func linearizedNPPart(plan RealizationPlan, np NPForm, part NPFormPart) (LinearizedPart, error) {
	if part.Role == "determiner" {
		class, err := classifyTerminal(part.Surface)
		if err != nil || class != terminalWord {
			return LinearizedPart{}, linearizeError(status.OutcomeInvalid, "NP_DETERMINER_TERMINAL_INVALID", "np_forms."+np.SlotID+".determiner", "NP determiner must be a licensed word terminal")
		}
		return LinearizedPart{Text: part.Surface, Kind: SegmentTerminal, SourceID: part.TerminalProvenance.ID, ConstructionID: plan.FamilyID.String(), NPPlanID: string(np.Plan.Plan.ID), NPRole: part.Role}, nil
	}
	if part.SlotForm == nil {
		return LinearizedPart{}, linearizeError(status.OutcomeInvalid, "NP_FORM_SLOT_PART_MISSING", "np_forms."+np.SlotID+"."+part.Role, "lexical NP part has no generated SlotForm")
	}
	form := part.SlotForm
	if !validSlotSurface(form.Surface) {
		return LinearizedPart{}, linearizeError(status.OutcomeInvalid, "NP_FORM_SLOT_SURFACE_INVALID", "np_forms."+np.SlotID+"."+part.Role, "NP part surface is invalid")
	}
	linearPart := LinearizedPart{Text: form.Surface, Kind: SegmentLexical, SourceID: form.Provenance.ID, NodeID: string(form.NodeID), LexemeID: form.LexemeID.String(), SenseID: form.SenseID.String(), ConstructionID: plan.FamilyID.String(), MorphTrace: append([]morph.GenerationTraceStep(nil), form.Trace...), NPPlanID: string(np.Plan.Plan.ID), NPRole: part.Role}
	if len(form.Trace) > 0 {
		linearPart.ProcedureID = form.Trace[len(form.Trace)-1].ProcedureID
	}
	return linearPart, nil
}

func capitalizeInitialLetter(value string) (string, bool) {
	for offset, character := range value {
		if !unicode.IsLetter(character) {
			continue
		}
		upper := unicode.ToUpper(character)
		if upper == character {
			return value, true
		}
		_, size := utf8.DecodeRuneInString(value[offset:])
		return value[:offset] + string(upper) + value[offset+size:], true
	}
	return value, false
}

func validateLinearizationConditions(conditions []construction.Condition, slots map[string]SlotForm, information map[meaning.Role]construction.InformationValue) error {
	for index, condition := range conditions {
		path := fmt.Sprintf("variant.conditions[%d]", index)
		switch condition.Kind {
		case construction.ConditionFeatureEquals:
			form, exists := slots[condition.SlotID]
			if !exists {
				return linearizeError(status.OutcomeIncomplete, "VARIANT_FEATURE_SLOT_UNAVAILABLE", path+".slot_id", "variant condition requires an explicitly realized slot feature")
			}
			value, known, featureErr := linearizationConditionFeature(form.Features, condition.Feature)
			if featureErr != "" {
				return linearizeError(status.OutcomeInvalid, "INVALID_SLOT_FEATURE_EVIDENCE", path+".feature", featureErr)
			}
			if !known {
				return linearizeError(status.OutcomeIncomplete, "VARIANT_FEATURE_UNKNOWN", path+".feature", "variant condition feature is absent or unknown in the explicit slot feature bundle")
			}
			if value != condition.Value {
				return linearizeError(status.OutcomeUnsupported, "VARIANT_FEATURE_CONDITION_MISMATCH", path+".value", "explicit slot feature does not satisfy the selected variant condition")
			}
		case construction.ConditionInformationStructure:
			value, exists := information[condition.Role]
			if !exists {
				return linearizeError(status.OutcomeIncomplete, "VARIANT_INFORMATION_STRUCTURE_UNKNOWN", path+".role", "variant condition requires explicit information-structure evidence for this semantic role")
			}
			if !validLinearizationInformationValue(value) {
				return linearizeError(status.OutcomeInvalid, "INVALID_INFORMATION_STRUCTURE_EVIDENCE", "information_structure."+string(condition.Role), "information-structure evidence has an unsupported value")
			}
			if value != condition.Info {
				return linearizeError(status.OutcomeUnsupported, "VARIANT_INFORMATION_STRUCTURE_MISMATCH", path+".information_structure", "explicit information-structure evidence does not satisfy the selected variant condition")
			}
		default:
			return linearizeError(status.OutcomeUnsupported, "VARIANT_CONDITION_UNSUPPORTED", path+".kind", "surface linearizer cannot evaluate this variant condition kind")
		}
	}
	return nil
}

func linearizationConditionFeature(values []features.FeatureValue, target features.FeatureID) (features.FeatureValueID, bool, string) {
	var found *features.FeatureValue
	for index := range values {
		if values[index].Feature != target {
			continue
		}
		if found != nil {
			return "", false, "slot feature evidence contains duplicate values for the condition feature"
		}
		value := values[index]
		found = &value
	}
	if found == nil {
		return "", false, ""
	}
	switch found.State {
	case features.StateKnown:
		if found.Value == "" || found.Variable != nil {
			return "", false, "known slot feature must have a value and no variable reference"
		}
		return found.Value, true, ""
	case features.StateAbsent:
		if found.Value != "" || found.Variable != nil {
			return "", false, "absent slot feature cannot carry a value or variable reference"
		}
		return "", false, ""
	case features.StateUnknown:
		if found.Value != "" {
			return "", false, "unknown slot feature cannot carry a known value"
		}
		return "", false, ""
	default:
		return "", false, "slot feature state is unsupported"
	}
}

func validLinearizationInformationValue(value construction.InformationValue) bool {
	switch value {
	case construction.InformationTopic, construction.InformationFocus, construction.InformationGiven, construction.InformationNew:
		return true
	default:
		return false
	}
}

func validLinearizationRole(role meaning.Role) bool {
	value := string(role)
	if value == "" || len(value) > construction.MaxFamilyTextBytes || strings.TrimSpace(value) != value || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

type terminalClass uint8

const (
	terminalWord terminalClass = iota
	terminalClosingPunctuation
	terminalOpeningPunctuation
)

func classifyTerminal(value string) (terminalClass, error) {
	if value == "" || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return 0, fmt.Errorf("terminal is empty, invalid UTF-8, or has boundary whitespace")
	}
	if strings.ContainsAny(value, "@<>${}") {
		return 0, fmt.Errorf("marker-like terminal syntax is not supported")
	}
	allPunctuation := true
	allWord := true
	previousSpace := true
	for _, r := range value {
		if unicode.IsControl(r) {
			return 0, fmt.Errorf("control characters are not supported in terminals")
		}
		if unicode.IsSpace(r) {
			if previousSpace {
				return 0, fmt.Errorf("terminal contains repeated or boundary whitespace")
			}
			previousSpace = true
			allPunctuation = false
			continue
		}
		previousSpace = false
		if !unicode.IsPunct(r) {
			allPunctuation = false
		}
		if !(unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsDigit(r) || strings.ContainsRune("'’ʼ-‐‑", r)) {
			allWord = false
		}
	}
	if allPunctuation {
		r, _ := utf8.DecodeRuneInString(value)
		if strings.ContainsRune("([{“‘«‹", r) {
			return terminalOpeningPunctuation, nil
		}
		return terminalClosingPunctuation, nil
	}
	if allWord && !previousSpace {
		return terminalWord, nil
	}
	return 0, fmt.Errorf("terminal must be a Unicode word literal or punctuation-only literal")
}

func validSlotSurface(value string) bool {
	if value == "" || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func containsNode(path []meaning.NodeID, node meaning.NodeID) bool {
	for _, candidate := range path {
		if candidate == node {
			return true
		}
	}
	return false
}

func wrapLinearizeConstructionError(err error) error {
	var constructionErr interface{ Result() status.Result }
	if errors.As(err, &constructionErr) {
		result := constructionErr.Result()
		return &LinearizationError{result: result}
	}
	return linearizeError(status.OutcomeInvalid, "INVALID_COMPILED_VARIANT", "plan.compiled_variant", err.Error())
}

func linearizeError(kind status.OutcomeKind, code, path, message string) error {
	return &LinearizationError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{
		Code: code, Path: path, Message: message, Severity: status.SeverityError,
	}}}}
}

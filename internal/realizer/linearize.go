package realizer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	defaultLinearizeMaxTextBytes = 64 * 1024
	maxLinearizeTextBytes        = 1 << 20
	defaultLinearizeMaxParts     = 4096
	maxLinearizeParts            = 16384
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
	if len(plan.Construction.Steps) != 0 && !explicitSurfaceNegation(plan, family) {
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

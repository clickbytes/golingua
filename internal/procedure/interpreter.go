package procedure

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxPlanBytes          = 64 << 10
	MaxOperations         = 128
	MaxStems              = 64
	MaxInputCodepoints    = 2_048
	MaxOutputCodepoints   = 1_024
	MaxOutputBytes        = 4_096
	MaxConditionsPerStep  = 32
	MaxTokenBytes         = 128
	MaxOperationTextBytes = 1_024
)

type Budget struct {
	MaxOperations       int
	MaxStems            int
	MaxInputCodepoints  int
	MaxOutputCodepoints int
	MaxOutputBytes      int
	MaxPlanBytes        int
}

func DefaultBudget() Budget {
	return Budget{MaxOperations: MaxOperations, MaxStems: MaxStems, MaxInputCodepoints: MaxInputCodepoints, MaxOutputCodepoints: MaxOutputCodepoints, MaxOutputBytes: MaxOutputBytes, MaxPlanBytes: MaxPlanBytes}
}

func (budget Budget) Validate() error {
	if budget.MaxOperations <= 0 || budget.MaxOperations > MaxOperations || budget.MaxStems <= 0 || budget.MaxStems > MaxStems || budget.MaxInputCodepoints <= 0 || budget.MaxInputCodepoints > MaxInputCodepoints || budget.MaxOutputCodepoints <= 0 || budget.MaxOutputCodepoints > MaxOutputCodepoints || budget.MaxOutputBytes <= 0 || budget.MaxOutputBytes > MaxOutputBytes || budget.MaxPlanBytes <= 0 || budget.MaxPlanBytes > MaxPlanBytes {
		return invalid("INVALID_PROCEDURE_BUDGET", "budget", "procedure limits must be positive and cannot exceed package ceilings")
	}
	return nil
}

type ProcedureError struct{ result status.Result }

func (failure *ProcedureError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "procedure execution failed"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *ProcedureError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func invalid(code, path, message string) error {
	return operationError(status.OutcomeInvalid, code, path, message)
}

func unknown(code, path, message string) error {
	return operationError(status.OutcomeUnknown, code, path, message)
}

func incomplete(code, path, message string) error {
	return operationError(status.OutcomeIncomplete, code, path, message)
}

func unsupported(code, path, message string) error {
	return operationError(status.OutcomeUnsupported, code, path, message)
}

func resourceLimit(name string, value, maximum int) error {
	return operationError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+name, fmt.Sprintf("procedure %s limit exceeded: %d > %d", name, value, maximum))
}

func operationError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &ProcedureError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// ValidateOperations checks the closed operation union, feature guards and
// finite plan bounds without executing or evaluating any data as code.
func ValidateOperations(steps []Operation, budget Budget) error {
	if err := budget.Validate(); err != nil {
		return err
	}
	if len(steps) == 0 {
		return invalid("EMPTY_PROCEDURE", "steps", "procedure requires at least one operation")
	}
	if len(steps) > budget.MaxOperations {
		return resourceLimit("operations", len(steps), budget.MaxOperations)
	}
	hasStemSelected := false
	for index, step := range steps {
		path := fmt.Sprintf("steps[%d]", index)
		if !step.Kind.Valid() {
			return unsupported("UNSUPPORTED_OPERATION", path+".kind", "operation kind is not implemented by this interpreter")
		}
		if len(step.When) > MaxConditionsPerStep {
			return resourceLimit("feature_conditions", len(step.When), MaxConditionsPerStep)
		}
		if step.Kind == OperationSelectStem {
			// Selecting a new explicit diagnostic stem is allowed at any point.
			// It also establishes the current string for subsequent operations.
			hasStemSelected = true
		} else if !hasStemSelected {
			return invalid("STEM_NOT_SELECTED", path, "a stem must be selected before text operations")
		}
		seen := make(map[features.FeatureID]struct{}, len(step.When))
		for _, condition := range step.When {
			if !validToken(string(condition.Feature)) || !validToken(string(condition.Value)) {
				return invalid("INVALID_FEATURE_CONDITION", path+".when", "feature condition IDs must be bounded lowercase tokens")
			}
			if _, exists := seen[condition.Feature]; exists {
				return invalid("DUPLICATE_FEATURE_CONDITION", path+".when", "feature dimension is repeated in the operation guard")
			}
			seen[condition.Feature] = struct{}{}
		}
		switch step.Kind {
		case OperationSelectStem:
			if !validToken(string(step.StemID)) || step.Text != "" || step.Suffix != "" || step.Replacement != nil || step.From != "" || step.To != nil {
				return invalid("INVALID_SELECT_STEM", path, "select_stem requires only a valid stem_id and optional feature guards")
			}
		case OperationAppend, OperationPrepend:
			if !validOperationText(step.Text, false) || step.StemID != "" || step.Suffix != "" || step.Replacement != nil || step.From != "" || step.To != nil {
				return invalid("INVALID_TEXT_OPERATION", path, "append/prepend requires only non-empty bounded text and optional guards")
			}
		case OperationStripSuffix:
			if !validOperationText(step.Suffix, false) || step.StemID != "" || step.Text != "" || step.Replacement != nil || step.From != "" || step.To != nil {
				return invalid("INVALID_STRIP_SUFFIX", path, "strip_suffix requires only a non-empty bounded suffix and optional guards")
			}
		case OperationReplaceSuffix:
			if !validOperationText(step.Suffix, false) || step.Replacement == nil || !validOperationText(*step.Replacement, true) || step.StemID != "" || step.Text != "" || step.From != "" || step.To != nil {
				return invalid("INVALID_REPLACE_SUFFIX", path, "replace_suffix requires a suffix and an explicit bounded replacement")
			}
		case OperationAlternation:
			if !validOperationText(step.From, false) || step.To == nil || !validOperationText(*step.To, true) || step.StemID != "" || step.Text != "" || step.Suffix != "" || step.Replacement != nil {
				return invalid("INVALID_ALTERNATION", path, "alternation requires one exact bounded from/to mapping")
			}
		}
	}
	encoded, err := json.Marshal(steps)
	if err != nil {
		return invalid("INVALID_PROCEDURE_JSON", "steps", "procedure cannot be serialized")
	}
	if len(encoded) > budget.MaxPlanBytes {
		return resourceLimit("plan_bytes", len(encoded), budget.MaxPlanBytes)
	}
	return nil
}

// ExecPlan executes a bounded declarative plan against explicit principal
// stems and feature values. It returns no partial form on any failure.
func ExecPlan(stems map[StemID]string, steps []Operation, bundle features.FeatureBundle, budget Budget) (ExecutionResult, error) {
	if err := ValidateOperations(steps, budget); err != nil {
		return ExecutionResult{}, err
	}
	if err := validateFeatureBundle(bundle); err != nil {
		return ExecutionResult{}, err
	}
	if len(stems) > budget.MaxStems {
		return ExecutionResult{}, resourceLimit("stems", len(stems), budget.MaxStems)
	}
	inputCodepoints := 0
	for id, form := range stems {
		if !validToken(string(id)) {
			return ExecutionResult{}, invalid("INVALID_STEM_ID", "stems", "stem ID is malformed")
		}
		if !validOperationText(form, false) {
			return ExecutionResult{}, invalid("INVALID_STEM_FORM", "stems."+string(id), "principal part is empty, invalid UTF-8, contains controls or exceeds its bound")
		}
		inputCodepoints += utf8.RuneCountInString(form)
		if inputCodepoints > budget.MaxInputCodepoints {
			return ExecutionResult{}, resourceLimit("input_codepoints", inputCodepoints, budget.MaxInputCodepoints)
		}
	}
	current := ""
	hasCurrent := false
	trace := make([]StepTrace, 0, len(steps))
	for index, step := range steps {
		applies, err := checkConditions(step.When, bundle)
		if err != nil {
			return ExecutionResult{}, err
		}
		if !applies {
			trace = append(trace, StepTrace{Index: index, Operation: step.Kind, StemID: step.StemID, Applied: false, OutputSize: len(current)})
			continue
		}
		if step.Kind != OperationSelectStem && !hasCurrent {
			return ExecutionResult{}, invalid("STEM_NOT_SELECTED", fmt.Sprintf("steps[%d]", index), "a stem must be selected before text operations")
		}
		switch step.Kind {
		case OperationSelectStem:
			form, exists := stems[step.StemID]
			if !exists {
				return ExecutionResult{}, unknown("UNKNOWN_STEM", "stems."+string(step.StemID), "required principal part was not supplied")
			}
			current, hasCurrent = form, true
		case OperationAppend:
			current += step.Text
		case OperationPrepend:
			current = step.Text + current
		case OperationStripSuffix:
			if !strings.HasSuffix(current, step.Suffix) {
				return ExecutionResult{}, unknown("SUFFIX_MISMATCH", fmt.Sprintf("steps[%d].suffix", index), "selected stem does not end with the required suffix")
			}
			current = strings.TrimSuffix(current, step.Suffix)
		case OperationReplaceSuffix:
			if !strings.HasSuffix(current, step.Suffix) {
				return ExecutionResult{}, unknown("SUFFIX_MISMATCH", fmt.Sprintf("steps[%d].suffix", index), "selected stem does not end with the required suffix")
			}
			current = strings.TrimSuffix(current, step.Suffix) + *step.Replacement
		case OperationAlternation:
			position := strings.Index(current, step.From)
			if position < 0 {
				return ExecutionResult{}, unknown("ALTERNATION_MISMATCH", fmt.Sprintf("steps[%d].from", index), "selected stem does not contain the declared alternation source")
			}
			current = current[:position] + *step.To + current[position+len(step.From):]
		}
		if len(current) > budget.MaxOutputBytes {
			return ExecutionResult{}, resourceLimit("output_bytes", len(current), budget.MaxOutputBytes)
		}
		codepoints := utf8.RuneCountInString(current)
		if codepoints > budget.MaxOutputCodepoints {
			return ExecutionResult{}, resourceLimit("output_codepoints", codepoints, budget.MaxOutputCodepoints)
		}
		trace = append(trace, StepTrace{Index: index, Operation: step.Kind, StemID: step.StemID, Applied: true, OutputSize: len(current)})
	}
	if !hasCurrent || current == "" {
		return ExecutionResult{}, unknown("EMPTY_FORM", "output", "procedure produced no surface form")
	}
	return ExecutionResult{Form: current, Trace: trace}, nil
}

func validateFeatureBundle(bundle features.FeatureBundle) error {
	if !validLanguage(string(bundle.Language)) || !validToken(string(bundle.ProfileVersion)) {
		return invalid("INVALID_FEATURE_BUNDLE", "features", "feature bundle requires a bounded language and profile version")
	}
	if len(bundle.Values) > features.MaxBundleValues {
		return resourceLimit("feature_values", len(bundle.Values), features.MaxBundleValues)
	}
	seen := make(map[features.FeatureID]struct{}, len(bundle.Values))
	for _, value := range bundle.Values {
		if !validToken(string(value.Feature)) {
			return invalid("INVALID_FEATURE_ID", "features", "feature ID is malformed")
		}
		if _, exists := seen[value.Feature]; exists {
			return invalid("DUPLICATE_FEATURE_ID", "features", "feature bundle contains duplicate dimensions")
		}
		seen[value.Feature] = struct{}{}
		switch value.State {
		case features.StateKnown:
			if !validToken(string(value.Value)) || value.Variable != nil {
				return invalid("INVALID_KNOWN_FEATURE", "features."+string(value.Feature), "known feature requires a value and cannot carry a variable")
			}
		case features.StateAbsent:
			if value.Value != "" || value.Variable != nil {
				return invalid("INVALID_ABSENT_FEATURE", "features."+string(value.Feature), "absent feature cannot carry a value or variable")
			}
		case features.StateUnknown:
			if value.Value != "" {
				return invalid("INVALID_UNKNOWN_FEATURE", "features."+string(value.Feature), "unknown feature cannot carry a concrete value")
			}
			if value.Variable != nil && (!validToken(string(value.Variable.ID)) || !validToken(string(value.Variable.Scope))) {
				return invalid("INVALID_FEATURE_VARIABLE", "features."+string(value.Feature), "scoped variable reference is malformed")
			}
		default:
			return invalid("INVALID_FEATURE_STATE", "features."+string(value.Feature), "feature state is unsupported")
		}
	}
	return nil
}

func checkConditions(conditions []FeatureCondition, bundle features.FeatureBundle) (bool, error) {
	values := make(map[features.FeatureID]features.FeatureValue, len(bundle.Values))
	for _, value := range bundle.Values {
		values[value.Feature] = value
	}
	for _, condition := range conditions {
		value, exists := values[condition.Feature]
		path := "features." + string(condition.Feature)
		if !exists || value.State == features.StateAbsent {
			return false, unknown("FEATURE_CONDITION_MISSING", path, "operation requires an explicitly supplied known feature value")
		}
		if value.State == features.StateUnknown {
			return false, incomplete("FEATURE_CONDITION_UNKNOWN", path, "operation guard depends on an unresolved feature")
		}
		if value.Value != condition.Value {
			return false, nil
		}
	}
	return true, nil
}

func validOperationText(value string, allowEmpty bool) bool {
	if (!allowEmpty && len(value) == 0) || len(value) > MaxOperationTextBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validToken(value string) bool {
	if len(value) == 0 || len(value) > MaxTokenBytes || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func validLanguage(value string) bool {
	if len(value) == 0 || len(value) > 35 || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if !(character >= 'a' && character <= 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

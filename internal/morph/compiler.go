package morph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	ParadigmSchemaVersion = 1
	MaxParadigms          = 1_024
	MaxCellsPerParadigm   = 512
	MaxRulesPerParadigm   = 512
	MaxParentsPerParadigm = 32
	MaxInheritanceDepth   = 32
)

type CompilerBudget struct {
	MaxParadigms        int
	MaxCells            int
	MaxRules            int
	MaxInheritanceDepth int
}

func DefaultCompilerBudget() CompilerBudget {
	return CompilerBudget{MaxParadigms: MaxParadigms, MaxCells: MaxParadigms * MaxCellsPerParadigm, MaxRules: MaxParadigms * MaxRulesPerParadigm, MaxInheritanceDepth: MaxInheritanceDepth}
}

func (budget CompilerBudget) Validate() error {
	if budget.MaxParadigms <= 0 || budget.MaxParadigms > MaxParadigms || budget.MaxCells <= 0 || budget.MaxCells > MaxParadigms*MaxCellsPerParadigm || budget.MaxRules <= 0 || budget.MaxRules > MaxParadigms*MaxRulesPerParadigm || budget.MaxInheritanceDepth <= 0 || budget.MaxInheritanceDepth > MaxInheritanceDepth {
		return morphError(status.OutcomeInvalid, "INVALID_COMPILER_BUDGET", "budget", "paradigm compiler bounds must be positive and cannot exceed package ceilings")
	}
	return nil
}

// CompiledParadigms is an immutable flattened view of validated paradigms.
type CompiledParadigms struct {
	items   []Paradigm
	byID    map[ParadigmID]int
	profile *features.LanguageFeatureProfile
}

// CompileParadigms validates one language-local inheritance set, detects
// cycles, duplicate cells and conflicting principal-part requirements, and
// returns flattened immutable paradigms.
func CompileParadigms(input []Paradigm, profile *features.LanguageFeatureProfile, compilerBudget CompilerBudget, procedureBudget procedure.Budget) (*CompiledParadigms, error) {
	if profile == nil {
		return nil, morphError(status.OutcomeUnknown, "FEATURE_PROFILE_UNAVAILABLE", "profile", "paradigm compilation requires a language feature profile")
	}
	if err := compilerBudget.Validate(); err != nil {
		return nil, err
	}
	if err := procedureBudget.Validate(); err != nil {
		return nil, wrapError(err)
	}
	if len(input) == 0 {
		return nil, morphError(status.OutcomeInvalid, "EMPTY_PARADIGM_SET", "paradigms", "at least one paradigm is required")
	}
	if len(input) > compilerBudget.MaxParadigms {
		return nil, morphLimit("paradigms", len(input), compilerBudget.MaxParadigms)
	}
	base := make(map[ParadigmID]Paradigm, len(input))
	totalCells, totalRules := 0, 0
	for index, paradigm := range input {
		path := fmt.Sprintf("paradigms[%d]", index)
		if paradigm.SchemaVersion != ParadigmSchemaVersion {
			return nil, morphError(status.OutcomeInvalid, "INVALID_PARADIGM_SCHEMA", path+".schema_version", "unsupported paradigm schema version")
		}
		if !validMorphToken(string(paradigm.ID)) {
			return nil, morphError(status.OutcomeInvalid, "INVALID_PARADIGM_ID", path+".id", "paradigm ID is malformed")
		}
		if _, exists := base[paradigm.ID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_PARADIGM_ID", path+".id", "paradigm ID is duplicated")
		}
		if paradigm.Language != profile.Language() || paradigm.ProfileVersion != profile.Version() {
			return nil, morphError(status.OutcomeIncomplete, "FEATURE_PROFILE_MISMATCH", path, "paradigm and feature profile language/version differ")
		}
		if !validMorphToken(string(paradigm.POS)) {
			return nil, morphError(status.OutcomeInvalid, "INVALID_PARADIGM_POS", path+".pos", "part-of-speech label is malformed")
		}
		if len(paradigm.Cells) > MaxCellsPerParadigm || len(paradigm.Rules) > MaxRulesPerParadigm {
			return nil, morphLimit("paradigm_records", max(len(paradigm.Cells), len(paradigm.Rules)), max(MaxCellsPerParadigm, MaxRulesPerParadigm))
		}
		if len(paradigm.Inherits) > MaxParentsPerParadigm {
			return nil, morphLimit("parents_per_paradigm", len(paradigm.Inherits), MaxParentsPerParadigm)
		}
		totalCells += len(paradigm.Cells)
		totalRules += len(paradigm.Rules)
		if totalCells > compilerBudget.MaxCells {
			return nil, morphLimit("cells", totalCells, compilerBudget.MaxCells)
		}
		if totalRules > compilerBudget.MaxRules {
			return nil, morphLimit("rules", totalRules, compilerBudget.MaxRules)
		}
		if err := validateLocalParadigm(paradigm, profile, procedureBudget); err != nil {
			return nil, wrapError(err)
		}
		base[paradigm.ID] = cloneParadigm(paradigm)
	}
	for id, paradigm := range base {
		seen := make(map[ParadigmID]struct{}, len(paradigm.Inherits))
		for _, parent := range paradigm.Inherits {
			if !validMorphToken(string(parent)) {
				return nil, morphError(status.OutcomeInvalid, "INVALID_PARADIGM_ID", "paradigms."+string(id)+".inherits", "inherited paradigm ID is malformed")
			}
			if parent == id {
				return nil, morphError(status.OutcomeInvalid, "PARADIGM_INHERITANCE_CYCLE", "paradigms."+string(id)+".inherits", "paradigm cannot inherit itself")
			}
			if _, duplicate := seen[parent]; duplicate {
				return nil, morphError(status.OutcomeInvalid, "DUPLICATE_PARADIGM_PARENT", "paradigms."+string(id)+".inherits", "inherited paradigm is listed more than once")
			}
			seen[parent] = struct{}{}
			if _, exists := base[parent]; !exists {
				return nil, morphError(status.OutcomeUnknown, "UNKNOWN_PARADIGM", "paradigms."+string(id)+".inherits", "inherited paradigm does not resolve")
			}
		}
	}
	if err := checkInheritanceCycles(base, compilerBudget.MaxInheritanceDepth); err != nil {
		return nil, err
	}
	resolved := make(map[ParadigmID]Paradigm, len(base))
	resolving := make(map[ParadigmID]bool, len(base))
	flattenedCells, flattenedRules := 0, 0
	var resolve func(ParadigmID) (Paradigm, error)
	resolve = func(id ParadigmID) (Paradigm, error) {
		if result, exists := resolved[id]; exists {
			return cloneParadigm(result), nil
		}
		if resolving[id] {
			return Paradigm{}, morphError(status.OutcomeInvalid, "PARADIGM_INHERITANCE_CYCLE", "paradigms."+string(id), "inheritance cycle detected")
		}
		resolving[id] = true
		current := cloneParadigm(base[id])
		cells := make(map[CellID]ParadigmCell)
		cellSignatures := make(map[string]CellID)
		rules := make(map[CellID]CellRule)
		requirements := make(map[procedure.StemID]PrincipalPartRequirement)
		for _, parentID := range current.Inherits {
			parent, err := resolve(parentID)
			if err != nil {
				return Paradigm{}, err
			}
			if parent.POS != current.POS {
				return Paradigm{}, morphError(status.OutcomeUnsupported, "INHERITED_POS_MISMATCH", "paradigms."+string(id)+".inherits", "inherited paradigm has a different part of speech")
			}
			for _, cell := range parent.Cells {
				if err := addInheritedCell(cells, cellSignatures, cell, "paradigms."+string(id)+".inherits"); err != nil {
					return Paradigm{}, err
				}
			}
			for _, rule := range parent.Rules {
				if old, exists := rules[rule.CellID]; exists {
					if !equalRule(old, rule) {
						return Paradigm{}, morphError(status.OutcomeInvalid, "CONFLICTING_PARADIGM_RULE", "paradigms."+string(id)+".inherits", "inherited paradigms define conflicting rules for one cell")
					}
					continue
				}
				rules[rule.CellID] = cloneRule(rule)
			}
			for _, requirement := range parent.PrincipalPartRequirements {
				if err := addRequirement(requirements, requirement, "paradigms."+string(id)+".inherits"); err != nil {
					return Paradigm{}, err
				}
			}
		}
		for _, cell := range current.Cells {
			if err := addCell(cells, cellSignatures, cell, "paradigms."+string(id)+".cells"); err != nil {
				return Paradigm{}, err
			}
		}
		for _, rule := range current.Rules {
			if _, exists := rules[rule.CellID]; exists {
				return Paradigm{}, morphError(status.OutcomeInvalid, "DUPLICATE_PARADIGM_RULE", "paradigms."+string(id)+".rules", "cell already has an inherited rule")
			}
			rules[rule.CellID] = cloneRule(rule)
		}
		for _, requirement := range current.PrincipalPartRequirements {
			if err := addRequirement(requirements, requirement, "paradigms."+string(id)+".principal_part_requirements"); err != nil {
				return Paradigm{}, err
			}
		}
		current.Cells = mapCells(cells)
		current.Rules = mapRules(rules)
		current.PrincipalPartRequirements = mapRequirements(requirements)
		current.Inherits = nil
		declaredStems := make(map[procedure.StemID]struct{}, len(current.PrincipalPartRequirements))
		for _, requirement := range current.PrincipalPartRequirements {
			declaredStems[requirement.StemID] = struct{}{}
		}
		for _, rule := range current.Rules {
			for stepIndex, step := range rule.Steps {
				if step.Kind == procedure.OperationSelectStem {
					if _, exists := declaredStems[step.StemID]; !exists {
						return Paradigm{}, morphError(status.OutcomeInvalid, "UNDECLARED_PRINCIPAL_STEM", fmt.Sprintf("paradigms.%s.rules.%s.steps[%d].stem_id", id, rule.CellID, stepIndex), "rule selects a stem absent from the flattened paradigm's principal-part requirements")
					}
				}
			}
		}
		if len(current.Cells) == 0 {
			return Paradigm{}, morphError(status.OutcomeInvalid, "EMPTY_PARADIGM_CELLS", "paradigms."+string(id)+".cells", "flattened paradigm has no cells")
		}
		flattenedCells += len(current.Cells)
		flattenedRules += len(current.Rules)
		if flattenedCells > compilerBudget.MaxCells {
			return Paradigm{}, morphLimit("flattened_cells", flattenedCells, compilerBudget.MaxCells)
		}
		if flattenedRules > compilerBudget.MaxRules {
			return Paradigm{}, morphLimit("flattened_rules", flattenedRules, compilerBudget.MaxRules)
		}
		resolving[id] = false
		resolved[id] = cloneParadigm(current)
		return current, nil
	}
	result := &CompiledParadigms{items: make([]Paradigm, 0, len(base)), byID: make(map[ParadigmID]int, len(base)), profile: profile}
	ids := make([]ParadigmID, 0, len(base))
	for id := range base {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		paradigm, err := resolve(id)
		if err != nil {
			return nil, err
		}
		result.byID[id] = len(result.items)
		result.items = append(result.items, cloneParadigm(paradigm))
	}
	return result, nil
}

func addRequirement(requirements map[procedure.StemID]PrincipalPartRequirement, requirement PrincipalPartRequirement, path string) error {
	if previous, exists := requirements[requirement.StemID]; exists {
		if previous != requirement {
			return morphError(status.OutcomeInvalid, "CONFLICTING_PRINCIPAL_PART_REQUIREMENT", path, "paradigm inheritance maps one stem to contradictory diagnostic forms")
		}
		return nil
	}
	for _, previous := range requirements {
		if previous.PartID == requirement.PartID {
			return morphError(status.OutcomeInvalid, "CONFLICTING_PRINCIPAL_PART_REQUIREMENT", path, "paradigm inheritance maps one diagnostic form to multiple stems")
		}
	}
	requirements[requirement.StemID] = requirement
	return nil
}

func validateLocalParadigm(paradigm Paradigm, profile *features.LanguageFeatureProfile, procedureBudget procedure.Budget) error {
	if len(paradigm.PrincipalPartRequirements) > MaxPrincipalParts {
		return morphLimit("principal_part_requirements", len(paradigm.PrincipalPartRequirements), MaxPrincipalParts)
	}
	seenStems := make(map[procedure.StemID]struct{}, len(paradigm.PrincipalPartRequirements))
	seenParts := make(map[PrincipalPartID]struct{}, len(paradigm.PrincipalPartRequirements))
	for index, requirement := range paradigm.PrincipalPartRequirements {
		path := fmt.Sprintf("paradigm.principal_part_requirements[%d]", index)
		if !validStemID(requirement.StemID) || !validPrincipalPartID(requirement.PartID) {
			return morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_REQUIREMENT", path, "stem and diagnostic part IDs are malformed")
		}
		if _, exists := seenStems[requirement.StemID]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_STEM_REQUIREMENT", path+".stem_id", "stem requirement is duplicated")
		}
		if _, exists := seenParts[requirement.PartID]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_PRINCIPAL_PART_REQUIREMENT", path+".part_id", "diagnostic part is mapped more than once")
		}
		seenStems[requirement.StemID] = struct{}{}
		seenParts[requirement.PartID] = struct{}{}
	}
	if len(paradigm.Cells) == 0 {
		return morphError(status.OutcomeInvalid, "EMPTY_PARADIGM_CELLS", "paradigm.cells", "paradigm must declare at least one cell")
	}
	cellIDs := make(map[CellID]struct{}, len(paradigm.Cells))
	cellSignatures := make(map[string]struct{}, len(paradigm.Cells))
	for index, cell := range paradigm.Cells {
		path := fmt.Sprintf("paradigm.cells[%d]", index)
		if !validMorphToken(string(cell.ID)) {
			return morphError(status.OutcomeInvalid, "INVALID_CELL_ID", path+".id", "cell ID is malformed")
		}
		if _, exists := cellIDs[cell.ID]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_CELL_ID", path+".id", "cell ID is duplicated")
		}
		cellIDs[cell.ID] = struct{}{}
		bundle, err := features.NewBundle(profile, cell.Features)
		if err != nil {
			return err
		}
		if len(bundle.Values) == 0 {
			return morphError(status.OutcomeInvalid, "EMPTY_CELL_FEATURES", path+".features", "cell requires at least one known feature value")
		}
		for _, value := range bundle.Values {
			if value.State != features.StateKnown {
				return morphError(status.OutcomeInvalid, "NONCONCRETE_CELL_FEATURE", path+".features", "paradigm cell keys require known feature values")
			}
		}
		signature := featureSignature(bundle.Values)
		if _, exists := cellSignatures[signature]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_CELL_FEATURES", path+".features", "two cells declare the same feature combination")
		}
		cellSignatures[signature] = struct{}{}
	}
	ruleCells := make(map[CellID]struct{}, len(paradigm.Rules))
	for index, rule := range paradigm.Rules {
		path := fmt.Sprintf("paradigm.rules[%d]", index)
		if rule.ProcedureID != "" {
			if len(rule.ProcedureID) > status.MaxTraceIDBytes {
				return morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_ID", path+".procedure_id", "procedure ID exceeds the trace identifier limit")
			}
			procedureID, err := ids.ParseProcedureID(rule.ProcedureID)
			if err != nil || procedureID.String() != rule.ProcedureID {
				return morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_ID", path+".procedure_id", "procedure ID must be a canonical procedure symbol")
			}
		}
		if rule.Provenance != (provenance.SourceRef{}) {
			if rule.ProcedureID == "" {
				return morphError(status.OutcomeInvalid, "MISSING_PROCEDURE_ID", path+".procedure_id", "a source-attributed rule requires a procedure ID")
			}
			if err := rule.Provenance.Validate(); err != nil {
				return morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_PROVENANCE", path+".provenance", err.Error())
			}
		}
		if _, exists := cellIDs[rule.CellID]; !exists {
			return morphError(status.OutcomeUnknown, "UNKNOWN_PARADIGM_CELL", path+".cell_id", "rule cell does not resolve in this paradigm")
		}
		if _, exists := ruleCells[rule.CellID]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_PARADIGM_RULE", path+".cell_id", "cell has multiple local rules")
		}
		ruleCells[rule.CellID] = struct{}{}
		if err := procedure.ValidateOperations(rule.Steps, procedureBudget); err != nil {
			return wrapError(err)
		}
		for stepIndex, step := range rule.Steps {
			for _, guard := range step.When {
				dimension, exists := profile.Lookup(guard.Feature)
				if !exists {
					return morphError(status.OutcomeUnsupported, "UNSUPPORTED_FEATURE_DIMENSION", fmt.Sprintf("%s.steps[%d].when", path, stepIndex), "operation guard uses an undeclared feature")
				}
				if !containsFeatureValue(dimension.Values, guard.Value) {
					return morphError(status.OutcomeUnsupported, "UNSUPPORTED_FEATURE_VALUE", fmt.Sprintf("%s.steps[%d].when", path, stepIndex), "operation guard value is not declared by the language profile")
				}
			}
		}
	}
	return nil
}

func checkInheritanceCycles(paradigms map[ParadigmID]Paradigm, maxDepth int) error {
	state := make(map[ParadigmID]uint8, len(paradigms))
	var visit func(ParadigmID, int) error
	visit = func(id ParadigmID, depth int) error {
		if depth > maxDepth {
			return morphLimit("inheritance_depth", depth, maxDepth)
		}
		if state[id] == 1 {
			return morphError(status.OutcomeInvalid, "PARADIGM_INHERITANCE_CYCLE", "paradigms."+string(id), "inheritance cycle detected")
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, parent := range paradigms[id].Inherits {
			if err := visit(parent, depth+1); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	ids := make([]ParadigmID, 0, len(paradigms))
	for id := range paradigms {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if err := visit(id, 1); err != nil {
			return err
		}
	}
	return nil
}

func addCell(cells map[CellID]ParadigmCell, signatures map[string]CellID, cell ParadigmCell, path string) error {
	if _, exists := cells[cell.ID]; exists {
		return morphError(status.OutcomeInvalid, "DUPLICATE_CELL_ID", path+"."+string(cell.ID), "cell ID is defined more than once in the inheritance result")
	}
	signature := featureSignature(cell.Features)
	if previous, exists := signatures[signature]; exists {
		return morphError(status.OutcomeInvalid, "DUPLICATE_CELL_FEATURES", path+"."+string(cell.ID), "cell duplicates feature combination from "+string(previous))
	}
	cells[cell.ID] = cloneCell(cell)
	signatures[signature] = cell.ID
	return nil
}

// Multiple inheritance may encounter the same shared ancestor through a
// diamond. Identical definitions are one inherited declaration; differences
// remain a duplicate/conflict and fail closed.
func addInheritedCell(cells map[CellID]ParadigmCell, signatures map[string]CellID, cell ParadigmCell, path string) error {
	if previous, exists := cells[cell.ID]; exists {
		if equalCell(previous, cell) {
			return nil
		}
		return morphError(status.OutcomeInvalid, "DUPLICATE_CELL_ID", path+"."+string(cell.ID), "inherited paradigms define conflicting versions of one cell")
	}
	return addCell(cells, signatures, cell, path)
}

func equalCell(left, right ParadigmCell) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func featureSignature(values []features.FeatureValue) string {
	ordered := append([]features.FeatureValue(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Feature < ordered[j].Feature })
	var builder strings.Builder
	for _, value := range ordered {
		builder.WriteString(string(value.Feature))
		builder.WriteByte('=')
		builder.WriteString(string(value.Value))
		builder.WriteByte(';')
	}
	return builder.String()
}

func mapCells(values map[CellID]ParadigmCell) []ParadigmCell {
	ids := make([]CellID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := make([]ParadigmCell, len(ids))
	for index, id := range ids {
		result[index] = cloneCell(values[id])
	}
	return result
}

func mapRules(values map[CellID]CellRule) []CellRule {
	ids := make([]CellID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := make([]CellRule, len(ids))
	for index, id := range ids {
		result[index] = cloneRule(values[id])
	}
	return result
}

func mapRequirements(values map[procedure.StemID]PrincipalPartRequirement) []PrincipalPartRequirement {
	ids := make([]procedure.StemID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := make([]PrincipalPartRequirement, len(ids))
	for index, id := range ids {
		result[index] = values[id]
	}
	return result
}

func cloneCell(source ParadigmCell) ParadigmCell {
	result := source
	result.Features = cloneFeatureValues(source.Features)
	return result
}

func cloneRule(source CellRule) CellRule {
	result := source
	result.Steps = make([]procedure.Operation, len(source.Steps))
	for index, operation := range source.Steps {
		result.Steps[index] = operation
		result.Steps[index].When = append([]procedure.FeatureCondition(nil), operation.When...)
		if operation.Replacement != nil {
			replacement := *operation.Replacement
			result.Steps[index].Replacement = &replacement
		}
		if operation.To != nil {
			to := *operation.To
			result.Steps[index].To = &to
		}
	}
	return result
}

func cloneParadigm(source Paradigm) Paradigm {
	result := source
	result.Inherits = append([]ParadigmID(nil), source.Inherits...)
	result.PrincipalPartRequirements = append([]PrincipalPartRequirement(nil), source.PrincipalPartRequirements...)
	result.Cells = make([]ParadigmCell, len(source.Cells))
	for index, cell := range source.Cells {
		result.Cells[index] = cloneCell(cell)
	}
	result.Rules = make([]CellRule, len(source.Rules))
	for index, rule := range source.Rules {
		result.Rules[index] = cloneRule(rule)
	}
	return result
}

func equalRule(left, right CellRule) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func cloneFeatureValues(values []features.FeatureValue) []features.FeatureValue {
	result := make([]features.FeatureValue, len(values))
	for index, value := range values {
		result[index] = value
		if value.Variable != nil {
			variable := *value.Variable
			result[index].Variable = &variable
		}
	}
	return result
}

func containsFeatureValue(values []features.FeatureValueID, candidate features.FeatureValueID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func wrapError(err error) error {
	if resultError, ok := err.(interface{ Result() status.Result }); ok {
		return &MorphError{result: resultError.Result()}
	}
	return morphError(status.OutcomeError, "MORPH_COMPILATION_ERROR", "paradigm", "paradigm compilation failed unexpectedly")
}

func morphLimit(resource string, value, maximum int) error {
	return morphError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("paradigm %s limit exceeded: %d > %d", resource, value, maximum))
}

func (compiled *CompiledParadigms) Lookup(id ParadigmID) (Paradigm, bool) {
	if compiled == nil {
		return Paradigm{}, false
	}
	index, exists := compiled.byID[id]
	if !exists {
		return Paradigm{}, false
	}
	return cloneParadigm(compiled.items[index]), true
}

func (compiled *CompiledParadigms) Paradigms() []Paradigm {
	if compiled == nil {
		return nil
	}
	result := make([]Paradigm, len(compiled.items))
	for index, paradigm := range compiled.items {
		result[index] = cloneParadigm(paradigm)
	}
	return result
}

// SelectCell resolves exactly one compatible cell. Unresolved required values
// return incomplete and multiple matches return ambiguous.
func (compiled *CompiledParadigms) SelectCell(id ParadigmID, bundle features.FeatureBundle) (ParadigmCell, error) {
	paradigm, exists := compiled.Lookup(id)
	if !exists {
		return ParadigmCell{}, morphError(status.OutcomeUnknown, "UNKNOWN_PARADIGM", "paradigm.id", "paradigm does not resolve")
	}
	if err := features.ValidateBundle(compiled.profile, bundle); err != nil {
		return ParadigmCell{}, wrapError(err)
	}
	values := make(map[features.FeatureID]features.FeatureValue, len(bundle.Values))
	for _, value := range bundle.Values {
		values[value.Feature] = value
	}
	matches := make([]ParadigmCell, 0, 1)
	possible := false
	for _, cell := range paradigm.Cells {
		matched, unresolved := true, false
		for _, requirement := range cell.Features {
			value, supplied := values[requirement.Feature]
			if !supplied || value.State == features.StateAbsent {
				matched, unresolved = false, true
				continue
			}
			if value.State == features.StateUnknown {
				matched, unresolved = false, true
				continue
			}
			if value.Value != requirement.Value {
				matched = false
			}
		}
		if matched {
			matches = append(matches, cloneCell(cell))
		} else if matched && unresolved {
			possible = true
		}
	}
	if len(matches) > 1 {
		return ParadigmCell{}, morphError(status.OutcomeAmbiguous, "AMBIGUOUS_PARADIGM_CELL", "features", "multiple paradigm cells match the supplied feature bundle")
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if possible {
		return ParadigmCell{}, morphError(status.OutcomeIncomplete, "INCOMPLETE_PARADIGM_FEATURES", "features", "unresolved feature values could select a paradigm cell")
	}
	return ParadigmCell{}, morphError(status.OutcomeUnknown, "UNKNOWN_PARADIGM_CELL", "features", "no paradigm cell matches the supplied feature bundle")
}

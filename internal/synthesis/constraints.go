package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	DefaultGoalCheckMaxWork = 262_144
	MaxGoalCheckMaxWork     = 262_144

	GoalCheckDispositionChecked      = "checked"
	GoalCheckDispositionRejected     = "rejected"
	GoalCheckDispositionUndetermined = "undetermined"
	GoalCheckScope                   = "structural_profile_only"
)

// GoalConstraintStatus reports the independently derived match for one input
// constraint. Fragments are paths into the submitted expression, not proof
// receipts or trusted constructor annotations.
type GoalConstraintStatus struct {
	ConstraintID string         `json:"constraint_id,omitempty"`
	Kind         ConstraintKind `json:"kind,omitempty"`
	Status       string         `json:"status"`
	Fragments    []string       `json:"fragments"`
	Code         string         `json:"code,omitempty"`
	Path         string         `json:"path,omitempty"`
	Message      string         `json:"message,omitempty"`
}

// GoalConstraintCheck separates expression typing from structural goal
// coverage. A satisfied result is only a check within the bounded structural
// profile; it does not establish that the proposition is true in any world.
type GoalConstraintCheck struct {
	Status           status.OutcomeKind     `json:"status"`
	Disposition      string                 `json:"disposition"`
	TypeCorrect      bool                   `json:"type_correct"`
	GoalTypeMatches  bool                   `json:"goal_type_matches"`
	GoalSatisfied    bool                   `json:"goal_satisfied"`
	Complete         bool                   `json:"complete"`
	CheckedType      *algebra.Type          `json:"checked_type,omitempty"`
	Interface        []algebra.PortBinding  `json:"interface"`
	Constraints      []GoalConstraintStatus `json:"constraints"`
	Unresolved       []Unresolved           `json:"unresolved"`
	Assumptions      []string               `json:"assumptions"`
	ConsistencyScope string                 `json:"consistency_scope"`
	VisitedNodes     int                    `json:"visited_nodes"`
	WorkUnits        int                    `json:"work_units"`
}

type goalCheckValue struct {
	kind algebra.TermKind
	id   string
}

func (value goalCheckValue) key() string { return string(value.kind) + "\x00" + value.id }

type goalCheckAtom struct {
	term *algebra.Term
	path string
}

type goalCheckInventory struct {
	atoms        []goalCheckAtom
	existentials map[string]algebra.Type
	holes        []Unresolved
	visited      int
	work         int
	maxWork      int
}

func (inventory *goalCheckInventory) charge(path string) error {
	if inventory.work >= inventory.maxWork {
		return goalCheckResource(path, "goal constraint traversal exceeds its logical work budget")
	}
	inventory.work++
	return nil
}

// CheckGoalConstraints independently checks a concrete candidate against one
// fully selected goal. It uses the typed checker for scope and type safety,
// then inspects the candidate AST directly; it does not trust search traces,
// constructor obligations, or fragment mappings.
func CheckGoalConstraints(primitiveContext algebra.Document, concreteGoal GoalSpec, candidateDoc algebra.Document, candidateInterface []algebra.PortBinding, requested algebra.Limits, maxWork int) (GoalConstraintCheck, error) {
	if maxWork < 0 {
		return GoalConstraintCheck{}, goalCheckInvalid("INVALID_GOAL_CHECK_WORK", "max_work", "work budget cannot be negative")
	}
	if maxWork == 0 {
		maxWork = DefaultGoalCheckMaxWork
	}
	if maxWork > MaxGoalCheckMaxWork {
		return GoalConstraintCheck{}, goalCheckResource("max_work", "work budget exceeds the hard limit")
	}

	if primitiveContext.Expression != nil || len(primitiveContext.Parameters) != 0 || len(primitiveContext.References) != 0 {
		return GoalConstraintCheck{}, goalCheckInvalid("INVALID_PRIMITIVE_CONTEXT", "context", "selected primitive context must contain declarations only")
	}
	if primitiveContext.SchemaVersion != algebra.SchemaVersion {
		return GoalConstraintCheck{}, goalCheckInvalid("INVALID_PRIMITIVE_CONTEXT", "context.schema_version", "primitive context schema version is unsupported")
	}
	if err := algebra.CheckDeclarations(primitiveContext); err != nil {
		return GoalConstraintCheck{}, err
	}

	goalLimits, err := NormalizeGoalLimits(concreteGoal.Limits)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	requestedLimits, err := algebra.NormalizeLimits(requested)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	candidateLimits, err := algebra.NormalizeLimits(candidateDoc.Limits)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	contextLimits, err := algebra.NormalizeLimits(primitiveContext.Limits)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	limits := minimumAlgebraLimits(goalLimits.TermLimits, requestedLimits, candidateLimits, contextLimits)
	if limits.MaxNodes > maxWork {
		limits.MaxNodes = maxWork
	}
	if limits.MaxDepth > maxWork {
		limits.MaxDepth = maxWork
	}
	if len(primitiveContext.Sorts)+len(primitiveContext.Signatures)+len(concreteGoal.References) > limits.MaxSymbols {
		return GoalConstraintCheck{}, goalCheckResource("references", "combined declarations exceed the effective checker limit")
	}

	unresolved, err := ValidateGoal(primitiveContext, concreteGoal)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	if len(primitiveContext.Sorts)+len(primitiveContext.Signatures)+len(concreteGoal.References) > algebra.DefaultMaxSymbols {
		return GoalConstraintCheck{}, goalCheckResource("references", "combined declarations exceed the bounded profile")
	}
	if len(candidateDoc.Sorts)+len(candidateDoc.Signatures)+len(candidateDoc.Parameters)+len(candidateDoc.References) > limits.MaxSymbols {
		return GoalConstraintCheck{}, goalCheckResource("candidate", "candidate declaration count exceeds the effective checker limit")
	}

	if candidateDoc.SchemaVersion != algebra.SchemaVersion {
		return GoalConstraintCheck{}, goalCheckInvalid("INVALID_CANDIDATE_SCHEMA_VERSION", "candidate.schema_version", "candidate document schema version is unsupported")
	}
	if len(candidateDoc.Parameters) != 0 {
		return GoalConstraintCheck{}, goalCheckInvalid("CANDIDATE_PARAMETERS_UNSUPPORTED", "candidate.parameters", "the bounded goal profile does not permit free parameters")
	}
	if err := validateCandidateInlineDeclarations(primitiveContext, concreteGoal.References, candidateDoc, limits); err != nil {
		return GoalConstraintCheck{}, err
	}
	if err := validateCandidateReferences(concreteGoal.References, candidateDoc.References); err != nil {
		return GoalConstraintCheck{}, err
	}

	checkedDocument := algebra.Document{
		SchemaVersion: algebra.SchemaVersion,
		Sorts:         primitiveContext.Sorts,
		Signatures:    primitiveContext.Signatures,
		References:    concreteGoal.References,
		Closed:        candidateDoc.Closed,
		Expression:    candidateDoc.Expression,
		Limits:        limits,
	}
	checked, err := algebra.Check(checkedDocument)
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	checkedType := checked.Type
	result := GoalConstraintCheck{
		Status: checked.Status, Disposition: GoalCheckDispositionUndetermined,
		TypeCorrect: true, GoalTypeMatches: algebra.TypeEqual(checked.Type, concreteGoal.ExpectedType),
		CheckedType: &checkedType, Interface: append([]algebra.PortBinding{}, candidateInterface...),
		Constraints:      make([]GoalConstraintStatus, 0, len(concreteGoal.Constraints)),
		Unresolved:       append([]Unresolved{}, unresolved...),
		Assumptions:      []string{"positive-facts-v1 structural constraints only; no fact or world-truth evaluation"},
		ConsistencyScope: GoalCheckScope,
		WorkUnits:        checked.NodeCount,
	}
	if len(unresolved) != 0 {
		result.Status = status.OutcomeIncomplete
		result.Complete = false
		result.Unresolved = append([]Unresolved{}, unresolved...)
		result.Constraints = undeterminedConstraintStatuses(concreteGoal.Constraints, unresolved)
		return result, nil
	}
	if checked.Status == status.OutcomeIncomplete {
		result.Status = status.OutcomeIncomplete
		result.Complete = false
		result.Unresolved = append(result.Unresolved, Unresolved{Code: "CANDIDATE_HOLE", Path: "candidate.expression", Message: "candidate contains an unresolved typed hole"})
		result.Constraints = undeterminedConstraintStatuses(concreteGoal.Constraints, result.Unresolved)
		return result, nil
	}

	if !result.GoalTypeMatches {
		return rejectGoalCandidate(result, concreteGoal, "GOAL_TYPE_MISMATCH", "candidate.expression", "candidate type does not equal the goal's expected type"), nil
	}
	openPorts := openGoalPortsForCheck(concreteGoal)
	leading, body := peelGoalLambdas(candidateDoc.Expression, len(openPorts))
	if len(leading) != len(openPorts) || len(candidateInterface) != len(openPorts) {
		return rejectGoalCandidate(result, concreteGoal, "GOAL_INTERFACE_MISMATCH", "candidate.interface", "leading lambda count must match the open, nonprojected goal ports"), nil
	}
	lambdaTypes := make(map[string]algebra.Type, len(leading))
	lambdaPaths := make(map[string]string, len(leading))
	for i, port := range openPorts {
		lambda := leading[i]
		binding := candidateInterface[i]
		if lambda == nil || lambda.Kind != algebra.TermLambda || lambda.Type == nil ||
			binding.Port != port.ID || binding.Binder != lambda.ID ||
			!algebra.TypeEqual(*lambda.Type, port.Type) {
			return rejectGoalCandidate(result, concreteGoal, "GOAL_INTERFACE_MISMATCH", fmt.Sprintf("candidate.interface[%d]", i), "interface ports, binder IDs, order, and types must match the goal lambdas exactly"), nil
		}
		lambdaTypes[lambda.ID] = *lambda.Type
		lambdaPaths[port.ID] = lambda.ID
	}
	if body == nil {
		return rejectGoalCandidate(result, concreteGoal, "MISSING_CANDIDATE_BODY", "candidate.expression", "candidate lambda interface has no proposition body"), nil
	}

	remainingWork := maxWork - checked.NodeCount
	if remainingWork < 0 {
		return GoalConstraintCheck{}, goalCheckResource("candidate.expression", "candidate type check exceeds the logical work budget")
	}
	inventory := goalCheckInventory{existentials: make(map[string]algebra.Type), maxWork: remainingWork}
	for i := range leading {
		if err := inventory.charge(fmt.Sprintf("candidate.expression.lambda[%d]", i)); err != nil {
			return GoalConstraintCheck{}, err
		}
		inventory.visited++
	}
	if err := walkGoalConstraintTerm(body, "candidate.expression"+goalLambdaBodySuffix(len(leading)), 1, &inventory, limits); err != nil {
		if algebraErrorIsIncomplete(err) {
			result.Status = status.OutcomeIncomplete
			result.Complete = false
			result.Unresolved = append(result.Unresolved, Unresolved{Code: "CANDIDATE_HOLE", Path: "candidate.expression", Message: "candidate contains an unresolved typed hole"})
			result.Constraints = undeterminedConstraintStatuses(concreteGoal.Constraints, result.Unresolved)
			result.VisitedNodes, result.WorkUnits = inventory.visited, checked.NodeCount+inventory.work
			return result, nil
		}
		var typed *algebra.Error
		if asAlgebraError(err, &typed) && typed.Status == status.OutcomeUnsupported {
			result.Status = status.OutcomeUnsupported
			result.Complete = false
			result.Constraints = undeterminedConstraintStatuses(concreteGoal.Constraints, []Unresolved{{Code: typed.Code, Path: typed.Path, Message: typed.Message}})
			result.Unresolved = append(result.Unresolved, Unresolved{Code: typed.Code, Path: typed.Path, Message: typed.Message})
			result.VisitedNodes, result.WorkUnits = inventory.visited, checked.NodeCount+inventory.work
			return result, nil
		}
		return GoalConstraintCheck{}, err
	}
	result.VisitedNodes = inventory.visited
	result.WorkUnits += inventory.work
	if len(inventory.holes) != 0 {
		result.Status = status.OutcomeIncomplete
		result.Complete = false
		result.GoalSatisfied = false
		result.Unresolved = append(result.Unresolved, inventory.holes...)
		result.Constraints = undeterminedConstraintStatuses(concreteGoal.Constraints, result.Unresolved)
		return result, nil
	}

	matcherRemaining := maxWork - checked.NodeCount - inventory.work
	if matcherRemaining <= 0 {
		return GoalConstraintCheck{}, goalCheckResource("candidate.expression", "goal constraint matching exceeds the logical work budget")
	}
	checker := newGoalProfileMatcher(primitiveContext, concreteGoal, lambdaPaths, lambdaTypes, inventory, matcherRemaining)
	matched, err := checker.match()
	if err != nil {
		return GoalConstraintCheck{}, err
	}
	result.Complete = true
	result.GoalSatisfied = matched.satisfied
	result.Constraints = matched.constraints
	if matched.satisfied {
		result.Status = status.OutcomeOK
		result.Disposition = GoalCheckDispositionChecked
	} else {
		result.Status = status.OutcomeOK
		result.Disposition = GoalCheckDispositionRejected
	}
	result.WorkUnits += checker.work
	return result, nil
}

func minimumAlgebraLimits(limits ...algebra.Limits) algebra.Limits {
	result := algebra.DefaultLimits()
	for _, current := range limits {
		if current.MaxNodes < result.MaxNodes {
			result.MaxNodes = current.MaxNodes
		}
		if current.MaxDepth < result.MaxDepth {
			result.MaxDepth = current.MaxDepth
		}
		if current.MaxTypeDepth < result.MaxTypeDepth {
			result.MaxTypeDepth = current.MaxTypeDepth
		}
		if current.MaxSymbols < result.MaxSymbols {
			result.MaxSymbols = current.MaxSymbols
		}
		if current.MaxPorts < result.MaxPorts {
			result.MaxPorts = current.MaxPorts
		}
	}
	return result
}

func validateCandidateInlineDeclarations(context algebra.Document, references []algebra.Declaration, candidate algebra.Document, limits algebra.Limits) error {
	selectedSorts := make(map[string]struct{}, len(context.Sorts))
	for _, sort := range context.Sorts {
		selectedSorts[sort.ID] = struct{}{}
	}
	seenSorts := make(map[string]struct{}, len(candidate.Sorts))
	for i, sort := range candidate.Sorts {
		if _, duplicate := seenSorts[sort.ID]; duplicate {
			return goalCheckInvalid("DUPLICATE_CANDIDATE_SORT", fmt.Sprintf("candidate.sorts[%d].id", i), "candidate inline sort is duplicated")
		}
		seenSorts[sort.ID] = struct{}{}
		if _, selected := selectedSorts[sort.ID]; !selected {
			return goalCheckInvalid("CANDIDATE_SORT_NOT_SELECTED", fmt.Sprintf("candidate.sorts[%d].id", i), "candidate inline sort is not in the selected primitive inventory")
		}
	}
	selectedSignatures := make(map[string]algebra.Signature, len(context.Signatures))
	for _, signature := range context.Signatures {
		selectedSignatures[signature.ID] = signature
	}
	seenSignatures := make(map[string]struct{}, len(candidate.Signatures))
	for i, signature := range candidate.Signatures {
		path := fmt.Sprintf("candidate.signatures[%d]", i)
		if _, duplicate := seenSignatures[signature.ID]; duplicate {
			return goalCheckInvalid("DUPLICATE_CANDIDATE_SIGNATURE", path+".id", "candidate inline signature is duplicated")
		}
		seenSignatures[signature.ID] = struct{}{}
		selected, exists := selectedSignatures[signature.ID]
		if !exists || !sameGoalSignature(selected, signature) {
			return goalCheckInvalid("CANDIDATE_SIGNATURE_NOT_SELECTED", path, "candidate inline signature does not match the selected typed primitive inventory")
		}
	}
	// Validate inline type graphs against trusted selected sorts before any
	// subsequent structural comparison or candidate checking.
	declarations := algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Sorts: context.Sorts,
		Signatures: candidate.Signatures, References: references, Limits: limits,
	}
	return algebra.CheckDeclarations(declarations)
}

func sameGoalSignature(selected, candidate algebra.Signature) bool {
	if selected.ID != candidate.ID || len(selected.Ports) != len(candidate.Ports) {
		return false
	}
	for i := range selected.Ports {
		if selected.Ports[i].Role != candidate.Ports[i].Role || !algebra.TypeEqual(selected.Ports[i].Type, candidate.Ports[i].Type) {
			return false
		}
	}
	return true
}

func validateCandidateReferences(goalRefs, candidateRefs []algebra.Declaration) error {
	if len(goalRefs) != len(candidateRefs) {
		return goalCheckInvalid("CANDIDATE_REFERENCE_SET_MISMATCH", "candidate.references", "candidate references must exactly match the goal declarations")
	}
	goalByID := make(map[string]algebra.Type, len(goalRefs))
	for _, ref := range goalRefs {
		goalByID[ref.ID] = ref.Type
	}
	seen := make(map[string]struct{}, len(candidateRefs))
	for i, ref := range candidateRefs {
		path := fmt.Sprintf("candidate.references[%d]", i)
		if _, duplicate := seen[ref.ID]; duplicate {
			return goalCheckInvalid("DUPLICATE_CANDIDATE_REFERENCE", path+".id", "candidate reference is duplicated")
		}
		seen[ref.ID] = struct{}{}
		typ, exists := goalByID[ref.ID]
		if !exists || !algebra.TypeEqual(typ, ref.Type) {
			return goalCheckInvalid("CANDIDATE_REFERENCE_SET_MISMATCH", path, "candidate reference ID and type must match the goal")
		}
	}
	return nil
}

func openGoalPortsForCheck(goal GoalSpec) []GoalPort {
	projected := make(map[string]struct{}, len(goal.ProjectPorts))
	for _, id := range goal.ProjectPorts {
		projected[id] = struct{}{}
	}
	ports := make([]GoalPort, 0, len(goal.Ports))
	for _, port := range goal.Ports {
		if port.Binding == nil {
			if _, isProjected := projected[port.ID]; !isProjected {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

func peelGoalLambdas(expression *algebra.Term, expected int) ([]*algebra.Term, *algebra.Term) {
	leading := make([]*algebra.Term, 0, expected)
	current := expression
	for current != nil && current.Kind == algebra.TermLambda && len(leading) < expected {
		leading = append(leading, current)
		current = current.Body
	}
	if current != nil && current.Kind == algebra.TermLambda {
		return append(leading, current), current.Body
	}
	return leading, current
}

func goalLambdaBodySuffix(count int) string {
	path := ""
	for range count {
		path += ".body"
	}
	return path
}

func walkGoalConstraintTerm(term *algebra.Term, path string, depth int, inventory *goalCheckInventory, limits algebra.Limits) error {
	if err := inventory.charge(path); err != nil {
		return err
	}
	inventory.visited++
	if depth > limits.MaxDepth {
		return goalCheckResource(path, "candidate AST traversal exceeds the effective depth limit")
	}
	if term == nil {
		return goalCheckInvalid("MISSING_CANDIDATE_TERM", path, "candidate contains a missing term")
	}
	switch term.Kind {
	case algebra.TermAnd:
		for i, operand := range term.Operands {
			if err := walkGoalConstraintTerm(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1, inventory, limits); err != nil {
				return err
			}
		}
	case algebra.TermExists:
		if term.Type == nil {
			return goalCheckInvalid("MISSING_EXISTS_TYPE", path+".type", "existential binder type is missing")
		}
		if _, duplicate := inventory.existentials[term.ID]; duplicate {
			return goalCheckInvalid("DUPLICATE_EXISTS_BINDER", path+".id", "existential binder ID is duplicated")
		}
		inventory.existentials[term.ID] = *term.Type
		if err := walkGoalConstraintTerm(term.Body, path+".body", depth+1, inventory, limits); err != nil {
			return err
		}
	case algebra.TermNot:
		return goalCheckUnsupported("NOT_UNSUPPORTED", path, "the positive structural goal profile does not interpret Not")
	case algebra.TermQuantifier:
		return goalCheckUnsupported("QUANTIFIER_UNSUPPORTED", path, "quantifiers are represented but not executed by the positive structural goal profile")
	case algebra.TermAsk:
		return goalCheckUnsupported("ASK_UNSUPPORTED", path, "Ask is represented but query answer projection is outside the positive structural goal profile")
	case algebra.TermAtom:
		for i, argument := range term.Args {
			argumentPath := fmt.Sprintf("%s.args[%d].value", path, i)
			if err := inventory.charge(argumentPath); err != nil {
				return err
			}
			inventory.visited++
			if argument.Value == nil {
				return goalCheckInvalid("MISSING_ATOM_ARGUMENT", argumentPath, "atom argument is missing")
			}
			if argument.Value.Kind == algebra.TermHole {
				inventory.holes = append(inventory.holes, Unresolved{Code: "CANDIDATE_HOLE", Path: argumentPath, Message: "candidate contains an unresolved typed hole"})
				continue
			}
			if argument.Value.Kind != algebra.TermVar && argument.Value.Kind != algebra.TermRef {
				return goalCheckUnsupported("UNSUPPORTED_ATOM_ARGUMENT", argumentPath, "the structural goal profile accepts only Var or Ref atom arguments")
			}
		}
		inventory.atoms = append(inventory.atoms, goalCheckAtom{term: term, path: path})
	case algebra.TermHole:
		inventory.holes = append(inventory.holes, Unresolved{Code: "CANDIDATE_HOLE", Path: path, Message: "candidate contains an unresolved typed hole"})
	case algebra.TermLambda:
		return goalCheckUnsupported("NESTED_LAMBDA_UNSUPPORTED", path, "only the ordered leading goal-port lambdas are supported")
	case algebra.TermApply:
		return goalCheckUnsupported("APPLY_UNSUPPORTED", path, "application is outside the positive structural goal profile")
	case algebra.TermParam:
		return goalCheckUnsupported("FREE_PARAMETER_UNSUPPORTED", path, "free parameters are outside the positive structural goal profile")
	case algebra.TermVar, algebra.TermRef:
		return goalCheckUnsupported("NON_ATOM_PROPOSITION", path, "the proposition body must be an atom, conjunction, or existential quantification")
	default:
		return goalCheckUnsupported("UNSUPPORTED_TERM_KIND", path+".kind", "term kind is outside the positive structural goal profile")
	}
	return nil
}

type goalProfileMatch struct {
	satisfied   bool
	constraints []GoalConstraintStatus
}

type goalCheckEdge struct {
	index int
	from  goalCheckValue
	to    goalCheckValue
	path  string
}

type goalProfileMatcher struct {
	goal        GoalSpec
	lambdaPorts map[string]string
	lambdaTypes map[string]algebra.Type
	inventory   goalCheckInventory
	signatures  map[string]algebra.Signature
	portValues  map[string]goalCheckValue
	usedAtoms   map[int]bool
	work        int
	maxWork     int
}

func newGoalProfileMatcher(context algebra.Document, goal GoalSpec, lambdaPorts map[string]string, lambdaTypes map[string]algebra.Type, inventory goalCheckInventory, remainingWork int) *goalProfileMatcher {
	signatures := make(map[string]algebra.Signature, len(context.Signatures))
	for _, signature := range context.Signatures {
		signatures[signature.ID] = signature
	}
	return &goalProfileMatcher{
		goal: goal, lambdaPorts: lambdaPorts, lambdaTypes: lambdaTypes,
		inventory: inventory, signatures: signatures,
		portValues: make(map[string]goalCheckValue, len(goal.Ports)),
		usedAtoms:  make(map[int]bool, len(inventory.atoms)), maxWork: remainingWork,
	}
}

func (matcher *goalProfileMatcher) charge(path string) error {
	if matcher.work >= matcher.maxWork {
		return goalCheckResource(path, "goal constraint matching exceeds its logical work budget")
	}
	matcher.work++
	return nil
}

func (matcher *goalProfileMatcher) match() (goalProfileMatch, error) {
	for _, port := range matcher.goal.Ports {
		if err := matcher.charge("ports." + port.ID); err != nil {
			return goalProfileMatch{}, err
		}
		if port.Binding != nil {
			if port.Binding.Kind == algebra.TermRef {
				matcher.portValues[port.ID] = goalCheckValue{kind: algebra.TermRef, id: port.Binding.ID}
			}
			continue
		}
		if !goalHasProjectPort(matcher.goal.ProjectPorts, port.ID) {
			binder, exists := matcher.lambdaPorts[port.ID]
			if !exists {
				return goalProfileMatch{}, goalCheckInvalid("GOAL_INTERFACE_MISMATCH", "candidate.interface", "open goal port has no matching leading lambda")
			}
			matcher.portValues[port.ID] = goalCheckValue{kind: algebra.TermVar, id: binder}
		}
	}

	pathIndex := -1
	for i := range matcher.goal.Constraints {
		if matcher.goal.Constraints[i].Kind == ConstraintPath {
			pathIndex = i
			break
		}
	}
	if pathIndex < 0 {
		return goalProfileMatch{}, goalCheckUnsupported("MISSING_PATH_CONSTRAINT", "constraints", "the initial structural profile requires one binary path")
	}
	pathConstraint := matcher.goal.Constraints[pathIndex]
	pathAtoms := make([]int, 0, matcher.goal.Constraints[pathIndex].AllowedAtomCounts[0])
	for i, atom := range matcher.inventory.atoms {
		if err := matcher.charge(atom.path); err != nil {
			return goalProfileMatch{}, err
		}
		if atom.term.Symbol == pathConstraint.Symbol {
			pathAtoms = append(pathAtoms, i)
		}
	}

	pathFragments, endpoints, expectedExists, pathOK, err := matcher.matchPath(pathConstraint, pathAtoms)
	if err != nil {
		return goalProfileMatch{}, err
	}
	if pathOK {
		for _, index := range pathAtoms {
			matcher.usedAtoms[index] = true
		}
	}
	statuses := make([]GoalConstraintStatus, len(matcher.goal.Constraints))
	for i, constraint := range matcher.goal.Constraints {
		if err := matcher.charge(fmt.Sprintf("constraints[%d]", i)); err != nil {
			return goalProfileMatch{}, err
		}
		statuses[i] = GoalConstraintStatus{ConstraintID: constraint.ID, Kind: constraint.Kind, Status: "rejected", Fragments: []string{}}
		if constraint.Kind == ConstraintPath {
			if pathOK {
				statuses[i].Status = "checked"
				statuses[i].Fragments = append([]string{}, pathFragments...)
			} else {
				statuses[i].Code = "PATH_STRUCTURE_MISMATCH"
				statuses[i].Path = fmt.Sprintf("constraints[%d]", i)
				statuses[i].Message = "candidate relation atoms do not form the required directed path with the requested occurrence count"
			}
			continue
		}
		if constraint.Kind != ConstraintUnary {
			return goalProfileMatch{}, goalCheckUnsupported("UNSUPPORTED_CONSTRAINT_KIND", fmt.Sprintf("constraints[%d].kind", i), "constraint kind is outside the structural profile")
		}
		target, known := endpoints[constraint.Port]
		if !pathOK || !known {
			statuses[i].Code = "UNRESOLVED_PROJECTED_ENDPOINT"
			statuses[i].Path = fmt.Sprintf("constraints[%d].port", i)
			statuses[i].Message = "projected endpoint could not be independently identified from the candidate path"
			continue
		}
		atomIndex, findErr := matcher.findUnaryAtom(constraint.Symbol, target)
		if findErr != nil {
			return goalProfileMatch{}, findErr
		}
		if atomIndex >= 0 {
			matcher.usedAtoms[atomIndex] = true
			statuses[i].Status = "checked"
			statuses[i].Fragments = []string{matcher.inventory.atoms[atomIndex].path}
		} else {
			statuses[i].Code = "MISSING_UNARY_OCCURRENCE"
			statuses[i].Path = fmt.Sprintf("constraints[%d]", i)
			statuses[i].Message = "no unused unary atom matches the requested symbol and endpoint value"
		}
	}

	authorizedAtoms := true
	for i := range matcher.inventory.atoms {
		if err := matcher.charge(matcher.inventory.atoms[i].path); err != nil {
			return goalProfileMatch{}, err
		}
		if !matcher.usedAtoms[i] {
			authorizedAtoms = false
			for j := range statuses {
				if statuses[j].Code == "" {
					statuses[j].Status = "rejected"
					statuses[j].Code = "UNAUTHORIZED_ATOM"
					statuses[j].Path = matcher.inventory.atoms[i].path
					statuses[j].Message = "candidate contains an atom not discharged by a goal constraint"
				}
			}
		}
	}

	exactExists := pathOK && len(expectedExists) == len(matcher.inventory.existentials)
	if exactExists {
		for id := range matcher.inventory.existentials {
			if _, expected := expectedExists[id]; !expected {
				exactExists = false
				break
			}
		}
	}
	if pathOK {
		matcher.portValues = endpoints
	}
	if pathOK && !exactExists {
		for i := range statuses {
			if statuses[i].Status == "checked" {
				statuses[i].Status = "rejected"
				statuses[i].Code = "EXISTENTIAL_SCOPE_MISMATCH"
				statuses[i].Path = "candidate.expression"
				statuses[i].Message = "existential binders must be exactly the path witnesses and projected endpoints"
				statuses[i].Fragments = []string{}
			}
		}
	}

	satisfied := pathOK && exactExists && authorizedAtoms
	for _, item := range statuses {
		if item.Status != "checked" {
			satisfied = false
		}
	}
	return goalProfileMatch{satisfied: satisfied, constraints: statuses}, nil
}

// matchPath walks the directed edge chain from the declared source endpoint.
// Projected endpoints are inferred only from a unique degree-one existential
// endpoint of the corresponding type.
func (matcher *goalProfileMatcher) matchPath(constraint Constraint, indices []int) ([]string, map[string]goalCheckValue, map[string]struct{}, bool, error) {
	count := constraint.AllowedAtomCounts[0]
	if len(indices) != count {
		return nil, nil, nil, false, nil
	}
	signature, exists := matcher.signatures[constraint.Symbol]
	if !exists || len(signature.Ports) != 2 {
		return nil, nil, nil, false, nil
	}
	if _, fromExists := signatureRole(signature, constraint.FromRole); !fromExists {
		return nil, nil, nil, false, nil
	}
	if _, toExists := signatureRole(signature, constraint.ToRole); !toExists || constraint.FromRole == constraint.ToRole {
		return nil, nil, nil, false, nil
	}
	edges := make([]goalCheckEdge, 0, len(indices))
	inDegree := make(map[string]int, len(indices)+1)
	outDegree := make(map[string]int, len(indices)+1)
	for _, index := range indices {
		atom := matcher.inventory.atoms[index]
		if err := matcher.charge(atom.path); err != nil {
			return nil, nil, nil, false, err
		}
		if len(atom.term.Args) != 2 {
			return nil, nil, nil, false, nil
		}
		values := make(map[string]goalCheckValue, 2)
		for _, arg := range atom.term.Args {
			if err := matcher.charge(atom.path); err != nil {
				return nil, nil, nil, false, err
			}
			if arg.Value == nil || (arg.Value.Kind != algebra.TermVar && arg.Value.Kind != algebra.TermRef) {
				return nil, nil, nil, false, nil
			}
			values[arg.Role] = goalCheckValue{kind: arg.Value.Kind, id: arg.Value.ID}
		}
		from, hasFrom := values[constraint.FromRole]
		to, hasTo := values[constraint.ToRole]
		if !hasFrom || !hasTo {
			return nil, nil, nil, false, nil
		}
		edges = append(edges, goalCheckEdge{index: index, from: from, to: to, path: atom.path})
		outDegree[from.key()]++
		inDegree[to.key()]++
	}

	start, startKnown := matcher.portValues[constraint.FromPort]
	end, endKnown := matcher.portValues[constraint.ToPort]
	projectedFrom := goalHasProjectPort(matcher.goal.ProjectPorts, constraint.FromPort)
	projectedTo := goalHasProjectPort(matcher.goal.ProjectPorts, constraint.ToPort)
	if !startKnown && !projectedFrom || !endKnown && !projectedTo {
		return nil, nil, nil, false, nil
	}
	if projectedFrom {
		candidate, unique := degreeEndpoint(edges, inDegree, outDegree, true, matcher.goalPortType(constraint.FromPort), matcher.inventory.existentials)
		if !unique {
			return nil, nil, nil, false, nil
		}
		start, startKnown = candidate, true
	}
	if projectedTo {
		candidate, unique := degreeEndpoint(edges, inDegree, outDegree, false, matcher.goalPortType(constraint.ToPort), matcher.inventory.existentials)
		if !unique {
			return nil, nil, nil, false, nil
		}
		end, endKnown = candidate, true
	}
	if !startKnown || !endKnown || !matcher.valueHasGoalType(start, constraint.FromPort) || !matcher.valueHasGoalType(end, constraint.ToPort) {
		return nil, nil, nil, false, nil
	}
	if projectedFrom && projectedTo && start.key() == end.key() {
		return nil, nil, nil, false, nil
	}

	used := make([]bool, len(edges))
	visited := map[string]struct{}{start.key(): {}}
	internal := make(map[string]struct{}, len(edges)-1)
	current := start
	fragments := make([]string, 0, len(edges))
	for step := range edges {
		if err := matcher.charge("candidate.path"); err != nil {
			return nil, nil, nil, false, err
		}
		selected := -1
		for i, edge := range edges {
			if !used[i] && edge.from.key() == current.key() {
				if selected >= 0 {
					return nil, nil, nil, false, nil
				}
				selected = i
			}
		}
		if selected < 0 {
			return nil, nil, nil, false, nil
		}
		used[selected] = true
		edge := edges[selected]
		fragments = append(fragments, edge.path)
		current = edge.to
		if step < len(edges)-1 {
			if current.kind != algebra.TermVar || !matcher.valueIsExistential(current) {
				return nil, nil, nil, false, nil
			}
			if _, repeated := visited[current.key()]; repeated {
				return nil, nil, nil, false, nil
			}
			visited[current.key()] = struct{}{}
			internal[current.id] = struct{}{}
		} else if current.key() == start.key() {
			if start.kind != algebra.TermRef || end.kind != algebra.TermRef || start.id != end.id {
				return nil, nil, nil, false, nil
			}
		} else if _, repeated := visited[current.key()]; repeated {
			return nil, nil, nil, false, nil
		}
	}
	if current.key() != end.key() || len(internal) != len(edges)-1 {
		return nil, nil, nil, false, nil
	}
	for _, wasUsed := range used {
		if !wasUsed {
			return nil, nil, nil, false, nil
		}
	}
	endpoints := make(map[string]goalCheckValue, len(matcher.portValues)+2)
	for id, value := range matcher.portValues {
		endpoints[id] = value
	}
	endpoints[constraint.FromPort], endpoints[constraint.ToPort] = start, end
	expected := make(map[string]struct{}, len(internal)+2)
	for id := range internal {
		expected[id] = struct{}{}
	}
	if projectedFrom {
		expected[start.id] = struct{}{}
	}
	if projectedTo {
		expected[end.id] = struct{}{}
	}
	return fragments, endpoints, expected, true, nil
}

func degreeEndpoint(edges []goalCheckEdge, inDegree, outDegree map[string]int, source bool, typ algebra.Type, existentials map[string]algebra.Type) (goalCheckValue, bool) {
	var found goalCheckValue
	matches := 0
	for _, edge := range edges {
		values := []goalCheckValue{edge.from, edge.to}
		for _, value := range values {
			if value.kind != algebra.TermVar {
				continue
			}
			actual, exists := existentials[value.id]
			if !exists || !algebra.TypeEqual(actual, typ) {
				continue
			}
			if source && outDegree[value.key()] == 1 && inDegree[value.key()] == 0 || !source && inDegree[value.key()] == 1 && outDegree[value.key()] == 0 {
				found = value
				matches++
			}
		}
	}
	return found, matches == 1
}

func (matcher *goalProfileMatcher) valueIsExistential(value goalCheckValue) bool {
	_, exists := matcher.inventory.existentials[value.id]
	return value.kind == algebra.TermVar && exists
}

func (matcher *goalProfileMatcher) valueHasGoalType(value goalCheckValue, portID string) bool {
	typ := matcher.goalPortType(portID)
	switch value.kind {
	case algebra.TermRef:
		for _, reference := range matcher.goal.References {
			if reference.ID == value.id {
				return algebra.TypeEqual(reference.Type, typ)
			}
		}
	case algebra.TermVar:
		if binderType, exists := matcher.lambdaTypes[value.id]; exists {
			return algebra.TypeEqual(binderType, typ)
		}
		if binderType, exists := matcher.inventory.existentials[value.id]; exists {
			return algebra.TypeEqual(binderType, typ)
		}
	}
	return false
}

func (matcher *goalProfileMatcher) goalPortType(portID string) algebra.Type {
	for _, port := range matcher.goal.Ports {
		if port.ID == portID {
			return port.Type
		}
	}
	return algebra.Type{}
}

func (matcher *goalProfileMatcher) findUnaryAtom(symbol string, target goalCheckValue) (int, error) {
	signature, exists := matcher.signatures[symbol]
	if !exists || len(signature.Ports) != 1 {
		return -1, nil
	}
	role := signature.Ports[0].Role
	for i, atom := range matcher.inventory.atoms {
		if err := matcher.charge(atom.path); err != nil {
			return -1, err
		}
		if matcher.usedAtoms[i] || atom.term.Symbol != symbol || len(atom.term.Args) != 1 {
			continue
		}
		arg := atom.term.Args[0]
		if arg.Role == role && arg.Value != nil && (arg.Value.Kind == algebra.TermVar || arg.Value.Kind == algebra.TermRef) &&
			(goalCheckValue{kind: arg.Value.Kind, id: arg.Value.ID}).key() == target.key() {
			return i, nil
		}
	}
	return -1, nil
}

func goalHasProjectPort(projected []string, id string) bool {
	for _, current := range projected {
		if current == id {
			return true
		}
	}
	return false
}

func undeterminedConstraintStatuses(constraints []Constraint, unresolved []Unresolved) []GoalConstraintStatus {
	byID := make(map[string]Unresolved, len(unresolved))
	for _, item := range unresolved {
		if item.ConstraintID != "" {
			byID[item.ConstraintID] = item
		}
	}
	result := make([]GoalConstraintStatus, 0, len(constraints))
	for _, constraint := range constraints {
		item := GoalConstraintStatus{ConstraintID: constraint.ID, Kind: constraint.Kind, Status: "undetermined", Fragments: []string{}}
		if unresolvedItem, exists := byID[constraint.ID]; exists {
			item.Code, item.Path, item.Message = unresolvedItem.Code, unresolvedItem.Path, unresolvedItem.Message
		}
		result = append(result, item)
	}
	return result
}

func rejectGoalCandidate(result GoalConstraintCheck, goal GoalSpec, code, path, message string) GoalConstraintCheck {
	result.Status = status.OutcomeOK
	result.Disposition = GoalCheckDispositionRejected
	result.GoalSatisfied = false
	result.Complete = true
	result.Constraints = make([]GoalConstraintStatus, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		result.Constraints = append(result.Constraints, GoalConstraintStatus{
			ConstraintID: constraint.ID, Kind: constraint.Kind, Status: "rejected", Fragments: []string{},
			Code: code, Path: path, Message: message,
		})
	}
	return result
}

func goalCheckInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func goalCheckUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}

func goalCheckResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

func algebraErrorIsIncomplete(err error) bool {
	var typed *algebra.Error
	return asAlgebraError(err, &typed) && typed.Status == status.OutcomeIncomplete
}

func asAlgebraError(err error, target **algebra.Error) bool {
	if err == nil {
		return false
	}
	value, ok := err.(*algebra.Error)
	if ok {
		*target = value
	}
	return ok
}

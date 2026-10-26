package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// ReplayCertifiedOperations recomputes a bounded sequence of public algebra
// operations. It does not trust claimed result types, interfaces, statuses, or
// step links, and it does not construct a replacement sequence.
func ReplayCertifiedOperations(primitiveContext algebra.Document, concreteGoal GoalSpec, operations []CertifiedOperation, requested GoalVerificationLimits) (algebra.Function, GoalVerificationStatistics, error) {
	var statistics GoalVerificationStatistics
	goalUnresolved, err := ValidateGoal(primitiveContext, concreteGoal)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	if len(goalUnresolved) != 0 {
		return algebra.Function{}, statistics, &algebra.Error{
			Status: status.OutcomeIncomplete, Code: "INCOMPLETE_REPLAY_GOAL",
			Path: "concrete_goal", Message: "operation replay requires a fully selected concrete goal",
		}
	}
	verificationLimits, err := NormalizeGoalVerificationLimits(requested)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	goalLimits, err := NormalizeGoalLimits(concreteGoal.Limits)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	contextLimits, err := algebra.NormalizeLimits(primitiveContext.Limits)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	stepLimit := minInt(verificationLimits.MaxReplaySteps, goalLimits.MaxSteps)
	if len(operations) == 0 {
		return algebra.Function{}, statistics, replayInvalid("MISSING_CERTIFIED_OPERATIONS", "operations", "at least one certified operation is required")
	}
	if len(operations) > stepLimit {
		return algebra.Function{}, statistics, replayResource("operations", "operation count exceeds the effective replay step limit")
	}
	if len(operations) > MaxGoalVerificationMaxReplaySteps {
		return algebra.Function{}, statistics, replayResource("operations", "operation count exceeds the hard replay step limit")
	}

	limits := minimumAlgebraLimits(goalLimits.TermLimits, verificationLimits.TermLimits, contextLimits)
	operatorContext, err := constructionContext(primitiveContext, concreteGoal.References, limits)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	if err := validateCertifiedOperationClaims(primitiveContext, operations, limits); err != nil {
		return algebra.Function{}, statistics, err
	}

	results := make(map[string]algebra.Function, len(operations))
	parents := make(map[string][]string, len(operations))
	orderedIDs := make([]string, 0, len(operations))
	for i, operation := range operations {
		path := fmt.Sprintf("operations[%d]", i)
		if err := checkReplayStepBudget(&statistics, stepLimit, path); err != nil {
			return algebra.Function{}, statistics, err
		}
		if !goalAllowsOperation(concreteGoal, operation.Op) {
			return algebra.Function{}, statistics, replayUnsupported("OPERATOR_NOT_ALLOWED", path+".op", "operation is outside the concrete goal's declared operator ceiling")
		}
		if _, duplicate := results[operation.ID]; duplicate {
			return algebra.Function{}, statistics, replayInvalid("DUPLICATE_OPERATION_ID", path+".id", "operation IDs must be unique")
		}
		inputs := make([]algebra.Function, len(operation.Inputs))
		seenInputs := make(map[string]struct{}, len(operation.Inputs))
		for j, inputID := range operation.Inputs {
			if _, duplicate := seenInputs[inputID]; duplicate {
				return algebra.Function{}, statistics, replayInvalid("DUPLICATE_OPERATION_INPUT", fmt.Sprintf("%s.inputs[%d]", path, j), "an operation cannot list the same input output twice")
			}
			seenInputs[inputID] = struct{}{}
			input, exists := results[inputID]
			if !exists {
				return algebra.Function{}, statistics, replayInvalid("NON_EARLIER_OPERATION_INPUT", fmt.Sprintf("%s.inputs[%d]", path, j), "inputs must reference an earlier operation output")
			}
			inputs[j] = input
		}
		if operation.Op == "permute" && len(inputs) != 1 {
			return algebra.Function{}, statistics, replayInvalid("IRRELEVANT_OPERATION_FIELD", path+".inputs", "permute requires exactly one earlier function input")
		}
		if operation.Op == "permute" && concreteGoalPathCount(concreteGoal) == 1 {
			inputs[0].Interface = append([]algebra.PortBinding{}, inputs[0].Interface...)
			if err := mapSingleRelationGoalAliases(operatorContext, concreteGoal, &inputs[0]); err != nil {
				return algebra.Function{}, statistics, err
			}
		}
		remaining := verificationLimits.MaxWorkNodes - statistics.WorkNodes
		if remaining <= 0 {
			return algebra.Function{}, statistics, replayResource(path, "operation replay exceeds the bounded work-node limit")
		}
		stepContext := operatorContext
		if remaining < stepContext.Limits.MaxNodes {
			stepContext.Limits.MaxNodes = remaining
		}
		if remaining < stepContext.Limits.MaxDepth {
			stepContext.Limits.MaxDepth = remaining
		}
		result, applyErr := applyCertifiedOperation(stepContext, concreteGoal, operation, inputs)
		if applyErr != nil {
			return algebra.Function{}, statistics, applyErr
		}
		checked, checkErr := checkReplayedFunction(stepContext, result)
		if checkErr != nil {
			return algebra.Function{}, statistics, checkErr
		}
		if checked.NodeCount > verificationLimits.MaxWorkNodes-statistics.WorkNodes {
			return algebra.Function{}, statistics, replayResource(path, "operation result exceeds the bounded work-node limit")
		}
		statistics.WorkNodes += checked.NodeCount
		statistics.ReplaySteps++
		if !algebra.TypeEqual(checked.Type, *operation.CheckedType) {
			return algebra.Function{}, statistics, replayInvalid("OPERATION_TYPE_CLAIM_MISMATCH", path+".checked_type", "claimed type differs from the recomputed operation result")
		}
		if !samePortBindings(result.Interface, operation.Interface) {
			return algebra.Function{}, statistics, replayInvalid("OPERATION_INTERFACE_CLAIM_MISMATCH", path+".interface", "claimed interface differs from the recomputed operation result")
		}
		if len(operation.Ports) > 0 && (operation.Op == "relation" || operation.Op == "join" || operation.Op == "restrict" || operation.Op == "bind") &&
			!sameStrings(interfacePorts(result.Interface), operation.Ports) {
			return algebra.Function{}, statistics, replayInvalid("OPERATION_PORTS_CLAIM_MISMATCH", path+".ports", "claimed output ports differ from the recomputed operation result")
		}
		results[operation.ID] = result
		parents[operation.ID] = append([]string{}, operation.Inputs...)
		orderedIDs = append(orderedIDs, operation.ID)
	}

	lastID := operations[len(operations)-1].ID
	if err := ensureReplayInputsReachFinal(lastID, orderedIDs, parents); err != nil {
		return algebra.Function{}, statistics, err
	}
	final := results[lastID]
	if concreteGoalPathCount(concreteGoal) == 1 {
		if err := mapSingleRelationGoalAliases(operatorContext, concreteGoal, &final); err != nil {
			return algebra.Function{}, statistics, err
		}
	}
	remaining := verificationLimits.MaxWorkNodes - statistics.WorkNodes
	if remaining <= 0 {
		return algebra.Function{}, statistics, replayResource("result", "final function check exceeds the bounded work-node limit")
	}
	finalContext := operatorContext
	if remaining < finalContext.Limits.MaxNodes {
		finalContext.Limits.MaxNodes = remaining
	}
	if remaining < finalContext.Limits.MaxDepth {
		finalContext.Limits.MaxDepth = remaining
	}
	checked, err := checkReplayedFunction(finalContext, final)
	if err != nil {
		return algebra.Function{}, statistics, err
	}
	if checked.NodeCount > remaining {
		return algebra.Function{}, statistics, replayResource("result", "final function exceeds the bounded work-node limit")
	}
	statistics.WorkNodes += checked.NodeCount
	if !algebra.TypeEqual(checked.Type, concreteGoal.ExpectedType) {
		return algebra.Function{}, statistics, replayInvalid("REPLAY_GOAL_TYPE_MISMATCH", "result.checked_type", "replayed function type does not match the concrete goal")
	}
	expectedPorts := openGoalPortIDs(concreteGoal.Ports, concreteGoal.ProjectPorts)
	if !sameStrings(interfacePorts(final.Interface), expectedPorts) {
		return algebra.Function{}, statistics, replayInvalid("REPLAY_GOAL_INTERFACE_MISMATCH", "result.interface", "replayed interface does not match the ordered open goal ports")
	}
	return final, statistics, nil
}

func goalAllowsOperation(goal GoalSpec, operation string) bool {
	for _, allowed := range goal.AllowedOperators {
		if allowed == operation {
			return true
		}
	}
	return false
}

func validateCertifiedOperationClaims(context algebra.Document, operations []CertifiedOperation, limits algebra.Limits) error {
	typeDeclarations := make([]algebra.Declaration, 0, len(operations)*2)
	for i, operation := range operations {
		path := fmt.Sprintf("operations[%d]", i)
		if !goalID(operation.ID) {
			return replayInvalid("INVALID_OPERATION_ID", path+".id", "operation ID must be bounded stable text")
		}
		if operation.ClaimedStatus != status.OutcomeOK {
			return replayInvalid("NON_SUCCESS_OPERATION_CLAIM", path+".claimed_status", "only successful operation claims can be replayed")
		}
		if operation.CheckedType == nil {
			return replayInvalid("MISSING_OPERATION_TYPE_CLAIM", path+".checked_type", "each operation must include its claimed output type")
		}
		if len(operation.Inputs) > MaxGoalVerificationMaxReplaySteps || len(operation.Mapping) > MaxGoalPorts || len(operation.Outputs) > MaxGoalPorts ||
			len(operation.Ports) > MaxGoalPorts || len(operation.Interface) > MaxGoalPorts || len(operation.Substitutions) > MaxGoalPorts {
			return replayResource(path, "operation claim exceeds a bounded list limit")
		}
		if len(operation.CheckedType.Kind) > 16 {
			return replayInvalid("INVALID_OPERATION_TYPE_CLAIM", path+".checked_type", "type kind text exceeds its bound")
		}
		typeDeclarations = append(typeDeclarations, algebra.Declaration{ID: fmt.Sprintf("operation-checked-type-%d", i), Type: *operation.CheckedType})
		if !goalID(operation.RuleID) || !goalID(operation.Op) {
			return replayInvalid("INVALID_OPERATION_RULE", path, "operator and rule IDs must be bounded stable text")
		}
		expectedRule := ruleIDForOperation(operation.Op)
		if expectedRule == "" {
			return replayUnsupported("UNSUPPORTED_CERTIFIED_OPERATOR", path+".op", "operator is outside the closed replay set")
		}
		if expectedRule != operation.RuleID {
			return replayInvalid("OPERATION_RULE_MISMATCH", path+".rule_id", "rule ID does not match the closed operator version mapping")
		}
		for j, text := range operation.Inputs {
			if !goalID(text) {
				return replayInvalid("INVALID_OPERATION_INPUT", fmt.Sprintf("%s.inputs[%d]", path, j), "input step ID must be bounded stable text")
			}
		}
		for j, binding := range operation.Interface {
			if !goalID(binding.Port) || !goalID(binding.Binder) {
				return replayInvalid("INVALID_OPERATION_INTERFACE", fmt.Sprintf("%s.interface[%d]", path, j), "interface IDs must be bounded stable text")
			}
		}
		for field, value := range map[string]string{
			"symbol": operation.Symbol, "port": operation.Port,
			"reference_id": operation.ReferenceID, "goal_port_id": operation.GoalPortID,
		} {
			if value != "" && !goalID(value) {
				return replayInvalid("INVALID_OPERATION_FIELD", path+"."+field, "operation field must be bounded stable text")
			}
		}
		for j, pair := range operation.Mapping {
			if !goalID(pair.LeftPort) || !goalID(pair.RightPort) {
				return replayInvalid("INVALID_OPERATION_MAPPING", fmt.Sprintf("%s.mapping[%d]", path, j), "mapping ports must be bounded stable text")
			}
		}
		for j, output := range operation.Outputs {
			if !goalID(output.Side) || !goalID(output.Port) || !goalID(output.As) {
				return replayInvalid("INVALID_OPERATION_OUTPUT", fmt.Sprintf("%s.outputs[%d]", path, j), "output side and port IDs must be bounded stable text")
			}
		}
		for j, port := range operation.Ports {
			if !goalID(port) {
				return replayInvalid("INVALID_OPERATION_PORT", fmt.Sprintf("%s.ports[%d]", path, j), "operation ports must be bounded stable text")
			}
		}
		for j, substitution := range operation.Substitutions {
			if !goalID(substitution.BinderID) || !goalID(substitution.ReferenceID) {
				return replayInvalid("INVALID_OPERATION_SUBSTITUTION", fmt.Sprintf("%s.substitutions[%d]", path, j), "substitution binder and reference IDs must be bounded stable text")
			}
			typeDeclarations = append(typeDeclarations, algebra.Declaration{ID: fmt.Sprintf("operation-substitution-type-%d-%d", i, j), Type: substitution.Type})
		}
		if operation.CheckedType != nil && (operation.Op == "relation" || operation.Op == "join" || operation.Op == "restrict" || operation.Op == "bind" || operation.Op == "project" || operation.Op == "permute") {
			if err := validateOperationFields(operation, path); err != nil {
				return err
			}
		}
	}
	return algebra.CheckDeclarations(algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Sorts: context.Sorts,
		Parameters: typeDeclarations, Limits: limits,
	})
}

func validateOperationFields(operation CertifiedOperation, path string) error {
	switch operation.Op {
	case "relation":
		if operation.Symbol == "" || len(operation.Inputs) != 0 || operation.Port != "" || operation.ReferenceID != "" || operation.GoalPortID != "" ||
			len(operation.Mapping) != 0 || len(operation.Outputs) != 0 || len(operation.Substitutions) != 0 {
			return replayInvalid("IRRELEVANT_OPERATION_FIELD", path, "relation requires only a symbol and optional output-port claim")
		}
	case "join":
		if len(operation.Inputs) != 2 || len(operation.Mapping) == 0 || operation.Symbol != "" || operation.Port != "" || operation.ReferenceID != "" || operation.GoalPortID != "" || len(operation.Substitutions) != 0 {
			return replayInvalid("IRRELEVANT_OPERATION_FIELD", path, "join requires two inputs and an explicit mapping")
		}
	case "restrict":
		if len(operation.Inputs) != 2 || operation.Port == "" || operation.ReferenceID != "" || len(operation.Mapping) != 0 || len(operation.Outputs) != 0 || len(operation.Substitutions) != 0 {
			return replayInvalid("IRRELEVANT_OPERATION_FIELD", path, "restrict requires a function, predicate, and selected port")
		}
	case "bind":
		if len(operation.Inputs) != 1 || operation.Port == "" || operation.ReferenceID == "" || operation.GoalPortID == "" || len(operation.Mapping) != 0 || len(operation.Outputs) != 0 || len(operation.Substitutions) != 1 || operation.Symbol != "" {
			return replayInvalid("IRRELEVANT_OPERATION_FIELD", path, "bind requires one function, port, goal reference, and one binder substitution")
		}
	case "project", "permute":
		if len(operation.Inputs) != 1 || operation.Symbol != "" || operation.Port != "" || operation.ReferenceID != "" || operation.GoalPortID != "" || len(operation.Mapping) != 0 || len(operation.Outputs) != 0 || len(operation.Substitutions) != 0 {
			return replayInvalid("IRRELEVANT_OPERATION_FIELD", path, "project and permute accept one function and an ordered port list")
		}
	default:
		return replayUnsupported("UNSUPPORTED_CERTIFIED_OPERATOR", path+".op", "operator is outside the closed replay set")
	}
	return nil
}

func applyCertifiedOperation(context algebra.Document, goal GoalSpec, operation CertifiedOperation, inputs []algebra.Function) (algebra.Function, error) {
	path := "operations." + operation.ID
	switch operation.Op {
	case "relation":
		return algebra.Relation(context, operation.Symbol)
	case "join":
		return algebra.Join(context, inputs[0], inputs[1], operation.Mapping, operation.Outputs)
	case "restrict":
		predicateSymbol, exists := simplePredicateSymbol(context, inputs[1])
		if operation.Symbol != "" && (!exists || operation.Symbol != predicateSymbol) {
			return algebra.Function{}, replayInvalid("PREDICATE_SYMBOL_CLAIM_MISMATCH", path+".symbol", "restriction symbol differs from the actual unary predicate input")
		}
		if operation.GoalPortID != "" && portAliasForGoal(goal, operation.GoalPortID) != operation.Port {
			return algebra.Function{}, replayInvalid("RESTRICT_GOAL_PORT_MISMATCH", path+".goal_port_id", "restriction port does not map to the claimed goal port")
		}
		return algebra.Restrict(context, inputs[0], operation.Port, inputs[1])
	case "bind":
		goalPort, found := goalPort(goal, operation.GoalPortID)
		if !found || goalPort.Binding == nil || goalPort.Binding.Kind != algebra.TermRef || goalPort.Binding.ID != operation.ReferenceID ||
			portAliasForGoal(goal, operation.GoalPortID) != operation.Port {
			return algebra.Function{}, replayInvalid("BIND_GOAL_REFERENCE_MISMATCH", path, "bind must match an explicitly Ref-bound concrete goal port")
		}
		index := interfaceIndexForReplay(inputs[0].Interface, operation.Port)
		if index < 0 {
			return algebra.Function{}, replayInvalid("UNKNOWN_BIND_PORT", path+".port", "bind port is not in its input interface")
		}
		substitution := operation.Substitutions[0]
		if substitution.BinderID != inputs[0].Interface[index].Binder || substitution.ReferenceID != operation.ReferenceID ||
			!algebra.TypeEqual(substitution.Type, *goalPort.Binding.Type) {
			return algebra.Function{}, replayInvalid("BIND_SUBSTITUTION_MISMATCH", path+".substitutions[0]", "binder substitution must match the selected interface binder and typed goal reference")
		}
		argument := &algebra.Term{Kind: algebra.TermRef, ID: operation.ReferenceID, Type: goalPort.Binding.Type}
		return algebra.Bind(context, inputs[0], operation.Port, argument)
	case "project":
		return algebra.Project(context, inputs[0], operation.Ports)
	case "permute":
		return algebra.Permute(context, inputs[0], operation.Ports)
	default:
		return algebra.Function{}, replayUnsupported("UNSUPPORTED_CERTIFIED_OPERATOR", path+".op", "operator is outside the closed replay set")
	}
}

func checkReplayedFunction(context algebra.Document, function algebra.Function) (algebra.CheckedExpression, error) {
	if function.Expression == nil {
		return algebra.CheckedExpression{}, replayInvalid("MISSING_REPLAY_EXPRESSION", "result.expression", "operator returned no expression")
	}
	document := context
	document.Expression = function.Expression
	document.Closed = false
	checked, err := algebra.Check(document)
	if err != nil {
		return algebra.CheckedExpression{}, err
	}
	term := function.Expression
	typ := checked.Type
	for i, binding := range function.Interface {
		if term == nil || term.Kind != algebra.TermLambda || term.ID != binding.Binder || term.Type == nil ||
			typ.Kind != algebra.TypeArrow || typ.From == nil || typ.To == nil || !algebra.TypeEqual(*typ.From, *term.Type) {
			return algebra.CheckedExpression{}, replayInvalid("REPLAY_INTERFACE_INVALID", fmt.Sprintf("result.interface[%d]", i), "function interface must match its ordered leading lambda binders and types")
		}
		typ = *typ.To
		term = term.Body
	}
	if term == nil || typ.Kind != algebra.TypeProp {
		return algebra.CheckedExpression{}, replayInvalid("REPLAY_INTERFACE_INVALID", "result.interface", "replayed function must return Prop after its declared leading lambdas")
	}
	if checked.Status != status.OutcomeOK {
		return algebra.CheckedExpression{}, &algebra.Error{Status: checked.Status, Code: "INCOMPLETE_REPLAY_FUNCTION", Path: "result.expression", Message: "replayed function contains an incomplete expression"}
	}
	return checked, nil
}

func mapSingleRelationGoalAliases(context algebra.Document, goal GoalSpec, function *algebra.Function) error {
	if function == nil || concreteGoalPathCount(goal) != 1 {
		return nil
	}
	path, err := uniquePath(goal.Constraints)
	if err != nil {
		return err
	}
	if len(path.AllowedAtomCounts) != 1 || path.AllowedAtomCounts[0] != 1 {
		return nil
	}
	signature, exists := goalSignature(context.Signatures, path.Symbol)
	if !exists {
		return replayInvalid("UNKNOWN_REPLAY_PATH_SYMBOL", "concrete_goal.constraints", "path relation is not declared in the replay context")
	}
	if _, exists := signatureRole(signature, path.FromRole); !exists {
		return replayInvalid("UNKNOWN_REPLAY_PATH_ROLE", "concrete_goal.constraints.from_role", "path role is not declared in the relation")
	}
	if _, exists := signatureRole(signature, path.ToRole); !exists {
		return replayInvalid("UNKNOWN_REPLAY_PATH_ROLE", "concrete_goal.constraints.to_role", "path role is not declared in the relation")
	}
	for i := range function.Interface {
		binding := &function.Interface[i]
		switch binding.Port {
		case path.FromRole:
			binding.Port = path.FromPort
		case path.ToRole:
			binding.Port = path.ToPort
		}
	}
	return nil
}

func concreteGoalPathCount(goal GoalSpec) int {
	count := 0
	for _, constraint := range goal.Constraints {
		if constraint.Kind == ConstraintPath {
			count++
		}
	}
	return count
}

func portAliasForGoal(goal GoalSpec, goalPortID string) string {
	path, err := uniquePath(goal.Constraints)
	if err != nil || len(path.AllowedAtomCounts) != 1 {
		return ""
	}
	if path.AllowedAtomCounts[0] == 1 {
		switch goalPortID {
		case path.FromPort:
			return path.FromRole
		case path.ToPort:
			return path.ToRole
		}
	}
	return goalPortID
}

func simplePredicateSymbol(context algebra.Document, predicate algebra.Function) (string, bool) {
	if len(predicate.Interface) != 1 {
		return "", false
	}
	term := predicate.Expression
	if term == nil || term.Kind != algebra.TermLambda {
		return "", false
	}
	term = term.Body
	if term == nil || term.Kind != algebra.TermAtom || len(term.Args) != 1 {
		return "", false
	}
	signature, exists := goalSignature(context.Signatures, term.Symbol)
	return term.Symbol, exists && len(signature.Ports) == 1
}

func ensureReplayInputsReachFinal(finalID string, orderedIDs []string, parents map[string][]string) error {
	reachable := make(map[string]struct{}, len(orderedIDs))
	stack := []string{finalID}
	for len(stack) != 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, visited := reachable[id]; visited {
			continue
		}
		reachable[id] = struct{}{}
		stack = append(stack, parents[id]...)
	}
	if len(reachable) != len(orderedIDs) {
		return replayInvalid("UNUSED_OPERATION_OUTPUT", "operations", "every operation output must contribute to the final replay result")
	}
	return nil
}

func samePortBindings(left, right []algebra.PortBinding) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func interfaceIndexForReplay(bindings []algebra.PortBinding, port string) int {
	for i, binding := range bindings {
		if binding.Port == port {
			return i
		}
	}
	return -1
}

func ruleIDForOperation(op string) string {
	switch op {
	case "relation":
		return CertificateRuleRelation
	case "join":
		return CertificateRuleJoin
	case "restrict":
		return CertificateRuleRestrict
	case "bind":
		return CertificateRuleBind
	case "project":
		return CertificateRuleProject
	case "permute":
		return CertificateRulePermute
	default:
		return ""
	}
}

func checkReplayStepBudget(statistics *GoalVerificationStatistics, maximum int, path string) error {
	if statistics.ReplaySteps >= maximum {
		return replayResource(path, "operation replay exceeds the bounded step limit")
	}
	return nil
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func replayInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func replayUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}

func replayResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

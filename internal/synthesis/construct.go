package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// Construct lowers a fully specified bounded goal through the public algebra
// operators. It constructs a typed representation; it does not evaluate facts
// or establish that the represented proposition is true.
func Construct(context algebra.Document, goal GoalSpec) (Result, error) {
	unresolved, err := ValidateGoal(context, goal)
	if err != nil {
		return Result{}, err
	}
	if len(unresolved) != 0 {
		return Result{Status: status.OutcomeIncomplete, Trace: []algebra.BuildTrace{}, Unresolved: unresolved, Obligations: []GoalObligation{}}, nil
	}
	limits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return Result{}, err
	}
	path, err := uniquePath(goal.Constraints)
	if err != nil {
		return Result{}, err
	}
	count := path.AllowedAtomCounts[0]
	signature, ok := goalSignature(context.Signatures, path.Symbol)
	if !ok {
		return Result{}, goalInvalid("UNKNOWN_RELATION_SYMBOL", "constraints", "relation symbol is not declared in the selected context")
	}
	_, okFrom := signatureRole(signature, path.FromRole)
	_, okTo := signatureRole(signature, path.ToRole)
	if !okFrom || !okTo {
		return Result{}, goalInvalid("UNKNOWN_RELATION_ROLE", "constraints", "path role is not declared by its relation")
	}

	operatorContext, err := constructionContext(context, goal.References, limits.TermLimits, goal.RequestContext)
	if err != nil {
		return Result{}, err
	}
	stepLimit := limits.MaxSteps
	stepCount := 0
	trace := make([]algebra.BuildTrace, 0, stepLimit)
	stepID := func() string { return fmt.Sprintf("s%d", stepCount) }
	consume := func() error {
		if stepCount >= stepLimit {
			return goalResource("limits.max_steps", "construction exceeded the bounded operator-step budget")
		}
		stepCount++
		return nil
	}
	appendTrace := func(op, symbol, resultKind string, inputs []string, mapping []algebra.JoinPair, outputs []algebra.JoinOutput, port string, ports []string, fn algebra.Function) string {
		id := stepID()
		checkedType := functionType(fn)
		trace = append(trace, algebra.BuildTrace{
			ID: id, Op: op, Symbol: symbol, Inputs: append([]string(nil), inputs...),
			Mapping: append([]algebra.JoinPair(nil), mapping...), Outputs: append([]algebra.JoinOutput(nil), outputs...),
			Port: port, Ports: append([]string(nil), ports...), ResultKind: resultKind,
			Status: status.OutcomeOK, CheckedType: &checkedType,
			Interface: append([]algebra.PortBinding(nil), fn.Interface...),
		})
		return id
	}

	// Construct the first relation occurrence.
	if err := consume(); err != nil {
		return Result{}, err
	}
	function, err := algebra.Relation(operatorContext, path.Symbol)
	if err != nil {
		return Result{}, err
	}
	previousID := appendTrace("relation", path.Symbol, "function", nil, nil, nil, "", interfacePorts(function.Interface), function)
	pathStepIDs := []string{previousID}
	portAliases := make(map[string]string, 2)
	var current algebra.Function
	if count == 1 {
		portAliases[path.FromPort] = path.FromRole
		portAliases[path.ToPort] = path.ToRole
		current = function
	} else {
		for occurrence := 1; occurrence < count; occurrence++ {
			if err := consume(); err != nil {
				return Result{}, err
			}
			next, relationErr := algebra.Relation(operatorContext, path.Symbol)
			if relationErr != nil {
				return Result{}, relationErr
			}
			nextID := appendTrace("relation", path.Symbol, "function", nil, nil, nil, "", interfacePorts(next.Interface), next)
			pathStepIDs = append(pathStepIDs, nextID)
			mapping := []algebra.JoinPair{{LeftPort: path.ToRole, RightPort: path.FromRole}}
			outputs := []algebra.JoinOutput{{Side: "left", Port: path.FromRole, As: path.FromPort}, {Side: "right", Port: path.ToRole, As: path.ToPort}}
			leftID := previousID
			if occurrence > 1 {
				mapping[0].LeftPort = path.ToPort
				outputs = []algebra.JoinOutput{{Side: "left", Port: path.FromPort, As: path.FromPort}, {Side: "right", Port: path.ToRole, As: path.ToPort}}
			}
			if err := consume(); err != nil {
				return Result{}, err
			}
			joined, joinErr := algebra.Join(operatorContext, function, next, mapping, outputs)
			if joinErr != nil {
				return Result{}, joinErr
			}
			previousID = appendTrace("join", "", "function", []string{leftID, nextID}, mapping, outputs, "", interfacePorts(joined.Interface), joined)
			pathStepIDs = append(pathStepIDs, previousID)
			function = joined
		}
		current = function
		portAliases[path.FromPort] = path.FromPort
		portAliases[path.ToPort] = path.ToPort
	}

	obligations := make([]GoalObligation, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		if constraint.Kind != ConstraintUnary {
			continue
		}
		if err := consume(); err != nil {
			return Result{}, err
		}
		predicate, relationErr := algebra.Relation(operatorContext, constraint.Symbol)
		if relationErr != nil {
			return Result{}, relationErr
		}
		predicateID := appendTrace("relation", constraint.Symbol, "function", nil, nil, nil, "", interfacePorts(predicate.Interface), predicate)
		if err := consume(); err != nil {
			return Result{}, err
		}
		port := portAliases[constraint.Port]
		if port == "" {
			return Result{}, goalInvalid("UNKNOWN_GOAL_PORT", "constraints", "unary restriction port has no path interface mapping")
		}
		priorID := previousID
		current, err = algebra.Restrict(operatorContext, current, port, predicate)
		if err != nil {
			return Result{}, err
		}
		previousID = appendTrace("restrict", constraint.Symbol, "function", []string{priorID, predicateID}, nil, nil, port, interfacePorts(current.Interface), current)
		obligations = append(obligations, GoalObligation{ConstraintID: constraint.ID, Kind: constraint.Kind, Status: "constructed_provenance", StepIDs: []string{predicateID, previousID}, Fragments: []string{}})
	}

	for i, goalPort := range goal.Ports {
		if goalPort.Binding == nil || goalPort.Binding.Kind != algebra.TermRef {
			continue
		}
		if err := consume(); err != nil {
			return Result{}, err
		}
		port := portAliases[goalPort.ID]
		if port == "" {
			return Result{}, goalInvalid("UNKNOWN_GOAL_PORT", fmt.Sprintf("ports[%d]", i), "bound goal port has no path interface mapping")
		}
		argument := &algebra.Term{Kind: algebra.TermRef, ID: goalPort.Binding.ID, Type: goalPort.Binding.Type}
		priorID := previousID
		current, err = algebra.Bind(operatorContext, current, port, argument)
		if err != nil {
			return Result{}, err
		}
		previousID = appendTrace("bind", "", "function", []string{priorID}, nil, nil, port, interfacePorts(current.Interface), current)
		portAliases[goalPort.ID] = ""
	}

	// Close explicitly projected ports only after all requested restrictions
	// and reference bindings have been applied.
	if len(goal.ProjectPorts) > 0 {
		retainedPorts := make([]string, 0, len(goal.Ports)-len(goal.ProjectPorts))
		projected := make(map[string]struct{}, len(goal.ProjectPorts))
		for _, portID := range goal.ProjectPorts {
			projected[portID] = struct{}{}
		}
		for _, goalPort := range goal.Ports {
			if goalPort.Binding != nil {
				continue
			}
			if _, closePort := projected[goalPort.ID]; closePort {
				continue
			}
			actualPort := portAliases[goalPort.ID]
			if actualPort == "" {
				return Result{}, goalInvalid("UNKNOWN_GOAL_PORT", "project_ports", "retained goal port has no constructed interface mapping")
			}
			retainedPorts = append(retainedPorts, actualPort)
		}
		if err := consume(); err != nil {
			return Result{}, err
		}
		priorID := previousID
		current, err = algebra.Project(operatorContext, current, retainedPorts)
		if err != nil {
			return Result{}, err
		}
		previousID = appendTrace("project", "", "function", []string{priorID}, nil, nil, "", retainedPorts, current)
	}

	// Map open interface names to goal ports, then permute only when both open
	// ports remain and their actual order differs from the requested order.
	for i := range current.Interface {
		for goalPort, actualPort := range portAliases {
			if actualPort == current.Interface[i].Port {
				current.Interface[i].Port = goalPort
			}
		}
	}
	desired := openGoalPortIDs(goal.Ports, goal.ProjectPorts)
	currentPorts := interfacePorts(current.Interface)
	if len(desired) == 2 && len(currentPorts) == 2 && (currentPorts[0] != desired[0] || currentPorts[1] != desired[1]) {
		actualOrder := make([]string, 0, 2)
		for _, id := range desired {
			alias := interfacePortForGoal(current.Interface, id)
			if alias == "" {
				return Result{}, goalInvalid("GOAL_INTERFACE_MISMATCH", "ports", "open goal port is absent from the constructed interface")
			}
			actualOrder = append(actualOrder, alias)
		}
		// current.Interface has already been annotated with goal IDs, so these
		// are stable goal-port names and carry the intended output mapping.
		if err := consume(); err != nil {
			return Result{}, err
		}
		priorID := previousID
		current, err = algebra.Permute(operatorContext, current, actualOrder)
		if err != nil {
			return Result{}, err
		}
		previousID = appendTrace("permute", "", "function", []string{priorID}, nil, nil, "", actualOrder, current)
	}

	// Restrict the exported interface to the declared goal order without
	// changing the AST or its binder IDs.
	expectedInterface := openGoalPortIDs(goal.Ports, goal.ProjectPorts)
	if len(current.Interface) != len(expectedInterface) {
		return Result{}, goalInvalid("GOAL_INTERFACE_MISMATCH", "ports", "constructed function has a different number of open goal ports")
	}
	for i, portID := range expectedInterface {
		if current.Interface[i].Port != portID {
			return Result{}, goalInvalid("GOAL_INTERFACE_MISMATCH", "ports", "constructed function interface order differs from the goal")
		}
	}
	if err := algebra.ValidateFunction(operatorContext, current); err != nil {
		return Result{}, err
	}

	document := operatorContext
	document.Expression = current.Expression
	document.Closed = false
	checked, err := algebra.Check(document)
	if err != nil {
		return Result{}, err
	}
	if !algebra.TypeEqual(checked.Type, goal.ExpectedType) {
		return Result{}, goalInvalid("CONSTRUCTED_TYPE_MISMATCH", "expected_type", "constructed expression type differs from the requested goal type")
	}
	obligations = append(obligations, GoalObligation{ConstraintID: path.ID, Kind: ConstraintPath, Status: "constructed_provenance", StepIDs: append([]string(nil), pathStepIDs...), Fragments: []string{}})
	obligations, err = mapConstructionFragments(goal, current.Expression, obligations)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Status: status.OutcomeOK, Expression: current.Expression, Checked: &checked,
		Interface: append([]algebra.PortBinding(nil), current.Interface...), CanonicalDocument: &document,
		Trace: trace, Unresolved: []Unresolved{}, Obligations: obligations,
	}, nil
}

func uniquePath(constraints []Constraint) (Constraint, error) {
	var path Constraint
	count := 0
	for _, constraint := range constraints {
		if constraint.Kind == ConstraintPath {
			path = constraint
			count++
		}
	}
	if count != 1 {
		return Constraint{}, goalInvalid("MISSING_OR_MULTIPLE_PATHS", "constraints", "construction requires exactly one validated path")
	}
	return path, nil
}

func constructionContext(context algebra.Document, references []algebra.Declaration, limits algebra.Limits, requestContext *algebra.RequestContext) (algebra.Document, error) {
	sorts := make([]algebra.SortDeclaration, len(context.Sorts))
	copy(sorts, context.Sorts)
	signatures := make([]algebra.Signature, len(context.Signatures))
	for i, signature := range context.Signatures {
		signatures[i] = signature
		signatures[i].Ports = make([]algebra.Port, len(signature.Ports))
		copy(signatures[i].Ports, signature.Ports)
		for j := range signatures[i].Ports {
			signatures[i].Ports[j].Type = cloneGoalType(signature.Ports[j].Type)
		}
	}
	refs := make([]algebra.Declaration, len(references))
	for i, ref := range references {
		refs[i] = algebra.Declaration{ID: ref.ID, Type: cloneGoalType(ref.Type)}
	}
	var contextCopy *algebra.RequestContext
	if requestContext != nil {
		cloned, cloneErr := algebra.CloneRequestContext(*requestContext)
		if cloneErr != nil {
			return algebra.Document{}, cloneErr
		}
		contextCopy = &cloned
	}
	document := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: sorts, Signatures: signatures, References: refs, RequestContext: contextCopy, Limits: limits}
	if err := algebra.CheckDeclarations(document); err != nil {
		return algebra.Document{}, err
	}
	return document, nil
}

func cloneGoalType(value algebra.Type) algebra.Type {
	copyValue := value
	if value.From != nil {
		from := cloneGoalType(*value.From)
		copyValue.From = &from
	}
	if value.To != nil {
		to := cloneGoalType(*value.To)
		copyValue.To = &to
	}
	return copyValue
}

func functionType(function algebra.Function) algebra.Type {
	var types []algebra.Type
	term := function.Expression
	for term != nil && term.Kind == algebra.TermLambda {
		if term.Type == nil {
			break
		}
		types = append(types, *term.Type)
		term = term.Body
	}
	out := algebra.Type{Kind: algebra.TypeProp}
	for i := len(types) - 1; i >= 0; i-- {
		from := cloneGoalType(types[i])
		prior := out
		out = algebra.Type{Kind: algebra.TypeArrow, From: &from, To: &prior}
	}
	return out
}

func interfacePorts(bindings []algebra.PortBinding) []string {
	ports := make([]string, len(bindings))
	for i, binding := range bindings {
		ports[i] = binding.Port
	}
	return ports
}

func openGoalPortIDs(ports []GoalPort, projectPorts []string) []string {
	projected := make(map[string]struct{}, len(projectPorts))
	for _, portID := range projectPorts {
		projected[portID] = struct{}{}
	}
	ids := make([]string, 0, len(ports))
	for _, port := range ports {
		_, closesPort := projected[port.ID]
		if port.Binding == nil && !closesPort {
			ids = append(ids, port.ID)
		}
	}
	return ids
}

func interfacePortForGoal(bindings []algebra.PortBinding, id string) string {
	for _, binding := range bindings {
		if binding.Port == id {
			return binding.Port
		}
	}
	return ""
}

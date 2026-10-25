package algebra

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/status"
)

// PortBinding connects a declared relation port to a leading lambda binder.
// Its position in Function.Interface is significant.
type PortBinding struct {
	Port   string `json:"port"`
	Binder string `json:"binder"`
}

// Function is a checked relation expression with an explicit ordered
// port-to-binder interface. It is data, not executable host code.
type Function struct {
	Expression *Term         `json:"expression"`
	Interface  []PortBinding `json:"interface"`
}

type JoinPair struct {
	LeftPort  string `json:"left_port"`
	RightPort string `json:"right_port"`
}

type JoinOutput struct {
	Side string `json:"side"` // "left" or "right"
	Port string `json:"port"`
	As   string `json:"as"`
}

// ValidateFunction checks the complete expression and verifies that its
// interface names exactly the ordered leading lambdas and their types.
func ValidateFunction(context Document, function Function) error {
	if function.Expression == nil {
		return invalid("MISSING_FUNCTION", "function.expression", "function expression is required")
	}
	expressionType, err := checkExpression(context, function.Expression)
	if err != nil {
		return err
	}
	seenPorts := make(map[string]bool, len(function.Interface))
	seenBinders := make(map[string]bool, len(function.Interface))
	term := function.Expression
	for i, binding := range function.Interface {
		path := fmt.Sprintf("function.interface[%d]", i)
		if !validText(binding.Port, 256) || !validText(binding.Binder, 256) {
			return invalid("INVALID_FUNCTION_INTERFACE", path, "port and binder IDs must be non-empty bounded text")
		}
		if seenPorts[binding.Port] || seenBinders[binding.Binder] {
			return invalid("DUPLICATE_FUNCTION_INTERFACE", path, "port and binder mappings must be one-to-one")
		}
		seenPorts[binding.Port], seenBinders[binding.Binder] = true, true
		if term == nil || term.Kind != TermLambda || term.ID != binding.Binder {
			return invalid("FUNCTION_INTERFACE_ORDER", path, "interface entries must match each leading lambda in order")
		}
		if expressionType.Kind != TypeArrow || expressionType.From == nil || expressionType.To == nil || term.Type == nil || !TypeEqual(*expressionType.From, *term.Type) {
			return invalid("FUNCTION_INTERFACE_TYPE", path, "interface binder type differs from the checked leading lambda type")
		}
		expressionType = *expressionType.To
		term = term.Body
	}
	if term == nil {
		return invalid("MISSING_FUNCTION_BODY", "function.expression", "function has no body")
	}
	if expressionType.Kind != TypeProp {
		return invalid("FUNCTION_RESULT_NOT_PROP", "function.expression", "function body must return Prop")
	}
	return nil
}

// Relation eta-expands a declared signature into an anonymous relation
// function. The caller is responsible for selecting only approved signatures.
func Relation(context Document, symbol string) (Function, error) {
	if err := CheckDeclarations(context); err != nil {
		return Function{}, err
	}
	var signature *Signature
	for i := range context.Signatures {
		if context.Signatures[i].ID == symbol {
			signature = &context.Signatures[i]
			break
		}
	}
	if signature == nil {
		return Function{}, invalid("UNKNOWN_RELATION", "symbol", "relation signature is not declared")
	}
	args := make([]Argument, len(signature.Ports))
	bindings := make([]PortBinding, len(signature.Ports))
	used := make(map[string]bool)
	for i, port := range signature.Ports {
		binder := freshOperatorID("rel_"+symbol+"_"+port.Role, used)
		args[i] = Argument{Role: port.Role, Value: &Term{Kind: TermVar, ID: binder}}
		bindings[i] = PortBinding{Port: port.Role, Binder: binder}
	}
	body := &Term{Kind: TermAtom, Symbol: symbol, Args: args}
	expression := wrapLambdas(bindings, signature.Ports, body)
	function := Function{Expression: expression, Interface: bindings}
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	return function, nil
}

// Bind beta-reduces one named interface port and removes only that port.
func Bind(context Document, function Function, port string, argument *Term) (Function, error) {
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	if argument == nil {
		return Function{}, invalid("MISSING_ARGUMENT", "argument", "argument term is required")
	}
	argType, err := checkExpression(context, argument)
	if err != nil {
		return Function{}, err
	}
	index := interfaceIndex(function.Interface, port)
	if index < 0 {
		return Function{}, invalid("UNKNOWN_FUNCTION_PORT", "port", "port is not in the function interface")
	}
	binders, body, err := peelLeading(function.Expression, len(function.Interface))
	if err != nil {
		return Function{}, err
	}
	if !TypeEqual(*binders[index].Type, argType) {
		return Function{}, invalid("ARGUMENT_TYPE_MISMATCH", "argument", "argument type does not match selected port")
	}
	replaced, err := SubstituteBinder(body, function.Interface[index].Binder, argument, context.Limits)
	if err != nil {
		return Function{}, err
	}
	newBindings := append([]PortBinding(nil), function.Interface[:index]...)
	newBindings = append(newBindings, function.Interface[index+1:]...)
	newPorts := make([]Port, 0, len(newBindings))
	for i, binding := range function.Interface {
		if i != index {
			newPorts = append(newPorts, Port{Role: binding.Port, Type: *cloneTypeValue(*binders[i].Type)})
		}
	}
	result := Function{Expression: wrapLambdas(newBindings, newPorts, replaced), Interface: newBindings}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

// Apply reduces a known lambda application. It refuses symbolic function
// variables rather than attempting evaluation or search.
func Apply(context Document, fnTerm, argument *Term) (*Term, error) {
	if fnTerm == nil || argument == nil {
		return nil, invalid("MISSING_ARGUMENT", "application", "function and argument terms are required")
	}
	application := &Term{Kind: TermApply, Function: fnTerm, Argument: argument}
	if _, err := checkExpression(context, application); err != nil {
		return nil, err
	}
	if fnTerm.Kind != TermLambda {
		return nil, &Error{Status: status.OutcomeUnsupported, Code: "SYMBOLIC_APPLICATION", Path: "function", Message: "application can only reduce a known Lambda term"}
	}
	result, err := SubstituteBinder(fnTerm.Body, fnTerm.ID, argument, context.Limits)
	if err != nil {
		return nil, err
	}
	if _, err := checkExpression(context, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Join combines two relations through explicit equal-typed port pairs. It
// adds existential witnesses for paired ports and preserves other ports.
func Join(context Document, left, right Function, mapping []JoinPair, outputs []JoinOutput) (Function, error) {
	limits, err := NormalizeLimits(context.Limits)
	if err != nil {
		return Function{}, err
	}
	if len(mapping) == 0 {
		return Function{}, &Error{Status: status.OutcomeIncomplete, Code: "MISSING_JOIN_MAPPING", Path: "mapping", Message: "an explicit port pair is required to join relations"}
	}
	if len(mapping) > limits.MaxNodes || len(outputs) > limits.MaxNodes {
		return Function{}, resourceLimit("mapping", "join interface mapping exceeds node budget")
	}
	if err := ValidateFunction(context, left); err != nil {
		return Function{}, err
	}
	if err := ValidateFunction(context, right); err != nil {
		return Function{}, err
	}
	leftCopy, leftIDs, err := FreshCopy(left.Expression, termIDs(right.Expression), limits)
	if err != nil {
		return Function{}, err
	}
	rightCopy, rightIDs, err := FreshCopy(right.Expression, termIDs(leftCopy), limits)
	if err != nil {
		return Function{}, err
	}
	leftBindings := remapInterface(left.Interface, leftIDs)
	rightBindings := remapInterface(right.Interface, rightIDs)
	leftBinders, leftBody, err := peelLeading(leftCopy, len(leftBindings))
	if err != nil {
		return Function{}, err
	}
	rightBinders, rightBody, err := peelLeading(rightCopy, len(rightBindings))
	if err != nil {
		return Function{}, err
	}
	leftTaken, rightTaken := make(map[string]bool), make(map[string]bool)
	shared := make([]*Term, 0, len(mapping))
	reserved := termIDs(leftCopy)
	for _, id := range termIDs(rightCopy) {
		reserved = append(reserved, id)
	}
	for i, pair := range mapping {
		li, ri := interfaceIndex(leftBindings, pair.LeftPort), interfaceIndex(rightBindings, pair.RightPort)
		if li < 0 || ri < 0 {
			return Function{}, invalid("UNKNOWN_JOIN_PORT", fmt.Sprintf("mapping[%d]", i), "join mapping names a port outside an input interface")
		}
		if leftTaken[pair.LeftPort] || rightTaken[pair.RightPort] {
			return Function{}, invalid("DUPLICATE_JOIN_PORT", fmt.Sprintf("mapping[%d]", i), "each input port may be paired at most once")
		}
		leftTaken[pair.LeftPort], rightTaken[pair.RightPort] = true, true
		if !TypeEqual(*leftBinders[li].Type, *rightBinders[ri].Type) {
			return Function{}, invalid("JOIN_TYPE_MISMATCH", fmt.Sprintf("mapping[%d]", i), "paired ports must have exactly equal types")
		}
		id := freshOperatorID(fmt.Sprintf("join_%d", i), reservedMap(reserved))
		reserved = append(reserved, id)
		witness := &Term{Kind: TermVar, ID: id}
		leftBody, err = SubstituteBinder(leftBody, leftBindings[li].Binder, witness, context.Limits)
		if err != nil {
			return Function{}, err
		}
		rightBody, err = SubstituteBinder(rightBody, rightBindings[ri].Binder, witness, context.Limits)
		if err != nil {
			return Function{}, err
		}
		shared = append(shared, &Term{Kind: TermExists, ID: id, Type: cloneTypeValue(*leftBinders[li].Type), Body: nil})
	}
	remaining := make([]PortBinding, 0, len(leftBindings)+len(rightBindings)-2*len(mapping))
	remainingPorts := make([]Port, 0, cap(remaining))
	remainingSides := make([]string, 0, cap(remaining))
	for i, binding := range leftBindings {
		if !leftTaken[binding.Port] {
			remaining = append(remaining, binding)
			remainingPorts = append(remainingPorts, Port{Role: binding.Port, Type: *cloneTypeValue(*leftBinders[i].Type)})
			remainingSides = append(remainingSides, "left")
		}
	}
	for i, binding := range rightBindings {
		if !rightTaken[binding.Port] {
			remaining = append(remaining, binding)
			remainingPorts = append(remainingPorts, Port{Role: binding.Port, Type: *cloneTypeValue(*rightBinders[i].Type)})
			remainingSides = append(remainingSides, "right")
		}
	}
	remaining, remainingPorts, err = applyJoinOutputs(remaining, remainingPorts, remainingSides, outputs)
	if err != nil {
		return Function{}, err
	}
	body := &Term{Kind: TermAnd, Operands: []*Term{leftBody, rightBody}}
	for i := len(shared) - 1; i >= 0; i-- {
		shared[i].Body = body
		body = shared[i]
	}
	result := Function{Expression: wrapLambdas(remaining, remainingPorts, body), Interface: remaining}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

// Restrict conjoins a unary predicate applied to one selected port variable.
func Restrict(context Document, function Function, port string, predicate Function) (Function, error) {
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	if err := ValidateFunction(context, predicate); err != nil {
		return Function{}, err
	}
	predicateExpression, predicateIDs, err := FreshCopy(predicate.Expression, termIDs(function.Expression), context.Limits)
	if err != nil {
		return Function{}, err
	}
	predicateBindings := remapInterface(predicate.Interface, predicateIDs)
	if len(predicate.Interface) != 1 {
		return Function{}, invalid("PREDICATE_ARITY", "predicate.interface", "restriction predicate must be unary")
	}
	selected := interfaceIndex(function.Interface, port)
	if selected < 0 {
		return Function{}, invalid("UNKNOWN_FUNCTION_PORT", "port", "port is not in the function interface")
	}
	fnBinders, fnBody, err := peelLeading(function.Expression, len(function.Interface))
	if err != nil {
		return Function{}, err
	}
	predBinders, predBody, err := peelLeading(predicateExpression, 1)
	if err != nil {
		return Function{}, err
	}
	if !TypeEqual(*fnBinders[selected].Type, *predBinders[0].Type) {
		return Function{}, invalid("PREDICATE_TYPE_MISMATCH", "predicate.interface[0]", "predicate port type must match selected function port")
	}
	argument := &Term{Kind: TermVar, ID: function.Interface[selected].Binder}
	predBody, err = SubstituteBinder(predBody, predicateBindings[0].Binder, argument, context.Limits)
	if err != nil {
		return Function{}, err
	}
	body := &Term{Kind: TermAnd, Operands: []*Term{fnBody, predBody}}
	portTypes := make([]Port, len(function.Interface))
	for i, binding := range function.Interface {
		portTypes[i] = Port{Role: binding.Port, Type: *cloneTypeValue(*fnBinders[i].Type)}
	}
	result := Function{Expression: wrapLambdas(function.Interface, portTypes, body), Interface: append([]PortBinding(nil), function.Interface...)}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

// Project retains an explicit subset of ports in their original relative
// order and existentially binds every omitted interface variable.
func Project(context Document, function Function, ports []string) (Function, error) {
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	limits, err := NormalizeLimits(context.Limits)
	if err != nil {
		return Function{}, err
	}
	if len(ports) > limits.MaxNodes {
		return Function{}, resourceLimit("ports", "projected port list exceeds node budget")
	}
	keep := make(map[string]bool, len(ports))
	for i, port := range ports {
		if interfaceIndex(function.Interface, port) < 0 {
			return Function{}, invalid("UNKNOWN_FUNCTION_PORT", fmt.Sprintf("ports[%d]", i), "projected port is not in the function interface")
		}
		if keep[port] {
			return Function{}, invalid("DUPLICATE_PROJECT_PORT", fmt.Sprintf("ports[%d]", i), "projected ports must be unique")
		}
		keep[port] = true
	}
	binders, body, err := peelLeading(function.Expression, len(function.Interface))
	if err != nil {
		return Function{}, err
	}
	remaining, remainingPorts := make([]PortBinding, 0, len(ports)), make([]Port, 0, len(ports))
	for i, binding := range function.Interface {
		if keep[binding.Port] {
			remaining = append(remaining, binding)
			remainingPorts = append(remainingPorts, Port{Role: binding.Port, Type: *cloneTypeValue(*binders[i].Type)})
		} else {
			body = &Term{Kind: TermExists, ID: binding.Binder, Type: cloneTypeValue(*binders[i].Type), Body: body}
		}
	}
	result := Function{Expression: wrapLambdas(remaining, remainingPorts, body), Interface: remaining}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

// Permute reorders all interface ports according to an explicit bijection.
func Permute(context Document, function Function, orderedPorts []string) (Function, error) {
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	limits, err := NormalizeLimits(context.Limits)
	if err != nil {
		return Function{}, err
	}
	if len(orderedPorts) > limits.MaxNodes {
		return Function{}, resourceLimit("ports", "permutation exceeds node budget")
	}
	if len(orderedPorts) != len(function.Interface) {
		return Function{}, invalid("INVALID_PERMUTATION", "ports", "permutation must list every interface port exactly once")
	}
	binders, body, err := peelLeading(function.Expression, len(function.Interface))
	if err != nil {
		return Function{}, err
	}
	bindings, portTypes := make([]PortBinding, len(orderedPorts)), make([]Port, len(orderedPorts))
	seen := make(map[string]bool)
	for i, port := range orderedPorts {
		index := interfaceIndex(function.Interface, port)
		if index < 0 || seen[port] {
			return Function{}, invalid("INVALID_PERMUTATION", fmt.Sprintf("ports[%d]", i), "permutation contains an unknown or repeated port")
		}
		seen[port] = true
		bindings[i] = function.Interface[index]
		portTypes[i] = Port{Role: port, Type: *cloneTypeValue(*binders[index].Type)}
	}
	result := Function{Expression: wrapLambdas(bindings, portTypes, body), Interface: bindings}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

func checkExpression(context Document, expression *Term) (Type, error) {
	document := context
	document.Expression = expression
	document.Closed = false
	checked, err := Check(document)
	if err != nil {
		return Type{}, err
	}
	if checked.Status == status.OutcomeIncomplete {
		return Type{}, &Error{Status: status.OutcomeIncomplete, Code: "INCOMPLETE_OPERATOR_INPUT", Path: "expression", Message: "operator input contains unresolved holes"}
	}
	return checked.Type, nil
}

func peelLeading(expression *Term, count int) ([]*Term, *Term, error) {
	binders := make([]*Term, count)
	current := expression
	for i := 0; i < count; i++ {
		if current == nil || current.Kind != TermLambda {
			return nil, nil, invalid("FUNCTION_INTERFACE_ORDER", fmt.Sprintf("expression.lambda[%d]", i), "expected leading lambda")
		}
		binders[i] = current
		current = current.Body
	}
	return binders, current, nil
}

func wrapLambdas(bindings []PortBinding, ports []Port, body *Term) *Term {
	result := body
	for i := len(bindings) - 1; i >= 0; i-- {
		typ := cloneTypeValue(ports[i].Type)
		result = &Term{Kind: TermLambda, ID: bindings[i].Binder, Type: typ, Body: result}
	}
	return result
}

func interfaceIndex(bindings []PortBinding, port string) int {
	for i, binding := range bindings {
		if binding.Port == port {
			return i
		}
	}
	return -1
}

func cloneTypeValue(value Type) *Type {
	copyValue := value
	if value.From != nil {
		copyValue.From = cloneTypeValue(*value.From)
	}
	if value.To != nil {
		copyValue.To = cloneTypeValue(*value.To)
	}
	return &copyValue
}

func freshOperatorID(base string, reserved map[string]bool) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "op"
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if len(candidate) > 256 {
			candidate = fmt.Sprintf("operator_%d", i)
		}
		if !reserved[candidate] {
			reserved[candidate] = true
			return candidate
		}
	}
}

func termIDs(term *Term) []string {
	ids := []string{}
	active := map[*Term]bool{}
	var visit func(*Term)
	visit = func(value *Term) {
		if value == nil || active[value] {
			return
		}
		active[value] = true
		if value.Kind == TermLambda || value.Kind == TermExists || value.Kind == TermVar {
			ids = append(ids, value.ID)
		}
		for _, arg := range value.Args {
			visit(arg.Value)
		}
		for _, child := range value.Operands {
			visit(child)
		}
		visit(value.Body)
		visit(value.Function)
		visit(value.Argument)
		delete(active, value)
	}
	visit(term)
	sort.Strings(ids)
	return ids
}

func reservedMap(ids []string) map[string]bool {
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result
}

func remapInterface(bindings []PortBinding, renames map[string]string) []PortBinding {
	result := append([]PortBinding(nil), bindings...)
	for i := range result {
		if id, ok := renames[result[i].Binder]; ok {
			result[i].Binder = id
		}
	}
	return result
}

func applyJoinOutputs(bindings []PortBinding, ports []Port, sides []string, outputs []JoinOutput) ([]PortBinding, []Port, error) {
	counts := map[string]int{}
	for _, binding := range bindings {
		counts[binding.Port]++
	}
	needsMapping := false
	for _, count := range counts {
		if count > 1 {
			needsMapping = true
		}
	}
	if len(outputs) == 0 {
		if needsMapping {
			return nil, nil, invalid("JOIN_OUTPUT_MAPPING_REQUIRED", "outputs", "duplicate remaining port names require explicit output mapping")
		}
		return bindings, ports, nil
	}
	if len(outputs) != len(bindings) {
		return nil, nil, invalid("INVALID_JOIN_OUTPUTS", "outputs", "output mapping must cover every remaining port exactly once")
	}
	newBindings, newPorts := make([]PortBinding, len(outputs)), make([]Port, len(outputs))
	seen := map[string]bool{}
	usedInput := map[string]bool{}
	for i, output := range outputs {
		if output.Side != "left" && output.Side != "right" {
			return nil, nil, invalid("INVALID_JOIN_OUTPUT_SIDE", fmt.Sprintf("outputs[%d].side", i), "side must be left or right")
		}
		if !validText(output.As, 256) || seen[output.As] {
			return nil, nil, invalid("INVALID_JOIN_OUTPUT_ID", fmt.Sprintf("outputs[%d].as", i), "output names must be unique bounded text")
		}
		seen[output.As] = true
		source := output.Side + ":" + output.Port
		found := -1
		for j, binding := range bindings {
			sidePort := sides[j] + ":" + binding.Port
			if sidePort == source {
				found = j
				break
			}
		}
		if found < 0 || usedInput[source] {
			return nil, nil, invalid("INVALID_JOIN_OUTPUTS", fmt.Sprintf("outputs[%d]", i), "output references an unknown or repeated input port")
		}
		usedInput[source] = true
		newBindings[i] = PortBinding{Port: output.As, Binder: bindings[found].Binder}
		newPorts[i] = Port{Role: output.As, Type: ports[found].Type}
	}
	return newBindings, newPorts, nil
}

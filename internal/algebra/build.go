package algebra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/status"
)

const BuildSchemaVersion = 1

// BuildArgument connects an atom port to an earlier expression step.
type BuildArgument struct {
	Role  string `json:"role"`
	Value string `json:"value"`
}

// BuildStep is the flattened tagged request union. DecodeBuildRequest rejects
// fields that do not belong to the selected Op; Build validates the same
// constraints for programmatically created requests.
type BuildStep struct {
	ID         string          `json:"id"`
	Op         string          `json:"op"`
	Symbol     string          `json:"symbol,omitempty"`
	Quantifier string          `json:"quantifier,omitempty"`
	Count      *int            `json:"count,omitempty"`
	Value      *Term           `json:"value,omitempty"`
	Args       []BuildArgument `json:"args,omitempty"`
	Inputs     []string        `json:"inputs,omitempty"`
	Left       string          `json:"left,omitempty"`
	Right      string          `json:"right,omitempty"`
	Mapping    []JoinPair      `json:"mapping,omitempty"`
	Outputs    []JoinOutput    `json:"outputs,omitempty"`
	Function   string          `json:"function,omitempty"`
	Port       string          `json:"port,omitempty"`
	Argument   string          `json:"argument,omitempty"`
	Predicate  string          `json:"predicate,omitempty"`
	Ports      []string        `json:"ports,omitempty"`
}

// BuildRequest is a finite construction recipe. It contains no final AST,
// inline signatures, or executable callbacks.
type BuildRequest struct {
	SchemaVersion  int             `json:"schema_version"`
	RequestContext *RequestContext `json:"request_context,omitempty"`
	Parameters     []Declaration   `json:"parameters,omitempty"`
	References     []Declaration   `json:"references,omitempty"`
	Closed         bool            `json:"closed,omitempty"`
	Limits         Limits          `json:"limits,omitempty"`
	Steps          []BuildStep     `json:"steps"`
	Result         string          `json:"result"`
}

type BuildTrace struct {
	ID           string             `json:"id"`
	Op           string             `json:"op"`
	Symbol       string             `json:"symbol,omitempty"`
	Inputs       []string           `json:"inputs,omitempty"`
	Mapping      []JoinPair         `json:"mapping,omitempty"`
	Outputs      []JoinOutput       `json:"outputs,omitempty"`
	Port         string             `json:"port,omitempty"`
	Ports        []string           `json:"ports,omitempty"`
	ResultKind   string             `json:"result_kind"`
	Status       status.OutcomeKind `json:"status"`
	CheckedType  *Type              `json:"checked_type,omitempty"`
	Expression   *Term              `json:"expression,omitempty"`
	Interface    []PortBinding      `json:"interface,omitempty"`
	Substitution Substitution       `json:"substitution,omitempty"`
}

type BuildResult struct {
	Status            status.OutcomeKind `json:"status"`
	ResultID          string             `json:"result_id"`
	Expression        *Term              `json:"expression,omitempty"`
	Interface         []PortBinding      `json:"interface,omitempty"`
	Checked           *CheckedExpression `json:"checked,omitempty"`
	Substitution      Substitution       `json:"substitution,omitempty"`
	CanonicalDocument *Document          `json:"canonical_document,omitempty"`
	Trace             []BuildTrace       `json:"trace"`
}

type buildValue struct {
	expression   *Term
	function     *Function
	substitution Substitution
	checked      *CheckedExpression
	kind         string
}

// DecodeBuildRequest strictly decodes one finite construction recipe.
func DecodeBuildRequest(data []byte) (BuildRequest, error) {
	if err := PreflightJSON(data); err != nil {
		return BuildRequest{}, err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil || root == nil {
		return BuildRequest{}, invalid("INVALID_BUILD_REQUEST", "request", "request must be a JSON object")
	}
	if err := validateBuildRequestShape(root); err != nil {
		return BuildRequest{}, err
	}
	var request BuildRequest
	if err := strictTypedUnmarshal(data, &request, "request"); err != nil {
		return BuildRequest{}, err
	}
	if request.SchemaVersion != BuildSchemaVersion {
		return BuildRequest{}, invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported build request version")
	}
	if raw, ok := root["request_context"]; ok {
		contextValue, contextErr := DecodeRequestContextJSON(raw)
		if contextErr != nil {
			return BuildRequest{}, contextErr
		}
		request.RequestContext = &contextValue
	}
	var err error
	if raw, ok := root["parameters"]; ok {
		request.Parameters, err = decodeBuildDeclarations(raw, "parameters")
		if err != nil {
			return BuildRequest{}, err
		}
	}
	if raw, ok := root["references"]; ok {
		request.References, err = decodeBuildDeclarations(raw, "references")
		if err != nil {
			return BuildRequest{}, err
		}
	}
	for i, step := range request.Steps {
		if step.Op == "value" {
			stepRaw, _ := rawObjectIndex(root["steps"], i)
			valueRaw := stepRaw["value"]
			term, err := DecodeTermJSON(valueRaw, fmt.Sprintf("steps[%d].value", i))
			if err != nil {
				return BuildRequest{}, err
			}
			step.Value = term
			request.Steps[i] = step
		}
	}
	limits, err := NormalizeLimits(request.Limits)
	if err != nil {
		return BuildRequest{}, err
	}
	if len(request.Steps) > 64 {
		return BuildRequest{}, resourceLimit("steps", "build has more than 64 steps")
	}
	for i, step := range request.Steps {
		path := fmt.Sprintf("steps[%d]", i)
		if err := validateBuildStepFields(step, path); err != nil {
			return BuildRequest{}, err
		}
		if err := validateBuildStepBounds(step, limits, path); err != nil {
			return BuildRequest{}, err
		}
	}
	return request, nil
}

// Build executes an explicitly ordered recipe using the ordinary typed
// checker. It constructs representations only; it does not evaluate facts.
func Build(context Document, request BuildRequest) (BuildResult, error) {
	if request.SchemaVersion != BuildSchemaVersion {
		return BuildResult{}, invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported build request version")
	}
	limits, err := NormalizeLimits(request.Limits)
	if err != nil {
		return BuildResult{}, err
	}
	if len(request.Steps) == 0 {
		return BuildResult{}, invalid("EMPTY_BUILD", "steps", "at least one build step is required")
	}
	if len(request.Steps) > 64 {
		return BuildResult{}, resourceLimit("steps", "build has more than 64 steps")
	}
	if len(context.Parameters) != 0 || len(context.References) != 0 || context.RequestContext != nil {
		return BuildResult{}, invalid("INVALID_BUILD_CONTEXT", "context", "build context supplies metadata declarations only")
	}
	if !validText(request.Result, 256) {
		return BuildResult{}, invalid("INVALID_RESULT_ID", "result", "result must name a prior step")
	}
	var requestContext *RequestContext
	if request.RequestContext != nil {
		cloned, cloneErr := CloneRequestContext(*request.RequestContext)
		if cloneErr != nil {
			return BuildResult{}, cloneErr
		}
		requestContext = &cloned
	}
	rawDocument := Document{SchemaVersion: SchemaVersion, Sorts: context.Sorts, Signatures: context.Signatures, Parameters: request.Parameters, References: request.References, RequestContext: requestContext, Limits: limits}
	if err := CheckDeclarations(rawDocument); err != nil {
		return BuildResult{}, err
	}
	document := Document{SchemaVersion: SchemaVersion, Sorts: cloneSorts(context.Sorts), Signatures: cloneSignatures(context.Signatures), Parameters: cloneDeclarations(request.Parameters), References: cloneDeclarations(request.References), RequestContext: requestContext, Limits: limits}
	values := make(map[string]buildValue, len(request.Steps))
	trace := make([]BuildTrace, 0, len(request.Steps))
	seen := make(map[string]bool, len(request.Steps))
	retainedNodes := 0
	for i, step := range request.Steps {
		path := fmt.Sprintf("steps[%d]", i)
		if !validText(step.ID, 256) {
			return BuildResult{}, invalid("INVALID_STEP_ID", path+".id", "step ID must be non-empty bounded text")
		}
		if seen[step.ID] {
			return BuildResult{}, invalid("DUPLICATE_STEP_ID", path+".id", "step ID is duplicated")
		}
		seen[step.ID] = true
		if err := validateBuildStepFields(step, path); err != nil {
			return BuildResult{}, err
		}
		if err := validateBuildStepBounds(step, limits, path); err != nil {
			return BuildResult{}, err
		}
		refs := buildStepReferences(step)
		for _, ref := range refs {
			if !validText(ref.id, 256) {
				return BuildResult{}, invalid("INVALID_STEP_REFERENCE", path+"."+ref.field, "step reference must be non-empty bounded text")
			}
			if !seen[ref.id] || ref.id == step.ID {
				return BuildResult{}, invalid("FORWARD_STEP_REFERENCE", path+"."+ref.field, "step references must point to an earlier step")
			}
		}
		value, inputs, err := executeBuildStep(document, step, values, path)
		if err != nil {
			return BuildResult{}, prefixBuildError(err, path)
		}
		if value.expression != nil {
			checked, checkErr := checkBuildExpression(document, value.expression)
			if checkErr != nil {
				return BuildResult{}, prefixBuildError(checkErr, path)
			}
			value.checked = &checked
			if value.function != nil {
				if err := ValidateFunction(document, *value.function); err != nil {
					return BuildResult{}, prefixBuildError(err, path)
				}
			}
			count, err := countTermNodes(value.expression, limits)
			if err != nil {
				return BuildResult{}, prefixBuildError(err, path)
			}
			if value.kind == "substitution" {
				count = 0
			}
			retainedNodes += count
			// Keep both the step value and a detached trace copy within the total cap.
			if count > 0 {
				retainedNodes += count
			}
			if retainedNodes > 16384 {
				return BuildResult{}, resourceLimit(path, "aggregate retained result and trace nodes exceed 16384")
			}
		} else if value.kind == "substitution" {
			count, err := countSubstitutionNodes(value.substitution, limits)
			if err != nil {
				return BuildResult{}, prefixBuildError(err, path)
			}
			retainedNodes += count * 2
			if retainedNodes > 16384 {
				return BuildResult{}, resourceLimit(path, "aggregate retained result and trace nodes exceed 16384")
			}
		}
		values[step.ID] = value
		traceRecord := traceForStep(step, value, inputs, limits)
		trace = append(trace, traceRecord)
	}
	final, ok := values[request.Result]
	if !ok {
		return BuildResult{}, invalid("UNKNOWN_RESULT_STEP", "result", "result does not name a build step")
	}
	result := BuildResult{Status: status.OutcomeOK, ResultID: request.Result, Trace: trace}
	if final.kind == "substitution" {
		count, err := countSubstitutionNodes(final.substitution, limits)
		if err != nil {
			return BuildResult{}, err
		}
		retainedNodes += count
		if retainedNodes > 16384 {
			return BuildResult{}, resourceLimit("result", "aggregate retained result and trace nodes exceed 16384")
		}
		result.Substitution = cloneSubstitution(final.substitution, limits)
		return result, nil
	}
	if final.expression == nil {
		return BuildResult{}, invalid("INVALID_RESULT_KIND", "result", "final step does not produce an expression or substitution")
	}
	document.Expression = final.expression
	document.Closed = request.Closed
	checked, err := Check(document)
	if err != nil {
		return BuildResult{}, prefixBuildError(err, "result")
	}
	if final.function != nil {
		if err := ValidateFunction(document, *final.function); err != nil {
			return BuildResult{}, prefixBuildError(err, "result")
		}
		result.Interface = append([]PortBinding(nil), final.function.Interface...)
	}
	result.Status, result.Checked = checked.Status, &checked
	result.Expression, err = copyTermDetached(final.expression, limits)
	if err != nil {
		return BuildResult{}, err
	}
	finalNodes, err := countTermNodes(final.expression, limits)
	if err != nil {
		return BuildResult{}, err
	}
	retainedNodes += finalNodes // final result and canonical document share this detached expression
	if retainedNodes > 16384 {
		return BuildResult{}, resourceLimit("result", "aggregate retained result and trace nodes exceed 16384")
	}
	canonical := document
	canonical.Expression = result.Expression
	result.CanonicalDocument = &canonical
	return result, nil
}

func executeBuildStep(context Document, step BuildStep, values map[string]buildValue, path string) (buildValue, []string, error) {
	lookup := func(id, field string) (buildValue, error) {
		v, ok := values[id]
		if !ok {
			return buildValue{}, invalid("UNKNOWN_STEP_REFERENCE", path+"."+field, "referenced step is unavailable")
		}
		return v, nil
	}
	expression := func(id, field string) (*Term, error) {
		v, err := lookup(id, field)
		if err != nil {
			return nil, err
		}
		if v.expression == nil || v.kind == "substitution" {
			return nil, invalid("STEP_KIND_MISMATCH", path+"."+field, "step must produce an expression")
		}
		return v.expression, nil
	}
	function := func(id, field string) (Function, error) {
		v, err := lookup(id, field)
		if err != nil {
			return Function{}, err
		}
		if v.function == nil {
			return Function{}, invalid("STEP_KIND_MISMATCH", path+"."+field, "step must produce a relation function")
		}
		return *v.function, nil
	}
	switch step.Op {
	case "relation":
		f, err := Relation(context, step.Symbol)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: f.Expression, function: &f, kind: "function"}, nil, nil
	case "value":
		if step.Value == nil {
			return buildValue{}, nil, invalid("MISSING_VALUE", path+".value", "typed term value is required")
		}
		if step.Value.Kind != TermParam && step.Value.Kind != TermRef && step.Value.Kind != TermHole {
			return buildValue{}, nil, invalid("INVALID_VALUE_KIND", path+".value.kind", "value steps allow only Param, Ref, or Hole")
		}
		checked, err := checkBuildExpression(context, step.Value)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: step.Value, kind: "expression", checked: &checked}, nil, nil
	case "atom":
		args := make([]Argument, len(step.Args))
		inputs := make([]string, 0, len(step.Args))
		for i, arg := range step.Args {
			term, err := expression(arg.Value, fmt.Sprintf("args[%d].value", i))
			if err != nil {
				return buildValue{}, nil, err
			}
			args[i] = Argument{Role: arg.Role, Value: term}
			inputs = append(inputs, arg.Value)
		}
		return buildValue{expression: &Term{Kind: TermAtom, Symbol: step.Symbol, Args: args}, kind: "expression"}, inputs, nil
	case "and":
		operands := make([]*Term, len(step.Inputs))
		for i, id := range step.Inputs {
			term, err := expression(id, fmt.Sprintf("inputs[%d]", i))
			if err != nil {
				return buildValue{}, nil, err
			}
			operands[i] = term
		}
		return buildValue{expression: &Term{Kind: TermAnd, Operands: operands}, kind: "expression"}, append([]string(nil), step.Inputs...), nil
	case "not":
		if len(step.Inputs) != 1 {
			return buildValue{}, nil, invalid("INVALID_NOT_INPUTS", path+".inputs", "not requires exactly one input")
		}
		value, err := lookup(step.Inputs[0], "inputs[0]")
		if err != nil {
			return buildValue{}, nil, err
		}
		if value.expression == nil || value.kind == "substitution" {
			return buildValue{}, nil, invalid("STEP_KIND_MISMATCH", path+".inputs[0]", "not requires an expression")
		}
		if value.function == nil {
			result, err := Negate(context, value.expression)
			if err != nil {
				return buildValue{}, nil, err
			}
			return buildValue{expression: result, kind: "expression"}, []string{step.Inputs[0]}, nil
		}
		fn, err := NegateFunction(context, *value.function)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: fn.Expression, function: &fn, kind: "function"}, []string{step.Inputs[0]}, nil
	case "quantifier":
		if len(step.Inputs) < 1 || len(step.Inputs) > 2 {
			return buildValue{}, nil, invalid("INVALID_QUANTIFIER_INPUTS", path+".inputs", "quantifier requires a predicate and optional restrictor")
		}
		value, err := lookup(step.Inputs[0], "inputs[0]")
		if err != nil {
			return buildValue{}, nil, err
		}
		if value.function == nil || len(value.function.Interface) != 1 {
			return buildValue{}, nil, invalid("QUANTIFIER_REQUIRES_UNARY_PREDICATE", path+".inputs[0]", "quantifier consumes exactly one leading predicate lambda")
		}
		binders, body, err := peelLeading(value.expression, 1)
		if err != nil {
			return buildValue{}, nil, err
		}
		var restrictor *Term
		if len(step.Inputs) == 2 {
			rv, err := lookup(step.Inputs[1], "inputs[1]")
			if err != nil {
				return buildValue{}, nil, err
			}
			if rv.function == nil || len(rv.function.Interface) != 1 {
				return buildValue{}, nil, invalid("QUANTIFIER_RESTRICTOR_INTERFACE", path+".inputs[1]", "restrictor must have one leading lambda")
			}
			rb, rbBody, err := peelLeading(rv.expression, 1)
			if err != nil {
				return buildValue{}, nil, err
			}
			if !TypeEqual(*binders[0].Type, *rb[0].Type) {
				return buildValue{}, nil, invalid("QUANTIFIER_RESTRICTOR_TYPE", path+".inputs[1]", "restrictor binder type must match the quantified variable")
			}
			restrictor, err = SubstituteBinder(rbBody, rb[0].ID, &Term{Kind: TermVar, ID: binders[0].ID}, context.Limits)
			if err != nil {
				return buildValue{}, nil, err
			}
		}
		result, err := NewQuantifier(context, binders[0].ID, *binders[0].Type, step.Quantifier, step.Count, restrictor, body)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: result, kind: "expression"}, append([]string{}, step.Inputs...), nil
	case "ask":
		if len(step.Inputs) < 1 || len(step.Inputs) > 2 {
			return buildValue{}, nil, invalid("INVALID_ASK_INPUTS", path+".inputs", "ask requires a proposition and optional restrictor")
		}
		value, err := lookup(step.Inputs[0], "inputs[0]")
		if err != nil {
			return buildValue{}, nil, err
		}
		if value.expression == nil || value.kind == "substitution" {
			return buildValue{}, nil, invalid("STEP_KIND_MISMATCH", path+".inputs[0]", "ask requires an expression")
		}
		var answers []Declaration
		var body *Term
		if value.function != nil {
			binders, peeled, err := peelLeading(value.expression, len(value.function.Interface))
			if err != nil {
				return buildValue{}, nil, err
			}
			answers = make([]Declaration, len(binders))
			for i, binder := range binders {
				answers[i] = Declaration{ID: binder.ID, Type: *cloneTypeValue(*binder.Type)}
			}
			body = peeled
		} else {
			body = value.expression
		}
		var restrictor *Term
		if len(step.Inputs) == 2 {
			rv, err := lookup(step.Inputs[1], "inputs[1]")
			if err != nil {
				return buildValue{}, nil, err
			}
			if len(answers) == 0 {
				if rv.function != nil || rv.expression == nil {
					return buildValue{}, nil, invalid("ASK_RESTRICTOR_INTERFACE", path+".inputs[1]", "yes/no Ask restrictor must be a proposition")
				}
				restrictor = rv.expression
			} else {
				if rv.function == nil || len(rv.function.Interface) != len(answers) {
					return buildValue{}, nil, invalid("ASK_RESTRICTOR_INTERFACE", path+".inputs[1]", "Ask restrictor must match the ordered answer interface")
				}
				rbinders, rbody, err := peelLeading(rv.expression, len(rv.function.Interface))
				if err != nil {
					return buildValue{}, nil, err
				}
				for i, binder := range rbinders {
					if !TypeEqual(answers[i].Type, *binder.Type) {
						return buildValue{}, nil, invalid("ASK_RESTRICTOR_TYPE", fmt.Sprintf("%s.inputs[1].interface[%d]", path, i), "restrictor binder type must match the ordered answer")
					}
					rbody, err = SubstituteBinder(rbody, binder.ID, &Term{Kind: TermVar, ID: answers[i].ID}, context.Limits)
					if err != nil {
						return buildValue{}, nil, err
					}
				}
				restrictor = rbody
			}
		}
		result, err := NewAsk(context, answers, restrictor, body)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: result, kind: "expression"}, append([]string{}, step.Inputs...), nil
	case "join":
		left, err := function(step.Left, "left")
		if err != nil {
			return buildValue{}, nil, err
		}
		right, err := function(step.Right, "right")
		if err != nil {
			return buildValue{}, nil, err
		}
		f, err := Join(context, left, right, step.Mapping, step.Outputs)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: f.Expression, function: &f, kind: "function"}, []string{step.Left, step.Right}, nil
	case "bind":
		f, err := function(step.Function, "function")
		if err != nil {
			return buildValue{}, nil, err
		}
		arg, err := expression(step.Argument, "argument")
		if err != nil {
			return buildValue{}, nil, err
		}
		out, err := Bind(context, f, step.Port, arg)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: out.Expression, function: &out, kind: "function"}, []string{step.Function, step.Argument}, nil
	case "apply":
		f, err := function(step.Function, "function")
		if err != nil {
			return buildValue{}, nil, err
		}
		arg, err := expression(step.Argument, "argument")
		if err != nil {
			return buildValue{}, nil, err
		}
		out, err := Apply(context, f.Expression, arg)
		if err != nil {
			return buildValue{}, nil, err
		}
		remaining := append([]PortBinding(nil), f.Interface...)
		if len(remaining) == 0 {
			return buildValue{}, nil, invalid("FUNCTION_INTERFACE_ORDER", "function", "function interface has no leading port")
		}
		remaining = remaining[1:]
		value := buildValue{expression: out, kind: "expression"}
		if len(remaining) > 0 {
			ports := make([]Port, 0, len(remaining))
			originalBinders, _, peelErr := peelLeading(f.Expression, len(f.Interface))
			if peelErr != nil {
				return buildValue{}, nil, peelErr
			}
			for i, b := range f.Interface {
				if i > 0 {
					ports = append(ports, Port{Role: b.Port, Type: *cloneTypeValue(*originalBinders[i].Type)})
				}
			}
			fn := Function{Expression: out, Interface: remaining}
			_ = ports
			if err := ValidateFunction(context, fn); err != nil {
				return buildValue{}, nil, err
			}
			value.function = &fn
			value.kind = "function"
		}
		return value, []string{step.Function, step.Argument}, nil
	case "restrict":
		f, err := function(step.Function, "function")
		if err != nil {
			return buildValue{}, nil, err
		}
		p, err := function(step.Predicate, "predicate")
		if err != nil {
			return buildValue{}, nil, err
		}
		out, err := Restrict(context, f, step.Port, p)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: out.Expression, function: &out, kind: "function"}, []string{step.Function, step.Predicate}, nil
	case "project":
		f, err := function(step.Function, "function")
		if err != nil {
			return buildValue{}, nil, err
		}
		out, err := Project(context, f, step.Ports)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: out.Expression, function: &out, kind: "function"}, []string{step.Function}, nil
	case "permute":
		f, err := function(step.Function, "function")
		if err != nil {
			return buildValue{}, nil, err
		}
		out, err := Permute(context, f, step.Ports)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{expression: out.Expression, function: &out, kind: "function"}, []string{step.Function}, nil
	case "unify":
		left, err := expression(step.Left, "left")
		if err != nil {
			return buildValue{}, nil, err
		}
		right, err := expression(step.Right, "right")
		if err != nil {
			return buildValue{}, nil, err
		}
		substitution, err := Unify(context, left, right)
		if err != nil {
			return buildValue{}, nil, err
		}
		checked, err := checkBuildExpression(context, left)
		if err != nil {
			return buildValue{}, nil, err
		}
		return buildValue{kind: "substitution", substitution: substitution, checked: &checked}, []string{step.Left, step.Right}, nil
	default:
		return buildValue{}, nil, &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_BUILD_OPERATION", Path: path + ".op", Message: "build operation is not supported"}
	}
}

func validateBuildRequestShape(root map[string]json.RawMessage) error {
	allowed := map[string]bool{"schema_version": true, "request_context": true, "parameters": true, "references": true, "closed": true, "limits": true, "steps": true, "result": true}
	if err := checkFields(root, allowed, "request"); err != nil {
		return err
	}
	for _, key := range []string{"schema_version", "steps", "result"} {
		if _, ok := root[key]; !ok {
			return invalid("MISSING_FIELD", "request."+key, "field is required")
		}
	}
	for key, raw := range root {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return invalid("NULL_FIELD", "request."+key, "null is not a valid field value")
		}
	}
	var steps []json.RawMessage
	if err := json.Unmarshal(root["steps"], &steps); err != nil || steps == nil {
		return invalid("INVALID_STEPS", "steps", "steps must be an array")
	}
	for i, raw := range steps {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil || object == nil {
			return invalid("INVALID_BUILD_STEP", fmt.Sprintf("steps[%d]", i), "step must be an object")
		}
		if err := validateBuildStepShape(object, fmt.Sprintf("steps[%d]", i)); err != nil {
			return err
		}
	}
	for _, key := range []string{"parameters", "references"} {
		if raw, ok := root[key]; ok {
			var entries []json.RawMessage
			if err := json.Unmarshal(raw, &entries); err != nil || entries == nil {
				return invalid("INVALID_DECLARATIONS", key, "field must be an array")
			}
		}
	}
	if raw, ok := root["limits"]; ok {
		var limitsObject map[string]json.RawMessage
		if err := json.Unmarshal(raw, &limitsObject); err != nil || limitsObject == nil {
			return invalid("INVALID_LIMITS", "limits", "limits must be an object")
		}
		allowedLimits := map[string]bool{"max_nodes": true, "max_depth": true, "max_type_depth": true, "max_symbols": true, "max_ports": true}
		if err := checkFields(limitsObject, allowedLimits, "limits"); err != nil {
			return err
		}
		for key, value := range limitsObject {
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return invalid("INVALID_LIMIT", "limits."+key, "limit cannot be null; omit it to use the default")
			}
		}
	}
	return nil
}

func validateBuildStepShape(object map[string]json.RawMessage, path string) error {
	if _, ok := object["id"]; !ok {
		return invalid("MISSING_FIELD", path+".id", "step id is required")
	}
	if _, ok := object["op"]; !ok {
		return invalid("MISSING_FIELD", path+".op", "step op is required")
	}
	var op string
	if err := json.Unmarshal(object["op"], &op); err != nil || !validText(op, 32) {
		return invalid("INVALID_BUILD_STEP", path+".op", "op must be non-empty bounded text")
	}
	fields := map[string][]string{
		"relation": {"id", "op", "symbol"}, "value": {"id", "op", "value"}, "atom": {"id", "op", "symbol", "args"}, "and": {"id", "op", "inputs"},
		"not": {"id", "op", "inputs"}, "quantifier": {"id", "op", "inputs", "quantifier", "count"}, "ask": {"id", "op", "inputs"},
		"join": {"id", "op", "left", "right", "mapping", "outputs"}, "bind": {"id", "op", "function", "port", "argument"}, "apply": {"id", "op", "function", "argument"},
		"restrict": {"id", "op", "function", "port", "predicate"}, "project": {"id", "op", "function", "ports"}, "permute": {"id", "op", "function", "ports"}, "unify": {"id", "op", "left", "right"},
	}
	allowed, known := fields[op]
	if !known {
		return &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_BUILD_OPERATION", Path: path + ".op", Message: "build operation is not supported"}
	} // Build emits typed unsupported for unknown operations.
	allowedSet := map[string]bool{}
	for _, key := range allowed {
		allowedSet[key] = true
	}
	if err := checkFields(object, allowedSet, path); err != nil {
		return err
	}
	optional := map[string]bool{"mapping": true, "outputs": true, "count": true}
	for _, key := range allowed {
		if optional[key] {
			continue
		}
		if _, ok := object[key]; !ok {
			return invalid("MISSING_FIELD", path+"."+key, "field is required")
		}
		if bytes.Equal(bytes.TrimSpace(object[key]), []byte("null")) {
			return invalid("NULL_FIELD", path+"."+key, "null is not a valid field value")
		}
	}
	for _, key := range []string{"mapping", "outputs"} {
		if raw, ok := object[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return invalid("NULL_FIELD", path+"."+key, "null is not a valid field value")
		}
	}
	for _, key := range []string{"args", "inputs", "ports"} {
		if raw, ok := object[key]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(raw, &list); err != nil || list == nil {
				return invalid("INVALID_ARRAY", path+"."+key, "field must be an array")
			}
		}
	}
	if op == "value" {
		term, err := DecodeTermJSON(object["value"], path+".value")
		if err != nil {
			return err
		}
		if term.Kind != TermParam && term.Kind != TermRef && term.Kind != TermHole {
			return invalid("INVALID_VALUE_KIND", path+".value.kind", "value permits only Param, Ref, or Hole")
		}
	}
	if op == "quantifier" {
		if raw, exists := object["count"]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return invalid("NULL_FIELD", path+".count", "count cannot be null")
		}
	}
	return nil
}

func checkFields(object map[string]json.RawMessage, allowed map[string]bool, path string) error {
	for key := range object {
		if !allowed[key] {
			return invalid("UNKNOWN_FIELD", path+"."+key, "field is not allowed here")
		}
	}
	return nil
}

func checkBuildExpression(context Document, expression *Term) (CheckedExpression, error) {
	document := context
	document.Expression = expression
	document.Closed = false
	return Check(document)
}

func decodeBuildDeclarations(raw json.RawMessage, path string) ([]Declaration, error) {
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil || entries == nil {
		return nil, invalid("INVALID_DECLARATIONS", path, "declarations must be an array")
	}
	result := make([]Declaration, len(entries))
	for i, entry := range entries {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(entry, &object); err != nil || object == nil {
			return nil, invalid("INVALID_DECLARATION", fmt.Sprintf("%s[%d]", path, i), "declaration must be an object")
		}
		if err := checkFields(object, map[string]bool{"id": true, "type": true}, fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return nil, err
		}
		if _, ok := object["id"]; !ok {
			return nil, invalid("MISSING_FIELD", fmt.Sprintf("%s[%d].id", path, i), "id is required")
		}
		typRaw, ok := object["type"]
		if !ok {
			return nil, invalid("MISSING_FIELD", fmt.Sprintf("%s[%d].type", path, i), "type is required")
		}
		var id string
		if err := json.Unmarshal(object["id"], &id); err != nil {
			return nil, invalid("INVALID_DECLARATION_ID", fmt.Sprintf("%s[%d].id", path, i), "id must be a string")
		}
		typ, err := DecodeTypeJSON(typRaw, fmt.Sprintf("%s[%d].type", path, i))
		if err != nil {
			return nil, err
		}
		result[i] = Declaration{ID: id, Type: typ}
	}
	return result, nil
}

func rawObjectIndex(raw json.RawMessage, index int) (map[string]json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || index < 0 || index >= len(values) {
		return nil, fmt.Errorf("index out of range")
	}
	var object map[string]json.RawMessage
	err := json.Unmarshal(values[index], &object)
	return object, err
}

type stepReference struct{ id, field string }

func buildStepReferences(s BuildStep) []stepReference {
	switch s.Op {
	case "atom":
		r := make([]stepReference, len(s.Args))
		for i, a := range s.Args {
			r[i] = stepReference{a.Value, fmt.Sprintf("args[%d].value", i)}
		}
		return r
	case "and":
		r := make([]stepReference, len(s.Inputs))
		for i, id := range s.Inputs {
			r[i] = stepReference{id, fmt.Sprintf("inputs[%d]", i)}
		}
		return r
	case "not", "quantifier", "ask":
		r := make([]stepReference, len(s.Inputs))
		for i, id := range s.Inputs {
			r[i] = stepReference{id, fmt.Sprintf("inputs[%d]", i)}
		}
		return r
	case "join", "unify":
		return []stepReference{{s.Left, "left"}, {s.Right, "right"}}
	case "bind", "apply":
		return []stepReference{{s.Function, "function"}, {s.Argument, "argument"}}
	case "restrict":
		return []stepReference{{s.Function, "function"}, {s.Predicate, "predicate"}}
	case "project", "permute":
		return []stepReference{{s.Function, "function"}}
	}
	return nil
}

func validateBuildStepFields(step BuildStep, path string) error {
	if !validText(step.Op, 32) {
		return invalid("INVALID_OPERATION", path+".op", "operation must be non-empty bounded text")
	}
	allowed := map[string]map[string]bool{
		"relation": {"symbol": true}, "value": {"value": true}, "atom": {"symbol": true, "args": true}, "and": {"inputs": true},
		"not": {"inputs": true}, "quantifier": {"inputs": true, "quantifier": true, "count": true}, "ask": {"inputs": true},
		"join": {"left": true, "right": true, "mapping": true, "outputs": true}, "bind": {"function": true, "port": true, "argument": true},
		"apply": {"function": true, "argument": true}, "restrict": {"function": true, "port": true, "predicate": true},
		"project": {"function": true, "ports": true}, "permute": {"function": true, "ports": true}, "unify": {"left": true, "right": true},
	}
	fields, known := allowed[step.Op]
	if !known {
		return &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_BUILD_OPERATION", Path: path + ".op", Message: "build operation is not supported"}
	}
	if step.Symbol != "" && !fields["symbol"] || step.Quantifier != "" && !fields["quantifier"] || step.Count != nil && !fields["count"] || step.Value != nil && !fields["value"] || step.Args != nil && !fields["args"] || step.Inputs != nil && !fields["inputs"] || step.Left != "" && !fields["left"] || step.Right != "" && !fields["right"] || step.Mapping != nil && !fields["mapping"] || step.Outputs != nil && !fields["outputs"] || step.Function != "" && !fields["function"] || step.Port != "" && !fields["port"] || step.Argument != "" && !fields["argument"] || step.Predicate != "" && !fields["predicate"] || step.Ports != nil && !fields["ports"] {
		return invalid("IRRELEVANT_FIELD", path, "step carries fields unrelated to its operation")
	}
	checkString := func(value, field string) error {
		if !validText(value, 256) {
			return invalid("INVALID_FIELD", path+"."+field, "field must be non-empty bounded text")
		}
		return nil
	}
	switch step.Op {
	case "relation":
		return checkString(step.Symbol, "symbol")
	case "value":
		if step.Value == nil {
			return invalid("MISSING_VALUE", path+".value", "typed term value is required")
		}
		if step.Value.Kind != TermParam && step.Value.Kind != TermRef && step.Value.Kind != TermHole {
			return invalid("INVALID_VALUE_KIND", path+".value.kind", "value permits only Param, Ref, or Hole")
		}
	case "atom":
		if err := checkString(step.Symbol, "symbol"); err != nil {
			return err
		}
		if step.Args == nil {
			return invalid("MISSING_FIELD", path+".args", "args array is required")
		}
	case "and":
		if step.Inputs == nil {
			return invalid("MISSING_FIELD", path+".inputs", "inputs array is required")
		}
	case "not":
		if step.Inputs == nil || len(step.Inputs) != 1 {
			return invalid("INVALID_NOT_INPUTS", path+".inputs", "not requires exactly one input")
		}
	case "quantifier":
		if step.Inputs == nil || len(step.Inputs) < 1 || len(step.Inputs) > 2 {
			return invalid("INVALID_QUANTIFIER_INPUTS", path+".inputs", "quantifier requires a predicate and optional restrictor")
		}
		switch step.Quantifier {
		case "some", "all":
			if step.Count != nil {
				return invalid("INVALID_QUANTIFIER_COUNT", path+".count", "some and all do not accept count")
			}
		case "exactly", "at_least", "at_most":
			if step.Count == nil || *step.Count < 0 {
				return invalid("INVALID_QUANTIFIER_COUNT", path+".count", "numeric quantifier requires a non-negative count")
			}
		default:
			return invalid("INVALID_QUANTIFIER_KIND", path+".quantifier", "quantifier kind is unsupported")
		}
	case "ask":
		if step.Inputs == nil || len(step.Inputs) < 1 || len(step.Inputs) > 2 {
			return invalid("INVALID_ASK_INPUTS", path+".inputs", "ask requires a proposition and optional restrictor")
		}
	case "join":
		if err := checkString(step.Left, "left"); err != nil {
			return err
		}
		return checkString(step.Right, "right")
	case "bind", "apply":
		if err := checkString(step.Function, "function"); err != nil {
			return err
		}
		if step.Op == "bind" {
			if err := checkString(step.Port, "port"); err != nil {
				return err
			}
		}
		return checkString(step.Argument, "argument")
	case "restrict":
		if err := checkString(step.Function, "function"); err != nil {
			return err
		}
		if err := checkString(step.Port, "port"); err != nil {
			return err
		}
		return checkString(step.Predicate, "predicate")
	case "project", "permute":
		if err := checkString(step.Function, "function"); err != nil {
			return err
		}
		if step.Ports == nil {
			return invalid("MISSING_FIELD", path+".ports", "ports array is required")
		}
	case "unify":
		if err := checkString(step.Left, "left"); err != nil {
			return err
		}
		return checkString(step.Right, "right")
	}
	return nil
}

func validateBuildStepBounds(step BuildStep, limits Limits, path string) error {
	if len(step.Args) > limits.MaxPorts || len(step.Inputs) > limits.MaxNodes || len(step.Mapping) > limits.MaxNodes || len(step.Outputs) > limits.MaxNodes || len(step.Ports) > limits.MaxNodes {
		return resourceLimit(path, "step array exceeds configured bounds")
	}
	for i, arg := range step.Args {
		if !validText(arg.Role, 256) {
			return invalid("INVALID_ROLE_ID", fmt.Sprintf("%s.args[%d].role", path, i), "argument role must be non-empty bounded text")
		}
		if !validText(arg.Value, 256) {
			return invalid("INVALID_STEP_REFERENCE", fmt.Sprintf("%s.args[%d].value", path, i), "argument must name an earlier step")
		}
	}
	for i, pair := range step.Mapping {
		if !validText(pair.LeftPort, 256) || !validText(pair.RightPort, 256) {
			return invalid("INVALID_JOIN_MAPPING", fmt.Sprintf("%s.mapping[%d]", path, i), "both mapped port IDs are required")
		}
	}
	for i, out := range step.Outputs {
		if !validText(out.Side, 8) || !validText(out.Port, 256) || !validText(out.As, 256) {
			return invalid("INVALID_JOIN_OUTPUT", fmt.Sprintf("%s.outputs[%d]", path, i), "side, source port, and output ID are required")
		}
	}
	for i, port := range step.Ports {
		if !validText(port, 256) {
			return invalid("INVALID_PORT_ID", fmt.Sprintf("%s.ports[%d]", path, i), "port ID must be non-empty bounded text")
		}
	}
	return nil
}

func prefixBuildError(err error, path string) error {
	if typed, ok := err.(*Error); ok {
		copyValue := *typed
		if copyValue.Path == "" {
			copyValue.Path = path
		} else if copyValue.Path == path || len(copyValue.Path) > len(path) && copyValue.Path[:len(path)+1] == path+"." {
			return &copyValue
		} else {
			copyValue.Path = path + "." + copyValue.Path
		}
		return &copyValue
	}
	return err
}

func traceForStep(step BuildStep, value buildValue, inputs []string, limits Limits) BuildTrace {
	record := BuildTrace{ID: step.ID, Op: step.Op, Symbol: step.Symbol, Inputs: append([]string(nil), inputs...), Mapping: append([]JoinPair(nil), step.Mapping...), Outputs: append([]JoinOutput(nil), step.Outputs...), Port: step.Port, Ports: append([]string(nil), step.Ports...), ResultKind: value.kind}
	if value.checked != nil {
		record.Status = value.checked.Status
		typ := cloneTypeValue(value.checked.Type)
		record.CheckedType = typ
	} else {
		record.Status = status.OutcomeOK
	}
	if value.expression != nil {
		record.Expression, _ = copyTermDetached(value.expression, limits)
	}
	if value.function != nil {
		record.Interface = append([]PortBinding(nil), value.function.Interface...)
	}
	if value.substitution != nil {
		record.Substitution = cloneSubstitution(value.substitution, limits)
	}
	return record
}

func cloneSubstitution(value Substitution, limits Limits) Substitution {
	if value == nil {
		return nil
	}
	result := make(Substitution, len(value))
	for id, term := range value {
		copyValue, err := copyTermDetached(term, limits)
		if err == nil {
			result[id] = copyValue
		}
	}
	return result
}

func countTermNodes(term *Term, limits Limits) (int, error) {
	if term == nil {
		return 0, nil
	}
	_, err := inspectTerm(term, limits)
	if err != nil {
		return 0, err
	}
	var count func(*Term, map[*Term]bool) int
	count = func(t *Term, active map[*Term]bool) int {
		if t == nil || active[t] {
			return 0
		}
		active[t] = true
		n := 1
		for _, a := range t.Args {
			n += count(a.Value, active)
		}
		for _, o := range t.Operands {
			n += count(o, active)
		}
		n += count(t.Body, active) + count(t.Function, active) + count(t.Argument, active)
		delete(active, t)
		return n
	}
	return count(term, map[*Term]bool{}), nil
}

func countSubstitutionNodes(value Substitution, limits Limits) (int, error) {
	total := 0
	keys := make([]string, 0, len(value))
	for id := range value {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		n, err := countTermNodes(value[id], limits)
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func cloneDeclarations(values []Declaration) []Declaration {
	result := make([]Declaration, len(values))
	for i, value := range values {
		result[i] = Declaration{ID: value.ID, Type: *cloneTypeValue(value.Type)}
	}
	return result
}
func cloneSorts(values []SortDeclaration) []SortDeclaration {
	return append([]SortDeclaration(nil), values...)
}
func cloneSignatures(values []Signature) []Signature {
	result := make([]Signature, len(values))
	for i, value := range values {
		result[i] = Signature{ID: value.ID, Label: value.Label, Ports: make([]Port, len(value.Ports))}
		for j, port := range value.Ports {
			result[i].Ports[j] = Port{Role: port.Role, Type: *cloneTypeValue(port.Type)}
		}
	}
	return result
}

func rejectDeclarationOverlap(parameters, references []Declaration) error {
	seen := map[string]bool{}
	for _, item := range parameters {
		seen[item.ID] = true
	}
	for _, item := range references {
		if seen[item.ID] {
			return invalid("DUPLICATE_DECLARATION_ID", "references", "parameter and reference IDs must be distinct")
		}
	}
	return nil
}

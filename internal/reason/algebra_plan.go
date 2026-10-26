package reason

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	maxAlgebraPlanAtoms     = 64
	maxAlgebraPlanVariables = 128
	betaQueryProfile        = "dse-positive-query-beta-v1"
)

// AlgebraPlanNodeKind is the checked positive query tree vocabulary.
type AlgebraPlanNodeKind string

const (
	AlgebraPlanAtom   AlgebraPlanNodeKind = "atom"
	AlgebraPlanAnd    AlgebraPlanNodeKind = "and"
	AlgebraPlanExists AlgebraPlanNodeKind = "exists"
)

// AlgebraPlanArgument retains a signature role and one namespace-aware value.
type AlgebraPlanArgument struct {
	Role  string             `json:"role"`
	Value AlgebraVariableKey `json:"value"`
	RefID string             `json:"ref_id,omitempty"`
}

// AlgebraPlanNode stores one source occurrence. Child IDs refer only to this
// plan's immutable tree and preserve source operand order.
type AlgebraPlanNode struct {
	ID       string                `json:"id"`
	Kind     AlgebraPlanNodeKind   `json:"kind"`
	Path     string                `json:"path"`
	Symbol   string                `json:"symbol,omitempty"`
	Args     []AlgebraPlanArgument `json:"args,omitempty"`
	Children []string              `json:"children,omitempty"`
	Binder   *AlgebraVariableKey   `json:"binder,omitempty"`
	Type     *algebra.Type         `json:"type,omitempty"`
}

// AlgebraPlanVariable records a typed binder or free parameter. UsedInAtom is
// descriptive only; false variables require a later explicit known-reference
// domain scan and are never assumed to have witnesses.
type AlgebraPlanVariable struct {
	Key        AlgebraVariableKey `json:"key"`
	Type       algebra.Type       `json:"type"`
	SourcePath string             `json:"source_path"`
	UsedInAtom bool               `json:"used_in_atom"`
}

// AlgebraPlanDescription is a detached, bounded, non-executable view.
type AlgebraPlanDescription struct {
	SchemaVersion       int                        `json:"schema_version"`
	Profile             string                     `json:"profile"`
	ContextID           string                     `json:"context_id"`
	At                  string                     `json:"at,omitempty"`
	FactsSnapshotID     string                     `json:"facts_snapshot_id"`
	FormalSnapshotID    string                     `json:"formal_snapshot_id"`
	Fingerprint         string                     `json:"fingerprint"`
	SourceFingerprint   string                     `json:"source_fingerprint"`
	LoweringProfile     string                     `json:"lowering_profile"`
	LoweringCertificate algebra.RewriteCertificate `json:"lowering_certificate"`
	Answers             []AlgebraAnswerPort        `json:"answers"`
	References          []algebra.Declaration      `json:"references"`
	Variables           []AlgebraPlanVariable      `json:"variables"`
	Nodes               []AlgebraPlanNode          `json:"nodes"`
	Root                string                     `json:"root"`
	Limits              AlgebraEvaluationLimits    `json:"limits"`
	Usage               AlgebraEvaluationUsage     `json:"usage"`
}

// AlgebraQueryPlan is an immutable compiled query. Its fields are private so
// callers cannot replace checked syntax, snapshots, or answer projections.
type AlgebraQueryPlan struct {
	profile           string
	contextID         string
	at                string
	factsSnapshotID   string
	formalSnapshotID  string
	fingerprint       string
	sourceFingerprint string
	loweringProfile   string
	lowering          algebra.RewriteCertificate
	answers           []AlgebraAnswerPort
	references        []algebra.Declaration
	variables         []AlgebraPlanVariable
	nodes             []AlgebraPlanNode
	root              string
	limits            AlgebraEvaluationLimits
	usage             AlgebraEvaluationUsage
	deadline          time.Time
	ctx               context.Context
	cancel            context.CancelFunc
}

// CompileAlgebraQuery checks, beta-lowers, and compiles one explicit positive
// query. It creates a plan only; it does not inspect facts or evaluate truth.
func CompileAlgebraQuery(parent context.Context, schema facts.NarySchema, environment algebra.RewriteEnvironment, source algebra.Document, request AlgebraQueryRequest, requested AlgebraEvaluationLimits) (*AlgebraQueryPlan, error) {
	if parent == nil {
		return nil, planInvalid("INVALID_CONTEXT", "context", "a caller context is required")
	}
	limits, err := NormalizeAlgebraEvaluationLimits(requested)
	if err != nil {
		return nil, err
	}
	limits.TermLimits, err = effectivePlanTermLimits(limits.TermLimits, schema.Document.Limits, environment.Document.Limits, source.Limits)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(time.Duration(limits.TimeoutMillis) * time.Millisecond)
	ctx, cancel := context.WithDeadline(parent, deadline)
	fail := func(err error) (*AlgebraQueryPlan, error) { cancel(); return nil, err }
	if err := contextPlanCheck(ctx); err != nil {
		return fail(err)
	}
	state := &planCompiler{ctx: ctx, limits: limits}
	if err := state.charge(len(schema.Document.Sorts)+len(schema.Document.Signatures)+len(environment.Document.Sorts)+len(environment.Document.Signatures)+len(source.Parameters)+len(source.References), "declarations"); err != nil {
		return fail(err)
	}
	if err := validateAlgebraPlanInputs(schema, environment, source, request, limits); err != nil {
		return fail(err)
	}

	// The selected fact schema has its own resource snapshot. The rewrite
	// environment separately pins the formal metadata snapshot.
	if err := compareSchemaWithFormal(schema.Document, environment); err != nil {
		return fail(err)
	}
	if err := compareSourceInventory(source, schema.Document); err != nil {
		return fail(err)
	}
	if err := compareDeclarations(source.Parameters, environment.Document.Parameters, "parameters"); err != nil {
		return fail(err)
	}
	if err := compareDeclarations(source.References, environment.Document.References, "references"); err != nil {
		return fail(err)
	}
	if err := state.checkContext(); err != nil {
		return fail(err)
	}

	// First check the self-contained source, then check it against the actual
	// selected declarations. Neither check treats labels as semantic selectors.
	checkedSource, err := algebra.Check(source)
	if err != nil {
		return fail(err)
	}
	if checkedSource.Status != status.OutcomeOK || len(checkedSource.Holes) != 0 {
		return fail(planIncomplete("QUERY_SOURCE_INCOMPLETE", "expression", "query source contains an unresolved hole"))
	}
	if err := state.charge(checkedSource.NodeCount, "source_check"); err != nil {
		return fail(err)
	}
	actual := source
	actual.Sorts = append([]algebra.SortDeclaration{}, schema.Document.Sorts...)
	actual.Signatures = clonePlanSignatures(schema.Document.Signatures)
	actual.Limits = limits.TermLimits
	checkedActual, err := algebra.Check(actual)
	if err != nil {
		return fail(err)
	}
	if checkedActual.Status != status.OutcomeOK || len(checkedActual.Holes) != 0 {
		return fail(planIncomplete("QUERY_SOURCE_INCOMPLETE", "expression", "query source is incomplete under the selected primitive declarations"))
	}
	if err := state.charge(checkedActual.NodeCount, "source_check"); err != nil {
		return fail(err)
	}

	if err := state.checkContext(); err != nil {
		return fail(err)
	}
	lowered, lowerCert, lowerWork, err := lowerQuery(source, environment, limits.TermLimits, state.remaining())
	if err != nil {
		return fail(err)
	}
	if err := state.charge(lowerWork, "beta_lowering"); err != nil {
		return fail(err)
	}
	if err := state.checkContext(); err != nil {
		return fail(err)
	}
	checkedLowered, err := algebra.Check(withSelectedInventory(source, schema.Document, lowered, limits.TermLimits))
	if err != nil {
		return fail(err)
	}
	if checkedLowered.Status != status.OutcomeOK || len(checkedLowered.Holes) != 0 {
		return fail(planIncomplete("QUERY_LOWERING_INCOMPLETE", "expression", "beta lowering did not produce a complete expression"))
	}
	if err := state.charge(checkedLowered.NodeCount, "lowered_check"); err != nil {
		return fail(err)
	}

	answers, leading, err := compileSelections(request.Select, lowered, checkedLowered)
	if err != nil {
		return fail(err)
	}
	body := lowered
	for range leading {
		if body == nil || body.Kind != algebra.TermLambda {
			return fail(planInvalid("LEADING_LAMBDA_MISMATCH", "expression", "selected leading lambda structure is inconsistent"))
		}
		body = body.Body
	}
	compileLimits := limits
	compileLimits.MaxWorkUnits = state.remaining()
	if compileLimits.MaxWorkUnits <= 0 {
		return fail(planResource("expression", "no logical work remains for plan compilation"))
	}
	compiler := &planCompiler{ctx: ctx, limits: compileLimits, parameterTypes: map[string]algebra.Type{}, referenceTypes: map[string]algebra.Type{}}
	for _, declaration := range source.Parameters {
		compiler.parameterTypes[declaration.ID] = typeCopy(declaration.Type)
	}
	for _, declaration := range source.References {
		compiler.referenceTypes[declaration.ID] = typeCopy(declaration.Type)
	}
	if err := compiler.pushSelectedLambdas(leading); err != nil {
		return fail(err)
	}
	root, err := compiler.compileTree(body, "expression.body", 1)
	if err != nil {
		return fail(err)
	}
	if compiler.atoms > limits.MaxPlanAtoms || len(compiler.variables) > limits.MaxVariables {
		return fail(planResource("expression", "compiled atom or variable count exceeds the structural ceiling"))
	}
	if err := state.charge(compiler.work, "plan_compile"); err != nil {
		return fail(err)
	}
	variables := compiler.variableList()
	for _, variable := range variables {
		if variable.Key.Kind == AlgebraSelectionParameter && !variable.UsedInAtom {
			return fail(planInvalid("UNUSED_PARAMETER_DECLARATION", variable.SourcePath, "unused parameter declarations cannot be selected as query outputs"))
		}
	}

	sourceBytes, err := algebra.Encode(source)
	if err != nil {
		return fail(err)
	}
	if err := state.charge(len(sourceBytes), "source_encode"); err != nil {
		return fail(err)
	}
	sourceHash := digestBytes("algebra-query-source-v1", sourceBytes)
	identity := struct {
		Version        int                 `json:"version"`
		Profile        string              `json:"profile"`
		Context        string              `json:"context"`
		At             string              `json:"at"`
		FactsSnapshot  string              `json:"facts_snapshot"`
		FormalSnapshot string              `json:"formal_snapshot"`
		Source         string              `json:"source"`
		Answers        []AlgebraAnswerPort `json:"answers"`
	}{1, request.Profile, request.ContextID, request.At, schema.SnapshotID, environment.SnapshotID, sourceHash, answers}
	identityBytes, _ := json.Marshal(identity)
	fingerprint := digestBytes("algebra-query-plan-v1", identityBytes)
	plan := &AlgebraQueryPlan{
		profile: request.Profile, contextID: request.ContextID, at: request.At,
		factsSnapshotID: schema.SnapshotID, formalSnapshotID: environment.SnapshotID,
		fingerprint: fingerprint, sourceFingerprint: sourceHash,
		loweringProfile: betaQueryProfile, lowering: clonePlanCertificate(lowerCert),
		answers: cloneAnswerPorts(answers), references: clonePlanDeclarations(source.References), variables: variables, nodes: compiler.nodes,
		root: root, limits: limits, usage: state.usage, deadline: deadline,
		ctx: ctx, cancel: cancel,
	}
	return plan, nil
}

// Describe returns a detached, non-executable plan summary.
func (plan *AlgebraQueryPlan) Describe() AlgebraPlanDescription {
	if plan == nil {
		return AlgebraPlanDescription{}
	}
	return AlgebraPlanDescription{
		SchemaVersion: AlgebraQuerySchemaVersion, Profile: plan.profile, ContextID: plan.contextID,
		At: plan.at, FactsSnapshotID: plan.factsSnapshotID, FormalSnapshotID: plan.formalSnapshotID,
		Fingerprint: plan.fingerprint, SourceFingerprint: plan.sourceFingerprint,
		LoweringProfile: plan.loweringProfile, LoweringCertificate: clonePlanCertificate(plan.lowering),
		Answers: cloneAnswerPorts(plan.answers), References: clonePlanDeclarations(plan.references), Variables: clonePlanVariables(plan.variables),
		Nodes: clonePlanNodes(plan.nodes), Root: plan.root, Limits: plan.limits, Usage: plan.usage,
	}
}

func (plan *AlgebraQueryPlan) close() {
	if plan != nil && plan.cancel != nil {
		plan.cancel()
		plan.cancel = nil
	}
}

type planCompiler struct {
	ctx            context.Context
	limits         AlgebraEvaluationLimits
	work           int
	nodes          []AlgebraPlanNode
	atoms          int
	variables      map[AlgebraVariableKey]*AlgebraPlanVariable
	variableOrder  []AlgebraVariableKey
	parameterTypes map[string]algebra.Type
	referenceTypes map[string]algebra.Type
	scopes         []map[string]AlgebraVariableKey
	usage          AlgebraEvaluationUsage
}

func (compiler *planCompiler) compileTree(term *algebra.Term, path string, depth int) (string, error) {
	if err := compiler.checkContext(); err != nil {
		return "", err
	}
	if term == nil {
		return "", planInvalid("MISSING_QUERY_NODE", path, "query tree node is required")
	}
	if depth > compiler.limits.TermLimits.MaxDepth {
		return "", planResource(path, "query tree exceeds depth limit")
	}
	compiler.work++
	if compiler.work > compiler.limits.MaxWorkUnits {
		return "", planResource(path, "plan compilation exceeds the remaining logical work budget")
	}
	if compiler.work > compiler.limits.MaxPlanAtoms+compiler.limits.MaxVariables+compiler.limits.TermLimits.MaxNodes {
		return "", planResource(path, "plan compilation traversal exceeds its structural bound")
	}
	nodeID := fmt.Sprintf("n%d", len(compiler.nodes)+1)
	idx := len(compiler.nodes)
	compiler.nodes = append(compiler.nodes, AlgebraPlanNode{ID: nodeID, Path: path})
	switch term.Kind {
	case algebra.TermAtom:
		if compiler.atoms >= compiler.limits.MaxPlanAtoms {
			return "", planResource(path, "query atom count exceeds the configured limit")
		}
		compiler.atoms++
		node := AlgebraPlanNode{ID: nodeID, Kind: AlgebraPlanAtom, Path: path, Symbol: term.Symbol, Args: make([]AlgebraPlanArgument, 0, len(term.Args))}
		if len(term.Args) > compiler.limits.TermLimits.MaxPorts {
			return "", planResource(path+".args", "atom arity exceeds the port limit")
		}
		for i, argument := range term.Args {
			argPath := fmt.Sprintf("%s.args[%d]", path, i)
			value, err := compiler.value(argument.Value, argPath)
			if err != nil {
				return "", err
			}
			node.Args = append(node.Args, AlgebraPlanArgument{Role: argument.Role, Value: value.key, RefID: value.refID})
		}
		compiler.nodes[idx] = node
	case algebra.TermAnd:
		if len(term.Operands) == 0 {
			return "", planInvalid("EMPTY_QUERY_CONJUNCTION", path, "And requires at least one operand")
		}
		node := AlgebraPlanNode{ID: nodeID, Kind: AlgebraPlanAnd, Path: path, Children: make([]string, 0, len(term.Operands))}
		for i, operand := range term.Operands {
			child, err := compiler.compileTree(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1)
			if err != nil {
				return "", err
			}
			node.Children = append(node.Children, child)
		}
		compiler.nodes[idx] = node
	case algebra.TermExists:
		key := AlgebraVariableKey{Kind: "exists", ID: term.ID}
		if err := compiler.addVariable(key, *term.Type, path+".id"); err != nil {
			return "", err
		}
		compiler.pushScope(term.ID, key)
		child, err := compiler.compileTree(term.Body, path+".body", depth+1)
		compiler.popScope()
		if err != nil {
			return "", err
		}
		typ := typeCopy(*term.Type)
		compiler.nodes[idx] = AlgebraPlanNode{ID: nodeID, Kind: AlgebraPlanExists, Path: path, Binder: &key, Type: &typ, Children: []string{child}}
	case algebra.TermApply, algebra.TermLambda:
		return "", planUnsupported("UNSUPPORTED_QUERY_CONSTRUCTOR", path, "query body may contain only Atom, And, and Exists after beta lowering")
	case algebra.TermHole:
		return "", planIncomplete("QUERY_HOLE_UNRESOLVED", path, "query contains an unresolved typed hole")
	default:
		return "", planInvalid("INVALID_QUERY_BODY", path, "query body must be proposition structure")
	}
	return nodeID, nil
}

type planValue struct {
	key   AlgebraVariableKey
	refID string
}

func (compiler *planCompiler) value(term *algebra.Term, path string) (planValue, error) {
	if term == nil {
		return planValue{}, planInvalid("MISSING_ATOM_ARGUMENT", path, "atom argument is required")
	}
	switch term.Kind {
	case algebra.TermRef:
		typ, ok := compiler.referenceTypes[term.ID]
		if !ok || term.Type == nil || !algebra.TypeEqual(typ, *term.Type) {
			return planValue{}, planInvalid("UNKNOWN_QUERY_REFERENCE", path, "reference is not declared with this exact type")
		}
		return planValue{key: AlgebraVariableKey{Kind: string(algebra.TermRef), ID: term.ID}, refID: term.ID}, nil
	case algebra.TermVar:
		key, ok := compiler.lookupBinder(term.ID)
		if !ok {
			return planValue{}, planInvalid("UNBOUND_QUERY_VARIABLE", path, "variable has no lexical binder")
		}
		variable := compiler.variables[key]
		variable.UsedInAtom = true
		return planValue{key: key}, nil
	case algebra.TermParam:
		typ, ok := compiler.parameterTypes[term.ID]
		if !ok || term.Type == nil || !algebra.TypeEqual(typ, *term.Type) {
			return planValue{}, planInvalid("UNKNOWN_QUERY_PARAMETER", path, "parameter is not declared with this exact type")
		}
		key := AlgebraVariableKey{Kind: string(algebra.TermParam), ID: term.ID}
		if err := compiler.addVariable(key, typ, "parameters."+term.ID); err != nil {
			return planValue{}, err
		}
		compiler.variables[key].UsedInAtom = true
		return planValue{key: key}, nil
	case algebra.TermHole:
		return planValue{}, planIncomplete("QUERY_HOLE_UNRESOLVED", path, "query contains an unresolved typed hole")
	default:
		return planValue{}, planUnsupported("UNSUPPORTED_ATOM_ARGUMENT", path, "atom arguments must be Ref, Var, or Param values")
	}
}

func (compiler *planCompiler) addVariable(key AlgebraVariableKey, typ algebra.Type, path string) error {
	if compiler.variables == nil {
		compiler.variables = make(map[AlgebraVariableKey]*AlgebraPlanVariable)
	}
	if _, exists := compiler.variables[key]; exists {
		return nil
	}
	if len(compiler.variables) >= compiler.limits.MaxVariables {
		return planResource(path, "query variable count exceeds the configured limit")
	}
	compiler.variables[key] = &AlgebraPlanVariable{Key: key, Type: typeCopy(typ), SourcePath: path}
	compiler.variableOrder = append(compiler.variableOrder, key)
	return nil
}

func (compiler *planCompiler) pushSelectedLambdas(lambdas []*algebra.Term) error {
	compiler.scopes = append(compiler.scopes, make(map[string]AlgebraVariableKey))
	for _, lambda := range lambdas {
		key := AlgebraVariableKey{Kind: AlgebraSelectionLambda, ID: lambda.ID}
		if err := compiler.addVariable(key, *lambda.Type, "expression"); err != nil {
			return err
		}
		compiler.scopes[len(compiler.scopes)-1][lambda.ID] = key
	}
	return nil
}

func (compiler *planCompiler) pushScope(id string, key AlgebraVariableKey) {
	compiler.scopes = append(compiler.scopes, map[string]AlgebraVariableKey{id: key})
}

func (compiler *planCompiler) popScope() {
	if len(compiler.scopes) > 0 {
		compiler.scopes = compiler.scopes[:len(compiler.scopes)-1]
	}
}

func (compiler *planCompiler) lookupBinder(id string) (AlgebraVariableKey, bool) {
	for i := len(compiler.scopes) - 1; i >= 0; i-- {
		if key, ok := compiler.scopes[i][id]; ok {
			return key, true
		}
	}
	return AlgebraVariableKey{}, false
}

func (compiler *planCompiler) variableList() []AlgebraPlanVariable {
	out := make([]AlgebraPlanVariable, 0, len(compiler.variableOrder))
	for _, key := range compiler.variableOrder {
		if value := compiler.variables[key]; value != nil {
			out = append(out, *value)
		}
	}
	return out
}

func (compiler *planCompiler) charge(amount int, path string) error {
	if amount < 0 || amount > compiler.limits.MaxWorkUnits-compiler.work {
		return planResource(path, "query compilation exceeds the logical work budget")
	}
	compiler.work += amount
	compiler.usage.WorkUnits += amount
	return compiler.checkContext()
}

func (compiler *planCompiler) remaining() int { return compiler.limits.MaxWorkUnits - compiler.work }

func (compiler *planCompiler) checkContext() error {
	select {
	case <-compiler.ctx.Done():
		if compiler.ctx.Err() == context.DeadlineExceeded {
			return planResource("limits.timeout_millis", "query planning deadline expired")
		}
		return &algebra.Error{Status: status.OutcomeIncomplete, Code: "QUERY_CANCELLED", Path: "context", Message: "query planning was cancelled"}
	default:
		return nil
	}
}

type binderSelection struct{ term *algebra.Term }

func compileSelections(selection []AlgebraSelection, root *algebra.Term, checked algebra.CheckedExpression) ([]AlgebraAnswerPort, []*algebra.Term, error) {
	leading := make([]*algebra.Term, 0)
	for node := root; node != nil && node.Kind == algebra.TermLambda; node = node.Body {
		leading = append(leading, node)
	}
	if len(leading) > DefaultAlgebraMaxVariables {
		return nil, nil, planResource("expression", "leading lambda count exceeds the variable ceiling")
	}
	params := make(map[string]algebra.Type, len(checked.FreeParameters))
	for _, declaration := range checked.FreeParameters {
		params[declaration.ID] = declaration.Type
	}
	seenLambda := make(map[int]bool)
	seenParams := make(map[string]bool)
	labels := make(map[string]bool)
	answers := make([]AlgebraAnswerPort, 0, len(selection))
	for i, item := range selection {
		path := fmt.Sprintf("select[%d]", i)
		if !validPlanText(item.As, 256) || labels[item.As] {
			return nil, nil, planInvalid("INVALID_ANSWER_LABEL", path+".as", "answer labels must be unique bounded UTF-8 text")
		}
		labels[item.As] = true
		switch item.Kind {
		case AlgebraSelectionLambda:
			if item.Index == nil || item.ID != "" || *item.Index < 0 || *item.Index >= len(leading) {
				return nil, nil, planInvalid("INVALID_LAMBDA_SELECTION", path, "lambda selection requires a valid zero-based index and no ID")
			}
			if seenLambda[*item.Index] {
				return nil, nil, planInvalid("DUPLICATE_ANSWER_SELECTION", path, "leading lambda is selected more than once")
			}
			seenLambda[*item.Index] = true
			lambda := leading[*item.Index]
			answers = append(answers, AlgebraAnswerPort{As: item.As, Variable: AlgebraVariableKey{Kind: AlgebraSelectionLambda, ID: lambda.ID}, Type: typeCopy(*lambda.Type)})
		case AlgebraSelectionParameter:
			if item.Index != nil || !validPlanText(item.ID, 256) {
				return nil, nil, planInvalid("INVALID_PARAMETER_SELECTION", path, "parameter selection requires an ID and no index")
			}
			typ, used := params[item.ID]
			if !used {
				return nil, nil, planInvalid("UNUSED_PARAMETER_SELECTION", path+".id", "only a used free parameter may be selected")
			}
			if seenParams[item.ID] {
				return nil, nil, planInvalid("DUPLICATE_ANSWER_SELECTION", path, "parameter is selected more than once")
			}
			seenParams[item.ID] = true
			answers = append(answers, AlgebraAnswerPort{As: item.As, Variable: AlgebraVariableKey{Kind: string(algebra.TermParam), ID: item.ID}, Type: typeCopy(typ)})
		default:
			return nil, nil, planUnsupported("UNSUPPORTED_ANSWER_SELECTION", path+".kind", "selection kind must be lambda or parameter")
		}
	}
	if len(seenLambda) != len(leading) {
		return nil, nil, planInvalid("UNSELECTED_OPEN_LAMBDA", "select", "every leading lambda must be explicitly selected")
	}
	if len(seenParams) != len(params) {
		return nil, nil, planInvalid("UNSELECTED_FREE_PARAMETER", "select", "every used free parameter must be explicitly selected")
	}
	return answers, leading, nil
}

func lowerQuery(source algebra.Document, environment algebra.RewriteEnvironment, limits algebra.Limits, remaining int) (*algebra.Term, algebra.RewriteCertificate, int, error) {
	if remaining <= 0 {
		return nil, algebra.RewriteCertificate{}, 0, planResource("limits.max_work_units", "no logical work remains for beta lowering")
	}
	profile := algebra.RewriteProfile{ID: betaQueryProfile, Rules: []algebra.RewriteRule{algebra.RewriteBetaV1}}
	env := environment
	env.Document.Limits = limits
	env.Document.Parameters = append([]algebra.Declaration{}, source.Parameters...)
	env.Document.References = append([]algebra.Declaration{}, source.References...)
	env.Document.Closed = source.Closed
	if err := algebra.ValidateRewriteEnvironment(env); err != nil {
		return nil, algebra.RewriteCertificate{}, 0, err
	}
	budget := algebra.RewriteBudget{MaxWorkNodes: minPositive(remaining, algebra.DefaultRewriteMaxWorkNodes)}
	result, err := algebra.Normalize(env, profile, source.Expression, budget)
	if err != nil {
		return nil, algebra.RewriteCertificate{}, 0, err
	}
	work := result.Statistics.WorkNodes
	if result.Status != status.OutcomeOK || !result.Normalized || result.Expression == nil {
		if result.Status == status.OutcomeResourceLimit {
			return nil, result.Certificate, work, planResource("expression", "beta lowering exhausted its bounded budget")
		}
		if result.Status == status.OutcomeIncomplete {
			return nil, result.Certificate, work, planIncomplete("QUERY_LOWERING_INCOMPLETE", "expression", "beta lowering did not complete")
		}
		return nil, result.Certificate, work, planUnsupported("QUERY_LOWERING_UNSUPPORTED", "expression", "beta lowering did not produce a supported normal form")
	}
	verifyRemaining := remaining - work
	if verifyRemaining <= 0 {
		return nil, result.Certificate, work, planResource("expression", "no logical work remains for independent beta-certificate verification")
	}
	verifyBudget := algebra.RewriteBudget{MaxWorkNodes: minPositive(verifyRemaining, algebra.DefaultRewriteMaxWorkNodes)}
	verified, err := algebra.VerifyRewriteCertificate(env, profile, source.Expression, result.Certificate, verifyBudget)
	work += verified.Statistics.WorkNodes
	if err != nil {
		return nil, result.Certificate, work, err
	}
	if verified.Status != status.OutcomeOK {
		switch verified.Status {
		case status.OutcomeResourceLimit:
			return nil, result.Certificate, work, planResource("expression", "beta certificate verification exhausted its bounded budget")
		case status.OutcomeIncomplete:
			return nil, result.Certificate, work, planIncomplete("QUERY_LOWERING_UNVERIFIED", "expression", "beta lowering certificate verification was incomplete")
		case status.OutcomeUnsupported:
			return nil, result.Certificate, work, planUnsupported("QUERY_LOWERING_UNVERIFIED", "expression", "beta lowering certificate verification was unsupported")
		default:
			return nil, result.Certificate, work, planInvalid("QUERY_LOWERING_UNVERIFIED", "expression", "beta lowering certificate verification did not succeed")
		}
	}
	if !verified.CertificateValid || !verified.NormalizationComplete {
		return nil, result.Certificate, work, planIncomplete("QUERY_LOWERING_UNVERIFIED", "expression", "beta lowering certificate was not independently verified")
	}
	return result.Expression, result.Certificate, work, nil
}

func validateAlgebraPlanInputs(schema facts.NarySchema, environment algebra.RewriteEnvironment, source algebra.Document, request AlgebraQueryRequest, limits AlgebraEvaluationLimits) error {
	if request.SchemaVersion != AlgebraQuerySchemaVersion {
		return planInvalid("INVALID_QUERY_SCHEMA_VERSION", "schema_version", "unsupported algebra query request version")
	}
	if request.Profile != AlgebraQueryProfile {
		return planUnsupported("UNSUPPORTED_QUERY_PROFILE", "profile", "query profile is not supported")
	}
	if !validPlanText(request.ContextID, 256) {
		return planInvalid("INVALID_QUERY_CONTEXT", "context_id", "context ID must be bounded valid UTF-8")
	}
	if schema.SnapshotID == "" || !validPlanText(schema.SnapshotID, 256) {
		return planInvalid("INVALID_FACT_SNAPSHOT", "metadata.facts_snapshot_id", "selected fact schema requires its own bounded snapshot ID")
	}
	if !validPlanText(environment.SnapshotID, 256) {
		return planInvalid("INVALID_FORMAL_SNAPSHOT", "metadata.formal_snapshot_id", "rewrite environment requires its own bounded formal snapshot ID")
	}
	if request.At != "" {
		if len(request.At) > 128 || !strings.HasSuffix(request.At, "Z") {
			return planInvalid("INVALID_QUERY_TIME", "at", "query instant must be an explicit RFC3339 UTC timestamp")
		}
		if _, err := time.Parse(time.RFC3339Nano, request.At); err != nil {
			return planInvalid("INVALID_QUERY_TIME", "at", "query instant must be an explicit RFC3339 UTC timestamp")
		}
	}
	if source.SchemaVersion != algebra.SchemaVersion || source.Expression == nil {
		return planInvalid("INVALID_QUERY_DOCUMENT", "expression", "a schema-version 1 self-contained expression document is required")
	}
	if len(source.Sorts)+len(source.Signatures)+len(source.Parameters)+len(source.References) > limits.TermLimits.MaxSymbols {
		return planResource("document", "query declarations exceed term symbol limits")
	}
	if schema.Document.SchemaVersion != algebra.SchemaVersion || schema.Document.Expression != nil || len(schema.Document.Parameters) != 0 || len(schema.Document.References) != 0 {
		return planInvalid("INVALID_FACT_SCHEMA", "metadata", "selected fact schema must contain declarations only")
	}
	if environment.Document.Expression != nil {
		return planInvalid("INVALID_FORMAL_ENVIRONMENT", "metadata.environment.expression", "rewrite environment must contain declarations only")
	}
	if err := algebra.CheckDeclarations(schema.Document); err != nil {
		return err
	}
	if err := algebra.CheckDeclarations(environment.Document); err != nil {
		return err
	}
	if err := algebra.CheckDeclarations(source); err != nil {
		return err
	}
	if err := algebra.ValidateRewriteEnvironment(environment); err != nil {
		return err
	}
	if len(request.Select) > limits.MaxVariables {
		return planResource("select", "answer selection count exceeds the variable limit")
	}
	return nil
}

func compareSchemaWithFormal(schema algebra.Document, environment algebra.RewriteEnvironment) error {
	formal := make(map[string]algebra.Signature, len(environment.Document.Signatures))
	for _, signature := range environment.Document.Signatures {
		formal[signature.ID] = signature
	}
	for i, signature := range schema.Signatures {
		candidate, ok := formal[signature.ID]
		if !ok || !samePlanSignature(signature, candidate) {
			return planInvalid("FACT_FORMAL_INVENTORY_MISMATCH", fmt.Sprintf("metadata.signatures[%d]", i), "fact and formal snapshots must declare identical primitive signatures")
		}
	}
	formalSorts := make(map[string]bool, len(environment.Document.Sorts))
	for _, sortDecl := range environment.Document.Sorts {
		formalSorts[sortDecl.ID] = true
	}
	for i, sortDecl := range schema.Sorts {
		if !formalSorts[sortDecl.ID] {
			return planInvalid("FACT_FORMAL_INVENTORY_MISMATCH", fmt.Sprintf("metadata.sorts[%d]", i), "fact and formal snapshots must contain the same nominal sorts")
		}
	}
	return nil
}

func compareSourceInventory(source, selected algebra.Document) error {
	selectedSorts := make(map[string]algebra.SortDeclaration, len(selected.Sorts))
	for _, item := range selected.Sorts {
		selectedSorts[item.ID] = item
	}
	for i, item := range source.Sorts {
		if actual, ok := selectedSorts[item.ID]; !ok || actual.ID != item.ID {
			return planInvalid("INLINE_SORT_NOT_SELECTED", fmt.Sprintf("sorts[%d]", i), "inline sort is not in the selected fact schema")
		}
	}
	selectedSignatures := make(map[string]algebra.Signature, len(selected.Signatures))
	for _, item := range selected.Signatures {
		selectedSignatures[item.ID] = item
	}
	for i, item := range source.Signatures {
		actual, ok := selectedSignatures[item.ID]
		if !ok || !samePlanSignature(item, actual) {
			return planInvalid("INLINE_SIGNATURE_NOT_SELECTED", fmt.Sprintf("signatures[%d]", i), "inline relation must exactly match an approved selected primitive")
		}
	}
	return nil
}

func compareDeclarations(source, selected []algebra.Declaration, path string) error {
	if len(source) != len(selected) {
		return planInvalid("QUERY_DECLARATION_MISMATCH", path, "source declarations must exactly match the selected normalization environment")
	}
	byID := make(map[string]algebra.Type, len(selected))
	for _, declaration := range selected {
		byID[declaration.ID] = declaration.Type
	}
	for i, declaration := range source {
		typ, ok := byID[declaration.ID]
		if !ok || !algebra.TypeEqual(typ, declaration.Type) {
			return planInvalid("QUERY_DECLARATION_MISMATCH", fmt.Sprintf("%s[%d]", path, i), "source ID and type must match the selected normalization environment")
		}
	}
	return nil
}

func samePlanSignature(a, b algebra.Signature) bool {
	if a.ID != b.ID || len(a.Ports) != len(b.Ports) {
		return false
	}
	for i := range a.Ports {
		if a.Ports[i].Role != b.Ports[i].Role || !algebra.TypeEqual(a.Ports[i].Type, b.Ports[i].Type) {
			return false
		}
	}
	return true
}

func clonePlanSignatures(in []algebra.Signature) []algebra.Signature {
	out := make([]algebra.Signature, len(in))
	for i, signature := range in {
		out[i] = algebra.Signature{ID: signature.ID, Label: signature.Label, Ports: make([]algebra.Port, len(signature.Ports))}
		for j, port := range signature.Ports {
			out[i].Ports[j] = algebra.Port{Role: port.Role, Type: typeCopy(port.Type)}
		}
	}
	return out
}

func typeCopy(value algebra.Type) algebra.Type {
	copyType := algebra.Type{Kind: value.Kind, SortID: value.SortID}
	if value.From != nil {
		from := typeCopy(*value.From)
		copyType.From = &from
	}
	if value.To != nil {
		to := typeCopy(*value.To)
		copyType.To = &to
	}
	return copyType
}

func withSelectedInventory(source algebra.Document, selected algebra.Document, expression *algebra.Term, limits algebra.Limits) algebra.Document {
	out := source
	out.Sorts = append([]algebra.SortDeclaration{}, selected.Sorts...)
	out.Signatures = clonePlanSignatures(selected.Signatures)
	out.Limits = limits
	out.Expression = expression
	return out
}

func digestBytes(domain string, payload []byte) string {
	h := sha256.New()
	h.Write([]byte(domain))
	h.Write([]byte{0})
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func validPlanText(value string, max int) bool {
	return value != "" && len(value) <= max && utf8.ValidString(value)
}

func effectivePlanTermLimits(requested algebra.Limits, documents ...algebra.Limits) (algebra.Limits, error) {
	effective, err := algebra.NormalizeLimits(requested)
	if err != nil {
		return algebra.Limits{}, err
	}
	for _, candidate := range documents {
		normalized, normalizeErr := algebra.NormalizeLimits(candidate)
		if normalizeErr != nil {
			return algebra.Limits{}, normalizeErr
		}
		if normalized.MaxNodes < effective.MaxNodes {
			effective.MaxNodes = normalized.MaxNodes
		}
		if normalized.MaxDepth < effective.MaxDepth {
			effective.MaxDepth = normalized.MaxDepth
		}
		if normalized.MaxTypeDepth < effective.MaxTypeDepth {
			effective.MaxTypeDepth = normalized.MaxTypeDepth
		}
		if normalized.MaxSymbols < effective.MaxSymbols {
			effective.MaxSymbols = normalized.MaxSymbols
		}
		if normalized.MaxPorts < effective.MaxPorts {
			effective.MaxPorts = normalized.MaxPorts
		}
	}
	return effective, nil
}

func contextPlanCheck(ctx context.Context) error {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return planResource("limits.timeout_millis", "query planning deadline expired")
		}
		return &algebra.Error{Status: status.OutcomeIncomplete, Code: "QUERY_CANCELLED", Path: "context", Message: "query planning was cancelled"}
	default:
		return nil
	}
}

func minPositive(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cloneAnswerPorts(in []AlgebraAnswerPort) []AlgebraAnswerPort {
	out := make([]AlgebraAnswerPort, len(in))
	for i, value := range in {
		out[i] = AlgebraAnswerPort{As: value.As, Variable: value.Variable, Type: typeCopy(value.Type)}
	}
	return out
}
func clonePlanDeclarations(in []algebra.Declaration) []algebra.Declaration {
	out := make([]algebra.Declaration, len(in))
	for i, declaration := range in {
		out[i] = algebra.Declaration{ID: declaration.ID, Type: typeCopy(declaration.Type)}
	}
	return out
}
func clonePlanVariables(in []AlgebraPlanVariable) []AlgebraPlanVariable {
	out := make([]AlgebraPlanVariable, len(in))
	for i, v := range in {
		out[i] = AlgebraPlanVariable{Key: v.Key, Type: typeCopy(v.Type), SourcePath: v.SourcePath, UsedInAtom: v.UsedInAtom}
	}
	return out
}
func clonePlanNodes(in []AlgebraPlanNode) []AlgebraPlanNode {
	out := make([]AlgebraPlanNode, len(in))
	for i, n := range in {
		out[i] = n
		out[i].Args = append([]AlgebraPlanArgument{}, n.Args...)
		out[i].Children = append([]string{}, n.Children...)
		if n.Binder != nil {
			b := *n.Binder
			out[i].Binder = &b
		}
		if n.Type != nil {
			typ := typeCopy(*n.Type)
			out[i].Type = &typ
		}
	}
	return out
}
func clonePlanCertificate(in algebra.RewriteCertificate) algebra.RewriteCertificate {
	out := in
	out.Profile.Rules = append([]algebra.RewriteRule{}, in.Profile.Rules...)
	out.Input = clonePlanTerm(in.Input)
	out.Output = clonePlanTerm(in.Output)
	out.Steps = make([]algebra.RewriteStep, len(in.Steps))
	for i, step := range in.Steps {
		out.Steps[i] = step
		out.Steps[i].Path = append([]int{}, step.Path...)
		out.Steps[i].Before = clonePlanTerm(step.Before)
		out.Steps[i].After = clonePlanTerm(step.After)
		out.Steps[i].BinderMapping = append([]algebra.BinderChange{}, step.BinderMapping...)
		out.Steps[i].SideConditions = append([]string{}, step.SideConditions...)
		out.Steps[i].ParameterMapping = make(algebra.Substitution, len(step.ParameterMapping))
		for id, term := range step.ParameterMapping {
			out.Steps[i].ParameterMapping[id] = clonePlanTerm(term)
		}
	}
	return out
}

func clonePlanTerm(term *algebra.Term) *algebra.Term {
	if term == nil {
		return nil
	}
	out := *term
	if term.Type != nil {
		value := typeCopy(*term.Type)
		out.Type = &value
	}
	out.Args = make([]algebra.Argument, len(term.Args))
	for i, argument := range term.Args {
		out.Args[i] = algebra.Argument{Role: argument.Role, Value: clonePlanTerm(argument.Value)}
	}
	out.Operands = make([]*algebra.Term, len(term.Operands))
	for i, operand := range term.Operands {
		out.Operands[i] = clonePlanTerm(operand)
	}
	out.Body, out.Function, out.Argument = clonePlanTerm(term.Body), clonePlanTerm(term.Function), clonePlanTerm(term.Argument)
	return &out
}

func planInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}
func planUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}
func planIncomplete(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeIncomplete, Code: code, Path: path, Message: message}
}
func planResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

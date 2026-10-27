package reason

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/status"
)

type algebraRoleReferenceKey struct {
	Symbol string
	Role   string
	RefID  string
}

type algebraEvidenceRow struct {
	key      string
	atom     facts.NaryAtom
	positive []facts.NaryFact
	negative []facts.NaryFact
}

// algebraEvaluationState owns one request's active evidence indexes and one
// shared work/deadline budget. The plan and fact view remain immutable.
type algebraEvaluationState struct {
	plan          *AlgebraQueryPlan
	view          *facts.NaryView
	schema        algebra.Document
	ctx           context.Context
	planCtx       context.Context
	cancel        context.CancelFunc
	limits        AlgebraEvaluationLimits
	usage         AlgebraEvaluationUsage
	knownRefs     map[string]algebra.Declaration
	planVariables map[AlgebraVariableKey]AlgebraPlanVariable
	sourceAtoms   []AlgebraPlanNode
	rows          []algebraEvidenceRow
	bySymbol      map[string][]int
	byRoleRef     map[algebraRoleReferenceKey][]int
	byAtom        map[string]int
	result        AlgebraEvaluationResult
}

// prepareAlgebraEvaluation validates the compiled plan against one actual
// immutable fact view, snapshots active evidence exactly once, and builds
// bounded request-local lookup indexes. It does not join rows or create
// answers.
func prepareAlgebraEvaluation(parent context.Context, plan *AlgebraQueryPlan, view *facts.NaryView, requested AlgebraEvaluationLimits) (*algebraEvaluationState, error) {
	if parent == nil {
		return nil, planInvalid("INVALID_CONTEXT", "context", "a caller context is required")
	}
	if plan == nil || plan.ctx == nil || plan.deadline.IsZero() || plan.profile != AlgebraQueryProfile || !validPlanText(plan.contextID, 256) || !validPlanText(plan.factsSnapshotID, 256) || !validPlanText(plan.formalSnapshotID, 256) {
		return nil, planInvalid("INVALID_QUERY_PLAN", "plan", "query plan is nil, uninitialized, or outside the supported profile")
	}
	if view == nil {
		return nil, planInvalid("INVALID_FACT_VIEW", "facts", "a validated request fact view is required")
	}
	limits, err := NormalizeAlgebraEvaluationLimits(requested)
	if err != nil {
		return nil, err
	}
	limits = intersectAlgebraEvaluationLimits(limits, plan.limits)
	if err := plan.initializedLimitsError(); err != nil {
		return nil, err
	}
	limits.TermLimits, err = effectivePlanTermLimits(limits.TermLimits, plan.limits.TermLimits)
	if err != nil {
		return nil, err
	}
	if plan.usage.WorkUnits > limits.MaxWorkUnits {
		return nil, planResource("limits.max_work_units", "query compilation already exceeds the evaluation work budget")
	}
	if view.SnapshotID() != plan.factsSnapshotID {
		return nil, planInvalid("FACT_SNAPSHOT_MISMATCH", "facts.snapshot_id", "fact view does not match the query plan's facts snapshot")
	}
	if view.ContextID() != plan.contextID {
		return nil, planInvalid("FACT_CONTEXT_MISMATCH", "facts.context_id", "fact view does not match the query plan context")
	}
	if !samePlanInstant(plan.at, view.At()) {
		return nil, planInvalid("FACT_INSTANT_MISMATCH", "facts.at", "fact view instant and query plan instant differ")
	}
	if view.Count() > facts.DefaultNaryMaxFacts {
		return nil, planResource("facts", "fact view exceeds the n-ary evidence ceiling")
	}

	deadline := time.Now().Add(time.Duration(limits.TimeoutMillis) * time.Millisecond)
	if parentDeadline, ok := parent.Deadline(); ok && parentDeadline.Before(deadline) {
		deadline = parentDeadline
	}
	if plan.deadline.Before(deadline) {
		deadline = plan.deadline
	}
	ctx, cancel := context.WithDeadline(parent, deadline)
	state := &algebraEvaluationState{
		plan: plan, view: view, ctx: ctx, planCtx: plan.ctx, cancel: cancel,
		limits: limits, usage: plan.usage, knownRefs: make(map[string]algebra.Declaration),
		bySymbol: make(map[string][]int), byRoleRef: make(map[algebraRoleReferenceKey][]int),
		byAtom: make(map[string]int), result: newAlgebraEvaluationResult(),
	}
	state.result.Status = status.OutcomeIncomplete
	state.result.Profile = plan.profile
	state.result.ContextID = plan.contextID
	state.result.At = plan.at
	state.result.FactsSnapshotID = plan.factsSnapshotID
	state.result.FormalSnapshotID = plan.formalSnapshotID
	state.result.PlanFingerprint = plan.fingerprint
	state.result.Columns = cloneAnswerPorts(plan.answers)
	state.result.Plan = func() *AlgebraPlanDescription { description := plan.Describe(); return &description }()
	state.result.Limits = limits
	state.result.Usage = state.usage
	state.result.Complete = false
	state.result.Knowledge = AlgebraKnowledgeUnknown
	if err := state.checkContext(); err != nil {
		cancel()
		return nil, err
	}

	// Reserve for the worst bounded SummaryAt materialization before any view
	// copy. The unused portion is released by charging only measured work below.
	// SummaryAt scans every stored record and argument, and the private index
	// then makes a second bounded pass over the active records. Reserve for both
	// passes before asking the view to materialize its detached summary.
	reservation := view.Count()*(2*algebra.DefaultMaxPorts+5) + facts.DefaultNaryMaxReferents + algebra.DefaultMaxSymbols + len(plan.nodes) + len(plan.variables) + len(plan.references)
	if reservation < 0 || reservation > state.remaining() {
		cancel()
		return nil, planResource("facts", "remaining work budget cannot reserve bounded fact summary materialization")
	}
	if err := state.checkContext(); err != nil {
		cancel()
		return nil, err
	}

	schema := view.Schema()
	if err := validateEvaluationSchema(schema); err != nil {
		cancel()
		return nil, err
	}
	summary, err := view.SummaryAt(plan.at)
	if err != nil {
		cancel()
		return nil, factsErrorAsAlgebra(err)
	}
	if summary.SnapshotID != plan.factsSnapshotID || summary.ContextID != plan.contextID || !samePlanInstant(summary.At, plan.at) {
		cancel()
		return nil, planInvalid("FACT_SUMMARY_MISMATCH", "facts.summary", "active evidence summary does not match the selected plan snapshot, context, and instant")
	}
	if err := state.checkContext(); err != nil {
		cancel()
		return nil, err
	}
	referents := view.Referents()
	known, err := mergeEvaluationReferences(schema, plan.references, referents, plan.requestContext, limits)
	if err != nil {
		cancel()
		return nil, err
	}
	if err := validateUsedKnownReferenceAttributes(plan.nodes, known, plan.requestContext, schema, plan.variables, plan.at); err != nil {
		cancel()
		return nil, err
	}
	if err := validateEvaluationPlan(plan, schema, known, limits); err != nil {
		cancel()
		return nil, err
	}
	state.schema = schema
	state.knownRefs = known

	actualWork, argumentCount := summaryMaterializationWork(summary, schema, known, plan)
	if actualWork > reservation {
		cancel()
		return nil, planResource("facts", "active evidence materialization exceeded its conservative reservation")
	}
	if err := state.charge(actualWork, "facts.summary"); err != nil {
		cancel()
		return nil, err
	}
	if argumentCount > facts.DefaultNaryMaxFacts*algebra.DefaultMaxPorts {
		cancel()
		return nil, planResource("facts.summary.atoms", "active evidence argument count exceeds the bounded index ceiling")
	}
	if err := state.buildIndexes(summary); err != nil {
		cancel()
		return nil, err
	}
	indexWork := len(state.rows) + argumentCount
	if err := state.charge(indexWork, "facts.index"); err != nil {
		cancel()
		return nil, err
	}
	state.usage.IndexedFacts = summary.ActiveFacts
	if err := state.checkContext(); err != nil {
		cancel()
		return nil, err
	}
	state.result.InputEvidence = AlgebraInputEvidenceSummary{
		StoredFacts: summary.StoredFacts, ActiveFacts: summary.ActiveFacts,
		InactiveFacts: summary.InactiveFacts, PositiveRecords: state.positiveRecordCount(),
		NegativeRecords: state.negativeRecordCount(), Conflicts: cloneEvaluationConflicts(summary.Conflicts),
	}
	state.result.EvidenceFingerprint, err = state.evidenceFingerprint()
	if err != nil {
		cancel()
		return nil, err
	}
	state.setClosedGroundAtomEvidence()
	state.result.Usage = state.usage
	state.result.Status = status.OutcomeIncomplete
	state.result.Knowledge = AlgebraKnowledgeUnknown
	state.result.Complete = false
	return state, nil
}

func validateUsedKnownReferenceAttributes(nodes []AlgebraPlanNode, known map[string]algebra.Declaration, requestContext *algebra.RequestContext, schema algebra.Document, variables []AlgebraPlanVariable, at string) error {
	used := make(map[string]bool)
	for _, node := range nodes {
		for _, argument := range node.Args {
			if argument.RefID != "" {
				used[argument.RefID] = true
			}
		}
	}
	for _, variable := range variables {
		if variable.UsedInAtom {
			continue
		}
		for id, declaration := range known {
			if algebra.TypeEqual(variable.Type, declaration.Type) {
				used[id] = true
			}
		}
	}
	ids := make([]string, 0, len(used))
	for id := range used {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	document := schema
	document.References = make([]algebra.Declaration, 0, len(ids))
	for _, id := range ids {
		declaration, ok := known[id]
		if !ok {
			return planInvalid("UNKNOWN_PLAN_REFERENCE", "references."+id, "query reference is absent from the explicit request/fact identity union")
		}
		document.References = append(document.References, declaration)
	}
	for _, id := range ids {
		attributes, err := algebra.EffectiveReferenceAttributes(document, requestContext, id)
		if err != nil {
			return err
		}
		if attributes.TemporalFrame != "" {
			declaration, ok := known[id]
			if !ok || declaration.Type.Kind != algebra.TypeEvent {
				return planInvalid("ATTRIBUTE_TYPE_MISMATCH", "references."+id+".temporal_frame", "temporal frame is available only on Event references")
			}
			if attributes.TemporalFrame != "present" {
				return planUnsupported("UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE", "references."+id+".temporal_frame", "only present Event references are supported by the bounded explicit-time query profile")
			}
			if at == "" {
				return planUnsupported("QUERY_INSTANT_REQUIRED", "at", "a present Event reference requires an explicit UTC query instant")
			}
		}
		if attributes.ReferenceTimeID != "" {
			return planUnsupported("UNSUPPORTED_REFERENCE_TIME", "references."+id+".reference_time_id", "reference-time links are not mapped by the positive-facts query profile")
		}
		if attributes.Multiplicity == "many" {
			return planUnsupported("UNSUPPORTED_MANY_REFERENCE", "references."+id+".multiplicity", "the positive-facts query profile does not expand or group a many-valued reference")
		}
	}
	return nil
}

func (plan *AlgebraQueryPlan) initializedLimitsError() error {
	if plan.limits.MaxWorkUnits <= 0 || plan.limits.MaxPlanAtoms <= 0 || plan.limits.MaxVariables <= 0 || plan.limits.MaxIntermediateRows <= 0 || plan.limits.MaxResults <= 0 || plan.limits.MaxProofNodes <= 0 || plan.limits.MaxProofsPerAnswer <= 0 || plan.limits.MaxProofDepth <= 0 || plan.limits.MaxOutputBytes <= 0 || plan.limits.TimeoutMillis <= 0 {
		return planInvalid("INVALID_QUERY_PLAN_LIMITS", "plan.limits", "compiled plan must carry normalized finite limits")
	}
	if _, err := algebra.NormalizeLimits(plan.limits.TermLimits); err != nil {
		return err
	}
	return nil
}

func intersectAlgebraEvaluationLimits(requested, compiled AlgebraEvaluationLimits) AlgebraEvaluationLimits {
	if compiled.MaxPlanAtoms < requested.MaxPlanAtoms {
		requested.MaxPlanAtoms = compiled.MaxPlanAtoms
	}
	if compiled.MaxVariables < requested.MaxVariables {
		requested.MaxVariables = compiled.MaxVariables
	}
	if compiled.MaxIntermediateRows < requested.MaxIntermediateRows {
		requested.MaxIntermediateRows = compiled.MaxIntermediateRows
	}
	if compiled.MaxResults < requested.MaxResults {
		requested.MaxResults = compiled.MaxResults
	}
	if compiled.MaxProofNodes < requested.MaxProofNodes {
		requested.MaxProofNodes = compiled.MaxProofNodes
	}
	if compiled.MaxProofsPerAnswer < requested.MaxProofsPerAnswer {
		requested.MaxProofsPerAnswer = compiled.MaxProofsPerAnswer
	}
	if compiled.MaxProofDepth < requested.MaxProofDepth {
		requested.MaxProofDepth = compiled.MaxProofDepth
	}
	if compiled.MaxWorkUnits < requested.MaxWorkUnits {
		requested.MaxWorkUnits = compiled.MaxWorkUnits
	}
	if compiled.MaxOutputBytes < requested.MaxOutputBytes {
		requested.MaxOutputBytes = compiled.MaxOutputBytes
	}
	if compiled.TimeoutMillis < requested.TimeoutMillis {
		requested.TimeoutMillis = compiled.TimeoutMillis
	}
	return requested
}

func (state *algebraEvaluationState) close() {
	if state != nil && state.cancel != nil {
		state.cancel()
		state.cancel = nil
	}
}

func (state *algebraEvaluationState) checkContext() error {
	if state == nil || state.ctx == nil {
		return planInvalid("INVALID_EVALUATION_STATE", "evaluation", "evaluation state is not initialized")
	}
	if state.planCtx != nil && state.planCtx.Err() != nil {
		if errors.Is(state.planCtx.Err(), context.DeadlineExceeded) {
			state.result.Status = status.OutcomeResourceLimit
			state.result.StopReason = AlgebraStopTimeout
			return planResource("limits.timeout_millis", "compiled query deadline expired")
		}
		state.result.Status = status.OutcomeIncomplete
		state.result.StopReason = AlgebraStopCancelled
		return &algebra.Error{Status: status.OutcomeIncomplete, Code: "QUERY_CANCELLED", Path: "context", Message: "compiled query was cancelled"}
	}
	select {
	case <-state.ctx.Done():
		if errors.Is(state.ctx.Err(), context.DeadlineExceeded) {
			state.result.Status = status.OutcomeResourceLimit
			state.result.StopReason = AlgebraStopTimeout
			return planResource("limits.timeout_millis", "evaluation deadline expired")
		}
		state.result.Status = status.OutcomeIncomplete
		state.result.StopReason = AlgebraStopCancelled
		return &algebra.Error{Status: status.OutcomeIncomplete, Code: "QUERY_CANCELLED", Path: "context", Message: "evaluation was cancelled"}
	default:
		return nil
	}
}

func (state *algebraEvaluationState) remaining() int {
	if state == nil || state.limits.MaxWorkUnits <= state.usage.WorkUnits {
		return 0
	}
	return state.limits.MaxWorkUnits - state.usage.WorkUnits
}

func (state *algebraEvaluationState) charge(amount int, path string) error {
	if err := state.checkContext(); err != nil {
		return err
	}
	if amount < 0 || amount > state.remaining() {
		state.result.Status = status.OutcomeResourceLimit
		state.result.StopReason = AlgebraStopMaxWorkUnits
		state.result.Complete = false
		return planResource(path, "evaluation exceeds the remaining logical work budget")
	}
	state.usage.WorkUnits += amount
	state.result.Usage = state.usage
	return state.checkContext()
}

func validateEvaluationSchema(schema algebra.Document) error {
	if schema.SchemaVersion != algebra.SchemaVersion || schema.Expression != nil || len(schema.Parameters) != 0 || len(schema.References) != 0 || schema.Closed {
		return planInvalid("INVALID_FACT_SCHEMA", "facts.schema", "fact schema must be a declaration-only algebra document")
	}
	if err := algebra.CheckDeclarations(schema); err != nil {
		return err
	}
	for i, signature := range schema.Signatures {
		if len(signature.Ports) > algebra.DefaultMaxPorts {
			return planResource(fmt.Sprintf("facts.schema.signatures[%d].ports", i), "relation arity exceeds the shared port limit")
		}
	}
	return nil
}

func mergeEvaluationReferences(schema algebra.Document, sourceRefs, viewRefs []algebra.Declaration, requestContext *algebra.RequestContext, limits AlgebraEvaluationLimits) (map[string]algebra.Declaration, error) {
	if len(sourceRefs) > facts.DefaultNaryMaxReferents || len(viewRefs) > facts.DefaultNaryMaxReferents {
		return nil, planResource("references", "source or fact referent count exceeds the shared ceiling")
	}
	known := make(map[string]algebra.Declaration, len(sourceRefs)+len(viewRefs))
	insert := func(declaration algebra.Declaration, path string) error {
		if !validPlanText(declaration.ID, 256) {
			return planInvalid("INVALID_KNOWN_REFERENCE", path+".id", "reference ID must be bounded valid UTF-8")
		}
		if prior, exists := known[declaration.ID]; exists {
			if !algebra.TypeEqual(prior.Type, declaration.Type) {
				return planInvalid("REFERENCE_TYPE_CONFLICT", path, "the same explicit reference ID has conflicting types")
			}
			for _, field := range []struct {
				name string
				old  *string
				next string
			}{{"temporal_frame", &prior.TemporalFrame, declaration.TemporalFrame}, {"reference_time_id", &prior.ReferenceTimeID, declaration.ReferenceTimeID}, {"multiplicity", &prior.Multiplicity, declaration.Multiplicity}} {
				if *field.old != "" && field.next != "" && *field.old != field.next {
					return planInvalid("REFERENCE_ATTRIBUTE_CONFLICT", path+"."+field.name, "the same explicit reference ID has conflicting known attributes")
				}
				if *field.old == "" {
					*field.old = field.next
				}
			}
			known[declaration.ID] = prior
			return nil
		}
		copyDeclaration := declaration
		copyDeclaration.Type = typeCopy(declaration.Type)
		known[declaration.ID] = copyDeclaration
		return nil
	}
	for i, declaration := range sourceRefs {
		if err := insert(declaration, fmt.Sprintf("plan.references[%d]", i)); err != nil {
			return nil, err
		}
	}
	for i, declaration := range viewRefs {
		if err := insert(declaration, fmt.Sprintf("facts.referents[%d]", i)); err != nil {
			return nil, err
		}
	}
	if requestContext != nil {
		for i, referent := range requestContext.Referents {
			declaration := algebra.Declaration{ID: referent.ID, Type: referent.Type, TemporalFrame: referent.TemporalFrame, ReferenceTimeID: referent.ReferenceTimeID, Multiplicity: referent.Multiplicity}
			if err := insert(declaration, fmt.Sprintf("request_context.referents[%d]", i)); err != nil {
				return nil, err
			}
		}
	}
	if len(known) > facts.DefaultNaryMaxReferents || len(known) > limits.TermLimits.MaxSymbols || len(known)+len(schema.Sorts)+len(schema.Signatures) > limits.TermLimits.MaxSymbols {
		return nil, planResource("references", "combined typed references exceed the shared symbol limits")
	}
	declarations := schema
	declarations.References = make([]algebra.Declaration, 0, len(known))
	ids := make([]string, 0, len(known))
	for id := range known {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		declarations.References = append(declarations.References, known[id])
	}
	if err := algebra.CheckDeclarations(declarations); err != nil {
		return nil, err
	}
	return known, nil
}

func validateEvaluationPlan(plan *AlgebraQueryPlan, schema algebra.Document, known map[string]algebra.Declaration, limits AlgebraEvaluationLimits) error {
	if len(plan.nodes) == 0 || len(plan.nodes) > limits.TermLimits.MaxNodes || len(plan.variables) > limits.MaxVariables || len(plan.answers) > limits.MaxVariables {
		return planResource("plan", "compiled plan exceeds the active structural limits")
	}
	if len(plan.references)+len(plan.variables)+len(schema.Sorts)+len(schema.Signatures) > limits.TermLimits.MaxSymbols {
		return planResource("plan", "query declarations exceed the active symbol limit")
	}
	variables := make(map[AlgebraVariableKey]AlgebraPlanVariable, len(plan.variables))
	for i, variable := range plan.variables {
		if !validPlanText(variable.Key.ID, 256) || variable.Key.Kind == "" {
			return planInvalid("INVALID_PLAN_VARIABLE", fmt.Sprintf("plan.variables[%d]", i), "compiled variable key is malformed")
		}
		if _, exists := variables[variable.Key]; exists {
			return planInvalid("DUPLICATE_PLAN_VARIABLE", fmt.Sprintf("plan.variables[%d]", i), "compiled variable key is duplicated")
		}
		if err := checkEvaluationType(schema, variable.Type, fmt.Sprintf("plan.variables[%d].type", i)); err != nil {
			return err
		}
		variables[variable.Key] = variable
	}
	nodes := make(map[string]AlgebraPlanNode, len(plan.nodes))
	atomCount := 0
	for i, node := range plan.nodes {
		if !validPlanText(node.ID, 256) || !validPlanText(node.Path, 4096) {
			return planInvalid("INVALID_PLAN_NODE", fmt.Sprintf("plan.nodes[%d]", i), "plan node ID and source path must be bounded")
		}
		if _, exists := nodes[node.ID]; exists {
			return planInvalid("DUPLICATE_PLAN_NODE", fmt.Sprintf("plan.nodes[%d].id", i), "plan node ID is duplicated")
		}
		nodes[node.ID] = node
		switch node.Kind {
		case AlgebraPlanAtom:
			atomCount++
			signature, ok := evaluationSignature(schema, node.Symbol)
			if !ok {
				return planInvalid("UNKNOWN_PLAN_RELATION", node.Path+".symbol", "plan relation is not declared in the selected fact schema")
			}
			if len(node.Args) != len(signature.Ports) || len(node.Args) > limits.TermLimits.MaxPorts {
				return planInvalid("PLAN_ARITY_MISMATCH", node.Path+".args", "plan atom does not match the selected signature arity")
			}
			for j, argument := range node.Args {
				port := signature.Ports[j]
				if argument.Role != port.Role {
					return planInvalid("PLAN_ROLE_MISMATCH", fmt.Sprintf("%s.args[%d].role", node.Path, j), "plan argument role does not match the declared port order")
				}
				var valueType algebra.Type
				if argument.Value.Kind == string(algebra.TermRef) {
					if argument.RefID == "" || argument.RefID != argument.Value.ID {
						return planInvalid("INVALID_PLAN_REFERENCE", fmt.Sprintf("%s.args[%d]", node.Path, j), "reference argument must retain its exact reference ID")
					}
					declaration, exists := known[argument.RefID]
					if !exists {
						return planInvalid("UNKNOWN_PLAN_REFERENCE", fmt.Sprintf("%s.args[%d]", node.Path, j), "reference is absent from the explicit request/fact identity union")
					}
					valueType = declaration.Type
				} else {
					if argument.RefID != "" {
						return planInvalid("INVALID_PLAN_REFERENCE", fmt.Sprintf("%s.args[%d]", node.Path, j), "only Ref arguments may carry ref_id")
					}
					variable, exists := variables[argument.Value]
					if !exists {
						return planInvalid("UNKNOWN_PLAN_VARIABLE", fmt.Sprintf("%s.args[%d]", node.Path, j), "argument variable has no checked declaration")
					}
					valueType = variable.Type
				}
				if !algebra.TypeEqual(valueType, port.Type) {
					return planInvalid("PLAN_ARGUMENT_TYPE_MISMATCH", fmt.Sprintf("%s.args[%d]", node.Path, j), "argument type does not match the actual selected relation port")
				}
			}
			if len(node.Children) != 0 || node.Binder != nil || node.Type != nil {
				return planInvalid("INVALID_PLAN_NODE_SHAPE", node.Path, "atom plan node contains fields from another constructor")
			}
		case AlgebraPlanAnd:
			if len(node.Children) == 0 || node.Symbol != "" || len(node.Args) != 0 || node.Binder != nil || node.Type != nil {
				return planInvalid("INVALID_PLAN_NODE_SHAPE", node.Path, "And plan node requires children only")
			}
		case AlgebraPlanExists:
			if node.Binder == nil || node.Type == nil || len(node.Children) != 1 || node.Symbol != "" || len(node.Args) != 0 {
				return planInvalid("INVALID_PLAN_NODE_SHAPE", node.Path, "Exists plan node requires one child, a binder, and its type")
			}
			variable, exists := variables[*node.Binder]
			if !exists || !algebra.TypeEqual(variable.Type, *node.Type) {
				return planInvalid("PLAN_BINDER_TYPE_MISMATCH", node.Path, "Exists binder must match its checked variable declaration")
			}
		default:
			return planUnsupported("UNSUPPORTED_PLAN_NODE", node.Path, "plan contains a constructor outside Atom/And/Exists")
		}
	}
	if atomCount > limits.MaxPlanAtoms {
		return planResource("plan", "plan atom count exceeds the active limit")
	}
	if _, exists := nodes[plan.root]; !exists {
		return planInvalid("MISSING_PLAN_ROOT", "plan.root", "compiled plan root does not exist")
	}
	seen := make(map[string]bool, len(nodes))
	active := make(map[string]bool, len(nodes))
	var visit func(string, int) error
	visit = func(id string, depth int) error {
		if depth > limits.TermLimits.MaxDepth {
			return planResource("plan", "plan tree exceeds the active depth limit")
		}
		if active[id] {
			return planInvalid("PLAN_TREE_CYCLE", "plan.nodes", "compiled plan tree contains a cycle")
		}
		if seen[id] {
			return planInvalid("PLAN_NODE_REUSED", "plan.nodes", "compiled plan tree reuses an occurrence node")
		}
		node, ok := nodes[id]
		if !ok {
			return planInvalid("MISSING_PLAN_CHILD", "plan.nodes", "compiled plan child does not exist")
		}
		active[id], seen[id] = true, true
		for _, child := range node.Children {
			if err := visit(child, depth+1); err != nil {
				return err
			}
		}
		active[id] = false
		return nil
	}
	if err := visit(plan.root, 1); err != nil {
		return err
	}
	if len(seen) != len(nodes) {
		return planInvalid("UNREACHABLE_PLAN_NODE", "plan.nodes", "compiled plan contains an unreachable node")
	}
	answerLabels := make(map[string]bool, len(plan.answers))
	for i, answer := range plan.answers {
		if !validPlanText(answer.As, 256) || answerLabels[answer.As] {
			return planInvalid("INVALID_PLAN_ANSWER", fmt.Sprintf("plan.answers[%d].as", i), "answer labels must be unique and bounded")
		}
		answerLabels[answer.As] = true
		variable, exists := variables[answer.Variable]
		if !exists || !algebra.TypeEqual(variable.Type, answer.Type) {
			return planInvalid("INVALID_PLAN_ANSWER", fmt.Sprintf("plan.answers[%d]", i), "answer must identify an exact selected variable and type")
		}
	}
	return nil
}

func checkEvaluationType(schema algebra.Document, typ algebra.Type, path string) error {
	document := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: append([]algebra.SortDeclaration{}, schema.Sorts...), Parameters: []algebra.Declaration{{ID: "evaluation_type", Type: typeCopy(typ)}}}
	if err := algebra.CheckDeclarations(document); err != nil {
		var algebraErr *algebra.Error
		if errors.As(err, &algebraErr) {
			copy := *algebraErr
			copy.Path = path
			return &copy
		}
		return err
	}
	return nil
}

func evaluationSignature(schema algebra.Document, id string) (algebra.Signature, bool) {
	for _, signature := range schema.Signatures {
		if signature.ID == id {
			return signature, true
		}
	}
	return algebra.Signature{}, false
}

func samePlanInstant(planAt, viewAt string) bool {
	if planAt == "" || viewAt == "" {
		return planAt == viewAt
	}
	if !strings.HasSuffix(planAt, "Z") || !strings.HasSuffix(viewAt, "Z") {
		return false
	}
	left, leftErr := time.Parse(time.RFC3339Nano, planAt)
	right, rightErr := time.Parse(time.RFC3339Nano, viewAt)
	return leftErr == nil && rightErr == nil && left.Equal(right)
}

func summaryMaterializationWork(summary facts.NarySummary, schema algebra.Document, known map[string]algebra.Declaration, plan *AlgebraQueryPlan) (int, int) {
	args := 0
	work := summary.StoredFacts + summary.ActiveFacts + len(summary.Atoms) + len(schema.Sorts) + len(schema.Signatures) + len(known) + len(plan.nodes) + len(plan.variables) + len(plan.references)
	for _, atom := range summary.Atoms {
		args += len(atom.Atom.Args)
	}
	return work + args, args
}

func (state *algebraEvaluationState) buildIndexes(summary facts.NarySummary) error {
	ordered := make([]facts.NaryAtomEvidence, 0, len(summary.Atoms))
	for _, evidence := range summary.Atoms {
		if len(evidence.Evidence.Positive) == 0 && len(evidence.Evidence.Negative) == 0 {
			continue
		}
		ordered = append(ordered, evidence)
	}
	sort.Slice(ordered, func(i, j int) bool { return evaluationAtomKey(ordered[i].Atom) < evaluationAtomKey(ordered[j].Atom) })
	state.rows = make([]algebraEvidenceRow, 0, len(ordered))
	for _, evidence := range ordered {
		if err := state.checkContext(); err != nil {
			return err
		}
		row := algebraEvidenceRow{
			key: evaluationAtomKey(evidence.Atom), atom: cloneEvaluationAtom(evidence.Atom),
			positive: cloneEvaluationFacts(evidence.Evidence.Positive), negative: cloneEvaluationFacts(evidence.Evidence.Negative),
		}
		sort.Slice(row.positive, func(i, j int) bool { return row.positive[i].ID < row.positive[j].ID })
		sort.Slice(row.negative, func(i, j int) bool { return row.negative[i].ID < row.negative[j].ID })
		index := len(state.rows)
		state.rows = append(state.rows, row)
		state.byAtom[row.key] = index
		state.bySymbol[row.atom.Symbol] = append(state.bySymbol[row.atom.Symbol], index)
		for _, argument := range row.atom.Args {
			if argument.Value.Kind == algebra.TermRef {
				key := algebraRoleReferenceKey{Symbol: row.atom.Symbol, Role: argument.Role, RefID: argument.Value.ID}
				state.byRoleRef[key] = append(state.byRoleRef[key], index)
			}
		}
	}
	return nil
}

func (state *algebraEvaluationState) evidenceFingerprint() (string, error) {
	refs := make([]algebra.Declaration, 0, len(state.knownRefs))
	ids := make([]string, 0, len(state.knownRefs))
	for id := range state.knownRefs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		declaration := state.knownRefs[id]
		declaration.Type = typeCopy(declaration.Type)
		refs = append(refs, declaration)
	}
	active := make([]facts.NaryAtomEvidence, 0, len(state.rows))
	for _, row := range state.rows {
		active = append(active, facts.NaryAtomEvidence{Atom: cloneEvaluationAtom(row.atom), Evidence: facts.NaryEvidence{
			State: evidenceState(row.positive, row.negative), Positive: cloneEvaluationFacts(row.positive), Negative: cloneEvaluationFacts(row.negative),
		}})
	}
	payload, err := json.Marshal(struct {
		Version          string                   `json:"version"`
		FactsSnapshotID  string                   `json:"facts_snapshot_id"`
		FormalSnapshotID string                   `json:"formal_snapshot_id"`
		ContextID        string                   `json:"context_id"`
		At               string                   `json:"at"`
		References       []algebra.Declaration    `json:"references"`
		Active           []facts.NaryAtomEvidence `json:"active_evidence"`
	}{"algebra-active-evidence-v1", state.plan.factsSnapshotID, state.plan.formalSnapshotID, state.plan.contextID, state.plan.at, refs, active})
	if err != nil {
		return "", planInvalid("EVIDENCE_FINGERPRINT_ENCODING", "facts", "active evidence fingerprint could not be encoded")
	}
	if len(payload) > state.limits.MaxOutputBytes || len(payload) > MaxAlgebraOutputBytes {
		return "", planResource("facts.fingerprint", "active evidence fingerprint exceeds the bounded payload size")
	}
	if err := state.charge(len(payload), "facts.fingerprint"); err != nil {
		return "", err
	}
	hash := sha256.Sum256(payload)
	return "algebra-active-evidence-v1:" + hex.EncodeToString(hash[:]), nil
}

func (state *algebraEvaluationState) setClosedGroundAtomEvidence() {
	if state == nil || state.plan == nil || len(state.plan.nodes) != 1 || len(state.plan.variables) != 0 || len(state.plan.answers) != 0 {
		return
	}
	node := state.plan.nodes[0]
	if node.ID != state.plan.root || node.Kind != AlgebraPlanAtom {
		return
	}
	atom := facts.NaryAtom{Symbol: node.Symbol, Args: make([]facts.NaryArgument, 0, len(node.Args))}
	for _, argument := range node.Args {
		if argument.Value.Kind != string(algebra.TermRef) || argument.RefID == "" || argument.RefID != argument.Value.ID {
			return
		}
		atom.Args = append(atom.Args, facts.NaryArgument{Role: argument.Role, Value: facts.NaryValue{Kind: algebra.TermRef, ID: argument.RefID}})
	}
	evidence := facts.NaryEvidence{State: facts.NaryUnknown, Positive: []facts.NaryFact{}, Negative: []facts.NaryFact{}}
	if index, exists := state.byAtom[evaluationAtomKey(atom)]; exists {
		row := state.rows[index]
		evidence.Positive = cloneEvaluationFacts(row.positive)
		evidence.Negative = cloneEvaluationFacts(row.negative)
		evidence.State = evidenceState(evidence.Positive, evidence.Negative)
	}
	state.result.GroundAtomEvidence = &evidence
}

func (state *algebraEvaluationState) positiveRecordCount() int {
	total := 0
	for _, row := range state.rows {
		total += len(row.positive)
	}
	return total
}

func (state *algebraEvaluationState) negativeRecordCount() int {
	total := 0
	for _, row := range state.rows {
		total += len(row.negative)
	}
	return total
}

func evaluationAtomKey(atom facts.NaryAtom) string {
	var builder strings.Builder
	writeEvaluationKeyPart(&builder, atom.Symbol)
	for _, argument := range atom.Args {
		writeEvaluationKeyPart(&builder, argument.Role)
		writeEvaluationKeyPart(&builder, string(argument.Value.Kind))
		writeEvaluationKeyPart(&builder, argument.Value.ID)
	}
	return builder.String()
}

func writeEvaluationKeyPart(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
}

func evidenceState(positive, negative []facts.NaryFact) facts.NaryKnowledge {
	switch {
	case len(positive) > 0 && len(negative) > 0:
		return facts.NaryBoth
	case len(positive) > 0:
		return facts.NaryPositive
	case len(negative) > 0:
		return facts.NaryNegative
	default:
		return facts.NaryUnknown
	}
}

func cloneEvaluationAtom(atom facts.NaryAtom) facts.NaryAtom {
	return facts.NaryAtom{Symbol: atom.Symbol, Args: append([]facts.NaryArgument{}, atom.Args...)}
}

func cloneEvaluationFacts(in []facts.NaryFact) []facts.NaryFact {
	out := make([]facts.NaryFact, len(in))
	for i, fact := range in {
		out[i] = fact
		out[i].Atom = cloneEvaluationAtom(fact.Atom)
		if fact.Source.Reference != nil {
			reference := *fact.Source.Reference
			out[i].Source.Reference = &reference
		}
	}
	return out
}

func cloneEvaluationConflicts(in []facts.NaryConflict) []facts.NaryConflict {
	out := make([]facts.NaryConflict, len(in))
	for i, conflict := range in {
		out[i] = conflict
		out[i].Atom = cloneEvaluationAtom(conflict.Atom)
		out[i].PositiveIDs = append([]string{}, conflict.PositiveIDs...)
		out[i].NegativeIDs = append([]string{}, conflict.NegativeIDs...)
	}
	sort.Slice(out, func(i, j int) bool { return evaluationAtomKey(out[i].Atom) < evaluationAtomKey(out[j].Atom) })
	return out
}

func factsErrorAsAlgebra(err error) error {
	if err == nil {
		return nil
	}
	var algebraErr *algebra.Error
	if errors.As(err, &algebraErr) {
		return algebraErr
	}
	var factErr *facts.StatusError
	if errors.As(err, &factErr) {
		result := factErr.Result()
		if len(result.Diagnostics) > 0 {
			diagnostic := result.Diagnostics[0]
			return &algebra.Error{Status: result.Status, Code: diagnostic.Code, Path: diagnostic.Path, Message: diagnostic.Message}
		}
	}
	return planInvalid("INVALID_FACT_VIEW", "facts", err.Error())
}

// algebraWitness is one complete positive grounding for every compiled plan
// variable. atomRows maps each source-occurrence node to one grouped evidence
// row; it deliberately retains existential bindings for later proof assembly.
type algebraWitness struct {
	bindings map[AlgebraVariableKey]facts.NaryValue
	atomRows map[string]int
}

type algebraBindingAddition struct {
	key   AlgebraVariableKey
	value facts.NaryValue
}

// enumerateAlgebraWitnesses performs deterministic left-deep positive joins
// over the already validated active fact index. The callback receives only
// complete witness tuples; stopping it early never reports exhaustion.
func enumerateAlgebraWitnesses(state *algebraEvaluationState, consume func(algebraWitness) (bool, error)) (bool, error) {
	if state == nil || state.plan == nil || state.ctx == nil || consume == nil {
		return false, planInvalid("INVALID_EVALUATION_STATE", "evaluation", "initialized evaluation state and internal consumer are required")
	}
	if err := state.checkContext(); err != nil {
		return false, err
	}
	atoms, err := collectEvaluationAtoms(state)
	if err != nil {
		return false, err
	}
	state.sourceAtoms = atoms
	if err := state.charge(len(state.plan.variables)+1, "plan.variables"); err != nil {
		return false, err
	}
	if state.planVariables == nil {
		state.planVariables = make(map[AlgebraVariableKey]AlgebraPlanVariable, len(state.plan.variables))
		for _, variable := range state.plan.variables {
			state.planVariables[variable.Key] = variable
		}
	}
	unused := make([]AlgebraPlanVariable, 0, len(state.plan.variables))
	for _, variable := range state.plan.variables {
		if !variable.UsedInAtom {
			unused = append(unused, variable)
		}
	}
	if len(unused) > state.limits.MaxVariables {
		return false, planResource("plan.variables", "unused variable count exceeds the active limit")
	}
	base := algebraWitness{bindings: make(map[AlgebraVariableKey]facts.NaryValue), atomRows: make(map[string]int)}
	states := []algebraWitness{base}
	for atomIndex, node := range atoms {
		if err := state.checkContext(); err != nil {
			return false, err
		}
		stream := atomIndex == len(atoms)-1 && len(unused) == 0
		next := make([]algebraWitness, 0)
		for stateIndex := range states {
			if err := state.charge(1+len(node.Args), node.Path+".index"); err != nil {
				return false, err
			}
			candidateRows, err := state.candidateRows(node, states[stateIndex].bindings)
			if err != nil {
				return false, err
			}
			for _, rowIndex := range candidateRows {
				if err := state.checkContext(); err != nil {
					return false, err
				}
				if err := state.charge(1+len(node.Args), node.Path+".match"); err != nil {
					return false, err
				}
				state.usage.JoinAttempts++
				state.result.Usage = state.usage
				row := state.rows[rowIndex]
				additions, additionCount, matched, err := state.matchAtom(node, row, states[stateIndex].bindings)
				if err != nil {
					return false, err
				}
				if !matched {
					continue
				}
				copyWork := 1 + len(states[stateIndex].bindings) + len(states[stateIndex].atomRows) + additionCount
				if err := state.charge(copyWork, node.Path+".witness"); err != nil {
					return false, err
				}
				witness := cloneAlgebraWitness(states[stateIndex])
				for i := 0; i < additionCount; i++ {
					witness.bindings[additions[i].key] = additions[i].value
				}
				witness.atomRows[node.ID] = rowIndex
				if stream {
					complete, err := state.finishAlgebraWitness(witness, len(atoms), consume)
					if err != nil {
						return false, err
					}
					if !complete {
						return false, nil
					}
					continue
				}
				if len(next) >= state.limits.MaxIntermediateRows {
					return false, state.stopIntermediateRows(node.Path)
				}
				next = append(next, witness)
				state.noteIntermediateRows(len(next))
			}
		}
		if !stream {
			states = next
			if len(states) == 0 {
				return true, nil
			}
		}
	}
	if len(unused) > 0 {
		if err := state.charge(len(state.knownRefs), "references.domain_index"); err != nil {
			return false, err
		}
		knownIDs := state.sortedKnownReferenceIDs()
		for variableIndex, variable := range unused {
			if err := state.checkContext(); err != nil {
				return false, err
			}
			next := make([]algebraWitness, 0)
			lastStage := variableIndex == len(unused)-1
			for stateIndex := range states {
				for _, id := range knownIDs {
					if err := state.checkContext(); err != nil {
						return false, err
					}
					if err := state.charge(1, "references.domain"); err != nil {
						return false, err
					}
					declaration := state.knownRefs[id]
					if !algebra.TypeEqual(variable.Type, declaration.Type) {
						continue
					}
					if existing, exists := states[stateIndex].bindings[variable.Key]; exists {
						_ = existing
						return false, planInvalid("DUPLICATE_DOMAIN_BINDING", variable.SourcePath, "an unused query variable was already bound before its domain stage")
					}
					if err := state.charge(1+len(states[stateIndex].bindings)+len(states[stateIndex].atomRows), "references.witness"); err != nil {
						return false, err
					}
					witness := cloneAlgebraWitness(states[stateIndex])
					witness.bindings[variable.Key] = facts.NaryValue{Kind: algebra.TermRef, ID: id}
					if lastStage {
						complete, err := state.finishAlgebraWitness(witness, len(atoms), consume)
						if err != nil {
							return false, err
						}
						if !complete {
							return false, nil
						}
						continue
					}
					if len(next) >= state.limits.MaxIntermediateRows {
						return false, state.stopIntermediateRows("references.domain")
					}
					next = append(next, witness)
					state.noteIntermediateRows(len(next))
				}
			}
			if !lastStage {
				states = next
				if len(states) == 0 {
					return true, nil
				}
			}
		}
	}
	if len(atoms) == 0 && len(unused) == 0 {
		continued, err := state.finishAlgebraWitness(base, 0, consume)
		if err != nil {
			return false, err
		}
		if !continued {
			return false, nil
		}
	}
	return true, nil
}

func collectEvaluationAtoms(state *algebraEvaluationState) ([]AlgebraPlanNode, error) {
	plan := state.plan
	if err := state.charge(len(plan.nodes), "plan.atom_scan"); err != nil {
		return nil, err
	}
	nodes := make(map[string]AlgebraPlanNode, len(plan.nodes))
	for _, node := range plan.nodes {
		nodes[node.ID] = node
	}
	atoms := make([]AlgebraPlanNode, 0, countPlanAtoms(plan))
	var walk func(string, int) error
	walk = func(id string, depth int) error {
		if err := state.checkContext(); err != nil {
			return err
		}
		if err := state.charge(1, "plan.atom_scan"); err != nil {
			return err
		}
		if depth > len(plan.nodes) {
			return planInvalid("PLAN_TREE_CYCLE", "plan.nodes", "compiled plan traversal encountered a cycle")
		}
		node, ok := nodes[id]
		if !ok {
			return planInvalid("MISSING_PLAN_CHILD", "plan.nodes", "compiled plan child does not exist")
		}
		switch node.Kind {
		case AlgebraPlanAtom:
			atoms = append(atoms, node)
		case AlgebraPlanAnd, AlgebraPlanExists:
			for _, child := range node.Children {
				if err := walk(child, depth+1); err != nil {
					return err
				}
			}
		default:
			return planUnsupported("UNSUPPORTED_PLAN_NODE", node.Path, "positive matcher supports only checked Atom/And/Exists nodes")
		}
		return nil
	}
	if err := walk(plan.root, 1); err != nil {
		return nil, err
	}
	return atoms, nil
}

func countPlanAtoms(plan *AlgebraQueryPlan) int {
	count := 0
	for _, node := range plan.nodes {
		if node.Kind == AlgebraPlanAtom {
			count++
		}
	}
	return count
}

func (state *algebraEvaluationState) candidateRows(node AlgebraPlanNode, bound map[AlgebraVariableKey]facts.NaryValue) ([]int, error) {
	rows, exists := state.bySymbol[node.Symbol]
	if !exists {
		return []int{}, nil
	}
	best := rows
	for _, argument := range node.Args {
		var refID string
		switch argument.Value.Kind {
		case string(algebra.TermRef):
			refID = argument.RefID
		default:
			if value, bound := bound[argument.Value]; bound {
				if value.Kind != algebra.TermRef {
					return []int{}, nil
				}
				refID = value.ID
			}
		}
		if refID != "" {
			indexed, ok := state.byRoleRef[algebraRoleReferenceKey{Symbol: node.Symbol, Role: argument.Role, RefID: refID}]
			if !ok {
				return []int{}, nil
			}
			if len(indexed) < len(best) {
				best = indexed
			}
		}
	}
	return best, nil
}

func (state *algebraEvaluationState) matchAtom(node AlgebraPlanNode, row algebraEvidenceRow, bound map[AlgebraVariableKey]facts.NaryValue) ([algebra.DefaultMaxPorts]algebraBindingAddition, int, bool, error) {
	if len(row.positive) == 0 || row.atom.Symbol != node.Symbol || len(row.atom.Args) != len(node.Args) {
		return [algebra.DefaultMaxPorts]algebraBindingAddition{}, 0, false, nil
	}
	var additions [algebra.DefaultMaxPorts]algebraBindingAddition
	additionCount := 0
	for i, argument := range node.Args {
		factArgument := row.atom.Args[i]
		if factArgument.Role != argument.Role || factArgument.Value.Kind != algebra.TermRef {
			return additions, additionCount, false, nil
		}
		var expectedType algebra.Type
		if argument.Value.Kind == string(algebra.TermRef) {
			if argument.RefID == "" || argument.RefID != argument.Value.ID {
				return additions, additionCount, false, planInvalid("INVALID_PLAN_REFERENCE", fmt.Sprintf("%s.args[%d]", node.Path, i), "reference argument must retain its exact ID")
			}
			declaration, ok := state.knownRefs[argument.RefID]
			if !ok {
				return additions, additionCount, false, planInvalid("UNKNOWN_PLAN_REFERENCE", fmt.Sprintf("%s.args[%d]", node.Path, i), "reference is absent from the explicit known identity domain")
			}
			expectedType = declaration.Type
			if factArgument.Value.ID != argument.RefID {
				return additions, additionCount, false, nil
			}
		} else {
			key := argument.Value
			variable, ok := state.planVariables[key]
			if !ok {
				return additions, additionCount, false, planInvalid("UNKNOWN_PLAN_VARIABLE", fmt.Sprintf("%s.args[%d]", node.Path, i), "variable key is absent from the checked plan")
			}
			expectedType = variable.Type
			if prior, exists := bound[key]; exists {
				if prior.Kind != algebra.TermRef || prior.ID != factArgument.Value.ID {
					return additions, additionCount, false, nil
				}
			} else if index := findAlgebraAddition(additions, additionCount, key); index >= 0 {
				if additions[index].value.Kind != algebra.TermRef || additions[index].value.ID != factArgument.Value.ID {
					return additions, additionCount, false, nil
				}
			} else {
				declaration, known := state.knownRefs[factArgument.Value.ID]
				if !known || !algebra.TypeEqual(declaration.Type, expectedType) {
					return additions, additionCount, false, nil
				}
				if additionCount >= len(additions) {
					return additions, additionCount, false, planResource(node.Path+".args", "atom variable bindings exceed the shared port limit")
				}
				additions[additionCount] = algebraBindingAddition{key: key, value: factArgument.Value}
				additionCount++
			}
		}
		port, ok := evaluationPort(state.schema, node.Symbol, argument.Role)
		if !ok || !algebra.TypeEqual(expectedType, port.Type) {
			return additions, additionCount, false, planInvalid("PLAN_ARGUMENT_TYPE_MISMATCH", fmt.Sprintf("%s.args[%d]", node.Path, i), "plan value type does not match the actual role type")
		}
		declaration, known := state.knownRefs[factArgument.Value.ID]
		if !known || !algebra.TypeEqual(declaration.Type, port.Type) {
			return additions, additionCount, false, planInvalid("FACT_REFERENCE_TYPE_MISMATCH", fmt.Sprintf("%s.args[%d]", node.Path, i), "fact reference does not match the actual selected role type")
		}
	}
	return additions, additionCount, true, nil
}

func findAlgebraAddition(additions [algebra.DefaultMaxPorts]algebraBindingAddition, count int, key AlgebraVariableKey) int {
	for i := 0; i < count; i++ {
		if additions[i].key == key {
			return i
		}
	}
	return -1
}

func evaluationPort(schema algebra.Document, symbol, role string) (algebra.Port, bool) {
	signature, ok := evaluationSignature(schema, symbol)
	if !ok {
		return algebra.Port{}, false
	}
	for _, port := range signature.Ports {
		if port.Role == role {
			return port, true
		}
	}
	return algebra.Port{}, false
}

func cloneAlgebraWitness(source algebraWitness) algebraWitness {
	copyOf := algebraWitness{bindings: make(map[AlgebraVariableKey]facts.NaryValue, len(source.bindings)), atomRows: make(map[string]int, len(source.atomRows))}
	for key, value := range source.bindings {
		copyOf.bindings[key] = value
	}
	for id, row := range source.atomRows {
		copyOf.atomRows[id] = row
	}
	return copyOf
}

func (state *algebraEvaluationState) finishAlgebraWitness(witness algebraWitness, expectedAtoms int, consume func(algebraWitness) (bool, error)) (bool, error) {
	if err := state.checkContext(); err != nil {
		return false, err
	}
	if err := state.charge(len(state.plan.variables)+len(witness.atomRows), "witness.validate"); err != nil {
		return false, err
	}
	if len(witness.atomRows) != expectedAtoms || len(witness.bindings) != len(state.plan.variables) {
		return false, planInvalid("INCOMPLETE_WITNESS", "plan", "completed witness must bind every plan variable and source atom occurrence")
	}
	for _, variable := range state.plan.variables {
		value, exists := witness.bindings[variable.Key]
		if !exists || value.Kind != algebra.TermRef {
			return false, planInvalid("INCOMPLETE_WITNESS", variable.SourcePath, "completed witness must retain an exact grounded reference")
		}
		declaration, known := state.knownRefs[value.ID]
		if !known || !algebra.TypeEqual(declaration.Type, variable.Type) {
			return false, planInvalid("WITNESS_TYPE_MISMATCH", variable.SourcePath, "grounded reference is not declared at the exact variable type")
		}
	}
	return consume(witness)
}

func (state *algebraEvaluationState) sortedKnownReferenceIDs() []string {
	ids := make([]string, 0, len(state.knownRefs))
	for id := range state.knownRefs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (state *algebraEvaluationState) noteIntermediateRows(count int) {
	if count > state.usage.IntermediateRows {
		state.usage.IntermediateRows = count
	}
	state.result.Usage = state.usage
}

func (state *algebraEvaluationState) stopIntermediateRows(path string) error {
	state.result.Status = status.OutcomeResourceLimit
	state.result.StopReason = AlgebraStopMaxIntermediateRows
	state.result.Complete = false
	return planResource("limits.max_intermediate_rows", "positive join stage exceeds the configured intermediate row limit at "+path)
}

// EvaluateAlgebraQuery enumerates finite positive witnesses over one explicit
// request-scoped fact view and retains a locally reassembled proof DAG.
func EvaluateAlgebraQuery(parent context.Context, plan *AlgebraQueryPlan, view *facts.NaryView, requested AlgebraEvaluationLimits) (AlgebraEvaluationResult, error) {
	state, err := prepareAlgebraEvaluation(parent, plan, view, requested)
	if err != nil {
		return AlgebraEvaluationResult{}, err
	}
	defer state.close()
	state.result.Assumptions = []string{
		"Only explicit positive records in the selected fact snapshot support answers.",
		"Evaluation exhausts the supplied finite known-reference domain under the selected evidence snapshot; absent evidence remains unknown and domain completeness is not claimed.",
	}
	answerIndex := make(map[string]int)
	seenGroundings := make(map[string]struct{})
	exhausted, matchErr := enumerateAlgebraWitnesses(state, func(witness algebraWitness) (bool, error) {
		if err := state.checkContext(); err != nil {
			return false, err
		}
		groundingKey, err := state.witnessGroundingKey(witness)
		if err != nil {
			return false, err
		}
		if _, duplicate := seenGroundings[groundingKey]; duplicate {
			return true, nil
		}
		values, tupleKey, err := state.answerTuple(witness)
		if err != nil {
			return false, err
		}
		answerPosition, hasAnswer := answerIndex[tupleKey]
		if !hasAnswer && len(state.result.Answers) >= state.limits.MaxResults {
			state.setResourceStop(AlgebraStopMaxResults, "limits.max_results", "answer count exceeds the configured result limit")
			state.recordStopError(planResource("limits.max_results", "answer count exceeds the configured result limit"))
			return false, nil
		}
		if hasAnswer && len(state.result.Answers[answerPosition].ProofIDs) >= state.limits.MaxProofsPerAnswer {
			state.setResourceStop(AlgebraStopMaxProofsPerAnswer, "limits.max_proofs_per_answer", "additional positive derivation exceeds the per-answer proof limit")
			state.recordStopError(planResource("limits.max_proofs_per_answer", "additional positive derivation exceeds the per-answer proof limit"))
			return false, nil
		}
		if err := state.charge(len(witness.bindings)+len(witness.atomRows)+len(values), "witness.materialize"); err != nil {
			if state.recordStopError(err) {
				return false, nil
			}
			return false, err
		}
		proofNodes, projectID, conflict, err := state.buildWitnessProof(witness)
		if err != nil {
			if state.recordStopError(err) {
				return false, nil
			}
			return false, err
		}
		answer := AlgebraAnswer{Values: append([]facts.NaryValue{}, values...), ProofIDs: []string{projectID}, HasConflict: conflict}
		if hasAnswer {
			prior := state.result.Answers[answerPosition]
			answer = AlgebraAnswer{ID: prior.ID, Values: append([]facts.NaryValue{}, prior.Values...), ProofIDs: append(append([]string{}, prior.ProofIDs...), projectID), HasConflict: prior.HasConflict || conflict}
		} else {
			answer.ID = fmt.Sprintf("a%d", len(state.result.Answers)+1)
		}
		committed, err := state.commitAlgebraCandidate(answer, answerPosition, hasAnswer, proofNodes)
		if err != nil {
			if state.recordStopError(err) {
				return false, nil
			}
			return false, err
		}
		if !committed {
			return false, nil
		}
		seenGroundings[groundingKey] = struct{}{}
		if !hasAnswer {
			answerIndex[tupleKey] = len(state.result.Answers) - 1
		}
		return true, nil
	})
	if matchErr != nil {
		if !state.recordStopError(matchErr) {
			return AlgebraEvaluationResult{}, matchErr
		}
	}
	if matchErr == nil && exhausted && state.result.StopReason == "" {
		state.result.Status = status.OutcomeOK
		state.result.Complete = true
		state.result.StopReason = ""
		if len(state.result.Answers) > 0 {
			state.result.Knowledge = AlgebraKnowledgeSupported
		} else {
			state.result.Knowledge = AlgebraKnowledgeUnknown
		}
	} else {
		state.result.Complete = false
		if len(state.result.Answers) > 0 {
			state.result.Knowledge = AlgebraKnowledgeSupported
		} else {
			state.result.Knowledge = AlgebraKnowledgeUnknown
		}
		if state.result.Status == status.OutcomeOK {
			state.result.Status = status.OutcomeIncomplete
		}
	}
	state.syncRetainedUsage()
	if err := state.finalizeAlgebraOutput(); err != nil {
		if !state.recordStopError(err) {
			return AlgebraEvaluationResult{}, err
		}
	}
	state.syncRetainedUsage()
	return state.result, nil
}

func (state *algebraEvaluationState) witnessGroundingKey(witness algebraWitness) (string, error) {
	if err := state.charge(len(witness.bindings)+len(witness.atomRows), "witness.identity"); err != nil {
		return "", err
	}
	keys := make([]AlgebraVariableKey, 0, len(witness.bindings))
	for key := range witness.bindings {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Kind != keys[j].Kind {
			return keys[i].Kind < keys[j].Kind
		}
		return keys[i].ID < keys[j].ID
	})
	var builder strings.Builder
	writeEvaluationKeyPart(&builder, "grounding-v1")
	for _, key := range keys {
		value := witness.bindings[key]
		writeEvaluationKeyPart(&builder, key.Kind)
		writeEvaluationKeyPart(&builder, key.ID)
		writeEvaluationKeyPart(&builder, string(value.Kind))
		writeEvaluationKeyPart(&builder, value.ID)
	}
	for _, node := range state.sourceAtoms {
		rowIndex, exists := witness.atomRows[node.ID]
		if !exists || rowIndex < 0 || rowIndex >= len(state.rows) {
			return "", planInvalid("MISSING_ATOM_WITNESS", node.Path, "grounding identity requires every source atom occurrence")
		}
		writeEvaluationKeyPart(&builder, node.ID)
		writeEvaluationKeyPart(&builder, state.rows[rowIndex].key)
	}
	return builder.String(), nil
}

func (state *algebraEvaluationState) answerTuple(witness algebraWitness) ([]facts.NaryValue, string, error) {
	if err := state.charge(len(state.plan.answers), "answer.tuple"); err != nil {
		return nil, "", err
	}
	values := make([]facts.NaryValue, 0, len(state.plan.answers))
	var key strings.Builder
	writeEvaluationKeyPart(&key, "answer-tuple-v1")
	for _, answer := range state.plan.answers {
		value, exists := witness.bindings[answer.Variable]
		if !exists || value.Kind != algebra.TermRef {
			return nil, "", planInvalid("MISSING_ANSWER_BINDING", "select."+answer.As, "selected output is not grounded to an exact reference")
		}
		declaration, known := state.knownRefs[value.ID]
		if !known || !algebra.TypeEqual(declaration.Type, answer.Type) {
			return nil, "", planInvalid("ANSWER_TYPE_MISMATCH", "select."+answer.As, "selected reference does not match its requested output type")
		}
		values = append(values, value)
		writeEvaluationKeyPart(&key, value.ID)
	}
	return values, key.String(), nil
}

func (state *algebraEvaluationState) commitAlgebraCandidate(answer AlgebraAnswer, answerPosition int, hasAnswer bool, stagedProof []AlgebraProofNode) (bool, error) {
	if err := state.checkContext(); err != nil {
		return false, err
	}
	proof := append(append([]AlgebraProofNode{}, state.result.Proof...), stagedProof...)
	answers := append([]AlgebraAnswer{}, state.result.Answers...)
	if hasAnswer {
		if answerPosition < 0 || answerPosition >= len(answers) {
			return false, planInvalid("ANSWER_INDEX_MISMATCH", "answers", "answer index changed during candidate staging")
		}
		answers[answerPosition] = answer
	} else {
		answers = append(answers, answer)
	}
	candidate := state.result
	candidate.Answers, candidate.Proof = answers, proof
	candidate.Knowledge = AlgebraKnowledgeSupported
	candidate.Complete = false
	candidate.Status = status.OutcomeIncomplete
	candidate.StopReason = ""
	candidate.Usage = state.usage
	candidate.Usage.Results = len(answers)
	candidate.Usage.ProofNodes = len(proof)
	candidate.Usage.Proofs = totalAlgebraProofs(answers)
	candidate.Usage.OutputBytes = 0
	if err := state.checkContext(); err != nil {
		return false, err
	}
	preview, err := json.Marshal(candidate)
	if err != nil {
		return false, planInvalid("RESULT_ENCODING", "result", "staged answer result could not be encoded")
	}
	if len(preview) > state.limits.MaxOutputBytes || len(preview) > MaxAlgebraOutputBytes {
		state.setResourceStop(AlgebraStopMaxOutputBytes, "limits.max_output_bytes", "retained result exceeds the configured serialized output limit")
		return false, planResource("limits.max_output_bytes", "retained result exceeds the configured serialized output limit")
	}
	if err := state.charge(len(preview), "result.output"); err != nil {
		return false, err
	}
	candidate.Usage = state.usage
	candidate.Usage.Results = len(answers)
	candidate.Usage.ProofNodes = len(proof)
	candidate.Usage.Proofs = totalAlgebraProofs(answers)
	candidate.Usage.OutputBytes = len(preview)
	payload, err := marshalAlgebraEvaluationResult(candidate, state.limits.MaxOutputBytes)
	if err != nil {
		state.setResourceStop(AlgebraStopMaxOutputBytes, "limits.max_output_bytes", "retained result exceeds the configured serialized output limit")
		return false, err
	}
	candidate.Usage.OutputBytes = len(payload)
	state.usage = candidate.Usage
	candidate.Usage = state.usage
	state.result = candidate
	return true, nil
}

func totalAlgebraProofs(answers []AlgebraAnswer) int {
	total := 0
	for _, answer := range answers {
		total += len(answer.ProofIDs)
	}
	return total
}

func marshalAlgebraEvaluationResult(result AlgebraEvaluationResult, maximum int) ([]byte, error) {
	for attempt := 0; attempt < 4; attempt++ {
		payload, err := json.Marshal(result)
		if err != nil {
			return nil, planInvalid("RESULT_ENCODING", "result", "evaluation result could not be encoded")
		}
		if len(payload) > maximum || len(payload) > MaxAlgebraOutputBytes {
			return nil, planResource("limits.max_output_bytes", "serialized evaluation result exceeds the configured output limit")
		}
		if result.Usage.OutputBytes == len(payload) {
			return payload, nil
		}
		result.Usage.OutputBytes = len(payload)
	}
	return nil, planResource("limits.max_output_bytes", "serialized output size did not stabilize within the fixed encoding bound")
}

func (state *algebraEvaluationState) syncRetainedUsage() {
	state.usage.Results = len(state.result.Answers)
	state.usage.ProofNodes = len(state.result.Proof)
	state.usage.Proofs = totalAlgebraProofs(state.result.Answers)
	state.result.Usage = state.usage
}

func (state *algebraEvaluationState) finalizeAlgebraOutput() error {
	if err := state.checkContext(); err != nil {
		return err
	}
	state.syncRetainedUsage()
	state.result.Usage.OutputBytes = 0
	payload, err := marshalAlgebraEvaluationResult(state.result, state.limits.MaxOutputBytes)
	if err != nil {
		state.setResourceStop(AlgebraStopMaxOutputBytes, "limits.max_output_bytes", "final evaluation result exceeds the configured serialized output limit")
		return err
	}
	if err := state.charge(len(payload), "result.output.final"); err != nil {
		return err
	}
	state.result.Usage = state.usage
	state.result.Usage.OutputBytes = len(payload)
	state.usage.OutputBytes = len(payload)
	final, err := marshalAlgebraEvaluationResult(state.result, state.limits.MaxOutputBytes)
	if err != nil {
		state.setResourceStop(AlgebraStopMaxOutputBytes, "limits.max_output_bytes", "final evaluation result exceeds the configured serialized output limit")
		return err
	}
	state.result.Usage.OutputBytes = len(final)
	state.usage.OutputBytes = len(final)
	return nil
}

func (state *algebraEvaluationState) setResourceStop(reason AlgebraEvaluationStopReason, path, message string) {
	state.result.Status = status.OutcomeResourceLimit
	state.result.StopReason = reason
	state.result.Complete = false
	if len(state.result.Answers) > 0 {
		state.result.Knowledge = AlgebraKnowledgeSupported
	} else {
		state.result.Knowledge = AlgebraKnowledgeUnknown
	}
	state.result.Usage = state.usage
	state.recordStopError(planResource(path, message))
}

func (state *algebraEvaluationState) recordStopError(err error) bool {
	if err == nil {
		return false
	}
	var algebraErr *algebra.Error
	if !errors.As(err, &algebraErr) {
		return false
	}
	switch algebraErr.Status {
	case status.OutcomeResourceLimit:
		state.result.Status = status.OutcomeResourceLimit
		state.result.Complete = false
		if state.result.StopReason == "" {
			switch {
			case strings.Contains(algebraErr.Path, "max_intermediate_rows"):
				state.result.StopReason = AlgebraStopMaxIntermediateRows
			case strings.Contains(algebraErr.Path, "max_results"):
				state.result.StopReason = AlgebraStopMaxResults
			case strings.Contains(algebraErr.Path, "max_proofs_per_answer"):
				state.result.StopReason = AlgebraStopMaxProofsPerAnswer
			case strings.Contains(algebraErr.Path, "max_proof_nodes"):
				state.result.StopReason = AlgebraStopMaxProofNodes
			case strings.Contains(algebraErr.Path, "max_proof_depth"):
				state.result.StopReason = AlgebraStopMaxProofDepth
			case strings.Contains(algebraErr.Path, "max_output_bytes"):
				state.result.StopReason = AlgebraStopMaxOutputBytes
			case strings.Contains(algebraErr.Path, "timeout"):
				state.result.StopReason = AlgebraStopTimeout
			default:
				state.result.StopReason = AlgebraStopMaxWorkUnits
			}
		}
	case status.OutcomeIncomplete:
		state.result.Status = status.OutcomeIncomplete
		state.result.Complete = false
		state.result.StopReason = AlgebraStopCancelled
	default:
		return false
	}
	for _, diagnostic := range state.result.Diagnostics {
		if diagnostic.Code == algebraErr.Code && diagnostic.Path == algebraErr.Path && diagnostic.Message == algebraErr.Message {
			return true
		}
	}
	state.result.Diagnostics = append(state.result.Diagnostics, *algebraErr)
	return true
}

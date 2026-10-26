package reason

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/status"
)

// AlgebraQueryKnowledge reports whether explicit positive support is present.
// It deliberately has no false or compound-negation state.
type AlgebraQueryKnowledge string

const (
	AlgebraKnowledgeSupported AlgebraQueryKnowledge = "supported"
	AlgebraKnowledgeUnknown   AlgebraQueryKnowledge = "unknown"
)

// Intrinsic query proof rules record how supplied positive evidence supports
// an answer. They are generic query rules, not domain axioms.
const (
	AlgebraProofFactMatch = "query.fact_match.v1"
	AlgebraProofAnd       = "query.and.v1"
	AlgebraProofExists    = "query.exists.v1"
	AlgebraProofProject   = "query.project.v1"
	AlgebraProofReference = "query.reference.v1"
)

// AlgebraBinding is one explicit, namespace-aware assignment to a source
// query variable. Ground values are exact request-local Ref identifiers.
type AlgebraBinding struct {
	Variable AlgebraVariableKey `json:"variable"`
	Value    facts.NaryValue    `json:"value"`
}

// AlgebraAnswer is one ordered selected tuple with every retained positive
// derivation ID. HasConflict reports opposing explicit evidence separately.
type AlgebraAnswer struct {
	ID          string            `json:"id"`
	Values      []facts.NaryValue `json:"values"`
	ProofIDs    []string          `json:"proof_ids"`
	HasConflict bool              `json:"has_conflict"`
}

// AlgebraProofNode is a proposed checkable node in a positive support DAG.
// This DTO does not establish its own validity or completeness.
type AlgebraProofNode struct {
	ID                      string                 `json:"id"`
	RuleID                  string                 `json:"rule_id"`
	PlanNodeID              string                 `json:"plan_node_id"`
	Premises                []string               `json:"premises"`
	FactIDs                 []string               `json:"fact_ids"`
	Sources                 []facts.SourceEvidence `json:"sources"`
	Atom                    *facts.NaryAtom        `json:"atom,omitempty"`
	Bindings                []AlgebraBinding       `json:"bindings"`
	Witness                 *AlgebraBinding        `json:"witness,omitempty"`
	Reference               *algebra.Declaration   `json:"reference,omitempty"`
	OpposingNegativeFactIDs []string               `json:"opposing_negative_fact_ids"`
	Depth                   int                    `json:"depth"`
}

// AlgebraInputEvidenceSummary reports the supplied active/inactive records
// and input conflicts. These are evidence counts, not truth conclusions.
type AlgebraInputEvidenceSummary struct {
	StoredFacts     int                  `json:"stored_facts"`
	ActiveFacts     int                  `json:"active_facts"`
	InactiveFacts   int                  `json:"inactive_facts"`
	PositiveRecords int                  `json:"positive_records"`
	NegativeRecords int                  `json:"negative_records"`
	Conflicts       []facts.NaryConflict `json:"conflicts"`
}

// AlgebraEvaluationResult contains a bounded result over the explicit query
// and fact snapshots. Complete scopes only the finite known-reference scan;
// it never claims that the request identities are a complete world domain.
type AlgebraEvaluationResult struct {
	SchemaVersion       int                         `json:"schema_version"`
	Status              status.OutcomeKind          `json:"status"`
	Profile             string                      `json:"profile"`
	Knowledge           AlgebraQueryKnowledge       `json:"knowledge"`
	ContextID           string                      `json:"context_id"`
	At                  string                      `json:"at,omitempty"`
	FactsSnapshotID     string                      `json:"facts_snapshot_id"`
	FormalSnapshotID    string                      `json:"formal_snapshot_id"`
	EvidenceFingerprint string                      `json:"evidence_fingerprint"`
	PlanFingerprint     string                      `json:"plan_fingerprint"`
	Columns             []AlgebraAnswerPort         `json:"columns"`
	Answers             []AlgebraAnswer             `json:"answers"`
	Proof               []AlgebraProofNode          `json:"proof"`
	InputEvidence       AlgebraInputEvidenceSummary `json:"input_evidence"`
	GroundAtomEvidence  *facts.NaryEvidence         `json:"ground_atom_evidence,omitempty"`
	Complete            bool                        `json:"complete"`
	StopReason          AlgebraEvaluationStopReason `json:"stop_reason,omitempty"`
	Diagnostics         []algebra.Error             `json:"diagnostics"`
	Usage               AlgebraEvaluationUsage      `json:"usage"`
	Limits              AlgebraEvaluationLimits     `json:"limits"`
	Assumptions         []string                    `json:"assumptions"`
	Plan                *AlgebraPlanDescription     `json:"plan,omitempty"`
}

// newAlgebraEvaluationResult creates an explicitly incomplete result with all
// required arrays initialized. It makes no claim that a query was evaluated.
func newAlgebraEvaluationResult() AlgebraEvaluationResult {
	return AlgebraEvaluationResult{
		SchemaVersion: AlgebraQuerySchemaVersion,
		Status:        status.OutcomeIncomplete,
		Profile:       AlgebraQueryProfile,
		Knowledge:     AlgebraKnowledgeUnknown,
		Columns:       []AlgebraAnswerPort{},
		Answers:       []AlgebraAnswer{},
		Proof:         []AlgebraProofNode{},
		InputEvidence: AlgebraInputEvidenceSummary{Conflicts: []facts.NaryConflict{}},
		Diagnostics:   []algebra.Error{},
		Limits:        DefaultAlgebraEvaluationLimits(),
		Assumptions:   []string{},
	}
}

// newAlgebraProofNode keeps mandatory list fields non-null when a producer
// starts a node. Semantic verification is performed by a later checker.
func newAlgebraProofNode() AlgebraProofNode {
	return AlgebraProofNode{
		Premises:                []string{},
		FactIDs:                 []string{},
		Sources:                 []facts.SourceEvidence{},
		Bindings:                []AlgebraBinding{},
		OpposingNegativeFactIDs: []string{},
	}
}

type algebraProofBuild struct {
	nodes []AlgebraProofNode
}

type algebraProofNodeResult struct {
	id       string
	depth    int
	bindings []AlgebraBinding
}

func (state *algebraEvaluationState) buildWitnessProof(witness algebraWitness) ([]AlgebraProofNode, string, bool, error) {
	builder := &algebraProofBuild{nodes: make([]AlgebraProofNode, 0)}
	planNodes := make(map[string]AlgebraPlanNode, len(state.plan.nodes))
	for _, node := range state.plan.nodes {
		planNodes[node.ID] = node
	}
	variables := state.planVariables
	var build func(string, int) (algebraProofNodeResult, error)
	build = func(id string, depth int) (algebraProofNodeResult, error) {
		if err := state.checkContext(); err != nil {
			return algebraProofNodeResult{}, err
		}
		if depth > state.limits.TermLimits.MaxDepth {
			return algebraProofNodeResult{}, planResource("proof", "source proof traversal exceeds the checked term depth")
		}
		if err := state.charge(1, "proof.tree"); err != nil {
			return algebraProofNodeResult{}, err
		}
		node, ok := planNodes[id]
		if !ok {
			return algebraProofNodeResult{}, planInvalid("MISSING_PLAN_CHILD", "plan.nodes", "proof source node does not exist")
		}
		switch node.Kind {
		case AlgebraPlanAtom:
			rowIndex, exists := witness.atomRows[node.ID]
			if !exists || rowIndex < 0 || rowIndex >= len(state.rows) {
				return algebraProofNodeResult{}, planInvalid("MISSING_ATOM_WITNESS", node.Path, "complete grounding has no matching evidence row for this atom occurrence")
			}
			row := state.rows[rowIndex]
			if len(row.positive) == 0 {
				return algebraProofNodeResult{}, planInvalid("NONPOSITIVE_ATOM_WITNESS", node.Path, "fact_match proof requires explicit positive evidence")
			}
			bindings, err := proofAtomBindings(node, witness, variables)
			if err != nil {
				return algebraProofNodeResult{}, err
			}
			factIDs := make([]string, 0, len(row.positive))
			sources := make([]facts.SourceEvidence, 0, len(row.positive))
			for _, fact := range row.positive {
				if err := state.charge(1, node.Path+".proof.facts"); err != nil {
					return algebraProofNodeResult{}, err
				}
				factIDs = append(factIDs, fact.ID)
				source := fact.Source
				if source.Reference != nil {
					reference := *source.Reference
					source.Reference = &reference
				}
				sources = append(sources, source)
			}
			opposing := make([]string, 0, len(row.negative))
			for _, fact := range row.negative {
				opposing = append(opposing, fact.ID)
			}
			sort.Strings(opposing)
			proofNode := newAlgebraProofNode()
			proofNode.RuleID, proofNode.PlanNodeID = AlgebraProofFactMatch, node.ID
			proofNode.FactIDs, proofNode.Sources = factIDs, sources
			atom := cloneEvaluationAtom(row.atom)
			proofNode.Atom, proofNode.Bindings = &atom, bindings
			proofNode.OpposingNegativeFactIDs = opposing
			proofNode.Depth = 1
			proofID, err := builder.add(state, proofNode)
			if err != nil {
				return algebraProofNodeResult{}, err
			}
			return algebraProofNodeResult{id: proofID, depth: 1, bindings: cloneAlgebraBindings(bindings)}, nil
		case AlgebraPlanAnd:
			proofNode := newAlgebraProofNode()
			proofNode.RuleID, proofNode.PlanNodeID = AlgebraProofAnd, node.ID
			outward := make([]AlgebraBinding, 0)
			maxDepth := 0
			for _, child := range node.Children {
				childResult, err := build(child, depth+1)
				if err != nil {
					return algebraProofNodeResult{}, err
				}
				proofNode.Premises = append(proofNode.Premises, childResult.id)
				if childResult.depth > maxDepth {
					maxDepth = childResult.depth
				}
				outward, err = mergeProofBindings(outward, childResult.bindings, node.Path)
				if err != nil {
					return algebraProofNodeResult{}, err
				}
			}
			proofNode.Bindings = cloneAlgebraBindings(outward)
			proofNode.Depth = maxDepth + 1
			proofID, err := builder.add(state, proofNode)
			if err != nil {
				return algebraProofNodeResult{}, err
			}
			return algebraProofNodeResult{id: proofID, depth: proofNode.Depth, bindings: outward}, nil
		case AlgebraPlanExists:
			if node.Binder == nil {
				return algebraProofNodeResult{}, planInvalid("MISSING_EXISTS_BINDER", node.Path, "checked Exists node has no binder")
			}
			childResult, err := build(node.Children[0], depth+1)
			if err != nil {
				return algebraProofNodeResult{}, err
			}
			value, exists := witness.bindings[*node.Binder]
			if !exists || value.Kind != algebra.TermRef {
				return algebraProofNodeResult{}, planInvalid("MISSING_EXISTS_WITNESS", node.Path, "Exists node requires its exact scoped reference witness")
			}
			variable, exists := variables[*node.Binder]
			if !exists {
				return algebraProofNodeResult{}, planInvalid("UNKNOWN_EXISTS_BINDER", node.Path, "Exists binder is absent from the checked variable inventory")
			}
			outward, found := removeProofBinding(childResult.bindings, *node.Binder)
			var premises []string
			premises = append(premises, childResult.id)
			if !variable.UsedInAtom {
				refID, refDepth, err := builder.addKnownReference(state, *node.Binder, variable, value)
				if err != nil {
					return algebraProofNodeResult{}, err
				}
				premises = append(premises, refID)
				if refDepth > childResult.depth {
					childResult.depth = refDepth
				}
			} else if !found {
				return algebraProofNodeResult{}, planInvalid("MISSING_EXISTS_BINDING", node.Path, "used Exists binder is absent from the child proof bindings")
			}
			witnessBinding := AlgebraBinding{Variable: *node.Binder, Value: value}
			proofNode := newAlgebraProofNode()
			proofNode.RuleID, proofNode.PlanNodeID = AlgebraProofExists, node.ID
			proofNode.Premises = premises
			proofNode.Witness = &witnessBinding
			proofNode.Bindings = cloneAlgebraBindings(outward)
			proofNode.Depth = childResult.depth + 1
			proofID, err := builder.add(state, proofNode)
			if err != nil {
				return algebraProofNodeResult{}, err
			}
			return algebraProofNodeResult{id: proofID, depth: proofNode.Depth, bindings: outward}, nil
		default:
			return algebraProofNodeResult{}, planUnsupported("UNSUPPORTED_PLAN_NODE", node.Path, "proof reconstruction supports only checked Atom/And/Exists nodes")
		}
	}

	root, err := build(state.plan.root, 1)
	if err != nil {
		return nil, "", false, err
	}
	project := newAlgebraProofNode()
	project.RuleID, project.PlanNodeID = AlgebraProofProject, state.plan.root
	project.Premises = append(project.Premises, root.id)
	maxDepth := root.depth
	answerBindings := make([]AlgebraBinding, 0, len(state.plan.answers))
	for _, answer := range state.plan.answers {
		value, exists := witness.bindings[answer.Variable]
		if !exists || value.Kind != algebra.TermRef {
			return nil, "", false, planInvalid("MISSING_ANSWER_BINDING", "select."+answer.As, "selected output is not grounded to an explicit reference")
		}
		declaration, known := state.knownRefs[value.ID]
		if !known || !algebra.TypeEqual(declaration.Type, answer.Type) {
			return nil, "", false, planInvalid("ANSWER_TYPE_MISMATCH", "select."+answer.As, "selected output reference does not match its declared type")
		}
		answerBindings = append(answerBindings, AlgebraBinding{Variable: answer.Variable, Value: value})
		if !state.planVariables[answer.Variable].UsedInAtom {
			variable := state.planVariables[answer.Variable]
			refID, refDepth, err := builder.addKnownReference(state, answer.Variable, variable, value)
			if err != nil {
				return nil, "", false, err
			}
			project.Premises = append(project.Premises, refID)
			if refDepth > maxDepth {
				maxDepth = refDepth
			}
		}
	}
	project.Bindings = answerBindings
	project.Depth = maxDepth + 1
	projectID, err := builder.add(state, project)
	if err != nil {
		return nil, "", false, err
	}
	conflict := false
	for _, rowIndex := range witness.atomRows {
		if rowIndex >= 0 && rowIndex < len(state.rows) && len(state.rows[rowIndex].negative) > 0 {
			conflict = true
		}
	}
	return builder.nodes, projectID, conflict, nil
}

func (builder *algebraProofBuild) add(state *algebraEvaluationState, node AlgebraProofNode) (string, error) {
	if node.Depth < 1 || node.Depth > state.limits.MaxProofDepth {
		state.setResourceStop(AlgebraStopMaxProofDepth, "limits.max_proof_depth", "proof DAG exceeds the configured depth limit")
		return "", planResource("limits.max_proof_depth", "proof DAG exceeds the configured depth limit")
	}
	if len(state.result.Proof)+len(builder.nodes) >= state.limits.MaxProofNodes {
		state.setResourceStop(AlgebraStopMaxProofNodes, "limits.max_proof_nodes", "proof DAG exceeds the configured node limit")
		return "", planResource("limits.max_proof_nodes", "proof DAG exceeds the configured node limit")
	}
	if err := state.charge(1+len(node.Premises)+len(node.Bindings)+len(node.FactIDs)+len(node.Sources)+len(node.OpposingNegativeFactIDs), "proof.node"); err != nil {
		return "", err
	}
	node.ID = fmt.Sprintf("p%d", len(state.result.Proof)+len(builder.nodes)+1)
	builder.nodes = append(builder.nodes, node)
	return node.ID, nil
}

func (builder *algebraProofBuild) addKnownReference(state *algebraEvaluationState, key AlgebraVariableKey, variable AlgebraPlanVariable, value facts.NaryValue) (string, int, error) {
	if err := state.checkContext(); err != nil {
		return "", 0, err
	}
	declaration, known := state.knownRefs[value.ID]
	if !known || value.Kind != algebra.TermRef || !algebra.TypeEqual(declaration.Type, variable.Type) {
		return "", 0, planInvalid("INVALID_REFERENCE_WITNESS", variable.SourcePath, "known-reference witness must preserve its exact declared ID and type")
	}
	proofNode := newAlgebraProofNode()
	proofNode.RuleID, proofNode.PlanNodeID = AlgebraProofReference, variable.SourcePath
	proofNode.Reference = &algebra.Declaration{ID: declaration.ID, Type: copyReasonType(declaration.Type)}
	proofNode.Bindings = []AlgebraBinding{{Variable: key, Value: value}}
	proofNode.Depth = 1
	proofID, err := builder.add(state, proofNode)
	return proofID, 1, err
}

func proofAtomBindings(node AlgebraPlanNode, witness algebraWitness, variables map[AlgebraVariableKey]AlgebraPlanVariable) ([]AlgebraBinding, error) {
	bindings := make([]AlgebraBinding, 0, len(node.Args))
	seen := make(map[AlgebraVariableKey]struct{}, len(node.Args))
	for _, argument := range node.Args {
		if argument.Value.Kind == string(algebra.TermRef) {
			continue
		}
		if _, exists := seen[argument.Value]; exists {
			continue
		}
		variable, known := variables[argument.Value]
		value, grounded := witness.bindings[argument.Value]
		if !known || !grounded || value.Kind != algebra.TermRef {
			return nil, planInvalid("INVALID_ATOM_BINDING", node.Path, "atom proof binding must identify one exact checked variable and known reference")
		}
		_ = variable
		seen[argument.Value] = struct{}{}
		bindings = append(bindings, AlgebraBinding{Variable: argument.Value, Value: value})
	}
	return bindings, nil
}

func mergeProofBindings(left, right []AlgebraBinding, path string) ([]AlgebraBinding, error) {
	out := cloneAlgebraBindings(left)
	positions := make(map[AlgebraVariableKey]int, len(out)+len(right))
	for i, binding := range out {
		positions[binding.Variable] = i
	}
	for _, binding := range right {
		if index, exists := positions[binding.Variable]; exists {
			prior := out[index].Value
			if prior.Kind != binding.Value.Kind || prior.ID != binding.Value.ID {
				return nil, planInvalid("PROOF_BINDING_CONFLICT", path, "conjunct proof children disagree about an exact variable binding")
			}
			continue
		}
		positions[binding.Variable] = len(out)
		out = append(out, binding)
	}
	return out, nil
}

func removeProofBinding(bindings []AlgebraBinding, key AlgebraVariableKey) ([]AlgebraBinding, bool) {
	out := make([]AlgebraBinding, 0, len(bindings))
	found := false
	for _, binding := range bindings {
		if binding.Variable == key {
			found = true
			continue
		}
		out = append(out, binding)
	}
	return out, found
}

func cloneAlgebraBindings(in []AlgebraBinding) []AlgebraBinding {
	out := make([]AlgebraBinding, len(in))
	copy(out, in)
	return out
}

func copyReasonType(value algebra.Type) algebra.Type {
	copyOf := algebra.Type{Kind: value.Kind, SortID: value.SortID}
	if value.From != nil {
		from := copyReasonType(*value.From)
		copyOf.From = &from
	}
	if value.To != nil {
		to := copyReasonType(*value.To)
		copyOf.To = &to
	}
	return copyOf
}

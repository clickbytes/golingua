package reason

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const AlgebraEvaluationProofScope = "supplied_positive_answer_witnesses"

type AlgebraEvaluationProofDisposition string

const (
	AlgebraProofVerifiedSupport AlgebraEvaluationProofDisposition = "positive_support_verified"
	AlgebraProofUndetermined    AlgebraEvaluationProofDisposition = "undetermined"
	AlgebraProofRejected        AlgebraEvaluationProofDisposition = "rejected"
)

// AlgebraEvaluationProofUsage reports work performed by this verifier. It is
// independent of untrusted producer usage and is not a CPU or energy metric.
type AlgebraEvaluationProofUsage struct {
	WorkUnits       int `json:"work_units"`
	VerifiedAnswers int `json:"verified_answers"`
	VerifiedNodes   int `json:"verified_nodes"`
	EvidenceRows    int `json:"evidence_rows"`
}

// AlgebraEvaluationProofVerification establishes only positive support for
// each supplied answer. Completeness is deliberately never inferred from one
// result document or from the finite known-reference domain.
type AlgebraEvaluationProofVerification struct {
	SchemaVersion           int                               `json:"schema_version"`
	Status                  status.OutcomeKind                `json:"status"`
	Disposition             AlgebraEvaluationProofDisposition `json:"disposition"`
	ProofValid              bool                              `json:"proof_valid"`
	PositiveSupportVerified bool                              `json:"positive_support_verified"`
	CompletenessVerified    bool                              `json:"completeness_verified"`
	Scope                   string                            `json:"scope"`
	Profile                 string                            `json:"profile"`
	ContextID               string                            `json:"context_id"`
	At                      string                            `json:"at,omitempty"`
	FactsSnapshotID         string                            `json:"facts_snapshot_id"`
	FormalSnapshotID        string                            `json:"formal_snapshot_id"`
	PlanFingerprint         string                            `json:"plan_fingerprint"`
	EvidenceFingerprint     string                            `json:"evidence_fingerprint"`
	VerifiedAnswers         int                               `json:"verified_answers"`
	VerifiedProofNodes      int                               `json:"verified_proof_nodes"`
	Assumptions             []string                          `json:"assumptions"`
	Diagnostics             []algebra.Error                   `json:"diagnostics"`
	Usage                   AlgebraEvaluationProofUsage       `json:"usage"`
}

type verifiedAlgebraProofNode struct {
	planNodeID string
	ruleID     string
	bindings   []AlgebraBinding
	conflict   bool
	depth      int
}

// VerifyAlgebraEvaluationProof independently replays the positive witness DAG
// against the immutable compiled plan and one actual active fact snapshot. It
// never evaluates the query, enumerates assignments, or calls the producer.
func VerifyAlgebraEvaluationProof(parent context.Context, plan *AlgebraQueryPlan, view *facts.NaryView, claim AlgebraEvaluationResult, requested AlgebraEvaluationLimits) (AlgebraEvaluationProofVerification, error) {
	result := newAlgebraEvaluationProofVerification(plan)
	limits, err := NormalizeAlgebraEvaluationLimits(requested)
	if err != nil {
		return result, err
	}
	state, err := prepareAlgebraEvaluation(parent, plan, view, limits)
	if err != nil {
		return algebraProofFailure(result, err)
	}
	defer state.close()
	result = newAlgebraEvaluationProofVerification(plan)
	result.ContextID = state.plan.contextID
	result.At = state.plan.at
	result.FactsSnapshotID = state.plan.factsSnapshotID
	result.FormalSnapshotID = state.plan.formalSnapshotID
	result.PlanFingerprint = state.plan.fingerprint
	result.EvidenceFingerprint = state.result.EvidenceFingerprint
	result.Usage.WorkUnits = state.usage.WorkUnits
	if err := state.checkContext(); err != nil {
		return algebraProofFailure(result, err)
	}
	if err := validateAlgebraEvaluationClaimShape(claim, state.limits); err != nil {
		return algebraProofFailure(result, err)
	}
	shapeCost := algebraEvaluationClaimShapeCost(claim)
	if err := state.charge(shapeCost, "proof.claim"); err != nil {
		return algebraProofFailure(result, err)
	}
	result.Usage.WorkUnits = state.usage.WorkUnits
	if err := validateAlgebraEvaluationClaimTypes(claim, state); err != nil {
		return algebraProofFailure(result, err)
	}
	if err := state.charge(len(claim.Columns)+countClaimReferenceTypes(claim), "proof.types"); err != nil {
		return algebraProofFailure(result, err)
	}
	result.Usage.WorkUnits = state.usage.WorkUnits
	if state.planVariables == nil {
		if err := state.charge(len(state.plan.variables), "proof.plan_variables"); err != nil {
			return algebraProofFailure(result, err)
		}
		state.planVariables = make(map[AlgebraVariableKey]AlgebraPlanVariable, len(state.plan.variables))
		for _, variable := range state.plan.variables {
			state.planVariables[variable.Key] = variable
		}
	}
	if err := matchAlgebraEvaluationClaimHeader(claim, state); err != nil {
		return algebraProofFailure(result, err)
	}
	if err := matchAlgebraEvaluationEvidenceSummary(claim, state); err != nil {
		return algebraProofFailure(result, err)
	}
	if err := state.checkContext(); err != nil {
		return algebraProofFailure(result, err)
	}
	if len(claim.Answers) == 0 {
		if len(claim.Proof) != 0 {
			return algebraProofFailure(result, planInvalid("UNUSED_PROOF_NODE", "proof", "an answer-free result cannot retain proof nodes"))
		}
		result.Status = status.OutcomeUnknown
		result.Disposition = AlgebraProofUndetermined
		result.Diagnostics = append(result.Diagnostics, algebra.Error{Status: status.OutcomeUnknown, Code: "NO_POSITIVE_WITNESSES", Path: "answers", Message: "no positive answer witnesses were supplied; absence does not establish a negative result"})
		result.Usage.WorkUnits = state.usage.WorkUnits
		return result, nil
	}
	verifiedNodes, roots, err := verifyAlgebraProofDAG(claim, state, &result)
	if err != nil {
		return algebraProofFailure(result, err)
	}
	if err := state.checkContext(); err != nil {
		return algebraProofFailure(result, err)
	}
	verifiedAnswers := make(map[string]struct{}, len(claim.Answers))
	answerTuples := make(map[string]struct{}, len(claim.Answers))
	usedRoots := make(map[string]struct{})
	for answerIndex, answer := range claim.Answers {
		if err := state.checkContext(); err != nil {
			return algebraProofFailure(result, err)
		}
		if err := state.charge(1+len(answer.Values)+len(answer.ProofIDs), fmt.Sprintf("answers[%d]", answerIndex)); err != nil {
			return algebraProofFailure(result, err)
		}
		if _, exists := verifiedAnswers[answer.ID]; exists {
			return algebraProofFailure(result, planInvalid("DUPLICATE_ANSWER_ID", fmt.Sprintf("answers[%d].id", answerIndex), "answer IDs must be unique"))
		}
		verifiedAnswers[answer.ID] = struct{}{}
		tuple := answerTupleIdentity(answer.Values)
		if _, exists := answerTuples[tuple]; exists {
			return algebraProofFailure(result, planInvalid("DUPLICATE_ANSWER_TUPLE", fmt.Sprintf("answers[%d].values", answerIndex), "answer tuples must be unique"))
		}
		answerTuples[tuple] = struct{}{}
		expectedValues := make([]facts.NaryValue, len(state.plan.answers))
		for i, selected := range state.plan.answers {
			if i >= len(answer.Values) {
				return algebraProofFailure(result, planInvalid("ANSWER_COLUMN_COUNT_MISMATCH", fmt.Sprintf("answers[%d].values", answerIndex), "answer tuple does not match the selected columns"))
			}
			value := answer.Values[i]
			if value.Kind != algebra.TermRef || !validPlanText(value.ID, 256) {
				return algebraProofFailure(result, planInvalid("INVALID_ANSWER_REFERENCE", fmt.Sprintf("answers[%d].values[%d]", answerIndex, i), "answer values must be exact bounded references"))
			}
			declaration, known := state.knownRefs[value.ID]
			if !known || !algebra.TypeEqual(declaration.Type, selected.Type) {
				return algebraProofFailure(result, planInvalid("ANSWER_REFERENCE_TYPE_MISMATCH", fmt.Sprintf("answers[%d].values[%d]", answerIndex, i), "answer reference is not declared at the selected column type"))
			}
			expectedValues[i] = facts.NaryValue{Kind: algebra.TermRef, ID: value.ID}
		}
		if len(answer.Values) != len(state.plan.answers) {
			return algebraProofFailure(result, planInvalid("ANSWER_COLUMN_COUNT_MISMATCH", fmt.Sprintf("answers[%d].values", answerIndex), "answer tuple does not match the selected columns"))
		}
		if len(answer.ProofIDs) == 0 {
			return algebraProofFailure(result, planInvalid("MISSING_ANSWER_PROOF", fmt.Sprintf("answers[%d].proof_ids", answerIndex), "every supplied answer requires at least one positive proof root"))
		}
		conflictAny := false
		for _, proofID := range answer.ProofIDs {
			if _, exists := usedRoots[proofID]; exists {
				return algebraProofFailure(result, planInvalid("DUPLICATE_PROOF_ROOT", fmt.Sprintf("answers[%d].proof_ids", answerIndex), "proof roots must be unique across answer records"))
			}
			usedRoots[proofID] = struct{}{}
			root, exists := roots[proofID]
			if !exists || root.ruleID != AlgebraProofProject {
				return algebraProofFailure(result, planInvalid("INVALID_ANSWER_PROOF_ROOT", fmt.Sprintf("answers[%d].proof_ids", answerIndex), "answer proof root must be a verified project node"))
			}
			if !sameAlgebraAnswerBindings(root.bindings, state.plan.answers, expectedValues) {
				return algebraProofFailure(result, planInvalid("ANSWER_PROOF_BINDING_MISMATCH", fmt.Sprintf("answers[%d].proof_ids", answerIndex), "project proof does not establish the exact ordered answer tuple"))
			}
			conflictAny = conflictAny || root.conflict
		}
		if conflictAny != answer.HasConflict {
			return algebraProofFailure(result, planInvalid("ANSWER_CONFLICT_MISMATCH", fmt.Sprintf("answers[%d].has_conflict", answerIndex), "conflict flag must reflect opposing records across the supplied derivations"))
		}
		result.VerifiedAnswers++
		result.Usage.VerifiedAnswers = result.VerifiedAnswers
		result.Usage.WorkUnits = state.usage.WorkUnits
	}
	if err := state.charge(len(claim.Proof), "proof.reachability"); err != nil {
		return algebraProofFailure(result, err)
	}
	if len(usedRoots) == 0 || !allAlgebraProofNodesReachable(claim, usedRoots, verifiedNodes) {
		return algebraProofFailure(result, planInvalid("UNREACHABLE_PROOF_NODE", "proof", "every proof node must be reachable from a supplied answer root"))
	}
	result.Status = status.OutcomeOK
	result.Disposition = AlgebraProofVerifiedSupport
	result.ProofValid = true
	result.PositiveSupportVerified = true
	result.CompletenessVerified = false
	result.VerifiedProofNodes = len(claim.Proof)
	result.Usage.VerifiedNodes = len(claim.Proof)
	result.Usage.EvidenceRows = state.result.InputEvidence.ActiveFacts
	result.Usage.WorkUnits = state.usage.WorkUnits
	return result, nil
}

func newAlgebraEvaluationProofVerification(plan *AlgebraQueryPlan) AlgebraEvaluationProofVerification {
	result := AlgebraEvaluationProofVerification{
		SchemaVersion: AlgebraQuerySchemaVersion,
		Status:        status.OutcomeIncomplete, Disposition: AlgebraProofUndetermined,
		Scope: AlgebraEvaluationProofScope, Profile: AlgebraQueryProfile,
		Assumptions: []string{
			"Only supplied positive answer witnesses are checked against the selected query and active fact snapshot.",
			"The verifier does not rerun query evaluation or search for additional answers.",
			"A finite known-reference domain and a producer Complete flag do not prove answer-set completeness.",
			"Conflicting negative records are reported as evidence conflict and do not erase positive support.",
		},
		Diagnostics: []algebra.Error{},
	}
	if plan != nil {
		result.ContextID, result.At = plan.contextID, plan.at
		result.FactsSnapshotID, result.FormalSnapshotID = plan.factsSnapshotID, plan.formalSnapshotID
		result.PlanFingerprint = plan.fingerprint
	}
	return result
}

func algebraProofFailure(result AlgebraEvaluationProofVerification, err error) (AlgebraEvaluationProofVerification, error) {
	var typed *algebra.Error
	if !errors.As(err, &typed) {
		return result, err
	}
	result.Status = typed.Status
	result.Disposition = AlgebraProofRejected
	result.ProofValid = false
	result.PositiveSupportVerified = false
	result.CompletenessVerified = false
	result.Diagnostics = append(result.Diagnostics, *typed)
	return result, nil
}

func validateAlgebraEvaluationClaimShape(claim AlgebraEvaluationResult, limits AlgebraEvaluationLimits) error {
	if claim.SchemaVersion != AlgebraQuerySchemaVersion {
		return planInvalid("INVALID_EVALUATION_PROOF_VERSION", "schema_version", "evaluation proof schema version must be 1")
	}
	if claim.Columns == nil || claim.Answers == nil || claim.Proof == nil || claim.Diagnostics == nil || claim.Assumptions == nil || claim.InputEvidence.Conflicts == nil {
		return planInvalid("NULL_EVALUATION_PROOF_ARRAY", "proof", "required evaluation-result arrays must be present and non-null")
	}
	if len(claim.Columns) > MaxAlgebraVariables || len(claim.Answers) > limits.MaxResults || len(claim.Proof) > limits.MaxProofNodes || len(claim.Diagnostics) > status.MaxDiagnostics || len(claim.Assumptions) > 32 || len(claim.InputEvidence.Conflicts) > facts.DefaultNaryMaxFacts {
		return planResource("proof", "evaluation proof arrays exceed the selected finite limits")
	}
	if !claim.Status.Valid() || (claim.Knowledge != AlgebraKnowledgeUnknown && claim.Knowledge != AlgebraKnowledgeSupported) {
		return planInvalid("INVALID_EVALUATION_PROOF_STATUS", "status", "producer status fields must use the closed evaluation vocabulary")
	}
	for i, column := range claim.Columns {
		if !validPlanText(column.As, 256) || !validPlanText(column.Variable.Kind, 32) || !validPlanText(column.Variable.ID, 256) {
			return planInvalid("INVALID_EVALUATION_PROOF_COLUMN", fmt.Sprintf("columns[%d]", i), "column identity fields are invalid or exceed their bound")
		}
	}
	for i, answer := range claim.Answers {
		if !validPlanText(answer.ID, 256) || answer.Values == nil || answer.ProofIDs == nil || len(answer.Values) > MaxAlgebraVariables || len(answer.ProofIDs) == 0 || len(answer.ProofIDs) > limits.MaxProofsPerAnswer {
			return planInvalid("INVALID_EVALUATION_PROOF_ANSWER", fmt.Sprintf("answers[%d]", i), "answer fields or required arrays exceed their bounds")
		}
		for j, value := range answer.Values {
			if value.Kind != algebra.TermRef || !validPlanText(value.ID, 256) {
				return planInvalid("INVALID_EVALUATION_PROOF_VALUE", fmt.Sprintf("answers[%d].values[%d]", i, j), "answer values must be bounded explicit references")
			}
		}
		for j, id := range answer.ProofIDs {
			if !validPlanText(id, 256) {
				return planInvalid("INVALID_EVALUATION_PROOF_ID", fmt.Sprintf("answers[%d].proof_ids[%d]", i, j), "proof root ID is invalid or too long")
			}
		}
	}
	if claim.GroundAtomEvidence != nil {
		if claim.GroundAtomEvidence.Positive == nil || claim.GroundAtomEvidence.Negative == nil || len(claim.GroundAtomEvidence.Positive) > facts.DefaultNaryMaxFacts || len(claim.GroundAtomEvidence.Negative) > facts.DefaultNaryMaxFacts {
			return planInvalid("INVALID_GROUND_ATOM_EVIDENCE", "ground_atom_evidence", "ground evidence arrays must be non-null and bounded")
		}
		for i, record := range claim.GroundAtomEvidence.Positive {
			if err := validateClaimFactRecord(record, fmt.Sprintf("ground_atom_evidence.positive[%d]", i)); err != nil {
				return err
			}
		}
		for i, record := range claim.GroundAtomEvidence.Negative {
			if err := validateClaimFactRecord(record, fmt.Sprintf("ground_atom_evidence.negative[%d]", i)); err != nil {
				return err
			}
		}
	}
	for i, conflict := range claim.InputEvidence.Conflicts {
		path := fmt.Sprintf("input_evidence.conflicts[%d]", i)
		if err := validateClaimAtom(conflict.Atom, path+".atom"); err != nil {
			return err
		}
		if conflict.PositiveIDs == nil || conflict.NegativeIDs == nil || len(conflict.PositiveIDs) > facts.DefaultNaryMaxFacts || len(conflict.NegativeIDs) > facts.DefaultNaryMaxFacts {
			return planInvalid("INVALID_EVALUATION_CONFLICT", path, "conflict fact ID arrays must be non-null and bounded")
		}
		for _, id := range append(append([]string{}, conflict.PositiveIDs...), conflict.NegativeIDs...) {
			if !validPlanText(id, 128) {
				return planInvalid("INVALID_EVALUATION_CONFLICT", path, "conflict fact IDs must be bounded UTF-8")
			}
		}
	}
	return nil
}

func algebraEvaluationClaimShapeCost(claim AlgebraEvaluationResult) int {
	work := len(claim.Columns) + len(claim.Answers) + len(claim.Proof) + len(claim.InputEvidence.Conflicts)
	for _, answer := range claim.Answers {
		work += len(answer.Values) + len(answer.ProofIDs)
	}
	for _, node := range claim.Proof {
		work += 1 + len(node.Premises) + len(node.FactIDs) + len(node.Sources) + len(node.Bindings) + len(node.OpposingNegativeFactIDs)
		if node.Atom != nil {
			work += len(node.Atom.Args)
		}
	}
	return work
}

func validateAlgebraEvaluationClaimTypes(claim AlgebraEvaluationResult, state *algebraEvaluationState) error {
	for i, column := range claim.Columns {
		if err := state.charge(1+state.limits.TermLimits.MaxTypeDepth, fmt.Sprintf("columns[%d].type", i)); err != nil {
			return err
		}
		document := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: state.schema.Sorts, Parameters: []algebra.Declaration{{ID: "proof_column", Type: column.Type}}, Limits: state.limits.TermLimits}
		if err := algebra.CheckDeclarations(document); err != nil {
			return err
		}
	}
	for i, node := range claim.Proof {
		if node.Reference == nil {
			continue
		}
		if err := state.charge(1+state.limits.TermLimits.MaxTypeDepth, fmt.Sprintf("proof[%d].reference.type", i)); err != nil {
			return err
		}
		document := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: state.schema.Sorts, Parameters: []algebra.Declaration{{ID: "proof_reference", Type: node.Reference.Type}}, Limits: state.limits.TermLimits}
		if err := algebra.CheckDeclarations(document); err != nil {
			return err
		}
	}
	return nil
}

func countClaimReferenceTypes(claim AlgebraEvaluationResult) int {
	count := 0
	for _, node := range claim.Proof {
		if node.Reference != nil {
			count++
		}
	}
	return count
}

func matchAlgebraEvaluationClaimHeader(claim AlgebraEvaluationResult, state *algebraEvaluationState) error {
	plan := state.plan
	if claim.Profile != plan.profile || claim.ContextID != plan.contextID || !samePlanInstant(claim.At, plan.at) || claim.FactsSnapshotID != plan.factsSnapshotID || claim.FormalSnapshotID != plan.formalSnapshotID {
		return planInvalid("EVALUATION_PROOF_CONTEXT_MISMATCH", "claim", "profile, context, instant, and both snapshots must match the actual plan and fact view")
	}
	if claim.PlanFingerprint != plan.fingerprint || claim.EvidenceFingerprint != state.result.EvidenceFingerprint {
		return planInvalid("EVALUATION_PROOF_FINGERPRINT_MISMATCH", "claim", "plan and active evidence fingerprints must match the actual selected inputs")
	}
	if len(claim.Columns) != len(plan.answers) {
		return planInvalid("EVALUATION_PROOF_COLUMNS_MISMATCH", "columns", "column list must exactly match the compiled answer projection")
	}
	for i, actual := range plan.answers {
		claimed := claim.Columns[i]
		if claimed.As != actual.As || claimed.Variable != actual.Variable || !algebra.TypeEqual(claimed.Type, actual.Type) {
			return planInvalid("EVALUATION_PROOF_COLUMNS_MISMATCH", fmt.Sprintf("columns[%d]", i), "column binding, role, order, or type differs from the compiled query")
		}
	}
	return nil
}

func matchAlgebraEvaluationEvidenceSummary(claim AlgebraEvaluationResult, state *algebraEvaluationState) error {
	actual := state.result.InputEvidence
	claimed := claim.InputEvidence
	if claimed.StoredFacts != actual.StoredFacts || claimed.ActiveFacts != actual.ActiveFacts || claimed.InactiveFacts != actual.InactiveFacts || claimed.PositiveRecords != actual.PositiveRecords || claimed.NegativeRecords != actual.NegativeRecords || !sameNaryConflicts(claimed.Conflicts, actual.Conflicts) {
		return planInvalid("EVALUATION_PROOF_EVIDENCE_SUMMARY_MISMATCH", "input_evidence", "evidence counts and conflicts must match the active selected fact snapshot")
	}
	if (claim.GroundAtomEvidence == nil) != (state.result.GroundAtomEvidence == nil) {
		return planInvalid("EVALUATION_PROOF_GROUND_EVIDENCE_MISMATCH", "ground_atom_evidence", "optional closed-atom evidence must match the selected fact snapshot")
	}
	if claim.GroundAtomEvidence != nil && !sameNaryEvidence(*claim.GroundAtomEvidence, *state.result.GroundAtomEvidence) {
		return planInvalid("EVALUATION_PROOF_GROUND_EVIDENCE_MISMATCH", "ground_atom_evidence", "closed-atom evidence differs from the selected fact snapshot")
	}
	return nil
}

func verifyAlgebraProofDAG(claim AlgebraEvaluationResult, state *algebraEvaluationState, result *AlgebraEvaluationProofVerification) (map[string]verifiedAlgebraProofNode, map[string]verifiedAlgebraProofNode, error) {
	if err := state.charge(len(state.plan.nodes)+len(state.plan.variables)+len(state.plan.references), "proof.plan_index"); err != nil {
		return nil, nil, err
	}
	planNodes := make(map[string]AlgebraPlanNode, len(state.plan.nodes))
	for _, node := range state.plan.nodes {
		if _, exists := planNodes[node.ID]; exists {
			return nil, nil, planInvalid("DUPLICATE_PLAN_NODE_ID", "plan.nodes", "compiled plan contains duplicate node identifiers")
		}
		planNodes[node.ID] = node
	}
	variablesByPath := make(map[string][]AlgebraPlanVariable, len(state.plan.variables))
	for _, variable := range state.plan.variables {
		variablesByPath[variable.SourcePath] = append(variablesByPath[variable.SourcePath], variable)
	}
	indices := make(map[string]int, len(claim.Proof))
	for i, node := range claim.Proof {
		if _, exists := indices[node.ID]; exists {
			return nil, nil, planInvalid("DUPLICATE_PROOF_NODE_ID", fmt.Sprintf("proof[%d].id", i), "proof node IDs must be unique")
		}
		indices[node.ID] = i
	}
	verified := make(map[string]verifiedAlgebraProofNode, len(claim.Proof))
	for i, node := range claim.Proof {
		if err := state.checkContext(); err != nil {
			return nil, nil, err
		}
		if err := validateAlgebraProofNodeShape(node, i, state.limits); err != nil {
			return nil, nil, err
		}
		for _, premise := range node.Premises {
			position, exists := indices[premise]
			if !exists || position >= i {
				return nil, nil, planInvalid("INVALID_PROOF_PREMISE_ORDER", fmt.Sprintf("proof[%d].premises", i), "every premise must identify an earlier proof node")
			}
		}
		verifiedNode, err := verifyOneAlgebraProofNode(node, i, verified, planNodes, variablesByPath, state)
		if err != nil {
			return nil, nil, err
		}
		verified[node.ID] = verifiedNode
		result.VerifiedProofNodes++
	}
	return verified, verified, nil
}

func allAlgebraProofNodesReachable(claim AlgebraEvaluationResult, roots map[string]struct{}, verified map[string]verifiedAlgebraProofNode) bool {
	nodes := make(map[string]AlgebraProofNode, len(claim.Proof))
	for _, node := range claim.Proof {
		nodes[node.ID] = node
	}
	seen := make(map[string]struct{}, len(claim.Proof))
	stack := make([]string, 0, len(roots))
	for id := range roots {
		stack = append(stack, id)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if _, exists := seen[id]; exists {
			continue
		}
		if _, ok := verified[id]; !ok {
			return false
		}
		node, ok := nodes[id]
		if !ok {
			return false
		}
		seen[id] = struct{}{}
		stack = append(stack, node.Premises...)
	}
	return len(seen) == len(claim.Proof)
}

func validateAlgebraProofNodeShape(node AlgebraProofNode, index int, limits AlgebraEvaluationLimits) error {
	path := fmt.Sprintf("proof[%d]", index)
	if !validPlanText(node.ID, 256) || !validPlanText(node.RuleID, 64) || !validPlanText(node.PlanNodeID, 256) || node.Depth < 1 || node.Depth > limits.MaxProofDepth {
		return planInvalid("INVALID_PROOF_NODE", path, "proof node identity, rule, plan reference, or depth is invalid")
	}
	if node.Premises == nil || node.FactIDs == nil || node.Sources == nil || node.Bindings == nil || node.OpposingNegativeFactIDs == nil {
		return planInvalid("NULL_PROOF_NODE_ARRAY", path, "proof node arrays must be present and non-null")
	}
	if len(node.Premises) > limits.MaxProofNodes || len(node.FactIDs) > facts.DefaultNaryMaxFacts || len(node.Sources) > facts.DefaultNaryMaxFacts || len(node.Bindings) > MaxAlgebraVariables || len(node.OpposingNegativeFactIDs) > facts.DefaultNaryMaxFacts {
		return planResource(path, "proof node arrays exceed the bounded verification limits")
	}
	for j, value := range node.Premises {
		if !validPlanText(value, 256) {
			return planInvalid("INVALID_PROOF_PREMISE", fmt.Sprintf("%s.premises[%d]", path, j), "premise ID is invalid or too long")
		}
	}
	for j, value := range node.FactIDs {
		if !validPlanText(value, 128) {
			return planInvalid("INVALID_PROOF_FACT_ID", fmt.Sprintf("%s.fact_ids[%d]", path, j), "fact ID is invalid or too long")
		}
	}
	for j, binding := range node.Bindings {
		if !validPlanText(binding.Variable.Kind, 32) || !validPlanText(binding.Variable.ID, 256) || binding.Variable.Kind != AlgebraSelectionLambda && binding.Variable.Kind != AlgebraSelectionParameter && binding.Variable.Kind != "exists" || binding.Value.Kind != algebra.TermRef || !validPlanText(binding.Value.ID, 256) {
			return planInvalid("INVALID_PROOF_BINDING", fmt.Sprintf("%s.bindings[%d]", path, j), "binding must retain a bounded query variable and explicit reference identity")
		}
	}
	for j, value := range node.OpposingNegativeFactIDs {
		if !validPlanText(value, 128) {
			return planInvalid("INVALID_PROOF_FACT_ID", fmt.Sprintf("%s.opposing_negative_fact_ids[%d]", path, j), "negative fact ID is invalid or too long")
		}
	}
	if node.Atom != nil {
		if !validPlanText(node.Atom.Symbol, 256) || node.Atom.Args == nil || len(node.Atom.Args) > algebra.DefaultMaxPorts {
			return planInvalid("INVALID_PROOF_ATOM", path+".atom", "atom symbol and ordered arguments are invalid or exceed their bound")
		}
		for j, argument := range node.Atom.Args {
			if !validPlanText(argument.Role, 256) || argument.Value.Kind != algebra.TermRef || !validPlanText(argument.Value.ID, 256) {
				return planInvalid("INVALID_PROOF_ATOM_ARGUMENT", fmt.Sprintf("%s.atom.args[%d]", path, j), "proof atom arguments must be named roles with explicit references")
			}
		}
	}
	if node.Witness != nil && (!validPlanText(node.Witness.Variable.Kind, 32) || node.Witness.Variable.Kind != "exists" || !validPlanText(node.Witness.Variable.ID, 256) || node.Witness.Value.Kind != algebra.TermRef || !validPlanText(node.Witness.Value.ID, 256)) {
		return planInvalid("INVALID_PROOF_WITNESS", path+".witness", "witness must name an exact scoped variable and reference")
	}
	if node.Reference != nil && (!validPlanText(node.Reference.ID, 256)) {
		return planInvalid("INVALID_PROOF_REFERENCE", path+".reference", "reference declaration ID is invalid or too long")
	}
	for j, source := range node.Sources {
		if err := validateClaimSource(source, fmt.Sprintf("%s.sources[%d]", path, j)); err != nil {
			return err
		}
	}
	return nil
}

func validateClaimAtom(atom facts.NaryAtom, path string) error {
	if !validPlanText(atom.Symbol, 256) || atom.Args == nil || len(atom.Args) > algebra.DefaultMaxPorts {
		return planInvalid("INVALID_EVALUATION_ATOM", path, "atom symbol and argument array must be bounded")
	}
	for i, arg := range atom.Args {
		if !validPlanText(arg.Role, 256) || arg.Value.Kind != algebra.TermRef || !validPlanText(arg.Value.ID, 256) {
			return planInvalid("INVALID_EVALUATION_ATOM_ARGUMENT", fmt.Sprintf("%s.args[%d]", path, i), "evidence atom arguments must be bounded named references")
		}
	}
	return nil
}

func validateClaimSource(source facts.SourceEvidence, path string) error {
	if !validPlanText(string(source.Kind), 32) || !validPlanText(source.ID, provenance.MaxSourceIDBytes) {
		return planInvalid("INVALID_EVALUATION_SOURCE", path, "source kind and ID must be bounded UTF-8")
	}
	switch source.Kind {
	case facts.SourceUser:
		if source.Reference != nil {
			return planInvalid("INVALID_EVALUATION_SOURCE", path+".reference", "user source cannot include an external reference")
		}
	case facts.SourceDocument:
		if source.Reference == nil {
			return planInvalid("INVALID_EVALUATION_SOURCE", path+".reference", "document source requires an external reference")
		}
		if err := source.Reference.Validate(); err != nil {
			return planInvalid("INVALID_EVALUATION_SOURCE", path+".reference", err.Error())
		}
	default:
		return planInvalid("INVALID_EVALUATION_SOURCE", path+".kind", "source kind must be user or document")
	}
	return nil
}

func validateClaimFactRecord(record facts.NaryFact, path string) error {
	if !validPlanText(record.ID, 128) {
		return planInvalid("INVALID_EVALUATION_FACT", path+".id", "fact ID must be bounded UTF-8")
	}
	if err := validateClaimAtom(record.Atom, path+".atom"); err != nil {
		return err
	}
	if record.Polarity != facts.Positive && record.Polarity != facts.Negative {
		return planInvalid("INVALID_EVALUATION_FACT", path+".polarity", "polarity must be positive or negative")
	}
	for name, value := range map[string]string{"from": record.ValidTime.From, "to": record.ValidTime.To} {
		if value != "" {
			if len(value) > 128 || !utf8.ValidString(value) || !strings.HasSuffix(value, "Z") {
				return planInvalid("INVALID_EVALUATION_FACT", path+".valid_time."+name, "time endpoint must be bounded RFC3339 UTC")
			}
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				return planInvalid("INVALID_EVALUATION_FACT", path+".valid_time."+name, "time endpoint must be bounded RFC3339 UTC")
			}
		}
	}
	if err := validateClaimSource(record.Source, path+".source"); err != nil {
		return err
	}
	if record.AssertionScope != facts.ScopeAsserted {
		return planInvalid("INVALID_EVALUATION_FACT", path+".assertion_scope", "fact scope must be asserted")
	}
	return nil
}

func verifyOneAlgebraProofNode(claim AlgebraProofNode, index int, prior map[string]verifiedAlgebraProofNode, planNodes map[string]AlgebraPlanNode, variablesByPath map[string][]AlgebraPlanVariable, state *algebraEvaluationState) (verifiedAlgebraProofNode, error) {
	path := fmt.Sprintf("proof[%d]", index)
	if err := state.charge(1+len(claim.Premises)+len(claim.FactIDs)+len(claim.Sources)+len(claim.Bindings)+len(claim.OpposingNegativeFactIDs), path); err != nil {
		return verifiedAlgebraProofNode{}, err
	}
	fieldsEmpty := func(atom *facts.NaryAtom, witness *AlgebraBinding, reference *algebra.Declaration, factIDs []string, sources []facts.SourceEvidence, negatives []string) bool {
		return atom == nil && witness == nil && reference == nil && len(factIDs) == 0 && len(sources) == 0 && len(negatives) == 0
	}
	switch claim.RuleID {
	case AlgebraProofFactMatch:
		planNode, ok := planNodes[claim.PlanNodeID]
		if !ok || planNode.Kind != AlgebraPlanAtom {
			return verifiedAlgebraProofNode{}, planInvalid("PROOF_SOURCE_NODE_MISMATCH", path+".plan_node_id", "fact_match must refer to one actual source Atom occurrence")
		}
		if len(claim.Premises) != 0 || claim.Atom == nil || claim.Witness != nil || claim.Reference != nil || len(claim.FactIDs) == 0 || len(claim.Sources) != len(claim.FactIDs) {
			return verifiedAlgebraProofNode{}, planInvalid("INVALID_FACT_MATCH_SHAPE", path, "fact_match requires one atom, exact positive records, no premises, and no unrelated fields")
		}
		rowIndex, exists := state.byAtom[evaluationAtomKey(*claim.Atom)]
		if !exists {
			return verifiedAlgebraProofNode{}, planInvalid("FACT_MATCH_NOT_ACTIVE", path+".atom", "claimed atom does not identify an active evidence row")
		}
		row := state.rows[rowIndex]
		if len(row.positive) == 0 || !sameNaryAtom(*claim.Atom, row.atom) {
			return verifiedAlgebraProofNode{}, planInvalid("FACT_MATCH_NOT_POSITIVE", path+".atom", "fact_match requires the exact grounded atom and active positive evidence")
		}
		bindings, err := expectedAlgebraAtomBindings(planNode, row, state)
		if err != nil {
			return verifiedAlgebraProofNode{}, err
		}
		if !sameAlgebraBindings(claim.Bindings, bindings) {
			return verifiedAlgebraProofNode{}, planInvalid("FACT_MATCH_BINDINGS_MISMATCH", path+".bindings", "atom bindings must match the source roles and exact grounded references")
		}
		if !sameStringSlice(claim.FactIDs, naryFactIDs(row.positive)) || !sameSourceSlice(claim.Sources, narySources(row.positive)) || !sameStringSlice(claim.OpposingNegativeFactIDs, naryFactIDs(row.negative)) {
			return verifiedAlgebraProofNode{}, planInvalid("FACT_MATCH_EVIDENCE_MISMATCH", path, "positive facts, source records, and opposing active IDs must match the selected time/scope-filtered snapshot")
		}
		return verifiedAlgebraProofNode{planNodeID: planNode.ID, ruleID: claim.RuleID, bindings: bindings, conflict: len(row.negative) != 0, depth: 1}, nil
	case AlgebraProofAnd:
		planNode, ok := planNodes[claim.PlanNodeID]
		if !ok || planNode.Kind != AlgebraPlanAnd || len(claim.Premises) != len(planNode.Children) || !fieldsEmpty(claim.Atom, claim.Witness, claim.Reference, claim.FactIDs, claim.Sources, claim.OpposingNegativeFactIDs) {
			return verifiedAlgebraProofNode{}, planInvalid("INVALID_AND_PROOF", path, "And proof must match one source And node and its ordered children")
		}
		merged := []AlgebraBinding{}
		maxDepth, conflict := 0, false
		for i, childID := range planNode.Children {
			childProof, ok := prior[claim.Premises[i]]
			if !ok || childProof.planNodeID != childID {
				return verifiedAlgebraProofNode{}, planInvalid("AND_PREMISE_MISMATCH", fmt.Sprintf("%s.premises[%d]", path, i), "And premises must prove the corresponding source children in order")
			}
			var mergeErr error
			merged, mergeErr = mergeAlgebraProofBindings(merged, childProof.bindings, path)
			if mergeErr != nil {
				return verifiedAlgebraProofNode{}, mergeErr
			}
			if childProof.depth > maxDepth {
				maxDepth = childProof.depth
			}
			conflict = conflict || childProof.conflict
		}
		if !sameAlgebraBindings(claim.Bindings, merged) || claim.Depth != maxDepth+1 {
			return verifiedAlgebraProofNode{}, planInvalid("AND_CONCLUSION_MISMATCH", path, "And bindings and depth must be recomputed from its premises")
		}
		return verifiedAlgebraProofNode{planNodeID: planNode.ID, ruleID: claim.RuleID, bindings: merged, conflict: conflict, depth: claim.Depth}, nil
	case AlgebraProofExists:
		planNode, ok := planNodes[claim.PlanNodeID]
		if !ok || planNode.Kind != AlgebraPlanExists || planNode.Binder == nil || len(planNode.Children) != 1 || claim.Witness == nil || claim.Reference != nil || claim.Atom != nil || len(claim.FactIDs) != 0 || len(claim.Sources) != 0 || len(claim.OpposingNegativeFactIDs) != 0 {
			return verifiedAlgebraProofNode{}, planInvalid("INVALID_EXISTS_PROOF", path, "Exists proof must match an actual source binder and carry only its explicit witness")
		}
		variable, ok := state.planVariables[*planNode.Binder]
		if !ok || claim.Witness.Variable != *planNode.Binder || claim.Witness.Value.Kind != algebra.TermRef {
			return verifiedAlgebraProofNode{}, planInvalid("EXISTS_WITNESS_BINDER_MISMATCH", path+".witness", "witness must name the exact scoped source binder")
		}
		declaration, ok := state.knownRefs[claim.Witness.Value.ID]
		if !ok || !algebra.TypeEqual(declaration.Type, variable.Type) {
			return verifiedAlgebraProofNode{}, planInvalid("EXISTS_WITNESS_TYPE_MISMATCH", path+".witness", "witness reference must be explicitly known at the binder's exact type")
		}
		if len(claim.Premises) != 1 && len(claim.Premises) != 2 {
			return verifiedAlgebraProofNode{}, planInvalid("EXISTS_PREMISE_COUNT", path+".premises", "Exists proof requires its body and, only for an unused binder, a reference premise")
		}
		body, ok := prior[claim.Premises[0]]
		if !ok || body.planNodeID != planNode.Children[0] {
			return verifiedAlgebraProofNode{}, planInvalid("EXISTS_BODY_MISMATCH", path+".premises[0]", "first premise must prove the exact Exists body")
		}
		bodyWitness, bodyHasWitness := findProofBinding(body.bindings, *planNode.Binder)
		outward, found := removeAlgebraProofBinding(body.bindings, *planNode.Binder)
		depth := body.depth
		if variable.UsedInAtom {
			if !found || !bodyHasWitness || bodyWitness != claim.Witness.Value || len(claim.Premises) != 1 {
				return verifiedAlgebraProofNode{}, planInvalid("EXISTS_SCOPE_MISMATCH", path, "used binder must be grounded in its body and has no separate reference premise")
			}
		} else {
			if found || len(claim.Premises) != 2 {
				return verifiedAlgebraProofNode{}, planInvalid("EXISTS_UNUSED_BINDER_MISMATCH", path, "unused binder requires a separate typed reference premise")
			}
			witnessProof, ok := prior[claim.Premises[1]]
			if !ok || witnessProof.ruleID != AlgebraProofReference || witnessProof.planNodeID != variable.SourcePath || len(witnessProof.bindings) != 1 || witnessProof.bindings[0].Variable != *planNode.Binder || witnessProof.bindings[0].Value != claim.Witness.Value {
				return verifiedAlgebraProofNode{}, planInvalid("EXISTS_REFERENCE_PREMISE_MISMATCH", path+".premises[1]", "unused binder premise must independently cite its exact declared reference")
			}
			if witnessProof.depth > depth {
				depth = witnessProof.depth
			}
		}
		if !sameAlgebraBindings(claim.Bindings, outward) || claim.Depth != depth+1 {
			return verifiedAlgebraProofNode{}, planInvalid("EXISTS_CONCLUSION_MISMATCH", path, "Exists scope, outward bindings, or depth does not follow from verified premises")
		}
		return verifiedAlgebraProofNode{planNodeID: planNode.ID, ruleID: claim.RuleID, bindings: outward, conflict: body.conflict, depth: claim.Depth}, nil
	case AlgebraProofReference:
		candidates := variablesByPath[claim.PlanNodeID]
		if len(candidates) == 0 || claim.Reference == nil || claim.Atom != nil || claim.Witness != nil || len(claim.Premises) != 0 || len(claim.FactIDs) != 0 || len(claim.Sources) != 0 || len(claim.OpposingNegativeFactIDs) != 0 || claim.Depth != 1 || len(claim.Bindings) != 1 {
			return verifiedAlgebraProofNode{}, planInvalid("INVALID_REFERENCE_PROOF", path, "reference rule must cite one actual unused variable path and no unrelated fields")
		}
		binding := claim.Bindings[0]
		var variable AlgebraPlanVariable
		foundVariable := false
		for _, candidate := range candidates {
			if candidate.Key == binding.Variable {
				variable, foundVariable = candidate, true
				break
			}
		}
		if !foundVariable {
			return verifiedAlgebraProofNode{}, planInvalid("REFERENCE_BINDING_MISMATCH", path+".bindings", "reference premise must name a plan variable at this source path")
		}
		if binding.Variable != variable.Key || binding.Value.Kind != algebra.TermRef || binding.Value.ID != claim.Reference.ID {
			return verifiedAlgebraProofNode{}, planInvalid("REFERENCE_BINDING_MISMATCH", path, "reference declaration and its namespace-aware binding must agree")
		}
		actual, known := state.knownRefs[claim.Reference.ID]
		if !known || !algebra.TypeEqual(claim.Reference.Type, actual.Type) || !algebra.TypeEqual(variable.Type, actual.Type) {
			return verifiedAlgebraProofNode{}, planInvalid("REFERENCE_TYPE_MISMATCH", path+".reference", "reference premise must use the exact selected declaration and variable type")
		}
		return verifiedAlgebraProofNode{planNodeID: variable.SourcePath, ruleID: claim.RuleID, bindings: []AlgebraBinding{binding}, depth: 1}, nil
	case AlgebraProofProject:
		if claim.PlanNodeID != state.plan.root || claim.Atom != nil || claim.Witness != nil || claim.Reference != nil || len(claim.FactIDs) != 0 || len(claim.Sources) != 0 || len(claim.OpposingNegativeFactIDs) != 0 || len(claim.Premises) == 0 {
			return verifiedAlgebraProofNode{}, planInvalid("INVALID_PROJECT_PROOF", path, "project rule must reference the actual query root and contain only ordered support premises")
		}
		body, ok := prior[claim.Premises[0]]
		if !ok || body.planNodeID != state.plan.root {
			return verifiedAlgebraProofNode{}, planInvalid("PROJECT_BODY_MISMATCH", path+".premises[0]", "first project premise must prove the actual query root")
		}
		bindings := make([]AlgebraBinding, 0, len(state.plan.answers))
		premiseIndex := 1
		depth, conflict := body.depth, body.conflict
		for _, answer := range state.plan.answers {
			variable, known := state.planVariables[answer.Variable]
			if !known {
				return verifiedAlgebraProofNode{}, planInvalid("UNKNOWN_ANSWER_VARIABLE", path+".bindings", "selected answer variable is absent from the compiled plan")
			}
			value, found := findProofBinding(body.bindings, answer.Variable)
			if !variable.UsedInAtom {
				if found {
					return verifiedAlgebraProofNode{}, planInvalid("PROJECT_UNUSED_VARIABLE_BOUND", path, "unused selected variable must be supplied only by its explicit reference premise")
				}
				if premiseIndex >= len(claim.Premises) {
					return verifiedAlgebraProofNode{}, planInvalid("PROJECT_REFERENCE_PREMISE_MISSING", path+".premises", "unused selected variable requires its typed known-reference premise")
				}
				ref, ok := prior[claim.Premises[premiseIndex]]
				if !ok || ref.ruleID != AlgebraProofReference || ref.planNodeID != variable.SourcePath || len(ref.bindings) != 1 || ref.bindings[0].Variable != answer.Variable {
					return verifiedAlgebraProofNode{}, planInvalid("PROJECT_REFERENCE_PREMISE_MISMATCH", fmt.Sprintf("%s.premises[%d]", path, premiseIndex), "selected unused output requires its own exact reference premise")
				}
				value, found = ref.bindings[0].Value, true
				if ref.depth > depth {
					depth = ref.depth
				}
				premiseIndex++
			} else if !found {
				return verifiedAlgebraProofNode{}, planInvalid("PROJECT_BINDING_MISSING", path+".bindings", "selected output must be grounded by its query proof")
			}
			declaration, exists := state.knownRefs[value.ID]
			if !found || value.Kind != algebra.TermRef || !exists || !algebra.TypeEqual(declaration.Type, answer.Type) {
				return verifiedAlgebraProofNode{}, planInvalid("PROJECT_BINDING_TYPE_MISMATCH", path+".bindings", "projected output must be an explicit reference at the selected column type")
			}
			bindings = append(bindings, AlgebraBinding{Variable: answer.Variable, Value: value})
		}
		if premiseIndex != len(claim.Premises) || claim.Depth != depth+1 || !sameAlgebraBindings(claim.Bindings, bindings) {
			return verifiedAlgebraProofNode{}, planInvalid("PROJECT_CONCLUSION_MISMATCH", path, "project must exactly bind columns in order, retain valid premises, and recompute depth")
		}
		return verifiedAlgebraProofNode{planNodeID: state.plan.root, ruleID: claim.RuleID, bindings: bindings, conflict: conflict, depth: claim.Depth}, nil
	default:
		return verifiedAlgebraProofNode{}, planUnsupported("UNSUPPORTED_PROOF_RULE", path+".rule_id", "only the versioned positive query proof rules are supported")
	}
}

func expectedAlgebraAtomBindings(node AlgebraPlanNode, row algebraEvidenceRow, state *algebraEvaluationState) ([]AlgebraBinding, error) {
	if len(node.Args) != len(row.atom.Args) {
		return nil, planInvalid("FACT_MATCH_ARITY_MISMATCH", node.Path, "active evidence row and source occurrence have different arity")
	}
	out := make([]AlgebraBinding, 0, len(node.Args))
	positions := make(map[AlgebraVariableKey]int, len(node.Args))
	for i, argument := range node.Args {
		factArgument := row.atom.Args[i]
		if factArgument.Role != argument.Role || factArgument.Value.Kind != algebra.TermRef {
			return nil, planInvalid("FACT_MATCH_ROLE_MISMATCH", fmt.Sprintf("%s.args[%d]", node.Path, i), "grounded fact roles must match the source signature in declared order")
		}
		if argument.Value.Kind == string(algebra.TermRef) {
			if argument.RefID == "" || argument.RefID != argument.Value.ID || factArgument.Value.ID != argument.RefID {
				return nil, planInvalid("FACT_MATCH_REFERENCE_MISMATCH", fmt.Sprintf("%s.args[%d]", node.Path, i), "explicit source references must retain their exact identity")
			}
			continue
		}
		variable, ok := state.planVariables[argument.Value]
		if !ok {
			return nil, planInvalid("FACT_MATCH_VARIABLE_UNKNOWN", fmt.Sprintf("%s.args[%d]", node.Path, i), "source variable is absent from the checked plan")
		}
		declaration, known := state.knownRefs[factArgument.Value.ID]
		if !known || !algebra.TypeEqual(declaration.Type, variable.Type) {
			return nil, planInvalid("FACT_MATCH_REFERENCE_TYPE", fmt.Sprintf("%s.args[%d]", node.Path, i), "grounding reference must be known at the source variable type")
		}
		if prior, exists := positions[argument.Value]; exists {
			if out[prior].Value.ID != factArgument.Value.ID {
				return nil, planInvalid("FACT_MATCH_REPEATED_VARIABLE", fmt.Sprintf("%s.args[%d]", node.Path, i), "repeated variable occurrences must retain one exact identity")
			}
			continue
		}
		positions[argument.Value] = len(out)
		out = append(out, AlgebraBinding{Variable: argument.Value, Value: factArgument.Value})
	}
	return out, nil
}

func mergeAlgebraProofBindings(left, right []AlgebraBinding, path string) ([]AlgebraBinding, error) {
	out := append([]AlgebraBinding{}, left...)
	positions := make(map[AlgebraVariableKey]int, len(left)+len(right))
	for i, binding := range out {
		positions[binding.Variable] = i
	}
	for _, binding := range right {
		if index, exists := positions[binding.Variable]; exists {
			if out[index].Value != binding.Value {
				return nil, planInvalid("PROOF_BINDING_CONFLICT", path, "conjunct premises disagree about one namespace-aware variable")
			}
			continue
		}
		positions[binding.Variable] = len(out)
		out = append(out, binding)
	}
	return out, nil
}

func removeAlgebraProofBinding(bindings []AlgebraBinding, key AlgebraVariableKey) ([]AlgebraBinding, bool) {
	out := make([]AlgebraBinding, 0, len(bindings))
	found := false
	for _, binding := range bindings {
		if binding.Variable == key {
			if found {
				return nil, false
			}
			found = true
			continue
		}
		out = append(out, binding)
	}
	return out, found
}

func findProofBinding(bindings []AlgebraBinding, key AlgebraVariableKey) (facts.NaryValue, bool) {
	for _, binding := range bindings {
		if binding.Variable == key {
			return binding.Value, true
		}
	}
	return facts.NaryValue{}, false
}

func sameAlgebraBindings(left, right []AlgebraBinding) bool {
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

func sameAlgebraAnswerBindings(bindings []AlgebraBinding, ports []AlgebraAnswerPort, values []facts.NaryValue) bool {
	if len(bindings) != len(ports) || len(values) != len(ports) {
		return false
	}
	for i, port := range ports {
		if bindings[i].Variable != port.Variable || bindings[i].Value != values[i] {
			return false
		}
	}
	return true
}

func sameNaryAtom(left, right facts.NaryAtom) bool {
	return left.Symbol == right.Symbol && reflect.DeepEqual(left.Args, right.Args)
}

func sameNaryFact(left, right facts.NaryFact) bool {
	return left.ID == right.ID && sameNaryAtom(left.Atom, right.Atom) && left.Polarity == right.Polarity && left.ValidTime == right.ValidTime && left.AssertionScope == right.AssertionScope && sameSourceEvidence(left.Source, right.Source)
}

func sameSourceEvidence(left, right facts.SourceEvidence) bool {
	if left.Kind != right.Kind || left.ID != right.ID || (left.Reference == nil) != (right.Reference == nil) {
		return false
	}
	if left.Reference == nil {
		return true
	}
	return *left.Reference == *right.Reference
}

func sameNaryEvidence(left, right facts.NaryEvidence) bool {
	if left.State != right.State || !sameNaryFactSlice(left.Positive, right.Positive) || !sameNaryFactSlice(left.Negative, right.Negative) {
		return false
	}
	return true
}

func sameNaryFactSlice(left, right []facts.NaryFact) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !sameNaryFact(left[i], right[i]) {
			return false
		}
	}
	return true
}

func sameNaryConflicts(left, right []facts.NaryConflict) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !sameNaryAtom(left[i].Atom, right[i].Atom) || !sameStringSlice(left[i].PositiveIDs, right[i].PositiveIDs) || !sameStringSlice(left[i].NegativeIDs, right[i].NegativeIDs) {
			return false
		}
	}
	return true
}

func sameSourceSlice(left, right []facts.SourceEvidence) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !sameSourceEvidence(left[i], right[i]) {
			return false
		}
	}
	return true
}

func naryFactIDs(input []facts.NaryFact) []string {
	ids := make([]string, len(input))
	for i, fact := range input {
		ids[i] = fact.ID
	}
	return ids
}

func narySources(input []facts.NaryFact) []facts.SourceEvidence {
	out := make([]facts.SourceEvidence, len(input))
	for i, fact := range input {
		out[i] = fact.Source
	}
	return out
}

func sameStringSlice(left, right []string) bool {
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

func answerTupleIdentity(values []facts.NaryValue) string {
	var builder strings.Builder
	for _, value := range values {
		writeEvaluationKeyPart(&builder, string(value.Kind))
		writeEvaluationKeyPart(&builder, value.ID)
	}
	return builder.String()
}

package synthesis

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	searchAlphaProfileID        = "dse-search-alpha-v1"
	searchIdentityVersion       = "dse-search-candidate-identity-v1"
	maxSearchIdentityBytes      = 16 << 20
	SearchStopAlphaIdentity     = SearchStopReason("alpha_identity_incomplete")
	SearchStopAlphaBudget       = SearchStopReason("alpha_normalization_budget")
	SearchStopUnresolved        = SearchStopReason("unresolved_goal")
	SearchStopConstruction      = SearchStopReason("construction_incomplete")
	SearchStopMaxAtoms          = SearchStopReason("max_atoms")
	SearchStopMaxSteps          = SearchStopReason("max_steps")
	SearchStopConstructionLimit = SearchStopReason("construction_resource_limit")
	SearchStopGoalVerification  = SearchStopReason("goal_verification_incomplete")
	searchArtifactProfileID     = "dse-search-artifact-v1"
	searchResultSchema          = 1
	searchConstructionReserve   = algebra.DefaultMaxNodes + MaxGoalSteps
)

// searchCandidateIdentity is request-local identity material. A key is
// available only after both normalization and independent certificate replay
// establish the licensed alpha-only normal form.
type searchCandidateIdentity struct {
	key             string
	complete        bool
	comparisonBytes []byte
	canonicalExpr   *algebra.Term
	normalization   algebra.NormalizationResult
	verification    algebra.VerificationResult
}

type searchIdentityDescriptor struct {
	Version                   string                      `json:"version"`
	SnapshotID                string                      `json:"snapshot_id"`
	ContextID                 string                      `json:"context_id"`
	RequestContextFingerprint string                      `json:"request_context_fingerprint,omitempty"`
	ExpectedType              algebra.Type                `json:"expected_type"`
	Ports                     []searchIdentityPort        `json:"ports"`
	References                []algebra.Declaration       `json:"references"`
	ProjectPorts              []string                    `json:"project_ports"`
	Constraints               []searchIdentityConstraint  `json:"constraints"`
	Choices                   []ChoiceAssignment          `json:"choices"`
	Fragments                 []searchIdentityFragmentMap `json:"fragments"`
	Expression                *algebra.Term               `json:"expression"`
}

type searchIdentityPort struct {
	ID      string       `json:"id"`
	Type    algebra.Type `json:"type"`
	Binding *GoalValue   `json:"binding,omitempty"`
}

type searchIdentityConstraint struct {
	ID                string         `json:"id"`
	Kind              ConstraintKind `json:"kind"`
	Symbol            string         `json:"symbol,omitempty"`
	FromPort          string         `json:"from_port,omitempty"`
	ToPort            string         `json:"to_port,omitempty"`
	FromRole          string         `json:"from_role,omitempty"`
	ToRole            string         `json:"to_role,omitempty"`
	AllowedAtomCounts []int          `json:"allowed_atom_counts,omitempty"`
	Port              string         `json:"port,omitempty"`
}

type searchIdentityFragmentMap struct {
	ConstraintID string         `json:"constraint_id"`
	Kind         ConstraintKind `json:"kind"`
	Fragments    []string       `json:"fragments"`
}

// certifySearchCandidate gives a fully constructed candidate a conservative
// alpha-equivalence identity. It never establishes that the represented
// proposition is true or that its language interpretation is approved.
func certifySearchCandidate(environment algebra.RewriteEnvironment, concreteGoal GoalSpec, choices []ChoiceAssignment, constructed Result, budget *searchBudget) (searchCandidateIdentity, bool, error) {
	if budget == nil {
		return searchCandidateIdentity{}, false, searchIdentityInvalid("INVALID_SEARCH_BUDGET", "budget", "a request-local search budget is required")
	}
	if constructed.Status != status.OutcomeOK {
		return searchCandidateIdentity{}, false, nil
	}
	if constructed.Expression == nil || constructed.Checked == nil || constructed.CanonicalDocument == nil {
		return searchCandidateIdentity{}, false, searchIdentityInvalid("INCOMPLETE_CONSTRUCTION_RESULT", "constructed", "an OK construction must include a checked expression and canonical document")
	}
	if constructed.Checked.Status != status.OutcomeOK {
		return searchCandidateIdentity{}, false, searchIdentityInvalid("CONSTRUCTION_RESULT_MISMATCH", "constructed.checked.status", "an OK construction must carry an OK checked-expression result")
	}
	if err := budget.checkDeadlineCancel(); !err {
		return searchCandidateIdentity{}, false, nil
	}

	profile := algebra.RewriteProfile{ID: searchAlphaProfileID, Rules: []algebra.RewriteRule{algebra.RewriteAlphaV1}}
	if err := algebra.ValidateRewriteEnvironment(environment); err != nil {
		return searchCandidateIdentity{}, false, err
	}
	remaining := budget.remainingWork()
	if remaining <= 0 {
		budget.stopReason = SearchStopMaxWorkUnits
		return searchCandidateIdentity{}, false, nil
	}
	workLimit := remaining
	if workLimit > algebra.DefaultRewriteMaxWorkNodes {
		workLimit = algebra.DefaultRewriteMaxWorkNodes
	}
	minimumAdmission, err := minimumAlphaAdmissionWork(environment, constructed.Checked.NodeCount, budget)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if budget.stop() != "" {
		return searchCandidateIdentity{}, false, nil
	}
	if minimumAdmission > workLimit {
		budget.stopReason = SearchStopMaxWorkUnits
		return searchCandidateIdentity{}, false, nil
	}
	rewriteBudget := defaultAlphaRewriteBudget(workLimit)
	normalized, err := algebra.Normalize(environment, profile, constructed.Expression, rewriteBudget)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if err := chargeCompletedRewriteWork(budget, normalized.Statistics.WorkNodes, workLimit); err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if !budget.checkDeadlineCancel() {
		return searchCandidateIdentity{}, false, nil
	}
	if err := validateIdentityReferences(environment.Document.References, concreteGoal.References); err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if err := validateIdentityReferences(constructed.CanonicalDocument.References, concreteGoal.References); err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if normalized.Status != status.OutcomeOK || !normalized.Normalized || normalized.Expression == nil {
		if rewriteResultIsWorkLimited(normalized.Status, normalized.Diagnostics) {
			budget.stopReason = SearchStopMaxWorkUnits
		} else if normalized.Status == status.OutcomeResourceLimit {
			budget.stopReason = SearchStopAlphaBudget
		} else {
			budget.stopReason = SearchStopAlphaIdentity
		}
		return searchCandidateIdentity{}, false, nil
	}
	equalConstruction, err := algebra.RewriteTermEqual(constructed.Expression, constructed.CanonicalDocument.Expression, environment.Document.Limits)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if !equalConstruction || normalized.Checked.Status != status.OutcomeOK || !algebra.TypeEqual(normalized.Checked.Type, concreteGoal.ExpectedType) {
		return searchCandidateIdentity{}, false, searchIdentityInvalid("CONSTRUCTION_DOCUMENT_MISMATCH", "constructed.canonical_document", "canonical document must contain the checked expression at the requested goal type")
	}

	remaining = budget.remainingWork()
	if remaining <= 0 {
		budget.stopReason = SearchStopMaxWorkUnits
		return searchCandidateIdentity{}, false, nil
	}
	if !budget.checkDeadlineCancel() {
		return searchCandidateIdentity{}, false, nil
	}
	minimumAdmission, err = minimumAlphaAdmissionWork(environment, constructed.Checked.NodeCount, budget)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if budget.stop() != "" {
		return searchCandidateIdentity{}, false, nil
	}
	if budget.remainingWork() < minimumAdmission {
		budget.stopReason = SearchStopMaxWorkUnits
		return searchCandidateIdentity{}, false, nil
	}
	verifyBudget := defaultAlphaRewriteBudget(minRewriteWork(remaining))
	verifyWorkLimit := verifyBudget.MaxWorkNodes
	verified, err := algebra.VerifyRewriteCertificate(environment, profile, constructed.Expression, normalized.Certificate, verifyBudget)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if err := chargeCompletedRewriteWork(budget, verified.Statistics.WorkNodes, verifyWorkLimit); err != nil {
		return searchCandidateIdentity{}, false, err
	}
	if !budget.checkDeadlineCancel() {
		return searchCandidateIdentity{}, false, nil
	}
	if verified.Status != status.OutcomeOK || !verified.CertificateValid || !verified.NormalizationComplete {
		if rewriteResultIsWorkLimited(verified.Status, verified.Diagnostics) {
			budget.stopReason = SearchStopMaxWorkUnits
		} else if verified.Status == status.OutcomeResourceLimit {
			budget.stopReason = SearchStopAlphaBudget
		} else {
			budget.stopReason = SearchStopAlphaIdentity
		}
		return searchCandidateIdentity{}, false, nil
	}

	limits, err := algebra.NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	identityExpression, err := cloneIdentityTerm(normalized.Expression, limits)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	requestContextFingerprint, err := goalContextFingerprint(environment.Document, concreteGoal)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	descriptor, err := buildSearchIdentityDescriptor(environment.SnapshotID, concreteGoal, choices, constructed.Obligations, identityExpression, limits, requestContextFingerprint)
	if err != nil {
		return searchCandidateIdentity{}, false, err
	}
	comparisonBytes, err := json.Marshal(descriptor)
	if err != nil {
		return searchCandidateIdentity{}, false, searchIdentityInvalid("SEARCH_IDENTITY_ENCODING", "identity", err.Error())
	}
	byteLimit := maxSearchIdentityBytes
	if budget.limits.MaxCandidateBytes < byteLimit {
		byteLimit = budget.limits.MaxCandidateBytes
	}
	if len(comparisonBytes) > byteLimit {
		if len(comparisonBytes) > budget.limits.MaxCandidateBytes {
			budget.stopReason = SearchStopMaxCandidateBytes
		} else {
			budget.stopReason = SearchStopAlphaBudget
		}
		return searchCandidateIdentity{}, false, nil
	}
	if !budget.checkDeadlineCancel() {
		return searchCandidateIdentity{}, false, nil
	}
	digestInput := append([]byte(searchIdentityVersion+"\x00"), comparisonBytes...)
	digest := sha256.Sum256(digestInput)
	return searchCandidateIdentity{
		key: hex.EncodeToString(digest[:]), complete: true,
		comparisonBytes: append([]byte{}, comparisonBytes...), canonicalExpr: identityExpression,
		normalization: normalized, verification: verified,
	}, true, nil
}

func (identity searchCandidateIdentity) sameCandidate(other searchCandidateIdentity, limits algebra.Limits) (bool, error) {
	if !identity.complete || !other.complete || identity.key == "" || identity.key != other.key {
		return false, nil
	}
	if !bytes.Equal(identity.comparisonBytes, other.comparisonBytes) {
		return false, nil
	}
	return algebra.RewriteTermEqual(identity.canonicalExpr, other.canonicalExpr, limits)
}

func defaultAlphaRewriteBudget(maxWork int) algebra.RewriteBudget {
	return algebra.RewriteBudget{
		MaxSteps:      algebra.DefaultRewriteMaxSteps,
		MaxExpansions: algebra.DefaultRewriteMaxExpansions,
		MaxTraceNodes: algebra.DefaultRewriteMaxTraceNodes,
		MaxWorkNodes:  maxWork,
	}
}

func minRewriteWork(remaining int) int {
	if remaining > algebra.DefaultRewriteMaxWorkNodes {
		return algebra.DefaultRewriteMaxWorkNodes
	}
	return remaining
}

func minimumAlphaAdmissionWork(environment algebra.RewriteEnvironment, termNodes int, budget *searchBudget) (int, error) {
	limits, err := algebra.NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return 0, err
	}
	if termNodes < 0 || termNodes > limits.MaxNodes {
		return 0, searchIdentityResource("constructed.checked.node_count", "checked construction node count is outside the configured term limit")
	}
	work := len(environment.Document.Sorts) + len(environment.Document.Signatures) + len(environment.Document.Parameters) + len(environment.Document.References)
	for _, signature := range environment.Document.Signatures {
		work += len(signature.Ports)
	}
	for aliasIndex, alias := range environment.Aliases {
		if !budget.checkDeadlineCancel() {
			return 0, nil
		}
		nodes, err := boundedDistinctTermNodes(alias.Expression, limits)
		if err != nil {
			return 0, err
		}
		work += 1 + nodes + len(alias.Ports) + len(alias.Interface)
		if work > algebra.DefaultRewriteMaxWorkNodes*8 {
			return 0, searchIdentityResource(fmt.Sprintf("environment.aliases[%d]", aliasIndex), "rewrite environment admission estimate exceeds its bounded profile")
		}
	}
	work += termNodes
	return work, nil
}

func boundedDistinctTermNodes(root *algebra.Term, limits algebra.Limits) (int, error) {
	if root == nil {
		return 0, searchIdentityInvalid("MISSING_ALIAS_EXPRESSION", "environment.aliases", "rewrite alias expression is required")
	}
	type pendingNode struct {
		term  *algebra.Term
		depth int
	}
	stack := []pendingNode{{term: root, depth: 1}}
	seen := make(map[*algebra.Term]struct{})
	for len(stack) != 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if item.term == nil {
			continue
		}
		if item.depth > limits.MaxDepth {
			return 0, searchIdentityResource("environment.aliases", "rewrite alias exceeds the configured depth bound")
		}
		if _, exists := seen[item.term]; exists {
			continue
		}
		seen[item.term] = struct{}{}
		if len(seen) > limits.MaxNodes {
			return 0, searchIdentityResource("environment.aliases", "rewrite alias exceeds the configured node bound")
		}
		switch item.term.Kind {
		case algebra.TermAtom:
			for i := len(item.term.Args) - 1; i >= 0; i-- {
				stack = append(stack, pendingNode{term: item.term.Args[i].Value, depth: item.depth + 1})
			}
		case algebra.TermAnd:
			for i := len(item.term.Operands) - 1; i >= 0; i-- {
				stack = append(stack, pendingNode{term: item.term.Operands[i], depth: item.depth + 1})
			}
		case algebra.TermExists, algebra.TermLambda:
			stack = append(stack, pendingNode{term: item.term.Body, depth: item.depth + 1})
		case algebra.TermApply:
			stack = append(stack, pendingNode{term: item.term.Argument, depth: item.depth + 1}, pendingNode{term: item.term.Function, depth: item.depth + 1})
		case algebra.TermVar, algebra.TermParam, algebra.TermRef, algebra.TermHole:
		default:
			return 0, searchIdentityInvalid("UNKNOWN_TERM_KIND", "environment.aliases", "rewrite alias contains a term kind outside the checked union")
		}
	}
	return len(seen), nil
}

func chargeCompletedRewriteWork(budget *searchBudget, work, granted int) error {
	if work < 0 || work > granted || work > budget.remainingWork() {
		return searchIdentityInvalid("INVALID_REWRITE_WORK_REPORT", "rewrite.statistics.work_nodes", fmt.Sprintf("rewrite component reported %d work nodes; grant was %d and request remainder was %d", work, granted, budget.remainingWork()))
	}
	// Charge completed work even if the request context expired during a phase.
	budget.usage.WorkUnits += work
	return nil
}

func chargeCompletedGoalVerificationWork(budget *searchBudget, work, granted int) error {
	if budget == nil || work < 0 || work > granted || work > budget.remainingWork() {
		return searchIdentityInvalid("INVALID_GOAL_VERIFICATION_WORK_REPORT", "verification.statistics.work_nodes", "goal verification reported work outside its finite grant")
	}
	// Charge work completed by a component even if the request deadline expired
	// while that component was running.
	budget.usage.WorkUnits += work
	return nil
}

func isIncompleteError(err error) bool {
	var typed *algebra.Error
	return errors.As(err, &typed) && typed.Status == status.OutcomeIncomplete
}

func fullyValidatedGoal(result GoalVerificationResult) bool {
	return result.Status == status.OutcomeOK && result.Disposition == GoalVerificationDispositionChecked &&
		result.TypeCorrect && result.RefinementChecked && result.ReplayChecked && result.CertificateValid &&
		result.GoalSatisfied && result.Complete
}

func rewriteResultIsWorkLimited(resultStatus status.OutcomeKind, diagnostics []algebra.Error) bool {
	if resultStatus != status.OutcomeResourceLimit {
		return false
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "REWRITE_WORK_LIMIT" {
			return true
		}
	}
	return false
}

func buildSearchIdentityDescriptor(snapshotID string, goal GoalSpec, choices []ChoiceAssignment, obligations []GoalObligation, expression *algebra.Term, limits algebra.Limits, requestContextFingerprint string) (searchIdentityDescriptor, error) {
	if len(goal.Ports) > MaxGoalPorts || len(goal.References) > algebra.DefaultMaxSymbols || len(goal.Constraints) > MaxGoalConstraints || len(choices) > MaxGoalConstraints+MaxGoalPorts+2 || len(obligations) > MaxGoalConstraints {
		return searchIdentityDescriptor{}, searchIdentityResource("identity", "goal identity metadata exceeds its bounded profile")
	}
	for _, choice := range choices {
		for _, field := range []string{choice.Kind, choice.AxisID, choice.ConstraintID, choice.Symbol, choice.FromRole, choice.ToRole, choice.PortID, choice.UnaryRole, choice.HoleID, choice.ReferenceID} {
			if len(field) > 256 || !utf8.ValidString(field) {
				return searchIdentityDescriptor{}, searchIdentityResource("choices", "choice identifier exceeds its bounded UTF-8 size")
			}
		}
	}
	descriptor := searchIdentityDescriptor{
		Version: searchIdentityVersion, SnapshotID: snapshotID, ContextID: goal.ContextID, RequestContextFingerprint: requestContextFingerprint,
		ExpectedType: goal.ExpectedType, Ports: make([]searchIdentityPort, len(goal.Ports)),
		References:   append([]algebra.Declaration{}, goal.References...),
		ProjectPorts: append([]string{}, goal.ProjectPorts...),
		Constraints:  make([]searchIdentityConstraint, len(goal.Constraints)),
		Choices:      append([]ChoiceAssignment{}, choices...), Expression: expression,
	}
	for i, port := range goal.Ports {
		descriptor.Ports[i] = searchIdentityPort{ID: port.ID, Type: port.Type}
		if port.Binding != nil {
			binding := *port.Binding
			if port.Binding.Type != nil {
				typ, err := cloneIdentityType(*port.Binding.Type, limits)
				if err != nil {
					return searchIdentityDescriptor{}, err
				}
				binding.Type = &typ
			}
			descriptor.Ports[i].Binding = &binding
		}
	}
	sort.Slice(descriptor.References, func(i, j int) bool { return descriptor.References[i].ID < descriptor.References[j].ID })
	sort.Strings(descriptor.ProjectPorts)
	for i, constraint := range goal.Constraints {
		descriptor.Constraints[i] = searchIdentityConstraint{
			ID: constraint.ID, Kind: constraint.Kind, Symbol: constraint.Symbol,
			FromPort: constraint.FromPort, ToPort: constraint.ToPort,
			FromRole: constraint.FromRole, ToRole: constraint.ToRole,
			AllowedAtomCounts: append([]int{}, constraint.AllowedAtomCounts...), Port: constraint.Port,
		}
	}
	obligationByID := make(map[string]GoalObligation, len(obligations))
	for _, obligation := range obligations {
		if _, exists := obligationByID[obligation.ConstraintID]; exists {
			return searchIdentityDescriptor{}, searchIdentityInvalid("DUPLICATE_CONSTRUCTION_OBLIGATION", "constructed.obligations", "construction returned more than one obligation for a goal constraint")
		}
		obligationByID[obligation.ConstraintID] = obligation
	}
	descriptor.Fragments = make([]searchIdentityFragmentMap, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		obligation, exists := obligationByID[constraint.ID]
		if !exists || obligation.Kind != constraint.Kind {
			return searchIdentityDescriptor{}, searchIdentityInvalid("CONSTRUCTION_OBLIGATION_MISMATCH", "constructed.obligations", "construction obligations do not match the concrete goal constraints")
		}
		delete(obligationByID, constraint.ID)
		if len(obligation.Fragments) > MaxGoalAtoms {
			return searchIdentityDescriptor{}, searchIdentityResource("constructed.obligations.fragments", "fragment count exceeds the goal atom bound")
		}
		fragments := append([]string{}, obligation.Fragments...)
		for _, fragment := range fragments {
			if len(fragment) > 4096 {
				return searchIdentityDescriptor{}, searchIdentityResource("constructed.obligations.fragments", "fragment path exceeds its bounded size")
			}
		}
		descriptor.Fragments = append(descriptor.Fragments, searchIdentityFragmentMap{ConstraintID: constraint.ID, Kind: constraint.Kind, Fragments: fragments})
	}
	if len(obligationByID) != 0 {
		return searchIdentityDescriptor{}, searchIdentityInvalid("EXTRA_CONSTRUCTION_OBLIGATION", "constructed.obligations", "construction returned an obligation not present in the concrete goal")
	}
	return descriptor, nil
}

func cloneIdentityTerm(term *algebra.Term, limits algebra.Limits) (*algebra.Term, error) {
	count := 0
	var clone func(*algebra.Term, int) (*algebra.Term, error)
	clone = func(source *algebra.Term, depth int) (*algebra.Term, error) {
		if source == nil {
			return nil, nil
		}
		if depth > limits.MaxDepth {
			return nil, searchIdentityResource("expression", "canonical expression exceeds the depth bound")
		}
		count++
		if count > limits.MaxNodes {
			return nil, searchIdentityResource("expression", "canonical expression exceeds the node bound")
		}
		result := &algebra.Term{Kind: source.Kind, ID: source.ID, Symbol: source.Symbol, Quantifier: source.Quantifier, Modality: source.Modality, ModalProfile: source.ModalProfile}
		if source.Count != nil {
			countCopy := *source.Count
			result.Count = &countCopy
		}
		if source.Type != nil {
			typ, err := cloneIdentityType(*source.Type, limits)
			if err != nil {
				return nil, err
			}
			result.Type = &typ
		}
		if source.Args != nil {
			result.Args = make([]algebra.Argument, len(source.Args))
			for i, argument := range source.Args {
				value, err := clone(argument.Value, depth+1)
				if err != nil {
					return nil, err
				}
				result.Args[i] = algebra.Argument{Role: argument.Role, Value: value}
			}
		}
		if source.Operands != nil {
			result.Operands = make([]*algebra.Term, len(source.Operands))
			for i, operand := range source.Operands {
				value, err := clone(operand, depth+1)
				if err != nil {
					return nil, err
				}
				result.Operands[i] = value
			}
		}
		var err error
		result.Body, err = clone(source.Body, depth+1)
		if err != nil {
			return nil, err
		}
		result.Function, err = clone(source.Function, depth+1)
		if err != nil {
			return nil, err
		}
		result.Argument, err = clone(source.Argument, depth+1)
		if err != nil {
			return nil, err
		}
		result.Restrictor, err = clone(source.Restrictor, depth+1)
		if err != nil {
			return nil, err
		}
		if source.Answers != nil {
			result.Answers = make([]algebra.Declaration, len(source.Answers))
			for i, answer := range source.Answers {
				result.Answers[i] = answer
				typ, cloneErr := cloneIdentityType(answer.Type, limits)
				if cloneErr != nil {
					return nil, cloneErr
				}
				result.Answers[i].Type = typ
			}
		}
		return result, nil
	}
	return clone(term, 1)
}

func cloneIdentityType(source algebra.Type, limits algebra.Limits) (algebra.Type, error) {
	visits := 0
	active := make(map[*algebra.Type]bool)
	var clone func(*algebra.Type, int) (*algebra.Type, error)
	clone = func(value *algebra.Type, depth int) (*algebra.Type, error) {
		if value == nil {
			return nil, nil
		}
		if depth > limits.MaxTypeDepth {
			return nil, searchIdentityResource("type", "canonical expression type exceeds the depth bound")
		}
		if active[value] {
			return nil, searchIdentityInvalid("CYCLIC_TYPE", "type", "canonical expression contains a cyclic type")
		}
		visits++
		if visits > limits.MaxNodes*limits.MaxTypeDepth {
			return nil, searchIdentityResource("type", "canonical expression type exceeds the node bound")
		}
		active[value] = true
		copyValue := &algebra.Type{Kind: value.Kind, SortID: value.SortID}
		from, err := clone(value.From, depth+1)
		if err == nil {
			copyValue.From = from
			var to *algebra.Type
			to, err = clone(value.To, depth+1)
			copyValue.To = to
		}
		delete(active, value)
		return copyValue, err
	}
	copyType, err := clone(&source, 1)
	if err != nil {
		return algebra.Type{}, err
	}
	return *copyType, nil
}

func validateIdentityReferences(environmentRefs, goalRefs []algebra.Declaration) error {
	if len(environmentRefs) != len(goalRefs) {
		return searchIdentityInvalid("SEARCH_REFERENCE_CONTEXT_MISMATCH", "environment.document.references", "rewrite environment must contain exactly the concrete goal references")
	}
	left := append([]algebra.Declaration{}, environmentRefs...)
	right := append([]algebra.Declaration{}, goalRefs...)
	sort.Slice(left, func(i, j int) bool { return left[i].ID < left[j].ID })
	sort.Slice(right, func(i, j int) bool { return right[i].ID < right[j].ID })
	for i := range left {
		if left[i].ID != right[i].ID || !algebra.TypeEqual(left[i].Type, right[i].Type) || left[i].TemporalFrame != right[i].TemporalFrame || left[i].ReferenceTimeID != right[i].ReferenceTimeID || left[i].Multiplicity != right[i].Multiplicity {
			return searchIdentityInvalid("SEARCH_REFERENCE_CONTEXT_MISMATCH", fmt.Sprintf("environment.document.references[%d]", i), "rewrite environment references must exactly match the concrete goal")
		}
	}
	return nil
}

func searchIdentityInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func searchIdentityResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "SEARCH_IDENTITY_LIMIT", Path: path, Message: message}
}

// Search enumerates the finite profile's explicit choices, constructs each
// complete candidate, and keeps only candidates with independently replayed
// alpha identities. It does not evaluate facts or prove the goal.
func Search(parent context.Context, primitiveContext algebra.Document, rewriteEnv algebra.RewriteEnvironment, goal GoalSpec, requested SearchLimits) (SearchResult, error) {
	budget, cancel, err := newSearchBudget(parent, requested)
	if err != nil {
		return SearchResult{}, err
	}
	defer cancel()
	result := SearchResult{
		SchemaVersion: searchResultSchema, Status: status.OutcomeIncomplete,
		Outcome: SearchOutcomeIncomplete, Profile: SearchSpaceProfile,
		Goal: goal, GoalFingerprint: "", RegistrySnapshotID: rewriteEnv.SnapshotID,
		ContextID: goal.ContextID, Limits: budget.limits,
		Space: emptySearchSpaceDescriptor(budget.limits), Candidates: []SearchCandidate{},
		Complete: false, UniqueWithinDeclaredSpace: false,
		Trace: []SearchTraceEntry{}, ExhaustedBounds: []SearchStopReason{},
		ValidationScope: "finite declared positive-facts path profile; independently verified goal constraints and operation replay plus alpha identity; not world truth, speaker intent, or language approval",
	}

	// Validate the complete user request before fingerprint serialization or
	// signature-index construction. Candidate-specific type/operator fit is
	// deliberately deferred to Construct for each explicit assignment.
	_, err = validateSearchGoal(primitiveContext, goal)
	if err != nil {
		return SearchResult{}, err
	}
	goalLimits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return SearchResult{}, err
	}
	if err := budget.checkDeadlineCancel(); !err {
		return finalizeSearchResult(result, nil, budget, false, SearchOutcomeIncomplete)
	}
	goalFingerprint, err := searchGoalFingerprint(goal, rewriteEnv.SnapshotID)
	if err != nil {
		return SearchResult{}, err
	}
	result.GoalFingerprint = goalFingerprint

	if err := validateSearchEnvironment(primitiveContext, rewriteEnv, goal); err != nil {
		return SearchResult{}, err
	}
	phaseWork := len(primitiveContext.Sorts) + len(primitiveContext.Signatures) + len(goal.Ports) + len(goal.References) + len(goal.Constraints)
	for _, signature := range primitiveContext.Signatures {
		phaseWork += len(signature.Ports)
	}
	if !chargeSearchWork(budget, phaseWork) {
		return finalizeSearchResult(result, nil, budget, false, SearchOutcomeIncomplete)
	}
	if !budget.checkDeadlineCancel() {
		return finalizeSearchResult(result, nil, budget, false, SearchOutcomeIncomplete)
	}
	index, err := NewSignatureIndex(primitiveContext)
	if err != nil {
		return SearchResult{}, err
	}
	if !budget.checkDeadlineCancel() {
		return finalizeSearchResult(result, nil, budget, false, SearchOutcomeIncomplete)
	}
	space, err := newFiniteCandidateSpace(primitiveContext, index, goal, budget)
	if err != nil {
		return SearchResult{}, err
	}
	result.Space = space.descriptorSnapshot()
	if space.stopReason != "" || budget.stop() != "" {
		return finalizeSearchResult(result, space, budget, false, SearchOutcomeIncomplete)
	}
	if space.blocked {
		return finalizeSearchResult(result, space, budget, false, SearchOutcomeUnresolvedGoal)
	}

	identityBuckets := make(map[string][]int)
	identities := make([]searchCandidateIdentity, 0)
	attempt := 0
	for space.hasNext() {
		if !budget.checkDeadlineCancel() {
			break
		}
		if len(result.Candidates) >= budget.limits.MaxCandidates {
			budget.stopReason = SearchStopMaxCandidates
			break
		}
		if !budget.consumeExpansion() {
			break
		}
		attempt++
		concreteGoal, choices, ok, nextErr := space.next()
		if nextErr != nil {
			return SearchResult{}, nextErr
		}
		if !ok {
			if space.stopReason == "" && budget.stop() == "" {
				return SearchResult{}, searchIdentityInvalid("SEARCH_CURSOR_STALLED", "space", "candidate cursor stopped without exhaustion or an operational bound")
			}
			break
		}
		outsideBound := false
		for _, choice := range choices {
			if choice.OutsidePathBound {
				outsideBound = true
				break
			}
		}
		if outsideBound {
			budget.stopReason = SearchStopMaxPathLength
			appendSearchTrace(budget, &result, SearchTraceEntry{Attempt: attempt, Stage: "enumerate", Code: string(SearchStopMaxPathLength), Status: status.OutcomeResourceLimit})
			break
		}

		if !reserveSearchWork(budget, searchConstructionReserve) {
			break
		}
		constructed, constructErr := Construct(primitiveContext, concreteGoal)
		if constructErr != nil {
			if isSearchPruneError(constructErr) {
				result.Rejected++
				appendSearchErrorTrace(budget, &result, attempt, "construct_rejected", constructErr, choices)
				continue
			}
			if isResourceError(constructErr) {
				budget.stopReason = constructionStopReason(constructErr)
				appendSearchErrorTrace(budget, &result, attempt, "construct_stopped", constructErr, choices)
				break
			}
			return SearchResult{}, constructErr
		}
		if constructed.Status != status.OutcomeOK {
			if constructed.Status == status.OutcomeIncomplete {
				budget.stopReason = SearchStopConstruction
				appendSearchTrace(budget, &result, SearchTraceEntry{Attempt: attempt, Stage: "construct", Code: string(SearchStopConstruction), Status: status.OutcomeIncomplete})
				break
			}
			return SearchResult{}, searchIdentityInvalid("UNEXPECTED_CONSTRUCTION_STATUS", "constructed.status", "construction returned a status outside the accepted complete/incomplete outcomes")
		}
		constructionWork := 0
		if constructed.Checked != nil {
			constructionWork += constructed.Checked.NodeCount
		}
		constructionWork += len(constructed.Trace)
		if constructionWork < 0 || constructionWork > searchConstructionReserve {
			return SearchResult{}, searchIdentityInvalid("CONSTRUCTION_WORK_OVERREPORT", "constructed", "constructor work exceeded its finite reservation")
		}
		budget.usage.WorkUnits -= searchConstructionReserve - constructionWork
		if !budget.checkDeadlineCancel() {
			break
		}

		identity, complete, identityErr := certifySearchCandidate(rewriteEnv, concreteGoal, choices, constructed, budget)
		if identityErr != nil {
			return SearchResult{}, identityErr
		}
		if !complete {
			if budget.stop() == "" {
				budget.stopReason = SearchStopAlphaIdentity
			}
			break
		}
		if !budget.checkDeadlineCancel() {
			break
		}
		remainingWork := budget.remainingWork()
		if remainingWork <= 0 {
			budget.stopReason = SearchStopGoalVerification
			break
		}
		producerWorkLimit := minInt(remainingWork, MaxGoalVerificationMaxWorkNodes)
		producerCertificateLimit := minInt(budget.limits.MaxCandidateBytes, MaxGoalVerificationMaxCertificateBytes)
		producerLimits := GoalVerificationLimits{
			MaxCertificateBytes: producerCertificateLimit,
			MaxReplaySteps:      minInt(goalLimits.MaxSteps, MaxGoalVerificationMaxReplaySteps),
			MaxWorkNodes:        producerWorkLimit,
			TermLimits:          goalLimits.TermLimits,
		}
		goalCertificate, producerStats, producerErr := CreateSynthesisCertificate(
			primitiveContext, rewriteEnv, goal, concreteGoal, choices,
			CertificateOriginSearch, constructed, producerLimits,
		)
		if chargeErr := chargeCompletedGoalVerificationWork(budget, producerStats.WorkNodes, producerWorkLimit); chargeErr != nil {
			return SearchResult{}, chargeErr
		}
		if !budget.checkDeadlineCancel() {
			break
		}
		if producerErr != nil {
			if isResourceError(producerErr) || isIncompleteError(producerErr) {
				budget.stopReason = SearchStopGoalVerification
				appendSearchErrorTrace(budget, &result, attempt, string(SearchStopGoalVerification), producerErr, choices)
				break
			}
			return SearchResult{}, producerErr
		}
		remainingWork = budget.remainingWork()
		if remainingWork <= 0 {
			budget.stopReason = SearchStopGoalVerification
			break
		}
		verifierWorkLimit := minInt(remainingWork, MaxGoalVerificationMaxWorkNodes)
		verifierLimits := GoalVerificationLimits{
			MaxCertificateBytes: producerCertificateLimit,
			MaxReplaySteps:      minInt(goalLimits.MaxSteps, MaxGoalVerificationMaxReplaySteps),
			MaxWorkNodes:        verifierWorkLimit,
			TermLimits:          goalLimits.TermLimits,
		}
		verification, verifyErr := VerifyGoalCertificate(primitiveContext, rewriteEnv, goal, *constructed.CanonicalDocument, goalCertificate, verifierLimits)
		if chargeErr := chargeCompletedGoalVerificationWork(budget, verification.Statistics.WorkNodes, verifierWorkLimit); chargeErr != nil {
			return SearchResult{}, chargeErr
		}
		if !budget.checkDeadlineCancel() {
			break
		}
		if verifyErr != nil {
			if isResourceError(verifyErr) || isIncompleteError(verifyErr) {
				budget.stopReason = SearchStopGoalVerification
				appendSearchErrorTrace(budget, &result, attempt, string(SearchStopGoalVerification), verifyErr, choices)
				break
			}
			return SearchResult{}, verifyErr
		}
		if verification.Status == status.OutcomeResourceLimit || verification.Status == status.OutcomeIncomplete || verification.Disposition == GoalVerificationDispositionUndetermined {
			budget.stopReason = SearchStopGoalVerification
			appendSearchTrace(budget, &result, SearchTraceEntry{Attempt: attempt, Stage: "goal_verification", Code: string(SearchStopGoalVerification), Status: verification.Status})
			break
		}
		if !fullyValidatedGoal(verification) {
			return SearchResult{}, searchIdentityInvalid("GOAL_CERTIFICATE_REJECTED", "candidate.verification", "candidate did not pass independent goal and operation verification")
		}
		constructed.Certificate = &goalCertificate
		constructed.Verification = &verification
		constructed.FullyValidated = true
		duplicateIndex := -1
		for _, priorIndex := range identityBuckets[identity.key] {
			same, compareErr := identity.sameCandidate(identities[priorIndex], primitiveContext.Limits)
			if compareErr != nil {
				return SearchResult{}, compareErr
			}
			if same {
				duplicateIndex = priorIndex
				break
			}
		}
		if duplicateIndex >= 0 {
			result.EquivalentDuplicates++
			appendSearchTrace(budget, &result, SearchTraceEntry{
				Attempt: attempt, Stage: "deduplicate", Code: "ALPHA_EQUIVALENT_CANDIDATE",
				CandidateID: result.Candidates[duplicateIndex].ID, ChoiceKey: compactChoiceKey(choices), Status: status.OutcomeOK,
			})
			continue
		}

		candidateID := fmt.Sprintf("c%d", len(result.Candidates)+1)
		candidate := buildSearchCandidate(candidateID, goalFingerprint, rewriteEnv.SnapshotID, goal.ContextID, concreteGoal, choices, constructed, identity)
		candidateBytes, marshalErr := json.Marshal(candidate)
		if marshalErr != nil {
			if isResourceError(marshalErr) {
				budget.stopReason = SearchStopMaxCandidateBytes
				appendSearchErrorTrace(budget, &result, attempt, "candidate_serialization_stopped", marshalErr, choices)
				break
			}
			return SearchResult{}, searchIdentityInvalid("SEARCH_CANDIDATE_ENCODING", "candidate", marshalErr.Error())
		}
		if !appendSearchCandidate(budget, &result, candidate, len(candidateBytes)) {
			break
		}
		identity.normalization = algebra.NormalizationResult{}
		identity.verification = algebra.VerificationResult{}
		identityBuckets[identity.key] = append(identityBuckets[identity.key], len(identities))
		identities = append(identities, identity)
		appendSearchTrace(budget, &result, SearchTraceEntry{
			Attempt: attempt, Stage: "candidate", Code: "CANDIDATE_ACCEPTED", CandidateID: candidateID,
			ChoiceKey: compactChoiceKey(choices), Status: status.OutcomeOK,
		})
	}

	if space.stopReason != "" && budget.stopReason == "" {
		budget.stopReason = space.stopReason
	}
	if budget.stop() != "" {
		return finalizeSearchResult(result, space, budget, false, SearchOutcomeIncomplete)
	}
	if space.blocked {
		return finalizeSearchResult(result, space, budget, false, SearchOutcomeUnresolvedGoal)
	}
	if space.hasNext() {
		// A cursor with remaining work that the loop did not consume is not
		// exhaustion; retain the verified prefix and report a bounded stop.
		budget.stopReason = SearchStopAlphaIdentity
		return finalizeSearchResult(result, space, budget, false, SearchOutcomeIncomplete)
	}
	if len(result.Candidates) == 0 {
		return finalizeSearchResult(result, space, budget, true, SearchOutcomeNoSolutionWithinBounds)
	}
	return finalizeSearchResult(result, space, budget, true, SearchOutcomeCandidates)
}

func emptySearchSpaceDescriptor(limits SearchLimits) SearchSpaceDescriptor {
	return SearchSpaceDescriptor{
		Profile: SearchSpaceProfile, PathPolicy: "explicit_counts",
		PathLengthChoices: []int{}, ExcludedExplicitPathLengths: []int{},
		Axes: []SearchAxisDescriptor{}, SourceConstraintIDs: []string{},
		InputUnresolved: []Unresolved{}, DomainRejections: []SearchDomainRejection{},
		BranchDepthMetric: "one layer per ordered choice axis; singleton axes count",
		MaxPathLength:     limits.MaxPathLength, MaxBranchDepth: limits.MaxBranchDepth,
	}
}

func finalizeSearchResult(result SearchResult, space *finiteCandidateSpace, budget *searchBudget, complete bool, outcome string) (SearchResult, error) {
	if space != nil {
		result.Space = space.descriptorSnapshot()
	}
	result.Usage = budget.usage
	result.Complete = complete
	result.UniqueWithinDeclaredSpace = complete && len(result.Candidates) == 1
	result.Outcome = outcome
	result.ValidationScope = SearchValidationScope
	result.Trace = nonNilSearchTrace(result.Trace)
	result.Candidates = nonNilSearchCandidates(result.Candidates)
	result.ExhaustedBounds = nonNilSearchStops(result.ExhaustedBounds)
	if outcome == SearchOutcomeUnresolvedGoal {
		result.Status = status.OutcomeIncomplete
		if budget.stopReason == "" {
			budget.stopReason = SearchStopUnresolved
		}
	} else if complete {
		result.Status = status.OutcomeOK
	} else if budget.stopReason == SearchStopUnresolved || budget.stopReason == SearchStopConstruction {
		result.Status = status.OutcomeIncomplete
	} else if budget.stopReason != "" {
		result.Status = status.OutcomeResourceLimit
	} else {
		result.Status = status.OutcomeIncomplete
	}
	if budget.stopReason != "" {
		result.StopReason = budget.stopReason
		result.Space.StopReason = budget.stopReason
		if budget.stopReason != SearchStopUnresolved && budget.stopReason != SearchStopConstruction {
			result.ExhaustedBounds = append(result.ExhaustedBounds, budget.stopReason)
		}
	}
	return result, nil
}

func searchGoalFingerprint(goal GoalSpec, snapshotID string) (string, error) {
	canonical := goal
	limits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return "", err
	}
	canonical.Limits = limits
	canonical.References = append([]algebra.Declaration{}, goal.References...)
	sort.Slice(canonical.References, func(i, j int) bool { return canonical.References[i].ID < canonical.References[j].ID })
	canonical.ProjectPorts = append([]string{}, goal.ProjectPorts...)
	sort.Strings(canonical.ProjectPorts)
	canonical.AllowedOperators = append([]string{}, goal.AllowedOperators...)
	sort.Strings(canonical.AllowedOperators)
	canonical.Provenance = append(canonical.Provenance[:0:0], goal.Provenance...)
	sort.Slice(canonical.Provenance, func(i, j int) bool { return canonical.Provenance[i].ID < canonical.Provenance[j].ID })
	canonical.Constraints = append([]Constraint{}, goal.Constraints...)
	for i := range canonical.Constraints {
		canonical.Constraints[i].AllowedAtomCounts = append([]int{}, goal.Constraints[i].AllowedAtomCounts...)
		sort.Ints(canonical.Constraints[i].AllowedAtomCounts)
	}
	payload, err := json.Marshal(struct {
		Version   string   `json:"version"`
		Snapshot  string   `json:"snapshot_id"`
		ContextID string   `json:"context_id"`
		Goal      GoalSpec `json:"goal"`
	}{Version: "dse-goal-fingerprint-v1", Snapshot: snapshotID, ContextID: goal.ContextID, Goal: canonical})
	if err != nil {
		return "", searchIdentityInvalid("GOAL_FINGERPRINT_ENCODING", "goal", err.Error())
	}
	if len(payload) > algebra.MaxDocumentBytes {
		return "", searchIdentityResource("goal", "canonical goal fingerprint payload exceeds 1 MiB")
	}
	digest := sha256.Sum256(append([]byte("dse-goal-fingerprint-v1\x00"), payload...))
	return hex.EncodeToString(digest[:]), nil
}

func validateSearchEnvironment(primitiveContext algebra.Document, rewriteEnv algebra.RewriteEnvironment, goal GoalSpec) error {
	if err := algebra.ValidateRewriteEnvironment(rewriteEnv); err != nil {
		return err
	}
	if err := validateIdentityReferences(rewriteEnv.Document.References, goal.References); err != nil {
		return err
	}
	if (rewriteEnv.Document.RequestContext == nil) != (goal.RequestContext == nil) {
		return searchIdentityInvalid("SEARCH_REQUEST_CONTEXT_MISMATCH", "environment.document.request_context", "rewrite environment must carry exactly the goal request context")
	}
	if goal.RequestContext != nil {
		environmentContext, contextErr := algebra.RequestContextFingerprint(*rewriteEnv.Document.RequestContext, rewriteEnv.Document)
		if contextErr != nil {
			return contextErr
		}
		goalContext, contextErr := goalContextFingerprint(primitiveContext, goal)
		if contextErr != nil {
			return contextErr
		}
		if environmentContext != goalContext {
			return searchIdentityInvalid("SEARCH_REQUEST_CONTEXT_MISMATCH", "environment.document.request_context", "rewrite environment context payload must match the goal")
		}
	}
	primitiveSorts := make(map[string]struct{}, len(primitiveContext.Sorts))
	environmentSorts := make(map[string]struct{}, len(rewriteEnv.Document.Sorts))
	for _, sort := range primitiveContext.Sorts {
		primitiveSorts[sort.ID] = struct{}{}
	}
	for _, sort := range rewriteEnv.Document.Sorts {
		environmentSorts[sort.ID] = struct{}{}
	}
	for sortID := range primitiveSorts {
		if _, ok := environmentSorts[sortID]; !ok {
			return searchIdentityInvalid("SEARCH_REWRITE_CONTEXT_MISMATCH", "environment.document.sorts", "rewrite environment must include the primitive context sorts")
		}
	}
	for _, signature := range primitiveContext.Signatures {
		var match *algebra.Signature
		for i := range rewriteEnv.Document.Signatures {
			if rewriteEnv.Document.Signatures[i].ID == signature.ID {
				match = &rewriteEnv.Document.Signatures[i]
				break
			}
		}
		if match == nil || len(match.Ports) != len(signature.Ports) {
			return searchIdentityInvalid("SEARCH_REWRITE_CONTEXT_MISMATCH", "environment.document.signatures", "rewrite environment must include every exact primitive signature")
		}
		for i, port := range signature.Ports {
			if match.Ports[i].Role != port.Role || !algebra.TypeEqual(match.Ports[i].Type, port.Type) {
				return searchIdentityInvalid("SEARCH_REWRITE_CONTEXT_MISMATCH", "environment.document.signatures", "rewrite environment primitive signature roles and types must match the indexed context")
			}
		}
	}
	goalLimits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return err
	}
	primitiveLimits, err := algebra.NormalizeLimits(primitiveContext.Limits)
	if err != nil {
		return err
	}
	environmentLimits, err := algebra.NormalizeLimits(rewriteEnv.Document.Limits)
	if err != nil {
		return err
	}
	if primitiveLimits != goalLimits.TermLimits || environmentLimits != goalLimits.TermLimits {
		return searchIdentityInvalid("SEARCH_TERM_LIMIT_MISMATCH", "environment.document.limits", "primitive, goal, and rewrite environments must use the same normalized term limits")
	}
	return nil
}

func chargeSearchWork(budget *searchBudget, work int) bool {
	ok, err := budget.consumeWork(work)
	if err != nil {
		return false
	}
	return ok
}

func reserveSearchWork(budget *searchBudget, reserve int) bool {
	if reserve < 0 || reserve > budget.remainingWork() || !budget.checkDeadlineCancel() {
		if budget.stopReason == "" {
			budget.stopReason = SearchStopMaxWorkUnits
		}
		return false
	}
	budget.usage.WorkUnits += reserve
	return true
}

func appendSearchCandidate(budget *searchBudget, result *SearchResult, candidate SearchCandidate, size int) bool {
	if !budget.checkDeadlineCancel() {
		return false
	}
	if budget.usage.Candidates >= budget.limits.MaxCandidates {
		budget.stopReason = SearchStopMaxCandidates
		return false
	}
	if size < 0 || size > budget.limits.MaxCandidateBytes-budget.usage.CandidateBytes {
		budget.stopReason = SearchStopMaxCandidateBytes
		return false
	}
	if size > algebra.MaxRewriteEnvelopeBytes {
		budget.stopReason = SearchStopMaxCandidateBytes
		return false
	}
	// Candidate count and serialized bytes are committed together only after
	// all checks, so stopped searches never report a phantom candidate.
	budget.usage.Candidates++
	budget.usage.CandidateBytes += size
	result.Candidates = append(result.Candidates, candidate)
	return true
}

func buildSearchCandidate(id, goalFingerprint, snapshotID, contextID string, concreteGoal GoalSpec, choices []ChoiceAssignment, construction Result, identity searchCandidateIdentity) SearchCandidate {
	identityProfile := algebra.RewriteProfile{ID: searchAlphaProfileID, Rules: []algebra.RewriteRule{algebra.RewriteAlphaV1}}
	certificate := DraftSynthesisCertificate{
		SchemaVersion: 1, Status: DraftGoalCertificateStatus, GoalVerified: false,
		GoalFingerprint: goalFingerprint, RegistrySnapshotID: snapshotID,
		ContextID: contextID, CandidateKey: identity.key,
		Choices:         append([]ChoiceAssignment{}, choices...),
		OperatorTrace:   append([]algebra.BuildTrace{}, construction.Trace...),
		Obligations:     append([]GoalObligation{}, construction.Obligations...),
		IdentityProfile: identityProfile, IdentityCertificate: identity.normalization.Certificate,
		IdentityVerification: identity.verification,
	}
	request := &algebra.ArtifactExportRequest{
		SchemaVersion: 1, Document: *construction.CanonicalDocument,
		Profile:   algebra.RewriteProfile{ID: searchArtifactProfileID, Rules: []algebra.RewriteRule{}},
		ContextID: contextID, ContextDependencies: []algebra.ContextDependency{},
		Provenance: append([]provenance.SourceRef{}, concreteGoal.Provenance...),
	}
	return SearchCandidate{
		ID: id, Key: identity.key, ConcreteGoal: concreteGoal,
		Choices: append([]ChoiceAssignment{}, choices...), Construction: construction,
		DraftCertificate: certificate, ExportRequest: request,
	}
}

func appendSearchTrace(budget *searchBudget, result *SearchResult, entry SearchTraceEntry) {
	if budget == nil || result == nil || !budget.retainTrace() {
		return
	}
	entry.Code = boundedSearchText(entry.Code, 256)
	entry.ConstraintID = boundedSearchText(entry.ConstraintID, 256)
	entry.InputPath = boundedSearchText(entry.InputPath, 256)
	entry.Message = boundedSearchText(entry.Message, 256)
	entry.CandidateID = boundedSearchText(entry.CandidateID, 64)
	entry.ChoiceKey = boundedSearchText(entry.ChoiceKey, 256)
	result.Trace = append(result.Trace, entry)
}

func appendSearchErrorTrace(budget *searchBudget, result *SearchResult, attempt int, stage string, err error, choices []ChoiceAssignment) {
	entry := SearchTraceEntry{Attempt: attempt, Stage: stage, ChoiceKey: compactChoiceKey(choices)}
	var typed *algebra.Error
	if errors.As(err, &typed) {
		entry.Code, entry.InputPath, entry.Message, entry.Status = typed.Code, typed.Path, typed.Message, typed.Status
	} else {
		entry.Code, entry.Message = "CONSTRUCTION_ERROR", err.Error()
		entry.Status = status.OutcomeError
	}
	appendSearchTrace(budget, result, entry)
}

func appendSearchErrorTraceWithCandidate(budget *searchBudget, result *SearchResult, attempt int, stage string, err error, choices []ChoiceAssignment) {
	appendSearchErrorTrace(budget, result, attempt, stage, err, choices)
}

func compactChoiceKey(choices []ChoiceAssignment) string {
	data, err := json.Marshal(choices)
	if err != nil {
		return "choices-unavailable"
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:8])
}

func boundedSearchText(value string, maximum int) string {
	if len(value) <= maximum && utf8.ValidString(value) {
		return value
	}
	if !utf8.ValidString(value) {
		value = strings.ToValidUTF8(value, "�")
	}
	if maximum <= 3 {
		return value[:maximum]
	}
	limit := maximum - 3
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit] + "..."
}

func isSearchPruneError(err error) bool {
	var typed *algebra.Error
	if !errors.As(err, &typed) {
		return false
	}
	switch typed.Code {
	case "PATH_LINK_TYPE_MISMATCH", "GOAL_ROLE_TYPE_MISMATCH", "OPERATOR_NOT_ALLOWED":
		return true
	default:
		return false
	}
}

func isResourceError(err error) bool {
	var typed *algebra.Error
	return errors.As(err, &typed) && typed.Status == status.OutcomeResourceLimit
}

func constructionStopReason(err error) SearchStopReason {
	var typed *algebra.Error
	if !errors.As(err, &typed) {
		return SearchStopConstructionLimit
	}
	switch typed.Path {
	case "limits.max_atoms":
		return SearchStopMaxAtoms
	case "limits.max_steps":
		return SearchStopMaxSteps
	default:
		message := strings.ToLower(typed.Message)
		if strings.Contains(message, "atom") && strings.Contains(message, "budget") {
			return SearchStopMaxAtoms
		}
		return SearchStopConstructionLimit
	}
}

func nonNilSearchCandidates(values []SearchCandidate) []SearchCandidate {
	if values == nil {
		return []SearchCandidate{}
	}
	return values
}

func nonNilSearchTrace(values []SearchTraceEntry) []SearchTraceEntry {
	if values == nil {
		return []SearchTraceEntry{}
	}
	return values
}

func nonNilSearchStops(values []SearchStopReason) []SearchStopReason {
	if values == nil {
		return []SearchStopReason{}
	}
	return values
}

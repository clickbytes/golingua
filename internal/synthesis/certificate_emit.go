package synthesis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

const maxEmittedCertificateBytes = 8 << 20

// CreateSynthesisCertificate emits bounded producer claims for a checked
// construction. The returned certificate is still unverified: its operations
// and structural obligations must be checked independently by the verifier.
func CreateSynthesisCertificate(primitiveContext algebra.Document, environment algebra.RewriteEnvironment, original, concrete GoalSpec, choices []ChoiceAssignment, origin string, constructed Result, requested GoalVerificationLimits) (SynthesisCertificate, GoalVerificationStatistics, error) {
	limits, err := NormalizeGoalVerificationLimits(requested)
	if err != nil {
		return SynthesisCertificate{}, GoalVerificationStatistics{}, err
	}
	budget := &certificateProducerBudget{limits: limits}
	certificate, err := createSynthesisCertificate(primitiveContext, environment, original, concrete, choices, origin, constructed, budget)
	return certificate, budget.statistics, err
}

type certificateProducerBudget struct {
	limits     GoalVerificationLimits
	statistics GoalVerificationStatistics
}

func (budget *certificateProducerBudget) charge(work int, path string) error {
	if work < 0 || work > budget.limits.MaxWorkNodes-budget.statistics.WorkNodes {
		budget.statistics.WorkNodes = budget.limits.MaxWorkNodes
		return certEmitResource(path, "certificate production exceeds the bounded work-node limit")
	}
	budget.statistics.WorkNodes += work
	return nil
}

func (budget *certificateProducerBudget) remaining() int {
	return budget.limits.MaxWorkNodes - budget.statistics.WorkNodes
}

func (budget *certificateProducerBudget) markTermResource(termNodeLimit int) {
	used := minInt(budget.remaining(), termNodeLimit)
	if used > 0 {
		budget.statistics.WorkNodes += used
	}
}

func createSynthesisCertificate(primitiveContext algebra.Document, environment algebra.RewriteEnvironment, original, concrete GoalSpec, choices []ChoiceAssignment, origin string, constructed Result, budget *certificateProducerBudget) (SynthesisCertificate, error) {
	if origin != CertificateOriginConstruct && origin != CertificateOriginSearch {
		return SynthesisCertificate{}, certEmitUnsupported("UNSUPPORTED_CERTIFICATE_ORIGIN", "origin", "origin must be construct or search")
	}
	if err := ValidateGoalRefinement(primitiveContext, original, concrete, choices, origin); err != nil {
		return SynthesisCertificate{}, err
	}
	if err := validateSearchEnvironment(primitiveContext, environment, concrete); err != nil {
		return SynthesisCertificate{}, err
	}
	if constructed.Status != status.OutcomeOK || constructed.Checked == nil || constructed.Checked.Status != status.OutcomeOK ||
		constructed.Expression == nil || constructed.CanonicalDocument == nil {
		return SynthesisCertificate{}, certEmitIncomplete("CONSTRUCTION_NOT_CHECKED", "constructed", "certificate production requires an OK checked construction and canonical document")
	}
	if len(constructed.Trace) == 0 || len(constructed.Trace) > MaxGoalSteps {
		return SynthesisCertificate{}, certEmitInvalid("INVALID_CONSTRUCTION_TRACE", "constructed.trace", "a bounded nonempty construction trace is required")
	}
	goalLimits, err := NormalizeGoalLimits(concrete.Limits)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	contextLimits, err := algebra.NormalizeLimits(primitiveContext.Limits)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	candidateLimits, err := algebra.NormalizeLimits(constructed.CanonicalDocument.Limits)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	effectiveTermLimits := minimumAlgebraLimits(goalLimits.TermLimits, budget.limits.TermLimits, contextLimits, candidateLimits)
	if len(constructed.Trace) > goalLimits.MaxSteps {
		return SynthesisCertificate{}, certEmitResource("constructed.trace", "construction trace exceeds the goal step limit")
	}

	canonicalExpressionEqual, err := algebra.RewriteTermEqual(constructed.Expression, constructed.CanonicalDocument.Expression, effectiveTermLimits)
	if err != nil {
		if isCertificateResourceError(err) {
			budget.markTermResource(effectiveTermLimits.MaxNodes)
		}
		return SynthesisCertificate{}, err
	}
	if !canonicalExpressionEqual {
		return SynthesisCertificate{}, certEmitInvalid("CONSTRUCTION_DOCUMENT_MISMATCH", "constructed.canonical_document.expression", "canonical document must contain the constructed expression")
	}
	checkedDocument := *constructed.CanonicalDocument
	checkedDocument.Limits = effectiveTermLimits
	checked, err := algebra.Check(checkedDocument)
	if err != nil {
		if isCertificateResourceError(err) {
			budget.markTermResource(effectiveTermLimits.MaxNodes)
		}
		return SynthesisCertificate{}, err
	}
	if checked.Status != status.OutcomeOK || !algebra.TypeEqual(checked.Type, concrete.ExpectedType) {
		return SynthesisCertificate{}, certEmitInvalid("CONSTRUCTION_TYPE_MISMATCH", "constructed.canonical_document.expression", "canonical candidate must be complete at the concrete goal type")
	}
	if err := budget.charge(checked.NodeCount, "constructed.canonical_document.expression"); err != nil {
		return SynthesisCertificate{}, err
	}
	if budget.remaining() < checked.NodeCount {
		return SynthesisCertificate{}, certEmitResource("candidate_fingerprint", "insufficient work budget for bounded candidate encoding")
	}
	candidateBytes, err := algebra.Encode(*constructed.CanonicalDocument)
	if err != nil {
		if isCertificateResourceError(err) {
			budget.markTermResource(effectiveTermLimits.MaxNodes)
		}
		return SynthesisCertificate{}, err
	}
	if err := budget.charge(checked.NodeCount, "candidate_fingerprint"); err != nil {
		return SynthesisCertificate{}, err
	}
	budget.statistics.VisitedNodes += checked.NodeCount * 2
	if budget.remaining() <= 0 {
		return SynthesisCertificate{}, certEmitResource("constraint_check", "no work budget remains for independent goal checking")
	}
	constraintCheck, err := CheckGoalConstraints(primitiveContext, concrete, checkedDocument, constructed.Interface, effectiveTermLimits, budget.remaining())
	if err != nil {
		if isCertificateResourceError(err) {
			budget.markTermResource(effectiveTermLimits.MaxNodes)
		}
		return SynthesisCertificate{}, err
	}
	if err := budget.charge(constraintCheck.WorkUnits, "constraint_check"); err != nil {
		return SynthesisCertificate{}, err
	}
	budget.statistics.VisitedNodes += constraintCheck.VisitedNodes
	if constraintCheck.Status != status.OutcomeOK || !constraintCheck.Complete || !constraintCheck.GoalSatisfied {
		return SynthesisCertificate{}, certEmitInvalid("CONSTRUCTION_GOAL_NOT_MATCHED", "constructed.canonical_document.expression", "independent structural checking did not match every concrete goal constraint")
	}
	obligations, err := certificateStructuralObligations(concrete, constraintCheck)
	if err != nil {
		return SynthesisCertificate{}, err
	}

	if err := budget.charge(len(constructed.Trace), "operations"); err != nil {
		return SynthesisCertificate{}, err
	}
	operations, err := certifiedOperationsFromTrace(concrete, constructed.Trace)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	remaining := budget.remaining()
	if remaining <= 0 {
		return SynthesisCertificate{}, certEmitResource("operation_replay", "no work budget remains for independent operation replay")
	}
	replayed, replayStats, err := ReplayCertifiedOperations(primitiveContext, concrete, operations, GoalVerificationLimits{
		MaxReplaySteps: minInt(goalLimits.MaxSteps, budget.limits.MaxReplaySteps), MaxWorkNodes: remaining, TermLimits: effectiveTermLimits,
	})
	budget.statistics.ReplaySteps += replayStats.ReplaySteps
	budget.statistics.VisitedNodes += replayStats.WorkNodes
	if chargeErr := budget.charge(replayStats.WorkNodes, "operation_replay"); chargeErr != nil {
		return SynthesisCertificate{}, chargeErr
	}
	if err != nil {
		return SynthesisCertificate{}, err
	}
	replayMatches, err := algebra.RewriteTermEqual(replayed.Expression, constructed.Expression, effectiveTermLimits)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	if !replayMatches || !samePortBindings(replayed.Interface, constructed.Interface) {
		return SynthesisCertificate{}, certEmitInvalid("PRODUCER_REPLAY_MISMATCH", "constructed.trace", "independent operation replay differs from the constructed expression or interface")
	}

	if err := budget.charge(len(original.Ports)+len(original.Constraints)+len(original.AllowedOperators), "goal_fingerprint"); err != nil {
		return SynthesisCertificate{}, err
	}
	goalFingerprint, err := searchGoalFingerprint(original, environment.SnapshotID)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	requestContextFingerprint, err := goalContextFingerprint(environment.Document, concrete)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	candidateDigest := sha256.Sum256(append([]byte("dse-candidate-document-v1\x00"), candidateBytes...))
	if err := budget.charge(len(original.References), "request_reference_boundary"); err != nil {
		return SynthesisCertificate{}, err
	}
	referenceBoundary, err := certificateReferenceBoundary(original.References)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	remaining = budget.remaining()
	if remaining <= 0 {
		return SynthesisCertificate{}, certEmitResource("scopes", "no work budget remains for scope traversal")
	}
	scopeLimits := effectiveTermLimits
	if remaining < scopeLimits.MaxNodes {
		scopeLimits.MaxNodes = remaining
	}
	scopes, scopeNodes, err := certificateScopes(checkedDocument.Expression, scopeLimits)
	if scopeNodes > 0 {
		if chargeErr := budget.charge(scopeNodes, "scopes"); chargeErr != nil {
			return SynthesisCertificate{}, chargeErr
		}
		budget.statistics.VisitedNodes += scopeNodes
	}
	if err != nil {
		return SynthesisCertificate{}, err
	}
	if err := budget.charge(len(scopes)+len(obligations)+len(choices), "certificate_fields"); err != nil {
		return SynthesisCertificate{}, err
	}
	certificate, err := cloneCertificateInputs(concrete, choices, operations, constructed.Interface, scopes, obligations)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	certificate.SchemaVersion = SynthesisCertificateSchemaVersion
	certificate.Format = SynthesisCertificateFormat
	certificate.ProducerStatus = SynthesisCertificateProducerStatus
	certificate.GoalVerified = false
	certificate.Origin = origin
	certificate.GoalFingerprint = goalFingerprint
	certificate.SnapshotID = environment.SnapshotID
	certificate.ContextID = concrete.ContextID
	certificate.RequestContextFingerprint = requestContextFingerprint
	certificate.Profile = concrete.Profile
	certificate.CandidateFingerprint = "sha256:" + hex.EncodeToString(candidateDigest[:])
	certificate.Assumptions = []CertificateAssumption{
		{Kind: CertificateAssumptionRegistrySnapshot, Value: environment.SnapshotID},
		{Kind: CertificateAssumptionRequestReferenceBound, Value: referenceBoundary},
		{Kind: CertificateAssumptionPositiveProfile, Value: concrete.Profile},
	}
	certificate.IdentityProof = nil
	certificate.IdentityKey = ""
	certificate = normalizeCertificate(certificate)

	encoded, err := json.Marshal(certificate)
	if err != nil {
		return SynthesisCertificate{}, certEmitInvalid("CERTIFICATE_ENCODING", "certificate", err.Error())
	}
	budget.statistics.CertificateBytes = len(encoded)
	certificateByteLimit := minInt(maxEmittedCertificateBytes, budget.limits.MaxCertificateBytes)
	if len(encoded) > certificateByteLimit {
		return SynthesisCertificate{}, certEmitResource("certificate", "emitted certificate exceeds the 8 MiB profile limit")
	}
	return certificate, nil
}

func certifiedOperationsFromTrace(goal GoalSpec, trace []algebra.BuildTrace) ([]CertifiedOperation, error) {
	byID := make(map[string]algebra.BuildTrace, len(trace))
	operations := make([]CertifiedOperation, 0, len(trace))
	for i, step := range trace {
		path := fmt.Sprintf("constructed.trace[%d]", i)
		if step.Status != status.OutcomeOK || step.CheckedType == nil {
			return nil, certEmitInvalid("INVALID_CONSTRUCTION_STEP", path, "each producer step must have an OK status and checked type")
		}
		if _, duplicate := byID[step.ID]; duplicate {
			return nil, certEmitInvalid("DUPLICATE_CONSTRUCTION_STEP", path+".id", "construction trace IDs must be unique")
		}
		byID[step.ID] = step
		op := CertifiedOperation{
			ID: step.ID, RuleID: ruleIDForOperation(step.Op), Op: step.Op,
			Inputs: append([]string{}, step.Inputs...), Symbol: step.Symbol,
			Mapping: append([]algebra.JoinPair{}, step.Mapping...), Outputs: append([]algebra.JoinOutput{}, step.Outputs...),
			Port: step.Port, Ports: append([]string{}, step.Ports...), ClaimedStatus: step.Status,
			CheckedType: step.CheckedType, Interface: append([]algebra.PortBinding{}, step.Interface...),
			Substitutions: []CertifiedSubstitution{},
		}
		if op.RuleID == "" {
			return nil, certEmitUnsupported("UNSUPPORTED_CONSTRUCTION_OPERATOR", path+".op", "construction trace contains an operator outside the certified profile")
		}
		switch step.Op {
		case "relation", "join", "project", "permute":
		case "restrict":
			goalPortID, ok := goalPortForAlias(goal, step.Port)
			if !ok {
				return nil, certEmitInvalid("UNMAPPED_RESTRICTION_PORT", path+".port", "restriction port must map to exactly one concrete goal port")
			}
			op.GoalPortID = goalPortID
		case "bind":
			goalPortID, ok := goalPortForAlias(goal, step.Port)
			if !ok {
				return nil, certEmitInvalid("UNMAPPED_BIND_PORT", path+".port", "bound port must map to exactly one concrete goal port")
			}
			port, ok := concretePort(goal, goalPortID)
			if !ok || port.Binding == nil || port.Binding.Kind != algebra.TermRef || port.Binding.Type == nil || len(step.Inputs) != 1 {
				return nil, certEmitInvalid("BIND_TRACE_NOT_GOAL_REFERENCE", path, "bind trace must correspond to one explicitly Ref-bound goal port")
			}
			input, exists := byID[step.Inputs[0]]
			if !exists {
				return nil, certEmitInvalid("BIND_TRACE_INPUT_MISSING", path+".inputs", "bind input must be an earlier traced operation")
			}
			binder := ""
			for _, binding := range input.Interface {
				if binding.Port == step.Port {
					if binder != "" {
						return nil, certEmitInvalid("AMBIGUOUS_BIND_TRACE_PORT", path+".port", "bind input has multiple matching interface ports")
					}
					binder = binding.Binder
				}
			}
			if binder == "" {
				return nil, certEmitInvalid("BIND_TRACE_PORT_MISSING", path+".port", "bind input does not expose the selected port")
			}
			op.GoalPortID = goalPortID
			op.ReferenceID = port.Binding.ID
			op.Substitutions = []CertifiedSubstitution{{BinderID: binder, ReferenceID: port.Binding.ID, Type: *port.Binding.Type}}
		default:
			return nil, certEmitUnsupported("UNSUPPORTED_CONSTRUCTION_OPERATOR", path+".op", "construction trace contains an operator outside the certified profile")
		}
		operations = append(operations, op)
	}
	return operations, nil
}

func goalPortForAlias(goal GoalSpec, alias string) (string, bool) {
	match := ""
	for _, port := range goal.Ports {
		if portAliasForGoal(goal, port.ID) != alias {
			continue
		}
		if match != "" {
			return "", false
		}
		match = port.ID
	}
	return match, match != ""
}

func concretePort(goal GoalSpec, id string) (GoalPort, bool) {
	for _, port := range goal.Ports {
		if port.ID == id {
			return port, true
		}
	}
	return GoalPort{}, false
}

func certificateStructuralObligations(goal GoalSpec, checked GoalConstraintCheck) ([]GoalObligation, error) {
	byID := make(map[string]GoalConstraintStatus, len(checked.Constraints))
	for _, constraint := range checked.Constraints {
		if _, duplicate := byID[constraint.ConstraintID]; duplicate {
			return nil, certEmitInvalid("DUPLICATE_CHECKED_CONSTRAINT", "constraint_check.constraints", "independent constraint results must have unique IDs")
		}
		byID[constraint.ConstraintID] = constraint
	}
	obligations := make([]GoalObligation, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		status, exists := byID[constraint.ID]
		if !exists || status.Status != "checked" {
			return nil, certEmitInvalid("CONSTRAINT_NOT_STRUCTURALLY_CHECKED", "constraint_check.constraints", "every concrete goal constraint must have a checked structural result")
		}
		obligations = append(obligations, GoalObligation{
			ConstraintID: constraint.ID, Kind: constraint.Kind, Status: "structural_claim",
			StepIDs: []string{}, Fragments: append([]string{}, status.Fragments...),
		})
	}
	return obligations, nil
}

func certificateReferenceBoundary(references []algebra.Declaration) (string, error) {
	ordered := append([]algebra.Declaration{}, references...)
	sortDeclarationsByID(ordered)
	payload, err := json.Marshal(ordered)
	if err != nil {
		return "", certEmitInvalid("REFERENCE_BOUNDARY_ENCODING", "references", err.Error())
	}
	if len(payload) > algebra.MaxDocumentBytes {
		return "", certEmitResource("references", "request reference boundary exceeds 1 MiB")
	}
	digest := sha256.Sum256(append([]byte("dse-request-reference-boundary-v1\x00"), payload...))
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func certificateScopes(root *algebra.Term, limits algebra.Limits) ([]CertificateScope, int, error) {
	if root == nil {
		return nil, 0, certEmitInvalid("MISSING_CANDIDATE_EXPRESSION", "candidate.expression", "scope inventory requires an expression")
	}
	if err := algebra.CheckDeclarations(algebra.Document{SchemaVersion: algebra.SchemaVersion, Limits: limits}); err != nil {
		return nil, 0, err
	}
	scopes := make([]CertificateScope, 0)
	active := make(map[*algebra.Term]bool)
	visited := 0
	var walk func(*algebra.Term, string, string, int) error
	walk = func(term *algebra.Term, path, parent string, depth int) error {
		if term == nil {
			return certEmitInvalid("MISSING_TERM_CHILD", path, "candidate term contains a missing child")
		}
		if depth > limits.MaxDepth {
			return certEmitResource(path, "scope traversal exceeds the term depth limit")
		}
		if visited >= limits.MaxNodes {
			return certEmitResource(path, "scope traversal exceeds the term node limit")
		}
		if active[term] {
			return certEmitInvalid("CYCLIC_CANDIDATE_TERM", path, "candidate term graph contains a cycle")
		}
		visited++
		active[term] = true
		defer delete(active, term)
		if term.Kind == algebra.TermLambda || term.Kind == algebra.TermExists {
			if term.Type == nil {
				return certEmitInvalid("MISSING_SCOPE_TYPE", path+".type", "binder type is required")
			}
			typeCopy, err := copyCertificateType(*term.Type)
			if err != nil {
				return err
			}
			scopes = append(scopes, CertificateScope{BinderID: term.ID, Kind: term.Kind, Type: typeCopy, Path: path, ParentBinderID: parent})
			parent = term.ID
		}
		switch term.Kind {
		case algebra.TermVar, algebra.TermParam, algebra.TermRef, algebra.TermHole:
		case algebra.TermAtom:
			for i, arg := range term.Args {
				if err := walk(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), parent, depth+1); err != nil {
					return err
				}
			}
		case algebra.TermAnd:
			for i, operand := range term.Operands {
				if err := walk(operand, fmt.Sprintf("%s.operands[%d]", path, i), parent, depth+1); err != nil {
					return err
				}
			}
		case algebra.TermExists, algebra.TermLambda:
			return walk(term.Body, path+".body", parent, depth+1)
		case algebra.TermModal:
			return walk(term.Body, path+".body", parent, depth+1)
		case algebra.TermApply:
			if err := walk(term.Function, path+".function", parent, depth+1); err != nil {
				return err
			}
			return walk(term.Argument, path+".argument", parent, depth+1)
		default:
			return certEmitUnsupported("UNSUPPORTED_SCOPE_TERM", path+".kind", "candidate term kind is outside the checked scope profile")
		}
		return nil
	}
	if err := walk(root, "expression", "", 1); err != nil {
		return nil, visited, err
	}
	return scopes, visited, nil
}

func isCertificateResourceError(err error) bool {
	var typed *algebra.Error
	return asAlgebraError(err, &typed) && typed.Status == status.OutcomeResourceLimit
}

func copyCertificateType(typ algebra.Type) (algebra.Type, error) {
	encoded, err := json.Marshal(typ)
	if err != nil {
		return algebra.Type{}, certEmitInvalid("CERTIFICATE_TYPE_COPY", "scope.type", err.Error())
	}
	copy, err := algebra.DecodeTypeJSON(encoded, "scope.type")
	if err != nil {
		return algebra.Type{}, err
	}
	return copy, nil
}

func cloneCertificateInputs(concrete GoalSpec, choices []ChoiceAssignment, operations []CertifiedOperation, bindings []algebra.PortBinding, scopes []CertificateScope, obligations []GoalObligation) (SynthesisCertificate, error) {
	input := struct {
		Goal        GoalSpec              `json:"goal"`
		Choices     []ChoiceAssignment    `json:"choices"`
		Operations  []CertifiedOperation  `json:"operations"`
		Interface   []algebra.PortBinding `json:"interface"`
		Scopes      []CertificateScope    `json:"scopes"`
		Obligations []GoalObligation      `json:"obligations"`
	}{concrete, choices, operations, bindings, scopes, obligations}
	encoded, err := json.Marshal(input)
	if err != nil {
		return SynthesisCertificate{}, certEmitInvalid("CERTIFICATE_COPY", "certificate", err.Error())
	}
	if len(encoded) > maxEmittedCertificateBytes {
		return SynthesisCertificate{}, certEmitResource("certificate", "certificate fields exceed the 8 MiB profile limit")
	}
	var copy struct {
		Goal        json.RawMessage       `json:"goal"`
		Choices     []ChoiceAssignment    `json:"choices"`
		Operations  []CertifiedOperation  `json:"operations"`
		Interface   []algebra.PortBinding `json:"interface"`
		Scopes      []CertificateScope    `json:"scopes"`
		Obligations []GoalObligation      `json:"obligations"`
	}
	if err := json.Unmarshal(encoded, &copy); err != nil {
		return SynthesisCertificate{}, certEmitInvalid("CERTIFICATE_COPY", "certificate", err.Error())
	}
	goalCopy, err := DecodeGoal(copy.Goal)
	if err != nil {
		return SynthesisCertificate{}, err
	}
	return SynthesisCertificate{
		ConcreteGoal: goalCopy, Choices: copy.Choices, Operations: copy.Operations,
		Interface: copy.Interface, Scopes: copy.Scopes, Obligations: copy.Obligations,
		Assumptions: []CertificateAssumption{},
	}, nil
}

func sortDeclarationsByID(declarations []algebra.Declaration) {
	for i := 1; i < len(declarations); i++ {
		for j := i; j > 0 && declarations[j].ID < declarations[j-1].ID; j-- {
			declarations[j], declarations[j-1] = declarations[j-1], declarations[j]
		}
	}
}

func certEmitInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func certEmitUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}

func certEmitIncomplete(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeIncomplete, Code: code, Path: path, Message: message}
}

func certEmitResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

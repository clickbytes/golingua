package synthesis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// VerifyGoalCertificate independently checks a producer certificate against
// the explicitly selected primitive and rewrite environments. It verifies a
// bounded structural profile only; it does not evaluate facts or prove truth.
func VerifyGoalCertificate(primitiveContext algebra.Document, environment algebra.RewriteEnvironment, original GoalSpec, candidate algebra.Document, certificate SynthesisCertificate, requested GoalVerificationLimits) (GoalVerificationResult, error) {
	result := newGoalVerificationResult()
	limits, err := NormalizeGoalVerificationLimits(requested)
	if err != nil {
		return result, err
	}
	result.ConsistencyScope = GoalVerificationConsistencyScope

	if _, err := validateSearchGoal(primitiveContext, original); err != nil {
		return result, err
	}
	if err := validateSearchEnvironment(primitiveContext, environment, original); err != nil {
		return result, err
	}
	result.SnapshotID = environment.SnapshotID
	result.ContextID = original.ContextID

	goalLimits, err := NormalizeGoalLimits(original.Limits)
	if err != nil {
		return result, err
	}
	contextLimits, err := algebra.NormalizeLimits(primitiveContext.Limits)
	if err != nil {
		return result, err
	}
	candidateLimits, err := algebra.NormalizeLimits(candidate.Limits)
	if err != nil {
		return result, err
	}
	effectiveTermLimits := minimumAlgebraLimits(goalLimits.TermLimits, limits.TermLimits, contextLimits, candidateLimits)
	if err := checkGoalTypesUnderLimits(primitiveContext, original, effectiveTermLimits); err != nil {
		return goalVerificationFailure(result, err), nil
	}
	workLimit := minInt(limits.MaxWorkNodes, DefaultGoalCheckMaxWork)
	workRemaining := workLimit
	charge := func(amount int, path string) error {
		if amount < 0 || amount > workRemaining {
			workRemaining = 0
			result.Statistics.WorkNodes = workLimit
			return certEmitResource(path, "certificate verification exceeds the bounded work-node limit")
		}
		workRemaining -= amount
		result.Statistics.WorkNodes += amount
		return nil
	}

	envelopeLimits := limits
	envelopeLimits.TermLimits = effectiveTermLimits
	if err := validateCertificateEnvelope(certificate, original, envelopeLimits); err != nil {
		return goalVerificationFailure(result, err), nil
	}
	if certificate.IdentityProof != nil || certificate.IdentityKey != "" {
		return goalVerificationUndetermined(result, certEmitUnsupported("UNSUPPORTED_IDENTITY_CLAIM", "identity_proof", "goal certificates in this profile do not carry alpha identity claims")), nil
	}
	declaredTypeContext := primitiveContext
	declaredTypeContext.Limits = effectiveTermLimits
	if err := validateCertificateDeclaredTypes(declaredTypeContext, certificate); err != nil {
		return goalVerificationFailure(result, err), nil
	}

	if err := checkGoalTypesUnderLimits(primitiveContext, certificate.ConcreteGoal, effectiveTermLimits); err != nil {
		return goalVerificationFailure(result, err), nil
	}
	if err := ValidateGoalRefinement(primitiveContext, original, certificate.ConcreteGoal, certificate.Choices, certificate.Origin); err != nil {
		return goalVerificationReject(result, err), nil
	}
	result.RefinementChecked = true
	if certificate.SnapshotID != environment.SnapshotID || certificate.ContextID != original.ContextID || certificate.Profile != original.Profile {
		return goalVerificationReject(result, certEmitInvalid("CERTIFICATE_CONTEXT_MISMATCH", "certificate", "snapshot, context, and profile claims must match the selected environment and request")), nil
	}
	goalContextFP, contextErr := goalContextFingerprint(primitiveContext, original)
	if contextErr != nil {
		return result, contextErr
	}
	if certificate.RequestContextFingerprint != goalContextFP {
		return goalVerificationReject(result, certEmitInvalid("CERTIFICATE_REQUEST_CONTEXT_MISMATCH", "certificate.request_context_fingerprint", "certificate request-context pin must match the caller-supplied goal context")), nil
	}
	if (candidate.RequestContext == nil) != (original.RequestContext == nil) {
		return goalVerificationReject(result, certEmitInvalid("CANDIDATE_REQUEST_CONTEXT_MISMATCH", "candidate.request_context", "candidate must preserve the request context carried by the goal")), nil
	}
	if !sameDeclarationInventory(certificate.ConcreteGoal.References, candidate.References) {
		return goalVerificationReject(result, certEmitInvalid("CANDIDATE_REFERENCE_ATTRIBUTES_MISMATCH", "candidate.references", "candidate references and their explicit attributes must match the concrete goal")), nil
	}
	if original.RequestContext != nil {
		candidateContext := primitiveContext
		candidateContext.References = append([]algebra.Declaration{}, candidate.References...)
		candidateContext.RequestContext = candidate.RequestContext
		candidateFP, fpErr := algebra.RequestContextFingerprint(*candidate.RequestContext, candidateContext)
		if fpErr != nil {
			return goalVerificationReject(result, fpErr), nil
		}
		if candidateFP != goalContextFP {
			return goalVerificationReject(result, certEmitInvalid("CANDIDATE_REQUEST_CONTEXT_MISMATCH", "candidate.request_context", "candidate request-context payload must match the caller-supplied goal context")), nil
		}
	}
	goalFingerprint, err := searchGoalFingerprint(original, environment.SnapshotID)
	if err != nil {
		return result, err
	}
	result.GoalFingerprint = goalFingerprint
	if certificate.GoalFingerprint != goalFingerprint {
		return goalVerificationReject(result, certEmitInvalid("GOAL_FINGERPRINT_MISMATCH", "certificate.goal_fingerprint", "goal fingerprint does not match the original request and selected snapshot")), nil
	}
	if err := validateCertificateAssumptions(certificate, original, environment); err != nil {
		return goalVerificationReject(result, err), nil
	}
	if err := validateCertifiedOperationClaims(primitiveContext, certificate.Operations, effectiveTermLimits); err != nil {
		return goalVerificationFailure(result, err), nil
	}
	encodedCertificate, err := json.Marshal(certificate)
	if err != nil {
		return result, certEmitInvalid("CERTIFICATE_ENCODING", "certificate", err.Error())
	}
	result.Statistics.CertificateBytes = len(encodedCertificate)
	if len(encodedCertificate) > minInt(limits.MaxCertificateBytes, MaxGoalVerificationMaxCertificateBytes) {
		return goalVerificationUndetermined(result, certEmitResource("certificate", "certificate exceeds the effective byte limit")), nil
	}
	result.Assumptions = []string{
		CertificateAssumptionRegistrySnapshot + "=" + environment.SnapshotID,
		CertificateAssumptionRequestReferenceBound + "=" + certificate.Assumptions[1].Value,
		CertificateAssumptionPositiveProfile + "=" + original.Profile,
	}

	if workRemaining <= 0 {
		return goalVerificationUndetermined(result, certEmitResource("candidate", "no work budget remains for candidate checking")), nil
	}
	constraintCheck, err := CheckGoalConstraints(primitiveContext, certificate.ConcreteGoal, candidate, certificate.Interface, effectiveTermLimits, workRemaining)
	if err != nil {
		if isCertificateResourceError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, effectiveTermLimits.MaxNodes)
			return goalVerificationUndetermined(result, err), nil
		}
		return result, err
	}
	result.ConstraintCheck = &constraintCheck
	result.TypeCorrect = constraintCheck.TypeCorrect
	if err := charge(constraintCheck.WorkUnits, "candidate_check"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	result.Statistics.VisitedNodes += constraintCheck.VisitedNodes
	if !constraintCheck.Complete {
		return goalVerificationUndetermined(result, firstConstraintDiagnostic(constraintCheck)), nil
	}
	result.Complete = true
	result.GoalSatisfied = constraintCheck.GoalSatisfied
	if err := validateObligationClaims(certificate.ConcreteGoal, certificate.Obligations, constraintCheck); err != nil {
		return goalVerificationReject(result, err), nil
	}

	if workRemaining <= 0 {
		return goalVerificationUndetermined(result, certEmitResource("operations", "no work budget remains for operation replay")), nil
	}
	replaySteps := minInt(limits.MaxReplaySteps, goalLimits.MaxSteps)
	if len(certificate.Operations) > replaySteps {
		return goalVerificationUndetermined(result, certEmitResource("operations", "certificate exceeds the effective replay-step limit")), nil
	}
	replayed, replayStats, err := ReplayCertifiedOperations(primitiveContext, certificate.ConcreteGoal, certificate.Operations, GoalVerificationLimits{
		MaxCertificateBytes: limits.MaxCertificateBytes,
		MaxReplaySteps:      replaySteps,
		MaxWorkNodes:        workRemaining,
		TermLimits:          effectiveTermLimits,
	})
	result.Statistics.ReplaySteps = replayStats.ReplaySteps
	result.Statistics.VisitedNodes += replayStats.WorkNodes
	if err != nil {
		if isCertificateResourceError(err) || isCertificateIncompleteError(err) || isCertificateUnsupportedError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, replayStats.WorkNodes)
			if isCertificateResourceError(err) && replayStats.WorkNodes == 0 {
				result.Statistics.WorkNodes += minInt(workRemaining, effectiveTermLimits.MaxNodes)
			}
			return goalVerificationUndetermined(result, err), nil
		}
		return goalVerificationReject(result, err), nil
	}
	if err := charge(replayStats.WorkNodes, "operation_replay"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	result.ReplayChecked = true

	checkedDocument := candidate
	checkedDocument.Limits = effectiveTermLimits
	checked, err := algebra.Check(checkedDocument)
	if err != nil {
		if isCertificateResourceError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, effectiveTermLimits.MaxNodes)
			return goalVerificationUndetermined(result, err), nil
		}
		return result, err
	}
	if err := charge(checked.NodeCount, "candidate_document_check"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	result.Statistics.VisitedNodes += checked.NodeCount
	replayedMatches, err := algebra.RewriteTermEqual(replayed.Expression, candidate.Expression, effectiveTermLimits)
	if err != nil {
		if isCertificateResourceError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, checked.NodeCount)
			return goalVerificationUndetermined(result, err), nil
		}
		return result, err
	}
	if err := charge(checked.NodeCount, "candidate_expression_equality"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	if !replayedMatches || !samePortBindings(replayed.Interface, certificate.Interface) {
		return goalVerificationReject(result, certEmitInvalid("REPLAY_CANDIDATE_MISMATCH", "candidate.expression", "replayed operations do not yield the submitted candidate expression and interface")), nil
	}

	scopes, scopeNodes, err := certificateScopes(candidate.Expression, effectiveTermLimits)
	if err != nil {
		if isCertificateResourceError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, scopeNodes)
			return goalVerificationUndetermined(result, err), nil
		}
		return result, err
	}
	if err := charge(scopeNodes, "certificate.scopes"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	result.Statistics.VisitedNodes += scopeNodes
	if err := compareCertificateScopes(certificate.Scopes, scopes); err != nil {
		return goalVerificationReject(result, err), nil
	}

	if !constraintCheck.GoalSatisfied {
		return goalVerificationReject(result, certEmitInvalid("GOAL_CONSTRAINTS_NOT_SATISFIED", "candidate.expression", "typed candidate does not satisfy every concrete structural constraint")), nil
	}
	result.GoalSatisfied = true
	if workRemaining <= 0 {
		return goalVerificationUndetermined(result, certEmitResource("candidate_fingerprint", "no work budget remains for candidate fingerprint verification")), nil
	}
	candidateBytes, err := algebra.Encode(candidate)
	if err != nil {
		if isCertificateResourceError(err) {
			result.Statistics.WorkNodes += minInt(workRemaining, checked.NodeCount)
			return goalVerificationUndetermined(result, err), nil
		}
		return result, err
	}
	if err := charge(checked.NodeCount, "candidate_fingerprint"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	if err := verifyCandidateFingerprintBytes(candidateBytes, certificate.CandidateFingerprint); err != nil {
		return goalVerificationReject(result, err), nil
	}
	if err := charge(len(certificate.Scopes)+len(certificate.Operations)+len(certificate.Obligations), "certificate_claims"); err != nil {
		return goalVerificationUndetermined(result, err), nil
	}
	result.CertificateValid = true
	result.Status = status.OutcomeOK
	result.Disposition = GoalVerificationDispositionChecked
	result.Complete = true
	return result, nil
}

func newGoalVerificationResult() GoalVerificationResult {
	return GoalVerificationResult{
		Status: status.OutcomeIncomplete, Disposition: GoalVerificationDispositionUndetermined,
		Assumptions: []string{}, Diagnostics: []algebra.Error{},
		ConsistencyScope:   GoalVerificationConsistencyScope,
		UniquenessVerified: false, SearchCompletenessVerified: false,
	}
}

func validateCertificateEnvelope(certificate SynthesisCertificate, original GoalSpec, limits GoalVerificationLimits) error {
	if certificate.SchemaVersion != SynthesisCertificateSchemaVersion || certificate.Format != SynthesisCertificateFormat {
		return certEmitInvalid("INVALID_CERTIFICATE_FORMAT", "certificate", "certificate schema version or format is unsupported")
	}
	if certificate.ProducerStatus != SynthesisCertificateProducerStatus || certificate.GoalVerified {
		return certEmitInvalid("INVALID_PRODUCER_CLAIM", "certificate", "producer status must remain unverified")
	}
	if certificate.Origin != CertificateOriginConstruct && certificate.Origin != CertificateOriginSearch {
		return certEmitInvalid("INVALID_CERTIFICATE_ORIGIN", "certificate.origin", "origin must be construct or search")
	}
	for _, field := range []struct{ name, value string }{
		{"goal_fingerprint", certificate.GoalFingerprint},
		{"snapshot_id", certificate.SnapshotID},
		{"context_id", certificate.ContextID},
		{"profile", certificate.Profile},
		{"candidate_fingerprint", certificate.CandidateFingerprint},
	} {
		if !goalID(field.value) {
			return certEmitInvalid("INVALID_CERTIFICATE_HEADER", "certificate."+field.name, "certificate header values must be bounded identifier text")
		}
	}
	if len(certificate.Interface) > MaxGoalPorts || len(certificate.Operations) == 0 || len(certificate.Operations) > minInt(limits.MaxReplaySteps, MaxGoalSteps) ||
		len(certificate.Scopes) > limits.TermLimits.MaxNodes || len(certificate.Obligations) > MaxGoalConstraints || len(certificate.Assumptions) > 3 || len(certificate.Choices) > MaxRefinementChoices {
		return certEmitResource("certificate", "certificate arrays exceed their bounded limits")
	}
	for i, binding := range certificate.Interface {
		if !goalID(binding.Port) || !goalID(binding.Binder) {
			return certEmitInvalid("INVALID_CERTIFICATE_INTERFACE", fmt.Sprintf("certificate.interface[%d]", i), "interface identifiers must be bounded nonempty text")
		}
	}
	if len(certificate.Assumptions) != 3 {
		return certEmitInvalid("CERTIFICATE_ASSUMPTIONS_MISMATCH", "certificate.assumptions", "exactly the three profile assumptions are required")
	}
	if certificate.IdentityProof != nil || certificate.IdentityKey != "" {
		return certEmitUnsupported("UNSUPPORTED_IDENTITY_CLAIM", "certificate.identity_proof", "identity claims are outside the goal certificate profile")
	}
	return nil
}

func validateCertificateDeclaredTypes(context algebra.Document, certificate SynthesisCertificate) error {
	declarations := make([]algebra.Declaration, 0, len(certificate.Scopes))
	for i, scope := range certificate.Scopes {
		if !goalID(scope.BinderID) || !utf8.ValidString(scope.Path) || len(scope.Path) == 0 || len(scope.Path) > 1024 ||
			strings.IndexByte(scope.Path, 0) >= 0 || scope.ParentBinderID != "" && !goalID(scope.ParentBinderID) {
			return certEmitInvalid("INVALID_CERTIFICATE_SCOPE", fmt.Sprintf("certificate.scopes[%d]", i), "scope binder and path fields must be bounded")
		}
		if scope.Kind != algebra.TermLambda && scope.Kind != algebra.TermExists {
			return certEmitInvalid("INVALID_CERTIFICATE_SCOPE_KIND", fmt.Sprintf("certificate.scopes[%d].kind", i), "scope kind must be Lambda or Exists")
		}
		declarations = append(declarations, algebra.Declaration{ID: fmt.Sprintf("certificate-scope-%d", i), Type: scope.Type})
	}
	if err := algebra.CheckDeclarations(algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Sorts: context.Sorts,
		Parameters: declarations, Limits: context.Limits,
	}); err != nil {
		return err
	}
	for i, obligation := range certificate.Obligations {
		if !goalID(obligation.ConstraintID) || obligation.Status == "" || len(obligation.Status) > 64 ||
			len(obligation.StepIDs) > MaxGoalSteps || len(obligation.Fragments) > MaxGoalAtoms {
			return certEmitInvalid("INVALID_CERTIFICATE_OBLIGATION", fmt.Sprintf("certificate.obligations[%d]", i), "obligation claims exceed bounded fields")
		}
		for j, value := range obligation.StepIDs {
			if !goalID(value) {
				return certEmitInvalid("INVALID_CERTIFICATE_OBLIGATION", fmt.Sprintf("certificate.obligations[%d].step_ids[%d]", i, j), "step IDs must be bounded identifier text")
			}
		}
		for j, value := range obligation.Fragments {
			if len(value) == 0 || len(value) > 1024 {
				return certEmitInvalid("INVALID_CERTIFICATE_OBLIGATION", fmt.Sprintf("certificate.obligations[%d].fragments[%d]", i, j), "fragment paths must be nonempty and bounded")
			}
		}
	}
	for i, assumption := range certificate.Assumptions {
		if len(assumption.Kind) == 0 || len(assumption.Kind) > 64 || len(assumption.Value) == 0 || len(assumption.Value) > 256 {
			return certEmitInvalid("INVALID_CERTIFICATE_ASSUMPTION", fmt.Sprintf("certificate.assumptions[%d]", i), "assumption fields must be nonempty and bounded")
		}
	}
	return nil
}

func checkGoalTypesUnderLimits(context algebra.Document, goal GoalSpec, limits algebra.Limits) error {
	declarations := make([]algebra.Declaration, 0, len(goal.Ports)+1)
	expected := goal.ExpectedType
	declarations = append(declarations, algebra.Declaration{ID: "verification-goal-expected-type", Type: expected})
	for i, port := range goal.Ports {
		declarations = append(declarations, algebra.Declaration{ID: fmt.Sprintf("verification-goal-port-%d", i), Type: port.Type})
		if port.Binding != nil && port.Binding.Type != nil {
			declarations = append(declarations, algebra.Declaration{ID: fmt.Sprintf("verification-goal-binding-%d", i), Type: *port.Binding.Type})
		}
	}
	document := algebra.Document{
		SchemaVersion: context.SchemaVersion, Sorts: context.Sorts, Signatures: context.Signatures,
		References: goal.References, Parameters: declarations, Limits: limits,
	}
	return algebra.CheckDeclarations(document)
}

func validateCertificateAssumptions(certificate SynthesisCertificate, original GoalSpec, environment algebra.RewriteEnvironment) error {
	boundary, err := certificateReferenceBoundary(original.References)
	if err != nil {
		return err
	}
	expected := []CertificateAssumption{
		{Kind: CertificateAssumptionRegistrySnapshot, Value: environment.SnapshotID},
		{Kind: CertificateAssumptionRequestReferenceBound, Value: boundary},
		{Kind: CertificateAssumptionPositiveProfile, Value: original.Profile},
	}
	if len(certificate.Assumptions) != len(expected) {
		return certEmitInvalid("CERTIFICATE_ASSUMPTIONS_MISMATCH", "certificate.assumptions", "assumption inventory differs from the fixed profile")
	}
	for i := range expected {
		if certificate.Assumptions[i] != expected[i] {
			return certEmitInvalid("CERTIFICATE_ASSUMPTIONS_MISMATCH", fmt.Sprintf("certificate.assumptions[%d]", i), "assumption kind or value differs from the selected request boundary")
		}
	}
	return nil
}

func validateObligationClaims(goal GoalSpec, claims []GoalObligation, checked GoalConstraintCheck) error {
	if len(claims) != len(goal.Constraints) || len(checked.Constraints) != len(goal.Constraints) {
		return certEmitInvalid("CERTIFICATE_OBLIGATIONS_MISMATCH", "certificate.obligations", "certificate must contain one structural claim per concrete constraint")
	}
	seen := make(map[string]struct{}, len(claims))
	actual := make(map[string]GoalConstraintStatus, len(checked.Constraints))
	for _, item := range checked.Constraints {
		if _, duplicate := actual[item.ConstraintID]; duplicate {
			return certEmitInvalid("DUPLICATE_VERIFIED_CONSTRAINT", "constraint_check.constraints", "constraint checker returned duplicate IDs")
		}
		actual[item.ConstraintID] = item
	}
	for i, goalConstraint := range goal.Constraints {
		claim := claims[i]
		if claim.ConstraintID != goalConstraint.ID || claim.Kind != goalConstraint.Kind || claim.Status != "structural_claim" || len(claim.StepIDs) != 0 {
			return certEmitInvalid("CERTIFICATE_OBLIGATION_MISMATCH", fmt.Sprintf("certificate.obligations[%d]", i), "obligation order and fields must match the concrete goal with no claimed proof step IDs")
		}
		if _, duplicate := seen[claim.ConstraintID]; duplicate {
			return certEmitInvalid("DUPLICATE_CERTIFICATE_OBLIGATION", fmt.Sprintf("certificate.obligations[%d].constraint_id", i), "constraint obligation is duplicated")
		}
		seen[claim.ConstraintID] = struct{}{}
		verified, exists := actual[claim.ConstraintID]
		if !exists || verified.Kind != claim.Kind || verified.Status != "checked" || !sameStrings(verified.Fragments, claim.Fragments) {
			return certEmitInvalid("CERTIFICATE_FRAGMENT_MISMATCH", fmt.Sprintf("certificate.obligations[%d].fragments", i), "fragment paths differ from independent structural matching")
		}
	}
	return nil
}

func compareCertificateScopes(claims, actual []CertificateScope) error {
	if len(claims) != len(actual) {
		return certEmitInvalid("CERTIFICATE_SCOPE_MISMATCH", "certificate.scopes", "scope records do not cover the actual candidate binders exactly")
	}
	for i := range actual {
		if claims[i].BinderID != actual[i].BinderID || claims[i].Kind != actual[i].Kind || claims[i].Path != actual[i].Path ||
			claims[i].ParentBinderID != actual[i].ParentBinderID || !algebra.TypeEqual(claims[i].Type, actual[i].Type) {
			return certEmitInvalid("CERTIFICATE_SCOPE_MISMATCH", fmt.Sprintf("certificate.scopes[%d]", i), "scope record differs from candidate binder identity, type, path, or parent")
		}
	}
	return nil
}

func verifyCandidateFingerprintBytes(encoded []byte, claimed string) error {
	digest := sha256.Sum256(append([]byte("dse-candidate-document-v1\x00"), encoded...))
	expected := "sha256:" + hex.EncodeToString(digest[:])
	if claimed != expected {
		return certEmitInvalid("CANDIDATE_FINGERPRINT_MISMATCH", "certificate.candidate_fingerprint", "candidate document fingerprint does not match the canonical candidate")
	}
	return nil
}

func firstConstraintDiagnostic(check GoalConstraintCheck) error {
	for _, item := range check.Constraints {
		if item.Status != "checked" {
			code, path, message := item.Code, item.Path, item.Message
			if code == "" {
				code, path, message = "GOAL_CONSTRAINT_NOT_SATISFIED", "candidate.expression", "candidate does not satisfy every structural goal constraint"
			}
			return certEmitIncomplete(code, path, message)
		}
	}
	if len(check.Unresolved) > 0 {
		item := check.Unresolved[0]
		return certEmitIncomplete(item.Code, item.Path, item.Message)
	}
	return certEmitIncomplete("CANDIDATE_CHECK_INCOMPLETE", "candidate", "candidate check did not complete")
}

func goalVerificationReject(result GoalVerificationResult, err error) GoalVerificationResult {
	result.Status = status.OutcomeInvalid
	result.Disposition = GoalVerificationDispositionRejected
	result.CertificateValid = false
	result.GoalSatisfied = result.ConstraintCheck != nil && result.ConstraintCheck.GoalSatisfied
	result.Diagnostics = append(result.Diagnostics, diagnosticFromError(err))
	return result
}

func goalVerificationUndetermined(result GoalVerificationResult, err error) GoalVerificationResult {
	var typed *algebra.Error
	if asAlgebraError(err, &typed) {
		result.Status = typed.Status
		if result.Status == status.OutcomeInvalid || result.Status == status.OutcomeOK {
			result.Status = status.OutcomeIncomplete
		}
	} else {
		result.Status = status.OutcomeIncomplete
	}
	result.Disposition = GoalVerificationDispositionUndetermined
	result.CertificateValid = false
	result.Complete = false
	result.Diagnostics = append(result.Diagnostics, diagnosticFromError(err))
	return result
}

func goalVerificationFailure(result GoalVerificationResult, err error) GoalVerificationResult {
	var typed *algebra.Error
	if asAlgebraError(err, &typed) && (typed.Status == status.OutcomeResourceLimit || typed.Status == status.OutcomeIncomplete || typed.Status == status.OutcomeUnsupported) {
		return goalVerificationUndetermined(result, err)
	}
	return goalVerificationReject(result, err)
}

func diagnosticFromError(err error) algebra.Error {
	var typed *algebra.Error
	if asAlgebraError(err, &typed) {
		return *typed
	}
	return algebra.Error{Status: status.OutcomeInvalid, Code: "GOAL_CERTIFICATE_VERIFICATION", Path: "certificate", Message: boundedVerificationMessage(err.Error())}
}

func boundedVerificationMessage(message string) string {
	if len(message) <= 1024 {
		return message
	}
	message = message[:1024]
	for !utf8.ValidString(message) {
		message = message[:len(message)-1]
	}
	return message
}

func isCertificateIncompleteError(err error) bool {
	var typed *algebra.Error
	return asAlgebraError(err, &typed) && typed.Status == status.OutcomeIncomplete
}

func isCertificateUnsupportedError(err error) bool {
	var typed *algebra.Error
	return asAlgebraError(err, &typed) && typed.Status == status.OutcomeUnsupported
}

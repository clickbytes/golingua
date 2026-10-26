package algebra

import (
	"encoding/json"

	"github.com/clickbytes/golingua/internal/status"
)

// VerifyEquivalenceCertificate independently checks a positive
// profile-relative equivalence certificate against the caller's actual
// inputs. It never normalizes or calls CompareExpressions.
func VerifyEquivalenceCertificate(environment RewriteEnvironment, left, right Document, profile RewriteProfile, contextID string, certificate EquivalenceCertificate, requested EquivalenceLimits) (EquivalenceVerification, error) {
	limits, err := NormalizeEquivalenceLimits(requested)
	if err != nil {
		return EquivalenceVerification{}, err
	}
	result := EquivalenceVerification{
		Status: status.OutcomeInvalid, Verdict: VerdictUndetermined,
		ClaimType: certificate.ClaimType, Scope: certificate.Scope,
		ContextID: contextID, SnapshotID: environment.SnapshotID,
		Assumptions: []string{
			"A valid claim is relative to the exact selected rewrite profile and metadata snapshot.",
			"References, parameters, roles, event identities, and binder scope retain their declared identity.",
			"This certificate does not establish truth, entailment, or general non-equivalence.",
		},
		Diagnostics: []Error{},
	}
	if !validEquivalenceText(contextID, 256) {
		return result, invalid("INVALID_EQUIVALENCE_CONTEXT", "context_id", "context ID must be non-empty bounded UTF-8")
	}
	if err := ValidateRewriteProfile(profile); err != nil {
		return result, err
	}
	if certificate.ClaimType != EquivalenceClaimProfile {
		result.Status = status.OutcomeUnsupported
		result.Diagnostics = append(result.Diagnostics, Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_EQUIVALENCE_CLAIM", Path: "certificate.claim_type", Message: "only positive selected-profile equivalence claims are supported"})
		return result, nil
	}
	if certificate.SchemaVersion != EquivalenceSchemaVersion || certificate.Scope != EquivalenceScope || certificate.SnapshotID != environment.SnapshotID || certificate.ContextID != contextID || certificate.Profile.ID != profile.ID || !sameRulesExact(certificate.Profile.Rules, profile.Rules) {
		return result, invalid("EQUIVALENCE_CERTIFICATE_CONTEXT_MISMATCH", "certificate", "version, scope, snapshot, context, and exact profile must match the actual verification request")
	}
	if len(certificate.SnapshotID) > 256 || len(certificate.ContextID) > 256 || len(certificate.ClaimType) > 64 || len(certificate.Scope) > 64 {
		return result, invalid("INVALID_EQUIVALENCE_CERTIFICATE", "certificate", "certificate header exceeds its text bounds")
	}
	if err := ValidateRewriteProfile(certificate.Profile); err != nil {
		return result, err
	}
	work := equivalenceWork{limit: limits.MaxWorkUnits}
	preparedEnvironment, _, environmentCost, err := prepareEquivalenceEnvironment(environment, limits.TermLimits)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			result.Statistics.WorkUnits = work.used
			return equivalenceVerificationPartial(result, typed.Status, typed), nil
		}
		return result, err
	}
	environment = preparedEnvironment
	if err := work.charge(environmentCost, "environment"); err != nil {
		result.Statistics.WorkUnits = work.used
		return equivalenceVerificationPartial(result, status.OutcomeResourceLimit, typedAlgebraError(err)), nil
	}
	leftAdmission, err := admitEquivalenceDocument(environment, left, limits.TermLimits, "left", &work)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			return equivalenceVerificationPartial(result, typed.Status, typed), nil
		}
		return result, err
	}
	leftFingerprint := leftAdmission.fingerprint
	result.LeftFingerprint = leftFingerprint
	rightAdmission, err := admitEquivalenceDocument(environment, right, limits.TermLimits, "right", &work)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			result.Statistics.WorkUnits = work.used
			return equivalenceVerificationPartial(result, typed.Status, typed), nil
		}
		return result, err
	}
	result.Statistics.WorkUnits = work.used
	result.LeftFingerprint = leftAdmission.fingerprint
	result.RightFingerprint = rightAdmission.fingerprint
	if certificate.LeftFingerprint != leftAdmission.fingerprint || certificate.RightFingerprint != rightAdmission.fingerprint {
		return result, invalid("EQUIVALENCE_INPUT_FINGERPRINT_MISMATCH", "certificate", "certificate source fingerprints do not match the actual canonical input documents")
	}
	if leftAdmission.checked.Status != status.OutcomeOK || rightAdmission.checked.Status != status.OutcomeOK {
		result.Status = status.OutcomeIncomplete
		result.Diagnostics = append(result.Diagnostics, Error{Status: status.OutcomeIncomplete, Code: "EQUIVALENCE_INPUT_INCOMPLETE", Path: "inputs", Message: "equivalence verification requires two complete expressions"})
		return result, nil
	}
	leftEnv, err := equivalenceSideEnvironment(environment, left, leftAdmission.limits)
	if err != nil {
		return result, err
	}
	rightEnv, err := equivalenceSideEnvironment(environment, right, rightAdmission.limits)
	if err != nil {
		return result, err
	}
	leftVerification, phase, err := verifyEquivalenceSide(leftEnv, profile, left.Expression, certificate.LeftRewrite, limits, &work, "left")
	result.Statistics.VerifiedWorkUnits += leftVerification.Statistics.WorkNodes
	result.Statistics.WorkUnits = work.used
	if err != nil {
		return result, err
	}
	if phase != status.OutcomeOK || !leftVerification.CertificateValid || !leftVerification.NormalizationComplete {
		return equivalenceVerificationPartial(result, phase, firstEquivalenceDiagnostic(leftVerification.Diagnostics, "LEFT_REWRITE_CERTIFICATE_INVALID", "left rewrite certificate did not independently verify")), nil
	}
	rightVerification, phase, err := verifyEquivalenceSide(rightEnv, profile, right.Expression, certificate.RightRewrite, limits, &work, "right")
	result.Statistics.VerifiedWorkUnits += rightVerification.Statistics.WorkNodes
	result.Statistics.WorkUnits = work.used
	if err != nil {
		return result, err
	}
	if phase != status.OutcomeOK || !rightVerification.CertificateValid || !rightVerification.NormalizationComplete {
		return equivalenceVerificationPartial(result, phase, firstEquivalenceDiagnostic(rightVerification.Diagnostics, "RIGHT_REWRITE_CERTIFICATE_INVALID", "right rewrite certificate did not independently verify")), nil
	}
	outputCompareCost := leftVerification.Checked.NodeCount + rightVerification.Checked.NodeCount
	if err := work.charge(outputCompareCost, "normal_form_comparison"); err != nil {
		result.Statistics.WorkUnits = work.used
		return equivalenceVerificationPartial(result, status.OutcomeResourceLimit, typedAlgebraError(err)), nil
	}
	equal, err := RewriteTermEqual(certificate.LeftRewrite.Output, certificate.RightRewrite.Output, intersectEquivalenceTermLimits(leftAdmission.limits, rightAdmission.limits))
	if err != nil {
		return result, err
	}
	if !equal {
		return result, invalid("EQUIVALENCE_NORMAL_FORMS_DIFFER", "certificate", "positive equivalence claim has unequal independently verified normal forms")
	}
	result.Statistics.ComparedNodes += outputCompareCost
	result.Statistics.WorkUnits = work.used
	encoded, err := json.Marshal(certificate)
	if err != nil {
		return result, invalid("EQUIVALENCE_CERTIFICATE_ENCODING", "certificate", "positive equivalence certificate could not be encoded")
	}
	if len(encoded) > limits.MaxCertificateBytes {
		return equivalenceVerificationPartial(result, status.OutcomeResourceLimit, Error{Status: status.OutcomeResourceLimit, Code: "EQUIVALENCE_CERTIFICATE_BYTES", Path: "certificate", Message: "combined equivalence certificate exceeds its byte limit"}), nil
	}
	if err := work.charge(len(encoded), "certificate"); err != nil {
		return equivalenceVerificationPartial(result, status.OutcomeResourceLimit, typedAlgebraError(err)), nil
	}
	result.Status = status.OutcomeOK
	result.CertificateValid = true
	result.EquivalenceProved = true
	result.Verdict = VerdictProvedEquivalent
	result.Statistics.CertificateBytes = len(encoded)
	result.Statistics.WorkUnits = work.used
	return result, nil
}

func equivalenceVerificationPartial(result EquivalenceVerification, phase status.OutcomeKind, diagnostic Error) EquivalenceVerification {
	result.CertificateValid = false
	result.EquivalenceProved = false
	result.Verdict = VerdictUndetermined
	if phase == "" || phase == status.OutcomeOK {
		phase = status.OutcomeIncomplete
	}
	result.Status = phase
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	return result
}

func firstEquivalenceDiagnostic(diagnostics []Error, code, message string) Error {
	if len(diagnostics) != 0 {
		return diagnostics[0]
	}
	return Error{Status: status.OutcomeIncomplete, Code: code, Path: "certificate", Message: message}
}

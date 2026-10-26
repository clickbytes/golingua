package algebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
)

const (
	EquivalenceSchemaVersion = 1
	EquivalenceClaimProfile  = "profile_equivalence_v1"
	EquivalenceScope         = "selected_rewrite_profile"
	DefaultEquivalenceWork   = 1_048_576
	MaxEquivalenceWork       = 16_777_216
	DefaultEquivalenceBytes  = 16 << 20
	MaxEquivalenceBytes      = 16 << 20
)

type EquivalenceVerdict string

const (
	VerdictProvedEquivalent    EquivalenceVerdict = "proved_equivalent"
	VerdictProvedNotEquivalent EquivalenceVerdict = "proved_not_equivalent"
	VerdictUndetermined        EquivalenceVerdict = "undetermined"
)

// EquivalenceLimits bounds all admission, normalization, verification, and
// comparison work in one request. Zero selects a finite default.
type EquivalenceLimits struct {
	MaxWorkUnits        int    `json:"max_work_units,omitempty"`
	MaxCertificateBytes int    `json:"max_certificate_bytes,omitempty"`
	TermLimits          Limits `json:"term_limits,omitempty"`
}

func DefaultEquivalenceLimits() EquivalenceLimits {
	return EquivalenceLimits{MaxWorkUnits: DefaultEquivalenceWork, MaxCertificateBytes: DefaultEquivalenceBytes, TermLimits: DefaultLimits()}
}

// DecodeEquivalenceCertificate strictly decodes a bounded positive profile-
// equivalence certificate. The supplied claim is verified separately against
// actual source documents, metadata, and context.
func DecodeEquivalenceCertificate(data []byte) (EquivalenceCertificate, error) {
	if err := PreflightJSONLimit(data, MaxEquivalenceBytes); err != nil {
		return EquivalenceCertificate{}, err
	}
	fields, err := strictObject(data, "equivalence_certificate", []string{"schema_version", "claim_type", "scope", "snapshot_id", "context_id", "profile", "left_fingerprint", "right_fingerprint", "left_rewrite", "right_rewrite"}, []string{"schema_version", "claim_type", "scope", "snapshot_id", "context_id", "profile", "left_fingerprint", "right_fingerprint", "left_rewrite", "right_rewrite"})
	if err != nil {
		return EquivalenceCertificate{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "equivalence_certificate.schema_version", 1, 1)
	if err != nil {
		return EquivalenceCertificate{}, err
	}
	values := make([]string, 0, 6)
	for _, field := range []string{"claim_type", "scope", "snapshot_id", "context_id", "left_fingerprint", "right_fingerprint"} {
		value, valueErr := rewriteJSONString(fields[field], "equivalence_certificate."+field, 256, false)
		if valueErr != nil {
			return EquivalenceCertificate{}, valueErr
		}
		values = append(values, value)
	}
	profile, err := decodeRewriteProfile(fields["profile"], "equivalence_certificate.profile")
	if err != nil {
		return EquivalenceCertificate{}, err
	}
	left, err := DecodeRewriteCertificate(fields["left_rewrite"])
	if err != nil {
		return EquivalenceCertificate{}, err
	}
	right, err := DecodeRewriteCertificate(fields["right_rewrite"])
	if err != nil {
		return EquivalenceCertificate{}, err
	}
	return EquivalenceCertificate{SchemaVersion: version, ClaimType: values[0], Scope: values[1], SnapshotID: values[2], ContextID: values[3], Profile: profile, LeftFingerprint: values[4], RightFingerprint: values[5], LeftRewrite: left, RightRewrite: right}, nil
}

// DecodeRewriteProfileJSON decodes a standalone finite profile document.
func DecodeRewriteProfileJSON(data []byte) (RewriteProfile, error) {
	if err := PreflightJSONLimit(data, MaxDocumentBytes); err != nil {
		return RewriteProfile{}, err
	}
	return decodeRewriteProfile(data, "profile")
}

func NormalizeEquivalenceLimits(requested EquivalenceLimits) (EquivalenceLimits, error) {
	if requested.MaxWorkUnits == 0 {
		requested.MaxWorkUnits = DefaultEquivalenceWork
	}
	if requested.MaxWorkUnits < 0 || requested.MaxWorkUnits > MaxEquivalenceWork {
		return EquivalenceLimits{}, invalid("INVALID_EQUIVALENCE_LIMIT", "limits.max_work_units", fmt.Sprintf("must be zero (default) or between 1 and %d", MaxEquivalenceWork))
	}
	if requested.MaxCertificateBytes == 0 {
		requested.MaxCertificateBytes = DefaultEquivalenceBytes
	}
	if requested.MaxCertificateBytes < 0 || requested.MaxCertificateBytes > MaxEquivalenceBytes {
		return EquivalenceLimits{}, invalid("INVALID_EQUIVALENCE_LIMIT", "limits.max_certificate_bytes", fmt.Sprintf("must be zero (default) or between 1 and %d", MaxEquivalenceBytes))
	}
	limits, err := NormalizeLimits(requested.TermLimits)
	if err != nil {
		return EquivalenceLimits{}, err
	}
	requested.TermLimits = limits
	return requested, nil
}

// EquivalenceStatistics counts bounded deterministic work. It is not CPU,
// energy, or wall-clock measurement.
type EquivalenceStatistics struct {
	WorkUnits         int `json:"work_units"`
	LeftWorkUnits     int `json:"left_work_units"`
	RightWorkUnits    int `json:"right_work_units"`
	VerifiedWorkUnits int `json:"verified_work_units"`
	ComparedNodes     int `json:"compared_nodes"`
	CertificateBytes  int `json:"certificate_bytes"`
}

type EquivalenceNormalizationReceipt struct {
	Status      status.OutcomeKind `json:"status"`
	Normalized  bool               `json:"normalized"`
	Statistics  RewriteStatistics  `json:"statistics"`
	Diagnostics []Error            `json:"diagnostics"`
}

// EquivalenceCertificate makes a positive claim only: both original inputs
// are independently reduced under the pinned profile and their complete
// verified normal forms are structurally equal.
type EquivalenceCertificate struct {
	SchemaVersion    int                `json:"schema_version"`
	ClaimType        string             `json:"claim_type"`
	Scope            string             `json:"scope"`
	SnapshotID       string             `json:"snapshot_id"`
	ContextID        string             `json:"context_id"`
	Profile          RewriteProfile     `json:"profile"`
	LeftFingerprint  string             `json:"left_fingerprint"`
	RightFingerprint string             `json:"right_fingerprint"`
	LeftRewrite      RewriteCertificate `json:"left_rewrite"`
	RightRewrite     RewriteCertificate `json:"right_rewrite"`
}

type EquivalenceResult struct {
	Status             status.OutcomeKind              `json:"status"`
	Verdict            EquivalenceVerdict              `json:"verdict"`
	Scope              string                          `json:"scope"`
	ContextID          string                          `json:"context_id"`
	SnapshotID         string                          `json:"snapshot_id"`
	Profile            RewriteProfile                  `json:"profile"`
	StructuralEqual    bool                            `json:"structural_equal"`
	LeftFingerprint    string                          `json:"left_fingerprint"`
	RightFingerprint   string                          `json:"right_fingerprint"`
	LeftNormalForm     *Term                           `json:"left_normal_form,omitempty"`
	RightNormalForm    *Term                           `json:"right_normal_form,omitempty"`
	LeftNormalization  EquivalenceNormalizationReceipt `json:"left_normalization"`
	RightNormalization EquivalenceNormalizationReceipt `json:"right_normalization"`
	Certificate        *EquivalenceCertificate         `json:"certificate,omitempty"`
	Assumptions        []string                        `json:"assumptions"`
	Statistics         EquivalenceStatistics           `json:"statistics"`
	Diagnostics        []Error                         `json:"diagnostics"`
}

// EquivalenceVerification reports the validity of one positive profile-relative
// certificate. It does not decide general non-equivalence or world truth.
type EquivalenceVerification struct {
	Status            status.OutcomeKind    `json:"status"`
	CertificateValid  bool                  `json:"certificate_valid"`
	EquivalenceProved bool                  `json:"equivalence_proved"`
	Verdict           EquivalenceVerdict    `json:"verdict"`
	ClaimType         string                `json:"claim_type"`
	Scope             string                `json:"scope"`
	ContextID         string                `json:"context_id"`
	SnapshotID        string                `json:"snapshot_id"`
	LeftFingerprint   string                `json:"left_fingerprint,omitempty"`
	RightFingerprint  string                `json:"right_fingerprint,omitempty"`
	Assumptions       []string              `json:"assumptions"`
	Diagnostics       []Error               `json:"diagnostics"`
	Statistics        EquivalenceStatistics `json:"statistics"`
}

type equivalenceWork struct {
	limit int
	used  int
}

func (work *equivalenceWork) charge(amount int, path string) error {
	if amount < 0 || amount > work.limit-work.used {
		return resourceLimit(path, "equivalence request exceeds its shared logical work budget")
	}
	work.used += amount
	return nil
}

type equivalenceAdmission struct {
	checked     CheckedExpression
	fingerprint string
	limits      Limits
}

// CompareExpressions performs label-insensitive structural comparison and
// bounded, proof-relative comparison under exactly the supplied rewrite rules.
// Different normal forms remain undetermined; this API emits no countermodel.
func CompareExpressions(environment RewriteEnvironment, left, right Document, profile RewriteProfile, contextID string, requested EquivalenceLimits) (EquivalenceResult, error) {
	limits, err := NormalizeEquivalenceLimits(requested)
	if err != nil {
		return EquivalenceResult{}, err
	}
	if !validEquivalenceText(contextID, 256) {
		return EquivalenceResult{}, invalid("INVALID_EQUIVALENCE_CONTEXT", "context_id", "context ID must be non-empty bounded UTF-8")
	}
	if err := matchEquivalenceRequestContexts(environment.Document, left, right, contextID); err != nil {
		return EquivalenceResult{}, err
	}
	if err := ValidateRewriteProfile(profile); err != nil {
		return EquivalenceResult{}, err
	}
	result := newEquivalenceResult(environment, profile, contextID)
	work := equivalenceWork{limit: limits.MaxWorkUnits}
	preparedEnvironment, effective, envWork, err := prepareEquivalenceEnvironment(environment, limits.TermLimits)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			finishEquivalencePartial(&result, typed.Status, []Error{typed}, "rewrite environment admission did not fit the shared work budget")
			result.Statistics.WorkUnits = work.used
			return result, nil
		}
		return EquivalenceResult{}, err
	}
	environment = preparedEnvironment
	if err := work.charge(envWork, "environment"); err != nil {
		finishEquivalencePartial(&result, status.OutcomeResourceLimit, []Error{typedAlgebraError(err)}, "rewrite environment exceeds the shared work budget")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	leftAdmission, err := admitEquivalenceDocument(environment, left, effective, "left", &work)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			finishEquivalencePartial(&result, typed.Status, []Error{typed}, "left source admission did not fit the shared work budget")
			result.Statistics.WorkUnits = work.used
			return result, nil
		}
		return EquivalenceResult{}, err
	}
	leftFingerprint := leftAdmission.fingerprint
	result.LeftFingerprint = leftFingerprint
	rightAdmission, err := admitEquivalenceDocument(environment, right, effective, "right", &work)
	if err != nil {
		if typed, ok := equivalenceBudgetError(err); ok {
			finishEquivalencePartial(&result, typed.Status, []Error{typed}, "right source admission did not fit the shared work budget")
			result.Statistics.WorkUnits = work.used
			return result, nil
		}
		return EquivalenceResult{}, err
	}
	result.LeftFingerprint, result.RightFingerprint = leftAdmission.fingerprint, rightAdmission.fingerprint
	result.Statistics.WorkUnits = work.used
	if leftAdmission.checked.Status == status.OutcomeIncomplete || rightAdmission.checked.Status == status.OutcomeIncomplete {
		result.Status = status.OutcomeIncomplete
		result.Verdict = VerdictUndetermined
		result.Diagnostics = append(result.Diagnostics, Error{Status: status.OutcomeIncomplete, Code: "EQUIVALENCE_INPUT_INCOMPLETE", Path: "inputs", Message: "one or both expressions contain unresolved holes"})
		return result, nil
	}
	compareCost := leftAdmission.checked.NodeCount + rightAdmission.checked.NodeCount
	if err := work.charge(compareCost, "structural_comparison"); err != nil {
		finishEquivalencePartial(&result, status.OutcomeResourceLimit, []Error{typedAlgebraError(err)}, "structural comparison exceeds the shared work budget")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	structurallyEqual, err := RewriteTermEqual(left.Expression, right.Expression, effective)
	if err != nil {
		return result, err
	}
	result.StructuralEqual = structurallyEqual
	result.Statistics.ComparedNodes = compareCost
	result.Statistics.WorkUnits = work.used

	leftEnv, err := equivalenceSideEnvironment(environment, left, leftAdmission.limits)
	if err != nil {
		return result, err
	}
	rightEnv, err := equivalenceSideEnvironment(environment, right, rightAdmission.limits)
	if err != nil {
		return result, err
	}
	leftNorm, phaseStatus, phaseErr := normalizeEquivalenceSide(leftEnv, profile, left.Expression, limits, &work, "left")
	result.LeftNormalization = normalizationReceipt(leftNorm)
	result.Statistics.LeftWorkUnits = leftNorm.Statistics.WorkNodes
	result.LeftNormalForm = leftNorm.Expression
	if phaseErr != nil {
		return result, phaseErr
	}
	if phaseStatus != status.OutcomeOK || !leftNorm.Normalized {
		finishEquivalencePartial(&result, phaseStatus, leftNorm.Diagnostics, "left normalization did not complete under the selected profile")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	rightNorm, phaseStatus, phaseErr := normalizeEquivalenceSide(rightEnv, profile, right.Expression, limits, &work, "right")
	result.RightNormalization = normalizationReceipt(rightNorm)
	result.Statistics.RightWorkUnits = rightNorm.Statistics.WorkNodes
	result.RightNormalForm = rightNorm.Expression
	if phaseErr != nil {
		return result, phaseErr
	}
	if phaseStatus != status.OutcomeOK || !rightNorm.Normalized {
		finishEquivalencePartial(&result, phaseStatus, rightNorm.Diagnostics, "right normalization did not complete under the selected profile")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	leftVerify, phaseStatus, phaseErr := verifyEquivalenceSide(leftEnv, profile, left.Expression, leftNorm.Certificate, limits, &work, "left")
	result.Statistics.VerifiedWorkUnits += leftVerify.Statistics.WorkNodes
	if phaseErr != nil {
		return result, phaseErr
	}
	if phaseStatus != status.OutcomeOK || !leftVerify.CertificateValid || !leftVerify.NormalizationComplete {
		finishEquivalencePartial(&result, phaseStatus, leftVerify.Diagnostics, "left rewrite receipt did not independently verify")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	rightVerify, phaseStatus, phaseErr := verifyEquivalenceSide(rightEnv, profile, right.Expression, rightNorm.Certificate, limits, &work, "right")
	result.Statistics.VerifiedWorkUnits += rightVerify.Statistics.WorkNodes
	if phaseErr != nil {
		return result, phaseErr
	}
	if phaseStatus != status.OutcomeOK || !rightVerify.CertificateValid || !rightVerify.NormalizationComplete {
		finishEquivalencePartial(&result, phaseStatus, rightVerify.Diagnostics, "right rewrite receipt did not independently verify")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	result.Status = status.OutcomeOK
	result.Verdict = VerdictUndetermined
	result.Statistics.WorkUnits = work.used
	formCompareCost := leftNorm.Checked.NodeCount + rightNorm.Checked.NodeCount
	if err := work.charge(formCompareCost, "normal_form_comparison"); err != nil {
		finishEquivalencePartial(&result, status.OutcomeResourceLimit, []Error{typedAlgebraError(err)}, "normal form comparison exceeds the shared work budget")
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	formsEqual, err := RewriteTermEqual(leftNorm.Expression, rightNorm.Expression, effective)
	if err != nil {
		return EquivalenceResult{}, err
	}
	result.Statistics.ComparedNodes += formCompareCost
	result.Statistics.WorkUnits = work.used
	if !formsEqual {
		return result, nil
	}
	certificate := EquivalenceCertificate{
		SchemaVersion: EquivalenceSchemaVersion, ClaimType: EquivalenceClaimProfile,
		Scope: EquivalenceScope, SnapshotID: environment.SnapshotID, ContextID: contextID,
		Profile: cloneEquivalenceProfile(profile), LeftFingerprint: leftAdmission.fingerprint,
		RightFingerprint: rightAdmission.fingerprint,
		LeftRewrite:      leftNorm.Certificate, RightRewrite: rightNorm.Certificate,
	}
	encoded, err := json.Marshal(certificate)
	if err != nil {
		return EquivalenceResult{}, invalid("EQUIVALENCE_CERTIFICATE_ENCODING", "certificate", "positive equivalence certificate could not be encoded")
	}
	if len(encoded) > limits.MaxCertificateBytes {
		result.Status = status.OutcomeResourceLimit
		result.Verdict = VerdictUndetermined
		result.Diagnostics = append(result.Diagnostics, Error{Status: status.OutcomeResourceLimit, Code: "EQUIVALENCE_CERTIFICATE_BYTES", Path: "certificate", Message: "combined equivalence certificate exceeds its byte limit"})
		return result, nil
	}
	if err := work.charge(len(encoded), "certificate"); err != nil {
		result.Status = status.OutcomeResourceLimit
		result.Verdict = VerdictUndetermined
		result.Diagnostics = append(result.Diagnostics, typedAlgebraError(err))
		result.Statistics.WorkUnits = work.used
		return result, nil
	}
	result.Certificate = &certificate
	result.Verdict = VerdictProvedEquivalent
	result.Statistics.CertificateBytes = len(encoded)
	result.Statistics.WorkUnits = work.used
	return result, nil
}

func matchEquivalenceRequestContexts(actualContext, left, right Document, contextID string) error {
	actual := actualContext.RequestContext
	if (actual == nil) != (left.RequestContext == nil) || (actual == nil) != (right.RequestContext == nil) {
		return invalid("EQUIVALENCE_REQUEST_CONTEXT_MISMATCH", "inputs.request_context", "both source documents and the selected environment must carry the same request context")
	}
	if actual == nil {
		return nil
	}
	if actual.ID != contextID || left.RequestContext.ID != contextID || right.RequestContext.ID != contextID {
		return invalid("EQUIVALENCE_REQUEST_CONTEXT_MISMATCH", "context_id", "request context IDs must match the supplied equivalence context ID")
	}
	_, expectedFingerprint, err := CanonicalRequestContext(*actual, actualContext)
	if err != nil {
		return err
	}
	for _, source := range []Document{left, right} {
		_, fingerprint, fingerprintErr := CanonicalRequestContext(*source.RequestContext, actualContext)
		if fingerprintErr != nil {
			return fingerprintErr
		}
		if fingerprint != expectedFingerprint {
			return invalid("EQUIVALENCE_REQUEST_CONTEXT_MISMATCH", "inputs.request_context", "source context payload differs from the independently selected context")
		}
	}
	return nil
}

func prepareEquivalenceEnvironment(environment RewriteEnvironment, requested Limits) (RewriteEnvironment, Limits, int, error) {
	baseLimits, err := NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return RewriteEnvironment{}, Limits{}, 0, err
	}
	limits, err := NormalizeLimits(requested)
	if err != nil {
		return RewriteEnvironment{}, Limits{}, 0, err
	}
	limits = intersectEquivalenceTermLimits(limits, baseLimits)
	environment.Document.Limits = limits
	if err := ValidateRewriteEnvironment(environment); err != nil {
		return RewriteEnvironment{}, Limits{}, 0, err
	}
	work, err := rewriteEnvironmentWork(environment, limits)
	if err != nil {
		return RewriteEnvironment{}, Limits{}, 0, err
	}
	return environment, limits, work, nil
}

func newEquivalenceResult(environment RewriteEnvironment, profile RewriteProfile, contextID string) EquivalenceResult {
	return EquivalenceResult{
		Status: status.OutcomeIncomplete, Verdict: VerdictUndetermined, Scope: EquivalenceScope,
		ContextID: contextID, SnapshotID: environment.SnapshotID, Profile: cloneEquivalenceProfile(profile),
		LeftNormalization:  EquivalenceNormalizationReceipt{Status: status.OutcomeIncomplete, Diagnostics: []Error{}},
		RightNormalization: EquivalenceNormalizationReceipt{Status: status.OutcomeIncomplete, Diagnostics: []Error{}},
		Assumptions: []string{
			"Verdicts are relative only to the exact selected rewrite profile and metadata snapshot.",
			"Reference and parameter IDs, roles, events, type structure, and scope remain identity sensitive.",
			"Open-world fact absence does not establish non-equivalence or falsity.",
		},
		Diagnostics: []Error{},
	}
}

func admitEquivalenceDocument(environment RewriteEnvironment, document Document, limits Limits, side string, work *equivalenceWork) (equivalenceAdmission, error) {
	path := side
	if document.SchemaVersion != SchemaVersion || document.Expression == nil {
		return equivalenceAdmission{}, invalid("INVALID_EQUIVALENCE_DOCUMENT", path, "version-1 self-contained expression document is required")
	}
	// Keep the submitted canonical document separate from the effective check
	// copy so its source fingerprint does not silently change with a lower cap.
	sourceDocument := document
	docLimits, err := NormalizeLimits(document.Limits)
	if err != nil {
		return equivalenceAdmission{}, err
	}
	limits = intersectEquivalenceTermLimits(limits, docLimits)
	document.Limits = limits
	checked, err := checkEquivalenceDocument(document, limits, work, path+".source")
	if err != nil {
		return equivalenceAdmission{}, err
	}
	if err := equivalenceInventorySubset(document, environment.Document, path); err != nil {
		return equivalenceAdmission{}, err
	}
	if err := equivalenceDeclarationSubset(document.Parameters, environment.Document.Parameters, path+".parameters"); err != nil {
		return equivalenceAdmission{}, err
	}
	if err := equivalenceDeclarationSubset(document.References, environment.Document.References, path+".references"); err != nil {
		return equivalenceAdmission{}, err
	}
	actual := document
	actual.Sorts = append([]SortDeclaration{}, environment.Document.Sorts...)
	actual.Signatures = cloneEquivalenceSignatures(environment.Document.Signatures)
	actual.Limits = limits
	actualChecked, err := checkEquivalenceDocument(actual, limits, work, path+".selected_context")
	if err != nil {
		return equivalenceAdmission{}, err
	}
	if checked.Status != status.OutcomeOK || actualChecked.Status != status.OutcomeOK {
		encoded, encErr := Encode(sourceDocument)
		if encErr != nil {
			return equivalenceAdmission{}, encErr
		}
		return equivalenceAdmission{checked: actualChecked, fingerprint: equivalenceFingerprint(encoded), limits: limits}, nil
	}
	encoded, err := Encode(sourceDocument)
	if err != nil {
		return equivalenceAdmission{}, err
	}
	return equivalenceAdmission{checked: actualChecked, fingerprint: equivalenceFingerprint(encoded), limits: limits}, nil
}

// checkEquivalenceDocument charges known declaration work before checking and
// lowers the node ceiling to the remaining shared budget for the tree walk.
func checkEquivalenceDocument(document Document, limits Limits, work *equivalenceWork, path string) (CheckedExpression, error) {
	declarations := declarationWork(document)
	if declarations < 0 || declarations >= work.limit-work.used {
		return CheckedExpression{}, resourceLimit(path, "declaration admission leaves no shared work for the expression")
	}
	if err := work.charge(declarations, path+".declarations"); err != nil {
		return CheckedExpression{}, err
	}
	remaining := work.limit - work.used
	if remaining <= 0 {
		return CheckedExpression{}, resourceLimit(path, "no shared work remains for expression admission")
	}
	if limits.MaxNodes > remaining {
		limits.MaxNodes = remaining
	}
	document.Limits = limits
	checked, err := Check(document)
	if err != nil {
		return CheckedExpression{}, err
	}
	if err := work.charge(checked.NodeCount, path+".nodes"); err != nil {
		return checked, err
	}
	return checked, nil
}

func equivalenceBudgetError(err error) (Error, bool) {
	var typed *Error
	if asAlgebraError(err, &typed) && (typed.Status == status.OutcomeResourceLimit || typed.Status == status.OutcomeIncomplete || typed.Status == status.OutcomeUnsupported) {
		return *typed, true
	}
	return Error{}, false
}

func equivalenceInventorySubset(source, actual Document, path string) error {
	actualSorts := make(map[string]SortDeclaration, len(actual.Sorts))
	for _, declaration := range actual.Sorts {
		actualSorts[declaration.ID] = declaration
	}
	for i, declaration := range source.Sorts {
		if _, ok := actualSorts[declaration.ID]; !ok {
			return invalid("UNKNOWN_EQUIVALENCE_SORT", fmt.Sprintf("%s.sorts[%d].id", path, i), "source sort is not in the selected metadata context")
		}
	}
	actualSigs := make(map[string]Signature, len(actual.Signatures))
	for _, signature := range actual.Signatures {
		actualSigs[signature.ID] = signature
	}
	for i, signature := range source.Signatures {
		selected, ok := actualSigs[signature.ID]
		if !ok {
			return invalid("UNKNOWN_EQUIVALENCE_SYMBOL", fmt.Sprintf("%s.signatures[%d].id", path, i), "source signature is not in the selected metadata context")
		}
		if len(signature.Ports) != len(selected.Ports) {
			return invalid("EQUIVALENCE_SIGNATURE_OVERRIDE", fmt.Sprintf("%s.signatures[%d]", path, i), "source signature does not exactly match the selected port inventory")
		}
		for j := range signature.Ports {
			if signature.Ports[j].Role != selected.Ports[j].Role || !TypeEqual(signature.Ports[j].Type, selected.Ports[j].Type) {
				return invalid("EQUIVALENCE_SIGNATURE_OVERRIDE", fmt.Sprintf("%s.signatures[%d].ports[%d]", path, i, j), "source signature role or type overrides the selected metadata")
			}
		}
	}
	return nil
}

func equivalenceDeclarationSubset(source, actual []Declaration, path string) error {
	selected := make(map[string]Type, len(actual))
	for _, declaration := range actual {
		selected[declaration.ID] = declaration.Type
	}
	for i, declaration := range source {
		typ, ok := selected[declaration.ID]
		if !ok {
			return invalid("UNKNOWN_EQUIVALENCE_DECLARATION", fmt.Sprintf("%s[%d].id", path, i), "parameter or reference is not in the selected typed context")
		}
		if !TypeEqual(declaration.Type, typ) {
			return invalid("EQUIVALENCE_DECLARATION_OVERRIDE", fmt.Sprintf("%s[%d].type", path, i), "source declaration overrides its selected type")
		}
	}
	return nil
}

func equivalenceSideEnvironment(environment RewriteEnvironment, source Document, limits Limits) (RewriteEnvironment, error) {
	copy := environment
	copy.Document = environment.Document
	copy.Document.Parameters = cloneEquivalenceDeclarations(source.Parameters)
	copy.Document.References = cloneEquivalenceDeclarations(source.References)
	copy.Document.Closed = source.Closed
	copy.Document.Limits = limits
	if err := ValidateRewriteEnvironment(copy); err != nil {
		return RewriteEnvironment{}, err
	}
	return copy, nil
}

func normalizeEquivalenceSide(environment RewriteEnvironment, profile RewriteProfile, input *Term, limits EquivalenceLimits, work *equivalenceWork, side string) (NormalizationResult, status.OutcomeKind, error) {
	remaining := work.limit - work.used
	if remaining <= 0 {
		return NormalizationResult{}, status.OutcomeResourceLimit, resourceLimit(side, "no logical work remains for normalization")
	}
	phaseWork := remaining
	if phaseWork > DefaultRewriteMaxWorkNodes {
		phaseWork = DefaultRewriteMaxWorkNodes
	}
	budget := RewriteBudget{MaxWorkNodes: phaseWork}
	result, err := Normalize(environment, profile, input, budget)
	if err != nil {
		var typed *Error
		if asAlgebraError(err, &typed) {
			if typed.Status == status.OutcomeInvalid {
				return NormalizationResult{}, typed.Status, err
			}
			return NormalizationResult{Status: typed.Status, Diagnostics: []Error{*typed}, Statistics: RewriteStatistics{}}, typed.Status, nil
		}
		return NormalizationResult{}, status.OutcomeError, err
	}
	if result.Statistics.WorkNodes > phaseWork {
		return NormalizationResult{}, status.OutcomeError, invalid("EQUIVALENCE_WORK_ACCOUNTING", side, "normalizer exceeded its granted phase work")
	}
	if err := work.charge(result.Statistics.WorkNodes, side+".normalize"); err != nil {
		return result, status.OutcomeResourceLimit, err
	}
	return result, result.Status, nil
}

func verifyEquivalenceSide(environment RewriteEnvironment, profile RewriteProfile, input *Term, certificate RewriteCertificate, limits EquivalenceLimits, work *equivalenceWork, side string) (VerificationResult, status.OutcomeKind, error) {
	remaining := work.limit - work.used
	if remaining <= 0 {
		return VerificationResult{}, status.OutcomeResourceLimit, resourceLimit(side, "no logical work remains for independent rewrite verification")
	}
	phaseWork := remaining
	if phaseWork > DefaultRewriteMaxWorkNodes {
		phaseWork = DefaultRewriteMaxWorkNodes
	}
	result, err := VerifyRewriteCertificate(environment, profile, input, certificate, RewriteBudget{MaxWorkNodes: phaseWork})
	if err != nil {
		var typed *Error
		if asAlgebraError(err, &typed) {
			if typed.Status == status.OutcomeInvalid {
				return VerificationResult{}, typed.Status, err
			}
			return VerificationResult{Status: typed.Status, Diagnostics: []Error{*typed}}, typed.Status, nil
		}
		return VerificationResult{}, status.OutcomeError, err
	}
	if result.Statistics.WorkNodes > phaseWork {
		return VerificationResult{}, status.OutcomeError, invalid("EQUIVALENCE_WORK_ACCOUNTING", side, "rewrite verifier exceeded its granted phase work")
	}
	if err := work.charge(result.Statistics.WorkNodes, side+".verify"); err != nil {
		return result, status.OutcomeResourceLimit, err
	}
	return result, result.Status, nil
}

func normalizationReceipt(result NormalizationResult) EquivalenceNormalizationReceipt {
	diagnostics := append([]Error{}, result.Diagnostics...)
	return EquivalenceNormalizationReceipt{Status: result.Status, Normalized: result.Normalized, Statistics: result.Statistics, Diagnostics: diagnostics}
}

func finishEquivalencePartial(result *EquivalenceResult, phase status.OutcomeKind, diagnostics []Error, message string) {
	result.Verdict = VerdictUndetermined
	result.Status = phase
	if result.Status == "" || result.Status == status.OutcomeOK {
		result.Status = status.OutcomeIncomplete
	}
	if len(diagnostics) != 0 {
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
	} else {
		result.Diagnostics = append(result.Diagnostics, Error{Status: result.Status, Code: "EQUIVALENCE_PHASE_INCOMPLETE", Path: "normalization", Message: message})
	}
}

func typedAlgebraError(err error) Error {
	var typed *Error
	if asAlgebraError(err, &typed) {
		return *typed
	}
	return Error{Status: status.OutcomeError, Code: "EQUIVALENCE_ERROR", Path: "equivalence", Message: err.Error()}
}

func asAlgebraError(err error, target **Error) bool {
	for err != nil {
		if typed, ok := err.(*Error); ok {
			*target = typed
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

func equivalenceFingerprint(encoded []byte) string {
	hash := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func intersectEquivalenceTermLimits(left, right Limits) Limits {
	result := left
	if right.MaxNodes < result.MaxNodes {
		result.MaxNodes = right.MaxNodes
	}
	if right.MaxDepth < result.MaxDepth {
		result.MaxDepth = right.MaxDepth
	}
	if right.MaxTypeDepth < result.MaxTypeDepth {
		result.MaxTypeDepth = right.MaxTypeDepth
	}
	if right.MaxSymbols < result.MaxSymbols {
		result.MaxSymbols = right.MaxSymbols
	}
	if right.MaxPorts < result.MaxPorts {
		result.MaxPorts = right.MaxPorts
	}
	return result
}

func declarationWork(document Document) int {
	work := 0
	add := func(value int) {
		if value < 0 || value > MaxEquivalenceWork-work {
			work = MaxEquivalenceWork + 1
			return
		}
		work += value
	}
	add(len(document.Sorts))
	add(len(document.Signatures))
	add(len(document.Parameters))
	add(len(document.References))
	for _, signature := range document.Signatures {
		add(len(signature.Ports))
		if work > MaxEquivalenceWork {
			return work
		}
	}
	return work
}

func cloneEquivalenceProfile(profile RewriteProfile) RewriteProfile {
	return RewriteProfile{ID: profile.ID, Rules: append([]RewriteRule{}, profile.Rules...)}
}

func cloneEquivalenceSignatures(signatures []Signature) []Signature {
	out := make([]Signature, len(signatures))
	for i, signature := range signatures {
		out[i] = Signature{ID: signature.ID, Label: signature.Label, Ports: append([]Port{}, signature.Ports...)}
	}
	return out
}

func cloneEquivalenceDeclarations(declarations []Declaration) []Declaration {
	out := make([]Declaration, len(declarations))
	for i, declaration := range declarations {
		out[i] = Declaration{ID: declaration.ID, Type: cloneEquivalenceType(declaration.Type)}
	}
	return out
}

func cloneEquivalenceType(value Type) Type {
	out := Type{Kind: value.Kind, SortID: value.SortID}
	if value.From != nil {
		copied := cloneEquivalenceType(*value.From)
		out.From = &copied
	}
	if value.To != nil {
		copied := cloneEquivalenceType(*value.To)
		out.To = &copied
	}
	return out
}

func validEquivalenceText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}

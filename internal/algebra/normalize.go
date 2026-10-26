package algebra

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/clickbytes/golingua/internal/status"
)

// RewriteStatistics records deterministic logical node-work units, not CPU
// time. Work units charge validated environment declarations/alias nodes,
// input admission, complete-tree scan nodes, and each applied step's full
// before/after roots. Certificate trace nodes are counted separately.
type RewriteStatistics struct {
	Steps            int `json:"steps"`
	Expansions       int `json:"expansions"`
	TraceNodes       int `json:"trace_nodes"`
	WorkNodes        int `json:"work_nodes"`
	CertificateBytes int `json:"certificate_bytes"`
}

// NormalizationResult contains the last checked root and its valid rewrite
// prefix. Normalized is true only after every licensed phase reaches a
// normal form under the selected profile.
type NormalizationResult struct {
	Status            status.OutcomeKind `json:"status"`
	Normalized        bool               `json:"normalized"`
	CanonicalDocument Document           `json:"canonical_document"`
	Checked           CheckedExpression  `json:"checked"`
	Expression        *Term              `json:"expression"`
	Certificate       RewriteCertificate `json:"certificate"`
	Statistics        RewriteStatistics  `json:"statistics"`
	Diagnostics       []Error            `json:"diagnostics,omitempty"`
}

// VerificationResult distinguishes a valid certified prefix from independently
// established profile-normal-form completion.
type VerificationResult struct {
	Status                status.OutcomeKind `json:"status"`
	CertificateValid      bool               `json:"certificate_valid"`
	NormalizationComplete bool               `json:"normalization_complete"`
	Checked               CheckedExpression  `json:"checked"`
	Statistics            RewriteStatistics  `json:"statistics"`
	Diagnostics           []Error            `json:"diagnostics,omitempty"`
}

// Normalize applies deterministic bottom-up rewrite phases. A blocked rule or
// exhausted budget returns the last valid prefix, never a false normal-form
// claim. Only malformed or invalid inputs return an error.
func Normalize(environment RewriteEnvironment, profile RewriteProfile, input *Term, requested RewriteBudget) (NormalizationResult, error) {
	budget, err := NormalizeRewriteBudget(requested)
	if err != nil {
		return NormalizationResult{}, err
	}
	if err := ValidateRewriteProfile(profile); err != nil {
		return NormalizationResult{}, err
	}
	if err := ValidateRewriteEnvironment(environment); err != nil {
		return NormalizationResult{}, err
	}
	if input == nil {
		return NormalizationResult{}, invalid("MISSING_EXPRESSION", "input", "rewrite input is required")
	}

	limits, err := NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	document := environment.Document
	document.Expression = input
	checked, err := Check(document)
	if err != nil {
		return NormalizationResult{}, err
	}
	if _, err := rewriteDocumentWire(document); err != nil {
		return NormalizationResult{}, err
	}
	inputBytes, err := rewriteJSONSize(input)
	if err != nil {
		return NormalizationResult{}, err
	}
	if inputBytes > MaxRewriteTermBytes {
		return NormalizationResult{}, resourceLimit("input", fmt.Sprintf("rewrite term is %d bytes; maximum is %d", inputBytes, MaxRewriteTermBytes))
	}
	current, err := copyTermDetached(input, limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	certificateInput, err := copyTermDetached(input, limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	rules := make([]RewriteRule, len(profile.Rules))
	copy(rules, profile.Rules)
	certificate := RewriteCertificate{SchemaVersion: 1, SnapshotID: environment.SnapshotID, Profile: RewriteProfile{ID: profile.ID, Rules: rules}, Input: certificateInput, Output: current, Steps: []RewriteStep{}}
	if environment.Document.RequestContext != nil {
		certificate.ContextFingerprint, err = RequestContextFingerprint(*environment.Document.RequestContext, environment.Document)
		if err != nil {
			return NormalizationResult{}, err
		}
	}
	state := normalizerState{environment: environment, profile: profile, budget: budget, limits: limits, certificate: certificate}
	envWork, err := rewriteEnvironmentWork(environment, limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	state.statistics.WorkNodes = envWork + checked.NodeCount
	if checked.Status == status.OutcomeIncomplete || len(checked.Holes) > 0 {
		state.certificate.Output = current
		state.certificate.Normalized = false
		return state.result(status.OutcomeIncomplete, &Error{Status: status.OutcomeIncomplete, Code: "REWRITE_INPUT_INCOMPLETE", Path: "input", Message: "rewrite input contains unresolved holes"}, checked)
	}
	if state.statistics.WorkNodes > budget.MaxWorkNodes {
		return state.partial(status.OutcomeResourceLimit, "REWRITE_WORK_LIMIT", "input admission exceeds the work-node budget", checked)
	}
	if err := state.updateCertificateSize(); err != nil {
		return NormalizationResult{}, err
	}
	if state.statistics.CertificateBytes > MaxRewriteCertificateBytes {
		return NormalizationResult{}, resourceLimit("certificate", "initial rewrite certificate exceeds its byte limit")
	}

	// Phase one repeatedly contracts known beta redexes and supplied aliases.
	if rewriteLicensed(profile, RewriteBetaV1) || rewriteLicensed(profile, RewriteAliasV1) {
		for {
			paths, nodeCount, scanErr := rewritePathsPostorder(current, limits)
			if scanErr != nil {
				return NormalizationResult{}, scanErr
			}
			if !state.chargeScan(nodeCount) {
				return state.partial(status.OutcomeResourceLimit, "REWRITE_WORK_LIMIT", "beta and alias scan exceeds the work-node budget", checked)
			}
			applied := false
			for _, path := range paths {
				node, pathErr := rewriteTermAt(current, path)
				if pathErr != nil {
					return NormalizationResult{}, pathErr
				}
				if rewriteLicensed(profile, RewriteBetaV1) && isBetaCandidate(node) {
					stopped, changed, applyErr := state.tryStep(current, RewriteBetaV1, path)
					if stopped != nil {
						return state.partialFromError(stopped, checked)
					}
					if applyErr != nil {
						return NormalizationResult{}, applyErr
					}
					if changed {
						current = state.certificate.Output
						applied = true
						break
					}
				}
				if rewriteLicensed(profile, RewriteAliasV1) && isAliasCandidate(node, environment.Aliases) {
					stopped, changed, applyErr := state.tryStep(current, RewriteAliasV1, path)
					if stopped != nil {
						return state.partialFromError(stopped, checked)
					}
					if applyErr != nil {
						return NormalizationResult{}, applyErr
					}
					if changed {
						current = state.certificate.Output
						applied = true
						break
					}
				}
			}
			if !applied {
				break
			}
		}
	}

	// Phase two performs one-way And normalization and restarts after each
	// successful local change so that newly exposed nested Ands are reconsidered.
	andRules := []RewriteRule{RewriteAndFlattenV1, RewriteAndOrderV1, RewriteAndDeduplicateV1}
	andLicensed := false
	for _, rule := range andRules {
		if rewriteLicensed(profile, rule) {
			andLicensed = true
		}
	}
	for andLicensed {
		paths, nodeCount, scanErr := rewritePathsPostorder(current, limits)
		if scanErr != nil {
			return NormalizationResult{}, scanErr
		}
		if !state.chargeScan(nodeCount) {
			return state.partial(status.OutcomeResourceLimit, "REWRITE_WORK_LIMIT", "And normalization scan exceeds the work-node budget", checked)
		}
		applied := false
		for _, path := range paths {
			node, pathErr := rewriteTermAt(current, path)
			if pathErr != nil {
				return NormalizationResult{}, pathErr
			}
			if node.Kind != TermAnd {
				continue
			}
			for _, rule := range andRules {
				if !rewriteLicensed(profile, rule) {
					continue
				}
				stopped, changed, applyErr := state.tryStep(current, rule, path)
				if stopped != nil {
					return state.partialFromError(stopped, checked)
				}
				if applyErr != nil {
					return NormalizationResult{}, applyErr
				}
				if changed {
					current = state.certificate.Output
					applied = true
					break
				}
			}
			if applied {
				break
			}
		}
		if !applied {
			break
		}
	}

	// Alpha naming is a final whole-root phase so it cannot perturb earlier
	// structural choices or make path selection depend on source binder names.
	if rewriteLicensed(profile, RewriteAlphaV1) {
		stopped, changed, applyErr := state.tryStep(current, RewriteAlphaV1, nil)
		if stopped != nil {
			return state.partialFromError(stopped, checked)
		}
		if applyErr != nil {
			return NormalizationResult{}, applyErr
		}
		if changed {
			current = state.certificate.Output
		}
	}
	state.certificate.Output = current
	state.certificate.Normalized = true
	finalChecked, err := checkRewriteOutput(environment, current)
	if err != nil {
		return NormalizationResult{}, err
	}
	return state.result(status.OutcomeOK, nil, finalChecked)
}

type normalizerState struct {
	environment RewriteEnvironment
	profile     RewriteProfile
	budget      RewriteBudget
	limits      Limits
	certificate RewriteCertificate
	statistics  RewriteStatistics
}

func (state *normalizerState) tryStep(current *Term, rule RewriteRule, path []int) (*Error, bool, error) {
	step, changed, err := ApplyRewrite(state.environment, state.profile, current, rule, path)
	beforeNodes := 0
	if step.Before != nil {
		if c, e := checkRewriteOutput(state.environment, current); e == nil {
			beforeNodes = c.NodeCount
		} else if err == nil {
			err = e
		}
	}
	if err != nil {
		var failure *Error
		if errors.As(err, &failure) && (failure.Status == status.OutcomeUnsupported || failure.Status == status.OutcomeResourceLimit || failure.Status == status.OutcomeIncomplete) {
			work := state.environmentWork() + beforeNodes
			if !state.addWork(work) {
				return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "rewrite attempt exceeds the work-node budget"}, false, nil
			}
			return failure, false, nil
		}
		return nil, false, err
	}
	work := state.environmentWork() + beforeNodes
	if changed {
		work += step.AfterNodes
	}
	if !state.addWork(work) {
		return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "rewrite attempt exceeds the work-node budget"}, false, nil
	}
	if !changed {
		return nil, false, nil
	}
	if state.statistics.Steps >= state.budget.MaxSteps {
		return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_STEP_LIMIT", Path: "budget.max_steps", Message: "successful rewrite step limit is exhausted"}, false, nil
	}
	if rule == RewriteAliasV1 && state.statistics.Expansions >= state.budget.MaxExpansions {
		return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_EXPANSION_LIMIT", Path: "budget.max_expansions", Message: "alias expansion limit is exhausted"}, false, nil
	}
	for _, term := range []*Term{step.Before, step.After} {
		size, sizeErr := rewriteJSONSize(term)
		if sizeErr != nil {
			return nil, false, sizeErr
		}
		if size > MaxRewriteTermBytes {
			return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_TERM_BYTES", Path: "step", Message: "candidate rewrite root exceeds the per-term byte limit"}, false, nil
		}
	}
	candidateDocument := state.environment.Document
	candidateDocument.Expression = step.After
	if _, wireErr := rewriteDocumentWire(candidateDocument); wireErr != nil {
		var failure *Error
		if errors.As(wireErr, &failure) && failure.Status == status.OutcomeResourceLimit {
			return failure, false, nil
		}
		return nil, false, wireErr
	}
	traceAdd := step.BeforeNodes + step.AfterNodes
	if traceAdd > state.budget.MaxTraceNodes-state.statistics.TraceNodes {
		return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_TRACE_LIMIT", Path: "budget.max_trace_nodes", Message: "candidate trace exceeds the retained trace-node budget"}, false, nil
	}
	candidate := state.certificate
	candidate.Steps = append(append([]RewriteStep(nil), state.certificate.Steps...), step)
	candidate.Output = step.After
	candidate.Normalized = false
	encoded, marshalErr := json.Marshal(candidate)
	if marshalErr != nil {
		if typed := rewriteMarshalerError(marshalErr); typed != nil {
			if typed.Status == status.OutcomeResourceLimit {
				return typed, false, nil
			}
			return nil, false, typed
		}
		return nil, false, invalid("REWRITE_CERTIFICATE_ENCODING", "certificate", marshalErr.Error())
	}
	if len(encoded) > MaxRewriteCertificateBytes {
		return &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_CERTIFICATE_BYTES", Path: "certificate", Message: "candidate certificate exceeds the serialized byte limit"}, false, nil
	}
	state.certificate = candidate
	state.statistics.Steps++
	state.statistics.TraceNodes += traceAdd
	state.statistics.CertificateBytes = len(encoded)
	if rule == RewriteAliasV1 {
		state.statistics.Expansions++
	}
	return nil, true, nil
}

func (state *normalizerState) environmentWork() int {
	value, _ := rewriteEnvironmentWork(state.environment, state.limits)
	return value
}
func (state *normalizerState) addWork(count int) bool {
	if count < 0 || count > state.budget.MaxWorkNodes-state.statistics.WorkNodes {
		return false
	}
	state.statistics.WorkNodes += count
	return true
}
func (state *normalizerState) chargeScan(nodes int) bool { return state.addWork(nodes) }
func (state *normalizerState) updateCertificateSize() error {
	bytes, err := json.Marshal(state.certificate)
	if err != nil {
		if typed := rewriteMarshalerError(err); typed != nil {
			return typed
		}
		return invalid("REWRITE_CERTIFICATE_ENCODING", "certificate", err.Error())
	}
	state.statistics.CertificateBytes = len(bytes)
	return nil
}

func (state *normalizerState) partial(statusValue status.OutcomeKind, code, message string, checked CheckedExpression) (NormalizationResult, error) {
	diagnostic := Error{Status: statusValue, Code: code, Path: "budget", Message: message}
	return state.result(statusValue, &diagnostic, checked)
}
func (state *normalizerState) partialFromError(failure *Error, checked CheckedExpression) (NormalizationResult, error) {
	return state.result(failure.Status, failure, checked)
}
func (state *normalizerState) result(statusValue status.OutcomeKind, diagnostic *Error, checked CheckedExpression) (NormalizationResult, error) {
	state.certificate.Normalized = statusValue == status.OutcomeOK
	documentForCheck := state.environment.Document
	documentForCheck.Expression = state.certificate.Output
	actualChecked, checkErr := Check(documentForCheck)
	if checkErr != nil {
		return NormalizationResult{}, checkErr
	}
	checked = actualChecked
	if err := state.updateCertificateSize(); err != nil {
		return NormalizationResult{}, err
	}
	current, err := copyTermDetached(state.certificate.Output, state.limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	doc, err := cloneRewriteDocument(state.environment.Document, state.limits)
	if err != nil {
		return NormalizationResult{}, err
	}
	doc.Expression = current
	result := NormalizationResult{Status: statusValue, Normalized: state.certificate.Normalized, CanonicalDocument: doc, Checked: checked, Expression: current, Certificate: state.certificate, Statistics: state.statistics}
	if diagnostic != nil {
		result.Diagnostics = []Error{*diagnostic}
	}
	return result, nil
}

// VerifyRewriteCertificate validates a certified prefix and independently
// checks completion. It never invokes Normalize or trusts the producer's
// strategy claim.
func VerifyRewriteCertificate(environment RewriteEnvironment, profile RewriteProfile, originalInput *Term, certificate RewriteCertificate, requested RewriteBudget) (VerificationResult, error) {
	budget, err := NormalizeRewriteBudget(requested)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := ValidateRewriteProfile(profile); err != nil {
		return VerificationResult{}, err
	}
	if err := ValidateRewriteEnvironment(environment); err != nil {
		return VerificationResult{}, err
	}
	if certificate.SnapshotID != environment.SnapshotID || certificate.Profile.ID != profile.ID || !sameRulesExact(certificate.Profile.Rules, profile.Rules) {
		return VerificationResult{}, invalid("REWRITE_CONTEXT_MISMATCH", "certificate", "certificate snapshot and exact rule profile must match verified inputs")
	}
	expectedContextFingerprint := ""
	if environment.Document.RequestContext != nil {
		expectedContextFingerprint, err = RequestContextFingerprint(*environment.Document.RequestContext, environment.Document)
		if err != nil {
			return VerificationResult{}, err
		}
	}
	if certificate.ContextFingerprint != expectedContextFingerprint {
		return VerificationResult{}, invalid("REWRITE_REQUEST_CONTEXT_MISMATCH", "certificate.context_fingerprint", "certificate request context differs from the independently selected context")
	}
	if originalInput == nil {
		return VerificationResult{}, invalid("MISSING_EXPRESSION", "input", "actual original input is required")
	}
	limits, err := NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return VerificationResult{}, err
	}
	inputDocument := environment.Document
	inputDocument.Expression = originalInput
	checkedInput, err := Check(inputDocument)
	if err != nil {
		return VerificationResult{}, err
	}
	if checkedInput.Status != status.OutcomeOK {
		return VerificationResult{Status: status.OutcomeIncomplete, CertificateValid: false, NormalizationComplete: false, Checked: checkedInput, Diagnostics: []Error{{Status: status.OutcomeIncomplete, Code: "REWRITE_INPUT_INCOMPLETE", Path: "input", Message: "rewrite verifier refuses an input with unresolved holes"}}}, nil
	}
	if certificate.SchemaVersion != 1 || certificate.Input == nil || certificate.Output == nil {
		return VerificationResult{}, invalid("INVALID_REWRITE_CERTIFICATE", "certificate", "schema version 1 and full input/output terms are required")
	}
	if len(certificate.SnapshotID) > 256 || len(certificate.ContextFingerprint) > 256 || len(certificate.Profile.ID) > 256 || len(certificate.Profile.Rules) > 6 {
		return VerificationResult{}, invalid("INVALID_REWRITE_CERTIFICATE", "certificate", "snapshot, profile, or rule list exceeds its bound")
	}
	inputBytes, err := rewriteJSONSize(originalInput)
	if err != nil {
		return VerificationResult{}, err
	}
	if inputBytes > MaxRewriteTermBytes {
		return VerificationResult{}, resourceLimit("input", "rewrite input exceeds term byte limit")
	}
	if len(certificate.Steps) > budget.MaxSteps {
		return VerificationResult{}, resourceLimit("certificate.steps", "certificate exceeds the successful step budget")
	}
	wireEstimate := 1024
	termBytes := 0
	admitTerm := func(term *Term, path string) (int, error) {
		if term == nil {
			return 0, invalid("MISSING_REWRITE_TERM", path, "certificate term is required")
		}
		checked, checkErr := checkRewriteOutput(environment, term)
		if checkErr != nil {
			return 0, checkErr
		}
		size, sizeErr := rewriteJSONSize(term)
		if sizeErr != nil {
			return 0, sizeErr
		}
		if size > MaxRewriteTermBytes {
			return 0, resourceLimit(path, "certificate term exceeds per-term byte limit")
		}
		if size > MaxRewriteCertificateBytes-termBytes {
			return 0, resourceLimit("certificate", "aggregate certificate terms exceed serialized byte limit")
		}
		termBytes += size
		wireEstimate += size + 8
		if wireEstimate > MaxRewriteCertificateBytes {
			return 0, resourceLimit("certificate", "certificate exceeds serialized byte limit")
		}
		return checked.NodeCount, nil
	}
	if _, err := admitTerm(certificate.Input, "certificate.input"); err != nil {
		return VerificationResult{}, err
	}
	if _, err := admitTerm(certificate.Output, "certificate.output"); err != nil {
		return VerificationResult{}, err
	}
	addJSONString := func(value, path string, max int) error {
		if !validOptionalText(value, max) {
			return invalid("INVALID_REWRITE_CERTIFICATE", path, "certificate text field exceeds its bound")
		}
		encoded, encodeErr := json.Marshal(value)
		if encodeErr != nil {
			return invalid("REWRITE_CERTIFICATE_ENCODING", path, encodeErr.Error())
		}
		wireEstimate += len(encoded) + 2
		if wireEstimate > MaxRewriteCertificateBytes {
			return resourceLimit("certificate", "certificate metadata exceeds serialized byte limit")
		}
		return nil
	}
	if err := addJSONString(certificate.SnapshotID, "certificate.snapshot_id", 256); err != nil {
		return VerificationResult{}, err
	}
	if err := addJSONString(certificate.ContextFingerprint, "certificate.context_fingerprint", 256); err != nil {
		return VerificationResult{}, err
	}
	if err := addJSONString(certificate.Profile.ID, "certificate.profile.id", 256); err != nil {
		return VerificationResult{}, err
	}
	for i, rule := range certificate.Profile.Rules {
		if err := addJSONString(string(rule), fmt.Sprintf("certificate.profile.rules[%d]", i), 32); err != nil {
			return VerificationResult{}, err
		}
	}
	equal, err := RewriteTermEqual(originalInput, certificate.Input, limits)
	if err != nil {
		return VerificationResult{}, err
	}
	if !equal {
		return VerificationResult{}, invalid("REWRITE_INPUT_MISMATCH", "certificate.input", "certificate input differs from the actual original input")
	}
	current, err := copyTermDetached(originalInput, limits)
	if err != nil {
		return VerificationResult{}, err
	}
	workEnv, err := rewriteEnvironmentWork(environment, limits)
	if err != nil {
		return VerificationResult{}, err
	}
	stats := RewriteStatistics{WorkNodes: workEnv + checkedInput.NodeCount}
	if stats.WorkNodes > budget.MaxWorkNodes {
		return VerificationResult{Status: status.OutcomeResourceLimit, CertificateValid: false, Diagnostics: []Error{{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "verification admission exceeds work budget"}}}, nil
	}
	for i, step := range certificate.Steps {
		stepPath := fmt.Sprintf("certificate.steps[%d]", i)
		if len(step.RuleID) > 32 || len(step.SnapshotID) > 256 || len(step.ProfileID) > 256 {
			return VerificationResult{}, invalid("INVALID_REWRITE_CERTIFICATE", stepPath, "step identifiers exceed their bounds")
		}
		if len(step.Path) > limits.MaxDepth {
			return VerificationResult{}, resourceLimit(stepPath+".path", "step path exceeds expression depth limit")
		}
		if len(step.BinderMapping) > 2*limits.MaxNodes || len(step.SideConditions) > 2*limits.MaxNodes {
			return VerificationResult{}, resourceLimit(stepPath, "step metadata exceeds bounded size")
		}
		if len(step.ParameterMapping) != 0 {
			return VerificationResult{}, invalid("REWRITE_PARAMETER_MAPPING", stepPath+".parameter_mapping", "rewrite rules in this kernel do not substitute free parameters")
		}
		if step.BeforeNodes < 0 || step.BeforeNodes > limits.MaxNodes || step.AfterNodes < 0 || step.AfterNodes > limits.MaxNodes {
			return VerificationResult{}, invalid("INVALID_REWRITE_NODE_COUNT", stepPath, "step node counts are outside the checked root bound")
		}
		wireEstimate += 512
		for _, metadata := range []struct {
			value, path string
			max         int
		}{
			{string(step.RuleID), stepPath + ".rule_id", 32},
			{step.SnapshotID, stepPath + ".snapshot_id", 256},
			{step.ProfileID, stepPath + ".profile_id", 256},
		} {
			if err := addJSONString(metadata.value, metadata.path, metadata.max); err != nil {
				return VerificationResult{}, err
			}
		}
		for _, index := range step.Path {
			if index < 0 {
				return VerificationResult{}, invalid("INVALID_REWRITE_PATH", stepPath+".path", "path indexes must be non-negative")
			}
			wireEstimate += len(strconv.Itoa(index)) + 2
		}
		for _, change := range step.BinderMapping {
			for _, field := range []struct {
				name, value string
				max         int
			}{{"before", change.Before, 256}, {"after", change.After, 256}, {"kind", change.Kind, 32}} {
				if err := addJSONString(field.value, stepPath+".binder_mapping."+field.name, field.max); err != nil {
					return VerificationResult{}, err
				}
			}
		}
		for _, condition := range step.SideConditions {
			if err := addJSONString(condition, stepPath+".side_conditions", 4096); err != nil {
				return VerificationResult{}, err
			}
		}
		if wireEstimate > MaxRewriteCertificateBytes {
			return VerificationResult{}, resourceLimit("certificate", "certificate exceeds serialized byte limit")
		}
		if _, err := admitTerm(step.Before, stepPath+".before"); err != nil {
			return VerificationResult{}, err
		}
		if _, err := admitTerm(step.After, stepPath+".after"); err != nil {
			return VerificationResult{}, err
		}
		beforeEqual, cmpErr := RewriteTermEqual(current, step.Before, limits)
		if cmpErr != nil {
			return VerificationResult{}, cmpErr
		}
		if !beforeEqual {
			return VerificationResult{}, invalid("REWRITE_CHAIN_GAP", fmt.Sprintf("certificate.steps[%d].before", i), "step input does not equal the preceding root")
		}
		if step.SnapshotID != environment.SnapshotID || step.ProfileID != profile.ID {
			return VerificationResult{}, invalid("REWRITE_CONTEXT_MISMATCH", fmt.Sprintf("certificate.steps[%d]", i), "step snapshot or profile differs from verification context")
		}
		if err := VerifyRewriteStep(environment, profile, step); err != nil {
			return VerificationResult{}, err
		}
		stepWork := workEnv + step.BeforeNodes + step.AfterNodes
		if !verificationCharge(&stats, budget.MaxWorkNodes, stepWork) {
			return VerificationResult{Status: status.OutcomeResourceLimit, CertificateValid: false, Statistics: stats, Diagnostics: []Error{{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: fmt.Sprintf("certificate.steps[%d]", i), Message: "step verification exceeds work budget"}}}, nil
		}
		traceAdd := step.BeforeNodes + step.AfterNodes
		if traceAdd > budget.MaxTraceNodes-stats.TraceNodes {
			return VerificationResult{}, resourceLimit("certificate.steps", "certificate trace exceeds trace-node budget")
		}
		stats.TraceNodes += traceAdd
		stats.Steps++
		if step.RuleID == RewriteAliasV1 {
			stats.Expansions++
			if stats.Expansions > budget.MaxExpansions {
				return VerificationResult{}, resourceLimit("certificate.steps", "certificate alias expansions exceed budget")
			}
		}
		current = step.After
	}
	outputEqual, err := RewriteTermEqual(current, certificate.Output, limits)
	if err != nil {
		return VerificationResult{}, err
	}
	if !outputEqual {
		return VerificationResult{}, invalid("REWRITE_OUTPUT_MISMATCH", "certificate.output", "certificate output differs from the final step root")
	}
	// Only after all complete roots and one-step claims have been bounded and
	// replayed may the aggregate object be serialized.
	encoded, err := json.Marshal(certificate)
	if err != nil {
		if typed := rewriteMarshalerError(err); typed != nil {
			return VerificationResult{}, typed
		}
		return VerificationResult{}, invalid("REWRITE_CERTIFICATE_ENCODING", "certificate", err.Error())
	}
	if len(encoded) > MaxRewriteCertificateBytes {
		return VerificationResult{}, resourceLimit("certificate", "certificate exceeds serialized byte limit")
	}
	stats.CertificateBytes = len(encoded)
	checkedFinal, err := checkRewriteOutput(environment, current)
	if err != nil {
		return VerificationResult{}, err
	}
	if !certificate.Normalized {
		return VerificationResult{Status: status.OutcomeIncomplete, CertificateValid: true, NormalizationComplete: false, Checked: checkedFinal, Statistics: stats, Diagnostics: []Error{{Status: status.OutcomeIncomplete, Code: "VALID_PREFIX_ONLY", Path: "certificate.normalized", Message: "rewrite prefix is valid but completion was not claimed"}}}, nil
	}
	complete, diag, scanWork, checkErr := independentlyNormalForm(environment, profile, current, limits, budget, stats.WorkNodes)
	if checkErr != nil {
		return VerificationResult{}, checkErr
	}
	stats.WorkNodes = scanWork
	if diag != nil {
		return VerificationResult{Status: diag.Status, CertificateValid: true, NormalizationComplete: false, Checked: checkedFinal, Statistics: stats, Diagnostics: []Error{*diag}}, nil
	}
	if !complete {
		return VerificationResult{}, invalid("CERTIFICATE_NOT_NORMALIZED", "certificate.normalized", "a licensed rewrite remains applicable to the output")
	}
	return VerificationResult{Status: status.OutcomeOK, CertificateValid: true, NormalizationComplete: true, Checked: checkedFinal, Statistics: stats}, nil
}

func independentlyNormalForm(environment RewriteEnvironment, profile RewriteProfile, root *Term, limits Limits, budget RewriteBudget, work int) (bool, *Error, int, error) {
	paths, nodes, err := rewritePathsPostorder(root, limits)
	if err != nil {
		return false, nil, work, err
	}
	if nodes > budget.MaxWorkNodes-work {
		return false, &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "normal-form inspection exceeds work budget"}, work, nil
	}
	work += nodes
	envWork, _ := rewriteEnvironmentWork(environment, limits)
	rules := []RewriteRule{RewriteBetaV1, RewriteAliasV1, RewriteAndFlattenV1, RewriteAndOrderV1, RewriteAndDeduplicateV1}
	for _, path := range paths {
		node, err := rewriteTermAt(root, path)
		if err != nil {
			return false, nil, work, err
		}
		for _, rule := range rules {
			if !rewriteLicensed(profile, rule) || !rewriteCandidateFor(rule, node, environment.Aliases) {
				continue
			}
			step, changed, applyErr := ApplyRewrite(environment, profile, root, rule, path)
			cost := envWork
			before, _ := Check(Document{SchemaVersion: environment.Document.SchemaVersion, Sorts: environment.Document.Sorts, Signatures: environment.Document.Signatures, Parameters: environment.Document.Parameters, References: environment.Document.References, Closed: environment.Document.Closed, Expression: root, Limits: environment.Document.Limits})
			cost += before.NodeCount
			if changed {
				cost += step.AfterNodes
			}
			if cost > budget.MaxWorkNodes-work {
				return false, &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "normal-form candidate inspection exceeds work budget"}, work, nil
			}
			work += cost
			if applyErr != nil {
				var failure *Error
				if errors.As(applyErr, &failure) && (failure.Status == status.OutcomeUnsupported || failure.Status == status.OutcomeResourceLimit || failure.Status == status.OutcomeIncomplete) {
					return false, failure, work, nil
				}
				return false, nil, work, applyErr
			}
			if changed {
				return false, nil, work, nil
			}
		}
	}
	if rewriteLicensed(profile, RewriteAlphaV1) {
		step, changed, applyErr := ApplyRewrite(environment, profile, root, RewriteAlphaV1, nil)
		before, _ := Check(Document{SchemaVersion: environment.Document.SchemaVersion, Sorts: environment.Document.Sorts, Signatures: environment.Document.Signatures, Parameters: environment.Document.Parameters, References: environment.Document.References, Closed: environment.Document.Closed, Expression: root, Limits: environment.Document.Limits})
		cost := envWork + before.NodeCount
		if changed {
			cost += step.AfterNodes
		}
		if cost > budget.MaxWorkNodes-work {
			return false, &Error{Status: status.OutcomeResourceLimit, Code: "REWRITE_WORK_LIMIT", Path: "budget.max_work_nodes", Message: "alpha normal-form inspection exceeds work budget"}, work, nil
		}
		work += cost
		if applyErr != nil {
			return false, nil, work, applyErr
		}
		if changed {
			return false, nil, work, nil
		}
	}
	return true, nil, work, nil
}

func rewriteCandidateFor(rule RewriteRule, node *Term, aliases []AliasDefinition) bool {
	switch rule {
	case RewriteBetaV1:
		return isBetaCandidate(node)
	case RewriteAliasV1:
		return isAliasCandidate(node, aliases)
	case RewriteAndFlattenV1, RewriteAndOrderV1, RewriteAndDeduplicateV1:
		return node != nil && node.Kind == TermAnd
	}
	return false
}
func isBetaCandidate(node *Term) bool {
	return node != nil && node.Kind == TermApply && node.Function != nil && node.Function.Kind == TermLambda
}
func isAliasCandidate(node *Term, aliases []AliasDefinition) bool {
	if node == nil || node.Kind != TermAtom {
		return false
	}
	_, ok := rewriteAlias(aliases, node.Symbol)
	return ok
}

func rewritePathsPostorder(root *Term, limits Limits) ([][]int, int, error) {
	paths := make([][]int, 0)
	active := make(map[*Term]bool)
	nodes := 0
	var walk func(*Term, []int, int) error
	walk = func(node *Term, path []int, depth int) error {
		if node == nil {
			return invalid("MISSING_TERM", "expression", "nil child in checked rewrite tree")
		}
		if depth > limits.MaxDepth {
			return resourceLimit("expression", "rewrite traversal exceeds depth limit")
		}
		if active[node] {
			return invalid("CYCLIC_TERM", "expression", "rewrite traversal found a pointer cycle")
		}
		nodes++
		if nodes > limits.MaxNodes {
			return resourceLimit("expression", "rewrite traversal exceeds node limit")
		}
		active[node] = true
		children := rewriteChildren(node)
		for i, child := range children {
			next := append(append([]int(nil), path...), i)
			if err := walk(child, next, depth+1); err != nil {
				return err
			}
		}
		delete(active, node)
		paths = append(paths, append([]int(nil), path...))
		return nil
	}
	if err := walk(root, nil, 1); err != nil {
		return nil, 0, err
	}
	return paths, nodes, nil
}

func rewriteEnvironmentWork(environment RewriteEnvironment, limits Limits) (int, error) {
	work := len(environment.Document.Sorts) + len(environment.Document.Signatures) + len(environment.Document.Parameters) + len(environment.Document.References)
	for _, signature := range environment.Document.Signatures {
		work += len(signature.Ports)
	}
	for _, alias := range environment.Aliases {
		_, err := inspectTerm(alias.Expression, limits)
		if err != nil {
			return 0, err
		}
		work += 1 + infoNodeCount(alias.Expression) + len(alias.Ports) + len(alias.Interface)
	}
	return work, nil
}

func infoNodeCount(root *Term) int {
	stack := []*Term{root}
	seen := make(map[*Term]bool)
	count := 0
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == nil || seen[node] {
			continue
		}
		seen[node] = true
		count++
		stack = append(stack, rewriteChildren(node)...)
	}
	return count
}

func rewriteJSONSize(value any) (int, error) {
	if term, ok := value.(*Term); ok {
		wire, err := rewriteTermWire(term, "term")
		if err != nil {
			return 0, err
		}
		encoded, err := json.Marshal(wire)
		if err != nil {
			return 0, invalid("REWRITE_JSON_ENCODING", "term", err.Error())
		}
		return len(encoded), nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return 0, invalid("REWRITE_JSON_ENCODING", "term", err.Error())
	}
	return len(encoded), nil
}

func rewriteMarshalerError(err error) *Error {
	var typed *Error
	if errors.As(err, &typed) {
		return typed
	}
	return nil
}

func cloneRewriteDocument(source Document, limits Limits) (Document, error) {
	copy := source
	copy.Sorts = append([]SortDeclaration(nil), source.Sorts...)
	copy.Signatures = make([]Signature, len(source.Signatures))
	for i, signature := range source.Signatures {
		copy.Signatures[i] = signature
		copy.Signatures[i].Ports = make([]Port, len(signature.Ports))
		for j, port := range signature.Ports {
			copy.Signatures[i].Ports[j] = port
			copy.Signatures[i].Ports[j].Type = *cloneTypeValue(port.Type)
		}
	}
	cloneDeclarations := func(values []Declaration) []Declaration {
		out := make([]Declaration, len(values))
		for i, value := range values {
			out[i] = value
			out[i].Type = *cloneTypeValue(value.Type)
		}
		return out
	}
	copy.Parameters = cloneDeclarations(source.Parameters)
	copy.References = cloneDeclarations(source.References)
	if source.RequestContext != nil {
		contextCopy, err := CloneRequestContext(*source.RequestContext)
		if err != nil {
			return Document{}, err
		}
		copy.RequestContext = &contextCopy
	}
	copy.Expression = nil
	return copy, nil
}

func verificationCharge(statistics *RewriteStatistics, limit, amount int) bool {
	if amount < 0 || amount > limit-statistics.WorkNodes {
		return false
	}
	statistics.WorkNodes += amount
	return true
}
func sameRulesExact(a, b []RewriteRule) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Ensure this file's typed failures remain ordinary errors rather than an
// unchecked panic if a caller supplies a malformed certificate object.
var _ error = (*Error)(nil)

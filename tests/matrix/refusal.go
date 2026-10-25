package matrix

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/status"
)

type typedOutcome interface{ Result() status.Result }

// RunRefusalAnalysis observes real source adapters for declared boundary input.
func RunRefusalAnalysis(t *testing.T, lang features.LanguageID, cases []RefusalCase) {
	t.Helper()
	runtime := Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(lang)
	if !ok || analyzer == nil {
		t.Fatalf("INFRA: source analyzer for %q is unavailable", lang)
	}
	for _, testCase := range cases {
		c := testCase
		t.Run(c.ID, func(t *testing.T) {
			if c.ID == "" || c.Language != string(lang) || c.InputOrigin == "" || c.OwnerAcceptance == "" || c.Requirement == "" || hasUnknownAllowance(c) && strings.TrimSpace(c.UnknownAllowanceReason) == "" {
				t.Fatalf("INFRA: invalid refusal declaration: %+v", c)
			}
			result, err := analyzer.Analyze(context.Background(), c.Input)
			logObservation(t, map[string]any{"check_class": "refusal", "stage": "analysis", "case_id": c.ID, "language": lang, "input": c.Input, "requirement": c.Requirement, "origin": c.InputOrigin, "owner_acceptance": c.OwnerAcceptance, "result": result, "error": errorText(err), "typed_error_result": refusalErrorResult(err)})
			if err != nil {
				var outcome typedOutcome
				if !errors.As(err, &outcome) {
					t.Errorf("INFRA: source analyzer returned an untyped error: %v", err)
					return
				}
				typed := outcome.Result()
				if infraOutcome(typed.Status) || infraDiagnostics(typed.Diagnostics) {
					t.Errorf("INFRA: source analyzer error: %v", err)
					return
				}
				if !acceptedRefusalStatus(typed.Status, c.AllowUnknownAnalysis, c.UnknownAllowanceReason, typed.Diagnostics) {
					t.Errorf("FAIL: source analyzer error is not an allowed refusal: %v", err)
					return
				}
				if len(result.Candidates) != 0 {
					t.Errorf("FAIL: invalid refusal error also returned candidates")
				}
				return
			}
			if infraOutcome(result.Outcome) || infraDiagnostics(result.Diagnostics) {
				t.Errorf("INFRA: source analysis exhausted resources or runtime is unavailable")
			}
			unknownAllowed := result.Outcome == status.OutcomeUnknown && c.AllowUnknownAnalysis && acceptedRefusalStatus(result.Outcome, true, c.UnknownAllowanceReason, result.Diagnostics)
			if !acceptedRefusalStatus(result.Outcome, c.AllowUnknownAnalysis, c.UnknownAllowanceReason, result.Diagnostics) {
				t.Errorf("FAIL: boundary input was not refused; source outcome=%s", result.Outcome)
			}
			for _, candidate := range result.Candidates {
				if candidate.Meaning != nil {
					t.Errorf("FAIL: refusal fabricated a Meaning candidate %q", candidate.ID)
				}
				if infraOutcome(candidate.Status) {
					t.Errorf("INFRA: candidate %q exhausted resources or failed", candidate.ID)
				}
				if !refusalOutcome(candidate.Status) && !(candidate.Status == status.OutcomeUnknown && unknownAllowed) {
					t.Errorf("FAIL: refusal retained candidate %q with non-refusal status %s", candidate.ID, candidate.Status)
				}
			}
		})
	}
}

// RunRefusalTranslation observes the production translation pipeline and
// keeps valid-but-unimplemented or partial analyses from counting as refusals.
func RunRefusalTranslation(t *testing.T, from, to features.LanguageID, cases []RefusalCase) {
	t.Helper()
	runtime := Runtime(t)
	for _, testCase := range cases {
		c := testCase
		t.Run(c.ID, func(t *testing.T) {
			if c.ID == "" || c.Language != string(from) || c.InputOrigin == "" || c.OwnerAcceptance == "" || c.Requirement == "" || hasUnknownAllowance(c) && strings.TrimSpace(c.UnknownAllowanceReason) == "" {
				t.Fatalf("INFRA: invalid refusal declaration: %+v", c)
			}
			result, err := runtime.Translate(context.Background(), engine.Request{From: from, To: to, Text: c.Input})
			logObservation(t, map[string]any{"check_class": "refusal", "stage": "translation", "case_id": c.ID, "from": from, "to": to, "input": c.Input, "requirement": c.Requirement, "origin": c.InputOrigin, "owner_acceptance": c.OwnerAcceptance, "result": result, "error": errorText(err), "typed_error_result": refusalErrorResult(err)})
			if err != nil {
				var outcome typedOutcome
				if !errors.As(err, &outcome) {
					t.Errorf("INFRA: translation returned an untyped error: %v", err)
					return
				}
				typed := outcome.Result()
				if infraOutcome(typed.Status) || infraDiagnostics(typed.Diagnostics) {
					t.Errorf("INFRA: translation error: %v", err)
					return
				}
				if !acceptedRefusalStatus(typed.Status, c.AllowUnknownTranslation, c.UnknownAllowanceReason, typed.Diagnostics) {
					t.Errorf("FAIL: translation error is not an allowed refusal: %v", err)
					return
				}
				if result.Text != "" || len(result.Candidates) != 0 {
					t.Errorf("FAIL: invalid refusal error also returned output")
				}
				return
			}
			if infraOutcome(result.Status) || infraDiagnostics(result.Diagnostics) {
				t.Errorf("INFRA: translation exhausted resources or runtime is unavailable")
			}
			unknownAllowed := result.Status == status.OutcomeUnknown && c.AllowUnknownTranslation && acceptedRefusalStatus(result.Status, true, c.UnknownAllowanceReason, result.Diagnostics)
			if !acceptedRefusalStatus(result.Status, c.AllowUnknownTranslation, c.UnknownAllowanceReason, result.Diagnostics) {
				t.Errorf("FAIL: boundary input was not refused; translation status=%s", result.Status)
			}
			if result.Text != "" {
				t.Errorf("FAIL: refusal returned translated text %q", result.Text)
			}
			for _, candidate := range result.Candidates {
				if candidate.Meaning != nil {
					t.Errorf("FAIL: refusal fabricated a Meaning in candidate %q", candidate.ID)
				}
				if strings.TrimSpace(candidate.Text) != "" {
					t.Errorf("FAIL: refusal returned candidate text %q", candidate.Text)
				}
				if infraOutcome(candidate.Outcome) {
					t.Errorf("INFRA: candidate %q exhausted resources or failed", candidate.ID)
				}
				if !refusalOutcome(candidate.Outcome) && !(candidate.Outcome == status.OutcomeUnknown && unknownAllowed) {
					t.Errorf("FAIL: refusal retained candidate %q with non-refusal status %s", candidate.ID, candidate.Outcome)
				}
			}
		})
	}
}

func refusalOutcome(outcome status.OutcomeKind) bool {
	return outcome == status.OutcomeUnsupported || outcome == status.OutcomeInvalid
}

func acceptedRefusalStatus(outcome status.OutcomeKind, allowUnknown bool, reason string, diagnostics []status.Diagnostic) bool {
	if refusalOutcome(outcome) {
		return len(diagnostics) > 0
	}
	return outcome == status.OutcomeUnknown && allowUnknown && strings.TrimSpace(reason) != "" && hasDiagnosticReason(diagnostics)
}

func hasUnknownAllowance(c RefusalCase) bool {
	return c.AllowUnknownAnalysis || c.AllowUnknownTranslation
}

func infraOutcome(outcome status.OutcomeKind) bool {
	return outcome == status.OutcomeResourceLimit || outcome == status.OutcomeError
}

func infraDiagnostics(diagnostics []status.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		code := strings.ToUpper(diagnostic.Code)
		if strings.Contains(code, "UNAVAILABLE") || strings.Contains(code, "RESOURCE_LIMIT") || strings.Contains(code, "CANCELED") || strings.Contains(code, "TIMEOUT") || strings.Contains(code, "RUNTIME") || strings.Contains(code, "PIPELINE") || strings.Contains(code, "ADAPTER") || strings.Contains(code, "ENGINE") {
			return true
		}
	}
	return false
}

func hasDiagnosticReason(diagnostics []status.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if strings.TrimSpace(diagnostic.Code) != "" && strings.TrimSpace(diagnostic.Message) != "" {
			return true
		}
	}
	return false
}

func refusalErrorResult(err error) any {
	var outcome typedOutcome
	if errors.As(err, &outcome) {
		return outcome.Result()
	}
	return nil
}

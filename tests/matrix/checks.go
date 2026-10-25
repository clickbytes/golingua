package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
)

var (
	runtimeOnce  sync.Once
	sharedEngine *engine.Engine
	runtimeErr   error
)

// Runtime returns the production-backed immutable language engine shared by
// matrix checks. Setup failures are infrastructure errors, not language gaps.
func Runtime(t testing.TB) *engine.Engine {
	t.Helper()
	runtimeOnce.Do(func() { sharedEngine, runtimeErr = NewRuntime() })
	if runtimeErr != nil {
		t.Fatalf("INFRA: compile matrix runtime: %v", runtimeErr)
	}
	return sharedEngine
}

// RunAnalysis independently checks source-language adapters against authored
// Meaning fixtures. Cases without fixtures remain visible as SPEC_PENDING.
func RunAnalysis(t *testing.T, lang features.LanguageID, cases []AnalysisCase) {
	t.Helper()
	if len(cases) == 0 {
		t.Errorf("SPEC_PENDING: no analysis cases were registered for %q", lang)
		return
	}
	for _, testCase := range cases {
		c := testCase
		t.Run(c.ID, func(t *testing.T) {
			t.Helper()
			want, identification, ok, err := loadFixture(c.MeaningID)
			if err != nil {
				t.Fatalf("INFRA: build fixture %q: %v", c.MeaningID, err)
			}
			if !ok {
				t.Errorf("SPEC_PENDING / NOT RUN: no independent Meaning fixture for %q", c.MeaningID)
				return
			}
			runtime := Runtime(t)
			if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("INFRA: fixture %q does not validate against CORE: %v", c.MeaningID, err)
			}
			analyzer, exists := runtime.SourceAnalyzerFor(lang)
			if !exists || analyzer == nil {
				t.Fatalf("INFRA: source analyzer for %q is unavailable", lang)
			}
			result, analyzeErr := analyzer.Analyze(context.Background(), c.Input)
			logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "result": result, "error": errorText(analyzeErr)})
			if analyzeErr != nil {
				if infraError(analyzeErr) {
					t.Errorf("INFRA: source analysis %q: %v", c.ID, analyzeErr)
				} else {
					t.Errorf("GAP_ANALYZE: source analysis %q failed: %v", c.ID, analyzeErr)
				}
				return
			}
			if result.Outcome == status.OutcomeUnsupported || len(result.Candidates) == 0 {
				t.Errorf("GAP_ANALYZE: source adapter did not produce a supported candidate (status %s)", result.Outcome)
				return
			}
			for _, candidate := range result.Candidates {
				if candidate.Status == status.OutcomeUnsupported || candidate.Meaning == nil {
					t.Errorf("GAP_ANALYZE: candidate %q has no supported Meaning (status %s)", candidate.ID, candidate.Status)
					continue
				}
				if candidate.Status != status.OutcomeOK {
					t.Errorf("FAIL: candidate %q has status %s", candidate.ID, candidate.Status)
					continue
				}
				if identification != nil {
					matches, compareErr := IdentificationMatches(*identification, candidate)
					if compareErr != nil || !matches {
						t.Errorf("FAIL: candidate %q does not match identification contract (matches=%t, error=%v)", candidate.ID, matches, compareErr)
					}
				} else {
					matches, compareErr := GraphMatches(want, *candidate.Meaning)
					if compareErr != nil || !matches {
						t.Errorf("FAIL: candidate %q Meaning does not match fixture (matches=%t, error=%v)", candidate.ID, matches, compareErr)
					}
					for _, context := range candidate.Contexts {
						for _, mention := range context.Context.Mentions {
							if mention.ReferentID != "" {
								t.Errorf("FAIL: candidate %q resolves identity without a fixture for it", candidate.ID)
							}
						}
					}
				}
			}
			if result.Outcome != status.OutcomeOK && result.Outcome != status.OutcomeAmbiguous {
				t.Errorf("FAIL: source analysis returned status %s", result.Outcome)
			}
		})
	}
}

// RunRealization checks target adapters with a fresh, authored Meaning and no
// source-language analysis call.
func RunRealization(t *testing.T, lang features.LanguageID, cases []RealizationCase) {
	t.Helper()
	if len(cases) == 0 {
		t.Errorf("SPEC_PENDING: no realization cases were registered for %q", lang)
		return
	}
	for _, testCase := range cases {
		c := testCase
		t.Run(c.ID, func(t *testing.T) {
			t.Helper()
			want, identification, ok, err := loadFixture(c.MeaningID)
			if err != nil {
				t.Fatalf("INFRA: build fixture %q: %v", c.MeaningID, err)
			}
			if !ok {
				t.Errorf("SPEC_PENDING / NOT RUN: no independent Meaning fixture for %q", c.MeaningID)
				return
			}
			runtime := Runtime(t)
			if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("INFRA: fixture %q does not validate against CORE: %v", c.MeaningID, err)
			}
			realizer, exists := runtime.TargetRealizerFor(lang)
			if !exists || realizer == nil {
				t.Fatalf("INFRA: target realizer for %q is unavailable", lang)
			}
			candidate := semantic.MeaningCandidate{
				ID: c.MeaningID, SemanticID: c.MeaningID, Meaning: &want,
				Status:   status.OutcomeOK,
				Evidence: semantic.EvidenceTrace{AlternativeID: c.MeaningID, Outcome: status.OutcomeOK},
			}
			if identification != nil {
				candidate.Contexts = []semantic.MeaningContext{{Meaning: want, Context: identification.Context}}
			}
			result, realizeErr := realizer.Realize(context.Background(), candidate)
			logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "result": result, "error": errorText(realizeErr)})
			if realizeErr != nil {
				if infraError(realizeErr) {
					t.Errorf("INFRA: target realization %q: %v", c.ID, realizeErr)
				} else {
					t.Errorf("GAP_REALIZE: target realization %q failed: %v", c.ID, realizeErr)
				}
				return
			}
			checkSurfaces(t, c, result.Outcome, result.Alternatives)
		})
	}
}

// RunTranslation checks one integrated engine request against the same
// independent I fixture and target-side allowance.
func RunTranslation(t *testing.T, c TranslationCase, analysis AnalysisCase, realization RealizationCase) {
	t.Helper()
	t.Run(c.ID, func(t *testing.T) {
		t.Helper()
		if c.AnalysisID != analysis.ID || c.RealizationID != realization.ID {
			t.Fatalf("SPEC_PENDING: translation references do not match supplied cases")
		}
		want, identification, ok, err := loadFixture(analysis.MeaningID)
		if err != nil {
			t.Fatalf("INFRA: build fixture %q: %v", analysis.MeaningID, err)
		}
		if !ok || realization.MeaningID != analysis.MeaningID {
			t.Errorf("SPEC_PENDING / NOT RUN: translation lacks a shared independent Meaning fixture")
			return
		}
		runtime := Runtime(t)
		if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			t.Fatalf("INFRA: fixture %q does not validate against CORE: %v", analysis.MeaningID, err)
		}
		sourceGap := false
		if identification != nil {
			analyzer, exists := runtime.SourceAnalyzerFor(features.LanguageID(c.From))
			if !exists || analyzer == nil {
				t.Fatalf("INFRA: source analyzer for %q is unavailable", c.From)
			}
			analysisResult, analysisErr := analyzer.Analyze(context.Background(), analysis.Input)
			logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "stage": "identification_source_analysis", "result": analysisResult, "error": errorText(analysisErr)})
			if analysisErr != nil {
				if infraError(analysisErr) {
					t.Errorf("INFRA: source analysis %q: %v", c.ID, analysisErr)
					return
				} else {
					sourceGap = true
					t.Errorf("GAP_ANALYZE: source analysis %q failed: %v", c.ID, analysisErr)
				}
			} else if analysisResult.Outcome == status.OutcomeUnsupported || len(analysisResult.Candidates) == 0 {
				sourceGap = true
				t.Errorf("GAP_ANALYZE: source analysis %q returned %s without candidates", c.ID, analysisResult.Outcome)
			} else {
				for _, candidate := range analysisResult.Candidates {
					if candidate.Status == status.OutcomeUnsupported || candidate.Meaning == nil {
						sourceGap = true
						t.Errorf("GAP_ANALYZE: source candidate %q has status %s and no supported Meaning", candidate.ID, candidate.Status)
						continue
					}
					if candidate.Status != status.OutcomeOK {
						t.Errorf("FAIL: source candidate %q has status %s", candidate.ID, candidate.Status)
						continue
					}
					if matches, matchErr := IdentificationMatches(*identification, candidate); matchErr != nil || !matches {
						t.Errorf("FAIL: source candidate %q does not match identification contract (matches=%t, error=%v)", candidate.ID, matches, matchErr)
					}
				}
			}
		}
		result, translateErr := runtime.Translate(context.Background(), engine.Request{From: features.LanguageID(c.From), To: features.LanguageID(c.To), Text: analysis.Input})
		logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "result": result, "error": errorText(translateErr)})
		if translateErr != nil {
			if infraError(translateErr) {
				t.Errorf("INFRA: translation %q: %v", c.ID, translateErr)
			} else {
				t.Errorf("FAIL: translation %q returned an error: %v", c.ID, translateErr)
			}
			return
		}
		if result.Status == status.OutcomeUnsupported {
			kind := "GAP_REALIZE"
			if sourceGap {
				kind = "GAP_ANALYZE"
			}
			for _, diagnostic := range result.Diagnostics {
				if strings.Contains(diagnostic.Code, "SOURCE") || strings.Contains(diagnostic.Code, "ANALYSIS") {
					kind = "GAP_ANALYZE"
					break
				}
			}
			t.Errorf("%s: translation returned unsupported: %v", kind, result.Diagnostics)
			return
		}
		if len(result.Candidates) == 0 {
			kind := "GAP_REALIZE"
			if sourceGap {
				kind = "GAP_ANALYZE"
			}
			t.Errorf("%s: translation returned no candidates", kind)
			return
		}
		converted := make([]engine.SurfaceAlternative, 0, len(result.Candidates))
		for _, candidate := range result.Candidates {
			if candidate.Meaning == nil {
				t.Errorf("GAP_ANALYZE: translation candidate %q has no Meaning", candidate.ID)
			} else if matches, compareErr := GraphMatches(want, *candidate.Meaning); compareErr != nil || !matches {
				t.Errorf("FAIL: translation candidate %q changed Meaning (matches=%t, error=%v)", candidate.ID, matches, compareErr)
			}
			converted = append(converted, engine.SurfaceAlternative{
				ID: candidate.ID, Outcome: candidate.Outcome, Text: candidate.Text,
				Assumptions: candidate.Assumptions, Losses: candidate.Losses,
				Diagnostics: candidate.Diagnostics,
			})
		}
		checkSurfaces(t, realization, result.Status, converted)
	})
}

func checkSurfaces(t *testing.T, c RealizationCase, outcome status.OutcomeKind, surfaces []engine.SurfaceAlternative) {
	t.Helper()
	if outcome == status.OutcomeUnsupported || outcome == status.OutcomeUnknown {
		t.Errorf("GAP_REALIZE: target adapter returned status %s", outcome)
		return
	}
	if outcome != status.OutcomeOK && outcome != status.OutcomeAmbiguous && outcome != status.OutcomeIncomplete {
		t.Errorf("FAIL: target adapter returned status %s", outcome)
	}
	allowed := make(map[string]bool, len(c.AllowedTexts))
	for _, text := range c.AllowedTexts {
		allowed[text] = true
	}
	seen := make(map[string]bool, len(surfaces))
	for _, surface := range surfaces {
		if surface.Text == "" {
			t.Errorf("FAIL: target candidate %q contains no text", surface.ID)
		} else if !allowed[surface.Text] {
			t.Errorf("FAIL: unexpected target text %q", surface.Text)
		}
		if surface.Text != "" {
			seen[surface.Text] = true
		}
		for _, failure := range surfaceRequirementFailures(c, surface) {
			t.Errorf("FAIL: target candidate %q %s", surface.ID, failure)
		}
		assumptionFields := make(map[string]bool, len(surface.Assumptions))
		for _, assumption := range surface.Assumptions {
			if !contains(c.AllowedAssumptionFields, assumption.Field) || len(assumption.Path) == 0 || strings.TrimSpace(assumption.Reason) == "" || strings.TrimSpace(assumption.Value) == "" {
				t.Errorf("FAIL: target candidate %q has unlicensed or incomplete assumption: %+v", surface.ID, assumption)
			}
			assumptionFields[assumption.Field] = true
		}
		for _, required := range c.RequiredAssumptionFields {
			if !assumptionFields[required] {
				t.Errorf("FAIL: target candidate %q lacks required assumption field %q", surface.ID, required)
			}
		}
		lossFields := make(map[string]bool, len(surface.Losses))
		for _, loss := range surface.Losses {
			if !contains(c.AllowedLossFields, loss.Field) || loss.Outcome != status.OutcomeIncomplete || len(loss.Path) == 0 || strings.TrimSpace(loss.Reason) == "" {
				t.Errorf("FAIL: target candidate %q has unlicensed or incomplete loss: %+v", surface.ID, loss)
			}
			lossFields[loss.Field] = true
		}
		for _, required := range c.RequiredLossFields {
			if !lossFields[required] {
				t.Errorf("FAIL: target candidate %q lacks required loss field %q", surface.ID, required)
			}
		}
		if outcome == status.OutcomeIncomplete || surface.Outcome == status.OutcomeIncomplete {
			if len(assumptionFields) == 0 && len(lossFields) == 0 {
				t.Errorf("FAIL: incomplete target candidate %q has neither an allowed assumption nor an allowed loss", surface.ID)
			}
		}
		if surface.Outcome != status.OutcomeOK && surface.Outcome != status.OutcomeAmbiguous && surface.Outcome != status.OutcomeIncomplete {
			if surface.Outcome == status.OutcomeUnsupported || surface.Outcome == status.OutcomeUnknown {
				t.Errorf("GAP_REALIZE: target candidate %q has status %s", surface.ID, surface.Outcome)
			} else {
				t.Errorf("FAIL: target candidate %q has status %s", surface.ID, surface.Outcome)
			}
		}
	}
	for _, missing := range missingRequiredTexts(c.RequiredTexts, seen) {
		t.Errorf("FAIL: expected target text %q was not produced", missing)
	}
	if len(allowed) == 0 {
		t.Errorf("SPEC_PENDING / NOT RUN: no allowed target texts for %q", c.ID)
	}
}

func missingRequiredTexts(required []string, seen map[string]bool) []string {
	var missing []string
	for _, text := range required {
		if !seen[text] {
			missing = append(missing, text)
		}
	}
	return missing
}

func surfaceRequirementFailures(c RealizationCase, surface engine.SurfaceAlternative) []string {
	var failures []string
	for _, requirement := range c.RequiredTextAssumptions {
		if requirement.Text != surface.Text {
			continue
		}
		matched := false
		for _, assumption := range surface.Assumptions {
			if assumption.Field == requirement.Field && assumption.Value == requirement.Value {
				matched = true
				break
			}
		}
		if !matched {
			failures = append(failures, "lacks the required "+requirement.Field+"="+requirement.Value+" assumption for "+requirement.Text)
		}
	}
	return failures
}

func logObservation(t testing.TB, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Logf("SHA=%s observation JSON encoding failed: %v; value=%+v", observedSHA(), err, value)
		return
	}
	t.Logf("SHA=%s observation=%s", observedSHA(), encoded)
}

func observedSHA() string {
	if value := strings.TrimSpace(os.Getenv("GOLINGUA_MATRIX_SHA")); value != "" {
		return value
	}
	return "UNRECORDED"
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func infraError(err error) bool {
	if err == nil {
		return false
	}
	var resultError interface{ Result() status.Result }
	if errors.As(err, &resultError) {
		for _, diagnostic := range resultError.Result().Diagnostics {
			if strings.Contains(diagnostic.Code, "PIPELINE_UNAVAILABLE") || strings.Contains(diagnostic.Code, "ADAPTER_UNAVAILABLE") || strings.Contains(diagnostic.Code, "RUNTIME_UNAVAILABLE") {
				return true
			}
		}
	}
	return strings.Contains(err.Error(), "unavailable")
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func loadFixture(id string) (meaning.Meaning, *interlingua.IdentificationMeaning, bool, error) {
	if graph, identification, ok, err := interlingua.Negation(id); err != nil || ok {
		return graph, identification, ok, err
	}
	if graph, ok, err := interlingua.Seeing(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, ok, err := interlingua.Reading(id); err != nil || ok {
		return graph, nil, ok, err
	}
	fixture, ok, err := interlingua.Identification(id)
	if err != nil || !ok {
		return meaning.Meaning{}, nil, ok, err
	}
	return fixture.Graph, &fixture, true, nil
}

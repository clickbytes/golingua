package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
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
			if gap := c.SemanticRequirement.RepresentationGap; gap != "" {
				defer t.Errorf("GAP_I: %s", gap)
			}
			want, identification, ok, err := loadFixture(c.MeaningID)
			if err != nil {
				t.Fatalf("INFRA: build fixture %q: %v", c.MeaningID, err)
			}
			if !ok {
				if gap := c.SemanticRequirement.RepresentationGap; gap != "" {
					runtime := Runtime(t)
					analyzer, exists := runtime.SourceAnalyzerFor(lang)
					if !exists || analyzer == nil {
						t.Fatalf("INFRA: source analyzer for %q is unavailable", lang)
					}
					result, analyzeErr := analyzer.Analyze(context.Background(), c.Input)
					logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "result": result, "error": errorText(analyzeErr)})
					if analyzeErr != nil && infraError(analyzeErr) {
						t.Errorf("INFRA: source analysis %q: %v", c.ID, analyzeErr)
					} else if analyzeErr != nil || result.Outcome == status.OutcomeUnsupported || len(result.Candidates) == 0 {
						t.Errorf("GAP_ANALYZE: source adapter did not produce a supported candidate for the explicit representation gap")
					}
					return
				}
				t.Errorf("SPEC_PENDING / NOT RUN: no independent Meaning fixture for %q", c.MeaningID)
				return
			}
			runtime := Runtime(t)
			if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("INFRA: fixture %q does not validate against CORE: %v", c.MeaningID, err)
			}
			allowedIDs, requiredIDs, contractErr := caseMeaningIDs(c.MeaningID, c.AllowedMeaningIDs, c.RequiredMeaningIDs)
			if contractErr != nil {
				t.Fatalf("SPEC_PENDING: analysis meaning contract is invalid: %v", contractErr)
			}
			fixtures, fixtureErr := loadMeaningFixtures(allowedIDs, runtime)
			if fixtureErr != nil {
				t.Fatalf("SPEC_PENDING / NOT RUN: load allowed analysis meanings: %v", fixtureErr)
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
				for _, meaningID := range MissingRequiredMeaningIDs(requiredIDs, nil) {
					t.Errorf("FAIL: required Meaning %q was not produced", meaningID)
				}
				return
			}
			observedRequired := make(map[string]bool, len(requiredIDs))
			for _, candidate := range result.Candidates {
				if candidate.Status == status.OutcomeUnsupported || candidate.Meaning == nil {
					t.Errorf("GAP_ANALYZE: candidate %q has no supported Meaning (status %s)", candidate.ID, candidate.Status)
					continue
				}
				if candidate.Status != status.OutcomeOK {
					t.Errorf("FAIL: candidate %q has status %s", candidate.ID, candidate.Status)
				}
				if identification != nil {
					matches, compareErr := IdentificationMatches(*identification, candidate)
					if compareErr != nil || !matches {
						t.Errorf("FAIL: candidate %q does not match identification contract (matches=%t, error=%v)", candidate.ID, matches, compareErr)
					} else {
						observedRequired[c.MeaningID] = true
					}
				} else {
					meaningID, matches, compareErr := MatchAllowedMeaning(allowedIDs, fixtures, *candidate.Meaning)
					if compareErr != nil || !matches {
						t.Errorf("FAIL: candidate %q Meaning does not match an allowed fixture (matches=%t, error=%v)", candidate.ID, matches, compareErr)
					} else {
						observedRequired[meaningID] = true
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
			for _, meaningID := range MissingRequiredMeaningIDs(requiredIDs, observedRequired) {
				t.Errorf("FAIL: required Meaning %q was not produced", meaningID)
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
			if gap := c.SemanticRequirement.RepresentationGap; gap != "" {
				defer t.Errorf("GAP_I: %s", gap)
			}
			contracts, err := realizationMeaningContracts(c)
			if err != nil {
				t.Fatalf("SPEC_PENDING: realization meaning contract is invalid: %v", err)
			}
			for _, contract := range contracts {
				variant := realizationVariantCase(c, contract)
				want, identification, ok, err := loadFixture(contract.MeaningID)
				if err != nil {
					t.Errorf("INFRA: build fixture %q: %v", contract.MeaningID, err)
					continue
				}
				if !ok {
					if gap := c.SemanticRequirement.RepresentationGap; gap != "" {
						logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "meaning_id": contract.MeaningID, "independent_realization_executed": false, "reason": gap})
						continue
					}
					t.Errorf("SPEC_PENDING / NOT RUN: no independent Meaning fixture for %q", contract.MeaningID)
					continue
				}
				runtime := Runtime(t)
				if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
					t.Errorf("INFRA: fixture %q does not validate against CORE: %v", contract.MeaningID, err)
					continue
				}
				realizer, exists := runtime.TargetRealizerFor(lang)
				if !exists || realizer == nil {
					t.Fatalf("INFRA: target realizer for %q is unavailable", lang)
				}
				candidate := semantic.MeaningCandidate{
					ID: contract.MeaningID, SemanticID: contract.MeaningID, Meaning: &want,
					Status:   status.OutcomeOK,
					Evidence: semantic.EvidenceTrace{AlternativeID: contract.MeaningID, Outcome: status.OutcomeOK},
				}
				if identification != nil {
					candidate.Contexts = []semantic.MeaningContext{{Meaning: want, Context: identification.Context}}
				}
				result, realizeErr := realizer.Realize(context.Background(), candidate)
				logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "meaning_id": contract.MeaningID, "result": result, "error": errorText(realizeErr)})
				if realizeErr != nil {
					if infraError(realizeErr) {
						t.Errorf("INFRA: target realization %q: %v", c.ID, realizeErr)
					} else {
						t.Errorf("GAP_REALIZE: target realization %q failed: %v", c.ID, realizeErr)
					}
					continue
				}
				checkSurfaces(t, variant, result.Outcome, result.Alternatives)
			}
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
		if analysis.MeaningID == "" || realization.MeaningID != analysis.MeaningID {
			t.Fatalf("SPEC_PENDING: translation analysis and realization must share a nonempty MeaningID")
		}
		if !reflect.DeepEqual(analysis.SemanticRequirement, realization.SemanticRequirement) {
			t.Fatalf("SPEC_PENDING: translation analysis and realization semantic requirements differ")
		}
		allowedIDs, requiredIDs, contractErr := caseMeaningIDs(analysis.MeaningID, analysis.AllowedMeaningIDs, analysis.RequiredMeaningIDs)
		if contractErr != nil {
			t.Fatalf("SPEC_PENDING: analysis meaning contract is invalid: %v", contractErr)
		}
		realizationContracts, contractErr := realizationMeaningContracts(realization)
		if contractErr != nil {
			t.Fatalf("SPEC_PENDING: realization meaning contract is invalid: %v", contractErr)
		}
		contractsByMeaning := make(map[string]RealizationCase, len(realizationContracts))
		for _, contract := range realizationContracts {
			contractsByMeaning[contract.MeaningID] = realizationVariantCase(realization, contract)
		}
		allowedSet := make(map[string]bool, len(allowedIDs))
		for _, meaningID := range allowedIDs {
			allowedSet[meaningID] = true
			if _, exists := contractsByMeaning[meaningID]; !exists {
				t.Fatalf("SPEC_PENDING: analysis allows Meaning %q without a target realization contract", meaningID)
			}
		}
		for meaningID := range contractsByMeaning {
			if !allowedSet[meaningID] {
				t.Fatalf("SPEC_PENDING: target realization declares Meaning %q outside the source allowed set", meaningID)
			}
		}
		gap := analysis.SemanticRequirement.RepresentationGap
		if gap != "" {
			defer t.Errorf("GAP_I: %s", gap)
		}
		want, identification, ok, err := loadFixture(analysis.MeaningID)
		if err != nil {
			t.Fatalf("INFRA: build fixture %q: %v", analysis.MeaningID, err)
		}
		if !ok {
			if gap != "" {
				runtime := Runtime(t)
				result, translateErr := runtime.Translate(context.Background(), engine.Request{From: features.LanguageID(c.From), To: features.LanguageID(c.To), Text: analysis.Input})
				logObservation(t, map[string]any{"sha": observedSHA(), "case_id": c.ID, "result": result, "error": errorText(translateErr)})
				if translateErr != nil && infraError(translateErr) {
					t.Errorf("INFRA: translation %q: %v", c.ID, translateErr)
				} else if translateErr != nil || result.Status == status.OutcomeUnsupported || len(result.Candidates) == 0 {
					t.Errorf("GAP_ANALYZE: translation did not produce a candidate under the explicit representation gap")
				}
				return
			}
			t.Errorf("SPEC_PENDING / NOT RUN: translation lacks a shared independent Meaning fixture")
			return
		}
		runtime := Runtime(t)
		if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			t.Fatalf("INFRA: fixture %q does not validate against CORE: %v", analysis.MeaningID, err)
		}
		fixtures, fixtureErr := loadMeaningFixtures(allowedIDs, runtime)
		if fixtureErr != nil {
			t.Fatalf("SPEC_PENDING / NOT RUN: load allowed translation meanings: %v", fixtureErr)
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
			for _, meaningID := range MissingRequiredMeaningIDs(requiredIDs, nil) {
				t.Errorf("FAIL: translation omitted required Meaning %q", meaningID)
			}
			return
		}
		if len(result.Candidates) == 0 {
			kind := "GAP_REALIZE"
			if sourceGap {
				kind = "GAP_ANALYZE"
			}
			t.Errorf("%s: translation returned no candidates", kind)
			for _, meaningID := range MissingRequiredMeaningIDs(requiredIDs, nil) {
				t.Errorf("FAIL: translation omitted required Meaning %q", meaningID)
			}
			return
		}
		convertedByMeaning := make(map[string][]engine.SurfaceAlternative, len(allowedIDs))
		observedMeaning := make(map[string]bool, len(allowedIDs))
		for _, candidate := range result.Candidates {
			if candidate.Meaning == nil {
				t.Errorf("GAP_ANALYZE: translation candidate %q has no Meaning", candidate.ID)
				continue
			}
			meaningID, matches, compareErr := MatchAllowedMeaning(allowedIDs, fixtures, *candidate.Meaning)
			if compareErr != nil || !matches {
				t.Errorf("FAIL: translation candidate %q Meaning is not allowed (matches=%t, error=%v)", candidate.ID, matches, compareErr)
				continue
			}
			observedMeaning[meaningID] = true
			convertedByMeaning[meaningID] = append(convertedByMeaning[meaningID], engine.SurfaceAlternative{
				ID: candidate.ID, Outcome: candidate.Outcome, Text: candidate.Text,
				Assumptions: candidate.Assumptions, Losses: candidate.Losses,
				Diagnostics: candidate.Diagnostics,
			})
		}
		for _, meaningID := range MissingRequiredMeaningIDs(requiredIDs, observedMeaning) {
			t.Errorf("FAIL: translation omitted required Meaning %q", meaningID)
		}
		for meaningID, surfaces := range convertedByMeaning {
			checkSurfaces(t, contractsByMeaning[meaningID], result.Status, surfaces)
		}
	})
}

func caseMeaningIDs(primary string, allowed, required []string) ([]string, []string, error) {
	if strings.TrimSpace(primary) == "" {
		return nil, nil, errors.New("primary MeaningID is empty")
	}
	allowedIDs := []string{primary}
	allowedSet := map[string]bool{primary: true}
	for _, id := range allowed {
		if strings.TrimSpace(id) == "" || allowedSet[id] {
			return nil, nil, fmt.Errorf("allowed MeaningID is empty or duplicated: %q", id)
		}
		allowedSet[id] = true
		allowedIDs = append(allowedIDs, id)
	}
	requiredIDs := []string{primary}
	requiredSet := map[string]bool{primary: true}
	for _, id := range required {
		if strings.TrimSpace(id) == "" || requiredSet[id] {
			return nil, nil, fmt.Errorf("required MeaningID is empty or duplicated: %q", id)
		}
		if !allowedSet[id] {
			return nil, nil, fmt.Errorf("required MeaningID %q is not allowed", id)
		}
		requiredSet[id] = true
		requiredIDs = append(requiredIDs, id)
	}
	return allowedIDs, requiredIDs, nil
}

func realizationMeaningContracts(c RealizationCase) ([]MeaningRealization, error) {
	if strings.TrimSpace(c.MeaningID) == "" {
		return nil, errors.New("primary MeaningID is empty")
	}
	contracts := []MeaningRealization{{MeaningID: c.MeaningID, AllowedTexts: c.AllowedTexts, RequiredTexts: c.RequiredTexts, RequiredTextAssumptions: c.RequiredTextAssumptions, RequiredTextGroups: c.RequiredTextGroups}}
	seen := map[string]bool{c.MeaningID: true}
	for _, variant := range c.MeaningVariants {
		if strings.TrimSpace(variant.MeaningID) == "" || seen[variant.MeaningID] {
			return nil, fmt.Errorf("variant MeaningID is empty or duplicated: %q", variant.MeaningID)
		}
		if len(variant.AllowedTexts) == 0 {
			return nil, fmt.Errorf("variant MeaningID %q has no allowed texts", variant.MeaningID)
		}
		seen[variant.MeaningID] = true
		contracts = append(contracts, variant)
	}
	return contracts, nil
}

func realizationVariantCase(parent RealizationCase, contract MeaningRealization) RealizationCase {
	variant := parent
	variant.MeaningID = contract.MeaningID
	variant.AllowedTexts = contract.AllowedTexts
	variant.RequiredTexts = contract.RequiredTexts
	variant.RequiredTextGroups = contract.RequiredTextGroups
	if contract.RequiredTextAssumptions != nil {
		variant.RequiredTextAssumptions = contract.RequiredTextAssumptions
	}
	return variant
}

func loadMeaningFixtures(ids []string, runtime *engine.Engine) (map[string]meaning.Meaning, error) {
	fixtures := make(map[string]meaning.Meaning, len(ids))
	for _, id := range ids {
		graph, _, ok, err := loadFixture(id)
		if err != nil {
			return nil, fmt.Errorf("build fixture %q: %w", id, err)
		}
		if !ok {
			return nil, fmt.Errorf("no authored fixture for %q", id)
		}
		if err := meaning.ValidateMeaning(graph, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			return nil, fmt.Errorf("fixture %q does not validate against CORE: %w", id, err)
		}
		fixtures[id] = graph
	}
	return fixtures, nil
}

// MatchAllowedMeaning returns the unique authored MeaningID matched by a candidate.
func MatchAllowedMeaning(ids []string, fixtures map[string]meaning.Meaning, candidate meaning.Meaning) (string, bool, error) {
	matchedID := ""
	for _, id := range ids {
		fixture, ok := fixtures[id]
		if !ok {
			return "", false, fmt.Errorf("missing allowed fixture %q", id)
		}
		matches, err := GraphMatches(fixture, candidate)
		if err != nil {
			return "", false, err
		}
		if matches {
			if matchedID != "" {
				return "", false, fmt.Errorf("candidate matches multiple allowed meanings: %q and %q", matchedID, id)
			}
			matchedID = id
		}
	}
	return matchedID, matchedID != "", nil
}

// MissingRequiredMeaningIDs reports required readings not observed in a result.
func MissingRequiredMeaningIDs(required []string, observed map[string]bool) []string {
	var missing []string
	for _, id := range required {
		if !observed[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

func checkSurfaces(t *testing.T, c RealizationCase, outcome status.OutcomeKind, surfaces []engine.SurfaceAlternative) {
	t.Helper()
	if outcome == status.OutcomeUnsupported || outcome == status.OutcomeUnknown {
		t.Errorf("GAP_REALIZE: target adapter returned status %s", outcome)
		return
	}
	if len(surfaces) == 0 {
		t.Errorf("FAIL: target adapter returned no surface alternatives")
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
	if invalid := invalidRequiredTextGroups(c.RequiredTextGroups, allowed); len(invalid) > 0 {
		t.Errorf("SPEC_PENDING: required text groups contain empty or unlicensed alternatives: %v", invalid)
	}
	for _, group := range missingRequiredTextGroups(c.RequiredTextGroups, seen) {
		t.Errorf("FAIL: required text group %d produced none of its alternatives", group)
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
	if graph, ok, err := interlingua.CompositionBatch(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, ok, err := interlingua.Composition(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, identification, ok, err := interlingua.Negation(id); err != nil || ok {
		return graph, identification, ok, err
	}
	if graph, ok, err := interlingua.Seeing(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, ok, err := interlingua.CorpusEvent(strings.TrimPrefix(id, "corpus.")); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, ok, err := interlingua.Contrast(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, ok, err := interlingua.Reading(id); err != nil || ok {
		return graph, nil, ok, err
	}
	if graph, identification, ok, err := interlingua.PendingDomain(id); err != nil || ok {
		return graph, identification, ok, err
	}
	fixture, ok, err := interlingua.Identification(id)
	if err != nil || !ok {
		return meaning.Meaning{}, nil, ok, err
	}
	return fixture.Graph, &fixture, true, nil
}

// invalidRequiredTextGroups identifies empty or unlicensed style groups.
func invalidRequiredTextGroups(groups [][]string, allowed map[string]bool) []int {
	var invalid []int
	for i, group := range groups {
		valid := len(group) > 0
		for _, text := range group {
			if text == "" || !allowed[text] {
				valid = false
			}
		}
		if !valid {
			invalid = append(invalid, i)
		}
	}
	return invalid
}

// missingRequiredTextGroups returns required groups without any observed text.
func missingRequiredTextGroups(groups [][]string, seen map[string]bool) []int {
	var missing []int
	for i, group := range groups {
		produced := false
		for _, text := range group {
			produced = produced || seen[text]
		}
		if !produced {
			missing = append(missing, i)
		}
	}
	return missing
}

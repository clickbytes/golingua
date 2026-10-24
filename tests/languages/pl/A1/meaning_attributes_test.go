package a1pl_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestMeaningAttributesAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), []matrix.AnalysisCase{
		{ID: "H-PL-reading-teacher-present", MeaningID: "reading.teacher-message.present", Input: "Nauczyciel czyta wiadomość."},
		{ID: "H-PL-reading-teacher-past", MeaningID: "reading.teacher-message.past", Input: "Nauczyciel czytał wiadomość."},
		{ID: "H-PL-reading-teachers-present", MeaningID: "reading.teachers-message.present", Input: "Nauczyciele czytają wiadomość."},
		{ID: "H-PL-reading-teachers-past", MeaningID: "reading.teachers-message.past", Input: "Nauczyciele czytali wiadomość."},
	})
}

func TestMeaningAttributesRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), []matrix.RealizationCase{
		{ID: "H-I-PL-reading-teacher-present", MeaningID: "reading.teacher-message.present", AllowedTexts: []string{"Nauczyciel czyta wiadomość."}, AllowedAssumptionFields: []string{"person"}},
		{ID: "H-I-PL-reading-teacher-past", MeaningID: "reading.teacher-message.past", AllowedTexts: []string{"Nauczyciel czytał wiadomość."}, AllowedAssumptionFields: []string{"person"}},
		{ID: "H-I-PL-reading-teachers-present", MeaningID: "reading.teachers-message.present", AllowedTexts: []string{"Nauczyciele czytają wiadomość."}, AllowedAssumptionFields: []string{"person"}},
		{ID: "H-I-PL-reading-teachers-past", MeaningID: "reading.teachers-message.past", AllowedTexts: []string{"Nauczyciele czytali wiadomość."}, AllowedAssumptionFields: []string{"person"}},
	})
}

func TestMeaningAttributesSeeingMultiplicity(t *testing.T) {
	base, ok, e := interlingua.Seeing("seeing.teacher-dog")
	if e != nil || !ok {
		t.Fatalf("Seeing fixture: %v", e)
	}
	for _, tc := range []struct {
		name string
		node meaning.NodeID
		text string
	}{
		{"plural teacher", "t", "Nauczyciele widzą psa."}, {"plural dog", "d", "Nauczyciel widzi psy."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := setMultiplicityPL(t, base, tc.node, meaning.MultiplicityMany)
			r := realizePL(t, g, "seeing.teacher-dog-"+tc.name)
			if !sameStrings(textSet(r.Alternatives), []string{tc.text}) {
				t.Errorf("FAIL: got %v, want %q (%s)", textSet(r.Alternatives), tc.text, r.Outcome)
			}
			assertNoAttributeAssumptions(t, r.Alternatives)
		})
	}
	negative, _, ok, e := interlingua.Negation("negation.seeing.teacher-dog")
	if e != nil || !ok {
		t.Fatalf("negative fixture: %v", e)
	}
	g := setMultiplicityPL(t, negative, "t", meaning.MultiplicityMany)
	r := realizePL(t, g, "negation.seeing.teacher-dog-many")
	if !sameStrings(textSet(r.Alternatives), []string{"Nauczyciele nie widzą psa."}) {
		t.Errorf("FAIL: negative plural target got %v (%s)", textSet(r.Alternatives), r.Outcome)
	}
	assertNoAttributeAssumptions(t, r.Alternatives)
}

func TestMeaningAttributesMassAnalysis(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("pl")
	if !ok {
		t.Fatal("INFRA: Polish analyzer unavailable")
	}
	result, err := analyzer.Analyze(context.Background(), "Nauczyciel dostał pieniądze.")
	logPL(t, "mass money", result, err)
	if err != nil || len(result.Candidates) == 0 {
		t.Fatalf("GAP_ANALYZE: money sentence has no candidates: %s %v", result.Outcome, err)
	}
	valid := false
	for _, c := range result.Candidates {
		if c.Meaning == nil {
			continue
		}
		event, ok := predicateNode(*c.Meaning)
		if !ok || event.TemporalFrame != meaning.TemporalFramePast {
			continue
		}
		teacher, teacherOK := referentConceptPL(*c.Meaning, "concept:teacher")
		money, moneyOK := referentConceptPL(*c.Meaning, "concept:money")
		if teacherOK && moneyOK && teacher.Multiplicity == meaning.MultiplicityOne && money.Multiplicity == "" {
			valid = true
		}
	}
	if !valid {
		t.Errorf("FAIL: no past receiving candidate with teacher=one and money multiplicity unknown")
	}
}

func TestMeaningAttributesRefuseUnsupportedAndConflicting(t *testing.T) {
	fixture, ok, e := interlingua.Identification("identification.classification.to")
	if e != nil || !ok {
		t.Fatalf("fixture: %v", e)
	}
	positiveBaseline := meaningCandidatePL(fixture.Graph, "classification-positive-baseline", fixture.Context)
	validateCandidatePL(t, positiveBaseline)
	positiveResult, positiveErr := targetPL(t).Realize(context.Background(), positiveBaseline)
	logPL(t, "classification valid positive baseline", positiveResult, positiveErr)
	requireSurfacePL(t, positiveResult, positiveErr)
	for _, tc := range []struct {
		name string
		edit func(*meaning.Meaning)
	}{
		{"past", func(g *meaning.Meaning) { setFrameInPlacePL(g, meaning.TemporalFramePast) }},
		{"many", func(g *meaning.Meaning) { setMultiplicityInPlacePL(g, "r", meaning.MultiplicityMany) }},
	} {
		t.Run("classification "+tc.name, func(t *testing.T) {
			g := cloneMeaningPL(t, fixture.Graph)
			tc.edit(&g)
			r := realizePLWithContext(t, g, fixture.Context, "classification-"+tc.name)
			assertNoTextPL(t, r, nil)
		})
	}
	negative, negativeContext, ok, e := interlingua.Negation("negation.classification.to")
	if e != nil || !ok || negativeContext == nil {
		t.Fatalf("negation fixture: %v", e)
	}
	contextGraph := setMultiplicityPL(t, negative, "r", meaning.MultiplicityMany)
	candidate := meaningCandidatePL(negative, "negative-context-many", negativeContext.Context)
	candidate.Contexts[0].Meaning = contextGraph
	baseCandidate := meaningCandidatePL(negative, "negative-context-baseline", negativeContext.Context)
	validateCandidatePL(t, baseCandidate)
	baseResult, baseErr := targetPL(t).Realize(context.Background(), baseCandidate)
	logPL(t, "negative classification valid baseline", baseResult, baseErr)
	requireSurfacePL(t, baseResult, baseErr)
	validateCandidatePL(t, candidate)
	r, err := targetPL(t).Realize(context.Background(), candidate)
	logPL(t, "negative classification context multiplicity mismatch", r, err)
	assertNoTextPL(t, r, err)
	// A valid source graph paired with a context graph that claims a different frame is refused.
	g := fixture.Graph
	contextGraph = setPredicateFramePL(t, g, meaning.TemporalFramePast)
	good := meaningCandidatePL(g, "mismatch-context", fixture.Context)
	good.Contexts[0].Meaning = contextGraph
	contextBaseline := meaningCandidatePL(g, "context-valid-baseline", fixture.Context)
	validateCandidatePL(t, contextBaseline)
	baselineResult, baselineErr := targetPL(t).Realize(context.Background(), contextBaseline)
	logPL(t, "context mismatch positive baseline", baselineResult, baselineErr)
	requireSurfacePL(t, baselineResult, baselineErr)
	validateCandidatePL(t, good)
	r, err = targetPL(t).Realize(context.Background(), good)
	logPL(t, "context polarity mismatch", r, err)
	assertNoTextPL(t, r, err)
}

func TestMeaningAttributesEvidenceConflictAndUnknownFallback(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("pl")
	if !ok {
		t.Fatal("INFRA: analyzer unavailable")
	}
	target := targetPL(t)
	for _, tc := range []struct {
		name, input string
		mutate      func(*meaning.Meaning)
	}{
		{"time conflict", "Nauczyciel czyta wiadomość.", func(g *meaning.Meaning) { setFrameInPlacePL(g, meaning.TemporalFramePast) }},
		{"number conflict", "Nauczyciel czyta wiadomość.", func(g *meaning.Meaning) { setMultiplicityConceptPL(g, "concept:teacher", meaning.MultiplicityMany) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, e := analyzer.Analyze(context.Background(), tc.input)
			logPL(t, tc.input, result, e)
			if e != nil || len(result.Candidates) == 0 {
				t.Fatalf("analysis unavailable: %s %v", result.Outcome, e)
			}
			candidate := result.Candidates[0]
			candidate.Evidence = candidate.Evidence.Clone()
			candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
			if candidate.Meaning == nil {
				t.Fatal("source candidate lacks Meaning")
			}
			assertSourceReadingPresentOnePL(t, candidate)
			assertReadingEvidencePL(t, candidate.Evidence)
			altered := cloneMeaningPL(t, *candidate.Meaning)
			tc.mutate(&altered)
			candidate.Meaning = &altered
			syncContextsPL(t, &candidate, altered)
			validateCandidatePL(t, candidate)
			validateCandidatePL(t, result.Candidates[0])
			baselineResult, baselineErr := target.Realize(context.Background(), result.Candidates[0])
			logPL(t, "valid source candidate baseline", baselineResult, baselineErr)
			requireSurfacePL(t, baselineResult, baselineErr)
			surf, realizeErr := target.Realize(context.Background(), candidate)
			logPL(t, "contradictory I and source evidence", surf, realizeErr)
			assertNoTextPL(t, surf, realizeErr)
		})
	}
	// A present third-person verb form without a separate tense value still conflicts with past I.
	result, e := analyzer.Analyze(context.Background(), "Nauczyciel czyta wiadomość.")
	if e != nil || len(result.Candidates) == 0 {
		t.Fatalf("present analysis unavailable: %s %v", result.Outcome, e)
	}
	candidate := result.Candidates[0]
	candidate.Evidence = candidate.Evidence.Clone()
	candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
	if candidate.Meaning == nil {
		t.Fatal("present candidate lacks Meaning")
	}
	assertSourceReadingPresentOnePL(t, candidate)
	assertReadingEvidencePL(t, candidate.Evidence)
	altered := cloneMeaningPL(t, *candidate.Meaning)
	setFrameInPlacePL(&altered, meaning.TemporalFramePast)
	candidate.Meaning = &altered
	syncContextsPL(t, &candidate, altered)
	if !setContradictoryPresentFormPL(&candidate.Evidence) {
		t.Fatal("review trap failed: no exact reading verb-form evidence changed")
	}
	validateCandidatePL(t, candidate)
	surf, realizeErr := target.Realize(context.Background(), candidate)
	logPL(t, "past I versus present form without tense", surf, realizeErr)
	assertNoTextPL(t, surf, realizeErr)
	// Unknown I attributes may use explicit defaults, but evidence must not silently restore them.
	result, e = analyzer.Analyze(context.Background(), "Nauczyciele czytali wiadomość.")
	if e != nil || len(result.Candidates) == 0 {
		t.Fatalf("past analysis: %s %v", result.Outcome, e)
	}
	candidate = result.Candidates[0]
	candidate.Evidence = candidate.Evidence.Clone()
	candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
	if candidate.Meaning == nil {
		t.Fatal("past candidate has no Meaning")
	}
	pastNode, _ := predicateNode(*candidate.Meaning)
	pastReader, _ := referentConceptPL(*candidate.Meaning, "concept:teacher")
	if pastNode.TemporalFrame != meaning.TemporalFramePast || pastReader.Multiplicity != meaning.MultiplicityMany {
		t.Fatal("unknown-attribute control requires known past and many before deletion")
	}
	unknown := cloneMeaningPL(t, *candidate.Meaning)
	setFrameInPlacePL(&unknown, "")
	setMultiplicityConceptPL(&unknown, "concept:teacher", "")
	candidate.Meaning = &unknown
	syncContextsPL(t, &candidate, unknown)
	validateCandidatePL(t, candidate)
	surf, realizeErr = target.Realize(context.Background(), candidate)
	logPL(t, "unknown attributes with source evidence", surf, realizeErr)
	if realizeErr != nil || len(textSet(surf.Alternatives)) == 0 {
		assertNoTextPL(t, surf, realizeErr)
	} else {
		if !sameStrings(textSet(surf.Alternatives), []string{"Nauczyciel czyta wiadomość."}) {
			t.Errorf("unknown attributes used past/plural evidence instead of declared defaults: %+v", surf.Alternatives)
		}
		for _, alt := range surf.Alternatives {
			if alt.Text != "" && (!hasAssumption(alt.Assumptions, "tense") || !hasAssumption(alt.Assumptions, "number")) {
				t.Errorf("FAIL: hidden evidence supplied unknown I fields without both explicit assumptions: %+v", alt)
			}
		}
	}
}

func TestMeaningAttributesMassKnownMultiplicityRefusal(t *testing.T) {
	// The mass noun route is not authorized to reinterpret known multiplicity as a count.
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("pl")
	if !ok {
		t.Fatal("INFRA: analyzer unavailable")
	}
	res, err := analyzer.Analyze(context.Background(), "Nauczyciel dostał pieniądze.")
	if err != nil || len(res.Candidates) == 0 {
		t.Fatalf("mass source unavailable: %s %v", res.Outcome, err)
	}
	for _, candidate := range res.Candidates {
		if candidate.Meaning == nil {
			continue
		}
		candidate.Evidence = candidate.Evidence.Clone()
		candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
		g := cloneMeaningPL(t, *candidate.Meaning)
		setMultiplicityConceptPL(&g, "concept:money", meaning.MultiplicityMany)
		candidate.Meaning = &g
		syncContextsPL(t, &candidate, g)
		validateCandidatePL(t, candidate)
		surf, e := targetPL(t).Realize(context.Background(), candidate)
		logPL(t, "mass multiplicity refusal", surf, e)
		assertNoTextPL(t, surf, e)
		return
	}
	t.Fatal("source money candidates had no Meaning")
}

func realizePL(t *testing.T, g meaning.Meaning, id string) engine.SurfaceResult {
	t.Helper()
	return realizePLWithContext(t, g, meaning.RequestContext{}, id)
}
func realizePLWithContext(t *testing.T, g meaning.Meaning, c meaning.RequestContext, id string) engine.SurfaceResult {
	t.Helper()
	candidate := meaningCandidatePL(g, id, c)
	validateCandidatePL(t, candidate)
	r, e := targetPL(t).Realize(context.Background(), candidate)
	logPL(t, id, r, e)
	if e != nil {
		t.Fatalf("valid prepared Meaning realization failed: %v", e)
	}
	return r
}
func targetPL(t *testing.T) engine.TargetRealizer {
	t.Helper()
	r, ok := matrix.Runtime(t).TargetRealizerFor("pl")
	if !ok {
		t.Fatal("INFRA: Polish target unavailable")
	}
	return r
}
func meaningCandidatePL(g meaning.Meaning, id string, c meaning.RequestContext) semantic.MeaningCandidate {
	candidate := semantic.MeaningCandidate{ID: id, SemanticID: id, Meaning: &g, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: id, Outcome: status.OutcomeOK}}
	if c.ID != "" {
		candidate.Contexts = []semantic.MeaningContext{{Meaning: g, Context: c}}
	}
	return candidate
}
func cloneMeaningPL(t *testing.T, g meaning.Meaning) meaning.Meaning {
	t.Helper()
	out, e := meaning.Build(g.Roots, g.Nodes, meaning.DefaultBudget())
	if e != nil {
		t.Fatalf("clone graph: %v", e)
	}
	return out
}
func setMultiplicityPL(t *testing.T, g meaning.Meaning, id meaning.NodeID, v meaning.Multiplicity) meaning.Meaning {
	t.Helper()
	g = cloneMeaningPL(t, g)
	setMultiplicityInPlacePL(&g, id, v)
	return g
}
func setMultiplicityInPlacePL(g *meaning.Meaning, id meaning.NodeID, v meaning.Multiplicity) {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			g.Nodes[i].Multiplicity = v
		}
	}
}
func setPredicateFramePL(t *testing.T, g meaning.Meaning, v meaning.TemporalFrame) meaning.Meaning {
	t.Helper()
	g = cloneMeaningPL(t, g)
	setFrameInPlacePL(&g, v)
	return g
}
func setFrameInPlacePL(g *meaning.Meaning, v meaning.TemporalFrame) {
	for i := range g.Nodes {
		if g.Nodes[i].Kind == meaning.KindEvent || g.Nodes[i].Kind == meaning.KindState {
			g.Nodes[i].TemporalFrame = v
		}
	}
}
func syncContextsPL(t *testing.T, c *semantic.MeaningCandidate, g meaning.Meaning) {
	for i := range c.Contexts {
		c.Contexts[i].Meaning = cloneMeaningPL(t, g)
	}
}
func predicateNode(g meaning.Meaning) (meaning.Node, bool) {
	for _, n := range g.Nodes {
		if n.Kind == meaning.KindEvent || n.Kind == meaning.KindState {
			return n, true
		}
	}
	return meaning.Node{}, false
}
func referentConceptPL(g meaning.Meaning, concept string) (meaning.Node, bool) {
	for _, n := range g.Nodes {
		if n.Kind == meaning.KindReferent && string(n.Concept) == concept {
			return n, true
		}
	}
	return meaning.Node{}, false
}
func setMultiplicityConceptPL(g *meaning.Meaning, concept string, value meaning.Multiplicity) {
	for i := range g.Nodes {
		if g.Nodes[i].Kind == meaning.KindReferent && string(g.Nodes[i].Concept) == concept {
			g.Nodes[i].Multiplicity = value
		}
	}
}
func textSet(xs []engine.SurfaceAlternative) []string {
	m := map[string]bool{}
	for _, x := range xs {
		if x.Text != "" {
			m[x.Text] = true
		}
	}
	r := []string{}
	for x := range m {
		r = append(r, x)
	}
	sort.Strings(r)
	return r
}
func sameStrings(a, b []string) bool {
	a = append([]string(nil), a...)
	b = append([]string(nil), b...)
	sort.Strings(a)
	sort.Strings(b)
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
func assertNoAttributeAssumptions(t *testing.T, xs []engine.SurfaceAlternative) {
	t.Helper()
	for _, x := range xs {
		for _, a := range x.Assumptions {
			if a.Field != "person" {
				t.Errorf("unexpected assumption for known I realization: %+v", a)
			}
		}
	}
}
func assertNoTextPL(t *testing.T, r engine.SurfaceResult, e error) {
	t.Helper()
	if e != nil {
		if isSetupErrorPL(e) {
			t.Errorf("INFRA: target refusal was a setup/runtime error: %v", e)
		}
		return
	}
	if r.Outcome == status.OutcomeOK || r.Outcome == status.OutcomeAmbiguous || r.Outcome == status.OutcomeIncomplete || len(r.Diagnostics) == 0 {
		t.Errorf("FAIL: refusal requires an explicit non-success outcome and diagnostic; outcome=%s diagnostics=%+v", r.Outcome, r.Diagnostics)
	}
	for _, a := range r.Alternatives {
		if a.Text != "" {
			t.Errorf("FAIL: invalid/conflicting Meaning produced text %q", a.Text)
		}
	}
}
func validateCandidatePL(t *testing.T, c semantic.MeaningCandidate) {
	t.Helper()
	runtime := matrix.Runtime(t)
	if c.Meaning == nil {
		t.Fatal("candidate missing Meaning")
	}
	if err := meaning.ValidateMeaning(*c.Meaning, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatalf("test setup Meaning invalid: %v", err)
	}
	if err := c.Evidence.Validate(); err != nil {
		t.Fatalf("test setup EvidenceTrace invalid: %v", err)
	}
	for i, ctx := range c.Contexts {
		if err := meaning.ValidateMeaning(ctx.Meaning, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			t.Fatalf("test setup context[%d] Meaning invalid: %v", i, err)
		}
		if err := meaning.ValidateReferences(ctx.Meaning, ctx.Context, meaning.DefaultBudget()); err != nil {
			t.Fatalf("test setup context[%d] references invalid: %v", i, err)
		}
	}
}
func requireSurfacePL(t *testing.T, r engine.SurfaceResult, e error) {
	t.Helper()
	if e != nil {
		t.Fatalf("valid positive baseline failed: %v", e)
	}
	for _, alt := range r.Alternatives {
		if alt.Text != "" {
			return
		}
	}
	t.Fatalf("valid positive baseline returned no surface: %+v", r)
}
func isSetupErrorPL(e error) bool {
	var typed interface{ Result() status.Result }
	if errors.As(e, &typed) {
		for _, d := range typed.Result().Diagnostics {
			if strings.Contains(d.Code, "PIPELINE_UNAVAILABLE") || strings.Contains(d.Code, "ADAPTER_UNAVAILABLE") || strings.Contains(d.Code, "RUNTIME_UNAVAILABLE") {
				return true
			}
		}
	}
	s := strings.ToLower(e.Error())
	return strings.Contains(s, "unavailable") || strings.Contains(s, "runtime") || strings.Contains(s, "adapter") || strings.Contains(s, "registry") || strings.Contains(s, "snapshot") || strings.Contains(s, "compile")
}
func hasAssumption(xs []realizer.FieldAssumption, field string) bool {
	for _, x := range xs {
		if x.Field == field {
			return true
		}
	}
	return false
}
func setContradictoryPresentFormPL(trace *semantic.EvidenceTrace) bool {
	changed, sawTense := false, false
	for i := range trace.Records {
		if trace.Records[i].CoreConceptID.String() != "concept:reading" {
			continue
		}
		featureValues := trace.Records[i].Features[:0]
		for _, value := range trace.Records[i].Features {
			if value.Feature == features.FeatureID("tense") {
				sawTense = true
				continue
			}
			if value.Feature == features.FeatureID("verb_form") {
				value.Value = features.FeatureValueID("third_person_singular")
				changed = true
			}
			featureValues = append(featureValues, value)
		}
		trace.Records[i].Features = featureValues
	}
	return changed && sawTense
}
func assertSourceReadingPresentOnePL(t *testing.T, c semantic.MeaningCandidate) {
	t.Helper()
	if c.Meaning == nil {
		t.Fatal("source candidate lacks Meaning")
	}
	event, ok := predicateNode(*c.Meaning)
	if !ok || event.TemporalFrame != meaning.TemporalFramePresent {
		t.Fatalf("source baseline frame is not present: %+v", event)
	}
	teacher, ok := referentConceptPL(*c.Meaning, "concept:teacher")
	if !ok || teacher.Multiplicity != meaning.MultiplicityOne {
		t.Fatalf("source teacher is not one: %+v", teacher)
	}
}
func assertReadingEvidencePL(t *testing.T, trace semantic.EvidenceTrace) {
	t.Helper()
	tense, form, readerNumber := false, false, false
	for _, record := range trace.Records {
		if record.CoreConceptID.String() == "concept:reading" {
			for _, f := range record.Features {
				if string(f.Feature) == "tense" && string(f.Value) == "present" {
					tense = true
				}
				if string(f.Feature) == "verb_form" && string(f.Value) == "finite" {
					form = true
				}
			}
		}
		if string(record.Role) == "reader" {
			for _, f := range record.Features {
				if string(f.Feature) == "number" && string(f.Value) == "singular" {
					readerNumber = true
				}
			}
		}
	}
	if !tense || !form || !readerNumber {
		t.Fatalf("test input lacks expected source evidence: tense=%t finite=%t reader_number=%t", tense, form, readerNumber)
	}
}
func logPL(t *testing.T, label string, value any, err error) {
	t.Helper()
	b, _ := json.Marshal(value)
	sha := os.Getenv("GOLINGUA_MATRIX_SHA")
	if sha == "" {
		sha = "UNRECORDED"
	}
	t.Logf("SHA=%s %s result=%s error=%v", sha, label, b, err)
}

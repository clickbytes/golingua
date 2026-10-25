package a1en_test

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
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English analyzer unavailable")
	}
	cases := []struct {
		input string
		ids   []string
	}{
		{"The teacher reads the message.", []string{"reading.teacher-message.present"}},
		{"The teacher read the message.", []string{"reading.teacher-message.past"}},
		{"The teachers read the message.", []string{"reading.teachers-message.present", "reading.teachers-message.past"}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, e := analyzer.Analyze(context.Background(), tc.input)
			logEN(t, tc.input, result, e)
			if e != nil || len(result.Candidates) == 0 {
				t.Fatalf("GAP_ANALYZE: %s %v", result.Outcome, e)
			}
			if result.Outcome != status.OutcomeOK && result.Outcome != status.OutcomeAmbiguous {
				t.Fatalf("unexpected analysis outcome: %s", result.Outcome)
			}
			matched := map[string]bool{}
			for _, candidate := range result.Candidates {
				if candidate.Meaning == nil || candidate.Status != status.OutcomeOK {
					t.Errorf("unsupported analysis candidate %q: %s", candidate.ID, candidate.Status)
					continue
				}
				candidateMatched := false
				for _, id := range tc.ids {
					want, yes, err := interlingua.Reading(id)
					if err != nil || !yes {
						t.Fatalf("fixture %s: %v", id, err)
					}
					same, err := matrix.GraphMatches(want, *candidate.Meaning)
					if err == nil && same {
						matched[id] = true
						candidateMatched = true
					}
				}
				if !candidateMatched {
					t.Errorf("FAIL: source produced an unlicensed Meaning candidate %q", candidate.ID)
				}
			}
			for _, id := range tc.ids {
				if !matched[id] {
					t.Errorf("FAIL: no candidate matched %s", id)
				}
			}
			if len(tc.ids) == 2 && len(matched) != 2 {
				t.Errorf("FAIL: plural read did not preserve both frames: %v", matched)
			}
		})
	}
}

func TestMeaningAttributesRealization(t *testing.T) {
	presentTexts := []string{"The teacher reads the message."}
	matrix.RunRealization(t, features.LanguageID("en"), []matrix.RealizationCase{
		{ID: "H-I-EN-reading-teacher-present", MeaningID: "reading.teacher-message.present", AllowedTexts: presentTexts, RequiredTexts: presentTexts, AllowedAssumptionFields: []string{"person", "determiner"}, RequiredAssumptionFields: []string{"determiner"}},
		{ID: "H-I-EN-reading-teacher-past", MeaningID: "reading.teacher-message.past", AllowedTexts: []string{"A teacher read a message.", "A teacher read the message.", "The teacher read a message.", "The teacher read the message."}, RequiredTexts: []string{"A teacher read a message.", "A teacher read the message.", "The teacher read a message.", "The teacher read the message."}, AllowedAssumptionFields: []string{"person", "determiner"}},
		{ID: "H-I-EN-reading-teachers-present", MeaningID: "reading.teachers-message.present", AllowedTexts: []string{"The teachers read the message."}, RequiredTexts: []string{"The teachers read the message."}, AllowedAssumptionFields: []string{"person", "determiner"}, RequiredAssumptionFields: []string{"determiner"}},
		{ID: "H-I-EN-reading-teachers-past", MeaningID: "reading.teachers-message.past", AllowedTexts: []string{"The teachers read a message.", "The teachers read the message."}, RequiredTexts: []string{"The teachers read a message.", "The teachers read the message."}, AllowedAssumptionFields: []string{"person", "determiner"}},
	})
}

func TestMeaningAttributesSeeingMultiplicity(t *testing.T) {
	base, ok, e := interlingua.Seeing("seeing.teacher-dog")
	if e != nil || !ok {
		t.Fatalf("fixture: %v", e)
	}
	for _, tc := range []struct {
		name  string
		node  meaning.NodeID
		texts []string
	}{{"plural teacher", "t", []string{"The teachers see a dog.", "The teachers see the dog."}}, {"plural dog", "d", []string{"A teacher sees the dogs.", "The teacher sees the dogs."}}} {
		t.Run(tc.name, func(t *testing.T) {
			g := setMultiplicityEN(t, base, tc.node, meaning.MultiplicityMany)
			r := realizeEN(t, g, "seeing-"+tc.name)
			if !sameTextEN(textsEN(r.Alternatives), tc.texts) {
				t.Errorf("FAIL: %v want %v outcome=%s", textsEN(r.Alternatives), tc.texts, r.Outcome)
			}
			assertNoTenseNumberEN(t, r.Alternatives)
		})
	}
	negative, _, ok, e := interlingua.Negation("negation.seeing.teacher-dog")
	if e != nil || !ok {
		t.Fatalf("negative fixture: %v", e)
	}
	g := setMultiplicityEN(t, negative, "t", meaning.MultiplicityMany)
	r := realizeEN(t, g, "negative plural teacher")
	if !sameTextEN(textsEN(r.Alternatives), []string{"The teachers do not see a dog.", "The teachers do not see the dog."}) {
		t.Errorf("FAIL: negative plural texts=%v outcome=%s", textsEN(r.Alternatives), r.Outcome)
	}
	assertNoTenseNumberEN(t, r.Alternatives)
}

func TestMeaningAttributesMassAnalysis(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: analyzer unavailable")
	}
	result, e := analyzer.Analyze(context.Background(), "The teacher got money.")
	logEN(t, "mass money", result, e)
	if e != nil || len(result.Candidates) == 0 {
		t.Fatalf("GAP_ANALYZE: %s %v", result.Outcome, e)
	}
	valid := false
	for _, c := range result.Candidates {
		if c.Meaning == nil {
			continue
		}
		p, ok := predicateEN(*c.Meaning)
		if !ok || p.TemporalFrame != meaning.TemporalFramePast {
			continue
		}
		teacher, tok := referentConceptEN(*c.Meaning, "concept:teacher")
		money, mok := referentConceptEN(*c.Meaning, "concept:money")
		if tok && mok && teacher.Multiplicity == meaning.MultiplicityOne && money.Multiplicity == "" {
			valid = true
		}
	}
	if !valid {
		t.Error("FAIL: no past receive candidate with teacher=one and money multiplicity unknown")
	}
}

func TestMeaningAttributesRefuseConflicts(t *testing.T) {
	positiveFixture, positiveOK, positiveErr := interlingua.Identification("identification.classification.this")
	if positiveErr != nil || !positiveOK {
		t.Fatalf("positive classification fixture: %v", positiveErr)
	}
	positiveBaseline := candidateEN(positiveFixture.Graph, "classification-positive-baseline", positiveFixture.Context)
	validateCandidateEN(t, positiveBaseline)
	baselineResult, baselineErr := targetEN(t).Realize(context.Background(), positiveBaseline)
	logEN(t, "classification valid positive baseline", baselineResult, baselineErr)
	requireSurfaceEN(t, baselineResult, baselineErr)
	for _, tc := range []struct {
		name string
		id   string
		edit func(*meaning.Meaning)
	}{
		{"classification past", "identification.classification.this", func(g *meaning.Meaning) { frameEN(g, meaning.TemporalFramePast) }},
		{"classification many", "identification.classification.this", func(g *meaning.Meaning) { multiplicityEN(g, "r", meaning.MultiplicityMany) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fic, ok, e := interlingua.Identification(tc.id)
			if e != nil || !ok {
				t.Fatalf("fixture: %v", e)
			}
			g, ctx := fic.Graph, fic.Context
			baselineCandidate := candidateEN(g, "classification-valid-baseline", ctx)
			validateCandidateEN(t, baselineCandidate)
			baseResult, baseErr := targetEN(t).Realize(context.Background(), baselineCandidate)
			logEN(t, "classification valid baseline", baseResult, baseErr)
			requireSurfaceEN(t, baseResult, baseErr)
			g = cloneMeaningEN(t, g)
			tc.edit(&g)
			candidate := candidateEN(g, tc.id, ctx)
			candidate.Contexts[0].Meaning = g
			validateCandidateEN(t, candidate)
			r, e := targetEN(t).Realize(context.Background(), candidate)
			logEN(t, "refusal "+tc.name, r, e)
			assertNoTextEN(t, r, e)
		})
	}
	negative, negativeContext, ok, e := interlingua.Negation("negation.classification.this")
	if e != nil || !ok || negativeContext == nil {
		t.Fatalf("negation fixture: %v", e)
	}
	contextMany := cloneMeaningEN(t, negative)
	multiplicityEN(&contextMany, "r", meaning.MultiplicityMany)
	candidate := candidateEN(negative, "negative-context-many", negativeContext.Context)
	candidate.Contexts[0].Meaning = contextMany
	negativeBaseline := candidateEN(negative, "negative-context-baseline", negativeContext.Context)
	validateCandidateEN(t, negativeBaseline)
	baselineResult, baselineErr = targetEN(t).Realize(context.Background(), negativeBaseline)
	logEN(t, "negative classification valid baseline", baselineResult, baselineErr)
	requireSurfaceEN(t, baselineResult, baselineErr)
	validateCandidateEN(t, candidate)
	r, e := targetEN(t).Realize(context.Background(), candidate)
	logEN(t, "negative context multiplicity mismatch", r, e)
	assertNoTextEN(t, r, e)
	// A valid source graph paired with a context graph that claims a different frame is refused.
	fic, ok, e := interlingua.Identification("identification.classification.this")
	if e != nil || !ok {
		t.Fatalf("fixture: %v", e)
	}
	g := fic.Graph
	ctxGraph := cloneMeaningEN(t, g)
	frameEN(&ctxGraph, meaning.TemporalFramePast)
	candidate = candidateEN(g, "mismatched-context", fic.Context)
	candidate.Contexts[0].Meaning = ctxGraph
	contextBaseline := candidateEN(g, "context-valid-baseline", fic.Context)
	validateCandidateEN(t, contextBaseline)
	baselineResult, baselineErr = targetEN(t).Realize(context.Background(), contextBaseline)
	logEN(t, "context mismatch positive baseline", baselineResult, baselineErr)
	requireSurfaceEN(t, baselineResult, baselineErr)
	validateCandidateEN(t, candidate)
	r, e = targetEN(t).Realize(context.Background(), candidate)
	logEN(t, "context/source frame mismatch", r, e)
	assertNoTextEN(t, r, e)
}

func TestMeaningAttributesEvidenceConflictAndUnknownFallback(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: analyzer unavailable")
	}
	target := targetEN(t)
	for _, tc := range []struct {
		name, input string
		edit        func(*meaning.Meaning)
	}{{"time conflict", "The teacher reads the message.", func(g *meaning.Meaning) { frameEN(g, meaning.TemporalFramePast) }}, {"number conflict", "The teacher reads the message.", func(g *meaning.Meaning) { setReferentConceptEN(g, "concept:teacher", meaning.MultiplicityMany) }}} {
		t.Run(tc.name, func(t *testing.T) {
			res, e := analyzer.Analyze(context.Background(), tc.input)
			logEN(t, tc.input, res, e)
			if e != nil || len(res.Candidates) == 0 {
				t.Fatalf("GAP_ANALYZE: %s %v", res.Outcome, e)
			}
			candidate := res.Candidates[0]
			candidate.Evidence = candidate.Evidence.Clone()
			candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
			if candidate.Meaning == nil {
				t.Fatal("candidate lacks Meaning")
			}
			assertSourceReadingPresentOneEN(t, candidate)
			assertReadingEvidenceEN(t, candidate.Evidence)
			g := cloneMeaningEN(t, *candidate.Meaning)
			tc.edit(&g)
			candidate.Meaning = &g
			for i := range candidate.Contexts {
				candidate.Contexts[i].Meaning = g
			}
			validateCandidateEN(t, candidate)
			baselineResult, baselineErr := target.Realize(context.Background(), res.Candidates[0])
			logEN(t, "valid source candidate baseline", baselineResult, baselineErr)
			requireSurfaceEN(t, baselineResult, baselineErr)
			surf, e := target.Realize(context.Background(), candidate)
			logEN(t, "I/evidence conflict", surf, e)
			assertNoTextEN(t, surf, e)
		})
	}
	// A finite present-form bundle without separate tense still cannot override a past Meaning frame.
	res, e := analyzer.Analyze(context.Background(), "The teacher reads the message.")
	if e != nil || len(res.Candidates) == 0 {
		t.Fatalf("present source analysis: %s %v", res.Outcome, e)
	}
	candidate := res.Candidates[0]
	candidate.Evidence = candidate.Evidence.Clone()
	candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
	if candidate.Meaning == nil {
		t.Fatal("candidate lacks Meaning")
	}
	assertSourceReadingPresentOneEN(t, candidate)
	assertReadingEvidenceEN(t, candidate.Evidence)
	g := cloneMeaningEN(t, *candidate.Meaning)
	frameEN(&g, meaning.TemporalFramePast)
	candidate.Meaning = &g
	for i := range candidate.Contexts {
		candidate.Contexts[i].Meaning = g
	}
	if !setContradictoryPresentFormEN(&candidate.Evidence) {
		t.Fatal("review trap failed: no reading verb-form evidence changed")
	}
	validateCandidateEN(t, candidate)
	surface, err := target.Realize(context.Background(), candidate)
	logEN(t, "past I versus present form without tense", surface, err)
	assertNoTextEN(t, surface, err)
	res, e = analyzer.Analyze(context.Background(), "The teacher read the message.")
	if e != nil || len(res.Candidates) == 0 {
		t.Fatalf("past analysis: %s %v", res.Outcome, e)
	}
	candidate = res.Candidates[0]
	candidate.Evidence = candidate.Evidence.Clone()
	candidate.Contexts = append([]semantic.MeaningContext(nil), candidate.Contexts...)
	g = cloneMeaningEN(t, *candidate.Meaning)
	frameEN(&g, "")
	setReferentConceptEN(&g, "concept:teacher", "")
	candidate.Meaning = &g
	for i := range candidate.Contexts {
		candidate.Contexts[i].Meaning = g
	}
	validateCandidateEN(t, candidate)
	surf, e := target.Realize(context.Background(), candidate)
	logEN(t, "unknown I with retained evidence", surf, e)
	if e != nil || len(textsEN(surf.Alternatives)) == 0 {
		assertNoTextEN(t, surf, e)
	} else {
		if !sameTextEN(textsEN(surf.Alternatives), []string{"The teacher reads the message."}) {
			t.Errorf("unknown frame used past evidence instead of declared present default: %+v", surf.Alternatives)
		}
		for _, alt := range surf.Alternatives {
			if alt.Text != "" && (!hasAssumptionEN(alt.Assumptions, "tense") || !hasAssumptionEN(alt.Assumptions, "number")) {
				t.Errorf("FAIL: missing I fields realized without both explicit assumptions: %+v", alt)
			}
		}
	}
}

func TestMeaningAttributesMassKnownMultiplicityRefusal(t *testing.T) {
	runtime := matrix.Runtime(t)
	a, ok := runtime.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("analyzer unavailable")
	}
	res, e := a.Analyze(context.Background(), "The teacher got money.")
	if e != nil || len(res.Candidates) == 0 {
		t.Fatalf("analysis unavailable: %s %v", res.Outcome, e)
	}
	for _, c := range res.Candidates {
		if c.Meaning == nil {
			continue
		}
		c.Evidence = c.Evidence.Clone()
		c.Contexts = append([]semantic.MeaningContext(nil), c.Contexts...)
		g := cloneMeaningEN(t, *c.Meaning)
		setReferentConceptEN(&g, "concept:money", meaning.MultiplicityMany)
		c.Meaning = &g
		for i := range c.Contexts {
			c.Contexts[i].Meaning = g
		}
		validateCandidateEN(t, c)
		surface, e := targetEN(t).Realize(context.Background(), c)
		logEN(t, "mass multiplicity refusal", surface, e)
		assertNoTextEN(t, surface, e)
		return
	}
	t.Fatal("money candidate missing Meaning")
}

func realizeEN(t *testing.T, g meaning.Meaning, id string) engine.SurfaceResult {
	t.Helper()
	c := candidateEN(g, id, meaning.RequestContext{})
	validateCandidateEN(t, c)
	r, e := targetEN(t).Realize(context.Background(), c)
	logEN(t, id, r, e)
	if e != nil {
		t.Errorf("target error: %v", e)
	}
	return r
}
func targetEN(t *testing.T) engine.TargetRealizer {
	t.Helper()
	r, ok := matrix.Runtime(t).TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	return r
}
func candidateEN(g meaning.Meaning, id string, ctx meaning.RequestContext) semantic.MeaningCandidate {
	c := semantic.MeaningCandidate{ID: id, SemanticID: id, Meaning: &g, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: id, Outcome: status.OutcomeOK}}
	if ctx.ID != "" {
		c.Contexts = []semantic.MeaningContext{{Meaning: g, Context: ctx}}
	}
	return c
}
func cloneMeaningEN(t *testing.T, g meaning.Meaning) meaning.Meaning {
	t.Helper()
	r, e := meaning.Build(g.Roots, g.Nodes, meaning.DefaultBudget())
	if e != nil {
		t.Fatalf("clone graph: %v", e)
	}
	return r
}
func setMultiplicityEN(t *testing.T, g meaning.Meaning, id meaning.NodeID, v meaning.Multiplicity) meaning.Meaning {
	g = cloneMeaningEN(t, g)
	multiplicityEN(&g, id, v)
	return g
}
func multiplicityEN(g *meaning.Meaning, id meaning.NodeID, v meaning.Multiplicity) {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			g.Nodes[i].Multiplicity = v
		}
	}
}
func frameEN(g *meaning.Meaning, v meaning.TemporalFrame) {
	for i := range g.Nodes {
		if g.Nodes[i].Kind == meaning.KindEvent || g.Nodes[i].Kind == meaning.KindState {
			g.Nodes[i].TemporalFrame = v
		}
	}
}
func predicateEN(g meaning.Meaning) (meaning.Node, bool) {
	for _, n := range g.Nodes {
		if n.Kind == meaning.KindEvent || n.Kind == meaning.KindState {
			return n, true
		}
	}
	return meaning.Node{}, false
}
func roleReferentEN(g meaning.Meaning, id meaning.NodeID, role string) (meaning.Node, bool) {
	nodes := map[meaning.NodeID]meaning.Node{}
	for _, n := range g.Nodes {
		nodes[n.ID] = n
	}
	for _, n := range g.Nodes {
		if n.ID == id {
			for _, p := range n.Participants {
				if string(p.Role) == role {
					v, ok := nodes[p.Target]
					return v, ok
				}
			}
		}
	}
	return meaning.Node{}, false
}
func referentConceptEN(g meaning.Meaning, concept string) (meaning.Node, bool) {
	for _, n := range g.Nodes {
		if n.Kind == meaning.KindReferent && string(n.Concept) == concept {
			return n, true
		}
	}
	return meaning.Node{}, false
}
func setReferentConceptEN(g *meaning.Meaning, concept string, value meaning.Multiplicity) {
	for i := range g.Nodes {
		if g.Nodes[i].Kind == meaning.KindReferent && string(g.Nodes[i].Concept) == concept {
			g.Nodes[i].Multiplicity = value
		}
	}
}
func textsEN(a []engine.SurfaceAlternative) []string {
	m := map[string]bool{}
	for _, x := range a {
		if x.Text != "" {
			m[x.Text] = true
		}
	}
	r := []string{}
	for s := range m {
		r = append(r, s)
	}
	sort.Strings(r)
	return r
}
func sameTextEN(a, b []string) bool {
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
func assertNoTenseNumberEN(t *testing.T, a []engine.SurfaceAlternative) {
	t.Helper()
	for _, x := range a {
		for _, v := range x.Assumptions {
			if v.Field == "tense" || v.Field == "number" {
				t.Errorf("known Meaning attr got assumption: %+v", v)
			}
		}
	}
}
func assertAllowedAssumptionsEN(t *testing.T, a []engine.SurfaceAlternative) {
	t.Helper()
	for _, x := range a {
		for _, v := range x.Assumptions {
			if v.Field != "person" && v.Field != "determiner" {
				t.Errorf("unexpected assumption for known I realization: %+v", v)
			}
		}
	}
}
func assertNoTextEN(t *testing.T, r engine.SurfaceResult, e error) {
	t.Helper()
	if e != nil {
		if isSetupErrorEN(e) {
			t.Errorf("INFRA: refusal came from runtime/setup error: %v", e)
		}
		return
	}
	if r.Outcome == status.OutcomeOK || r.Outcome == status.OutcomeAmbiguous || r.Outcome == status.OutcomeIncomplete || len(r.Diagnostics) == 0 {
		t.Errorf("FAIL: refusal requires an explicit non-success outcome and diagnostic; outcome=%s diagnostics=%+v", r.Outcome, r.Diagnostics)
	}
	for _, a := range r.Alternatives {
		if a.Text != "" {
			t.Errorf("FAIL: invalid input realized %q", a.Text)
		}
	}
}
func validateCandidateEN(t *testing.T, c semantic.MeaningCandidate) {
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
func requireSurfaceEN(t *testing.T, r engine.SurfaceResult, e error) {
	t.Helper()
	if e != nil {
		t.Fatalf("valid baseline failed: %v", e)
	}
	for _, alt := range r.Alternatives {
		if alt.Text != "" {
			return
		}
	}
	t.Fatalf("valid baseline returned no surface: %+v", r)
}
func isSetupErrorEN(e error) bool {
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
func hasAssumptionEN(a []realizer.FieldAssumption, key string) bool {
	for _, v := range a {
		if v.Field == key {
			return true
		}
	}
	return false
}
func setContradictoryPresentFormEN(trace *semantic.EvidenceTrace) bool {
	changed, sawTense := false, false
	for i := range trace.Records {
		if trace.Records[i].CoreConceptID.String() != "concept:reading" {
			continue
		}
		kept := trace.Records[i].Features[:0]
		for _, value := range trace.Records[i].Features {
			if value.Feature == features.FeatureID("tense") {
				sawTense = true
				continue
			}
			if value.Feature == features.FeatureID("verb_form") {
				value.Value = features.FeatureValueID("third_person_singular")
				changed = true
			}
			kept = append(kept, value)
		}
		trace.Records[i].Features = kept
	}
	return changed && sawTense
}
func assertSourceReadingPresentOneEN(t *testing.T, c semantic.MeaningCandidate) {
	t.Helper()
	if c.Meaning == nil {
		t.Fatal("source candidate lacks Meaning")
	}
	p, ok := predicateEN(*c.Meaning)
	if !ok || p.TemporalFrame != meaning.TemporalFramePresent {
		t.Fatalf("source baseline frame not present: %+v", p)
	}
	teacher, ok := referentConceptEN(*c.Meaning, "concept:teacher")
	if !ok || teacher.Multiplicity != meaning.MultiplicityOne {
		t.Fatalf("source teacher is not one: %+v", teacher)
	}
}
func assertReadingEvidenceEN(t *testing.T, trace semantic.EvidenceTrace) {
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
func logEN(t *testing.T, label string, v any, e error) {
	t.Helper()
	b, _ := json.Marshal(v)
	sha := os.Getenv("GOLINGUA_MATRIX_SHA")
	if sha == "" {
		sha = "UNRECORDED"
	}
	t.Logf("SHA=%s %s result=%s error=%v", sha, label, b, e)
}

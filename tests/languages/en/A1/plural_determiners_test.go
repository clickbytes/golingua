package a1en_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestPluralDeterminersIndependentSeeingAndReading(t *testing.T) {
	rt := matrix.Runtime(t)
	target, ok := rt.TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	cases := []struct {
		id         string
		want       []string
		determiner map[string]map[string]string
	}{
		{"a1.12.amb.007.pl", []string{"Teachers see birds.", "The teachers see birds.", "Teachers see the birds.", "The teachers see the birds."}, map[string]map[string]string{
			"Teachers see birds.": {"experiencer": "bare", "stimulus": "bare"}, "The teachers see birds.": {"experiencer": "definite", "stimulus": "bare"},
			"Teachers see the birds.": {"experiencer": "bare", "stimulus": "definite"}, "The teachers see the birds.": {"experiencer": "definite", "stimulus": "definite"},
		}},
		{"a1.12.amb.003.pl", []string{"A mother reads a book.", "A mother reads the book.", "The mother reads a book.", "The mother reads the book."}, map[string]map[string]string{
			"A mother reads a book.": {"reader": "indefinite", "material": "indefinite"}, "A mother reads the book.": {"reader": "indefinite", "material": "definite"},
			"The mother reads a book.": {"reader": "definite", "material": "indefinite"}, "The mother reads the book.": {"reader": "definite", "material": "definite"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			fic, found, err := interlingua.AmbiguityExpansion(tc.id)
			if err != nil || !found {
				t.Fatalf("fixture %s unavailable: %v", tc.id, err)
			}
			before, _ := json.Marshal(fic.Graph)
			c := candidateEN(fic.Graph, tc.id, fic.Context)
			validateCandidateEN(t, c)
			r, err := target.Realize(context.Background(), c)
			logEN(t, "I->EN "+tc.id, r, err)
			if err != nil || len(r.Alternatives) != 4 || !sameTextEN(textsEN(r.Alternatives), tc.want) {
				t.Errorf("want exactly 4 licensed determiner combinations %v, got outcome=%s texts=%v err=%v", tc.want, r.Outcome, textsEN(r.Alternatives), err)
			}
			assertDeterminerAssumptionsForRoles(t, fic.Graph, r.Alternatives, tc.determiner)
			after, _ := json.Marshal(fic.Graph)
			if string(before) != string(after) {
				t.Error("target realization mutated the independent Meaning graph")
			}
		})
	}
	fic, found, err := interlingua.TimeExpansion("a1.10.time.021.en")
	if err != nil || !found {
		t.Fatalf("past reading fixture unavailable: %v", err)
	}
	c := candidateEN(fic.Graph, "a1.10.time.021.en", fic.Context)
	validateCandidateEN(t, c)
	r, err := target.Realize(context.Background(), c)
	logEN(t, "I->EN past reading", r, err)
	want := []string{"A mother read a letter.", "A mother read the letter.", "The mother read a letter.", "The mother read the letter."}
	if err != nil || len(r.Alternatives) != 4 || !sameTextEN(textsEN(r.Alternatives), want) {
		t.Errorf("past reading must retain four determiner choices: outcome=%s texts=%v err=%v", r.Outcome, textsEN(r.Alternatives), err)
	}
	assertDeterminerAssumptionsForRoles(t, fic.Graph, r.Alternatives, map[string]map[string]string{
		"A mother read a letter.": {"reader": "indefinite", "material": "indefinite"}, "A mother read the letter.": {"reader": "indefinite", "material": "definite"},
		"The mother read a letter.": {"reader": "definite", "material": "indefinite"}, "The mother read the letter.": {"reader": "definite", "material": "definite"},
	})
}

func TestPluralDeterminersPreserveActualEnglishSource(t *testing.T) {
	rt := matrix.Runtime(t)
	analyzer, ok := rt.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English source unavailable")
	}
	target, ok := rt.TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	for _, tc := range []struct {
		input string
		roles map[string]string
		text  string
	}{
		{"The teachers see the birds.", map[string]string{"experiencer": "definite", "stimulus": "definite"}, "The teachers see the birds."},
		{"Teachers see birds.", map[string]string{"experiencer": "bare", "stimulus": "bare"}, "Teachers see birds."},
		{"A mother reads a letter.", map[string]string{"reader": "indefinite", "material": "indefinite"}, "A mother reads a letter."},
	} {
		t.Run(tc.input, func(t *testing.T) {
			src, err := analyzer.Analyze(context.Background(), tc.input)
			logEN(t, "actual EN source "+tc.input, src, err)
			if err != nil {
				t.Fatalf("source analyzer: %v", err)
			}
			count := 0
			for _, cand := range src.Candidates {
				if cand.Status != status.OutcomeOK || cand.Meaning == nil {
					continue
				}
				count++
				for role, want := range tc.roles {
					if got := sourceDeterminerForRole(cand, role); got != want {
						t.Errorf("source role %s determiner=%q want %q", role, got, want)
					}
				}
				dst, e := target.Realize(context.Background(), cand)
				logEN(t, "same-language adapter "+tc.input, dst, e)
				if e != nil || dst.Outcome != status.OutcomeOK || !sameTextEN(textsEN(dst.Alternatives), []string{tc.text}) {
					t.Errorf("known source determiners should retain one surface: outcome=%s texts=%v err=%v", dst.Outcome, textsEN(dst.Alternatives), e)
				}
				for _, alt := range dst.Alternatives {
					if len(alt.Assumptions) != 0 || len(alt.Losses) != 0 {
						t.Errorf("known source determiner produced assumptions/losses: %+v %+v", alt.Assumptions, alt.Losses)
					}
				}
			}
			if count == 0 {
				t.Error("source produced no complete supported Meaning")
			}
		})
	}
}

func TestPluralDeterminersSeparateNominalAndPersonalReadingSources(t *testing.T) {
	rt := matrix.Runtime(t)
	analyzer, ok := rt.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English source unavailable")
	}
	source, err := analyzer.Analyze(context.Background(), "Teachers read letters.")
	logEN(t, "nominal plural read disambiguation", source, err)
	if err != nil {
		t.Fatalf("source analyzer: %v", err)
	}
	frames := map[meaning.TemporalFrame]bool{}
	readingCandidates := 0
	for _, candidate := range source.Candidates {
		if candidate.Status == status.OutcomeOK && candidate.Meaning != nil {
			predicate, found := predicateEN(*candidate.Meaning)
			if found && predicate.Concept == "concept:reading" {
				readingCandidates++
				frames[predicate.TemporalFrame] = true
			}
		}
		if candidate.Status != status.OutcomeOK {
			for _, record := range candidate.Evidence.Records {
				if record.FamilyID.String() == "construction:en.transitive_personal" {
					t.Errorf("nominal plural reading retained unresolved personal-construction candidate %q", candidate.ID)
					break
				}
			}
		}
	}
	if readingCandidates != 2 || !frames[meaning.TemporalFramePresent] || !frames[meaning.TemporalFramePast] {
		t.Errorf("nominal plural read should retain exactly present and past readings, got %d candidates in frames %v", readingCandidates, frames)
	}

	target, ok := rt.TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	for _, tc := range []struct {
		id, pronoun, count, anchor string
	}{
		{"a1.10.time.009.en", "I", "one", "speaker"},
		{"a1.10.time.011.en", "We", "many", "speaker_inclusive_group"},
		{"a1.10.time.010.en", "You", "", "addressee"},
	} {
		t.Run(tc.pronoun, func(t *testing.T) {
			fic, found, fixtureErr := interlingua.TimeExpansion(tc.id)
			if fixtureErr != nil || !found {
				t.Fatalf("fixture %s unavailable: %v", tc.id, fixtureErr)
			}
			predicate, found := predicateEN(fic.Graph)
			if !found {
				t.Fatal("fixture lacks predicate")
			}
			reader, found := roleReferentEN(fic.Graph, predicate.ID, "reader")
			if !found || string(reader.Multiplicity) != tc.count {
				t.Errorf("reader multiplicity=%q, want %q", reader.Multiplicity, tc.count)
			}
			anchored := false
			for _, mention := range fic.Context.Mentions {
				if mention.NodeID == reader.ID && mention.Anchor != nil && mention.Anchor.Kind == "participant" && mention.Anchor.Value == tc.anchor {
					anchored = true
				}
			}
			if !anchored {
				t.Errorf("reader lacks expected participant anchor %q", tc.anchor)
			}
			candidate := candidateEN(fic.Graph, tc.id, fic.Context)
			validateCandidateEN(t, candidate)
			result, realizeErr := target.Realize(context.Background(), candidate)
			logEN(t, "I->EN personal reading "+tc.id, result, realizeErr)
			if realizeErr != nil || len(result.Alternatives) == 0 {
				t.Fatalf("personal reading target unavailable: outcome=%s error=%v", result.Outcome, realizeErr)
			}
			for _, alternative := range result.Alternatives {
				if !strings.HasPrefix(alternative.Text, tc.pronoun+" read ") {
					t.Errorf("personal reading target lost pronoun or tense: %q", alternative.Text)
				}
			}
		})
	}
}

func TestPluralDeterminersNegativeIndependentChoices(t *testing.T) {
	rt := matrix.Runtime(t)
	target, ok := rt.TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	for _, tc := range []struct {
		id         string
		want       []string
		aux        string
		determiner map[string]map[string]string
	}{
		{"a1.8.neg.019", []string{"A mother does not hear birds.", "A mother does not hear the birds.", "The mother does not hear birds.", "The mother does not hear the birds."}, "does not hear", map[string]map[string]string{
			"A mother does not hear birds.": {"experiencer": "indefinite", "stimulus": "bare"}, "A mother does not hear the birds.": {"experiencer": "indefinite", "stimulus": "definite"},
			"The mother does not hear birds.": {"experiencer": "definite", "stimulus": "bare"}, "The mother does not hear the birds.": {"experiencer": "definite", "stimulus": "definite"},
		}},
		{"a1.8.neg.020", []string{"Doctors do not see a car.", "Doctors do not see the car.", "The doctors do not see a car.", "The doctors do not see the car."}, "do not see", map[string]map[string]string{
			"Doctors do not see a car.": {"experiencer": "bare", "stimulus": "indefinite"}, "Doctors do not see the car.": {"experiencer": "bare", "stimulus": "definite"},
			"The doctors do not see a car.": {"experiencer": "definite", "stimulus": "indefinite"}, "The doctors do not see the car.": {"experiencer": "definite", "stimulus": "definite"},
		}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			g, found, err := interlingua.NegationExpansion(tc.id)
			if err != nil || !found {
				t.Fatalf("fixture unavailable: %v", err)
			}
			c := candidateEN(g, tc.id, meaning.RequestContext{})
			validateCandidateEN(t, c)
			r, err := target.Realize(context.Background(), c)
			logEN(t, "I->EN "+tc.id, r, err)
			if err != nil || len(r.Alternatives) != 4 || !sameTextEN(textsEN(r.Alternatives), tc.want) {
				t.Errorf("expected four negative determiner choices: outcome=%s texts=%v err=%v", r.Outcome, textsEN(r.Alternatives), err)
			}
			for _, a := range r.Alternatives {
				if !strings.Contains(a.Text, tc.aux) {
					t.Errorf("wrong auxiliary/verb for frame: %q", a.Text)
				}
			}
			assertDeterminerAssumptionsForRoles(t, g, r.Alternatives, tc.determiner)
		})
	}
}

func TestPluralDeterminersParticipantsAndThisControl(t *testing.T) {
	rt := matrix.Runtime(t)
	pl, ok := rt.SourceAnalyzerFor("pl")
	if !ok {
		t.Fatal("INFRA: Polish source unavailable")
	}
	enSource, ok := rt.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English source unavailable")
	}
	enTarget, ok := rt.TargetRealizerFor("en")
	if !ok {
		t.Fatal("INFRA: English target unavailable")
	}
	for _, input := range []string{"Widzę psy."} {
		src, err := pl.Analyze(context.Background(), input)
		logEN(t, "PL personal source "+input, src, err)
		if err != nil {
			t.Fatalf("Polish source: %v", err)
		}
		var supported int
		got := []string{}
		for _, c := range src.Candidates {
			if c.Status == status.OutcomeOK && c.Meaning != nil {
				supported++
				dst, e := enTarget.Realize(context.Background(), c)
				logEN(t, "PL->EN "+input, dst, e)
				if e != nil {
					t.Errorf("target: %v", e)
					continue
				}
				got = append(got, textsEN(dst.Alternatives)...)
			}
		}
		want := []string{"I see dogs.", "I see the dogs."}
		if supported == 0 || !sameTextEN(got, want) {
			t.Errorf("expected first-person plural-object choices %v; got %v", want, got)
		}
		for _, s := range got {
			if strings.Contains(s, "a dogs") {
				t.Errorf("ungrammatical plural indefinite: %q", s)
			}
		}
	}
	we, err := enSource.Analyze(context.Background(), "We see dogs.")
	logEN(t, "EN We source", we, err)
	if err != nil {
		t.Fatalf("EN We source: %v", err)
	}
	anchored := false
	for _, c := range we.Candidates {
		if c.Status != status.OutcomeOK || c.Meaning == nil {
			continue
		}
		for _, ctx := range c.Contexts {
			for _, m := range ctx.Context.Mentions {
				if m.Anchor != nil && m.Anchor.Value == "speaker_inclusive_group" {
					anchored = true
				}
			}
		}
	}
	if !anchored {
		t.Error("We source did not preserve speaker-inclusive participant anchor")
	}
	this, err := enSource.Analyze(context.Background(), "This window is black.")
	logEN(t, "This window source control", this, err)
	if err != nil || len(this.Candidates) == 0 {
		t.Errorf("This window source regression: outcome=%s err=%v", this.Outcome, err)
	}
}

func TestPluralDeterminersRejectMalformedActualSources(t *testing.T) {
	rt := matrix.Runtime(t)
	analyzer, ok := rt.SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English source unavailable")
	}
	for _, input := range []string{"Teacher sees the cat.", "The teacher sees dog.", "The teacher sees a dogs.", "The teacher sees an dog.", "The teacher sees a apple."} {
		t.Run(input, func(t *testing.T) {
			r, err := analyzer.Analyze(context.Background(), input)
			logEN(t, "malformed source "+input, r, err)
			if err != nil {
				t.Fatalf("unexpected infrastructure error: %v", err)
			}
			for _, c := range r.Candidates {
				if c.Status == status.OutcomeOK && c.Meaning != nil {
					t.Errorf("malformed sentence received supported Meaning: %q", input)
				}
			}
			if r.Outcome != status.OutcomeUnsupported && r.Outcome != status.OutcomeInvalid {
				t.Errorf("expected controlled source refusal, got %s", r.Outcome)
			}
		})
	}
}

func sourceDeterminerForRole(c semantic.MeaningCandidate, role string) string {
	for _, rec := range c.Evidence.Records {
		if string(rec.Role) == role {
			for _, f := range rec.Features {
				if string(f.Feature) == "determiner" && f.State == "known" {
					return string(f.Value)
				}
			}
		}
	}
	return ""
}

func assertDeterminerAssumptionsForRoles(t *testing.T, g meaning.Meaning, alts []engine.SurfaceAlternative, expected map[string]map[string]string) {
	t.Helper()
	p, ok := predicateEN(g)
	if !ok {
		t.Fatal("fixture has no predicate")
	}
	for _, role := range []string{"experiencer", "stimulus", "reader", "material"} {
		var targetID meaning.NodeID
		for _, part := range p.Participants {
			if string(part.Role) == role {
				targetID = part.Target
			}
		}
		if targetID == "" {
			for _, values := range expected {
				if _, required := values[role]; required {
					t.Errorf("expected determiner role %s is absent from fixture", role)
					break
				}
			}
			continue
		}
		path, found := runtime.MeaningPath(g, targetID)
		if !found {
			t.Fatalf("no Meaning path for role %s", role)
		}
		for _, a := range alts {
			roleValues, hasSurface := expected[a.Text]
			if !hasSurface {
				t.Errorf("surface %q has no explicit determiner expectation", a.Text)
				continue
			}
			want, roleExpected := roleValues[role]
			if !roleExpected {
				t.Errorf("surface %q lacks an expected determiner value for role %s", a.Text, role)
				continue
			}
			count := 0
			for _, ass := range a.Assumptions {
				if ass.Field == "determiner" && fmt.Sprint(ass.Path) == fmt.Sprint(path) {
					count++
					if ass.Value != want {
						t.Errorf("surface %q role %s assumption=%q, want %q", a.Text, role, ass.Value, want)
					}
				}
			}
			if count != 1 {
				t.Errorf("surface %q has %d determiner assumptions for role %s path %v (expected %q)", a.Text, count, role, path, want)
			}
		}
	}
	for _, a := range alts {
		want := len(expected[a.Text])
		got := 0
		for _, ass := range a.Assumptions {
			if ass.Field == "determiner" {
				got++
			}
		}
		if got != want {
			t.Errorf("surface %q has %d determiner assumptions total, want exactly %d role-bound assumptions", a.Text, got, want)
		}
	}
}

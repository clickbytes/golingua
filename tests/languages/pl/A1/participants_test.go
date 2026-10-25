package a1pl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	runtimepipe "github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestPersonalSourcePreservesParticipantAnchors(t *testing.T) {
	engine := matrix.Runtime(t)
	adapter, ok := engine.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || adapter == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	for _, tc := range []struct {
		input, verb, multiplicity, anchor string
	}{
		{"Widzę psa.", "Widzę", "one", "speaker"},
		{"Widzimy psa.", "Widzimy", "many", "speaker_inclusive_group"},
		{"Widzisz psa.", "Widzisz", "one", "addressee"},
		{"Widzicie psa.", "Widzicie", "many", "addressee"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			result, err := adapter.Analyze(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("analyze: %v", err)
			}
			if !hasPersonalSourceCandidate(result.Candidates, tc.input, tc.verb, "seeing", tc.multiplicity, tc.anchor) {
				t.Fatalf("no seeing Meaning preserved person multiplicity %s, anchor %s, and actual VP span", tc.multiplicity, tc.anchor)
			}
		})
	}
	for _, tc := range []struct{ input, multiplicity, anchor string }{
		{"Ja widzę psa.", "one", "speaker"},
		{"My widzimy psa.", "many", "speaker_inclusive_group"},
		{"Ty widzisz psa.", "one", "addressee"},
		{"Wy widzicie psa.", "many", "addressee"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			result, err := adapter.Analyze(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("analyze: %v", err)
			}
			if !hasPersonalSourceCandidate(result.Candidates, tc.input, "", "seeing", tc.multiplicity, tc.anchor) {
				t.Fatalf("no explicit personal subject Meaning with %s anchor", tc.anchor)
			}
		})
	}
}

func TestPersonalKnownSpeakerHasOnePolishTarget(t *testing.T) {
	engine := matrix.Runtime(t)
	polish, ok := engine.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || polish == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	fixture, found, err := interlingua.TimeExpansion("a1.10.time.001.en")
	if err != nil || !found {
		t.Fatalf("load speaker I fixture: found=%t err=%v", found, err)
	}
	result, err := polish.Realize(context.Background(), personalFixtureCandidate("known-speaker", fixture))
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != status.OutcomeOK || len(result.Alternatives) != 1 || len(result.Diagnostics) != 0 {
		t.Fatalf("known speaker has no unresolved target choice: outcome=%s alternatives=%d diagnostics=%+v", result.Outcome, len(result.Alternatives), result.Diagnostics)
	}
	surface := result.Alternatives[0]
	if surface.Text != "Widzę psa." || surface.Outcome != status.OutcomeOK || len(surface.Assumptions) != 0 || len(surface.Losses) != 0 {
		t.Errorf("known speaker realization changed information or invented uncertainty: %+v", surface)
	}
}

func TestEnglishYouFixturesKeepCountUnknownAndPolishTargetsRemainIndependent(t *testing.T) {
	engine := matrix.Runtime(t)
	english, ok := engine.SourceAnalyzerFor(features.LanguageID("en"))
	if !ok || english == nil {
		t.Fatal("INFRA: English source analyzer unavailable")
	}
	polish, ok := engine.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || polish == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	for _, id := range []string{"a1.10.time.002.en", "a1.10.time.006.en", "a1.10.time.010.en"} {
		fixture, ok, err := interlingua.TimeExpansion(id)
		if err != nil || !ok {
			t.Fatalf("load I fixture %s: ok=%t error=%v", id, ok, err)
		}
		if personalSubjectMultiplicity(fixture.Graph) != "" || fixture.Context.Mentions[0].Anchor == nil || fixture.Context.Mentions[0].Anchor.Value != "addressee" {
			t.Errorf("%s encodes a count or lacks addressee anchor: graph=%+v context=%+v", id, fixture.Graph, fixture.Context)
		}
	}
	for _, tc := range []struct{ input, predicate string }{
		{"You see the cat.", "seeing"},
		{"You hear the bird.", "hearing"},
		{"You read the message.", "reading"},
	} {
		result, err := english.Analyze(context.Background(), tc.input)
		if err != nil {
			t.Fatalf("analyze %q: %v", tc.input, err)
		}
		found := 0
		for _, candidate := range result.Candidates {
			if candidate.Status != status.OutcomeOK || !hasPredicate(candidate, tc.predicate, "") {
				continue
			}
			subjectID := personalSubjectID(*candidate.Meaning, tc.predicate)
			if subjectID == "" || nodeByID(*candidate.Meaning, subjectID).Concept != "concept:person" {
				continue
			}
			found++
			if nodeByID(*candidate.Meaning, subjectID).Multiplicity != "" || !candidateHasParticipantAnchor(candidate, subjectID, "addressee") {
				t.Errorf("%q personal candidate inferred addressee count or lost anchor: %+v", tc.input, candidate.Meaning)
			}
		}
		if found == 0 {
			t.Errorf("%q produced no supported personal source candidate", tc.input)
		}
	}
	fixture, ok, err := interlingua.TimeExpansion("a1.10.time.002.en")
	if err != nil || !ok {
		t.Fatalf("load independent I fixture: ok=%t error=%v", ok, err)
	}
	candidate := personalFixtureCandidate("a1.10.time.002.en", fixture)
	original, err := json.Marshal(candidate.Meaning)
	if err != nil {
		t.Fatalf("snapshot independent Meaning: %v", err)
	}
	result, err := polish.Realize(context.Background(), candidate)
	if err != nil {
		t.Fatalf("realize independent I: %v", err)
	}
	assertSurfaceNumberAssumption(t, result.Alternatives, "Widzisz kota.", "singular")
	assertSurfaceNumberAssumption(t, result.Alternatives, "Widzicie kota.", "plural")
	if result.Outcome != status.OutcomeAmbiguous || len(result.Alternatives) != 2 {
		t.Errorf("unknown addressee count must retain exactly two number branches: outcome=%s alternatives=%d", result.Outcome, len(result.Alternatives))
	}
	current, err := json.Marshal(candidate.Meaning)
	if err != nil || !bytes.Equal(original, current) {
		t.Fatal("target realization mutated independent input Meaning")
	}
}

func TestPersonalIndependentPastReadingUsesGenderEvidenceAssumptions(t *testing.T) {
	engine := matrix.Runtime(t)
	polish, ok := engine.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || polish == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	fixture, ok, err := interlingua.TimeExpansion("a1.10.time.009.en.past")
	if err != nil || !ok {
		t.Fatalf("load past I fixture: ok=%t error=%v", ok, err)
	}
	result, err := polish.Realize(context.Background(), personalFixtureCandidate("a1.10.time.009.en.past", fixture))
	if err != nil {
		t.Fatalf("realize past I: %v", err)
	}
	assertSurfaceGenderAssumption(t, result.Alternatives, "Czytałem książkę.", "masculine")
	assertSurfaceGenderAssumption(t, result.Alternatives, "Czytałam książkę.", "feminine")
}

func TestPolishPastPersonalSourceGenderRestrictsItsTarget(t *testing.T) {
	engine := matrix.Runtime(t)
	source, ok := engine.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || source == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	polish, ok := engine.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || polish == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	result, err := source.Analyze(context.Background(), "Czytałam list.")
	if err != nil {
		t.Fatalf("analyze Polish past reading: %v", err)
	}
	found := false
	for _, candidate := range result.Candidates {
		if !hasPredicate(candidate, "reading", meaning.TemporalFramePast) || candidate.Status != status.OutcomeOK {
			continue
		}
		found = true
		target, err := polish.Realize(context.Background(), candidate)
		if err != nil {
			t.Fatalf("realize analyzed Polish Meaning: %v", err)
		}
		female := false
		for _, alternative := range target.Alternatives {
			if alternative.Text == "Czytałam list." {
				female = true
				for _, assumption := range alternative.Assumptions {
					if assumption.Field == "gender" {
						t.Errorf("known source gender was redundantly assumed: %+v", assumption)
					}
				}
			}
			if alternative.Text == "Czytałem list." {
				t.Errorf("feminine source predicate evidence allowed masculine target: %+v", alternative)
			}
		}
		if !female {
			t.Errorf("no feminine past reading target in %+v", target.Alternatives)
		}
	}
	if !found {
		t.Fatal("Polish source produced no OK past reading Meaning")
	}
}

func TestParticipantAnchorMisalignmentFailsClosedForTargets(t *testing.T) {
	engine := matrix.Runtime(t)
	fixture, ok, err := interlingua.TimeExpansion("a1.10.time.001.en")
	if err != nil || !ok {
		t.Fatalf("load speaker I fixture: ok=%t error=%v", ok, err)
	}
	type target struct {
		name string
		lang features.LanguageID
	}
	for _, selectedTarget := range []target{{"en", "en"}, {"pl", "pl"}} {
		realizer, exists := engine.TargetRealizerFor(selectedTarget.lang)
		if !exists || realizer == nil {
			t.Fatalf("INFRA: %s target realizer unavailable", selectedTarget.lang)
		}
		for _, makeCase := range []struct {
			name  string
			build func() (semantic.MeaningCandidate, meaning.RequestContext)
		}{
			{"conflicting-same-actor-anchors", func() (semantic.MeaningCandidate, meaning.RequestContext) {
				candidate := personalFixtureCandidate("conflicting", fixture)
				ctx := fixture.Context
				other := ctx.Mentions[0]
				other.ID = "conflicting-anchor"
				anchor := meaning.DeicticAnchor{Kind: "participant", Value: "addressee"}
				other.Anchor = &anchor
				ctx.Mentions = append(ctx.Mentions, other)
				candidate.Contexts = []semantic.MeaningContext{{Meaning: fixture.Graph, Context: ctx}}
				return candidate, ctx
			}},
			{"stale-first-paired-context", func() (semantic.MeaningCandidate, meaning.RequestContext) {
				candidate := personalFixtureCandidate("stale-first", fixture)
				stale := fixture.Graph
				stale.Nodes = append([]meaning.Node(nil), stale.Nodes...)
				for i := range stale.Nodes {
					if stale.Nodes[i].Kind == meaning.KindEvent {
						stale.Nodes[i].TemporalFrame = meaning.TemporalFramePast
					}
				}
				withoutAnchor := fixture.Context
				withoutAnchor.Mentions = nil
				candidate.Contexts = []semantic.MeaningContext{
					{Meaning: stale, Context: withoutAnchor},
					{Meaning: fixture.Graph, Context: fixture.Context},
				}
				same, err := meaning.StructuralEquivalence(fixture.Graph, stale, meaning.DefaultBudget())
				if err != nil || same {
					t.Fatalf("stale paired graph is not structurally distinct: same=%t error=%v", same, err)
				}
				return candidate, withoutAnchor
			}},
			{"participant-anchor-on-object", func() (semantic.MeaningCandidate, meaning.RequestContext) {
				graph := fixture.Graph
				graph.Nodes = append([]meaning.Node(nil), graph.Nodes...)
				for i := range graph.Nodes {
					if graph.Nodes[i].ID == "r2" {
						graph.Nodes[i].Concept = "concept:person"
					}
				}
				ctx := fixture.Context
				anchor := meaning.DeicticAnchor{Kind: "participant", Value: "speaker"}
				ctx.Mentions = []meaning.Mention{{ID: "object-anchor", NodeID: "r2", Anchor: &anchor}}
				candidate := semantic.MeaningCandidate{ID: "object-anchor", SemanticID: "object-anchor", Meaning: &graph, Status: status.OutcomeOK,
					Evidence: semantic.EvidenceTrace{AlternativeID: "object-anchor", Outcome: status.OutcomeOK},
					Contexts: []semantic.MeaningContext{{Meaning: graph, Context: ctx}}}
				return candidate, ctx
			}},
		} {
			t.Run(string(selectedTarget.lang)+"/"+makeCase.name, func(t *testing.T) {
				candidate, validationContext := makeCase.build()
				if err := meaning.ValidateReferences(*candidate.Meaning, validationContext, meaning.DefaultBudget()); err != nil {
					t.Fatalf("constructed invalid-context diagnostic: %v", err)
				}
				for _, paired := range candidate.Contexts {
					if err := meaning.ValidateReferences(paired.Meaning, paired.Context, meaning.DefaultBudget()); err != nil {
						t.Fatalf("paired context is not valid against its graph: %v", err)
					}
				}
				if makeCase.name == "stale-first-paired-context" {
					same, err := meaning.StructuralEquivalence(*candidate.Meaning, candidate.Contexts[0].Meaning, meaning.DefaultBudget())
					if err != nil || same {
						t.Fatalf("first paired graph is not stale relative to actual Meaning: same=%t error=%v", same, err)
					}
				}
				result, err := realizer.Realize(context.Background(), candidate)
				if err != nil {
					var resultError interface{ Result() status.Result }
					if !errors.As(err, &resultError) || (resultError.Result().Status != status.OutcomeInvalid && resultError.Result().Status != status.OutcomeUnsupported) {
						t.Fatalf("realize diagnostic candidate returned untyped/nonrefusal error: %v", err)
					}
					return
				}
				if result.Outcome != status.OutcomeInvalid && result.Outcome != status.OutcomeUnsupported {
					t.Errorf("invalid participant anchor outcome=%s, want invalid or unsupported", result.Outcome)
				}
				for _, alternative := range result.Alternatives {
					if alternative.Text != "" {
						t.Errorf("invalid/misaligned participant anchor generated surface %q", alternative.Text)
					}
				}
			})
		}
	}
}

func TestPolishImplicitRoleEvidenceRejectsDuplicateAndStaleProofs(t *testing.T) {
	engine := matrix.Runtime(t)
	adapter, ok := engine.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || adapter == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	production, ok := adapter.(*runtimepipe.SourcePipeline)
	if !ok || production == nil {
		t.Fatalf("INFRA: Polish analyzer has unexpected implementation %T", adapter)
	}
	for _, mode := range []string{"duplicate", "stale-derivation"} {
		t.Run(mode, func(t *testing.T) {
			pipeline := *production
			builder := &personalEvidenceMutationBuilder{base: pipeline.EvidenceBuilder, mode: mode}
			pipeline.EvidenceBuilder = builder
			result, err := pipeline.Analyze(context.Background(), "Widzę psa.")
			if !builder.mutated {
				t.Fatal("fixture did not produce a complete implicit-role proof to corrupt")
			}
			if err == nil {
				for _, candidate := range result.Candidates {
					if candidate.Status == status.OutcomeOK && hasPredicate(candidate, "seeing", "") {
						t.Fatalf("malformed implicit proof created an OK seeing Meaning: %+v", candidate.Meaning)
					}
				}
				t.Fatalf("malformed implicit proof was accepted without a controlled error: outcome=%s", result.Outcome)
			}
			var resultError interface{ Result() status.Result }
			if !errors.As(err, &resultError) {
				t.Fatalf("malformed implicit proof returned untyped error %T: %v", err, err)
			}
			failure := resultError.Result()
			if failure.Status != status.OutcomeInvalid {
				t.Errorf("malformed proof error outcome=%s, want invalid: %+v", failure.Status, failure.Diagnostics)
			}
			t.Logf("controlled error outcome=%s diagnostics=%+v", failure.Status, failure.Diagnostics)
		})
	}
}

type personalEvidenceMutationBuilder struct {
	base    runtimepipe.CompositionEvidenceBuilder
	mode    string
	mutated bool
}

func (builder *personalEvidenceMutationBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtimepipe.CompositionEvidence, error) {
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil {
		return evidence, err
	}
	if len(evidence.ImplicitRoles) == 0 {
		return evidence, nil
	}
	builder.mutated = true
	switch builder.mode {
	case "duplicate":
		evidence.ImplicitRoles = append(evidence.ImplicitRoles, evidence.ImplicitRoles[0])
	case "stale-derivation":
		evidence.ImplicitRoles[0].DerivationID = "stale-derivation"
	}
	return evidence, nil
}

func personalFixtureCandidate(id string, fixture interlingua.IdentificationMeaning) semantic.MeaningCandidate {
	graph := fixture.Graph
	return semantic.MeaningCandidate{
		ID: id, SemanticID: id, Meaning: &graph, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: id, Outcome: status.OutcomeOK},
		Contexts: []semantic.MeaningContext{{Meaning: graph, Context: fixture.Context}},
	}
}

func hasPersonalSourceCandidate(candidates []semantic.MeaningCandidate, input, verb, predicate, multiplicity, anchor string) bool {
	start := strings.Index(input, verb)
	for _, candidate := range candidates {
		if candidate.Meaning == nil || candidate.Status != status.OutcomeOK || !hasPredicate(candidate, predicate, "") {
			continue
		}
		subjectID := personalSubjectID(*candidate.Meaning, predicate)
		if subjectID == "" {
			continue
		}
		subject := nodeByID(*candidate.Meaning, subjectID)
		if subject.Kind != meaning.KindReferent || subject.Concept != "concept:person" || multiplicity != "" && string(subject.Multiplicity) != multiplicity {
			continue
		}
		anchored := false
		for _, paired := range candidate.Contexts {
			if len(paired.Context.Referents) != 0 {
				continue
			}
			for _, mention := range paired.Context.Mentions {
				if mention.NodeID == subjectID && mention.ReferentID == "" && mention.EvidenceID == "" && mention.Anchor != nil && mention.Anchor.Kind == "participant" && mention.Anchor.Value == anchor {
					anchored = true
				}
			}
		}
		if !anchored {
			continue
		}
		if verb != "" {
			foundVPProof := false
			for _, record := range candidate.Evidence.Records {
				if record.Reason == "implicit participant role experiencer licensed by construction evidence; no surface subject" &&
					record.CoreConceptID.String() == "concept:"+predicate && record.FamilyID.String() == "construction:pl.transitive_personal" {
					if record.SourceSpan.Start != start || record.SourceSpan.End != start+len(verb) || input[record.SourceSpan.Start:record.SourceSpan.End] != verb {
						continue
					}
					foundVPProof = true
				}
				if strings.HasPrefix(record.SenseID.String(), "sense:pl.ja.") || strings.HasPrefix(record.SenseID.String(), "sense:pl.my.") ||
					strings.HasPrefix(record.SenseID.String(), "sense:pl.ty.") || strings.HasPrefix(record.SenseID.String(), "sense:pl.wy.") {
					return false
				}
			}
			if !foundVPProof {
				continue
			}
		}
		return true
	}
	return false
}

func candidateHasParticipantAnchor(candidate semantic.MeaningCandidate, nodeID meaning.NodeID, expected string) bool {
	for _, paired := range candidate.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.NodeID == nodeID && mention.ReferentID == "" && mention.EvidenceID == "" && mention.Anchor != nil &&
				mention.Anchor.Kind == "participant" && mention.Anchor.Value == expected {
				return true
			}
		}
	}
	return false
}

func personalSubjectMultiplicity(graph meaning.Meaning) meaning.Multiplicity {
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindReferent && node.Concept == "concept:person" {
			return node.Multiplicity
		}
	}
	return ""
}

func personalSubjectID(graph meaning.Meaning, predicate string) meaning.NodeID {
	for _, node := range graph.Nodes {
		if node.Kind != meaning.KindEvent || string(node.Concept) != "concept:"+predicate {
			continue
		}
		role := meaning.Role("experiencer")
		if predicate == "reading" {
			role = "reader"
		}
		for _, participant := range node.Participants {
			if participant.Role == role {
				return participant.Target
			}
		}
	}
	return ""
}

func nodeByID(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func hasPredicate(candidate semantic.MeaningCandidate, predicate string, tense meaning.TemporalFrame) bool {
	if candidate.Meaning == nil {
		return false
	}
	for _, node := range candidate.Meaning.Nodes {
		if node.Kind == meaning.KindEvent && string(node.Concept) == "concept:"+predicate && (tense == "" || node.TemporalFrame == tense) {
			return true
		}
	}
	return false
}

func assertSurfaceNumberAssumption(t *testing.T, alternatives []engine.SurfaceAlternative, surface, number string) {
	t.Helper()
	for _, alternative := range alternatives {
		if alternative.Text != surface {
			continue
		}
		for _, assumption := range alternative.Assumptions {
			if assumption.Field == "number" && assumption.Value == number {
				return
			}
		}
	}
	t.Errorf("missing %q with number=%s assumption in %+v", surface, number, alternatives)
}

func assertSurfaceGenderAssumption(t *testing.T, alternatives []engine.SurfaceAlternative, surface, gender string) {
	t.Helper()
	for _, alternative := range alternatives {
		if alternative.Text != surface {
			continue
		}
		for _, assumption := range alternative.Assumptions {
			if assumption.Field == "gender" && assumption.Value == gender {
				return
			}
		}
	}
	t.Errorf("missing %q with gender=%s assumption in %+v", surface, gender, alternatives)
}

package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestIdentificationMeaning(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	fixtures := map[string]interlingua.IdentificationMeaning{}
	for _, id := range []string{
		"identification.classification.this",
		"identification.classification.to",
		"identification.black.this",
		"identification.black.ten",
	} {
		fixture, ok, err := interlingua.Identification(id)
		if err != nil || !ok {
			t.Fatalf("INFRA: load identification fixture %q: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(fixture.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Fatalf("fixture %q does not validate against CORE: %v", id, err)
		}
		if err := meaning.ValidateReferences(fixture.Graph, fixture.Context, meaning.DefaultBudget()); err != nil {
			t.Fatalf("fixture %q has invalid reference context: %v", id, err)
		}
		if len(fixture.Context.Referents) != 0 || len(fixture.Context.Mentions) != 1 ||
			fixture.Context.Mentions[0].ReferentID != "" || fixture.Context.Mentions[0].EvidenceID != "" ||
			fixture.Context.Mentions[0].Anchor == nil {
			t.Errorf("fixture %q must contain only one unresolved anchored mention", id)
		}
		if !hasOnlyPredicationAndReferent(fixture.Graph) {
			t.Errorf("fixture %q contains nodes outside the state/referent contract", id)
		}
		candidate := candidateFor(fixture)
		if matches, err := matrix.IdentificationMatches(fixture, candidate); err != nil || !matches {
			t.Errorf("fixture %q should match itself: matches=%t error=%v", id, matches, err)
		}
		renamedGraph := renameMeaning(fixture.Graph, map[meaning.NodeID]meaning.NodeID{"s": "predicate", "r": "referent"})
		renamedContext := cloneRequestContext(fixture.Context)
		renamedContext.ID = "opaque-context-changed"
		renamedContext.Mentions[0].ID = "opaque-mention-changed"
		renamedContext.Mentions[0].NodeID = "referent"
		renamed := semantic.MeaningCandidate{
			ID: id, SemanticID: id, Meaning: &renamedGraph,
			Contexts: []semantic.MeaningContext{{Meaning: renamedGraph, Context: renamedContext}},
		}
		if matches, err := matrix.IdentificationMatches(fixture, renamed); err != nil || !matches {
			t.Errorf("renamed/reordered fixture %q with opaque context IDs changed should match: matches=%t error=%v", id, matches, err)
		}
		fixtures[id] = fixture
	}

	classification := fixtures["identification.classification.this"]
	black := fixtures["identification.black.this"]
	assertIdentificationRejects(t, "classification versus black", classification, candidateFor(black))

	dogGraph := cloneMeaning(classification.Graph)
	dog, err := ids.NewConceptID("dog")
	if err != nil {
		t.Fatalf("create dog concept ID: %v", err)
	}
	for index := range dogGraph.Nodes {
		if dogGraph.Nodes[index].Kind == meaning.KindReferent {
			dogGraph.Nodes[index].Concept = meaning.ConceptRefFromID(dog)
		}
	}
	dogCandidate := candidateFor(classification)
	dogCandidate.Meaning = &dogGraph
	dogCandidate.Contexts[0].Meaning = dogGraph
	assertIdentificationRejects(t, "cat changed to dog", classification, dogCandidate)

	blackPredicateMissing := cloneMeaning(black.Graph)
	blackPredicateMissing.Roots = []meaning.NodeID{"r"}
	blackPredicateMissing.Nodes = []meaning.Node{blackPredicateMissing.Nodes[1]}
	missingPredicate := candidateFor(black)
	missingPredicate.Meaning = &blackPredicateMissing
	missingPredicate.Contexts[0].Meaning = blackPredicateMissing
	assertIdentificationRejects(t, "black predicate omitted", black, missingPredicate)

	bearerMissing := cloneMeaning(black.Graph)
	for index := range bearerMissing.Nodes {
		if bearerMissing.Nodes[index].Kind == meaning.KindState {
			bearerMissing.Nodes[index].Participants = nil
		}
	}
	missingBearer := candidateFor(black)
	missingBearer.Meaning = &bearerMissing
	missingBearer.Contexts[0].Meaning = bearerMissing
	assertIdentificationRejects(t, "black bearer edge omitted", black, missingBearer)

	wrongMentionTarget := candidateFor(black)
	wrongMentionTarget.Contexts[0].Context.Mentions[0].NodeID = "s"
	assertIdentificationRejects(t, "mention targets state", black, wrongMentionTarget)

	catID, err := ids.NewConceptID("cat")
	if err != nil {
		t.Fatalf("create cat concept ID: %v", err)
	}
	extraReferentGraph := cloneMeaning(black.Graph)
	second, err := meaning.NewReferent("r2", catID)
	if err != nil {
		t.Fatalf("create second cat referent: %v", err)
	}
	for index := range extraReferentGraph.Nodes {
		if extraReferentGraph.Nodes[index].Kind == meaning.KindState {
			extraReferentGraph.Nodes[index].Participants[0].Target = "r2"
		}
	}
	extraReferentGraph.Nodes = append(extraReferentGraph.Nodes, second)
	discourse, err := meaning.NewDiscourse("d", meaning.ActStatement, []meaning.NodeID{"s", "r"})
	if err != nil {
		t.Fatalf("create discourse root for extra referent: %v", err)
	}
	extraReferentGraph.Nodes = append(extraReferentGraph.Nodes, discourse)
	extraReferentGraph, err = meaning.Build([]meaning.NodeID{"d"}, extraReferentGraph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build extra cat referent graph: %v", err)
	}
	extraReferent := candidateFor(black)
	extraReferent.Meaning = &extraReferentGraph
	extraContext := cloneRequestContext(black.Context)
	extraContext.Mentions[0].NodeID = "r"
	extraReferent.Contexts[0] = semantic.MeaningContext{Meaning: extraReferentGraph, Context: extraContext}
	assertIdentificationRejects(t, "extra distinct cat referent with state edge pointing elsewhere", black, extraReferent)

	wrongAnchor := candidateFor(classification)
	wrongAnchor.Contexts[0].Context.Mentions[0].Anchor.Value = "demonstrative"
	assertIdentificationRejects(t, "anchor changed from proximal to demonstrative", classification, wrongAnchor)

	fabricatedIdentity := candidateFor(classification)
	fabricatedContext := cloneRequestContext(classification.Context)
	fabricatedContext.Referents = []meaning.Referent{{ID: "individual-1", Concept: meaning.ConceptRefFromID(catID)}}
	fabricatedContext.Mentions[0].ReferentID = "individual-1"
	fabricatedContext.Mentions[0].EvidenceID = "evidence-1"
	fabricatedIdentity.Contexts[0].Context = fabricatedContext
	assertIdentificationRejects(t, "fabricated explicit identity", classification, fabricatedIdentity)
}

func assertIdentificationRejects(t *testing.T, name string, want interlingua.IdentificationMeaning, got semantic.MeaningCandidate) {
	t.Helper()
	matches, err := matrix.IdentificationMatches(want, got)
	if matches {
		t.Errorf("%s unexpectedly matched identification contract (error=%v)", name, err)
	}
}

func candidateFor(fixture interlingua.IdentificationMeaning) semantic.MeaningCandidate {
	graph := cloneMeaning(fixture.Graph)
	return semantic.MeaningCandidate{
		ID: "identification-fixture", SemanticID: "identification-fixture", Meaning: &graph,
		Contexts: []semantic.MeaningContext{{Meaning: graph, Context: cloneRequestContext(fixture.Context)}},
	}
}

func cloneMeaning(source meaning.Meaning) meaning.Meaning {
	graph, err := meaning.Build(source.Roots, source.Nodes, meaning.DefaultBudget())
	if err != nil {
		panic(err)
	}
	return graph
}

func cloneRequestContext(source meaning.RequestContext) meaning.RequestContext {
	result := source
	result.Referents = append([]meaning.Referent(nil), source.Referents...)
	result.Mentions = append([]meaning.Mention(nil), source.Mentions...)
	for index := range result.Mentions {
		if result.Mentions[index].Anchor != nil {
			anchor := *result.Mentions[index].Anchor
			result.Mentions[index].Anchor = &anchor
		}
	}
	return result
}

func hasOnlyPredicationAndReferent(graph meaning.Meaning) bool {
	stateCount, referentCount := 0, 0
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindState:
			stateCount++
		case meaning.KindReferent:
			referentCount++
		default:
			return false
		}
	}
	return stateCount == 1 && referentCount == 1
}

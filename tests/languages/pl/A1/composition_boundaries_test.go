package a1pl_test

import (
	"context"
	"testing"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestCompositionBoundaries(t *testing.T) {
	target := targetPL(t)
	base, ok, err := interlingua.Composition("composition.black-cat-sees-dog")
	if err != nil || !ok {
		t.Fatalf("load independent composition fixture: ok=%t error=%v", ok, err)
	}
	positive := meaningCandidatePL(base, "composition-positive", meaning.RequestContext{})
	validateCandidatePL(t, positive)
	positiveResult, positiveErr := target.Realize(context.Background(), positive)
	logPL(t, "positive independent composition I", positiveResult, positiveErr)
	if positiveErr != nil || !sameStrings(textSet(positiveResult.Alternatives), []string{"Czarny kot widzi psa."}) {
		t.Fatalf("positive independent I did not realize exactly: outcome=%s texts=%v error=%v", positiveResult.Outcome, textSet(positiveResult.Alternatives), positiveErr)
	}

	// A graph without a property root is still a valid, independently supported
	// simple seeing Meaning; absence of q alone must not be rejected.
	simple, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("load simple seeing fixture: ok=%t error=%v", ok, err)
	}
	simpleCandidate := meaningCandidatePL(simple, "simple-seeing-no-property", meaning.RequestContext{})
	validateCandidatePL(t, simpleCandidate)
	simpleResult, simpleErr := target.Realize(context.Background(), simpleCandidate)
	logPL(t, "simple seeing without q", simpleResult, simpleErr)
	if simpleErr != nil || len(textSet(simpleResult.Alternatives)) == 0 {
		t.Errorf("simple seeing without q was refused: outcome=%s diagnostics=%+v error=%v", simpleResult.Outcome, simpleResult.Diagnostics, simpleErr)
	}

	catCat, ok, err := interlingua.Composition("composition.black-cat-sees-cat")
	if err != nil || !ok {
		t.Fatalf("load cat/cat composition fixture: ok=%t error=%v", ok, err)
	}
	cases := []struct {
		name string
		edit func(meaning.Meaning) (meaning.Meaning, meaning.RequestContext)
	}{
		{"bearer is object", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return retargetBlackPL(t, g, "r2"), meaning.RequestContext{}
		}},
		{"bearer is second cat despite matching concept", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return retargetBlackPL(t, catCat, "r2"), meaning.RequestContext{}
		}},
		{"event time absent", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFramePL(&g, "e", "")
			return buildBoundaryPL(t, g), meaning.RequestContext{}
		}},
		{"event past", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFramePL(&g, "e", meaning.TemporalFramePast)
			return buildBoundaryPL(t, g), meaning.RequestContext{}
		}},
		{"subject many", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeMultiplicityPL(&g, "r1", meaning.MultiplicityMany)
			return buildBoundaryPL(t, g), meaning.RequestContext{}
		}},
		{"object multiplicity absent", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeMultiplicityPL(&g, "r2", "")
			return buildBoundaryPL(t, g), meaning.RequestContext{}
		}},
		{"property is timed", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFramePL(&g, "q", meaning.TemporalFramePresent)
			return buildBoundaryPL(t, g), meaning.RequestContext{}
		}},
		{"additional property root", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return addSecondBlackPL(t, g), meaning.RequestContext{}
		}},
		{"negation scopes event alongside property", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return scopeEventPL(t, g), meaning.RequestContext{}
		}},
		{"anchored subject context", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return g, compositionContextPL("r1", true)
		}},
		{"context graph has a different bearer", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return g, compositionContextPL("r1", false)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, requestContext := tc.edit(cloneMeaningPL(t, base))
			candidate := meaningCandidatePL(graph, "composition-boundary-"+tc.name, requestContext)
			if tc.name == "context graph has a different bearer" {
				contextGraph := retargetBlackPL(t, cloneMeaningPL(t, graph), "r2")
				candidate.Contexts = []semantic.MeaningContext{{Meaning: contextGraph, Context: requestContext}}
			}
			validateCandidatePL(t, candidate)
			result, realizeErr := target.Realize(context.Background(), candidate)
			logPL(t, tc.name, result, realizeErr)
			assertNoTextPL(t, result, realizeErr)
		})
	}
}

func buildBoundaryPL(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	built, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build valid boundary Meaning: %v", err)
	}
	return built
}

func setNodeFramePL(graph *meaning.Meaning, id meaning.NodeID, value meaning.TemporalFrame) {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == id {
			graph.Nodes[index].TemporalFrame = value
		}
	}
}

func setNodeMultiplicityPL(graph *meaning.Meaning, id meaning.NodeID, value meaning.Multiplicity) {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == id {
			graph.Nodes[index].Multiplicity = value
		}
	}
}

func retargetBlackPL(t *testing.T, graph meaning.Meaning, referent meaning.NodeID) meaning.Meaning {
	t.Helper()
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == "q" {
			graph.Nodes[index].Participants[0].Target = referent
		}
	}
	return buildBoundaryPL(t, graph)
}

func addSecondBlackPL(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	black, err := ids.NewConceptID("black")
	if err != nil {
		t.Fatal(err)
	}
	state, err := meaning.NewState("q2", black, []meaning.Participant{{Role: "bearer", Target: "r1"}})
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes = append(graph.Nodes, state)
	graph.Roots = append(graph.Roots, "q2")
	return buildBoundaryPL(t, graph)
}

func scopeEventPL(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	negation, err := meaning.NewNegation("neg", "e")
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes = append(graph.Nodes, negation)
	graph.Roots = []meaning.NodeID{"neg", "q"}
	return buildBoundaryPL(t, graph)
}

func compositionContextPL(subject meaning.NodeID, anchored bool) meaning.RequestContext {
	context := meaning.RequestContext{ID: "composition-boundary-context", Mentions: []meaning.Mention{
		{ID: "subject", NodeID: subject}, {ID: "object", NodeID: "r2"},
	}}
	if anchored {
		context.Mentions[0].Anchor = &meaning.DeicticAnchor{Kind: "deixis", Value: "proximal"}
	}
	return context
}

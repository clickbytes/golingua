package interlingua_test

import (
	"errors"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
)

func TestMeaningAttributesReason(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}

	for _, multiplicity := range []meaning.Multiplicity{"", meaning.MultiplicityOne} {
		t.Run("unresolved time and multiplicity="+string(multiplicity)+" remain plannable", func(t *testing.T) {
			input := meaningAttributeQuestion(t, "", multiplicity, multiplicity, "")
			planned, err := reason.Plan(input, core, meaning.DefaultBudget())
			if err != nil {
				t.Fatalf("baseline question should be plannable: %v", err)
			}
			if planned.Args != [2]meaning.ReferentID{"teacher-1", "dog-1"} {
				t.Errorf("baseline arguments = %v, want teacher-1/dog-1", planned.Args)
			}
		})
	}

	for _, frame := range []meaning.TemporalFrame{meaning.TemporalFramePast, meaning.TemporalFramePresent} {
		frame := frame
		t.Run("known "+string(frame)+" frame is unsupported even with At", func(t *testing.T) {
			input := meaningAttributeQuestion(t, frame, "", "", "2026-10-24")
			_, err := reason.Plan(input, core, meaning.DefaultBudget())
			assertMeaningAttributeRefusal(t, err, "UNSUPPORTED_QUESTION_TEMPORAL_FRAME", "meaning.nodes.e.temporal_frame")
		})
	}

	for _, role := range []string{"experiencer", "stimulus"} {
		role := role
		t.Run("many "+role+" referent is unsupported", func(t *testing.T) {
			teacherMultiplicity, dogMultiplicity := meaning.MultiplicityOne, meaning.MultiplicityOne
			targetID := "t"
			if role == "experiencer" {
				teacherMultiplicity = meaning.MultiplicityMany
			} else {
				dogMultiplicity = meaning.MultiplicityMany
				targetID = "d"
			}
			input := meaningAttributeQuestion(t, "", teacherMultiplicity, dogMultiplicity, "")
			_, err := reason.Plan(input, core, meaning.DefaultBudget())
			assertMeaningAttributeRefusal(t, err, "UNSUPPORTED_QUESTION_MULTIPLICITY", "meaning.nodes."+targetID+".multiplicity")
		})
	}
}

func meaningAttributeQuestion(t *testing.T, frame meaning.TemporalFrame, teacherMultiplicity, dogMultiplicity meaning.Multiplicity, at string) reason.QuestionInput {
	t.Helper()
	seeing, err := ids.NewConceptID("seeing")
	if err != nil {
		t.Fatalf("build seeing concept ID: %v", err)
	}
	teacher, err := ids.NewConceptID("teacher")
	if err != nil {
		t.Fatalf("build teacher concept ID: %v", err)
	}
	dog, err := ids.NewConceptID("dog")
	if err != nil {
		t.Fatalf("build dog concept ID: %v", err)
	}
	event, err := meaning.NewEvent("e", seeing, []meaning.Participant{
		{Role: "experiencer", Target: "t"}, {Role: "stimulus", Target: "d"},
	})
	if err != nil {
		t.Fatalf("build seeing event: %v", err)
	}
	event.TemporalFrame = frame
	teacherNode, err := meaning.NewReferent("t", teacher)
	if err != nil {
		t.Fatalf("build teacher referent: %v", err)
	}
	teacherNode.Multiplicity = teacherMultiplicity
	dogNode, err := meaning.NewReferent("d", dog)
	if err != nil {
		t.Fatalf("build dog referent: %v", err)
	}
	dogNode.Multiplicity = dogMultiplicity
	question, err := meaning.NewQuestion("q", "e")
	if err != nil {
		t.Fatalf("build question scope: %v", err)
	}
	discourse, err := meaning.NewDiscourse("root", meaning.ActQuestion, []meaning.NodeID{"q"})
	if err != nil {
		t.Fatalf("build question discourse: %v", err)
	}
	graph, err := meaning.Build([]meaning.NodeID{"root"}, []meaning.Node{event, teacherNode, dogNode, question, discourse}, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build question Meaning: %v", err)
	}
	return reason.QuestionInput{
		Meaning: graph,
		Context: meaning.RequestContext{
			ID: "ctx",
			Referents: []meaning.Referent{
				{ID: "teacher-1", Concept: "concept:teacher"},
				{ID: "dog-1", Concept: "concept:dog"},
			},
			Mentions: []meaning.Mention{
				{ID: "m1", NodeID: "t", ReferentID: "teacher-1", EvidenceID: "ev1"},
				{ID: "m2", NodeID: "d", ReferentID: "dog-1", EvidenceID: "ev2"},
			},
		},
		Kind: reason.QuestionYesNo,
		Mapping: reason.RelationMapping{
			Predicate:      "relation:seeing",
			RoleToArgument: map[meaning.Role]int{"experiencer": 0, "stimulus": 1},
		},
		At: at,
	}
}

func assertMeaningAttributeRefusal(t *testing.T, err error, code, path string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s refusal", code)
	}
	var resultError interface{ Result() status.Result }
	if !errors.As(err, &resultError) {
		t.Fatalf("expected typed status error for %s, got %T: %v", code, err, err)
	}
	result := resultError.Result()
	if result.Status != status.OutcomeUnsupported {
		t.Fatalf("status = %q, want unsupported; error=%v", result.Status, err)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code && diagnostic.Path == path {
			return
		}
	}
	t.Fatalf("missing diagnostic %s at %s: %+v", code, path, result.Diagnostics)
}

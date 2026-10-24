// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

// Seeing returns the language-independent teacher/dog seeing contract.
func Seeing(id string) (meaning.Meaning, bool, error) {
	var experiencer, stimulus meaning.NodeID
	switch id {
	case "seeing.teacher-dog":
		experiencer, stimulus = "t", "d"
	case "seeing.dog-teacher":
		experiencer, stimulus = "d", "t"
	default:
		return meaning.Meaning{}, false, nil
	}

	seeing, err := ids.NewConceptID("seeing")
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	teacher, err := ids.NewConceptID("teacher")
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	dog, err := ids.NewConceptID("dog")
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	event, err := meaning.NewEvent("e", seeing, []meaning.Participant{
		{Role: "experiencer", Target: experiencer},
		{Role: "stimulus", Target: stimulus},
	})
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	event.TemporalFrame = meaning.TemporalFramePresent
	teacherNode, err := meaning.NewReferent("t", teacher)
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	teacherNode.Multiplicity = meaning.MultiplicityOne
	dogNode, err := meaning.NewReferent("d", dog)
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	dogNode.Multiplicity = meaning.MultiplicityOne
	graph, err := meaning.Build([]meaning.NodeID{"e"}, []meaning.Node{event, teacherNode, dogNode}, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	return graph, true, nil
}

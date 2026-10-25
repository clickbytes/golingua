// Package interlingua stores independent, hand-authored A1.2 contrast meanings.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type contrastSpec struct {
	event, role1, concept1, role2, concept2 string
	multiplicity1, multiplicity2            meaning.Multiplicity
	negative, blackObject                   bool
}

var contrastSpecs = map[string]contrastSpec{
	"m5.seeing.teachers-dog":      {"seeing", "experiencer", "teacher", "stimulus", "dog", meaning.MultiplicityMany, meaning.MultiplicityOne, false, false},
	"m5.neg.hearing.teachers-dog": {"hearing", "experiencer", "teacher", "stimulus", "dog", meaning.MultiplicityMany, meaning.MultiplicityOne, true, false},
	"m5.black.dog-sees-black-cat": {"seeing", "experiencer", "dog", "stimulus", "cat", meaning.MultiplicityOne, meaning.MultiplicityOne, false, true},
	"m5.neg.seeing.cat-dog":       {"seeing", "experiencer", "cat", "stimulus", "dog", meaning.MultiplicityOne, meaning.MultiplicityOne, true, false},
	"m5.neg.seeing.teacher-book":  {"seeing", "experiencer", "teacher", "stimulus", "book", meaning.MultiplicityOne, meaning.MultiplicityOne, true, false},
	"m5.hearing.dogs-cat":         {"hearing", "experiencer", "dog", "stimulus", "cat", meaning.MultiplicityMany, meaning.MultiplicityOne, false, false},
	"m5.black.teacher-sees-cats":  {"seeing", "experiencer", "teacher", "stimulus", "cat", meaning.MultiplicityOne, meaning.MultiplicityMany, false, true},
}

// Contrast builds a Meaning graph from the authored A1.2 event, role, count,
// polarity, and property specification. No sentence parser is involved.
func Contrast(id string) (meaning.Meaning, bool, error) {
	spec, ok := contrastSpecs[id]
	if !ok {
		return meaning.Meaning{}, false, nil
	}
	newConcept := func(name string) (ids.ConceptID, error) {
		return ids.NewConceptID(name)
	}
	eventConcept, err := newConcept(spec.event)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("event concept: %w", err)
	}
	firstConcept, err := newConcept(spec.concept1)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("first participant concept: %w", err)
	}
	secondConcept, err := newConcept(spec.concept2)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("second participant concept: %w", err)
	}
	event, err := meaning.NewEvent("e", eventConcept, []meaning.Participant{
		{Role: meaning.Role(spec.role1), Target: "r1"},
		{Role: meaning.Role(spec.role2), Target: "r2"},
	})
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("event: %w", err)
	}
	event.TemporalFrame = meaning.TemporalFramePresent
	first, err := meaning.NewReferent("r1", firstConcept)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("first participant: %w", err)
	}
	first.Multiplicity = spec.multiplicity1
	second, err := meaning.NewReferent("r2", secondConcept)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("second participant: %w", err)
	}
	second.Multiplicity = spec.multiplicity2

	nodes := []meaning.Node{event, first, second}
	roots := []meaning.NodeID{"e"}
	if spec.negative {
		scope, err := meaning.NewNegation("neg", "e")
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("negation scope: %w", err)
		}
		nodes = append(nodes, scope)
		roots = []meaning.NodeID{"neg"}
	}
	if spec.blackObject {
		black, err := newConcept("black")
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("black property concept: %w", err)
		}
		state, err := meaning.NewState("q", black, []meaning.Participant{{Role: "bearer", Target: "r2"}})
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("black property: %w", err)
		}
		nodes = append(nodes, state)
		roots = append(roots, "q")
	}
	graph, err := meaning.Build(roots, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build contrast Meaning %q: %w", id, err)
	}
	return graph, true, nil
}

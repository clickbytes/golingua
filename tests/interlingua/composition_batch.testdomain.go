// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type batchEvent struct {
	id, concept  string
	participants []meaning.Participant
	negative     bool
}

type batchSpec struct {
	events       []batchEvent
	referents    map[meaning.NodeID]string
	blackBearers []meaning.NodeID
}

// CompositionBatch returns a hand-authored graph for one bounded M6 sentence.
func CompositionBatch(id string) (meaning.Meaning, bool, error) {
	if id == "composition.batch.06" {
		return Composition("composition.black-dog-sees-cat")
	}
	if id == "composition.batch.07" {
		return Composition("composition.black-bird-sees-dog")
	}
	if id == "composition.batch.08" {
		return Composition("composition.black-cat-sees-cat")
	}
	if id == "composition.batch.09" {
		return Seeing("seeing.dog-teacher")
	}

	specs := map[string]batchSpec{
		"composition.batch.01": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}, {"e2", "hearing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r3"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "teacher", "r2": "dog", "r3": "cat"},
		},
		"composition.batch.02": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}, {"e2", "hearing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r3"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "cat", "r2": "dog", "r3": "bird"},
		},
		"composition.batch.03": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}, {"e2", "hearing", []meaning.Participant{{Role: "experiencer", Target: "r3"}, {Role: "stimulus", Target: "r4"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "teacher", "r2": "cat", "r3": "dog", "r4": "bird"},
		},
		"composition.batch.04": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "cat", "r2": "dog"}, blackBearers: []meaning.NodeID{"r1", "r2"},
		},
		"composition.batch.05": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "cat", "r2": "dog"}, blackBearers: []meaning.NodeID{"r2"},
		},
		"composition.batch.10": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, true}},
			referents: map[meaning.NodeID]string{"r1": "cat", "r2": "dog"}, blackBearers: []meaning.NodeID{"r2"},
		},
		"composition.batch.11": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, true}},
			referents: map[meaning.NodeID]string{"r1": "cat", "r2": "dog"}, blackBearers: []meaning.NodeID{"r1"},
		},
		"composition.batch.12": {
			events:    []batchEvent{{"e1", "seeing", []meaning.Participant{{Role: "experiencer", Target: "r1"}, {Role: "stimulus", Target: "r2"}}, false}, {"e2", "looking", []meaning.Participant{{Role: "agent", Target: "r1"}, {Role: "target", Target: "r3"}}, false}},
			referents: map[meaning.NodeID]string{"r1": "teacher", "r2": "dog", "r3": "cat"},
		},
	}
	spec, ok := specs[id]
	if !ok {
		return meaning.Meaning{}, false, nil
	}

	var nodes []meaning.Node
	var roots []meaning.NodeID
	for _, eventSpec := range spec.events {
		conceptID, err := ids.NewConceptID(eventSpec.concept)
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("event concept %s: %w", eventSpec.concept, err)
		}
		event, err := meaning.NewEvent(meaning.NodeID(eventSpec.id), conceptID, eventSpec.participants)
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("event %s: %w", eventSpec.id, err)
		}
		event.TemporalFrame = meaning.TemporalFramePresent
		if eventSpec.negative {
			scope, scopeErr := meaning.NewNegation(meaning.NodeID("n"+eventSpec.id[1:]), event.ID)
			if scopeErr != nil {
				return meaning.Meaning{}, false, fmt.Errorf("negation for %s: %w", eventSpec.id, scopeErr)
			}
			nodes = append(nodes, scope)
			roots = append(roots, scope.ID)
		} else {
			roots = append(roots, event.ID)
		}
		nodes = append(nodes, event)
	}
	blackID, err := ids.NewConceptID("black")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("black concept: %w", err)
	}
	for i, bearer := range spec.blackBearers {
		state, stateErr := meaning.NewState(meaning.NodeID(fmt.Sprintf("q%d", i+1)), blackID, []meaning.Participant{{Role: "bearer", Target: bearer}})
		if stateErr != nil {
			return meaning.Meaning{}, false, fmt.Errorf("black state: %w", stateErr)
		}
		nodes = append(nodes, state)
		roots = append(roots, state.ID)
	}
	for refID, name := range spec.referents {
		conceptID, conceptErr := ids.NewConceptID(name)
		if conceptErr != nil {
			return meaning.Meaning{}, false, fmt.Errorf("referent concept %s: %w", name, conceptErr)
		}
		referent, refErr := meaning.NewReferent(refID, conceptID)
		if refErr != nil {
			return meaning.Meaning{}, false, fmt.Errorf("referent %s: %w", refID, refErr)
		}
		referent.Multiplicity = meaning.MultiplicityOne
		nodes = append(nodes, referent)
	}
	graph, err := meaning.Build(roots, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build batch fixture %q: %w", id, err)
	}
	return graph, true, nil
}

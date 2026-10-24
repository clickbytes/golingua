// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

// Composition returns one authored black-subject and seeing-event Meaning.
// The two referent nodes remain separate even when they share a ConceptID.
func Composition(id string) (meaning.Meaning, bool, error) {
	var subjectName, objectName string
	switch id {
	case "composition.black-cat-sees-dog":
		subjectName, objectName = "cat", "dog"
	case "composition.black-dog-sees-cat":
		subjectName, objectName = "dog", "cat"
	case "composition.black-bird-sees-dog":
		subjectName, objectName = "bird", "dog"
	case "composition.black-cat-sees-cat":
		subjectName, objectName = "cat", "cat"
	default:
		return meaning.Meaning{}, false, nil
	}

	seeingID, err := ids.NewConceptID("seeing")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("seeing concept: %w", err)
	}
	blackID, err := ids.NewConceptID("black")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("black concept: %w", err)
	}
	subjectID, err := ids.NewConceptID(subjectName)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("subject concept %q: %w", subjectName, err)
	}
	objectID, err := ids.NewConceptID(objectName)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("object concept %q: %w", objectName, err)
	}

	event, err := meaning.NewEvent("e", seeingID, []meaning.Participant{
		{Role: "experiencer", Target: "r1"},
		{Role: "stimulus", Target: "r2"},
	})
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("seeing event: %w", err)
	}
	event.TemporalFrame = meaning.TemporalFramePresent
	state, err := meaning.NewState("q", blackID, []meaning.Participant{{Role: "bearer", Target: "r1"}})
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("black property state: %w", err)
	}
	subject, err := meaning.NewReferent("r1", subjectID)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("subject referent: %w", err)
	}
	subject.Multiplicity = meaning.MultiplicityOne
	object, err := meaning.NewReferent("r2", objectID)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("object referent: %w", err)
	}
	object.Multiplicity = meaning.MultiplicityOne

	graph, err := meaning.Build([]meaning.NodeID{"e", "q"}, []meaning.Node{event, state, subject, object}, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build composition Meaning %q: %w", id, err)
	}
	return graph, true, nil
}

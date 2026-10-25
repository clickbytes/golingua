// Package interlingua stores authored corpus-event Meaning fixtures.
package interlingua

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type corpusEventSpec struct {
	event, role1, concept1, role2, concept2 string
	past                                    bool
}

var corpusEventSpecs = map[string]corpusEventSpec{
	"S003":           {event: "seeing", role1: "experiencer", concept1: "teacher", role2: "stimulus", concept2: "book", past: false},
	"S010":           {event: "seeing", role1: "experiencer", concept1: "doctor", role2: "stimulus", concept2: "car", past: false},
	"S016":           {event: "seeing", role1: "experiencer", concept1: "father", role2: "stimulus", concept2: "dog", past: false},
	"S023":           {event: "seeing", role1: "experiencer", concept1: "child", role2: "stimulus", concept2: "book", past: false},
	"S023.offspring": {event: "seeing", role1: "experiencer", concept1: "child_of_person", role2: "stimulus", concept2: "book", past: false},
	"S028":           {event: "hearing", role1: "experiencer", concept1: "teacher", role2: "stimulus", concept2: "bird", past: false},
	"S034":           {event: "hearing", role1: "experiencer", concept1: "doctor", role2: "stimulus", concept2: "mother", past: false},
	"S040":           {event: "hearing", role1: "experiencer", concept1: "child", role2: "stimulus", concept2: "father", past: false},
	"S040.offspring": {event: "hearing", role1: "experiencer", concept1: "child_of_person", role2: "stimulus", concept2: "father", past: false},
	"S043":           {event: "hearing", role1: "experiencer", concept1: "neighbor", role2: "stimulus", concept2: "bird", past: false},
	"S051":           {event: "looking", role1: "agent", concept1: "teacher", role2: "target", concept2: "book", past: false},
	"S055":           {event: "looking", role1: "agent", concept1: "teacher", role2: "target", concept2: "tree", past: false},
	"S063":           {event: "looking", role1: "agent", concept1: "mother", role2: "target", concept2: "dog", past: false},
	"S074":           {event: "looking", role1: "agent", concept1: "child", role2: "target", concept2: "car", past: false},
	"S074.offspring": {event: "looking", role1: "agent", concept1: "child_of_person", role2: "target", concept2: "car", past: false},
	"S076":           {event: "receiving", role1: "recipient", concept1: "teacher", role2: "theme", concept2: "book", past: true},
	"S081":           {event: "receiving", role1: "recipient", concept1: "doctor", role2: "theme", concept2: "book", past: true},
	"S086":           {event: "receiving", role1: "recipient", concept1: "mother", role2: "theme", concept2: "book", past: true},
	"S099":           {event: "receiving", role1: "recipient", concept1: "child", role2: "theme", concept2: "message", past: true},
	"S099.offspring": {event: "receiving", role1: "recipient", concept1: "child_of_person", role2: "theme", concept2: "message", past: true},
}

// CorpusEvent builds the independently authored event graph for a selected
// sentence from the draft bilingual corpus. It does not parse sentence text.
func CorpusEvent(id string) (meaning.Meaning, bool, error) {
	spec, ok := corpusEventSpecs[id]
	if !ok {
		return meaning.Meaning{}, false, nil
	}
	concept := func(value string) (ids.ConceptID, error) {
		return ids.NewConceptID(strings.TrimPrefix(value, "concept:"))
	}
	eventConcept, err := concept(spec.event)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("event concept: %w", err)
	}
	firstConcept, err := concept(spec.concept1)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("first participant concept: %w", err)
	}
	secondConcept, err := concept(spec.concept2)
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
	if spec.past {
		event.TemporalFrame = meaning.TemporalFramePast
	} else {
		event.TemporalFrame = meaning.TemporalFramePresent
	}
	first, err := meaning.NewReferent("r1", firstConcept)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("first participant: %w", err)
	}
	first.Multiplicity = meaning.MultiplicityOne
	second, err := meaning.NewReferent("r2", secondConcept)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("second participant: %w", err)
	}
	second.Multiplicity = meaning.MultiplicityOne
	graph, err := meaning.Build([]meaning.NodeID{"e"}, []meaning.Node{event, first, second}, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build corpus Meaning %q: %w", id, err)
	}
	return graph, true, nil
}

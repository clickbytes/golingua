// Package interlingua stores authored, independent A1 action event fixtures.
package interlingua

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type actionParticipantSpec struct{ role, concept, multiplicity string }
type actionExpansionSpec struct {
	id, event, temporal string
	participants        []actionParticipantSpec
}

var actionExpansionSpecs = []actionExpansionSpec{
	{id: "a1.4.actions.001", event: "eating", temporal: "present", participants: []actionParticipantSpec{{"eater", "dog", "one"}, {"food", "bread", ""}}},
	{id: "a1.4.actions.002", event: "eating", temporal: "present", participants: []actionParticipantSpec{{"eater", "mother", "one"}, {"food", "egg_food", "one"}}},
	{id: "a1.4.actions.003", event: "eating", temporal: "present", participants: []actionParticipantSpec{{"eater", "teacher", "one"}, {"food", "apple_fruit", "one"}}},
	{id: "a1.4.actions.004", event: "eating", temporal: "present", participants: []actionParticipantSpec{{"eater", "father", "one"}, {"food", "soup_food", ""}}},
	{id: "a1.4.actions.005", event: "drinking", temporal: "present", participants: []actionParticipantSpec{{"drinker", "dog", "one"}, {"liquid", "water", ""}}},
	{id: "a1.4.actions.006", event: "drinking", temporal: "present", participants: []actionParticipantSpec{{"drinker", "cat", "one"}, {"liquid", "milk_food", ""}}},
	{id: "a1.4.actions.007", event: "drinking", temporal: "present", participants: []actionParticipantSpec{{"drinker", "teacher", "one"}, {"liquid", "tea_beverage", ""}}},
	{id: "a1.4.actions.008", event: "drinking", temporal: "present", participants: []actionParticipantSpec{{"drinker", "doctor", "one"}, {"liquid", "coffee_beverage", ""}}},
	{id: "a1.4.actions.009", event: "sleeping", temporal: "present", participants: []actionParticipantSpec{{"sleeper", "teacher", "one"}}},
	{id: "a1.4.actions.010", event: "sleeping", temporal: "present", participants: []actionParticipantSpec{{"sleeper", "dog", "one"}}},
	{id: "a1.4.actions.011", event: "breathing", temporal: "present", participants: []actionParticipantSpec{{"organism", "cat", "one"}}},
	{id: "a1.4.actions.012", event: "sneezing", temporal: "present", participants: []actionParticipantSpec{{"organism", "mother", "one"}}},
	{id: "a1.4.actions.013", event: "running", temporal: "present", participants: []actionParticipantSpec{{"runner", "dog", "one"}}},
	{id: "a1.4.actions.014", event: "running", temporal: "present", participants: []actionParticipantSpec{{"runner", "cat", "one"}}},
	{id: "a1.4.actions.015", event: "coughing", temporal: "present", participants: []actionParticipantSpec{{"organism", "father", "one"}}},
	{id: "a1.4.actions.016", event: "waking_up", temporal: "present", participants: []actionParticipantSpec{{"person", "doctor", "one"}}},
	{id: "a1.4.actions.017", event: "walking", temporal: "present", participants: []actionParticipantSpec{{"walker", "teacher", "one"}}},
	{id: "a1.4.actions.018", event: "walking", temporal: "present", participants: []actionParticipantSpec{{"walker", "mother", "one"}}},
	{id: "a1.4.actions.019", event: "falling_asleep", temporal: "present", participants: []actionParticipantSpec{{"sleeper", "dog", "one"}}},
	{id: "a1.4.actions.020", event: "walking", temporal: "present", participants: []actionParticipantSpec{{"walker", "horse", "one"}}},
	{id: "a1.4.actions.021", event: "opening_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "teacher", "one"}, {"object", "window", "one"}}},
	{id: "a1.4.actions.022", event: "opening_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "mother", "one"}, {"object", "window", "one"}}},
	{id: "a1.4.actions.023", event: "closing_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "teacher", "one"}, {"object", "window", "one"}}},
	{id: "a1.4.actions.024", event: "closing_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "doctor", "one"}, {"object", "book", "one"}}},
	{id: "a1.4.actions.025", event: "opening_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "mother", "one"}, {"object", "book", "one"}}},
	{id: "a1.4.actions.026", event: "closing_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "father", "one"}, {"object", "bag", "one"}}},
	{id: "a1.4.actions.027", event: "carrying", temporal: "present", participants: []actionParticipantSpec{{"carrier", "teacher", "one"}, {"load", "book", "one"}}},
	{id: "a1.4.actions.028", event: "carrying", temporal: "present", participants: []actionParticipantSpec{{"carrier", "mother", "one"}, {"load", "bag", "one"}}},
	{id: "a1.4.actions.029", event: "holding_object", temporal: "present", participants: []actionParticipantSpec{{"holder", "doctor", "one"}, {"object", "key_tool", "one"}}},
	{id: "a1.4.actions.030", event: "holding_object", temporal: "present", participants: []actionParticipantSpec{{"holder", "father", "one"}, {"object", "book", "one"}}},
	{id: "a1.4.actions.031", event: "pulling_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "dog", "one"}, {"theme", "bag", "one"}}},
	{id: "a1.4.actions.032", event: "pushing_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "teacher", "one"}, {"theme", "table", "one"}}},
	{id: "a1.4.actions.033", event: "cutting_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "teacher", "one"}, {"material", "paper_material", ""}}},
	{id: "a1.4.actions.034", event: "folding_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "mother", "one"}, {"material", "paper_material", ""}}},
	{id: "a1.4.actions.035", event: "tearing_object", temporal: "present", participants: []actionParticipantSpec{{"agent", "dog", "one"}, {"material", "paper_material", ""}}},
	{id: "a1.4.actions.036", event: "buying", temporal: "present", participants: []actionParticipantSpec{{"buyer", "teacher", "one"}, {"goods", "book", "one"}}},
	{id: "a1.4.actions.037", event: "buying", temporal: "present", participants: []actionParticipantSpec{{"buyer", "doctor", "one"}, {"goods", "car", "one"}}},
	{id: "a1.4.actions.038", event: "selling", temporal: "present", participants: []actionParticipantSpec{{"seller", "mother", "one"}, {"goods", "book", "one"}}},
	{id: "a1.4.actions.039", event: "selling", temporal: "present", participants: []actionParticipantSpec{{"seller", "father", "one"}, {"goods", "car", "one"}}},
	{id: "a1.4.actions.040", event: "giving", temporal: "present", participants: []actionParticipantSpec{{"giver", "teacher", "one"}, {"recipient", "mother", "one"}, {"theme", "book", "one"}}},
	{id: "a1.4.actions.041", event: "giving", temporal: "present", participants: []actionParticipantSpec{{"giver", "mother", "one"}, {"recipient", "father", "one"}, {"theme", "key_tool", "one"}}},
	{id: "a1.4.actions.042", event: "giving", temporal: "present", participants: []actionParticipantSpec{{"giver", "doctor", "one"}, {"recipient", "teacher", "one"}, {"theme", "message", "one"}}},
	{id: "a1.4.actions.043", event: "liking", temporal: "present", participants: []actionParticipantSpec{{"experiencer", "teacher", "one"}, {"target", "dog", "one"}}},
	{id: "a1.4.actions.044", event: "liking", temporal: "present", participants: []actionParticipantSpec{{"experiencer", "mother", "one"}, {"target", "cat", "one"}}},
	{id: "a1.4.actions.045", event: "living", temporal: "present", participants: []actionParticipantSpec{{"organism", "bird", "one"}}},
	{id: "a1.4.actions.046", event: "living", temporal: "present", participants: []actionParticipantSpec{{"organism", "fish_animal", "one"}}},
	{id: "a1.4.actions.047", event: "working", temporal: "present", participants: []actionParticipantSpec{{"worker", "doctor", "one"}}},
}

// ActionExpansion builds one authored event graph without parsing either language.
func ActionExpansion(id string) (meaning.Meaning, bool, error) {
	var spec *actionExpansionSpec
	for i := range actionExpansionSpecs {
		if actionExpansionSpecs[i].id == id {
			spec = &actionExpansionSpecs[i]
			break
		}
	}
	if spec == nil {
		return meaning.Meaning{}, false, nil
	}
	eventConcept, err := ids.NewConceptID(spec.event)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("event concept: %w", err)
	}
	participants := make([]meaning.Participant, len(spec.participants))
	nodes := make([]meaning.Node, 1, len(spec.participants)+1)
	for i, participant := range spec.participants {
		target := meaning.NodeID(fmt.Sprintf("r%d", i+1))
		participants[i] = meaning.Participant{Role: meaning.Role(participant.role), Target: target}
		concept, err := ids.NewConceptID(strings.TrimPrefix(participant.concept, "concept:"))
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("participant concept %d: %w", i+1, err)
		}
		referent, err := meaning.NewReferent(target, concept)
		if err != nil {
			return meaning.Meaning{}, false, fmt.Errorf("participant %d: %w", i+1, err)
		}
		if participant.multiplicity == "one" {
			referent.Multiplicity = meaning.MultiplicityOne
		}
		nodes = append(nodes, referent)
	}
	event, err := meaning.NewEvent("e", eventConcept, participants)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("event: %w", err)
	}
	if spec.temporal == "present" {
		event.TemporalFrame = meaning.TemporalFramePresent
	} else if spec.temporal == "past" {
		event.TemporalFrame = meaning.TemporalFramePast
	} else {
		return meaning.Meaning{}, false, fmt.Errorf("unknown temporal frame %q", spec.temporal)
	}
	nodes[0] = event
	graph, err := meaning.Build([]meaning.NodeID{"e"}, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build action Meaning %q: %w", id, err)
	}
	return graph, true, nil
}

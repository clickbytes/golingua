// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

// Reading returns an independent teacher/message graph with authored temporal
// and multiplicity attributes. It does not invoke a language analyzer.
func Reading(id string) (meaning.Meaning, bool, error) {
	var frame meaning.TemporalFrame
	teacherMultiplicity := meaning.MultiplicityOne
	switch id {
	case "reading.teacher-message.present":
		frame = meaning.TemporalFramePresent
	case "reading.teacher-message.past":
		frame = meaning.TemporalFramePast
	case "reading.teachers-message.present":
		frame = meaning.TemporalFramePresent
		teacherMultiplicity = meaning.MultiplicityMany
	case "reading.teachers-message.past":
		frame = meaning.TemporalFramePast
		teacherMultiplicity = meaning.MultiplicityMany
	default:
		return meaning.Meaning{}, false, nil
	}
	readingID, err := ids.NewConceptID("reading")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("reading event concept: %w", err)
	}
	teacherID, err := ids.NewConceptID("teacher")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("reading agent concept: %w", err)
	}
	messageID, err := ids.NewConceptID("message")
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("reading material concept: %w", err)
	}
	event, err := meaning.NewEvent("e", readingID, []meaning.Participant{
		{Role: "reader", Target: "t"},
		{Role: "material", Target: "m"},
	})
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build reading event: %w", err)
	}
	event.TemporalFrame = frame
	teacher, err := meaning.NewReferent("t", teacherID)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build reader referent: %w", err)
	}
	teacher.Multiplicity = teacherMultiplicity
	message, err := meaning.NewReferent("m", messageID)
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build message referent: %w", err)
	}
	message.Multiplicity = meaning.MultiplicityOne
	graph, err := meaning.Build([]meaning.NodeID{"e"}, []meaning.Node{event, teacher, message}, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, fmt.Errorf("build reading Meaning: %w", err)
	}
	return graph, true, nil
}

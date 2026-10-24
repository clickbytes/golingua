// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

// IdentificationMeaning pairs a Meaning graph with its unresolved reference
// context. It does not identify an individual or infer the deictic referent.
type IdentificationMeaning struct {
	Graph   meaning.Meaning
	Context meaning.RequestContext
}

// Identification returns one of the independent identification contracts.
// The neutral anchor modes are: proximal (explicit near deixis), demonstrative
// (explicit pointing with distance unresolved), and unresolved (reference mode
// unresolved, including a possible non-deictic reading). None resolves identity.
// Language-specific surface alternatives belong to realization cases, not I.
func Identification(id string) (IdentificationMeaning, bool, error) {
	var concept string
	var role meaning.Role
	var anchor meaning.DeicticAnchor
	switch id {
	case "identification.classification.this":
		concept, role = "class_membership", "member"
		anchor = meaning.DeicticAnchor{Kind: "deixis", Value: "proximal"}
	case "identification.classification.to":
		concept, role = "class_membership", "member"
		anchor = meaning.DeicticAnchor{Kind: "deixis", Value: "unresolved"}
	case "identification.black.this":
		concept, role = "black", "bearer"
		anchor = meaning.DeicticAnchor{Kind: "deixis", Value: "proximal"}
	case "identification.black.ten":
		concept, role = "black", "bearer"
		anchor = meaning.DeicticAnchor{Kind: "deixis", Value: "demonstrative"}
	default:
		return IdentificationMeaning{}, false, nil
	}

	stateID, err := ids.NewConceptID(concept)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("identification state concept: %w", err)
	}
	catID, err := ids.NewConceptID("cat")
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("identification referent concept: %w", err)
	}
	state, err := meaning.NewState("s", stateID, []meaning.Participant{{Role: role, Target: "r"}})
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build identification state: %w", err)
	}
	state.TemporalFrame = meaning.TemporalFramePresent
	referent, err := meaning.NewReferent("r", catID)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build identification referent: %w", err)
	}
	referent.Multiplicity = meaning.MultiplicityOne
	graph, err := meaning.Build([]meaning.NodeID{"s"}, []meaning.Node{state, referent}, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build identification Meaning: %w", err)
	}
	context := meaning.RequestContext{
		ID:       "matrix-identification",
		Mentions: []meaning.Mention{{ID: "subject", NodeID: "r", Anchor: &anchor}},
	}
	if err := meaning.ValidateReferences(graph, context, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("validate identification reference context: %w", err)
	}
	return IdentificationMeaning{Graph: graph, Context: context}, true, nil
}

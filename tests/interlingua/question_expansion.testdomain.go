// Package interlingua holds hand-authored A1.9 question Meaning fixtures.
package interlingua

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type questionSpec struct {
	predicate, roles, kind, tense, anchor string
	negative                              bool
}

var questionSpecs = map[int]questionSpec{
	1:  {predicate: "class_membership", roles: "member=dog:one", kind: "state", anchor: "proximal"},
	2:  {predicate: "class_membership", roles: "member=window:one", kind: "state", anchor: "distal"},
	3:  {predicate: "black", roles: "bearer=cat:many", kind: "state", anchor: "proximal"},
	4:  {predicate: "black", roles: "bearer=book:many", kind: "state", anchor: "distal"},
	5:  {predicate: "sleeping", roles: "sleeper=dog:one", kind: "event", tense: "present"},
	6:  {predicate: "eating", roles: "eater=cat:one,food=bread:mass", kind: "event", tense: "past"},
	7:  {predicate: "seeing", roles: "experiencer=doctor:one,stimulus=bird:one", kind: "event", tense: "present"},
	8:  {predicate: "hearing", roles: "experiencer=teacher:one,stimulus=telephone_device:one", kind: "event", tense: "present", negative: true},
	9:  {predicate: "sleeping", roles: "sleeper=child:many", kind: "event", tense: "present", negative: true},
	10: {predicate: "black", roles: "bearer=book:one", kind: "state", tense: "present", negative: true},
	45: {predicate: "class_membership", roles: "member=bird:one", kind: "state", anchor: "proximal"},
	46: {predicate: "class_membership", roles: "member=book:one", kind: "state", anchor: "distal"},
	47: {predicate: "class_membership", roles: "member=dog:many", kind: "state", anchor: "proximal"},
	48: {predicate: "black", roles: "bearer=dog:many", kind: "state", anchor: "distal"},
	49: {predicate: "black", roles: "bearer=entity:unspecified", kind: "state", anchor: "proximal"},
	50: {predicate: "black", roles: "bearer=entity:unspecified", kind: "state", anchor: "distal"},
	54: {predicate: "reading", roles: "reader=teacher:one,material=book:one", kind: "event", tense: "past", negative: true},
}

// QuestionExpansion returns a proposition under a question act, plus a bounded
// unresolved anchor when the source form overtly marks deixis. It adds no WH
// variable, answer, or identity inference.
func QuestionExpansion(id string) (IdentificationMeaning, bool, error) {
	parts := strings.Split(id, ".")
	if (len(parts) != 5 && len(parts) != 6) || parts[0] != "a1" || parts[1] != "9" || parts[2] != "q" {
		return IdentificationMeaning{}, false, nil
	}
	var index int
	if _, err := fmt.Sscanf(parts[3], "%d", &index); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("parse question fixture index: %w", err)
	}
	spec, ok := questionSpecs[index]
	if !ok || (parts[4] != "en" && parts[4] != "pl") {
		return IdentificationMeaning{}, false, nil
	}
	variant := ""
	if len(parts) == 6 {
		variant = parts[5]
	}
	if variant != "" && variant != "age" && variant != "offspring" {
		return IdentificationMeaning{}, false, nil
	}
	predicateID, err := ids.NewConceptID(spec.predicate)
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	participants := []meaning.Participant{}
	referents := []meaning.Node{}
	for n, role := range strings.Split(spec.roles, ",") {
		pair := strings.Split(strings.TrimSpace(role), "=")
		if len(pair) != 2 {
			return IdentificationMeaning{}, false, fmt.Errorf("invalid question role %q", role)
		}
		entity := strings.Split(pair[1], ":")
		concept := entity[0]
		if concept == "child" && variant == "offspring" {
			concept = "child_of_person"
		}
		conceptID, err := ids.NewConceptID(concept)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		nodeID := meaning.NodeID(fmt.Sprintf("r%d", n+1))
		ref, err := meaning.NewReferent(nodeID, conceptID)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		if len(entity) > 1 {
			if entity[1] == "one" {
				ref.Multiplicity = meaning.MultiplicityOne
			}
			if entity[1] == "many" {
				ref.Multiplicity = meaning.MultiplicityMany
			}
		}
		referents = append(referents, ref)
		participants = append(participants, meaning.Participant{Role: meaning.Role(pair[0]), Target: nodeID})
	}
	var proposition meaning.Node
	if spec.kind == "event" {
		proposition, err = meaning.NewEvent("pred", predicateID, participants)
	} else {
		proposition, err = meaning.NewState("pred", predicateID, participants)
	}
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	if spec.tense == "past" {
		proposition.TemporalFrame = meaning.TemporalFramePast
	} else {
		proposition.TemporalFrame = meaning.TemporalFramePresent
	}
	nodes := append([]meaning.Node{proposition}, referents...)
	target := meaning.NodeID("pred")
	if spec.negative {
		neg, e := meaning.NewNegation("neg", target)
		if e != nil {
			return IdentificationMeaning{}, false, e
		}
		nodes = append(nodes, neg)
		target = "neg"
	}
	question, err := meaning.NewQuestion("question", target)
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	nodes = append(nodes, question)
	discourse, err := meaning.NewDiscourse("root", meaning.ActQuestion, []meaning.NodeID{"question"})
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	nodes = append(nodes, discourse)
	graph, err := meaning.Build([]meaning.NodeID{"root"}, nodes, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	context := meaning.RequestContext{ID: "a1-9-question"}
	if spec.anchor != "" {
		anchor := spec.anchor
		if parts[4] == "pl" {
			switch index {
			case 1, 45, 47, 49:
				anchor = "unresolved"
			case 2, 50:
				anchor = "distal"
			case 3:
				anchor = "demonstrative"
			case 4, 48:
				anchor = "distal"
			}
		}
		context.Mentions = []meaning.Mention{{ID: "deictic", NodeID: meaning.NodeID("r1"), Anchor: &meaning.DeicticAnchor{Kind: "deixis", Value: anchor}}}
	}
	if err := meaning.ValidateReferences(graph, context, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, err
	}
	return IdentificationMeaning{Graph: graph, Context: context}, true, nil
}

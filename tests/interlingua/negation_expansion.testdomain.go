// Package interlingua stores independently authored negative Meaning graphs for A1.8.
package interlingua

import (
	"fmt"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"strings"
)

type NegationExpansionSpec struct {
	Index                            int
	Predicate, Roles, Temporal, Kind string
}

var negationExpansionSpecs = []NegationExpansionSpec{
	{Index: 1, Predicate: "sleeping", Roles: "sleeper=dog:one", Temporal: "present", Kind: "event"},
	{Index: 2, Predicate: "eating", Roles: "eater=cat:one,food=bread:mass", Temporal: "present", Kind: "event"},
	{Index: 3, Predicate: "drinking", Roles: "drinker=child:one,liquid=milk_food:mass", Temporal: "present", Kind: "event"},
	{Index: 4, Predicate: "reading", Roles: "reader=doctor:one,material=letter_document:one", Temporal: "present", Kind: "event"},
	{Index: 5, Predicate: "writing_content", Roles: "author=teacher:one,document=message:one", Temporal: "present", Kind: "writing_content_gap"},
	{Index: 6, Predicate: "running", Roles: "runner=dog:one", Temporal: "present", Kind: "event"},
	{Index: 7, Predicate: "walking", Roles: "walker=cat:one", Temporal: "present", Kind: "event"},
	{Index: 8, Predicate: "flying_motion", Roles: "mover=bird:one", Temporal: "present", Kind: "event"},
	{Index: 9, Predicate: "opening_object", Roles: "agent=doctor:one,object=door:one", Temporal: "present", Kind: "event"},
	{Index: 10, Predicate: "closing_object", Roles: "agent=teacher:one,object=window:one", Temporal: "present", Kind: "event"},
	{Index: 11, Predicate: "seeing", Roles: "experiencer=dog:one,stimulus=cat:one", Temporal: "present", Kind: "event"},
	{Index: 12, Predicate: "hearing", Roles: "experiencer=cat:one,stimulus=child:one", Temporal: "present", Kind: "event"},
	{Index: 13, Predicate: "eating", Roles: "eater=doctor:many,food=apple_fruit:many", Temporal: "present", Kind: "event"},
	{Index: 14, Predicate: "drinking", Roles: "drinker=child:many,liquid=water:mass", Temporal: "present", Kind: "event"},
	{Index: 15, Predicate: "reading", Roles: "reader=teacher:many,material=book:one", Temporal: "present", Kind: "event"},
	{Index: 16, Predicate: "sleeping", Roles: "sleeper=dog:many", Temporal: "present", Kind: "event"},
	{Index: 17, Predicate: "running", Roles: "runner=cat:many", Temporal: "present", Kind: "event"},
	{Index: 18, Predicate: "eating", Roles: "eater=bird:many,food=bread:mass", Temporal: "present", Kind: "event"},
	{Index: 19, Predicate: "hearing", Roles: "experiencer=mother:one,stimulus=bird:many", Temporal: "present", Kind: "event"},
	{Index: 20, Predicate: "seeing", Roles: "experiencer=doctor:many,stimulus=car:one", Temporal: "present", Kind: "event"},
	{Index: 21, Predicate: "black", Roles: "bearer=dog:one", Temporal: "present", Kind: "state"},
	{Index: 22, Predicate: "black", Roles: "bearer=cat:one", Temporal: "present", Kind: "state"},
	{Index: 23, Predicate: "black", Roles: "bearer=book:one", Temporal: "present", Kind: "state"},
	{Index: 24, Predicate: "black", Roles: "bearer=window:one", Temporal: "present", Kind: "state"},
	{Index: 25, Predicate: "black", Roles: "bearer=door:one", Temporal: "present", Kind: "state"},
	{Index: 26, Predicate: "black", Roles: "bearer=book:many", Temporal: "present", Kind: "state"},
	{Index: 27, Predicate: "black", Roles: "bearer=cat:many", Temporal: "present", Kind: "state"},
	{Index: 28, Predicate: "sadness", Roles: "bearer=teacher:one", Temporal: "present", Kind: "state_gap"},
	{Index: 29, Predicate: "sadness", Roles: "bearer=child:one", Temporal: "present", Kind: "state_gap"},
	{Index: 30, Predicate: "sadness", Roles: "bearer=mother:one", Temporal: "present", Kind: "state_gap"},
	{Index: 31, Predicate: "sleeping", Roles: "sleeper=dog:one", Temporal: "past", Kind: "event"},
	{Index: 32, Predicate: "eating", Roles: "eater=cat:many,food=bread:mass", Temporal: "past", Kind: "event"},
	{Index: 33, Predicate: "drinking", Roles: "drinker=mother:one,liquid=water:mass", Temporal: "past", Kind: "event"},
	{Index: 34, Predicate: "reading", Roles: "reader=doctor:many,material=letter_document:one", Temporal: "past", Kind: "event"},
	{Index: 35, Predicate: "seeing", Roles: "experiencer=dog:one,stimulus=cat:one", Temporal: "past", Kind: "event"},
	{Index: 36, Predicate: "hearing", Roles: "experiencer=teacher:one,stimulus=telephone_device:one", Temporal: "past", Kind: "event"},
	{Index: 37, Predicate: "running", Roles: "runner=child:many", Temporal: "past", Kind: "event"},
	{Index: 38, Predicate: "closing_object", Roles: "agent=doctor:one,object=door:one", Temporal: "past", Kind: "event"},
	{Index: 39, Predicate: "black", Roles: "bearer=book:one", Temporal: "past", Kind: "state"},
	{Index: 40, Predicate: "black", Roles: "bearer=window:one", Temporal: "past", Kind: "state"},
	{Index: 41, Predicate: "sleeping", Roles: "sleeper=dog:one", Temporal: "present", Kind: "frequency_gap"},
	{Index: 42, Predicate: "eating", Roles: "eater=cat:many", Temporal: "present", Kind: "frequency_gap"},
	{Index: 43, Predicate: "sleeping", Roles: "sleeper=dog:one", Temporal: "present", Kind: "quantified_location_gap"},
	{Index: 44, Predicate: "drinking", Roles: "drinker=cat:one,liquid=milk_food:mass", Temporal: "present", Kind: "quantifier_scope_gap"},
	{Index: 45, Predicate: "sleeping", Roles: "sleeper=dog:one", Temporal: "present", Kind: "focus_gap"},
}

// NegationExpansion builds a hand-authored graph; gap cases intentionally have no graph.
func NegationExpansion(id string) (meaning.Meaning, bool, error) {
	var spec NegationExpansionSpec
	variant := ""
	for _, s := range negationExpansionSpecs {
		base := fmt.Sprintf("a1.8.neg.%03d", s.Index)
		if id == base || id == base+".age" || id == base+".offspring" {
			spec = s
			if strings.HasSuffix(id, ".age") {
				variant = "age"
			}
			if strings.HasSuffix(id, ".offspring") {
				variant = "offspring"
			}
			break
		}
	}
	if spec.Index == 0 || strings.HasSuffix(spec.Kind, "_gap") {
		return meaning.Meaning{}, false, nil
	}
	if spec.Kind != "event" && spec.Kind != "state" {
		return meaning.Meaning{}, false, nil
	}
	participants := []meaning.Participant{}
	referents := []meaning.Node{}
	roleParts := strings.Split(spec.Roles, ",")
	for i, part := range roleParts {
		fields := strings.Split(strings.TrimSpace(part), "=")
		if len(fields) != 2 {
			return meaning.Meaning{}, false, fmt.Errorf("bad role %q", part)
		}
		val := strings.Split(fields[1], ":")
		concept := val[0]
		mult := ""
		if len(val) > 1 {
			mult = val[1]
		}
		if concept == "child" && (variant == "offspring" || (variant == "" && id == fmt.Sprintf("a1.8.neg.%03d.offspring", spec.Index))) {
			concept = "child_of_person"
		}
		cid, err := ids.NewConceptID(concept)
		if err != nil {
			return meaning.Meaning{}, false, err
		}
		rid := meaning.NodeID(fmt.Sprintf("r%d", i+1))
		ref, err := meaning.NewReferent(rid, cid)
		if err != nil {
			return meaning.Meaning{}, false, err
		}
		if mult == "one" {
			ref.Multiplicity = meaning.MultiplicityOne
		}
		if mult == "many" {
			ref.Multiplicity = meaning.MultiplicityMany
		}
		referents = append(referents, ref)
		participants = append(participants, meaning.Participant{Role: meaning.Role(fields[0]), Target: rid})
	}
	predicateID, err := ids.NewConceptID(spec.Predicate)
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	var pred meaning.Node
	if spec.Kind == "state" {
		pred, err = meaning.NewState("pred", predicateID, participants)
	} else {
		pred, err = meaning.NewEvent("pred", predicateID, participants)
	}
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	if spec.Temporal == "past" {
		pred.TemporalFrame = meaning.TemporalFramePast
	} else {
		pred.TemporalFrame = meaning.TemporalFramePresent
	}
	scopeTarget := meaning.NodeID("pred")
	nodes := append([]meaning.Node{pred}, referents...)
	neg, err := meaning.NewNegation("neg", scopeTarget)
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	nodes = append([]meaning.Node{neg}, nodes...)
	graph, err := meaning.Build([]meaning.NodeID{"neg"}, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, false, err
	}
	return graph, true, nil
}

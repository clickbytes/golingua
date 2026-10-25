// Package interlingua stores independent, hand-authored A1 identification fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type identificationExpansionSpec struct {
	pair                                                       int
	kind, predicate, concept, multiplicity, anchorEN, anchorPL string
}

var identificationExpansionSpecs = []identificationExpansionSpec{
	{pair: 1, kind: "classification", predicate: "class_membership", concept: "dog", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 2, kind: "classification", predicate: "class_membership", concept: "bird", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 3, kind: "classification", predicate: "class_membership", concept: "book", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 4, kind: "classification", predicate: "class_membership", concept: "car", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 5, kind: "classification", predicate: "class_membership", concept: "table", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 6, kind: "classification", predicate: "class_membership", concept: "chair", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 7, kind: "classification", predicate: "class_membership", concept: "window", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 8, kind: "classification", predicate: "class_membership", concept: "apple_fruit", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 9, kind: "classification", predicate: "class_membership", concept: "key_tool", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 10, kind: "classification", predicate: "class_membership", concept: "bag", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 11, kind: "classification", predicate: "class_membership", concept: "teacher", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 12, kind: "classification", predicate: "class_membership", concept: "doctor", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 13, kind: "classification", predicate: "class_membership", concept: "mother", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 14, kind: "classification", predicate: "class_membership", concept: "father", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 15, kind: "classification", predicate: "class_membership", concept: "horse", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 16, kind: "classification", predicate: "class_membership", concept: "fish_animal", multiplicity: "one", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 17, kind: "classification", predicate: "class_membership", concept: "cat", multiplicity: "one", anchorEN: "distal", anchorPL: "unresolved"},
	{pair: 18, kind: "classification", predicate: "class_membership", concept: "dog", multiplicity: "one", anchorEN: "distal", anchorPL: "unresolved"},
	{pair: 19, kind: "classification", predicate: "class_membership", concept: "book", multiplicity: "one", anchorEN: "distal", anchorPL: "unresolved"},
	{pair: 20, kind: "classification", predicate: "class_membership", concept: "apple_fruit", multiplicity: "one", anchorEN: "distal", anchorPL: "unresolved"},
	{pair: 21, kind: "classification", predicate: "class_membership", concept: "cat", multiplicity: "many", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 22, kind: "classification", predicate: "class_membership", concept: "dog", multiplicity: "many", anchorEN: "distal", anchorPL: "unresolved"},
	{pair: 23, kind: "property", predicate: "black", concept: "dog", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 24, kind: "property", predicate: "black", concept: "bird", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 25, kind: "property", predicate: "black", concept: "horse", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 26, kind: "property", predicate: "black", concept: "car", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 27, kind: "property", predicate: "black", concept: "book", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 28, kind: "property", predicate: "black", concept: "bag", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 29, kind: "property", predicate: "black", concept: "window", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 30, kind: "property", predicate: "black", concept: "apple_fruit", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 31, kind: "property", predicate: "black", concept: "cat", multiplicity: "many", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 32, kind: "property", predicate: "black", concept: "book", multiplicity: "many", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 33, kind: "classification", predicate: "class_membership", concept: "teacher", multiplicity: "many", anchorEN: "proximal", anchorPL: "unresolved"},
	{pair: 34, kind: "property", predicate: "black", concept: "horse", multiplicity: "many", anchorEN: "distal", anchorPL: "distal"},
	{pair: 35, kind: "missing_colour", predicate: "red", concept: "book", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 36, kind: "missing_colour", predicate: "white", concept: "car", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 37, kind: "missing_colour", predicate: "blue", concept: "bag", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
	{pair: 38, kind: "missing_colour", predicate: "green", concept: "apple_fruit", multiplicity: "one", anchorEN: "proximal", anchorPL: "demonstrative"},
}

// IdentificationExpansionFixtureIDs returns IDs for the supported, authored one-referent fixtures.
func IdentificationExpansionFixtureIDs() []string {
	var result []string
	for _, spec := range identificationExpansionSpecs {
		if spec.kind == "missing_colour" {
			continue
		}
		result = append(result, fmt.Sprintf("a1.5.identification.%03d.en", spec.pair), fmt.Sprintf("a1.5.identification.%03d.pl", spec.pair))
	}
	return result
}

// IdentificationExpansion builds an independent graph and unresolved reference context.
func IdentificationExpansion(id string) (IdentificationMeaning, bool, error) {
	var spec *identificationExpansionSpec
	var anchorValue string
	for i := range identificationExpansionSpecs {
		candidate := &identificationExpansionSpecs[i]
		if id == fmt.Sprintf("a1.5.identification.%03d.en", candidate.pair) {
			spec, anchorValue = candidate, candidate.anchorEN
			break
		}
		if id == fmt.Sprintf("a1.5.identification.%03d.pl", candidate.pair) {
			spec, anchorValue = candidate, candidate.anchorPL
			break
		}
	}
	if spec == nil || spec.kind == "missing_colour" {
		return IdentificationMeaning{}, false, nil
	}
	role := meaning.Role("member")
	if spec.kind == "property" {
		role = "bearer"
	} else if spec.kind != "classification" {
		return IdentificationMeaning{}, false, fmt.Errorf("unknown identification kind %q", spec.kind)
	}
	predicateID, err := ids.NewConceptID(spec.predicate)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("predicate concept: %w", err)
	}
	conceptID, err := ids.NewConceptID(spec.concept)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("referent concept: %w", err)
	}
	state, err := meaning.NewState("s", predicateID, []meaning.Participant{{Role: role, Target: "r"}})
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("state: %w", err)
	}
	state.TemporalFrame = meaning.TemporalFramePresent
	referent, err := meaning.NewReferent("r", conceptID)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("referent: %w", err)
	}
	if spec.multiplicity == "one" {
		referent.Multiplicity = meaning.MultiplicityOne
	} else if spec.multiplicity == "many" {
		referent.Multiplicity = meaning.MultiplicityMany
	} else {
		return IdentificationMeaning{}, false, fmt.Errorf("unsupported multiplicity %q", spec.multiplicity)
	}
	graph, err := meaning.Build([]meaning.NodeID{"s"}, []meaning.Node{state, referent}, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build graph: %w", err)
	}
	fixture := IdentificationMeaning{Graph: graph, Context: meaning.RequestContext{ID: "a1.5-identification", Mentions: []meaning.Mention{{ID: "subject", NodeID: "r", Anchor: &meaning.DeicticAnchor{Kind: "deixis", Value: anchorValue}}}}}
	if err := meaning.ValidateReferences(graph, fixture.Context, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("validate unresolved mention: %w", err)
	}
	return fixture, true, nil
}

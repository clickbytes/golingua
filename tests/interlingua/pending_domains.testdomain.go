// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

// PendingDomain returns independently authored representable meanings and components.
// A supported component does not imply that a complete sentence graph exists.
func PendingDomain(id string) (meaning.Meaning, *IdentificationMeaning, bool, error) {
	switch id {
	case "quantity.exactly-two-cats.component":
		catID, err := ids.NewConceptID("cat")
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("exact-two cat concept: %w", err)
		}
		cat, err := meaning.NewReferent("cats", catID)
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("exact-two cat referent: %w", err)
		}
		cat.Multiplicity = meaning.MultiplicityMany
		count := 2
		quantifier, err := meaning.NewQuantifier("exactly_two", cat.ID, meaning.Quantifier{Kind: meaning.QuantifierExactly, Count: &count})
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("exact-two cat quantifier: %w", err)
		}
		graph, err := meaning.Build([]meaning.NodeID{quantifier.ID}, []meaning.Node{quantifier, cat}, meaning.DefaultBudget())
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("build exact-two cat Meaning: %w", err)
		}
		return graph, nil, true, nil
	case "question.classification.this", "question.classification.to":
		baseID := "identification.classification.this"
		if id == "question.classification.to" {
			baseID = "identification.classification.to"
		}
		fixture, ok, err := Identification(baseID)
		if err != nil || !ok {
			if err == nil {
				err = fmt.Errorf("identification fixture %q is unavailable", baseID)
			}
			return meaning.Meaning{}, nil, false, err
		}
		question, err := meaning.NewQuestion("question", fixture.Graph.Roots[0])
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("build classification question: %w", err)
		}
		discourse, err := meaning.NewDiscourse("root", meaning.ActQuestion, []meaning.NodeID{question.ID})
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("build question discourse: %w", err)
		}
		nodes := append([]meaning.Node{question, discourse}, fixture.Graph.Nodes...)
		graph, err := meaning.Build([]meaning.NodeID{discourse.ID}, nodes, meaning.DefaultBudget())
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("build questioned classification Meaning: %w", err)
		}
		fixture.Graph = graph
		if err := meaning.ValidateReferences(graph, fixture.Context, meaning.DefaultBudget()); err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("validate question reference context: %w", err)
		}
		return graph, &fixture, true, nil
	default:
		return meaning.Meaning{}, nil, false, nil
	}
}

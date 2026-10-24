// Package interlingua stores independent, hand-authored Meaning fixtures.
package interlingua

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/meaning"
)

// Negation returns an explicit negative scope over an existing authored
// identification or seeing graph. Classification keeps its unresolved
// request context alongside the wrapped graph.
func Negation(id string) (meaning.Meaning, *IdentificationMeaning, bool, error) {
	var base meaning.Meaning
	var contextFixture *IdentificationMeaning
	switch id {
	case "negation.classification.this":
		fixture, ok, err := Identification("identification.classification.this")
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("load proximal classification fixture: %w", err)
		}
		if !ok {
			return meaning.Meaning{}, nil, false, fmt.Errorf("proximal classification fixture is unavailable")
		}
		base = fixture.Graph
		contextFixture = &fixture
	case "negation.classification.to":
		fixture, ok, err := Identification("identification.classification.to")
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("load unresolved classification fixture: %w", err)
		}
		if !ok {
			return meaning.Meaning{}, nil, false, fmt.Errorf("unresolved classification fixture is unavailable")
		}
		base = fixture.Graph
		contextFixture = &fixture
	case "negation.seeing.teacher-dog":
		graph, ok, err := Seeing("seeing.teacher-dog")
		if err != nil {
			return meaning.Meaning{}, nil, false, fmt.Errorf("load teacher-dog seeing fixture: %w", err)
		}
		if !ok {
			return meaning.Meaning{}, nil, false, fmt.Errorf("teacher-dog seeing fixture is unavailable")
		}
		base = graph
	default:
		return meaning.Meaning{}, nil, false, nil
	}
	if len(base.Roots) != 1 {
		return meaning.Meaning{}, nil, false, fmt.Errorf("negation fixture %q requires one source root", id)
	}
	scope, err := meaning.NewNegation("neg", base.Roots[0])
	if err != nil {
		return meaning.Meaning{}, nil, false, fmt.Errorf("build negation scope for %q: %w", id, err)
	}
	nodes := append([]meaning.Node{scope}, base.Nodes...)
	graph, err := meaning.Build([]meaning.NodeID{scope.ID}, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, nil, false, fmt.Errorf("build negated Meaning %q: %w", id, err)
	}
	if contextFixture != nil {
		contextFixture.Graph = graph
		return graph, contextFixture, true, nil
	}
	return graph, nil, true, nil
}

package facts

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/meaning"
)

// Budget bounds a single request's explicit fact inventory.
type Budget struct {
	MaxFacts     int
	MaxReferents int
}

func DefaultBudget() Budget { return Budget{MaxFacts: MaxFacts, MaxReferents: meaning.MaxReferents} }

// Set owns defensive copies of one request context and its assertions. It
// does not persist or amend the ontology. Conflicting facts retain both IDs.
type Set struct {
	context meaning.RequestContext
	facts   []Fact
	byID    map[string]int
}

// Build checks the input schema before constructing an immutable Set.
func Build(ontology *core.ConceptGraph, input Input, budget Budget) (*Set, error) {
	if input.SchemaVersion != SchemaVersion {
		return nil, factError("INVALID_FACT_SCHEMA_VERSION", "schema_version", "unsupported fact input schema version")
	}
	return NewSet(ontology, input.Context, input.Facts, budget)
}

// NewSet validates all input before exposing an immutable request snapshot.
func NewSet(ontology *core.ConceptGraph, context meaning.RequestContext, input []Fact, budget Budget) (*Set, error) {
	if budget == (Budget{}) {
		budget = DefaultBudget()
	}
	if budget.MaxFacts <= 0 || budget.MaxFacts > MaxFacts || budget.MaxReferents <= 0 || budget.MaxReferents > meaning.MaxReferents {
		return nil, factError("INVALID_FACT_BUDGET", "budget", "fact budget must be positive and within package ceilings")
	}
	if len(input) > budget.MaxFacts || len(context.Referents) > budget.MaxReferents {
		return nil, factLimit("facts_or_referents", len(input)+len(context.Referents))
	}
	if err := validateContext(ontology, context); err != nil {
		return nil, err
	}
	result := &Set{context: cloneContext(context), facts: make([]Fact, 0, len(input)), byID: make(map[string]int, len(input))}
	for index, item := range input {
		if err := validateFact(ontology, context, item, index); err != nil {
			return nil, err
		}
		if _, exists := result.byID[item.ID]; exists {
			return nil, factError("DUPLICATE_FACT_ID", fmt.Sprintf("facts[%d].id", index), "fact ID is duplicated")
		}
		result.byID[item.ID] = index
		result.facts = append(result.facts, cloneFact(item))
	}
	return result, nil
}

func (set *Set) Context() meaning.RequestContext {
	if set == nil {
		return meaning.RequestContext{}
	}
	return cloneContext(set.context)
}

func (set *Set) Facts() []Fact {
	if set == nil {
		return nil
	}
	result := make([]Fact, len(set.facts))
	for index, fact := range set.facts {
		result[index] = cloneFact(fact)
	}
	return result
}

func (set *Set) ByID(id string) (Fact, bool) {
	if set == nil {
		return Fact{}, false
	}
	index, exists := set.byID[id]
	if !exists {
		return Fact{}, false
	}
	return cloneFact(set.facts[index]), true
}

// EvidenceFor returns both explicit polarities in input order. Empty slices
// mean unknown; callers may report a conflict when both are non-empty.
func (set *Set) EvidenceFor(predicate core.RelationID, args [2]meaning.ReferentID) (positive, negative []Fact) {
	if set == nil {
		return nil, nil
	}
	for _, fact := range set.facts {
		if fact.Predicate != predicate.String() || fact.ArgReferentIDs != args {
			continue
		}
		if fact.Polarity == Positive {
			positive = append(positive, cloneFact(fact))
		} else {
			negative = append(negative, cloneFact(fact))
		}
	}
	return positive, negative
}

func cloneFact(fact Fact) Fact {
	if fact.Source.Reference != nil {
		reference := *fact.Source.Reference
		fact.Source.Reference = &reference
	}
	return fact
}

func cloneContext(context meaning.RequestContext) meaning.RequestContext {
	copyOf := context
	copyOf.Referents = append([]meaning.Referent(nil), context.Referents...)
	copyOf.Mentions = append([]meaning.Mention(nil), context.Mentions...)
	for index := range copyOf.Mentions {
		if context.Mentions[index].Anchor != nil {
			anchor := *context.Mentions[index].Anchor
			copyOf.Mentions[index].Anchor = &anchor
		}
	}
	return copyOf
}

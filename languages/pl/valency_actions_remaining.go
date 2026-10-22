package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// remainingActionFrames records bounded lexical patterns; clause complements
// still require construction-level data.
var remainingActionFrames = []frameDeclaration{
	{"pl.skarzyc_sie.complain.action", "pl.skarzyc_sie.complain", "pl.transitive_oblique", []meaning.Role{"complainant", "recipient", "problem"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "complainant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "oblique", Role: "problem", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.zartowac.joke.action", "pl.zartowac.joke", "pl.oblique", []meaning.Role{"speaker", "audience"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "audience", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
	}},
	{"pl.wolac.prefer.action", "pl.wolac.prefer", "pl.transitive_oblique", []meaning.Role{"chooser", "preferred", "alternative"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "chooser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "preferred", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "alternative", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.probowac.try.action", "pl.probowac.try", "pl.oblique", []meaning.Role{"agent", "goal"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "goal", Required: true, GrammaticalFunction: "oblique", Case: "genitive"},
	}},
}

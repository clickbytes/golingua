package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// productionActionFrames records explicit Polish case and adposition choices.
var productionActionFrames = []frameDeclaration{
	{"pl.krasc.steal.action", "pl.krasc.steal", "pl.ditransitive", []meaning.Role{"taker", "owner", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "taker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "owner", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.kategoryzowac.categorize.action", "pl.kategoryzowac.categorize", "pl.transitive_oblique", []meaning.Role{"classifier", "item", "class"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "classifier", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "item", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "class", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "jako"},
	}},
	{"pl.rysowac.draw.action", "pl.rysowac.draw", "pl.oblique", []meaning.Role{"artist", "medium"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "artist", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "medium", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "na"},
	}},
	{"pl.nagrywac.record.action", "pl.nagrywac.record", "pl.transitive", []meaning.Role{"recorder", "source"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recorder", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "source", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wspolpracowac.cooperate.action", "pl.wspolpracowac.cooperate", "pl.transitive_oblique", []meaning.Role{"participant_a", "participant_b", "goal"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant_a", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "participant_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "goal", Required: true, GrammaticalFunction: "complement", Case: "instrumental", Adposition: "nad"},
	}},
	{"pl.ubierac_sie.dress.action", "pl.ubierac_sie.dress", "pl.transitive_oblique", []meaning.Role{"dresser", "clothes"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "dresser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "clothes", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "w"},
	}},
}

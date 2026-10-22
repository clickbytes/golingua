package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// activityActionFrames records selected Polish argument cases. These frames do
// not approve construction bindings.
var activityActionFrames = []frameDeclaration{
	{"pl.porownywac.compare.action", "pl.porownywac.compare", "pl.double_np", []meaning.Role{"comparer", "item_a", "item_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "comparer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object_a", Role: "item_a", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "object_b", Role: "item_b", Required: true, GrammaticalFunction: "complement", Case: "accusative"},
	}},
	{"pl.ranic.injure.action", "pl.ranic.injure", "pl.transitive", []meaning.Role{"cause", "patient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cause", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "patient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.sadzic.plant.action", "pl.sadzic.plant", "pl.transitive", []meaning.Role{"planter", "seed"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "planter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "seed", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.ostrzegac.warn.action", "pl.ostrzegac.warn", "pl.transitive_oblique", []meaning.Role{"warner", "recipient", "risk"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "warner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "risk", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "przed"},
	}},
	{"pl.polecac.recommend.action", "pl.polecac.recommend", "pl.ditransitive", []meaning.Role{"recommender", "recipient", "option"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recommender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "option", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.zaczynac.start.action", "pl.zaczynac.start", "pl.transitive", []meaning.Role{"activity"}, []valency.ArgumentSlot{
		{ID: "object", Role: "activity", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.konczyc.finish.action", "pl.konczyc.finish", "pl.transitive", []meaning.Role{"activity"}, []valency.ArgumentSlot{
		{ID: "object", Role: "activity", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.kontynuowac.continue.action", "pl.kontynuowac.continue", "pl.transitive", []meaning.Role{"activity"}, []valency.ArgumentSlot{
		{ID: "object", Role: "activity", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

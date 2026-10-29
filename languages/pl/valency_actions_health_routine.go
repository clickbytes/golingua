package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var healthRoutineActionFrames = []frameDeclaration{
	{"pl.dostac.receive_message.action", "pl.dostac.receive_message", "pl.transitive_oblique", []meaning.Role{"recipient", "sender", "message"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "message", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "sender", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.zazywac.medicine.action", "pl.zazywac.medicine", "pl.transitive", []meaning.Role{"patient", "medicine"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "patient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "medicine", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pisac.document.action", "pl.pisac.document", "pl.transitive", []meaning.Role{"author", "document"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "author", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "document", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pisac.instrument.action", "pl.pisac.instrument", "pl.transitive", []meaning.Role{"writer", "instrument"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "instrument", Required: true, GrammaticalFunction: "object", Case: "instrumental"},
	}},
	{"pl.gotowac.boil.action", "pl.gotowac.boil", "pl.transitive", []meaning.Role{"cook", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cook", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zbierac.harvest.action", "pl.zbierac.harvest", "pl.transitive", []meaning.Role{"harvester", "crop"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "harvester", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "crop", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var communicationLastFrames = []frameDeclaration{
	{"pl.zamawiac_usluge.hire.action", "pl.zamawiac_usluge.hire", "pl.oblique", []meaning.Role{"client", "provider"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "client", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "provider", Role: "provider", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "u"},
	}},
	{"pl.zle_zrozumiec.misunderstand.action", "pl.zle_zrozumiec.misunderstand", "pl.transitive", []meaning.Role{"interpreter", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "interpreter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "content", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.tlumaczyc_ustnie.interpret.action", "pl.tlumaczyc_ustnie.interpret", "pl.transitive_oblique", []meaning.Role{"interpreter", "speaker", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "interpreter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "speaker", Role: "speaker", Required: true, GrammaticalFunction: "possessor", Case: "genitive"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
}

package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var phraseLevelActionFrames = []frameDeclaration{
	{"pl.nie_lubic.dislike.action", "pl.nie_lubic.dislike", "pl.transitive", []meaning.Role{"experiencer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.czerpac_przyjemnosc.enjoy.action", "pl.czerpac_przyjemnosc.enjoy", "pl.oblique", []meaning.Role{"experiencer", "source"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "source", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.dochodzic_do_siebie.recover.action", "pl.dochodzic_do_siebie.recover", "pl.intransitive", []meaning.Role{"patient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "patient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.robic_zakupy.shop.action", "pl.robic_zakupy.shop", "pl.intransitive", []meaning.Role{"shopper"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "shopper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}

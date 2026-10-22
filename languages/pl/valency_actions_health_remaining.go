package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var healthRemainingFrames = []frameDeclaration{
	{"pl.zachorowac.ill.action", "pl.zachorowac.ill", "pl.intransitive", []meaning.Role{"patient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "patient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.myc_zeby.brush.action", "pl.myc_zeby.brush", "pl.oblique", []meaning.Role{"person", "brush"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "complement", Role: "brush", Required: true, GrammaticalFunction: "complement", Case: "instrumental"},
	}},
	{"pl.wymieniac_sie.exchange.action", "pl.wymieniac_sie.exchange", "pl.transitive_oblique", []meaning.Role{"party_a", "party_b", "goods_a", "goods_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "party_a", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique_party", Role: "party_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement_a", Role: "goods_a", Required: true, GrammaticalFunction: "complement", Case: "instrumental"},
		{ID: "oblique_goods", Role: "goods_b", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.goic_sie.heal.action", "pl.goic_sie.heal", "pl.intransitive", []meaning.Role{"tissue"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "tissue", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}

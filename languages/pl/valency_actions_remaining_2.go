package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var remainingActionFrames2 = []frameDeclaration{
	{"pl.pozyczac_komus.lend.action", "pl.pozyczac_komus.lend", "pl.ditransitive", []meaning.Role{"lender", "borrower", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "lender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "borrower", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zamawiac_towary.order.action", "pl.zamawiac_towary.order", "pl.transitive_oblique", []meaning.Role{"customer", "supplier", "goods"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "customer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "supplier", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "u"},
	}},
	{"pl.pisac.type.action", "pl.pisac.type", "pl.oblique", []meaning.Role{"typist", "device"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "typist", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "device", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "na"},
	}},
	{"pl.chciec.desire.action", "pl.chciec.desire", "pl.oblique", []meaning.Role{"experiencer", "desired"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "desired", Required: true, GrammaticalFunction: "oblique", Case: "genitive"},
	}},
}

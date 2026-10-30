package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var polishCommerceOmissionFrames = []frameDeclaration{
	{"pl.kupowac.buy.omit_seller", "pl.kupowac.buy", "pl.transitive", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.kupic.buy.omit_seller", "pl.kupic.buy", "pl.transitive", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.kupowac.buy.negative", "pl.kupowac.buy", "pl.transitive_negative", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "genitive"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.kupic.buy.negative", "pl.kupic.buy", "pl.transitive_negative", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "genitive"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.sprzedawac.sell.omit_buyer", "pl.sprzedawac.sell", "pl.transitive", []meaning.Role{"seller", "goods", "buyer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "seller", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "counterparty", Role: "buyer", Required: false, GrammaticalFunction: "oblique", Case: "dative"},
	}},
}

var polishCommerceOmittedRoles = map[string][]meaning.Role{
	"pl.kupowac.buy.omit_seller":    {"seller"},
	"pl.kupic.buy.omit_seller":      {"seller"},
	"pl.kupowac.buy.negative":       {"seller"},
	"pl.kupic.buy.negative":         {"seller"},
	"pl.sprzedawac.sell.omit_buyer": {"buyer"},
	"pl.pisac.write.message_only":   {"recipient"},
	"pl.pisac.write.negative":       {"recipient"},
	"pl.pisac.write.recipient_only": {"message"},
}

func polishCommerceOmissionBindings() []constructionBindingDeclaration {
	bindings := []constructionBindingDeclaration{
		{"pl.kupowac.buy", "pl.kupowac.buy.omit_seller", "pl.transitive", "pl.transitive.svo"},
		{"pl.kupic.buy", "pl.kupic.buy.omit_seller", "pl.transitive", "pl.transitive.svo"},
		{"pl.kupowac.buy", "pl.kupowac.buy.negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
		{"pl.kupic.buy", "pl.kupic.buy.negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
		{"pl.sprzedawac.sell", "pl.sprzedawac.sell.omit_buyer", "pl.transitive", "pl.transitive.svo"},
	}
	return bindings
}

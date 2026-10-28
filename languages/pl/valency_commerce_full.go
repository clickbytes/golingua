package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var polishCommerceFullPurchaseFrames = []frameDeclaration{
	{"pl.kupowac.buy.from_seller", "pl.kupowac.buy", "pl.commerce.buy_from", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "seller", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
}

func polishCommerceFullPurchaseBindings() []constructionBindingDeclaration {
	return []constructionBindingDeclaration{
		{"pl.kupowac.buy", "pl.kupowac.buy.from_seller", "pl.commerce.buy_from", "pl.commerce.buy_from.svoc.od"},
		{"pl.kupowac.buy", "pl.kupowac.buy.from_seller", "pl.commerce.buy_from", "pl.commerce.buy_from.svoc.od.period"},
	}
}

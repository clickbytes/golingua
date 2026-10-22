package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var exchangeActionFrames = []frameDeclaration{
	{"pl.zamykac_na_klucz.lock.action", "pl.zamykac_na_klucz.lock", "pl.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.skladac_z_czesci.assemble.action", "pl.skladac_z_czesci.assemble", "pl.transitive", []meaning.Role{"assembler", "product"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "assembler", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "product", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.targowac_sie.bargain.action", "pl.targowac_sie.bargain", "pl.transitive_oblique", []meaning.Role{"buyer_or_seller", "counterparty", "goods"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer_or_seller", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "counterparty", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "goods", Required: true, GrammaticalFunction: "complement", Case: "accusative", Adposition: "o"},
	}},
	{"pl.dawac_w_prezencie.gift.action", "pl.dawac_w_prezencie.gift", "pl.ditransitive", []meaning.Role{"giver", "recipient", "gift"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "gift", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dzielic_sie.share.action", "pl.dzielic_sie.share", "pl.transitive_oblique", []meaning.Role{"owner", "user", "resource"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "owner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "user", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "resource", Required: true, GrammaticalFunction: "complement", Case: "instrumental"},
	}},
	{"pl.uczyc_sie.acquire_knowledge.action", "pl.uczyc_sie.acquire_knowledge", "pl.oblique", []meaning.Role{"learner", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "learner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "content", Required: true, GrammaticalFunction: "oblique", Case: "genitive"},
	}},
}

package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var transferCognitionFrames = []frameDeclaration{
	{"pl.ogladac.view.action", "pl.ogladac.view", "pl.transitive", []meaning.Role{"viewer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "viewer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.ogladac_ekran.watch.action", "pl.ogladac_ekran.watch", "pl.transitive", []meaning.Role{"viewer", "display"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "viewer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "display", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wyobrazac_sobie.imagine.action", "pl.wyobrazac_sobie.imagine", "pl.transitive", []meaning.Role{"imaginer", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "imaginer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "content", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.miec_nadzieje.hope.action", "pl.miec_nadzieje.hope", "pl.transitive_oblique", []meaning.Role{"experiencer", "desired_outcome"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "desired_outcome", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.dostawac_zaplate.receive.action", "pl.dostawac_zaplate.receive", "pl.transitive_oblique", []meaning.Role{"payee", "payer", "value"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "payee", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "value", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "payer", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.zwracac_pieniadze.refund.action", "pl.zwracac_pieniadze.refund", "pl.ditransitive", []meaning.Role{"payer", "recipient", "money"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "payer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "money", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.przelewac_pieniadze.transfer.action", "pl.przelewac_pieniadze.transfer", "pl.ditransitive", []meaning.Role{"source", "recipient", "money"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "source", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "money", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

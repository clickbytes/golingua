package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var communicationMoreFrames = []frameDeclaration{
	{"pl.stwierdzac.state.action", "pl.stwierdzac.state", "pl.ditransitive", []meaning.Role{"speaker", "recipient", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: false, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "content", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.nie_zgadzac_sie.disagree.action", "pl.nie_zgadzac_sie.disagree", "pl.transitive_oblique", []meaning.Role{"speaker", "other_party", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "other_party", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "content", Required: true, GrammaticalFunction: "complement", Case: "genitive", Adposition: "co_do"},
	}},
	{"pl.przyjmowac.accept_offer.action", "pl.przyjmowac.accept_offer", "pl.transitive_oblique", []meaning.Role{"acceptor", "offerer", "proposal"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "acceptor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "offerer", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
		{ID: "object", Role: "proposal", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wysylac.send_message.action", "pl.wysylac.send_message", "pl.ditransitive", []meaning.Role{"sender", "recipient", "message"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "message", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

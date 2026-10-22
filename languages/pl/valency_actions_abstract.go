package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// abstractActionFrames records selected Polish case and adposition patterns.
var abstractActionFrames = []frameDeclaration{
	{"pl.zgadzac_sie.agree.action", "pl.zgadzac_sie.agree", "pl.transitive_oblique", []meaning.Role{"speaker", "other_party", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "other_party", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "content", Required: true, GrammaticalFunction: "complement", Case: "genitive", Adposition: "co_do"},
	}},
	{"pl.odmawiac.refuse.action", "pl.odmawiac.refuse", "pl.ditransitive", []meaning.Role{"refuser", "other_party", "proposal"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "refuser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "other_party", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "proposal", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.dyskutowac.discuss.action", "pl.dyskutowac.discuss", "pl.transitive_oblique", []meaning.Role{"participant_a", "participant_b", "topic"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant_a", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "participant_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
		{ID: "complement", Role: "topic", Required: true, GrammaticalFunction: "complement", Case: "locative", Adposition: "o"},
	}},
	{"pl.uznawac.believe.action", "pl.uznawac.believe", "pl.transitive", []meaning.Role{"believer", "proposition"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "believer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "proposition", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dziedziczyc.inherit.action", "pl.dziedziczyc.inherit", "pl.transitive_oblique", []meaning.Role{"heir", "predecessor", "property"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "heir", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "property", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "predecessor", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "po"},
	}},
	{"pl.wzrastac.increase.action", "pl.wzrastac.increase", "pl.intransitive", []meaning.Role{"bearer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "bearer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.spadac.reduce.action", "pl.spadac.reduce", "pl.intransitive", []meaning.Role{"bearer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "bearer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}

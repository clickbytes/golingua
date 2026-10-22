package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// socialActionFrames declares only role signatures actually present in CORE and
// independently reviewed Polish case/adposition choices for these candidates.
var socialActionFrames = []frameDeclaration{
	{"pl.witac.greet.action", "pl.witac.greet", "pl.transitive", []meaning.Role{"greeter", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "greeter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dziekowac.thank.action", "pl.dziekowac.thank", "pl.transitive", []meaning.Role{"thanker", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "thanker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.przepraszac.apologize.action", "pl.przepraszac.apologize", "pl.transitive", []meaning.Role{"speaker", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.doradzac.advise.action", "pl.doradzac.advise", "pl.transitive", []meaning.Role{"adviser", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "adviser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.przekonywac.persuade.action", "pl.przekonywac.persuade", "pl.transitive", []meaning.Role{"persuader", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "persuader", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.chwalic.praise.action", "pl.chwalic.praise", "pl.transitive", []meaning.Role{"speaker", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.krytykowac.criticize.action", "pl.krytykowac.criticize", "pl.transitive", []meaning.Role{"critic", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "critic", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.szeptac.whisper.action", "pl.szeptac.whisper", "pl.transitive", []meaning.Role{"speaker", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.dzwonic.call.action", "pl.dzwonic.call", "pl.oblique", []meaning.Role{"caller", "callee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "caller", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "callee", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.poznawac.meet.action", "pl.poznawac.meet", "pl.transitive", []meaning.Role{"participant_a", "participant_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant_a", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "participant_b", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pomagac.help.action", "pl.pomagac.help", "pl.transitive", []meaning.Role{"helper", "beneficiary"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "helper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "beneficiary", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.kochac.love.action", "pl.kochac.love", "pl.transitive", []meaning.Role{"experiencer", "person"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "person", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.lubic.like.action", "pl.lubic.like", "pl.transitive", []meaning.Role{"experiencer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.bac_sie.fear.action", "pl.bac_sie.fear", "pl.transitive", []meaning.Role{"experiencer", "threat"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "threat", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.martwic_sie.worry.action", "pl.martwic_sie.worry", "pl.oblique", []meaning.Role{"experiencer", "concern"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "concern", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "o"},
	}},
	{"pl.potrzebowac.need.action", "pl.potrzebowac.need", "pl.transitive", []meaning.Role{"experiencer", "requirement"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "requirement", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.znajdowac.find.action", "pl.znajdowac.find", "pl.transitive", []meaning.Role{"finder", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "finder", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.gubic.lose.action", "pl.gubic.lose", "pl.transitive", []meaning.Role{"holder", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.szukac.search.action", "pl.szukac.search", "pl.transitive", []meaning.Role{"searcher", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "searcher", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.rezerwowac.reserve.action", "pl.rezerwowac.reserve", "pl.transitive", []meaning.Role{"requester", "resource"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "requester", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "resource", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.anulowac.cancel.action", "pl.anulowac.cancel", "pl.transitive", []meaning.Role{"requester", "reservation"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "requester", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "reservation", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.oszczedzac.save.action", "pl.oszczedzac.save", "pl.transitive", []meaning.Role{"saver", "money"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "saver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "money", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wydawac.spend.action", "pl.wydawac.spend", "pl.transitive", []meaning.Role{"spender", "money"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "spender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "money", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zarabiac.earn.action", "pl.zarabiac.earn", "pl.transitive", []meaning.Role{"earner", "money"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "earner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "money", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

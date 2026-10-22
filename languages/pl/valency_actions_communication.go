package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// communicationActionFrames records the selected Polish cases/adpositions for
// the CORE roles; these frames do not approve construction bindings.
var communicationActionFrames = []frameDeclaration{
	{"pl.informowac.inform.action", "pl.informowac.inform", "pl.transitive_oblique", []meaning.Role{"informant", "recipient", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "informant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "content", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "o"},
	}},
	{"pl.wyjasniac.explain.action", "pl.wyjasniac.explain", "pl.ditransitive", []meaning.Role{"explainer", "recipient", "subject"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "explainer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "subject", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.opisywac.describe.action", "pl.opisywac.describe", "pl.ditransitive", []meaning.Role{"speaker", "recipient", "referent"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "referent", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.przedstawiac.introduce.action", "pl.przedstawiac.introduce", "pl.ditransitive", []meaning.Role{"introducer", "person_a", "person_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "introducer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "person_a", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "person_b", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.prosic.request.action", "pl.prosic.request", "pl.transitive_oblique", []meaning.Role{"requester", "addressee", "requested"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "requester", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "addressee", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "requested", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "o"},
	}},
	{"pl.zapraszac.invite.action", "pl.zapraszac.invite", "pl.transitive_oblique", []meaning.Role{"inviter", "invitee", "occasion"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "inviter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "invitee", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "occasion", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.oferowac.offer.action", "pl.oferowac.offer", "pl.ditransitive", []meaning.Role{"offerer", "recipient", "proposed"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "offerer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "proposed", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.obiecywac.promise.action", "pl.obiecywac.promise", "pl.ditransitive", []meaning.Role{"promiser", "addressee", "commitment"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "promiser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "commitment", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "addressee", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
}

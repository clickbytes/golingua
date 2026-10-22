package pl

// communicationMoreEntries adds pending assertion, disagreement and offer
// acceptance senses with explicit role-bearing valency frames.
var communicationMoreEntries = []lexicalEntry{
	{"pl.stwierdzac.v", "stwierdzać", "verb", []lexicalMeaning{{"pl.stwierdzac.state", "Expressing a proposition as an assertion of fact or opinion.", "stating"}}},
	{"pl.nie_zgadzac_sie.v", "nie zgadzać się", "verb_reflexive", []lexicalMeaning{{"pl.nie_zgadzac_sie.disagree", "Expressing rejection of an assertion or position without necessarily refusing action.", "disagreeing"}}},
	{"pl.przyjmowac.v", "przyjmować", "verb", []lexicalMeaning{{"pl.przyjmowac.accept_offer", "Expressing consent to terms proposed by another participant.", "accepting_offer"}}},
}

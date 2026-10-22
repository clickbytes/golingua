package pl

// abstractActionEntries adds pending candidates for selected discourse,
// cognition and quantity-change events without inferring morphology.
var abstractActionEntries = []lexicalEntry{
	{"pl.zgadzac_sie.v", "zgadzać się", "verb_reflexive", []lexicalMeaning{{"pl.zgadzac_sie.agree", "Expressing acceptance of a proposal, assertion, or position.", "agreeing"}}},
	{"pl.odmawiac.v", "odmawiać", "verb", []lexicalMeaning{{"pl.odmawiac.refuse", "Declining a request, offer, or proposed course of action.", "refusing"}}},
	{"pl.dyskutowac.v", "dyskutować", "verb", []lexicalMeaning{{"pl.dyskutowac.discuss", "Exchanging views or information about a shared topic.", "discussing"}}},
	{"pl.uznawac.v", "uznawać", "verb", []lexicalMeaning{{"pl.uznawac.believe", "Adopting a proposition as likely or true without establishing its actual truth.", "forming_belief"}}},
	{"pl.dziedziczyc.v", "dziedziczyć", "verb", []lexicalMeaning{{"pl.dziedziczyc.inherit", "Receiving an asset because of a predecessor's death under inheritance rules.", "inheriting_property"}}},
	{"pl.wzrastac.v", "wzrastać", "verb", []lexicalMeaning{{"pl.wzrastac.increase", "A measurable magnitude becoming larger between two observations.", "increasing_quantity"}}},
}

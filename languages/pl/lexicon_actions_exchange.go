package pl

// exchangeActionEntries adds pending physical, transfer and negotiation senses.
var exchangeActionEntries = []lexicalEntry{
	{"pl.zamykac_na_klucz.v", "zamykać na klucz", "verb", []lexicalMeaning{{"pl.zamykac_na_klucz.lock", "Engaging a mechanism that prevents opening or unauthorized access.", "locking_object"}}},
	{"pl.skladac_z_czesci.v", "składać z części", "verb", []lexicalMeaning{{"pl.skladac_z_czesci.assemble", "Joining separate components to produce a larger structured artifact.", "assembling_object"}}},
	{"pl.targowac_sie.v", "targować się", "verb_reflexive", []lexicalMeaning{{"pl.targowac_sie.bargain", "Negotiating transaction terms with the aim of changing the proposed price.", "bargaining"}}},
	{"pl.dawac_w_prezencie.v", "dawać w prezencie", "verb", []lexicalMeaning{{"pl.dawac_w_prezencie.gift", "Transferring something to a recipient as a gift without expected payment.", "gifting"}}},
	{"pl.dzielic_sie.v", "dzielić się", "verb_reflexive", []lexicalMeaning{{"pl.dzielic_sie.share", "Allowing multiple participants to access or use a resource under an arrangement.", "sharing_resource"}}},
}

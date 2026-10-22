package pl

// outcomeActionEntries adds command and outcome candidates with explicit senses.
var outcomeActionEntries = []lexicalEntry{
	{"pl.rozkazywac.v", "rozkazywać", "verb", []lexicalMeaning{{"pl.rozkazywac.command", "Directing another participant to act with asserted authority.", "commanding"}}},
	{"pl.udac_sie.v", "udać się", "verb_reflexive", []lexicalMeaning{{"pl.udac_sie.succeed", "An attempted action achieving its specified objective.", "succeeding"}}},
	{"pl.nie_udac_sie.v", "nie udać się", "verb_reflexive", []lexicalMeaning{{"pl.nie_udac_sie.fail", "An attempted action ending without achieving its specified objective.", "failing"}}},
}

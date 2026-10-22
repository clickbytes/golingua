package pl

// activityActionEntries adds pending Polish candidates with explicit valency.
// No morphology or construction is inferred.
var activityActionEntries = []lexicalEntry{
	{"pl.porownywac.v", "porównywać", "verb", []lexicalMeaning{{"pl.porownywac.compare", "Examining at least two items in terms of their similarities or differences.", "comparing"}}},
	{"pl.ranic.v", "ranić", "verb", []lexicalMeaning{{"pl.ranic.injure", "Causing bodily tissue damage to a living participant.", "injuring_person"}}},
	{"pl.sadzic.v", "sadzić", "verb", []lexicalMeaning{{"pl.sadzic.plant", "Placing a seed in growing medium to initiate cultivation.", "planting_seed"}}},
	{"pl.ostrzegac.v", "ostrzegać", "verb", []lexicalMeaning{{"pl.ostrzegac.warn", "Alerting a recipient to a risk or unfavorable possible occurrence.", "warning"}}},
	{"pl.polecac.v", "polecać", "verb", []lexicalMeaning{{"pl.polecac.recommend", "Expressing positive preference for an option as suitable for a goal.", "recommending"}}},
	{"pl.zaczynac.v", "zaczynać", "verb", []lexicalMeaning{{"pl.zaczynac.start", "An activity passing from not occurring into an occurring state.", "starting_activity"}}},
	{"pl.konczyc.v", "kończyć", "verb", []lexicalMeaning{{"pl.konczyc.finish", "An activity reaching its intended or defined terminal boundary.", "finishing_activity"}}},
	{"pl.kontynuowac.v", "kontynuować", "verb", []lexicalMeaning{{"pl.kontynuowac.continue", "An activity remaining in progress across a later time segment.", "continuing_activity"}}},
}

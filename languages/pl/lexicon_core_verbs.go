package pl

// coreVerbsEntries declares the reviewed Polish verbs protected by explicit valency frames in valency.go.
// Every entry is an authored editorial decision recorded in
// data/pl/1000-core-review.tsv; senses link to CORE as pending candidates
// without transfer authority. No paradigm is asserted for new lexemes yet.
var coreVerbsEntries = []lexicalEntry{
	{"pl.pytac.v", "pytać", "verb", []lexicalMeaning{{"pl.pytac.request_info", "A communicative act requesting information from an addressee", "asking"}}},
	{"pl.odpowiadac.v", "odpowiadać", "verb", []lexicalMeaning{{"pl.odpowiadac.answer", "A communicative act supplying information in response to a prior request", "answering"}}},
	{"pl.czytac.v", "czytać", "verb", []lexicalMeaning{{"pl.czytac.interpret", "An event of interpreting written or otherwise recorded symbolic content", "reading"}}},
	{"pl.uczyc_sie.v", "uczyć się", "verb_reflexive", []lexicalMeaning{{"pl.uczyc_sie.acquire", "An event in which a participant acquires knowledge or skill", "learning"}, {"pl.uczyc_sie.acquire_knowledge", "Coming to know a fact or relationship through evidence or instruction.", "acquiring_knowledge"}}},
	{"pl.uczyc.v", "uczyć", "verb", []lexicalMeaning{{"pl.uczyc.instruct", "An event in which an instructor intentionally supports another's learning", "teaching"}}},
	{"pl.jesc.v", "jeść", "verb", []lexicalMeaning{{"pl.jesc.eat", "An event of taking food into a body for nourishment", "eating"}}},
	{"pl.spac.v", "spać", "verb", []lexicalMeaning{{"pl.spac.sleep", "A physiological period of sleep undergone by a living participant", "sleeping"}}},
}

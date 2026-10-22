package pl

// remainingActionEntries2 adds pending lending, ordering, desire and event-noun
// candidates without asserting verb morphology.
var remainingActionEntries2 = []lexicalEntry{
	{"pl.pozyczac_komus.v", "pożyczać komuś", "verb", []lexicalMeaning{{"pl.pozyczac_komus.lend", "Temporarily allowing another participant to use an item with an expectation of return.", "lending_object"}}},
	{"pl.zamawiac_towary.v", "zamawiać towary", "verb", []lexicalMeaning{{"pl.zamawiac_towary.order", "Submitting a request for specified goods or services to a supplier.", "ordering_goods"}}},
	{"pl.chciec.v", "chcieć", "verb", []lexicalMeaning{{"pl.chciec.desire", "Experiencing a directed want for an outcome, object, or state.", "desiring"}}},
}

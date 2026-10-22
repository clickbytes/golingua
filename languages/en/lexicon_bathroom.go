package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishBathroomSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_bathroom.go",
	Version: "2026-10-22-batch-31", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name bathroom fixtures and cleaning supplies.
var englishBathroomEntries = []lexicalEntry{
	{"en.toilet_fixture.n", "toilet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toilet_fixture.direct", "a sanitary fixture used to dispose of bodily waste into a drainage system", "toilet_fixture"}}},
	{"en.shower_fixture.n", "shower", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shower_fixture.direct", "a water outlet and fittings designed for washing the body by spray", "shower_fixture"}}},
	{"en.bathtub.n", "bathtub", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bathtub.direct", "a large fixed or portable basin intended for bathing by immersion", "bathtub"}}},
	{"en.sink_fixture.n", "sink", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sink_fixture.direct", "a fixed basin with a drain for washing hands, objects, or food", "sink_fixture"}}},
	{"en.faucet.n", "faucet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.faucet.direct", "a valve-operated outlet controlling water flow from a supply pipe", "faucet"}}},
	{"en.toothpaste.n", "toothpaste", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toothpaste.direct", "a preparation applied to a toothbrush to help clean teeth", "toothpaste"}}},
	{"en.shampoo.n", "shampoo", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shampoo.direct", "a liquid or gel cleansing preparation intended for hair", "shampoo"}}},
	{"en.trash_can.n", "trash can", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.trash_can.direct", "a container for temporary collection of discarded waste", "trash_can"}}},
	{"en.cleaning_sponge.n", "sponge", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cleaning_sponge.direct", "a porous absorbent cleaning implement used to wipe or scrub surfaces", "cleaning_sponge"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishSettlementSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_settlements.go",
	Version: "2026-10-22-batch-21", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish populated places, territory and structures.
var englishSettlementEntries = []lexicalEntry{
	{"en.settlement.n", "settlement", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.settlement.direct", "a locality where people live in a relatively persistent concentration", "settlement"}}},
	{"en.village.n", "village", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.village.direct", "a small human settlement generally less urbanized than a town", "village"}}},
	{"en.town.n", "town", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.town.direct", "a settled urban locality smaller in scale or status than a major city", "town"}}},
	{"en.city.n", "city", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.city.direct", "a large or administratively recognized urban settlement", "city"}}},
	{"en.country_territory.n", "territory", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.country_territory.direct", "the geographic territory associated with a recognized country, not its government", "country_territory"}}},
	{"en.region_area.n", "region", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.region_area.direct", "a bounded or conventionally identified portion of a larger territory", "region_area"}}},
	{"en.neighborhood_area.n", "neighborhood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.neighborhood_area.direct", "a local district within a settlement where people live or work", "neighborhood_area"}}},
	{"en.address_location.n", "address", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.address_location.direct", "a location identified by an address system rather than a named physical structure", "address_location"}}},
	{"en.building.n", "building", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.building.direct", "a constructed structure with a roof or enclosure designed for human use", "building"}}},
	{"en.apartment.n", "apartment", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.apartment.direct", "a self-contained residential unit within a larger building", "apartment"}}},
	{"en.room.n", "room", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.room.direct", "a bounded interior space within a building, typically divided by walls", "room"}}},
}

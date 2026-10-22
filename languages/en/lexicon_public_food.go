package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPublicFoodSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_public_food.go",
	Version: "2026-10-22-batch-31", License: "unknown", Author: "golingua contributors",
}

// These pending nouns distinguish foods from public service premises.
var englishPublicFoodEntries = []lexicalEntry{
	{"en.pizza_food.n", "pizza", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pizza_food.direct", "a baked flat dough base topped with sauce, cheese, or other ingredients", "pizza_food"}}},
	{"en.ice_cream_food.n", "ice cream", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ice_cream_food.direct", "a frozen sweet dessert typically made from milk or comparable ingredients", "ice_cream_food"}}},
	{"en.chocolate_food.n", "chocolate", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.chocolate_food.direct", "a sweet or bitter food preparation derived mainly from cocoa solids or butter", "chocolate_food"}}},
	{"en.honey_food.n", "honey", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.honey_food.direct", "a viscous sweet substance produced by bees from plant nectar or secretions", "honey_food"}}},
	{"en.grape_fruit_food.n", "grape", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.grape_fruit_food.direct", "a small edible grape berry, distinct from grapefruit", "grape_fruit_food"}}},
	{"en.gas_station.n", "gas station", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.gas_station.direct", "a roadside service facility where motor vehicles can obtain fuel", "gas_station"}}},
	{"en.bank_building.n", "bank", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bank_building.direct", "premises used by a banking institution to serve customers", "bank_building"}}},
	{"en.pharmacy_building.n", "pharmacy", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pharmacy_building.direct", "premises where medicines are dispensed or sold under pharmacy practice", "pharmacy_building"}}},
}

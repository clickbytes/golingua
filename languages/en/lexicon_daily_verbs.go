package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishDailyVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_daily_verbs.go",
	Version: "2026-10-22-batch-32", License: "unknown", Author: "golingua contributors",
}

// These pending verbs name common food, home and routine activities.
var englishDailyVerbEntries = []lexicalEntry{
	{"en.printing_document.n", "print", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.printing_document.direct", "produce a physical copy of document content with printing equipment", "printing_document"}}},
	{"en.cooking_food.n", "cook", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cooking_food.direct", "prepare food by combining ingredients and applying kitchen processes", "cooking_food"}}},
	{"en.baking_food.n", "bake", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.baking_food.direct", "cook food using sustained dry heat in an enclosed oven", "baking_food"}}},
	{"en.boiling_food.n", "boil", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.boiling_food.direct", "cook food by immersing it in boiling liquid", "boiling_food"}}},
	{"en.frying_food.n", "fry", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.frying_food.direct", "cook food using hot fat or oil in contact with its surface", "frying_food"}}},
	{"en.shopping.n", "shop", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shopping.direct", "visit or use sellers to select items for potential or actual purchase", "shopping"}}},
	{"en.cleaning_home.n", "clean", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cleaning_home.direct", "remove dirt and restore order or cleanliness in a dwelling", "cleaning_home"}}},
	{"en.dressing_self.n", "dress", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.dressing_self.direct", "put clothing onto one’s own body", "dressing_self"}}},
	{"en.brushing_teeth.n", "brush", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.brushing_teeth.direct", "clean teeth by rubbing them with a brush", "brushing_teeth"}}},
	{"en.bathing.n", "bathe", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bathing.direct", "wash the body by immersion in water or a bath-like vessel", "bathing"}}},
	{"en.planting_seed.n", "plant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.planting_seed.direct", "place a seed in growing medium to initiate cultivation", "planting_seed"}}},
	{"en.watering_plant.n", "water", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.watering_plant.direct", "supply water to a plant or its growing medium", "watering_plant"}}},
	{"en.harvesting_crop.n", "harvest", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.harvesting_crop.direct", "collect cultivated plants or plant products when ready for use", "harvesting_crop"}}},
	{"en.playing_game.n", "play", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.playing_game.direct", "take part in a rule-governed activity for recreation or competition", "playing_game"}}},
}

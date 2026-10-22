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
	{"en.printing_document.v", "print", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.printing_document.event", "produce a physical copy of document content with printing equipment", "printing_document"}}},
	{"en.cooking_food.v", "cook", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.cooking_food.event", "prepare food by combining ingredients and applying kitchen processes", "cooking_food"}}},
	{"en.baking_food.v", "bake", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.baking_food.event", "cook food using sustained dry heat in an enclosed oven", "baking_food"}}},
	{"en.boiling_food.v", "boil", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.boiling_food.event", "cook food by immersing it in boiling liquid", "boiling_food"}}},
	{"en.frying_food.v", "fry", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.frying_food.event", "cook food using hot fat or oil in contact with its surface", "frying_food"}}},
	{"en.shopping.v", "shop", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.shopping.event", "visit or use sellers to select items for potential or actual purchase", "shopping"}}},
	{"en.cleaning_home.v", "clean", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.cleaning_home.event", "remove dirt and restore order or cleanliness in a dwelling", "cleaning_home"}}},
	{"en.dressing_self.v", "dress", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.dressing_self.event", "put clothing onto one’s own body", "dressing_self"}}},
	{"en.brushing_teeth.v", "brush", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.brushing_teeth.event", "clean teeth by rubbing them with a brush", "brushing_teeth"}}},
	{"en.bathing.v", "bathe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.bathing.event", "wash the body by immersion in water or a bath-like vessel", "bathing"}}},
	{"en.planting_seed.v", "plant", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.planting_seed.event", "place a seed in growing medium to initiate cultivation", "planting_seed"}}},
	{"en.watering_plant.v", "water", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.watering_plant.event", "supply water to a plant or its growing medium", "watering_plant"}}},
	{"en.harvesting_crop.v", "harvest", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.harvesting_crop.event", "collect cultivated plants or plant products when ready for use", "harvesting_crop"}}},
	{"en.playing_game.v", "play", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.playing_game.event", "take part in a rule-governed activity for recreation or competition", "playing_game"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishFoodSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_food.go",
	Version: "2026-10-22-batch-20", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name common foods, fruits, ingredients and meals.
var englishFoodEntries = []lexicalEntry{
	{"en.beverage.n", "beverage", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.beverage.direct", "a consumable liquid intended primarily for drinking", "beverage"}}},
	{"en.meal.n", "meal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.meal.direct", "a socially organized portion or occasion of eating food, distinct from the act of eating", "meal"}}},
	{"en.breakfast_meal.n", "breakfast", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.breakfast_meal.direct", "a meal conventionally taken near the start of a person’s waking day", "breakfast_meal"}}},
	{"en.lunch_meal.n", "lunch", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lunch_meal.direct", "a meal conventionally eaten around the middle of a day", "lunch_meal"}}},
	{"en.dinner_meal.n", "dinner", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.dinner_meal.direct", "a substantial meal conventionally eaten later in the day", "dinner_meal"}}},
	{"en.bread.n", "bread", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bread.direct", "a baked staple food made principally from flour-and-water dough", "bread"}}},
	{"en.rice_food.n", "rice", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rice_food.direct", "edible rice grains typically prepared as a staple dish or ingredient", "rice_food"}}},
	{"en.pasta_food.n", "pasta", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pasta_food.direct", "shaped dough food commonly cooked in boiling water", "pasta_food"}}},
	{"en.potato_food.n", "potato", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.potato_food.direct", "an edible potato tuber considered as food", "potato_food"}}},
	{"en.vegetable_food.n", "vegetable", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.vegetable_food.direct", "an edible plant part conventionally treated as a vegetable in cooking", "vegetable_food"}}},
	{"en.fruit_food.n", "fruit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fruit_food.direct", "an edible sweet or tart plant product treated as fruit in food contexts", "fruit_food"}}},
	{"en.apple_fruit.n", "apple", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.apple_fruit.direct", "an edible pome fruit produced by an apple tree", "apple_fruit"}}},
	{"en.banana_fruit.n", "banana", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.banana_fruit.direct", "an elongated edible fruit of a banana plant", "banana_fruit"}}},
	{"en.orange_fruit.n", "orange", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.orange_fruit.direct", "an edible citrus fruit with a segmented interior and orange peel", "orange_fruit"}}},
	{"en.lemon_fruit.n", "lemon", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lemon_fruit.direct", "an acidic yellow citrus fruit commonly used for juice or flavor", "lemon_fruit"}}},
	{"en.tomato_food.n", "tomato", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tomato_food.direct", "the edible tomato fruit used as an ingredient in dishes", "tomato_food"}}},
	{"en.carrot_food.n", "carrot", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.carrot_food.direct", "the edible taproot of a cultivated carrot plant", "carrot_food"}}},
	{"en.onion_food.n", "onion", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.onion_food.direct", "an edible layered bulb of the onion plant", "onion_food"}}},
	{"en.lettuce_food.n", "lettuce", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lettuce_food.direct", "edible leafy material from cultivated lettuce plants", "lettuce_food"}}},
	{"en.meat_food.n", "meat", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.meat_food.direct", "edible animal muscle or associated tissue prepared or intended for consumption", "meat_food"}}},
	{"en.beef_food.n", "beef", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.beef_food.direct", "meat obtained from cattle", "beef_food"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishFoodMoreSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_food_more.go",
	Version: "2026-10-22-batch-21", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name additional foods, ingredients and beverages.
var englishFoodMoreEntries = []lexicalEntry{
	{"en.pork_food.n", "pork", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pork_food.direct", "meat obtained from pigs", "pork_food"}}},
	{"en.chicken_meat.n", "chicken", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.chicken_meat.direct", "meat obtained from a chicken, distinct from the live bird", "chicken_meat"}}},
	{"en.fish_meat.n", "fish", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fish_meat.direct", "fish flesh considered as food, distinct from the live fish", "fish_meat"}}},
	{"en.egg_food.n", "egg", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.egg_food.direct", "a bird egg intended or prepared for human consumption", "egg_food"}}},
	{"en.milk_food.n", "milk", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.milk_food.direct", "a nutrient-rich mammalian secretion consumed as a drink or ingredient", "milk_food"}}},
	{"en.cheese_food.n", "cheese", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cheese_food.direct", "a food made by coagulating and processing milk proteins and fats", "cheese_food"}}},
	{"en.yogurt_food.n", "yogurt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.yogurt_food.direct", "a fermented milk food with thick or drinkable consistency", "yogurt_food"}}},
	{"en.butter_food.n", "butter", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.butter_food.direct", "a concentrated milk-fat spread or cooking ingredient", "butter_food"}}},
	{"en.soup_food.n", "soup", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.soup_food.direct", "a liquid-dominant prepared dish containing stock and other ingredients", "soup_food"}}},
	{"en.salad_food.n", "salad", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.salad_food.direct", "a prepared dish of mixed raw or cooked ingredients, commonly served cold", "salad_food"}}},
	{"en.sandwich_food.n", "sandwich", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sandwich_food.direct", "a prepared food portion with filling held between or within bread", "sandwich_food"}}},
	{"en.cake_food.n", "cake", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cake_food.direct", "a sweet baked food made from a structured batter or dough", "cake_food"}}},
	{"en.cookie_food.n", "cookie", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cookie_food.direct", "a small sweet baked portion intended to be eaten individually", "cookie_food"}}},
	{"en.sugar_food.n", "sugar", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sugar_food.direct", "a sweet crystalline carbohydrate used as food or an ingredient", "sugar_food"}}},
	{"en.salt_food.n", "salt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.salt_food.direct", "edible sodium chloride or similar salt used for seasoning or preservation", "salt_food"}}},
	{"en.pepper_spice.n", "pepper", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pepper_spice.direct", "a pungent seasoning obtained from pepper plants", "pepper_spice"}}},
	{"en.coffee_beverage.n", "coffee", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coffee_beverage.direct", "a beverage prepared by extracting roasted ground coffee beans with water", "coffee_beverage"}}},
	{"en.tea_beverage.n", "tea", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tea_beverage.direct", "a beverage prepared by infusing tea leaves or other plant material in water", "tea_beverage"}}},
	{"en.juice_beverage.n", "juice", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.juice_beverage.direct", "a drink consisting principally of extracted fruit or vegetable liquid", "juice_beverage"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishAnimalSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_animals.go",
	Version: "2026-10-22-batch-15", License: "unknown", Author: "golingua contributors",
}

// These candidates distinguish animal classes and selected domestic species.
// They declare no morphology, biological taxonomy, or food-use senses.
var englishAnimalEntries = []lexicalEntry{
	{"en.mammal.n", "mammal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mammal.class", "a vertebrate of the mammalian class, typically nursing its young with milk", "mammal"}}},
	{"en.bird.n", "bird", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bird.class", "a feathered vertebrate of the bird class; flight ability is not required", "bird"}}},
	{"en.fish_animal.n", "fish", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fish_animal.aquatic_vertebrate", "an aquatic gill-breathing vertebrate, as an animal rather than prepared food", "fish_animal"}}},
	{"en.reptile.n", "reptile", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.reptile.class", "a vertebrate in a reptilian lineage, generally bearing dry scales", "reptile"}}},
	{"en.amphibian.n", "amphibian", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.amphibian.class", "a vertebrate of the amphibian class, often using both aquatic and terrestrial habitats", "amphibian"}}},
	{"en.insect.n", "insect", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.insect.six_legged_arthropod", "an arthropod with a six-legged adult body plan", "insect"}}},
	{"en.spider.n", "spider", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.spider.arachnid", "an arachnid with eight walking legs, distinct from insects", "spider"}}},
	{"en.worm_animal.n", "worm", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.worm_animal.soft_invertebrate", "an elongated soft-bodied invertebrate, without assuming one taxonomic family", "worm_animal"}}},
	{"en.horse.n", "horse", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.horse.domestic_equine", "an individual domestic horse belonging to the equine species", "horse"}}},
	{"en.cow.n", "cow", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cow.adult_female_bovine", "an adult female of domestic cattle, not cattle in general", "cow"}}},
	{"en.pig.n", "pig", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pig.domestic_animal", "an individual domestic pig raised or kept by people", "pig"}}},
	{"en.sheep.n", "sheep", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sheep.domestic_animal", "an individual domestic sheep of either sex", "sheep"}}},
	{"en.rabbit.n", "rabbit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rabbit.lagomorph", "a small lagomorph of the rabbit type, distinct from a hare", "rabbit"}}},
	{"en.mouse_animal.n", "mouse", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mouse_animal.small_rodent", "a small rodent of a mouse type, not a computer device", "mouse_animal"}}},
	{"en.rat.n", "rat", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rat.rodent", "a rodent of a rat type, distinct from a mouse", "rat"}}},
	{"en.bear.n", "bear", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bear.ursid", "a large omnivorous or carnivorous mammal of the bear family", "bear"}}},
	{"en.monkey.n", "monkey", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.monkey.non_ape_primate", "a non-ape simian primate, without restricting to one species", "monkey"}}},
	{"en.chicken_animal.n", "chicken", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.chicken_animal.domestic_bird", "an individual domestic chicken, rather than its meat", "chicken_animal"}}},
	{"en.duck.n", "duck", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.duck.waterfowl", "a waterfowl bird of the duck type", "duck"}}},
	{"en.goose.n", "goose", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.goose.waterfowl", "a large waterfowl bird of the goose type", "goose"}}},
	{"en.frog.n", "frog", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.frog.amphibian", "a usually tailless adult amphibian of the frog group", "frog"}}},
	{"en.snake.n", "snake", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.snake.limbless_reptile", "an elongated limbless reptile of the snake lineage", "snake"}}},
	{"en.lizard.n", "lizard", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lizard.squamate", "a typically four-legged squamate reptile, distinct from snakes", "lizard"}}},
	{"en.turtle.n", "turtle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.turtle.shelled_reptile", "a reptile with a protective bony shell of the turtle group", "turtle"}}},
}

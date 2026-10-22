package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishNounSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_nouns.go",
	Version: "2026-10-22-batch-01", License: "unknown", Author: "golingua contributors",
}

// These clear nominal candidates are independent of the Polish lexical layer.
// They carry no inferred morphology, valency or construction binding.
var englishNounEntries = []lexicalEntry{
	{"en.person.n", "person", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.person.human", "a human individual without a specified age or social role", "person"}}},
	{"en.adult.n", "adult", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.adult.person", "a person past the child stage, without asserting a legal age threshold", "adult"}}},
	{"en.animal.n", "animal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.animal.organism", "a living nonhuman organism capable of animal life processes", "animal"}}},
	{"en.plant.n", "plant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.plant.organism", "a living organism of the plant kingdom", "plant"}}},
	{"en.tree.n", "tree", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tree.woody_plant", "a woody perennial plant with a persistent main stem", "tree"}}},
	{"en.artifact.n", "artifact", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.artifact.made_object", "a physical object intentionally made or modified by an agent", "artifact"}}},
	{"en.door.n", "door", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.door.barrier", "a movable barrier used to open or close an entrance", "door"}}},
	{"en.vehicle.n", "vehicle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.vehicle.conveyance", "an artifact designed to carry people or goods between places", "vehicle"}}},
	{"en.location.n", "location", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.location.place", "a spatial region that can serve as the site of an event or entity", "location"}}},
	{"en.body_part.n", "body part", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.body_part.anatomical", "a physical component of an organism's body", "body_part"}}},
	{"en.event.n", "event", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.event.occurrence", "an occurrence with participants or a change that can be located in time", "event"}}},
}

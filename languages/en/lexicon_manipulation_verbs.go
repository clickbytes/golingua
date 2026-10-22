package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishManipulationVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_manipulation_verbs.go",
	Version: "2026-10-22-batch-25", License: "unknown", Author: "golingua contributors",
}

// These pending verb candidates name deliberate physical actions.
var englishManipulationVerbEntries = []lexicalEntry{
	{"en.intentional_action.v", "act", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.intentional_action.event", "undertake an event deliberately to produce an effect", "intentional_action"}}},
	{"en.manipulating_object.v", "manipulate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.manipulating_object.event", "intentionally control an object’s position, shape, or state by physical action", "manipulating_object"}}},
	{"en.touching_object.v", "touch", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.touching_object.event", "make physical surface contact with an object without requiring it to move", "touching_object"}}},
	{"en.grasping.v", "grasp", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.grasping.event", "close a hand or implement around an object to hold it securely", "grasping"}}},
	{"en.holding_object.v", "hold", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.holding_object.event", "maintain physical support or control of an object without releasing it", "holding_object"}}},
	{"en.releasing_object.v", "release", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.releasing_object.event", "end a maintained physical hold or restraint on an object", "releasing_object"}}},
	{"en.lifting_object.v", "lift", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.lifting_object.event", "move a supported object to a higher position", "lifting_object"}}},
	{"en.lowering_object.v", "lower", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.lowering_object.event", "deliberately move a supported object to a lower position", "lowering_object"}}},
	{"en.placing_object.v", "place", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.placing_object.event", "move an object to a target position and leave it there", "placing_object"}}},
	{"en.inserting_object.v", "insert", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.inserting_object.event", "move an object through an opening into a containing region", "inserting_object"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishMovementMoreSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_movement_more.go",
	Version: "2026-10-22-batch-33", License: "unknown", Author: "golingua contributors",
}

// These pending verbs cover waiting, staying, resting and navigation.
var englishMovementMoreEntries = []lexicalEntry{
	{"en.waiting.v", "wait", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.waiting.event", "remain ready or inactive until an anticipated condition or event occurs", "waiting"}}},
	{"en.staying_at_place.v", "stay", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.staying_at_place.event", "maintain presence at a location without an implied change of position", "staying_at_place"}}},
	{"en.resting.v", "rest", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.resting.event", "temporarily reduce activity to recover energy or comfort", "resting"}}},
	{"en.navigation.v", "navigate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.navigation.event", "choose or follow a route to reach a spatial destination", "navigation"}}},
}

package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishExperienceNounSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_experience_nouns.go",
	Version: "2026-10-22-batch-33", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name common bodily and experiential states.
var englishExperienceNounEntries = []lexicalEntry{
	{"en.hunger_state.n", "hunger", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hunger_state.direct", "a bodily condition signaling need or desire for food", "hunger_state"}}},
	{"en.thirst_state.n", "thirst", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.thirst_state.direct", "a bodily condition signaling need or desire for fluids", "thirst_state"}}},
	{"en.fatigue_state.n", "fatigue", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fatigue_state.direct", "reduced capacity for sustained activity after exertion or insufficient recovery", "fatigue_state"}}},
	{"en.pain_sensation.n", "pain", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pain_sensation.direct", "an unpleasant sensory and affective experience associated with actual or potential bodily harm", "pain_sensation"}}},
}

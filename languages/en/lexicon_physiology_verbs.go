package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPhysiologyVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_physiology_verbs.go",
	Version: "2026-10-22-batch-27", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name basic respiratory actions.
var englishPhysiologyVerbEntries = []lexicalEntry{
	{"en.breathing.v", "breathe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.breathing.event", "repeatedly move respiratory gases into and out of a living organism", "breathing"}}},
	{"en.inhaling.v", "inhale", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.inhaling.event", "draw air or another respiratory gas into the respiratory system", "inhaling"}}},
	{"en.exhaling.v", "exhale", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.exhaling.event", "expel respiratory gas from the respiratory system", "exhaling"}}},
	{"en.coughing.v", "cough", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.coughing.event", "forcefully expel air to clear or respond to irritation in the airways", "coughing"}}},
	{"en.sneezing.v", "sneeze", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.sneezing.event", "involuntarily expel air through the nose and mouth in a sneeze reflex", "sneezing"}}},
	{"en.swallowing.v", "swallow", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.swallowing.event", "move a substance from the mouth through the throat toward the stomach", "swallowing"}}},
}

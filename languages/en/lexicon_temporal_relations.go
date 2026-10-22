package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishTemporalRelationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_temporal_relations.go",
	Version: "2026-10-22-batch-33", License: "unknown", Author: "golingua contributors",
}

// These pending entries express temporal overlap and interval boundaries.
var englishTemporalRelationEntries = []lexicalEntry{
	{"en.temporal_overlap.n", "overlap", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.temporal_overlap.direct", "a relation in which two temporal intervals share some, but not necessarily all, of their duration", "temporal_overlap"}}},
	{"en.occurs_during.prep", "during", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.occurs_during.relation", "an event takes place within the time span of another event or interval", "occurs_during"}}},
	{"en.interval_begins_with.prep", "at the beginning of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.interval_begins_with.relation", "an event or instant marks the starting boundary of a temporal interval", "interval_begins_with"}}},
	{"en.interval_ends_with.prep", "at the end of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.interval_ends_with.relation", "an event or instant marks the ending boundary of a temporal interval", "interval_ends_with"}}},
}

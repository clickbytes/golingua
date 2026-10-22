package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishSocialGroupsSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_social_groups.go",
	Version: "2026-10-22-batch-09", License: "unknown", Author: "golingua contributors",
}

// These people-group nouns distinguish social relation and gathering type.
// Candidate links remain pending and carry no inferred morphology or frames.
var englishSocialGroupEntries = []lexicalEntry{
	{"en.group_of_people.n", "group", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.group_of_people.collective", "a plurality of people treated as one socially relevant collective", "group_of_people"}}},
	{"en.family_group.n", "family", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.family_group.kinship", "a set of people linked by kinship or established family relationships", "family_group"}}},
	{"en.household.n", "household", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.household.domestic", "people sharing domestic living arrangements and ordinary household resources", "household"}}},
	{"en.community.n", "community", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.community.shared_identity", "people connected by common locality, practice, or shared social identity", "community"}}},
	{"en.crowd.n", "crowd", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.crowd.gathering", "a gathering of people physically together without requiring formal membership", "crowd"}}},
	{"en.audience.n", "audience", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.audience.recipients", "people gathered or addressed as recipients of a performance or message", "audience"}}},
}

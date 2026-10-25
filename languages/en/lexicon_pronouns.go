package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPersonalPronounSource = provenance.SourceRef{ID: "golingua-en-personal-pronouns-358", Path: "languages/en/lexicon_pronouns.go", Version: "2026-10-25-358", License: "internal", Author: "Codex by owner delegation"}

var englishPersonalPronounEntries = []lexicalEntry{
	{"en.i.pron", "I", lexicon.PartOfSpeech("pronoun"), []lexicalMeaning{{"en.i.speaker", "first person singular speaker pronoun", "person"}}},
	{"en.we.pron", "we", lexicon.PartOfSpeech("pronoun"), []lexicalMeaning{{"en.we.speaker_group", "first person plural inclusive speaker pronoun", "person"}}},
	{"en.you.pron", "you", lexicon.PartOfSpeech("pronoun"), []lexicalMeaning{{"en.you.addressee", "second person addressee pronoun with unspecified number", "person"}}},
}

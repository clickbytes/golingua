package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishProductionVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_production_verbs.go",
	Version: "2026-10-22-batch-30", License: "unknown", Author: "golingua contributors",
}

// These pending verbs name work, creation, recording and document editing.
var englishProductionVerbEntries = []lexicalEntry{
	{"en.working.v", "work", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.working.event", "apply effort to perform an assigned or productive task", "working"}}},
	{"en.making_object.v", "make", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.making_object.event", "bring a new artifact or object into existence by action", "making_object"}}},
	{"en.manufacturing.v", "manufacture", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.manufacturing.event", "produce goods by organized processes using equipment or labor", "manufacturing"}}},
	{"en.designing_object.v", "design", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.designing_object.event", "form a specification of how a new object or system should function or appear", "designing_object"}}},
	{"en.drawing_image.v", "draw", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.drawing_image.event", "create lines or marks to represent forms on a surface", "drawing_image"}}},
	{"en.photographing.v", "photograph", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.photographing.event", "capture a still image of a subject using a camera", "photographing"}}},
	{"en.recording_sound.v", "record", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.recording_sound.event", "capture sound into a reproducible recording", "recording_sound"}}},
	{"en.writing_document.v", "write", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.writing_document.event", "create or revise a text intended as a durable document", "writing_document"}}},
	{"en.typing_text.v", "type", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.typing_text.event", "enter characters through a keyboard or comparable input device", "typing_text"}}},
	{"en.editing_document.v", "edit", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.editing_document.event", "change existing document content to improve or alter its form or meaning", "editing_document"}}},
}

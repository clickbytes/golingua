package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishDocumentSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_documents.go",
	Version: "2026-10-22-batch-23", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish information-bearing objects and media.
var englishDocumentEntries = []lexicalEntry{
	{"en.newspaper.n", "newspaper", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.newspaper.direct", "a regularly issued publication containing news and related articles", "newspaper"}}},
	{"en.document.n", "document", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.document.direct", "an intentionally recorded information artifact with continuing content and form", "document"}}},
	{"en.message.n", "message", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.message.direct", "a discrete communicated or recorded unit intended to convey information to a recipient", "message"}}},
	{"en.letter_document.n", "letter", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.letter_document.direct", "a written document composed as correspondence to a specific recipient", "letter_document"}}},
	{"en.email_message.n", "email", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.email_message.direct", "an electronically transmitted message addressed through an email system", "email_message"}}},
	{"en.text_message.n", "text message", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.text_message.direct", "a short written message exchanged through a messaging network or service", "text_message"}}},
	{"en.notebook.n", "notebook", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.notebook.direct", "a bound or organized set of pages intended for notes", "notebook"}}},
	{"en.paper_sheet.n", "sheet of paper", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.paper_sheet.direct", "a single thin piece of paper made for writing, printing, or wrapping", "paper_sheet"}}},
	{"en.pen.n", "pen", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pen.direct", "a handheld writing instrument that delivers ink to a surface", "pen"}}},
	{"en.pencil.n", "pencil", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pencil.direct", "a handheld writing instrument that deposits graphite or similar material", "pencil"}}},
	{"en.map_document.n", "map", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.map_document.direct", "a spatial representation showing locations, routes, or boundaries", "map_document"}}},
	{"en.calendar_object.n", "calendar", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.calendar_object.direct", "a physical or digital display organizing dates and scheduled occasions", "calendar_object"}}},
	{"en.photograph.n", "photograph", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.photograph.direct", "a still visual image produced by recording light from a scene", "photograph"}}},
	{"en.drawing_artifact.n", "drawing", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.drawing_artifact.direct", "a visual representation produced mainly through lines or marks", "drawing_artifact"}}},
	{"en.music_recording.n", "music recording", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.music_recording.direct", "a durable representation of an audible musical performance or composition", "music_recording"}}},
	{"en.digital_file.n", "digital file", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.digital_file.direct", "an addressable stored unit of digital data distinct from its information content", "digital_file"}}},
	{"en.website.n", "website", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.website.direct", "a persistent collection of related web resources identified by a site address", "website"}}},
	{"en.application_software.n", "application", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.application_software.direct", "a program product intended to support an end user’s tasks", "application_software"}}},
}

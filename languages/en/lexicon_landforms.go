package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishLandformSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_landforms.go",
	Version: "2026-10-22-batch-22", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish surfaces from basic landforms.
var englishLandformEntries = []lexicalEntry{
	{"en.surface.n", "surface", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.surface.direct", "an identifiable exterior boundary of an object or region", "surface"}}},
	{"en.ground_surface.n", "ground", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ground_surface.direct", "the outer surface of terrain on which a person or object can stand", "ground_surface"}}},
	{"en.hill.n", "hill", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hill.direct", "a raised landform smaller or less dominant than a mountain", "hill"}}},
	{"en.mountain.n", "mountain", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mountain.direct", "a substantial elevation of Earth’s surface rising above surrounding terrain", "mountain"}}},
	{"en.valley.n", "valley", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.valley.direct", "a low-lying elongated landform between higher surrounding terrain", "valley"}}},
}

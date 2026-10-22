package en

import "github.com/clickbytes/golingua/internal/provenance"

type lexicalGroup struct {
	source  provenance.SourceRef
	entries []lexicalEntry
}

// englishLexiconGroups gathers language-local domains in a stable order.
func englishLexiconGroups() []lexicalGroup {
	return []lexicalGroup{
		{source: englishLexicalSource, entries: englishEntries},
		{source: englishNounSource, entries: englishNounEntries},
	}
}

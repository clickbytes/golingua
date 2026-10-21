package en

import (
	"github.com/clickbytes/golingua/internal/morph"
)

// inCodeMorphRecords reuses the declared regular seed paradigms only for
// lexical entries whose spelling matches those declared rules. Entries with
// irregular or unreviewed forms deliberately have no paradigm assignment.
func inCodeMorphRecords() ([]lexemeRecord, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	records := make([]lexemeRecord, 0, len(englishParadigmByLexeme))
	for _, lexeme := range document.Lexemes {
		if lexeme.ParadigmID == "" {
			continue
		}
		records = append(records, lexemeRecord{
			Lexeme:         lexeme,
			PrincipalParts: morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "base", Form: lexeme.Lemma}}},
		})
	}
	return records, nil
}

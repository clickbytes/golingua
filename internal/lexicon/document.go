package lexicon

const SchemaVersion = 1

// Document is the closed v1 JSON envelope for a lexical snapshot.
type Document struct {
	SchemaVersion int               `json:"schema_version"`
	Lexemes       []MorphLexeme     `json:"lexemes"`
	Senses        []LexicalSense    `json:"senses"`
	Bindings      []SemanticBinding `json:"bindings"`
}

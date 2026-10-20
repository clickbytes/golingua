package importer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/lexicon"
)

const (
	LexiconFileSchemaVersion = lexicon.SchemaVersion
	MaxLexiconFileBytes      = lexicon.MaxDocumentBytes
)

// LexiconFile keeps morphology, lexical senses, and semantic bindings separate.
type LexiconFile struct {
	SchemaVersion int                       `json:"schema_version"`
	Lexemes       []lexicon.MorphLexeme     `json:"lexemes"`
	Senses        []lexicon.LexicalSense    `json:"senses"`
	Bindings      []lexicon.SemanticBinding `json:"bindings"`
}

// DecodeLexiconFile decodes one bounded strict lexical file. CORE is only
// required when a record declares a semantic binding.
func DecodeLexiconFile(data []byte, ontology *core.ConceptGraph) (LexiconFile, error) {
	if len(data) > MaxLexiconFileBytes {
		return LexiconFile{}, resourceLimit("lexicon_bytes", len(data), MaxLexiconFileBytes)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		code := "INVALID_LEXICON_FILE"
		if errors.Is(err, errDuplicateJSONKey) {
			code = "DUPLICATE_JSON_KEY"
		}
		return LexiconFile{}, invalid(code, "lexicon", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document lexicon.Document
	if err := decoder.Decode(&document); err != nil {
		return LexiconFile{}, invalid("INVALID_LEXICON_FILE", "lexicon", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return LexiconFile{}, invalid("TRAILING_LEXICON_DATA", "lexicon", "lexicon file must contain exactly one JSON value")
	}
	if err := lexicon.ValidateDocument(document, ontology, lexicon.DefaultBudget()); err != nil {
		return LexiconFile{}, invalid("INVALID_LEXICON_FILE", "lexicon", err.Error())
	}
	return LexiconFile{SchemaVersion: document.SchemaVersion, Lexemes: document.Lexemes, Senses: document.Senses, Bindings: document.Bindings}, nil
}

func (file LexiconFile) Document() lexicon.Document {
	return lexicon.Document{SchemaVersion: file.SchemaVersion, Lexemes: file.Lexemes, Senses: file.Senses, Bindings: file.Bindings}
}

func (file LexiconFile) Validate(ontology *core.ConceptGraph) error {
	if file.SchemaVersion != LexiconFileSchemaVersion {
		return invalid("LEXICON_SCHEMA_VERSION", "schema_version", fmt.Sprintf("unsupported lexicon file schema version %d", file.SchemaVersion))
	}
	if err := lexicon.ValidateDocument(file.Document(), ontology, lexicon.DefaultBudget()); err != nil {
		return invalid("INVALID_LEXICON_FILE", "lexicon", err.Error())
	}
	return nil
}

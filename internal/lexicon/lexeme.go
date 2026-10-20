package lexicon

import "github.com/clickbytes/golingua/internal/ids"

// LanguageID is a canonical lowercase language tag (for example, "pl" or
// "en-us"). This package does not normalize spelling or infer language.
type LanguageID string

// PartOfSpeech is a bounded language-data label, not a grammatical model.
type PartOfSpeech string

// Property carries one bounded morphology feature without prescribing the
// feature inventory for any language.
type Property struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MorphLexeme is a language-specific form inventory entry. Concept mapping
// is deliberately stored separately in SemanticBinding.
type MorphLexeme struct {
	ID         ids.LexemeID `json:"id"`
	Language   LanguageID   `json:"language"`
	Lemma      string       `json:"lemma"`
	POS        PartOfSpeech `json:"pos"`
	Properties []Property   `json:"properties,omitempty"`
	ParadigmID string       `json:"paradigm_id,omitempty"`
}

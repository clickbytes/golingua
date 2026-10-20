package lexicon

import "github.com/clickbytes/golingua/internal/ids"

// ValencyID is an opaque reference to a separately validated valency frame.
type ValencyID string

// LexicalSense belongs to one MorphLexeme. Different senses remain separate
// even when their surface lemma is identical.
type LexicalSense struct {
	ID         ids.SenseID  `json:"id"`
	LexemeID   ids.LexemeID `json:"lexeme_id"`
	Register   string       `json:"register,omitempty"`
	Usage      []string     `json:"usage,omitempty"`
	ValencyIDs []ValencyID  `json:"valency_ids,omitempty"`
}

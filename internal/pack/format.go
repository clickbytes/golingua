// Package pack compiles validated CORE and lexicon snapshots into immutable
// in-memory indexes. It does not activate data or provide a translation runtime.
package pack

const FormatVersion = 1

// Identity pins a pack to one exact validated source snapshot.
type Identity struct {
	FormatVersion  int    `json:"format_version"`
	DatasetID      string `json:"dataset_id"`
	DatasetVersion string `json:"dataset_version"`
	Digest         string `json:"digest"`
}

// Summary reports the scope actually present in this pack.
type Summary struct {
	Identity      Identity `json:"identity"`
	Concepts      int      `json:"concepts"`
	RelationTypes int      `json:"relation_types"`
	Lexemes       int      `json:"lexemes"`
	Senses        int      `json:"senses"`
	Bindings      int      `json:"bindings"`
	Morphology    string   `json:"morphology"`
	Construction  string   `json:"construction"`
}

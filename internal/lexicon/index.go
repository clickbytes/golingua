package lexicon

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
)

// Index is an immutable compiled lexical snapshot. All internal slices and
// maps are private; accessors return defensive copies.
type Index struct {
	lexemes         []MorphLexeme
	senses          []LexicalSense
	bindings        []SemanticBinding
	byLexeme        map[string]int
	byLemma         map[string][]int
	sensesByLexeme  map[string][]int
	bindingsBySense map[string][]int
	bindingsByCore  map[string][]int
}

// Decode reads one strict versioned JSON document and validates its bounds
// and references. A CORE snapshot is needed only when bindings are present.
func Decode(data []byte, ontology *core.ConceptGraph, budget Budget) (Document, error) {
	if err := budget.Validate(); err != nil {
		return Document{}, err
	}
	if len(data) > budget.MaxBytes {
		return Document{}, resourceLimit("input_bytes", len(data), budget.MaxBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return Document{}, invalid("INVALID_LEXICON_JSON", "document", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Document{}, invalid("TRAILING_JSON", "document", "expected exactly one JSON value")
	}
	if err := ValidateDocument(document, ontology, budget); err != nil {
		return Document{}, err
	}
	return cloneDocument(document), nil
}

// Encode returns a deterministic bounded JSON representation after
// validating all records and any supplied CORE bindings.
func Encode(document Document, ontology *core.ConceptGraph, budget Budget) ([]byte, error) {
	if err := ValidateDocument(document, ontology, budget); err != nil {
		return nil, err
	}
	canonical := cloneDocument(document)
	if canonical.Lexemes == nil {
		canonical.Lexemes = []MorphLexeme{}
	}
	if canonical.Senses == nil {
		canonical.Senses = []LexicalSense{}
	}
	if canonical.Bindings == nil {
		canonical.Bindings = []SemanticBinding{}
	}
	sort.Slice(canonical.Lexemes, func(i, j int) bool { return canonical.Lexemes[i].ID.String() < canonical.Lexemes[j].ID.String() })
	sort.Slice(canonical.Senses, func(i, j int) bool { return canonical.Senses[i].ID.String() < canonical.Senses[j].ID.String() })
	sort.Slice(canonical.Bindings, func(i, j int) bool {
		return semanticBindingSortKey(canonical.Bindings[i]) < semanticBindingSortKey(canonical.Bindings[j])
	})
	for index := range canonical.Lexemes {
		sort.Slice(canonical.Lexemes[index].Properties, func(i, j int) bool {
			return canonical.Lexemes[index].Properties[i].Key < canonical.Lexemes[index].Properties[j].Key
		})
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, invalid("INVALID_LEXICON_JSON", "document", err.Error())
	}
	if len(encoded) > budget.MaxBytes {
		return nil, resourceLimit("input_bytes", len(encoded), budget.MaxBytes)
	}
	return encoded, nil
}

// Build compiles and freezes a validated lexical snapshot. Unmapped
// MorphLexeme data can be compiled without a CORE graph.
func Build(document Document, ontology *core.ConceptGraph, budget Budget) (*Index, error) {
	if err := ValidateDocument(document, ontology, budget); err != nil {
		return nil, err
	}
	document = cloneDocument(document)
	sort.Slice(document.Lexemes, func(i, j int) bool { return document.Lexemes[i].ID.String() < document.Lexemes[j].ID.String() })
	sort.Slice(document.Senses, func(i, j int) bool { return document.Senses[i].ID.String() < document.Senses[j].ID.String() })
	sort.Slice(document.Bindings, func(i, j int) bool {
		return semanticBindingSortKey(document.Bindings[i]) < semanticBindingSortKey(document.Bindings[j])
	})
	index := &Index{
		lexemes: document.Lexemes, senses: document.Senses, bindings: document.Bindings,
		byLexeme: make(map[string]int, len(document.Lexemes)),
		byLemma:  make(map[string][]int), sensesByLexeme: make(map[string][]int),
		bindingsBySense: make(map[string][]int), bindingsByCore: make(map[string][]int),
	}
	for position, lexeme := range index.lexemes {
		id := lexeme.ID.String()
		index.byLexeme[id] = position
		key := lemmaKey(lexeme.Language, lexeme.Lemma)
		index.byLemma[key] = append(index.byLemma[key], position)
	}
	for position, sense := range index.senses {
		index.sensesByLexeme[sense.LexemeID.String()] = append(index.sensesByLexeme[sense.LexemeID.String()], position)
	}
	for position, binding := range index.bindings {
		index.bindingsBySense[binding.SenseID.String()] = append(index.bindingsBySense[binding.SenseID.String()], position)
		if binding.Target == nil {
			index.bindingsByCore[binding.CoreRef.String()] = append(index.bindingsByCore[binding.CoreRef.String()], position)
		}
	}
	return index, nil
}

func lemmaKey(language LanguageID, lemma string) string { return string(language) + "\x00" + lemma }

func (index *Index) Len() (lexemes, senses, bindings int) {
	if index == nil {
		return 0, 0, 0
	}
	return len(index.lexemes), len(index.senses), len(index.bindings)
}

func (index *Index) ByLexemeID(id ids.LexemeID) (MorphLexeme, bool) {
	if index == nil {
		return MorphLexeme{}, false
	}
	position, exists := index.byLexeme[id.String()]
	if !exists {
		return MorphLexeme{}, false
	}
	return cloneLexeme(index.lexemes[position]), true
}

// BySenseID returns a defensive copy of the exact lexical sense.
func (index *Index) BySenseID(id ids.SenseID) (LexicalSense, bool) {
	if index == nil {
		return LexicalSense{}, false
	}
	key := id.String()
	position := sort.Search(len(index.senses), func(position int) bool {
		return index.senses[position].ID.String() >= key
	})
	if position >= len(index.senses) || index.senses[position].ID != id {
		return LexicalSense{}, false
	}
	return cloneSense(index.senses[position]), true
}

// LookupLexeme is a named lookup alias for ByLexemeID.
func (index *Index) LookupLexeme(id ids.LexemeID) (MorphLexeme, bool) {
	return index.ByLexemeID(id)
}

// ByLemma returns every exact language+lemma match, retaining homographs and
// same-lemma entries with different parts of speech.
func (index *Index) ByLemma(language LanguageID, lemma string) []MorphLexeme {
	if index == nil {
		return nil
	}
	positions := index.byLemma[lemmaKey(language, lemma)]
	result := make([]MorphLexeme, len(positions))
	for i, position := range positions {
		result[i] = cloneLexeme(index.lexemes[position])
	}
	return result
}

func (index *Index) SensesForLexeme(id ids.LexemeID) []LexicalSense {
	if index == nil {
		return nil
	}
	positions := index.sensesByLexeme[id.String()]
	result := make([]LexicalSense, len(positions))
	for i, position := range positions {
		result[i] = cloneSense(index.senses[position])
	}
	return result
}

func (index *Index) BindingsForSense(id ids.SenseID) []SemanticBinding {
	if index == nil {
		return nil
	}
	return index.copyBindings(index.bindingsBySense[id.String()])
}

// ByCoreConcept returns every mapping to a concept, including pending and
// non-equivalent relations. Callers must inspect kind and status explicitly.
func (index *Index) ByCoreConcept(id ids.ConceptID) []SemanticBinding {
	if index == nil {
		return nil
	}
	return index.copyBindings(index.bindingsByCore[id.String()])
}

// TransferBindings returns every approved equivalent mapping for one sense.
// Multiple results remain explicit; this method never selects one arbitrarily.
func (index *Index) TransferBindings(id ids.SenseID) []SemanticBinding {
	bindings := index.BindingsForSense(id)
	result := make([]SemanticBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.TransferAuthorized() {
			result = append(result, binding)
		}
	}
	return result
}

// TargetBindingsForSense returns all explicitly tagged targets for a sense.
// Results are detached; runtime callers must still validate selected metadata.
func (index *Index) TargetBindingsForSense(id ids.SenseID) []SemanticBinding {
	bindings := index.BindingsForSense(id)
	result := make([]SemanticBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Target != nil {
			result = append(result, binding)
		}
	}
	return result
}

// ExecutableTargetBindingsForSense applies the separate approval, exact-map,
// profile, and provenance gate without selecting among multiple targets.
func (index *Index) ExecutableTargetBindingsForSense(id ids.SenseID) []SemanticBinding {
	bindings := index.TargetBindingsForSense(id)
	result := make([]SemanticBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.TargetExecutionAuthorized() {
			result = append(result, binding)
		}
	}
	return result
}

func (index *Index) copyBindings(positions []int) []SemanticBinding {
	result := make([]SemanticBinding, len(positions))
	for i, position := range positions {
		result[i] = index.bindings[position]
		result[i].Target = cloneSemanticTarget(index.bindings[position].Target)
	}
	return result
}

// Document returns a defensive data copy of the compiled snapshot.
func (index *Index) Document() Document {
	if index == nil {
		return Document{}
	}
	return cloneDocument(Document{SchemaVersion: SchemaVersion, Lexemes: index.lexemes, Senses: index.senses, Bindings: index.bindings})
}

func cloneDocument(source Document) Document {
	result := Document{SchemaVersion: source.SchemaVersion, Lexemes: make([]MorphLexeme, len(source.Lexemes)), Senses: make([]LexicalSense, len(source.Senses)), Bindings: make([]SemanticBinding, len(source.Bindings))}
	for i, binding := range source.Bindings {
		result.Bindings[i] = binding
		result.Bindings[i].Target = cloneSemanticTarget(binding.Target)
	}
	for i, lexeme := range source.Lexemes {
		result.Lexemes[i] = cloneLexeme(lexeme)
	}
	for i, sense := range source.Senses {
		result.Senses[i] = cloneSense(sense)
	}
	return result
}

func cloneLexeme(source MorphLexeme) MorphLexeme {
	result := source
	result.Properties = append([]Property(nil), source.Properties...)
	return result
}

func cloneSense(source LexicalSense) LexicalSense {
	result := source
	result.Usage = append([]string(nil), source.Usage...)
	result.ValencyIDs = append([]ValencyID(nil), source.ValencyIDs...)
	return result
}

func semanticBindingSortKey(binding SemanticBinding) string {
	if binding.Target == nil {
		return binding.SenseID.String() + "\x00core\x00" + binding.CoreRef.String()
	}
	encoded, _ := json.Marshal(canonicalSemanticTarget(binding.Target))
	return binding.SenseID.String() + "\x00target\x00" + string(encoded)
}

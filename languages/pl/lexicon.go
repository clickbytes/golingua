package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

type lexicalMeaning struct {
	id, description, candidateCore string
}

type lexicalEntry struct {
	id, lemma string
	pos       lexicon.PartOfSpeech
	meanings  []lexicalMeaning
}

// These are locally authored lexical distinctions, not generated translations
// of CORE labels. Candidate links have no transfer authority.
var polishEntries = []lexicalEntry{
	{"pl.pies.n", "pies", "noun", []lexicalMeaning{{"pl.pies.animal", "domestic dog", "dog"}}},
	{"pl.kot.n", "kot", "noun", []lexicalMeaning{{"pl.kot.animal", "domestic cat", "cat"}}},
	{"pl.dziecko.n", "dziecko", "noun", []lexicalMeaning{{"pl.dziecko.person", "young person", "child"}}},
	{"pl.nauczyciel.n", "nauczyciel", "noun", []lexicalMeaning{{"pl.nauczyciel.role", "person acting as instructor", "teacher"}}},
	{"pl.lekarz.n", "lekarz", "noun", []lexicalMeaning{{"pl.lekarz.role", "medical practitioner", "doctor"}}},
	{"pl.dom.n", "dom", "noun", []lexicalMeaning{{"pl.dom.building", "residential building", "house"}, {"pl.dom.household", "people living together as a household", ""}}},
	{"pl.szkola.n", "szkoła", "noun", []lexicalMeaning{{"pl.szkola.facility", "building used for instruction", "school"}, {"pl.szkola.institution", "institution providing instruction", ""}}},
	{"pl.ksiazka.n", "książka", "noun", []lexicalMeaning{{"pl.ksiazka.physical", "physical book for reading", "book"}, {"pl.ksiazka.work", "written work regardless of physical carrier", ""}}},
	{"pl.reka.n", "ręka", "noun", []lexicalMeaning{{"pl.reka.hand", "terminal grasping part", "hand"}, {"pl.reka.arm", "arm including more than the hand", ""}}},
	{"pl.oko.n", "oko", "noun", []lexicalMeaning{{"pl.oko.organ", "organ for vision", "eye"}}},
	{"pl.woda.n", "woda", "noun", []lexicalMeaning{{"pl.woda.substance", "water as a substance", "water"}}},
	{"pl.widziec.v", "widzieć", "verb", []lexicalMeaning{{"pl.widziec.visual", "perceive visually", "seeing"}}},
	{"pl.patrzec.v", "patrzeć", "verb", []lexicalMeaning{{"pl.patrzec.directed", "direct visual attention to a target", "looking"}}},
	{"pl.slyszec.v", "słyszeć", "verb", []lexicalMeaning{{"pl.slyszec.auditory", "perceive through hearing", "hearing"}}},
	{"pl.dac.v", "dać", "verb", []lexicalMeaning{{"pl.dac.transfer", "initiate transfer to a recipient", "giving"}}},
	{"pl.dostac.v", "dostać", "verb", []lexicalMeaning{{"pl.dostac.receive", "become recipient of a theme", "receiving"}}},
}

var polishLexicalSource = provenance.SourceRef{
	ID: "golingua-pl-authored", Path: "languages/pl/lexicon.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var polishFrameBySense = map[string]lexicon.ValencyID{
	"pl.widziec.visual":   "pl.widziec.visual.transitive",
	"pl.patrzec.directed": "pl.patrzec.directed.na",
	"pl.slyszec.auditory": "pl.slyszec.auditory.transitive",
	"pl.dostac.receive":   "pl.dostac.receive.transitive",
}

var polishParadigmByLexeme = map[string]string{
	"pl.kot.n": "pl_masc_animal_regular",
}

// LexicalDocument constructs the Polish lexical data from Go declarations.
// Pending candidate links preserve possible CORE relationships without
// licensing translation. This function does not invent inflection or valency.
func LexicalDocument() (lexicon.Document, error) {
	if err := polishLexicalSource.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical source: %w", err)
	}
	document := lexicon.Document{SchemaVersion: lexicon.SchemaVersion}
	for _, entry := range polishEntries {
		lexemeID, err := ids.NewLexemeID(entry.id)
		if err != nil {
			return lexicon.Document{}, fmt.Errorf("Polish lexeme %s: %w", entry.id, err)
		}
		lexeme := lexicon.MorphLexeme{ID: lexemeID, Language: "pl", Lemma: entry.lemma, POS: entry.pos, ParadigmID: polishParadigmByLexeme[entry.id]}
		if entry.id == "pl.kot.n" {
			lexeme.Properties = []lexicon.Property{{Key: "gender", Value: "masculine_animal"}}
		}
		document.Lexemes = append(document.Lexemes, lexeme)
		for _, meaning := range entry.meanings {
			senseID, err := ids.NewSenseID(meaning.id)
			if err != nil {
				return lexicon.Document{}, fmt.Errorf("Polish sense %s: %w", meaning.id, err)
			}
			sense := lexicon.LexicalSense{ID: senseID, LexemeID: lexemeID, Usage: []string{meaning.description}}
			if frameID, ok := polishFrameBySense[meaning.id]; ok {
				sense.ValencyIDs = []lexicon.ValencyID{frameID}
			}
			document.Senses = append(document.Senses, sense)
			if meaning.candidateCore == "" {
				continue
			}
			conceptID, err := ids.NewConceptID(meaning.candidateCore)
			if err != nil {
				return lexicon.Document{}, fmt.Errorf("Polish candidate %s: %w", meaning.candidateCore, err)
			}
			document.Bindings = append(document.Bindings, lexicon.SemanticBinding{
				SenseID: senseID, CoreRef: conceptID, Mapping: lexicon.MappingCandidate,
				Status: lexicon.BindingPending, Provenance: polishLexicalSource,
			})
		}
	}
	return document, nil
}

// CompileLexicon validates the in-code Polish data against an exact CORE
// snapshot. The resulting index includes pending mappings for review; its
// TransferBindings view remains empty until separately approved data exists.
func CompileLexicon(ontology *core.ConceptGraph) (*lexicon.Index, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	return lexicon.Build(document, ontology, lexicon.DefaultBudget())
}

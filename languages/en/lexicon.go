package en

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

// English distinctions are authored independently from Polish lexical items.
// The proposed CORE references remain pending and cannot authorize transfer.
var englishEntries = []lexicalEntry{
	{"en.dog.n", "dog", "noun", []lexicalMeaning{{"en.dog.animal", "domestic canine animal", "dog"}}},
	{"en.cat.n", "cat", "noun", []lexicalMeaning{{"en.cat.animal", "domestic feline animal", "cat"}}},
	{"en.child.n", "child", "noun", []lexicalMeaning{{"en.child.person", "young person", "child"}}},
	{"en.teacher.n", "teacher", "noun", []lexicalMeaning{{"en.teacher.role", "person acting as instructor", "teacher"}}},
	{"en.doctor.n", "doctor", "noun", []lexicalMeaning{{"en.doctor.role", "medical practitioner", "doctor"}, {"en.doctor.academic", "holder of a doctoral degree", ""}}},
	{"en.house.n", "house", "noun", []lexicalMeaning{{"en.house.building", "residential building", "house"}, {"en.house.institution", "legislative or other institutional body", ""}}},
	{"en.school.n", "school", "noun", []lexicalMeaning{{"en.school.facility", "place built for instruction", "school"}, {"en.school.institution", "organization providing instruction", ""}, {"en.school.fish", "group of fish moving together", ""}}},
	{"en.book.n", "book", "noun", []lexicalMeaning{{"en.book.physical", "physical bound reading material", "book"}, {"en.book.work", "written work regardless of carrier", ""}}},
	{"en.hand.n", "hand", "noun", []lexicalMeaning{{"en.hand.body", "terminal grasping part of an arm", "hand"}, {"en.hand.worker", "person employed for manual work", ""}}},
	{"en.eye.n", "eye", "noun", []lexicalMeaning{{"en.eye.organ", "organ of visual perception", "eye"}}},
	{"en.water.n", "water", "noun", []lexicalMeaning{{"en.water.substance", "water as a substance", "water"}}},
	{"en.see.v", "see", "verb", []lexicalMeaning{{"en.see.visual", "perceive visually", "seeing"}, {"en.see.understand", "understand an idea", ""}}},
	{"en.look.v", "look", "verb", []lexicalMeaning{{"en.look.directed", "direct visual attention toward a target", "looking"}, {"en.look.appear", "have a visible appearance", ""}}},
	{"en.hear.v", "hear", "verb", []lexicalMeaning{{"en.hear.auditory", "perceive through hearing", "hearing"}}},
	{"en.get.v", "get", "verb", []lexicalMeaning{{"en.get.receive", "become recipient of something", "receiving"}, {"en.get.understand", "understand an idea", ""}}},
	{"en.give.v", "give", "verb", []lexicalMeaning{{"en.give.transfer", "initiate transfer to a recipient", "giving"}}},
	{"en.read.v", "read", "verb", []lexicalMeaning{{"en.read.text", "interpret written symbolic content", "reading"}}},
	{"en.learn.v", "learn", "verb", []lexicalMeaning{{"en.learn.knowledge", "acquire knowledge or skill", "learning"}}},
	{"en.sleep.v", "sleep", "verb", []lexicalMeaning{{"en.sleep.physiological", "undergo physiological sleep", "sleeping"}}},
}

var englishLexicalSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var englishFrameBySense = map[string]lexicon.ValencyID{
	"en.see.visual":    "en.see.visual.transitive",
	"en.look.directed": "en.look.directed.at",
	"en.hear.auditory": "en.hear.auditory.transitive",
	"en.get.receive":   "en.get.receive.transitive",
	"en.read.text":     "en.read.text.transitive",
}

var englishParadigmByLexeme = map[string]string{
	"en.dog.n": "en_noun_regular", "en.teacher.n": "en_noun_regular",
	"en.doctor.n": "en_noun_regular", "en.school.n": "en_noun_regular",
	"en.hand.n": "en_noun_regular", "en.eye.n": "en_noun_regular",
	"en.look.v": "en_verb_regular",
}

// LexicalDocument builds the English lexical records from Go declarations.
// It keeps polysemous senses distinct and leaves unmapped meanings unresolved.
func LexicalDocument() (lexicon.Document, error) {
	document := lexicon.Document{SchemaVersion: lexicon.SchemaVersion}
	for _, group := range englishLexiconGroups() {
		if err := group.source.Validate(); err != nil {
			return lexicon.Document{}, fmt.Errorf("English lexical source %q: %w", group.source.ID, err)
		}
		for _, entry := range group.entries {
			lexemeID, err := ids.NewLexemeID(entry.id)
			if err != nil {
				return lexicon.Document{}, fmt.Errorf("English lexeme %s: %w", entry.id, err)
			}
			document.Lexemes = append(document.Lexemes, lexicon.MorphLexeme{
				ID: lexemeID, Language: "en", Lemma: entry.lemma, POS: entry.pos,
				ParadigmID: englishParadigmByLexeme[entry.id],
			})
			for _, meaning := range entry.meanings {
				senseID, err := ids.NewSenseID(meaning.id)
				if err != nil {
					return lexicon.Document{}, fmt.Errorf("English sense %s: %w", meaning.id, err)
				}
				sense := lexicon.LexicalSense{ID: senseID, LexemeID: lexemeID, Usage: []string{meaning.description}}
				if frameID, ok := englishFrameBySense[meaning.id]; ok {
					sense.ValencyIDs = []lexicon.ValencyID{frameID}
				}
				document.Senses = append(document.Senses, sense)
				if meaning.candidateCore == "" {
					continue
				}
				conceptID, err := ids.NewConceptID(meaning.candidateCore)
				if err != nil {
					return lexicon.Document{}, fmt.Errorf("English candidate %s: %w", meaning.candidateCore, err)
				}
				document.Bindings = append(document.Bindings, lexicon.SemanticBinding{
					SenseID: senseID, CoreRef: conceptID, Mapping: lexicon.MappingCandidate,
					Status: lexicon.BindingPending, Provenance: group.source,
				})
			}
		}
	}
	return document, nil
}

// CompileLexicon validates in-code English data against an exact CORE graph.
// All candidate links remain pending and absent from TransferBindings.
func CompileLexicon(ontology *core.ConceptGraph) (*lexicon.Index, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	return lexicon.Build(document, ontology, lexicon.DefaultBudget())
}

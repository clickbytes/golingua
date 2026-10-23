package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var polishNounSource = provenance.SourceRef{
	ID: "golingua-pl-noun-authored", Path: "languages/pl/morphology.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var polishSentenceFormSource = provenance.SourceRef{
	ID: "golingua-pl-sentence-forms-authored", Path: "languages/pl/morphology.go",
	Version: "2026-10-23", License: "internal", Author: "golingua contributors",
}

// Only cells whose outputs are explicitly declared here are generated. This
// regular masculine animate pattern does not imply a complete Polish noun
// declension for every lexical item.
func inCodeParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	type cell struct{ id, number, grammaticalCase, suffix string }
	cells := []cell{
		{"sg_nom", "singular", "nominative", ""},
		{"sg_gen", "singular", "genitive", "a"},
		{"sg_acc", "singular", "accusative", "a"},
		{"sg_inst", "singular", "instrumental", "em"},
		{"pl_nom", "plural", "nominative", "y"},
		{"pl_gen", "plural", "genitive", "ów"},
	}
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion,
		ID:            "pl_masc_animal_regular", Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: "root", PartID: "base", Required: true}},
	}
	for _, item := range cells {
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
			ID: morph.CellID(item.id), Features: []features.FeatureValue{
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(item.number)},
				{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(item.grammaticalCase)},
				{Feature: "gender", State: features.StateKnown, Value: "masculine_animal"},
			},
		})
		steps := []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "root"}}
		if item.suffix != "" {
			steps = append(steps, procedure.Operation{Kind: procedure.OperationAppend, Text: item.suffix})
		}
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{
			CellID: morph.CellID(item.id), ProcedureID: "procedure:pl_masc_animal_regular." + item.id,
			Provenance: polishNounSource, Steps: steps,
		})
	}
	result := append([]morph.Paradigm{paradigm}, attestedNounParadigms(profile)...)
	return append(result, morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion,
		ID:            "pl_widziec_present_third_singular", Language: "pl", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: "form", PartID: "present_third_singular", Required: true}},
		Cells: []morph.ParadigmCell{{ID: "present_third_singular", Features: []features.FeatureValue{
			{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
			{Feature: "tense", State: features.StateKnown, Value: "present"},
			{Feature: "mood", State: features.StateKnown, Value: "indicative"},
			{Feature: "voice", State: features.StateKnown, Value: "active"},
			{Feature: "person", State: features.StateKnown, Value: "third"},
			{Feature: "number", State: features.StateKnown, Value: "singular"},
		}}},
		Rules: []morph.CellRule{{CellID: "present_third_singular", ProcedureID: "procedure:pl_widziec.present_third_singular", Provenance: polishSentenceFormSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "form"}}}},
	})
}

func inCodeMorphRecords() ([]lexemeRecord, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	attested, err := attestedNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	foundCat := false
	foundSee := false
	for _, lexeme := range document.Lexemes {
		if lexeme.ID.String() == "lexeme:pl.widziec.v" {
			foundSee = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "present_third_singular", Form: "widzi"}},
			}})
		}
		if lexeme.ID.String() == "lexeme:pl.kot.n" {
			foundCat = true
			attested = append(attested, lexemeRecord{
				Lexeme: lexeme,
				PrincipalParts: morph.PrincipalParts{
					LexemeID: lexeme.ID,
					Parts:    []morph.PrincipalPart{{ID: "base", Form: lexeme.Lemma}},
				},
			})
		}
	}
	if !foundCat || !foundSee {
		return nil, fmt.Errorf("Polish in-code morphology lexeme is missing")
	}
	return attested, nil
}

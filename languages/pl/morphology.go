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
	return append([]morph.Paradigm{paradigm}, attestedNounParadigms(profile)...)
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
	for _, lexeme := range document.Lexemes {
		if lexeme.ID.String() == "lexeme:pl.kot.n" {
			records := []lexemeRecord{{
				Lexeme: lexeme,
				PrincipalParts: morph.PrincipalParts{
					LexemeID: lexeme.ID,
					Parts: []morph.PrincipalPart{{ID: "base", Form: lexeme.Lemma}},
				},
			}}
			return append(records, attested...), nil
		}
	}
	return nil, fmt.Errorf("Polish in-code morphology lexeme is missing")
}

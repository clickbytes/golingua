package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
)

var identificationEntries = []lexicalEntry{
	{"pl.byc.v", "być", "verb", []lexicalMeaning{{"pl.byc.class_membership", "instantiate a class given by a referent", "class_membership"}}},
	{"pl.czarny.adj", "czarny", "adjective", []lexicalMeaning{{"pl.czarny.black", "black color attributed to one bearer", "black"}}},
}

func identificationParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	verb := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "pl_identification_byc_present_3sg",
		Language: "pl", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{
			StemID: "pl_identification_byc_present_3sg_form", PartID: "present_3sg_form", Required: true,
		}, {
			StemID: "pl_identification_byc_present_3pl_form", PartID: "present_3pl_form", Required: true,
		}},
		Cells: []morph.ParadigmCell{{
			ID: "present_indicative_active_imperfective_3sg",
			Features: []features.FeatureValue{
				{Feature: "aspect", State: features.StateKnown, Value: "imperfective"},
				{Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: "singular"},
				{Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
			},
		}, {
			ID: "present_indicative_active_imperfective_3pl",
			Features: []features.FeatureValue{
				{Feature: "aspect", State: features.StateKnown, Value: "imperfective"},
				{Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: "plural"},
				{Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
			},
		}},
		Rules: []morph.CellRule{{
			CellID:      "present_indicative_active_imperfective_3sg",
			ProcedureID: "procedure:pl_identification_byc_present_3sg.present_indicative_active_imperfective_3sg",
			Provenance:  polishIdentificationSource,
			Steps:       []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_identification_byc_present_3sg_form"}},
		}, {
			CellID:      "present_indicative_active_imperfective_3pl",
			ProcedureID: "procedure:pl_identification_byc_present_3sg.present_indicative_active_imperfective_3pl",
			Provenance:  polishIdentificationSource,
			Steps:       []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_identification_byc_present_3pl_form"}},
		}},
	}
	adjective := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "pl_identification_czarny_masculine_animal_nominative",
		Language: "pl", ProfileVersion: profile.Version(), POS: "adjective",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{
			StemID: "pl_identification_czarny_form", PartID: "positive_singular_masculine_animal_nominative_form", Required: true,
		}},
		Cells: []morph.ParadigmCell{
			{ID: "positive_singular_masculine_animal_nominative", Features: []features.FeatureValue{
				{Feature: "case", State: features.StateKnown, Value: "nominative"},
				{Feature: "degree", State: features.StateKnown, Value: "positive"},
				{Feature: "gender", State: features.StateKnown, Value: "masculine_animal"},
				{Feature: "number", State: features.StateKnown, Value: "singular"},
			}},
			{ID: "positive_singular_neuter_nominative", Features: []features.FeatureValue{
				{Feature: "case", State: features.StateKnown, Value: "nominative"},
				{Feature: "degree", State: features.StateKnown, Value: "positive"},
				{Feature: "gender", State: features.StateKnown, Value: "neuter"},
				{Feature: "number", State: features.StateKnown, Value: "singular"},
			}},
		},
		Rules: []morph.CellRule{
			{CellID: "positive_singular_masculine_animal_nominative", ProcedureID: "procedure:pl_identification_czarny_masculine_animal_nominative.positive_singular_masculine_animal_nominative", Provenance: polishIdentificationSource, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_identification_czarny_form"}}},
			{CellID: "positive_singular_neuter_nominative", ProcedureID: "procedure:pl_identification_czarny_masculine_animal_nominative.positive_singular_neuter_nominative", Provenance: polishIdentificationSource, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_identification_czarny_form"}, {Kind: procedure.OperationStripSuffix, Suffix: "y"}, {Kind: procedure.OperationAppend, Text: "e"}}},
		},
	}
	return []morph.Paradigm{verb, adjective}
}

func identificationMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	wanted := map[string]struct{}{"lexeme:pl.byc.v": {}, "lexeme:pl.czarny.adj": {}}
	found := make(map[string]lexicon.MorphLexeme, len(wanted))
	for _, lexeme := range lexemes {
		name := lexeme.ID.String()
		if _, ok := wanted[name]; ok {
			found[name] = lexeme
		}
	}
	verb, hasVerb := found["lexeme:pl.byc.v"]
	adjective, hasAdjective := found["lexeme:pl.czarny.adj"]
	if !hasVerb || !hasAdjective {
		return nil, fmt.Errorf("Polish identification morphology lexeme is missing")
	}
	if verb.Lemma != "być" || verb.POS != "verb" || verb.ParadigmID != "pl_identification_byc_present_3sg" || adjective.Lemma != "czarny" || adjective.POS != "adjective" || adjective.ParadigmID != "pl_identification_czarny_masculine_animal_nominative" {
		return nil, fmt.Errorf("Polish identification morphology lexeme has an incompatible declaration")
	}
	return []lexemeRecord{
		{Lexeme: verb, PrincipalParts: morph.PrincipalParts{LexemeID: verb.ID, Parts: []morph.PrincipalPart{{ID: "present_3sg_form", Form: "jest"}, {ID: "present_3pl_form", Form: "są"}}}},
		{Lexeme: adjective, PrincipalParts: morph.PrincipalParts{LexemeID: adjective.ID, Parts: []morph.PrincipalPart{{ID: "positive_singular_masculine_animal_nominative_form", Form: "czarny"}}}},
	}, nil
}

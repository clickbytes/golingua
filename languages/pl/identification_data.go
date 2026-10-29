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
	{"pl.czerwony.adj", "czerwony", "adjective", []lexicalMeaning{{"pl.czerwony.red", "red color attributed to one bearer", "red"}}},
	{"pl.bialy.adj", "biały", "adjective", []lexicalMeaning{{"pl.bialy.white", "white color attributed to one bearer", "white"}}},
	{"pl.niebieski.adj", "niebieski", "adjective", []lexicalMeaning{{"pl.niebieski.blue", "blue color attributed to one bearer", "blue"}}},
	{"pl.zielony.adj", "zielony", "adjective", []lexicalMeaning{{"pl.zielony.green", "green color attributed to one bearer", "green"}}},
	{"pl.otwarty.adj", "otwarty", "adjective", []lexicalMeaning{{"pl.otwarty.open", "open state attributed to a bearer", "open_state"}}},
	{"pl.smutny.adj", "smutny", "adjective", []lexicalMeaning{{"pl.smutny.sad", "sad affect attributed to one bearer", "sad"}}},
}

type polishAttributionSpec struct {
	sense, concept, lexeme string
}

var polishAttributionSpecs = []polishAttributionSpec{
	{sense: "sense:pl.czarny.black", concept: "concept:black", lexeme: "lexeme:pl.czarny.adj"},
	{sense: "sense:pl.czerwony.red", concept: "concept:red", lexeme: "lexeme:pl.czerwony.adj"},
	{sense: "sense:pl.bialy.white", concept: "concept:white", lexeme: "lexeme:pl.bialy.adj"},
	{sense: "sense:pl.niebieski.blue", concept: "concept:blue", lexeme: "lexeme:pl.niebieski.adj"},
	{sense: "sense:pl.zielony.green", concept: "concept:green", lexeme: "lexeme:pl.zielony.adj"},
	{sense: "sense:pl.otwarty.open", concept: "concept:open_state", lexeme: "lexeme:pl.otwarty.adj"},
	{sense: "sense:pl.smutny.sad", concept: "concept:sad", lexeme: "lexeme:pl.smutny.adj"},
}

func polishAttributionForSense(sense string) (polishAttributionSpec, bool) {
	for _, spec := range polishAttributionSpecs {
		if spec.sense == sense {
			return spec, true
		}
	}
	return polishAttributionSpec{}, false
}

func polishAttributionForConcept(concept string) (polishAttributionSpec, bool) {
	for _, spec := range polishAttributionSpecs {
		if spec.concept == concept {
			return spec, true
		}
	}
	return polishAttributionSpec{}, false
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
	return []morph.Paradigm{verb}
}

func identificationMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	wanted := map[string]struct{}{"lexeme:pl.byc.v": {}}
	found := make(map[string]lexicon.MorphLexeme, len(wanted))
	for _, lexeme := range lexemes {
		name := lexeme.ID.String()
		if _, ok := wanted[name]; ok {
			found[name] = lexeme
		}
	}
	verb, hasVerb := found["lexeme:pl.byc.v"]
	if !hasVerb {
		return nil, fmt.Errorf("Polish identification morphology lexeme is missing")
	}
	if verb.Lemma != "być" || verb.POS != "verb" || verb.ParadigmID != "pl_identification_byc_present_3sg" {
		return nil, fmt.Errorf("Polish identification morphology lexeme has an incompatible declaration")
	}
	return []lexemeRecord{{Lexeme: verb, PrincipalParts: morph.PrincipalParts{LexemeID: verb.ID, Parts: []morph.PrincipalPart{{ID: "present_3sg_form", Form: "jest"}, {ID: "present_3pl_form", Form: "są"}}}}}, nil
}

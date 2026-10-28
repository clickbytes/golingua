package en

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishIdentificationSource = provenance.SourceRef{
	ID: "golingua-en-identification-318", URI: "https://github.com/clickbytes/golingua/issues/318",
	Version: "2026-10-24", License: "internal", Author: "Codex by owner delegation",
}

var identificationEntries = []lexicalEntry{
	{"en.be.v", "be", "verb", []lexicalMeaning{{"en.be.class_membership", "instantiate the class given by a referent", "class_membership"}}},
	{"en.black.adj", "black", "adjective", []lexicalMeaning{{"en.black.color", "black color attributed to one bearer", "black"}}},
}

func identificationParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	verb := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "en_identification_be_present_3sg",
		Language: "en", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{
			StemID: "en_identification_be_present_3sg_form", PartID: "third_person_singular_form", Required: true,
		}, {
			StemID: "en_identification_be_present_3pl_form", PartID: "third_person_plural_form", Required: true,
		}},
		Cells: []morph.ParadigmCell{{
			ID: "finite_present_indicative_active_third_singular",
			Features: []features.FeatureValue{
				{Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: "singular"},
				{Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
				{Feature: "verb_number", State: features.StateKnown, Value: "singular"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
			},
		}, {
			ID: "finite_present_indicative_active_third_plural",
			Features: []features.FeatureValue{
				{Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: "plural"},
				{Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"},
				{Feature: "verb_number", State: features.StateKnown, Value: "plural"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
			},
		}},
		Rules: []morph.CellRule{{
			CellID:      "finite_present_indicative_active_third_singular",
			ProcedureID: "procedure:en_identification_be_present_3sg.finite_present_indicative_active_third_singular",
			Provenance:  englishIdentificationSource,
			Steps:       []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "en_identification_be_present_3sg_form"}},
		}, {
			CellID:      "finite_present_indicative_active_third_plural",
			ProcedureID: "procedure:en_identification_be_present_3sg.finite_present_indicative_active_third_plural",
			Provenance:  englishIdentificationSource,
			Steps:       []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "en_identification_be_present_3pl_form"}},
		}},
	}
	adjective := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "en_identification_black_positive",
		Language: "en", ProfileVersion: profile.Version(), POS: "adjective",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{
			StemID: "en_identification_black_positive_form", PartID: "positive_form", Required: true,
		}},
		Cells: []morph.ParadigmCell{{
			ID:       "positive",
			Features: []features.FeatureValue{{Feature: "adjective_degree", State: features.StateKnown, Value: "positive"}},
		}},
		Rules: []morph.CellRule{{
			CellID: "positive", ProcedureID: "procedure:en_identification_black_positive.positive",
			Provenance: englishIdentificationSource,
			Steps:      []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "en_identification_black_positive_form"}},
		}},
	}
	return []morph.Paradigm{verb, adjective}
}

func identificationMorphRecords(lexeme lexicon.MorphLexeme) (morph.PrincipalParts, bool) {
	switch lexeme.ID.String() {
	case "lexeme:en.be.v":
		return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "third_person_singular_form", Form: "is"}, {ID: "third_person_plural_form", Form: "are"}}}, true
	case "lexeme:en.black.adj":
		return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "positive_form", Form: "black"}}}, true
	default:
		return morph.PrincipalParts{}, false
	}
}

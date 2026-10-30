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

var englishAdjectiveDataSource = provenance.SourceRef{
	ID: "golingua-en-authored-adjectives-435", Path: "languages/en/identification_data.go",
	Version: "2026-10-28", License: "internal", Author: "Codex by owner delegation",
}

var englishPastAttributionSource = provenance.SourceRef{
	ID: "golingua-en-past-attribution-437", Path: "languages/en/identification_data.go",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var identificationEntries = []lexicalEntry{
	{"en.be.v", "be", "verb", []lexicalMeaning{{"en.be.class_membership", "instantiate the class given by a referent", "class_membership"}}},
	{"en.black.adj", "black", "adjective", []lexicalMeaning{{"en.black.color", "black color attributed to one bearer", "black"}}},
	{"en.red.adj", "red", "adjective", []lexicalMeaning{{"en.red.color", "red color attributed to one bearer", "red"}}},
	{"en.white.adj", "white", "adjective", []lexicalMeaning{{"en.white.color", "white color attributed to one bearer", "white"}}},
	{"en.blue.adj", "blue", "adjective", []lexicalMeaning{{"en.blue.color", "blue color attributed to one bearer", "blue"}}},
	{"en.green.adj", "green", "adjective", []lexicalMeaning{{"en.green.color", "green color attributed to one bearer", "green"}}},
	{"en.open.adj", "open", "adjective", []lexicalMeaning{{"en.open.state", "state of a bearer configured to permit passage, entry, or access", "open_state"}}},
	{"en.sad.adj", "sad", "adjective", []lexicalMeaning{{"en.sad.state", "sad affect attributed to one bearer", "sad"}}},
}

type englishAttributionSpec struct {
	sense, concept, lexeme string
	indefiniteArticle      string
}

var englishAttributionSpecs = []englishAttributionSpec{
	{sense: "sense:en.black.color", concept: "concept:black", lexeme: "lexeme:en.black.adj", indefiniteArticle: "a"},
	{sense: "sense:en.red.color", concept: "concept:red", lexeme: "lexeme:en.red.adj", indefiniteArticle: "a"},
	{sense: "sense:en.white.color", concept: "concept:white", lexeme: "lexeme:en.white.adj", indefiniteArticle: "a"},
	{sense: "sense:en.blue.color", concept: "concept:blue", lexeme: "lexeme:en.blue.adj", indefiniteArticle: "a"},
	{sense: "sense:en.green.color", concept: "concept:green", lexeme: "lexeme:en.green.adj", indefiniteArticle: "a"},
	{sense: "sense:en.open.state", concept: "concept:open_state", lexeme: "lexeme:en.open.adj", indefiniteArticle: "an"},
	{sense: "sense:en.sad.state", concept: "concept:sad", lexeme: "lexeme:en.sad.adj", indefiniteArticle: "a"},
}

func englishAttributionForSense(sense string) (englishAttributionSpec, bool) {
	for _, spec := range englishAttributionSpecs {
		if spec.sense == sense {
			return spec, true
		}
	}
	return englishAttributionSpec{}, false
}

func englishAttributionForConcept(concept string) (englishAttributionSpec, bool) {
	for _, spec := range englishAttributionSpecs {
		if spec.concept == concept {
			return spec, true
		}
	}
	return englishAttributionSpec{}, false
}

func identificationParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	verb := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "en_identification_be_present_3sg",
		Language: "en", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{
			StemID: "en_identification_be_present_3sg_form", PartID: "third_person_singular_form", Required: true,
		}, {
			StemID: "en_identification_be_present_3pl_form", PartID: "third_person_plural_form", Required: true,
		}, {
			StemID: "en_identification_be_past_singular_form", PartID: "past_singular_form", Required: true,
		}, {
			StemID: "en_identification_be_past_plural_form", PartID: "past_plural_form", Required: true,
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
		}, {
			ID: "finite_past_indicative_active_third_singular",
			Features: []features.FeatureValue{
				{Feature: "mood", State: features.StateKnown, Value: "indicative"}, {Feature: "number", State: features.StateKnown, Value: "singular"},
				{Feature: "person", State: features.StateKnown, Value: "third"}, {Feature: "tense", State: features.StateKnown, Value: "past"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "verb_number", State: features.StateKnown, Value: "singular"},
				{Feature: "voice", State: features.StateKnown, Value: "active"},
			},
		}, {
			ID: "finite_past_indicative_active_third_plural",
			Features: []features.FeatureValue{
				{Feature: "mood", State: features.StateKnown, Value: "indicative"}, {Feature: "number", State: features.StateKnown, Value: "plural"},
				{Feature: "person", State: features.StateKnown, Value: "third"}, {Feature: "tense", State: features.StateKnown, Value: "past"},
				{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "verb_number", State: features.StateKnown, Value: "plural"},
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
		}, {
			CellID: "finite_past_indicative_active_third_singular", ProcedureID: "procedure:en_identification_be_past.finite_singular", Provenance: englishPastAttributionSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "en_identification_be_past_singular_form"}},
		}, {
			CellID: "finite_past_indicative_active_third_plural", ProcedureID: "procedure:en_identification_be_past.finite_plural", Provenance: englishPastAttributionSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "en_identification_be_past_plural_form"}},
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
	result := []morph.Paradigm{verb, adjective}
	for _, entry := range identificationEntries[2:] {
		paradigmID := "en_identification_" + entry.lemma + "_positive"
		stemID := paradigmID + "_form"
		result = append(result, morph.Paradigm{
			SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(paradigmID),
			Language: "en", ProfileVersion: profile.Version(), POS: "adjective",
			PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: procedure.StemID(stemID), PartID: "positive_form", Required: true}},
			Cells:                     []morph.ParadigmCell{{ID: "positive", Features: []features.FeatureValue{{Feature: "adjective_degree", State: features.StateKnown, Value: "positive"}}}},
			Rules: []morph.CellRule{{CellID: "positive", ProcedureID: "procedure:" + paradigmID + ".positive", Provenance: englishAdjectiveDataSource,
				Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: procedure.StemID(stemID)}}}},
		})
	}
	return result
}

func identificationMorphRecords(lexeme lexicon.MorphLexeme) (morph.PrincipalParts, bool) {
	switch lexeme.ID.String() {
	case "lexeme:en.be.v":
		return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
			{ID: "third_person_singular_form", Form: "is"}, {ID: "third_person_plural_form", Form: "are"},
			{ID: "past_singular_form", Form: "was"}, {ID: "past_plural_form", Form: "were"},
		}}, true
	case "lexeme:en.black.adj":
		return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "positive_form", Form: "black"}}}, true
	case "lexeme:en.red.adj", "lexeme:en.white.adj", "lexeme:en.blue.adj", "lexeme:en.green.adj", "lexeme:en.open.adj", "lexeme:en.sad.adj":
		return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "positive_form", Form: lexeme.Lemma}}}, true
	default:
		return morph.PrincipalParts{}, false
	}
}

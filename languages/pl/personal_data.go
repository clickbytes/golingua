package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var polishPersonalSource = provenance.SourceRef{
	ID: "golingua-pl-personal-pronouns-358", Path: "languages/pl/personal_data.go",
	Version: "2026-10-25-358", License: "internal", Author: "Codex by owner delegation",
}

var polishPersonalEntries = []lexicalEntry{
	{"pl.ja.pron", "ja", "pronoun", []lexicalMeaning{{"pl.ja.speaker", "first person singular speaker pronoun", "person"}}},
	{"pl.my.pron", "my", "pronoun", []lexicalMeaning{{"pl.my.speaker_group", "first person plural inclusive speaker pronoun", "person"}}},
	{"pl.ty.pron", "ty", "pronoun", []lexicalMeaning{{"pl.ty.addressee", "second person singular addressee pronoun", "person"}}},
	{"pl.wy.pron", "wy", "pronoun", []lexicalMeaning{{"pl.wy.addressee_group", "second person plural addressee pronoun", "person"}}},
}

type polishPersonalPronoun struct {
	lexeme, cell, form, person, number string
}

var polishPersonalPronouns = []polishPersonalPronoun{
	{"pl.ja.pron", "nominative_first_singular", "ja", "first", "singular"},
	{"pl.my.pron", "nominative_first_plural", "my", "first", "plural"},
	{"pl.ty.pron", "nominative_second_singular", "ty", "second", "singular"},
	{"pl.wy.pron", "nominative_second_plural", "wy", "second", "plural"},
}

func polishPersonalParadigmID(lexeme string) morph.ParadigmID {
	return morph.ParadigmID("pl_personal_" + lexeme)
}

func polishPersonalParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := make([]morph.Paradigm, 0, len(polishPersonalPronouns))
	for _, pronoun := range polishPersonalPronouns {
		id := polishPersonalParadigmID(pronoun.lexeme)
		stem := procedure.StemID("form")
		part := morph.PrincipalPartID("form")
		cell := morph.CellID(pronoun.cell)
		result = append(result, morph.Paradigm{
			SchemaVersion: morph.ParadigmSchemaVersion, ID: id, Language: "pl", ProfileVersion: profile.Version(), POS: "pronoun",
			PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: stem, PartID: part, Required: true}},
			Cells: []morph.ParadigmCell{{ID: cell, Features: []features.FeatureValue{
				{Feature: "case", State: features.StateKnown, Value: "nominative"},
				{Feature: "person", State: features.StateKnown, Value: features.FeatureValueID(pronoun.person)},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(pronoun.number)},
			}}},
			Rules: []morph.CellRule{{CellID: cell, ProcedureID: "procedure:" + string(id) + "." + string(cell), Provenance: polishPersonalSource, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stem}}}},
		})
	}
	return result
}

func polishPersonalMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	byID := make(map[string]lexicon.MorphLexeme, len(polishPersonalPronouns))
	for _, lexeme := range lexemes {
		byID[lexeme.ID.String()] = lexeme
	}
	result := make([]lexemeRecord, 0, len(polishPersonalPronouns))
	for _, pronoun := range polishPersonalPronouns {
		lexeme, ok := byID["lexeme:"+pronoun.lexeme]
		if !ok || lexeme.POS != "pronoun" || lexeme.ParadigmID != string(polishPersonalParadigmID(pronoun.lexeme)) {
			return nil, fmt.Errorf("Polish personal pronoun morphology lexeme %s is missing or incompatible", pronoun.lexeme)
		}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "form", Form: pronoun.form}}}})
	}
	return result, nil
}

func polishPersonalConstructionBindings() []constructionBindingDeclaration {
	prodropVariants := []string{
		"pl.transitive_personal.prodrop.present.first.singular",
		"pl.transitive_personal.prodrop.present.first.plural",
		"pl.transitive_personal.prodrop.present.second.singular",
		"pl.transitive_personal.prodrop.present.second.plural",
	}
	explicitVariants := []string{
		"pl.transitive_personal_explicit.present.first.singular",
		"pl.transitive_personal_explicit.present.first.plural",
		"pl.transitive_personal_explicit.present.second.singular",
		"pl.transitive_personal_explicit.present.second.plural",
	}
	readingPastVariants := make([]string, 0, 8)
	for _, person := range []string{"first", "second"} {
		for _, number := range []string{"singular", "plural"} {
			for _, gender := range personalPastGenders(number) {
				readingPastVariants = append(readingPastVariants, "pl.transitive_personal.prodrop.past."+person+"."+number+"."+gender)
			}
		}
	}
	result := make([]constructionBindingDeclaration, 0, len(prodropVariants)*3+len(explicitVariants)*3+len(readingPastVariants))
	for _, sense := range []struct{ sense, frame string }{
		{"pl.widziec.visual", "pl.widziec.visual.transitive"},
		{"pl.slyszec.auditory", "pl.slyszec.auditory.transitive"},
		{"pl.czytac.interpret", "pl.czytac.interpret.frame"},
	} {
		for _, variant := range prodropVariants {
			result = append(result, constructionBindingDeclaration{sense.sense, sense.frame, "pl.transitive_personal", variant})
		}
		for _, variant := range explicitVariants {
			result = append(result, constructionBindingDeclaration{sense.sense, sense.frame, "pl.transitive_personal_explicit", variant})
		}
		if sense.sense == "pl.czytac.interpret" {
			for _, variant := range readingPastVariants {
				result = append(result, constructionBindingDeclaration{sense.sense, sense.frame, "pl.transitive_personal", variant})
			}
		}
	}
	return result
}

func personalPastGenders(number string) []string {
	if number == "singular" {
		return []string{"masculine_personal", "feminine"}
	}
	return []string{"masculine_personal", "non_masculine_personal"}
}

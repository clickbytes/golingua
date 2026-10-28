package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var eatingDrinkingMorphSource = provenance.SourceRef{
	ID: "wsjp_pl_jesc_present", URI: "https://wsjp.pl/haslo/podglad/27652/jesc/3062287/posilek",
	Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var drinkingMorphSource = provenance.SourceRef{
	ID: "wsjp_pl_pic_present", URI: "https://wsjp.pl/haslo/podglad/2324/pic/4698432/kawe",
	Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

// Only explicitly declared noun and adjective forms enter these paradigms.
func inCodeParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := attestedNounParadigms(profile)
	result = append(result, pluralOnlyNounParadigm(profile))
	result = append(result, treeNounParadigm(profile))
	result = append(result, doroslyAdjectiveParadigm(profile))
	result = append(result, identificationParadigms(profile)...)
	result = append(result, eatingDrinkingParadigms(profile)...)
	return append(result, polishPersonalParadigms(profile)...)
}

func eatingDrinkingParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := make([]morph.Paradigm, 0, 2)
	for _, item := range []struct {
		id    string
		parts []struct{ partID, form string }
	}{
		{"pl_verb_jesc_present_3sg_3pl", []struct{ partID, form string }{{"present_3sg_form", "je"}, {"present_3pl_form", "jedzą"}}},
		{"pl_verb_pic_present_3sg_3pl", []struct{ partID, form string }{{"present_3sg_form", "pije"}, {"present_3pl_form", "piją"}}},
	} {
		source := eatingDrinkingMorphSource
		if item.id == "pl_verb_pic_present_3sg_3pl" {
			source = drinkingMorphSource
		}
		paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(item.id), Language: "pl", ProfileVersion: profile.Version(), POS: "verb"}
		for _, cell := range []struct{ id, number, partID string }{{"present_3sg", "singular", "present_3sg_form"}, {"present_3pl", "plural", "present_3pl_form"}} {
			stemID := procedure.StemID(item.id + "_" + cell.partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: morph.PrincipalPartID(cell.partID), Required: true})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: morph.CellID("present_indicative_active_imperfective_" + cell.id), Features: []features.FeatureValue{
				{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)}, {Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"}, {Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
			}})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: morph.CellID("present_indicative_active_imperfective_" + cell.id), ProcedureID: "procedure:" + item.id + ".present_" + cell.number, Provenance: source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}}})
		}
		result = append(result, paradigm)
	}
	return result
}

func eatingDrinkingMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	forms := map[string]map[string]string{
		"lexeme:pl.jesc.v": {"present_3sg_form": "je", "present_3pl_form": "jedzą"},
		"lexeme:pl.pic.v":  {"present_3sg_form": "pije", "present_3pl_form": "piją"},
	}
	result := make([]lexemeRecord, 0, len(forms))
	for _, lexeme := range lexemes {
		parts, ok := forms[lexeme.ID.String()]
		if !ok {
			continue
		}
		if lexeme.Lemma == "jeść" {
			lexeme.ParadigmID = "pl_verb_jesc_present_3sg_3pl"
		} else {
			lexeme.ParadigmID = "pl_verb_pic_present_3sg_3pl"
		}
		principalParts := morph.PrincipalParts{LexemeID: lexeme.ID}
		for id, form := range parts {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID(id), Form: form})
		}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: principalParts})
		delete(forms, lexeme.ID.String())
	}
	if len(forms) != 0 {
		return nil, fmt.Errorf("Polish eating/drinking morphology lexeme is missing")
	}
	return result, nil
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
	pluralOnly, err := pluralOnlyNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, pluralOnly...)
	treeNouns, err := treeNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, treeNouns...)
	adjectives, err := doroslyAdjectiveRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, adjectives...)
	identification, err := identificationMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, identification...)
	personal, err := polishPersonalMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, personal...)
	foodVerbs, err := eatingDrinkingMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, foodVerbs...)
	foundCat := false
	foundBird := false
	foundSee := false
	foundHear := false
	foundLook := false
	foundRead := false
	foundReceive := false
	for _, lexeme := range document.Lexemes {
		if lexeme.ID.String() == "lexeme:pl.kot.n" {
			foundCat = true
		}
		if lexeme.ID.String() == "lexeme:pl.ptak.n" {
			foundBird = true
		}
		if lexeme.ID.String() == "lexeme:pl.widziec.v" {
			foundSee = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
					{ID: "present_stem", Form: "widz"},
					{ID: "past_stem", Form: "widzia"},
					{ID: "past_masculine_personal_plural_stem", Form: "widzie"},
				},
			}})
		}
		if lexeme.ID.String() == "lexeme:pl.slyszec.v" {
			foundHear = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "present_stem", Form: "słysz"}},
			}})
		}
		if lexeme.ID.String() == "lexeme:pl.patrzec.v" {
			foundLook = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "present_stem", Form: "patrz"}},
			}})
		}
		if lexeme.ID.String() == "lexeme:pl.czytac.v" {
			foundRead = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
					{ID: "present_stem", Form: "czyt"},
					{ID: "past_stem", Form: "czyta"},
				},
			}})
		}
		if lexeme.ID.String() == "lexeme:pl.dostac.v" {
			foundReceive = true
			attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
					{ID: "future_stem", Form: "dostan"},
					{ID: "past_stem", Form: "dosta"},
				},
			}})
		}
	}
	if !foundCat || !foundBird || !foundSee || !foundHear || !foundLook || !foundRead || !foundReceive {
		return nil, fmt.Errorf("Polish in-code morphology lexeme is missing")
	}
	return attested, nil
}

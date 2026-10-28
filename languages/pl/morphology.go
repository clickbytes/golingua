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
	result = append(result, polishPresentThirdPersonParadigms(profile)...)
	return append(result, polishPersonalParadigms(profile)...)
}

func polishPresentThirdPersonParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := make([]morph.Paradigm, 0, 13)
	for _, item := range []struct {
		id, singular, plural string
		source               provenance.SourceRef
	}{
		{"pl_verb_jesc_present_3sg_3pl", "je", "jedzą", eatingDrinkingMorphSource},
		{"pl_verb_pic_present_3sg_3pl", "pije", "piją", drinkingMorphSource},
		{"pl_verb_otwierac_obiekt_present_3sg_3pl", "otwiera", "otwierają", provenance.SourceRef{ID: "wsjp-pl-otwierac", URI: "https://wsjp.pl/haslo/podglad/16726/otwierac/4886610/okno", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_zamykac_obiekt_present_3sg_3pl", "zamyka", "zamykają", provenance.SourceRef{ID: "wsjp-pl-zamykac", URI: "https://wsjp.pl/haslo/podglad/4812/zamykac/5096451/okno", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_niesc_present_3sg_3pl", "niesie", "niosą", provenance.SourceRef{ID: "wsjp-pl-niesc", URI: "https://wsjp.pl/haslo/podglad/34633/niesc", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_trzymac_present_3sg_3pl", "trzyma", "trzymają", provenance.SourceRef{ID: "wsjp-pl-trzymac", URI: "https://wsjp.pl/haslo/podglad/17843/trzymac/5009629/w-dloni", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_ciagnac_present_3sg_3pl", "ciągnie", "ciągną", provenance.SourceRef{ID: "wsjp-pl-ciagnac", URI: "https://wsjp.pl/haslo/podglad/4760/ciagnac/5135673/woz", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_popychac_present_3sg_3pl", "popycha", "popychają", provenance.SourceRef{ID: "wsjp-pl-popychac", URI: "https://wsjp.pl/haslo/podglad/28202/popychac/5027558/wozek", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_pchac_present_3sg_3pl", "pcha", "pchają", provenance.SourceRef{ID: "wsjp-pl-pchac", URI: "https://wsjp.pl/haslo/podglad/28194/pchac/2702490/wozek", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_ciac_present_3sg_3pl", "tnie", "tną", provenance.SourceRef{ID: "wsjp-pl-ciac", URI: "https://wsjp.pl/haslo/podglad/1652/ciac/5013762/na-czesci", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_skladac_present_3sg_3pl", "składa", "składają", provenance.SourceRef{ID: "wsjp-pl-skladac", URI: "https://wsjp.pl/haslo/podglad/23658/skladac/5027088/parasol", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_drzec_present_3sg_3pl", "drze", "drą", provenance.SourceRef{ID: "wsjp-pl-drzec", URI: "https://wsjp.pl/haslo/podglad/36741/drzec/4891089/kartke", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_rozdzierac_present_3sg_3pl", "rozdziera", "rozdzierają", provenance.SourceRef{ID: "dobryslownik-pl-rozdzierac", URI: "https://dobryslownik.pl/slowo/rozdziera%C4%87/48819/", Version: "2026-10-28", License: "unknown", Author: "Dobry słownik języka polskiego i poradnia językowa"}},
	} {
		paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(item.id), Language: "pl", ProfileVersion: profile.Version(), POS: "verb"}
		for _, cell := range []struct{ id, number, partID string }{{"present_3sg", "singular", "present_3sg_form"}, {"present_3pl", "plural", "present_3pl_form"}} {
			stemID := procedure.StemID(item.id + "_" + cell.partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: morph.PrincipalPartID(cell.partID), Required: true})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: morph.CellID("present_indicative_active_imperfective_" + cell.id), Features: []features.FeatureValue{
				{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)}, {Feature: "person", State: features.StateKnown, Value: "third"},
				{Feature: "tense", State: features.StateKnown, Value: "present"}, {Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
			}})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: morph.CellID("present_indicative_active_imperfective_" + cell.id), ProcedureID: "procedure:" + item.id + ".present_" + cell.number, Provenance: item.source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}}})
		}
		result = append(result, paradigm)
	}
	return result
}

func polishPresentThirdPersonMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	forms := map[string]struct{ paradigmID, singular, plural string }{
		"lexeme:pl.jesc.v":            {"pl_verb_jesc_present_3sg_3pl", "je", "jedzą"},
		"lexeme:pl.pic.v":             {"pl_verb_pic_present_3sg_3pl", "pije", "piją"},
		"lexeme:pl.otwierac_obiekt.v": {"pl_verb_otwierac_obiekt_present_3sg_3pl", "otwiera", "otwierają"},
		"lexeme:pl.zamykac_obiekt.v":  {"pl_verb_zamykac_obiekt_present_3sg_3pl", "zamyka", "zamykają"},
		"lexeme:pl.niesc.v":           {"pl_verb_niesc_present_3sg_3pl", "niesie", "niosą"},
		"lexeme:pl.trzymac.v":         {"pl_verb_trzymac_present_3sg_3pl", "trzyma", "trzymają"},
		"lexeme:pl.ciagnac.v":         {"pl_verb_ciagnac_present_3sg_3pl", "ciągnie", "ciągną"},
		"lexeme:pl.popychac.v":        {"pl_verb_popychac_present_3sg_3pl", "popycha", "popychają"},
		"lexeme:pl.pchac.v":           {"pl_verb_pchac_present_3sg_3pl", "pcha", "pchają"},
		"lexeme:pl.ciac.v":            {"pl_verb_ciac_present_3sg_3pl", "tnie", "tną"},
		"lexeme:pl.skladac.v":         {"pl_verb_skladac_present_3sg_3pl", "składa", "składają"},
		"lexeme:pl.drzec.v":           {"pl_verb_drzec_present_3sg_3pl", "drze", "drą"},
		"lexeme:pl.rozdzierac.v":      {"pl_verb_rozdzierac_present_3sg_3pl", "rozdziera", "rozdzierają"},
	}
	result := make([]lexemeRecord, 0, len(forms))
	for _, lexeme := range lexemes {
		formsForLexeme, ok := forms[lexeme.ID.String()]
		if !ok {
			continue
		}
		lexeme.ParadigmID = formsForLexeme.paradigmID
		principalParts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
			{ID: "present_3sg_form", Form: formsForLexeme.singular},
			{ID: "present_3pl_form", Form: formsForLexeme.plural},
		}}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: principalParts})
		delete(forms, lexeme.ID.String())
	}
	if len(forms) != 0 {
		return nil, fmt.Errorf("Polish present-tense verb morphology lexeme is missing")
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
	presentVerbs, err := polishPresentThirdPersonMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, presentVerbs...)
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

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

var sleepingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-spac-past", URI: "https://wsjp.pl/haslo/podglad/3615/spac/4703056/w-nocy",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var givingPerfectivePastSource = provenance.SourceRef{
	ID: "wsjp-pl-dac-past", URI: "https://wsjp.pl/haslo/podglad/6368/dac",
	Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var givingImperfectivePastSource = provenance.SourceRef{
	ID: "golingua-pl-dawac-past-authored", Path: "languages/pl/morphology.go",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var helpingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-pomagac-present", URI: "https://wsjp.pl/haslo/podglad/25622/pomagac/2729386/w-nauce",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var thankingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-dziekowac-present", URI: "https://wsjp.pl/haslo/podglad/22696/dziekowac",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var usingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-uzywac-present", URI: "https://wsjp.pl/haslo/podglad/15960/uzywac",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var cuttingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-kroic-present", URI: "https://wsjp.pl/haslo/podglad/1564/kroic/2531996/chleb",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var cuttingPastMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-ciac-past", URI: "https://wsjp.pl/haslo/podglad/1652/ciac/5013763/szabla",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var conversingMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-rozmawiac-present", URI: "https://wsjp.pl/haslo/podglad/21480/rozmawiac",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var pushingPastMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-pchac-wozek-past", URI: "https://wsjp.pl/haslo/podglad/28194/pchac/2702490/wozek",
	Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var polishHaveMorphSource = provenance.SourceRef{
	ID: "wsjp-pl-miec-have-present", URI: "https://wsjp.pl/haslo/podglad/3243/miec",
	Version: "2026-10-30", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

var polishBelongingMorphSource = provenance.SourceRef{
	ID: "golingua-pl-nalezec-present-authored", URI: "https://github.com/clickbytes/golingua/issues/476#issuecomment-6528414953",
	Version: "2026-10-30", License: "internal", Author: "Codex by owner delegation",
}

// Only explicitly declared noun and adjective forms enter these paradigms.
func inCodeParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := attestedNounParadigms(profile)
	result = append(result, pluralOnlyNounParadigm(profile))
	result = append(result, treeNounParadigm(profile))
	result = append(result, doroslyAdjectiveParadigm(profile))
	result = append(result, positiveAdjectiveParadigms(profile)...)
	result = append(result, identificationParadigms(profile)...)
	result = append(result, polishPresentThirdPersonParadigms(profile)...)
	result = append(result, polishDaćPastParadigm(profile))
	result = append(result, polishKupicPastParadigm(profile))
	result = append(result, polishHavePersonalParadigm(profile))
	return append(result, polishPersonalParadigms(profile)...)
}

func polishHavePersonalParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	const id = "pl_verb_miec_personal"
	paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(id), Language: "pl", ProfileVersion: profile.Version(), POS: "verb"}
	for _, form := range []struct{ person, number, surface string }{
		{"first", "singular", "mam"}, {"second", "singular", "masz"}, {"third", "singular", "ma"},
		{"first", "plural", "mamy"}, {"second", "plural", "macie"}, {"third", "plural", "mają"},
	} {
		cellID := morph.CellID("present_" + form.person + "_" + form.number)
		partID := morph.PrincipalPartID("present_" + form.person + "_" + form.number)
		stemID := procedure.StemID(id + "_" + string(partID))
		paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: partID, Required: true})
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
			{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
			{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(form.number)}, {Feature: "person", State: features.StateKnown, Value: features.FeatureValueID(form.person)},
			{Feature: "tense", State: features.StateKnown, Value: "present"}, {Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
		}})
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + id + "." + string(cellID), Provenance: polishHaveMorphSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}}})
	}
	return paradigm
}

func polishKupicPastParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	const id = "pl_verb_kupic_past_3sg_3pl"
	paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(id), Language: "pl", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: "pl_verb_kupic_past_stem", PartID: "past_stem", Required: true}}}
	for _, cell := range []struct{ id, number, gender, ending string }{
		{"past_3sg_masc", "singular", "masculine_personal", "ł"}, {"past_3sg_fem", "singular", "feminine", "ła"}, {"past_3sg_neut", "singular", "neuter", "ło"},
		{"past_3pl_masc_personal", "plural", "masculine_personal", "li"}, {"past_3pl_non_masc_personal", "plural", "non_masculine_personal", "ły"},
	} {
		cellID := morph.CellID(cell.id)
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
			{Feature: "aspect", State: features.StateKnown, Value: "perfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
			{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(cell.gender)}, {Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)},
			{Feature: "person", State: features.StateKnown, Value: "third"}, {Feature: "tense", State: features.StateKnown, Value: "past"}, {Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
		}})
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + id + "." + cell.id, Provenance: provenance.SourceRef{ID: "golingua-pl-kupic-past-authored", Path: "languages/pl/morphology.go", Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation"}, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_verb_kupic_past_stem"}, {Kind: procedure.OperationAppend, Text: cell.ending}}})
	}
	return paradigm
}

func polishDaćPastParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	const id = "pl_verb_dac_past_3sg_masculine"
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: id, Language: "pl", ProfileVersion: profile.Version(), POS: "verb",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: "pl_verb_dac_past_stem", PartID: "past_stem", Required: true}},
		Cells:                     []morph.ParadigmCell{},
		Rules:                     []morph.CellRule{},
	}
	for _, cell := range []struct{ id, number, gender, ending string }{
		{"past_3sg_masculine_personal", "singular", "masculine_personal", "ł"},
		{"past_3sg_feminine", "singular", "feminine", "ła"},
		{"past_3sg_neuter", "singular", "neuter", "ło"},
		{"past_3pl_masculine_personal", "plural", "masculine_personal", "li"},
		{"past_3pl_non_masculine_personal", "plural", "non_masculine_personal", "ły"},
	} {
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: morph.CellID(cell.id), Features: []features.FeatureValue{
			{Feature: "aspect", State: features.StateKnown, Value: "perfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
			{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(cell.gender)}, {Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)},
			{Feature: "person", State: features.StateKnown, Value: "third"}, {Feature: "tense", State: features.StateKnown, Value: "past"},
			{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
		}})
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: morph.CellID(cell.id), ProcedureID: "procedure:" + id + "." + cell.id, Provenance: givingPerfectivePastSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "pl_verb_dac_past_stem"}, {Kind: procedure.OperationAppend, Text: cell.ending}}})
	}
	return paradigm
}

func polishPresentThirdPersonParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := make([]morph.Paradigm, 0, 32)
	for _, item := range []struct {
		id, singular, plural string
		source               provenance.SourceRef
	}{
		{"pl_verb_jesc_present_3sg_3pl", "je", "jedzą", eatingDrinkingMorphSource},
		{"pl_verb_pic_present_3sg_3pl", "pije", "piją", drinkingMorphSource},
		{"pl_verb_nalezec_present_3sg_3pl", "należy", "należą", polishBelongingMorphSource},
		{"pl_verb_otwierac_obiekt_present_3sg_3pl", "otwiera", "otwierają", provenance.SourceRef{ID: "wsjp-pl-otwierac", URI: "https://wsjp.pl/haslo/podglad/16726/otwierac/4886610/okno", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_zamykac_obiekt_present_3sg_3pl", "zamyka", "zamykają", provenance.SourceRef{ID: "wsjp-pl-zamykac", URI: "https://wsjp.pl/haslo/podglad/4812/zamykac/5096451/okno", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_niesc_present_3sg_3pl", "niesie", "niosą", provenance.SourceRef{ID: "wsjp-pl-niesc", URI: "https://wsjp.pl/haslo/podglad/34633/niesc", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_trzymac_present_3sg_3pl", "trzyma", "trzymają", provenance.SourceRef{ID: "wsjp-pl-trzymac", URI: "https://wsjp.pl/haslo/podglad/17843/trzymac/5009629/w-dloni", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_ciagnac_present_3sg_3pl", "ciągnie", "ciągną", provenance.SourceRef{ID: "wsjp-pl-ciagnac", URI: "https://wsjp.pl/haslo/podglad/4760/ciagnac/5135673/woz", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_popychac_present_3sg_3pl", "popycha", "popychają", provenance.SourceRef{ID: "wsjp-pl-popychac", URI: "https://wsjp.pl/haslo/podglad/28202/popychac/5027558/wozek", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_pchac_present_3sg_3pl", "pcha", "pchają", provenance.SourceRef{ID: "wsjp-pl-pchac", URI: "https://wsjp.pl/haslo/podglad/28194/pchac/2702490/wozek", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_ciac_present_3sg_3pl", "tnie", "tną", provenance.SourceRef{ID: "wsjp-pl-ciac", URI: "https://wsjp.pl/haslo/podglad/1652/ciac/5013762/na-czesci", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_kroic_present_3sg_3pl", "kroi", "kroją", cuttingMorphSource},
		{"pl_verb_rozmawiac_present_3sg_3pl", "rozmawia", "rozmawiają", conversingMorphSource},
		{"pl_verb_skladac_present_3sg_3pl", "składa", "składają", provenance.SourceRef{ID: "wsjp-pl-skladac-model", URI: "https://wsjp.pl/haslo/podglad/23658/skladac/5027090/model", Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_drzec_present_3sg_3pl", "drze", "drą", provenance.SourceRef{ID: "wsjp-pl-drzec", URI: "https://wsjp.pl/haslo/podglad/36741/drzec/4891089/kartke", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_rozdzierac_present_3sg_3pl", "rozdziera", "rozdzierają", provenance.SourceRef{ID: "dobryslownik-pl-rozdzierac", URI: "https://dobryslownik.pl/slowo/rozdziera%C4%87/48819/", Version: "2026-10-28", License: "unknown", Author: "Dobry słownik języka polskiego i poradnia językowa"}},
		{"pl_verb_spac_present_3sg_3pl", "śpi", "śpią", provenance.SourceRef{ID: "wsjp-pl-spac", URI: "https://wsjp.pl/haslo/podglad/3615/spac/4703056/w-nocy", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_oddychac_present_3sg_3pl", "oddycha", "oddychają", provenance.SourceRef{ID: "wsjp-pl-oddychac", URI: "https://wsjp.pl/haslo/podglad/4316/oddychac/1737962/ustami", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_kichac_present_3sg_3pl", "kicha", "kichają", provenance.SourceRef{ID: "wsjp-pl-kichac", URI: "https://wsjp.pl/haslo/podglad/92187/kichac/5229899/w-chorobie", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_biec_present_3sg_3pl", "biegnie", "biegną", provenance.SourceRef{ID: "wsjp-pl-biec", URI: "https://wsjp.pl/haslo/podglad/25072/biec/3986429/na-pomoc", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_kaszlec_present_3sg_3pl", "kaszle", "kaszlą", provenance.SourceRef{ID: "nck-pl-kaszlec", URI: "https://nck.pl/projekty-kulturalne/projekty/ojczysty-dodaj-do-ulubionych/ciekawostki-jezykowe/KASLAC_na_wiele_sposobow%2Ccltt%2Ck", Version: "2026-10-28", License: "unknown", Author: "Narodowe Centrum Kultury"}},
		{"pl_verb_budzic_sie_present_3sg_3pl", "budzi", "budzą", provenance.SourceRef{ID: "wsjp-pl-budzic-sie", URI: "https://wsjp.pl/haslo/podglad/3047/budzic-sie/4841359/ze-snu", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_chodzic_present_3sg_3pl", "chodzi", "chodzą", provenance.SourceRef{ID: "wsjp-pl-chodzic", URI: "https://wsjp.pl/haslo/podglad/10552/chodzic", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_zasypiac_present_3sg_3pl", "zasypia", "zasypiają", provenance.SourceRef{ID: "wsjp-pl-zasypiac", URI: "https://wsjp.pl/haslo/podglad/3218/zasypiac/516873/zaczynac-spac", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_zyc_present_3sg_3pl", "żyje", "żyją", provenance.SourceRef{ID: "wsjp-pl-zyc", URI: "https://wsjp.pl/haslo/podglad/2312/zyc", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_pracowac_present_3sg_3pl", "pracuje", "pracują", provenance.SourceRef{ID: "wsjp-pl-pracowac", URI: "https://wsjp.pl/haslo/podglad/19959/pracowac/5068993/w-banku", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_pomagac_present_3sg_3pl", "pomaga", "pomagają", helpingMorphSource},
		{"pl_verb_dziekowac_present_3sg_3pl", "dziękuje", "dziękują", thankingMorphSource},
		{"pl_verb_uzywac_present_3sg_3pl", "używa", "używają", usingMorphSource},
		{"pl_verb_latac_present_3sg_3pl", "lata", "latają", provenance.SourceRef{ID: "wsjp-pl-latac", URI: "https://wsjp.pl/haslo/podglad/38367/latac/4212091/samolot-ptak", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_isc_present_3sg_3pl", "idzie", "idą", provenance.SourceRef{ID: "wsjp-pl-isc", URI: "https://wsjp.pl/haslo/podglad/19012/isc/5066139/do-domu", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_dawac_present_3sg_3pl", "daje", "dają", provenance.SourceRef{ID: "wsjp-pl-dawac", URI: "https://wsjp.pl/haslo/podglad/6365/dawac/4488497/prezent", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_kupowac_present_3sg_3pl", "kupuje", "kupują", provenance.SourceRef{ID: "wsjp-pl-kupowac", URI: "https://wsjp.pl/index.php/haslo/podglad/7275/kupowac/2372345/mieszkanie", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_sprzedawac_present_3sg_3pl", "sprzedaje", "sprzedają", provenance.SourceRef{ID: "wsjp-pl-sprzedawac", URI: "https://wsjp.pl/haslo/podglad/18191/sprzedawac/4925066/w-sklepie", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_lubic_present_3sg_3pl", "lubi", "lubią", provenance.SourceRef{ID: "wsjp-pl-lubic", URI: "https://wsjp.pl/haslo/podglad/37948/lubic/4064834/ludzi", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_ogladac_present_3sg_3pl", "ogląda", "oglądają", provenance.SourceRef{ID: "wsjp-pl-ogladac", URI: "https://wsjp.pl/index.php/haslo/podglad/20347/ogladac/4739983/towar", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
		{"pl_verb_pisac_present_3sg_3pl", "pisze", "piszą", provenance.SourceRef{ID: "wsjp-pl-pisac", URI: "https://wsjp.pl/haslo/podglad/9846/pisac/3960164/do-kolegi", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	} {
		pos := "verb"
		if item.id == "pl_verb_budzic_sie_present_3sg_3pl" {
			pos = "verb_reflexive"
		}
		paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(item.id), Language: "pl", ProfileVersion: profile.Version(), POS: lexicon.PartOfSpeech(pos)}
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
		if item.id == "pl_verb_pic_present_3sg_3pl" || item.id == "pl_verb_spac_present_3sg_3pl" || item.id == "pl_verb_oddychac_present_3sg_3pl" || item.id == "pl_verb_pisac_present_3sg_3pl" || item.id == "pl_verb_dawac_present_3sg_3pl" || item.id == "pl_verb_kupowac_present_3sg_3pl" || item.id == "pl_verb_otwierac_obiekt_present_3sg_3pl" || item.id == "pl_verb_zamykac_obiekt_present_3sg_3pl" || item.id == "pl_verb_lubic_present_3sg_3pl" || item.id == "pl_verb_trzymac_present_3sg_3pl" || item.id == "pl_verb_pchac_present_3sg_3pl" || item.id == "pl_verb_ciagnac_present_3sg_3pl" || item.id == "pl_verb_ciac_present_3sg_3pl" || item.id == "pl_verb_skladac_present_3sg_3pl" || item.id == "pl_verb_biec_present_3sg_3pl" || item.id == "pl_verb_rozdzierac_present_3sg_3pl" || item.id == "pl_verb_kichac_present_3sg_3pl" || item.id == "pl_verb_kaszlec_present_3sg_3pl" || item.id == "pl_verb_ogladac_present_3sg_3pl" || item.id == "pl_verb_sprzedawac_present_3sg_3pl" {
			stemID := procedure.StemID(item.id + "_past_stem")
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: "past_stem", Required: true})
			if item.id == "pl_verb_kaszlec_present_3sg_3pl" {
				paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: procedure.StemID(item.id + "_past_masc_plural_stem"), PartID: "past_masc_plural_stem", Required: true})
			}
			for _, cell := range []struct{ id, number, gender, ending string }{
				{"past_3sg_masc", "singular", "masculine_personal", "ł"},
				{"past_3sg_fem", "singular", "feminine", "ła"},
				{"past_3sg_neut", "singular", "neuter", "ło"},
				{"past_3pl_masc_personal", "plural", "masculine_personal", "li"},
				{"past_3pl_non_masc_personal", "plural", "non_masculine_personal", "ły"},
			} {
				if item.id == "pl_verb_ciagnac_present_3sg_3pl" {
					switch cell.id {
					case "past_3sg_masc":
						cell.ending = "ął"
					case "past_3sg_fem":
						cell.ending = "ęła"
					case "past_3sg_neut":
						cell.ending = "ęło"
					case "past_3pl_masc_personal":
						cell.ending = "ęli"
					case "past_3pl_non_masc_personal":
						cell.ending = "ęły"
					}
				}
				if item.id == "pl_verb_ciac_present_3sg_3pl" {
					switch cell.id {
					case "past_3sg_masc":
						cell.ending = "ął"
					case "past_3sg_fem":
						cell.ending = "ęła"
					case "past_3sg_neut":
						cell.ending = "ęło"
					case "past_3pl_masc_personal":
						cell.ending = "ęli"
					case "past_3pl_non_masc_personal":
						cell.ending = "ęły"
					}
				}
				cellID := morph.CellID(cell.id)
				paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
					{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)}, {Feature: "person", State: features.StateKnown, Value: "third"},
					{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(cell.gender)}, {Feature: "tense", State: features.StateKnown, Value: "past"},
					{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
				}})
				cellSource := drinkingMorphSource
				if item.id == "pl_verb_spac_present_3sg_3pl" {
					cellSource = sleepingMorphSource
				} else if item.id == "pl_verb_oddychac_present_3sg_3pl" {
					cellSource = item.source
				} else if item.id == "pl_verb_pisac_present_3sg_3pl" {
					cellSource = provenance.SourceRef{ID: "wsjp-pl-pisac", URI: "https://wsjp.pl/haslo/podglad/9846/pisac/3960164/do-kolegi", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}
				} else if item.id == "pl_verb_dawac_present_3sg_3pl" {
					cellSource = givingImperfectivePastSource
				} else if item.id == "pl_verb_kupowac_present_3sg_3pl" {
					cellSource = provenance.SourceRef{ID: "golingua-pl-kupowac-past-authored", Path: "languages/pl/morphology.go", Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation"}
				} else if item.id == "pl_verb_otwierac_obiekt_present_3sg_3pl" || item.id == "pl_verb_zamykac_obiekt_present_3sg_3pl" {
					cellSource = item.source
				} else if item.id == "pl_verb_lubic_present_3sg_3pl" || item.id == "pl_verb_trzymac_present_3sg_3pl" {
					cellSource = item.source
				} else if item.id == "pl_verb_pchac_present_3sg_3pl" {
					cellSource = pushingPastMorphSource
				} else if item.id == "pl_verb_ciagnac_present_3sg_3pl" {
					cellSource = item.source
				} else if item.id == "pl_verb_ciac_present_3sg_3pl" {
					cellSource = cuttingPastMorphSource
				} else if item.id == "pl_verb_rozdzierac_present_3sg_3pl" {
					cellSource = item.source
				} else if item.id == "pl_verb_ogladac_present_3sg_3pl" {
					cellSource = provenance.SourceRef{ID: "golingua-pl-ogladac-past-authored", Path: "languages/pl/morphology.go", Version: "2026-10-30", License: "internal", Author: "Codex by owner delegation"}
				} else if item.id == "pl_verb_sprzedawac_present_3sg_3pl" {
					cellSource = provenance.SourceRef{ID: "golingua-pl-sprzedawac-past-authored", Path: "languages/pl/morphology.go", Version: "2026-10-30", License: "internal", Author: "Codex by owner delegation"}
				} else if item.id == "pl_verb_kichac_present_3sg_3pl" || item.id == "pl_verb_kaszlec_present_3sg_3pl" {
					cellSource = item.source
				}
				if item.id == "pl_verb_skladac_present_3sg_3pl" {
					cellSource = item.source
				}
				if item.id == "pl_verb_biec_present_3sg_3pl" {
					cellSource = provenance.SourceRef{ID: "wsjp-pl-biec-motion", URI: "https://wsjp.pl/haslo/podglad/25072/biec/3986428/po-schodach", Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}
				}
				cellStemID := stemID
				if item.id == "pl_verb_kaszlec_present_3sg_3pl" && cell.id == "past_3pl_masc_personal" {
					cellStemID = procedure.StemID(item.id + "_past_masc_plural_stem")
				}
				paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + item.id + "." + cell.id, Provenance: cellSource,
					Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: cellStemID}, {Kind: procedure.OperationAppend, Text: cell.ending}}})
			}
		}
		if item.id == "pl_verb_niesc_present_3sg_3pl" {
			for _, stem := range []struct{ partID, form string }{{"past_masc_stem", "nió"}, {"past_stem", "nios"}, {"past_masc_plural_stem", "nie"}} {
				stemID := procedure.StemID(item.id + "_" + stem.partID)
				paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: morph.PrincipalPartID(stem.partID), Required: true})
			}
			for _, cell := range []struct{ id, number, gender, partID, ending string }{
				{"past_3sg_masc", "singular", "masculine_personal", "past_masc_stem", "sł"},
				{"past_3sg_fem", "singular", "feminine", "past_stem", "ła"},
				{"past_3sg_neut", "singular", "neuter", "past_stem", "ło"},
				{"past_3pl_masc_personal", "plural", "masculine_personal", "past_masc_plural_stem", "śli"},
				{"past_3pl_non_masc_personal", "plural", "non_masculine_personal", "past_stem", "ły"},
			} {
				cellID := morph.CellID(cell.id)
				paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
					{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(cell.number)}, {Feature: "person", State: features.StateKnown, Value: "third"},
					{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(cell.gender)}, {Feature: "tense", State: features.StateKnown, Value: "past"},
					{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
				}})
				stemID := procedure.StemID(item.id + "_" + cell.partID)
				paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + item.id + "." + cell.id, Provenance: item.source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}, {Kind: procedure.OperationAppend, Text: cell.ending}}})
			}
		}
		if item.id == "pl_verb_jesc_present_3sg_3pl" {
			for _, cell := range []struct {
				partID string
				forms  []struct{ id, number, gender, ending string }
			}{
				{"past_singular_stem", []struct{ id, number, gender, ending string }{{"past_3sg_masc", "singular", "masculine_personal", "ł"}, {"past_3sg_fem", "singular", "feminine", "ła"}, {"past_3sg_neut", "singular", "neuter", "ło"}}},
				{"past_pl_masc", []struct{ id, number, gender, ending string }{{"past_3pl_masc_personal", "plural", "masculine_personal", "li"}}},
				{"past_pl_nonmasc", []struct{ id, number, gender, ending string }{{"past_3pl_non_masc_personal", "plural", "non_masculine_personal", "ły"}}},
			} {
				stemID := procedure.StemID(item.id + "_" + cell.partID)
				paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: morph.PrincipalPartID(cell.partID), Required: true})
				for _, form := range cell.forms {
					cellID := morph.CellID(form.id)
					paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
						{Feature: "aspect", State: features.StateKnown, Value: "imperfective"}, {Feature: "mood", State: features.StateKnown, Value: "indicative"},
						{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(form.number)}, {Feature: "person", State: features.StateKnown, Value: "third"},
						{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(form.gender)}, {Feature: "tense", State: features.StateKnown, Value: "past"},
						{Feature: "verb_form", State: features.StateKnown, Value: "finite"}, {Feature: "voice", State: features.StateKnown, Value: "active"},
					}})
					paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + item.id + "." + form.id, Provenance: eatingDrinkingMorphSource, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}, {Kind: procedure.OperationAppend, Text: form.ending}}})
				}
			}
		}
		result = append(result, paradigm)
	}
	return result
}

func polishPresentThirdPersonMorphRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	forms := map[string]struct{ paradigmID, singular, plural string }{
		"lexeme:pl.miec.v":            {"pl_verb_miec_personal", "ma", "mają"},
		"lexeme:pl.nalezec.v":         {"pl_verb_nalezec_present_3sg_3pl", "należy", "należą"},
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
		"lexeme:pl.kroic.v":           {"pl_verb_kroic_present_3sg_3pl", "kroi", "kroją"},
		"lexeme:pl.rozmawiac.v":       {"pl_verb_rozmawiac_present_3sg_3pl", "rozmawia", "rozmawiają"},
		"lexeme:pl.skladac.v":         {"pl_verb_skladac_present_3sg_3pl", "składa", "składają"},
		"lexeme:pl.drzec.v":           {"pl_verb_drzec_present_3sg_3pl", "drze", "drą"},
		"lexeme:pl.rozdzierac.v":      {"pl_verb_rozdzierac_present_3sg_3pl", "rozdziera", "rozdzierają"},
		"lexeme:pl.spac.v":            {"pl_verb_spac_present_3sg_3pl", "śpi", "śpią"},
		"lexeme:pl.oddychac.v":        {"pl_verb_oddychac_present_3sg_3pl", "oddycha", "oddychają"},
		"lexeme:pl.kichac.v":          {"pl_verb_kichac_present_3sg_3pl", "kicha", "kichają"},
		"lexeme:pl.biec.v":            {"pl_verb_biec_present_3sg_3pl", "biegnie", "biegną"},
		"lexeme:pl.kaszlec.v":         {"pl_verb_kaszlec_present_3sg_3pl", "kaszle", "kaszlą"},
		"lexeme:pl.budzic_sie.v":      {"pl_verb_budzic_sie_present_3sg_3pl", "budzi", "budzą"},
		"lexeme:pl.chodzic.v":         {"pl_verb_chodzic_present_3sg_3pl", "chodzi", "chodzą"},
		"lexeme:pl.zasypiac.v":        {"pl_verb_zasypiac_present_3sg_3pl", "zasypia", "zasypiają"},
		"lexeme:pl.zyc.v":             {"pl_verb_zyc_present_3sg_3pl", "żyje", "żyją"},
		"lexeme:pl.pracowac.v":        {"pl_verb_pracowac_present_3sg_3pl", "pracuje", "pracują"},
		"lexeme:pl.pomagac.v":         {"pl_verb_pomagac_present_3sg_3pl", "pomaga", "pomagają"},
		"lexeme:pl.dziekowac.v":       {"pl_verb_dziekowac_present_3sg_3pl", "dziękuje", "dziękują"},
		"lexeme:pl.uzywac.v":          {"pl_verb_uzywac_present_3sg_3pl", "używa", "używają"},
		"lexeme:pl.latac.v":           {"pl_verb_latac_present_3sg_3pl", "lata", "latają"},
		"lexeme:pl.isc.v":             {"pl_verb_isc_present_3sg_3pl", "idzie", "idą"},
		"lexeme:pl.dawac.v":           {"pl_verb_dawac_present_3sg_3pl", "daje", "dają"},
		"lexeme:pl.kupowac.v":         {"pl_verb_kupowac_present_3sg_3pl", "kupuje", "kupują"},
		"lexeme:pl.kupic.v":           {"pl_verb_kupic_past_3sg_3pl", "kupuje", "kupują"},
		"lexeme:pl.sprzedawac.v":      {"pl_verb_sprzedawac_present_3sg_3pl", "sprzedaje", "sprzedają"},
		"lexeme:pl.lubic.v":           {"pl_verb_lubic_present_3sg_3pl", "lubi", "lubią"},
		"lexeme:pl.ogladac.v":         {"pl_verb_ogladac_present_3sg_3pl", "ogląda", "oglądają"},
		"lexeme:pl.pisac.v":           {"pl_verb_pisac_present_3sg_3pl", "pisze", "piszą"},
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
		if lexeme.ID.String() == "lexeme:pl.miec.v" {
			principalParts.Parts = []morph.PrincipalPart{
				{ID: "present_first_singular", Form: "mam"}, {ID: "present_second_singular", Form: "masz"},
				{ID: "present_third_singular", Form: "ma"}, {ID: "present_first_plural", Form: "mamy"},
				{ID: "present_second_plural", Form: "macie"}, {ID: "present_third_plural", Form: "mają"},
			}
		}
		if lexeme.ID.String() == "lexeme:pl.pic.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "pi"})
		}
		if lexeme.ID.String() == "lexeme:pl.otwierac_obiekt.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "otwiera"})
		}
		if lexeme.ID.String() == "lexeme:pl.zamykac_obiekt.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "zamyka"})
		}
		if lexeme.ID.String() == "lexeme:pl.dawac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "dawa"})
		}
		if lexeme.ID.String() == "lexeme:pl.kupowac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "kupowa"})
		}
		if lexeme.ID.String() == "lexeme:pl.kupic.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "kupi"})
		}
		if lexeme.ID.String() == "lexeme:pl.spac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "spa"})
		}
		if lexeme.ID.String() == "lexeme:pl.oddychac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "oddycha"})
		}
		if lexeme.ID.String() == "lexeme:pl.pisac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "pisa"})
		}
		if lexeme.ID.String() == "lexeme:pl.lubic.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "lubi"})
		}
		if lexeme.ID.String() == "lexeme:pl.trzymac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "trzyma"})
		}
		if lexeme.ID.String() == "lexeme:pl.niesc.v" {
			principalParts.Parts = append(principalParts.Parts,
				morph.PrincipalPart{ID: "past_masc_stem", Form: "nió"},
				morph.PrincipalPart{ID: "past_stem", Form: "nios"},
				morph.PrincipalPart{ID: "past_masc_plural_stem", Form: "nie"},
			)
		}
		if lexeme.ID.String() == "lexeme:pl.jesc.v" {
			principalParts.Parts = append(principalParts.Parts,
				morph.PrincipalPart{ID: "past_singular_stem", Form: "jad"},
				morph.PrincipalPart{ID: "past_pl_masc", Form: "jed"},
				morph.PrincipalPart{ID: "past_pl_nonmasc", Form: "jad"},
			)
		}
		if lexeme.ID.String() == "lexeme:pl.pchac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "pcha"})
		}
		if lexeme.ID.String() == "lexeme:pl.ciagnac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "ciągn"})
		}
		if lexeme.ID.String() == "lexeme:pl.ciac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "ci"})
		}
		if lexeme.ID.String() == "lexeme:pl.skladac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "składa"})
		}
		if lexeme.ID.String() == "lexeme:pl.biec.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "bieg"})
		}
		if lexeme.ID.String() == "lexeme:pl.rozdzierac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "rozdziera"})
		}
		if lexeme.ID.String() == "lexeme:pl.kichac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "kicha"})
		}
		if lexeme.ID.String() == "lexeme:pl.kaszlec.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "kaszla"}, morph.PrincipalPart{ID: "past_masc_plural_stem", Form: "kaszle"})
		}
		if lexeme.ID.String() == "lexeme:pl.ogladac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "ogląda"})
		}
		if lexeme.ID.String() == "lexeme:pl.sprzedawac.v" {
			principalParts.Parts = append(principalParts.Parts, morph.PrincipalPart{ID: "past_stem", Form: "sprzedawa"})
		}
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
	positiveAdjectives, err := positiveAdjectiveRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	attested = append(attested, positiveAdjectives...)
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
	dacFound := false
	for _, lexeme := range document.Lexemes {
		if lexeme.ID.String() != "lexeme:pl.dac.v" {
			continue
		}
		dacFound = true
		lexeme.ParadigmID = "pl_verb_dac_past_3sg_masculine"
		attested = append(attested, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "past_stem", Form: "da"}}}})
	}
	if !dacFound {
		return nil, fmt.Errorf("Polish dać past morphology lexeme is missing")
	}
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
				LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
					{ID: "present_stem", Form: "słysz"},
					{ID: "past_stem", Form: "słysza"},
					{ID: "past_masculine_personal_plural_stem", Form: "słysze"},
				},
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

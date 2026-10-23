package pl

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var attestedNounSource = provenance.SourceRef{
	ID:      "golingua-pl-noun-forms-authored",
	Path:    "languages/pl/morphology_attested_nouns.go",
	Version: "2026-10-23-authored-v3",
	License: "unknown",
	Author:  "golingua contributors",
}

var attestedNounExternalSources = map[string]provenance.SourceRef{
	"pl.kot.n": {
		ID: "wsjp-pl-kot-n", URI: "https://wsjp.pl/haslo/podglad/29368/kot/5136345/domowy",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.ptak.n": {
		ID: "wsjp-pl-ptak-n", URI: "https://wsjp.pl/haslo/podglad/20014/ptak/4177524/stworzenie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.kon.n": {
		ID: "wsjp-pl-kon-n", URI: "https://wsjp.pl/haslo/podglad/16472/kon/5069035/zwierze",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.miasto.n": {
		ID: "wsjp-pl-miasto-n", URI: "https://wsjp.pl/haslo/podglad/3383/miasto/2451641/teren",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.mieszkanie.n": {
		ID: "wsjp-pl-mieszkanie-n", URI: "https://wsjp.pl/haslo/podglad/3382/mieszkanie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.rzeka.n": {
		ID: "wsjp-pl-rzeka-n", URI: "https://wsjp.pl/haslo/podglad/1735/rzeka/667911/woda",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.siostra.n": {
		ID: "wsjp-pl-siostra-n", URI: "https://wsjp.pl/haslo/podglad/8259/siostra/1883127/w-rodzinie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.mewa.n": {
		ID: "wsjp-pl-mewa-n", URI: "https://wsjp.pl/haslo/podglad/37575/mewa/4059050/ptak",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
}

var treeNounFormSource = provenance.SourceRef{
	ID:      "golingua-pl-tree-forms-wsj",
	URI:     "https://wsjp.pl/haslo/podglad/10895/drzewo/4680852/roslina",
	Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

// attestedNounForms contains every declared case/number form explicitly.
// Shared paradigms select these principal parts; they NEVER invent a suffix,
// sound change, missing form or an unreviewed alternate reading.
type attestedNounEntry struct {
	id, lemma, gender string
	singular, plural  [7]string
}

var attestedNounForms = []attestedNounEntry{
	{"pl.matka.n", "matka", "feminine", [7]string{"matka", "matki", "matce", "matkę", "matką", "matce", "matko"}, [7]string{"matki", "matek", "matkom", "matki", "matkami", "matkach", "matki"}},
	{"pl.ojciec.n", "ojciec", "masculine_personal", [7]string{"ojciec", "ojca", "ojcu", "ojca", "ojcem", "ojcu", "ojcze"}, [7]string{"ojcowie", "ojców", "ojcom", "ojców", "ojcami", "ojcach", "ojcowie"}},
	{"pl.pies.n", "pies", "masculine_animal", [7]string{"pies", "psa", "psu", "psa", "psem", "psie", "psie"}, [7]string{"psy", "psów", "psom", "psy", "psami", "psach", "psy"}},
	{"pl.dom.n", "dom", "masculine_inanimate", [7]string{"dom", "domu", "domowi", "dom", "domem", "domu", "domu"}, [7]string{"domy", "domów", "domom", "domy", "domami", "domach", "domy"}},
	{"pl.woda.n", "woda", "feminine", [7]string{"woda", "wody", "wodzie", "wodę", "wodą", "wodzie", "wodo"}, [7]string{"wody", "wód", "wodom", "wody", "wodami", "wodach", "wody"}},
	{"pl.szkola.n", "szkoła", "feminine", [7]string{"szkoła", "szkoły", "szkole", "szkołę", "szkołą", "szkole", "szkoło"}, [7]string{"szkoły", "szkół", "szkołom", "szkoły", "szkołami", "szkołach", "szkoły"}},
	{"pl.ksiazka.n", "książka", "feminine", [7]string{"książka", "książki", "książce", "książkę", "książką", "książce", "książko"}, [7]string{"książki", "książek", "książkom", "książki", "książkami", "książkach", "książki"}},
	{"pl.reka.n", "ręka", "feminine", [7]string{"ręka", "ręki", "ręce", "rękę", "ręką", "ręce", "ręko"}, [7]string{"ręce", "rąk", "rękom", "ręce", "rękami", "rękach", "ręce"}},
	{"pl.oko.n", "oko", "neuter", [7]string{"oko", "oka", "oku", "oko", "okiem", "oku", "oko"}, [7]string{"oczy", "oczu", "oczom", "oczy", "oczami", "oczach", "oczy"}},
	{"pl.dziecko.n", "dziecko", "neuter", [7]string{"dziecko", "dziecka", "dziecku", "dziecko", "dzieckiem", "dziecku", "dziecko"}, [7]string{"dzieci", "dzieci", "dzieciom", "dzieci", "dziećmi", "dzieciach", "dzieci"}},
	{"pl.stol.n", "stół", "masculine_inanimate", [7]string{"stół", "stołu", "stołowi", "stół", "stołem", "stole", "stole"}, [7]string{"stoły", "stołów", "stołom", "stoły", "stołami", "stołach", "stoły"}},
	{"pl.nauczyciel.n", "nauczyciel", "masculine_personal", [7]string{"nauczyciel", "nauczyciela", "nauczycielowi", "nauczyciela", "nauczycielem", "nauczycielu", "nauczycielu"}, [7]string{"nauczyciele", "nauczycieli", "nauczycielom", "nauczycieli", "nauczycielami", "nauczycielach", "nauczyciele"}},
	{"pl.lekarz.n", "lekarz", "masculine_personal", [7]string{"lekarz", "lekarza", "lekarzowi", "lekarza", "lekarzem", "lekarzu", "lekarzu"}, [7]string{"lekarze", "lekarzy", "lekarzom", "lekarzy", "lekarzami", "lekarzach", "lekarze"}},
	{"pl.samochod.n", "samochód", "masculine_inanimate", [7]string{"samochód", "samochodu", "samochodowi", "samochód", "samochodem", "samochodzie", "samochodzie"}, [7]string{"samochody", "samochodów", "samochodom", "samochody", "samochodami", "samochodach", "samochody"}},
	{"pl.ryba.n", "ryba", "feminine", [7]string{"ryba", "ryby", "rybie", "rybę", "rybą", "rybie", "rybo"}, [7]string{"ryby", "ryb", "rybom", "ryby", "rybami", "rybach", "ryby"}},
	{"pl.szpital.n", "szpital", "masculine_inanimate", [7]string{"szpital", "szpitala", "szpitalowi", "szpital", "szpitalem", "szpitalu", "szpitalu"}, [7]string{"szpitale", "szpitali", "szpitalom", "szpitale", "szpitalami", "szpitalach", "szpitale"}},
	{"pl.pizza.n", "pizza", "feminine", [7]string{"pizza", "pizzy", "pizzy", "pizzę", "pizzą", "pizzy", "pizzo"}, [7]string{"pizze", "pizz", "pizzom", "pizze", "pizzami", "pizzach", "pizze"}},
	{"pl.sasiad.n", "sąsiad", "masculine_personal", [7]string{"sąsiad", "sąsiada", "sąsiadowi", "sąsiada", "sąsiadem", "sąsiedzie", "sąsiedzie"}, [7]string{"sąsiedzi", "sąsiadów", "sąsiadom", "sąsiadów", "sąsiadami", "sąsiadach", "sąsiedzi"}},
	{"pl.klient.n", "klient", "masculine_personal", [7]string{"klient", "klienta", "klientowi", "klienta", "klientem", "kliencie", "kliencie"}, [7]string{"klienci", "klientów", "klientom", "klientów", "klientami", "klientach", "klienci"}},
	{"pl.jedzenie.n", "jedzenie", "neuter", [7]string{"jedzenie", "jedzenia", "jedzeniu", "jedzenie", "jedzeniem", "jedzeniu", "jedzenie"}, [7]string{"jedzenia", "jedzeń", "jedzeniom", "jedzenia", "jedzeniami", "jedzeniach", "jedzenia"}},
	{"pl.wiadomosc.n", "wiadomość", "feminine", [7]string{"wiadomość", "wiadomości", "wiadomości", "wiadomość", "wiadomością", "wiadomości", "wiadomości"}, [7]string{"wiadomości", "wiadomości", "wiadomościom", "wiadomości", "wiadomościami", "wiadomościach", "wiadomości"}},
	{"pl.kot.n", "kot", "masculine_animal", [7]string{"kot", "kota", "kotu", "kota", "kotem", "kocie", "kocie"}, [7]string{"koty", "kotów", "kotom", "koty", "kotami", "kotach", "koty"}},
	{"pl.ptak.n", "ptak", "masculine_animal", [7]string{"ptak", "ptaka", "ptakowi", "ptaka", "ptakiem", "ptaku", "ptaku"}, [7]string{"ptaki", "ptaków", "ptakom", "ptaki", "ptakami", "ptakach", "ptaki"}},
	{"pl.kon.n", "koń", "masculine_animal", [7]string{"koń", "konia", "koniowi", "konia", "koniem", "koniu", "koniu"}, [7]string{"konie", "koni", "koniom", "konie", "końmi", "koniach", "konie"}},
	{"pl.miasto.n", "miasto", "neuter", [7]string{"miasto", "miasta", "miastu", "miasto", "miastem", "mieście", "miasto"}, [7]string{"miasta", "miast", "miastom", "miasta", "miastami", "miastach", "miasta"}},
	{"pl.mieszkanie.n", "mieszkanie", "neuter", [7]string{"mieszkanie", "mieszkania", "mieszkaniu", "mieszkanie", "mieszkaniem", "mieszkaniu", "mieszkanie"}, [7]string{"mieszkania", "mieszkań", "mieszkaniom", "mieszkania", "mieszkaniami", "mieszkaniach", "mieszkania"}},
	{"pl.rzeka.n", "rzeka", "feminine", [7]string{"rzeka", "rzeki", "rzece", "rzekę", "rzeką", "rzece", "rzeko"}, [7]string{"rzeki", "rzek", "rzekom", "rzeki", "rzekami", "rzekach", "rzeki"}},
	{"pl.siostra.n", "siostra", "feminine", [7]string{"siostra", "siostry", "siostrze", "siostrę", "siostrą", "siostrze", "siostro"}, [7]string{"siostry", "sióstr", "siostrom", "siostry", "siostrami", "siostrach", "siostry"}},
	{"pl.mewa.n", "mewa", "feminine", [7]string{"mewa", "mewy", "mewie", "mewę", "mewą", "mewie", "mewo"}, [7]string{"mewy", "mew", "mewom", "mewy", "mewami", "mewach", "mewy"}},
}

var attestedNounCases = [7]string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"}

func attestedNounGender(lexemeID string) (string, bool) {
	switch lexemeID {
	case "pl.kot.n", "pl.ptak.n", "pl.kon.n":
		return "masculine_animal", true
	case "pl.miasto.n", "pl.mieszkanie.n":
		return "neuter", true
	case "pl.rzeka.n", "pl.siostra.n", "pl.mewa.n":
		return "feminine", true
	case "pl.matka.n":
		return "feminine", true
	case "pl.ojciec.n":
		return "masculine_personal", true
	case "pl.pies.n":
		return "masculine_animal", true
	case "pl.dom.n":
		return "masculine_inanimate", true
	case "pl.woda.n":
		return "feminine", true
	case "pl.szkola.n":
		return "feminine", true
	case "pl.ksiazka.n":
		return "feminine", true
	case "pl.reka.n":
		return "feminine", true
	case "pl.oko.n":
		return "neuter", true
	case "pl.dziecko.n":
		return "neuter", true
	case "pl.stol.n":
		return "masculine_inanimate", true
	case "pl.nauczyciel.n":
		return "masculine_personal", true
	case "pl.lekarz.n":
		return "masculine_personal", true
	case "pl.samochod.n":
		return "masculine_inanimate", true
	case "pl.ryba.n":
		return "feminine", true
	case "pl.szpital.n":
		return "masculine_inanimate", true
	case "pl.pizza.n":
		return "feminine", true
	case "pl.sasiad.n", "pl.klient.n":
		return "masculine_personal", true
	case "pl.jedzenie.n":
		return "neuter", true
	case "pl.wiadomosc.n":
		return "feminine", true
	default:
		return "", false
	}
}

func pluralOnlyNounParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion,
		ID:            "pl_noun_plural_only_attested", Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
	}
	for _, grammaticalCase := range attestedNounCases {
		partID := morph.PrincipalPartID("pl_" + grammaticalCase)
		cellID := morph.CellID(partID)
		stemID := procedure.StemID(partID)
		paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
			StemID: stemID, PartID: partID, Required: true,
		})
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
			ID: cellID,
			Features: []features.FeatureValue{
				{Feature: "number", State: features.StateKnown, Value: "plural"},
				{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
			},
		})
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{
			CellID: cellID, ProcedureID: "procedure:pl_noun_plural_only_attested." + string(cellID), Provenance: attestedNounSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
		})
	}
	return paradigm
}

func pluralOnlyNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	for _, lexeme := range lexemes {
		if lexeme.ID.String() != "lexeme:pl.pieniadze.n" {
			continue
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, len(attestedNounCases))}
		forms := [7]string{"pieniądze", "pieniędzy", "pieniądzom", "pieniądze", "pieniędzmi", "pieniądzach", "pieniądze"}
		for index, grammaticalCase := range attestedNounCases {
			parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID("pl_" + grammaticalCase), Form: forms[index]})
		}
		return []lexemeRecord{{Lexeme: lexeme, PrincipalParts: parts}}, nil
	}
	return nil, fmt.Errorf("Polish plural-only noun lexeme is missing")
}

func attestedNounParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	genders := []string{"masculine_personal", "masculine_animal", "masculine_inanimate", "feminine", "neuter"}
	result := make([]morph.Paradigm, 0, len(genders)+len(attestedNounExternalSources))
	for _, gender := range genders {
		paradigmID := morph.ParadigmID("pl_noun_attested_" + gender)
		result = append(result, attestedNounParadigm(profile, paradigmID, gender, attestedNounSource))
	}
	for _, entry := range attestedNounForms {
		source, external := attestedNounExternalSources[entry.id]
		if !external {
			continue
		}
		paradigmID := attestedNounParadigmID(entry.id, entry.gender)
		result = append(result, attestedNounParadigm(profile, paradigmID, entry.gender, source))
	}
	return result
}

func attestedNounParadigmID(lexemeID, gender string) morph.ParadigmID {
	if _, external := attestedNounExternalSources[lexemeID]; external {
		stableSuffix := strings.NewReplacer(".", "_", "-", "_").Replace(lexemeID)
		return morph.ParadigmID("pl_noun_attested_" + stableSuffix)
	}
	return morph.ParadigmID("pl_noun_attested_" + gender)
}

func attestedNounParadigm(profile *features.LanguageFeatureProfile, id morph.ParadigmID, gender string, source provenance.SourceRef) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: id, Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
	}
	for _, number := range []string{"singular", "plural"} {
		prefix := "sg"
		if number == "plural" {
			prefix = "pl"
		}
		for _, grammaticalCase := range attestedNounCases {
			partID := morph.PrincipalPartID(prefix + "_" + grammaticalCase)
			stemID := procedure.StemID(partID)
			cellID := morph.CellID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
				ID: cellID,
				Features: []features.FeatureValue{
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
					{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(gender)},
				},
			})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{
				CellID: cellID, ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID),
				Provenance: source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
			})
		}
	}
	return paradigm
}

func attestedNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	indexed := make(map[string]lexicon.MorphLexeme, len(lexemes))
	for _, lexeme := range lexemes {
		indexed[lexeme.ID.String()] = lexeme
	}
	result := make([]lexemeRecord, 0, len(attestedNounForms))
	for _, entry := range attestedNounForms {
		lexeme, ok := indexed["lexeme:"+entry.id]
		if !ok || lexeme.Lemma != entry.lemma || lexeme.ParadigmID != string(attestedNounParadigmID(entry.id, entry.gender)) {
			return nil, fmt.Errorf("Polish attested noun %s has no compatible lexical declaration", entry.id)
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, 14)}
		for caseIndex, grammaticalCase := range attestedNounCases {
			parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID("sg_" + grammaticalCase), Form: entry.singular[caseIndex]})
			parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID("pl_" + grammaticalCase), Form: entry.plural[caseIndex]})
		}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: parts})
	}
	return result, nil
}

func treeNounParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion:  morph.ParadigmSchemaVersion,
		ID:             "pl_noun_drzewo_attested",
		Language:       "pl",
		ProfileVersion: profile.Version(),
		POS:            "noun",
	}
	for _, number := range []string{"singular", "plural"} {
		prefix := "sg"
		if number == "plural" {
			prefix = "pl"
		}
		for _, grammaticalCase := range attestedNounCases {
			partID := morph.PrincipalPartID(prefix + "_" + grammaticalCase)
			stemID := procedure.StemID(partID)
			cellID := morph.CellID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
				ID: cellID,
				Features: []features.FeatureValue{
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
					{Feature: "gender", State: features.StateKnown, Value: "neuter"},
				},
			})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{
				CellID: cellID, ProcedureID: "procedure:pl_noun_drzewo_attested." + string(cellID),
				Provenance: treeNounFormSource,
				Steps:      []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
			})
		}
	}
	return paradigm
}

func treeNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	for _, lexeme := range lexemes {
		if lexeme.ID.String() != "lexeme:pl.drzewo.n" {
			continue
		}
		if lexeme.Lemma != "drzewo" || lexeme.ParadigmID != "pl_noun_drzewo_attested" {
			return nil, fmt.Errorf("Polish tree noun has an incompatible lexical declaration")
		}
		singular := [7]string{"drzewo", "drzewa", "drzewu", "drzewo", "drzewem", "drzewie", "drzewo"}
		plural := [7]string{"drzewa", "drzew", "drzewom", "drzewa", "drzewami", "drzewach", "drzewa"}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, 14)}
		for index, grammaticalCase := range attestedNounCases {
			parts.Parts = append(parts.Parts,
				morph.PrincipalPart{ID: morph.PrincipalPartID("sg_" + grammaticalCase), Form: singular[index]},
				morph.PrincipalPart{ID: morph.PrincipalPartID("pl_" + grammaticalCase), Form: plural[index]},
			)
		}
		return []lexemeRecord{{Lexeme: lexeme, PrincipalParts: parts}}, nil
	}
	return nil, fmt.Errorf("Polish tree noun lexeme is missing")
}

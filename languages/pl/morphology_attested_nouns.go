package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var attestedNounSource = provenance.SourceRef{
	ID:      "golingua-pl-noun-forms-authored",
	Path:    "languages/pl/morphology_attested_nouns.go",
	Version: "2026-10-22-reviewed-candidates-v1",
	License: "unknown",
	Author:  "golingua contributors",
}

// attestedNounForms contains every declared case/number form explicitly.
// Shared paradigms select these principal parts; they NEVER invent a suffix,
// sound change, missing form or an unreviewed alternate reading.
type attestedNounEntry struct {
	id, lemma, gender string
	singular, plural  [7]string
}

var attestedNounForms = []attestedNounEntry{
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
}

var attestedNounCases = [7]string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"}

func attestedNounGender(lexemeID string) (string, bool) {
	switch lexemeID {
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
	default:
		return "", false
	}
}

func attestedNounParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	genders := []string{"masculine_personal", "masculine_animal", "masculine_inanimate", "feminine", "neuter"}
	result := make([]morph.Paradigm, 0, len(genders))
	for _, gender := range genders {
		paradigm := morph.Paradigm{
			SchemaVersion:  morph.ParadigmSchemaVersion,
			ID:             morph.ParadigmID("pl_noun_attested_" + gender),
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
						{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(gender)},
					},
				})
				paradigm.Rules = append(paradigm.Rules, morph.CellRule{
					CellID:      cellID,
					ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID),
					Provenance:  attestedNounSource,
					Steps:       []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
				})
			}
		}
		result = append(result, paradigm)
	}
	return result
}

func attestedNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	indexed := make(map[string]lexicon.MorphLexeme, len(lexemes))
	for _, lexeme := range lexemes {
		indexed[lexeme.ID.String()] = lexeme
	}
	result := make([]lexemeRecord, 0, len(attestedNounForms))
	for _, entry := range attestedNounForms {
		lexeme, ok := indexed["lexeme:"+entry.id]
		if !ok || lexeme.Lemma != entry.lemma || lexeme.ParadigmID != "pl_noun_attested_"+entry.gender {
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

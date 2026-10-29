package pl

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var doroslyAdjectiveSource = provenance.SourceRef{
	ID:      "wsjp-pl-dorosly-adjective",
	URI:     "https://wsjp.pl/haslo/podglad/16860/dorosly/2823640/czlowiek",
	Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

type positiveAdjectiveEntry struct {
	lexemeID, lemma, class, masculinePersonalPlural string
	source                                          provenance.SourceRef
}

// WSJP positive-degree cells use separate hard -y and soft -ski ending sets.
// Masculine-personal plural nominatives are explicit per lexeme.
var positiveAdjectiveEntries = []positiveAdjectiveEntry{
	{"pl.czarny.adj", "czarny", "hard_y", "czarni", provenance.SourceRef{ID: "wsjp-pl-czarny-color", URI: "https://wsjp.pl/haslo/podglad/5086/czarny/4935813/kolor", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	{"pl.czerwony.adj", "czerwony", "hard_y", "czerwoni", provenance.SourceRef{ID: "wsjp-pl-czerwony-color", URI: "https://wsjp.pl/haslo/podglad/414/czerwony/4226346/kolor", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	{"pl.bialy.adj", "biały", "hard_y", "biali", provenance.SourceRef{ID: "wsjp-pl-bialy-color", URI: "https://wsjp.pl/haslo/podglad/8926/bialy/5144529/kolor", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	{"pl.niebieski.adj", "niebieski", "soft_i", "niebiescy", provenance.SourceRef{ID: "wsjp-pl-niebieski-color", URI: "https://wsjp.pl/haslo/podglad/416/niebieski/5008200/kolor", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	{"pl.zielony.adj", "zielony", "hard_y", "zieloni", provenance.SourceRef{ID: "wsjp-pl-zielony-color", URI: "https://wsjp.pl/haslo/podglad/417/zielony/5108628/kolor", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
	{"pl.otwarty.adj", "otwarty", "hard_y", "otwarci", provenance.SourceRef{ID: "golingua-pl-authored-otwarty-adjective-morphology", Path: "languages/pl/morphology_attested_adjectives.go", Version: "2026-10-28", License: "internal", Author: "Codex by owner delegation"}},
	{"pl.smutny.adj", "smutny", "hard_y", "smutni", provenance.SourceRef{ID: "wsjp-pl-smutny-affect", URI: "https://wsjp.pl/haslo/podglad/5774/smutny/4769515/chlopiec", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"}},
}

type polishAdjectiveForm struct {
	number, gender, grammaticalCase string
	positive, comparative           string
}

// The table follows the documented WSJP degree and agreement cells for
// adjective lexemes currently present in the Polish catalog.
var doroslyAdjectiveForms = []polishAdjectiveForm{
	{"singular", "masculine_personal", "nominative", "dorosły", "doroślejszy"},
	{"singular", "masculine_personal", "genitive", "dorosłego", "doroślejszego"},
	{"singular", "masculine_personal", "dative", "dorosłemu", "doroślejszemu"},
	{"singular", "masculine_personal", "accusative", "dorosłego", "doroślejszego"},
	{"singular", "masculine_personal", "instrumental", "dorosłym", "doroślejszym"},
	{"singular", "masculine_personal", "locative", "dorosłym", "doroślejszym"},
	{"singular", "masculine_personal", "vocative", "dorosły", ""},
	{"singular", "masculine_animal", "nominative", "dorosły", "doroślejszy"},
	{"singular", "masculine_animal", "genitive", "dorosłego", "doroślejszego"},
	{"singular", "masculine_animal", "dative", "dorosłemu", "doroślejszemu"},
	{"singular", "masculine_animal", "accusative", "dorosłego", "doroślejszego"},
	{"singular", "masculine_animal", "instrumental", "dorosłym", "doroślejszym"},
	{"singular", "masculine_animal", "locative", "dorosłym", "doroślejszym"},
	{"singular", "masculine_animal", "vocative", "dorosły", ""},
	{"singular", "masculine_inanimate", "nominative", "dorosły", "doroślejszy"},
	{"singular", "masculine_inanimate", "genitive", "dorosłego", "doroślejszego"},
	{"singular", "masculine_inanimate", "dative", "dorosłemu", "doroślejszemu"},
	{"singular", "masculine_inanimate", "accusative", "dorosły", "doroślejszy"},
	{"singular", "masculine_inanimate", "instrumental", "dorosłym", "doroślejszym"},
	{"singular", "masculine_inanimate", "locative", "dorosłym", "doroślejszym"},
	{"singular", "masculine_inanimate", "vocative", "dorosły", ""},
	{"singular", "neuter", "nominative", "dorosłe", "doroślejsze"},
	{"singular", "neuter", "genitive", "dorosłego", "doroślejszego"},
	{"singular", "neuter", "dative", "dorosłemu", "doroślejszemu"},
	{"singular", "neuter", "accusative", "dorosłe", "doroślejsze"},
	{"singular", "neuter", "instrumental", "dorosłym", "doroślejszym"},
	{"singular", "neuter", "locative", "dorosłym", "doroślejszym"},
	{"singular", "neuter", "vocative", "dorosłe", ""},
	{"singular", "feminine", "nominative", "dorosła", "doroślejsza"},
	{"singular", "feminine", "genitive", "dorosłej", "doroślejszej"},
	{"singular", "feminine", "dative", "dorosłej", "doroślejszej"},
	{"singular", "feminine", "accusative", "dorosłą", "doroślejszą"},
	{"singular", "feminine", "instrumental", "dorosłą", "doroślejszą"},
	{"singular", "feminine", "locative", "dorosłej", "doroślejszej"},
	{"singular", "feminine", "vocative", "dorosła", ""},
	{"plural", "masculine_personal", "nominative", "dorośli", "doroślejsi"},
	{"plural", "masculine_personal", "genitive", "dorosłych", "doroślejszych"},
	{"plural", "masculine_personal", "dative", "dorosłym", "doroślejszym"},
	{"plural", "masculine_personal", "accusative", "dorosłych", "doroślejszych"},
	{"plural", "masculine_personal", "instrumental", "dorosłymi", "doroślejszymi"},
	{"plural", "masculine_personal", "locative", "dorosłych", "doroślejszych"},
	{"plural", "masculine_personal", "vocative", "dorośli", ""},
	{"plural", "non_masculine_personal", "nominative", "dorosłe", "doroślejsze"},
	{"plural", "non_masculine_personal", "genitive", "dorosłych", "doroślejszych"},
	{"plural", "non_masculine_personal", "dative", "dorosłym", "doroślejszym"},
	{"plural", "non_masculine_personal", "accusative", "dorosłe", "doroślejsze"},
	{"plural", "non_masculine_personal", "instrumental", "dorosłymi", "doroślejszymi"},
	{"plural", "non_masculine_personal", "locative", "dorosłych", "doroślejszych"},
	{"plural", "non_masculine_personal", "vocative", "dorosłe", ""},
}

func doroslyAdjectiveParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: "pl_adjective_dorosly_attested",
		Language: "pl", ProfileVersion: profile.Version(), POS: "adjective",
	}
	seenParts := make(map[morph.PrincipalPartID]struct{})
	for _, form := range doroslyAdjectiveForms {
		partID := doroslyAdjectivePartID("positive", form.positive)
		stemID := procedure.StemID(partID)
		if _, exists := seenParts[partID]; !exists {
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			seenParts[partID] = struct{}{}
		}
		paradigm = appendDoroslyAdjectiveCell(paradigm, form, "positive", stemID, false)
		if form.comparative == "" {
			continue
		}
		comparativePart := doroslyAdjectivePartID("comparative", form.comparative)
		comparativeStem := procedure.StemID(comparativePart)
		if _, exists := seenParts[comparativePart]; !exists {
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: comparativeStem, PartID: comparativePart, Required: true,
			})
			seenParts[comparativePart] = struct{}{}
		}
		paradigm = appendDoroslyAdjectiveCell(paradigm, form, "comparative", comparativeStem, false)
		paradigm = appendDoroslyAdjectiveCell(paradigm, form, "superlative", comparativeStem, true)
	}
	return paradigm
}

func appendDoroslyAdjectiveCell(paradigm morph.Paradigm, form polishAdjectiveForm, degree string, stem procedure.StemID, superlative bool) morph.Paradigm {
	cellID := morph.CellID(doroslyAdjectiveCellID(degree, form))
	paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
		ID: cellID,
		Features: []features.FeatureValue{
			{Feature: "degree", State: features.StateKnown, Value: features.FeatureValueID(degree)},
			{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(form.number)},
			{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(form.gender)},
			{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(form.grammaticalCase)},
		},
	})
	steps := []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stem}}
	if superlative {
		steps = append(steps, procedure.Operation{Kind: procedure.OperationPrepend, Text: "naj"})
	}
	paradigm.Rules = append(paradigm.Rules, morph.CellRule{
		CellID: cellID, ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID),
		Provenance: doroslyAdjectiveSource, Steps: steps,
	})
	return paradigm
}

func doroslyAdjectivePartID(degree, form string) morph.PrincipalPartID {
	return morph.PrincipalPartID(degree + "_form_" + hex.EncodeToString([]byte(form)))
}

func doroslyAdjectiveCellID(degree string, form polishAdjectiveForm) string {
	number := "sg"
	if form.number == "plural" {
		number = "pl"
	}
	gender := map[string]string{
		"masculine_personal":     "mp",
		"masculine_animal":       "ma",
		"masculine_inanimate":    "mi",
		"non_masculine_personal": "nmp",
		"feminine":               "f",
		"neuter":                 "n",
	}[form.gender]
	grammaticalCase := map[string]string{
		"nominative": "nom", "genitive": "gen", "dative": "dat", "accusative": "acc",
		"instrumental": "inst", "locative": "loc", "vocative": "voc",
	}[form.grammaticalCase]
	return strings.Join([]string{degree, number, gender, grammaticalCase}, "_")
}

func doroslyAdjectiveRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	for _, lexeme := range lexemes {
		if lexeme.ID.String() != "lexeme:pl.dorosly.adj" {
			continue
		}
		if lexeme.Lemma != "dorosły" || lexeme.ParadigmID != "pl_adjective_dorosly_attested" || lexeme.POS != "adjective" {
			return nil, fmt.Errorf("Polish adjective dorosły has an incompatible lexical declaration")
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, len(doroslyAdjectiveForms)*2)}
		seenParts := make(map[morph.PrincipalPartID]struct{})
		for _, form := range doroslyAdjectiveForms {
			positivePart := doroslyAdjectivePartID("positive", form.positive)
			if _, exists := seenParts[positivePart]; !exists {
				parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: positivePart, Form: form.positive})
				seenParts[positivePart] = struct{}{}
			}
			if form.comparative != "" {
				comparativePart := doroslyAdjectivePartID("comparative", form.comparative)
				if _, exists := seenParts[comparativePart]; !exists {
					parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: comparativePart, Form: form.comparative})
					seenParts[comparativePart] = struct{}{}
				}
			}
		}
		return []lexemeRecord{{Lexeme: lexeme, PrincipalParts: parts}}, nil
	}
	return nil, fmt.Errorf("Polish adjective dorosły lexeme is missing")
}

func positiveAdjectiveForms(entry positiveAdjectiveEntry) []polishAdjectiveForm {
	root := entry.lemma
	switch entry.class {
	case "hard_y":
		root = strings.TrimSuffix(root, "y")
	case "soft_i":
		root = strings.TrimSuffix(root, "i")
	}
	suffixes := map[string]map[string]string{
		"hard_y": {"masculine_gen": "ego", "masculine_dat": "emu", "masculine_oblique": "ym", "neuter_nom": "e", "feminine_nom": "a", "feminine_oblique": "ej", "feminine_acc": "ą", "plural_nom": "e", "plural_gen": "ych", "plural_dat": "ym", "plural_instr": "ymi"},
		"soft_i": {"masculine_gen": "iego", "masculine_dat": "iemu", "masculine_oblique": "im", "neuter_nom": "ie", "feminine_nom": "a", "feminine_oblique": "iej", "feminine_acc": "ą", "plural_nom": "ie", "plural_gen": "ich", "plural_dat": "im", "plural_instr": "imi"},
	}[entry.class]
	form := func(key string) string { return root + suffixes[key] }
	forms := make([]polishAdjectiveForm, 0, 49)
	add := func(number, gender string, values [7]string) {
		for index, grammaticalCase := range attestedAdjectiveCases {
			forms = append(forms, polishAdjectiveForm{number: number, gender: gender, grammaticalCase: grammaticalCase, positive: values[index]})
		}
	}
	for _, gender := range []string{"masculine_personal", "masculine_animal", "masculine_inanimate"} {
		accusative := form("masculine_gen")
		if gender == "masculine_inanimate" {
			accusative = entry.lemma
		}
		add("singular", gender, [7]string{entry.lemma, form("masculine_gen"), form("masculine_dat"), accusative, form("masculine_oblique"), form("masculine_oblique"), entry.lemma})
	}
	add("singular", "neuter", [7]string{form("neuter_nom"), form("masculine_gen"), form("masculine_dat"), form("neuter_nom"), form("masculine_oblique"), form("masculine_oblique"), form("neuter_nom")})
	add("singular", "feminine", [7]string{form("feminine_nom"), form("feminine_oblique"), form("feminine_oblique"), form("feminine_acc"), form("feminine_acc"), form("feminine_oblique"), form("feminine_nom")})
	add("plural", "masculine_personal", [7]string{entry.masculinePersonalPlural, form("plural_gen"), form("plural_dat"), form("plural_gen"), form("plural_instr"), form("plural_gen"), entry.masculinePersonalPlural})
	add("plural", "non_masculine_personal", [7]string{form("plural_nom"), form("plural_gen"), form("plural_dat"), form("plural_nom"), form("plural_instr"), form("plural_gen"), form("plural_nom")})
	return forms
}

var attestedAdjectiveCases = []string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"}

func positiveAdjectiveParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := make([]morph.Paradigm, 0, len(positiveAdjectiveEntries))
	for _, entry := range positiveAdjectiveEntries {
		paradigmID := positiveAdjectiveParadigmID(entry.lexemeID)
		paradigm := morph.Paradigm{SchemaVersion: morph.ParadigmSchemaVersion, ID: paradigmID, Language: "pl", ProfileVersion: profile.Version(), POS: "adjective"}
		for _, form := range positiveAdjectiveForms(entry) {
			cellID := morph.CellID(doroslyAdjectiveCellID("positive", form))
			partID := positiveAdjectivePartID(cellID)
			stemID := procedure.StemID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{StemID: stemID, PartID: partID, Required: true})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{ID: cellID, Features: []features.FeatureValue{
				{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(form.grammaticalCase)},
				{Feature: "degree", State: features.StateKnown, Value: "positive"},
				{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(form.gender)},
				{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(form.number)},
			}})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{CellID: cellID, ProcedureID: "procedure:" + string(paradigmID) + "." + string(cellID), Provenance: entry.source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}}})
		}
		result = append(result, paradigm)
	}
	return result
}

func positiveAdjectiveParadigmID(lexemeID string) morph.ParadigmID {
	return morph.ParadigmID("pl_adjective_positive_" + strings.ReplaceAll(strings.TrimPrefix(lexemeID, "pl."), ".adj", ""))
}

func positiveAdjectivePartID(cellID morph.CellID) morph.PrincipalPartID {
	return morph.PrincipalPartID("positive_form_" + hex.EncodeToString([]byte(cellID)))
}

func positiveAdjectiveRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	byID := make(map[string]lexicon.MorphLexeme, len(lexemes))
	for _, lexeme := range lexemes {
		byID[lexeme.ID.String()] = lexeme
	}
	result := make([]lexemeRecord, 0, len(positiveAdjectiveEntries))
	for _, entry := range positiveAdjectiveEntries {
		lexeme, ok := byID["lexeme:"+entry.lexemeID]
		if !ok || lexeme.Lemma != entry.lemma || lexeme.POS != "adjective" || lexeme.ParadigmID != string(positiveAdjectiveParadigmID(entry.lexemeID)) {
			return nil, fmt.Errorf("Polish adjective %s has a missing or incompatible lexical declaration", entry.lemma)
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID}
		for _, form := range positiveAdjectiveForms(entry) {
			cellID := morph.CellID(doroslyAdjectiveCellID("positive", form))
			parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: positiveAdjectivePartID(cellID), Form: form.positive})
		}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: parts})
	}
	return result, nil
}

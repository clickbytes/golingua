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

type polishAdjectiveForm struct {
	number, gender, grammaticalCase string
	positive, comparative           string
}

// The table follows the documented WSJP degree and agreement cells for the
// single adjective lexeme currently present in the Polish catalog.
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

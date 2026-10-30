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

var polishOwnershipNounMorphSource = provenance.SourceRef{
	ID: "golingua-pl-wlasciciel-instrumental-476", URI: "https://github.com/clickbytes/golingua/issues/476",
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
	result = append(result, polishOwnershipPredicateParadigm(profile))
	return append(result, polishPersonalParadigms(profile)...)
}

func polishOwnershipPredicateParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	const id = "pl_noun_ownership_predicate"
	stem := procedure.StemID(id + "_stem")
	cell := morph.CellID("singular_instrumental")
	return morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(id), Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
		PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: stem, PartID: "stem", Required: true}},
		Cells: []morph.ParadigmCell{{ID: cell, Features: []features.FeatureValue{
			{Feature: "number", State: features.StateKnown, Value: "singular"},
			{Feature: "case", State: features.StateKnown, Value: "instrumental"},
			{Feature: "gender", State: features.StateKnown, Value: "masculine_personal"},
		}}},
		Rules: []morph.CellRule{{CellID: cell, ProcedureID: "procedure:" + id + ".singular_instrumental", Provenance: polishOwnershipNounMorphSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stem}, {Kind: procedure.OperationAppend, Text: "em"}},
		}},
	}
}

func inCodeMorphRecords() ([]lexemeRecord, error) {
	authored, err := decodePolishLexemeDocument(lexemesJSON)
	if err != nil {
		return nil, err
	}
	authoredByID := make(map[string]lexicon.MorphLexeme, len(authored.Records))
	for _, record := range authored.Records {
		authoredByID[record.Lexeme.ID.String()] = record.Lexeme
	}
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	var records []lexemeRecord
	attested, err := attestedNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, attested...)
	pluralOnly, err := pluralOnlyNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, pluralOnly...)
	treeNouns, err := treeNounRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, treeNouns...)
	dorosly, err := doroslyAdjectiveRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, dorosly...)
	positiveAdjectives, err := positiveAdjectiveRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, positiveAdjectives...)
	identification, err := identificationMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, identification...)
	personal, err := polishPersonalMorphRecords(document.Lexemes)
	if err != nil {
		return nil, err
	}
	records = append(records, personal...)
	foundOwner, foundCat, foundBird := false, false, false
	for _, lexeme := range document.Lexemes {
		switch lexeme.ID.String() {
		case "lexeme:pl.wlasciciel.n":
			foundOwner = true
			records = append(records, lexemeRecord{Lexeme: lexeme, PrincipalParts: morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{{ID: "stem", Form: "właściciel"}}}})
		case "lexeme:pl.kot.n":
			foundCat = true
		case "lexeme:pl.ptak.n":
			foundBird = true
		}
	}
	if !foundOwner || !foundCat || !foundBird {
		return nil, fmt.Errorf("Polish in-code noun morphology lexeme is missing")
	}
	filtered := make([]lexemeRecord, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		id := record.Lexeme.ID.String()
		if dataOwned, exists := authoredByID[id]; exists {
			if !samePolishMorphLexicalIdentity(record.Lexeme, dataOwned) {
				return nil, fmt.Errorf("Polish authored morphology lexeme %q conflicts with lexical identity", id)
			}
			continue
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("Polish in-code morphology repeats lexeme ID %q", id)
		}
		seen[id] = struct{}{}
		filtered = append(filtered, record)
	}
	return filtered, nil
}

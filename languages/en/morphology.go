package en

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var sentenceMorphSource = provenance.SourceRef{
	ID: "golingua-en-sentence-forms-authored", Path: "languages/en/morphology.go",
	Version: "2026-10-23", License: "internal", Author: "golingua contributors",
}

func inCodeParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	makeParadigm := func(id, pos, feature, value string) morph.Paradigm {
		return morph.Paradigm{
			SchemaVersion: morph.ParadigmSchemaVersion, ID: morph.ParadigmID(id), Language: "en", ProfileVersion: profile.Version(), POS: lexicon.PartOfSpeech(pos),
			PrincipalPartRequirements: []morph.PrincipalPartRequirement{{StemID: "form", PartID: "attested", Required: true}},
			Cells:                     []morph.ParadigmCell{{ID: "attested", Features: []features.FeatureValue{{Feature: features.FeatureID(feature), State: features.StateKnown, Value: features.FeatureValueID(value)}}}},
			Rules: []morph.CellRule{{CellID: "attested", ProcedureID: "procedure:" + id + ".attested", Provenance: sentenceMorphSource,
				Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: "form"}}}},
		}
	}
	return []morph.Paradigm{
		makeParadigm("en_noun_singular_attested", "noun", "noun_number", "singular"),
		makeParadigm("en_see_present_third_singular", "verb", "verb_form", "third_person_singular"),
		makeParadigm("en_determiner_definite", "determiner", "determiner", "definite"),
	}
}

// inCodeMorphRecords uses a regular seed only where its spelling rule applies.
// The sentence slice adds explicit singular nouns and one attested verb form;
// missing English forms remain unavailable.
func inCodeMorphRecords() ([]lexemeRecord, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	records := make([]lexemeRecord, 0, len(englishParadigmByLexeme))
	for _, lexeme := range document.Lexemes {
		if lexeme.ParadigmID == "" {
			continue
		}
		part := morph.PrincipalPart{ID: "base", Form: lexeme.Lemma}
		if lexeme.ParadigmID == "en_noun_singular_attested" || lexeme.ParadigmID == "en_determiner_definite" {
			part.ID = "attested"
		}
		if lexeme.ParadigmID == "en_see_present_third_singular" {
			part = morph.PrincipalPart{ID: "attested", Form: "sees"}
		}
		records = append(records, lexemeRecord{
			Lexeme:         lexeme,
			PrincipalParts: morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{part}},
		})
	}
	return records, nil
}

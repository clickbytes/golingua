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
		makeParadigm("en_determiner_definite", "determiner", "determiner", "definite"),
		makeParadigm("en_determiner_indefinite", "determiner", "determiner", "indefinite"),
	}
}

// inCodeMorphRecords uses a regular seed only where its spelling rule applies.
// The sentence slice adds explicit singular nouns and attested verb forms;
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
		if lexeme.ParadigmID == "en_noun_singular_attested" || lexeme.ParadigmID == "en_determiner_definite" || lexeme.ParadigmID == "en_determiner_indefinite" {
			part.ID = "attested"
		}
		if lexeme.ID.String() == "lexeme:en.a.det" {
			part = morph.PrincipalPart{ID: "attested", Form: "a"}
		}
		if lexeme.ID.String() == "lexeme:en.an.det" {
			part = morph.PrincipalPart{ID: "attested", Form: "an"}
		}
		principalParts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{part}}
		switch lexeme.ID.String() {
		case "lexeme:en.see.v":
			principalParts = finiteEnglishPrincipalParts(lexeme, "see", "sees", "saw", "seen", "seeing")
		case "lexeme:en.hear.v":
			principalParts = finiteEnglishPrincipalParts(lexeme, "hear", "hears", "heard", "heard", "hearing")
		case "lexeme:en.look.v":
			principalParts = finiteEnglishPrincipalParts(lexeme, "look", "looks", "looked", "looked", "looking")
		case "lexeme:en.get.v":
			principalParts = finiteEnglishPrincipalParts(lexeme, "get", "gets", "got", "gotten", "getting")
		case "lexeme:en.read.v":
			principalParts = finiteEnglishPrincipalParts(lexeme, "read", "reads", "read", "read", "reading")
		}
		if lexeme.ParadigmID == "en_noun_principal_parts" {
			principalParts = englishNounPrincipalParts(lexeme)
		}
		records = append(records, lexemeRecord{
			Lexeme:         lexeme,
			PrincipalParts: principalParts,
		})
	}
	return records, nil
}

func englishNounPrincipalParts(lexeme lexicon.MorphLexeme) morph.PrincipalParts {
	pluralByLexeme := map[string]string{
		"lexeme:en.child.n":         "children",
		"lexeme:en.person.n":        "people",
		"lexeme:en.mouse_animal.n":  "mice",
		"lexeme:en.goose.n":         "geese",
		"lexeme:en.foot.n":          "feet",
		"lexeme:en.tooth.n":         "teeth",
		"lexeme:en.leaf_plant.n":    "leaves",
		"lexeme:en.knife_utensil.n": "knives",
	}
	plural := pluralByLexeme[lexeme.ID.String()]
	if plural == "" {
		return morph.PrincipalParts{LexemeID: lexeme.ID}
	}
	return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
		{ID: "singular", Form: lexeme.Lemma},
		{ID: "plural", Form: plural},
	}}
}

func finiteEnglishPrincipalParts(lexeme lexicon.MorphLexeme, base, thirdPersonSingular, past, pastParticiple, presentParticiple string) morph.PrincipalParts {
	return morph.PrincipalParts{LexemeID: lexeme.ID, Parts: []morph.PrincipalPart{
		{ID: "base", Form: base},
		{ID: "third_person_singular", Form: thirdPersonSingular},
		{ID: "past", Form: past},
		{ID: "past_participle", Form: pastParticiple},
		{ID: "present_participle", Form: presentParticiple},
	}}
}

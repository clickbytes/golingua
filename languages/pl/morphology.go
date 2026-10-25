package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/morph"
)

// Only explicitly declared noun and adjective forms enter these paradigms.
func inCodeParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	result := attestedNounParadigms(profile)
	result = append(result, pluralOnlyNounParadigm(profile))
	result = append(result, treeNounParadigm(profile))
	result = append(result, doroslyAdjectiveParadigm(profile))
	return append(result, identificationParadigms(profile)...)
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

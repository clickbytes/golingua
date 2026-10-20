// Package en loads the small, bundled English morphology seed.
package en

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

//go:embed features.json
var profileJSON []byte

//go:embed manifest.json
var manifestJSON []byte

//go:embed morphology/paradigms.json
var paradigmsJSON []byte

//go:embed morphology/lexemes.json
var lexemesJSON []byte

type lexemeRecord struct {
	Lexeme         lexicon.MorphLexeme  `json:"lexeme"`
	PrincipalParts morph.PrincipalParts `json:"principal_parts"`
}

type lexemeDocument struct {
	SchemaVersion int            `json:"schema_version"`
	Records       []lexemeRecord `json:"records"`
}

type manifest struct {
	SchemaVersion int    `json:"schema_version"`
	DatasetID     string `json:"dataset_id"`
	Language      string `json:"language"`
	Profile       string `json:"profile"`
	Morphology    struct {
		Paradigms string `json:"paradigms"`
		Lexemes   string `json:"lexemes"`
	} `json:"morphology"`
	Capabilities struct {
		Generate           string `json:"morphology_generate"`
		Analyze            string `json:"morphology_analyze"`
		SentenceAnalysis   string `json:"sentence_analysis"`
		MeaningRealization string `json:"meaning_realization"`
	} `json:"capabilities"`
}

// Seed is an immutable, compiled view of the English starter data.
type Seed struct {
	profile    *features.LanguageFeatureProfile
	lexicon    *lexicon.Index
	lexemes    []lexicon.MorphLexeme
	paradigms  *morph.CompiledParadigms
	exceptions *morph.CompiledExceptions
	provenance *provenance.Index
	parts      map[string]morph.PrincipalParts
	index      *morph.InverseIndex
}

// Load strictly decodes the bundled profiles, paradigms and lexical records,
// then builds the finite relation table used by Analyze.
func Load() (*Seed, error) {
	var packageManifest manifest
	manifestDecoder := json.NewDecoder(bytes.NewReader(manifestJSON))
	manifestDecoder.DisallowUnknownFields()
	if err := manifestDecoder.Decode(&packageManifest); err != nil {
		return nil, fmt.Errorf("English manifest: %w", err)
	}
	var manifestTrailing any
	if err := manifestDecoder.Decode(&manifestTrailing); err != io.EOF {
		return nil, fmt.Errorf("English manifest: expected exactly one JSON value")
	}
	if packageManifest.SchemaVersion != 1 || packageManifest.DatasetID != "golingua-en-seed-v1" || packageManifest.Language != "en" || packageManifest.Profile != "features.json" || packageManifest.Morphology.Paradigms != "morphology/paradigms.json" || packageManifest.Morphology.Lexemes != "morphology/lexemes.json" || packageManifest.Capabilities.Generate != "implemented" || packageManifest.Capabilities.Analyze != "implemented" || packageManifest.Capabilities.SentenceAnalysis != "unsupported" || packageManifest.Capabilities.MeaningRealization != "unsupported" {
		return nil, fmt.Errorf("English manifest is inconsistent with the bundled morphology package")
	}
	profile, err := features.DecodeProfile(profileJSON)
	if err != nil {
		return nil, fmt.Errorf("English feature profile: %w", err)
	}
	if profile.Language() != "en" {
		return nil, fmt.Errorf("English feature profile has language %q", profile.Language())
	}
	paradigms, err := morph.DecodeParadigms(paradigmsJSON)
	if err != nil {
		return nil, fmt.Errorf("English paradigms: %w", err)
	}
	compiledParadigms, err := morph.CompileParadigms(paradigms, profile, morph.DefaultCompilerBudget(), procedure.DefaultBudget())
	if err != nil {
		return nil, fmt.Errorf("compile English paradigms: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(lexemesJSON))
	decoder.DisallowUnknownFields()
	var document lexemeDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("English lexemes: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("English lexemes: expected exactly one JSON value")
	}
	if document.SchemaVersion != 1 || document.Records == nil || len(document.Records) == 0 || len(document.Records) > 128 {
		return nil, fmt.Errorf("English lexeme document has unsupported schema or record count")
	}
	lexemes := make([]lexicon.MorphLexeme, len(document.Records))
	parts := make(map[string]morph.PrincipalParts, len(document.Records))
	for index, record := range document.Records {
		if record.Lexeme.Language != "en" || record.PrincipalParts.LexemeID != record.Lexeme.ID {
			return nil, fmt.Errorf("English lexeme record %d has inconsistent language or principal-parts identity", index)
		}
		lexemes[index] = record.Lexeme
		parts[record.Lexeme.ID.String()] = record.PrincipalParts
	}
	compiledLexicon, err := lexicon.Build(lexicon.Document{SchemaVersion: lexicon.SchemaVersion, Lexemes: lexemes, Senses: []lexicon.LexicalSense{}, Bindings: []lexicon.SemanticBinding{}}, nil, lexicon.DefaultBudget())
	if err != nil {
		return nil, fmt.Errorf("compile English lexemes: %w", err)
	}
	exceptions, err := morph.CompileExceptions([]morph.ExceptionRule{}, profile)
	if err != nil {
		return nil, fmt.Errorf("compile English exceptions: %w", err)
	}
	provenanceRecords := make([]provenance.RecordRef, 0)
	for _, paradigm := range paradigms {
		for _, rule := range paradigm.Rules {
			if rule.ProcedureID != "" && rule.Provenance.ID != "" {
				provenanceRecords = append(provenanceRecords, provenance.RecordRef{Kind: provenance.RecordProcedure, ID: rule.ProcedureID, Source: rule.Provenance})
			}
		}
	}
	provenanceIndex, err := provenance.BuildIndex(provenanceRecords)
	if err != nil {
		return nil, fmt.Errorf("build English provenance index: %w", err)
	}
	seed := &Seed{profile: profile, lexicon: compiledLexicon, lexemes: lexemes, paradigms: compiledParadigms, exceptions: exceptions, provenance: provenanceIndex, parts: parts}
	readings := make([]morph.Reading, 0, 64)
	for _, sourceLexeme := range lexemes {
		lexeme, exists := compiledLexicon.LookupLexeme(sourceLexeme.ID)
		if !exists {
			return nil, fmt.Errorf("English lexeme %q did not compile", sourceLexeme.ID.String())
		}
		paradigm, exists := compiledParadigms.Lookup(morph.ParadigmID(lexeme.ParadigmID))
		if !exists {
			return nil, fmt.Errorf("English lexeme %q refers to an unavailable paradigm", lexeme.ID.String())
		}
		for _, cell := range paradigm.Cells {
			bundle, err := features.NewBundle(profile, cell.Features)
			if err != nil {
				return nil, fmt.Errorf("English cell %q feature bundle: %w", cell.ID, err)
			}
			result, err := morph.Generate(lexeme, parts[lexeme.ID.String()], bundle, compiledParadigms, exceptions, procedure.DefaultBudget())
			if err != nil {
				return nil, fmt.Errorf("generate English %q cell %q: %w", lexeme.Lemma, cell.ID, err)
			}
			if result.State != morph.GenerationGenerated || len(result.Forms) == 0 {
				return nil, fmt.Errorf("English seed %q cell %q did not generate a form (state %q)", lexeme.Lemma, cell.ID, result.State)
			}
			for _, form := range result.Forms {
				readings = append(readings, form.Reading())
			}
		}
	}
	table, err := morph.BuildFiniteRelationTable(readings, profile, morph.DefaultFiniteTableBudget())
	if err != nil {
		return nil, fmt.Errorf("build English finite relation table: %w", err)
	}
	seed.index, err = morph.BuildInverseIndex(table)
	if err != nil {
		return nil, fmt.Errorf("build English inverse index: %w", err)
	}
	return seed, nil
}

// Generate resolves one bundled lexeme by stable ID and a language-local
// feature request.
func (seed *Seed) Generate(lexemeID string, values []features.FeatureValue) (morph.GenerationResult, error) {
	if seed == nil || seed.profile == nil || seed.lexicon == nil {
		return morph.GenerationResult{}, fmt.Errorf("English seed is unavailable")
	}
	parsedID, err := ids.ParseLexemeID(lexemeID)
	if err != nil || parsedID.String() != lexemeID {
		return morph.GenerationResult{}, fmt.Errorf("invalid English lexeme ID")
	}
	lexeme, exists := seed.lexicon.LookupLexeme(parsedID)
	if !exists {
		return morph.GenerationResult{State: morph.GenerationUnknown, Forms: []morph.GeneratedForm{}}, nil
	}
	bundle, err := features.NewBundle(seed.profile, values)
	if err != nil {
		return morph.GenerationResult{}, err
	}
	return morph.Generate(lexeme, seed.parts[lexemeID], bundle, seed.paradigms, seed.exceptions, procedure.DefaultBudget())
}

// Analyze returns every licensed reading for an English surface in the seed.
func (seed *Seed) Analyze(surface string) ([]morph.Reading, error) {
	if seed == nil || seed.index == nil {
		return nil, fmt.Errorf("English seed is unavailable")
	}
	return seed.index.Analyze(surface)
}

// Lexemes returns the immutable seed's lexemes as defensive copies.
func (seed *Seed) Lexemes() []lexicon.MorphLexeme {
	if seed == nil || seed.lexicon == nil {
		return nil
	}
	result := make([]lexicon.MorphLexeme, len(seed.lexemes))
	for index, lexeme := range seed.lexemes {
		result[index] = lexeme
		result[index].Properties = append([]lexicon.Property(nil), lexeme.Properties...)
	}
	return result
}

// Provenance returns the immutable source index for bundled procedures.
func (seed *Seed) Provenance() *provenance.Index {
	if seed == nil {
		return nil
	}
	return seed.provenance
}

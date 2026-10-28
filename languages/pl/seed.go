// Package pl loads the small, bundled Polish morphology seed.
package pl

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

//go:embed morphology/exceptions.json
var exceptionsJSON []byte

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
		Paradigms  string `json:"paradigms"`
		Lexemes    string `json:"lexemes"`
		Exceptions string `json:"exceptions"`
	} `json:"morphology"`
	Capabilities struct {
		Generate           string `json:"morphology_generate"`
		Analyze            string `json:"morphology_analyze"`
		SentenceAnalysis   string `json:"sentence_analysis"`
		MeaningRealization string `json:"meaning_realization"`
	} `json:"capabilities"`
}

// Seed is an immutable, compiled view of the Polish starter data.
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

// Load strictly decodes the bundled profile, manifest, paradigms and lexical
// records, then builds the finite relation table for inverse analysis.
func Load() (*Seed, error) {
	return LoadWithExceptions(nil)
}

// LoadWithExceptions builds the bundled Polish seed with additional explicit
// morphology exceptions. The supplied rules are copied into the compiled
// exception set and are not used to mutate the bundled source data.
func LoadWithExceptions(extra []morph.ExceptionRule) (*Seed, error) {
	var packageManifest manifest
	manifestDecoder := json.NewDecoder(bytes.NewReader(manifestJSON))
	manifestDecoder.DisallowUnknownFields()
	if err := manifestDecoder.Decode(&packageManifest); err != nil {
		return nil, fmt.Errorf("Polish manifest: %w", err)
	}
	var manifestTrailing any
	if err := manifestDecoder.Decode(&manifestTrailing); err != io.EOF {
		return nil, fmt.Errorf("Polish manifest: expected exactly one JSON value")
	}
	if packageManifest.SchemaVersion != 1 || packageManifest.DatasetID != "golingua-pl-seed-v1" || packageManifest.Language != "pl" || packageManifest.Profile != "features.json" || packageManifest.Morphology.Paradigms != "morphology/paradigms.json" || packageManifest.Morphology.Lexemes != "morphology/lexemes.json" || packageManifest.Morphology.Exceptions != "morphology/exceptions.json" || packageManifest.Capabilities.Generate != "implemented" || packageManifest.Capabilities.Analyze != "implemented" || packageManifest.Capabilities.SentenceAnalysis != "unsupported" || packageManifest.Capabilities.MeaningRealization != "unsupported" {
		return nil, fmt.Errorf("Polish manifest is inconsistent with the bundled morphology package")
	}
	profile, err := features.DecodeProfile(profileJSON)
	if err != nil {
		return nil, fmt.Errorf("Polish feature profile: %w", err)
	}
	if profile.Language() != "pl" {
		return nil, fmt.Errorf("Polish feature profile has language %q", profile.Language())
	}
	paradigms, err := morph.DecodeParadigms(paradigmsJSON)
	if err != nil {
		return nil, fmt.Errorf("Polish paradigms: %w", err)
	}
	paradigms = append(paradigms, inCodeParadigms(profile)...)
	compiledParadigms, err := morph.CompileParadigms(paradigms, profile, morph.DefaultCompilerBudget(), procedure.DefaultBudget())
	if err != nil {
		return nil, fmt.Errorf("compile Polish paradigms: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(lexemesJSON))
	decoder.DisallowUnknownFields()
	var document lexemeDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("Polish lexemes: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("Polish lexemes: expected exactly one JSON value")
	}
	if document.SchemaVersion != 1 || document.Records == nil || len(document.Records) == 0 || len(document.Records) > 128 {
		return nil, fmt.Errorf("Polish lexeme document has unsupported schema or record count")
	}
	additionalRecords, err := inCodeMorphRecords()
	if err != nil {
		return nil, err
	}
	if len(document.Records)+len(additionalRecords) > 128 {
		return nil, fmt.Errorf("Polish lexeme count exceeds seed limit")
	}
	document.Records = append(document.Records, additionalRecords...)
	lexemes := make([]lexicon.MorphLexeme, len(document.Records))
	parts := make(map[string]morph.PrincipalParts, len(document.Records))
	for index, record := range document.Records {
		if record.Lexeme.Language != "pl" || record.PrincipalParts.LexemeID != record.Lexeme.ID {
			return nil, fmt.Errorf("Polish lexeme record %d has inconsistent language or principal-parts identity", index)
		}
		lexemes[index] = record.Lexeme
		parts[record.Lexeme.ID.String()] = record.PrincipalParts
	}
	compiledLexicon, err := lexicon.Build(lexicon.Document{SchemaVersion: lexicon.SchemaVersion, Lexemes: lexemes, Senses: []lexicon.LexicalSense{}, Bindings: []lexicon.SemanticBinding{}}, nil, lexicon.DefaultBudget())
	if err != nil {
		return nil, fmt.Errorf("compile Polish lexemes: %w", err)
	}
	exceptionRules, err := morph.DecodeExceptions(exceptionsJSON)
	if err != nil {
		return nil, fmt.Errorf("Polish exceptions: %w", err)
	}
	combinedExceptionRules := make([]morph.ExceptionRule, 0, len(exceptionRules)+len(extra))
	combinedExceptionRules = append(combinedExceptionRules, exceptionRules...)
	combinedExceptionRules = append(combinedExceptionRules, extra...)
	exceptions, err := morph.CompileExceptions(combinedExceptionRules, profile)
	if err != nil {
		return nil, fmt.Errorf("compile Polish exceptions: %w", err)
	}
	provenanceRecords := make([]provenance.RecordRef, 0, len(combinedExceptionRules))
	for _, paradigm := range paradigms {
		for _, rule := range paradigm.Rules {
			if rule.ProcedureID != "" && rule.Provenance.ID != "" {
				provenanceRecords = append(provenanceRecords, provenance.RecordRef{Kind: provenance.RecordProcedure, ID: rule.ProcedureID, Source: rule.Provenance})
			}
		}
	}
	for _, rule := range combinedExceptionRules {
		provenanceRecords = append(provenanceRecords, provenance.RecordRef{Kind: provenance.RecordException, ID: string(rule.ID), Source: rule.Provenance})
	}
	provenanceIndex, err := provenance.BuildIndex(provenanceRecords)
	if err != nil {
		return nil, fmt.Errorf("build Polish provenance index: %w", err)
	}
	seed := &Seed{profile: profile, lexicon: compiledLexicon, lexemes: lexemes, paradigms: compiledParadigms, exceptions: exceptions, provenance: provenanceIndex, parts: parts}
	readings := make([]morph.Reading, 0, 32)
	for _, sourceLexeme := range lexemes {
		lexeme, exists := compiledLexicon.LookupLexeme(sourceLexeme.ID)
		if !exists {
			return nil, fmt.Errorf("Polish lexeme %q did not compile", sourceLexeme.ID.String())
		}
		paradigm, exists := compiledParadigms.Lookup(morph.ParadigmID(lexeme.ParadigmID))
		if !exists {
			return nil, fmt.Errorf("Polish lexeme %q refers to an unavailable paradigm", lexeme.ID.String())
		}
		for _, cell := range paradigm.Cells {
			bundle, err := features.NewBundle(profile, cell.Features)
			if err != nil {
				return nil, fmt.Errorf("Polish cell %q feature bundle: %w", cell.ID, err)
			}
			result, err := morph.Generate(lexeme, parts[lexeme.ID.String()], bundle, compiledParadigms, exceptions, procedure.DefaultBudget())
			if err != nil {
				return nil, fmt.Errorf("generate Polish %q cell %q: %w", lexeme.Lemma, cell.ID, err)
			}
			if result.State == morph.GenerationBlocked {
				continue
			}
			if result.State != morph.GenerationGenerated || len(result.Forms) == 0 {
				return nil, fmt.Errorf("Polish seed %q cell %q did not generate a form (state %q)", lexeme.Lemma, cell.ID, result.State)
			}
			for _, form := range result.Forms {
				readings = append(readings, form.Reading())
			}
		}
	}
	table, err := morph.BuildFiniteRelationTable(readings, profile, morph.DefaultFiniteTableBudget())
	if err != nil {
		return nil, fmt.Errorf("build Polish finite relation table: %w", err)
	}
	seed.index, err = morph.BuildInverseIndex(table)
	if err != nil {
		return nil, fmt.Errorf("build Polish inverse index: %w", err)
	}
	return seed, nil
}

// Generate resolves one bundled lexeme by stable ID and a language-local
// feature request.
func (seed *Seed) Generate(lexemeID string, values []features.FeatureValue) (morph.GenerationResult, error) {
	if seed == nil || seed.profile == nil || seed.lexicon == nil {
		return morph.GenerationResult{}, fmt.Errorf("Polish seed is unavailable")
	}
	parsedID, err := ids.ParseLexemeID(lexemeID)
	if err != nil || parsedID.String() != lexemeID {
		return morph.GenerationResult{}, fmt.Errorf("invalid Polish lexeme ID")
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

// Analyze returns every licensed reading for a Polish surface in the seed.
func (seed *Seed) Analyze(surface string) ([]morph.Reading, error) {
	if seed == nil || seed.index == nil {
		return nil, fmt.Errorf("Polish seed is unavailable")
	}
	return seed.index.Analyze(surface)
}

// RuntimeMorphology exposes immutable compiled indexes to the trusted Go
// sentence adapter. The profile and indexes have no mutating API.
func (seed *Seed) RuntimeMorphology() (*features.LanguageFeatureProfile, *morph.InverseIndex, *morph.CompiledParadigms, *morph.CompiledExceptions) {
	if seed == nil {
		return nil, nil, nil, nil
	}
	return seed.profile, seed.index, seed.paradigms, seed.exceptions
}

// PrincipalParts returns a defensive copy for one attested lexeme.
func (seed *Seed) PrincipalParts(id string) (morph.PrincipalParts, bool) {
	if seed == nil {
		return morph.PrincipalParts{}, false
	}
	parts, found := seed.parts[id]
	parts.Parts = append([]morph.PrincipalPart(nil), parts.Parts...)
	return parts, found
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

// Provenance returns the immutable source index for bundled procedures and exceptions.
func (seed *Seed) Provenance() *provenance.Index {
	if seed == nil {
		return nil
	}
	return seed.provenance
}

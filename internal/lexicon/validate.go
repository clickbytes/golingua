package lexicon

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxDocumentBytes = 4 << 20
	MaxLexemes       = 100_000
	MaxSenses        = 250_000
	MaxBindings      = 250_000
	MaxProperties    = 64
	MaxUsageNotes    = 32
	MaxValencyRefs   = 64
	MaxTextBytes     = 1_024
	MaxFeatureBytes  = 128
)

// Budget bounds one lexical snapshot. Limits cannot exceed package ceilings.
type Budget struct {
	MaxLexemes  int
	MaxSenses   int
	MaxBindings int
	MaxBytes    int
}

func DefaultBudget() Budget {
	return Budget{MaxLexemes: MaxLexemes, MaxSenses: MaxSenses, MaxBindings: MaxBindings, MaxBytes: MaxDocumentBytes}
}

func (budget Budget) Validate() error {
	if budget.MaxLexemes <= 0 || budget.MaxLexemes > MaxLexemes || budget.MaxSenses <= 0 || budget.MaxSenses > MaxSenses || budget.MaxBindings <= 0 || budget.MaxBindings > MaxBindings || budget.MaxBytes <= 0 || budget.MaxBytes > MaxDocumentBytes {
		return invalid("INVALID_LEXICON_BUDGET", "budget", "lexicon limits must be positive and cannot exceed package ceilings")
	}
	return nil
}

// LexiconError retains a typed outcome and bounded diagnostic.
type LexiconError struct{ result status.Result }

func (failure *LexiconError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "lexicon operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *LexiconError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func invalid(code, path, message string) error {
	return lexiconError(status.OutcomeInvalid, code, path, message)
}

func unknown(code, path, message string) error {
	return lexiconError(status.OutcomeUnknown, code, path, message)
}

func resourceLimit(resource string, value, maximum int) error {
	return lexiconError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("lexicon %s limit exceeded: %d > %d", resource, value, maximum))
}

func lexiconError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &LexiconError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// ValidateDocument checks shape and all cross-record references. A CORE graph
// is needed only when the document contains semantic bindings; an unmapped
// morphology lexicon is valid without CORE.
func ValidateDocument(document Document, ontology *core.ConceptGraph, budget Budget) error {
	if err := budget.Validate(); err != nil {
		return err
	}
	if document.SchemaVersion != SchemaVersion {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported lexicon schema version")
	}
	if len(document.Lexemes) == 0 {
		return invalid("EMPTY_LEXICON", "lexemes", "lexicon requires at least one morph lexeme")
	}
	if len(document.Lexemes) > budget.MaxLexemes {
		return resourceLimit("lexemes", len(document.Lexemes), budget.MaxLexemes)
	}
	if len(document.Senses) > budget.MaxSenses {
		return resourceLimit("senses", len(document.Senses), budget.MaxSenses)
	}
	if len(document.Bindings) > budget.MaxBindings {
		return resourceLimit("bindings", len(document.Bindings), budget.MaxBindings)
	}
	meter := documentMeter{maximum: budget.MaxBytes, used: 64}
	lexemes := make(map[string]MorphLexeme, len(document.Lexemes))
	for index, lexeme := range document.Lexemes {
		path := fmt.Sprintf("lexemes[%d]", index)
		if err := meter.add(80, lexeme.ID.String(), string(lexeme.Language), lexeme.Lemma, string(lexeme.POS), lexeme.ParadigmID); err != nil {
			return err
		}
		if !validTypedLexemeID(lexeme.ID) {
			return invalid("INVALID_LEXEME_ID", path+".id", "lexeme ID must use the lexeme namespace")
		}
		key := lexeme.ID.String()
		if _, exists := lexemes[key]; exists {
			return invalid("DUPLICATE_LEXEME_ID", path+".id", "lexeme ID is duplicated")
		}
		if !validLanguage(lexeme.Language) {
			return invalid("INVALID_LANGUAGE_ID", path+".language", "language tag must be canonical lowercase ASCII")
		}
		if !validText(lexeme.Lemma, MaxTextBytes) {
			return invalid("INVALID_LEMMA", path+".lemma", "lemma is empty, untrimmed, invalid UTF-8, contains controls, or exceeds its bound")
		}
		if !validLabel(string(lexeme.POS), 64) {
			return invalid("INVALID_PART_OF_SPEECH", path+".pos", "part of speech must be a bounded lowercase label")
		}
		if lexeme.ParadigmID != "" && !validReferenceID(lexeme.ParadigmID) {
			return invalid("INVALID_PARADIGM_ID", path+".paradigm_id", "paradigm ID is malformed or exceeds its bound")
		}
		if len(lexeme.Properties) > MaxProperties {
			return resourceLimit("properties", len(lexeme.Properties), MaxProperties)
		}
		properties := make(map[string]struct{}, len(lexeme.Properties))
		for propertyIndex, property := range lexeme.Properties {
			propertyPath := fmt.Sprintf("%s.properties[%d]", path, propertyIndex)
			if err := meter.add(24, property.Key, property.Value); err != nil {
				return err
			}
			if !validLabel(property.Key, MaxFeatureBytes) || !validText(property.Value, MaxTextBytes) {
				return invalid("INVALID_LEXICAL_PROPERTY", propertyPath, "property key or value is malformed")
			}
			if _, exists := properties[property.Key]; exists {
				return invalid("DUPLICATE_LEXICAL_PROPERTY", propertyPath+".key", "property key is duplicated")
			}
			properties[property.Key] = struct{}{}
		}
		lexemes[key] = lexeme
	}
	senses := make(map[string]LexicalSense, len(document.Senses))
	for index, sense := range document.Senses {
		path := fmt.Sprintf("senses[%d]", index)
		if err := meter.add(72, sense.ID.String(), sense.LexemeID.String(), sense.Register); err != nil {
			return err
		}
		if !validTypedSenseID(sense.ID) {
			return invalid("INVALID_SENSE_ID", path+".id", "sense ID must use the sense namespace")
		}
		key := sense.ID.String()
		if _, exists := senses[key]; exists {
			return invalid("DUPLICATE_SENSE_ID", path+".id", "sense ID is duplicated")
		}
		if !validTypedLexemeID(sense.LexemeID) {
			return invalid("INVALID_LEXEME_ID", path+".lexeme_id", "sense lexeme ID must use the lexeme namespace")
		}
		if _, exists := lexemes[sense.LexemeID.String()]; !exists {
			return unknown("UNKNOWN_LEXEME", path+".lexeme_id", "sense lexeme does not resolve")
		}
		if sense.Register != "" && !validLabel(sense.Register, 64) {
			return invalid("INVALID_REGISTER", path+".register", "register must be a bounded lowercase label")
		}
		if len(sense.Usage) > MaxUsageNotes {
			return resourceLimit("usage_notes", len(sense.Usage), MaxUsageNotes)
		}
		for usageIndex, usage := range sense.Usage {
			if err := meter.add(8, usage); err != nil {
				return err
			}
			if !validText(usage, MaxTextBytes) {
				return invalid("INVALID_USAGE", fmt.Sprintf("%s.usage[%d]", path, usageIndex), "usage note is malformed or exceeds its bound")
			}
		}
		if len(sense.ValencyIDs) > MaxValencyRefs {
			return resourceLimit("valency_refs", len(sense.ValencyIDs), MaxValencyRefs)
		}
		valencies := make(map[ValencyID]struct{}, len(sense.ValencyIDs))
		for refIndex, valencyID := range sense.ValencyIDs {
			if err := meter.add(8, string(valencyID)); err != nil {
				return err
			}
			if !validReferenceID(string(valencyID)) {
				return invalid("INVALID_VALENCY_ID", fmt.Sprintf("%s.valency_ids[%d]", path, refIndex), "valency ID is malformed or exceeds its bound")
			}
			if _, exists := valencies[valencyID]; exists {
				return invalid("DUPLICATE_VALENCY_ID", path+".valency_ids", "valency reference is duplicated")
			}
			valencies[valencyID] = struct{}{}
		}
		senses[key] = sense
	}
	seenBindings := make(map[string]struct{}, len(document.Bindings))
	parsedBindings := make([]ids.ConceptID, len(document.Bindings))
	for index, binding := range document.Bindings {
		path := fmt.Sprintf("bindings[%d]", index)
		if err := meter.add(112, binding.SenseID.String(), binding.CoreRef.String(), string(binding.Mapping), string(binding.Status), binding.Provenance.ID, binding.Provenance.Path, binding.Provenance.URI, binding.Provenance.Version, binding.Provenance.License); err != nil {
			return err
		}
		if !validTypedSenseID(binding.SenseID) {
			return invalid("INVALID_SENSE_ID", path+".sense_id", "binding sense ID must use the sense namespace")
		}
		if _, exists := senses[binding.SenseID.String()]; !exists {
			return unknown("UNKNOWN_SENSE", path+".sense_id", "binding sense does not resolve")
		}
		if binding.CoreRef.String() == "" {
			return invalid("INVALID_CONCEPT_ID", path+".core_ref", "binding CORE concept ID is empty")
		}
		coreRef, err := ids.ParseConceptID(binding.CoreRef.String())
		if err != nil {
			return invalid("INVALID_CONCEPT_ID", path+".core_ref", "binding CORE concept ID is malformed")
		}
		parsedBindings[index] = coreRef
		if !binding.Mapping.Valid() {
			return invalid("INVALID_MAPPING_KIND", path+".mapping_kind", "mapping kind is empty or unsupported")
		}
		if !binding.Status.Valid() {
			return invalid("INVALID_BINDING_STATUS", path+".status", "binding status is empty or unsupported")
		}
		if err := binding.Provenance.Validate(); err != nil {
			return invalid("INVALID_PROVENANCE", path+".provenance", "binding provenance is malformed")
		}
		key := binding.SenseID.String() + "\x00" + coreRef.String()
		if _, exists := seenBindings[key]; exists {
			return invalid("DUPLICATE_SEMANTIC_BINDING", path, "sense-to-concept pair has multiple competing bindings")
		}
		seenBindings[key] = struct{}{}
	}
	if len(document.Bindings) > 0 && ontology == nil {
		return unknown("CORE_GRAPH_REQUIRED", "bindings", "CORE snapshot is required to validate semantic bindings")
	}
	for index, coreRef := range parsedBindings {
		if _, exists := ontology.Lookup(coreRef); !exists {
			return unknown("UNKNOWN_CONCEPT", fmt.Sprintf("bindings[%d].core_ref", index), "binding concept does not resolve in the supplied CORE snapshot")
		}
	}
	return nil
}

type documentMeter struct {
	used    int
	maximum int
}

func (meter *documentMeter) add(overhead int, values ...string) error {
	if overhead > meter.maximum-meter.used {
		return resourceLimit("document_bytes", meter.maximum+1, meter.maximum)
	}
	cost := overhead
	for _, value := range values {
		if len(value) > meter.maximum-meter.used-cost-2 {
			return resourceLimit("document_bytes", meter.maximum+1, meter.maximum)
		}
		cost += len(value) + 2 // UTF-8 payload and JSON quotes.
		for _, r := range value {
			switch r {
			case '"', '\\':
				cost++
			case '\u2028', '\u2029':
				cost += 3
			}
		}
	}
	if cost > meter.maximum-meter.used {
		return resourceLimit("document_bytes", meter.used+cost, meter.maximum)
	}
	meter.used += cost
	return nil
}

func validTypedLexemeID(id ids.LexemeID) bool {
	_, err := ids.ParseLexemeID(id.String())
	return err == nil
}

func validTypedSenseID(id ids.SenseID) bool {
	_, err := ids.ParseSenseID(id.String())
	return err == nil
}

func validLanguage(language LanguageID) bool {
	value := string(language)
	if value == "" || len(value) > 35 || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z') && !(index > 0 && r >= '0' && r <= '9') {
				return false
			}
		}
	}
	return true
}

func validLabel(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

func validReferenceID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	colon := 0
	for index, r := range value {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' && r != '.' && r != ':' {
			return false
		}
		if r == ':' {
			colon++
			if index == 0 || index == len(value)-1 {
				return false
			}
		}
	}
	return colon <= 1
}

func validText(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

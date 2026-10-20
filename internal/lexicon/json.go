package lexicon

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

func (document *Document) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("lexicon document must be a JSON object")
	}
	for _, field := range []string{"schema_version", "lexemes", "senses", "bindings"} {
		value, exists := fields[field]
		if !exists {
			return fmt.Errorf("lexicon document field %q is required", field)
		}
		if (field == "lexemes" || field == "senses" || field == "bindings") && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("lexicon document field %q cannot be null", field)
		}
	}
	type wireDocument Document
	var decoded wireDocument
	if err := decodeExactObject(data, map[string]bool{"schema_version": true, "lexemes": true, "senses": true, "bindings": true}, &decoded); err != nil {
		return err
	}
	*document = Document(decoded)
	return nil
}

func (lexeme *MorphLexeme) UnmarshalJSON(data []byte) error {
	var wire struct {
		ID         string       `json:"id"`
		Language   LanguageID   `json:"language"`
		Lemma      string       `json:"lemma"`
		POS        PartOfSpeech `json:"pos"`
		Properties []Property   `json:"properties,omitempty"`
		ParadigmID string       `json:"paradigm_id,omitempty"`
	}
	allowed := map[string]bool{"id": true, "language": true, "lemma": true, "pos": true, "properties": true, "paradigm_id": true}
	if err := rejectNullFields(data, "properties", "paradigm_id"); err != nil {
		return err
	}
	if err := decodeExactObject(data, allowed, &wire); err != nil {
		return err
	}
	id, err := ids.ParseLexemeID(wire.ID)
	if err != nil {
		return fmt.Errorf("lexeme id: %w", err)
	}
	*lexeme = MorphLexeme{ID: id, Language: wire.Language, Lemma: wire.Lemma, POS: wire.POS, Properties: wire.Properties, ParadigmID: wire.ParadigmID}
	return nil
}

func (lexeme MorphLexeme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID         string       `json:"id"`
		Language   LanguageID   `json:"language"`
		Lemma      string       `json:"lemma"`
		POS        PartOfSpeech `json:"pos"`
		Properties []Property   `json:"properties,omitempty"`
		ParadigmID string       `json:"paradigm_id,omitempty"`
	}{ID: lexeme.ID.String(), Language: lexeme.Language, Lemma: lexeme.Lemma, POS: lexeme.POS, Properties: lexeme.Properties, ParadigmID: lexeme.ParadigmID})
}

func (sense *LexicalSense) UnmarshalJSON(data []byte) error {
	var wire struct {
		ID         string      `json:"id"`
		LexemeID   string      `json:"lexeme_id"`
		Register   string      `json:"register,omitempty"`
		Usage      []string    `json:"usage,omitempty"`
		ValencyIDs []ValencyID `json:"valency_ids,omitempty"`
	}
	allowed := map[string]bool{"id": true, "lexeme_id": true, "register": true, "usage": true, "valency_ids": true}
	if err := rejectNullFields(data, "register", "usage", "valency_ids"); err != nil {
		return err
	}
	if err := decodeExactObject(data, allowed, &wire); err != nil {
		return err
	}
	id, err := ids.ParseSenseID(wire.ID)
	if err != nil {
		return fmt.Errorf("sense id: %w", err)
	}
	lexemeID, err := ids.ParseLexemeID(wire.LexemeID)
	if err != nil {
		return fmt.Errorf("sense lexeme id: %w", err)
	}
	*sense = LexicalSense{ID: id, LexemeID: lexemeID, Register: wire.Register, Usage: wire.Usage, ValencyIDs: wire.ValencyIDs}
	return nil
}

func (sense LexicalSense) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID         string      `json:"id"`
		LexemeID   string      `json:"lexeme_id"`
		Register   string      `json:"register,omitempty"`
		Usage      []string    `json:"usage,omitempty"`
		ValencyIDs []ValencyID `json:"valency_ids,omitempty"`
	}{ID: sense.ID.String(), LexemeID: sense.LexemeID.String(), Register: sense.Register, Usage: sense.Usage, ValencyIDs: sense.ValencyIDs})
}

func (binding *SemanticBinding) UnmarshalJSON(data []byte) error {
	var wire struct {
		SenseID    string          `json:"sense_id"`
		CoreRef    string          `json:"core_ref"`
		Mapping    MappingKind     `json:"mapping_kind"`
		Status     BindingStatus   `json:"status"`
		Provenance json.RawMessage `json:"provenance"`
	}
	allowed := map[string]bool{"sense_id": true, "core_ref": true, "mapping_kind": true, "status": true, "provenance": true}
	if err := decodeExactObject(data, allowed, &wire); err != nil {
		return err
	}
	senseID, err := ids.ParseSenseID(wire.SenseID)
	if err != nil {
		return fmt.Errorf("binding sense id: %w", err)
	}
	coreRef, err := ids.ParseConceptID(wire.CoreRef)
	if err != nil {
		return fmt.Errorf("binding core ref: %w", err)
	}
	var source provenance.SourceRef
	if err := decodeExactObject(wire.Provenance, map[string]bool{"id": true, "path": true, "uri": true, "version": true, "license": true, "author": true, "checksum_sha256": true}, &source); err != nil {
		return fmt.Errorf("binding provenance: %w", err)
	}
	*binding = SemanticBinding{SenseID: senseID, CoreRef: coreRef, Mapping: wire.Mapping, Status: wire.Status, Provenance: source}
	return nil
}

func (binding SemanticBinding) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenseID    string               `json:"sense_id"`
		CoreRef    string               `json:"core_ref"`
		Mapping    MappingKind          `json:"mapping_kind"`
		Status     BindingStatus        `json:"status"`
		Provenance provenance.SourceRef `json:"provenance"`
	}{SenseID: binding.SenseID.String(), CoreRef: binding.CoreRef.String(), Mapping: binding.Mapping, Status: binding.Status, Provenance: binding.Provenance})
}

func (property *Property) UnmarshalJSON(data []byte) error {
	type wireProperty Property
	var decoded wireProperty
	if err := decodeExactObject(data, map[string]bool{"key": true, "value": true}, &decoded); err != nil {
		return err
	}
	*property = Property(decoded)
	return nil
}

func decodeExactObject(data []byte, allowed map[string]bool, target any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unknown or mis-cased field %q", field)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func rejectNullFields(data []byte, names ...string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for _, name := range names {
		if value, exists := fields[name]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("field %q cannot be null", name)
		}
	}
	return nil
}

package importer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
)

const (
	ConceptFileSchemaVersion = 1
	MaxConceptFileBytes      = core.MaxConceptDocumentBytes
)

// ConceptFile is the importer envelope for language-neutral CORE definitions.
type ConceptFile struct {
	SchemaVersion int                       `json:"schema_version"`
	Relations     []core.RelationDefinition `json:"relations"`
	Concepts      []core.ConceptDefinition  `json:"concepts"`
}

// DecodeConceptFile reuses core's strict versioned decoder and validator.
func DecodeConceptFile(data []byte) (ConceptFile, error) {
	if len(data) > MaxConceptFileBytes {
		return ConceptFile{}, resourceLimit("concept_bytes", len(data), MaxConceptFileBytes)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		code := "INVALID_CONCEPT_FILE"
		if errors.Is(err, errDuplicateJSONKey) {
			code = "DUPLICATE_JSON_KEY"
		}
		return ConceptFile{}, invalid(code, "concepts", err.Error())
	}
	document, err := core.DecodeConceptDocument(data)
	if err != nil {
		return ConceptFile{}, invalid("INVALID_CONCEPT_FILE", "concepts", err.Error())
	}
	if _, err := core.CompileConceptGraph(document.Concepts, document.Relations); err != nil {
		return ConceptFile{}, invalid("INVALID_CONCEPT_FILE", "concepts", err.Error())
	}
	return ConceptFile{SchemaVersion: document.SchemaVersion, Relations: document.Relations, Concepts: document.Concepts}, nil
}

func (file ConceptFile) Validate() error {
	if file.SchemaVersion != ConceptFileSchemaVersion {
		return invalid("CONCEPT_SCHEMA_VERSION", "schema_version", fmt.Sprintf("unsupported concept file schema version %d", file.SchemaVersion))
	}
	_, err := core.CompileConceptGraph(file.Concepts, file.Relations)
	if err != nil {
		return invalid("INVALID_CONCEPT_FILE", "concepts", err.Error())
	}
	return nil
}

func (file ConceptFile) MarshalJSON() ([]byte, error) {
	if err := file.Validate(); err != nil {
		return nil, err
	}
	type wire ConceptFile
	return json.Marshal(wire(file))
}

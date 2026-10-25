// Package metamodel defines versioned, source-neutral metadata documents for
// formal symbols. It does not load, validate, or execute documents.
package metamodel

import (
	"encoding/json"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
)

const DocumentSchemaVersion = 1
const PositiveFactsProfile = "positive-facts-v1"

type SymbolKind string

const (
	SymbolPrimitive          SymbolKind = "primitive"
	SymbolAlias              SymbolKind = "alias"
	SymbolOpaqueUnformalized SymbolKind = "opaque_unformalized"
)

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
)

// Document is a declarative package manifest. Labels and descriptions are
// metadata; only IDs, signatures, constraints, and alias expressions denote
// formal structure.
type Document struct {
	SchemaVersion int                       `json:"schema_version"`
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	Profile       string                    `json:"profile"`
	Provenance    []provenance.SourceRef    `json:"provenance"`
	Sorts         []algebra.SortDeclaration `json:"sorts"`
	Symbols       []Symbol                  `json:"symbols"`
	Dependencies  []Dependency              `json:"dependencies"`
	TypicalEdges  []TypicalEdge             `json:"typical_edges"`
	Axioms        []json.RawMessage         `json:"axioms"`
}

type Symbol struct {
	ID           string                 `json:"id"`
	Kind         SymbolKind             `json:"kind"`
	Label        string                 `json:"label,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ReviewStatus ReviewStatus           `json:"review_status"`
	Provenance   []provenance.SourceRef `json:"provenance"`
	Ports        []algebra.Port         `json:"ports,omitempty"`
	Constraints  []Constraint           `json:"constraints,omitempty"`
	Alias        *Alias                 `json:"alias,omitempty"`
}

type Alias struct {
	Expression *algebra.Term `json:"expression"`
	Interface  []PortBinding `json:"interface"`
}

// PortBinding connects an externally named port to one leading lambda binder.
type PortBinding struct {
	Port   string `json:"port"`
	Binder string `json:"binder"`
}

type ConstraintKind string

const (
	ConstraintArity    ConstraintKind = "arity"
	ConstraintPortType ConstraintKind = "port_type"
)

// Constraint is a closed tagged record. Its interpretation is structural
// arity/type checking only; it cannot encode world rules or inequalities.
type Constraint struct {
	Kind  ConstraintKind `json:"kind"`
	Count *int           `json:"count,omitempty"`
	Role  string         `json:"role,omitempty"`
	Type  *algebra.Type  `json:"type,omitempty"`
}

type Dependency struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Fingerprint string `json:"fingerprint"`
}

// TypicalEdge is descriptive metadata between declared symbols. It is never
// promoted to an axiom or fact by this model.
type TypicalEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

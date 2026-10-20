// Package ids defines stable external symbols and snapshot-local numeric IDs.
// Display labels and language-specific strings are not identifiers.
package ids

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MaxNamespaceBytes = 32
	MaxNameBytes      = 256
	MaxNameSegments   = 16
	MaxSegmentBytes   = 64
	MaxSymbolBytes    = MaxNamespaceBytes + 1 + MaxNameBytes
)

type Namespace string

const (
	NamespaceConcept      Namespace = "concept"
	NamespaceLexeme       Namespace = "lexeme"
	NamespaceSense        Namespace = "sense"
	NamespaceConstruction Namespace = "construction"
	NamespaceProcedure    Namespace = "procedure"
)

var (
	ErrInvalidSymbol     = errors.New("invalid symbol")
	ErrInvalidNamespace  = errors.New("invalid symbol namespace")
	ErrInvalidName       = errors.New("invalid symbol name")
	ErrUnicodeSymbol     = errors.New("symbol identifiers must use ASCII characters")
	ErrPrototypeKey      = errors.New("prototype-sensitive names are not allowed")
	ErrNamespaceMismatch = errors.New("symbol namespace does not match typed ID")
)

// Symbol is a stable, human-readable identifier split into a graph namespace
// and a hierarchical name. Both fields must follow the ASCII identifier
// grammar validated by Canonical and NewSymbol.
type Symbol struct {
	Namespace Namespace
	Name      string
}

// NewSymbol validates and canonicalizes a symbol. Canonical identifiers use
// lowercase ASCII; natural-language labels belong in data, not in Symbol.Name.
func NewSymbol(namespace Namespace, name string) (Symbol, error) {
	canonicalNamespace, err := canonicalNamespace(string(namespace))
	if err != nil {
		return Symbol{}, err
	}
	canonicalName, err := canonicalName(name)
	if err != nil {
		return Symbol{}, err
	}
	return Symbol{Namespace: Namespace(canonicalNamespace), Name: canonicalName}, nil
}

// Parse reads the stable external form namespace:name. It accepts ASCII case
// variants and returns lowercase canonical fields. Non-ASCII identifiers are
// rejected so composed/decomposed or confusable Unicode cannot alias a symbol.
func Parse(value string) (Symbol, error) {
	if len(value) == 0 || len(value) > MaxSymbolBytes || !utf8.ValidString(value) {
		return Symbol{}, fmt.Errorf("%w: empty, invalid UTF-8, or exceeds %d bytes", ErrInvalidSymbol, MaxSymbolBytes)
	}
	if strings.Count(value, ":") != 1 {
		return Symbol{}, fmt.Errorf("%w: expected namespace:name", ErrInvalidSymbol)
	}
	parts := strings.SplitN(value, ":", 2)
	return NewSymbol(Namespace(parts[0]), parts[1])
}

// Canonical returns the validated stable external form namespace:name.
func (symbol Symbol) Canonical() (string, error) {
	canonical, err := NewSymbol(symbol.Namespace, symbol.Name)
	if err != nil {
		return "", err
	}
	return string(canonical.Namespace) + ":" + canonical.Name, nil
}

func canonicalNamespace(value string) (string, error) {
	if len(value) == 0 || len(value) > MaxNamespaceBytes {
		return "", fmt.Errorf("%w: empty or exceeds %d bytes", ErrInvalidNamespace, MaxNamespaceBytes)
	}
	canonical, err := canonicalSegment(value)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidNamespace, err)
	}
	if isPrototypeKey(canonical) {
		return "", fmt.Errorf("%w: namespace %q", ErrPrototypeKey, value)
	}
	return canonical, nil
}

func canonicalName(value string) (string, error) {
	if len(value) == 0 || len(value) > MaxNameBytes || !utf8.ValidString(value) {
		return "", fmt.Errorf("%w: empty, invalid UTF-8, or exceeds %d bytes", ErrInvalidName, MaxNameBytes)
	}
	segments := strings.Split(value, ".")
	if len(segments) > MaxNameSegments {
		return "", fmt.Errorf("%w: has more than %d segments", ErrInvalidName, MaxNameSegments)
	}
	canonicalSegments := make([]string, len(segments))
	for index, segment := range segments {
		if len(segment) > MaxSegmentBytes {
			return "", fmt.Errorf("%w: segment %d exceeds %d bytes", ErrInvalidName, index, MaxSegmentBytes)
		}
		canonical, err := canonicalSegment(segment)
		if err != nil {
			return "", fmt.Errorf("%w: segment %d: %w", ErrInvalidName, index, err)
		}
		if isPrototypeKey(canonical) {
			return "", fmt.Errorf("%w: name segment %q", ErrPrototypeKey, segment)
		}
		canonicalSegments[index] = canonical
	}
	return strings.Join(canonicalSegments, "."), nil
}

func canonicalSegment(value string) (string, error) {
	if value == "" {
		return "", errors.New("empty segment")
	}
	for _, character := range value {
		if character > 127 {
			return "", ErrUnicodeSymbol
		}
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '-') {
			return "", errors.New("segment contains a character outside [A-Za-z0-9_-]")
		}
	}
	return strings.ToLower(value), nil
}

func isPrototypeKey(value string) bool {
	switch value {
	case "__proto__", "prototype", "constructor":
		return true
	default:
		return false
	}
}

// ConceptID, LexemeID, SenseID, ConstructionID, and ProcedureID are distinct
// stable identity types. Their representation cannot be interchanged through
// ordinary assignment.
type ConceptID struct{ symbol Symbol }
type LexemeID struct{ symbol Symbol }
type SenseID struct{ symbol Symbol }
type ConstructionID struct{ symbol Symbol }
type ProcedureID struct{ symbol Symbol }

func NewConceptID(name string) (ConceptID, error) {
	symbol, err := NewSymbol(NamespaceConcept, name)
	return ConceptID{symbol: symbol}, err
}

func ParseConceptID(value string) (ConceptID, error) {
	symbol, err := parseTyped(value, NamespaceConcept)
	return ConceptID{symbol: symbol}, err
}

func (id ConceptID) Symbol() Symbol { return id.symbol }
func (id ConceptID) String() string { return canonicalString(id.symbol) }

func NewLexemeID(name string) (LexemeID, error) {
	symbol, err := NewSymbol(NamespaceLexeme, name)
	return LexemeID{symbol: symbol}, err
}

func ParseLexemeID(value string) (LexemeID, error) {
	symbol, err := parseTyped(value, NamespaceLexeme)
	return LexemeID{symbol: symbol}, err
}

func (id LexemeID) Symbol() Symbol { return id.symbol }
func (id LexemeID) String() string { return canonicalString(id.symbol) }

func NewSenseID(name string) (SenseID, error) {
	symbol, err := NewSymbol(NamespaceSense, name)
	return SenseID{symbol: symbol}, err
}

func ParseSenseID(value string) (SenseID, error) {
	symbol, err := parseTyped(value, NamespaceSense)
	return SenseID{symbol: symbol}, err
}

func (id SenseID) Symbol() Symbol { return id.symbol }
func (id SenseID) String() string { return canonicalString(id.symbol) }

func NewConstructionID(name string) (ConstructionID, error) {
	symbol, err := NewSymbol(NamespaceConstruction, name)
	return ConstructionID{symbol: symbol}, err
}

func ParseConstructionID(value string) (ConstructionID, error) {
	symbol, err := parseTyped(value, NamespaceConstruction)
	return ConstructionID{symbol: symbol}, err
}

func (id ConstructionID) Symbol() Symbol { return id.symbol }
func (id ConstructionID) String() string { return canonicalString(id.symbol) }

func NewProcedureID(name string) (ProcedureID, error) {
	symbol, err := NewSymbol(NamespaceProcedure, name)
	return ProcedureID{symbol: symbol}, err
}

func ParseProcedureID(value string) (ProcedureID, error) {
	symbol, err := parseTyped(value, NamespaceProcedure)
	return ProcedureID{symbol: symbol}, err
}

func (id ProcedureID) Symbol() Symbol { return id.symbol }
func (id ProcedureID) String() string { return canonicalString(id.symbol) }

func parseTyped(value string, expected Namespace) (Symbol, error) {
	symbol, err := Parse(value)
	if err != nil {
		return Symbol{}, err
	}
	if symbol.Namespace != expected {
		return Symbol{}, fmt.Errorf("%w: expected %q, got %q", ErrNamespaceMismatch, expected, symbol.Namespace)
	}
	return symbol, nil
}

func canonicalString(symbol Symbol) string {
	value, _ := symbol.Canonical()
	return value
}

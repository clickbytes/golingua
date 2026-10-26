// Package algebra defines a small, domain-neutral typed expression language.
// It checks representation and scope; it does not interpret symbols or prove
// that propositions are true.
package algebra

import (
	"fmt"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
)

// TypeKind is the closed set of built-in and compositional type forms.
type TypeKind string

const (
	TypeEntity   TypeKind = "Entity"
	TypeEvent    TypeKind = "Event"
	TypeInterval TypeKind = "Interval"
	TypeProp     TypeKind = "Prop"
	TypeNominal  TypeKind = "Nominal"
	TypeArrow    TypeKind = "Arrow"
	TypeQuestion TypeKind = "Question"
)

// Type is either a built-in sort, a declared nominal sort, or a function type.
// Fields irrelevant to Kind must be zero; Check enforces the tagged shape.
type Type struct {
	Kind   TypeKind `json:"kind"`
	SortID string   `json:"sort_id,omitempty"`
	From   *Type    `json:"from,omitempty"`
	To     *Type    `json:"to,omitempty"`
}

// SortDeclaration gives a nominal sort a stable ID. Label is presentation data.
type SortDeclaration struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

// Port is one ordered argument position of a relation signature.
type Port struct {
	Role string `json:"role"`
	Type Type   `json:"type"`
}

// Signature declares an uninterpreted proposition-valued symbol.
type Signature struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Ports []Port `json:"ports"`
}

// Declaration assigns a type to a free parameter or explicit reference.
type Declaration struct {
	ID   string `json:"id"`
	Type Type   `json:"type"`
}

// Limits bound structural checking. A zero field selects its conservative
// default; negative or over-cap values are invalid.
type Limits struct {
	MaxNodes     int `json:"max_nodes,omitempty"`
	MaxDepth     int `json:"max_depth,omitempty"`
	MaxTypeDepth int `json:"max_type_depth,omitempty"`
	MaxSymbols   int `json:"max_symbols,omitempty"`
	MaxPorts     int `json:"max_ports,omitempty"`
}

const (
	DefaultMaxNodes     = 4096
	DefaultMaxDepth     = 64
	DefaultMaxTypeDepth = 32
	DefaultMaxSymbols   = 512
	DefaultMaxPorts     = 16
)

// DefaultLimits returns the finite default checker budget.
func DefaultLimits() Limits {
	return Limits{
		MaxNodes: DefaultMaxNodes, MaxDepth: DefaultMaxDepth,
		MaxTypeDepth: DefaultMaxTypeDepth, MaxSymbols: DefaultMaxSymbols,
		MaxPorts: DefaultMaxPorts,
	}
}

// NormalizeLimits applies defaults and rejects negative or above-cap requests.
func NormalizeLimits(requested Limits) (Limits, error) {
	defaults := DefaultLimits()
	values := []struct {
		name  string
		value *int
		def   int
		cap   int
	}{
		{"max_nodes", &requested.MaxNodes, defaults.MaxNodes, defaults.MaxNodes},
		{"max_depth", &requested.MaxDepth, defaults.MaxDepth, defaults.MaxDepth},
		{"max_type_depth", &requested.MaxTypeDepth, defaults.MaxTypeDepth, defaults.MaxTypeDepth},
		{"max_symbols", &requested.MaxSymbols, defaults.MaxSymbols, defaults.MaxSymbols},
		{"max_ports", &requested.MaxPorts, defaults.MaxPorts, defaults.MaxPorts},
	}
	for _, item := range values {
		if *item.value == 0 {
			*item.value = item.def
			continue
		}
		if *item.value < 0 || *item.value > item.cap {
			return Limits{}, invalid("INVALID_LIMIT", "limits."+item.name,
				fmt.Sprintf("must be zero (default) or between 1 and %d", item.cap))
		}
	}
	return requested, nil
}

// TypeEqual reports exact structural type equality. It does not infer
// subtyping. It is cycle-safe for caller-constructed pointer graphs.
func TypeEqual(left, right Type) bool {
	type pair struct{ left, right *Type }
	stack := []pair{{&left, &right}}
	seen := make(map[pair]struct{})
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current.left == nil || current.right == nil {
			if current.left != current.right {
				return false
			}
			continue
		}
		key := current
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if current.left.Kind != current.right.Kind || current.left.SortID != current.right.SortID {
			return false
		}
		stack = append(stack, pair{current.left.From, current.right.From}, pair{current.left.To, current.right.To})
	}
	return true
}

func validText(value string, max int) bool {
	return value != "" && len(value) <= max && utf8.ValidString(value)
}

func validOptionalText(value string, max int) bool {
	return len(value) <= max && utf8.ValidString(value)
}

// Error is a typed checker failure with a stable status, code, and input path.
type Error struct {
	Status  status.OutcomeKind `json:"status"`
	Code    string             `json:"code"`
	Path    string             `json:"path"`
	Message string             `json:"message"`
}

func (failure *Error) Error() string {
	if failure == nil {
		return "algebra check failed"
	}
	if failure.Path == "" {
		return failure.Code + ": " + failure.Message
	}
	return failure.Code + " at " + failure.Path + ": " + failure.Message
}

func invalid(code, path, message string) error {
	return &Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func resourceLimit(path, message string) error {
	return &Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

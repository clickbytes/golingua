package algebra

import "github.com/clickbytes/golingua/internal/status"

// TermKind is the closed tagged union of expression constructors.
type TermKind string

const (
	TermVar        TermKind = "Var"
	TermParam      TermKind = "Param"
	TermRef        TermKind = "Ref"
	TermHole       TermKind = "Hole"
	TermAtom       TermKind = "Atom"
	TermAnd        TermKind = "And"
	TermExists     TermKind = "Exists"
	TermLambda     TermKind = "Lambda"
	TermApply      TermKind = "Apply"
	TermNot        TermKind = "Not"
	TermQuantifier TermKind = "Quantifier"
	TermAsk        TermKind = "Ask"
)

// Argument assigns one expression to a signature role. The sequence of Args
// must match the signature's declared port order exactly.
type Argument struct {
	Role  string `json:"role"`
	Value *Term  `json:"value"`
}

// Term is the typed IR node. Check enforces that only fields belonging to Kind
// are populated. Label is optional presentation data and never affects meaning.
type Term struct {
	Kind       TermKind      `json:"kind"`
	ID         string        `json:"id,omitempty"`
	Label      string        `json:"label,omitempty"`
	Type       *Type         `json:"type,omitempty"`
	Symbol     string        `json:"symbol,omitempty"`
	Args       []Argument    `json:"args,omitempty"`
	Operands   []*Term       `json:"operands,omitempty"`
	Body       *Term         `json:"body,omitempty"`
	Function   *Term         `json:"function,omitempty"`
	Argument   *Term         `json:"argument,omitempty"`
	Quantifier string        `json:"quantifier,omitempty"`
	Count      *int          `json:"count,omitempty"`
	Restrictor *Term         `json:"restrictor,omitempty"`
	Answers    []Declaration `json:"answers,omitempty"`
}

// Requirement describes an unresolved typed hole. It does not assert a fact
// or introduce an existential witness.
type Requirement struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Type Type   `json:"type"`
}

// CheckedExpression is the structural checker result. Status reports only
// whether the expression is structurally complete; it makes no truth claim.
type CheckedExpression struct {
	Status         status.OutcomeKind `json:"status"`
	Type           Type               `json:"type"`
	FreeParameters []Declaration      `json:"free_parameters,omitempty"`
	Holes          []Requirement      `json:"holes,omitempty"`
	Requirements   []Requirement      `json:"requirements,omitempty"`
	NodeCount      int                `json:"node_count"`
}

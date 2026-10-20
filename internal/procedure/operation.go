// Package procedure defines a closed declarative string-operation DSL.
// Plans contain data only and are never evaluated as code.
package procedure

import "github.com/clickbytes/golingua/internal/features"

type StemID string

type OperationKind string

const (
	OperationAppend        OperationKind = "append"
	OperationPrepend       OperationKind = "prepend"
	OperationSelectStem    OperationKind = "select_stem"
	OperationStripSuffix   OperationKind = "strip_suffix"
	OperationReplaceSuffix OperationKind = "replace_suffix"
	OperationAlternation   OperationKind = "alternation"
)

func (kind OperationKind) Valid() bool {
	switch kind {
	case OperationAppend, OperationPrepend, OperationSelectStem, OperationStripSuffix, OperationReplaceSuffix, OperationAlternation:
		return true
	default:
		return false
	}
}

// FeatureCondition makes one operation conditional on known profile data.
type FeatureCondition struct {
	Feature features.FeatureID      `json:"feature"`
	Value   features.FeatureValueID `json:"value"`
}

// Operation is a tagged union. Only fields required by Kind may be set.
// Replacement and To are pointers so an explicitly empty replacement is
// distinct from an omitted replacement (a zero exponent is representable).
type Operation struct {
	Kind        OperationKind      `json:"kind"`
	StemID      StemID             `json:"stem_id,omitempty"`
	Text        string             `json:"text,omitempty"`
	Suffix      string             `json:"suffix,omitempty"`
	Replacement *string            `json:"replacement,omitempty"`
	From        string             `json:"from,omitempty"`
	To          *string            `json:"to,omitempty"`
	When        []FeatureCondition `json:"when,omitempty"`
}

// StepTrace reports a bounded applied step without storing arbitrary logs.
type StepTrace struct {
	Index      int           `json:"index"`
	Operation  OperationKind `json:"operation"`
	StemID     StemID        `json:"stem_id,omitempty"`
	Applied    bool          `json:"applied"`
	OutputSize int           `json:"output_size"`
}

type ExecutionResult struct {
	Form  string      `json:"form"`
	Trace []StepTrace `json:"trace"`
}

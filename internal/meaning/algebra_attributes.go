package meaning

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
)

// ApplyReferenceAttributes copies explicit attributes from one validated
// Meaning node onto a caller-selected typed algebra reference. It does not
// infer identity, source evidence, cardinality, or facts from concept labels.
func ApplyReferenceAttributes(reference algebra.Declaration, node *Node) (algebra.Declaration, error) {
	if node == nil {
		return algebra.Declaration{}, fmt.Errorf("meaning node is required")
	}
	out := reference
	switch node.Kind {
	case KindEvent, KindState:
		if reference.Type.Kind != algebra.TypeEvent || node.Multiplicity != "" {
			return algebra.Declaration{}, fmt.Errorf("event/state attributes require an Event reference")
		}
		if node.TemporalFrame != "" && !node.TemporalFrame.Valid() {
			return algebra.Declaration{}, fmt.Errorf("unknown temporal frame")
		}
		if out.TemporalFrame != "" && node.TemporalFrame != "" && out.TemporalFrame != string(node.TemporalFrame) {
			return algebra.Declaration{}, fmt.Errorf("conflicting explicit temporal frame")
		}
		if node.TemporalFrame != "" {
			out.TemporalFrame = string(node.TemporalFrame)
		}
	case KindReferent:
		if (reference.Type.Kind != algebra.TypeEntity && reference.Type.Kind != algebra.TypeNominal) || node.TemporalFrame != "" {
			return algebra.Declaration{}, fmt.Errorf("referent multiplicity requires an Entity or Nominal reference")
		}
		if node.Multiplicity != "" && !node.Multiplicity.Valid() {
			return algebra.Declaration{}, fmt.Errorf("unknown referent multiplicity")
		}
		if out.Multiplicity != "" && node.Multiplicity != "" && out.Multiplicity != string(node.Multiplicity) {
			return algebra.Declaration{}, fmt.Errorf("conflicting explicit multiplicity")
		}
		if node.Multiplicity != "" {
			out.Multiplicity = string(node.Multiplicity)
		}
	default:
		return algebra.Declaration{}, fmt.Errorf("node kind %q does not provide reference attributes", node.Kind)
	}
	return out, nil
}

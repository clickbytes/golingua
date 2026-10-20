package lexicon

import (
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

type MappingKind string

const (
	MappingEquivalent MappingKind = "equivalent"
	MappingBroad      MappingKind = "broad"
	MappingNarrow     MappingKind = "narrow"
	MappingRelated    MappingKind = "related"
	MappingCandidate  MappingKind = "candidate"
)

func (kind MappingKind) Valid() bool {
	switch kind {
	case MappingEquivalent, MappingBroad, MappingNarrow, MappingRelated, MappingCandidate:
		return true
	default:
		return false
	}
}

type BindingStatus string

const (
	BindingPending  BindingStatus = "pending"
	BindingApproved BindingStatus = "approved"
)

func (status BindingStatus) Valid() bool {
	return status == BindingPending || status == BindingApproved
}

// SemanticBinding connects one sense to a CORE concept. Only approved
// equivalent bindings can authorize a semantic transfer.
type SemanticBinding struct {
	SenseID    ids.SenseID          `json:"sense_id"`
	CoreRef    ids.ConceptID        `json:"core_ref"`
	Mapping    MappingKind          `json:"mapping_kind"`
	Status     BindingStatus        `json:"status"`
	Provenance provenance.SourceRef `json:"provenance"`
}

// TransferAuthorized reports whether this binding may license semantic
// transfer. It does not select among multiple bindings or validate CORE.
func (binding SemanticBinding) TransferAuthorized() bool {
	if binding.Mapping != MappingEquivalent || binding.Status != BindingApproved {
		return false
	}
	if _, err := ids.ParseSenseID(binding.SenseID.String()); err != nil {
		return false
	}
	if _, err := ids.ParseConceptID(binding.CoreRef.String()); err != nil {
		return false
	}
	return binding.Provenance.Validate() == nil
}

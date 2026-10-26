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
	Target     *SemanticTarget      `json:"target,omitempty"`
	Mapping    MappingKind          `json:"mapping_kind"`
	Status     BindingStatus        `json:"status"`
	Provenance provenance.SourceRef `json:"provenance"`
}

// TransferAuthorized reports whether this binding may license semantic
// transfer. It does not select among multiple bindings or validate CORE.
func (binding SemanticBinding) TransferAuthorized() bool {
	if binding.Target != nil {
		return false
	}
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

// TargetExecutionAuthorized is separate from legacy Sense-to-CORE transfer.
// Full document and selected-registry validation remains the caller's job.
func (binding SemanticBinding) TargetExecutionAuthorized() bool {
	if binding.Target == nil || binding.CoreRef.String() != "" || binding.Mapping != MappingEquivalent || binding.Status != BindingApproved || binding.Provenance.Validate() != nil {
		return false
	}
	frames := make(map[string]struct{}, len(binding.Target.RolePorts))
	sense := LexicalSense{ValencyIDs: make([]ValencyID, 0, len(binding.Target.RolePorts))}
	for _, mapping := range binding.Target.RolePorts {
		if _, exists := frames[mapping.FrameID]; !exists {
			frames[mapping.FrameID] = struct{}{}
			sense.ValencyIDs = append(sense.ValencyIDs, ValencyID(mapping.FrameID))
		}
	}
	return validateSemanticTarget(binding.Target, sense, "target") == nil && binding.Target.ExecutionAuthorized(binding.Mapping, binding.Status, true)
}

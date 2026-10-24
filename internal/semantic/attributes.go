package semantic

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
)

// NodeAttributesEvidence supplies explicit, language-neutral scalar
// attributes for one bounded phrase candidate.
type NodeAttributesEvidence struct {
	CandidateID   string
	TemporalFrame meaning.TemporalFrame
	Multiplicity  meaning.Multiplicity
}

func validateNodeAttributes(evidence []NodeAttributesEvidence, candidates map[string]parser.PhraseCandidate) (map[string]NodeAttributesEvidence, error) {
	validated := make(map[string]NodeAttributesEvidence, len(evidence))
	for index, attribute := range evidence {
		path := fmt.Sprintf("node_attributes[%d]", index)
		candidate, exists := candidates[attribute.CandidateID]
		if attribute.CandidateID == "" || !exists {
			return nil, semanticError(status.OutcomeInvalid, "INVALID_NODE_ATTRIBUTES_EVIDENCE", path+".candidate_id", "node attributes require an existing phrase candidate")
		}
		if attribute.TemporalFrame == "" && attribute.Multiplicity == "" {
			return nil, semanticError(status.OutcomeInvalid, "EMPTY_NODE_ATTRIBUTES_EVIDENCE", path, "node attributes must supply at least one explicit attribute")
		}
		if attribute.TemporalFrame != "" {
			if !attribute.TemporalFrame.Valid() {
				return nil, semanticError(status.OutcomeInvalid, "INVALID_TEMPORAL_FRAME", path+".temporal_frame", "temporal frame is unsupported")
			}
			if !candidate.Category.IsPredicate() {
				return nil, semanticError(status.OutcomeInvalid, "TEMPORAL_FRAME_CANDIDATE_KIND", path+".candidate_id", "temporal frame evidence requires a predicate phrase candidate")
			}
		}
		if attribute.Multiplicity != "" {
			if !attribute.Multiplicity.Valid() {
				return nil, semanticError(status.OutcomeInvalid, "INVALID_MULTIPLICITY", path+".multiplicity", "referent multiplicity is unsupported")
			}
			if candidate.Category != construction.CategoryNP {
				return nil, semanticError(status.OutcomeInvalid, "MULTIPLICITY_CANDIDATE_KIND", path+".candidate_id", "multiplicity evidence requires an NP phrase candidate")
			}
		}
		if _, duplicate := validated[attribute.CandidateID]; duplicate {
			return nil, semanticError(status.OutcomeInvalid, "DUPLICATE_NODE_ATTRIBUTES_EVIDENCE", path+".candidate_id", "candidate may have only one node attributes evidence record")
		}
		validated[attribute.CandidateID] = attribute
	}
	return validated, nil
}

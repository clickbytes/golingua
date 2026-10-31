package semantic

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

// HasMeaningEvidenceAddresses reports whether a trace uses graph-local IDs.
// Such a trace cannot be merged into a different graph without an explicit
// mapping, even when the graphs have been proven semantically equivalent.
func HasMeaningEvidenceAddresses(trace EvidenceTrace) bool {
	for _, record := range trace.Records {
		if record.MeaningNodeID != "" || record.PredicateNodeID != "" {
			return true
		}
	}
	return false
}

// ValidateEvidenceMeaning verifies the graph-local part of an evidence trace.
// Legacy records without addresses remain readable, but this check never infers
// their target from a concept. Trace syntax/provenance validation stays separate.
func ValidateEvidenceMeaning(trace EvidenceTrace, graph meaning.Meaning) error {
	if len(trace.Records) > DefaultEvidenceBudget().MaxRecords {
		return evidenceError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "evidence.records", "evidence address inventory exceeds the record budget")
	}
	if !HasMeaningEvidenceAddresses(trace) {
		return nil
	}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		return err
	}
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	for i, record := range trace.Records {
		if record.MeaningNodeID == "" && record.PredicateNodeID == "" {
			continue
		}
		path := fmt.Sprintf("evidence.records[%d]", i)
		node, exists := nodes[record.MeaningNodeID]
		if !exists || record.CoreConceptID.String() == "" || string(node.Concept) != record.CoreConceptID.String() {
			return evidenceError(status.OutcomeInvalid, "EVIDENCE_NODE_MISMATCH", path+".meaning_node_id", "evidence must address an existing node with its exact CORE concept")
		}
		if record.Origin == EvidenceOriginGenerated {
			return evidenceError(status.OutcomeInvalid, "INVALID_GENERATED_EVIDENCE", path, "generated binding evidence cannot claim a source feature address")
		}
		if record.PredicateNodeID == "" {
			if node.Kind == meaning.KindReferent && record.Role != "" {
				return evidenceError(status.OutcomeInvalid, "EVIDENCE_ARGUMENT_OWNER_MISSING", path+".predicate_node_id", "argument evidence must identify its owning predicate")
			}
			continue
		}
		owner, found := nodes[record.PredicateNodeID]
		if !found || (owner.Kind != meaning.KindEvent && owner.Kind != meaning.KindState) || record.Role == "" || node.Kind != meaning.KindReferent {
			return evidenceError(status.OutcomeInvalid, "EVIDENCE_ARGUMENT_MISMATCH", path, "argument evidence requires a predicate, role and referent")
		}
		matches := 0
		for _, participant := range owner.Participants {
			if participant.Role == record.Role && participant.Target == record.MeaningNodeID {
				matches++
			}
		}
		if matches != 1 {
			return evidenceError(status.OutcomeInvalid, "EVIDENCE_ARGUMENT_MISMATCH", path, "evidence address does not match the exact predicate-role-referent link")
		}
	}
	return nil
}

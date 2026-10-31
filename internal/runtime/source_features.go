package runtime

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type sourceFeatureAddress struct {
	node, owner meaning.NodeID
}

// SourceFeatures is the compatibility lookup for an unambiguous concept/role.
// A repeated concept is not an occurrence identity. New consumers should pass
// the owning predicate to SourceFeaturesAt instead of asking this method to guess.
func SourceFeatures(source semantic.MeaningCandidate, role meaning.Role, concept ids.ConceptID) (map[features.FeatureID]features.FeatureValueID, bool) {
	addresses := sourceFeatureAddresses(source.Meaning, role, concept)
	if len(addresses) != 1 {
		return nil, false
	}
	return sourceFeaturesForAddress(source, addresses[0], role, concept)
}

// SourceFeaturesAt reads one exact predicate, or its argument in role. The
// supplied concept must agree with the addressed node. Old unaddressed records
// are usable only if this concept/role identifies one occurrence in the graph.
// A disagreement is not resolved by record order or by copying another node.
func SourceFeaturesAt(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) (map[features.FeatureID]features.FeatureValueID, bool) {
	address, found := sourceFeatureAddressAt(source.Meaning, predicateID, role, concept)
	if !found {
		return nil, false
	}
	return sourceFeaturesForAddress(source, address, role, concept)
}

// SourceFeaturesPresentAt distinguishes conflicting evidence from no evidence.
// It follows exactly the same occurrence binding as SourceFeaturesAt.
func SourceFeaturesPresentAt(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) bool {
	address, found := sourceFeatureAddressAt(source.Meaning, predicateID, role, concept)
	if !found {
		return false
	}
	var legacy []sourceFeatureAddress
	legacyLoaded := false
	for _, record := range source.Evidence.Records {
		if record.MeaningNodeID == "" && record.PredicateNodeID == "" && !legacyLoaded && record.Role == role && record.CoreConceptID == concept {
			legacy = sourceFeatureAddresses(source.Meaning, role, concept)
			legacyLoaded = true
		}
		if sourceRecordMatches(record, address, role, concept, legacy) {
			return true
		}
	}
	return false
}

// SourceEvidenceAppliesAt applies the same occurrence check to a lexical trace
// record used for non-morphological source information such as lexical manner.
func SourceEvidenceAppliesAt(source semantic.MeaningCandidate, record semantic.EvidenceRecord, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) bool {
	address, found := sourceFeatureAddressAt(source.Meaning, predicateID, role, concept)
	if !found {
		return false
	}
	var legacy []sourceFeatureAddress
	if record.MeaningNodeID == "" && record.PredicateNodeID == "" {
		legacy = sourceFeatureAddresses(source.Meaning, role, concept)
	}
	return sourceRecordMatches(record, address, role, concept, legacy)
}

func sourceFeaturesForAddress(source semantic.MeaningCandidate, address sourceFeatureAddress, role meaning.Role, concept ids.ConceptID) (map[features.FeatureID]features.FeatureValueID, bool) {
	values := make(map[features.FeatureID]features.FeatureValueID)
	found := false
	var legacy []sourceFeatureAddress
	legacyLoaded := false
	for _, record := range source.Evidence.Records {
		if record.MeaningNodeID == "" && record.PredicateNodeID == "" && !legacyLoaded && record.Role == role && record.CoreConceptID == concept {
			legacy = sourceFeatureAddresses(source.Meaning, role, concept)
			legacyLoaded = true
		}
		if !sourceRecordMatches(record, address, role, concept, legacy) {
			continue
		}
		found = true
		for _, feature := range record.Features {
			if feature.State != features.StateKnown {
				continue
			}
			if previous, exists := values[feature.Feature]; exists && previous != feature.Value {
				return nil, false
			}
			values[feature.Feature] = feature.Value
		}
	}
	return values, found
}

func sourceRecordMatches(record semantic.EvidenceRecord, address sourceFeatureAddress, role meaning.Role, concept ids.ConceptID, legacy []sourceFeatureAddress) bool {
	if record.Origin != "" || record.Outcome != status.OutcomeOK || record.CoreConceptID != concept || record.Role != role {
		return false
	}
	if record.MeaningNodeID != "" || record.PredicateNodeID != "" {
		return record.MeaningNodeID == address.node && record.PredicateNodeID == address.owner
	}
	return len(legacy) == 1 && legacy[0] == address
}

func sourceFeatureAddressAt(graph *meaning.Meaning, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) (sourceFeatureAddress, bool) {
	if graph == nil || predicateID == "" || concept.String() == "" || len(graph.Nodes) > meaning.MaxMeaningNodes {
		return sourceFeatureAddress{}, false
	}
	var predicate meaning.Node
	count := 0
	for _, node := range graph.Nodes {
		if node.ID == predicateID {
			predicate = node
			count++
		}
	}
	if count != 1 {
		return sourceFeatureAddress{}, false
	}
	if role == "" {
		return sourceFeatureAddress{node: predicate.ID}, string(predicate.Concept) == concept.String()
	}
	if (predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState) || len(predicate.Participants) > meaning.MaxParticipants {
		return sourceFeatureAddress{}, false
	}
	var target meaning.NodeID
	count = 0
	for _, participant := range predicate.Participants {
		if participant.Role == role {
			target = participant.Target
			count++
		}
	}
	if count != 1 {
		return sourceFeatureAddress{}, false
	}
	count = 0
	for _, node := range graph.Nodes {
		if node.ID == target && node.Kind == meaning.KindReferent && string(node.Concept) == concept.String() {
			count++
		}
	}
	return sourceFeatureAddress{node: target, owner: predicate.ID}, count == 1
}

func sourceFeatureAddresses(graph *meaning.Meaning, role meaning.Role, concept ids.ConceptID) []sourceFeatureAddress {
	if graph == nil || concept.String() == "" || len(graph.Nodes) > meaning.MaxMeaningNodes {
		return nil
	}
	var addresses []sourceFeatureAddress
	referents := make(map[meaning.NodeID]bool)
	for _, node := range graph.Nodes {
		if string(node.Concept) != concept.String() {
			continue
		}
		if role == "" {
			addresses = append(addresses, sourceFeatureAddress{node: node.ID})
		}
		if node.Kind == meaning.KindReferent {
			referents[node.ID] = true
		}
	}
	if role == "" {
		return addresses
	}
	for _, node := range graph.Nodes {
		if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState {
			continue
		}
		if len(node.Participants) > meaning.MaxParticipants {
			return nil
		}
		var target meaning.NodeID
		count := 0
		for _, participant := range node.Participants {
			if participant.Role == role {
				target = participant.Target
				count++
			}
		}
		if count == 1 && referents[target] {
			addresses = append(addresses, sourceFeatureAddress{node: target, owner: node.ID})
		}
	}
	return addresses
}

// ValidateSourceFeatureEvidence refuses stale addresses, legacy evidence that
// could belong to more than one occurrence, and conflicting known values before
// target defaults can mistake any of them for missing source information.
func ValidateSourceFeatureEvidence(source semantic.MeaningCandidate) error {
	if source.Meaning == nil {
		return nil
	}
	if err := semantic.ValidateEvidenceMeaning(source.Evidence, *source.Meaning); err != nil {
		return err
	}
	if len(source.Evidence.Records) > semantic.DefaultEvidenceBudget().MaxRecords {
		return runtimeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "source.evidence", "source evidence exceeds the record budget")
	}
	type key struct {
		address sourceFeatureAddress
		role    meaning.Role
		feature features.FeatureID
	}
	known := make(map[key]features.FeatureValueID)
	type legacyKey struct {
		role    meaning.Role
		concept ids.ConceptID
	}
	legacyCache := make(map[legacyKey][]sourceFeatureAddress)
	for _, record := range source.Evidence.Records {
		if len(record.Features) > features.MaxBundleValues {
			return runtimeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "source.evidence.features", "source feature bundle exceeds the existing feature limit")
		}
		if record.Origin != "" || record.Outcome != status.OutcomeOK || len(record.Features) == 0 {
			continue
		}
		address := sourceFeatureAddress{node: record.MeaningNodeID, owner: record.PredicateNodeID}
		if address.node == "" {
			lookup := legacyKey{record.Role, record.CoreConceptID}
			addresses, cached := legacyCache[lookup]
			if !cached {
				addresses = sourceFeatureAddresses(source.Meaning, record.Role, record.CoreConceptID)
				legacyCache[lookup] = addresses
			}
			if len(addresses) > 1 {
				return runtimeError(status.OutcomeUnsupported, "SOURCE_FEATURE_OCCURRENCE_AMBIGUOUS", "source.evidence", "unaddressed source features match multiple predicate or argument occurrences")
			}
			if len(addresses) == 0 {
				continue
			}
			address = addresses[0]
		}
		for _, feature := range record.Features {
			if feature.State != features.StateKnown {
				continue
			}
			k := key{address, record.Role, feature.Feature}
			if previous, exists := known[k]; exists && previous != feature.Value {
				return runtimeError(status.OutcomeUnsupported, "SOURCE_FEATURE_CONFLICT", "source.evidence", "known source features conflict at the same predicate or argument occurrence")
			}
			known[k] = feature.Value
		}
	}
	return nil
}

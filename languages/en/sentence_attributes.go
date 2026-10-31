package en

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func englishMeaningNode(source semantic.MeaningCandidate, nodeID meaning.NodeID) (meaning.Node, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func englishReferentForRole(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role) (meaning.Node, bool) {
	predicate, exists := englishMeaningNode(source, predicateID)
	if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return meaning.Node{}, false
	}
	var referentID meaning.NodeID
	for _, participant := range predicate.Participants {
		if participant.Role != role {
			continue
		}
		if referentID != "" {
			return meaning.Node{}, false
		}
		referentID = participant.Target
	}
	referent, exists := englishMeaningNode(source, referentID)
	return referent, exists && referent.Kind == meaning.KindReferent
}

func englishMultiplicityNumber(multiplicity meaning.Multiplicity) string {
	switch multiplicity {
	case meaning.MultiplicityOne:
		return "singular"
	case meaning.MultiplicityMany:
		return "plural"
	default:
		return ""
	}
}

func englishNumberClass(properties []lexicon.Property, paradigmID string) (mass, pluralOnly bool) {
	for _, property := range properties {
		switch property.Key {
		case "number_behavior":
			mass = mass || property.Value == "mass"
			pluralOnly = pluralOnly || property.Value == "plural_only"
		case "grammatical_number":
			pluralOnly = pluralOnly || property.Value == "plural_only"
		}
	}
	pluralOnly = pluralOnly || paradigmID == "en_noun_plural_only_attested"
	return mass, pluralOnly
}

// englishTargetNumber returns the receiving grammar's number, whether the
// choice came from Meaning, and whether that Meaning is realizable by the
// selected local lexical class.
func englishTargetNumber(source semantic.MeaningCandidate, referentID meaning.NodeID, properties []lexicon.Property, paradigmID string) (string, bool, bool) {
	referent, exists := englishMeaningNode(source, referentID)
	if !exists || referent.Kind != meaning.KindReferent {
		return "", false, false
	}
	mass, pluralOnly := englishNumberClass(properties, paradigmID)
	if mass || pluralOnly {
		if referent.Multiplicity != "" {
			return "", false, false
		}
		if pluralOnly {
			return "plural", false, true
		}
		return "singular", false, true
	}
	if number := englishMultiplicityNumber(referent.Multiplicity); number != "" {
		return number, true, true
	}
	return "singular", false, true
}

func englishSourceFeaturesPresent(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) bool {
	return runtime.SourceFeaturesPresentAt(source, predicateID, role, concept)
}

func englishFrameEvidenceConsistent(source semantic.MeaningCandidate, predicateID meaning.NodeID, concept ids.ConceptID) bool {
	predicate, exists := englishMeaningNode(source, predicateID)
	if !exists || predicate.TemporalFrame == "" {
		return true
	}
	values, found := runtime.SourceFeaturesAt(source, predicateID, "", concept)
	if !found {
		return !englishSourceFeaturesPresent(source, predicateID, "", concept)
	}
	if tense := string(values["tense"]); tense != "" && tense != string(predicate.TemporalFrame) {
		return false
	}
	if values["verb_form"] == "past" && predicate.TemporalFrame != meaning.TemporalFramePast {
		return false
	}
	if values["verb_form"] == "third_person_singular" && predicate.TemporalFrame != meaning.TemporalFramePresent {
		return false
	}
	return true
}

func englishNumberFeaturesAgree(values map[features.FeatureID]features.FeatureValueID, expected string) bool {
	observed := ""
	for _, feature := range []features.FeatureID{"number", "noun_number", "verb_number"} {
		value := string(values[feature])
		if value == "" {
			continue
		}
		if value != "singular" && value != "plural" || observed != "" && observed != value {
			return false
		}
		observed = value
	}
	return observed == "" || observed == expected
}

func englishPredicateNumberApplies(predicate realizer.TargetCandidate, role meaning.Role) bool {
	core := predicate.TransferBinding.CoreRef.String()
	if core == "concept:class_membership" && role == "member" || core == "concept:black" && role == "bearer" {
		return true
	}
	for _, argument := range predicate.Frame.Arguments {
		if argument.Role == role && argument.GrammaticalFunction == "subject" {
			return true
		}
	}
	return false
}

func englishPredicateSubjectRoleForTarget(predicate realizer.TargetCandidate) meaning.Role {
	core := predicate.TransferBinding.CoreRef.String()
	if core == "concept:class_membership" {
		return "member"
	}
	if core == "concept:black" {
		return "bearer"
	}
	var role meaning.Role
	for _, argument := range predicate.Frame.Arguments {
		if argument.GrammaticalFunction != "subject" {
			continue
		}
		if role != "" {
			return ""
		}
		role = argument.Role
	}
	return role
}

func englishMultiplicityEvidenceConsistent(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, referent meaning.Node, predicateConcept ids.ConceptID, checkPredicate bool) bool {
	expected := englishMultiplicityNumber(referent.Multiplicity)
	if expected == "" {
		return true
	}
	predicate, exists := englishMeaningNode(source, predicateID)
	if !exists {
		return false
	}
	participantFound := false
	for _, participant := range predicate.Participants {
		if participant.Role == role && participant.Target == referent.ID {
			if participantFound {
				return false
			}
			participantFound = true
		}
	}
	if !participantFound {
		return false
	}
	concept, err := ids.ParseConceptID(string(referent.Concept))
	if err == nil {
		values, found := runtime.SourceFeaturesAt(source, predicateID, role, concept)
		if !found {
			if englishSourceFeaturesPresent(source, predicateID, role, concept) {
				return false
			}
		} else if !englishNumberFeaturesAgree(values, expected) {
			return false
		}
	}
	if checkPredicate {
		values, found := runtime.SourceFeaturesAt(source, predicateID, "", predicateConcept)
		if !found {
			if englishSourceFeaturesPresent(source, predicateID, "", predicateConcept) {
				return false
			}
		} else if !englishNumberFeaturesAgree(values, expected) {
			return false
		}
	}
	return true
}

func englishPredicateEvidenceConsistent(source semantic.MeaningCandidate, predicate realizer.TargetCandidate) bool {
	if !englishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
		return false
	}
	meaningPredicate, exists := englishMeaningNode(source, predicate.NodeID)
	if !exists {
		return false
	}
	for _, participant := range meaningPredicate.Participants {
		referent, exists := englishMeaningNode(source, participant.Target)
		if !exists || referent.Kind != meaning.KindReferent || referent.Multiplicity == "" {
			continue
		}
		if !englishMultiplicityEvidenceConsistent(source, predicate.NodeID, participant.Role, referent, predicate.TransferBinding.CoreRef, englishPredicateNumberApplies(predicate, participant.Role)) {
			return false
		}
	}
	return true
}

func englishIdentificationAttributesAllowed(source semantic.MeaningCandidate, predicateID meaning.NodeID, role string, referentID meaning.NodeID, coreID ids.ConceptID) bool {
	predicate, predicateExists := englishMeaningNode(source, predicateID)
	referent, referentExists := englishMeaningNode(source, referentID)
	if !predicateExists || !referentExists || referent.Kind != meaning.KindReferent {
		return false
	}
	if predicate.TemporalFrame == meaning.TemporalFramePast {
		if role != "bearer" {
			return false
		}
		if _, attribution := englishAttributionForConcept(string(predicate.Concept)); !attribution {
			return false
		}
	}
	return englishFrameEvidenceConsistent(source, predicateID, coreID) &&
		englishMultiplicityEvidenceConsistent(source, predicateID, meaning.Role(role), referent, coreID, true)
}

func englishIdentificationDefaults(source semantic.MeaningCandidate, predicatePath, referentPath []meaning.NodeID, predicate, referent meaning.Node, predicateConcept ids.ConceptID) []realizer.FieldAssumption {
	assumptions := make([]realizer.FieldAssumption, 0, 4)
	if predicate.TemporalFrame == "" {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "tense", Value: "present", Reason: "Meaning carries no temporal frame; this licensed English identification route uses present"})
	}
	// A nominal or demonstrative subject uses third-person agreement by
	// construction; it does not posit a new semantic participant feature.
	personKnown := referent.Kind == meaning.KindReferent && !englishParticipantAnchorIntent(source, referent.ID)
	if values, found := runtime.SourceFeaturesAt(source, predicate.ID, "", predicateConcept); found {
		personKnown = personKnown || values["person"] != ""
	}
	if !personKnown {
		assumptions = append(assumptions, realizer.FieldAssumption{Path: predicatePath, Field: "person", Value: "third", Reason: "Meaning carries no grammatical person; this route uses the approved third-person form"})
	}
	if referent.Multiplicity == "" {
		assumptions = append(assumptions,
			realizer.FieldAssumption{Path: predicatePath, Field: "number", Value: "singular", Reason: "Meaning carries no multiplicity; this licensed class predicate uses the singular form"},
			realizer.FieldAssumption{Path: referentPath, Field: "number", Value: "singular", Reason: "Meaning carries no multiplicity; this licensed count noun uses the singular form"},
		)
	}
	return assumptions
}

func englishPredicateNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, candidates []realizer.TargetCandidate) (string, bool, bool) {
	if source.Meaning == nil {
		return "", false, false
	}
	role := englishPredicateSubjectRoleForTarget(predicate)
	if role == "" {
		return "singular", false, true
	}
	referent, exists := englishReferentForRole(source, predicate.NodeID, role)
	if !exists {
		return "", false, false
	}
	number := ""
	semanticNumber := referent.Multiplicity != ""
	found := false
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != referent.ID || candidate.Outcome != status.OutcomeOK && candidate.Outcome != status.OutcomeIncomplete {
			continue
		}
		candidateNumber, _, ok := englishTargetNumber(source, referent.ID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if !ok || found && number != candidateNumber {
			return "", semanticNumber, false
		}
		number, found = candidateNumber, true
	}
	if !found {
		if value := englishMultiplicityNumber(referent.Multiplicity); value != "" {
			return value, true, true
		}
		return "singular", false, true
	}
	return number, semanticNumber, true
}

func englishPresentNegationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route englishPresentNegationRoute, candidates []realizer.TargetCandidate) (string, bool) {
	_, event, sourceRoute, found := englishNegatedPresent(source)
	if !found || event.ID != predicate.NodeID || sourceRoute.family != route.family || sourceRoute.sense != route.sense || event.TemporalFrame == meaning.TemporalFramePast {
		return "", false
	}
	referent, exists := englishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || referent.Multiplicity != "" && !englishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, predicate.TransferBinding.CoreRef, true) {
		return "", false
	}
	number, _, ok := englishPredicateNumber(source, predicate, candidates)
	if !ok || !englishFrameEvidenceConsistent(source, event.ID, predicate.TransferBinding.CoreRef) {
		return "", false
	}
	return number, true
}

// englishCountRoleNumber uses semantic cardinality first. Source grammatical
// number can refine an otherwise open count NP only when its lexical contract
// does not explicitly keep cardinality unresolved (for example pluralia tantum).
func englishCountRoleNumber(source semantic.MeaningCandidate, index *lexicon.Index, predicateID meaning.NodeID, role meaning.Role, referent meaning.Node) string {
	if number := englishMultiplicityNumber(referent.Multiplicity); number != "" {
		return number
	}
	concept := coreConcept(referent)
	if index != nil {
		for _, record := range source.Evidence.Records {
			if !runtime.SourceEvidenceAppliesAt(source, record, predicateID, role, concept) {
				continue
			}
			sense, found := index.BySenseID(record.SenseID)
			if !found {
				continue
			}
			word, found := index.ByLexemeID(sense.LexemeID)
			if !found {
				continue
			}
			for _, property := range word.Properties {
				if property.Key == "multiplicity_behavior" && property.Value == "lexical_plural_not_cardinality" {
					return ""
				}
			}
		}
	}
	if values, found := runtime.SourceFeaturesAt(source, predicateID, role, concept); found {
		if values["number"] != "" {
			return string(values["number"])
		}
		return string(values["noun_number"])
	}
	return ""
}

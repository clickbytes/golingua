package valency

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
)

// ValidationRequest identifies the selected event/state and the two approved
// links needed to realize it: lexical Sense→CORE and Sense→frame→construction.
type ValidationRequest struct {
	Meaning             meaning.Meaning
	NodeID              meaning.NodeID
	SemanticBinding     lexicon.SemanticBinding
	ConstructionBinding ConstructionBinding
}

// ValidateFrame checks one candidate frame against one explicitly selected
// Meaning node and immutable CORE snapshot. It does not select among senses;
// callers must validate each candidate separately and preserve ambiguity.
func ValidateFrame(frame ValencyFrame, request ValidationRequest, ontology *core.ConceptGraph) (ValidationResult, error) {
	if err := frame.Validate(); err != nil {
		return ValidationResult{}, err
	}
	if ontology == nil {
		return ValidationResult{}, invalid("INVALID_CORE_GRAPH", "core", "ontology snapshot is nil")
	}
	if err := meaning.ValidateMeaning(request.Meaning, ontology, meaning.DefaultBudget()); err != nil {
		return ValidationResult{}, err
	}
	construction := request.ConstructionBinding
	if err := validateBindingStructure(construction); err != nil {
		return ValidationResult{}, err
	}
	if construction.SenseID != frame.SenseID {
		return ValidationResult{}, invalid("SENSE_FRAME_IDENTITY_MISMATCH", "construction_binding.sense_id", "construction binding sense must match the frame sense")
	}
	if construction.ValencyFrameID != frame.ID {
		return ValidationResult{}, invalid("FRAME_IDENTITY_MISMATCH", "construction_binding.valency_frame_id", "construction binding must reference this exact valency frame")
	}
	if !containsConstruction(frame.AllowedFamilyIDs, construction.FamilyID) {
		return ValidationResult{}, unsupported("FAMILY_NOT_LICENSED_BY_FRAME", "construction_binding.family_id", "valency frame does not license this construction family")
	}
	semantic := request.SemanticBinding
	if !semantic.TransferAuthorized() {
		return ValidationResult{}, unsupported("SEMANTIC_TRANSFER_NOT_APPROVED", "semantic_binding", "frame validation requires an approved equivalent Sense-to-CORE binding with provenance")
	}
	if semantic.SenseID != frame.SenseID {
		return ValidationResult{}, invalid("SENSE_BINDING_IDENTITY_MISMATCH", "semantic_binding.sense_id", "semantic binding sense must match the frame sense")
	}

	var selected *meaning.Node
	for index := range request.Meaning.Nodes {
		node := &request.Meaning.Nodes[index]
		if node.ID == request.NodeID {
			selected = node
			break
		}
	}
	if selected == nil {
		return ValidationResult{}, unknown("UNKNOWN_MEANING_NODE", "node_id", "selected Meaning node does not exist")
	}
	if selected.Kind != meaning.KindEvent && selected.Kind != meaning.KindState {
		return ValidationResult{}, invalid("INVALID_VALENCY_TARGET", "node_id", "valency frame must validate an event or state node")
	}
	conceptID, err := ids.ParseConceptID(string(selected.Concept))
	if err != nil || conceptID != semantic.CoreRef {
		return ValidationResult{}, invalid("CORE_BINDING_IDENTITY_MISMATCH", "meaning_node.concept", "selected Meaning concept must match the approved semantic binding")
	}
	concept, exists := ontology.Lookup(conceptID)
	if !exists {
		return ValidationResult{}, unknown("UNKNOWN_CONCEPT", "semantic_binding.core_ref", "bound CORE concept does not exist")
	}
	declaredRoles := make(map[string]struct{}, len(concept.RoleTypes))
	for _, role := range concept.RoleTypes {
		declaredRoles[role.Name] = struct{}{}
	}
	frameRoles := make(map[meaning.Role]ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		if _, exists := declaredRoles[string(argument.Role)]; !exists {
			return ValidationResult{}, invalid("FRAME_ROLE_NOT_IN_CORE", "frame.arguments."+argument.ID, "frame role must be declared by the bound CORE concept")
		}
		frameRoles[argument.Role] = argument
	}
	participants := make(map[meaning.Role]meaning.Node, len(selected.Participants))
	for _, participant := range selected.Participants {
		if _, exists := frameRoles[participant.Role]; !exists {
			return ValidationResult{}, unsupported("EXTRA_SEMANTIC_ROLE", "meaning_node.participants."+string(participant.Role), "selected frame does not license this semantic role")
		}
		target, exists := meaningNode(request.Meaning, participant.Target)
		if !exists {
			return ValidationResult{}, unknown("UNKNOWN_ROLE_TARGET", "meaning_node.participants."+string(participant.Role), "participant target does not resolve")
		}
		participants[participant.Role] = target
	}
	for role, argument := range frameRoles {
		if _, exists := participants[role]; !exists && argument.Required {
			return ValidationResult{}, incomplete("MISSING_REQUIRED_ROLE", "meaning_node.participants."+string(role), "required semantic role is missing from the selected Meaning")
		}
	}
	result := ValidationResult{SoftConstraintMisses: []SoftConstraintMiss{}}
	for _, constraint := range frame.SelectionalConstraints {
		target, present := participants[constraint.Role]
		if !present {
			continue
		}
		for _, allowedID := range constraint.AllowedConceptIDs {
			if _, exists := ontology.Lookup(allowedID); !exists {
				return ValidationResult{}, unknown("UNKNOWN_SELECTIONAL_CONCEPT", "frame.selectional_constraints."+constraint.ID, "selectional constraint refers to a CORE concept that does not exist")
			}
		}
		targetID, err := ids.ParseConceptID(string(target.Concept))
		if err != nil {
			return ValidationResult{}, invalid("INVALID_ROLE_TARGET_CONCEPT", "meaning_node.participants."+string(constraint.Role), "role target has a malformed CORE concept ID")
		}
		targetConcept, exists := ontology.Lookup(targetID)
		if !exists {
			return ValidationResult{}, unknown("UNKNOWN_ROLE_TARGET_CONCEPT", "meaning_node.participants."+string(constraint.Role), "role target concept is absent from CORE")
		}
		matches := constraintMatches(constraint, targetID, targetConcept.Kind, ontology)
		if matches {
			continue
		}
		if constraint.Strength == ConstraintHard {
			return ValidationResult{}, unsupported("HARD_SELECTIONAL_RESTRICTION", "meaning_node.participants."+string(constraint.Role), "role target violates a hard selectional restriction")
		}
		result.SoftConstraintMisses = append(result.SoftConstraintMisses, SoftConstraintMiss{ConstraintID: constraint.ID, Role: constraint.Role, Actual: targetID})
	}
	sortSoftMisses(result.SoftConstraintMisses)
	return result, nil
}

func meaningNode(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func containsConstruction(values []ids.ConstructionID, candidate ids.ConstructionID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func constraintMatches(constraint SelectionalConstraint, actual ids.ConceptID, actualKind core.ConceptKind, ontology *core.ConceptGraph) bool {
	for _, kind := range constraint.AllowedKinds {
		if kind == actualKind {
			return true
		}
	}
	for _, allowed := range constraint.AllowedConceptIDs {
		if isSubtypeOrSame(ontology, actual, allowed) {
			return true
		}
	}
	return false
}

func isSubtypeOrSame(ontology *core.ConceptGraph, actual, allowed ids.ConceptID) bool {
	current := actual
	for {
		if current == allowed {
			return true
		}
		concept, exists := ontology.Lookup(current)
		if !exists || concept.Parent == nil {
			return false
		}
		current = *concept.Parent
	}
}

func sortSoftMisses(misses []SoftConstraintMiss) {
	for i := 1; i < len(misses); i++ {
		for j := i; j > 0; j-- {
			left, right := misses[j-1], misses[j]
			if left.ConstraintID < right.ConstraintID || (left.ConstraintID == right.ConstraintID && left.Role <= right.Role) {
				break
			}
			misses[j-1], misses[j] = misses[j], misses[j-1]
		}
	}
}

// validateBindingStructure is shared by index construction and selector use.
func validateBindingStructure(binding ConstructionBinding) error {
	if !binding.Approved {
		return unsupported("UNAPPROVED_CONSTRUCTION_BINDING", "binding.approved", "only explicitly approved construction bindings may enter the runtime index")
	}
	if !validSenseID(binding.SenseID) || !validFrameID(binding.ValencyFrameID) || !validConstructionID(binding.FamilyID) || !validConstructionID(binding.VariantID) {
		return invalid("INVALID_CONSTRUCTION_BINDING", "binding", "binding requires canonical sense, frame, family and variant IDs")
	}
	if err := binding.Provenance.Validate(); err != nil {
		return invalid("INVALID_CONSTRUCTION_PROVENANCE", "binding.provenance", fmt.Sprintf("construction binding provenance is invalid: %v", err))
	}
	return nil
}

func validateCandidateStructure(candidate ConstructionCandidate) error {
	if !validSenseID(candidate.SenseID) || !validFrameID(candidate.ValencyFrameID) || !validConstructionID(candidate.FamilyID) || !validConstructionID(candidate.VariantID) {
		return invalid("INVALID_CONSTRUCTION_CANDIDATE", "candidate", "candidate requires canonical sense, frame, family and variant IDs")
	}
	if err := candidate.Provenance.Validate(); err != nil {
		return invalid("INVALID_CONSTRUCTION_CANDIDATE_PROVENANCE", "candidate.provenance", fmt.Sprintf("construction candidate provenance is invalid: %v", err))
	}
	return nil
}

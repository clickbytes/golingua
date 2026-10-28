// Package valency models sense-specific semantic argument licensing. It keeps
// semantic roles separate from language-specific grammatical functions,
// case, and adposition policy.
package valency

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxFrames                = 100_000
	MaxBindings              = 250_000
	MaxArgumentsPerFrame     = 64
	MaxSemanticRoles         = 64
	MaxConstraintsPerFrame   = 256
	MaxConceptsPerConstraint = 64
	MaxKindsPerConstraint    = 16
	MaxFamiliesPerFrame      = 128
	MaxIndexItems            = 1_000_000
	MaxFrameTextBytes        = 256
)

// FrameID identifies one manually authored valency frame. It is a bounded,
// lowercase local identifier; sense and construction references use typed IDs.
type FrameID string

// ArgumentSlot assigns a semantic role to one argument position in a frame.
// Required is explicit; optional slots remain absent unless Meaning supplies
// their role. GrammaticalFunction, Case, and Adposition are descriptive
// language-level licensing data and do not follow from SemanticRole.
type ArgumentSlot struct {
	ID                  string       `json:"id"`
	Role                meaning.Role `json:"role"`
	Required            bool         `json:"required"`
	GrammaticalFunction string       `json:"grammatical_function,omitempty"`
	Case                string       `json:"case,omitempty"`
	Adposition          string       `json:"adposition,omitempty"`
}

// ConstraintStrength distinguishes a hard selection restriction from a soft
// preference. A soft mismatch is reported but never rejects a frame.
type ConstraintStrength string

const (
	ConstraintHard ConstraintStrength = "hard"
	ConstraintSoft ConstraintStrength = "soft"
)

// SelectionalConstraint restricts the CORE concepts that may fill one role.
// Concept IDs match exact concepts or their CORE descendants; AllowedKinds
// are checked against the target's declared CORE kind.
type SelectionalConstraint struct {
	ID                 string             `json:"id"`
	Role               meaning.Role       `json:"role"`
	Strength           ConstraintStrength `json:"strength"`
	AllowedConceptIDs  []ids.ConceptID    `json:"allowed_concept_ids,omitempty"`
	ExcludedConceptIDs []ids.ConceptID    `json:"excluded_concept_ids,omitempty"`
	AllowedKinds       []core.ConceptKind `json:"allowed_kinds,omitempty"`
}

// ValencyFrame licenses semantic argument roles for one lexical sense. It
// intentionally has no subject=agent or object=theme default.
type ValencyFrame struct {
	ID                     FrameID                 `json:"id"`
	SenseID                ids.SenseID             `json:"sense_id"`
	SemanticRoles          []meaning.Role          `json:"semantic_roles"`
	Arguments              []ArgumentSlot          `json:"arguments"`
	SelectionalConstraints []SelectionalConstraint `json:"selectional_constraints,omitempty"`
	AllowedFamilyIDs       []ids.ConstructionID    `json:"allowed_family_ids"`
	Provenance             provenance.SourceRef    `json:"provenance"`
}

// ConstructionBinding is an explicitly approved link from a lexical sense
// through its valency frame to one construction family and variant.
type ConstructionBinding struct {
	SenseID        ids.SenseID          `json:"sense_id"`
	ValencyFrameID FrameID              `json:"valency_frame_id"`
	FamilyID       ids.ConstructionID   `json:"family_id"`
	VariantID      ids.ConstructionID   `json:"variant_id"`
	Approved       bool                 `json:"approved"`
	Provenance     provenance.SourceRef `json:"provenance"`
}

// ConstructionCandidate is an unapproved proposal for a sense/frame/family/
// variant route. It is stored separately from ConstructionBinding and can
// never authorize parsing or realization.
type ConstructionCandidate struct {
	SenseID        ids.SenseID          `json:"sense_id"`
	ValencyFrameID FrameID              `json:"valency_frame_id"`
	FamilyID       ids.ConstructionID   `json:"family_id"`
	VariantID      ids.ConstructionID   `json:"variant_id"`
	Provenance     provenance.SourceRef `json:"provenance"`
}

// SoftConstraintMiss records a preference mismatch without treating it as a
// refusal or changing the selected frame.
type SoftConstraintMiss struct {
	ConstraintID string        `json:"constraint_id"`
	Role         meaning.Role  `json:"role"`
	Actual       ids.ConceptID `json:"actual_concept"`
}

// ValidationResult carries accepted soft preference misses separately from
// hard validation failures.
type ValidationResult struct {
	SoftConstraintMisses []SoftConstraintMiss `json:"soft_constraint_misses,omitempty"`
}

// Validate checks a frame's internal structure and canonical identities.
func (frame ValencyFrame) Validate() error {
	if !validFrameID(frame.ID) {
		return invalid("INVALID_VALENCY_FRAME_ID", "frame.id", "frame ID must be a bounded lowercase identifier")
	}
	if !validSenseID(frame.SenseID) {
		return invalid("INVALID_SENSE_ID", "frame.sense_id", "frame requires a canonical typed sense ID")
	}
	if err := frame.Provenance.Validate(); err != nil {
		return invalid("INVALID_FRAME_PROVENANCE", "frame.provenance", "valency frame requires valid provenance")
	}
	if len(frame.Arguments) == 0 || len(frame.Arguments) > MaxArgumentsPerFrame {
		return resourceLimitOrInvalid("arguments", len(frame.Arguments), MaxArgumentsPerFrame)
	}
	if len(frame.SelectionalConstraints) > MaxConstraintsPerFrame {
		return resourceLimit("constraints", len(frame.SelectionalConstraints), MaxConstraintsPerFrame)
	}
	if len(frame.SemanticRoles) > MaxSemanticRoles {
		return resourceLimit("semantic_roles", len(frame.SemanticRoles), MaxSemanticRoles)
	}
	if len(frame.AllowedFamilyIDs) > MaxFamiliesPerFrame {
		return resourceLimit("allowed_families", len(frame.AllowedFamilyIDs), MaxFamiliesPerFrame)
	}
	if len(frame.AllowedFamilyIDs) == 0 {
		return invalid("MISSING_ALLOWED_FAMILY", "frame.allowed_family_ids", "frame must explicitly license at least one construction family")
	}
	roles := make(map[meaning.Role]struct{}, len(frame.SemanticRoles))
	for index, role := range frame.SemanticRoles {
		if !validToken(string(role)) {
			return invalid("INVALID_SEMANTIC_ROLE", fmt.Sprintf("frame.semantic_roles[%d]", index), "semantic role must be a bounded lowercase token")
		}
		if _, exists := roles[role]; exists {
			return invalid("DUPLICATE_SEMANTIC_ROLE", fmt.Sprintf("frame.semantic_roles[%d]", index), "semantic role is duplicated")
		}
		roles[role] = struct{}{}
	}
	if len(roles) == 0 {
		return invalid("EMPTY_SEMANTIC_ROLES", "frame.semantic_roles", "frame must declare at least one semantic role")
	}
	arguments := make(map[string]struct{}, len(frame.Arguments))
	argumentRoles := make(map[meaning.Role]struct{}, len(frame.Arguments))
	for index, argument := range frame.Arguments {
		path := fmt.Sprintf("frame.arguments[%d]", index)
		if !validToken(argument.ID) {
			return invalid("INVALID_ARGUMENT_SLOT_ID", path+".id", "argument slot ID must be a bounded lowercase token")
		}
		if _, exists := arguments[argument.ID]; exists {
			return invalid("DUPLICATE_ARGUMENT_SLOT_ID", path+".id", "argument slot ID is duplicated")
		}
		arguments[argument.ID] = struct{}{}
		if _, exists := roles[argument.Role]; !exists {
			return invalid("UNDECLARED_ARGUMENT_ROLE", path+".role", "argument role must be declared in semantic_roles")
		}
		if _, exists := argumentRoles[argument.Role]; exists {
			return invalid("DUPLICATE_ARGUMENT_ROLE", path+".role", "each role has one bounded slot in this Meaning schema")
		}
		argumentRoles[argument.Role] = struct{}{}
		for field, value := range map[string]string{"grammatical_function": argument.GrammaticalFunction, "case": argument.Case, "adposition": argument.Adposition} {
			if value != "" && !validToken(value) {
				return invalid("INVALID_ARGUMENT_POLICY", path+"."+field, "language-specific policy must be a bounded lowercase token")
			}
		}
	}
	if len(argumentRoles) != len(roles) {
		return invalid("INCOMPLETE_ARGUMENT_SLOTS", "frame.arguments", "every declared semantic role requires exactly one argument slot")
	}
	constraintIDs := make(map[string]struct{}, len(frame.SelectionalConstraints))
	for index, constraint := range frame.SelectionalConstraints {
		path := fmt.Sprintf("frame.selectional_constraints[%d]", index)
		if !validToken(constraint.ID) {
			return invalid("INVALID_CONSTRAINT_ID", path+".id", "constraint ID must be a bounded lowercase token")
		}
		if _, exists := constraintIDs[constraint.ID]; exists {
			return invalid("DUPLICATE_CONSTRAINT_ID", path+".id", "constraint ID is duplicated")
		}
		constraintIDs[constraint.ID] = struct{}{}
		if _, exists := roles[constraint.Role]; !exists {
			return invalid("UNKNOWN_CONSTRAINT_ROLE", path+".role", "constraint role is not declared by the frame")
		}
		if constraint.Strength != ConstraintHard && constraint.Strength != ConstraintSoft {
			return invalid("INVALID_CONSTRAINT_STRENGTH", path+".strength", "constraint strength must be hard or soft")
		}
		if len(constraint.AllowedConceptIDs) == 0 && len(constraint.AllowedKinds) == 0 && len(constraint.ExcludedConceptIDs) == 0 {
			return invalid("EMPTY_SELECTIONAL_CONSTRAINT", path, "constraint must name allowed concept IDs or concept kinds")
		}
		if len(constraint.ExcludedConceptIDs) > MaxConceptsPerConstraint {
			return resourceLimit("constraint_excluded_concepts", len(constraint.ExcludedConceptIDs), MaxConceptsPerConstraint)
		}
		seenExcluded := make(map[string]struct{}, len(constraint.ExcludedConceptIDs))
		for conceptIndex, conceptID := range constraint.ExcludedConceptIDs {
			if !validConceptID(conceptID) {
				return invalid("INVALID_CONSTRAINT_CONCEPT_ID", fmt.Sprintf("%s.excluded_concept_ids[%d]", path, conceptIndex), "excluded concept ID must be canonical")
			}
			if _, exists := seenExcluded[conceptID.String()]; exists {
				return invalid("DUPLICATE_EXCLUDED_CONSTRAINT_CONCEPT", path+".excluded_concept_ids", "excluded concept ID is duplicated")
			}
			seenExcluded[conceptID.String()] = struct{}{}
		}
		if len(constraint.AllowedConceptIDs) > MaxConceptsPerConstraint {
			return resourceLimit("constraint_concepts", len(constraint.AllowedConceptIDs), MaxConceptsPerConstraint)
		}
		if len(constraint.AllowedKinds) > MaxKindsPerConstraint {
			return resourceLimit("constraint_kinds", len(constraint.AllowedKinds), MaxKindsPerConstraint)
		}
		seenConcepts := make(map[string]struct{}, len(constraint.AllowedConceptIDs))
		for conceptIndex, conceptID := range constraint.AllowedConceptIDs {
			if !validConceptID(conceptID) {
				return invalid("INVALID_CONSTRAINT_CONCEPT_ID", fmt.Sprintf("%s.allowed_concept_ids[%d]", path, conceptIndex), "constraint concept ID must be canonical")
			}
			if _, exists := seenConcepts[conceptID.String()]; exists {
				return invalid("DUPLICATE_CONSTRAINT_CONCEPT", path+".allowed_concept_ids", "allowed concept ID is duplicated")
			}
			seenConcepts[conceptID.String()] = struct{}{}
		}
		seenKinds := make(map[core.ConceptKind]struct{}, len(constraint.AllowedKinds))
		for kindIndex, kind := range constraint.AllowedKinds {
			if !kind.Valid() {
				return invalid("INVALID_CONSTRAINT_KIND", fmt.Sprintf("%s.allowed_kinds[%d]", path, kindIndex), "allowed concept kind is unsupported")
			}
			if _, exists := seenKinds[kind]; exists {
				return invalid("DUPLICATE_CONSTRAINT_KIND", path+".allowed_kinds", "allowed concept kind is duplicated")
			}
			seenKinds[kind] = struct{}{}
		}
	}
	seenFamilies := make(map[string]struct{}, len(frame.AllowedFamilyIDs))
	for index, familyID := range frame.AllowedFamilyIDs {
		if !validConstructionID(familyID) {
			return invalid("INVALID_CONSTRUCTION_FAMILY_ID", fmt.Sprintf("frame.allowed_family_ids[%d]", index), "family ID must be a canonical typed construction ID")
		}
		if _, exists := seenFamilies[familyID.String()]; exists {
			return invalid("DUPLICATE_CONSTRUCTION_FAMILY_ID", "frame.allowed_family_ids", "family ID is duplicated")
		}
		seenFamilies[familyID.String()] = struct{}{}
	}
	return nil
}

func validFrameID(id FrameID) bool {
	value := string(id)
	if value == "" || len(value) > MaxFrameTextBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func validToken(value string) bool {
	if value == "" || len(value) > MaxFrameTextBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func validSenseID(id ids.SenseID) bool {
	parsed, err := ids.ParseSenseID(id.String())
	return err == nil && parsed.String() == id.String()
}

func validConceptID(id ids.ConceptID) bool {
	parsed, err := ids.ParseConceptID(id.String())
	return err == nil && parsed.String() == id.String()
}

func validConstructionID(id ids.ConstructionID) bool {
	parsed, err := ids.ParseConstructionID(id.String())
	return err == nil && parsed.String() == id.String()
}

func canonicalFrame(frame ValencyFrame) ValencyFrame {
	result := cloneFrame(frame)
	sort.Slice(result.SemanticRoles, func(i, j int) bool { return result.SemanticRoles[i] < result.SemanticRoles[j] })
	sort.Slice(result.Arguments, func(i, j int) bool { return result.Arguments[i].ID < result.Arguments[j].ID })
	for index := range result.SelectionalConstraints {
		constraint := &result.SelectionalConstraints[index]
		sort.Slice(constraint.AllowedConceptIDs, func(i, j int) bool {
			return constraint.AllowedConceptIDs[i].String() < constraint.AllowedConceptIDs[j].String()
		})
		sort.Slice(constraint.ExcludedConceptIDs, func(i, j int) bool {
			return constraint.ExcludedConceptIDs[i].String() < constraint.ExcludedConceptIDs[j].String()
		})
		sort.Slice(constraint.AllowedKinds, func(i, j int) bool { return constraint.AllowedKinds[i] < constraint.AllowedKinds[j] })
	}
	sort.Slice(result.SelectionalConstraints, func(i, j int) bool { return result.SelectionalConstraints[i].ID < result.SelectionalConstraints[j].ID })
	sort.Slice(result.AllowedFamilyIDs, func(i, j int) bool { return result.AllowedFamilyIDs[i].String() < result.AllowedFamilyIDs[j].String() })
	return result
}

func cloneFrame(frame ValencyFrame) ValencyFrame {
	result := frame
	result.SemanticRoles = append([]meaning.Role(nil), frame.SemanticRoles...)
	result.Arguments = append([]ArgumentSlot(nil), frame.Arguments...)
	result.SelectionalConstraints = make([]SelectionalConstraint, len(frame.SelectionalConstraints))
	for index, constraint := range frame.SelectionalConstraints {
		result.SelectionalConstraints[index] = constraint
		result.SelectionalConstraints[index].AllowedConceptIDs = append([]ids.ConceptID(nil), constraint.AllowedConceptIDs...)
		result.SelectionalConstraints[index].ExcludedConceptIDs = append([]ids.ConceptID(nil), constraint.ExcludedConceptIDs...)
		result.SelectionalConstraints[index].AllowedKinds = append([]core.ConceptKind(nil), constraint.AllowedKinds...)
	}
	result.AllowedFamilyIDs = append([]ids.ConstructionID(nil), frame.AllowedFamilyIDs...)
	return result
}

func cloneBinding(binding ConstructionBinding) ConstructionBinding { return binding }

func cloneCandidate(candidate ConstructionCandidate) ConstructionCandidate { return candidate }

func resourceLimitOrInvalid(name string, count, maximum int) error {
	if count > maximum {
		return resourceLimit(name, count, maximum)
	}
	return invalid("EMPTY_"+strings.ToUpper(name), "frame."+name, "frame requires at least one "+name)
}

func invalid(code, path, message string) error {
	return valencyError(status.OutcomeInvalid, code, path, message)
}

func unknown(code, path, message string) error {
	return valencyError(status.OutcomeUnknown, code, path, message)
}

func incomplete(code, path, message string) error {
	return valencyError(status.OutcomeIncomplete, code, path, message)
}

func unsupported(code, path, message string) error {
	return valencyError(status.OutcomeUnsupported, code, path, message)
}

func resourceLimit(resource string, value, maximum int) error {
	return valencyError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("valency %s limit exceeded: %d > %d", resource, value, maximum))
}

func valencyError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &Error{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// Error retains the typed outcome and bounded diagnostic for a valency failure.
type Error struct{ result status.Result }

func (failure *Error) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "valency operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed outcome.
func (failure *Error) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

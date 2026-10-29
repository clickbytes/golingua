package construction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	MaxFamilies      = 100_000
	MaxLanguageBytes = 35
)

// ConstructionPlanID is a deterministic digest of one validated family
// structure. It is not a surface string or a ConceptID.
type ConstructionPlanID string

// Valid reports whether the plan identity is the canonical SHA-256 form.
func (id ConstructionPlanID) Valid() bool {
	value := string(id)
	if len(value) != len("plan:")+sha256.Size*2 || !strings.HasPrefix(value, "plan:") || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "plan:"))
	return err == nil
}

// OperatorID identifies an operator the family permits. Operators are typed
// metadata here; their scope and composition are compiled in a later stage.
type OperatorID string

// Capability identifies a bounded family-level operation or plan feature.
type Capability string

const CapabilityImplicitParticipant Capability = "implicit_participant"

// ConstructionFamily is one immutable language-local structural template.
// It is independent of any ConceptID or SenseID; sense-specific licensing is
// held by approved valency.ConstructionBinding records.
type ConstructionFamily struct {
	ID           ids.ConstructionID   `json:"id"`
	Language     features.LanguageID  `json:"language"`
	SlotSpecs    []SlotSpec           `json:"slot_specs"`
	Constraints  []FamilyConstraint   `json:"constraints,omitempty"`
	Operators    []OperatorID         `json:"operators,omitempty"`
	Capabilities []Capability         `json:"capabilities,omitempty"`
	Provenance   provenance.SourceRef `json:"provenance"`
}

// ImplicitParticipantFamilySlot returns an eligible optional person-role NP
// slot only when the family explicitly grants implicit-participant authority.
func ImplicitParticipantFamilySlot(family ConstructionFamily, slotID string) (SlotSpec, bool) {
	if family.Validate() != nil {
		return SlotSpec{}, false
	}
	licensed := false
	for _, capability := range family.Capabilities {
		if capability == CapabilityImplicitParticipant {
			licensed = true
			break
		}
	}
	if !licensed {
		return SlotSpec{}, false
	}
	for index, slot := range family.SlotSpecs {
		if slot.ID != slotID {
			continue
		}
		if slot.Category != CategoryNP || slot.ArgumentBinding == nil || slot.Optionality != SlotOptional ||
			slot.Cardinality.Minimum != 0 || slot.Cardinality.Maximum != 1 {
			return SlotSpec{}, false
		}
		copyOf := cloneFamily(family)
		return copyOf.SlotSpecs[index], true
	}
	return SlotSpec{}, false
}

// ImplicitParticipantSlot reports whether a selected variant omits one
// eligible zero-minimum, single, role-bound NP under family authority.
func ImplicitParticipantSlot(family ConstructionFamily, variant Variant, slotID string) (SlotSpec, bool) {
	if variant.FamilyID != family.ID || variant.Validate(family) != nil {
		return SlotSpec{}, false
	}
	slot, eligible := ImplicitParticipantFamilySlot(family, slotID)
	if !eligible {
		return SlotSpec{}, false
	}
	for _, token := range variant.Sequence {
		if token.Kind == TokenSlot && token.SlotID == slotID {
			return SlotSpec{}, false
		}
	}
	return slot, true
}

// Validate checks identities, bounds, slot references, and provenance.
func (family ConstructionFamily) Validate() error {
	if !validFamilyID(family.ID) {
		return invalid("INVALID_CONSTRUCTION_FAMILY_ID", "family.id", "family ID must be a canonical typed construction ID")
	}
	if !validLanguage(family.Language) {
		return invalid("INVALID_FAMILY_LANGUAGE", "family.language", "family language must be a canonical lowercase language tag")
	}
	if len(family.SlotSpecs) == 0 || len(family.SlotSpecs) > MaxSlotsPerFamily {
		if len(family.SlotSpecs) > MaxSlotsPerFamily {
			return resourceLimit("slots", len(family.SlotSpecs), MaxSlotsPerFamily)
		}
		return invalid("EMPTY_CONSTRUCTION_FAMILY", "family.slot_specs", "family requires at least one phrase slot")
	}
	if len(family.Constraints) > MaxConstraintsPerFamily {
		return resourceLimit("constraints", len(family.Constraints), MaxConstraintsPerFamily)
	}
	if len(family.Operators) > MaxFamilyItems || len(family.Capabilities) > MaxFamilyItems {
		return resourceLimit("family_items", len(family.Operators)+len(family.Capabilities), MaxFamilyItems)
	}
	itemCount := len(family.SlotSpecs) + len(family.Constraints) + len(family.Operators) + len(family.Capabilities)
	if itemCount > MaxFamilyItems {
		return resourceLimit("family_items", itemCount, MaxFamilyItems)
	}
	if err := family.Provenance.Validate(); err != nil {
		return invalid("INVALID_FAMILY_PROVENANCE", "family.provenance", "construction family requires valid provenance")
	}
	slotsByID := make(map[string]SlotSpec, len(family.SlotSpecs))
	positions := make(map[int]struct{}, len(family.SlotSpecs))
	argumentRoles := make(map[string]struct{})
	modifierSlots := make([]SlotSpec, 0, 1)
	for index, slot := range family.SlotSpecs {
		if err := validateSlot(slot, index); err != nil {
			return err
		}
		if _, exists := slotsByID[slot.ID]; exists {
			return invalid("DUPLICATE_SLOT_ID", "family.slot_specs", "slot ID is duplicated")
		}
		if _, exists := positions[slot.Position]; exists {
			return invalid("DUPLICATE_SLOT_POSITION", "family.slot_specs", "slot position is duplicated")
		}
		slotsByID[slot.ID] = slot
		positions[slot.Position] = struct{}{}
		if slot.ArgumentBinding != nil {
			role := string(slot.ArgumentBinding.SemanticRole)
			if _, exists := argumentRoles[role]; exists {
				return invalid("DUPLICATE_ARGUMENT_ROLE_BINDING", "family.slot_specs."+slot.ID, "one family slot may bind a semantic role only once")
			}
			argumentRoles[role] = struct{}{}
		}
		if slot.ModifierBinding != nil {
			modifierSlots = append(modifierSlots, slot)
		}
	}
	for _, slot := range modifierSlots {
		target, exists := slotsByID[slot.ModifierBinding.TargetSlot]
		if !exists {
			return unknown("UNKNOWN_MODIFIER_TARGET_SLOT", "family.slot_specs."+slot.ID+".modifier_binding.target_slot", "modifier target slot does not exist")
		}
		if target.ID == slot.ID || target.Category != CategoryNP || target.ArgumentBinding == nil || target.Optionality != SlotRequired || target.Cardinality.Minimum != 1 || target.Cardinality.Maximum != 1 {
			return invalid("INVALID_MODIFIER_TARGET_SLOT", "family.slot_specs."+slot.ID+".modifier_binding.target_slot", "modifier must target a distinct required single NP slot with an argument binding")
		}
	}
	for position := 0; position < len(family.SlotSpecs); position++ {
		if _, exists := positions[position]; !exists {
			return invalid("NONCONTIGUOUS_SLOT_ORDER", "family.slot_specs", "slot positions must be contiguous starting at zero")
		}
	}
	constraintIDs := make(map[string]struct{}, len(family.Constraints))
	for index, constraint := range family.Constraints {
		path := fmt.Sprintf("family.constraints[%d]", index)
		if !validToken(constraint.ID) {
			return invalid("INVALID_FAMILY_CONSTRAINT_ID", path+".id", "constraint ID must be a bounded lowercase token")
		}
		if _, exists := constraintIDs[constraint.ID]; exists {
			return invalid("DUPLICATE_FAMILY_CONSTRAINT_ID", path+".id", "constraint ID is duplicated")
		}
		constraintIDs[constraint.ID] = struct{}{}
		if constraint.Strength != ConstraintHard && constraint.Strength != ConstraintSoft {
			return invalid("INVALID_FAMILY_CONSTRAINT_STRENGTH", path+".strength", "constraint strength must be hard or soft")
		}
		if constraint.ControllerSlot == "" || constraint.TargetSlot == "" || constraint.ControllerSlot == constraint.TargetSlot {
			return invalid("INVALID_FAMILY_CONSTRAINT_SLOTS", path, "constraint requires two distinct slot references")
		}
		if _, exists := slotsByID[constraint.ControllerSlot]; !exists {
			return unknown("UNKNOWN_CONSTRAINT_SLOT", path+".controller_slot", "constraint controller slot does not exist")
		}
		if _, exists := slotsByID[constraint.TargetSlot]; !exists {
			return unknown("UNKNOWN_CONSTRAINT_SLOT", path+".target_slot", "constraint target slot does not exist")
		}
		switch constraint.Kind {
		case ConstraintPrecedes:
			if len(constraint.Features) != 0 {
				return invalid("PRECEDENCE_HAS_FEATURES", path+".features", "precedence constraint cannot carry agreement features")
			}
			if slotsByID[constraint.ControllerSlot].Position >= slotsByID[constraint.TargetSlot].Position {
				return invalid("PRECEDENCE_CONFLICTS_WITH_SLOT_ORDER", path, "precedes constraint must agree with the family's ordered slots")
			}
		case ConstraintAgreement:
			if len(constraint.Features) == 0 {
				return invalid("AGREEMENT_WITHOUT_FEATURES", path+".features", "agreement constraint requires at least one feature dimension")
			}
			if len(constraint.Features) > MaxFeaturesPerConstraint {
				return resourceLimit("constraint_features", len(constraint.Features), MaxFeaturesPerConstraint)
			}
			seenFeatures := make(map[features.FeatureID]struct{}, len(constraint.Features))
			for featureIndex, feature := range constraint.Features {
				if !validToken(string(feature)) {
					return invalid("INVALID_AGREEMENT_FEATURE", fmt.Sprintf("%s.features[%d]", path, featureIndex), "agreement feature must be a bounded lowercase token")
				}
				if _, exists := seenFeatures[feature]; exists {
					return invalid("DUPLICATE_AGREEMENT_FEATURE", path+".features", "agreement feature is duplicated")
				}
				seenFeatures[feature] = struct{}{}
			}
		default:
			return invalid("INVALID_FAMILY_CONSTRAINT_KIND", path+".kind", "constraint kind must be precedes or agreement")
		}
	}
	if err := validateUniqueTokens("operators", family.Operators); err != nil {
		return err
	}
	if err := validateUniqueTokens("capabilities", family.Capabilities); err != nil {
		return err
	}
	return nil
}

func validFamilyID(id ids.ConstructionID) bool {
	parsed, err := ids.ParseConstructionID(id.String())
	return err == nil && parsed.String() == id.String()
}

func validLanguage(language features.LanguageID) bool {
	value := string(language)
	if len(value) == 0 || len(value) > MaxLanguageBytes || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if !(character >= 'a' && character <= 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

func validateUniqueTokens[T ~string](path string, values []T) error {
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if !validToken(string(value)) {
			return invalid("INVALID_FAMILY_"+strings.ToUpper(strings.TrimSuffix(path, "s")), fmt.Sprintf("family.%s[%d]", path, index), "value must be a bounded lowercase token")
		}
		if _, exists := seen[string(value)]; exists {
			return invalid("DUPLICATE_FAMILY_"+strings.ToUpper(strings.TrimSuffix(path, "s")), "family."+path, "value is duplicated")
		}
		seen[string(value)] = struct{}{}
	}
	return nil
}

func canonicalFamily(family ConstructionFamily) ConstructionFamily {
	result := cloneFamily(family)
	sort.Slice(result.SlotSpecs, func(i, j int) bool {
		if result.SlotSpecs[i].Position != result.SlotSpecs[j].Position {
			return result.SlotSpecs[i].Position < result.SlotSpecs[j].Position
		}
		return result.SlotSpecs[i].ID < result.SlotSpecs[j].ID
	})
	for index := range result.Constraints {
		sort.Slice(result.Constraints[index].Features, func(i, j int) bool {
			return result.Constraints[index].Features[i] < result.Constraints[index].Features[j]
		})
	}
	sort.Slice(result.Constraints, func(i, j int) bool { return result.Constraints[i].ID < result.Constraints[j].ID })
	sort.Slice(result.Operators, func(i, j int) bool { return result.Operators[i] < result.Operators[j] })
	sort.Slice(result.Capabilities, func(i, j int) bool { return result.Capabilities[i] < result.Capabilities[j] })
	return result
}

func cloneFamily(family ConstructionFamily) ConstructionFamily {
	result := family
	result.SlotSpecs = append([]SlotSpec(nil), family.SlotSpecs...)
	for index, slot := range result.SlotSpecs {
		if slot.ArgumentBinding != nil {
			binding := *slot.ArgumentBinding
			result.SlotSpecs[index].ArgumentBinding = &binding
		}
		if slot.ModifierBinding != nil {
			binding := *slot.ModifierBinding
			result.SlotSpecs[index].ModifierBinding = &binding
		}
	}
	result.Constraints = make([]FamilyConstraint, len(family.Constraints))
	for index, constraint := range family.Constraints {
		result.Constraints[index] = constraint
		result.Constraints[index].Features = append([]features.FeatureID(nil), constraint.Features...)
	}
	result.Operators = append([]OperatorID(nil), family.Operators...)
	result.Capabilities = append([]Capability(nil), family.Capabilities...)
	return result
}

func planID(family ConstructionFamily) ConstructionPlanID {
	canonical := canonicalFamily(family)
	var builder strings.Builder
	writePlanPart(&builder, string(canonical.Language))
	writePlanPart(&builder, strconv.Itoa(len(canonical.SlotSpecs)))
	for _, slot := range canonical.SlotSpecs {
		writePlanPart(&builder, slot.ID)
		writePlanPart(&builder, strconv.Itoa(slot.Position))
		writePlanPart(&builder, string(slot.Category))
		writePlanPart(&builder, strconv.Itoa(slot.Cardinality.Minimum))
		writePlanPart(&builder, strconv.Itoa(slot.Cardinality.Maximum))
		writePlanPart(&builder, string(slot.Optionality))
		writePlanPart(&builder, slot.Case)
		writePlanPart(&builder, slot.Adposition)
		if slot.ArgumentBinding == nil {
			writePlanPart(&builder, "-")
		} else {
			writePlanPart(&builder, string(slot.ArgumentBinding.SemanticRole))
		}
		if slot.ModifierBinding != nil {
			writePlanPart(&builder, "modifier_binding")
			writePlanPart(&builder, slot.ModifierBinding.TargetSlot)
			writePlanPart(&builder, string(slot.ModifierBinding.SemanticRole))
		}
	}
	writePlanPart(&builder, strconv.Itoa(len(canonical.Constraints)))
	for _, constraint := range canonical.Constraints {
		writePlanPart(&builder, constraint.ID)
		writePlanPart(&builder, string(constraint.Kind))
		writePlanPart(&builder, string(constraint.Strength))
		writePlanPart(&builder, constraint.ControllerSlot)
		writePlanPart(&builder, constraint.TargetSlot)
		writePlanPart(&builder, strconv.Itoa(len(constraint.Features)))
		for _, feature := range constraint.Features {
			writePlanPart(&builder, string(feature))
		}
	}
	writePlanPart(&builder, strconv.Itoa(len(canonical.Operators)))
	for _, operator := range canonical.Operators {
		writePlanPart(&builder, string(operator))
	}
	writePlanPart(&builder, strconv.Itoa(len(canonical.Capabilities)))
	for _, capability := range canonical.Capabilities {
		writePlanPart(&builder, string(capability))
	}
	digest := sha256.Sum256([]byte(builder.String()))
	return ConstructionPlanID("plan:" + hex.EncodeToString(digest[:]))
}

func writePlanPart(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func roleLayoutKey(slots []SlotSpec) string {
	canonical := append([]SlotSpec(nil), slots...)
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].Position < canonical[j].Position })
	var builder strings.Builder
	for _, slot := range canonical {
		writePlanPart(&builder, strconv.Itoa(slot.Position))
		writePlanPart(&builder, string(slot.Category))
		writePlanPart(&builder, strconv.Itoa(slot.Cardinality.Minimum))
		writePlanPart(&builder, strconv.Itoa(slot.Cardinality.Maximum))
		writePlanPart(&builder, string(slot.Optionality))
		if slot.ArgumentBinding == nil {
			writePlanPart(&builder, "-")
		} else {
			writePlanPart(&builder, string(slot.ArgumentBinding.SemanticRole))
		}
	}
	modifierCount := 0
	for _, slot := range canonical {
		if slot.ModifierBinding != nil {
			modifierCount++
		}
	}
	if modifierCount > 0 {
		writePlanPart(&builder, "modifier_bindings")
		writePlanPart(&builder, strconv.Itoa(modifierCount))
		for _, slot := range canonical {
			if slot.ModifierBinding == nil {
				continue
			}
			writePlanPart(&builder, strconv.Itoa(slot.Position))
			writePlanPart(&builder, string(slot.Category))
			writePlanPart(&builder, slot.ID)
			writePlanPart(&builder, slot.ModifierBinding.TargetSlot)
			writePlanPart(&builder, string(slot.ModifierBinding.SemanticRole))
		}
	}
	return builder.String()
}

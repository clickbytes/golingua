package construction

import (
	"sort"
	"strconv"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/valency"
)

// CompiledVariant is a validated immutable sequence plus boundary tokens for
// a later bounded parser. Sequence order never modifies semantic role data.
type CompiledVariant struct {
	Variant      Variant
	FamilyID     ids.ConstructionID
	PlanID       ConstructionPlanID
	FirstAllowed []SequenceToken
	LastAllowed  []SequenceToken
}

// CompileVariant validates a variant against one family and computes every
// possible first/last token after optional boundary slots are omitted.
func CompileVariant(variant Variant, family ConstructionFamily) (CompiledVariant, error) {
	if err := family.Validate(); err != nil {
		return CompiledVariant{}, err
	}
	if err := variant.Validate(family); err != nil {
		return CompiledVariant{}, err
	}
	slots := make(map[string]SlotSpec, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		slots[slot.ID] = slot
	}
	counts := make(map[string]int, len(slots))
	hasRequiredToken := false
	for index, token := range variant.Sequence {
		if token.Kind != TokenSlot {
			if err := token.validate("variant.sequence[" + strconv.Itoa(index) + "]"); err != nil {
				return CompiledVariant{}, err
			}
			if !token.Optional {
				hasRequiredToken = true
			}
			continue
		}
		counts[token.SlotID]++
		if counts[token.SlotID] > 1 {
			return CompiledVariant{}, invalid("DUPLICATE_VARIANT_SLOT", "variant.sequence["+strconv.Itoa(index)+"].slot_id", "a family slot may occur only once in a variant sequence")
		}
		if slots[token.SlotID].Optionality == SlotRequired {
			hasRequiredToken = true
		}
	}
	for _, slot := range family.SlotSpecs {
		count := counts[slot.ID]
		if slot.Optionality == SlotRequired && count != 1 {
			return CompiledVariant{}, invalid("MISSING_REQUIRED_VARIANT_SLOT", "variant.sequence", "every required family slot must occur exactly once")
		}
		if slot.Optionality == SlotOptional && count > 1 {
			return CompiledVariant{}, invalid("DUPLICATE_VARIANT_SLOT", "variant.sequence", "optional family slot may occur at most once")
		}
	}
	if !hasRequiredToken {
		return CompiledVariant{}, invalid("EMPTY_POSSIBLE_VARIANT", "variant.sequence", "variant may omit every token and produce an empty sequence")
	}
	first := boundaryTokens(variant.Sequence, slots, false)
	last := boundaryTokens(variant.Sequence, slots, true)
	return CompiledVariant{
		Variant:      cloneVariant(variant),
		FamilyID:     family.ID,
		PlanID:       planID(family),
		FirstAllowed: first,
		LastAllowed:  last,
	}, nil
}

// ValidateVariantForFrame checks that a sequence remains compatible with one
// sense-specific frame. A generic family/variant pair can be checked against
// each approved frame binding without copying either structure.
func ValidateVariantForFrame(variant Variant, family ConstructionFamily, frame valency.ValencyFrame) (CompiledVariant, error) {
	compiled, err := CompileVariant(variant, family)
	if err != nil {
		return CompiledVariant{}, err
	}
	if err := frame.Validate(); err != nil {
		return CompiledVariant{}, err
	}
	if !familyAllowsFrame(family, frame) {
		return CompiledVariant{}, unsupported("FRAME_FAMILY_NOT_LICENSED", "frame.allowed_family_ids", "valency frame does not license this construction family")
	}
	familySlots := make(map[string]SlotSpec, len(family.SlotSpecs))
	roleSlots := make(map[string]SlotSpec)
	for _, slot := range family.SlotSpecs {
		familySlots[slot.ID] = slot
		if slot.ArgumentBinding != nil {
			roleSlots[string(slot.ArgumentBinding.SemanticRole)] = slot
		}
	}
	frameArguments := make(map[string]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		frameArguments[string(argument.Role)] = argument
	}
	sequenceSlots := make(map[string]struct{}, len(variant.Sequence))
	for _, token := range variant.Sequence {
		if token.Kind == TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	for role := range roleSlots {
		if _, exists := frameArguments[role]; !exists {
			return CompiledVariant{}, unsupported("FAMILY_ROLE_NOT_LICENSED_BY_FRAME", "frame.arguments", "family binds a semantic role not licensed by this sense frame")
		}
	}
	for role, argument := range frameArguments {
		if !argument.Required {
			continue
		}
		slot, exists := roleSlots[role]
		if !exists {
			return CompiledVariant{}, unsupported("REQUIRED_FRAME_ROLE_UNBOUND", "family.slot_specs", "required frame role has no family argument slot")
		}
		if _, exists := sequenceSlots[slot.ID]; !exists {
			if _, licensed := ImplicitParticipantSlot(family, variant, slot.ID); licensed {
				continue
			}
			return CompiledVariant{}, unsupported("REQUIRED_FRAME_ROLE_OMITTED", "variant.sequence", "variant omits a slot bound to a required frame role")
		}
	}
	return compiled, nil
}

func familyAllowsFrame(family ConstructionFamily, frame valency.ValencyFrame) bool {
	for _, familyID := range frame.AllowedFamilyIDs {
		if familyID == family.ID {
			return true
		}
	}
	return false
}

func boundaryTokens(sequence []SequenceToken, slots map[string]SlotSpec, reverse bool) []SequenceToken {
	result := []SequenceToken{}
	for offset := 0; offset < len(sequence); offset++ {
		index := offset
		if reverse {
			index = len(sequence) - 1 - offset
		}
		token := sequence[index]
		result = append(result, CloneSequenceTokens([]SequenceToken{token})[0])
		if token.Kind == TokenTerminal && !token.Optional || token.Kind == TokenSlot && slots[token.SlotID].Optionality == SlotRequired {
			break
		}
	}
	if reverse {
		reverseTokens(result)
	}
	return result
}

func reverseTokens(tokens []SequenceToken) {
	for left, right := 0, len(tokens)-1; left < right; left, right = left+1, right-1 {
		tokens[left], tokens[right] = tokens[right], tokens[left]
	}
}

// VariantIndex is an immutable family-local index by variant ID and possible
// boundary token. It preserves all variants that share a boundary.
type VariantIndex struct {
	family  ConstructionFamily
	records []CompiledVariant
	byID    map[string]int
	byFirst map[string][]int
	byLast  map[string][]int
}

// CompileVariantIndex validates and indexes variants for one family. Sense
// compatibility is checked at each binding with ValidateVariantForFrame.
func CompileVariantIndex(family ConstructionFamily, variants []Variant) (*VariantIndex, error) {
	if err := family.Validate(); err != nil {
		return nil, err
	}
	if len(variants) == 0 {
		return nil, invalid("EMPTY_VARIANT_INDEX", "variants", "variant index requires at least one variant")
	}
	if len(variants) > MaxFamilyItems {
		return nil, resourceLimit("variants", len(variants), MaxFamilyItems)
	}
	records := make([]CompiledVariant, len(variants))
	byID := make(map[string]int, len(variants))
	for index, variant := range variants {
		compiled, err := CompileVariant(variant, family)
		if err != nil {
			return nil, err
		}
		key := variant.ID.String()
		if _, exists := byID[key]; exists {
			return nil, invalid("DUPLICATE_VARIANT_ID", "variants", "variant ID is duplicated within the family")
		}
		records[index] = compiled
		byID[key] = index
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Variant.ID.String() < records[j].Variant.ID.String() })
	byID = make(map[string]int, len(records))
	byFirst := make(map[string][]int)
	byLast := make(map[string][]int)
	for index, record := range records {
		byID[record.Variant.ID.String()] = index
		seenFirst := make(map[string]struct{}, len(record.FirstAllowed))
		for _, token := range record.FirstAllowed {
			keys, _ := token.InputKeys()
			for _, key := range keys {
				if _, exists := seenFirst[key]; exists {
					continue
				}
				seenFirst[key] = struct{}{}
				byFirst[key] = append(byFirst[key], index)
			}
		}
		seenLast := make(map[string]struct{}, len(record.LastAllowed))
		for _, token := range record.LastAllowed {
			keys, _ := token.InputKeys()
			for _, key := range keys {
				if _, exists := seenLast[key]; exists {
					continue
				}
				seenLast[key] = struct{}{}
				byLast[key] = append(byLast[key], index)
			}
		}
	}
	return &VariantIndex{family: canonicalFamily(family), records: records, byID: byID, byFirst: byFirst, byLast: byLast}, nil
}

// ByID returns one compiled variant from this family.
func (index *VariantIndex) ByID(variantID VariantID) (CompiledVariant, bool) {
	if index == nil || !validVariantID(variantID) {
		return CompiledVariant{}, false
	}
	position, exists := index.byID[variantID.String()]
	if !exists {
		return CompiledVariant{}, false
	}
	return cloneCompiled(index.records[position]), true
}

// ByFirstToken returns every compiled variant that may begin with token.
func (index *VariantIndex) ByFirstToken(token SequenceToken) ([]CompiledVariant, error) {
	if index == nil {
		return []CompiledVariant{}, unknown("VARIANT_INDEX_UNAVAILABLE", "index", "variant index is unavailable")
	}
	key, err := token.Key()
	if err != nil {
		return []CompiledVariant{}, err
	}
	return index.recordsAt(index.byFirst[key]), nil
}

// ByLastToken returns every compiled variant that may end with token.
func (index *VariantIndex) ByLastToken(token SequenceToken) ([]CompiledVariant, error) {
	if index == nil {
		return []CompiledVariant{}, unknown("VARIANT_INDEX_UNAVAILABLE", "index", "variant index is unavailable")
	}
	key, err := token.Key()
	if err != nil {
		return []CompiledVariant{}, err
	}
	return index.recordsAt(index.byLast[key]), nil
}

func (index *VariantIndex) recordsAt(positions []int) []CompiledVariant {
	result := make([]CompiledVariant, len(positions))
	for offset, position := range positions {
		result[offset] = cloneCompiled(index.records[position])
	}
	return result
}

func cloneCompiled(record CompiledVariant) CompiledVariant {
	return CompiledVariant{
		Variant:      cloneVariant(record.Variant),
		FamilyID:     record.FamilyID,
		PlanID:       record.PlanID,
		FirstAllowed: CloneSequenceTokens(record.FirstAllowed),
		LastAllowed:  CloneSequenceTokens(record.LastAllowed),
	}
}

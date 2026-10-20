package construction

import (
	"sort"
	"strconv"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/valency"
)

// FamilyRecord pairs one immutable family template with its deterministic
// structural plan identity.
type FamilyRecord struct {
	Family ConstructionFamily
	PlanID ConstructionPlanID
}

// Registry is an immutable index over language-local construction templates
// and their separately stored approved sense bindings.
type Registry struct {
	families         []FamilyRecord
	bindingsByFamily map[string][]valency.ConstructionBinding
	byID             map[string]int
	byLanguage       map[string][]int
	byRoleLayout     map[string][]int
}

// CompileRegistry validates each unique family and links the already-validated
// approved bindings from a valency index. One family is stored once regardless
// of how many senses refer to it.
func CompileRegistry(families []ConstructionFamily, valencyIndex *valency.Index) (*Registry, error) {
	if len(families) == 0 {
		return nil, invalid("EMPTY_CONSTRUCTION_REGISTRY", "families", "registry requires at least one family")
	}
	if len(families) > MaxFamilies {
		return nil, resourceLimit("families", len(families), MaxFamilies)
	}
	records := make([]FamilyRecord, len(families))
	byID := make(map[string]int, len(families))
	byLanguage := make(map[string][]int)
	byRoleLayout := make(map[string][]int)
	itemCount := 0
	for index, family := range families {
		if err := family.Validate(); err != nil {
			return nil, err
		}
		itemCount += len(family.SlotSpecs) + len(family.Constraints) + len(family.Operators) + len(family.Capabilities)
		for _, constraint := range family.Constraints {
			itemCount += len(constraint.Features)
		}
		if itemCount > MaxFamilyItems {
			return nil, resourceLimit("registry_items", itemCount, MaxFamilyItems)
		}
		key := family.ID.String()
		if _, exists := byID[key]; exists {
			return nil, invalid("DUPLICATE_CONSTRUCTION_FAMILY_ID", "families", "family ID is duplicated")
		}
		records[index] = FamilyRecord{Family: canonicalFamily(family), PlanID: planID(family)}
		byID[key] = index
		byLanguage[string(family.Language)] = append(byLanguage[string(family.Language)], index)
		layoutKey := roleLayoutKey(family.SlotSpecs)
		byRoleLayout[layoutKey] = append(byRoleLayout[layoutKey], index)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Family.ID.String() < records[j].Family.ID.String() })
	// Rebuild index positions over the sorted, immutable backing slice.
	byID = make(map[string]int, len(records))
	byLanguage = make(map[string][]int)
	byRoleLayout = make(map[string][]int)
	for index, record := range records {
		byID[record.Family.ID.String()] = index
		byLanguage[string(record.Family.Language)] = append(byLanguage[string(record.Family.Language)], index)
		key := roleLayoutKey(record.Family.SlotSpecs)
		byRoleLayout[key] = append(byRoleLayout[key], index)
	}
	bindingsByFamily := make(map[string][]valency.ConstructionBinding)
	if valencyIndex != nil {
		for index, binding := range valencyIndex.Bindings() {
			familyKey := binding.FamilyID.String()
			if _, exists := byID[familyKey]; !exists {
				return nil, unknown("UNKNOWN_CONSTRUCTION_FAMILY", "bindings["+strconv.Itoa(index)+"].family_id", "approved valency binding refers to no family in this registry")
			}
			bindingsByFamily[familyKey] = append(bindingsByFamily[familyKey], binding)
		}
	}
	for familyID := range bindingsByFamily {
		sort.Slice(bindingsByFamily[familyID], func(i, j int) bool {
			left, right := bindingsByFamily[familyID][i], bindingsByFamily[familyID][j]
			if left.SenseID.String() != right.SenseID.String() {
				return left.SenseID.String() < right.SenseID.String()
			}
			if left.ValencyFrameID != right.ValencyFrameID {
				return left.ValencyFrameID < right.ValencyFrameID
			}
			return left.VariantID.String() < right.VariantID.String()
		})
	}
	return &Registry{families: records, bindingsByFamily: bindingsByFamily, byID: byID, byLanguage: byLanguage, byRoleLayout: byRoleLayout}, nil
}

// FamilyFor performs a type-safe lookup by language and family ID. It never
// falls back to a ConceptID or a family from another language.
func (registry *Registry) FamilyFor(language features.LanguageID, familyID ids.ConstructionID) (FamilyRecord, error) {
	if registry == nil {
		return FamilyRecord{}, unknown("CONSTRUCTION_REGISTRY_UNAVAILABLE", "registry", "construction registry is unavailable")
	}
	if !validLanguage(language) || !validFamilyID(familyID) {
		return FamilyRecord{}, invalid("INVALID_FAMILY_QUERY", "query", "family query requires canonical language and construction family IDs")
	}
	position, exists := registry.byID[familyID.String()]
	if !exists {
		return FamilyRecord{}, unknown("UNKNOWN_CONSTRUCTION_FAMILY", "family_id", "construction family is not registered")
	}
	record := registry.families[position]
	if record.Family.Language != language {
		return FamilyRecord{}, unsupported("CONSTRUCTION_LANGUAGE_MISMATCH", "language", "construction family is registered for a different language")
	}
	return cloneRecord(record), nil
}

// ByID returns a registered family by stable typed ID, independent of language.
func (registry *Registry) ByID(familyID ids.ConstructionID) (FamilyRecord, bool) {
	if registry == nil || !validFamilyID(familyID) {
		return FamilyRecord{}, false
	}
	position, exists := registry.byID[familyID.String()]
	if !exists {
		return FamilyRecord{}, false
	}
	return cloneRecord(registry.families[position]), true
}

// ByLanguage returns every family for one language in stable family ID order.
func (registry *Registry) ByLanguage(language features.LanguageID) []FamilyRecord {
	if registry == nil || !validLanguage(language) {
		return []FamilyRecord{}
	}
	positions := registry.byLanguage[string(language)]
	result := make([]FamilyRecord, len(positions))
	for offset, position := range positions {
		result[offset] = cloneRecord(registry.families[position])
	}
	return result
}

// ByRoleLayout returns all families with the supplied ordered slot structure.
// Language-specific constraints do not form part of the role layout, so PL and
// EN families can share a layout while retaining separate family records.
func (registry *Registry) ByRoleLayout(slots []SlotSpec) []FamilyRecord {
	if registry == nil || len(slots) == 0 || len(slots) > MaxSlotsPerFamily {
		return []FamilyRecord{}
	}
	key := roleLayoutKey(slots)
	positions := registry.byRoleLayout[key]
	result := make([]FamilyRecord, len(positions))
	for offset, position := range positions {
		result[offset] = cloneRecord(registry.families[position])
	}
	return result
}

// BindingsByFamilyID returns separately stored approved sense bindings. The
// family template is not cloned or specialized for each binding.
func (registry *Registry) BindingsByFamilyID(familyID ids.ConstructionID) []valency.ConstructionBinding {
	if registry == nil || !validFamilyID(familyID) {
		return []valency.ConstructionBinding{}
	}
	bindings := registry.bindingsByFamily[familyID.String()]
	return append([]valency.ConstructionBinding(nil), bindings...)
}

// Families returns all family records in stable ID order.
func (registry *Registry) Families() []FamilyRecord {
	if registry == nil {
		return []FamilyRecord{}
	}
	result := make([]FamilyRecord, len(registry.families))
	for index, record := range registry.families {
		result[index] = cloneRecord(record)
	}
	return result
}

func cloneRecord(record FamilyRecord) FamilyRecord {
	return FamilyRecord{Family: cloneFamily(record.Family), PlanID: record.PlanID}
}

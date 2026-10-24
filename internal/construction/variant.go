package construction

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	MaxSequenceTokens  = 256
	MaxConditions      = 128
	MaxTerminalBytes   = 256
	MaxTerminalInputs  = 16
	MaxVariantPriority = 1_000
)

// SequenceTokenKind selects exactly one variant of the slot/terminal union.
type SequenceTokenKind string

const (
	TokenSlot     SequenceTokenKind = "slot"
	TokenTerminal SequenceTokenKind = "terminal"
)

// VariantID is distinct from a family ConstructionID at the Go type level.
// Its external spelling uses the construction namespace shared by this graph.
type VariantID string

// NewVariantID creates a canonical typed variant identity.
func NewVariantID(name string) (VariantID, error) {
	id, err := ids.NewConstructionID(name)
	if err != nil {
		return "", err
	}
	return VariantID(id.String()), nil
}

// ParseVariantID parses the canonical external variant identity.
func ParseVariantID(value string) (VariantID, error) {
	id, err := ids.ParseConstructionID(value)
	if err != nil {
		return "", err
	}
	return VariantID(id.String()), nil
}

// String returns the stable external variant identity.
func (id VariantID) String() string { return string(id) }

// SequenceToken is either one family slot reference or one terminal. A
// terminal may accept bounded alternate source spellings and may be omitted
// during analysis; linearization always emits its canonical Terminal value.
// Repeated terminal text is allowed; a slot reference may occur only once.
type SequenceToken struct {
	Kind           SequenceTokenKind `json:"kind"`
	SlotID         string            `json:"slot_id,omitempty"`
	Terminal       string            `json:"terminal,omitempty"`
	Optional       bool              `json:"optional,omitempty"`
	AcceptedInputs []string          `json:"accepted_inputs,omitempty"`
}

// Key returns a bounded, unambiguous index key for a well-formed token.
func (token SequenceToken) Key() (string, error) {
	if err := token.validate("token"); err != nil {
		return "", err
	}
	value := token.SlotID
	if token.Kind == TokenTerminal {
		value = token.Terminal
	}
	return string(token.Kind) + "\x00" + value, nil
}

// InputKeys returns the canonical terminal key and each explicitly accepted
// source spelling. Slot tokens have one key. The returned slice is owned by
// the caller.
func (token SequenceToken) InputKeys() ([]string, error) {
	if err := token.validate("token"); err != nil {
		return nil, err
	}
	if token.Kind == TokenSlot {
		key, _ := token.Key()
		return []string{key}, nil
	}
	keys := make([]string, 0, len(token.AcceptedInputs)+1)
	keys = append(keys, string(TokenTerminal)+"\x00"+token.Terminal)
	for _, accepted := range token.AcceptedInputs {
		keys = append(keys, string(TokenTerminal)+"\x00"+accepted)
	}
	return keys, nil
}

// MatchesInput reports whether an analyzed terminal is licensed by this
// token, including an explicitly declared alternate source spelling.
func (token SequenceToken) MatchesInput(value string) bool {
	if token.Kind != TokenTerminal {
		return false
	}
	if value == token.Terminal {
		return true
	}
	for _, accepted := range token.AcceptedInputs {
		if value == accepted {
			return true
		}
	}
	return false
}

func (token SequenceToken) validate(path string) error {
	switch token.Kind {
	case TokenSlot:
		if !validToken(token.SlotID) || token.Terminal != "" || token.Optional || len(token.AcceptedInputs) != 0 {
			return invalid("INVALID_SLOT_SEQUENCE_TOKEN", path, "slot token requires only a bounded slot ID")
		}
	case TokenTerminal:
		if token.SlotID != "" || !validTerminal(token.Terminal) {
			return invalid("INVALID_TERMINAL_SEQUENCE_TOKEN", path, "terminal token requires only a bounded non-empty terminal")
		}
		if len(token.AcceptedInputs) > MaxTerminalInputs {
			return resourceLimit(path+".accepted_inputs", len(token.AcceptedInputs), MaxTerminalInputs)
		}
		seen := map[string]struct{}{token.Terminal: {}}
		for index, accepted := range token.AcceptedInputs {
			if !validTerminal(accepted) {
				return invalid("INVALID_ACCEPTED_TERMINAL", path+".accepted_inputs["+strconv.Itoa(index)+"]", "accepted source terminal must be a bounded non-empty terminal")
			}
			if _, duplicate := seen[accepted]; duplicate {
				return invalid("DUPLICATE_ACCEPTED_TERMINAL", path+".accepted_inputs", "accepted source terminals must be unique and differ from the canonical terminal")
			}
			seen[accepted] = struct{}{}
		}
	default:
		return invalid("INVALID_SEQUENCE_TOKEN_KIND", path+".kind", "sequence token kind must be slot or terminal")
	}
	return nil
}

func validTerminal(value string) bool {
	if value == "" || len(value) > MaxTerminalBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// Markedness describes the conventional status of a licensed variant. It does
// not imply topic, focus, truth, or an inferred discourse fact.
type Markedness string

const (
	MarkednessNeutral Markedness = "neutral"
	MarkednessMarked  Markedness = "marked"
)

// DirectionPolicy states which direction may consult this variant.
type DirectionPolicy string

const (
	DirectionBoth    DirectionPolicy = "both"
	DirectionRealize DirectionPolicy = "realize"
	DirectionAnalyze DirectionPolicy = "analyze"
)

// ConditionKind is a closed set of predicates whose evidence must be supplied
// explicitly by a later planner.
type ConditionKind string

const (
	ConditionFeatureEquals        ConditionKind = "feature_equals"
	ConditionInformationStructure ConditionKind = "information_structure"
)

// InformationValue is an explicitly represented discourse status.
type InformationValue string

const (
	InformationTopic InformationValue = "topic"
	InformationFocus InformationValue = "focus"
	InformationGiven InformationValue = "given"
	InformationNew   InformationValue = "new"
)

// Condition is a closed typed predicate; no field contains executable code.
type Condition struct {
	Kind    ConditionKind           `json:"kind"`
	SlotID  string                  `json:"slot_id,omitempty"`
	Feature features.FeatureID      `json:"feature,omitempty"`
	Value   features.FeatureValueID `json:"value,omitempty"`
	Role    meaning.Role            `json:"role,omitempty"`
	Info    InformationValue        `json:"information_structure,omitempty"`
}

// Variant gives one family a linearization order and explicit licensing data.
// It cannot change semantic role assignments.
type Variant struct {
	ID              VariantID            `json:"id"`
	FamilyID        ids.ConstructionID   `json:"family_id"`
	Sequence        []SequenceToken      `json:"sequence"`
	Conditions      []Condition          `json:"conditions,omitempty"`
	Markedness      Markedness           `json:"markedness"`
	Priority        int                  `json:"priority"`
	DirectionPolicy DirectionPolicy      `json:"direction_policy"`
	Provenance      provenance.SourceRef `json:"provenance"`
}

// VariantEmitsAdposition reports whether the exact selected sequence contains
// the required adposition after its predicate and before the licensed argument
// slot. This keeps PP surface material explicit in the language-local variant.
func VariantEmitsAdposition(variant Variant, family ConstructionFamily, slotID, adposition string) bool {
	if slotID == "" || !validToken(adposition) {
		return false
	}
	var predicatePosition, argumentPosition = -1, -1
	for position, token := range variant.Sequence {
		if token.Kind != TokenSlot {
			continue
		}
		for _, slot := range family.SlotSpecs {
			if slot.ID != token.SlotID {
				continue
			}
			if slot.IsMainPredicate() && predicatePosition < 0 {
				predicatePosition = position
			}
			if token.SlotID == slotID {
				argumentPosition = position
			}
			break
		}
	}
	if predicatePosition < 0 || argumentPosition <= predicatePosition {
		return false
	}
	for position := predicatePosition + 1; position < argumentPosition; position++ {
		token := variant.Sequence[position]
		if token.Kind == TokenTerminal && token.Terminal == adposition {
			return true
		}
	}
	return false
}

// ConstructionVariant is the descriptive name for Variant in external APIs.
type ConstructionVariant = Variant

func (variant Variant) Validate(family ConstructionFamily) error {
	if !validVariantID(variant.ID) || !validFamilyID(variant.FamilyID) {
		return invalid("INVALID_VARIANT_ID", "variant.id", "variant and family IDs must be canonical construction IDs")
	}
	if variant.FamilyID != family.ID {
		return invalid("VARIANT_FAMILY_IDENTITY_MISMATCH", "variant.family_id", "variant must reference the exact family being compiled")
	}
	if len(variant.Sequence) == 0 || len(variant.Sequence) > MaxSequenceTokens {
		if len(variant.Sequence) > MaxSequenceTokens {
			return resourceLimit("sequence_tokens", len(variant.Sequence), MaxSequenceTokens)
		}
		return invalid("EMPTY_VARIANT_SEQUENCE", "variant.sequence", "variant requires at least one sequence token")
	}
	if len(variant.Conditions) > MaxConditions {
		return resourceLimit("conditions", len(variant.Conditions), MaxConditions)
	}
	if variant.Markedness != MarkednessNeutral && variant.Markedness != MarkednessMarked {
		return invalid("INVALID_VARIANT_MARKEDNESS", "variant.markedness", "markedness must be neutral or marked")
	}
	if variant.Priority < 0 || variant.Priority > MaxVariantPriority {
		return invalid("INVALID_VARIANT_PRIORITY", "variant.priority", "priority is outside the supported range")
	}
	switch variant.DirectionPolicy {
	case DirectionBoth, DirectionRealize, DirectionAnalyze:
	default:
		return invalid("INVALID_DIRECTION_POLICY", "variant.direction_policy", "direction policy must be both, realize, or analyze")
	}
	if err := variant.Provenance.Validate(); err != nil {
		return invalid("INVALID_VARIANT_PROVENANCE", "variant.provenance", "variant requires valid provenance")
	}
	slots := make(map[string]SlotSpec, len(family.SlotSpecs))
	roles := make(map[meaning.Role]struct{})
	for _, slot := range family.SlotSpecs {
		slots[slot.ID] = slot
		if slot.ArgumentBinding != nil {
			roles[slot.ArgumentBinding.SemanticRole] = struct{}{}
		}
	}
	for index, token := range variant.Sequence {
		if err := token.validate("variant.sequence[" + strconv.Itoa(index) + "]"); err != nil {
			return err
		}
		if token.Kind == TokenSlot {
			if _, exists := slots[token.SlotID]; !exists {
				return unknown("UNKNOWN_VARIANT_SLOT", "variant.sequence["+strconv.Itoa(index)+"].slot_id", "sequence refers to no slot in the family")
			}
		}
	}
	for index, condition := range variant.Conditions {
		path := "variant.conditions[" + strconv.Itoa(index) + "]"
		switch condition.Kind {
		case ConditionFeatureEquals:
			if _, exists := slots[condition.SlotID]; !exists || !validToken(string(condition.Feature)) || !validToken(string(condition.Value)) || condition.Role != "" || condition.Info != "" {
				return invalid("INVALID_FEATURE_CONDITION", path, "feature condition requires a family slot, feature and value only")
			}
		case ConditionInformationStructure:
			if _, exists := roles[condition.Role]; !exists || condition.SlotID != "" || condition.Feature != "" || condition.Value != "" || !validInformationValue(condition.Info) {
				return invalid("INVALID_INFORMATION_STRUCTURE_CONDITION", path, "information-structure condition requires a bound semantic role and explicit value only")
			}
		default:
			return invalid("INVALID_VARIANT_CONDITION_KIND", path+".kind", "condition kind is not supported")
		}
	}
	seenConditions := make(map[string]struct{}, len(variant.Conditions))
	for _, condition := range variant.Conditions {
		key := string(condition.Kind) + "\x00" + condition.SlotID + "\x00" + string(condition.Feature) + "\x00" + string(condition.Value) + "\x00" + string(condition.Role) + "\x00" + string(condition.Info)
		if _, exists := seenConditions[key]; exists {
			return invalid("DUPLICATE_VARIANT_CONDITION", "variant.conditions", "condition is duplicated")
		}
		seenConditions[key] = struct{}{}
	}
	return nil
}

func validVariantID(id VariantID) bool {
	parsed, err := ParseVariantID(string(id))
	return err == nil && parsed == id
}

func validInformationValue(value InformationValue) bool {
	switch value {
	case InformationTopic, InformationFocus, InformationGiven, InformationNew:
		return true
	default:
		return false
	}
}

func cloneVariant(variant Variant) Variant {
	result := variant
	result.Sequence = CloneSequenceTokens(variant.Sequence)
	result.Conditions = append([]Condition(nil), variant.Conditions...)
	return result
}

// CloneSequenceTokens returns a deep copy of a sequence and its alternate
// terminal spellings.
func CloneSequenceTokens(tokens []SequenceToken) []SequenceToken {
	result := append([]SequenceToken(nil), tokens...)
	for index := range result {
		result[index].AcceptedInputs = append([]string(nil), tokens[index].AcceptedInputs...)
	}
	return result
}

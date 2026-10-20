package construction

import (
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	MaxPhrasePlans = 100_000
	MaxPhraseItems = 1_000_000
)

type PhraseID string
type PhraseKind string
type PhraseInputKind string

const (
	PhraseNP           PhraseKind      = "np"
	PhraseVP           PhraseKind      = "vp"
	PhrasePP           PhraseKind      = "pp"
	PhraseInputLexical PhraseInputKind = "lexical"
	PhraseInputNP      PhraseInputKind = "np"
	PhraseInputVP      PhraseInputKind = "vp"
	PhraseInputPP      PhraseInputKind = "pp"
)

// PhraseInput names a required or optional runtime input. IDs are placeholders,
// never lexical content or a claim that a corresponding lexical item exists.
type PhraseInput struct {
	ID       string          `json:"id"`
	Role     string          `json:"role"`
	Kind     PhraseInputKind `json:"kind"`
	Required bool            `json:"required"`
}

type PhraseChild struct {
	Role   string   `json:"role"`
	PlanID PhraseID `json:"plan_id"`
}

type FeatureAgreement struct {
	ControllerInput string             `json:"controller_input"`
	TargetInput     string             `json:"target_input"`
	Feature         features.FeatureID `json:"feature"`
}

// PhrasePlan is one node in an immutable phrase-plan DAG. Kind-specific fields
// are checked by CompileNP, CompileVP, and CompilePP.
type PhrasePlan struct {
	ID           PhraseID               `json:"id"`
	TemplateID   string                 `json:"template_id"`
	Language     features.LanguageID    `json:"language"`
	Kind         PhraseKind             `json:"kind"`
	Inputs       []PhraseInput          `json:"inputs,omitempty"`
	Children     []PhraseChild          `json:"children,omitempty"`
	Features     features.FeatureBundle `json:"features,omitempty"`
	Agreements   []FeatureAgreement     `json:"agreements,omitempty"`
	FrameID      valency.FrameID        `json:"frame_id,omitempty"`
	GovernedCase features.FeatureValue  `json:"governed_case,omitempty"`
}

type CompiledPhrasePlan struct {
	Plan PhrasePlan
}

type PhraseIndex struct {
	plans []CompiledPhrasePlan
	byID  map[PhraseID]int
}

type PhraseTemplateInput struct {
	Role     string          `json:"role"`
	Kind     PhraseInputKind `json:"kind"`
	Required bool            `json:"required"`
}

type PhraseTemplateChild struct {
	Role     string     `json:"role"`
	Kind     PhraseKind `json:"kind"`
	Required bool       `json:"required"`
}

type PhraseTemplate struct {
	ID         string                `json:"id"`
	Kind       PhraseKind            `json:"kind"`
	Inputs     []PhraseTemplateInput `json:"inputs"`
	Children   []PhraseTemplateChild `json:"children"`
	FeatureIDs []features.FeatureID  `json:"feature_ids,omitempty"`
}

type PhraseProfile struct {
	SchemaVersion int                 `json:"schema_version"`
	Language      features.LanguageID `json:"language"`
	Version       string              `json:"version"`
	Templates     []PhraseTemplate    `json:"templates"`
}

func CompilePhraseGraph(plans []PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile, frames map[valency.FrameID]valency.ValencyFrame) (*PhraseIndex, error) {
	if len(plans) == 0 {
		return nil, invalid("EMPTY_PHRASE_GRAPH", "plans", "phrase graph requires at least one plan")
	}
	if len(plans) > MaxPhrasePlans {
		return nil, resourceLimit("phrase_plans", len(plans), MaxPhrasePlans)
	}
	if profile == nil {
		return nil, unknown("PHRASE_PROFILE_UNAVAILABLE", "profile", "phrase compilation requires an explicit language feature profile")
	}
	if phraseProfile.Language != profile.Language() {
		return nil, unsupported("PHRASE_PROFILE_LANGUAGE_MISMATCH", "profile.language", "phrase and feature profiles must use the same language")
	}
	if err := ValidatePhraseProfile(phraseProfile, profile); err != nil {
		return nil, err
	}
	byID := make(map[PhraseID]int, len(plans))
	compiled := make([]CompiledPhrasePlan, len(plans))
	items := 0
	for i, plan := range plans {
		items += len(plan.Inputs) + len(plan.Children) + len(plan.Features.Values) + len(plan.Agreements)
		if items > MaxPhraseItems {
			return nil, resourceLimit("phrase_items", items, MaxPhraseItems)
		}
		if _, exists := byID[plan.ID]; exists {
			return nil, invalid("DUPLICATE_PHRASE_ID", "plans", "phrase plan ID is duplicated")
		}
		var result CompiledPhrasePlan
		var err error
		switch plan.Kind {
		case PhraseNP:
			result, err = compilePhraseNode(plan, phraseProfile, profile, nil)
		case PhraseVP:
			result, err = compilePhraseNode(plan, phraseProfile, profile, frames)
		case PhrasePP:
			result, err = compilePhraseNode(plan, phraseProfile, profile, nil)
		default:
			err = invalid("INVALID_PHRASE_KIND", "kind", "phrase kind must be np, vp, or pp")
		}
		if err != nil {
			return nil, err
		}
		byID[plan.ID] = i
		compiled[i] = result
	}
	for i, plan := range plans {
		seenRoles := map[string]struct{}{}
		for _, child := range plan.Children {
			position, exists := byID[child.PlanID]
			if !exists {
				return nil, unknown("UNKNOWN_PHRASE_CHILD", "plans["+strconv.Itoa(i)+"].children", "child plan reference does not exist")
			}
			if plans[position].Language != plan.Language {
				return nil, unsupported("PHRASE_LANGUAGE_MISMATCH", "children", "nested phrase plans must use the same language")
			}
			if _, exists := seenRoles[child.Role]; exists && !(plan.Kind == PhraseNP && child.Role == "modifier") {
				return nil, invalid("DUPLICATE_PHRASE_CHILD_ROLE", "children", "child role is duplicated")
			}
			seenRoles[child.Role] = struct{}{}
			if !allowedPhraseChild(plan.Kind, child.Role, plans[position].Kind) {
				return nil, unsupported("PHRASE_CHILD_KIND_UNSUPPORTED", "children", "child kind is not allowed for this phrase role")
			}
			for _, template := range phraseProfile.Templates {
				if template.ID == plan.TemplateID {
					for _, childSpec := range template.Children {
						if childSpec.Role == child.Role && childSpec.Kind != plans[position].Kind {
							return nil, unsupported("PHRASE_CHILD_TEMPLATE_KIND_MISMATCH", "children."+child.Role, "child kind does not match the selected phrase template")
						}
					}
				}
			}
		}
	}
	state := make([]uint8, len(plans))
	var visit func(int) error
	visit = func(position int) error {
		if state[position] == 1 {
			return invalid("PHRASE_CYCLE", "plans", "phrase graph must be acyclic")
		}
		if state[position] == 2 {
			return nil
		}
		state[position] = 1
		for _, child := range plans[position].Children {
			if err := visit(byID[child.PlanID]); err != nil {
				return err
			}
		}
		state[position] = 2
		return nil
	}
	for i := range plans {
		if err := visit(i); err != nil {
			return nil, err
		}
	}
	sort.Slice(compiled, func(i, j int) bool { return compiled[i].Plan.ID < compiled[j].Plan.ID })
	byID = make(map[PhraseID]int, len(compiled))
	for i, plan := range compiled {
		byID[plan.Plan.ID] = i
	}
	return &PhraseIndex{plans: compiled, byID: byID}, nil
}

func (index *PhraseIndex) ByID(id PhraseID) (CompiledPhrasePlan, bool) {
	if index == nil {
		return CompiledPhrasePlan{}, false
	}
	i, ok := index.byID[id]
	if !ok {
		return CompiledPhrasePlan{}, false
	}
	result := index.plans[i]
	result.Plan = clonePhrasePlan(result.Plan)
	return result, true
}

func (index *PhraseIndex) Plans() []CompiledPhrasePlan {
	if index == nil {
		return []CompiledPhrasePlan{}
	}
	result := make([]CompiledPhrasePlan, len(index.plans))
	for i, plan := range index.plans {
		result[i] = CompiledPhrasePlan{Plan: clonePhrasePlan(plan.Plan)}
	}
	return result
}

func compilePhraseNode(plan PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile, frames map[valency.FrameID]valency.ValencyFrame) (CompiledPhrasePlan, error) {
	if profile == nil {
		return CompiledPhrasePlan{}, unknown("PHRASE_PROFILE_UNAVAILABLE", "profile", "phrase compilation requires an explicit language feature profile")
	}
	if !validToken(string(plan.ID)) || !validToken(plan.TemplateID) || plan.Language != profile.Language() || plan.Language != phraseProfile.Language {
		return CompiledPhrasePlan{}, unsupported("PHRASE_PROFILE_MISMATCH", "language", "phrase requires a canonical ID and matching language profile")
	}
	var template *PhraseTemplate
	for i := range phraseProfile.Templates {
		if phraseProfile.Templates[i].ID == plan.TemplateID {
			template = &phraseProfile.Templates[i]
			break
		}
	}
	if template == nil || template.Kind != plan.Kind {
		return CompiledPhrasePlan{}, unknown("UNKNOWN_PHRASE_TEMPLATE", "template_id", "phrase template is absent or has a different kind")
	}
	if len(plan.Inputs) > MaxFamilyItems || len(plan.Children) > MaxFamilyItems {
		return CompiledPhrasePlan{}, resourceLimit("phrase_node_items", len(plan.Inputs)+len(plan.Children), MaxFamilyItems)
	}
	inputs := map[string]PhraseInput{}
	inputIDs := map[string]struct{}{}
	templateInputs := map[string]PhraseTemplateInput{}
	for _, input := range template.Inputs {
		templateInputs[input.Role] = input
	}
	for _, input := range plan.Inputs {
		if !validToken(input.ID) || !validToken(input.Role) {
			return CompiledPhrasePlan{}, invalid("INVALID_PHRASE_INPUT", "inputs", "phrase input requires bounded ID and role")
		}
		if input.Kind != PhraseInputLexical && input.Kind != PhraseInputNP && input.Kind != PhraseInputVP && input.Kind != PhraseInputPP {
			return CompiledPhrasePlan{}, invalid("INVALID_PHRASE_INPUT_KIND", "inputs", "phrase input kind is unsupported")
		}
		if _, exists := inputs[input.Role]; exists {
			return CompiledPhrasePlan{}, invalid("DUPLICATE_PHRASE_INPUT_ROLE", "inputs", "phrase input role is duplicated")
		}
		spec, exists := templateInputs[input.Role]
		if !exists || spec.Kind != input.Kind || spec.Required != input.Required {
			return CompiledPhrasePlan{}, unsupported("PHRASE_INPUT_NOT_IN_TEMPLATE", "inputs."+input.Role, "phrase input must match the selected template role, kind, and requiredness")
		}
		if _, exists := inputIDs[input.ID]; exists {
			return CompiledPhrasePlan{}, invalid("DUPLICATE_PHRASE_INPUT_ID", "inputs", "phrase input ID is duplicated")
		}
		inputIDs[input.ID] = struct{}{}
		inputs[input.Role] = input
	}
	for role, spec := range templateInputs {
		if spec.Required {
			if input, ok := inputs[role]; !ok || !input.Required {
				return CompiledPhrasePlan{}, incomplete("MISSING_TEMPLATE_INPUT", "inputs."+role, "required template input is absent")
			}
		}
	}
	templateChildren := map[string]PhraseTemplateChild{}
	for _, child := range template.Children {
		templateChildren[child.Role] = child
	}
	for _, child := range plan.Children {
		if _, ok := templateChildren[child.Role]; !ok {
			if plan.Kind != PhraseVP || !strings.HasPrefix(child.Role, "argument_") {
				return CompiledPhrasePlan{}, unsupported("PHRASE_CHILD_NOT_IN_TEMPLATE", "children."+child.Role, "child role is not declared by the selected template")
			}
		}
	}
	for role, spec := range templateChildren {
		if spec.Required {
			found := false
			for _, child := range plan.Children {
				if child.Role == role {
					found = true
				}
			}
			if !found {
				return CompiledPhrasePlan{}, incomplete("MISSING_TEMPLATE_CHILD", "children."+role, "required template child is absent")
			}
		}
	}
	requiredInput := func(role string, kind PhraseInputKind) error {
		input, ok := inputs[role]
		if !ok || !input.Required || input.Kind != kind {
			return invalid("MISSING_REQUIRED_PHRASE_INPUT", "inputs."+role, "required typed phrase input is missing")
		}
		return nil
	}
	switch plan.Kind {
	case PhraseNP:
		if err := requiredInput("head", PhraseInputLexical); err != nil {
			return CompiledPhrasePlan{}, err
		}
		if err := allowedInputRoles(inputs, "head", "determiner", "modifier", "quantity"); err != nil {
			return CompiledPhrasePlan{}, err
		}
		for _, child := range plan.Children {
			if child.Role != "modifier" && child.Role != "quantity" {
				return CompiledPhrasePlan{}, invalid("INVALID_NP_CHILD_ROLE", "children", "NP child role must be modifier or quantity")
			}
		}
	case PhraseVP:
		if err := requiredInput("verb", PhraseInputLexical); err != nil {
			return CompiledPhrasePlan{}, err
		}
		if plan.FrameID == "" {
			return CompiledPhrasePlan{}, incomplete("MISSING_VP_VALENCY_FRAME", "frame_id", "VP plan requires an explicit sense-specific valency frame")
		}
		frame, ok := frames[plan.FrameID]
		if !ok {
			return CompiledPhrasePlan{}, unknown("UNKNOWN_VP_VALENCY_FRAME", "frame_id", "VP frame is not present in the supplied frame set")
		}
		if frame.ID != plan.FrameID {
			return CompiledPhrasePlan{}, invalid("VP_FRAME_INDEX_ID_MISMATCH", "frame_id", "frame index key must match the frame's internal ID")
		}
		if err := frame.Validate(); err != nil {
			return CompiledPhrasePlan{}, err
		}
		if _, exists := inputs["auxiliary"]; exists {
			return CompiledPhrasePlan{}, unsupported("AUXILIARY_LICENSE_UNAVAILABLE", "inputs.auxiliary", "no approved language provider or evidence currently licenses an auxiliary")
		}
		childRoles := map[string]struct{}{}
		for _, child := range plan.Children {
			if child.Role == "auxiliary" {
				return CompiledPhrasePlan{}, unsupported("AUXILIARY_LICENSE_UNAVAILABLE", "children.auxiliary", "no approved language provider or evidence currently licenses an auxiliary")
			}
			childRoles[child.Role] = struct{}{}
		}
		for _, argument := range frame.Arguments {
			role := "argument_" + string(argument.Role)
			input, exists := inputs[role]
			_, hasChild := childRoles[role]
			if exists && hasChild {
				return CompiledPhrasePlan{}, invalid("DUPLICATE_VP_ARGUMENT_PLAN", "inputs."+role, "an argument may use an input or nested plan, not both")
			}
			if exists && ((input.Kind != PhraseInputNP && input.Kind != PhraseInputPP) || input.Required != argument.Required) {
				return CompiledPhrasePlan{}, invalid("INVALID_VP_ARGUMENT_INPUT", "inputs."+role, "argument input kind and requiredness must match the valency frame")
			}
			if argument.Required && !exists && !hasChild {
				return CompiledPhrasePlan{}, unsupported("REQUIRED_VP_ARGUMENT_MISSING", "inputs."+role, "required frame role needs an explicit NP or PP input")
			}
		}
		allowedRoles := append([]string{"verb", "auxiliary"}, argumentInputRoles(frame)...)
		if err := allowedInputRoles(inputs, allowedRoles...); err != nil {
			return CompiledPhrasePlan{}, err
		}
		for _, child := range plan.Children {
			if child.Role != "auxiliary" && !hasFrameArgument(frame, child.Role) {
				return CompiledPhrasePlan{}, invalid("INVALID_VP_CHILD_ROLE", "children", "VP child role must be auxiliary")
			}
		}
	case PhrasePP:
		if err := requiredInput("preposition", PhraseInputLexical); err != nil {
			return CompiledPhrasePlan{}, err
		}
		if err := allowedInputRoles(inputs, "preposition"); err != nil {
			return CompiledPhrasePlan{}, err
		}
		found := false
		for _, child := range plan.Children {
			if child.Role == "complement" {
				found = true
			} else {
				return CompiledPhrasePlan{}, invalid("INVALID_PP_CHILD_ROLE", "children", "PP child role must be complement")
			}
		}
		if !found {
			return CompiledPhrasePlan{}, incomplete("MISSING_PP_COMPLEMENT", "children", "PP plan requires a nested complement plan")
		}
	default:
		return CompiledPhrasePlan{}, invalid("INVALID_PHRASE_KIND", "kind", "phrase kind must be np, vp, or pp")
	}
	if len(plan.Features.Values) > 0 {
		if plan.Features.Language != profile.Language() || plan.Features.ProfileVersion != profile.Version() {
			return CompiledPhrasePlan{}, unsupported("PHRASE_FEATURE_PROFILE_MISMATCH", "features", "feature bundle must use the exact phrase profile")
		}
		if err := features.ValidateBundle(profile, plan.Features); err != nil {
			return CompiledPhrasePlan{}, err
		}
		allowed := map[features.FeatureID]struct{}{}
		for _, id := range template.FeatureIDs {
			allowed[id] = struct{}{}
		}
		for _, value := range plan.Features.Values {
			if _, ok := allowed[value.Feature]; !ok {
				return CompiledPhrasePlan{}, unsupported("PHRASE_FEATURE_NOT_IN_TEMPLATE", "features."+string(value.Feature), "feature is not declared by the selected phrase template")
			}
		}
	}
	if plan.GovernedCase.Feature != "" {
		if plan.Kind != PhrasePP || plan.GovernedCase.Feature != "case" {
			return CompiledPhrasePlan{}, invalid("INVALID_GOVERNED_CASE", "governed_case", "governed case is only valid as the case feature on a PP")
		}
		bundle, err := features.NewBundle(profile, []features.FeatureValue{plan.GovernedCase})
		if err != nil {
			return CompiledPhrasePlan{}, err
		}
		if err := features.ValidateBundle(profile, bundle); err != nil {
			return CompiledPhrasePlan{}, err
		}
		if !containsFeatureID(template.FeatureIDs, plan.GovernedCase.Feature) {
			return CompiledPhrasePlan{}, unsupported("GOVERNED_CASE_NOT_IN_TEMPLATE", "governed_case", "case feature is not declared by the selected PP template")
		}
	} else if plan.Kind != PhrasePP && plan.GovernedCase.State != "" {
		return CompiledPhrasePlan{}, invalid("UNEXPECTED_GOVERNED_CASE", "governed_case", "governed case field is only allowed on PP")
	}
	for _, agreement := range plan.Agreements {
		if agreement.ControllerInput == agreement.TargetInput {
			return CompiledPhrasePlan{}, invalid("SELF_PHRASE_AGREEMENT", "agreements", "agreement requires distinct explicit inputs")
		}
		if _, ok := inputsByID(plan.Inputs)[agreement.ControllerInput]; !ok {
			return CompiledPhrasePlan{}, unknown("UNKNOWN_AGREEMENT_CONTROLLER", "agreements", "agreement controller input is missing")
		}
		if _, ok := inputsByID(plan.Inputs)[agreement.TargetInput]; !ok {
			return CompiledPhrasePlan{}, unknown("UNKNOWN_AGREEMENT_TARGET", "agreements", "agreement target input is missing")
		}
		if _, ok := profile.Lookup(agreement.Feature); !ok {
			return CompiledPhrasePlan{}, unknown("UNKNOWN_AGREEMENT_FEATURE", "agreements", "agreement feature is not declared by the language profile")
		}
		if !containsFeatureID(template.FeatureIDs, agreement.Feature) {
			return CompiledPhrasePlan{}, unsupported("AGREEMENT_FEATURE_NOT_IN_TEMPLATE", "agreements", "agreement feature is not declared by the selected phrase template")
		}
	}
	if plan.Kind != PhraseVP && plan.FrameID != "" {
		return CompiledPhrasePlan{}, invalid("UNEXPECTED_PHRASE_FIELDS", "plan", "frame reference is only allowed on VP")
	}
	if plan.Kind == PhrasePP && len(plan.Agreements) > 0 {
		return CompiledPhrasePlan{}, invalid("UNEXPECTED_PP_AGREEMENT", "agreements", "PP plans use governed case, not input agreement rules")
	}
	if plan.GovernedCase.Feature == "" && (plan.GovernedCase.State != "" || plan.GovernedCase.Value != "" || plan.GovernedCase.Variable != nil) {
		return CompiledPhrasePlan{}, invalid("INVALID_GOVERNED_CASE", "governed_case", "governed case requires a complete feature value")
	}
	return CompiledPhrasePlan{Plan: clonePhrasePlan(plan)}, nil
}

func allowedPhraseChild(parent PhraseKind, role string, child PhraseKind) bool {
	switch parent {
	case PhraseNP:
		return (role == "modifier" || role == "quantity") && (child == PhraseNP || child == PhrasePP)
	case PhraseVP:
		return (role == "auxiliary" && child == PhraseVP) || (strings.HasPrefix(role, "argument_") && (child == PhraseNP || child == PhrasePP))
	case PhrasePP:
		return role == "complement" && child == PhraseNP
	}
	return false
}

func hasFrameArgument(frame valency.ValencyFrame, childRole string) bool {
	if !strings.HasPrefix(childRole, "argument_") {
		return false
	}
	role := strings.TrimPrefix(childRole, "argument_")
	for _, argument := range frame.Arguments {
		if string(argument.Role) == role {
			return true
		}
	}
	return false
}
func allowedInputRoles(inputs map[string]PhraseInput, allowed ...string) error {
	set := map[string]struct{}{}
	for _, role := range allowed {
		set[role] = struct{}{}
	}
	for role := range inputs {
		if _, ok := set[role]; !ok {
			return invalid("UNEXPECTED_PHRASE_INPUT_ROLE", "inputs."+role, "input role is not supported for this phrase kind")
		}
	}
	return nil
}
func argumentInputRoles(frame valency.ValencyFrame) []string {
	result := make([]string, 0, len(frame.Arguments))
	for _, arg := range frame.Arguments {
		result = append(result, "argument_"+string(arg.Role))
	}
	return result
}
func inputsByID(inputs []PhraseInput) map[string]PhraseInput {
	result := map[string]PhraseInput{}
	for _, input := range inputs {
		result[input.ID] = input
	}
	return result
}
func containsFeatureID(values []features.FeatureID, candidate features.FeatureID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func clonePhrasePlan(plan PhrasePlan) PhrasePlan {
	result := plan
	result.Inputs = append([]PhraseInput(nil), plan.Inputs...)
	result.Children = append([]PhraseChild(nil), plan.Children...)
	result.Agreements = append([]FeatureAgreement(nil), plan.Agreements...)
	result.Features.Values = append([]features.FeatureValue(nil), plan.Features.Values...)
	for i := range result.Features.Values {
		if result.Features.Values[i].Variable != nil {
			variable := *result.Features.Values[i].Variable
			result.Features.Values[i].Variable = &variable
		}
	}
	if plan.GovernedCase.Variable != nil {
		variable := *plan.GovernedCase.Variable
		result.GovernedCase.Variable = &variable
	}
	return result
}

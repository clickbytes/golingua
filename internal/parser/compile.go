package parser

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	maxPlanScalarBytes      = 2048
	maxPlanStringTotalBytes = 1 << 20
)

// Compile validates and snapshots construction plans into an immutable parser
// index. Each accepted input contributes exactly one sense/frame permission.
func Compile(inputs []PlanInput, budget Budget) (*Index, error) {
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, newParserError("inputs", "compiler requires at least one plan input")
	}
	if len(inputs) > construction.MaxCompiledPlans {
		return nil, newCompileLimitError("inputs", len(inputs), construction.MaxCompiledPlans)
	}

	itemCount := 0
	for i, input := range inputs {
		itemCount = countInputItems(itemCount, input)
		if itemCount > construction.MaxCompileItems {
			return nil, newCompileLimitError(fmt.Sprintf("inputs[%d]", i), itemCount, construction.MaxCompileItems)
		}
	}

	plans := make([]compiledPlan, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for i, input := range inputs {
		path := fmt.Sprintf("inputs[%d]", i)
		compiled, err := compilePlanInput(input, path)
		if err != nil {
			return nil, err
		}
		permission := compiled.permission
		key := string(compiled.plan.ID) + "\x00" + permission.SenseID.String() + "\x00" + string(permission.FrameID)
		if _, exists := seen[key]; exists {
			return nil, newParserError(path, "duplicate plan, sense, and frame input")
		}
		seen[key] = struct{}{}
		plans = append(plans, compiled)
	}
	if len(plans) == 0 {
		return nil, newParserError("inputs", "compiler produced no plans")
	}
	sort.Slice(plans, func(i, j int) bool {
		left, right := plans[i], plans[j]
		if left.plan.ID != right.plan.ID {
			return left.plan.ID < right.plan.ID
		}
		if left.permission.SenseID != right.permission.SenseID {
			return left.permission.SenseID.String() < right.permission.SenseID.String()
		}
		return left.permission.FrameID < right.permission.FrameID
	})
	index := &Index{
		plans: plans, byLanguage: make(map[features.LanguageID][]int),
		byFirstSymbol: make(map[planLookupKey][]int),
	}
	for i := range index.plans {
		plan := index.plans[i]
		language := plan.family.Language
		index.byLanguage[language] = append(index.byLanguage[language], i)
		seenKeys := make(map[planLookupKey]struct{}, len(plan.sequence))
		for _, symbol := range plan.sequence {
			key := planLookupKey{Language: language}
			if symbol.Slot != nil {
				key.Category = symbol.Slot.Category
			} else {
				key.Terminal = symbol.Terminal
			}
			if _, exists := seenKeys[key]; !exists {
				index.byFirstSymbol[key] = append(index.byFirstSymbol[key], i)
				seenKeys[key] = struct{}{}
			}
			if symbol.Slot == nil || symbol.Slot.Optionality != construction.SlotOptional {
				break
			}
		}
	}
	return index, nil
}

func compilePlanInput(input PlanInput, path string) (compiledPlan, error) {
	if err := input.Family.Validate(); err != nil {
		return compiledPlan{}, wrapValidationError(path+".family", err)
	}
	if err := input.Frame.Validate(); err != nil {
		return compiledPlan{}, wrapValidationError(path+".frame", err)
	}
	if err := input.Variant.Validate(input.Family); err != nil {
		return compiledPlan{}, wrapValidationError(path+".variant", err)
	}
	if _, err := construction.ValidateVariantForFrame(input.Variant, input.Family, input.Frame); err != nil {
		return compiledPlan{}, wrapValidationError(path+".variant", err)
	}
	if err := input.Binding.Provenance.Validate(); err != nil {
		return compiledPlan{}, newParserError(path+".binding.provenance", "approved construction binding requires valid provenance")
	}
	if !input.Binding.Approved {
		return compiledPlan{}, newParserError(path+".binding.approved", "construction binding must be explicitly approved")
	}
	if err := validatePlanScalars(input.Plan, path+".plan"); err != nil {
		return compiledPlan{}, err
	}
	if !input.Plan.ValidateIdentity() {
		return compiledPlan{}, newParserError(path+".plan", "compiled construction plan identity or permission digest is invalid")
	}
	if input.Plan.FamilyID != input.Family.ID || input.Plan.VariantID != input.Variant.ID || input.Plan.Language != string(input.Family.Language) {
		return compiledPlan{}, newParserError(path+".plan", "plan family, variant, or language does not match its supplied records")
	}
	if input.Variant.FamilyID != input.Family.ID {
		return compiledPlan{}, newParserError(path+".variant.family_id", "variant does not reference the supplied family")
	}
	if input.Variant.DirectionPolicy != construction.DirectionBoth && input.Variant.DirectionPolicy != construction.DirectionAnalyze {
		return compiledPlan{}, newParserError(path+".variant.direction_policy", "parser index requires an analysis-capable variant")
	}
	if input.Binding.SenseID != input.Frame.SenseID || input.Binding.ValencyFrameID != input.Frame.ID || input.Binding.FamilyID != input.Family.ID || input.Binding.VariantID.String() != input.Variant.ID.String() {
		return compiledPlan{}, newParserError(path+".binding", "approved construction binding does not match the supplied sense, frame, family, and variant")
	}
	if !sequenceEqual(input.Plan.Linearization, input.Variant.Sequence) {
		return compiledPlan{}, newParserError(path+".plan.linearization", "compiled plan sequence differs from the supplied variant")
	}
	if err := validatePlanArgumentBindings(input.Plan, input.Family, input.Frame, path+".plan.argument_bindings"); err != nil {
		return compiledPlan{}, err
	}
	permission := construction.SensePermission{SenseID: input.Frame.SenseID, FrameID: input.Frame.ID}
	if !hasPermission(input.Plan.Permissions, permission) {
		return compiledPlan{}, newParserError(path+".plan.permissions", "compiled plan does not grant the supplied sense and frame permission")
	}
	family := cloneFamily(input.Family)
	frame := cloneFrame(input.Frame)
	variant := cloneVariant(input.Variant)
	plan := cloneCompiledPlan(input.Plan)
	slots := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		slots[slot.ID] = cloneSlot(slot)
	}
	sequence := make([]compiledSymbol, len(variant.Sequence))
	headSlotID := ""
	for i, token := range variant.Sequence {
		symbol := compiledSymbol{Position: i}
		switch token.Kind {
		case construction.TokenSlot:
			slot, ok := slots[token.SlotID]
			if !ok {
				return compiledPlan{}, newParserError(path+fmt.Sprintf(".variant.sequence[%d]", i), "variant references a missing family slot")
			}
			slotCopy := cloneSlot(slot)
			symbol.Slot = &slotCopy
			if slot.Category == construction.CategoryVP {
				if headSlotID != "" {
					return compiledPlan{}, newParserError(path+".variant.sequence", "analysis plan must have exactly one VP predicate slot")
				}
				headSlotID = slot.ID
			}
		case construction.TokenTerminal:
			symbol.Terminal = token.Terminal
		default:
			return compiledPlan{}, newParserError(path+fmt.Sprintf(".variant.sequence[%d]", i), "variant contains an unsupported sequence token")
		}
		sequence[i] = symbol
	}
	if headSlotID == "" {
		return compiledPlan{}, newParserError(path+".variant.sequence", "analysis plan must have exactly one VP predicate slot")
	}
	return compiledPlan{
		plan: plan, family: family, variant: variant, frame: frame, headSlotID: headSlotID,
		permission: permission, slots: slots, sequence: sequence,
	}, nil
}

func hasPermission(permissions []construction.SensePermission, wanted construction.SensePermission) bool {
	for _, permission := range permissions {
		if permission == wanted {
			return true
		}
	}
	return false
}

func sequenceEqual(left, right []construction.SequenceToken) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func validatePlanArgumentBindings(plan construction.CompiledConstructionPlan, family construction.ConstructionFamily, frame valency.ValencyFrame, path string) error {
	slotsByRole := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		if slot.ArgumentBinding != nil {
			slotsByRole[string(slot.ArgumentBinding.SemanticRole)] = slot
		}
	}
	expected := make(map[string]construction.CompiledArgumentBinding, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		slot, exists := slotsByRole[string(argument.Role)]
		if !exists {
			if argument.Required {
				return newParserError(path, "required frame argument has no family slot binding")
			}
			continue
		}
		if slot.Case != "" && argument.Case != "" && slot.Case != argument.Case {
			return newParserError(path, "frame and family case policies conflict")
		}
		if slot.Adposition != "" && argument.Adposition != "" && slot.Adposition != argument.Adposition {
			return newParserError(path, "frame and family adposition policies conflict")
		}
		casePolicy := argument.Case
		if casePolicy == "" {
			casePolicy = slot.Case
		}
		adposition := argument.Adposition
		if adposition == "" {
			adposition = slot.Adposition
		}
		expected[argument.ID] = construction.CompiledArgumentBinding{
			ArgumentID: argument.ID, Role: string(argument.Role), Required: argument.Required,
			SlotID: slot.ID, CasePolicy: casePolicy, Adposition: adposition,
		}
	}
	if len(plan.ArgumentBindings) != len(expected) {
		return newParserError(path, "compiled argument bindings do not match the supplied frame and family")
	}
	for _, binding := range plan.ArgumentBindings {
		want, exists := expected[binding.ArgumentID]
		if !exists || binding != want {
			return newParserError(path, "compiled argument bindings do not match the supplied frame and family")
		}
		delete(expected, binding.ArgumentID)
	}
	if len(expected) != 0 {
		return newParserError(path, "compiled argument bindings omit a frame argument mapping")
	}
	return nil
}

func countInputItems(current int, input PlanInput) int {
	count := current
	add := func(size int) {
		if count > construction.MaxCompileItems || size > construction.MaxCompileItems-count {
			count = construction.MaxCompileItems + 1
			return
		}
		count += size
	}
	add(len(input.Family.SlotSpecs))
	add(len(input.Family.Constraints))
	add(len(input.Family.Operators))
	add(len(input.Family.Capabilities))
	if count > construction.MaxCompileItems {
		return count
	}
	for _, constraint := range input.Family.Constraints {
		add(len(constraint.Features))
		if count > construction.MaxCompileItems {
			return count
		}
	}
	add(len(input.Frame.SemanticRoles))
	add(len(input.Frame.Arguments))
	add(len(input.Frame.SelectionalConstraints))
	add(len(input.Frame.AllowedFamilyIDs))
	for _, constraint := range input.Frame.SelectionalConstraints {
		add(len(constraint.AllowedConceptIDs))
		add(len(constraint.AllowedKinds))
		if count > construction.MaxCompileItems {
			return count
		}
	}
	add(len(input.Plan.Permissions))
	add(len(input.Plan.ArgumentBindings))
	add(len(input.Plan.Constraints))
	add(len(input.Plan.Linearization))
	add(len(input.Plan.ProcedureIDs))
	for _, constraint := range input.Plan.Constraints {
		add(len(constraint.Features))
		if count > construction.MaxCompileItems {
			return count
		}
	}
	add(len(input.Variant.Sequence))
	add(len(input.Variant.Conditions))
	return count
}

func validatePlanScalars(plan construction.CompiledConstructionPlan, path string) error {
	if err := plan.ProfileProvenance.Validate(); err != nil {
		return newParserError(path+".profile_provenance", "compiled plan requires valid profile provenance")
	}
	if !validPlanVersion(string(plan.ProfileVersion)) {
		return newParserError(path+".profile_version", "compiled plan profile version must be a bounded lowercase token")
	}
	bytes := 0
	add := func(value, field string) error {
		if len(value) > maxPlanScalarBytes {
			return newParserError(path+"."+field, "compiled plan scalar exceeds the byte limit")
		}
		if len(value) > maxPlanStringTotalBytes-bytes {
			return newParserError(path, "compiled plan string data exceeds the aggregate byte limit")
		}
		bytes += len(value)
		return nil
	}
	addSource := func(sourcePath construction.SourcePath, field string) error {
		for _, value := range []struct{ value, name string }{
			{sourcePath.SlotID, field + ".slot_id"},
			{sourcePath.ArgumentID, field + ".argument_id"},
			{string(sourcePath.Feature), field + ".feature"},
		} {
			if err := add(value.value, value.name); err != nil {
				return err
			}
		}
		return nil
	}
	for _, value := range []struct{ value, name string }{
		{string(plan.ID), "id"}, {plan.Language, "language"}, {string(plan.ProfileVersion), "profile_version"},
		{plan.ProfileProvenance.ID, "profile_provenance.id"}, {plan.ProfileProvenance.Path, "profile_provenance.path"},
		{plan.ProfileProvenance.URI, "profile_provenance.uri"}, {plan.ProfileProvenance.Version, "profile_provenance.version"},
		{plan.ProfileProvenance.License, "profile_provenance.license"}, {plan.ProfileProvenance.Author, "profile_provenance.author"},
		{plan.ProfileProvenance.ChecksumSHA256, "profile_provenance.checksum_sha256"},
		{plan.PermissionDigest, "permission_digest"}, {plan.FamilyID.String(), "family_id"}, {plan.VariantID.String(), "variant_id"},
	} {
		if err := add(value.value, value.name); err != nil {
			return err
		}
	}
	for i, permission := range plan.Permissions {
		if err := add(permission.SenseID.String(), fmt.Sprintf("permissions[%d].sense_id", i)); err != nil {
			return err
		}
		if err := add(string(permission.FrameID), fmt.Sprintf("permissions[%d].frame_id", i)); err != nil {
			return err
		}
	}
	for i, binding := range plan.ArgumentBindings {
		for _, value := range []struct{ value, name string }{
			{binding.ArgumentID, "argument_id"}, {binding.Role, "role"}, {binding.SlotID, "slot_id"},
			{binding.CasePolicy, "case_policy"}, {binding.Adposition, "adposition"},
		} {
			if err := add(value.value, fmt.Sprintf("argument_bindings[%d].%s", i, value.name)); err != nil {
				return err
			}
		}
	}
	for i, constraint := range plan.Constraints {
		field := fmt.Sprintf("constraints[%d]", i)
		for _, value := range []struct{ value, name string }{
			{constraint.ID, ".id"}, {string(constraint.Kind), ".kind"}, {string(constraint.Strength), ".strength"},
			{constraint.CasePolicy, ".case_policy"}, {constraint.BeforeSlot, ".before_slot"},
			{constraint.AfterSlot, ".after_slot"}, {constraint.OptionalSlot, ".optional_slot"},
		} {
			if err := add(value.value, field+value.name); err != nil {
				return err
			}
		}
		if err := addSource(constraint.Target, field+".target"); err != nil {
			return err
		}
		if err := addSource(constraint.Controller, field+".controller"); err != nil {
			return err
		}
		for j, feature := range constraint.Features {
			if err := add(string(feature), fmt.Sprintf("%s.features[%d]", field, j)); err != nil {
				return err
			}
		}
		if predicate := constraint.Condition; predicate != nil {
			if err := add(string(predicate.Kind), field+".condition.kind"); err != nil {
				return err
			}
			if err := addSource(predicate.Source, field+".condition.source"); err != nil {
				return err
			}
			for _, value := range []struct{ value, name string }{
				{string(predicate.Value), ".condition.value"}, {predicate.SlotID, ".condition.slot_id"},
			} {
				if err := add(value.value, field+value.name); err != nil {
					return err
				}
			}
		}
	}
	for i, token := range plan.Linearization {
		if err := add(string(token.Kind), fmt.Sprintf("linearization[%d].kind", i)); err != nil {
			return err
		}
		if err := add(token.SlotID, fmt.Sprintf("linearization[%d].slot_id", i)); err != nil {
			return err
		}
		if err := add(token.Terminal, fmt.Sprintf("linearization[%d].terminal", i)); err != nil {
			return err
		}
	}
	for i, procedure := range plan.ProcedureIDs {
		if err := add(procedure, fmt.Sprintf("procedure_ids[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

func validPlanVersion(value string) bool {
	if len(value) == 0 || len(value) > features.MaxTokenBytes || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func cloneFamily(family construction.ConstructionFamily) construction.ConstructionFamily {
	family.SlotSpecs = append([]construction.SlotSpec(nil), family.SlotSpecs...)
	for i := range family.SlotSpecs {
		family.SlotSpecs[i] = cloneSlot(family.SlotSpecs[i])
	}
	family.Constraints = append([]construction.FamilyConstraint(nil), family.Constraints...)
	for i := range family.Constraints {
		family.Constraints[i].Features = append([]features.FeatureID(nil), family.Constraints[i].Features...)
	}
	family.Operators = append([]construction.OperatorID(nil), family.Operators...)
	family.Capabilities = append([]construction.Capability(nil), family.Capabilities...)
	return family
}

func cloneSlot(slot construction.SlotSpec) construction.SlotSpec {
	if slot.ArgumentBinding != nil {
		binding := *slot.ArgumentBinding
		slot.ArgumentBinding = &binding
	}
	return slot
}

func cloneFrame(frame valency.ValencyFrame) valency.ValencyFrame {
	frame.SemanticRoles = append([]meaning.Role(nil), frame.SemanticRoles...)
	frame.Arguments = append([]valency.ArgumentSlot(nil), frame.Arguments...)
	frame.SelectionalConstraints = append([]valency.SelectionalConstraint(nil), frame.SelectionalConstraints...)
	for i := range frame.SelectionalConstraints {
		frame.SelectionalConstraints[i].AllowedConceptIDs = append([]ids.ConceptID(nil), frame.SelectionalConstraints[i].AllowedConceptIDs...)
		frame.SelectionalConstraints[i].AllowedKinds = append([]core.ConceptKind(nil), frame.SelectionalConstraints[i].AllowedKinds...)
	}
	frame.AllowedFamilyIDs = append([]ids.ConstructionID(nil), frame.AllowedFamilyIDs...)
	return frame
}

func cloneVariant(variant construction.Variant) construction.Variant {
	variant.Sequence = append([]construction.SequenceToken(nil), variant.Sequence...)
	variant.Conditions = append([]construction.Condition(nil), variant.Conditions...)
	return variant
}

func cloneCompiledPlan(plan construction.CompiledConstructionPlan) construction.CompiledConstructionPlan {
	plan.Permissions = append([]construction.SensePermission(nil), plan.Permissions...)
	plan.ArgumentBindings = append([]construction.CompiledArgumentBinding(nil), plan.ArgumentBindings...)
	plan.Constraints = append([]construction.CompiledConstraint(nil), plan.Constraints...)
	for i := range plan.Constraints {
		plan.Constraints[i].Features = append([]features.FeatureID(nil), plan.Constraints[i].Features...)
		if plan.Constraints[i].Condition != nil {
			condition := *plan.Constraints[i].Condition
			plan.Constraints[i].Condition = &condition
		}
	}
	plan.Linearization = append([]construction.SequenceToken(nil), plan.Linearization...)
	plan.ProcedureIDs = append([]string(nil), plan.ProcedureIDs...)
	return plan
}

func newCompileLimitError(path string, actual, limit int) *ParseError {
	return &ParseError{result: status.Result{
		Status: status.OutcomeResourceLimit,
		Diagnostics: []status.Diagnostic{{
			Code: "PARSER_COMPILE_LIMIT", Message: fmt.Sprintf("compile input exceeds bounded limit %d (got %d)", limit, actual),
			Path: path, Severity: status.SeverityError,
		}},
	}}
}

func wrapValidationError(path string, err error) error {
	if err == nil {
		return nil
	}
	typed, ok := err.(interface{ Result() status.Result })
	if !ok {
		return newParserError(path, err.Error())
	}
	result := typed.Result()
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Path == "" {
			result.Diagnostics[i].Path = path
		} else {
			result.Diagnostics[i].Path = path + "." + result.Diagnostics[i].Path
		}
	}
	return &ParseError{result: result}
}

package synthesis

import (
	"fmt"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const MaxRefinementChoices = 64

// ValidateGoalRefinement independently checks that concrete is an admissible
// refinement of original and that search-origin choices name those exact
// refinements. It validates claims; it never generates a candidate.
func ValidateGoalRefinement(primitiveContext algebra.Document, original, concrete GoalSpec, choices []ChoiceAssignment, origin string) error {
	if len(choices) > MaxRefinementChoices {
		return goalResource("choices", "choice count exceeds the bounded goal-axis limit")
	}
	for i, choice := range choices {
		if err := validateRefinementChoiceShape(choice, i); err != nil {
			return err
		}
	}

	if _, err := validateSearchGoal(primitiveContext, original); err != nil {
		return err
	}
	if origin != CertificateOriginConstruct && origin != CertificateOriginSearch {
		return goalUnsupported("UNSUPPORTED_CERTIFICATE_ORIGIN", "origin", "certificate origin must be construct or search")
	}
	if origin == CertificateOriginConstruct {
		if len(choices) != 0 {
			return goalInvalid("CONSTRUCT_CHOICES_NOT_ALLOWED", "choices", "construct-origin refinement cannot include search choices")
		}
		if unresolved, validateErr := ValidateGoal(primitiveContext, original); validateErr != nil {
			return validateErr
		} else if len(unresolved) != 0 {
			return refinementIncomplete("ORIGINAL_GOAL_INCOMPLETE", "original", "construct origin requires a fully selected original goal")
		}
	}

	concreteUnresolved, err := ValidateGoal(primitiveContext, concrete)
	if err != nil {
		return err
	}
	if len(concreteUnresolved) != 0 {
		return refinementIncomplete("INCOMPLETE_REFINEMENT", "concrete_goal", "concrete goal still has unresolved choices")
	}

	if err := compareRefinementFields(original, concrete); err != nil {
		return err
	}
	if err := compareGoalRequestContext(primitiveContext, original, concrete); err != nil {
		return err
	}
	if origin == CertificateOriginSearch {
		if err := validateSearchRefinementChoices(primitiveContext, original, concrete, choices); err != nil {
			return err
		}
	}
	return nil
}

func goalContextFingerprint(declarations algebra.Document, goal GoalSpec) (string, error) {
	if goal.RequestContext == nil {
		return "", nil
	}
	document := declarations
	document.SchemaVersion = algebra.SchemaVersion
	document.Expression = nil
	document.Parameters = nil
	document.References = append([]algebra.Declaration{}, goal.References...)
	document.RequestContext = goal.RequestContext
	return algebra.RequestContextFingerprint(*goal.RequestContext, document)
}

func compareGoalRequestContext(primitiveContext algebra.Document, original, concrete GoalSpec) error {
	if (original.RequestContext == nil) != (concrete.RequestContext == nil) {
		return refinementInvalid("REFINEMENT_REQUEST_CONTEXT_CHANGED", "request_context", "request context presence must be preserved")
	}
	if original.RequestContext == nil {
		return nil
	}
	fingerprint := func(goal GoalSpec) (string, error) {
		document := primitiveContext
		document.Expression = nil
		document.Parameters = nil
		document.References = append([]algebra.Declaration{}, goal.References...)
		document.RequestContext = goal.RequestContext
		return algebra.RequestContextFingerprint(*goal.RequestContext, document)
	}
	originalFingerprint, err := fingerprint(original)
	if err != nil {
		return err
	}
	concreteFingerprint, err := fingerprint(concrete)
	if err != nil {
		return err
	}
	if originalFingerprint != concreteFingerprint {
		return refinementInvalid("REFINEMENT_REQUEST_CONTEXT_CHANGED", "request_context", "request context payload must be preserved")
	}
	return nil
}

func validateRefinementChoiceShape(choice ChoiceAssignment, index int) error {
	fields := []struct {
		name     string
		value    string
		required bool
	}{
		{"kind", choice.Kind, true}, {"axis_id", choice.AxisID, true},
		{"constraint_id", choice.ConstraintID, false}, {"symbol", choice.Symbol, false},
		{"from_role", choice.FromRole, false}, {"to_role", choice.ToRole, false},
		{"port_id", choice.PortID, false}, {"unary_role", choice.UnaryRole, false},
		{"hole_id", choice.HoleID, false}, {"reference_id", choice.ReferenceID, false},
	}
	for _, field := range fields {
		if len(field.value) > 256 || !utf8.ValidString(field.value) || field.required && !goalID(field.value) || !field.required && field.value != "" && !goalID(field.value) {
			return goalInvalid("INVALID_REFINEMENT_CHOICE", fmt.Sprintf("choices[%d].%s", index, field.name), "choice text must be bounded non-empty identifier text when present")
		}
	}
	if choice.AtomCount < 0 || choice.AtomCount > MaxGoalAtoms {
		return goalInvalid("INVALID_REFINEMENT_CHOICE", fmt.Sprintf("choices[%d].atom_count", index), "choice atom count is outside the positive-facts profile bound")
	}
	return nil
}

func compareRefinementFields(original, concrete GoalSpec) error {
	if original.SchemaVersion != concrete.SchemaVersion || original.Profile != concrete.Profile || original.ContextID != concrete.ContextID {
		return refinementInvalid("REFINEMENT_CONTEXT_CHANGED", "concrete_goal", "schema, profile, and context ID must be preserved")
	}
	if !algebra.TypeEqual(original.ExpectedType, concrete.ExpectedType) {
		return refinementInvalid("REFINEMENT_EXPECTED_TYPE_CHANGED", "expected_type", "expected type must be preserved")
	}
	if len(original.Ports) != len(concrete.Ports) {
		return refinementInvalid("REFINEMENT_PORTS_CHANGED", "ports", "goal port inventory and order must be preserved")
	}
	for i, originalPort := range original.Ports {
		concretePort := concrete.Ports[i]
		path := fmt.Sprintf("ports[%d]", i)
		if originalPort.ID != concretePort.ID || !algebra.TypeEqual(originalPort.Type, concretePort.Type) {
			return refinementInvalid("REFINEMENT_PORTS_CHANGED", path, "goal port IDs, order, and types must be preserved")
		}
		if err := comparePortBinding(originalPort, concretePort, original.References, path); err != nil {
			return err
		}
	}
	if !sameDeclarationInventory(original.References, concrete.References) {
		return refinementInvalid("REFINEMENT_REFERENCES_CHANGED", "references", "reference IDs and types must be preserved without merging or redefinition")
	}
	if !sameStringSet(original.ProjectPorts, concrete.ProjectPorts) {
		return refinementInvalid("REFINEMENT_PROJECTION_CHANGED", "project_ports", "the explicit projection set must be preserved")
	}
	if !sameStringSet(original.AllowedOperators, concrete.AllowedOperators) {
		return refinementInvalid("REFINEMENT_OPERATOR_CEILING_CHANGED", "allowed_operators", "the operator ceiling must be preserved")
	}
	if !sameSourceRefs(original.Provenance, concrete.Provenance) {
		return refinementInvalid("REFINEMENT_PROVENANCE_CHANGED", "provenance", "source references and their order must be preserved")
	}
	originalLimits, err := NormalizeGoalLimits(original.Limits)
	if err != nil {
		return err
	}
	concreteLimits, err := NormalizeGoalLimits(concrete.Limits)
	if err != nil {
		return err
	}
	if originalLimits != concreteLimits {
		return refinementInvalid("REFINEMENT_LIMITS_CHANGED", "limits", "normalized goal limits must be preserved")
	}
	if len(original.Constraints) != len(concrete.Constraints) {
		return refinementInvalid("REFINEMENT_CONSTRAINTS_CHANGED", "constraints", "constraint inventory and order must be preserved")
	}
	for i, originalConstraint := range original.Constraints {
		concreteConstraint := concrete.Constraints[i]
		path := fmt.Sprintf("constraints[%d]", i)
		if originalConstraint.ID != concreteConstraint.ID || originalConstraint.Kind != concreteConstraint.Kind ||
			originalConstraint.FromPort != concreteConstraint.FromPort || originalConstraint.ToPort != concreteConstraint.ToPort {
			return refinementInvalid("REFINEMENT_CONSTRAINTS_CHANGED", path, "constraint IDs, kinds, order, and path endpoints must be preserved")
		}
		if originalConstraint.Symbol != "" && originalConstraint.Symbol != concreteConstraint.Symbol {
			return refinementInvalid("REFINEMENT_SYMBOL_CHANGED", path+".symbol", "a selected source symbol cannot be replaced")
		}
		if originalConstraint.FromRole != "" && originalConstraint.FromRole != concreteConstraint.FromRole {
			return refinementInvalid("REFINEMENT_ROLE_CHANGED", path+".from_role", "a selected source role cannot be replaced")
		}
		if originalConstraint.ToRole != "" && originalConstraint.ToRole != concreteConstraint.ToRole {
			return refinementInvalid("REFINEMENT_ROLE_CHANGED", path+".to_role", "a selected target role cannot be replaced")
		}
		if originalConstraint.Port != "" && originalConstraint.Port != concreteConstraint.Port {
			return refinementInvalid("REFINEMENT_UNARY_PORT_CHANGED", path+".port", "a selected unary target port cannot be replaced")
		}
		if originalConstraint.Kind == ConstraintPath {
			if len(concreteConstraint.AllowedAtomCounts) != 1 {
				return refinementInvalid("PATH_COUNT_NOT_SELECTED", path+".allowed_atom_counts", "concrete path must select exactly one occurrence count")
			}
			if len(originalConstraint.AllowedAtomCounts) != 0 && !containsCount(originalConstraint.AllowedAtomCounts, concreteConstraint.AllowedAtomCounts[0]) {
				return refinementInvalid("PATH_COUNT_NOT_ALLOWED", path+".allowed_atom_counts", "selected count was not declared by the original goal")
			}
		}
	}
	return nil
}

func comparePortBinding(original, concrete GoalPort, references []algebra.Declaration, path string) error {
	if original.Binding == nil {
		if concrete.Binding != nil {
			return refinementInvalid("REFINEMENT_OPEN_PORT_BOUND", path+".binding", "an unbound goal port must remain unbound")
		}
		return nil
	}
	if original.Binding.Kind == algebra.TermRef {
		if concrete.Binding == nil || concrete.Binding.Kind != algebra.TermRef || concrete.Binding.ID != original.Binding.ID || concrete.Binding.Type == nil ||
			!algebra.TypeEqual(*concrete.Binding.Type, *original.Binding.Type) {
			return refinementInvalid("REFINEMENT_REFERENCE_BINDING_CHANGED", path+".binding", "an explicit reference binding must be preserved exactly")
		}
		return nil
	}
	if original.Binding.Kind != algebra.TermHole {
		return refinementInvalid("UNSUPPORTED_ORIGINAL_BINDING", path+".binding", "original binding kind is outside the refinement profile")
	}
	if concrete.Binding == nil || concrete.Binding.Kind != algebra.TermRef || concrete.Binding.Type == nil ||
		!algebra.TypeEqual(*concrete.Binding.Type, *original.Binding.Type) ||
		!declaredReferenceOfType(references, concrete.Binding.ID, *original.Binding.Type) {
		return refinementInvalid("INVALID_HOLE_RESOLUTION", path+".binding", "a typed Hole may resolve only to an existing declared reference of the exact type")
	}
	return nil
}

func sameDeclarationInventory(left, right []algebra.Declaration) bool {
	if len(left) != len(right) {
		return false
	}
	byID := make(map[string]algebra.Declaration, len(left))
	for _, declaration := range left {
		byID[declaration.ID] = declaration
	}
	for _, declaration := range right {
		prior, exists := byID[declaration.ID]
		if !exists || !algebra.TypeEqual(prior.Type, declaration.Type) ||
			prior.TemporalFrame != declaration.TemporalFrame || prior.ReferenceTimeID != declaration.ReferenceTimeID || prior.Multiplicity != declaration.Multiplicity {
			return false
		}
	}
	return true
}

func declaredReferenceOfType(references []algebra.Declaration, id string, typ algebra.Type) bool {
	for _, reference := range references {
		if reference.ID == id && algebra.TypeEqual(reference.Type, typ) {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	set := make(map[string]struct{}, len(left))
	for _, value := range left {
		set[value] = struct{}{}
	}
	for _, value := range right {
		if _, exists := set[value]; !exists {
			return false
		}
	}
	return true
}

func sameSourceRefs(left, right []provenance.SourceRef) bool {
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

func containsCount(counts []int, selected int) bool {
	for _, count := range counts {
		if count == selected {
			return true
		}
	}
	return false
}

func validateSearchRefinementChoices(context algebra.Document, original, concrete GoalSpec, choices []ChoiceAssignment) error {
	pathIndex := -1
	for i, constraint := range original.Constraints {
		if constraint.Kind == ConstraintPath {
			pathIndex = i
			break
		}
	}
	if pathIndex < 0 {
		return refinementIncomplete("ORIGINAL_PATH_UNRESOLVED", "constraints", "search refinement requires an original binary path constraint")
	}
	pathOriginal := original.Constraints[pathIndex]
	pathConcrete := concrete.Constraints[pathIndex]
	if len(choices) == 0 {
		return refinementInvalid("MISSING_SEARCH_CHOICES", "choices", "search origin requires the explicit finite-axis choice sequence")
	}
	position := 0
	if err := checkPathSignatureChoice(choices, &position, pathOriginal, pathConcrete); err != nil {
		return err
	}
	if err := checkPathLengthChoice(choices, &position, pathOriginal, pathConcrete); err != nil {
		return err
	}
	axisNumber := 3
	for i, constraint := range original.Constraints {
		if constraint.Kind != ConstraintUnary {
			continue
		}
		if position >= len(choices) {
			return refinementInvalid("MISSING_SEARCH_CHOICE", fmt.Sprintf("choices[%d]", position), "choice sequence omits a unary constraint axis")
		}
		choice := choices[position]
		concreteConstraint := concrete.Constraints[i]
		if choice.Kind != "unary" || choice.AxisID != fmt.Sprintf("axis-%d", axisNumber) || choice.ConstraintID != constraint.ID ||
			choice.Symbol != concreteConstraint.Symbol || choice.PortID != concreteConstraint.Port || choice.UnaryRole == "" ||
			choice.FromRole != "" || choice.ToRole != "" || choice.AtomCount != 0 || choice.HoleID != "" || choice.ReferenceID != "" || choice.OutsidePathBound {
			return refinementInvalid("SEARCH_CHOICE_MISMATCH", fmt.Sprintf("choices[%d]", position), "unary choice does not match its ordered source constraint and concrete selection")
		}
		signature, exists := goalSignature(context.Signatures, concreteConstraint.Symbol)
		if !exists || len(signature.Ports) != 1 || signature.Ports[0].Role != choice.UnaryRole ||
			!algebra.TypeEqual(signature.Ports[0].Type, goalPortType(concrete, concreteConstraint.Port)) {
			return refinementInvalid("SEARCH_CHOICE_MISMATCH", fmt.Sprintf("choices[%d].unary_role", position), "unary role must be the declared typed port of its selected primitive")
		}
		if constraint.Symbol != "" && choice.Symbol != constraint.Symbol || constraint.Port != "" && choice.PortID != constraint.Port {
			return refinementInvalid("SEARCH_CHOICE_CHANGED_FIXED_FIELD", fmt.Sprintf("choices[%d]", position), "unary choice changed a field fixed by the original goal")
		}
		position++
		axisNumber++
	}
	for _, port := range original.Ports {
		if port.Binding == nil || port.Binding.Kind != algebra.TermHole {
			continue
		}
		if position >= len(choices) {
			return refinementInvalid("MISSING_SEARCH_CHOICE", fmt.Sprintf("choices[%d]", position), "choice sequence omits a typed-hole reference axis")
		}
		choice := choices[position]
		concretePort, _ := goalPort(concrete, port.ID)
		if choice.Kind != "hole_binding" || choice.AxisID != fmt.Sprintf("axis-%d", axisNumber) || choice.PortID != port.ID ||
			choice.HoleID != port.Binding.ID || concretePort.Binding == nil || choice.ReferenceID != concretePort.Binding.ID ||
			choice.ConstraintID != "" || choice.Symbol != "" || choice.FromRole != "" || choice.ToRole != "" || choice.AtomCount != 0 || choice.UnaryRole != "" || choice.OutsidePathBound {
			return refinementInvalid("SEARCH_CHOICE_MISMATCH", fmt.Sprintf("choices[%d]", position), "hole choice must bind the original hole to the concrete declared reference")
		}
		if !declaredReferenceOfType(original.References, choice.ReferenceID, *port.Binding.Type) {
			return refinementInvalid("INVALID_HOLE_RESOLUTION", fmt.Sprintf("choices[%d].reference_id", position), "hole choice reference is not declared with the exact hole type")
		}
		position++
		axisNumber++
	}
	if position != len(choices) {
		return refinementInvalid("EXTRA_SEARCH_CHOICE", fmt.Sprintf("choices[%d]", position), "choice sequence contains an unrecognized or duplicate axis")
	}
	return nil
}

func checkPathSignatureChoice(choices []ChoiceAssignment, position *int, original, concrete Constraint) error {
	if *position >= len(choices) {
		return refinementInvalid("MISSING_SEARCH_CHOICE", fmt.Sprintf("choices[%d]", *position), "choice sequence omits the binary signature axis")
	}
	choice := choices[*position]
	if choice.Kind != "path_signature" || choice.AxisID != "axis-1" || choice.ConstraintID != original.ID ||
		choice.Symbol != concrete.Symbol || choice.FromRole != concrete.FromRole || choice.ToRole != concrete.ToRole ||
		choice.PortID != concrete.FromPort || choice.AtomCount != 0 || choice.UnaryRole != "" || choice.HoleID != "" || choice.ReferenceID != "" || choice.OutsidePathBound {
		return refinementInvalid("SEARCH_CHOICE_MISMATCH", fmt.Sprintf("choices[%d]", *position), "binary signature choice does not match the concrete path selection")
	}
	if original.Symbol != "" && choice.Symbol != original.Symbol || original.FromRole != "" && choice.FromRole != original.FromRole || original.ToRole != "" && choice.ToRole != original.ToRole {
		return refinementInvalid("SEARCH_CHOICE_CHANGED_FIXED_FIELD", fmt.Sprintf("choices[%d]", *position), "binary signature choice changed a field fixed by the original goal")
	}
	*position++
	return nil
}

func checkPathLengthChoice(choices []ChoiceAssignment, position *int, original, concrete Constraint) error {
	if *position >= len(choices) {
		return refinementInvalid("MISSING_SEARCH_CHOICE", fmt.Sprintf("choices[%d]", *position), "choice sequence omits the path-length axis")
	}
	choice := choices[*position]
	if len(concrete.AllowedAtomCounts) != 1 || choice.Kind != "path_length" || choice.AxisID != "axis-2" ||
		choice.ConstraintID != original.ID || choice.AtomCount != concrete.AllowedAtomCounts[0] || choice.OutsidePathBound ||
		choice.Symbol != "" || choice.FromRole != "" || choice.ToRole != "" || choice.PortID != "" || choice.UnaryRole != "" || choice.HoleID != "" || choice.ReferenceID != "" {
		return refinementInvalid("SEARCH_CHOICE_MISMATCH", fmt.Sprintf("choices[%d]", *position), "path-length choice does not match the concrete in-bound count")
	}
	if len(original.AllowedAtomCounts) != 0 && !containsCount(original.AllowedAtomCounts, choice.AtomCount) {
		return refinementInvalid("SEARCH_CHOICE_NOT_ALLOWED", fmt.Sprintf("choices[%d].atom_count", *position), "selected count was not one of the original path choices")
	}
	*position++
	return nil
}

func goalPort(goal GoalSpec, id string) (GoalPort, bool) {
	for _, port := range goal.Ports {
		if port.ID == id {
			return port, true
		}
	}
	return GoalPort{}, false
}

func goalPortType(goal GoalSpec, id string) algebra.Type {
	port, exists := goalPort(goal, id)
	if !exists {
		return algebra.Type{}
	}
	return port.Type
}

func refinementInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func refinementIncomplete(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeIncomplete, Code: code, Path: path, Message: message}
}

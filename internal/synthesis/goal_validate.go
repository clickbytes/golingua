package synthesis

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// Unresolved records an explicit goal choice that is still required. It is
// not a candidate construction or a claim that the goal is satisfiable.
type Unresolved struct {
	ConstraintID string        `json:"constraint_id,omitempty"`
	Code         string        `json:"code"`
	Path         string        `json:"path"`
	Message      string        `json:"message"`
	Hole         *algebra.Type `json:"hole,omitempty"`
}

// ValidateGoal checks the bounded positive-facts goal contract. It validates
// structure and declarations only; it does not construct a term or prove
// consistency with any world.
func ValidateGoal(context algebra.Document, goal GoalSpec) ([]Unresolved, error) {
	return validateGoal(context, goal, false)
}

// validateSearchGoal checks the goal's formal request shape while leaving
// candidate-specific fit and operator-budget admission to the bounded search.
// It does not establish that any candidate exists or that the goal is true.
func validateSearchGoal(context algebra.Document, goal GoalSpec) ([]Unresolved, error) {
	return validateGoal(context, goal, true)
}

func validateGoal(context algebra.Document, goal GoalSpec, forSearch bool) ([]Unresolved, error) {
	if context.Expression != nil || len(context.Parameters) != 0 || len(context.References) != 0 {
		return nil, goalInvalid("INVALID_CONTEXT", "context", "context must contain declarations only; request references belong to the goal")
	}
	if context.SchemaVersion != algebra.SchemaVersion {
		return nil, goalInvalid("INVALID_CONTEXT", "context.schema_version", "context declaration version is unsupported")
	}
	if len(context.Sorts)+len(context.Signatures) > algebra.DefaultMaxSymbols {
		return nil, goalResource("context", "context declaration count exceeds the bounded profile")
	}
	if err := algebra.CheckDeclarations(context); err != nil {
		return nil, err
	}
	if goal.SchemaVersion != GoalSchemaVersion {
		return nil, goalInvalid("INVALID_GOAL_SCHEMA_VERSION", "schema_version", "goal schema version is unsupported")
	}
	if !goalID(goal.Profile) || goal.Profile != GoalProfile {
		return nil, goalInvalid("UNSUPPORTED_GOAL_PROFILE", "profile", "goal profile is unsupported")
	}
	if !goalID(goal.ContextID) {
		return nil, goalInvalid("INVALID_CONTEXT_ID", "context_id", "context ID must be non-empty UTF-8, whitespace-free, and at most 256 bytes")
	}
	limits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return nil, err
	}
	if len(goal.Ports) == 0 || len(goal.Ports) > MaxGoalPorts {
		return nil, goalBoundError("ports", len(goal.Ports), MaxGoalPorts)
	}
	if len(goal.ProjectPorts) > MaxGoalPorts {
		return nil, goalResource("project_ports", "projected port count exceeds the bounded goal-port limit")
	}
	if len(goal.Constraints) > MaxGoalConstraints {
		return nil, goalBoundError("constraints", len(goal.Constraints), MaxGoalConstraints)
	}
	if len(goal.Provenance) > MaxGoalSources {
		return nil, goalBoundError("provenance", len(goal.Provenance), MaxGoalSources)
	}
	if len(goal.References)+len(context.Sorts)+len(context.Signatures) > algebra.DefaultMaxSymbols {
		return nil, goalResource("references", "combined declaration count exceeds the bounded profile")
	}

	portIndex := make(map[string]int, len(goal.Ports))
	for i, port := range goal.Ports {
		path := fmt.Sprintf("ports[%d]", i)
		if !goalID(port.ID) {
			return nil, goalInvalid("INVALID_GOAL_PORT_ID", path+".id", "port ID must be non-empty UTF-8, whitespace-free, and at most 256 bytes")
		}
		if _, exists := portIndex[port.ID]; exists {
			return nil, goalInvalid("DUPLICATE_GOAL_PORT_ID", path+".id", "goal port ID is duplicated")
		}
		portIndex[port.ID] = i
	}
	projectedPorts := make(map[string]struct{}, len(goal.ProjectPorts))
	for i, portID := range goal.ProjectPorts {
		path := fmt.Sprintf("project_ports[%d]", i)
		if !goalID(portID) {
			return nil, goalInvalid("INVALID_PROJECT_PORT_ID", path, "projected port ID must be a non-empty bounded identifier")
		}
		portNo, exists := portIndex[portID]
		if !exists {
			return nil, goalInvalid("UNKNOWN_GOAL_PORT", path, "projected port is not declared by the goal")
		}
		if _, exists := projectedPorts[portID]; exists {
			return nil, goalInvalid("DUPLICATE_PROJECT_PORT", path, "projected port is duplicated")
		}
		if goal.Ports[portNo].Binding != nil {
			return nil, goalInvalid("BOUND_PROJECT_PORT", path, "a Ref or Hole bound port cannot also be projected")
		}
		projectedPorts[portID] = struct{}{}
	}

	// Validate all caller-provided types through the same bounded declaration
	// checker used by the expression algebra. Unique synthetic declaration IDs
	// avoid changing or conflating the caller's namespaces.
	typeDeclarations := make([]algebra.Declaration, 0, 2+len(goal.Ports)*2)
	addType := func(label string, typ algebra.Type) {
		typeDeclarations = append(typeDeclarations, algebra.Declaration{ID: fmt.Sprintf("goal-type-%s-%d", label, len(typeDeclarations)), Type: typ})
	}
	addType("expected", goal.ExpectedType)
	for _, port := range goal.Ports {
		addType("port", port.Type)
		if port.Binding != nil && port.Binding.Type != nil {
			addType("binding", *port.Binding.Type)
		}
	}
	if err := algebra.CheckDeclarations(algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Sorts: context.Sorts,
		Parameters: typeDeclarations, Limits: limits.TermLimits,
	}); err != nil {
		return nil, err
	}

	references := make(map[string]algebra.Declaration, len(goal.References))
	for i, ref := range goal.References {
		path := fmt.Sprintf("references[%d]", i)
		if !goalID(ref.ID) {
			return nil, goalInvalid("INVALID_GOAL_REFERENCE_ID", path+".id", "reference ID must be non-empty UTF-8, whitespace-free, and at most 256 bytes")
		}
		if _, exists := references[ref.ID]; exists {
			return nil, goalInvalid("DUPLICATE_GOAL_REFERENCE_ID", path+".id", "goal reference ID is duplicated")
		}
		references[ref.ID] = ref
	}
	combined := context
	combined.Parameters = nil
	combined.References = goal.References
	combined.Expression = nil
	combined.Limits = limits.TermLimits
	if err := algebra.CheckDeclarations(combined); err != nil {
		return nil, err
	}

	sourceIDs := make(map[string]struct{}, len(goal.Provenance))
	for i, source := range goal.Provenance {
		path := fmt.Sprintf("provenance[%d]", i)
		if err := source.Validate(); err != nil {
			return nil, goalInvalid("INVALID_GOAL_SOURCE", path, err.Error())
		}
		if _, exists := sourceIDs[source.ID]; exists {
			return nil, goalInvalid("DUPLICATE_GOAL_SOURCE_ID", path+".id", "source ID is duplicated")
		}
		sourceIDs[source.ID] = struct{}{}
	}

	operators := make(map[string]struct{}, len(goal.AllowedOperators))
	for i, op := range goal.AllowedOperators {
		if _, ok := allowedGoalOperators[op]; !ok {
			return nil, goalUnsupported("UNSUPPORTED_GOAL_OPERATOR", fmt.Sprintf("allowed_operators[%d]", i), "operator is not in the bounded goal profile")
		}
		if _, exists := operators[op]; exists {
			return nil, goalInvalid("DUPLICATE_GOAL_OPERATOR", fmt.Sprintf("allowed_operators[%d]", i), "operator is duplicated")
		}
		operators[op] = struct{}{}
	}

	// ExpectedType describes the declared output interface in port order.
	openPorts := make([]algebra.Type, 0, len(goal.Ports))
	for _, port := range goal.Ports {
		_, projected := projectedPorts[port.ID]
		if port.Binding == nil && !projected {
			openPorts = append(openPorts, port.Type)
		}
	}
	expected := algebra.Type{Kind: algebra.TypeProp}
	for i := len(openPorts) - 1; i >= 0; i-- {
		from := openPorts[i]
		prior := expected
		expected = algebra.Type{Kind: algebra.TypeArrow, From: &from, To: &prior}
	}
	if !algebra.TypeEqual(goal.ExpectedType, expected) {
		return nil, goalInvalid("GOAL_EXPECTED_TYPE_MISMATCH", "expected_type", "expected type must be the ordered goal-port interface returning Prop")
	}

	holes := make(map[string]struct{})
	refBindingCount := 0
	for i, port := range goal.Ports {
		if port.Binding == nil {
			continue
		}
		binding := port.Binding
		path := fmt.Sprintf("ports[%d].binding", i)
		if binding.ID == "" || len(binding.ID) > 256 || !utf8.ValidString(binding.ID) {
			return nil, goalInvalid("INVALID_GOAL_BINDING_ID", path+".id", "binding ID must be a non-empty bounded UTF-8 identifier")
		}
		if binding.Type == nil || !algebra.TypeEqual(*binding.Type, port.Type) {
			return nil, goalInvalid("GOAL_BINDING_TYPE_MISMATCH", path+".type", "binding type must exactly match its goal port")
		}
		switch binding.Kind {
		case algebra.TermRef:
			ref, exists := references[binding.ID]
			if !exists {
				return nil, goalInvalid("UNDECLARED_GOAL_REFERENCE", path+".id", "reference binding must name a declared goal reference")
			}
			if !algebra.TypeEqual(ref.Type, port.Type) {
				return nil, goalInvalid("GOAL_BINDING_TYPE_MISMATCH", path+".type", "reference declaration type must match its goal port")
			}
			refBindingCount++
		case algebra.TermHole:
			if _, exists := holes[binding.ID]; exists {
				return nil, goalInvalid("DUPLICATE_GOAL_HOLE_ID", path+".id", "hole ID is duplicated")
			}
			holes[binding.ID] = struct{}{}
		default:
			return nil, goalInvalid("INVALID_GOAL_BINDING_KIND", path+".kind", "only explicit Ref and typed Hole bindings are supported")
		}
	}

	pathCount := 0
	unaryCount := 0
	pathCounts := make([]int, 0, len(goal.Constraints))
	var unresolved []Unresolved
	constraintIDs := make(map[string]struct{}, len(goal.Constraints))
	for i, constraint := range goal.Constraints {
		path := fmt.Sprintf("constraints[%d]", i)
		if !goalID(constraint.ID) {
			return nil, goalInvalid("INVALID_CONSTRAINT_ID", path+".id", "constraint ID must be non-empty UTF-8, whitespace-free, and at most 256 bytes")
		}
		if _, exists := constraintIDs[constraint.ID]; exists {
			return nil, goalInvalid("DUPLICATE_CONSTRAINT_ID", path+".id", "constraint ID is duplicated")
		}
		constraintIDs[constraint.ID] = struct{}{}
		switch constraint.Kind {
		case ConstraintPath:
			pathCount++
			if constraint.Port != "" {
				return nil, goalInvalid("IRRELEVANT_CONSTRAINT_FIELD", path+".port", "path constraint cannot set unary port")
			}
			if len(goal.Ports) != 2 {
				return nil, goalUnsupported("UNSUPPORTED_PATH_INTERFACE", path, "the initial path profile requires exactly two goal ports")
			}
			if constraint.FromPort != "" {
				if _, ok := portIndex[constraint.FromPort]; !ok {
					return nil, goalInvalid("UNKNOWN_GOAL_PORT", path+".from_port", "path endpoint is not a declared goal port")
				}
			}
			if constraint.ToPort != "" {
				if _, ok := portIndex[constraint.ToPort]; !ok {
					return nil, goalInvalid("UNKNOWN_GOAL_PORT", path+".to_port", "path endpoint is not a declared goal port")
				}
			}
			if constraint.FromPort != "" && constraint.ToPort != "" && constraint.FromPort == constraint.ToPort {
				return nil, goalUnsupported("UNSUPPORTED_PATH_CYCLE", path, "a path cannot connect a goal port to itself")
			}
			if constraint.FromRole != "" && constraint.ToRole != "" && constraint.FromRole == constraint.ToRole {
				return nil, goalInvalid("CONFLICTING_ROLE_MAPPING", path, "one relation role cannot map to both path endpoints")
			}
			if constraint.Symbol == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_SYMBOL", Path: path + ".symbol", Message: "select a declared binary relation symbol"})
			} else {
				sig, found := goalSignature(context.Signatures, constraint.Symbol)
				if !found {
					return nil, goalInvalid("UNKNOWN_RELATION_SYMBOL", path+".symbol", "relation symbol is not declared in the selected context")
				}
				if len(sig.Ports) != 2 {
					return nil, goalUnsupported("UNSUPPORTED_PATH_SIGNATURE", path+".symbol", "path relation must have exactly two declared ports")
				}
				if constraint.FromRole != "" {
					if _, ok := signatureRole(sig, constraint.FromRole); !ok {
						return nil, goalInvalid("UNKNOWN_RELATION_ROLE", path+".from_role", "role is not declared by the relation")
					}
				}
				if constraint.ToRole != "" {
					if _, ok := signatureRole(sig, constraint.ToRole); !ok {
						return nil, goalInvalid("UNKNOWN_RELATION_ROLE", path+".to_role", "role is not declared by the relation")
					}
				}
				if constraint.FromRole != "" && constraint.FromPort != "" {
					from, _ := signatureRole(sig, constraint.FromRole)
					fromPort, _ := portIndex[constraint.FromPort]
					if !forSearch && !algebra.TypeEqual(goal.Ports[fromPort].Type, sig.Ports[from].Type) {
						return nil, goalInvalid("GOAL_ROLE_TYPE_MISMATCH", path+".from_role", "source goal port type must match its selected relation role")
					}
				}
				if constraint.ToRole != "" && constraint.ToPort != "" {
					to, _ := signatureRole(sig, constraint.ToRole)
					toPort, _ := portIndex[constraint.ToPort]
					if !forSearch && !algebra.TypeEqual(goal.Ports[toPort].Type, sig.Ports[to].Type) {
						return nil, goalInvalid("GOAL_ROLE_TYPE_MISMATCH", path+".to_role", "target goal port type must match its selected relation role")
					}
				}
				if constraint.FromRole != "" && constraint.ToRole != "" {
					from, _ := signatureRole(sig, constraint.FromRole)
					to, _ := signatureRole(sig, constraint.ToRole)
					if constraint.FromRole == constraint.ToRole || from == to {
						return nil, goalInvalid("CONFLICTING_ROLE_MAPPING", path, "path endpoints must map to distinct relation roles")
					}
					if !forSearch && len(constraint.AllowedAtomCounts) == 1 && constraint.AllowedAtomCounts[0] > 1 && !algebra.TypeEqual(sig.Ports[from].Type, sig.Ports[to].Type) {
						return nil, goalInvalid("PATH_LINK_TYPE_MISMATCH", path, "repeated path relations require matching types at the joined roles")
					}
				}
			}
			if constraint.FromPort == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_FROM_PORT", Path: path + ".from_port", Message: "select the source goal port"})
			}
			if constraint.ToPort == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_TO_PORT", Path: path + ".to_port", Message: "select the target goal port"})
			}
			if constraint.FromRole == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_FROM_ROLE", Path: path + ".from_role", Message: "map the source endpoint to a relation role"})
			}
			if constraint.ToRole == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_TO_ROLE", Path: path + ".to_role", Message: "map the target endpoint to a relation role"})
			}
			if len(constraint.AllowedAtomCounts) == 0 {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_COUNT", Path: path + ".allowed_atom_counts", Message: "select an allowed path length"})
			} else {
				seenCounts := make(map[int]struct{}, len(constraint.AllowedAtomCounts))
				for j, count := range constraint.AllowedAtomCounts {
					countPath := fmt.Sprintf("%s.allowed_atom_counts[%d]", path, j)
					if count <= 0 {
						return nil, goalInvalid("INVALID_ATOM_COUNT", countPath, "path length must be positive")
					}
					if count > MaxGoalAtoms {
						return nil, goalResource(countPath, "path length exceeds the profile atom cap")
					}
					if !forSearch && count > limits.MaxAtoms {
						return nil, goalResource(countPath, "path length exceeds the requested atom budget")
					}
					if _, exists := seenCounts[count]; exists {
						return nil, goalInvalid("DUPLICATE_ATOM_COUNT", countPath, "path length choice is duplicated")
					}
					seenCounts[count] = struct{}{}
					pathCounts = append(pathCounts, count)
				}
				if len(seenCounts) > 1 {
					unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "UNDERDETERMINED_COUNT", Path: path + ".allowed_atom_counts", Message: "choose one path length; no choice is selected automatically"})
				}
			}
		case ConstraintUnary:
			unaryCount++
			if constraint.FromPort != "" || constraint.ToPort != "" || constraint.FromRole != "" || constraint.ToRole != "" || len(constraint.AllowedAtomCounts) != 0 {
				return nil, goalInvalid("IRRELEVANT_CONSTRAINT_FIELD", path, "unary constraint contains path-only fields")
			}
			var unarySignature algebra.Signature
			if constraint.Symbol == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_SYMBOL", Path: path + ".symbol", Message: "select a declared unary relation symbol"})
			} else {
				var found bool
				unarySignature, found = goalSignature(context.Signatures, constraint.Symbol)
				if !found {
					return nil, goalInvalid("UNKNOWN_RELATION_SYMBOL", path+".symbol", "relation symbol is not declared in the selected context")
				}
				if len(unarySignature.Ports) != 1 {
					return nil, goalUnsupported("UNSUPPORTED_UNARY_SIGNATURE", path+".symbol", "unary restriction requires a one-port relation")
				}
			}
			if constraint.Port == "" {
				unresolved = append(unresolved, Unresolved{ConstraintID: constraint.ID, Code: "MISSING_PORT", Path: path + ".port", Message: "select the restricted goal port"})
			} else {
				portNo, ok := portIndex[constraint.Port]
				if !ok {
					return nil, goalInvalid("UNKNOWN_GOAL_PORT", path+".port", "unary restriction port is not declared")
				}
				if !forSearch && constraint.Symbol != "" && !algebra.TypeEqual(goal.Ports[portNo].Type, unarySignature.Ports[0].Type) {
					return nil, goalInvalid("GOAL_ROLE_TYPE_MISMATCH", path, "goal port type must match the unary relation port")
				}
			}
		default:
			return nil, goalUnsupported("UNSUPPORTED_CONSTRAINT_KIND", path+".kind", "constraint kind is outside the bounded goal profile")
		}
	}
	if pathCount == 0 {
		unresolved = append(unresolved, Unresolved{Code: "MISSING_PATH", Path: "constraints", Message: "provide one binary path constraint"})
	} else if pathCount > 1 {
		return nil, goalUnsupported("MULTIPLE_PATHS", "constraints", "the initial profile supports one path constraint")
	}
	if pathCount == 1 && len(goal.Ports) != 2 {
		return nil, goalUnsupported("UNSUPPORTED_PATH_INTERFACE", "ports", "a binary path requires exactly two goal ports")
	}
	plan := deriveGoalPlan(context.Signatures, refBindingCount, goal.Ports, goal.ProjectPorts, goal.Constraints)
	if !forSearch {
		if err := validateGoalOperatorNeeds(operators, plan.requiredOperators); err != nil {
			return nil, err
		}
		if plan.maxSteps > limits.MaxSteps {
			return nil, goalResource("limits.max_steps", "the requested goal exceeds the bounded operator-step budget")
		}
		if _, err := goalAtomUpperBound(pathCounts, unaryCount, limits.MaxAtoms); err != nil {
			return nil, err
		}
	}
	for id := range holes {
		for i := range goal.Ports {
			binding := goal.Ports[i].Binding
			if binding != nil && binding.Kind == algebra.TermHole && binding.ID == id {
				holeType := *binding.Type
				unresolved = append(unresolved, Unresolved{Code: "UNRESOLVED_HOLE", Path: fmt.Sprintf("ports[%d].binding", i), Message: "a typed hole must be resolved before lowering", Hole: &holeType})
			}
		}
	}
	return unresolved, nil
}

var allowedGoalOperators = map[string]struct{}{
	GoalOperatorRelation: {}, GoalOperatorJoin: {}, GoalOperatorRestrict: {},
	GoalOperatorBind: {}, GoalOperatorPermute: {}, GoalOperatorProject: {},
}

func goalID(value string) bool {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r == 0 || r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			return false
		}
	}
	return true
}

func goalSignature(signatures []algebra.Signature, id string) (algebra.Signature, bool) {
	for _, signature := range signatures {
		if signature.ID == id {
			return signature, true
		}
	}
	return algebra.Signature{}, false
}

func signatureRole(signature algebra.Signature, role string) (int, bool) {
	for i, port := range signature.Ports {
		if port.Role == role {
			return i, true
		}
	}
	return 0, false
}

type goalPlan struct {
	requiredOperators map[string]struct{}
	maxSteps          int
}

func deriveGoalPlan(signatures []algebra.Signature, refBindings int, ports []GoalPort, projectPorts []string, constraints []Constraint) goalPlan {
	needs := make(map[string]struct{})
	plan := goalPlan{requiredOperators: needs}
	for _, constraint := range constraints {
		if constraint.Kind != ConstraintPath || constraint.Symbol == "" || constraint.FromPort == "" || constraint.ToPort == "" || constraint.FromRole == "" || constraint.ToRole == "" || len(constraint.AllowedAtomCounts) == 0 {
			continue
		}
		signature, found := goalSignature(signatures, constraint.Symbol)
		if !found {
			continue
		}
		_, foundFrom := signatureRole(signature, constraint.FromRole)
		_, foundTo := signatureRole(signature, constraint.ToRole)
		if !foundFrom || !foundTo {
			continue
		}
		needs[GoalOperatorRelation] = struct{}{}
		for _, count := range constraint.AllowedAtomCounts {
			steps := 2*count - 1
			if count > 1 {
				needs[GoalOperatorJoin] = struct{}{}
			}
			if len(ports) == 2 && len(projectPorts) == 0 && ports[0].Binding == nil && ports[1].Binding == nil {
				var firstOutput string
				switch count {
				case 1:
					if signature.Ports[0].Role == constraint.FromRole {
						firstOutput = constraint.FromPort
					} else if signature.Ports[0].Role == constraint.ToRole {
						firstOutput = constraint.ToPort
					}
				default:
					firstOutput = constraint.FromPort
				}
				if firstOutput != "" && firstOutput != ports[0].ID {
					needs[GoalOperatorPermute] = struct{}{}
					steps++
				}
			}
			if steps > plan.maxSteps {
				plan.maxSteps = steps
			}
		}
		break
	}
	for _, constraint := range constraints {
		if constraint.Kind == ConstraintUnary && constraint.Symbol != "" && constraint.Port != "" {
			needs[GoalOperatorRestrict] = struct{}{}
			plan.maxSteps += 2
		}
	}
	if refBindings > 0 {
		needs[GoalOperatorBind] = struct{}{}
		plan.maxSteps += refBindings
	}
	if len(projectPorts) > 0 {
		needs[GoalOperatorProject] = struct{}{}
		plan.maxSteps++
	}
	return plan
}

func validateGoalOperatorNeeds(available map[string]struct{}, needs map[string]struct{}) error {
	for name := range needs {
		if _, ok := available[name]; !ok {
			return goalUnsupported("OPERATOR_NOT_ALLOWED", "allowed_operators", "goal requires an operator that is not allowed: "+name)
		}
	}
	return nil
}

func goalAtomUpperBound(counts []int, unaryCount, limit int) (int, error) {
	maxPath := 1
	for _, count := range counts {
		if count > maxPath {
			maxPath = count
		}
	}
	if maxPath+unaryCount > limit {
		return 0, goalResource("limits.max_atoms", "path occurrences and unary constraints exceed the atom budget")
	}
	return maxPath + unaryCount, nil
}

func goalBoundError(path string, count, cap int) error {
	if count > cap {
		return goalResource(path, fmt.Sprintf("count %d exceeds limit %d", count, cap))
	}
	return goalInvalid("INVALID_GOAL_COUNT", path, "at least one item is required")
}

func goalInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

func goalUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}

func goalResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}

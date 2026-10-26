package synthesis

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const SearchSpaceProfile = "finite-path-search-v1"

const (
	SearchStopMaxPathLength  SearchStopReason = "max_path_length"
	SearchStopMaxBranchDepth SearchStopReason = "max_branch_depth"
)

// ChoiceAssignment records one explicit value selected from one search axis.
// It records syntactic choices only; it is not evidence of truth or intent.
type ChoiceAssignment struct {
	Kind             string `json:"kind"`
	AxisID           string `json:"axis_id"`
	ConstraintID     string `json:"constraint_id,omitempty"`
	Symbol           string `json:"symbol,omitempty"`
	FromRole         string `json:"from_role,omitempty"`
	ToRole           string `json:"to_role,omitempty"`
	AtomCount        int    `json:"atom_count,omitempty"`
	PortID           string `json:"port_id,omitempty"`
	UnaryRole        string `json:"unary_role,omitempty"`
	HoleID           string `json:"hole_id,omitempty"`
	ReferenceID      string `json:"reference_id,omitempty"`
	OutsidePathBound bool   `json:"outside_path_bound,omitempty"`
}

// SearchAxisDescriptor identifies one ordered choice axis and its finite
// cardinality. Axis order is the mixed-radix cursor order.
type SearchAxisDescriptor struct {
	AxisID       string `json:"axis_id"`
	Kind         string `json:"kind"`
	ConstraintID string `json:"constraint_id,omitempty"`
	PortID       string `json:"port_id,omitempty"`
	OptionCount  int    `json:"option_count"`
}

// SearchDomainRejection explains an empty candidate-choice domain. It only
// reports structural/type fit against declarations, not world meaning.
type SearchDomainRejection struct {
	AxisID  string `json:"axis_id"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SearchSpaceDescriptor reports the finite declared search dimensions. It
// does not claim that any candidates were explored or that the goal is true.
type SearchSpaceDescriptor struct {
	Profile                     string                  `json:"profile"`
	PathPolicy                  string                  `json:"path_policy"`
	PathLengthChoices           []int                   `json:"path_length_choices"`
	MaxPathLength               int                     `json:"max_path_length"`
	ExcludedExplicitPathLengths []int                   `json:"excluded_explicit_path_lengths"`
	AxisCount                   int                     `json:"axis_count"`
	BranchDepth                 int                     `json:"branch_depth"`
	BranchDepthMetric           string                  `json:"branch_depth_metric"`
	MaxBranchDepth              int                     `json:"max_branch_depth"`
	Axes                        []SearchAxisDescriptor  `json:"axes"`
	SourceConstraintIDs         []string                `json:"source_constraint_ids"`
	InputUnresolved             []Unresolved            `json:"input_unresolved"`
	DomainRejections            []SearchDomainRejection `json:"domain_rejections"`
	BlockedByUnresolved         bool                    `json:"blocked_by_unresolved,omitempty"`
	Ready                       bool                    `json:"ready"`
	DeclaredConfigurations      string                  `json:"declared_configurations,omitempty"`
	ConfigurationCountExact     bool                    `json:"configuration_count_exact"`
	StopReason                  SearchStopReason        `json:"stop_reason,omitempty"`
}

type candidateAxis struct {
	descriptor   SearchAxisDescriptor
	options      []ChoiceAssignment
	unarySymbols []unaryChoice
	unaryPorts   []goalPortChoice
	unaryCount   int
}

type unaryChoice struct {
	symbol string
	role   string
	type_  algebra.Type
}

type goalPortChoice struct {
	id    string
	type_ algebra.Type
}

func (axis candidateAxis) optionCount() int {
	if axis.descriptor.Kind == "unary_pair" {
		return axis.unaryCount
	}
	return len(axis.options)
}

func (axis candidateAxis) optionAt(index int, budget *searchBudget) (ChoiceAssignment, bool) {
	if index < 0 || index >= axis.optionCount() || budget == nil {
		return ChoiceAssignment{}, false
	}
	if axis.descriptor.Kind != "unary_pair" {
		return axis.options[index], true
	}
	seen := 0
	for _, symbol := range axis.unarySymbols {
		for _, port := range axis.unaryPorts {
			ok, err := budget.consumeWork(1)
			if err != nil || !ok {
				return ChoiceAssignment{}, false
			}
			if !algebra.TypeEqual(symbol.type_, port.type_) {
				continue
			}
			if seen == index {
				return ChoiceAssignment{
					Kind: "unary", AxisID: axis.descriptor.AxisID,
					ConstraintID: axis.descriptor.ConstraintID, Symbol: symbol.symbol,
					PortID: port.id, UnaryRole: symbol.role,
				}, true
			}
			seen++
		}
	}
	return ChoiceAssignment{}, false
}

// finiteCandidateSpace stores only individual axes and one cursor position per
// axis. It never allocates the Cartesian product of all candidate choices.
type finiteCandidateSpace struct {
	goal       GoalSpec
	axes       []candidateAxis
	positions  []int
	budget     *searchBudget
	descriptor SearchSpaceDescriptor
	blocked    bool
	done       bool
	stopReason SearchStopReason
}

// newFiniteCandidateSpace validates the formal goal before deriving axes.
// The caller supplies a signature index built from the same selected,
// approved-primitive declaration context.
func newFiniteCandidateSpace(context algebra.Document, index *SignatureIndex, goal GoalSpec, budget *searchBudget) (*finiteCandidateSpace, error) {
	if budget == nil {
		return nil, searchSpaceInvalid("INVALID_SEARCH_BUDGET", "budget", "a normalized request-local search budget is required")
	}
	unresolved, err := validateSearchGoal(context, goal)
	if err != nil {
		return nil, err
	}
	if index == nil {
		return nil, searchSpaceInvalid("MISSING_SIGNATURE_INDEX", "context.signatures", "a validated signature index is required")
	}
	space := &finiteCandidateSpace{
		budget: budget,
		descriptor: SearchSpaceDescriptor{
			Profile:           SearchSpaceProfile,
			PathPolicy:        "explicit_counts",
			PathLengthChoices: []int{}, ExcludedExplicitPathLengths: []int{},
			Axes: []SearchAxisDescriptor{}, SourceConstraintIDs: []string{},
			InputUnresolved: []Unresolved{}, DomainRejections: []SearchDomainRejection{},
			BranchDepthMetric: "one layer per ordered choice axis; singleton axes count",
			MaxPathLength:     budget.limits.MaxPathLength, MaxBranchDepth: budget.limits.MaxBranchDepth,
		},
	}

	if !budget.checkDeadlineCancel() {
		space.markStopped(budget.stop())
		return space, nil
	}
	ownedGoal, complete := cloneGoalForSearch(goal, budget)
	if !complete {
		space.markStopped(budget.stop())
		return space, nil
	}
	space.goal = ownedGoal
	space.descriptor.InputUnresolved, complete = cloneUnresolvedForSearch(unresolved, budget)
	if !complete {
		space.markStopped(budget.stop())
		return space, nil
	}
	for _, constraint := range space.goal.Constraints {
		space.descriptor.SourceConstraintIDs = append(space.descriptor.SourceConstraintIDs, constraint.ID)
	}

	pathIndex := -1
	for i, constraint := range space.goal.Constraints {
		if constraint.Kind == ConstraintPath {
			pathIndex = i
			break
		}
	}
	if pathIndex < 0 {
		space.blocked = true
		space.descriptor.BlockedByUnresolved = true
		return space, nil
	}
	path := space.goal.Constraints[pathIndex]
	fromPort, fromFound := searchGoalPort(space.goal.Ports, path.FromPort)
	toPort, toFound := searchGoalPort(space.goal.Ports, path.ToPort)
	if !fromFound || !toFound {
		// Endpoint identity is not inferred from argument order or equal types.
		space.blocked = true
		space.descriptor.BlockedByUnresolved = true
		return space, nil
	}

	pathAxisID := "axis-1"
	if !space.chargeWork(len(context.Signatures)) {
		return space, nil
	}
	binaryChoices := index.BinaryChoices(fromPort.Type, toPort.Type, path.FromRole, path.ToRole)
	pathOptions := make([]ChoiceAssignment, 0, len(binaryChoices))
	for _, choice := range binaryChoices {
		if path.Symbol != "" && choice.Symbol != path.Symbol {
			continue
		}
		if !space.chargeWork(1) {
			return space, nil
		}
		pathOptions = append(pathOptions, ChoiceAssignment{
			Kind: "path_signature", AxisID: pathAxisID, ConstraintID: path.ID,
			Symbol: choice.Symbol, FromRole: choice.FromRole, ToRole: choice.ToRole,
			PortID: path.FromPort,
		})
	}
	space.addAxis(candidateAxis{descriptor: SearchAxisDescriptor{
		AxisID: pathAxisID, Kind: "binary_signature", ConstraintID: path.ID,
		OptionCount: len(pathOptions),
	}, options: pathOptions})
	if len(pathOptions) == 0 {
		space.descriptor.DomainRejections = append(space.descriptor.DomainRejections, SearchDomainRejection{
			AxisID: pathAxisID, Code: "NO_BINARY_TYPE_FIT",
			Message: "no declared binary signature and role assignment matches the requested endpoint types and filters",
		})
	}

	pathLengths, explicit := searchPathLengths(path.AllowedAtomCounts, budget.limits.MaxPathLength)
	if explicit {
		space.descriptor.PathPolicy = "explicit_counts"
	} else {
		space.descriptor.PathPolicy = "default_range"
	}
	space.descriptor.PathLengthChoices = append([]int{}, pathLengths...)
	lengthOptions := make([]ChoiceAssignment, 0, len(pathLengths))
	for _, count := range pathLengths {
		if !space.chargeWork(1) {
			return space, nil
		}
		lengthOptions = append(lengthOptions, ChoiceAssignment{
			Kind: "path_length", AxisID: "axis-2", ConstraintID: path.ID,
			AtomCount: count, OutsidePathBound: count > budget.limits.MaxPathLength,
		})
		if explicit && count > budget.limits.MaxPathLength {
			space.descriptor.ExcludedExplicitPathLengths = append(space.descriptor.ExcludedExplicitPathLengths, count)
		}
	}
	space.addAxis(candidateAxis{descriptor: SearchAxisDescriptor{
		AxisID: "axis-2", Kind: "path_length", ConstraintID: path.ID,
		OptionCount: len(lengthOptions),
	}, options: lengthOptions})

	unarySymbols, complete := space.collectUnarySymbols(context.Signatures, index, space.goal.Ports)
	if !complete {
		return space, nil
	}
	axisNumber := 3
	for _, constraint := range space.goal.Constraints {
		if constraint.Kind != ConstraintUnary {
			continue
		}
		axisID := fmt.Sprintf("axis-%d", axisNumber)
		axisNumber++
		axis, rejection, complete := space.makeUnaryAxis(axisID, constraint, unarySymbols, space.goal.Ports)
		if !complete {
			return space, nil
		}
		space.addAxis(axis)
		if rejection != nil {
			space.descriptor.DomainRejections = append(space.descriptor.DomainRejections, *rejection)
		}
	}

	refs := append([]algebra.Declaration{}, space.goal.References...)
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })
	for i, port := range space.goal.Ports {
		if port.Binding == nil || port.Binding.Kind != algebra.TermHole {
			continue
		}
		axisID := fmt.Sprintf("axis-%d", axisNumber)
		axisNumber++
		options := make([]ChoiceAssignment, 0)
		for _, ref := range refs {
			if !space.chargeWork(1) {
				return space, nil
			}
			if algebra.TypeEqual(*port.Binding.Type, ref.Type) {
				options = append(options, ChoiceAssignment{
					Kind: "hole_binding", AxisID: axisID, PortID: port.ID,
					HoleID: port.Binding.ID, ReferenceID: ref.ID,
				})
			}
		}
		space.addAxis(candidateAxis{descriptor: SearchAxisDescriptor{
			AxisID: axisID, Kind: "hole_reference", PortID: port.ID,
			OptionCount: len(options),
		}, options: options})
		if len(options) == 0 {
			space.descriptor.DomainRejections = append(space.descriptor.DomainRejections, SearchDomainRejection{
				AxisID: axisID, Code: "NO_REFERENCE_FOR_HOLE_TYPE",
				Message: fmt.Sprintf("no declared request reference has the exact type required by hole at ports[%d]", i),
			})
		}
	}

	space.descriptor.AxisCount = len(space.axes)
	space.descriptor.BranchDepth = len(space.axes)
	space.positions = make([]int, len(space.axes))
	space.descriptor.Ready = true
	space.calculateConfigurationCount()
	if space.descriptor.BranchDepth > budget.limits.MaxBranchDepth {
		space.markStopped(SearchStopMaxBranchDepth)
	}
	return space, nil
}

// hasNext reports whether a complete declared configuration remains. It does
// not reserve an expansion or candidate slot; the search engine does that.
func (space *finiteCandidateSpace) hasNext() bool {
	if space == nil || space.done || space.blocked || space.stopReason != "" || space.budget == nil {
		return false
	}
	if space.budget.stop() != "" || !space.budget.checkDeadlineCancel() {
		space.markStopped(space.budget.stop())
		return false
	}
	for _, axis := range space.axes {
		if axis.optionCount() == 0 {
			return false
		}
	}
	return true
}

// next lazily emits one detached concrete GoalSpec and its assignments. It
// does not charge a candidate or expansion; those reservations belong to the
// search engine before it invokes this cursor.
func (space *finiteCandidateSpace) next() (GoalSpec, []ChoiceAssignment, bool, error) {
	if !space.hasNext() {
		return GoalSpec{}, nil, false, nil
	}
	assignments := make([]ChoiceAssignment, len(space.axes))
	for i, axis := range space.axes {
		choice, found := axis.optionAt(space.positions[i], space.budget)
		if !found {
			if space.budget.stop() != "" {
				space.markStopped(space.budget.stop())
				return GoalSpec{}, nil, false, nil
			}
			return GoalSpec{}, nil, false, searchSpaceInvalid("INVALID_SEARCH_AXIS", axis.descriptor.AxisID, "axis cursor selected an unavailable option")
		}
		assignments[i] = choice
	}
	candidate, complete := cloneGoalForSearch(space.goal, space.budget)
	if !complete {
		space.markStopped(space.budget.stop())
		return GoalSpec{}, nil, false, nil
	}
	if err := applyChoiceAssignments(&candidate, assignments, space.budget); err != nil {
		if space.budget.stop() != "" {
			space.markStopped(space.budget.stop())
			return GoalSpec{}, nil, false, nil
		}
		return GoalSpec{}, nil, false, err
	}
	space.advance()
	return candidate, assignments, true, nil
}

func (space *finiteCandidateSpace) descriptorSnapshot() SearchSpaceDescriptor {
	if space == nil {
		return SearchSpaceDescriptor{}
	}
	copyDescriptor := space.descriptor
	copyDescriptor.PathLengthChoices = append([]int{}, space.descriptor.PathLengthChoices...)
	copyDescriptor.ExcludedExplicitPathLengths = append([]int{}, space.descriptor.ExcludedExplicitPathLengths...)
	copyDescriptor.Axes = append([]SearchAxisDescriptor{}, space.descriptor.Axes...)
	copyDescriptor.SourceConstraintIDs = append([]string{}, space.descriptor.SourceConstraintIDs...)
	copyDescriptor.InputUnresolved = append([]Unresolved{}, space.descriptor.InputUnresolved...)
	for i := range copyDescriptor.InputUnresolved {
		if space.descriptor.InputUnresolved[i].Hole != nil {
			typ := cloneGoalType(*space.descriptor.InputUnresolved[i].Hole)
			copyDescriptor.InputUnresolved[i].Hole = &typ
		}
	}
	copyDescriptor.DomainRejections = append([]SearchDomainRejection{}, space.descriptor.DomainRejections...)
	if space.stopReason != "" {
		copyDescriptor.StopReason = space.stopReason
	}
	return copyDescriptor
}

func (space *finiteCandidateSpace) addAxis(axis candidateAxis) {
	space.axes = append(space.axes, axis)
	space.descriptor.Axes = append(space.descriptor.Axes, axis.descriptor)
	space.descriptor.AxisCount = len(space.axes)
	space.descriptor.BranchDepth = len(space.axes)
}

func (space *finiteCandidateSpace) chargeWork(units int) bool {
	ok, err := space.budget.consumeWork(units)
	if err != nil {
		space.markStopped(SearchStopMaxWorkUnits)
		return false
	}
	if !ok {
		space.markStopped(space.budget.stop())
		return false
	}
	return true
}

func (space *finiteCandidateSpace) markStopped(reason SearchStopReason) {
	if reason == "" {
		return
	}
	if space.stopReason == "" {
		space.stopReason = reason
	}
	if space.budget != nil && space.budget.stopReason == "" {
		space.budget.stopReason = reason
	}
	space.descriptor.StopReason = space.stopReason
	space.descriptor.Ready = false
}

func (space *finiteCandidateSpace) calculateConfigurationCount() {
	if space.blocked || !space.descriptor.Ready {
		space.descriptor.ConfigurationCountExact = false
		return
	}
	product := big.NewInt(1)
	for _, axis := range space.axes {
		product.Mul(product, big.NewInt(int64(axis.optionCount())))
	}
	space.descriptor.DeclaredConfigurations = product.String()
	space.descriptor.ConfigurationCountExact = true
}

func (space *finiteCandidateSpace) advance() {
	if len(space.positions) == 0 {
		space.done = true
		return
	}
	for i := len(space.positions) - 1; i >= 0; i-- {
		space.positions[i]++
		if space.positions[i] < space.axes[i].optionCount() {
			return
		}
		space.positions[i] = 0
	}
	space.done = true
}

func (space *finiteCandidateSpace) collectUnarySymbols(signatures []algebra.Signature, index *SignatureIndex, ports []GoalPort) ([]unaryChoice, bool) {
	seenTypes := make([]algebra.Type, 0, len(ports))
	bySymbol := make(map[string]unaryChoice)
	for _, port := range ports {
		known := false
		for _, seen := range seenTypes {
			if algebra.TypeEqual(seen, port.Type) {
				known = true
				break
			}
		}
		if known {
			continue
		}
		seenTypes = append(seenTypes, cloneGoalType(port.Type))
		if !space.chargeWork(len(signatures)) {
			return nil, false
		}
		choices := index.UnaryChoices(port.Type)
		for _, choice := range choices {
			if _, exists := bySymbol[choice.Symbol]; exists {
				continue
			}
			if !space.chargeWork(1) {
				return nil, false
			}
			signature, found := index.Lookup(choice.Symbol)
			if !found || len(signature.Ports) != 1 || signature.Ports[0].Role != choice.UnaryRole {
				return nil, false
			}
			bySymbol[choice.Symbol] = unaryChoice{symbol: choice.Symbol, role: choice.UnaryRole, type_: signature.Ports[0].Type}
		}
	}
	result := make([]unaryChoice, 0, len(bySymbol))
	for _, choice := range bySymbol {
		result = append(result, choice)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].symbol != result[j].symbol {
			return result[i].symbol < result[j].symbol
		}
		return result[i].role < result[j].role
	})
	return result, true
}

func (space *finiteCandidateSpace) makeUnaryAxis(axisID string, constraint Constraint, symbols []unaryChoice, ports []GoalPort) (candidateAxis, *SearchDomainRejection, bool) {
	filteredSymbols := make([]unaryChoice, 0, len(symbols))
	for _, symbol := range symbols {
		if constraint.Symbol == "" || constraint.Symbol == symbol.symbol {
			filteredSymbols = append(filteredSymbols, symbol)
		}
	}
	filteredPorts := make([]goalPortChoice, 0, len(ports))
	for _, port := range ports {
		if constraint.Port != "" && constraint.Port != port.ID {
			continue
		}
		portType, complete := cloneSearchType(port.Type, space.budget)
		if !complete {
			space.markStopped(space.budget.stop())
			return candidateAxis{}, nil, false
		}
		filteredPorts = append(filteredPorts, goalPortChoice{id: port.ID, type_: portType})
	}
	sort.Slice(filteredPorts, func(i, j int) bool { return filteredPorts[i].id < filteredPorts[j].id })
	axis := candidateAxis{
		descriptor:   SearchAxisDescriptor{AxisID: axisID, Kind: "unary_pair", ConstraintID: constraint.ID},
		unarySymbols: filteredSymbols, unaryPorts: filteredPorts,
	}
	for _, symbol := range filteredSymbols {
		for _, port := range filteredPorts {
			if !space.chargeWork(1) {
				return candidateAxis{}, nil, false
			}
			if algebra.TypeEqual(symbol.type_, port.type_) {
				axis.unaryCount++
			}
		}
	}
	axis.descriptor.OptionCount = axis.unaryCount
	if axis.unaryCount == 0 {
		return axis, &SearchDomainRejection{
			AxisID: axisID, Code: "NO_UNARY_TYPE_FIT",
			Message: "no declared unary signature and goal-port pairing has an exact type match",
		}, true
	}
	return axis, nil, true
}

func searchPathLengths(explicit []int, maximum int) ([]int, bool) {
	if len(explicit) != 0 {
		result := append([]int{}, explicit...)
		sort.Ints(result)
		return result, true
	}
	result := make([]int, maximum)
	for i := range result {
		result[i] = i + 1
	}
	return result, false
}

func searchGoalPort(ports []GoalPort, id string) (GoalPort, bool) {
	for _, port := range ports {
		if port.ID == id {
			return port, true
		}
	}
	return GoalPort{}, false
}

func cloneGoalForSearch(source GoalSpec, budget *searchBudget) (GoalSpec, bool) {
	copyGoal := source
	var complete bool
	copyGoal.ExpectedType, complete = cloneSearchType(source.ExpectedType, budget)
	if !complete {
		return GoalSpec{}, false
	}
	copyGoal.Ports = make([]GoalPort, len(source.Ports))
	for i, port := range source.Ports {
		if !chargeSearchCopy(budget) {
			return GoalSpec{}, false
		}
		copyPort := port
		copyPort.Type, complete = cloneSearchType(port.Type, budget)
		if !complete {
			return GoalSpec{}, false
		}
		if port.Binding != nil {
			binding := *port.Binding
			if port.Binding.Type != nil {
				bindingType, cloneOK := cloneSearchType(*port.Binding.Type, budget)
				if !cloneOK {
					return GoalSpec{}, false
				}
				binding.Type = &bindingType
			}
			copyPort.Binding = &binding
		}
		copyGoal.Ports[i] = copyPort
	}
	copyGoal.References = make([]algebra.Declaration, len(source.References))
	for i, ref := range source.References {
		if !chargeSearchCopy(budget) {
			return GoalSpec{}, false
		}
		typ, cloneOK := cloneSearchType(ref.Type, budget)
		if !cloneOK {
			return GoalSpec{}, false
		}
		copyGoal.References[i] = ref
		copyGoal.References[i].Type = typ
	}
	copyGoal.ProjectPorts, complete = cloneSearchStrings(source.ProjectPorts, budget)
	if !complete {
		return GoalSpec{}, false
	}
	copyGoal.AllowedOperators, complete = cloneSearchStrings(source.AllowedOperators, budget)
	if !complete {
		return GoalSpec{}, false
	}
	copyGoal.Constraints = make([]Constraint, len(source.Constraints))
	for i, constraint := range source.Constraints {
		if !chargeSearchCopy(budget) {
			return GoalSpec{}, false
		}
		copyConstraint := constraint
		copyConstraint.AllowedAtomCounts = make([]int, len(constraint.AllowedAtomCounts))
		for j, count := range constraint.AllowedAtomCounts {
			if !chargeSearchCopy(budget) {
				return GoalSpec{}, false
			}
			copyConstraint.AllowedAtomCounts[j] = count
		}
		copyGoal.Constraints[i] = copyConstraint
	}
	copyGoal.Provenance = make([]provenance.SourceRef, len(source.Provenance))
	for i, item := range source.Provenance {
		if !chargeSearchCopy(budget) {
			return GoalSpec{}, false
		}
		copyGoal.Provenance[i] = item
	}
	if source.RequestContext != nil {
		if !chargeSearchCopy(budget) {
			return GoalSpec{}, false
		}
		contextCopy, err := algebra.CloneRequestContext(*source.RequestContext)
		if err != nil {
			return GoalSpec{}, false
		}
		copyGoal.RequestContext = &contextCopy
	}
	return copyGoal, true
}

func cloneSearchStrings(source []string, budget *searchBudget) ([]string, bool) {
	copyStrings := make([]string, len(source))
	for i, value := range source {
		if !chargeSearchCopy(budget) {
			return nil, false
		}
		copyStrings[i] = value
	}
	return copyStrings, true
}

func cloneSearchType(source algebra.Type, budget *searchBudget) (algebra.Type, bool) {
	if !chargeSearchCopy(budget) {
		return algebra.Type{}, false
	}
	copyType := source
	if source.From != nil {
		from, complete := cloneSearchType(*source.From, budget)
		if !complete {
			return algebra.Type{}, false
		}
		copyType.From = &from
	}
	if source.To != nil {
		to, complete := cloneSearchType(*source.To, budget)
		if !complete {
			return algebra.Type{}, false
		}
		copyType.To = &to
	}
	return copyType, true
}

func cloneUnresolvedForSearch(source []Unresolved, budget *searchBudget) ([]Unresolved, bool) {
	result := make([]Unresolved, len(source))
	for i, item := range source {
		if !chargeSearchCopy(budget) {
			return nil, false
		}
		result[i] = item
		if item.Hole != nil {
			typ, complete := cloneSearchType(*item.Hole, budget)
			if !complete {
				return nil, false
			}
			result[i].Hole = &typ
		}
	}
	return result, true
}

func chargeSearchCopy(budget *searchBudget) bool {
	ok, err := budget.consumeWork(1)
	return err == nil && ok
}

func applyChoiceAssignments(goal *GoalSpec, assignments []ChoiceAssignment, budget *searchBudget) error {
	for _, assignment := range assignments {
		switch assignment.Kind {
		case "path_signature":
			constraint := goalConstraint(goal, assignment.ConstraintID, ConstraintPath)
			if constraint == nil {
				return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "path constraint disappeared from detached goal")
			}
			constraint.Symbol, constraint.FromRole, constraint.ToRole = assignment.Symbol, assignment.FromRole, assignment.ToRole
		case "path_length":
			constraint := goalConstraint(goal, assignment.ConstraintID, ConstraintPath)
			if constraint == nil {
				return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "path constraint disappeared from detached goal")
			}
			constraint.AllowedAtomCounts = []int{assignment.AtomCount}
		case "unary":
			constraint := goalConstraint(goal, assignment.ConstraintID, ConstraintUnary)
			if constraint == nil {
				return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "unary constraint disappeared from detached goal")
			}
			constraint.Symbol, constraint.Port = assignment.Symbol, assignment.PortID
		case "hole_binding":
			port := goalPortByID(goal.Ports, assignment.PortID)
			if port == nil || port.Binding == nil || port.Binding.Kind != algebra.TermHole || port.Binding.ID != assignment.HoleID {
				return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "typed hole disappeared from detached goal")
			}
			ref := goalReferenceByID(goal.References, assignment.ReferenceID)
			if ref == nil || !algebra.TypeEqual(ref.Type, *port.Binding.Type) {
				return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "hole assignment does not name an exact-type declared reference")
			}
			typ, complete := cloneSearchType(ref.Type, budget)
			if !complete {
				return searchSpaceInvalid("SEARCH_STOPPED", assignment.AxisID, "search budget stopped while copying a selected reference")
			}
			port.Binding = &GoalValue{Kind: algebra.TermRef, ID: ref.ID, Type: &typ}
		default:
			return searchSpaceInvalid("INVALID_SEARCH_ASSIGNMENT", assignment.AxisID, "assignment kind is not in the finite search profile")
		}
	}
	return nil
}

func goalConstraint(goal *GoalSpec, id string, kind ConstraintKind) *Constraint {
	if goal == nil {
		return nil
	}
	for i := range goal.Constraints {
		if goal.Constraints[i].ID == id && goal.Constraints[i].Kind == kind {
			return &goal.Constraints[i]
		}
	}
	return nil
}

func goalPortByID(ports []GoalPort, id string) *GoalPort {
	for i := range ports {
		if ports[i].ID == id {
			return &ports[i]
		}
	}
	return nil
}

func goalReferenceByID(references []algebra.Declaration, id string) *algebra.Declaration {
	for i := range references {
		if references[i].ID == id {
			return &references[i]
		}
	}
	return nil
}

func searchSpaceInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}

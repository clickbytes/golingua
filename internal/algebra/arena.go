package algebra

import (
	"fmt"
)

// NodeID identifies one syntax node inside an Arena. Zero is reserved to
// represent an absent edge; IDs in a valid arena are dense starting at one.
type NodeID uint32

// NodeArgument is an ordered role edge in an arena Atom.
type NodeArgument struct {
	Role  string `json:"role"`
	Value NodeID `json:"value"`
}

// Node is one detached syntax record. Identity carries the source BinderID,
// ParameterID, ReferenceID, or HoleID according to Kind; NodeID is never a
// substitute for that semantic identity.
type Node struct {
	NodeID     NodeID         `json:"node_id"`
	Kind       TermKind       `json:"kind"`
	Identity   string         `json:"identity,omitempty"`
	Label      string         `json:"label,omitempty"`
	Type       *Type          `json:"type,omitempty"`
	Symbol     string         `json:"symbol,omitempty"`
	Args       []NodeArgument `json:"args,omitempty"`
	Operands   []NodeID       `json:"operands,omitempty"`
	Body       NodeID         `json:"body,omitempty"`
	Function   NodeID         `json:"function,omitempty"`
	Argument   NodeID         `json:"argument,omitempty"`
	Quantifier string         `json:"quantifier,omitempty"`
	Count      *int           `json:"count,omitempty"`
	Restrictor NodeID         `json:"restrictor,omitempty"`
	Answers    []Declaration  `json:"answers,omitempty"`
}

// Arena is an immutable, checked, artifact-local DAG. Its fields remain
// private; all exported accessors return detached values.
type Arena struct {
	nodes   []Node
	root    NodeID
	context Document
	checked CheckedExpression
	limits  Limits
}

// NewArena validates and detaches one dense DAG against a declaration-only
// context. It does not infer symbol approval, metadata provenance, or facts.
func NewArena(context Document, root NodeID, nodes []Node) (*Arena, error) {
	if context.Expression != nil {
		return nil, invalid("ARENA_CONTEXT_HAS_EXPRESSION", "context.expression", "arena context must contain declarations only")
	}
	if err := CheckDeclarations(context); err != nil {
		return nil, err
	}
	limits, err := NormalizeLimits(context.Limits)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, invalid("EMPTY_ARENA", "nodes", "arena requires at least one syntax node")
	}
	if len(nodes) > limits.MaxNodes {
		return nil, resourceLimit("nodes", fmt.Sprintf("arena node count %d exceeds limit %d", len(nodes), limits.MaxNodes))
	}
	if root == 0 || uint64(root) > uint64(len(nodes)) {
		return nil, invalid("INVALID_ARENA_ROOT", "root", "root NodeID must identify a declared node")
	}

	copyNodes := make([]Node, len(nodes))
	for i, node := range nodes {
		want := NodeID(i + 1)
		if node.NodeID != want {
			return nil, invalid("NON_DENSE_NODE_IDS", fmt.Sprintf("nodes[%d].node_id", i), "NodeIDs must be unique and dense starting at one")
		}
		if node.Kind == TermAsk && len(node.Answers) > limits.MaxPorts {
			return nil, resourceLimit(fmt.Sprintf("nodes[%d].answers", i), fmt.Sprintf("answer count %d exceeds limit %d", len(node.Answers), limits.MaxPorts))
		}
		copyNodes[i] = node
		copyNodes[i].Count = nil
		if node.Count != nil {
			count := *node.Count
			copyNodes[i].Count = &count
		}
		if node.Kind == TermAsk {
			copyNodes[i].Answers = make([]Declaration, len(node.Answers))
			for j, answer := range node.Answers {
				typeCopy, typeErr := cloneStandaloneType(&answer.Type, limits)
				if typeErr != nil {
					return nil, typeErr
				}
				copyNodes[i].Answers[j] = Declaration{ID: answer.ID, Type: typeCopy}
			}
		}
		if node.Type != nil {
			typeCopy, typeErr := cloneStandaloneType(node.Type, limits)
			if typeErr != nil {
				return nil, typeErr
			}
			copyNodes[i].Type = &typeCopy
		}
		copyNodes[i].Args = append([]NodeArgument(nil), node.Args...)
		copyNodes[i].Operands = append([]NodeID(nil), node.Operands...)
	}

	children := make([][]NodeID, len(copyNodes)+1)
	totalEdges := 0
	for i, node := range copyNodes {
		path := fmt.Sprintf("nodes[%d]", i)
		if err := validateArenaNodeShape(node, path); err != nil {
			return nil, err
		}
		if len(node.Args) > limits.MaxNodes || len(node.Operands) > limits.MaxNodes || len(node.Args)+len(node.Operands) > limits.MaxNodes {
			return nil, resourceLimit(path, "node edge count exceeds the configured node limit")
		}
		edges := make([]NodeID, 0, len(node.Args)+len(node.Operands)+4)
		for j, argument := range node.Args {
			if argument.Value != 0 {
				if err := validArenaEdge(argument.Value, len(copyNodes), fmt.Sprintf("%s.args[%d].value", path, j)); err != nil {
					return nil, err
				}
				edges = append(edges, argument.Value)
			}
		}
		for j, operand := range node.Operands {
			if operand != 0 {
				if err := validArenaEdge(operand, len(copyNodes), fmt.Sprintf("%s.operands[%d]", path, j)); err != nil {
					return nil, err
				}
				edges = append(edges, operand)
			}
		}
		for _, edge := range []struct {
			field string
			child NodeID
		}{{"body", node.Body}, {"function", node.Function}, {"argument", node.Argument}, {"restrictor", node.Restrictor}} {
			field, child := edge.field, edge.child
			if child != 0 {
				if err := validArenaEdge(child, len(copyNodes), path+"."+field); err != nil {
					return nil, err
				}
				edges = append(edges, child)
			}
		}
		totalEdges += len(edges)
		if totalEdges > limits.MaxNodes {
			return nil, resourceLimit("nodes", "arena edge count exceeds the configured node limit")
		}
		children[node.NodeID] = edges
	}

	state := make([]uint8, len(copyNodes)+1)
	expanded := make([]int, len(copyNodes)+1)
	heights := make([]int, len(copyNodes)+1)
	var measure func(NodeID) (int, int, error)
	measure = func(id NodeID) (int, int, error) {
		if state[id] == 1 {
			return 0, 0, invalid("ARENA_CYCLE", fmt.Sprintf("nodes[%d]", id-1), "arena edges must form an acyclic graph")
		}
		if state[id] == 2 {
			return expanded[id], heights[id], nil
		}
		state[id] = 1
		count, height := 1, 1
		for _, child := range children[id] {
			childCount, childHeight, childErr := measure(child)
			if childErr != nil {
				return 0, 0, childErr
			}
			if childCount > limits.MaxNodes-count {
				return 0, 0, resourceLimit(fmt.Sprintf("nodes[%d]", id-1), "expanded syntax occurrence count exceeds the configured node limit")
			}
			count += childCount
			if childHeight+1 > height {
				height = childHeight + 1
			}
		}
		if height > limits.MaxDepth {
			return 0, 0, resourceLimit(fmt.Sprintf("nodes[%d]", id-1), "expanded syntax depth exceeds the configured depth limit")
		}
		state[id], expanded[id], heights[id] = 2, count, height
		return count, height, nil
	}
	for id := NodeID(1); uint64(id) <= uint64(len(copyNodes)); id++ {
		if _, _, err := measure(id); err != nil {
			return nil, err
		}
	}

	reachable := make([]bool, len(copyNodes)+1)
	stack := []NodeID{root}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[id] {
			continue
		}
		reachable[id] = true
		stack = append(stack, children[id]...)
	}
	for id := NodeID(1); uint64(id) <= uint64(len(copyNodes)); id++ {
		if !reachable[id] {
			return nil, invalid("UNREACHABLE_ARENA_NODE", fmt.Sprintf("nodes[%d]", id-1), "every arena node must be reachable from the root")
		}
	}

	rootTerm, err := arenaTerm(root, copyNodes, limits)
	if err != nil {
		return nil, err
	}
	ownedContext, err := cloneArenaContext(context, limits)
	if err != nil {
		return nil, err
	}
	document := ownedContext
	document.Expression = rootTerm
	checked, err := Check(document)
	if err != nil {
		return nil, err
	}
	ownedChecked, err := cloneCheckedExpression(checked, limits)
	if err != nil {
		return nil, err
	}
	return &Arena{nodes: copyNodes, root: root, context: ownedContext, checked: ownedChecked, limits: limits}, nil
}

// Root returns the artifact-local syntax root ID.
func (arena *Arena) Root() NodeID {
	if arena == nil {
		return 0
	}
	return arena.root
}

// Count returns the number of records, not the expanded term occurrence
// count (which is separately bounded during construction).
func (arena *Arena) Count() int {
	if arena == nil {
		return 0
	}
	return len(arena.nodes)
}

// Nodes returns detached syntax records in dense NodeID order.
func (arena *Arena) Nodes() []Node {
	if arena == nil {
		return []Node{}
	}
	copyNodes, err := cloneArenaNodes(arena.nodes, arena.limits)
	if err != nil {
		return []Node{}
	}
	return copyNodes
}

// Node returns one detached syntax record when id is present.
func (arena *Arena) Node(id NodeID) (Node, bool) {
	if arena == nil || id == 0 || uint64(id) > uint64(len(arena.nodes)) {
		return Node{}, false
	}
	copyNodes, err := cloneArenaNodes([]Node{arena.nodes[id-1]}, arena.limits)
	if err != nil {
		return Node{}, false
	}
	copyNodes[0].NodeID = id
	return copyNodes[0], true
}

// Checked returns a detached structural-check result.
func (arena *Arena) Checked() CheckedExpression {
	if arena == nil {
		return CheckedExpression{}
	}
	copyResult, err := cloneCheckedExpression(arena.checked, arena.limits)
	if err != nil {
		return CheckedExpression{}
	}
	return copyResult
}

// Document reconstructs a detached checked document from arena records and
// the declaration context.
func (arena *Arena) Document() (Document, error) {
	if arena == nil || arena.root == 0 || len(arena.nodes) == 0 {
		return Document{}, invalid("UNINITIALIZED_ARENA", "arena", "a validated arena is required")
	}
	context, err := cloneArenaContext(arena.context, arena.limits)
	if err != nil {
		return Document{}, err
	}
	root, err := arenaTerm(arena.root, arena.nodes, arena.limits)
	if err != nil {
		return Document{}, err
	}
	context.Expression = root
	if _, err := Check(context); err != nil {
		return Document{}, err
	}
	return context, nil
}

func validateArenaNodeShape(node Node, path string) error {
	identityKind := node.Kind == TermVar || node.Kind == TermParam || node.Kind == TermRef || node.Kind == TermHole || node.Kind == TermExists || node.Kind == TermLambda || node.Kind == TermQuantifier
	if !identityKind && node.Identity != "" {
		return invalid("MALFORMED_ARENA_NODE", path+".identity", "identity is not allowed for this node kind")
	}
	if node.Kind != TermAtom && node.Symbol != "" {
		return invalid("MALFORMED_ARENA_NODE", path+".symbol", "symbol is not allowed for this node kind")
	}
	if node.Kind != TermAtom && len(node.Args) != 0 {
		return invalid("MALFORMED_ARENA_NODE", path+".args", "args are not allowed for this node kind")
	}
	if node.Kind != TermAnd && len(node.Operands) != 0 {
		return invalid("MALFORMED_ARENA_NODE", path+".operands", "operands are not allowed for this node kind")
	}
	if node.Kind != TermLambda && node.Kind != TermExists && node.Kind != TermNot && node.Kind != TermQuantifier && node.Kind != TermAsk && node.Body != 0 {
		return invalid("MALFORMED_ARENA_NODE", path+".body", "body edge is not allowed for this node kind")
	}
	if node.Kind != TermApply && (node.Function != 0 || node.Argument != 0) {
		return invalid("MALFORMED_ARENA_NODE", path, "function/argument edges are not allowed for this node kind")
	}
	if node.Kind != TermParam && node.Kind != TermRef && node.Kind != TermHole && node.Kind != TermLambda && node.Kind != TermExists && node.Kind != TermQuantifier && node.Type != nil {
		return invalid("MALFORMED_ARENA_NODE", path+".type", "type is not allowed for this node kind")
	}
	if node.Kind != TermQuantifier && node.Quantifier != "" {
		return invalid("MALFORMED_ARENA_NODE", path+".quantifier", "quantifier is not allowed for this node kind")
	}
	if node.Kind != TermQuantifier && node.Count != nil {
		return invalid("MALFORMED_ARENA_NODE", path+".count", "count is not allowed for this node kind")
	}
	if node.Kind != TermQuantifier && node.Kind != TermAsk && node.Restrictor != 0 {
		return invalid("MALFORMED_ARENA_NODE", path+".restrictor", "restrictor edge is not allowed for this node kind")
	}
	if node.Kind != TermAsk && node.Answers != nil {
		return invalid("MALFORMED_ARENA_NODE", path+".answers", "answers are not allowed for this node kind")
	}
	return nil
}

func validArenaEdge(id NodeID, count int, path string) error {
	if id == 0 || uint64(id) > uint64(count) {
		return invalid("INVALID_ARENA_EDGE", path, "edge must identify a declared node")
	}
	return nil
}

func arenaTerm(root NodeID, nodes []Node, limits Limits) (*Term, error) {
	if root == 0 || uint64(root) > uint64(len(nodes)) {
		return nil, invalid("INVALID_ARENA_ROOT", "root", "root NodeID must identify a declared node")
	}
	cache := make([]*Term, len(nodes)+1)
	var build func(NodeID) (*Term, error)
	build = func(id NodeID) (*Term, error) {
		if id == 0 {
			return nil, nil
		}
		if uint64(id) > uint64(len(nodes)) {
			return nil, invalid("INVALID_ARENA_EDGE", "nodes", "edge must identify a declared node")
		}
		if cache[id] != nil {
			return cache[id], nil
		}
		node := nodes[id-1]
		term := &Term{Kind: node.Kind, ID: node.Identity, Label: node.Label, Symbol: node.Symbol, Quantifier: node.Quantifier}
		if node.Count != nil {
			count := *node.Count
			term.Count = &count
		}
		if node.Kind == TermAsk {
			term.Answers = make([]Declaration, len(node.Answers))
			for i, answer := range node.Answers {
				typeCopy, err := cloneStandaloneType(&answer.Type, limits)
				if err != nil {
					return nil, err
				}
				term.Answers[i] = Declaration{ID: answer.ID, Type: typeCopy}
			}
		}
		if node.Type != nil {
			typeCopy, err := cloneStandaloneType(node.Type, limits)
			if err != nil {
				return nil, err
			}
			term.Type = &typeCopy
		}
		cache[id] = term
		if node.Args != nil {
			term.Args = make([]Argument, len(node.Args))
			for i, argument := range node.Args {
				value, err := build(argument.Value)
				if err != nil {
					return nil, err
				}
				term.Args[i] = Argument{Role: argument.Role, Value: value}
			}
		}
		if node.Operands != nil {
			term.Operands = make([]*Term, len(node.Operands))
			for i, child := range node.Operands {
				value, err := build(child)
				if err != nil {
					return nil, err
				}
				term.Operands[i] = value
			}
		}
		var err error
		term.Body, err = build(node.Body)
		if err != nil {
			return nil, err
		}
		term.Restrictor, err = build(node.Restrictor)
		if err != nil {
			return nil, err
		}
		term.Function, err = build(node.Function)
		if err != nil {
			return nil, err
		}
		term.Argument, err = build(node.Argument)
		if err != nil {
			return nil, err
		}
		return term, nil
	}
	return build(root)
}

func cloneArenaNodes(nodes []Node, limits Limits) ([]Node, error) {
	if len(nodes) > limits.MaxNodes {
		return nil, resourceLimit("nodes", "arena node count exceeds the configured node limit")
	}
	copyNodes := make([]Node, len(nodes))
	for i, node := range nodes {
		copyNodes[i] = node
		copyNodes[i].Count = nil
		if node.Count != nil {
			count := *node.Count
			copyNodes[i].Count = &count
		}
		if node.Kind == TermAsk {
			copyNodes[i].Answers = make([]Declaration, len(node.Answers))
			for j, answer := range node.Answers {
				typeCopy, typeErr := cloneStandaloneType(&answer.Type, limits)
				if typeErr != nil {
					return nil, typeErr
				}
				copyNodes[i].Answers[j] = Declaration{ID: answer.ID, Type: typeCopy}
			}
		}
		if node.Type != nil {
			typeCopy, err := cloneStandaloneType(node.Type, limits)
			if err != nil {
				return nil, err
			}
			copyNodes[i].Type = &typeCopy
		}
		copyNodes[i].Args = append([]NodeArgument(nil), node.Args...)
		copyNodes[i].Operands = append([]NodeID(nil), node.Operands...)
	}
	return copyNodes, nil
}

func cloneArenaContext(context Document, limits Limits) (Document, error) {
	if context.Expression != nil {
		return Document{}, invalid("ARENA_CONTEXT_HAS_EXPRESSION", "context.expression", "arena context must contain declarations only")
	}
	copyContext := context
	copyContext.Expression = nil
	copyContext.Sorts = append([]SortDeclaration(nil), context.Sorts...)
	copyContext.Signatures = make([]Signature, len(context.Signatures))
	for i, signature := range context.Signatures {
		copyContext.Signatures[i] = Signature{ID: signature.ID, Label: signature.Label, Ports: make([]Port, len(signature.Ports))}
		for j, port := range signature.Ports {
			typeCopy, err := cloneStandaloneType(&port.Type, limits)
			if err != nil {
				return Document{}, err
			}
			copyContext.Signatures[i].Ports[j] = Port{Role: port.Role, Type: typeCopy}
		}
	}
	cloneDeclarations := func(values []Declaration) ([]Declaration, error) {
		result := make([]Declaration, len(values))
		for i, declaration := range values {
			typeCopy, err := cloneStandaloneType(&declaration.Type, limits)
			if err != nil {
				return nil, err
			}
			result[i] = Declaration{ID: declaration.ID, Type: typeCopy}
		}
		return result, nil
	}
	var err error
	copyContext.Parameters, err = cloneDeclarations(context.Parameters)
	if err != nil {
		return Document{}, err
	}
	copyContext.References, err = cloneDeclarations(context.References)
	if err != nil {
		return Document{}, err
	}
	return copyContext, nil
}

func cloneCheckedExpression(value CheckedExpression, limits Limits) (CheckedExpression, error) {
	copyValue := value
	typeCopy, err := cloneStandaloneType(&value.Type, limits)
	if err != nil {
		return CheckedExpression{}, err
	}
	copyValue.Type = typeCopy
	copyDeclarations := func(values []Declaration) ([]Declaration, error) {
		result := make([]Declaration, len(values))
		for i, declaration := range values {
			typeCopy, err := cloneStandaloneType(&declaration.Type, limits)
			if err != nil {
				return nil, err
			}
			result[i] = Declaration{ID: declaration.ID, Type: typeCopy}
		}
		return result, nil
	}
	copyValue.FreeParameters, err = copyDeclarations(value.FreeParameters)
	if err != nil {
		return CheckedExpression{}, err
	}
	copyRequirements := func(values []Requirement) ([]Requirement, error) {
		result := make([]Requirement, len(values))
		for i, requirement := range values {
			typeCopy, err := cloneStandaloneType(&requirement.Type, limits)
			if err != nil {
				return nil, err
			}
			result[i] = Requirement{Kind: requirement.Kind, ID: requirement.ID, Type: typeCopy}
		}
		return result, nil
	}
	copyValue.Holes, err = copyRequirements(value.Holes)
	if err != nil {
		return CheckedExpression{}, err
	}
	copyValue.Requirements, err = copyRequirements(value.Requirements)
	if err != nil {
		return CheckedExpression{}, err
	}
	return copyValue, nil
}

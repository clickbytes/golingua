package algebra

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/status"
)

const SchemaVersion = 1

// Document is the in-memory schema consumed by the structural checker. JSON
// decoding and byte-size enforcement are intentionally implemented separately.
type Document struct {
	SchemaVersion int               `json:"schema_version"`
	Sorts         []SortDeclaration `json:"sorts,omitempty"`
	Signatures    []Signature       `json:"signatures,omitempty"`
	Parameters    []Declaration     `json:"parameters,omitempty"`
	References    []Declaration     `json:"references,omitempty"`
	Closed        bool              `json:"closed,omitempty"`
	Expression    *Term             `json:"expression"`
	Limits        Limits            `json:"limits,omitempty"`
}

type checker struct {
	limits      Limits
	sorts       map[string]struct{}
	signatures  map[string]Signature
	parameters  map[string]Declaration
	references  map[string]Declaration
	scope       *checkerScope
	activeTerms map[*Term]bool
	freeParams  map[string]Declaration
	holes       map[string]Requirement
	nodeCount   int
	typeVisits  int
}

// Check validates one typed expression against its inline declarations. It
// returns type/scope information only; it neither evaluates propositions nor
// claims that they are true.
func Check(document Document) (CheckedExpression, error) {
	checker, _, err := prepareChecker(document, true)
	if err != nil {
		return CheckedExpression{}, err
	}
	resultType, err := checker.visit(document.Expression, "expression", 1)
	if err != nil {
		return CheckedExpression{}, err
	}
	if document.Closed && len(checker.freeParams) != 0 {
		ids := sortedDeclarationIDs(checker.freeParams)
		return CheckedExpression{}, invalid("FREE_PARAMETER_IN_CLOSED_TERM", "expression", "closed expression has free parameter(s): "+joinIDs(ids))
	}
	return checkedResult(checker, resultType), nil
}

// CheckDeclarations validates inline sorts, signatures, parameters, and
// references without fabricating an expression or a hole.
func CheckDeclarations(document Document) error {
	_, _, err := prepareChecker(document, false)
	return err
}

func prepareChecker(document Document, requireExpression bool) (*checker, Limits, error) {
	if document.SchemaVersion != SchemaVersion {
		return nil, Limits{}, invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported algebra document version")
	}
	limits, err := NormalizeLimits(document.Limits)
	if err != nil {
		return nil, Limits{}, err
	}
	if requireExpression && document.Expression == nil {
		return nil, Limits{}, invalid("MISSING_EXPRESSION", "expression", "expression is required")
	}
	totalSymbols := len(document.Sorts) + len(document.Signatures) + len(document.Parameters) + len(document.References)
	if totalSymbols > limits.MaxSymbols {
		return nil, Limits{}, resourceLimit("limits.max_symbols", fmt.Sprintf("symbol declarations %d exceed limit %d", totalSymbols, limits.MaxSymbols))
	}
	checker := &checker{
		limits: limits, sorts: make(map[string]struct{}, len(document.Sorts)),
		signatures: make(map[string]Signature, len(document.Signatures)),
		parameters: make(map[string]Declaration, len(document.Parameters)),
		references: make(map[string]Declaration, len(document.References)),
		scope:      newCheckerScope(), activeTerms: make(map[*Term]bool),
		freeParams: make(map[string]Declaration), holes: make(map[string]Requirement),
	}
	if err := checker.readSorts(document.Sorts); err != nil {
		return nil, Limits{}, err
	}
	if err := checker.readSignatures(document.Signatures); err != nil {
		return nil, Limits{}, err
	}
	if err := checker.readDeclarations("parameters", document.Parameters, checker.parameters); err != nil {
		return nil, Limits{}, err
	}
	if err := checker.readDeclarations("references", document.References, checker.references); err != nil {
		return nil, Limits{}, err
	}
	return checker, limits, nil
}

func checkedResult(checker *checker, resultType Type) CheckedExpression {
	free := make([]Declaration, 0, len(checker.freeParams))
	for _, id := range sortedDeclarationIDs(checker.freeParams) {
		free = append(free, checker.freeParams[id])
	}
	holes := make([]Requirement, 0, len(checker.holes))
	for _, id := range sortedRequirementIDs(checker.holes) {
		holes = append(holes, checker.holes[id])
	}
	resultStatus := status.OutcomeOK
	if len(holes) != 0 {
		resultStatus = status.OutcomeIncomplete
	}
	return CheckedExpression{
		Status: resultStatus, Type: resultType, FreeParameters: free,
		Holes: holes, Requirements: append([]Requirement(nil), holes...), NodeCount: checker.nodeCount,
	}
}

func (c *checker) readSorts(sorts []SortDeclaration) error {
	for i, declaration := range sorts {
		path := fmt.Sprintf("sorts[%d]", i)
		if !validText(declaration.ID, 256) {
			return invalid("INVALID_SORT_ID", path+".id", "sort ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if !validOptionalText(declaration.Label, 1024) {
			return invalid("INVALID_LABEL", path+".label", "label must be valid UTF-8 and at most 1024 bytes")
		}
		if _, exists := c.sorts[declaration.ID]; exists {
			return invalid("DUPLICATE_SORT_ID", path+".id", "sort ID is duplicated")
		}
		c.sorts[declaration.ID] = struct{}{}
	}
	return nil
}

func (c *checker) readSignatures(signatures []Signature) error {
	for i, signature := range signatures {
		path := fmt.Sprintf("signatures[%d]", i)
		if !validText(signature.ID, 256) {
			return invalid("INVALID_SYMBOL_ID", path+".id", "symbol ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if !validOptionalText(signature.Label, 1024) {
			return invalid("INVALID_LABEL", path+".label", "label must be valid UTF-8 and at most 1024 bytes")
		}
		if len(signature.Ports) > c.limits.MaxPorts {
			return resourceLimit(path+".ports", fmt.Sprintf("port count %d exceeds limit %d", len(signature.Ports), c.limits.MaxPorts))
		}
		if _, exists := c.signatures[signature.ID]; exists {
			return invalid("DUPLICATE_SYMBOL_ID", path+".id", "symbol ID is duplicated")
		}
		seenRoles := make(map[string]struct{}, len(signature.Ports))
		for j, port := range signature.Ports {
			portPath := fmt.Sprintf("%s.ports[%d]", path, j)
			if !validText(port.Role, 256) {
				return invalid("INVALID_ROLE_ID", portPath+".role", "role ID must be non-empty valid UTF-8 and at most 256 bytes")
			}
			if _, exists := seenRoles[port.Role]; exists {
				return invalid("DUPLICATE_ROLE_ID", portPath+".role", "role ID is duplicated in this signature")
			}
			seenRoles[port.Role] = struct{}{}
			if err := c.validateType(&port.Type, portPath+".type", 1, make(map[*Type]bool)); err != nil {
				return err
			}
			if port.Type.Kind == TypeArrow {
				return invalid("INVALID_RELATION_PORT_TYPE", portPath+".type", "relation port must have a base sort type")
			}
		}
		c.signatures[signature.ID] = signature
	}
	return nil
}

func (c *checker) readDeclarations(name string, declarations []Declaration, into map[string]Declaration) error {
	for i, declaration := range declarations {
		path := fmt.Sprintf("%s[%d]", name, i)
		if !validText(declaration.ID, 256) {
			return invalid("INVALID_DECLARATION_ID", path+".id", "declaration ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if _, exists := into[declaration.ID]; exists {
			return invalid("DUPLICATE_DECLARATION_ID", path+".id", "declaration ID is duplicated")
		}
		if err := c.validateType(&declaration.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
			return err
		}
		into[declaration.ID] = declaration
	}
	return nil
}

func (c *checker) validateType(value *Type, path string, depth int, active map[*Type]bool) error {
	if value == nil {
		return invalid("MISSING_TYPE", path, "type is required")
	}
	if depth > c.limits.MaxTypeDepth {
		return resourceLimit(path, fmt.Sprintf("type depth exceeds limit %d", c.limits.MaxTypeDepth))
	}
	if active[value] {
		return invalid("CYCLIC_TYPE", path, "type definition contains a pointer cycle")
	}
	c.typeVisits++
	maxTypeVisits := c.limits.MaxNodes * c.limits.MaxTypeDepth
	if c.typeVisits > maxTypeVisits {
		return resourceLimit(path, fmt.Sprintf("type node checks exceed limit %d", maxTypeVisits))
	}
	active[value] = true
	defer delete(active, value)
	switch value.Kind {
	case TypeEntity, TypeEvent, TypeInterval, TypeProp, TypeQuestion:
		if value.SortID != "" || value.From != nil || value.To != nil {
			return invalid("MALFORMED_TYPE", path, "built-in type cannot carry sort or arrow fields")
		}
	case TypeNominal:
		if !validText(value.SortID, 256) || value.From != nil || value.To != nil {
			return invalid("MALFORMED_TYPE", path, "nominal type requires only a valid sort_id")
		}
		if _, exists := c.sorts[value.SortID]; !exists {
			return invalid("UNKNOWN_SORT", path+".sort_id", "nominal sort is not declared")
		}
	case TypeArrow:
		if value.SortID != "" || value.From == nil || value.To == nil {
			return invalid("MALFORMED_TYPE", path, "arrow type requires from and to and cannot carry sort_id")
		}
		if err := c.validateType(value.From, path+".from", depth+1, active); err != nil {
			return err
		}
		if err := c.validateType(value.To, path+".to", depth+1, active); err != nil {
			return err
		}
	default:
		return invalid("UNKNOWN_TYPE_KIND", path+".kind", "type kind is unsupported")
	}
	return nil
}

func (c *checker) visit(term *Term, path string, depth int) (Type, error) {
	if term == nil {
		return Type{}, invalid("MISSING_TERM", path, "term is required")
	}
	if depth > c.limits.MaxDepth {
		return Type{}, resourceLimit(path, fmt.Sprintf("expression depth exceeds limit %d", c.limits.MaxDepth))
	}
	if c.activeTerms[term] {
		return Type{}, invalid("CYCLIC_TERM", path, "expression contains a pointer cycle")
	}
	c.nodeCount++
	if c.nodeCount > c.limits.MaxNodes {
		return Type{}, resourceLimit(path, fmt.Sprintf("expression node count exceeds limit %d", c.limits.MaxNodes))
	}
	if !validOptionalText(term.Label, 1024) {
		return Type{}, invalid("INVALID_LABEL", path+".label", "label must be valid UTF-8 and at most 1024 bytes")
	}
	if err := validateTermShape(term, path); err != nil {
		return Type{}, err
	}
	c.activeTerms[term] = true
	defer delete(c.activeTerms, term)
	switch term.Kind {
	case TermVar:
		if !validText(term.ID, 256) {
			return Type{}, invalid("INVALID_BINDER_ID", path+".id", "binder ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		value, exists := c.scope.lookup(term.ID)
		if !exists {
			return Type{}, invalid("UNBOUND_VARIABLE", path+".id", "variable does not refer to an in-scope binder")
		}
		return value, nil
	case TermParam:
		if !validText(term.ID, 256) {
			return Type{}, invalid("INVALID_PARAMETER_ID", path+".id", "parameter ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		declaration, exists := c.parameters[term.ID]
		if !exists {
			return Type{}, invalid("UNDECLARED_PARAMETER", path+".id", "parameter ID is not declared")
		}
		if err := c.validateType(term.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
			return Type{}, err
		}
		if !TypeEqual(*term.Type, declaration.Type) {
			return Type{}, invalid("PARAMETER_TYPE_MISMATCH", path+".type", "parameter type differs from its declaration")
		}
		c.freeParams[term.ID] = declaration
		return declaration.Type, nil
	case TermRef:
		if !validText(term.ID, 256) {
			return Type{}, invalid("INVALID_REFERENCE_ID", path+".id", "reference ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		declaration, exists := c.references[term.ID]
		if !exists {
			return Type{}, invalid("UNDECLARED_REFERENCE", path+".id", "reference ID is not declared")
		}
		if err := c.validateType(term.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
			return Type{}, err
		}
		if !TypeEqual(*term.Type, declaration.Type) {
			return Type{}, invalid("REFERENCE_TYPE_MISMATCH", path+".type", "reference type differs from its declaration")
		}
		return declaration.Type, nil
	case TermHole:
		if !validText(term.ID, 256) {
			return Type{}, invalid("INVALID_HOLE_ID", path+".id", "hole ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if _, exists := c.holes[term.ID]; exists {
			return Type{}, invalid("DUPLICATE_HOLE_ID", path+".id", "hole ID is duplicated")
		}
		if err := c.validateType(term.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
			return Type{}, err
		}
		c.holes[term.ID] = Requirement{Kind: "hole", ID: term.ID, Type: *term.Type}
		return *term.Type, nil
	case TermAtom:
		return c.visitAtom(term, path, depth)
	case TermNot:
		bodyType, err := c.visit(term.Body, path+".body", depth+1)
		if err != nil {
			return Type{}, err
		}
		if bodyType.Kind != TypeProp {
			return Type{}, invalid("EXPECTED_PROP", path+".body", "Not body must have type Prop")
		}
		return Type{Kind: TypeProp}, nil
	case TermQuantifier:
		return c.visitQuantifier(term, path, depth)
	case TermAsk:
		return c.visitAsk(term, path, depth)
	case TermAnd:
		if len(term.Operands) == 0 {
			return Type{}, invalid("EMPTY_CONJUNCTION", path+".operands", "And requires at least one proposition")
		}
		for i, operand := range term.Operands {
			operandType, err := c.visit(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1)
			if err != nil {
				return Type{}, err
			}
			if operandType.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", fmt.Sprintf("%s.operands[%d]", path, i), "And operands must have type Prop")
			}
		}
		return Type{Kind: TypeProp}, nil
	case TermExists, TermLambda:
		if !validText(term.ID, 256) {
			return Type{}, invalid("INVALID_BINDER_ID", path+".id", "binder ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if err := c.validateType(term.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
			return Type{}, err
		}
		if err := c.scope.bind(term.ID, *term.Type, path+".id"); err != nil {
			return Type{}, err
		}
		bodyType, bodyErr := c.visit(term.Body, path+".body", depth+1)
		c.scope.unbind(term.ID)
		if bodyErr != nil {
			return Type{}, bodyErr
		}
		if term.Kind == TermExists {
			if bodyType.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "Exists body must have type Prop")
			}
			return Type{Kind: TypeProp}, nil
		}
		argumentType := *term.Type
		return Type{Kind: TypeArrow, From: &argumentType, To: &bodyType}, nil
	case TermApply:
		functionType, err := c.visit(term.Function, path+".function", depth+1)
		if err != nil {
			return Type{}, err
		}
		if functionType.Kind != TypeArrow || functionType.From == nil || functionType.To == nil {
			return Type{}, invalid("NOT_A_FUNCTION", path+".function", "application target must have an arrow type")
		}
		argumentType, err := c.visit(term.Argument, path+".argument", depth+1)
		if err != nil {
			return Type{}, err
		}
		if !TypeEqual(argumentType, *functionType.From) {
			return Type{}, invalid("ARGUMENT_TYPE_MISMATCH", path+".argument", "argument type does not match function input type")
		}
		return *functionType.To, nil
	default:
		return Type{}, invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind is unsupported")
	}
}

func (c *checker) visitAtom(term *Term, path string, depth int) (Type, error) {
	if !validText(term.Symbol, 256) {
		return Type{}, invalid("INVALID_SYMBOL_ID", path+".symbol", "symbol ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	signature, exists := c.signatures[term.Symbol]
	if !exists {
		return Type{}, invalid("UNKNOWN_SYMBOL", path+".symbol", "symbol has no declared signature")
	}
	if len(term.Args) > c.limits.MaxPorts {
		return Type{}, resourceLimit(path+".args", fmt.Sprintf("argument count %d exceeds port limit %d", len(term.Args), c.limits.MaxPorts))
	}
	seen := make(map[string]struct{}, len(term.Args))
	for i, argument := range term.Args {
		argPath := fmt.Sprintf("%s.args[%d]", path, i)
		if !validText(argument.Role, 256) {
			return Type{}, invalid("INVALID_ROLE_ID", argPath+".role", "argument role must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if _, duplicate := seen[argument.Role]; duplicate {
			return Type{}, invalid("DUPLICATE_ARGUMENT_ROLE", argPath+".role", "argument role is duplicated")
		}
		seen[argument.Role] = struct{}{}
		known := false
		for _, port := range signature.Ports {
			if argument.Role == port.Role {
				known = true
				break
			}
		}
		if !known {
			return Type{}, invalid("UNKNOWN_ARGUMENT_ROLE", argPath+".role", "argument role is not declared by the symbol signature")
		}
	}
	if len(term.Args) != len(signature.Ports) {
		return Type{}, invalid("ARITY_MISMATCH", path+".args", fmt.Sprintf("symbol expects %d arguments, got %d", len(signature.Ports), len(term.Args)))
	}
	for i, port := range signature.Ports {
		argument := term.Args[i]
		argPath := fmt.Sprintf("%s.args[%d]", path, i)
		if argument.Role != port.Role {
			return Type{}, invalid("ARGUMENT_ORDER_MISMATCH", argPath+".role", "argument roles must follow the declared signature order")
		}
		argumentType, err := c.visit(argument.Value, argPath+".value", depth+1)
		if err != nil {
			return Type{}, err
		}
		if !TypeEqual(argumentType, port.Type) {
			return Type{}, invalid("ARGUMENT_TYPE_MISMATCH", argPath+".value", "argument type differs from its declared port type")
		}
	}
	return Type{Kind: TypeProp}, nil
}

func validateTermShape(term *Term, path string) error {
	bad := func(field string) error {
		return invalid("MALFORMED_TERM", path+"."+field, "field is not allowed for this term kind")
	}
	usesType := term.Kind == TermParam || term.Kind == TermRef || term.Kind == TermHole || term.Kind == TermLambda || term.Kind == TermExists || term.Kind == TermQuantifier
	if !usesType && term.Type != nil {
		return bad("type")
	}
	if term.Kind != TermAtom && term.Symbol != "" {
		return bad("symbol")
	}
	if term.Kind != TermAtom && len(term.Args) != 0 {
		return bad("args")
	}
	if term.Kind != TermAnd && len(term.Operands) != 0 {
		return bad("operands")
	}
	usesBody := term.Kind == TermLambda || term.Kind == TermExists || term.Kind == TermNot || term.Kind == TermQuantifier || term.Kind == TermAsk
	if !usesBody && term.Body != nil {
		return bad("body")
	}
	if term.Kind != TermApply && (term.Function != nil || term.Argument != nil) {
		return bad("function/argument")
	}
	if term.Kind != TermQuantifier && term.Quantifier != "" {
		return bad("quantifier")
	}
	if term.Kind != TermQuantifier && term.Count != nil {
		return bad("count")
	}
	if term.Kind != TermQuantifier && term.Kind != TermAsk && term.Restrictor != nil {
		return bad("restrictor")
	}
	if term.Kind != TermAsk && term.Answers != nil {
		return bad("answers")
	}
	usesID := term.Kind == TermVar || term.Kind == TermParam || term.Kind == TermRef || term.Kind == TermHole || term.Kind == TermLambda || term.Kind == TermExists || term.Kind == TermQuantifier
	if !usesID && term.ID != "" {
		return bad("id")
	}
	switch term.Kind {
	case TermVar:
	case TermParam, TermRef, TermHole:
		if term.Type == nil {
			return invalid("MALFORMED_TERM", path+".type", "typed term requires a type")
		}
	case TermAtom:
		if !termHasFields(term, "symbol") {
			return invalid("MALFORMED_TERM", path, "Atom requires a symbol")
		}
	case TermAnd:
		if term.Operands == nil {
			return invalid("MALFORMED_TERM", path+".operands", "And requires an operand list")
		}
	case TermExists, TermLambda:
		if term.ID == "" || term.Type == nil || term.Body == nil {
			return invalid("MALFORMED_TERM", path, "binder term requires id, type, and body")
		}
	case TermNot:
		if term.Body == nil {
			return invalid("MALFORMED_TERM", path+".body", "Not requires a body")
		}
	case TermQuantifier:
		if term.ID == "" || term.Type == nil || term.Body == nil || term.Quantifier == "" {
			return invalid("MALFORMED_TERM", path, "Quantifier requires id, type, quantifier, and body")
		}
	case TermAsk:
		if term.Body == nil {
			return invalid("MALFORMED_TERM", path+".body", "Ask requires a body")
		}
	case TermApply:
		if term.Function == nil || term.Argument == nil {
			return invalid("MALFORMED_TERM", path, "Apply requires function and argument")
		}
	default:
		return invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind is unsupported")
	}
	return nil
}

// termHasFields distinguishes an empty required string from an omitted field.
// The in-memory form cannot represent JSON null/field presence, so the checked
// structural contract treats the required string's nonempty value as present.
func termHasFields(term *Term, field string) bool {
	return field == "symbol" && term.Symbol != ""
}

func sortedDeclarationIDs(values map[string]Declaration) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func sortedRequirementIDs(values map[string]Requirement) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func joinIDs(ids []string) string {
	result := ""
	for i, id := range ids {
		if i != 0 {
			result += ", "
		}
		result += id
	}
	return result
}

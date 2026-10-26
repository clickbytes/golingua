package algebra

import "fmt"

type checkerScope struct {
	active   map[string]Type
	declared map[string]struct{}
}

func newCheckerScope() *checkerScope {
	return &checkerScope{active: make(map[string]Type), declared: make(map[string]struct{})}
}

func (scope *checkerScope) bind(id string, value Type, path string) error {
	if _, exists := scope.declared[id]; exists {
		return invalid("DUPLICATE_BINDER_ID", path, "binder ID must be globally unique")
	}
	scope.declared[id] = struct{}{}
	scope.active[id] = value
	return nil
}

func (scope *checkerScope) unbind(id string) { delete(scope.active, id) }

func (scope *checkerScope) lookup(id string) (Type, bool) {
	value, ok := scope.active[id]
	return value, ok
}

func (c *checker) visitQuantifier(term *Term, path string, depth int) (Type, error) {
	if !validText(term.ID, 256) {
		return Type{}, invalid("INVALID_BINDER_ID", path+".id", "binder ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	if err := c.validateType(term.Type, path+".type", 1, make(map[*Type]bool)); err != nil {
		return Type{}, err
	}
	if !isDataType(*term.Type) {
		return Type{}, invalid("INVALID_QUANTIFIER_BINDER_TYPE", path+".type", "quantifier binder must have a data sort type")
	}
	switch term.Quantifier {
	case "some", "all":
		if term.Count != nil {
			return Type{}, invalid("INVALID_QUANTIFIER_COUNT", path+".count", "some and all do not accept count")
		}
	case "exactly", "at_least", "at_most":
		if term.Count == nil {
			return Type{}, invalid("MISSING_QUANTIFIER_COUNT", path+".count", "numeric quantifier requires a count")
		}
		if *term.Count < 0 {
			return Type{}, invalid("INVALID_QUANTIFIER_COUNT", path+".count", "count must be non-negative")
		}
	default:
		return Type{}, invalid("INVALID_QUANTIFIER_KIND", path+".quantifier", "quantifier kind is unsupported")
	}
	if err := c.scope.bind(term.ID, *term.Type, path+".id"); err != nil {
		return Type{}, err
	}
	if term.Restrictor != nil {
		typ, err := c.visit(term.Restrictor, path+".restrictor", depth+1)
		if err != nil {
			c.scope.unbind(term.ID)
			return Type{}, err
		}
		if typ.Kind != TypeProp {
			c.scope.unbind(term.ID)
			return Type{}, invalid("EXPECTED_PROP", path+".restrictor", "quantifier restrictor must have type Prop")
		}
	}
	typ, err := c.visit(term.Body, path+".body", depth+1)
	c.scope.unbind(term.ID)
	if err != nil {
		return Type{}, err
	}
	if typ.Kind != TypeProp {
		return Type{}, invalid("EXPECTED_PROP", path+".body", "quantifier body must have type Prop")
	}
	return Type{Kind: TypeProp}, nil
}

func (c *checker) visitAsk(term *Term, path string, depth int) (Type, error) {
	if len(term.Answers) > c.limits.MaxPorts {
		return Type{}, resourceLimit(path+".answers", fmt.Sprintf("answer count %d exceeds limit %d", len(term.Answers), c.limits.MaxPorts))
	}
	bound := make([]string, 0, len(term.Answers))
	for i, answer := range term.Answers {
		answerPath := fmt.Sprintf("%s.answers[%d]", path, i)
		if answer.TemporalFrame != "" || answer.ReferenceTimeID != "" || answer.Multiplicity != "" {
			return Type{}, invalid("MALFORMED_ANSWER", answerPath, "Ask answer declarations cannot carry reference attributes")
		}
		if !validText(answer.ID, 256) {
			return Type{}, invalid("INVALID_ANSWER_ID", answerPath+".id", "answer ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if answer.TemporalFrame != "" || answer.ReferenceTimeID != "" || answer.Multiplicity != "" {
			return Type{}, invalid("MALFORMED_ANSWER", answerPath, "Ask answer declarations cannot carry reference attributes")
		}
		if err := c.validateType(&answer.Type, answerPath+".type", 1, make(map[*Type]bool)); err != nil {
			for _, id := range bound {
				c.scope.unbind(id)
			}
			return Type{}, err
		}
		if !isDataType(answer.Type) {
			for _, id := range bound {
				c.scope.unbind(id)
			}
			return Type{}, invalid("INVALID_ANSWER_TYPE", answerPath+".type", "Ask answer variable must have a data sort type")
		}
		if err := c.scope.bind(answer.ID, answer.Type, answerPath+".id"); err != nil {
			for _, id := range bound {
				c.scope.unbind(id)
			}
			return Type{}, err
		}
		bound = append(bound, answer.ID)
	}
	cleanup := func() {
		for _, id := range bound {
			c.scope.unbind(id)
		}
	}
	if term.Restrictor != nil {
		typ, err := c.visit(term.Restrictor, path+".restrictor", depth+1)
		if err != nil {
			cleanup()
			return Type{}, err
		}
		if typ.Kind != TypeProp {
			cleanup()
			return Type{}, invalid("EXPECTED_PROP", path+".restrictor", "Ask restrictor must have type Prop")
		}
	}
	typ, err := c.visit(term.Body, path+".body", depth+1)
	cleanup()
	if err != nil {
		return Type{}, err
	}
	if typ.Kind != TypeProp {
		return Type{}, invalid("EXPECTED_PROP", path+".body", "Ask body must have type Prop")
	}
	return Type{Kind: TypeQuestion}, nil
}

func isDataType(value Type) bool {
	switch value.Kind {
	case TypeEntity, TypeEvent, TypeInterval, TypeNominal:
		return true
	default:
		return false
	}
}

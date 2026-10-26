package algebra

// Negate constructs a typed Not proposition. It does not evaluate the body.
func Negate(context Document, body *Term) (*Term, error) {
	if body == nil {
		return nil, invalid("MISSING_NOT_BODY", "body", "Not body is required")
	}
	checked, err := checkExpression(context, body)
	if err != nil {
		return nil, err
	}
	if checked.Kind != TypeProp {
		return nil, invalid("EXPECTED_PROP", "body", "Not body must have type Prop")
	}
	result := &Term{Kind: TermNot, Body: body}
	if _, err := checkExpression(context, result); err != nil {
		return nil, err
	}
	return result, nil
}

// NegateFunction wraps a predicate body while keeping its typed lambda ports.
func NegateFunction(context Document, function Function) (Function, error) {
	if err := ValidateFunction(context, function); err != nil {
		return Function{}, err
	}
	binders, body, err := peelLeading(function.Expression, len(function.Interface))
	if err != nil {
		return Function{}, err
	}
	wrapped := &Term{Kind: TermNot, Body: body}
	ports := make([]Port, len(binders))
	for i, binder := range binders {
		ports[i] = Port{Role: function.Interface[i].Port, Type: *cloneTypeValue(*binder.Type)}
	}
	result := Function{Expression: wrapLambdas(function.Interface, ports, wrapped), Interface: append([]PortBinding{}, function.Interface...)}
	if err := ValidateFunction(context, result); err != nil {
		return Function{}, err
	}
	return result, nil
}

// NewQuantifier builds a represented scope without assigning truth or
// completeness semantics to its quantifier kind.
func NewQuantifier(context Document, binder string, typ Type, kind string, count *int, restrictor, body *Term) (*Term, error) {
	var countCopy *int
	if count != nil {
		value := *count
		countCopy = &value
	}
	typCopy := cloneTypeValue(typ)
	result := &Term{Kind: TermQuantifier, ID: binder, Type: typCopy, Quantifier: kind, Count: countCopy, Restrictor: restrictor, Body: body}
	if _, err := checkExpression(context, result); err != nil {
		return nil, err
	}
	return result, nil
}

// NewAsk builds a typed Question wrapper. Answer declarations bind in both
// the optional restrictor and body; an empty answer list represents yes/no.
func NewAsk(context Document, answers []Declaration, restrictor, body *Term) (*Term, error) {
	answerCopy := make([]Declaration, len(answers))
	for i, answer := range answers {
		answerCopy[i] = Declaration{ID: answer.ID, Type: *cloneTypeValue(answer.Type)}
	}
	result := &Term{Kind: TermAsk, Answers: answerCopy, Restrictor: restrictor, Body: body}
	if _, err := checkExpression(context, result); err != nil {
		return nil, err
	}
	return result, nil
}

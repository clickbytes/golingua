package algebra

import (
	"fmt"
	"strconv"

	"github.com/clickbytes/golingua/internal/status"
)

// SubstituteParameter replaces free Param occurrences with the same ID. It
// does not rename binders, references, or holes that share the parameter ID.
// The replacement's available syntactic type must match each occurrence; a
// final algebra.Check with the caller's declarations remains authoritative.
func SubstituteParameter(term *Term, id string, replacement *Term, limits Limits) (*Term, error) {
	limits, err := NormalizeLimits(limits)
	if err != nil {
		return nil, err
	}
	if !validText(id, 256) {
		return nil, invalid("INVALID_PARAMETER_ID", "parameter_id", "parameter ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	if term == nil || replacement == nil {
		return nil, invalid("MISSING_TERM", "term", "term and replacement are required")
	}
	if replacement.Kind == TermParam && replacement.ID == id {
		return copyTermDetached(term, limits)
	}
	replacementType, err := standaloneTermType(replacement, limits)
	if err != nil {
		return nil, err
	}
	info, err := inspectTerm(term, limits)
	if err != nil {
		return nil, err
	}
	if _, present := info.parameters[id]; !present {
		return copyTermDetached(term, limits)
	}
	avoid := copySet(info.allIDs)
	replacementInfo, err := inspectTerm(replacement, limits)
	if err != nil {
		return nil, err
	}
	for freeID := range replacementInfo.freeVars {
		avoid[freeID] = struct{}{}
	}
	state := newCopyState(limits, avoid, replacementInfo.freeVars, false)
	copy, err := state.walk(term, "expression", 1, 0, 0, id, replacement, &replacementType)
	if err != nil {
		return nil, err
	}
	return copy, nil
}

// SubstituteBinder replaces free Var occurrences for one binder ID in the
// supplied body. The removed binder's declared type is not present in this
// API; callers must run algebra.Check on the resulting enclosing term.
func SubstituteBinder(body *Term, binderID string, replacement *Term, limits Limits) (*Term, error) {
	return SubstituteBinderReserved(body, binderID, replacement, nil, limits)
}

// SubstituteBinderReserved replaces free Var occurrences for one binder ID
// while reserving IDs owned by the surrounding term. Those IDs are considered
// before capture-avoidance freshening or insertion, so generated binder names
// cannot collide with binders outside body.
func SubstituteBinderReserved(body *Term, binderID string, replacement *Term, reservedIDs []string, limits Limits) (*Term, error) {
	limits, err := NormalizeLimits(limits)
	if err != nil {
		return nil, err
	}
	if !validText(binderID, 256) {
		return nil, invalid("INVALID_BINDER_ID", "binder_id", "binder ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	if body == nil || replacement == nil {
		return nil, invalid("MISSING_TERM", "term", "body and replacement are required")
	}
	if len(reservedIDs) > limits.MaxNodes {
		return nil, resourceLimit("reserved_ids", "reserved ID count exceeds the configured node limit")
	}
	reserved := make(map[string]struct{}, len(reservedIDs))
	for i, id := range reservedIDs {
		if !validText(id, 256) {
			return nil, invalid("INVALID_RESERVED_ID", fmt.Sprintf("reserved_ids[%d]", i), "reserved ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		reserved[id] = struct{}{}
	}
	info, err := inspectTerm(body, limits)
	if err != nil {
		return nil, err
	}
	if _, present := info.freeVars[binderID]; !present {
		return copyTermDetached(body, limits)
	}
	replacementInfo, err := inspectTerm(replacement, limits)
	if err != nil {
		return nil, err
	}
	free := replacementInfo.freeVars
	avoid := copySet(info.allIDs)
	for id := range reserved {
		avoid[id] = struct{}{}
	}
	for freeID := range free {
		avoid[freeID] = struct{}{}
	}
	state := newCopyState(limits, avoid, free, false)
	return state.walk(body, "expression", 1, 0, 0, binderID, replacement, nil)
}

// FreshCopy deep-copies a term while assigning every Lambda/Exists binder a
// fresh ID. Free variables and Param/Ref/Hole IDs keep their namespaces and
// values; the returned map covers only binder IDs.
func FreshCopy(term *Term, reservedIDs []string, limits Limits) (*Term, map[string]string, error) {
	limits, err := NormalizeLimits(limits)
	if err != nil {
		return nil, nil, err
	}
	if term == nil {
		return nil, nil, invalid("MISSING_TERM", "term", "term is required")
	}
	info, err := inspectTerm(term, limits)
	if err != nil {
		return nil, nil, err
	}
	avoid := copySet(info.allIDs)
	for _, id := range reservedIDs {
		if !validText(id, 256) {
			return nil, nil, invalid("INVALID_RESERVED_ID", "reserved_ids", "reserved IDs must be non-empty valid UTF-8 and at most 256 bytes")
		}
		avoid[id] = struct{}{}
	}
	state := newCopyState(limits, avoid, nil, true)
	copy, err := state.walk(term, "expression", 1, 0, 0, "", nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return copy, state.binderMap, nil
}

type termInfo struct {
	allIDs     map[string]struct{}
	binderIDs  map[string]struct{}
	freeVars   map[string]struct{}
	parameters map[string]struct{}
}

func inspectTerm(term *Term, limits Limits) (termInfo, error) {
	info := termInfo{allIDs: make(map[string]struct{}), binderIDs: make(map[string]struct{}), freeVars: make(map[string]struct{}), parameters: make(map[string]struct{})}
	active := make(map[*Term]bool)
	bound := make(map[string]int)
	nodes := 0
	var walk func(*Term, string, int) error
	walk = func(value *Term, path string, depth int) error {
		if value == nil {
			return nil
		}
		if depth > limits.MaxDepth {
			return resourceLimit(path, fmt.Sprintf("expression depth exceeds limit %d", limits.MaxDepth))
		}
		if active[value] {
			return invalid("CYCLIC_TERM", path, "expression contains a pointer cycle")
		}
		nodes++
		if nodes > limits.MaxNodes {
			return resourceLimit(path, fmt.Sprintf("expression node count exceeds limit %d", limits.MaxNodes))
		}
		active[value] = true
		defer delete(active, value)
		if value.Kind == TermParam && value.ID != "" {
			info.parameters[value.ID] = struct{}{}
		}
		if value.Kind == TermVar {
			if !validText(value.ID, 256) {
				return invalid("INVALID_BINDER_ID", path+".id", "variable ID must be non-empty valid UTF-8 and at most 256 bytes")
			}
			info.allIDs[value.ID] = struct{}{}
			if bound[value.ID] == 0 {
				info.freeVars[value.ID] = struct{}{}
			}
		}
		if value.Kind == TermLambda || value.Kind == TermExists || value.Kind == TermQuantifier {
			if !validText(value.ID, 256) {
				return invalid("INVALID_BINDER_ID", path+".id", "binder ID must be non-empty valid UTF-8 and at most 256 bytes")
			}
			if _, exists := info.binderIDs[value.ID]; exists {
				return invalid("DUPLICATE_BINDER_ID", path+".id", "binder IDs must be unique in a copied term")
			}
			info.binderIDs[value.ID] = struct{}{}
			info.allIDs[value.ID] = struct{}{}
			bound[value.ID]++
			if err := walk(value.Restrictor, path+".restrictor", depth+1); err != nil {
				return err
			}
			if err := walk(value.Body, path+".body", depth+1); err != nil {
				return err
			}
			bound[value.ID]--
			return nil
		}
		if value.Kind == TermAsk {
			if len(value.Answers) > limits.MaxPorts {
				return resourceLimit(path+".answers", fmt.Sprintf("answer count %d exceeds limit %d", len(value.Answers), limits.MaxPorts))
			}
			for i, answer := range value.Answers {
				if !validText(answer.ID, 256) {
					return invalid("INVALID_ANSWER_ID", fmt.Sprintf("%s.answers[%d].id", path, i), "answer ID must be non-empty valid UTF-8 and at most 256 bytes")
				}
				if !isDataType(answer.Type) {
					return invalid("INVALID_ANSWER_TYPE", fmt.Sprintf("%s.answers[%d].type", path, i), "Ask answer variable must have a data sort type")
				}
				if _, exists := info.binderIDs[answer.ID]; exists {
					return invalid("DUPLICATE_BINDER_ID", fmt.Sprintf("%s.answers[%d].id", path, i), "binder IDs must be unique in a copied term")
				}
				info.binderIDs[answer.ID] = struct{}{}
				info.allIDs[answer.ID] = struct{}{}
				bound[answer.ID]++
			}
			if err := walk(value.Restrictor, path+".restrictor", depth+1); err != nil {
				return err
			}
			if err := walk(value.Body, path+".body", depth+1); err != nil {
				return err
			}
			for _, answer := range value.Answers {
				bound[answer.ID]--
			}
			return nil
		}
		for i, arg := range value.Args {
			if err := walk(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
				return err
			}
		}
		for i, operand := range value.Operands {
			if err := walk(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
		if err := walk(value.Restrictor, path+".restrictor", depth+1); err != nil {
			return err
		}
		if err := walk(value.Body, path+".body", depth+1); err != nil {
			return err
		}
		if err := walk(value.Function, path+".function", depth+1); err != nil {
			return err
		}
		return walk(value.Argument, path+".argument", depth+1)
	}
	if err := walk(term, "expression", 1); err != nil {
		return termInfo{}, err
	}
	return info, nil
}

func copySet(source map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(source))
	for id := range source {
		result[id] = struct{}{}
	}
	return result
}

type copyState struct {
	limits              Limits
	avoid               map[string]struct{}
	freeReplacementVars map[string]struct{}
	freshenAll          bool
	active              map[*Term]bool
	renamed             map[string]string
	generated           map[string]struct{}
	binderMap           map[string]string
	nodes               int
	typeNodes           int
}

func newCopyState(limits Limits, avoid, freeVars map[string]struct{}, freshenAll bool) *copyState {
	return &copyState{limits: limits, avoid: avoid, freeReplacementVars: freeVars, freshenAll: freshenAll, active: make(map[*Term]bool), renamed: make(map[string]string), generated: make(map[string]struct{}), binderMap: make(map[string]string)}
}

func (state *copyState) walk(value *Term, path string, depth, shadow int, typeDepth int, target string, replacement *Term, replacementType *Type) (*Term, error) {
	if value == nil {
		return nil, nil
	}
	if depth > state.limits.MaxDepth {
		return nil, resourceLimit(path, fmt.Sprintf("expression depth exceeds limit %d", state.limits.MaxDepth))
	}
	if state.active[value] {
		return nil, invalid("CYCLIC_TERM", path, "expression contains a pointer cycle")
	}
	if value.Kind == TermParam && replacementType != nil && value.ID == target {
		if value.Type == nil || !TypeEqual(*value.Type, *replacementType) {
			return nil, invalid("SUBSTITUTION_TYPE_MISMATCH", path+".type", "replacement type differs from parameter type")
		}
		return state.cloneInserted(replacement, path, depth)
	}
	if value.Kind == TermVar && target != "" && value.ID == target && shadow == 0 {
		return state.cloneInserted(replacement, path, depth)
	}
	state.nodes++
	if state.nodes > state.limits.MaxNodes {
		return nil, resourceLimit(path, fmt.Sprintf("expression node count exceeds limit %d", state.limits.MaxNodes))
	}
	state.active[value] = true
	defer delete(state.active, value)
	copyValue := *value
	copyValue.Type = nil
	copyValue.Count = nil
	if value.Count != nil {
		count := *value.Count
		copyValue.Count = &count
	}
	if value.Type != nil {
		typeCopy, err := state.cloneType(value.Type, path+".type", 1, make(map[*Type]bool))
		if err != nil {
			return nil, err
		}
		copyValue.Type = typeCopy
	}
	if value.Kind == TermVar {
		if renamed, ok := state.renamed[value.ID]; ok {
			copyValue.ID = renamed
		}
	}
	if value.Kind == TermLambda || value.Kind == TermExists || value.Kind == TermQuantifier || value.Kind == TermAsk {
		ids := []string{}
		if value.Kind == TermAsk {
			copyValue.Answers = make([]Declaration, len(value.Answers))
			for i, answer := range value.Answers {
				copyValue.Answers[i] = answer
				typ, err := state.cloneType(&answer.Type, fmt.Sprintf("%s.answers[%d].type", path, i), 1, make(map[*Type]bool))
				if err != nil {
					return nil, err
				}
				copyValue.Answers[i].Type = *typ
				ids = append(ids, answer.ID)
			}
		} else {
			ids = append(ids, value.ID)
		}
		type renameRestore struct {
			id, previous string
			existed      bool
		}
		restores := make([]renameRestore, 0, len(ids))
		nextShadow := shadow
		for i, original := range ids {
			fresh := state.freshenAll
			if _, capture := state.freeReplacementVars[original]; capture {
				fresh = true
			}
			newID := original
			if fresh {
				newID = state.freshID(original)
				state.binderMap[original] = newID
			}
			if value.Kind == TermAsk {
				copyValue.Answers[i].ID = newID
			} else {
				copyValue.ID = newID
			}
			previous, existed := state.renamed[original]
			restores = append(restores, renameRestore{id: original, previous: previous, existed: existed})
			state.renamed[original] = newID
			if target != "" && original == target {
				nextShadow++
			}
		}
		restrictor, err := state.walk(value.Restrictor, path+".restrictor", depth+1, nextShadow, typeDepth, target, replacement, replacementType)
		if err == nil {
			copyValue.Restrictor = restrictor
		}
		var body *Term
		if err == nil {
			body, err = state.walk(value.Body, path+".body", depth+1, nextShadow, typeDepth, target, replacement, replacementType)
		}
		for _, restore := range restores {
			if restore.existed {
				state.renamed[restore.id] = restore.previous
			} else {
				delete(state.renamed, restore.id)
			}
		}
		if err != nil {
			return nil, err
		}
		copyValue.Body = body
		return &copyValue, nil
	}
	copyValue.Args = make([]Argument, len(value.Args))
	for i, argument := range value.Args {
		copyValue.Args[i] = argument
		cloned, err := state.walk(argument.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Args[i].Value = cloned
	}
	copyValue.Operands = make([]*Term, len(value.Operands))
	for i, operand := range value.Operands {
		cloned, err := state.walk(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Operands[i] = cloned
	}
	copyValue.Body = nil
	if value.Body != nil {
		body, err := state.walk(value.Body, path+".body", depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Body = body
	}
	copyValue.Restrictor = nil
	if value.Restrictor != nil {
		restrictor, err := state.walk(value.Restrictor, path+".restrictor", depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Restrictor = restrictor
	}
	if value.Function != nil {
		function, err := state.walk(value.Function, path+".function", depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Function = function
	}
	if value.Argument != nil {
		argument, err := state.walk(value.Argument, path+".argument", depth+1, shadow, typeDepth, target, replacement, replacementType)
		if err != nil {
			return nil, err
		}
		copyValue.Argument = argument
	}
	return &copyValue, nil
}

func (state *copyState) freshID(original string) string {
	for suffix := 1; ; suffix++ {
		candidate := original + "_fresh_" + strconv.Itoa(suffix)
		if len(candidate) > 256 {
			candidate = "binder_fresh_" + strconv.Itoa(suffix)
		}
		if _, used := state.avoid[candidate]; used {
			continue
		}
		if _, used := state.generated[candidate]; used {
			continue
		}
		state.generated[candidate] = struct{}{}
		state.avoid[candidate] = struct{}{}
		return candidate
	}
}

func (state *copyState) cloneType(value *Type, path string, depth int, active map[*Type]bool) (*Type, error) {
	if value == nil {
		return nil, nil
	}
	if depth > state.limits.MaxTypeDepth {
		return nil, resourceLimit(path, fmt.Sprintf("type depth exceeds limit %d", state.limits.MaxTypeDepth))
	}
	if active[value] {
		return nil, invalid("CYCLIC_TYPE", path, "type contains a pointer cycle")
	}
	state.typeNodes++
	if state.typeNodes > state.limits.MaxNodes*state.limits.MaxTypeDepth {
		return nil, resourceLimit(path, "type node checks exceed configured bound")
	}
	active[value] = true
	defer delete(active, value)
	copyValue := *value
	var err error
	copyValue.From, err = state.cloneType(value.From, path+".from", depth+1, active)
	if err != nil {
		return nil, err
	}
	copyValue.To, err = state.cloneType(value.To, path+".to", depth+1, active)
	if err != nil {
		return nil, err
	}
	return &copyValue, nil
}

func (state *copyState) cloneInserted(term *Term, path string, depth int) (*Term, error) {
	info, err := inspectTerm(term, state.limits)
	if err != nil {
		return nil, err
	}
	avoid := copySet(state.avoid)
	for id := range info.allIDs {
		avoid[id] = struct{}{}
		state.avoid[id] = struct{}{}
	}
	inserted := newCopyState(state.limits, avoid, info.freeVars, true)
	inserted.nodes = state.nodes
	inserted.typeNodes = state.typeNodes
	copy, err := inserted.walk(term, path+".replacement", depth, 0, 0, "", nil, nil)
	state.nodes = inserted.nodes
	state.typeNodes = inserted.typeNodes
	if err != nil {
		return nil, err
	}
	for id := range inserted.generated {
		state.avoid[id] = struct{}{}
	}
	return copy, nil
}

func copyTermDetached(term *Term, limits Limits) (*Term, error) {
	state := newCopyState(limits, make(map[string]struct{}), nil, false)
	return state.walk(term, "expression", 1, 0, 0, "", nil, nil)
}

func standaloneTermType(term *Term, limits Limits) (Type, error) {
	active := make(map[*Term]bool)
	bound := make(map[string]Type)
	nodes := 0
	var infer func(*Term, string, int) (Type, error)
	infer = func(value *Term, path string, depth int) (Type, error) {
		if value == nil {
			return Type{}, invalid("MISSING_TERM", path, "term is required")
		}
		if depth > limits.MaxDepth {
			return Type{}, resourceLimit(path, fmt.Sprintf("expression depth exceeds limit %d", limits.MaxDepth))
		}
		if active[value] {
			return Type{}, invalid("CYCLIC_TERM", path, "expression contains a pointer cycle")
		}
		nodes++
		if nodes > limits.MaxNodes {
			return Type{}, resourceLimit(path, fmt.Sprintf("expression node count exceeds limit %d", limits.MaxNodes))
		}
		active[value] = true
		defer delete(active, value)
		switch value.Kind {
		case TermVar:
			if typ, ok := bound[value.ID]; ok {
				return typ, nil
			}
			return Type{}, &Error{Status: status.OutcomeUnsupported, Code: "UNRESOLVED_VARIABLE_TYPE", Path: path, Message: "a free Var type requires an enclosing checker context"}
		case TermParam, TermRef, TermHole:
			if value.Type == nil {
				return Type{}, invalid("MISSING_TYPE", path+".type", "term type is required")
			}
			copyType, err := cloneStandaloneType(value.Type, limits)
			if err != nil {
				return Type{}, err
			}
			return copyType, nil
		case TermAtom:
			for i, arg := range value.Args {
				if _, err := infer(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
					return Type{}, err
				}
			}
			return Type{Kind: TypeProp}, nil
		case TermAnd:
			if len(value.Operands) == 0 {
				return Type{}, invalid("EMPTY_CONJUNCTION", path+".operands", "And requires at least one operand")
			}
			for i, operand := range value.Operands {
				typ, err := infer(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1)
				if err != nil {
					return Type{}, err
				}
				if typ.Kind != TypeProp {
					return Type{}, invalid("EXPECTED_PROP", path, "And operands must have type Prop")
				}
			}
			return Type{Kind: TypeProp}, nil
		case TermExists:
			if value.Type == nil {
				return Type{}, invalid("MISSING_TYPE", path+".type", "Exists binder type is required")
			}
			bound[value.ID] = *value.Type
			bodyType, err := infer(value.Body, path+".body", depth+1)
			delete(bound, value.ID)
			if err != nil {
				return Type{}, err
			}
			if bodyType.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "Exists body must have type Prop")
			}
			return Type{Kind: TypeProp}, nil
		case TermNot:
			body, err := infer(value.Body, path+".body", depth+1)
			if err != nil {
				return Type{}, err
			}
			if body.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "Not body must have type Prop")
			}
			return Type{Kind: TypeProp}, nil
		case TermModal:
			body, err := infer(value.Body, path+".body", depth+1)
			if err != nil {
				return Type{}, err
			}
			if body.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "Modal body must have type Prop")
			}
			return Type{Kind: TypeProp}, nil
		case TermQuantifier:
			if value.Type == nil || !isDataType(*value.Type) {
				return Type{}, invalid("INVALID_QUANTIFIER_BINDER_TYPE", path+".type", "quantifier binder must have a data sort type")
			}
			if value.Quantifier == "exactly" || value.Quantifier == "at_least" || value.Quantifier == "at_most" {
				if value.Count == nil {
					return Type{}, invalid("MISSING_QUANTIFIER_COUNT", path+".count", "numeric quantifier requires a count")
				}
				if *value.Count < 0 {
					return Type{}, invalid("INVALID_QUANTIFIER_COUNT", path+".count", "count must be non-negative")
				}
			} else if (value.Quantifier == "some" || value.Quantifier == "all") && value.Count != nil {
				return Type{}, invalid("INVALID_QUANTIFIER_COUNT", path+".count", "some and all do not accept count")
			} else if value.Quantifier != "some" && value.Quantifier != "all" {
				return Type{}, invalid("INVALID_QUANTIFIER_KIND", path+".quantifier", "quantifier kind is unsupported")
			}
			bound[value.ID] = *value.Type
			if value.Restrictor != nil {
				t, err := infer(value.Restrictor, path+".restrictor", depth+1)
				if err != nil {
					delete(bound, value.ID)
					return Type{}, err
				}
				if t.Kind != TypeProp {
					delete(bound, value.ID)
					return Type{}, invalid("EXPECTED_PROP", path+".restrictor", "quantifier restrictor must have type Prop")
				}
			}
			body, err := infer(value.Body, path+".body", depth+1)
			delete(bound, value.ID)
			if err != nil {
				return Type{}, err
			}
			if body.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "quantifier body must have type Prop")
			}
			return Type{Kind: TypeProp}, nil
		case TermAsk:
			if len(value.Answers) > limits.MaxPorts {
				return Type{}, resourceLimit(path+".answers", fmt.Sprintf("answer count %d exceeds limit %d", len(value.Answers), limits.MaxPorts))
			}
			for i, answer := range value.Answers {
				if !isDataType(answer.Type) {
					return Type{}, invalid("INVALID_ANSWER_TYPE", fmt.Sprintf("%s.answers[%d].type", path, i), "Ask answer variable must have a data sort type")
				}
				bound[answer.ID] = answer.Type
			}
			if value.Restrictor != nil {
				t, err := infer(value.Restrictor, path+".restrictor", depth+1)
				if err != nil {
					for _, a := range value.Answers {
						delete(bound, a.ID)
					}
					return Type{}, err
				}
				if t.Kind != TypeProp {
					for _, a := range value.Answers {
						delete(bound, a.ID)
					}
					return Type{}, invalid("EXPECTED_PROP", path+".restrictor", "Ask restrictor must have type Prop")
				}
			}
			body, err := infer(value.Body, path+".body", depth+1)
			for _, a := range value.Answers {
				delete(bound, a.ID)
			}
			if err != nil {
				return Type{}, err
			}
			if body.Kind != TypeProp {
				return Type{}, invalid("EXPECTED_PROP", path+".body", "Ask body must have type Prop")
			}
			return Type{Kind: TypeQuestion}, nil
		case TermLambda:
			if value.Type == nil {
				return Type{}, invalid("MISSING_TYPE", path+".type", "Lambda binder type is required")
			}
			argType, err := cloneStandaloneType(value.Type, limits)
			if err != nil {
				return Type{}, err
			}
			bound[value.ID] = *value.Type
			bodyType, err := infer(value.Body, path+".body", depth+1)
			delete(bound, value.ID)
			if err != nil {
				return Type{}, err
			}
			return Type{Kind: TypeArrow, From: &argType, To: &bodyType}, nil
		case TermApply:
			functionType, err := infer(value.Function, path+".function", depth+1)
			if err != nil {
				return Type{}, err
			}
			argumentType, err := infer(value.Argument, path+".argument", depth+1)
			if err != nil {
				return Type{}, err
			}
			if functionType.Kind != TypeArrow || functionType.From == nil || functionType.To == nil {
				return Type{}, invalid("NOT_A_FUNCTION", path+".function", "application target must have arrow type")
			}
			if !TypeEqual(argumentType, *functionType.From) {
				return Type{}, invalid("ARGUMENT_TYPE_MISMATCH", path+".argument", "argument type does not match function input type")
			}
			return *functionType.To, nil
		default:
			return Type{}, invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind is unsupported")
		}
	}
	return infer(term, "replacement", 1)
}

func cloneStandaloneType(value *Type, limits Limits) (Type, error) {
	state := newCopyState(limits, make(map[string]struct{}), nil, false)
	copy, err := state.cloneType(value, "type", 1, make(map[*Type]bool))
	if err != nil {
		return Type{}, err
	}
	if copy == nil {
		return Type{}, invalid("MISSING_TYPE", "type", "type is required")
	}
	return *copy, nil
}

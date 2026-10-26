package algebra

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/status"
)

// ApplyRewrite attempts one licensed rewrite at an explicit path in a checked
// full-root expression. A false changed result means the rule is inapplicable.
func ApplyRewrite(environment RewriteEnvironment, profile RewriteProfile, before *Term, rule RewriteRule, path []int) (RewriteStep, bool, error) {
	if err := ValidateRewriteEnvironment(environment); err != nil {
		return RewriteStep{}, false, err
	}
	if err := ValidateRewriteProfile(profile); err != nil {
		return RewriteStep{}, false, err
	}
	if !rewriteLicensed(profile, rule) {
		return RewriteStep{}, false, &Error{Status: status.OutcomeUnsupported, Code: "UNLICENSED_REWRITE_RULE", Path: "rule_id", Message: "rewrite rule is not licensed by the selected profile"}
	}
	limits, checkedBefore, err := checkRewriteRoot(environment, before)
	if err != nil {
		return RewriteStep{}, false, err
	}
	if len(path) > limits.MaxDepth {
		return RewriteStep{}, false, resourceLimit("path", "rewrite path exceeds expression depth limit")
	}
	if _, err := rewriteTermAt(before, path); err != nil {
		return RewriteStep{}, false, err
	}
	rootIDs, err := rewriteReservedIDs(before, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	after, err := copyTermDetached(before, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	workingTarget, err := rewriteTermAt(after, path)
	if err != nil {
		return RewriteStep{}, false, err
	}
	var replacement *Term
	var binderMapping []BinderChange
	var sideConditions []string
	var alphaRenames map[string]string
	switch rule {
	case RewriteAlphaV1:
		if len(path) != 0 {
			return RewriteStep{}, false, invalid("ALPHA_PATH_NOT_ROOT", "path", "alpha normalization applies only to the whole root")
		}
		after, alphaRenames, err = canonicalizeBinders(after, limits)
		if err != nil {
			return RewriteStep{}, false, err
		}
		sideConditions = []string{"binder IDs use deterministic preorder names", "bound Var occurrences follow their exact binder", "parameter and reference IDs are preserved"}
	case RewriteBetaV1:
		if workingTarget.Kind != TermApply || workingTarget.Function == nil || workingTarget.Function.Kind != TermLambda {
			return RewriteStep{}, false, nil
		}
		lambda := workingTarget.Function
		bodyInfo, inspectErr := inspectTerm(lambda.Body, limits)
		if inspectErr != nil {
			return RewriteStep{}, false, inspectErr
		}
		if _, used := bodyInfo.freeVars[lambda.ID]; !used {
			dependencies, depErr := rewriteDependencies(workingTarget.Argument, limits)
			if depErr != nil {
				return RewriteStep{}, false, depErr
			}
			if len(dependencies.parameters)+len(dependencies.references)+len(dependencies.holes) > 0 {
				return RewriteStep{}, false, &Error{Status: status.OutcomeUnsupported, Code: "DEPENDENCY_ERASING_BETA", Path: "path", Message: "beta reduction would erase a free parameter, reference, or hole dependency"}
			}
		}
		freshBody, renameMap, copyErr := FreshCopy(lambda.Body, rootIDs, limits)
		if copyErr != nil {
			return RewriteStep{}, false, copyErr
		}
		reserved := append([]string(nil), rootIDs...)
		reserved = append(reserved, binderIDsInPreorder(freshBody)...)
		replaced, subErr := SubstituteBinderReserved(freshBody, lambda.ID, workingTarget.Argument, reserved, limits)
		if subErr != nil {
			return RewriteStep{}, false, subErr
		}
		replacement = replaced
		alphaRenames = renameMap
		sideConditions = []string{"redex function is a Lambda", "substitution is capture avoiding in the enclosing root scope", "free parameter and reference dependencies are preserved"}
	case RewriteAliasV1:
		if workingTarget.Kind != TermAtom {
			return RewriteStep{}, false, nil
		}
		alias, found := rewriteAlias(environment.Aliases, workingTarget.Symbol)
		if !found {
			return RewriteStep{}, false, nil
		}
		if len(workingTarget.Args) != len(alias.Ports) || len(alias.Interface) != len(alias.Ports) {
			return RewriteStep{}, false, invalid("ALIAS_APPLICATION_SHAPE", "path", "alias atom does not match its declared ordered interface")
		}
		freshAlias, renameMap, copyErr := FreshCopy(alias.Expression, rootIDs, limits)
		if copyErr != nil {
			return RewriteStep{}, false, copyErr
		}
		_, body, peelErr := peelLeading(freshAlias, len(alias.Interface))
		if peelErr != nil {
			return RewriteStep{}, false, peelErr
		}
		reserved := append([]string(nil), rootIDs...)
		reserved = append(reserved, binderIDsInPreorder(freshAlias)...)
		for i, binding := range alias.Interface {
			if workingTarget.Args[i].Role != binding.Port || binding.Port != alias.Ports[i].Role {
				return RewriteStep{}, false, invalid("ALIAS_APPLICATION_ROLE", fmt.Sprintf("path.args[%d].role", i), "alias argument roles must match the ordered interface")
			}
			freshBinder, exists := renameMap[binding.Binder]
			if !exists {
				return RewriteStep{}, false, invalid("ALIAS_BINDER_MAPPING", fmt.Sprintf("aliases.%s.interface[%d]", alias.ID, i), "fresh alias copy is missing an interface binder")
			}
			body, err = SubstituteBinderReserved(body, freshBinder, workingTarget.Args[i].Value, reserved, limits)
			if err != nil {
				return RewriteStep{}, false, err
			}
		}
		replacement = body
		sideConditions = []string{"alias definition is supplied by the validated rewrite environment: " + alias.ID, "alias ports and arguments match in declared order", "alias binders are freshened against the complete root", "alias body is instantiated by capture-avoiding substitution"}
		for _, oldBinder := range binderIDsInPreorder(alias.Expression) {
			sideConditions = append(sideConditions, fmt.Sprintf("alias binder correspondence %q -> %q", oldBinder, renameMap[oldBinder]))
		}
		for i, binding := range alias.Interface {
			sideConditions = append(sideConditions, fmt.Sprintf("alias interface binder %q -> %q is removed at port %q and instantiated from argument %d", binding.Binder, renameMap[binding.Binder], binding.Port, i))
		}
	case RewriteAndFlattenV1:
		if workingTarget.Kind != TermAnd {
			return RewriteStep{}, false, nil
		}
		flattened := make([]*Term, 0, len(workingTarget.Operands))
		for _, operand := range workingTarget.Operands {
			if operand.Kind == TermAnd {
				flattened = append(flattened, operand.Operands...)
			} else {
				flattened = append(flattened, operand)
			}
		}
		if len(flattened) == len(workingTarget.Operands) {
			return RewriteStep{}, false, nil
		}
		replacement = &Term{Kind: TermAnd, Operands: flattened}
		sideConditions = []string{"only directly nested And operands are flattened", "no binder, quantifier, discourse, or modal boundary is crossed"}
	case RewriteAndOrderV1:
		if workingTarget.Kind != TermAnd {
			return RewriteStep{}, false, nil
		}
		outer, scopeErr := rewriteOuterScope(before, path)
		if scopeErr != nil {
			return RewriteStep{}, false, scopeErr
		}
		keys := make([]string, len(workingTarget.Operands))
		ordered := append([]*Term(nil), workingTarget.Operands...)
		for i, operand := range ordered {
			keys[i], err = rewriteOrderKey(operand, outer, limits)
			if err != nil {
				return RewriteStep{}, false, err
			}
		}
		indices := make([]int, len(ordered))
		for i := range indices {
			indices[i] = i
		}
		sort.SliceStable(indices, func(i, j int) bool { return keys[indices[i]] < keys[indices[j]] })
		sortedOperands := make([]*Term, len(ordered))
		changed := false
		for i, index := range indices {
			sortedOperands[i] = ordered[index]
			if i != index {
				changed = true
			}
		}
		if !changed {
			return RewriteStep{}, false, nil
		}
		replacement = &Term{Kind: TermAnd, Operands: sortedOperands}
		sideConditions = []string{"And operands use stable alpha-invariant structural ordering", "labels are ignored and roles, references, parameters, and order-sensitive constructors remain explicit"}
	case RewriteAndDeduplicateV1:
		if workingTarget.Kind != TermAnd {
			return RewriteStep{}, false, nil
		}
		kept := make([]*Term, 0, len(workingTarget.Operands))
		for _, operand := range workingTarget.Operands {
			binderFree, scanErr := rewriteBinderFree(operand, limits)
			if scanErr != nil {
				return RewriteStep{}, false, scanErr
			}
			duplicate := false
			if binderFree {
				for _, prior := range kept {
					priorFree, _ := rewriteBinderFree(prior, limits)
					if priorFree {
						equal, eqErr := RewriteTermEqual(operand, prior, limits)
						if eqErr != nil {
							return RewriteStep{}, false, eqErr
						}
						if equal {
							duplicate = true
							break
						}
					}
				}
			}
			if !duplicate {
				kept = append(kept, operand)
			}
		}
		if len(kept) == len(workingTarget.Operands) {
			return RewriteStep{}, false, nil
		}
		replacement = &Term{Kind: TermAnd, Operands: kept}
		sideConditions = []string{"only structurally identical binder-free operands are deduplicated", "reference and event IDs remain identity sensitive", "singleton And remains an And"}
	default:
		return RewriteStep{}, false, &Error{Status: status.OutcomeUnsupported, Code: "UNKNOWN_REWRITE_RULE", Path: "rule_id", Message: "rewrite rule is not supported"}
	}

	if replacement != nil {
		after, err = replaceRewritePath(after, path, replacement)
		if err != nil {
			return RewriteStep{}, false, err
		}
	}
	checkedAfter, err := checkRewriteOutput(environment, after)
	if err != nil {
		return RewriteStep{}, false, err
	}
	if !TypeEqual(checkedBefore.Type, checkedAfter.Type) {
		return RewriteStep{}, false, invalid("REWRITE_TYPE_CHANGED", "after", "rewrite changed the full-root result type")
	}
	beforeDeps, err := rewriteDependencies(before, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	afterDeps, err := rewriteDependencies(after, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	if !sameStringSet(beforeDeps.parameters, afterDeps.parameters) || !sameStringSet(beforeDeps.references, afterDeps.references) {
		return RewriteStep{}, false, &Error{Status: status.OutcomeUnsupported, Code: "DEPENDENCY_ERASING_REWRITE", Path: "after", Message: "rewrite would erase or introduce a free parameter or reference"}
	}
	equal, err := RewriteTermEqual(before, after, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	if equal {
		return RewriteStep{}, false, nil
	}
	binderMapping, err = rewriteBinderChanges(before, after, alphaRenames, rule)
	if err != nil {
		return RewriteStep{}, false, err
	}
	beforeCopy, err := copyTermDetached(before, limits)
	if err != nil {
		return RewriteStep{}, false, err
	}
	return RewriteStep{
		RuleID: rule, Path: append([]int{}, path...), Before: beforeCopy, After: after,
		BinderMapping: binderMapping, ParameterMapping: Substitution{}, SideConditions: sideConditions,
		SnapshotID: environment.SnapshotID, ProfileID: profile.ID,
		BeforeNodes: checkedBefore.NodeCount, AfterNodes: checkedAfter.NodeCount,
	}, true, nil
}

// VerifyRewriteStep independently replays one named, licensed rule at its
// recorded path and compares every certificate claim with the recomputed step.
func VerifyRewriteStep(environment RewriteEnvironment, profile RewriteProfile, step RewriteStep) error {
	if step.SnapshotID != environment.SnapshotID || step.ProfileID != profile.ID {
		return invalid("REWRITE_CONTEXT_MISMATCH", "step", "step snapshot and profile IDs must match the supplied environment and profile")
	}
	if len(step.ParameterMapping) != 0 {
		return invalid("REWRITE_PARAMETER_MAPPING", "step.parameter_mapping", "rewrite rules in this kernel do not substitute free parameters")
	}
	expected, changed, err := ApplyRewrite(environment, profile, step.Before, step.RuleID, step.Path)
	if err != nil {
		return err
	}
	if !changed {
		return invalid("REWRITE_NOT_APPLICABLE", "step", "recorded rewrite is not applicable or makes no change")
	}
	if step.Before == nil || step.After == nil {
		return invalid("MISSING_REWRITE_TERM", "step", "full-root before and after terms are required")
	}
	beforeEqual, err := RewriteTermEqual(step.Before, expected.Before, environment.Document.Limits)
	if err != nil {
		return err
	}
	afterEqual, err := RewriteTermEqual(step.After, expected.After, environment.Document.Limits)
	if err != nil {
		return err
	}
	if !beforeEqual || !afterEqual {
		return invalid("REWRITE_RESULT_MISMATCH", "step.after", "recorded full-root result differs from replay")
	}
	if step.RuleID != expected.RuleID || !sameIntSlice(step.Path, expected.Path) || !sameBinderChanges(step.BinderMapping, expected.BinderMapping) || !sameStringSlice(step.SideConditions, expected.SideConditions) || step.SnapshotID != expected.SnapshotID || step.ProfileID != expected.ProfileID || step.BeforeNodes != expected.BeforeNodes || step.AfterNodes != expected.AfterNodes {
		return invalid("REWRITE_CERTIFICATE_MISMATCH", "step", "recorded rule metadata differs from independent replay")
	}
	return nil
}

// RewriteTermEqual compares terms while ignoring presentation labels. It is
// bounded and preserves every syntactic ID, role, type, and child order.
func RewriteTermEqual(left, right *Term, limits Limits) (bool, error) {
	limits, err := NormalizeLimits(limits)
	if err != nil {
		return false, err
	}
	if left == nil || right == nil {
		return left == right, nil
	}
	if _, err := inspectTerm(left, limits); err != nil {
		return false, err
	}
	if _, err := inspectTerm(right, limits); err != nil {
		return false, err
	}
	var equalType func(*Type, *Type, int, map[*Type]bool, map[*Type]bool) (bool, error)
	typeVisits := 0
	equalType = func(a, b *Type, depth int, activeA, activeB map[*Type]bool) (bool, error) {
		if a == nil || b == nil {
			return a == b, nil
		}
		if depth > limits.MaxTypeDepth {
			return false, resourceLimit("type", "type depth exceeds configured limit")
		}
		if activeA[a] || activeB[b] {
			return false, invalid("CYCLIC_TYPE", "type", "term equality encountered cyclic type data")
		}
		typeVisits++
		if typeVisits > limits.MaxNodes*limits.MaxTypeDepth {
			return false, resourceLimit("type", "type comparison exceeds configured bound")
		}
		if a.Kind != b.Kind || a.SortID != b.SortID {
			return false, nil
		}
		activeA[a], activeB[b] = true, true
		fromEqual, err := equalType(a.From, b.From, depth+1, activeA, activeB)
		if err == nil && fromEqual {
			var toEqual bool
			toEqual, err = equalType(a.To, b.To, depth+1, activeA, activeB)
			fromEqual = toEqual
		}
		delete(activeA, a)
		delete(activeB, b)
		return fromEqual, err
	}
	activePairs := make(map[[2]*Term]bool)
	termNodes := 0
	var equalTerm func(*Term, *Term, int) (bool, error)
	equalTerm = func(a, b *Term, depth int) (bool, error) {
		if a == nil || b == nil {
			return a == b, nil
		}
		if depth > limits.MaxDepth {
			return false, resourceLimit("term", "term equality depth exceeds configured limit")
		}
		pair := [2]*Term{a, b}
		if activePairs[pair] {
			return false, invalid("CYCLIC_TERM", "term", "term equality encountered cyclic term data")
		}
		termNodes++
		if termNodes > limits.MaxNodes {
			return false, resourceLimit("term", "term equality exceeds configured node bound")
		}
		if a.Kind != b.Kind || a.ID != b.ID || a.Symbol != b.Symbol || a.Quantifier != b.Quantifier || a.Modality != b.Modality || a.ModalProfile != b.ModalProfile ||
			len(a.Args) != len(b.Args) || len(a.Operands) != len(b.Operands) || len(a.Answers) != len(b.Answers) ||
			(a.Count == nil) != (b.Count == nil) || (a.Count != nil && *a.Count != *b.Count) {
			return false, nil
		}
		activePairs[pair] = true
		defer delete(activePairs, pair)
		typesEqual, err := equalType(a.Type, b.Type, 1, make(map[*Type]bool), make(map[*Type]bool))
		if err != nil || !typesEqual {
			return typesEqual, err
		}
		for i := range a.Answers {
			if a.Answers[i].ID != b.Answers[i].ID {
				return false, nil
			}
			answerTypesEqual, answerTypeErr := equalType(&a.Answers[i].Type, &b.Answers[i].Type, 1, make(map[*Type]bool), make(map[*Type]bool))
			if answerTypeErr != nil || !answerTypesEqual {
				return answerTypesEqual, answerTypeErr
			}
		}
		for i := range a.Args {
			if a.Args[i].Role != b.Args[i].Role {
				return false, nil
			}
			ok, e := equalTerm(a.Args[i].Value, b.Args[i].Value, depth+1)
			if e != nil || !ok {
				return ok, e
			}
		}
		for i := range a.Operands {
			ok, e := equalTerm(a.Operands[i], b.Operands[i], depth+1)
			if e != nil || !ok {
				return ok, e
			}
		}
		for _, child := range [][2]*Term{{a.Restrictor, b.Restrictor}, {a.Body, b.Body}, {a.Function, b.Function}, {a.Argument, b.Argument}} {
			ok, e := equalTerm(child[0], child[1], depth+1)
			if e != nil || !ok {
				return ok, e
			}
		}
		return true, nil
	}
	return equalTerm(left, right, 1)
}

func checkRewriteRoot(environment RewriteEnvironment, expression *Term) (Limits, CheckedExpression, error) {
	if expression == nil {
		return Limits{}, CheckedExpression{}, invalid("MISSING_EXPRESSION", "expression", "full-root expression is required")
	}
	limits, err := NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return Limits{}, CheckedExpression{}, err
	}
	document := environment.Document
	document.Expression = expression
	checked, err := Check(document)
	if err != nil {
		return Limits{}, CheckedExpression{}, err
	}
	if checked.Status != status.OutcomeOK || len(checked.Holes) > 0 {
		return Limits{}, CheckedExpression{}, &Error{Status: status.OutcomeIncomplete, Code: "REWRITE_INPUT_INCOMPLETE", Path: "expression", Message: "rewrite kernel refuses unresolved holes"}
	}
	return limits, checked, nil
}

func checkRewriteOutput(environment RewriteEnvironment, expression *Term) (CheckedExpression, error) {
	document := environment.Document
	document.Expression = expression
	checked, err := Check(document)
	if err != nil {
		return CheckedExpression{}, err
	}
	if checked.Status != status.OutcomeOK || len(checked.Holes) > 0 {
		return CheckedExpression{}, &Error{Status: status.OutcomeIncomplete, Code: "REWRITE_OUTPUT_INCOMPLETE", Path: "after", Message: "rewrite output contains unresolved holes"}
	}
	return checked, nil
}

func rewriteLicensed(profile RewriteProfile, rule RewriteRule) bool {
	for _, licensed := range profile.Rules {
		if licensed == rule {
			return true
		}
	}
	return false
}

func rewriteTermAt(root *Term, path []int) (*Term, error) {
	current := root
	for i, index := range path {
		children := rewriteChildren(current)
		if index < 0 || index >= len(children) {
			return nil, invalid("INVALID_REWRITE_PATH", fmt.Sprintf("path[%d]", i), "child index is outside the term")
		}
		current = children[index]
	}
	return current, nil
}

func rewriteChildren(term *Term) []*Term {
	if term == nil {
		return nil
	}
	switch term.Kind {
	case TermAtom:
		out := make([]*Term, len(term.Args))
		for i, arg := range term.Args {
			out[i] = arg.Value
		}
		return out
	case TermAnd:
		return term.Operands
	case TermExists, TermLambda, TermNot, TermModal:
		if term.Body != nil {
			return []*Term{term.Body}
		}
	case TermQuantifier, TermAsk:
		children := make([]*Term, 0, 2)
		if term.Restrictor != nil {
			children = append(children, term.Restrictor)
		}
		if term.Body != nil {
			children = append(children, term.Body)
		}
		return children
	case TermApply:
		out := make([]*Term, 0, 2)
		if term.Function != nil {
			out = append(out, term.Function)
		}
		if term.Argument != nil {
			out = append(out, term.Argument)
		}
		return out
	}
	return nil
}

func replaceRewritePath(root *Term, path []int, replacement *Term) (*Term, error) {
	if len(path) == 0 {
		return replacement, nil
	}
	parent, err := rewriteTermAt(root, path[:len(path)-1])
	if err != nil {
		return nil, err
	}
	index := path[len(path)-1]
	switch parent.Kind {
	case TermAtom:
		if index < 0 || index >= len(parent.Args) {
			break
		}
		parent.Args[index].Value = replacement
		return root, nil
	case TermAnd:
		if index < 0 || index >= len(parent.Operands) {
			break
		}
		parent.Operands[index] = replacement
		return root, nil
	case TermExists, TermLambda, TermNot, TermModal:
		if index == 0 {
			parent.Body = replacement
			return root, nil
		}
	case TermQuantifier, TermAsk:
		if parent.Restrictor != nil {
			if index == 0 {
				parent.Restrictor = replacement
				return root, nil
			}
			index--
		}
		if index == 0 {
			parent.Body = replacement
			return root, nil
		}
	case TermApply:
		if index == 0 {
			parent.Function = replacement
			return root, nil
		}
		if index == 1 {
			parent.Argument = replacement
			return root, nil
		}
	}
	return nil, invalid("INVALID_REWRITE_PATH", fmt.Sprintf("path[%d]", len(path)-1), "child index is outside the term")
}

func rewriteReservedIDs(term *Term, limits Limits) ([]string, error) {
	if _, err := inspectTerm(term, limits); err != nil {
		return nil, err
	}
	ids := make(map[string]struct{})
	stack := []*Term{term}
	seen := make(map[*Term]bool)
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current == nil || seen[current] {
			continue
		}
		seen[current] = true
		switch current.Kind {
		case TermVar, TermLambda, TermExists, TermQuantifier, TermParam, TermRef, TermHole:
			if current.ID != "" {
				ids[current.ID] = struct{}{}
			}
		case TermAsk:
			for _, answer := range current.Answers {
				ids[answer.ID] = struct{}{}
			}
		}
		stack = append(stack, rewriteChildren(current)...)
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

func canonicalizeBinders(root *Term, limits Limits) (*Term, map[string]string, error) {
	copy, err := copyTermDetached(root, limits)
	if err != nil {
		return nil, nil, err
	}
	binders := binderIDsInPreorder(copy)
	mapping := make(map[string]string, len(binders))
	for i, id := range binders {
		mapping[id] = "norm_b" + strconv.Itoa(i+1)
	}
	if err := remapBinderIDs(copy, mapping, limits); err != nil {
		return nil, nil, err
	}
	return copy, mapping, nil
}

func remapBinderIDs(root *Term, mapping map[string]string, limits Limits) error {
	type item struct {
		term  *Term
		depth int
	}
	stack := []item{{root, 1}}
	seen := make(map[*Term]bool)
	nodes := 0
	for len(stack) > 0 {
		it := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if it.term == nil {
			continue
		}
		if it.depth > limits.MaxDepth {
			return resourceLimit("expression", "binder remap depth exceeds configured limit")
		}
		if seen[it.term] {
			return invalid("CYCLIC_TERM", "expression", "binder remap encountered shared or cyclic term nodes")
		}
		seen[it.term] = true
		nodes++
		if nodes > limits.MaxNodes {
			return resourceLimit("expression", "binder remap exceeds configured node limit")
		}
		if next, ok := mapping[it.term.ID]; (it.term.Kind == TermLambda || it.term.Kind == TermExists || it.term.Kind == TermQuantifier) && ok {
			it.term.ID = next
		}
		if it.term.Kind == TermAsk {
			for i := range it.term.Answers {
				if next, ok := mapping[it.term.Answers[i].ID]; ok {
					it.term.Answers[i].ID = next
				}
			}
		}
		if it.term.Kind == TermVar {
			if next, ok := mapping[it.term.ID]; ok {
				it.term.ID = next
			}
		}
		children := rewriteChildren(it.term)
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, item{term: children[i], depth: it.depth + 1})
		}
	}
	return nil
}

func binderIDsInPreorder(root *Term) []string {
	var out []string
	var walk func(*Term)
	walk = func(t *Term) {
		if t == nil {
			return
		}
		switch t.Kind {
		case TermLambda, TermExists, TermQuantifier:
			out = append(out, t.ID)
		case TermAsk:
			for _, answer := range t.Answers {
				out = append(out, answer.ID)
			}
		}
		for _, child := range rewriteChildren(t) {
			walk(child)
		}
	}
	walk(root)
	return out
}

func rewriteBinderChanges(before, after *Term, renames map[string]string, rule RewriteRule) ([]BinderChange, error) {
	beforeIDs := binderIDsInPreorder(before)
	afterIDs := binderIDsInPreorder(after)
	afterSet := make(map[string]bool, len(afterIDs))
	for _, id := range afterIDs {
		afterSet[id] = true
	}
	out := make([]BinderChange, 0, len(beforeIDs)+len(afterIDs))
	classifiedAfter := make(map[string]bool, len(afterIDs))
	if rule == RewriteAlphaV1 {
		for _, old := range beforeIDs {
			next := renames[old]
			classifiedAfter[next] = true
			kind := "renamed"
			if next == old {
				kind = "retained"
			}
			out = append(out, BinderChange{Before: old, After: next, Kind: kind})
		}
		return out, nil
	}
	for _, old := range beforeIDs {
		if next, ok := renames[old]; ok {
			kind := "renamed"
			if old == next {
				kind = "retained"
			}
			out = append(out, BinderChange{Before: old, After: next, Kind: kind})
			classifiedAfter[next] = true
			continue
		}
		if afterSet[old] {
			out = append(out, BinderChange{Before: old, After: old, Kind: "retained"})
			classifiedAfter[old] = true
		} else {
			out = append(out, BinderChange{Before: old, After: "", Kind: "removed"})
		}
	}
	for _, id := range afterIDs {
		if !classifiedAfter[id] {
			out = append(out, BinderChange{Before: "", After: id, Kind: "introduced"})
			classifiedAfter[id] = true
		}
	}
	return out, nil
}

func rewriteAlias(aliases []AliasDefinition, id string) (AliasDefinition, bool) {
	for _, alias := range aliases {
		if alias.ID == id {
			return alias, true
		}
	}
	return AliasDefinition{}, false
}

type rewriteDeps struct{ parameters, references, holes map[string]bool }

func rewriteDependencies(root *Term, limits Limits) (rewriteDeps, error) {
	deps := rewriteDeps{parameters: map[string]bool{}, references: map[string]bool{}, holes: map[string]bool{}}
	if _, err := inspectTerm(root, limits); err != nil {
		return deps, err
	}
	stack := []*Term{root}
	seen := make(map[*Term]bool)
	for len(stack) > 0 {
		t := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if t == nil || seen[t] {
			continue
		}
		seen[t] = true
		switch t.Kind {
		case TermParam:
			deps.parameters[t.ID] = true
		case TermRef:
			deps.references[t.ID] = true
		case TermHole:
			deps.holes[t.ID] = true
		}
		stack = append(stack, rewriteChildren(t)...)
	}
	return deps, nil
}
func sameStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}

func rewriteOuterScope(root *Term, path []int) (map[string]int, error) {
	scope := map[string]int{}
	current := root
	ordinal := 0
	for depth, index := range path {
		switch current.Kind {
		case TermExists, TermLambda, TermQuantifier:
			scope[current.ID] = ordinal
			ordinal++
		case TermAsk:
			for _, answer := range current.Answers {
				scope[answer.ID] = ordinal
				ordinal++
			}
		}
		children := rewriteChildren(current)
		if index < 0 || index >= len(children) {
			return nil, invalid("INVALID_REWRITE_PATH", fmt.Sprintf("path[%d]", depth), "child index is outside the term")
		}
		current = children[index]
	}
	return scope, nil
}

func rewriteOrderKey(term *Term, outer map[string]int, limits Limits) (string, error) {
	if _, err := inspectTerm(term, limits); err != nil {
		return "", err
	}
	var b strings.Builder
	local := map[string]int{}
	next := 0
	nodes := 0
	type localRestore struct {
		id      string
		old     int
		existed bool
	}
	bindLocals := func(ids []string) []localRestore {
		restores := make([]localRestore, 0, len(ids))
		for _, id := range ids {
			old, existed := local[id]
			restores = append(restores, localRestore{id: id, old: old, existed: existed})
			local[id] = next
			next++
		}
		return restores
	}
	restoreLocals := func(restores []localRestore) {
		for i := len(restores) - 1; i >= 0; i-- {
			item := restores[i]
			if item.existed {
				local[item.id] = item.old
			} else {
				delete(local, item.id)
			}
		}
	}
	var walk func(*Term, int) error
	walk = func(t *Term, depth int) error {
		if t == nil {
			b.WriteString("nil;")
			return nil
		}
		if depth > limits.MaxDepth {
			return resourceLimit("order_key", "ordering key depth exceeds limit")
		}
		nodes++
		if nodes > limits.MaxNodes {
			return resourceLimit("order_key", "ordering key exceeds node limit")
		}
		token := func(s string) { fmt.Fprintf(&b, "%d:%s;", len(s), s) }
		token(string(t.Kind))
		switch t.Kind {
		case TermVar:
			if value, ok := local[t.ID]; ok {
				token("local")
				token(strconv.Itoa(value))
			} else if value, ok := outer[t.ID]; ok {
				token("outer")
				token(strconv.Itoa(value))
			} else {
				token("free")
				token(t.ID)
			}
		case TermParam, TermRef, TermHole:
			token(t.ID)
			if t.Type != nil {
				if err := rewriteTypeKey(&b, t.Type, 1, limits); err != nil {
					return err
				}
			}
		case TermAtom:
			token(t.Symbol)
			for _, a := range t.Args {
				token(a.Role)
				if err := walk(a.Value, depth+1); err != nil {
					return err
				}
			}
		case TermAnd:
			token(strconv.Itoa(len(t.Operands)))
			for _, o := range t.Operands {
				if err := walk(o, depth+1); err != nil {
					return err
				}
			}
		case TermNot:
			return walk(t.Body, depth+1)
		case TermModal:
			token(t.Modality)
			token(t.ModalProfile)
			return walk(t.Body, depth+1)
		case TermExists, TermLambda:
			if err := rewriteTypeKey(&b, t.Type, 1, limits); err != nil {
				return err
			}
			restores := bindLocals([]string{t.ID})
			err := walk(t.Body, depth+1)
			restoreLocals(restores)
			return err
		case TermQuantifier:
			token(t.Quantifier)
			if t.Count == nil {
				token("no-count")
			} else {
				token(strconv.Itoa(*t.Count))
			}
			if err := rewriteTypeKey(&b, t.Type, 1, limits); err != nil {
				return err
			}
			restores := bindLocals([]string{t.ID})
			if t.Restrictor == nil {
				token("no-restrictor")
			} else {
				token("restrictor")
				if err := walk(t.Restrictor, depth+1); err != nil {
					restoreLocals(restores)
					return err
				}
			}
			err := walk(t.Body, depth+1)
			restoreLocals(restores)
			return err
		case TermAsk:
			token(strconv.Itoa(len(t.Answers)))
			ids := make([]string, len(t.Answers))
			for i, answer := range t.Answers {
				ids[i] = answer.ID
				if err := rewriteTypeKey(&b, &answer.Type, 1, limits); err != nil {
					return err
				}
			}
			restores := bindLocals(ids)
			if t.Restrictor == nil {
				token("no-restrictor")
			} else {
				token("restrictor")
				if err := walk(t.Restrictor, depth+1); err != nil {
					restoreLocals(restores)
					return err
				}
			}
			err := walk(t.Body, depth+1)
			restoreLocals(restores)
			return err
		case TermApply:
			if err := walk(t.Function, depth+1); err != nil {
				return err
			}
			if err := walk(t.Argument, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(term, 1); err != nil {
		return "", err
	}
	return b.String(), nil
}

func rewriteTypeKey(b *strings.Builder, t *Type, depth int, limits Limits) error {
	if t == nil {
		b.WriteString("type:nil;")
		return nil
	}
	if depth > limits.MaxTypeDepth {
		return resourceLimit("order_key.type", "type key depth exceeds limit")
	}
	fmt.Fprintf(b, "type:%d:%s;", len(t.Kind), t.Kind)
	if t.Kind == TypeNominal {
		fmt.Fprintf(b, "%d:%s;", len(t.SortID), t.SortID)
	}
	if t.Kind == TypeArrow {
		if err := rewriteTypeKey(b, t.From, depth+1, limits); err != nil {
			return err
		}
		return rewriteTypeKey(b, t.To, depth+1, limits)
	}
	return nil
}

func rewriteBinderFree(term *Term, limits Limits) (bool, error) {
	if _, err := inspectTerm(term, limits); err != nil {
		return false, err
	}
	stack := []*Term{term}
	seen := map[*Term]bool{}
	for len(stack) > 0 {
		t := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if t == nil || seen[t] {
			continue
		}
		seen[t] = true
		if t.Kind == TermLambda || t.Kind == TermExists || t.Kind == TermQuantifier || (t.Kind == TermAsk && len(t.Answers) != 0) {
			return false, nil
		}
		stack = append(stack, rewriteChildren(t)...)
	}
	return true, nil
}

func sameIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func sameBinderChanges(a, b []BinderChange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

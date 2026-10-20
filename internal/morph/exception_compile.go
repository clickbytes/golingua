package morph

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
)

type CompiledExceptions struct {
	rules   []ExceptionRule
	byID    map[ExceptionID]int
	profile *features.LanguageFeatureProfile
}

func compileExceptions(input []ExceptionRule, profile *features.LanguageFeatureProfile) (*CompiledExceptions, error) {
	if profile == nil {
		return nil, morphError(status.OutcomeUnknown, "FEATURE_PROFILE_UNAVAILABLE", "profile", "exception compilation requires a language feature profile")
	}
	if len(input) > MaxExceptionRules {
		return nil, morphLimit("exception_rules", len(input), MaxExceptionRules)
	}
	rules := make([]ExceptionRule, len(input))
	byID := make(map[ExceptionID]int, len(input))
	totalForms, totalOverrides := 0, 0
	for index, rule := range input {
		if rule.SchemaVersion != ExceptionSchemaVersion {
			return nil, morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_SCHEMA", fmt.Sprintf("exceptions[%d].schema_version", index), "unsupported exception schema version")
		}
		if err := validateExceptionReview(rule); err != nil {
			return nil, err
		}
		if rule.Priority < 0 || rule.Priority > MaxExceptionPriority {
			return nil, morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_PRIORITY", "exception.priority", "exception priority is outside its bound")
		}
		if err := validateExceptionSelector(rule, profile); err != nil {
			return nil, err
		}
		if err := validateExceptionEffect(rule.Effect); err != nil {
			return nil, err
		}
		totalForms += len(rule.Effect.Forms)
		totalOverrides += len(rule.OverrideOf)
		if totalForms > MaxExceptionFormRecords {
			return nil, morphLimit("exception_form_records", totalForms, MaxExceptionFormRecords)
		}
		if totalOverrides > MaxExceptionOverrides {
			return nil, morphLimit("exception_override_records", totalOverrides, MaxExceptionOverrides)
		}
		if _, exists := byID[rule.ID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_EXCEPTION_ID", "exceptions."+string(rule.ID), "exception ID is duplicated")
		}
		byID[rule.ID] = index
		rules[index] = cloneExceptionRule(rule)
	}
	for _, rule := range rules {
		seen := make(map[ExceptionID]struct{}, len(rule.OverrideOf))
		for _, targetID := range rule.OverrideOf {
			if targetID == rule.ID {
				return nil, morphError(status.OutcomeInvalid, "EXCEPTION_OVERRIDE_CYCLE", "exceptions."+string(rule.ID)+".override_of", "an exception cannot override itself")
			}
			if _, exists := seen[targetID]; exists {
				return nil, morphError(status.OutcomeInvalid, "DUPLICATE_EXCEPTION_OVERRIDE", "exceptions."+string(rule.ID)+".override_of", "override target is duplicated")
			}
			seen[targetID] = struct{}{}
			targetIndex, exists := byID[targetID]
			if !exists {
				return nil, morphError(status.OutcomeUnknown, "UNKNOWN_EXCEPTION_OVERRIDE", "exceptions."+string(rule.ID)+".override_of", "override target does not resolve")
			}
			target := rules[targetIndex]
			if !selectorsOverlap(rule, target) {
				return nil, morphError(status.OutcomeInvalid, "NONOVERLAPPING_EXCEPTION_OVERRIDE", "exceptions."+string(rule.ID)+".override_of", "override target cannot match the same request")
			}
			if exceptionSpecificity(rule) < exceptionSpecificity(target) {
				return nil, morphError(status.OutcomeInvalid, "LESS_SPECIFIC_EXCEPTION_OVERRIDE", "exceptions."+string(rule.ID)+".override_of", "an exception cannot override a more-specific rule")
			}
			if rule.Priority <= target.Priority {
				return nil, morphError(status.OutcomeInvalid, "EXCEPTION_OVERRIDE_PRIORITY", "exceptions."+string(rule.ID)+".priority", "explicit override must have strictly higher priority")
			}
		}
	}
	if err := validateOverrideCycles(rules, byID); err != nil {
		return nil, err
	}
	if err := validateOverlappingRules(rules); err != nil {
		return nil, err
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	byID = make(map[ExceptionID]int, len(rules))
	for index, rule := range rules {
		byID[rule.ID] = index
	}
	return &CompiledExceptions{rules: rules, byID: byID, profile: profile}, nil
}

func validateExceptionSelector(rule ExceptionRule, profile *features.LanguageFeatureProfile) error {
	selector := rule.Selector
	if selector.Language != profile.Language() {
		return morphError(status.OutcomeIncomplete, "EXCEPTION_LANGUAGE_MISMATCH", "exception.selector.language", "exception language must match its feature profile")
	}
	if len(selector.Features) > features.MaxDimensions {
		return morphLimit("exception_feature_conditions", len(selector.Features), features.MaxDimensions)
	}
	seenFeatures := make(map[features.FeatureID]struct{}, len(selector.Features))
	for index, condition := range selector.Features {
		if !validMorphToken(string(condition.Feature)) || !validMorphToken(string(condition.Value)) {
			return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_FEATURE", fmt.Sprintf("exception.selector.features[%d]", index), "feature condition IDs are malformed")
		}
		if _, exists := seenFeatures[condition.Feature]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_EXCEPTION_FEATURE", "exception.selector.features", "feature selector repeats a dimension")
		}
		seenFeatures[condition.Feature] = struct{}{}
		dimension, exists := profile.Lookup(condition.Feature)
		if !exists {
			return morphError(status.OutcomeUnsupported, "UNSUPPORTED_EXCEPTION_FEATURE", "exception.selector.features", "exception selector feature is not in the profile")
		}
		if !containsFeatureValue(dimension.Values, condition.Value) {
			return morphError(status.OutcomeUnsupported, "UNSUPPORTED_EXCEPTION_FEATURE_VALUE", "exception.selector.features", "exception selector value is not in the profile")
		}
	}
	if selector.POS != "" && !validMorphToken(selector.POS) {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_POS", "exception.selector.pos", "part-of-speech selector is malformed")
	}
	if selector.ParadigmID != "" && !validMorphToken(string(selector.ParadigmID)) {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_PARADIGM", "exception.selector.paradigm_id", "paradigm selector is malformed")
	}
	if selector.LexemeID != "" {
		lexemeID, err := ids.ParseLexemeID(selector.LexemeID)
		if err != nil || lexemeID.String() != selector.LexemeID {
			return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_LEXEME", "exception.selector.lexeme_id", "lexeme selector is malformed")
		}
	}
	if selector.CellID != "" && !validMorphToken(string(selector.CellID)) {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_CELL", "exception.selector.cell_id", "cell selector is malformed")
	}
	if err := validateScopeFields(rule.Scope, selector); err != nil {
		return err
	}
	return nil
}

func validateScopeFields(scope ExceptionScope, selector ExceptionSelector) error {
	valid := false
	switch scope {
	case ExceptionScopeLanguage:
		valid = selector.POS == "" && selector.ParadigmID == "" && selector.LexemeID == "" && selector.CellID == ""
	case ExceptionScopePOS:
		valid = selector.POS != "" && selector.ParadigmID == "" && selector.LexemeID == "" && selector.CellID == ""
	case ExceptionScopeParadigm:
		valid = selector.POS != "" && selector.ParadigmID != "" && selector.LexemeID == "" && selector.CellID == ""
	case ExceptionScopeLexeme:
		valid = selector.POS != "" && selector.ParadigmID == "" && selector.LexemeID != "" && selector.CellID == ""
	case ExceptionScopeCell:
		valid = selector.POS != "" && selector.ParadigmID != "" && selector.LexemeID == "" && selector.CellID != ""
	case ExceptionScopeLexemeCell:
		valid = selector.POS != "" && selector.ParadigmID != "" && selector.LexemeID != "" && selector.CellID != ""
	}
	if !valid {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_SCOPE", "exception.selector", "selector identity fields do not match the declared exception scope")
	}
	return nil
}

func validateExceptionEffect(effect ExceptionEffect) error {
	seenForms := make(map[string]struct{}, len(effect.Forms))
	for index, form := range effect.Forms {
		if !validExceptionForm(form) {
			return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_FORM", fmt.Sprintf("exception.effect.forms[%d]", index), "exception form is empty or outside its bound")
		}
		if _, exists := seenForms[form]; exists {
			return morphError(status.OutcomeInvalid, "DUPLICATE_EXCEPTION_FORM", fmt.Sprintf("exception.effect.forms[%d]", index), "exception effect repeats a form")
		}
		seenForms[form] = struct{}{}
	}
	if len(effect.Forms) > MaxExceptionOutputs {
		return morphLimit("exception_outputs", len(effect.Forms), MaxExceptionOutputs)
	}
	switch effect.Kind {
	case ExceptionEffectBlock:
		if len(effect.Forms) != 0 || effect.StemID != "" || effect.StemForm != "" {
			return morphError(status.OutcomeInvalid, "INVALID_BLOCK_EFFECT", "exception.effect", "block has no forms or stem override")
		}
	case ExceptionEffectReplace, ExceptionEffectSuppletion:
		if len(effect.Forms) != 1 || effect.StemID != "" || effect.StemForm != "" {
			return morphError(status.OutcomeInvalid, "INVALID_REPLACEMENT_EFFECT", "exception.effect", "replace and suppletion require exactly one explicit form")
		}
	case ExceptionEffectAddVariant:
		if len(effect.Forms) == 0 || effect.StemID != "" || effect.StemForm != "" {
			return morphError(status.OutcomeInvalid, "INVALID_ADD_VARIANT_EFFECT", "exception.effect", "add_variant requires one or more explicit forms")
		}
	case ExceptionEffectStemOverride:
		if len(effect.Forms) != 0 || !validMorphToken(string(effect.StemID)) || !validExceptionForm(effect.StemForm) {
			return morphError(status.OutcomeInvalid, "INVALID_STEM_OVERRIDE_EFFECT", "exception.effect", "stem_override requires only a valid stem ID and explicit form")
		}
	default:
		return morphError(status.OutcomeUnsupported, "UNSUPPORTED_EXCEPTION_EFFECT", "exception.effect.kind", "exception effect is not supported")
	}
	return nil
}

func validateOverrideCycles(rules []ExceptionRule, byID map[ExceptionID]int) error {
	state := make(map[ExceptionID]uint8, len(rules))
	var visit func(ExceptionID) error
	visit = func(id ExceptionID) error {
		if state[id] == 1 {
			return morphError(status.OutcomeInvalid, "EXCEPTION_OVERRIDE_CYCLE", "exceptions."+string(id)+".override_of", "exception override graph contains a cycle")
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, target := range rules[byID[id]].OverrideOf {
			if err := visit(target); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for _, rule := range rules {
		if err := visit(rule.ID); err != nil {
			return err
		}
	}
	return nil
}

func validateOverlappingRules(rules []ExceptionRule) error {
	for leftIndex := 0; leftIndex < len(rules); leftIndex++ {
		left := rules[leftIndex]
		for rightIndex := leftIndex + 1; rightIndex < len(rules); rightIndex++ {
			right := rules[rightIndex]
			if !selectorsOverlap(left, right) {
				continue
			}
			if left.Effect.Kind == ExceptionEffectAddVariant && right.Effect.Kind == ExceptionEffectAddVariant && exceptionSpecificity(left) == exceptionSpecificity(right) {
				if duplicateExceptionForms(left.Effect.Forms, right.Effect.Forms) {
					return morphError(status.OutcomeInvalid, "DUPLICATE_EXCEPTION_VARIANT", "exceptions", "overlapping add_variant rules repeat a surface form")
				}
				continue
			}
			winner, loser := moreSpecific(left, right)
			if !containsExceptionID(winner.OverrideOf, loser.ID) {
				return morphError(status.OutcomeInvalid, "CONFLICTING_EXCEPTION_RULES", "exceptions", "overlapping exception rules require an explicit override from the more-specific rule")
			}
		}
	}
	return nil
}

func moreSpecific(left, right ExceptionRule) (ExceptionRule, ExceptionRule) {
	leftSpecificity, rightSpecificity := exceptionSpecificity(left), exceptionSpecificity(right)
	if leftSpecificity > rightSpecificity || (leftSpecificity == rightSpecificity && left.Priority > right.Priority) {
		return left, right
	}
	return right, left
}

func selectorsOverlap(left, right ExceptionRule) bool {
	a, b := left.Selector, right.Selector
	if a.Language != b.Language || fieldsConflict(a.POS, b.POS) || fieldsConflict(string(a.ParadigmID), string(b.ParadigmID)) || fieldsConflict(a.LexemeID, b.LexemeID) || fieldsConflict(string(a.CellID), string(b.CellID)) {
		return false
	}
	leftIndex, rightIndex := 0, 0
	for leftIndex < len(a.Features) && rightIndex < len(b.Features) {
		leftCondition, rightCondition := a.Features[leftIndex], b.Features[rightIndex]
		switch {
		case leftCondition.Feature < rightCondition.Feature:
			leftIndex++
		case leftCondition.Feature > rightCondition.Feature:
			rightIndex++
		default:
			if leftCondition.Value != rightCondition.Value {
				return false
			}
			leftIndex++
			rightIndex++
		}
		if leftIndex == len(a.Features) || rightIndex == len(b.Features) {
			break
		}
	}
	return true
}

func fieldsConflict(left, right string) bool { return left != "" && right != "" && left != right }

func duplicateExceptionForms(left, right []string) bool {
	seen := make(map[string]struct{}, len(left))
	for _, form := range left {
		seen[form] = struct{}{}
	}
	for _, form := range right {
		if _, exists := seen[form]; exists {
			return true
		}
	}
	return false
}

func cloneExceptionRule(source ExceptionRule) ExceptionRule {
	result := source
	result.Selector.Features = append([]ExceptionFeatureMatch(nil), source.Selector.Features...)
	sort.Slice(result.Selector.Features, func(i, j int) bool { return result.Selector.Features[i].Feature < result.Selector.Features[j].Feature })
	result.Effect.Forms = append([]string(nil), source.Effect.Forms...)
	result.OverrideOf = append([]ExceptionID(nil), source.OverrideOf...)
	return result
}

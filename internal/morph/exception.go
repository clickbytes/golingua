package morph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	ExceptionSchemaVersion  = 1
	MaxExceptionRules       = 2_048
	MaxExceptionOutputs     = 64
	MaxExceptionIDBytes     = 64
	MaxExceptionPriority    = 1_000_000
	MaxExceptionTrace       = 32
	MaxExceptionFormRecords = 8_192
	MaxExceptionOverrides   = 16_384
)

type ExceptionID string
type ExceptionScope string
type ExceptionEffectKind string

const (
	ExceptionScopeLanguage      ExceptionScope      = "language"
	ExceptionScopePOS           ExceptionScope      = "pos"
	ExceptionScopeParadigm      ExceptionScope      = "paradigm"
	ExceptionScopeLexeme        ExceptionScope      = "lexeme"
	ExceptionScopeCell          ExceptionScope      = "cell"
	ExceptionScopeLexemeCell    ExceptionScope      = "lexeme_cell"
	ExceptionEffectBlock        ExceptionEffectKind = "block"
	ExceptionEffectReplace      ExceptionEffectKind = "replace"
	ExceptionEffectAddVariant   ExceptionEffectKind = "add_variant"
	ExceptionEffectStemOverride ExceptionEffectKind = "stem_override"
	ExceptionEffectSuppletion   ExceptionEffectKind = "suppletion"
	ExceptionSourceRegular                          = "regular"
	ExceptionSource                                 = "exception"
)

type ExceptionSelector struct {
	Language   features.LanguageID     `json:"language"`
	POS        string                  `json:"pos,omitempty"`
	ParadigmID ParadigmID              `json:"paradigm_id,omitempty"`
	LexemeID   string                  `json:"lexeme_id,omitempty"`
	CellID     CellID                  `json:"cell_id,omitempty"`
	Features   []ExceptionFeatureMatch `json:"features,omitempty"`
}

// ExceptionFeatureMatch keeps selectors independent from runtime
// operations while using the same language-local feature identifiers.
type ExceptionFeatureMatch struct {
	Feature features.FeatureID      `json:"feature"`
	Value   features.FeatureValueID `json:"value"`
}

type ExceptionEffect struct {
	Kind     ExceptionEffectKind `json:"kind"`
	Forms    []string            `json:"forms,omitempty"`
	StemID   procedure.StemID    `json:"stem_id,omitempty"`
	StemForm string              `json:"stem_form,omitempty"`
}

type ExceptionRule struct {
	SchemaVersion int                  `json:"schema_version"`
	ID            ExceptionID          `json:"id"`
	Scope         ExceptionScope       `json:"scope"`
	Selector      ExceptionSelector    `json:"selector"`
	Effect        ExceptionEffect      `json:"effect"`
	Priority      int                  `json:"priority"`
	OverrideOf    []ExceptionID        `json:"override_of,omitempty"`
	Provenance    provenance.SourceRef `json:"provenance"`
	ReviewedBy    string               `json:"reviewed_by"`
	ReviewedAt    string               `json:"reviewed_at"`
}

type ExceptionRequest struct {
	Language   features.LanguageID
	POS        string
	ParadigmID ParadigmID
	LexemeID   string
	CellID     CellID
	Features   features.FeatureBundle
}

type ResolvedForm struct {
	Form        string                `json:"form"`
	Source      string                `json:"source"`
	ExceptionID ExceptionID           `json:"exception_id,omitempty"`
	Provenance  *provenance.SourceRef `json:"provenance,omitempty"`
}

type ExceptionTraceStep struct {
	RuleID   ExceptionID         `json:"rule_id"`
	Effect   ExceptionEffectKind `json:"effect"`
	SourceID string              `json:"source_id"`
}

type ExceptionResolution struct {
	Blocked              bool                        `json:"blocked,omitempty"`
	RequiresRegeneration bool                        `json:"requires_regeneration,omitempty"`
	StemOverrides        map[procedure.StemID]string `json:"stem_overrides,omitempty"`
	Forms                []ResolvedForm              `json:"forms"`
	Source               string                      `json:"source"`
	Trace                []ExceptionTraceStep        `json:"trace,omitempty"`
}

// CompileExceptions returns an order-independent, immutable exception set.
func CompileExceptions(rules []ExceptionRule, profile *features.LanguageFeatureProfile) (*CompiledExceptions, error) {
	return compileExceptions(rules, profile)
}

func DecodeExceptions(data []byte) ([]ExceptionRule, error) {
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("exception document exceeds 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var rawRules []json.RawMessage
	if err := decoder.Decode(&rawRules); err != nil {
		return nil, err
	}
	if rawRules == nil {
		return nil, fmt.Errorf("exception document must be a JSON array")
	}
	if len(rawRules) > MaxExceptionRules {
		return nil, fmt.Errorf("exception rule count exceeds %d", MaxExceptionRules)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("expected exactly one exception JSON value")
	}
	rules := make([]ExceptionRule, len(rawRules))
	for index, raw := range rawRules {
		if err := validateExceptionJSON(raw); err != nil {
			return nil, fmt.Errorf("exceptions[%d]: %w", index, err)
		}
		if err := json.Unmarshal(raw, &rules[index]); err != nil {
			return nil, fmt.Errorf("exceptions[%d]: %w", index, err)
		}
	}
	return rules, nil
}

func validateExceptionJSON(data []byte) error {
	if err := exactJSONFields(data,
		[]string{"schema_version", "id", "scope", "selector", "effect", "priority", "override_of", "provenance", "reviewed_by", "reviewed_at"},
		[]string{"schema_version", "id", "scope", "selector", "effect", "priority", "provenance", "reviewed_by", "reviewed_at"},
		[]string{"override_of", "schema_version", "id", "scope", "selector", "effect", "priority", "provenance", "reviewed_by", "reviewed_at"}); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(data, &fields)
	var selector map[string]json.RawMessage
	if err := json.Unmarshal(fields["selector"], &selector); err != nil {
		return fmt.Errorf("selector must be an object")
	}
	if err := exactJSONFields(fields["selector"], []string{"language", "pos", "paradigm_id", "lexeme_id", "cell_id", "features"}, []string{"language"}, []string{"language", "pos", "paradigm_id", "lexeme_id", "cell_id", "features"}); err != nil {
		return fmt.Errorf("selector: %w", err)
	}
	if rawFeatures, exists := selector["features"]; exists {
		var featureMatches []json.RawMessage
		if err := json.Unmarshal(rawFeatures, &featureMatches); err != nil || featureMatches == nil {
			return fmt.Errorf("selector.features must be an array")
		}
		if len(featureMatches) > features.MaxDimensions {
			return fmt.Errorf("selector.features exceeds %d entries", features.MaxDimensions)
		}
		for index, condition := range featureMatches {
			if err := exactJSONFields(condition, []string{"feature", "value"}, []string{"feature", "value"}, nil); err != nil {
				return fmt.Errorf("selector.features[%d]: %w", index, err)
			}
		}
	}
	var effect map[string]json.RawMessage
	if err := json.Unmarshal(fields["effect"], &effect); err != nil {
		return fmt.Errorf("effect must be an object")
	}
	if err := exactJSONFields(fields["effect"], []string{"kind", "forms", "stem_id", "stem_form"}, []string{"kind"}, []string{"kind", "forms", "stem_id", "stem_form"}); err != nil {
		return fmt.Errorf("effect: %w", err)
	}
	var provenanceFields map[string]json.RawMessage
	if err := json.Unmarshal(fields["provenance"], &provenanceFields); err != nil {
		return fmt.Errorf("provenance must be an object")
	}
	return exactJSONFields(fields["provenance"], []string{"id", "path", "uri", "version", "license"}, []string{"id", "version", "license"}, []string{"id", "path", "uri", "version", "license"})
}

func exactJSONFields(data []byte, allowedNames, requiredNames, nonNullNames []string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	allowed := make(map[string]struct{}, len(allowedNames))
	for _, name := range allowedNames {
		allowed[name] = struct{}{}
	}
	for name := range fields {
		if _, exists := allowed[name]; !exists {
			return fmt.Errorf("unknown or mis-cased field %q", name)
		}
	}
	for _, name := range requiredNames {
		if _, exists := fields[name]; !exists {
			return fmt.Errorf("required field %q is missing", name)
		}
	}
	for _, name := range nonNullNames {
		if raw, exists := fields[name]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("field %q cannot be null", name)
		}
	}
	return nil
}

// ResolveExceptions applies the most-specific matching rule. Base forms are
// retained only for add_variant. A block never falls through to base forms.
func (compiled *CompiledExceptions) ResolveExceptions(request ExceptionRequest, baseForms []string) (ExceptionResolution, error) {
	if compiled == nil || compiled.profile == nil {
		return ExceptionResolution{}, morphError(status.OutcomeUnknown, "EXCEPTION_SET_UNAVAILABLE", "exceptions", "compiled exception set is unavailable")
	}
	if request.Language != compiled.profile.Language() || request.Features.Language != compiled.profile.Language() || request.Features.ProfileVersion != compiled.profile.Version() {
		return ExceptionResolution{}, morphError(status.OutcomeIncomplete, "EXCEPTION_PROFILE_MISMATCH", "request", "exception request and compiled profile differ")
	}
	if err := features.ValidateBundle(compiled.profile, request.Features); err != nil {
		return ExceptionResolution{}, wrapError(err)
	}
	if !validMorphToken(request.POS) || !validMorphToken(string(request.ParadigmID)) || !validMorphToken(string(request.CellID)) {
		return ExceptionResolution{}, morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_TARGET", "request", "exception target requires bounded POS, paradigm and cell IDs")
	}
	if request.LexemeID != "" {
		lexemeID, err := ids.ParseLexemeID(request.LexemeID)
		if err != nil || lexemeID.String() != request.LexemeID {
			return ExceptionResolution{}, morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_TARGET", "request.lexeme_id", "exception target lexeme ID is invalid")
		}
	}
	if len(baseForms) > MaxExceptionOutputs {
		return ExceptionResolution{}, morphLimit("base_forms", len(baseForms), MaxExceptionOutputs)
	}
	for index, form := range baseForms {
		if !validPrincipalForm(form, MaxPrincipalPartRunes) {
			return ExceptionResolution{}, morphError(status.OutcomeInvalid, "INVALID_BASE_FORM", fmt.Sprintf("base_forms[%d]", index), "base form is empty or outside its bound")
		}
	}
	featureValues := make(map[features.FeatureID]features.FeatureValue, len(request.Features.Values))
	for _, value := range request.Features.Values {
		featureValues[value.Feature] = value
	}
	matching := make([]ExceptionRule, 0)
	bestSpecificity := -1
	for _, rule := range compiled.rules {
		matched, _, _ := matchesException(rule, request, featureValues)
		specificity := exceptionSpecificity(rule)
		if matched {
			if specificity > bestSpecificity {
				bestSpecificity = specificity
				matching = matching[:0]
			}
			if specificity == bestSpecificity {
				matching = append(matching, rule)
			}
		}
	}
	possibleMoreSpecific := false
	unresolvedStatus := status.OutcomeUnknown
	for _, rule := range compiled.rules {
		matched, maybe, state := matchesException(rule, request, featureValues)
		if !matched && maybe && exceptionSpecificity(rule) >= bestSpecificity {
			possibleMoreSpecific = true
			if state == status.OutcomeIncomplete {
				unresolvedStatus = status.OutcomeIncomplete
			}
		}
	}
	if possibleMoreSpecific {
		return ExceptionResolution{}, morphError(unresolvedStatus, "EXCEPTION_SELECTOR_UNRESOLVED", "request.features", "an unresolved feature may select a more-specific exception")
	}
	if len(matching) == 0 {
		if possibleMoreSpecific {
			return ExceptionResolution{}, morphError(unresolvedStatus, "EXCEPTION_SELECTOR_UNRESOLVED", "request.features", "an unresolved feature may select an exception")
		}
		if len(baseForms) == 0 {
			return ExceptionResolution{}, morphError(status.OutcomeUnknown, "UNKNOWN_MORPHOLOGICAL_FORM", "base_forms", "no regular form or matching exception was supplied")
		}
		return regularResolution(baseForms), nil
	}
	active, err := activeMatchingRules(matching)
	if err != nil {
		return ExceptionResolution{}, err
	}
	return applyExceptionRules(active, baseForms)
}

func (compiled *CompiledExceptions) Lookup(id ExceptionID) (ExceptionRule, bool) {
	if compiled == nil {
		return ExceptionRule{}, false
	}
	index, exists := compiled.byID[id]
	if !exists {
		return ExceptionRule{}, false
	}
	return cloneExceptionRule(compiled.rules[index]), true
}

func regularResolution(forms []string) ExceptionResolution {
	result := ExceptionResolution{Forms: make([]ResolvedForm, len(forms)), Source: ExceptionSourceRegular}
	for index, form := range forms {
		result.Forms[index] = ResolvedForm{Form: form, Source: ExceptionSourceRegular}
	}
	return result
}

func applyExceptionRules(rules []ExceptionRule, base []string) (ExceptionResolution, error) {
	result := ExceptionResolution{Source: ExceptionSource, Forms: []ResolvedForm{}, Trace: make([]ExceptionTraceStep, 0, len(rules))}
	for _, rule := range rules {
		result.Trace = append(result.Trace, ExceptionTraceStep{RuleID: rule.ID, Effect: rule.Effect.Kind, SourceID: rule.Provenance.ID})
	}
	if len(result.Trace) > MaxExceptionTrace {
		return ExceptionResolution{}, morphLimit("exception_trace", len(result.Trace), MaxExceptionTrace)
	}
	if len(rules) == 1 {
		rule := rules[0]
		switch rule.Effect.Kind {
		case ExceptionEffectBlock:
			result.Blocked = true
			return result, nil
		case ExceptionEffectReplace, ExceptionEffectSuppletion:
			result.Forms = []ResolvedForm{exceptionForm(rule, rule.Effect.Forms[0])}
			return result, nil
		case ExceptionEffectAddVariant:
			result.Forms = append(result.Forms, regularForms(base)...)
			for _, form := range rule.Effect.Forms {
				result.Forms = append(result.Forms, exceptionForm(rule, form))
			}
			return dedupeResolutionForms(result), nil
		case ExceptionEffectStemOverride:
			result.RequiresRegeneration = true
			result.StemOverrides = map[procedure.StemID]string{rule.Effect.StemID: rule.Effect.StemForm}
			return result, nil
		}
	}
	if allAddVariants(rules) {
		result.Forms = append(result.Forms, regularForms(base)...)
		for _, rule := range rules {
			for _, form := range rule.Effect.Forms {
				result.Forms = append(result.Forms, exceptionForm(rule, form))
			}
		}
		return dedupeResolutionForms(result), nil
	}
	return ExceptionResolution{}, morphError(status.OutcomeInvalid, "AMBIGUOUS_EXCEPTION_RESOLUTION", "exceptions", "multiple equally specific exceptions have no unique effect")
}

func activeMatchingRules(rules []ExceptionRule) ([]ExceptionRule, error) {
	active := make([]ExceptionRule, 0, len(rules))
	for _, candidate := range rules {
		overridden := false
		for _, other := range rules {
			if candidate.ID == other.ID {
				continue
			}
			if containsExceptionID(other.OverrideOf, candidate.ID) {
				overridden = true
				break
			}
		}
		if !overridden {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return nil, morphError(status.OutcomeInvalid, "EXCEPTION_OVERRIDE_CYCLE", "exceptions", "matching exception override chain has no winner")
	}
	if len(active) > 1 && !allAddVariants(active) {
		return nil, morphError(status.OutcomeInvalid, "AMBIGUOUS_EXCEPTION_RESOLUTION", "exceptions", "multiple equally specific exceptions remain active")
	}
	sort.Slice(active, func(i, j int) bool { return active[i].ID < active[j].ID })
	return active, nil
}

func exceptionForm(rule ExceptionRule, form string) ResolvedForm {
	provenanceCopy := rule.Provenance
	return ResolvedForm{Form: form, Source: ExceptionSource, ExceptionID: rule.ID, Provenance: &provenanceCopy}
}

func regularForms(forms []string) []ResolvedForm {
	result := make([]ResolvedForm, len(forms))
	for index, form := range forms {
		result[index] = ResolvedForm{Form: form, Source: ExceptionSourceRegular}
	}
	return result
}

func dedupeResolutionForms(result ExceptionResolution) ExceptionResolution {
	seen := make(map[string]struct{}, len(result.Forms))
	forms := result.Forms[:0]
	for _, form := range result.Forms {
		key := form.Source + "\x00" + string(form.ExceptionID) + "\x00" + form.Form
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		forms = append(forms, form)
	}
	result.Forms = forms
	return result
}

func allAddVariants(rules []ExceptionRule) bool {
	if len(rules) == 0 {
		return false
	}
	for _, rule := range rules {
		if rule.Effect.Kind != ExceptionEffectAddVariant {
			return false
		}
	}
	return true
}

func matchesException(rule ExceptionRule, request ExceptionRequest, values map[features.FeatureID]features.FeatureValue) (bool, bool, status.OutcomeKind) {
	selector := rule.Selector
	if selector.Language != request.Language || (selector.POS != "" && selector.POS != request.POS) || (selector.ParadigmID != "" && selector.ParadigmID != request.ParadigmID) || (selector.CellID != "" && selector.CellID != request.CellID) {
		return false, false, ""
	}
	identityUnresolved := selector.LexemeID != "" && request.LexemeID == ""
	if selector.LexemeID != "" && request.LexemeID != "" && selector.LexemeID != request.LexemeID {
		return false, false, ""
	}
	unresolved, unresolvedStatus := identityUnresolved, status.OutcomeUnknown
	for _, condition := range selector.Features {
		value, exists := values[condition.Feature]
		if !exists || value.State == features.StateAbsent {
			unresolved = true
			continue
		}
		if value.State == features.StateUnknown {
			unresolved = true
			unresolvedStatus = status.OutcomeIncomplete
			continue
		}
		if value.Value != condition.Value {
			return false, false, ""
		}
	}
	if unresolved {
		return false, true, unresolvedStatus
	}
	return true, false, ""
}

func exceptionSpecificity(rule ExceptionRule) int {
	rank := map[ExceptionScope]int{ExceptionScopeLanguage: 1, ExceptionScopePOS: 2, ExceptionScopeParadigm: 3, ExceptionScopeLexeme: 4, ExceptionScopeCell: 5, ExceptionScopeLexemeCell: 6}[rule.Scope]
	return rank*100 + len(rule.Selector.Features)
}

func containsExceptionID(values []ExceptionID, candidate ExceptionID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func validExceptionForm(form string) bool {
	return validPrincipalForm(form, MaxPrincipalPartRunes) && utf8.RuneCountInString(form) <= MaxPrincipalPartRunes
}

func validateExceptionReview(rule ExceptionRule) error {
	if !validMorphToken(string(rule.ID)) || len(rule.ID) > MaxExceptionIDBytes {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_ID", "exception.id", "exception ID is malformed")
	}
	if err := rule.Provenance.Validate(); err != nil {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_PROVENANCE", "exception.provenance", err.Error())
	}
	if len(rule.ReviewedBy) > 64 || !validMorphToken(rule.ReviewedBy) {
		return morphError(status.OutcomeInvalid, "MISSING_EXCEPTION_REVIEWER", "exception.reviewed_by", "exception requires a named reviewer identifier")
	}
	if len(rule.ReviewedAt) > 64 {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_REVIEW_TIME", "exception.reviewed_at", "review time exceeds its bound")
	}
	if _, err := time.Parse(time.RFC3339, rule.ReviewedAt); err != nil {
		return morphError(status.OutcomeInvalid, "INVALID_EXCEPTION_REVIEW_TIME", "exception.reviewed_at", "review time must use RFC3339")
	}
	return nil
}

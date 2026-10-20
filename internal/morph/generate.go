package morph

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

type GenerationState string

const (
	GenerationGenerated   GenerationState = "generated"
	GenerationBlocked     GenerationState = "blocked"
	GenerationIncomplete  GenerationState = "incomplete"
	GenerationUnavailable GenerationState = "unavailable"
	GenerationUnknown     GenerationState = "unknown"
)

const (
	MaxGeneratedForms  = 64
	MaxGenerationTrace = 160
)

type GenerationTraceStep struct {
	Operation   procedure.OperationKind `json:"operation,omitempty"`
	Applied     *bool                   `json:"applied,omitempty"`
	OutputSize  int                     `json:"output_size,omitempty"`
	ProcedureID string                  `json:"procedure_id,omitempty"`
	ParadigmID  ParadigmID              `json:"paradigm_id,omitempty"`
	CellID      CellID                  `json:"cell_id,omitempty"`
	Exception   ExceptionID             `json:"exception_id,omitempty"`
	Effect      ExceptionEffectKind     `json:"effect,omitempty"`
	SourceID    string                  `json:"source_id,omitempty"`
}

type GeneratedForm struct {
	Surface        string                  `json:"surface"`
	LexemeID       string                  `json:"lexeme_id"`
	Language       features.LanguageID     `json:"language"`
	ProfileVersion features.ProfileVersion `json:"profile_version"`
	POS            string                  `json:"pos"`
	ParadigmID     ParadigmID              `json:"paradigm_id"`
	CellID         CellID                  `json:"cell_id"`
	Features       []features.FeatureValue `json:"features"`
	Source         string                  `json:"source"`
	ExceptionID    ExceptionID             `json:"exception_id,omitempty"`
	Provenance     *provenance.SourceRef   `json:"provenance,omitempty"`
	Trace          []GenerationTraceStep   `json:"trace"`
}

type GenerationResult struct {
	State       GenerationState       `json:"state"`
	Forms       []GeneratedForm       `json:"forms"`
	Trace       []GenerationTraceStep `json:"trace"`
	Diagnostics []status.Diagnostic   `json:"diagnostics,omitempty"`
}

// Generate resolves one language-specific lexeme and feature request through
// a compiled paradigm and optional exceptions. Unknown evidence is returned as
// an explicit state; invalid, unsupported, incomplete and resource-limit
// failures remain typed errors.
func Generate(lexeme lexicon.MorphLexeme, parts PrincipalParts, bundle features.FeatureBundle, paradigms *CompiledParadigms, exceptions *CompiledExceptions, budget procedure.Budget) (GenerationResult, error) {
	if err := budget.Validate(); err != nil {
		return GenerationResult{}, morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_BUDGET", "budget", err.Error())
	}
	lexemeID, err := ids.ParseLexemeID(lexeme.ID.String())
	if err != nil || lexemeID.String() != lexeme.ID.String() {
		return GenerationResult{}, morphError(status.OutcomeInvalid, "INVALID_LEXEME_ID", "lexeme.id", "generation requires a canonical typed lexeme ID")
	}
	if lexeme.Language == "" || !validPrincipalForm(lexeme.Lemma, MaxPrincipalPartRunes) || !validMorphToken(string(lexeme.POS)) {
		return GenerationResult{}, morphError(status.OutcomeInvalid, "INVALID_MORPH_LEXEME", "lexeme", "lexeme language, lemma and POS must be valid and bounded")
	}
	if len(lexeme.Properties) > features.MaxBundleValues {
		return GenerationResult{}, morphLimit("lexeme_properties", len(lexeme.Properties), features.MaxBundleValues)
	}
	seenProperties := make(map[string]struct{}, len(lexeme.Properties))
	for _, property := range lexeme.Properties {
		if !validMorphToken(property.Key) || !validMorphToken(property.Value) {
			return GenerationResult{}, morphError(status.OutcomeInvalid, "INVALID_MORPH_LEXEME_PROPERTY", "lexeme.properties", "lexeme property keys and values must be bounded tokens")
		}
		if _, exists := seenProperties[property.Key]; exists {
			return GenerationResult{}, morphError(status.OutcomeInvalid, "DUPLICATE_MORPH_LEXEME_PROPERTY", "lexeme.properties", "lexeme properties must have unique keys")
		}
		seenProperties[property.Key] = struct{}{}
	}
	if paradigms == nil || paradigms.profile == nil {
		return GenerationResult{}, morphError(status.OutcomeUnknown, "PARADIGM_SET_UNAVAILABLE", "paradigms", "compiled paradigm set is unavailable")
	}
	if lexeme.Language != lexicon.LanguageID(paradigms.profile.Language()) || bundle.Language != paradigms.profile.Language() || bundle.ProfileVersion != paradigms.profile.Version() {
		return GenerationResult{}, morphError(status.OutcomeIncomplete, "MORPHOLOGY_PROFILE_MISMATCH", "request", "lexeme, feature bundle and paradigm profile differ")
	}
	bundle, err = features.NewBundle(paradigms.profile, bundle.Values)
	if err != nil {
		return GenerationResult{}, wrapError(err)
	}
	if parts.LexemeID != lexeme.ID {
		return GenerationResult{}, morphError(status.OutcomeInvalid, "PRINCIPAL_PART_LEXEME_MISMATCH", "principal_parts.lexeme_id", "principal parts belong to a different lexeme")
	}
	if lexeme.ParadigmID == "" {
		return generationWithDiagnostic(GenerationUnknown, "PARADIGM_UNAVAILABLE", "lexeme has no assigned paradigm", "lexeme.paradigm_id"), nil
	}
	paradigm, exists := paradigms.Lookup(ParadigmID(lexeme.ParadigmID))
	if !exists {
		return unknownGeneration(morphError(status.OutcomeUnknown, "UNKNOWN_PARADIGM", "lexeme.paradigm_id", "lexeme paradigm is not in the compiled set")), nil
	}
	if string(paradigm.POS) != string(lexeme.POS) {
		return GenerationResult{}, morphError(status.OutcomeUnsupported, "PARADIGM_POS_MISMATCH", "lexeme.pos", "lexeme POS differs from the assigned paradigm")
	}
	if err := features.ValidateFeatureRequest(paradigms.profile, bundle); err != nil {
		return GenerationResult{}, wrapError(err)
	}
	cell, err := paradigms.SelectCell(paradigm.ID, bundle)
	if err != nil {
		if isOutcome(err, status.OutcomeUnknown) {
			return unknownGeneration(err), nil
		}
		return GenerationResult{}, err
	}
	request := ExceptionRequest{Language: bundle.Language, POS: string(lexeme.POS), ParadigmID: paradigm.ID, LexemeID: lexeme.ID.String(), CellID: cell.ID, Features: bundle}
	preResolved := ExceptionResolution{}
	preMatched := false
	if exceptions != nil {
		preResolved, err = exceptions.ResolveExceptions(request, nil)
		if err == nil {
			preMatched = true
			if preResolved.Blocked {
				return GenerationResult{State: GenerationBlocked, Forms: []GeneratedForm{}, Trace: exceptionGenerationTrace(preResolved.Trace, paradigm.ID, cell.ID)}, nil
			}
			if !preResolved.RequiresRegeneration && !allTraceEffects(preResolved.Trace, ExceptionEffectAddVariant) {
				if err := normalizeResolutionForms(&preResolved); err != nil {
					return GenerationResult{}, err
				}
				return generatedResult(lexeme, paradigm, cell, bundle, preResolved, nil, nil), nil
			}
		} else if !isNoMatchingException(err) {
			if isOutcome(err, status.OutcomeUnknown) {
				return unknownGeneration(err), nil
			}
			return GenerationResult{}, err
		}
	}
	var rule *CellRule
	for index := range paradigm.Rules {
		if paradigm.Rules[index].CellID == cell.ID {
			matchedRule := cloneRule(paradigm.Rules[index])
			rule = &matchedRule
			break
		}
	}
	if rule == nil {
		path := "paradigms." + string(paradigm.ID) + ".cells." + string(cell.ID) + ".rule"
		return generationWithDiagnostic(GenerationIncomplete, "CELL_RULE_UNAVAILABLE", "selected paradigm cell has no generation rule", path), nil
	}
	if len(paradigm.PrincipalPartRequirements) > MaxPrincipalParts {
		return GenerationResult{}, morphLimit("principal_part_requirements", len(paradigm.PrincipalPartRequirements), MaxPrincipalParts)
	}
	var overrides map[procedure.StemID]string
	if preMatched && preResolved.RequiresRegeneration {
		overrides = preResolved.StemOverrides
	}
	stems, err := resolveGenerationStems(parts, paradigm.PrincipalPartRequirements, overrides, budget)
	if err != nil {
		if isOutcome(err, status.OutcomeUnknown) {
			return unknownGeneration(err), nil
		}
		return GenerationResult{}, err
	}
	for stemID, form := range overrides {
		declared := false
		for _, requirement := range paradigm.PrincipalPartRequirements {
			if requirement.StemID == stemID {
				declared = true
				break
			}
		}
		if !declared {
			return GenerationResult{}, morphError(status.OutcomeInvalid, "UNDECLARED_STEM_OVERRIDE", "exception.effect.stem_id", "stem override does not target a declared paradigm stem")
		}
		stems[stemID] = form
	}
	if len(stems) > budget.MaxStems {
		return GenerationResult{}, morphLimit("stems", len(stems), budget.MaxStems)
	}
	execution, err := procedure.ExecPlan(stems, rule.Steps, bundle, budget)
	if err != nil {
		if isOutcome(err, status.OutcomeUnknown) {
			return unknownGeneration(err), nil
		}
		return GenerationResult{}, err
	}
	normalized, err := NormalizeSurface(execution.Form, DefaultNormalizationBudget())
	if err != nil {
		return GenerationResult{}, err
	}
	resolution := regularResolution([]string{normalized})
	if exceptions != nil {
		resolution, err = exceptions.ResolveExceptions(request, []string{normalized})
		if err != nil {
			if isOutcome(err, status.OutcomeUnknown) {
				return unknownGeneration(err), nil
			}
			return GenerationResult{}, err
		}
	}
	if err := normalizeResolutionForms(&resolution); err != nil {
		return GenerationResult{}, err
	}
	if len(resolution.Forms) > MaxGeneratedForms {
		return GenerationResult{}, morphLimit("generated_forms", len(resolution.Forms), MaxGeneratedForms)
	}
	return generatedResult(lexeme, paradigm, cell, bundle, resolution, execution.Trace, rule), nil
}

func normalizeResolutionForms(resolution *ExceptionResolution) error {
	seen := make(map[string]struct{}, len(resolution.Forms))
	forms := make([]ResolvedForm, 0, len(resolution.Forms))
	for index := range resolution.Forms {
		normalized, err := NormalizeSurface(resolution.Forms[index].Form, DefaultNormalizationBudget())
		if err != nil {
			return err
		}
		resolution.Forms[index].Form = normalized
		key := resolution.Forms[index].Source + "\x00" + string(resolution.Forms[index].ExceptionID) + "\x00" + normalized
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		forms = append(forms, resolution.Forms[index])
	}
	resolution.Forms = forms
	return nil
}

func resolveGenerationStems(parts PrincipalParts, requirements []PrincipalPartRequirement, overrides map[procedure.StemID]string, budget procedure.Budget) (map[procedure.StemID]string, error) {
	if len(requirements) > MaxPrincipalParts {
		return nil, morphLimit("principal_part_requirements", len(requirements), MaxPrincipalParts)
	}
	stems := make(map[procedure.StemID]string, len(requirements))
	seenStems := make(map[procedure.StemID]struct{}, len(requirements))
	seenParts := make(map[PrincipalPartID]struct{}, len(requirements))
	for _, requirement := range requirements {
		if !validStemID(requirement.StemID) || !validPrincipalPartID(requirement.PartID) {
			return nil, morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_REQUIREMENT", "principal_part_requirements", "stem and principal-part IDs must be bounded tokens")
		}
		if _, exists := seenStems[requirement.StemID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_STEM_REQUIREMENT", "principal_part_requirements", "stem requirement is duplicated")
		}
		if _, exists := seenParts[requirement.PartID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_PRINCIPAL_PART_REQUIREMENT", "principal_part_requirements", "principal part is mapped more than once")
		}
		seenStems[requirement.StemID] = struct{}{}
		seenParts[requirement.PartID] = struct{}{}
		form, err := ResolveStem(parts, requirement, budget)
		if err == nil {
			stems[requirement.StemID] = form
			continue
		}
		if isOutcome(err, status.OutcomeUnknown) {
			if !requirement.Required {
				continue
			}
			if _, replaced := overrides[requirement.StemID]; replaced {
				continue
			}
		}
		return nil, err
	}
	return stems, nil
}

func generatedResult(lexeme lexicon.MorphLexeme, paradigm Paradigm, cell ParadigmCell, bundle features.FeatureBundle, resolution ExceptionResolution, operationTrace []procedure.StepTrace, rule *CellRule) GenerationResult {
	forms := make([]GeneratedForm, len(resolution.Forms))
	procedureID, sourceID := "", ""
	if rule != nil {
		procedureID, sourceID = rule.ProcedureID, rule.Provenance.ID
	}
	trace := generationTrace(operationTrace, resolution.Trace, paradigm.ID, cell.ID, procedureID, sourceID)
	for index, form := range resolution.Forms {
		formTrace := cloneGenerationTrace(trace)
		var provenanceCopy *provenance.SourceRef
		if form.Provenance != nil {
			copy := *form.Provenance
			provenanceCopy = &copy
		} else if form.Source == ExceptionSourceRegular && rule != nil && rule.Provenance.ID != "" {
			copy := rule.Provenance
			provenanceCopy = &copy
		}
		forms[index] = GeneratedForm{Surface: form.Form, LexemeID: lexeme.ID.String(), Language: features.LanguageID(lexeme.Language), ProfileVersion: bundle.ProfileVersion, POS: string(lexeme.POS), ParadigmID: paradigm.ID, CellID: cell.ID, Features: cloneFeatureValues(bundle.Values), Source: form.Source, ExceptionID: form.ExceptionID, Provenance: provenanceCopy, Trace: formTrace}
	}
	state := GenerationGenerated
	if resolution.Blocked {
		state = GenerationBlocked
	}
	return GenerationResult{State: state, Forms: forms, Trace: trace}
}

func unknownGeneration(err error) GenerationResult {
	result := GenerationResult{State: GenerationUnknown, Forms: []GeneratedForm{}}
	if failure, ok := err.(interface{ Result() status.Result }); ok {
		result.Diagnostics = append([]status.Diagnostic(nil), failure.Result().Diagnostics...)
	}
	return result
}

func isOutcome(err error, kind status.OutcomeKind) bool {
	failure, ok := err.(interface{ Result() status.Result })
	return ok && failure.Result().Status == kind
}

func isNoMatchingException(err error) bool {
	failure, ok := err.(interface{ Result() status.Result })
	if !ok {
		return false
	}
	result := failure.Result()
	return result.Status == status.OutcomeUnknown && len(result.Diagnostics) > 0 && result.Diagnostics[0].Code == "UNKNOWN_MORPHOLOGICAL_FORM"
}

func allTraceEffects(trace []ExceptionTraceStep, effect ExceptionEffectKind) bool {
	if len(trace) == 0 {
		return false
	}
	for _, step := range trace {
		if step.Effect != effect {
			return false
		}
	}
	return true
}

func exceptionGenerationTrace(trace []ExceptionTraceStep, paradigmID ParadigmID, cellID CellID) []GenerationTraceStep {
	return generationTrace(nil, trace, paradigmID, cellID, "", "")
}

func generationTrace(operations []procedure.StepTrace, exceptions []ExceptionTraceStep, paradigmID ParadigmID, cellID CellID, procedureID, sourceID string) []GenerationTraceStep {
	result := make([]GenerationTraceStep, 0, len(operations)+len(exceptions))
	for _, step := range operations {
		applied := step.Applied
		result = append(result, GenerationTraceStep{Operation: step.Operation, Applied: &applied, OutputSize: step.OutputSize, ProcedureID: procedureID, ParadigmID: paradigmID, CellID: cellID, SourceID: sourceID})
	}
	for _, step := range exceptions {
		result = append(result, GenerationTraceStep{Exception: step.RuleID, Effect: step.Effect, ParadigmID: paradigmID, CellID: cellID, SourceID: step.SourceID})
	}
	return result
}

func generationWithDiagnostic(state GenerationState, code, message, path string) GenerationResult {
	diagnostic := status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning}
	return GenerationResult{State: state, Forms: []GeneratedForm{}, Diagnostics: []status.Diagnostic{diagnostic}}
}

// BoundedTrace converts the detailed morphology trace to the shared bounded
// status trace, marking truncation when the morphology trace is longer.
func (result GenerationResult) BoundedTrace() status.Trace {
	count := min(len(result.Trace), status.MaxTraceSteps)
	trace := status.Trace{Steps: make([]status.TraceStep, 0, count), Truncated: len(result.Trace) > count}
	for _, step := range result.Trace[:count] {
		trace.Steps = append(trace.Steps, status.TraceStep{
			RuleID: string(step.Exception), ProcedureID: step.ProcedureID,
			ParadigmID: string(step.ParadigmID), CellID: string(step.CellID),
			SourceID: step.SourceID,
		})
	}
	return trace
}

func (form GeneratedForm) Reading() Reading {
	var source *provenance.SourceRef
	if form.Provenance != nil {
		copy := *form.Provenance
		source = &copy
	}
	return Reading{Surface: form.Surface, LexemeID: form.LexemeID, POS: form.POS, ParadigmID: form.ParadigmID, CellID: form.CellID, Features: features.FeatureBundle{Language: form.Language, ProfileVersion: form.ProfileVersion, Values: cloneFeatureValues(form.Features)}, Source: form.Source, ExceptionID: form.ExceptionID, Provenance: source, Trace: cloneGenerationTrace(form.Trace)}
}

func cloneGenerationTrace(source []GenerationTraceStep) []GenerationTraceStep {
	result := make([]GenerationTraceStep, len(source))
	copy(result, source)
	for index, step := range source {
		if step.Applied != nil {
			applied := *step.Applied
			result[index].Applied = &applied
		}
	}
	return result
}

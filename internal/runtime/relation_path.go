package runtime

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type RelationPathFeaturePlanner interface {
	Features(context.Context, engine.RelationPathRequest, engine.RelationPathVariant, engine.RelationPathPart, lexicon.MorphLexeme) ([]features.FeatureValue, error)
}

type RelationPathMorphology interface {
	Lexemes() []lexicon.MorphLexeme
	Generate(string, []features.FeatureValue) (morph.GenerationResult, error)
}

type RelationPathRealizer struct {
	Language   features.LanguageID
	Plan       engine.RelationPathPlan
	Morphology RelationPathMorphology
	Features   RelationPathFeaturePlanner
}

var _ engine.RelationPathTargetRealizer = (*RelationPathRealizer)(nil)
var _ engine.RelationPathTargetVerifier = (*RelationPathRealizer)(nil)

func (r *RelationPathRealizer) RelationPathPlan() (engine.RelationPathPlan, error) {
	if r == nil || r.Language == "" {
		return engine.RelationPathPlan{}, relationPathError(status.OutcomeUnsupported, "RELATION_PATH_TARGET_UNAVAILABLE", "target.relation_path", "relation path plan is not configured")
	}
	encoded, err := json.Marshal(r.Plan)
	if err != nil || len(encoded) > engine.MaxRelationPlanBytes {
		return engine.RelationPathPlan{}, relationPathError(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "target.relation_path.plan", "configured plan exceeds its encoding bound")
	}
	plan, err := engine.DecodeRelationPathPlan(encoded, string(r.Language))
	if err != nil {
		return engine.RelationPathPlan{}, relationPathError(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "target.relation_path.plan", "configured plan failed strict validation: "+err.Error())
	}
	return cloneRelationPathPlan(plan), nil
}

func (r *RelationPathRealizer) RealizeRelationPath(ctx context.Context, request engine.RelationPathRequest) (engine.RelationPathResult, error) {
	if ctx == nil {
		return relationPathFailure(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "relation path realization requires a context"), nil
	}
	if err := ctx.Err(); err != nil {
		return relationPathFailure(status.OutcomeResourceLimit, "RELATION_PATH_CANCELED", "context", "relation path realization was canceled"), nil
	}
	if r == nil || r.Morphology == nil || r.Features == nil || r.Language == "" || request.Package == nil || string(request.Package.Language()) != string(r.Language) {
		return relationPathFailure(status.OutcomeUnsupported, "RELATION_PATH_TARGET_UNAVAILABLE", "target.relation_path", "selected package, language plan, morphology and feature planner are required"), nil
	}
	budget, err := semantic.NormalizeDSEBudget(request.Budget)
	if err != nil {
		return relationPathFailure(status.OutcomeInvalid, "INVALID_DSE_BUDGET", "budget", "relation path realization requires a valid finite DSE budget"), nil
	}
	plan, err := r.RelationPathPlan()
	if err != nil {
		return relationPathFailure(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "target.relation_path.plan", err.Error()), nil
	}
	seed, ok, seedWork := relationPathSeed(r.Morphology.Lexemes(), request.Selection.Lexeme, budget.MaxWork)
	if seedWork > budget.MaxWork {
		return relationPathFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation path morphology seed scan exceeded the work limit"), nil
	}
	if !ok {
		return relationPathFailure(status.OutcomeUnsupported, "RELATION_PATH_SEED_UNAVAILABLE", "target.relation_path.lexeme", "the exact selected verb lexeme is absent from morphology seed"), nil
	}
	var result engine.RelationPathResult
	result.Status = status.OutcomeUnknown
	compatible, work := 0, seedWork
	failed, ambiguous := false, false
	for _, variant := range plan.Variants {
		work++
		if work > budget.MaxWork {
			return relationPathFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation path plan scan exceeded the work limit"), nil
		}
		if err := ctx.Err(); err != nil {
			return relationPathFailure(status.OutcomeResourceLimit, "RELATION_PATH_CANCELED", "context", "relation path realization was canceled"), nil
		}
		if !relationVariantRoleCompatible(variant, request) {
			continue
		}
		compatible++
		alts, outcome, diagnostic, partWork := r.realizeRelationVariant(ctx, request, plan, variant, seed, budget.MaxWork-work)
		work += partWork
		if outcome != status.OutcomeOK {
			if outcome == status.OutcomeInvalid {
				return relationPathFailure(outcome, diagnostic.Code, diagnostic.Path, diagnostic.Message), nil
			}
			failed = true
			if outcome == status.OutcomeAmbiguous {
				ambiguous = true
			}
			result.Diagnostics = appendRelationPathDiagnostic(result.Diagnostics, diagnostic)
			if outcome == status.OutcomeResourceLimit {
				return relationPathFailure(outcome, diagnostic.Code, diagnostic.Path, diagnostic.Message), nil
			}
			continue
		}
		if len(result.Alternatives)+len(alts) > engine.MaxRelationTargetAlternatives {
			return relationPathFailure(status.OutcomeResourceLimit, "RELATION_PATH_ALTERNATIVE_LIMIT", "target.alternatives", "relation path alternatives exceed the configured finite limit"), nil
		}
		appendCandidates := make([]engine.RelationPathAlternative, 0, len(alts))
		for _, candidate := range alts {
			duplicate := false
			for _, prior := range result.Alternatives {
				if candidate.ID == prior.ID {
					if reflect.DeepEqual(candidate, prior) {
						duplicate = true
						break
					}
					return relationPathFailure(status.OutcomeInvalid, "RELATION_PATH_ID_COLLISION", "target.alternatives", "stable alternative ID collision with different evidence"), nil
				}
			}
			if !duplicate {
				appendCandidates = append(appendCandidates, candidate)
			}
		}
		outputBytes := relationPathOutputBytes(append(append([]engine.RelationPathAlternative(nil), result.Alternatives...), appendCandidates...))
		if outputBytes > budget.MaxOutputBytes {
			return relationPathFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.output_bytes", "relation path output exceeds the finite byte limit"), nil
		}
		result.Alternatives = append(result.Alternatives, appendCandidates...)
	}
	if compatible == 0 {
		return relationPathFailure(status.OutcomeUnsupported, "RELATION_PATH_ORIENTATION_UNAVAILABLE", "target.relation_path.plan", "no plan variant maps both ordered arguments to the selected frame subject and object roles"), nil
	}
	if len(result.Alternatives) == 0 {
		if !failed {
			return relationPathFailure(status.OutcomeIncomplete, "RELATION_PATH_UNCOVERED", "target.relation_path", "no complete surface alternative was produced"), nil
		}
		if ambiguous {
			result.Status = status.OutcomeAmbiguous
		} else {
			result.Status = status.OutcomeIncomplete
		}
		return result, nil
	}
	if failed {
		result.Status = status.OutcomeIncomplete
	} else if len(result.Alternatives) > 1 {
		result.Status = status.OutcomeAmbiguous
		result.Diagnostics = appendRelationPathDiagnostic(result.Diagnostics, status.Diagnostic{Code: "RELATION_PATH_AMBIGUOUS", Path: "target.alternatives", Message: "multiple fully covered relation path alternatives remain licensed", Severity: status.SeverityWarning})
	} else {
		result.Status = status.OutcomeOK
	}
	encodedResult, err := json.Marshal(result)
	if err != nil || len(encodedResult) > budget.MaxOutputBytes {
		return relationPathFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.output_bytes", "relation path result exceeds the finite serialized output limit"), nil
	}
	if err := engine.ValidateRelationPathResult(request, plan, result); err != nil {
		return relationPathFailure(status.OutcomeInvalid, "RELATION_PATH_RESULT_INVALID", "target.relation_path.result", err.Error()), nil
	}
	return result, nil
}

func (r *RelationPathRealizer) VerifyRelationPath(ctx context.Context, request engine.RelationPathRequest, alternative engine.RelationPathAlternative) error {
	if ctx == nil {
		return relationPathError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "relation path verification requires a context")
	}
	if err := ctx.Err(); err != nil {
		return relationPathError(status.OutcomeResourceLimit, "RELATION_PATH_CANCELED", "context", "relation path verification was canceled")
	}
	plan, err := r.RelationPathPlan()
	if err != nil {
		return err
	}
	if err := engine.ValidateRelationPathAlternative(request, plan, alternative); err != nil {
		return relationPathError(status.OutcomeUnsupported, "RELATION_PATH_ALTERNATIVE_INVALID", "target.relation_path.alternative", err.Error())
	}
	result, err := r.RealizeRelationPath(ctx, request)
	if err != nil {
		return err
	}
	if result.Status != status.OutcomeOK && result.Status != status.OutcomeAmbiguous && result.Status != status.OutcomeIncomplete {
		return relationPathError(result.Status, "RELATION_PATH_REGENERATION_INCOMPLETE", "target.relation_path", "fresh realization did not establish a complete verifiable result")
	}
	count := 0
	for _, candidate := range result.Alternatives {
		if reflect.DeepEqual(candidate, alternative) {
			count++
		}
	}
	if count != 1 {
		return relationPathError(status.OutcomeUnsupported, "RELATION_PATH_ALTERNATIVE_UNVERIFIED", "target.relation_path.alternative", "the exact alternative must occur once in fresh regeneration")
	}
	if err := ctx.Err(); err != nil {
		return relationPathError(status.OutcomeResourceLimit, "RELATION_PATH_CANCELED", "context", "relation path verification was canceled")
	}
	return nil
}

func (r *RelationPathRealizer) realizeRelationVariant(ctx context.Context, request engine.RelationPathRequest, plan engine.RelationPathPlan, variant engine.RelationPathVariant, seed lexicon.MorphLexeme, maxWork int) ([]engine.RelationPathAlternative, status.OutcomeKind, status.Diagnostic, int) {
	base := []engine.RelationPathAlternative{{VariantID: variant.ID}}
	work := 0
	for _, part := range variant.Parts {
		work++
		if work > maxWork {
			return nil, status.OutcomeResourceLimit, relationPathDiagnostic("RESOURCE_LIMIT", "budget.work", "relation path construction work limit exceeded"), work
		}
		if err := ctx.Err(); err != nil {
			return nil, status.OutcomeResourceLimit, relationPathDiagnostic("RELATION_PATH_CANCELED", "context", "relation path realization was canceled"), work
		}
		if part.Slot != "verb" {
			for i := range base {
				sp := engine.RelationPathSurfacePart{Step: part.Step, Text: part.Literal}
				if part.Slot != "" {
					sp.BinderID = relationPathBinder(request, part.Slot)
					sp.Port = relationPathPort(request, part.Step, part.Slot)
					sp.GrammaticalFunction = part.GrammaticalFunction
				}
				base[i].Parts = append(base[i].Parts, sp)
			}
			continue
		}
		featuresPlanned, err := r.Features.Features(ctx, request, variant, part, seed)
		if err != nil {
			return nil, status.OutcomeUnsupported, relationPathDiagnostic("RELATION_PATH_FEATURES_UNAVAILABLE", "target.features", err.Error()), work
		}
		generated, err := r.Morphology.Generate(seed.ID.String(), featuresPlanned)
		if err != nil {
			return nil, status.OutcomeUnsupported, relationPathDiagnostic("RELATION_PATH_MORPHOLOGY_UNAVAILABLE", "target.morphology", err.Error()), work
		}
		if generated.State != morph.GenerationGenerated || len(generated.Forms) == 0 {
			return nil, status.OutcomeIncomplete, relationPathDiagnostic("RELATION_PATH_VERB_FORM_UNAVAILABLE", "target.morphology", "selected verb has no generated form for this plan slot"), work
		}
		if len(base)*len(generated.Forms) > engine.MaxRelationTargetAlternatives {
			return nil, status.OutcomeResourceLimit, relationPathDiagnostic("RELATION_PATH_ALTERNATIVE_LIMIT", "target.alternatives", "relation path exceeds the finite alternative limit"), work
		}
		next := make([]engine.RelationPathAlternative, 0, len(base)*len(generated.Forms))
		for _, alt := range base {
			for _, form := range generated.Forms {
				work++
				if work > maxWork {
					return nil, status.OutcomeResourceLimit, relationPathDiagnostic("RESOURCE_LIMIT", "budget.work", "relation path morphology expansion exceeded the work limit"), work
				}
				if form.LexemeID != seed.ID.String() || string(form.Language) != string(r.Language) || form.Surface == "" {
					return nil, status.OutcomeInvalid, relationPathDiagnostic("RELATION_PATH_FORM_IDENTITY_MISMATCH", "target.morphology", "generated form does not belong to the exact selected verb lexeme and language"), work
				}
				copyAlt := alt
				copyAlt.Parts = append(append([]engine.RelationPathSurfacePart(nil), alt.Parts...), engine.RelationPathSurfacePart{Step: part.Step, Text: form.Surface, Generated: cloneRelationPathForm(form)})
				next = append(next, copyAlt)
			}
		}
		base = next
	}
	for i := range base {
		base[i].ID = engine.RelationPathAlternativeID(base[i].VariantID, base[i].Parts)
		if base[i].ID == "" {
			return nil, status.OutcomeInvalid, relationPathDiagnostic("RELATION_PATH_ID_UNAVAILABLE", "target.alternatives", "alternative evidence could not be canonically encoded"), work
		}
		if err := engine.ValidateRelationPathAlternative(request, plan, base[i]); err != nil {
			return nil, status.OutcomeInvalid, relationPathDiagnostic("RELATION_PATH_ALTERNATIVE_INVALID", "target.relation_path", err.Error()), work
		}
	}
	deduplicated := make([]engine.RelationPathAlternative, 0, len(base))
	seen := map[string]engine.RelationPathAlternative{}
	for _, alt := range base {
		if prior, exists := seen[alt.ID]; exists {
			if !reflect.DeepEqual(prior, alt) {
				return nil, status.OutcomeInvalid, relationPathDiagnostic("RELATION_PATH_ID_COLLISION", "target.alternatives", "stable alternative ID collision with different evidence"), work
			}
			continue
		}
		seen[alt.ID] = alt
		deduplicated = append(deduplicated, alt)
	}
	return deduplicated, status.OutcomeOK, status.Diagnostic{}, work
}

func relationVariantRoleCompatible(variant engine.RelationPathVariant, request engine.RelationPathRequest) bool {
	for _, part := range variant.Parts {
		if part.Slot != "x" && part.Slot != "y" && part.Slot != "z" {
			continue
		}
		if relationPathPosition(part.Step, part.Slot) != part.Port {
			return false
		}
		port := relationPathPort(request, part.Step, part.Slot)
		role, ok := request.Selection.PortRoles[port]
		if !ok {
			return false
		}
		if part.GrammaticalFunction == "subject" && role != request.Selection.SubjectRole || part.GrammaticalFunction == "object" && role != request.Selection.ObjectRole {
			return false
		}
	}
	return true
}

func relationPathPosition(step, slot string) string {
	if step == "first_atom" && slot == "x" || step == "second_atom" && slot == "z" {
		return "first"
	}
	if step == "first_atom" && slot == "z" || step == "second_atom" && slot == "y" {
		return "second"
	}
	return ""
}

func relationPathPort(request engine.RelationPathRequest, step, slot string) string {
	if relationPathPosition(step, slot) == "first" {
		return request.OrderedPorts[0]
	}
	if relationPathPosition(step, slot) == "second" {
		return request.OrderedPorts[1]
	}
	return ""
}

func relationPathBinder(request engine.RelationPathRequest, slot string) string {
	switch slot {
	case "x":
		return request.BinderX
	case "y":
		return request.BinderY
	case "z":
		return request.BinderZ
	default:
		return ""
	}
}

func relationPathSeed(seeds []lexicon.MorphLexeme, selected lexicon.MorphLexeme, maxWork int) (lexicon.MorphLexeme, bool, int) {
	var found lexicon.MorphLexeme
	count := 0
	work := 0
	for _, seed := range seeds {
		work++
		if work > maxWork {
			return lexicon.MorphLexeme{}, false, work
		}
		if seed.ID == selected.ID && seed.Language == selected.Language && seed.Lemma == selected.Lemma && seed.POS == selected.POS {
			found, count = seed, count+1
		}
	}
	return found, count == 1, work
}

func relationPathOutputBytes(alternatives []engine.RelationPathAlternative) int {
	total := 0
	for _, alternative := range alternatives {
		for _, part := range alternative.Parts {
			total += len(part.Text)
		}
	}
	return total
}

func cloneRelationPathForm(form morph.GeneratedForm) *morph.GeneratedForm {
	encoded, err := json.Marshal(form)
	if err == nil {
		var copy morph.GeneratedForm
		if json.Unmarshal(encoded, &copy) == nil {
			return &copy
		}
	}
	copy := form
	copy.Features = append(copy.Features[:0:0], form.Features...)
	copy.Trace = append(copy.Trace[:0:0], form.Trace...)
	return &copy
}

func cloneRelationPathPlan(plan engine.RelationPathPlan) engine.RelationPathPlan {
	result := plan
	result.Variants = make([]engine.RelationPathVariant, len(plan.Variants))
	for i, variant := range plan.Variants {
		result.Variants[i] = variant
		result.Variants[i].Parts = append([]engine.RelationPathPart(nil), variant.Parts...)
	}
	return result
}

func relationPathDiagnostic(code, path, message string) status.Diagnostic {
	return status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
}

func appendRelationPathDiagnostic(current []status.Diagnostic, diagnostic status.Diagnostic) []status.Diagnostic {
	if len(current) >= status.MaxDiagnostics {
		return current
	}
	return append(current, diagnostic)
}
func relationPathFailure(kind status.OutcomeKind, code, path, message string) engine.RelationPathResult {
	return engine.RelationPathResult{Status: kind, Diagnostics: []status.Diagnostic{relationPathDiagnostic(code, path, message)}}
}
func relationPathError(kind status.OutcomeKind, code, path, message string) error {
	return targetPipelineError(kind, code, path, message)
}

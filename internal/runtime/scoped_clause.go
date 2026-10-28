package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// ScopedFeaturePlanner makes case and agreement policy explicit and local to
// the selected language profile.
type ScopedFeaturePlanner interface {
	Features(context.Context, engine.ScopedTargetRequest, construction.ScopedClausePart, lexicon.MorphLexeme, lexicon.MorphLexeme) ([]features.FeatureValue, error)
}

type ScopedMorphology interface {
	Lexemes() []lexicon.MorphLexeme
	Generate(string, []features.FeatureValue) (morph.GenerationResult, error)
}

// ScopedClauseRealizer realizes only the exact lexical selections supplied by
// the checked source-side scope analysis.
type ScopedClauseRealizer struct {
	Language   features.LanguageID
	Plan       construction.ScopedClausePlan
	Valency    *valency.Index
	Morphology ScopedMorphology
	Features   ScopedFeaturePlanner
}

var _ engine.ScopedTargetRealizer = (*ScopedClauseRealizer)(nil)
var _ engine.ScopedTargetVerifier = (*ScopedClauseRealizer)(nil)

// ScopedPlan returns a defensively decoded copy, so callers cannot mutate the
// plan used by this realizer and malformed in-memory plans are refused.
func (r *ScopedClauseRealizer) ScopedPlan() (construction.ScopedClausePlan, error) {
	if r == nil || r.Language == "" {
		return construction.ScopedClausePlan{}, targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_UNAVAILABLE", "target.scope", "scope realizer has no configured language plan")
	}
	data, err := json.Marshal(r.Plan)
	if err != nil {
		return construction.ScopedClausePlan{}, targetPipelineError(status.OutcomeInvalid, "SCOPED_PLAN_INVALID", "target.scope.plan", "scope plan cannot be encoded for validation")
	}
	plan, err := construction.DecodeScopedClausePlan(data, r.Language)
	if err != nil {
		return construction.ScopedClausePlan{}, targetPipelineError(status.OutcomeInvalid, "SCOPED_PLAN_INVALID", "target.scope.plan", "scope plan failed strict validation: "+err.Error())
	}
	return plan, nil
}

// VerifyScope recomputes the exact request and requires one unique occurrence
// of the complete alternative, including its generated forms and identity.
func (r *ScopedClauseRealizer) VerifyScope(ctx context.Context, req engine.ScopedTargetRequest, alternative engine.ScopedTargetAlternative) error {
	if ctx == nil {
		return targetPipelineError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "scope verification requires a context")
	}
	if err := ctx.Err(); err != nil {
		return targetPipelineError(status.OutcomeResourceLimit, "SCOPED_TARGET_CANCELED", "context", "scope verification was canceled")
	}
	result, err := r.RealizeScope(ctx, req)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return targetPipelineError(status.OutcomeResourceLimit, "SCOPED_TARGET_CANCELED", "context", "scope verification was canceled")
	}
	if result.Status != status.OutcomeOK && result.Status != status.OutcomeAmbiguous {
		return targetPipelineError(result.Status, "SCOPED_TARGET_REFUSED", "target.scope", "scope realization did not establish a verifiable result")
	}
	matches := 0
	for _, candidate := range result.Alternatives {
		if reflect.DeepEqual(candidate, alternative) {
			matches++
		}
	}
	if matches != 1 {
		return targetPipelineError(status.OutcomeUnsupported, "SCOPED_ALTERNATIVE_UNVERIFIED", "target.scope.alternative", "the exact alternative must occur exactly once in a fresh realization")
	}
	return nil
}

func (r *ScopedClauseRealizer) RealizeScope(ctx context.Context, req engine.ScopedTargetRequest) (engine.ScopedTargetResult, error) {
	if ctx == nil {
		return scopedFailure(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "scope realization requires a context"), nil
	}
	if err := ctx.Err(); err != nil {
		return scopedFailure(status.OutcomeResourceLimit, "SCOPED_TARGET_CANCELED", "context", "scope realization was canceled"), nil
	}
	if r == nil || r.Valency == nil || r.Morphology == nil || r.Features == nil || r.Language == "" || r.Plan.Language != r.Language || r.Plan.SchemaVersion != construction.ScopedClauseSchemaVersion {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_TARGET_UNAVAILABLE", "target.scope", "scope construction, valency, morphology, and language feature planning must be configured"), nil
	}
	if req.CheckedExpression == nil || req.BinderX == "" || req.BinderY == "" || req.BinderEvent == "" || req.BinderX == req.BinderY || req.NegationLocation != engine.ScopedNegationOuter && req.NegationLocation != engine.ScopedNegationInner {
		return scopedFailure(status.OutcomeInvalid, "INVALID_SCOPED_TARGET_REQUEST", "request", "checked expression, distinct binders, and recognized negation placement are required"), nil
	}
	selections := []engine.ScopedTargetLexicalSelection{req.XSelection, req.YSelection, req.EventSelection}
	for _, s := range selections {
		if s.Binding.SenseID != s.Sense.ID || s.Sense.LexemeID != s.Lexeme.ID || s.Validation.SenseID != s.Sense.ID.String() || !s.Validation.Authorized || s.Validation.Status != s.Binding.Status || s.Validation.Mapping != s.Binding.Mapping || !s.Binding.TargetExecutionAuthorized() || s.Lexeme.Language != lexicon.LanguageID(r.Language) || s.Lexeme.ID.String() == "" || s.Lexeme.Lemma == "" || s.Lexeme.POS == "" {
			return scopedFailure(status.OutcomeUnsupported, "SCOPED_SELECTION_UNAUTHORIZED", "request.selections", "each exact lexical selection must be authorized and internally consistent"), nil
		}
	}
	if !verifyScopedShape(req) {
		return scopedFailure(status.OutcomeInvalid, "SCOPED_SHAPE_MISMATCH", "request.paths", "the exact Not(All(...)) or All(Not(...)) tree, binders, ports, event frame, and selected symbols must match the supplied paths"), nil
	}
	if req.XSelection.Role != req.Roles.XRole || req.XSelection.Port != req.Roles.XPort || req.YSelection.Role != req.Roles.YRole || req.YSelection.Port != req.Roles.YPort || req.EventSelection.Port != req.Roles.EventPort || req.Roles.XRole == "" || req.Roles.YRole == "" || req.Roles.XRole == req.Roles.YRole || req.Roles.XPort == req.Roles.YPort || req.Roles.EventPort == "" {
		return scopedFailure(status.OutcomeInvalid, "SCOPED_ROLE_IDENTITY_MISMATCH", "request.roles", "selected binder roles and ports must match the explicit request roles"), nil
	}
	if req.EventSelection.Binding.Target == nil {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_EVENT_TARGET_UNAVAILABLE", "request.event", "selected event has no explicit target port mapping"), nil
	}
	frames := r.Valency.FramesBySenseID(req.EventSelection.Sense.ID)
	var frame *valency.ValencyFrame
	for i := range frames {
		hasFrame := false
		for _, p := range req.EventSelection.Binding.Target.RolePorts {
			if p.FrameID == string(frames[i].ID) {
				hasFrame = true
			}
		}
		if hasFrame {
			if frame != nil {
				return scopedFailure(status.OutcomeAmbiguous, "SCOPED_FRAME_AMBIGUOUS", "request.event", "selected event maps to multiple valency frames"), nil
			}
			frame = &frames[i]
		}
	}
	if frame == nil {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_FRAME_UNAVAILABLE", "request.event", "the selected event sense has no exact licensed valency frame"), nil
	}
	if !scopedFrameRole(*frame, req.Roles.XRole, req.Roles.XPort, req.EventSelection.Binding) || !scopedFrameRole(*frame, req.Roles.YRole, req.Roles.YPort, req.EventSelection.Binding) {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_FRAME_ROLE_MISMATCH", "request.roles", "frame roles and target ports do not license both selected binders"), nil
	}
	subject := ""
	subjects, objects := 0, 0
	objectRole := ""
	for _, a := range frame.Arguments {
		if a.GrammaticalFunction == "subject" {
			subjects++
			if string(a.Role) == req.Roles.XRole {
				subject = "x"
			} else if string(a.Role) == req.Roles.YRole {
				subject = "y"
			}
		}
		if a.GrammaticalFunction == "object" {
			objects++
			objectRole = string(a.Role)
			if string(a.Role) != req.Roles.XRole && string(a.Role) != req.Roles.YRole {
				return scopedFailure(status.OutcomeUnsupported, "SCOPED_OBJECT_ROLE_MISMATCH", "request.roles", "the frame object must be one of the two selected binder roles"), nil
			}
		}
	}
	if subject == "" || subjects != 1 || objects != 1 || (subject == "x" && objectRole != req.Roles.YRole) || (subject == "y" && objectRole != req.Roles.XRole) {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_SUBJECT_UNAVAILABLE", "request.roles", "the selected frame must have exactly one subject and one object matching distinct selected roles"), nil
	}
	var variant *construction.ScopedClauseVariant
	for i := range r.Plan.Variants {
		v := &r.Plan.Variants[i]
		if v.Negation == string(req.NegationLocation) && v.SubjectBinder == subject {
			if variant != nil {
				return scopedFailure(status.OutcomeAmbiguous, "SCOPED_VARIANT_AMBIGUOUS", "plan.variants", "multiple variants match this scope and subject"), nil
			}
			variant = v
		}
	}
	if variant == nil {
		return scopedFailure(status.OutcomeUnsupported, "SCOPED_VARIANT_UNAVAILABLE", "plan.variants", "no compiled construction preserves this scope and subject binding"), nil
	}
	if req.EventTemporalFrame == "" {
		return scopedFailure(status.OutcomeIncomplete, "SCOPED_TEMPORAL_FRAME_UNKNOWN", "request.event_temporal_frame", "the event temporal frame is required by the compiled construction"), nil
	}
	lexemes := make(map[string]lexicon.MorphLexeme)
	for _, item := range r.Morphology.Lexemes() {
		lexemes[item.ID.String()] = item
	}
	for _, s := range selections {
		seed, ok := lexemes[s.Lexeme.ID.String()]
		if !ok || seed.Language != s.Lexeme.Language || seed.Lemma != s.Lexeme.Lemma || seed.POS != s.Lexeme.POS {
			return scopedFailure(status.OutcomeUnsupported, "SCOPED_SEED_LEXEME_MISMATCH", "request.selections", "selected lexeme does not exactly match the morphology seed"), nil
		}
	}
	if result := enforceScopedConstraints(*variant, req, lexemes); result != nil {
		return *result, nil
	}
	constructionID, parseErr := ids.ParseConstructionID(variant.ID)
	if parseErr != nil || constructionID.String() != variant.ID {
		return scopedFailure(status.OutcomeInvalid, "SCOPED_CONSTRUCTION_ID_INVALID", "plan.variant.id", "selected construction ID is not canonical"), nil
	}
	base := []engine.ScopedTargetAlternative{{ConstructionID: constructionID, Parts: []engine.ScopedSurfacePart{}}}
	for _, part := range variant.Parts {
		if err := ctx.Err(); err != nil {
			return scopedFailure(status.OutcomeResourceLimit, "SCOPED_TARGET_CANCELED", "context", "scope realization was canceled"), nil
		}
		if part.Literal != "" {
			for i := range base {
				base[i].Parts = append(base[i].Parts, engine.ScopedSurfacePart{Text: part.Literal})
			}
			continue
		}
		selection, binder := scopedPartSelection(req, part.Slot)
		seed := lexemes[selection.Lexeme.ID.String()]
		subjectSelection := req.XSelection
		if subject == "y" {
			subjectSelection = req.YSelection
		}
		values, err := r.Features.Features(ctx, req, part, seed, lexemes[subjectSelection.Lexeme.ID.String()])
		if err != nil {
			return scopedFailure(status.OutcomeUnsupported, "SCOPED_FEATURE_PLAN_UNAVAILABLE", "target.features", err.Error()), nil
		}
		generated, err := r.Morphology.Generate(selection.Lexeme.ID.String(), values)
		if err != nil {
			return scopedFailure(status.OutcomeUnsupported, "SCOPED_MORPHOLOGY_UNAVAILABLE", "target.morphology", err.Error()), nil
		}
		if generated.State != morph.GenerationGenerated || len(generated.Forms) == 0 {
			return scopedFailure(status.OutcomeIncomplete, "SCOPED_FORM_UNAVAILABLE", "target.morphology", "the selected lexeme has no complete generated form for this construction slot"), nil
		}
		if len(base)*len(generated.Forms) > engine.MaxScopedTargetAlternatives {
			return scopedFailure(status.OutcomeResourceLimit, "SCOPED_ALTERNATIVE_LIMIT", "target.alternatives", "scope realization exceeds the complete-alternative limit"), nil
		}
		next := make([]engine.ScopedTargetAlternative, 0, len(base)*len(generated.Forms))
		for _, alt := range base {
			for _, form := range generated.Forms {
				if form.LexemeID != selection.Lexeme.ID.String() || form.Language != r.Language || form.Surface == "" {
					return scopedFailure(status.OutcomeInvalid, "SCOPED_FORM_IDENTITY_MISMATCH", "target.morphology", "generated form does not belong to the exact selected lexeme and language"), nil
				}
				copyAlt := alt
				copyAlt.Parts = append(append([]engine.ScopedSurfacePart(nil), alt.Parts...), engine.ScopedSurfacePart{Text: form.Surface, BinderID: binder, Port: selection.Port, Generated: cloneScopedForm(form)})
				next = append(next, copyAlt)
			}
		}
		base = next
	}
	for i := range base {
		base[i].ID = scopedAlternativeID(base[i], req)
	}
	return engine.ScopedTargetResult{Status: status.OutcomeOK, Alternatives: base}, nil
}

func scopedFailure(kind status.OutcomeKind, code, path, message string) engine.ScopedTargetResult {
	return engine.ScopedTargetResult{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}
}
func verifyScopedShape(req engine.ScopedTargetRequest) bool {
	root := req.CheckedExpression
	path := "input.expression"
	outerNeg := req.NegationLocation == engine.ScopedNegationOuter
	if outerNeg {
		if root.Kind != algebra.TermNot || root.Body == nil || req.Paths.Negation != path {
			return false
		}
		root, path = root.Body, path+".body"
	}
	if root.Kind != algebra.TermQuantifier || root.Quantifier != "all" || root.ID != req.BinderX || root.Restrictor == nil || root.Body == nil || root.Type == nil || root.Type.Kind != algebra.TypeEntity || req.Paths.OuterQuantifier != path || req.Paths.Restrictor != path+".restrictor" {
		return false
	}
	if !scopedUnaryAtom(root.Restrictor, req.BinderX, req.XSelection.Binding) {
		return false
	}
	body, bodyPath := root.Body, path+".body"
	innerNeg := !outerNeg
	if innerNeg {
		if body.Kind != algebra.TermNot || body.Body == nil || req.Paths.Negation != bodyPath {
			return false
		}
		body, bodyPath = body.Body, bodyPath+".body"
	}
	if body.Kind != algebra.TermExists || body.ID != req.BinderY || body.ID == root.ID || body.Type == nil || body.Type.Kind != algebra.TypeEntity || body.Body == nil || req.Paths.InnerExistential != bodyPath {
		return false
	}
	if body.Body.Kind != algebra.TermAnd || len(body.Body.Operands) != 2 {
		return false
	}
	var predicate, event *algebra.Term
	var predicatePath, eventPath string
	for i, term := range body.Body.Operands {
		if term == nil {
			return false
		}
		p := fmt.Sprintf("%s.body.operands[%d]", bodyPath, i)
		if term.Kind == algebra.TermExists {
			if event != nil {
				return false
			}
			event, eventPath = term, p
		} else {
			if predicate != nil || term.Kind != algebra.TermAtom {
				return false
			}
			predicate, predicatePath = term, p
		}
	}
	if predicate == nil || event == nil || predicate.Symbol != string(req.YSelection.Binding.Target.SymbolRef) || event.ID != req.BinderEvent || event.ID == root.ID || event.ID == body.ID || event.Type == nil || event.Type.Kind != algebra.TypeEvent || event.Body == nil || event.TemporalFrame != req.EventTemporalFrame || req.Paths.InnerPredicate != predicatePath || req.Paths.EventBinder != eventPath || req.Paths.EventAtom != eventPath+".body" {
		return false
	}
	if !scopedUnaryAtom(predicate, req.BinderY, req.YSelection.Binding) || event.Body.Kind != algebra.TermAtom || event.Body.Symbol != string(req.EventSelection.Binding.Target.SymbolRef) || len(event.Body.Args) != 3 {
		return false
	}
	seen := map[string]int{}
	for _, arg := range event.Body.Args {
		if arg.Value == nil || arg.Value.Kind != algebra.TermVar {
			return false
		}
		port := ""
		switch arg.Value.ID {
		case req.BinderX:
			port = req.Roles.XPort
		case req.BinderY:
			port = req.Roles.YPort
		case req.BinderEvent:
			port = req.Roles.EventPort
		}
		if port == "" || arg.Role != port {
			return false
		}
		seen[arg.Value.ID]++
	}
	return seen[req.BinderX] == 1 && seen[req.BinderY] == 1 && seen[req.BinderEvent] == 1
}
func scopedUnaryAtom(term *algebra.Term, binder string, binding lexicon.SemanticBinding) bool {
	return binding.Target != nil && binding.Target.ReferencePort != nil && term != nil && term.Kind == algebra.TermAtom && term.Symbol == string(binding.Target.SymbolRef) && len(term.Args) == 1 && term.Args[0].Role == binding.Target.ReferencePort.PortID && term.Args[0].Value != nil && term.Args[0].Value.Kind == algebra.TermVar && term.Args[0].Value.ID == binder
}
func scopedFrameRole(frame valency.ValencyFrame, role, port string, binding lexicon.SemanticBinding) bool {
	if binding.Target == nil {
		return false
	}
	declared := false
	for _, a := range frame.Arguments {
		if string(a.Role) == role && a.GrammaticalFunction != "" {
			declared = true
		}
	}
	mapped := false
	for _, p := range binding.Target.RolePorts {
		if p.FrameID == string(frame.ID) && p.Role == role && p.PortID == port {
			mapped = true
		}
	}
	return declared && mapped
}
func scopedPartSelection(req engine.ScopedTargetRequest, slot string) (engine.ScopedTargetLexicalSelection, string) {
	switch slot {
	case "x":
		return req.XSelection, req.BinderX
	case "y":
		return req.YSelection, req.BinderY
	default:
		return req.EventSelection, req.BinderEvent
	}
}
func enforceScopedConstraints(v construction.ScopedClauseVariant, req engine.ScopedTargetRequest, seeds map[string]lexicon.MorphLexeme) *engine.ScopedTargetResult {
	for _, c := range v.Constraints {
		var actual string
		if c.Binder == "event" && c.Feature == "temporal_frame" {
			actual = req.EventTemporalFrame
		} else {
			s, _ := scopedPartSelection(req, c.Binder)
			for _, p := range seeds[s.Lexeme.ID.String()].Properties {
				if p.Key == c.Feature {
					actual = p.Value
				}
			}
		}
		if actual == "" {
			x := scopedFailure(status.OutcomeUnsupported, "SCOPED_CONSTRAINT_UNKNOWN", "plan.constraints", fmt.Sprintf("constraint %s on %s is unknown", c.Feature, c.Binder))
			return &x
		}
		found := false
		for _, allowed := range c.Values {
			if allowed == actual {
				found = true
			}
		}
		if !found {
			x := scopedFailure(status.OutcomeUnsupported, "SCOPED_CONSTRAINT_MISMATCH", "plan.constraints", "selected lexical evidence does not satisfy a construction constraint")
			return &x
		}
	}
	return nil
}
func scopedAlternativeID(a engine.ScopedTargetAlternative, req engine.ScopedTargetRequest) string {
	var b strings.Builder
	b.WriteString(a.ConstructionID.String())
	for _, selection := range []engine.ScopedTargetLexicalSelection{req.XSelection, req.YSelection, req.EventSelection} {
		b.WriteByte(0)
		b.WriteString(selection.Sense.ID.String())
		b.WriteByte(0)
		b.WriteString(selection.Lexeme.ID.String())
		b.WriteByte(0)
		b.WriteString(selection.Binding.Provenance.ID)
	}
	for _, p := range a.Parts {
		b.WriteByte(0)
		b.WriteString(p.BinderID)
		b.WriteByte(0)
		b.WriteString(p.Port)
		b.WriteByte(0)
		b.WriteString(p.Text)
		if p.Generated != nil {
			b.WriteByte(0)
			encoded, _ := json.Marshal(p.Generated)
			b.Write(encoded)
		}
	}
	sum := sha256.Sum256([]byte(b.String()))
	return "scoped-" + hex.EncodeToString(sum[:12])
}
func cloneScopedForm(f morph.GeneratedForm) *morph.GeneratedForm {
	f.Features = append([]features.FeatureValue(nil), f.Features...)
	f.Trace = append([]morph.GenerationTraceStep(nil), f.Trace...)
	return &f
}

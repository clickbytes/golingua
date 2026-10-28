package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

type relationSourceSelection struct {
	Status                 status.OutcomeKind
	Diagnostic             status.Diagnostic
	Alternative            RelationSourceAlternative
	Variant                engine.RelationPathVariant
	Mapping                language.SourcePredicateMapping
	Binding                lexicon.SemanticBinding
	Sense                  lexicon.LexicalSense
	Lexeme                 lexicon.MorphLexeme
	Frame                  valency.ValencyFrame
	Validation             language.TargetValidation
	Signature              algebra.Signature
	HeadPort               algebra.Port
	RolePorts              map[string]algebra.Port
	PortRoles              map[string]string
	SubjectRole            string
	ObjectRole             string
	VerbReadings           [2]RelationSourcePartEvidence
	RelationPlanProvenance provenance.SourceRef
	SourceProvenance       provenance.SourceRef
	BindingProvenance      provenance.SourceRef
	FrameProvenance        provenance.SourceRef
	Work                   int
}

// selectRelationSourceMetadata selects one explicitly authorized source-plan
// mapping for a complete relation-path alternative. It emits metadata evidence
// only and creates no algebra terms or references.
func selectRelationSourceMetadata(ctx context.Context, alternative RelationSourceAlternative, relationPlan engine.RelationPathPlan, options DSEOptions) relationSourceSelection {
	out := relationSourceSelection{Status: status.OutcomeInvalid, Alternative: cloneRelationSourceAlternative(alternative), RolePorts: map[string]algebra.Port{}, PortRoles: map[string]string{}}
	fail := func(kind status.OutcomeKind, code, path, message string) relationSourceSelection {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	if ctx == nil {
		return fail(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "relation source metadata selection requires a context")
	}
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_BUDGET", "budget", "relation source metadata selection requires a valid finite DSE budget")
	}
	if options.Package == nil || options.Package.LexiconIndex() == nil || options.Package.ValencyIndex() == nil || options.Plan == nil {
		return fail(status.OutcomeInvalid, "MISSING_SOURCE_METADATA", "source", "selected semantic package and source plan are required")
	}
	if err := ctx.Err(); err != nil {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
	}
	strictPlan, err := revalidateRelationPathPlan(relationPlan, string(options.Package.Language()))
	if err != nil {
		return fail(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "relation_plan", err.Error())
	}
	variant, variantCount := relationPlanVariant(strictPlan, alternative.VariantID, &out.Work, budget.MaxWork)
	if out.Work > budget.MaxWork {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation source plan scan exceeded its finite work limit")
	}
	if variantCount != 1 || variant == nil {
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_VARIANT_INVALID", "alternative.variant_id", "source alternative must name exactly one variant in the strict compiled plan")
	}
	out.Variant = *variant
	if err := validateRelationSourceEvidence(alternative, *variant, &out.Work, budget.MaxWork, ctx); err != nil {
		if errors.Is(err, errRelationBindingResource) {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation source evidence scan exceeded its finite work limit")
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
		}
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_EVIDENCE_INVALID", "alternative", err.Error())
	}
	first, second, ok := relationSourceVerbEvidence(alternative)
	if !ok {
		return fail(status.OutcomeIncomplete, "RELATION_SOURCE_VERBS_MISSING", "alternative.parts", "both independent verb readings require explicit source evidence")
	}
	out.VerbReadings = [2]RelationSourcePartEvidence{cloneRelationSourcePart(first), cloneRelationSourcePart(second)}
	if !sameRelationSourceIdentity(first, second) {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_SENSE_MISMATCH", "alternative.parts", "the two relation verb readings must identify the same exact lexeme and sense")
	}
	senseID, parseErr := ids.ParseSenseID(first.Sense.ID.String())
	if parseErr != nil || senseID.String() != first.Sense.ID.String() {
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_SENSE_INVALID", "alternative.parts", "verb reading sense ID is not canonical")
	}
	selectedSense, exists := options.Package.LexiconIndex().BySenseID(senseID)
	if !exists || selectedSense.ID != first.Sense.ID || selectedSense.LexemeID != first.Sense.LexemeID {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_SENSE_UNAVAILABLE", "alternative.parts", "source verb sense is absent from the selected semantic package")
	}
	selectedLexeme, exists := options.Package.LexiconIndex().ByLexemeID(selectedSense.LexemeID)
	if !exists || first.Lexeme == nil || !sameRelationSourceLexemeIdentity(*first.Lexeme, selectedLexeme) || selectedLexeme.Language != options.Package.Language() || selectedLexeme.POS != "verb" {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_LEXEME_UNAVAILABLE", "alternative.parts", "source verb lexeme identity does not match the selected language package")
	}
	if first.Reading == nil || second.Reading == nil || first.Reading.LexemeID != selectedLexeme.ID.String() || second.Reading.LexemeID != selectedLexeme.ID.String() || first.Reading.POS != "verb" || second.Reading.POS != "verb" {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_READING_UNAVAILABLE", "alternative.parts", "both source readings must retain known verb morphology for the selected lexeme")
	}
	if err := ctx.Err(); err != nil {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
	}
	rows, err := options.Package.ValidateTargets(options.Selected)
	if err != nil {
		return fail(status.OutcomeInvalid, "TARGET_VALIDATION_FAILED", "selected", err.Error())
	}
	sourceDoc := options.Plan.Document()
	if sourceDoc.Language != string(options.Package.Language()) {
		return fail(status.OutcomeInvalid, "SOURCE_PLAN_LANGUAGE_MISMATCH", "source_plan.language", "source plan and selected package languages differ")
	}
	bindings := options.Package.LexiconIndex().TargetBindingsForSense(senseID)
	if len(bindings) > lexicon.MaxBindings {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.bindings", "relation source binding inventory exceeds its finite cap")
	}
	var matches []relationSourceSelection
	var sawSenseBinding bool
	for _, mapping := range sourceDoc.Predicates {
		if !relationSourceChargeBinding(&out.Work, budget.MaxWork, 1) {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source predicate mapping scan exceeded its finite work limit")
		}
		if err := ctx.Err(); err != nil {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
		}
		if mapping.SenseID != senseID {
			continue
		}
		sawSenseBinding = true
		allowed := false
		for _, allowedID := range mapping.AllowedVariantIDs {
			if !relationSourceChargeBinding(&out.Work, budget.MaxWork, 1) {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source variant authorization scan exceeded its finite work limit")
			}
			if string(allowedID) == alternative.VariantID {
				allowed = true
			}
		}
		if !allowed || mapping.FrameID == "" || mapping.HeadPortID == "" {
			continue
		}
		frame, frameOK, frameErr := relationSourceFrame(ctx, options.Package, mapping.FrameID, senseID, &out.Work, budget.MaxWork)
		if frameErr != nil {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
		}
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source frame scan exceeded its finite work limit")
		}
		if !frameOK {
			continue
		}
		subjectRole, objectRole, frameOK := relationSourceFrameRoles(frame)
		if !frameOK {
			continue
		}
		for _, candidate := range bindings {
			if !relationSourceChargeBinding(&out.Work, budget.MaxWork, 1) {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "semantic binding scan exceeded its finite work limit")
			}
			if err := ctx.Err(); err != nil {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
			}
			if candidate.Target == nil || candidate.Target.Kind != lexicon.SemanticTargetSymbol || candidate.Mapping != lexicon.MappingEquivalent || candidate.Status != lexicon.BindingApproved || !candidate.TargetExecutionAuthorized() {
				continue
			}
			validation, authorized, validationErr := relationSourceAuthorizedValidation(ctx, rows, candidate, &out.Work, budget.MaxWork)
			if validationErr != nil {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
			}
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "target validation scan exceeded its finite work limit")
			}
			if !authorized {
				continue
			}
			signature, signatureOK, signatureErr := relationSourceSelectedSignature(ctx, options.Selected, candidate.Target.SymbolRef, &out.Work, budget.MaxWork)
			if signatureErr != nil {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
			}
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "selected signature scan exceeded its finite work limit")
			}
			if !signatureOK {
				continue
			}
			head, entityPorts, signatureOK := relationSourceSignaturePorts(signature, mapping.HeadPortID)
			if !signatureOK {
				continue
			}
			rolePorts, portRoles, mapped, roleErr := relationSourceRolePorts(ctx, candidate.Target.RolePorts, frame, entityPorts, head.Role, subjectRole, objectRole, &out.Work, budget.MaxWork)
			if roleErr != nil {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
			}
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "semantic role-port scan exceeded its finite work limit")
			}
			if !mapped {
				continue
			}
			if !relationSourcePartsAgreeWithFrame(alternative, entityPorts, portRoles, subjectRole, objectRole) {
				continue
			}
			candidateSelection := relationSourceSelection{
				Status: status.OutcomeOK, Alternative: cloneRelationSourceAlternative(alternative), Variant: *variant,
				Mapping: cloneSourcePredicateMapping(mapping), Binding: cloneRelationSourceBinding(candidate),
				Sense: cloneRelationSourceSense(selectedSense), Lexeme: cloneRelationSourceLexeme(selectedLexeme),
				Frame: cloneRelationSourceFrame(frame), Validation: validation, Signature: cloneRelationSourceSignature(signature),
				HeadPort: head, RolePorts: rolePorts, PortRoles: portRoles, SubjectRole: subjectRole, ObjectRole: objectRole,
				VerbReadings:           [2]RelationSourcePartEvidence{cloneRelationSourcePart(first), cloneRelationSourcePart(second)},
				RelationPlanProvenance: strictPlan.Provenance,
				SourceProvenance:       sourceDoc.Provenance, BindingProvenance: candidate.Provenance, FrameProvenance: frame.Provenance,
				Work: out.Work,
			}
			matches = append(matches, candidateSelection)
		}
	}
	if err := ctx.Err(); err != nil {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source metadata selection was canceled")
	}
	if len(matches) > 1 {
		return fail(status.OutcomeAmbiguous, "AMBIGUOUS_RELATION_SOURCE_BINDING", "relation_source", "multiple approved source mappings and symbol bindings fully match the exact variant, sense, frame, and ports")
	}
	if len(matches) == 1 {
		matches[0].Work = out.Work
		return matches[0]
	}
	if sawSenseBinding {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_MAPPING_UNAVAILABLE", "relation_source", "no authorized equivalent symbol binding matches the source plan, selected signature, and frame roles")
	}
	return fail(status.OutcomeUnsupported, "RELATION_SOURCE_MAPPING_UNKNOWN", "relation_source", "selected source plan has no mapping for the exact verb sense and approved variant")
}

var errRelationBindingResource = fmt.Errorf("relation source metadata work limit exceeded")

func relationPlanVariant(plan engine.RelationPathPlan, id string, work *int, limit int) (*engine.RelationPathVariant, int) {
	var found *engine.RelationPathVariant
	count := 0
	for i := range plan.Variants {
		if !relationSourceChargeBinding(work, limit, 1) {
			return nil, count
		}
		if plan.Variants[i].ID == id {
			found = &plan.Variants[i]
			count++
		}
	}
	return found, count
}

func validateRelationSourceEvidence(alternative RelationSourceAlternative, variant engine.RelationPathVariant, work *int, limit int, ctx context.Context) error {
	if len(alternative.Parts) == 0 || len(alternative.Parts) > engine.MaxRelationPlanParts || alternative.VariantID != variant.ID {
		return fmt.Errorf("source alternative has invalid variant or part count")
	}
	expected := make([]relationExpectedSourcePart, 0, len(alternative.Parts))
	for index, part := range variant.Parts {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return errRelationBindingResource
		}
		if part.Slot == "verb" {
			expected = append(expected, relationExpectedSourcePart{index: index, part: part, verb: true})
			continue
		}
		tokens, err := text.Tokenize(part.Literal, text.TokenizeOptions{})
		if err != nil {
			return err
		}
		for _, token := range tokens {
			expected = append(expected, relationExpectedSourcePart{index: index, part: part, literal: token.Normalized})
		}
	}
	if len(expected) != len(alternative.Parts) {
		if len(expected) != len(alternative.Parts)+1 || expected[len(expected)-1].literal != "." || expected[len(expected)-1].part.Slot != "" {
			return fmt.Errorf("source alternative does not cover every strict plan token")
		}
		expected = expected[:len(expected)-1]
	}
	priorEnd, priorIndex := -1, -1
	for i, want := range expected {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return errRelationBindingResource
		}
		got := alternative.Parts[i]
		if got.PartIndex != want.index || got.Step != want.part.Step || got.Slot != want.part.Slot || got.Literal != want.part.Literal || got.Port != want.part.Port || got.PortPosition != want.part.Port || got.GrammaticalFunction != want.part.GrammaticalFunction || got.TokenIndex <= priorIndex || got.Span.Start < priorEnd || got.Span.Start < 0 || got.Span.End <= got.Span.Start || got.Surface == "" {
			return fmt.Errorf("source part %d differs from strict plan order or token provenance", i)
		}
		priorIndex, priorEnd = got.TokenIndex, got.Span.End
		if want.verb {
			if got.Reading == nil || got.Lexeme == nil || got.Sense == nil || got.Reading.LexemeID != got.Lexeme.ID.String() || got.Sense.LexemeID != got.Lexeme.ID {
				return fmt.Errorf("verb part %d lacks internally consistent reading, lexeme, and sense evidence", i)
			}
		} else if got.Reading != nil || got.Lexeme != nil || got.Sense != nil {
			return fmt.Errorf("literal part %d cannot carry lexical reading evidence", i)
		}
	}
	return nil
}

type relationExpectedSourcePart struct {
	index   int
	part    engine.RelationPathPart
	literal string
	verb    bool
}

func relationSourceVerbEvidence(alternative RelationSourceAlternative) (RelationSourcePartEvidence, RelationSourcePartEvidence, bool) {
	var first, second RelationSourcePartEvidence
	firstCount, secondCount := 0, 0
	for _, part := range alternative.Parts {
		if part.Slot != "verb" {
			continue
		}
		switch part.Step {
		case "first_atom":
			first, firstCount = part, firstCount+1
		case "second_atom":
			second, secondCount = part, secondCount+1
		default:
			return first, second, false
		}
	}
	return first, second, firstCount == 1 && secondCount == 1
}

func sameRelationSourceIdentity(first, second RelationSourcePartEvidence) bool {
	return first.Lexeme != nil && second.Lexeme != nil && first.Sense != nil && second.Sense != nil &&
		first.Lexeme.ID == second.Lexeme.ID && first.Lexeme.Language == second.Lexeme.Language && first.Lexeme.Lemma == second.Lexeme.Lemma && first.Lexeme.POS == second.Lexeme.POS &&
		first.Sense.ID == second.Sense.ID && first.Sense.LexemeID == second.Sense.LexemeID
}

func sameRelationSourceLexemeIdentity(source, selected lexicon.MorphLexeme) bool {
	return source.ID == selected.ID && source.Language == selected.Language && source.Lemma == selected.Lemma && source.POS == selected.POS
}

func relationSourceFrame(ctx context.Context, pack *language.SemanticLanguagePackage, frameID valency.FrameID, senseID ids.SenseID, work *int, limit int) (valency.ValencyFrame, bool, error) {
	var found valency.ValencyFrame
	count := 0
	for _, frame := range pack.ValencyIndex().FramesBySenseID(senseID) {
		if err := ctx.Err(); err != nil {
			return valency.ValencyFrame{}, false, err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return valency.ValencyFrame{}, false, nil
		}
		if frame.ID == frameID && frame.SenseID == senseID {
			found, count = frame, count+1
		}
	}
	return found, count == 1, nil
}

func relationSourceFrameRoles(frame valency.ValencyFrame) (string, string, bool) {
	if frame.ID == "" || frame.SenseID.String() == "" || len(frame.SemanticRoles) != 2 || len(frame.Arguments) != 2 {
		return "", "", false
	}
	declared := make(map[string]bool, 2)
	for _, role := range frame.SemanticRoles {
		if role == "" || declared[string(role)] {
			return "", "", false
		}
		declared[string(role)] = true
	}
	subject, object := "", ""
	for _, argument := range frame.Arguments {
		role := string(argument.Role)
		if !declared[role] {
			return "", "", false
		}
		declared[role] = false
		switch argument.GrammaticalFunction {
		case "subject":
			if subject != "" {
				return "", "", false
			}
			subject = role
		case "object":
			if object != "" {
				return "", "", false
			}
			object = role
		default:
			return "", "", false
		}
	}
	for _, remaining := range declared {
		if remaining {
			return "", "", false
		}
	}
	return subject, object, subject != "" && object != "" && subject != object
}

func relationSourceSelectedSignature(ctx context.Context, selected algebra.Document, symbol lexicon.SymbolRef, work *int, limit int) (algebra.Signature, bool, error) {
	var found algebra.Signature
	count := 0
	for _, signature := range selected.Signatures {
		if err := ctx.Err(); err != nil {
			return algebra.Signature{}, false, err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return algebra.Signature{}, false, nil
		}
		if signature.ID == string(symbol) {
			found, count = signature, count+1
		}
	}
	return found, count == 1, nil
}

func relationSourceSignaturePorts(signature algebra.Signature, headPortID string) (algebra.Port, []algebra.Port, bool) {
	if signature.ID == "" || len(signature.Ports) != 3 {
		return algebra.Port{}, nil, false
	}
	seen := map[string]bool{}
	var head algebra.Port
	entities := make([]algebra.Port, 0, 2)
	for _, port := range signature.Ports {
		if port.Role == "" || seen[port.Role] {
			return algebra.Port{}, nil, false
		}
		seen[port.Role] = true
		switch port.Type.Kind {
		case algebra.TypeEvent:
			if head.Role != "" || !bareRelationType(port.Type, algebra.TypeEvent) {
				return algebra.Port{}, nil, false
			}
			head = port
		case algebra.TypeEntity:
			if !bareRelationType(port.Type, algebra.TypeEntity) {
				return algebra.Port{}, nil, false
			}
			entities = append(entities, port)
		default:
			return algebra.Port{}, nil, false
		}
	}
	return head, entities, head.Role != "" && head.Role == headPortID && len(entities) == 2
}

func relationSourceRolePorts(ctx context.Context, rows []lexicon.SemanticRolePort, frame valency.ValencyFrame, entities []algebra.Port, headPort, subject, object string, work *int, limit int) (map[string]algebra.Port, map[string]string, bool, error) {
	roles := map[string]algebra.Port{}
	portRoles := map[string]string{}
	seenPorts := map[string]bool{}
	seenRoles := map[string]bool{}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, nil, false, err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return nil, nil, false, nil
		}
		if row.FrameID != string(frame.ID) {
			continue
		}
		if row.PortID == headPort || row.Role == "" || seenRoles[row.Role] || seenPorts[row.PortID] {
			return nil, nil, false, nil
		}
		var port *algebra.Port
		for i := range entities {
			if entities[i].Role == row.PortID {
				port = &entities[i]
				break
			}
		}
		if port == nil || !algebra.TypeEqual(row.Type, port.Type) || row.Type.Kind != algebra.TypeEntity {
			return nil, nil, false, nil
		}
		seenRoles[row.Role], seenPorts[row.PortID] = true, true
		roles[row.Role] = *port
		portRoles[row.PortID] = row.Role
	}
	if len(roles) != 2 || roles[subject].Role == "" || roles[object].Role == "" || roles[subject].Role == roles[object].Role {
		return nil, nil, false, nil
	}
	return roles, portRoles, true, nil
}

func relationSourcePartsAgreeWithFrame(alternative RelationSourceAlternative, entities []algebra.Port, portRoles map[string]string, subject, object string) bool {
	for _, evidence := range alternative.Parts {
		if evidence.Slot != "x" && evidence.Slot != "y" && evidence.Slot != "z" {
			continue
		}
		if evidence.PortPosition != "first" && evidence.PortPosition != "second" {
			return false
		}
		position := 0
		if evidence.PortPosition == "second" {
			position = 1
		}
		if position >= len(entities) || evidence.Port != evidence.PortPosition {
			return false
		}
		role := portRoles[entities[position].Role]
		if evidence.GrammaticalFunction == "subject" && role != subject || evidence.GrammaticalFunction == "object" && role != object {
			return false
		}
		if evidence.GrammaticalFunction != "subject" && evidence.GrammaticalFunction != "object" {
			return false
		}
	}
	return true
}

func relationSourceAuthorizedValidation(ctx context.Context, rows []language.TargetValidation, binding lexicon.SemanticBinding, work *int, limit int) (language.TargetValidation, bool, error) {
	var match language.TargetValidation
	count := 0
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return language.TargetValidation{}, false, err
		}
		if !relationSourceChargeBinding(work, limit, 1) {
			return language.TargetValidation{}, false, nil
		}
		if row.SenseID == binding.SenseID.String() && row.Kind == lexicon.SemanticTargetSymbol && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
			match = row
			count++
		}
	}
	if count != 1 {
		return language.TargetValidation{}, false, nil
	}
	match.Unresolved = append(match.Unresolved[:0:0], match.Unresolved...)
	return match, true, nil
}

func relationSourceChargeBinding(work *int, limit, amount int) bool {
	if amount < 0 {
		return false
	}
	*work += amount
	return *work <= limit
}

func revalidateRelationPathPlan(plan engine.RelationPathPlan, languageID string) (engine.RelationPathPlan, error) {
	encoded, err := json.Marshal(plan)
	if err != nil || len(encoded) > engine.MaxRelationPlanBytes {
		return engine.RelationPathPlan{}, fmt.Errorf("relation plan exceeds its finite encoding bound")
	}
	return engine.DecodeRelationPathPlan(encoded, languageID)
}

func bareRelationType(value algebra.Type, kind algebra.TypeKind) bool {
	return value.Kind == kind && value.SortID == "" && value.From == nil && value.To == nil
}

func cloneRelationSourceAlternative(alternative RelationSourceAlternative) RelationSourceAlternative {
	copy := alternative
	copy.Parts = make([]RelationSourcePartEvidence, len(alternative.Parts))
	for index, part := range alternative.Parts {
		copy.Parts[index] = cloneRelationSourcePart(part)
	}
	return copy
}

func cloneRelationSourcePart(part RelationSourcePartEvidence) RelationSourcePartEvidence {
	copy := part
	if part.Reading != nil {
		copy.Reading = cloneRelationSourceReading(*part.Reading)
	}
	if part.Lexeme != nil {
		lexeme := cloneRelationSourceLexeme(*part.Lexeme)
		copy.Lexeme = &lexeme
	}
	if part.Sense != nil {
		sense := cloneRelationSourceSense(*part.Sense)
		copy.Sense = &sense
	}
	return copy
}

func cloneRelationSourceBinding(binding lexicon.SemanticBinding) lexicon.SemanticBinding {
	copy := binding
	if binding.Target != nil {
		target := *binding.Target
		target.RolePorts = append([]lexicon.SemanticRolePort(nil), binding.Target.RolePorts...)
		if binding.Target.ReferencePort != nil {
			port := *binding.Target.ReferencePort
			target.ReferencePort = &port
		}
		copy.Target = &target
	}
	return copy
}

func cloneRelationSourceFrame(frame valency.ValencyFrame) valency.ValencyFrame {
	copy := frame
	copy.SemanticRoles = append([]meaning.Role(nil), frame.SemanticRoles...)
	copy.Arguments = append([]valency.ArgumentSlot(nil), frame.Arguments...)
	copy.SelectionalConstraints = append([]valency.SelectionalConstraint(nil), frame.SelectionalConstraints...)
	for index := range copy.SelectionalConstraints {
		copy.SelectionalConstraints[index].AllowedConceptIDs = append(copy.SelectionalConstraints[index].AllowedConceptIDs[:0:0], frame.SelectionalConstraints[index].AllowedConceptIDs...)
		copy.SelectionalConstraints[index].AllowedKinds = append(copy.SelectionalConstraints[index].AllowedKinds[:0:0], frame.SelectionalConstraints[index].AllowedKinds...)
	}
	copy.AllowedFamilyIDs = append(copy.AllowedFamilyIDs[:0:0], frame.AllowedFamilyIDs...)
	return copy
}

func cloneRelationSourceSignature(signature algebra.Signature) algebra.Signature {
	copy := signature
	copy.Ports = append([]algebra.Port(nil), signature.Ports...)
	for index := range copy.Ports {
		copy.Ports[index].Type = cloneRelationSourceType(signature.Ports[index].Type)
	}
	return copy
}

func cloneRelationSourceType(value algebra.Type) algebra.Type {
	copy := value
	if value.From != nil {
		from := cloneRelationSourceType(*value.From)
		copy.From = &from
	}
	if value.To != nil {
		to := cloneRelationSourceType(*value.To)
		copy.To = &to
	}
	return copy
}

func cloneSourcePredicateMapping(mapping language.SourcePredicateMapping) language.SourcePredicateMapping {
	copy := mapping
	copy.AllowedFamilyIDs = append(copy.AllowedFamilyIDs[:0:0], mapping.AllowedFamilyIDs...)
	copy.AllowedVariantIDs = append(copy.AllowedVariantIDs[:0:0], mapping.AllowedVariantIDs...)
	if mapping.ReferenceCondition != nil {
		condition := *mapping.ReferenceCondition
		copy.ReferenceCondition = &condition
	}
	return copy
}

package runtime

import (
	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// scopedSourceSelection is lexical metadata only. It deliberately contains no
// algebra expression or reference IDs; a later composer owns those decisions.
type scopedSourceSelection struct {
	Status     status.OutcomeKind
	Diagnostic status.Diagnostic
	Variant    construction.ScopedClauseVariant
	Bindings   []ScopedSourceBinding
	Nouns      map[string]scopedNounTarget
	Event      scopedEventTarget
	Provenance []provenance.SourceRef
	Work       int
}

type scopedNounTarget struct {
	Binding lexicon.SemanticBinding
	Port    lexicon.SemanticReferencePort
	Symbol  lexicon.SymbolRef
}

type scopedEventTarget struct {
	Binding       lexicon.SemanticBinding
	Frame         valency.ValencyFrame
	Mapping       language.SourcePredicateMapping
	Head          algebra.Port
	RolePorts     map[string]algebra.Port
	Symbol        lexicon.SymbolRef
	SubjectRole   string
	ObjectRole    string
	SubjectBinder string
	ObjectBinder  string
}

func selectScopedSourceMetadata(alt ScopedSourceAlternative, plan construction.ScopedClausePlan, sourcePlan *language.SourcePlan, pack *language.SemanticLanguagePackage, selected algebra.Document, requested semantic.DSEBudget) scopedSourceSelection {
	out := scopedSourceSelection{Status: status.OutcomeInvalid, Bindings: append([]ScopedSourceBinding(nil), alt.Bindings...), Nouns: map[string]scopedNounTarget{}, Event: scopedEventTarget{}}
	fail := func(kind status.OutcomeKind, code, path, message string) scopedSourceSelection {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	budget, err := semantic.NormalizeDSEBudget(requested)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_BUDGET", "budget", "source metadata selection requires a valid finite DSE budget")
	}
	if sourcePlan == nil || pack == nil || pack.LexiconIndex() == nil || pack.ValencyIndex() == nil {
		return fail(status.OutcomeInvalid, "MISSING_SOURCE_METADATA", "source", "source plan and selected semantic package are required")
	}
	if alt.VariantID == "" || alt.SubjectBinder == "" || alt.SubjectBinder != "x" && alt.SubjectBinder != "y" || alt.Negation != "outer" && alt.Negation != "inner" {
		return fail(status.OutcomeInvalid, "INVALID_ALTERNATIVE", "alternative", "complete source alternative has invalid variant or scope metadata")
	}
	var variant *construction.ScopedClauseVariant
	for i := range plan.Variants {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
		}
		if plan.Variants[i].ID == alt.VariantID {
			if variant != nil {
				return fail(status.OutcomeAmbiguous, "AMBIGUOUS_VARIANT", "variant_id", "variant identifier is duplicated")
			}
			variant = &plan.Variants[i]
		}
	}
	if variant == nil || variant.SubjectBinder != alt.SubjectBinder || variant.Negation != alt.Negation {
		return fail(status.OutcomeInvalid, "VARIANT_MISMATCH", "variant_id", "alternative does not match exactly one scoped clause variant")
	}
	out.Variant = *variant
	byBinder := map[string]ScopedSourceBinding{}
	seen := map[string]bool{}
	eventCount := 0
	for _, b := range alt.Bindings {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
		}
		if b.Binder != "x" && b.Binder != "y" && b.Binder != "event" {
			return fail(status.OutcomeInvalid, "INVALID_BINDER", "bindings", "alternative contains an unknown binder")
		}
		if b.Sense.ID.String() == "" || b.Lexeme.ID.String() == "" || b.Sense.LexemeID != b.Lexeme.ID {
			return fail(status.OutcomeInvalid, "INVALID_BINDING_EVIDENCE", "bindings", "binding sense and lexeme evidence disagree")
		}
		if b.Binder == "event" {
			eventCount++
			if eventCount > 1 {
				return fail(status.OutcomeInvalid, "DUPLICATE_EVENT_BINDING", "bindings.event", "event binder must occur exactly once")
			}
		}
		if prior, exists := byBinder[b.Binder]; exists {
			if prior.Sense.ID != b.Sense.ID || prior.Lexeme.ID != b.Lexeme.ID {
				return fail(status.OutcomeAmbiguous, "REPEATED_BINDER_DISAGREEMENT", "bindings."+b.Binder, "repeated binder evidence must resolve to the same sense and lexeme")
			}
			continue
		}
		seen[b.Binder] = true
		byBinder[b.Binder] = b
	}
	if !seen["x"] || !seen["y"] || !seen["event"] || eventCount != 1 {
		return fail(status.OutcomeIncomplete, "MISSING_BINDER_EVIDENCE", "bindings", "x, y, and exactly one event sense require selected lexical evidence")
	}
	rows, err := pack.ValidateTargets(selected)
	if err != nil {
		return fail(status.OutcomeInvalid, "TARGET_VALIDATION_FAILED", "selected", err.Error())
	}
	for _, binder := range []string{"x", "y"} {
		b := byBinder[binder]
		var candidates []lexicon.SemanticBinding
		for _, target := range pack.LexiconIndex().TargetBindingsForSense(b.Sense.ID) {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
			}
			if target.Target != nil && target.Target.Kind == lexicon.SemanticTargetReferencePredicate && target.Mapping == lexicon.MappingEquivalent && target.Status == lexicon.BindingApproved && target.TargetExecutionAuthorized() {
				candidates = append(candidates, target)
			}
		}
		valid := candidates[:0]
		for _, candidate := range candidates {
			if !uniqueAuthorizedValidation(rows, candidate, lexicon.SemanticTargetReferencePredicate) {
				continue
			}
			if candidate.Target.ReferencePort == nil || !signatureHasPort(selected, candidate.Target.SymbolRef, candidate.Target.ReferencePort.PortID, algebra.TypeEntity) {
				continue
			}
			valid = append(valid, candidate)
		}
		if len(valid) == 0 {
			return fail(status.OutcomeUnsupported, "REFERENCE_TARGET_UNAVAILABLE", "bindings."+binder, "no unique authorized equivalent reference predicate target has a selected Entity port")
		}
		if len(valid) != 1 {
			return fail(status.OutcomeAmbiguous, "AMBIGUOUS_REFERENCE_TARGET", "bindings."+binder, "multiple authorized reference predicate targets match the selected context")
		}
		t := valid[0]
		out.Nouns[binder] = scopedNounTarget{Binding: t, Port: *t.Target.ReferencePort, Symbol: t.Target.SymbolRef}
		out.Provenance = append(out.Provenance, t.Provenance)
	}
	var eventBinding *lexicon.SemanticBinding
	var eventFrame *valency.ValencyFrame
	var eventMap *language.SourcePredicateMapping
	for _, mapping := range sourcePlan.Document().Predicates {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
		}
		if mapping.SenseID != byBinder["event"].Sense.ID {
			continue
		}
		// A source-plan mapping authorizes only the variants it names. In
		// particular, an empty list must not become an implicit wildcard for
		// the scoped-source path.
		allowedVariant := false
		for _, allowedID := range mapping.AllowedVariantIDs {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
			}
			if string(allowedID) == variant.ID {
				allowedVariant = true
			}
		}
		if !allowedVariant {
			continue
		}
		frame, ok := frameFor(pack.Frames(), mapping.FrameID, mapping.SenseID)
		if !ok || !frameRolesMatchVariant(frame, *variant) {
			continue
		}
		for _, candidate := range pack.LexiconIndex().TargetBindingsForSense(mapping.SenseID) {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
			}
			if candidate.Target == nil || candidate.Target.Kind != lexicon.SemanticTargetSymbol || candidate.Mapping != lexicon.MappingEquivalent || candidate.Status != lexicon.BindingApproved || !candidate.TargetExecutionAuthorized() || !uniqueAuthorizedValidation(rows, candidate, lexicon.SemanticTargetSymbol) {
				continue
			}
			if !hasRolePortsForFrame(candidate.Target.RolePorts, string(mapping.FrameID)) {
				continue
			}
			if eventBinding != nil {
				return fail(status.OutcomeAmbiguous, "AMBIGUOUS_EVENT_TARGET", "event", "multiple authorized source mappings select the event sense")
			}
			c, f, m := candidate, frame, mapping
			eventBinding, eventFrame, eventMap = &c, &f, &m
		}
	}
	if eventBinding == nil {
		return fail(status.OutcomeUnsupported, "EVENT_TARGET_UNAVAILABLE", "event", "no unique authorized symbol target maps the selected event sense and frame")
	}
	if alt.SubjectBinder != "x" && alt.SubjectBinder != "y" {
		return fail(status.OutcomeInvalid, "INVALID_SUBJECT_BINDER", "subject_binder", "subject binder must be x or y")
	}
	roles := map[string]algebra.Port{}
	rolePortIDs := map[string]string{}
	for _, rp := range eventBinding.Target.RolePorts {
		if rp.FrameID == string(eventFrame.ID) {
			if _, duplicate := roles[rp.Role]; duplicate {
				return fail(status.OutcomeAmbiguous, "AMBIGUOUS_ROLE_PORT", "event.role_ports", "selected frame has duplicate mapped role ports")
			}
			if signaturePort(selected, eventBinding.Target.SymbolRef, rp.PortID, algebra.TypeEntity) {
				roles[rp.Role] = algebra.Port{Role: rp.PortID, Type: rp.Type}
				rolePortIDs[rp.Role] = rp.PortID
			}
		}
	}
	if len(roles) != 2 {
		return fail(status.OutcomeUnsupported, "EVENT_ROLE_PORTS_UNAVAILABLE", "event.role_ports", "selected frame requires exactly two distinct mapped Entity role ports")
	}
	args := map[string]valency.ArgumentSlot{}
	for _, a := range eventFrame.Arguments {
		args[string(a.Role)] = a
	}
	var subjectRole, objectRole string
	for role, arg := range args {
		if arg.GrammaticalFunction == "subject" {
			if subjectRole != "" {
				return fail(status.OutcomeAmbiguous, "AMBIGUOUS_FRAME_SUBJECT", "frame.arguments", "frame has multiple subject roles")
			}
			subjectRole = role
		}
		if arg.GrammaticalFunction == "object" && (variant.FrameObjectCase == "" || arg.Case == variant.FrameObjectCase) {
			if objectRole != "" {
				return fail(status.OutcomeAmbiguous, "AMBIGUOUS_FRAME_OBJECT", "frame.arguments", "frame has multiple object roles for selected case")
			}
			objectRole = role
		}
	}
	if subjectRole == "" || objectRole == "" || subjectRole == objectRole || len(roles) != 2 || roles[subjectRole].Role == "" || roles[objectRole].Role == "" {
		return fail(status.OutcomeUnsupported, "FRAME_ROLE_MAPPING_UNAVAILABLE", "frame.arguments", "selected frame does not map distinct subject and explicitly cased object roles")
	}
	headMatches := 0
	var head algebra.Port
	for _, p := range signaturePorts(selected, eventBinding.Target.SymbolRef) {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "source metadata selection work limit exceeded")
		}
		if p.Type.Kind == algebra.TypeEvent {
			headMatches++
			head = p
		}
	}
	if headMatches != 1 || head.Role != eventMap.HeadPortID || len(signaturePorts(selected, eventBinding.Target.SymbolRef)) != 3 || rolePortIDs[subjectRole] == rolePortIDs[objectRole] {
		return fail(status.OutcomeAmbiguous, "EVENT_HEAD_PORT_AMBIGUOUS", "event.signature", "selected symbol must expose exactly one Event port")
	}
	entityPorts := 0
	for _, p := range signaturePorts(selected, eventBinding.Target.SymbolRef) {
		if p.Type.Kind == algebra.TypeEntity {
			entityPorts++
		}
	}
	if entityPorts != 2 || len(roles) != 2 || rolePortIDs[subjectRole] == "" || rolePortIDs[objectRole] == "" || !signaturePort(selected, eventBinding.Target.SymbolRef, rolePortIDs[subjectRole], algebra.TypeEntity) || !signaturePort(selected, eventBinding.Target.SymbolRef, rolePortIDs[objectRole], algebra.TypeEntity) {
		return fail(status.OutcomeUnsupported, "EVENT_SIGNATURE_SHAPE", "event.signature", "event signature must contain one Event port and exactly two mapped Entity ports")
	}
	objectBinder := "x"
	if variant.SubjectBinder == "x" {
		objectBinder = "y"
	}
	out.Event = scopedEventTarget{Binding: *eventBinding, Frame: *eventFrame, Mapping: *eventMap, Head: head, RolePorts: roles, Symbol: eventBinding.Target.SymbolRef, SubjectRole: subjectRole, ObjectRole: objectRole, SubjectBinder: variant.SubjectBinder, ObjectBinder: objectBinder}
	out.Provenance = append(out.Provenance, sourcePlan.Document().Provenance, eventBinding.Provenance, eventFrame.Provenance)
	out.Status = status.OutcomeOK
	return out
}

func uniqueAuthorizedValidation(rows []language.TargetValidation, binding lexicon.SemanticBinding, kind lexicon.SemanticTargetKind) bool {
	n := 0
	for _, row := range rows {
		if row.SenseID == binding.SenseID.String() && row.Kind == kind && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
			n++
		}
	}
	return n == 1
}

func signaturePorts(doc algebra.Document, symbol lexicon.SymbolRef) []algebra.Port {
	for _, sig := range doc.Signatures {
		if sig.ID == string(symbol) {
			return append([]algebra.Port(nil), sig.Ports...)
		}
	}
	return nil
}
func signaturePort(doc algebra.Document, symbol lexicon.SymbolRef, id string, kind algebra.TypeKind) bool {
	for _, p := range signaturePorts(doc, symbol) {
		if p.Role == id && p.Type.Kind == kind {
			return true
		}
	}
	return false
}
func signatureHasPort(doc algebra.Document, symbol lexicon.SymbolRef, id string, kind algebra.TypeKind) bool {
	return signaturePort(doc, symbol, id, kind)
}
func frameFor(frames []valency.ValencyFrame, id valency.FrameID, sense interface{ String() string }) (valency.ValencyFrame, bool) {
	for _, f := range frames {
		if f.ID == id && f.SenseID.String() == sense.String() {
			return f, true
		}
	}
	return valency.ValencyFrame{}, false
}
func hasRolePortsForFrame(rows []lexicon.SemanticRolePort, frame string) bool {
	n := 0
	for _, r := range rows {
		if r.FrameID == frame {
			n++
		}
	}
	return n >= 2
}

func frameRolesMatchVariant(frame valency.ValencyFrame, variant construction.ScopedClauseVariant) bool {
	var subj, obj string
	for _, arg := range frame.Arguments {
		if arg.GrammaticalFunction == "subject" {
			if subj != "" {
				return false
			}
			subj = string(arg.Role)
		}
		if arg.GrammaticalFunction == "object" && (variant.FrameObjectCase == "" || arg.Case == variant.FrameObjectCase) {
			if obj != "" {
				return false
			}
			obj = string(arg.Role)
		}
	}
	if subj == "" || obj == "" || subj == obj {
		return false
	}
	return true
}

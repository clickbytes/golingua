package engine

import (
	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// selectAlgebraEventRelationBinding resolves the exact three-port event
// predicate and its two explicitly mapped Entity roles from selected metadata.
// The caller must have checked the expression before constructing plan.
func selectAlgebraEventRelationBinding(plan algebraRelationEventPathPlan, pack *language.SemanticLanguagePackage, selected algebra.Document, requested semantic.DSEBudget) algebraRelationBindingResult {
	return selectAlgebraEventRelationBindingForFrame(plan, pack, selected, requested, "")
}

// selectAlgebraEventRelationBindingForFrame applies an explicit frame constraint
// for source-plan verification. The ordinary target selector intentionally
// remains unconstrained and continues to report multiple valid frames.
func selectAlgebraEventRelationBindingForFrame(plan algebraRelationEventPathPlan, pack *language.SemanticLanguagePackage, selected algebra.Document, requested semantic.DSEBudget, frameID valency.FrameID) algebraRelationBindingResult {
	out := algebraRelationBindingResult{Status: status.OutcomeInvalid}
	fail := func(kind status.OutcomeKind, code, path, message string) algebraRelationBindingResult {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	budget, err := semantic.NormalizeDSEBudget(requested)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_BUDGET", "budget", "event relation binding requires a valid finite DSE budget")
	}
	if pack == nil || pack.LexiconIndex() == nil || pack.ValencyIndex() == nil || pack.Language() == "" {
		return fail(status.OutcomeInvalid, "MISSING_RELATION_METADATA", "package", "a selected language package with lexicon and valency data is required")
	}
	if !validAlgebraEventRelationPlan(plan) {
		return fail(status.OutcomeInvalid, "INVALID_EVENT_RELATION_PATH", "relation.signature", "event relation path must have one Event port and two ordered Entity ports matching both atoms")
	}
	selectedSignature, found := selectedRelationSignature(selected, plan.Symbol)
	if !found || !sameRelationSignature(selectedSignature, plan.Signature) {
		return fail(status.OutcomeInvalid, "SELECTED_SIGNATURE_MISMATCH", "selected.signatures", "selected declaration must contain the exact typed three-port signature used by the checked path")
	}
	rows, err := pack.ValidateTargets(selected)
	if err != nil {
		return fail(status.OutcomeInvalid, "TARGET_VALIDATION_FAILED", "selected", err.Error())
	}
	rowCounts := make(map[string]int)
	for _, row := range rows {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event relation target validation scan exceeded its finite work limit")
		}
		if row.Kind == lexicon.SemanticTargetSymbol && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
			rowCounts[row.SenseID]++
		}
	}
	bindings := pack.LexiconIndex().Document().Bindings
	if len(bindings) > lexicon.MaxBindings {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.bindings", "lexicon binding scan exceeds its finite cap")
	}
	var matches []algebraRelationBinding
	sawSymbol := false
	for _, candidate := range bindings {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event relation metadata selection work limit exceeded")
		}
		if candidate.Target == nil || candidate.Target.Kind != lexicon.SemanticTargetSymbol || candidate.Target.SymbolRef != lexicon.SymbolRef(plan.Symbol) {
			continue
		}
		sawSymbol = true
		if candidate.Mapping != lexicon.MappingEquivalent || candidate.Status != lexicon.BindingApproved || !candidate.TargetExecutionAuthorized() || rowCounts[candidate.SenseID.String()] != 1 {
			continue
		}
		sense, ok := pack.LexiconIndex().BySenseID(candidate.SenseID)
		if !ok {
			continue
		}
		lexeme, ok := pack.LexiconIndex().ByLexemeID(sense.LexemeID)
		if !ok || sense.LexemeID != lexeme.ID || lexeme.Language != pack.Language() || lexeme.POS != "verb" {
			continue
		}
		var validation language.TargetValidation
		validationCount := 0
		for _, row := range rows {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event relation validation lookup exceeded its finite work limit")
			}
			if row.SenseID == candidate.SenseID.String() && row.Kind == lexicon.SemanticTargetSymbol && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
				validation, validationCount = cloneRelationValidation(row), validationCount+1
			}
		}
		if validationCount != 1 {
			continue
		}
		frames := pack.ValencyIndex().FramesBySenseID(candidate.SenseID)
		for _, frame := range frames {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event relation frame scan exceeded its finite work limit")
			}
			if frame.SenseID != candidate.SenseID || frameID != "" && frame.ID != frameID || !relationFrameComplete(frame) {
				continue
			}
			subjectRole, objectRole, rolesOK := eventRelationFrameRoles(frame)
			if !rolesOK {
				continue
			}
			portRoles, mapped := eventRelationPortRoles(candidate.Target.RolePorts, frame.ID, plan, budget.MaxWork-out.Work)
			out.Work += mapped.work
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event relation role-port scan exceeded its finite work limit")
			}
			if !mapped.ok || len(portRoles) != 2 {
				continue
			}
			mappedRoles := make(map[string]string, 2)
			valid := true
			for port, role := range portRoles {
				if role != subjectRole && role != objectRole {
					valid = false
					break
				}
				if _, exists := mappedRoles[role]; exists {
					valid = false
					break
				}
				mappedRoles[role] = port
			}
			if !valid || len(mappedRoles) != 2 || mappedRoles[subjectRole] == "" || mappedRoles[objectRole] == "" {
				continue
			}
			matches = append(matches, algebraRelationBinding{
				SemanticBinding: cloneRelationSemanticBinding(candidate),
				LexicalSense:    cloneRelationSense(sense),
				Lexeme:          cloneRelationLexeme(lexeme),
				Validation:      validation,
				SenseID:         candidate.SenseID.String(),
				LexemeID:        sense.LexemeID.String(),
				Language:        string(lexeme.Language),
				Lemma:           lexeme.Lemma,
				POS:             string(lexeme.POS),
				Symbol:          candidate.Target.SymbolRef,
				Frame:           cloneEventRelationFrame(frame),
				SubjectRole:     subjectRole,
				ObjectRole:      objectRole,
				PortRoles:       portRoles,
				PortIDs: map[string]string{
					plan.Signature.Ports[plan.EntityPortPositions[0]].Role: plan.Signature.Ports[plan.EntityPortPositions[0]].Role,
					plan.Signature.Ports[plan.EntityPortPositions[1]].Role: plan.Signature.Ports[plan.EntityPortPositions[1]].Role,
				},
			})
		}
	}
	if len(matches) > 1 {
		return fail(status.OutcomeAmbiguous, "AMBIGUOUS_EVENT_RELATION_BINDING", "relation", "multiple approved sense/frame mappings fully match the selected event symbol and ports")
	}
	if len(matches) == 1 {
		out.Status, out.Binding = status.OutcomeOK, matches[0]
		return out
	}
	if sawSymbol {
		return fail(status.OutcomeUnsupported, "EVENT_RELATION_BINDING_UNAVAILABLE", "relation", "symbol has no approved equivalent binding with a complete matching event frame and role mapping")
	}
	return fail(status.OutcomeUnknown, "EVENT_RELATION_SYMBOL_UNKNOWN", "relation.symbol", "selected semantic package has no binding for the exact IR event symbol")
}

func validAlgebraEventRelationPlan(plan algebraRelationEventPathPlan) bool {
	if plan.Symbol == "" || plan.Signature.ID != plan.Symbol || len(plan.Signature.Ports) != 3 || plan.EventPortPosition < 0 || plan.EventPortPosition >= len(plan.Signature.Ports) {
		return false
	}
	if plan.EventPort != plan.Signature.Ports[plan.EventPortPosition].Role || !bareRelationPortType(plan.Signature.Ports[plan.EventPortPosition].Type, algebra.TypeEvent) {
		return false
	}
	seen := map[int]bool{plan.EventPortPosition: true}
	for _, position := range plan.EntityPortPositions {
		if position < 0 || position >= len(plan.Signature.Ports) || seen[position] || !bareRelationPortType(plan.Signature.Ports[position].Type, algebra.TypeEntity) {
			return false
		}
		seen[position] = true
	}
	if len(seen) != len(plan.Signature.Ports) || plan.Signature.Ports[plan.EntityPortPositions[0]].Role == plan.Signature.Ports[plan.EntityPortPositions[1]].Role {
		return false
	}
	if plan.FirstPorts != [2]string{plan.Signature.Ports[plan.EntityPortPositions[0]].Role, plan.Signature.Ports[plan.EntityPortPositions[1]].Role} || plan.SecondPorts != plan.FirstPorts {
		return false
	}
	if plan.FirstEvent == nil || plan.SecondEvent == nil || plan.FirstEventBinder == "" || plan.SecondEventBinder == "" || plan.FirstEventBinder == plan.SecondEventBinder || plan.FirstEventBinder != plan.FirstEvent.ID || plan.SecondEventBinder != plan.SecondEvent.ID || plan.FirstEventAtom != plan.FirstEvent.Body || plan.SecondEventAtom != plan.SecondEvent.Body || plan.FirstAtom != plan.FirstEventAtom || plan.SecondAtom != plan.SecondEventAtom || plan.FirstEventPath == "" || plan.SecondEventPath == "" || plan.FirstAtomPath == "" || plan.SecondAtomPath == "" {
		return false
	}
	for index, term := range []*algebra.Term{plan.FirstEvent, plan.SecondEvent} {
		if term == nil || term.Kind != algebra.TermExists || !termHasKind(term, algebra.TypeEvent) || term.ID == "" || term.Body == nil || term.Body.Kind != algebra.TermAtom || term.Body.Symbol != plan.Symbol || len(term.Body.Args) != 3 {
			return false
		}
		for position, port := range plan.Signature.Ports {
			arg := term.Body.Args[position]
			if arg.Role != port.Role || arg.Value == nil || arg.Value.Kind != algebra.TermVar {
				return false
			}
			if position == plan.EventPortPosition && arg.Value.ID != term.ID {
				return false
			}
			if position == plan.EntityPortPositions[0] || position == plan.EntityPortPositions[1] {
				entityIDs := [2]string{}
				entityIDs[0] = term.Body.Args[plan.EntityPortPositions[0]].Value.ID
				entityIDs[1] = term.Body.Args[plan.EntityPortPositions[1]].Value.ID
				if index == 0 && entityIDs != [2]string{plan.FirstBinder, plan.WitnessBinder} {
					return false
				}
				if index == 1 && entityIDs != [2]string{plan.WitnessBinder, plan.SecondBinder} {
					return false
				}
			}
		}
	}
	return plan.FirstBinder != "" && plan.SecondBinder != "" && plan.WitnessBinder != "" && plan.FirstBinder != plan.SecondBinder && plan.FirstBinder != plan.WitnessBinder && plan.SecondBinder != plan.WitnessBinder
}

func bareRelationPortType(typeValue algebra.Type, kind algebra.TypeKind) bool {
	return typeValue.Kind == kind && typeValue.SortID == "" && typeValue.From == nil && typeValue.To == nil
}

func eventRelationFrameRoles(frame valency.ValencyFrame) (string, string, bool) {
	subject, object := "", ""
	roles := make(map[meaning.Role]bool, 2)
	for _, role := range frame.SemanticRoles {
		roles[role] = true
	}
	for _, arg := range frame.Arguments {
		if !roles[arg.Role] {
			return "", "", false
		}
		roles[arg.Role] = false
		switch arg.GrammaticalFunction {
		case "subject":
			if subject != "" {
				return "", "", false
			}
			subject = string(arg.Role)
		case "object":
			if object != "" {
				return "", "", false
			}
			object = string(arg.Role)
		default:
			return "", "", false
		}
	}
	for _, remaining := range roles {
		if remaining {
			return "", "", false
		}
	}
	return subject, object, subject != "" && object != "" && subject != object
}

type eventRelationPortRoleScan struct {
	ok   bool
	work int
}

func eventRelationPortRoles(rows []lexicon.SemanticRolePort, frame valency.FrameID, plan algebraRelationEventPathPlan, remaining int) (map[string]string, eventRelationPortRoleScan) {
	result := make(map[string]string, 2)
	scan := eventRelationPortRoleScan{ok: true}
	seenPorts := make(map[string]bool, 2)
	seenRoles := make(map[string]bool, 2)
	for _, row := range rows {
		scan.work++
		if scan.work > remaining {
			scan.ok = false
			return nil, scan
		}
		if row.FrameID != string(frame) {
			continue
		}
		if row.PortID == plan.EventPort || seenPorts[row.PortID] || seenRoles[row.Role] || row.Role == "" {
			scan.ok = false
			return nil, scan
		}
		position := -1
		for _, candidate := range plan.EntityPortPositions {
			if plan.Signature.Ports[candidate].Role == row.PortID {
				position = candidate
				break
			}
		}
		if position < 0 || !algebra.TypeEqual(row.Type, plan.Signature.Ports[position].Type) || row.Type.Kind != algebra.TypeEntity {
			scan.ok = false
			return nil, scan
		}
		seenPorts[row.PortID], seenRoles[row.Role] = true, true
		result[row.PortID] = row.Role
	}
	if len(result) != 2 {
		scan.ok = false
	}
	return result, scan
}

func cloneEventRelationFrame(frame valency.ValencyFrame) valency.ValencyFrame {
	frame.SemanticRoles = append([]meaning.Role(nil), frame.SemanticRoles...)
	frame.Arguments = append([]valency.ArgumentSlot(nil), frame.Arguments...)
	frame.SelectionalConstraints = append([]valency.SelectionalConstraint(nil), frame.SelectionalConstraints...)
	for index := range frame.SelectionalConstraints {
		frame.SelectionalConstraints[index].AllowedConceptIDs = append(frame.SelectionalConstraints[index].AllowedConceptIDs[:0:0], frame.SelectionalConstraints[index].AllowedConceptIDs...)
		frame.SelectionalConstraints[index].AllowedKinds = append(frame.SelectionalConstraints[index].AllowedKinds[:0:0], frame.SelectionalConstraints[index].AllowedKinds...)
	}
	frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs[:0:0], frame.AllowedFamilyIDs...)
	return frame
}

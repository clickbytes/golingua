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

type algebraRelationBinding struct {
	SemanticBinding lexicon.SemanticBinding
	LexicalSense    lexicon.LexicalSense
	Lexeme          lexicon.MorphLexeme
	Validation      language.TargetValidation
	SenseID         string
	LexemeID        string
	Language        string
	Lemma           string
	POS             string
	Symbol          lexicon.SymbolRef
	Frame           valency.ValencyFrame
	SubjectRole     string
	ObjectRole      string
	PortRoles       map[string]string
	PortIDs         map[string]string
}

type algebraRelationBindingResult struct {
	Status     status.OutcomeKind
	Diagnostic status.Diagnostic
	Binding    algebraRelationBinding
	Work       int
}

// selectAlgebraRelationBinding resolves the exact symbol and ordered ports in
// a structural relation path against the caller-selected semantic package and
// declaration document. The caller must have checked the IR first.
func selectAlgebraRelationBinding(plan algebraRelationPathPlan, pack *language.SemanticLanguagePackage, selected algebra.Document, requested semantic.DSEBudget) algebraRelationBindingResult {
	out := algebraRelationBindingResult{Status: status.OutcomeInvalid}
	fail := func(kind status.OutcomeKind, code, path, message string) algebraRelationBindingResult {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	budget, err := semantic.NormalizeDSEBudget(requested)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_BUDGET", "budget", "relation binding requires a valid finite DSE budget")
	}
	if pack == nil || pack.LexiconIndex() == nil || pack.ValencyIndex() == nil || plan.Symbol == "" || len(plan.Signature.Ports) != 2 {
		return fail(status.OutcomeInvalid, "MISSING_RELATION_METADATA", "relation", "a selected package and complete binary relation path are required")
	}
	if plan.Signature.ID != plan.Symbol || plan.Signature.Ports[0].Type.Kind != algebra.TypeEntity || plan.Signature.Ports[1].Type.Kind != algebra.TypeEntity || plan.Signature.Ports[0].Role == plan.Signature.Ports[1].Role || plan.FirstPorts != [2]string{plan.Signature.Ports[0].Role, plan.Signature.Ports[1].Role} || plan.SecondPorts != plan.FirstPorts {
		return fail(status.OutcomeInvalid, "INVALID_RELATION_PATH", "relation.signature", "relation path ports do not match its ordered Entity signature")
	}
	selectedSignature, hasSelectedSignature := selectedRelationSignature(selected, plan.Symbol)
	if !hasSelectedSignature || !sameRelationSignature(selectedSignature, plan.Signature) {
		return fail(status.OutcomeInvalid, "SELECTED_SIGNATURE_MISMATCH", "selected.signatures", "selected declaration must contain the exact typed relation signature used by the checked path")
	}
	rows, err := pack.ValidateTargets(selected)
	if err != nil {
		return fail(status.OutcomeInvalid, "TARGET_VALIDATION_FAILED", "selected", err.Error())
	}
	rowCounts := map[string]int{}
	for _, row := range rows {
		if row.Kind == lexicon.SemanticTargetSymbol && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
			rowCounts[row.SenseID]++
		}
	}

	bindings := pack.LexiconIndex().Document().Bindings
	if len(bindings) > lexicon.MaxBindings {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.bindings", "lexicon binding scan exceeds its finite cap")
	}
	var matches []algebraRelationBinding
	var sawSymbol bool
	for _, candidate := range bindings {
		out.Work++
		if out.Work > budget.MaxWork {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation metadata selection work limit exceeded")
		}
		if candidate.Target == nil || candidate.Target.Kind != lexicon.SemanticTargetSymbol || candidate.Target.SymbolRef != lexicon.SymbolRef(plan.Symbol) {
			continue
		}
		sawSymbol = true
		if candidate.Mapping != lexicon.MappingEquivalent || candidate.Status != lexicon.BindingApproved || !candidate.TargetExecutionAuthorized() || rowCounts[candidate.SenseID.String()] != 1 {
			continue
		}
		lexicalSense, ok := pack.LexiconIndex().BySenseID(candidate.SenseID)
		if !ok {
			continue
		}
		lexeme, ok := pack.LexiconIndex().ByLexemeID(lexicalSense.LexemeID)
		if !ok {
			continue
		}
		if lexicalSense.LexemeID != lexeme.ID || lexeme.Language != pack.Language() || lexeme.POS != "verb" {
			continue
		}
		var validation language.TargetValidation
		for _, row := range rows {
			if row.SenseID == candidate.SenseID.String() && row.Kind == lexicon.SemanticTargetSymbol && row.Status == lexicon.BindingApproved && row.Mapping == lexicon.MappingEquivalent && row.Authorized && len(row.Unresolved) == 0 {
				validation = cloneRelationValidation(row)
				break
			}
		}
		frames := pack.ValencyIndex().FramesBySenseID(candidate.SenseID)
		for _, frame := range frames {
			out.Work++
			if out.Work > budget.MaxWork {
				return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation metadata selection work limit exceeded")
			}
			if !relationFrameComplete(frame) {
				continue
			}
			roleMap := map[string]lexicon.SemanticRolePort{}
			valid := true
			for _, rp := range candidate.Target.RolePorts {
				if rp.FrameID != string(frame.ID) {
					continue
				}
				portIndex := -1
				for i, port := range plan.Signature.Ports {
					if port.Role == rp.PortID {
						portIndex = i
						break
					}
				}
				if _, duplicate := roleMap[rp.Role]; duplicate || portIndex < 0 || !algebra.TypeEqual(rp.Type, plan.Signature.Ports[portIndex].Type) {
					valid = false
					break
				}
				roleMap[rp.Role] = rp
			}
			if !valid || len(roleMap) != 2 {
				continue
			}
			arguments := map[string]valency.ArgumentSlot{}
			for _, arg := range frame.Arguments {
				arguments[string(arg.Role)] = arg
			}
			var subject, object string
			for role, arg := range arguments {
				if arg.GrammaticalFunction == "subject" {
					if subject != "" {
						valid = false
						break
					}
					subject = role
				}
				if arg.GrammaticalFunction == "object" {
					if object != "" {
						valid = false
						break
					}
					object = role
				}
			}
			if !valid || subject == "" || object == "" || subject == object || len(arguments) != 2 {
				continue
			}
			firstRole, secondRole := "", ""
			for role, rp := range roleMap {
				if rp.PortID == plan.Signature.Ports[0].Role {
					firstRole = role
				}
				if rp.PortID == plan.Signature.Ports[1].Role {
					secondRole = role
				}
			}
			if !((firstRole == subject && secondRole == object) || (firstRole == object && secondRole == subject)) || roleMap[firstRole].PortID == roleMap[secondRole].PortID {
				continue
			}
			matches = append(matches, algebraRelationBinding{
				SemanticBinding: cloneRelationSemanticBinding(candidate), LexicalSense: cloneRelationSense(lexicalSense), Lexeme: cloneRelationLexeme(lexeme), Validation: validation,
				SenseID: candidate.SenseID.String(), LexemeID: lexicalSense.LexemeID.String(), Language: string(lexeme.Language), Lemma: lexeme.Lemma, POS: string(lexeme.POS), Symbol: candidate.Target.SymbolRef,
				Frame: frame, SubjectRole: subject, ObjectRole: object,
				PortRoles: map[string]string{plan.Signature.Ports[0].Role: firstRole, plan.Signature.Ports[1].Role: secondRole},
				PortIDs:   map[string]string{plan.Signature.Ports[0].Role: plan.Signature.Ports[0].Role, plan.Signature.Ports[1].Role: plan.Signature.Ports[1].Role},
			})
		}
	}
	if len(matches) > 1 {
		return fail(status.OutcomeAmbiguous, "AMBIGUOUS_RELATION_BINDING", "relation", "multiple approved sense/frame mappings fully match the selected symbol and ports")
	}
	if len(matches) == 1 {
		out.Status, out.Binding = status.OutcomeOK, matches[0]
		return out
	}
	if sawSymbol {
		return fail(status.OutcomeUnsupported, "RELATION_BINDING_UNAVAILABLE", "relation", "symbol has no approved equivalent binding with a complete matching frame and role mapping")
	}
	return fail(status.OutcomeUnknown, "RELATION_SYMBOL_UNKNOWN", "relation.symbol", "selected semantic package has no binding for the exact IR symbol")
}

func selectedRelationSignature(document algebra.Document, symbol string) (algebra.Signature, bool) {
	var found algebra.Signature
	count := 0
	for _, signature := range document.Signatures {
		if signature.ID == symbol {
			found = signature
			count++
		}
	}
	return found, count == 1
}

func sameRelationSignature(left, right algebra.Signature) bool {
	if left.ID != right.ID || len(left.Ports) != len(right.Ports) {
		return false
	}
	for index := range left.Ports {
		if left.Ports[index].Role != right.Ports[index].Role || !algebra.TypeEqual(left.Ports[index].Type, right.Ports[index].Type) {
			return false
		}
	}
	return true
}

func cloneRelationSemanticBinding(binding lexicon.SemanticBinding) lexicon.SemanticBinding {
	result := binding
	if binding.Target != nil {
		target := *binding.Target
		target.RolePorts = append([]lexicon.SemanticRolePort(nil), binding.Target.RolePorts...)
		if binding.Target.ReferencePort != nil {
			reference := *binding.Target.ReferencePort
			target.ReferencePort = &reference
		}
		result.Target = &target
	}
	return result
}

func cloneRelationSense(sense lexicon.LexicalSense) lexicon.LexicalSense {
	result := sense
	result.Usage = append([]string(nil), sense.Usage...)
	result.ValencyIDs = append([]lexicon.ValencyID(nil), sense.ValencyIDs...)
	return result
}

func cloneRelationLexeme(lexeme lexicon.MorphLexeme) lexicon.MorphLexeme {
	result := lexeme
	result.Properties = append([]lexicon.Property(nil), lexeme.Properties...)
	return result
}

func cloneRelationValidation(validation language.TargetValidation) language.TargetValidation {
	result := validation
	result.Unresolved = append(result.Unresolved[:0:0], validation.Unresolved...)
	return result
}

func relationFrameComplete(frame valency.ValencyFrame) bool {
	if frame.ID == "" || frame.SenseID.String() == "" || len(frame.SemanticRoles) != 2 || len(frame.Arguments) != 2 {
		return false
	}
	roles := map[meaning.Role]bool{}
	for _, role := range frame.SemanticRoles {
		if role == "" || roles[role] {
			return false
		}
		roles[role] = true
	}
	for _, arg := range frame.Arguments {
		if !roles[arg.Role] {
			return false
		}
		roles[arg.Role] = false
	}
	for _, present := range roles {
		if present {
			return false
		}
	}
	return true
}

package engine

import (
	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/status"
)

// algebraScopeLexicalMatch is the language-local evidence selected for a
// scope atom. Binding carries its provenance and lexical identity; semantic
// roles retain the explicit target mappings used for the event participants.
type algebraScopeLexicalMatch struct {
	Binding    lexicon.SemanticBinding
	Sense      lexicon.LexicalSense
	Lexeme     lexicon.MorphLexeme
	Validation language.TargetValidation
	OuterRole  string
	InnerRole  string
	OuterPort  string
	InnerPort  string
	EventPort  string
}

type algebraScopeSelection struct {
	Status status.OutcomeKind
	Code   string
	Match  *algebraScopeLexicalMatch
}

func selectAlgebraScopeBindings(plan algebraScopePlan, pack *language.SemanticLanguagePackage, validations []language.TargetValidation, signatures map[string]algebra.Signature) (outer, inner, event algebraScopeSelection) {
	if pack == nil || pack.LexiconIndex() == nil {
		failure := algebraScopeSelection{Status: status.OutcomeInvalid, Code: "MISSING_SEMANTIC_PACKAGE"}
		return failure, failure, failure
	}
	outer = selectScopeUnaryBinding(pack, validations, plan.Restrictor, plan.OuterPort)
	inner = selectScopeUnaryBinding(pack, validations, plan.Predicate, plan.InnerPort)
	event = selectScopeEventBinding(pack, validations, plan, signatures)
	return outer, inner, event
}

func selectScopeUnaryBinding(pack *language.SemanticLanguagePackage, validations []language.TargetValidation, atom *algebra.Term, port string) algebraScopeSelection {
	if atom == nil || atom.Kind != algebra.TermAtom || atom.Symbol == "" || port == "" {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_UNARY_ATOM"}
	}
	var candidates []algebraScopeLexicalMatch
	matched := false
	invalidMatch := false
	for _, binding := range pack.LexiconIndex().Document().Bindings {
		target := binding.Target
		if target == nil || target.Kind != lexicon.SemanticTargetReferencePredicate || string(target.SymbolRef) != atom.Symbol || target.ReferencePort == nil || target.ReferencePort.PortID != port {
			continue
		}
		matched = true
		validation, valid := scopeBindingValidation(validations, binding)
		sense, senseOK := pack.LexiconIndex().BySenseID(binding.SenseID)
		lexeme, lexemeOK := pack.LexiconIndex().ByLexemeID(sense.LexemeID)
		if !binding.TargetExecutionAuthorized() || !valid || !targetValidationAuthorized(validations, binding) || !validation.Authorized || validation.Status != lexicon.BindingApproved || validation.Mapping != lexicon.MappingEquivalent || len(validation.Unresolved) != 0 || !senseOK || !lexemeOK || lexeme.Language != pack.Language() || lexeme.POS != "noun" || !algebra.TypeEqual(target.ReferencePort.Type, algebra.Type{Kind: algebra.TypeEntity}) {
			invalidMatch = true
			continue
		}
		candidates = append(candidates, algebraScopeLexicalMatch{Binding: binding, Sense: sense, Lexeme: lexeme, Validation: validation})
	}
	if len(candidates) > 1 {
		return algebraScopeSelection{Status: status.OutcomeAmbiguous, Code: "AMBIGUOUS_SCOPE_BINDING"}
	}
	if len(candidates) == 1 {
		return algebraScopeSelection{Status: status.OutcomeOK, Match: &candidates[0]}
	}
	if matched || invalidMatch {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_BINDING"}
	}
	return algebraScopeSelection{Status: status.OutcomeUnknown, Code: "MISSING_SCOPE_BINDING"}
}

func selectScopeEventBinding(pack *language.SemanticLanguagePackage, validations []language.TargetValidation, plan algebraScopePlan, signatures map[string]algebra.Signature) algebraScopeSelection {
	atom := plan.EventAtom
	if atom == nil || atom.Kind != algebra.TermAtom || atom.Symbol == "" {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_EVENT_ATOM"}
	}
	signature, exists := signatures[atom.Symbol]
	if !exists || len(signature.Ports) != len(atom.Args) || plan.EventPort == "" || plan.EventOuterPort == "" || plan.EventInnerPort == "" {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_EVENT_SIGNATURE"}
	}
	portTypes := make(map[string]algebra.Type, len(signature.Ports))
	for _, port := range signature.Ports {
		if port.Role == "" || portTypes[port.Role].Kind != "" {
			return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_EVENT_SIGNATURE"}
		}
		portTypes[port.Role] = port.Type
	}
	if portTypes[plan.EventPort].Kind != algebra.TypeEvent || portTypes[plan.EventOuterPort].Kind != algebra.TypeEntity || portTypes[plan.EventInnerPort].Kind != algebra.TypeEntity {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_EVENT_SIGNATURE"}
	}
	var candidates []algebraScopeLexicalMatch
	matched, invalidMatch := false, false
	for _, binding := range pack.LexiconIndex().Document().Bindings {
		target := binding.Target
		if target == nil || target.Kind != lexicon.SemanticTargetSymbol || string(target.SymbolRef) != atom.Symbol {
			continue
		}
		matched = true
		validation, valid := scopeBindingValidation(validations, binding)
		sense, senseOK := pack.LexiconIndex().BySenseID(binding.SenseID)
		lexeme, lexemeOK := pack.LexiconIndex().ByLexemeID(sense.LexemeID)
		if !binding.TargetExecutionAuthorized() || !valid || !targetValidationAuthorized(validations, binding) || !validation.Authorized || validation.Status != lexicon.BindingApproved || validation.Mapping != lexicon.MappingEquivalent || len(validation.Unresolved) != 0 || !senseOK || !lexemeOK || lexeme.Language != pack.Language() || lexeme.POS != "verb" {
			invalidMatch = true
			continue
		}
		mapping := make(map[string]lexicon.SemanticRolePort, len(target.RolePorts))
		bad := len(target.RolePorts) != 2
		frame := ""
		roles := make(map[string]bool, 2)
		for _, row := range target.RolePorts {
			if row.PortID == "" || row.Role == "" || row.FrameID == "" || mapping[row.PortID].PortID != "" || roles[row.Role] {
				bad = true
				break
			}
			if frame == "" {
				frame = row.FrameID
			} else if frame != row.FrameID {
				bad = true
			}
			mapping[row.PortID], roles[row.Role] = row, true
		}
		outerMapping, hasOuter := mapping[plan.EventOuterPort]
		innerMapping, hasInner := mapping[plan.EventInnerPort]
		if !hasOuter || !hasInner || outerMapping.Role == innerMapping.Role || outerMapping.Type.Kind != algebra.TypeEntity || innerMapping.Type.Kind != algebra.TypeEntity || !algebra.TypeEqual(outerMapping.Type, portTypes[plan.EventOuterPort]) || !algebra.TypeEqual(innerMapping.Type, portTypes[plan.EventInnerPort]) {
			bad = true
		}
		for _, port := range signature.Ports {
			_, mapped := mapping[port.Role]
			if port.Role == plan.EventPort {
				if mapped || port.Type.Kind != algebra.TypeEvent {
					bad = true
				}
			} else if !mapped {
				bad = true
			}
		}
		if bad || len(mapping) != 2 {
			invalidMatch = true
			continue
		}
		candidates = append(candidates, algebraScopeLexicalMatch{Binding: binding, Sense: sense, Lexeme: lexeme, Validation: validation, OuterRole: outerMapping.Role, InnerRole: innerMapping.Role, OuterPort: plan.EventOuterPort, InnerPort: plan.EventInnerPort, EventPort: plan.EventPort})
	}
	if len(candidates) > 1 {
		return algebraScopeSelection{Status: status.OutcomeAmbiguous, Code: "AMBIGUOUS_SCOPE_BINDING"}
	}
	if len(candidates) == 1 {
		return algebraScopeSelection{Status: status.OutcomeOK, Match: &candidates[0]}
	}
	if matched || invalidMatch {
		return algebraScopeSelection{Status: status.OutcomeInvalid, Code: "INVALID_SCOPE_BINDING"}
	}
	return algebraScopeSelection{Status: status.OutcomeUnknown, Code: "MISSING_SCOPE_BINDING"}
}

func scopeBindingValidation(validations []language.TargetValidation, binding lexicon.SemanticBinding) (language.TargetValidation, bool) {
	if binding.Target == nil {
		return language.TargetValidation{}, false
	}
	return findTargetValidation(validations, binding)
}

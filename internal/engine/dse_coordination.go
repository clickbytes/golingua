package engine

import (
	"reflect"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// enrichDSECoordination independently enriches each of the two source clauses
// before restoring the checked conjunction for target realization.
func enrichDSECoordination(engine *Engine, sourcePackage *language.SemanticLanguagePackage, document algebra.Document, candidate semantic.DSECandidate) (algebra.Document, []DSEReferenceConceptMapping, status.OutcomeKind, string, string) {
	root := document.Expression
	if root == nil || root.Kind != algebra.TermAnd || len(root.Operands) != 2 || root.Operands[0] == nil || root.Operands[1] == nil || root.Operands[0].Kind != algebra.TermAtom || root.Operands[1].Kind != algebra.TermAtom || len(candidate.ClauseSources) != 2 {
		return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_COORDINATION_PROVENANCE_REQUIRED", "translation coordination requires exactly two direct Atoms and two independently analyzed clause sources"
	}
	flat := make([]semantic.DSERoleSource, 0, len(candidate.RoleSources))
	for _, clause := range candidate.ClauseSources {
		flat = append(flat, clause.RoleSources...)
	}
	if !reflect.DeepEqual(flat, candidate.RoleSources) || candidate.ClauseSources[0].HeadEventReferenceID == "" || candidate.ClauseSources[1].HeadEventReferenceID == "" || candidate.ClauseSources[0].HeadEventReferenceID == candidate.ClauseSources[1].HeadEventReferenceID {
		return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_COORDINATION_PROVENANCE_MISMATCH", "coordinated clause provenance must flatten exactly to aggregate role sources and retain two distinct event references"
	}
	working := document
	allMappings := []DSEReferenceConceptMapping{}
	for i, clause := range candidate.ClauseSources {
		atom := root.Operands[i]
		if !dseCoordinationClauseFrameMatches(sourcePackage, clause, atom) {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_COORDINATION_FRAME_MISMATCH", "each clause must match its exact approved source frame and semantic role ports"
		}
		child := candidate
		child.Source = clause.Source
		child.HeadEventReferenceID = clause.HeadEventReferenceID
		child.ClauseSources = nil
		child.RoleSources = clause.RoleSources
		working.Expression = atom
		var outcome status.OutcomeKind
		var code, message string
		var mappings []DSEReferenceConceptMapping
		working, mappings, outcome, code, message = enrichDSEContext(engine, working, child)
		if outcome != status.OutcomeOK {
			return algebra.Document{}, nil, outcome, code, message
		}
		allMappings = append(allMappings, mappings...)
	}
	working.Expression = root
	return working, allMappings, status.OutcomeOK, "", ""
}

func dseCoordinationClauseFrameMatches(sourcePackage *language.SemanticLanguagePackage, clause semantic.DSEClauseSource, atom *algebra.Term) bool {
	if sourcePackage == nil || sourcePackage.ValencyIndex() == nil || atom == nil || clause.Source.SenseID == "" || clause.Source.FrameID == "" {
		return false
	}
	var frameFound bool
	var allowedRoles map[string]bool
	for _, frame := range sourcePackage.Frames() {
		if string(frame.ID) == clause.Source.FrameID && frame.SenseID.String() == clause.Source.SenseID {
			if frameFound {
				return false
			}
			frameFound = true
			allowedRoles = map[string]bool{}
			for _, role := range frame.SemanticRoles {
				allowedRoles[string(role)] = true
			}
		}
	}
	if !frameFound {
		return false
	}
	sense, err := ids.ParseSenseID(clause.Source.SenseID)
	if err != nil {
		return false
	}
	bindings := sourcePackage.LexiconIndex().TargetBindingsForSense(sense)
	selected := 0
	ports := map[string]string{}
	for _, binding := range bindings {
		if binding.Target == nil || binding.Target.Kind != lexicon.SemanticTargetSymbol || binding.Target.SymbolRef != lexicon.SymbolRef(atom.Symbol) || !binding.TargetExecutionAuthorized() {
			continue
		}
		framePorts := map[string]string{}
		for _, port := range binding.Target.RolePorts {
			if port.FrameID != clause.Source.FrameID {
				continue
			}
			if port.Role == "" || port.PortID == "" {
				return false
			}
			if prior, exists := framePorts[port.Role]; exists && prior != port.PortID {
				return false
			}
			framePorts[port.Role] = port.PortID
		}
		complete := len(framePorts) == len(allowedRoles)
		for role := range allowedRoles {
			if framePorts[role] == "" {
				complete = false
			}
		}
		if complete {
			selected++
			ports = framePorts
		}
	}
	if selected != 1 {
		return false
	}
	rolesByPort := map[string]string{}
	for role, port := range ports {
		if prior, exists := rolesByPort[port]; exists && prior != role {
			return false
		}
		rolesByPort[port] = role
	}
	sources := map[string]semantic.DSERoleSource{}
	for _, role := range clause.RoleSources {
		if role.Role == "" || role.ReferenceID == "" || role.WHGap != nil || sources[role.Role].Role != "" {
			return false
		}
		sources[role.Role] = role
	}
	seenEvent := false
	seenRoles := map[string]bool{}
	for _, arg := range atom.Args {
		if arg.Value == nil || arg.Value.Type == nil {
			return false
		}
		if arg.Value.Type.Kind == algebra.TypeEvent {
			if seenEvent || arg.Value.Kind != algebra.TermRef || arg.Value.ID != clause.HeadEventReferenceID {
				return false
			}
			seenEvent = true
			continue
		}
		role, exists := rolesByPort[arg.Role]
		source, sourced := sources[role]
		if !exists || !sourced || seenRoles[role] || arg.Value.Type.Kind != algebra.TypeEntity || arg.Value.Kind != algebra.TermRef || arg.Value.ID != source.ReferenceID {
			return false
		}
		seenRoles[role] = true
	}
	if !seenEvent || len(seenRoles) != len(allowedRoles) || len(seenRoles) != len(sources) {
		return false
	}
	for role := range sources {
		if !allowedRoles[role] || !seenRoles[role] {
			return false
		}
	}
	return true
}

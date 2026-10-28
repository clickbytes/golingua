package engine

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
)

func scopedVariant(plan construction.ScopedClausePlan, alt ScopedTargetAlternative, req ScopedTargetRequest, subject string) (construction.ScopedClauseVariant, bool) {
	negation := string(req.NegationLocation)
	if negation != "outer" && negation != "inner" {
		return construction.ScopedClauseVariant{}, false
	}
	if subject == "" {
		return construction.ScopedClauseVariant{}, false
	}
	var found construction.ScopedClauseVariant
	count := 0
	for _, v := range plan.Variants {
		if v.ID == alt.ConstructionID.String() && v.Negation == negation && v.SubjectBinder == subject {
			found = v
			count++
		}
	}
	return found, count == 1
}

func validateScopeFrame(req ScopedTargetRequest, pack *language.SemanticLanguagePackage, plan construction.ScopedClausePlan) (string, error) {
	if pack == nil || req.EventSelection.Binding.Target == nil {
		return "", fmt.Errorf("event role frame evidence is absent")
	}
	target := req.EventSelection.Binding.Target
	frameIDs := map[string]bool{}
	ports := map[string]lexicon.SemanticRolePort{}
	if len(target.RolePorts) != 2 {
		return "", fmt.Errorf("event target must map exactly two participant ports")
	}
	for _, p := range target.RolePorts {
		if p.FrameID == "" || p.PortID == "" || p.Role == "" {
			return "", fmt.Errorf("event target contains incomplete frame mapping")
		}
		if _, exists := ports[p.PortID]; exists {
			return "", fmt.Errorf("event target duplicates a participant port")
		}
		frameIDs[p.FrameID] = true
		ports[p.PortID] = p
	}
	x, xok := ports[req.Roles.XPort]
	y, yok := ports[req.Roles.YPort]
	if !xok || !yok || x.FrameID != y.FrameID || x.Role != req.Roles.XRole || y.Role != req.Roles.YRole || x.Role == y.Role {
		return "", fmt.Errorf("X and Y do not map to distinct roles in one event frame")
	}
	matched := 0
	subjectBinder := ""
	for _, frame := range pack.Frames() {
		if !frameIDs[string(frame.ID)] || frame.SenseID != req.EventSelection.Sense.ID {
			continue
		}
		roles := map[string]bool{}
		objectRoles := map[string]bool{}
		objectCase := ""
		for _, arg := range frame.Arguments {
			if arg.GrammaticalFunction == "subject" {
				roles[string(arg.Role)] = true
			}
			if arg.GrammaticalFunction == "object" {
				objectRoles[string(arg.Role)] = true
				objectCase = arg.Case
			}
		}
		if len(roles) != 1 || len(objectRoles) != 1 {
			return "", fmt.Errorf("frame must have exactly one subject and one object role")
		}
		binder := ""
		if roles[req.Roles.XRole] && objectRoles[req.Roles.YRole] {
			matched++
			binder = "x"
		} else if roles[req.Roles.YRole] && objectRoles[req.Roles.XRole] {
			matched++
			binder = "y"
		}
		if binder != "" {
			variants := 0
			for _, variant := range plan.Variants {
				if variant.Negation == string(req.NegationLocation) && variant.SubjectBinder == binder && variant.FrameObjectCase != "" {
					variants++
					if objectCase != variant.FrameObjectCase {
						return "", fmt.Errorf("frame object case does not match selected construction variant")
					}
				}
			}
			if variants > 1 {
				return "", fmt.Errorf("multiple frame object case constraints match selected scope")
			}
			subjectBinder = binder
		}
	}
	if matched != 1 || subjectBinder == "" {
		return "", fmt.Errorf("event selection does not resolve to exactly one matching frame subject")
	}
	return subjectBinder, nil
}

func validateScopeVariantParts(variant construction.ScopedClauseVariant, actual []ScopedSurfacePart, req ScopedTargetRequest) error {
	if len(variant.Parts) != len(actual) {
		return fmt.Errorf("surface part count differs from construction plan")
	}
	for i, want := range variant.Parts {
		got := actual[i]
		if want.Literal != "" {
			if got.Text != want.Literal || got.Generated != nil || got.BinderID != "" || got.Port != "" {
				return fmt.Errorf("literal part %d differs from the exact plan", i)
			}
			continue
		}
		binder, selection := req.BinderEvent, req.EventSelection
		switch want.Slot {
		case "x":
			binder, selection = req.BinderX, req.XSelection
		case "y":
			binder, selection = req.BinderY, req.YSelection
		case "event":
		default:
			return fmt.Errorf("plan contains unknown slot")
		}
		if got.BinderID != binder || got.Port != selection.Port || got.Generated == nil || got.Generated.LexemeID != selection.Lexeme.ID.String() || got.Generated.Surface != got.Text {
			return fmt.Errorf("slot part %d lacks exact binder, port, lexeme, or generated form", i)
		}
		if want.Slot == "event" && req.EventTemporalFrame != "" && req.EventTemporalFrame != "past" && req.EventTemporalFrame != "present" {
			return fmt.Errorf("requested event temporal frame is unsupported")
		}
		if want.Case != "" {
			value, present, valid := generatedKnownFeature(got, "case")
			if !valid {
				return fmt.Errorf("generated noun case is duplicated or unresolved")
			}
			if variant.CasePolicy == "marked" && (!present || value != want.Case) {
				return fmt.Errorf("marked-case construction requires the exact known planned noun case")
			}
			if variant.CasePolicy == "unmarked" && present && value != want.Case {
				return fmt.Errorf("generated noun case conflicts with planned case")
			}
		}
	}
	for _, constraint := range variant.Constraints {
		if constraint.Feature == "temporal_frame" && constraint.Binder == "event" {
			if !containsScopeValue(constraint.Values, req.EventTemporalFrame) {
				return fmt.Errorf("event temporal frame violates construction constraint")
			}
			continue
		}
		matched := 0
		for i := range variant.Parts {
			if variant.Parts[i].Slot == constraint.Binder {
				matched++
				if !featureAllowed(actual[i], constraint.Feature, constraint.Values) {
					return fmt.Errorf("generated slot occurrence does not satisfy feature constraint")
				}
			}
		}
		if matched == 0 {
			return fmt.Errorf("constrained slot is absent")
		}
	}
	return nil
}

func generatedFeatureMatches(part ScopedSurfacePart, feature, value string) bool {
	actual, present, valid := generatedKnownFeature(part, feature)
	return valid && present && actual == value
}

func generatedKnownFeature(part ScopedSurfacePart, feature string) (string, bool, bool) {
	if part.Generated == nil {
		return "", false, false
	}
	value, found := "", false
	for _, f := range part.Generated.Features {
		if string(f.Feature) == feature {
			if found || string(f.State) != "known" || f.Value == "" || f.Variable != nil {
				return "", true, false
			}
			value, found = string(f.Value), true
		}
	}
	return value, found, true
}

func featureAllowed(part ScopedSurfacePart, feature string, values []string) bool {
	for _, value := range values {
		if generatedFeatureMatches(part, feature, value) {
			return true
		}
	}
	return false
}

func containsScopeValue(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value && value != "" {
			return true
		}
	}
	return false
}

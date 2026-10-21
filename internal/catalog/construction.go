package catalog

import (
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// MaxResolvedPlans bounds a single concept lookup independently of catalog size.
const MaxResolvedPlans = 256

// PlanPath records an authorized route through semantic and construction data.
// It is an index result, not a generated sentence.
type PlanPath struct {
	ConceptID ids.ConceptID                         `json:"concept_id"`
	SenseID   ids.SenseID                           `json:"sense_id"`
	FrameID   valency.FrameID                       `json:"frame_id"`
	Plan      construction.CompiledConstructionPlan `json:"plan"`
}

type PlanResolution struct {
	Status      status.OutcomeKind  `json:"status"`
	Paths       []PlanPath          `json:"paths"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// ResolveApprovedPlans walks Concept→Sense→Frame→ConstructionPlan using only
// approved equivalent lexical links and approved construction bindings.
// An allowed family in a frame is not itself a binding, and a compiled plan
// is not a license for every sense sharing its structure.
func ResolveApprovedPlans(graph *core.ConceptGraph, words *lexicon.Index, frames *valency.Index, plans *construction.PlanIndex, conceptID ids.ConceptID) PlanResolution {
	result := PlanResolution{Status: status.OutcomeUnsupported, Paths: []PlanPath{}}
	if graph == nil || words == nil || frames == nil || plans == nil {
		return planDiagnostic(status.OutcomeInvalid, "MISSING_CATALOG_INDEX", "all compiled catalog indexes are required")
	}
	parsed, err := ids.ParseConceptID(conceptID.String())
	if err != nil || parsed.String() != conceptID.String() {
		return planDiagnostic(status.OutcomeInvalid, "INVALID_CONCEPT_ID", "concept ID must be canonical")
	}
	if _, exists := graph.Lookup(conceptID); !exists {
		return planDiagnostic(status.OutcomeUnknown, "UNKNOWN_CONCEPT", "concept is not present in the selected CORE snapshot")
	}
	approvedSenses := make(map[string]ids.SenseID)
	for _, binding := range words.ByCoreConcept(conceptID) {
		if binding.TransferAuthorized() {
			approvedSenses[binding.SenseID.String()] = binding.SenseID
		}
	}
	if len(approvedSenses) == 0 {
		return planDiagnostic(status.OutcomeUnsupported, "NO_APPROVED_SENSE", "concept has no approved equivalent lexical sense")
	}
	approvedConstructions := make(map[string]struct{})
	for _, binding := range frames.Bindings() {
		if _, exists := approvedSenses[binding.SenseID.String()]; !exists {
			continue
		}
		key := binding.SenseID.String() + "\x00" + string(binding.ValencyFrameID) + "\x00" + binding.FamilyID.String() + "\x00" + binding.VariantID.String()
		approvedConstructions[key] = struct{}{}
	}
	if len(approvedConstructions) == 0 {
		return planDiagnostic(status.OutcomeUnsupported, "NO_APPROVED_CONSTRUCTION", "approved senses have no approved construction bindings")
	}
	for _, plan := range plans.Plans() {
		if !plan.ValidateIdentity() {
			return planDiagnostic(status.OutcomeInvalid, "INVALID_CONSTRUCTION_PLAN", "compiled plan identity is invalid")
		}
		for _, permission := range plan.Permissions {
			if _, exists := approvedSenses[permission.SenseID.String()]; !exists {
				continue
			}
			key := permission.SenseID.String() + "\x00" + string(permission.FrameID) + "\x00" + plan.FamilyID.String() + "\x00" + string(plan.VariantID)
			if _, exists := approvedConstructions[key]; !exists {
				continue
			}
			if len(result.Paths) == MaxResolvedPlans {
				return planDiagnostic(status.OutcomeResourceLimit, "PLAN_PATH_LIMIT", "concept resolves to too many construction plans")
			}
			result.Paths = append(result.Paths, PlanPath{ConceptID: conceptID, SenseID: permission.SenseID, FrameID: permission.FrameID, Plan: plan})
		}
	}
	if len(result.Paths) == 0 {
		return planDiagnostic(status.OutcomeUnsupported, "NO_COMPILED_CONSTRUCTION_PLAN", "no compiled plan matches the approved sense, frame, family and variant")
	}
	sort.Slice(result.Paths, func(i, j int) bool {
		left, right := result.Paths[i], result.Paths[j]
		if left.SenseID.String() != right.SenseID.String() {
			return left.SenseID.String() < right.SenseID.String()
		}
		if left.FrameID != right.FrameID {
			return left.FrameID < right.FrameID
		}
		return left.Plan.ID < right.Plan.ID
	})
	result.Status = status.OutcomeOK
	return result
}

func planDiagnostic(kind status.OutcomeKind, code, message string) PlanResolution {
	return PlanResolution{Status: kind, Paths: []PlanPath{}, Diagnostics: []status.Diagnostic{{Code: code, Message: message, Path: "catalog.construction", Severity: status.SeverityError}}}
}

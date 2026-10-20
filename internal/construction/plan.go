package construction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

type CompiledArgumentBinding struct {
	ArgumentID string `json:"argument_id"`
	Role       string `json:"role"`
	Required   bool   `json:"required"`
	SlotID     string `json:"slot_id"`
	CasePolicy string `json:"case_policy,omitempty"`
	Adposition string `json:"adposition,omitempty"`
}

type SensePermission struct {
	SenseID ids.SenseID     `json:"sense_id"`
	FrameID valency.FrameID `json:"frame_id"`
}

type CompiledConstructionPlan struct {
	ID                ConstructionPlanID        `json:"id"`
	Language          string                    `json:"language"`
	ProfileVersion    features.ProfileVersion   `json:"profile_version"`
	ProfileProvenance provenance.SourceRef      `json:"profile_provenance"`
	Permissions       []SensePermission         `json:"sense_permissions"`
	FamilyID          ids.ConstructionID        `json:"family_id"`
	VariantID         VariantID                 `json:"variant_id"`
	ArgumentBindings  []CompiledArgumentBinding `json:"argument_bindings"`
	Constraints       []CompiledConstraint      `json:"constraints"`
	Linearization     []SequenceToken           `json:"linearization"`
	ProcedureIDs      []string                  `json:"procedure_ids,omitempty"`
}

func canonicalPlanID(plan CompiledConstructionPlan) ConstructionPlanID {
	canonical := plan
	canonical.ID = ""
	canonical.Permissions = nil
	canonical.ArgumentBindings = append([]CompiledArgumentBinding(nil), plan.ArgumentBindings...)
	sort.Slice(canonical.ArgumentBindings, func(i, j int) bool {
		return canonical.ArgumentBindings[i].ArgumentID < canonical.ArgumentBindings[j].ArgumentID
	})
	canonical.Constraints = append([]CompiledConstraint(nil), plan.Constraints...)
	for i := range canonical.Constraints {
		canonical.Constraints[i].Features = append([]features.FeatureID(nil), canonical.Constraints[i].Features...)
		sort.Slice(canonical.Constraints[i].Features, func(a, b int) bool {
			return canonical.Constraints[i].Features[a] < canonical.Constraints[i].Features[b]
		})
	}
	sort.Slice(canonical.Constraints, func(i, j int) bool { return canonical.Constraints[i].ID < canonical.Constraints[j].ID })
	canonical.ProcedureIDs = append([]string(nil), plan.ProcedureIDs...)
	sort.Strings(canonical.ProcedureIDs)
	data, _ := json.Marshal(canonical)
	digest := sha256.Sum256(data)
	return ConstructionPlanID("plan:" + hex.EncodeToString(digest[:]))
}

func normalizeProcedureID(value string) string { return value }

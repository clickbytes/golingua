package runtime

import (
	"sort"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// SentenceAnalysisCandidates exposes only successful, exact predicate routes
// from a source result to the typed language bridge. Participant records are
// intentionally ignored because they are not independent sentence analyses.
func SentenceAnalysisCandidates(result semantic.Result, senseIDs ...ids.SenseID) []language.AnalysisCandidate {
	allowedSenses := make(map[ids.SenseID]bool, len(senseIDs))
	for _, senseID := range senseIDs {
		if senseID.String() != "" {
			allowedSenses[senseID] = true
		}
	}
	if len(allowedSenses) == 0 {
		return []language.AnalysisCandidate{}
	}
	type route struct {
		sense   ids.SenseID
		core    ids.ConceptID
		frame   string
		family  ids.ConstructionID
		variant ids.ConstructionID
	}
	unique := make(map[route]struct{})
	for _, candidate := range result.Candidates {
		if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
			continue
		}
		for _, record := range candidate.Evidence.Records {
			if record.Outcome != status.OutcomeOK || !allowedSenses[record.SenseID] || record.CoreConceptID.String() == "" ||
				string(record.FrameID) == "" || record.FamilyID.String() == "" || string(record.VariantID) == "" {
				continue
			}
			variantID, err := ids.ParseConstructionID(string(record.VariantID))
			if err != nil {
				continue
			}
			unique[route{sense: record.SenseID, core: record.CoreConceptID, frame: string(record.FrameID), family: record.FamilyID, variant: variantID}] = struct{}{}
		}
	}
	routes := make([]route, 0, len(unique))
	for value := range unique {
		routes = append(routes, value)
	}
	sort.Slice(routes, func(i, j int) bool {
		left, right := routes[i], routes[j]
		if left.sense != right.sense {
			return left.sense.String() < right.sense.String()
		}
		if left.core != right.core {
			return left.core.String() < right.core.String()
		}
		if left.frame != right.frame {
			return left.frame < right.frame
		}
		if left.family != right.family {
			return left.family.String() < right.family.String()
		}
		return left.variant.String() < right.variant.String()
	})
	resultCandidates := make([]language.AnalysisCandidate, 0, len(routes))
	for _, value := range routes {
		resultCandidates = append(resultCandidates, language.AnalysisCandidate{
			SenseID: value.sense, CoreRef: value.core, ValencyFrameID: value.frame,
			FamilyID: value.family, VariantID: value.variant,
		})
	}
	return resultCandidates
}

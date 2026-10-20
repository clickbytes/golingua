package analyzer

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/status"
	textpkg "github.com/clickbytes/golingua/internal/text"
)

// matchMWEs appends every flat fixed/inflected MWE match while retaining all
// component reading alternatives. It never removes token-level analyses.
func matchMWEs(lattice *LexicalLattice, plans []construction.CompiledMWEPlan, lexical *lexicon.Index, budget Budget, used *int) error {
	if lattice == nil {
		return newLatticeSetupError("lattice", "lattice is required")
	}
	if err := budget.Validate(); err != nil {
		return err
	}
	if plans == nil {
		markMWEIncomplete(lattice, "MWE_INDEX_UNAVAILABLE", "mwe_plans", "MWE inventory is unavailable")
		return nil
	}
	if len(plans) == 0 {
		return nil
	}
	if used == nil || *used < 0 || *used > budget.MaxCandidates {
		return newLatticeSetupError("budget.candidates", "shared candidate count is outside the configured budget")
	}
	if lexical == nil {
		return latticeDataError(status.OutcomeUnknown, "MWE_LEXICON_UNAVAILABLE", "lexicon", "MWE matching requires an explicit lexical snapshot")
	}

	planLimit := len(plans)
	if planLimit > budget.MaxMWEPlans {
		planLimit = budget.MaxMWEPlans
		id := boundedID(plans[planLimit].Plan.ID)
		path := "mwe_plans[" + fmt.Sprint(planLimit) + "]"
		if id != "" {
			path += "." + id
		}
		markMWEIncomplete(lattice, "MWE_PLAN_BUDGET_EXCEEDED", path, "remaining MWE plans were not examined because the plan budget was exhausted")
	}
	// Bound copy and sort work before allocating; plans beyond this prefix are
	// omitted by the explicit plan budget.
	orderedPlans := append([]construction.CompiledMWEPlan(nil), plans[:planLimit]...)
	sort.SliceStable(orderedPlans, func(i, j int) bool {
		return orderedPlans[i].Plan.ID < orderedPlans[j].Plan.ID
	})
	for index := 1; index < planLimit; index++ {
		if orderedPlans[index-1].Plan.ID == orderedPlans[index].Plan.ID {
			return latticeDataError(status.OutcomeInvalid, "DUPLICATE_MWE_PLAN_ID", mwePlanPath(index, orderedPlans[index].Plan.ID), "MWE plan IDs must be unique")
		}
	}

	for planIndex := 0; planIndex < planLimit; planIndex++ {
		plan := orderedPlans[planIndex].Plan
		if err := validateMWEPlanForMatching(plan, lattice.Language, lexical); err != nil {
			return err
		}
		if len(plan.Components) < 2 || len(plan.Components) > construction.MaxSequenceTokens {
			return latticeDataError(status.OutcomeInvalid, "INVALID_MWE_COMPONENT_COUNT", mwePlanPath(planIndex, plan.ID)+".components", "compiled MWE component count is outside the supported range")
		}
		if err := validateMWEComponents(plan, planIndex); err != nil {
			return err
		}
		if hasMWEInput(plan) {
			markMWEIncomplete(lattice, "MWE_INPUT_REQUIRES_PARSING", mwePlanPath(planIndex, plan.ID), "MWEInput matching requires phrase parsing and was not evaluated")
			continue
		}

		for start := 0; start+len(plan.Components) <= len(lattice.Tokens); start++ {
			components, matched := matchMWEAt(lattice, plan, start)
			if !matched {
				continue
			}
			addition := 1 // the MWE edge itself
			for _, component := range components {
				addition += len(component.ReadingIDs)
			}
			if addition > budget.MaxCandidates-*used {
				span := lattice.Tokens[start].Token.Span
				end := lattice.Tokens[start+len(plan.Components)-1].Token.Span.End
				path := fmt.Sprintf("%s.tokens[%d:%d].span[%d:%d]", mwePlanPath(planIndex, plan.ID), start, start+len(plan.Components), span.Start, end)
				markMWEIncomplete(lattice, "MWE_CANDIDATE_BUDGET_EXCEEDED", path, "matching MWE was not retained because the candidate budget was exhausted")
				return nil
			}
			first := lattice.Tokens[start].Token.Span
			last := lattice.Tokens[start+len(plan.Components)-1].Token.Span
			lattice.MWEs = append(lattice.MWEs, MWEEdge{
				PlanID:     plan.ID,
				LexemeID:   plan.LexemeID,
				SenseID:    plan.SenseID,
				TokenStart: start,
				TokenEnd:   start + len(plan.Components),
				Span:       textpkg.Span{Start: first.Start, End: last.End},
				Components: components,
			})
			*used += addition
		}
	}
	return nil
}

func validateMWEComponents(plan construction.MWEPlan, planIndex int) error {
	seen := make(map[string]struct{}, len(plan.Components))
	for index, component := range plan.Components {
		path := fmt.Sprintf("%s.components[%d]", mwePlanPath(planIndex, plan.ID), index)
		if component.Position != index || component.ID == "" {
			return latticeDataError(status.OutcomeInvalid, "INVALID_MWE_COMPONENT_ORDER", path, "compiled MWE component positions must be contiguous and IDs nonempty")
		}
		if _, exists := seen[component.ID]; exists {
			return latticeDataError(status.OutcomeInvalid, "DUPLICATE_MWE_COMPONENT_ID", path, "compiled MWE component IDs must be unique")
		}
		seen[component.ID] = struct{}{}
		switch component.Kind {
		case construction.MWEFixed:
			if len(component.Features.Values) != 0 {
				return latticeDataError(status.OutcomeInvalid, "INVALID_FIXED_MWE_COMPONENT", path, "fixed MWE component cannot constrain features")
			}
		case construction.MWEInflected:
			if len(component.Features.Values) == 0 || component.Features.Language != plan.Language {
				return latticeDataError(status.OutcomeInvalid, "INVALID_INFLECTED_MWE_COMPONENT", path, "inflected MWE component requires language-scoped feature constraints")
			}
		default:
			if component.Kind != construction.MWEInput {
				return latticeDataError(status.OutcomeInvalid, "INVALID_MWE_COMPONENT_KIND", path, "compiled MWE component kind is unsupported")
			}
		}
	}
	return nil
}

func validateMWEPlanForMatching(plan construction.MWEPlan, language features.LanguageID, lexical *lexicon.Index) error {
	if plan.ID == "" || plan.Language != language {
		return latticeDataError(status.OutcomeUnsupported, "MWE_LANGUAGE_MISMATCH", "mwe_plans."+boundedID(plan.ID), "MWE plan language or identity does not match the lattice")
	}
	whole, found := lexical.ByLexemeID(plan.LexemeID)
	if !found {
		return latticeDataError(status.OutcomeUnknown, "UNKNOWN_MWE_LEXEME", "mwe_plans."+boundedID(plan.ID)+".lexeme_id", "MWE lexeme is absent from the lexical snapshot")
	}
	if features.LanguageID(whole.Language) != language {
		return latticeDataError(status.OutcomeUnsupported, "MWE_LEXEME_LANGUAGE_MISMATCH", "mwe_plans."+boundedID(plan.ID)+".lexeme_id", "MWE lexeme language does not match the lattice")
	}
	senseFound := false
	for _, sense := range lexical.SensesForLexeme(plan.LexemeID) {
		if sense.ID == plan.SenseID {
			senseFound = true
			break
		}
	}
	if !senseFound {
		return latticeDataError(status.OutcomeUnknown, "UNKNOWN_MWE_SENSE", "mwe_plans."+boundedID(plan.ID)+".sense_id", "MWE sense is not attached to its lexical entry")
	}
	for _, component := range plan.Components {
		if component.Kind != construction.MWEFixed && component.Kind != construction.MWEInflected && component.Kind != construction.MWEInput {
			return latticeDataError(status.OutcomeInvalid, "INVALID_MWE_COMPONENT_KIND", "mwe_plans."+boundedID(plan.ID)+".components", "compiled MWE contains an unsupported component kind")
		}
		if component.Kind == construction.MWEInput {
			continue
		}
		lexeme, ok := lexical.ByLexemeID(component.LexemeID)
		if !ok {
			return latticeDataError(status.OutcomeUnknown, "UNKNOWN_MWE_COMPONENT_LEXEME", "mwe_plans."+boundedID(plan.ID)+".components."+boundedID(component.ID), "MWE component lexeme is absent from the lexical snapshot")
		}
		if features.LanguageID(lexeme.Language) != language {
			return latticeDataError(status.OutcomeUnsupported, "MWE_COMPONENT_LANGUAGE_MISMATCH", "mwe_plans."+boundedID(plan.ID)+".components."+boundedID(component.ID), "MWE component language does not match the lattice")
		}
	}
	return nil
}

func hasMWEInput(plan construction.MWEPlan) bool {
	for _, component := range plan.Components {
		if component.Kind == construction.MWEInput {
			return true
		}
	}
	return false
}

func matchMWEAt(lattice *LexicalLattice, plan construction.MWEPlan, start int) ([]MWEComponentMatch, bool) {
	matches := make([]MWEComponentMatch, 0, len(plan.Components))
	for componentIndex, component := range plan.Components {
		tokenIndex := start + componentIndex
		node := lattice.Tokens[tokenIndex]
		if node.Token.Kind != textpkg.TokenWord {
			return nil, false
		}
		readingIDs := make([]string, 0)
		for _, candidate := range node.Readings {
			if candidate.Reading.LexemeID != component.LexemeID.String() {
				continue
			}
			if component.Kind == construction.MWEInflected && !featureSubsetMatches(component.Features, candidate.Reading.Features) {
				continue
			}
			readingIDs = append(readingIDs, candidate.Reading.ID)
		}
		if len(readingIDs) == 0 {
			return nil, false
		}
		matches = append(matches, MWEComponentMatch{
			ComponentID: component.ID,
			TokenIndex:  tokenIndex,
			ReadingIDs:  readingIDs,
		})
	}
	return matches, true
}

func featureSubsetMatches(required, actual features.FeatureBundle) bool {
	if required.Language != actual.Language || required.ProfileVersion != actual.ProfileVersion {
		return false
	}
	for _, wanted := range required.Values {
		found := false
		for _, present := range actual.Values {
			if sameFeatureValue(wanted, present) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func sameFeatureValue(left, right features.FeatureValue) bool {
	if left.Feature != right.Feature || left.State != right.State || left.Value != right.Value {
		return false
	}
	if left.Variable == nil || right.Variable == nil {
		return left.Variable == nil && right.Variable == nil
	}
	return left.Variable.ID == right.Variable.ID && left.Variable.Scope == right.Variable.Scope
}

func markMWEIncomplete(lattice *LexicalLattice, code, path, message string) {
	lattice.Outcome = status.OutcomeIncomplete
	for _, existing := range lattice.Diagnostics {
		if existing.Code == code && existing.Path == path {
			return
		}
	}
	diagnostic := status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning}
	if len(lattice.Diagnostics) < status.MaxDiagnostics {
		lattice.Diagnostics = append(lattice.Diagnostics, diagnostic)
		return
	}
	// Keep a bounded marker if the diagnostic budget was already exhausted.
	last := &lattice.Diagnostics[status.MaxDiagnostics-1]
	if last.Code != "ADDITIONAL_MWE_INCOMPLETENESS" {
		*last = status.Diagnostic{
			Code:     "ADDITIONAL_MWE_INCOMPLETENESS",
			Message:  "additional MWE matches or plans were not evaluated; lattice is incomplete",
			Path:     "mwe_plans",
			Severity: status.SeverityWarning,
		}
	}
}

func latticeDataError(kind status.OutcomeKind, code, path, message string) error {
	return NewLatticeError(status.Result{
		Status: kind,
		Diagnostics: []status.Diagnostic{{
			Code:     code,
			Message:  message,
			Path:     path,
			Severity: status.SeverityError,
		}},
	})
}

func mwePlanPath(index int, planID string) string {
	return "mwe_plans[" + fmt.Sprint(index) + "]." + boundedID(planID)
}

func boundedID(value string) string {
	if len(value) <= 128 && utf8.ValidString(value) {
		return value
	}
	return "<invalid-id>"
}

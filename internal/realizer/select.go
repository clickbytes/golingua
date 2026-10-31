// Package realizer selects licensed target-language lexical and construction
// alternatives for Meaning graphs. It does not generate surface text.
package realizer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultMaxTargetCandidates = 4096
	maxTargetCandidates        = 16384
	defaultMaxTargetExpansions = 16384
	maxTargetExpansions        = 65536
)

// Budget bounds Meaning traversal, candidate retention, and inspected
// lexical/license records while selecting targets.
type Budget struct {
	MaxCandidates int
	MaxExpansions int
}

// DefaultBudget returns conservative realization-selection limits.
func DefaultBudget() Budget {
	return Budget{MaxCandidates: defaultMaxTargetCandidates, MaxExpansions: defaultMaxTargetExpansions}
}

func (budget Budget) Validate() error {
	if budget.MaxCandidates <= 0 || budget.MaxCandidates > maxTargetCandidates || budget.MaxExpansions <= 0 || budget.MaxExpansions > maxTargetExpansions {
		return selectionError(status.OutcomeInvalid, "INVALID_REALIZER_BUDGET", "budget", "realizer limits must be positive and within package ceilings")
	}
	return nil
}

// Input contains semantic alternatives and the immutable indexes that license
// target lexical and construction choices. Variants are indexed per family.
type Input struct {
	Candidates               []semantic.MeaningCandidate
	Outcome                  status.OutcomeKind
	Diagnostics              []status.Diagnostic
	CandidatePolicy          TargetCandidatePolicy
	Operators                []construction.Operator
	OperatorCoverageComplete bool
	Language                 lexicon.LanguageID
	Lexicon                  *lexicon.Index
	Valency                  *valency.Index
	Constructions            *construction.Registry
	Variants                 map[ids.ConstructionID]*construction.VariantIndex
	Core                     *core.ConceptGraph
	Budget                   Budget
}

// TargetCandidatePolicy may retain or omit an already licensed candidate
// after considering the complete candidate set for its source Meaning. It
// cannot add or modify lexical or construction licenses.
type TargetCandidatePolicy interface {
	KeepTargetCandidate(semantic.MeaningCandidate, TargetCandidate, []TargetCandidate) bool
}

// TargetCandidate is one exact approved target-side lexical and construction
// license for a Meaning concept node. It contains no generated surface text.
type TargetCandidate struct {
	SourceCandidateID    string
	SemanticID           string
	NodeID               meaning.NodeID
	Kind                 meaning.NodeKind
	QuestionAnswer       *semantic.QuestionAnswer `json:"question_answer,omitempty"`
	Roles                []meaning.Role
	Outcome              status.OutcomeKind
	Diagnostics          []status.Diagnostic
	Lexeme               lexicon.MorphLexeme
	Sense                lexicon.LexicalSense
	TransferBinding      lexicon.SemanticBinding
	Frame                valency.ValencyFrame
	ConstructionBinding  valency.ConstructionBinding
	Family               construction.FamilyRecord
	CompiledVariant      construction.CompiledVariant
	ConstructionPlan     construction.ConstructionPlan
	SoftConstraintMisses []valency.SoftConstraintMiss
	MissingOmittedRoles  []meaning.Role
}

// Result preserves every retained target alternative and the combined status.
type Result struct {
	Outcome     status.OutcomeKind  `json:"outcome"`
	Candidates  []TargetCandidate   `json:"candidates"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// SelectionError reports invalid setup or inputs as a typed status result.
type SelectionError struct{ result status.Result }

func (err *SelectionError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "target selection error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns the typed error outcome with a defensive diagnostics copy.
func (err *SelectionError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

// SelectTargets enumerates all bounded target lexical/frame/construction
// choices licensed by approved equivalent CORE transfers. It never chooses a
// single alternative by score or insertion order and never generates text.
func SelectTargets(input Input) (Result, error) {
	budget := input.Budget
	if budget == (Budget{}) {
		budget = DefaultBudget()
	}
	if err := budget.Validate(); err != nil {
		return Result{}, err
	}
	if !validLanguage(input.Language) || input.Lexicon == nil || input.Valency == nil || input.Constructions == nil || input.Core == nil {
		return Result{}, selectionError(status.OutcomeInvalid, "MISSING_REALIZATION_INDEX", "indexes", "target language, lexicon, valency, construction and CORE indexes are required")
	}
	if !input.Outcome.Valid() {
		return Result{}, selectionError(status.OutcomeInvalid, "INVALID_SEMANTIC_OUTCOME", "outcome", "semantic input outcome is unsupported")
	}
	if len(input.Candidates) > budget.MaxCandidates || len(input.Diagnostics) > status.MaxDiagnostics || len(input.Operators) > maxTargetExpansions || len(input.Operators) > budget.MaxExpansions {
		return Result{}, selectionError(status.OutcomeResourceLimit, "REALIZER_INPUT_LIMIT", "input", "semantic candidates, diagnostics, or operators exceed package limits")
	}
	if input.Outcome != status.OutcomeOK && len(input.Diagnostics) == 0 {
		return Result{}, selectionError(status.OutcomeInvalid, "MISSING_SEMANTIC_RESULT_DIAGNOSTIC", "diagnostics", "non-ok semantic input outcome requires at least one diagnostic")
	}
	for index, diagnostic := range input.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return Result{}, selectionError(status.OutcomeInvalid, "INVALID_SEMANTIC_DIAGNOSTIC", fmt.Sprintf("diagnostics[%d]", index), err.Error())
		}
	}
	for index, operator := range input.Operators {
		if err := operator.Validate(); err != nil {
			return Result{}, selectionError(outcomeFromError(err), "INVALID_OPERATOR_EVIDENCE", fmt.Sprintf("operators[%d]", index), err.Error())
		}
	}
	totalMeaningNodes := 0
	totalMeaningWork := 0
	for _, candidate := range input.Candidates {
		if candidate.Meaning == nil {
			continue
		}
		if len(candidate.Meaning.Nodes) > budget.MaxExpansions-totalMeaningNodes {
			return Result{}, selectionError(status.OutcomeResourceLimit, "MEANING_INVENTORY_LIMIT", "candidates", "aggregate Meaning nodes exceed the target selection budget")
		}
		totalMeaningNodes += len(candidate.Meaning.Nodes)
		graphWork := meaningGraphWork(candidate.Meaning.Nodes)
		if graphWork > budget.MaxExpansions-totalMeaningWork {
			return Result{}, selectionError(status.OutcomeResourceLimit, "MEANING_WORK_LIMIT", "candidates", "aggregate Meaning node and edge work exceeds the target selection budget")
		}
		totalMeaningWork += graphWork
	}
	if len(input.Operators) > budget.MaxExpansions-totalMeaningWork {
		return Result{}, selectionError(status.OutcomeResourceLimit, "OPERATOR_WORK_LIMIT", "operators", "operator evidence exceeds the remaining target selection budget")
	}
	totalMeaningWork += len(input.Operators)

	lexemeCount, senseCount, bindingCount := input.Lexicon.Len()
	if lexemeCount > lexicon.MaxLexemes || senseCount > lexicon.MaxSenses || bindingCount > lexicon.MaxBindings {
		return Result{}, selectionError(status.OutcomeResourceLimit, "LEXICON_INDEX_LIMIT", "lexicon", "lexicon inventory exceeds the supported selection bound")
	}
	ordered := append([]semantic.MeaningCandidate(nil), input.Candidates...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].SemanticID != ordered[j].SemanticID {
			return ordered[i].SemanticID < ordered[j].SemanticID
		}
		return ordered[i].ID < ordered[j].ID
	})
	result := Result{Outcome: input.Outcome, Candidates: []TargetCandidate{}, Diagnostics: append([]status.Diagnostic(nil), input.Diagnostics...)}
	finish := func() (Result, error) {
		// Keep each candidate's status and diagnostics local to the node and
		// route that produced it. Result.Outcome aggregates all candidates and
		// may be ambiguous solely because multiple valid routes remain; copying
		// that aggregate back onto every candidate would make independent role
		// referents unusable during realization.
		return result, nil
	}
	if !input.OperatorCoverageComplete {
		result.Outcome = mergeOutcome(result.Outcome, status.OutcomeIncomplete)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, "OPERATOR_COVERAGE_UNKNOWN", "operator evidence coverage was not declared complete", "operators")
	}
	if len(ordered) == 0 {
		result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnknown)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, "NO_SEMANTIC_CANDIDATES", "semantic input contains no meaning candidates", "candidates")
		return finish()
	}
	work := totalMeaningWork
	nodeCounts := make(map[string]int)
	type frameRejection struct {
		diagnostic status.Diagnostic
	}
	deferredFrameRejections := make(map[string][]frameRejection)
	deferredFrameOutcomes := make(map[string]status.OutcomeKind)
	seenCandidateIDs := make(map[string]struct{}, len(ordered))
	for candidateIndex, source := range ordered {
		if !boundedSourceID(source.ID) || !boundedSourceID(source.SemanticID) {
			return Result{}, selectionError(status.OutcomeInvalid, "INVALID_SOURCE_CANDIDATE_ID", fmt.Sprintf("candidates[%d]", candidateIndex), "source candidate and semantic IDs must be non-empty bounded UTF-8")
		}
		if _, duplicate := seenCandidateIDs[source.ID]; duplicate {
			return Result{}, selectionError(status.OutcomeInvalid, "DUPLICATE_SOURCE_CANDIDATE_ID", fmt.Sprintf("candidates[%d].id", candidateIndex), "source candidate IDs must be unique")
		}
		seenCandidateIDs[source.ID] = struct{}{}
		if !source.Status.Valid() {
			return Result{}, selectionError(status.OutcomeInvalid, "INVALID_CANDIDATE_STATUS", fmt.Sprintf("candidates[%d].status", candidateIndex), "semantic candidate status is unsupported")
		}
		if err := source.Evidence.ValidateWithBudget(semantic.DefaultEvidenceBudget()); err != nil || source.Evidence.Outcome != source.Status {
			message := "semantic candidate evidence outcome must match its status"
			if err != nil {
				message = err.Error()
			}
			return Result{}, selectionError(status.OutcomeInvalid, "INVALID_SOURCE_EVIDENCE", fmt.Sprintf("candidates[%d].evidence", candidateIndex), message)
		}
		if source.Status == status.OutcomeOK && source.Meaning == nil {
			return Result{}, selectionError(status.OutcomeInvalid, "MISSING_MEANING_GRAPH", fmt.Sprintf("candidates[%d].meaning", candidateIndex), "an OK semantic candidate requires a Meaning graph")
		}
		result.Outcome = mergeOutcome(result.Outcome, source.Status)
		result.Diagnostics = boundedDiagnostics(result.Diagnostics, source.Evidence.Diagnostics)
		if source.Meaning == nil {
			continue
		}
		if err := meaning.ValidateMeaning(*source.Meaning, input.Core, meaning.DefaultBudget()); err != nil {
			return Result{}, selectionError(status.OutcomeInvalid, "INVALID_MEANING_GRAPH", fmt.Sprintf("candidates[%d].meaning", candidateIndex), err.Error())
		}
		if source.QuestionAnswer != nil {
			if err := source.QuestionAnswer.Validate(*source.Meaning, input.Core); err != nil {
				return Result{}, selectionError(status.OutcomeInvalid, "INVALID_QUESTION_ANSWER", fmt.Sprintf("candidates[%d].question_answer", candidateIndex), err.Error())
			}
			if err := validateQuestionDiscourse(*source.Meaning, *source.QuestionAnswer); err != nil {
				return Result{}, selectionError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_SCOPE", fmt.Sprintf("candidates[%d].question_answer", candidateIndex), err.Error())
			}
			if !hasExactQuestionOperator(*source.Meaning, input.Operators, *source.QuestionAnswer) {
				return Result{}, selectionError(status.OutcomeUnsupported, "QUESTION_OPERATOR_MISMATCH", fmt.Sprintf("candidates[%d].question_answer", candidateIndex), "typed WH answer requires exactly one matching WH construction operator")
			}
		} else if hasQuestionAct(source.Meaning.Nodes) {
			// Meaning v1 questions without an explicit typed answer are not
			// silently treated as the one-gap WH route.
			result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
			result.Diagnostics = appendDiagnostic(result.Diagnostics, "QUESTION_ANSWER_UNAVAILABLE", "Meaning question has no explicit typed answer role for this realization route", "meaning.discourse")
		}
		nodes := append([]meaning.Node(nil), source.Meaning.Nodes...)
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
		conceptCount := 0
		rolesByNode := make(map[meaning.NodeID][]meaning.Role)
		knownNodes := make(map[meaning.NodeID]meaning.Node, len(nodes))
		for _, node := range nodes {
			knownNodes[node.ID] = node
			if node.Concept != "" {
				conceptCount++
			}
			if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState {
				for _, participant := range node.Participants {
					rolesByNode[participant.Target] = append(rolesByNode[participant.Target], participant.Role)
				}
			}
		}
		if conceptCount == 0 {
			result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnknown)
			result.Diagnostics = appendDiagnostic(result.Diagnostics, "NO_CONCEPT_TARGET", "Meaning graph has no concept-bearing node for target selection", "meaning.nodes")
			continue
		}
		scopeWork := meaningGraphWork(nodes)
		if len(input.Operators) > budget.MaxExpansions-work {
			result.Outcome = status.OutcomeResourceLimit
			result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "operator evidence validation reached its configured bound", "budget.expansions")
			return finish()
		}
		work += len(input.Operators)
		for operatorIndex, operator := range input.Operators {
			target, exists := knownNodes[operator.SemanticScope.TargetNodeID]
			if !exists || (target.Kind != meaning.KindEvent && target.Kind != meaning.KindState) {
				return Result{}, selectionError(status.OutcomeInvalid, "INVALID_OPERATOR_TARGET", fmt.Sprintf("operators[%d].semantic_scope.target_node_id", operatorIndex), "explicit target operator must point to an event or state in each Meaning candidate")
			}
			if operator.SemanticScope.ScopeNodeID != "" {
				if _, exists := knownNodes[operator.SemanticScope.ScopeNodeID]; !exists {
					return Result{}, selectionError(status.OutcomeInvalid, "INVALID_OPERATOR_SCOPE", fmt.Sprintf("operators[%d].semantic_scope.scope_node_id", operatorIndex), "explicit target operator scope node must resolve in each Meaning candidate")
				}
			}
		}
		unsupportedDiscourse := hasUnsupportedDiscourseAct(nodes, input.Operators) || source.QuestionAnswer == nil && hasQuestionAct(nodes)
		for _, node := range nodes {
			if node.Concept == "" {
				continue
			}
			nodeKey := source.ID + "\x00" + string(node.ID)
			nodeCounts[nodeKey] = 0
			alternateQuestionGapRejected := false
			if len(input.Operators) > budget.MaxExpansions-work {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "operator target lookup reached its configured bound", "budget.expansions")
				return finish()
			}
			work += len(input.Operators)
			nodeOperators := operatorsForTarget(input.Operators, node.ID)
			if len(nodeOperators) > construction.MaxOperatorsPerPlan {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "operators for one construction plan exceed the package limit", "budget.operators")
				return finish()
			}
			nodeOutcome := mergeOutcome(input.Outcome, source.Status)
			nodeDiagnostics := boundedDiagnostics(input.Diagnostics, source.Evidence.Diagnostics)
			if !input.OperatorCoverageComplete {
				nodeOutcome = mergeOutcome(nodeOutcome, status.OutcomeIncomplete)
				nodeDiagnostics = appendDiagnostic(nodeDiagnostics, "OPERATOR_COVERAGE_UNKNOWN", "operator evidence coverage was not declared complete", "operators")
			}
			if scopeWork > budget.MaxExpansions-work {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "Meaning scope traversal reached its configured bound", "budget.expansions")
				return finish()
			}
			work += scopeWork
			if hasUnmappedScope(nodes, node.ID, nodeOperators) {
				nodeOutcome = mergeOutcome(nodeOutcome, status.OutcomeIncomplete)
				nodeDiagnostics = appendDiagnostic(nodeDiagnostics, "OPERATOR_MAPPING_UNAVAILABLE", "Meaning contains a scoped operator without an explicit construction operator mapping", "meaning."+string(node.ID))
			}
			if unsupportedDiscourse {
				nodeOutcome = mergeOutcome(nodeOutcome, status.OutcomeUnsupported)
				nodeDiagnostics = appendDiagnostic(nodeDiagnostics, "DISCOURSE_ACT_UNSUPPORTED", "non-statement discourse act has no explicit supported realization mapping", "meaning.discourse")
			}
			result.Diagnostics = boundedDiagnostics(result.Diagnostics, nodeDiagnostics)
			result.Outcome = mergeOutcome(result.Outcome, nodeOutcome)
			concept, err := ids.ParseConceptID(string(node.Concept))
			if err != nil {
				return Result{}, selectionError(status.OutcomeInvalid, "INVALID_TARGET_CONCEPT", "meaning."+string(node.ID), "target concept ID is malformed")
			}
			bindings := input.Lexicon.ByCoreConcept(concept)
			sort.Slice(bindings, func(i, j int) bool {
				if bindings[i].SenseID != bindings[j].SenseID {
					return bindings[i].SenseID.String() < bindings[j].SenseID.String()
				}
				return bindings[i].CoreRef.String() < bindings[j].CoreRef.String()
			})
			if node.Kind == meaning.KindReferent {
				licensed := false
				for _, transfer := range bindings {
					work++
					if work > budget.MaxExpansions {
						result.Outcome = status.OutcomeResourceLimit
						result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate expansion reached its configured bound", "budget.expansions")
						return finish()
					}
					if transfer.CoreRef != concept || !transfer.TransferAuthorized() {
						continue
					}
					sense, exists := input.Lexicon.BySenseID(transfer.SenseID)
					if !exists {
						return Result{}, selectionError(status.OutcomeInvalid, "UNKNOWN_TRANSFER_SENSE", "lexicon.bindings", "approved CORE transfer refers to an absent lexical sense")
					}
					lexeme, exists := input.Lexicon.ByLexemeID(sense.LexemeID)
					if !exists {
						return Result{}, selectionError(status.OutcomeInvalid, "UNKNOWN_SENSE_LEXEME", "lexicon.senses", "target sense refers to an absent lexeme")
					}
					if lexeme.Language != input.Language {
						continue
					}
					licensed = true
					if len(result.Candidates) >= budget.MaxCandidates {
						result.Outcome = status.OutcomeResourceLimit
						result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate retention reached its configured bound", "budget.candidates")
						return finish()
					}
					roles := append([]meaning.Role(nil), rolesByNode[node.ID]...)
					sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
					result.Candidates = append(result.Candidates, TargetCandidate{SourceCandidateID: source.ID, SemanticID: source.SemanticID, NodeID: node.ID, Kind: node.Kind, Roles: roles, Outcome: nodeOutcome, Diagnostics: append([]status.Diagnostic(nil), nodeDiagnostics...), Lexeme: lexeme, Sense: sense, TransferBinding: transfer})
					nodeCounts[nodeKey]++
				}
				if nodeCounts[nodeKey] == 0 && !licensed {
					result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
					result.Diagnostics = appendDiagnostic(result.Diagnostics, "NO_LICENSED_TARGET", "no approved target transfer was found for this referent", "meaning."+string(node.ID))
				}
				continue
			}
			if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState {
				result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
				result.Diagnostics = appendDiagnostic(result.Diagnostics, "UNSUPPORTED_CONCEPT_NODE", "target lexical selection supports event, state and referent nodes", "meaning."+string(node.ID))
				continue
			}
			licensed := false
			for _, transfer := range bindings {
				work++
				if work > budget.MaxExpansions {
					result.Outcome = status.OutcomeResourceLimit
					result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate expansion reached its configured bound", "budget.expansions")
					return finish()
				}
				if transfer.CoreRef != concept || !transfer.TransferAuthorized() {
					continue
				}
				sense, exists := input.Lexicon.BySenseID(transfer.SenseID)
				if !exists {
					return Result{}, selectionError(status.OutcomeInvalid, "UNKNOWN_TRANSFER_SENSE", "lexicon.bindings", "approved CORE transfer refers to an absent lexical sense")
				}
				lexeme, exists := input.Lexicon.ByLexemeID(sense.LexemeID)
				if !exists {
					return Result{}, selectionError(status.OutcomeInvalid, "UNKNOWN_SENSE_LEXEME", "lexicon.senses", "target sense refers to an absent lexeme")
				}
				if lexeme.Language != input.Language {
					continue
				}
				frames := input.Valency.FramesBySenseID(sense.ID)
				for _, frame := range frames {
					work++
					if work > budget.MaxExpansions {
						result.Outcome = status.OutcomeResourceLimit
						result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate expansion reached its configured bound", "budget.expansions")
						return finish()
					}
					for _, familyID := range frame.AllowedFamilyIDs {
						if len(nodeOperators) > 0 && nodeOperators[0].InputFamilyID != familyID {
							continue
						}
						// Query the precompiled route, not every unrelated sense in
						// this family. Charge the lookup and each inspected binding.
						work++
						if work > budget.MaxExpansions {
							result.Outcome = status.OutcomeResourceLimit
							result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "frame/family lookup reached its configured bound", "budget.expansions")
							return finish()
						}
						familyBindings := input.Valency.BindingsForFrameFamily(frame.ID, familyID)
						if len(familyBindings) > budget.MaxExpansions-work {
							result.Outcome = status.OutcomeResourceLimit
							result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "valency family binding inventory exceeds the remaining selection budget", "budget.expansions")
							return finish()
						}
						for _, binding := range familyBindings {
							work++
							if work > budget.MaxExpansions {
								result.Outcome = status.OutcomeResourceLimit
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate expansion reached its configured bound", "budget.expansions")
								return finish()
							}
							if binding.SenseID != sense.ID || binding.ValencyFrameID != frame.ID || binding.FamilyID != familyID || !binding.Approved {
								continue
							}
							work++
							if work > budget.MaxExpansions {
								result.Outcome = status.OutcomeResourceLimit
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "exact construction license lookup reached its configured bound", "budget.expansions")
								return finish()
							}
							if !input.Constructions.HasExactBinding(binding) {
								return Result{}, selectionError(status.OutcomeInvalid, "CONSTRUCTION_INDEX_MISMATCH", "constructions."+familyID.String(), "exact valency license is absent from the construction registry")
							}
							family, err := input.Constructions.FamilyFor(features.LanguageID(input.Language), familyID)
							if err != nil {
								continue
							}
							// WH-only families are not applicable to a Meaning with no
							// question act. Do not let their required answer-gap roles
							// become failures of an otherwise complete declarative target.
							// Typed questions still pass through full validation below.
							if source.QuestionAnswer == nil && !hasQuestionAct(nodes) && familySupportsWHQuestion(family.Family) {
								continue
							}
							variantIndex := input.Variants[familyID]
							if variantIndex == nil {
								result.Outcome = mergeOutcome(result.Outcome, status.OutcomeIncomplete)
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "VARIANT_INDEX_UNAVAILABLE", "exact target construction variant index is unavailable", "constructions."+familyID.String())
								continue
							}
							compiledVariant, exists := variantIndex.ByID(construction.VariantID(binding.VariantID.String()))
							if !exists || compiledVariant.FamilyID != familyID || compiledVariant.Variant.FamilyID != familyID {
								result.Outcome = mergeOutcome(result.Outcome, status.OutcomeIncomplete)
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_VARIANT_UNAVAILABLE", "exact bound target variant is unavailable", "constructions."+familyID.String())
								continue
							}
							if compiledVariant.PlanID != family.PlanID {
								result.Outcome = mergeOutcome(result.Outcome, status.OutcomeIncomplete)
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_VARIANT_INDEX_STALE", "target variant index was compiled from a different construction family plan", "constructions."+familyID.String())
								continue
							}
							family, err = construction.BindFrame(family, frame)
							if err != nil {
								licensed = true
								result.Outcome = mergeOutcome(result.Outcome, outcomeFromError(err))
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_FRAME_FAMILY_MISMATCH", err.Error(), "constructions."+familyID.String())
								continue
							}
							if compiledVariant.Variant.DirectionPolicy != construction.DirectionBoth && compiledVariant.Variant.DirectionPolicy != construction.DirectionRealize {
								continue
							}
							var validatedVariant construction.CompiledVariant
							if source.QuestionAnswer != nil && node.ID == source.QuestionAnswer.TargetNodeID {
								validatedVariant, err = construction.ValidateVariantForQuestionGap(compiledVariant.Variant, family.Family, frame, source.QuestionAnswer.Role)
							} else {
								validatedVariant, err = construction.ValidateVariantForFrame(compiledVariant.Variant, family.Family, frame)
							}
							if err != nil {
								licensed = true
								if source.QuestionAnswer != nil && node.ID == source.QuestionAnswer.TargetNodeID && isAlternateQuestionGapVariantError(err, compiledVariant.Variant, family.Family, frame, source.QuestionAnswer.Role) {
									alternateQuestionGapRejected = true
									continue
								}
								result.Outcome = mergeOutcome(result.Outcome, outcomeFromError(err))
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_FRAME_VARIANT_MISMATCH", err.Error(), "constructions."+familyID.String())
								continue
							}
							if source.QuestionAnswer != nil && node.ID == source.QuestionAnswer.TargetNodeID && !questionGapVariantHasOnlyOmittedRequiredRole(validatedVariant.Variant, family.Family, frame, source.QuestionAnswer.Role) {
								licensed = true
								result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "QUESTION_GAP_ADDITIONAL_ROLE_OMITTED", "all required roles except the typed WH answer role must have explicit target slots", "constructions."+familyID.String())
								continue
							}
							compiledVariant = validatedVariant
							plan, err := construction.Compose(input.Constructions, frame, familyID, *source.Meaning, nodeOperators)
							if err != nil {
								licensed = true
								result.Outcome = mergeOutcome(result.Outcome, outcomeFromError(err))
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_OPERATOR_COMPOSITION_REJECTED", err.Error(), "meaning."+string(node.ID))
								continue
							}
							validationFrame := frame
							if source.QuestionAnswer != nil && node.ID == source.QuestionAnswer.TargetNodeID {
								var gapOK bool
								validationFrame, gapOK = frameForQuestionGapValidation(frame, source.QuestionAnswer.Role)
								if !gapOK {
									licensed = true
									result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
									result.Diagnostics = appendDiagnostic(result.Diagnostics, "QUESTION_GAP_FRAME_ROLE_MISMATCH", "typed WH answer role is not required by the selected valency frame", "frame.arguments")
									continue
								}
							}
							validation, err := valency.ValidateFrame(validationFrame, valency.ValidationRequest{Meaning: *source.Meaning, NodeID: node.ID, SemanticBinding: transfer, ConstructionBinding: binding}, input.Core)
							if err != nil {
								licensed = true
								if alternativeRoleSetMismatch(err) {
									deferredFrameOutcomes[nodeKey] = mergeOutcome(deferredFrameOutcomes[nodeKey], outcomeFromError(err))
									if len(deferredFrameRejections[nodeKey]) < status.MaxDiagnostics {
										deferredFrameRejections[nodeKey] = append(deferredFrameRejections[nodeKey], frameRejection{status.Diagnostic{Code: "TARGET_FRAME_REJECTED", Message: err.Error(), Path: "meaning." + string(node.ID), Severity: status.SeverityWarning}})
									}
								} else {
									result.Outcome = mergeOutcome(result.Outcome, outcomeFromError(err))
									result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_FRAME_REJECTED", err.Error(), "meaning."+string(node.ID))
								}
								continue
							}
							licensed = true
							if len(result.Candidates) >= budget.MaxCandidates {
								result.Outcome = status.OutcomeResourceLimit
								result.Diagnostics = appendDiagnostic(result.Diagnostics, "RESOURCE_LIMIT", "target candidate retention reached its configured bound", "budget.candidates")
								return finish()
							}
							roles := append([]meaning.Role(nil), rolesByNode[node.ID]...)
							if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState {
								roles = roles[:0]
								for _, participant := range node.Participants {
									roles = append(roles, participant.Role)
								}
							}
							sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
							var questionAnswer *semantic.QuestionAnswer
							if source.QuestionAnswer != nil && source.QuestionAnswer.TargetNodeID == node.ID {
								answerCopy := *source.QuestionAnswer
								questionAnswer = &answerCopy
							}
							if len(validation.MissingOmittedRoles) > 0 {
								nodeOutcome = mergeOutcome(nodeOutcome, status.OutcomeIncomplete)
								result.Outcome = mergeOutcome(result.Outcome, status.OutcomeIncomplete)
								nodeDiagnostics = appendDiagnostic(nodeDiagnostics, "TARGET_ROLE_ABSENT_BY_FRAME_LICENSE", "source Meaning omits a role explicitly permitted by this exact frame", "meaning."+string(node.ID)+".participants")
							}
							result.Candidates = append(result.Candidates, TargetCandidate{SourceCandidateID: source.ID, SemanticID: source.SemanticID, NodeID: node.ID, Kind: node.Kind, QuestionAnswer: questionAnswer, Roles: roles, Outcome: nodeOutcome, Diagnostics: append([]status.Diagnostic(nil), nodeDiagnostics...), Lexeme: lexeme, Sense: sense, TransferBinding: transfer, Frame: frame, ConstructionBinding: binding, Family: family, CompiledVariant: compiledVariant, ConstructionPlan: plan, SoftConstraintMisses: append([]valency.SoftConstraintMiss(nil), validation.SoftConstraintMisses...), MissingOmittedRoles: append([]meaning.Role(nil), validation.MissingOmittedRoles...)})
							nodeCounts[nodeKey]++
						}
					}
				}
			}
			if nodeCounts[nodeKey] == 0 {
				if source.QuestionAnswer != nil && node.ID == source.QuestionAnswer.TargetNodeID && alternateQuestionGapRejected {
					result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
					result.Diagnostics = appendDiagnostic(result.Diagnostics, "NO_MATCHING_QUESTION_GAP_VARIANT", "no licensed realization variant omits the explicitly queried semantic role", "constructions."+string(node.ID))
				} else if !licensed {
					result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
					result.Diagnostics = appendDiagnostic(result.Diagnostics, "NO_LICENSED_TARGET", "no approved target transfer and exact construction license was found", "meaning."+string(node.ID))
				}
			}
		}
	}
	if input.CandidatePolicy != nil {
		bySourceID := make(map[string]semantic.MeaningCandidate, len(ordered))
		for _, source := range ordered {
			bySourceID[source.ID] = source
		}
		all := result.Candidates
		filtered := make([]TargetCandidate, 0, len(all))
		licensedPredicates := make(map[string]TargetCandidate)
		retainedPredicates := make(map[string]bool)
		predicateNodes := make(map[string]map[meaning.NodeID]bool, len(ordered))
		for _, source := range ordered {
			predicateNodes[source.ID] = policySentencePredicateIDs(source)
		}
		for _, candidate := range all {
			source := bySourceID[candidate.SourceCandidateID]
			key := candidate.SourceCandidateID + "\x00" + string(candidate.NodeID)
			sentencePredicate := predicateNodes[candidate.SourceCandidateID][candidate.NodeID]
			if sentencePredicate {
				licensedPredicates[key] = candidate
			}
			if input.CandidatePolicy.KeepTargetCandidate(source, candidate, all) {
				filtered = append(filtered, candidate)
				if sentencePredicate {
					retainedPredicates[key] = true
				}
			}
		}
		rejectedPredicates := make([]string, 0, len(licensedPredicates))
		for key := range licensedPredicates {
			if !retainedPredicates[key] {
				rejectedPredicates = append(rejectedPredicates, key)
			}
		}
		sort.Strings(rejectedPredicates)
		for _, key := range rejectedPredicates {
			result.Outcome = mergeOutcome(result.Outcome, status.OutcomeUnsupported)
			result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_PREDICATE_POLICY_REJECTED", "target policy rejected every licensed candidate for the sentence predicate", "meaning."+string(licensedPredicates[key].NodeID))
		}
		result.Candidates = filtered
	}
	// A role-set mismatch rejects one frame, not the Meaning when another
	// retained frame for the same source node covers its participants.
	retainedNodes := make(map[string]bool)
	for _, candidate := range result.Candidates {
		retainedNodes[candidate.SourceCandidateID+"\x00"+string(candidate.NodeID)] = true
	}
	rejectedKeys := make([]string, 0, len(deferredFrameRejections))
	for key := range deferredFrameRejections {
		rejectedKeys = append(rejectedKeys, key)
	}
	sort.Strings(rejectedKeys)
	for _, key := range rejectedKeys {
		if retainedNodes[key] {
			continue
		}
		result.Outcome = mergeOutcome(result.Outcome, deferredFrameOutcomes[key])
		for _, rejection := range deferredFrameRejections[key] {
			result.Diagnostics = boundedDiagnostics(result.Diagnostics, []status.Diagnostic{rejection.diagnostic})
		}
	}
	// Construction and lexical alternatives are counted after the optional
	// language-local policy has retained only the applicable licensed routes.
	referentCandidateCounts := make(map[string]int)
	for _, candidate := range result.Candidates {
		if candidate.Kind == meaning.KindReferent {
			referentCandidateCounts[candidate.SourceCandidateID+"\x00"+string(candidate.NodeID)]++
		}
	}
	referentKeys := make([]string, 0, len(referentCandidateCounts))
	for key := range referentCandidateCounts {
		referentKeys = append(referentKeys, key)
	}
	sort.Strings(referentKeys)
	for _, key := range referentKeys {
		if referentCandidateCounts[key] <= 1 {
			continue
		}
		parts := strings.SplitN(key, "\x00", 2)
		result.Outcome = mergeOutcome(result.Outcome, status.OutcomeAmbiguous)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_ALTERNATIVES_AMBIGUOUS", "multiple licensed target senses remain for this referent", "meaning."+parts[1])
	}
	eventCandidateCounts := make(map[string]int)
	for _, candidate := range result.Candidates {
		if candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState {
			eventCandidateCounts[candidate.SourceCandidateID+"\x00"+string(candidate.NodeID)]++
		}
	}
	for key, count := range eventCandidateCounts {
		if count <= 1 {
			continue
		}
		parts := strings.SplitN(key, "\x00", 2)
		path := "meaning"
		if len(parts) == 2 {
			path += "." + parts[1]
		}
		result.Outcome = mergeOutcome(result.Outcome, status.OutcomeAmbiguous)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, "TARGET_ALTERNATIVES_AMBIGUOUS", "multiple licensed target lexical and construction alternatives remain for this predicate", path)
	}
	sort.Slice(result.Candidates, func(i, j int) bool { return targetLess(result.Candidates[i], result.Candidates[j]) })
	return finish()
}

// policySentencePredicateIDs identifies sentence roots through existing IR
// wrappers. Supporting referents and attributive predicates are not roots.
func policySentencePredicateIDs(source semantic.MeaningCandidate) map[meaning.NodeID]bool {
	result := make(map[meaning.NodeID]bool)
	if source.Meaning == nil {
		return result
	}
	nodes := make(map[meaning.NodeID]meaning.Node, len(source.Meaning.Nodes))
	for _, node := range source.Meaning.Nodes {
		nodes[node.ID] = node
	}
	var visit func(meaning.NodeID, int)
	visit = func(id meaning.NodeID, depth int) {
		if depth >= meaning.MaxMeaningDepth {
			return
		}
		node := nodes[id]
		switch node.Kind {
		case meaning.KindDiscourse:
			if len(node.Content) == 1 && (node.Act == meaning.ActStatement || source.QuestionAnswer != nil && node.Act == meaning.ActQuestion) {
				visit(node.Content[0], depth+1)
			}
		case meaning.KindScope:
			if node.Operator == meaning.OperatorNegation {
				visit(node.Target, depth+1)
			}
		case meaning.KindQuestion:
			if source.QuestionAnswer != nil {
				visit(node.Target, depth+1)
			}
		case meaning.KindEvent, meaning.KindState:
			result[id] = true
		}
	}
	for _, root := range source.Meaning.Roots {
		visit(root, 0)
	}
	return result
}

func operatorsForTarget(operators []construction.Operator, target meaning.NodeID) []construction.Operator {
	result := make([]construction.Operator, 0, len(operators))
	for _, operator := range operators {
		if operator.SemanticScope.TargetNodeID == target {
			result = append(result, operator)
		}
	}
	return result
}

func hasUnmappedScope(nodes []meaning.Node, target meaning.NodeID, operators []construction.Operator) bool {
	for _, node := range meaningAncestors(nodes, target) {
		switch node.Kind {
		case meaning.KindScope, meaning.KindModality, meaning.KindQuestion, meaning.KindQuantifier:
			mapped := false
			if node.Kind != meaning.KindQuantifier {
				for _, operator := range operators {
					if operatorMatchesWrapper(operator, node, target) {
						mapped = true
						break
					}
				}
			}
			if !mapped {
				return true
			}
		}
	}
	return false
}

func meaningGraphWork(nodes []meaning.Node) int {
	work := len(nodes)
	for _, node := range nodes {
		switch node.Kind {
		case meaning.KindEvent, meaning.KindState:
			work += len(node.Participants)
		case meaning.KindQuestion, meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
			if node.Target != "" {
				work++
			}
		case meaning.KindDiscourse:
			work += len(node.Content)
		}
	}
	return work
}

// meaningAncestors returns wrappers and discourse containers whose semantic
// target/content chain reaches target. It deliberately excludes predicate
// participant edges, which do not scope their participant's lexical choice.
func meaningAncestors(nodes []meaning.Node, target meaning.NodeID) []meaning.Node {
	parents := make(map[meaning.NodeID][]meaning.NodeID, len(nodes))
	byID := make(map[meaning.NodeID]meaning.Node, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
		switch node.Kind {
		case meaning.KindQuestion, meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
			parents[node.Target] = append(parents[node.Target], node.ID)
		case meaning.KindDiscourse:
			for _, child := range node.Content {
				parents[child] = append(parents[child], node.ID)
			}
		}
	}
	seen := map[meaning.NodeID]struct{}{target: {}}
	queue := []meaning.NodeID{target}
	for head := 0; head < len(queue); head++ {
		for _, parent := range parents[queue[head]] {
			if _, exists := seen[parent]; exists {
				continue
			}
			seen[parent] = struct{}{}
			queue = append(queue, parent)
		}
	}
	result := make([]meaning.Node, 0, len(seen)-1)
	for id := range seen {
		if id != target {
			result = append(result, byID[id])
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func hasUnsupportedDiscourseAct(nodes []meaning.Node, operators []construction.Operator) bool {
	for _, node := range nodes {
		if node.Kind != meaning.KindDiscourse {
			continue
		}
		if node.Act == meaning.ActQuestion && len(node.Content) == 1 {
			question, found := meaningNodeIDInNodes(nodes, node.Content[0])
			if !found || question.Kind != meaning.KindQuestion || !hasQuestionWrapperOperator(operators, question) {
				return true
			}
			continue
		}
		if node.Act != meaning.ActStatement {
			// Operator evidence is scoped to a predicate, and this API has no
			// explicit link from a discourse act to its realizing operator.
			// A question operator elsewhere in the graph cannot license it.
			return true
		}
	}
	return false
}

func validateQuestionDiscourse(graph meaning.Meaning, answer semantic.QuestionAnswer) error {
	questionCount, questionID := 0, meaning.NodeID("")
	discourseCount := 0
	discourseID := meaning.NodeID("")
	predicateCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindQuestion {
			questionCount++
			questionID = node.ID
			if node.Target != answer.TargetNodeID {
				return fmt.Errorf("question wrapper target does not match the typed answer target")
			}
		}
		if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState {
			predicateCount++
			if node.ID != answer.TargetNodeID || node.Kind != meaning.KindEvent {
				return fmt.Errorf("WH realization supports only one queried event predicate")
			}
		}
		if node.Kind == meaning.KindScope || node.Kind == meaning.KindModality || node.Kind == meaning.KindQuantifier {
			return fmt.Errorf("WH realization does not support additional scoped operators")
		}
		if node.Kind == meaning.KindDiscourse && node.Act == meaning.ActQuestion {
			discourseCount++
			discourseID = node.ID
		} else if node.Kind == meaning.KindDiscourse {
			return fmt.Errorf("WH realization does not support additional discourse acts")
		}
	}
	if questionCount != 1 || discourseCount != 1 || predicateCount != 1 || questionID == "" || discourseID == "" {
		return fmt.Errorf("WH realization requires exactly one question wrapper and one question discourse node")
	}
	for _, node := range graph.Nodes {
		if node.ID == discourseID && (len(node.Content) != 1 || node.Content[0] != questionID) {
			return fmt.Errorf("question discourse must directly contain the unique question wrapper")
		}
	}
	if len(graph.Roots) != 1 || graph.Roots[0] != discourseID {
		return fmt.Errorf("WH realization requires the question discourse node to be the sole root")
	}
	return nil
}

func hasQuestionAct(nodes []meaning.Node) bool {
	for _, node := range nodes {
		if node.Kind == meaning.KindQuestion || node.Kind == meaning.KindDiscourse && node.Act == meaning.ActQuestion {
			return true
		}
	}
	return false
}

func familySupportsWHQuestion(family construction.ConstructionFamily) bool {
	hasQuestionOperator, hasQuestionCapability := false, false
	for _, operator := range family.Operators {
		if operator == "question-wh" {
			hasQuestionOperator = true
			break
		}
	}
	for _, capability := range family.Capabilities {
		if capability == "question-wh" {
			hasQuestionCapability = true
			break
		}
	}
	return hasQuestionOperator && hasQuestionCapability
}

func questionGapVariantHasOnlyOmittedRequiredRole(variant construction.Variant, family construction.ConstructionFamily, frame valency.ValencyFrame, answerRole meaning.Role) bool {
	roleSlots := make(map[string]string, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		if slot.ArgumentBinding != nil {
			roleSlots[string(slot.ArgumentBinding.SemanticRole)] = slot.ID
		}
	}
	sequenceSlots := make(map[string]struct{}, len(variant.Sequence))
	for _, token := range variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	answerSlot, exists := roleSlots[string(answerRole)]
	if !exists {
		return false
	}
	if _, included := sequenceSlots[answerSlot]; included {
		return false
	}
	for _, argument := range frame.Arguments {
		if !argument.Required || string(argument.Role) == string(answerRole) {
			continue
		}
		slotID, bound := roleSlots[string(argument.Role)]
		if !bound {
			return false
		}
		if _, included := sequenceSlots[slotID]; !included {
			return false
		}
	}
	return true
}

func isAlternateQuestionGapVariantError(err error, variant construction.Variant, family construction.ConstructionFamily, frame valency.ValencyFrame, answerRole meaning.Role) bool {
	var resultError interface{ Result() status.Result }
	if !errors.As(err, &resultError) {
		return false
	}
	result := resultError.Result()
	filledGap := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == "QUESTION_GAP_SLOT_FILLED" {
			filledGap = true
			break
		}
	}
	if !filledGap {
		return false
	}
	hasQuestionOperator, hasQuestionCapability := false, false
	for _, operator := range family.Operators {
		if operator == "question-wh" {
			hasQuestionOperator = true
			break
		}
	}
	for _, capability := range family.Capabilities {
		if capability == "question-wh" {
			hasQuestionCapability = true
			break
		}
	}
	if !hasQuestionOperator || !hasQuestionCapability {
		return false
	}
	roleSlots := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		if slot.ArgumentBinding != nil {
			roleSlots[string(slot.ArgumentBinding.SemanticRole)] = slot
		}
	}
	answerSlot, exists := roleSlots[string(answerRole)]
	if !exists || answerSlot.Category != construction.CategoryNP || answerSlot.Optionality != construction.SlotOptional || answerSlot.Cardinality.Minimum != 0 || answerSlot.Cardinality.Maximum != 1 {
		return false
	}
	sequenceSlots := make(map[string]struct{}, len(variant.Sequence))
	for _, token := range variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	if _, answerIncluded := sequenceSlots[answerSlot.ID]; !answerIncluded {
		return false
	}
	omittedRole := ""
	for _, argument := range frame.Arguments {
		if !argument.Required {
			continue
		}
		slot, bound := roleSlots[string(argument.Role)]
		if !bound {
			return false
		}
		if _, included := sequenceSlots[slot.ID]; included {
			continue
		}
		if omittedRole != "" || argument.Role == answerRole || slot.Category != construction.CategoryNP || slot.Optionality != construction.SlotOptional || slot.Cardinality.Minimum != 0 || slot.Cardinality.Maximum != 1 {
			return false
		}
		omittedRole = string(argument.Role)
	}
	return omittedRole != ""
}

// frameForQuestionGapValidation adapts only the validation view of one frame:
// the explicit typed answer supplies its otherwise-required event role. The
// approved frame stored on the selected target candidate remains unchanged.
func frameForQuestionGapValidation(frame valency.ValencyFrame, answerRole meaning.Role) (valency.ValencyFrame, bool) {
	copy := frame
	copy.Arguments = append([]valency.ArgumentSlot(nil), frame.Arguments...)
	for index := range copy.Arguments {
		if copy.Arguments[index].Role != answerRole {
			continue
		}
		if !copy.Arguments[index].Required {
			return valency.ValencyFrame{}, false
		}
		copy.Arguments[index].Required = false
		return copy, true
	}
	return valency.ValencyFrame{}, false
}

func hasExactQuestionOperator(graph meaning.Meaning, operators []construction.Operator, answer semantic.QuestionAnswer) bool {
	var question meaning.Node
	questionCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindQuestion {
			question = node
			questionCount++
		}
	}
	if questionCount != 1 || question.Target != answer.TargetNodeID || len(operators) != 1 {
		return false
	}
	return hasQuestionWrapperOperator(operators, question)
}

func hasQuestionWrapperOperator(operators []construction.Operator, question meaning.Node) bool {
	if question.Kind != meaning.KindQuestion {
		return false
	}
	matched := 0
	for _, operator := range operators {
		if operator.SemanticScope.ScopeNodeID != question.ID {
			continue
		}
		if operator.Kind != construction.OperatorQuestion || operator.Value != construction.ValueWH || operator.SemanticScope.TargetNodeID != question.Target || !operatorMatchesWrapper(operator, question, question.Target) {
			return false
		}
		matched++
	}
	return matched == 1
}

func meaningNodeIDInNodes(nodes []meaning.Node, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func boundedDiagnostics(first, second []status.Diagnostic) []status.Diagnostic {
	result := make([]status.Diagnostic, 0, minInt(status.MaxDiagnostics, len(first)+len(second)))
	for _, source := range [][]status.Diagnostic{first, second} {
		for _, diagnostic := range source {
			if len(result) >= status.MaxDiagnostics {
				return result
			}
			found := false
			for _, existing := range result {
				if existing == diagnostic {
					found = true
					break
				}
			}
			if !found {
				result = append(result, diagnostic)
			}
		}
	}
	return result
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func targetLess(left, right TargetCandidate) bool {
	if left.SourceCandidateID != right.SourceCandidateID {
		return left.SourceCandidateID < right.SourceCandidateID
	}
	if left.NodeID != right.NodeID {
		return left.NodeID < right.NodeID
	}
	if left.Sense.ID != right.Sense.ID {
		return left.Sense.ID.String() < right.Sense.ID.String()
	}
	if left.Frame.ID != right.Frame.ID {
		return left.Frame.ID < right.Frame.ID
	}
	if left.Family.Family.ID != right.Family.Family.ID {
		return left.Family.Family.ID.String() < right.Family.Family.ID.String()
	}
	return left.ConstructionBinding.VariantID.String() < right.ConstructionBinding.VariantID.String()
}

func validLanguage(language lexicon.LanguageID) bool {
	value := string(language)
	if len(value) == 0 || len(value) > construction.MaxLanguageBytes || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if !(character >= 'a' && character <= 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

func boundedSourceID(value string) bool {
	return len(value) > 0 && len(value) <= 256 && utf8.ValidString(value) && validPlanField(value)
}

func appendDiagnostic(diagnostics []status.Diagnostic, code, message, path string) []status.Diagnostic {
	if len(diagnostics) < status.MaxDiagnostics {
		diagnostics = append(diagnostics, status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning})
	}
	return diagnostics
}

func mergeOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	priority := map[status.OutcomeKind]int{
		status.OutcomeOK: 0, status.OutcomeUnsupported: 1, status.OutcomeUnknown: 2,
		status.OutcomeAmbiguous: 3, status.OutcomeIncomplete: 4,
		status.OutcomeResourceLimit: 7, status.OutcomeError: 8, status.OutcomeInvalid: 9,
	}
	if priority[right] > priority[left] {
		return right
	}
	return left
}

func outcomeFromError(err error) status.OutcomeKind {
	if err == nil {
		return status.OutcomeOK
	}
	var resultError interface{ Result() status.Result }
	if errors.As(err, &resultError) {
		result := resultError.Result()
		if result.Status.Valid() {
			return result.Status
		}
	}
	return status.OutcomeError
}

func alternativeRoleSetMismatch(err error) bool {
	var resultError interface{ Result() status.Result }
	if !errors.As(err, &resultError) {
		return false
	}
	diagnostics := resultError.Result().Diagnostics
	if len(diagnostics) != 1 {
		return false
	}
	return diagnostics[0].Code == "EXTRA_SEMANTIC_ROLE" || diagnostics[0].Code == "MISSING_REQUIRED_ROLE"
}

func selectionError(kind status.OutcomeKind, code, path, message string) error {
	return &SelectionError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}

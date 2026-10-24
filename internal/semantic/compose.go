// Package semantic composes bounded parser evidence with approved lexical,
// valency and construction records. It never guesses a missing transfer.
package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultMaxSemanticAlternatives = 4096
	maxSemanticAlternatives        = 16384
	defaultMaxSemanticExpansions   = 16384
	maxSemanticExpansions          = 65536
)

const (
	// DiagnosticUnresolvedRoleSense marks a partial candidate whose only
	// semantic uncertainty is an unapproved competing lexical sense in a role.
	DiagnosticUnresolvedRoleSense = "UNRESOLVED_ROLE_SENSE"
	// UnresolvedRoleSenseReason describes the exact unresolved condition above.
	UnresolvedRoleSenseReason = "role candidate has a competing sense without an approved equivalent CORE binding"
)

// Budget bounds semantic alternatives and expansion work independently from
// parser and Meaning graph budgets.
type Budget struct {
	MaxAlternatives int
	MaxExpansions   int
	Evidence        EvidenceBudget
	Meaning         meaning.Budget
}

func DefaultBudget() Budget {
	return Budget{MaxAlternatives: defaultMaxSemanticAlternatives, MaxExpansions: defaultMaxSemanticExpansions, Evidence: DefaultEvidenceBudget(), Meaning: meaning.DefaultBudget()}
}

func (budget Budget) Validate() error {
	if budget.MaxAlternatives <= 0 || budget.MaxAlternatives > maxSemanticAlternatives || budget.MaxExpansions <= 0 || budget.MaxExpansions > maxSemanticExpansions {
		return semanticError(status.OutcomeInvalid, "INVALID_SEMANTIC_BUDGET", "budget", "semantic limits must be positive and within package ceilings")
	}
	if err := budget.Evidence.Validate(); err != nil {
		return err
	}
	return budget.Meaning.Validate()
}

// ReferenceEvidence either resolves a mention to explicit request-local
// identity evidence or preserves an unresolved deictic anchor. Equal CORE
// concepts alone never join mentions into one referent.
type ReferenceEvidence struct {
	CandidateID string
	ReferentID  meaning.ReferentID
	EvidenceID  string
	Anchor      *meaning.DeicticAnchor
}

// OperatorEvidence supplies a construction operator and its explicit target
// candidate. Only operators representable in Meaning v1 are composed.
type OperatorEvidence struct {
	Operator          construction.Operator
	TargetCandidateID string
	Span              text.Span
}

// CompositionInput contains already bounded parser and lexical structures.
// Operators and reference identity are supplied explicitly by the caller.
type CompositionInput struct {
	Forest                   parser.Forest
	Candidates               []parser.PhraseCandidate
	Lattice                  analyzer.LexicalLattice
	SourceBytes              int
	Lexicon                  *lexicon.Index
	Valency                  *valency.Index
	Core                     *core.ConceptGraph
	Constructions            *construction.Registry
	References               []ReferenceEvidence
	NodeAttributes           []NodeAttributesEvidence
	RequestContextID         string
	Operators                []OperatorEvidence
	OperatorCoverageComplete bool
	Budget                   Budget
}

// ComposeResult preserves supported graphs and evidence-only competitors.
type ComposeResult struct {
	Outcome     status.OutcomeKind
	Candidates  []MeaningCandidate
	Diagnostics []status.Diagnostic
}

type transfer struct {
	binding lexicon.SemanticBinding
	sense   ids.SenseID
}
type roleOption struct {
	role      meaning.Role
	candidate parser.PhraseCandidate
	mapping   transfer
	span      text.Span
}

// Compose converts each retained parser alternative independently, retaining
// all approved equivalent mappings and every unresolved competitor.
func Compose(input CompositionInput) (ComposeResult, error) {
	budget := input.Budget
	if budget == (Budget{}) {
		budget = DefaultBudget()
	}
	if err := budget.Validate(); err != nil {
		return ComposeResult{}, err
	}
	if input.RequestContextID == "" {
		input.RequestContextID = "semantic-request"
	}
	if input.Lexicon == nil || input.Valency == nil || input.Core == nil || input.Constructions == nil {
		return ComposeResult{}, semanticError(status.OutcomeInvalid, "MISSING_SEMANTIC_INDEX", "indexes", "lexicon, valency, CORE and construction indexes are required")
	}
	if len(input.Forest.Nodes) > parser.DefaultBudget().MaxNodes || len(input.Candidates) > parser.DefaultBudget().MaxExpansions || len(input.Operators) > construction.MaxOperatorsPerPlan {
		return ComposeResult{}, semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "input", "semantic input exceeds its bounded inventory")
	}
	if len(input.Forest.Roots) > parser.DefaultBudget().MaxNodes || len(input.Forest.Unmatched) > parser.DefaultBudget().MaxExpansions || len(input.Lattice.Tokens) > status.DefaultBounds().MaxInputBytes {
		return ComposeResult{}, semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "input", "semantic input exceeds its bounded inventory")
	}
	if len(input.Forest.Diagnostics) > status.MaxDiagnostics {
		return ComposeResult{}, semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest.diagnostics", "forest diagnostic count exceeds the status limit")
	}
	for i, diagnostic := range input.Forest.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_FOREST_DIAGNOSTIC", fmt.Sprintf("forest.diagnostics[%d]", i), err.Error())
		}
	}
	sourceBytes := input.SourceBytes
	if sourceBytes == 0 {
		for _, token := range input.Lattice.Tokens {
			if token.Token.Span.End > sourceBytes {
				sourceBytes = token.Token.Span.End
			}
		}
	}
	if sourceBytes < 0 || sourceBytes > status.DefaultBounds().MaxInputBytes {
		return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_SOURCE_LENGTH", "source_bytes", "source length is outside the bounded input range")
	}
	if err := validateCompositionInputs(input, sourceBytes); err != nil {
		return ComposeResult{}, err
	}
	candidateByID := make(map[string]parser.PhraseCandidate, len(input.Candidates))
	for _, candidate := range input.Candidates {
		candidateByID[candidate.ID] = candidate
	}
	attributesByCandidate, err := validateNodeAttributes(input.NodeAttributes, candidateByID)
	if err != nil {
		return ComposeResult{}, err
	}
	referenceByCandidate := make(map[string]ReferenceEvidence, len(input.References))
	for _, reference := range input.References {
		if _, exists := candidateByID[reference.CandidateID]; !exists || reference.CandidateID == "" {
			return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_REFERENCE_EVIDENCE", "references", "reference evidence requires an existing candidate")
		}
		resolved := reference.ReferentID != "" && reference.EvidenceID != ""
		unresolved := reference.ReferentID == "" && reference.EvidenceID == "" && reference.Anchor != nil
		if !resolved && !unresolved {
			return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_REFERENCE_EVIDENCE", "references", "reference evidence requires both identity fields or an unresolved anchor")
		}
		if reference.Anchor != nil {
			anchorContext := meaning.RequestContext{
				ID:       "reference-validation",
				Mentions: []meaning.Mention{{ID: "mention", NodeID: "node", Anchor: reference.Anchor}},
			}
			if err := anchorContext.Validate(); err != nil {
				return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_REFERENCE_ANCHOR", "references.anchor", err.Error())
			}
		}
		if _, exists := referenceByCandidate[reference.CandidateID]; exists {
			return ComposeResult{}, semanticError(status.OutcomeInvalid, "DUPLICATE_REFERENCE_EVIDENCE", "references", "candidate has more than one reference resolution")
		}
		referenceByCandidate[reference.CandidateID] = reference
	}
	result := ComposeResult{Outcome: input.Forest.Outcome, Diagnostics: append([]status.Diagnostic(nil), input.Forest.Diagnostics...)}
	if !result.Outcome.Valid() {
		return ComposeResult{}, semanticError(status.OutcomeInvalid, "INVALID_FOREST_OUTCOME", "forest.outcome", "parser forest has an invalid outcome")
	}
	work := 0
	rootByID := make(map[parser.NodeID]parser.PackedNode, len(input.Forest.Nodes))
	phraseNodeByCandidate := make(map[string]parser.NodeID)
	for _, node := range input.Forest.Nodes {
		rootByID[node.ID] = node
		for _, candidateID := range node.CandidateIDs {
			if _, exists := phraseNodeByCandidate[candidateID]; !exists {
				phraseNodeByCandidate[candidateID] = node.ID
			}
		}
	}
	for _, rootID := range input.Forest.Roots {
		root, exists := rootByID[rootID]
		if !exists {
			return ComposeResult{}, semanticError(status.OutcomeInvalid, "UNKNOWN_FOREST_ROOT", "forest.roots", "parser root does not resolve")
		}
		for _, alternative := range root.Alternatives {
			predicateNodes := make([]parser.PackedNode, 0, 1)
			for _, slot := range alternative.Slots {
				for _, nodeID := range slot.NodeIDs {
					if node, found := rootByID[nodeID]; found && node.Category.IsPredicate() {
						// Packed predicate nodes can be shared by several readings. The
						// alternative's slot binding is the candidate selection for
						// this exact derivation; retain it when composing meaning.
						if len(slot.CandidateIDs) != 0 {
							node.CandidateIDs = append([]string(nil), slot.CandidateIDs...)
						}
						predicateNodes = append(predicateNodes, node)
					}
				}
			}
			work++
			if work > budget.MaxExpansions || len(result.Candidates) >= budget.MaxAlternatives {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic alternative or expansion budget exhausted", "budget"))
				return finalizeComposition(result, input.Core, budget)
			}
			if err := composeAlternative(&result, input, alternative, root.Span, sourceBytes, candidateByID, referenceByCandidate, attributesByCandidate, predicateNodes, phraseNodeByCandidate, &work, budget); err != nil {
				return ComposeResult{}, err
			}
		}
	}
	for _, unmatched := range input.Forest.Unmatched {
		candidate, exists := candidateByID[unmatched.CandidateID]
		if !exists {
			result.Outcome = mergeSemanticOutcome(result.Outcome, status.OutcomeInvalid)
			result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("UNKNOWN_UNMATCHED_CANDIDATE", "unmatched parser candidate does not resolve in the supplied candidate inventory", "forest.unmatched"))
			continue
		}
		appendEvidenceOnly(&result, input, candidate, status.OutcomeUnsupported, "UNMATCHED_PARSE_CANDIDATE", "candidate was not used by a retained construction parse", budget)
		if len(result.Candidates) >= budget.MaxAlternatives {
			result.Outcome = status.OutcomeResourceLimit
			result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic alternative budget exhausted", "budget"))
			break
		}
	}
	if len(result.Candidates) == 0 {
		result.Outcome = mergeSemanticOutcome(result.Outcome, status.OutcomeUnknown)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("NO_SEMANTIC_ALTERNATIVES", "no parser alternatives were available for semantic composition", "forest"))
	}
	return finalizeComposition(result, input.Core, budget)
}

func validateCompositionInputs(input CompositionInput, sourceBytes int) error {
	maxAlternatives := parser.DefaultBudget().MaxAlternatives
	maxNestedItems := parser.DefaultBudget().MaxExpansions
	var alternatives, nestedItems int
	var textBytes int
	addText := func(values ...string) bool {
		for _, value := range values {
			textBytes += len(value)
			if textBytes > maxEvidenceTextBytes {
				return false
			}
		}
		return true
	}
	checkSpan := func(span text.Span, path string) error {
		if span.Start < 0 || span.End <= span.Start || span.End > sourceBytes {
			return semanticError(status.OutcomeInvalid, "INVALID_SOURCE_SPAN", path, "source span must be a non-empty byte range within the declared source length")
		}
		return nil
	}
	checkInventory := func(path string) error {
		if alternatives > maxAlternatives || nestedItems > maxNestedItems {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", path, "nested semantic input exceeds its bounded inventory")
		}
		return nil
	}
	for i, node := range input.Forest.Nodes {
		if !addText(string(node.ID), string(node.Category)) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
		}
		if err := checkSpan(node.Span, fmt.Sprintf("forest.nodes[%d].span", i)); err != nil {
			return err
		}
		alternatives += len(node.Alternatives)
		nestedItems += len(node.CandidateIDs) + len(node.Alternatives)
		if err := checkInventory("forest"); err != nil {
			return err
		}
		for _, candidateID := range node.CandidateIDs {
			if !addText(candidateID) {
				return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
			}
		}
		for j, alternative := range node.Alternatives {
			if !addText(string(alternative.PlanID), alternative.FamilyID.String(), string(alternative.VariantID), alternative.Permission.SenseID.String(), string(alternative.Permission.FrameID), alternative.DerivationID) {
				return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
			}
			nestedItems += len(alternative.Slots) + len(alternative.Unresolved)
			if err := checkInventory("forest"); err != nil {
				return err
			}
			for _, unresolved := range alternative.Unresolved {
				if !addText(unresolved) {
					return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
				}
			}
			for k, slot := range alternative.Slots {
				if !addText(slot.SlotID, string(slot.Role)) {
					return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
				}
				if slot.WHGap != nil && !addText(slot.WHGap.SlotID, string(slot.WHGap.Role), string(slot.WHGap.ScopeNodeID)) {
					return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
				}
				nestedItems += len(slot.CandidateIDs) + len(slot.NodeIDs)
				if err := checkInventory("forest"); err != nil {
					return err
				}
				for _, candidateID := range slot.CandidateIDs {
					if !addText(candidateID) {
						return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
					}
				}
				for _, nodeID := range slot.NodeIDs {
					if !addText(string(nodeID)) {
						return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest", "nested parser text exceeds the semantic text budget")
					}
				}
				if slot.WHGap != nil && (len(slot.WHGap.SlotID)+len(slot.WHGap.ScopeNodeID)) > 512 {
					return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("forest.nodes[%d].alternatives[%d].slots[%d]", i, j, k), "nested parser metadata exceeds its text bound")
				}
			}
		}
	}
	for _, rootID := range input.Forest.Roots {
		if !addText(string(rootID)) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest.roots", "root identifiers exceed the semantic text budget")
		}
	}
	candidateIDs := make(map[string]struct{}, len(input.Candidates))
	for i, candidate := range input.Candidates {
		if !addText(candidate.ID, string(candidate.Category), candidate.Case, candidate.LexemeID.String()) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "candidates", "nested candidate text exceeds the semantic text budget")
		}
		if candidate.ID == "" {
			return semanticError(status.OutcomeInvalid, "INVALID_CANDIDATE_ID", fmt.Sprintf("candidates[%d].id", i), "phrase candidate ID must be non-empty")
		}
		if _, exists := candidateIDs[candidate.ID]; exists {
			return semanticError(status.OutcomeInvalid, "DUPLICATE_CANDIDATE_ID", fmt.Sprintf("candidates[%d].id", i), "phrase candidate IDs must be unique")
		}
		candidateIDs[candidate.ID] = struct{}{}
		if err := checkSpan(candidate.Span, fmt.Sprintf("candidates[%d].span", i)); err != nil {
			return err
		}
		nestedItems += len(candidate.ReadingIDs) + len(candidate.SenseIDs) + len(candidate.Features)
		if err := checkInventory("candidates"); err != nil {
			return err
		}
		for _, value := range candidate.ReadingIDs {
			if !addText(value) {
				return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "candidates", "nested candidate text exceeds the semantic text budget")
			}
		}
		for _, value := range candidate.SenseIDs {
			if !addText(value.String()) {
				return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "candidates", "nested candidate text exceeds the semantic text budget")
			}
		}
		for feature, value := range candidate.Features {
			if !addText(string(feature), string(value)) {
				return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "candidates", "candidate feature text exceeds the semantic text budget")
			}
		}
	}
	for i, token := range input.Lattice.Tokens {
		if !addText(token.Token.Surface, token.Token.Normalized, token.Token.Folded, string(token.Token.Kind)) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "lattice", "lexical token text exceeds the semantic text budget")
		}
		if err := checkSpan(token.Token.Span, fmt.Sprintf("lattice.tokens[%d].span", i)); err != nil {
			return err
		}
		nestedItems += len(token.Readings)
		if err := checkInventory("lattice"); err != nil {
			return err
		}
	}
	for i, unmatched := range input.Forest.Unmatched {
		if !addText(unmatched.CandidateID, unmatched.Reason) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "forest.unmatched", "unmatched candidate text exceeds the semantic text budget")
		}
		if len(unmatched.CandidateID)+len(unmatched.Reason) > 1024 {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("forest.unmatched[%d]", i), "unmatched candidate metadata exceeds the text bound")
		}
	}
	if len(input.References) > len(input.Candidates) || !addText(input.RequestContextID) {
		return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "references", "reference inventory or request context exceeds the semantic text budget")
	}
	if len(input.NodeAttributes) > len(input.Candidates) {
		return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "node_attributes", "node attribute evidence exceeds the candidate inventory")
	}
	for i, attribute := range input.NodeAttributes {
		if !addText(attribute.CandidateID, string(attribute.TemporalFrame), string(attribute.Multiplicity)) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("node_attributes[%d]", i), "node attribute evidence exceeds the semantic text budget")
		}
	}
	for i, reference := range input.References {
		if !addText(reference.CandidateID, string(reference.ReferentID), reference.EvidenceID) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("references[%d]", i), "reference evidence exceeds the semantic text budget")
		}
		if reference.Anchor != nil && !addText(reference.Anchor.Kind, reference.Anchor.Value) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("references[%d].anchor", i), "reference anchor text exceeds the semantic text budget")
		}
	}
	for i, operator := range input.Operators {
		if !addText(operator.TargetCandidateID, string(operator.Operator.Kind), string(operator.Operator.Value), string(operator.Operator.SemanticScope.TargetNodeID), string(operator.Operator.SemanticScope.ScopeNodeID)) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("operators[%d]", i), "operator evidence exceeds the semantic text budget")
		}
		provenance := operator.Operator.Provenance
		if !addText(provenance.ID, provenance.Path, provenance.URI, provenance.Version, provenance.License, provenance.Author, provenance.ChecksumSHA256) {
			return semanticError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("operators[%d]", i), "operator provenance exceeds the semantic text budget")
		}
	}
	return nil
}

func composeAlternative(result *ComposeResult, input CompositionInput, alternative parser.ParseAlternative, span text.Span, sourceBytes int, candidates map[string]parser.PhraseCandidate, references map[string]ReferenceEvidence, attributes map[string]NodeAttributesEvidence, predicateNodes []parser.PackedNode, phraseNodes map[string]parser.NodeID, work *int, budget Budget) error {
	baseRecords := []EvidenceRecord{}
	if span.End <= span.Start {
		span = firstSourceSpan(input.Lattice)
	}
	if span.End <= span.Start {
		return semanticError(status.OutcomeInvalid, "MISSING_SOURCE_SPAN", "forest", "parse alternative has no source span for bounded evidence")
	}
	if len(alternative.Unresolved) > 0 {
		// Keep composing the known parts. The missing slots will remain absent
		// from the graph and the resulting candidate will carry an incomplete
		// status; no placeholder participant is fabricated.
	}
	partialStatus, partialReason := status.OutcomeOK, ""
	if len(alternative.Unresolved) > 0 {
		partialStatus, partialReason = status.OutcomeIncomplete, "one or more construction slots remain unresolved"
	}
	if len(predicateNodes) != 1 {
		appendAlternativeEvidenceOnly(result, input, alternative, span, status.OutcomeIncomplete, "PREDICATE_NODE_UNRESOLVED", "parse alternative does not identify exactly one predicate phrase", baseRecords, budget)
		return nil
	}
	predicateNode := predicateNodes[0]
	if len(predicateNode.CandidateIDs) != 1 {
		appendAlternativeEvidenceOnly(result, input, alternative, predicateNode.Span, status.OutcomeIncomplete, "PREDICATE_CANDIDATE_UNRESOLVED", "predicate phrase contains multiple candidate IDs", baseRecords, budget)
		return nil
	}
	predicate, exists := candidates[predicateNode.CandidateIDs[0]]
	if !exists {
		return semanticError(status.OutcomeInvalid, "UNKNOWN_PHRASE_CANDIDATE", "predicate", "predicate phrase candidate does not resolve")
	}
	if !containsSense(predicate.SenseIDs, alternative.Permission.SenseID) {
		reason := "predicate candidate does not contain the exact parse permission sense"
		records := []EvidenceRecord{{SourceSpan: predicate.Span, CandidateID: predicate.ID, ReadingID: firstString(predicate.ReadingIDs), SenseID: alternative.Permission.SenseID, Outcome: status.OutcomeUnsupported, Reason: reason}}
		appendAlternativeEvidenceOnly(result, input, alternative, predicate.Span, status.OutcomeUnsupported, "PREDICATE_SENSE_MISMATCH", reason, records, budget)
		return nil
	}
	predicateMappings := transferMappings(input.Lexicon, predicate, &alternative.Permission.SenseID)
	if len(predicateMappings) == 0 {
		reason := "predicate sense has no approved equivalent CORE binding"
		records := []EvidenceRecord{{SourceSpan: predicate.Span, CandidateID: predicate.ID, ReadingID: firstString(predicate.ReadingIDs), SenseID: alternative.Permission.SenseID, Outcome: status.OutcomeUnsupported, Reason: reason}}
		appendAlternativeEvidenceOnly(result, input, alternative, predicate.Span, status.OutcomeUnsupported, "UNLICENSED_SEMANTIC_TRANSFER", reason, records, budget)
		return nil
	}
	frame, binding, ok := exactFrameBinding(input, alternative)
	if !ok {
		appendAlternativeEvidenceOnly(result, input, alternative, predicate.Span, status.OutcomeUnsupported, "EXACT_FRAME_BINDING_UNAVAILABLE", "exact sense, frame, family and variant construction binding is unavailable", baseRecords, budget)
		return nil
	}
	roleOptions, unresolvedStatus, unresolvedReason := resolveRoles(input, alternative, candidates)
	unresolvedRoleRecords := []EvidenceRecord{}
	if unresolvedReason != "" {
		if unresolvedStatus == status.OutcomeUnsupported {
			unmappedRecords := unmappedRoleEvidence(alternative, candidates, input.Lexicon, unresolvedReason)
			appendAlternativeEvidenceOnly(result, input, alternative, span, unresolvedStatus, "ROLE_MAPPING_UNAVAILABLE", unresolvedReason, unmappedRecords, budget)
			return nil
		}
		if unresolvedStatus == status.OutcomeIncomplete {
			unresolvedRoleRecords = unmappedRoleEvidence(alternative, candidates, input.Lexicon, unresolvedReason)
		}
		partialStatus = mergeSemanticOutcome(partialStatus, unresolvedStatus)
		partialReason = mergePartialReason(partialReason, unresolvedReason)
	}
	if len(result.Candidates) >= budget.MaxAlternatives {
		result.Outcome = status.OutcomeResourceLimit
		result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic alternative budget exhausted", "budget"))
		return nil
	}
	for _, predicateMapping := range predicateMappings {
		if *work >= budget.MaxExpansions {
			result.Outcome = status.OutcomeResourceLimit
			result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic mapping expansion budget exhausted", "budget"))
			return nil
		}
		*work++
		combinations := roleCombinations(roleOptions, budget.MaxExpansions-*work)
		if len(combinations) == 0 && len(roleOptions) > 0 {
			appendAlternativeEvidenceOnly(result, input, alternative, span, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "role mapping combinations exceed semantic budget", baseRecords, budget)
			return nil
		}
		if len(roleOptions) == 0 {
			combinations = [][]roleOption{{}}
		}
		for _, combination := range combinations {
			*work += len(input.Operators) + 1
			if *work > budget.MaxExpansions {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic mapping expansion budget exhausted", "budget"))
				return nil
			}
			candidate, _, misses, outcome, diagnostics, err := buildMeaningCandidate(input, alternative, predicate, predicateMapping, frame, binding, combination, references, attributes, partialStatus, partialReason, phraseNodes, sourceBytes, budget)
			if err != nil {
				return err
			}
			candidate.Evidence.Outcome = outcome
			candidate.Evidence.Records = append(candidate.Evidence.Records, unresolvedRoleRecords...)
			if unresolvedReason == UnresolvedRoleSenseReason && len(unresolvedRoleRecords) > 0 {
				candidate.Evidence.Diagnostics = appendDiagnostic(candidate.Evidence.Diagnostics, semanticDiagnostic(DiagnosticUnresolvedRoleSense, unresolvedReason, "roles"))
			}
			candidate.Status = outcome
			candidate.SoftConstraintMisses = misses
			if err := candidate.Evidence.ValidateWithBudget(budget.Evidence); err != nil {
				return err
			}
			result.Candidates = append(result.Candidates, candidate)
			result.Outcome = mergeSemanticOutcome(result.Outcome, outcome)
			result.Diagnostics = appendDiagnostics(result.Diagnostics, diagnostics...)
			if len(result.Candidates) >= budget.MaxAlternatives {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "semantic alternative budget exhausted", "budget"))
				return nil
			}
		}
	}
	return nil
}

func buildMeaningCandidate(input CompositionInput, alternative parser.ParseAlternative, predicate parser.PhraseCandidate, predicateMapping transfer, frame valency.ValencyFrame, binding valency.ConstructionBinding, roles []roleOption, references map[string]ReferenceEvidence, attributes map[string]NodeAttributesEvidence, partialStatus status.OutcomeKind, partialReason string, phraseNodes map[string]parser.NodeID, sourceBytes int, budget Budget) (MeaningCandidate, []EvidenceRecord, []valency.SoftConstraintMiss, status.OutcomeKind, []status.Diagnostic, error) {
	id := semanticAlternativeID(alternative, predicate.ID, predicateMapping.binding.CoreRef, roles)
	nodes := []meaning.Node{}
	participants := []meaning.Participant{}
	refs := meaning.RequestContext{ID: input.RequestContextID, Referents: []meaning.Referent{}, Mentions: []meaning.Mention{}}
	refNodeByIdentity := map[string]meaning.NodeID{}
	refNodeIndexByIdentity := map[string]int{}
	refConceptByIdentity := map[string]ids.ConceptID{}
	records := []EvidenceRecord{{SourceSpan: predicate.Span, CandidateID: predicate.ID, PhraseNodeID: phraseNodes[predicate.ID], ReadingID: firstString(predicate.ReadingIDs), SenseID: predicateMapping.sense, CoreConceptID: predicateMapping.binding.CoreRef, BindingKey: BindingKeyFor(predicateMapping.sense, predicateMapping.binding.CoreRef), PlanID: alternative.PlanID, FamilyID: alternative.FamilyID, VariantID: alternative.VariantID, FrameID: alternative.Permission.FrameID, Outcome: status.OutcomeOK, Features: evidenceFeatures(predicate.Features), Provenance: predicateMapping.binding.Provenance}}
	for index, role := range roles {
		identity := "candidate:" + role.candidate.ID
		refEvidence, resolved := references[role.candidate.ID]
		resolved = resolved && refEvidence.ReferentID != ""
		if resolved {
			identity = "referent:" + string(refEvidence.ReferentID)
		}
		nodeID, exists := refNodeByIdentity[identity]
		if exists && refConceptByIdentity[identity] != role.mapping.binding.CoreRef {
			diagnostic := semanticDiagnostic("REFERENT_CONCEPT_CONFLICT", "explicitly co-referred mentions have incompatible CORE concepts", "references")
			return evidenceOnlyMeaningCandidate(id, status.OutcomeAmbiguous, records, diagnostic), records, nil, status.OutcomeAmbiguous, []status.Diagnostic{diagnostic}, nil
		}
		if !exists {
			nodeID = meaning.NodeID("ref" + strconv.Itoa(index))
			node, err := meaning.NewReferent(nodeID, role.mapping.binding.CoreRef)
			if err != nil {
				return MeaningCandidate{}, nil, nil, status.OutcomeInvalid, nil, err
			}
			refNodeIndexByIdentity[identity] = len(nodes)
			nodes = append(nodes, node)
			refNodeByIdentity[identity] = nodeID
			refConceptByIdentity[identity] = role.mapping.binding.CoreRef
			if resolved {
				if refs.ID == "" {
					refs.ID = input.RequestContextID
				}
				refs.Referents = append(refs.Referents, meaning.Referent{ID: refEvidence.ReferentID, Concept: meaning.ConceptRefFromID(role.mapping.binding.CoreRef)})
			}
		}
		if attribute, supplied := attributes[role.candidate.ID]; supplied && attribute.Multiplicity != "" {
			nodeIndex := refNodeIndexByIdentity[identity]
			current := nodes[nodeIndex].Multiplicity
			if current != "" && current != attribute.Multiplicity {
				diagnostic := semanticDiagnostic("REFERENT_MULTIPLICITY_CONFLICT", "explicitly co-referred mentions have conflicting multiplicity evidence", "node_attributes")
				return evidenceOnlyMeaningCandidate(id, status.OutcomeAmbiguous, records, diagnostic), records, nil, status.OutcomeAmbiguous, []status.Diagnostic{diagnostic}, nil
			}
			nodes[nodeIndex].Multiplicity = attribute.Multiplicity
		}
		mention := meaning.Mention{ID: meaning.MentionID("m" + strconv.Itoa(index)), NodeID: nodeID}
		if resolved {
			mention.ReferentID = refEvidence.ReferentID
			mention.EvidenceID = refEvidence.EvidenceID
		}
		if refEvidence.Anchor != nil {
			anchor := *refEvidence.Anchor
			mention.Anchor = &anchor
		}
		refs.Mentions = append(refs.Mentions, mention)
		participants = append(participants, meaning.Participant{Role: role.role, Target: nodeID})
		records = append(records, EvidenceRecord{SourceSpan: role.span, CandidateID: role.candidate.ID, PhraseNodeID: phraseNodes[role.candidate.ID], ReadingID: firstString(role.candidate.ReadingIDs), SenseID: role.mapping.sense, CoreConceptID: role.mapping.binding.CoreRef, BindingKey: BindingKeyFor(role.mapping.sense, role.mapping.binding.CoreRef), Role: role.role, Outcome: status.OutcomeOK, Features: evidenceFeatures(role.candidate.Features), Provenance: role.mapping.binding.Provenance})
	}
	kind := meaning.KindEvent
	if concept, ok := input.Core.Lookup(predicateMapping.binding.CoreRef); !ok {
		diagnostic := semanticDiagnostic("UNKNOWN_PREDICATE_CONCEPT", "approved binding points to a CORE concept absent from the graph", "predicate")
		return evidenceOnlyMeaningCandidate(id, status.OutcomeUnknown, records, diagnostic), records, nil, status.OutcomeUnknown, []status.Diagnostic{diagnostic}, nil
	} else if concept.Kind == core.KindProperty {
		kind = meaning.KindState
	} else if concept.Kind != core.KindEvent {
		diagnostic := semanticDiagnostic("PREDICATE_KIND_UNSUPPORTED", "CORE concept kind does not license an event or state Meaning node", "predicate")
		return evidenceOnlyMeaningCandidate(id, status.OutcomeUnsupported, records, diagnostic), records, nil, status.OutcomeUnsupported, []status.Diagnostic{diagnostic}, nil
	}
	event, err := meaning.NewEvent("event", predicateMapping.binding.CoreRef, participants)
	if kind == meaning.KindState {
		event, err = meaning.NewState("event", predicateMapping.binding.CoreRef, participants)
	}
	if err != nil {
		return MeaningCandidate{}, nil, nil, status.OutcomeInvalid, nil, err
	}
	if attribute, supplied := attributes[predicate.ID]; supplied {
		event.TemporalFrame = attribute.TemporalFrame
	}
	nodes = append(nodes, event)
	operators, operatorNodes, opRecords, opOutcome, opDiagnostics := materializeOperators(input, alternative, predicate, sourceBytes)
	nodes = append(nodes, operatorNodes...)
	records = append(records, opRecords...)
	graphNodes := append([]meaning.Node(nil), nodes...)
	if len(nodes) == 0 {
		return MeaningCandidate{}, nil, nil, status.OutcomeInvalid, nil, errors.New("empty semantic graph")
	}
	roots := []meaning.NodeID{"event"}
	for _, node := range operatorNodes {
		if node.ID != "event" {
			roots = append(roots, node.ID)
		}
	}
	if len(input.Operators) == 1 && len(operatorNodes) == 1 &&
		operatorNodes[0].Kind == meaning.KindScope && operatorNodes[0].Operator == meaning.OperatorNegation && operatorNodes[0].Target == "event" {
		roots = []meaning.NodeID{operatorNodes[0].ID}
	}
	graph, err := meaning.Build(roots, graphNodes, budget.Meaning)
	if err != nil {
		outcome, diagnostics := errorOutcome(err)
		return MeaningCandidate{ID: id, SemanticID: id, Meaning: nil, Status: outcome, Evidence: EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: records, Diagnostics: diagnostics}}, records, nil, outcome, diagnostics, nil
	}
	if err := meaning.ValidateMeaning(graph, input.Core, budget.Meaning); err != nil {
		outcome, diagnostics := errorOutcome(err)
		return MeaningCandidate{ID: id, SemanticID: id, Meaning: &graph, Contexts: []MeaningContext{{Meaning: graph, Context: refs}}, Status: outcome, Evidence: EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: records, Diagnostics: diagnostics}}, records, nil, outcome, diagnostics, nil
	}
	if refs.ID != "" {
		if err := meaning.ValidateReferences(graph, refs, budget.Meaning); err != nil {
			outcome, diagnostics := errorOutcome(err)
			return MeaningCandidate{ID: id, SemanticID: id, Meaning: &graph, Contexts: []MeaningContext{{Meaning: graph, Context: refs}}, Status: outcome, Evidence: EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: records, Diagnostics: diagnostics}}, records, nil, outcome, diagnostics, nil
		}
	}
	validation, frameErr := valency.ValidateFrame(frame, valency.ValidationRequest{Meaning: graph, NodeID: "event", SemanticBinding: predicateMapping.binding, ConstructionBinding: binding}, input.Core)
	outcome := status.OutcomeOK
	diagnostics := []status.Diagnostic{}
	if frameErr != nil {
		outcome, diagnostics = errorOutcome(frameErr)
	}
	if partialStatus != "" && partialStatus != status.OutcomeOK {
		outcome = mergeSemanticOutcome(outcome, partialStatus)
		diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("PARTIAL_PARSE_EVIDENCE", partialReason, "parse"))
	}
	if opOutcome != "" && opOutcome != status.OutcomeOK {
		outcome = mergeSemanticOutcome(outcome, opOutcome)
		diagnostics = appendDiagnostics(diagnostics, opDiagnostics...)
	}
	if len(input.Operators) > 1 {
		// Meaning v1 has scoped wrappers but no representation for the order
		// between multiple operators. Preserve the scopes and construction plan
		// evidence, while withholding a complete interpretation.
		outcome = mergeSemanticOutcome(outcome, status.OutcomeIncomplete)
		diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("OPERATOR_ORDER_NOT_REPRESENTED", "multiple explicit operators have no ordered scope representation in Meaning v1", "operators"))
	}
	if !input.OperatorCoverageComplete {
		outcome = mergeSemanticOutcome(outcome, status.OutcomeIncomplete)
		diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("OPERATOR_COVERAGE_UNKNOWN", "operator evidence coverage was not declared complete", "operators"))
	}
	if len(operators) > 0 {
		if _, err := construction.Compose(input.Constructions, frame, alternative.FamilyID, graph, operators); err != nil {
			opStatus, opDiags := errorOutcome(err)
			outcome = mergeSemanticOutcome(outcome, opStatus)
			diagnostics = appendDiagnostics(diagnostics, opDiags...)
		}
	}
	diagnostics = appendDiagnostics(diagnostics, frameDiagnostic(validation.SoftConstraintMisses)...)
	semanticID, identityErr := semanticGraphID(graph, refs, id)
	if identityErr != nil {
		identityOutcome, identityDiagnostics := errorOutcome(identityErr)
		outcome = mergeSemanticOutcome(outcome, identityOutcome)
		diagnostics = appendDiagnostics(diagnostics, identityDiagnostics...)
	}
	candidate := MeaningCandidate{ID: id, SemanticID: semanticID, Meaning: &graph, Contexts: []MeaningContext{{Meaning: graph, Context: refs}}, Status: outcome, Evidence: EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: append([]EvidenceRecord(nil), records...), Diagnostics: append([]status.Diagnostic(nil), diagnostics...)}, SoftConstraintMisses: append([]valency.SoftConstraintMiss(nil), validation.SoftConstraintMisses...), DerivationIDs: []string{alternative.DerivationID}}
	if err := candidate.Evidence.ValidateWithBudget(budget.Evidence); err != nil {
		return MeaningCandidate{}, nil, nil, "", nil, err
	}
	return candidate, records, validation.SoftConstraintMisses, outcome, diagnostics, nil
}

func materializeOperators(input CompositionInput, alternative parser.ParseAlternative, predicate parser.PhraseCandidate, sourceBytes int) ([]construction.Operator, []meaning.Node, []EvidenceRecord, status.OutcomeKind, []status.Diagnostic) {
	operators := []construction.Operator{}
	nodes := []meaning.Node{}
	records := []EvidenceRecord{}
	outcome := status.OutcomeOK
	diagnostics := []status.Diagnostic{}
	for index, evidence := range input.Operators {
		if evidence.Span.Start < 0 || evidence.Span.End <= evidence.Span.Start || evidence.Span.End > sourceBytes {
			outcome = mergeSemanticOutcome(outcome, status.OutcomeUnsupported)
			diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("INVALID_OPERATOR_SPAN", "operator evidence must have a bounded non-empty source span", "operators"))
			continue
		}
		if evidence.TargetCandidateID != predicate.ID {
			outcome = mergeSemanticOutcome(outcome, status.OutcomeUnsupported)
			diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("OPERATOR_TARGET_UNSUPPORTED", "operator target is not the selected predicate candidate", "operators"))
			continue
		}
		op := evidence.Operator
		op.SemanticScope.TargetNodeID = "event"
		scopeID := meaning.NodeID("scope" + strconv.Itoa(index))
		op.SemanticScope.ScopeNodeID = scopeID
		var node meaning.Node
		var err error
		switch op.Kind {
		case construction.OperatorNegation:
			if op.Value != construction.ValueNegated {
				err = errors.New("unsupported negation value")
			} else {
				node, err = meaning.NewNegation(scopeID, "event")
			}
		case construction.OperatorQuestion:
			outcome = mergeSemanticOutcome(outcome, status.OutcomeUnsupported)
			reason := "question subtype has no explicit representation in Meaning v1"
			diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("QUESTION_TYPE_UNSUPPORTED", reason, "operators"))
			records = append(records, EvidenceRecord{SourceSpan: evidence.Span, CandidateID: predicate.ID, PlanID: alternative.PlanID, FamilyID: alternative.FamilyID, VariantID: alternative.VariantID, FrameID: alternative.Permission.FrameID, Outcome: status.OutcomeUnsupported, Reason: reason})
			continue
		case construction.OperatorModality:
			modality := meaning.Modality(op.Value)
			node, err = meaning.NewModalityNode(scopeID, modality, "event")
		default:
			outcome = mergeSemanticOutcome(outcome, status.OutcomeUnsupported)
			reason := "explicit construction operator cannot be represented in Meaning v1"
			diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("OPERATOR_NOT_IN_MEANING_V1", reason, "operators"))
			span := evidence.Span
			if span.End <= span.Start {
				span = predicate.Span
			}
			records = append(records, EvidenceRecord{SourceSpan: span, CandidateID: predicate.ID, PlanID: alternative.PlanID, FamilyID: alternative.FamilyID, VariantID: alternative.VariantID, FrameID: alternative.Permission.FrameID, Outcome: status.OutcomeUnsupported, Reason: reason})
			continue
		}
		if err != nil {
			outcome = mergeSemanticOutcome(outcome, status.OutcomeUnsupported)
			diagnostics = appendDiagnostic(diagnostics, semanticDiagnostic("INVALID_OPERATOR_EVIDENCE", err.Error(), "operators"))
			continue
		}
		operators = append(operators, op)
		nodes = append(nodes, node)
		span := evidence.Span
		if span.End <= span.Start {
			span = predicate.Span
		}
		records = append(records, EvidenceRecord{SourceSpan: span, CandidateID: predicate.ID, PlanID: alternative.PlanID, FamilyID: alternative.FamilyID, VariantID: alternative.VariantID, FrameID: alternative.Permission.FrameID, Outcome: status.OutcomeOK, Reason: "explicit operator evidence supplied by caller", Provenance: op.Provenance})
	}
	return operators, nodes, records, outcome, diagnostics
}

func resolveRoles(input CompositionInput, alternative parser.ParseAlternative, candidates map[string]parser.PhraseCandidate) ([][]roleOption, status.OutcomeKind, string) {
	options := [][]roleOption{}
	var unresolvedReason string
	for _, slot := range alternative.Slots {
		if slot.Role == "" || len(slot.CandidateIDs) == 0 {
			continue
		}
		if len(slot.CandidateIDs) != 1 {
			return nil, status.OutcomeIncomplete, "role slot has multiple candidates; Meaning v1 has no group participant representation"
		}
		candidate, ok := candidates[slot.CandidateIDs[0]]
		if !ok {
			return nil, status.OutcomeUnknown, "role candidate does not resolve"
		}
		if candidate.Category.IsPredicate() {
			continue
		}
		mappings := transferMappings(input.Lexicon, candidate, nil)
		if len(mappings) == 0 {
			return nil, status.OutcomeUnsupported, "role candidate has no approved equivalent Sense-to-CORE binding"
		}
		mappedSenses := map[ids.SenseID]bool{}
		for _, mapping := range mappings {
			mappedSenses[mapping.sense] = true
		}
		for _, sense := range candidate.SenseIDs {
			if !mappedSenses[sense] {
				// Keep the approved readings available while retaining this competing
				// sense as unresolved evidence on the partial Meaning candidate.
				unresolvedReason = UnresolvedRoleSenseReason
			}
		}
		roleChoices := make([]roleOption, 0, len(mappings))
		for _, mapping := range mappings {
			roleChoices = append(roleChoices, roleOption{role: slot.Role, candidate: candidate, mapping: mapping, span: candidate.Span})
		}
		options = append(options, roleChoices)
	}
	if unresolvedReason != "" {
		return options, status.OutcomeIncomplete, unresolvedReason
	}
	return options, status.OutcomeOK, ""
}

func mergePartialReason(existing, next string) string {
	if next == "" || existing == next {
		return existing
	}
	if existing == "" {
		return next
	}
	return existing + "; " + next
}

func unmappedRoleEvidence(alternative parser.ParseAlternative, candidates map[string]parser.PhraseCandidate, index *lexicon.Index, reason string) []EvidenceRecord {
	records := []EvidenceRecord{}
	for _, slot := range alternative.Slots {
		if slot.Role == "" {
			continue
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, exists := candidates[candidateID]
			if !exists || candidate.Span.End <= candidate.Span.Start {
				continue
			}
			mapped := map[ids.SenseID]bool{}
			for _, sense := range candidate.SenseIDs {
				if len(index.TransferBindings(sense)) > 0 {
					mapped[sense] = true
				}
			}
			for _, sense := range candidate.SenseIDs {
				if mapped[sense] {
					continue
				}
				records = append(records, EvidenceRecord{SourceSpan: candidate.Span, CandidateID: candidate.ID, ReadingID: firstString(candidate.ReadingIDs), SenseID: sense, Role: slot.Role, Outcome: status.OutcomeUnsupported, Reason: reason})
			}
		}
	}
	return records
}

func roleCombinations(options [][]roleOption, limit int) [][]roleOption {
	if limit <= 0 {
		return nil
	}
	result := [][]roleOption{{}}
	for _, choices := range options {
		next := make([][]roleOption, 0)
		for _, prefix := range result {
			for _, choice := range choices {
				if len(next) >= limit {
					return nil
				}
				copyPrefix := append([]roleOption(nil), prefix...)
				next = append(next, append(copyPrefix, choice))
			}
		}
		result = next
	}
	return result
}

func transferMappings(index *lexicon.Index, candidate parser.PhraseCandidate, only *ids.SenseID) []transfer {
	result := []transfer{}
	seen := map[string]bool{}
	for _, sense := range candidate.SenseIDs {
		if only != nil && sense != *only {
			continue
		}
		for _, binding := range index.TransferBindings(sense) {
			key := sense.String() + "\x00" + binding.CoreRef.String()
			if !seen[key] {
				result = append(result, transfer{binding: binding, sense: sense})
				seen[key] = true
			}
		}
	}
	return result
}

func exactFrameBinding(input CompositionInput, alternative parser.ParseAlternative) (valency.ValencyFrame, valency.ConstructionBinding, bool) {
	frames := input.Valency.FramesBySenseID(alternative.Permission.SenseID)
	var selected valency.ValencyFrame
	found := false
	for _, frame := range frames {
		if frame.ID == alternative.Permission.FrameID {
			selected = frame
			found = true
			break
		}
	}
	if !found {
		return valency.ValencyFrame{}, valency.ConstructionBinding{}, false
	}
	for _, binding := range input.Valency.BindingsByFamilyID(alternative.FamilyID) {
		if binding.SenseID == alternative.Permission.SenseID && binding.ValencyFrameID == alternative.Permission.FrameID && binding.FamilyID == alternative.FamilyID && binding.VariantID.String() == string(alternative.VariantID) && binding.Approved {
			return selected, binding, true
		}
	}
	return valency.ValencyFrame{}, valency.ConstructionBinding{}, false
}

func appendAlternativeEvidenceOnly(result *ComposeResult, input CompositionInput, alternative parser.ParseAlternative, span text.Span, outcome status.OutcomeKind, code, reason string, records []EvidenceRecord, budget Budget) {
	id := alternativeID(alternative, code, reason)
	diagnostic := semanticDiagnostic(code, reason, "alternative")
	trace := EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: append([]EvidenceRecord(nil), records...), Diagnostics: []status.Diagnostic{diagnostic}}
	trace.Records = append(trace.Records, EvidenceRecord{SourceSpan: span, PlanID: alternative.PlanID, FamilyID: alternative.FamilyID, VariantID: alternative.VariantID, FrameID: alternative.Permission.FrameID, Outcome: outcome, Reason: reason})
	if err := trace.ValidateWithBudget(budget.Evidence); err != nil {
		kind, diagnostics := errorOutcome(err)
		result.Outcome = mergeSemanticOutcome(result.Outcome, kind)
		result.Diagnostics = appendDiagnostics(result.Diagnostics, diagnostics...)
		return
	}
	result.Candidates = append(result.Candidates, MeaningCandidate{ID: id, SemanticID: id, Meaning: nil, Status: outcome, Evidence: trace, DerivationIDs: []string{alternative.DerivationID}})
	result.Outcome = mergeSemanticOutcome(result.Outcome, outcome)
	result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic)
}

func appendEvidenceOnly(result *ComposeResult, input CompositionInput, candidate parser.PhraseCandidate, outcome status.OutcomeKind, code, reason string, budget Budget) {
	if candidate.Span.End <= candidate.Span.Start {
		result.Outcome = mergeSemanticOutcome(result.Outcome, status.OutcomeInvalid)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("INVALID_EVIDENCE_SPAN", "unmatched candidate has no valid source span, so its evidence cannot be retained", "candidate"))
		return
	}
	id := "evidence:" + stableHash(candidate.ID+"\x00"+code+"\x00"+reason)
	records := []EvidenceRecord{}
	if len(candidate.SenseIDs) == 0 {
		records = append(records, EvidenceRecord{SourceSpan: candidate.Span, CandidateID: candidate.ID, ReadingID: firstString(candidate.ReadingIDs), Outcome: outcome, Reason: reason})
	} else {
		for _, sense := range candidate.SenseIDs {
			records = append(records, EvidenceRecord{SourceSpan: candidate.Span, CandidateID: candidate.ID, ReadingID: firstString(candidate.ReadingIDs), SenseID: sense, Outcome: outcome, Reason: reason})
		}
	}
	trace := EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: records, Diagnostics: []status.Diagnostic{semanticDiagnostic(code, reason, "candidate")}}
	if trace.ValidateWithBudget(budget.Evidence) != nil {
		result.Outcome = mergeSemanticOutcome(result.Outcome, status.OutcomeResourceLimit)
		result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("RESOURCE_LIMIT", "candidate evidence could not be retained within the evidence budget", "evidence"))
		return
	}
	result.Candidates = append(result.Candidates, MeaningCandidate{ID: id, SemanticID: id, Status: outcome, Evidence: trace})
	result.Outcome = mergeSemanticOutcome(result.Outcome, outcome)
	result.Diagnostics = appendDiagnostic(result.Diagnostics, trace.Diagnostics[0])
}

func finalizeComposition(result ComposeResult, ontology *core.ConceptGraph, budget Budget) (ComposeResult, error) {
	result.Diagnostics = trimDiagnostics(result.Diagnostics)
	if len(result.Candidates) == 0 && result.Outcome == status.OutcomeOK {
		result.Outcome = status.OutcomeUnknown
		result.Diagnostics = appendDiagnostic(result.Diagnostics, semanticDiagnostic("NO_SEMANTIC_ALTERNATIVES", "no semantic graph or evidence-only candidate survived composition", "result"))
	}
	if len(result.Candidates) > 0 {
		resolved, err := Disambiguate(result.Candidates, result.Outcome, result.Diagnostics, ontology, budget.Evidence)
		if err != nil {
			return ComposeResult{}, err
		}
		result.Outcome = resolved.Outcome
		result.Candidates = resolved.Candidates
		result.Diagnostics = resolved.Diagnostics
	}
	if result.Outcome != status.OutcomeOK && len(result.Diagnostics) == 0 {
		result.Diagnostics = []status.Diagnostic{semanticDiagnostic("SEMANTIC_RESULT_INCOMPLETE", "semantic composition did not establish a complete unique meaning", "result")}
	}
	return result, nil
}

func errorOutcome(err error) (status.OutcomeKind, []status.Diagnostic) {
	type resultError interface{ Result() status.Result }
	var failure resultError
	if errors.As(err, &failure) {
		r := failure.Result()
		if r.Status.Valid() && r.Status != "" {
			return r.Status, append([]status.Diagnostic(nil), r.Diagnostics...)
		}
	}
	return status.OutcomeError, []status.Diagnostic{semanticDiagnostic("SEMANTIC_VALIDATION_ERROR", err.Error(), "meaning")}
}

func frameDiagnostic(misses []valency.SoftConstraintMiss) []status.Diagnostic {
	out := []status.Diagnostic{}
	for _, miss := range misses {
		out = append(out, semanticDiagnostic("SOFT_SELECTIONAL_MISS", "soft preference mismatch retained without rejecting the candidate", string(miss.Role)))
	}
	return out
}
func semanticError(kind status.OutcomeKind, code, path, message string) error {
	return &composeError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{semanticDiagnostic(code, message, path)}}}
}

type composeError struct{ result status.Result }

func (e *composeError) Error() string {
	if e == nil || len(e.result.Diagnostics) == 0 {
		return "semantic composition error"
	}
	return e.result.Diagnostics[0].Message
}
func (e *composeError) Result() status.Result {
	if e == nil {
		return status.Result{}
	}
	r := e.result
	r.Diagnostics = append([]status.Diagnostic(nil), r.Diagnostics...)
	return r
}
func semanticDiagnostic(code, message, path string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning}
}
func appendDiagnostic(ds []status.Diagnostic, d status.Diagnostic) []status.Diagnostic {
	for _, old := range ds {
		if old.Code == d.Code && old.Path == d.Path && old.Message == d.Message {
			return ds
		}
	}
	if len(ds) >= status.MaxDiagnostics {
		return ds
	}
	return append(ds, d)
}
func appendDiagnostics(ds []status.Diagnostic, more ...status.Diagnostic) []status.Diagnostic {
	for _, d := range more {
		ds = appendDiagnostic(ds, d)
	}
	return ds
}
func trimDiagnostics(ds []status.Diagnostic) []status.Diagnostic {
	if len(ds) > status.MaxDiagnostics {
		return append([]status.Diagnostic(nil), ds[:status.MaxDiagnostics]...)
	}
	return append([]status.Diagnostic(nil), ds...)
}

func evidenceFeatures(values map[features.FeatureID]features.FeatureValueID) []features.FeatureValue {
	if len(values) == 0 {
		return nil
	}
	result := make([]features.FeatureValue, 0, len(values))
	for feature, value := range values {
		result = append(result, features.FeatureValue{Feature: feature, State: features.StateKnown, Value: value})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].Feature < result[right].Feature })
	return result
}

func mergeSemanticOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	if left == "" {
		return right
	}
	if right == "" || left == right {
		return left
	}
	rank := func(v status.OutcomeKind) int {
		switch v {
		case status.OutcomeInvalid:
			return 8
		case status.OutcomeError:
			return 7
		case status.OutcomeResourceLimit:
			return 6
		case status.OutcomeUnsupported:
			return 5
		case status.OutcomeAmbiguous:
			return 4
		case status.OutcomeIncomplete:
			return 3
		case status.OutcomeUnknown:
			return 2
		case status.OutcomeOK:
			return 1
		}
		return 0
	}
	if rank(right) > rank(left) {
		return right
	}
	return left
}

func firstSourceSpan(l analyzer.LexicalLattice) text.Span {
	if len(l.Tokens) > 0 {
		return l.Tokens[0].Token.Span
	}
	return text.Span{}
}
func containsSense(values []ids.SenseID, needle ids.SenseID) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
func firstString(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
func alternativeID(a parser.ParseAlternative, parts ...string) string {
	return "alt:" + stableHash(strings.Join(append([]string{a.DerivationID, string(a.PlanID), a.Permission.SenseID.String(), string(a.Permission.FrameID)}, parts...), "\x00"))
}
func semanticAlternativeID(a parser.ParseAlternative, predicate string, concept ids.ConceptID, roles []roleOption) string {
	parts := []string{predicate, concept.String()}
	for _, role := range roles {
		parts = append(parts, string(role.role), role.candidate.ID, role.mapping.sense.String(), role.mapping.binding.CoreRef.String())
	}
	return alternativeID(a, parts...)
}
func semanticGraphID(graph meaning.Meaning, context meaning.RequestContext, alternativeID string) (string, error) {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	referents := make(map[meaning.NodeID]meaning.ReferentID)
	for _, mention := range context.Mentions {
		if mention.ReferentID != "" {
			referents[mention.NodeID] = mention.ReferentID
		}
	}
	memo := make(map[meaning.NodeID]string, len(nodes))
	var signature func(meaning.NodeID) (string, error)
	signature = func(id meaning.NodeID) (string, error) {
		if value, ok := memo[id]; ok {
			return value, nil
		}
		node, ok := nodes[id]
		if !ok {
			return "", fmt.Errorf("semantic graph target %q is absent", id)
		}
		parts := []string{string(node.Kind), string(node.Concept), string(node.Operator), string(node.Modality), string(node.Act)}
		if node.TemporalFrame != "" {
			parts = append(parts, "temporal_frame", string(node.TemporalFrame))
		}
		if node.Multiplicity != "" {
			parts = append(parts, "multiplicity", string(node.Multiplicity))
		}
		switch node.Kind {
		case meaning.KindReferent:
			if referent := referents[id]; referent != "" {
				parts = append(parts, "identity", string(referent))
			} else {
				parts = append(parts, "unresolved", alternativeID)
			}
		case meaning.KindEvent, meaning.KindState:
			participants := append([]meaning.Participant(nil), node.Participants...)
			sort.Slice(participants, func(i, j int) bool { return participants[i].Role < participants[j].Role })
			for _, participant := range participants {
				child, err := signature(participant.Target)
				if err != nil {
					return "", err
				}
				parts = append(parts, string(participant.Role), child)
			}
		case meaning.KindDiscourse:
			for _, target := range node.Content {
				child, err := signature(target)
				if err != nil {
					return "", err
				}
				parts = append(parts, child)
			}
		default:
			if node.Target != "" {
				child, err := signature(node.Target)
				if err != nil {
					return "", err
				}
				parts = append(parts, child)
			}
			if node.Quantifier != nil {
				parts = append(parts, string(node.Quantifier.Kind))
				if node.Quantifier.Count != nil {
					parts = append(parts, strconv.Itoa(*node.Quantifier.Count))
				}
			}
		}
		value := stableHash(strings.Join(parts, "\x00"))
		memo[id] = value
		return value, nil
	}
	rootSignatures := make([]string, 0, len(graph.Roots))
	for _, root := range graph.Roots {
		value, err := signature(root)
		if err != nil {
			return "", err
		}
		rootSignatures = append(rootSignatures, value)
	}
	sort.Strings(rootSignatures)
	return "meaning:" + stableHash(context.ID+"\x00"+strings.Join(rootSignatures, "\x00")), nil
}

func evidenceOnlyMeaningCandidate(id string, outcome status.OutcomeKind, records []EvidenceRecord, diagnostic status.Diagnostic) MeaningCandidate {
	if id == "" {
		id = "evidence:" + stableHash(diagnostic.Code+"\x00"+diagnostic.Message)
	}
	trace := EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: append([]EvidenceRecord(nil), records...), Diagnostics: []status.Diagnostic{diagnostic}}
	return MeaningCandidate{ID: id, SemanticID: id, Status: outcome, Evidence: trace}
}

func stableHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

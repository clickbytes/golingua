package semantic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/synthesis"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

// ComposeDSE composes every retained parse and lexical reading independently.
// It uses only the explicitly selected language package, source plan, algebra
// declarations, and request-local identity/context supplied by the caller.
func ComposeDSE(input DSEInput) (DSEResult, error) {
	budget, err := NormalizeDSEBudget(input.Budget)
	if err != nil {
		return DSEResult{}, err
	}
	if input.Context == nil {
		return DSEResult{}, dseError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "DSE composition requires a request context")
	}
	if err := input.Context.Err(); err != nil {
		return DSEResult{}, dseError(status.OutcomeResourceLimit, "DSE_CANCELED", "context", "DSE composition was canceled or timed out")
	}
	if input.Package == nil || input.Plan == nil {
		return DSEResult{}, dseError(status.OutcomeInvalid, "MISSING_SELECTED_SOURCE_METADATA", "metadata", "a selected semantic language package and source plan are required")
	}
	if input.Selected.SchemaVersion != algebra.SchemaVersion || input.Selected.Expression != nil || input.Selected.RequestContext != nil || len(input.Selected.Parameters) != 0 || len(input.Selected.References) != 0 {
		return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_SELECTED_DECLARATIONS", "selected", "selected algebra input must contain declarations only")
	}
	if err := algebra.CheckDeclarations(input.Selected); err != nil {
		return DSEResult{}, err
	}
	if input.Plan.Language() != string(input.Package.Language()) || input.Plan.Language() != string(input.Language) {
		return DSEResult{}, dseError(status.OutcomeInvalid, "SOURCE_LANGUAGE_MISMATCH", "language", "plan, package, and analyzed forest languages must match exactly")
	}
	if input.ReferenceContext.SchemaVersion != 1 || input.ReferenceContext.ContextID == "" {
		return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_SOURCE_REFERENCE_CONTEXT", "reference_context", "an explicit validated source reference context is required")
	}
	if input.ReferenceContext.RequestContext != nil && input.ReferenceContext.RequestContext.ID != input.ReferenceContext.ContextID {
		return DSEResult{}, dseError(status.OutcomeInvalid, "REQUEST_CONTEXT_ID_MISMATCH", "reference_context.request_context.id", "request context ID must match the explicit source context ID")
	}
	if len(input.Forest.Nodes) > parser.DefaultBudget().MaxNodes || len(input.Forest.Roots) > parser.DefaultBudget().MaxNodes || len(input.Candidates) > parser.DefaultBudget().MaxExpansions || len(input.Lattice.Tokens) > status.DefaultBounds().MaxInputBytes || len(input.ReferenceContext.References) > algebra.DefaultMaxSymbols || len(input.Operators) > construction.MaxOperatorsPerPlan {
		return DSEResult{}, dseError(status.OutcomeResourceLimit, "DSE_INPUT_LIMIT", "input", "source evidence exceeds the finite DSE profile bounds")
	}
	if !input.Forest.Outcome.Valid() || input.SourceBytes < 0 || input.SourceBytes > status.DefaultBounds().MaxInputBytes {
		return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_SOURCE_FOREST", "forest", "forest outcome or source byte count is invalid")
	}
	if len(input.Forest.Diagnostics) > status.MaxDiagnostics {
		return DSEResult{}, dseError(status.OutcomeResourceLimit, "FOREST_DIAGNOSTIC_LIMIT", "forest.diagnostics", "parser diagnostic count exceeds its finite cap")
	}
	for i, diagnostic := range input.Forest.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_FOREST_DIAGNOSTIC", fmt.Sprintf("forest.diagnostics[%d]", i), err.Error())
		}
	}
	if input.Forest.Outcome == status.OutcomeAmbiguous {
		if resolved, ok := dseResolveExplicitWHGapOmissions(input.Forest); ok {
			input.Forest = resolved
		}
	}
	if input.Forest.Outcome != status.OutcomeOK && input.Forest.Outcome != status.OutcomeIncomplete && input.Forest.Outcome != status.OutcomeResourceLimit {
		return DSEResult{}, dseError(status.OutcomeUnsupported, "UNSUPPORTED_SOURCE_FOREST", "forest.outcome", "only valid, incomplete, or resource-limited forests can be composed")
	}
	if input.Forest.Outcome == status.OutcomeResourceLimit {
		return DSEResult{Status: status.OutcomeResourceLimit, Complete: false, Candidates: []DSECandidate{}, Residuals: []DSEResidual{}, Diagnostics: []status.Diagnostic{}, WholeVerificationScope: "type_scope_context_only"}, nil
	}
	if err := validateDSEContext(input); err != nil {
		return DSEResult{}, err
	}
	mentionRefs, err := dseMentionReferences(input)
	if err != nil {
		return DSEResult{}, err
	}
	budgetContext, cancel := context.WithCancel(input.Context)
	defer cancel()
	input.Context = budgetContext

	result := DSEResult{Status: status.OutcomeOK, Complete: input.Forest.Outcome == status.OutcomeOK || dseForestIsIncompleteOnlyForWHGaps(input.Forest), Candidates: []DSECandidate{}, Residuals: []DSEResidual{}, Diagnostics: []status.Diagnostic{}, WholeVerificationScope: "type_scope_context_only"}
	candidates := make(map[string]parser.PhraseCandidate, len(input.Candidates))
	for i, candidate := range input.Candidates {
		if candidate.ID == "" || len(candidate.ID) > 256 || candidate.Span.Start < 0 || candidate.Span.End < candidate.Span.Start || candidate.Span.End > input.SourceBytes {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_PHRASE_CANDIDATE", fmt.Sprintf("candidates[%d]", i), "candidate ID or source span is invalid")
		}
		if _, duplicate := candidates[candidate.ID]; duplicate {
			return DSEResult{}, dseError(status.OutcomeInvalid, "DUPLICATE_PHRASE_CANDIDATE", fmt.Sprintf("candidates[%d].id", i), "candidate IDs must be unique")
		}
		candidates[candidate.ID] = candidate
	}
	attrs := make(map[string]NodeAttributesEvidence, len(input.NodeAttributes))
	for i, row := range input.NodeAttributes {
		candidate, ok := candidates[row.CandidateID]
		if !ok || row.CandidateID == "" {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_NODE_ATTRIBUTES_EVIDENCE", fmt.Sprintf("node_attributes[%d].candidate_id", i), "attributes must target an actual phrase candidate")
		}
		if _, duplicate := attrs[row.CandidateID]; duplicate {
			return DSEResult{}, dseError(status.OutcomeInvalid, "DUPLICATE_NODE_ATTRIBUTES_EVIDENCE", fmt.Sprintf("node_attributes[%d]", i), "candidate attributes may be supplied only once")
		}
		if row.TemporalFrame != "" && (!row.TemporalFrame.Valid() || !candidate.Category.IsPredicate()) {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_TEMPORAL_FRAME_EVIDENCE", fmt.Sprintf("node_attributes[%d]", i), "temporal frame must be supported and target a predicate candidate")
		}
		if row.Multiplicity != "" && (!row.Multiplicity.Valid() || candidate.Category != construction.CategoryNP) {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_MULTIPLICITY_EVIDENCE", fmt.Sprintf("node_attributes[%d]", i), "multiplicity must be supported and target an NP candidate")
		}
		if row.TemporalFrame == "" && row.Multiplicity == "" {
			return DSEResult{}, dseError(status.OutcomeInvalid, "EMPTY_NODE_ATTRIBUTES_EVIDENCE", fmt.Sprintf("node_attributes[%d]", i), "attribute record must contain a known value")
		}
		attrs[row.CandidateID] = row
	}
	referenceEvidence := make(map[string]ReferenceEvidence, len(input.References))
	for i, row := range input.References {
		if _, ok := candidates[row.CandidateID]; !ok || row.CandidateID == "" {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_REFERENCE_EVIDENCE", fmt.Sprintf("references[%d]", i), "reference evidence must target an actual phrase candidate")
		}
		if _, duplicate := referenceEvidence[row.CandidateID]; duplicate {
			return DSEResult{}, dseError(status.OutcomeInvalid, "DUPLICATE_REFERENCE_EVIDENCE", fmt.Sprintf("references[%d]", i), "candidate reference evidence must be unique")
		}
		resolved := row.ReferentID != "" && row.EvidenceID != ""
		unresolved := row.ReferentID == "" && row.EvidenceID == "" && row.Anchor != nil
		if (!resolved && !unresolved) || len(string(row.ReferentID)) > 256 || len(row.EvidenceID) > 256 || !utf8.ValidString(string(row.ReferentID)) || !utf8.ValidString(row.EvidenceID) {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_REFERENCE_EVIDENCE", fmt.Sprintf("references[%d]", i), "identity evidence requires both reference and evidence IDs, or an explicit unresolved anchor")
		}
		if row.Anchor != nil {
			anchorContext := meaning.RequestContext{ID: "dse-reference-validation", Mentions: []meaning.Mention{{ID: "mention", NodeID: "node", Anchor: row.Anchor}}}
			if err := anchorContext.Validate(); err != nil {
				return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_REFERENCE_ANCHOR", fmt.Sprintf("references[%d].anchor", i), err.Error())
			}
		}
		referenceEvidence[row.CandidateID] = row
	}
	rootByID := make(map[parser.NodeID]parser.PackedNode, len(input.Forest.Nodes))
	for i, node := range input.Forest.Nodes {
		if node.ID == "" {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_FOREST_NODE", fmt.Sprintf("forest.nodes[%d].id", i), "packed node ID must be non-empty")
		}
		if _, duplicate := rootByID[node.ID]; duplicate {
			return DSEResult{}, dseError(status.OutcomeInvalid, "DUPLICATE_FOREST_NODE", fmt.Sprintf("forest.nodes[%d].id", i), "packed node IDs must be unique")
		}
		rootByID[node.ID] = node
	}
	for _, unmatched := range input.Forest.Unmatched {
		if _, ok := candidates[unmatched.CandidateID]; !ok {
			return DSEResult{}, dseError(status.OutcomeInvalid, "UNKNOWN_UNMATCHED_CANDIDATE", "forest.unmatched", "unmatched candidate does not resolve")
		}
		if unmatched.TokenStart < 0 || unmatched.TokenEnd < unmatched.TokenStart || unmatched.TokenEnd > len(input.Lattice.Tokens) {
			return DSEResult{}, dseError(status.OutcomeInvalid, "INVALID_UNMATCHED_SPAN", "forest.unmatched", "unmatched token span is outside the analyzed lattice")
		}
		res := DSEResidual{Code: "UNMATCHED_SOURCE_CANDIDATE", CandidateID: unmatched.CandidateID, Path: fmt.Sprintf("tokens[%d:%d]", unmatched.TokenStart, unmatched.TokenEnd), Message: "candidate is outside retained complete parse interpretations"}
		result.Residuals = append(result.Residuals, res)
		result.Complete = false
	}
	work := 0
	resourceStopped := false
	resourceCode, resourcePath := "DSE_WORK_LIMIT", "forest"
	for rootIndex, rootID := range input.Forest.Roots {
		if resourceStopped {
			break
		}
		if err := input.Context.Err(); err != nil {
			return dseResourcePrefix(result, "DSE_CANCELED", "forest.roots", work), nil
		}
		root, ok := rootByID[rootID]
		if !ok {
			return DSEResult{}, dseError(status.OutcomeInvalid, "UNKNOWN_FOREST_ROOT", fmt.Sprintf("forest.roots[%d]", rootIndex), "root ID does not resolve to a packed node")
		}
		for altIndex, alternative := range root.Alternatives {
			if !chargeDSEWork(&result, &work, budget.MaxWork) {
				return dseResourcePrefix(result, "DSE_WORK_LIMIT", "forest", work), nil
			}
			if len(result.Candidates) >= budget.MaxAlternatives {
				return dseResourcePrefix(result, "DSE_ALTERNATIVE_LIMIT", "forest", work), nil
			}
			if len(alternative.Unresolved) != 0 && !dseUnresolvedSlotsAreWHGaps(alternative) {
				if appendDSEIncomplete(&result, fmt.Sprintf("root:%s:%d", rootID, altIndex), DSESourceRecord{DerivationID: alternative.DerivationID, Span: root.Span}, "UNRESOLVED_PARSE_SLOTS", "parse alternative has unresolved source slots", budget, work) {
					return result, nil
				}
				continue
			}
			if !dsePermissionMatches(input.Plan.Document(), alternative) {
				if appendDSEIncomplete(&result, fmt.Sprintf("root:%s:%d", rootID, altIndex), DSESourceRecord{DerivationID: alternative.DerivationID, FamilyID: alternative.FamilyID.String(), VariantID: alternative.VariantID.String(), Span: root.Span}, "SOURCE_PERMISSION_MISMATCH", "source plan does not authorize the exact sense, frame, family, and variant", budget, work) {
					return result, nil
				}
				continue
			}
			sourceMapping, hasSourceMapping := sourcePredicateMapping(input.Plan.Document(), alternative.Permission.SenseID.String(), string(alternative.Permission.FrameID))
			nominalRelationMapping := hasSourceMapping && sourceMapping.PredicateKind == "relation" &&
				containsConstructionID(sourceMapping.AllowedFamilyIDs, alternative.FamilyID) &&
				(len(sourceMapping.AllowedVariantIDs) == 0 || containsVariantID(sourceMapping.AllowedVariantIDs, alternative.VariantID))
			vpIDs := make([]string, 0, 2)
			roleSlots := make(map[string]parser.SlotBinding)
			for _, slot := range alternative.Slots {
				containsFrameRole := false
				for _, frameRole := range input.Package.Frames() {
					if frameRole.ID == alternative.Permission.FrameID {
						for _, role := range frameRole.SemanticRoles {
							if string(role) == string(slot.Role) {
								containsFrameRole = true
							}
						}
						break
					}
				}
				isPredicateSlot := false
				for _, id := range slot.CandidateIDs {
					if c, found := candidates[id]; found && c.Category.IsPredicate() {
						isPredicateSlot = true
					}
				}
				// The parser only compiles an unbound NP as the main predicate
				// when its exact source construction marks that slot
				// NominalPredicate. Preserve that distinction here: accept it only
				// for an explicitly mapped relation sense/frame and an exact
				// authorized family/variant, with one selected nominal head.
				if !isPredicateSlot && slot.Role == "" && nominalRelationMapping && len(slot.CandidateIDs) == 1 {
					if c, found := candidates[slot.CandidateIDs[0]]; found && c.Category == input.Plan.Document().ReferenceCategory && candidateHasSense(c, alternative.Permission.SenseID.String()) {
						isPredicateSlot = true
					}
				}
				for _, id := range slot.CandidateIDs {
					if c, found := candidates[id]; found && c.Category.IsPredicate() {
						vpIDs = appendUniqueString(vpIDs, id)
					}
				}
				if slot.Role == "" && nominalRelationMapping && len(slot.CandidateIDs) == 1 {
					if c, found := candidates[slot.CandidateIDs[0]]; found && c.Category == input.Plan.Document().ReferenceCategory && candidateHasSense(c, alternative.Permission.SenseID.String()) {
						vpIDs = appendUniqueString(vpIDs, slot.CandidateIDs[0])
					}
				}
				if !containsFrameRole && !isPredicateSlot {
					result.Complete = false
					result.Residuals = append(result.Residuals, DSEResidual{Code: "UNMAPPED_SOURCE_SLOT", Path: "forest.alternatives.slots", Message: "non-predicate source slot is not covered by this composition profile"})
				}
				if !containsFrameRole {
					continue
				}
				if _, exists := roleSlots[string(slot.Role)]; exists {
					return DSEResult{}, dseError(status.OutcomeInvalid, "DUPLICATE_SOURCE_ROLE_SLOT", fmt.Sprintf("forest.roots[%d].alternatives[%d].slots", rootIndex, altIndex), "semantic role appears more than once")
				}
				roleSlots[string(slot.Role)] = slot
			}
			if len(vpIDs) == 0 {
				if appendDSEIncomplete(&result, fmt.Sprintf("root:%s:%d", rootID, altIndex), DSESourceRecord{DerivationID: alternative.DerivationID, FamilyID: alternative.FamilyID.String(), VariantID: alternative.VariantID.String(), Span: root.Span}, "MISSING_PREDICATE_CANDIDATE", "parse alternative has no predicate phrase candidate", budget, work) {
					return result, nil
				}
				continue
			}
			for _, vpID := range vpIDs {
				vp := candidates[vpID]
				if !candidateHasSense(vp, alternative.Permission.SenseID.String()) {
					if appendDSEIncomplete(&result, fmt.Sprintf("%s:%s", alternative.DerivationID, vpID), dseRecord(vp, alternative), "PERMISSION_SENSE_NOT_IN_CANDIDATE", "exact parse permission sense is absent from predicate candidate readings", budget, work) {
						return result, nil
					}
					continue
				}
				mapping, hasMapping := sourcePredicateMapping(input.Plan.Document(), alternative.Permission.SenseID.String(), string(alternative.Permission.FrameID))
				if !hasMapping {
					if appendDSEIncomplete(&result, fmt.Sprintf("%s:%s", alternative.DerivationID, vpID), dseRecord(vp, alternative), "UNMAPPED_SOURCE_PREDICATE", "source plan has no exact sense/frame mapping", budget, work) {
						return result, nil
					}
					continue
				}
				if !containsConstructionID(mapping.AllowedFamilyIDs, alternative.FamilyID) || (len(mapping.AllowedVariantIDs) > 0 && !containsVariantID(mapping.AllowedVariantIDs, alternative.VariantID)) {
					if appendDSEIncomplete(&result, fmt.Sprintf("%s:%s", alternative.DerivationID, vpID), dseRecord(vp, alternative), "SOURCE_FAMILY_VARIANT_NOT_ALLOWED", "source plan does not authorize the exact parse family and variant", budget, work) {
						return result, nil
					}
					continue
				}
				frames := input.Package.Frames()
				var frame *valency.ValencyFrame
				for i := range frames {
					if frames[i].ID == alternative.Permission.FrameID && frames[i].SenseID == alternative.Permission.SenseID {
						f := frames[i]
						frame = &f
						break
					}
				}
				if frame == nil {
					if appendDSEIncomplete(&result, fmt.Sprintf("%s:%s", alternative.DerivationID, vpID), dseRecord(vp, alternative), "UNKNOWN_SOURCE_FRAME", "exact parse permission frame is not in selected package", budget, work) {
						return result, nil
					}
					continue
				}
				roleSelections := make([]dseRoleSelection, 0, len(frame.SemanticRoles))
				invalidSlots := false
				for _, role := range frame.SemanticRoles {
					slot, exists := roleSlots[string(role)]
					if !exists || (slot.WHGap == nil && len(slot.CandidateIDs) == 0) || (slot.WHGap != nil && (len(slot.CandidateIDs) != 0 || string(slot.WHGap.Role) != string(role) || slot.WHGap.ScopeNodeID == "" || !dseWHGapMatchesSource(slot.WHGap, input.Lattice))) {
						invalidSlots = true
						break
					}
					roleSelections = append(roleSelections, dseRoleSelection{role: string(role), slot: slot})
				}
				if invalidSlots {
					if appendDSEIncomplete(&result, fmt.Sprintf("%s:%s", alternative.DerivationID, vpID), dseRecord(vp, alternative), "MISSING_SOURCE_ROLE_CANDIDATE", "every required semantic frame role needs an explicit source NP candidate; WH gaps are outside this profile", budget, work) {
						return result, nil
					}
					continue
				}
				for _, supplied := range input.ReferenceContext.RoleReferences {
					knownRole := false
					for _, frameRole := range frame.SemanticRoles {
						if supplied.Role == string(frameRole) {
							knownRole = true
							break
						}
					}
					if !knownRole {
						result.Complete = false
						result.Residuals = append(result.Residuals, DSEResidual{Code: "UNUSED_ROLE_REFERENCE_MAPPING", Path: "reference_context.role_references", Message: "caller role mapping is not consumed by the exact selected frame"})
					}
				}
				for _, supplied := range input.ReferenceContext.PortReferences {
					if supplied.PortID != mapping.HeadPortID {
						result.Complete = false
						result.Residuals = append(result.Residuals, DSEResidual{Code: "UNUSED_PORT_REFERENCE_MAPPING", Path: "reference_context.port_references", Message: "only the exact source-plan Event head port is supported by this profile"})
					}
				}
				var expand func(int, []dseSelectedRole)
				var composeErr error
				expand = func(pos int, selectedRoles []dseSelectedRole) {
					if resourceStopped {
						return
					}
					if err := input.Context.Err(); err != nil {
						resourceStopped, resourceCode, resourcePath = true, "DSE_CANCELED", "context"
						return
					}
					if len(result.Candidates) >= budget.MaxAlternatives {
						resourceStopped, resourceCode, resourcePath = true, "DSE_ALTERNATIVE_LIMIT", "forest"
						return
					}
					if pos == len(roleSelections) {
						if !chargeDSEWork(&result, &work, budget.MaxWork) {
							resourceStopped, resourceCode, resourcePath = true, "DSE_WORK_LIMIT", "forest"
							return
						}
						candidate, complete, err := composeDSESelection(input, alternative, vp, mapping, frame, selectedRoles, referenceEvidence, attrs, mentionRefs)
						if err != nil {
							composeErr = err
							return
						}
						if !complete {
							result.Complete = false
						}
						candidateBytes, marshalErr := json.Marshal(candidate)
						if marshalErr != nil {
							composeErr = dseError(status.OutcomeError, "DSE_CANDIDATE_ENCODING", "candidate", "candidate could not be encoded within the validated profile")
							return
						}
						if len(candidateBytes) > budget.MaxOutputBytes-result.OutputBytes {
							result = dseResourcePrefix(result, "DSE_OUTPUT_LIMIT", candidate.ID, work)
							return
						}
						result.OutputBytes += len(candidateBytes)
						result.Candidates = append(result.Candidates, candidate)
						if input.Context.Err() != nil {
							resourceStopped, resourceCode, resourcePath = true, "DSE_CANCELED", "context"
						}
						return
					}
					slot := roleSelections[pos].slot
					if slot.WHGap != nil {
						selectedRoles = append(selectedRoles, dseSelectedRole{role: roleSelections[pos].role, gap: slot.WHGap})
						expand(pos+1, selectedRoles)
						selectedRoles = selectedRoles[:len(selectedRoles)-1]
						return
					}
					for _, candidateID := range slot.CandidateIDs {
						if !chargeDSEWork(&result, &work, budget.MaxWork) {
							resourceStopped, resourceCode, resourcePath = true, "DSE_WORK_LIMIT", "forest"
							return
						}
						c, exists := candidates[candidateID]
						if !exists {
							result.Complete = false
							result.Residuals = append(result.Residuals, DSEResidual{Code: "UNKNOWN_SLOT_CANDIDATE", CandidateID: candidateID, Message: "slot candidate ID is absent from actual phrase candidate inventory"})
							continue
						}
						if c.Category != construction.CategoryNP {
							result.Complete = false
							result.Residuals = append(result.Residuals, DSEResidual{Code: "UNSUPPORTED_ROLE_PHRASE_CATEGORY", CandidateID: c.ID, Message: "semantic frame roles require an explicitly parsed NP candidate in this profile"})
							continue
						}
						for _, sense := range c.SenseIDs {
							if !chargeDSEWork(&result, &work, budget.MaxWork) {
								resourceStopped, resourceCode, resourcePath = true, "DSE_WORK_LIMIT", "forest"
								return
							}
							selectedRoles = append(selectedRoles, dseSelectedRole{role: roleSelections[pos].role, candidate: c, senseID: sense.String()})
							expand(pos+1, selectedRoles)
							selectedRoles = selectedRoles[:len(selectedRoles)-1]
							if resourceStopped {
								return
							}
						}
					}
				}
				expand(0, nil)
				if composeErr != nil {
					return DSEResult{}, composeErr
				}
				if resourceStopped {
					return dseResourcePrefix(result, resourceCode, resourcePath, work), nil
				}
			}
		}
	}
	if input.Forest.Outcome != status.OutcomeOK && !dseForestIsIncompleteOnlyForWHGaps(input.Forest) {
		result.Complete = false
	}
	if !input.OperatorCoverageComplete {
		result.Complete = false
		result.Residuals = append(result.Residuals, DSEResidual{Code: "OPERATOR_COVERAGE_INCOMPLETE", Message: "source operator evidence is not declared complete"})
		for i := range result.Candidates {
			if result.Candidates[i].Status == status.OutcomeOK {
				result.Candidates[i].Status = status.OutcomeIncomplete
				result.Candidates[i].Residuals = append(result.Candidates[i].Residuals, DSEResidual{Code: "OPERATOR_COVERAGE_INCOMPLETE", Message: "source operator coverage is incomplete"})
			}
		}
	}
	if len(result.Residuals) > 0 {
		result.Complete = false
	}
	if len(result.Candidates) == 0 && result.Status == status.OutcomeOK {
		result.Status = status.OutcomeIncomplete
	}
	if !result.Complete && result.Status == status.OutcomeOK {
		result.Status = status.OutcomeIncomplete
	}
	return result, nil
}

func dseUnresolvedSlotsAreWHGaps(alternative parser.ParseAlternative) bool {
	if len(alternative.Unresolved) == 0 {
		return true
	}
	gapSlots := make(map[string]struct{}, len(alternative.Unresolved))
	for _, slot := range alternative.Slots {
		if slot.WHGap == nil || len(slot.CandidateIDs) != 0 || slot.WHGap.ScopeNodeID == "" || string(slot.WHGap.SlotID) != slot.SlotID {
			continue
		}
		gapSlots[slot.SlotID] = struct{}{}
	}
	if len(gapSlots) != len(alternative.Unresolved) {
		return false
	}
	for _, unresolved := range alternative.Unresolved {
		if _, ok := gapSlots[unresolved]; !ok {
			return false
		}
	}
	return true
}

func dseForestIsIncompleteOnlyForWHGaps(forest parser.Forest) bool {
	if forest.Outcome != status.OutcomeIncomplete || len(forest.Unmatched) != 0 || len(forest.Roots) == 0 {
		return false
	}
	byID := make(map[parser.NodeID]parser.PackedNode, len(forest.Nodes))
	for _, node := range forest.Nodes {
		byID[node.ID] = node
	}
	seenAlternative := false
	for _, rootID := range forest.Roots {
		root, ok := byID[rootID]
		if !ok || root.Category != "clause" || len(root.Alternatives) == 0 {
			return false
		}
		for _, alternative := range root.Alternatives {
			seenAlternative = true
			if len(alternative.Unresolved) == 0 || !dseUnresolvedSlotsAreWHGaps(alternative) {
				return false
			}
		}
	}
	return seenAlternative
}

// dseResolveExplicitWHGapOmissions removes only parser alternatives that are
// identical to a sibling apart from omission of an explicitly supplied WH
// gap. Optional argument slots otherwise admit both branches even when the
// same parse context supplies a scoped gap for that slot.
func dseResolveExplicitWHGapOmissions(forest parser.Forest) (parser.Forest, bool) {
	if forest.Outcome != status.OutcomeAmbiguous || len(forest.Unmatched) != 0 || len(forest.Roots) == 0 {
		return parser.Forest{}, false
	}
	nodesByID := make(map[parser.NodeID]parser.PackedNode, len(forest.Nodes))
	for _, node := range forest.Nodes {
		nodesByID[node.ID] = node
	}
	resolvedRoots := make(map[parser.NodeID][]parser.ParseAlternative, len(forest.Roots))
	for _, rootID := range forest.Roots {
		root, ok := nodesByID[rootID]
		if !ok || root.Category != "clause" || len(root.Alternatives) < 2 {
			return parser.Forest{}, false
		}
		gapIndex := -1
		for i, alternative := range root.Alternatives {
			count := 0
			for _, slot := range alternative.Slots {
				if slot.WHGap != nil {
					if slot.WHGap.ScopeNodeID == "" || slot.WHGap.SlotID != slot.SlotID || string(slot.WHGap.Role) != string(slot.Role) || len(slot.CandidateIDs) != 0 {
						return parser.Forest{}, false
					}
					count++
				}
			}
			if count == 1 {
				if gapIndex >= 0 {
					return parser.Forest{}, false
				}
				gapIndex = i
			} else if count != 0 {
				return parser.Forest{}, false
			}
		}
		if gapIndex < 0 {
			return parser.Forest{}, false
		}
		gapAlternative := root.Alternatives[gapIndex]
		for i, alternative := range root.Alternatives {
			if i != gapIndex && !dseSameParseExceptWHGap(gapAlternative, alternative) {
				return parser.Forest{}, false
			}
		}
		resolvedRoots[rootID] = []parser.ParseAlternative{gapAlternative}
	}
	filtered := forest
	filtered.Nodes = append([]parser.PackedNode(nil), forest.Nodes...)
	for i := range filtered.Nodes {
		if alternatives, ok := resolvedRoots[filtered.Nodes[i].ID]; ok {
			filtered.Nodes[i].Alternatives = alternatives
		}
	}
	filtered.Outcome = status.OutcomeOK
	return filtered, true
}

func dseSameParseExceptWHGap(withGap, withoutGap parser.ParseAlternative) bool {
	if withGap.PlanID != withoutGap.PlanID || withGap.FamilyID != withoutGap.FamilyID || withGap.VariantID != withoutGap.VariantID || withGap.Permission != withoutGap.Permission || len(withGap.Slots) != len(withoutGap.Slots) || len(withGap.Unresolved) != len(withoutGap.Unresolved) {
		return false
	}
	for i := range withGap.Unresolved {
		if withGap.Unresolved[i] != withoutGap.Unresolved[i] {
			return false
		}
	}
	gapCount := 0
	for i := range withGap.Slots {
		a, b := withGap.Slots[i], withoutGap.Slots[i]
		if a.SlotID != b.SlotID || a.Role != b.Role || len(a.CandidateIDs) != len(b.CandidateIDs) || len(a.NodeIDs) != len(b.NodeIDs) || b.WHGap != nil {
			return false
		}
		for j := range a.CandidateIDs {
			if a.CandidateIDs[j] != b.CandidateIDs[j] {
				return false
			}
		}
		for j := range a.NodeIDs {
			if a.NodeIDs[j] != b.NodeIDs[j] {
				return false
			}
		}
		if a.WHGap != nil {
			gapCount++
		}
	}
	return gapCount == 1
}

type dseRoleSelection struct {
	role string
	slot parser.SlotBinding
}
type dseSelectedRole struct {
	role      string
	candidate parser.PhraseCandidate
	senseID   string
	gap       *parser.WHGapEvidence
}

func dseWHGapMatchesSource(gap *parser.WHGapEvidence, lattice analyzer.LexicalLattice) bool {
	if gap == nil || gap.Span.Start < 0 || gap.Span.End <= gap.Span.Start {
		return false
	}
	for _, token := range lattice.Tokens {
		if token.Token.Kind == text.TokenWord && token.Token.Span == gap.Span {
			return true
		}
	}
	return false
}

func composeDSESelection(input DSEInput, alternative parser.ParseAlternative, vp parser.PhraseCandidate, mapping language.SourcePredicateMapping, frame *valency.ValencyFrame, roles []dseSelectedRole, refsByCandidate map[string]ReferenceEvidence, attrs map[string]NodeAttributesEvidence, mentionRefs map[string]dseMentionReference) (DSECandidate, bool, error) {
	pack := input.Package
	bindings := pack.LexiconIndex().TargetBindingsForSense(alternative.Permission.SenseID)
	selectedIndex := -1
	var target *lexicon.SemanticTarget
	for i := range bindings {
		b := bindings[i]
		if !b.TargetExecutionAuthorized() || b.Target == nil || b.Target.Kind == lexicon.SemanticTargetReferencePredicate {
			continue
		}
		matched := false
		for _, rp := range b.Target.RolePorts {
			if rp.FrameID == string(alternative.Permission.FrameID) {
				matched = true
				break
			}
		}
		if matched {
			if selectedIndex >= 0 {
				return incompleteDSE(vp, alternative, "AMBIGUOUS_FRAME_TARGET", "multiple approved targets cover the exact frame")
			}
			selectedIndex = i
			target = b.Target
		}
	}
	if selectedIndex < 0 || target == nil {
		return incompleteDSE(vp, alternative, "MISSING_APPROVED_FRAME_TARGET", "no approved target covers the exact frame")
	}
	roleRefs := make([]language.SemanticRoleReference, 0, len(roles))
	roleReferenceByRole := make(map[string]string, len(roles))
	roleBasisByRole := make(map[string]string, len(roles))
	referenceByCandidate := make(map[string]string, len(roles))
	var answer *algebra.Declaration
	var answerRole string
	var err error
	for _, selected := range roles {
		if selected.gap != nil {
			if _, hasRoleReference := sourceRoleReference(input.ReferenceContext, selected.role); hasRoleReference {
				return incompleteDSE(vp, alternative, "WH_ROLE_REFERENCE_CONFLICT", "a WH answer role cannot also be bound to a caller-supplied referent")
			}
			if answer != nil || selected.gap.ScopeNodeID == "" || string(selected.gap.Role) != selected.role {
				return incompleteDSE(vp, alternative, "INVALID_WH_GAP", "the selected question variant must provide one exact frame-role gap")
			}
			port, ok := dseTargetRolePort(target, string(alternative.Permission.FrameID), selected.role)
			if !ok {
				return incompleteDSE(vp, alternative, "UNMAPPED_WH_ROLE", "the exact selected frame role has no authorized target port")
			}
			answerID := "answer_" + strings.TrimPrefix(string(selected.gap.ScopeNodeID), "wh_")
			if len(answerID) > 256 {
				return incompleteDSE(vp, alternative, "WH_ANSWER_ID_LIMIT", "question answer binder exceeds the bounded identifier profile")
			}
			answer = &algebra.Declaration{ID: answerID, Type: cloneDSEType(port.Type)}
			answerRole = selected.role
			continue
		}
		fallback, hasFallback := sourceRoleReference(input.ReferenceContext, selected.role)
		refID, resolved, anchored, basis, err := resolveDSEReference(selected.candidate.ID, fallback, hasFallback, refsByCandidate, mentionRefs)
		if err != nil {
			return DSECandidate{}, false, err
		}
		if anchored || !resolved {
			return incompleteDSE(vp, alternative, "MISSING_ROLE_REFERENCE_MAPPING", "caller did not explicitly bind a source semantic role")
		}
		roleRefs = append(roleRefs, language.SemanticRoleReference{Role: selected.role, ReferenceID: refID})
		roleReferenceByRole[selected.role] = refID
		roleBasisByRole[selected.role] = basis
		if previous, exists := referenceByCandidate[selected.candidate.ID]; exists && previous != refID {
			return DSECandidate{}, false, dseError(status.OutcomeInvalid, "CANDIDATE_REFERENCE_CONFLICT", "reference_context.role_references", "one source candidate cannot resolve to different caller references")
		}
		referenceByCandidate[selected.candidate.ID] = refID
	}
	relationPredicate := mapping.PredicateKind == "relation"
	headRef, ok := "", false
	if !relationPredicate {
		headRef, ok = sourcePortReference(input.ReferenceContext, mapping.HeadPortID)
	}
	if input.RequireCandidateHeadReference && !relationPredicate {
		resolvedHead, resolved, anchoredHead, _, resolveErr := resolveDSEReference(vp.ID, "", false, refsByCandidate, mentionRefs)
		if resolveErr != nil {
			return DSECandidate{}, false, resolveErr
		}
		if anchoredHead || !resolved {
			return incompleteDSE(vp, alternative, "MISSING_CANDIDATE_EVENT_REFERENCE", "the predicate candidate has no explicit, resolved Event identity in this request")
		}
		if ok && headRef != resolvedHead {
			return DSECandidate{}, false, dseError(status.OutcomeInvalid, "CANDIDATE_EVENT_REFERENCE_CONFLICT", "reference_context.port_references", "the candidate Event identity conflicts with the supplied Event port mapping")
		}
		headRef, ok = resolvedHead, true
	}
	if !relationPredicate && !ok {
		return incompleteDSE(vp, alternative, "MISSING_HEAD_EVENT_REFERENCE", "caller must explicitly bind the mapped Event head port")
	}
	if !relationPredicate {
		if sourceEvidence, exists := refsByCandidate[vp.ID]; exists {
			if sourceEvidence.ReferentID != "" && string(sourceEvidence.ReferentID) != headRef {
				return DSECandidate{}, false, dseError(status.OutcomeInvalid, "HEAD_REFERENCE_EVIDENCE_CONFLICT", "reference_context.port_references", "caller Event binding conflicts with resolved source evidence")
			}
		}
	}
	request := language.SemanticBindingRequest{SenseID: alternative.Permission.SenseID, FrameID: alternative.Permission.FrameID, BindingIndex: &selectedIndex, ContextID: input.ReferenceContext.ContextID, References: cloneAlgebraDeclarations(input.ReferenceContext.References), RequestContext: cloneDSEContext(input.ReferenceContext.RequestContext), RoleReferences: roleRefs}
	if !relationPredicate {
		request.PortReferences = []language.SemanticPortReference{{PortID: mapping.HeadPortID, ReferenceID: headRef}}
	}
	var instantiated language.SemanticInstantiationResult
	var predicate *algebra.Term
	var candidate DSECandidate
	candidate = dseCandidateStart(vp, alternative)
	if relationPredicate {
		candidate.PredicateKind = "relation"
	} else {
		candidate.HeadEventReferenceID = headRef
	}
	for _, selected := range roles {
		if selected.gap != nil {
			gap := DSEWHGapEvidence{SlotID: selected.gap.SlotID, Role: string(selected.gap.Role), ScopeNodeID: string(selected.gap.ScopeNodeID), Span: selected.gap.Span}
			source := DSESourceRecord{ReadingIDs: []string{}, Span: selected.gap.Span}
			candidate.RoleSources = append(candidate.RoleSources, DSERoleSource{Role: selected.role, Source: source, WHGap: &gap, ResolutionBasis: "parser_wh_gap"})
			continue
		}
		candidate.RoleSources = append(candidate.RoleSources, DSERoleSource{Role: selected.role, Source: dseRoleRecord(selected.candidate, selected.senseID), ReferenceID: roleReferenceByRole[selected.role], ResolutionBasis: roleBasisByRole[selected.role]})
	}
	instantiatedStatus := status.OutcomeOK
	if answer != nil {
		if relationPredicate {
			return incompleteDSE(vp, alternative, "RELATION_QUESTION_UNSUPPORTED", "relation predicates in this selected source profile do not license WH questions")
		}
		predicate, err = buildDSEQuestionAtom(input.Selected, target, string(alternative.Permission.FrameID), mapping.HeadPortID, roleReferenceByRole, headRef, answerRole, answer.ID)
		if err != nil {
			return incompleteDSE(vp, alternative, "QUESTION_TARGET_REFUSED", boundedDSEMessage(err.Error()))
		}
		typ := algebra.Type{Kind: algebra.TypeProp}
		candidate.FragmentChecks = append(candidate.FragmentChecks, DSEFragmentCheck{Kind: "predicate", Status: status.OutcomeOK, CheckedType: &typ, Interface: []algebra.PortBinding{}})
	} else {
		instantiated, err = language.InstantiateSemanticBinding(pack, input.Selected, input.Rewrite, request)
		if err != nil {
			return incompleteDSE(vp, alternative, "PREDICATE_INSTANTIATION_REFUSED", err.Error())
		}
		instantiatedStatus = instantiated.Status
	}
	if answer == nil && instantiated.Symbol != nil {
		predicate = instantiated.Symbol.Document.Expression
		checked := instantiated.Symbol.Checked
		candidate.Checked = &checked
		typ := checked.Type
		candidate.FragmentChecks = append(candidate.FragmentChecks, DSEFragmentCheck{Kind: "predicate", BindingKey: instantiated.BindingKey, Status: instantiated.Status, CheckedType: &typ, Interface: append([]algebra.PortBinding{}, instantiated.Symbol.Interface...)})
	} else if answer == nil && instantiated.Goal != nil {
		goalWire, _ := json.Marshal(instantiated.Goal)
		candidate.GoalSpecs = append(candidate.GoalSpecs, goalWire)
		if instantiated.Construction != nil && instantiated.Construction.CanonicalDocument != nil && instantiated.Construction.Checked != nil {
			predicate = instantiated.Construction.CanonicalDocument.Expression
			checked := *instantiated.Construction.Checked
			candidate.Checked = &checked
			typ := checked.Type
			documentWire, encodeErr := algebra.Encode(*instantiated.Construction.CanonicalDocument)
			if encodeErr != nil {
				return DSECandidate{}, false, encodeErr
			}
			goalEvidence, evidenceErr := json.Marshal(struct {
				Goal              json.RawMessage                   `json:"goal"`
				CanonicalDocument json.RawMessage                   `json:"canonical_document"`
				Certificate       *synthesis.SynthesisCertificate   `json:"certificate"`
				Verification      *synthesis.GoalVerificationResult `json:"verification"`
				Scope             string                            `json:"scope"`
			}{Goal: goalWire, CanonicalDocument: documentWire, Certificate: instantiated.Certificate, Verification: instantiated.Verification, Scope: "verified instantiated goal fragment only; not the enclosing source composition"})
			if evidenceErr != nil {
				return DSECandidate{}, false, evidenceErr
			}
			candidate.FragmentChecks = append(candidate.FragmentChecks, DSEFragmentCheck{Kind: "goal_fragment", BindingKey: instantiated.BindingKey, Status: instantiated.Status, CheckedType: &typ, Interface: append([]algebra.PortBinding{}, instantiated.Construction.Interface...), GoalEvidence: goalEvidence})
		}
	} else if answer == nil {
		return incompleteDSE(vp, alternative, "UNSUPPORTED_TARGET_RESULT", "selected semantic target did not produce a typed symbol or goal expression")
	}
	if predicate == nil {
		return incompleteDSE(vp, alternative, "INCOMPLETE_TARGET_EXPRESSION", "selected target has no fully constructed proposition")
	}
	if instantiatedStatus != status.OutcomeOK {
		candidate.Status = status.OutcomeIncomplete
		candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "OPEN_TARGET_INTERFACE", CandidateID: vp.ID, Message: "semantic target retains an unbound interface port"})
	}
	roleCoverage := len(roleReferenceByRole)
	if answer != nil {
		roleCoverage++
	}
	if roleCoverage != len(frame.SemanticRoles) {
		return incompleteDSE(vp, alternative, "SOURCE_ROLE_COVERAGE_MISMATCH", "explicit role bindings do not cover the selected frame")
	}
	// Rebuild a standalone context after adding explicit source attributes and
	// non-asserted conditions. Conditions are not appended to the proposition.
	requestContext := cloneDSEContext(input.ReferenceContext.RequestContext)
	if requestContext == nil {
		requestContext = &algebra.RequestContext{SchemaVersion: algebra.RequestContextSchemaVersion, ID: input.ReferenceContext.ContextID, Referents: []algebra.ContextReferent{}, Mentions: []algebra.ContextMention{}, ResolutionEvidence: []algebra.ResolutionEvidence{}, Conditions: []algebra.ContextCondition{}}
	}
	if requestContext.Conditions == nil {
		requestContext.Conditions = []algebra.ContextCondition{}
	}
	conditionCoverage := []DSEConditionCoverage{}
	for _, selected := range roles {
		if selected.gap != nil {
			continue
		}
		evidence, exists := refsByCandidate[selected.candidate.ID]
		if !exists || evidence.Anchor == nil {
			continue
		}
		condition := algebra.ContextCondition{Kind: algebra.ContextPresupposition, Code: "deictic_anchor", TargetReferentIDs: []string{roleReferenceByRole[selected.role]}, Detail: evidence.Anchor.Value}
		condition.ID = dseConditionID(condition)
		conditionID := addDSECondition(requestContext, condition)
		conditionCoverage = append(conditionCoverage, DSEConditionCoverage{Kind: language.SourceConditionKind(condition.Kind), Code: condition.Code, ConditionID: conditionID, TargetReferenceIDs: append([]string{}, condition.TargetReferentIDs...)})
	}
	roleNPConditions, roleNPChecks, roleNPResiduals, err := instantiateRolePredicates(input, alternative, roles, roleReferenceByRole, request, requestContext)
	if err != nil {
		return DSECandidate{}, false, err
	}
	candidate.FragmentChecks = append(candidate.FragmentChecks, roleNPChecks...)
	candidate.Residuals = append(candidate.Residuals, roleNPResiduals...)
	for _, cond := range roleNPConditions {
		conditionID := addDSECondition(requestContext, cond)
		conditionCoverage = append(conditionCoverage, DSEConditionCoverage{Kind: language.SourceConditionKind(cond.Kind), Code: cond.Code, ConditionID: conditionID, TargetReferenceIDs: append([]string{}, cond.TargetReferentIDs...)})
	}
	featureHeadRef := headRef
	if relationPredicate {
		featureHeadRef = ""
	}
	featureConditions, coverage, featureResiduals, err := instantiateFeatureConditions(input, alternative, vp, roles, target, mapping, roleReferenceByRole, featureHeadRef, request, requestContext)
	if err != nil {
		return DSECandidate{}, false, err
	}
	for i, cond := range featureConditions {
		conditionID := addDSECondition(requestContext, cond)
		if i < len(coverage) {
			coverage[i].ConditionID = conditionID
		}
	}
	conditionCoverage = append(conditionCoverage, coverage...)
	candidate.Residuals = append(candidate.Residuals, featureResiduals...)
	if instantiatedStatus == status.OutcomeOK {
		if err := applyOperatorMappings(input, vp, &predicate, &candidate); err != nil {
			return DSECandidate{}, false, err
		}
	} else if len(input.Operators) > 0 {
		candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "NEGATION_REQUIRES_CLOSED_PREDICATE", CandidateID: vp.ID, Message: "negation cannot be applied to a target with an open interface"})
	}
	final := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: cloneAlgebraSorts(input.Selected.Sorts), Signatures: cloneAlgebraSignatures(input.Selected.Signatures), References: cloneAlgebraDeclarations(input.ReferenceContext.References), Closed: true, Expression: predicate, Limits: input.Selected.Limits, RequestContext: requestContext}
	if answer != nil {
		asked, askErr := algebra.NewAsk(final, []algebra.Declaration{*answer}, nil, predicate)
		if askErr != nil {
			return incompleteDSE(vp, alternative, "QUESTION_SCOPE_REFUSED", boundedDSEMessage(askErr.Error()))
		}
		final.Expression = asked
	}
	for _, selected := range roles {
		if err := mergeDSECandidateAttributes(&final, selected.candidate, roleReferenceByRole[selected.role], attrs); err != nil {
			return DSECandidate{}, false, err
		}
	}
	if !relationPredicate {
		if err := mergeDSECandidateAttributes(&final, vp, headRef, attrs); err != nil {
			return DSECandidate{}, false, err
		}
	}
	if err := algebra.CheckDeclarations(algebra.Document{SchemaVersion: final.SchemaVersion, Sorts: final.Sorts, Signatures: final.Signatures, References: final.References, Limits: final.Limits, RequestContext: final.RequestContext}); err != nil {
		return DSECandidate{}, false, err
	}
	checked, err := algebra.Check(final)
	if err != nil {
		return DSECandidate{}, false, err
	}
	encoded, err := algebra.Encode(final)
	if err != nil {
		return DSECandidate{}, false, err
	}
	candidate.CanonicalDocument = append(json.RawMessage{}, encoded...)
	candidate.Checked = &checked
	candidate.ConditionCoverage = conditionCoverage
	if len(candidate.Residuals) > 0 || candidate.Status == status.OutcomeIncomplete {
		candidate.Status = status.OutcomeIncomplete
		candidate.Scope = "typed representation; unresolved/residual source evidence; no world-fact claim"
		return candidate, false, nil
	}
	candidate.Status = status.OutcomeOK
	candidate.Scope = "typed algebra representation and explicit request context; no world-fact or language-quality claim"
	return candidate, true, nil
}

func dseTargetRolePort(target *lexicon.SemanticTarget, frameID, role string) (lexicon.SemanticRolePort, bool) {
	if target == nil {
		return lexicon.SemanticRolePort{}, false
	}
	var found lexicon.SemanticRolePort
	for _, port := range target.RolePorts {
		if port.FrameID != frameID || port.Role != role {
			continue
		}
		if found.PortID != "" {
			return lexicon.SemanticRolePort{}, false
		}
		found = port
	}
	return found, found.PortID != ""
}

func buildDSEQuestionAtom(selected algebra.Document, target *lexicon.SemanticTarget, frameID, headPort string, roleRefs map[string]string, headRef, answerRole, answerID string) (*algebra.Term, error) {
	if target == nil || target.Kind != lexicon.SemanticTargetSymbol || target.SymbolRef == "" || headPort == "" || headRef == "" || answerRole == "" || answerID == "" {
		return nil, fmt.Errorf("WH questions require one approved predicate symbol and explicit event/answer bindings")
	}
	var signature *algebra.Signature
	for i := range selected.Signatures {
		if selected.Signatures[i].ID == string(target.SymbolRef) {
			if signature != nil {
				return nil, fmt.Errorf("predicate symbol has duplicate selected signatures")
			}
			signature = &selected.Signatures[i]
		}
	}
	if signature == nil || len(signature.Ports) == 0 || len(signature.Ports) > algebra.DefaultMaxPorts {
		return nil, fmt.Errorf("approved predicate signature is unavailable or outside the finite profile")
	}
	args := make([]algebra.Argument, 0, len(signature.Ports))
	usedRoles := make(map[string]bool, len(roleRefs)+1)
	usedHead := false
	for _, port := range signature.Ports {
		var value *algebra.Term
		if port.Role == headPort {
			if usedHead {
				return nil, fmt.Errorf("event port is repeated in the predicate signature")
			}
			usedHead = true
			typ := cloneDSEType(port.Type)
			value = &algebra.Term{Kind: algebra.TermRef, ID: headRef, Type: &typ}
		} else {
			var rolePort *lexicon.SemanticRolePort
			for i := range target.RolePorts {
				candidate := &target.RolePorts[i]
				if candidate.FrameID == frameID && candidate.PortID == port.Role {
					if rolePort != nil {
						return nil, fmt.Errorf("predicate port has duplicate exact frame-role mappings")
					}
					rolePort = candidate
				}
			}
			if rolePort == nil || !algebra.TypeEqual(rolePort.Type, port.Type) || usedRoles[rolePort.Role] {
				return nil, fmt.Errorf("predicate port lacks one exact typed frame-role mapping")
			}
			usedRoles[rolePort.Role] = true
			if rolePort.Role == answerRole {
				value = &algebra.Term{Kind: algebra.TermVar, ID: answerID}
			} else {
				refID := roleRefs[rolePort.Role]
				if refID == "" {
					return nil, fmt.Errorf("non-question frame role lacks an explicit reference")
				}
				typ := cloneDSEType(port.Type)
				value = &algebra.Term{Kind: algebra.TermRef, ID: refID, Type: &typ}
			}
		}
		args = append(args, algebra.Argument{Role: port.Role, Value: value})
	}
	if !usedHead || len(usedRoles) != len(roleRefs)+1 || !usedRoles[answerRole] {
		return nil, fmt.Errorf("predicate signature does not cover the selected frame roles and WH answer exactly")
	}
	return &algebra.Term{Kind: algebra.TermAtom, Symbol: string(target.SymbolRef), Args: args}, nil
}

func instantiateRolePredicates(input DSEInput, alternative parser.ParseAlternative, roles []dseSelectedRole, roleRefs map[string]string, request language.SemanticBindingRequest, rc *algebra.RequestContext) ([]algebra.ContextCondition, []DSEFragmentCheck, []DSEResidual, error) {
	var conditions []algebra.ContextCondition
	var checks []DSEFragmentCheck
	var residuals []DSEResidual
	planRow := language.SourcePredicateMapping{}
	for _, r := range input.Plan.Document().Predicates {
		if r.SenseID == alternative.Permission.SenseID && string(r.FrameID) == string(alternative.Permission.FrameID) {
			planRow = r
			break
		}
	}
	if planRow.ReferenceCondition == nil {
		for _, r := range roles {
			residuals = append(residuals, DSEResidual{Code: "MISSING_REFERENCE_CONDITION", CandidateID: r.candidate.ID, Message: "source plan does not declare a non-asserted placement for nominal predicate conditions"})
		}
		return conditions, checks, residuals, nil
	}
	for _, selected := range roles {
		if selected.gap != nil {
			continue
		}
		_, ok := input.Package.LexiconIndex().ByLexemeID(selected.candidate.LexemeID)
		if !ok {
			residuals = append(residuals, DSEResidual{Code: "UNKNOWN_ROLE_LEXEME", CandidateID: selected.candidate.ID, Message: "selected NP lexeme is absent from package"})
			continue
		}
		if residual := dseReferenceNonHeadResidual(input.Lattice, selected.candidate, input.Plan.Document().ReferenceNonHeadPOSAllowlist); residual != nil {
			residuals = append(residuals, *residual)
		}
		senseID, parseErr := ids.ParseSenseID(selected.senseID)
		if parseErr != nil {
			residuals = append(residuals, DSEResidual{Code: "INVALID_NP_SENSE_ID", CandidateID: selected.candidate.ID, Message: "NP candidate sense ID is malformed"})
			continue
		}
		bindings := input.Package.LexiconIndex().TargetBindingsForSense(senseID)
		found := false
		var selectedCondition *algebra.ContextCondition
		var selectedCheck *DSEFragmentCheck
		for index, binding := range bindings {
			if !binding.TargetExecutionAuthorized() || binding.Target == nil || binding.Target.Kind != lexicon.SemanticTargetReferencePredicate || binding.Target.ReferencePort == nil {
				continue
			}
			found = true
			ctx := language.SemanticBindingRequest{SenseID: senseID, BindingIndex: &index, ContextID: request.ContextID, References: cloneAlgebraDeclarations(request.References), RequestContext: cloneDSEContext(request.RequestContext), RoleReferences: []language.SemanticRoleReference{}, PortReferences: []language.SemanticPortReference{{PortID: binding.Target.ReferencePort.PortID, ReferenceID: roleRefs[selected.role]}}}
			inst, err := language.InstantiateSemanticBinding(input.Package, input.Selected, input.Rewrite, ctx)
			if err != nil {
				residuals = append(residuals, DSEResidual{Code: "REFERENCE_PREDICATE_REFUSED", CandidateID: selected.candidate.ID, Message: boundedDSEMessage(err.Error())})
				continue
			}
			if inst.Symbol == nil || inst.Symbol.Document.Expression == nil || inst.Status != status.OutcomeOK {
				residuals = append(residuals, DSEResidual{Code: "REFERENCE_PREDICATE_INCOMPLETE", CandidateID: selected.candidate.ID, Message: "reference predicate did not produce a closed typed proposition"})
				continue
			}
			expr := inst.Symbol.Document.Expression
			cond := algebra.ContextCondition{Kind: algebra.ContextConditionKind(planRow.ReferenceCondition.Kind), Code: planRow.ReferenceCondition.Code, TargetReferentIDs: []string{roleRefs[selected.role]}, Expression: expr, Detail: ""}
			cond.ID = dseConditionID(cond)
			if selectedCondition != nil {
				residuals = append(residuals, DSEResidual{Code: "AMBIGUOUS_REFERENCE_PREDICATE", CandidateID: selected.candidate.ID, Message: "NP sense has multiple approved reference-predicate targets; no target is selected implicitly"})
				selectedCondition = nil
				selectedCheck = nil
				break
			}
			selectedCondition = &cond
			typ := inst.Symbol.Checked.Type
			check := DSEFragmentCheck{Kind: "reference_condition", BindingKey: inst.BindingKey, Status: inst.Status, CheckedType: &typ, Interface: append([]algebra.PortBinding{}, inst.Symbol.Interface...)}
			selectedCheck = &check
		}
		if !found {
			residuals = append(residuals, DSEResidual{Code: "MISSING_REFERENCE_PREDICATE", CandidateID: selected.candidate.ID, Message: "NP sense has no approved reference_predicate target"})
		}
		if selectedCondition != nil && selectedCheck != nil {
			conditions = append(conditions, *selectedCondition)
			checks = append(checks, *selectedCheck)
		}
	}
	return conditions, checks, residuals, nil
}

func dseReferenceNonHeadResidual(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate, allowedPOS []string) *DSEResidual {
	if candidate.TokenStart < 0 || candidate.TokenEnd <= candidate.TokenStart || candidate.TokenEnd > len(lattice.Tokens) {
		return &DSEResidual{Code: "INVALID_REFERENCE_CANDIDATE_SPAN", CandidateID: candidate.ID, Message: "NP candidate token span is outside the analyzed lexical lattice"}
	}
	headReadings := make(map[string]bool, len(candidate.ReadingIDs))
	for _, id := range candidate.ReadingIDs {
		headReadings[id] = true
	}
	headFound := false
	for index := candidate.TokenStart; index < candidate.TokenEnd; index++ {
		token := lattice.Tokens[index]
		isHead := false
		for _, reading := range token.Readings {
			if headReadings[reading.Reading.ID] && reading.Lexeme.ID == candidate.LexemeID {
				isHead = true
				headFound = true
			}
		}
		if isHead {
			continue
		}
		if len(token.Readings) == 0 {
			return &DSEResidual{Code: "UNMAPPED_REFERENCE_NONHEAD", CandidateID: candidate.ID, Path: fmt.Sprintf("tokens[%d]", index), Message: "non-head token inside an NP has no lexical POS allowed by the source plan"}
		}
		for _, reading := range token.Readings {
			if !containsString(allowedPOS, string(reading.Lexeme.POS)) {
				return &DSEResidual{Code: "UNMAPPED_REFERENCE_NONHEAD_POS", CandidateID: candidate.ID, Path: fmt.Sprintf("tokens[%d]", index), Message: "non-head NP POS is not covered by the source-plan allowlist"}
			}
		}
	}
	if !headFound {
		return &DSEResidual{Code: "REFERENCE_HEAD_READING_MISSING", CandidateID: candidate.ID, Message: "NP phrase span does not contain its selected lexical head reading"}
	}
	return nil
}

func instantiateFeatureConditions(input DSEInput, alternative parser.ParseAlternative, vp parser.PhraseCandidate, roles []dseSelectedRole, target *lexicon.SemanticTarget, mapping language.SourcePredicateMapping, roleRefs map[string]string, headRef string, request language.SemanticBindingRequest, rc *algebra.RequestContext) ([]algebra.ContextCondition, []DSEConditionCoverage, []DSEResidual, error) {
	allCandidates := []parser.PhraseCandidate{vp}
	for _, role := range roles {
		if role.gap == nil {
			allCandidates = append(allCandidates, role.candidate)
		}
	}
	doc := input.Plan.Document()
	scopedFeatures := make(map[string]bool, len(doc.FeatureConditions)+len(doc.RequiredFeatureIDs))
	for _, condition := range doc.FeatureConditions {
		scopedFeatures[condition.FeatureID] = true
	}
	for _, feature := range doc.RequiredFeatureIDs {
		scopedFeatures[feature] = true
	}
	conditions := []algebra.ContextCondition{}
	coverage := []DSEConditionCoverage{}
	residuals := []DSEResidual{}
	for _, candidate := range allCandidates {
		known := map[string]string{}
		featureIDs := make([]string, 0, len(candidate.Features))
		for feature := range candidate.Features {
			featureIDs = append(featureIDs, string(feature))
		}
		sort.Strings(featureIDs)
		for _, feature := range featureIDs {
			var value string
			for id, v := range candidate.Features {
				if string(id) == feature {
					value = string(v)
					break
				}
			}
			known[feature] = value
		}
		knownIDs := make([]string, 0, len(known))
		for feature := range known {
			knownIDs = append(knownIDs, feature)
		}
		sort.Strings(knownIDs)
		for _, feature := range knownIDs {
			if !scopedFeatures[feature] {
				continue
			}
			value := known[feature]
			var row *language.SourceFeatureCondition
			for i := range doc.FeatureConditions {
				if doc.FeatureConditions[i].FeatureID == feature && doc.FeatureConditions[i].ValueID == value {
					row = &doc.FeatureConditions[i]
					break
				}
			}
			if row == nil {
				residuals = append(residuals, DSEResidual{Code: "UNMAPPED_SOURCE_FEATURE", CandidateID: candidate.ID, Message: "known source feature has no explicit feature-condition mapping"})
				continue
			}
			refID := ""
			if candidate.ID == vp.ID {
				refID = headRef
			}
			for _, role := range roles {
				if role.candidate.ID != candidate.ID {
					continue
				}
				roleRef := roleRefs[role.role]
				if refID != "" && refID != roleRef {
					return nil, nil, nil, dseError(status.OutcomeInvalid, "AMBIGUOUS_FEATURE_REFERENCE", "feature_conditions.port_id", "feature condition port maps to competing role references")
				}
				refID = roleRef
			}
			if refID == "" {
				residuals = append(residuals, DSEResidual{Code: "FEATURE_REFERENCE_UNBOUND", CandidateID: candidate.ID, Message: "feature condition port does not identify an explicitly bound source reference"})
				continue
			}
			conditionContext := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: cloneAlgebraSorts(input.Selected.Sorts), Signatures: cloneAlgebraSignatures(input.Selected.Signatures), References: cloneAlgebraDeclarations(request.References), Limits: input.Selected.Limits, RequestContext: cloneDSEContext(request.RequestContext)}
			function, err := algebra.Relation(conditionContext, row.SymbolRef)
			if err != nil {
				return nil, nil, nil, err
			}
			boundType := cloneDSEType(row.Type)
			function, err = algebra.Bind(conditionContext, function, row.PortID, &algebra.Term{Kind: algebra.TermRef, ID: refID, Type: &boundType})
			if err != nil {
				return nil, nil, nil, err
			}
			cond := algebra.ContextCondition{Kind: algebra.ContextConditionKind(row.Condition), Code: row.Code, TargetReferentIDs: []string{refID}, Expression: function.Expression}
			cond.ID = dseConditionID(cond)
			conditions = append(conditions, cond)
			coverage = append(coverage, DSEConditionCoverage{Kind: row.Condition, Code: row.Code, ConditionID: cond.ID, TargetReferenceIDs: []string{refID}})
		}
	}
	allKnown := map[string]bool{}
	for _, candidate := range allCandidates {
		for feature := range candidate.Features {
			allKnown[string(feature)] = true
		}
	}
	for _, required := range doc.RequiredFeatureIDs {
		if !allKnown[required] {
			residuals = append(residuals, DSEResidual{Code: "REQUIRED_FEATURE_UNKNOWN", CandidateID: vp.ID, Message: "required source feature is unknown across this selected interpretation"})
		}
	}
	return conditions, coverage, residuals, nil
}

func applyOperatorMappings(input DSEInput, vp parser.PhraseCandidate, predicate **algebra.Term, candidate *DSECandidate) error {
	plan := input.Plan.Document()
	type operatorKey struct {
		target string
		kind   construction.OperatorKind
		value  construction.OperatorValue
		start  int
		end    int
	}
	seen := make(map[operatorKey]struct{}, len(input.Operators))
	for _, operator := range input.Operators {
		key := operatorKey{target: operator.TargetCandidateID, kind: operator.Operator.Kind, value: operator.Operator.Value, start: operator.Span.Start, end: operator.Span.End}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		if operator.TargetCandidateID != vp.ID {
			candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "UNSUPPORTED_OPERATOR_TARGET", CandidateID: operator.TargetCandidateID, Message: "operator targets a source candidate outside the selected predicate"})
			continue
		}
		if operator.Span.Start < 0 || operator.Span.End <= operator.Span.Start || operator.Span.End > input.SourceBytes {
			candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "INVALID_OPERATOR_SPAN", CandidateID: vp.ID, Message: "operator span is not a valid source range"})
			continue
		}
		if operator.Operator.Kind != construction.OperatorNegation || operator.Operator.Value != construction.ValueNegated {
			candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "UNSUPPORTED_SOURCE_OPERATOR_VALUE", CandidateID: vp.ID, Message: "only explicitly negated operator evidence is supported by this profile"})
			continue
		}
		mapped := false
		for _, row := range plan.Operators {
			if row.SourceKind == string(operator.Operator.Kind) && row.Operator == algebra.TermNot {
				*predicate = &algebra.Term{Kind: algebra.TermNot, Body: *predicate}
				mapped = true
				break
			}
		}
		if !mapped {
			candidate.Residuals = append(candidate.Residuals, DSEResidual{Code: "UNMAPPED_SOURCE_OPERATOR", CandidateID: vp.ID, Path: fmt.Sprintf("tokens[%d:%d]", operator.Span.Start, operator.Span.End), Message: "source operator has no explicit supported algebra mapping"})
		}
	}
	return nil
}

func mergeDSECandidateAttributes(document *algebra.Document, candidate parser.PhraseCandidate, referenceID string, evidence map[string]NodeAttributesEvidence) error {
	row, ok := evidence[candidate.ID]
	if !ok {
		return nil
	}
	for i := range document.References {
		if document.References[i].ID != referenceID {
			continue
		}
		current := &document.References[i]
		if row.TemporalFrame != "" {
			if current.TemporalFrame != "" && current.TemporalFrame != string(row.TemporalFrame) {
				return dseError(status.OutcomeInvalid, "CONFLICTING_TEMPORAL_FRAME", "references."+referenceID, "source evidence conflicts with caller-provided temporal frame")
			}
			current.TemporalFrame = string(row.TemporalFrame)
		}
		if row.Multiplicity != "" {
			if current.Multiplicity != "" && current.Multiplicity != string(row.Multiplicity) {
				return dseError(status.OutcomeInvalid, "CONFLICTING_MULTIPLICITY", "references."+referenceID, "source evidence conflicts with caller-provided multiplicity")
			}
			current.Multiplicity = string(row.Multiplicity)
		}
		return nil
	}
	for i := range document.RequestContext.Referents {
		if document.RequestContext.Referents[i].ID != referenceID {
			continue
		}
		current := &document.RequestContext.Referents[i]
		if row.TemporalFrame != "" {
			if current.TemporalFrame != "" && current.TemporalFrame != string(row.TemporalFrame) {
				return dseError(status.OutcomeInvalid, "CONFLICTING_TEMPORAL_FRAME", "request_context.referents."+referenceID, "source evidence conflicts with caller-provided temporal frame")
			}
			current.TemporalFrame = string(row.TemporalFrame)
		}
		if row.Multiplicity != "" {
			if current.Multiplicity != "" && current.Multiplicity != string(row.Multiplicity) {
				return dseError(status.OutcomeInvalid, "CONFLICTING_MULTIPLICITY", "request_context.referents."+referenceID, "source evidence conflicts with caller-provided multiplicity")
			}
			current.Multiplicity = string(row.Multiplicity)
		}
		return nil
	}
	return dseError(status.OutcomeInvalid, "ATTRIBUTE_REFERENCE_UNDECLARED", "references."+referenceID, "source attribute candidate has no matching declared reference")
}

func validateDSEContext(input DSEInput) error {
	refs := input.ReferenceContext
	if refs.RequestContext != nil {
		cloned, err := algebra.CloneRequestContext(*refs.RequestContext)
		if err != nil {
			return err
		}
		refs.RequestContext = &cloned
	}
	doc := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: cloneAlgebraSorts(input.Selected.Sorts), Signatures: cloneAlgebraSignatures(input.Selected.Signatures), References: cloneAlgebraDeclarations(refs.References), Limits: input.Selected.Limits, RequestContext: cloneDSEContext(refs.RequestContext)}
	if err := algebra.CheckDeclarations(doc); err != nil {
		return err
	}
	if refs.RequestContext != nil {
		if _, err := algebra.CheckRequestContext(*refs.RequestContext, doc); err != nil {
			return err
		}
	}
	if input.Rewrite.SnapshotID == "" || input.Rewrite.Document.SchemaVersion != algebra.SchemaVersion {
		return dseError(status.OutcomeInvalid, "INVALID_REWRITE_ENVIRONMENT", "rewrite", "rewrite environment must be an explicitly selected checked algebra snapshot")
	}
	return nil
}

type dseMentionReference struct {
	status       string
	referenceIDs []string
}

func dseMentionReferences(input DSEInput) (map[string]dseMentionReference, error) {
	out := make(map[string]dseMentionReference)
	if input.ReferenceContext.RequestContext == nil {
		return out, nil
	}
	doc := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: cloneAlgebraSorts(input.Selected.Sorts), Signatures: cloneAlgebraSignatures(input.Selected.Signatures), References: cloneAlgebraDeclarations(input.ReferenceContext.References), Limits: input.Selected.Limits, RequestContext: cloneDSEContext(input.ReferenceContext.RequestContext)}
	assessment, err := algebra.CheckRequestContext(*doc.RequestContext, doc)
	if err != nil {
		return nil, err
	}
	for _, resolution := range assessment.Mentions {
		out[resolution.MentionID] = dseMentionReference{status: string(resolution.Status), referenceIDs: append([]string{}, resolution.CandidateReferentIDs...)}
	}
	return out, nil
}

// resolveDSEReference gives corroborated mention/evidence identity precedence.
// A role mapping is an explicit caller assumption only when source evidence
// supplies no resolved identity; it cannot override a resolved mention.
func resolveDSEReference(candidateID, fallback string, hasFallback bool, evidence map[string]ReferenceEvidence, mentions map[string]dseMentionReference) (string, bool, bool, string, error) {
	resolvedID, basis := "", ""
	if row, ok := evidence[candidateID]; ok && row.ReferentID != "" && row.EvidenceID != "" {
		resolvedID, basis = string(row.ReferentID), "source_identity_evidence"
	}
	if mention, ok := mentions[candidateID]; ok {
		switch mention.status {
		case string(algebra.MentionResolved):
			if len(mention.referenceIDs) != 1 {
				return "", false, false, "", dseError(status.OutcomeInvalid, "INVALID_RESOLVED_MENTION", "request_context.mentions."+candidateID, "resolved mention must have one reference ID")
			}
			if resolvedID != "" && resolvedID != mention.referenceIDs[0] {
				return "", false, false, "", dseError(status.OutcomeInvalid, "SOURCE_IDENTITY_CONFLICT", "request_context.mentions."+candidateID, "source evidence and request-context resolution disagree")
			}
			if resolvedID != "" {
				basis = "source_and_context_resolution"
			} else {
				basis = "request_context_resolution"
			}
			resolvedID = mention.referenceIDs[0]
		case string(algebra.MentionAmbiguous):
			if !hasFallback {
				return "", false, true, "", nil
			}
		}
	}
	if resolvedID != "" {
		if hasFallback && fallback != resolvedID {
			return "", false, false, "", dseError(status.OutcomeInvalid, "ROLE_REFERENCE_CONFLICT", "reference_context.role_references."+candidateID, "explicit role mapping conflicts with resolved mention identity")
		}
		return resolvedID, true, false, basis, nil
	}
	if hasFallback {
		return fallback, true, false, "caller_role_mapping_assumption", nil
	}
	if row, ok := evidence[candidateID]; ok && row.Anchor != nil {
		return "", false, true, "", nil
	}
	return "", false, false, "", nil
}

func sourcePredicateMapping(doc language.SourcePlanDocument, sense, frame string) (language.SourcePredicateMapping, bool) {
	for _, row := range doc.Predicates {
		if row.SenseID.String() == sense && string(row.FrameID) == frame {
			return row, true
		}
	}
	return language.SourcePredicateMapping{}, false
}
func dsePermissionMatches(plan language.SourcePlanDocument, alt parser.ParseAlternative) bool {
	_, ok := sourcePredicateMapping(plan, alt.Permission.SenseID.String(), string(alt.Permission.FrameID))
	if !ok {
		return false
	}
	for _, p := range plan.Predicates {
		if p.SenseID == alt.Permission.SenseID && p.FrameID == alt.Permission.FrameID {
			return containsConstructionID(p.AllowedFamilyIDs, alt.FamilyID) && (len(p.AllowedVariantIDs) == 0 || containsVariantID(p.AllowedVariantIDs, alt.VariantID))
		}
	}
	return false
}
func containsConstructionID(values []ids.ConstructionID, id ids.ConstructionID) bool {
	for _, v := range values {
		if v == id {
			return true
		}
	}
	return false
}
func containsVariantID(values []construction.VariantID, id construction.VariantID) bool {
	for _, v := range values {
		if v == id {
			return true
		}
	}
	return false
}
func chargeDSEWork(result *DSEResult, work *int, limit int) bool {
	if *work >= limit {
		result.Work = *work
		return false
	}
	*work++
	result.Work = *work
	return true
}
func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func appendUniqueString(values []string, value string) []string {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}
func candidateHasSense(candidate parser.PhraseCandidate, sense string) bool {
	for _, v := range candidate.SenseIDs {
		if v.String() == sense {
			return true
		}
	}
	return false
}
func sourceRoleReference(context language.SourceReferenceContext, role string) (string, bool) {
	value := ""
	found := false
	for _, row := range context.RoleReferences {
		if row.Role == role {
			if found {
				return "", false
			}
			value = row.ReferenceID
			found = true
		}
	}
	return value, found
}
func sourcePortReference(context language.SourceReferenceContext, port string) (string, bool) {
	value := ""
	found := false
	for _, row := range context.PortReferences {
		if row.PortID == port {
			if found {
				return "", false
			}
			value = row.ReferenceID
			found = true
		}
	}
	return value, found
}
func dseRecord(candidate parser.PhraseCandidate, alternative parser.ParseAlternative) DSESourceRecord {
	reading := ""
	if len(candidate.ReadingIDs) > 0 {
		reading = candidate.ReadingIDs[0]
	}
	return DSESourceRecord{CandidateID: candidate.ID, ReadingID: reading, ReadingIDs: append([]string{}, candidate.ReadingIDs...), SenseID: alternative.Permission.SenseID.String(), FrameID: string(alternative.Permission.FrameID), FamilyID: alternative.FamilyID.String(), VariantID: alternative.VariantID.String(), DerivationID: alternative.DerivationID, Span: candidate.Span}
}
func dseRoleRecord(candidate parser.PhraseCandidate, sense string) DSESourceRecord {
	reading := ""
	if len(candidate.ReadingIDs) > 0 {
		reading = candidate.ReadingIDs[0]
	}
	return DSESourceRecord{CandidateID: candidate.ID, ReadingID: reading, ReadingIDs: append([]string{}, candidate.ReadingIDs...), SenseID: sense, Span: candidate.Span}
}
func dseCandidateStart(candidate parser.PhraseCandidate, alternative parser.ParseAlternative) DSECandidate {
	return DSECandidate{ID: dseStableID(candidate.ID, alternative.DerivationID, alternative.Permission.SenseID.String()), Status: status.OutcomeOK, Source: dseRecord(candidate, alternative), RoleSources: []DSERoleSource{}, GoalSpecs: []json.RawMessage{}, FragmentChecks: []DSEFragmentCheck{}, ConditionCoverage: []DSEConditionCoverage{}, Residuals: []DSEResidual{}, Scope: "typed representation; no world-fact claim"}
}
func incompleteDSE(candidate parser.PhraseCandidate, alternative parser.ParseAlternative, code, message string) (DSECandidate, bool, error) {
	out := dseCandidateStart(candidate, alternative)
	out.Status = status.OutcomeIncomplete
	out.Residuals = append(out.Residuals, DSEResidual{Code: code, CandidateID: candidate.ID, Message: boundedDSEMessage(message)})
	return out, false, nil
}
func appendDSEIncomplete(result *DSEResult, id string, source DSESourceRecord, code, message string, budget DSEBudget, work int) bool {
	candidate := DSECandidate{ID: dseStableID(id, code), Status: status.OutcomeIncomplete, Source: source, GoalSpecs: []json.RawMessage{}, FragmentChecks: []DSEFragmentCheck{}, ConditionCoverage: []DSEConditionCoverage{}, Residuals: []DSEResidual{{Code: code, CandidateID: source.CandidateID, Message: message}}, Scope: "unresolved source interpretation; no meaning selected"}
	if len(result.Candidates) >= budget.MaxAlternatives || len(result.Candidates) >= MaxDSEAlternatives {
		*result = dseResourcePrefix(*result, "DSE_ALTERNATIVE_LIMIT", candidate.ID, work)
		return true
	}
	encoded, err := json.Marshal(candidate)
	if err != nil {
		*result = dseResourcePrefix(*result, "DSE_CANDIDATE_ENCODING", candidate.ID, work)
		return true
	}
	if len(encoded) > budget.MaxOutputBytes-result.OutputBytes {
		*result = dseResourcePrefix(*result, "DSE_OUTPUT_LIMIT", candidate.ID, work)
		return true
	}
	result.OutputBytes += len(encoded)
	result.Candidates = append(result.Candidates, candidate)
	result.Complete = false
	return false
}
func dseStableID(parts ...string) string {
	h := sha256.Sum256([]byte(fmt.Sprint(parts)))
	return "dse:" + hex.EncodeToString(h[:12])
}
func dseConditionID(condition algebra.ContextCondition) string {
	copy := condition
	copy.ID = ""
	encoded, _ := json.Marshal(copy)
	sum := sha256.Sum256(encoded)
	return "condition:" + hex.EncodeToString(sum[:12])
}
func addDSECondition(request *algebra.RequestContext, condition algebra.ContextCondition) string {
	if request == nil {
		return condition.ID
	}
	content := condition
	content.ID = ""
	encoded, _ := json.Marshal(content)
	for _, existing := range request.Conditions {
		old := existing
		old.ID = ""
		oldEncoded, _ := json.Marshal(old)
		if string(oldEncoded) == string(encoded) {
			return existing.ID
		}
	}
	request.Conditions = append(request.Conditions, condition)
	return condition.ID
}
func dseResourcePrefix(result DSEResult, code, path string, work int) DSEResult {
	result.Status = status.OutcomeResourceLimit
	result.Complete = false
	result.Work = work
	result.Residuals = append(result.Residuals, DSEResidual{Code: code, Path: path, Message: "finite source composition budget was exhausted; retained prefix is incomplete"})
	return result
}
func dseError(kind status.OutcomeKind, code, path, message string) error {
	diag := status.Diagnostic{Code: code, Path: path, Message: boundedDSEMessage(message), Severity: status.SeverityError}
	res, err := status.NewResult(kind, "", []status.Diagnostic{diag}, nil)
	if err != nil {
		return fmt.Errorf("DSE failure: %s", message)
	}
	return &DSEError{result: res}
}
func boundedDSEMessage(message string) string {
	if len(message) > status.MaxDiagnosticMessageBytes {
		limit := status.MaxDiagnosticMessageBytes
		for limit > 0 && !utf8.ValidString(message[:limit]) {
			limit--
		}
		return message[:limit]
	}
	return message
}
func cloneDSEContext(source *algebra.RequestContext) *algebra.RequestContext {
	if source == nil {
		return nil
	}
	copy, err := algebra.CloneRequestContext(*source)
	if err != nil {
		return nil
	}
	return &copy
}
func cloneAlgebraDeclarations(source []algebra.Declaration) []algebra.Declaration {
	out := append([]algebra.Declaration{}, source...)
	for i := range out {
		out[i].Type = cloneDSEType(out[i].Type)
	}
	return out
}
func cloneAlgebraSorts(source []algebra.SortDeclaration) []algebra.SortDeclaration {
	return append([]algebra.SortDeclaration{}, source...)
}
func cloneAlgebraSignatures(source []algebra.Signature) []algebra.Signature {
	out := make([]algebra.Signature, len(source))
	for i, s := range source {
		out[i] = s
		out[i].Ports = append([]algebra.Port{}, s.Ports...)
		for j := range out[i].Ports {
			out[i].Ports[j].Type = cloneDSEType(out[i].Ports[j].Type)
		}
	}
	return out
}
func cloneDSEType(source algebra.Type) algebra.Type {
	out := source
	if source.From != nil {
		x := cloneDSEType(*source.From)
		out.From = &x
	}
	if source.To != nil {
		x := cloneDSEType(*source.To)
		out.To = &x
	}
	return out
}

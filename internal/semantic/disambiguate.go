package semantic

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// MeaningCandidate is one graph or evidence-only semantic interpretation.
// SemanticID identifies equivalent meanings; ID identifies this derivation.
type MeaningCandidate struct {
	ID                   string                       `json:"id"`
	SemanticID           string                       `json:"semantic_id"`
	Meaning              *meaning.Meaning             `json:"meaning,omitempty"`
	Contexts             []MeaningContext             `json:"request_contexts,omitempty"`
	Status               status.OutcomeKind           `json:"status"`
	Evidence             EvidenceTrace                `json:"evidence"`
	SoftConstraintMisses []valency.SoftConstraintMiss `json:"soft_constraint_misses,omitempty"`
	DerivationIDs        []string                     `json:"derivation_ids,omitempty"`
}

// MeaningContext keeps each derivation's request context beside the exact
// graph whose local node IDs it addresses.
type MeaningContext struct {
	Meaning meaning.Meaning        `json:"meaning"`
	Context meaning.RequestContext `json:"context"`
}

// Result retains semantic alternatives without choosing among them.
type Result struct {
	Outcome     status.OutcomeKind  `json:"outcome"`
	Candidates  []MeaningCandidate  `json:"candidates"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// DisambiguationError reports invalid inputs or exhausted retention budgets as
// a typed status result.
type DisambiguationError struct {
	result status.Result
}

func (err *DisambiguationError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "semantic disambiguation error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns the typed error outcome with a defensive diagnostics copy.
func (err *DisambiguationError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

// Disambiguate merges candidates that have the same semantic identity while
// retaining distinct meanings and all available evidence. It never ranks or
// selects one interpretation over another.
func Disambiguate(candidates []MeaningCandidate, inputOutcome status.OutcomeKind, diagnostics []status.Diagnostic, ontology *core.ConceptGraph, budget EvidenceBudget) (Result, error) {
	if err := budget.Validate(); err != nil {
		return Result{}, err
	}
	if !inputOutcome.Valid() {
		return Result{}, disambiguationError("INVALID_INPUT_OUTCOME", "input_outcome", "input outcome is unsupported")
	}
	if ontology == nil {
		return Result{}, disambiguationError("MISSING_CORE_GRAPH", "core", "a CORE graph is required to verify candidate meanings")
	}
	if len(candidates) > budget.MaxRecords || len(candidates) > maxEvidenceRecords {
		return Result{}, disambiguationError("RESOURCE_LIMIT", "candidates", "candidate count exceeds the disambiguation budget")
	}
	if len(diagnostics) > status.MaxDiagnostics {
		return Result{}, disambiguationError("RESOURCE_LIMIT", "diagnostics", "diagnostic count exceeds the status limit")
	}
	for i, diagnostic := range diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return Result{}, disambiguationError("INVALID_DIAGNOSTIC", fmt.Sprintf("diagnostics[%d]", i), err.Error())
		}
	}

	// Validate and account for the complete input before allocating retained
	// copies. The aggregate ceiling prevents many individually valid traces from
	// multiplying the request's memory use.
	var records, graphItems, derivationItems, missItems, textBytes int
	textBytes = diagnosticTextBytes(diagnostics)
	if textBytes > budget.MaxTextBytes {
		return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.text_bytes", "diagnostics exceed the text budget")
	}
	seenIDs := make(map[string]struct{}, len(candidates))
	for i := range candidates {
		candidate := &candidates[i]
		if !boundedID(candidate.ID, 256) || !boundedID(candidate.SemanticID, 256) {
			return Result{}, disambiguationError("INVALID_CANDIDATE_ID", fmt.Sprintf("candidates[%d]", i), "candidate and semantic IDs must be non-empty bounded UTF-8")
		}
		if _, duplicate := seenIDs[candidate.ID]; duplicate {
			return Result{}, disambiguationError("DUPLICATE_CANDIDATE_ID", fmt.Sprintf("candidates[%d].id", i), "candidate IDs must be unique")
		}
		seenIDs[candidate.ID] = struct{}{}
		if !candidate.Status.Valid() {
			return Result{}, disambiguationError("INVALID_CANDIDATE_STATUS", fmt.Sprintf("candidates[%d].status", i), "candidate status is unsupported")
		}
		if candidate.Status == status.OutcomeOK && candidate.Meaning == nil {
			return Result{}, disambiguationError("MISSING_MEANING_GRAPH", fmt.Sprintf("candidates[%d].meaning", i), "an OK candidate requires a validated meaning graph")
		}
		if candidate.Meaning != nil {
			if err := meaning.ValidateMeaning(*candidate.Meaning, ontology, meaning.DefaultBudget()); err != nil {
				return Result{}, disambiguationError("INVALID_MEANING_GRAPH", fmt.Sprintf("candidates[%d].meaning", i), err.Error())
			}
			if len(candidate.Contexts) == 0 {
				return Result{}, disambiguationError("MISSING_MEANING_CONTEXT", fmt.Sprintf("candidates[%d].request_contexts", i), "a meaning graph requires its request-local reference context")
			}
			var representativeContext *MeaningContext
			representativeIndex := -1
			for contextIndex, context := range candidate.Contexts {
				if err := meaning.ValidateMeaning(context.Meaning, ontology, meaning.DefaultBudget()); err != nil {
					return Result{}, disambiguationError("INVALID_CONTEXT_MEANING", fmt.Sprintf("candidates[%d].request_contexts[%d].meaning", i, contextIndex), err.Error())
				}
				if err := meaning.ValidateReferences(context.Meaning, context.Context, meaning.DefaultBudget()); err != nil {
					return Result{}, disambiguationError("INVALID_MEANING_CONTEXT", fmt.Sprintf("candidates[%d].request_contexts[%d]", i, contextIndex), err.Error())
				}
				structurallyEqual, _ := meaning.CanonicalEqual(*candidate.Meaning, context.Meaning, meaning.DefaultBudget())
				if structurallyEqual && representativeContext == nil {
					contextCopy := context
					representativeContext = &contextCopy
					representativeIndex = contextIndex
				}
				contextItems, contextText := contextSize(context.Context)
				graphCount, graphText, exceeds := meaningSize(context.Meaning, budget.MaxRecords-graphItems)
				if exceeds {
					return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.context_meaning", "aggregate context meaning graphs exceed the graph budget")
				}
				graphItems += graphCount
				textBytes += graphText
				graphItems += contextItems
				textBytes += contextText
				if graphItems > budget.MaxRecords || textBytes > budget.MaxTextBytes {
					return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.request_context", "aggregate request context exceeds the semantic budget")
				}
			}
			if representativeContext == nil {
				return Result{}, disambiguationError("CONTEXT_MEANING_MISMATCH", fmt.Sprintf("candidates[%d].request_contexts", i), "candidate graph must be retained beside at least one context that identifies its local referent nodes")
			}
			for contextIndex, context := range candidate.Contexts {
				if contextIndex == representativeIndex {
					// Canonical structural equality and ValidateReferences above
					// establish this exact graph/context pair. Requiring semantic
					// identity here would incorrectly reject unresolved referents.
					continue
				}
				equivalent, err := meaning.SemanticEquivalence(representativeContext.Meaning, context.Meaning, representativeContext.Context, context.Context, meaning.DefaultBudget())
				if err != nil || !equivalent {
					return Result{}, disambiguationError("CONTEXT_MEANING_MISMATCH", fmt.Sprintf("candidates[%d].request_contexts[%d]", i, contextIndex), "every retained request context must identify the candidate's semantic meaning")
				}
			}
			items, graphText, exceeds := meaningSize(*candidate.Meaning, budget.MaxRecords-graphItems)
			if exceeds {
				return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.meaning_nodes", "aggregate meaning graph size exceeds the record budget")
			}
			graphItems += items
			if graphItems > budget.MaxRecords {
				return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.meaning_nodes", "aggregate meaning graph size exceeds the record budget")
			}
			textBytes += graphText
			if textBytes > budget.MaxTextBytes {
				return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.text_bytes", "meaning graph text exceeds the text budget")
			}
		}
		if err := candidate.Evidence.ValidateWithBudget(budget); err != nil {
			return Result{}, fmt.Errorf("candidate %q evidence: %w", candidate.ID, err)
		}
		if candidate.Evidence.Outcome != candidate.Status {
			return Result{}, disambiguationError("EVIDENCE_STATUS_MISMATCH", fmt.Sprintf("candidates[%d].evidence.outcome", i), "evidence outcome must match candidate status")
		}
		if len(candidate.SoftConstraintMisses) > budget.MaxRecords || len(candidate.DerivationIDs) > budget.MaxRecords {
			return Result{}, disambiguationError("RESOURCE_LIMIT", fmt.Sprintf("candidates[%d]", i), "candidate metadata exceeds the record budget")
		}
		derivationItems += len(candidate.DerivationIDs) + 1 // include the alternative identity retained below
		missItems += len(candidate.SoftConstraintMisses)
		if derivationItems > budget.MaxRecords || missItems > budget.MaxRecords {
			return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.candidate_metadata", "aggregate candidate metadata exceeds the record budget")
		}
		records += len(candidate.Evidence.Records)
		if records > budget.MaxRecords {
			return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.records", "aggregate evidence record count exceeds the budget")
		}
		textBytes += candidateTextBytes(*candidate)
		if textBytes > budget.MaxTextBytes {
			return Result{}, disambiguationError("RESOURCE_LIMIT", "budget.text_bytes", "aggregate candidate text exceeds the budget")
		}
		for j, id := range candidate.DerivationIDs {
			if !boundedID(id, 256) {
				return Result{}, disambiguationError("INVALID_DERIVATION_ID", fmt.Sprintf("candidates[%d].derivation_ids[%d]", i, j), "derivation ID must be non-empty bounded UTF-8")
			}
		}
		for j, miss := range candidate.SoftConstraintMisses {
			if !boundedID(miss.ConstraintID, 256) || !boundedID(string(miss.Role), 256) || !boundedID(miss.Actual.String(), 256) {
				return Result{}, disambiguationError("INVALID_SOFT_CONSTRAINT_MISS", fmt.Sprintf("candidates[%d].soft_constraint_misses[%d]", i, j), "soft constraint miss has an invalid identity")
			}
		}
	}

	// Stable ordering makes the merged output independent of map iteration and
	// ensures the selected representative is deterministic.
	ordered := append([]MeaningCandidate(nil), candidates...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].SemanticID != ordered[j].SemanticID {
			return ordered[i].SemanticID < ordered[j].SemanticID
		}
		if ordered[i].ID != ordered[j].ID {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].Evidence.AlternativeID < ordered[j].Evidence.AlternativeID
	})

	merged := make([]MeaningCandidate, 0, len(ordered))
	bySemantic := make(map[string][]int, len(ordered))
	for _, source := range ordered {
		candidate := cloneMeaningCandidate(source)
		candidate.DerivationIDs = appendUnique(candidate.DerivationIDs, candidate.ID)
		candidate.DerivationIDs = appendUnique(candidate.DerivationIDs, candidate.Evidence.AlternativeID)
		candidate.SoftConstraintMisses = uniqueSoftMisses(candidate.SoftConstraintMisses)
		position := -1
		for _, possible := range bySemantic[candidate.SemanticID] {
			current := merged[possible]
			if current.Meaning == nil || candidate.Meaning == nil {
				continue
			}
			if contextsEquivalent(current.Contexts, candidate.Contexts) {
				position = possible
				break
			}
		}
		if position < 0 {
			bySemantic[candidate.SemanticID] = append(bySemantic[candidate.SemanticID], len(merged))
			merged = append(merged, candidate)
			continue
		}
		current := &merged[position]
		current.Status = mergeOutcome(current.Status, candidate.Status)
		if current.Meaning == nil && candidate.Meaning != nil {
			current.Meaning = candidate.Meaning
		}
		current.DerivationIDs = appendUnique(current.DerivationIDs, candidate.ID)
		current.DerivationIDs = appendUnique(current.DerivationIDs, candidate.Evidence.AlternativeID)
		for _, id := range candidate.DerivationIDs {
			current.DerivationIDs = appendUnique(current.DerivationIDs, id)
		}
		for _, miss := range candidate.SoftConstraintMisses {
			current.SoftConstraintMisses = appendSoftMissUnique(current.SoftConstraintMisses, miss)
		}
		for _, context := range candidate.Contexts {
			current.Contexts = appendContextUnique(current.Contexts, context)
		}
		mergeEvidence(&current.Evidence, candidate.Evidence, current.SemanticID)
		current.Evidence.Outcome = current.Status
		current.Evidence.Diagnostics = mergeDiagnostics(current.Evidence.Diagnostics, candidate.Evidence.Diagnostics)
	}

	resultDiagnostics := mergeDiagnostics(nil, diagnostics)
	for _, candidate := range merged {
		resultDiagnostics = mergeDiagnostics(resultDiagnostics, candidate.Evidence.Diagnostics)
	}
	outcome := resultOutcome(inputOutcome, candidates, merged)
	if outcome != status.OutcomeOK && len(resultDiagnostics) == 0 {
		resultDiagnostics = append(resultDiagnostics, status.Diagnostic{
			Code: "DISAMBIGUATION_INCOMPLETE", Message: "semantic alternatives remain unresolved", Severity: status.SeverityWarning,
		})
	}
	return Result{Outcome: outcome, Candidates: merged, Diagnostics: resultDiagnostics}, nil
}

func cloneMeaningCandidate(source MeaningCandidate) MeaningCandidate {
	copy := source
	copy.Meaning = cloneMeaning(source.Meaning)
	copy.Evidence = source.Evidence.Clone()
	copy.Contexts = make([]MeaningContext, len(source.Contexts))
	for index, context := range source.Contexts {
		copy.Contexts[index] = MeaningContext{Meaning: *cloneMeaning(&context.Meaning), Context: cloneContext(context.Context)}
	}
	copy.SoftConstraintMisses = append([]valency.SoftConstraintMiss(nil), source.SoftConstraintMisses...)
	copy.DerivationIDs = append([]string(nil), source.DerivationIDs...)
	return copy
}

func appendContextUnique(contexts []MeaningContext, value MeaningContext) []MeaningContext {
	for _, context := range contexts {
		if reflect.DeepEqual(context, value) {
			return contexts
		}
	}
	return append(contexts, MeaningContext{Meaning: *cloneMeaning(&value.Meaning), Context: cloneContext(value.Context)})
}

func contextsEquivalent(leftContexts, rightContexts []MeaningContext) bool {
	for _, left := range leftContexts {
		for _, right := range rightContexts {
			equivalent, err := meaning.SemanticEquivalence(left.Meaning, right.Meaning, left.Context, right.Context, meaning.DefaultBudget())
			if err == nil && equivalent {
				return true
			}
		}
	}
	return false
}

func contextSize(context meaning.RequestContext) (items, textBytes int) {
	items = len(context.Referents) + len(context.Mentions)
	textBytes = len(context.ID)
	for _, referent := range context.Referents {
		textBytes += len(referent.ID) + len(referent.Concept)
	}
	for _, mention := range context.Mentions {
		textBytes += len(mention.ID) + len(mention.NodeID) + len(mention.ReferentID) + len(mention.EvidenceID)
		if mention.Anchor != nil {
			textBytes += len(mention.Anchor.Kind) + len(mention.Anchor.Value)
		}
	}
	return items, textBytes
}

func cloneContext(source meaning.RequestContext) meaning.RequestContext {
	copy := source
	copy.Referents = append([]meaning.Referent(nil), source.Referents...)
	copy.Mentions = make([]meaning.Mention, len(source.Mentions))
	for index, mention := range source.Mentions {
		copy.Mentions[index] = mention
		if mention.Anchor != nil {
			anchor := *mention.Anchor
			copy.Mentions[index].Anchor = &anchor
		}
	}
	return copy
}

func cloneMeaning(source *meaning.Meaning) *meaning.Meaning {
	if source == nil {
		return nil
	}
	copy := *source
	copy.Roots = append([]meaning.NodeID(nil), source.Roots...)
	copy.Nodes = make([]meaning.Node, len(source.Nodes))
	for i, node := range source.Nodes {
		copy.Nodes[i] = node
		copy.Nodes[i].Participants = append([]meaning.Participant(nil), node.Participants...)
		copy.Nodes[i].Content = append([]meaning.NodeID(nil), node.Content...)
		if node.Quantifier != nil {
			quantifier := *node.Quantifier
			if node.Quantifier.Count != nil {
				count := *node.Quantifier.Count
				quantifier.Count = &count
			}
			copy.Nodes[i].Quantifier = &quantifier
		}
	}
	return &copy
}

func mergeEvidence(target *EvidenceTrace, source EvidenceTrace, semanticID string) {
	seen := make(map[EvidenceRecord]struct{}, len(target.Records)+len(source.Records))
	for _, record := range target.Records {
		seen[record] = struct{}{}
	}
	for _, record := range source.Records {
		if _, exists := seen[record]; exists {
			continue
		}
		seen[record] = struct{}{}
		target.Records = append(target.Records, record)
	}
	target.AlternativeID = semanticID
}

func mergeDiagnostics(target, source []status.Diagnostic) []status.Diagnostic {
	for _, diagnostic := range source {
		found := false
		for _, existing := range target {
			if existing == diagnostic {
				found = true
				break
			}
		}
		if !found && len(target) < status.MaxDiagnostics {
			target = append(target, diagnostic)
		}
	}
	return target
}

func resultOutcome(input status.OutcomeKind, inputCandidates, merged []MeaningCandidate) status.OutcomeKind {
	candidateOutcome := status.OutcomeOK
	completeSemanticIDs := make(map[string]struct{}, len(inputCandidates))
	unsupported := false
	for _, candidate := range inputCandidates {
		switch candidate.Status {
		case status.OutcomeOK:
			completeSemanticIDs[candidate.SemanticID] = struct{}{}
		case status.OutcomeUnsupported:
			unsupported = true
			candidateOutcome = mergeOutcome(candidateOutcome, status.OutcomeUnsupported)
		case status.OutcomeUnknown:
			candidateOutcome = mergeOutcome(candidateOutcome, status.OutcomeUnknown)
		case status.OutcomeIncomplete:
			candidateOutcome = mergeOutcome(candidateOutcome, status.OutcomeIncomplete)
		case status.OutcomeInvalid, status.OutcomeResourceLimit, status.OutcomeError:
			candidateOutcome = mergeOutcome(candidateOutcome, candidate.Status)
		case status.OutcomeAmbiguous:
			candidateOutcome = mergeOutcome(candidateOutcome, status.OutcomeAmbiguous)
		}
	}
	if unsupported && len(completeSemanticIDs) > 0 && candidateOutcome != status.OutcomeInvalid && candidateOutcome != status.OutcomeResourceLimit && candidateOutcome != status.OutcomeError {
		candidateOutcome = status.OutcomeIncomplete
	}
	if candidateOutcome == status.OutcomeOK {
		if len(completeSemanticIDs) == 0 {
			candidateOutcome = status.OutcomeUnknown
		} else if len(completeSemanticIDs) > 1 || len(merged) > 1 {
			candidateOutcome = status.OutcomeAmbiguous
		}
	}
	return mergeOutcome(input, candidateOutcome)
}

func mergeOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	if left == right {
		return left
	}
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

func candidateTextBytes(candidate MeaningCandidate) int {
	count := len(candidate.ID) + len(candidate.SemanticID) + len(candidate.Evidence.AlternativeID)
	for _, id := range candidate.DerivationIDs {
		count += len(id)
	}
	for _, miss := range candidate.SoftConstraintMisses {
		count += len(miss.ConstraintID) + len(miss.Role) + len(miss.Actual.String())
	}
	for _, record := range candidate.Evidence.Records {
		for _, value := range evidenceRecordStrings(record) {
			count += len(value)
		}
	}
	count += diagnosticTextBytes(candidate.Evidence.Diagnostics)
	return count
}

func meaningSize(graph meaning.Meaning, limit int) (items, textBytes int, exceeds bool) {
	items = len(graph.Roots) + len(graph.Nodes)
	if items > limit {
		return items, 0, true
	}
	for _, root := range graph.Roots {
		textBytes += len(root)
	}
	for _, node := range graph.Nodes {
		textBytes += len(node.ID) + len(node.Kind) + len(node.Concept) + len(node.Target)
		textBytes += len(node.Operator) + len(node.Modality) + len(node.Act)
		items += len(node.Participants) + len(node.Content)
		if items > limit {
			return items, textBytes, true
		}
		for _, participant := range node.Participants {
			textBytes += len(participant.Role) + len(participant.Target)
		}
		for _, target := range node.Content {
			textBytes += len(target)
		}
		if node.Quantifier != nil {
			textBytes += len(node.Quantifier.Kind)
			items++
			if items > limit {
				return items, textBytes, true
			}
		}
	}
	return items, textBytes, false
}

func diagnosticTextBytes(diagnostics []status.Diagnostic) int {
	total := 0
	for _, diagnostic := range diagnostics {
		total += len(diagnostic.Code) + len(diagnostic.Message) + len(diagnostic.Path)
	}
	return total
}

func boundedID(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && validEvidenceText(value, maximum)
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendSoftMissUnique(values []valency.SoftConstraintMiss, value valency.SoftConstraintMiss) []valency.SoftConstraintMiss {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func uniqueSoftMisses(values []valency.SoftConstraintMiss) []valency.SoftConstraintMiss {
	result := make([]valency.SoftConstraintMiss, 0, len(values))
	for _, value := range values {
		result = appendSoftMissUnique(result, value)
	}
	return result
}

func disambiguationError(code, path, message string) error {
	kind := status.OutcomeInvalid
	if code == "RESOURCE_LIMIT" {
		kind = status.OutcomeResourceLimit
	}
	return &DisambiguationError{result: status.Result{
		Status: kind,
		Diagnostics: []status.Diagnostic{{
			Code: code, Path: path, Message: message, Severity: status.SeverityError,
		}},
	}}
}

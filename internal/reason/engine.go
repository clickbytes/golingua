package reason

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxDerivations = 4096
	MaxWork        = 100_000
	MaxDepth       = 32
	MaxProofFacts  = 64
)

type Budget struct {
	MaxDerivations int
	MaxWork        int
	MaxDepth       int
}

func DefaultBudget() Budget {
	return Budget{MaxDerivations: MaxDerivations, MaxWork: MaxWork, MaxDepth: MaxDepth}
}

// Query asks about two known request referents at one explicit UTC instant.
// An empty At reads only timeless facts, never a time-bounded assertion.
type Query struct {
	Predicate string                `json:"predicate"`
	Args      [2]meaning.ReferentID `json:"args"`
	At        string                `json:"at,omitempty"`
}

// Evaluate computes the bounded positive closure of explicit request facts.
// Negative evidence remains direct; rules never derive negation. A completed
// closure yields true, false, conflict or unknown with deterministic proof.
func Evaluate(set *facts.Set, rules []Rule, query Query, budget Budget) (Result, error) {
	if set == nil || set.CoreGraph() == nil {
		return Result{}, reasonInvalid("FACT_SET_UNAVAILABLE", "facts", "reasoner requires a validated request FactSet")
	}
	if budget == (Budget{}) {
		budget = DefaultBudget()
	}
	if budget.MaxDerivations <= 0 || budget.MaxDerivations > MaxDerivations || budget.MaxWork <= 0 || budget.MaxWork > MaxWork || budget.MaxDepth <= 0 || budget.MaxDepth > MaxDepth {
		return Result{}, reasonInvalid("INVALID_REASON_BUDGET", "budget", "reason budget must be positive and within package ceilings")
	}
	instant, err := validateQuery(set, query)
	if err != nil {
		return Result{}, err
	}
	compiled, err := compileRules(set.CoreGraph(), rules)
	if err != nil {
		return Result{}, err
	}
	allFacts := set.Facts()
	sort.Slice(allFacts, func(i, j int) bool { return allFacts[i].ID < allFacts[j].ID })
	proofs := make(map[Claim]ProofNode, len(allFacts))
	byID := make(map[string]ProofNode, len(allFacts))
	bySubject := map[string][]Claim{}
	byObject := map[string][]Claim{}
	queue := make([]Claim, 0, len(allFacts))
	negative := make([]facts.Fact, 0)
	selected := Claim{Predicate: query.Predicate, Args: query.Args}
	add := func(claim Claim, proof ProofNode) bool {
		if _, exists := proofs[claim]; exists {
			return false
		}
		proofs[claim] = proof
		byID[proof.ID] = proof
		bySubject[indexKey(claim.Predicate, claim.Args[0])] = append(bySubject[indexKey(claim.Predicate, claim.Args[0])], claim)
		byObject[indexKey(claim.Predicate, claim.Args[1])] = append(byObject[indexKey(claim.Predicate, claim.Args[1])], claim)
		queue = append(queue, claim)
		return true
	}
	sequence := 0
	for _, fact := range allFacts {
		if !activeAt(fact.ValidTime, instant, query.At != "") {
			continue
		}
		claim := Claim{Predicate: fact.Predicate, Args: fact.ArgReferentIDs}
		if fact.Polarity == facts.Negative {
			if claim == selected {
				negative = append(negative, fact)
			}
			continue
		}
		sequence++
		add(claim, ProofNode{ID: fmt.Sprintf("proof.%d", sequence), PremiseFactIDs: []string{fact.ID}, Sources: []facts.SourceEvidence{fact.Source}, Conclusion: claim})
	}
	work, derivations := 0, 0
	derive := func(rule Rule, conclusion Claim, premises ...ProofNode) error {
		work++
		if work > budget.MaxWork {
			return reasonLimit("work", work)
		}
		if _, exists := proofs[conclusion]; exists {
			return nil
		}
		depth := 0
		factIDs := map[string]struct{}{}
		premiseIDs := make([]string, 0, len(premises))
		for _, premise := range premises {
			if premise.Depth > depth {
				depth = premise.Depth
			}
			premiseIDs = append(premiseIDs, premise.ID)
			for _, id := range premise.PremiseFactIDs {
				factIDs[id] = struct{}{}
			}
		}
		depth++
		if depth > budget.MaxDepth {
			return reasonLimit("depth", depth)
		}
		derivations++
		if derivations > budget.MaxDerivations {
			return reasonLimit("derivations", derivations)
		}
		sequence++
		proof := ProofNode{ID: fmt.Sprintf("proof.%d", sequence), RuleID: rule.ID, PremiseProofIDs: premiseIDs, Conclusion: conclusion, Depth: depth}
		for id := range factIDs {
			proof.PremiseFactIDs = append(proof.PremiseFactIDs, id)
		}
		if len(proof.PremiseFactIDs) > MaxProofFacts {
			return reasonLimit("proof_facts", len(proof.PremiseFactIDs))
		}
		sort.Strings(proof.PremiseFactIDs)
		for _, id := range proof.PremiseFactIDs {
			fact, exists := set.ByID(id)
			if !exists {
				return reasonInvalid("MISSING_PROOF_FACT", "proof.premise_fact_ids", "proof source fact is unavailable")
			}
			proof.Sources = append(proof.Sources, fact.Source)
		}
		add(conclusion, proof)
		return nil
	}
	for head := 0; head < len(queue); head++ {
		claim := queue[head]
		current := proofs[claim]
		for _, rule := range compiled {
			if rule.Predicate != claim.Predicate {
				continue
			}
			switch rule.Kind {
			case RuleSymmetry:
				if err := derive(rule, Claim{Predicate: claim.Predicate, Args: [2]meaning.ReferentID{claim.Args[1], claim.Args[0]}}, current); err != nil {
					return Result{}, err
				}
			case RuleTransitivity:
				for _, before := range append([]Claim(nil), byObject[indexKey(claim.Predicate, claim.Args[0])]...) {
					if err := derive(rule, Claim{Predicate: claim.Predicate, Args: [2]meaning.ReferentID{before.Args[0], claim.Args[1]}}, proofs[before], current); err != nil {
						return Result{}, err
					}
				}
				for _, after := range append([]Claim(nil), bySubject[indexKey(claim.Predicate, claim.Args[1])]...) {
					if err := derive(rule, Claim{Predicate: claim.Predicate, Args: [2]meaning.ReferentID{claim.Args[0], after.Args[1]}}, current, proofs[after]); err != nil {
						return Result{}, err
					}
				}
			}
		}
	}
	result := Result{Truth: TruthUnknown, NegativeEvidence: negative}
	root, positive := proofs[selected]
	if positive {
		result.Proof = collectProof(root, byID)
	}
	switch {
	case positive && len(negative) > 0:
		result.Truth = TruthConflict
	case positive:
		result.Truth = TruthTrue
	case len(negative) > 0:
		result.Truth = TruthFalse
	}
	return result, nil
}

func validateQuery(set *facts.Set, query Query) (time.Time, error) {
	predicate, err := core.ParseRelationID(query.Predicate)
	if err != nil || predicate.String() != query.Predicate {
		return time.Time{}, reasonInvalid("INVALID_QUERY_PREDICATE", "query.predicate", "query predicate must be a canonical CORE relation ID")
	}
	relation, exists := set.CoreGraph().Relation(predicate)
	if !exists {
		return time.Time{}, reasonInvalid("UNKNOWN_QUERY_PREDICATE", "query.predicate", "query predicate is not declared in CORE")
	}
	context := set.Context()
	for index, arg := range query.Args {
		found := false
		for _, referent := range context.Referents {
			if referent.ID != arg {
				continue
			}
			found = true
			id, _ := ids.ParseConceptID(string(referent.Concept))
			concept, _ := set.CoreGraph().Lookup(id)
			allowed := relation.SubjectKind
			if index == 1 {
				allowed = relation.ObjectKind
			}
			if !allowsKind(allowed, concept.Kind) {
				return time.Time{}, reasonInvalid("QUERY_RELATION_KIND_MISMATCH", "query.args", "query referent violates the relation signature")
			}
			break
		}
		if !found {
			return time.Time{}, reasonInvalid("UNKNOWN_QUERY_REFERENT", "query.args", "query referent is not in this request context")
		}
	}
	if query.At == "" {
		return time.Time{}, nil
	}
	instant, err := time.Parse(time.RFC3339Nano, query.At)
	if err != nil || !strings.HasSuffix(query.At, "Z") {
		return time.Time{}, reasonInvalid("INVALID_QUERY_TIME", "query.at", "query time must be an RFC3339 UTC timestamp")
	}
	return instant, nil
}

func activeAt(interval facts.ValidTime, at time.Time, timed bool) bool {
	if !timed {
		return interval.From == "" && interval.To == ""
	}
	if interval.From != "" {
		start, _ := time.Parse(time.RFC3339Nano, interval.From)
		if at.Before(start) {
			return false
		}
	}
	if interval.To != "" {
		end, _ := time.Parse(time.RFC3339Nano, interval.To)
		if at.After(end) {
			return false
		}
	}
	return true
}

func indexKey(predicate string, arg meaning.ReferentID) string {
	return predicate + "\x00" + string(arg)
}

func collectProof(root ProofNode, byID map[string]ProofNode) []ProofNode {
	seen := map[string]struct{}{}
	var result []ProofNode
	var visit func(ProofNode)
	visit = func(node ProofNode) {
		if _, exists := seen[node.ID]; exists {
			return
		}
		seen[node.ID] = struct{}{}
		for _, id := range node.PremiseProofIDs {
			visit(byID[id])
		}
		result = append(result, cloneProof(node))
	}
	visit(root)
	return result
}

func cloneProof(node ProofNode) ProofNode {
	node.PremiseProofIDs = append([]string(nil), node.PremiseProofIDs...)
	node.PremiseFactIDs = append([]string(nil), node.PremiseFactIDs...)
	node.Sources = append([]facts.SourceEvidence(nil), node.Sources...)
	for index := range node.Sources {
		if node.Sources[index].Reference != nil {
			ref := *node.Sources[index].Reference
			node.Sources[index].Reference = &ref
		}
	}
	return node
}

func allowsKind(allowed []core.ConceptKind, actual core.ConceptKind) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, kind := range allowed {
		if kind == actual {
			return true
		}
	}
	return false
}

func validID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

type StatusError struct{ result status.Result }

func (failure *StatusError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "reasoning failed"
	}
	return failure.result.Diagnostics[0].Message
}
func (failure *StatusError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}
func reasonInvalid(code, path, message string) error {
	return &StatusError{result: status.Result{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}
func reasonLimit(path string, count int) error {
	return &StatusError{result: status.Result{Status: status.OutcomeResourceLimit, Diagnostics: []status.Diagnostic{{Code: "REASON_RESOURCE_LIMIT", Path: path, Message: fmt.Sprintf("reasoning %s limit exceeded at %d", path, count), Severity: status.SeverityError}}}}
}

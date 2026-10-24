package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// MeaningPath returns one stable, bounded root-to-node path for attaching
// realization assumptions to typed meaning fields.
func MeaningPath(graph meaning.Meaning, target meaning.NodeID) ([]meaning.NodeID, bool) {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	if _, exists := nodes[target]; !exists {
		return nil, false
	}
	parents := make(map[meaning.NodeID]meaning.NodeID, len(nodes))
	depths := make(map[meaning.NodeID]int, len(nodes))
	roots := append([]meaning.NodeID(nil), graph.Roots...)
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	queue := make([]meaning.NodeID, 0, len(roots))
	for _, root := range roots {
		if _, exists := nodes[root]; !exists {
			continue
		}
		if _, seen := depths[root]; seen {
			continue
		}
		depths[root] = 1
		queue = append(queue, root)
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current == target {
			path := []meaning.NodeID{current}
			for {
				parent, exists := parents[current]
				if !exists {
					break
				}
				path = append(path, parent)
				current = parent
			}
			for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
				path[left], path[right] = path[right], path[left]
			}
			return path, len(path) <= meaning.MaxMeaningDepth
		}
		if depths[current] >= meaning.MaxMeaningDepth {
			continue
		}
		children := meaningChildren(nodes[current])
		sort.Slice(children, func(i, j int) bool { return children[i] < children[j] })
		for _, child := range children {
			if _, exists := nodes[child]; !exists {
				continue
			}
			if _, seen := depths[child]; seen {
				continue
			}
			parents[child] = current
			depths[child] = depths[current] + 1
			queue = append(queue, child)
		}
	}
	return nil, false
}

func meaningChildren(node meaning.Node) []meaning.NodeID {
	switch node.Kind {
	case meaning.KindDiscourse:
		return append([]meaning.NodeID(nil), node.Content...)
	case meaning.KindEvent, meaning.KindState:
		children := make([]meaning.NodeID, len(node.Participants))
		for index, participant := range node.Participants {
			children[index] = participant.Target
		}
		return children
	case meaning.KindQuestion, meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
		if node.Target != "" {
			return []meaning.NodeID{node.Target}
		}
	}
	return nil
}

// BareDeclarativeEvidenceBuilder attests absence of separate semantic
// operators only for a complete parse whose predicate has an explicitly
// supported sense and finite-form feature signature. Other input keeps
// operator coverage unknown and cannot become an OK semantic candidate.
type BareDeclarativeEvidenceBuilder struct {
	SupportedSenses                     map[ids.SenseID]bool
	RequiredPredicateFeatures           map[features.FeatureID]features.FeatureValueID
	PredicateFeatureAlternativesBySense map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID
	// AllowedUnlexicalizedTerminals names exact function words licensed by the
	// source's compiled construction sequence, such as English "at" or Polish
	// "na". A complete parse must still consume every input token.
	AllowedUnlexicalizedTerminals map[string]bool
}

// NoOperatorEvidenceBuilder attests an empty operator inventory only for an
// already complete, supported declarative Meaning candidate. It accepts one
// statement discourse wrapper around one lexical target; question, scope,
// modality and quantifier structures remain outside this bounded path.
type NoOperatorEvidenceBuilder struct{}

func (NoOperatorEvidenceBuilder) Build(ctx context.Context, candidate semantic.MeaningCandidate) ([]construction.Operator, bool, error) {
	if ctx == nil {
		return nil, false, fmt.Errorf("operator evidence requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
		return nil, false, nil
	}
	graph := candidate.Meaning
	if len(graph.Roots) != 1 {
		return nil, false, nil
	}
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	root := nodes[graph.Roots[0]]
	if root.Kind == meaning.KindDiscourse {
		if root.Act != meaning.ActStatement || len(root.Content) != 1 {
			return nil, false, nil
		}
		content, exists := nodes[root.Content[0]]
		if !exists || content.Kind != meaning.KindEvent && content.Kind != meaning.KindState && content.Kind != meaning.KindReferent {
			return nil, false, nil
		}
	}
	for _, node := range candidate.Meaning.Nodes {
		if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState && node.Kind != meaning.KindReferent && node.Kind != meaning.KindDiscourse {
			return nil, false, nil
		}
	}
	return []construction.Operator{}, true, nil
}

func (builder BareDeclarativeEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (CompositionEvidence, error) {
	if err := ctx.Err(); err != nil {
		return CompositionEvidence{}, err
	}
	evidence := CompositionEvidence{RequestContextID: stableRequestContextID(lattice)}
	if forest.Outcome != status.OutcomeOK && forest.Outcome != status.OutcomeAmbiguous || len(forest.Roots) == 0 || len(forest.Unmatched) != 0 || len(lattice.MWEs) != 0 {
		return evidence, nil
	}
	for index, token := range lattice.Tokens {
		if token.Token.Kind == text.TokenPunctuation && token.Token.Normalized == "." && index == len(lattice.Tokens)-1 {
			continue
		}
		if token.Token.Kind != text.TokenWord {
			return evidence, nil
		}
		if len(token.Readings) == 0 {
			terminal := token.Token.Folded
			if terminal == "" {
				terminal = token.Token.Normalized
			}
			if !builder.AllowedUnlexicalizedTerminals[terminal] {
				return evidence, nil
			}
		}
	}
	byCandidate := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byCandidate[candidate.ID] = candidate
	}
	rootCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		rootCount++
		for _, alternative := range node.Alternatives {
			if len(alternative.Unresolved) != 0 || !builder.SupportedSenses[alternative.Permission.SenseID] {
				return evidence, nil
			}
			predicateFound := false
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, id := range slot.CandidateIDs {
					candidate, exists := byCandidate[id]
					if !exists || !candidate.Category.IsPredicate() || !builder.matchesPredicateFeatures(alternative.Permission.SenseID, candidate.Features) {
						return evidence, nil
					}
					predicateFound = true
				}
			}
			if !predicateFound {
				return evidence, nil
			}
		}
	}
	if rootCount == 0 {
		return evidence, nil
	}
	evidence.OperatorCoverageComplete = true
	return evidence, nil
}

// stableRequestContextID makes candidate IDs reusable across identical
// sentence requests without persisting request context between invocations.
func stableRequestContextID(lattice analyzer.LexicalLattice) string {
	var signature strings.Builder
	write := func(value string) {
		signature.WriteString(strconv.Itoa(len(value)))
		signature.WriteByte(':')
		signature.WriteString(value)
	}
	write(string(lattice.Language))
	for _, token := range lattice.Tokens {
		write(strconv.Itoa(token.Index))
		write(strconv.Itoa(token.Token.Span.Start))
		write(strconv.Itoa(token.Token.Span.End))
		write(token.Token.Surface)
		write(token.Token.Normalized)
		write(token.Token.Folded)
		write(string(token.Token.Kind))
	}
	digest := sha256.Sum256([]byte(signature.String()))
	return "request-" + hex.EncodeToString(digest[:16])
}

func (builder BareDeclarativeEvidenceBuilder) matchesPredicateFeatures(senseID ids.SenseID, actual map[features.FeatureID]features.FeatureValueID) bool {
	if alternatives, exists := builder.PredicateFeatureAlternativesBySense[senseID]; exists {
		for _, required := range alternatives {
			if predicateFeaturesMatch(actual, required) {
				return true
			}
		}
		return false
	}
	return predicateFeaturesMatch(actual, builder.RequiredPredicateFeatures)
}

func predicateFeaturesMatch(actual, required map[features.FeatureID]features.FeatureValueID) bool {
	if len(required) == 0 {
		return false
	}
	for feature, value := range required {
		if actual[feature] != value {
			return false
		}
	}
	return true
}

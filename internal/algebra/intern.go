package algebra

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
)

const (
	DefaultArtifactCacheEntries    = 16
	MaxArtifactCacheEntries        = 16
	DefaultArtifactCacheDescriptor = 16 << 20
	MaxArtifactCacheDescriptor     = 16 << 20
)

// ArtifactCache is a request-local FIFO cache. Its byte budget counts only
// serialized collision-check descriptors, not Go heap overhead.
type ArtifactCache struct {
	mu                 sync.Mutex
	maxEntries         int
	maxDescriptorBytes int
	descriptorBytes    int
	entries            map[string][]*artifactCacheEntry
	fifo               *list.List
}

type artifactCacheEntry struct {
	key                string
	semanticDescriptor []byte
	sourceDescriptor   []byte
	registryDescriptor []byte
	artifact           *Artifact
	descriptorBytes    int
	order              *list.Element
}

// NewArtifactCache creates a bounded cache. Zero values select the documented
// defaults; negative or over-capacity values are rejected.
func NewArtifactCache(maxEntries, maxDescriptorBytes int) (*ArtifactCache, error) {
	if maxEntries < 0 || maxDescriptorBytes < 0 || maxEntries > MaxArtifactCacheEntries || maxDescriptorBytes > MaxArtifactCacheDescriptor {
		return nil, invalid("INVALID_ARTIFACT_CACHE_BOUNDS", "cache", "cache bounds must be zero defaults or within the supported caps")
	}
	if maxEntries == 0 {
		maxEntries = DefaultArtifactCacheEntries
	}
	if maxDescriptorBytes == 0 {
		maxDescriptorBytes = DefaultArtifactCacheDescriptor
	}
	return &ArtifactCache{
		maxEntries: maxEntries, maxDescriptorBytes: maxDescriptorBytes,
		entries: make(map[string][]*artifactCacheEntry), fifo: list.New(),
	}, nil
}

// ExportArtifact returns an immutable checked artifact. A nil cache disables
// both whole-artifact reuse and node interning. A cache without the complete
// trusted registry structure still enables node interning, but bypasses
// whole-artifact lookup and storage.
func ExportArtifact(environment ArtifactEnvironment, source Document, cache *ArtifactCache) (*Artifact, bool, error) {
	if cache == nil {
		artifact, err := NewArtifact(environment, source, false)
		return artifact, false, err
	}
	if cache.entries == nil || cache.fifo == nil || cache.maxEntries <= 0 || cache.maxDescriptorBytes <= 0 {
		return nil, false, invalid("UNINITIALIZED_ARTIFACT_CACHE", "cache", "artifact cache must be created with NewArtifactCache")
	}
	prepared, err := prepareArtifactSource(environment, source)
	if err != nil {
		return nil, false, err
	}
	if len(prepared.selected.registryStructure) == 0 {
		artifact, buildErr := buildArtifactFromPrepared(prepared, true)
		return artifact, false, buildErr
	}
	meaningDescriptor, err := artifactMeaningBytes(prepared.envelope, prepared.document.Expression)
	if err != nil {
		return nil, false, err
	}
	semanticDescriptor := append([]byte("golingua-artifact-cache-descriptor-v1\x00"), meaningDescriptor...)
	key := artifactDigest("golingua-artifact-cache-v1\x00", meaningDescriptor)
	sourceDescriptor, err := artifactSourceResourceBytes(prepared, true)
	if err != nil {
		return nil, false, err
	}
	registryDescriptor := append([]byte(nil), prepared.selected.registryStructure...)
	descriptorSize := len(semanticDescriptor) + len(sourceDescriptor) + len(registryDescriptor)

	cache.mu.Lock()
	defer cache.mu.Unlock()
	for _, entry := range cache.entries[key] {
		if bytes.Equal(entry.semanticDescriptor, semanticDescriptor) &&
			bytes.Equal(entry.sourceDescriptor, sourceDescriptor) &&
			bytes.Equal(entry.registryDescriptor, registryDescriptor) {
			return entry.artifact, true, nil
		}
	}
	artifact, err := buildArtifactFromPrepared(prepared, true)
	if err != nil {
		return nil, false, err
	}
	if descriptorSize > cache.maxDescriptorBytes {
		return artifact, false, nil
	}
	for cache.fifo.Len() >= cache.maxEntries || cache.descriptorBytes > cache.maxDescriptorBytes-descriptorSize {
		cache.evictOldest()
	}
	entry := &artifactCacheEntry{
		key: key, semanticDescriptor: append([]byte(nil), semanticDescriptor...),
		sourceDescriptor: append([]byte(nil), sourceDescriptor...), registryDescriptor: registryDescriptor,
		artifact: artifact, descriptorBytes: descriptorSize,
	}
	entry.order = cache.fifo.PushBack(entry)
	cache.entries[key] = append(cache.entries[key], entry)
	cache.descriptorBytes += descriptorSize
	return artifact, false, nil
}

func (cache *ArtifactCache) evictOldest() {
	oldest := cache.fifo.Front()
	if oldest == nil {
		return
	}
	entry := oldest.Value.(*artifactCacheEntry)
	cache.fifo.Remove(oldest)
	cache.descriptorBytes -= entry.descriptorBytes
	entries := cache.entries[entry.key]
	for i, candidate := range entries {
		if candidate == entry {
			entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	if len(entries) == 0 {
		delete(cache.entries, entry.key)
	} else {
		cache.entries[entry.key] = entries
	}
}

// BuildArena checks a complete document, then stores its expression in a
// dense immutable arena. When useInterning is true, only exact syntax nodes
// with identical already-interned children are shared.
func BuildArena(document Document, useInterning bool) (*Arena, error) {
	checked, err := Check(document)
	if err != nil {
		return nil, err
	}
	limits, err := NormalizeLimits(document.Limits)
	if err != nil {
		return nil, err
	}
	if checked.NodeCount < 1 || checked.NodeCount > limits.MaxNodes {
		return nil, resourceLimit("expression", "checked expression occurrence count exceeds the configured node limit")
	}
	context := document
	context.Expression = nil

	nodes := make([]Node, 0, checked.NodeCount)
	buckets := make(map[[sha256.Size]byte][]NodeID)
	var visit func(*Term, int) (NodeID, error)
	visit = func(term *Term, depth int) (NodeID, error) {
		if term == nil {
			return 0, invalid("MISSING_TERM", "expression", "checked expression unexpectedly contains a missing child")
		}
		if depth > limits.MaxDepth {
			return 0, resourceLimit("expression", fmt.Sprintf("expression depth exceeds limit %d", limits.MaxDepth))
		}
		candidate := Node{
			Kind: term.Kind, Identity: term.ID, Label: term.Label,
			Type: term.Type, Symbol: term.Symbol, Quantifier: term.Quantifier,
			Modality: term.Modality, ModalProfile: term.ModalProfile,
		}
		if term.Count != nil {
			count := *term.Count
			candidate.Count = &count
		}
		if term.Kind == TermAsk {
			candidate.Answers = make([]Declaration, len(term.Answers))
			copy(candidate.Answers, term.Answers)
		}
		if term.Args != nil {
			candidate.Args = make([]NodeArgument, len(term.Args))
			for i, argument := range term.Args {
				child, childErr := visit(argument.Value, depth+1)
				if childErr != nil {
					return 0, childErr
				}
				candidate.Args[i] = NodeArgument{Role: argument.Role, Value: child}
			}
		}
		if term.Operands != nil {
			candidate.Operands = make([]NodeID, len(term.Operands))
			for i, operand := range term.Operands {
				child, childErr := visit(operand, depth+1)
				if childErr != nil {
					return 0, childErr
				}
				candidate.Operands[i] = child
			}
		}
		var childErr error
		if term.Body != nil {
			candidate.Body, childErr = visit(term.Body, depth+1)
			if childErr != nil {
				return 0, childErr
			}
		}
		if term.Restrictor != nil {
			candidate.Restrictor, childErr = visit(term.Restrictor, depth+1)
			if childErr != nil {
				return 0, childErr
			}
		}
		if term.Function != nil {
			candidate.Function, childErr = visit(term.Function, depth+1)
			if childErr != nil {
				return 0, childErr
			}
		}
		if term.Argument != nil {
			candidate.Argument, childErr = visit(term.Argument, depth+1)
			if childErr != nil {
				return 0, childErr
			}
		}

		if useInterning {
			fingerprint, hashErr := arenaNodeFingerprint(candidate)
			if hashErr != nil {
				return 0, hashErr
			}
			for _, id := range buckets[fingerprint] {
				if equalArenaNode(nodes[id-1], candidate) {
					return id, nil
				}
			}
			if len(nodes) >= limits.MaxNodes {
				return 0, resourceLimit("expression", "unique arena node count exceeds the configured node limit")
			}
			candidate.NodeID = NodeID(len(nodes) + 1)
			nodes = append(nodes, candidate)
			buckets[fingerprint] = append(buckets[fingerprint], candidate.NodeID)
			return candidate.NodeID, nil
		}
		if len(nodes) >= limits.MaxNodes {
			return 0, resourceLimit("expression", "arena node count exceeds the configured node limit")
		}
		candidate.NodeID = NodeID(len(nodes) + 1)
		nodes = append(nodes, candidate)
		return candidate.NodeID, nil
	}

	root, err := visit(document.Expression, 1)
	if err != nil {
		return nil, err
	}
	return NewArena(context, root, nodes)
}

func arenaNodeFingerprint(node Node) ([sha256.Size]byte, error) {
	node.NodeID = 0
	encoded, err := json.Marshal(node)
	if err != nil {
		return [sha256.Size]byte{}, invalid("ARENA_NODE_ENCODING", "node", err.Error())
	}
	return sha256.Sum256(encoded), nil
}

func equalArenaNode(left, right Node) bool {
	if left.Kind != right.Kind || left.Identity != right.Identity || left.Label != right.Label || left.Symbol != right.Symbol ||
		left.Body != right.Body || left.Function != right.Function || left.Argument != right.Argument ||
		left.Quantifier != right.Quantifier || left.Restrictor != right.Restrictor || left.Modality != right.Modality || left.ModalProfile != right.ModalProfile ||
		len(left.Args) != len(right.Args) || len(left.Operands) != len(right.Operands) {
		return false
	}
	if (left.Count == nil) != (right.Count == nil) || (left.Count != nil && *left.Count != *right.Count) || len(left.Answers) != len(right.Answers) {
		return false
	}
	for i := range left.Answers {
		if left.Answers[i].ID != right.Answers[i].ID || !TypeEqual(left.Answers[i].Type, right.Answers[i].Type) {
			return false
		}
	}
	if (left.Type == nil) != (right.Type == nil) {
		return false
	}
	if left.Type != nil && !TypeEqual(*left.Type, *right.Type) {
		return false
	}
	for i := range left.Args {
		if left.Args[i] != right.Args[i] {
			return false
		}
	}
	for i := range left.Operands {
		if left.Operands[i] != right.Operands[i] {
			return false
		}
	}
	return true
}

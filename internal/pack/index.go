package pack

import (
	"github.com/clickbytes/golingua/internal/core"
	conceptgraph "github.com/clickbytes/golingua/internal/graph"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
)

// Pack is an immutable in-memory view of one validated CORE/lexicon snapshot.
// Its indexes and underlying graph/index expose defensive-copy accessors.
type Pack struct {
	identity  Identity
	graph     *core.ConceptGraph
	adjacency *conceptgraph.GraphReader
	lexicon   *lexicon.Index
	symbols   *ids.SymbolTable
}

func (pack *Pack) Identity() Identity {
	if pack == nil {
		return Identity{}
	}
	return pack.identity
}

func (pack *Pack) Summary() Summary {
	if pack == nil {
		return Summary{}
	}
	lexemes, senses, bindings := pack.lexicon.Len()
	return Summary{
		Identity: pack.identity, Concepts: pack.graph.Len(), Lexemes: lexemes,
		Senses: senses, Bindings: bindings, Morphology: "not_included",
		Construction: "not_included",
	}
}

// Resolve maps a stable symbol to a dense ID scoped to this pack only.
func (pack *Pack) Resolve(symbol ids.Symbol) (ids.NodeID, bool) {
	if pack == nil {
		return 0, false
	}
	return pack.symbols.Resolve(symbol)
}

// SymbolAt returns the stable symbol represented by a pack-local ID.
func (pack *Pack) SymbolAt(id ids.NodeID) (ids.Symbol, bool) {
	if pack == nil {
		return ids.Symbol{}, false
	}
	return pack.symbols.SymbolAt(id)
}

// Concept returns an immutable graph value through its defensive-copy API.
func (pack *Pack) Concept(id ids.ConceptID) (core.Concept, bool) {
	if pack == nil {
		return core.Concept{}, false
	}
	return pack.graph.Lookup(id)
}

// Lexicon returns the immutable lexical index; its accessors return copies.
func (pack *Pack) Lexicon() *lexicon.Index {
	if pack == nil {
		return nil
	}
	return pack.lexicon
}

// Graph returns the immutable bounded adjacency index for this pack's CORE.
func (pack *Pack) Graph() *conceptgraph.GraphReader {
	if pack == nil {
		return nil
	}
	return pack.adjacency
}

// Concepts returns defensive copies in canonical ID order.
func (pack *Pack) Concepts() []core.Concept {
	if pack == nil {
		return nil
	}
	return pack.graph.Concepts()
}

package ids

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

const MaxSymbols = 1 << 20

// NodeID is a dense position valid only within one frozen symbol-table snapshot.
type NodeID uint32

var (
	ErrSymbolCollision = errors.New("canonical symbol collision")
	ErrSymbolLimit     = errors.New("symbol table limit exceeded")
	ErrBuilderFrozen   = errors.New("symbol table builder is frozen")
	ErrNodeIDOverflow  = errors.New("symbol count exceeds NodeID capacity")
)

// SymbolTableBuilder collects validated stable symbols. Freeze sorts their
// canonical forms before assigning dense IDs, making IDs independent of Add
// order for the same symbol set.
type SymbolTableBuilder struct {
	symbols map[string]Symbol
	frozen  bool
}

func NewSymbolTableBuilder() *SymbolTableBuilder {
	return &SymbolTableBuilder{symbols: make(map[string]Symbol)}
}

// Add validates a symbol and rejects any second symbol with the same canonical
// form, including case variants that would otherwise alias after normalization.
func (builder *SymbolTableBuilder) Add(symbol Symbol) error {
	if builder == nil {
		return errors.New("nil symbol table builder")
	}
	if builder.frozen {
		return ErrBuilderFrozen
	}
	if builder.symbols == nil {
		builder.symbols = make(map[string]Symbol)
	}
	canonical, err := symbol.Canonical()
	if err != nil {
		return err
	}
	if _, exists := builder.symbols[canonical]; exists {
		return fmt.Errorf("%w: %s", ErrSymbolCollision, canonical)
	}
	if len(builder.symbols) >= MaxSymbols {
		return fmt.Errorf("%w: maximum is %d", ErrSymbolLimit, MaxSymbols)
	}
	normalized, err := NewSymbol(symbol.Namespace, symbol.Name)
	if err != nil {
		return err
	}
	builder.symbols[canonical] = normalized
	return nil
}

// Freeze returns an immutable table with dense IDs assigned in canonical
// lexical order. The builder cannot be changed after this call succeeds.
func (builder *SymbolTableBuilder) Freeze() (*SymbolTable, error) {
	if builder == nil {
		return nil, errors.New("nil symbol table builder")
	}
	if builder.frozen {
		return nil, ErrBuilderFrozen
	}
	if uint64(len(builder.symbols)) > uint64(math.MaxUint32)+1 {
		return nil, ErrNodeIDOverflow
	}
	canonical := make([]string, 0, len(builder.symbols))
	for symbol := range builder.symbols {
		canonical = append(canonical, symbol)
	}
	sort.Strings(canonical)

	table := &SymbolTable{
		symbols: make([]Symbol, len(canonical)),
		ids:     make(map[string]NodeID, len(canonical)),
	}
	for index, key := range canonical {
		if uint64(index) > uint64(math.MaxUint32) {
			return nil, ErrNodeIDOverflow
		}
		table.symbols[index] = builder.symbols[key]
		table.ids[key] = NodeID(index)
	}
	builder.frozen = true
	return table, nil
}

// SymbolTable is an immutable snapshot-local stable-symbol index. Its storage
// is private and its accessors return values rather than mutable slices.
type SymbolTable struct {
	symbols []Symbol
	ids     map[string]NodeID
}

func (table *SymbolTable) Len() int {
	if table == nil {
		return 0
	}
	return len(table.symbols)
}

// Resolve maps a stable symbol to its dense snapshot-local NodeID.
func (table *SymbolTable) Resolve(symbol Symbol) (NodeID, bool) {
	if table == nil {
		return 0, false
	}
	canonical, err := symbol.Canonical()
	if err != nil {
		return 0, false
	}
	id, exists := table.ids[canonical]
	return id, exists
}

// SymbolAt returns a copy of the stable symbol for a snapshot-local NodeID.
func (table *SymbolTable) SymbolAt(id NodeID) (Symbol, bool) {
	if table == nil || uint64(id) >= uint64(len(table.symbols)) {
		return Symbol{}, false
	}
	return table.symbols[id], true
}

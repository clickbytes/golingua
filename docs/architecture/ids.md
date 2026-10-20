# Stable identifiers and snapshot-local IDs

Stable symbols name source definitions across versions. Dense `NodeID` values are only compact positions inside one frozen runtime snapshot; never persist them or use them as public identity. A new snapshot may assign different numeric IDs when its symbol set changes.

## External symbol form

The canonical external form is `namespace:name`, for example `concept:animal.dog`. `Namespace` and `Name` are separate fields in Go. A name may contain dot-separated nonempty segments. Each segment uses ASCII letters, digits, `_` or `-`; canonical output uses lowercase ASCII. Whitespace, controls, separators inside a segment, empty segments and names exceeding the documented byte/segment bounds are invalid.

Identifiers are intentionally ASCII-only. Concept, lexeme, sense, construction and procedure IDs are machine identifiers, not display labels or words in Polish/English. Rejecting non-ASCII here prevents canonically composed/decomposed and confusable Unicode spellings from creating aliases without imposing an implicit normalization policy. Human-readable Unicode definitions belong in separate data fields.

The JavaScript-sensitive segments `__proto__`, `prototype` and `constructor` are rejected in namespaces and names so identifiers remain safe across JSON/map tooling as well as Go. No namespace or concept is inferred from a name prefix; callers must supply the namespace explicitly.

## Distinct stable ID types

`ConceptID`, `LexemeID`, `SenseID`, `ConstructionID` and `ProcedureID` are separate Go types whose constructors enforce their namespace. They cannot be substituted for each other by assignment. `Symbol` is the generic source-table key when a table intentionally contains more than one symbol kind.

## Dense runtime table

`SymbolTableBuilder` validates and collects symbols, rejects duplicate canonical forms, then sorts canonical strings during `Freeze`. The same set therefore receives the same dense order regardless of insertion order. `Resolve` performs a canonical-key lookup; `SymbolAt` indexes a frozen slice. The table exposes no mutable maps or slices. The symbol count is bounded by `MaxSymbols`; `NodeID` is a `uint32` and is checked for representational overflow before assignment.

Use stable symbols in source definitions, diagnostics and cross-version references. Use `NodeID` only while traversing the exact snapshot that produced it. Do not derive symbols from lexical content or from prefixes, and do not treat a `NodeID` as a `ConceptID`.

# Neutral CORE data

`data/core/1000-concept-proposals.tsv` is the clean, owner-supplied editorial record of the 1000-concept inventory. It preserves the original row numbers, statuses, source IDs, license labels, definitions, and all ten columns. It is **not** read at startup and does not generate declarations automatically. The executable source of truth is the statically typed Go data in `internal/catalog/concepts.go` and its domain files.

The TSV distinguishes 49 existing concepts from 951 proposed concepts. Those statuses describe editorial history only; they are not binding approvals. The 49 original definitions remain unchanged. New adopted definitions are split by domain in Go and are validated by `catalog.Compile()`.

For new declarations, use a stable `ConceptID`, a distinct definition, a same-kind parent when appropriate, semantic roles only when justified, and a `provenance.SourceRef`. The current proposal source reference is `golingua-core-proposed-2026-10-21`, path `data/core/1000-concept-proposals.tsv`, version `2026-10-21-draft`, license `unknown`, with no asserted author. Cite provenance through `SourceRef` fields (`ID`, `Path` or `URI`, `Version`, `License`, and `Author` when confirmed). Provenance records origin; it does not confer a right to reuse external data.

The literal license label `unknown` is retained. An unknown license does not equal permission to import external dictionaries or datasets. No external dictionary has been imported for this inventory.

CORE stores language-neutral concepts, not PL/EN lexical senses, inflection, valency, constructions, or direct language pairs. English definitions are metadata, not EN lexical entries. A concept of kind `relation` is not a `RelationDefinition` schema for graph edges. Never create approved PL↔EN bindings from TSV status, a matching name, or a definition alone.

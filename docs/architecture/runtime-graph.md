# Runtime graph indexes and bounded traversal

`core.CompileConceptGraph` validates source definitions first. `graph.Compile` then assigns snapshot-local dense `NodeID` values in sorted stable-symbol order and builds immutable outgoing, incoming, relation and parent indexes. Query-time lookups use typed IDs and integer adjacency; stable-symbol parsing and JSON are not part of traversal. A rebuilt snapshot may use different numeric IDs.

The index uses bounded adjacency lists behind the `GraphReader` API. The API does not expose a storage layout, so a later implementation may choose CSR or another representation without changing traversal callers. The current builder caps snapshots at 100,000 concepts and 262,144 combined definition/taxonomy edges, checking counts before copying graph nodes or allocating adjacency. It uses in-memory slices and maps only; there is no graph database or SQL connection on the hot path.

## Edge semantics

The typed `Concept.Parent` link is the hard taxonomy edge used by `Ancestors` and `IsA`; it is indexed as a required parent edge. A named relation such as `is_a` is not inferred from its string spelling. Definition relations remain distinct typed edges with their explicit `required` or `typical` strength. Outgoing, incoming and relation indexes return defensive copies in deterministic `RelationID`/`NodeID` order.

`Targets` returns direct relation targets by default. It follows further links only when that relation schema declares `Transitive=true`. Typical edges are skipped unless the caller explicitly opts in; their strength is preserved in results. Cycles terminate through a visited set.

## Budgets and outcomes

`Budget` requires positive `MaxNodes`, `MaxEdges` and `MaxDepth`, with hard ceilings of 100,000 nodes, 262,144 edges and 100,000 depth. The default per-operation budget is 10,000 nodes, 20,000 edges and depth 128. `ProjectConcept`, `Targets` and `Ancestors` return no partial result when work exceeds a limit; the error carries `status.OutcomeResourceLimit` and a `RESOURCE_LIMIT` diagnostic. Invalid budgets return `invalid`; a missing root or relation schema returns `unknown`.

Projection output sorts nodes and edges before returning, so equivalent source sets produce byte-order-stable values regardless of definition insertion order. These graph operations do not infer facts beyond declared taxonomy and explicitly transitive relation schemas.

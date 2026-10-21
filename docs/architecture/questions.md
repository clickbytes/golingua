# Typed question execution

`internal/reason.Plan` accepts a typed `MeaningGraph`, request context and an explicit `RelationMapping` supplied by a trusted caller. The graph must have one question discourse act and one question scope over an event/state predicate. The caller maps two declared CORE semantic roles to the subject/object positions of one binary CORE relation. Text labels and lexical similarity never choose the relation.

Every grounded participant must resolve through a request `Mention` to one `ReferentID`. A WH question leaves exactly one mapped role absent; its answer candidates are request referents whose CORE kinds fit the missing relation argument. Yes/no questions have two grounded arguments. Unsupported operator scopes, unresolved identity, missing role mappings or unsupported shapes produce typed outcomes instead of a guess.

`Execute` delegates yes/no truth to the bounded reasoner. `true` and `false` require positive proof or direct negative evidence. Opposing evidence is `conflict`; no evidence is `unknown`. WH execution evaluates each bounded, type-compatible referent and returns all positive or conflicting candidates with proof evidence. Multiple candidates are `ambiguous`; it never selects the first. Answers contain stable referent IDs, not surface words. Content planning and language realization are later stages.

This API is library-only. No owner supplied question input or PL/EN answer CLI exists yet, so manual answer evaluation remains `NOT RUN`.

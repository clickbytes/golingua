# Request scoped references and equivalence

`ReferentID` identifies one entity instance only within one `RequestContext`. A `ConceptRef` identifies a class; it does not make two referents the same instance. `MentionID` identifies a referring occurrence. A mention can link to a referent only with an explicit bounded `EvidenceID`; absent links remain unresolved. Deictic anchors retain bounded context supplied by the caller and do not resolve identity by themselves.

Reference context is request scoped and must not be persisted as session memory. Context IDs form a boundary: semantic comparison across different contexts returns `unknown`, even when their referent labels happen to match. A missing identity binding returns `unknown`; conflicting bindings return `ambiguous`.

## Three different comparisons

- `StructuralEquivalence` compares canonical Meaning structure, including local node IDs and ordered discourse content.
- `SemanticEquivalence` ignores local node IDs and node input order while preserving roles, discourse order, operators, quantifier values and explicit referent identity. It does not infer identity from a concept, lemma or graph shape.
- `SurfaceEquivalence` compares exact, caller supplied SHA-256 fingerprints from `SurfaceFingerprint`. It performs no spelling, whitespace or language normalization and stores no surface text.

These comparisons answer different questions. Structural equality does not prove entity identity; semantic equality does not prove world truth; surface equality does not prove semantic equality. Each operation validates finite graph and context bounds before returning a comparison.

The model has no cross-request memory, automatic anaphora resolver or source provenance verifier. `EvidenceID` is an explicit caller supplied reference to evidence; this layer validates its shape but does not certify the evidence itself.

Semantic equivalence also compares [temporal frame and referent multiplicity](meaning-attributes.md). Resolving two mentions to the same request identity does not erase different attributes. Identification contexts must preserve those attributes in the corresponding predicate/referent, even when their local node IDs differ.

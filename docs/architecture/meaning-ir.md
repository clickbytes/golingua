# Meaning IR v1

`internal/meaning` represents one request-scoped interpretation. `ConceptGraph` describes stable language-neutral kinds and role signatures; a `Meaning` graph links concrete event, state and referent nodes and scopes operators over them. A concept definition is not an event in the world, and a Meaning node is not a stored fact.

## Shape and identity

The JSON envelope has `schema_version`, `roots` and tagged `nodes`. Node IDs are bounded, local identifiers for one graph. Concept fields use stable `concept:*` IDs. Events and states hold explicit role-to-node participants. Questions, scopes and modalities point to one target; discourse nodes preserve ordered content. The v1 node kinds are `event`, `state`, `referent`, `question`, `scope`, `modality` and `discourse`.

```json
{
  "schema_version": 1,
  "roots": ["question_1"],
  "nodes": [
    {"id":"event_1","kind":"event","concept":"concept:read","participants":[{"role":"agent","target":"person_1"},{"role":"theme","target":"book_1"}]},
    {"id":"person_1","kind":"referent","concept":"concept:person"},
    {"id":"book_1","kind":"referent","concept":"concept:book"},
    {"id":"question_1","kind":"question","target":"event_1"}
  ]
}
```

The sample illustrates shape only; it does not add lexical data or claim an approved semantic interpretation. Swapping `agent` and `theme` changes the meaning. Canonicalization sorts nodes and role assignments while keeping each role attached to its target; discourse content order remains intact.

## Validation and boundaries

`Build` and `ValidateStructure` check the closed tagged union, bounded IDs, unique node/root IDs, resolved references, acyclic structure and node/edge/depth budgets. `ValidateMeaning` additionally resolves concept IDs and checks event/state/referent concept kinds, declared CORE roles and allowed target kinds. A role signature describes which roles may occur; this version does not infer that every declared role must be present.

Strict JSON decoding rejects unknown fields, extra JSON values, unsupported schema versions and oversized input. Invalid data returns `invalid`; unresolved node/concept references return `unknown`; exhausted work returns `resource_limit`. Errors carry the shared typed status plus bounded diagnostics. Canonical JSON, equality and SHA-256 hash use no pointer addresses.

There are no grammatical case, gender, inflection, tense morphology or surface-order fields. Semantic time and richer quantification/operator inventories can be added in later scoped issues. A `question` node is an information-seeking wrapper, and `scope` nodes keep negation or hypothetical operators attached to their target. The IR has no truth/assertion field, so questions and negative scopes do not become positive asserted facts. Discourse act labels do not certify that content is true.

The graph is a bounded semantic representation, not a parser, translator, persistent world model or evidence store. No language-specific data belongs in CORE.

## CLI inspection

`golingua meaning canonicalize --json` reads one Meaning document from stdin, applies the strict structural validator and prints canonical JSON. It does not load ontology data; callers that need cross-CORE role checks use `meaning.ValidateMeaning` with a compiled `core.ConceptGraph`.

```sh
golingua meaning canonicalize --json < meaning.json
```

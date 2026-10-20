# ConceptGraph definitions

`Concept` identifies a language-neutral category or meaning type. It is not a word, lexical sense, utterance participant, or fact about a specific referent. `ConceptID`, `LexemeID` and `SenseID` have separate typed namespaces; a lexical sense may bind to a concept only through a separate explicit mapping.

## Definitions and relation schemas

The source shape is versioned in [`concepts.schema.json`](../../schemas/core/concepts.schema.json). A document contains a bounded list of concept definitions and relation schemas. Unknown object fields are rejected by the schema. The Go compiler also validates the typed API and limits, so callers that bypass JSON Schema still receive invalid diagnostics.

Concept kinds (`entity`, `event`, `property`, `relation`, `abstract`) describe semantic categories only. Parent links provide taxonomy, must resolve without cycles, and must use the same kind unless the parent is a shared `abstract` root. Role signatures and relation schemas list allowed endpoint kinds. An empty allowed-kind list means any valid kind. Relation IDs and concept IDs are explicit; names and prefixes never infer ontology.

Definition edges retain their declared `required` or `typical` strength. This graph stores source descriptions and does not perform reasoning. A `typical` edge is never promoted to a required relation or a certain fact. `SemanticEffect` is a bounded language-neutral declaration; it carries no POS, lemma, case, tense, inflection, or surface order.

## Compiled snapshot

`CompileConceptGraph` validates all IDs, kinds, endpoint signatures, dangling references and parent cycles before returning a graph. Invalid input returns bounded diagnostics with stable codes and paths. Successful graphs use deterministic key order and private storage. `Lookup` returns a defensive value copy; modifying the returned slices cannot mutate the compiled snapshot.

Concepts describe types, not instances. Distinct entities that happen to share one concept belong in a future `MeaningGraph` as separate referents. This package does not create referents, sentence facts, lexical mappings, or language-specific grammar.

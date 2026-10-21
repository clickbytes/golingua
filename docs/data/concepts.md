# CORE ConceptDefinition catalog

`internal/catalog` is the executable, language-neutral source of truth for CORE. The TSV at `data/core/1000-concept-proposals.tsv` is an editorial record of the supplied inventory; it is not parsed at runtime. The catalog compiles 1000 `ConceptDefinition` records toward the editorial target of 1000: 49 original definitions and 951 definitions accepted from the supplied proposals. The 49 original records and their definitions remain unchanged. The interim 187-record growth batch from PR #144 was replaced by this owner-supplied inventory; shared IDs are represented once in the resulting catalog.

New records are static, typed Go declarations split into domain files under `internal/catalog/concepts_*.go`. `conceptGroups()` registers the base and domain groups deterministically. `Entries()` validates each group's `SourceRef`, creates stable IDs through `ids.NewConceptID`, and returns fresh values with copied role signatures. `Compile()` delegates to `core.CompileConceptGraph`; `Count()` reports the compiled graph size.

## Adding a concept

1. Choose a stable `concept:<name>` ID for one language-neutral meaning. The TSV `NO` is only an editorial row number and must never become a runtime ID.
2. Choose the matching `ConceptKind` and, when justified, a parent of the same kind. Parents must exist and the taxonomy must remain acyclic.
3. Write a short definition that distinguishes the meaning from nearby concepts. The English text in `Definition_EN` is metadata, not an English lexeme or lexical sense.
4. Add only semantic participant `RoleType`s supported by the definition. A TSV dash means no roles (`nil`). Role names are not grammatical cases, prepositions, subject/object labels, or language-specific valency frames.
5. Preserve provenance in `SourceRef` (`ID`, `Path`, `Version`, `License`, and confirmed `Author`). `unknown` is a truthful license value, not a permission to import external data.
6. Add the typed declaration to the relevant domain function in deterministic row order. Update the actual compiled counts in the data documentation.

`DRAFT_REVIEW` in the TSV is editorial history only. It is not a `SemanticBinding` status and does not make the compiler a linguistic-quality checker. Every adopted definition requires human semantic review; structural compilation only checks IDs, kinds, parents, roles, cycles, and related invariants.

`core.KindRelation` describes a kind of concept, such as `concept:located_in`. It is distinct from `core.RelationDefinition` (`relation:<name>`), which defines a schema for typed `DefinitionEdge`s. A `KindRelation` does not automatically create a `RelationDefinition` or an asserted graph edge.

CORE concepts do not contain PL/EN lexemes or approved `SemanticBinding`s. Language-local lexical senses need their own evidence and provenance. Do not infer an approved binding from an ID, parent, or English definition; never create direct PL↔EN pairs here. In particular, the physical carrier `concept:book` does not stand for the work's information content.

The 49 original meanings, including the distinction between `concept:looking` (directed visual attention) and `concept:seeing` (visual perception), remain unchanged. The inventory keeps `concept:giving` and `concept:receiving` as distinct transfer viewpoints; `concept:buying` and `concept:selling` as separate transaction perspectives; `concept:school` as a facility and `concept:school_institution` as an organization; and calendar intervals as meanings rather than grammatical tense or number. These distinctions do not imply any language binding.

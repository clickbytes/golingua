# #476 — source feature evidence bound to Meaning occurrences

Base: `36490cbc30d63d1ce99ffacc7ba7a1c32682ec0c` (product `b063258`).
Date: 2026-10-31. Work is serial, on `main` only, by owner instruction.

## Reproduced defect

`SourceFeatures` used `(role, CORE concept)` as an evidence key. Independently
authored Meaning with two seeing events, each with a different cat as stimulus,
accepted an unaddressed `determiner=definite` record as known source information.
This is not an occurrence identity. The before-change regression actually fails;
this observation is not a claim that a particular full multi-clause translation
was demonstrated incorrect.

## Contract and implementation

Source `EvidenceRecord` now optionally carries `meaning_node_id` and
`predicate_node_id`. These are local IDs in its candidate's primary Meaning,
not CORE IDs, parser IDs, permanent referents or proof of world truth.

- Predicate-local lexical evidence addresses its exact node; owner is empty.
- Argument evidence addresses its referent and owning predicate. The declared
  role must connect exactly those nodes, with the exact declared concept.
- A referent reused in multiple events can have separate referring occurrences
  and corresponding source grammatical evidence. Owner and role are therefore
  part of the key, not just the referent node.
- Mapped native source predicates, arguments, implicit-argument licensing records
  and modifier predicates carry their proper address. A licensing record for an
  omitted participant addresses the predicate, not an invented surface mention.
- Generated target binding evidence cannot claim source addresses. Its source
  trace and the separate request context retain their existing responsibilities.

`SourceFeaturesAt` takes the owning predicate (or the predicate itself for an
empty role) and verifies the concept/argument link. All existing product callers
of the old helper were migrated. Lexical aspect/manner and NP determiner readers
follow the same rule. Legacy unaddressed records are still usable only when the
concept/role has one possible occurrence; repeated occurrences are never merged
or selected by record order. An independent Meaning without source evidence
continues to use the language's existing explicit alternatives/assumptions.

At the target boundary, stale links, ambiguous unaddressed features, and conflicting
known values are rejected before an adapter can mistake them for missing evidence.
Record and feature limits remain bounded. Returned feature maps are independent.
Trace validation checks field shape and text bounds; graph validation checks the
exact correspondence. The new optional fields are included in trace JSON.

A semantic equivalence boolean does not provide an ID mapping. Disambiguation
therefore merges addressed traces only in the same canonical graph namespace.
Equivalent candidates in different namespaces remain separate until an explicit
mapping exists. Reordering nodes is harmless; changing IDs requires remapping the
trace and context. No name-based or concept-based rebinding is performed.

## Verification

Focused tests use independently authored repeated-event/referent graphs, distinct
and shared referents, present/past evidence, wrong owners, wrong concepts, stale
IDs, absent IDs, conflicting values, legacy ambiguity, unique legacy compatibility,
generated-origin misuse, budgets, mutation isolation, actual PL/EN analysis and
same-language realization. Explicit graph/context/trace renaming preserves output;
graph-only renaming refuses; removing source evidence from `The cat sees a cat.`
retains the four independent English article combinations.

Controls include `Czytałam list.`, nominal repeated cats, and plural past negative
seeing. Technical checks include formatting, vet, build, focused race checks and
the full existing matrix. Exact command outputs, CLI observations, source hashes,
post-integration Actions and the final SHA are recorded in #476 and delivery
artifacts. The matrix is observational and still contains outstanding gaps.

## Boundaries

This is not a Meaning schema migration, nor a replacement of semantic properties
with parser history. Source grammatical evidence remains separate from Meaning;
fully independent representation of all distinctions remains a separate task.
There is no new coordination parser, source-aligned coordinated target acceptance,
new language, lexeme, assumed fact, expected-text expansion, GAP removal or budget
increase. General synonym-specific planning and duplicated language allowlists
remain open. No new full DSE acceptance or independent second-agent review is
claimed. Owner language acceptance remains NOT EVALUATED.

# Construction templates

The files under `languages/{pl,en}/constructions/` are language-local,
structural templates. They describe NP, VP, and PP slots and permitted slot
orders. They do not contain lexemes, sentence examples, semantic roles, or
sense-specific valency frames. A family is shared independently of any one
lexeme or sense.

`internal/construction.DecodeConstructionPackage` reads schema version 1 JSON,
rejects unknown and duplicate object fields, checks its declared language
against the family, and compiles the variants into a deterministic
`VariantIndex`. Compilation proves only that the structure is well formed.
These files are not loaded by the application runtime today.

In particular, the Polish OVS order is a marked structural option; it does
not assert case marking, topic/focus, or interchangeable grammatical roles.
The English auxiliary-question and Polish `czy` templates represent limited
orders, not complete question grammars. A polarity operator and meaning
evidence must be supplied explicitly by a future planner.

## Current catalog coverage

The catalog now exposes `catalog.ResolveApprovedPlans`: it follows a CORE
concept through an **approved equivalent** lexical binding, a sense-specific
valency frame, an **approved** construction binding and the exact compiled
family/variant plan. It returns `unsupported` at the first missing license.
Multiple matching paths remain separate and ordered; the resolver does not
choose a meaning or produce surface text.

The PL and EN packages have independently declared transitive families and
variants in `languages/*/constructions/transitive.json`. PL also declares an
intransitive family. The four current PL and five current EN valency frames
name allowed families, but `CompileValency` supplies no approved construction
bindings. All sense-to-CORE links in the in-code lexical batches are pending.
Therefore the current catalog has **zero authorized construction paths**, and
the CLI `translate` remains unsupported.

Structural templates can be shared by compatible senses only after their
roles, grammatical functions, case or adposition policy, variant and operator
constraints are reviewed for each language. In particular `patrzeć na` and
`look at` have oblique target arguments; a generic direct object template
must not silently stand in for those frames. A frame's `AllowedFamilyIDs` is
only a candidate family reference. It is never treated as approval.

The owner's sentence-level review has not run. Future review should add
explicit construction bindings and compiled plans before advertising any
sentence analysis or realization capability in the language manifests.
No external lexical or linguistic source is claimed by the current structural
templates; their provenance records identify them as internal project data.

## Polish structural family correction — G07-03

`languages/pl/valency.go` now assigns a compatible construction family per sense instead of marking all frames `pl.transitive`. `pl.intransitive` applies to `spać` and the currently modeled `uczyć się`; `pl.oblique` (new `languages/pl/constructions/oblique.json`) describes a PP complement for `patrzeć na`; `pl.ditransitive` (new `ditransitive.json`) can describe a dative recipient and accusative theme for `dać`. Remaining listed transitive frames have a single NP complement. Semantic participant names match their referenced CORE concepts.

These declarations are **not** approved sense-to-family bindings. Optional `o co` (`pytać`), `na co` (`odpowiadać`), and discipline/content complements (`uczyć`/`uczyć się`) remain out of scope until semantic roles and language-local arguments are reviewed. No sentence capability was activated.

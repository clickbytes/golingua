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
The CLI compiles the construction packages for its explicitly enabled route;
other families remain structural data until a provider and stage adapter use
their exact licensed plans.

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
intransitive family. `CompileValency` exposes pending construction candidates
separately from approved bindings. Following the owner's delegated review, the
index contains four approved PL routes (`widzieć`, `słyszeć`, `dostać`,
`patrzeć na`) and seven approved EN sense/frame/family/variant routes across
`see`, `hear`, three `get` variants, `read`, and `look at`. G05 #201 binds
`look at` to the English oblique family with an exact `at` terminal; the Polish
route carries `na` and an accusative complement. G08 #66 (PR #213) added the
mass-object `get` variant; G10-08 #228 (PR #230) added an indefinite-object
variant for count singular `receiving` objects where source definiteness is
missing.

`catalog --summary --json` reports candidate and approved construction-binding
counts separately. Current counts are **zero pending candidates, four approved
PL routes and seven approved EN routes**. These bindings license only
their exact sense/frame/family/variant routes. The catalog can resolve exact
approved plans; the compiled CLI registry exposes `seeing`, `hearing`,
`looking`, and `receiving` sentence routes. The approved `read` route does not
yet have a sentence provider and stage adapter. `Bindings()` returns only approved records;
`Candidates()` is review input and never licenses a runtime path.

Structural templates can be shared by compatible senses only after their
roles, grammatical functions, case or adposition policy, variant and operator
constraints are reviewed for each language. In particular `patrzeć na` and
`look at` have oblique target arguments; a generic direct object template
must not silently stand in for those frames. A frame's `AllowedFamilyIDs` is
only a candidate family reference. It is never treated as approval.

The owner's sentence-level review has not run. The compiled CLI enables the
`seeing`, `hearing`, `looking`, and `receiving` provider routes. The base
language manifests continue to report their general analysis/realization
capabilities as unsupported; the runtime package is generated for these exact
routes and verified against its provider evidence.
No external lexical or linguistic source is claimed by the current structural
templates; their provenance records identify them as internal project data.

## Polish structural family correction — G07-03

`languages/pl/valency.go` now assigns a compatible construction family per sense instead of marking all frames `pl.transitive`. `pl.intransitive` applies to `spać` and the currently modeled `uczyć się`; `pl.oblique` (new `languages/pl/constructions/oblique.json`) describes a PP complement for `patrzeć na`; `pl.ditransitive` (new `ditransitive.json`) can describe a dative recipient and accusative theme for `dać`. Remaining listed transitive frames have a single NP complement. Semantic participant names match their referenced CORE concepts.

These declarations are **not** approved sense-to-family bindings. Optional `o co` (`pytać`), `na co` (`odpowiadać`), and discipline/content complements (`uczyć`/`uczyć się`) remain out of scope until semantic roles and language-local arguments are reviewed. No sentence capability was activated for these families.

## Polish movement and transport sense frames

The G07-03 motion batch adds 24 reviewed `LexicalSense` candidates with 24 authored `ValencyFrame` declarations. Each frame uses exactly its CORE role signature, while Polish case and adpositions stay local. One-role motions use `pl.intransitive`, direct-object frames use `pl.transitive`, and PP complements use `pl.oblique`. The families are only structural; `Approved` binding count remains zero, and the motion verbs have no asserted morphology paradigms or sentence capabilities.

## New PL families for reviewed multi-argument verbs

`languages/pl/constructions/transitive_oblique.json` and `double_np.json` provide structural slots, respectively (subject NP + verb VP + object NP + oblique PP) and (subject NP + verb VP + object NP + complement NP). `ditransitive.json` now has consistent `recipient` and `object` slot identifiers, matching the corresponding PL valency frames. Sense-specific cases such as `z`+genitive, `do`+genitive or dative recipient are stored in `valency_actions_multirole.go`, **not** in CORE.

Other new PL valency records remain without approved construction bindings. Template validity alone does not imply parsing, realization or PL↔EN transfer.

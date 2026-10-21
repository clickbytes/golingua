# Construction coverage of the first PL/EN catalog batch

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

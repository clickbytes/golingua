# Semantic alternatives and disambiguation

G05-04 maps parser alternatives to bounded Meaning candidates only when an
approved lexical sense-to-CORE binding and the exact sense/frame/construction
permission are available. It preserves evidence and uncertainty; it does not
pick a meaning because it has a higher score or appears first.

## Evidence required for a Meaning candidate

Each candidate is built from one parser alternative, one approved
`MappingEquivalent` `SemanticBinding`, its exact `ValencyFrame`, and the
approved `ConstructionBinding` named by that frame and construction plan.
Every event, state, and referent concept must come from an explicit approved
binding. Every participant role must come from the construction's semantic
role binding, and every target must have explicit lexical or reference
evidence. Token order, grammatical function, or a common-sense guess cannot
swap roles or supply a missing participant.

An evidence trace links source byte spans, phrase and reading IDs, sense and
CORE IDs, plan/family/variant/frame IDs, binding provenance, and the reason a
candidate survived or was rejected. Empty IDs mean the evidence is unavailable;
they never authorize an inferred mapping. Evidence remains request-scoped and
does not become part of the language-neutral Meaning graph.

## Validation and partial candidates

`meaning.ValidateMeaning` checks graph structure, concept existence and kind,
declared roles, and allowed role-target kinds. It does not require every role
declared by a frame to be present. Validate each candidate frame separately
with `valency.ValidateFrame`; that check enforces frame permissions,
Sense-to-CORE approval, required roles, and selectional constraints.

When a required role is missing, retain the observed participants and the
partial Meaning candidate with `incomplete` and its `MISSING_REQUIRED_ROLE`
diagnostic. Do not invent a referent or placeholder participant to make the
frame pass. If a valid Meaning graph cannot be constructed because a concept,
target, or binding is unknown, retain an evidence-only candidate without
claiming Meaning validation passed. Explicitly malformed graphs are `invalid`;
known hard contradictions may be rejected with their typed `unsupported`
diagnostic.

## Outcomes

| Outcome | Meaning for disambiguation |
| --- | --- |
| `ok` | The evidence and search are complete, and one semantically distinct, fully validated Meaning remains. |
| `unknown` | Available evidence cannot establish a required identity or mapping. Missing evidence is not a contradiction. |
| `unsupported` | A requested transfer or construction has no approved license, or explicit evidence conflicts with a hard policy. No counterpart is invented. |
| `ambiguous` | More than one semantically distinct, licensed Meaning remains after explicit hard constraints. Preserve every alternative. |
| `incomplete` | Some candidate, role, or search path remains unresolved, including a missing required role or incomplete parser input. Do not report a unique complete result. |
| `resource_limit` | A configured bound stopped work. Retain only the bounded partial alternatives and state that coverage is incomplete. |

These outcomes are distinct. A missing or unknown feature cannot prove a hard
match or a hard contradiction. A soft constraint may rank or annotate a
candidate, but cannot remove it. A score never turns ambiguity, unsupported
evidence, or incomplete coverage into `ok`.

## Identity, deduplication, and uniqueness

Semantic identity follows the Meaning graph: concept identity, role-to-target
assignments, and explicit operator, modality, quantifier, question, and
discourse scope. Surface order, token spelling, local Meaning node IDs, and
lexical insertion order do not define semantic identity. Use semantic graph
equivalence only when its reference evidence resolves the compared referents
in the same request context. If equivalence is unknown or ambiguous, keep the
candidates separate.

Equivalent Meaning graphs may share one semantic alternative only if all
supporting senses, frames, constructions, bindings, and source evidence remain
attached to it. Different role assignments or operator scopes remain distinct
alternatives even when they use the same words. Never deduplicate by concept
alone or by a digest that omits graph structure.

A unique `ok` result requires complete input and search coverage, exactly one
semantic equivalence class of fully validated Meaning candidates, and explicit
hard evidence that rules out every competitor. An overlapping unsupported,
unmapped, or unresolved reading prevents a confident unique `ok` result unless
explicit hard evidence eliminates it. If an alternative limit is reached,
the result remains `incomplete` even when only one candidate was retained.

## Operators and referents

The current composition API can represent explicit negation and modality when
the construction operator, value, target candidate, and source span are
provided and licensed. Question subtypes, hypothetical scope, tense, voice,
quantifier, and discourse operators are not all represented by this composition
path; retain them as unsupported evidence until the Meaning/operator schema
can preserve their exact value and scope. Surface punctuation or word order
alone does not create an operator. Referent identity requires explicit
request-scoped reference evidence; matching concepts or repeated names do not
establish that two mentions denote the same individual. Unresolved or
conflicting reference evidence remains `unknown` or `ambiguous` rather than
being guessed.

## Review boundary

The owner supplies Polish and English text for manual review. Until then,
language quality remains **NOT RUN**. Do not add generated sentence suites,
gold outputs, test fixtures, score thresholds, or comparisons with the former
JavaScript implementation.

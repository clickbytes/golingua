# Semantic operator scopes and asserted projection

Meaning nodes carry explicit scope references. `question`, `negation`, `hypothetical`, `modality` and quantifier wrappers point to their operand; no operation moves an operator across another scope. `meaning.OperatorForNode` exposes a typed view without changing the stored source graph.

## Asserted and unresolved content

`ProjectAsserted` accepts an event or state as asserted content only when it is reachable through an explicit `Discourse` node with `act: statement`. The output records the semantic node and the exact source path. It is evidence that the input meaning represents a statement, not proof that the proposition is true in the world.

Question, negation, hypothetical, modal and quantifier scopes go into `unresolved` with a typed reason and source path. Their operands are not projected as positive assertions. A request/question discourse act is also unresolved. An event/state without an explicit statement act is unresolved rather than assumed to be asserted. Unknown or missing context is not treated as false.

Quantifiers are bounded semantic values (`some`, `all`, `exactly`, `at_least`, `at_most`) scoped over one operand. Numeric quantifiers require an explicit bounded count; missing or null counts are not interpreted as zero. They do not map directly to Polish or English grammatical number. No event time is inferred from lexical words or from a tense-shaped field in Meaning.

## Supported combinations and limits

Scope nesting remains explicit. The current projection does not commute, simplify or infer through nested operators. Double negation has no licensed construction in this version and returns typed `unsupported` with `UNLICENSED_DOUBLE_NEGATION`; no partial projection is returned. Other scoped operators remain unresolved until a later issue licenses a specific interpretation.

Structural validation checks dangling nodes, acyclic references, and bounded node/edge/depth work before projection. Budget exhaustion is `resource_limit`; malformed scope data is `invalid`; missing references are `unknown`; unsupported composition remains distinct from each of those outcomes.

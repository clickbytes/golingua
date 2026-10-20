# Bounded construction parser

The parser consumes source-spanned lexical evidence and explicit phrase
candidates. It matches those candidates against one exact compiled construction
plan at a time. It does not infer POS categories from language-neutral code or
build a `MeaningGraph` from surface order.

## Inputs and evidence

Each parser plan joins a `CompiledConstructionPlan` with its exact family,
variant, sense-specific valency frame and approved construction binding. The
family supplies slot categories, cardinality, optionality, case and adposition
policies that are not copied into the compiled plan. Invalid or mismatched
joins are rejected during index compilation.

A phrase candidate has a stable candidate ID, NP/VP/PP category, half-open token
range and source byte span, plus lexical reading IDs, senses and head features.
Its category and feature evidence come from a language-specific phrase builder
or explicit caller. The parser checks that referenced token and reading
evidence exists in the lexical lattice. A lexical POS string alone does not
imply NP, VP or PP.

Construction slot bindings assign semantic roles. OVS surface order cannot
assign roles. When a plan requires a case, only an explicit matching case value
licenses that path. An unknown or conflicting feature cannot prove a hard
constraint; unresolved evidence stays represented and marks coverage
incomplete. Soft constraints are advisory and do not remove candidates.

Optional slots may be omitted only when licensed by the family and variant. A
WH gap additionally needs explicit evidence naming its slot, semantic role and
question scope. `ParseContext` is trusted caller supplied evidence; the parser
checks bounded identifier shape but does not validate scope against a
`MeaningGraph`. It never infers an extracted argument from word order.

## Packed forest and limits

Equal subspans with the same phrase category share one forest node. Each node
retains every candidate and derivation edge that reaches it. Construction
alternatives preserve distinct sense/frame permissions; multiple derivations
with one permission remain separate edges and do not manufacture semantic
ambiguity. The parser does not select by score, insertion order or first match.
Each root alternative's slot binding references shared phrase node IDs, keeping
explicit links from construction interpretations to phrase evidence.
The compiled index includes language-local buckets keyed by possible first
terminal or phrase category, including starts exposed by optional leading
slots. Request-time parsing consults those buckets at the sentence start.

Node, alternative and expansion budgets are independent and finite. Exhaustion
returns the retained partial forest with `incomplete` and a resource-limit
diagnostic. Ambiguity is represented by alternatives and never overrides
incomplete coverage. Unsupported or unresolved lexical evidence remains
visible in the forest result.

## Current boundary

The parser returns syntax, slot assignments, lexical evidence references and
licensed sense/frame permissions. A later semantic stage may construct a
meaning only after checking approved sense-to-CORE mappings and other required
licenses. This parser does not infer a translation, a fact or a unique reading.

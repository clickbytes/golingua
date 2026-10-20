# Target lexicalization

Target lexicalization sits between semantic composition and surface
realization. The current bounded API selects target lexical and construction
alternatives for validated Meaning nodes and packages one selected alternative
per `BuildPlan` call. It does not generate inflected forms or sentence text.

## Candidate construction and licensing

`SelectTargets` starts from an explicit target language, lexical index,
valency index, construction registry, family-local compiled variant indexes,
and CORE graph. It considers concept-bearing Meaning nodes, semantic roles,
and explicitly supplied operators. Language profiles and grammar remain local
to the target language. This API has no target feature-profile or case-
governance index input; constraints not represented by its supplied licenses
remain unresolved.

Apply licensing filters before ordering. A target lexical sense can transfer a
CORE concept only through an `approved` `equivalent` semantic binding. Pending,
`broad`, `narrow`, `related`, and `candidate` bindings may be retained as
unlicensed evidence, but they do not authorize a target word. Each candidate
must satisfy the exact target sense, valency frame, approved construction
binding, registered family, and family-local compiled variant. The variant
must permit realization in the requested direction and validate against the
selected frame. The exact binding must also appear in both the valency index
and construction registry. A family or plan name by itself grants no
permission to a sense.

Construction role bindings determine which target slots express Meaning
participants. Word order, lexical category, frequency, and surface position do
not assign semantic roles. Hard constraints are enforced when represented by
the supplied valency and construction licenses. Missing target profile data,
case governance, or unresolved feature evidence remains visible and cannot be
treated as a match.

## Stable ordering after licensing

The current API does not score or prefer licensed candidates because no
ranking policy or preference data is supplied. It sorts retained alternatives
by stable source, Meaning node, sense, frame, family, and variant identifiers;
this ordering is for reproducibility and does not imply semantic preference.
It cannot create a binding, permit an unlicensed construction, repair a known
contradiction, or remove a legal alternative. A future ranking policy must use
explicit bounded target-language data or caller policy. No implicit universal
grammar or hidden corpus frequency is assumed.

## Alternatives and evidence

Keep distinct licensed lexical senses, lexical entries, and constructions as
separate alternatives until their target forms are known to be equivalent for
the requested Meaning. Do not merge candidates merely because they share a
lemma, a CORE concept, or a spelling. If alternatives are later recognized as
equivalent, retain the supporting sense, frame, construction, variant, and
provenance evidence for each derivation.

Each retained candidate remains traceable to the source candidate and Meaning
node it serves, the target sense and approved binding. Predicate candidates
also retain the exact valency and construction permissions considered;
referent candidates carry lexical transfer evidence without predicate
licenses. A missing mapping or construction is a visible failure, not a
guessed fallback. The caller retains the validated source Meaning and supplies
it to `BuildPlan` with one selected candidate, so the graph is not duplicated
in each output alternative.

## Assumptions, loss, and unsupported cases

The Meaning graph is the source of semantic roles and operators. Lexicalization
does not infer participants, referent identity, tense, polarity, information
structure, or omitted content from token order. Any target distinction required
by a construction must come from explicit Meaning content or explicit
request-scoped evidence. If that evidence is absent, keep the affected
candidate unresolved.

Some source distinctions may have no supported target expression, and some
target constructions may express distinctions not represented in Meaning.
Record such cases as explicit field-level loss or unresolved evidence. Do not
silently drop a Meaning role, operator, or other semantic distinction to
obtain a fluent-looking string. If no licensed target lexical/construction
path remains, report the operation as unsupported. Missing evidence remains
unknown; if it leaves overall search coverage unresolved, report the result as
incomplete. Exhausted work bounds produce a resource-limited result.

This stage chooses lexical and construction plans, not inflected surface forms.
Target feature-profile checks, preference ranking, morphology, and linearized
surface generation remain future work where their required inputs are absent.
Later stages may supply forms only through an explicitly supported target
paradigm, procedure, or exception. They may not invent a missing lexical form
or change the Meaning graph to make realization succeed.

## Bounded work and current status

Candidate enumeration and license inspection have finite bounds. Scanning the
lexical index to build lookup maps is bounded by lexicon package ceilings;
Meaning graph traversal, operator checks, candidate retention, and license
enumeration are additionally bounded by caller-visible expansion limits. The
selector does not build an unbounded Cartesian product of senses, mappings,
and construction plans. On exhaustion, retained alternatives remain available
and coverage is marked resource-limited; the selector never reports a unique
complete result after truncation.

The current runtime performs bounded target lexical/construction candidate
selection and per-candidate plan packaging, but not sentence realization.
This document does not imply that any particular target-language data set is
complete or approved.

# DSE and Meaning v1 compatibility (RFC)

**Status:** normative compatibility contract for the proposed DSE route; implementation `NOT RUN`; owner acceptance `NOT_EVALUATED`.

This document defines how future DSE work coexists with the existing request-scoped Meaning v1. It does not add an API, route, schema version, migration, or fallback in the current product. Until an explicitly selected DSE integration is implemented and reviewed, the existing Meaning v1 paths remain unchanged.

## 1. Authority and opt-in

Each request uses one explicitly selected semantic route:

| Route selection | Authority | Behavior |
| --- | --- | --- |
| No DSE opt-in / existing route | Meaning v1 | Existing parser, candidate, context, evaluation, and realization behavior remains authoritative. |
| Explicit DSE opt-in on a future supported route | DSE expression | One DSE term is authoritative for that selected route; v1 may be used only through a reviewed, semantics-preserving conversion. |

An opt-in must name and pin the DSE registry identity, version, and content fingerprint for the request. DSE metadata lives in a separately versioned registry selected by request/configuration; it is not silently merged into, or used to mutate, the current CORE/lexicon `Pack`. The Pack is an immutable snapshot identified by its own dataset version and digest. A DSE registry fingerprint is not a Pack digest and must not be presented as one.

If the explicitly selected DSE route returns `unsupported`, `unknown`, `incomplete`, `invalid`, or `resource_limit`, the caller receives that result and its diagnostics. The engine MUST NOT silently retry the same request through Meaning v1 and return the fallback as though DSE had succeeded. A caller may explicitly submit a separate v1 request, but that is a separate interpretation with a separate result. A request may not carry competing v1 and DSE authorities for one semantic alternative.

All nonselected routes remain v1-authoritative. Selection and future integration belong to a separately reviewed implementation leaf (planned #397); this RFC does not claim that configuration or CLI controls exist.

## 2. Existing v1 structures and identity boundaries

Meaning v1 is a closed, versioned, request-scoped graph ([current types](../../../internal/meaning/types.go)). A `Meaning` has graph-local `NodeID`s, roots, and tagged nodes such as event, state, referent, scope, modality, quantifier, question, and discourse. Events/states contain a `ConceptRef` and role-to-node participants. Referents carry a `ConceptRef` and optional `Multiplicity`; event/state nodes carry an optional `TemporalFrame`. Scope and discourse are explicit. A graph is not a fact store and has no implicit truth or existence field.

Important existing identity distinctions:

| Existing item | Scope / meaning | DSE compatibility rule |
| --- | --- | --- |
| `ConceptRef` | stable CORE concept identifier | may be copied only through an exact declared mapping; it is not an instance ID |
| `NodeID` | unique only within one Meaning graph | do not reuse as a persistent DSE binder or request referent ID |
| `BinderID` | lexical scope in a DSE term | never convert to a `NodeID` or entity identity |
| `ParameterID` | free typed input in a DSE term | not an entity, anaphor, or implicit argument in v1 |
| `HoleID` | unresolved typed DSE expression slot | not a v1 node, witness, or silently omitted field |
| `ReferentID` | identity inside one `RequestContext` | preserve only in the same explicitly paired context; never globalize it |
| `MentionID` | one referring occurrence inside one request context | preserve occurrence and its node link; do not resolve it by concept or surface label |
| `EvidenceID` | caller-supplied identifier attached to an explicit resolution | its shape is validated; it is not authenticated proof of identity |
| `Role` | semantic participant role declared for a CORE concept | DSE `PortID` is distinct; map only by an explicit reviewed role/port binding |

`MeaningCandidate` keeps a candidate ID, semantic ID, optional Meaning graph, status, evidence trace, and `MeaningContext` pairs. Each context contains the exact graph whose local node IDs it addresses and a request-local `RequestContext`. Converters MUST keep each context paired with its graph and validate that pairing. They may not carry context entries over to a structurally different graph or resolve an unresolved mention while converting.

The v1 context validator checks local IDs, mention links, and shape. In particular, an empty `Mention.ReferentID` is unresolved; a nonempty one requires an existing request-local referent and an evidence ID. This validation does not authenticate the evidence source. The request context is not durable memory. See [request-scoped references](../references.md) and [`internal/meaning/reference.go`](../../../internal/meaning/reference.go).

## 3. Partial conversion contract

Conversion is allowed only when the target representation can preserve the relevant structure and all request-local links. A conversion must either:

1. produce a validated result and a mapping record for every converted node, binder, port, reference, and context link; or
2. refuse with a typed diagnostic identifying the unsupported or lossy part.

There is no “best effort” success that drops semantic content. A type check alone does not prove that two representations express the same meaning. Conversion also does not prove that the represented proposition is true.

| DSE structure / information | Possible v1 mapping | Required behavior if exact mapping is unavailable |
| --- | --- | --- |
| Typed atomic proposition | Only an explicit mapping to a supported v1 event/state/role structure | refuse or retain as DSE-only; do not make a v1 event from a symbol label |
| `Lambda` / bound variable | Only after scope-safe reduction to a supported representation | refuse a residual function; never turn a binder into an entity instance |
| Free `ParameterID` | No implicit v1 equivalent | refuse unless an explicit request contract supplies and maps the parameter |
| `Exists` / hidden witness | Only an exact supported quantifier/scope mapping that preserves its bound and scope | refuse if v1 cannot express the quantifier or if an existential is confused with known existence |
| `HoleID` | No complete v1 equivalent | preserve as DSE incomplete, or refuse conversion; never drop it |
| Ordered relation ports | Explicit port-to-role mapping with compatible sorts and arity | refuse unmapped, duplicated, reordered, or ambiguous roles |
| Distinct entity terms/references | Distinct v1 graph nodes and/or context referents, as required by source identity | refuse rather than merge by concept, label, or graph shape |
| `TemporalFrame` | Exact `past`/`present` value on the mapped event/state | preserve empty as unresolved; never default empty to present or past |
| `Multiplicity` | Exact `one` or `many` on the mapped referent | preserve empty as unresolved; never infer one/many from missing data or grammatical guesswork |
| Scope, polarity, quantifier, modality | Exact supported v1 scope/operator node and target relationship | refuse or retain DSE-only when scope/value cannot be represented; never project the positive operand as an assertion |
| Dialogue act / assertion force | Exact explicit discourse representation | absence remains absence; no proposition is made an assertion by default |
| Request context and mentions | Same validated graph/context pair, preserving context ID, IDs, anchors, and evidence links | refuse if pairing, scope, identity, or mention linkage would change |
| Evidence trace | Separate request-scoped trace with its own typed records and source spans | never merge evidence into Meaning or treat identifiers/provenance as identity proof |

The table is not permission to map every formula that resembles a v1 graph. The converter must verify the exact structure, types, scope, operator values, and context. For example, a two-edge DSE formula with an existential intermediate is not automatically one v1 `seeing` event; it can be mapped only if an explicit mapping defines and preserves that structure. No conversion may add a parent, individual, event, or fact that the source did not represent.

## 4. Attributes, ambiguity, and assumptions

Missing semantic fields remain unresolved. The v1 zero value for `TemporalFrame` means unknown; missing `Multiplicity` does not mean `one`. Mass nouns and plural-only morphology do not determine individual cardinality. A converter must preserve these distinctions and must not manufacture a value to satisfy a grammar or target construction. See [temporal frame and multiplicity](../meaning-attributes.md).

Multiple licensed DSE candidates remain separate alternatives. Conversion must not merge them because they share a label, target text, `ConceptRef`, or apparent graph shape. Where v1 equivalence is unknown across distinct request contexts, identity remains unknown; where compatible readings remain unresolved, report ambiguity rather than selecting the first candidate. Existing [`MeaningCandidate`](../../../internal/semantic/disambiguate.go) and [`OutcomeKind`](../../../internal/status/status.go) retain statuses and derivation evidence for this reason.

Grammar assumptions are not semantic content. A language target may declare a field-level assumption for a grammatical distinction absent from the source, if the route explicitly supports that policy. It may not use an assumption to fill a required semantic participant, resolve an identity, invent a time or multiplicity, or remove a DSE/v1 field. Semantic content loss is a separate loss record with a path and reason; if a required distinction cannot be preserved, conversion or realization refuses rather than reporting an assumption as a substitute.

No proposition is asserted merely because it is represented, typed, lexicalized, or realized. Questions, requests, hypothetical content, negation, and modality retain explicit scope and discourse force. A positive projection from a question, negation, or hypothesis is not a compatible conversion. See [operator scopes](../semantic-operators.md) and [bounded reasoning](../reasoning.md).

## 5. Result and evidence boundaries

Use current `status.OutcomeKind` values only with their existing meanings: `ok`, `ambiguous`, `incomplete`, `unknown`, `unsupported`, `invalid`, `resource_limit`, and `error`. A DSE typing result, goal result, consistency result, and fact-evaluation result are separate claims, not interchangeable status labels. In particular:

- `unsupported` means the selected profile cannot handle the construct; it is not a negative answer;
- `unknown` means the available evidence establishes neither side in the active profile;
- `incomplete` means unresolved content or proof obligations remain;
- `resource_limit` means bounded work stopped before completion, not that no solution exists;
- `invalid` is reserved for malformed or structurally contradictory input.

Existing `EvidenceTrace` is request-scoped and may include source spans, senses, bindings, roles, features, construction plans, outcomes, reasons, and provenance. Its shape and IDs do not authenticate a source document, prove identity, or establish world truth. DSE certificates are a separate proof object governed by the DSE verifier profile. Neither may silently substitute for the other. See [`internal/semantic/evidence.go`](../../../internal/semantic/evidence.go) and [lexicalization](../lexicalization.md).

## 6. Runtime and data boundaries

The existing v1 Pack is an immutable CORE/lexicon snapshot pinned by dataset ID, version, and digest ([Pack decisions](../03-data-decisions.md)). DSE uses its own pinned metadata registry and fingerprint. Changing one cannot silently mutate or reinterpret the other. No schema migration, automatic import, Pack extension, or public API change is authorized by this compatibility RFC.

Future configuration may explicitly opt a route into DSE and supply the DSE registry fingerprint. Any API field or CLI option required for that route must be proposed and reviewed in its implementation leaf; it must not overload the existing dataset ID/digest or infer opt-in from an installed registry. A missing or mismatched DSE registry fails closed. Existing v1 clients and routes keep their behavior when DSE is not selected.

The DSE evaluator consumes explicit facts under a named profile. The v1 Meaning graph is not a fact set, the CORE ontology is not an implicit fact database, and lexical bindings are not evidence that a proposition is true. No fact is synthesized from a translated sentence or from a conversion. Any future bridge from DSE derivations to v1 evidence must preserve premises, rule IDs, registry/profile versions, and derivation IDs explicitly.

## 7. Verification and delivery status

This issue is documentation-only. It does not authorize Go changes, migrations, API additions, runtime commands, new tests, test runners, or workflow changes. Later work follows the current `AGENTS.md` and its own active leaf; this RFC does not create standing test authorization.

Review status, mathematical correctness assessment, technical implementation, language evaluation, and owner acceptance are separate. This compatibility design has not been executed in a runtime: implementation/runtime `NOT RUN`, owner acceptance `NOT_EVALUATED`. Future implementation evidence must name the actual base SHA, registry fingerprint, selected routes, exact commands, outputs, and unresolved limitations. Documentation alone is not a runtime pass.

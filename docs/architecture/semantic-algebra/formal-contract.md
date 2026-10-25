# DSE formal contract (RFC)

**Status:** normative design for DSE-01; implementation `NOT RUN`; owner language acceptance `NOT_EVALUATED`.

This document defines the typed, domain-independent expression and bounded reasoning contract used by the Dynamic Semantic Expressions (DSE) work. It does not claim that these types, operators, a synthesis command, or a proof checker exist in the current binary. Later issues may implement this contract in stages. Until then, existing Meaning v1 and PL/EN adapters remain authoritative for the routes they already support.

The key separation is:

1. metadata describes symbols, signatures, sorts, and explicitly trusted axioms;
2. a typed term represents a proposition or function without deciding whether it is true;
3. a GoalSpec constrains a search for such terms;
4. a verifier checks typing, goal constraints, and any certificate claims separately;
5. an evaluator may derive a result from explicit facts in its declared profile;
6. a language adapter may realize a represented meaning where a licensed construction exists.

No layer may infer semantics from a label, identifier spelling, or language-specific code path.

## 1. Normative terms and dependency direction

“MUST”, “MUST NOT”, and “MAY” describe this contract. A symbol's printable label is explanatory data; its stable ID and declared signature identify it. Labels such as `parent`, `mother`, or `grandparent` MUST NOT dispatch to dedicated Go methods, callbacks, or algorithms.

| Layer | Owns | May depend on | MUST NOT depend on |
| --- | --- | --- | --- |
| Typed IR | sort/type IDs, binders, terms, structural validation | neutral IDs and values | domain registry, parser, PL/EN plugins |
| Symbol registry | sort declarations, symbol signatures, port metadata, provenance | typed IR | language adapters or executable callbacks |
| Construction / synthesis | GoalSpec, candidate generation, resource bounds | typed IR, registry, generic operators | label-based behavior or a domain-specific constructor |
| Verification | typing, goal obligations, certificate checking | typed IR, registry, explicit verifier profile | synthesis search decisions as proof of correctness |
| Facts / evaluation | explicit facts, match environments, evidence | typed IR, declared facts and allowlisted rules | implicit global facts or language guesses |
| Meaning / utterance | proposition roots, participants, context, speech act | typed IR and explicit mappings | treating a proposition as an asserted fact by default |
| Language realization | construction plans, morphology, lexical bindings | meaning and language-local grammar | changing the meaning to make a surface form fit |

Dependencies point from higher layers to lower, neutral contracts. The typed IR is reusable when a domain registry or language adapter changes. Existing `internal/core`, `internal/meaning`, and `internal/lexicon` structures are inputs to the compatibility design; this RFC does not silently redefine or migrate them.

Each DSE request pins an explicitly selected, separately versioned metadata registry. The DSE registry is not an implicit extension of the existing CORE/lexicon `Pack` and cannot replace it by load order. Generic relation ports use neutral `PortID`s and sorts; domain roles remain data and never choose operator behavior. Existing Meaning v1 remains authoritative for every route not explicitly selected for DSE. A selected DSE route has one authoritative expression representation; a request must not silently combine conflicting v1 and DSE interpretations. Detailed route selection and conversion rules belong to the separate compatibility contract.

## 2. Sorts, signatures, and types

The core sort vocabulary is:

| Type | Meaning |
| --- | --- |
| `Entity` | an entity-valued term; it does not assert identity, existence, or a domain class |
| `Event` | an event-valued term |
| `Interval` | a temporal interval-valued term |
| `Prop` | a proposition-valued term; it is not a Boolean or a truth value |
| `Nominal(s)` | a nominal sort identified by a declared `SortID s` in metadata |
| `A → B` | a function type from `A` to `B`; arrows associate to the right |

Domain-specific nominal sorts are declared as data. Any subtype relation must also be declared and validated; until a subtype rule is in the active profile, type checking uses exact sort equality. Neither English descriptions nor ID prefixes imply a sort or subtype. A relation signature contains an ordered list of `PortID`s, each with a sort, and has result sort `Prop`. Port order is semantic: a `RoleID` or human-readable port name does not reorder arguments.

For example, a caller may declare an otherwise uninterpreted binary relation symbol `r` with ports `left: Entity`, `right: Entity` and type `Entity → Entity → Prop`. This declaration gives `r` a type only. It says nothing about its extension, transitivity, truth, or intended English translation.

## 3. Identities, scope, and holes

The IR uses distinct, stable ID namespaces:

| ID | Refers to | Binding rule |
| --- | --- | --- |
| `BinderID` | a variable introduced by `Lambda` or `Exists` | `Var(id)` is valid only in that binder's lexical scope |
| `ParameterID` | an intentionally free, typed input parameter | remains free until explicitly bound; never aliases a referent |
| `ReferenceID` | a referent supplied by a request or context | points to an explicit external reference record; never acts as a binder |
| `HoleID` | an unresolved expression slot with an expected type | is not a variable, witness, fact, or negative result |
| `PortID` | an ordered argument position in a relation signature | connected only by an explicit port mapping |

IDs, not display names, establish identity and scope. Alpha-equivalent binders may be rendered with different display names while retaining equivalent structure. Implementations MUST validate unique binder declarations, reject unbound `Var` references, and preserve free parameters, reference IDs, and holes as separate categories. A hole or free parameter prevents a claim of complete evaluation unless the active goal explicitly treats it as an input.

Parameter substitution and beta reduction are different operations. `BindParam(t,p,a)` replaces only free `Param(p)` occurrences with a type-compatible `a`; it preserves every `Var(BinderID)` occurrence and does not remove a lambda. Beta reduction for `Apply(Lambda(b:A,body),arg:A)` substitutes `arg` for occurrences of exactly `Var(b)` in `body`, then removes that outer lambda. It never substitutes a `Param` merely because its ID or display name resembles `b`. Both operations are type-preserving and capture-avoiding. Before substituting a composed/open term, implementations alpha-rename nested binders whose IDs would capture free binder references in the argument; imported term copies with repeated binder IDs are freshened before composition. Parameter IDs, binder IDs, reference IDs, and hole IDs remain separate namespaces throughout. Display-name collisions never cause capture. If a name-based serializer cannot preserve these distinctions, it must alpha-rename or refuse.

## 4. Term forms

The initial IR grammar is deliberately small:

```text
Var(BinderID)                 bound variable
Param(ParameterID, Type)       free typed input
Ref(ReferenceID, Type)         externally supplied referent
Hole(HoleID, Type)             unresolved typed slot
Atom(SymbolID, ordered terms)  primitive proposition
And(Prop...)                   finite conjunction
Exists(BinderID: Type, Prop)   existential proposition
Lambda(BinderID: Type, Term)   typed function
Apply(Function, Argument)      function application
```

The following generic operations are derived over those forms. They have no domain-specific implementation branches. Port-aware operations use a sidecar **function interface**, not a new IR term constructor: an ordered mapping `PortID → BinderID` for distinct leading `Lambda` arguments of a typed relation function. The mapping is one-to-one and records the relation signature's port order and types. It does not change the function's type. A bare registered relation symbol may be eta-expanded from its declared signature for these operators—for `r:A→B→Prop`, the expansion is `λa:A.λb:B.Atom(r,Var(a),Var(b))`. A complete `Atom` is already a `Prop`; it can be used as a relation function only by such a well-typed eta expansion over its declared free inputs. There is no implicit port matching from labels, argument order guesses, or role names.

| Form / operation | Type and premises | Result | Failure | Execution boundary |
| --- | --- | --- | --- | --- |
| `Var(b)` | `b` is declared in the current scope with type `A` | term of type `A` | `unbound_binder` or scope error | structural check only |
| `Param(p,A)` | `p` is a declared free input | term of type `A` | `undeclared_parameter` or conflicting type | structural check only |
| `Ref(i,A)` | request context declares reference `i` with type compatible with `A` | term of type `A` | `unknown_reference` or type mismatch | context lookup only; does not prove facts about the referent |
| `Hole(h,A)` | hole ID is unique and expected type is well formed | incomplete term of type `A` | duplicate hole or malformed type | cannot be evaluated as a witness |
| `Atom(s,args)` | `s` has a registered signature; full arity, order, and argument types match | `Prop` | `unknown_symbol`, `arity_mismatch`, `type_mismatch` | records a proposition; no truth is implied |
| `And(p1…pn)` | every operand has type `Prop`; `n ≥ 1` | `Prop` | `type_mismatch` or empty conjunction | positive profile can evaluate finite conjunctions |
| `Exists(b:A,p)` | `p:Prop`; binder `b:A` is fresh in its scope | `Prop` | ill-scoped binder or non-`Prop` body | positive finite profile searches explicit witness facts within bounds |
| `Lambda(b:A,t)` | body `t:B` is well scoped | `A → B` | ill-scoped binder or invalid body | represents a function; not itself an assertion |
| `Apply(f,a)` | `f:A→B` and `a:A` | `B`; if `f=Lambda(b,body)`, beta-substitute exact `Var(b)` occurrences and remove the outer lambda | `not_a_function`, argument type mismatch, or invalid scope | beta-reduction is structural and bounded; may yield a residual function |
| `BindParam(t,p,arg)` | `Param(p,A)` occurs free in `t`; `arg:A` | `t` with only free `Param(p)` occurrences replaced | missing/duplicate parameter binding or type mismatch | capture-avoiding free-parameter substitution; does not add a fact |
| `Bind(t,interface,port,arg)` | `interface` maps `port` to a leading lambda binder `b:A`; `arg:A` | term with that exact lambda binder applied to `arg` and removed; remaining interface is retained | absent/duplicate port, non-leading binder, or type mismatch | capture-avoiding beta binding; does not add a fact |
| `Join(P,Q,interfaces,map)` | interfaces map distinct leading lambda binders; `map` explicitly pairs one output port of `P` with one input port of `Q` of equal type | alpha-rename component scopes disjointly; replace the paired binder occurrences by one fresh existential witness; conjoin bodies; existentially bind the witness; wrap remaining binders in declared output order | missing/ambiguous mapping, type mismatch, binder collision, or output-port collision | finite positive join only; bounded by expansion and witness limits |
| `Restrict(P,interface,port,C)` | `interface` maps `port` to a leading binder `b:A`; `C:A→Prop` | same ordered lambda interface with body `Pbody ∧ Apply(C,Var(b))` | unknown port, non-leading binder, or incompatible predicate type | positive restriction only; does not discard the port |
| `Project(P,interface,ports)` | selected ports are distinct members of the interface | selected leading binders remain in their declared order; every unselected interface binder becomes an explicit existential in the body | unknown/duplicate port, missing interface, or invalid scope | finite positive projection; witness enumeration is bounded; free params/refs/holes are never discarded |

An `And` node is logical conjunction. It does not mean database update, default closure, or assertion in the world. `Join` MUST record both function interfaces and the selected port mapping in the term or certificate; equality of labels such as `parent` or `child` is not sufficient to connect ports. Component lambda scopes are made disjoint before construction. The shared witness type must equal both paired port types. Remaining input and output binders are wrapped in the explicitly declared output interface order. `Restrict` keeps the original interface. `Project` existentially binds only unselected interface arguments; it is not deletion of an argument that might still affect meaning and cannot drop free parameters, references, or holes.

For a relation `r: A → B → Prop`, the general two-step join connects `r`'s second port in the first copy to `r`'s first port in the second copy, with a fresh witness `z:B` only when the declared sorts permit it. For the homogeneous case `r: Entity → Entity → Prop`, the result is:

```text
λx:Entity. λy:Entity. ∃z:Entity.
  (r(x,z) ∧ r(z,y))
```

This is an anonymous expression generated from a goal and signature. It is not a stored `Grandparent` function, a method on `Parent`, or a claim that `r` is transitive. The expression says only that the two `r` propositions share an existential intermediate entity.

## 5. Goal specification and synthesis

A `GoalSpec` describes a search problem, not its answer. At minimum it contains:

- a desired type and explicitly ordered open ports;
- typed structural constraints or obligations over those ports and registered symbols;
- any permitted generic operators and declared proof profile;
- finite resource bounds (expansions, depth, candidates, and evaluation witnesses).

A goal MUST NOT require a prewritten final AST or a named derived concept. A search MAY return several candidates. Every candidate is checked independently against the goal. A candidate's metadata label, synthesis path, or type alone does not establish that it satisfies the goal. A name or alias for a successful expression is optional metadata and is not used to select an algorithm.

For example, a goal can supply the symbol `r`, input ports `x:Entity` and `y:Entity`, require a two-edge positive path from `x` to `y`, and permit `Join`, `Exists`, and `And`. The synthesizer can construct the expression in §4 without knowing what `r` means. A finite synthesis constraint `allowed_atom_counts: [1, 2]` with the same required symbol, endpoints, and connected-path obligation admits both a direct `r(x,y)` candidate and a two-edge candidate. The verifier checks each separately against that one structural obligation. This finite search configuration is not an `Or` constructor in the proposition language: neither returned term contains disjunction. Both are distinct valid candidates; the engine MUST retain both unless an explicit constraint ranks or excludes one. They are not presumed equivalent.

The initial executable synthesis profile is **positive finite conjunctive construction**: `Atom`, `And`, typed binders, `Exists`, application, and the generic operations above. Its domain-axiom and domain-rule allowlist is empty: only structural alpha/beta reduction and positive matching against supplied facts execute. Metadata alone cannot authorize a domain axiom or rule. Any future profile that permits one requires a separately reviewed implementation and explicit allowlisting with provenance. Search may beta-reduce lambdas into this fragment. It does not include unrestricted higher-order unification, arbitrary recursion, theorem proving, negation-as-failure, disjunction, universal quantification, or unconstrained search. Unsupported constructs return `unsupported`; exhausted bounds return `resource_limit`, never “no solution”.

Registered atomic relation symbols return `Prop` in this executable profile. Function and lambda terms may be represented and type-checked, but finite query evaluation accepts them only after bounded reduction to the supported positive fragment. A residual function, hole, or unsupported term is reported as incomplete or unsupported rather than treated as an executable fact.

## 6. Four separate verification claims

The system MUST report these claims independently:

1. **Well typed:** each term and operator obeys the declared signatures and scope rules.
2. **Goal satisfied:** a candidate meets the explicit structural and constraint obligations in the selected synthesis profile.
3. **Constraints consistent:** only decidable consistency checks in the named finite profile have been completed. This is not a claim about arbitrary first-order or higher-order axioms.
4. **Conclusion follows from facts:** an evaluator produced a derivation from supplied facts and any rules explicitly authorized in its profile.

None implies the next. For example, `r(Entity,Entity) → Prop` is well typed whether or not a relation fact exists; it does not choose a desired meaning. A goal can be type-correct yet have no candidate under its constraints. A candidate can satisfy the syntactic goal without being true of any supplied facts. Metadata may declare domain axioms only with provenance and an explicit allowlisted rule profile; unlisted or unsupported axiom kinds are not executed. Labels and descriptions are never axioms.

### Independent checks and certificates

Synthesis output includes the candidate term, its goal obligations, and a certificate that names the signature versions, operator steps, substitutions, and resource bounds used. A separate verifier checks the term and certificate against the original GoalSpec and registry. It MUST NOT treat “the synthesizer returned it” as proof. If the verifier cannot replay a step, the claim is incomplete or unsupported; it is not silently accepted.

Certificates prove only what their checker profile states. A typing certificate proves typing, not truth. A finite fact derivation proves that its conclusion follows from those facts and allowlisted rules, not that the fact set is complete or true in the world. A candidate list with no witness is not a proof of negation. Search completeness may be reported only relative to a named finite grammar, signatures, allowed operators, and fully exhausted bounds; it does not prove global nonexistence.

## 7. Evaluation profile and knowledge states

The initial evaluator profile accepts finite explicit facts whose symbols and tuple types match the registry, and evaluates positive conjunctive propositions by bounded matching. A successful match returns a witness/environment and evidence pointing to each input fact. Evaluation is monotonic: adding facts may add derivations but does not invalidate an earlier derivation.

In the initial profile, an absent matching fact always yields `unknown`. There is no completeness-certificate branch or other route to derive `false` in this profile. Empty result sets, exhausted witness enumeration, or a missing symbol never imply `false`. This RFC establishes no global closed-world assumption. Negation, universal claims, quantifiers over an assumed complete domain, and closed-world queries require a later reviewed profile and explicit semantics; they are not executable here.

The evaluator distinguishes **representation** from **knowledge**. `Prop` is a proposition term; a derived `true`/entailed result is a separate evaluation outcome with proof evidence. A proposition without a proof may remain unknown even when it is well formed and representable. An unsatisfied GoalSpec means no candidate met the requested constraints in the checked search space; it does not mean the proposition is false in the world.

## 8. Result states

Use the existing engine status vocabulary where it has the same meaning; do not overload one state to imply a different claim. A result carries a stage and diagnostics so that the same status word at distinct stages is not ambiguous.

| State | Meaning in this contract | Not equivalent to |
| --- | --- | --- |
| `ok` | the requested stage completed under its declared profile and all required checks passed | truth outside the checked profile; owner acceptance |
| `ambiguous` | more than one licensed candidate remains | error or permission to choose arbitrarily |
| `incomplete` | a result is represented, but required evidence, fields, or obligations remain unresolved | `false` or a complete proof |
| `unknown` | evidence does not establish true or false in the active profile | empty-world truth or a negative answer |
| `unsupported` | the input is well formed enough to identify a construct outside the active profile | invalid syntax or no solution |
| `invalid` | malformed IDs/types/scope or an internally contradictory request that fails validation | merely unknown data |
| `resource_limit` | a declared bound stopped search, verification, or evaluation | exhaustive failure or proof of no result |

`not satisfied` is a goal-check result, not a world-truth value. It MAY be returned only when the verifier has decided the stated constraints for the stated finite candidate; it MUST NOT be inferred from an incomplete search. Diagnostics identify the stage, profile, and bound that produced a non-success state. In particular, `resource_limit` and `unsupported` remain visible.

## 9. Neutral examples and negative controls

All examples below use an arbitrary symbol `r`; they do not add a domain concept or approved lexical meaning.

### Two-step construction

Declare `r: Entity → Entity → Prop`, with ordered ports `left`, `right`. A path goal with input ports `start:Entity`, `end:Entity` and two linked `r` edges synthesizes:

```text
λstart:Entity. λend:Entity. ∃middle:Entity.
  And(Atom(r,start,middle), Atom(r,middle,end))
```

The output is a typed formula. Without facts or an allowed axiom, it is not an assertion that a path exists. Renaming `r` to any other valid `SymbolID` changes the data consumed by the same algorithm, not the algorithm selected.

### Type error

If `r: Entity → Entity → Prop`, `Atom(r, intervalRef, entityRef)` is rejected because `intervalRef:Interval` is not `Entity`. This does not test whether the relation is true.

### Capture avoidance

For a concrete beta step, let `bx` and `by` be distinct binder IDs and `py` a free parameter ID. The reduction is:

```text
Apply(
  Lambda(bx:Entity, Lambda(by:Entity,
    Atom(r, Var(bx), Var(by)))),
  Param(py,Entity))
→ Lambda(by:Entity, Atom(r, Param(py,Entity), Var(by)))
```

Only `Var(bx)` is replaced, the outer lambda is removed, `Var(by)` remains bound, and `Param(py)` remains free. If `py` and `by` have the same display name, the result is unchanged; IDs govern substitution. More generally, if an argument contains a free `Var` whose ID collides with a nested binder in the body, that nested binder is first alpha-renamed to a fresh ID.

Separately, let `b1` bind a displayed variable named `x`, and let `pY` be a free parameter displayed as `y`:

```text
Lambda(b1:Entity, Atom(r, Var(b1), Param(pY,Entity)))
```

Substituting `pY` with a distinct free parameter `pX` (whose display name happens to be `x`) leaves `Param(pX)` free. It is not captured by `b1`, because `ParameterID` and `BinderID` differ. A name-based exporter must alpha-rename the binder or refuse an unsafe encoding.

### Two valid candidates

For the finite goal constraint `allowed_atom_counts: [1, 2]` and a connected positive path over `r` between the declared endpoints, `λx.λy.r(x,y)` and the existential formula above both satisfy the same structural goal. They remain two separate candidates, not a disjunctive proposition. The engine does not choose one from the English meaning of `r`.

### Unsatisfied constraints

Suppose a finite goal requires at least one `Atom(r,...)` and also sets `max_atoms=0`. The finite consistency check reports `constraints_consistent=false` and the outer request status is `invalid` with a `constraint_contradiction` diagnostic. No candidate goal verdict is fabricated for an invalid request, and no `false` conclusion about a real-world relation is emitted.

### Search budget

If search reaches `max_expansions` before checking every candidate, it returns `resource_limit` with the bound and partial candidates. It MUST NOT return an empty “no solution” result or claim completeness.

## 10. Representation, evaluation, and language realization boundaries

| Operation | Input | Output claim | Required behavior on unsupported or lossy cases |
| --- | --- | --- | --- |
| Representation | typed terms, GoalSpec, or a Meaning mapping | a well-formed proposition/function structure | preserve open ports, references, multiplicity, context, polarity, and unresolved slots; reject or report residual meaning if not expressible |
| Evaluation | proposition plus explicit facts and named profile | derivation/evidence, unknown, or profile refusal | no silent fact creation, no closed-world default, no truth claim from an empty result |
| Language realization | proposition/utterance plus language-local lexicon and construction plans | one or more licensed surface candidates with assumptions/losses | keep ambiguity; reject unsupported constructions; report every loss instead of editing the proposition to fit |

A represented proposition is not automatically an assertion, question, presupposition, or request. Utterance force and discourse context require explicit structure. Conversion to or from Meaning v1 is a separate, partial mapping: where references, scope, types, proof status, or holes cannot be preserved, the converter must expose the residual or refuse. This RFC does not authorize a schema migration or claim a lossless adapter.

Negation, universal quantification, WH/gap binding, modal operators, temporal projection, and presupposition require future formal contracts. Their mention here is a boundary, not implementation permission or an assertion that the initial profile supports them.

## 11. Extension rule and completion claims

An extension adds a typed form, operator, axiom rule, or evaluation rule only with a declared signature, premises, result, validation errors, resource bound, and verifier/evaluator semantics. New domain symbols may be added as data and used by existing generic operators without recompiling domain-specific code. New operator algorithms require code and a reviewed versioned profile; a metadata record cannot name or execute arbitrary code.

Every result names its active registry/profile version and relevant limits. “Complete” means complete only relative to the finite grammar, explicit inputs, supported rules, and exhausted bounds named in that result. It never means complete natural-language understanding, complete domain knowledge, or owner approval. Technical implementation status, agent assessment, and owner acceptance remain separate.

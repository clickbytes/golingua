# Finite explicit algebra construction

**Status:** implemented for the bounded command described here; the coordinator owns the scoped mathematical assessment, supported by independent Luna review, and owner language acceptance is `NOT_EVALUATED`. This command constructs typed representations. It does not search for a construction, evaluate facts, prove a proposition, or add a derived symbol to a registry.

The earlier [formal contract](formal-contract.md) and [metadata contract](metadata.md) describe broader planned layers. This page documents the narrower implemented composition and unification surface in the current CLI. The RFC’s broader synthesis and evaluation capabilities remain outside this command.

## Command and request

`algebra build` takes a strict version-1 construction recipe and one or more explicitly selected metadata packages:

```sh
go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/composition/metadata.json \
  --input data/semantic/examples/composition/two-hop.json --json
```

Repeat `--metadata <package.json>` for every required package. Package dependencies must be supplied explicitly; paths inside metadata are not opened. `--input <file>` reads only that file. Omitting `--input`, or specifying `--input -`, reads stdin. `--json` is required. The command does not need a language runtime or application service.

The request has `schema_version`, optional `parameters`, `references`, `closed`, and `limits`, a required ordered `steps` array, and a required `result` step ID. It contains no final AST, inline sort/signature declarations, callback, or host code. See the [construction JSON Schema](../../../schemas/semantic/construction.schema.json) and the [neutral metadata example](../../../data/semantic/examples/composition/metadata.json).

Every step has a unique ID, an operation, and only that operation’s fields. Step references must point backward to earlier steps. Each step is checked using the ordinary algebra checker before it is retained. The final result ID may select an earlier step even when later trace steps are present.

The CLI admits only approved primitive signatures from the explicitly selected metadata registries. Pending, opaque, or alias symbols are not operational signatures for construction. They cannot be used to manufacture a new named derived symbol. An unknown or non-operational relation is refused. The metadata validator can inspect aliases, but this construction profile does not expand or execute them.

## Operators

Each relation or composition result is a typed expression with an optional explicit interface: an ordered list mapping port IDs to distinct leading lambda binder IDs. The interface is carried as build metadata; it is not a new term constructor. Operators select ports only through this mapping, never by guessing from equal types, labels, or role names.

| Operation | Behavior |
| --- | --- |
| `relation` | Eta-expands a selected primitive signature into a lambda for each declared port. A zero-port relation remains a proposition-valued expression. |
| `value` | Introduces a typed `Param`, `Ref`, or `Hole`. Other term constructors are not accepted as literal values. |
| `atom` | Creates one primitive atom from earlier step values in the signature’s declared ordered role sequence. |
| `and` | Conjoins earlier proposition-valued expressions. |
| `bind` | Beta-substitutes one named port, including a non-first port, and removes that port from the interface. The argument must be a checked expression of exactly the selected type. |
| `apply` | Applies and reduces a known lambda expression. Symbolic or unknown function execution is refused. When the function carries an interface, its first port is removed. |
| `join` | Requires explicit left/right port pairs of exactly equal type. Each pair becomes one existential witness shared by both bodies; the bodies are conjoined. Unpaired ports remain in left-then-right order. |
| `restrict` | Conjoins an explicitly supplied unary predicate applied to one selected port variable. Predicate and port types must match. |
| `project` | Keeps the requested subset in source order and existentially binds every unselected interface variable. An explicit empty `ports` array projects all ports away. |
| `permute` | Reorders every interface port by an explicit full bijection. It does not infer an inverse fact. |
| `unify` | Returns a typed substitution for the bounded first-order fragment described below. It is not a proposition result. |

Join orientation is never guessed. If the mapping is absent or empty, the operator returns `incomplete` with `MISSING_JOIN_MAPPING`. If remaining ports on both inputs have duplicate names, `outputs` must explicitly map each side and source port to a unique output name. This permits n-ary composition while keeping role/interface identity explicit.

For example, the supplied binary relation `r` has ordered ports `a` and `b`. The `two-hop.json` recipe returns the structure:

```text
λa. λb. ∃z. (r(a,z) ∧ r(z,b))
```

The term is an anonymous result of that recipe. It does not assert transitivity, state that any pair is related, or modify the registry. `three-hop.json` and `other-relation.json` exercise the same operator loop with three links and a different relation. `nary.json` composes a three-port relation and explicitly renames the colliding remaining `c` ports.

The [`apply-free.json`](../../../data/semantic/examples/composition/apply-free.json) example applies a relation function to a declared free parameter; [`bind-nonfirst.json`](../../../data/semantic/examples/composition/bind-nonfirst.json) binds the named second port. `restrict.json`, `project.json`, and `permute.json` give corresponding explicit recipes.

## Identity, copying, and substitution

`Param`, `Ref`, `Hole`, and `Var` have distinct identities. `SubstituteParameter` replaces only matching free parameter occurrences. Binder substitution replaces occurrences of the exact `Var` for the selected binder; it does not replace a `Param` that shares a display spelling. `Apply` removes a known lambda only after substituting its binder.

Substitution is capture-avoiding. Copies freshen binder IDs while preserving free parameter, reference, and hole IDs. Operator inputs are not mutated. Every built expression is checked, and a final expression is checked again with `closed` applied at the final result only. Intermediate open steps may therefore retain declared free parameters. A final closed expression with remaining free parameters is refused.

## Bounded unification

`unify` treats declared `Param` IDs as typed variables and supports only `Param`, `Ref`, `Atom`, and ordered `And` structure. It checks both operands against the actual supplied declarations and signatures. Atom symbols, ordered roles, types, and reference IDs must match exactly; differently named references are not equated. A successful result is a substitution map, including `{}` for an empty substitution, and the CLI reports `result_kind: "substitution"`.

The unifier follows existing substitutions and performs an occurs check, so a parameter cannot be assigned a structure containing itself. Conflicts are `invalid`; a hole remains `incomplete`. `Lambda`, `Apply`, `Exists`, arrow-typed variables/references, higher-order unification, alpha-equivalence search, and fact matching are unsupported by this unifier. No result means “the terms were not unified under this bounded procedure”; it is not a claim about truth in a model.

## Limits and statuses

The input JSON is limited to 1 MiB and nesting depth 256. A recipe contains at most 64 steps. The ordinary checker defaults and caps each expression at 4096 nodes and depth 64, types at depth 32, declarations at 512, and relation ports at 16. Retained step results plus detached trace expressions/substitutions are bounded to 16384 aggregate term nodes. Exceeding a limit returns `resource_limit`; it is not a failed proof or an empty result.

A typed `Hole` can remain in an expression result and yields `incomplete` with its requirement. Using unresolved holes in an operator that cannot proceed may return an incomplete refusal. Malformed requests and type/scope conflicts return `invalid`; unknown operations or unsupported term forms return `unsupported`. No failure invents a term, a port mapping, or a witness to make the recipe succeed.

Exit codes are `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, and `incomplete` 7. Infrastructure errors use 1. These statuses describe the construction procedure, not linguistic acceptance. The CLI’s `ok` means the explicit recipe produced a well-typed representation under the selected declarations; it does not say that the represented proposition is true.

## Examples and verification record

All examples in [`data/semantic/examples/composition/`](../../../data/semantic/examples/composition/) use neutral relation symbols and `internal-neutral-example` provenance for issue #381. The metadata has no axioms. These are mathematical data examples, not claims about English or Polish, and their author approval flag is not owner language acceptance.

CLI observations for this work are kept separately in the DSE evidence record when the coordinator records them. The existence of example files or this document is not itself an observation. Broader DSE synthesis/evaluation and owner acceptance remain separate from this implemented finite composition command.

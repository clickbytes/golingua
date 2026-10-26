# Negation, quantifiers, and Ask scopes

This contract distinguishes representable scope from executable query scope.
The typed IR can carry `Not`, `Quantifier`, and `Ask` terms. Their presence does
not add truth rules, a complete domain, or language-specific behavior.

## Typed terms

| Term | Type and fields | Binding behavior |
| --- | --- | --- |
| `Not` | `Prop`; requires `body` | Does not introduce a binder. |
| `Quantifier` | `Prop`; requires `id`, data-sort `type`, `quantifier`, and `body`; optional `restrictor` and numeric `count` | Its binder is in scope in both restrictor and body. |
| `Ask` | `Question`; requires `body`; ordered `answers` may be empty; optional `restrictor` | Each answer ID is in scope in both restrictor and body. |

The supported quantifier labels are `some`, `all`, `exactly`, `at_least`, and
`at_most`. `some` and `all` forbid `count`; the other three require a
non-negative integer, including zero. The checker confirms the term's typed
shape and binder scope. It does not assign these labels an evaluation rule.

An omitted or `null` restrictor means unrestricted scope. An empty object `{}`
is malformed; it is not an empty or unrestricted domain. `Not`, quantifiers,
and `Ask` introduce no existential witness merely by being represented.

`Ask.answers` is an ordered list of typed declarations. An empty list represents
a yes/no question; a non-empty list defines the answer columns and their order.
`Project` can choose answer ports before building an `Ask`; ports omitted by
`Project` are existentially closed by that operator. `Ask` itself binds its
answer declarations in both its optional restrictor and body. Its result type
is `Question`, distinct from the `Prop` type of its body.

## Build recipes

The version-1 construction DSL adds three operations:

- `not`: one earlier expression or predicate function; a function keeps its
  ordered leading lambda interface while its proposition body is wrapped.
- `quantifier`: `inputs` contains a one-lambda predicate and optionally a
  one-lambda restrictor of the same type. `quantifier` names the scope kind;
  `count` is present only for numeric kinds.
- `ask`: `inputs` contains an earlier proposition or function, and optionally
  a restrictor. A function's leading interface becomes the ordered answers.
  A proposition creates a yes/no `Ask`. A non-empty answer interface requires
  a restrictor function with the same ordered types; a yes/no restrictor is a
  proposition.

Each input names an earlier step. There is no raw-AST escape, inference from
labels, or implicit domain construction. The result remains a typed expression
and operation trace; it is not inserted into the fact store.

## Evaluation boundary

The `positive-facts-v1` evaluator accepts a root `Ask` whose body, after beta
lowering, is a positive `Atom`, `And`, or `Exists` expression. Its optional
restrictor is conjoined with the body, and its embedded ordered answers are the
only projection authority. `algebra evaluate` and evaluation-proof verification
therefore omit `--select` for an `Ask`; supplying it is refused. A non-Ask
closed positive proposition can use an explicit empty selection (`[]`), which
produces the yes/no answer shape. A missing match, including an empty fact set,
is `unknown`, not false.

`Not` and all quantifier kinds remain unsupported for evaluation, including
`exactly 3` when three matching objects happen to be present. The current finite
fact snapshot is not a complete world domain. A nested `Not`, quantifier, or
`Ask` is not dropped to make the positive body executable. Malformed types,
binder scope, and negative counts are `invalid`; unsupported operators are
`unsupported`; resource exhaustion remains `resource_limit`.

The capability manifest reports representation and evaluation separately.
`Not` and quantifiers are representable but not evaluatable in this profile.
`Ask` evaluation is conditional on the root positive-body shape above; the
manifest does not claim general `Question` evaluation. Positive answer proofs
attest explicit witnesses only. They do not prove full question truth,
completeness, quantifier semantics, or owner acceptance. No PL/EN parser or
realizer integration is part of this scope.

Neutral construction, query, fact, normalization, and export/replay inputs are
in [the scope examples](../../../data/semantic/examples/scopes/). The actual CLI
commands and exits are recorded in [DSE-06-01 evidence](../../work/dse/evidence/06-01/README.md).

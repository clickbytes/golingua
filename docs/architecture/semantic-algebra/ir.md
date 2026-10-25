# Typed expression IR v1

This page documents the typed expression representation and checker implemented for DSE leaf #379. It describes structural typing and scope validation only. The [formal contract](formal-contract.md) remains the normative design reference; the strict wire shape is defined by [`expression.schema.json`](../../../schemas/semantic/expression.schema.json). This IR is not an ontology, fact store, truth value, or language-specific meaning.

## Document and declarations

A version-1 document has `schema_version` and `expression`, with optional inline `sorts`, `signatures`, `parameters`, `references`, `closed`, and `limits`:

| Field | Purpose |
| --- | --- |
| `sorts` | Nominal sort declarations with stable `id` and optional presentation `label`. |
| `signatures` | Uninterpreted proposition symbols, each with an ordered `ports` list. Every port has `role` and `type`. |
| `parameters` | Typed free `ParameterID` declarations. A `Param` must match one exactly. |
| `references` | Explicit typed request references. A `Ref` must match one exactly; this is not identity discovery. |
| `closed` | When `true`, reject an expression containing a free `Param`. |
| `limits` | Optional finite checker limits. Zero selects the corresponding default. |

Labels are presentation only. They do not identify a type, symbol, binder, parameter, or reference and do not affect checking. IDs are nonempty valid UTF-8, at most 256 bytes; labels are at most 1024 UTF-8 bytes.

## Types and signatures

The exact type kinds are `Entity`, `Event`, `Interval`, `Prop`, `Nominal`, and `Arrow`. `Nominal` requires a `sort_id` declared in `sorts`; `Arrow` requires `from` and `to`. Equality is structural and exact; no subtype is inferred. A relation signature's ports must have base types (`Entity`, `Event`, `Interval`, `Prop`, or a declared `Nominal`); an `Arrow` port is rejected. `Prop` is a proposition sort, not a Boolean or a truth result.

Signature port order is semantic. An `Atom` must supply the full list of arguments in that exact order, with each argument's `role` equal to its corresponding port role and its value type exactly equal to the port type. Duplicate roles, unknown roles, wrong order, arity mismatch, and type mismatch are distinct structural failures. A symbol ID has no built-in interpretation.

Arrow types remain available for function terms. `Lambda` produces an arrow type and `Apply` checks the function and argument types. The checker does not execute beta reduction; function typing is not function execution.

## Term variants

The tagged `kind` values are capitalized and closed:

| `kind` | Required fields besides `kind` | Meaning for structural checking |
| --- | --- | --- |
| `Var` | `id` | Refers to an in-scope `Lambda` or `Exists` binder by `BinderID`. |
| `Param` | `id`, `type` | Free typed parameter, declared in `parameters`; reported in `free_parameters`. |
| `Ref` | `id`, `type` | Explicit request reference, declared in `references`. |
| `Hole` | `id`, `type` | Typed unresolved slot; returns `incomplete` and a requirement, never a witness or fact. |
| `Atom` | `symbol`, `args` | Applies one declared proposition signature to its ordered role arguments. |
| `And` | nonempty `operands` | Conjunction whose operands must all have type `Prop`. |
| `Exists` | `id`, `type`, `body` | Binds a fresh ID in the body; body must have type `Prop`. |
| `Lambda` | `id`, `type`, `body` | Binds a fresh ID in the body and returns an `Arrow` type. |
| `Apply` | `function`, `argument` | Checks exact arrow-domain equality and returns the arrow codomain; does not reduce. |

Every variant may carry an optional `label`. Fields not allowed for a variant are refused even when supplied as `null` or a zero value. Binder IDs are globally unique in one expression; variable lookup uses IDs and lexical scope, not labels. `ParameterID`, `ReferenceID`, `BinderID`, and `HoleID` remain distinct namespaces. References are declarations, not proof of entity identity. Holes remain unresolved and make no existence claim.

## Limits and outcomes

`Decode` enforces a 1 MiB document-byte limit before parsing, rejects invalid UTF-8, duplicate JSON keys and trailing documents, and bounds JSON nesting at 256 before typed decoding. The checker defaults and caps are:

| Limit | Default/cap |
| --- | ---: |
| Expression nodes | 4096 |
| Expression depth | 64 |
| Type depth | 32 |
| Aggregate sort/signature/parameter/reference declarations | 512 |
| Ports and atom arguments | 16 |

Zero uses the default; a negative value or value above the cap is invalid. Exceeding a work bound returns `resource_limit`, not a claim that an expression is false or has no interpretation. A valid expression with unresolved typed holes returns `incomplete`; free parameters are reported and are rejected only when `closed` is true. Successful checking means structural checking completed, not that a proposition is true.

`Decode` and `Encode` preserve the in-memory document contract, including labels, ordered roles, binder IDs, parameters, references, and limits. `Encode` validates without mutating the caller and permits incomplete hole documents. Canonical output is semantically structured JSON, not a byte-for-byte copy of the source.

## Boundaries

The initial CLI command checks and canonicalizes this IR. It does not synthesize expressions, execute `Apply`, evaluate facts, prove a proposition, load a runtime registry, or connect a language adapter. See the [DSE formal contract](formal-contract.md), [Meaning v1 compatibility contract](compatibility.md), [CLI usage](../../cli.md), and [#379 evidence index](../../work/dse/evidence/03-01/README.md).

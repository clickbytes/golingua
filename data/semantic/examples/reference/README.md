# Neutral request-context examples

These coordinator-authored records exercise request-local referent identity,
explicit mention evidence, assumptions, presuppositions, losses, and context
preservation. They do not encode owner-approved language meanings. The URNs in
`SourceRef` identify neutral diagnostic proposals; they are not external
evidence.

`resolved-document.json`, `unresolved-document.json`, and
`conflicting-document.json` differ in explicit mention evidence. Referents A
and B intentionally share a type and classification while retaining separate
IDs. `zero-arity-document.json` carries a condition with an empty `args: []`.
The `metadata-zero-arity.json` package adds the `t0` primitive used by that
condition and by the description goal; `metadata.json` intentionally omits it.

`description-goal.json` binds goal port `y` to Ref A and leaves `x` open. Its
constructed expression remains an open description, not a chosen entity or a
presupposition witness. The build, normalize, export, and replay requests show
how the same context payload is retained across request forms. The
`presupposition-is-not-fact-document.json` and `facts.json` are separate inputs;
the condition does not add evidence to the fact set.

These inputs are small CLI examples, not a test corpus or acceptance gate. See
the [reference contract](../../../../docs/architecture/semantic-algebra/reference.md)
and [#392 evidence](../../../../docs/work/dse/evidence/06-02/README.md) for
observed results and limits.

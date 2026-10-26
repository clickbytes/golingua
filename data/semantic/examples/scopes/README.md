# Neutral scope examples

These are small, source-neutral examples for typed construction and the
positive `Ask` evaluator. They are not language data or owner-approved semantic
rules.

- `build/` contains explicit recipes for `Not`, both `Not(All(...))` and
  `All(Not(...))`, numeric `exactly 0`, and ordered `Ask` projections.
- `query/` contains the corresponding typed expressions, including negative
  scope cases and reversed answer order.
- `facts-simple.json` contains explicit positive `r(A,B)` and `r(B,C)` records;
  `facts-empty.json` has no records. Missing matches remain unknown.
- `normalization/`, `export/`, and `replay/` hold the neutral inputs for the
  recorded normalization and artifact flow. The normalization request is
  declaration-free because selected metadata supplies its inventory.

The [scope contract](../../../../docs/architecture/semantic-algebra/scopes.md)
and [CLI usage](../../../../docs/cli.md) describe the limits and refusals. The
observed command receipts are in
[DSE-06-01 evidence](../../../../docs/work/dse/evidence/06-01/README.md).

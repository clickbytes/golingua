# Neutral positive-query example

These files show the #389 finite positive-facts workflow over an explicitly supplied neutral metadata package and fact input. They are mathematical API examples, not language data or a claim about a complete world model.

The constructed query asks for `x` and `y` such that `r(x, middle)`, `r(middle, y)`, and `p(y)`. The saved `candidate.json` is the checked canonical query document; `certificate.json` is its construction certificate; `select.json` selects the two leading lambda binders by position.

- `facts-simple.json` produces answer `(A, C)` with one positive proof.
- `facts-branch-conflict.json` produces `(A, C)` with two distinct path witnesses. The `A→B` proof leaf records both the positive fact ID and its opposing negative fact ID; the answer reports `has_conflict: true` while remaining supported by positive evidence.
- `facts-changed-target.json` produces `(A, D)` from the supplied `A→B→D` path and `p(D)`.

The result snapshots are literal CLI outputs from the run recorded in `docs/work/dse/evidence/05-02/`. The fact and formal metadata snapshot IDs are separate. No fact is inferred from absence, and no current-time value is supplied.

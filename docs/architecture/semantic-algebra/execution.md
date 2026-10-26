# Finite positive-query evaluation

Issue #389 adds a terminal path from a checked algebra query and explicit fact records to bounded answer candidates with a locally assembled proof DAG. It does not add general Datalog, closed-world negation, a world model, a language route, or the later #390 proof-equivalence verifier.

## Inputs and command

The command requires named files; it does not read stdin:

```sh
golingua algebra evaluate \
  --metadata data/semantic/examples/evaluation/metadata.json \
  --expression data/semantic/examples/evaluation/candidate.json \
  --facts data/semantic/examples/evaluation/facts-simple.json \
  --select data/semantic/examples/evaluation/select.json \
  --context-id eval-neutral --profile positive-facts-v1 --json
```

The expression is a self-contained typed algebra document whose declarations are checked against explicitly selected metadata. The facts file supplies the request-scoped referent domain and assertions. The selection file is a strict JSON array of either `{ "kind": "lambda", "index": 0, "as": "start" }` or `{ "kind": "parameter", "id": "p", "as": "out" }`. Lambda indices select leading binders after bounded beta lowering; parameter IDs remain in a distinct namespace. An empty array is a closed query. `--context-id` must match the fact document; `--profile` must be `positive-facts-v1`.

Optional `--at` accepts UTC RFC3339Nano ending in `Z`. If the fact input already has an instant, the flag must denote the same instant; otherwise it fills the instant on a copied request. If neither provides an instant, only timeless records are active. The evaluator never chooses the current time.

`--evaluation-limit name=value` may be repeated with unique names: `max_plan_atoms`, `max_variables`, `max_intermediate_rows`, `max_results`, `max_proof_nodes`, `max_proofs_per_answer`, `max_proof_depth`, `max_work_units`, `max_output_bytes`, and `timeout_millis`. Zero selects a finite default. Defaults are respectively 64, 128, 256, 64, 4096, 8, 128, 1,048,576, 8 MiB, and 2,000 ms; hard caps are 128, 512, 4,096, 512, 16,384, 64, 128, 16,777,216, 16 MiB, and 30,000 ms. Shared algebra term limits also apply. A resource limit or cancellation returns only fully committed answers and proof nodes, with `complete: false`; a candidate is staged before it is committed so a partial proof cannot appear in the retained prefix. The separate CLI response envelope is capped at 16 MiB.

## Result scope

`knowledge` is `supported` when at least one answer has explicit positive support and `unknown` otherwise. It is never `false`. Negative records and positive/negative conflicts are reported as evidence; negative evidence does not negate a positive proof. Missing evidence is `unknown`, not negation-as-failure. `complete: true` means the matcher exhausted its finite scan over the supplied typed reference union and selected evidence snapshot. It does not claim that the reference domain or fact set is complete for the world.

The result includes distinct facts-snapshot and formal-snapshot identifiers. These pins describe different selected inputs and must not be equated. The finite reference domain is the union of the query's explicit typed references and fact-input referents; unused lambda and Exists binders range only over that explicit typed union. Repeated scoped groundings can produce separate proof IDs up to the proof-per-answer cap. Ordered answer tuples retain exact Ref IDs, including distinct event IDs of the same type.

## Proof DAG

Each matching source Atom occurrence creates a `query.fact_match.v1` leaf containing its exact grounded Atom, every positive FactID and corresponding source record, the variable bindings relevant to that Atom, and any opposing negative FactIDs for the same atom. The evaluator does not select an arbitrary first source. `query.and.v1` nodes preserve the source conjunction's child order. `query.exists.v1` nodes record the exact scoped binder-to-Ref witness and remove that binder from outward bindings. An unused Exists or selected leading lambda receives a `query.reference.v1` node backed by a declared typed Ref, not an invented fact. The root `query.project.v1` node contains only selected columns. Premises point to earlier nodes; IDs are deterministic within a result (`p1`, `p2`, ...), and answer IDs are similarly local (`a1`, `a2`, ...).

This is a bounded construction against the already checked plan and active fact rows. The #389 runtime does not independently replay arbitrary submitted proof DAGs, establish unrestricted equivalence, infer transitive rules, or prove global consistency. The separate #390 issue owns later proof verification.

## Recorded neutral run

The saved example uses explicit `r` and `p` relations. With `facts-simple.json`, the CLI returned `(A, C)` and a proof whose `r` leaves cite `f-r-ab` and `f-r-bc`, and whose `p` leaf cites `f-p-c`. With the branch-conflict input, the same answer has two path witnesses (`A→B→C` and `A→D→C`); the first path leaf also lists `f-r-ab-neg`, and `has_conflict` is true. The changed-target input selects `(A, D)`. Empty and negative-only inputs produce no supported answer and retain `knowledge: unknown`.

The captured outputs, exact command records, source hashes, and the separate bounded inspection of emitted proof nodes are linked from [#389 evidence](../../work/dse/evidence/05-02/README.md). Agent technical assessment and owner language acceptance are separate; owner acceptance is **NOT_EVALUATED**.

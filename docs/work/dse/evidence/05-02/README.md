# DSE-05-02 / issue #389 — query evaluator evidence

This receipt records the frozen branch CLI run for #389, based on `cf21625b7b85bcf2735c89bb02ffd11fdcf6993f` on `dse/05-02-query-evaluator`. It is local technical evidence, not a PR, main-branch CI receipt, or owner acceptance.

## Frozen build and source

The frozen executable was `/tmp/golingua-dse/bin/golingua-389-final`, SHA-256 `22e48da11cf8291f67c87d691a48d739c4f43ddb3964a8a2d53bdfe00bdf069e`. `environment.json` preserves the coordinator's literal technical observation, including seven source-file hashes and the observed `gofmt`, `go vet ./...`, `go build ./cmd/golingua`, frozen build, `git diff --check`, `--help`, and `--version` exits. The executable reported `0.1.0-alpha.1`.

`observations.jsonl` is the literal 38-record CLI log from the unchanged observation script. It includes synthesis and independent `verify-goal` setup, followed by fact/query/resource observations. Exit counts were 0: 27, 2: 2, 3: 1, 7: 1, 4: 7; stderr was empty and the largest saved stdout was 13,638 bytes. Each record's stdout file is copied under `cli-outputs/`; the JSONL itself retains the original `/tmp` paths and hashes. `inputs/` copies the 39 original small independent JSON inputs byte for byte. `cli-outputs/` also keeps the generated candidate, certificate, selection, result JSON files, and stderr captures. These are observed neutral formal inputs, not a new test runner or language corpus.

The product example in `data/semantic/examples/evaluation/` is the small reusable CLI walkthrough. Its metadata contains eight neutral predicates; this query uses `r` and `p`. The saved canonical candidate, certificate, selection, and simple/conflict results were copied from the frozen CLI output.

## Observed behavior

- Construct, then verify the generated candidate in a separate CLI process: both returned `ok`; the certificate was accepted by the independent #388 goal checker.
- On the simple facts, the query returned `(A, C)` with proof leaves citing `f-r-ab`, `f-r-bc`, and `f-p-c`. On changed-target facts it returned `(A, D)`. Reordered argument arrays preserved role-based matching; swapping role values did not produce support.
- On branch-conflict facts, `(A, C)` retained two distinct proofs. The `A→B` leaf included positive `f-r-ab` and opposing negative `f-r-ab-neg`; the answer marked `has_conflict: true` while remaining positively supported.
- Empty, negative-only, no-transitive-rule, no-explosion, empty-domain and post-interval cases returned `knowledge: unknown`, not false. Timed start/end were active; no `at` did not activate bounded-time records.
- Repeated ternary variables, Exists, parameter namespace, separate event references, zero-arity atoms, beta-lowered input, unused lambda, unused Exists and disconnected joins completed with distinct proof/grounding behavior as appropriate. Unused leading lambda and Exists variables ranged over the explicit typed reference union only.
- Resource caps preserved completed prefixes where available and reported `complete: false`; proof-node/depth and intermediate-row caps retained no incomplete proof. Wrong context was invalid; residual Apply was unsupported; a typed Hole remained incomplete.

`proof-observation.json` is the coordinator's separate structural inspection of the returned DAGs against the selected fact rows, roles, types, time, scopes, Exists witnesses, projection, and premise ordering. It covered 135 DAG nodes and 30 answers across 30 result objects. This is bounded inspection of this run, not a general proof-equivalence checker, completeness proof for the world, or a substitute for the future #390 work.

The result distinguishes the fact snapshot from the formal metadata snapshot. Positive evidence supports `knowledge: supported`; absence remains `unknown`, negative evidence is preserved separately, and the declared reference/fact domain is not claimed complete. No arbitrary first source record is selected: a matching leaf retains all applicable positive FactIDs and source records, plus opposing negative FactIDs.

## Assessment status

The observed CLI workflow and bounded proof structure were technically demonstrated for this neutral profile. The captured statuses are runtime outcomes, not a percentage quality score. No language behavior was tested. Owner language acceptance is **NOT_EVALUATED**. The issue-map entry remains ACTIVE with PR and merge fields unset; PR review and main CI are pending their canonical receipts.

The documented construct → verify-goal → evaluate shell flow was also executed with relocated temporary files and the frozen binary: [observation](doc-flow-observation.json), [literal output](doc-flow-output.jsonl). Both command outputs were ok and the evaluated tuple was (A, C).

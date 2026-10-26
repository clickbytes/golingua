# DSE-05-03 / issue #390 — equivalence and proof-checking evidence

This receipt records branch-local technical observations for issue #390. The
issue-map entry is still `ACTIVE`; no PR or merge SHA is claimed here. PR review
and main-branch CI remain pending their canonical GitHub receipts. Owner
language acceptance is **NOT_EVALUATED**.

## Frozen source and executable

The observed executable was
`/tmp/golingua-dse/bin/golingua-390-final`, SHA-256
`3d16d21d57cf8163ec553bd222ac1c8c34e4d7153d744ea68c7a0ce02840a447`. The
coordinator's technical receipt records the matching `golingua-390-verified`
build hash, `0.1.0-alpha.1`, `gofmt -l`, `go vet ./...`, `go build`, `--help`,
and `--version`; each command exited 0. The seven frozen source/schema hashes
are preserved in [source-hashes.json](source-hashes.json), and the full
command observations are in [environment.json](environment.json).

## Literal CLI observations

`observations.jsonl` preserves the original 37-record CLI log, and
`extra-observations.jsonl` preserves three additional open-world/evaluation
controls. Their combined exit counts are 0: 16, 2: 14, 3: 3, 4: 6, and 5: 1.
All recorded stderr values are empty; the largest saved stdout is 5,107 bytes.
Each log retains its original temporary paths and stdout hashes. Copies of the
referenced inputs and stdout/stderr artifacts are under `inputs/` and
`cli-outputs/` for inspection. The evaluation facts include the exact
`facts-simple.json` input from the preceding #389 evidence receipt, copied as
`inputs/facts-simple-389.json`.

The observations show:

- Alpha, beta, and explicitly licensed conjunction-order comparisons returned
  `proved_equivalent`; separate `verify-proof --claim equivalence` runs accepted
  each emitted certificate. The independent check did not call comparison as
  its proof oracle.
- Conjunction reordering without the rule, role swaps, distinct reference IDs,
  and distinct event IDs returned `undetermined`; they did not produce a
  negative equivalence proof. Entailment and countermodel claim types were
  refused as `unsupported`.
- The supplied evaluation witness was independently accepted with one positive
  answer and 14 verified proof nodes. `completeness_verified` remained false.
  The tampered selector and answer, conflict, context, depth, evidence
  fingerprint, Exists witness, fact ID, premise, profile, rule, and source
  controls were refused. An empty witness set returned `unknown`, not false.
- Comparison work/certificate limits returned `resource_limit`; resource
  exhaustion did not become a negative proof.
- Existing rewrite and goal proof routes remained usable. This issue did not
  change the pre-existing binary question API or `SemanticEquivalence`.

The corrected evaluation selector is the exact input
`inputs/select-evaluation-original.json`. A different selector fixture was
retained as a negative control and correctly failed fingerprint matching; the
earlier temporary fixture mismatch was an input-wiring error, not a code change
or a changed expected result.

## Examples and schemas

The reusable neutral command files live in
[`data/semantic/examples/proofs/`](../../../../../data/semantic/examples/proofs/).
The comparison certificate and evaluation result are copies of actual CLI
outputs. The independent JSON Schema 4.26.0 observation in
[schema-observation.json](schema-observation.json) reports no validation errors
for the three equivalence certificates or the evaluation proof. Equivalence
and evaluation schemas are additive definitions under `$defs`; the existing
proof schema root union was not changed to include these new document kinds.

The three commands shown in the documentation were subsequently run against
the frozen `golingua-390-verified` executable. All exited 0: comparison returned
`proved_equivalent`, its independent certificate check returned valid, and the
evaluation proof returned `positive_support_verified` with completeness false.
The [flow record](doc-flow-observation.json) and its three JSON outputs are
preserved under `cli-outputs/`; these are additional to the original 40
observations above. These checks concern neutral algebraic inputs; they do not
establish a complete model checker, countermodels, entailment, answer
completeness, language meanings, or owner acceptance.

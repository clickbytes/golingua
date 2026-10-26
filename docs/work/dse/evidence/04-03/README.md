# DSE-04-03: independent verification of finite constructions

This record documents the frozen local CLI snapshot for issue #388. It is
technical evidence for a bounded structural profile, not an owner acceptance
or a claim about facts or world truth. The owner-language status is
**NOT_EVALUATED**. The issue remains active at this snapshot; its PR and merge
receipt are still pending.

## Frozen inputs and commands

The base was `ec10a2cb178d7ad5587884c730a7762fed506d90` on
`dse/04-03-synthesis-verifier`. The frozen binary was
`/tmp/golingua-dse/bin/golingua-388-final`, SHA-256
`dad3fe833b8748fb2e9f1bcfa4a2b54873fe0f7a1fec269ab1469b769ccc5ac2`.
`environment.json` records its source-file hashes and actual commands. The
source checks `gofmt -l internal/synthesis internal/cli/algebra.go
internal/cli/cli.go`, `go vet ./...`,
`go build -o /tmp/golingua-dse/bin/golingua-388-final ./cmd/golingua`, and
`git diff --check` exited 0; `--help` and `--version` also exited 0.

The literal coordinator CLI receipts are `observations.jsonl` (55 records) and
`extra-observations.jsonl` (25 records). The copied small JSON files in
`inputs/` and `cli-files/` retain the input snapshots, mutations, and captured
stdout documents needed to inspect these records. The two JSONL files preserve
their original `/tmp/golingua-dse/...` paths; the captured files are included
here under `cli-files/`. The 8 MiB oversize certificate input was not copied.
`certificate-oversize-manifest.json` preserves its observed byte length and
SHA-256, a recipe that recreates those exact bytes from the included valid
certificate, and the small `verify-wire-oversize` refusal receipt. The receipt
is exit 4 with `resource_limit` and `JSON_BYTE_LIMIT`.

The intentional 10,000-space input is also omitted from the repository; its
[manifest](whitespace-input-manifest.json) preserves the exact bytes, digest,
recipe and requested raw-byte limit. The original experiment and its refusal
receipt are unchanged.

Across the 80 records, exit codes were 0: 28, 2: 41, 3: 2, 4: 8, and 7: 1.
Seven construct candidates, eight ambiguous-search candidates, and two
explicit-reference search candidates were produced and independently
verified. The successful end-to-end path was construction, extraction of its
canonical candidate and certificate, `verify-goal` in a separate command,
then export and replay in a new process. Every retained search alternative
also passed the independent verifier. A single candidate did not establish
search uniqueness or completeness.

The refusal controls included wrong path orientation/count, missing unary
constraint, extra atoms or unused existentials, free variables, altered goal,
candidate, snapshot, status, binder, scope, obligation, assumption,
interface, operation, and rule claims; malformed/duplicate/trailing wire
input; disallowed operators; and work, step, and byte limits. Refusals were
invalid (exit 2), unsupported (3), resource-limited (4), or incomplete (7),
according to the case. They were not interpreted as a successful verification.
The additional records exercised shared references, one-hop and projected
goals, reverse roles, duplicate unary constraints, and typed holes. The typed
hole remained incomplete. The final run corrected an earlier observed
`interface: null` output to an empty array and repeated the same inputs; the
pre-correction output is preserved outside this evidence folder and is not
counted as final-source evidence.

`schema-observation.json` records validation with `jsonschema` 4.26.0 installed
outside the repository: the schema meta-check passed and 18 captured
certificates (17 producer outputs plus one whitespace-budget control) had no
schema validation errors. This tool is not a product or CI
dependency. `agent-assessment.json` records the separate bounded assessment;
neither that assessment nor the CLI receipts constitute owner approval.

The corrected synthesis documentation workflow was also executed with `bash -e`
and `go run ./cmd/golingua`, with its temporary files relocated to an isolated
directory. [The observation](doc-flow-observation.json) records exit 0 and
unchanged source hashes; [the replay output](doc-flow-replay.json) is `ok`.

## Interpretation and limits

The verifier independently checks types, goal refinement, operator replay,
certificate claims, scopes, and structural constraint coverage against the
selected metadata and goal. The producer's `goal_verified: false` remains an
untrusted producer field. `fully_validated` is a field on the synthesis result;
the separate `GoalVerificationResult` reports its component checks, and the
synthesis flag is true only when its required checks succeed. This is not fact
evaluation, global theory consistency, theorem proving, world-truth evidence,
or a language route. Search `origin` is a label, not evidence that search ran;
one candidate certificate cannot establish search completeness or uniqueness.
Resource exhaustion and unresolved inputs are not acceptance.

The 500-reference response-size observation in the #387 synthesis contract is
historical: it predates the goal-certificate payload fields, which change
candidate sizes. It is not a measurement of the #388 output contract.

The repository includes a small reproducible example at
[`data/semantic/examples/verification`](../../../../../data/semantic/examples/verification/README.md),
the [synthesis contract](../../../../architecture/semantic-algebra/synthesis.md),
and the [certificate schema](../../../../../schemas/semantic/synthesis-certificate.schema.json).
The frozen source is not yet represented by a merged PR or a main-branch CI
receipt. The issue map therefore leaves its PR and merge fields empty. Owner
language acceptance remains **NOT_EVALUATED**.

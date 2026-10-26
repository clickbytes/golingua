# DSE-04-01: Goal construction evidence

This record preserves the final local CLI observations for #386 after the
bounded unary-validation correction was integrated into the source used for
the final build. It separates those results from earlier pre-integration
observations and the first validator API probe. The coordinator's structural
inspection is bounded review evidence; it is not a production proof checker,
world-truth evaluation, or quality gate.

## Final source and command record

- [`observations.jsonl`](observations.jsonl) contains 59 literal command
  records, including `gofmt`, `go vet ./...`, `go build ./cmd/golingua`, help,
  version, positive construction, refusals, and artifact export/replay.
- [`environment.json`](environment.json) records the exact binary
  `/tmp/golingua-dse/bin/golingua-386-corrected`, its SHA-256, Go version,
  source hashes, input hashes, and product-data hashes. It records no final
  source-hash mismatches. Post-build inputs are explicitly marked as created
  after the binary was built.
- [`invariants.json`](invariants.json) records coordinator inspection of the
  final outputs: 14 goal checks, explicit topology and role orientation,
  distinct fragment occurrences, goal interface/type preservation, and
  operator budgets. These are scoped inspection results, not an independent
  proof of goal truth.

The final observations include successful one-, two-, and three-hop
constructions, unary restrictions, fixed references, port permutation,
post-build symbols and roles, and stdin input. Missing orientation/symbol,
multiple permitted counts, and typed holes return `incomplete` without a
candidate expression or export request. Malformed inputs and unknown roles are
refused. Atom and operator budgets return `resource_limit`; a disallowed join
returns `unsupported`. The post-integration controls show that an unknown
nonempty unary symbol is `invalid` even when its port is absent, while a known
symbol with the wrong arity is `unsupported`. An unresolved path does not mask
an invalid unary symbol. A reference object with an extra `label` field is
rejected.

The observations also record canonical-document checking without metadata,
refusal of inline declarations when metadata is selected, and successful
metadata checking after removing only the inline `sorts` and `signatures` for
that separate check. The synthesized export request retains the hydrated
canonical document and passes the existing export flow. Replay in a distinct
process returns an artifact exactly equal to the exported artifact; a wrong
expected context is refused. Existing catalog summary, metadata inspection,
build, and explicit-facts conflict controls are recorded as unchanged
successful CLI operations. These checks establish structural/runtime behavior
for this bounded command only, not world truth or language integration.

## Inputs and pipeline artifacts

- [`inputs/independent/`](inputs/independent/) preserves the independent goal
  inputs used for the main construction/refusal observations.
- [`inputs/extra/`](inputs/extra/) preserves additional post-build inputs,
  including new symbols/roles and malformed controls.
- [`inputs/validation-controls/`](inputs/validation-controls/) contains the two
  post-integration unary controls and the exact pre-integration validator
  source note.
- [`pipeline/`](pipeline/) contains the literal canonical document, export
  request/output, replay request/output, wrong-context request, and the
  external-check document used in the separate-process flow.
- [`pre-integration/`](pre-integration/) retains the earlier 36-record
  observation log and its environment snapshot. These records predate the
  unary-validation integration and are not substituted for the final log.
  `initial-validator-api.json` is an earlier API diagnostic only; its renamed
  goal used a base metadata package and is not evidence for final hash
  invariance.

The product's nine goal examples are in
[`data/semantic/examples/goals/`](../../../../../data/semantic/examples/goals/).
This evidence directory copies the independent and diagnostic input sets so
the final command record can be inspected without treating them as product
fixtures or a test runner.

## Scope and status

The command constructs finite typed representations from explicit constraints
and selected approved primitive metadata. Its obligation paths identify
distinct occurrences in the final expression; the `constructed_provenance`
label does not independently prove that a constraint is true. The selected
metadata, artifact, and replay context do not create a fact store, world model,
language route, automatic approval, or owner acceptance.

Technical command checks and coordinator inspection are recorded separately
from owner language evaluation. Owner acceptance remains **NOT_EVALUATED**.
The issue-map entry remains active with PR and merge receipt unset; PR/main CI
status is pending the canonical GitHub completion receipt.

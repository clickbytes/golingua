# DSE-06-01 runtime evidence

This record covers the final local #391 source snapshot after the closed-query
selection correction. It separates CLI observations, structural schema checks,
the coordinator's semantic review, and owner acceptance.

## Frozen inputs and execution

- Binary: `/tmp/golingua-dse/bin/golingua-391-final`
- SHA-256: `41753732f4d459e5aaad68768cc16d875b73a1cc6971bb9d576347f596586b66`
- Source hashes: [`391-source-hashes.json`](391-source-hashes.json), 19 Go/schema
  files
- Technical commands: [`391-technical-observation.json`](391-technical-observation.json)
  records `gofmt -l` (empty output), `go vet ./...`, build, `--help`, and
  `--version`; each exited 0.
- CLI records: [`cli-observations-67.jsonl`](cli-observations-67.jsonl),
  [`build-observations-18.jsonl`](build-observations-18.jsonl), and
  [`extra-observations-7.jsonl`](extra-observations-7.jsonl). Their stdout files
  are copied under [`outputs/`](outputs/); inputs are under [`inputs/`](inputs/).
  The JSONL command paths preserve the original `/tmp/golingua-dse` invocation
  provenance; the matching inputs and outputs are included here by basename in
  their respective subdirectories. Inputs added for the seven extra controls
  are in [`inputs/extra/`](inputs/extra/).
- Schema observations:
  [`391-schema-observation.json`](391-schema-observation.json) records 21
  supplied DSL/document/artifact files checked with jsonschema 4.26.0; all had
  zero validation errors and the meta-schemas were valid.

The three literal command logs contain 92 records: exit 0: 75, exit 2: 9,
exit 3: 6, and exit 4: 2. These counts summarize observed commands; they are
not an automated quality score.

The seven literal command blocks shown in `docs/cli.md` were also run with
`go run ./cmd/golingua` from the repository root. Their commands, exit codes,
and stdout hashes are in
[`documentation-flow-observation.json`](documentation-flow-observation.json);
all seven exited 0. The matching stdout files are in
[`documentation-flow/`](documentation-flow/). These are additional doc-flow
receipts and are not part of the 92 final CLI records above.

## Observed behavior

The CLI exercised explicit construction of `Not(All(...))`, `All(Not(...))`,
`exactly 0`, and Ask over one or two projected ports. Seven construction
requests and their returned documents passed the declared schema. The final
CLI record includes check, normalize, rewrite-certificate verification, export,
and replay flows for representative scopes. The evaluator results are captured
in [`391-semantic-observation.json`](391-semantic-observation.json):

- Projecting `r.a` over the supplied `r(A,B)` and `r(B,C)` facts returns `(A)`
  and `(B)`.
- Projecting `r.b` over those same facts returns `(B)` and `(C)`.
- The two-column Ask returns `(A,B)` and `(B,C)`; reversing its declared answer
  order returns `(B,A)` and `(C,B)`.
- A supported yes/no Ask returns one empty tuple. A nonmatching Ask and the same
  Ask over an empty fact set return no tuples with `knowledge: "unknown"`.
- Evaluation-proof verification accepts the captured positive witness; a proof
  for a different Ask is refused.

The facts file remained byte-identical through evaluation. These results
demonstrate explicit record support and projection order only. They do not
establish world truth, question completeness, or quantifier semantics.

Negative controls observed malformed counts, invalid scopes, type mismatches,
duplicate Ask answers, and explicit `--select` on Ask as invalid. Ask bodies
containing Not or Quantifier were refused as unsupported. Quantifiers and Not
remain representable through build/check/normalization but outside query
evaluation; `exactly 3` remains unsupported even with three explicit matching
records. Low work and answer limits returned `resource_limit` rather than a
negative answer.

The final correction removed a selection-count gate for non-Ask closed
propositions while retaining the CLI's explicit `--select` file requirement.
With an explicit empty selection, the closed `r(A,B)` query over these facts
returns supported and `r(B,A)` returns unknown. Ask rejects any externally
supplied selection because its ordered `answers` field owns projection.

The pre-correction behavior is preserved separately in
[`historical/pre-correction/`](historical/pre-correction/): three outputs show
the earlier `QUERY_SELECTION_REQUIRED` response for closed queries. This is a
historical receipt only; it is not included in the 92 final command records.
An earlier invocation also combined an inline declaration inventory with
selected metadata and was rejected under the existing inventory contract; the
successful final invocations use the corrected declaration-free normalization
request and unchanged expected semantics.

## Reproduction inputs

The neutral inputs are copied under [`inputs/`](inputs/), and a small curated
set is available as [product examples](../../../../../data/semantic/examples/scopes/).
For example, from the repository root:

```sh
go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/build/not-all.json --json

go run ./cmd/golingua algebra evaluate \
  --metadata data/semantic/examples/scopes/metadata.json \
  --expression data/semantic/examples/scopes/query/ask-first.json \
  --facts data/semantic/examples/scopes/facts-simple.json \
  --context-id eval-neutral --profile positive-facts-v1 --json
```

The JSONL receipts include exact argv and exit codes for all other inputs.
The separate coordinator check of tuples and fact immutability is recorded in
[`391-semantic-observation.json`](391-semantic-observation.json). It is an
agent assessment, not owner acceptance.

## Statuses and limits

The worker accidentally invoked `go test ./...` during implementation. It exited
nonzero; its raw output was not retained. This was outside the planned minimal
verification, did not add tests, and was not used as an acceptance gate. The
required formatting, vet, build, help/version, schema, and actual CLI checks
above are recorded separately.

Coordinator technical and semantic review are recorded as local observations.
Owner language acceptance remains `NOT_EVALUATED`. The issue is still awaiting
its PR and actual PR/main CI receipt; this evidence does not claim a merge or
owner approval.

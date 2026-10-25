# DSE-03-01 evidence index

This directory records observations for the bounded typed-IR/checker leaf. Its JSON files under [`inputs/`](inputs/) are explicit neutral controls, not a broad corpus or a pass/fail runner. A control's presence specifies an input; it does not assert that a command was run or passed.

## Control inputs

| Group | Existing input files | Independently specified purpose |
| --- | --- | --- |
| Positive term/type forms | [`atom.json`](inputs/atom.json), [`lambda.json`](inputs/lambda.json), [`apply.json`](inputs/apply.json), [`nominal-sort.json`](inputs/nominal-sort.json) | Declared relation roles, typed binders/application, and declared nominal sorts. These inputs describe representability/type checks, not reduction or truth. |
| Free and explicit values | [`parameter-open.json`](inputs/parameter-open.json), [`parameter-closed.json`](inputs/parameter-closed.json), [`unknown-reference.json`](inputs/unknown-reference.json) | Declared open parameter, refusal when a parameter occurs in a closed document, and explicit-reference declaration requirements. |
| Unresolved holes | [`hole.json`](inputs/hole.json) | Typed hole remains a requirement and yields `incomplete`; it is not converted into an `Exists` witness or fact. |
| Shape and scope refusals | [`wrong-type.json`](inputs/wrong-type.json), [`dangling-binder.json`](inputs/dangling-binder.json), [`duplicate-binder.json`](inputs/duplicate-binder.json), [`same-label-distinct-binders.json`](inputs/same-label-distinct-binders.json), [`duplicate-role.json`](inputs/duplicate-role.json), [`arrow-relation-port.json`](inputs/arrow-relation-port.json) | Exact type/role checks, lexical IDs, distinct binder identity despite repeated labels, and base-sort-only signature ports. |
| Strict wire refusals | [`wrong-version.json`](inputs/wrong-version.json), [`unknown-field.json`](inputs/unknown-field.json), [`irrelevant-null.json`](inputs/irrelevant-null.json), [`duplicate-key.json`](inputs/duplicate-key.json), [`trailing-document.json`](inputs/trailing-document.json), [`invalid-utf8.json`](inputs/invalid-utf8.json), [`unknown-symbol.json`](inputs/unknown-symbol.json), [`unknown-sort.json`](inputs/unknown-sort.json), [`invalid-limit.json`](inputs/invalid-limit.json) | Version, field union, duplicate/trailing JSON, UTF-8, explicit declarations, and limits. Each is a specified negative input; outcomes belong in the observation record. |
| Canonical round-trip | [`roundtrip.json`](inputs/roundtrip.json), [`roundtrip-complex.json`](inputs/roundtrip-complex.json) | Preserve document meaning fields through decode/encode, including richer term structure. These are not byte-identical JSON requirements. |
| Resource bounds | [`term-depth-limit.json`](inputs/term-depth-limit.json), [`raw-depth-limit.json`](inputs/raw-depth-limit.json) | Expression and pre-decode raw JSON nesting boundaries. The committed raw-depth input is below the document byte cap. |

The separate document byte-limit input is generated temporarily under `/tmp` during a local probe; the oversized blob is not committed. Do not replace it with a large repository fixture.

## Observation receipt

The coordinator executed every recorded probe through the actual CLI. The independent expectations were fixed before execution: valid types/scope remain structural only, invalid shapes/scope are refused, typed holes stay incomplete, and declared bounds stop work without a false/no-solution claim. All recorded statuses and exits match those requirements. No expected language corpus or test runner was changed.

- Literal commands, stdout/stderr, exits and input/binary hashes: [`observations.jsonl`](observations.jsonl).
- Toolchain, base SHA plus exact working-tree source fingerprints, build/smoke observations and binary hash: [`environment.json`](environment.json).
- A complex canonical replay preserves binders, labels, ordered roles, references, free parameters, holes and limits; it remains `incomplete`, not a fabricated complete proposition.
- The initial raw-depth probe exposed an oversized diagnostic path being converted into an internal-result error. The algebra-local CLI adapter now safely bounds diagnostics, marks truncated paths explicitly, and preserves `resource_limit`/exit 4. Ordinary exact paths and statuses are preserved.
- `gofmt`, `go vet ./...`, build, help/version and coordinator/read-only reviews completed. CI PR/main receipts belong to the issue after they actually run.
- Owner language acceptance: `NOT_EVALUATED`.
- DSE synthesis, fact evaluation, truth proving, and language integration: outside this leaf; do not infer their availability from a successful checker result.

The [formal RFC](../../../../architecture/semantic-algebra/formal-contract.md), [IR implementation contract](../../../../architecture/semantic-algebra/ir.md), and [CLI reference](../../../../cli.md) define what each observation means. `ok` means the requested structural check completed; it is not a truth or owner-acceptance claim.

# DSE-04-02 bounded search evidence

This evidence records the frozen local CLI observations for issue #387. It
separates the literal 50-command record, coordinator inspection of those
outputs, and the byte-cap boundary probe. The inspection is bounded structural
review; it is not an independent goal verifier, world-truth evaluation, a
language-quality score, or a quality gate.

## Source and command snapshot

- [`observations.jsonl`](observations.jsonl) preserves 50 literal records:
  `gofmt`, `go vet ./...`, `go build ./cmd/golingua`, help/version, complete
  searches, refusals, bounded stops, stdin, post-build metadata, and the
  candidate export/replay and independent alpha-verification flow. The recorded
  gofmt, vet, build, help, and version commands exited 0.
- [`environment.json`](environment.json) records the source base
  `14df22d0d663bc72ab88c6352f14507798ee5756`, the exact source hashes, inputs,
  product data, schema, pipeline, Go version, and binary. Its 529 source hashes
  have no mismatches against the recorded snapshot. The binary was
  `/tmp/golingua-dse/bin/golingua-387-final`, SHA-256
  `a561d324ee7b5047e82b6142672147a165832be0f2e647ed0d864d9170fe0604`.
  This source snapshot was uncommitted and pinned by file hashes; this is not a
  claim that it was merged or that its PR/main CI completed.
- [`invariants.json`](invariants.json) records coordinator inspection of 59
  candidate observations and 22 per-run payload/completeness checks. The checks cover topology, unary targets, distinct
  constraint-fragment occurrences, goal interface and type, draft status,
  alpha-only identity, empty artifact profiles, and candidate count/byte
  accounting. The inspection also records key and trace invariance under
  metadata insertion-order and label changes, distinct typed-hole reference
  choices, exact equality of legacy construct/catalog outputs, and equal
  exported/replayed artifacts in separate processes. These checks are scoped
  observations, not an independent goal proof.

## Search outcomes

The independent neutral inputs in [`inputs/independent/`](inputs/independent/)
include metadata and malformed controls. The recorded CLI outputs include:

- `unique`: one complete candidate; `ambiguous`: eight; `unary-choice`: four;
  and `hole-references`: two, preserving distinct declared reference IDs.
- The default-length example returned 12 candidates. A type-fit empty domain
  and a typed hole with no matching reference both returned a complete empty
  result scoped to the finite profile, not a claim that a proposition is false.
- Typed-length pruning and a disallowed operator each retained one complete
  candidate and recorded one rejected choice. Unknown constraints, roles, and
  symbols were refused rather than silently omitted.
- Candidate, expansion, work, candidate-byte, branch-depth, atom, path-length,
  and timeout bounds returned `resource_limit` with `complete: false`; retained
  prefixes were not marked unique. Trace retention omitted seven trace entries
  while preserving all eight candidates and complete search status.
- Strict flag controls refused unknown, duplicate, negative, fractional,
  overflowing, and over-cap limits, plus `--input`, `--no-cache`, and search
  limits in construct mode.

Work units are bounded logical accounting, not CPU time or energy. Timeout
stops are observed at cooperative phase and iteration checkpoints, not as a
hard real-time guarantee.

The post-build inputs in [`inputs/post-build/`](inputs/post-build/) were created
only after the binary was built. A new nominal sort, symbols, and roles still
produced eight candidates without rebuilding; the before/after binary hashes
match. The recorded candidate pipeline performs metadata-bound checking,
export, and replay in a separate process. Independent alpha verification
succeeds for the candidate certificate and reference context; a wrong alpha
snapshot is refused. This validates the alpha-identity receipt only. Each
candidate's `draft_goal_unverified` certificate keeps `goal_verified: false`;
there is no independent goal proof, fact/world-truth claim, or language route.

The exact pipeline inputs and outputs are in
[`inputs/pipeline/`](inputs/pipeline/). Candidate choices retain their
symbols, orientations, references, and fragment mappings. Metadata order or
labels do not change candidate keys or traces; no sense or relation is merged
by type. The explicit artifact profile is empty and is separate from both the
search goal profile and the alpha-only identity profile.

## Whole-response byte-cap probe

The separate [`inputs/wire-boundary/`](inputs/wire-boundary/) record contains
the 146,238-byte goal with 500 declared neutral references, the compact probe
summary, and the before/after binary hash. The frozen CLI exited 4 and returned
valid JSON of 16,380,520 bytes including its newline, under the 16,777,216-byte
whole-response limit. It retained 36 candidates and discarded one suffix
candidate, reporting `resource_limit`, `incomplete`, `complete: false`,
`unique_within_declared_space: false`, and `stop_reason: max_response_bytes`.
`usage.candidate_bytes` was 16,226,243, exactly the sum of the retained
serialized candidate payloads. This demonstrates the separate whole-response
cap; it is a byte-limit diagnostic, not a language corpus result or score.

The 16,380,520-byte raw stdout remains outside the repository at
`/tmp/golingua-dse/387-wire-probe-real/output-500-256.json`, SHA-256
`df4c90cc02c7280a9d81939897f4ab495cfc5e85b33b6f3b0abdfaeca144c845`. It is
intentionally not duplicated into repository evidence; the compact summary
records the raw path, hash, and byte count. The product [search examples](../../../../../data/semantic/examples/search/README.md)
are small neutral formal inputs, not language cases or owner-approved meanings.

## Scope and status

The implementation enumerates a finite declared search space and preserves
alternatives. Complete empty results, alpha identity, construction provenance,
and byte-bound behavior have only the scopes described above. Search does not
evaluate facts, prove world truth, choose speaker intent, or create persistent
learning or a language route. The task adds no fact store.

Technical command observations and coordinator inspection are separate from
owner acceptance. Owner language acceptance remains **NOT_EVALUATED**. The
issue-map entry remains `ACTIVE`; this evidence snapshot has no PR or merge
receipt, and PR/main CI status awaits the canonical completion receipt.

The coordinator clarified the documentation export example as construct-only and added explicit search candidate-index extraction after review. The separate [documentation-flow receipt](doc-flow-observation.json) records that exact jq extraction (index 0 from the unique completed example), export, replay-request assembly and replay with the frozen binary; all four commands exited 0. This is separate from the 50-command main log.

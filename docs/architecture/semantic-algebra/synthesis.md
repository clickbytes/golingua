# Finite goal construction and search

`algebra synthesize --mode construct` turns a fully specified goal into a
checked, typed expression using a small explicit operator set.
`algebra synthesize --mode search` enumerates constructions over a finite set
of declared choices. A goal is a request, not an AST, recipe, fact query, or
claim that the requested relation holds in a world. Both modes use the bounded
`positive-facts-v1` profile.

## Command and inputs

```sh
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/goals/metadata.json \
  --goal data/semantic/examples/goals/two-hop-property.json --json
```

Repeat `--metadata <file>` to select explicit packages and their declared
dependencies. `--goal` is required and names one JSON file; `--goal -` reads
stdin. `--json` and an explicit mode are required. Both modes reject `--input`
and `--no-cache`. Metadata is loaded through the validated registry path, and
only approved primitive signatures are available. No package path is read from
the goal itself.

The strict version-1 goal document has `schema_version`, `profile`,
`context_id`, `expected_type`, `ports`, `references`, `allowed_operators`,
`constraints`, `provenance`, and `limits`; optional `project_ports` names ports
to close. The arrays and `limits` object are required; the goal contains no
expression or inline sort/signature declarations.
Reference declarations have exactly `id` and `type`. `context_id` is a required,
bounded request identifier; it does not authenticate a caller or identify an
entity. Types and declared references are checked against the selected
metadata. Source references preserve attribution but do not establish owner
approval.

Each goal port is either open or explicitly bound to a typed `Ref` or `Hole`.
`expected_type` must be the arrow type of the open, nonprojected ports in their
declared order, returning `Prop`. Binding a reference or projecting an open
port removes that port from the output interface; a Ref/Hole-bound port cannot
also be projected. Projection existentially closes only the named goal ports.
A typed hole remains unresolved: construction returns `incomplete`
before lowering and does not treat the hole as an operator argument or witness.

The initial profile supports one binary path constraint and unary
restrictions on its endpoints. A path names its endpoint ports and relation
roles; each relation occurrence and each unary restriction consumes one atom
from the shared atom budget. Path counts are explicit choices. Missing choices,
missing path symbol or roles, or a missing unary symbol/port remain unresolved;
multiple permitted path counts remain underdetermined and are never resolved by
choosing the first. An unknown nonempty symbol, unknown role, conflicting role
mapping, unsupported graph shape, or invalid type is a refusal, even when
another field is unresolved. Labels and identifier spellings do not select
operators or infer roles.

## Construction and output

The allowed public operators are `relation`, `join`, `restrict`, `bind`,
`project`, and `permute`. The constructor calls the shared algebra operators with explicit
ports and role mappings. It checks the completed expression against the
selected declarations and the requested type. Construction does not perform
fact lookup, graph search, general planning, beta/alias expansion, or language
realization.

The JSON result includes `status`, `mode`, `goal_profile`, `context_id`,
`metadata`, `result`, and `validation_scope`. A complete result contains the
checked expression, interface, canonical document, operator trace, and
obligations mapping each input constraint to distinct final-atom paths and
construction step IDs. An obligation status of `constructed_provenance` records
which construction steps supplied that fragment; it is not an independent
proof that the goal is true. Incomplete results do not contain a candidate
expression or export request. Exit codes follow the algebra status mapping:
`ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5,
`ambiguous` 6, `incomplete` 7, and infrastructure/error 1.

Only a complete construction includes `export_request`. It records the
canonical document, goal context ID and provenance, empty context dependencies,
and an explicit empty rewrite profile named `dse-construct-artifact-v1`. This
profile is separate from the goal's `positive-facts-v1` profile. Before the
request is included, the CLI creates a `constructed_unverified` certificate
and independently checks it against the original goal, candidate document,
and explicitly selected metadata. The certificate's `goal_verified` remains
`false`: it is the producer's untrusted claim record, while the separate
`GoalVerificationResult` reports the component checks. `fully_validated` is a
field on the synthesis result and is true only when all required checks pass;
only then does it include an export request. Export stores the checked
representation; it does not normalize it or evaluate facts. Replay still
requires independently selected metadata, context, and profile.

## Independent goal verification

Construction checks types and operator plans. The independent verifier then
rechecks the candidate document, constraint coverage, operator trace, scope,
and certificate claims without calling the constructor or search engine. Use
the canonical document from the construction output as the candidate input
and the accompanying producer certificate as an untrusted claim:

```sh
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/verification/metadata.json \
  --goal data/semantic/examples/verification/goal.json --json > /tmp/synthesis.json
jq -e '.result.fully_validated == true' /tmp/synthesis.json >/dev/null
jq -c '.result.canonical_document' /tmp/synthesis.json > /tmp/candidate.json
jq -c '.result.certificate' /tmp/synthesis.json > /tmp/certificate.json
go run ./cmd/golingua algebra verify-goal \
  --goal data/semantic/examples/verification/goal.json \
  --candidate /tmp/candidate.json --certificate /tmp/certificate.json \
  --metadata data/semantic/examples/verification/metadata.json --json > /tmp/verification.json
jq -e '.result.status == "ok" and .result.type_correct and .result.refinement_checked and .result.replay_checked and .result.certificate_valid and .result.goal_satisfied and .result.complete' /tmp/verification.json >/dev/null
jq -c '.export_request' /tmp/synthesis.json > /tmp/export-request.json
go run ./cmd/golingua algebra export \
  --metadata data/semantic/examples/verification/metadata.json \
  --input /tmp/export-request.json --json > /tmp/export.json
jq -c --slurpfile exported /tmp/export.json \
  '{schema_version:.schema_version,profile:.profile,context_id:.context_id,context_dependencies:.context_dependencies,provenance:.provenance,artifact:$exported[0].artifact}' \
  /tmp/export-request.json > /tmp/replay-request.json
go run ./cmd/golingua algebra replay \
  --metadata data/semantic/examples/verification/metadata.json \
  --input /tmp/replay-request.json --json
```

`verify-goal` requires named goal, candidate, and certificate files; it does
not read these three inputs from stdin. Repeat `--metadata` for explicitly
selected packages. The candidate is a self-contained algebra document whose
inline declarations must agree with the selected packages. The certificate
cannot select packages or establish its own snapshot. Optional repeatable
`--verify-limit name=value` accepts only `max_certificate_bytes`,
`max_replay_steps`, and `max_work_nodes`; duplicate, malformed, unknown, or
over-cap values are refused. Defaults and caps are 8 MiB for the certificate,
64 replay steps, and 262,144 logical work nodes. The certificate size is
checked before it is read into the verifier.

The result keeps `type_correct`, `refinement_checked`, `replay_checked`,
`certificate_valid`, `goal_satisfied`, and `complete` separate. A usable
verified result requires all of these checks to finish successfully. A
producer's `goal_verified` field is never trusted. `origin` is likewise a
producer-supplied label, not proof that finite search ran. One verified
candidate does not establish search completeness or uniqueness. A `Hole`,
unsupported claim, or resource stop cannot become a validated candidate.
Reported assumptions are limited to the selected typed registry snapshot,
the request's explicit reference boundary, and the positive structural
profile; they are not domain axioms or claims about named real-world entities.
Goal certificates do not accept `identity_proof` or `identity_key` fields.
Search's optional alpha-normalization check is a separate, narrower identity
step used for deduplication; it is not part of goal verification.

This verification is bounded to the selected positive structural profile. It
checks that typed terms match the requested finite goal and its explicit
constraints. It does not evaluate stored facts, establish world truth, prove
global theory consistency, or imply owner language acceptance. The neutral
[verification examples](../../../data/semantic/examples/verification/README.md)
include one accepted candidate and a wrong-orientation candidate that is
refused.

## Finite search

Search uses the same goal and metadata inputs, with an explicit search mode:

```sh
go run ./cmd/golingua algebra synthesize --mode search \
  --metadata data/semantic/examples/search/metadata.json \
  --goal data/semantic/examples/search/ambiguous.json --json
```

Repeatable `--search-limit name=value` flags are accepted only in search mode.
The exact names, defaults, and hard caps are:

| Limit | Default | Hard cap |
| --- | ---: | ---: |
| `max_path_length` | 3 | 16 |
| `max_branch_depth` | 64 | 64 |
| `max_expansions` | 128 | 4096 |
| `max_candidates` | 16 | 128 |
| `max_trace_entries` | 64 | 256 |
| `max_work_units` | 1,048,576 | 16,777,216 |
| `max_candidate_bytes` | 8 MiB | 16 MiB |
| `timeout_millis` | 2000 | 30000 |

Zero selects the finite default. Negative, over-cap, duplicate, malformed, or
unknown limits are refused. The search space varies the declared relation and
role choices, unary target choices, path length, and typed hole bindings to
matching declared references. Choices are kept in deterministic identifier
order, including every legal role orientation and reference ID; the search
does not pick the first match or infer speaker intent from a type. Empty path
counts enumerate lengths from one through the configured `max_path_length`.
An explicitly requested length outside that bound stops with a partial
`resource_limit` result instead of silently dropping that branch. Each search
axis contributes one branch-depth level.

Candidate construction still uses the goal's `max_atoms` (16 total path and
unary atoms), `max_steps` (64 operator calls), and shared type and declaration
limits. Candidate-specific type mismatch and disallowed-operator choices can
be rejected while the remaining space is searched. Unknown symbols, roles,
malformed constraints, and invalid goal types remain request errors. A typed
hole can bind only to an explicitly declared, type-matching reference. A hole
with no matching reference produces a complete empty result within this finite
profile; no new reference or witness is invented.

The result preserves the original goal and reports the declared space,
candidate choices, per-candidate constructions, independent goal-verification
results, trace, rejections, and usage. Only fully verified candidates are
retained with an `export_request`. `complete: true` means only that this finite
declared space was exhausted.
`no_solution_within_profile_bounds` with status `ok` means no candidate was
found in that space, not that a proposition is false in the world. Any
operational stop or unresolved goal keeps `complete` false and cannot claim
uniqueness, even when a valid prefix candidate is retained. Trace omission is
reported separately and does not itself make the space incomplete. There is no
ranking or confidence score.

Candidates are deduplicated only after licensed `alpha_v1` normalization and
independent verification of that normalization certificate. The key compares
the full bounded semantic descriptor and normalized term, preserving chosen
symbols, role orientations, reference IDs, and constraint-fragment mappings.
No beta reduction, alias expansion, or conjunction equivalence is added to the
search identity profile. Each candidate has a `draft_goal_unverified`
certificate with `goal_verified: false`, plus a separate independently
verified goal certificate and result. The draft certificate and alpha identity
are not a goal proof, fact check, world-truth claim, or language approval.

Every candidate includes its own `export_request` with the canonical
construction document, context, provenance, and the explicit empty
`dse-search-artifact-v1` rewrite profile. The search result does not select a
single candidate for export. Inspect alternatives first, then deliberately
extract the chosen candidate's request for the existing export/replay flow.
Export and replay still require explicit metadata and context; these examples
do not add a fact store, named derived definitions, or language routes.

## Limits and boundaries

The goal-level hard limits are 16 total path and unary atoms, 64 public
operator calls, 16 goal ports, 16 source references, 32 constraints, and 1 MiB
of goal JSON. Shared type checking additionally bounds expressions to 4,096 nodes and
depth 64, types to depth 32, declarations to 512, and relation ports to 16.
Goal limits can be lowered; a resource limit is not evidence that no
construction exists. `max_candidate_bytes` (8 MiB default, 16 MiB hard cap)
counts retained serialized candidate payloads only. It excludes the base goal,
space descriptor, and trace; `usage.candidate_bytes` is the sum of retained
candidate payload sizes. The CLI separately caps the complete encoded response
at 16 MiB. If that cap requires dropping a suffix of candidates, it reports a
resource-limited incomplete result and the number discarded. A historical
#387 CLI probe, run before goal-certificate fields changed candidate payload
sizes, used 500 declared neutral references with 256-byte IDs in a 146,238-byte
goal. It exited 4 with a 16,380,520-byte response including the newline, under
the 16,777,216-byte cap. The result retained 36 candidates and discarded one;
it reported `resource_limit`, `incomplete`, `complete: false`,
`unique_within_declared_space: false`, and `stop_reason: max_response_bytes`.
`usage.candidate_bytes` was 16,226,243, exactly the serialized size sum of the
retained candidates. The raw response is outside the repository at
`/tmp/golingua-dse/387-wire-probe-real/output-500-256.json` (SHA-256
`df4c90cc02c7280a9d81939897f4ab495cfc5e85b33b6f3b0abdfaeca144c845`); the
compact command and result record is
`/tmp/golingua-dse/387-wire-probe-real/summary.json`. The frozen binary hash
was unchanged before and after
(`a561d324ee7b5047e82b6142672147a165832be0f2e647ed0d864d9170fe0604`). This is
a byte-limit diagnostic, not a language corpus result, world-truth evidence,
or owner approval. Work units are logical accounting, not CPU time or energy.
Timeout checks are cooperative at bounded phase and iteration checkpoints,
not a hard real-time guarantee.

The initial profile is not a general graph synthesizer, theorem prover, fact
evaluator, equivalence checker, or natural-language route. It does not infer
negative facts or closed-world truth. A checked expression establishes only
structural well-typedness under the selected registry. Independent goal proof,
world truth, and owner language acceptance are separate; owner language
acceptance remains **NOT_EVALUATED**.

## Examples and continuation

The [goal examples](../../../data/semantic/examples/goals/README.md) include a
two-hop path with a unary property, a three-hop path, a one-hop minimal case,
an explicit reference binding, and typed-hole/underspecified requests. The
[goal schema](../../../schemas/semantic/goals.schema.json) defines the closed
wire shape. The [search examples](../../../data/semantic/examples/search/README.md)
include unique, ambiguous, underdetermined, typed-hole, and candidate-pruning
inputs with neutral formal metadata. They are bounded inputs, not a
gold-output corpus or an automatic quality gate.

For a compact **construct-mode** export/replay flow, save the synthesis result
and extract its canonical document and top-level export request:

```sh
jq -c '.result.canonical_document' /tmp/synthesis.json > /tmp/expression.json
go run ./cmd/golingua algebra check --input /tmp/expression.json --json
jq -c '.export_request' /tmp/synthesis.json > /tmp/export-request.json
go run ./cmd/golingua algebra export --metadata data/semantic/examples/goals/metadata.json --input /tmp/export-request.json --json > /tmp/export.json
jq -c --slurpfile exported /tmp/export.json \
  '{schema_version:1,profile:.export_request.profile,context_id:.export_request.context_id,context_dependencies:.export_request.context_dependencies,provenance:.export_request.provenance,artifact:$exported[0].artifact}' \
  /tmp/synthesis.json > /tmp/replay-request.json
go run ./cmd/golingua algebra replay --metadata data/semantic/examples/goals/metadata.json --input /tmp/replay-request.json --json
```

Search results use a candidate-indexed shape. Inspect the alternatives and
explicitly choose an index before extraction. This example uses a goal with
one candidate in its completed declared space:

```sh
go run ./cmd/golingua algebra synthesize --mode search --metadata data/semantic/examples/search/metadata.json --goal data/semantic/examples/search/unique.json --json > /tmp/search.json
jq '.result.candidates | to_entries[] | {index:.key,id:.value.id,choices:.value.choices}' /tmp/search.json
selected_candidate=0
jq -c --argjson chosen "$selected_candidate" '.result.candidates[$chosen].export_request' /tmp/search.json > /tmp/search-export-request.json
go run ./cmd/golingua algebra export --metadata data/semantic/examples/search/metadata.json --input /tmp/search-export-request.json --json > /tmp/search-export.json
jq -c --slurpfile exported /tmp/search-export.json \
  '{schema_version:1,profile:.profile,context_id:.context_id,context_dependencies:.context_dependencies,provenance:.provenance,artifact:$exported[0].artifact}' \
  /tmp/search-export-request.json > /tmp/search-replay-request.json
go run ./cmd/golingua algebra replay --metadata data/semantic/examples/search/metadata.json --input /tmp/search-replay-request.json --json
```

The canonical document includes hydrated declarations. Structural `algebra
check` can check that document without metadata. If checking with metadata,
remove only the inline `sorts` and `signatures` first; metadata-bound checking
rejects inline declarations. The export request should retain the canonical
document as produced.

The implementation and observed CLI record belong to the active DSE issue and
its evidence; prior observations predating the final bounded validation change
are not evidence for that final source snapshot. Implementation status,
agent assessment, mathematical review, and owner acceptance are separate
records. Owner language acceptance remains **NOT_EVALUATED**. See the
[DSE issue map](../../work/dse/issue-map.json) for current status.

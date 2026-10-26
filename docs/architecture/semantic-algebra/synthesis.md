# Finite goal construction

`algebra synthesize --mode construct` turns a fully specified goal into a
checked, typed expression using a small explicit operator set. The goal is a
request for construction, not an AST, recipe, fact query, or claim that the
requested relation holds in a world. This page describes the current bounded
`positive-facts-v1` profile.

## Command and inputs

```sh
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/goals/metadata.json \
  --goal data/semantic/examples/goals/two-hop-property.json --json
```

Repeat `--metadata <file>` to select explicit packages and their declared
dependencies. `--goal` is required and names one JSON file; `--goal -` reads
stdin. `--json` and `--mode construct` are required. This command does not
accept `--input`, `--no-cache`, or a search mode. Metadata is loaded through the
validated registry path, and only approved primitive signatures are available
to construction. No package path is read from the goal itself.

The strict version-1 goal document has `schema_version`, `profile`,
`context_id`, `expected_type`, `ports`, `references`, `allowed_operators`,
`constraints`, `provenance`, and `limits`. The arrays and `limits` object are
required; the goal contains no expression or inline sort/signature declarations.
Reference declarations have exactly `id` and `type`. `context_id` is a required,
bounded request identifier; it does not authenticate a caller or identify an
entity. Types and declared references are checked against the selected
metadata. Source references preserve attribution but do not establish owner
approval.

Each goal port is either open or explicitly bound to a typed `Ref` or `Hole`.
`expected_type` must be the arrow type of the open ports in their declared order,
returning `Prop`. Binding a reference removes that port from the output
interface. A typed hole remains unresolved: construction returns `incomplete`
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

The allowed public operators are `relation`, `join`, `restrict`, `bind`, and
`permute`. The constructor calls the shared algebra operators with explicit
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
profile is separate from the goal's `positive-facts-v1` profile. Export stores
the checked representation; it does not normalize it or certify goal
satisfaction. Replay later requires independently selected metadata, context,
and profile.

## Limits and boundaries

The initial hard limits are 16 total path and unary atoms, 64 public operator
calls, 16 goal ports, 16 source references, 32 constraints, and 1 MiB of goal
JSON. Shared type checking additionally bounds expressions to 4,096 nodes and
depth 64, types to depth 32, declarations to 512, and relation ports to 16.
Limits can be lowered; a resource limit is not evidence that no construction
exists.

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
wire shape. They are bounded inputs, not a gold-output corpus or an automatic
quality gate.

For a compact export/replay flow, save the synthesis result and extract the
canonical document and export request:

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

The canonical document includes hydrated declarations. Structural `algebra
check` can check that document without metadata. If checking with metadata,
remove only the inline `sorts` and `signatures` first; metadata-bound checking
rejects inline declarations. The export request should retain the canonical
document as produced.

The implementation and observed CLI record belong to the active DSE issue and
its evidence; prior observations predating the final bounded validation change
are not evidence for that final source snapshot. Implementation status,
agent assessment, mathematical review, and owner acceptance are separate
records. See the [DSE issue map](../../work/dse/issue-map.json) for current
status.

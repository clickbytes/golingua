# DSE-02-02 #384: evidence record

This directory preserves the literal local observations for #384 and the exact input requests used to obtain them. The JSONL log has 55 records; it is evidence of command behavior, not a gold set, test runner, truth oracle, or owner acceptance. Product examples are separately available in [`data/semantic/examples/facts`](../../../../../data/semantic/examples/facts/).

## Frozen run

The [environment record](environment.json) identifies the executable, source base and per-file source hashes. It records the source as uncommitted but pinned by those hashes; it must not be described as a committed build. The executable was `/tmp/golingua-dse/bin/golingua-384-final`, SHA-256 `ee1d04f6abf51d53e19e61542ee5508334c0e338f5769416663468d536df616e`, built from base `a7415414b3c19ac2aa2c1ca1d8fb7fa57186e146` with Go 1.27.0 on linux/amd64. The recorded `gofmt`, `go vet ./...`, build, help, and version commands exited 0. The source hash inventory is in the environment file; these records were copied without rebuilding the frozen binary.

The [observations](observations.jsonl) contain literal argv, exit codes, stdout, and stderr for each command. The [view API observation](view-api-observation.json) records a separate isolated Go probe (`GOPROXY=off`) for absence, explicit time selection, inclusive boundaries, conflicts, year-one instants, and mutation isolation.

## Observed behavior

- `facts-arities` accepted unary, binary, and ternary records; `facts-zero-arity` accepted a zero-argument atom. `facts-role-swap` kept role/value assignments distinct, and `facts-distinct-events` retained separate event references. `facts-same-sort-instances` kept same-sort referents distinct.
- `facts-conflict` and the valid stdin request returned `ok` with two active opposite-polarity records, evidence state `both`, and one reported conflict. `facts-negative-only` returned evidence state `negative`. An empty view was valid with zero records.
- The explicit UTC endpoints `time-start` and `time-end` were inclusive and reported `both`. `time-before` and `time-after` had zero active records and reported `unknown`; `time-no-at` did not activate bounded records; `time-disjoint` had one active and one inactive record and no active conflict. The API probe independently reports inclusive endpoints and `unknown` for an empty lookup at a nonempty document time. Absence is unknown evidence, not falsehood.
- Malformed requests were refused with typed `invalid` statuses for duplicate IDs/roles, missing or extra roles, unknown symbols/referents, wrong sorts, variables/holes, mismatched snapshot pins, bad source provenance, invalid/non-UTC instants, reversed intervals, null limits, unknown fields, duplicate JSON keys, nonasserted scopes, and absent metadata/invalid subcommands. `fact-limit` returned `resource_limit` (exit 4). These are parser/validation observations, not claims about world truth.
- No bridge from this request-scoped view to the existing binary `Fact`/`Set` or `question` runtime was exercised; that mapping remains unavailable. Owner language acceptance is **NOT_EVALUATED**.

## Inputs and scope

[`inputs/independent`](inputs/independent/) preserves the prepared independent request directory, including the 14 positive/valid neutral requests and its explicit refusal controls. [`inputs/extra`](inputs/extra/) preserves the separate small diagnostic requests used for malformed-input, time, provenance, and pin observations; neither directory is a generated corpus or runner. The shared neutral metadata is [`data/semantic/examples/facts/metadata.json`](../../../../../data/semantic/examples/facts/metadata.json). No input creates facts from a sentence or Meaning graph.

The records document a local branch observation only. The #384 PR and main CI canonical receipt are pending; no PR URL, merge SHA, or CI result is asserted here. The active issue map points #384 to this evidence directory. The coordinator's assessment remains separate from owner acceptance.

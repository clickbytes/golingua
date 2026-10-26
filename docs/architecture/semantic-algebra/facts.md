# Request-scoped n-ary evidence

**Status:** implemented as a bounded `algebra facts validate` path. It accepts explicitly supplied ground evidence and summarizes which records apply at an explicit query instant. This is an evidence inventory, not a truth evaluator. Coordinator assessment and owner acceptance are separate; owner language acceptance remains `NOT_EVALUATED`.

## CLI request

`algebra facts validate` requires one or more explicit metadata packages and a version-1 JSON request. `--input <file>` reads one file; without `--input` or with `--input -`, it reads stdin. `--json` is required. Repeat `--metadata <file>` for explicitly selected packages and required dependencies; paths or URLs found inside JSON are never opened.

```sh
./golingua algebra facts validate \
  --metadata data/semantic/examples/facts/metadata.json \
  --input data/semantic/examples/facts/arities.json --json
./golingua algebra facts validate \
  --metadata data/semantic/examples/facts/metadata.json \
  --input data/semantic/examples/facts/conflict.json --json
./golingua algebra facts validate \
  --metadata data/semantic/examples/facts/metadata.json \
  --input data/semantic/examples/facts/time-start.json --json
```

`facts.schema.json` gives the strict wire shape and the [neutral examples](../../../data/semantic/examples/facts/). Runtime validation remains authoritative for registry membership, role/type matching, unique IDs, source validation, and interval ordering.

The request has `schema_version`, a required `context_id`, required `referents` and `facts` arrays, optional `snapshot_id`, optional query instant `at`, and optional request limits. Referents declare exact request-local IDs and algebra types; runtime accepts only ground base sorts `Entity`, `Event`, `Interval`, and declared `Nominal` sorts. A referent's type is not a ConceptID and does not identify it with another referent of the same type.

Each fact carries an ID, one `atom` (`symbol` and named role/value pairs), explicit `positive` or `negative` polarity, a `source`, `assertion_scope: "asserted"`, and optional closed `valid_time`. Every value has a kind and ID. Only `Ref` values are ground; `Var`, `Param`, and `Hole` are rejected as fact values. Only the selected metadata's approved primitive relation signatures are available. Pending, opaque, and alias symbols are not operational facts relations. Arity 1, 2, and 3 use the same named-role validation; arguments are returned in declared signature order while preserving each role/value association. Values are never sorted to make roles fit.

The `snapshot_id` is optional as an expected pin. When supplied, it must match the selected schema snapshot, which is computed from the explicitly selected registry IDs, versions, resource fingerprints, and formal denotation IDs. `context_id` names this request's evidence namespace; it is not authentication or an entity-identity claim. No CORE ConceptID/participant mapping or network/source loading is performed.

## Evidence and time semantics

Positive and negative records are explicit evidence. Both may coexist for the same exact atom and remain available with their IDs, source records, and validity intervals. The view reports such active opposition as conflict data; it does not reject the request, perform explosion, or assign a world truth value. Its `positive`, `negative`, `both`, and `unknown` labels describe supplied evidence states, not a complete four-valued logic. No evidence means `unknown`, never `false`.

The query instant is the optional UTC RFC3339Nano `at` value. CLI passes this value explicitly to `SummaryAt`; it does not use the current clock or infer a relative past/present. If `at` is omitted or empty, only timeless records (both validity endpoints empty) are active. With an explicit instant, empty endpoints are unbounded; specified interval endpoints are closed and inclusive. Stored facts remain in the returned `facts` list even when inactive; summary counts and evidence states reflect the chosen instant. Opposing records in disjoint windows do not create an active conflict.

Atoms are grouped by selected symbol, ordered declared roles, and exact `Ref` IDs. Swapping values between same-typed roles changes the atom. Distinct same-sort referents and distinct event IDs remain separate even if their participants match. Type equality is not entity identity.

## Output and limits

For a valid view, the command returns engine `status: "ok"` (exit 0), including when the summary contains conflicts. Its JSON fields are:

- `snapshot_id`, `context_id`, and `at` for the selected request scope;
- `referents` and canonical-ordered `facts`, retaining source, polarity, assertion scope, and validity data;
- `summary`, containing `stored_facts`, `active_facts`, `inactive_facts`, first-seen `atoms` with their evidence state, and active `conflicts`;
- `metadata` with selected package IDs, versions, and resource fingerprints;
- `absence_means: "unknown"` and `validation_scope: "explicit n-ary ground evidence; not world-truth or owner approval"`.

An empty facts list is a valid empty view with zero counts and empty arrays. Invalid input or grounding/metadata mismatch returns a typed refusal; exceeding limits returns `resource_limit`. Standard exits are `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, and infrastructure failure 1.

The JSON request limit is 1 MiB and nesting is capped at 256. Limits are 4096 facts, 512 referents, 512 aggregate schema declarations/referents, and 16 ports per relation. IDs and UTC strings are bounded by the runtime; legacy fact/source IDs retain their 128-byte restriction, while referent, symbol, and role IDs use the bounded 256-byte form. Strict decoding rejects unknown/null fields, duplicate keys, invalid UTF-8, and malformed nested types.

This new request-scoped n-ary view is separate from the existing binary `facts.Fact`/`Set` boundary. It does not import facts from `Meaning`, a parsed sentence, or a definition; it does not map into the existing `question`/reasoner route. No global facts state, persistent database, or learning behavior is added. The current command validates and summarizes evidence only; it does not provide a natural-language fact mapping or a new evaluator capability.

The neutral metadata and requests in this directory do not claim linguistic validity or owner approval. CLI observations are distinct from agent assessment; owner language acceptance remains `NOT_EVALUATED`.

# Stored typed artifacts

**Status:** implemented as a bounded artifact export, inspection, and replay path. This stores one checked syntax root; it does not evaluate facts, execute rewrites, normalize, or prove a proposition. Coordinator assessment and independent technical review are separate from owner acceptance, which remains `NOT_EVALUATED`.

## Commands and handoff

The three commands are `algebra export`, `algebra inspect`, and `algebra replay`. Each requires `--json`. Export and replay require one or more explicit `--metadata <file>` flags; inspection deliberately takes no metadata. Input comes from one `--input <file>`, or stdin when omitted or `--input -`. File input is not combined with stdin. Limits are 1 MiB for export requests, 8 MiB for an artifact envelope, and 16 MiB for replay requests. Metadata selection is bounded to 16 package files, each at most 1 MiB.

The example metadata and requests in [`data/semantic/examples/storage/`](../../../data/semantic/examples/storage/) are neutral structural examples for issue #383. `shared-refs.json`, `distinct-refs.json`, `distinct-events.json`, and `namespace-identities.json` exercise identity distinctions; the other requests cover alias-shaped expression input, zero-arity atoms, typed holes, and labels. They are requests, not expected-output fixtures or language claims.

### Build, export, replay in a separate process

First build the existing two-hop recipe and retain its canonical document. Then put that document into an export request carrying the chosen profile, request context, context dependency pins, and provenance. The extraction below intentionally uses a small Python command-line transform rather than a product runner:

```sh
go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/composition/metadata.json \
  --input data/semantic/examples/composition/two-hop.json --json > /tmp/storage-build.json
python3 -c 'import json; b=json.load(open("/tmp/storage-build.json")); r=json.load(open("data/semantic/examples/storage/shared-refs.json")); r["document"]=b["canonical_document"]; json.dump(r,open("/tmp/storage-export-request.json","w"))'
go run ./cmd/golingua algebra export \
  --metadata data/semantic/examples/storage/metadata.json \
  --input /tmp/storage-export-request.json --json > /tmp/storage-export.json
```

The metadata selected for export must contain the build document's inline sorts and signatures as an exact typed subset. The exporter compares that inventory against the selected validated registry and then uses the registry inventory. It does not strip, override, or trust build-supplied declarations. The normal metadata-selected `algebra check` route rejects inline sorts/signatures; export has this explicit comparison because it accepts a checked build result as input.

For replay, make a new request in a separate process. Its expected profile/context/dependencies/provenance are carried outside the artifact, rather than accepted from the artifact as authority:

```sh
python3 -c 'import json; e=json.load(open("/tmp/storage-export.json"))["artifact"]; r=json.load(open("/tmp/storage-export-request.json")); json.dump({"schema_version":1,"profile":r["profile"],"context_id":r["context_id"],"context_dependencies":r["context_dependencies"],"provenance":r["provenance"],"artifact":e},open("/tmp/storage-replay-request.json","w"))'
go run ./cmd/golingua algebra replay \
  --metadata data/semantic/examples/storage/metadata.json \
  --input /tmp/storage-replay-request.json --json > /tmp/storage-replay.json
python3 -c 'import json; json.dump(json.load(open("/tmp/storage-export.json"))["artifact"],open("/tmp/storage-artifact.json","w"))'
go run ./cmd/golingua algebra inspect --input /tmp/storage-artifact.json --json
```

`--no-cache` is an export-only flag. It disables node interning and request-local whole-artifact cache use; it does not change the checked expression or meaning fingerprint. Replay always validates the supplied artifact against the independently selected metadata and expected request fields. Inspection reads only the bare envelope, checks its declared structure and receipts, and returns `metadata_verified: false`; it cannot establish registry approval or authenticate external resources.

The command JSON uses these outer fields:

| Command | Result fields |
| --- | --- |
| `export` | `status`, `artifact`, `checked`, `unique_nodes`, `expanded_nodes`, `interning_enabled`, `cache_hit`, `metadata` |
| `inspect` | `status`, `validation_scope`, `metadata_verified`, `normalizer_version`, `registry_snapshot_id`, `required_registry_pins`, `profile`, `context_id`, `context_dependencies`, `provenance`, `resource_fingerprint`, `meaning_fingerprint`, `cache_key`, `root`, `unique_nodes`, `expanded_nodes`, `checked` |
| `replay` | `status`, `artifact`, `checked`, `unique_nodes`, `expanded_nodes`, `metadata` |

Observed command outcomes and environment are recorded by the coordinator in the DSE-03-03 evidence record; this document does not substitute for those receipts.

## Identity and artifact boundaries

An artifact is version 1 and contains exactly one checked root. Node IDs are dense, artifact-local syntax indices, not entity identities. Binder, parameter, reference, and hole identities are their exact payload IDs in their respective namespaces. Repeated occurrences of a `Var` may share one node; distinct event binders and distinct references remain distinct. There is no multi-root identity schema.

`context_id` names the request context. Context dependency IDs and fingerprints are explicit pins. They are not authenticated entity identities, URLs, file paths, or instructions to load code. Registry pins and source provenance describe explicitly selected resources; neither review status nor provenance asserts owner linguistic approval.

The public `resource_fingerprint` is calculated from the canonical public envelope body with the three receipt fields removed. Private registry-structure bytes are not part of it. `meaning_fingerprint` is a conservative typed-root and selected-metadata projection that includes the normalizer version, profile, context, declarations, and exact semantic IDs while ignoring labels and local node numbers. The `cache_key` uses its own versioned domain. These hashes support reproducibility and lookup; they are not proof of logical equivalence, signatures, authenticity, or truth.

The artifact retains labels and resource detail even though labels are omitted from the meaning key. Cache collision checks compare full semantic, source-resource, and available private registry-structure descriptors, not only digests. The optional cache is request-local, bounded to 16 entries and 16 MiB of serialized collision descriptors, and is neither a persistent database nor a global memory or learning system. If the complete registry structure is unavailable, whole-artifact cache lookup is bypassed while artifact export remains valid. `--no-cache` can change DAG sharing and node counts, but not checked IR or meaning fingerprint. No CPU, energy, or performance claim follows from this storage path.

Replay recomputes the three receipts, verifies graph density, edges, acyclicity, reachability, and expanded-node limits, and checks against the explicit selected metadata. Effective limits are the stricter of artifact and request limits; replay preserves original artifact node IDs and limits. A typed `Hole` remains incomplete and is never a witness. New normalizer versions, metadata snapshots, profiles, or contexts affect cache identity and mismatched replay is refused.

The format and request fragments are in [`expression.schema.json`](../../../schemas/semantic/expression.schema.json), under `$defs` (`artifact_envelope`, `artifact_export_request`, and `artifact_replay_request`); the existing root expression schema is unchanged. The shared checker bounds are 4096 expanded nodes, depth 64, type depth 32, 512 declarations, and 16 ports. Sparse/dangling/cyclic/orphaned graphs, mismatched versions/snapshots/profiles/contexts/fingerprints, and exceeded bounds produce typed refusals. Status exit codes remain `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; infrastructure failures use 1.

For the contracts around typed expressions and identity, see the [formal contract](formal-contract.md), [compatibility contract](compatibility.md), [composition contract](composition.md), and [normalization contract](normalization.md). This artifact path stores a representation; it does not run the rewrite rules or evaluate facts.

# DSE-03-03: immutable expression DAG and validated replay

Coordinator assessment: **ACCEPT within this bounded artifact contract**, supported by independent Luna source/output review. Owner language acceptance: **NOT_EVALUATED**. An artifact fingerprint is not a logical equivalence proof, signature or authentication of an entity reference.

## Literal execution

observations.jsonl contains 78 actual CLI/process records: argv, stdin when used, stdout, stderr, exit and selected process IDs. environment.json pins the base SHA, compiled binary and source hashes from the uncommitted implementation used for these observations. input-path-map.json connects temporary input paths to neutral data and tracked replay/tampering files.

The separate-process-export and separate-process-replay records have different PIDs. Their parsed artifact objects are identical, including the three receipts, all node records/IDs, typed binder/reference payloads, role edges, limits and provenance. The pipeline begins with the existing two-hop build recipe, exports its actual checked canonical document, then replays and inspects it; the independent algebra check on that canonical document succeeds.

The isolated API observation in cache-api-observation.json records request-local cache invalidation with the actual compiled registry loader/factory. Its first diagnostic program referenced an unavailable statistics getter and failed compilation; removing that diagnostic-only call changed no product code or acceptance requirement. The completed probe exits 0 and records real misses/hits. The temporary source copy excluded .git and product-extraneous files, used GOPROXY=off, and is not a committed test or runner. Its temporary location and source hash are retained; this is not an OpenCode delegation/isolation claim.

## Observed requirements

- Eight input cases export, structurally inspect and replay. Seven are complete; the typed hole remains incomplete/7 in every stage and never becomes a witness.
- Cache and no-cache preserve meaning fingerprints, cache keys, checked types and exact identity payloads. The repeated-ref case stores 4 versus 7 syntax records while the expanded occurrence count remains 7.
- Distinct refs and two separately bound Event instances remain distinct. The legal same textual identity in Var and Ref namespaces remains separated by term kind. Alias expressions remain pinned typed expressions, without automatic normalization.
- Resource labels are preserved, and changing a display label changes the resource fingerprint while retaining conservative meaning identity. No alpha-equivalence or label-driven rule selection is used for interning.
- The API probe observes first export miss, exact-repeat immutable hit, and misses after changing the actual compiled registry, selected profile, context ID, context dependency fingerprint or typed reference ID. A label-only source edit misses rather than returning old presentation data. Missing complete registry structure bypasses the whole-artifact cache. Hash-bucket collision comparisons use complete descriptors, not digest equality.
- Changed expected profile/context/dependencies/provenance, unknown normalizer/version, dangling edge/reference, nondense IDs, cycle, orphan, altered resource/meaning/cache receipts and too-small expanded-node budget are refused with explicit invalid/unsupported/resource-limit results.
- The expanded-budget control has four unique syntax nodes but seven occurrences; max_nodes=4 is refused. Interning does not evade the expanded-expression bound.
- Real metadata changes distinguish semantic snapshot mismatch from exact resource/version pin mismatch. Replay never derives its trusted metadata from the artifact.
- Inspect reports metadata_verified=false and requires no selected metadata. Supplying metadata to inspect, --no-cache to replay, missing export metadata or a wire-supplied environment is rejected. Stdin export works.
- Earlier metadata fingerprints match the recorded baseline binary for neutral and after-build resources. Existing normalize and check paths still execute successfully.

## Format and limits

The format contains one checked root. NodeID is a local syntax index; Binder/Param/Ref/Hole identity is a separate payload. Constructors/getters detach data. Replay validates dense IDs, all edges, reachability, acyclicity, depth, expanded occurrences, typing and scopes under independently selected metadata/profile/context/provenance before returning an immutable artifact. Requested limits can tighten wire limits.

Default/hard expression bounds remain 4096 expanded nodes, depth 64, type depth 32, 512 declarations and 16 ports. Export requests are limited to 1 MiB, public artifact envelopes to 8 MiB, replay requests to 16 MiB and JSON nesting to 256. Whole-artifact cache bounds are 16 entries and 16 MiB of serialized comparison descriptors, not heap/CPU/energy guarantees. Private full registry structure is separately bounded and never included in public resource fingerprint material.

Resource identity covers the public envelope body excluding the three receipts. Conservative meaning identity excludes presentation labels and local NodeIDs while preserving typed IDs/roles/order, normalizer version, formal registry/profile and explicit context dependencies. Complete comparison descriptors protect in-memory lookup collisions. No persistent knowledge store, automatic learning, network/path loading from data, language changes, tests, runner, golds or CI change was introduced.

The current normalizer version is a required dependency, not a migration mechanism. Profiles are recorded dependencies here; storage does not invoke normalization or evaluate facts. Context IDs/pins delimit caller-supplied request dependencies and do not authenticate entity identity. Provenance remains explicit data, separate from owner acceptance.

Gofmt, vet, build, help/version and the CLI commands were actually executed. Schema syntax, local references and preservation of the existing expression-schema root were checked; no new schema-validation dependency was installed. PR/main CI and post-merge receipt are pending in this pre-PR snapshot and will be recorded on #383 after actual execution.

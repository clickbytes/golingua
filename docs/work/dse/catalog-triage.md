# #385 — catalog triage and bounded pilot

This page records a source-neutral inventory and a small, explicit pilot. Catalog entries, formal symbols, language bindings, and runtime coverage are separate quantities. The current catalog contains 1002 source records; the 1000-concept value is an editorial target, not the current count or a target number of primitives. No catalog entry becomes a formal fact merely because it has a parent, roles, or a descriptive definition.

## Exporting the catalog

The compiled binary provides:

```sh
golingua algebra catalog export --json
```

It reads the compiled catalog in the binary and accepts no `--input`, `--metadata`, or `--no-cache` option. Its JSON envelope is `status`, `validation_scope`, and `inventory`; the inventory contains `schema_version`, `format`, `version`, `counts`, `source_records`, and `packages`.

The source inventory preserves each typed CORE definition's ID, kind, parent, role names and allowed-kind unions, definition edges and strengths, semantic effects, description, and original `SourceRef`. The packages carry the same IDs and descriptions as `opaque_unformalized` symbols with `review_status: pending`. They deliberately contain no formal ports inferred from source roles, no axioms, no aliases, no dependencies, and no typical edges. A CORE event's role union is source metadata, not a relation signature.

For the observed 1002-entry catalog, the exporter produces deterministic chunks of 400, 400, and 202 symbols. Each chunk is an independently validated/individually loadable metadata package. The metadata loader limits the selected aggregate to 512 declarations, so these three chunks cannot be selected together as one context. To inspect an emitted chunk, ordinary JSON extraction is sufficient; it does not need a converter or a new runner:

```sh
golingua algebra catalog export --json > /tmp/catalog-export.json
python3 -c 'import json; x=json.load(open("/tmp/catalog-export.json")); json.dump(x["inventory"]["packages"][0], open("/tmp/catalog-part-1.json","w"))'
golingua algebra metadata validate --metadata /tmp/catalog-part-1.json --json
```

The export is bounded to 4096 source records, 400 records per chunk, 16 chunks, 1 MiB per package, and 16 MiB combined. A limit or package-validation error returns no partial inventory. Existing `catalog --summary` and `catalog --details` continue to report catalog/source counts separately from metadata packages, primitive classifications, and supported routes.

## Review ledger

[`catalog-review.tsv`](../../../data/semantic/catalog-review.tsv) has one row per source catalog ID (1002 unique rows). Its classification vocabulary distinguishes `primitive_candidate`, `derived_metadata`, `opaque_unformalized`, and `excluded_from_pilot`; review and owner status are separate columns. At this snapshot there is one `primitive_candidate`: `concept:reading`, marked `agent_reviewed`, with owner status `NOT_EVALUATED` and a versioned mapping. The other 1001 remain `opaque_unformalized`, `pending`, `NOT_EVALUATED`, with `no_mapping`. There are no `derived_metadata` or excluded rows.

The ledger is a review aid, not a registry input or automatic approval mechanism. The full catalog export remains opaque and pending, including `concept:reading`. Descriptions, parent links, and role names do not generate an equality alias, implication, typicality rule, or world fact. The initial profile has no executable axioms. An equality alias, a one-way implication, and descriptive typical information are different claims and require separate review and representation.

## Explicit pilot package

The separate [`pilot metadata package`](../../../data/semantic/pilot/metadata.json) contains three approved primitives for structural use: neutral `pilot.r(left: Entity, right: Entity)`, neutral `pilot.mark(value: Entity)`, and mapped `concept:reading(event: Event, reader: Entity, material: Entity)`. The neutral symbols have no CORE or lexical mapping. The reading relation preserves the source CORE reader/material roles and adds an explicit event-identity port. That port is a versioned authored mapping decision; it is absent from the original CORE role list and must not be described as an implicit source fact or an unversioned change to `concept:reading`.

The accompanying [`mapping record`](../../../data/semantic/pilot/mapping.json) names the PL anchors (`sense:pl.czytac.interpret`, `frame:pl.czytac.interpret.frame`, `construction:pl.transitive.svo`) and EN anchors (`sense:en.read.text`, `frame:en.read.text.transitive`, `construction:en.transitive.svo.reading`). It preserves the original CORE, lexical-review, and construction-review provenance separately from the new mapping source. The approved registry status gates only this bounded structural pilot use. The existing lexical/construction approvals are not approval of a DSE equivalence, a new PL/EN route, or language quality; owner acceptance remains `NOT_EVALUATED`.

Seeing and hearing remain opaque in the inventory. Their source role definitions include `Entity | Event` unions that the current metadata relation-port type cannot represent directly. Narrowing those unions to `Entity` without an explicit reviewed mapping would lose information, so the ledger does not infer a signature for either concept.

The copied [`reading-check`](../../../data/semantic/pilot/reading-check.json), [`reading-build`](../../../data/semantic/pilot/reading-build.json), and [`two-hop`](../../../data/semantic/pilot/two-hop.json) inputs demonstrate metadata validation, a typed reading atom, and neutral composition. They are small explicit requests, not a corpus or automated test runner. The export chunk containing `concept:reading` and the pilot package must not be loaded together because they duplicate that symbol ID; that pair alone does not exceed the aggregate declaration budget. Selecting all three inventory chunks together does exceed the 512-declaration aggregate limit.

## Current observations and status

On the active branch, the frozen CLI returned `ok` for catalog export (1002 records; chunks 400/400/202), pilot metadata validation, the reading check/build, and the neutral two-hop build. A wrong event argument was refused with `ARGUMENT_TYPE_MISMATCH`; the old `concept:seeing` request was refused with `UNKNOWN_SYMBOL` in the small pilot. The existing catalog summary/details were observed unchanged. These are structural CLI observations, not proof of a world fact or a language-quality decision.

The pilot's corrected original provenance references were revalidated after the first CLI recording. Because provenance participates in the resource fingerprint, do not reuse a fingerprint from an earlier pre-correction observation as the fingerprint of the final metadata file. The final evidence receipt is pending at `docs/work/dse/evidence/02-03/`; no PR, merge, or main-CI result is claimed here. Agent assessment and owner acceptance remain separate; owner language acceptance is `NOT_EVALUATED`.

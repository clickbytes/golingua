# Versioned metadata packages

The metadata package API and CLI add an explicit, source-neutral registry around the typed IR. This page describes current bounded behavior for #380. A decoded `Document` is not an activated registry; `Compile` or `LoadPackages` validates it before producing immutable snapshots.

## CLI selection

```sh
golingua algebra metadata inspect --metadata package.json [--metadata dependency.json ...] --json
golingua algebra metadata validate --metadata package.json [--metadata dependency.json ...] --json
golingua algebra check --input expression.json --metadata package.json [--metadata dependency.json ...] --json
```

Each repeated `--metadata` selects one file explicitly. The loader reads only those paths; package data cannot request filesystem or network access. A maximum of 16 files is accepted, with a 1 MiB wire limit per file. `algebra check` may read its expression from stdin when `--input` is absent or `-`. When metadata is selected, inline expression `sorts` and `signatures` are refused, then approved primitive declarations are hydrated into the same algebra checker. The canonical result includes the hydrated declarations and selected package IDs, versions, and resource fingerprints.

`inspect` presents immutable package metadata, symbol kind/review classification, provenance, counts, and both fingerprints. `validate` reports structural/profile checks completed. Neither operation claims theorem consistency, world truth, language approval, or owner acceptance. The `Decode` API checks the strict wire shape only; it does not activate a package.

## Package contents and validation

Schema version 1 requires `id`, `version`, profile `positive-facts-v1`, provenance, sorts, symbols, dependencies, typical edges, and an empty `axioms` array. The strict [package schema](../../../schemas/semantic/package.schema.json) links the [symbol variants](../../../schemas/semantic/symbols.schema.json); types and terms reuse the [typed expression schema](../../../schemas/semantic/expression.schema.json). Unknown fields, duplicate JSON keys, null fields, malformed tagged variants, nonempty axioms, and relation ports with `Arrow` types are rejected.

The package loader checks bounded IDs and counts, each `SourceRef`, unique source IDs within each provenance record, sort and symbol ID collisions, type declarations, structural constraints, typical-edge references, alias closure/interface/type, and dependency pins. `arity.count` must equal the declared port count. `port_type` must equal the named port's declared type. These constraints only describe structure; they cannot encode domain laws. The initial axiom allowlist is empty. `typical_edges` are descriptive metadata and never become facts.

All dependency packages must be explicitly supplied. The loader checks exact ID, version, resource-fingerprint pins, missing packages, duplicates, cycles, and references restricted to a package's declared dependency closure. It deduplicates shared transitive packages by package ID and returns a deterministic dependency-first order. Any error returns no partial registry list. There is no path lookup, import callback, handler name, script, or label-driven dispatch.

## Immutable snapshots and identity

Registry constructors are private; callers receive snapshots only through validation. Documents and symbols are returned as detached copies. The operational signature view exposes only `approved` `primitive` symbols. `pending` primitives, `opaque_unformalized` records, and aliases remain inspectable but are not operational symbols. Alias definitions are checked as real closed terms against declared primitive/alias signatures and dependency sorts; free parameters/references, holes, invalid port/binder mappings, wrong result types, opaque uses, and alias cycles are refused. External `algebra check` currently refuses alias use as `unsupported`; it does not expand aliases.

Fingerprints use versioned SHA-256 domains. The resource fingerprint includes package metadata, including labels, descriptions, provenance, review state, edges, and declared dependency pins. The formal denotation ID omits labels, descriptions, provenance, review state, and typical-edge annotations. It retains package/sort/symbol IDs, profile, ordered signatures, constraints, alias structure without term labels, and sorted direct dependency IDs paired with each resolved dependency's formal denotation ID. Dependency resource fingerprints and version strings are not copied into that formal projection. This is conservative structural identity, not proof of arbitrary equivalence.

`SourceRef` is attribution metadata only: it is not proof of permission, factual truth, semantic approval, or language acceptance. Likewise, `review_status: approved` gates operational access for a primitive; it does not assert a fact. A successful `algebra check` establishes only that the typed expression passed structural checks under the selected registry.

## Examples and current evidence

The [neutral package](../../../data/semantic/examples/metadata/neutral.json) declares one nominal sort, one approved two-port primitive with structural arity/type constraints, an explicit two-step alias, and one pending opaque record. Its paired [expression](../../../data/semantic/examples/metadata/expression.json) uses the primitive and declared references without inline signatures. The #380 [evidence index](../../work/dse/evidence/02-01/README.md) separates observed CLI behavior from the planned post-build two-package experiment.

The formal design references remain [the DSE formal contract](formal-contract.md), [Meaning v1 compatibility](compatibility.md), and [typed IR](ir.md). These contracts do not imply facts, theorem consistency, synthesis, language integration, or owner acceptance.

## Catalog inventory and the #385 pilot

`algebra catalog export --json` serializes the compiled catalog as an inert inventory with `schema_version`, `format`, `version`, `counts`, `source_records`, and `packages`. It takes no file, registry, or cache flags. Source records preserve their typed CORE definitions and original provenance. The emitted metadata chunks intentionally classify every catalog symbol as `opaque_unformalized` and `pending`, with no inferred ports, aliases, axioms, or relation rules. For the current 1002 records, the deterministic package sizes are 400/400/202. Each package is separately loadable; the 512 aggregate declaration limit prevents selecting all chunks together. See the [catalog triage ledger and pilot notes](../../work/dse/catalog-triage.md).

The separate [`pilot package`](../../../data/semantic/pilot/metadata.json) is an explicitly reviewed, bounded structural selection: two neutral mathematical primitives and a versioned mapping for `concept:reading`. The mapping adds an `Event` identity port while retaining the CORE reader/material roles. It is documented in the [mapping record](../../../data/semantic/pilot/mapping.json) and is distinct from the original CORE and lexical/construction `SourceRef`s. This mapping does not assert that the source CORE definition contains an event port, establish unrestricted equivalence, create a fact, or approve a natural-language route. The `approved` package status enables only its declared structural use; owner language acceptance remains `NOT_EVALUATED`.

The pilot's `reading-check.json`, `reading-build.json`, and `two-hop.json` are explicit neutral/structural requests, not a test runner. A changed provenance record changes the resource fingerprint; do not reuse a pre-correction value as the final pilot fingerprint. Export inventory chunks and the pilot manifest must not be selected together because of the duplicate `concept:reading` ID and aggregate metadata limit. Broader catalog mappings and language integration remain outside this pilot.

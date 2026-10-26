# DSE-02-03 #385 — catalog and pilot evidence

This directory records the active-branch observations for the catalog inventory, inert metadata chunks, ledger, and separate structural pilot. It is not a language-quality test set or an acceptance record.

## Frozen command record

[`environment.json`](environment.json) identifies the uncommitted source state pinned by per-file hashes, Go version, frozen executable path, and executable hash. The observations are preserved literally in [`observations.jsonl`](observations.jsonl) (32 command records); [`invariants.json`](invariants.json) records the independently checked catalog/ledger invariants. The binary is `/tmp/golingua-dse/bin/golingua-385-final`, SHA-256 `1b54cae2e274697b6c1369015a1754aff33c2ce6cd3ff7a3f8c22697499a47eb`, based on `b63a2560dd0d2ca24c76fad3a80e3937a8f0285d`. The captured commands include `gofmt`, `go vet ./...`, build, `--help`, and `--version`, all with exit 0. This evidence job did not rebuild the binary.

## Catalog export and ledger

The actual `algebra catalog export --json` observation returned `ok` with 1002 source records and three packages sized 400/400/202. Repeated export was byte-identical. All emitted symbols were opaque and pending; the independent `catalog --summary` and `catalog --details` observations matched their pre-export snapshots. Each chunk validated independently. Selecting all three together returned `resource_limit` (exit 4) because the selected aggregate exceeded 512 declarations. A `concept:seeing` check against an inert export package returned `unsupported` (`SYMBOL_NOT_OPERATIONAL`, exit 3).

The TSV ledger contains 1002 unique IDs with exact source coverage. It identifies one reading `primitive_candidate` reviewed by the agent, owner status `NOT_EVALUATED`, and a mapping pointer; 1001 records remain opaque/pending with no mapping. This classification does not promote the export package: the full catalog export still contains only opaque/pending symbols, including `concept:reading`.

## Pilot observations

The separate [`pilot metadata`](../../../../../data/semantic/pilot/metadata.json) has three structurally approved primitives. Final-provenance commands in the log returned `ok` for metadata validation, the reading check, the reading build, and the neutral two-hop build. A wrong event argument returned `invalid` with `ARGUMENT_TYPE_MISMATCH`; an old `concept:seeing` request in the small pilot returned `invalid` with `UNKNOWN_SYMBOL`; selecting pilot metadata with the inventory chunk containing the same `concept:reading` ID returned `invalid` for the duplicate symbol.

These commands establish only bounded structural checks. The `concept:reading` pilot adds an authored `Event` identity port while preserving `reader` and `material`; it does not establish unrestricted CORE equivalence, facts, inference rules, or a new PL/EN translation route. The ledger's candidate classification and package review status are separate from owner language acceptance, which remains **NOT_EVALUATED**.

The original independent inputs are copied under [`inputs/independent`](inputs/independent/), and the exact emitted chunks under [`exported-packages`](exported-packages/). The first pilot outputs used pre-correction provenance. Their inputs are retained under [`pre-provenance-correction`](pre-provenance-correction/); `RECONSTRUCTION.txt` explicitly marks those JSON files as semantically reconstructed, not byte-identical snapshots. The invariants record that the reconstructed metadata reproduced the earlier output. Use the final-provenance command records for the current package. Since provenance changes the resource fingerprint, no earlier fingerprint is presented as the final fingerprint.

Final pilot data hashes:

- `metadata.json`: `e97ec4bd03d3c19793a0e77519a89e83949b63b6c3ffe8f40b78a4e3dd85f9b9`
- `mapping.json`: `b776586a9f3bd59eae53744a8ee7f6b1a144de53bf9ab28e56034323891dbb17`
- `catalog-review.tsv`: `f55facf62db8800ca99861d0964fb136aa123ca0edf62eead1210683142e6919`

The canonical issue-map entry remains active with no PR URL or merge SHA yet. PR review and main CI are pending; no merge receipt or owner acceptance is claimed. The coordinator's technical assessment and independent review remain separate from owner language acceptance.

The final staging check found trailing tabs for the empty last TSV column. The coordinator applied the reviewer-specified column reorder (`lexical_anchors` before nonempty `reason`); all 1002 row dictionaries and IDs were verified unchanged. The final ledger hash above covers that editorial ordering. Source, binary, metadata and literal CLI observations are unchanged.

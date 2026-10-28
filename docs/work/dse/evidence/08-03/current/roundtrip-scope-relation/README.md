# #398 scope/relation round-trip evidence

Run date: 2026-10-28 UTC. Job base: `bff69b4`; the built binary embeds source revision `bff69b46c04dc6f1c3d31fb170bfe3bd261d96d6` and reports `vcs.modified=true`. Binary: `/tmp/golingua-dse/398-roundtrip-scope-relation/golingua`, SHA-256 `88699ad2d387ca0b30fa986ccad2eb2a0ba549825e863a1dd1cb4dd16d416f13`. Build command and runtime pins are recorded in [`environment.json`](environment.json). Runtime configs, language packages, source/target plans, context, metadata and independent documents are under [`inputs/`](inputs/); their hashes are in [`inputs.sha256`](inputs.sha256). The active snapshot used was `golingua-authored` version `2026-10-23`, digest `e62fdf8903c59fd27690aa46e58b00d199a10a0e1108869e3aa8d47a8f7196f0`.

The runtime snapshot copy under `inputs/snapshot-store/` was used for the run but is not committed as a duplicate of the dataset. To reproduce, place a store with the recorded active snapshot at that path or create one there using the CLI snapshot workflow, then verify its digest and the recorded input hashes. The configuration and manifest retain the paths actually used; the listed snapshot hashes document the omitted files.

[`commands.jsonl`](commands.jsonl) records all 98 production CLI invocations with exact argv, cwd, times, exit codes, embedded source revision, binary hash, and stdout/stderr paths and hashes. Raw outputs are in [`outputs/`](outputs/). No runner was added. The flow and selected candidate details are in [`roundtrip-results.json`](roundtrip-results.json).

## Observed flows

Each row is a full source → target → source sequence. The returned surface was reanalyzed. Both target and return candidate documents were compared with independent IRs.

| Round-trip | Source analysis | First translation | Target analysis | Return translation | Return analysis | Projected comparisons (target / return) | Round-trip status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Outer PL → EN → PL | ok | ok | incomplete | incomplete | ok | proved_equivalent / proved_equivalent | incomplete |
| Outer EN → PL → EN | incomplete | incomplete | ok | ok | incomplete | proved_equivalent / proved_equivalent | incomplete |
| Inner PL → EN → PL | ok | ok | incomplete | incomplete | ok | proved_equivalent / proved_equivalent | incomplete |
| Inner EN → PL → EN | incomplete | incomplete | ok | ok | incomplete | proved_equivalent / proved_equivalent | incomplete |
| Relation PL → EN → PL | ok | ok | incomplete | incomplete | ok | proved_equivalent / proved_equivalent | incomplete |
| Relation EN → PL → EN | incomplete | incomplete | ok | ok | incomplete | proved_equivalent / proved_equivalent | incomplete |

All 12 explicit `algebra realize` surface retrievals returned a unique text-bearing `ok` alternative and matched the corresponding text selected from `translate`. Three return realizations to Polish had aggregate `incomplete` with a unique `ok` text-bearing alternative; those aggregate statuses are preserved. Every end-to-end round-trip remains `incomplete` because at least one stage in each sequence was incomplete.

For each EN → PL first translation, the aggregate remained `incomplete`. Selection was conditional and by the unique candidate with `outcome: ok` and nonempty text. The other candidate had `outcome: unsupported`, no text, and diagnostic `DSE_SOURCE_CANDIDATE_NOT_COMPLETE`. The selected IDs and rejected candidates are recorded in `roundtrip-results.json` and the raw translation outputs. The same rule was applied to any later incomplete translation; no incomplete aggregate is promoted to `ok`.

## Comparison scope and preserved attempts

Full-context comparisons were attempted for all target and return analyses. All 12 returned `invalid` with `EQUIVALENCE_REQUEST_CONTEXT_MISMATCH`: the analyzed documents carry request context `dse-398-empty`, while the independent scope files carry `scope-contrast-context`; the relation document also has a different context. These results remain in the raw outputs.

The context-projected comparisons remove only `request_context`. For scope, the timed expression files omit inline declarations, so the final comparison uses the existing `compare-not-all.json` / `compare-all-not.json` copies, which supply declarations from the independent acceptance documents while keeping the timed expression identical. The expression equality was checked. For relation, the existing present-tensed independent document was used. All 12 projected comparisons returned `proved_equivalent` under the selected `attrs-beta-alpha` rewrite profile. This is conditional on that profile and metadata and does not compare request contexts, establish world truth, or certify linguistic quality.

Two correction attempts are retained rather than hidden: the first scope comparison used untimed upstream copies and is stored under each comparison's `*_initial_attempt`; the first corrected timed compare without inline declaration supply returned `UNKNOWN_SYMBOL` and is recorded as `context_projected_without_acceptance_declarations`. The final projected comparisons use the existing timed files and declaration-bearing compare copies. No expected expression was changed.

The relation source is the existing timed function description; the independent relation document specifies `present` on both events. No owner approval is implied. English/Polish linguistic acceptance: **NOT EVALUATED**. Go tests and test runner: **NOT RUN**.

## Additional comparisons in one explicit empty context

A separate follow-up used exactly 12 additional `algebra compare` invocations, recorded in [`inputs/same-empty-context/commands.jsonl`](inputs/same-empty-context/commands.jsonl), with raw output in [`outputs/same-empty-context/`](outputs/same-empty-context/). The preflight audit is [`precompare-validation.json`](inputs/same-empty-context/precompare-validation.json): all 12 analyzed documents had an explicit empty `dse-398-empty` request context; the existing independent contexts were either empty (`scope-contrast-context` on timed scope originals) or missing (relation). No nonempty referents, mentions, references, speaker/addressee, conditions, or resolution evidence were found.

The compare copies derive from the existing independent timed documents. For scope, the existing declaration-bearing comparison copy was used; its expression was checked identical to the timed independent file. For relation, the existing present-tensed independent file was used directly. Each independent comparison copy received the exact empty `request_context` object from its analyzed counterpart, with ID `dse-398-empty`. Before/after checks confirm expression, references, temporal fields (within expression), signatures and sorts were unchanged; derivation records and hashes are included in the preflight audit. The new results are in [`results.json`](inputs/same-empty-context/results.json), and inputs are hashed in [`sha256.txt`](inputs/same-empty-context/sha256.txt).

All 12 comparisons returned `proved_equivalent` under `attrs-beta-alpha` with an exact matching `dse-398-empty` context ID. This verifies the supplied documents after placing both comparison operands in the same explicitly empty context. It does **not** claim equivalence under the original differing context IDs: the earlier 12 `EQUIVALENCE_REQUEST_CONTEXT_MISMATCH` results remain unchanged above and in the original outputs. No expression, reference, time, or signature was rewritten for this comparison set. Owner linguistic acceptance remains **NOT EVALUATED**; Go tests and runner remain **NOT RUN**.

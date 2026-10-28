# WH round trips

The four recorded flows cover subject and object questions in both directions: EN→PL→EN and PL→EN→PL. `roundtrip-results.json` retains each stage, answer role, selected candidate ID, status, assumptions and residuals. Exact CLI commands and output hashes are in `commands.jsonl` and `inputs/compare-corrected/commands.jsonl`; raw stdout and stderr are in `outputs/`.

All source, target and return analyses returned `ok`. Each translation remained `incomplete`, and each surface was selected explicitly from a unique text-bearing candidate; no incomplete aggregate was promoted to `ok`. The answer roles remain `experiencer` for subject WH and `stimulus` for object WH.

The corrected comparisons after explicit context projection returned `proved_equivalent` under the selected profile. Full-context comparisons remained `invalid` with `CONTEXT_REFERENCE_MISSING`. These results do not establish full context equivalence, world truth or linguistic approval. Initial failed comparison attempts remain recorded.

The configured snapshot store was used locally but is not committed as a duplicate dataset. Reproduction requires a `golingua-authored` snapshot at `inputs/snapshot-store/` with the digest and file hashes recorded in `environment.json` and `inputs.sha256`. Source revision and binary hash are recorded there and in the command manifests.

Agent linguistic assessment is separate from owner acceptance: **NOT EVALUATED**. Go tests and a test runner were not used for this evidence.

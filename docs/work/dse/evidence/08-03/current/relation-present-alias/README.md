# Anonymous relation with an explicit present-time definition

These observations use source commit `ef45e22f0957d22de4b1c8bcdb51f60dfba86fce`. Exact commands, binary and output hashes, exit codes and timestamps are in `commands.jsonl`. The copied runtime configuration refers to the recorded local snapshot store and files under `/tmp/golingua-dse/398-scope-relation-config`; reproducing the source analyses requires restoring those paths and the active `golingua-authored` snapshot. No test runner or language corpus is introduced here.

The independently authored `two-hop.json` remains unchanged. A separate, explicitly selected `vision-event-present-binary-proposal.json` defines its binary alias as an existential vision event with `temporal_frame: present`. This is an agent-reviewed formal proposal for this bounded slice, not a world fact, a general meaning of *see*, or owner acceptance. The earlier [untimed proposal and its incomplete previews](../relation-alias-preview/README.md) remain unchanged.

Metadata validation returned `ok`. The PL source analysis returned `ok`; the EN aggregate returned `incomplete`, retaining one successful visual-sense candidate and another unsupported candidate. Each explicitly selected successful candidate compared with the independent anonymous relation as `proved_equivalent` under the recorded alias profile after projecting the request context. This establishes expression equivalence under the selected proposal, not full context equivalence or an unambiguous English analysis.

Normalization of the independent anonymous relation returned `ok`; separate verification reported `certificate_valid: true` and `normalization_complete: true`. The expanded independent IR realized in EN and PL with `ok` and no residuals. A control with the first event's entity roles swapped returned `undetermined`, which does not prove non-equivalence. No assertion about world truth follows from any of these checks.

Linguistic assessment: bounded agent proposal. Owner linguistic acceptance: **NOT EVALUATED**. The remaining EN ambiguity and the WH context limitations remain open in #398.

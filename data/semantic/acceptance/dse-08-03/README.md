# DSE-08-03 independent formal candidates

These small typed documents record the independent requirements for active issue #398. They are agent-authored formal candidates for CLI diagnosis, not owner-approved language meanings, production golds, or a test corpus. They contain no source trace or copied sentence evidence. Surface proposals and known ambiguities remain in `docs/work/dse/vertical-slices.md`.

- `wh/subject-ask.json` asks for the experiencer while holding stimulus `B` fixed.
- `wh/object-ask.json` asks for the stimulus while holding experiencer `A` fixed.
- `wh/role-swap-*.json` holds the same context and reference attributes constant while swapping the answer role.
- `relation/two-hop-goal.json` composes the existing `pilot.r` relation twice through one shared Entity binder. `relation/two-hop.json` is the corresponding independently written typed expression. Neither asserts an English or Polish lexical mapping for `pilot.r`.
- `scope/not-all.json` and `scope/all-not.json` preserve distinct negation placement. The event variable is existentially closed inside each teacher/dog clause because the selected perception relation has an explicit Event port.
- `wh/subject-ask-past.json` is a diagnostic refusal control for unsupported past-time evaluation.

The request-context conditions in the WH files are explicit descriptions/presuppositions only. They are not fact records. The diagnostic `facts-active.json` used by the evaluator is synthetic; its answers do not establish real-world facts or approve the source phrases.

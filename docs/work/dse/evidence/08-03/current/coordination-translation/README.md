# Configured DSE coordination bridge — #398, 2026-10-28

Code `56ddff9`, same draft PR #422. The bounded bridge validates two source clause records against the checked `And`, exact approved source frames/symbol targets, typed Event identity and role→port→Entity reference mapping. It enriches both clauses separately and preserves all source roles/receipts. Repeated agreeing lexical descriptors can be consumed idempotently; inconsistent mappings/coverage refuse consumption. Definiteness and other unsupported conditions remain.

`final-pl-en` returns sixteen explicit alternatives, e.g. `The teacher sees the dog and the dog sees the teacher.` Status is `ambiguous`; all article assumptions and warnings remain. `final-en-pl` refuses the target and retains both `definite_reference` conditions. Missing second-event identity gives `incomplete` with no candidates; contradictory supplied class gives `invalid`, without target text. Four existing EN/PL subject/object WH translation controls still return their texts as `incomplete`, with residuals/assumptions retained.

The pinned runtime configs include separate source contexts for EN and PL. They use the same existing snapshot store as the preceding #398 evidence; no store mutation was performed. `environment.json` records the immutable final binary and code SHA. `commands.jsonl` records actual commands, outputs, exits and hashes. The code change passed gofmt, go vet, build, help/version smoke and diff checks; no test files/runner/CI gate were added.

## Partial round-trip observation

`roundtrip-selected-en.json` records an agent diagnostic selection of the first all-definite PL→EN variant and its assumptions. This is not an owner preview selection or a claim that the other fifteen alternatives disappear. The selected text is reanalyzed using an explicitly supplied EN identity context matching its token positions. `roundtrip-pl-en-analyze` returns one complete checked conjunction with E1 A→B and E2 B→A. `roundtrip-pl-en-expression-compare` proves equivalence to the independent conjunction **after explicitly projecting request context**. The full original and selected-context documents remain recorded.

The added EN definite presuppositions are explained by the retained target assumptions, not silently removed from the originals. Full context equivalence for this direction is not proved. The reverse EN→PL→IR round-trip is **NOT RUN**, because configured EN→PL produces no target text. Thus this is not six-flow acceptance of the whole relation slice, nor a realization of the generic lambda/existential relation goal.

## Remaining work

The bridge subtask is implemented but the full #398 remains PARTIAL/BLOCKED. Source/target scope cannot be claimed from these two direct Atoms. The separate [scope contract audit](../scope-contract/README.md) identifies the missing binder/time realization contract. WH context residuals and owner language acceptance remain separate, unresolved statuses.

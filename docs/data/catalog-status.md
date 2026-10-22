# Compiled CORE / PL / EN inventory

From the repository root:

```sh
go run ./cmd/golingua catalog --summary --json
go run ./cmd/golingua catalog --details --json
```

The CLI compiles the current CORE, lexical and language-specific valency snapshots. It reports statuses of **real Go data**, not proposed editorial labels and not percentages of translation accuracy.

| Inventory | CORE | PL | EN |
| --- | ---: | ---: | ---: |
| CORE concepts | 1000 | — | — |
| Lexemes | — | 982 | 376 |
| Lexical senses | — | 1007 | 385 |
| Pending concept binding records | — | 1006 | 377 |
| Approved concept bindings | — | 0 | 0 |
| Concepts without a binding | — | 0 | 624 |
| Authored valency frames | — | 273 | 5 |

Approved PL+EN pairs: **0**. The `valency_frames` field is separately compiled from `pl.CompileValency()` and `en.CompileValency()`. `catalog --details` has one per-concept row per CORE ID, including true pending/absent/approved states.

The Polish editorial `data/pl/1000-core-review.tsv` contains 1000 proposed decisions and 1000 pending candidate rows. These rows produce 1006 typed pending binding records over all 1000 concepts. Candidate mappings, approved mappings, linguistic quality, and morphological coverage must not be conflated.

The English ledger `data/en/1000-core-review.tsv` covers the exact 1000 compiled CORE IDs. It records 19 existing pending candidates, 357 newly assessed concepts with 357 English noun lemmas, and leaves 624 concepts `UNREVIEWED`; it does not infer lemmas from CORE definitions. The existing `en.school.institution` sense is now linked as a pending candidate; eight pre-existing English senses remain unmapped and are listed in `docs/languages/en.md`. The current EN increments add 357 noun lexemes, with no new morphology, valency, or construction coverage.

Polish `forms` is a separate partial subsystem: **15 newly supplied noun inventories with 14 explicitly declared case/number forms each**, plus the pre-existing narrow seed. These form strings are not mechanically generated from an inflection class, but selected from declared principal parts through five shared morphology paradigms. No reviewed sentence construction, analyzed clause or generated PL↔EN translation is activated. Current language manifests keep sentence analysis/realization unsupported.

The repo's technical CI builds the Go program and exercises `catalog` and the small `forms` smoke. **Owner semantic and inflection review: NOT RUN.**

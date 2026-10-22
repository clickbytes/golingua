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
| Lexemes | — | 957 | 19 |
| Lexical senses | — | 980 | 28 |
| Pending concept bindings | — | 979 | 19 |
| Approved concept bindings | — | 0 | 0 |
| Concepts without a binding | — | 24 | 981 |
| Authored valency frames | — | 270 | 5 |

Approved PL+EN pairs: **0**. The `valency_frames` field is separately compiled from `pl.CompileValency()` and `en.CompileValency()`. `catalog --details` has one per-concept row per CORE ID, including true pending/absent/approved states.

The Polish editorial `data/pl/1000-core-review.tsv` contains 1000 proposed decisions, 983 pending candidate links and 17 decisions without links; 7 event proposals still lack typed Go lexical/frame entries. Candidate mappings, approved mappings, linguistic quality, and morphological coverage must not be conflated.

Polish `forms` is a separate partial subsystem: **15 newly supplied noun inventories with 14 explicitly declared case/number forms each**, plus the pre-existing narrow seed. These form strings are not mechanically generated from an inflection class, but selected from declared principal parts through five shared morphology paradigms. No reviewed sentence construction, analyzed clause or generated PL↔EN translation is activated. Current language manifests keep sentence analysis/realization unsupported.

The repo's technical CI builds the Go program and exercises `catalog` and the small `forms` smoke. **Owner semantic and inflection review: NOT RUN.**

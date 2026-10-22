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
| Lexemes | — | 982 | 1000 |
| Lexical senses | — | 1007 | 1009 |
| Pending concept binding records | — | 963 | 12 |
| Approved concept bindings | — | 43 | 989 |
| Concepts without a binding | — | 0 | 0 |
| Authored valency frames | — | 273 | 5 |

Approved PL+EN pairs: **38**. The PL ledger review in #152 batch 1 approved 43 exact lexical equivalents across 43 concepts; 957 concepts remain pending in PL. The EN review approved 989 equivalents; its 12 pending records cover 11 concepts. The `valency_frames` field is separately compiled from `pl.CompileValency()` and `en.CompileValency()`. `catalog --details` has one per-concept row per CORE ID, including true pending/absent/approved states.

The construction index contains zero pending candidates and four approved sense/frame/family/variant bindings per language (PL 4, EN 4), following the owner's delegated review for #63. `catalog --summary` reports these separately. The bindings do not approve lexical CORE links or activate sentence translation; EN `look at` remains unbound because its oblique construction template is absent.

The Polish editorial `data/pl/1000-core-review.tsv` contains 1000 decisions: 43 approved equivalent rows and 957 pending candidate rows. The compiled graph contains 1006 binding records over all 1000 concepts. Candidate mappings, approved mappings, linguistic quality, and morphological coverage must not be conflated.

The English ledger `data/en/1000-core-review.tsv` covers the exact 1000 compiled CORE IDs. It records 19 existing pending candidates and 981 newly assessed concepts, including 694 noun, 264 verb and 23 preposition candidates, with no `UNREVIEWED` concepts; twenty delegated lexical batches approved 989 equivalent links across 989 concepts, while 12 binding records remain pending across 11 concepts. Eight first-batch candidates and three second-batch concepts remain pending for documented semantic or sense-boundary reasons. Eight pre-existing English senses remain unmapped and are listed in `docs/languages/en.md`. These lexical batches added no morphology or valency; the separate delegated construction review approved four EN routes.

Polish `forms` is a separate partial subsystem: **15 newly supplied noun inventories with 14 explicitly declared case/number forms each**, plus the pre-existing narrow seed. These form strings are not mechanically generated from an inflection class, but selected from declared principal parts through five shared morphology paradigms. No reviewed sentence construction, analyzed clause or generated PL↔EN translation is activated. Current language manifests keep sentence analysis/realization unsupported.

The repo's technical CI builds the Go program and exercises `catalog` and the small `forms` smoke. **Owner semantic and inflection review: NOT RUN.**

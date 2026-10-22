# Catalog status

Compilable inventory from the current typed Go source:

```sh
go run ./cmd/golingua catalog --summary --json
go run ./cmd/golingua catalog --details --json
```

After G07-03's entity + motion batches, expected structural inventory (verify by executing CLI and CI, not by assuming linguistic accuracy):

| Measure | CORE | PL | EN |
| --- | ---: | ---: | ---: |
| Concept definitions | 1000 | — | — |
| Lexemes | — | 728 | 19 |
| Lexical senses | — | 744 | 28 |
| Pending bindings | — | 743 | 19 |
| Approved bindings | — | 0 | 0 |
| CORE concepts absent from local bindings | — | 257 | 981 |

Approved concept pairs PL+EN: **0**. `catalog --details --json` reports 1000 per-concept rows and does not treat pending candidates as approved.

Separate Polish editorial worklist: **1000 rows** with 983 proposed pending binding candidates and 17 records with no candidate; **240 event candidates not yet compiled in the PL Go lexicon**. These numbers must not be mixed with `catalog.Inspect()` counters. PL has 36 authored valency frames; existence of frames or templates alone is not an approved construction path. The PL morphology generator is a tiny independent seed, not a generator for the 704 lexical entries.

English remains the original 19-lexeme/28-sense starter layer until #61 starts. Both PL/EN manifests still report unsupported sentence analysis and realization. Translating text requires independently approved bindings and runtime providers (#63/#123).

Manual linguistic acceptance on owner-supplied inputs remains **NOT RUN**.

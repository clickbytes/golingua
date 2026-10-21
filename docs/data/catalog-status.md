# Catalog status

Run the compiled in-code inventory from the repository root:

```sh
golingua catalog --summary --json
golingua catalog --details --json
```

The summary compiles the actual CORE and PL/EN indexes. The details view includes one `per_concept` row per active stable ConceptID and its real `pl`/`en` states (`approved`, `pending`, `non_transferable`, or `absent`). The summary omits per-concept rows. Neither view treats candidates as approved.

## Observed on main after the G07-03 nature and body batch

The CLI returned `status: ok`, **1000 CORE concepts / target 1000**, PL **210 lexemes, 217 senses, 216 bindings** and EN **19 lexemes, 28 senses, 19 bindings**. PL has **216 pending, 0 approved, 0 non-transferable, 784 absent** concepts; EN has **19 pending, 0 approved, 0 non-transferable, 981 absent**. Approved PL+EN pairs: **0**. `catalog --details --json` has one row for each of the 1000 compiled IDs.

The editorial worklist `data/pl/1000-core-review.tsv` covers all 1000 rows (479 resolved, 521 `DEFERRED` for later batches). On top of the People slice, batch 2 implements the reviewed nature domains (animals, plants, substances, landforms) and the human body domain in compiled Go; the remaining reviewed row categories (places, objects, food, and the basic nouns left from the original 49) are still worklist-only. Worklist decisions are not catalog bindings until they exist as typed Go data.

These values are catalog inventory counts. Lexemes, senses, bindings, inflected forms, constructions, and translation capability are separate layers; `catalog` does not count forms or constructions. The CORE proposals do not add language data or activate translation. Manual evaluation on owner-supplied examples remains **NOT RUN**.

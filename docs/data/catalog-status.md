# Catalog status

Run the compiled in-code inventory from the repository root:

```sh
golingua catalog --summary --json
golingua catalog --details --json
```

The summary compiles the actual CORE and PL/EN indexes. The details view includes one `per_concept` row per active stable ConceptID and its real `pl`/`en` states (`approved`, `pending`, `non_transferable`, or `absent`). The summary omits per-concept rows. Neither view treats candidates as approved.

## Observed on main after the G07-03 People batch

The CLI returned `status: ok`, **1000 CORE concepts / target 1000**, PL **75 lexemes, 81 senses, 80 bindings** and EN **19 lexemes, 28 senses, 19 bindings**. PL has **80 pending, 0 approved, 0 non-transferable, 920 absent** concepts; EN has **19 pending, 0 approved, 0 non-transferable, 981 absent**. Approved PL+EN pairs: **0**. `catalog --details --json` has one row for each of the 1000 compiled IDs.

The editorial worklist `data/pl/1000-core-review.tsv` covers all 1000 rows (479 resolved, 521 `DEFERRED` for later batches), but only the People-domain slice is implemented in compiled Go so far, with a few additional `candidate` links on existing multi-sense lexemes. Worklist decisions are not catalog bindings until they exist as typed Go data.

These values are catalog inventory counts. Lexemes, senses, bindings, inflected forms, constructions, and translation capability are separate layers; `catalog` does not count forms or constructions. The CORE proposals do not add language data or activate translation. Manual evaluation on owner-supplied examples remains **NOT RUN**.

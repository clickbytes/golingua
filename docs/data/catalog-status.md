# Catalog status

Run the compiled in-code inventory from the repository root:

```sh
golingua catalog --summary --json
golingua catalog --details --json
```

The summary compiles the actual CORE and PL/EN indexes. The details view includes one `per_concept` row per active stable ConceptID and its real `pl`/`en` states (`approved`, `pending`, `non_transferable`, or `absent`). The summary omits per-concept rows. Neither view treats candidates as approved.

## Observed on main after the G07-03 nature substantives batch

The CLI returned `status: ok`, **1000 CORE concepts / target 1000**, PL **451 lexemes, 460 senses, 459 bindings** and EN unchanged (**19/28/19**). PL has **459 pending, 0 approved, 0 non-transferable, 541 absent** concepts. Approved PL+EN pairs: **0**. `catalog --details --json` has one row for each of the 1000 compiled IDs.

The editorial worklist `data/pl/1000-core-review.tsv` covers all 1000 rows (533 resolved after the locomotion/substances editorial pass, 467 `DEFERRED` for the remaining verb, adjective and abstract domains). Compiled Go so far covers the reviewed People, nature (animals/plants/substances/landforms), body, places, objects, food, original-49 basics noun domains, and the seven core verbs behind explicit valency frames. The remaining DEFERRED rows (521 at the last check) still await their editorial pass, which will add the next verb frames per sense. Worklist decisions are not catalog bindings until they exist as typed Go data.

These values are catalog inventory counts. Lexemes, senses, bindings, inflected forms, constructions, and translation capability are separate layers; `catalog` does not count forms or constructions. The CORE proposals do not add language data or activate translation. Manual evaluation on owner-supplied examples remains **NOT RUN**.

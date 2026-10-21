# Catalog status

Run the compiled in-code inventory from the repository root:

```sh
golingua catalog --summary --json
golingua catalog --details --json
```

The summary compiles the actual CORE and PL/EN indexes. The details view includes one `per_concept` row per active stable ConceptID and its real `pl`/`en` states (`approved`, `pending`, `non_transferable`, or `absent`). The summary omits per-concept rows. Neither view treats candidates as approved.

## Observed on this branch

The CLI returned `status: ok`, **1000 CORE concepts / target 1000**, PL **16 lexemes, 20 senses, 16 bindings** and EN **19 lexemes, 28 senses, 19 bindings**. PL has **16 pending, 0 approved, 0 non-transferable, 984 absent** concepts; EN has **19 pending, 0 approved, 0 non-transferable, 981 absent**. Approved PL+EN pairs: **0**. `catalog --details --json` has one row for each of the 1000 compiled IDs.

These values are catalog inventory counts. Lexemes, senses, bindings, inflected forms, constructions, and translation capability are separate layers; `catalog` does not count forms or constructions. The CORE proposals do not add language data or activate translation. Manual evaluation on owner-supplied examples remains **NOT RUN**.

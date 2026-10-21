# Catalog status

Run the compiled in-code inventory from the repository root:

```sh
golingua catalog --summary --json
golingua catalog --details --json
```

The summary compiles the actual CORE and PL/EN indexes. The details view includes one `per_concept` row per active stable ConceptID and its real `pl`/`en` states (`approved`, `pending`, `non_transferable`, or `absent`). The summary omits per-concept rows. Neither view treats candidates as approved.

## Observed on main after the G07-03 batch-7 noun groups

The CLI returned `status: ok`, **1000 CORE concepts / target 1000**, PL **680
lexemes, 691 senses, 690 bindings** and EN unchanged (**19/28/19**). PL has
**690 pending, 0 approved, 0 non-transferable, 310 absent** concepts.
Approved PL+EN pairs: **0**. `catalog --details --json` has one row for each
of the 1000 compiled IDs.

The editorial worklist `data/pl/1000-core-review.tsv` covers all 1000 rows
after the closing editorial pass — **1000 resolved, 0 `DEFERRED`**. Decisions
include nouns, abstracts and relations plus all remaining verb domains;
verbs without per-sense valency frames stay worklist-only and are not
compiled into Go (same rule as the locomotion batch). Compiled Go so far
cover the reviewed People, nature (animals/plants/substances/landforms),
body, places, objects, food, original-49 basics noun domains, the seven
core verbs behind explicit valency frames, and the batch-7 noun groups
(qualities, abstracts, relations, transport, media, supplies plus the
dishware/service-place fractions routed into food and places). Worklist
decisions are not catalog bindings until they exist as typed Go data;
further verb frames and morphology are separate decisions after the
owner's first manual review.

These values are catalog inventory counts. Lexemes, senses, bindings, inflected forms, constructions, and translation capability are separate layers; `catalog` does not count forms or constructions. The CORE proposals do not add language data or activate translation. Manual evaluation on owner-supplied examples remains **NOT RUN**.

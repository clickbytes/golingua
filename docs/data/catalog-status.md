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
| Lexemes | — | 983 | 1003 |
| Lexical senses | — | 1007 | 1012 |
| Pending concept binding records | — | 124 | 4 |
| Approved concept bindings | — | 882 | 997 |
| Concepts without a binding | — | 0 | 0 |
| Authored valency frames | — | 273 | 5 |

Approved PL+EN pairs: **881**. Delegated PL review batches 1–20 and audit #192 (#152, #154, #156, #158, #160, #162, #164, #166, #168, #170, #172, #174, #176, #178, #180, #182, #184, #186, #188, and #190) approved 882 exact lexical equivalents across 882 concepts; 118 concepts remain pending in PL. The EN review plus #192 approved 996 equivalents; G10-12 #237 added a separately sourced agent-reviewed equivalent, bringing the total to 997. Four pending EN binding records cover 3 concepts. The `valency_frames` field is separately compiled from `pl.CompileValency()` and `en.CompileValency()`. `catalog --details` has one per-concept row per CORE ID, including true pending/absent/approved states.

The construction index contains zero pending candidates and 46 approved sense/frame/family/variant bindings (PL 4, EN 42), following reviews #63, #201, #213, #230, #233, and G10-12 #237. The EN bindings include mass-object routes and `a`/`an` determiner alternatives for the exact `seeing`, `hearing`, `looking`, and `receiving` routes. An indefinite article is licensed only when the target lexeme has an explicit `indefinite_article=a` or `indefinite_article=an` property; source analysis checks that declared allomorph and rejects plural readings after `a/an`. `catalog --summary` reports construction bindings separately from the 1006 PL and 1001 EN lexical binding records. These routes do not approve lexical CORE links; unresolved lexical candidates and sentence statuses remain visible.

The Polish editorial `data/pl/1000-core-review.tsv` contains 1000 decisions: 882 approved equivalent rows and 118 pending candidate rows. The compiled graph contains 1006 binding records over all 1000 concepts. Candidate mappings, approved mappings, linguistic quality, and morphological coverage must not be conflated.

The English ledger `data/en/1000-core-review.tsv` covers the exact 1000 compiled CORE IDs. It records 19 original-seed rows and 981 later authored concepts, including 694 noun, 264 verb and 23 preposition candidates, with no `UNREVIEWED` concepts; twenty delegated lexical batches and audit #192 approved 996 equivalent links across 996 concepts, and G10-12 #237 separately reviewed the existing `en.offspring.child` sense as equivalent to `concept:child_of_person`. In total 997 links are approved, while 4 binding records remain pending across 3 concepts. Three concepts remain pending for documented semantic or sense-boundary reasons. Eight pre-existing English senses remain unmapped and are listed in `docs/languages/en.md`. These lexical batches added no morphology or valency; construction reviews currently license 42 EN bindings, including `look at` (#201), mass-object `receiving` (#213), the receiving object alternative (#230), subject/object alternatives across existing routes (#233), and `a/an` allomorph routes (#237).

Polish `forms` is a separate partial subsystem: **22 noun inventories with 14 explicitly declared case/number forms each, plus seven case forms for plural-only `pieniądze`**, alongside the pre-existing narrow seed. These form strings are selected from declared principal parts through shared morphology paradigms; they are not inferred for arbitrary lexemes. The compiled CLI activates limited `seeing`, `hearing`, `looking`, and `receiving` sentence routes. Current base language manifests keep general sentence analysis/realization unsupported.

The repo's technical CI builds the Go program and exercises `catalog` and the small `forms` smoke. **Owner semantic and inflection review: NOT RUN.**

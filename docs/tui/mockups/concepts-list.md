# F3 Concepts list and search layout

**DESIGN / RECORDED — NOT LIVE GO OUTPUT.** F3 receives the real compiled
`CatalogResult.Inventory.PerConcept` list from
`app.Service.Catalog(ctx, app.CatalogRequest{Details:true})`. The list contains
stable ConceptIDs with one actual PL and EN coverage status each. This screen
filters the returned ID text locally; it does not show concept descriptions or
infer translations from IDs. TUI implementation is not available yet.

## 80 × 24 · matched ConceptID query

The query `concept:ab` returns the four rows below when applied as a
case-sensitive substring filter over the observed 1000 returned IDs. Rows and
coverage values match actual catalog output in canonical ConceptID order.

```text
+------------------------------------------------------------------------------+
| GoLingua | F3 Concepts | compiled catalog | RECORDED / DESIGN ONLY           |
| F1 Translate   F2 Forms   F3 Concepts   F4 Languages   F5 Settings           |
|+----------------------------------------------------------------------------+|
| F3 Concepts | search the returned ConceptID strings                          |
| Focus: > Query | Tab -> result list -> page controls                         |
| Catalog: ok | 1000 concepts | 1000 coverage rows                             |
| Query: concept:ab                                                            |
| Filter uses actual ConceptID text only; no labels are guessed                |
| ConceptID | PL | EN                                                          |
| concept:abdomen | approved | approved                                        |
| concept:ability | approved | approved                                        |
| concept:absence | pending | approved                                         |
| concept:abstract | pending | approved                                        |
| Matches: 4 | Page 1/1 | [Prev disabled] [Next disabled]                      |
| Coverage statuses are actual inventory fields, not acceptance                |
| Enter applies Query / selects ID by focus; details are a separate view       |
| No definitions, relations, or translations are inferred from an ID           |
| No-hit filter keeps the loaded catalog and its counts unchanged              |
| Tab/Shift-Tab moves focus; focus is visible without color                    |
| Up/Down selects; PageUp/PageDown scroll long result lists                    |
| Prompt: > shared shell input; it is not a catalog filter request             |
| F1-F5 global | Enter Query/Select | Esc clear selection | Ctrl-C exit        |
+------------------------------------------------------------------------------+
```

## 120 × 32 · full catalog viewport

With an empty query, the wider viewport shows the first eight actual rows from
the same 1000-entry response. `PageDown` continues through the local list; the
catalog service has no filter or paging arguments.

```text
+----------------------------------------------------------------------------------------------------------------------+
| GoLingua | F3 Concepts | full inventory list | RECORDED / DESIGN ONLY                                                |
| F1 Translate   F2 Forms   F3 Concepts   F4 Languages   F5 Settings                                                   |
|+--------------------------------------------------------------------------------------------------------------------+|
| F3 Concepts | filter actual ConceptID text; coverage comes from compiled catalog                                     |
| Focus: > Query | Tab -> result list -> scroll controls                                                               |
| Catalog: ok | 1000 concepts | target 1000 | 1000 returned rows                                                       |
| Query: (empty)                                                                                                       |
| Filter is a client-side substring match over returned ConceptID strings only                                         |
| ConceptID | PL | EN                                                                                                  |
| concept:abdomen | approved | approved                                                                                |
| concept:ability | approved | approved                                                                                |
| concept:absence | pending | approved                                                                                 |
| concept:abstract | pending | approved                                                                                |
| concept:accepting_offer | approved | approved                                                                        |
| concept:accessibility | approved | approved                                                                          |
| concept:accuracy | approved | approved                                                                               |
| concept:acidity | approved | approved                                                                                |
| Rows 1-8 of 1000 | [PageUp disabled] [PageDown available]                                                            |
| Current output shows approved and pending; no unobserved statuses are fabricated                                     |
| absent and non_transferable are valid enum values but absent from this run                                           |
| No-hit query not-a-concept-id -> 0 client-side matches; catalog remains loaded                                       |
| No concept label, definition, relation, or translation is inferred from its ID                                       |
| Enter applies Query / selects ID by focus; details are a separate view                                               |
| Tab/Shift-Tab: Query -> results -> scroll controls                                                                   |
| Up/Down selects a row | PageUp/PageDown scrolls the canonical result list                                            |
| Esc clears selection and preserves the query                                                                         |
| Prompt: > shared shell input; no filter or page argument is sent to Catalog                                          |
| F1-F5 navigation remains global; catalog coverage does not indicate runtime readiness                                |
| Resize recomputes visible rows from the same returned inventory                                                      |
| Ctrl-C exits and restores the terminal                                                                               |
+----------------------------------------------------------------------------------------------------------------------+
```

## Recorded CLI evidence

The catalog was built from implementation SHA
`35a3d667ff62af63cea6f0e535466736a50613ee`. Exact command, status, exit code,
stderr, full stdout byte count and SHA-256, plus sample rows are in the
[G11-12 evidence comment](https://github.com/clickbytes/golingua/issues/285#issuecomment-6361783032).
The [client-side filter observations](https://github.com/clickbytes/golingua/issues/285#issuecomment-6361795181)
record query `concept:ab` (four actual rows) and the no-hit query
`not-a-concept-id` (zero actual matches). These are local filters over the
returned inventory, not CLI/server filtering features.

Observed statuses in this CLI output are `approved` and `pending`. The enum in
`internal/catalog/count.go` also defines `absent` and `non_transferable`, but
neither occurred in this run; no mock row is fabricated for them. A `pending`
status remains a pending candidate, not an approved binding.

The empty-inventory state is **API-contract-only / NOT RUN**: this invocation
returned all 1000 CORE concepts. The mockup does not show an empty service
response as if it had been observed. If Catalog reports an error, keep its
returned status and diagnostics; do not replace the result with an empty
inventory. Catalog counts and coverage do not establish runtime readiness or
linguistic acceptance.

## Filtering, focus, and navigation

- `Query` accepts a stable ConceptID or fragment and performs a case-sensitive
  substring match against `ConceptCoverage.ConceptID` values already returned by
  the service. It does not search glosses, definitions, lexical forms, or
  relations. The service still returns the full inventory.
- The list preserves canonical CORE ID order. Each row shows only the ID and the
  actual `pl` and `en` coverage values. Pending is not styled or described as
  approved; focus and status remain understandable without color.
- Empty query means all returned IDs. A query with zero matches shows `No
  concepts match this ConceptID filter.` The catalog remains loaded and its
  inventory counts are unchanged. This is distinct from an empty service result.
- `Tab` / `Shift-Tab` visits Query, results, and scroll controls. `Enter` applies
  the local filter from Query; after moving focus to results, `Up` / `Down`
  selects a result and `Enter` retains the stable ID selection. `PageUp` /
  `PageDown` scrolls or pages the client-side list. Concept details are a separate
  view.
- `Esc` clears the current row selection while preserving the query. `F1`–`F5`
  remain global. The bottom prompt is part of the shared shell and is not sent as
  an argument to Catalog.
- Resize preserves the query and selected ID while recalculating the visible
  rows. No concept record is mutated.

This layout is for list/search only. Definition and relation fields belong to a
separate F3 detail view and appear only after an actual `LookupConcept` call.

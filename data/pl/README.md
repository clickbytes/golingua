# Polish editorial worklist — 1000 CORE meanings

`data/pl/1000-core-review.tsv` is a project-authored **editorial ledger**, not executable runtime data or an approved translation table. Exactly 1000 distinct CORE IDs are represented. Each row carries the proposed PL lemma, sense identity, mapping relationship, provenance note, valency frame reference where implemented, and morphology scope. An entry flagged `REVIEWED` describes an **agent's draft choice**, not linguistic acceptance by the owner.

## Worklist status

| Editorial state | Rows |
| --- | ---: |
| REVIEWED_LEXICALIZED | 965 |
| REVIEWED_BROADER_NARROWER | 18 |
| REVIEWED_RELATED | 2 |
| REVIEWED_NO_DIRECT_FORM | 15 |
| DEFERRED / UNREVIEWED | 0 |
| **Total** | **1000** |

Proposed semantic candidate rows: **983 pending**, **17 without a proposed binding**. This total is independent of the compiled `catalog` counts: **954 pending PL candidates have typed Go declarations**, and **29 event candidates remain worklist-only**, requiring individual lexical and valency review. `DEFERRED=0` means initial editorial triage is complete, not that every entry is executable or accepted. No mapping is `approved`.

## Compiled state after G07-03 batches (2026-10-22)

- CORE: **1000** compiled concept definitions.
- PL: **934 lexemes**, **955 senses**, **954 pending semantic bindings**, **46 CORE absent**.
- EN unchanged: **19 lexemes**, **28 senses**, **19 pending bindings**.
- PL valency: **247 authored, compiled frames**, some single-role, some NP/PP/recipient-based. No approved construction binding.
- Polish morphology: **15 selected noun lexemes** supplied with 14 explicit form strings each (7 cases × 2 numbers), using five shared gender-class selection paradigms. Previous `kot`, `student`, `czytać` seeds remain. No automatic inflection for the remaining lexical inventory.

`MORPHOLOGY_COVERAGE=attested_14_case_number_cells_pending_owner_review` in the TSV identifies the noun slice, including multiple senses of the same lexeme. `no_authorized_paradigm` identifies new verb entries without inflection. Missing cells and lexemes must not be filled through speculative analogy. Inverse analysis preserves distinct syncretic readings, with normalization constrained to NFC.

Selected spelling/semantic corrections were made explicitly in the worklist (such as `kaszleć`, `popychać`, `odblokowywać`, `zamarzać`, `topnieć`, `palić się`). These are candidate decisions, not approved lemma/sense transfers. Certain nominalized relation labels are metadata labels, not working prepositions or constructions. Disputed broad/narrow mappings need language review.

Batch 8 adds pending lexical senses and explicit accusative transitive frames for `otwierać`, `zamykać`, `łamać`, `podgrzewać`, `chłodzić`, `podpalać`, `gasić` and the battery sense of `ładować`. It corrects the battery lexeme ID to `pl.ladowac_baterie.v`, distinct from the existing `pl.ladowac.v` (`lądować`). No verb morphology is declared.

Batch 9 adds pending senses and role-specific frames for `informować`, `wyjaśniać`, `opisywać`, `przedstawiać`, `prosić`, `zapraszać`, `oferować` and `obiecywać`. The dative, accusative and selected prepositional complements are explicit; `rozkazywać` stays outside the compiled lexicon because its requested-action clause needs a construction-level representation. No new verb morphology or construction binding is declared.

Batch 10 adds pending senses for `porównywać`, `ranić`, `sadzić`, `ostrzegać`, `polecać`, `zaczynać`, `kończyć` and `kontynuować`, with selected case frames. Infinitival complements for the activity verbs are outside these frames. No new verb morphology or construction binding is declared.

Batch 11 adds six pending candidates: `kraść` (correcting the worklist typo), `kategoryzować`, `rysować`, `nagrywać`, `współpracować` and `ubierać się`, with selected case and adposition frames. Their bindings remain pending; no morphology or construction binding is declared.

Batch 12 adds seven pending senses: `zgadzać się`, `odmawiać`, `dyskutować`, `uznawać`, `dziedziczyć`, `wzrastać` and the quantity-change sense of `spadać`. It corrects the worklist form `rośnieć` to `wzrastać`; the existing motion sense of `spadać` stays separate. No morphology or construction binding is declared.

Batch 13 adds four pending senses: `skarżyć się`, `żartować`, `wolać` and `próbować`. Nominal arguments are recorded; clause complements for `próbować` remain outside the frame. No morphology or construction binding is declared.

Batch 14 adds four pending communication senses: `stwierdzać`, `nie zgadzać się`, `przyjmować` an offer, and `wysyłać` a message. The latter shares the existing `wysyłać` lexeme ID with goods shipping while keeping separate senses and frames. No morphology or construction binding is declared.

Batch 15 adds five pending senses for receiving a message, taking medicine, writing a document, boiling food and harvesting a crop. The message, writing and cooking senses reuse existing lexemes; no morphology or construction binding is declared.

Batch 16 adds four pending multiword event candidates: `nie lubić`, `czerpać przyjemność`, `dochodzić do siebie` and `robić zakupy`, each with a bounded case frame. Their multiword lemmas receive no inferred morphology or construction bindings.

## Licensing and manual review

`SourceRef` uses project-authored references and `License: unknown` where no explicit license decision exists. This is **not** permission to copy third-party dictionaries. There are no newly approved PL↔EN mappings; the owner-selected manual linguistic acceptance of these new batches remains **NOT RUN**.

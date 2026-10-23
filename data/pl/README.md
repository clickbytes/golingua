# Polish editorial worklist — 1000 CORE meanings

`data/pl/1000-core-review.tsv` is a project-authored **editorial ledger**, not executable runtime data or an approved translation table. Exactly 1000 distinct CORE IDs are represented. Each row carries the proposed PL lemma, sense identity, mapping relationship, provenance note, valency frame reference where implemented, and morphology scope. An entry flagged `REVIEWED` describes an **agent's draft choice**, not linguistic acceptance by the owner.

## Worklist status

| Editorial state | Rows |
| --- | ---: |
| REVIEWED_LEXICALIZED | 980 |
| REVIEWED_BROADER_NARROWER | 19 |
| REVIEWED_RELATED | 1 |
| REVIEWED_NO_DIRECT_FORM | 0 |
| DEFERRED / UNREVIEWED | 0 |
| **Total** | **1000** |

Proposed semantic rows: **882 approved and 118 pending**, **0 without a proposed binding**. This row count is independent of compiled bindings: **1006 typed binding records cover 1000 CORE concepts**, because six concepts each have two gender-specific lexical senses. `DEFERRED=0` means initial editorial triage is complete, not that every entry is executable or accepted. The 882 individually approved equivalents are recorded in delegated review issues #152, #154, #156, #158, #160, #162, #164, #166, #168, #170, #172, #174, #176, #178, #180, #182, #184, #186, #188, and #190 (batches 1–20), followed by six PL approvals in #192.

## Compiled state after pending audit #192 (2026-10-23)

- CORE: **1000** compiled concept definitions.
- PL: **982 lexemes**, **1007 senses**, **1006 semantic bindings (882 approved; 124 pending) over 1000 concepts**, **0 CORE absent**.
- EN: **1001 lexemes**, **1010 senses**, **996 approved and 5 pending binding records** over 1000 concepts.
- PL valency: **273 authored, compiled frames**, some single-role, some NP/PP/recipient-based; four separately approved construction bindings.
- Polish morphology: **17 selected noun lexemes** supplied with 14 explicit form strings each (7 cases × 2 numbers), using five shared gender-class selection paradigms. Previous `kot`, `student`, `czytać` seeds remain. No automatic inflection for the remaining lexical inventory.

`MORPHOLOGY_COVERAGE=attested_14_case_number_cells_pending_owner_review` in the TSV identifies the noun slice, including multiple senses of the same lexeme. `no_authorized_paradigm` identifies new verb entries without inflection. Missing cells and lexemes must not be filled through speculative analogy. Inverse analysis preserves distinct syncretic readings, with normalization constrained to NFC.

Selected spelling/semantic corrections were made explicitly in the worklist (such as `kaszleć`, `popychać`, `odblokowywać`, `zamarzać`, `topnieć`, `palić się`). The #192 correction from `slodycz` to `słodycz` has an approved dedicated sense. Each other lexical binding keeps its recorded status. Certain nominalized relation labels are metadata labels, not working prepositions or constructions. Disputed broad/narrow mappings need language review.

Batch 8 adds pending lexical senses and explicit accusative transitive frames for `otwierać`, `zamykać`, `łamać`, `podgrzewać`, `chłodzić`, `podpalać`, `gasić` and the battery sense of `ładować`. It corrects the battery lexeme ID to `pl.ladowac_baterie.v`, distinct from the existing `pl.ladowac.v` (`lądować`). No verb morphology is declared.

Batch 9 added pending senses and role-specific frames for `informować`, `wyjaśniać`, `opisywać`, `przedstawiać`, `prosić`, `zapraszać`, `oferować` and `obiecywać`. At that stage `rozkazywać` stayed outside the compiled lexicon because its requested-action clause needed a construction-level representation; batch 22 later added a bounded nominal frame (`rozkazywać komuś coś`) without licensing a clause construction. No new verb morphology or construction binding is declared.

Batch 10 adds pending senses for `porównywać`, `ranić`, `sadzić`, `ostrzegać`, `polecać`, `zaczynać`, `kończyć` and `kontynuować`, with selected case frames. Infinitival complements for the activity verbs are outside these frames. No new verb morphology or construction binding is declared.

Batch 11 adds six pending candidates: `kraść` (correcting the worklist typo), `kategoryzować`, `rysować`, `nagrywać`, `współpracować` and `ubierać się`, with selected case and adposition frames. Their bindings remain pending; no morphology or construction binding is declared.

Batch 12 adds seven pending senses: `zgadzać się`, `odmawiać`, `dyskutować`, `uznawać`, `dziedziczyć`, `wzrastać` and the quantity-change sense of `spadać`. It corrects the worklist form `rośnieć` to `wzrastać`; the existing motion sense of `spadać` stays separate. No morphology or construction binding is declared.

Batch 13 adds four pending senses: `skarżyć się`, `żartować`, `wolać` and `próbować`. Nominal arguments are recorded; clause complements for `próbować` remain outside the frame. No morphology or construction binding is declared.

Batch 14 adds four pending communication senses: `stwierdzać`, `nie zgadzać się`, `przyjmować` an offer, and `wysyłać` a message. The latter shares the existing `wysyłać` lexeme ID with goods shipping while keeping separate senses and frames. No morphology or construction binding is declared.

Batch 15 adds five pending senses for receiving a message, taking medicine, writing a document, boiling food and harvesting a crop. The message, writing and cooking senses reuse existing lexemes; no morphology or construction binding is declared.

Batch 16 adds four pending multiword event candidates: `nie lubić`, `czerpać przyjemność`, `dochodzić do siebie` and `robić zakupy`, each with a bounded case frame. Their multiword lemmas receive no inferred morphology or construction bindings.

Batch 17 adds six pending senses for watching a screen, imagining, hoping, receiving payment, refunding payment and transferring money. The selected case and adposition frames are explicit; no morphology or construction bindings are declared.

Batch 18 adds four verbal candidates for lending, ordering goods, typing and desiring, plus event nouns for locomotion and navigation. Only the verbal entries receive valency frames; the event nouns have no verbal valency. The TSV `giving` frame reference was synchronized with its existing compiled frame, with its binding still pending.

Batch 19 adds pending senses for locking, assembling, bargaining, gifting and sharing, plus an acquiring-knowledge sense on the existing `uczyć się` lexeme. All six verbal senses have selected frames; bindings remain pending and no morphology is added.

Batch 20 adds pending senses for `zachorować`, brushing teeth, exchanging goods and tissue healing. Their selected argument frames are explicit; no morphology or construction bindings are declared.

Batch 21 adds pending candidates for service hiring, misunderstanding (`źle zrozumieć`) and spoken interpreting. Their selected argument frames are explicit; no morphology or construction bindings are declared.

Batch 22 adds pending candidates for commanding, succeeding and failing. The two outcome senses use dative experiencer and nominative goal; no morphology or construction bindings are declared.

Batch 23 adds five pending multiword candidates already identified in the ledger: `osoba starsza`, `rodzic adopcyjny`, `członek zespołu`, `roślina kwitnąca` and `grunt rolny`. They have no inferred morphology, valency or construction bindings.

Batch 24 adds gendered lexical variants for stepparent, sibling, grandparent, grandchild, aunt/uncle and cousin. The paired senses remain separate pending narrow candidates; they share six CORE concepts. No morphology, valency or construction binding is added.

Batch 25 adds `małe dziecko` as a broader pending candidate for toddler and `przekazanie` as a narrower event noun for transfer. Neither adds morphology, valency or construction bindings.

Batch 26 adds pending category senses for `część ciała` and `zdarzenie`. These transparent labels have no inferred morphology, valency or construction bindings.

Batch 27 adds `celowe działanie` as a pending event-noun candidate for intentional action, replacing the overly broad worklist proposal `dokonywać`. No morphology, valency or construction binding is declared.

Batch 28 adds `komunikowanie` as a pending event noun for the sender-to-addressee communication event; the broader polysemous `komunikacja` remains noted as a related term. No morphology, noun valency or construction binding is declared.

Batch 29 adds `byt` as a pending, broad, register-limited candidate for the CORE continuant category `entity`, based on the philosophical noun sense documented by WSJP PAN. The external definition is not copied; the source's licensing remains unknown. No morphology, valency or construction binding is declared.

## Licensing and manual review

`SourceRef` uses project-authored references and `License: unknown` where no explicit license decision exists. This is **not** permission to copy third-party dictionaries. There are no newly approved PL↔EN mappings; owner approval of semantic bindings remains **NOT RUN**. Owner-directed Codex lexical assessment for the three cited candidates is recorded below; it does not include owner approval or CLI inflection/translation review.

Owner-directed lexical assessment (Codex, 2026-10-22) checked three pending candidates against WSJP PAN entries: `celowe działanie` matches the recorded sense of `celowy` and the collocation `celowe działanie`; `komunikowanie` is the gerund of `komunikować` (“convey information”) and fits the sender-to-addressee event; `byt` is documented in a philosophical sense as anything existing in some way, so its mapping to CORE `entity` is broad, register-limited and must remain pending. These checks assess lexical fit only: they do not approve bindings or validate inflection, noun argument realization, or sentence translations. External source licensing remains unknown, and no dictionary definitions or inflection tables were copied.

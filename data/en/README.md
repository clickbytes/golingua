# English lexical worklist — 1000 CORE meanings

`1000-core-review.tsv` is an editorial ledger for the exact 1000 compiled CORE IDs. Typed Go declarations in `languages/en` are the runtime source. The ledger keeps separate fields for lexical lineage, semantic binding status, provenance and the reason for each decision.

## Current scope (2026-10-23)

| Editorial lineage | Rows |
| --- | ---: |
| `EXISTING_PENDING` (original seed designation) | 19 |
| `REVIEWED_LEXICALIZED` (later authored rows) | 981 |
| `UNREVIEWED` | 0 |
| **Total** | **1000** |

The lineage field records where the candidate originated; it does not report its current semantic approval. The compiled snapshot has **1000 lexemes, 1009 senses and 1001 binding records**. Delegated review in #61 and the pending audit in #192 approved **996 equivalent links**. **Five records on four concepts** remain pending: `entity`, `child_of_person`, `aunt_uncle` (two gender-specific records) and `employer`. Their exact reasons are in the ledger and #192. No concept is absent from the candidate inventory.

Eight pre-existing senses remain unmapped: `en.doctor.academic`, `en.house.institution`, `en.school.fish`, `en.book.work`, `en.hand.worker`, `en.see.understand`, `en.look.appear`, and `en.get.understand`. A linked candidate does not imply that other senses of the same word are linked.

## Runtime snapshot

`catalog --summary --json` reports CORE 1000; PL 982 lexemes / 1007 senses / 1006 bindings with 882 approved concepts and 118 pending; EN 1000 lexemes / 1009 senses / 1001 bindings with 996 approved concepts and four pending; 880 concepts have approved links in both languages. The CLI also reports 273 PL and five EN valency frames. These counts are inventory facts, not a translation-quality score.

Sentence analysis and realization remain unsupported in the default runtime. `forms` serves its separately declared morphology seed. Owner review of actual sentence output remains NOT RUN.

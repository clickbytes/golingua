# English lexical worklist — 1000 CORE meanings

`1000-core-review.tsv` is an editorial ledger, not a runtime loader or an approved translation table. It contains the exact 1000 IDs from the compiled CORE snapshot and preserves CORE metadata for audit. Rows do not infer English words from `DEFINITION_EN`.

## Current scope (2026-10-22)

| Editorial state | Rows |
| --- | ---: |
| `EXISTING_PENDING` (original seed) | 19 |
| `REVIEWED_LEXICALIZED` (new batches) | 33 |
| `UNREVIEWED` | 948 |
| **Total** | **1000** |

The 19 `EXISTING_PENDING` plus 33 `REVIEWED_LEXICALIZED` rows correspond to 52 compiled candidate bindings, not approved equivalents. All bindings remain `pending`; no owner approval is recorded. The nine senses in the original seed that currently have no CORE candidate binding remain unresolved: `en.doctor.academic`, `en.house.institution`, `en.school.institution`, `en.school.fish`, `en.book.work`, `en.hand.worker`, `en.see.understand`, `en.look.appear`, and `en.get.understand`.

The authored increments add candidate nouns `person`, `adult`, `animal`, `plant`, `tree`, `artifact`, `door`, `vehicle`, `location`, `body part`, `event`, `spouse`, `romantic partner`, `friend`, `acquaintance`, `neighbor`, `stranger`, `guest`, `resident`, `citizen`, `visitor`, `tourist`, `traveler`, `parent`, `mother`, `father`, `sibling`, `brother`, `sister`, `grandparent`, `grandchild`, `cousin`, and `offspring`. These increments declare no morphology, valency, or construction coverage. Their draft lexical fit was assessed independently from Polish data; bindings remain pending and need owner review. `offspring` is a candidate for the age-independent child-of-person concept, separate from the existing young-person sense of `child`. Spot checks against Cambridge Dictionary distinguish `acquaintance` from `friend`, the person sense of `stranger`, and the leisure travel sense of `tourist`; no definition text is copied. The ledger records that US `neighbor`/`traveler` correspond to UK `neighbour`/`traveller`, and excludes the separate UK sports-team sense of `tourist`. `SourceRef` identifies project-authored Go declarations with license `unknown`; no third-party definitions or forms were copied.

`UNREVIEWED` means editorial work has not yet been performed. It is not a deferral or an assertion that English lacks a direct form. The ledger is therefore a complete inventory scaffold but **not** completion of English editorial review or G07-04. Future increments should replace these rows only after decisions are made, with senses, lexical coverage, morphology/valency notes, provenance, reviewer, date, and rationale recorded explicitly.

## Runtime snapshot

`catalog --summary --json` currently reports CORE 1000; PL 982 lexemes / 1007 senses / 1006 binding records / 0 absent; EN 52 lexemes / 61 senses / 52 pending binding records / 948 absent; 273 PL and 5 EN valency frames; 0 approved concepts and 0 approved pairs. EN absence is not evidence that a concept has no English expression.

Sentence analysis and realization remain unsupported. `forms` only serves its separately declared morphology seed, and neither this ledger nor pending mappings activate translation. Owner review of lexical choices and real CLI translation output remains NOT RUN.

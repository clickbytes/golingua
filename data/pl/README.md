# Polish review of the 1000 CORE concepts

`data/pl/1000-core-review.tsv` contains **1000** different CORE IDs, language-specific candidate lemmas and editorial rationales. It is an **editorial worklist only**: never compiled at runtime, never an approved bilingual dictionary, and not an automated quality oracle.

## Editorial inventory

| PL_WORK_STATUS | Rows |
| --- | ---: |
| REVIEWED_LEXICALIZED | 965 |
| REVIEWED_BROADER_NARROWER | 18 |
| REVIEWED_RELATED | 2 |
| REVIEWED_NO_DIRECT_FORM | 15 |
| DEFERRED / UNREVIEWED | 0 |
| **Total** | **1000** |

983 worklist rows name a proposed sense/binding; 17 have no binding. All 983 proposed connections are **pending**. The ledger's status `REVIEWED` means that an agent has written a candidate decision; it does NOT certify correctness by the project owner, an authoritative dictionary or an executable grammar. In particular, `MAPPING_KIND=equivalent` in TSV is not `lexicon.BindingApproved`.

## Current compiled Go scope

After adding the institution/food entity candidates:

- Typed Polish lexicon: **728 lexemes / 744 lexical senses / 743 pending bindings** (257 CORE absent).
- Worklist-only: **240 event concept proposals** without fully authored Polish sense-specific valency; do not claim these as runtime bindings.
- Compiled valency: **12** semantically role-aligned Polish frames; no approved construction bindings or sentence realization.
- Morpho-phonological coverage: small separate seed, not the 704-lexeme vocabulary; unimplemented cells have no generated forms.

`PL_LEMMAS`, `LEXEME_SENSE_IDS`, `MAPPING_KIND`, `BINDING_STATUS`, `VALENCY_FRAME_IDS`, `MORPHOLOGY_COVERAGE`, `CONSTRUCTION_COVERAGE` and provenance fields may refer to worklist-only candidates. Inspect `catalog --summary --json` to see compiled runtime data; the ledger alone cannot establish runtime coverage.

All new data is project-authored candidate metadata, sources have `unknown` licence unless a definite licence is recorded. `unknown` is not permission to copy any outside lexicon. Disputed forms (including inflection, homonymy, dialect and semantic narrowing) require a separate owner decision. `winogrono` is not pluralia tantum; `lody` in its food sense uses plural form. Nominalized relation labels are not functional adpositions or syntactic operators.

**Manual linguistic acceptance: NOT RUN.** Owner-supplied examples are needed before approving bindings or declaring any PL↔EN translation readiness.

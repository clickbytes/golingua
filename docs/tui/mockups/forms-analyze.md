# F2 Analyze mode layout

**DESIGN / RECORDED — NOT LIVE GO OUTPUT.** F2 Analyze calls the existing
surface-form operation and displays only returned `FormsResult` fields. It is
separate from F2 Generate; no forms are generated here and no reading is
selected as linguistically correct.

## 80 × 24 · Polish single reading

The input `dostaliśmy` returned one actual reading from the PL morphology seed.
Features below are grouped for the narrow view; the reading ID is compact here
and the full CLI stdout is linked in the evidence record.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · F2 Analyze · PL · recorded observation                            │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F2 Analyze · Surface form                                                    │
    │ Focus: surface · language and Analyze are Tab controls                       │
    │ Language: Polish (PL)                                                        │
    │ Surface draft: dostaliśmy                                                    │
    │ Result: ok · 1 returned reading · exit 0                                     │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ Reading 1 of 1 · focus is for inspection only                                │
    │ Lexeme: lexeme:pl.dostac.v · POS verb                                        │
    │ Cell: past_1pl_masculine_personal                                            │
    │ Features: tense=past · person=first · number=plural                          │
    │ Features: gender=masculine_personal · aspect=perfective                      │
    │ Features: mood=indicative · verb_form=finite · voice=active                  │
    │ ID: 5e63e1fc…311d05d6                                                        │
    │ Source: regular · paradigm pl_verb_perfective_finite                         │
    │ Source and generation trace are returned reading fields.                     │
    │ Analyze mode does not generate forms.                                        │
    │ Shared prompt draft: > dostaliśmy                                            │
    │ Analyze [Enter] · Enter submits from surface or action focus.                │
    │ Tab/Shift-Tab focus · F1–F5 global · Esc returns to surface                  │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32 · English multiple readings

The actual surface `read` returns eight readings in deterministic service order.
This view shows the first four and the focused reading details; the result list
scrolls independently. Focus identifies the current row for inspection only.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · F2 Analyze · EN · recorded observation                                                                    │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F2 Analyze · Surface form · ordered readings                                                                         │
    │ Language: English (EN) · focus: readings list · row 1 for inspection                                                 │
    │ Surface draft: read                                                                                                  │
    │ Result: ok · 8 returned readings · exit 0                                                                            │
    │ The focused row is not ranked as the best or linguistically approved.                                                │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Returned order · first 4 of 8 visible · list scrolls                                                                 │
    │ > 1 · imperative · lexeme:en.read.v · cell imperative                                                                │
    │    mood=imperative · verb_form=imperative · voice=active                                                             │
    │   2 · indicative past · lexeme:en.read.v · cell past_finite                                                          │
    │    mood=indicative · tense=past · verb_form=finite · voice=active                                                    │
    │   3 · past participle · lexeme:en.read.v · cell past_participle                                                      │
    │    verb_form=past_participle                                                                                         │
    │   4 · present first plural · lexeme:en.read.v · cell present_first_plural                                            │
    │    mood=indicative · person=first · tense=present · verb_number=plural                                               │
    │ Rows 5–8 continue in returned service order.                                                                         │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Focused reading details · row 1                                                                                      │
    │ ID: 54fa35ec2c799955aca8f5cfd92523e82828067aafc83b2586ccdbb4febbf501                                                 │
    │ POS: verb · source: regular · paradigm: en_verb_finite_principal_parts                                               │
    │ Features: mood=imperative · verb_form=imperative · voice=active                                                      │
    │ Provenance: golingua-en-morphology-seed · version 2026-10-23-g10-01                                                  │
    │ No reading is discarded; select/focus changes presentation only.                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Shared prompt draft: > read                                                                                          │
    │ Analyze [Enter] · surface submission uses Forms(Language, Surface).                                                  │
    │ Tab order visible · Esc returns to surface · F1–F5 global                                                            │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Recorded CLI states

All probes use implementation SHA
`34bfbe832cd7f5874e68161dfc7c288291eb7c05`, built with
`go build -o /tmp/golingua-g11-10-forms-34bfbe8 ./cmd/golingua`. PL/EN
morphology is bundled in the binary, so runtime snapshot/digest is **N/A**.
The [issue evidence comment](https://github.com/clickbytes/golingua/issues/281#issuecomment-6361471056)
contains each exact command, raw stdout, stderr, exit code, and status.

| Request | Observed result | Source / evidence |
| --- | --- | --- |
| PL surface `dostaliśmy` | `ok`, one reading; `lexeme:pl.dostac.v`, `past_1pl_masculine_personal`; exit 0 | Existing CLI example in [manual review](../../manual-review.md); current probe in [G11-10 evidence](https://github.com/clickbytes/golingua/issues/281#issuecomment-6361471056) |
| EN surface `read` | `ok`, eight actual ordered readings for `lexeme:en.read.v`; exit 0 | Existing `read/reads` observations in [sentence data README](../../../data/sentences/README.md); current probe in [G11-10 evidence](https://github.com/clickbytes/golingua/issues/281#issuecomment-6361471056) |
| EN surface `not` | `unknown`, zero readings, `UNKNOWN_SURFACE_FORM`; exit 5 | `not` occurs in existing negated sentence observations in [G10-17](../../../data/sentences/observed-2026-10-23-g10-17-present-negation.tsv); current probe in [G11-10 evidence](https://github.com/clickbytes/golingua/issues/281#issuecomment-6361471056) |
| Invalid language `xx`, surface `read` | `invalid`, `INVALID_FORMS_ARGUMENTS`; exit 2 | Current probe in [G11-10 evidence](https://github.com/clickbytes/golingua/issues/281#issuecomment-6361471056); this is request validation, not a linguistic judgement |

`unknown` means the current inverse morphology table returned no licensed
reading for that surface. It does not mean the surface is not a word. The
invalid request remains distinct from an unknown surface.

## Unavailable morphology contract state

The normal CLI always loads the bundled PL/EN morphology, so this failure could
not be produced through the recorded command and is **NOT RUN**. If the typed
service returns a loader failure, F2 shows the returned `error` and
`MORPHOLOGY_LOAD_FAILED` diagnostic with no readings. An unavailable inverse
index is represented by its returned `unknown` status and
`INVERSE_INDEX_UNAVAILABLE` diagnostic, not silently relabeled. These are
API-contract design states, **not recorded CLI output**; the mockup does not
invent a sample response.

## Status and interaction boundary

| Returned state | F2 behavior |
| --- | --- |
| `ok` with one or many readings | Preserve returned order and every real reading; show surface, lexeme, POS, cell and feature values. |
| `unknown` | Preserve diagnostics and show no reading rows; do not infer a lemma. |
| `invalid` | Preserve request diagnostics; do not treat it as an unknown word. |
| `error` / unavailable loader | Show only returned failure diagnostics; no readings. This path is API-contract-only above. |
| `resource_limit` or another actual status | Preserve its exact status and diagnostics; do not map it to `unknown`. |

The F2 workspace owns language, surface, Analyze action, and result-list focus.
The global prompt supplies the current surface draft. `Enter` submits only when
the surface or Analyze action owns focus; `↑`/`↓` move through actual reading
rows. `Esc` from a subview returns focus to the surface without changing it.
`Tab`/`Shift-Tab` follow the visible order; F1–F5 remain global navigation.
Focus remains visible without color, and resizing preserves the surface draft
and reading-list position. Polish diacritics are preserved as returned.

The service source is [`app.Service.Forms`](../../../internal/app/service.go);
`morph.Reading` fields are defined in
[`internal/morph/finite_table.go`](../../../internal/morph/finite_table.go).
The TUI and owner language acceptance remain **NOT RUN**.

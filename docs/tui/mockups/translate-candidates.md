# F1 candidate/details overlay layout

**DESIGN / RECORDED — NOT LIVE GO OUTPUT.** These frames use only candidate
fields present in repository recordings. Selection means keyboard focus for
inspection; it does not change the F1 result or mean linguistic approval.

The overlay presents `engine.TranslationAlternative` values already returned
inside `app.TranslationResponse.Result`. A response `Failure` stays in F1; the
overlay does not call `app.Service.Translate`, run a preview, or fill absent
fields.

## 80 × 24 · text-bearing candidate detail

Recorded S001 PL→EN has outer status `ambiguous`, no `result.text`, and four
candidates. This view shows two actual rows and the full ID of the focused
candidate. Typed assumption/evidence values are not serialized in this TSV, so
the detail says so instead of deriving values from a diagnostic code.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · F1 Translate · Candidate details · recorded                       │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ Candidate list → detail · recorded S001 PL→EN                                │
    │ Input: Nauczyciel widzi psa.                                                 │
    │ Outer: ambiguous · result.text absent · 4 candidates                         │
    │ Focus: candidate 1 detail · selected for inspection only                     │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ Candidates (2 of 4 visible; list scrolls)                                    │
    │ > 1 incomplete · The teacher sees the dog.                                   │
    │   ID: translation:bdb419…159acbc1                                            │
    │   2 incomplete · The teacher sees a dog.                                     │
    │   ID: translation:5436afb2…c6c73f67                                          │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ Detail · candidate outcome incomplete                                        │
    │ Text: The teacher sees the dog.                                              │
    │ Diagnostic: EXPLICIT_REALIZATION_ASSUMPTION                                  │
    │ Full candidate ID:                                                           │
    │ translation:bdb419f9a32f7583968a651331624dd2778dd5ed2618f9cd10a9649b159acbc1 │
    │ Typed assumptions/evidence: not captured in this source TSV.                 │
    │ Esc back to list; Esc again to F1; F1 and prompt state retained.             │
    │ Tab focus · F1–F5 global · Preview selection ≠ linguistic approval           │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32 · candidate without text

Recorded S004 PL→EN has outer status `ambiguous` and two candidates. The focused
candidate has outcome `incomplete` but no text; its ID and diagnostics are still
shown. The second actual candidate has text, as recorded. `Esc` returns one
level and preserves F1's prompt/result state.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · F1 Translate · Candidate details · recorded historical example                                            │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Candidate list → detail · recorded S004 PL→EN                                                                        │
    │ Input: Nauczyciel widzi dom.                                                                                         │
    │ Outer: ambiguous · result.text absent · 2 actual candidates                                                          │
    │ Focus: candidate 1 detail · selection is for inspection only                                                         │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Candidates (2 actual rows)                                                                                           │
    │ > 1 incomplete · no candidate text                                                                                   │
    │   ID: translation:6f9c90343458ce2ab5575811180158db8c8dae519fea20d9e1d12217bfe313bf                                   │
    │   2 incomplete · text is present in the source recording                                                             │
    │   ID: translation:c26fd3031a82129147748344db9c3562697225f70630657a4d673c2838fed6bc                                   │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Detail · candidate outcome incomplete                                                                                │
    │ Candidate ID:                                                                                                        │
    │ translation:6f9c90343458ce2ab5575811180158db8c8dae519fea20d9e1d12217bfe313bf                                         │
    │ Text: none returned (source TSV value: null)                                                                         │
    │ Candidate diagnostic: PRINCIPAL_PARTS_UNAVAILABLE                                                                    │
    │ Outer diagnostic: DISAMBIGUATION_INCOMPLETE;PRINCIPAL_PARTS_UNAVAILABLE                                              │
    │ Preview is unavailable because this candidate has no text.                                                           │
    │ Assumptions/evidence: not captured in this source TSV row.                                                           │
    │ Source: observed-2026-10-23-g10-07-runtime-preview.tsv                                                               │
    │ Record: S004 · implementation SHA 0da6ad7726af85af6ac92fade774061d3c4e7815                                           │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ Esc returns to list; Esc again returns to F1.                                                                        │
    │ F1 request, prompt draft, and list position are preserved.                                                           │
    │ Focus is labeled in text; never rely on color alone.                                                                 │
    │ F1–F5 remain global navigation · no candidate is approved                                                            │
    │ Recorded design state only · NOT LIVE GO OUTPUT                                                                      │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Recorded examples and field boundaries

| Record | Outer result | Candidate fields shown | Source |
| --- | --- | --- | --- |
| S001 PL→EN | `ambiguous`, exit 6, `result.text` absent, 4 candidates | First two actual candidates have outcome `incomplete`, actual text and IDs; candidate diagnostic `EXPLICIT_REALIZATION_ASSUMPTION`. | [G10-20 full corpus](../../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv), implementation SHA `a07c6cdecf4fbbcf7e72c5ca08881ca213050627` |
| S003 EN→PL | `incomplete`, exit 7 | Candidate `Nauczyciel widzi książkę.` · ID `translation:1909abbbfcf09557a6b0bed33553cf8dd81268e4b135c6eeaac45d7be0719f8f`; preserves Polish diacritics when wrapped. | [G10-20 full corpus](../../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv), implementation SHA `a07c6cdecf4fbbcf7e72c5ca08881ca213050627` |
| S004 PL→EN | `ambiguous`, 2 candidates, `result.text` absent | Candidate `translation:6f9c90343458ce2ab5575811180158db8c8dae519fea20d9e1d12217bfe313bf`: outcome `incomplete`, text `null`, diagnostic `PRINCIPAL_PARTS_UNAVAILABLE`. A second candidate has a text value in the source recording. | [G10-07 runtime preview](../../../data/sentences/observed-2026-10-23-g10-07-runtime-preview.tsv), implementation SHA `0da6ad7726af85af6ac92fade774061d3c4e7815` |

`TranslationAlternative` has typed `Evidence` and `Assumptions` fields in
[`internal/engine/result.go`](../../../internal/engine/result.go). The cited
TSV rows record candidate IDs, outcomes, text and diagnostic codes, but do not
serialize those typed values. In these two details their values are therefore
**not captured**. A diagnostic named `EXPLICIT_REALIZATION_ASSUMPTION` is shown
as a diagnostic, not converted into an assumed field/value/reason. Other saved
manual-review summaries belong to their own recorded implementation SHAs and
must not be joined to these candidate IDs.

| Status at the boundary | Overlay behavior |
| --- | --- |
| Outer `ok` with a returned candidate | Show only the returned ID, outcome and text. |
| Outer `ambiguous`; candidate outcome `incomplete`, with text | Preserve outer `ambiguous`; show the actual candidate text and ID. Do not select automatically. |
| Candidate outcome `incomplete`, no text | Show the actual ID and diagnostics with “no text”; preview is unavailable. |
| `unsupported`, `unknown`, `invalid`, `resource_limit`, or `error` response/failure | Keep the returned result/failure in F1. Show no candidate row unless an actual candidate was returned. |

The detail view separates outer status from candidate outcome. If a later
implementation offers a technical preview, it must use the selected candidate
ID explicitly under G11-29; that action is not owner or linguistic approval.

## Focus and resizing

The overlay owns its candidate-list/detail focus while open. `↑`/`↓` move among
actual rows; `Enter` opens the focused detail; `Esc` closes detail to the list,
then returns to F1 without changing its request, prompt draft or scroll position.
`Tab`/`Shift-Tab` follow the visible focus order. F1–F5 remain global screen
navigation as defined by the [keymap](../keymap.md). Focus is labeled in text,
not color alone. Full IDs and Polish text wrap or scroll without changing their
bytes; resizing preserves the selection and the F1 state.

Candidate details belong to G11-09. Preview execution belongs to G11-29 and the
reusable candidate details UI to G11-26; neither is implemented by this design.

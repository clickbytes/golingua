# F1 Translate screen layout

**DESIGN / NOT LIVE GO OUTPUT — recorded examples and UI-only states.**
This screen uses the existing `app.Service.Translate` result. Input comes from
the shared prompt; the screen displays only returned status, text, candidates
and diagnostics. It never chooses an ambiguous candidate automatically.

## 80 × 24 · recorded simple result

S001 shows an actual `ok` EN→PL observation. The displayed text and compact ID
are from the cited recording, not a new CLI run; the ID abbreviation stands for
the full recorded candidate ID listed in the source file.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: recorded · F1 selected · Focus: prompt                   │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F1 Translate · EN → PL · recorded S001                                       │
    │ Input (shared prompt): The teacher sees the dog.                             │
    │ Result status: ok · exit 0                                                   │
    │ Text: Nauczyciel widzi psa.                                                  │
    │ Candidate: ok · ID translation:301b34…99e4e                                  │
    │ Source: g10-20-full-corpus.tsv · SHA a07c6cdecf4f…                           │
    │ The text is present because this recorded status is ok.                      │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ PROMPT · F1 context · focus=input                                            │
    │ > The teacher sees the dog.                                                  │
    │ Editing is separate from submission; Enter sends to F1.                      │
    │ The shared prompt stays below the result workspace.                          │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                 │
    │ Enter Submit · Esc Back · Ctrl+C Exit                                        │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32 · recorded ambiguous result

S001 PL→EN returned four actual candidates and no selected `result.text`. The
first two candidate texts and full IDs are shown; the result-list area scrolls
for the remaining candidates. This preserves `ambiguous` even though candidates
have their own `incomplete` status.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: recorded · F1 selected · Focus: result list                                                      │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F1 Translate · PL → EN · recorded S001                                                                               │
    │ Input (shared prompt): Nauczyciel widzi psa.                                                                         │
    │ Result: ambiguous · exit 6 · result.text is absent                                                                   │
    │ Diagnostic: TARGET_ALTERNATIVES_AMBIGUOUS · four actual candidates                                                   │
    │ No candidate is selected automatically; 2 of 4 are visible here.                                                     │
    │ 1 · The teacher sees the dog. · candidate status incomplete                                                          │
    │    Candidate ID: translation:bdb419f9a32f7583968a651331624dd2778dd5ed2618f9cd10a9649b159acbc1                        │
    │ 2 · The teacher sees a dog. · candidate status incomplete                                                            │
    │    Candidate ID: translation:5436afb2822fad3cf56c0dc3aa7ffdca8ac6cc76677e6d0a80e88eccc6c73f67                        │
    │ Candidates 3–4 continue in the independently scrolling result list.                                                  │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ PROMPT · F1 context · focus remains available                                                                        │
    │ > Nauczyciel widzi psa.                                                                                              │
    │ Enter submits to F1; result text is separate from the prompt draft.                                                  │
    │ The screen reports only fields returned by Translate.                                                                │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                                                         │
    │ Enter Submit · Esc Back · Ctrl+C Exit                                                                                │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Busy and runtime unavailable states

Busy is UI state only; it is not an engine status or result. Keep the input
visible and leave result text and candidates blank until the service returns:

    F1 Translate · REQUEST IN PROGRESS
    Waiting for app.Service.Translate · no result received yet

When no runtime is selected, show that availability state and do not fill the
result panel with guessed text. The recorded `unsupported` sample below has no
output text and the diagnostic `LANGUAGE_REGISTRY_UNAVAILABLE`.

## Recorded status examples

All rows below are actual repository observations. Historical rows are labeled
as such; the mockup does not imply they match a current runtime. `-`/`null`
means no result text was returned.

| Status | Recorded request | Actual result/candidate | Source |
| --- | --- | --- | --- |
| `ok` | S001 EN→PL, `The teacher sees the dog.` | `Nauczyciel widzi psa.` · candidate ID `translation:301b34e2c5651449b8f9ae741938e7c5f2ec8fcdb02ea39c748a22615dd99e4e` · outcome `ok` | [G10-20 full corpus](../../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv), SHA `a07c6cdecf4f…` |
| `ambiguous` | S001 PL→EN, `Nauczyciel widzi psa.` | `result.text` absent; four real candidates, including the two shown above | [G10-20 full corpus](../../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv), SHA `a07c6cdecf4f…` |
| `incomplete` | S003 EN→PL, `The teacher sees the book.` | Candidate `Nauczyciel widzi książkę.` · ID `translation:1909abbbfcf09557a6b0bed33553cf8dd81268e4b135c6eeaac45d7be0719f8f` · `PARTIAL_PARSE_EVIDENCE;UNRESOLVED_ROLE_SENSE` | [G10-20 full corpus](../../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv), SHA `a07c6cdecf4f…` |
| `unsupported` | Historical S001, PL→EN | No text · `LANGUAGE_REGISTRY_UNAVAILABLE` | [recorded registry-unavailable run](../../../data/sentences/observed-2026-10-23-d05212e.tsv) |
| `unknown` | Historical S043, PL→EN | No text · `NO_SEMANTIC_ALTERNATIVES;NO_START_PLAN;UNKNOWN_SURFACE` | [initial provider run (pre-#197)](../../../data/sentences/observed-2026-10-23-provider-runtime.tsv) |
| `invalid` | S036 PL→EN, refused preview ID | No result text · `INVALID_PREVIEW_ID` | [explicit preview observation](../../../data/sentences/observed-2026-10-23-g10-06-explicit-preview.tsv), SHA `0da6ad7726af…` |

`invalid` here is the recorded status of a rejected candidate preview ID, not a
judgment that the sentence itself is invalid. `unknown` and `unsupported` are
also distinct: the former records no semantic start plan; the latter records
an unavailable language registry in its cited observation.

## Layout boundary

The F1 workspace owns direction, returned result presentation and candidate
list scrolling. The shared prompt owns typed input and focus; F1 submits it to
`Service.Translate`. Candidate detail/preview overlays are deferred to G11-09.
Runtime-unavailable, busy and recorded-result labels must not be combined into
a fabricated live response.

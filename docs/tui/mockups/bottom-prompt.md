# Fixed bottom prompt layout

**DESIGN / NOT LIVE GO OUTPUT — illustrative layout states, not a live TUI.**
The prompt is one shell component below the independently scrolling workspace.
It is not repeated in F1–F5 and does not call the engine by itself. The active
screen owns the operation that follows an explicit submission.

## 80 × 24 · one-line draft

The prompt has focus. Its history preview is read-only; typing or pasting edits
the draft. The visible IME pre-edit string is separate from committed draft
text, so Enter cannot submit text while composition is still active.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: unavailable · F1 Translate selected · Focus: prompt      │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE · scrolls independently above this fixed prompt                    │
    │ The selected screen owns content; no engine result is represented.           │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ PROMPT HISTORY · one recent draft · session preview                          │
    │ 1 · Earlier draft: Cześć, Łódź — 😊                                           │
    │ PROMPT · focus=input · one line · editing draft                              │
    │ > Cześć, Łódź — 😊▏                                                           │
    │ IME pre-edit: か · uncommitted; not part of the draft yet                     │
    │ Typing or paste edits only; the draft is not submitted.                      │
    │ Enter submits after composition ends · ? shows help                          │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                 │
    │ Enter Submit · Esc Back · Ctrl+C Exit                                        │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32 · bounded multiline draft

This view shows two pasted lines in a four-line prompt viewport, with a small
read-only history preview. IME composition remains visible as pre-edit text;
Enter commits that text first, and a later Enter submits the draft.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: unavailable · F1 Translate selected · Focus: prompt                                              │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE · scrolls independently above this fixed prompt                                                            │
    │ Screen-owned content stays separate from the shared prompt.                                                          │
    │                                                                                                                      │
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
    │ PROMPT HISTORY · two recent drafts · session preview                                                                 │
    │ 1 · The teacher reads the message.                                                                                   │
    │ 2 · Cześć, Łódź — 😊                                                                                                  │
    │ PROMPT · focus=input · bounded multiline · draft 2/4 lines                                                           │
    │ > Cześć, Łódź — first pasted line                                                                                    │
    │   Second pasted line stays in the same draft.▏                                                                       │
    │ IME pre-edit: か · composition active · not yet committed                                                             │
    │ Enter while composing commits text; it does not submit.                                                              │
    │ After commit, a later Enter submits to the active screen.                                                            │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                                                         │
    │ Enter Submit · Esc Back · Ctrl+C Exit                                                                                │
    │ Typing and paste edit the draft; submission is explicit.                                                             │
    │ The slash palette is a separate overlay above this prompt (G11-07).                                                  │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

The multiline display demonstrates pasted line breaks; it does not define a
new newline shortcut or change the routing in the [prompt contract](../command-contract.md).

## Overflow and resize

The prompt uses at most four visible draft rows. If a draft is longer, scroll
inside that reserved prompt area and retain every line; the workspace has its
own scroll position and the keyboard footer stays fixed. The history preview
can collapse to a one-row summary before reducing the draft viewport.

Overflow preview:

    Prompt viewport · 4 visible rows · showing draft lines 3–6 of 6
    History preview · collapsed to one row · footer remains visible
    Workspace scroll and prompt scroll are independent

On resize from 120×32 to 80×24, recompute the prompt and workspace viewports
without discarding draft text, caret position, history preview, focus or IME
pre-edit text. Below 80×24, use the compact warning in the [global shell](global-shell.md)
and preserve the draft state until the terminal is enlarged.

## Component boundary

The shell prompt owns its draft, edit focus, bounded viewport, history preview
and IME composition display. History is session-only and is not persisted. The active screen owns routing and the operation
performed on submission. The palette remains a separate overlay designed in
G11-07. All states here are layout illustrations, not engine results or a
change to the existing command contract.

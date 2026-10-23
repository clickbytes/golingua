# Slash palette layout

**DESIGN / NOT LIVE GO OUTPUT — illustrative overlay states, not a live TUI.**
The palette is a separate navigation overlay immediately above the persistent
prompt. It filters the existing registry only: `/translate`, `/forms`,
`/concepts`, `/languages`, `/settings`. The selected row is marked in text;
keyboard focus and selection are both understandable without color.

## 80 × 24 · partial prefix

The `/fo` prefix produces one highlighted match. The palette stays above the
prompt, and the active F1 workspace remains behind it.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: unavailable · F1 selected · Focus: palette query         │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE · active screen continues behind the overlay                       │
    │ Its content scrolls independently; the prompt and footer stay fixed.         │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ PALETTE · prefix /fo · 1 result · query has focus                            │
    │ > /forms · Navigate to F2 Forms                                              │
    │ The highlight is a selection; it does not activate by itself.                │
    │ Results 1–1 of 1 · Enter chooses · Esc cancels                               │
    │ ↑↓ changes selection when there is more than one result.                     │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ PROMPT · remains visible below the palette · F1 context                      │
    │ > /fo                                                                        │
    │ Typing filters only; no command runs until chosen.                           │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                 │
    │ ↑↓ Select · Enter Choose · Esc Back                                          │
    │ Ctrl+C Exit                                                                  │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32 · empty prefix

The initial `/` query shows all five existing navigation commands. The first
result is selected, but navigation occurs only after `Enter`.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua · Runtime: unavailable · F1 selected · Focus: palette query                                                 │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE · active screen continues behind the overlay                                                               │
    │ The palette occupies only the space above the persistent prompt.                                                     │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    │                                                                                                                      │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ PALETTE · query / · five registered navigation commands · query has focus                                            │
    │ > /translate · Navigate to F1 Translate                                                                              │
    │   /forms · Navigate to F2 Forms                                                                                      │
    │   /concepts · Navigate to F3 Concepts                                                                                │
    │   /languages · Navigate to F4 Languages                                                                              │
    │   /settings · Navigate to F5 Settings                                                                                │
    │ Selected item 1/5 · all five results fit in this viewport                                                            │
    │ ↑↓ Select · Enter Choose · Esc Cancel                                                                                │
    │ Prefix matching is literal; no fuzzy results are added.                                                              │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ PROMPT · visible below palette · current query draft: /                                                              │
    │ > /                                                                                                                  │
    │ The query remains editable; typing does not navigate.                                                                │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens · Tab Focus · ←→ Tabs · ? Help                                                                         │
    │ ↑↓ Select · Enter Choose · Esc Back · Ctrl+C Exit                                                                    │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Exact, empty and invalid queries

- Exact `/forms` is ready to choose, but does not navigate until `Enter`:

      PALETTE · exact match · focus=query
      > /forms · ready; Enter to navigate

- No match `/xyz` keeps the overlay open, preserves the query and makes no
  service call or screen change:

      PALETTE · query /xyz · focus=query
      No registered commands match · query retained

- Invalid `/translate --from en` shows an input error and preserves the full
  text for correction. Do not pass it to the CLI or engine:

      PALETTE · invalid command · focus=query
      /translate --from en
      Arguments are not supported · text retained

- On successful navigation, close the palette, activate the selected screen and
  clear only the command text as specified by the
  [command contract](../command-contract.md).
- `Esc` cancels the palette action and restores the previous focus; it does not
  navigate. Up/Down changes the highlighted result when multiple results exist.

## Clipping and scroll

At 80×24, constrain the overlay to the rows above the prompt and footer. When
five results do not fit, scroll within the palette list and keep the query,
selected result, prompt and footer visible:

    PALETTE · query / · focus=query · results 1–3/5 · selected 2/5
      /translate
    > /forms
      /concepts
    Remaining results scroll inside the overlay · prompt/footer stay fixed

Do not let the overlay cover the prompt or global footer. Keep query focus and
the highlighted result distinguishable using text and the `>` marker.

## Ownership boundary

The palette owns its temporary query, result list, selected index and scroll
position. The persistent prompt remains a separate shell component below it;
the active screen owns the command's navigation destination. All command names
come from the existing registry. This artifact adds no slash verbs, engine
operation or live behavior.

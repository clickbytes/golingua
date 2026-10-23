# Global TUI shell mockup

**DESIGN / NOT LIVE GO OUTPUT — illustrative only, not a live TUI.** The shell
frames below reserve the current screen's workspace and the shared prompt. The
mockup contains no translation result, language data, or business-screen
controls. The selected tab is bracketed; keyboard focus has a separate textual
marker and does not depend on color.

## 80 × 24

The active tab is F1 Translate; keyboard focus is on its tab. Runtime is
unavailable. The workspace is intentionally neutral.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua                         Runtime: unavailable · Focus: F1 tab        │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE — scrollable active screen content starts here                     │
    │ Screen-specific content is defined by separate mockups.                      │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    │                                                                              │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ Prompt slot — fixed bottom control; component layout belongs to G11-06       │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens  Tab Focus  ←→ Tabs  Enter Select  Esc Back ? Help  Ctrl+C Exit│
    │                                                                              │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32

The same shell uses the extra width for breathing room. Tab order, prompt
ownership and workspace ownership remain the same; later mockups define the
contents of individual screens.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua                         Runtime: unavailable · Focus: F1 tab                                                │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE — scrollable active screen content starts here                                                             │
    │ Screen-specific content is defined by separate mockups.                                                              │
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
    │ Prompt slot — fixed bottom control; component layout belongs to G11-06                                               │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ F1–F5 Screens  Tab Focus  ←→ Tabs  Enter Select  Esc Back ? Help  Ctrl+C Exit                                        │
    │                                                                                                                      │
    └──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Layout rules

- Keep the title/runtime status, all F1–F5 tabs, workspace, prompt slot and
  keyboard footer visible at 80×24.
- At wider sizes, let only the active screen's workspace expand. Do not add
  another navigation mode or infer new engine capability.
- The prompt slot remains bottom-anchored; G11-06 defines its internal fields,
  history and focus treatment.
- The footer labels keyboard access. The selected tab and current focus remain
  distinguishable without color.
- Below 80×24, use the compact-terminal behavior in the product contract and
  preserve the explicit request to enlarge the terminal when status or focus
  cannot remain visible.

## Compact and long-header states

At a size below the usable minimum, replace the workspace with an explicit
compact warning. Keep the selected screen, runtime status, keyboard focus and
draft in state; never silently clip their values. This 60×6 inset illustrates
the insufficient-size state rather than adding another screen:

    60×6 · Runtime: unavailable
    Selected tab: F1 Translate · Focus: prompt
    Draft: “The teacher reads the message.” (retained)
    Workspace paused; status, focus and draft are preserved.
    Enlarge terminal to at least 80×24 to continue.
    After resize, restore the selected screen and focus.

At 80 columns, a long runtime path wraps at a path separator. The selected tab
and the keyboard-focus marker remain on their own readable header row, while
the complete path is retained:

    GoLingua · Runtime: /home/ppotepa/golingua/config/runtime/production/
               2026-10-23/pl-en.json
    Selected tab: [F1 Translate] · Focus: F1 tab
    Draft: “The teacher reads the message.” · Status: ready

The first row is the start of a long header, not a second runtime field. Do not
truncate its path or hide the selected tab, focus, draft or result status when
wrapping consumes a row.

## Component ownership

- The shell owns the header/runtime indicator, F1–F5 tabs, workspace slot,
  bottom prompt slot and keyboard footer.
- Each F1–F5 screen owns only its content inside the workspace slot.
- The command palette is a separate overlay above the prompt and is designed
  in G11-07; this mockup does not render it or the prompt's internal fields.

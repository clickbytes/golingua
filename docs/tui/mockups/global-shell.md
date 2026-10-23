# Global TUI shell mockup

**Design artifact only — not captured Go output and not a live TUI.** The shell
frames below reserve the current screen's workspace and the shared prompt. The
mockup contains no translation result, language data, or business-screen
controls. Focus is labeled in text and shown on the selected tab, not by color.

## 80 × 24

The active tab is F1 Translate; focus is on the screen tabs. Runtime is
unavailable. The workspace is intentionally neutral.

    ┌──────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua                         Runtime: unavailable · Focus: screen tabs   │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings             │
    ├──────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE — active screen content starts here                                │
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
    │ F1–F5 Screens  Tab Focus  ←→ Tabs  Enter Select  Esc Back  Ctrl+C Exit       │
    │                                                                              │
    └──────────────────────────────────────────────────────────────────────────────┘

## 120 × 32

The same shell uses the extra width for breathing room. Tab order, prompt
ownership and workspace ownership remain the same; later mockups define the
contents of individual screens.

    ┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │ GoLingua                         Runtime: unavailable · Focus: screen tabs                                           │
    │ [F1 Translate]  F2 Forms  F3 Concepts  F4 Languages  F5 Settings                                                     │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
    │ WORKSPACE — active screen content starts here                                                                        │
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
    │ F1–F5 Screens  Tab Focus  ←→ Tabs  Enter Select  Esc Back  Ctrl+C Exit                                               │
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

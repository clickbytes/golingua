# F2 Generate mode layout

**DESIGN / RECORDED — NOT LIVE GO OUTPUT.** This layout submits only the explicit
language, lexeme ID, and language-local `key=value` features to
`app.Service.Forms`. It displays the returned `FormsResult` without inventing
paradigms, defaults, readings, or linguistic approval. The TUI is not yet
implemented.

## 80 × 24 · Polish generated form

The frame uses the actual PL request recorded below. Feature rows are plain text;
there are no guessed dropdown inventories. The compact result shows the returned
surface, status, state, POS, profile, paradigm/cell, source, features, and a
trace summary. `Details` exposes the remaining returned provenance and trace
fields without changing the result.

```text
+------------------------------------------------------------------------------+
| GoLingua | F2 Generate | Polish (PL) | RECORDED / DESIGN ONLY                |
| F1 Translate   F2 Forms   F3 Concepts   F4 Languages   F5 Settings           |
|+----------------------------------------------------------------------------+|
| F2 Generate | submit a lexeme and explicit Polish feature values             |
| Focus: > Lexeme ID | Tab -> features -> Add/Remove -> Generate               |
| Language: pl | Lexeme ID: lexeme:pl.widziec.v                                |
| Features: type language-local key=value pairs; no guessed options            |
| verb_form=finite | tense=present | aspect=imperfective | mood=indicative     |
| voice=active | person=third | number=plural                                  |
| Feature rows: [ Add ] [ Remove ] | Actions: [ Generate ] [ Details ]         |
|+----------------------------------------------------------------------------+|
| Result: ok | state=generated | 1 returned form | exit 0                      |
| Surface: widzą                                                               |
| Language: pl | POS: verb | Profile: v2026-10-23.1                            |
| Paradigm / cell: pl_verb_present_i / present_3pl                             |
| Source: regular | Provenance ID: golingua-pl-morphology-seed                 |
| Returned features: aspect=imperfective, mood=indicative, number=plural       |
| person=third, tense=present, verb_form=finite, voice=active                  |
| Trace summary: select_stem (4) -> append (6)                                 |
| Details shows all returned provenance and trace fields                       |
| Prompt: > (shared shell input; separate from F2 field draft)                 |
| Tab/Shift-Tab focus | Enter Generate | Esc Back | F1-F5 tabs | Ctrl-C Exit   |
+------------------------------------------------------------------------------+
```

## 120 × 32 · English generated form

This wider view uses an actual EN request. All returned forms stay visible in
service order. If an output has many forms, the results area scrolls and
`Up`/`Down` inspects them; the interface does not choose a preferred form.

```text
+----------------------------------------------------------------------------------------------------------------------+
| GoLingua | F2 Generate | English (EN) | RECORDED / DESIGN ONLY                                                       |
| F1 Translate   F2 Forms   F3 Concepts   F4 Languages   F5 Settings                                                   |
|+--------------------------------------------------------------------------------------------------------------------+|
| F2 Generate | submit a validated lexeme ID and explicit English feature values                                       |
| Focus: > feature row 1 | Tab -> next input; focus visible without color                                              |
| Language: English (en)                                                                                               |
| Lexeme ID: lexeme:en.look.v                                                                                          |
| Feature editor: bounded key=value rows; service validates language profile values                                    |
| verb_form=finite | tense=present | mood=indicative                                                                   |
| voice=active | person=third | verb_number=singular                                                                   |
| Feature rows: [ Add ] [ Remove ] | Actions: [ Generate ] [ Details ]                                                 |
|+--------------------------------------------------------------------------------------------------------------------+|
| Result: ok | state=generated | 1 returned form | exit 0                                                              |
| Surface: looks                                                                                                       |
| Lexeme: lexeme:en.look.v | Language: en | POS: verb                                                                  |
| Profile version: v2026-10-23.1                                                                                       |
| Paradigm / cell: en_verb_finite_principal_parts / present_third_singular                                             |
| Source: regular                                                                                                      |
| Provenance ID: golingua-en-morphology-seed                                                                           |
| Provenance path: languages/en/morphology/paradigms.json                                                              |
| Provenance version: 2026-10-23-g10-01 | license: internal                                                            |
| Returned features: mood=indicative, person=third, tense=present                                                      |
| verb_form=finite, verb_number=singular, voice=active                                                                 |
| Trace summary: select_stem applied=true | output_size=5                                                              |
| Details opens exact GeneratedForm provenance and trace fields.                                                       |
| No ranking is applied; all returned forms remain visible in engine order.                                            |
| Tab/Shift-Tab moves focus; Up/Down inspects returned forms.                                                          |
| Esc returns from details without clearing the submitted request.                                                     |
| Prompt: > (shared shell input; separate from F2 field draft).                                                        |
| Tab focus | Enter Generate | Esc Back | F1-F5 tabs | Ctrl-C Exit                                                     |
+----------------------------------------------------------------------------------------------------------------------+
```

## Recorded CLI evidence

The observations were captured from implementation SHA
`15dec7bb604cefb5425671ae82cdbbc528e246cf`, using binary
`/tmp/golingua-g11-11-15dec7bb` built with
`go build -o /tmp/golingua-g11-11-15dec7bb ./cmd/golingua`. The exact argv,
stdout, stderr, and exit codes are in the [G11-11 evidence comment](https://github.com/clickbytes/golingua/issues/283#issuecomment-6361599808).
Additional probes, including unsupported features and no-matching-cell requests,
are recorded in the [follow-up evidence comment](https://github.com/clickbytes/golingua/issues/283#issuecomment-6361679655).

| Request | Observed result |
| --- | --- |
| PL `lexeme:pl.widziec.v`, `verb_form=finite tense=present aspect=imperfective mood=indicative voice=active person=third number=plural` | `ok`, `state=generated`, `widzą`, exit 0; one returned form, cell `present_3pl`. |
| EN `lexeme:en.look.v`, `verb_form=finite tense=present mood=indicative voice=active person=third verb_number=singular` | `ok`, `state=generated`, `looks`, exit 0; one returned form, cell `present_third_singular`. |
| PL unknown lexeme `lexeme:pl.nieistniejacy.v`, `verb_form=finite` | `unknown`, `FORM_UNAVAILABLE`, exit 5. This records missing morphology for the supplied ID; it is not a failed translation. |
| PL `dostac` past without gender; PL `dorosly` superlative vocative | `unknown`, `UNKNOWN_PARADIGM_CELL`, exit 5; no cell matched the supplied bundle, so this is not `incomplete`. |
| PL request with a lexeme but no feature | `invalid`, `INVALID_FORMS_ARGUMENTS`, exit 2. CLI parsing rejects generation without at least one feature. A future UI may prevent submission and show the local validation message. |
| PL undeclared feature dimension `not_a_profile_feature=value` | `unsupported`, `UNSUPPORTED_FEATURE_DIMENSION`, exit 3; the response has no generation state. Preserve the status and diagnostic without relabeling it `invalid`. |

For the PL form, `GeneratedForm` returned profile `v2026-10-23.1`, paradigm
`pl_verb_present_i`, source `regular`, provenance ID
`golingua-pl-morphology-seed`, path `languages/pl/morphology/paradigms.json`,
version `2026-10-23-g10-02`, license `internal`, and the two-step
`select_stem`/`append` trace. For EN it returned profile `v2026-10-23.1`,
paradigm `en_verb_finite_principal_parts`, source `regular`, provenance ID
`golingua-en-morphology-seed`, path `languages/en/morphology/paradigms.json`,
version `2026-10-23-g10-01`, license `internal`, and a `select_stem` trace.
All returned feature values are shown in the corresponding frames. Full JSON is
preserved in the linked evidence comment.

## Status handling

| Actual API response | Presentation |
| --- | --- |
| `status=ok`, `state=generated` | Show every returned form in order and expose its actual fields. |
| `status=incomplete`, `state=incomplete` | Show the returned forms (possibly zero) and diagnostics; keep `incomplete`. |
| `status=unsupported`, `state=blocked` or `unavailable` | Keep `unsupported` and the exact state; show zero forms only when the response contains none. |
| `status=unsupported` without a generation state | Preserve the returned diagnostics (for example `UNSUPPORTED_FEATURE_DIMENSION`); do not invent a generation state. |
| `status=unknown`, `state=unknown` | Keep `unknown` and its actual diagnostic; do not infer a lemma or form. |
| `status=invalid` | Show the actual validation diagnostic; no generated form is implied. |
| `len(Forms)==0` | Show “0 forms returned” beside the original status and diagnostics; never synthesize a form. |
| Any other status | Preserve it and all returned diagnostics; never remap it to `unknown`. |

The enum and mapping come from `internal/morph/generate.go` and
`internal/app/service.go`. On this bundled data, `incomplete`, `blocked`, and
`unavailable` are **API-contract-only / CLI NOT RUN**: the current PL and EN
paradigm data has no cells without generation rules, the PL exceptions file is
empty, and normal seed generation does not emit those states. The mockup does
not fabricate JSON or claim those states were observed. Agent assessment is
separate from owner acceptance; linguistic acceptance remains **NOT RUN**.

## Input, focus, and navigation

- `Language` accepts the supported local choices `pl` or `en`. With Language
  focused, `Left` / `Right` changes the choice; `Enter` commits it. Keep typed
  feature names and values visible on language change; never rewrite them
  silently.
- `Lexeme ID` is a bounded text input. The screen may check identifier syntax,
  but the compiled language seed remains authoritative for whether the lexeme
  exists and can generate forms.
- Feature rows are explicit text `key=value` pairs. They are not dropdowns and
  the screen does not add default features. The service validates them against
  the selected language profile; returned diagnostics remain visible.
- `Tab` / `Shift-Tab` visits Language, Lexeme ID, each feature row, Add/Remove,
  Generate, Details (after a response), then returned forms. Focus is visible
  without color.
- `Enter` commits a text field or feature row, submits only when Generate is
  focused, and opens details only when Details is focused. `Up` / `Down` moves
  through feature rows or returned forms. `Esc` leaves a detail view without
  clearing the draft. `F1`–`F5` remain global.
- Resizing or switching modes preserves the draft and last response. The shared
  bottom prompt remains part of the global shell.

This is a layout specification only. It does not add engine behavior or accept
any generated form as linguistically correct.

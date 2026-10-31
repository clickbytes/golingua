# #476 — basic realization, negative route direction and lexical number

Base: `6dc2c51608ecd6ef474e69c2e5ec64c064a2a932` (product `851aa00`).
Date: 2026-10-31. Serial work on `main` only, by owner instruction.

## Reproduced failures and fixes

1. The English valency compiler promoted every negative binding into a compact
   target route, including bindings licensed for analysis only. Opening/closing
   consequently acquired two competing negative frames and target selection
   refused. Compact routes are now added only when the same sense/frame already
   has a target-capable variant in the declared construction package. Explicit
   compact bindings are retained once. Missing variants are errors. A variant
   licensed in both directions is also valid source evidence; a realize-only
   variant is not. Four already-existing bare-object variants are licensed on
   the dedicated closing negative frame. No competing frames are merged and no
   new whole-sentence templates are introduced.
2. Door attribution selected number by two literal variant names and missed past
   and bare-plural recipes. It now reads the selected bearer's number condition,
   respects explicit semantic multiplicity, and retains all four choices for
   unresolved cardinality. Nominal third-person agreement is a construction rule,
   not a newly assumed speech-participant property.
3. Real source evidence exposed a bug absent from independent realization: Polish
   lexical plural `drzwi` supplied grammatical plural to the compact article plan
   while another local planner selected singular, producing bare singular `door`.
   Both compact-route selection and article planning now share a number resolver.
   It prefers explicit Meaning multiplicity; otherwise source lexeme metadata
   `multiplicity_behavior=lexical_plural_not_cardinality` prevents guessing a
   count. The existing licensed singular/plural routes retain their explicit
   assumptions. Evidence is read for the exact predicate/role occurrence.
4. The existing `latać` lexeme had no past principal part/cells. Its data now
   includes `past_stem=lata` and five third-person imperfective past cells,
   implemented with the existing stem-selection/append procedure. Forms are
   `latał`, `latała`, `latało`, `latali`, `latały`. Present forms and the existing
   paradigm ID remain unchanged. Source: [WSJP PAN, latać, samolot/ptak](https://wsjp.pl/haslo/podglad/38367/latac/4212091/samolot-ptak), consulted
   2026-10-31; source provenance is recorded in the data. This does not introduce
   `lecieć`, a new sense, perfective aspect, or a per-sentence translation rule.

## Explicit contract corrections

Five existing realization records were corrected; no Meaning fixture, role,
operator, multiplicity, representation-gap declaration or budget was changed.

- EN->PL `1061/1063` retain the required texts `Ptak latał.` / `Ptaki latały.`.
  They require the explicit lexical choice `aspect=imperfective` and permit a
  referent-local article loss only when actual English source evidence exists.
- PL->EN `1062/1064` formerly demanded only `the` without an assumption, although
  their own note required determiner alternatives. The source supplies count and
  past, not an English article. Both a/the (singular) or bare/the (plural) are
  now required with their matching assumptions. Source-conditioned aspect loss
  is permitted, consistently with the existing simple-past policy.
- PL->EN negative `876` keeps all eight required texts and the original number
  and determiner requirements. Its contract now permits the source aspect loss,
  matching the previously accepted neighboring past-negative contracts.

Focused tests independently require exactly one loss at the appropriate event
or referent with actual source evidence and zero such losses after evidence is
removed from the same graph/context. They also require the Polish aspect choice.
These are declared contract corrections, not hidden additions to AllowedTexts.
Owner linguistic acceptance remains NOT EVALUATED.

## Observable results and verification

Before-change tests reproduce the six independent realization failures and the
analysis-only license leak. Separate before-change tests reproduce bare singular
`door` with real source evidence and the invented person assumption. The new
49 supplemental checks include manually authored IR, source graph comparisons,
role/scope/time/missing-argument contrasts, article/agreement refusals, explicit
one/many/unresolved door cardinality, occurrence-local losses and forward/inverse
morphology. Existing medical/academic ambiguity of English `doctor` is preserved.

Actual CLI produces:

- `Lekarz nie otwiera drzwi.`: eight licensed English article/count alternatives.
- `Nauczyciel nie zamyka okna.`: four English article alternatives with `does not`.
- `Lekarz nie zamykał drzwi.`: eight alternatives with `did not` and aspect loss.
- `Drzwi były czarne.`: `A door was black.`, `The door was black.`,
  `Doors were black.`, `The doors were black.`, with explicit count/article choices.
- `The bird flew.` / `The birds flew.`: `Ptak latał.` / `Ptaki latały.`,
  with explicit aspect choice and source article loss.

These outputs can remain ambiguous/incomplete; that is not an infrastructure
error or a reason to suppress alternatives. Before this change the door-past CLI
already retained two partial texts but also reported a variant-feature refusal;
this delivery does not claim that all six previously produced no text.

Local Go 1.27.3 verification: formatting, vet, build, help/version, focused tests,
real before/after CLI and the existing complete `go test ./tests/... -count=1 -v`.
The final local run contains 3000 registered checks, 18 status improvements and
no previous PASS regression. Integration: 194->202 PASS, 309 FAIL, 497->489 GAP;
EN->PL 40->42, PL->EN 154->160. All six integration GAP_REALIZE now pass; other
independent realization gaps and the remaining 489 integration gaps stay open.
The full command still exits 1. An intermediate code-only run before the final
annotation/contract corrections gave 196 PASS; the final 202 result includes the
five disclosed contract corrections. The 49 supplemental checks are not added
to the 1000-case total. Race and post-integration Actions evidence are recorded
in #476 and delivery artifacts.

## Boundaries

This closes this execution slice, not the complete plan. General determiner and
word-order contract review, duplicated language allowlists, general relational
predication, independent IR for GAP_I, general phrase/WH composition and DSE
acceptance remain open. No model, timeout increase, new runner, branch or PR is
introduced. No second-agent review or complete language acceptance is claimed.

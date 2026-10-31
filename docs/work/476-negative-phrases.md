# #476 — phrase-owned negative realization

Base: `88c201a6362255572089dee4a7f326dfe3483a55` (product code `c8ccd14`).
Date: 2026-10-31. Branch: `main` only, by explicit owner instruction.

## Delivered scope

The existing English auxiliary planner is shared by three explicitly licensed
negative families: ordinary transitive clauses, object-modified seeing, and
three-role giving. The family capability `vp-auxiliary-negation` opts a recipe
into this path; it is not a lexical, frame, operator, or agreement license.
Exact approved bindings and scope checks remain required.

The two migrated families now contain 9 variants instead of 37:

| Family | Before | After | Directions |
| --- | ---: | ---: | --- |
| `en.attributed_seeing_object_negative` | 8 | 3 | 2 source, 1 target |
| `en.ditransitive_negative` | 29 | 6 | 4 source, 2 target |

Target recipes contain slots, the licensed adposition and punctuation. NP plans
own determiner alternatives and `a/an`; the VP owns `do/does/did + not + base`.
The frame determines argument roles. Local choices are bounded and referent-local,
not stored as products of whole-sentence article templates. Target-only variants
cannot enter source parsing; source-only variants cannot generate target text.

Source agreement is checked against the consumed subject and actual auxiliary,
including an accepted input spelling, not only the canonical template terminal.
The Polish object-modified negative construction uses its existing case and
agreement rules for singular/plural participants; no Polish forms were invented.

The generic parser retains an exactly identical derivation only once. Optional
terminals and a prebuilt NP group previously reached the same derivation twice,
causing `DUPLICATE_CANDIDATE_ID` later. The deduplication compares full parse
alternatives within a root, including bindings and unresolved evidence. It does
not merge different readings, participant identities or source evidence, and it
does not remove search budgets.

## Observable flows

Newly supported bounded combinations include:

- `The teachers did not see the black cats.` →
  `Nauczyciele nie widzieli czarnych kotów.` (also the licensed OVS order).
- `Teachers do not see black cats.` →
  `Nauczyciele nie widzą czarnych kotów.`
- `Nauczyciele nie widzieli czarnych kotów.` → four English bare/the choices.
- `Teachers do not give water to a doctor.` →
  `Nauczyciele nie dają lekarzowi wody.`
- `The teacher does not give an apple to the doctor.` →
  `Nauczyciel nie daje lekarzowi jabłka.`

These observations do not claim general adjective, quantifier, reflexive,
coordination or question coverage. Source lexical alternatives such as the
unbound competing reading of `apple` are still visible as incomplete/ambiguous.
The single-subject past object-modified seeing and past `to` transfer already
worked before this change; they are regression controls, not new capabilities.

## Contract corrections

The unified compact negative route is now recognized as English simple past by
the existing aspect-loss policy. Three corpus realization contracts
(`A1.8-NEG-866/868/870-PL-EN`) permit `aspect` loss when actual Polish source
evidence provides lexical aspect. No text, semantic role, required article
assumption, or representation-gap declaration was changed. A focused test
requires exactly one loss on the correct event for these actual sources, and
requires no such loss after source evidence is removed from the same IR/context.
This correction does not count as a newly translated corpus sentence.

The old catalog test requiring red/white/blue/green to be absent now validates
their already-existing entity-bearer property signatures instead. The generic
`color` concept still cannot be assigned an invented bearer role, and missing
independent fixtures remain missing. No new color data or GAP removal is included.
New scope/attachment controls forbid silently reusing the old meaning's surface;
they do not assert that a future valid extension must always remain unsupported.

## Verification and limits

Local Go 1.27.3: formatting, vet, build, actual before/after CLI, independently
authored IR realization, source graph comparison, invalid agreement/articles,
repeated-concept referents, changed attachment/scope, and focused race checks.
The complete local run executes the existing `go test ./tests/... -count=1 -v`.
All 3000 registered statuses match the baseline: 194 integration PASS, 309 FAIL,
497 GAP (360 GAP_I, 131 GAP_ANALYZE, 6 GAP_REALIZE); the command still exits 1.
Supplemental tests demonstrate the new combinations, not additional corpus PASS.
Post-integration Actions evidence and final SHA are recorded in issue #476.

General occurrence-addressed `SourceFeatures`, independent IR for the remaining
GAP_I cases, general NP/VP/WH composition, synonym-specific planning and other
historical boundary contracts remain open. There is no new full DSE acceptance,
no new language, no timeout increase, and no claim of an independent second-agent
review. Owner language acceptance remains NOT EVALUATED.

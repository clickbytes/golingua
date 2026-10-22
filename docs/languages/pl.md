# Polish Go lexicon and limited morphology

The Polish semantic source is `languages/pl/lexicon.go` plus the explicitly ordered, typed `lexicon_*.go` groups. `pl.LexicalDocument()` constructs separate `MorphLexeme`, `LexicalSense` and pending `SemanticBinding` objects; `pl.CompileLexicon(coreGraph)` freezes them after validation against the supplied compiled CORE.

## Source inventory (after G07-03 entity and valency fix)

**728 Polish lexemes, 744 senses and 743 pending candidate bindings** cover 743 of 1000 CORE concepts. **257 CORE remain absent from the Polish compiled lexicon.** Approved bindings: **0**. The editorial ledger `data/pl/1000-core-review.tsv` has 983 proposed candidate mappings; **240 event candidate meanings are worklist-only** and have no Go lexical binding yet. Counts are source-derived; run `golingua catalog --summary --json` to verify the observed snapshot.

The 24 reviewed motion/transport senses and their frames are declared in `lexicon_motion.go` / `valency_motion.go` without additional morphology or approved construction permission. The lexicon preserves polysemy: `skóra` (body/material), `ryba` (animal/food), `bank` (institution/premises), `szkoła` (building/institution), `sklep`, `restauracja`, `szpital`, `biblioteka` (place/organization). Equal lemmas may have multiple senses; no implied equivalence is licensed by their shared string. New institution and food senses live in `lexicon_institutions_food.go` or were appended to their existing lemma in `lexicon_places.go`.

`pl.CompileValency()` provides **36** explicitly scoped frames: `widzieć`, `patrzeć na`, `słyszeć`, `dostać`, `dać`, `pytać`, `odpowiadać`, `czytać`, `uczyć`, `uczyć się`, `jeść`, `spać`. Semantic roles use actual CORE signatures: e.g. `sleeper`, `reader/material`, `sender/addressee`, `instructor/learner`. `patrzeć na` uses the oblique PP family, `dać` a ditransitive structure, `uczyć się` and `spać` the intransitive family, not a generic accusative transitive family. Only currently supportable core roles are licensed; optional topic/content complements not represented in CORE require future separate review.

All these valency frame families are **structural declarations without approved construction bindings**. `pl.transitive`, `pl.intransitive`, `pl.oblique` and `pl.ditransitive` do not imply working sentence generation or a runtime plugin.

## Morphology seed versus full lexicon

The existing `pl.Load()` builds a separate small seed of authored morphology and an inverse index; it does **not** make all 704 semantic lexemes inflectable. Only declared cells work:

- `student`: masculine-personal noun selected case/number cells.
- `czytać`: selected present indicative person/number cells.
- `kot`: 6 masculine-animal noun cells.

`lexeme:czytac-v` is a separate morphology seed ID from `lexeme:pl.czytac.v`; do not unify them without an explicit migration and a real review of paradigms. The seed lacks general noun paradigms, adjective agreement, past/future/conditional, aspect realizations, and irregulars. Unknown forms return typed `unknown`/`unsupported`; inverse analysis keeps all distinct readings. The seed loader's 128-record limit does not determine lexical inventory size; finite table defaults to 100000 readings/32 MiB.

`languages/pl/manifest.json` correctly retains `sentence_analysis: unsupported` and `meaning_realization: unsupported`. For owner evaluation use `golingua forms --language pl ... --json` or the `catalog` commands; translation remains unavailable without separately approved data and adapters.

**Owner manual linguistic acceptance: NOT RUN.** No CI green check replaces semantic/inflection review.

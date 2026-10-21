# Polish morphology seed

## In-code lexical catalog

`pl.LexicalDocument()` declares Polish lexical slices in Go, authored in
G07-03 batches: 426 lexemes, 435 separate senses and 434 **pending**
candidate links to the neutral CORE catalog (files
`lexicon.go`, `lexicon_people.go`, `lexicon_nature.go`, `lexicon_body.go`,
`lexicon_places.go`, `lexicon_objects.go`, `lexicon_food.go`,
`lexicon_basics.go` and the group aggregator `lexicon_groups.go`).
`pl.CompileLexicon(coreGraph)`
validates and freezes them against the exact supplied CORE graph. The second
sense of `ksiazka.work` deliberately has no CORE link. The new
`dom.household`, `szkola.institution`, `reka.arm` and
`dziecko.kinship` senses do use `candidate` links to at most one CORE
concept each. Candidate links are not approved semantic bindings and
cannot authorize translation.
Batch 2 authored the nature group (animals, plants, substances, landforms;
`staw` for the water body is a separate lexeme `pl.staw_zbiornik.n`) and the
body group (`skóra` is one lexeme with the `part` and `material` senses of
the polysemy; `ryba` and `owoc` gained second food senses merged manually).
Batch 3 adds the places group (settlements,
buildings, rooms, routes), the objects group (domestic objects, tools,
machines, technology), the food group (food, beverages, garments) and the
basics group (resolved original-49 nouns, leftover person entries and
property nouns). Reviewable verbs stay worklist-only until valency batches.
The remaining reviewed rows are 566 absent concepts; re-check the totals
with `catalog --summary --json` before each further batch.

The full editorial decision record behind this slice and the rest of the
1000 concepts is `data/pl/1000-core-review.tsv`; see `data/pl/README.md`.
Rows marked there as reviewed but not yet implemented in Go are
worklist-only and add nothing to the compiled catalog.

`pl.CompileValency()` separately compiles the initial four sense-specific
argument frames (`widzieć`, `patrzeć na`, `słyszeć`, `dostać`) and the
batch-4 core-verb frames for `pytać` (addressee accusative, topic with
`o`), `odpowiadać na` (content with `na`), `czytać` (theme accusative),
`uczyć` (learner accusative, discipline genitive), `uczyć się` (content
genitive), `jeść` (theme accusative) and one-agent `spać`. The frames
record Polish case and prepositions but have no approved construction
binding, and none of these verbs has a seed paradigm yet.

The lexical catalog supplies one additional in-code morphology entry, `kot`,
with six explicitly declared cells for a masculine animal noun. It supplies no
sentence constructions. Source
metadata points to the project-authored Go declarations
and records the license as `unknown` until the project establishes one.
Counts refer only to the declared in-code slices, not to all Polish
vocabulary.

The Go package `github.com/clickbytes/golingua/languages/pl` loads the bundled Polish profile, manifest, paradigms, and lexical principal parts. `pl.Load()` compiles the seed into the shared generator and finite relation analyzer. `Seed.Generate(lexemeID, featureValues)` and `Seed.Analyze(surface)` are Go APIs; the CLI exposes them through `forms`, but this package does not realize sentences.

## Included classes

- One explicitly scoped masculine personal noun pattern, represented by `student`, for singular and plural across the seven profile cases. It uses the `t → ci` alternation in locative/vocative singular and nominative/vocative plural.
- One present indicative active finite verb pattern represented by `czytać`, with all person/number endings and explicit principal stem `czyt`.
- An in-code masculine animal noun pattern for `kot`: singular nominative, genitive, accusative and instrumental; plural nominative and genitive. Other cells remain unsupported.

The finite relation table retains syncretic readings. For `studenta`, analysis returns both singular genitive and accusative; for plural `studentów`, it returns plural genitive and accusative. NFC is the only surface normalization; case and diacritics remain significant.

This seed does not define a general Polish declension/conjugation system. It does not cover other noun stem alternations, adjective agreement, past/future/conditional forms, aspect, stress, phonology, or lexemes outside the listed patterns. `morphology/exceptions.json` is an empty array; the loader validates and compiles formal exception records when independently reviewed rules are added. No exception/block rule is claimed by this seed. Missing or unsupported cells do not get guessed forms. No owner-selected manual example has been reviewed yet; local API smoke is implementation verification only.

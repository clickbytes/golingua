# Polish morphology seed

## In-code lexical catalog

`pl.LexicalDocument()` now declares a first Polish lexical batch in Go:
16 lexemes, 20 separate senses and 16 **pending** candidate links to the
neutral CORE catalog. `pl.CompileLexicon(coreGraph)` validates and freezes
them against the exact supplied CORE graph. The second senses of `dom`,
`szkoła`, `książka` and `ręka` deliberately have no CORE link yet. Candidate
links are not approved semantic bindings and cannot authorize translation.

`pl.CompileValency()` separately compiles four sense-specific argument frames
for `widzieć`, `patrzeć na`, `słyszeć` and `dostać`. The frames record Polish
case and the preposition `na`; they have no approved construction binding.

The lexical catalog supplies one additional in-code morphology entry, `kot`,
with six explicitly declared cells for a masculine animal noun. It supplies no
sentence constructions. Source
metadata points to the project-authored Go declarations
and records the license as `unknown` until the project establishes one.
Counts refer only to this initial batch, not to all Polish vocabulary.

The Go package `github.com/clickbytes/golingua/languages/pl` loads the bundled Polish profile, manifest, paradigms, and lexical principal parts. `pl.Load()` compiles the seed into the shared generator and finite relation analyzer. `Seed.Generate(lexemeID, featureValues)` and `Seed.Analyze(surface)` are Go APIs; the CLI exposes them through `forms`, but this package does not realize sentences.

## Included classes

- One explicitly scoped masculine personal noun pattern, represented by `student`, for singular and plural across the seven profile cases. It uses the `t → ci` alternation in locative/vocative singular and nominative/vocative plural.
- One present indicative active finite verb pattern represented by `czytać`, with all person/number endings and explicit principal stem `czyt`.
- An in-code masculine animal noun pattern for `kot`: singular nominative, genitive, accusative and instrumental; plural nominative and genitive. Other cells remain unsupported.

The finite relation table retains syncretic readings. For `studenta`, analysis returns both singular genitive and accusative; for plural `studentów`, it returns plural genitive and accusative. NFC is the only surface normalization; case and diacritics remain significant.

This seed does not define a general Polish declension/conjugation system. It does not cover other noun stem alternations, adjective agreement, past/future/conditional forms, aspect, stress, phonology, or lexemes outside the listed patterns. `morphology/exceptions.json` is an empty array; the loader validates and compiles formal exception records when independently reviewed rules are added. No exception/block rule is claimed by this seed. Missing or unsupported cells do not get guessed forms. No owner-selected manual example has been reviewed yet; local API smoke is implementation verification only.

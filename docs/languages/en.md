# English morphology seed

## In-code lexical catalog

`en.LexicalDocument()` declares an initial English batch in Go: 19 lexemes,
28 distinct senses and 19 **pending** candidate links to neutral CORE.
`en.CompileLexicon(coreGraph)` validates those records against the exact CORE
snapshot. Meanings such as the institutional sense of `school` and the
understanding sense of `see` remain separate and unmapped. English records are
authored independently of Polish records; equal counts and word pairs are not
assumed.

`en.CompileValency()` compiles five sense-specific frames. In particular,
directed `look` uses an `at` oblique argument, whereas visual `see` uses a
direct object. There are no approved construction bindings. Candidate links
cannot authorize translation. Source metadata points to the project-authored
Go files and records the license as `unknown` until established.

This catalog is separate from the five-entry morphology seed below. It does
not add inflection paradigms, auxiliaries or sentence realization.

The Go package `github.com/clickbytes/golingua/languages/en` loads the bundled `features.json`, `manifest.json`, `morphology/paradigms.json`, and `morphology/lexemes.json`. `en.Load()` validates and compiles the package into a shared morphology engine, then builds a finite relation table for inverse analysis. Its `Seed.Generate(lexemeID, featureValues)` and `Seed.Analyze(surface)` methods expose morphology from Go. The repository CLI has no `forms` command, so this seed does not claim CLI integration or sentence translation.

## Included classes

- Regular count nouns: singular base and plural `+s` (seed entries: `book`, `cat`).
- A bounded regular verb pattern for base, third person singular `+s`, past and past participle `+ed`, and present participle `+ing` (seed entry: `walk`).
- An explicit principal parts paradigm for `go`: `go`, `goes`, `went`, `gone`, `going`. These listed forms are selected as principal parts; the package does not infer them from the lemma or label them as reviewed exception rules.
- Regular adjective degree for positive, comparative `+er`, and superlative `+est` (seed entry: `bright`).

`Analyze` preserves all readings: for example, the finite table contains separate `past` and `past_participle` readings for `walked`. Surfaces use NFC normalization only; case and diacritics remain significant.

This is a small declared seed, not a complete English grammar. It does not implement silent-e deletion, consonant doubling, y-to-i changes, irregular plural rules, person/number agreement, auxiliaries, or sentence realization. Additional irregular exception rules require their own valid provenance and review metadata. No owner-selected manual example has been reviewed yet; the local API smoke is implementation verification only.

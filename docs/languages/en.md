# English morphology seed

## In-code lexical catalog

`en.LexicalDocument()` declares 71 English lexemes, 80 distinct senses and 71
**pending** candidate links to neutral CORE (19 original lexemes plus 52
project-authored noun candidates).
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

The morphology catalog assigns existing regular paradigms to seven additional entries:
`dog`, `teacher`, `doctor`, `school`, `hand`, `eye` and `look`. It does not add
auxiliaries or sentence realization.

The 52 new noun entries (including `person`, `adult`, `animal`, `plant`, `tree`,
`artifact`, `door`, `vehicle`, `location`, `body part`, `event`, `spouse`,
`romantic partner`, `friend`, `acquaintance`, `neighbor`, `stranger`, `guest`,
`resident`, `citizen`, `visitor`, `tourist`, `traveler`, `parent`, `mother`,
`father`, `sibling`, `brother`, `sister`, `grandparent`, `grandchild`, `cousin`,
`offspring`, `infant`, `toddler`, `adolescent`, `older adult`, `stepparent`,
`adoptive parent`, `host`, `aunt`, `uncle`, `passenger`, `driver`, `pedestrian`,
`worker`, `employee`, `employer`, `manager`, `farmer`, `cook`, and `waiter`) currently have
no assigned paradigm, form inventory, valency frame, or construction binding.
They are lexical candidates only. The complete editorial ledger is
[`data/en/1000-core-review.tsv`](../../data/en/1000-core-review.tsv): it has
1000 exact current CORE IDs, 19 `EXISTING_PENDING`, 51 `REVIEWED_LEXICALIZED` rows and 930
`UNREVIEWED` rows. The nine existing unmapped senses are `en.doctor.academic`,
`en.house.institution`, `en.school.institution`, `en.school.fish`,
`en.book.work`, `en.hand.worker`, `en.see.understand`, `en.look.appear`, and
`en.get.understand`. Their mappings still need separate lexical decisions.

This remains an incremental data update, not completion of the 1000-concept English
editorial review. All 71 bindings remain pending; none is owner-approved.

The Go package `github.com/clickbytes/golingua/languages/en` loads the bundled `features.json`, `manifest.json`, `morphology/paradigms.json`, and `morphology/lexemes.json`, plus in-code entries. `en.Load()` validates and compiles the package into a shared morphology engine, then builds a finite relation table for inverse analysis. Its `Seed.Generate(lexemeID, featureValues)` and `Seed.Analyze(surface)` methods are available through CLI `forms`. This seed does not claim sentence translation.

## Included classes

- Regular count nouns: singular base and plural `+s` (seed entries: `book`, `cat`).
- A bounded regular verb pattern for base, third person singular `+s`, past and past participle `+ed`, and present participle `+ing` (seed entry: `walk`).
- An explicit principal parts paradigm for `go`: `go`, `goes`, `went`, `gone`, `going`. These listed forms are selected as principal parts; the package does not infer them from the lemma or label them as reviewed exception rules.
- Regular adjective degree for positive, comparative `+er`, and superlative `+est` (seed entry: `bright`).

`Analyze` preserves all readings: for example, the finite table contains separate `past` and `past_participle` readings for `walked`. Surfaces use NFC normalization only; case and diacritics remain significant.

This is a small declared seed, not a complete English grammar. It does not implement silent-e deletion, consonant doubling, y-to-i changes, irregular plural rules, person/number agreement, auxiliaries, or sentence realization. Additional irregular exception rules require their own valid provenance and review metadata. No owner-selected manual example has been reviewed yet; the local API smoke is implementation verification only.

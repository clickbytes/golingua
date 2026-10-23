# English morphology seed

## In-code lexical catalog

`en.LexicalDocument()` declares 1002 English lexemes, 1011 distinct senses and 1001
CORE binding records (the concept inventory plus the sentence-only determiner `a`).
Twenty delegated review batches and audit #192 approved 996 exact equivalent
links; 5 records remain pending. The additional determiner sense has no CORE
binding because it is used only to analyze English noun phrases.
`en.CompileLexicon(coreGraph)` validates those records against the exact CORE
snapshot. Meanings such as the institutional sense of `school` and the
understanding sense of `see` remain separate and unmapped. English records are
authored independently of Polish records; equal counts and word pairs are not
assumed.

`en.CompileValency()` compiles five sense-specific frames. In particular,
directed `look` uses an `at` oblique argument, whereas visual `see` uses a
direct object. Five exact EN sense/frame/family/variant routes (`see`, `hear`,
`get` in its receive sense, `read` with text, and `look at`) have approved
construction bindings after delegated review. `look at` uses the distinct
oblique `at` + NP construction assessed in G05 #201.
Remaining candidate lexical links stay pending; neither they nor the approved
EN links alone authorize translation.
Source metadata points to the project-authored Go files and records the license
as `unknown` until established.

The morphology catalog assigns paradigms to a bounded, explicit subset of the
English lexicon. G10-03 adds selected count-noun number classes alongside the
G10-01 finite verb classes and explicit principal parts. Most of the 1000
lexical entries still have no assigned morphology. It does not add auxiliaries
or unrestricted sentence realization.

The 981 newly lexicalized entries cover people and social roles, organizations, basic concepts,
animals, plants, body parts, materials, vehicles, household objects, communication verbs
and physical action verbs. Examples
include `person`, `friend`, `company`, `communication`, `turtle`, `root`, `heart`,
`air`, `glass`, `car`, `table`, `bottle`, `hammer`, `computer`, `breakfast`, `beef`, `shirt`, `shoe`, `village`, `room`, `road`, `airport`, `bridge`, `hill`, `valley`, `island`, `river`, `cloud`, `document`, `message`, `map`, `website`, `length`, `mass`, `pitch`, `joy`, `anxiety`, and verbs such as `speak`, `explain`, `thank`, `send`, `translate`, `act`, `lift`, `buy`, `sell`, `pay`, `donate`, `search`, `desire`, `prefer`, `breathe`, `swallow`, `walk`, `run`, `swim`, `drive`, `listen`, `remember`, `decide`, `hope`, `permit`, `quantity`, `accuracy`, and `manufacture`, `money`, `passport`, `toilet`, `pizza`, `friendship`, and `ownership`. Noun candidates have
no assigned paradigm, form inventory, valency frame, or construction binding. Verb candidates likewise have no assigned morphology or valency. They are lexical candidates only. The complete editorial ledger is
[`data/en/1000-core-review.tsv`](../../data/en/1000-core-review.tsv): it has
1000 exact current CORE IDs, 19 `EXISTING_PENDING`, 981 `REVIEWED_LEXICALIZED` rows and no
`UNREVIEWED` rows. The eight existing unmapped senses are `en.doctor.academic`,
`en.house.institution`, `en.school.fish`,
`en.book.work`, `en.hand.worker`, `en.see.understand`, `en.look.appear`, and
`en.get.understand`. Their mappings still need separate lexical decisions.

The ledger has a lexical candidate for every current CORE concept. Twenty delegated review batches and audit #192 approved 996 exact equivalents across 996 concepts and
retained 5 pending binding records across 4 concepts. The remaining candidates are recorded in issue #192. The existing
`en.school.institution` sense is a pending candidate; eight existing senses
remain unmapped. Construction approvals do not imply these lexical decisions.

The Go package `github.com/clickbytes/golingua/languages/en` loads the bundled `features.json`, `manifest.json`, `morphology/paradigms.json`, and `morphology/lexemes.json`, plus in-code entries. `en.Load()` validates and compiles the package into a shared morphology engine, then builds a finite relation table for inverse analysis. Its `Seed.Generate(lexemeID, featureValues)` and `Seed.Analyze(surface)` methods are available through CLI `forms`. The compiled CLI enables exact `seeing`, `hearing`, `looking`, and `receiving` routes when it loads an active snapshot. The `looking` slice uses `looks` and exact `at`; `receiving` uses the declared principal parts `get`, `gets`, `got`, `gotten`, `getting`, analyzes `got`, and realizes past tense only.

## Included classes

- Selected count nouns have singular/plural cells for regular `+s`, sibilant
  `+es`, or consonant + `y → ies`. Assignment is explicit per lexeme; there is
  no automatic spelling classifier. Examples include `book/books`,
  `bus/buses`, `branch/branches`, and `city/cities`.
- Selected irregular nouns list both principal parts directly: `child/children`,
  `person/people`, `mouse/mice`, `goose/geese`, `foot/feet`, `tooth/teeth`,
  `leaf/leaves`, and `knife/knives`. `fish` (animal sense) and `sheep` have
  separate singular/plural cells selecting the same surface form, so analysis
  preserves both number readings. These entries do not license general rules
  for other irregular nouns.
- `food`, `money`, and `water` remain explicitly singular-only mass nouns.
- English finite regular verb classes add person/number present cells, a person-invariant past cell, imperative, and both participles. Subclasses cover plain endings, sibilants, final silent `e`, and consonant + `y`; each is attached only to explicitly selected entries such as `print`, `brush`, `bake`, and `fry`.
- An explicit finite principal-parts paradigm covers `see`, `hear`, `look`, and `get`, and the standalone seed example `go`. Each lists base, third-person singular, past, past participle, and present participle. Listed forms are selected directly; the package does not infer irregular forms from the lemma or claim they are reviewed exception rules.
- Regular adjective degree for positive, comparative `+er`, and superlative `+est` (seed entry: `bright`).

`Analyze` preserves all readings: for example, the finite table contains separate `past` and `past_participle` readings for `walked`. Surfaces use NFC normalization only; case and diacritics remain significant.

This is a bounded declared seed, not a complete English grammar. The spelling
classes do not implement consonant doubling, infer unlisted irregular plurals,
provide auxiliaries, or enable unrestricted sentence realization. The English
lexical catalog currently has no adjective or pronoun lexemes; `bright` remains
a standalone morphology seed example. The compiled sentence routes use finite
`see`, `hear`, `look` and past `get` forms. English `food`, `money`, and `water`
are marked as mass nouns for target number selection. The transitive source
construction recognizes `the`, `a`, and a bare object where current
lexical/morphology data allow it. The oblique `look at` route recognizes the
exact `at` terminal. Owner-authored `looking` CLI observations are in
[`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../../data/sentences/observed-2026-10-23-g05-14-looking.tsv);
owner sign-off remains **NOT RUN**.

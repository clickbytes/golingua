# UniMorph morphology import

`golingua-import unimorph` reads a local, three-column UniMorph TSV stream:
`lemma<TAB>surface<TAB>semicolon-separated-tags`. It produces a bounded JSON
report of observed forms and review candidates. It does not download a corpus,
activate a paradigm, generate unobserved forms, assign a ConceptID, or infer a
lexical meaning.

The reviewed PL and EN maps live in:

- `languages/pl/import/unimorph-profile.json`
- `languages/en/import/unimorph-profile.json`

Each file maps exact tag spellings to the matching existing language feature
profile version. A map may qualify a tag by POS where one source tag has more
than one use. A tag without an applicable mapping remains in `unmapped_tags`;
the importer does not normalize or guess what it means. Check the mappings
against the exact UniMorph release and source documentation before use.

Example source rows (tabs separate the three columns):

```text
dom	dom	N;SG;NOM
dom	domu	N;SG;GEN
house	house	N;SG
house	houses	N;PL
```

Run with the repository's source and feature profiles:

```sh
go run ./cmd/golingua-import unimorph \
  --input ./data/unimorph-pl.tsv \
  --mapping ./languages/pl/import/unimorph-profile.json \
  --features ./languages/pl/features.json \
  --source-id unimorph-pl \
  --source-version 1 \
  --license CC-BY-SA-3.0 \
  --source-uri https://github.com/unimorph/pol
```

Use a real release identifier, license label, and source location for the
corpus being imported. Provenance is metadata and does not establish permission
to redistribute. No UniMorph corpus is bundled or downloaded by this repository.

Output groups only observed records by language, lemma, and POS. A class
fingerprint summarizes the set of mapped feature signatures in that group; it
is a review aid, not a compiled class. Identical rows are counted and
deduplicated. Missing forms do not block a candidate. The report retains source
tags, source line, provenance, and mapped versus unmapped tags. Unmapped POS
stays empty, so a reviewer can see the unresolved record instead of an inferred
part of speech. Candidate status is always `needs_review` and activation is
always `not_activated`.

The stem candidate list includes only a source form exactly equal to the lemma,
labelled `observed_exact_lemma_form`. Other surfaces are never split into a
stem and ending. `potential_exceptions` stays empty unless a future reviewed
rule can identify an exception; a surface differing from its lemma is not by
itself evidence of irregularity.

The current Polish feature profile distinguishes several masculine classes,
while a generic UniMorph masculine tag does not select one of them. Such tags
remain unmapped until a language specific mapping is independently reviewed.

Inputs are bounded by bytes, row length, record count, field size, and tags per
row. Malformed row structure or invalid UTF-8 fails with a diagnostic. Unknown
tags are preserved and do not discard otherwise valid rows. There is no
automatic comparison with legacy JavaScript data, no translation evaluation,
and no test fixture runner.

References: [UniMorph schema](https://unimorph.github.io/schema/) describes
records as a lemma, inflected form, and bundle of morphological features.
The [UniMorph Polish repository](https://github.com/unimorph/pol) identifies
its data sources and license; verify the exact release and intended use before
importing it.

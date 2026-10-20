# New golingua dataset format

`golingua-import` reads a local manifest and the definition files it names. It
validates them as offline data; it does not activate a runtime, fetch a URI, or
execute content. The formats are owned by golingua and do not mirror the legacy
JavaScript data.

## Files

The manifest uses schema version 1 and identifies exactly one concepts file and
one lexicon file. Files are named `concepts.json` and `lexicon.json` here for
clarity; the manifest paths must remain relative to the selected data directory.

```json
{
  "schema_version": 1,
  "dataset_id": "demo-pl-en",
  "version": "1",
  "sources": [{
    "id": "hand-authored",
    "path": "sources/demo",
    "version": "1",
    "license": "CC0-1.0"
  }],
  "files": [
    {"id": "concepts", "path": "concepts.json", "source_id": "hand-authored"},
    {"id": "lexicon", "path": "lexicon.json", "source_id": "hand-authored"}
  ]
}
```

`concepts.json` is the versioned language-neutral CORE document accepted by
`internal/core`. Its `concepts` contain stable `concept:` IDs, a kind such as
`entity` or `event`, and optional explicitly declared roles and definition
edges. Language-specific labels and parts of speech do not belong there.

```json
{
  "schema_version": 1,
  "relations": [],
  "concepts": [
    {"id": "concept:dwelling", "kind": "entity", "definition": "a place to live"}
  ]
}
```

`lexicon.json` is the versioned lexical document accepted by
`internal/lexicon`:

```json
{
  "schema_version": 1,
  "lexemes": [
    {"id": "lexeme:pl.dom", "language": "pl", "lemma": "dom", "pos": "noun"},
    {"id": "lexeme:en.house", "language": "en", "lemma": "house", "pos": "noun"}
  ],
  "senses": [],
  "bindings": []
}
```

MorphLexeme records carry language, lemma and POS. LexicalSense records are
separate and point to a lexeme. SemanticBinding records are separate again and
point from a sense to a CORE concept. An unmapped lexeme may have no sense or
binding; this records morphology without asserting semantic equivalence.
Bindings require a matching CORE concept and retain provenance. Only an
approved equivalent binding can authorize semantic transfer.

## Run

```sh
go run ./cmd/golingua-import --manifest ./data/manifest.json --data-dir ./data
```

The command prints counts only after all listed files pass strict schema,
reference, checksum, size, and path validation. Unknown or repeated JSON fields,
duplicate IDs, missing references, malformed records, symlink escapes, and
resource-limit violations fail closed. No Go tests or external importer
dependencies are required by this format.

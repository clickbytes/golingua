# Lexical records and semantic bindings

`internal/lexicon` separates a language-specific `MorphLexeme` from its `LexicalSense` records and any `SemanticBinding` to CORE. A MorphLexeme has no CORE field and can be validated and compiled without an ontology snapshot. Lemma matching is exact and preserves case and Unicode spelling; this layer does not normalize forms or infer morphology.

## Versioned document

The v1 JSON document has `schema_version`, `lexemes`, `senses` and `bindings`. IDs are serialized as stable `lexeme:*`, `sense:*` and `concept:*` strings. Decoding is bounded, rejects unknown or mis-cased fields and trailing JSON, and validates all lexical links. `Encode` emits a deterministic order. The following is a shape example only; it is not approved dictionary data:

```json
{
  "schema_version": 1,
  "lexemes": [
    {"id":"lexeme:bank-n","language":"en","lemma":"bank","pos":"noun"}
  ],
  "senses": [
    {"id":"sense:bank-financial","lexeme_id":"lexeme:bank-n","register":"neutral"},
    {"id":"sense:bank-river","lexeme_id":"lexeme:bank-n","register":"neutral"}
  ],
  "bindings": []
}
```

Multiple senses and same-lemma lexemes remain separate. `ByLemma` returns every exact language+lemma match, including homographs and different parts of speech. Indexes are compiled into an immutable snapshot and return defensive copies. Missing register or usage stays unspecified; it is not interpreted as a default value.

## CORE mapping and transfer boundary

Bindings distinguish `equivalent`, `broad`, `narrow`, `related` and `candidate`, with `pending` or `approved` status and a bounded `provenance.SourceRef`. The sense-to-concept pair is unique in one document. A CORE snapshot is required when bindings exist so each concept reference can be resolved. Without bindings, CORE is optional.

Only `approved` + `equivalent` reports `TransferAuthorized`. Pending or non-equivalent bindings remain indexed for inspection but cannot authorize semantic transfer. Multiple approved equivalent bindings for a sense are all returned; callers must resolve ambiguity explicitly. The model records attribution metadata but does not verify a source or interpret its license.

The lexicon model does not implement a morphology engine, paradigm interpreter, importer, language seed or translation pipeline. Those belong to later queued issues. IDs, properties, text, record counts and JSON size have finite limits; invalid records, unknown links and exhausted budgets retain distinct status outcomes.

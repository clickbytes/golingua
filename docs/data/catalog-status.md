# Catalog status

Run the current in-code inventory from the repository root:

```sh
go run ./cmd/golingua catalog --summary --json
go run ./cmd/golingua catalog --details --json
```

The summary compiles the actual CORE and language snapshots. The details
view reports one row per stable ConceptID with `pl` and `en` values:
`approved`, `pending`, `non_transferable`, or `absent`. The report counts a
concept once even if it has multiple senses or lexical candidates. It counts
lexemes and senses separately from concepts and inflected forms.

Current catalog after the first CORE growth batch: 236 CORE concepts toward an
editorial target of about 1000; PL has 16 lexemes, 20 senses and 16 pending
candidate bindings; EN has 19 lexemes, 28 senses and 19 pending candidate
bindings. No binding is approved for transfer, so approved PL+EN pairs are
zero. The count is not a measure of working sentence translations or
linguistic quality. Manual assessment on owner supplied examples remains
NOT RUN.

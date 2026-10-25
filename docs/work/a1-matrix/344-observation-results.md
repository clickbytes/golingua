# #344 observation metadata results

Status: implementation evidence only; owner acceptance remains `NOT_EVALUATED`.

`TestInventory` emits one individual `matrix_inventory=` JSON object for each registered direction (schema version 1), including source SHA (`GOLINGUA_MATRIX_SHA`, otherwise `UNRECORDED`), case/pair/group/direction, declared tags, source input and semantics, allowed/required Meaning IDs, and owner status. It emits one `matrix_snapshot=` pack identity and `matrix_sense_association=` records joining catalogue senses to lexemes, language, POS, and declared paradigm. Empty paradigms are `not-declared`; absent joins are `unknown`. These are catalogue associations, not executed morphology evidence (`morphology_executed:false`).

Runtime observation logs distinguish `native_analysis`, `independent_realization`, and `integration` with `adapter_executed`; source probes retain separate stages. A GAPI realization remains explicitly unexecuted with its reason. Actual realization logs include the authored Meaning input and context when present. Existing checks and acceptance behavior were not changed.

Optional `person` and applicable `number` fields were added only to non-gap common-NP realization contracts in A1.4/A1.8. Existing required assumptions, texts, meanings, and quantity semantics were not changed. The A1.8 English door cases retain explicit singular/plural text assumptions.

## Local checks

- `gofmt` on the six modified Go files: completed.
- `go test ./tests/translation -run '^TestInventory$' -count=1 -v`: PASS; 1,000 `matrix_inventory` records, one snapshot record, 2,023 catalogue association records. `GOLINGUA_MATRIX_SHA` was unset, so source SHA is `UNRECORDED`. Pack identity digest: `5fe8d84d0a4ffbae85ad34d1864015690d32ab2c49112fad26a07450c76d9d77`.
- `go test ./tests/matrix -run '^(TestRequiredSurfaceContracts|TestRequiredTextGroups|TestContextMatchesBindsAnchorsToGraphRoles|TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions|TestMatchAllowedCandidateUsesMatchingVariantContext|TestMatchAllowedCandidatePropagatesInvalidExpectedContextAndBudget)$' -count=1 -v`: PASS (6 focused tests).
- `git diff --check`: PASS.
- Full matrix, CLI, vet, and build: NOT RUN; root owns consolidated checks.

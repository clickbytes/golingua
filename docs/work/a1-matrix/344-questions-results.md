# #344 M8-Q58 results

Source selection: [58 curated pairs](344-questions-selection.md). Contract scope: [brief](344-questions-brief.md).

The A1.9 batch registers 116 directional cases and 58 pair IDs. Its independent yes/no fixture graphs validate against CORE. The fixture package contains no bilingual input strings and no fabricated WH-variable graphs.

## Correction verification

Commands run after the bounded review correction:

```text
go test ./tests/interlingua -run '^TestQuestionExpansion' -count=1
go test ./tests/matrix -run '^TestQuestionFixtureLoaderRoutesOnlyAnchoredContexts$' -count=1
go test ./tests/translation -run '^TestInventory$' -count=1
go test ./tests/languages/en/A1.9 ./tests/languages/pl/A1.9 ./tests/translation/en-pl ./tests/translation/pl-en -run 'QuestionExpansion|OtherDomains/A1.9-' -count=1 -v
```

The CORE fixture proof, anchored/non-deictic loader routing proof, and inventory pass. The scoped native/integration command exits 1 on current language gaps: 252 `GAP_ANALYZE`, 34 `GAP_REALIZE`, 246 `GAP_I`, and 428 `FAIL` markers. It reports zero `SPEC_PENDING` and zero `INFRA`. Raw output: `/tmp/m8-q58-correction.log`; SHA is unrecorded for this working tree. `git diff --check` passes.

Coordinator's pre-correction full matrix observation covered 652 records / 1,956 checks: 255 PASS, 68 FAIL, 510 GAP_ANALYZE, 253 GAP_REALIZE, 870 GAP_I, zero SPEC_PENDING/INFRA. The coordinator's separate pre-correction CLI run produced 111 unsupported and 5 unknown outcomes across 116 directions. Neither the full matrix nor CLI was rerun by this correction worker; these values are reported as coordinator observations, not correction verification.

Language owner acceptance remains `NOT_EVALUATED`. The batch does not certify CEFR coverage or claim question handling is complete. `go vet`, `go build`, smoke CLI and full repository test runs were `NOT RUN` by this worker; broader verification remains with the coordinator.

Coordinator final run (2481b2b+working-tree):652records/1956checks;255PASS,68FAIL,510GAP_ANALYZE,253GAP_REALIZE,870GAP_I;0SPEC_PENDING/INFRA/NOT RUN; matrix exit1. All116 CLI inputs executed:111unsupported/exit3,5unknown/exit5, [responses](evidence/344-questions-cli.json). The English destination question was curated to explicit walking before the final CLI run; this avoids equating generic going with walking. Reviewed and corrected count/reference choices for PL drzwi, allowed-field consistency, known-reference assumptions, and nil-context routing for nondeictic questions. The inventory now rejects required assumption fields that are not allowed. Vet/build/help/version/diff passed. Language gaps remain; owner acceptance NOT_EVALUATED.

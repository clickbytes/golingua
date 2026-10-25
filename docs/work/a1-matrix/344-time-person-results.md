# #344 M8-TIME-PERSON — results

Added the first 16 authored time agreement pairs as 32 directional contracts, with source fixture links, per-frame English `read` meanings, age-child/offspring alternatives, target candidate requirements, and local number/gender/determiner assumptions. The remaining 40 A1.10 rows stay queued. Owner acceptance: `NOT_EVALUATED`.

## Checks and observations

- `gofmt` completed for the four new Go files and the scoped registry/loader files.
- `go test ./tests/translation -run '^TestInventory$' -count=1 -v` — PASS; all 342 pair registrations and directional links are structurally valid.
- Focused `TestTimeExpansion` proofs ran in the combined native/integration command — PASS; CORE/reference checks and explicit personal-time variants are valid.
- `go test -json ./tests/languages/en/A1.10 ./tests/languages/pl/A1.10 ./tests/translation/en-pl ./tests/translation/pl-en -run 'TestTimePerson|TestOtherDomains$' -count=1` — native and integration packages exit nonzero on current language gaps. The A1.10 observations include 32 `GAP_ANALYZE` and 17 `GAP_REALIZE` occurrences; captured output has zero `SPEC_PENDING` and zero `INFRA` diagnostics. `TestOtherDomains` also runs older unrelated domains.
- `git diff --check` — PASS.

These observations are agent-run results, not owner language acceptance. The run did not execute the full matrix, CLI, vet, or build.

Coordinator acceptance of this component: independent review found no discrepancy in16 pairs/32 records. Full local run source dae7b6b+working-tree:684records/2052checks;255PASS,80FAIL,574GAP_ANALYZE,273GAP_REALIZE,870GAP_I;0SPEC_PENDING/INFRA/NOT RUN, exit1. All32 CLI calls executed:26unsupported/exit3,6unknown/exit5; [responses](evidence/344-time-person-cli.json). Vet/build/help/version/diff passed. The separate earlier TIME56 assignment returned PARTIAL without edits and was narrowed to an independent-I component plus this16-pair component; remaining40 are still queued in #344, not reported complete. Independent fixture review added row22 present alongside its past EN primary and removed duplicate age aliases; all four focused proofs pass. Source I never receives guessed you-count or biological sex. Owner acceptance NOT_EVALUATED.

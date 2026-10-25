# #344 M8-TIME-OTHER — results

Added the remaining 40 A1.10 pairs as 80 directional records. Ordinary cases use the existing `TimeExpansion` I variants; rows 49–56 carry explicit `RepresentationGap` contracts and no fabricated fixture graphs. Owner acceptance: `NOT_EVALUATED`.

## Checks and observations

- `gofmt` completed for the four new language files and scoped translation registry files.
- `go test ./tests/translation -run '^TestInventory$' -count=1` — PASS; registrations resolve and inventory links are consistent.
- `go test ./tests/interlingua -run '^TestTimeExpansion' -count=1` — PASS; existing temporal fixture proofs validate against CORE and references.
- `go test -v ./tests/languages/en/A1.10 ./tests/languages/pl/A1.10 ./tests/translation/en-pl ./tests/translation/pl-en -run 'TestTimeOther|TestOtherDomains$' -count=1` — native/integration packages exit nonzero on expected current language gaps. The command also selects earlier unrelated cases through `TestOtherDomains`. Captured output contains zero `SPEC_PENDING` and zero `INFRA` diagnostics.
- Rows 49–56 logged `independent_realization_executed:false` for both language directions with the stated future/progressive/frequency/overlap reason; no independent target graph was fabricated.
- `git diff --check` — PASS.

The full matrix, CLI, vet, and build were not run. These are agent observations, not owner language acceptance.

Coordinator full run source83d7bc3+working-tree:764records/2292checks;255PASS,92FAIL,694GAP_ANALYZE,333GAP_REALIZE,918GAP_I;0SPEC_PENDING/INFRA/NOT RUN, exit1. All80 CLI calls RUN:73unsupported/exit3,3incomplete/exit7,1ambiguous/exit6,3unknown/exit5. [Responses](evidence/344-time-other-cli.json). Statuses are engine observations, not automatic quality acceptance. Independent read-only review found no discrepancy in the40 pairs, time/child alternatives or door-count contracts. Vet/build/help/version/diff passed. All56 TIME pairs are now registered; owner acceptance NOT_EVALUATED.

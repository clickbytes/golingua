# #344 correction — results

Added explicit `.pl` variants for negation fixture IDs `a1.8.neg.009`, `.025`, and `.038`. Each keeps the original base ID's known-one English graph and changes only door multiplicity to empty for the Polish plural-only source. Polish analysis and corresponding English realization now reference `.pl`; English analysis and corresponding Polish realization retain the base ID. Owner acceptance: `NOT_EVALUATED`.

The English targets from unknown-count Polish I include doctor a/the × door a/the/bare and door number alternatives for pairs 9 and 38, and singular/plural article choices for pair 25. Per-text determiner and number assumptions license the choices; person remains optional.

## Checks and observations

- `gofmt` completed for the four scoped Go files.
- `go test ./tests/interlingua -run '^TestNegationExpansion' -count=1` — PASS; the focused proof validates CORE, unchanged predicate/frame/roles and negation scope, base door=one versus `.pl` door count empty, and rejects `.pl` aliases outside the three authorized IDs.
- `go test ./tests/translation -run '^TestInventory$' -count=1` — PASS.
- `go test -v ./tests/languages/en/A1.8 ./tests/languages/pl/A1.8 -run 'TestNegation(Expansion(Analysis|Realization))?/(A1\\.8-NEG-(817|818|849|850|875|876))' -count=1` — expected language gaps; zero `SPEC_PENDING` and zero `INFRA` diagnostics.
- `go test -v ./tests/translation/en-pl ./tests/translation/pl-en -run '^TestOtherDomains$' -count=1` — the existing integration wrapper selects A1.8 under this broad group; these records show unsupported/unknown language behavior. Zero `SPEC_PENDING` and zero `INFRA` diagnostics. Other unrelated domains also run.
- `git diff --check` — PASS.

No engine behavior or owner acceptance is claimed. Full matrix, CLI, vet, and build were not run.

Coordinator review confirmed that only PL source door multiplicity and its EN choices changed; English source I retains one. Final full matrix source17d85e8+working-tree:764records/2292checks, exit1;255PASS,92FAIL,694GAP_ANALYZE,333GAP_REALIZE,918GAP_I;zeroSPEC_PENDING/INFRA/NOT RUN. Record statuses are unchanged from the previous run. All6 affected CLI inputs RUN:5unsupported/exit3 and1unknown/exit5, [responses](evidence/344-door-count-cli.json). Vet/build/help/version/diff passed. This is a contract correction, not an engine improvement or owner approval. Owner NOT_EVALUATED.

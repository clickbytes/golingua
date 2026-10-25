# #344 AMB43 — results

Added 43 pairs / 86 directions. Canonical variants include English present/past reads (9–14), age-child/offspring readings for rows 12, 15–20, 31–32, 41, Polish role swaps for rows 21–26, two bound anchors for rows 27–32, and entity-level exact-two quantifiers for rows 35, 36, 41. Full gaps are 33, 34, 37–40, 43; the PL row 37 contract preserves both readings while English row 37 requires absence only. OwnerAcceptance: `NOT_EVALUATED`.

Controls and observed language-run outcomes are reported below after execution.

## Observed controls

- `gofmt` on the changed Go files: completed.
- `go test ./tests/interlingua -run '^TestAmbiguityExpansion' -count=1`: PASS; 4 focused independent-I tests.
- `go test ./tests/translation -run '^TestInventory$' -count=1`: PASS.
- `go test ./tests/languages/en/A1.12 ./tests/languages/pl/A1.12 ./tests/translation/en-pl ./tests/translation/pl-en -run 'TestAmbiguity|TestOtherDomains' -count=1 -v`: EXIT 1 on expected language-contract failures. AMB43 native records emitted 28 declared `GAP_I` observations; integrated records emitted 14. No AMB43 `SPEC_PENDING` or `INFRA` diagnostics were observed. `TestOtherDomains` also exercises other registered domains, so its exit is not an isolated AMB43 result.
- `git diff --check`: PASS.
- Full matrix, CLI, vet, and build: NOT RUN (outside this component).

After this addition the declared inventory is 467 pairs / 934 directions; `ambiguity_reference` has 45 pairs. OwnerAcceptance remains `NOT_EVALUATED`.


## Bounded correction evidence

The source analysis contracts now link every licensed canonical I ID, including English past rows 9–14 and all four row-12 frame/child combinations. Polish rows 21–26 require both base and `.swap`; target English text vectors are role-correct SVO. Polish deictic targets use the supplied per-role reference vectors, including separate offspring vectors for rows 31–32. A standard-Go catalogue-link proof compares each source AnalysisCase’s allowed/required IDs directly with `AmbiguityExpansionIDs`; full-gap rows have no fixture IDs.

Coordinator final review: after the bounded worker correction, PL base targets still contained36 truncated initials. Root restored the72 Allowed/Required lists directly from authored full PL sentences, with the stated generic teacher target for13, and specialized both EN-source absence gap strings for37. No runtime output was used to author these repairs. Final full source913a003+working-tree:934records/2802checks, exit1;291PASS,248FAIL,866GAP_ANALYZE,395GAP_REALIZE,1002GAP_I;0SPEC_PENDING/INFRA/NOT RUN. All86 unchanged CLI inputs previously ran:44unsupported/exit3,30incomplete/exit7,12ambiguous/exit6; [responses](evidence/344-ambiguity-cli.json). Fixture/catalogue/inventory checks, vet/build/help/version/diff passed. Source CLI inputs/product were unchanged by contract correction. Owner NOT_EVALUATED.

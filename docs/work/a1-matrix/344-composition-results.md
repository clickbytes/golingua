# A1.13 composition-boundary results

Added 33 selected pairs and 66 directions as `A1.13-COMP-1301..1366`, PairIDs `a1.13-comp-001..033`, group `composition_boundary`, registered in `A1.13-composition`. Native source and target records share primary MeaningIDs and semantic requirements. Child row 4 uses base age-child and `.offspring` with separate target sentence contracts. Row 33 retains one mother referent across both seeing roles.

Request targets preserve ActRequest without asserting event time. EN-source imperatives include Polish singular/plural forms with `number` assumptions, consistent across coordinated actions. PL-source `drzwi` rows 21, 23, and 24 offer English singular/plural door forms with number and determiner assumptions; English-source I retains door multiplicity one. `me`/`mi` in row 26 stays linked to the speaker anchor. The authored selection is preserved in [selection](344-composition-selection.md).

Checks and observations:

- `gofmt` on all assigned new and edited Go files: completed.
- `go test ./tests/interlingua -run '^TestCompositionExpansion' -count=1`: PASS, 3 tests.
- `go test ./tests/translation -run '^TestInventory$' -count=1`: PASS.
- Native `go test ./tests/languages/en/A1.13 ./tests/languages/pl/A1.13 -run '^TestComposition' -count=1 -v`: exit 1, expected language-contract failures and explicit gaps; 52 `GAP_I`, zero `SPEC_PENDING` and `INFRA`.
- Integration `go test ./tests/translation/en-pl ./tests/translation/pl-en -run '^TestOtherDomains$/A1\.13-COMP-' -count=1 -v`: exit 1, expected language-contract failures and explicit gaps; 26 `GAP_I`, zero `SPEC_PENDING` and `INFRA`.
- Full realization strings were inspected across the 66 records; no one-character/scalar-index outputs. `git diff --check`: PASS.

Full-matrix, CLI, vet, and build checks were NOT RUN; root owns final consolidation. Rows 12–18, 22, 25, 27, 29–30, and 32 remain full I gaps. OwnerAcceptance: `NOT_EVALUATED`.

Coordinator review verified exactly1000 unique directions in500 pairs. It found and corrected unknown PL door counts for request21/23/24; the two-source focused proofs now pass. All66 CLI inputs RUN:58unsupported/exit3,8unknown/exit5, [responses](evidence/344-composition-cli.json). Inventory and independent-I checks passed. Full matrix, vet/build/smoke and final CI are consolidated after reporting changes; no PASS claimed yet for those checks. Earlier technical CI40195470643 caught formatting in seven older files; root applied gofmt without semantic changes. Owner NOT_EVALUATED.

# #344 M8-LQ48 — wyniki lokalne

Dodano 48 par / 96 kierunków A1.7; licznik całej macierzy wynosi 223 pary / 446 rekordów. Każda pełna analiza, realizacja i integracja pozostaje `GAP_I`; powierzchnie w danych są propozycjami, nie zatwierdzonymi goldami.

Kontrole uruchomione:

- `go test ./tests/translation -run '^TestInventory$' -count=1` — PASS.
- `go test ./tests/interlingua -run 'Test(QuantifierAndReferentCannotCarryExistenceTime|CoreExistenceCannotBindCountedReferent|WomanConceptIsNotAvailableForCountContracts|PendingDomainRepresentableMeaning)$' -count=1` — PASS.
- `git diff --check` — PASS.

Pełna macierz, CLI, `go vet`, `go build` i smoke CLI: `NOT RUN` w tym zleceniu; wykonuje je koordynator.

Coordinator: all96 CLI inputs RUN:88 unsupported/exit3 and8 unknown/exit5; [responses](evidence/344-location-cli.json). Full local matrix source c4e47c3+working-tree:446records/1338checks, exit1, raw log `/tmp/golingua-matrix-expansion/m8-location-full.log`. Vet/build/help/version/diff and focused proofs/inventory passed. Independent review found lowercase English article alternatives; coordinator corrected sentence capitalization. Also corrected negative HAVE assertion wording from positive to negated, without changing expectations. Owner acceptance NOT_EVALUATED.

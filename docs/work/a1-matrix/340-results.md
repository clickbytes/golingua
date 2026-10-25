# Wyniki A1-M4 #340

Base: `78bba8dbf9cec5d09981f0ae1d1c7d2026d5d393`. Kontrola wykonawcy dotyczy `base+working-tree`; akceptacja językowa właściciela: `NOT_EVALUATED`. Wszystkie 16 par pozostaje oznaczonych `draft_unreviewed`.

## Weryfikacja

- `go test ./tests/interlingua -run 'Corpus(EventFixturesValidateAndPreserveAuthoredStructure|MeaningAlternativesRequireEachAllowedReading)' -count=1` — PASS; 16 podstawowych i 4 alternatywne grafy walidują się względem CORE. Wąski test potwierdza rozpoznanie obu dozwolonych grafów, wykrycie brakującego wymaganego odczytania i odmowę grafu spoza zbioru.
- `go test ./tests/languages/en/A1.1 ./tests/languages/pl/A1.1 -run CorpusRealization -count=1` — PASS; 32 przypadki realizacji.
- `go test ./tests/translation -run Inventory -count=1` — PASS; 32 nowe kierunkowe ID mają kompletne, jednoznaczne odwołania.
- `GOLINGUA_MATRIX_SHA=78bba8dbf9cec5d09981f0ae1d1c7d2026d5d393+working-tree go test ./tests/... -count=1 -v` — exit 1. Raport: 56 rekordów, 168 kontroli; PASS 109, FAIL 23, GAP_ANALYZE 6, GAP_REALIZE 6, GAP_I 24, SPEC_PENDING 0, INFRA 0, NOT RUN 0. W nowych 32 rekordach: analysis PASS 14 / FAIL 18, realization PASS 32, integration PASS 27 / FAIL 5. Odczyt `offspring` jest dopuszczony tylko z grafem `child_of_person`; wymagane alternatywy nieobecne w analizie pozostają FAIL. FAIL/GAP nie zostały dostrojone do runtime.
- Reporter dla końcowego logu — PASS; rozpoznaje 56 jawnych ID, tagi pokrycia i trzy warstwy statusu. Surowy log: `/tmp/golingua-m4-final.log`; raport: `/tmp/golingua-m4-final-report.md`.
- `git diff --check` i `gofmt` — PASS.
- `go vet ./...`, build CLI/smoke i CI PR/main — NOT RUN przez wykonawcę; pozostają do weryfikacji koordynatora.

Koordynator zaobserwował niezależnie 32 wejścia CLI: EN→PL kończy się `incomplete` (exit 7), PL→EN zwraca `ambiguous` (exit 6) z wariantami rodzajników. To obserwacje runtime, nie źródło kontraktów ani zatwierdzenie zdań. Dowody koordynatora znajdują się w `docs/work/a1-matrix/evidence/340-cli.json`.

Ograniczenie: zakres podnosi 16 wybranych par korpusu do wykonywalnej macierzy w obu kierunkach. Dla czterech zdań z `dziecko/child` zachowuje dwa wymagane odczytania. Nie zatwierdza pozostałych 84 par ani językowego statusu wybranych propozycji i nie dodaje gramatyki.

Redakcja integracyjna koordynatora: usunięto automatyczne przepisywanie warunków wariantu według pozycji tekstu w tablicy. Warunki per-Meaning są jawne; dla tekstów offspring zapisano właściwe wartości rodzajników. To ograniczona korekta integracji po jedynym review wykonawcy, bez kolejnego przebiegu implementacyjnego.

Koordynator: gofmt, `go vet ./...`, build, `--help`, `--version`, diff-check PASS. Pełny przebieg po integracji z `GOLINGUA_MATRIX_SHA=78bba8dbf9cec5d09981f0ae1d1c7d2026d5d393+working-tree` i raport zachowały powyższe wyniki 109/23/36 GAP. Surowy log `/tmp/golingua-matrix-expansion/m4-full.log`. Same poprawne teksty CLI nie dowodzą kompletności analizy: niektóre wyniki zawierają PARTIAL_PARSE_EVIDENCE/UNRESOLVED_ROLE_SENSE, a EN pomija wymagane odczytanie potomstwa dla child. Te rozbieżności nadal są widoczne.

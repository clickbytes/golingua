# #344 M8-PO34 — wyniki lokalne

Status implementacji: deklaratywne kontrakty dodane; brakujące pełne grafy pozostają jawne jako `GAP_I`. Nie przedstawiamy tych zdań jako zatwierdzonych językowo.

- 34 pary / 68 rekordów kierunkowych, 175 par / 350 kierunków w całej macierzy.
- Dodano rejestrację A1.6 i sprawdzenie, że `temporary_possession`, `legal_ownership`, `siblinghood` i `friendship_relationship` nie mogą być użyte jako predykaty stanów Meaning.
- Zaktualizowano SemanticRequirement w sześciu istniejących parach HAVE, bez zmiany ich zdań, identyfikatorów ani propozycji tłumaczeń.
- Wynik CLI na 68 wejściach: 68/68 uruchomiono lokalnie; wszystkie zwróciły `unsupported` (exit 3), a nie poprawne tłumaczenie. Pełne odpowiedzi: [evidence](evidence/344-possession-cli.json).
- Właścicielska akceptacja językowa: `NOT_EVALUATED`.

## Kontrole lokalne

- `go test ./tests/translation -run '^TestInventory$' -count=1`: PASS.
- `go test ./tests/interlingua -run TestPossessionAndKinshipRelationsAreNotMeaningStates -count=1`: PASS.
- `go test ./tests/... -count=1 -v`: exit 1; raport wygenerowano z logu. Wszystkie 68 nowych integracji mają `GAP_I`; lokalne adaptery nie dostarczyły licencjonowanego pełnego znaczenia. Pozostałe niezaliczone testy są dotychczasowymi jawnie raportowanymi lukami macierzy.
- Raport macierzy: 350 kierunkowych rekordów, 68 nowych wierszy A1.6; nowe rekordy pokazują `GAP_I` dla analizy, realizacji i integracji.
- CLI `cmd/golingua` na 68 wejściach: RUN; 68 × `unsupported`/exit 3. Nie jest to PASS tłumaczenia.

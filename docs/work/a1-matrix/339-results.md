# Wyniki A1-M3 #339

Base kodu: `cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd`. Ocena agenta: `DONE` dla zamknięcia kontraktów i wykazania reprezentowalności/braków. Akceptacja językowa właściciela: `NOT_EVALUATED`.

I zawiera poprawny graf `Question` → klasyfikacja kota dla pytań yes/no z zachowanym aktem pytania, kotwicą proximal/unresolved i nierozstrzygniętą referencją. Dla dokładnie dwóch kotów I zawiera `QuantifierExactly(2)` nad referentem `cat` o multiplicity `many`; wymagany otwarty domain dyskursu pozostaje jawny. Kontrole wykazują, że Meaning nie przyjmuje `located_in` ani `temporary_possession` jako `State` (`CONCEPT_KIND_MISMATCH`); rola `located_in` w CORE to `occupant/location`, a tymczasowego posiadania `holder/object`. On-support wymaga innego predykatu/relacji i nie jest utożsamione z `located_in`. Posiadanie ma też wymagane wiązanie osoby mówiącej lub zaimka żeńskiego. WH-location pozostawia miejsce niewypełnione, a jego brak reprezentacji wynika z relacji KindRelation.

Lokalne kontrole:

- `go test ./tests/interlingua ./tests/matrix ./tests/translation -run 'PendingDomain|Inventory|RequiredSurface' -count=1 -v` — PASS.
- `GOLINGUA_MATRIX_SHA=cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd go test ./tests/... -count=1 -v` — exit 1; polecenie użyło SHA bazowego mimo zmodyfikowanego working tree, więc wynik jest wstępną obserwacją `base+working-tree`, a nie odtwarzalnym wynikiem dla samego base. Log wykazał 0 `SPEC_PENDING`, 24 kontrole `GAP_I`, 6 `GAP_ANALYZE` i 6 `GAP_REALIZE`. Dodatkowe kontrole testowe: PASS; techniczny produktowy CI nie jest tym wynikiem.
- Reporter na tym samym surowym logu z rzeczywistym exit code — PASS; 12 rekordów zachowało statusy per analiza/realizacja/integracja, `GAP_I` jest statusem wiersza, a surowe warstwowe GAP pozostają w logu.
- `git diff --check` — PASS. `gofmt` — PASS.

Rzeczywiste CLI dla wszystkich 12 zdań na bazowym runtime jest zapisane przez koordynatora w `docs/work/a1-matrix/evidence/339-cli.json`: 12/12 zwróciło `unsupported`, exit 3. To obserwacja silnika, nie akceptacja przypadków ani expected. Nie zostały uruchomione: pełny `go vet ./...`, build CLI/smoke i CI PR/main — `NOT RUN` przez wykonawcę; koordynator przejmuje te kontrole.

Limity: nowe kontrakty opisują brakującą semantykę, nie dodają gramatyki. Zakres nie certyfikuje A1 ani nie zatwierdza językowo zdań. Kolejny etap pozostaje QUEUED do odbioru tego PR i zamknięcia child.

## Weryfikacja koordynatora

Po review wykonano `go vet ./...`, build CLI, `--help`, `--version`, gofmt i diff-check: PASS. Powtórzony pełny przebieg z `GOLINGUA_MATRIX_SHA=cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd+working-tree` zakończył się exit 1; raport: 36 PASS, 24 GAP_I, 6 GAP_ANALYZE, 6 GAP_REALIZE, 0 SPEC_PENDING/FAIL/INFRA, dodatkowe 152 kontrole PASS. Lokalny log: `/tmp/golingua-matrix-expansion/m3-full.log`; raport: `m3-report.md`. Niezależna realizacja ośmiu rekordów z GAP_I nie została wywołana — brak legalnego grafu, jawne `independent_realization_executed:false` w logu. Analiza i integracja tych rekordów zostały wykonane. CI PR/main potwierdza issue #339.

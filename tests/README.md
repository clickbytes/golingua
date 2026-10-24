# Lokalna macierz znaczenia i języków

Na polecenie właściciela macierz A1 wykonujemy standardowym `go test`, **poza CI językowym**. Techniczne CI nadal sprawdza formatowanie, vet, build i dotychczasowe smoke. Próbka to 12 par zdań i 24 rekordy kierunkowe; nie jest pełnym A1 ani certyfikacją CEFR. Kontekst tematów: [deskryptory Rady Europy](https://rm.coe.int/16809ea0d4). Zdania są własną redakcją, bez porównań do JavaScriptu.

## Dane i odpowiedzialności

```text
tests/
  interlingua/              # niezależne, ręcznie zapisane wzorce I
  languages/pl/A1/          # polska analiza i polska realizacja
  languages/en/A1/          # angielska analiza i angielska realizacja
  translation/cases.go      # 24 stabilne ID i odwołania do obu stron
  translation/pl-en/        # kontrole PL→I→EN
  translation/en-pl/        # kontrole EN→I→PL
  matrix/                  # wspólne typy, walidowany runtime i pomocniki
```

`*.testdomain.go` to zwykłe dane Go; `*_test.go` wykonuje kontrole. Przyszłe A1.1 i A1.2 oznaczają kolejne partie A1 konkretnego języka i wymagają osobno określonego zakresu.

Każdy oryginalny rekord zachowuje ID w `translation.Cases`. Odwołuje się do analizy `ID:analyze` w języku źródłowym oraz realizacji `ID:realize` w docelowym. Analiza przechowuje dokładne wejście, `MustPreserve`, nierozstrzygnięcia i pochodzenie wejścia. Realizacja przechowuje osobno `ProposalTexts`, autorstwo propozycji i ewentualny zamknięty kontrakt `AllowedTexts` oraz `AllowedAssumptionFields`. Brak wzorca I nie jest dowodem błędu parsera. Propozycje innych domen nie stają się automatycznie asercjami tekstowymi.

W A1-T01 zamknięto dwa znaczenia kontrolne: `seeing.teacher-dog` i `seeing.dog-teacher`. Wzorce zawierają zdarzenie `concept:seeing`, role `experiencer`/`stimulus` i odrębne referenty. Nie ustalają konkretnej tożsamości, czasu, liczby ani angielskiej określoności. Realizacja z samego I musi ujawnić potrzebne założenia; angielski zachowuje cztery warianty rodzajników. Zamiana ról zmienia kontrakt.

`GraphMatches` porównuje strukturę z bijekcją lokalnych NodeID, niezależnie od kolejności węzłów i ról. Zachowuje współdzielenie węzłów i odrębność referentów o tym samym ConceptID. To porównanie kontraktu z nierozstrzygniętymi referentami, **nie dowód identyczności osobników** ani zamiennik `meaning.SemanticEquivalence`. Wzorce powstają bez analizatora, a realizacja nie uruchamia analizy innego języka.

Runtime kompiluje `bundled.Candidate`, waliduje pakiet i rejestr języków, następnie używa produkcyjnych konstruktorów PL/EN z CLI. Nie wymaga konfiguracji ani aktywnego store użytkownika.

## Uruchamianie i odczyt wyniku

Z katalogu głównego zapisz SHA wraz z poleceniem; dla niezatwierdzonych zmian dodaj opis working tree:

```sh
export GOLINGUA_MATRIX_SHA="$(git rev-parse HEAD)+working-tree"
go test ./tests/interlingua/... -run Seeing -count=1 -v
go test ./tests/languages/pl/A1/... -run Seeing -count=1 -v
go test ./tests/languages/en/A1/... -run Seeing -count=1 -v
go test ./tests/translation/... -run Seeing -count=1 -v
go test ./tests/... -count=1 -v
```

Po commicie użyj dokładnego SHA bez `+working-tree`. Kontrole adapterów logują SHA oraz surowy wynik z kandydaturami, grafami, tekstami, założeniami, stratami i diagnostykami. Bez zmiennej SHA log zawiera `UNRECORDED`; nie przypisuj takiego logu do konkretnego commita. Polecenie, exit code i ocenę zapisz obok logu. Obserwacje nie nadpisują oczekiwań.

| Ocena | Znaczenie |
| --- | --- |
| PASS | Wykonany przypadek spełnia kontrakt; standardowe `go test` pokazuje PASS |
| FAIL | Wynik narusza kontrakt |
| GAP_ANALYZE | Brak wymaganej analizy języka |
| GAP_REALIZE | Brak wymaganej realizacji języka |
| GAP_I | I nie potrafi wyrazić uzgodnionego znaczenia |
| SPEC_PENDING / NOT RUN | Brak niezależnego, zamkniętego wzorca do porównania |
| INFRA | Runtime lub środowisko uniemożliwia wykonanie |

Status silnika (`ok`, `ambiguous`, `incomplete`, `unsupported`) jest oddzielony od oceny. Samo `ok` nie daje PASS. Nie używamy `t.Skip`, procentowych progów ani automatycznej akceptacji propozycji. Pełne `go test ./tests/...` ma niezerowy exit code, dopóki jakiekolwiek zadeklarowane wymaganie pozostaje niezaliczone. Inwentarz kontroluje wszystkie 24 ID i odwołania, również bez wzorców I.

Akceptacja językowa właściciela pozostaje `NOT_EVALUATED`. Zamknięcie kontraktu i ocena techniczna przez agenta nie oznaczają jego akceptacji.

## Odtworzenie obserwacji CLI

Macierz nie potrzebuje store. CLI używa jawnie przygotowanego snapshotu; przykład zapisuje go poza repozytorium:

```sh
go build -o /tmp/golingua-a1 ./cmd/golingua
/tmp/golingua-a1 snapshot --apply --store /tmp/golingua-a1-store --json
cat > /tmp/golingua-a1-runtime.json <<'JSON'
{"schema_version":1,"dataset_store":"/tmp/golingua-a1-store","dataset_id":"golingua-authored","language_package_dirs":[]}
JSON
/tmp/golingua-a1 translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config /tmp/golingua-a1-runtime.json --json
/tmp/golingua-a1 translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config /tmp/golingua-a1-runtime.json --json
```

W obserwacji A1-T01 na bazie `a922fff84a663976ec46324ff6a7a9359665e38e` EN→PL zwrócił `ok`/exit 0 i `Nauczyciel widzi psa.`; zamiana ról dała `Pies widzi nauczyciela.`. PL→EN zwrócił `ambiguous`/exit 6 i cztery warianty `a/the` dla obu uczestników, także po zamianie ról. Ocena agenta: zachowane role i jawna niejednoznaczność rodzajników. Właściciel: `NOT_EVALUATED`. Surowe JSON sesji: `/tmp/golingua-a1-t01-cli.json` (lokalny dowód, nie fixture produktu).

Pierwsza dostawa danych: [#309](https://github.com/clickbytes/golingua/issues/309). Fundament: [#312](https://github.com/clickbytes/golingua/issues/312); epik: [#311](https://github.com/clickbytes/golingua/issues/311). Szczegóły: [CLI](../docs/cli.md), [przegląd ręczny](../docs/manual-review.md).

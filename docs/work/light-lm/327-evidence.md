# Odbiór LIGHT-F3

Koordynator wykonał odbiór 2026-10-25 (Europe/Warsaw) na `3505a7aebabacad7b999f9447d79b6c28f2faf2c`, Go 1.27.0, Linux amd64. Ocena agenta: PASS dla zamkniętego [kontraktu #327](327-contract.md). Akceptacja językowa właściciela: NOT_EVALUATED. Kolejny commit zawiera tylko dowody i ich indeks. Merge i rzeczywiste CI końcowego main są osobną bramką, zapisaną w issue/PR i następnym checkpointcie rejestru repo.

## Wykonane kontrole

| Kontrola | Wynik |
| --- | --- |
| `gofmt -l` wszystkich śledzonych Go, `git diff --check` | PASS |
| `go vet ./...` | PASS, exit 0 |
| `go build -o /tmp/golingua-light-f3 ./cmd/golingua` | PASS, exit 0 |
| `--help`, `--version`, `catalog --summary --json`, `catalog --details --json` | PASS; wersja 0.1.0-alpha.1 |
| `forms --language pl --lexeme lexeme:pl.pies.n --feature number=plural --feature case=genitive --feature gender=masculine_animal --json` i `forms --language pl --surface psów --json` | PASS |
| `go test ./tests/... -run 'Composition\|Generalization\|MeaningAttributes\|Identification\|Negation\|Seeing\|Inventory' -count=1 -v` | PASS, exit 0 |
| `go test ./tests/... -count=1 -v` | FAIL, exit 1: wyłącznie 36 SPEC_PENDING, 0 INFRA, wszystkie 24 oryginalne ID obecne |

Testy miały `GOLINGUA_MATRIX_SHA=3505a7aebabacad7b999f9447d79b6c28f2faf2c`; surowe logi `/tmp/light-f3-{selected,full,vet,build}.log`. Lokalny katalog wykazał 135 zatwierdzonych wiązań konstrukcji (PL12, EN123); poprawiono nieaktualną liczbę w dokumentacji. Testy językowe nie zostały dodane do CI.

Niezaliczone kontrole pełnej macierzy nadal dotyczą analizy, realizacji i integracji 12 rekordów: `A1-LQ-001-EN-PL`, `A1-LQ-002-PL-EN`, `A1-LQ-003-EN-PL`, `A1-LQ-004-PL-EN` oraz analogicznych `A1-PO-001–004` i `A1-Q-001–004`. Powodem jest brak zamkniętego niezależnego I, a nie udowodniony błąd parsera. Sześć nadrzędnych testów raportuje te 36 niezaliczonych subtestów. Brak innych niepowodzeń i pominięć. To nie jest zaliczenie pełnego A1.

## Znaczenie i realizacja

Cztery wzorce Composition zapisano ręcznie, przed adapterami. Kontrola porównuje cały graf przy jednej bijekcji NodeID i nieuporządkowanych korzeniach; nie porównuje dwóch niezależnych projekcji. Alpha-renaming i przestawienie węzłów zachowują zgodność. Zmiana roli, zgubiona cecha, błędny bearer i scalenie dwóch kotów są wykrywane. Oba referenty pozostają nierozstrzygnięte, także przy równym ConceptID.

Analiza i realizacja każdego języka działają osobno. Realizacja dostaje niezależne I bez parsera drugiego języka. Kontrole obejmują też brak/past czasu, brak/many wielości, czasowany stan cechy, drugi modyfikator, negację złożonej frazy, zakotwiczony i sprzeczny kontekst. Poprawne strukturalnie i względem CORE nieobsługiwane I daje brak tekstu z jawną odmową. Wykryty podczas pracy defekt EN (sam `TARGET_ALTERNATIVES_AMBIGUOUS` po odrzuceniu wszystkich powierzchni) usunięto przez `TARGET_LEXICAL_COVERAGE_UNAVAILABLE` i Unsupported, zachowując wcześniejsze diagnostyki. Kontrolne proste Seeing nadal daje tekst.

Izolowane kopie analizatora i indeksu wykazały: pending mapa black → `MODIFIER_TRANSFER_UNRESOLVED`, competing pending mapa → `MODIFIER_TRANSFER_AMBIGUOUS`; żaden wariant nie materializuje I ani nie pomija AP. Dodatnia kontrola przed i po wykazała, że współdzielony runtime nie został zmodyfikowany.

Identyfikacja psa przechodzi trzy niezależne analizy i trzy realizacje w każdym języku (klasyfikacja, black, negacja). EN odmawia niezależnego klasyfikującego I dla water/money i eye bez licencjonowanego `a`. Przypadki nie zmieniają oryginalnych 24 rekordów.

Osiem integracji porównuje pośrednie I każdego kandydata z wzorcem, dokładny zestaw tekstów, brak strat i jawne założenia. Dla PL→EN każdy wariant ma dokładnie dwa determiner z właściwą wartością i ścieżką do odpowiedniego referenta przez rolę; kot/kot nie jest wyjątkiem.

## Rzeczywiste CLI

[40 obserwacji JSONL](327-observations.jsonl) zawiera SHA, dokładne polecenia, exit, status, kandydatury, znaczenia, teksty, założenia, straty i diagnostyki. To obserwacje, nie źródło oczekiwań; obszerne Evidence parsera pominięto. Nowe kombinacje są autorskimi próbami agenta w zamkniętym kontrakcie, nie akceptacją właściciela.

| Wejście EN→PL | Tekst |
| --- | --- |
| The black cat sees the dog. | Czarny kot widzi psa. |
| The black dog sees the cat. | Czarny pies widzi kota. |
| The black bird sees the dog. | Czarny ptak widzi psa. |
| The black cat sees the cat. | Czarny kot widzi kota. |

Te cztery wyniki mają `incomplete` / exit7 z jawnym person=third, bez utraty cechy lub uczestnika. Odwrotne cztery wejścia PL dają `ambiguous` / exit6 oraz dokładnie cztery kombinacje `A/The black … sees a/the …`, z założeniami określoności. Status silnika nie jest automatyczną oceną jakości.

Sześć przepływów psa daje `To jest pies.`, `Ten pies jest czarny.`, `To nie jest pies.` oraz warianty EN this/that/it zgodne z nierozstrzygnięciem polskiego wskazania. Cztery oryginalne koty zachowały teksty i kontrakt deixis. Kontrole Reading zachowały present/past i dwie interpretacje mnogiego `read`; negacja, role dog→teacher i pieniądze zachowały znaczenie.

Wszystkie 12 wejść CLI z #326 zachowało dokładnie status, teksty, grafy, założenia, straty i diagnostyki. Fish nadal zachowuje dwa odczytania liczby EN oraz znany brak `PRINCIPAL_PARTS_UNAVAILABLE` dodatkowego sensu PL→EN; nie zaliczono tego braku jako obsłużonego wyniku.

Kontrole rodzajnika: `The teacher looks at an offspring.` ponownie daje `Nauczyciel patrzy na dziecko.` po uzupełnieniu istniejącego leksemu o count. `The black cat sees an offspring.` daje `Czarny kot widzi dziecko.`. `A black offspring sees the dog.` ma poprawne I, ale PL nie ma wymaganej nijakiej formy przymiotnika: `unknown`, brak tekstu; niezależny target EN realizuje wszystkie cztery zgodne warianty. `An black offspring…` daje `unsupported`, bez znaczenia. Nie dodano brakującej odmiany ani nowego leksemu.

## Odtworzenie i TUI

Na wskazanym commicie:

```sh
go build -o /tmp/golingua-light-f3 ./cmd/golingua
/tmp/golingua-light-f3 snapshot --apply --store /tmp/golingua-light-f3-store --json
```

Plik `/tmp/golingua-light-f3-runtime.json`:

```json
{"schema_version":1,"dataset_store":"/tmp/golingua-light-f3-store","dataset_id":"golingua-authored","language_package_dirs":[]}
```

Snapshot `5fe8d84d0a4ffbae85ad34d1864015690d32ab2c49112fad26a07450c76d9d77`, aktywowany z potwierdzoną trwałością. Polecenia tłumaczenia są w JSONL. Przykład:

```sh
/tmp/golingua-light-f3 translate --from en --to pl --text 'The black cat sees the dog.' --runtime-config /tmp/golingua-light-f3-runtime.json --json
TERM=xterm-256color XDG_CONFIG_HOME=/tmp/golingua-light-f3-config \
  script -q -f -c /tmp/golingua-light-f3 /tmp/light-f3-tui.typescript
```

Rzeczywista sesja TUI: F5, wpisanie configu, Ctrl+S → `ok · ready:true`. F1, `The black cat sees the dog.` → jeden polski tekst, incomplete; Tab do candidates, Down wybiera, Enter otwiera szczegóły z person=third. Esc zamyka szczegóły, kolejne Esc ustawia prompt. Alt+D i `Czarny kot widzi psa.` → cztery warianty, ambiguous; szczegóły pokazały dwa determiner na odrębnych referentach. Preview wybranego wariantu zachował ambiguous i pokazał jego tekst. `To jest pies.` dało trzy warianty It/That/This; powrót do EN→PL i `This is a cat` dał `To jest kot.`. Ctrl+C zakończyło proces exit0 i przywróciło terminal. Ustawienia użytkownika nie zostały zmienione.

SHA-256 transcriptu: `d2ee2ed78f5c8137c9109b09a5d612600aedb7afb6ab46f051dfed44b545acf0`; binarium: `5f7e7986c9de3ce46f6299a8eaae9e5457969c9680690929f64cd32faef1b854`. Koordynator przejrzał diff i wyniki osobiście. Przebiegi i korekty workerów są jawne w [zleceniach](327-jobs.md). Ten zakres nie dodaje pamięci, uczenia ani dowodu oszczędności energii.

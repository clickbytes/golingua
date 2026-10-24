# Odbiór LIGHT-F2

Wykonano 2026-10-24 na `f21d2b168760a538a6f90deefe8c561b1e070773`, Linux amd64, Go 1.27.0. Koordynator: PASS dla zamkniętego zakresu [#326](https://github.com/clickbytes/golingua/issues/326). Akceptacja językowa właściciela: NOT_EVALUATED. Późniejszy commit dowodów zmienia wyłącznie dokumentację. Końcowe SHA oraz rzeczywiste CI PR/main są osobną bramką zapisaną w issue i następnym checkpointcie rejestru repo.

## Procedury i zachowanie form

Wykonano 124 wywołania CLI przed i po zmianie: 56 generacji PL (cztery leksemy × 14 komórek), 10 komórek `look`, 7 istniejących kontroli oraz analizę wszystkich 51 różnych powierzchni. [Obserwacje JSONL](326-observations.jsonl) zawierają dokładne polecenia, exit codes, formy/odczyty, cechy, ID, provenance i ślady. Bazą porównania był binarny F1 z `99aefcd`, którego kod produkcyjny jest identyczny z main `1707590`.

Zachowano wszystkie statusy, powierzchnie, ID leksemów i komórek, cechy, źródła oraz diagnostyki. Oczekiwane różnice obejmują przypisanie paradygmatu szkoła/ryba/look, ślad procedury i wynikający z jego treści hash odczytu. Pierwsze porównanie wykazało różne hashe odczytów; sprawdzono ich przyczynę i porównano zawartość odczytów bez hasha, zachowując kontrolę provenance i stable IDs. Nie było zmiany interpretacji ani brakującej formy.

Wszystkie 24 proceduralne komórki PL pokazały `select_stem → strip_suffix → append`. `looked`, `printed`, `cooked` pokazały wspólną klasę i `select_stem → append`. Analiza `looked` zachowała `past_finite` i `past_participle`. `rzece` nadal ma celownik i miejscownik liczby pojedynczej, `szkoły` zachowuje swoje cztery odczyty. `szkół`, `psów`, `kocie` i `ptakiem` pozostały poprawnymi jawnymi formami.

| Leksem | Sześć form proceduralnych: sg acc/ins/voc, pl dat/ins/loc |
| --- | --- |
| szkoła | szkołę, szkołą, szkoło, szkołom, szkołami, szkołach |
| ryba | rybę, rybą, rybo, rybom, rybami, rybach |
| rzeka | rzekę, rzeką, rzeko, rzekom, rzekami, rzekach |
| mewa | mewę, mewą, mewo, mewom, mewami, mewach |

## Przepływy zdaniowe

12 wejść CLI zachowało względem base status, teksty, pośrednie grafy, założenia, straty i diagnostyki. S051 pochodzi z korpusu repo; negacja, liczba mnoga i kombinacje szkoła/ryba/rzeka są jawnie oznaczonymi próbami agenta.

- `The teacher looks at the book.` → `Nauczyciel patrzy na książkę.`; odwrotnie cztery warianty `A/The teacher ... a/the book` z jawną określonością.
- `The teacher does not look at the book.` → `Nauczyciel nie patrzy na książkę.`; negacja zachowana także w drugą stronę.
- `The teachers look at the book.` → `Nauczyciele patrzą na książkę.`; PL→EN zachowuje mnogie `teachers` i dwa warianty rodzajnika obiektu.
- `The teacher sees the school/river.` → `Nauczyciel widzi szkołę/rzekę.`; obie formy powstały proceduralnie. School zachowuje dwa istniejące odczytania o identycznym tekście, bez ich arbitralnego scalenia.
- `The teacher sees the fish.` zachowuje dwa znaczenia liczby: `Nauczyciel widzi rybę.` i `Nauczyciel widzi ryby.`. `Nauczyciel widzi rybę.` daje istniejące warianty EN dla sensu zwierzęcia; dodatkowy sens nadal kończy się `PRINCIPAL_PARTS_UNAVAILABLE` bez tekstu. Ten brak istniał na base i nie jest zaliczony jako obsłużone tłumaczenie. Nie dodawano form dla innego sensu w ramach #326.

`incomplete` i `ambiguous` są statusami silnika; nie utożsamiono ich z oceną jakości. JSONL zachowuje również kandydata bez tekstu oraz jego diagnostykę.

## Kontrole

| Polecenie | Wynik |
| --- | --- |
| `gofmt -l` wszystkich śledzonych plików Go | PASS, pusty wynik |
| `git diff --check` oraz lokalne linki dokumentacji | PASS |
| `go vet ./...` | PASS, exit 0 |
| `go build -o /tmp/golingua-light-f2 ./cmd/golingua` | PASS, exit 0 |
| `--help`, `--version`, `catalog --summary --json`, `catalog --details --json` | PASS, wersja 0.1.0-alpha.1 |
| `go test ./tests/... -run 'MeaningAttributes\|Identification\|Negation\|Seeing\|Inventory' -count=1 -v` | PASS, exit 0 |
| `go test ./tests/... -count=1 -v` | FAIL, exit 1: dokładnie 36 SPEC_PENDING, 0 INFRA, wszystkie 24 ID obecne |

Testy uruchomiono z `GOLINGUA_MATRIX_SHA=f21d2b168760a538a6f90deefe8c561b1e070773`. Niezaliczone kontrole to analiza, realizacja i integracja dla 12 rekordów: A1-LQ-001–004, A1-PO-001–004, A1-Q-001–004, z oryginalnymi kierunkami EN-PL/PL-EN. Powodem pozostaje niezamknięta niezależna specyfikacja I. Nie było innych niepowodzeń ani pominięć. To nie jest ukończenie pełnego A1.

Niezależny review Luna Low nie znalazł rozbieżności PL/EN z kontraktem. Koordynator sprawdził diff i sam wykonał powyższe kontrole. Nie dodano testów, danych masowych ani zmian CI. TUI NOT RUN dla #326: aktywny zakres wymaga CLI i nie zmienia interfejsu TUI.

## Odtworzenie

Na wskazanym commicie:

```sh
go build -o /tmp/golingua-light-f2 ./cmd/golingua
/tmp/golingua-light-f2 snapshot --apply --store /tmp/golingua-light-f2-store --json
```

Config `/tmp/golingua-light-f2-runtime.json`:

```json
{"schema_version":1,"dataset_store":"/tmp/golingua-light-f2-store","dataset_id":"golingua-authored","language_package_dirs":[]}
```

Snapshot: `180978604ce944fa2dcfdfd7f07acb42d2d0392a8e80417b9e183f1365b0aa06`, status ok, zastosowany i trwały. Wszystkie polecenia CLI są w JSONL. Przykłady:

```sh
/tmp/golingua-light-f2 forms --language pl --lexeme lexeme:pl.rzeka.n --feature number=plural --feature case=locative --feature gender=feminine --json
/tmp/golingua-light-f2 forms --language pl --surface rzekach --json
/tmp/golingua-light-f2 forms --language en --surface looked --json
/tmp/golingua-light-f2 translate --from en --to pl --text 'The teacher looks at the book.' --runtime-config /tmp/golingua-light-f2-runtime.json --json
```

Eksploracyjne `forms --help`, generacja bez cech lub `verb_form=base` i nieistniejące `participle_type` dały jawne błędy argumentów/profilu. Poprawiono polecenia według istniejącego kontraktu komórek; te próby nie są dowodem błędu odmiany ani zaliczonymi kontrolami. Lokalne surowe logi: `/tmp/light-f2-{selected,full,vet}.log`, obserwacje przed/po w `/tmp/light-f2-{before,after}-forms.json` oraz `/tmp/light-f2-{before-cli,cli}.json`.

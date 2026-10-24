# Odbiór LIGHT-F4

Koordynator wykonał odbiór 2026-10-25 (Europe/Warsaw). Base `57e4882e79eb6596caed5126108fa272177e6982`; kod po zmianie `e2173740f1bf930f5f37302f98d95b5972599b48`. Ocena lokalnego zakresu [#328](https://github.com/clickbytes/golingua/issues/328): PASS. Akceptacja językowa właściciela: NOT_EVALUATED. Późniejszy commit dodaje wyłącznie dowody; końcowy merge i rzeczywiste CI main są osobną bramką zapisaną w issue/PR.

## Zakres i pomiar

Usunięto pełny Lexicon.Document i dwie mapy katalogu z każdego SelectTargets. Lookup sensu używa binary search, a zwracane rekordy nadal mają defensywne kopie. Zgodność niezmiennego CORE/pack jest obliczana raz przez Engine.New; Translate zachowuje tę samą kolejność kontroli i diagnostyki mismatch/missing. Nie zmieniono danych językowych, Meaning, budżetów, walidacji kandydatów ani CI.

Go 1.27.0, Linux amd64, Intel Core i7-11800H, 16 logicznych CPU, GOMAXPROCS=16. Po trzy oddzielne procesy dla base i after, identyczny [program pomiaru](328-measure.go.txt), bez współbieżnych zadań pomiarowych. Każdy proces inicjalizuje produkcyjny runtime przez matrix.NewRuntime, następnie 10 razy rozgrzewa i 100 razy mierzy każde z czterech wejść. JSON jest serializowany po mierzonych pętlach; liczniki alokacji obejmują także niewielką obsługę pomiaru. Poniżej mediana trzech median/p95 oraz trzech średnich alokacji. MB oznacza 1 000 000 bajtów; to alokacje, nie zajęta jednocześnie pamięć.

| Wejście | Mediana ms przed → po | p95 ms przed → po | MB/żądanie przed → po | Alokacje/żądanie przed → po |
| --- | --- | --- | --- | --- |
| The teacher sees the dog. | 2,149 → 1,010 | 4,622 → 1,952 | 3,930 → 0,800 | 16606 → 10490 |
| The teacher reads the message. | 2,094 → 0,866 | 4,308 → 1,440 | 3,850 → 0,719 | 14654 → 8538 |
| Nauczyciel nie widzi psa. | 5,454 → 3,946 | 7,710 → 6,136 | 6,095 → 2,960 | 47068 → 40948 |
| Czarny kot widzi psa. | 6,921 → 5,175 | 9,436 → 7,552 | 7,046 → 3,918 | 60890 → 54880 |

We wszystkich próbach statusy i liczności kandydatów były takie same: dwa wejścia EN incomplete/1 i dwa PL ambiguous/4. Wynik wykazuje około 3,13 MB mniej alokacji na żądanie w tym zakresie. Nie jest procentową bramką jakości ani gwarancją czasu na innym sprzęcie.

Inicjalizacja danych Go: przed 1056–1112 ms, mediana 1097 ms; po 1074–1361 ms, mediana 1093 ms. Alokacje inicjalizacji pozostają około 763 MB, żywa sterta po GC około 18,98 MB. Nie wykazano poprawy startu lub pamięci rezydentnej. Trzy świeże procesy rzeczywistego CLI, obejmujące wczytanie snapshotu i pojedynczy przekład, miały przed 878–907 ms, po 918–966 ms. Cache systemu operacyjnego i obciążenie stanowiska nie były kontrolowane; pomiar CLI nie wykazuje przyspieszenia. Zmiana dotyczy powtarzanych żądań załadowanego silnika. Nie mierzono energii ani nie porównywano z modelem neuronowym.

[Surowe pomiary](328-measurements.json) zawierają wszystkie sześć przebiegów, metadane, świeże procesy CLI oraz parametry profili. Nie wybrano tylko najszybszego przebiegu.

## Profile

Oddzielne profile wykonano po inicjalizacji i rozgrzaniu: 1000 wywołań, po 250 każdego wejścia. Profil alokacji końcowy pomniejszono przez `-base` o profil przed pętlą, aby wyłączyć inicjalizację. Surowe profile są w [328-profiles](328-profiles/manifest.json), z rozmiarami i SHA256; nie są zależnością produktu ani CI. Raporty: [alokacje przed](328-profiles/before-alloc-top.txt), [alokacje po](328-profiles/after-alloc-top.txt), [CPU przed](328-profiles/before-cpu-top.txt), [CPU po](328-profiles/after-cpu-top.txt).

W profilu alokacji przed zmianą SelectTargets miał flat 1404,67 MiB; cloneDocument 911,07 MiB, reflect.deepValueEqual 360,66 MiB, Concepts 278,20 MiB. Po zmianie SelectTargets miał flat 133,37 MiB, a trzy usunięte operacje nie miały próbek w różnicowym profilu żądań (sprawdzono także pełny raport bez odcinania małych pozycji). Całkowite próbkowane alokacje profilu spadły z 5021,76 do 2031,18 MiB. CPU dla 1000 wywołań: 8,01 s próbek przed i 4,45 s po, obejmuje pracę równoległego GC; nie utożsamiono tego z czasem ściennym ani energią. Koszty walidacji Meaning, kanonizacji IDs i klonowania konstrukcji pozostają poza tym zamkniętym zakresem; nie uruchomiono dodatkowego backlogu automatycznie.

## Zachowanie i kontrole

[40 obserwacji CLI](328-observations.jsonl) porównano z #327: identyczne exit, status, wszystkie ID/kandydatury, grafy, teksty, założenia, straty i diagnostyki. Jedynie obszerne Evidence parsera jest pominięte w obu zapisach. Znane odmowy dla nieobsługiwanej polskiej formy przymiotnika i dodatkowego sensu fish nadal są odmowami; ich niezaliczony zakres nie został ukryty.

[Standalone próba niezmienności](328-check.go.txt) uruchomiona na końcowym SHA dała [12 poprawnych obserwacji](328-check-results.json): znany/nil/missing lookup, obronne Usage/ValencyIDs/Properties/CORE views, identyczny pełny wynik Translate po mutacji kopii, ten sam i oddzielny równoważny CORE, mismatch i brakujące zasoby. `engine.New` nadal dopuszcza konfigurację mismatch; request ma invalid/RUNTIME_SNAPSHOT_MISMATCH. Brak CORE lub pack nadal ma unknown/RUNTIME_SNAPSHOT_UNAVAILABLE. To mała ręcznie uruchamiana próba, bez pakietu testowego, runnera lub bramki w CI.

| Kontrola końcowego SHA | Wynik |
| --- | --- |
| gofmt wszystkich śledzonych Go, diff check i lokalne linki | PASS |
| `go vet ./...` | PASS |
| `go build -o /tmp/golingua-light-f4 ./cmd/golingua` | PASS |
| `--help`, `--version`, catalog summary/details | PASS |
| Generacja pies plural genitive masculine_animal i analiza psów | PASS |
| `go test ./tests/... -run 'Composition\|Generalization\|MeaningAttributes\|Identification\|Negation\|Seeing\|Inventory' -count=1 -v` | PASS |
| `go test ./tests/... -count=1 -v` | FAIL, wyłącznie 36 SPEC_PENDING, 0 INFRA, 24 oryginalne ID |

Testy używały `GOLINGUA_MATRIX_SHA=e2173740f1bf930f5f37302f98d95b5972599b48`, raw `/tmp/light-f4-{selected,full,vet,build}.log`. Niezaliczone są analiza/realizacja/integracja 12 rekordów LQ/PO/Q (001–004, oryginalne kierunki). Żaden przypadek nie jest pominięty. Techniczne CI pozostało bez zmian. TUI NOT RUN dla #328: zmiana nie dotyczy interfejsu, kontrakt wymaga pomiaru/CLI; rzeczywiste przepływy TUI odebrano w #327. Nie ma obowiązkowej obserwacji tego zakresu pozostawionej NOT RUN.

## Odtworzenie pomiaru

Z końcowego checkoutu zawierającego tekst programu wybierz base lub after SHA powyżej. Program nie jest uruchamiany w aktywnym checkoutcie:

```sh
measure_sha=57e4882e79eb6596caed5126108fa272177e6982
measure_copy=$(mktemp -d /tmp/golingua-f4-reproduce-XXXXXX)
git archive "$measure_sha" | tar -x -C "$measure_copy"
mkdir -p "$measure_copy/cmd/light-f4-measure"
cp docs/work/light-lm/328-measure.go.txt "$measure_copy/cmd/light-f4-measure/main.go"
(cd "$measure_copy" && go build -o /tmp/golingua-f4-measure ./cmd/light-f4-measure)
GOMAXPROCS=16 /tmp/golingua-f4-measure -sha "$measure_sha" -out /tmp/golingua-f4-repeat-1
GOMAXPROCS=16 /tmp/golingua-f4-measure -sha "$measure_sha" -out /tmp/golingua-f4-repeat-2
GOMAXPROCS=16 /tmp/golingua-f4-measure -sha "$measure_sha" -out /tmp/golingua-f4-repeat-3
GOMAXPROCS=16 /tmp/golingua-f4-measure -sha "$measure_sha" -out /tmp/golingua-f4-repeat-profile -profile
go tool pprof -top -alloc_space -base /tmp/golingua-f4-repeat-profile/allocs-start.pprof /tmp/golingua-f4-repeat-profile/allocs-end.pprof
go tool pprof -top -cum /tmp/golingua-f4-repeat-profile/cpu.pprof
```

Powtórz z after SHA i innymi katalogami wyników; nie uruchamiaj pomiarów równolegle. Do odtworzenia próby niezmienności użyj after SHA, skopiuj `328-check.go.txt` do `cmd/light-f4-check/main.go` w tej kopii, zbuduj i uruchom z `-sha "$measure_sha" -out /tmp/golingua-f4-check.json`. Program nie jest zgodny z base przed wprowadzeniem BySenseID.

Rzeczywisty CLI:

```sh
go build -o /tmp/golingua-light-f4 ./cmd/golingua
/tmp/golingua-light-f4 snapshot --apply --store /tmp/golingua-light-f4-store --json
```

Config `/tmp/golingua-light-f4-runtime.json`:

```json
{"schema_version":1,"dataset_store":"/tmp/golingua-light-f4-store","dataset_id":"golingua-authored","language_package_dirs":[]}
```

Snapshot `5fe8d84d0a4ffbae85ad34d1864015690d32ab2c49112fad26a07450c76d9d77`, aktywny i trwały, identyczny z #327. Polecenia wszystkich 40 obserwacji są w JSONL. W celu pomiaru świeżego procesu uruchom wybrane polecenie oddzielnie ze stałym GOMAXPROCS=16; w tym pomiarze nie czyściliśmy cache OS. Koordynator przejrzał diff i wyniki; zlecenia oraz korekty są w [rejestrze pracy](328-jobs.md).

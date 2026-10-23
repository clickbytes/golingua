# Bieżące CLI

CLI udostępnia wersję, pomoc, deklarowane capabilities, operacje na Meaning JSON, typowane `question`, komendę `translate` oraz odmianę `forms`. Samo istnienie komendy `translate` nie oznacza, że środowisko ma zainstalowane języki lub może wykonać translację.

## Działające polecenia

Uruchom z katalogu repozytorium:

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

`--version` zwraca wersję technicznego podglądu v0. `--help` pokazuje dostępne komendy. `capabilities --json` opisuje gotowość bez wybranego runtime'u, dlatego obecnie zwraca `status: "unsupported"`, puste tablice `languages` i `features` oraz `ready: false`; sama komenda kończy się kodem 0. `forms` i `catalog` mają własne komendy opisane poniżej.

## Translacja

Użyj jawnych flag, aktywnego snapshotu i wymaganej odpowiedzi JSON:

```sh
go run ./cmd/golingua snapshot --apply --store ./store --json
go run ./cmd/golingua translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config ./runtime.json --json
```

`runtime.json` wskazuje aktywny snapshot. W skompilowanym CLI nie trzeba podawać `language_package_dirs`: binarka kompiluje ograniczone pakiety PL/EN dla dokładnych szlaków `seeing`, `hearing`, `looking` oraz `receiving`. Trzy pierwsze obsługują jawne formy teraźniejsze 3. osoby liczby pojedynczej; `receiving` obsługuje angielskie `got` oraz polskie 3. osobowe formy czasu przeszłego liczby pojedynczej. `looking` zachowuje `look at` / `patrzeć na` i polski biernik. Bez `--runtime-config` odpowiedź pozostaje `unsupported`, ponieważ proces nie ma wybranego snapshotu.

Opcjonalny `--preview-id <id>` wybiera podgląd wskazanego kandydata, jeśli silnik zwrócił kilka alternatyw. Wybór zachowuje ogólny status `ambiguous` lub `incomplete`; nie zmienia Meaning. Po dojściu do etapu wyboru kandydata identyfikator, który nie wskazuje kompletnego zachowanego kandydata, zwraca `invalid`. Jeśli runtime jest niedostępny, żądanie kończy się `unsupported`, bo nie ma alternatyw do sprawdzenia. Flagi mogą wystąpić w dowolnej kolejności. `--from`, `--to`, `--text` i `--json` są wymagane; wartości tekstowe nie mogą być puste. Nieznane, powtórzone, błędne lub niekompletne argumenty zwracają JSON ze statusem `invalid` i kodem wyjścia 2. Nie podawaj argumentów pozycyjnych.

Wynik translacji ma pola `status`, `from`, `to`, opcjonalne `text`, opcjonalne `preview_id`, `candidates` oraz `diagnostics`. Niepoprawne żądanie zwraca te same pola z `status: "invalid"`; błędna składnia samych flag jest raportowana jako JSON `status`/`diagnostics`, bo nie powstało jeszcze żądanie do silnika. Obecne ograniczone szlaki obejmują zdania oznajmujące `seeing`, `hearing` i `looking` w teraźniejszej 3. osobie liczby pojedynczej oraz `receiving` w czasie przeszłym, 3. osobie liczby pojedynczej. Angielska analiza przyjmuje jawne `the`, `a` oraz brak rodzajnika przy rzeczownikach niepoliczalnych z bieżącego inwentarza; polska realizacja używa jawnych form przypadków tylko dla wybranych rzeczowników. Brak odpowiedniej formy lub konstrukcji pozostaje `unsupported` albo `unknown`; kilka dopuszczalnych interpretacji pozostaje `ambiguous`. Tłumaczenie PL→EN może zwrócić tekst `incomplete`, ponieważ `MeaningGraph` nie koduje określoności potrzebnej do wyboru angielskiego rodzajnika. Status `ok` oznacza wykonanie skonfigurowanego pipeline'u, nie ocenę językową.

Tekst wejściowy ma limit 64 KiB, a każdy tekst wyjściowy również 64 KiB; łączny tekst alternatyw ma limit 64 KiB. Silnik ogranicza rozmiar całego wyniku do 4 MiB oraz liczbę kandydatów do 256. Identyfikatory języków mają maksymalnie 35 znaków, a identyfikator podglądu maksymalnie 256 bajtów. Przekroczenie limitu wejścia lub retencji alternatyw zwraca `resource_limit`, a nie obcięty tekst. Niepoprawny albo zbyt duży serializowany wynik zwraca `invalid`.

## Odmiana i analiza form

`forms` ładuje obecny ograniczony seed morfologiczny PL albo EN. Generacja wymaga identyfikatora leksemu i cech językowych; analiza przyjmuje formę powierzchniową:

```sh
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.kot.n --feature number=singular --feature case=nominative --feature gender=masculine_animal --json
go run ./cmd/golingua forms --language en --surface dogs --json
```

Wynik zawiera status oraz `forms` lub `readings`. Brak leksemu albo odczytu daje `unknown`. To nie jest translacja ani ocena jakości językowej. Ocena na słowach właściciela pozostaje `NOT RUN`.

W G07-03 `forms` obejmuje także 15 jawnie zadeklarowanych rzeczowników PL po 14 standardowych komórek case/number. Są to pełne listy głównych form zapisane jako principal parts, nie automatyczne reguły dla całego słownika. Przykład: `forms --language pl --lexeme lexeme:pl.pies.n --feature number=plural --feature case=genitive --feature gender=masculine_animal --json`. Test techniczny generatora/analizatora w CI nie jest akceptacją językową.

## Inwentarz katalogu

`catalog --summary --json` kompiluje obecne dane Go CORE, PL i EN i pokazuje
oddzielne liczby konceptów, leksemów, sensów oraz statusy bindingów.
`catalog --details --json` dodaje listę stabilnych ConceptID ze statusami PL/EN. `pending`
nie upoważnia do tłumaczenia. Liczba 1000 jest celem redakcyjnym, a raport nie
mierzy poprawności językowej.

`catalog --summary` raportuje liczbę pending construction candidates i
approved construction bindings oddzielnie dla PL i EN. Tylko approved binding
licencjonuje dokładną trasę sense/frame/family/variant; sam nie wystarcza do
tłumaczenia zdania bez zatwierdzonych powiązań leksykalnych i runtime.

## Snapshot katalogu Go

`snapshot` buduje walidowane pliki CORE i wspólnego leksykonu PL/EN z bieżących
deklaracji Go. Podgląd podaje skrót i liczności; aktywacja zapisuje te same dane
do wskazanego lokalnego magazynu i ustawia jego aktywny wskaźnik:

```sh
go run ./cmd/golingua snapshot --preview --store <katalog-magazynu> --json
go run ./cmd/golingua snapshot --apply --store <katalog-magazynu> --json
```

Obie komendy wymagają jawnej ścieżki i mogą utworzyć katalog magazynu. Snapshot
zachowuje `pending` i pochodzenie każdego bindingu. Nie zawiera walencji ani
konstrukcji; skompilowana binarka dostarcza je osobno i aktywuje tylko dokładne
trasy dozwolone przez dane snapshotu. Dla bieżącego CLI plik runtime może
wyglądać tak:

```json
{
  "schema_version": 1,
  "dataset_store": "./store",
  "dataset_id": "golingua-authored"
}
```

Do użycia CORE przez `question` skonfiguruj `dataset_store` na ten katalog i
`dataset_id` na `golingua-authored`.

## Pytania na faktach z żądania

Komenda `question` wykonuje typowane zapytanie na `MeaningGraph` i `FactSet` podanych w jednym żądaniu. Nie parsuje polskiego ani angielskiego tekstu i nie realizuje odpowiedzi powierzchniowej.

```sh
go run ./cmd/golingua question --runtime-config <runtime.json> --json < question-request.json
```

Runtime config wybiera aktywny CORE/dataset snapshot. Konfiguracja `question` oraz bieżące skompilowane CLI mogą pominąć `language_package_dirs`; aplikacje używające zewnętrznych katalogów pakietów nadal mogą je podać. JSON ze standardowego wejścia ma limit 4 MiB, schemat zamknięty i wersję `schema_version: 1`:

- `question`: `reason.QuestionInput` — typowany `meaning`, request-scoped `context`, `kind` (`yes_no` albo `wh`), jawne `mapping` relacji i ról, opcjonalne `gap_role` i `at`.
- `facts`: tablica `facts.Fact` dla tego samego request context; każdy fakt wymaga jawnego źródła i polaryzacji.
- `rules`: opcjonalne reguły `symmetry`/`transitivity`, które CORE jawnie deklaruje. Są walidowane względem wybranego snapshotu.

Wynik JSON zawiera `status`, wybrany `goal`, zaplanowane `question`, `answer` z prawdą/kandydatami i proof IDs, oraz tylko te fakty/źródła, które stanowią dowód. Brak dowodu daje `unknown`; nie tworzy faktu negatywnego. Niepoprawny snapshot, Meaning, FactSet lub mapowanie daje typowany status błędu. `question` nie jest dowodem jakości językowej; ręczny odbiór wymaga wejść dostarczonych przez właściciela.

Wyjście sukcesu i błędów domenowych jest pojedynczym obiektem JSON na stdout. Exit code odpowiada statusowi: `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; `error` i błędy infrastruktury używają kodu 1. Komenda `capabilities --json` sama kończy się 0 i opisuje gotowość przez pole `ready`.

Nieznana lub niepoprawna komenda zwraca `invalid` i exit code 2. Ogólne błędy procesu/serializacji mają osobne kody; `unsupported`, `invalid` i awaria programu nie są jedną kategorią.

## Przyszłe funkcje

Pozostałe sensy, konstrukcje, czasy i formy wymagają własnych etapów danych i implementacji. Aktualne obserwacje dla 25 istniejących przykładów `looking` w obu kierunkach zapisano w [`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../data/sentences/observed-2026-10-23-g05-14-looking.tsv), a dla `receiving` w [`data/sentences/observed-2026-10-23-g05-13-receiving.tsv`](../data/sentences/observed-2026-10-23-g05-13-receiving.tsv). To zapis działania CLI, nie gold ani automatyczna ocena jakości. Właścicielski odbiór językowy pozostaje `NOT RUN`; delegowana ocena agenta jest odrębna. Procedura i statusy są opisane w [manual review](manual-review.md).

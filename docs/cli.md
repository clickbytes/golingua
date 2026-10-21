# Bieżące CLI

CLI udostępnia wersję, pomoc, deklarowane capabilities, operacje na Meaning JSON i komendę `translate`. Samo istnienie komendy `translate` nie oznacza, że środowisko ma zainstalowane języki lub może wykonać translację.

## Działające polecenia

Uruchom z katalogu repozytorium:

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

`--version` zwraca wersję bootstrapu. `--help` pokazuje dostępne komendy. `capabilities --json` obecnie zwraca `status: "unsupported"`, puste tablice `languages` i `features`, oraz `ready: false`; sama komenda kończy się kodem 0.

## Translacja

Użyj jawnych flag oraz wymaganej odpowiedzi JSON:

```sh
go run ./cmd/golingua translate --from pl --to en --text 'tekst właściciela' --json
```

Opcjonalny `--preview-id <id>` wybiera podgląd wskazanego kandydata, jeśli silnik zwrócił kilka alternatyw. Wybór zachowuje ogólny status `ambiguous` lub `incomplete`; nie zmienia Meaning. Po dojściu do etapu wyboru kandydata identyfikator, który nie wskazuje kompletnego zachowanego kandydata, zwraca `invalid`. Jeśli runtime jest niedostępny, żądanie kończy się `unsupported`, bo nie ma alternatyw do sprawdzenia. Flagi mogą wystąpić w dowolnej kolejności. `--from`, `--to`, `--text` i `--json` są wymagane; wartości tekstowe nie mogą być puste. Nieznane, powtórzone, błędne lub niekompletne argumenty zwracają JSON ze statusem `invalid` i kodem wyjścia 2. Nie podawaj argumentów pozycyjnych.

Wynik translacji ma pola `status`, `from`, `to`, opcjonalne `text`, opcjonalne `preview_id`, `candidates` oraz `diagnostics`. Niepoprawne żądanie zwraca te same pola z `status: "invalid"`; błędna składnia samych flag jest raportowana jako JSON `status`/`diagnostics`, bo nie powstało jeszcze żądanie do silnika. Przy pustej domyślnej konfiguracji CLI nie ładuje providerów językowych. Dlatego obecnie prawidłowo sformowane żądanie zwraca jawny wynik `unsupported` z diagnostyką; nie jest to próba tłumaczenia ani wynik PL↔EN do oceny. CLI nie obiecuje poprawnego tłumaczenia. Dostępne są statusy `ok`, `unsupported`, `unknown`, `ambiguous`, `incomplete`, `invalid`, `resource_limit` i `error`; `ok` będzie zwracane wyłącznie po rzeczywistym wyniku silnika.

Tekst wejściowy ma limit 64 KiB, a każdy tekst wyjściowy również 64 KiB; łączny tekst alternatyw ma limit 64 KiB. Silnik ogranicza rozmiar całego wyniku do 4 MiB oraz liczbę kandydatów do 256. Identyfikatory języków mają maksymalnie 35 znaków, a identyfikator podglądu maksymalnie 256 bajtów. Przekroczenie limitu wejścia lub retencji alternatyw zwraca `resource_limit`, a nie obcięty tekst. Niepoprawny albo zbyt duży serializowany wynik zwraca `invalid`. Nie ma komendy `forms`.

Wyjście sukcesu i błędów domenowych jest pojedynczym obiektem JSON na stdout. Exit code odpowiada statusowi: `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; `error` i błędy infrastruktury używają kodu 1. Komenda `capabilities --json` sama kończy się 0 i opisuje gotowość przez pole `ready`.

Nieznana lub niepoprawna komenda zwraca `invalid` i exit code 2. Ogólne błędy procesu/serializacji mają osobne kody; `unsupported`, `invalid` i awaria programu nie są jedną kategorią.

## Przyszłe funkcje

Obsługiwane kierunki i ręczna ocena translacji zależą od skonfigurowanych providerów oraz danych. Procedura oceny wyników dostarczonych przez właściciela jest opisana w [manual review](manual-review.md).

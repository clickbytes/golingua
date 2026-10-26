# DSE-06-02: referenty i kontekst żądania

Ten zapis dokumentuje lokalne kontrole końcowego źródła #392. Obejmuje typowane
referenty, jawne rozstrzyganie wzmianek, warunki niezależne od faktów oraz
zachowanie payloadu przez dostępne przepływy CLI. Jest to obserwacja techniczna
i ocena koordynatora; akceptacja językowa właściciela pozostaje
`NOT_EVALUATED`.

## Zamrożone źródło i kontrole techniczne

- Gałąź: `dse/06-02-references-context`; baza issue: `80926ec5bdec30010fba9f5e265993881da9f06a`.
- Końcowy binarny plik: `/tmp/golingua-dse/bin/golingua-392-final`.
- SHA-256 binarnego pliku: `2d7aa77db048d8aa992584a1ff3d1880774ba5bdf038345e77781bf4e592d927`.
- [`392-source-hashes.json`](392-source-hashes.json) zawiera skróty 34 plików Go i schematów użytych do zbudowania tego pliku.
- [`392-technical-observation.json`](392-technical-observation.json) rejestruje `gofmt -l` (pusty wynik), `go vet ./...`, build do wskazanego pliku, `--help` i `--version`; wszystkie komendy zakończyły się kodem 0.
- [`392-schema-observation.json`](392-schema-observation.json) zapisuje walidację 19 wejść i artefaktów przy użyciu jsonschema 4.26.0; metaschematy i wszystkie sprawdzone dokumenty miały zero błędów.

## Rzeczywiste wywołania CLI

Trzy literalne dzienniki zachowują argv, kody wyjścia, ścieżki i SHA-256 stdout:

- [`392-cli-observations.jsonl`](392-cli-observations.jsonl): 40 komend, kody 0: 26, 2: 14.
- [`392-extra-cli-observations.jsonl`](392-extra-cli-observations.jsonl): 6 komend, kody 0: 5, 2: 1.
- [`392-final-controls.jsonl`](392-final-controls.jsonl): 3 komendy, kody 0: 1, 2: 1, 4: 1.

Łącznie to 49 wywołań: kod 0: 32, kod 2: 16, kod 4: 1. Wyjścia są skopiowane
pod [`outputs/`](outputs/), wejścia pod [`inputs/`](inputs/). Dzienniki zachowują
oryginalne ścieżki `/tmp/golingua-dse/`; kopie w tym katalogu można powiązać po
nazwie i grupie obserwacji. Trzy generowane wejścia mają teraz odrębne ścieżki
`*-input.json`; są skopiowane osobno obok odpowiadających im stdout.

Dodatkowy, oddzielny test zgodności nil-context jest zapisany w
[`392-final-nil-observation.json`](392-final-nil-observation.json), a jego
stdout w [`outputs/final-controls/`](outputs/final-controls/). Bieżący
`golingua-392-final` i zamrożony binarny #391 dały identyczne
`resource_fingerprint`, `meaning_fingerprint` i `cache_key` przy eksporcie
starego żądania bez `request_context`. Ten dodatkowy test nie wchodzi do
powyższych 49 rekordów.

## Zaobserwowane zachowanie

[`392-semantic-observation.json`](392-semantic-observation.json) zawiera
sprawdzone przez koordynatora porównania named records przed i po normalizacji,
eksporcie i odtworzeniu:

- Rozstrzygnięty zestaw wskazuje `m1 → A`, `m2 → A`; nierozstrzygnięty zachowuje
  `mU` jako `unresolved`; sprzeczne dowody dla `m1` pokazują `ambiguous` z
  kandydatami A i B. Powtórne wskazanie A nie scala ani nie usuwa B.
- Referenty A i B mają zgodny typ i `concept_id`, lecz ich ID pozostają różne.
  Jawny cel opisu wiąże tylko port `y` z Ref A. Wynik
  `λx.r(x, A)` zachowuje otwarty port `x`; obecność B nie jest przesłanką do
  wyboru ani domknięcia tego portu.
- Kinds warunków (`assumption`, `presupposition`, `loss`) są zachowane. Osobna
  próba ewaluacji z warunkiem presupozycji `p(A)` nadal zwraca `knowledge:
  "unknown"`; kontekst nie tworzy rekordu faktu.
- Warunek z zerową arnością zachowuje `args: []` przez sprawdzenie,
  normalizację, eksport i odtworzenie. Opis z tym warunkiem wymaga
  `metadata-zero-arity.json`, które deklaruje `t0`; zwykłe `metadata.json` go
  nie zawiera.
- Porównanie i weryfikacja równoważności z tym samym kontekstem zakończyły się
  powodzeniem. Payload zmieniony przy tym samym ID został odrzucony. Odtworzenie
  artefaktu wymaga osobnego oczekiwanego kontekstu, a jego brak lub obcy payload
  zostały odrzucone. Przeszukanie celu zachowało opis otwarty; weryfikacja
  zwróconego certyfikatu celu przeszła.
- Limit referentów zakończył się `resource_limit` (kod 4). Błędny typ, dangling
  referent/wzmianka, niezgodny context ID, niedozwolona morfologia, brak
  wymaganej kotwicy i zmieniony pin zwróciły odmowy zamiast zgadywania.

Wyniki potwierdzają strukturę, typy i jawne powiązania w wybranym profilu.
Nie dowodzą istnienia, unikalności, prawdy świata, pełności wyszukiwania ani
akceptacji językowej właściciela.

## Odtworzenie

Neutralne wejścia użyte przez koordynatora znajdują się w
[`inputs/independent/`](inputs/independent/); generowane żądania i dodatkowe
kontrole są rozdzielone w [`inputs/cli/`](inputs/cli/),
[`inputs/extra/`](inputs/extra/) i [`inputs/final-controls/`](inputs/final-controls/).
Wyjścia są pogrupowane według tych samych trzech dzienników. Przykłady produktu
są w [katalogu reference](../../../../../data/semantic/examples/reference/),
a ich komendy opisano w [CLI](../../../../cli.md) i
[kontrakcie kontekstu](../../../../architecture/semantic-algebra/reference.md).

Nie uruchamiano `go test` w tym zadaniu. Nie dodano runnera, goldów ani
automatycznej oceny. Właściciel nie zaakceptował jeszcze językowo wyniku
(`NOT_EVALUATED`); PR oraz rzeczywisty CI PR/main pozostają oczekujące.

## Literalna kontrola dokumentacji

Koordynator uruchomił dosłownie cztery bloki shell z sekcji CLI: łącznie 10 wywołań GoLingua, z ekstrakcją kandydata/certyfikatu przez jq w jednym bloku. Wszystkie bloki zakończyły się kodem 0. [Osobny zapis](392-docflow-observation.json) i stdout w `outputs/doc-flow/` uzupełniają 49 wcześniejszych obserwacji; nie są do nich doliczone.

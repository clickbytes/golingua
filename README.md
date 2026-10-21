# golingua

**Nowy, niezależny prototyp silnika językowego w Go.** Koncepty, gramatyka, walencja, procedury i wyjątki są projektowane od zera. Polski i angielski pozostają osobnymi pakietami językowymi.

**Status technicznego podglądu v0:** CLI nie tłumaczy PL↔EN ani nie analizuje/realizuje swobodnych zdań. `forms` generuje i analizuje wybrane formy z małych seedów PL/EN, `catalog` raportuje stan danych Go, a `question` wykonuje typowane zapytania na MeaningGraph i faktach podanych jawnie w JSON. Docelowo: około 1000 opisanych konceptów CORE; obecnie 49. Brak pamięci i uczenia w v0.

## Jak rozwijamy

**Jedno zadanie, jeden agent, jeden PR; EPIC po EPIC-u.** Na tym etapie **nie przenosimy starej implementacji JavaScript i nie porównujemy zachowania Go do wyników starego silnika**. Nie prowadzimy masowej macierzy testów ani automatycznego testowania tłumaczeń. Właściciel projektu dostarcza zdania/słowa PL i EN; uruchamiamy Go CLI i oceniamy wyjście **ręcznie**. Zapisujemy tylko zauważone błędy i poprawki, bez obowiązkowych „goldów”, coverage %, test runnera czy progu liczby przykładów.

CI kontroluje jedynie, czy projekt jest sformatowany, kompiluje się oraz czy CLI startuje (gofmt, go vet, go build, --help/--version). **Zielone CI nie znaczy, że tłumaczenie działa poprawnie.**

Istniejące źródło [clickbytes/flexLingua](https://github.com/clickbytes/flexLingua) może być konsultowane przy projektowaniu i ewentualnym imporcie danych, lecz nie jest wzorcem API, kodu ani wyników. Nie ma obowiązku importu jego słowników lub wcześniejszych testów.

## Dokumentacja

- [Architektura golingua](docs/architecture/README.md) — kontrakty i kolejność projektowania.
- [Generator i finite relation table](docs/architecture/morph-backends.md) — API ograniczonego generowania i analizy morfologicznej.
- [English morphology seed](docs/languages/en.md) — obecne klasy, API i ograniczenia EN.
- [Polish morphology seed](docs/languages/pl.md) — obecne klasy, API i ograniczenia PL.
- [Protokół ręcznego sprawdzania](docs/manual-review.md) — użytkownik podaje wejścia, agent pokazuje faktyczny output.
- [Planowanie treści odpowiedzi](docs/architecture/content-planning.md) — wybór aktu i treści odpowiedzi na podstawie typowanego wyniku zapytania.
- [Realizacja odpowiedzi dialogowej](docs/architecture/dialogue-realization.md) — bezpieczny most do wspólnego pipeline’u konstrukcji, morfologii i linearyzacji.
- [CLI](docs/cli.md) — rzeczywiste komendy bootstrapu i jawne granice obecnej obsługi.
- [Przykłady ręczne](examples/manual/README.md) — na razie bez domyślnych zdań.
- [Status katalogu](docs/data/catalog-status.md) — bieżące liczby CORE/PL/EN i status zatwierdzenia bindingów.
- [Granice bezpieczeństwa](docs/security.md) — walidacja wejścia i importów.
- [Podgląd v0](docs/release/v0.md) — możliwości i ograniczenia tego wydania.

Typowane węzły: ConceptDefinition, MorphLexeme, LexicalSense, SemanticBinding, ValencyFrame, ConstructionFamily/Variant, ProcedurePlan, Paradigm, MeaningGraph. Wspólna reprezentacja znaczenia niezależna od PL/EN. **Nie implementujemy wymaganego API `ROCK.VERB.MORPH.PAST`**; wystarczają typowane identyfikatory i zwykłe wywołania Go.

## Uruchomienie podglądu v0

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
go run ./cmd/golingua catalog --summary --json
go run ./cmd/golingua forms --language en --surface dogs --json
```

`capabilities --json` raportuje `ready: false` dla translacji. `translate` bez zatwierdzonych danych i providerów zwraca jawne `unsupported`. Dla odmiany własnego słowa użyj `forms --language pl|en --surface <forma> --json`; wymagane cechy generacji są opisane w [CLI](docs/cli.md). Nie ma TUI, serwera, bazy, FST, pamięci ani uczenia.

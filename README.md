# golingua

**Nowy, niezależny prototyp silnika językowego w Go.** Koncepty, gramatyka, walencja, procedury i wyjątki są projektowane od zera. Polski i angielski pozostają osobnymi pakietami językowymi.

**Status technicznego podglądu v0:** CLI nie tłumaczy PL↔EN ani nie analizuje/realizuje swobodnych zdań. `forms` generuje i analizuje ograniczony, jawnie opisany zestaw form PL/EN (w PL dodatkowo 15 rzeczowników z 14 deklarowanymi formami każdy), `catalog` raportuje stan danych Go, a `question` wykonuje typowane zapytania na MeaningGraph i faktach podanych jawnie w JSON. CORE ma już 1000 skompilowanych konceptów. Stan Go: PL 982 leksemów / 1007 sensów / 1006 kandydackich bindingów na 1000 konceptach (273 ramy walencji), EN 104 leksemy / 113 sensów / 104 kandydackie bindingi (103 koncepty, 5 ram). Ledger EN ma 104 kandydackie bindingi i 897 konceptów jeszcze nieopracowanych redakcyjnie. Nie ma zatwierdzonych powiązań ani realizacji dowolnych zdań. Brak pamięci i uczenia w v0.

## Jak rozwijamy

**Jedno aktywne zadanie i jeden agent; domyślnie jeden PR.** W #60 właściciel dopuścił sekwencyjne commity bezpośrednio na main. Zmiany w Go nie są zatwierdzonymi tłumaczeniami. Na tym etapie **nie przenosimy starej implementacji JavaScript i nie porównujemy zachowania Go do wyników starego silnika**. Nie prowadzimy masowej macierzy testów ani automatycznego testowania tłumaczeń. Właściciel projektu dostarcza zdania/słowa PL i EN; uruchamiamy Go CLI i oceniamy wyjście **ręcznie**. Zapisujemy tylko zauważone błędy i poprawki, bez obowiązkowych „goldów”, coverage %, test runnera czy progu liczby przykładów.

CI kontroluje gofmt, go vet, build, start CLI (--help/--version) oraz kompilację katalogu i obu indeksów walencji w smoke `catalog --summary/--details --json`. **Zielone CI nie znaczy, że tłumaczenie działa poprawnie.**

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

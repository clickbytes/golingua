# golingua

**Nowy, niezależny prototyp silnika językowego w Go.** Koncepty, gramatyka, walencja, procedury i wyjątki są projektowane od zera. Polski i angielski pozostają osobnymi pakietami językowymi.

**Status:** CLI pozostaje bootstrapem bez tłumaczenia PL↔EN, rozumowania ani generowania zdań. Biblioteka udostępnia już ograniczoną morfologię oraz mały seed EN przez Go API (bez CLI `forms`). Docelowo: około 1000 konceptów z leksyką w obu językach, bez pamięci i uczenia w v0.

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
- [CLI](docs/cli.md) — rzeczywiste komendy bootstrapu i jawne granice obecnej obsługi.
- [Przykłady ręczne](examples/manual/README.md) — na razie bez domyślnych zdań.
- [EPIC G00](https://github.com/clickbytes/golingua/issues/1) — aktywny etap; następne etapy czekają.

Typowane węzły: ConceptDefinition, MorphLexeme, LexicalSense, SemanticBinding, ValencyFrame, ConstructionFamily/Variant, ProcedurePlan, Paradigm, MeaningGraph. Wspólna reprezentacja znaczenia niezależna od PL/EN. **Nie implementujemy wymaganego API `ROCK.VERB.MORPH.PAST`**; wystarczają typowane identyfikatory i zwykłe wywołania Go.

## Uruchomienie aktualnego bootstrapu

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

Na dziś `translate` zwraca jawne `LANGUAGE_RUNTIME_UNAVAILABLE`. Nie ma TUI, serwera, bazy, FST, pamięci ani uczenia.

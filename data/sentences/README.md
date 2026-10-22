# 100 par zdań PL/EN powiązanych z CORE

Na polecenie właściciela z 2026-10-23 zapisano 100 **redakcyjnych propozycji** zdań w [100-pl-en.tsv](100-pl-en.tsv). Dla każdego wiersza [100-meanings.jsonl](100-meanings.jsonl) zawiera odpowiadający mu przykład request-scoped `MeaningGraph` z konceptem zdarzenia i dwoma osobnymi referentami. `S001` w obu plikach oznacza ten sam przykład. Zdania nie są nowymi `ConceptDefinition`, `LexicalSense` ani zatwierdzonymi tłumaczeniami.

Materiał obejmuje po 25 zdań dla `seeing`, `hearing`, `looking` i `receiving`. Każdy użyty koncept i sens zdarzenia ma obecnie osobny zatwierdzony binding PL/EN w katalogu, ale zdania, formy czasowników, wybór rodzajnika, czas/aspekt oraz konstrukcje nie zostały przez to zatwierdzone. Angielska trasa konstrukcyjna `look at` nie ma jeszcze approved bindingu. Powtarzane kombinacje uczestników służą sprawdzeniu zamiany ról i odmiany; nie stanowią miary jakości.

Wszystkie 100 struktur `MeaningGraph` sprawdzono przez `meaning.ValidateMeaning` względem skompilowanego CORE. Skompilowane indeksy leksykalne PL/EN potwierdziły `approved` i `equivalent` dla każdego wskazanego konceptu oraz dokładnych sensów zdarzeń. Ta kontrola dotyczy identyfikatorów i ról, a nie poprawności powierzchni zdań.

Status wszystkich wierszy to `draft_unreviewed`. Autorem powierzchni PL/EN jest Codex; licencja tych nowych propozycji nie została osobno ustalona. Ręczna ocena właściciela: **NOT RUN**. `MeaningGraph` przechowuje tutaj typ zdarzenia i role uczestników; nie koduje czasu, aspektu, liczby ani określoności z zapisanych zdań. Dlatego pliki nie są goldami ani automatyczną bramką expected output.

## Obserwacja CLI

W roboczej gałęzi #195 dodano jawne, ograniczone formy potrzebne dla grupy `seeing`:
`widzi` i `sees`, polskie paradygmaty `matka` i `ojciec` oraz angielskie
formy liczby pojedynczej uczestników. Komenda `forms` potwierdza analizę tych
form. Nie daje to jeszcze wyników `translate`: zatwierdzone szablony
`pl.transitive` i `en.transitive` mają sloty `subject`/`object` bez ról.
`construction.BindFrame` może je teraz skompilować dla dokładnej ramy
walencji, która jawnie zapisuje `ArgumentSlot.ID` i rolę. Providery zdań
i adaptery CLI nadal nie są zarejestrowane. Angielskie rodzajniki wymagają osobnej
analizy frazy i reprezentacji określoności; przykładowy MeaningGraph jej nie
przechowuje. Ręczna ocena jakości zdań pozostaje **NOT RUN**.

Lokalne uruchomienie skompilowanego `pl.NewSentenceSource` na tym snapshotcie
zwróciło `ok` oraz role `experiencer=teacher`, `stimulus=dog` dla
`Nauczyciel widzi psa.`. Dla `Pies widzi nauczyciela.` role zostały zamienione
zgodnie z przypadkami. `Nauczyciel słyszy psa.` pozostało `unsupported`, bo
brakuje jawnej formy `słyszy` w seedzie. Ta obserwacja dotyczy analizy PL,
nie przekładu w CLI ani ręcznej oceny jakości.

Na kodzie `d05212e` zbudowano `go build -o /tmp/golingua-g05 ./cmd/golingua` i dla każdego z 100 wierszy uruchomiono obie komendy:

```sh
/tmp/golingua-g05 translate --from pl --to en --text '<pl z wiersza>' --json
/tmp/golingua-g05 translate --from en --to pl --text '<en z wiersza>' --json
```

Dosłowne statusy i kody diagnostyczne są w [observed-2026-10-23-d05212e.tsv](observed-2026-10-23-d05212e.tsv). Wszystkie 200 wywołań zwróciły exit code 3, `unsupported` i `LANGUAGE_REGISTRY_UNAVAILABLE`; żadne nie zwróciło tłumaczenia. Jest to obserwacja stanu runtime, a nie ocena językowa. Do rzeczywistego sprawdzenia PL↔EN potrzebne są providerzy, `StageFactory`, reguły fraz i form slotów w #123 oraz ocena właściciela po ich uruchomieniu.

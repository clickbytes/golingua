# 100 par zdań PL/EN powiązanych z CORE

Na polecenie właściciela z 2026-10-23 zapisano 100 **redakcyjnych propozycji** zdań w [100-pl-en.tsv](100-pl-en.tsv). Dla każdego wiersza [100-meanings.jsonl](100-meanings.jsonl) zawiera odpowiadający mu przykład request-scoped `MeaningGraph` z konceptem zdarzenia i dwoma osobnymi referentami. `S001` w obu plikach oznacza ten sam przykład. Zdania nie są nowymi `ConceptDefinition`, `LexicalSense` ani zatwierdzonymi tłumaczeniami.

Materiał obejmuje po 25 zdań dla `seeing`, `hearing`, `looking` i `receiving`. Każdy użyty koncept i sens zdarzenia ma obecnie osobny zatwierdzony binding PL/EN w katalogu, ale zdania, formy czasowników, wybór rodzajnika, czas/aspekt oraz konstrukcje nie zostały przez to zatwierdzone. Angielska trasa konstrukcyjna `look at` nie ma jeszcze approved bindingu. Powtarzane kombinacje uczestników służą sprawdzeniu zamiany ról i odmiany; nie stanowią miary jakości.

Wszystkie 100 struktur `MeaningGraph` sprawdzono przez `meaning.ValidateMeaning` względem skompilowanego CORE. Skompilowane indeksy leksykalne PL/EN potwierdziły `approved` i `equivalent` dla każdego wskazanego konceptu oraz dokładnych sensów zdarzeń. Ta kontrola dotyczy identyfikatorów i ról, a nie poprawności powierzchni zdań.

Status wszystkich wierszy to `draft_unreviewed`. Autorem powierzchni PL/EN jest Codex; licencja tych nowych propozycji nie została osobno ustalona. Ręczna ocena właściciela: **NOT RUN**. `MeaningGraph` przechowuje tutaj typ zdarzenia i role uczestników; nie koduje czasu, aspektu, liczby ani określoności z zapisanych zdań. Dlatego pliki nie są goldami ani automatyczną bramką expected output.

## Obserwacja CLI

Gałąź #195 rejestruje skompilowane pakiety i etapy PL/EN dla ograniczonego,
zatwierdzonego szlaku `seeing`. Używa analizy składniowej, dokładnych bindingów,
ram walencji, planu konstrukcji, jawnych form morfologicznych i linearyzacji.
Angielski artykuł jest analizowany jako cecha `definite`; przy tłumaczeniu z PL
określoność nie występuje w `MeaningGraph`, więc wybór `the` jest jawnie
oznaczany jako założenie i wynik pozostaje `incomplete`.

Przykładowe lokalne wywołania na aktywnym snapshotcie:

```sh
golingua translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config runtime.json --json
golingua translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config runtime.json --json
```

Pierwsze zwróciło dosłowny tekst `The teacher sees the dog.` ze statusem
`incomplete` i diagnostykami `EXPLICIT_REALIZATION_ASSUMPTION` oraz
`TRANSLATION_NOT_OK`. Drugie zwróciło `Nauczyciel widzi psa.` ze statusem
`ok`. To są obserwacje działania pipeline'u, nie ręczna ocena jakości zdań.

Wcześniej, przed rejestracją providerów, 200 analogicznych wywołań zwracało
`unsupported / LANGUAGE_REGISTRY_UNAVAILABLE`; ten historyczny zapis znajduje
się w [observed-2026-10-23-d05212e.tsv](observed-2026-10-23-d05212e.tsv).
Po aktywowaniu bieżącego snapshotu uruchomiono wszystkie 100 par w obu
kierunkach. Pełne wyjścia z kodami wyjścia, statusami, kodami diagnostyk i
tekstami znajdują się w
[observed-2026-10-23-provider-runtime.tsv](observed-2026-10-23-provider-runtime.tsv).
Spośród 200 wywołań 12 miało status `ok`, 12 `incomplete`, 8 `ambiguous`,
4 `unknown`, a 164 `unsupported`. Tylko część przykładów korzysta z obecnego
szlaku `seeing`; pozostałe pokazują brak obsługi leksykalnej lub konstrukcyjnej.
Nie są to goldy ani automatyczna miara poprawności.

Ręczna ocena właściciela pozostaje **NOT RUN**. Nie zmieniano statusów
redakcyjnych wierszy. Odbiór językowy wymaga oceny konkretnych wyjść przez
właściciela projektu.

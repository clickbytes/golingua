# 100 par zdań PL/EN powiązanych z CORE

Na polecenie właściciela z 2026-10-23 zapisano 100 **redakcyjnych propozycji** zdań w [100-pl-en.tsv](100-pl-en.tsv). Dla każdego wiersza [100-meanings.jsonl](100-meanings.jsonl) zawiera odpowiadający mu przykład request-scoped `MeaningGraph` z konceptem zdarzenia i dwoma osobnymi referentami. `S001` w obu plikach oznacza ten sam przykład. Zdania nie są nowymi `ConceptDefinition`, `LexicalSense` ani zatwierdzonymi tłumaczeniami.

Materiał obejmuje po 25 zdań dla `seeing`, `hearing`, `looking` i `receiving`. Każdy użyty koncept i sens zdarzenia ma obecnie osobny zatwierdzony binding PL/EN w katalogu, ale zdania, formy czasowników, wybór rodzajnika, czas/aspekt oraz konstrukcje nie zostały przez to zatwierdzone. Angielska trasa konstrukcyjna `look at` nie ma jeszcze approved bindingu. Powtarzane kombinacje uczestników służą sprawdzeniu zamiany ról i odmiany; nie stanowią miary jakości.

Wszystkie 100 struktur `MeaningGraph` sprawdzono przez `meaning.ValidateMeaning` względem skompilowanego CORE. Skompilowane indeksy leksykalne PL/EN potwierdziły `approved` i `equivalent` dla każdego wskazanego konceptu oraz dokładnych sensów zdarzeń. Ta kontrola dotyczy identyfikatorów i ról, a nie poprawności powierzchni zdań.

Status wszystkich wierszy to `draft_unreviewed`. Autorem powierzchni PL/EN jest Codex; licencja tych nowych propozycji nie została osobno ustalona. Akceptacja właściciela: **NOT RUN**. `MeaningGraph` przechowuje tutaj typ zdarzenia i role uczestników; nie koduje czasu, aspektu, liczby ani określoności z zapisanych zdań. Dlatego pliki nie są goldami ani automatyczną bramką expected output.

## Obserwacja CLI

Gałęzie #195, #197 i #199 rejestrują skompilowane pakiety oraz etapy PL/EN dla
ograniczonych, zatwierdzonych szlaków `seeing`, `hearing` i `receiving`. Używają analizy
składniowej, dokładnych bindingów, ram walencji, planu konstrukcji, jawnych form
morfologicznych i linearyzacji. Angielski artykuł analizuje `the` i `a` jako
cechy; konstrukcja dopuszcza również brak rodzajnika dla bieżących rzeczowników
niepoliczalnych. Przy tłumaczeniu z PL określoność nie występuje w
`MeaningGraph`, więc wybór `the` jest jawnie oznaczany jako założenie i wynik
pozostaje `incomplete`. Trasa `receiving` zachowuje czas przeszły i uzgadnia
polski czasownik z rodzajem podmiotu w liczbie pojedynczej.

Przykładowe lokalne wywołania na aktywnym snapshotcie:

```sh
golingua translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config runtime.json --json
golingua translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config runtime.json --json
golingua translate --from pl --to en --text 'Nauczyciel słyszy psa.' --runtime-config runtime.json --json
golingua translate --from en --to pl --text 'The teacher hears the dog.' --runtime-config runtime.json --json
```

Pierwsze zwróciło dosłowny tekst `The teacher sees the dog.` ze statusem
`incomplete` i diagnostykami `EXPLICIT_REALIZATION_ASSUMPTION` oraz
`TRANSLATION_NOT_OK`. Drugie zwróciło `Nauczyciel widzi psa.` ze statusem
`ok`. To są obserwacje działania pipeline'u, nie ręczna ocena jakości zdań.

Wcześniejszy wynik 100 par z gałęzi #195 (przed aktywacją `hearing`) jest
zachowany w [observed-2026-10-23-provider-runtime.tsv](observed-2026-10-23-provider-runtime.tsv);
poza `seeing` pokazywał głównie brak obsługi. Jego statusy nie opisują bieżącego
runtime'u po #197.

Po #197 uruchomiono wszystkie 25 istniejących propozycji `hearing` w obu
kierunkach, łącznie z kodami wyjścia, statusami, diagnostykami i tekstami w
[observed-2026-10-23-g05-12-hearing.tsv](observed-2026-10-23-g05-12-hearing.tsv).
Przykład `S026` PL→EN zwrócił literalny tekst `The teacher hears the dog.` ze
statusem `incomplete`; EN→PL zwrócił `Nauczyciel słyszy psa.` ze statusem `ok`.
W tej próbce nierozstrzygnięte przypadki `child`/`doctor` pozostają
`ambiguous` albo `unsupported`. To są statusy techniczne z wywołań CLI, nie
ocena jakości ani gold.

Po #199 uruchomiono wszystkie 25 istniejących propozycji `receiving` w obu
kierunkach, z kodami wyjścia, statusami, diagnostykami i tekstami w
[observed-2026-10-23-g05-13-receiving.tsv](observed-2026-10-23-g05-13-receiving.tsv).
Angielski→polski wygenerował tekst dla wody, jedzenia, wiadomości i pieniędzy;
polski→angielski zachował tekst, lecz zwrócił `incomplete` z powodu jawnego
założenia rodzajnika `the`. Sensy książki oraz część odczytań `child` i `doctor`
pozostają niejednoznaczne lub niewspierane. To obserwacje runtime i ocena agenta
na wejściach właściciela; nie są akceptacją właściciela ani goldem.

**Ocena agenta:** w obsłużonych przykładach dobór `dostał` / `dostała` / `dostało`
zgadza się z rodzajem podmiotu, a `dostać` zachowuje przeszły czas wejścia
`got`. Polskie formy i angielskie zdania z `the` są gramatycznie możliwe; wybór
określoności w EN wymaga kontekstu, którego `MeaningGraph` nie zawiera, więc
wynik PL→EN pozostaje niekompletny. Nierozstrzygnięte odczytania `book`, `doctor`
i `child` są poprawnie ujawnione przez runtime; nie należy ich maskować przez
automatyczny wybór leksykalny. Pełny odbiór semantyczny wymaga kontekstu lub
rozstrzygnięcia tych sensów.

Wcześniej, przed rejestracją providerów, 200 analogicznych wywołań zwracało
`unsupported / LANGUAGE_REGISTRY_UNAVAILABLE`; ten historyczny zapis znajduje
się w [observed-2026-10-23-d05212e.tsv](observed-2026-10-23-d05212e.tsv).
Po aktywacji runtime'u #195, który obsługiwał tylko `seeing`, uruchomiono
wszystkie 100 par w obu kierunkach. Historyczne wyjścia znajdują się w
[observed-2026-10-23-provider-runtime.tsv](observed-2026-10-23-provider-runtime.tsv):
12 wywołań miało status `ok`, 12 `incomplete`, 8 `ambiguous`, 4 `unknown`,
a 164 `unsupported`. Te statusy opisują stan przed #197, nie są goldami ani
automatyczną miarą poprawności.

Akceptacja właściciela pozostaje **NOT RUN**. Nie zmieniano statusów
redakcyjnych wierszy. Właściciel delegował agentowi autonomiczną ocenę i
poprawianie rzeczywistych przykładów; ocena agenta nie jest akceptacją
właściciela.

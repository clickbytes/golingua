# 100 par zdań PL/EN powiązanych z CORE

Na polecenie właściciela z 2026-10-23 zapisano 100 **redakcyjnych propozycji** zdań w [100-pl-en.tsv](100-pl-en.tsv). Dla każdego wiersza [100-meanings.jsonl](100-meanings.jsonl) zawiera odpowiadający mu przykład request-scoped `MeaningGraph` z konceptem zdarzenia i dwoma osobnymi referentami. `S001` w obu plikach oznacza ten sam przykład. Zdania nie są nowymi `ConceptDefinition`, `LexicalSense` ani zatwierdzonymi tłumaczeniami.

Materiał obejmuje po 25 zdań dla `seeing`, `hearing`, `looking` i `receiving`. Każdy użyty koncept i sens zdarzenia ma obecnie osobny zatwierdzony binding PL/EN w katalogu, ale zdania, wybór rodzajnika i własności MeaningGraph nie zostały przez to zatwierdzone. W G05 #201 zatwierdzono dokładne szlaki konstrukcyjne EN `look at` oraz PL `patrzeć na` + biernik. Powtarzane kombinacje uczestników służą sprawdzeniu zamiany ról i odmiany; nie stanowią miary jakości.

Wszystkie 100 struktur `MeaningGraph` sprawdzono przez `meaning.ValidateMeaning` względem skompilowanego CORE. Skompilowane indeksy leksykalne PL/EN potwierdziły `approved` i `equivalent` dla każdego wskazanego konceptu oraz dokładnych sensów zdarzeń. Ta kontrola dotyczy identyfikatorów i ról, a nie poprawności powierzchni zdań.

Status wszystkich wierszy to `draft_unreviewed`. Autorem powierzchni PL/EN jest Codex; licencja tych nowych propozycji nie została osobno ustalona. Akceptacja właściciela: **NOT RUN**. `MeaningGraph` przechowuje tutaj typ zdarzenia i role uczestników; nie koduje czasu, aspektu, liczby ani określoności z zapisanych zdań. Dlatego pliki nie są goldami ani automatyczną bramką expected output.

## Obserwacja CLI

Gałęzie #195, #197, #199 i #201 rejestrują skompilowane pakiety oraz etapy PL/EN dla
ograniczonych, zatwierdzonych szlaków `seeing`, `hearing`, `looking` i `receiving`. Używają analizy
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
golingua translate --from en --to pl --text 'The teacher looks at the dog.' --runtime-config runtime.json --json
golingua translate --from pl --to en --text 'Nauczyciel patrzy na psa.' --runtime-config runtime.json --json
```

Pierwsze zwróciło dosłowny tekst `The teacher sees the dog.` ze statusem
`incomplete` i diagnostykami `EXPLICIT_REALIZATION_ASSUMPTION` oraz
`TRANSLATION_NOT_OK`. Drugie zwróciło `Nauczyciel widzi psa.` ze statusem
`ok`. To są obserwacje działania pipeline'u, nie ręczna ocena jakości zdań.

Po G05 #201 uruchomiono wszystkie 25 istniejących propozycji `looking` w obu
kierunkach; dosłowne kody wyjścia, statusy, diagnostyki i teksty są w
[observed-2026-10-23-g05-14-looking.tsv](observed-2026-10-23-g05-14-looking.tsv).
Wyniki: 12 `ok`, 12 `incomplete`, 8 `ambiguous` i 18 `unsupported` spośród 50
wywołań. Jednoznaczne podmioty z `dog`, `car` i `tree` tłumaczą się EN→PL ze
statusem `ok`; PL→EN zachowuje tekst, ale pozostaje `incomplete` z powodu
jawnego założenia rodzajnika `the`. Formy `drzewo` i `tree` są dostępne po obu
stronach. Nierozstrzygnięte `book`, `house`, `doctor` i `child` nie są wymuszane;
ich statusy pozostają `ambiguous` lub `unsupported` zależnie od kierunku i
pozostałych ról zdania.

**Ocena agenta:** w wyjściach z jednoznacznymi podmiotami zachowano relację
patrzenia skierowanego na cel. Angielskie `look at` jest oddzielone od
bezpośredniego `see`, a polskie `patrzeć na` zachowuje wymagane `na` i biernik
(`psa`, `samochód`, `drzewo`). `Nauczyciel`, `Matka`, `Ojciec` i `Dziecko`
zachowują rolę agensa. Niejednoznaczne wejścia nie powinny być oceniane jako
błędy leksykalne, dopóki ich odczytanie nie zostanie rozstrzygnięte. Właścicielska
akceptacja zdań i form nadal **NOT RUN**.

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

Po G10-06 #225 wszystkie 100 istniejących par uruchomiono ponownie w obu
kierunkach, a tekstowe kandydatury dla wyników `incomplete`/`ambiguous`
wybrano jawnie przez `--preview-id`. Dokładne odpowiedzi, ID, diagnostyki i
wyniki podglądów są w
[observed-2026-10-23-g10-07-runtime-preview.tsv](observed-2026-10-23-g10-07-runtime-preview.tsv).
W tej obserwacji EN→PL miało 60 wyników `ok` i 40 `incomplete`; PL→EN miało
75 `incomplete` i 25 `ambiguous`. Kandydatury z nierozstrzygniętymi sensami
pozostają widoczne. Dla policzalnych rzeczowników w liczbie pojedynczej `book` i `message` realizacja z PL
przyjęła `the`, podczas gdy zapisane EN pary mają `a`; brak określoności w
MeaningGraph pozostaje jawny i jest śledzony w G10-08 #228. Akceptacja
właściciela pozostaje **NOT RUN**.

Po G10-08 #228 zapisano nowe obserwacje z istniejących S076/S077/S079/S080 w
[observed-2026-10-23-g10-08-indefinite-receiving.tsv](observed-2026-10-23-g10-08-indefinite-receiving.tsv).
PL→EN zachowuje teraz warianty `a` i `the` dla policzalnych `book` i
`message`, a oba mają osobne założenia i własne candidate ID. Zbiorczy wynik
jest `ambiguous`; jawny `--preview-id` wyświetla wybrany tekst bez zmiany tego
statusu. Jawnie niepoliczalne `water` i `money` nadal realizują się bez
rodzajnika. EN→PL dla tych samych istniejących par zachowało teksty polskie.
Wiersze pozostają `draft_unreviewed`; akceptacja właściciela jest **NOT RUN**.

Po G10-10 #233 wszystkie 100 par uruchomiono ponownie w obu kierunkach.
Dokładne statusy, kandydatury i wybrane podglądy zapisano w
[observed-2026-10-23-g10-10-determiner-alternatives.tsv](observed-2026-10-23-g10-10-determiner-alternatives.tsv).
PL→EN zachowuje `a/the` dla podmiotu i dopełnienia/celu w istniejących
transitive i oblique routes, a masowe `water` / `money` nadal mają wariant bez
rodzajnika. EN→PL zachowało dotychczasowe teksty. Wyniki wciąż pokazują
niejednoznaczność i niezatwierdzone odczytania; akceptacja właściciela jest
**NOT RUN**.

Po G10-11 #235 (PR #236) uzupełniono sentence-facing formy istniejącego EN `household`; zapis CLI dla 10 już obecnych rekordów i obu kierunków znajduje się w [observed-2026-10-23-g10-11-household-morphology.tsv](observed-2026-10-23-g10-11-household-morphology.tsv). Warianty `household` i `house` pozostają osobnymi kandydatami, a wyniki zachowują `ambiguous`/`incomplete`; akceptacja właściciela jest **NOT RUN**.

Po G10-12 #237 ponownie uruchomiono istniejące rekordy S021–S025, S036–S040,
S071–S075 i S096–S100 w obu kierunkach. Dokładne odpowiedzi kandydatów,
kody wyjścia, statusy, diagnostyki i teksty znajdują się w
[observed-2026-10-23-g10-12-offspring-a-an.tsv](observed-2026-10-23-g10-12-offspring-a-an.tsv).
EN→PL zwróciło 16 wyników `ok` i 4 `incomplete`; niekompletne odczytania
`book` zachowują `PARTIAL_PARSE_EVIDENCE` i `UNRESOLVED_ROLE_SENSE`. PL→EN
zachowało 20 wieloznacznych wyników z odrębnymi kandydaturami `child` i
`offspring` oraz wariantami `a`/`an`/`the`; brak określoności w źródle nadal
jest jawnym założeniem, a nie automatycznie wybranym tłumaczeniem. W tej
próbce nie wystąpiły `ROLE_LEXICALIZATION_UNAVAILABLE` ani
`PRINCIPAL_PARTS_UNAVAILABLE`.

W tym samym pliku osobny wiersz `DERIVED-EN-01` rejestruje dodatkową,
pochodną próbę `An offspring sees a dog.` → `Dziecko widzi psa.` (`ok`, exit
0). Forma `A offspring...` została odrzucona przez parser; jawny zero plural
ma formy `offspring` dla singular i plural, a `an` jest analizowany jako
determiner indefinite. To obserwacje CLI i ocena agenta, nie formalny gold ani
akceptacja właściciela. Akceptacja językowa właściciela pozostaje **NOT RUN**.

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

Po G10-12 #237 (PR #238) ponownie uruchomiono istniejące rekordy S021–S025, S036–S040,
S071–S075 i S096–S100 w obu kierunkach. Dokładne odpowiedzi kandydatów,
kody wyjścia, statusy, diagnostyki i teksty znajdują się w
[observed-2026-10-23-g10-12-offspring-a-an.tsv](observed-2026-10-23-g10-12-offspring-a-an.tsv).
W 20 istniejących rekordach EN→PL zwróciło 15 wyników `ok` i 5 `incomplete`;
niekompletne odczytania `book` zachowują `PARTIAL_PARSE_EVIDENCE` i
`UNRESOLVED_ROLE_SENSE`. PL→EN
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

Po G10-13 #239 (PR #240) obserwacje dla istniejących S021–S025, S036–S040,
S071–S075 i S096–S100 ponownie uruchomiono w obu kierunkach. Wszystkie 40
wywołań zachowało statusy i zbiory tekstów z poprzedniej obserwacji: EN→PL
15 `ok` / 5 `incomplete`, PL→EN 20 `ambiguous`. Plik
[observed-2026-10-23-g10-13-plural-agreement.tsv](observed-2026-10-23-g10-13-plural-agreement.tsv)
zawiera też 10 osobno oznaczonych prób pochodnych. W polach diagnostyk `-`
oznacza brak diagnostyki. Pięć zgodnych prób EN→PL (jedna singularna i cztery
mnogie) dało `ok`; trzy wejścia PL→EN zachowały mnogie formy czasowników w
kandydatach `ambiguous`, a dwie niezgodne pary liczby nie uzyskały parse.
`forms` wygenerowało `see`, `hear`, `look`, `widzą`, `słyszą` i `patrzą` dla
komórek trzeciej osoby mnogiej. To obserwacje działania CLI, nie gold ani odbiór
językowy; akceptacja właściciela pozostaje **NOT RUN**.

## Pełny przebieg G10-14 — 2026-10-23

Na aktualnym CLI z implementacji
`33b8830cae5b285fbfafacdbb0e90c85c7916aa6` uruchomiono wszystkie 100 istniejących
wierszy w obu kierunkach: 200 rzeczywistych wywołań. Dokładne statusy, kody
wyjścia, kandydatury tekstowe i diagnostyki są w
[observed-2026-10-23-g10-14-full-corpus.tsv](observed-2026-10-23-g10-14-full-corpus.tsv)
(592 wiersze kandydatur; `-` w polach diagnostyki oznacza brak diagnostyki).
EN→PL zwróciło 60 wyników `ok` i 40 `incomplete`; każdy wynik zachował tekst.
PL→EN zwróciło 100 wyników `ambiguous` z widocznymi kandydaturami; wszystkie
592 kandydatury mają tekst, a każda z 492 kandydatur PL→EN ma status lokalny
`incomplete` z jawnymi założeniami/alternatywami.

**Ocena agenta:** `The child sees the dog.` daje `Dziecko widzi psa.` (`ok`).
`Dziecko widzi psa.` zachowuje warianty `child`/`offspring` oraz rodzajniki
`a`/`the` jako alternatywy. W EN→PL niekompletne wyniki pozostawiają tekst i
evidence, gdy konkurencyjny sens nie ma zatwierdzonego CORE odpowiednika (np.
`en.book.work`, `en.doctor.academic`, `en.house.institution`); agent nie uznaje
tych przypadków za awarię realizatora ani nie awansuje bindingów. Odbiór
językowy właściciela pozostaje **NOT RUN**.

## Polski czas przeszły mnogi G10-15 — 2026-10-23

Na implementacji `c97a6441c62102b098f734d0945f379064402e95` uruchomiono cztery
oznaczone jako pochodne smoke CLI oraz cztery wywołania `forms`; dokładne wyniki
znajdują się w
[observed-2026-10-23-g10-15-plural-past.tsv](observed-2026-10-23-g10-15-plural-past.tsv).
Angielskie podmioty mnogie `teachers` i `children` uzyskały polskie formy
`dostali` i `dostały`. W kierunku PL→EN kandydatury zachowały `got`, ale
określoność i alternatywne sensy nadal dają status `incomplete`/`ambiguous`.
Ponowne uruchomienie S001–S100 w obu kierunkach (200 wywołań) zachowało
wszystkie statusy i zbiory tekstów kandydatów z G10-14; nie dodawano wierszy
do korpusu. Ocena agenta jest zapisana oddzielnie od odbioru właściciela,
który pozostaje **NOT RUN**.

## Negacja `seeing` G10-16 — 2026-10-23

Na implementacji `08dd82c87e201161e47d8e277048fe6848e13f73` EN
`The teacher does not see the dog.` → PL `Nauczyciel nie widzi psa.` zwróciło
`ok` (exit 0). W kierunku PL→EN ten sam przykład wyprodukował kandydata
`The teacher does not see the dog.` ze statusem `incomplete` (exit 7), ponieważ
Meaning nie określa angielskich rodzajników. Dodatnie counterpart EN→PL
zachowało `ok`; PL→EN zachowało dotychczasowe cztery warianty dodatnie i nie
wybrało rodziny negatywnej. Szczegóły znajdują się w
[observed-2026-10-23-g10-16-negated-seeing.tsv](observed-2026-10-23-g10-16-negated-seeing.tsv).

**Ocena agenta:** ograniczona negacja poprawnie tworzy i zachowuje scope Meaning
w EN→PL oraz realizuje jawne terminale `does not` / `nie`. Angielski kandydat
z polskiego wejścia zachowuje negację, ale status `incomplete` jest prawidłowy,
bo określoność nie jest zakodowana w Meaning. Ponowny przebieg wszystkich 100
istniejących par w obu kierunkach wykonał 200 wywołań i zachował statusy oraz
zbiory tekstów kandydatów G10-14. Odbiór językowy właściciela pozostaje
**NOT RUN**.

## Negacja teraźniejsza `seeing`, `hearing`, `looking` G10-17 — 2026-10-23

Na implementacji `e1c9166b5ac7fa73fc39e650512fd033c4b4e05d` uruchomiono 18
pochodnych wejść CLI na snapshotcie utworzonym z bieżącego katalogu. EN→PL
zwróciło `ok` dla 3sg/3pl `does/do not see`, `hear` i `look at`, dając
odpowiednio `nie widzi/widzą`, `nie słyszy/słyszą` i `nie patrzy/patrzą na`.
PL→EN zachowało odpowiadające kandydatury z `does/do not`; wszystkie sześć
pozostało `incomplete`, ponieważ określoność nie występuje w polskim Meaning.
Zamiana ról w obu kierunkach zachowała właściwego experiencera/agensa i
stimulus/target. Cztery próby niezgodnej liczby albo czasu przeszłego zwróciły
`unsupported`.

Ponownie uruchomiono wszystkie 100 istniejących par w obu kierunkach — 200
wywołań. Każda grupa zachowała zbiorczy status i zbiór par `(candidate
outcome, text)` względem G10-14; zmienionych grup: 0. EN→PL miało 60 `ok` i
40 `incomplete`, a PL→EN 100 `ambiguous`. Nie dodano ani nie edytowano wierszy
korpusu. Dokładne wyniki smoke i skrót porównania są w
[obserwacji G10-17](observed-2026-10-23-g10-17-present-negation.tsv).

**Ocena agenta:** rozpoznane trasy zachowują negację, role, liczbę i osobną
rekcję oblique dla `look at` / `patrzeć na`. Status `incomplete` PL→EN
ujawnia brak określoności w źródle, a próby spoza zakresu nie zostały
zaakceptowane przez nową konstrukcję. Akceptacja językowa właściciela pozostaje
**NOT RUN**.


## Rodzajniki na negatywnych trasach G10-18 — 2026-10-23

Negacja `see`, `hear` i `look at` zachowuje jawne warianty `a`, `an` oraz `the`
dla podmiotu i dopełnienia/celu w obu liczbach. `a/an` wymaga właściwości
leksemu i liczby pojedynczej; brak określoności w polskim Meaning pozostawia
oddzielne angielskie kandydatury ze statusem zbiorczym `ambiguous`. Próby CLI,
próby ochronne i wynik regresji 200 wywołań są w
[obserwacji G10-18](observed-2026-10-23-g10-18-negative-article-alternatives.tsv).
Ocena agenta jest zapisana osobno; odbiór właściciela pozostaje **NOT RUN**.

## Trasa `reading` G10-19 — 2026-10-23

G10-19 aktywuje istniejące zatwierdzone sensy PL/EN `reading` w present 3sg/3pl.
EN→PL `The teacher reads the message.` i `The teachers read the message.`
zwróciło `ok`, odpowiednio `Nauczyciel czyta wiadomość.` i `Nauczyciele
czytają wiadomość.`. PL→EN zachowało odpowiednie angielskie teksty jako
kandydaty `incomplete` z dwoma jawnymi założeniami `the`, ponieważ polskie
Meaning nie przechowuje określoności. `forms` analizuje `read/reads` oraz
`czyta/czytają`; `czytać` ma jawny aspekt imperfective i używa stabilnego ID
`lexeme:pl.czytac.v`. Niezgodne uzgodnienie, homograficzny past `read` oraz
negacja pozostają `unsupported`. Regresja 100 istniejących par wykonała 200
wywołań i zachowała wszystkie statusy oraz zbiory `(candidate outcome, text)`
G10-14; zero zmienionych grup. Nie zmieniono 100-zdaniowego korpusu. Szczegóły
i dokładne wyniki są w
[obserwacji G10-19](observed-2026-10-23-g10-19-reading-route.tsv).

**Ocena agenta:** role czytelnika i materiału, liczba oraz ograniczenie czasu
są zachowane; PL→EN jawnie ujawnia brak określoności. Akceptacja językowa
właściciela pozostaje **NOT RUN**.

## Trasa `reading` past G10-20 — 2026-10-23

G10-20 dodaje zdaniową analizę i realizację czasu przeszłego dla istniejących
sensów `reading`, bez dodawania wierszy do korpusu. EN→PL dla
`The teacher/mother/child read the message.` generuje odpowiednio
`Nauczyciel czytał wiadomość.`, `Matka czytała wiadomość.` i
`Dziecko czytało wiadomość.`. Dla `teachers` i `children` angielskie `read` ma
zarówno teraźniejszą, jak i przeszłą analizę, a runtime zachowuje oba polskie
teksty jako `ambiguous`. PL→EN używa simple past `read`; brak określoności w
polskim źródle zachowuje alternatywy `a/the`, a `dziecko` również ma
alternatywę leksemu.

Paradygmat `pl.czytac.v` zawiera 15 jawnych komórek czasu przeszłego
(1./2./3. osoba, liczba pojedyncza/mnoga, odpowiednia klasa rodzaju) ze
źródłem WSJP. Istniejący korpus S001–S100 uruchomiono ponownie w obu kierunkach:
200 wywołań, bez zmian w statusach ani parach `(candidate outcome, text)`
względem G10-14. Dokładne wyjścia tras i form są w
[obserwacji G10-20](observed-2026-10-23-g10-20-reading-past.tsv); przebieg
pełnego korpusu jest w
[osobnej tabeli](observed-2026-10-23-g10-20-full-corpus.tsv). Ocena agenta jest
zapisana, akceptacja właściciela pozostaje **NOT RUN**.

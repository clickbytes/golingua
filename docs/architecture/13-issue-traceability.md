# 13 — Mapa pracy szeregowej po resecie testów

**Status: żywa mapa etapów; G10-20 #253 (PR #255) rozszerza istniejący szlak `reading` o jawny simple past EN i polskie formy przeszłe.** G00–G08 są technicznie zamknięte; ręczna akceptacja językowa właściciela pozostaje osobnym statusem. G04 #5/#207 ma zamknięty closeout techniczny. G06 #7 zamyka zakres techniczny przez #128 i delegowaną ocenę #130; akceptacja właściciela pozostaje **NOT RUN**. G08 #9 ukończono po wszystkich child issues (#65–70); właściciel delegował autonomiczną implementację i ocenę, więc jego odbiór pozostaje **NOT RUN**, ale nie blokuje technicznego zamknięcia. #66 w PR #213 dodało wariant bez rodzajnika dla jawnie niepoliczalnych leksemów na trasie `receiving`. G10 #214 zamknięto po PR #255; aktywnym epikiem jest G11 #254, a aktywnym zadaniem G11-02 #258; G11-01 #256 ukończono w PR #257; G10-01–08 (#215/#217/#219/#221/#223/#225/#227/#228) ukończono w PR #216/#218/#220/#222/#224/#226/#229/#230; G10-09 #231 zsynchronizował liczniki w PR #232; G10-10 #233 (PR #234) rozszerzył alternatywy rodzajnika na istniejące trasy; G10-11 #235 (PR #236) uzupełnił morfologię `household`; G10-12 #237 (PR #238) ukończono; G10-13 #239 (PR #240) dodało mnogie formy czasu teraźniejszego i twarde uzgodnienie liczby; G10-14 #241 zapisało wyniki 100 wierszy PL/EN; G10-15 #243 rozszerzyło polski czas przeszły `receiving` na liczbę mnogą; G10-16 #245 dodaje negację `seeing` i jej obsługę w realizatorze; G10-17 #247 rozszerza negację na trzy trasy i obie liczby; G10-18 #249 dodaje warianty `a/an/the`; G10-19 #251 aktywuje present `reading`; G10-20 #253 (PR #255) dodało `reading` w past i zamknęło techniczny zakres G10. `catalog --summary --json` raportuje: CORE 1000; PL 983 leksemy / 1007 sensów / 1006 bindingów (882 zatwierdzone koncepty, 118 konceptów pending); EN 1003 leksemy / 1012 sensów / 1001 bindingów (997 zatwierdzonych konceptów, 3 koncepty pending); 881 zatwierdzonych par PL+EN. Liczba pending binding records PL wynosi 124, bo część konceptów ma alternatywne kandydatury; EN ma 4 pending records. Angielskie determinery `a` i `an` są sentence-only i nie mają CORE bindingów. Kompilator CORE udostępnia obecnie 173 dokładne relacje pochodzące z event concepts z dwoma zadeklarowanymi rolami. Historyczne G00 #78 nie jest obecnym aktywnym zadaniem.

| Kolejność | EPIC | Co zbudować | Jak oceniamy na tym etapie |
| --- | --- | --- | --- |
| 0 | [G00 #1](https://github.com/clickbytes/golingua/issues/1) | Bootstrap, granice typów, diagnostyka, źródła danych, prosty CLI | kompilacja Go i zrozumiała pomoc CLI |
| 1 | [G01 #2](https://github.com/clickbytes/golingua/issues/2) | ConceptGraph, identyfikatory, MeaningGraph | czy API ładnie reprezentuje i wypisuje pojęcia, bez unit test suite |
| 2 | [G02 #3](https://github.com/clickbytes/golingua/issues/3) | Lexicon, morphology, procedury i wyjątki | wybrane przez właściciela słowa i odmiany |
| 3 | [G03 #4](https://github.com/clickbytes/golingua/issues/4) | ValencyFrame, ConstructionFamily/Variant, NP/VP/PP, liczebniki/time/MWE | ręczna próba wybranych konstrukcji użytkownika |
| 4 | [G04 #5](https://github.com/clickbytes/golingua/issues/5) | Tworzenie/import danych i kompilacja niemutowalnych pakietów | implementacja gotowa; #207 weryfikuje snapshot i odmianę na CLI przed closeoutem |
| 5 | [G05 #6](https://github.com/clickbytes/golingua/issues/6) | Analyze i Realize PL↔EN | rzeczywiste wejścia właściciela; ocena agenta jest delegowana, akceptacja właściciela osobna |
| 6 | [G06 #7](https://github.com/clickbytes/golingua/issues/7) | ograniczony reasoner i ContentPlanner | delegowana próba typed CLI; akceptacja językowa właściciela osobno |
| 7 | [G07 #8](https://github.com/clickbytes/golingua/issues/8) | około 1000 konceptów z PL/EN | zakres katalogu ukończony; closeout #205, ręczny odbiór zdań pozostaje osobnym NOT RUN |
| 8 | [G08 #9](https://github.com/clickbytes/golingua/issues/9) | uporządkowanie CLI, prosta diagnostyka, optymalizacja tylko gdy potrzebna | ograniczona ocena agenta na zapisanych trasach; akceptacja właściciela osobno |
| 9 | [G09 #10](https://github.com/clickbytes/golingua/issues/10) | przyszła pamięć i uczenie | DEFERRED, brak implementacji v0 |
| 10 | [G10 #214](https://github.com/clickbytes/golingua/issues/214) | proceduralny model języka PL/EN: odmiana, cechy i konstrukcje prowadzące do szerszej obsługi zdań | G10-01–08 ukończone przez PR #216/#218/#220/#222/#224/#226/#229/#230; G10-09 #231 przez PR #232; G10-10 #233 (PR #234) rozszerzyło warianty rodzajnika; G10-11 #235 (PR #236) dodało formy `household`; G10-12 #237 (PR #238) dodało `offspring`, zero plural i rozróżnianie `a/an`; G10-13 #239 (PR #240) dodało present 3pl i uzgodnienie `number`; G10-14 #241 (PR #242) ponownie uruchomiło 100 wierszy; G10-15 #243 rozszerzyło polski past `receiving`; G10-16 #245 dodało singularną negację `seeing`; G10-17 #247 obejmuje negację teraźniejszą na wszystkich trzech trasach i w obu liczbach; G10-18 #249 zachowuje tam jawne alternatywy rodzajnika; G10-19 #251 aktywuje present `reading` z `read/reads` i `czyta/czytają`; G10-20 #253 (PR #255) dodało past `read` oraz 15 jawnych komórek przeszłych `czytać`; techniczny zakres G10 jest zamknięty

| 11 | [G11 #254](https://github.com/clickbytes/golingua/issues/254) | spójne CLI deweloperskie i pięcioekranowy TUI oparty o te same usługi Go | G11-01 #256 COMPLETE (PR #257); G11-02 #258 ACTIVE: typowane usługi in-process dla CLI/TUI |

### G10-19 #251 — aktywacja `reading` w present — 2026-10-23

Aktywuje istniejące zatwierdzone sensy `pl.czytac.interpret` i `en.read.text`
w trzeciej osobie liczby pojedynczej i mnogiej. `read/reads` ma jawne principal
parts; polski paradygmat `czytać` został podłączony do stabilnego ID
`lexeme:pl.czytac.v` i dostał jawny imperfective, wymagany przez analizę
zdania. EN→PL `The teacher reads the message.` oraz `The teachers read the
message.` zwróciły `ok` z `czyta/czytają`. PL→EN zachowało odpowiednie teksty
`The teacher reads the message.` i `The teachers read the message.` jako
kandydatów `incomplete` z jawnymi założeniami `the`, bo polski Meaning nie
koduje określoności. Błędne uzgodnienie, homograficzny past `read` i negacja
pozostały `unsupported`. Ponowny przebieg S001–S100 (200 wywołań) zachował
statusy i zbiory `(candidate outcome, text)` G10-14 we wszystkich grupach.
Szczegóły form i CLI: [obserwacja G10-19](../../data/sentences/observed-2026-10-23-g10-19-reading-route.tsv).
Ocena agenta jest zapisana; akceptacja językowa właściciela pozostaje
**NOT RUN**.

### G10-20 #253 (PR #255) — dodanie past do `reading` — 2026-10-23

Dodaje trzecią osobę liczby pojedynczej i mnogiej dla angielskiego simple past
`read` oraz dla polskiego imperfective `czytać`, z jawną zgodą rodzaju w PL.
Obserwacje CLI: EN→PL `The teacher/mother/child read the message.`
zwróciło odpowiednio `Nauczyciel czytał wiadomość.`, `Matka czytała
wiadomość.` i `Dziecko czytało wiadomość.` ze statusem `ok`. `The teachers /
children read ...` zachowało warianty present/past jako `ambiguous`. PL→EN
wygenerowało past `read`; liczba pojedyncza i mnoga pozostają `ambiguous` przy
wariantach rodzajnika, a `dziecko` ma też alternatywę leksemu. Obecny korpus
S001–S100 uruchomiono ponownie: 200 wywołań, zero zmienionych grup statusu i
zbiorów `(candidate outcome, text)` względem G10-14. Szczegółowe obserwacje są
w [tabeli tras i form](../../data/sentences/observed-2026-10-23-g10-20-reading-past.tsv)
oraz [tabeli pełnego korpusu](../../data/sentences/observed-2026-10-23-g10-20-full-corpus.tsv).
Ocena agenta jest zapisana; odbiór językowy właściciela pozostaje **NOT RUN**.

### G11-01 #256 — kontrakt CLI/TUI — 2026-10-23

G10 #214 zamknięto po PR #255; G11 #254 jest aktywnym epikiem z jednym
aktywnym child issue. G11-01 ustala, że no-args na interaktywnym TTY otworzy
TUI dopiero po wdrożeniu G11-04; brak TTY i jawne `--help` zachowują pomoc, a
istniejące polecenia nie zmieniają JSON-u ani exit codes. Wybrano Bubble Tea v2
(`charm.land/bubbletea/v2`), pięć ekranów F1–F5, ustawienia UI jako wersjonowany
JSON pod `os.UserConfigDir()`, potwierdzane `Apply` snapshotu i fail-closed
prezentację statusów. Usługi TUI mają pozostać in-process i współdzielić
engine/CLI API. PR #257 scalono do `main`. Szczegóły są w [kontrakcie](../tui/README.md), [screen map](../tui/screens.md),
[keymap](../tui/keymap.md), [binding map](../tui/bindings.md) i
[kontrakcie błędów](../tui/errors.md). Dokumentacja nie oznacza, że TUI już
działa; obserwacja TUI pozostaje **NOT RUN**.

### G11-02 #258 — typowany serwis aplikacyjny — 2026-10-23

`internal/app.Service` dostarcza wspólne metody `Translate`, `Forms`,
`Catalog`, `LookupConcept` i `Snapshot` wraz z typowanymi żądaniami i
wynikami. `cmd/golingua` przekazuje jedną instancję serwisu do istniejących
handlerów, a `internal/cli` zachowuje parsowanie, JSON, stderr i kody procesu.
Question pozostaje poza migracją. Snapshot `Apply` wymaga flagi potwierdzenia
na granicy usługi; CLI ustawia ją wyłącznie przy jawnym `--apply`. Szczegóły
bindingów są w [mapie TUI](../tui/bindings.md), a lokalne obserwacje w
[manual review](../manual-review.md). Vet/build i CLI smoke przeszły lokalnie;
GitHub Actions `go-ci/build` run #39922637020 dla PR #259 przeszedł. Akceptacja
właściciela pozostaje **NOT RUN**.

## Kryterium wykonania tasku

1. Jeden ACTIVE EPIC, jeden task i jeden agent; zwykle jeden PR. Wyjątek właściciela dla #60: sekwencyjne commity na main i CI.
2. Wprowadź nazwane struktury/funkcje/pliki; nie rozbudowuj rozwiązania ponad granice etapu.
3. CI Go: format, vet, build i start CLI. Brak obowiązkowych unit tests.
4. Gdy funkcja wytwarza formy/zdania, uruchom dostępne przykłady właściciela i zapisz dokładne obserwacje; agent samodzielnie ocenia i poprawia w granicach delegacji. Akceptacji właściciela nie oznaczaj jako wykonanej, jeśli jej nie udzielił.
5. Zapisz realne błędy i poprawki w issue; nie ma automatycznej akceptacji językowej na podstawie starszych wyników.
6. Po scaleniu i ponownym sprawdzeniu głównej gałęzi rozpocznij kolejny task.

## Bieżące szlaki i obserwacje

Aktualny stan katalogu po G10-20 #253, bez zmian w licznikach leksykalnych: 983 leksemy PL, 1007 sensów, 1006 bindingów (882 approved i 124 pending), 1003 leksemy EN, 1012 sensów i 1001 CORE bindingów (997 approved i 4 pending records). Wszystkie 1000 konceptów mają rekordy; PL ma 118 pending konceptów, a EN 3. Angielskie determinery `a` i `an` są sentence-only i nie mają CORE bindingów. Indeks konstrukcji ma 117 zatwierdzonych wiązań z sensem/ramą (PL 8, EN 109) po review #63, #201, #213, #230, #233, #237, #245, #247, #249, #251 i #253. Katalog raportuje 881 PL+EN approved pairs. CLI obsługuje ograniczone szlaki `seeing`, `hearing`, `looking`, `reading` i `receiving`, w tym negację `seeing`, `hearing` i `looking` w liczbie pojedynczej i mnogiej; rzeczywiste wyjścia 25 przykładów `seeing` znajdują się w [obserwacjach początkowego runtime'u](../../data/sentences/observed-2026-10-23-provider-runtime.tsv), `hearing` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-12-hearing.tsv), `receiving` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-13-receiving.tsv), `looking` w [najnowszej tabeli](../../data/sentences/observed-2026-10-23-g05-14-looking.tsv), negacji singularnego `seeing` w [obserwacji G10-16](../../data/sentences/observed-2026-10-23-g10-16-negated-seeing.tsv), wszystkich tras w [obserwacji G10-17](../../data/sentences/observed-2026-10-23-g10-17-present-negation.tsv), a alternatyw rodzajnika na tych trasach w [obserwacji G10-18](../../data/sentences/observed-2026-10-23-g10-18-negative-article-alternatives.tsv), a trasy `reading` w [obserwacji G10-19](../../data/sentences/observed-2026-10-23-g10-19-reading-route.tsv) i [obserwacji G10-20](../../data/sentences/observed-2026-10-23-g10-20-reading-past.tsv). Pełny przebieg 100 par po G10-20 zachował statusy i zbiory tekstów względem G10-14 (200 wywołań, zero zmienionych grup). Katalog G07 i jego zadania są technicznie ukończone; 118 konceptów PL i 3 EN pozostają pending, bez automatycznego awansu. Akceptacja językowa właściciela dla wyników zdań pozostaje **NOT RUN**; obserwacje działania i ocena agenta są udokumentowane niezależnie. Nie wyciągać wniosków o jakości na podstawie liczników lub samego statusu wykonania.

## Co zostało świadomie usunięte

Nie istnieje obowiązek przeniesienia dawnych zdań, wejść oracli i spodziewanych tekstów JS. Nie planujemy nowego katalogu `fixtures/parity`, automatycznych runnerów w `internal/testing`, protokołu niezależnego recenzowania expected zdań ani przelicznika „ile zdań przeszło”. To nowe źródło danych i nowy silnik.

Po decyzji użytkownika przyjmujemy również, że nie ma wymaganego dotted/property API `ROCK.VERB...`. Typowanie i kompozycja pozostają kluczowe, sama nawigacja API nie.

G03-08 #74, G03-09 #75 i G04-07 #76 są zamknięte historycznie. G04 #38–43/#76/#207 zakończyły implementację i obserwację techniczną. G05 #123, #195, #197, #199 i #201 są zakończone; #203 zsynchronizował mapę i G05. G07 #58–64, #145, #152–192 są zamknięte; #205 zsynchronizował closeout epiku, bez zmiany nierozstrzygniętych kandydatur ani statusu oceny językowej. G06 #128 (PR #209) i #130 (PR #210) zakończyły techniczny zakres epiku; delegowana ocena użyła przykładu S001 i tymczasowego faktu demonstracyjnego. Runtime realizuje dodatnie odpowiedzi yes/no PL/EN na zatwierdzonych trasach ze statusem `incomplete` i jawnymi założeniami. Odpowiedzi false mogą użyć ograniczonej negacji `seeing`, `hearing` i `looking` w 3. osobie liczby pojedynczej lub mnogiej; pozostałe odpowiedzi negatywne i WH zachowują reasoning/proof, ale tekst pozostaje `unsupported` bez zgodnej konstrukcji. Akceptacja językowa właściciela pozostaje **NOT RUN**. G08 #9 zamknięto po #65–70 i scaleniu PR #213; na S077 `Nauczyciel dostał wodę.` obserwowano `The teacher got water.` / `incomplete` / exit 7 z jawnym założeniem rodzajnika, a EN→PL zwróciło `Nauczyciel dostał wodę.` / `ok` / exit 0. Rzeczownik policzalny w S079 zachował `The teacher got the message.` / `incomplete` / exit 7. Trzy rekordy pozostają `draft_unreviewed`; odbiór językowy właściciela pozostaje **NOT RUN** i nie blokuje dalszej pracy zgodnie z jego decyzją z 2026-10-23. G10-01 #215 zakończono przez PR #216, który dodał angielskie finite classes. G10-02 #217 zakończono przez PR #218: dodano polskie finite cells dla wybranych czasowników, z aspektami i uzgodnieniem. Obserwacje na 100 istniejących parach zachowały dotychczasowe kody wyjścia/statusy i zbiory tekstów kandydatów; rozszerzenie morfologii samo nie oznacza, że 100 przykładów jest już w pełni tłumaczonych. G10-03 #219 (PR #220) dodało angielskie komórki liczby rzeczowników dla wybranych regularnych klas oraz jawnych form nieregularnych. G10-04 #221 (PR #222) dodało 14 jawnych komórek ośmiu polskich rzeczowników oraz 133 komórki przymiotnika `dorosły`; `mewa` pozostaje leksemem morfologicznym bez sensu i bindingu. Jego przykłady CLI działają, ale językowa akceptacja właściciela pozostaje NOT RUN. G10-05 #223 (PR #224) zachowuje tekst kandydata dla zatwierdzonego odczytania, gdy współistnieje niezatwierdzony sens roli; status pozostaje `incomplete`, a konkurencyjny sens trafia do evidence. Obserwacje S003, S031 i S076 są w `data/sentences/observed-2026-10-23-g10-05-partial-lexicalization.tsv`. Bindingi nie zmieniły statusu, a akceptacja językowa właściciela pozostaje NOT RUN. G10-06 #225 (PR #226) umożliwia jawny podgląd zachowanego tekstu `incomplete` przez `--preview-id` bez zmiany statusu; szczegóły są w `data/sentences/observed-2026-10-23-g10-06-explicit-preview.tsv`. G10-07 #227 obserwuje istniejące 100 par w obu kierunkach; wynik preview zapisano w `data/sentences/observed-2026-10-23-g10-07-runtime-preview.tsv`. Przegląd wskazał niedookreślony rodzajnik w PL→EN `receiving`; G10-08 #228 dodaje warianty `a` i `the` dla policzalnego obiektu, zachowując wariant bez rodzajnika dla jawnych rzeczowników masowych. Kandydaci zachowują statusy lokalne, a status wyniku zbiorczego ujawnia alternatywy. Szczegóły CLI są w `data/sentences/observed-2026-10-23-g10-08-indefinite-receiving.tsv`; lexical bindingi pozostają bez zmian, a akceptacja językowa właściciela **NOT RUN**. G10-09 #231 porządkuje pozostałe opisy liczników leksykalnych i konstrukcyjnych po #230. G10-10 #233 (PR #234) dodaje jawne bindingi konstrukcyjne EN dla alternatyw `a/the` w podmiocie i obiekcie/celu na obecnych trasach; `a` wymaga właściwości leksykalnej, a warianty nie zmieniają statusów lexical bindings. Obserwacje wszystkich 100 istniejących par zapisano w `data/sentences/observed-2026-10-23-g10-10-determiner-alternatives.tsv`; PL→EN pozostaje `ambiguous`, a akceptacja właściciela **NOT RUN**. G10-11 #235 (PR #236) dodaje regularną klasę `en_noun_regular` i article property `a` do istniejącego `en.household.n`; w 10 wybranych rekordach usunięto 10 diagnostyk `PRINCIPAL_PARTS_UNAVAILABLE`, zachowując `ambiguous`/`incomplete` oraz nierozstrzygnięte odczyty. Szczegóły: `data/sentences/observed-2026-10-23-g10-11-household-morphology.tsv`; akceptacja właściciela **NOT RUN**. Przymiotników i zaimków EN nie ma jeszcze jako leksemów katalogu.

G10-12 #237 (PR #238) dodało źródłowo oceniony binding `en.offspring.child → concept:child_of_person`, jawny zero plural `offspring` i allomorf `an`. Dla 20 istniejących rekordów EN→PL zwróciło 15 `ok` i 5 `incomplete`, a PL→EN 20 `ambiguous`; oddzielna próba pochodna `An offspring sees a dog.` zwróciła `ok`. Dokładne wyjścia i SHA zapisano w `data/sentences/observed-2026-10-23-g10-12-offspring-a-an.tsv`.

G10-13 #239 (PR #240) dodało analizę i realizację trzeciej osoby liczby mnogiej dla `seeing`, `hearing` i `looking` oraz twarde uzgodnienie podmiotu z czasownikiem w EN i PL. EN `The offspring sees a dog.` → PL `Dziecko widzi psa.` i `The offspring see a dog.` → `Dzieci widzą psa.` zwracają `ok`; EN `The children sees...` i PL `Dzieci widzi...` nie uzyskują pełnego parsowania. PL→EN realizuje mnogie czasowniki w widocznych kandydatach, ale pozostaje `ambiguous` z powodu alternatyw leksemu/rodzajnika. Uruchomione 40 wywołań dla tych samych istniejących rekordów zachowało statusy i zbiory tekstów względem G10-12. Wszystkie 10 odrębnie oznaczonych prób mnogich i SHA `c61f401d1abfc1247602b7a0b76b14b5180973df` zapisano w `data/sentences/observed-2026-10-23-g10-13-plural-agreement.tsv`. Formy `see/hear/look` oraz `widzą/słyszą/patrzą` generują się; angielskie `got` jest w rodzinie bez uzgodnienia liczby. Odbiór językowy właściciela pozostaje **NOT RUN**.
G10-14 #241 (PR #242) rejestruje 200 wywołań rzeczywistego CLI dla S001–S100 PL↔EN na implementacji `33b8830cae5b285fbfafacdbb0e90c85c7916aa6`. EN→PL zwróciło 60 `ok` i 40 `incomplete`; PL→EN zachowało 100 wyników `ambiguous` z widocznymi kandydaturami. Niekompletne statusy ujawniają konkurencyjne niezatwierdzone sensy i założenia rodzajnika. Wszystkie wyniki, kandydatury oraz diagnostyki zapisano w `data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv`; odbiór właściciela pozostaje **NOT RUN**.

G10-15 #243 rozszerza polski `receiving` w czasie przeszłym na 3. osobę liczby mnogiej oraz dobiera `dostali/dostały` z jawnego numeru podmiotu i klasy jego rodzaju. `The teachers got water.` → `Nauczyciele dostali wodę.` i `The children got water.` → `Dzieci dostały wodę.` zwracają `ok`. Odwrócone wejścia zachowują odpowiednio `The teachers got water.` (`incomplete`) i `The children got water.` (`ambiguous`) wśród kandydatów, zgodnie z istniejącymi założeniami określoności. `forms` generuje i analizuje oba czasowniki. Ponowny przebieg S001–S100 wykonał 200 wywołań: EN→PL 60 `ok`/40 `incomplete`, PL→EN 100 `ambiguous`; statusy i zbiory tekstów kandydatów nie zmieniły się wobec G10-14. Dokładne smoke CLI: [tabela G10-15](../../data/sentences/observed-2026-10-23-g10-15-plural-past.tsv). Agent sprawdził formy i te przykłady; akceptacja właściciela pozostaje **NOT RUN**.

G10-16 #245 dodaje osobne rodziny konstrukcji EN/PL i obsługę tożsamościowego operatora negacji na trasie `seeing` w 3. osobie liczby pojedynczej. `The teacher does not see the dog.` → `Nauczyciel nie widzi psa.` zwraca `ok`; wejście PL→EN zachowuje `The teacher does not see the dog.` jako kandydata `incomplete` z jawnymi założeniami rodzajnika. Dodatnie counterparts nie wybierają rodziny negatywnej. Ponowny przebieg S001–S100 wykonał 200 wywołań i zachował statusy oraz zbiory tekstów kandydatów G10-14. Szczegóły są w [tabeli G10-16](../../data/sentences/observed-2026-10-23-g10-16-negated-seeing.tsv); kod mechanizmu znajduje się w commicie `08dd82c87e201161e47d8e277048fe6848e13f73`. Ocena agenta nie stanowi odbioru językowego właściciela, który pozostaje **NOT RUN**.

G10-17 #247 (PR #248) rozszerza negację present na `seeing`, `hearing` i `looking` w liczbie pojedynczej i mnogiej. EN używa zgodnego `does/do not` i bazowej formy czasownika; PL zachowuje jawne formy `nie widzi/widzą`, `nie słyszy/słyszą` i `nie patrzy/patrzą`. `look at` oraz `patrzeć na` pozostają osobną rodziną oblique, a zgodność liczby jest wymagana. Smoke CLI obejmuje oba kierunki, zamianę ról, błędną zgodność oraz negację w nieobsługiwanym czasie przeszłym; pełny przebieg S001–S100 zachował statusy i zbiory tekstów G10-14. Szczegóły są w [tabeli G10-17](../../data/sentences/observed-2026-10-23-g10-17-present-negation.tsv). Ocena agenta: negacja i role są zachowane na zadeklarowanych trasach; PL→EN pozostaje `incomplete` z powodu angielskiej określoności. Akceptacja językowa właściciela pozostaje **NOT RUN**.

G10-18 #249 dodaje jawne warianty `a`, `an` i `the` do ujemnych konstrukcji EN dla `see`, `hear` i `look at`, zachowując właściwości allomorficzne i ograniczenia liczby. Cztery wejścia EN→PL z określonymi rodzajnikami zwróciły `ok`; `A teachers...` oraz błędne `an dog` pozostały `unsupported`. Dwa polskie wejścia zachowały po cztery alternatywy angielskie i status `ambiguous`. Ponowny przebieg S001–S100 wykonał 200 wywołań i zachował wszystkie statusy oraz zbiory `(candidate outcome, text)` G10-14. Szczegóły są w [obserwacji G10-18](../../data/sentences/observed-2026-10-23-g10-18-negative-article-alternatives.tsv). Ocena agenta: warianty rodzajnika i kontrola allomorfów są zgodne z zadeklarowanymi trasami; akceptacja właściciela pozostaje **NOT RUN**.

# DSE-08-03: niezależne specyfikacje wycinków PL↔IR↔EN

Status: wymagane sześć ścieżek dla każdego z trzech wycinków nie jest domknięte. WH źródła i round-trip powierzchni są teraz wykonywalne, ale niezależna realizacja i tłumaczenia zwracają `incomplete`; relacja i scope pozostają nieobsłużone w produkcyjnych ścieżkach językowych. Cały #398 jest **PARTIAL**, a akceptacja językowa właściciela **NOT EVALUATED**. To nie są goldy ani twierdzenia o zatwierdzonej jakości językowej. Oczekiwania IR poniżej wynikają z formuł i ról, nie z outputu parsera lub generatora.

## Wspólny kontrakt wykonania

Dla każdego wycinka wymagane są niezależnie: PL→IR, EN→IR, realizacja niezależnego IR→PL, IR→EN, PL→IR→EN→IR i EN→IR→PL→IR. Niezależny IR nie może zawierać Source Evidence ani być skopiowany z analizy. Porównanie ma sprawdzać typy, role, identyfikację referentów, bindery i zakres; sam zgodny round-trip nie potwierdza poprawności.

Jeżeli wejście ma więcej niż jeden dopuszczalny odczyt, zapisujemy wszystkie kandydatury albo odmawiamy rozstrzygnięcia. Nie obniżamy oczekiwań do obecnego wsparcia runtime. Każda nowa powierzchnia lub mapowanie leksykalne zaproponowane tutaj jest propozycją agenta, nie zatwierdzeniem właściciela.

## 1. Dwukrokowa kompozycja relacji

### Niezależny cel formalny

Wejściowa deklaracja to dowolna binarna relacja `r(a: Entity, b: Entity)`. GoalSpec ma porty `x: Entity`, `y: Entity`, typ wyniku `Entity → Entity → Prop` i dokładnie dwa wystąpienia tej samej relacji, połączone przez wspólną zmienną `z`:

```text
λx:Entity. λy:Entity. ∃z:Entity. And(r(a=x,b=z), r(a=z,b=y))
```

Oczekiwany obiekt jest anonimowym wyrażeniem ścieżki `r∘r`; nie ma osobnego ConceptID, aliasu z zachowaniem algorytmu ani aksjomatu przechodniości. Zamiana portów `a` i `b` w którymkolwiek kroku albo rozdzielenie `z` musi zmienić lub unieważnić wynik. Istniejący neutralny przykład `data/semantic/examples/goals/two-hop-property.json` zawiera tę samą ścieżkę oraz dodatkową restrykcję `p(y)`; należy użyć osobnego, mniejszego GoalSpec bez tej restrykcji.

### Dane i dopuszczone językowe propozycje

- Neutralne metadane `r` i formalny GoalSpec są w `data/semantic/examples/goals/metadata.json` oraz `two-hop-property.json`.
- `pilot.r` ma dwa uporządkowane porty Entity. W angielskim pakiecie źródłowym istnieje `sense:en.link.indirect`, ale jego opis jawnie nazywa je techniczną konwencją nazewniczą. To nie jest zatwierdzone znaczenie angielskiego *link* ani dowód bilingwalności. Odpowiednika PL nie ma w bieżącym pakiecie WH.
- Najbliższy istniejący zestaw common lexemes to teacher/dog/see w `data/semantic/examples/source-composition/wh-questions/packages/{en,pl}.json`. Propozycja wejść do oceny: EN “The teacher sees the dog, and the dog sees the teacher.”; PL “Nauczyciel widzi psa, a pies widzi nauczyciela.” Są to dwa jawne fakty z tym samym pośrednim referentem, nie zatwierdzona para przekładowa. Kandydacka instancja drogi to `∃e1,e2,z (see(e1,A,B) AND see(e2,B,A))`, gdzie A jest nauczycielem, B psem; po domknięciu referentu wydarzenia relacja Entity–Entity `R(a,b) := ∃e see(e,a,b)` daje anonimową ścieżkę `R∘R(A,A)`. Sprawdzenie ma potwierdzić to przejście jawnie z danych/GoalSpec; nie zakładamy, że sama klauzulowa koniunkcja jest równoważna celowi lambda-ścieżki. Bez takiego mapowania przypadek pozostaje nierozstrzygnięty, nie PASS.

### Kryteria kontroli

Niezależny dokument IR musi przechodzić obie realizacje bez tekstu źródłowego, a PL i EN muszą zachować oba zastosowania `r` i wspólny binder. Każdy z sześciu kierunków ma zachować te same porty i role. Osobno sprawdzamy kontrolę negatywną: pojedyncze `r(x,y)`, odwrócone argumenty i dwie krawędzie bez wspólnego referentu nie mogą zostać przedstawione jako ten sam wynik. Nie wyciągamy wniosku o prawdziwości ścieżki bez jawnych faktów.

Ograniczenie przed wykonaniem: obecny EN pilot `link.indirect` jest tylko bindingiem GoalSpec; PL go nie ma. Wybór poprawnej naturalnej powierzchni dla anonimowej kompozycji i jej mapowania pozostaje otwarty. Brak danych/mappingu nie może być zastąpiony kodem wybieranym po nazwie konceptu.

## 2. Pytanie WH o wskazaną rolę

### Niezależny cel formalny

Używamy dwumiejscowej relacji `see(experiencer, stimulus)` oraz oddzielnie dostarczonych referentów `A` i `B`. Wymagane są dwa różne pytania:

```text
Ask answers=[x:Entity] body=see(experiencer=x, stimulus=B)  # pytana rola experiencer
Ask answers=[y:Entity] body=see(experiencer=A, stimulus=y)  # pytana rola stimulus
```

Zamiana ról ma zamienić answer variable i pozostawioną stałą referencję. Na jawnych faktach `see(A,B)` evaluator może zwrócić odpowiednio `A` lub `B` z dowodem. Pytanie nie jest asercją i nie dopisuje `see(A,B)` do faktów. Brak pasującego faktu pozostaje `unknown`, nie `false`.

### Dane i ograniczenia odczytu

Istniejące źródłowe propozycje to `data/semantic/examples/source-composition/wh-questions/cases/{en,pl}-{subject,object}.json` oraz README obok nich. Przykłady `Who sees the dog?` / `Kto widzi psa?` wiążą lukę z `experiencer`; `The teacher sees whom?` / `Kogo widzi nauczyciel?` wiążą ją z `stimulus`. Pierwsza para kontekstu używa jawnego drugiego argumentu; druga używa jawnego pierwszego argumentu. Pochodzenie wpisów jest agentowe, nie właścicielskie.

Oczekiwany zakres jest ograniczony do dokładnie wskazanej luki i powiązanej z nią roli. W obecnych przykładach answer variable ma typ `Entity`; brak restryktora `Person`, więc te dane nie uzasadniają pełnej semantyki *who/kto*. Angielski object przykład jest pytaniem echo. Nie traktujemy go jako dowodu obsługi neutralnego szyku WH ani nie zamieniamy cicho na inną konstrukcję.

### Kryteria kontroli

Sześć wymaganych kierunków musi zachować `Question` oraz odpowiednią rolę luki. IR→PL i IR→EN startują z niezależnego `Ask`, bez Source Evidence. Dla każdego języka sprawdzamy zamianę `experiencer`↔`stimulus` i odpowiednią zmianę answer variable. Dodatkowo wykonujemy `algebra evaluate` na niezależnie podanych dodatnich faktach; zapisujemy proof/witness i kontrolę z pustymi faktami jako `unknown`. Evaluator nie może użyć pytania jako faktu.

Aktualna, ograniczona obserwacja CLI: skonfigurowane źródłowe analizy EN/PL tworzą cztery kompletne `Ask`; niezależna realizacja i tłumaczenie zwracają tekst, ale pozostają `incomplete` z jawnymi resztami kontekstu lub założeniami. Przy kanonicznych dokumentach z deklaracjami sygnatur evaluator zwraca odpowiedzi dla czterech wariantów z jawnym faktem i instantem. Puste lub wygasłe fakty dają `unknown`; brak instanta i `past` są jawnie odrzucane. Source-analysis API nie przyjmuje faktów. To częściowe działanie WH, nie spełnienie sześciu ścieżek.

## 3. Kontrast zakresów negacji i kwantyfikatora

### Niezależne cele formalne

Używamy tych samych deklaracji `Teacher(x)`, `Dog(y)` i `see(experiencer, stimulus)`, w tym samym kontekście referencji. Dwa cele są różnymi drzewami typowanymi:

```text
Not(Quantifier x:Entity all
  restrictor=Teacher(x)
  body=Exists y:Entity (Dog(y) AND see(x,y)))

Quantifier x:Entity all
  restrictor=Teacher(x)
  body=Not(Exists y:Entity (Dog(y) AND see(x,y))))
```

Pierwszy mówi, że nie każdy nauczyciel widzi jakiegoś psa (dla proponowanej analizy `a dog` pozostaje pod zakresem `every`). Drugi mówi, że dla każdego nauczyciela nie istnieje widziany przez niego pies. Nie wolno utożsamić drzew, przepchnąć negacji przez `all`, ani uznać drugiej za pierwszą na podstawie samej powierzchni. To są cele reprezentacji i realizacji; evaluator pozytywnych faktów nie wykonuje `Not`/kwantyfikatorów, więc prawdziwość tych odczytań nie jest kryterium tego wycinka.

### Powierzchnie proponowane, nie zatwierdzone

EN propozycje: “Not every teacher sees a dog.” dla pierwszego drzewa oraz “Every teacher sees no dog.” dla drugiego. PL propozycje: “Nie każdy nauczyciel widzi jakiegoś psa.” oraz “Każdy nauczyciel nie widzi żadnego psa.” Są to propozycje agenta do niezależnej oceny językowej. Przegląd 2026-10-27 nie uznał ich za jednoznaczne: w EN istotna jest interakcja kwantyfikatora uniwersalnego z negacją i indefinitem, a w PL szyk `wszyscy/każdy ... nie` może być odczytany jak `nie wszyscy`. Źródła językoznawcze uzasadniają ostrożność, ale nie zatwierdzają tych konkretnych zdań ani przekładu: [Wu & Ionin, “English inverse scope”](https://doi.org/10.1177/0267658321992830), [Przybyszewski, “O użyciu operatorów kwantyfikacji ogólnej w zdaniach przeczących”](https://www.jezyk-polski.pl/index.php/jp/issue/download/17/19).

Jako jawniejsze **niezweryfikowane propozycje** rozważamy “There is at least one teacher who sees no dogs.” / „Istnieje co najmniej jeden nauczyciel, który nie widzi żadnego psa.” dla `Not(All(...))` oraz “For every teacher, there is no dog that teacher sees.” / „Dla każdego nauczyciela nie istnieje pies, którego ten nauczyciel widzi.” dla `All(Not(...))`. W klasycznej logice pierwsza para ma odpowiadać `∃x(Teacher(x) ∧ ¬∃y(Dog(y) ∧ see(x,y)))`, czyli podanej formule `Not(All(...))`; druga jawnie stawia kwantyfikator każdego nauczyciela na zewnątrz negowanego istnienia psa. To uzasadnienie formalne nie rozstrzyga wszystkich odczytań naturalnych ani nie jest akceptacją właściciela. Przed użyciem jako wymagań CLI wymagane są niezależna ocena językowa i próby zachowujące alternatywy.

### Kryteria kontroli

Każdy z sześciu kierunków ma zachować exact binder `x`, restryktor `Teacher(x)`, existential `y` i pozycję `Not`. Zamiana zakresu musi dać różne, sprawdzone dokumenty IR lub jawną niejednoznaczność/odmowę; brak mappingu nie może produkować dodatniego `see` ani gubić `Not`. Realizacja niezależnych drzew nie może użyć tekstu źródłowego. Nie wykonujemy truth evaluation dla tych drzew.

Istniejące neutralne przepisy `data/semantic/examples/scopes/build/not-all.json` i `build/all-not.json` oraz odpowiadające dokumenty `query/` pokazują struktury `Not(All(...))` i `All(Not(...))`, lecz używają nieinterpretowanych symboli, nie leksyki PL/EN. Obecne DSE plany nie mają mapowania kwantyfikatorów; zwykłe source plany mapują jedynie operator negacji. To granica danych/konstrukcji wymagająca osobnej obsługi, nie podstawa do zmiany oczekiwanych drzew.

## Stan dowodów

| Wycinek | Niezależne wymaganie | CLI PL→IR / EN→IR | IR→PL / IR→EN | Oba round-tripy | Dodatkowy evaluator | Akceptacja właściciela |
| --- | --- | --- | --- | --- | --- | --- |
| Dwukrokowa relacja | GoalSpec path `r∘r`; agent-review proposal definiuje `r(a,b)` jako istnienie eventu widzenia, owner acceptance NOT EVALUATED | PARTIAL: `algebra synthesize` tworzy dwa `pilot.r` z binderem; CLI `algebra compare` dowodzi równoważności z niezależnym dwueventowym wyrażeniem przy `alias_v1`. Syntax probe: proste zdania EN/PL mają po jednym root `transitive`, koordynowane zdania `forest_outcome=unknown`, zero roots; token `and` / `a` nie ma licencjonowanego odczytu. `algebra analyze` pokazuje dalsze `UNSUPPORTED_SOURCE_FOREST` | PARTIAL: produkcyjna realizacja EN i PL zwraca `NO_EXACT_REALIZATION_BINDING`; formalny alias nie jest ładowany przez językowe ścieżki | NOT COMPLETE: oba source→target kierunki odrzucone przed round-trip | N/A | NOT EVALUATED |
| WH role | Dwa różne root `Ask`; rola zmienia odpowiedź | PARTIAL: 4 źródłowe analizy i 4 analizy przetłumaczonych powierzchni zwracają `ok`; wszystkie 8 porównań `proved_equivalent` w zakresie wyrażenia | PARTIAL: cztery niezależne realizacje zwracają tekst, ale `incomplete` z powodu jawnych reszt kontekstu; cztery tłumaczenia zwracają tekst, ale `incomplete` z powodu reszt lub założeń | PARTIAL: oba round-tripy mają tekst i ponowną analizę `ok`; porównanie celowo pomija request context, więc nie dowodzi pełnej równoważności kontekstu | PARTIAL: cztery odpowiedzi na jawnych faktach; empty/expired → `unknown`; brak instantu/past → odmowa | NOT EVALUATED |
| Scope | Dwa różne drzewa `Not(All(...))` / `All(...Not(...))` | PARTIAL: oba niezależne IR przechodzą `algebra check`; porównanie zachowuje różne normal forms (`undetermined`). Syntax probe: zero roots; `every` / `każdy` oraz All-Not `no` / `żadnego` nie mają licencjonowanego odczytu, a token początkowy nie ma start construction. `algebra analyze` pokazuje dalsze `UNSUPPORTED_SOURCE_FOREST` | PARTIAL: niezależna realizacja obu scope w EN i PL zwraca `NO_EXACT_REALIZATION_BINDING` | NOT COMPLETE: source→target zatrzymuje się na analizie | nie dotyczy; brak ewaluacji zakresu | NOT EVALUATED |

## Bieżące próby CLI

Najnowszy przebieg z bazowego SHA `3e9b7dcdbd389c54b65c02d8877a05dde2553b78` i aktywnymi zmianami #398 zwrócił `ok` dla czterech skonfigurowanych analiz WH. Każda z czterech niezależnych realizacji tworzy tekst, lecz całość ma status `incomplete`, bo zachowuje niezużyte referenty i warunki kontekstu. Cztery tłumaczenia również tworzą tekst i kończą jako `incomplete`: realizacje polskie raportują reszty, a angielskie jawne założenia gramatyczne, w tym osobę i określoność. Powierzchnie z obu kierunków tłumaczenia ponownie analizują się jako `ok`.

Wszystkie osiem porównań WH zwróciło `proved_equivalent` tylko w profilu wyrażenia. Pliki porównawcze jawnie wykluczają `request_context`; wcześniejsze pełne porównanie zgłosiło `CONTEXT_REFERENCE_MISSING`, ponieważ identyfikatory/referenty kontekstu wejściowego nie odpowiadają niezależnym dokumentom. Wynik porównania nie dowodzi równoważności kontekstu ani nie usuwa założeń. Negatywna kontrola `Whom does the teachers see?` jest odrzucona jako `UNSUPPORTED_SOURCE_FOREST`; poprawna, pojedyncza liczba `Whom does the teacher see?` jest analizowana jako pytanie o rolę `stimulus`.

Zachowane próby evaluator-a zwracają A/B dla czterech EN/PL pytań przy jawnym syntetycznym fakcie; pusty i wygasły facts set pozostają `unknown`, brak instantu i `past` są odrzucane. To wąski obieg odpowiedzi, nie dowód na pozostałe ścieżki językowe.

Najnowsze przebiegi produkcyjnego EN/PL dla relacji i scope zapisane w `current/slices/current-run-commands.jsonl` zwracają `UNSUPPORTED_SOURCE_FOREST` na analizie i tłumaczeniu. Dokładniejsza izolacja `SourcePipeline.AnalyzeSyntax` jest zapisana w `current/slices/outputs/source-syntax-diagnostics-{en,pl}.stdout.jsonl`: oba pojedyncze zdania kontrolne mają po jednym root konstrukcji przechodniej, lecz zdania złożone oraz scope mają `forest_outcome=unknown` i zero roots. Przyczyną jest brak licencjonowanego odczytu `and` / `a` w koordynacji oraz brak odczytów operatorów/kwantyfikatorów `every`, `no`, `każdy`, `żadnego` i odpowiadających im start construction. Stąd `UNSUPPORTED_SOURCE_FOREST` jest komunikatem o odrzuconym wyniku analizy; composition nie dostaje poprawnego lasu, który mógłby zgubić. Kod sondy jest w `current/slices/probes/source-syntax-probe.go.txt`; uruchomiono ją w odłączonym worktree na commitcie `995436e98f9d05c99275e2c812bf9b63c98b63d9`. Dla scope próbowano obu wcześniejszych par powierzchni (`Not every...` / `Nie każdy...` oraz `Every...no...` / `Każdy...żadnego...`); niezależna realizacja nadal zwraca `NO_EXACT_REALIZATION_BINDING`.

Osobny formalny proposal relacji zapisany w `data/semantic/acceptance/dse-08-03/relation/vision-event-binary-proposal.json` definiuje `pilot.r(a,b)` przez istnienie eventu `pilot.vision_event`. Przy wybranym profilu zawierającym `alias_v1`, aktualny `algebra compare` zwraca `proved_equivalent` dla anonimowej ścieżki `r∘r` i niezależnego wyrażenia dwóch eventów. To dowód względem wybranego aliasu i profilu, nie produkcyjna analiza/realizacja PL/EN ani owner acceptance. Komenda i surowy wynik znajdują się w `current/slices/current-run-commands.jsonl` oraz `slices/outputs/compare-two-hop-see-alias-proposal.stdout.json`.

Dwa niezależne dokumenty scope przechodzą `algebra check`. `algebra compare` zachowuje `Not(All(...))` i `All(...Not(...))` jako różne normal forms, ale zwraca `undetermined`; nie dowodzi to nierównoważności ani prawdziwości odczytań.

Niezależne formalne propozycje, w tym GoalSpec, znajdują się w `data/semantic/acceptance/dse-08-03/`. Aktualne komendy i stdout/stderr są w `docs/work/dse/evidence/08-03/current/slices/current-run-commands.jsonl` oraz `slices/outputs/`; pozostałe odtwarzalne próby evaluator-a i historycznej rozbieżności są w `slices/commands.jsonl` i `replay/commands.jsonl`.

Występuje rozbieżność w historycznym logu bazowym: zarejestrowany wynik temporalnej odmowy nie odtwarza się na zachowanym binarium i tych samych projected-inputs. Projected docs nie mają wymaganych deklaracji sygnatur i bieżące powtórzenie zwraca `UNKNOWN_SYMBOL`. Prawidłowe próby evaluator-a używają kanonicznych `Ask` z sygnaturami, zapisanych w `current/inputs/`. Historycznego logu nie traktujemy jako potwierdzonego wyniku; szczegóły i surowe wyjścia są w `docs/work/dse/evidence/08-03/current/README.md`.

## Źródła i granice diagnozy

- Aktywne kryteria: [issue #398](https://github.com/clickbytes/golingua/issues/398).
- WH propozycje i granice semantyczne: `data/semantic/examples/source-composition/wh-questions/README.md`, `cases/`, `contexts/`.
- Neutralne ścieżki, metadata i GoalSpec: `data/semantic/examples/goals/two-hop-property.json`, `data/semantic/examples/goals/metadata.json`.
- Neutralny WH evaluator i jawne fakty: `data/semantic/examples/scopes/query/ask-first.json`, `data/semantic/examples/scopes/facts-simple.json`, `docs/architecture/semantic-algebra/scopes.md`.
- Neutralne kontrasty zakresu: `data/semantic/examples/scopes/build/not-all.json`, `build/all-not.json`, `query/not-all.json`, `query/all-not.json`.
- Odczyt granic w implementacji (nie wynik CLI): `internal/engine/algebra_realization.go`, `internal/engine/dse_translate.go`, `internal/runtime/compiled_plan_realizer.go`, `internal/runtime/source_dse.go`, `internal/cli/algebra.go`.

# DSE-08-03: niezależne specyfikacje wycinków PL↔IR↔EN

Status: specyfikacje są propozycją techniczną; pełny wymagany przebieg sześciu ścieżek dla #398 jest **NOT RUN**. Wykonano diagnostykę bazową oraz ograniczone próby aktualnego CLI opisane w `docs/work/dse/evidence/08-03/`; akceptacja językowa właściciela **NOT EVALUATED**. Nie są to goldy ani twierdzenia o zatwierdzonej jakości językowej. Oczekiwania IR poniżej wynikają z formuł i ról, nie z outputu parsera lub generatora.

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

Aktualna, ograniczona obserwacja CLI: skonfigurowane źródłowe analizy EN/PL tworzą cztery kompletne `Ask`; przy kanonicznych dokumentach z deklaracjami sygnatur evaluator zwraca odpowiedzi dla czterech wariantów z jawnym faktem i instantem. Puste lub wygasłe fakty dają `unknown`; brak instanta i `past` są jawnie odrzucane. Niezależna realizacja tych `Ask` nadal zwraca `UNSUPPORTED_EXPRESSION_PART`. Source-analysis API nie przyjmuje faktów. To częściowe działanie evaluator-a, nie spełnienie sześciu ścieżek.

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

EN propozycje: “Not every teacher sees a dog.” dla pierwszego drzewa oraz “Every teacher sees no dog.” dla drugiego. PL propozycje: “Nie każdy nauczyciel widzi jakiegoś psa.” oraz “Każdy nauczyciel nie widzi żadnego psa.” Są to propozycje agenta do niezależnej oceny językowej. W szczególności zakres *a/no dog*, negatywna zgodność w polskim, rodzaj określoności i możliwe alternatywne odczytania muszą być jawne; nie przypisujemy im właścicielskiej akceptacji. Jeśli zdanie ma więcej niż jeden dopuszczalny zakres, zachowujemy alternatywy albo odmawiamy pojedynczego odczytania.

### Kryteria kontroli

Każdy z sześciu kierunków ma zachować exact binder `x`, restryktor `Teacher(x)`, existential `y` i pozycję `Not`. Zamiana zakresu musi dać różne, sprawdzone dokumenty IR lub jawną niejednoznaczność/odmowę; brak mappingu nie może produkować dodatniego `see` ani gubić `Not`. Realizacja niezależnych drzew nie może użyć tekstu źródłowego. Nie wykonujemy truth evaluation dla tych drzew.

Istniejące neutralne przepisy `data/semantic/examples/scopes/build/not-all.json` i `build/all-not.json` oraz odpowiadające dokumenty `query/` pokazują struktury `Not(All(...))` i `All(Not(...))`, lecz używają nieinterpretowanych symboli, nie leksyki PL/EN. Obecne DSE plany nie mają mapowania kwantyfikatorów; zwykłe source plany mapują jedynie operator negacji. To granica danych/konstrukcji wymagająca osobnej obsługi, nie podstawa do zmiany oczekiwanych drzew.

## Stan dowodów

| Wycinek | Niezależne wymaganie | CLI PL→IR / EN→IR | IR→PL / IR→EN | Oba round-tripy | Dodatkowy evaluator | Akceptacja właściciela |
| --- | --- | --- | --- | --- | --- | --- |
| Dwukrokowa relacja | GoalSpec path `r∘r`; językowe mapowanie pozostaje do ustalenia | PARTIAL: `algebra synthesize` buduje dwa `pilot.r` z jednym binderem i `algebra compare` dowodzi zgodności z niezależnym IR; źródłowe EN/PL odrzucone jako `UNSUPPORTED_SOURCE_FOREST` | NOT ACCEPTED: niezależna realizacja EN i PL odrzucona (`UNSUPPORTED_EXPRESSION_PART`) | NOT RUN | N/A | NOT EVALUATED |
| WH role | Dwa różne root `Ask`; rola zmienia odpowiedź | PARTIAL: 4 skonfigurowane `algebra analyze` zwróciły kompletne `Ask`; każde porównanie z niezależnym EN/PL IR dało `proved_equivalent` | NOT ACCEPTED: niezależna realizacja EN i PL odrzucona (`UNSUPPORTED_EXPRESSION_PART`) | NOT RUN | PARTIAL: cztery odpowiedzi na jawnych faktach; empty/expired → `unknown`; brak instantu/past → odmowa | NOT EVALUATED |
| Scope | Dwa różne drzewa `Not(All(...))` / `All(...Not(...))` | PARTIAL: oba niezależne IR przechodzą `algebra check`; porównanie zachowuje różne normal forms (`undetermined`); źródłowe EN/PL odrzucone jako `UNSUPPORTED_SOURCE_FOREST` | NOT ACCEPTED: niezależna realizacja obu scope w EN i PL odrzucona (`UNSUPPORTED_EXPRESSION_PART`) | NOT RUN | nie dotyczy; brak ewaluacji zakresu | NOT EVALUATED |

## Wykonane próby bazowe

Na bazowym SHA `dab9d23b48fe429ca4221c5efdf8e081bc6405b3` cztery `algebra analyze` zwróciły kompletne WH `Ask`; bieżące powtórzenie dało identyczne raw outputy. Każdy wynik porównany z niezależnym Meaning przez `algebra compare` dał `proved_equivalent`. Osobny role-swap zachował odmienne normal forms; comparator odpowiedział `undetermined`, a nie dowiódł nierównoważności. W niezależnych faktach cztery kontrolowane EN/PL subject/object zapytania zwróciły odpowiednio A/B z proof; pusty i wygasły facts set pozostały `unknown`; brak instantu oraz `past` zostały odrzucone. To domyka tylko wąski WH evaluator.

Niezależny GoalSpec `r∘r` z `pilot.r` utworzył przez `algebra synthesize` kanoniczny wyraz dwóch relacji ze wspólnym binderem; niezależny IR przeszedł `algebra check`, a `algebra compare` dowiodło ich równoważności w wybranym profilu. Nie powstała realizacja tekstowa: źródłowe dwuklauzulowe wejścia EN i PL zwracają `UNSUPPORTED_SOURCE_FOREST`, a niezależna realizacja zwraca `UNSUPPORTED_EXPRESSION_PART`.

Dwa niezależne dokumenty scope przechodzą `algebra check`. `algebra compare` zachowuje `Not(All(...))` i `All(...Not(...))` jako różne normal forms, ale wynik to `undetermined`; nie jest to dowód nierównoważności ani prawdziwości odczytań. Proponowane wejścia scope w EN i PL zwracają `UNSUPPORTED_SOURCE_FOREST`; realizacja obu niezależnych struktur w obu językach zwraca `UNSUPPORTED_EXPRESSION_PART`.

Niezależne formalne propozycje, w tym GoalSpec, znajdują się w `data/semantic/acceptance/dse-08-03/`. Pełne komendy, stdout/stderr, exit codes, porównania, syntetyczne fakty i wszystkie odmowy CLI są w `docs/work/dse/evidence/08-03/current/`.

Występuje rozbieżność w historycznym logu bazowym: zarejestrowany wynik temporalnej odmowy nie odtwarza się na zachowanym binarium i tych samych projected-inputs. Projected docs nie mają wymaganych deklaracji sygnatur i bieżące powtórzenie zwraca `UNKNOWN_SYMBOL`. Prawidłowe próby evaluator-a używają kanonicznych `Ask` z sygnaturami, zapisanych w `current/inputs/`. Historycznego logu nie traktujemy jako potwierdzonego wyniku; szczegóły i surowe wyjścia są w `docs/work/dse/evidence/08-03/current/README.md`.

## Źródła i granice diagnozy

- Aktywne kryteria: [issue #398](https://github.com/clickbytes/golingua/issues/398).
- WH propozycje i granice semantyczne: `data/semantic/examples/source-composition/wh-questions/README.md`, `cases/`, `contexts/`.
- Neutralne ścieżki, metadata i GoalSpec: `data/semantic/examples/goals/two-hop-property.json`, `data/semantic/examples/goals/metadata.json`.
- Neutralny WH evaluator i jawne fakty: `data/semantic/examples/scopes/query/ask-first.json`, `data/semantic/examples/scopes/facts-simple.json`, `docs/architecture/semantic-algebra/scopes.md`.
- Neutralne kontrasty zakresu: `data/semantic/examples/scopes/build/not-all.json`, `build/all-not.json`, `query/not-all.json`, `query/all-not.json`.
- Odczyt granic w implementacji (nie wynik CLI): `internal/engine/algebra_realization.go`, `internal/engine/dse_translate.go`, `internal/runtime/compiled_plan_realizer.go`, `internal/runtime/source_dse.go`, `internal/cli/algebra.go`.

# 07 — Proceduralna morfologia, generatory i wyjątki

**Częściowo wdrożone w G02.** Typowane operacje, kompilator paradygmatów, formalny kompilator wyjątków, generator i finite relation analyzer istnieją w `internal/procedure` oraz `internal/morph`. Podstawowa zasada: opisujemy **klasy i reguły**, a nie przechowujemy wszystkich regularnych form pojedynczych słów.

## Osobne komponenty

`MorphLexeme`: lemat, POS, język, cechy własne i przypisana klasa. Nie wymaga zatwierdzonego `CoreRef`, aby generator mógł go odmienić.

`Paradigm`: typowana, lokalna dla języka klasa stabilnych komórek cech i operacji; może dziedziczyć wspólne przepisy. Implementacja Go znajduje się w `internal/morph`. Komórka ma stabilny `CellID` oraz jawny zestaw znanych wartości z `FeatureProfile`. Reguła wskazuje komórkę i deklaratywny plan.

`StemSet / PrincipalParts`: tematy i dodatkowe formy diagnostyczne dla nieregularności (np. `go → went`). `ResolveStem` zwraca wyłącznie formę podaną jawnie dla wymaganego principal part; brak wymaganej formy daje `unknown`. Nie udawaj, że każdą formę można wywnioskować z samego lematu.

`FeatureProfile`: dopuszczone pola i wartości należą do **języka**, a nie do wspólnego CORE. Neutralna unifikacja cech w `internal/features`.

`ProcedurePlan`: wspólny, ograniczony i zamknięty zestaw kroków (`select_stem`, `append`, `prepend`, `strip_suffix`, `replace_suffix`, `alternation`). Nieznane operacje i niezadeklarowane warianty są odrzucane; nie ma `eval`, poleceń powłoki ani kodu z JSON. Interpreter liczy limity wejścia/wyjścia w punktach kodowych oraz bajtach, waliduje guardy cech i zwraca ślad kroków. Znana cecha niedopasowana do guarda pomija krok; brak cechy kończy się `unknown`, a nierozstrzygnięta wartość `incomplete`. Operacje `strip_suffix`, `replace_suffix` i `alternation` wymagają dokładnego dopasowania; brak dopasowania nie jest cichym fallbackiem. Obecna alternacja oznacza jedno literalne zastąpienie pierwszego wystąpienia, nie ogólną obsługę procesów fonologicznych.

Kompilator sprawdza wersję profilu, komórki i reguły, cykle dziedziczenia, konflikty oraz wymagania form diagnostycznych. JSON paradygmatu ma jawne wymagane pola, odrzuca nieznane lub błędnie zapisane klucze oraz ogranicza dokument do 1 MiB. Dziedziczenie zachowuje stabilne sloty; wspólny identyczny przodek w diamencie jest scalany raz, a sprzeczne definicje są błędem. Interpretor pozostaje niezależny od języka, zaś `Paradigm` i jego etykiety należą do języka i wersji profilu.

`ExceptionRule`: tylko potrzebny, recenzowany wyjątek z określonym zakresem i efektem. Dokładna precedencja, kontrola konfliktów i provenance są opisane w [kontrakcie wyjątków](morph-exceptions.md).

## Generator

`Generate(MorphLexeme, PrincipalParts, FeatureBundle, CompiledParadigms, CompiledExceptions, Budget)` wybiera komórkę, jawne principal parts, wykonuje plan i rozstrzyga wyjątki. Zwraca `generated`, `blocked`, `unavailable` albo `unknown`; błędne, niekompletne i przekraczające limit żądania pozostają typowanymi błędami.

Reguły sprawdzają POS, profil, dostępność komórki i zgodność wyjątków. Kiedy forma jest zablokowana, generator **nie może** wrócić do regularnej postaci przez obejście. Gdy jej nie znamy, odpowiedź brzmi `unknown/unavailable`, nie wymyślona forma. Pusta końcówka (zero exponent) nie jest brakującą komórką.

### Angielski finite: osoba, liczba i czas

G10-01 dodaje do angielskiego profilu `verb_form=finite` oraz jawne komórki czasu teraźniejszego dla sześciu kombinacji osoby i liczby. Tylko 3. osoba liczby pojedynczej ma odmienną powierzchnię; pozostałe wybierają formę podstawową. Czas przeszły ma jedną komórkę wspólną, bo jego powierzchnia nie zmienia się przez osobę ani liczbę. Imperative oraz present/past participle pozostają osobnymi formami.

Regularne procedury wybierają wyłącznie przypisane im leksemy: zwykłe `+s/+ed/+ing`, sybilant `+es`, końcowe nieme `e` oraz spółgłoska + `y`. Nie jest to automatyczny klasyfikator ortografii. `see`, `hear`, `look` i `get` mają jawne principal parts i wybierają je bez zgadywania. Angielski adapter zdań analizuje/generuje ograniczone formy finite teraźniejsze 3. osoby lp. i przeszłe; G10-01 nie oznacza obsługi wszystkich konstrukcji zdaniowych.

### Angielski rzeczownik: liczba

G10-03 dodaje jawne komórki singular/plural dla wybranych rzeczowników. Procedury obejmują regularne `+s`, sybilant `+es` i dokładne `y → ies`; przypisanie klasy jest indywidualne i nie klasyfikuje całego słownika. `child/children`, `person/people`, `mouse/mice`, `goose/geese`, `foot/feet`, `tooth/teeth`, `leaf/leaves` i `knife/knives` zapisują oba principal parts. `fish` w sensie zwierzęcia oraz `sheep` mają osobne odczyty liczby pojedynczej i mnogiej o tej samej powierzchni. Nie powstają uogólnione reguły nieregularności; masowe `food`, `money` i `water` pozostają singular-only.

### Polski finite: aspekt, osoba, liczba i uzgodnienie czasu przeszłego

G10-02 przypisuje jawne klasy tylko `widzieć`, `słyszeć`, `patrzeć` i `dostać`. Pierwsze trzy mają sześć komórek czasu teraźniejszego z aspektem niedokonanym. `Dostać` ma sześć komórek przyszłych czasu dokonanego oraz komórki przeszłe dla osób i liczb. Czas przeszły rozróżnia rodzaj w liczbie pojedynczej oraz klasę męskoosobową i niemęskoosobową w liczbie mnogiej. Tematy są zapisane jako principal parts; procedury wybierają końcówkę komórki. Nie jest to klasyfikator wszystkich polskich czasowników ani pełny model aspektu, trybu, strony, imiesłowów czy czasowników złożonych.

### Polskie rzeczowniki i przymiotnik

G10-04 deklaruje pełne 7 przypadków × 2 liczby dla ośmiu jawnie wybranych rzeczowników. `kot` i `ptak` przechodzą z ograniczonej sześciokomórkowej reguły na indywidualne, źródłowane formy. Pozostałe nowe wpisy to `koń`, `miasto`, `mieszkanie`, `rzeka`, `siostra` i `mewa`. Każda reguła wskazuje własny rekord WSJP PAN oraz licencję `unknown`; żaden wpis nie tworzy produktywnej reguły dla innych leksemów. `mewa` jest obecnie wyłącznie leksemem morfologicznym bez sensu i bindingu CORE.

Jedyny obecny polski przymiotnik `dorosły` ma jawne komórki stopnia równego, wyższego i najwyższego dla dostępnych przypadków, liczb i klas zgody. Formy comparativus są wymienione w principal parts; procedura najwyższego stopnia dodaje `naj-` do tych form. WSJP nie przedstawia vocative dla stopni wyższych, więc tych komórek nie deklarujemy.

### Wyjątki

| Effect | Semantyka |
| --- | --- |
| `block` | nie generuj i nie akceptuj danej komórki |
| `replace` | wynik jawny **zamiast** regularnego |
| `addVariant` | świadomie dopuszczona dodatkowa forma |
| `stemOverride` | inny temat w określonych warunkach |
| `suppletion` | osobna nieprzewidywalna forma bez fikcyjnej reguły |
| `constructionException` | rząd/strona/operator specyficzny dla sensu: **Valency/Construction**, nie morphology cell |

Precedence: jawna hierarchia zakresów + deklaracja nadpisania, **nie** kolejność pliku. Dwie sprzeczne reguły bez rozstrzygnięcia → odmowa kompilacji.

## Analizator form

`BuildFiniteRelationTable` waliduje odczyty, zapisuje relację deterministycznie i wiąże ją z SHA-256. `BuildInverseIndex` tworzy indeks powierzchnia→odczyty; `Analyze(surface)` zwraca wszystkie licencjonowane homografy i synkretyczne interpretacje z leksymem, POS, paradygmatem, komórką, cechami, źródłem, provenance i śladem.

Powierzchnia jest normalizowana wyłącznie do Unicode NFC. Wielkość liter i diakrytyki pozostają bez zmian; nie ma case folding ani usuwania akcentów. Odczyt wyjątkowy wymaga zgodnego ID reguły, provenance i śladu źródła. Blokada nie dopuszcza regularnego fallbacku. Tabela oraz zapytania mają jawne limity bajtów, liczby odczytów i długości powierzchni.

Ani FST, ani lazy backend nie są obowiązkowym warunkiem. Nie udowadniamy teraz przewagi wydajności nad starym JavaScriptem.

## Rzeczywiste przykłady użytkownika

Właściciel może podawać konkretne leksemy i żądane formy PL/EN. CLI zwraca pełną listę, dostępne cechy i informację `generated / exception / blocked / unknown`. Na podstawie obserwacji poprawiamy klasę, temat lub wyjątek.

**Nie tworzymy obowiązkowych goldów, prób jednostkowych dla każdej komórki, matrycy odmiany ani porównań z poprzednim silnikiem.** Dane ze starszych leksykonów mogą służyć do analizy językoznawczej, ale Go model jest projektowany od nowa.

## Inne języki i przyszłość

Operacje generyczne mogą wspierać prefiksy, alternacje, harmonię i root templates, lecz nie deklarują automatycznej kompletnej morfologii arabskiej czy tureckiej. PL/EN są celem v0. Generatory muszą być ograniczone zasobami i nie mogą wykonywać kodu użytkownika, nawet jeśli prawie niczego automatycznie nie testujemy.

# 07 — Proceduralna morfologia, generatory i wyjątki

**PROPOSED, G02.** Podstawowa zasada: opisujemy **klasy i reguły**, a nie przechowujemy wszystkich regularnych form pojedynczych słów.

## Osobne komponenty

`MorphLexeme`: lemat, POS, język, cechy własne i przypisana klasa. Nie wymaga zatwierdzonego `CoreRef`, aby generator mógł go odmienić.

`Paradigm`: typowana, lokalna dla języka klasa stabilnych komórek cech i operacji; może dziedziczyć wspólne przepisy. Implementacja Go znajduje się w `internal/morph`. Komórka ma stabilny `CellID` oraz jawny zestaw znanych wartości z `FeatureProfile`. Reguła wskazuje komórkę i deklaratywny plan.

`StemSet / PrincipalParts`: tematy i dodatkowe formy diagnostyczne dla nieregularności (np. `go → went`). `ResolveStem` zwraca wyłącznie formę podaną jawnie dla wymaganego principal part; brak wymaganej formy daje `unknown`. Nie udawaj, że każdą formę można wywnioskować z samego lematu.

`FeatureProfile`: dopuszczone pola i wartości należą do **języka**, a nie do wspólnego CORE. Neutralna unifikacja cech w `internal/features`.

`ProcedurePlan`: wspólny, ograniczony i zamknięty zestaw kroków (`select_stem`, `append`, `prepend`, `strip_suffix`, `replace_suffix`, `alternation`). Nieznane operacje i niezadeklarowane warianty są odrzucane; nie ma `eval`, poleceń powłoki ani kodu z JSON. Interpreter liczy limity wejścia/wyjścia w punktach kodowych oraz bajtach, waliduje guardy cech i zwraca ślad kroków. Znana cecha niedopasowana do guarda pomija krok; brak cechy kończy się `unknown`, a nierozstrzygnięta wartość `incomplete`. Operacje `strip_suffix`, `replace_suffix` i `alternation` wymagają dokładnego dopasowania; brak dopasowania nie jest cichym fallbackiem. Obecna alternacja oznacza jedno literalne zastąpienie pierwszego wystąpienia, nie ogólną obsługę procesów fonologicznych.

Kompilator sprawdza wersję profilu, komórki i reguły, cykle dziedziczenia, konflikty oraz wymagania form diagnostycznych. JSON paradygmatu ma jawne wymagane pola, odrzuca nieznane lub błędnie zapisane klucze oraz ogranicza dokument do 1 MiB. Dziedziczenie zachowuje stabilne sloty; wspólny identyczny przodek w diamencie jest scalany raz, a sprzeczne definicje są błędem. Interpretor pozostaje niezależny od języka, zaś `Paradigm` i jego etykiety należą do języka i wersji profilu.

`ExceptionRule`: tylko potrzebny wyjątek z określonym zakresem i efektem.

## Generator

`(LexemeID, FeatureBundle) → select paradigm/cell → pick stem → operations → exception → legal surface variants`.

Reguły sprawdzają POS, profil, dostępność komórki i zgodność wyjątków. Kiedy forma jest zablokowana, generator **nie może** wrócić do regularnej postaci przez obejście. Gdy jej nie znamy, odpowiedź brzmi `unknown/unavailable`, nie wymyślona forma. Pusta końcówka (zero exponent) nie jest brakującą komórką.

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

`Analyze("read")` powinien móc zwrócić wiele interpretacji dla jednego zapisu, np. formy odpowiadające różnym komórkom. Generator i analyzer korzystają z tych samych reguł, ale nie muszą działać przez ten sam algorytm. Na początek wystarczy prosta, ograniczona finite relation table + reverse index.

Ani FST, ani lazy backend nie są obowiązkowym warunkiem. Nie udowadniamy teraz przewagi wydajności nad starym JavaScriptem.

## Rzeczywiste przykłady użytkownika

Właściciel może podawać konkretne leksemy i żądane formy PL/EN. CLI zwraca pełną listę, dostępne cechy i informację `generated / exception / blocked / unknown`. Na podstawie obserwacji poprawiamy klasę, temat lub wyjątek.

**Nie tworzymy obowiązkowych goldów, prób jednostkowych dla każdej komórki, matrycy odmiany ani porównań z poprzednim silnikiem.** Dane ze starszych leksykonów mogą służyć do analizy językoznawczej, ale Go model jest projektowany od nowa.

## Inne języki i przyszłość

Operacje generyczne mogą wspierać prefiksy, alternacje, harmonię i root templates, lecz nie deklarują automatycznej kompletnej morfologii arabskiej czy tureckiej. PL/EN są celem v0. Generatory muszą być ograniczone zasobami i nie mogą wykonywać kodu użytkownika, nawet jeśli prawie niczego automatycznie nie testujemy.

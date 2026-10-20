# 03 — Dane, kompilacja i format runtime

**PROPOSED.** Źródłem prawdy nowego silnika są nowe struktury Go i jego nowe, proste dane. Nie przenosimy starego kodu JS. Wcześniejsze leksykony/koncepcje mogą być pomocniczym źródłem wiedzy lub opcjonalnym zbiorem do importu, lecz **nie ma obowiązku ich kopiowania ani porównywania wyników**.

## Docelowe poziomy danych

| Poziom | Rola | Przykładowe pliki |
| --- | --- | --- |
| Definicje | małe, czytelne, wersjonowane deklaracje znaczeń, klas odmiany, konstrukcji, operatorów i wyjątków | `data/core/*.json`, `languages/pl/**`, `languages/en/**` |
| Opcjonalny import | narzędzie do przetwarzania dodatkowych słowników PL/EN bez włączania ich automatycznie | `internal/importer/*`, `imports/*` |
| Pakiet runtime | skompilowane, walidowane, niemutowalne indeksy i plany | `internal/pack/*` |
| Wyniki ręczne | tekst/diagnostyka z faktycznego wywołania Go CLI; opcjonalne notatki | `examples/manual/*.md` lub komentarz issue |

Brak `tests/cases`, masowego importu starych expected, obowiązkowego gold-index i runnera. Nie wybieramy od razu SQLite, FST ani specjalnego binarnego formatu pakietu.

## Własności kompilatora

1. Definicje mają jawny typ i symbol, np. `ConceptID`, `LexemeID`, `SenseID`, `ParadigmID`, `ConstructionID`. Nazwa nie określa automatycznie ontologii ani POS.
2. Powiązania do nieistniejących symboli są błędami danych; nie zgaduj brakujących mapowań.
3. Rozpoznawanie wyjątku nie zależy od kolejności JSON. Konflikty muszą powodować czytelną diagnostykę.
4. Rzeczywiście wspólne paradygmaty i konstrukcje dzielą plany; nie materializujemy ich oddzielnie dla każdego czasownika.
5. Indeks odwrócony form zachowuje wszystkie legalne odczytania. Usuwanie alternatyw nie jest optymalizacją.
6. Staged data może być niekompletne; nie powinno po cichu stać się aktywną wiedzą.
7. Zmiana danych tworzy nowy snapshot pakietu; istniejący runtime nie powinien mieć mutowanych globalnych map.

## Wydajność i magazyn

W nowym prototypie Go najpierw proste struktury: tablice i mapy numerycznych ID, indeksy `surface→readings`, `lemma→lexemes`, `concept→senses`, `sense→valency` i `construction→plans`. Liczne relacje ontologiczne mogą korzystać z indeksu sąsiedztwa, ale nie potrzeba grafowej bazy do 1000 konceptów.

Baza SQLite mogłaby pomóc redaktorowi danych, nie musi być w runtime. FST/lazy morphology są odroczone: nie przenoś wymagań porównawczych ze starego engine. Benchmark optymalizacyjny wolno uruchomić później, gdy użytkownik wyraźnie go potrzebuje; nie jest standardowym gate przed PR.

## Projekt od zera, licencje

Do rozwoju PL/EN można wykorzystać projektowo własne definicje i ręczne reguły. Źródła zewnętrzne wymagają odnotowania licencji i pochodzenia; to **kwestia jakości danych**, a nie automatyczna ścieżka testowa. Zweryfikuj konkretne wydanie, jeśli importujesz obce zasoby. Nie pobieraj danych z internetu w czasie obsługi zdań.

## Aktualny status

Bootstrap posiada `go.mod`, CLI, generyczny pakiet typowanych wyników oraz `internal/ids` z odrębnymi stable ID wrappers i deterministyczną snapshot-local symbol table (G01-01). `internal/core` dodaje typowane definicje i niemutowalny ConceptGraph z walidacją relacji (G01-02); `internal/graph` kompiluje bounded adjacency i traversal (G01-03); `internal/meaning` definiuje versioned request-scoped Meaning IR, cross-CORE role validation i fail-closed operator projection (G01-04/G01-05). Wersjonowany JSON Schema opisuje format ConceptGraph. G00-04 dodaje `SourceRef` i dokumentację manifestów, ale nie dodaje loadera pakietów, aktywnego słownika ani pakietów językowych. Pozostałe planowane struktury leksykalne i składniowe są **PROPOSED**, dopóki ich implementacyjne issue nie zakończy się rzeczywistym scaleniem. Nie powołuj się na procenty pokrycia, jeśli nie ustaliliśmy jeszcze docelowej zawartości nowego słownika.

### Pack CORE + lexicon (format v1)

`internal/pack` kompiluje dokładnie jeden zwalidowany, aktywny snapshot wskazany przez `staging.Store.Active`. Pack jest niemutowalny i związany z `dataset_id`, wersją oraz digestem zawartości snapshotu. Gęste `ids.NodeID` są stabilne tylko w ramach tego packa; publiczne identyfikatory Concept/Lexeme/Sense pozostają tożsamością źródłową. CORE udostępnia niemutowalny indeks sąsiedztwa z `internal/graph`; `lexicon.Index` zapewnia lemma→lexemes, lexeme→senses, sense→bindings i concept→bindings, zachowując homonimy oraz wiele odczytów.

Pack v1 działa wyłącznie na CORE i leksykonie obecnego snapshotu. Nie zawiera profili językowych, principal parts, paradigmatów, generowanych form, walencji ani konstrukcji. Import propozycji UniMorph nie aktywuje morfologii. Pole podsumowania `morphology`/`construction` ma wartość `not_included`; brak tych danych nie oznacza braku znaczenia ani niepowodzenia ładowania. `golingua-import pack inspect` tylko odczytuje aktywny snapshot i pokazuje przypiętą tożsamość oraz liczniki; nie zmienia wskaźnika ani nie deklaruje gotowości translacji.

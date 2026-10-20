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

Bootstrap posiada `go.mod`, CLI, generyczny pakiet typowanych wyników oraz `internal/ids` z odrębnymi stable ID wrappers i deterministyczną snapshot-local symbol table (G01-01). G00-04 dodaje `SourceRef` i dokumentację manifestów, ale nie dodaje loadera, aktywnego słownika, walidatora pełnych rekordów ani pakietów językowych. Pozostałe planowane struktury leksykalne, grafowe i składniowe są **PROPOSED**, dopóki ich implementacyjne issue nie zakończy się rzeczywistym scaleniem. Nie powołuj się na procenty pokrycia, jeśli nie ustaliliśmy jeszcze docelowej zawartości nowego słownika.

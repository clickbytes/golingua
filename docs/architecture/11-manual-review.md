# 11 — Ręczne sprawdzanie, minimalny CI i prosta diagnostyka

**Nadrzędna decyzja 2026-10-20:** obecny etap golingua **nie posiada automatycznej macierzy poprawności języka, formalnych goldów, test runnerów ani bramki zgodności z poprzednim JavaScriptem**.

## Co naprawdę sprawdzamy w każdym PR

**Technicznie (GitHub CI):** gofmt, go vet, go build, uruchomienie CLI z `--version` i `--help`. To weryfikuje, że projekt się buduje i startuje. Nie jest dowodem jakości tłumaczenia.

**Językowo (na realnych danych):** właściciel dostarcza zdania/słowa i kierunek. Agent, gdy ma możliwość faktycznego wykonania Go, zapisuje komendę, wynik, SHA i ewentualne MeaningGraph/diagnostyki. Od 2026-10-23 właściciel deleguje agentowi autonomiczną ocenę i poprawki dla dostarczonych przykładów. Ocena agenta i akceptacja właściciela to odrębne statusy; agent kontynuuje bez oczekiwania na odpowiedź, ale nie przypisuje właścicielowi własnej oceny. Jeżeli nie uruchomiono CLI, status brzmi `NOT RUN`.

Można mieć zero udostępnionych przykładów na starcie i nadal pracować nad architekturą. Nie wolno udawać, że brak błędów zgłoszonych przez użytkownika to 100% poprawności.

## Minimalne zapisy

Krótka notatka lub komentarz:
```text
SHA: <sha>
Kierunek: pl → en
Wejście: <zdanie od użytkownika>
Polecenie: <rzeczywiste CLI>
Wynik: <dokładny output i ewentualny status>
Ocena agenta: OK | DO POPRAWY | NIEPEWNE | BRAK OBSŁUGI | NOT RUN
Akceptacja właściciela: OK | DO POPRAWY | NOT RUN
Poprawka: <klasa/leksem/rekcja/generator/ConstructionPlan>
```

Takie notatki nie są automatycznie wykonywane, nie są testami obowiązkowymi i nie tworzą formalnego gold standardu. [Instrukcja ręcznego review](../manual-review.md).

## Co musi pozostać w kodzie, mimo rezygnacji z testów

- Walidacja typów i ograniczeń wejścia, bo inaczej program nie jest bezpiecznie uruchamialny.
- Rozróżnianie `unknown`, `unsupported`, `ambiguous`, `invalid` i `resource_limit`.
- Wybór wszystkich dopuszczonych odczytań formy, bez fałszywej pewności.
- Blokowanie nielegalnych wyjątków, cykli konstrukcji i nieobsługiwanych kombinacji.
- Trace z LexemeID/SenseID/ConstructionID i wykorzystanych reguł, przynajmniej w trybie diagnostycznym.
- Odróżnienie faktów od definicji pojęć i pytania od asercji, jeśli później dojdzie reasoning.

To **wymagania logiki**, nie zadania polegające na napisaniu setek sprawdzianów. Ich niedokładności wykrywamy i korygujemy na konkretnych przykładach właściciela.

## Czego nie implementować

- `tests/cases` / `fixtures/golden` / `internal/testing/*` jako framework oceny językowej.
- Automatycznego go test `./...` w CI albo wymogu pisania `*_test.go` dla każdej funkcji.
- Migracji poprzednich 593 zdań, niezależnych reviewer approvals per sentence, procentowego coverage sentence/act.
- Benchmarku Go względem JS jako warunku zakończenia EPIC.
- Autogeneratora zdań testowych, który później porównuje output z samym sobą.

## Później

Gdy baza będzie stabilniejsza, można osobną decyzją użytkownika dodać konkretne testy i pomiary. Nie projektuj już dziś ich skomplikowanych kontraktów. Jedyny obecny obowiązek to uczciwa informacja: czy kod się kompiluje, co faktycznie powiedział Go, jak wynik ocenił agent i czy właściciel osobno go zaakceptował.

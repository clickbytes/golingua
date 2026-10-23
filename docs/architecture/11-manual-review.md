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

## G10-12 — `offspring` i allomorfy `a/an`

Agent porównał istniejący sens `en.offspring.child` („a person's child,
considered in relation to that parent”) z zatwierdzonym polskim
`pl.dziecko.kinship` („a person in relation to their own parent, at any age”).
Oba opisują tę samą relację rodzic–dziecko; odróżniono ją od angielskiego
`en.child.person` oznaczającego młodą osobę. Zapis agenta z provenance znajduje
się w [G10-12 #237](https://github.com/clickbytes/golingua/issues/237#issuecomment-6357615617).
Binding `en.offspring.child → concept:child_of_person` ma status `approved`;
akceptacja właściciela pozostaje **NOT RUN**.

Rzeczywiste wywołania dla 20 istniejących rekordów S021–S025, S036–S040,
S071–S075 i S096–S100, w obu kierunkach, zapisano w
[tabeli obserwacji G10-12](../../data/sentences/observed-2026-10-23-g10-12-offspring-a-an.tsv).
W 20 istniejących rekordach EN→PL zwróciło 15 `ok` i 5 `incomplete`;
niekompletne przypadki zachowały diagnostyki nierozstrzygniętego sensu `book`. PL→EN zwróciło 20 wyników
`ambiguous`, w których widoczne są leksykalne alternatywy `child`/`offspring`
i warianty rodzajnika. Pochodna próba `An offspring sees a dog.` →
`Dziecko widzi psa.` zwróciła `ok`; nie jest częścią korpusu właściciela. Parser odrzuca
`A offspring...`, a `an` wyklucza mnogą analizę `offspring`. Agent uznaje
rozróżnienie form i jawność niepewności za zgodne z zadeklarowanym modelem;
pełna ocena językowa właściciela nadal **NOT RUN**.

## G10-13 — present 3pl i uzgodnienie podmiotu z czasownikiem

Issue #239 / PR #240 włącza komórki trzeciej osoby liczby mnogiej na już
istniejących trasach `seeing`, `hearing` i `looking` w PL i EN. Rodzime cechy
`noun_number`/`verb_number` są normalizowane do `number` na granicy analizy i
realizacji; twarda reguła konstrukcji odrzuca niezgodne liczby. Angielskie
przeszłe `got` pozostało w rodzinie bez uzgodnienia, a polski przeszły `dostać`
nadal wybiera jawne cechy liczby i rodzaju.

W istniejących 40 wywołaniach dla S021–S025, S036–S040, S071–S075 i S096–S100 nie zmieniły się statusy ani zbiory tekstów względem poprzedniej obserwacji. Dziesięć osobno oznaczonych prób pochodnych oraz dokładne wyjścia CLI są w [tabeli G10-13](../../data/sentences/observed-2026-10-23-g10-13-plural-agreement.tsv). Formy mnogie generują się i są analizowane; `offspring` dobiera liczbę czasownika, a niezgodne kombinacje podmiotu i czasownika nie otrzymują pełnego parsowania. PL→EN nadal pokazuje alternatywy leksykalne/rodzajnikowe jako `ambiguous`. Agent ocenia zachowanie jawnych form i cechy liczby jako spójne z trasą; ocena językowa właściciela pozostaje **NOT RUN**.

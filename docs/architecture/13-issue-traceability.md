# 13 — Mapa pracy szeregowej po resecie testów

**Status: żywa mapa etapów, aktualizacja 2026-10-23; bieżący `main` przed PR #208: `3d2633e35b905cdf96d15b3084d606ff3383d313`.** G00–G03 zamknięte, G09 odroczony. G05 #6 i G07 #8 są zamknięte technicznie; ręczna akceptacja właściciela pozostaje osobnym statusem. G04 #5 ma zakończone zadania implementacyjne, a #207 wykonuje końcową obserwację CLI i closeout techniczny. G06 #7 jest następny w kolejce; #128 wymaga realizacji danych odpowiedzi PL/EN, a #130 wejścia do oceny typed CLI. G08 #9 pozostaje późniejszym etapem. `catalog --summary --json` raportuje: CORE 1000; PL 982 leksemy / 1007 sensów / 1006 bindingów (882 zatwierdzone koncepty, 118 konceptów pending); EN 1002 leksemy / 1011 sensów / 1001 bindingów (996 zatwierdzonych konceptów, 4 koncepty pending); 880 zatwierdzonych par PL+EN. Liczba pending binding records PL wynosi 124, bo część konceptów ma alternatywne kandydatury. Dodatkowy determiner EN `a` jest sentence-only i nie ma CORE bindingu. Historyczne G00 #78 nie jest obecnym aktywnym zadaniem.

| Kolejność | EPIC | Co zbudować | Jak oceniamy na tym etapie |
| --- | --- | --- | --- |
| 0 | [G00 #1](https://github.com/clickbytes/golingua/issues/1) | Bootstrap, granice typów, diagnostyka, źródła danych, prosty CLI | kompilacja Go i zrozumiała pomoc CLI |
| 1 | [G01 #2](https://github.com/clickbytes/golingua/issues/2) | ConceptGraph, identyfikatory, MeaningGraph | czy API ładnie reprezentuje i wypisuje pojęcia, bez unit test suite |
| 2 | [G02 #3](https://github.com/clickbytes/golingua/issues/3) | Lexicon, morphology, procedury i wyjątki | wybrane przez właściciela słowa i odmiany |
| 3 | [G03 #4](https://github.com/clickbytes/golingua/issues/4) | ValencyFrame, ConstructionFamily/Variant, NP/VP/PP, liczebniki/time/MWE | ręczna próba wybranych konstrukcji użytkownika |
| 4 | [G04 #5](https://github.com/clickbytes/golingua/issues/5) | Tworzenie/import danych i kompilacja niemutowalnych pakietów | implementacja gotowa; #207 weryfikuje snapshot i odmianę na CLI przed closeoutem |
| 5 | [G05 #6](https://github.com/clickbytes/golingua/issues/6) | Analyze i Realize PL↔EN | rzeczywiste wejścia właściciela; ocena agenta jest delegowana, akceptacja właściciela osobna |
| 6 | [G06 #7](https://github.com/clickbytes/golingua/issues/7) | ograniczony reasoner i ContentPlanner | manualne pytania i FactSet od użytkownika |
| 7 | [G07 #8](https://github.com/clickbytes/golingua/issues/8) | około 1000 konceptów z PL/EN | zakres katalogu ukończony; closeout #205, ręczny odbiór zdań pozostaje osobnym NOT RUN |
| 8 | [G08 #9](https://github.com/clickbytes/golingua/issues/9) | uporządkowanie CLI, prosta diagnostyka, optymalizacja tylko gdy potrzebna | build CI, ręczne użycie, brak porównania z JS |
| 9 | [G09 #10](https://github.com/clickbytes/golingua/issues/10) | przyszła pamięć i uczenie | DEFERRED, brak implementacji v0 |

## Kryterium wykonania tasku

1. Jeden ACTIVE EPIC, jeden task i jeden agent; zwykle jeden PR. Wyjątek właściciela dla #60: sekwencyjne commity na main i CI.
2. Wprowadź nazwane struktury/funkcje/pliki; nie rozbudowuj rozwiązania ponad granice etapu.
3. CI Go: format, vet, build i start CLI. Brak obowiązkowych unit tests.
4. Gdy funkcja wytwarza formy/zdania, uruchom dostępne przykłady właściciela i zapisz dokładne obserwacje; agent samodzielnie ocenia i poprawia w granicach delegacji. Akceptacji właściciela nie oznaczaj jako wykonanej, jeśli jej nie udzielił.
5. Zapisz realne błędy i poprawki w issue; nie ma automatycznej akceptacji językowej na podstawie starszych wyników.
6. Po scaleniu i ponownym sprawdzeniu głównej gałęzi rozpocznij kolejny task.

## Bieżące szlaki i obserwacje

Lokalny stan po audycie #192 zawiera 982 leksemy PL, 1007 sensów, 1006 bindingów (882 approved i 124 pending), 1002 leksemy EN, 1011 sensów i 1001 CORE bindingów. Wszystkie 1000 konceptów mają rekordy; PL ma 118 pending konceptów, a EN 4. Dodatkowy angielski determiner `a` służy wyłącznie analizie konstrukcji i nie ma CORE bindingu. Indeks konstrukcji ma 9 zatwierdzonych tras (PL 4, EN 5) po review #63 i #201. Katalog raportuje 880 PL+EN approved pairs. CLI obsługuje ograniczone szlaki `seeing`, `hearing`, `looking` i `receiving`; rzeczywiste wyjścia 25 przykładów `seeing` znajdują się w [obserwacjach początkowego runtime'u](../../data/sentences/observed-2026-10-23-provider-runtime.tsv), `hearing` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-12-hearing.tsv), `receiving` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-13-receiving.tsv), a `looking` w [najnowszej tabeli](../../data/sentences/observed-2026-10-23-g05-14-looking.tsv). We wszystkich czterech zestawach uruchomiono 25 istniejących par w obu kierunkach (100 par, 200 uruchomień CLI). Katalog G07 i jego zadania są technicznie ukończone; 118 konceptów PL i 4 EN pozostają pending, bez automatycznego awansu. Akceptacja językowa właściciela dla wyników zdań pozostaje **NOT RUN**; obserwacje działania i ocena agenta są udokumentowane niezależnie. Nie wyciągać wniosków o jakości na podstawie liczników lub samego statusu wykonania.

## Co zostało świadomie usunięte

Nie istnieje obowiązek przeniesienia dawnych zdań, wejść oracli i spodziewanych tekstów JS. Nie planujemy nowego katalogu `fixtures/parity`, automatycznych runnerów w `internal/testing`, protokołu niezależnego recenzowania expected zdań ani przelicznika „ile zdań przeszło”. To nowe źródło danych i nowy silnik.

Po decyzji użytkownika przyjmujemy również, że nie ma wymaganego dotted/property API `ROCK.VERB...`. Typowanie i kompozycja pozostają kluczowe, sama nawigacja API nie.

G03-08 #74, G03-09 #75 i G04-07 #76 są zamknięte historycznie. G04 #38–43/#76 zakończyły implementację; #207 zapisuje aktualny smoke snapshotu i morfologii, a odbiór właściciela pozostaje odrębny. G05 #123, #195, #197, #199 i #201 są zakończone; #203 zsynchronizował mapę i G05. G07 #58–64, #145, #152–192 są zamknięte; #205 zsynchronizował closeout epiku, bez zmiany nierozstrzygniętych kandydatur ani statusu oceny językowej. G06 #128 jest następnym taskiem do realizacji na bazie delegowanego przeglądu; #130 pozostaje wejściem do ręcznej oceny CLI. G08 #65/#66 są późniejszym etapem zadań językowych i ręcznego przeglądu.

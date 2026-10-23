# 13 — Mapa pracy szeregowej po resecie testów

**Status: żywa mapa etapów, aktualizacja 2026-10-23 po merge #202; bazowy `main` przed closeout PR #204: `54ae9d11ade0a16a2c602555a28d3b47e1dca756`.** G00–G03 zamknięte, G09 odroczony, G07 #8 oczekuje na kolejkę mimo zakończenia CORE #145, PL #60, EN #61 i review konstrukcji #63. Audyt #192 zakończony: `catalog --summary` raportuje PL 982 leksemy / 1007 sensów / 1006 bindingów (882 zatwierdzone koncepty, 118 pending) oraz EN 1002 leksemy / 1011 sensów / 1001 bindingów (996 zatwierdzonych konceptów, 4 pending; dodatkowy determiner `a` jest sentence-only i nie ma CORE bindingu). G05 #6 zakończył implementacyjne tasks #123, #195, #197, #199 i #201; PR #202 scala ostatni szlak `looking`. G05 closeout task #203 synchronizuje tę mapę. G04, G06 i G08 oczekują na swoje zależności. Historyczne G00 #78 nie jest obecnym aktywnym zadaniem.

| Kolejność | EPIC | Co zbudować | Jak oceniamy na tym etapie |
| --- | --- | --- | --- |
| 0 | [G00 #1](https://github.com/clickbytes/golingua/issues/1) | Bootstrap, granice typów, diagnostyka, źródła danych, prosty CLI | kompilacja Go i zrozumiała pomoc CLI |
| 1 | [G01 #2](https://github.com/clickbytes/golingua/issues/2) | ConceptGraph, identyfikatory, MeaningGraph | czy API ładnie reprezentuje i wypisuje pojęcia, bez unit test suite |
| 2 | [G02 #3](https://github.com/clickbytes/golingua/issues/3) | Lexicon, morphology, procedury i wyjątki | wybrane przez właściciela słowa i odmiany |
| 3 | [G03 #4](https://github.com/clickbytes/golingua/issues/4) | ValencyFrame, ConstructionFamily/Variant, NP/VP/PP, liczebniki/time/MWE | ręczna próba wybranych konstrukcji użytkownika |
| 4 | [G04 #5](https://github.com/clickbytes/golingua/issues/5) | Tworzenie/import danych i kompilacja niemutowalnych pakietów | kontrola zawartości, legalności danych i wyników CLI, bez starej macierzy |
| 5 | [G05 #6](https://github.com/clickbytes/golingua/issues/6) | Analyze i Realize PL↔EN | rzeczywiste wejścia właściciela; ocena agenta jest delegowana, akceptacja właściciela osobna |
| 6 | [G06 #7](https://github.com/clickbytes/golingua/issues/7) | ograniczony reasoner i ContentPlanner | manualne pytania i FactSet od użytkownika |
| 7 | [G07 #8](https://github.com/clickbytes/golingua/issues/8) | około 1000 konceptów z PL/EN | przegląd deklarowanych danych i ręczne tłumaczenia, bez procentu jakości |
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

Lokalny stan po audycie #192 zawiera 982 leksemy PL, 1007 sensów, 1006 bindingów (882 approved i 124 pending), 1002 leksemy EN, 1011 sensów i 1001 CORE bindingów. Wszystkie 1000 konceptów mają rekordy; PL ma 118 pending konceptów, a EN 4. Dodatkowy angielski determiner `a` służy wyłącznie analizie konstrukcji i nie ma CORE bindingu. Indeks konstrukcji ma 9 zatwierdzonych tras (PL 4, EN 5) po review #63 i #201. Katalog raportuje 880 PL+EN approved pairs. CLI obsługuje ograniczone szlaki `seeing`, `hearing`, `looking` i `receiving`; rzeczywiste wyjścia 25 przykładów `seeing` znajdują się w [obserwacjach początkowego runtime'u](../../data/sentences/observed-2026-10-23-provider-runtime.tsv), `hearing` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-12-hearing.tsv), `receiving` w [tej tabeli](../../data/sentences/observed-2026-10-23-g05-13-receiving.tsv), a `looking` w [najnowszej tabeli](../../data/sentences/observed-2026-10-23-g05-14-looking.tsv). We wszystkich czterech zestawach uruchomiono 25 istniejących par w obu kierunkach. Akceptacja językowa właściciela pozostaje **NOT RUN**; wynik działania i ocena agenta są udokumentowane niezależnie. Nie wyciągać wniosków o jakości na podstawie liczników lub samego statusu wykonania.

## Co zostało świadomie usunięte

Nie istnieje obowiązek przeniesienia dawnych zdań, wejść oracli i spodziewanych tekstów JS. Nie planujemy nowego katalogu `fixtures/parity`, automatycznych runnerów w `internal/testing`, protokołu niezależnego recenzowania expected zdań ani przelicznika „ile zdań przeszło”. To nowe źródło danych i nowy silnik.

Po decyzji użytkownika przyjmujemy również, że nie ma wymaganego dotted/property API `ROCK.VERB...`. Typowanie i kompozycja pozostają kluczowe, sama nawigacja API nie.

G03-08 #74, G03-09 #75 i G04-07 #76 są zamknięte historycznie. G05 #123, #195, #197, #199 i #201 są zakończone; #203 synchronizuje mapę zamknięcia. G06 #128 oczekuje na zatwierdzone konstrukcje odpowiedzi, a #130 na wejścia właściciela do ręcznej oceny CLI. G08 #65/#66 oczekują na działające funkcje i rzeczywiste przykłady właściciela. G07 #60/#61/#63 zakończono w udokumentowanym zakresie.

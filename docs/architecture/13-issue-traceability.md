# 13 — Mapa pracy szeregowej po resecie testów

**Status: żywa mapa etapów, aktualizacja 2026-10-23.** G00–G03 zamknięte, G09 odroczony, G07 #8 oczekuje: CORE #145 i inwentarze PL #60 oraz EN #61 zakończone. Audyt #192 zakończony: lokalny `catalog --summary` raportuje PL 982 leksemy / 1007 sensów / 1006 bindingów (882 zatwierdzone koncepty, 118 pending) oraz EN 1001 leksemów / 1010 sensów / 1001 bindingów (996 zatwierdzonych konceptów, 4 pending). G05 #6 jest aktywny; #123 i #195 zakończono po merge, a jedynym aktywnym zadaniem jest #197. G04, G06 i G08 oczekują na swoje zależności i ręczny odbiór. Historyczne G00 #78 nie jest obecnym aktywnym zadaniem.

| Kolejność | EPIC | Co zbudować | Jak oceniamy na tym etapie |
| --- | --- | --- | --- |
| 0 | [G00 #1](https://github.com/clickbytes/golingua/issues/1) | Bootstrap, granice typów, diagnostyka, źródła danych, prosty CLI | kompilacja Go i zrozumiała pomoc CLI |
| 1 | [G01 #2](https://github.com/clickbytes/golingua/issues/2) | ConceptGraph, identyfikatory, MeaningGraph | czy API ładnie reprezentuje i wypisuje pojęcia, bez unit test suite |
| 2 | [G02 #3](https://github.com/clickbytes/golingua/issues/3) | Lexicon, morphology, procedury i wyjątki | wybrane przez właściciela słowa i odmiany |
| 3 | [G03 #4](https://github.com/clickbytes/golingua/issues/4) | ValencyFrame, ConstructionFamily/Variant, NP/VP/PP, liczebniki/time/MWE | ręczna próba wybranych konstrukcji użytkownika |
| 4 | [G04 #5](https://github.com/clickbytes/golingua/issues/5) | Tworzenie/import danych i kompilacja niemutowalnych pakietów | kontrola zawartości, legalności danych i wyników CLI, bez starej macierzy |
| 5 | [G05 #6](https://github.com/clickbytes/golingua/issues/6) | Analyze i Realize PL↔EN | tylko manualne zdania, które dostarczy użytkownik |
| 6 | [G06 #7](https://github.com/clickbytes/golingua/issues/7) | ograniczony reasoner i ContentPlanner | manualne pytania i FactSet od użytkownika |
| 7 | [G07 #8](https://github.com/clickbytes/golingua/issues/8) | około 1000 konceptów z PL/EN | przegląd deklarowanych danych i ręczne tłumaczenia, bez procentu jakości |
| 8 | [G08 #9](https://github.com/clickbytes/golingua/issues/9) | uporządkowanie CLI, prosta diagnostyka, optymalizacja tylko gdy potrzebna | build CI, ręczne użycie, brak porównania z JS |
| 9 | [G09 #10](https://github.com/clickbytes/golingua/issues/10) | przyszła pamięć i uczenie | DEFERRED, brak implementacji v0 |

## Kryterium wykonania tasku

1. Jeden ACTIVE EPIC, jeden task i jeden agent; zwykle jeden PR. Wyjątek właściciela dla #60: sekwencyjne commity na main i CI.
2. Wprowadź nazwane struktury/funkcje/pliki; nie rozbudowuj rozwiązania ponad granice etapu.
3. CI Go: format, vet, build i start CLI. Brak obowiązkowych unit tests.
4. Gdy funkcja wytwarza formy/zdania, daj właścicielowi możliwość ręcznego sprawdzenia dokładnego outputu. Jeśli nie ma inputu lub możliwości wykonania — oznacz „nie oceniono ręcznie”.
5. Zapisz realne błędy i poprawki w issue; nie ma automatycznej akceptacji językowej na podstawie starszych wyników.
6. Po scaleniu i ponownym sprawdzeniu głównej gałęzi rozpocznij kolejny task.

## Bieżący ręczny odbiór PL

Lokalny stan po audycie #192 zawiera 982 leksemy PL, 1007 sensów, 1006 bindingów (882 approved i 124 pending), 1001 leksemów EN, 1010 sensów i 1001 bindingów. Wszystkie 1000 konceptów mają rekordy; PL ma 118 pending konceptów, a EN 4. Osiem sense/frame/family/variant construction bindings (4 PL, 4 EN) zatwierdzono po delegowanym przeglądzie #63. Katalog raportuje 880 PL+EN approved pairs. CLI ma ograniczone, uruchamialne szlaki `seeing` i `hearing`; rzeczywisty output dla 25 istniejących przykładów `hearing` w obu kierunkach zapisano w [tabeli obserwacji](../../data/sentences/observed-2026-10-23-g05-12-hearing.tsv). Ocena językowa właściciela nadal **NOT RUN**. Nie wyciągać wniosków o jakości na podstawie liczników lub samego statusu wykonania.

## Co zostało świadomie usunięte

Nie istnieje obowiązek przeniesienia dawnych zdań, wejść oracli i spodziewanych tekstów JS. Nie planujemy nowego katalogu `fixtures/parity`, automatycznych runnerów w `internal/testing`, protokołu niezależnego recenzowania expected zdań ani przelicznika „ile zdań przeszło”. To nowe źródło danych i nowy silnik.

Po decyzji użytkownika przyjmujemy również, że nie ma wymaganego dotted/property API `ROCK.VERB...`. Typowanie i kompozycja pozostają kluczowe, sama nawigacja API nie.

G03-08 #74, G03-09 #75 i G04-07 #76 są zamknięte historycznie. G05 #123 i #195 zakończono; #197 rozwija zatwierdzony szlak `hearing`. G06 #128 oczekuje na zatwierdzone konstrukcje odpowiedzi, a #130 na wejścia i ręczną ocenę właściciela. G08 #65/#66 oczekują na działające funkcje i przykłady właściciela. G07 #60/#61/#63 zakończono w udokumentowanym zakresie.

# 12 — Decyzje po resecie prototypu 2026-10-20

**Ważna aktualizacja:** użytkownik wyraźnie odrzucił obowiązkowe porównywanie Go ze starym silnikiem oraz rozbudowaną automatyczną macierz. **Piszemy od zera i ręcznie sprawdzamy output użytkownika.** Ta decyzja ma pierwszeństwo przed starszymi sformułowaniami w issues.

## Obowiązujące decyzje

| ID | Decyzja |
| --- | --- |
| D-01 | Nowy kod Go; stary JS jedynie opcjonalnym źródłem inspiracji/danych. Brak implementacyjnej zgodności ze starą architekturą. |
| D-02 | Jeden agent, jeden aktywny EPIC/task/PR. |
| D-03 | Kontrola CI minimalna: gofmt, vet, build, help/version; bez automatycznych testów językowych i Go unit suite. |
| D-04 | Tłumaczenia i odmiany sprawdza ręcznie właściciel na wybranych wejściach. |
| D-05 | ConceptGraph pozostaje niezależny od języków. |
| D-06 | MorphLexeme ≠ LexicalSense ≠ SemanticBinding ≠ Concept; możliwa morfologia bez CORE. |
| D-07 | ValencyFrame per sens, ConstructionFamily/Variant jako first-class graf konstrukcji. |
| D-08 | NP/VP/PP, liczebniki, czas, pytanie, negacja, MWE są kompozycyjnymi konstrukcjami/operacjami. |
| D-09 | Współdzielone generatory, principal parts, wyjątki `block/replace/addVariant` i rozdzielenie inverse forms. |
| D-10 | Language-specific feature profiles; neutralna struktura unifikacji w `internal/features`. |
| D-11 | Typowane ID, indeksy i niemutowalne plany; bez konieczności fluent `ROCK.VERB.MORPH.PAST` API. |
| D-12 | Języki jako data plugins, bez `plugin.Open` i bez rozgałęziania CORE po PL/EN. |
| D-13 | Brak z góry wybranego FST, SQL, graph DB i obowiązkowych benchmarków. |
| D-14 | MeaningGraph opisuje wypowiedź; FactSet ewentualnie opisuje jawne fakty; nie tworzymy globalnych domysłów. |
| D-15 | Około 1000 konceptów PL/EN to cel słownikowy, nie certyfikat jakości mowy ani 1000 testów. |
| D-16 | Pamięć długoterminowa i uczenie odroczone. |
| D-17 | Licencje/provenance zewnętrznych zasobów są obowiązkiem prawnym, nie frameworkiem testowym. |
| D-18 | Każde issue zachowuje ścieżki, symbole i DoD funkcji, ale nie wymaga goldów/fixtures/CI sprawdzającego tekst. |

## Ryzyka, które akceptujemy w świadomym prototypie

- **Ręczna ocena nie pokrywa wszystkich kombinacji.** Błędy mogą pojawić się później; nie obiecujemy coverage ani CEFR.
- **Ręcznie wprowadzane reguły mogą być niepełne.** Generator ma zwracać unsupported zamiast wymyślonej odmiany, kiedy brak danych.
- **Brak regresyjnej macierzy** zwiększa ryzyko powtórnego błędu przy zmianie paradygmatu; to zaakceptowana cena uproszczenia prac teraz.
- **1000 konceptów bez szerokiej gramatyki** nie oznacza swobodnego języka; rozwijamy konstruktory i generatory niezależnie od liczby słów.
- **Zielony CI nie sprawdza semantyki.** Poprawność zależy od późniejszej ręcznej recenzji rezultatów.
- **Wydajność Go nie jest automatyczna.** Na razie proste typowane struktury, profilowanie tylko gdy pojawi się realny problem.
- **Fakty językowe a fakty świata** to różne dziedziny; answer bez faktów musi pozostać unknown.

## Otwarta przyszłość

Formalne testy jednostkowe, testy paradygmatów, benchmarki i porównania silników mogą kiedyś wrócić **wyłącznie po nowej decyzji właściciela**. Nie przemycaj ich poprzez wymogi implementacyjne G02–G08. Uczenie i pamięć są osobnym etapem G09.

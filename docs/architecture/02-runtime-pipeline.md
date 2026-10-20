# 02 — Przepływy wykonania (nowy Go)

**PROPOSED.** Nie jest to opis istniejącego już działającego parsera.

## Kompilacja wiedzy

`źródło danych JSON (zapisane reguły) → walidator → Concept/Sense/Paradigm/Valency/Construction/Procedure → skompilowane, niemutowalne plany i indeksy`.

Zamiast tworzyć tysiące zagnieżdżonych obiektów jak `ROCK.VERB.MORPH.PAST`, łączymy wspólne struktury numerycznymi ID. Materializujemy tylko to, co rzeczywiście potrzebne do działania. Typy danych muszą być prostsze niż w prototypowym JS.

## Analiza

`text → tokenizacja → wszystkie odczytania morfologiczne → kandydaci sensów → dopuszczone konstrukcje i walencja → MeaningGraph`.

Nie wybieramy kandydata wyłącznie według kolejności importu. Gdy brak słowa/cechy/konstrukcji — `unsupported`/`unknown`; gdy kilka znaczeń — `ambiguous`. Jeśli limit przerwie analizę, wynik ma być niekompletny, nie fałszywie pewny.

## Realizacja

`MeaningGraph → zatwierdzony lub spójny SemanticBinding → ValencyFrame → językowa ConstructionFamily/Variant → procedura NP/VP/PP → formy fleksyjne → tekst`.

Czas, negacja, liczba, rekcja, role i referenty nie mogą znikać po drodze tylko dlatego, że silnik wybrał łatwiejszy szablon. Zmiana szyku PL może dać różną powierzchnię przy tych samych rolach, ale EN SVO ma własne reguły.

## Odpowiedź w kolejnych etapach

`QuestionMeaning + jawne FactSet → Reasoner (ograniczone reguły) → ContentPlanner → ResponseMeaning → PL/EN ConstructionPlan → odpowiedź`.

Brak FactSet nie jest dowodem fałszu; świat nie jest zamkniętą tabelą. Nie planujemy pamięci ani uczenia teraz.

## CLI jako główny wskaźnik funkcjonalny

Po dodaniu operacji Go udostępni przykładowo `golingua translate --from pl --to en --text "..."` i `golingua forms --language pl --word "..."`. Dokładne flagi ustali issue wprowadzające funkcję; nie należy ich na razie reklamować jako zaimplementowanych. Właściciel podaje wejścia, agent pokazuje wyniki, a korekty przeprowadzamy na regułach.

**Nie ma skryptu sprawdzającego setki zdań, założeń procentowych ani porównania do starego JS.** Minimalny CI tylko buduje Go. Ograniczenia runtime (maksymalna liczba kandydatów/węzłów) są cechą bezpieczeństwa, a nie celem generowania automatycznych testów.

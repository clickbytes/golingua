# 00 — Produkt i cel nowego prototypu

**Golingua = proceduralny silnik pojęciowo-językowy napisany od zera w Go.** Na tym etapie nie ma modelu uczącego się i nie ma trwałej pamięci. Pierwsze języki: PL, EN.

## Cele

- Typowany ConceptGraph i około **1000 konceptów** z leksemami/sensami PL i EN jako docelowy rozmiar prototypu, nie deklaracja sprawności całego języka.
- Konstrukcje jako first-class elementy: walencja per sens, rodziny konstrukcji, warianty, operator składniowy i generator.
- Formy fleksyjne proceduralne: lemat/klasa + wymagane tematy + wyjątki, bez ręcznego duplikowania wszystkich regularnych form.
- Analyze(surface) i Generate(lexeme,features); później pełne PL↔EN przez neutralny MeaningGraph; jeszcze później proste odpowiedzi na podane jawnie fakty.
- Dane jako moduły języka, kompilowane do małych niemutowalnych struktur wykonawczych; Go nie kopiuje monolitu JS.

## Jak sprawdzamy postęp

**Ręcznie.** Użytkownik daje słowo lub zdanie, kierunek i ewentualne oczekiwania. Po zbudowaniu funkcjonalności uruchamiamy Go CLI, pokazujemy rzeczywisty wynik i poprawiamy mechanizm. Dopóki użytkownik nie dostarczy wejścia, nie ma wymogu generowania przykładowych tłumaczeń.

Nie istnieje obligatoryjny *numerical accuracy*, zestaw goldów, testów gramatycznych, automatyczny runner, baseline ze starej implementacji ani konieczna liczba zdań przed merge. Ocenę działania zapisujemy w issue krótkim komentarzem, nie systemem metryk. CI sprawdza **tylko budowanie i start programu**.

## Uporządkowanie informacji

`ConceptGraph` opisuje znaczenia. `LexicalGraph` rozróżnia słowo, jego sensy i język. `ValencyGraph` określa argumenty sensu. `ConstructionGraph` definiuje, jak w danym języku zbudować NP/VP/PP i całe zdanie. `MorphologyGraph/ProcedureGraph` generują powierzchnię. `MeaningGraph` opisuje konkretną wypowiedź. `FactSet`, jeśli dojdzie później, opisuje sytuację — nie jest ontologią.

`Concept → Sense/Valency → Construction → Generator → Surface` jest skrótem przepływu, **nie** sztywnym drzewem obiektów ani koniecznym wyrażeniem w C#-owym stylu `ROCK.VERB...`.

## Granice wersji

- Nie przewidujemy swobodnej konwersacji porównywalnej z dużym LLM; model może wykonywać tylko zaprojektowane, licencjonowane konstrukcje i operacje.
- Brak automatycznej nauki, trwałej pamięci, zewnętrznych faktów, dużej infrastruktury i wymagania 10×/100× wydajności.
- Nie oznaczamy brakującego konceptu jako „błędne tłumaczenie” bez rzeczywistego wywołania.
- Brak dedukcji znaczenia na podstawie identycznej nazwy słowa w dwóch językach.
- Nie dopisujemy fałszywych faktów podczas tworzenia pytań, negacji i hipotez.

## Zakończenie v0

Uruchamialny Go CLI z jawnie zadeklarowanym zakresem PL/EN, później około 1000 konceptów (liczenie osobno od zdolności zdaniowych). Poprawność pierwszych przypadków oceniana **ręcznie przez właściciela**, nie przez automatyczne testy. Zakończenie EPIC oznacza sprawdzone pliki/funkcje i zrozumiały status możliwości, nie określony procent zdanych zdań.

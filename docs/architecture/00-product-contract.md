# 00 — Cel produktu i definicja gotowości

**Status:** PROPOSED / G00-02. **Produkt:** golingua, jeden proces Go, bez internetowej zależności runtime. **Źródło migracji:** [pinned flexLingua](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92). **Zadania:** [G00 #1](https://github.com/clickbytes/golingua/issues/1), [G05 #6](https://github.com/clickbytes/golingua/issues/6), [G06 #7](https://github.com/clickbytes/golingua/issues/7), [G07 #8](https://github.com/clickbytes/golingua/issues/8), [G08 #9](https://github.com/clickbytes/golingua/issues/9).

## Misja, nie marketing

Zbudować eksperymentalny, **bardzo lekki, ograniczony i audytowalny** silnik rozumienia, translacji oraz odpowiedzi PL↔EN: `text → analyses → typed MeaningGraph → licensed response/target plan → text`. W odróżnieniu od LLM generującego tokeny probabilistycznie, ten wariant ma wykorzystywać jawne koncepty, reguły, konstrukcje, walencję, ograniczenia i dowody. Porównanie szybkości jest hipotezą, nie obietnicą 10×/100×.

Wersja v0 ma być kompletna **w zadeklarowanym ograniczonym podzbiorze**, a nie jako język naturalny bez ograniczeń. Brak dowodu daje `unknown`/`incomplete`/`unsupported` zamiast zgadywanego poprawnie brzmiącego zdania.

## Zakres produktu v0

- Dwa niezależne **pakiety PL i EN** z profilami cech i własnymi konstrukcjami. Brak parsera na parę języków; semantyczna reprezentacja pośrednia jest jedna.
- 1 000 **unikalnych, zatwierdzonych ConceptID z poprawnym Sense→Concept binding w PL i EN** jako cel leksykalno-semantyczny, nie „1000 słów/odmian/testów”, nie 1000 zdań i nie certyfikat A1/CEFR. Nie liczyć abstrakcyjnych kategorii, operatorów, synonimów ani wariantów form jako kolejnych konceptów.
- Proceduralna morfologia: klasa, formy diagnostyczne, dopuszczone alternacje, wyjątki, wszystkie możliwe inverse readings; bounded finite table jako pierwszy backend.
- First-class `ConstructionFamily/Variant` i osobny `ValencyFrame` per sense: legalność pytania, negacji, strony, szyku i rekcji zatwierdzana przed realizacją.
- Typowany MeaningGraph: stany, zdarzenia, referenty, role, polaryzacja, pytania, podstawowe operatory, zasięg, czas i kwantyfikacja tylko w rzeczywiście zweryfikowanych subsetach.
- Minimalny request-scoped `FactSet` z proof-based yes/no i WH oraz ContentPlanner wybierający treść i licencjonowaną konstrukcję. Brak zewnętrznej faktografii.
- Trace audytowy: źródłowe span/case/SenseID/ConstructionID/ProcedureID + `informationLoss` i kontrolowana odmowa.
- Jeden CLI Go i czytelny JSON, bez przymusu serwera, GUI ani procesów pomocniczych.

## Poza zakresem v0

- **Uczenie się** form, wyjątków, konceptów lub faktów przez runtime; samodzielne dodawanie reguł lub autofaktyzacja.
- Trwała pamięć użytkownika, import historii rozmów, zgadywanie referenta po podobnym słowie.
- Pełne NLP, wiedza ogólna, swobodny kreatywny dialog, internetowy retrieval, emocjonalny/personality LM lub gwarancja jakości na nieznanej domenie.
- Wymóg „property navigation” typu `ROCK.VERB.MORPH.PAST()`; identyfikatory są typowane i odwołania przez normalne Go API.
- Z góry wymagane SQLite, FST, kompilowanie Go jako `.so`, własny format binarny i rozbudowany rozproszony system usług.

## Stan audytowanego źródła, nie gotowość targetu

Na `flexLingua/main@79c4a958` policzono **1002** wpisy CORE (w tym **51** abstrakcyjnych kategorii), **943** różne ConceptID z mapowaniami leksykalnymi w obu językach, **1005** wpisów leksykonu PL i **1006** EN, **465/462** wiązań morfologicznych, **27/13** paradygmatów PL/EN oraz **51/51** ramek gramatycznych. To statyczne liczby source, nie dowód 943 poprawnie zrealizowanych zdań ani 943 zatwierdzonych *equivalent* bindingów w nowym kontrakcie.

Przy zachowaniu źródłowych 943 zatwierdzonych odpowiedników potrzebnych jest **57 dodatkowych dwujęzycznych ConceptID**, aby osiągnąć 1000. Spośród 1002 źródłowych ID: 51 to kategorie, z pozostałych 951 osiem nie ma bezpośredniego mapowania PL/EN. Zatem gdyby wszystkie osiem nadawało się do zatwierdzonego mapowania, co najmniej **49 nowych** właściwych konceptów musiałoby zostać dodanych. To dolne ograniczenie arytmetyczne, a nie lista 49 zatwierdzonych konceptów.

## Siedem niezależnych wymiarów pokrycia

| Miara | Licznik i minimalny warunek |
| --- | --- |
| catalog | ConceptID istnieje, ma poprawny typ i relacje |
| bilingual sense | ten sam ConceptID ma approved PL SenseBinding i approved EN SenseBinding, zgodne znaczeniowo |
| morphology | wymagane formy zostały potwierdzone i generator/analyzer odtwarzają wszystkie licencjonowane readings |
| construction | zatwierdzony Sense ma ValencyBinding i przynajmniej jeden plan konstrukcji dla konkretnego kierunku |
| sentence | jawna, niezależnie zatwierdzona para sentence/MeaningGraph/diagnostics przeszła test |
| generation | docelowy surface przekazuje role, scope i wszystkie wymuszane cechy albo jawnie raportuje stratę/refusal |
| dialogue | na podanym FactSet odpowiedź jest dowodliwa, używa licencjonowanej konstrukcji i nie wymyśla faktów |

Denominatory są osobne. `1000 bilingual concepts` nie oznacza 1000 sentence-ready/dialoque-ready. Raport musi pokazywać stan wszystkich siedmiu, także wartość zero.

## Minimalne zakończenie v0 — kontrakt obserwowalny

1. Rozpoznaj w PL i EN konkretny, wcześniej zatwierdzony zestaw zdań, a dla każdej wypowiedzi zachowaj właściwy MeaningGraph i rolę agenta/tematu.
2. Zrealizuj to znaczenie w drugim języku wyłącznie przez `approved SenseBinding + licensed ConstructionPlan + MorphGenerator`.
3. Zwróć różne legalne odczytania jako alternatywy; ograniczenie eksploracji nie może stać się dowodem unikalności.
4. Dla pytania z jawnie podanymi faktami zwróć answer/clarify/unknown z `Proof`; brak dowodu ≠ fałsz.
5. Pokaż odmowę dla unlicensed construction i formy zablokowanej przez wyjątek.
6. CI na kodzie Go i istniejących goldach, testy nowych kompozycji, benchmark na tym samym hoście; bez fałszywych „x100”.

Gdy v0 nie spełnia docelowego pokrycia, raportuj `NOT MET` i nie podpisuj release jako kompletnego.

# Architektura golingua — Go od zera, ocena ręczna

**Status architektury: runtime zdań PL/EN obsługuje pięć ograniczonych szlaków: `seeing`, `hearing`, `looking`, `reading` i `receiving`.** `seeing`, `hearing`, `looking` i `reading` mają jawne formy teraźniejsze 3. osoby liczby pojedynczej i mnogiej; `receiving` ma angielskie `got` i jawne polskie formy czasu przeszłego 3. osoby liczby pojedynczej i mnogiej. `reading` używa angielskiego `read/reads` w teraźniejszości i angielskiego simple past `read`; homograficzne mnogie `read` pozostaje niejednoznaczne czasowo. Polski imperfective `czytać` ma jawne formy teraźniejsze `czyta/czytają` oraz 15 jawnych komórek przeszłych dla osoby, liczby i rodzaju. `seeing`, `hearing` i `looking` obsługują również negację teraźniejszą w 3. osobie liczby pojedynczej i mnogiej: EN `does/do not` oraz PL `nie` z jawnymi formami czasownika. `looking` zachowuje angielskie `look at` i polskie `patrzeć na` z biernikiem, także w osobnej rodzinie negatywnej. Dla obsługiwanych dodatnich i negatywnych szlaków angielskie konstrukcje zachowują zatwierdzone warianty `a`, `an` i `the` dla podmiotu i dopełnienia/celu, gdy polskie źródło nie rozstrzyga określoności; wybór allomorfu wynika z jawnej właściwości leksemu `indefinite_article=a/an`, a analiza po `a/an` odrzuca liczbę mnogą. Na `reading` brak określoności pozostaje jawnym założeniem kandydata; polskie źródło może zwrócić kilka angielskich wariantów rodzajnika. Jawne obiekty masowe w `receiving` zachowują wariant bez rodzajnika. Indeks ma 117 zatwierdzonych wiązań konstrukcji z sensem/ramą (PL 8, EN 109), w tym 48 wariantów rodzajnika na negatywnych trasach EN. Nierozstrzygnięte sensy i brakujące formy nadal zwracają status braku obsługi, niejednoznaczności lub nierozstrzygnięcia. PL→EN dla polskiej negacji zachowuje kandydatury z tekstem, ale pozostaje `ambiguous`, gdy określoność wymagana przez angielskie rodzajniki nie wynika ze źródła; `reading` jawnie zachowuje założenia określoności; PL→EN może być `ambiguous` z powodu zatwierdzonych wariantów rodzajnika. Istnieją rzeczywiste obserwacje CLI dla przykładów zdaniowych; ocena agenta na delegowanych wejściach i akceptacja właściciela pozostają osobnymi statusami.

Katalog Go można zapisać jako walidowany snapshot poleceniem `golingua snapshot --apply --store <katalog> --json`. `--preview` pokazuje skrót i liczności przed aktywacją. Snapshot obejmuje CORE oraz leksemy, sensy i bindingi PL/EN z ich obecnymi statusami i pochodzeniem; morfologia, walencja i plany konstrukcji są kompilowane z zaufanego kodu Go. `translate --runtime-config <path>` wymaga aktywnego snapshotu i weryfikuje dokładne trasy przed wykonaniem.

## Fundamentalna decyzja (2026-10-20)

Właściciel zrezygnował z dotychczasowego podejścia do porównywania ze starym JavaScriptem, z obowiązkowych goldów i automatycznej macierzy. **Pisze się nowy silnik Go od zera**: źródło flexLingua może pomóc zrozumieć domenę i później dostarczyć opcjonalne dane, ale nie narzuca zachowania ani planu kodu. Efekty tłumaczeń i odmiany sprawdzamy na rzeczywistych wejściach właściciela przez [procedurę](../manual-review.md); od 2026-10-23 właściciel deleguje agentowi autonomiczną ocenę i poprawianie, bez utożsamiania tego z jego własną akceptacją. Nie dopisuj `*_test.go`, nowych test runnerów lub fixture trees.

Minimalne CI: formatowanie, vet, build, start CLI. Nie używamy CI jako językowego wskaźnika jakości.

## Kanoniczne rozdziały

1. [00 — cel i ograniczenia](00-product-contract.md)
2. [01 — model i zależności grafów](01-graph-layers.md)
3. [IDs — stable symbols i dense runtime IDs](ids.md)
4. [ConceptGraph — typed concepts and relations](concepts.md)
5. [Runtime graph — adjacency and bounded traversal](runtime-graph.md)
6. [Meaning IR v1](meaning-ir.md)
7. [Semantic operators and asserted projection](semantic-operators.md)
8. [Request scoped references and equivalence](references.md)
9. [Lexical records and semantic bindings](lexicon-model.md)
10. [Language-local feature profiles and unification](feature-profiles.md)
11. [02 — przepływ analizy/realizacji](02-runtime-pipeline.md)
12. [03 — format źródła i runtime](03-data-decisions.md)
13. [04 — wyniki, diagnostyki i limity](04-results-and-limits.md) — kontrakt informacyjny, bez bramek testowych
14. [05 — źródła danych i manifesty](05-data-sources.md) — otoczka źródeł, bez importera i parity gate
15. [06 — walencja i ConstructionGraph](06-constructions-and-valency.md)
16. [07 — generatory, paradygmaty i wyjątki](07-morphology-procedures-exceptions.md)
   - [Kontrakt formalnych wyjątków morfologicznych](morph-exceptions.md)
   - [Generator i finite relation table](morph-backends.md)
17. [08 — MeaningGraph, fakty i rozumowanie](08-meaning-facts-reasoning.md)
    - [Request-scoped world facts](world-facts.md)
    - [Bounded request reasoning](reasoning.md)
    - [Typed question execution](questions.md)
    - [Content planning](content-planning.md)
    - [Dialogue response realization](dialogue-realization.md)
18. [09 — kontrakt pluginów językowych](09-language-plugin-contract.md)
19. [10 — dane, słowniki i źródła](10-import-provenance.md)
20. [11 — minimalne kontrole i ręczny review](11-manual-review.md)
21. [12 — decyzje/ryzyka](12-decisions-and-risks.md)
22. [13 — EPIC-i i kolejność](13-issue-traceability.md)
23. [14 — przykładowe przebiegi projektowe](14-worked-vertical-slices.md)

Dokumenty 04 i 05 opisują typy/statusy oraz manifesty/pochodzenie; nie są bramkami do testów, goldów ani parity. G00-03 dodaje `04-results-and-limits.md`, a G00-04 dodaje `05-data-sources.md`. Nie tworzymy dokumentu `05-parity-policy.md`, generatorów raportów ani oracli.

## Co jest normatywne

- Oddzielne ConceptDefinition, MorphLexeme, LexicalSense, SemanticBinding, ValencyFrame, ConstructionFamily/Variant/Operator, Procedure, Paradigm, MeaningGraph i request-scoped FactSet.
- Oddzielna wiedza semantyczna i powierzchnia języka.
- Proceduralne formy/zdania z legalną obsługą wyjątków.
- Typowane błędy i ograniczenie pracy runtime.
- Języki jako moduły danych, bez kopiowania słownika do CORE.
- Jedna kolejka zadań: najwyżej jeden aktywny epic/task, reszta QUEUED, jeden agent i jeden implementacyjny PR.

**Nie jest normatywne:** obligatoryjne API `ROCK.VERB.MORPH.PAST`; odtwarzanie dawnego JS; liczba testów, goldów i wymaganych przykładów; od razu FST/SQL; uczenie i trwała pamięć.

Proponowane nazwy struktur Go są szkicami. Implementator ma zachować **kontrakt znaczeniowy i rozdział odpowiedzialności**, ale może uzasadnić inne nazewnictwo/szczegóły w aktywnym issue.

# Architektura golingua — Go od zera, ocena ręczna

**Status architektury: techniczne komponenty zaimplementowane, językowy runtime nieaktywny.** Istnieją CLI/bootstrap, typowane identyfikatory, 1000-komponentowy ConceptGraph, Meaning IR, graph indexes, parser/realizer/planery jako biblioteki, małe seedy morfologii i niezależne leksykony PL/EN. Osiem dokładnych konstrukcyjnych tras PL/EN ma zatwierdzone bindingi, a 989 EN Sense→CORE links przeszło delegowany przegląd w dwudziestu partiach; brak zatwierdzonych PL lexical links i skonfigurowanych providerów językowych, więc nie ma zatwierdzonych par PL+EN i `translate` poprawnie zwraca `unsupported`. Implementacja technicznego parsera/realizera nie jest równoznaczna z możliwością generowania polskiego lub angielskiego zdania.

## Fundamentalna decyzja (2026-10-20)

Właściciel zrezygnował z dotychczasowego podejścia do porównywania ze starym JavaScriptem, z obowiązkowych goldów i automatycznej macierzy. **Pisze się nowy silnik Go od zera**: źródło flexLingua może pomóc zrozumieć domenę i później dostarczyć opcjonalne dane, ale nie narzuca zachowania ani planu kodu. Przez pierwsze etapy efekty tłumaczeń i odmiany sprawdzane są **wyłącznie ręcznie**, na wejściach właściciela, przez [procedurę](../manual-review.md). Nie dopisuj `*_test.go`, nowych test runnerów lub fixture trees.

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

- Oddzielne ConceptDefinition, MorphLexeme, LexicalSense, SemanticBinding, ValencyFrame, ConstructionFamily/Variant/Operator, Procedure, Paradigm, MeaningGraph i ewentualny późniejszy FactSet.
- Oddzielna wiedza semantyczna i powierzchnia języka.
- Proceduralne formy/zdania z legalną obsługą wyjątków.
- Typowane błędy i ograniczenie pracy runtime.
- Języki jako moduły danych, bez kopiowania słownika do CORE.
- Jedna kolejka zadań (G00 ACTIVE, następne QUEUED), jeden agent i jeden PR.

**Nie jest normatywne:** obligatoryjne API `ROCK.VERB.MORPH.PAST`; odtwarzanie dawnego JS; liczba testów, goldów i wymaganych przykładów; od razu FST/SQL; uczenie i trwała pamięć.

Proponowane nazwy struktur Go są szkicami. Implementator ma zachować **kontrakt znaczeniowy i rozdział odpowiedzialności**, ale może uzasadnić inne nazewnictwo/szczegóły w aktywnym issue.

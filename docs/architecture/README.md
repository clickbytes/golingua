# Architektura golingua — Go od zera, ocena ręczna

**Status architektury: częściowo wdrożone typy danych, nie wdrożony silnik.** Bieżący runtime ma CLI/bootstrap, typowane statusy, metadane źródeł, stable symbols, snapshot-local table NodeID, kompilowany ConceptGraph, bounded adjacency/traversal, versioned Meaning IR i fail-closed asserted projection. Nie ma jeszcze leksykonu, parsera ani realizera. Wyjście `translate` pozostaje unsupported dopóki realizer/parsing nie powstanie.

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
9. [02 — przepływ analizy/realizacji](02-runtime-pipeline.md)
10. [03 — format źródła i runtime](03-data-decisions.md)
11. [04 — wyniki, diagnostyki i limity](04-results-and-limits.md) — kontrakt informacyjny, bez bramek testowych
12. [05 — źródła danych i manifesty](05-data-sources.md) — otoczka źródeł, bez importera i parity gate
13. [06 — walencja i ConstructionGraph](06-constructions-and-valency.md)
14. [07 — generatory, paradygmaty i wyjątki](07-morphology-procedures-exceptions.md)
15. [08 — MeaningGraph, fakty i rozumowanie](08-meaning-facts-reasoning.md)
16. [09 — kontrakt pluginów językowych](09-language-plugin-contract.md)
17. [10 — dane, słowniki i źródła](10-import-provenance.md)
18. [11 — minimalne kontrole i ręczny review](11-manual-review.md)
19. [12 — decyzje/ryzyka](12-decisions-and-risks.md)
20. [13 — EPIC-i i kolejność](13-issue-traceability.md)
21. [14 — przykładowe przebiegi projektowe](14-worked-vertical-slices.md)

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

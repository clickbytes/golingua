# Architektura golingua — Go od zera, ocena ręczna

**Status architektury: PROPOSED, nie wdrożony silnik.** Bieżący Go ma tylko CLI/bootstrap i minimalny CI. Wyjście `translate` pozostaje unsupported dopóki realizer/parsing nie powstanie.

## Fundamentalna decyzja (2026-10-20)

Właściciel zrezygnował z dotychczasowego podejścia do porównywania ze starym JavaScriptem, z obowiązkowych goldów i automatycznej macierzy. **Pisze się nowy silnik Go od zera**: źródło flexLingua może pomóc zrozumieć domenę i później dostarczyć opcjonalne dane, ale nie narzuca zachowania ani planu kodu. Przez pierwsze etapy efekty tłumaczeń i odmiany sprawdzane są **wyłącznie ręcznie**, na wejściach właściciela, przez [procedurę](../manual-review.md). Nie dopisuj `*_test.go`, nowych test runnerów lub fixture trees.

Minimalne CI: formatowanie, vet, build, start CLI. Nie używamy CI jako językowego wskaźnika jakości.

## Kanoniczne rozdziały

1. [00 — cel i ograniczenia](00-product-contract.md)
2. [01 — model i zależności grafów](01-graph-layers.md)
3. [02 — przepływ analizy/realizacji](02-runtime-pipeline.md)
4. [03 — format źródła i runtime](03-data-decisions.md)
5. [04 — wyniki, diagnostyki i limity](04-results-and-limits.md) — kontrakt informacyjny, bez bramek testowych
6. [06 — walencja i ConstructionGraph](06-constructions-and-valency.md)
7. [07 — generatory, paradygmaty i wyjątki](07-morphology-procedures-exceptions.md)
8. [08 — MeaningGraph, fakty i rozumowanie](08-meaning-facts-reasoning.md)
9. [09 — kontrakt pluginów językowych](09-language-plugin-contract.md)
10. [10 — dane, słowniki i źródła](10-import-provenance.md)
11. [11 — minimalne kontrole i ręczny review](11-manual-review.md)
12. [12 — decyzje/ryzyka](12-decisions-and-risks.md)
13. [13 — EPIC-i i kolejność](13-issue-traceability.md)
14. [14 — przykładowe przebiegi projektowe](14-worked-vertical-slices.md)

Nie tworzymy dokumentów 04/05 jako bramek do testów, goldów ani parity. G00-03 udostępnia niebramkujący `04-results-and-limits.md` opisujący statusy i limity kodu. G00-04 opisze prosty sposób opisywania źródeł i danych; bez generatorów raportów i oracli.

## Co jest normatywne

- Oddzielne ConceptDefinition, MorphLexeme, LexicalSense, SemanticBinding, ValencyFrame, ConstructionFamily/Variant/Operator, Procedure, Paradigm, MeaningGraph i ewentualny późniejszy FactSet.
- Oddzielna wiedza semantyczna i powierzchnia języka.
- Proceduralne formy/zdania z legalną obsługą wyjątków.
- Typowane błędy i ograniczenie pracy runtime.
- Języki jako moduły danych, bez kopiowania słownika do CORE.
- Jedna kolejka zadań (G00 ACTIVE, następne QUEUED), jeden agent i jeden PR.

**Nie jest normatywne:** obligatoryjne API `ROCK.VERB.MORPH.PAST`; odtwarzanie dawnego JS; liczba testów, goldów i wymaganych przykładów; od razu FST/SQL; uczenie i trwała pamięć.

Proponowane nazwy struktur Go są szkicami. Implementator ma zachować **kontrakt znaczeniowy i rozdział odpowiedzialności**, ale może uzasadnić inne nazewnictwo/szczegóły w aktywnym issue.

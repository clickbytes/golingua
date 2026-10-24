# Architektura golingua — Go od zera, ocena ręczna

Aktualizacja 2026-10-24: [LIGHT-F](../work/light-lm/README.md) utrwala czas odniesienia i wielość uczestników w samodzielnym I. [Kontrakt atrybutów](meaning-attributes.md) obowiązuje w analizie, realizacji, porównaniach i projekcji. Nowsza decyzja właściciela dopuszcza ograniczone lokalne testy `tests/` opisane w [macierzy](../../tests/README.md); wcześniejszy zakaz testów poniżej jest historyczny. CI nadal sprawdza technikę, a akceptacja językowa właściciela pozostaje osobna.

**Status architektury: runtime zdań PL/EN obsługuje pięć ograniczonych szlaków zdarzeń (`seeing`, `hearing`, `looking`, `reading`, `receiving`) oraz ograniczoną dodatnią konstrukcję `black` + podmiot `seeing`.** Konstrukcja złożona wymaga teraźniejszego zdarzenia, dwóch referentów w liczbie pojedynczej i nieczasowanego stanu `black` przypisanego do referenta podmiotu; nie rozstrzyga tożsamości ani płci. Pozostałe szlaki zachowują ograniczenia opisane poniżej: formy czasu i liczby są dostępne tylko w jawnie zaimplementowanych komórkach, a nierozstrzygnięte rodzajniki, czas mnogiego angielskiego `read` i brakujące formy pozostają widoczne jako warianty lub odmowa.

Polski imperfective `czytać` ma jawne formy teraźniejsze `czyta/czytają` oraz 15 komórek przeszłych dla osoby, liczby i rodzaju. Angielski `look` korzysta obecnie ze wspólnej regularnej procedury odmiany czasownika zamiast osobnego zestawu ręcznie zapisanych principal parts; zachowuje `look`, `looks`, `looked` (past i past participle) oraz `looking`. Polskie `patrzeć` i angielskie `look at` pozostają związane z trasą dopełnienia, również w rodzinie negatywnej. `seeing`, `hearing` i `looking` obsługują ograniczoną negację teraźniejszą (`does/do not` po angielsku, `nie` po polsku); `receiving` ma angielskie `got` oraz jawne polskie formy przeszłe 3. osoby. Czytanie zachowuje angielskie `read/reads` i homograficzne mnogie `read` jako niejednoznaczne czasowo. Angielskie warianty `a`, `an` i `the`, polskie założenia oraz jawne straty są ograniczone do zatwierdzonych tras i nie są domyślnie rozstrzygane.

Bieżący `catalog --summary --json` wykazuje 135 zatwierdzonych wiązań konstrukcji z sensem/ramą (PL 12, EN 123), w tym siedem dodanych przez #327. Szczegóły złożonej trasy i jej granice opisuje [kontrakt #327](../work/light-lm/327-contract.md). Istnieją obserwacje CLI dla ograniczonych przykładów; ocena agenta i akceptacja właściciela pozostają osobnymi statusami.

Katalog Go można zapisać jako walidowany snapshot poleceniem `golingua snapshot --apply --store <katalog> --json`. `--preview` pokazuje skrót i liczności przed aktywacją. Snapshot obejmuje CORE oraz leksemy, sensy i bindingi PL/EN z ich obecnymi statusami i pochodzeniem; morfologia, walencja i plany konstrukcji są kompilowane z zaufanego kodu Go. `translate --runtime-config <path>` wymaga aktywnego snapshotu i weryfikuje dokładne trasy przed wykonaniem.

## Fundamentalna decyzja (2026-10-20)

Właściciel zrezygnował z dotychczasowego podejścia do porównywania ze starym JavaScriptem, z obowiązkowych goldów i automatycznej macierzy. **Pisze się nowy silnik Go od zera**: źródło flexLingua może pomóc zrozumieć domenę i później dostarczyć opcjonalne dane, ale nie narzuca zachowania ani planu kodu. Efekty tłumaczeń i odmiany sprawdzamy na rzeczywistych wejściach właściciela przez [procedurę](../manual-review.md); ograniczone lokalne kontrole opisuje [macierz PL/EN](../../tests/README.md), a agentowa ocena nie jest akceptacją właściciela.

Minimalne CI: formatowanie, vet, build, start CLI. Nie używamy CI jako językowego wskaźnika jakości.

## Kanoniczne rozdziały

1. [00 — cel i ograniczenia](00-product-contract.md)
2. [01 — model i zależności grafów](01-graph-layers.md)
3. [IDs — stable symbols i dense runtime IDs](ids.md)
4. [ConceptGraph — typed concepts and relations](concepts.md)
5. [Runtime graph — adjacency and bounded traversal](runtime-graph.md)
6. [Meaning IR v1](meaning-ir.md)
   - [Czas odniesienia i wielość referenta](meaning-attributes.md)
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

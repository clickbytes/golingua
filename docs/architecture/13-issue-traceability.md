# 13 — Mapa EPIC/issue → kontrakt → dowód

**Status:** plan w GitHub, bez twierdzenia że poniższe features już istnieją. **Zero równoległych agentów**. Stan rozpoczęcia G00-02: jeden ACTIVE EPIC G00, task G00-01 zmergowany; pozostałe QUEUED. Datę i aktualność sprawdza się w issues, nie w tym statycznym pliku.

| Kolejność | EPIC | Zakres issues | Kontrakt do przeczytania | Co jest dowodem DoD |
| --- | --- | --- | --- | --- |
| 0 | [G00 #1](https://github.com/clickbytes/golingua/issues/1) | G00-01…06 (#11–16) | 00,01,02,03,11,12 + nowe G00-03/04 docs | Go build/CI, contract docs, status and manifest checks |
| 1 | [G01 #2](https://github.com/clickbytes/golingua/issues/2) | G01-01…06 (#18–23) | 01,02,08,12 | IDs, ontology, scope, referents tests, cross-graph semantics |
| 2 | [G02 #3](https://github.com/clickbytes/golingua/issues/3) | G02-01…07 (#24–30) | 01,03,07,09,10 | Lexeme without CORE, morphology generate/inverse gold, exceptions |
| 3 | [G03 #4](https://github.com/clickbytes/golingua/issues/4) | G03-01…09 (#31–37, [#74](https://github.com/clickbytes/golingua/issues/74), [#75](https://github.com/clickbytes/golingua/issues/75)) | 01,06,07,09,14 | shared family, per-sense valency, operators, NP/VP/PP and specialised phrase plans |
| 4 | [G04 #5](https://github.com/clickbytes/golingua/issues/5) | G04-01…07 (#38–43, [#76](https://github.com/clickbytes/golingua/issues/76)) | 03,07,09,10,11 | staged import, review, immutable pack, typed EvidenceGraph |
| 5 | [G05 #6](https://github.com/clickbytes/golingua/issues/6) | G05-01…08 (#44–51) | 02,06,07,08,09,11,14 | source→meaning→target both directions, parser alternatives, exact golds |
| 6 | [G06 #7](https://github.com/clickbytes/golingua/issues/7) | G06-01…06 (#52–57) | 00,02,08,11,14 | request FactSet, ProofSteps, content planning, grounded answer |
| 7 | [G07 #8](https://github.com/clickbytes/golingua/issues/8) | G07-01…07 (#58–64) | 00,03,07,09,10,11 | independently reviewed >=1000 bilingual ConceptID; per-capability coverage |
| 8 | [G08 #9](https://github.com/clickbytes/golingua/issues/9) | G08-01…06 (#65–70) | 00,02,03,11,12 | 593 legacy case trace, held-out, fair benchmark, CI/release |
| 9 | [G09 #10](https://github.com/clickbytes/golingua/issues/10) | G09-01…03 (#71–73) | 08,12, separate docs/future to be written | **DEFERRED** plan memory/learning/personality; no v0 runtime |

W każdej kwestii „exact path vs issue”: pliki ścieżek są **docelowe**, aż do implementacji. Doprecyzowania w tych docs są kontraktem przed kodem, ale nazwy exported Go symboli mogą wymagać małej korekty zatwierdzonej w issue. W razie konfliktu z modelem semantycznym nie wprowadzać odmiennych interfejsów tylko w jednym module.

## Przegląd backlogu na G00-02 (stan i poprawki)

**Mocne strony pierwotnych 62 zadań:** 62/62 mają kroki, 62/62 DoD i wyraźne OUT, około 319 konkretnych ścieżek, 10 niezależnych faz. To wystarczające do nawigowania PR po kodzie po ustanowieniu kontraktu, lecz **nie wystarcza do odtwarzania semantyki bez docs**.

**Po audycie: 65 planowanych zadań (pierwotne 62 + trzy zadania naprawiające braki), 10 EPIC-ów.** Wszystkie trzy nowe tasks są QUEUED, a G00 pozostaje ACTIVE.

**Potwierdzone uzupełnienia zakresu:**
1. Neutralna reprezentacja cech i direction DAG — doc 01+07; poprawić [G02-02 #25](https://github.com/clickbytes/golingua/issues/25): obecny plan umieszcza współdzielone typy w `internal/morph`, powinny znaleźć się w neutralnym `internal/features`.
2. First-class phrase subconstruction NP/VP/PP — doc 06, [G03-08 #74](https://github.com/clickbytes/golingua/issues/74) (QUEUED).
3. Liczebniki, czas, MWE/idiom as licensed procedures — doc 06, [G03-09 #75](https://github.com/clickbytes/golingua/issues/75) (QUEUED).
4. EvidenceGraph jako queryable provenance/capability index — docs 01+11, [G04-07 #76](https://github.com/clickbytes/golingua/issues/76) (QUEUED).
5. Bez required dotted/fluent API — doc 12, korekta [G03-02 #32](https://github.com/clickbytes/golingua/issues/32) i README. Zastępuje je typed SymbolID/ConstructionID/ProcedureID.
6. Oddzielne osie statusu i completeness — doc 02; ostateczny wire format w G00-03.
7. Approved, semantically exact bilingual concept count i gap to 1000 — doc 00/10/11; G07-01 wybiera definicję i katalog.

## Katalog warunków zależności

- G00-01 already merged; G00-02 docs-only; G00-03 statuses; G00-04 source/golds; G00-05 workflow; G00-06 quality.
- G01 produces identifiers, pure CORE/Meaning and graph traversal **before** G02 relies on them.
- G02 introduces lexical senses and morph generation/analysis **before** G03 compiles construction plans.
- G03-01 valency before G03-02/03 families and variants, G03-04/05 operators and compiler, G03-06 language registry, G03-07 seed, G03-08 phrase plans, G03-09 specialised policies. G03-08/09 do not independently implement sentence parsing, which belongs to G05.
- G04 imports candidates/compiles packs from already fixed graph and construction schemas. G04-07 evidence-backed activation after staging and compiler. Runtime source remains read-only.
- G05 uses language packages from G04 to realize and analyze, then tests both directions; G06 reasons and plans from explicit facts; G07 scales coverage; G08 challenges generalization and benchmarks.
- G09 planning only, no v0 memory, learning or broad personality.

## Bramka każdego tasku

1. Fresh main SHA + exact issue + owner + no open unrelated PR.
2. Przeczytane właściwe docs i source anchors, recorded expected/provenance/gold approval.
3. Kod tylko własnego tasku, source-independent negative tests, no alternate legacy path.
4. Local command proof only if actually run, CI PASS on head and merge SHA, review diff/semantic impact.
5. Epic checklist + issue handoff updated, następny task starts only after merge, nie równolegle.

## Status audytu kontraktowego

Ten plik jest mapą projektu. **Nie certyfikuje** że kontrakty mogą być już poprawnie kompilowane, importowane lub tłumaczone; to wymaga implementacji i testów kolejnych issue. Jeżeli przyszły issue ujawnia niewspierane wymaganie, zwraca blocker zamiast dostosowywać oracle do implementacji.

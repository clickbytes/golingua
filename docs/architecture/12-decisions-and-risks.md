# 12 — Decyzje architektoniczne, rozstrzygnięte kolizje i świadome odroczenia

**Status:** dokument decyzji **projektowych**, nie relacja z już napisanego runtime. Granice v0 pozostają proponowane do implementacji poszczególnymi issues. Stan golingua na bazowym commicie G00-02 to tylko Go bootstrap.

## Decyzje wiążące kolejne PR-y, o ile właściciel ich nie zmieni

| ID | Decyzja | Dlaczego / ograniczenie | Implementacja |
| --- | --- | --- | --- |
| D-01 | Go-first, jeden proces | typowanie i profilowanie zamiast mechanicznej translacji JS | G00–G08 |
| D-02 | Languages remain data plugins | bez Go `plugin.Open()`, osobne profil PL/EN | G03-06 |
| D-03 | CORE nie posiada słów | Concept ≠ Lexeme ≠ Sense; brak rekcji w ConceptGraph | G01/G02 |
| D-04 | First-class ConstructionFamily | rodzina/variant/plan są osobnymi modelami, nie `grammar.frames` per event | G03 |
| D-05 | Walencja per SenseID | role semantyczne i funkcje argumentów nie są tym samym | G03-01 |
| D-06 | NP/VP/PP też są konstrukcjami | kompozycyjne podstruktury, bez per-verb sentence tables | G03-08 |
| D-07 | Operatory gramatyczne są typowane | nie dowolny kod ani nielegalny iloczyn wszystkich wariantów | G03-04/05 |
| D-08 | MorphLexeme bez CORE | analityczna odmiana importu nie wymaga zatwierdzonej semantyki | G02-01 |
| D-09 | Paradigms + principal parts + formal exceptions | regularne formy proceduralne; rzeczywiste nieregularności jawne | G02-03/04 |
| D-10 | Generate↔Analyze bounded, ALL readings | błędne „first reading” nie może udawać unikalności | G02-05/G05 |
| D-11 | Finite relation tables na start | FST/lazy wymaga benchmarku, nie domyślnie lepsze | G02/G08 |
| D-12 | Stable SymbolID, runtime dense ID lokalne dla pack | zero stable runtime uint32 across versions | G01-01 |
| D-13 | Neutral feature mechanics w `internal/features` | dependency DAG i niezależność profili języków | korekta G02-02 |
| D-14 | Typed WorldFact oddzielnie od Concept | dog class ≠ konkretny pies w pokoju | G06-01 |
| D-15 | Proof-based reasoner, not pattern guess | unknown ≠ false, typical ≠ proof | G06-02/03 |
| D-16 | ContentPlanner oddzielnie od SurfaceRealizer | wybór treści i wybór słów niezależne | G06-04/05 |
| D-17 | Memory/learning absent in Go v0 | przyszłe interfejsy bez automutacji danych | G09 DEFERRED |
| D-18 | Provenance + approval before activation | source data ≠ trusted sense mapping; no autoapproval | G04/G07 |
| D-19 | EvidenceGraph jako typowane indeksy | każda capability ma konkretny proof/review/CI SHA | G04-07 |
| D-20 | Immutable snapshot, compiler offline | szybki hot path; brak request-time reinterpretacji JSON | G04-06 |
| D-21 | No mandatory SQLite/FST/Graph DB | decyzja storage/backend po profilowaniu | G04/G08 |
| D-22 | 1000 to approved bilingual ConceptIDs | nie 1000 leksemów/słowotworów/zdań | G07 |
| D-23 | Semantic truth > string success | role/scope/referent/time zachowane lub explicit loss | G05/G08 |
| D-24 | Brak required dotted/fluent property API | **PORZUCAMY** wymaganie `ROCK.VERB.MORPH.PAST`; wystarczy typowany Go method/ID lookup. Nie twórz node per inflection cell. | G00-02/G03-02 |
| D-25 | Jeden agent / epic / issue / PR | uniknięcie konfliktów i ułatwienie review | AGENTS + EPICs |
| D-26 | Source SHA pinned, gold independent | JS may contain bugs; no oracle derived from Go output | G00-04/G08 |
| D-27 | Input uncertainty preserves incompleteness | niepełny parser nie ma prawa zwracać confident unique | G00-03/G05-03 |
| D-28 | Go personality/persistent context poza v0 | styl i pamięć nie mogą zmienić znaczenia | G09 |

**D-24 jest rozstrzygnięciem bieżącego użytkownika:** wygodne ścieżki nawigacyjne były szkicem interfejsu, nie celem architektury. Ich porzucenie nie zmienia potrzeby first-class ConceptID/ConstructionID/ProcedureID ani proceduralnych generatorów.

## Ryzyka wysokie i ograniczenia

| Ryzyko | Przykład regresji | Ochrona |
| --- | --- | --- |
| R-01: pozorny concept coverage | 943 common lexical links reklamowane jako 943 sentence translations | siedem liczników coverage, brak inflated denominator |
| R-02: zbyt agresywne generowanie | `goed` obok zatwierdzonego `went`; błędna polska odmiana | principal parts, exceptions, independent golds, inverse parity |
| R-03: konstrukcja bez walencji | `look` dostaje obiekt bez `at`, mimo podobieństwa do `see` | sense-specific licensed ValencyFrame |
| R-04: semantics loss | znikają negacja, czas, kwantyfikacja lub relacje w dyskursie | exact MeaningGraph comparison, loss path, refusal |
| R-05: generowanie nielegalnych zdaniowych wariantów | wszystkie combinations tense×voice×question bez warunków | operator typechecking and family-specific constraints |
| R-06: fałszywy pewny odczyt | parser osiąga budżet po jednej gałęzi | typed incomplete and resource_limit distinct |
| R-07: niepoprawna identity | `dog_1` i `dog_2` połączone po ConceptID | typed local referents, proof-backed coreference |
| R-08: fikcyjna odpowiedź z ontologii | pytanie „dlaczego pies tu jest” bez WorldFacts | reasoner with explicit source FactSet and proofs |
| R-09: nieautoryzowany import | zewnętrzne formy, brak licencji, data corruption | staging, hash, review, atomic promotion |
| R-10: nadmiarowe grafy/mikroserwisy | węzeł w SQL na każdą formę przy request | immutable direct indexes, profile before DB |
| R-11: rozjazd Go packages | `core↔graph`, `construction↔morph` circular import | fixed dependency DAG, design review |
| R-12: fałszywe badania | Go szybki, bo pominął ambig/negację | compare complete semantic outcomes same workloads |

## Otwarte decyzje, których nie zamykać w ciemno

- Konkretne pola wire JSON, stale enum numbers, resource budgets i backward compatibility source — G00-03/G01/G04; przykład schematu w docs jest **pseudokontraktem**, nie automatycznie produkcyjnym schema.
- Wybór store staging (plain JSON transaction vs SQLite), fizycznego pack format, FST/lazy, indeksu CSR vs zwykła adjacency — **DEFERRED** do dedykowanych pomiarów i wymogów.
- Które realne 57 dodatkowych konceptów lub mappingów do 1000 po review — G07; liczba 57 to baseline arytmetyczny, nie zatwierdzony katalog.
- Uczciwe licencje i prawa do danych zewnętrznych — nie deklarować rozstrzygnięcia dla wszystkich dostawców bez weryfikacji konkretnego pliku/wydania.
- Liczba konstrukcji/rekursja/zakresy modalności wspierane w v0 — dopiero po zatwierdzeniu niezależnych przykładów, nie po nazwie „light LM”.
- Czy oddzielny `EvidenceGraph` będzie wewnętrznym pakietem `internal/evidence` czy modułem compiled pack — G04-07 decyduje API bez cyklu.

## Reguła zmian decyzji

Jeśli task ujawnia sprzeczność: wpisz konkretny przypadek z reprodukowalnym source SHA / expected, opisz opcje, rozstrzygnij we właściwym ADR i dopiero pisz kod. Nie rozszerzaj poza DoD, nie omijaj review, nie dokładaj legacy/v2 fallbacku jako obejścia.

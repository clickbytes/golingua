# Rozszerzenie macierzy A1 według zjawisk

Decyzja właściciela: 2026-10-25, audyt PL/EN realizowany szeregowo. Epik [#337](https://github.com/clickbytes/golingua/issues/337). Pierwotny snapshot audytu na base `9bf866bc620ec1ad4ae994876f57bc47885aac71` miał 24 rekordy kierunkowe (12 par) / 72 kontrole, w tym 36 `SPEC_PENDING` dla 12 kierunków LQ/PO/Q; to historyczny wynik sprzed M3, nie aktualny licznik. M2–M7 zostały scalone, a ich main CI było zielone. M8 #344 został zaimplementowany w [PR #351](https://github.com/clickbytes/golingua/pull/351); końcowy zapis CI i scalenia znajduje się w issue #344 i PR. Wartości 141, 206 i 268 pojawiały się w archiwalnych etapowych snapshotach z 2026-10-25; są historyczne i zostały zastąpione przez obecne 500 par. Wcześniejszy raport A1.5 na źródle `a9eb60e+working-tree` wykazał 72 `GAP_I`; jest to wynik tego etapu, nie obecny wynik pełnej macierzy. Również historyczne 24 `GAP_I` M3 i 36 `SPEC_PENDING` M2 odnoszą się tylko do wskazanych dawnych przebiegów. Aktualny stan i pełny run: [wyniki #344](344-results.md).

| Etap | Issue | Zakres | Stan |
| --- | --- | --- | --- |
| M2 | [#338](https://github.com/clickbytes/golingua/issues/338) | Inwentarz, warianty i raport | DONE; merged, main CI green |
| M3 | [#339](https://github.com/clickbytes/golingua/issues/339) | Kontrakty LQ/PO/Q | DONE; merged, main CI green |
| M4 | [#340](https://github.com/clickbytes/golingua/issues/340) | Reprezentatywny korpus | DONE; merged, main CI green |
| M5 | [#341](https://github.com/clickbytes/golingua/issues/341) | Kontrasty | DONE; merged, main CI green |
| M6 | [#342](https://github.com/clickbytes/golingua/issues/342) | Kompozycja i przekształcenia | DONE; merged, main CI green |
| M7 | [#343](https://github.com/clickbytes/golingua/issues/343) | Odmowy i niejednoznaczność | DONE; merged, main CI green |
| M8 | [#344](https://github.com/clickbytes/golingua/issues/344) | Zbalansowana macierz, obserwacje i raport luk | Implementacja ukończona; draft PR #351, CI PR/main i merge pending |

Obecny inwentarz ma **500 par / 1000 kierunków** w partiach A1–A1.13. Wszystkie cele grupowe osiągnięto: actions 65, identification 40, possession 40, location_quantity 50, negation 50, questions 60, time_agreement 60, case_adjective_order 45, ambiguity_reference 45 i composition_boundary 45. To cele planistyczne próby, nie kryteria poprawności. Grupy i tagi opisują wymagania; nie dowodzą obsługi konstrukcji ani pokrycia gramatyki. Partie nie wyznaczają limitu rekordów i nie są certyfikacją CEFR. Raport `tests/results/latest.md` zmieni się automatycznie dopiero po pushu do `main`; nie edytuje się go ręcznie.

Wynik techniczny M8 i obserwacje są oddzielone od akceptacji językowej właściciela, która pozostaje `NOT_EVALUATED`. `GAP`, `FAIL`, `SPEC_PENDING`, `NOT RUN` i `INFRA` pozostają jawne. 32 odmowy, kontrole metamorficzne, wieloznaczności i wsparcia są dodatkowymi kategoriami, osobnymi od PASS zwykłego tłumaczenia. Obserwacyjny workflow nie zmienia technicznego `go-ci`.

## Delegacja

Dzisiejszy audyt OpenCode (`/tmp/golingua-opencode-audit-20261025/coordinator-note.md`) wykazał brak potwierdzenia odmów shell/network/subagent i brak workera w diagnostycznej liście agentów. OpenCode nie dostał implementacji. Dopuszczony zastępca: Luna Low, jedno małe zadanie i jeden writer naraz. Koordynator zachowuje projekt kontraktu, review, integrację, rzeczywiste kontrole i GitHub.

Briefy i dowody poszczególnych etapów znajdują się w tym katalogu; surowe obserwacje pełnej macierzy publikuje workflow jako artefakt.

## #344 M8-PO34 — A1.6 possession

Dodano 34 pary / 68 kierunków jako kontrakty deklaratywne. Pełne znaczenia posiadania, przynależności, własności prawnej, pokrewieństwa i przyjaźni pozostają jawnie `GAP_I` tam, gdzie model Meaning nie może reprezentować relacji CORE. Zakres i lista wejść: [brief](344-possession-brief.md), [wybór](344-possession-selection.md), [wyniki](344-possession-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 M8-LQ48 — A1.7 location and quantity

Dodano dokładnie 48 par / 96 kierunków jako deklaratywne kontrakty A1.7. Pełne znaczenia lokalizacji oraz egzystencji z liczebnością pozostają jawne `GAP_I`; graf dwóch kotów jest komponentem samej liczebności. Zakres i lista: [brief](344-location-brief.md), [wybór](344-location-selection.md), [wyniki](344-location-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 reference matching prerequisite

Context matching now supports multiple anchored mentions bound to graph nodes by the same structural bijection, including variant-specific source matching and integration source-candidate linkage. Existing one-referent identification checks remain strict. Details: [brief](344-reference-matching-brief.md) and [results](344-reference-matching-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 M8-Q58 — A1.9 questions

Added 58 authored question pairs / 116 directions. Yes/no fixture graphs cover complete propositions with question act, time and polarity; WH binding, full possession/location, quantity binding, open-state and other unsupported requirements remain explicit gaps without invented graphs. Row 30 explicitly says walking in both languages; the question still lacks its destination-variable binding. The Polish-only plural `drzwi` target retains four English number/deixis candidates. Selection, contract and observed run: [brief](344-questions-brief.md), [selection](344-questions-selection.md), [results](344-questions-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 M8-TIME-PERSON — A1.10 osobowe ramy czasu

Dodano wybrane 16 par / 32 kierunki dla PL↔EN. Kontrakty zachowują ramy czasu, liczbę uczestnika, kotwice osób, warianty `read`, dziecko/offspring oraz lokalne założenia fleksyjne. Pozostałe 40 przypadków A1.10 pozostaje queued. Szczegóły: [brief](344-time-person-brief.md), [selection](344-time-person-selection.md), [wyniki](344-time-person-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 M8-TIME-OTHER — pozostałe A1.10

Dodano 40 par / 80 kierunków, domykając wybrane 56 par A1.10. Zwykłe ramy i liczebność używają istniejących I; przyszłość, progresywność, częstotliwość oraz nakładanie się zdarzeń pozostają jawnie `GAP_I`, bez osłabionych grafów. TIME45 zachowuje nierozstrzygniętą liczebność polskiego pluralia tantum. Szczegóły: [brief](344-time-other-brief.md), [selection](344-time-other-selection.md), [wyniki](344-time-other-results.md). OwnerAcceptance: `NOT_EVALUATED`.

## #344 plural-only door correction

Pairs A1.8 negation 9, 25, and 38 now keep Polish `drzwi` count unresolved in source I while English singular sources remain one. This is a contract correction from the architecture rule, not an engine change or owner approval. See [brief](344-door-count-brief.md) and [results](344-door-count-results.md). OwnerAcceptance: `NOT_EVALUATED`.


## #344 M8-CASE42 — A1.11 case and adjective order

Added 42 selected pairs / 84 directions. OVS and dative/accusative-fronted inputs retain ordinary role-based I graphs; on-support, open-state, instrument-role, writing-with-instrument, reciprocal conversation, and accompaniment gaps remain explicit without substitute graphs. Row 6 keeps Polish `drzwi` count unknown. Full list and observations: [brief](344-case-brief.md), [selection](344-case-selection.md), [results](344-case-results.md). OwnerAcceptance: `NOT_EVALUATED`.


## #344 M8-AMB43 — A1.12 ambiguity and reference

Added 43 selected pairs / 86 directions. The independent catalogue preserves licensed read frames, child readings, Polish role swaps, deictic anchors, entity counts, and distinct mentions; unsupported scope, attachment, and parent binding remain full I gaps. See [brief](344-ambiguity-brief.md), [selection](344-ambiguity-selection.md), and [results](344-ambiguity-results.md). OwnerAcceptance: `NOT_EVALUATED`.


## #344 M8-COMP33 — A1.13 composition boundaries

Added 33 selected pairs / 66 directions. Ordinary event/state compositions preserve authored role sharing, time, negation scope, exact black bearers, and row-33 explicit reflexivity; the five licensed imperatives remain request acts without asserted time. Group structure, connective relations, explicit politeness, posture change, modality, embedded content, and contrast remain full I gaps. See [brief](344-composition-brief.md), [selection](344-composition-selection.md), and [results](344-composition-results.md). OwnerAcceptance: `NOT_EVALUATED`.

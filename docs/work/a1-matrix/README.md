# Rozszerzenie macierzy A1 według zjawisk

Decyzja właściciela: 2026-10-25, wykonanie audytu macierzy w całości z zachowaniem pracy szeregowej. Epik [#337](https://github.com/clickbytes/golingua/issues/337). Pierwotny base audytu: `9bf866bc620ec1ad4ae994876f57bc47885aac71`; przebieg wyjściowy ujawnił 36 kontroli `SPEC_PENDING` dla 12 kierunków LQ/PO/Q. M2 usunął ograniczenia inwentarza i raportu; po scaleniu PR techniczny CI main był zielony. M3 opiera się na `cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd`.

| Etap | Issue | Zakres | Stan |
| --- | --- | --- | --- |
| M2 | [#338](https://github.com/clickbytes/golingua/issues/338) | Inwentarz, warianty i raport | DONE; merged, main CI green |
| M3 | [#339](https://github.com/clickbytes/golingua/issues/339) | Kontrakty LQ/PO/Q | DONE; merged, main CI green |
| M4 | [#340](https://github.com/clickbytes/golingua/issues/340) | Reprezentatywny korpus | DONE; merged, main CI green |
| M5 | [#341](https://github.com/clickbytes/golingua/issues/341) | Kontrasty | DONE; merged, main CI green |
| M6 | [#342](https://github.com/clickbytes/golingua/issues/342) | Kompozycja i przekształcenia | DONE; merged, main CI green |
| M7 | [#343](https://github.com/clickbytes/golingua/issues/343) | Odmowy i niejednoznaczność | DONE; merged, main CI green |
| M8 | [#344](https://github.com/clickbytes/golingua/issues/344) | Zbalansowana macierz, obserwowane pokrycie i raport luk | ACTIVE; jeden PR #351, kolejne małe partie szeregowo |

Obecny inwentarz obejmuje 382 par / 764 kierunki. Po M8-TIME-OTHER grupa `time_agreement` obejmuje 60 par / 120 kierunków; A1.10 zawiera 56 wybranych par w dwóch sekwencyjnych komponentach. Grupy działań, identyfikacji, posiadania, lokalizacji/ilości i negacji osiągnęły zaplanowaną liczebność; nie oznacza to obsługi tych konstrukcji przez silnik. Cel 500 par / 1000 kierunków pozostaje celem rozwojowym. Tagi opisują deklarowane wymagania; katalog słów nie jest pokryciem gramatyki. Nie jest to certyfikacja CEFR.

Ocena agenta wynika z rzeczywistych obserwacji. Akceptacja językowa właściciela: `NOT_EVALUATED`. Braki `GAP`, `FAIL`, `SPEC_PENDING`, `NOT RUN` i `INFRA` pozostają jawne. Poprawna odmowa błędnego wejścia jest odrębną kategorią od poprawnego tłumaczenia. Obserwacyjny workflow macierzy nie zmienia wyniku technicznego `go-ci`.

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

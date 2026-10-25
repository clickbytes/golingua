# Rozszerzenie macierzy A1 według zjawisk

Decyzja właściciela: 2026-10-25, wykonanie audytu macierzy w całości z zachowaniem pracy szeregowej. Epik [#337](https://github.com/clickbytes/golingua/issues/337). Pierwotny base audytu: `9bf866bc620ec1ad4ae994876f57bc47885aac71`; przebieg wyjściowy ujawnił 36 kontroli `SPEC_PENDING` dla 12 kierunków LQ/PO/Q. M2 usunął ograniczenia inwentarza i raportu; po scaleniu PR techniczny CI main był zielony. M3 opiera się na `cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd`.

| Etap | Issue | Zakres | Stan |
| --- | --- | --- | --- |
| M2 | [#338](https://github.com/clickbytes/golingua/issues/338) | Inwentarz, warianty i raport | DONE; merged, main CI green |
| M3 | [#339](https://github.com/clickbytes/golingua/issues/339) | Kontrakty LQ/PO/Q | DONE; merged, main CI green |
| M4 | [#340](https://github.com/clickbytes/golingua/issues/340) | Reprezentatywny korpus | ACTIVE; implementation in review |
| M5 | [#341](https://github.com/clickbytes/golingua/issues/341) | Kontrasty | ACTIVE; implementation in review |
| M6 | [#342](https://github.com/clickbytes/golingua/issues/342) | Kompozycja i przekształcenia | ACTIVE; implementation in review |
| M7 | [#343](https://github.com/clickbytes/golingua/issues/343) | Odmowy i niejednoznaczność | ACTIVE; implementation in review |
| M8 | [#344](https://github.com/clickbytes/golingua/issues/344) | Infrastruktura inwentarza i zbalansowana macierz | ACTIVE; infrastruktura gotowa, partia A1.4 dodaje 47 par działań; 500 par pozostaje celem |

Obecny inwentarz obejmuje 103 pary / 206 kierunków; A1.4 dodaje 47 ręcznie opisanych par i domyka grupę `actions` do 65. Cel 500 par / 1000 kierunków jest celem rozwojowym: dalsze zdania dobieramy z wykazanych luk, bez jednego masowego batcha. Deklarowane tagi nie dowodzą wykonania konstrukcji ani opanowania języka; katalog słów nie jest pokryciem gramatyki. Nie jest to certyfikacja CEFR.

Ocena agenta wynika z rzeczywistych obserwacji. Akceptacja językowa właściciela: `NOT_EVALUATED`. Braki `GAP`, `FAIL`, `SPEC_PENDING`, `NOT RUN` i `INFRA` pozostają jawne. Poprawna odmowa błędnego wejścia jest odrębną kategorią od poprawnego tłumaczenia. Obserwacyjny workflow macierzy nie zmienia wyniku technicznego `go-ci`.

## Delegacja

Dzisiejszy audyt OpenCode (`/tmp/golingua-opencode-audit-20261025/coordinator-note.md`) wykazał brak potwierdzenia odmów shell/network/subagent i brak workera w diagnostycznej liście agentów. OpenCode nie dostał implementacji. Dopuszczony zastępca: Luna Low, jedno małe zadanie i jeden writer naraz. Koordynator zachowuje projekt kontraktu, review, integrację, rzeczywiste kontrole i GitHub.

Briefy i dowody poszczególnych etapów znajdują się w tym katalogu; surowe obserwacje pełnej macierzy publikuje workflow jako artefakt.

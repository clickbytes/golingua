# LIGHT R zachowanie znaczenia i gramatyka PL EN

Decyzja właściciela 2026-10-25: wdrożyć sześć szeregowych pionów po [audycie](../architecture-audit/README.md). [Epik #354](https://github.com/clickbytes/golingua/issues/354). Koordynator integruje i odbiera małe zadania delegowane; najwyżej jeden writer. Każdy pion kończy się działającym przepływem CLI, niezależnym I, kontrolą analizy/realizacji, PR i rzeczywistym CI main. Pełna macierz pozostaje obserwacyjna. Akceptacja językowa właściciela: NOT_EVALUATED.

| Pion | Issue | Stan |
| --- | --- | --- |
| Negacja PL i przypadek dopełnienia | [355](https://github.com/clickbytes/golingua/issues/355) | DONE — PR #361, CI PR/main SUCCESS |
| Liczba i potrzebne formy referentów | [356](https://github.com/clickbytes/golingua/issues/356) | DONE — PR #362, CI PR/main SUCCESS |
| Przeszłość i uzgodnienie | [357](https://github.com/clickbytes/golingua/issues/357) | ACTIVE |
| Mówiący i adresat | [358](https://github.com/clickbytes/golingua/issues/358) | QUEUED |
| Alternatywy określoności EN | [359](https://github.com/clickbytes/golingua/issues/359) | QUEUED |
| Wiek i potomek child children | [360](https://github.com/clickbytes/golingua/issues/360) | QUEUED |

Zakres nie obejmuje HAVE, przestrzeni, ilości/WH, nowego schematu I ani uczenia. [LIGHT-L #329](https://github.com/clickbytes/golingua/issues/329) oczekuje na zakończenie tego epiku. Przypadki i wymagania pochodzą z istniejącej macierzy; rozszerzenia danych ograniczają się do wybranego pionu.

OpenCode 2026-10-25: użyto dziennej kontroli tego samego środowiska i obrazu v2.0.15 (sha256:cf24fe4124f8ceb2253e52157f651fbf3bdc16bca9e3e0c2595fd7e32107a9f9). Na hoście brak komendy opencode; dostępny izolowany obraz bez zmian. Wcześniejsza kontrola nie wykazała workera w debug agents ani trzech wymaganych odmów narzędziowych. Bez ponawiania niezmienionej niespełnionej próby i bez --auto; dopuszczone małe zadania Luna Low. Szczegóły dziennej kontroli są w rejestrze [LIGHT-F](../light-lm/README.md); lokalna notatka /tmp/golingua-opencode-audit-20261025/coordinator-note.md.

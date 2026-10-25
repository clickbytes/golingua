# A1-M7 #343 — zakres

Base: `032860473454f1dada4288d4f48e5abfdc4e9d64`; branch: `a1-m7-boundaries`. Zakres danych to dokładnie 8 wejść EN i 8 wejść PL z `/tmp/golingua-matrix-expansion/m7-selected.json`, każde z pochodzeniem i `OwnerAcceptance: NOT_EVALUATED`.

Dane odmów nie są rekordami przekładu i nie trafiają do `translation.Cases`. Osiem wejść w każdym języku sprawdza natywną analizę i integrację tłumaczenia: 32 nazwane podprzypadki. PASS oznacza `unsupported`/`invalid` bez grafu Meaning i bez tekstu. Jawne `unknown` ma osobną opt-in dla dokładnych etapów: analiza natywna pustego wejścia i nieznanych leksemów oraz tłumaczenie wyłącznie nieznanych leksemów. Każda taka zgoda ma niepusty powód i wymaga niepustej diagnostyki, bez znaczenia/tekstu/sukcesu. Nieznane wejście pustego pipeline tłumaczenia pozostaje `invalid`; inne `unknown`, `incomplete`, `ok`, `ambiguous`, generic error, resource limit ani infrastruktura nie zaliczają się.

Reporter wydziela granice odmowy w oddzielnej tabeli z pakietem i nazwą kontroli; rekordowe liczniki i dodatkowe kontrole pozostają odrębne. `TestAmbiguityAssumptionsFollowRoles` na rzeczywistym PL→EN wymaga czterech wariantów `a/the`, weryfikuje każde znaczenie względem niezależnego Seeing oraz mapuje `determiner` i wartość do ścieżki NodeID wyprowadzonej z ról `experiencer`/`stimulus`.

Bez zmian engine, gramatyki, CI i korpusu normalnych translation cases.

# Końcowy odbiór DSE na main

Źródło `386f0d98ca12bde7e53880109db4f634958321c6`, hash binarki i wejść: `environment.json`. `commands.jsonl` zawiera argv, exit code, stderr i pliki stdout. Binarkę zbudowano przez `go build -o /tmp/golingua-dse-400 ./cmd/golingua`. Snapshot odtworzono lokalnie, nie jest dodany do repozytorium; jego digest jest w `snapshot.json`. Konfiguracje wskazują `/tmp/golingua-dse-400-store` oraz ścieżki do zasobów repozytorium.

Neutralny `construct` → `verify` → `evaluate` jest `ok`, odpowiedzi mają dowody faktów. Niezależne I→PL/EN jest `ok` dla jawnego profilu czasu teraźniejszego; to nie ustala prawdziwości ani akceptacji językowej. Produkcyjne tłumaczenia WH w obu kierunkach zwracają tekst i `incomplete` (exit 7), z jawnymi założeniami. Pierwsza próba EN→PL z ogólną konfiguracją pilota pozostaje w dowodach jako `incomplete` bez tekstu; konfiguracja WH pozwala na tekst, ale nie usuwa reszt. Akceptacja właściciela: NOT EVALUATED.

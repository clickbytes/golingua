# A1-M7 #343 — obserwacje

Base: `032860473454f1dada4288d4f48e5abfdc4e9d64`; pełna obserwacja wykonana na `+working-tree`.

## Komendy

```sh
GOLINGUA_MATRIX_SHA='032860473454f1dada4288d4f48e5abfdc4e9d64+working-tree' go test ./tests/... -count=1 -v
python3 scripts/report-language-matrix.py --sha '032860473454f1dada4288d4f48e5abfdc4e9d64+working-tree' --exit-code 1 --log /tmp/golingua-m7-full.log
```

Pełne Go test zakończyło się kodem `1` z powodu znanych GAP/FAIL zwykłej macierzy. Pierwszy przebieg wykazał sześć wyników `unknown`; review rozdzielił brak wejściowych tokenów i brak wiedzy o nieznanych leksemach od odmowy wadliwej składni. Dodano jawne, per-stage zgody tylko dla dokładnie tych przypadków, z wymaganym uzasadnieniem i diagnostyką. Końcowy raport: `/tmp/golingua-m7-report.md`; surowy log: `/tmp/golingua-m7-full.log`.

## Granice odmowy

Końcowy reporter wykrył wszystkie 32 podprzypadki w oddzielnej sekcji: 32 PASS i 0 FAIL, poza zwykłymi licznikami rekordów. Wcześniejsze sześć `unknown` zostało sklasyfikowanych per przypadek: natywny analizator dla pustego wejścia i nieznanych leksemów PL/EN; tłumaczenie dla nieznanych leksemów PL/EN. Te sześć dozwolonych wyników ma konkretne diagnostyki braku tokenów/znanych odczytów i nie zwraca grafu ani tekstu. Tłumaczenie pustego wejścia zwraca `invalid` bez outputu. Żadne inne `unknown` ani resource-limit/untyped runtime error nie jest dozwolone. Wszystkie surowe obserwacje zawierają `check_class: refusal`, SHA, wejście, pochodzenie, wynik i typed-error result, jeżeli wystąpił.

Kontrola `TestAmbiguityAssumptionsFollowRoles` PASS: cztery angielskie kombinacje rodzajników zachowały właściwy graf; każde założenie `determiner` ma oczekiwaną wartość i ścieżkę z referenta odpowiedniej roli.

## Zwykła macierz i pozostałe kontrole

Zwykłe liczniki macierzy pozostały odrębne: `PASS 175`, `FAIL 37`, `GAP_ANALYZE 28`, `GAP_REALIZE 48`, `GAP_I 48`, `SPEC_PENDING 0`, `INFRA 0`, `NOT RUN 0`. Reporter nie dodaje odmów do tych liczników. Wymagania kontraktów nie zostały dopasowane do wyniku silnika.

`gofmt` i `git diff --check` przechodzą. `go vet ./...`, `go build ./cmd/golingua`, smoke CLI i 16 prób CLI koordynatora: `NOT RUN` przez wykonawcę. Akceptacja językowa właściciela: `NOT_EVALUATED`.

Koordynator: review doprecyzował, że dla tego jednorzędowego zdarzenia ścieżka założenia ma być dokładnie event→argument roli, nie dowolny ciąg zawierający ID. Rzeczywiste16 CLI:12 unsupported/exit3,2 invalid/exit2 (puste żądanie),2 unknown/exit5 (nieznane leksemy), bez wspieranego Meaning lub tekstu. Pełne raw JSON w evidence/343-cli.json. Vet/build/help/version/gofmt/diff PASS; końcowy lokalny przebieg i CI zapisane poniżej/issue #343.

Ostateczny przebieg po integracji: `/tmp/golingua-matrix-expansion/m7-full.log`, exit1. Raport z tego samego logu:112 rekordów/336 kontroli zwykłych (175 PASS,37 FAIL,124 GAP), osobno32/32 PASS odmów, dodatkowo PASS warunków rodzajnika. OVS pozostaje oddzielnym GAP_ANALYZE w metamorficznych. Nie ma nowych SPEC_PENDING/INFRA/NOT RUN.

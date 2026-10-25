# A1-M6 #342 — obserwacje

Base: `573274f48148cb65a050bc38c0107bb167264a52`; pełna macierz została uruchomiona na `+working-tree`.

## Komendy

```sh
go test ./tests/interlingua -run CompositionBatch -count=1
go test ./tests/translation -run Inventory -count=1
GOLINGUA_MATRIX_SHA='573274f48148cb65a050bc38c0107bb167264a52+working-tree' go test ./tests/... -count=1 -v
python3 scripts/report-language-matrix.py --sha '573274f48148cb65a050bc38c0107bb167264a52+working-tree' --exit-code 1 --log /tmp/golingua-m6-full.log
```

Kontrola niezależnych grafów przeszła, podobnie inwentarz. Pełna macierz zakończyła się kodem `1` z powodu jawnych FAIL/GAP. Raport: `/tmp/golingua-m6-report.md`; surowy log: `/tmp/golingua-m6-full.log`.

## Wynik

Raport zawiera 112 rekordów kierunkowych i wszystkie trzy kontrole na rekord. Liczniki rekordów: `PASS 175`, `FAIL 37`, `GAP_ANALYZE 28`, `GAP_REALIZE 48`, `GAP_I 48`, `SPEC_PENDING 0`, `INFRA 0`, `NOT RUN 0`. Dziewięć oddzielnych kontroli metamorficznych jest raportowanych package-qualified: role swap PL/EN, liczba PL/EN, czas PL/EN, polaryzacja PL/EN i SVO→OVS PL. Osiem pierwszych ma PASS; OVS ma `GAP_ANALYZE`, ponieważ analizator zwrócił `unsupported` dla `Nauczyciela widzi pies.`. Nie pominięto tej kontroli.

Nowe rekordy kompozycji pozostają GAP/FAIL w wielu kierunkach: analiza wielozdarzeniowych zdań i negacja z przymiotnikiem nie są obecnie obsługiwane, a realizacja nie obejmuje tej kompozycji. Wymagań tekstowych ani grafów nie dopasowano do outputu. Angielskie warianty rodzajników pozostają niezależne dla każdego count NP i nie dodają kwantyfikatora `all`.

`go vet ./...`, `go build ./cmd/golingua`, smoke CLI oraz 24 osobne próby CLI delegowane koordynatorowi: `NOT RUN` przez wykonawcę. Akceptacja językowa właściciela: `NOT_EVALUATED`.

Koordynator: review grafów współdzielenia, przypisania black i scope negacji oraz metamorficznych normalizacji zakończony. Gofmt/vet/build/help/version/diff-check PASS. Pełny przebieg base+working-tree powtórzony w `/tmp/golingua-matrix-expansion/m6-full.log`, liczniki jak wyżej. Wszystkie24 wywołania CLI wykonano; selected fields w `evidence/342-cli.json`, pełny stdout lokalnie w m6-cli.json. CI PR/main potwierdza issue #342.

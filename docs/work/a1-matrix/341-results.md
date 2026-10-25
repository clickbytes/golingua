# A1-M5 #341 — wyniki obserwacji

Base: `645e9f9255c3ec65548f87a91343fa1566f078ea`; wyniki poniżej z working tree oznaczonym `+working-tree`. Dane 16 par są propozycją koordynatora, a akceptacja językowa właściciela pozostaje `NOT_EVALUATED`.

## Zakres i komendy

Dodano 32 jawne, kierunkowe rekordy A1.2 z wybranych 16 par. Kontrakty posiadania opisują wymagania bez fikcyjnych grafów; pozostałe kontrasty mają niezależnie zapisane grafy I albo wskazują istniejące Reading. Angielskie `read` w liczbie mnogiej zachowuje obowiązkowe odczyty present i past oraz osobne polskie realizacje.

Pełna obserwacja:

```sh
GOLINGUA_MATRIX_SHA='645e9f9255c3ec65548f87a91343fa1566f078ea+working-tree' go test ./tests/... -count=1 -v
python3 scripts/report-language-matrix.py --sha '645e9f9255c3ec65548f87a91343fa1566f078ea+working-tree' --exit-code 1 --log /tmp/golingua-m5-correction-full.log
```

Pierwsza komenda zakończyła się kodem `1`, ponieważ macierz ma rzeczywiste FAIL/GAP. Raport po korekcie wygenerowano do `/tmp/golingua-m5-correction-report.md`; pełny log do `/tmp/golingua-m5-correction-full.log`.

## Wynik

Raport zawiera 88 kierunkowych rekordów i obserwuje trzy kontrole każdego rekordu. Po korekcie łącznie: `PASS 151`, `FAIL 39`, `GAP_ANALYZE 11`, `GAP_REALIZE 15`, `GAP_I 48`, `SPEC_PENDING 0`, `INFRA 0`, `NOT RUN 0` dla rekordów. Zmiana trzech asercji z FAIL do PASS wynika wyłącznie z usunięcia nieuzasadnionego wymagania jawnego `person`; pozostałe FAIL/GAP zachowano. Pakiety macierzy, interlingua i inwentarza przeszły; pakiety adapterów/języków kończą się FAIL zgodnie z niezamkniętymi wymaganiami i rozbieżnościami.

W nowych rekordach A1.2 widać między innymi:

- cztery przypadki posiadania oznaczone `GAP_I` z powodu ograniczeń reprezentacji, bez przypisywania im udawanego grafu;
- analizę obu odczytów `read` EN dla wymaganych grafów present/past, lecz realizacja PL→EN nie zachowuje obu wymaganych kontraktów;
- `GAP_ANALYZE`/`GAP_REALIZE` dla czarnego atrybutu przypisanego do obiektu;
- FAIL w odmianie i negacji oraz przy kontraktach szyku/określoności tam, gdzie rzeczywisty wynik nie spełnił oczekiwań.

Wymagania i oczekiwane warianty nie zostały dopasowane do outputu. To propozycje kontraktów i obserwacje silnika, nie zatwierdzenie językowe właściciela ani ocena poziomu CEFR. Osobny wynik CLI dla wybranych wejść, wykonany przez koordynatora, znajduje się w `docs/work/a1-matrix/evidence/341-cli.json`.

## Pozostałe kontrole

- `gofmt` wykonano na nowych i zmienionych plikach Go.
- `git diff --check` przechodzi.
- `go vet ./...`, `go build ./cmd/golingua` i smoke CLI: `NOT RUN` przez wykonawcę; koordynator przeprowadza kontrole techniczne.

## Korekta po review

Usunięto `person` z wymaganych założeń realizacji: pospolity podmiot gramatyczny sam ustala osobę, więc jawny atrybut `person` jest dozwolonym szczegółem implementacyjnym, a nie wymaganiem kontraktu. Dla angielskich wariantów pozostają wymagane nierozstrzygnięte wybory `determiner`. Identyczne warunki per tekst zostały zredukowane do pojedynczych unikalnych par tekst/pole/wartość; nie deklarują one ścieżek semantycznych podmiotu ani dopełnienia.

Koordynator po review: `go vet ./...`, build, `--help`, `--version`, gofmt i diff-check PASS. Pełny przebieg z base+working-tree powtórzony do `/tmp/golingua-matrix-expansion/m5-full.log`; reporter potwierdził151 PASS,39 FAIL,11 GAP_ANALYZE,15 GAP_REALIZE,48 GAP_I,0 SPEC_PENDING/INFRA/NOT RUN. CLI wszystkich32 kierunków wykonane niezależnie; selected output w evidence/341-cli.json. Nie zmieniano gramatyki ani oczekiwań na podstawie silnika. CI PR/main zapisane w issue #341.

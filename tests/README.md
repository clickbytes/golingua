# Macierz zachowania PL/EN

Macierz opisuje 24 kierunkowe rekordy z próbki A1: po cztery ID z domen `AC`, `ID`, `LQ`, `NEG`, `PO` i `Q` (`001`–`004`). To ograniczona próbka, nie pełne A1 ani certyfikacja CEFR; [deskryptory Rady Europy](https://rm.coe.int/16809ea0d4) są jedynie kontekstem. Zdania są własną redakcją, nie kopiami z egzaminu. Oryginalne ID, teksty, wymagania i provenance pozostają w danych językowych; akceptacja właściciela jest zapisana osobno jako `NOT_EVALUATED`.

## Układ i kontrakt

```text
tests/
  interlingua/                 # niezależne, ręcznie zapisane znaczenia I
  languages/pl/A1/             # analiza i realizacja po polsku
  languages/en/A1/             # analiza i realizacja po angielsku
  translation/cases.go         # 24 stabilne ID i odwołania do obu stron
  translation/pl-en/           # PL → I → EN
  translation/en-pl/           # EN → I → PL
  matrix/                      # wspólne typy, runtime produkcyjny i kontrole
```

A1.1 i A1.2 oznaczają przyszłe partie A1 konkretnego języka; wymagają osobno określonego zakresu i nie zastępują tej próbki.

Pliki `*.testdomain.go` zawierają dane. Pliki `*_test.go` wykonują lokalne kontrole standardowym `go test`; te testy językowe nie są częścią CI. Każdy przypadek analizy zachowuje dokładne wejście, wymagania, nierozstrzygnięcia i pochodzenie wejścia. Przypadek realizacji zachowuje `ProposalTexts`, pochodzenie propozycji i akceptację właściciela oddzielnie od zamkniętych `AllowedTexts`, dozwolonych/obowiązkowych założeń i strat. Propozycja nie staje się automatycznie oczekiwanym wynikiem.

Integracja wskazuje ID analizy i realizacji; nie kopiuje ich treści. Inwentarz sprawdza dokładnie 24 ID, wszystkie odwołania, brak osieroconych przypadków oraz kompletność wymaganych pól provenance. Analiza i realizacja PL działają niezależnie od EN i odwrotnie. Wykonanie przekładu uruchamia oba adaptery produkcyjne przez rzeczywisty pipeline.

Wzorce `MeaningGraph` są tworzone niezależnie od parserów. Ograniczone porównanie grafów jest niewrażliwe na lokalne NodeID i kolejność węzłów, ale zachowuje role i bijekcję referentów. Nie scala odrębnych węzłów tylko dlatego, że mają ten sam ConceptID. Wzorce nie dopisują tożsamości osobników ani płci zwierzęcia. Kontekst wskazania przechowuje osobną nierozstrzygniętą Mention bez `ReferentID` i `EvidenceID`.

Negacja w obecnym wycinku ma własne znaczenia I: `negation.classification.this` i `.to` umieszczają jeden zakres negacji nad stanem `class_membership`; `negation.seeing.teacher-dog` umieszcza go nad zdarzeniem widzenia. Każdy graf ma jeden korzeń będący zakresem negacji, bez dodatkowego korzenia predykatu. Dla klasyfikacji adapter celu odrzuca kontekst, którego graf ma inną polaryzację niż znaczenie.

Wynik silnika (`ok`, `ambiguous`, `incomplete`, `unsupported`) jest oddzielony od oceny przypadku. Kontrole sprawdzają rzeczywiste grafy, konteksty, teksty, założenia, straty i diagnostyki; sam status `ok` nie oznacza PASS. Brak zamkniętego wzorca daje `SPEC_PENDING`, nie domyślny PASS ani udowodniony błąd parsera. Nie używamy `t.Skip`.

| Ocena | Znaczenie |
| --- | --- |
| `PASS` / `FAIL` | Wykonany przypadek spełnia kontrakt / wynik jest z nim sprzeczny. |
| `GAP_ANALYZE` / `GAP_REALIZE` | Brakuje wymaganej analizy / realizacji danego języka. |
| `GAP_I` | Obecna reprezentacja nie potrafi wyrazić wymaganego znaczenia. |
| `SPEC_PENDING` / `NOT RUN` | Brakuje zamkniętego kontraktu / kontrola nie została wykonana. |
| `INFRA` | Runtime lub konfiguracja uniemożliwia wiarygodne wykonanie. |

Żaden GAP, FAIL, SPEC_PENDING, NOT RUN ani INFRA nie jest pomijany ani zaliczany jako zielony wynik.

## Identyfikacja objęta zamkniętym kontraktem

Znaczenie klasyfikacji to `class_membership(member → cat)`: wskazany referent jest klasyfikowany jako kot. To nie jest identyczność z innym osobnikiem ani aktualizacja ontologii. Znaczenie opisu koloru to `black(bearer → cat)`: czarny kolor przypisany temu wskazanemu kotu.

| ID I | Kotwica nierozstrzygniętej Mention |
| --- | --- |
| `identification.classification.this` | `deixis=proximal` |
| `identification.classification.to` | `deixis=unresolved` |
| `identification.black.this` | `deixis=proximal` |
| `identification.black.ten` | `deixis=demonstrative` (odległość nieznana) |

Realizacja EN dla nierozstrzygniętego `to` zachowuje warianty `This is a cat.`, `That is a cat.` i `It is a cat.`; dla `ten` zachowuje `This cat is black.` i `That cat is black.`. Każdy wybrany angielski wariant ma jawne założenie `reference`. Rodzime I z kotwicą `proximal` daje tylko wariant `This` bez takiego założenia. Polski `to`/`ten` nie koduje jawnie bliskości `this`; przekład z EN deklaruje stratę `deixis`. Założenia i straty są odrębnymi informacjami.

## Odbiór identyfikacji

Przed testami zapisz commit, którego wyników używasz:

```sh
export GOLINGUA_MATRIX_SHA="$(git rev-parse HEAD)"
```

Jeśli working tree zawiera zmiany, przed uruchomieniem obejrzyj i zapisz zakres diffu, a SHA oznacz jako `$(git rev-parse HEAD)+working-tree`. Bez tej zmiennej logi pokazują `UNRECORDED` i nie są dowodem przypisanym do SHA. Zachowuj surowy output oraz obok niego dokładne polecenie i exit code; nie zmieniaj expected na podstawie outputu.

```sh
go test ./tests/interlingua/... -run Identification -count=1 -v
go test ./tests/languages/pl/A1/... -run Identification -count=1 -v
go test ./tests/languages/en/A1/... -run Identification -count=1 -v
go test ./tests/translation/... -run Identification -count=1 -v
```

## Odbiór negacji

Uruchom cztery niezależne kontrole negacji oraz kontrolne Seeing i inwentarz:

```sh
go test ./tests/interlingua/... -run Negation -count=1 -v
go test ./tests/languages/pl/A1/... -run Negation -count=1 -v
go test ./tests/languages/en/A1/... -run Negation -count=1 -v
go test ./tests/translation/... -run Negation -count=1 -v
go test ./tests/... -run 'Negation|Identification|Seeing|Inventory' -count=1 -v
```

Ostatnie polecenie sprawdza również zgodność polaryzacji grafu znaczenia i kontekstu wskazania w obu adapterach. Przykłady negacji nie zastępują kontroli zamiany ról z istniejących przypadków Seeing.

## Seeing, inwentarz i pełny obraz

```sh
go test ./tests/interlingua/... -run Seeing -count=1 -v
go test ./tests/languages/pl/A1/... -run Seeing -count=1 -v
go test ./tests/languages/en/A1/... -run Seeing -count=1 -v
go test ./tests/translation/... -run Seeing -count=1 -v
go test ./tests/translation/... -run Inventory -count=1 -v
go test ./tests/... -count=1 -v
```

Pełna macierz jest celowo niezaliczona, dopóki pozostałe wymagania nie mają zamkniętych wzorców i działających tras. Na końcowym snapshotcie oczekiwany był niezerowy exit wyłącznie z **36 kontrolami `SPEC_PENDING` dotyczącymi 12 rekordów domen LQ, PO i Q**. Pozostałe 12 rekordów obejmuje Seeing, identyfikację i negację. Każdy z 24 oryginalnych rekordów nadal pojawia się w raporcie; `SPEC_PENDING` ani `NOT RUN` nie są pomijane i nie oznaczają PASS. Językowe kontrole pozostają poza CI; techniczne CI repozytorium pozostaje bez zmian.

## Obserwacje CLI dla czterech oryginalnych zdań

Na dostępnej próbce otrzymano następujące obserwacje z CLI:

| ID | Wejście i kierunek | Tekst kandydata | Status silnika / exit |
| --- | --- | --- | --- |
| `A1-ID-001-EN-PL` | `This is a cat` · EN→PL | `To jest kot.` | `incomplete` / 7; strata `deixis` |
| `A1-ID-002-PL-EN` | `To jest kot.` · PL→EN | `This is a cat.`, `That is a cat.`, `It is a cat.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |
| `A1-ID-003-EN-PL` | `This cat is black.` · EN→PL | `Ten kot jest czarny.` | `incomplete` / 7; strata `deixis` |
| `A1-ID-004-PL-EN` | `Ten kot jest czarny.` · PL→EN | `This cat is black.`, `That cat is black.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |

Są to wyniki silnika zgodne z ograniczonym kontraktem testowanej próbki; `incomplete` i `ambiguous` są jawne i nie rozstrzygają tożsamości ani płci. Końcowy surowy JSON z SHA i poleceniami: `/tmp/a1-t05-cli.json` (pierwsze cztery obserwacje to identyfikacja; kolejne to wybrane kontrole istniejących tras). Odtworzenie tych czterech wejść wykonuje się po przygotowaniu runtime poniżej.

## Obserwacje CLI dla negacji

Na końcowym snapshotcie CLI zwróciło poniższe wyniki. Status i exit CLI są obserwacją silnika; założenia i straty pozostają osobnymi polami.

| ID | Wejście i kierunek | Teksty kandydatów | Status / exit |
| --- | --- | --- | --- |
| `A1-NEG-001-EN-PL` | `This is not a cat.` · EN→PL | `To nie jest kot.` | `incomplete` / 7; strata `deixis` |
| `A1-NEG-002-PL-EN` | `To nie jest kot.` · PL→EN | `It is not a cat.`, `That is not a cat.`, `This is not a cat.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |
| `A1-NEG-003-EN-PL` | `The teacher does not see the dog.` · EN→PL | `Nauczyciel nie widzi psa.` | `ok` / 0 |
| `A1-NEG-004-PL-EN` | `Nauczyciel nie widzi psa.` · PL→EN | `The teacher does not see the dog.`, `The teacher does not see a dog.`, `A teacher does not see the dog.`, `A teacher does not see a dog.` | `ambiguous` / 6; dobór `determiner` jest jawny dla każdej opcji |

Wszystkie cztery grafy pośrednie mają jeden korzeń zakresu negacji. Dodatkowe kontrole zachowały zamianę ról, liczbę mnogą przy słuchaniu/patrzeniu, niejednoznaczność czytania, a podwójne lub przestawione `not`/`nie` otrzymały `unsupported`. Surowe obserwacje z poleceniami, exit code i SHA: `/tmp/a1-neg-final-cli.json`; testy I, PL, EN, przekładu, kontrole, vet, build i pełnej macierzy: `/tmp/a1-neg-final-{i,pl,en,translation,controls,vet,build,full}.log`. Pełny przebieg zakończył się niezerowo wyłącznie 36 przypadkami `SPEC_PENDING` opisanymi wyżej.

## Przygotowanie świeżego snapshotu i CLI

Z katalogu repozytorium zbuduj bieżący program i snapshot do katalogów tymczasowych:

```sh
go build -o /tmp/golingua-a1-neg ./cmd/golingua
/tmp/golingua-a1-neg snapshot --apply --store /tmp/golingua-a1-neg-store --json
cat > /tmp/golingua-a1-neg-runtime.json <<'JSON'
{"schema_version":1,"dataset_store":"/tmp/golingua-a1-neg-store","dataset_id":"golingua-authored","language_package_dirs":[]}
JSON
```

Uruchom cztery rekordy osobno; `--text` zachowuje dokładny zapis wejścia:

```sh
/tmp/golingua-a1-neg translate --from en --to pl --text 'This is a cat' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'To jest kot.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from en --to pl --text 'This cat is black.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'Ten kot jest czarny.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
```

Odtwórz cztery nowe przepływy negacji tym samym snapshotem:

```sh
/tmp/golingua-a1-neg translate --from en --to pl --text 'This is not a cat.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'To nie jest kot.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from en --to pl --text 'The teacher does not see the dog.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'Nauczyciel nie widzi psa.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
```

## Odtworzenie zaobserwowanych przepływów TUI

Obserwacje czterech przepływów identyfikacji wykonano w rzeczywistym terminalu z binarium zbudowanym na kodzie A1-T04 i izolowaną konfiguracją użytkownika. Uruchomienie w TTY (ścieżka binarium może wskazywać własny build):

```sh
TERM=xterm-256color XDG_CONFIG_HOME=/tmp/golingua-a1-neg-config /tmp/golingua-a1-neg
```

W UI naciśnij `F5`, wpisz `/tmp/golingua-a1-neg-runtime.json` i `Ctrl+S` (zaobserwowane: ustawienia zapisane, runtime `ok`, `ready:true`). Naciśnij `F1`, wpisz dokładne zdanie i Enter. Dla EN→PL sprawdź `This is a cat` oraz `This cat is black.`. Użyj `Alt+D`, by przełączyć kierunek PL→EN, i sprawdź `To jest kot.` oraz `Ten kot jest czarny.`. `Tab` przechodzi między kierunkiem, kandydatami i podglądem; `Up`/`Down` wybierają kandydata, Enter otwiera szczegóły, `Esc` wraca/resetuje fokus, a `Ctrl+U` czyści prompt. Podgląd zachowuje status `incomplete` lub `ambiguous`; w szczegółach widoczne są odpowiednio strata `deixis` lub wybór `reference`. Zakończ `Ctrl+C`.

Rzeczywista obserwacja potwierdziła cztery przepływy TUI identyfikacji i czysty exit po `Ctrl+C`; konfiguracja użytkownika nie została zmieniona. Notatka dowodowa: `/tmp/a1-t05-tui-observation.md` (zawiera też build metadata i sumę binarium).

Dla negacji użyto świeżego binarium `/tmp/golingua-a1-neg` z runtime `/tmp/golingua-a1-neg-runtime.json` oraz odizolowanego `XDG_CONFIG_HOME`. TUI pokazało `To nie jest kot.` ze statusem `incomplete` i stratą `deixis`, `Nauczyciel nie widzi psa.` ze statusem `ok`, trzy warianty EN dla `To nie jest kot.` ze statusem `ambiguous` i widocznym wyborem `reference=this`, oraz cztery warianty rodzajników dla `Nauczyciel nie widzi psa.` ze statusem `ambiguous` i założeniami `determiner`. Podgląd zachował status `ambiguous` i wybrany tekst. Wszystkie cztery przepływy zakończono `Ctrl+C` z exit 0; ustawienia użytkownika nie uległy zmianie. Notatka: `/tmp/a1-neg-tui-observation.md`; surowy transcript: `/tmp/a1-neg-tui.typescript`.

## Status i pochodzenie

Ocena techniczna agenta dla zamkniętej identyfikacji i negacji: PASS. Akceptacja językowa właściciela: `NOT_EVALUATED`. Pozostałe domeny A1 pozostają poza ukończonym zakresem.

Powiązane zadania: epik [#311](https://github.com/clickbytes/golingua/issues/311), fundament [#312](https://github.com/clickbytes/golingua/issues/312), identyfikacja I [#314](https://github.com/clickbytes/golingua/issues/314), PL [#316](https://github.com/clickbytes/golingua/issues/316), EN [#318](https://github.com/clickbytes/golingua/issues/318), integracja [#320](https://github.com/clickbytes/golingua/issues/320), negacja [#322](https://github.com/clickbytes/golingua/issues/322). Dodatkowe informacje: [architektura](../docs/architecture/README.md), [CLI](../docs/cli.md), [przegląd ręczny](../docs/manual-review.md).

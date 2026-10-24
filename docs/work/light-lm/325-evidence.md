# Odbiór lokalny LIGHT-F1

Wykonano 2026-10-24 na [99aefcd5658f89f1b07ead90f3fd0b03923435b7](https://github.com/clickbytes/golingua/commit/99aefcd5658f89f1b07ead90f3fd0b03923435b7), Linux amd64, Go 1.27.0. Późniejszy commit dowodów zmienia wyłącznie dokumentację. Ocena koordynatora: **PASS dla zakresu #325**. Akceptacja językowa właściciela: **NOT_EVALUATED**. Merge i rzeczywiste CI PR/main są osobną bramką; ich końcowe SHA, runy i wynik zapisuje [issue #325](https://github.com/clickbytes/golingua/issues/325).

## Kontrole

| Polecenie/kontrola | Wynik |
| --- | --- |
| `gofmt -l` wszystkich śledzonych plików Go | PASS, brak plików do formatowania |
| `git diff --check` i lokalne linki zmienianych dokumentów | PASS |
| `go vet ./...` | PASS, exit 0 |
| `go build -o /tmp/golingua-light-f1 ./cmd/golingua` | PASS, exit 0 |
| `--help`, `--version` | PASS; wersja `0.1.0-alpha.1` |
| `catalog --summary --json`, `catalog --details --json` | PASS |
| `forms --language pl --lexeme lexeme:pl.pies.n --feature number=plural --feature case=genitive --feature gender=masculine_animal --json` oraz analiza `--surface psów` | PASS |
| `go test ./tests/... -run 'MeaningAttributes\|Identification\|Negation\|Seeing\|Inventory' -count=1 -v` | PASS, exit 0, sześć pakietów wykonujących testy |
| `go test ./tests/... -count=1 -v` | FAIL, exit 1, wyłącznie 36 SPEC_PENDING opisanych poniżej |

Testy uruchomiono z `GOLINGUA_MATRIX_SHA=99aefcd`. `MeaningAttributes` sprawdza niezależne I, analizę i realizację PL/EN bez parsera drugiego języka, JSON/hash/equality/projekcję, homografię, liczbę obu ról, negację, mass/plural_only, jawne założenia dla unknown oraz odmowy sprzecznego Evidence/kontekstu. Kontrprzykłady mają walidowane grafy, request references i EvidenceTrace oraz dodatnie kontrole tej samej trasy. Zachowano wcześniejsze kontrole Identification/Negation/Seeing i zamianę ról. Nie dodano testów językowych do CI.

Pełna macierz wykazuje wszystkie **24 oryginalne ID**. Brak zamkniętego niezależnego I daje SPEC_PENDING, nie udowodniony błąd analizatora:

| Rekordy | Liczba niezaliczonych kontroli |
| --- | --- |
| A1-LQ-001-EN-PL, A1-LQ-002-PL-EN, A1-LQ-003-EN-PL, A1-LQ-004-PL-EN | 12: analiza, realizacja i integracja |
| A1-PO-001-EN-PL, A1-PO-002-PL-EN, A1-PO-003-EN-PL, A1-PO-004-PL-EN | 12: analiza, realizacja i integracja |
| A1-Q-001-EN-PL, A1-Q-002-PL-EN, A1-Q-003-EN-PL, A1-Q-004-PL-EN | 12: analiza, realizacja i integracja |

Nie było innych niepowodzeń ani INFRA. To nadal nie jest zaliczenie pełnego A1.

## CLI i pośrednie znaczenia

[Obserwacje JSONL](325-observations.jsonl) zawierają dokładne polecenia, exit codes, SHA, kandydatury, grafy, teksty, założenia, straty i diagnostyki. Pominięto obszerne rekordy dowodów parsowania; zapis nie jest źródłem oczekiwań testów. Każdy wiersz jest osobnym obiektem JSON, a pierwszy opisuje metadane.

| Wejście | Zaobserwowany wynik |
| --- | --- |
| `The teacher reads the message.` | present, teacher=one; `Nauczyciel czyta wiadomość.` |
| `The teacher read the message.` | past, teacher=one; `Nauczyciel czytał wiadomość.` |
| `The teachers read the message.` | dwa różne I, present/past, teacher=many; `Nauczyciele czytają wiadomość.` i `Nauczyciele czytali wiadomość.` |
| `Nauczyciele czytają wiadomość.` | present/many; `The teachers read the message.` z jawną określonością |
| `Nauczyciele czytali wiadomość.` | past/many; `The teachers read a message.` i `The teachers read the message.` |
| `The teacher sees the dogs.` | teacher=one, dog=many; `Nauczyciel widzi psy.` |
| `Nauczyciele nie widzą psa.` | negacja i teacher=many; `The teachers do not see a dog.` i `The teachers do not see the dog.` |
| `The teachers do not see the dog.` | `Nauczyciele nie widzą psa.`, zachowana negacja i wielość |
| S080: `Nauczyciel dostał pieniądze.` / `The teacher got money.` | past; money bez multiplicity, poprawne lokalne formy pieniędzy/money |
| `The dog sees the teacher.` | role pozostają odwrócone; `Pies widzi nauczyciela.` |

Wszystkie cztery pierwotne wejścia identyfikacji zachowały uzgodnione teksty i wskazanie. Znane time/multiplicity nie dostają fałszywych założeń tense/number. Status `incomplete` może nadal oznaczać jawne założenie trzeciej osoby, rodzajnika albo stratę bliskości; `ambiguous` zachowuje rozróżnione odczytania lub licencjonowane warianty. Nie oceniano poprawności samym statusem.

Dodatkowe próbne `The teacher received money.` dało unsupported: istniejący korpus i trasa używają `get/got`. Nie dopisano nowej trasy ani nie zaliczono tego wejścia jako obsłużonego. Granicę ustalono z S080 i wcześniejszych kontraktów, bez zmiany semantycznych oczekiwań dla pieniędzy.

Typowane `question` z pustymi faktami i pojedynczymi referentami zwróciło unknown. Dodanie `temporal_frame=past` dało `UNSUPPORTED_QUESTION_TEMPORAL_FRAME`, a `multiplicity=many` — `UNSUPPORTED_QUESTION_MULTIPLICITY`, oba exit 3. Pełne wejścia tych trzech obserwacji są w JSONL. Reasoner nie zgubił pól ani nie uznał braku dowodu za fałsz.

## Powtórzenie CLI i TUI

Na wskazanym commicie przygotowano świeży runtime:

```sh
go build -o /tmp/golingua-light-f1 ./cmd/golingua
/tmp/golingua-light-f1 snapshot --apply --store /tmp/golingua-light-f1-final-store --json
```

Plik `/tmp/golingua-light-f1-final-runtime.json`:

```json
{"schema_version":1,"dataset_store":"/tmp/golingua-light-f1-final-store","dataset_id":"golingua-authored","language_package_dirs":[]}
```

Snapshot: `golingua-authored`, digest `0f0f631ce9455fc78f78075d68f1fd5d41dff452a68d096781b285f7481b2460`, status ok, aktywacja i trwałość potwierdzone. Polecenia tłumaczenia z JSONL można odtworzyć z tym runtime. Input dla `question` należy przekazać jako JSON na stdin.

Rzeczywista sesja TUI, exit 0:

```sh
TERM=xterm-256color XDG_CONFIG_HOME=/tmp/golingua-light-f1-config \
  script -q -f -c /tmp/golingua-light-f1 /tmp/light-f1-final-tui.typescript
```

F5: wpisanie powyższej ścieżki configu, Ctrl+S; widoczne `ok · ready: true`. F1: `The teacher read the message.` pokazało `Nauczyciel czytał wiadomość.`. Następnie `The teachers read the message.` pokazało dwie kandydatury: `Nauczyciele czytali wiadomość.` i `Nauczyciele czytają wiadomość.`. Alt+D i `Nauczyciele czytali wiadomość.` pokazały dwa warianty EN z `a/the message`. W szczegółach wybranej kandydatury były wyłącznie dwa założenia określoności, bez tense/number. Ctrl+C przywróciło terminal. SHA-256 surowego lokalnego transcriptu: `a1e2f824e24ed88161b29e6cbddb5c71fdbe550f38ff604d53d8db23f6c554c3`.

Zakres nie dodaje aspektu, nowych czasów, ogólnej koreferencji, pamięci ani uczenia. Dalsze naprawy proceduralnej odmiany, uogólniania i kosztu runtime pozostają osobnymi pozycjami w [rejestrze](README.md).

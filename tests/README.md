# Macierz językowa według poziomu i partii

Na polecenie właściciela z 2026-10-23 powstaje jawna macierz przypadków do realizacji. Pierwszym celem jest A1 dla PL↔EN. To mała próbka zachowań komunikacyjnych, nie pełne pokrycie ani certyfikacja CEFR. Jako kontekst doboru tematów wykorzystano [deskryptory Rady Europy](https://rm.coe.int/16809ea0d4); zdania i propozycje przekładów są własną redakcją, nie kopią materiału egzaminacyjnego.

## Układ

```text
tests/
  README.md
  A1/
    cases.go
    identification.testdomain.go
    possession.testdomain.go
    location_quantity.testdomain.go
    actions.testdomain.go
    negation.testdomain.go
    questions.testdomain.go
```

`A1/` jest pierwszą partią. Przyszłe `A1.1/`, `A1.2/` oznaczą kolejne partie tego samego celu A1, nie odrębne poziomy CEFR. Nie tworzymy ich przed określeniem zakresu. Nazwa domeny mówi, jakie zachowanie sprawdzamy; stabilne ID identyfikuje konkretny kierunek i wejście. Nowa partia nie zastępuje ani nie przepisuje starszych przypadków.

Pierwsza próbka zawiera 12 par zdań i 24 jawne rekordy kierunkowe. Obejmuje identyfikację/opis, posiadanie, lokalizację/liczbę, proste czynności i zamianę ról, przeczenie oraz pytania. Włącza dokładne wejście właściciela `This is a cat`, znaną działającą trasę widzenia i nowe konstrukcje, których silnik obecnie może nie obsługiwać. Nie dobieramy macierzy wyłącznie do istniejącej implementacji.

## Co opisuje rekord

- `ID`, `From`, `To`, `Input`: stabilna tożsamość, kierunek i dokładny tekst źródłowy.
- `MustPreserve`: wymagane zachowanie znaczenia, np. role uczestników, negacja, liczba albo pytanie zamiast stwierdzenia.
- `ReferenceTexts`: niepełna lista przykładowych propozycji przekładu. Nie jest listą wszystkich dopuszczalnych wyników ani automatyczną asercją równości napisu.
- `AmbiguityNote`: czego nie można rozstrzygnąć bez kontekstu. Przykładowo PL→EN nie wymusza jednego rodzajnika, a polskie „to” wymaga rozważenia `this`/`that`/`it`.
- `InputOrigin`, `ReferenceOrigin`: pochodzenie wejścia oraz oddzielne autorstwo propozycji. Wynik obecnego silnika nie staje się źródłem oczekiwania.
- `OwnerAcceptance`: akceptacja właściciela; początkowo `NOT_EVALUATED`. Dostarczenie samego wejścia nie oznacza zatwierdzenia proponowanego przekładu.

Brak obsługi wpisu jest luką do zrealizowania. Nie zmieniamy oczekiwania na `unsupported` tylko po to, żeby przypadek wyglądał na zaliczony. Nie wymuszamy `ok` ani jednego tekstu, gdy wejście dopuszcza kilka odczytań. Obserwację, ocenę agenta i akceptację właściciela zapisujemy oddzielnie.

## Odczyt i wykonanie

Ta pierwsza dostawa jest **macierzą danych do realizacji, bez runnera**. Pliki `*.testdomain.go` są zwykłymi kompilowanymi źródłami Go z rekordami; sam przyrostek nie uruchamia ich przez `go test`. Nie ma ukrytych pominięć, testów zawsze zielonych ani automatycznej akceptacji propozycji.

Można obejrzeć konkretną domenę standardowym narzędziem Go:

```sh
go doc -src ./tests/A1 IdentificationCases
go doc -src ./tests/A1 QuestionCases
```

Przypadek wykonuje się dziś jawnym CLI z wybranym aktywnym runtime’em, np.:

```sh
go run ./cmd/golingua translate --from en --to pl --text 'This is a cat' --runtime-config .local/runtime.json --json
```

Ścieżka `.local/runtime.json` jest lokalną konfiguracją użytkownika, nie częścią macierzy; przygotowanie snapshotu i konfiguracji opisuje [CLI](../docs/cli.md). Zapisuj rzeczywisty status, tekst/kandydatury, diagnostyki, exit code i SHA, bez nadpisywania propozycji przez aktualny output. Format opisu obserwacji jest w [procedurze przeglądu](../docs/manual-review.md).

Decyzja o wykonywanych testach i ich asercjach jest osobnym krokiem. Obecna próbka nie dodaje `*_test.go`, frameworka, automatycznej bramki językowej w CI, procentowych metryk ani porównania z JS. Istniejące gofmt/vet/build sprawdzają tylko poprawność techniczną pakietu.

## Rzeczywiste obserwacje startowe

Silnik: `0e2933c911f9b86ba92d53740a7dde310d55f9e9`. Kompilowany lokalnie jako `/tmp/golingua-unsupported-analysis`; aktywna konfiguracja `/home/ppotepa/git/work/golingua/.local/runtime.json`. Zapis nie twierdzi, że wykonano całą macierz.

| ID | Dokładne wejście EN→PL | Wynik silnika / exit | Ocena agenta | Akceptacja właściciela |
| --- | --- | --- | --- | --- |
| A1-ID-001-EN-PL | `This is a cat` | `unsupported` / 3; brak tekstu, `UNKNOWN_SURFACE` dla This/is, `NO_START_PLAN`, `UNMATCHED_PARSE_CANDIDATE` | BRAK OBSŁUGI konstrukcji identyfikacyjnej | NOT_EVALUATED |
| A1-AC-001-EN-PL | `The teacher sees the dog.` | `ok` / 0; `Nauczyciel widzi psa.` | przekład zachowuje role i sens na tym wejściu | NOT_EVALUATED |

Pozostałe rekordy tej próbki nie zostały wykonane w ramach #309: **NOT RUN**. Historyczne pochodzenie części wejść nie jest nowym uruchomieniem ani akceptacją właściciela. Pełne surowe JSON z dwóch powyższych wywołań zapisano lokalnie jako `/tmp/golingua-a1-owner-observed.json` i `/tmp/golingua-a1-existing-observed.json`; są dowodami sesji, nie fixtures produktu. Zlecenie i końcowe wyniki kontroli technicznych: [issue #309](https://github.com/clickbytes/golingua/issues/309).

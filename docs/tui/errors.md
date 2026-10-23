# Statusy, diagnostyki i błędy TUI

TUI prezentuje status engine jako część wyniku. Kolor może wspierać rozróżnienie, lecz każdy status ma także tekstową etykietę. Interfejs nie podnosi ani nie obniża statusu na podstawie tego, czy znaleziono tekst.

| Status | Prezentacja |
| --- | --- |
| `ok` | Pokaż tekst i rzeczywiste evidence. `ok` opisuje wykonanie obecnego pipeline'u, nie odbiór językowy. |
| `ambiguous` | Pokaż każdą zwróconą kandydaturę i jej ID. Nie wybieraj jednej bez działania użytkownika. |
| `incomplete` | Pokaż dostępny tekst kandydata, założenia i nierozstrzygnięte evidence. Nie zmieniaj statusu po podglądzie. |
| `unsupported` | Powiedz, że dla tej trasy brak obsługi; nie twórz tekstu zastępczego. |
| `unknown` | Pokaż, którego leksemu, konceptu lub danych nie znaleziono; nie utożsamiaj tego z brakiem całego języka. |
| `invalid` | Wskaż diagnostykę i pole wejściowe, jeśli istnieje; pozwól poprawić formularz. |
| `resource_limit` | Powiedz, że przekroczono jawny limit i wynik nie został obcięty. |
| `error` | Pokaż bezpieczny komunikat techniczny i identyfikator diagnostyki; szczegóły są w Technical. |

## Diagnostyki i podgląd

Simple pokazuje status i komunikat. Normal pokazuje candidate listę, jawne założenia i istotne ostrzeżenia. Technical dodaje surowe diagnostics/evidence i parametry żądania przydatne w ręcznym CLI. Poziom zmienia tylko widoczność, nigdy żądanie, selekcję, ranking ani status.

Podgląd kandydata wybiera rzeczywiste ID kandydata. Jest to prezentacja alternatywy, nie zatwierdzenie jej językowej poprawności. Odbiór agenta i akceptacja właściciela pozostają osobnymi zapisami.

## Błędy TUI

- Nieprawidłowy runtime config lub brak aktywnego snapshotu nie zamienia się w `unsupported` udający wynik translacji; pokaż diagnostykę konfiguracji.
- Nieczytelny plik preferencji powoduje ostrzeżenie i wartości domyślne. Nie nadpisuj pliku bez jawnego `Save`.
- Niepowodzenie `Preview` lub `Apply` pokazuje rzeczywisty błąd stagingu. `Apply` nie jest ponawiane w tle.
- Błąd terminala lub brak TTY kończy TUI bez ANSI w stdout i bez pozostawienia terminala w raw mode.

## Kody wyjścia CLI

| Kod | Znaczenie |
| ---: | --- |
| 0 | `ok`; wyjątek: `capabilities` bez runtime też kończy się 0 z `ready:false` |
| 1 | `error` |
| 2 | `invalid` / składnia żądania |
| 3 | `unsupported` |
| 4 | `resource_limit` |
| 5 | `unknown` |
| 6 | `ambiguous` |
| 7 | `incomplete` |

TUI nie mapuje statusów na kod procesu CLI, bo nie uruchamia procesu potomnego. Kod tabeli dotyczy istniejących poleceń CLI.

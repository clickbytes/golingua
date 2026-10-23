# 04 — Wyniki, diagnostyki i limity

Ten dokument opisuje kontrakt wyników Go oraz jego ograniczone użycie przez parser, morfologię, leksykę i translację PL/EN. Skompilowany runtime obsługuje obecnie jawne szlaki `seeing` i `hearing` w teraźniejszej 3. osobie liczby pojedynczej; pozostałe trasy zwracają odpowiedni status braku obsługi lub nierozstrzygnięcia.

## Wynik operacji

`internal/status.Result` jest wspólną strukturą wyniku. Jej pola `status`, `text`, `diagnostics` i opcjonalny `trace` mają ograniczony rozmiar i walidację. Każdy status inny niż `ok` wymaga co najmniej jednej diagnostyki; nie da się utworzyć wyniku błędnego z pustym wyjaśnieniem przez `status.NewResult`.

| Status | Znaczenie |
| --- | --- |
| `ok` | Operacja została wykonana; sam status nie twierdzi nic o jakości językowej. |
| `unsupported` | Żądana operacja lub konstrukcja nie jest obsługiwana. |
| `unknown` | Dostępne dane nie rozstrzygają odpowiedzi; brak informacji nie jest fałszem. |
| `ambiguous` | Pozostaje więcej niż jedna dopuszczalna interpretacja. |
| `incomplete` | Przetwarzanie nie dało kompletnego wyniku; częściowy `text` nie jest pełnym sukcesem. |
| `invalid` | Żądanie lub dane nie spełniają wymaganego formatu. |
| `resource_limit` | Jawny limit wejścia lub pracy został przekroczony. To osobny wynik, nie odmowa z powodu braku obsługi. |
| `error` | Wystąpił błąd wykonania lub infrastruktury. |

Skrócony przykład obserwowanego wyjścia `translate` PL→EN. Status `incomplete` zachowuje tekst, ale ujawnia jawne założenia dotyczące angielskich rodzajników:

```json
{"status":"incomplete","from":"pl","to":"en","candidates":[{"outcome":"incomplete","text":"The teacher sees the dog."}]}
```

Wyjście błędnej komendy:

```json
{"status":"invalid","diagnostics":[{"code":"INVALID_COMMAND","message":"unknown or invalid command","path":"command","severity":"error"}]}
```

CLI wypisuje te wyniki na stdout jako JSON. Nieznana komenda i nieobsługiwana translacja nie są zgłaszane jako sukces. Jeżeli samo zapisanie JSON do stdout zawiedzie, CLI zwraca kod infrastruktury 1 i zapisuje krótką diagnostykę na stderr; w takiej sytuacji stdout nie może być wiarygodnym kanałem JSON.

## Diagnostyki

`Diagnostic` zawiera stabilny `code`, czytelny `message`, opcjonalny `path` i `severity` (`info`, `warning`, `error`). Kody używają wielkich liter, cyfr i podkreśleń. Każdy wynik może nieść najwyżej 16 diagnostyk; kod ma maksymalnie 64 bajty, wiadomość 2048 bajtów, a ścieżka 512 bajtów. Wartości muszą być poprawnym UTF-8.

## Budżety

`status.DefaultBounds()` zwraca następujące wartości początkowe:

| Budżet | Limit |
| --- | ---: |
| Wejście UTF-8, mierzone w bajtach | 16 KiB |
| Głębokość grafu | 64 |
| Liczba analiz/kandydatów | 256 |

`Bounds.Validate` wymaga dodatnich limitów; zero nie oznacza „bez limitu”. `CheckInputText` sprawdza UTF-8 i limit bajtów; niepoprawny UTF-8 zwraca `ErrInvalidInput`. Metody `CheckInputBytes`, `CheckGraphDepth` i `CheckAnalyses` zwracają `LimitExceededError` po przekroczeniu. `LimitExceededError.Result()` daje status `resource_limit` i diagnostykę `RESOURCE_LIMIT`. Ujemna wartość albo niepoprawna konfiguracja budżetu jest błędem konfiguracji, a nie przekroczeniem limitu.

Budżety są używane przez operacje runtime. `translate` bez `--runtime-config` zwraca `unsupported`, a z aktywnym snapshotem uruchamia wyłącznie skonfigurowane, skompilowane adaptery.

## Ślad pochodzenia

Opcjonalny `Trace` może zawierać najwyżej 32 kroki, a każdy krok musi wskazywać przynajmniej jeden `RuleID`, `SenseID` lub `ConstructionID`. Identyfikator ma maksymalnie 128 bajtów. `Truncated` jawnie oznacza pominięcie dalszych kroków; ślad nie może zastąpić brakującego dowodu ani uzasadniać interpretacji samą nazwą słowa.

## CLI i kody wyjścia

| Kod | Wynik |
| ---: | --- |
| 0 | `ok` albo poprawne zapytanie o capabilities/help/version |
| 1 | `error` lub błąd zapisu/serializacji |
| 2 | `invalid` albo niepoprawne użycie CLI |
| 3 | `unsupported` |
| 4 | `resource_limit` |
| 5 | `unknown` |
| 6 | `ambiguous` |
| 7 | `incomplete` |

`capabilities --json` bez wybranego runtime'u zwraca pole `status: unsupported`, puste listy języków i funkcji oraz `ready: false`; komenda sama kończy się kodem 0. `--help` i `--version` pozostają krótkimi odpowiedziami tekstowymi. `translate` bez aktywnego runtime'u zwraca `unsupported` i kod 3.

Nie ma tu goldów, testów automatycznych jakości językowej ani bramki procentowej. Zapis obserwacji CLI dla redakcyjnych 100 zdań nie jest oceną poprawności. Ręczna ocena pozostaje `NOT RUN`, dopóki właściciel nie oceni konkretnych wyników.

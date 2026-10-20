# 04 — Wyniki, diagnostyki i limity

Ten dokument opisuje **zaimplementowany, generyczny kontrakt wyników Go**, a nie działającą obsługę tłumaczeń. Parser, morfologia, leksyka i translacja nadal nie są zaimplementowane. `translate` musi zatem zwracać `unsupported`.

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

Przykładowe wyjście `translate` z obecnego bootstrapu:

```json
{"status":"unsupported","diagnostics":[{"code":"LANGUAGE_RUNTIME_UNAVAILABLE","message":"sentence translation is not implemented in the bootstrap","path":"translate","severity":"error"}]}
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

Obecne CLI nie przyjmuje jeszcze tekstu do analizy. Powyższe budżety są zaimplementowanym, typowanym API dla przyszłych operacji runtime; nie oznaczają, że nieistniejący parser lub translator już ich używa. `translate` na tym etapie zwraca `unsupported` bez przetwarzania argumentów jako tekstu językowego.

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

`capabilities --json` opisuje dostępne możliwości procesu: komenda kończy się kodem 0, a pole `status: unsupported` wewnątrz odpowiedzi oznacza, że żadne języki ani funkcje językowe nie są obecnie dostępne. `--help` i `--version` pozostają krótkimi odpowiedziami tekstowymi. `translate` zwraca wynik JSON ze statusem `unsupported` i kodem 3.

Nie ma tu goldów, runnera, testów automatycznych jakości językowej ani bramki procentowej. Ręczna ocena tłumaczeń pozostaje nieoceniona, dopóki właściciel nie poda rzeczywistego wejścia.

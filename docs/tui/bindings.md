# Bindingi TUI → typowane usługi Go

TUI i CLI używają jednej logiki aplikacyjnej. CLI nie jest uruchamiane jako proces potomny TUI. `internal/app.Service` udostępnia wspólne, typowane operacje; adaptery ekranów mogą używać tych metod bez parsowania JSON-u CLI.

| Ekran / działanie | Źródło prawdy | Reguła bindingu |
| --- | --- | --- |
| F1 Translate | `Service.Translate(ctx, TranslateRequest)` → `TranslationResponse`; runtime ładowany przez wstrzyknięty `internal/runtime` adapter | Nie omijaj walidacji runtime'u ani stage adapterów. Pokaż status, candidate IDs, assumptions i diagnostics tak, jak zwraca engine. Błędy przygotowania runtime'u pozostają osobnym `status.Result`. |
| F2 Analyze / Generate | `Service.Forms(ctx, FormsRequest)` → `FormsResult`, delegowane do seedów `languages/pl` lub `languages/en` | Zachowaj wszystkie zgodne odczytania i ich cechy; UI nie wybiera „najbardziej prawdopodobnej” formy. Przekazuj wyłącznie jawne cechy profilu języka. |
| F3 Details | `Service.LookupConcept(ctx, ConceptID)` → `core.Concept` i informacja o trafieniu | Definicja i relacje pojawiają się tylko po prawdziwym lookupie w skompilowanym grafie. |
| F3 Coverage / F4 inventory | `Service.Catalog(ctx, CatalogRequest)` → `CatalogResult`, zbudowany przez `catalog.Inspect(graph, polish, english)` | Liczniki oznaczają inwentarz katalogu, a nie gotowość runtime'u ani ocenę jakości. |
| F4 Runtime state | wynik wybranego runtime config i ograniczone capabilities | `capabilities` bez runtime ma `ready:false`; nie wyciągaj z tego wniosku, że katalog PL/EN jest pusty. |
| F5 Preview / Apply | `Service.Snapshot(ctx, SnapshotRequest)` → `SnapshotResult`; kandydat i store pochodzą z `bundled`/`staging` | Preview jest bez zapisu. `Apply` wymaga `Confirmed: true`; warstwa UI musi najpierw uzyskać jawne potwierdzenie użytkownika. |

## Kierunek zależności

```text
cmd/golingua
  ├── internal/cli ───┐
  └── internal/tui ───┴──> internal/app ─── engine / languages / catalog / core / staging
                              └── injected runtime loader
```

`internal/app` nie importuje `internal/cli` ani `cmd/golingua`. `internal/tui` zależy od typowanych żądań/wyników, a nie od JSON-u CLI. `internal/cli` zachowuje stdout, stderr, składnię i kody procesu.

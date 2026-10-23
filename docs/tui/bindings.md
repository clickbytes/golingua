# Bindingi TUI → typowane usługi Go

TUI i CLI używają jednej logiki aplikacyjnej. CLI nie jest uruchamiane jako proces potomny TUI. G11-02 dostarczy `internal/app` jako in-process fasadę nad obecnymi usługami; nazwy metod adaptera zostaną ustalone w tym etapie. Poniżej są źródła prawdy i granice, nie obietnica, że adapter już istnieje.

| Ekran / działanie | Źródło prawdy | Reguła bindingu |
| --- | --- | --- |
| F1 Translate | `engine.Engine.Translate(engine.Request)` z konfiguracją wczytaną przez `internal/runtime` | Nie omijaj walidacji runtime'u ani stage adapterów. Pokaż status, candidate IDs, assumptions i diagnostics tak, jak zwraca engine. |
| F2 Analyze | `languages/pl` lub `languages/en` morphology seed: `Analyze(surface)` | Zachowaj wszystkie zgodne odczytania i ich cechy; UI nie wybiera „najbardziej prawdopodobnej” formy. |
| F2 Generate | morphology seed: `Generate(lexemeID, featureValues)` | Przekazuj wyłącznie jawne cechy profilu języka. Brak zadeklarowanej komórki pozostaje widoczny. |
| F3 Details | `core.ConceptGraph.Lookup(ConceptID)` | Definicja i relacje pojawiają się tylko po prawdziwym lookupie w skompilowanym grafie. |
| F3 Coverage / F4 inventory | `catalog.Inspect(graph, polish, english)` | Liczniki oznaczają inwentarz katalogu, a nie gotowość runtime'u ani ocenę jakości. |
| F4 Runtime state | wynik wybranego runtime config i ograniczone capabilities | `capabilities` bez runtime ma `ready:false`; nie wyciągaj z tego wniosku, że katalog PL/EN jest pusty. |
| F5 Preview / Apply | `bundled.Candidate`, `staging.OpenStore`, store `Preview` / `Apply` | Preview jest bez zapisu. Apply jest osobną operacją i wymaga jawnego potwierdzenia użytkownika. |

## Kierunek zależności

```text
cmd/golingua
  ├── internal/cli ───┐
  └── internal/tui ───┴── internal/app ─── engine / languages / catalog / staging
                                               └── runtime / core
```

To diagram odpowiedzialności. G11-02 ma doprecyzować implementację i udostępnić adaptery bez przesuwania reguł językowych do CLI/TUI. `internal/tui` zależy od typowanych żądań/wyników, a nie od JSON-u CLI. `internal/cli` zachowuje stdout, stderr, składnię i kody procesu.

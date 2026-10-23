# Bindingi TUI do typowanych usług Go

CLI i TUI używają wspólnego `internal/app.Service`; TUI nie uruchamia CLI jako
procesu potomnego ani nie parsuje jego JSON-u.

| Ekran/działanie | Kontrakt usługi | Zasada prezentacji |
| --- | --- | --- |
| F1 Translate i Preview kandydata | `Service.Translate(ctx, TranslateRequest)` → `TranslationResponse` | Pokazuj zwrócony status, kandydatów, ID, założenia i diagnostyki. Preview ponawia zachowane żądanie z jawnym zwróconym PreviewID; nie zmienia pierwotnej odpowiedzi. |
| F2 Analyze / Generate | `Service.Forms(ctx, FormsRequest)` → `FormsResult` | Zachowaj wszystkie zwrócone odczytania/formy i ich cechy. Przekazuj tylko jawne wejście; nie wnioskuj o cechach językowych. |
| F3 katalog i szczegóły | `Service.Catalog(ctx, CatalogRequest{Details:true})`; `Service.LookupConcept(ctx, ConceptID)` → `core.Concept` i informacja o trafieniu | Filtruj lokalnie i dosłownie po ConceptID. Szczegóły pochodzą wyłącznie z rekordu zwróconego przez LookupConcept. |
| F4 inwentarz | `Service.Catalog(ctx, CatalogRequest{})` | Pokazuj rzeczywiste liczniki leksemów, sensów, bindingów i pokrycia PL/EN niezależnie od gotowości runtime'u. |
| F4 gotowość runtime'u i F5 Validate | `Service.ValidateRuntime(ctx, RuntimeValidationRequest)` | F4 sprawdza wybraną ścieżkę runtime'u, a F5 szkic ścieżki. Poprawnie wczytana konfiguracja nie obiecuje obsługi każdego zdania. Walidacja F5 nie wybiera ani nie zapisuje szkicu. |
| F5 Snapshot Preview / Apply | `Service.Snapshot(ctx, SnapshotRequest)` → `SnapshotResult` | Preview używa jawnej ścieżki store ze szkicu i może utworzyć katalogi; nie aktywuje danych. Apply wymaga potwierdzenia, przekazuje zatwierdzony digest w `ExpectedDigest`, a usługa sprawdza go przed aktywacją. Niezgodność kończy się bez Apply. |

Zmiana ścieżki store w F5 unieważnia wcześniejszy Preview. Snapshot Apply nie
zmienia wybranego runtime'u ani nie zapisuje ustawień. Preferencje stają się
aktywne dopiero po udanym, atomowym Save.

## Kierunek zależności

```text
cmd/golingua
  ├── internal/cli ───┐
  └── internal/tui ───┴──> internal/app ─── engine / languages / catalog / core / staging
                              └── injected runtime loader
```

`internal/app` nie importuje `internal/cli` ani `cmd/golingua`. stdout, stderr,
składnia komend i kody wyjścia CLI pozostają osobnym kontraktem.

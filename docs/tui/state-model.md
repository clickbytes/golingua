# Model stanu, zdarzeń i fokusu TUI

Ten dokument opisuje model MVU zaimplementowanego TUI GoLingua. TUI udostępnia
pięć ekranów i korzysta z typowanych metod `internal/app.Service`; brak TTY
zachowuje ścieżkę nieinteraktywnego CLI.

## Własność stanu

Model główny posiada stan powłoki, wybór ekranu, rozmiar terminala i cykl życia
aplikacji. Każdy ekran posiada własny szkic wejściowy, wynik, selekcję i
identyfikatory oczekujących operacji. Przełączenie ekranu nie kasuje roboczego
stanu pozostałych ekranów.

| Właściciel | Stan |
| --- | --- |
| Powłoka | aktywny ekran F1–F5, właściciel i cel fokusu, rozmiar terminala, globalne komunikaty oraz stan startu/zamykania |
| F1 Translate | kierunek, edytowany tekst, `preview_id`, wynik i rewizja `TranslationResponse`, osobne tokeny Translate/Preview |
| F2 Forms | tryb Analyze/Generate, język, forma albo LexemeID i cechy, `FormsResult` oraz token |
| F3 Concepts | filtr, wybrany `ConceptID`, `CatalogResult`, szczegóły `LookupConcept` i tokeny zapytań |
| F4 Languages | inwentarz `CatalogResult` i osobno wynik walidacji wybranego runtime'u |
| F5 Settings | aktywne preferencje i szkic, walidacja runtime'u, stan zapisu, ścieżka store, wynik snapshotu i dialog |

Model przechowuje dokładne wyniki i statusy usług. Nie wybiera najlepszego
kandydata, nie zmienia statusu i nie uzupełnia brakujących danych. Stan
prezentacji, taki jak fokus, przewinięcie i rozwinięte szczegóły, nie trafia do
`engine`, katalogu ani danych językowych.

`internal/app.Service` jest źródłem operacji: `Translate` zwraca
`TranslationResponse`, `Forms` — `FormsResult`, `Catalog` — `CatalogResult`,
`LookupConcept` — rekord i informację o trafieniu, a `Snapshot` —
`SnapshotResult`. Model TUI zachowuje te wartości jako typowane wyniki.
Warstwa terminala nie uruchamia procesu CLI i nie parsuje jego JSON-u.

## Zdarzenia i aktualizacja

Warstwa zdarzeń zamienia wejście terminala i zakończenie pracy usług na
wiadomości dla jednego modelu głównego. Aktualizacja modelu jest synchroniczna;
wywołania serwisu, ładowanie ustawień i zapis działają poza funkcją aktualizacji
i zwracają osobną wiadomość zakończenia.

| Rodzaj wiadomości | Działanie modelu |
| --- | --- |
| `ScreenSelected(screen)` | Ustaw aktywny ekran; zachowaj szkice, wyniki, selekcję i fokus nieaktywnego ekranu |
| `FocusMoved(owner, target)` | Zapisz jawnego właściciela i cel fokusu; fokus nie wynika wyłącznie z koloru ani pozycji kursora |
| `InputChanged(screen, field, revision)` | Zaktualizuj szkic i zwiększ rewizję wejścia tego ekranu |
| `OperationSubmitted(screen, operation, revision, request)` | Zachowaj niezmienny, typowany request, nadaj token i oznacz tę operację jako oczekującą |
| `ServiceCompleted(token, result, error)` | Skoreluj odpowiedź z ekranem i operacją; zastosuj tylko do pasującego, oczekującego tokenu |
| `TerminalResized(width, height)` | Zapisz rozmiar i przelicz układ bez odrzucania stanu formularzy |
| `SettingsLoaded(result)` / `SettingsSaved(result)` | Zaktualizuj wyłącznie preferencje oraz widoczny stan ich odczytu/zapisu |
| `QuitRequested` / `ShutdownCompleted` | Przejdź do kontrolowanego zamknięcia i zakończ po przywróceniu terminala |

Zdarzenia klawiatury są tłumaczone na wybór ekranu, fokusu albo akcję
posiadaną przez aktualny fokus. Bieżące kolejności kontrolek są: F1
prompt/direction/candidates/preview/result/tabs; F2
prompt/language/mode/[features]/run/result/tabs; F3 prompt/list/result/tabs;
F4 refresh/result/tabs; F5 prompt/validate/level/default_direction/store/
snapshot_preview/snapshot_apply/save/cancel/result/tabs. W dialogu snapshotu
fokus obejmuje tylko Cancel i Apply.

## Tożsamość odpowiedzi asynchronicznej

Każde wywołanie usługi otrzymuje unikalny, rosnący `request_id` w obrębie
uruchomienia TUI oraz metadane: ekran, operację i rewizję wysłanego szkicu.
Zadanie asynchroniczne przechowuje kopię requestu i ten token. Nie modyfikuje
modelu bezpośrednio.

- Oczekujący token jest śledzony per ekran i operację, nie jako jeden globalny
  „ostatni request”. Wolna translacja F1 nie blokuje wyniku Forms F2.
- Odpowiedź aktualizuje właścicielski stan tylko wtedy, gdy jej token nadal
  pasuje do oczekującej operacji. Nowsza operacja tego samego rodzaju zastępuje
  starszą; spóźniona odpowiedź starszej operacji jest ignorowana.
- Zmiana ekranu sama w sobie nie unieważnia pracy. Pasujący wynik może uzupełnić
  nieaktywny ekran, a użytkownik zobaczy go po powrocie.
- Jeżeli użytkownik edytował szkic po wysłaniu requestu, wynik zachowuje
  rewizję wysłanego wejścia i jest oznaczony jako odnoszący się do starszego
  szkicu. Nie przedstawiaj go jako wyniku aktualnej treści.
- Zmiana konfiguracji runtime'u zwiększa jego rewizję. Odpowiedź z poprzedniej
  rewizji nie może nadpisać stanu bieżącej konfiguracji.
- Wynik `error`, anulowanie lub błąd dekodowania kończy oczekującą operację i
  trafia do jej ekranu. Nie jest po cichu zamieniany w `unsupported` ani `ok`.

Schemat korelacji:

```mermaid
sequenceDiagram
    actor User
    participant Root as Root Update
    participant F1 as F1 state
    participant F2 as F2 state
    participant App as app.Service
    User->>Root: Submit F1 revision 8
    Root->>F1: pending token 41
    Root->>App: Translate(request 41)
    User->>Root: Submit F2 revision 3
    Root->>F2: pending token 42
    Root->>App: Forms(request 42)
    App-->>Root: Completed(token 42, result)
    Root->>F2: apply when token still matches
    App-->>Root: Completed(token 41, result)
    Root->>F1: apply independently when token still matches
```

Jeśli ten sam ekran wysłał nowsze zapytanie F1 tokenem 43, zakończenie tokenu
41 jest przestarzałe i nie zmienia wyniku F1. Nie odrzuca to niezależnego
tokenu 42 dla F2.

## Właściciel fokusu i przejścia

Fokus jest jawną parą właściciela i celu, na przykład `(F2, surface)` albo
`(shell, screen_tabs)`. Model powłoki przechowuje aktywny cel; każdy ekran
przechowuje ostatni poprawny fokus lokalny. Pasek kart i globalny prompt są
kontrolkami powłoki. Pola, listy i akcje są kontrolkami ekranu.

- Po starcie fokus wskazuje widoczne, interaktywne pole albo pierwszy element
  zgodny z układem i rozmiarem terminala.
- F1–F5 wybierają ekran bez kasowania lokalnego fokusu i szkicu. Po powrocie
  ekran odzyskuje swój poprzedni cel fokusu, jeśli nadal istnieje; w przeciwnym
  razie wybierany jest jego pierwszy dostępny cel.
- `Tab` i `Shift-Tab` przechodzą po jawnej kolejności kontrolek. Aktywna kontrolka
  jest zawsze widoczna także bez koloru. Szczegółową kolejność definiują
  keymapa i specyfikacja danego ekranu.
- `Enter` wysyła akcję do właściciela fokusu. `Esc` cofa modalny/podwidokowy
  stan zgodnie z keymapą i nigdy sam nie zatwierdza zmian.
- Zmiana rozmiaru aktualizuje układ, lecz nie przenosi fokusu na nieistniejący
  element i nie usuwa szkicu. Jeśli bieżący element znika w układzie zwartym,
  wybierz jawnie dostępny cel tego samego ekranu.
- Dialog potwierdzenia snapshotu jest właścicielem fokusu do czasu zatwierdzenia
  albo anulowania. Dialog pokazuje ścieżkę store i digest podglądanego
  kandydata. Przed zastosowaniem UI odświeża podgląd dla tej ścieżki; jeśli
  digest się zmienił, ponownie pyta o potwierdzenie. Anulowanie nie wywołuje
  `Snapshot(Apply:true)`. Po potwierdzeniu do `SnapshotRequest` trafia ta sama
  ścieżka oraz `Confirmed: true`; ekran pokazuje digest zwrócony przez Apply.

## Cykl życia

1. Inicjalizacja powłoki sprawdza warunki interaktywnego terminala przed wejściem
   w tryb interaktywny. Jawne CLI/help/version i brak TTY zachowują ścieżkę CLI.
2. Start TUI wczytuje wersjonowane preferencje i żąda szczegółowego katalogu
   konceptów. Nie stosuje snapshotu ani nie wykonuje tłumaczenia bez działania
   użytkownika. Wejście na F4 odświeża inwentarz i walidację runtime'u.
3. Zdarzenia użytkownika aktualizują wyłącznie model przypisany do ich
   właściciela. Dłuższa praca zwraca wiadomość z tokenem.
4. Kontrolowane wyjście przywraca terminal. Gdy snapshot Apply już trwa,
   Ctrl+C/Ctrl+D czeka na wynik tej operacji.

Zmiana ekranu nie anuluje poprawnej pracy. Odpowiedź może uzupełnić nieaktywny
ekran, jeśli token, rewizja i przechwycona ścieżka nadal pasują. Zmiana ścieżki
store unieważnia podgląd; zmiana runtime'u unieważnia wcześniejszą walidację.
Przewinięcie, filtr i selekcja należą do ekranu i przetrwają resize.

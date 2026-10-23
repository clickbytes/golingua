# Prompt i slash commands

Ten dokument opisuje routing wspólnego dolnego edytora w aktualnym TUI. Cel
wpisu zależy od ekranu i fokusu; TUI nie uruchamia procesu CLI i nie stanowi
funkcji czatu.

## Składnia

Wejście promptu jest interpretowane według pierwszego znaku:

| Wejście | Zachowanie |
| --- | --- |
| `Ctrl+P` | Otwórz paletę nawigacji z dowolnego ekranu |
| początkowy `/` w promptcie F1–F3 | Otwórz paletę nawigacji zamiast wstawiać znak do szkicu |
| nazwa lub prefiks w otwartej palecie | Filtruj zarejestrowane trasy; Enter wybiera dokładną nazwę lub zaznaczoną sugestię |
| tekst bez początkowego `/` | Przekaż do aktywnego pola tekstowego zgodnie z routingiem ekranowym |

Nazwa komendy składa się z małych liter ASCII, cyfr i myślników; zaczyna się
małą literą. Białe znaki ani argumenty nie są częścią komendy w tej wersji.
`/translate --from en` jest niepoprawnym poleceniem TUI, a parser CLI nigdy
nie jest dla niego wywoływany. Wielkość liter musi odpowiadać zarejestrowanej
nazwie. Nieznany lub źle uformowany slash command pokazuje błąd wejścia w
powłoce, nie status tłumaczenia.

## Rejestr i kontekst

Rejestr globalny zawiera wyłącznie nawigację do pięciu ekranów:

| Komenda | Ekran | Zachowanie po wyborze |
| --- | --- | --- |
| `/translate` | F1 Translate | Aktywuj F1 i przywróć jego ostatni poprawny fokus |
| `/forms` | F2 Forms | Aktywuj F2; zachowaj wybrany tryb Analyze/Generate |
| `/concepts` | F3 Concepts | Aktywuj F3 i zachowaj filtr oraz selekcję |
| `/languages` | F4 Languages | Aktywuj F4 i zachowaj jego fokus |
| `/settings` | F5 Settings | Aktywuj F5 i zachowaj niezapisane preferencje |

Komendy są dostępne z każdego ekranu; paleta oznacza bieżący ekran i może
ustawić go jako pierwszą podpowiedź, ale nie ukrywa innych ekranów. Inicjalna
kolejność to F1–F5. G11-04 nie dodaje slash commands uruchamiających operacje
silnika, zapisujących preferencje ani stosujących snapshot. Te działania
pozostają jawnymi kontrolkami właściwych ekranów.

## Paleta, wybór i błędne wejście

- Otwarta paleta filtruje rejestr po wpisanym prefiksie; filtr jest prosty i
  deterministyczny, bez zgadywania ani fuzzy matching.
- Dokładna nazwa jest oznaczona jako gotowa do wyboru, ale wykona się dopiero
  po `Enter`. Przy wielu wynikach `↑`/`↓` zmienia zaznaczenie, a `Enter`
  wybiera zaznaczoną komendę. Niepełny prefiks sam nie przełącza ekranu.
- Wybranie komendy nawigacyjnej wysyła `ScreenSelected(screen)` do modelu z
  [kontraktu stanu](state-model.md). Prompt czyści tylko własny tekst komendy;
  zachowuje szkice, wyniki i fokus ekranów.
- `Esc` zamyka paletę bez akcji i wraca do poprzedniego celu fokusu. `Tab` /
  `Shift-Tab` zachowują globalny kontrakt fokusu z [mapy klawiszy](keymap.md).
- `Enter` dla pustej lub niejednoznacznej selekcji nie wykonuje komendy.
  Nieznane `/name`, niedozwolone argumenty lub niepoprawny identyfikator
  pokazują krótki komunikat z zachowanym tekstem, aby można go poprawić.
  Błąd nie wywołuje usługi ani nie modyfikuje ekranu.
- Lista sugestii pokazuje tylko komendy z globalnego rejestru. Puste wyniki
  pozostawiają paletę otwartą i umożliwiają dalsze wpisywanie lub anulowanie.

## Zwykły tekst i routing ekranowy

Tekst bez początkowego `/` trafia do aktywnego celu tekstowego. Przejście
klawiaturą do innego pola zmienia cel routingu; sama zmiana ekranu zachowuje
jego poprzedni fokus. Wysyłanie requestu następuje akcją ekranu, nigdy samym
wpisaniem tekstu.

| Ekran | Odbiorca zwykłego tekstu |
| --- | --- |
| F1 Translate | Edytowany tekst źródłowy; `Enter` uruchamia `Service.Translate` z kierunkiem wybranym na ekranie |
| F2 Forms / Analyze | Pole formy powierzchniowej; `Enter` uruchamia `Service.Forms` w trybie Analyze |
| F2 Forms / Generate | LexemeID w prompcie i jawne pary `key=value` w features; Enter na polu albo run uruchamia Generate |
| F3 Concepts | Dosłowny filtr podciągu wyłącznie po ConceptID, z rozróżnieniem wielkości liter; Enter na prompt/list/result wywołuje LookupConcept dla aktualnie wybranego ID |
| F4 Languages | Brak pola tekstowego; Enter na Refresh ponawia niezależne Catalog inventory i ValidateRuntime |
| F5 Settings | Edytowane pole wskazane fokusem: runtime config albo snapshot store. Validate sprawdza szkic runtime'u; Preview, Apply, Save i Cancel są osobnymi kontrolkami |

Zwykły tekst nigdy nie staje się rozmową z agentem, zapytaniem do sieci,
argumentem CLI ani poleceniem powłoki. Katalog snapshotu jest jawnie podaną
ścieżką; Preview może utworzyć katalogi, ale nie aktywuje snapshotu. Apply ma
domyślną akcję Cancel, przed zastosowaniem ponawia Preview i weryfikuje digest.
Zapis ustawień nie uruchamia Apply, a Apply nie wybiera runtime'u ani nie
zapisuje preferencji.

## Granice i status

Kontrakt opisuje routing i walidację wpisu. [Model stanu](state-model.md)
posiada globalny draft, fokus i przejścia; [mapa ekranów](screens.md) opisuje
ich pola; [kontrakt błędów](errors.md) zachowuje rozróżnienie diagnostyk TUI i
statusów silnika. Nieznana slash command nie jest `status.invalid` z engine.

Paleta jest nawigacyjna i rejestruje wyłącznie `/translate`, `/forms`,
`/concepts`, `/languages` i `/settings`. Ukośnik w edytorze ścieżki F5 jest
zwykłym znakiem. Dokument nie zmienia CLI, który nadal wymaga jawnego
podpolecenia i zachowuje swoje dotychczasowe JSON, stderr i kody wyjścia.

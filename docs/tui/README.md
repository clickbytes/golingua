# GoLingua TUI — kontrakt produktu

**Stan: G11-V1 dostarcza powłokę Bubble Tea v2, F1 Translate i wybór/sprawdzenie/zapis ścieżki runtime w F5; G11-V2 dodaje kandydatów, podgląd i paletę nawigacji.** F2–F4 są jawnymi placeholderami. Pozostałe opisy ekranów i makiety przedstawiają docelowy kontrakt. CLI i TUI używają tych samych usług Go.

## Uruchamianie

| Wywołanie | Zachowanie docelowe | Zachowanie G11-V1 |
| --- | --- | --- |
| `golingua` na interaktywnym TTY | Otwiera F1 Translate | Otwiera F1 Translate |
| `golingua` bez TTY | Wyświetla pomoc bez ANSI | Wyświetla pomoc CLI |
| `golingua --help`, `-h` lub `help` | Zawsze wyświetla pomoc CLI | Wyświetla pomoc CLI |
| `golingua --version` | Zwraca samą wersję | Zwraca samą wersję |
| dowolne istniejące polecenie (`translate`, `forms`, `catalog`, `snapshot`, `meaning`, `question`, `capabilities`) | Zachowuje składnię, JSON na stdout i kody wyjścia | Zachowuje obecną obsługę |

TUI działa tylko wtedy, gdy stdin i stdout są terminalami. TUI renderuje ANSI wyłącznie w terminalu; nie uruchamia się przy pracy przez potok ani przekierowanym stdout. W pierwszej wersji nie dodajemy osobnego podpolecenia `tui`. Jawne `--help` zawsze wygrywa z automatycznym startem TUI.

## Wspólne usługi, bez drugiego silnika

CLI i TUI mają wywoływać tę samą typowaną warstwę `internal/app`. Usługi używają istniejących pakietów `engine`, `languages/pl`, `languages/en`, `catalog`, `core` i `staging`. TUI nie tworzy drugiego parsera, realizatora, leksykonu ani zbioru reguł, nie uruchamia własnej kopii `golingua` jako procesu potomnego i nie interpretuje statusu silnika po swojemu.

TUI pokazuje tylko dane zwrócone przez wybraną usługę. Brak formy pozostaje brakiem formy; brak ekwiwalentu leksykalnego nie staje się wymyślonym tłumaczeniem. Podgląd kandydata nie jest akceptacją językową.

## Pięć ekranów

- **F1 Translate** — wybór PL↔EN, tekst, wynik, status, rzeczywiste kandydatury i ich założenia. Szczegóły w [screen map](screens.md).
- **F2 Forms** — analiza `surface` albo generacja z leksemu i jawnych cech. Nie uzupełnia nieznanych komórek.
- **F3 Concepts** — filtrowanie po stabilnym ConceptID, rzeczywiste pokrycie katalogu i opcjonalne odczytanie definicji/relacji z istniejącego ConceptGraph.
- **F4 Languages** — wyłącznie PL i EN, liczniki katalogu i osobny stan wybranego runtime'u.
- **F5 Settings** — poziom wyświetlania, domyślny kierunek, ścieżka runtime config i katalog snapshotu; podgląd i osobno potwierdzone zastosowanie snapshotu.

Nie dodajemy innych zakładek ani funkcji asystenta czatu. Granice danych i pakietów opisuje [mapa bindingów](bindings.md), statusy i błędy — [kontrakt błędów](errors.md), własność modelu, zdarzenia, odpowiedzi asynchroniczne i fokus — [kontrakt stanu](state-model.md), a routing promptu — [kontrakt komend](command-contract.md).

Makieta wspólnej ramy w 80×24 i szerokim terminalu, wraz ze stanem kompaktowym i zawijaniem długiej ścieżki, jest oznaczona jako projekt w [global-shell mockup](mockups/global-shell.md); nie pokazuje działającego TUI ani wyniku silnika.

Dolny, wspólny prompt ma osobną makietę [fixed bottom prompt layout](mockups/bottom-prompt.md) z wersją jednoliniową, ograniczonym szkicem wieloliniowym, historią i stanem kompozycji IME. To również projekt układu, nie zachowanie działającego TUI.

Paleta komend ma osobny projekt [slash palette layout](mockups/slash-palette.md), ograniczony do pięciu istniejących komend nawigacyjnych. Pokazuje dopasowania prefiksu, fokus, wybór, brak wyników i błędny wpis bez uruchamiania silnika.

Ekran F1 Translate ma własną [makietę](mockups/translate.md) kierunku, wyniku, kandydatów i statusów. Przykłady z danych obserwacyjnych są oznaczone jako historyczne; w mockupie nie ma wyniku wymyślonego dla niedostępnego runtime'u.

Lista i szczegóły kandydatów mają osobną [makietę nakładki](mockups/translate-candidates.md). Pokazuje rzeczywiste ID, stan kandydata, diagnostykę i jawny brak tekstu bez utożsamiania podglądu z akceptacją.

F2 Analyze ma osobną [makietę analizy formy powierzchniowej](mockups/forms-analyze.md). Pokazuje odczyty PL/EN zwracane przez `forms`, stan `unknown` i błędne żądanie; nie generuje form.

F2 Generate ma osobną [makietę generowania form](mockups/forms-generate.md). Pokazuje rzeczywiste formy PL/EN i pola zwrócone przez `FormsResult`; feature rows pozostają jawnym wejściem `key=value`, bez zgadywanych list wartości. `incomplete`, `blocked` i `unavailable` są opisane jako stany kontraktu API, których zwykły CLI nie odtwarza na aktualnym seedzie.

F3 Concepts list/search ma osobną [makietę listy konceptów](mockups/concepts-list.md). Pokazuje rzeczywiste ConceptID i statusy pokrycia PL/EN; filtr działa lokalnie po ID. Brak wyników nie oznacza pustego katalogu, a etykiety, definicje i tłumaczenia nie są wyprowadzane z ID.

## Ustawienia

Preferencje są wersjonowanym JSON w `os.UserConfigDir()/golingua/settings.json`. Zawierają tylko ustawienia interfejsu: poziom wyświetlania, domyślny kierunek PL↔EN, ścieżkę runtime config i katalog snapshotu. Nie przechowują tekstów zdań, leksykonu, znaczeń, historii, faktów, uczenia ani sekretów.

Zapis następuje po jawnym poleceniu **Save**: plik tymczasowy w tym samym katalogu, bezpieczne uprawnienia użytkownika i atomowa zamiana. Niepoprawny plik daje ostrzeżenie i wartości domyślne; aplikacja nie nadpisuje go po cichu. Brak możliwości odczytu/zapisu ustawień nie blokuje samego CLI. Zapis preferencji nigdy nie stosuje snapshotu. `Preview` i `Apply` snapshotu są osobnymi działaniami, a `Apply` zawsze wymaga potwierdzenia.

## Biblioteka TUI

Wybrano [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) z modułem `charm.land/bubbletea/v2`. Zweryfikowana wersja referencyjna G11-01 to [v2.0.10](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.10), licencja projektu upstream to [MIT](https://github.com/charmbracelet/bubbletea/blob/main/LICENSE). Biblioteka dostarcza cykl zdarzeń i renderowanie terminala; decyzje i reguły GoLingua pozostają w `internal/app`. Zależność v2.0.10 jest przypięta w `go.mod`.

## Dostępność i rozmiar

Minimalny układ docelowy to około 80×24 znaki. Przy mniejszym oknie ekran przechodzi do zwartego układu, a jeśli nie można zachować widocznego statusu i fokusu — pokazuje prośbę o powiększenie terminala. Zmiana rozmiaru nie odrzuca danych formularza. TUI obsługuje klawiaturę zgodnie z [mapą klawiszy](keymap.md); mysz nie jest wymagana.

## Etapy

G11-01–12 ukończyły kontrakty, wspólne usługi i makiety F1, F2 oraz listy F3. Działa zakres G11-V1; kolejne zakresy pozostają planem. [Epic G11](https://github.com/clickbytes/golingua/issues/254) zastępuje niewykonane mikroetapy G11-13–42 ośmioma kolejnymi zakresami, z których każdy ma dać używalny przepływ w terminalu:

| Zakres | Obserwowalny rezultat |
| --- | --- |
| G11-V1 | Rama pięciu zakładek z jednym dolnym promptem i działającym F1 Translate; CLI bez TTY zachowuje obecne działanie. |
| G11-V2 | Szczegóły rzeczywistych kandydatów, jawny podgląd po PreviewID i pięć komend nawigacyjnych z paletą. |
| G11-V3 | F2 Analyze i Generate z jawnymi, lokalnymi dla języka cechami `key=value`. |
| G11-V4 | F3 Concepts: lista, filtr po ConceptID i szczegóły rzeczywistego rekordu. |
| G11-V5 | F4 Languages: inwentarz PL/EN oddzielony od gotowości wybranego runtime'u. |
| G11-V6 | F5 Settings: pozostałe preferencje (ścieżka runtime’u działa od V1). |
| G11-V7 | F5 snapshot: podgląd i osobno potwierdzone zastosowanie. |
| G11-V8 | Dokumentacja CLI/TUI, przegląd integracyjny i zamknięcie epiku. |

Makiety i dokumentację aktualizujemy razem z odpowiadającym im działającym zakresem. Kolejne issue otwieramy po zamknięciu poprzedniego.

## Obsługa dostarczona w V1

F1: wpisz tekst w dolnym prompcie i naciśnij `Enter`. `Alt+D` albo `Enter` na kontrolce direction zmienia EN↔PL. Wynik pokazuje faktyczny status, tekst lub jego brak i diagnostyki. Wynik starszego szkicu jest oznaczony. Lista kandydatów i ich podgląd działają od V2.

F5: wpisz ścieżkę runtime config; `Enter` wybiera i sprawdza ją przez `Service.ValidateRuntime`. `Ctrl+S` albo `Enter` na kontrolce save jawnie zapisuje wersję 1 ustawień i bezwzględną ścieżkę. Tymczasowy plik ma uprawnienia 0600 i jest atomowo przemianowany w tym samym katalogu. Nieaktualny runtime zachowuje ścieżkę i diagnostyki. Niepoprawne ustawienia wywołują ostrzeżenie, bez zapisu do czasu jawnego Save. Pozostałe preferencje należą do V6; snapshot Preview/Apply nie są dostępne w V1.

`Tab`/`Shift+Tab` przechodzą po jawnie nazwanych kontrolkach, w tym tabs. Na tabs działają `Left`/`Right`; fokus result pozwala przewijać `Up`/`Down`/`PgUp`/`PgDown`. `Ctrl+U` czyści pole. `?` pokazuje pomoc, `Esc` zamyka ją i wraca do promptu, `Ctrl+C`/`Ctrl+D` bezpiecznie kończy program. Zmiana zakładki i rozmiaru zachowuje szkice.

## Obsługa dostarczona w V2

F1 pokazuje wszystkie rzeczywiście zwrócone kandydatury. `Tab` przechodzi kolejno przez prompt, direction, candidates, preview, result i tabs. Na candidates `Up`/`Down` wybierają do inspekcji; `Enter` otwiera szczegóły z pełnym ID, statusem, tekstem lub jego brakiem, założeniami i diagnostykami. `Esc` wraca do listy. Szczegóły i wynik przewija się strzałkami oraz `PgUp`/`PgDown`.

`Enter` na preview ponownie wywołuje `Service.Translate` z zapisanym żądaniem wyniku i jawnym `PreviewID`; szkic może być już inny. Odpowiedź podglądu jest osobna: pierwotny status i wynik pozostają zachowane. Kandydat bez tekstu nadal ma szczegóły, ale akcja preview wyświetla brak możliwości podglądu. Wybór ani podgląd nie oznaczają akceptacji językowej.

`Ctrl+P` z każdego ekranu otwiera wspólną paletę z początkowym `/`. Wpisz nazwę bez ponownego ukośnika, np. `forms`, albo wyczyść pole i wpisz `/forms`. Dostępne są wyłącznie `/translate`, `/forms`, `/concepts`, `/languages`, `/settings`. Prefiks filtruje listę; `Up`/`Down` wybierają sugestię, `Enter` wykonuje dokładną nazwę lub jawną selekcję. Nieznane nazwy i argumenty dają błąd bez wywołania usługi. `Esc` anuluje bez zmiany szkicu i fokusu ekranu. W F1 początkowy `/` otwiera paletę; w F5 pole ścieżki przyjmuje ukośniki jako część ścieżki, a nawigację otwiera `Ctrl+P`. Przełączanie ekranów i resize zachowują selekcję oraz szkic.

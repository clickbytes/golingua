# GoLingua TUI — kontrakt produktu

**Stan: zatwierdzony kontrakt G11-01; interfejs TUI nie jest jeszcze zaimplementowany.** Do implementacji G11-04 dołączone zostanie Bubble Tea v2, a pięć ekranów ma używać tych samych usług Go co CLI. Ta dokumentacja nie oznacza, że TUI jest już dostępne.

## Uruchamianie

| Wywołanie | Zachowanie docelowe | Zachowanie na `main` przed implementacją G11-04 |
| --- | --- | --- |
| `golingua` na interaktywnym TTY | Otwiera F1 Translate | Wyświetla pomoc CLI |
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

## Ustawienia

Preferencje są wersjonowanym JSON w `os.UserConfigDir()/golingua/settings.json`. Zawierają tylko ustawienia interfejsu: poziom wyświetlania, domyślny kierunek PL↔EN, ścieżkę runtime config i katalog snapshotu. Nie przechowują tekstów zdań, leksykonu, znaczeń, historii, faktów, uczenia ani sekretów.

Zapis następuje po jawnym poleceniu **Save**: plik tymczasowy w tym samym katalogu, bezpieczne uprawnienia użytkownika i atomowa zamiana. Niepoprawny plik daje ostrzeżenie i wartości domyślne; aplikacja nie nadpisuje go po cichu. Brak możliwości odczytu/zapisu ustawień nie blokuje samego CLI. Zapis preferencji nigdy nie stosuje snapshotu. `Preview` i `Apply` snapshotu są osobnymi działaniami, a `Apply` zawsze wymaga potwierdzenia.

## Biblioteka TUI

Wybrano [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) z modułem `charm.land/bubbletea/v2`. Zweryfikowana wersja referencyjna G11-01 to [v2.0.10](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.10), licencja projektu upstream to [MIT](https://github.com/charmbracelet/bubbletea/blob/main/LICENSE). Biblioteka dostarcza cykl zdarzeń i renderowanie terminala; decyzje i reguły GoLingua pozostają w `internal/app`. Zależność zostanie przypięta w `go.mod` w etapie implementacji.

## Dostępność i rozmiar

Minimalny układ docelowy to około 80×24 znaki. Przy mniejszym oknie ekran przechodzi do zwartego układu, a jeśli nie można zachować widocznego statusu i fokusu — pokazuje prośbę o powiększenie terminala. Zmiana rozmiaru nie odrzuca danych formularza. TUI obsługuje klawiaturę zgodnie z [mapą klawiszy](keymap.md); mysz nie jest wymagana.

## Etapy

G11-01 zamroził kontrakt produktu, a G11-02 udostępnił usługi in-process z [mapy bindingów](bindings.md). G11-03 definiuje MVU, fokus i odpowiedzi asynchroniczne w [kontrakcie stanu](state-model.md). G11-04 definiuje routing promptu i slash commands w [kontrakcie komend](command-contract.md). Późniejsze etapy dodadzą osobne makiety, shell, ekrany i ustawienia; implementacja TUI jeszcze nie istnieje, a podręcznik CLI jest osobnym zadaniem G11-39.

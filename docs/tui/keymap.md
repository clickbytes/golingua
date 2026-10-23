# Klawisze i zachowanie terminala

| Klawisz | Obecne działanie |
| --- | --- |
| `F1`–`F5` | Otwórz odpowiadający ekran i zachowaj jego stan. Wejście na F4 odświeża katalog oraz walidację runtime'u. |
| `Tab` / `Shift+Tab` | Przejdź po kolejności fokusu aktywnego ekranu, łącznie z paskiem zakładek. |
| `Left` / `Right` | Na pasku zakładek otwórz sąsiedni ekran. W dialogu snapshotu zmień zaznaczoną akcję. |
| `Up` / `Down` | Zmień wybór kandydatury/konceptu albo przewiń fokusowany wynik. W dialogu snapshotu przełącz Cancel/Apply. |
| `Enter` | Wykonaj akcję wskazaną fokusem. Snapshot Apply najpierw otwiera dialog z domyślnym Cancel. |
| `Esc` | Zamknij pomoc lub podwidok; na F5 resetuje szkice, chyba że trwa zapis albo snapshot Apply. W dialogu zamyka go bez Apply. |
| `PgUp` / `PgDown` | Przewiń fokusowane wyniki i szczegóły konceptu F3. |
| `Ctrl+U` | Wyczyść aktywny edytor. |
| `?` | Pokaż pomoc klawiatury. |
| `Ctrl+C` / `Ctrl+D` | Zakończ działanie i przywróć terminal; podczas snapshot Apply poczekaj na wynik operacji. |

Ctrl+P otwiera paletę nawigacji z dowolnego ekranu. Początkowy ukośnik w
promptcie F1–F3 także otwiera paletę. Rejestr zawiera dokładnie `/translate`,
`/forms`, `/concepts`, `/languages` i `/settings`; komendy wyłącznie
przełączają ekran. Ukośnik w ścieżce store F5 jest zwykłym znakiem.

Shell jest projektowany dla około 80×24. Resize przelicza zawijanie i widoczne
przewinięcie bez utraty szkiców, selekcji ani filtrów. Przy 40×12 lub więcej
renderuje ramę; poniżej 40×12 pokazuje informację o minimalnym rozmiarze. Fokus
jest nazwany, nie zależy wyłącznie od koloru. Mysz nie jest wymagana.

TUI startuje tylko przy interaktywnym stdin i stdout oraz `TERM` różnym od
`dumb`. Bez TTY CLI zachowuje ścieżkę nieinteraktywną i nie emituje ekranu ANSI
TUI. Bubble Tea przywraca terminal podczas kontrolowanego wyjścia.

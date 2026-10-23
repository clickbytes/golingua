# Klawisze i zachowanie terminala

| Klawisz | Działanie |
| --- | --- |
| `F1`–`F5` | Przejdź bezpośrednio do odpowiadającego ekranu |
| `Tab` / `Shift-Tab` | Przenieś fokus do następnego/poprzedniego kontrolera; ekranowy pasek zakładek też jest osiągalny z klawiatury |
| `←` / `→` | Na ekranowym pasku wybierz poprzednią/następną zakładkę; w innym kontekście zachowaj działanie pola |
| `↑` / `↓` | Przesuń zaznaczenie lub kursor w bieżącej liście |
| `Enter` | Uruchom zaznaczoną akcję lub zatwierdź wybrany element |
| `Esc` | Zamknij podwidok lub wróć o jeden poziom; nie zatwierdza zmian |
| `?` | Pokaż skróconą pomoc dla bieżącego ekranu |
| `Ctrl-C` | Bezpiecznie zakończ działanie i przywróć terminal |

Brak myszy nie ogranicza dostępu do żadnej akcji. Fokus jest zawsze widoczny. Bieżąca zakładka, aktywne pole i status wyniku nie mogą zależeć wyłącznie od koloru.

## Rozmiar i terminal

Docelowa minimalna przestrzeń robocza to około 80 kolumn × 24 wiersze. Po resize ekran przelicza układ bez wymazywania wprowadzonych danych. Gdy terminal jest mniejszy, interfejs przechodzi do układu zwartego; jeśli nadal nie może pokazać pola wejściowego, fokusu i statusu, pokazuje komunikat `Powiększ terminal do co najmniej 80×24`, a nie ucięty wynik.

TUI uruchamia się tylko przy interaktywnym stdin i stdout. Przy braku TTY albo nieobsługiwanym terminalu nie włącza raw mode i nie drukuje kodów ANSI do potoku. Wyjście przez `Ctrl-C`, EOF lub błąd inicjalizacji przywraca stan terminala.

Taby F1–F5 pozostają stałe. Gdy terminal przechwytuje F-keys, użytkownik może przejść `Tab`/`Shift-Tab` do widocznego paska zakładek, wybrać ekran klawiszami `←`/`→` i otworzyć go przez `Enter`.

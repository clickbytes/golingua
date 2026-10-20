# Ręczna weryfikacja językowa golingua

**Obowiązuje od 2026-10-20.** Na tym etapie **nie ma automatycznej macierzy translacji**, porównania z flexLingua, niezależnych goldów ani liczby przypadków wymaganej przed PR/merge. Użytkownik przekazuje słowa lub zdania do sprawdzenia, agent uruchamia nowy silnik Go i relacjonuje *zaobserwowany* wynik, bez zgadywania.

## Jeden przypadek — prosta procedura

1. Użytkownik podaje tekst i kierunek `pl→en` lub `en→pl`; może wskazać ręcznie poprawny wariant, lecz nie musi.
2. Po implementacji CLI uruchom **dokładnie ten tekst**; załącz rzeczywistą komendę, SHA oraz output/diagnostykę. Brak możliwości uruchomienia → **NOT RUN**, nie udawaj wyniku.
3. Użytkownik ocenia: `OK`, `BŁĄD`, `NIEPEWNE` lub `BRAK OBSŁUGI`. Krótkie uzasadnienie i oczekiwana poprawka wystarczą.
4. Jeśli jest błąd, popraw właściwy fragment leksyki/morfologii/konstrukcji/CORE, a nie zapisuj całej pary zdaniowej jako wyjątku.
5. Można ponownie wykonać **to samo** wejście dla manualnej kontroli poprawki. Nic nie trafia automatycznie do CI.

Nie ma obowiązku zapisywania każdego outputu do pliku. Jeśli zapis jest przydatny, prowadź opcjonalnie krótką notatkę w issue lub `examples/manual/`:

```text
SHA: <commit>
Kierunek: pl -> en
Wejście: <zdanie użytkownika>
Komenda: <dokładna komenda CLI>
Rzeczywiste wyjście: <tekst/JSON>
Ocena właściciela: OK | BŁĄD | NIEPEWNE | BRAK OBSŁUGI | NIE OCENIONO
Uwaga: <co poprawić>
```

Wyniki diagnostyczne muszą odróżniać: brak leksemu, brak sensu, nieobsługiwaną konstrukcję, wieloznaczność, limit i rzeczywisty błąd programu. **Zielone CI = da się zbudować i uruchomić CLI, nie jakość językowa.**

## Co NIE jest wymagane

- Przenoszenie dawnych zdań lub ich oczekiwanych rezultatów z JS.
- Tworzenie `tests/cases`, `fixtures`, formalnych goldów, oracle, raportów PASS% ani test runnera.
- Uruchamianie masowego automatycznego `go test ./...`.
- Testowanie wszystkich kombinacji konstrukcji, liczb, czasów w jednym PR.
- Przyjmowanie przez agenta roli arbitra zgodności językowej bez właściciela.

## Bezpieczeństwo i ograniczenia

Ograniczenie testów **nie** oznacza rezygnacji z walidacji danych wejściowych, ograniczania zasobów, jawnych wyjątków ani poprawnego rozdzielenia ról semantycznych. To własności runtime. Na tym etapie oceniamy ręcznie przede wszystkim realne zachowanie PL/EN. Gdy projekt dojrzeje, użytkownik może osobno zlecić powrót do testów automatycznych.

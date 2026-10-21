# Ręczny przegląd outputu golingua

**Obowiązuje od 2026-10-20.** Nie ma automatycznej macierzy translacji, porównania z flexLingua, niezależnych goldów ani liczby przypadków wymaganej przed PR/merge. Właściciel może przekazać dowolne słowo lub zdanie; agent uruchamia dostępne CLI i relacjonuje *zaobserwowany* wynik, bez zgadywania. Sama dostępność komendy nie oznacza, że skonfigurowano runtime językowy.

## Jeden przypadek — prosta procedura

1. Właściciel podaje dowolny tekst i kierunek (`pl→en` albo `en→pl`). Może wskazać akceptowalny wariant, lecz nie musi. Nie wymyślaj tekstu właściciela, jeśli go nie podał.
2. Zapisz pełny SHA, dokładną komendę, surowe stdout, stderr i exit code z uruchomienia. Nie poprawiaj ani nie parafrazuj outputu. Jeśli komenda nie została uruchomiona, wpisz **NOT RUN** i nie twórz przykładowego wyniku.
3. Właściciel ocenia output jako `OK`, `BŁĄD`, `NIEPEWNE` albo `BRAK OBSŁUGI`. Krótkie uzasadnienie wystarczy. Status techniczny uruchomienia (PASS/FAIL/INFRA/NOT RUN) zapisuje się oddzielnie od oceny językowej.
4. Diagnozuj błąd programu osobno od złej fleksji, sensu lub tłumaczenia. Błąd uruchomienia, parsowania argumentów, panic, serializacji lub niezgodny kontrakt to problem programu. Niepoprawna forma językowa przy działającym CLI to obserwacja językowa, nie awaria builda.
5. Jeśli potrzebna jest poprawka, zmień właściwą leksykę/morfologię/konstrukcję/CORE. Nie zapisuj całej pary zdaniowej jako wyjątku. To samo wejście można uruchomić ponownie ręcznie; żaden wynik nie trafia automatycznie do CI.

### Status technicznego podglądu v0

W technicznym podglądzie v0 `forms` i `catalog` są dostępne lokalnie. Właściciel
może podać słowo PL/EN do oceny odmiany; agent uruchomi dokładnie tę formę
przez `forms` i zapisze literalne wyjście. Ocena translacji zdania pozostaje
`NOT RUN` do czasu zatwierdzenia danych i uruchomienia providerów zdań.

CLI przyjmuje `translate --from <lang> --to <lang> --text <text> --json`, waliduje żądanie i zwraca typowy JSON. Domyślna konfiguracja nie zawiera jednak providerów językowych, więc poprawnie sformułowane żądanie zwraca `unsupported` i exit code 3. Taki wynik potwierdza tylko uczciwą odmowę; nie jest wynikiem translacji ani podstawą do oceny PL/EN. Błędne flagi lub para językowa zwracają `invalid` i exit code 2. `forms` udostępnia ograniczoną odmianę i analizę obecnych seedów PL/EN; ocena na wejściu właściciela pozostaje `NOT RUN`.

Komenda `question --runtime-config <path> --json` przyjmuje typowany MeaningGraph, jawny FactSet oraz mapowanie ról. Nie przyjmuje zdania w języku naturalnym ani nie realizuje odpowiedzi tekstowej. Jej wynik można ręcznie ocenić jako rozumowanie wyłącznie na wejściach dostarczonych przez właściciela. Bez takich wejść, aktywnego CORE snapshotu i uruchomienia konkretnej komendy wpisz **NOT RUN**; nie twórz przykładowych faktów ani twierdzeń.

Bez tekstu dostarczonego przez właściciela oznacz ocenę językową **NOT EVALUATED**. Jeśli nie uruchomiono konkretnej próby na wejściu właściciela, oznacz ją **NOT RUN**. Smoki techniczne z tekstem zastępczym nie są wejściem do oceny językowej. Dokładny kontrakt i przykład komendy opisuje [docs/cli.md](cli.md); nie przedstawiaj go jako dowodu, że provider jest skonfigurowany.

Nie ma obowiązku zapisywania każdego outputu do pliku. Jeśli zapis jest przydatny, prowadź opcjonalnie krótką notatkę w issue lub `examples/manual/`:

```text
Implementacja/wersja: <pełny SHA lub wersja>
Tryb: <lokalna komenda / rzeczywisty serwis>
Kierunek: <pl -> en / en -> pl>
Wejście właściciela: <dokładny tekst albo prywatny odnośnik>
Komenda: <dokładne polecenie, jeśli uruchomiono>
Wynik techniczny: PASS | FAIL | INFRA | NOT RUN
Exit code: <kod albo NOT RUN>
stdout: <literalny output albo NOT RUN>
stderr: <literalny output albo NOT RUN>
Ocena właściciela: OK | BŁĄD | NIEPEWNE | BRAK OBSŁUGI | NOT EVALUATED
Uwaga: <opcjonalna diagnoza lub poprawka>
```

Nie zapisuj prywatnego wejścia lub rozmowy do repo bez potrzeby i zgody właściciela. Notatki w [`examples/manual/`](../examples/manual/README.md) są opcjonalne i nie są czytane przez CI.

Wyniki diagnostyczne muszą odróżniać: brak leksemu, brak sensu, nieobsługiwaną konstrukcję, wieloznaczność, limit i rzeczywisty błąd programu. **Zielone CI = da się zbudować i uruchomić CLI, nie jakość językowa.**

## Co NIE jest wymagane

- Przenoszenie dawnych zdań lub ich oczekiwanych rezultatów z JS.
- Tworzenie `tests/cases`, `fixtures`, formalnych goldów, oracle, raportów PASS% ani test runnera.
- Uruchamianie masowego automatycznego `go test ./...`.
- Testowanie wszystkich kombinacji konstrukcji, liczb, czasów w jednym PR.
- Przyjmowanie przez agenta roli arbitra zgodności językowej bez właściciela.

## Bezpieczeństwo i ograniczenia

Ograniczenie testów **nie** oznacza rezygnacji z walidacji danych wejściowych, ograniczania zasobów, jawnych wyjątków ani poprawnego rozdzielenia ról semantycznych. To własności runtime. Na tym etapie oceniamy ręcznie przede wszystkim realne zachowanie PL/EN. Gdy projekt dojrzeje, użytkownik może osobno zlecić powrót do testów automatycznych.

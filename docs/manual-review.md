# Ręczny przegląd outputu golingua

**Obowiązuje od 2026-10-20, z delegacją właściciela z 2026-10-23.** Nie ma automatycznej macierzy translacji, porównania z flexLingua, niezależnych goldów ani liczby przypadków wymaganej przed PR/merge. Właściciel może przekazać słowa, zdania lub korpus; agent uruchamia dostępne CLI, sam ocenia konkretne obserwacje i poprawia błędy w granicach aktywnego zadania. Ocena agenta nie jest podpisem ani akceptacją właściciela. Nie wymyślaj niepodanych wyników oczekiwanych. Sama dostępność komendy nie oznacza, że skonfigurowano runtime językowy.

## Jeden przypadek — prosta procedura

1. Właściciel podaje dowolny tekst i kierunek (`pl→en` albo `en→pl`). Może wskazać akceptowalny wariant, lecz nie musi. Nie wymyślaj tekstu właściciela, jeśli go nie podał.
2. Zapisz pełny SHA, dokładną komendę, surowe stdout, stderr i exit code z uruchomienia. Nie poprawiaj ani nie parafrazuj outputu. Jeśli komenda nie została uruchomiona, wpisz **NOT RUN** i nie twórz przykładowego wyniku.
3. Właściciel ocenia output jako `OK`, `BŁĄD`, `NIEPEWNE` albo `BRAK OBSŁUGI`. Krótkie uzasadnienie wystarczy. Status techniczny uruchomienia (PASS/FAIL/INFRA/NOT RUN) zapisuje się oddzielnie od oceny językowej.
4. Diagnozuj błąd programu osobno od złej fleksji, sensu lub tłumaczenia. Błąd uruchomienia, parsowania argumentów, panic, serializacji lub niezgodny kontrakt to problem programu. Niepoprawna forma językowa przy działającym CLI to obserwacja językowa, nie awaria builda.
5. Jeśli potrzebna jest poprawka, zmień właściwą leksykę/morfologię/konstrukcję/CORE. Nie zapisuj całej pary zdaniowej jako wyjątku. To samo wejście można uruchomić ponownie ręcznie; żaden wynik nie trafia automatycznie do CI.

### Status technicznego podglądu v0

`forms` i `catalog` są dostępne lokalnie. CLI może też załadować ograniczony,
skompilowany runtime PL/EN dla dokładnych tras `seeing`, `hearing`, `looking` i `receiving`, gdy `translate`
otrzyma `--runtime-config` wskazujący aktywny snapshot. Pierwsze trzy trasy obsługują teraźniejszą
3. osobę liczby pojedynczej, a `receiving` — czas przeszły w 3. osobie liczby pojedynczej.
Bez tej konfiguracji
żądanie translacji zwraca `unsupported` i exit code 3.

Na polecenie właściciela uruchomiono translację 100 zapisanych par w obu
kierunkach; obserwowane statusy i teksty zapisano w
[`data/sentences/observed-2026-10-23-provider-runtime.tsv`](../data/sentences/observed-2026-10-23-provider-runtime.tsv).
Po aktywacji trasy `hearing` uruchomiono też 25 zapisanych par w obu kierunkach;
wyniki z kodami wyjścia, statusami, diagnostykami i tekstami zapisano w
[`data/sentences/observed-2026-10-23-g05-12-hearing.tsv`](../data/sentences/observed-2026-10-23-g05-12-hearing.tsv).
Po aktywacji `receiving` uruchomiono 25 istniejących par z tego samego korpusu
w obu kierunkach; wyniki zapisano w
[`data/sentences/observed-2026-10-23-g05-13-receiving.tsv`](../data/sentences/observed-2026-10-23-g05-13-receiving.tsv).
Po aktywacji `looking` uruchomiono istniejące pary S051–S075 w obu kierunkach;
wyniki zapisano w
[`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../data/sentences/observed-2026-10-23-g05-14-looking.tsv).
To są obserwacje działania programu, a nie ocena językowa. Właściciel nie ocenił
jeszcze konkretnych wyjść, więc ręczny odbiór translacji i ocena form pozostają
`NOT RUN`. `forms` udostępnia ograniczoną odmianę oraz analizę obecnych seedów
PL/EN.

Komenda `question --runtime-config <path> --json` przyjmuje typowany MeaningGraph, jawny FactSet oraz mapowanie ról. Nie przyjmuje zdania w języku naturalnym. `--language pl|en` opcjonalnie uruchamia realizację odpowiedzi. Z delegacji właściciela agent może sprawdzić ograniczony przykład oparty na dostępnych danych projektu oraz jawnie oznaczonym fakcie demonstracyjnym; taki fakt jest agent-authored, nietrwały i nie jest wejściem ani akceptacją właściciela. Bez uruchomienia komendy wpisz **NOT RUN**. Właścicielska akceptacja językowa pozostaje osobnym statusem.

W bieżącym runtime dodatnia odpowiedź yes/no może zwrócić tekst PL/EN ze statusem `incomplete` i jawnymi założeniami. Odpowiedź przecząca oraz tekst WH pozostają `unsupported` do czasu dodania zatwierdzonych konstrukcji negacji i samodzielnej frazy NP. Strukturalna odpowiedź reasonera, proof i źródło faktu pozostają dostępne w wyniku.

Bez tekstu dostarczonego przez właściciela oznacz ocenę językową **NOT EVALUATED**. Jeśli nie uruchomiono konkretnej próby na wejściu właściciela, oznacz ją **NOT RUN**. Gdy właściciel deleguje autonomiczną ocenę, agent może opisać własną ocenę na tych wejściach, ale status akceptacji właściciela pozostaje osobny. Smoki techniczne z tekstem zastępczym nie są wejściem do oceny językowej. Dokładny kontrakt i przykład komendy opisuje [docs/cli.md](cli.md); nie przedstawiaj go jako dowodu, że provider jest skonfigurowany.

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

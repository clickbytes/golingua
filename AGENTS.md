# golingua — jednoosobowy tryb pracy

## Nadrzędna decyzja właściciela (2026-10-20)

**Nowy Go od zera. Bez obowiązku odtwarzania starego JavaScriptu. Bez porównań wyjść do flexLingua. Bez rozbudowanych testów automatycznych.** Przez bieżące etapy jedyną oceną jakości językowej jest **ręczne sprawdzenie wyników tłumaczeń i form dostarczonych przez właściciela projektu**. Nie generuj sam masowych zestawów zdaniowych, nie dodawaj test runnerów, goldów, fixture-gates, automatycznej akceptacji expected ani metryk procentowych poprawności na tej podstawie.

Źródło `clickbytes/flexLingua` można czytać dla inspiracji lub metadanych i licencji, nie kopiować kodu ani zmuszać Go do powtarzania zachowania JS.

## Bramka szeregowa

- Dokładnie jeden wykonujący agent. Najwyżej jeden ACTIVE EPIC, jedno ACTIVE child issue, jedna gałąź i jeden implementacyjny PR. Wszystkie przyszłe issues QUEUED.
- Zanim edytujesz kod: sprawdź main SHA, otwarte PR i status epiców. W komentarzu issue podaj owner, pliki/symbole, plan, oczekiwany rezultat obserwowalny w CLI.
- Twórz jeden PR na issue. Przeglądaj diff, rzeczywisty CI i ręcznie obserwowane wyniki. Po merge sprawdź main i dopiero rozpocznij następny task.
- Jawny wyjątek właściciela dla **G07 #60 (2026-10-21)**: sekwencyjne commity bezpośrednio do `main` bez PR, każdy po przeglądzie różnic i z realnym CI; bez równoległych agentów. Pozostałe issues używają standardowego branch + PR. Wyjątek nie oznacza zatwierdzenia znaczeń.
- Connector GitHub nie jest terminalem użytkownika. Nie podawaj fikcyjnego lokalnego PASS. Zapisz NOT RUN, jeśli nie uruchomiono CLI na rzeczywistym środowisku.

## Narzędzia pracy Codex

- Jeśli RTK jest zainstalowany, preferuj `rtk <polecenie>` dla terminala i `gh` (np. `rtk git status`, `rtk gh issue list`). Bez RTK używaj zwykłego polecenia; RTK jest opcjonalnym narzędziem agenta, nie zależnością projektu.
- Używaj CodeGraph do wyszukiwania symboli, zebrania lokalnego kontekstu i sprawdzenia wpływu zmiany, jeśli indeks jest dostępny. Wyniki indeksu mogą być nieaktualne lub niepełne; sprawdzaj ważne ustalenia w plikach źródłowych. CodeGraph nie jest autorytetem dla semantyki produktu ani danych językowych.
- Brak RTK, CodeGraph, ich konfiguracji lub indeksu nie może blokować builda ani CI.

## Minimalny CI, niemal zero testów

- Wymagamy `gofmt`, `go vet ./...`, `go build ./cmd/golingua` i lekkiego smoke `--help`/`--version`.
- **Nie ma obligatoryjnego `go test ./...`, automatycznych testów gramatycznych, fixtures, formalnych goldów, matrix runnera, porównywania z JS, fuzzy benchmark gate ani liczby zdań do zdania.**
- Nie dodawaj plików `*_test.go` w bieżącej fazie, chyba że właściciel wyraźnie zmieni tę politykę.
- Po implementacji funkcji językowej zgłoś właścicielowi gotowe komendy CLI i wynik uruchomienia; właściciel poda zdania PL/EN i oceni output. Jeśli nie ma dostępnego lokalnego runtime, dokumentuj, że ocena ręczna pozostaje otwarta, nie zgaduj.
- Nie zmieniaj implementacji wyłącznie po to, by dopasować output do starej wersji JS.

## Architektura

Czytaj [`docs/architecture/README.md`](docs/architecture/README.md), właściwy kontrakt i aktywne issue. Concept ≠ leksem ≠ sens ≠ walencja ≠ konstrukcja ≠ procedura; MeaningGraph jest znaczeniem wypowiedzi, a nie przechowywaną ontologią. Gramatyka i profile cech są lokalne dla języka.

Wspólne typy cech → `internal/features`; CORE nie importuje pluginów PL/EN; parser i realizer używają skompilowanych planów konstrukcji; morfologia korzysta z paradygmatów oraz jawnych wyjątków. Do sprawdzenia po implementacji należą realne przypadki negacji, zamiany ról, polisemii i odmiany, **ręcznie**, na przykładach właściciela.

Nawigacja `ROCK.VERB.MORPH.PAST` nie jest wymaganiem. Trzymamy stabilne ConceptID/LexemeID/ConstructionID i opcjonalne gęste ID runtime.

## Rozwój danych

1000 konceptów PL i EN jest celem docelowym, nie wymogiem, aby w każdym PR generować ogromny słownik. Dodawaj tylko dane potrzebne do bieżącej fazy. Provenance i licencje danych zewnętrznych zachowuj; automatyczne importy nie mogą tworzyć niezatwierdzonych faktów ani uruchamiać kodu.

Brak pamięci i samouczenia w Go v0. Bez własnego serwera, FST, grafowej bazy i pluginów `.so`, dopóki rzeczywiste potrzeby ich nie uzasadnią.

## Zmiana zakresu

Jeśli stare issue nadal mówi o goldach, niezależnym oracle, matrycy testowej, test_*.go czy regresji względem flexLingua, jest to **stare wymaganie zastąpione decyzją 2026-10-20**. Takie issue musi zostać poprawione przed implementacją, nie wykonywane literalnie. Nie wolno rozszerzać zakresu bez zgody właściciela.

@RTK.md

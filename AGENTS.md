# golingua — jeden koordynator, praca szeregowa

## Nadrzędna decyzja właściciela (2026-10-20)

**Nowy Go od zera. Bez obowiązku odtwarzania starego JavaScriptu. Bez porównań wyjść do flexLingua. Bez rozbudowanych testów automatycznych.** Jakość sprawdzamy na rzeczywistych wejściach PL/EN dostarczonych przez właściciela albo zapisanych w jego korpusie. Właściciel 2026-10-23 delegował agentowi autonomiczną implementację, ocenę i poprawianie tych wyników; nie czekaj na jego odpowiedź, aby kontynuować. Ocenę agenta i akceptację właściciela zapisuj osobno; nie twierdź, że właściciel zatwierdził wynik, jeśli tego nie zrobił. Nie twórz dodatkowych masowych zestawów bez polecenia; nie dodawaj test runnerów, goldów, fixture-gates, automatycznej akceptacji expected ani metryk procentowych poprawności.

Źródło `clickbytes/flexLingua` można czytać dla inspiracji lub metadanych i licencji, nie kopiować kodu ani zmuszać Go do powtarzania zachowania JS.

## Bramka szeregowa

- Jeden koordynator odpowiada za wynik. Najwyżej jeden ACTIVE EPIC, jedno ACTIVE child issue, jedna gałąź i jeden implementacyjny PR. W danej chwili tylko jeden agent edytuje aktywny checkout; pozostali mogą wykonywać ograniczony przegląd bez zapisu. Wszystkie przyszłe issues QUEUED.
- Zanim edytujesz kod: sprawdź main SHA, otwarte PR i status epiców. W komentarzu issue podaj owner, pliki/symbole, plan, oczekiwany rezultat obserwowalny w CLI.
- Domyślnie twórz jeden PR na issue. Przeglądaj diff, rzeczywisty CI i ręcznie obserwowane wyniki. Po merge sprawdź main i dopiero rozpocznij następny task. Jawne polecenie właściciela wykonania konkretnej zmiany bezpośrednio na `main` jest wyjątkiem jednorazowym: sprawdź świeży HEAD, zakres diffu i rzeczywisty CI; nie przenoś wyjątku na kolejne zadania.
- Jawny wyjątek właściciela dla **G07 #60 (2026-10-21)**: sekwencyjne commity bezpośrednio do `main` bez PR, każdy po przeglądzie różnic i z realnym CI; bez równoległych agentów. Pozostałe issues używają standardowego branch + PR. Wyjątek nie oznacza zatwierdzenia znaczeń.
- Connector GitHub nie jest terminalem użytkownika. Nie podawaj fikcyjnego lokalnego PASS. Zapisz NOT RUN, jeśli nie uruchomiono CLI na rzeczywistym środowisku.

## Ograniczona delegacja — decyzja właściciela (2026-10-23)

- Koordynator **rozważa** delegację na początku issue, ale nie uruchamia pomocnika automatycznie. Deleguj tylko zadanie z krótkim zakresem, wskazanymi wejściami i plikami, sprawdzalnym wynikiem oraz spodziewaną oszczędnością czasu większą od kosztu koordynacji. Małe lub silnie powiązane zmiany wykonaj bez delegowania.
- Najwyżej dwóch pomocników może równolegle czytać źródła, przeglądać diff albo analizować obserwacje. Nie edytują repozytorium, nie piszą do GitHub i nie uznają danych językowych za zatwierdzone.
- Jeśli pomocnik ma edytować kod lub dokument, pracuje **jeden** taki wykonawca naraz, na jawnie przydzielonych plikach aktywnego issue. Koordynator nie edytuje tych plików równocześnie. Pomocnik przekazuje diff i rzeczywiste obserwacje; koordynator sprawdza je, integruje, uruchamia wymagane kontrole i sam wykonuje commit, PR oraz merge.
- Pomocnik nie otwiera własnej gałęzi, issue ani PR, nie zmienia zakresu i nie wysyła wiadomości do osób trzecich. Jedna bramka szeregowa obowiązuje także podczas delegacji. Odnotuj delegację w issue tylko wtedy, gdy ma znaczenie dla wyniku lub pochodzenia oceny.
- OpenCode lub inny agent jest opcjonalnym narzędziem koordynatora, nie zależnością projektu. Sprawdź dostępność narzędzia przed użyciem; brak instalacji nie zatrzymuje pracy. Nie wymagaj delegowania dla każdego issue.

## Narzędzia pracy Codex

- Jeśli RTK jest zainstalowany, preferuj `rtk <polecenie>` dla terminala i `gh` (np. `rtk git status`, `rtk gh issue list`). Bez RTK używaj zwykłego polecenia; RTK jest opcjonalnym narzędziem agenta, nie zależnością projektu.
- Używaj CodeGraph do wyszukiwania symboli, zebrania lokalnego kontekstu i sprawdzenia wpływu zmiany, jeśli indeks jest dostępny. Wyniki indeksu mogą być nieaktualne lub niepełne; sprawdzaj ważne ustalenia w plikach źródłowych. CodeGraph nie jest autorytetem dla semantyki produktu ani danych językowych.
- Brak RTK, CodeGraph, ich konfiguracji lub indeksu nie może blokować builda ani CI.

## Minimalny CI, niemal zero testów

- CI wymaga `gofmt`, `go vet ./...`, `go build ./cmd/golingua` i lekkiego smoke `--help`/`--version`. Po zmianie Go sprawdź lokalnie właściwe komendy przed PR. Przy zmianie samych dokumentów sprawdź diff i linki oraz rzeczywisty CI PR i `main`; nie powtarzaj lokalnego builda ani prób CLI bez konkretnej potrzeby.
- **Nie ma obligatoryjnego `go test ./...`, automatycznych testów gramatycznych, fixtures, formalnych goldów, matrix runnera, porównywania z JS, fuzzy benchmark gate ani liczby zdań do zdania.**
- Nie dodawaj plików `*_test.go` w bieżącej fazie, chyba że właściciel wyraźnie zmieni tę politykę.
- Po implementacji funkcji językowej uruchom lokalne CLI na dostępnych wejściach właściciela/korpusu, zapisz rzeczywiste wyniki i samodzielnie popraw konkretne błędy w granicach issue. Jeśli nie ma runtime, oznacz próbę `NOT RUN`. Akceptacja właściciela pozostaje osobnym statusem i nie jest warunkiem czekania na kolejne zadanie.
- Nie zmieniaj implementacji wyłącznie po to, by dopasować output do starej wersji JS.

## Konkluzywne zadania

- Po G11-R1 kolejne implementacyjne child issue mają kończyć się działającym przepływem użytkownika w terminalu, z rzeczywiście obserwowanym wynikiem na dostępnym wejściu. Jeśli obowiązkowego przepływu nie sprawdzono (`NOT RUN`), wynik zadania to `PARTIAL`/`BLOCKED`, nie `DONE`, chyba że właściciel wyraźnie zwolni z tego kryterium. Odrębna akceptacja językowa właściciela może pozostać `NOT EVALUATED`. Makiety i komponenty są częścią takiego zadania, nie osobnymi PR-ami tylko dla liczby zamkniętych issues.
- Jeden pionowy zakres może mieć kilka sekwencyjnych commitów w jednym PR. Dziel go dalej tylko wtedy, gdy samodzielny wynik będzie używalny i możliwy do przeglądu. Po merge domknij tytuł/body child issue oraz wskaźnik aktywnego zadania w epiku przed otwarciem następnego.
- `ok`, `incomplete` i `ambiguous` są statusami silnika, nie automatyczną oceną jakości. Dla zdań bez rozstrzygniętej określoności zachowuj jawne kandydatury; nie wymuszaj pojedynczego tłumaczenia przez zgadywanie.

## Architektura

Czytaj [`docs/architecture/README.md`](docs/architecture/README.md), właściwy kontrakt i aktywne issue. Concept ≠ leksem ≠ sens ≠ walencja ≠ konstrukcja ≠ procedura; MeaningGraph jest znaczeniem wypowiedzi, a nie przechowywaną ontologią. Gramatyka i profile cech są lokalne dla języka.

Wspólne typy cech → `internal/features`; CORE nie importuje pluginów PL/EN; parser i realizer używają skompilowanych planów konstrukcji; morfologia korzysta z paradygmatów oraz jawnych wyjątków. Do sprawdzenia po implementacji należą realne przypadki negacji, zamiany ról, polisemii i odmiany, **ręcznie**, na przykładach właściciela.

Nawigacja `ROCK.VERB.MORPH.PAST` nie jest wymaganiem. Trzymamy stabilne ConceptID/LexemeID/ConstructionID i opcjonalne gęste ID runtime.

## Rozwój danych

1000 konceptów PL i EN jest celem docelowym, nie wymogiem, aby w każdym PR generować ogromny słownik. Dodawaj tylko dane potrzebne do bieżącej fazy. Provenance i licencje danych zewnętrznych zachowuj; automatyczne importy nie mogą tworzyć niezatwierdzonych faktów ani uruchamiać kodu.

Brak pamięci i samouczenia w Go v0. Bez własnego serwera, FST, grafowej bazy i pluginów `.so`, dopóki rzeczywiste potrzeby ich nie uzasadnią.

## Zmiana zakresu

Jeśli stare issue nadal mówi o goldach, niezależnym oracle, matrycy testowej, test_*.go czy regresji względem flexLingua, jest to **stare wymaganie zastąpione decyzją 2026-10-20**. Takie issue musi zostać poprawione przed implementacją, nie wykonywane literalnie. Niewykonane mikroetapy G11-13–G11-42 zastępuje decyzja właściciela z 2026-10-23 o pionowych zakresach G11-V1–V8. Inne rozszerzenia zakresu wymagają zgody właściciela.

@RTK.md

## Koordynacja wykonania — Codex Sol Medium (uzupełnienie z 2026-10-23)

**Jedyny koordynator:** bieżąca główna sesja Codex z modelem Sol Medium. Koordynator posiada decyzje o zakresie, kolejności, delegacji, integracji, review, issue/epic, branch/PR, weryfikacji CI i merge. Pomocnicy nie przejmują roli koordynatora. Ustawienia roli opisano w `docs/agent-execution.md`.

- **Priorytet wymagań:** najnowsza decyzja właściciela i twarde zasady v0 / AGENTS.md -> aktywne child issue -> wyraźnie powiązane, aktualne kontrakty -> epic (kolejność) -> dokumenty historyczne. Implementacja jest źródłem faktów o *tym, co istnieje*, a issue o *tym, co trzeba zrobić*. Konflikt między issue i twardą decyzją zgłoś; nie rozstrzygaj go przez cichą zmianę zakresu.
- **Bez v2 i migracji:** nie projektuj drugiej wersji produktu, migracji danych, adapterów zgodności, abstrakcji na przyszłość, redesignów ani hipotetycznych rozszerzeń bez wyraźnego polecenia właściciela. `Bubble Tea v2` to zatwierdzona wersja zewnętrznej biblioteki dla G11, a nie pozwolenie na GoLingua v2.
- **Szczegółowe issue wyznacza zadanie:** nie implementuj zadań z queued issue, sugestii z docs, FIXME/TODO ani życzeń wykrytych przy przeglądzie. Tylko kryteria aktywnego child issue mogą rozszerzyć jego zakres. W razie niejasności zapisz ryzyko/bloker. Nie twórz dodatkowego backlogu automatycznie.
- **Docs / GitHub:** wyłącznie Sol aktualizuje GitHub issues, epics i PR-y. Dokumenty w repo zmieniaj tylko wtedy, gdy aktywne issue lub decyzja właściciela wyraźnie tego wymaga; samo przeczytanie nie uruchamia aktualizacji docs. Delegowani workerzy nigdy nie piszą do GitHub.
- **Testy:** nie dodawaj nowego `*_test.go`, fixture, golden file, test runnera, benchmarku ani testowej bramki, chyba że aktywne issue *wyraźnie* wymaga konkretnych testów i nie narusza decyzji właściciela. Wywołanie istniejących komend weryfikacji (`gofmt`, `go vet`, `go build`, smoke oraz adekwatne ręczne CLI) nie jest tworzeniem testów.
- **Budżet pracy:** jeden worker implementacyjny naraz; najwyżej dwóch łącznie oddelegowanych pomocników (odczyt / review). Standardowo maksymalnie jeden przebieg implementacji i jedna korekta po review. Po drugim nieudanym przebiegu Sol robi tylko jasno uzasadnioną, ograniczoną poprawkę albo zwraca `BLOCKED` / `PARTIAL` z dowodami. Bez samoczynnego resetowania zadania i ponawiania pętli.
- **Deleguj tylko zamknięte zadania:** precyzyjne wejście, base SHA, dopuszczone pliki/symbole, jednoznaczny wynik, konkretne zakazy, budżet i format odpowiedzi. Codex Luna Low jest preferowany do researchu i małych, lokalnych prac. OpenCode `--auto` wyłącznie dla jednoznacznych, mniej ryzykownych prac w odizolowanym środowisku roboczym; brak OpenCode nie blokuje rozwoju.
- **Sposób zakończenia:** `DONE` (w pełni wykazane), `PARTIAL` (konkretne niedokończone kryteria) albo `BLOCKED` (dokładny powód i próby). `NOT RUN` nigdy nie jest `PASS`. Sol sprawdza zakres diffu, wymagane kontrole i realne obserwacje; ocena agenta nie stanowi zgody właściciela.
- **Jeden PR na pionowy zakres:** jeden aktywny child, jedna implementacyjna gałąź/PR; nie rozbijaj funkcji na mikro-issues/PR-y bez samodzielnego używalnego rezultatu. Po PR i CI aktualizuj status child + epic tylko raz, nie generuj rozbudowanego logu procesu w issue.

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

## Delegacja małych zadań — najnowsza decyzja właściciela (2026-10-23)

- **Zawsze wydzielaj małe, zamknięte zadania i deleguj ich wykonanie do OpenCode.** Każde zlecenie ma dokładny JOB BRIEF: issue, base SHA, wejścia, dozwolone pliki/symbole, obserwowalny wynik, zakazy, limit pracy i format przekazania. GPT Luna Low może pomagać w analizie, review i małych lokalnych zadaniach. Koordynator odpowiada za podział pracy, redakcję, integrację, ocenę i weryfikację; nie zastępuje delegacji samodzielnym implementowaniem całej funkcji.
- Najwyżej **czterech pomocników łącznie** może pracować równocześnie, z uwzględnieniem rzeczywistych limitów narzędzi. Najwyżej jeden jest wykonawcą edytującym kod; pozostali wykonują ograniczoną analizę lub review bez zapisu. Nie piszą do GitHub i nie uznają danych językowych za zatwierdzone. Nie uruchamiaj czterech tylko po to, aby wykorzystać limit.
- Jeśli pomocnik ma edytować kod lub dokument, pracuje **jeden** taki wykonawca naraz, na jawnie przydzielonych plikach aktywnego issue. Koordynator nie edytuje tych plików równocześnie. Pomocnik przekazuje diff i rzeczywiste obserwacje; koordynator sprawdza je, integruje, uruchamia wymagane kontrole i sam wykonuje commit, PR oraz merge.
- Pomocnik nie otwiera własnej gałęzi, issue ani PR, nie zmienia zakresu i nie wysyła wiadomości do osób trzecich. Jedna bramka szeregowa obowiązuje także podczas delegacji. Odnotuj delegację w issue tylko wtedy, gdy ma znaczenie dla wyniku lub pochodzenia oceny.
- OpenCode jest narzędziem wykonawczym koordynatora, nie zależnością produktu ani CI. Sprawdź jego dostępność i konfigurację przed użyciem. Jeśli jest niedostępny, zleć dopuszczony mały zakres GPT Luna Low; jeśli plan właściciela wymaga konkretnie OpenCode albo nie można zapewnić wymaganej izolacji/uprawnień, zgłoś dokładny bloker zamiast obchodzić bramkę.

## OpenCode — jawna konfiguracja i aktualny wybór modelu (2026-10-23)

- **Przed każdą delegacją do OpenCode utwórz lub świadomie zaktualizuj jawny `opencode.json` w katalogu głównym jego izolowanej, jednorazowej kopii kodu.** Sam plik `.opencode/agents/*.md` ani historyczny przykład uruchomienia nie wystarczają. Konfiguracja ma wskazywać wybranego workera, dokładny model, limit 12 kroków i minimalne uprawnienia do przydzielonych plików. Nie umieszczaj w niej sekretów ani poświadczeń; nie dodawaj jej automatycznie do repozytorium produktu.
- **Pełny przegląd OpenCode wykonuj przy pierwszym użyciu danego dnia, nie przed każdym zleceniem.** Sprawdź rzeczywiście zainstalowaną wersję OpenCode i pomoc używanych poleceń. Odczytaj aktualną dokumentację i schemat konfiguracji, zamiast zakładać zgodność ze starą składnią. Dla V2 używaj `permissions`, `shell` i `subagent`; nie traktuj nieobsługiwanego pola jako działającego zabezpieczenia. Źródła: [konfiguracja](https://dev.opencode.ai/v2/docs/config/), [uprawnienia](https://dev.opencode.ai/v2/docs/permissions/), [agenci](https://dev.opencode.ai/v2/docs/agents/).
- **Sprawdź w tym samym kontenerze, katalogu roboczym i środowisku uruchomienia, co worker, jakie źródła konfiguracji i agenci są faktycznie wczytani.** Użyj dostępnych w tej wersji poleceń diagnostycznych (np. `debug config`, `debug agents`) i sprawdź wynik, nie tylko kod wyjścia. Brak wskazanego workera, niewczytany `opencode.json`, nieznane pola lub nieoczekiwane nadpisanie reguł wymagają wyjaśnienia przed edycją kodu. Nie wypisuj sekretów z konfiguracji.
- **W ramach dziennego przeglądu odśwież listę modeli dostępnych w tym środowisku** (np. `opencode models`, po sprawdzeniu bieżącej pomocy), zweryfikuj dokładne identyfikatory i aktualny koszt w katalogu dostawcy. Wykonaj krótką próbę odpowiedzi wybranego modelu z jawną konfiguracją. Historyczny identyfikator, wcześniejsza udana próba ani obecność modelu w katalogu nie dowodzą jego bieżącej dostępności. Korzystaj wyłącznie z modeli aktualnie bezpłatnych; nie przechodź na płatny model, plan ani płatny fallback. Jeśli właściciel jawnie przypiął konkretny model, jego niedostępność zgłoś zamiast po cichu wybierać inny.
- **W ramach dziennego przeglądu sprawdź aktualne rankingi i wyniki dotyczące planowanego rodzaju pracy**, w szczególności edycji kodu, pracy narzędziowej i wykonywania instrukcji. Porównuj tylko modele dostępne i dopuszczone kosztowo. Zapisz datę sprawdzenia, linki do źródeł, identyfikator wybranego modelu i krótkie uzasadnienie. Oddziel wyniki autorów benchmarku od deklaracji dostawcy; ogólny ranking czatu nie zastępuje oceny pracy nad kodem. Nie utożsamiaj podobnych nazw, wariantów ani darmowego endpointu z modelem z rankingu bez potwierdzenia. Brak wiarygodnego, aktualnego wyniku oznacz jako brak danych; nie wymyślaj pozycji ani oceny jakości. Ranking pomaga w wyborze, ale nie zastępuje przeglądu diffu i rzeczywistych kontroli.
- Wynik pełnego przeglądu można ponownie wykorzystać tego samego dnia dla tego samego środowiska i rodzaju pracy. Zapisz datę, środowisko, wersję/obraz, model, koszt, źródła i wyniki prób w lokalnej notatce koordynatora poza danymi produktu; brak dostępnej notatki oznacza potrzebę przeglądu. **Odśwież odpowiednią część przeglądu natychmiast**, gdy model przestaje być dostępny, zmienia się jego cena lub identyfikator, wersja OpenCode, obraz, dostawca, konfiguracja albo wymagany rodzaj pracy. Nie ponawiaj bez zmian niedostępnego modelu. Nie twórz zadania cyklicznego ani automatyzacji: chodzi o kontrolę podczas pracy. Przed każdym zleceniem nadal przygotuj jawny `opencode.json` i sprawdź jego faktyczne wczytanie oraz wymagane odmowy uprawnień; dzienna notatka nie zastępuje tych kontroli.
- Przed `--auto` wykonaj w izolacji rzeczywiste próby odmowy dla powłoki, narzędzi sieciowych, podagentów i dostępu poza katalogiem roboczym. Udana odpowiedź modelu lub deklaracja `deny` w pliku nie wystarcza. Zapisz obserwowane wyniki, wersję, model i źródła konfiguracji bez sekretów. Jeśli zabroniona operacja się wykona albo nie można sprawdzić odmów, nie przekazuj implementacji temu workerowi. Gdy plan właściciela wymaga takiej delegacji, zgłoś `BLOCKED`; nie obchodź bramki przez YOLO na hoście. Kontener bez `.git` i poświadczeń, zewnętrzny timeout oraz przegląd wszystkich zmienionych ścieżek pozostają obowiązkowe.

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
- **Budżet pracy:** jeden worker implementacyjny naraz; najwyżej czterech łącznie oddelegowanych pomocników (wykonawca oraz odczyt / review). Standardowo maksymalnie jeden przebieg implementacji i jedna korekta po review. Po drugim nieudanym przebiegu Sol integruje/redaguje tylko jasno uzasadnioną, ograniczoną poprawkę przekazaną przez wykonawcę albo zwraca `BLOCKED` / `PARTIAL` z dowodami. Bez samoczynnego resetowania zadania i ponawiania pętli.
- **Deleguj tylko zamknięte zadania:** precyzyjne wejście, base SHA, dopuszczone pliki/symbole, jednoznaczny wynik, konkretne zakazy, budżet i format odpowiedzi. Codex Luna Low jest preferowany do researchu i małych, lokalnych prac. OpenCode `--auto` wyłącznie dla jednoznacznych, mniej ryzykownych prac w odizolowanym środowisku roboczym; brak OpenCode nie blokuje rozwoju.
- **Sposób zakończenia:** `DONE` (w pełni wykazane), `PARTIAL` (konkretne niedokończone kryteria) albo `BLOCKED` (dokładny powód i próby). `NOT RUN` nigdy nie jest `PASS`. Sol sprawdza zakres diffu, wymagane kontrole i realne obserwacje; ocena agenta nie stanowi zgody właściciela.
- **Jeden PR na pionowy zakres:** jeden aktywny child, jedna implementacyjna gałąź/PR; nie rozbijaj funkcji na mikro-issues/PR-y bez samodzielnego używalnego rezultatu. Po PR i CI aktualizuj status child + epic tylko raz, nie generuj rozbudowanego logu procesu w issue.

## Domknięcie G11 — decyzja właściciela 2026-10-23

V4 → V5 → V6 → V7 → V8 wykonujemy szeregowo, po jednym child issue i PR; kolejny zakres dopiero po merge, zielonym CI main i zamknięciu poprzednika. Limit czterech pomocników uwzględnia rzeczywiste limity narzędzi; jeden wykonawca pisze naraz. Role opcjonalne: OpenCode wykonawca, Luna Low analiza, Luna Low review, OpenCode przegląd dowodów/propozycja dokumentacji bez zapisu. Nie uruchamiaj wszystkich dla wykorzystania limitu. Aktualnie dostępny darmowy model OpenCode wybieramy po wymaganym przeglądzie; brak OpenCode lub niespełniona bramka delegacji dopuszcza małe zadania Luna Low, bez obchodzenia izolacji.

## Macierz poziomów językowych — nowsza decyzja właściciela 2026-10-23

Właściciel jawnie zmienił wcześniejszy zakaz macierzy: zaczynamy od małej próbki A1 dla PL↔EN w `tests/A1/*.testdomain.go`. Kolejne `tests/A1.1/`, `tests/A1.2/` to następne partie tego samego celu A1, otwierane dopiero po określeniu zakresu. Pierwsza dostawa #309 jest deklaratywną macierzą przypadków do realizacji, bez runnera i bez `*_test.go`; sama nazwa `.testdomain.go` nie uruchamia testów Go. Ta decyzja jest wyjątkiem od wcześniejszego zakazu katalogów macierzy, nie pozwoleniem na JS parity, masowe automatyczne generowanie, procentowe bramki ani niezamówioną zmianę gramatyki/CI.

Przypadki opisują docelowe zachowanie języka, w tym brakujące konstrukcje. Nie dopasowuj oczekiwań do aktualnego outputu silnika i nie uznawaj `unsupported` za zaliczenie brakującej funkcji. Jawne wejście, kierunek, wymagania semantyczne, przykładowe propozycje przekładu, niejednoznaczność i pochodzenie mają być rozdzielone. Propozycja agenta nie oznacza akceptacji właściciela; obserwacje i `NOT RUN` zapisujemy osobno. Reguły i pierwsza próbka: [`tests/README.md`](tests/README.md). Pozostałe zasady pracy szeregowej i wymagane kontrole nadal obowiązują.

## LIGHT LM i wykonywalna macierz — decyzja właściciela 2026-10-24

Właściciel polecił wdrożyć macierz I, niezależnej analizy/realizacji PL/EN i integracji oraz naprawy audytu fundamentu LIGHT LM. Ograniczone standardowe testy Go w `tests/` i autorskie wzorce I są teraz jawnie dopuszczone w zakresie aktywnego issue; zastępuje to wcześniejszy zakaz `*_test.go` dla tego zakresu. Testy językowe pozostają poza CI. Nie tworzymy własnego frameworka, masowego korpusu, JS parity ani procentowej bramki. Pełne uruchomienie nadal ma wykazywać wszystkie niezaliczone wymagania, bez `t.Skip` i bez zmiany oczekiwań pod output.

Naprawy są prowadzone szeregowo w [LIGHT-F #324](https://github.com/clickbytes/golingua/issues/324), a przyszłe jawne uczenie ma osobny, oczekujący [LIGHT-L #329](https://github.com/clickbytes/golingua/issues/329). Sam kierunek LIGHT LM nie dowodzi oszczędności energii ani równoważności z modelem neuronowym. Nie dodaje automatycznej pamięci do bieżącego zakresu; uczenie wymaga zamkniętego kontraktu, pochodzenia, obsługi konfliktu i cofnięcia.

Decyzje, zakresy zleceń, odbiór i odtwarzalne wyniki mają odzwierciedlenie w repo: [rejestr pracy](docs/work/light-lm/README.md) i właściwe issue/PR. Koordynator zachowuje własność semantyki oraz GitHub; jeden writer, małe delegowane zadania i rzeczywiste CI PR/main pozostają obowiązkowe. Ocena agenta i akceptacja językowa właściciela są osobnymi statusami.

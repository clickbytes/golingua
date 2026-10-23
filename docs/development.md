# Workflow rozwoju

Repozytorium jest rozwijane szeregowo: najwyżej jeden aktywny EPIC i child task. Codex Sol Medium jest jedynym koordynatorem i autorem operacji GitHub; Codex Luna Low i opcjonalny OpenCode YOLO wykonują wyłącznie zadania przydzielone przez Sol. Maksymalnie dwóch pomocników łącznie, z czego tylko jeden w danej chwili może pisać kod (i tylko w przydzielonych plikach). Domyślnie korzystamy z jednej working branch i implementacyjnego PR; wyjątek zatwierdzony dla G07 #60 (2026-10-21) pozwala na sekwencyjne, zweryfikowane commity bezpośrednio na main. Właściciel może też jawnie polecić jednorazową zmianę wprost na `main`; wymaga to sprawdzenia aktualnego SHA, zakresu i rzeczywistego CI i nie zmienia standardowego procesu kolejnych issue. Zanim zaczniesz task, odczytaj aktualny `main`, kolejkę issue i otwarte PR. W issue zapisz ownera, bazowy SHA, plany per plik/symbol, dodatnie i ujemne zachowanie oraz komendy weryfikacji. Nie aktywuj kolejnego taska, dopóki bieżący PR nie zostanie zreviewowany, scalony, a CI na `main` nie przejdzie.

## Branch, commit, PR i merge

W **G07 #60** obowiązuje wyjątek: nie twórz PR tylko dlatego, że wymaga go poniższy standard; sprawdź main SHA, brak konfliktującej pracy, dokonaj commit + push w jednym batchu, a następnie zweryfikuj rzeczywisty CI. Dla innych tasks wraca opisany niżej zwykły flow.

1. Utwórz jedną gałąź od świeżego `main` i zmieniaj wyłącznie zakres aktywnego issue.
2. Uruchom wymagane lokalne kontrole techniczne. Nie dodawaj ani nie uruchamiaj nowych testów Go w standardowym workflow; wyjątkiem są testy dokładnie zlecone w aktywnym issue i wyraźnie zatwierdzone przez właściciela. Nadal uruchamiaj istniejące gofmt/vet/build oraz odpowiednie CLI smoke. Wyniki zapisuj jako PASS, FAIL, INFRA lub NOT RUN z komendą i dowodem.
3. Otwórz jeden PR powiązany z issue. Sprawdź cały diff i wynik obowiązkowego `go-ci` na PR head.
4. Po scaleniu sprawdź SHA `main` i `go-ci` na merge commit. Zaktualizuj issue/epic z odnośnikiem do PR i workflow. Dopiero wtedy aktywuj następny child.
5. Manualna ocena PL/EN jest możliwa tylko na przykładach właściciela; bez nich zapisz NOT EVALUATED. Nie wymyślaj oczekiwanych wyników.

## Minimalne kontrole

```sh
gofmt -l <zmienione-pliki-go>
go vet ./...
go build ./cmd/golingua
go run ./cmd/golingua --version
go run ./cmd/golingua --help
```

Obowiązkowy CI to `go-ci` (`gofmt`, `vet`, `build`, start CLI oraz smoke `catalog --summary/--details --json`, które kompilują CORE/PL/EN i walencję). Nie dodawaj test runnera, goldów, zgodności z JavaScript ani ilościowego progu jakości języka. Brak przebiegu obowiązkowego TUI/CLI w issue oznacza PARTIAL/BLOCKED, nie DONE (chyba że właściciel jawnie zwolni dane kryterium).

## Opcjonalne narzędzia agenta

RTK i CodeGraph są opcjonalnymi narzędziami pracy, nie zależnościami Go. Ich brak ani brak indeksu nie może blokować kompilacji lub CI.

### RTK

Zainstaluj RTK zgodnie z aktualną instrukcją projektu RTK i sprawdź `rtk --version`. W tym repo integrację hooka Codex wytwarza `rtk init --codex`; sprawdź zmiany przed commitem. Hook jest lokalny dla projektu i wymaga zaufania hookom przez Codex. RTK może przepisywać wspierane komendy Bash przed standardowym approval/sandbox sprawdzeniem Codex. W zwykłej pracy preferuj `rtk git ...`, `rtk gh ...` i inne `rtk <polecenie>`. Można użyć surowego polecenia, jeśli RTK nie jest zainstalowany albo wynik jest nieczytelny.

Sprawdź integrację przez:

```sh
rtk --version
rtk gh issue list --repo clickbytes/golingua
rtk git status
```

### CodeGraph

Zainstaluj CodeGraph zgodnie z instrukcją dostawcy i zweryfikuj `codegraph --version`. Zbuduj lokalny indeks dla bieżącego repo zgodnie z `codegraph init --help`, a potem użyj `codegraph status` i `codegraph query main`. Dostępne polecenia zależą od wersji. Używaj CodeGraph do znalezienia symboli, kontekstu kodu i potencjalnego wpływu zmiany. Sprawdzaj ważne wyniki w plikach źródłowych; graf może być nieaktualny lub niepełny i nie rozstrzyga semantyki produktu.

`.codegraph/.gitignore` pozostawia bazę, WAL/SHM, logi, PID i artefakty lokalne poza repo. Commituj sam plik `.gitignore`, nigdy lokalny indeks. Po błędzie uprawnień lub nieaktualnym indeksie odnotuj wynik, napraw/odtwórz lokalny indeks z prawami bieżącego użytkownika albo użyj `rg` i zwykłego czytania plików. Nie zmieniaj właściciela plików przez `sudo` w ramach zwykłego workflow.

### Codex: konfiguracja projektu i użytkownika

Projektowy `.codex/config.toml` zawiera wpis MCP CodeGraph i jest ładowany tylko dla zaufanego repozytorium. Hook RTK jest w `.codex/hooks.json`; po zmianie trust lub konfiguracji uruchom ponownie Codex, zaakceptuj hook w zaufanym repo i sprawdź serwer przez `codex mcp list` / próbę połączenia w kliencie. Konfiguracja projektu ma wyższy priorytet niż ustawienia użytkownika dla tych wpisów. Nie kopiuj MCP CodeGraph ani hooka RTK do `~/.codex/config.toml`, jeśli są już dostarczane projektowo. Ustawienia osobiste, takie jak model i trust innych projektów, pozostają osobiste. Użytkownik może dodać osobny globalny CodeGraph dla innych repozytoriów, ale wtedy usuń duplikat w projekcie albo globalny wpis.

Polecenie MCP CodeGraph w `.codex/config.toml` odzwierciedla wyjście `codegraph install --print-config codex` dla lokalnie sprawdzonej wersji narzędzia. Po aktualizacji CodeGraph sprawdź wydrukowaną konfigurację i dostosuj argumenty komendy. Projekt nie instaluje narzędzi z sieci ani nie tworzy indeksu podczas builda/CI.

## Koordynator i delegacja (G11-P0 #289)

Obowiązuje [protokół wykonania](agent-execution.md) i [szablon job brief](agent-job-brief-template.md). Sol wybiera delegację tylko gdy oszczędza pracę; Luna Low wykonuje ograniczony research/review lub małą edycję, OpenCode YOLO jest opcjonalnym wykonawcą w izolowanym środowisku. Workerzy nie zapisują GitHub i nie zmieniają epika, issue ani planowania. Po jednej implementacji i najwyżej jednej uzasadnionej korekcie Sol kończy zadanie jako DONE/PARTIAL/BLOCKED; bez ponawiania identycznych błędów.

Nie wdrażamy hipotetycznego GoLingua v2, migracji czy adapterów zgodności. Zatwierdzona biblioteka Bubble Tea v2 nie jest nową wersją GoLingua.

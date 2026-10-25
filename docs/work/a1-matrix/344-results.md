# A1-M8 #344 — wyniki końcowe

Status implementacji: ukończony zakres w [PR #351](https://github.com/clickbytes/golingua/pull/351). Operacyjny odbiór CI i scalenia jest zapisany w issue #344 i PR. Epik: [#337](https://github.com/clickbytes/golingua/issues/337); child: [#344](https://github.com/clickbytes/golingua/issues/344). Ocena agenta i obserwacje nie są akceptacją właściciela: `NOT_EVALUATED`.

M2–M7 są scalone z zielonym main CI. M8 obejmuje obecnie 500 par / 1000 kierunków A1–A1.13. Wszystkie zaplanowane grupy mają docelową liczebność: actions 65, identification 40, possession 40, location_quantity 50, negation 50, questions 60, time_agreement 60, case_adjective_order 45, ambiguity_reference 45, composition_boundary 45. To cele planistyczne, nie bramki poprawności. Partie nie narzucają limitu inwentarza ani nie certyfikują CEFR.

## Pełny przebieg obserwacyjny

Na źródle `f113035+working-tree` wykonano `go test ./tests/... -count=1 -v`; kod wyjścia wyniósł 1 z powodu widocznych wyników macierzy. Raport i surowy log tego przebiegu: `/tmp/golingua-matrix-expansion/m8-final-report.md` i `/tmp/golingua-matrix-expansion/m8-final-full.log`.

Wszystkie 3000 kontroli dla 1000 kierunków są zachowane:

| Status | Liczba kontroli |
|---|---:|
| PASS | 297 |
| FAIL | 244 |
| GAP_ANALYZE | 946 |
| GAP_REALIZE | 433 |
| GAP_I | 1080 |
| SPEC_PENDING | 0 |
| INFRA | 0 |
| Check NOT RUN | 0 |

Raport zachowuje też 205 dodatkowych kontroli liściowych oraz oddzielne 32 kontrole odmów. Runtime ma 362 jawne znaczniki `independent_realization_executed:false` wraz z ID i powodami; obejmują one również dodatkowe kontrole. To liczba niewykonanych wywołań adaptera, nie brakujących wyników kontroli. Statusy źródłowej analizy, niezależnej realizacji i integracji są liczone osobno.

Pełny raport rozróżnia zadeklarowane tagi i wymagania, autorskie `meaning_input`, zaobserwowane grafy/ślady źródłowe i wynikowe oraz pola trace `family_id`/`variant_id`/`frame_id`, cechy i diagnostyki. Powiązanie zaobserwowanego sense z lexeme/paradigm jest metadanymi katalogu, a nie dowodem wykonania każdej formy lub procedury morfologicznej. Nie wyciąga się procentowej oceny poprawności ani wniosków o pokryciu słownika.

Wejścia nowych partii uruchomiono w etapowych próbach CLI. Zestawienia etapów i zakresów są w [rejestrze pracy](README.md). Zachowane odpowiedzi CLI obejmują [M3](evidence/339-cli.json), [M4](evidence/340-cli.json), [M5](evidence/341-cli.json), [M6](evidence/342-cli.json), [M7](evidence/343-cli.json), [A1.4](evidence/344-actions-cli.json), [A1.5](evidence/344-identification-cli.json), [A1.6](evidence/344-possession-cli.json), [A1.7](evidence/344-location-cli.json), [A1.8](evidence/344-negation-cli.json), [A1.9](evidence/344-questions-cli.json), [A1.10 osobowe](evidence/344-time-person-cli.json), [A1.10 pozostałe](evidence/344-time-other-cli.json), [A1.11](evidence/344-case-cli.json), [A1.12](evidence/344-ambiguity-cli.json) i [A1.13](evidence/344-composition-cli.json). Są to obserwacje silnika, oddzielone od kontraktów i akceptacji właściciela.

## Odtworzenie

```sh
matrix_sha=$(git rev-parse HEAD)
GOLINGUA_MATRIX_SHA="$matrix_sha" go test ./tests/... -count=1 -v > /tmp/golingua-m8-full.log 2>&1
matrix_exit=$?
python3 -B scripts/report-language-matrix.py --sha "$matrix_sha" --exit-code "$matrix_exit" --log /tmp/golingua-m8-full.log > /tmp/golingua-m8-report.md
```

Na tym źródle koordynator potwierdził `gofmt`, `git diff --check`, `go vet ./...`, `go build ./cmd/golingua`, smoke `--help`/`--version` i wymagane kontrole skupione: PASS. Pełne `go test` ma exit 1 opisany powyżej. Wyniki rzeczywistego CI PR/main i scalenia są utrzymywane w odbiorze GitHub #344 / PR #351.

Zmiana nie dodaje gramatyki produktu ani mechanizmu uczenia. Aktualizację raportu `tests/results/latest.md` wykona istniejący workflow dopiero po zmianie `main`; plik nie jest ręcznie edytowany.

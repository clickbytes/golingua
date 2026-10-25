# Fundament LIGHT LM

Decyzja właściciela z 2026-10-24: rozwijać silnik pojęciowo-językowy z proceduralną odmianą i przyszłym uczeniem symbolicznym; naprawić problemy wykazane w audycie i zachować decyzje, zlecenia oraz dowody w repozytorium. Koordynator odpowiada za semantykę, integrację, ocenę i GitHub. Akceptacja językowa właściciela pozostaje osobna.

## Kolejność

| Zakres | GitHub | Stan |
| --- | --- | --- |
| Fundament znaczenia, procedur i uogólniania | [epik 324](https://github.com/clickbytes/golingua/issues/324) | lokalny zakres odebrany; końcowe merge/CI w epiku |
| Czas i wielość uczestników w I | [325](https://github.com/clickbytes/golingua/issues/325) | DONE |
| Wspólne proceduralne klasy odmiany | [326](https://github.com/clickbytes/golingua/issues/326) | DONE |
| Składanie i realizacja niezależne od konkretnego kota | [327](https://github.com/clickbytes/golingua/issues/327) | DONE |
| Ograniczenie kopiowania danych w runtime | [328](https://github.com/clickbytes/golingua/issues/328) | odbiór lokalny PASS; końcowe merge/CI w issue |
| Jawne uczenie i trwałe użycie nabytej wiedzy | [epik 329](https://github.com/clickbytes/golingua/issues/329) | QUEUED |

Jeden ACTIVE child, jedna gałąź/PR i jeden writer. Kolejny zakres rozpoczyna się po merge, zielonym CI końcowego main i zamknięciu poprzednika. Pełna macierz jest wykonywana lokalnie oraz raportowana po pushu na `main` przez osobny, nieblokujący workflow; techniczne `go-ci` pozostaje bez językowej bramki. Najnowszy raport: [`tests/results/latest.md`](../../../tests/results/latest.md), surowy log: artefakt danego uruchomienia Actions.

## Punkt wyjścia

Audyt obejmował main `38870480b03a95b58ee89bd7cb7011cd3c80060c`. Rzeczywiste próby wykazały identyczne grafy dla `The teacher reads/read the message.` oraz `The teacher/teachers see the dog.`; różnice zachowywało Evidence parsera. Bez niego realizator wybierał teraźniejszość i pojedynczość z jawnymi założeniami. Analiza `This is a dog.` i `To jest pies.` działała, ale realizacja nie zwracała tekstu. `The black cat sees the dog.` było unsupported. To ograniczenia zaobserwowanych tras, nie ocena pełnego A1.

Pomiar lokalny Go 1.27.0, Linux amd64, i7-11800H, GOMAXPROCS=16: inicjalizacja 1078 ms, żywa sterta po GC 18 676 168 B; mediana załadowanego tłumaczenia 1,82–4,88 ms i 3,81–5,81 MB alokacji na żądanie (trzy krótkie wejścia, po 10 rozgrzewających i 100 mierzonych wywołań). Nie mierzono energii ani przewagi nad równoważnym modelem neuronowym. Profil wskazał kopiowanie leksykonu/map w SelectTargets i porównywanie całego CORE. Reprodukcja i pomiar porównawczy należą do #328.

## Wykonanie i dokumentacja

- [Kontrakt czasu i wielości](../../architecture/meaning-attributes.md).
- [Briefy i przekazania 325](325-jobs.md).
- [Lokalne dowody odbioru 325](325-evidence.md) i [obserwacje CLI](325-observations.jsonl); bieżący stan merge/CI jest zapisany w issue/PR.
- #325 scalono przez [PR 330](https://github.com/clickbytes/golingua/pull/330): main `1707590ead2892f16f065fe52eaaf935e9e9216e`; [CI PR](https://github.com/clickbytes/golingua/actions/runs/40134700533) i [CI main](https://github.com/clickbytes/golingua/actions/runs/40134811089) SUCCESS. Drzewo merge jest zgodne ze sprawdzonym head PR.
- [Kontrakt klas 326](326-contract.md) i [briefy oraz przekazania](326-jobs.md).
- [Dowody odbioru 326](326-evidence.md) i [obserwacje CLI](326-observations.jsonl).
- #326 scalono przez [PR 331](https://github.com/clickbytes/golingua/pull/331): main `11772b75729ac1ef1f7dc1f4abff51a5f7269363`, drzewo zgodne z odebranym head. [CI PR](https://github.com/clickbytes/golingua/actions/runs/40135648747) i [CI main](https://github.com/clickbytes/golingua/actions/runs/40135704117) SUCCESS.
- [Kontrakt składania 327](327-contract.md), [briefy oraz przekazania](327-jobs.md), [dowody odbioru](327-evidence.md) i [obserwacje CLI](327-observations.jsonl).
- Ocena techniczna agenta nie oznacza akceptacji właściciela, obecnie NOT EVALUATED.

OpenCode: dzienna kontrola 2026-10-24, izolowany obraz `ghcr.io/anomalyco/opencode:2.0.15`, jawny worker i limit 12 kroków. Config wskazywał plik, ale `debug agents` zwrócił pustą listę; katalogi modeli były puste, a katalog kosztów odpowiadał 403. Model, darmowy koszt i worker nie zostały zweryfikowane. Próba odpowiedzi i odmowy uprawnień: NOT RUN; `--auto` nie uruchomiono. Zgodnie z AGENTS stosujemy ograniczone zadania Luna Low. Źródła kontroli: [config](https://dev.opencode.ai/v2/docs/config/), [permissions](https://dev.opencode.ai/v2/docs/permissions/), [agents](https://dev.opencode.ai/v2/docs/agents/). Konfiguracja narzędzia i poświadczenia nie są danymi produktu.

Odświeżenie 2026-10-25 (Europe/Warsaw), ten sam obraz v2.0.15, nowy izolowany kontener bez `.git` i poświadczeń: jawny config został odczytany; `models` pokazało dziewięć darmowych modeli, a [katalog cen](https://models.dev/api.json) pobrany w kontenerze potwierdził zerowe stawki. Diagnostyczne `opencode/glm-5.3-flash-free` odpowiedziało `READY`, koszt próby narzędziowej 0. `debug agents` nadal nie wykazało wskazanego workera, mimo przyjęcia `--agent` w run. Rzeczywisty odczyt pliku poza workspace został odrzucony przez `external_directory`; dla shell, sieci i subagenta nie zaobserwowano prób narzędziowych (model zgłosił ich brak), więc wymagane trzy odmowy pozostają NOT VERIFIED. Pełna bramka nadal niespełniona, bez `--auto`; kolejne zamknięte zadania używają Luna Low. Przegląd [SWE-bench](https://www.swebench.com/) i [Terminal-Bench](https://www.tbench.ai/) nie dostarczył odczytywalnego wyniku dla dokładnego darmowego endpointu, więc nie przypisano mu pozycji w rankingu. Notatka i surowe próby koordynatora: `/tmp/golingua-opencode-audit-20261025/`; kontener po kontroli usunięto.

#327 scalono przez [PR 332](https://github.com/clickbytes/golingua/pull/332): main `57e4882e79eb6596caed5126108fa272177e6982`, drzewo zgodne z head `8a8817761754d1c64c5518ab7ce92ec62fa2136f`. [CI PR](https://github.com/clickbytes/golingua/actions/runs/40140701015) i [CI main](https://github.com/clickbytes/golingua/actions/runs/40140774211) SUCCESS. Po #327 aktywowano ograniczenie kopiowania runtime #328; uczenie #329 pozostaje QUEUED.

- [Kontrakt kosztu runtime 328](328-contract.md), [zlecenia](328-jobs.md), [dowody i reprodukcja](328-evidence.md), [pomiary](328-measurements.json), [profile](328-profiles/manifest.json) oraz [obserwacje CLI](328-observations.jsonl).

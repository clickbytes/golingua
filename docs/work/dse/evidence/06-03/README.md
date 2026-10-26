# #393: atrybuty semantyczne — lokalne evidence

To zapis rzeczywistych obserwacji z zamrożonego lokalnego wykonania, a nie
gold, test runner ani akceptacja językowa. Pliki wejściowe i odpowiedzi
pozostają w repo, więc raport nie zależy od katalogu `/tmp`.

## Uruchomione CLI

Binarium: `/tmp/golingua-dse/bin/golingua-393-final`, SHA-256
`ede4172b2fb757c7f69e34a19d62cfa7d388b94b1ab51223f3ebdbb707e66a45`.
Zarejestrowano 114 wywołań CLI: 79 zakończonych kodem 0, 28 kodem 2 i 7 kodem
3. Rozkład według surowych rejestrów:

- 99 głównych obserwacji: 70×0, 23×2, 6×3.
- 12 dodatkowych obserwacji: 7×0, 4×2, 1×3.
- 3 kontrole kontekstu: 2×0, 1×2.

Kody 2 i 3 są oczekiwanymi odmowami dla błędnych atrybutów albo
nieobsługiwanej ewaluacji; nie są zaliczonymi wynikami semantycznymi. Osobna
kontrola zgodności legacy CLI miała jeszcze jedno wywołanie z kodem 0; nie
wchodzi ono do powyższych 114 rekordów.
Przykładowy, ponownie uruchomiony przepływ z repozytoryjnych danych znajduje
się w [obserwacjach docflow](observations/docflow/docflow-commands.jsonl).
Jego surowe odpowiedzi potwierdzają: kontrola dokumentu, budowa Modal,
ewaluacja `one`, eksport, inspekcja i replay zwróciły `ok`; ewaluacja
`past` i `many` zwróciła `unsupported` (kod 3); porównanie `Not(Modal(P))`
z `Modal(Not(P))` dało `verdict: undetermined`. Inspekcja zachowuje jawne pola,
ale raportuje `metadata_verified: false`.

## Artefakty

- `observations/393-cli-observations.jsonl`, `393-extra-cli-observations.jsonl`
  i `393-context-controls.jsonl` zachowują argumenty, kody wyjścia i skróty
  odpowiedzi dla wszystkich 114 wywołań.
- Katalogi `observations/cli/` i `observations/extra-cli/` zawierają
  odpowiadające im wejścia/żądania i surowe odpowiedzi CLI. Pełna lista
  skopiowanych plików, rozmiary i SHA-256 są w
  [`observations-file-manifest.json`](observations-file-manifest.json).
- `inputs/independent/` i `inputs/extra/` przechowują odpowiednio 46
  niezależnych plików wejściowych i 31 plików dla dodatkowych kontroli.
- `observations/393-adapter-observations.jsonl` zawiera sześć oddzielnych
  prób API adaptera Meaning; nie są to wywołania CLI.
- `393-semantic-observation.json` podsumowuje zachowane wartości i odmowy.
  `393-final-legacy-observation.json` oraz oba pliki `*-export.json`
  porównują trzy fingerprinty z baseline #392: wartości resource, meaning i
  cache pozostały identyczne.
- `393-schema-observation.json` raportuje 42 artefakty walidacji schematów
  bez błędów; `393-source-hashes.json` zapisuje skróty 37 plików źródłowych i
  schematów. `393-technical-observation.json` zawiera rzeczywiste komendy
  `gofmt`, `go vet ./...`, build oraz `--help`/`--version`, wszystkie z kodem
  0. Nie uruchamiano `go test` ani nie dodawano testów.

Przykłady produktu są w [`data/semantic/examples/attributes/`](../../../../../data/semantic/examples/attributes/README.md).
Wyniki pokazują zachowanie reprezentacji oraz wąski profil ewaluacji, nie
twierdzą, że czas, modalność albo `many` są już obsługiwane przez ewaluator.
Ocena agenta i status kontroli technicznych są odrębne od akceptacji
właściciela: `NOT_EVALUATED`. PR i CI `main` pozostają oczekujące; status
zadania znajduje się w [issue map](../../issue-map.json).

Koordynator niezależnie uruchomił wszystkie 17 literalnych poleceń nowej dokumentacji (8 z `docs/cli.md`, 9 z przykładów): 15 × kod 0 oraz 2 × oczekiwane `unsupported`, kod 3. Osobny zapis: [root docflow](observations/393-root-docflow-observation.json).

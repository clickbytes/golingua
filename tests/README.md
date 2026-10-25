# Macierz zachowania PL/EN

Obecny inwentarz obejmuje **500 par / 1000 kierunków**: próbkę A1 oraz kolejne partie A1.1–A1.13. Zaplanowane liczebności wszystkich dziesięciu grup zostały osiągnięte; to opis zakresu próby, nie obsługi gramatycznej przez silnik. Partie A1.1–A1.13 nie wyznaczają limitu rekordów ani poziomu biegłości i nie stanowią certyfikacji CEFR. Właściciel może rozwijać inwentarz o kolejne jawnie wybrane partie. Każdy kierunek ma `PairID` i `PrimaryGroup`; `translation.LanguageBatches` rejestruje wspólne zbiory, a `coverage-targets.json` zawiera cele planistyczne, nie bramki poprawności. Teksty, wymagania i akceptacja właściciela są rozdzielone; niezatwierdzone kontrakty mają `NOT_EVALUATED`. Aktualny odbiór i status PR opisuje [podsumowanie #344](../docs/work/a1-matrix/344-results.md).

## Układ i kontrakt

```text
tests/
  interlingua/                 # niezależne, ręcznie zapisane znaczenia I
  languages/pl/A1/             # analiza i realizacja po polsku
  languages/en/A1/             # analiza i realizacja po angielsku
  languages/pl/A1.1..A1.13/   # kolejne wybrane partie po polsku
  languages/en/A1.1..A1.13/   # kolejne wybrane partie po angielsku
  translation/cases.go         # deklaratywne ID i odwołania do obu stron
  translation/pl-en/           # PL → I → EN
  translation/en-pl/           # EN → I → PL
  matrix/                      # wspólne typy, runtime produkcyjny i kontrole
```

A1.1–A1.13 oznaczają kolejne partie próby PL/EN z jawnie wybranymi zjawiskami. Każda zachowuje własne briefy, listę wejść i obserwacje w [rejestrze etapów](../docs/work/a1-matrix/README.md); pełny wybór został też uruchomiony w staged CLI evidence. Akceptacja językowa właściciela pozostaje odrębna od obserwacji agenta.

Pliki `*.testdomain.go` zawierają dane. Standardowe kontrole Go rozdzielają analizę źródła, niezależną realizację każdego wymaganego Meaning i integrację przez pipeline produkcyjny. Zbiory `AllowedTexts`, `RequiredTexts`, grupy one-of i warianty per Meaning określają osobno dopuszczalne style i wymagane wyniki; propozycja nie staje się automatycznie oczekiwanym wynikiem. Pliki `*_test.go` wykonują te kontrole oraz osobne testy odmów, metamorficzne, wieloznaczności i wsparcia. Nie dolicza się ich do PASS zwykłego tłumaczenia. Obserwacyjny workflow GitHub Actions na push do `main` publikuje [`tests/results/latest.md`](results/latest.md) i surowy log jako artefakt; raport nie blokuje technicznego `go-ci` i zmieni się dopiero po merge na `main`. Każda analiza zachowuje wejście, wymagania, nierozstrzygnięcia i pochodzenie. Realizacja zachowuje `ProposalTexts`, `AllowedTexts`, `RequiredTexts`, dozwolone/obowiązkowe założenia i straty. `RequiredTexts` musi być podzbiorem `AllowedTexts`; pusta lista `AllowedTexts` pozostawia kontrakt otwarty.

Integracja wskazuje ID analizy i realizacji; nie kopiuje ich treści. Inwentarz sprawdza unikalność ID, poprawny kierunek, wszystkie odwołania, brak osieroconych przypadków oraz kompletność wymaganych pól provenance. `CoverageTags` w rekordzie integracji opisują zadeklarowane zjawiska; nie dowodzą wykonania ani poprawności konstrukcji. Raport oddziela je od statusów rzeczywiście zaobserwowanych kontroli i pokazuje kontrole dodatkowe poza rekordami integracji. Analiza i realizacja PL działają niezależnie od EN i odwrotnie. Wykonanie przekładu uruchamia oba adaptery produkcyjne przez rzeczywisty pipeline.

Wzorce `MeaningGraph` są tworzone niezależnie od parserów. Ograniczone porównanie grafów jest niewrażliwe na lokalne NodeID i kolejność węzłów, ale zachowuje role i bijekcję referentów. Nie scala odrębnych węzłów tylko dlatego, że mają ten sam ConceptID. Wzorce nie dopisują tożsamości osobników ani płci zwierzęcia. Kontekst wskazania przechowuje osobną nierozstrzygniętą Mention bez `ReferentID` i `EvidenceID`.

Negacja w obecnym wycinku ma własne znaczenia I: `negation.classification.this` i `.to` umieszczają jeden zakres negacji nad stanem `class_membership`; `negation.seeing.teacher-dog` umieszcza go nad zdarzeniem widzenia. Każdy graf ma jeden korzeń będący zakresem negacji, bez dodatkowego korzenia predykatu. Dla klasyfikacji adapter celu odrzuca kontekst, którego graf ma inną polaryzację niż znaczenie.

Wynik silnika (`ok`, `ambiguous`, `incomplete`, `unsupported`) jest oddzielony od oceny przypadku. Kontrole sprawdzają rzeczywiste grafy, konteksty, teksty, założenia, straty i diagnostyki; sam status `ok` nie oznacza PASS. Brak zamkniętego wzorca daje `SPEC_PENDING`, nie domyślny PASS ani udowodniony błąd parsera. Nie używamy `t.Skip`.

| Ocena | Znaczenie |
| --- | --- |
| `PASS` / `FAIL` | Wykonany przypadek spełnia kontrakt / wynik jest z nim sprzeczny. |
| `GAP_ANALYZE` / `GAP_REALIZE` | Brakuje wymaganej analizy / realizacji danego języka. |
| `GAP_I` | Obecna reprezentacja nie potrafi wyrazić wymaganego znaczenia. |
| `SPEC_PENDING` / `NOT RUN` | Brakuje zamkniętego kontraktu / kontrola nie została wykonana. |
| `INFRA` | Runtime lub konfiguracja uniemożliwia wiarygodne wykonanie. |

Żaden GAP, FAIL, SPEC_PENDING, NOT RUN ani INFRA nie jest pomijany ani zaliczany jako zielony wynik.

## Identyfikacja objęta zamkniętym kontraktem

Znaczenie klasyfikacji to `class_membership(member → cat)`: wskazany referent jest klasyfikowany jako kot. To nie jest identyczność z innym osobnikiem ani aktualizacja ontologii. Znaczenie opisu koloru to `black(bearer → cat)`: czarny kolor przypisany temu wskazanemu kotu.

| ID I | Kotwica nierozstrzygniętej Mention |
| --- | --- |
| `identification.classification.this` | `deixis=proximal` |
| `identification.classification.to` | `deixis=unresolved` |
| `identification.black.this` | `deixis=proximal` |
| `identification.black.ten` | `deixis=demonstrative` (odległość nieznana) |

Realizacja EN dla nierozstrzygniętego `to` zachowuje warianty `This is a cat.`, `That is a cat.` i `It is a cat.`; dla `ten` zachowuje `This cat is black.` i `That cat is black.`. Każdy wybrany angielski wariant ma jawne założenie `reference`. Rodzime I z kotwicą `proximal` daje tylko wariant `This` bez takiego założenia. Polski `to`/`ten` nie koduje jawnie bliskości `this`; przekład z EN deklaruje stratę `deixis`. Założenia i straty są odrębnymi informacjami.

## Odbiór identyfikacji

Przed testami zapisz commit, którego wyników używasz:

```sh
export GOLINGUA_MATRIX_SHA="$(git rev-parse HEAD)"
```

Jeśli working tree zawiera zmiany, przed uruchomieniem obejrzyj i zapisz zakres diffu, a SHA oznacz jako `$(git rev-parse HEAD)+working-tree`. Bez tej zmiennej logi pokazują `UNRECORDED` i nie są dowodem przypisanym do SHA. Zachowuj surowy output oraz obok niego dokładne polecenie i exit code; nie zmieniaj expected na podstawie outputu.

```sh
go test ./tests/interlingua/... -run Identification -count=1 -v
go test ./tests/languages/pl/A1/... -run Identification -count=1 -v
go test ./tests/languages/en/A1/... -run Identification -count=1 -v
go test ./tests/translation/... -run Identification -count=1 -v
```

## Odbiór negacji

Uruchom cztery niezależne kontrole negacji oraz kontrolne Seeing i inwentarz:

```sh
go test ./tests/interlingua/... -run Negation -count=1 -v
go test ./tests/languages/pl/A1/... -run Negation -count=1 -v
go test ./tests/languages/en/A1/... -run Negation -count=1 -v
go test ./tests/translation/... -run Negation -count=1 -v
go test ./tests/... -run 'Negation|Identification|Seeing|Inventory' -count=1 -v
```

Ostatnie polecenie sprawdza również zgodność polaryzacji grafu znaczenia i kontekstu wskazania w obu adapterach. Przykłady negacji nie zastępują kontroli zamiany ról z istniejących przypadków Seeing.

## Seeing, inwentarz i pełny obraz

Zakres [LIGHT-F1](../docs/work/light-lm/README.md) dodaje ograniczone kontrole `MeaningAttributes` dla istniejących tras, bez nowych rekordów A1. Wzorce Reading są zapisane niezależnie od analizatorów; Seeing i Identification mają jawne present/one. Testy sprawdzają osobno analizę, przygotowane I bez Evidence parsera, integrację, walidację i diagnostykę sprzecznych danych. Kontrakt: [czas i wielość](../docs/architecture/meaning-attributes.md).

```sh
go test ./tests/interlingua/... -run MeaningAttributes -count=1 -v
go test ./tests/languages/pl/A1/... -run MeaningAttributes -count=1 -v
go test ./tests/languages/en/A1/... -run MeaningAttributes -count=1 -v
go test ./tests/translation/... -run MeaningAttributes -count=1 -v
go test ./tests/... -run 'MeaningAttributes|Negation|Identification|Seeing|Inventory' -count=1 -v
```

Present Reading używa istniejącej trasy the/the z jawnymi założeniami określoności; past zachowuje licencjonowane warianty rodzajnika. To granica dotychczasowego zakresu G10-19/G10-20, nie nowa ocena pełnego języka.

## Złożona cecha podmiotu i uogólnienie

Kontrakt [LIGHT-F3](../docs/work/light-lm/327-contract.md) dodaje cztery niezależne I: `composition.black-cat-sees-dog`, `composition.black-dog-sees-cat`, `composition.black-bird-sees-dog` i `composition.black-cat-sees-cat`. Każde łączy stan `black` z rolą `bearer` wskazującą dokładnie referent podmiotu oraz zdarzenie `seeing` w czasie teraźniejszym. Oba referenty mają `one`; pozostają odrębnymi węzłami także wtedy, gdy oba mają ConceptID kota. Przykładowe uogólnienie referenta psa obejmuje `This is a dog` / `To jest pies.`, `This dog is black.` / `Ten pies jest czarny.` i `This is not a dog.` / `To nie jest pies.`.

Złożona trasa widzenia jest ograniczona do jednego teraźniejszego zdarzenia, jednego nieczasowanego stanu `black`, podmiotu z zatwierdzonym leksemem i obsługiwaną formą oraz dwóch referentów o znanej liczbie pojedynczej. Nie rozstrzyga tożsamości ani płci. PL realizuje zgodę z podmiotem; EN zachowuje jawne warianty określoności. Inne modyfikatory, czasy, liczby, brak wiązania lub formy wymaganej do realizacji wymagają odmowy. Kontrole są lokalne i nie zmieniają rekordów próbki ani CI.

```sh
go test ./tests/interlingua/... -run Composition -count=1 -v
go test ./tests/languages/pl/A1/... -run Composition -count=1 -v
go test ./tests/languages/en/A1/... -run Composition -count=1 -v
go test ./tests/translation/... -run Composition -count=1 -v
go test ./tests/languages/pl/A1/... -run Generalization -count=1 -v
go test ./tests/languages/en/A1/... -run Generalization -count=1 -v
go test ./tests/... -run 'Composition|Generalization' -count=1 -v
```

```sh
go test ./tests/interlingua/... -run Seeing -count=1 -v
go test ./tests/languages/pl/A1/... -run Seeing -count=1 -v
go test ./tests/languages/en/A1/... -run Seeing -count=1 -v
go test ./tests/translation/... -run Seeing -count=1 -v
go test ./tests/translation/... -run Inventory -count=1 -v
go test ./tests/... -count=1 -v
```

**Historia M3:** wynik zapisany dla jego ówczesnego snapshotu wykazał 24 `GAP_I`, 6 `GAP_ANALYZE`, 6 `GAP_REALIZE` i zero `SPEC_PENDING`; zobacz [odbiór M3](../docs/work/a1-matrix/339-results.md). I dla dokładnie dwóch kotów reprezentuje legalny komponent `QuantifierExactly(2)`, ale nie pełne teraźniejsze zdanie egzystencjalne: brak wiązania z rolą istnienia i brak czasu. Obecne pełne wymaganie „There are two cats” pozostaje `GAP_I`; wyjaśnia to [audyt liczebności](../docs/work/a1-matrix/344-quantity-audit.md). Wcześniejszy wynik M2 dla 24 rekordów / 72 kontroli, w tym 36 `SPEC_PENDING`, jest historycznym snapshotem na podanym base SHA w [wynikach M2](../docs/work/a1-matrix/338-results.md), a nie aktualnym licznikiem. Aktualny pełny przebieg #344, wszystkie 1000 kierunków oraz podział statusów opisuje [wynik #344](../docs/work/a1-matrix/344-results.md). `SPEC_PENDING` i `NOT RUN` pozostają widoczne i nie oznaczają PASS.

## Obserwacje CLI dla czterech oryginalnych zdań

Na dostępnej próbce otrzymano następujące obserwacje z CLI:

| ID | Wejście i kierunek | Tekst kandydata | Status silnika / exit |
| --- | --- | --- | --- |
| `A1-ID-001-EN-PL` | `This is a cat` · EN→PL | `To jest kot.` | `incomplete` / 7; strata `deixis` |
| `A1-ID-002-PL-EN` | `To jest kot.` · PL→EN | `This is a cat.`, `That is a cat.`, `It is a cat.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |
| `A1-ID-003-EN-PL` | `This cat is black.` · EN→PL | `Ten kot jest czarny.` | `incomplete` / 7; strata `deixis` |
| `A1-ID-004-PL-EN` | `Ten kot jest czarny.` · PL→EN | `This cat is black.`, `That cat is black.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |

Są to wyniki silnika zgodne z ograniczonym kontraktem testowanej próbki; `incomplete` i `ambiguous` są jawne i nie rozstrzygają tożsamości ani płci. Końcowy surowy JSON z SHA i poleceniami: `/tmp/a1-t05-cli.json` (pierwsze cztery obserwacje to identyfikacja; kolejne to wybrane kontrole istniejących tras). Odtworzenie tych czterech wejść wykonuje się po przygotowaniu runtime poniżej.

## Obserwacje CLI dla negacji

Na końcowym snapshotcie CLI zwróciło poniższe wyniki. Status i exit CLI są obserwacją silnika; założenia i straty pozostają osobnymi polami.

| ID | Wejście i kierunek | Teksty kandydatów | Status / exit |
| --- | --- | --- | --- |
| `A1-NEG-001-EN-PL` | `This is not a cat.` · EN→PL | `To nie jest kot.` | `incomplete` / 7; strata `deixis` |
| `A1-NEG-002-PL-EN` | `To nie jest kot.` · PL→EN | `It is not a cat.`, `That is not a cat.`, `This is not a cat.` | `ambiguous` / 6; założenie `reference` dla każdej opcji |
| `A1-NEG-003-EN-PL` | `The teacher does not see the dog.` · EN→PL | `Nauczyciel nie widzi psa.` | `ok` / 0 |
| `A1-NEG-004-PL-EN` | `Nauczyciel nie widzi psa.` · PL→EN | `The teacher does not see the dog.`, `The teacher does not see a dog.`, `A teacher does not see the dog.`, `A teacher does not see a dog.` | `ambiguous` / 6; dobór `determiner` jest jawny dla każdej opcji |

Wszystkie cztery grafy pośrednie mają jeden korzeń zakresu negacji. Dodatkowe kontrole zachowały zamianę ról, liczbę mnogą przy słuchaniu/patrzeniu, niejednoznaczność czytania, a podwójne lub przestawione `not`/`nie` otrzymały `unsupported`. Surowe obserwacje z poleceniami, exit code i SHA: `/tmp/a1-neg-final-cli.json`; testy I, PL, EN, przekładu, kontrole, vet, build i pełnej macierzy: `/tmp/a1-neg-final-{i,pl,en,translation,controls,vet,build,full}.log`. To historyczny przebieg sprzed finalnej macierzy #344; jego 36 przypadków `SPEC_PENDING` nie jest bieżącym wynikiem. Aktualny przebieg i reprodukcja są w [wynikach #344](../docs/work/a1-matrix/344-results.md).

## Przygotowanie świeżego snapshotu i CLI

Z katalogu repozytorium zbuduj bieżący program i snapshot do katalogów tymczasowych:

```sh
go build -o /tmp/golingua-a1-neg ./cmd/golingua
/tmp/golingua-a1-neg snapshot --apply --store /tmp/golingua-a1-neg-store --json
cat > /tmp/golingua-a1-neg-runtime.json <<'JSON'
{"schema_version":1,"dataset_store":"/tmp/golingua-a1-neg-store","dataset_id":"golingua-authored","language_package_dirs":[]}
JSON
```

Uruchom cztery rekordy osobno; `--text` zachowuje dokładny zapis wejścia:

```sh
/tmp/golingua-a1-neg translate --from en --to pl --text 'This is a cat' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'To jest kot.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from en --to pl --text 'This cat is black.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'Ten kot jest czarny.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
```

Odtwórz cztery nowe przepływy negacji tym samym snapshotem:

```sh
/tmp/golingua-a1-neg translate --from en --to pl --text 'This is not a cat.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'To nie jest kot.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from en --to pl --text 'The teacher does not see the dog.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
/tmp/golingua-a1-neg translate --from pl --to en --text 'Nauczyciel nie widzi psa.' --runtime-config /tmp/golingua-a1-neg-runtime.json --json
```

## Odmowy i jawne założenia wieloznaczności

Granice wejścia są osobną kategorią poza `translation.Cases`: 16 autorskich, wadliwych lub nieobsługiwanych wejść sprawdza natywną analizę i pipeline tłumaczenia, łącznie 32 kontrole. Odmowa zalicza się przy `unsupported`/`invalid` bez grafu Meaning ani tekstu. Wąsko dozwolone `unknown` wymaga niepustego uzasadnienia i diagnostyki wyłącznie dla pustego wejścia w natywnym analizatorze oraz nieznanych leksemów w analizatorze i przekładzie. Dla pustego wejścia pipeline tłumaczenia wymagany jest `invalid`. Inne `unknown`, `incomplete`, `ok`, błąd infrastruktury i limit zasobów nie są poprawną odmową. Raport oddziela te granice od zwykłych rekordów i ich liczników PASS. Pochodzenie danych nie oznacza akceptacji właściciela (`NOT_EVALUATED`).

Dodatkowa kontrola `TestAmbiguityAssumptionsFollowRoles` tłumaczy rzeczywiste `Nauczyciel widzi psa.` PL→EN, wymaga czterech wariantów rodzajników i sprawdza, że każda założona wartość `determiner` ma ścieżkę do referenta roli `experiencer` albo `stimulus` zgodnego z niezależnym grafem Seeing.

## Odtworzenie zaobserwowanych przepływów TUI

Obserwacje czterech przepływów identyfikacji wykonano w rzeczywistym terminalu z binarium zbudowanym na kodzie A1-T04 i izolowaną konfiguracją użytkownika. Uruchomienie w TTY (ścieżka binarium może wskazywać własny build):

```sh
TERM=xterm-256color XDG_CONFIG_HOME=/tmp/golingua-a1-neg-config /tmp/golingua-a1-neg
```

W UI naciśnij `F5`, wpisz `/tmp/golingua-a1-neg-runtime.json` i `Ctrl+S` (zaobserwowane: ustawienia zapisane, runtime `ok`, `ready:true`). Naciśnij `F1`, wpisz dokładne zdanie i Enter. Dla EN→PL sprawdź `This is a cat` oraz `This cat is black.`. Użyj `Alt+D`, by przełączyć kierunek PL→EN, i sprawdź `To jest kot.` oraz `Ten kot jest czarny.`. `Tab` przechodzi między kierunkiem, kandydatami i podglądem; `Up`/`Down` wybierają kandydata, Enter otwiera szczegóły, `Esc` wraca/resetuje fokus, a `Ctrl+U` czyści prompt. Podgląd zachowuje status `incomplete` lub `ambiguous`; w szczegółach widoczne są odpowiednio strata `deixis` lub wybór `reference`. Zakończ `Ctrl+C`.

Rzeczywista obserwacja potwierdziła cztery przepływy TUI identyfikacji i czysty exit po `Ctrl+C`; konfiguracja użytkownika nie została zmieniona. Notatka dowodowa: `/tmp/a1-t05-tui-observation.md` (zawiera też build metadata i sumę binarium).

Dla negacji użyto świeżego binarium `/tmp/golingua-a1-neg` z runtime `/tmp/golingua-a1-neg-runtime.json` oraz odizolowanego `XDG_CONFIG_HOME`. TUI pokazało `To nie jest kot.` ze statusem `incomplete` i stratą `deixis`, `Nauczyciel nie widzi psa.` ze statusem `ok`, trzy warianty EN dla `To nie jest kot.` ze statusem `ambiguous` i widocznym wyborem `reference=this`, oraz cztery warianty rodzajników dla `Nauczyciel nie widzi psa.` ze statusem `ambiguous` i założeniami `determiner`. Podgląd zachował status `ambiguous` i wybrany tekst. Wszystkie cztery przepływy zakończono `Ctrl+C` z exit 0; ustawienia użytkownika nie uległy zmianie. Notatka: `/tmp/a1-neg-tui-observation.md`; surowy transcript: `/tmp/a1-neg-tui.typescript`.

## Status i pochodzenie

Wyniki poniżej dokumentują wcześniejszy zamknięty wycinek identyfikacji i negacji; nie zmieniają aktualnej oceny całej macierzy. Aktualny audyt to epik [#337](https://github.com/clickbytes/golingua/issues/337), child [#344](https://github.com/clickbytes/golingua/issues/344) i draft [PR #351](https://github.com/clickbytes/golingua/pull/351). M2–M7 są scalone z zielonym main CI; M8 jest zaimplementowany, a końcowe CI PR/main i merge oczekują. Akceptacja językowa właściciela pozostaje `NOT_EVALUATED`; [rejestr etapów i wyniki](../docs/work/a1-matrix/README.md) oraz [końcowe podsumowanie M8](../docs/work/a1-matrix/344-results.md) oddzielają status implementacji od obserwacji.

Powiązane wcześniejsze zadania: epik [#311](https://github.com/clickbytes/golingua/issues/311), fundament [#312](https://github.com/clickbytes/golingua/issues/312), identyfikacja I [#314](https://github.com/clickbytes/golingua/issues/314), PL [#316](https://github.com/clickbytes/golingua/issues/316), EN [#318](https://github.com/clickbytes/golingua/issues/318), integracja [#320](https://github.com/clickbytes/golingua/issues/320), negacja [#322](https://github.com/clickbytes/golingua/issues/322). Dodatkowe informacje: [architektura](../docs/architecture/README.md), [CLI](../docs/cli.md), [przegląd ręczny](../docs/manual-review.md).

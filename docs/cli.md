# Bieżące CLI

CLI udostępnia wersję, pomoc, deklarowane capabilities, operacje na Meaning JSON, typowane `question`, komendę `translate` oraz odmianę `forms`. Samo istnienie komendy `translate` nie oznacza, że środowisko ma zainstalowane języki lub może wykonać translację.

## Działające polecenia

Uruchom z katalogu repozytorium:

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

`--version` zwraca wersję technicznego podglądu v0. `--help` pokazuje dostępne komendy. `capabilities --json` opisuje gotowość bez wybranego runtime'u, dlatego obecnie zwraca `status: "unsupported"`, puste tablice `languages` i `features` oraz `ready: false`; sama komenda kończy się kodem 0. `forms` i `catalog` mają własne komendy opisane poniżej.

## Interfejs terminalowy (G11-V1/V2)

Bezargumentowe `golingua` otwiera TUI, gdy stdin i stdout są TTY (poza `TERM=dumb`). Bez TTY pozostaje pomoc bez ANSI. Jawne komendy, `--help`, `--version`, JSON i kody wyjścia zachowują dotychczasowy kontrakt.

F1 tłumaczy przez wspólny `app.Service`: wpisz tekst, `Enter`; `Alt+D` zmienia EN/PL. F5 przyjmuje ścieżkę runtime config: `Enter` wybiera ją i sprawdza; `Ctrl+S` jawnie zapisuje bezwzględną ścieżkę w `os.UserConfigDir()/golingua/settings.json`. Zapis nigdy nie stosuje snapshotu. Ostrzeżenie o błędnych ustawieniach nie powoduje ich automatycznego nadpisania. Sprawdzenie runtime’u oznacza możliwość jego wczytania, nie obsługę każdego zdania.

`F1`–`F5` przełączają ekrany, `Tab`/`Shift+Tab` zmieniają fokus. Na pasku zakładek użyj strzałek. Fokus wyniku pozwala przewijać `Up`/`Down`/`PgUp`/`PgDown`; `Ctrl+U` czyści pole. `?` pokazuje pomoc, `Esc` wraca do promptu, `Ctrl+C`/`Ctrl+D` kończy i przywraca terminal. F2–F4 są placeholderami. Szczegóły w [dokumentacji TUI](tui/README.md).

F1 udostępnia od V2 kontrolki candidates i preview: strzałki wybierają kandydata, `Enter` pokazuje szczegóły, a `Enter` na preview jawnie wywołuje usługę z jego `PreviewID`. Podgląd zachowuje status pierwotnego wyniku. Brak tekstu pozostaje jawnym brakiem. `Ctrl+P` otwiera na każdym ekranie paletę pięciu komend nawigacyjnych (`/translate`, `/forms`, `/concepts`, `/languages`, `/settings`); `Esc` anuluje i zachowuje szkic. W F5 ukośnik w polu ścieżki jest częścią ścieżki.

## Translacja

Użyj jawnych flag, aktywnego snapshotu i wymaganej odpowiedzi JSON:

```sh
go run ./cmd/golingua snapshot --apply --store ./store --json
go run ./cmd/golingua translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config ./runtime.json --json
```

`runtime.json` wskazuje aktywny snapshot. W skompilowanym CLI nie trzeba podawać `language_package_dirs`: binarka kompiluje ograniczone pakiety PL/EN dla dokładnych szlaków `seeing`, `hearing`, `looking`, `reading` oraz `receiving`. `seeing`, `hearing`, `looking` i `reading` obsługują jawne formy teraźniejsze 3. osoby liczby pojedynczej i mnogiej; `seeing`, `hearing` i `looking` mają też ograniczoną negację teraźniejszą w obu liczbach (`does/do not` / `nie`), a ich warianty EN zachowują zadeklarowane `a`, `an` i `the`; `a/an` wymaga jawnej właściwości leksykalnej i liczby pojedynczej. `reading` używa angielskich `read/reads` i polskich `czyta/czytają` w teraźniejszości oraz angielskiego simple past `read` i jawnych polskich form przeszłych `czytać` w 3. osobie liczby pojedynczej i mnogiej. Angielskie mnogie `read` pozostaje niejednoznaczne między czasem teraźniejszym i przeszłym; konstrukcja zachowuje obie analizy. Negacja `reading` pozostaje poza zakresem. Przy realizacji PL→EN brak określoności zostaje widoczny jako jawne założenie rodzajnika; kandydat może pozostać `incomplete` albo wynik może być `ambiguous`, gdy kilka wariantów rodzajnika jest licencjonowanych. `receiving` obsługuje angielskie `got` oraz polskie 3. osobowe formy czasu przeszłego liczby pojedynczej i mnogiej, jeśli liczba i klasa rodzaju podmiotu są znane. `looking` zachowuje `look at` / `patrzeć na` i polski biernik. Bez `--runtime-config` odpowiedź pozostaje `unsupported`, ponieważ proces nie ma wybranego snapshotu.

Opcjonalny `--preview-id <id>` jawnie kopiuje tekst wskazanego kandydata do głównego pola `text`, gdy ogólny status wyniku to `ambiguous` lub `incomplete`. Kandydat musi mieć tekst oraz własny status `ok` albo `incomplete`; status wyniku i kandydata, Meaning oraz evidence pozostają bez zmian. Kandydat bez tekstu lub oznaczony `unsupported`, `unknown`, `invalid`, `error` albo `resource_limit` zwróci `invalid`. Bez `--preview-id` silnik nie wybiera żadnego niekompletnego tekstu automatycznie. Jeśli runtime jest niedostępny, żądanie kończy się `unsupported`, bo nie ma alternatyw do sprawdzenia. Flagi mogą wystąpić w dowolnej kolejności. `--from`, `--to`, `--text` i `--json` są wymagane; wartości tekstowe nie mogą być puste. Nieznane, powtórzone, błędne lub niekompletne argumenty zwracają JSON ze statusem `invalid` i kodem wyjścia 2. Nie podawaj argumentów pozycyjnych.

Wynik translacji ma pola `status`, `from`, `to`, opcjonalne `text`, opcjonalne `preview_id`, `candidates` oraz `diagnostics`. Każdy candidate może zawierać jawne `assumptions` z polem, wartością, powodem i ścieżką do węzła Meaning; dzięki temu `incomplete` ujawnia, jakie cechy wybrał planer. W wynikach `incomplete` i `ambiguous` pole `text` jest ustawiane tylko po jawnym wskazaniu tekstowego kandydata przez `--preview-id`. Niepoprawne żądanie zwraca te same pola z `status: "invalid"`; błędna składnia samych flag jest raportowana jako JSON `status`/`diagnostics`, bo nie powstało jeszcze żądanie do silnika. Obecne ograniczone szlaki obejmują `seeing`, `hearing` i `looking` w teraźniejszej 3. osobie liczby pojedynczej i mnogiej, `reading` w teraźniejszości i przeszłości w tych liczbach, negację `seeing`, `hearing` i `looking` w 3. osobie liczby pojedynczej i mnogiej oraz `receiving` w czasie przeszłym, 3. osobie liczby pojedynczej i mnogiej. Angielska analiza przyjmuje jawne `the`, `a` oraz brak rodzajnika przy rzeczownikach niepoliczalnych z bieżącego inwentarza. Przy realizacji EN z polskiego źródła, które nie koduje określoności, konstrukcje transitive i oblique zachowują warianty `a/the` przy podmiocie oraz obiekcie/celu dla zatwierdzonych sensów; na `reading` planer ujawnia jawne warianty rodzajnika; wynik pozostaje `ambiguous`, gdy kilka form jest licencjonowanych. Wariant `a` wymaga jawnej właściwości `indefinite_article=a` na wszystkich dostępnych leksemach referentu; nie wyciąga się wniosku o rodzajniku z pisowni. W `receiving` jawnie masowy obiekt zachowuje wariant bez rodzajnika, również z wariantami rodzajnika przy podmiocie. Każdy wybór rodzajnika ma osobne założenie, wynik zbiorczy pozostaje `ambiguous`, a tekst konkretnego wariantu można wybrać przez `--preview-id`. Polska realizacja używa jawnych form przypadków tylko dla wybranych rzeczowników. Negacja PL→EN zachowuje tekstowego kandydata (np. `The teachers do not hear the dog.`), ale status pozostaje `incomplete`, gdy źródło nie rozstrzyga angielskich rodzajników. Brak odpowiedniej formy lub konstrukcji pozostaje `unsupported` albo `unknown`; kilka dopuszczalnych interpretacji pozostaje `ambiguous`. Status `ok` oznacza wykonanie skonfigurowanego pipeline'u, nie ocenę językową.

Tekst wejściowy ma limit 64 KiB, a każdy tekst wyjściowy również 64 KiB; łączny tekst alternatyw ma limit 64 KiB. Silnik ogranicza rozmiar całego wyniku do 4 MiB oraz liczbę kandydatów do 256. Każdy candidate może zwrócić do 64 assumptions; powód jednego assumption ma limit 1024 bajtów. Identyfikatory języków mają maksymalnie 35 znaków, a identyfikator podglądu maksymalnie 256 bajtów. Przekroczenie limitu wejścia lub retencji alternatyw zwraca `resource_limit`, a nie obcięty tekst. Niepoprawny albo zbyt duży serializowany wynik zwraca `invalid`.

## Odmiana i analiza form

`forms` ładuje obecny ograniczony seed morfologiczny PL albo EN. Generacja wymaga identyfikatora leksemu i cech językowych; analiza przyjmuje formę powierzchniową:

```sh
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.kot.n --feature number=singular --feature case=nominative --feature gender=masculine_animal --json
go run ./cmd/golingua forms --language en --surface dogs --json
go run ./cmd/golingua forms --language en --lexeme lexeme:en.look.v --feature verb_form=finite --feature tense=present --feature mood=indicative --feature voice=active --feature person=third --feature verb_number=singular --json
go run ./cmd/golingua forms --language en --lexeme lexeme:en.get.v --feature verb_form=finite --feature tense=past --feature mood=indicative --feature voice=active --json
```

Wynik zawiera status oraz `forms` lub `readings`. Brak leksemu albo odczytu daje `unknown`. To nie jest translacja ani ocena jakości językowej. Ocena na słowach właściciela pozostaje `NOT RUN`.

W G07-03 `forms` obejmuje także 15 jawnie zadeklarowanych rzeczowników PL po 14 standardowych komórek case/number. Są to pełne listy głównych form zapisane jako principal parts, nie automatyczne reguły dla całego słownika. Przykład: `forms --language pl --lexeme lexeme:pl.pies.n --feature number=plural --feature case=genitive --feature gender=masculine_animal --json`. Test techniczny generatora/analizatora w CI nie jest akceptacją językową.

G10-01 rozszerza angielskie `forms` o finite osoby/liczby/czasu oraz imperative i participles. Analiza form podstawowych zwraca wszystkie zgodne odczyty osoby/liczby; angielski czas przeszły jest wspólny dla osób i liczb. Obsługiwane są tylko przypisane, jawne klasy odmiany i wpisane principal parts. Przykłady są obserwacją generatora/analizatora, nie oceną zdań ani akceptacją językową.

G10-02 rozszerza polskie `forms` o jawne osoby/liczby dla wybranych czasowników oraz zgodę rodzaju/liczby w przeszłym `dostać`. W profilu PL `dostać` jako czasownik dokonany realizuje przyszłość przez sześć jawnych komórek z `tense=future`; liczba mnoga czasu przeszłego rozróżnia `gender=masculine_personal` i `gender=non_masculine_personal`.

```sh
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.widziec.v --feature verb_form=finite --feature tense=present --feature aspect=imperfective --feature mood=indicative --feature voice=active --feature person=first --feature number=singular --json
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.dostac.v --feature verb_form=finite --feature tense=future --feature aspect=perfective --feature mood=indicative --feature voice=active --feature person=second --feature number=singular --json
go run ./cmd/golingua forms --language pl --surface dostaliśmy --json
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.kon.n --feature number=plural --feature case=instrumental --feature gender=masculine_animal --json
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.miasto.n --feature number=plural --feature case=genitive --feature gender=neuter --json
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.dorosly.adj --feature degree=superlative --feature number=plural --feature gender=non_masculine_personal --feature case=nominative --json
go run ./cmd/golingua forms --language pl --surface dorośli --json
```

G10-04 dodał pełne 14 komórek przypadek/liczba dla `kot`, `ptak`, `koń`,
`miasto`, `mieszkanie`, `rzeka`, `siostra` i `mewa`. Dla istniejącego
przymiotnika `dorosły` CLI generuje udokumentowane stopnie: równy, wyższy i
najwyższy. Wyższe stopnie nie mają zadeklarowanych form wołacza, więc żądanie
takiej komórki pozostaje niedostępne. `mewa` jest leksemem morfologicznym bez
sensu i bindingu CORE, więc jego formy są dostępne dla `forms`, ale nie tworzą
nowej ścieżki tłumaczenia semantycznego.

G10-03 dodaje ograniczone angielskie klasy liczby rzeczowników: regularne `+s`,
sybilant `+es`, spółgłoska + `y → ies`, jawne principal parts dla wybranych
nieregularności oraz pluralia synkretyczne `fish` i `sheep`. Klasa jest
przypisana jawnie do wybranych leksemów; rzeczowniki masowe `food`, `money` i
`water` pozostają bez formy mnogiej.

```sh
go run ./cmd/golingua forms --language en --lexeme lexeme:en.bus.n --feature noun_number=plural --json
go run ./cmd/golingua forms --language en --lexeme lexeme:en.city.n --feature noun_number=plural --json
go run ./cmd/golingua forms --language en --lexeme lexeme:en.child.n --feature noun_number=plural --json
go run ./cmd/golingua forms --language en --surface buses --json
go run ./cmd/golingua forms --language en --surface children --json
go run ./cmd/golingua forms --language en --surface fish --json
```

## Inwentarz katalogu

`catalog --summary --json` kompiluje obecne dane Go CORE, PL i EN i pokazuje
oddzielne liczby konceptów, leksemów, sensów oraz statusy bindingów.
`catalog --details --json` dodaje listę stabilnych ConceptID ze statusami PL/EN. `pending`
nie upoważnia do tłumaczenia. Liczba 1000 jest celem redakcyjnym, a raport nie
mierzy poprawności językowej.

`catalog --summary` raportuje liczbę pending construction candidates i
approved construction bindings oddzielnie dla PL i EN. Tylko approved binding
licencjonuje dokładną trasę sense/frame/family/variant; sam nie wystarcza do
tłumaczenia zdania bez zatwierdzonych powiązań leksykalnych i runtime.

## Snapshot katalogu Go

`snapshot` buduje walidowane pliki CORE i wspólnego leksykonu PL/EN z bieżących
deklaracji Go. Podgląd podaje skrót i liczności; aktywacja zapisuje te same dane
do wskazanego lokalnego magazynu i ustawia jego aktywny wskaźnik:

```sh
go run ./cmd/golingua snapshot --preview --store <katalog-magazynu> --json
go run ./cmd/golingua snapshot --apply --store <katalog-magazynu> --json
```

Obie komendy wymagają jawnej ścieżki i mogą utworzyć katalog magazynu. Snapshot
zachowuje `pending` i pochodzenie każdego bindingu. Nie zawiera walencji ani
konstrukcji; skompilowana binarka dostarcza je osobno i aktywuje tylko dokładne
trasy dozwolone przez dane snapshotu. Dla bieżącego CLI plik runtime może
wyglądać tak:

```json
{
  "schema_version": 1,
  "dataset_store": "./store",
  "dataset_id": "golingua-authored"
}
```

Do użycia CORE przez `question` skonfiguruj `dataset_store` na ten katalog i
`dataset_id` na `golingua-authored`.

## Pytania na faktach z żądania

Komenda `question` wykonuje typowane zapytanie na `MeaningGraph` i `FactSet` podanych w jednym żądaniu. Nie parsuje polskiego ani angielskiego tekstu. Opcjonalna flaga `--language pl|en` uruchamia odpowiedź przez zweryfikowany target runtime:

```sh
go run ./cmd/golingua question --runtime-config <runtime.json> --json < question-request.json
go run ./cmd/golingua question --runtime-config <runtime.json> --language pl --json < question-request.json
```

Runtime config wybiera aktywny CORE/dataset snapshot. Konfiguracja `question` oraz bieżące skompilowane CLI mogą pominąć `language_package_dirs`; aplikacje używające zewnętrznych katalogów pakietów nadal mogą je podać. JSON ze standardowego wejścia ma limit 4 MiB, schemat zamknięty i wersję `schema_version: 1`:

- `question`: `reason.QuestionInput` — typowany `meaning`, request-scoped `context`, `kind` (`yes_no` albo `wh`), jawne `mapping` relacji i ról, opcjonalne `gap_role` i `at`.
- `facts`: tablica `facts.Fact` dla tego samego request context; każdy fakt wymaga jawnego źródła i polaryzacji.
- `rules`: opcjonalne reguły `symmetry`/`transitivity`, które CORE jawnie deklaruje. Są walidowane względem wybranego snapshotu.

Wynik JSON zawiera `status`, wybrany `goal`, zaplanowane `question`, `answer` z prawdą/kandydatami i proof IDs, oraz tylko te fakty/źródła, które stanowią dowód. Przy `--language` wynik zawiera też `response` z `response_meaning`, konstrukcjami i alternatywami. Każda alternatywa pokazuje tekst, status, dokładne identyfikatory dowodu konstrukcji, jawne `assumptions`, straty i diagnostykę. Brak dowodu daje `unknown`; nie tworzy faktu negatywnego. Niepoprawny snapshot, Meaning, FactSet lub mapowanie daje typowany status błędu.

Obecne dodatnie odpowiedzi yes/no mogą użyć istniejących zatwierdzonych konstrukcji zdaniowych PL/EN. Gdy Meaning nie określa czasu, osoby/liczby lub angielskiego rodzajnika, CLI zachowuje tekst ze statusem `incomplete` i jawnymi założeniami. Odpowiedzi `false` mogą użyć ograniczonej negacji teraźniejszej `seeing`, `hearing` lub `looking` w 3. osobie liczby pojedynczej albo mnogiej, jeśli liczba wynika z requestu; inne negacje pozostają `unsupported`. Odpowiedź WH zachowuje znaleziony referent i proof, ale jej tekst pozostaje `unsupported`, dopóki runtime nie ma samodzielnej konstrukcji NP. Błędy realizacji zachowują ustrukturyzowaną odpowiedź, dowód i źródło faktu. `question` nadal nie przyjmuje pytania w języku naturalnym. Ocenę delegowaną przez właściciela zapisuje się oddzielnie od jego akceptacji; ręczna akceptacja właściciela pozostaje osobnym statusem.

Wyjście sukcesu i błędów domenowych jest pojedynczym obiektem JSON na stdout. Exit code odpowiada statusowi: `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; `error` i błędy infrastruktury używają kodu 1. Komenda `capabilities --json` sama kończy się 0 i opisuje gotowość przez pole `ready`.

Nieznana lub niepoprawna komenda zwraca `invalid` i exit code 2. Ogólne błędy procesu/serializacji mają osobne kody; `unsupported`, `invalid` i awaria programu nie są jedną kategorią.

## Bieżąca obserwacja funkcji językowych

Po G10-18 uruchomiono wszystkie 100 istniejących wierszy korpusu w obu kierunkach: 200 wywołań CLI. EN→PL nadal ma 60 `ok` i 40 `incomplete`, PL→EN nadal ma 100 `ambiguous`; statusy i zbiory kandydatów tekstowych zgadzają się z [obserwacją G10-14](../data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv). Szczegóły nowych negatywnych alternatyw rodzajnika, prób liczby mnogiej i błędnego allomorfu oraz podsumowanie regresji są w [obserwacji G10-18](../data/sentences/observed-2026-10-23-g10-18-negative-article-alternatives.tsv). Niegramatyczne `A teachers...` i błędne `an dog` pozostały `unsupported`; polskie wejścia zachowały wiele jawnie założonych wariantów EN ze statusem `ambiguous`. G10-17 pozostaje w [osobnej obserwacji](../data/sentences/observed-2026-10-23-g10-17-present-negation.tsv). Przykłady dodatnie nie wybrały rodzin negatywnych; niezgodna liczba i nieobsługiwany czas przeszły pozostały poza nowymi trasami. G10-16 wcześniej dodało ograniczoną negację `seeing` w liczbie pojedynczej; historyczne wyjścia są w [obserwacji G10-16](../data/sentences/observed-2026-10-23-g10-16-negated-seeing.tsv). G10-13 dodało obsługę czasu teraźniejszego 3pl na ograniczonych trasach `seeing`, `hearing` i `looking` oraz twarde uzgodnienie `number`; szczegóły są w [obserwacji G10-13](../data/sentences/observed-2026-10-23-g10-13-plural-agreement.tsv). Pozostałe sensy, konstrukcje, czasy i formy wymagają własnych etapów danych i implementacji. Aktualne obserwacje dla 25 istniejących przykładów `looking` w obu kierunkach zapisano w [`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../data/sentences/observed-2026-10-23-g05-14-looking.tsv), a dla `receiving` w [`data/sentences/observed-2026-10-23-g05-13-receiving.tsv`](../data/sentences/observed-2026-10-23-g05-13-receiving.tsv). To zapis działania CLI, nie gold ani automatyczna ocena jakości. Właścicielski odbiór językowy pozostaje `NOT RUN`; delegowana ocena agenta jest odrębna. Procedura i statusy są opisane w [manual review](manual-review.md).

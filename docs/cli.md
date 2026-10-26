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

## Typowany IR: `algebra check`

`algebra check` dekoduje jeden version-1 dokument algebry, sprawdza typy i zakres binderów, a następnie zwraca strukturalny wynik oraz `canonical_document`. Może czytać jeden jawny plik albo stdin:

```sh
go run ./cmd/golingua algebra check --input docs/work/dse/evidence/03-01/inputs/atom.json --json
go run ./cmd/golingua algebra check --json < docs/work/dse/evidence/03-01/inputs/atom.json
go run ./cmd/golingua algebra check --input - --json < docs/work/dse/evidence/03-01/inputs/hole.json
```

Obie flagi można podać w dowolnej kolejności. `--json` jest wymagane. Bez `--input`, albo przy `--input -`, wejściem jest stdin; przy ścieżce pliku komenda czyta wyłącznie ten plik i nie dołącza stdin. Nieznane, brakujące, puste lub powtórzone argumenty dają JSON `status: "invalid"` i exit 2. Błąd otwarcia/odczytu pliku lub zapisu wyniku jest błędem infrastruktury wypisanym na stderr z exit 1.

Sukces zwraca `status`, `checked` (typ, wolne parametry, dziury/wymagania i licznik węzłów) oraz znormalizowany `canonical_document`. `status: "incomplete"` z exit 7 zachowuje typowane `Hole`, wymagania i diagnostykę; nie zamienia dziury w świadek lub fakt. Typowane odmowy mapują się na kody: `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7, a `error` 1. Zbyt długa ścieżka diagnostyczna lub wiadomość jest jawnie skracana na granicy UTF-8 do limitów wspólnego statusu (512/2048 bajtów); status i kod pozostają zachowane, a skrócenie ścieżki ma osobną adnotację.

Dokument ma limit 1 MiB; parser ogranicza zagnieżdżenie, a checker stosuje limity węzłów, głębokości, deklaracji, typów i portów. Szczegółowy [kontrakt IR](architecture/semantic-algebra/ir.md), [schemat JSON](../schemas/semantic/expression.schema.json), pliki kontrolne i status evidence są w [DSE-03-01](work/dse/evidence/03-01/README.md). Ta komenda nie redukuje `Apply`, nie wykonuje syntezy, nie ładuje rejestru runtime, nie ewaluje faktów i nie dowodzi prawdziwości propozycji. Poprawny typ oznacza wyłącznie poprawną strukturę według aktywnego checkera.

### Jawne pakiety metadanych

`algebra metadata inspect` i `algebra metadata validate` przyjmują jeden lub więcej jawnie wskazanych plików pakietu. `algebra check` może wskazać te same pliki, aby uzupełnić deklaracje z wybranego rejestru:

```sh
go run ./cmd/golingua algebra metadata inspect --metadata data/semantic/examples/metadata/neutral.json --json
go run ./cmd/golingua algebra metadata validate --metadata data/semantic/examples/metadata/neutral.json --json
go run ./cmd/golingua algebra check --input data/semantic/examples/metadata/expression.json --metadata data/semantic/examples/metadata/neutral.json --json
```

Powtórz `--metadata <plik>` dla kolejnych pakietów, w tym wszystkich jawnych zależności. Wczytanie obejmuje maksymalnie 16 plików, każdy do 1 MiB. Nie podawaj `--metadata` w danych pakietu jako ścieżki; CLI otwiera wyłącznie pliki wybrane flagą. Każdy pakiet przechodzi ścisły dekoder, walidację profilu, pochodzenia, deklaracji, relacji zależności, ograniczeń i definicji aliasów, zanim zostanie utworzony niezmienny snapshot. Błąd któregokolwiek pakietu zwraca odmowę bez częściowej listy rejestrów.

`inspect` pokazuje identyfikatory i wersje pakietów, oba fingerprinty, sorty, klasy symboli, status review i provenance. `validate` raportuje wyłącznie walidację struktury profilu `positive-facts-v1`; nie potwierdza spójności teorii, prawdziwości faktów ani akceptacji właściciela. `SourceRef` jest informacją o pochodzeniu, nie potwierdzeniem uprawnień ani semantyki.

Przy `algebra check --metadata ...` wejściowy dokument wyrażenia nie może zawierać własnych `sorts` ani `signatures`. Checker udostępnia wyłącznie zatwierdzone symbole `primitive`. Użycie symbolu `pending`, `opaque_unformalized` lub aliasu zwraca `unsupported`; alias można obejrzeć i walidator sprawdza jego zamkniętą definicję, lecz bieżący profil go nie rozwija podczas sprawdzania wyrażeń. Nieznany identyfikator symbolu pozostaje `invalid`. Wyrażenie jest sprawdzane przez ten sam typowany checker i wynik zawiera kanoniczny dokument z deklaracjami z rejestru oraz ID, wersję i fingerprint zasobów wybranych pakietów.

Fingerprint zasobu obejmuje metadane dokumentu i jawne piny zależności. Formalny `formal_denotation_id` pomija etykiety, opisy, provenance, status review i krawędzie typowości; obejmuje identyfikatory i strukturę formalną, a dla zależności ich ID oraz formalne denotation ID. Fingerprinty zasobów zależności i ich wersje nie trafiają do tej formalnej projekcji. Żadna klasa lub etykieta nie wybiera procedury. Początkowy allowlist aksjomatów jest pusty; `typical_edges` pozostają opisem, a ograniczenia `arity` i `port_type` są wyłącznie strukturalnymi warunkami deklaracji. Szczegóły: [kontrakt metadanych](architecture/semantic-algebra/metadata.md), [schemat pakietu](../schemas/semantic/package.schema.json), [schemat symboli](../schemas/semantic/symbols.schema.json) i [neutralne przykłady](../data/semantic/examples/metadata/neutral.json).

## Bounded normalization and certificate verification

`algebra normalize` applies the explicitly listed finite rewrite profile to a typed expression using one or more explicitly selected metadata packages. `algebra verify-rewrite` independently checks the submitted certificate against the original request and those packages:

```sh
go run ./cmd/golingua algebra normalize --metadata data/semantic/examples/normalization/metadata.json --input data/semantic/examples/normalization/ab.json --json > /tmp/golingua-normalized.json
python3 -c 'import json; r=json.load(open("data/semantic/examples/normalization/ab.json")); n=json.load(open("/tmp/golingua-normalized.json")); json.dump({"schema_version":1,"request":r,"certificate":n["result"]["certificate"]},open("/tmp/golingua-verification.json","w"))'
go run ./cmd/golingua algebra verify-rewrite --metadata data/semantic/examples/normalization/metadata.json --input /tmp/golingua-verification.json --json
```

Both commands require `--json` and at least one `--metadata <file>`; repeat that flag for each selected package and every explicit dependency. Input is stdin by default or one `--input <file>`; `--input -` selects stdin. The request carries the original expression document, rewrite profile, and optional bounded budget. The certificate cannot supply trusted aliases, registry files, or its own authoritative environment: the CLI rebuilds the environment and formal snapshot from the selected validated metadata.

Normalization returns its complete `NormalizationResult`, including the last checked expression, canonical document, certificate, statistics, and diagnostics. If a limit is reached after valid steps, it preserves that prefix, returns `resource_limit` (exit 4), and sets `normalized` false. The verifier returns separate `certificate_valid` and `normalization_complete` fields. A checked prefix is not a complete normalization; a complete normalization is only normal form under the selected finite profile. Neither command evaluates facts, proves truth, establishes unrestricted logical equivalence, or synthesizes a term from a goal. An empty rewrite list permits no simplification and therefore treats unchanged input as normal under that empty profile.

The six rule IDs are `alpha_v1`, `beta_v1`, `alias_v1`, `and_flatten_v1`, `and_order_v1`, and `and_deduplicate_v1`. Beta and alias reductions that would erase free parameter or reference dependencies are refused as unsupported. Conjunction ordering is a deterministic structural convention; deduplication removes only exact, identity-sensitive, binder-free duplicate operands. It does not merge operands containing binders, alpha-equivalent `Exists` terms, distinct event instances, or distinct reference IDs. The 262,144 logical work units count environment declarations and ports plus alias AST/interface nodes, initial root admission, whole-tree scans, and each applied step's full before/after roots. Candidate attempts also charge the environment inventory; verifier completion inspections charge their inspected roots. These are not CPU-time or energy measurements. Other limits include 128 steps, 64 alias expansions, 32,768 trace nodes, 1 MiB request/document/term, 8 MiB certificate, and 16 MiB verification envelope. See the [normalization contract](architecture/semantic-algebra/normalization.md), [proof schema](../schemas/semantic/proof.schema.json), and [neutral inputs](../data/semantic/examples/normalization/).

These examples are neutral mathematical controls from issue #382, not language data or owner-approved meanings. The coordinator owns the scoped mathematical assessment, supported by independent Luna review; owner language acceptance remains `NOT_EVALUATED`. Runtime evidence and its actual commands/exit codes are recorded separately from these docs.

### Jawne składanie: `algebra build`

`algebra build` wykonuje version-1 recepturę jawnych kroków konstrukcji na wskazanych pakietach metadanych. Każdy krok wskazuje wcześniejsze ID kroku; nie ma tu automatycznego wyszukiwania planu ani wnioskowania z etykiet.

```sh
go run ./cmd/golingua algebra metadata validate --metadata data/semantic/examples/composition/metadata.json --json
go run ./cmd/golingua algebra build --metadata data/semantic/examples/composition/metadata.json --input data/semantic/examples/composition/two-hop.json --json
go run ./cmd/golingua algebra build --metadata data/semantic/examples/composition/metadata.json --input data/semantic/examples/composition/three-hop.json --json
go run ./cmd/golingua algebra build --metadata data/semantic/examples/composition/metadata.json --input data/semantic/examples/composition/nary.json --json
```

Wymagany jest co najmniej jeden `--metadata`; powtarzaj flagę dla wszystkich jawnych pakietów i ich zależności. `--input <plik>` czyta wyłącznie wskazany plik; bez flagi albo z `--input -` receptura pochodzi ze stdin. `--json` jest wymagane. Wyjście zawiera `status`, `result_id`, `result_kind`, term wyrażenia lub substytucję, interfejs portów, `checked`, `canonical_document` dla termu, uporządkowany `trace` oraz wybrane pakiety metadanych. Unifikacja ma `result_kind: "substitution"`; również pusty wynik jest jawnie serializowany jako `substitution: {}` i nie oznacza dowodu ani propozycji.

Obsługiwane kroki to `relation`, `value`, `atom`, `and`, `join`, `bind`, `apply`, `restrict`, `project`, `permute` i `unify`. `join` wymaga jawnych par portów tego samego typu; brak mapowania zwraca `incomplete`, a zderzające się nazwy portów wyjściowych wymagają jawnego `outputs`. `Bind` może wskazać dowolny port interfejsu. `Apply` redukuje znany lambda-term. `Project` wiąże egzystencjalnie pominięte porty, a `Permute` wymaga pełnej bijekcji. Kopiowanie i podstawianie zachowują odrębność Param/Ref/Hole/Var i świeże ID binderów; receptura oraz operatory nie zmieniają wejściowych danych.

Profil unifikacji jest ograniczony do zadeklarowanych Param jako zmiennych, Ref, Atom i uporządkowanego And. Typy, symbole, role argumentów i ID referencji muszą być zgodne; occurs check odrzuca podstawienie cykliczne. Hole pozostaje `incomplete`, a wyrażenia lambda, Apply, Exists, arrow-variable i zadania higher-order są `unsupported` w unifikatorze. Sam `build` może użyć wspieranych operacji `Apply`/`Exists` do konstrukcji termu; odmowa unifikacji nie rozstrzyga prawdziwości termów.

`build` używa wyłącznie zatwierdzonych primitive signatures z jawnego rejestru. Pending, aliasy i symbole opaque nie są tutaj rozwijane ani uruchamiane; pole review nie oznacza akceptacji językowej. Wynik jest sprawdzoną reprezentacją, a nie nowym faktem, świadkiem świata, regułą przechodniości ani zarejestrowanym aliasem. Limit wejścia wynosi 1 MiB, maksymalnie 64 kroki, 4096 węzłów i głębokość 64 na wyrażenie, 32 na typ, 512 deklaracji, 16 portów oraz 16384 węzłów termów zachowanych łącznie w rezultatach i śladzie. Przekroczenie daje `resource_limit`, nigdy zgadywanie mapowania. Kody wyjścia pozostają mapowane do statusu: `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7, `error` 1. Pełny opis zachowania znajduje się w [kontrakcie składania](architecture/semantic-algebra/composition.md), [schemacie receptury](../schemas/semantic/construction.schema.json) i [przykładach](../data/semantic/examples/composition/).

### Przechowywanie sprawdzonych wyrażeń

`algebra export`, `algebra replay` i `algebra inspect` obsługują jeden version-1 artifact z jednym sprawdzonym korzeniem. Export i replay wymagają jawnie wybranych pakietów metadanych; inspect przyjmuje sam envelope i raportuje `metadata_verified: false`. Wejście może być plikiem albo stdin (`--input -`); `--no-cache` dotyczy wyłącznie exportu.

Limity wejściowe to 1 MiB dla żądania exportu, 8 MiB dla artifact envelope i 16 MiB dla żądania replay. Registry wybierane są osobnymi flagami `--metadata`; żaden path, URL ani kod z dokumentu nie jest automatycznie wczytywany. Export porównuje inline sorts/signatures z dokładnym typowanym podzbiorem wybranego registry, a następnie korzysta z inventory rejestru. Replay niezależnie sprawdza wybrane metadane, profil, context i provenance; inspekcja nie waliduje approval registry.

Wynik export/replay zawiera `status`, `artifact`, `checked`, `unique_nodes`, `expanded_nodes` i `metadata`; export dodatkowo `interning_enabled` i `cache_hit`. Inspect zawiera `validation_scope`, `metadata_verified`, `normalizer_version`, `registry_snapshot_id`, `required_registry_pins`, profil/context/provenance, trzy receipt fields, `root`, liczniki i `checked`. Hashes są receiptami dla konserwatywnych projekcji, nie dowodem równoważności logicznej ani autentyczności. Cache jest opcjonalny, request-local i ograniczony; nie jest trwałą pamięcią. Przechowywanie nie uruchamia rewrites, nie normalizuje wyrażeń, nie ewaluje faktów i niczego nie dowodzi. Szczegóły i neutralne wejścia są w [kontrakcie runtime storage](architecture/semantic-algebra/runtime-storage.md) oraz [katalogu przykładów](../data/semantic/examples/storage/).

## Interfejs terminalowy

Bezargumentowe `golingua` otwiera TUI, gdy stdin i stdout są TTY (poza `TERM=dumb`). Bez TTY pozostaje pomoc bez ANSI. Jawne komendy, `--help`, `--version`, JSON i kody wyjścia zachowują dotychczasowy kontrakt.

F1 tłumaczy przez wspólny `app.Service`: wpisz tekst i naciśnij `Enter`; `Alt+D` zmienia EN/PL. Up/Down wybiera rzeczywiste kandydatury, Enter otwiera ich szczegóły, a kontrolka Preview ponawia jawne żądanie z PreviewID. F2 udostępnia Analyze i Generate. Analyze przyjmuje formę powierzchniową; Generate przyjmuje LexemeID i jawne pary `key=value`. Żądanie nie uzupełnia cech. W Simple Enter na wyniku pokazuje szczegóły wszystkich odczytań/form.

F3 filtruje katalog lokalnie, dosłownie i z rozróżnieniem wielkości liter wyłącznie po ConceptID; Enter wywołuje lookup wybranego rekordu. F4 pokazuje rzeczywisty inwentarz PL/EN i osobno walidację runtime'u. Wejście przez F-key, tabs lub paletę odświeża oba odczyty; Refresh ponawia je. Wczytanie runtime config nie oznacza obsługi każdego zdania, a błąd runtime'u nie czyści inwentarza.

F5 edytuje wersjonowane ustawienia. Runtime path, display level, default direction i snapshot store pozostają szkicem do udanego `Save` (`Ctrl+S`); zapis atomowy używa trybu 0600. `Cancel`/`Esc` odrzuca szkic. Walidacja runtime'u sprawdza szkic, ale go nie wybiera. Starsze pliki v1 bez `display_level`, `default_from` i `snapshot_store` dostają wartości domyślne w pamięci bez automatycznego przepisu. Simple pokazuje skrót, Normal listy/ID/cechy/założenia, Technical pełny JSON. Szczegóły kandydatów i odczytań pozostają dostępne.

Snapshot Preview wymaga jawnej ścieżki store; nie aktywuje danych, choć może utworzyć katalogi. Apply wymaga udanego bieżącego Preview i dostępnego z klawiatury dialogu z domyślnym Cancel. Przed aktywacją następuje ponowny preview i weryfikacja digestu; zmiana digestu wymaga ponownego potwierdzenia. Apply nie zapisuje ustawień ani nie wybiera runtime'u. Ctrl+C/Ctrl+D czeka na już trwający Apply.

F1–F5 przełączają ekrany, Tab/Shift+Tab zmienia fokus, Left/Right na pasku wybiera ekran. F3–F5 focus order i operacje, resize i terminal opisano w [dokumentacji TUI](tui/README.md). Ctrl+P otwiera paletę pięciu tras nawigacyjnych; początkowy slash w promptcie F1–F3 także ją otwiera, a ukośnik w ścieżce F5 jest literalny. Bez TTY CLI zachowuje pomoc bez ANSI.

### Krótki przegląd ręczny

Uruchom `go run ./cmd/golingua` w terminalu, sprawdź F1 na zwykłym zdaniu PL lub EN, następnie F2 Analyze na istniejącej formie, F3 na wpisanym ConceptID, F4 przez Refresh i F5 przez walidację wybranej ścieżki. W F5 można osobno wpisać ścieżkę store, uruchomić Preview i anulować dialog Apply; nie trzeba stosować snapshotu do przejrzenia pozostałych ekranów. Na istniejących przykładach CLI sprawdź pojedynczą komendę `forms --language en --surface dogs --json` oraz `catalog --summary --json`. Bez TTY sprawdź domyślne wejście przez `go run ./cmd/golingua </dev/null`; `--help` pozostaje jawną komendą CLI. Ten przegląd nie jest test runnerem ani nowym korpusem.

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

Obecne dodatnie odpowiedzi yes/no mogą użyć istniejących zatwierdzonych konstrukcji zdaniowych PL/EN. Gdy Meaning nie określa czasu, osoby/liczby lub angielskiego rodzajnika, CLI zachowuje tekst ze statusem `incomplete` i jawnymi założeniami. Odpowiedzi `false` mogą użyć ograniczonej negacji teraźniejszej `seeing`, `hearing` lub `looking` dla dopuszczonych pojedynczych referentów; inne negacje pozostają `unsupported`. Pytanie z jawnym `temporal_frame` daje `UNSUPPORTED_QUESTION_TEMPORAL_FRAME`, a argument z `multiplicity=many` daje `UNSUPPORTED_QUESTION_MULTIPLICITY`: reasoner nie mapuje jeszcze względnego czasu na fakty ani nie agreguje grup. Odpowiedź WH zachowuje znaleziony referent i proof, ale jej tekst pozostaje `unsupported`, dopóki runtime nie ma samodzielnej konstrukcji NP. Błędy realizacji zachowują ustrukturyzowaną odpowiedź, dowód i źródło faktu. `question` nadal nie przyjmuje pytania w języku naturalnym. Ocenę delegowaną przez właściciela zapisuje się oddzielnie od jego akceptacji; ręczna akceptacja właściciela pozostaje osobnym statusem.

Wyjście sukcesu i błędów domenowych jest pojedynczym obiektem JSON na stdout. Exit code odpowiada statusowi: `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; `error` i błędy infrastruktury używają kodu 1. Komenda `capabilities --json` sama kończy się 0 i opisuje gotowość przez pole `ready`.

Nieznana lub niepoprawna komenda zwraca `invalid` i exit code 2. Ogólne błędy procesu/serializacji mają osobne kody; `unsupported`, `invalid` i awaria programu nie są jedną kategorią.

## Bieżąca obserwacja funkcji językowych

Po G10-18 uruchomiono wszystkie 100 istniejących wierszy korpusu w obu kierunkach: 200 wywołań CLI. EN→PL nadal ma 60 `ok` i 40 `incomplete`, PL→EN nadal ma 100 `ambiguous`; statusy i zbiory kandydatów tekstowych zgadzają się z [obserwacją G10-14](../data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv). Szczegóły nowych negatywnych alternatyw rodzajnika, prób liczby mnogiej i błędnego allomorfu oraz podsumowanie regresji są w [obserwacji G10-18](../data/sentences/observed-2026-10-23-g10-18-negative-article-alternatives.tsv). Niegramatyczne `A teachers...` i błędne `an dog` pozostały `unsupported`; polskie wejścia zachowały wiele jawnie założonych wariantów EN ze statusem `ambiguous`. G10-17 pozostaje w [osobnej obserwacji](../data/sentences/observed-2026-10-23-g10-17-present-negation.tsv). Przykłady dodatnie nie wybrały rodzin negatywnych; niezgodna liczba i nieobsługiwany czas przeszły pozostały poza nowymi trasami. G10-16 wcześniej dodało ograniczoną negację `seeing` w liczbie pojedynczej; historyczne wyjścia są w [obserwacji G10-16](../data/sentences/observed-2026-10-23-g10-16-negated-seeing.tsv). G10-13 dodało obsługę czasu teraźniejszego 3pl na ograniczonych trasach `seeing`, `hearing` i `looking` oraz twarde uzgodnienie `number`; szczegóły są w [obserwacji G10-13](../data/sentences/observed-2026-10-23-g10-13-plural-agreement.tsv). Pozostałe sensy, konstrukcje, czasy i formy wymagają własnych etapów danych i implementacji. Aktualne obserwacje dla 25 istniejących przykładów `looking` w obu kierunkach zapisano w [`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../data/sentences/observed-2026-10-23-g05-14-looking.tsv), a dla `receiving` w [`data/sentences/observed-2026-10-23-g05-13-receiving.tsv`](../data/sentences/observed-2026-10-23-g05-13-receiving.tsv). To zapis działania CLI, nie gold ani automatyczna ocena jakości. Właścicielski odbiór językowy pozostaje `NOT RUN`; delegowana ocena agenta jest odrębna. Procedura i statusy są opisane w [manual review](manual-review.md).

# Bieżące CLI

CLI udostępnia wersję, pomoc, deklarowane capabilities, operacje na Meaning JSON, typowane `question`, komendę `translate` oraz odmianę `forms`. Samo istnienie komendy `translate` nie oznacza, że środowisko ma zainstalowane języki lub może wykonać translację.

## Działające polecenia

Uruchom z katalogu repozytorium:

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

`--version` zwraca wersję technicznego podglądu v0. `--help` pokazuje dostępne komendy. `capabilities --json` bez runtime'u zwraca `status: "unsupported"` i `ready: false`. Z `--runtime-config <path>` ładuje dokładnie wskazany runtime albo zwraca typowaną odmowę; odpowiedź rozdziela `representation`, `analysis`, `realization` i `evaluation` dla faktycznie skonfigurowanych zasobów. `forms` i `catalog` mają własne komendy opisane poniżej.

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

### Jawne fakty n-arne: `algebra facts validate`

Polecenie sprawdza ograniczony, jawnie przekazany zestaw faktów względem wybranych pakietów metadanych i podsumowuje, które rekordy obowiązują w podanym kontekście czasu:

```sh
./golingua algebra facts validate --metadata data/semantic/examples/facts/metadata.json --input data/semantic/examples/facts/arities.json --json
./golingua algebra facts validate --metadata data/semantic/examples/facts/metadata.json --input data/semantic/examples/facts/conflict.json --json
cat data/semantic/examples/facts/empty.json | ./golingua algebra facts validate --metadata data/semantic/examples/facts/metadata.json --json
```

Wymagane jest `--json` oraz co najmniej jeden jawny `--metadata`; można powtarzać tę flagę dla wybranych pakietów i wszystkich ich zależności. `--input <plik>` wybiera pojedynczy plik, a brak `--input` lub `--input -` czyta stdin. Dane pakietu nie mogą wskazywać plików do otwarcia. Żądanie zawiera wersję schematu, `context_id`, tablice `referents` i `facts`, opcjonalny pin `snapshot_id`, opcjonalny UTC `at` i limity. Szczegółowy [kontrakt faktów](architecture/semantic-algebra/facts.md), [schemat JSON](../schemas/semantic/facts.schema.json) oraz [neutralne przykłady](../data/semantic/examples/facts/) opisują format.

Każda wartość faktu jest jawnym `Ref` do żądania. `Var`, `Param` i `Hole`, nieznane referenty, symbole niezatwierdzone do operacyjnego użycia oraz assertion scope inny niż `asserted` są odrzucane. Dopuszczone są zatwierdzone prymitywne relacje z wybranych metadanych; aliasy, symbole pending i opaque nie są faktami operacyjnymi. Argumenty są sprawdzane po nazwanych rolach i zwracane w kolejności portów relacji, bez sortowania wartości. Typ referentu nie scala jego tożsamości z innym referentem tego samego sortu.

Źródła dodatnie i ujemne mogą współistnieć dla tego samego atomu. Taki aktywny konflikt jest poprawnymi danymi i nie oznacza wybuchu ani prawdy/fałszu świata. Brak dopasowanego dowodu daje `unknown`, nigdy negację. Bez `at` aktywne są tylko rekordy bez obu granic czasu. Jawny `at` musi być UTC RFC3339Nano; granice przedziału są domknięte, puste końce nieograniczone, a żądanie nie używa zegara bieżącego ani określeń względnych. Fakty pozostają w wyniku także wtedy, gdy są nieaktywne w danym momencie.

Odpowiedź JSON ma `status`, `snapshot_id`, `context_id`, `at`, `referents`, kanonicznie uporządkowane `facts`, `summary`, `metadata`, `absence_means` i `validation_scope`. `summary` podaje liczby `stored_facts`, `active_facts` i `inactive_facts`, atomy z ich stanem dowodowym oraz aktywne konflikty. Poprawny konflikt kończy się `ok` i exit 0; błędne żądanie jest typowaną odmową, zwykle exit 2, a przekroczenie limitu używa `resource_limit` i exit 4. Maksymalny JSON to 1 MiB, zagnieżdżenie 256, do 4096 faktów, 512 referentów i 16 portów.

Ten request-scoped widok jest odrębny od istniejących binarnych `facts.Fact`/`Set` i komendy `question`. Nie importuje faktów z Meaning ani zdań, nie tworzy globalnej pamięci lub bazy i nie rozszerza starego reasonera. Obserwacja CLI, ocena agenta i akceptacja właściciela są osobne; akceptacja językowa pozostaje `NOT_EVALUATED`.

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

### Porównanie i niezależne sprawdzanie dowodów

`algebra compare` porównuje dwa dokumenty względem jawnego, skończonego
profilu przepisywania. `algebra verify-proof` ma oddzielne typowane ścieżki dla
roszczeń `equivalence`, `rewrite`, `goal` i `evaluation`; każda wymaga
odpowiednich wejść i niezależnie sprawdza dostarczony certyfikat lub proof.
Przykłady dla porównania i weryfikacji równoważności są w
[examples/proofs/equivalence](../data/semantic/examples/proofs/equivalence/), a
dla weryfikacji świadków ewaluacji w
[examples/proofs/evaluation](../data/semantic/examples/proofs/evaluation/).

```sh
golingua algebra compare \
  --left data/semantic/examples/proofs/equivalence/alpha-left.json \
  --right data/semantic/examples/proofs/equivalence/alpha-right.json \
  --profile data/semantic/examples/proofs/equivalence/profile-alpha.json \
  --context-id proof-neutral \
  --metadata data/semantic/examples/proofs/equivalence/metadata.json --json

golingua algebra verify-proof --claim equivalence \
  --left data/semantic/examples/proofs/equivalence/alpha-left.json \
  --right data/semantic/examples/proofs/equivalence/alpha-right.json \
  --profile data/semantic/examples/proofs/equivalence/profile-alpha.json \
  --context-id proof-neutral \
  --metadata data/semantic/examples/proofs/equivalence/metadata.json \
  --certificate data/semantic/examples/proofs/equivalence/certificate-alpha.json --json

golingua algebra verify-proof --claim evaluation \
  --expression data/semantic/examples/proofs/evaluation/evaluation-expression.json \
  --facts data/semantic/examples/proofs/evaluation/facts-branch-conflict.json \
  --select data/semantic/examples/proofs/evaluation/select-evaluation-original.json \
  --context-id eval-neutral --profile positive-facts-v1 \
  --metadata data/semantic/examples/proofs/evaluation/metadata.json \
  --proof data/semantic/examples/proofs/evaluation/evaluation-proof.json --json
```

Wymagane są `--json` oraz jawne pakiety `--metadata` (flaga może być
powtórzona). Porównanie przyjmuje dwa dokumenty, plik profilu i `context-id`;
weryfikacja równoważności dostaje także certyfikat z porównania. Weryfikacja
ewaluacji dostaje oryginalne wyrażenie, fakty, selekcję, `context-id`, profil
`positive-facts-v1` i proof. Limity porównania podaje się jako powtarzane
`--comparison-limit name=value`; wartości zero wybierają skończone domyślne
limity, a nie nielimitowane działanie. Limity ewaluacji mają osobną flagę
`--evaluation-limit name=value`.

Pozytywny wynik `equivalence` jest ograniczony do wybranego snapshotu metadanych,
profilu i kontekstu. Alpha, beta lub porządkowanie `And` są dozwolone tylko,
gdy dany profil jawnie licencjonuje regułę. Weryfikator niezależnie odtwarza
obie strony; różne postacie normalne, zamiana ról, inne ID referencji lub
zdarzeń oraz nieporządek `And` bez licencji pozostają `undetermined`, a nie
dowodem nierównoważności. Zakres nie obejmuje pełnego model checkera,
kontrmodeli ani entailmentu; te typy roszczeń są jawnie odrzucane jako
`unsupported`.

Weryfikacja `evaluation` sprawdza wyłącznie dostarczone dodatnie świadki
odpowiedzi względem wybranego wyrażenia, aktywnych faktów, ról, referencji,
czasu, kontekstu i profilu. Nie uruchamia ewaluatora ani nie szuka kolejnych
odpowiedzi. `completeness_verified` pozostaje `false`; brak świadków oznacza
`unknown`, nie fałsz. Dodatni i ujemny dowód źródłowy mogą współistnieć;
konflikt nie usuwa dodatniego wsparcia. Zasoby lub błędy weryfikacji nie są
zamieniane na negatywny dowód.

Limity są skończone. Porównanie domyślnie zużywa do 1 048 576 jednostek pracy
(twardy limit 16 777 216) i ogranicza certyfikat do 16 MiB. Ewaluacja domyślnie
zużywa do 1 048 576 jednostek pracy (twardy limit 16 777 216), ogranicza proof
do 4096 węzłów (twardo 16 384) oraz 8 MiB wyjścia (twardo 16 MiB) i ma domyślny
limit czasu 2 sekundy (maksymalnie 30 sekund). Wspólne limity typowanego terminu
również obowiązują. Przekroczenie limitu zachowuje status odmowy lub
`resource_limit` i nie dowodzi wyniku negatywnego. Przykłady oraz szczegółowy zakres są opisane w
[kontrakcie proofów](architecture/semantic-algebra/proofs.md) i
[neutralnym katalogu przykładów](../data/semantic/examples/proofs/README.md);
rzeczywiste CLI-obserwacje, wejścia i skróty plików znajdują się w
[#390 evidence](work/dse/evidence/05-03/README.md). Akceptacja językowa
właściciela pozostaje `NOT_EVALUATED`.

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

### Zakresy negacji, kwantyfikatorów i Ask

`algebra build` może jawnie zbudować termy `Not`, `Quantifier` i `Ask` z
wcześniejszych kroków receptury. Przykłady neutralne znajdują się w
[`data/semantic/examples/scopes/`](../data/semantic/examples/scopes/):

```sh
go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/build/not-all.json --json

go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/build/all-not.json --json

go run ./cmd/golingua algebra normalize \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/normalization/not-all.json --json

go run ./cmd/golingua algebra export \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/export/not-all.json --json

go run ./cmd/golingua algebra replay \
  --metadata data/semantic/examples/scopes/metadata.json \
  --input data/semantic/examples/scopes/replay/not-all.json --json

go run ./cmd/golingua algebra evaluate \
  --metadata data/semantic/examples/scopes/metadata.json \
  --expression data/semantic/examples/scopes/query/ask-first.json \
  --facts data/semantic/examples/scopes/facts-simple.json \
  --context-id eval-neutral --profile positive-facts-v1 --json

go run ./cmd/golingua algebra evaluate \
  --metadata data/semantic/examples/scopes/metadata.json \
  --expression data/semantic/examples/scopes/query/ask-second.json \
  --facts data/semantic/examples/scopes/facts-simple.json \
  --context-id eval-neutral --profile positive-facts-v1 --json
```

W przykładach `not-all` i `all-not` kolejność operatorów jest jawna i różna;
oba przepisywania oraz eksport/odtworzenie artefaktu są zapisane w evidence.
Przykłady Ask pytają o pierwszy albo drugi port tej samej relacji i tego samego
zestawu faktów. Wyniki pokazują odpowiednio `(A),(B)` oraz `(B),(C)`. `Ask`
przenosi projekcję w uporządkowanym polu `answers`, więc `evaluate` i
`verify-proof --claim evaluation` nie przyjmują zewnętrznego `--select`; jego
podanie jest `invalid`. `Ask` bez odpowiedzi jest pytaniem yes/no i dodatnie
wsparcie zwraca pustą krotkę. Brak dopasowania, także przy pustym zestawie
faktów, zwraca `knowledge: "unknown"`, nie odpowiedź negatywną.

Termy są opisane w [kontrakcie zakresów](architecture/semantic-algebra/scopes.md).
`Not` wymaga `Prop`. `some`/`all` nie przyjmują `count`; `exactly`, `at_least`
i `at_most` wymagają nieujemnej liczby, w tym zera. Pominięty lub `null`
restrictor oznacza brak restrictora; pusty obiekt `{}` jest nieprawidłowy.
`Not` i kwantyfikatory są typowane i składane, ale ich ewaluacja pozostaje
`unsupported`, również gdy znana jest liczba dopasowanych faktów. Nie dodano
profilu kompletnej dziedziny ani semantyki prawda/fałsz dla tych operatorów.

### Request-scoped referenty i warunki

Opcjonalne `request_context` w dokumencie przenosi jawne, typowane referenty,
kotwice mówiącego i adresata, wzmianki, rekordy `resolution_evidence` oraz
warunki `assumption`, `presupposition` i `loss`. To dane bieżącego żądania;
nie tworzą pamięci rozmowy ani faktów. Identyfikatory referentów pozostają
odrębne nawet przy tym samym typie i `concept_id`. Każda zadeklarowana
tożsamość wzmianki wymaga zgodnego dowodu; brak celu daje `unresolved`, a
sprzeczne jawne dowody dają `ambiguous` z widocznymi kandydatami.

Sprawdź samodzielny dokument z kontekstem:

```sh
go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/resolved-document.json --json

go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/unresolved-document.json --json

go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/conflicting-document.json --json
```

Komenda wymaga `request_context` i zwraca sprawdzenie struktury oraz osobne
`request_context_assessment`. `status: "ok"` oznacza poprawność typów i
deklaracji, nie rozstrzygnięcie każdej wzmianki ani prawdziwość warunków.
Przy wybranym `--metadata` dokument wejściowy nie może zawierać własnych
`sorts` ani `signatures`; dla dokumentów samodzielnych flagę można pominąć.
Format opisuje [kontrakt kontekstu](architecture/semantic-algebra/reference.md)
i [schemat](../schemas/semantic/context.schema.json).

Kontekst zachowuje się przez jawne budowanie, normalizację, syntezę,
certyfikację, eksport i odtworzenie. Na przykład:

```sh
go run ./cmd/golingua algebra build \
  --metadata data/semantic/examples/reference/metadata.json \
  --input data/semantic/examples/reference/build-request.json --json

go run ./cmd/golingua algebra normalize \
  --metadata data/semantic/examples/reference/metadata.json \
  --input data/semantic/examples/reference/normalization-request.json --json

go run ./cmd/golingua algebra export \
  --metadata data/semantic/examples/reference/metadata.json \
  --input data/semantic/examples/reference/export-request.json --json

go run ./cmd/golingua algebra replay \
  --metadata data/semantic/examples/reference/metadata.json \
  --input data/semantic/examples/reference/replay-request.json --json
```

Eksport bierze kontekst z dokumentu. Odtworzenie wymaga odrębnego
`request_context` w żądaniu i odrzuca brak lub inną treść, nawet jeśli ID się
zgadza. Podobnie porównanie oraz weryfikacja certyfikatu wymagają zgodnego
payloadu po obu stronach i w środowisku; nil-context zachowuje wcześniejszy
format i fingerprinty.

Cele mogą jawnie wiązać port z referencją, pozostawiając inny port otwarty:

```sh
reference_tmp="$(mktemp -d)"
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/reference/metadata-zero-arity.json \
  --goal data/semantic/examples/reference/description-goal.json --json \
  > "$reference_tmp/synthesis.json"

jq -c '.result.canonical_document' "$reference_tmp/synthesis.json" \
  > "$reference_tmp/candidate.json"
jq -c '.result.certificate' "$reference_tmp/synthesis.json" \
  > "$reference_tmp/certificate.json"

go run ./cmd/golingua algebra verify-goal \
  --goal data/semantic/examples/reference/description-goal.json \
  --candidate "$reference_tmp/candidate.json" \
  --certificate "$reference_tmp/certificate.json" \
  --metadata data/semantic/examples/reference/metadata-zero-arity.json --json
```

The resulting description is `λx.r(x, A)`: it keeps `x` open and uses only
the goal's explicit binding for `A`. The context also contains `B` with the
same type and concept classification, but this does not select or merge `B`.
`metadata-zero-arity.json` adds the declared `t0` primitive required by one
condition in this goal; the ordinary eight-symbol package omits it. Empty
`args: []` in that condition survives checking and serialization.

A presupposition is not promoted into a positive fact. This neutral example
checks the separate fact-evaluation input:

```sh
go run ./cmd/golingua algebra evaluate \
  --metadata data/semantic/examples/reference/metadata.json \
  --expression data/semantic/examples/reference/presupposition-is-not-fact-document.json \
  --facts data/semantic/examples/reference/facts.json \
  --select data/semantic/examples/reference/empty-selection.json \
  --context-id ctx-reference --profile positive-facts-v1 --json
```

The observed knowledge remains `unknown`; conditions are preserved metadata,
not evidence rows in the separate fact set. Context and condition size limits
are finite; exceeding them returns `resource_limit` rather than resolving an
identity or assuming a condition. See [#392 observations](work/dse/evidence/06-02/README.md)
for the literal commands, exit codes, output hashes, limits, and owner status.

W zakresie pytań evaluator obsługuje wyłącznie root `Ask` z dodatnim `Atom`/`And`/`Exists`
po beta-lowering. Wbudowany restrictor zostaje połączony z body. Zamknięte
pozytywne wyrażenie niebędące `Ask` może użyć jawnej pustej selekcji `[]`; brak
zewnętrznej selekcji nie zmienia go w `Ask`. Karta `algebra capabilities
--json` odróżnia możliwość reprezentacji od warunkowej ewaluacji i nie deklaruje
ogólnej obsługi typu `Question`. Dodatni proof potwierdza podane świadki, nie
kompletność odpowiedzi ani prawdziwość całego pytania. Nie dodano parsera ani
realizatora PL/EN dla tych scope'ów.

Rzeczywiste komendy, wejścia, wyjścia, skróty i kody wyjścia znajdują się w
[evidence DSE-06-01](work/dse/evidence/06-01/README.md). Odbiór techniczny
koordynatora i ocena językowa właściciela są osobnymi statusami; właścicielskie
`NOT_EVALUATED` nie jest zgodą semantyczną.

### Konstrukcja z jawnego celu

`algebra synthesize` obsługuje ograniczony tryb `construct`: buduje sprawdzone,
typowane wyrażenie dla w pełni określonego celu i jawnych pakietów metadanych.

```sh
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/goals/metadata.json \
  --goal data/semantic/examples/goals/two-hop-property.json --json
```

Wymagane są `--mode construct`, `--goal` i `--json`; `--goal -` czyta stdin.
Można powtarzać `--metadata` dla wybranych pakietów i ich jawnych zależności.
W tym trybie `--input` i `--no-cache` są odrzucane.
Dokument celu nie zawiera AST ani inline sortów/sygnatur. Wymaga między innymi
`context_id` (identyfikatora żądania, nie uwierzytelnienia ani tożsamości
referentu), typu oczekiwanego, uporządkowanych portów, jawnych referencji,
operatorów, ograniczeń, provenance i limitów; opcjonalne `project_ports`
wskazuje otwarte porty do egzystencjalnego zamknięcia. Dziura `Hole` pozostaje
nierozstrzygnięta i daje `incomplete`; nie jest świadkiem ani argumentem
operatora.

Profil `positive-facts-v1` obsługuje jeden jawny binarny constraint ścieżki oraz
jednoargumentowe ograniczenia unary na jej końcach. Obsługiwane operatory to
`relation`, `join`, `restrict`, `bind`, `project` i `permute`. Projekcja
zamyka egzystencjalnie tylko wymienione otwarte porty; nie może pokrywać
portów związanych przez `Ref` lub `Hole`. Brak wyboru liczby atomów
pozostaje unresolved, a wiele dozwolonych liczności pozostaje
underdetermined; komenda nie wybiera pierwszej. Nieznany symbol lub rola,
niezgodny typ i nieobsługiwany kształt grafu są odmowami, również gdy inne
elementy celu są niedookreślone. To nie jest ogólny planer grafów, wyszukiwarka,
ewaluator faktów ani dowód twierdzenia.

Wynik podaje `status`, `mode`, `goal_profile`, `context_id`, `metadata`,
`result` i zakres walidacji. Pełny rezultat zawiera sprawdzone wyrażenie,
interfejs, dokument kanoniczny, ślad operatorów i mapowanie ograniczeń na
odrębne ścieżki atomów końcowego wyrażenia. `constructed_provenance` wiąże
ograniczenie z krokami konstrukcji; nie jest niezależnym dowodem jego
prawdziwości. Konstrukcja dołącza certyfikat producenta ze statusem
`constructed_unverified` i `goal_verified: false`, a obok osobny wynik
niezależnej weryfikacji z komponentami kontroli. `fully_validated` jest polem
wyniku syntezy i ma wartość true tylko wtedy, gdy wymagane kontrole strukturalne
i certyfikatu zakończyły się powodzeniem; nie jest to ocena prawdy w świecie. Tylko w pełni zweryfikowany wynik zawiera `export_request` z pustym
profilem przepisywania `dse-construct-artifact-v1`; export nie normalizuje
wyrażenia ani nie ewaluuje faktów. Statusy używają kodów: `ok` 0, `invalid` 2,
`unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6,
`incomplete` 7, błąd infrastruktury 1.

Limity obejmują 16 atomów łącznie dla ścieżki i unary, 64 wywołania operatorów,
16 portów, 16 źródeł, 32 ograniczenia i 1 MiB wejściowego JSON celu. Wspólny
checker ogranicza wyrażenia do 4096 węzłów i głębokości 64, typy do głębokości
32, deklaracje do 512 i porty relacji do 16. Szczegółowe [granice konstrukcji](architecture/semantic-algebra/synthesis.md),
[schemat celu](../schemas/semantic/goals.schema.json) oraz [przykłady](../data/semantic/examples/goals/README.md)
opisują format. Walidacja strukturalna nie dowodzi prawdy świata; akceptacja
językowa właściciela pozostaje **NOT_EVALUATED**.

#### Wyszukiwanie w skończonym profilu

Tryb `search` enumeruje jawnie ograniczony zbiór konstrukcji dla tego samego
profilu celu. Nie wybiera jednej interpretacji ani nie sprawdza faktów świata.

```sh
go run ./cmd/golingua algebra synthesize --mode search \
  --metadata data/semantic/examples/search/metadata.json \
  --goal data/semantic/examples/search/ambiguous.json --json

go run ./cmd/golingua algebra synthesize --mode search \
  --metadata data/semantic/examples/search/metadata.json \
  --goal data/semantic/examples/search/ambiguous.json \
  --search-limit max_candidates=1 --json
```

Wymagane są `--mode search`, `--metadata`, `--goal` i `--json`; `--goal -`
czyta stdin. `--input` i `--no-cache` są odrzucane. Można powtarzać
`--search-limit name=value`, ale tylko w trybie search. Obsługiwane nazwy to
`max_path_length`, `max_branch_depth`, `max_expansions`, `max_candidates`,
`max_trace_entries`, `max_work_units`, `max_candidate_bytes` i
`timeout_millis`. Klucze muszą być unikalne, a wartości nieujemnymi
dziesiętnymi liczbami całkowitymi; nieznane, błędne i powtórzone klucze są
odrzucane. Zero wybiera ograniczoną wartość domyślną, a nie limit bez końca.

Wynik zawiera profil celu, `context_id`, metadane, pełne `result`, zakres
walidacji i wszystkie zachowane kandydatury. Każda zachowana kandydatura ma
osobny wynik niezależnej weryfikacji, zachowuje wybrane role, symbole, długość
ścieżki, referencje i mapowanie ograniczeń, a własne `export_request` dostaje
tylko po przejściu wymaganych kontroli. `complete: true` oznacza wyłącznie wyczerpanie
zadeklarowanej skończonej przestrzeni. Pusta kompletna lista z
`no_solution_within_profile_bounds` nie oznacza fałszu w świecie. Limit,
timeout albo nierozstrzygnięty cel dają wynik niekompletny; zachowana
kandydatura w prefiksie nie oznacza unikalności. Wynik nie ma rankingu ani
oceny pewności. Każda kandydatura ma szkicowy certyfikat
`draft_goal_unverified`; `goal_verified` pozostaje `false`. Weryfikacja
tożsamości alfa nie jest niezależnym dowodem spełnienia celu.

Po obejrzeniu `choices` wybierz jawnie indeks kandydatury. Dla zapisanego
wyniku wyszukiwania wyciągnij jej własne żądanie; pełny przepływ pokazuje
[kontrakt syntezy](architecture/semantic-algebra/synthesis.md).

```sh
selected_candidate=0
jq -c --argjson chosen "$selected_candidate" '.result.candidates[$chosen].export_request' /tmp/search.json > /tmp/search-export-request.json
```

`max_candidate_bytes` ma domyślnie 8 MiB i twardy limit 16 MiB. Liczy dokładne
zserializowane payloady zachowanych kandydatur, nie cały obiekt odpowiedzi;
`usage.candidate_bytes` jest ich sumą. CLI stosuje osobny limit 16 MiB na
całą odpowiedź JSON. Jeśli wymaga to usunięcia końcowych kandydatur, wynik
zostaje oznaczony `resource_limit`/`incomplete`, z `stop_reason` równym
`max_response_bytes` i liczbą odrzuconych elementów
`response_candidates_discarded`. Historyczna próba #387, sprzed zmiany
rozmiaru payloadów o pola certyfikatu celu, użyła 500 neutralnych,
zadeklarowanych referencji z ID długości 256 bajtów i wejścia 146 238 bajtów.
Zamrożone CLI zwróciło exit 4,
`resource_limit`/`incomplete` oraz 16 380 520 bajtów stdout z końcowym
newline, poniżej limitu 16 777 216 bajtów: 36 kandydatur pozostało, jedna
końcowa kandydatura została odrzucona, `complete: false`,
`unique_within_declared_space: false` i
`stop_reason: max_response_bytes`. `usage.candidate_bytes` równało się dokładnej
sumie 16 226 243 bajtów zachowanych payloadów. Zapis CLI jest poza repozytorium
w `/tmp/golingua-dse/387-wire-probe-real/output-500-256.json`, SHA-256
`df4c90cc02c7280a9d81939897f4ab495cfc5e85b33b6f3b0abdfaeca144c845`; skrócony
rekord polecenia i wyniku znajduje się w
`/tmp/golingua-dse/387-wire-probe-real/summary.json`. Hash zamrożonego binarium
przed i po próbie był taki sam
(`a561d324ee7b5047e82b6142672147a165832be0f2e647ed0d864d9170fe0604`). To
diagnostyka limitu bajtów, nie korpus językowy, dowód prawdy świata ani odbiór
właściciela. Limit pracy mierzy logiczne jednostki, nie CPU ani energię;
timeout jest sprawdzany kooperacyjnie na granicach faz i iteracji.

Wybór eksportu należy wykonać jawnie po obejrzeniu alternatyw. Przykłady
znajdują się w [katalogu search](../data/semantic/examples/search/README.md),
a pełne ograniczenia i statusy w [kontrakcie syntezy](architecture/semantic-algebra/synthesis.md).
Te formalne przykłady nie są korpusem językowym ani akceptacją semantyczną;
akceptacja językowa właściciela pozostaje **NOT_EVALUATED**.

#### Niezależna weryfikacja celu

`algebra verify-goal` sprawdza certyfikat konstrukcji niezależnie od
producenta. Podaj jawnie cel, kanoniczny dokument-kandydata, certyfikat i
wybrane pakiety metadanych:

```sh
go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/verification/metadata.json \
  --goal data/semantic/examples/verification/goal.json --json > /tmp/synthesis.json
jq -e '.result.fully_validated == true' /tmp/synthesis.json >/dev/null
jq -c '.result.canonical_document' /tmp/synthesis.json > /tmp/candidate.json
jq -c '.result.certificate' /tmp/synthesis.json > /tmp/certificate.json
go run ./cmd/golingua algebra verify-goal \
  --goal data/semantic/examples/verification/goal.json \
  --candidate /tmp/candidate.json --certificate /tmp/certificate.json \
  --metadata data/semantic/examples/verification/metadata.json --json > /tmp/verification.json
jq -e '.result.status == "ok" and .result.type_correct and .result.refinement_checked and .result.replay_checked and .result.certificate_valid and .result.goal_satisfied and .result.complete' /tmp/verification.json >/dev/null
jq -c '.export_request' /tmp/synthesis.json > /tmp/export-request.json
go run ./cmd/golingua algebra export \
  --metadata data/semantic/examples/verification/metadata.json \
  --input /tmp/export-request.json --json > /tmp/export.json
jq -c --slurpfile exported /tmp/export.json \
  '{schema_version:.schema_version,profile:.profile,context_id:.context_id,context_dependencies:.context_dependencies,provenance:.provenance,artifact:$exported[0].artifact}' \
  /tmp/export-request.json > /tmp/replay-request.json
go run ./cmd/golingua algebra replay \
  --metadata data/semantic/examples/verification/metadata.json \
  --input /tmp/replay-request.json --json
```

Wymagane są trzy nazwane pliki: `--goal`, `--candidate` i `--certificate`;
ta komenda nie czyta ich ze stdin. Można powtarzać `--metadata` dla jawnie
wybranych pakietów. Dokument kandydata musi być samodzielnym dokumentem algebry;
jego deklaracje inline są porównywane z wybranymi metadanymi. Certyfikat nie
wybiera rejestru i nie uwierzytelnia własnego snapshotu. Powtarzalne
`--verify-limit name=value` przyjmuje wyłącznie `max_certificate_bytes`,
`max_replay_steps` i `max_work_nodes`. Limity domyślne i maksymalne wynoszą
odpowiednio 8 MiB, 64 kroki replay i 262 144 logiczne jednostki pracy. Limit
certyfikatu jest normalizowany przed odczytem pliku. Nieznane, powtórzone,
niecałkowite, ujemne lub przekraczające limit wartości są odrzucane.

Wynik oddziela `type_correct`, `refinement_checked`, `replay_checked`,
`certificate_valid`, `goal_satisfied` i `complete`. Akceptacja wymaga
powodzenia wszystkich tych kontroli; flaga producenta `goal_verified` nie jest
zaufana. Pole `origin` także jest etykietą producenta i nie dowodzi, że użyto
wyszukiwania. Pojedynczy zweryfikowany kandydat nie dowodzi kompletności ani
unikalności przestrzeni wyszukiwania. `Hole`, nieobsługiwane twierdzenie lub
limit zasobów nie stają się zweryfikowanym kandydatem.

Weryfikacja dotyczy ograniczonego strukturalnego profilu pozytywnego. Nie
ewaluuje jawnych faktów, nie ustala prawdy świata ani spójności całej teorii i
nie jest akceptacją językową właściciela. Przykłady w
[katalogu weryfikacji](../data/semantic/examples/verification/README.md)
zawierają neutralny zaakceptowany dokument oraz kandydata z błędną orientacją,
który jest odrzucany.

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

Istniejący runtime bez pola `semantic_profile` zachowuje trasę Meaning v1. Jawne
`"semantic_profile": "dse"` wybiera osobny przepływ DSE; `dse_profile` wtedy
obowiązkowo przypina wersję profilu, fingerprint całego wybranego registry,
pakiety metadanych z ich ID/wersją/fingerprintem oraz dla `en` i `pl` pakiet
źródłowy i docelowy, plan analizy i kontekst referencji. Pakiety językowe i plany
przypinają SHA-256 dokładnych bajtów pliku; kontekst przypina `context_id`, wersję
schematu i SHA-256. Ścieżki względne są liczone względem pliku runtime. Błędny
pin lub brak wymaganej zdolności odrzuca aktywację. DSE nie wraca automatycznie
do Meaning v1.

W tej iteracji runtime DSE obsługuje wyłącznie jeden dodatni, bezpośredni Atom.
Kontekst i warunki, których target nie zużywa, pozostają w DSE trace/residualach;
nieobsłużony warunek kończy się jawnym `unsupported` przed wywołaniem generatora.
Warunek `lexical_descriptor` może zostać zużyty tylko dla dokładnego źródłowego
sensu z zatwierdzonym wiązaniem referencji i jednym zatwierdzonym, równoważnym
transferem do CORE; wynik zapisuje to w `consumed_conditions`. Każdy referent
uczestnika dostaje koncept wyłącznie z dokładnego źródłowego sensu powiązanego
z jednym zatwierdzonym transfer bindingiem CORE. Brak, konflikt lub kilka mapowań
kończy się jawną odmową przed generacją. Wynik DSE ma `dse_trace` z pełnym
składaniem, GoalSpec/certyfikatami, canonical document, pinami i pokryciem
realizatora; alternatywy nie zawierają konkurencyjnego grafu Meaning. Jawne
`--preview-id` może pokazać wybrany przez wywołującego wariant, ale zbiorczy
`ambiguous` oraz pozostałe warianty i ich założenia zostają zachowane.

Przykładowe sprawdzenie skonfigurowanego profilu:

```sh
golingua capabilities --runtime-config ./runtime-dse.json --json
golingua translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config ./runtime-dse.json --json
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

## Eksport katalogu do inwentarza DSE

`algebra catalog export --json` eksportuje skompilowany katalog z binarki jako ograniczony, nieaktywny inwentarz. Polecenie nie przyjmuje `--input`, `--metadata` ani `--no-cache` i zwraca JSON o polach `status`, `validation_scope`, `inventory`; sam inwentarz zawiera `schema_version`, `format`, `version`, `counts`, `source_records` i `packages`.

```sh
go run ./cmd/golingua algebra catalog export --json > /tmp/catalog-export.json
python3 -c 'import json; x=json.load(open("/tmp/catalog-export.json")); json.dump(x["inventory"]["packages"][0], open("/tmp/catalog-part-1.json","w"))'
go run ./cmd/golingua algebra metadata validate --metadata /tmp/catalog-part-1.json --json
go run ./cmd/golingua algebra metadata validate --metadata data/semantic/pilot/metadata.json --json
go run ./cmd/golingua algebra check --metadata data/semantic/pilot/metadata.json --input data/semantic/pilot/reading-check.json --json
go run ./cmd/golingua algebra build --metadata data/semantic/pilot/metadata.json --input data/semantic/pilot/reading-build.json --json
go run ./cmd/golingua algebra build --metadata data/semantic/pilot/metadata.json --input data/semantic/pilot/two-hop.json --json
```

`python3` w przykładzie jedynie wyciąga jeden obiekt JSON do pliku, aby przekazać go do istniejącego walidatora. Eksport ma 1002 rekordy źródłowe w deterministycznych chunkach 400/400/202. Każdy chunk jest niezależny; limit 512 deklaracji nie pozwala wybrać wszystkich trzech jednocześnie. Limity eksportu to 4096 rekordów, 400 wpisów/chunk, 16 chunków, 1 MiB/pakiet i 16 MiB łącznie. Błąd limitu lub walidacji zwraca odmowę bez częściowego inwentarza.

Źródłowe definicje zachowują ID, kind, parent, role unions, definition edges, semantic effects, description i provenance. Pakiety eksportu zachowują identyfikatory i opisy jako `opaque_unformalized`/`pending`; nie zamieniają source roles na formalne ports i nie tworzą rules, aliases, axioms ani facts. Eksport nie podmienia liczników `catalog --summary`/`--details`: docelowy rozmiar 1000 jest celem redakcyjnym, nie rzeczywistą liczbą rekordów ani primitive signatures.

Oddzielny pilot [`metadata.json`](../data/semantic/pilot/metadata.json) deklaruje `pilot.r`, `pilot.mark` i wersjonowane mapowanie `concept:reading`. Mapowanie jawnie dodaje port `event: Event` do istniejących ról `reader` i `material`; ten port nie występuje w oryginalnym CORE role list. [Mapping record](../data/semantic/pilot/mapping.json) zachowuje nowe źródło mapowania oddzielnie od oryginalnych CORE/PL/EN source refs. To ograniczona decyzja do strukturalnego użycia pilota, bez twierdzenia o równoważności całej definicji CORE ani o akceptacji PL/EN. Lexical anchors nie są automatyczną licencją nowej trasy tłumaczenia.

Na aktywnej gałęzi walidacja pilota, check `reading`, build relacji reading i neutralny build dwuetapowy zwróciły `ok`. Niepoprawny typ portu został odrzucony przez `ARGUMENT_TYPE_MISMATCH`, a `concept:seeing` jest `UNKNOWN_SYMBOL` w małym pilocie. Resource fingerprint zależy od provenance; wartości z obserwacji sprzed korekty źródeł nie są fingerprintem końcowego manifestu. Pełny ledger 1002 ID znajduje się w [`catalog-review.tsv`](../data/semantic/catalog-review.tsv), a szczegóły zakresu w [catalog triage](work/dse/catalog-triage.md). Akceptacja językowa właściciela pozostaje `NOT_EVALUATED`; PR i CI na `main` są odrębnymi, oczekującymi statusami.

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

### Ewaluacja jawnych faktów: `algebra evaluate`

`algebra evaluate` wykonuje sprawdzone zapytanie na jednym jawnie wybranym,
typowanym widoku faktów. Wszystkie wejścia są nazwanymi plikami; polecenie nie
czyta stdin:

```sh
golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/evaluation/metadata.json \
  --goal data/semantic/examples/evaluation/goal.json --json > /tmp/synthesis.json
jq -e '.result.fully_validated == true' /tmp/synthesis.json >/dev/null
jq -c '.result.canonical_document' /tmp/synthesis.json > /tmp/query.json
jq -c '.result.certificate' /tmp/synthesis.json > /tmp/certificate.json
golingua algebra verify-goal \
  --metadata data/semantic/examples/evaluation/metadata.json \
  --goal data/semantic/examples/evaluation/goal.json \
  --candidate /tmp/query.json --certificate /tmp/certificate.json --json
jq -c '[.result.certificate.interface | to_entries[] | {kind:"lambda",index:.key,as:.value.port}]' \
  /tmp/synthesis.json > /tmp/select.json
golingua algebra evaluate \
  --metadata data/semantic/examples/evaluation/metadata.json \
  --expression /tmp/query.json \
  --facts data/semantic/examples/evaluation/facts-simple.json \
  --select /tmp/select.json --context-id eval-neutral \
  --profile positive-facts-v1 --json
```

Wymagane są `--metadata`, `--expression`, `--facts`, `--select`, `--context-id`,
`--profile positive-facts-v1` i `--json`. Nie można podać `-` ani użyć stdin.
Plik selekcji to ścisła tablica JSON: lambda wybiera binder przez pozycję w
kolejności po ograniczonej redukcji beta, a parametr przez jego jawne ID.
`--context-id` musi zgadzać się z wejściem faktów. Powtarzalne
`--evaluation-limit name=value` obsługuje `max_plan_atoms`, `max_variables`,
`max_intermediate_rows`, `max_results`, `max_proof_nodes`,
`max_proofs_per_answer`, `max_proof_depth`, `max_work_units`,
`max_output_bytes` i `timeout_millis`; nazwy muszą być unikalne, a zero oznacza
ograniczoną wartość domyślną. `--at` jest opcjonalnym znacznikiem czasu UTC i
nigdy nie uruchamia zegara bieżącego.

`knowledge: supported` oznacza co najmniej jeden wynik wsparty jawnymi
pozytywnymi rekordami; bez takiego wsparcia wynik to `unknown`, nigdy `false`.
Negatywne rekordy i konflikty są osobno widoczne i nie anulują pozytywnego
wsparcia. `complete: true` oznacza wyczerpanie skończonego skanowania podanych
referentów i wybranego snapshotu faktów. Nie dowodzi kompletności domeny świata.
Limit lub anulowanie zwraca wyłącznie zatwierdzony prefiks odpowiedzi i dowodów,
z `complete: false`. Fakty i formalne metadane mają odrębne identyfikatory
snapshotu.

### Atrybuty referencji i Modal

Wpis referencji może nieść `temporal_frame` (`past` lub `present`) i
`reference_time_id` typu Interval tylko dla Event, a `multiplicity` (`one` lub
`many`) tylko dla Entity/Nominal. Atrybuty są opcjonalne: nieobecne pole
pozostaje nieznane. Deklaracje z dokumentu i jego jawnego kontekstu są łączone;
sprzeczne wartości, nieznany cel lub cel złego typu są odrzucane. Atrybuty nie
przechodzą na Param ani bindery Ask/lambda. `inspect` pokazuje zapisane pola,
ale `metadata_verified` pozostaje `false`; do ich weryfikacji względem
metadanych służą jawny export i replay.

Neutralne przykłady znajdują się w
[`data/semantic/examples/attributes/`](../data/semantic/examples/attributes/README.md).
Wywołania kontroli i ewaluacji korzystają z neutralnego pakietu oraz jawnych danych:

```sh
golingua algebra check --input data/semantic/examples/attributes/past.json --json
golingua algebra build --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/modal-build.json --json
golingua algebra evaluate --metadata data/semantic/examples/attributes/metadata.json --expression data/semantic/examples/attributes/one.json --facts data/semantic/examples/attributes/facts-with-event.json --select data/semantic/examples/attributes/select-empty.json --context-id eval-neutral --profile positive-facts-v1 --json
golingua algebra evaluate --metadata data/semantic/examples/attributes/metadata.json --expression data/semantic/examples/attributes/past.json --facts data/semantic/examples/attributes/facts-with-event.json --select data/semantic/examples/attributes/select-empty.json --context-id eval-neutral --profile positive-facts-v1 --json
golingua algebra compare --left data/semantic/examples/attributes/not-modal.json --right data/semantic/examples/attributes/modal-not.json --profile data/semantic/examples/attributes/compare-profile.json --context-id eval-neutral --metadata data/semantic/examples/attributes/metadata.json --json
```

Eksport zachowuje atrybuty jawnych referencji. `inspect` samodzielnie pokazuje
strukturę, a `replay` wymaga wybranego pakietu metadanych:

```sh
golingua algebra export --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/past-export-request.json --json
golingua algebra inspect --input data/semantic/examples/attributes/past-artifact-input.json --json
golingua algebra replay --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/past-replay-request.json --json
```

Te trzy polecenia zwróciły kod 0; inspekcja zwróciła
`metadata_verified: false`.

W obserwacji #393 `check`, `build`, ewaluacja referencji `one` i
`compare` kończyły się kodem 0; ewaluacja `past` zwracała `unsupported`
(kod 3), ponieważ profil ewaluacji nie obsługuje względnego czasu. Porównanie
zakresów modalnych zwracało `undetermined`, bez wnioskowania o równoważności.
`many` i Modal można reprezentować, typować, budować i zachowywać w artefakcie,
lecz ewaluator obecnie zwraca dla nich `unsupported`, a nie `false`. Puste,
`null`, sprzeczne lub źle typowane atrybuty są odrzucane. Szczegóły kontraktu
opisuje [atrybuty referencji i Modal](architecture/semantic-algebra/attributes.md);
pełne wejścia i wyjścia CLI są w [evidence #393](work/dse/evidence/06-03/README.md).
Odbiór językowy właściciela pozostaje `NOT_EVALUATED`.

DAG dowodowy zachowuje ID i źródła wszystkich pozytywnych faktów dla
pasującego atomu oraz ID przeciwnego dowodu, jeśli istnieje. Węzły `And`
zachowują kolejność źródłową; węzeł `Exists` przechowuje dokładne świadectwo
zakresowe, ale nie udostępnia go jako odpowiedzi; projekcja zawiera tylko jawnie
wybrane kolumny. Nieużywane bindery są ograniczone do jawnej unii typowanych
referentów z zapytania i wejścia faktów. To nie jest negacja przez brak,
rekurencyjny Datalog, globalny solver ani niezależny weryfikator dowodów #390.

Pełne ograniczenia i zapisany przykład znajdują się w [kontrakcie ewaluacji](architecture/semantic-algebra/execution.md) i [neutralnych danych](../data/semantic/examples/evaluation/README.md). Obserwacje branch CLI, ocena agenta i akceptacja właściciela są osobne; akceptacja językowa pozostaje **NOT_EVALUATED**.

### Wiązania leksykalne

Polecenia `algebra binding` operują na jawnym pakiecie językowym wersji 1. Każde polecenie wymaga `--language-package`; `validate` i `instantiate` wymagają też jednego lub większej liczby jawnych `--metadata`. `inspect` pokazuje zawartość pakietu bez walidacji jego celów wobec metadanych i zwraca `metadata_verified: false`. `validate` sprawdza podpisy i autoryzację odrębnie. `instantiate` przyjmuje `--input <plik>` albo `--input -` dla stdin.

```sh
golingua algebra binding inspect --language-package data/semantic/languages/pl/pilot.json --json
golingua algebra binding validate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --json
golingua algebra binding instantiate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/pl-see-request.json --json
```

Wynik `instantiate` dla celu `symbol` umieszcza dokument w `.result.symbol.document`. Można przekazać go do zwykłego kontrolera wyrażeń:

```sh
TASK_TMP=$(mktemp -d)
golingua algebra binding instantiate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/pl-see-request.json --json > "$TASK_TMP/instantiation.json"
jq '.result.symbol.document' "$TASK_TMP/instantiation.json" > "$TASK_TMP/document.json"
golingua algebra check --input "$TASK_TMP/document.json" --json
```

Dla celu `goal` wynik zawiera `.result.goal`, `.result.construction.canonical_document` i `.result.certificate`. Każdy element można sprawdzić oddzielnie; kontrola celu ponownie wymaga wybranych metadanych:

```sh
TASK_TMP=$(mktemp -d)
golingua algebra binding instantiate --language-package data/semantic/languages/en/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/en-link-indirect-request.json --json > "$TASK_TMP/instantiation.json"
jq '.result.goal' "$TASK_TMP/instantiation.json" > "$TASK_TMP/goal.json"
jq '.result.construction.canonical_document' "$TASK_TMP/instantiation.json" > "$TASK_TMP/candidate.json"
jq '.result.certificate' "$TASK_TMP/instantiation.json" > "$TASK_TMP/certificate.json"
golingua algebra check --input "$TASK_TMP/candidate.json" --json
golingua algebra verify-goal --metadata data/semantic/examples/lexical-bindings/metadata.json --goal "$TASK_TMP/goal.json" --candidate "$TASK_TMP/candidate.json" --certificate "$TASK_TMP/certificate.json" --json
```


Kilka pasujących wiązań daje `ambiguous`; wybór wymaga jawnego `binding_index`. Brak referencji, nieznana rola/port, zły typ albo kontekst przypięty do innego żądania nie są naprawiane przez zgadywanie. Szczegóły rozdziału między ramą, rolą i portem, granic kontekstu oraz ograniczeń pilota opisuje [kontrakt wiązań leksykalnych](architecture/semantic-algebra/lexical-bindings.md). Próbki i obserwacje znajdują się w [przykładach](../data/semantic/examples/lexical-bindings/README.md) oraz [evidence #394](work/dse/evidence/07-01/README.md). Akceptacja językowa właściciela pozostaje **NOT EVALUATED**.

### Analiza źródłowa: `algebra analyze`

To jawne wejście do ograniczonego pipeline'u składniowego. Wymaga aktywnego snapshotu runtime, pakietu i planu dla tego samego języka, kontekstu oraz wybranego zestawu metadanych. Polecenie nie zmienia domyślnego `translate` i nie przechodzi do starej kompozycji CORE.

```sh
TASK_TMP=$(mktemp -d)
repo=$(pwd)
printf '{"schema_version":1,"dataset_store":"%s/store","dataset_id":"golingua-authored","language_package_dirs":[]}\n' "$TASK_TMP" > "$TASK_TMP/runtime.json"
golingua snapshot --apply --store "$TASK_TMP/store" --json

golingua algebra analyze --language en --text 'The teacher sees the dog.' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package "$repo/data/semantic/languages/en/source.json" \
  --source-plan "$repo/data/semantic/languages/en/source-plan.json" \
  --context "$repo/data/semantic/examples/source-composition/contexts/main.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/en.json"
jq '.composition.candidates[0].canonical_document' "$TASK_TMP/en.json" > "$TASK_TMP/en-document.json"
golingua algebra check --input "$TASK_TMP/en-document.json" --json

golingua algebra analyze --language pl --text 'Nauczyciel widzi psa.' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package "$repo/data/semantic/languages/pl/source.json" \
  --source-plan "$repo/data/semantic/languages/pl/source-plan.json" \
  --context "$repo/data/semantic/examples/source-composition/contexts/main.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/pl.json"
jq '.composition.candidates[0].canonical_document' "$TASK_TMP/pl.json" > "$TASK_TMP/pl-document.json"
printf '%s\n' '{"id":"source-profile-empty","rules":[]}' > "$TASK_TMP/profile.json"
golingua algebra compare --left "$TASK_TMP/en-document.json" --right "$TASK_TMP/pl-document.json" \
  --profile "$TASK_TMP/profile.json" --context-id source-395-main \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/comparison.json"
jq '.result.certificate' "$TASK_TMP/comparison.json" > "$TASK_TMP/equivalence-certificate.json"
golingua algebra verify-proof --claim equivalence \
  --left "$TASK_TMP/en-document.json" --right "$TASK_TMP/pl-document.json" \
  --profile "$TASK_TMP/profile.json" --context-id source-395-main \
  --certificate "$TASK_TMP/equivalence-certificate.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json
```

Samo istnienie warunku w kontekście nie tworzy faktu. Identyfikatory wzmianek muszą pochodzić z rzeczywistych kandydatów składniowych i potwierdzonej evidencji rozstrzygnięcia; jawne `role_references` są założeniem wywołującego. Atrybuty referencji, presupozycje i zakres operatorów są zachowywane w dokumentach. Pełne kontrole porównania potwierdziły sześć par dla tego pilota; nie są oceną prawdy świata ani akceptacją językową właściciela (`NOT EVALUATED`). Zobacz [kontrakt składania analizy](architecture/semantic-algebra/source-composition.md), [przykłady](../data/semantic/examples/source-composition/) i [evidence #395](work/dse/evidence/07-02/README.md).


#### Jawny WH gap w składaniu źródłowym

Warianty WH są częścią osobnego profilu `algebra analyze`. W przykładach #395 pakiet i plan przypinają dokładny frame/variant; kontekst jawnie dostarcza event oraz rolę niebędącą luką. Wynik jest typed `Ask`, nie tekstową odpowiedzią ani wywołaniem `question` na MeaningGraph.

```sh
TASK_TMP=$(mktemp -d)
printf '{"schema_version":1,"dataset_store":"%s/store","dataset_id":"golingua-authored","language_package_dirs":[]}\n' "$TASK_TMP" > "$TASK_TMP/runtime.json"
golingua snapshot --apply --store "$TASK_TMP/store" --json

golingua algebra analyze --language en --text 'Who sees the dog?' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package data/semantic/examples/source-composition/wh-questions/packages/en.json \
  --source-plan data/semantic/languages/en/source-plan.json \
  --context data/semantic/examples/source-composition/wh-questions/contexts/subject.json \
  --metadata data/semantic/examples/source-composition/wh-questions/metadata.json --json

golingua algebra analyze --language en --text 'The teacher sees whom?' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package data/semantic/examples/source-composition/wh-questions/packages/en.json \
  --source-plan data/semantic/languages/en/source-plan.json \
  --context data/semantic/examples/source-composition/wh-questions/contexts/object.json \
  --metadata data/semantic/examples/source-composition/wh-questions/metadata.json --json

golingua algebra analyze --language pl --text 'Kto widzi psa?' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package data/semantic/examples/source-composition/wh-questions/packages/pl.json \
  --source-plan data/semantic/languages/pl/source-plan.json \
  --context data/semantic/examples/source-composition/wh-questions/contexts/subject.json \
  --metadata data/semantic/examples/source-composition/wh-questions/metadata.json --json

golingua algebra analyze --language pl --text 'Kogo widzi nauczyciel?' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package data/semantic/examples/source-composition/wh-questions/packages/pl.json \
  --source-plan data/semantic/languages/pl/source-plan.json \
  --context data/semantic/examples/source-composition/wh-questions/contexts/object.json \
  --metadata data/semantic/examples/source-composition/wh-questions/metadata.json --json
```

The four selected examples return `status=ok`, `composition.complete=true`, and a single `Ask` candidate. The answer variable occupies `experiencer` for `Who/Kto` and `stimulus` for `whom/kogo`; the complementary role and event remain explicit caller references. `role_sources[].source.span` and `role_sources[].wh_gap.span` identify the exact interrogative token, and the construction consumes the final `?`. A probe without `?` returns `unsupported`. The Polish object variant places the verb before the subject. The parser trace retains an `incomplete` forest because the WH argument has no NP candidate; composition treats it as covered only when the explicit gap exactly matches that frame slot. A caller-supplied reference for the answer role produces an `incomplete` result (`WH_ROLE_REFERENCE_CONFLICT`). The answer type is `Entity` and has no `Person` restrictor, so these probes do not encode the full lexical semantics of `who/whom/kto/kogo`. A declarative control still uses `en.transitive.svo` and returns an `Atom`. These examples preserve their authored #395 provenance and are not owner language approval or general WH coverage. Literal stdout/stderr, canonical documents, checks, input copies and binary hash are in [evidence #395](work/dse/evidence/07-02/README.md). Independent `algebra check` reports `Question` for each extracted document.

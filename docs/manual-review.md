# Ręczny przegląd outputu golingua

**Obowiązuje od 2026-10-20, z delegacją właściciela z 2026-10-23.** Nie ma automatycznej macierzy translacji, porównania z flexLingua, niezależnych goldów ani liczby przypadków wymaganej przed PR/merge. Właściciel może przekazać słowa, zdania lub korpus; agent uruchamia dostępne CLI, sam ocenia konkretne obserwacje i poprawia błędy w granicach aktywnego zadania. Ocena agenta nie jest podpisem ani akceptacją właściciela. Nie wymyślaj niepodanych wyników oczekiwanych. Sama dostępność komendy nie oznacza, że skonfigurowano runtime językowy.

## Jeden przypadek — prosta procedura

1. Właściciel podaje dowolny tekst i kierunek (`pl→en` albo `en→pl`). Może wskazać akceptowalny wariant, lecz nie musi. Nie wymyślaj tekstu właściciela, jeśli go nie podał.
2. Zapisz pełny SHA, dokładną komendę, surowe stdout, stderr i exit code z uruchomienia. Nie poprawiaj ani nie parafrazuj outputu. Jeśli komenda nie została uruchomiona, wpisz **NOT RUN** i nie twórz przykładowego wyniku.
3. Właściciel ocenia output jako `OK`, `BŁĄD`, `NIEPEWNE` albo `BRAK OBSŁUGI`. Krótkie uzasadnienie wystarczy. Status techniczny uruchomienia (PASS/FAIL/INFRA/NOT RUN) zapisuje się oddzielnie od oceny językowej.
4. Diagnozuj błąd programu osobno od złej fleksji, sensu lub tłumaczenia. Błąd uruchomienia, parsowania argumentów, panic, serializacji lub niezgodny kontrakt to problem programu. Niepoprawna forma językowa przy działającym CLI to obserwacja językowa, nie awaria builda.
5. Jeśli potrzebna jest poprawka, zmień właściwą leksykę/morfologię/konstrukcję/CORE. Nie zapisuj całej pary zdaniowej jako wyjątku. To samo wejście można uruchomić ponownie ręcznie; żaden wynik nie trafia automatycznie do CI.

### Status technicznego podglądu v0

`forms` i `catalog` są dostępne lokalnie. CLI może też załadować ograniczony,
skompilowany runtime PL/EN dla dokładnych tras `seeing`, `hearing`, `looking` i `receiving`, gdy `translate`
otrzyma `--runtime-config` wskazujący aktywny snapshot. Pierwsze trzy trasy obsługują teraźniejszą
3. osobę liczby pojedynczej, a `receiving` — czas przeszły w 3. osobie liczby pojedynczej.
Bez tej konfiguracji
żądanie translacji zwraca `unsupported` i exit code 3.

Na polecenie właściciela uruchomiono translację 100 zapisanych par w obu
kierunkach; obserwowane statusy i teksty zapisano w
[`data/sentences/observed-2026-10-23-provider-runtime.tsv`](../data/sentences/observed-2026-10-23-provider-runtime.tsv).
Po aktywacji trasy `hearing` uruchomiono też 25 zapisanych par w obu kierunkach;
wyniki z kodami wyjścia, statusami, diagnostykami i tekstami zapisano w
[`data/sentences/observed-2026-10-23-g05-12-hearing.tsv`](../data/sentences/observed-2026-10-23-g05-12-hearing.tsv).
Po aktywacji `receiving` uruchomiono 25 istniejących par z tego samego korpusu
w obu kierunkach; wyniki zapisano w
[`data/sentences/observed-2026-10-23-g05-13-receiving.tsv`](../data/sentences/observed-2026-10-23-g05-13-receiving.tsv).
Po aktywacji `looking` uruchomiono istniejące pary S051–S075 w obu kierunkach;
wyniki zapisano w
[`data/sentences/observed-2026-10-23-g05-14-looking.tsv`](../data/sentences/observed-2026-10-23-g05-14-looking.tsv).
To są obserwacje działania programu, a nie ocena językowa. Właściciel nie ocenił
jeszcze konkretnych wyjść, więc ręczny odbiór translacji i ocena form pozostają
`NOT RUN`. `forms` udostępnia ograniczoną odmianę oraz analizę obecnych seedów
PL/EN.

Po G10-01 ponownie uruchomiono wszystkie 100 istniejących par w obu kierunkach
na aktywnym snapshotcie. Dokładne kody wyjścia, statusy, diagnostyki i wyjścia
kandydatów zapisano w
[`data/sentences/observed-2026-10-23-g10-01-finite-verbs.tsv`](../data/sentences/observed-2026-10-23-g10-01-finite-verbs.tsv).
Kontrola konkretnych tras potwierdziła zachowanie dotychczasowych statusów:
EN→PL `seeing`, `hearing`, `looking` zwracają istniejące formy, a `receiving`
zwraca `Nauczyciel dostał wodę.` ze statusem `ok`. PL→EN zachowuje jawne
założenia wyboru rodzajnika. W rozszerzonym inwentarzu czasowników `water` nie
ma przypisanego paradygmatu, bo homografia z rzeczownikiem masowym `water`
tworzyła dodatkowy odczyt czasownikowy w istniejącej trasie `receiving`.
Odbiór właściciela nadal **NOT RUN**.

### G10-02 — polskie paradygmaty finite — 2026-10-23

`golingua` zbudowano z implementacji `54f7807030c4aa052bd95f1423a5be7503c12fea`. Aktywny snapshot `golingua-authored` w `/tmp/golingua-g10-02-store` ma digest `f950ea7251fd2a84dac477a309b7c2e0e6bdb1057a3950b27d0306b58764828e`. `forms` wygenerowało wybrane komórki: `widzi`, `widzą`, `słyszę`, `patrzy`, `dostanie`, `dostaniemy`, `dostałaś`, `dostałyśmy`, `dostaliście` i `dostali`. `forms --surface dostaliśmy --json` zwróciło `past_1pl_masculine_personal` z `aspect=perfective`, `gender=masculine_personal`, `number=plural`, `person=first`, `tense=past`, `verb_form=finite`, `mood=indicative` i `voice=active`.

Przykładowe komendy:

```sh
./golingua forms --language pl --lexeme lexeme:pl.widziec.v --feature verb_form=finite --feature tense=present --feature aspect=imperfective --feature mood=indicative --feature voice=active --feature person=third --feature number=plural --json
./golingua forms --language pl --surface dostaliśmy --json
./golingua translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config /tmp/golingua-g10-02-runtime.json --json
./golingua translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config /tmp/golingua-g10-02-runtime.json --json
```

| Wejście | Kierunek | Exit/status | Tekst kandydata |
| --- | --- | --- | --- |
| `The teacher sees the dog.` | EN→PL | `0` / `ok` | `Nauczyciel widzi psa.` |
| `Nauczyciel widzi psa.` | PL→EN | `7` / `incomplete` | `The teacher sees the dog.`; jawne założenia targetu pozostają widoczne |
| `The teacher got water.` | EN→PL | `0` / `ok` | `Nauczyciel dostał wodę.` |
| `Nauczyciel dostał wodę.` | PL→EN | `7` / `incomplete` | `The teacher got water.`; jawne założenia targetu pozostają widoczne |

Po zmianie ponownie wykonano 200 CLI uruchomień na 100 istniejących, właścicielskich par zdań; wyniki zapisano w [`data/sentences/observed-2026-10-23-g10-02-pl-finite-verbs.tsv`](../data/sentences/observed-2026-10-23-g10-02-pl-finite-verbs.tsv). W porównaniu z obserwacjami G10-01 kody wyjścia, statusy, zbiory tekstów kandydatów i zbiory diagnostyk pozostały takie same. Nie oznacza to, że wszystkie pary są tłumaczone w pełni; część nadal jest `unsupported`, `ambiguous` albo `incomplete`. Ocena agenta nie jest odbiorem właściciela; akceptacja właściciela pozostaje **NOT RUN**.

Komenda `question --runtime-config <path> --json` przyjmuje typowany MeaningGraph, jawny FactSet oraz mapowanie ról. Nie przyjmuje zdania w języku naturalnym. `--language pl|en` opcjonalnie uruchamia realizację odpowiedzi. Z delegacji właściciela agent może sprawdzić ograniczony przykład oparty na dostępnych danych projektu oraz jawnie oznaczonym fakcie demonstracyjnym; taki fakt jest agent-authored, nietrwały i nie jest wejściem ani akceptacją właściciela. Bez uruchomienia komendy wpisz **NOT RUN**. Właścicielska akceptacja językowa pozostaje osobnym statusem.

W bieżącym runtime dodatnia odpowiedź yes/no może zwrócić tekst PL/EN ze statusem `incomplete` i jawnymi założeniami. Odpowiedź przecząca oraz tekst WH pozostają `unsupported` do czasu dodania zatwierdzonych konstrukcji negacji i samodzielnej frazy NP. Strukturalna odpowiedź reasonera, proof i źródło faktu pozostają dostępne w wyniku.

### Delegowana ocena typed `question` — 2026-10-23

Po merge PR #209 uruchomiono zbudowane z `main` CLI (`1d41cbe`) na jednym
zapisanym przykładzie S001 (`seeing`, teacher, dog). Fakt `agent_demo_seeing_001`
był tymczasowy, request-scoped i jawnie agent-authored; nie zapisano go w
katalogu. Wyniki są obserwacją działania i oceną agenta, nie danymi ani
akceptacją właściciela.

Przykładowa komenda dodatnia PL:

```sh
/tmp/golingua-g06-130-review question --runtime-config /tmp/golingua-g06-128-runtime.json --language pl --json < /tmp/golingua-g06-128-question.json
```

EN używa tej samej komendy z `--language en`. Literalne pola powierzchniowe:

| Przypadek | Exit/status | Zaobserwowany wynik |
| --- | --- | --- |
| Dodatnia odpowiedź PL | `7` / `incomplete` | `Nauczyciel widzi psa.`; założenia: czas teraźniejszy, 3. osoba, liczba pojedyncza predykatu oraz domyślna liczba leksemu dla uczestników. |
| Dodatnia odpowiedź EN | `7` / `incomplete` | `The teacher sees the dog.`; założenia: czas teraźniejszy, 3. osoba, liczba pojedyncza, rodzajniki określone i liczba pojedyncza rzeczowników. |
| Jawny fakt ujemny | `3` / `unsupported` | Odpowiedź `false` i dowód pozostają w JSON; tekstu brak, bo nie ma zatwierdzonej konstrukcji negacji. |
| WH z luką `stimulus` | `3` / `unsupported` | Reasoner zwraca `dog_1` i proof; tekstu brak, bo nie ma samodzielnej konstrukcji NP. |
| Odwrócona kolejność ról | `2` / `invalid` | `QUERY_EVENT_ROLE_ORDER_MISMATCH`. |

Istniejące zdanie EN→PL zwróciło `ok` i `Nauczyciel widzi psa.`; PL→EN zwróciło
`incomplete` i `The teacher sees the dog.`. Kontrole techniczne tego commita:
`gofmt`, `git diff --check`, `go vet ./...`, `go build ./cmd/golingua`, help/version
i rzeczywiste CI PR #209 przeszły. `go test ./...` nie uruchamiano. Właścicielska
ocena językowa pozostaje **NOT RUN**.

### Delegowany przegląd tras `translate` — 2026-10-23

Ograniczony przegląd użył istniejących, oznaczonych `draft_unreviewed` rekordów
S001, S026, S053 i S077 z `data/sentences/100-pl-en.tsv`: po jednym dla tras
`seeing`, `hearing`, `looking` i `receiving`, w obu kierunkach. Bazowy `main`
przed poprawką miał SHA
`6d7b2e85c1f8f57175f47f2bc00d54eb0c79e280`. Końcowe wyniki poniżej uruchomiono
z binarki zbudowanej z review commita
`8ed705d161dd03965a06ce1ff04e900bb1b3a92c`.

Komendy miały postać:

```sh
/tmp/golingua-g08-65-review-final translate --from pl --to en --text 'Nauczyciel widzi psa.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from en --to pl --text 'The teacher sees the dog.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from pl --to en --text 'Nauczyciel słyszy psa.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from en --to pl --text 'The teacher hears the dog.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from pl --to en --text 'Nauczyciel patrzy na psa.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from en --to pl --text 'The teacher looks at the dog.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from pl --to en --text 'Nauczyciel dostał wodę.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
/tmp/golingua-g08-65-review-final translate --from en --to pl --text 'The teacher got water.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
```

Literalne pola wyjścia po poprawce:

| Rekord | Kierunek | Exit/status | Tekst | Assumptions / diagnostyka |
| --- | --- | --- | --- | --- |
| S001 | PL→EN | `7` / `incomplete` | `The teacher sees the dog.` | `determiner=definite` dla `event/ref0` i `event/ref1`; candidate `EXPLICIT_REALIZATION_ASSUMPTION`, wynik `TRANSLATION_NOT_OK`. |
| S001 | EN→PL | `0` / `ok` | `Nauczyciel widzi psa.` | brak assumptions i diagnostyk. |
| S026 | PL→EN | `7` / `incomplete` | `The teacher hears the dog.` | `determiner=definite` dla obu uczestników; candidate `EXPLICIT_REALIZATION_ASSUMPTION`, wynik `TRANSLATION_NOT_OK`. |
| S026 | EN→PL | `0` / `ok` | `Nauczyciel słyszy psa.` | brak assumptions i diagnostyk. |
| S053 | PL→EN | `7` / `incomplete` | `The teacher looks at the dog.` | `determiner=definite` dla obu uczestników; candidate `EXPLICIT_REALIZATION_ASSUMPTION`, wynik `TRANSLATION_NOT_OK`. |
| S053 | EN→PL | `0` / `ok` | `Nauczyciel patrzy na psa.` | brak assumptions i diagnostyk. |
| S077 | PL→EN | `7` / `incomplete` | `The teacher got the water.` | `determiner=definite` dla `event/ref0` (water) i `event/ref1` (teacher); candidate `EXPLICIT_REALIZATION_ASSUMPTION`, wynik `TRANSLATION_NOT_OK`. |
| S077 | EN→PL | `0` / `ok` | `Nauczyciel dostał wodę.` | brak assumptions i diagnostyk. |

Assumptions są dołączone do każdego candidate z wartością, powodem i ścieżką
Meaning; identyczne powtórzenia diagnostyki target są zwracane raz. S077 pokazuje
konkretny wybór `the water`, ale rekord źródłowy jest `draft_unreviewed`, a
Meaning nie koduje określoności. To jawne założenie, nie rozstrzygnięcie, czy
wariant jest językowo akceptowalny. Akceptacja właściciela pozostaje
**NOT RUN**. Kontrole review commita: `gofmt`, `git diff --check`, `go vet ./...`,
`go build ./cmd/golingua`, `--help` i `--version`; `go test ./...` nie uruchamiano.

### G08-02 — bare object dla angielskiego rzeczownika niepoliczalnego — 2026-10-23

Ręczne uruchomienia wykonano z binarki zbudowanej z commita
`e9d36e9568eda689117bad04dc8d61f215719d12` i snapshotu
`/tmp/golingua-g06-128-runtime.json`. Wejścia pochodzą z istniejących rekordów
S077, S079 i S076 w `data/sentences/100-pl-en.tsv`; rekordy nadal są
`draft_unreviewed`. Właściciel wybrał dla S077 angielskie `water` bez
rodzajnika. Pozostałe wyjścia są obserwacją agenta, nie ich akceptacją.

Komendy translacji (filtr `jq` ogranicza wydruk do pól istotnych dla ręcznej
obserwacji; exit code pochodzi z CLI przez `set -o pipefail`):

```sh
set -o pipefail; /tmp/golingua-g08-66-final translate --from pl --to en --text 'Nauczyciel dostał wodę.' --runtime-config /tmp/golingua-g06-128-runtime.json --json | jq -c '{status,text,candidates:[.candidates[]|{outcome,text,assumptions}],diagnostics}'
set -o pipefail; /tmp/golingua-g08-66-final translate --from pl --to en --text 'Nauczyciel dostał wiadomość.' --runtime-config /tmp/golingua-g06-128-runtime.json --json | jq -c '{status,text,candidates:[.candidates[]|{outcome,text,assumptions}],diagnostics}'
set -o pipefail; /tmp/golingua-g08-66-final translate --from en --to pl --text 'The teacher got water.' --runtime-config /tmp/golingua-g06-128-runtime.json --json | jq -c '{status,text,candidates:[.candidates[]|{outcome,text,assumptions}],diagnostics}'
/tmp/golingua-g08-66-final translate --from pl --to en --text 'Nauczyciel dostał książkę.' --runtime-config /tmp/golingua-g06-128-runtime.json --json
```

| Rekord | Kierunek | Exit/status | Tekst / obserwacja |
| --- | --- | --- | --- |
| S077 | PL→EN | `7` / `incomplete` | Candidate: `The teacher got water.`. Założenia: `determiner=bare` dla `event/ref0` (water) oraz `determiner=definite` dla `event/ref1` (teacher). Diagnostyka `TRANSLATION_NOT_OK`; stdout JSON, stderr pusty. |
| S079 | PL→EN | `7` / `incomplete` | Candidate: `The teacher got the message.`. Oba uczestniki mają `determiner=definite`; wariant dla rzeczownika policzalnego pozostał bez zmian. Diagnostyka `TRANSLATION_NOT_OK`; stdout JSON, stderr pusty. |
| S077 | EN→PL | `0` / `ok` | `Nauczyciel dostał wodę.`; bez assumptions i diagnostyk; stderr pusty. |
| S076 | PL→EN | `3` / `unsupported` | Tekstu brak: polski `książkę` ma konkurencyjny sens `pl.ksiazka.work` bez zatwierdzonego równoważnego CORE bindingu (`ROLE_MAPPING_UNAVAILABLE`); to istniejąca luka leksykalna, nie wynik wyboru bare-object. |

S077 potwierdza wybraną formę bez rodzajnika i zachowuje `incomplete`, bo
Meaning nie koduje określoności obiektu ani odbiorcy. S079 pokazuje, że jawnie
policzalny leksem korzysta ze starego wariantu z `the`. Właścicielska akceptacja
językowa pozostaje **NOT RUN**.

### G10-05 — częściowa leksykalizacja przy nierozstrzygniętej polisemii — 2026-10-23

Uruchomienia wykonano binarką z implementacji
`2ac848623e9b2ad07428702feb365cc44870c7e0` i aktywnym snapshotem wskazanym w
`/tmp/golingua-g10-04-runtime.json`. Są to trzy istniejące rekordy właścicielskiej
listy zdań, wszystkie nadal `draft_unreviewed`. Dokładne kody, teksty, diagnostyki
i nierozstrzygnięte wpisy evidence zapisano w
[`observed-2026-10-23-g10-05-partial-lexicalization.tsv`](../data/sentences/observed-2026-10-23-g10-05-partial-lexicalization.tsv).

```sh
/tmp/golingua-g10-05 translate --from en --to pl --text 'The doctor hears the dog.' --runtime-config /tmp/golingua-g10-04-runtime.json --json
/tmp/golingua-g10-05 translate --from pl --to en --text 'Nauczyciel widzi książkę.' --runtime-config /tmp/golingua-g10-04-runtime.json --json
/tmp/golingua-g10-05 translate --from en --to pl --text 'The teacher got a book.' --runtime-config /tmp/golingua-g10-04-runtime.json --json
```

Każda komenda zwróciła exit `7` i status `incomplete`; tekst znajduje się w
`candidates[0].text`, podczas gdy główne pole `text` pozostaje puste. Dla
`S003` jawne założenie rodzajnika `the` jest dodatkowo widoczne w diagnostyce.
Evidence zachowuje odpowiednio `en.doctor.academic`, `pl.ksiazka.work` i
`en.book.work` jako odczytania bez zatwierdzonego ekwiwalentu. Żaden binding nie
został dodany ani zatwierdzony.

**Ocena agenta:** `Lekarz słyszy psa.` i `Nauczyciel dostał książkę.` zachowują
sens i role z istniejących zdań; `The teacher sees the book.` jest możliwym
tekstem docelowym, a określoność pozostaje jawnym założeniem, bo polski input jej
nie koduje. To ocena tych konkretnych outputów, nie automatyczna bramka ani
akceptacja właściciela. Akceptacja właściciela pozostaje **NOT RUN**.

### G10-06 — jawny podgląd niekompletnego kandydata — 2026-10-23

Wywołania wykonano binarką z implementacji
`0da6ad7726af85af6ac92fade774061d3c4e7815` i tym samym aktywnym snapshotem.
Dokładne komendy, ID, statusy, teksty i diagnostyki zapisano w
[`observed-2026-10-23-g10-06-explicit-preview.tsv`](../data/sentences/observed-2026-10-23-g10-06-explicit-preview.tsv).
Te rekordy pochodzą z istniejącej listy zdań i pozostają `draft_unreviewed`.

```sh
/tmp/golingua-g10-06 translate --from en --to pl --text 'The doctor hears the dog.' --runtime-config /tmp/golingua-g10-04-runtime.json --preview-id translation:c71dddd4d2e8c7e1013e13ffa65a2510df64293a831e5c45950fd582cb90ad8f --json
/tmp/golingua-g10-06 translate --from pl --to en --text 'Nauczyciel widzi książkę.' --runtime-config /tmp/golingua-g10-04-runtime.json --preview-id translation:beb0937ce5db7011fecea34b89abdb49fc96ca44a31b56314d0bceb718da69b4 --json
/tmp/golingua-g10-06 translate --from en --to pl --text 'The teacher got a book.' --runtime-config /tmp/golingua-g10-04-runtime.json --preview-id translation:3a59356f7882da043c32eb99336aaf2e1649de7a532ba1419d1235c605c074d1 --json
```

S031, S003 i S076 zwróciły exit `7` oraz status `incomplete`, a ich jawnie
wskazany tekst pojawił się zarówno w `text`, jak i w wybranym candidate. Próba
na S036 pokazała podgląd `The child hears the dog.` z zachowanym statusem
`ambiguous` i exit `6`; wskazanie drugiego kandydata, który ma status
`unsupported` i nie ma tekstu, zwróciło `invalid` / exit `2`. Powtórzone
identyczne wywołania S031 zwróciły to samo candidate ID, więc można skopiować
ID z odpowiedzi i użyć go w kolejnej komendzie.

**Ocena agenta:** podgląd ujawnia tylko tekst jawnie wskazanego kandydata i nie
zmienia statusu ani evidence. Akceptacja językowa właściciela pozostaje
**NOT RUN**.

### G10-07 — wyniki zapisanych 100 par w obu kierunkach — 2026-10-23

Każdy z 100 istniejących wierszy uruchomiono jako PL→EN i EN→PL na binarce z
implementacji `0da6ad7726af85af6ac92fade774061d3c4e7815` i tym samym aktywnym
snapshotem. Wszystkie 200 bazowych odpowiedzi oraz kandydatury zapisano w
[`observed-2026-10-23-g10-07-runtime-preview.tsv`](../data/sentences/observed-2026-10-23-g10-07-runtime-preview.tsv).
Dla tekstowych kandydatów z wynikiem `incomplete` lub `ambiguous` wykonano osobne
wywołanie z `--preview-id`; wyniki `ok` miały tekst bez preview.

EN→PL zwróciło status `ok` dla 60 wejść i `incomplete` dla 40; wszystkie miały
tekst. PL→EN zwróciło `incomplete` dla 75 i `ambiguous` dla 25; każde wejście
miało co najmniej jednego tekstowego kandydata, a w zachowanych alternatywach
znalazły się również przypadki bez tekstu. CLI potwierdziło 140 jawnych
podglądów: wskazany tekst pojawił się w głównym `text`, a status wyniku i
kandydata pozostał bez zmian. Różne teksty nie były rozstrzygane kolejnością.

**Ocena agenta:** wyjścia dla prostych tras `seeing`, `hearing` i `looking`
zachowują role i konstrukcje wejścia. W `receiving` polski input nie przenosi
określoności. Dla pojedynczych dopełnień `book` i `message` obecna angielska
kandydatura wybiera `the`, podczas gdy zapisane pary EN zawierają `a`; wybór
pozostaje założeniem i nie zmieniam statusu na `ok`. Ten konkretny przypadek
jest ujęty w kolejce jako G10-08 #228. Akceptacja językowa właściciela pozostaje
**NOT RUN**.

### G10-08 — alternatywy rodzajnika dla `receiving` — 2026-10-23

Po G10-08 uruchomiono istniejące rekordy S076, S077, S079 i S080 z pliku
[`100-pl-en.tsv`](../data/sentences/100-pl-en.tsv) w obu kierunkach. Dokładne
kody wyjścia, statusy, candidate ID, teksty i jawne podglądy zapisano w
[`observed-2026-10-23-g10-08-indefinite-receiving.tsv`](../data/sentences/observed-2026-10-23-g10-08-indefinite-receiving.tsv).
PL→EN dla S076 zwróciło `The teacher got the book.` i `The teacher got a book.`;
S079 zwróciło `The teacher got the message.` i `The teacher got a message.`.
Wynik zbiorczy pozostał `ambiguous`, każdy wariant miał własny status
`incomplete` i osobne założenie rodzajnika, a oba teksty dało się jawnie
wyświetlić przez `--preview-id`. S077 i S080 zachowały teksty bez rodzajnika
(`The teacher got water.` i `The teacher got money.`). EN→PL zachowało
obserwowane teksty polskie dla tych czterech istniejących wejść.

Podczas pracy wykryto i naprawiono błąd agregacji: globalne `ambiguous` selekcji
oznaczało wcześniej również podmiot jako niejednoznaczny i blokowało jego
leksykalizację. Statusy i diagnostyki kandydata są teraz lokalne dla jego węzła
i trasy; status zbiorczy nadal pokazuje nierozstrzygnięte alternatywy.
Akceptacja językowa właściciela i redakcyjny status zdań nadal pozostają osobne;
właścicielska akceptacja pozostaje **NOT RUN**.

### G10-10 — alternatywy rodzajnika EN w transitive i oblique — 2026-10-23

Na implementacji [`51c9a07980d814aa67fac2fd9eb829ac95c1bfaf`](../data/sentences/observed-2026-10-23-g10-10-determiner-alternatives.tsv)
uruchomiono wszystkie 100 istniejących wierszy PL/EN w obu kierunkach — 200
odpowiedzi CLI. Zapis zawiera wyniki, wszystkie zachowane kandydatury i 16
oddzielnych wywołań `--preview-id` na próbkach tras `seeing`, `looking` i
`receiving`.

PL→EN: wszystkie 100 wyników zbiorczych ma status `ambiguous`; każdy wynik ma
co najmniej jeden tekstowy kandydat. Zapisano 402 kandydatury, w tym 370 z
tekstem i 32 bez tekstu. EN→PL: 60 wyników ma status `ok`, 40 `incomplete`;
wszystkie 100 mają tekst. S001 udostępnia cztery warianty `a/the` dla podmiotu
i obiektu; S051 zachowuje `look at` przy tych samych wyborach rodzajnika;
S076 ma cztery count-NP warianty, a S077 zachowuje bezrodzajnikowe `water` i
dwie alternatywy rodzajnika dla podmiotu. Wszystkie 16 jawnych podglądów
zwróciło tekst wskazanego kandydata i jego identyfikator.

**Ocena agenta:** polskie źródło nie określa rodzajnika, dlatego warianty
pozostają osobnymi, jawnymi kandydatami z założeniami; statusy `incomplete` i
`ambiguous` nie są zamieniane na `ok`. Wariant `a` dopuszczono tylko dla
leksemów z jawną właściwością `indefinite_article=a`; wymowa innych leksemów
nie jest zgadywana. Nierozstrzygnięte bindingi i brakujące teksty kandydatów
pozostają widoczne. Akceptacja językowa właściciela pozostaje **NOT RUN**.

### G10-11 #235 (PR #236) — regularne formy EN dla `household` — 2026-10-23

Na implementacji [`be6bf497b6bae0780bec1cc988be873f37edac83`](../data/sentences/observed-2026-10-23-g10-11-household-morphology.tsv) uruchomiono oba kierunki dla 10 istniejących rekordów: S004, S009, S014, S019, S024, S052, S057, S062, S067 i S072 — 20 odpowiedzi CLI. `en.household.n` korzysta teraz z regularnej klasy `en_noun_regular`, a sentence-facing właściwość `indefinite_article=a` dopuszcza `a household`; `forms` rozpoznało `household` i `households`.

PL→EN zachowało status `ambiguous` we wszystkich 10 wynikach. W 84 zachowanych kandydaturach 80 ma tekst, w tym 40 wariantów `household`; 4 kandydatury nadal są `unsupported` z powodu nierozstrzygniętego `child_of_person`. EN→PL zachowało 10 wyników `incomplete` z tekstem oraz `PARTIAL_PARSE_EVIDENCE`/`UNRESOLVED_ROLE_SENSE`. W poprzedniej obserwacji tych rekordów 10 kandydatur miało `PRINCIPAL_PARTS_UNAVAILABLE`; po zmianie w wybranej próbce nie pozostała żadna taka diagnostyka. Statusy semantyczne, lexical bindings oraz redakcyjny status zdań nie zostały zmienione.

**Ocena agenta:** dodane formy są regularnym zapisem istniejącego leksemu i odsłaniają kandydatury dla już zapisanych odczytów `dom`; nie rozstrzygają, czy w zdaniu chodzi o budynek, czy gospodarstwo domowe. Wyniki nadal wymagają ręcznej oceny; akceptacja właściciela pozostaje **NOT RUN**.

Bez tekstu dostarczonego przez właściciela oznacz ocenę językową **NOT EVALUATED**. Jeśli nie uruchomiono konkretnej próby na wejściu właściciela, oznacz ją **NOT RUN**. Gdy właściciel deleguje autonomiczną ocenę, agent może opisać własną ocenę na tych wejściach, ale status akceptacji właściciela pozostaje osobny. Techniczne smoke testy z tekstem zastępczym nie są wejściem do oceny językowej. Dokładny kontrakt i przykład komendy opisuje [docs/cli.md](cli.md); nie przedstawiaj go jako dowodu, że provider jest skonfigurowany.

Nie ma obowiązku zapisywania każdego outputu do pliku. Jeśli zapis jest przydatny, prowadź opcjonalnie krótką notatkę w issue lub `examples/manual/`:

```text
Implementacja/wersja: <pełny SHA lub wersja>
Tryb: <lokalna komenda / rzeczywisty serwis>
Kierunek: <pl -> en / en -> pl>
Wejście właściciela: <dokładny tekst albo prywatny odnośnik>
Komenda: <dokładne polecenie, jeśli uruchomiono>
Wynik techniczny: PASS | FAIL | INFRA | NOT RUN
Exit code: <kod albo NOT RUN>
stdout: <literalny output albo NOT RUN>
stderr: <literalny output albo NOT RUN>
Ocena właściciela: OK | BŁĄD | NIEPEWNE | BRAK OBSŁUGI | NOT EVALUATED
Uwaga: <opcjonalna diagnoza lub poprawka>
```

Nie zapisuj prywatnego wejścia lub rozmowy do repo bez potrzeby i zgody właściciela. Notatki w [`examples/manual/`](../examples/manual/README.md) są opcjonalne i nie są czytane przez CI.

Wyniki diagnostyczne muszą odróżniać: brak leksemu, brak sensu, nieobsługiwaną konstrukcję, wieloznaczność, limit i rzeczywisty błąd programu. **Zielone CI = da się zbudować i uruchomić CLI, nie jakość językowa.**

## Co NIE jest wymagane

- Przenoszenie dawnych zdań lub ich oczekiwanych rezultatów z JS.
- Tworzenie `tests/cases`, `fixtures`, formalnych goldów, oracle, raportów PASS% ani test runnera.
- Uruchamianie masowego automatycznego `go test ./...`.
- Testowanie wszystkich kombinacji konstrukcji, liczb, czasów w jednym PR.
- Przyjmowanie przez agenta roli arbitra zgodności językowej bez właściciela.

## Bezpieczeństwo i ograniczenia

Ograniczenie testów **nie** oznacza rezygnacji z walidacji danych wejściowych, ograniczania zasobów, jawnych wyjątków ani poprawnego rozdzielenia ról semantycznych. To własności runtime. Na tym etapie oceniamy ręcznie przede wszystkim realne zachowanie PL/EN. Gdy projekt dojrzeje, użytkownik może osobno zlecić powrót do testów automatycznych.

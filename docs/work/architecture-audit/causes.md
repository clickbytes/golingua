# Rodziny przyczyn i granice diagnozy

Każda kontrola ma dokładnie jedną przyczynę główną; dodatkowe blokery są w osobnej kolumnie checks.tsv. Liczby dotyczą kontroli, nie niezależnych błędów. „Ustalone” oznacza potwierdzoną granicę w kodzie i wyniku. „Ograniczona” oznacza, że nie rozstrzygnięto głębszej przyczyny dla każdego rekordu; „hipoteza” wymaga oceny językowej. Rodzina nie czyni wszystkich jej rekordów jednym defektem.

## REF

Kategoria: **kontrola odniesienia**. Kontrole: **297**. Pewność: **ustalone**. Pierwsza potwierdzona granica: wymagania wybranego rekordu.

Spełnia zapisany kontrakt; nie dowodzi kompletności języka.

Źródło: [`tests/matrix/checks.go:61`](../../../tests/matrix/checks.go#L61). Przedstawiciel: `A1-AC-001-EN-PL`, `analysis`; wynik `PASS`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-CORE

Kategoria: **dane**. Kontrole: **402**. Pewność: **ustalone co do katalogu**. Pierwsza potwierdzona granica: CORE → legalny predykat i role.

Brakuje wymaganego konceptu lub roli. Wymagany dobór semantyczny pozostaje oceną agenta.

Źródło: [`internal/catalog/concepts_relations.go:9`](../../../internal/catalog/concepts_relations.go#L9). Przedstawiciel: `A1-LQ-001-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-EXIST

Kategoria: **dane**. Kontrole: **156**. Pewność: **ustalone**. Pierwsza potwierdzona granica: CORE existence → rola liczonego referenta.

Istnieje komponent exactly(n), lecz brak czasowanego predykatu istnienia z rolą; dodatkowo domena i wiązanie ilości wymagają kontraktu.

Źródło: [`tests/interlingua/quantity_time_gap_test.go:48`](../../../tests/interlingua/quantity_time_gap_test.go#L48). Przedstawiciel: `A1-LQ-003-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## I-REL

Kategoria: **reprezentacja znaczenia**. Kontrole: **162**. Pewność: **ustalone**. Pierwsza potwierdzona granica: CORE KindRelation → Meaning state.

Walidator nie przyjmuje KindRelation jako state/event. Alternatywą może być jawny predykat stanu w danych, po decyzji semantycznej; nie dowodzi konieczności redesignu.

Źródło: [`tests/interlingua/possession_gap_test.go:15`](../../../tests/interlingua/possession_gap_test.go#L15). Przedstawiciel: `A1-Q-003-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## I-WH

Kategoria: **reprezentacja znaczenia**. Kontrole: **204**. Pewność: **ustalone**. Pierwsza potwierdzona granica: pełne I → typowana zmienna WH.

Question przechowuje target, nie zmienną i jej restrykcję. GapRole działa jako osobne wejście reasonera, nie jest samowystarczalnym I.

Źródło: [`internal/meaning/types.go:48`](../../../internal/meaning/types.go#L48). Przedstawiciel: `A1.9-Q-1009-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## I-TIME

Kategoria: **reprezentacja znaczenia**. Kontrole: **72**. Pewność: **ustalone**. Pierwsza potwierdzona granica: wymaganie czasu/aspektu → I.

Tylko past/present; brak frequency/progressive i relacji czasowej między zdarzeniami.

Źródło: [`internal/meaning/attributes.go:1`](../../../internal/meaning/attributes.go#L1). Przedstawiciel: `A1.10-TIME-1097-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## I-SCOPE

Kategoria: **reprezentacja znaczenia**. Kontrole: **24**. Pewność: **ustalone**. Pierwsza potwierdzona granica: wiązanie zmiennej/kwantyfikatora → I.

Kwantyfikator ma target/count, brak wiązania zmiennej z rolą zdarzenia i wybranego zakresu dystrybutywnego.

Źródło: [`internal/meaning/types.go:48`](../../../internal/meaning/types.go#L48). Przedstawiciel: `A1.12-AMB-1265-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## I-COMPOSE

Kategoria: **reprezentacja znaczenia**. Kontrole: **60**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: pełna relacja między treściami → I.

Brak wybranego operatora/grupy/relacji w schemacie. Dla modality/politeness to ocena pełności wymagań, nie dowód niemożności kodowania każdego odczytu.

Źródło: [`internal/meaning/types.go:48`](../../../internal/meaning/types.go#L48). Przedstawiciel: `A1.13-COMP-1323-EN-PL`, `analysis`; wynik `GAP_I`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## M-INDEX

Kategoria: **morfologia**. Kontrole: **762**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: token → odczyt indeksu form.

Indeks nie zawiera wymaganej powierzchni. Log dowodzi braku odczytu, nie rozstrzyga dla każdego tokenu braku leksemu, principal parts, komórki albo segmentacji.

Źródło: [`internal/analyzer/scan.go:142`](../../../internal/analyzer/scan.go#L142). Przedstawiciel: `A1-Q-002-PL-EN`, `analysis`; wynik `GAP_ANALYZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## A-PLAN

Kategoria: **plugin analizy**. Kontrole: **184**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: odczyty/frazy → pełne parsowanie.

Brak pełnej trasy dla tego wejścia; przy znanych formach blokują zakresy/cechy planu lub dostarczone frazy. Dokładny odrzucony constraint nie jest w każdym logu.

Źródło: [`internal/parser/match.go:75`](../../../internal/parser/match.go#L75). Przedstawiciel: `A1-Q-001-EN-PL`, `analysis`; wynik `GAP_ANALYZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-SENSE

Kategoria: **dane**. Kontrole: **37**. Pewność: **ustalone**. Pierwsza potwierdzona granica: sens roli → zatwierdzony binding.

Zatwierdzony graf zachowany, lecz konkurencyjny sens bez równoważnego bindingu utrzymuje incomplete.

Źródło: [`internal/semantic/compose.go:830`](../../../internal/semantic/compose.go#L830). Przedstawiciel: `A1.1-AC-101-EN-PL`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-NUMBER

Kategoria: **dane**. Kontrole: **14**. Pewność: **ustalone**. Pierwsza potwierdzona granica: właściwość leksemu → multiplicity I.

Brak number_behavior=count dla letter/window/apple. Analizator zapisuje multiplicity tylko dla jawnych count nouns.

Źródło: [`languages/en/lexicon.go:1271`](../../../languages/en/lexicon.go#L1271). Przedstawiciel: `A1.10-TIME-1041-EN-PL`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## A-NUMBER

Kategoria: **plugin analizy**. Kontrole: **2**. Pewność: **ustalone**. Pierwsza potwierdzona granica: negacja → przypadek slotu → synkretyczny odczyt.

Negatywna konstrukcja odziedziczyła accusative zamiast genitive. Książki wybiera accusative plural, traci genitive singular i zmienia liczbę Meaning.

Źródło: [`languages/pl/valency.go:77`](../../../languages/pl/valency.go#L77). Przedstawiciel: `A1.2-NEG-227-PL-EN`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-POLYSEMY

Kategoria: **dane**. Kontrole: **24**. Pewność: **ustalone**. Pierwsza potwierdzona granica: forma child/children → inventory sensów.

Child ma tylko sens wieku; znaczenie potomka przypisano osobnemu leksemowi offspring. Nie istnieje przejście od child do wymaganego child_of_person.

Źródło: [`languages/en/lexicon.go:31`](../../../languages/en/lexicon.go#L31). Przedstawiciel: `A1.1-AC-107-EN-PL`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## A-OVS

Kategoria: **plugin analizy**. Kontrole: **12**. Pewność: **ustalone**. Pierwsza potwierdzona granica: synkretyczne frazy → wszystkie licencjonowane role.

Brak pełnej analizy OVS; dla pluralnych koty/psy brak wymaganej alternatywy zamiany ról. Nie jest to błąd bijekcji porównania I.

Źródło: [`languages/pl/constructions/transitive.json:1`](../../../languages/pl/constructions/transitive.json#L1). Przedstawiciel: `A1.12-AMB-1242-PL-EN`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-NEG-CASE

Kategoria: **plugin realizacji**. Kontrole: **6**. Pewność: **ustalone**. Pierwsza potwierdzona granica: negatywna konstrukcja → przypadek dopełnienia.

Rodzina nie zmienia accusative ramy na genitive. Generator poprawnie wykonuje błędne żądanie: nie widzi książkę, nie słyszy ptaki.

Źródło: [`languages/pl/constructions/transitive_negative.json:7`](../../../languages/pl/constructions/transitive_negative.json#L7). Przedstawiciel: `A1.2-NEG-228-EN-PL`, `realization`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-DETERMINER

Kategoria: **plugin realizacji**. Kontrole: **62**. Pewność: **ustalone**. Pierwsza potwierdzona granica: nieokreśloność → wszystkie wymagane warianty.

Polityka plural bez źródłowego rodzajnika zachowuje tylko definite; reading ma osobny wariant z terminalami the. Brak wymaganych bare plural i wariantów a/the; I nie musi zgadywać określoności.

Źródło: [`languages/en/sentence_target.go:537`](../../../languages/en/sentence_target.go#L537). Przedstawiciel: `A1.10-TIME-1042-PL-EN`, `realization`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## C-AMB

Kategoria: **kontrakt testu**. Kontrole: **2**. Pewność: **hipoteza**. Pierwsza potwierdzona granica: licencjonowane odczyty → zbiór allowed I.

To jest ryba dopuszcza kandydaturę food w kodzie; kontrakt dopuszcza tylko animal. Nie ustalono niezależnie, czy food i przypisane one są poprawne dla tego kontekstu.

Źródło: [`tests/languages/pl/A1.5/identification.testdomain.go:22`](../../../tests/languages/pl/A1.5/identification.testdomain.go#L22). Przedstawiciel: `A1.5-ID-532-PL-EN`, `analysis`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-CONTEXT

Kategoria: **plugin realizacji**. Kontrole: **20**. Pewność: **ustalone**. Pierwsza potwierdzona granica: Meaning + participant anchor → NP/osoba.

Speaker/addressee nie steruje leksykalizacją ani osobą; występuje The person sees zamiast I see. PL ma także brak principal parts person jako osobny bloker.

Źródło: [`languages/en/sentence_target.go:152`](../../../languages/en/sentence_target.go#L152). Przedstawiciel: `A1.10-TIME-1002-PL-EN`, `realization`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-PARTS

Kategoria: **dane**. Kontrole: **29**. Pewność: **ustalone**. Pierwsza potwierdzona granica: wybrany leksem → principal parts.

Wybrany docelowy leksem nie ma attested principal parts w seed. Katalog leksemów jest szerszy od morfologii.

Źródło: [`internal/runtime/slot_execution.go:138`](../../../internal/runtime/slot_execution.go#L138). Przedstawiciel: `A1.10-TIME-1001-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## M-CELL

Kategoria: **morfologia**. Kontrole: **7**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: features slotu → komórka/formy.

Żądana komórka nie została rozwiązana. Log i sondy pokazują odmowę; nie każdy rekord ma osobny ślad odrzuconych komórek.

Źródło: [`internal/realizer/execute.go:1`](../../../internal/realizer/execute.go#L1). Przedstawiciel: `A1.10-TIME-1033-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-PAST-AGREE

Kategoria: **plugin realizacji**. Kontrole: **4**. Pewność: **ustalone**. Pierwsza potwierdzona granica: past plan cech → hard agreement.

Dla past helper pomija number; rodzina wymaga number na subject i verb. Saw istnieje, odmowa jest przed poprawną linearyzacją.

Źródło: [`languages/en/sentence_target.go:145`](../../../languages/en/sentence_target.go#L145). Przedstawiciel: `A1.10-TIME-1034-PL-EN`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## D-LICENSE

Kategoria: **dane**. Kontrole: **156**. Pewność: **ustalone**. Pierwsza potwierdzona granica: sens/ramy → dokładna licencja konstrukcji celu.

Brak zatwierdzonej kombinacji transferu i dokładnej konstrukcji przed filtrem polityki celu; sama obecność słowa/ramy nie wystarcza.

Źródło: [`internal/realizer/select.go:543`](../../../internal/realizer/select.go#L543). Przedstawiciel: `A1.10-TIME-1045-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-POLICY

Kategoria: **plugin realizacji**. Kontrole: **47**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: wybór → retained surface candidate.

Brak kandydata powierzchniowego po wyborze/politykach/kontrolach kształtu. Dokładna gałąź odrzucenia nie jest logowana dla każdego rekordu.

Źródło: [`internal/runtime/target.go:112`](../../../internal/runtime/target.go#L112). Przedstawiciel: `A1.10-TIME-1081-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-COMPOSE

Kategoria: **plugin realizacji**. Kontrole: **68**. Pewność: **ustalone**. Pierwsza potwierdzona granica: wiele węzłów leksykalnych → konstrukcja pokrywająca.

Żadna wybrana konstrukcja nie pokrywa wszystkich węzłów; obsługiwana black-subject+seeing jest węższa od ogólnej kompozycji.

Źródło: [`internal/runtime/target.go:117`](../../../internal/runtime/target.go#L117). Przedstawiciel: `A1.11-CASE-1105-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-OPERATOR

Kategoria: **plugin realizacji**. Kontrole: **116**. Pewność: **ustalone**. Pierwsza potwierdzona granica: zakres/akt → jawny mapping konstrukcyjny.

Brak pełnego mapowania operatorów/aktu w adapterze celu; nie jest to dowód braku operatora w I.

Źródło: [`internal/runtime/target.go:83`](../../../internal/runtime/target.go#L83). Przedstawiciel: `A1-Q-001-EN-PL`, `realization`; wynik `GAP_REALIZE`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## C-ASSUMPTION

Kategoria: **kontrakt testu**. Kontrole: **71**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: tekst/założenia → allowed pola kontraktu.

Zwykle poprawny tekst nominalnej 3. osoby odrzucony za person=third lub brak wymaganych wariantów. Kontrakt i annotator są niespójne; nie zmieniano żadnej strony.

Źródło: [`tests/matrix/checks.go:629`](../../../tests/matrix/checks.go#L629). Przedstawiciel: `A1.11-CASE-1101-EN-PL`, `realization`; wynik `FAIL`. Odtworzenie: wejście i kierunek z [records.tsv](records.tsv), polecenie macierzy w [raporcie](README.md); szczegółowe linie dowodu w [checks.tsv](checks.tsv).

## R-SURFACE

Kategoria: **plugin realizacji**. Kontrole: **0**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: wybrana realizacja → wymagany zbiór tekstów.

Brakuje wymaganego tekstu/wariantu albo pojawia się niedozwolony. Konkretne komunikaty w surowym logu; bez dowodu nie przypisano każdego do morfologii.

Źródło: [`tests/matrix/checks.go:629`](../../../tests/matrix/checks.go#L629).

## A-SEMANTICS

Kategoria: **plugin analizy**. Kontrole: **0**. Pewność: **ograniczona**. Pierwsza potwierdzona granica: parse → wymagane Meaning/context.

Nie został zachowany pełny wymagany graf/kontekst. Szczegółowe komunikaty i obserwacje pozwalają odtworzyć dalszą diagnozę.

Źródło: [`internal/semantic/compose.go:455`](../../../internal/semantic/compose.go#L455).

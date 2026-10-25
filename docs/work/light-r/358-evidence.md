# R4 odbiór uczestników wypowiedzi

Issue [#358](https://github.com/clickbytes/golingua/issues/358), base 232705cf71b86561ec07034fa302b31d5f465044 + working tree. Ocena agenta: odebrano lokalnie; końcowy PR i rzeczywisty CI PR/main zapisuje issue. Akceptacja językowa właściciela NOT_EVALUATED.

## Delegacja i granice

Jeden wykonawca Luna Low pracował szeregowo; drugi Luna Low prowadził ograniczone przeglądy bez zapisu. Każdy brief wskazywał issue, base, dozwolone pliki, wynik, zakazy, budżet i przekazanie. Pomocnicy nie pisali do GitHub. Koordynator redagował i integrował ograniczone poprawki wskazane w review oraz wykonał odbiór.

| Zlecenie | Zamknięty zakres |
| --- | --- |
| A | runtime/source.go i semantic/compose.go: typed implicit role, budżety, exact derivation/VP/span/frame proof |
| B | construction family/variant oraz runtime slot execution, realizer Execute/Linearize: licencjonowane pominięcie NP, walidacja context |
| C | wyłącznie EN: trzy zaimki, morfologia, personal family, analiza i target |
| D | wyłącznie kontrakty porównania anchorów i sześć wymagań straty count ty/wy→you |
| E | wyłącznie PL source/data: cztery zaimki, rodziny explicit/pro-drop, exact bindings i dowody |
| F | wyłącznie PL target: cechy z warunków wariantu, number/gender alternatives, context refusal |
| G | jeden focused plik participants_test.go: niezależne I, prawdziwe source spans, negatywne konteksty i dowody |
| H | jeden test composition_test.go: EN→PL bez pozornego założenia third, PL→EN nadal warianty rodzajnika |

OpenCode: użyto dziennej kontroli tego samego obrazu v2.0.15 i środowiska z [rejestru](README.md). Worker nie spełnia bramki agent listing / trzech odmów uprawnień. Implementacja przez OpenCode NOT RUN; bez --auto i płatnego fallbacku. Użyto dopuszczonego Luna Low.

## Wykazane zachowanie

[38 rzeczywistych wywołań CLI](358-observations.jsonl) obejmuje oba kierunki seeing/hearing/reading, Ja/My/Ty/Wy, trzecioosobowy brak podmiotu, błędną zgodę i wcześniejsze kontrole nominalne. Widzę/Widzimy daje I/We; polskie ty/wy zachowuje znaną liczbę, a EN you pozostawia ją nieznaną. EN→PL you daje dwa warianty number, zaś homograficzne read zachowuje present/past i potrzebne polskie warianty rodzaju. Sześć niepoprawnych lub niewspieranych wejść nie daje tekstu. This is a dog zachowuje jawną stratę deixis zamiast udawać pełne rozróżnienie odległości.

Nowe focused kontrole wykonują produkcyjne adaptery: autorskie I .001/.002 oraz .009.past, rzeczywisty EN you, wszystkie cztery PL pro-drop i jawne zaimki. Potwierdzają brak mutacji I, dokładny span VP i brak fikcyjnego tokenu zaimka, known speaker→jeden OK target, unknown you→dwa warianty, źródłowe Czytałam list→wyłącznie żeńska forma bez założenia gender. Duplicate/stale implicit proof daje typed invalid. Sprzeczne, niepasujące i stale konteksty obu targetów są jawnie odrzucane bez tekstu. Focused kontrole i osiem kontrastów composition PASS.

Konstrukcja explicit wymaga NP; pro-drop ma pominięty opcjonalny NP i capability. Próba Widzę psa wykazała potrzebę tego rozdziału, nie attestation na filtrowanym lesie. Wszystkie role pozostają w Meaning. CORE i schemat I nie dostały osoby gramatycznej ani rodzaju; kontekst używa istniejących participant anchors. Generyczny nominalny NP wyznacza third przez gramatykę, bez założenia o świecie.

Review i pełna macierz wykryły regresje dwóch podłączeń; przywrócono dawną trasę third-person be w EN i odrzucono personal families przy nominalnym target subject. Dodatkowe ograniczone integracje koordynatora obejmują binding ogólnej rodziny do ramy przed walidacją implicit role, walidację sparowanego grafu, poprawkę I See→I see, kanoniczną formę bazową EN you bez count inference i typed refusal przed agregacją. Referent alternatives są teraz liczone po CandidatePolicy, tak jak event alternatives; odfiltrowane zaimki nie tworzą fałszywego ambiguous.

## Kontrole i ograniczenia

Gofmt, go vet ./..., build cmd/golingua, --help/--version i diff check PASS. Pełna obserwacyjna macierz: polecenie GOLINGUA_MATRIX_SHA=232705cf71b86561ec07034fa302b31d5f465044+working-tree go test ./tests/... -count=1 -v; exit1. Końcowe liczności: 187 FAIL, 858 GAP_ANALYZE, 1080 GAP_I, 410 GAP_REALIZE, 465 PASS. Jedyna dodatkowa niezaliczona kontrola to dotychczasowy SVO→OVS GAP_ANALYZE. Log: /tmp/golingua-light-r/358-final-matrix.log; po merge surowy log publikuje workflow main.

Dziewięć dawnych PASS integracji EN→PL zmienia się na FAIL: A1.1-AC-101/103/111/117/125/127/129, A1.2-NEG-228, A1.8-NEG-839. Porównano stare i nowe obserwacje: każda ma nadal jeden taki sam niekompletny source candidate z PARTIAL_PARSE_EVIDENCE / UNRESOLVED_ROLE_SENSE, tym samym grafem i tekstem. To ujawnione fałszywe PASS: pozorne target person=third maskowało istniejący niekompletny source. Nie przywrócono tego założenia ani nie poluzowano macierzy. Engine nadal propaguje niekompletność źródła; target-only jest OK. Cztery EN→PL composition wymagają teraz mocniejszego OK, bez założeń, strat i diagnostyk, bo ich kompletne źródła nie mają tego braku.

Braki książka/work i child pozostają widoczne; R6 zajmuje się child. Nie dodano drinking, future, personal past seeing/hearing, dowolnego trzecioosobowego pro-drop ani rozstrzygania tożsamości. OVS pozostaje istniejącym GAP. Kontekst z innymi lokalnymi ID nie jest niejawnie przepinany przez nową osobową realizację.

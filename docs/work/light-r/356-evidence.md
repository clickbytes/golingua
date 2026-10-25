# R2 odbiór liczby i form referentów

Base 50688307f7fd83df32264a5c0d6b554123c52ddd + working tree. Ocena agenta lokalnie odebrana, właściciel NOT_EVALUATED. Końcowe merge/CI: [#356](https://github.com/clickbytes/golingua/issues/356).

Luna Low był jedynym writerem w zakresie przydzielonych danych/adapterów i focused testu. Koordynator doprecyzował obowiązkowy kierunek EN proximal → PL to z deixis loss i skierował rozróżnienie ten/to do istniejących deklaratywnych variant.conditions zamiast dodatkowego source forest helpera. Koordynator usunął uboczną, równoważną redakcję dwóch warunków source po przekazaniu; brak zmiany semantyki.

[11 rzeczywistych wywołań CLI](356-observations.jsonl): This window/apple is black → To okno/jabłko jest czarne (incomplete z jawną stratą deixis); odwrotnie oba warianty This/That z jawnym założeniem reference. The mother read the letter → Matka czytała list; odwrotnie past read z wariantami rodzajnika. The teacher sees the windows/apples → Nauczyciel widzi okna/jabłka; polskie okna/jabłka zachowują plural. Okno pozostaje singular. Brak samodzielnej decyzji o określoności, tożsamości ani płci biologicznej.

Focused testy używają niezależnych I identification.029/030.en i time.021.en, sprawdzają analizę one/many w ordinary clauses, realizację, produkcyjną integrację i odrzucenie błędnego ten/to gender. Ponowny focused przebieg po redakcji koordynatora obejmuje też R1. Gofmt, go vet ./..., build, --help/--version i diff check PASS.

Pełne `GOLINGUA_MATRIX_SHA=50688307f7fd83df32264a5c0d6b554123c52ddd+working-tree go test ./tests/... -count=1 -v`: exit1, 330 PASS,243 FAIL,920 GAP_ANALYZE,427 GAP_REALIZE,1080 GAP_I. Brak utraty wcześniejszego PASS względem R1. Wzrost FAIL o2 wynika z wykonywania wcześniej brakujących etapów; nie zmieniano expected. Jedyny niezaliczony dodatkowy kontrast to istniejący SVO→OVS PL GAP_ANALYZE. Log /tmp/golingua-light-r/356-matrix.log; po merge surowe wykonanie publikuje workflow obserwacyjny.

Granice: pluralna atrybucja, inne przymiotniki, nowe domeny i rozszerzenie I poza zakresem. Rodzajniki plural i pełne warianty reading pozostają w R5. Nowe formy są autorskimi danymi agenta, nie akceptacją językową właściciela.

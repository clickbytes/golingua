# R1 odbiór negacji PL

Base e41b1d124fdd4b177c720964b88fc56a7493a664 + working tree, 2026-10-25. Ocena agenta: zakres lokalnie odebrany. Akceptacja językowa właściciela NOT_EVALUATED. Końcowe merge/CI: [issue #355](https://github.com/clickbytes/golingua/issues/355).

Mały wykonawca Luna Low zmienił languages/pl/valency.go i tests/languages/pl/A1/negation_case_test.go. Pierwszy test z psem nie rozróżniał acc/gen; koordynator odrzucił ten dowód jako niewystarczający i zlecił jedną korektę na książce, ptakach oraz na+acc. Nie zatwierdzono pending book.work ani nie zmieniono expected istniejącego korpusu.

Rzeczywiste CLI i argumenty/exit codes: [obserwacje](355-observations.jsonl). Analiza Nauczyciel nie widzi książki zachowuje stimulus book one, present i zakres negacji. Powstają cztery warianty rodzajników EN, wynik ambiguous. The teacher does not see the book daje Nauczyciel nie widzi książki (incomplete z jawną niepełnością źródłowego sensu). Matka nie słyszy ptaków daje warianty EN z the birds; niezależne I hearing(mother,bird many) daje Matka nie słyszy ptaków. Dodatnie widzi książkę i ujemne nie patrzy na książkę zachowują biernik.

The mother does not hear birds pozostaje unsupported przy analizie EN bare plural. To istniejący odrębny brak trasy NP, przeznaczony do R5, nie błąd naprawy przypadka PL. Nie zadeklarowano zaliczenia tego wejścia. Status incomplete nie oznacza zaakceptowania całego znaczenia ani rozstrzygnięcia book.work.

Wykonane: gofmt zmienionych plików, go vet ./..., go build -o /tmp/golingua-light-r/golingua ./cmd/golingua, --help, --version i git diff --check: PASS. Focused TestNegativeCaseRealizationContrasts oraz TestNegativeSeeingBookAnalysisKeepsSingularStimulus: PASS (obejmują niezależne I, realizację i analizę).

Pełne `GOLINGUA_MATRIX_SHA=e41b1d124fdd4b177c720964b88fc56a7493a664+working-tree go test ./tests/... -count=1 -v`: exit 1. 3000 kontroli: PASS 306, FAIL 241, GAP_ANALYZE 940, GAP_REALIZE 433, GAP_I 1080. Brak utraty wcześniejszego PASS w zarejestrowanych 3000 kontrolach. Nowe focused testy są obecne w pełnym logu. Nie zmieniono 1000 rekordów ani ich wymagań. Log lokalny /tmp/golingua-light-r/355-matrix.log; po merge surowy log publikuje istniejący workflow obserwacyjny.

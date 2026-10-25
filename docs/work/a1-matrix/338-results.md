# Odbiór A1-M2

Base `9bf866bc620ec1ad4ae994876f57bc47885aac71`, zmiany working tree obejrzane przed wykonaniem. Ocena techniczna koordynatora: PASS dla zakresu infrastruktury. Akceptacja językowa właściciela: `NOT_EVALUATED`.

Wykonano lokalnie: kontrola gofmt wszystkich śledzonych Go (brak niesformatowanych), `go vet ./...`, `go build -o /tmp/golingua-matrix-expansion/golingua ./cmd/golingua`, `--help`, `--version` (`0.1.0-alpha.1`) oraz `git diff --check`: PASS.

`GOLINGUA_MATRIX_SHA=9bf866bc620ec1ad4ae994876f57bc47885aac71+working-tree go test ./tests/... -count=1 -v` zakończyło się exit 1. Raporter, wykonany na tym samym logu z przekazanym rzeczywistym exit code, wykazał 24 rekordy / 72 kontrole: 36 PASS, 36 SPEC_PENDING, zero FAIL/GAP/INFRA/NOT RUN w tych rekordach oraz 147 dodatkowych kontroli PASS, z jednoznacznym pakietem i nazwą. Istniejące braki pozostały widoczne.

Kontrole mechanizmu obejmują niezależność wariantu dozwolonego od wymaganego, wykrywanie brakującego wymaganego tekstu i błędnej wartości warunku reference. Sprawdzono rozpoznawanie A1.1, B1 i C2.1; nie dodano fikcyjnych zdań produktu.

Reprodukcja: ustawić GOLINGUA_MATRIX_SHA na aktualne SHA (z dopiskiem +working-tree, jeśli potrzebny), uruchomić pełne polecenie powyżej do pliku, następnie `python3 -B scripts/report-language-matrix.py --sha <SHA> --exit-code <rzeczywisty-kod> --log <plik>`. Surowe lokalne pliki: `/tmp/golingua-matrix-expansion/m2-full.log` i `m2-report.md`; trwały raport i log publikuje workflow main.

Raportowane tagi to deklarowany zakres przypadku, nie śledzenie wykonanych konstrukcji. Zmiana nie dodaje nowej gramatyki ani procentowej bramki i nie zmienia technicznego CI. Stan PR i main CI zapisany w issue #338 po scaleniu.

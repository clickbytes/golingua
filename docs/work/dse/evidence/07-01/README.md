# Evidence #394 — typowane wiązania leksykalne

## Zakres i przypięte wykonanie

Obserwacje wykonano na zamrożonym binarium `/tmp/golingua-dse/bin/golingua-394-final`, SHA-256 `dd8e95c102ffcbb339cc036c517ab81fb90df4163d080648565b4297bec6991b`. Plik `reports/394-technical-observation.json` zapisuje rzeczywiste polecenia `gofmt -l`, `go vet ./...`, `go build -o ... ./cmd/golingua`, `--help` i `--version`; wszystkie miały kod 0. Nie uruchamiano `go test`.

Źródła CLI, schematu i dane są wskazane w `reports/394-source-hashes.json`; raport obejmuje 10 ścieżek i potwierdza ich zgodność z zamrożonym źródłem użytym do zbudowania binarium. `reports/394-schema-observation.json` zawiera 36 walidacji przykładowych wejść względem schematów; wszystkie zwróciły pustą listę błędów. To kontrole techniczne i strukturalne, nie akceptacja językowa.

## Obserwacje CLI

`outputs/main/observations.jsonl` zawiera 41 wywołań na wejściach z `inputs/independent/`: 17 z kodem 0, 11 z kodem 2, 7 z kodem 5, 5 z kodem 3 i 1 z kodem 6. Surowe odpowiedzi są skopiowane do `outputs/main/` pod nazwami wskazanymi przez `stdout_file`. Log zachowuje pierwotne ścieżki `/tmp/golingua-dse/394-independent-inputs`; ich lokalnymi odpowiednikami są pliki w `inputs/independent/`.

Te kontrole obejmują pakiet z jedną lub wieloma pasującymi kandydaturami, jawny wybór `binding_index`, zmienione nazwy symboli, otwarte parametry, brak/nieznane metadane, zły typ, nieznaną rolę lub port, niedozwolony profil, wstrzyknięcie zmiennej i przypięty kontekst. `ambiguous` nie wybiera samodzielnie celu. Przypadki błędów są zachowane jako odmowy, a nie pomijane.

`outputs/pilot/observations.jsonl` zawiera odrębną serię 20 wywołań na produkcyjnych przykładach PL/EN: 19 z kodem 0 i 1 z kodem 2. Surowe odpowiedzi i użyte wejścia pochodzą z `outputs/pilot/`; pliki przykładowe są również dostępne w [`data/semantic/examples/lexical-bindings/`](../../../../../data/semantic/examples/lexical-bindings/README.md). Pilotaż zwraca cztery percepcyjne symbole PL/EN oraz dwa techniczne cele `link` EN. Dla pary percepcyjnej PL/EN zwrócono ten sam typowany atom; `see` i `hear` pozostają odrębnymi symbolami. Zamiana mapowania ról zmienia atom. Kandydat dwukrawędziowy ma wspólny pośredni binder `Exists`; jego kontrola dokumentu i niezależne `verify-goal` zwracają kod 0. Referencja zdarzenia podana jako bodziec typu `Event` jest odrzucona (`ROLE_REFERENCE_TYPE_MISMATCH`, kod 2).

Zestaw nie mierzy poprawności tłumaczeń ani parsowania. `pilot.r` jest neutralną relacją uporządkowaną. `link.direct` i `link.indirect` to osobne techniczne etykiety zdefiniowane w danych pilota, nie ogólna deklaracja znaczenia angielskiego słowa „link”. Weryfikacja celu ogranicza się do jawnego profilu strukturalnego i nie ocenia prawdziwości faktów.

## Odtwarzalny przepływ z dokumentacji

Polecenia z [sekcji CLI](../../../../cli.md) zostały wykonane zamrożonym binarium. `docflow/symbol/` zachowuje `inspect`, `validate`, instancjację symbolu oraz wyodrębniony dokument i kontrolę algebraiczną; każdy z pięciu wyników ma status `ok`. `docflow/goal/` zachowuje instancjację celu, wyodrębniony cel/kandydata/certyfikat, kontrolę kandydata i `verify-goal`; instancjacja i obie kontrole mają status `ok`, a weryfikacja zwraca `certificate_valid: true`, `goal_satisfied: true` i `complete: true`. Zakres w odpowiedzi pozostaje `structural_profile_only`, nie jest twierdzeniem o świecie ani akceptacją językową.

## Zachowanie istniejącego katalogu

`reports/394-final-legacy-observation.json` zapisuje dwa osobne porównania ze stanem bazowym: `catalog --summary --json` i `algebra catalog export --json`. Oba końcowe wyjścia były bajtowo identyczne z bazą. Zachowano po jednym końcowym wyjściu w `reports/catalog-summary-final.json` i `reports/catalog-export-final.json`; SHA-256 końcowego strumienia eksportu to `54f27ad1ea12943e0ad4d115156edca952e03ab5e840c4af6336944066f818d1`, taki sam jak SHA wyniku bazowego. Nie duplikowano dużego strumienia bazowego.

Oddzielny plik `reports/394-pilot-root-semantic-observation.json` potwierdza, że pakiety pilota powstały po zbudowaniu binarium i nie zmieniły jego SHA-256. `reports/394-semantic-observation.json` zapisuje relację bezpośrednią, ścieżkę pośrednią, odrębne symbole dla percepcji wzrokowej i słuchowej oraz wynik inspekcji `metadata_verified: false`.

Pochodzenie czterech wcześniejszych wiązań CORE jest zachowane jako historia w danych, lecz nie jest użyte do automatycznego aktywowania nowych celów. Ocena nowych wiązań jest ograniczona do pilota; akceptacja językowa właściciela pozostaje **NOT EVALUATED**. PR i CI `main` są oczekujące w chwili utworzenia tego evidence.

Koordynator niezależnie uruchomił trzy literalne bloki z `docs/cli.md`: wszystkie 8 wywołań GoLingua zakończyło się kodem 0, w tym osobne `check` i `verify-goal`. [Zapis wykonania](root-docflow/observation.json), [dokładne wywołania](root-docflow/commands.jsonl).

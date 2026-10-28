# DSE-09-01: niezależność od domeny

Baza audytu: `3643bdcd69298f85fff2826431502743ad4c3024`. Binarka zbudowana przed przygotowaniem danych: SHA-256 `2c792432f78f883cf5e546b4096b46e54b1b1f373bbc989342120b7a65feb4c5`. Materiał kontrolny i mapy identyfikatorów znajdują się w `data/semantic/acceptance/independence/`; komendy, exit codes, dane wyjściowe i certyfikaty w `evidence/09-01/`. To propozycje neutralnych danych, nie akceptacja językowa właściciela.

## Przegląd kodu

- `internal/metamodel/validate.go` i `internal/metamodel/registry.go`: ładowanie i walidacja deklaratywnych pakietów i sygnatur. Nazwy i etykiety nie są wyborem algorytmu.
- `internal/algebra/`: typowana reprezentacja operatorów i jawnych symboli; brak importu modułów PL/EN do algebry.
- `internal/synthesis/construct.go`, `search.go`, `replay.go` i `goal_verify.go`: konstrukcja, ograniczone wyszukiwanie oraz niezależne odtworzenie i weryfikacja korzystają z celu, sygnatur i ogólnych operatorów, bez znalezionej tablicy wyników ani dispatchu po nazwach domenowych.
- `internal/reason/algebra_plan.go` i `algebra_evaluate.go`: kompilacja według rodzaju termu i deklarowanych ról oraz ewaluacja względem jawnych faktów. Osobny starszy adapter MeaningGraph w `internal/reason/match.go` korzysta z CORE i dostarczonego mapowania relacji; nie jest implementacją tego przepływu algebry. `internal/semantic/compose.go` zawiera starszą językową obsługę domyślnego konceptu `person`, poza badaną ścieżką syntezy i ewaluacji algebry.

Przegląd nie dowodzi braku każdej możliwej zależności w całym produkcie. Testowany jest ograniczony profil konstrukcji i dodatnich faktów; negacja, ogólne kwantyfikatory i inne nieobsługiwane operatory nie są tym dowodem objęte. Synteza ani certyfikat nie dowodzą prawdziwości faktów w świecie.

## Wyniki

Dwa pakiety opisują węzły tras oraz jednostki przekazania. Każdy dostarcza własne symbole, role, nominalny SortID i fakty po zbudowaniu tej samej binarki. Konstruktor wygenerował anonimowy wynik dla dwóch i trzech kroków z restrykcją; wszystkie osiem przypadków (oryginały i przemianowania) otrzymało `ok`, `fully_validated=true`, niezależne `verify-goal: ok` i ewaluację `ok`. W obu domenach łańcuch dwustopniowy zwrócił dwa wyniki, a trzystopniowy jeden; po przemianowaniu wszystkie pary referentów są równe według map `id-map.json`. Odrębne węzły nie zostały scalone. Wyniki są ograniczone do dostarczonych faktów i tych typów celów.

Przemianowano symbole, role, własne sorty, porty, referenty i identyfikatory faktów; etykiety usunięto. Dla wszystkich ośmiu przypadków eksport w osobnych procesach z `--no-cache` zachował fingerprint znaczenia względem zwykłego eksportu. Reprezentacja artefaktu może się różnić, ponieważ zmienia się współdzielenie węzłów. Po próbach SHA-256 binarki pozostał taki sam.

Nieznany port zwrócił `UNKNOWN_RELATION_ROLE`; weryfikacja konstruktu bez wymaganej restrykcji zwróciła `REFINEMENT_CONSTRAINTS_CHANGED`, a podmieniony certyfikat `GOAL_FINGERPRINT_MISMATCH`. Cel dopuszczający dwa i trzy kroki zwrócił dwa kandydaty bez wskazania jednego jako jedynego rozwiązania. Limit jednej jednostki pracy zwrócił `resource_limit` i `incomplete`. Nie dowodzi to ogólnej kompletności wyszukiwania ani obsługi dowolnych typów ograniczeń.

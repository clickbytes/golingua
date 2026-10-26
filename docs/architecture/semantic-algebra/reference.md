# Kontekst i jawne referenty

`request_context` to opcjonalny, lokalny dla jednego żądania opis referentów,
wzmianek, jawnych powiązań tożsamościowych oraz warunków semantycznych. Nie jest
pamięcią rozmowy ani ontologią. Pole może wystąpić w dokumencie algebry lub
specyfikacji celu; brak pola zachowuje dotychczasowe dokumenty i ich fingerprinty.
Pełny kształt określa [schemat kontekstu](../../../schemas/semantic/context.schema.json),
a wspólny typowany term [schemat wyrażeń](../../../schemas/semantic/expression.schema.json).

## Referenty i wzmianki

Każdy referent ma własne `id` i typ. Opcjonalny `concept_id` jest klasyfikacją;
ten sam koncept nie łączy identyfikatorów. `speaker_referent_id` i
`addressee_referent_id` wskazują wyłącznie jawne referenty w tym kontekście.
Identyfikator referencji użyty przez wyrażenie musi mieć zgodny typ i być
zadeklarowany w kontekście. Identyfikatory binderów (`Var`), parametrów,
referencji i dziur pozostają osobnymi przestrzeniami.

Wzmianka z zadeklarowanym `referent_id` wymaga pasującego rekordu
`resolution_evidence`. Bez celu wzmianka pozostaje `unresolved`; sprzeczne,
poparte cele pozostają `ambiguous` z listą kandydatów. Powtórne jawne wskazania
tego samego referenta zachowują współreferencję. Checker nie wybiera celu na
podstawie zgodnego konceptu, roli mówiącego ani samego podobieństwa typów.
`concept_id` musi używać istniejącej postaci, na przykład
`concept:category.k`; nie jest nowym typem i nie tworzy równoważności.
Pochodzenie używa standardowego `provenance.SourceRef` (`id`, `version`,
`license` oraz dokładnie jedno z pól `path` i `uri`, z opcjonalnymi polami
autora i sumy kontrolnej), a nie rekordu źródła faktu. `SourceRef` dokumentuje
pochodzenie, ale sam nie uwierzytelnia źródła.

## Warunki

Warunek ma jawny `id`, `kind` (`assumption`, `presupposition` albo `loss`),
nieprzezroczysty `code`, listę `target_referent_ids` oraz opcjonalny typowany
`expression` lub opisowy `detail`. Wyrażenia sprawdza się względem faktycznie
wybranych deklaracji. Checker zachowuje je i raportuje; nie wykonuje ich,
nie uznaje za fakty, nie rozstrzyga istnienia ani unikalności i nie zmienia
statusu referencji. Źródłowe dane parsera nie są wymagane do przeniesienia
jawnie podanego kontekstu.

## CLI i zachowanie przepływu

```sh
go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/resolved-document.json --json

go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/unresolved-document.json --json

go run ./cmd/golingua algebra context check \
  --input data/semantic/examples/reference/conflicting-document.json --json

go run ./cmd/golingua algebra synthesize --mode construct \
  --metadata data/semantic/examples/reference/metadata-zero-arity.json \
  --goal data/semantic/examples/reference/description-goal.json --json
```

`algebra context check` wymaga jawnego `request_context`. Dokument
self-contained można sprawdzić bez pakietu; przy `--metadata` dokument wejściowy
nie może powtarzać `sorts` ani `signatures`. Wynik `status: "ok"` oznacza
poprawność struktury i typów. Osobna `request_context_assessment` pokazuje
`resolved`, `unresolved` lub `ambiguous`; `ok` nie oznacza, że wszystkie wzmianki
zostały rozstrzygnięte ani że warunki są prawdziwe.

Kontekst jest przenoszony przez budowanie, normalizację, konstrukcję i
wyszukiwanie celów, certyfikaty, eksport, odtworzenie, porównanie oraz ocenę
zapytania. Fingerprint wiąże cały payload, a nie tylko `id`. Eksport bierze
kontekst z dokumentu; odtworzenie wymaga osobnego, oczekiwanego
`request_context` w żądaniu i odrzuca brak lub różną treść. Porównanie wyrażeń
wymaga zgodnego kontekstu obu dokumentów i środowiska. Zmiana treści przy tym
samym identyfikatorze nie jest zgodnością. Fingerprint kanonizuje nazwane
rekordy według ID; normalizacja zachowuje kolejność warunków źródłowych, a
eksport może uporządkować nazwane rekordy według ID. Porównuj więc zawartość
rekordów o tym samym ID, nie surową kolejność tablic. Kolejność argumentów
`Atom` i odpowiedzi `Ask` pozostaje semantyczna i nie jest przestawiana.

Przykład opisu w `description-goal.json` wiąże port `y` jawnie z `Ref A`, ale
pozostawia port `x` otwarty. Wynik ma postać `λx.r(x, A)`; obecność referentów
`A` i `B` tego samego typu i klasy nie pozwala wybrać `B` ani domknąć `x`.
Warunki opisu pozostają warunkami, nie świadkami. Ten przykład używa
`metadata-zero-arity.json`, bo zawiera symbol `t0` dla warunku z atomem o
zerowej arności; zwykły pakiet `metadata.json` go nie deklaruje. Puste
`args: []` jest zachowywane w sprawdzaniu i serializacji.

Nie ma tu heurystyk koreferencji, tożsamości płciowej, pamięci między żądaniami
ani automatycznego wnioskowania z morfologii.

Ograniczenia są skończone: maksymalnie 1 MiB na kontekst, 512 referentów,
512 wzmianek, 1024 rekordy dowodowe, 128 warunków i 4096 węzłów wyrażeń
warunków. Przekroczenie zwraca `resource_limit`. Żaden z tych limitów nie
implikuje kompletności rozwiązania ani prawdy o świecie.

## Przykłady i obserwacje

Neutralne, autorskie wejścia znajdują się w
[`data/semantic/examples/reference/`](../../../data/semantic/examples/reference/).
Rzeczywiste wyjścia, argumenty, kody zakończenia i skróty lokalnej kontroli są
w [evidence #392](../../work/dse/evidence/06-02/README.md). Przykłady są
diagnostyczne; ich URN-y i `SourceRef` nie są zewnętrznym dowodem ani akceptacją
właściciela. Ocena językowa właściciela pozostaje `NOT_EVALUATED`.

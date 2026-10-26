# Atrybuty referencji i zakres modalny

To neutralne przykłady typowanego IR. Symbole `r`, `s`, `p`, `q`, `t` i `e2`
mają wyłącznie jawne sygnatury; ich etykiety nie kodują reguł. Pakiet
`metadata.json` nie pochodzi z języka naturalnego.

## Zachowanie pól

`past.json` i `present.json` przypisują ramę czasu do referencji `ev1` typu
Event oraz wiążą ją z referencją `T` typu Interval. `one.json` i `many.json`
przypisują wielość do referencji `C` typu Entity. `unknown-time.json` i
`unknown-number.json` celowo pomijają te pola. Brak wartości pozostaje
nieznany; silnik nie wyprowadza czasu z gramatyki ani liczby z kształtu
zapytania. `context-only-reference-interval.json` zawiera `T` wyłącznie w
jawnym kontekście żądania.

`modal.json`, `not-modal.json` i `modal-not.json` odróżniają strukturę Modal od
Not oraz pokazują, że zmiana kolejności operatorów zmienia zakres. Profil
modalny jest jawnie przekazanym, nieinterpretowanym identyfikatorem. Sam wpis
profilu nie dodaje praw modalnych.

## Odtworzenie

Z poziomu katalogu repozytorium można sprawdzić reprezentację i zbudować
przykładowy term:

```sh
golingua algebra check --input data/semantic/examples/attributes/past.json --json
golingua algebra build --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/modal-build.json --json
golingua algebra evaluate --metadata data/semantic/examples/attributes/metadata.json --expression data/semantic/examples/attributes/one.json --facts data/semantic/examples/attributes/facts-with-event.json --select data/semantic/examples/attributes/select-empty.json --context-id eval-neutral --profile positive-facts-v1 --json
golingua algebra evaluate --metadata data/semantic/examples/attributes/metadata.json --expression data/semantic/examples/attributes/past.json --facts data/semantic/examples/attributes/facts-with-event.json --select data/semantic/examples/attributes/select-empty.json --context-id eval-neutral --profile positive-facts-v1 --json
```

Pierwsze trzy polecenia kończą się statusem `ok`; ewaluacja `past.json` zwraca
`unsupported`, ponieważ obecny profil ewaluacji nie obsługuje względnej ramy
czasu. Analogicznie `many.json` i Modal są poprawnymi reprezentacjami, ale nie
są obsługiwane przez obecny ewaluator. Nie jest to wnioskowanie o prawdziwości
ani o dokładnej liczbie obiektów.

Żądania eksportu, artefakty do inspekcji i żądania replay są odrębnymi,
jawnymi wejściami. Poniższe przykłady pokazują zachowanie pól przy eksporcie i
odtworzeniu:

```sh
golingua algebra export --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/past-export-request.json --json
golingua algebra inspect --input data/semantic/examples/attributes/past-artifact-input.json --json
golingua algebra replay --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/past-replay-request.json --json
golingua algebra inspect --input data/semantic/examples/attributes/context-only-reference-interval-artifact-input.json --json
golingua algebra replay --metadata data/semantic/examples/attributes/metadata.json --input data/semantic/examples/attributes/context-only-reference-interval-replay-request.json --json
```

Te pięć poleceń zakończyło się kodem 0. Inspekcja obu artefaktów zwraca
`metadata_verified: false`; replay sprawdza je przy jawnie przekazanym pakiecie
metadanych.

Pełne wejścia i surowe odpowiedzi obserwacji #393 znajdują się w
[`docs/work/dse/evidence/06-03/`](../../../../docs/work/dse/evidence/06-03/README.md).
Status akceptacji językowej właściciela: `NOT_EVALUATED`.

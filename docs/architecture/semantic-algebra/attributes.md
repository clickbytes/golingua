# Atrybuty referencji i Modal

Ta część kontraktu opisuje wersjonowany, typowany zapis atrybutów jawnych
referencji oraz ograniczony operator `Modal`. Nie dodaje ontologii ani
domyślnych faktów.

## Atrybuty

Pola `temporal_frame` i `reference_time_id` dotyczą wyłącznie jawnej
deklaracji referencji typu `Event`. Obsługiwane wartości `temporal_frame` to
`past` i `present`; `reference_time_id` wskazuje jawnie zadeklarowany obiekt
typu `Interval`. Pole może pochodzić z deklaracji dokumentu albo z kontekstu
żądania. Gdy oba miejsca podają tę samą referencję, znane wartości muszą być
zgodne. Deklaracja celu nie może wisieć ani wskazywać referencji innego typu.

Pole `multiplicity` dotyczy tylko referencji typu `Entity` lub `Nominal` i
przyjmuje `one` albo `many`. Te wartości są informacją o jawnie wybranej
referencji. `many` nie oznacza dokładnej liczby ani kompletnego zbioru.

Pola są opcjonalne. Ich brak zachowuje wartość nieznaną; nie jest zastępowany
przez czas bieżący, domysł gramatyczny ani liczbę argumentów. Atrybutów nie
można przypisywać deklaracjom Param ani binderom `Ask`/lambda. Jawny kontekst
jest osobnym wejściem żądania, a zgodność jego wpisów z dokumentem jest
sprawdzana przed użyciem.

Inspekcja artefaktu zwraca jego strukturalne deklaracje, ale oznacza
`metadata_verified: false`. Same wartości atrybutów nie uwierzytelniają
referencji ani nie dowodzą jej tożsamości. Eksport i odtworzenie weryfikują
artefakt względem jawnie wybranych metadanych oraz kontekstu.

## Modal

Wersjonowany term `Modal` ma jawny rodzaj (`possible`, `necessary` lub
`likely`), opcjonalny identyfikator profilu i ciało typu Prop. Rodzaj i profil
są danymi strukturalnymi. Obecny profil nie definiuje praw modalnych ani nie
wyprowadza semantycznych konsekwencji z identyfikatora profilu. `Not` i
`Modal` są różnymi konstruktorami, a ich zagnieżdżenie zachowuje kolejność i
zakres.

## Granica ewaluacji

Typowanie, budowa, normalizacja, eksport i odtworzenie zachowują atrybuty.
Ograniczony ewaluator faktów obsługuje jedynie dotychczasowy profil atomowego
dodatniego wsparcia. Jawne `one`, brak atrybutów i dopasowane fakty mogą być
przetwarzane w tym profilu. Względne ramy czasu, `many` i Modal są obecnie
zwracane jako `unsupported`; nie są zamieniane na `false`. Brak atrybutu nadal
pozostaje nieznany. W pokazanym przykładzie `Not(Modal(P))` i
`Modal(Not(P))` mają różne postacie normalne, a profil nie licencjonuje
zamiany ich zakresów; to porównanie zwraca `undetermined`. Identyczne
wyrażenia mogą być porównywalne bez takiej reguły.

Adapter Meaning przyjmuje już utworzony typowany MeaningGraph i jawne
mapowanie jego znanych pól na referencje. Nie parsuje języka, nie rozwiązuje
świata i nie tworzy faktów.

Neutralne przykłady i rzeczywiste lokalne obserwacje są opisane w
[`data/semantic/examples/attributes/`](../../../data/semantic/examples/attributes/README.md)
oraz [evidence #393](../../work/dse/evidence/06-03/README.md). Techniczna
obserwacja wykonania i ocena agenta są odrębne od akceptacji językowej
właściciela, która pozostaje `NOT_EVALUATED`.

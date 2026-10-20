# Warianty i linearyzacja

`Variant` wybiera uporządkowaną sekwencję tokenów rodziny: slotów NP/VP/PP i
jawnych terminali. Wariant ma własny `construction:` ID, `FamilyID`, jawne
warunki, markedness, priorytet i policy kierunku (`both`, `realize` albo
`analyze`). `CompileVariant` wiąże go z jedną rodziną i tworzy bounded plan z
dozwolonymi tokenami brzegowymi dla późniejszego parsera.

Każdy wymagany slot rodziny występuje w sekwencji dokładnie raz; opcjonalny może
wystąpić zero albo jeden raz. Brak wymaganej pozycji, nieznany slot,
zduplikowany slot token, błędny terminal, nieznany warunek lub nieprawidłowe
provenance zatrzymują kompilację. Literalny terminal może powtórzyć się w
sekwencji, jeśli ma odrębną funkcję (np. powtórzone słowo pomocnicze); kontrola
duplikatów dotyczy tożsamości slotów, nie samego tekstu. `VariantIndex` zwraca
wszystkie warianty pasujące do first/last tokenów i nie usuwa wieloznaczności.

`ValidateVariantForFrame` sprawdza osobno konkretną walencję: frame musi
licencjonować FamilyID, role family muszą istnieć w tym frame, a każda rola
wymagana przez frame musi być obecna w sekwencji wariantu. Jedna rodzina i jeden
variant mogą być sprawdzane wobec wielu sense frames bez kopiowania szablonu.

## Kolejność powierzchni a znaczenie

Sekwencja ustala kolejność powierzchniową. Nie zmienia `SemanticRole`,
`Meaning.Participant` ani `LexicalSense`; linearyzator nie przypisuje ról na
podstawie pozycji.

- W angielskim transitive SVO szablon może mapować slot podmiotu do
  `experiencer`, a dopełnienie do `theme`. W `The cat sees the dog` kot
  wypełnia `experiencer`, a pies `theme`; w `The dog sees the cat` role
  wypełniają odwrotne referenty. Rodzina zachowuje to samo slot→role mapowanie,
  a Meaning jawnie wskazuje uczestników.
- Polski OVS może uporządkować sloty obiektu, czasownika i podmiotu. Role
  pozostają przypisane z walencji; przypadek musi pozwalać odzyskać te role.
  Sam szyk OVS nie ustawia `topic`, `focus`, `given` ani `new`.

To są strukturalne ilustracje, nie zatwierdzone wejścia językowe ani goldy.
Jeśli InformationStructure nie została dostarczona, pozostaje unknown. Warunek
`information_structure` wymaga jawnej wartości; markedness i priorytet nie
uzupełniają jej automatycznie.

## Warunki

Warunki są zamkniętym unionem typowanych predykatów: na tym etapie feature
equality dla slotu albo jawna wartość InformationStructure dla semantic role.
Kod nie wykonuje tekstu z JSON. Nierozpoznany rodzaj warunku, puste wartości,
zły slot albo rola spoza rodziny są odrzucane. Zgodność feature value z profilem
języka należy do późniejszego kompilatora constraints.

Granice indeksu uwzględniają pominięcie opcjonalnych slotów na początku i końcu
sekwencji. Jeśli wariant zaczyna się opcjonalnym NP przed wymaganym VP, oba
tokeny mogą być pierwsze; oba indeksowane są dla bounded parsera. Kompilator
odrzuca sekwencję, która mogłaby całkowicie zniknąć.

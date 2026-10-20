# Walencja per Sense

`ValencyFrame` wiąże role semantyczne z konkretnym `SenseID`. Sam `ConceptID`
nie określa składni czasownika: angielskie `see` i `look at` mogą odnosić się do
podobnego zdarzenia, ale licencjonują inne ramy; polskie `widzieć` i `patrzeć
na` także różnią się rekcją. Każdy transfer wymaga zatwierdzonego równoważnego
`LexicalSense → CORE` oraz zatwierdzonego `ConstructionBinding` z dokładnym
`SenseID`, `ValencyFrameID`, `FamilyID` i `VariantID`. Frame przechowuje
`SourceRef` dla ręcznie opracowanej reguły; binding ma własne provenance.

## Role i argumenty

Rola semantyczna (`agent`, `theme`, `experiencer` itp.) nie jest funkcją
składniową (`subject`, `object`, `oblique`) ani przypadkiem. Frame zapisuje te
wymiary oddzielnie w `ArgumentSlot`; nie istnieje globalna reguła
`agent = subject` lub `theme = object`. `Case` i `Adposition` są jawnymi,
językowymi ograniczeniami argumentu, a ich brak nie uprawnia do wywnioskowania
rekcji.

Każdy slot ma jeden `Role` i jawne `Required`. Brak roli wymaganej zwraca
`incomplete/MISSING_REQUIRED_ROLE`; dodatkowa rola, której frame nie licencjonuje,
zwraca `unsupported/EXTRA_SEMANTIC_ROLE`. Opcjonalny slot może pozostać pusty.
Walidator niczego do `Meaning` nie dopisuje. Bieżący Meaning przypisuje daną
rolę najwyżej raz; dlatego schemat frame nie przyjmuje powtórzonych slotów dla
tej samej roli.

## Ograniczenia selekcyjne

`SelectionalConstraint` wymienia dozwolone `ConceptID` (włącznie z ich
potomkami w hierarchii CORE) lub rodzaje konceptów. Ograniczenie `hard` odrzuca
niezgodny cel jako `unsupported/HARD_SELECTIONAL_RESTRICTION`. Ograniczenie
`soft` zwraca `SoftConstraintMiss` w zaakceptowanym wyniku i nie odrzuca ramy.
Nieznany koncept pozostaje `unknown`; nie jest traktowany jako niezgodny typ.

## Licencjonowanie i indeks

Frame wymienia `AllowedFamilyIDs`. Indeks przyjmuje wyłącznie jawnie zatwierdzone
bindingi z prawidłowym provenance; binding musi wskazać istniejący frame o tym
samym `SenseID`, a jego family musi znajdować się na liście dozwolonych rodzin.
Brak frame'u, niezatwierdzony mapping, nieznany family lub błędne powiązanie
nie uruchamia fallbacku po `ConceptID`. `FramesBySenseID` zachowuje wszystkie
ramy sensu. `BindingsByFamilyID` zwraca osobne bindingi w porządku
`SenseID/FrameID/VariantID`; wspólna rodzina nie scala sensów ani nie usuwa
niejednoznaczności.

## Przykład ram

Poniższe opisy ilustrują rozdzielenie poziomów, nie stanowią zatwierdzonych
danych leksykalnych:

| Sense | Role semantyczne | Slot językowy |
| --- | --- | --- |
| EN `see` | experiencer, theme | experiencer: subject; theme: direct object |
| EN `look at` | experiencer, theme | experiencer: subject; theme: oblique z `at` |
| PL `widzieć` | experiencer, theme | experiencer: podmiot w mianowniku; theme: dopełnienie zgodne z zatwierdzoną ramą |
| PL `patrzeć na` | experiencer, theme | experiencer: podmiot w mianowniku; theme: oblique z `na` i jawnie licencjonowanym przypadkiem |

Bez zatwierdzonego sensu, semantic binding i construction binding wynik to
odmowa transferu, a nie zgadywanie najbliższej ramy. Walidator wymaga
konkretnego `Meaning.NodeID`, więc wiele pasujących zdarzeń nie jest
rozstrzyganych kolejnością.

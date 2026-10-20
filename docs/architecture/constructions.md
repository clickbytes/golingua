# ConstructionFamily i zapytania konstrukcji

`ConstructionFamily` to jeden językowy szablon strukturalny. Ma własny typed
`construction:` ID i nie zależy od `ConceptID` ani `SenseID`. Zatwierdzony
`valency.ConstructionBinding` łączy konkretny sens i jego `ValencyFrameID` z
rodziną oraz wariantem. Wielokrotne bindingi wskazują ten sam szablon; indeks
nie kopiuje rodziny per czasownik.

## Rodzina, wariant, plan i powierzchnia

- **Family** opisuje współdzielony kształt, np. NP–VP–NP.
- **Variant** wybiera dozwoloną odmianę rodziny, np. szyk lub licencjonowaną
  alternację. Jest identyfikowany oddzielnie przez binding.
- **ConstructionPlanID** to deterministyczny SHA-256 struktury rodziny po
  walidacji. Identyfikuje plan strukturalny, nie zdanie ani powierzchniowy tekst.
- **Surface** powstaje dopiero później przez kompozycję slotów, walencji,
  operatorów i generatorów języka.

Registry zapewnia `FamilyFor(language, familyID)`, `ByID`, `ByLanguage`,
`ByRoleLayout` oraz `BindingsByFamilyID`. `FamilyFor` wymaga zgodności obu
typowanych kluczy i nie ma fallbacku do konceptu. `ByRoleLayout` może zwrócić
rodziny PL i EN o takim samym układzie ról; ich constraints i policy pozostają
osobne.

## Sloty i ograniczenia

`SlotSpec` opisuje uporządkowany slot NP, VP lub PP, cardinality, required lub
optional, opcjonalne powiązanie z rolą semantyczną, przypadek i adpozycję.
Rola nie wyznacza automatycznie subject, object ani przypadku. W walencji
konkretnego sensu brakuje zgadywania rekcji.

Family constraint ma jawny typ (`precedes` albo `agreement`) i siłę `hard` lub
`soft`. W tym kroku registry waliduje kształt i referencje slotów, a nie
interpretuje constraintów; ich kompilację prowadzi późniejszy etap. Operator i
capability są ograniczonymi identyfikatorami, bez `eval` i bez kodu wykonywalnego
w JSON. Każda rodzina zachowuje `SourceRef`.

## Reużycie jednej rodziny przez wiele sensów

Poniższa ilustracja pokazuje wyłącznie kształt indeksu. Identyfikatory `demo`
są syntetyczne i nie są zatwierdzonymi danymi leksykalnymi ani przykładami
językowymi. Registry ma jeden wpis rodziny, a każdy wiersz to oddzielny
`ConstructionBinding` wskazujący ten sam `FamilyID`. Jedyny strukturalny wpis
ma sloty `NP(role=agent) → VP → NP(role=theme)`; to kształt przykładowej
rodziny transitive active, nie reguła globalna dla innych sensów:

| Jeden wpis family | SenseID w osobnych bindingach |
| --- | --- |
| `construction:demo.en.transitive` | `sense:demo.01`, `sense:demo.02`, `sense:demo.03`, `sense:demo.04`, `sense:demo.05`, `sense:demo.06`, `sense:demo.07`, `sense:demo.08`, `sense:demo.09`, `sense:demo.10`, `sense:demo.11` |

`BindingsByFamilyID(construction:demo.en.transitive)` zachowuje wszystkie 11
odrębnych sense/frame references; `ByID` nadal zwraca jedną rodzinę. To
pokazuje wielokrotne licencjonowanie szablonu bez tworzenia kopii struktury.
Przed umieszczeniem rzeczywistych sensów w seedzie każdy binding wymaga
zatwierdzonej walencji i provenance.

## Różne profile językowe

PL i EN mogą mieć taki sam role layout, ale muszą mieć oddzielne rodziny i
constraints. Na przykład PL może jawnie wymagać slotu PP z `na` oraz odpowiedniego
przypadku, a EN slotu PP z `at`. Są to przykłady kształtu, nie zatwierdzone
ramy dla konkretnych sensów. `ByRoleLayout` znajduje oba wpisy, a
`FamilyFor(pl, plFamilyID)` zwraca tylko rodzinę PL.

Duplicate family/slot/constraint IDs, nieciągła kolejność slotów, niespójna
cardinality, dangling slot reference, niepoprawne provenance lub zły język są
odrzucane przed wpisaniem do immutable registry. Nieznany `FamilyID` zwraca
`unknown`; próba użycia rodziny innego języka zwraca `unsupported`.

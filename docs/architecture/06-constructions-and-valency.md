# 06 — ConstructionGraph: podstawowa jednostka granulacji składni

**Status:** PROPOSED. Issues: [G03-01…07 #31–37](https://github.com/clickbytes/golingua/issues/4), [G05-03…06](https://github.com/clickbytes/golingua/issues/6), [G07-06](https://github.com/clickbytes/golingua/issues/63). Źródło wzorców: pinned `languages/pl/grammar.json`, `languages/en/grammar.json`, `bridge.json`, `docs/adr/generic-language-feature-contract.md` i `feature-driven-word-order.md`.

## 1. Cztery rzeczy, których nie scalać

1. `ConceptDefinition`, np. `event.read`, ma semantyczne role `agent` i `theme` i ich ograniczenia typów.
2. `LexicalSense`, np. `pl.czytac.sense.1`, ma **ValencyFrame** opisującą legalną realizację argumentów: np. użycie biernika w określonej konstrukcji. Dwie jednostki z podobnym concept/frame nie muszą mieć tej samej rekcji.
3. `ConstructionFamily`, np. `pl.transitive.active`, to typowy, wielokrotnie używany wzór slots/constraints/semantic operator capabilities. Rodzina jest odrębnym węzłem, może służyć dziesiątkom sensów.
4. `ConstructionVariant` ma jawny `VariantID`, kolejność, warunki i koszt/preferencję: np. `pl.transitive.active.svo` oraz `pl.transitive.active.ovs`. Wariant nie jest odrębnym conceptem ani automatyczną zmianą sensu.

`ValencyBinding` jest typowaną krawędzią `SenseID + ValencyFrameID → ConstructionFamilyID/VariantID`. Nie wyprowadzać automatycznie z samego faktu, że concept ma `agent/theme`. Poszczególne role semantyczne ≠ funkcje składniowe (`agent ≠ zawsze subject`).

## 2. Proponowany minimalny kontrakt źródłowy

To schemat poglądowy, **nie gotowy parser JSON**:

```json
{
  "id": "pl.transitive.active",
  "language": "pl",
  "family": "transitive",
  "slots": [
    {"id":"subject","category":"NP","required":true},
    {"id":"predicate","category":"FINITE_VP","required":true},
    {"id":"object","category":"NP","required":true}
  ],
  "constraints": [
    {"op":"govern","slot":"subject","casePolicy":"nominative"},
    {"op":"agree","target":"predicate","controller":"subject","features":["number","person"]},
    {"op":"govern","slot":"object","casePolicy":"from_valency"}
  ],
  "variants": [
    {"id":"svo","sequence":["subject","predicate","object"],"default":true},
    {"id":"ovs","sequence":["object","predicate","subject"],
     "when":["licensed_by_valency","roles_recoverable"]}
  ]
}
```

Dla EN analogiczna rodzina może wymagać SVO i niezależnej strategii pomocniczych w pytaniu. `casePolicy:from_valency` znaczy: *kompilator wymaga jawnej rekcji dostarczonej przez wybraną ramę*, a nie domyślnie acc dla każdego czasownika.

Następujący pseudokod Go jest wzorcem **zwykłych typowanych wywołań**, nie obowiązkowym „fluent property tree”:

```go
plan, err := compiler.CompileConstruction(
  languageID, senseID, valencyID, familyID, variantID, semanticOperators,
)
surface, err := realizer.Realize(plan, meaning, constraints)
```

Żadnego obowiązkowego `ROCK.VERB.MORPH.PAST`. Język nie ma globalnej puli właściwości generowanych dla każdego słowa.

## 3. ValencyFrame: pola obowiązkowe

- `SenseID` i `FrameID`; identyfikatory stable i reviewed.
- Mapowania `semanticRole → argumentSlotId` o jawnej krotności, np. `agent → subject`; to jest decyzja konkretnej ramy.
- Dla slotu: `NP/PP/clause`, wymagany/optional, semantyczny typ referenta, case/adposition/government, control, możliwy implicit subject tylko po dowodzie.
- Własne `licensedFamilyIDs/allowedOperators/blockedVariants`, w tym voice, WH gap, negation, tense/aspect compatibility.
- Alternacje jak `see vs look at`, `widzieć vs patrzeć na` są właściwością sensów i ramek, nie synonimem wymieniającym wyraz bez zmiany rekcji.

**Negative witness:** utworzenie dla `look at` konstrukcji z dopełnieniem bez wymaganej przyimkowej realizacji musi zakończyć się `unsupported`, nawet jeżeli graf semantyczny percepcji jest zgodny.

## 4. Konstrukcje są kompozycyjne, lecz nie tworzą nieograniczonego iloczynu

`BaseFamily + (Tense,Aspect,Polarity,Voice,Question,Focus) → OperatorComposition → Validation → ConstructionPlan`.

`ConstructionOperator` opisuje typed preconditions i output scope, a nie „doklej znak nie”. Operator negacji może mieć inne zachowanie przy pytaniu, modalności lub biernej; kompilator nie zakłada komutatywności. Zmiana voice nie może zmienić ról semantycznych; pominięty agent ma jawny operator/`omittedRoles` i warunek licencji.

**Plan kompilacji:**
1. Resolve i validate `SenseBinding + ValencyFrame`; brak zatwierdzonego sense → brak realizer.
2. Select explicit `FamilyID` i potwierdź zgodność slotów z sense frame.
3. Unify językowy `FeatureProfile`, case/government, controller agreement, operator scopes.
4. Expand tylko dozwolone warianty, validate required slots occurrence exactly once (`@verb` nigdy nie znika), precedence i information structure.
5. Sort stable plan components; compile współdzielony `ConstructionPlanID` i oddzielnie sense-specific constraints. Dwie ramy różne semantycznie mogą korzystać z jednego planu struktury bez utraty identity sensu.
6. Zachowaj status `canAnalyze` i `canRealize` osobno, wraz z reviewed evidence IDs.

W source po 51 ramach zdarzeń wykryto odpowiednio około **19 PL / 15 EN** różnych struktur po znormalizowaniu ramek bez concept-specific key; grupy po 25 są identyczne. To argument za dedup, nie dowód, że każda z 51 ramek ma pełną walencję i coverage.

## 5. Generatory fraz jako konstrukcje niższego poziomu

Nie tylko pełne zdania są konstrukcjami. Własne identyfikatory i plany mogą mieć również:
- `NominalPhrase(NP)`: noun head, determinant/article, demonstrative, numeral, possessive, adjective agreement, countability, noun phrase case and definiteness. Flexibility PL i determiner policies EN w pluginie. MWE nominatywno-fleksyjne mogą posiadać internal head/anchors.
- `FiniteVerbPhrase(VP)`: lexeme/sense, finite/nonfinite, tense/aspect/mood/voice, auxiliary selection, person/number/gender controlled agreement, negation placement, idiomatic restrictions.
- `PrepositionalPhrase(PP)`: lexically licensed adposition, governed case, phrase head and agreement constraints.
- `Clause/Coordination/Relative/Complement`: operator scope i ograniczenia rekursji; zgody/stylistyka nie są globalnymi operatorami bez dowodów.
- `NumericQuantity`: quantity jako semantyczna wartość; language-local rule wybiera przypadek NP oraz wymaganą liczbę/zgodę czasownika. **Nie deklarować** uogólnienia wyłącznie dlatego, że istnieje tabela wartości 0..20 w źródle.
- `TimeExpression`: dzień/deiktyczny czas/własna data-godzina i kontekst, nie automatycznie `tense`.
- `MultiwordConstruction`: stałe i odmienne człony, idiomy, separable elements, restrictions. Sens idiomatyczny nie jest literalnym sense słów-składników.

Wymagane **proceduralne rozszerzenie planu**: NP i VP mają własne plany/generatory i własne acceptance fixtures. G03-07/G05-06 nie mogą ograniczyć się do jednego sklejenia stringów z gotowych ręcznych form.

## 6. Wariacje PL/EN — nie zgaduj znaczenia z szyku

PL: `Dziewczynka widzi psa` i `Psa widzi dziewczynka` mogą zachować role `agent=girl` i `theme=dog` dzięki przypadkom; tego wymagają właściwe morfologiczne readings. EN: `The girl sees the dog` / `The dog sees the girl` mają **inne role** przy standardowym SVO. Konstrukcja PL OVS nie jest automatycznym dowodem topic/focus; brak takiej evidence = unknown information structure.

**Negative oracle** (sam schemat testu; wymagane niezależne zatwierdzenie):
- Dublowanie `predicate` w sekwencji → compile invalid.
- Brak wymaganej roli `theme` dla czasownika wymagającego dopełnienia → unsupported.
- PL OVS bez morfologicznego dowodu przypisania ról → ambiguous/incomplete, nie confident unique.
- Pytanie przekształcone w asercję lub negacja przesunięta do innego scope → semantic mismatch FAIL.
- Czasownik niemający licencji biernej → brak automatycznego passive generator.
- Dwa niezgodne controller agreements → compile conflict, bez „ostatni wygrywa”.

## 7. Granice wdrażania

G03 implementuje **typy, kompilację i mały PL/EN zestaw planów**; G05 wykonuje parser i realizer. Nie robić „uniwersalnej gramatyki wszystkich języków” ani list ręcznych wyjątków w Go switch/if po ConceptID. Dowolna składnia niewspierana przez udowodniony capability pozostaje unsupported. Dalsza szczegółowość fraz wymaga odpowiednio małych osobnych issues — wskazane w [mapie braków](13-issue-traceability.md).

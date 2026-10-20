# 06 — Konstrukcje językowe jako poziom granulacji

**PROPOSED, G03.** Konstrukcja językowa to **pierwszorzędny element architektury**. Nie jest jedynie ciągiem tokenów ani wygodną nazwą konstrukcji w JS.

## Warstwy

`ConceptDefinition` określa znaczenie i role, np. `event.visual_perception` oraz `agent/theme`. `LexicalSense` identyfikuje konkretny sens leksemu. `ValencyFrame` mówi, **jak ten sens** realizuje argumenty i jakie ograniczenia obowiązują. `ConstructionFamily` określa językowy wzorzec (np. transitive active, predication, possessive). `ConstructionVariant` wybiera legalny wariant szyku/formy. `ConstructionOperator` składa pytanie, negację, czas, modalność itp. `ConstructionPlan` to wynik kompilacji potrzebny parserowi i realizatorowi.

Nie można wyprowadzać rekcji słowa z samego konceptu. `see` i `look at` mogą dotyczyć percepcji, ale przyjmują różne konstrukcje. `widzieć` i `patrzeć na` również mają odmienny rząd.

## Przykładowa rodzina (ilustracja, nie gotowy JSON schema)

```json
{
  "id": "pl.transitive.active",
  "language": "pl",
  "slots": [
    {"id":"subject","kind":"NP","required":true},
    {"id":"verb","kind":"FINITE_VP","required":true},
    {"id":"object","kind":"NP","required":true}
  ],
  "constraints": [
    {"kind":"govern","slot":"subject","case":"nom"},
    {"kind":"agree","target":"verb","controller":"subject","features":["person","number"]},
    {"kind":"govern","slot":"object","policy":"from_valency"}
  ],
  "variants": [
    {"id":"svo","order":["subject","verb","object"]},
    {"id":"ovs","order":["object","verb","subject"],"requires":"roles_recoverable"}
  ]
}
```

Nie przyjmuj globalnego `object=acc` dla każdego polskiego czasownika. `from_valency` wymaga zadeklarowanej rekcji konkretnego sensu. Sam szyk OVS nie dowodzi focus/topic. Angielski ma inne porządki; wyrażenie `The dog sees the girl` nie jest semantycznym odpowiednikiem `The girl sees the dog`.

## Walencja

Dla każdego `ValencyFrame`: `SenseID`, role semantyczne, typy argumentów, obowiązkowość, funkcja składniowa, przypadek/przyimek, alternacje, kontrolery zgody i licencja na operatory. Brak ramy dla sensu nie może włączyć automatycznie „najbliższego” wzorca.

## Podkonstrukcje i generatory

**Nie ograniczamy ConstructionGraph do całych zdań.**

- `NP`: rzeczownik główny, przydawki, rodzaj/żywotność, liczba, przypadek, determiner/article, dzierżawcze, przysłówkowe modyfikatory, liczebniki.
- `VP`: czasownik i wymagane argumenty, finite/nonfinite, tense/aspect/mood, polaryzacja, voice, pomocnicze, zgoda z rzeczywistym kontrolerem.
- `PP`: licencjonowana przez walencję adpozycja/przyimek, narzucony przypadek, zagnieżdżona NP.
- `Clausal`: pytanie yes/no lub WH, dopełnienie, zdanie względne, koordynacja — w miarę rzeczywistej implementacji.
- `Quantified NP`: wartość 5 w znaczeniu nie oznacza automatycznie gramatycznej liczby pojedynczej. PL i EN dobierają własną realizację.
- `Temporal expression`: `jutro` nie jest równoznaczne ze zmianą gramatycznego czasu zdarzenia.
- `MWE/idiom`: stałe i odmieniane komponenty, rama, oddzielny sens idiomatyczny. `break the ice` nie może automatycznie tworzyć literalnego zdarzenia rozbijania lodu.

Warianty konstrukcji nie powinny być mnożone do wszystkich połączeń. `Family + Operator + Valency + Profile → check → Plan`. Konstruktor ocenia legalność, a potem zleca pracę generatorom form. Operatory nie muszą komutować, zwłaszcza przy negacji, pytaniu i modalności.

## Kompilacja i runtime

Jedna rodzina konstrukcji może obsługiwać wiele sensów. Dane trzymamy raz, a `ConstructionBinding` ma konkretne `SenseID, ValencyFrameID, FamilyID, VariantID`. Zależności na kontrolerze zgody, obowiązkowym miejscu czasownika, sekwencji i typach muszą być zweryfikowane przed użyciem.

Projektowane Go API może być zwykłym `CompileConstruction(lang,sense,frame,family,variant,operators)`. **Nie implementuj `ROCK.VERB.MORPH.PAST` jako koniecznej nawigacji lub przechowywanego drzewa.**

## Jak będziemy oceniać

Ręcznie, na zdaniach, które poda właściciel. Na początek wystarczą konkretne wejścia pokazujące problem (np. inny podmiot/dopełnienie albo niewłaściwy przypadek). Koryguj źródłową konstrukcję/rekcję/generator, **nie dopisuj tablicy całych zdań**. Nie twórz automatycznego generatora przykładów, zestawu expected ani test runnera.

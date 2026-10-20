# 01 — Granulacja grafów, identyfikatory i zależności Go

**Status:** PROPOSED. Zasada: **grafy są warstwami logicznymi jednego procesu**, nie osobnymi bazami, mikroserwisami ani fizycznymi tablicami na każdy człon ścieżki.

## Węzły i odpowiedzialność

| Węzeł / struktura | Jeden rekord reprezentuje | Właściciel | NIE wolno w nim umieścić |
| --- | --- | --- | --- |
| `ConceptDefinition` / ConceptGraph | znaczenie typu: `animal.dog`, `event.read`, role i relacje | `internal/core` | lemma, case, surface order, user facts |
| `MorphLexeme` / LexicalGraph | leksem w jednym języku, lemma/POS i inherent features | `internal/lexicon` | niezatwierdzony CoreRef jako automatyczny fakt |
| `LexicalSense` | jedno znaczenie konkretnego leksemu, nawet jeśli jest wieloznaczny | `internal/lexicon` | składni każdej realizacji jako globalnego CORE |
| `SemanticBinding` | `SenseID → CoreRef`, typ mapowania, review/provenance | `internal/lexicon` | utożsamienia `related` z `equivalent` |
| `ValencyFrame` / ValencyGraph | dopuszczone argumenty, ich role/funkcje/rekcję danego `SenseID` | `internal/valency` | reguł słowotwórczych i faktów świata |
| `ConstructionFamily` | wspólny, językowy wzór składni, np. transitive / intransitive | `internal/construction` | kopii każdego konceptu i leksemu |
| `ConstructionVariant` | konkretną dopuszczoną sekwencję/scope/markedness, np. PL OVS | `internal/construction` | nowego sensu tylko dlatego, że zmienił się szyk |
| `ProcedurePlan` / ProcedureGraph | mały typowany plan generowania, rozpoznania lub transformacji | `internal/procedure` | arbitralnego kodu z JSON, I/O, mutacji CORE |
| `Paradigm` / MorphologyGraph | klasy odmiany i reguły komórek + wyjątki per leksem | `internal/morph` | konceptów jako obowiązkowej morfologii |
| `EvidenceRecord` / EvidenceGraph | świadectwo: source, review, fixture, capability, reguła, wynik | `internal/evidence` planowane lub `internal/provenance` staging | twierdzenia „implemented” bez testu |
| `MeaningGraph` | **instancję interpretacji zdania**: event/state/question/operator/roles/referents | `internal/meaning` | polskiego przypadku, angielskiego szyku, inferred world fact |
| `WorldFact` / request FactSet | pojedynczy, czasowo ugruntowany fakt o **konkretnych referentach** | `internal/facts` w G06 | definicji słownikowej jako zastępnika faktu użytkownika |

Nazwa `EvidenceGraph` oznacza sieć stabilnych, typowanych referencji między dowodami i produktami; na poziomie v0 można ją przechowywać jako indeksowane rekordy, nie potrzebuje grafowej bazy. Source `languages/<code>/graph.json` w flexLingua jest *grafem topologii pluginu*, nie semantycznym ConceptGraph. Nie kopiować tych typów 1:1.

## Typowane krawędzie; relacja nie jest dowolnym stringiem

| Od → do | Relacja (proposed) | Wymagany dowód/ograniczenie |
| --- | --- | --- |
| Concept → Concept | `is_a`, required definition relation / typical link | valid typowanie, jawna transitivity/symmetry, typical nigdy nie dowodzi |
| MorphLexeme → Paradigm | `inflects_by` | morphology ready, zgodny POS i feature profile |
| MorphLexeme → LexicalSense | `has_sense` | SenseID unikalny, wszystkie sense niezależne |
| LexicalSense → Concept | `maps_to` (SemanticBinding) | `approved equivalent` dla bezstratnego transferu; broader/narrower jako informacja, nie substitution |
| LexicalSense → ValencyFrame | `licenses_frame` | sense-specific, stabilne role i selekcja typów |
| ValencyFrame → ConstructionFamily | `licenses_family` | frame/language/role cardinality compatible |
| ConstructionFamily → ConstructionVariant | `allows_variant` | all slots bound, negation/question/voice conditions |
| ConstructionPlan → ProcedurePlan / Paradigm | `invokes` | poprawne typowane argumenty, bound cost |
| Evidence → {Binding, Plan, Rule, Fact, Case} | `supports` lub `tests` | reviewed source ID, immutable evidence and rev |

Nie tworzyć stałych grafowych ścieżek `ROCK.VERB.MORPH.PAST`. Past to **cecha wejścia generatora**, nie obiekt ontologiczny ani węzeł przechowywany pod każdym słowem.

## Identyfikatory — trzy odrębne przestrzenie

- **SymbolID** stabilny i zewnętrzny, np. `animal.dog`, `en.go.1`, `pl.transitive.active`. Tożsamość zależy od jawnego namespace i nie może być zgadywana po części tekstowej nazwy.
- **Typed ID**: `ConceptID`, `LexemeID`, `SenseID`, `ValencyID`, `ConstructionID`, `ProcedureID`, `EvidenceID`. Nieakceptowalne są porównania ID różnych typów na podstawie samego stringa.
- **Runtime NodeID (uint32 lub równoważny)**: lokalny dla **jednego immutable pack snapshotu**, wyznaczany deterministycznie podczas kompilacji. Nie zapisuje się go jako trwałej semantycznej referencji; dodanie nowego symbolu może zmienić indeksy wszystkich kolejnych.

Dedup, sort i canonicalization muszą działać bez kolizji; hash może być indeksem pomocniczym, **nie dowodem tożsamości**. Zachowaj publiczny SymbolID w trace.

## Kierunki zależności pakietów — DAG bez cykli

`internal/ids`, `internal/status` i `internal/provenance` to liście bazowe. **Proponowana korekta issue G02-02:** współdzielona unifikacja cech nie powinna leżeć tylko w `internal/morph/features.go`, bo jest wykorzystywana również przez `construction` i `valency`. Neutralne, typowane mechanizmy umieścić w `internal/features`, językowe profile w pluginach; morfologia importuje ten pakiet, nie odwrotnie.

`internal/core` zawiera typy i definicje, nie importuje `internal/graph`; `internal/graph` zależy od `core`. `internal/meaning` może walidować względem `core`; `core` nigdy nie importuje `meaning`. `internal/procedure` implementuje bezpieczne, generyczne operacje i **nie importuje `morph` ani `construction`**. Zależności:

`ids/status/features → core → graph,meaning,lexicon; lexicon+features+procedure → morph; meaning+lexicon+features → valency; valency+features+procedure → construction; all validated data → pack; pack + analyzer/parser/realizer → engine; meaning+facts → reason → planner → dialogue; engine → cli`.

`internal/importer` zapisuje kandydujące schematy i staging, ale nie może modyfikować aktywnego `pack`; pack compiler czyta wyłącznie zatwierdzone dane. Żaden komponent CORE nie importuje `languages/pl` ani `languages/en`.

## Własności grafu

1. Ontology `parent`/`is_a` acykliczne i twarde; inne relacje mogą tworzyć cykle, wymagają visited set i budżetu.
2. Role `agent/theme/recipient` są semantyczne; `subject/object` i przypadek wyznacza wyłącznie konstrukcja+walencja.
3. `definitionGraph.required` jest ograniczeniem typu, `typical` to sygnał preferencji **bez siły logicznej**. Brak obserwacji nie jest sprzecznością.
4. Dwa referenty tego samego konceptu nie stają się automatycznie jedną instancją. Identity wymaga dowodu.
5. Tylko jawnie licencjonowane alternacje konstrukcji mogą zamieniać realization plan; negacja/pytanie/warunek nie zmienia się w assertion.
6. Powtarzalny porządek ID/edge/trace i jawny resource_limit. Żaden limit nie może po cichu usunąć alternatyw przy komunikacie `unique`.

### Minimalny przykład

`event.visual_perception` to ConceptID; `en.see.1` to MorphLexeme; `en.see.1.sense.1` to LexicalSense związany z konceptem; ValencyFrame podaje, że sense dopuszcza obserwującego i postrzeganą encję; `en.construction.transitive.active` określa SVO; ProcedurePlan dobiera `sees`; konkretne `girl_1` i `dog_2` istnieją dopiero w MeaningGraph/FactSet, nie jako nowe słownikowe koncepty.

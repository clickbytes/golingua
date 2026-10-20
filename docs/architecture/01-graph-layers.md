# 01 — Warstwy grafowe i granice typów

**Częściowo wdrożone.** `internal/core` kompiluje i waliduje language-neutral ConceptGraph oraz relacje; `internal/graph` buduje snapshot-local adjacency i bounded traversal. Pozostałe warstwy niżej są proponowanym zakresem, dopóki ich issue nie zostaną scalone. Jeden proces Go; logiczne grafy nie muszą być osobnymi bazami danych.

| Warstwa | Właściciel (proponowany) | Odpowiedzialność |
| --- | --- | --- |
| ConceptGraph | `internal/core` | znaczenia, ontologia, typy relacji, `is_a` |
| LexicalGraph | `internal/lexicon` | MorphLexeme, LexicalSense, sense↔concept |
| ValencyGraph | `internal/valency` | wymagane/optional role per sens, case/adposition/rekcja |
| ConstructionGraph | `internal/construction` | familie konstrukcji, warianty, operatory i podkonstrukcje NP/VP/PP |
| MorphologyGraph | `internal/morph` | cechy, paradygmaty, tematy, komórki, wyjątki, inverse lookup |
| ProcedureGraph | `internal/procedure` | ograniczony interpreter typowanych, współdzielonych operacji |
| Evidence/Provenance | `internal/provenance` | źródła danych, ręczne adnotacje, identyfikatory reguł w diagnostyce |
| MeaningGraph | `internal/meaning` | znaczenie jednego zdania, role, referenty, operatory |
| FactSet (później) | `internal/facts` | jawnie podane fakty o konkretnych instancjach, nie konceptach |

## ID

Stabilne symboliczne `ConceptID, LexemeID, SenseID, FamilyID, VariantID, ProcedureID`. Wewnątrz niemutowalnego pack można używać gęstych `NodeID`; nie powinny być identyfikatorami trwałymi między wersjami. Hash nie jest tożsamością.

**Ważna zmiana względem wygodnej nawigacji:** `PAST` to cecha/komórka generatora, nie węzeł ConceptGraph pod każdym czasownikiem. Nie trzeba implementować fluent API.

## Typowane krawędzie

- `Concept is_a Concept` i inne typowane relacje, required/typical nie są równoważne.
- `MorphLexeme has_sense LexicalSense`; `LexicalSense maps_to Concept` z jawnym typem zgodności.
- `LexicalSense licenses ValencyFrame`; `ValencyFrame licenses ConstructionFamily/Variant`.
- `ConstructionPlan invokes ProcedurePlan/Paradigm`; wyjątki mogą zastępować wynik tylko w deklarowanym zakresie.
- `MeaningGraph` tworzy konkretne participant/referent instancje; nie duplikuje definicji konceptu.

## Zależności Go

Proponowany kierunek: `internal/ids` i `internal/status` → neutralne `internal/features` i `internal/core` → `internal/lexicon`/`internal/meaning` → `internal/morph`/`internal/valency` → `internal/construction` → `internal/parser`/`internal/realizer` → `internal/engine`/CLI. `internal/procedure` nie może importować `construction` i `morph`, bo te dwa korzystają z procedur. Osobny kompilator/loader zasobów pracuje offline.

Nie wkładaj polskiej fleksji do `core`, nie implementuj transferu parami języków w `engine` i nie twórz jednego `if concept == X` per wyraz.

## Właściwości zachowania, sprawdzane teraz ręcznie

- Typ `dog` to nie konkretny pies; dwa referenty mogą należeć do jednej kategorii.
- `agent` to semantyczna rola, nie synonim podmiotu i przypadku mianownika.
- Przypadek gramatyczny/szyk powierzchniowy należy do pluginu, nie do MeaningGraph.
- Multiple readings pozostają alternatywami, nie pierwszym znalezionym wynikiem.
- Ograniczenia i brak danych skutkują jawnym komunikatem, nie automatycznym „best guess”.

Te inwarianty mają być zakodowane, ale nie wymagamy na tym etapie formalnej macierzy testowej.

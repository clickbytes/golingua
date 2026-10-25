# #344 M8-ID38 selection

Selected source: `/tmp/golingua-matrix-expansion/m8-identification-selected.json` (coordinator_authored_A1_identification_2026-10-25). All rows have `OwnerAcceptance=NOT_EVALUATED`; these are proposals, not owner-approved language data.

| Pair | Polish | English | Contract note |
|---|---|---|---|
| 001 | To jest pies. | This is a dog. | masculine animate noun, irregular stem pies |
| 002 | To jest ptak. | This is a bird. | animal class and masculine noun |
| 003 | To jest książka. | This is a book. | feminine inanimate noun class |
| 004 | To jest samochód. | This is a car. | masculine inanimate class |
| 005 | To jest stół. | This is a table. | masculine class with stem alternation |
| 006 | To jest krzesło. | This is a chair. | neuter class noun ending -o |
| 007 | To jest okno. | This is a window. | neuter artifact, not a place relation |
| 008 | To jest jabłko. | This is an apple. | neuter fruit and an article |
| 009 | To jest klucz. | This is a key. | physical key rather than abstract solution |
| 010 | To jest torba. | This is a bag. | feminine container classification |
| 011 | To jest nauczyciel. | This is a teacher. | profession class; not identity or biological sex |
| 012 | To jest lekarz. | This is a doctor. | profession, no specific person identity |
| 013 | To jest matka. | This is a mother. | kinship class without guessing whose parent |
| 014 | To jest ojciec. | This is a father. | parent class, no child referent invented |
| 015 | To jest koń. | This is a horse. | animal class and Polish consonant alternation |
| 016 | To jest ryba. | This is a fish. | singular animal, not a food portion |
| 017 | To jest kot. | That is a cat. | explicit English distal reference; Polish to leaves distance unresolved |
| 018 | To jest pies. | That is a dog. | explicit English distal reference; Polish to leaves distance unresolved |
| 019 | To jest książka. | That is a book. | explicit English distal reference; Polish to leaves distance unresolved |
| 020 | To jest jabłko. | That is an apple. | explicit English distal reference; Polish to leaves distance unresolved |
| 021 | To są koty. | These are cats. | plural classification with proximal source and unresolved Polish reference |
| 022 | To są psy. | Those are dogs. | plural distal classification, not universal class quantification |
| 023 | Ten pies jest czarny. | This dog is black. | masculine animate adjective agreement, no biological sex |
| 024 | Ten ptak jest czarny. | This bird is black. | black bearer bird rather than a closed cat list |
| 025 | Ten koń jest czarny. | This horse is black. | horse grammatical gender versus sex |
| 026 | Ten samochód jest czarny. | This car is black. | inanimate masculine bearer |
| 027 | Ta książka jest czarna. | This book is black. | feminine demonstrative and predicate adjective |
| 028 | Ta torba jest czarna. | This bag is black. | container feminine agreement |
| 029 | To okno jest czarne. | This window is black. | neuter demonstrative to, not classificatory to jest |
| 030 | To jabłko jest czarne. | This apple is black. | neuter count fruit bearer; no health inference |
| 031 | Te koty są czarne. | These cats are black. | nonpersonal plural adjective agreement; unknown distance in Polish te |
| 032 | Te książki są czarne. | These books are black. | feminine noun plural still nonpersonal adjective form |
| 033 | To są nauczyciele. | These are teachers. | masculine personal plural classification without race/colour ambiguity |
| 034 | Tamte konie są czarne. | Those horses are black. | both sources explicitly distal; do not report an invented deixis loss |
| 035 | Ta książka jest czerwona. | This book is red. | red bearer predicate absent from current CORE |
| 036 | Ten samochód jest biały. | This car is white. | white bearer predicate absent; cannot substitute black |
| 037 | Ta torba jest niebieska. | This bag is blue. | blue bearer predicate absent, feminine agreement |
| 038 | To jabłko jest zielone. | This apple is green. | green bearer predicate absent, neuter agreement |

Pairs 001–034 use independently built `class_membership(member)` or `black(bearer)` state graphs with present time, one referent of declared multiplicity, and exactly one unresolved anchored mention. Pairs 035–038 require red/white/blue/green color properties and remain `GAP_I`: CORE has black only, and generic `color` has no bearer role/value. No substitute black graphs are authored.

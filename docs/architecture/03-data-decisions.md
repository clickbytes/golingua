# 03 — Źródła danych, etap redakcyjny i runtime pack

**Status:** PROPOSED; kontrakty szczegółowe G04. Początkowa decyzja: **nie wybieramy bazy danych, FST ani własnego binarnego protokołu jako warunku v0**. Optymalizacja ma wynikać z profili i parity, nie z estetyki reprezentacji grafu.

## Trzy artefakty o trzech celach

| Warstwa | Postać v0 | Cel | Czy runtime modyfikuje? |
| --- | --- | --- | --- |
| Repo source of truth | czytelne JSON (ze schema) + reviewed source IDs | edycja paradygmatów, konstrukcji, sensów i wyjątków | nie |
| Import staging | lokalne wersjonowane rekordy kandydatów + review manifest i atomic promotion | import masowy, konflikty, provenance, licencje | nie, to oddzielne narzędzie offline |
| Compiled pack | zwalidowany immutable snapshot z indeksami liczbowymi i checksums; fizyczny format ustala G04-06 | szybki `Analyze/Generate/Realize/Reason` | nie |

SQLite jest **opcją edytorskiego staging**, jeśli potrzebna transakcyjność, nie podstawową grafową bazą ani hot-path runtime. FST/two-level morphology to osobny kandydat porównywany w G08; obecne JS ADR `morphology-backend-scaling` świadomie odkłada FST/lazy. W Go zaczynamy od skończonych relacji i indeksów odwróconych, ale nie projektujemy publicznego API uzależnionego od reprezentacji.

## Przykładowa *docelowa* topologia języka

```text
data/core/concepts.json
languages/pl/manifest.json
languages/pl/features.json
languages/pl/lexicon/{lexemes,senses,bindings}.json
languages/pl/morphology/{paradigms,lexemes}.json
languages/pl/valency/*.json
languages/pl/constructions/{families,variants,operators}.json
languages/pl/import/unimorph-profile.json
languages/en/... (symetryczna struktura, INNE reguły)
imports/manifests/{source,review}.json
fixtures/{core,morph,sentences,reasoning,coverage}/...
```

Nie są to istniejące pliki w bootstrap, a ścieżki docs i issues wyznaczają obowiązki przyszłych PR-ów. Odpowiedniki językowe nie mają wymuszonej symetrii reguł. Semantyka CORE jest wspólna, a POS, case, agreement i word order pozostają językowe.

## Wersjonowanie i unikanie mieszania ID

- Stable SymbolID i odrębne namespace dla concept, lexeme, sense, valency, construction, procedure i evidence; rozszerzanie CORE nie zmienia znaczenia istniejącego ID bez kontrolowanej migracji.
- `packHash` hash kanonicznych, posortowanych wejść (plus wersja kompilatora i schematów), nie zegar systemowy. ID typu `uint32` to tylko snapshot-local offsets i mogą zmienić się po dodaniu rekordu.
- Parser gotowego pakietu odrzuca nieznane wersje, checksum mismatch, dangling relation, niezatwierdzoną mapę i unsupported operator.
- Każdy import ma `sourceRepo/URI + version/revision + checksum + author + license + review`. Jeśli którykolwiek wymagany element jest nieznany, status jest `pending/unverified`, nie domyślne „approved”.

## Migracja pinned source

Przed implementacją importu wszystkie reguły schema i pliki weryfikować na [source anchor](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92). Na tym SHA: CORE 1002, PL lexical items 1005, EN 1006, PL morphological bindings 465, EN 462, PL/EN source frame 51/51. Ta statystyka to **próg kontroli kompletności odczytu**; nie znaczy zatwierdzonej semantycznej zgodności nowego modelu.

Import istniejących `forms`: traktować je jako obserwacje/test material. Docelowa definicja regularnej formy to lemma+paradigm+principal parts; jeżeli reguła nie odtwarza obserwacji, kandydat konfliktuje lub wymaga wyjątków. Nie usuwać zatwierdzonych explicit form, dopóki odpowiednik proceduralny nie ma niezależnego pokrycia. Nie utożsamiać lemma spelling z concept ID.

## Strategia wydajności

Indeksować:
1. `ConceptID → type/parent/relations` (dense immutable arrays / typed adjacency).
2. `normalizedSurface → []MorphReading` (wszystkie homografy).
3. `(Language,Lemma,POS) → []LexemeID`.
4. `ConceptID → []approved SenseBinding`.
5. `SenseID → []ValencyFrameID`.
6. `(Language,FamilyID,VariantID) → compiled ConstructionPlanID`.
7. `RelationType + subject/object → indexed WorldFact candidates`.
8. `CapabilityID → evidence/review/source/test IDs`.

Jeden proces Go, bez request-time SQL lub reparsing source grammar JSON. Nie budować specjalnego indeksu `Past` jako nowego węzła dla każdego leksemu. Cache tylko pure/immutable i wersjonowany hash pack; per-request fakty izolowane. Optymalizować dopiero po tym samym workload i bez spadku semantycznego pokrycia.

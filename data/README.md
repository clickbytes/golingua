# Data golingua

Ten katalog zawiera redakcyjny zapis 1000 znaczeń CORE (`core/1000-concept-proposals.tsv`) oraz 1000 decyzji PL (`pl/1000-core-review.tsv`), obok materiałów do importu i manifestów. **TSV nie jest ładowane jako aktywny runtime source**; kompilowalne deklaracje CORE i PL znajdują się w `internal/catalog/` i `languages/pl/`.

Planowany podział:

- `core/` — neutralne ConceptID, definicje i relacje pojęć;
- manifesty zbiorów — `schema_version`, `dataset_id` i rejestr `SourceRef`;
- rekordy konceptów, leksemów, sensów, paradygmatów i konstrukcji — osobne typy z osobnymi identyfikatorami.

Leksemy i gramatyka języka należą do `languages/<kod>/`, nie do neutralnego CORE. Lexeme i Sense nie są Concept. Mapowanie sensu do konceptu wymaga osobnego, zatwierdzonego SemanticBinding; wspólny zapis słowa lub tłumaczenie nie tworzy takiego mapowania.

Każdy rekord danych powinien zachować identyfikator źródła, jego wersję i licencję. Wartość licencji `unknown` jest jawnym brakiem ustalenia, nie zgodą na kopiowanie. Nie dodawaj tu starych goldów, expected translations ani automatycznych danych z flexLingua.

Projekt ma kompilowany katalog CORE 1000 oraz kompilowany leksykon PL (728 leksemów) i EN (19 leksemów), ale nie ma aktywowanych, zatwierdzonych ścieżek tłumaczenia zdań. Ręczna ocena znaczeń i odmian nadal jest odrębnym warunkiem jakości.

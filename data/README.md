# Data golingua

Ten katalog jest przeznaczony na przyszłe neutralne dane CORE, wersjonowane manifesty i ręcznie zatwierdzone deklaracje. Aktualny bootstrap nie ładuje jeszcze danych z tego katalogu.

Planowany podział:

- `core/` — neutralne ConceptID, definicje i relacje pojęć;
- manifesty zbiorów — `schema_version`, `dataset_id` i rejestr `SourceRef`;
- rekordy konceptów, leksemów, sensów, paradygmatów i konstrukcji — osobne typy z osobnymi identyfikatorami.

Leksemy i gramatyka języka należą do `languages/<kod>/`, nie do neutralnego CORE. Lexeme i Sense nie są Concept. Mapowanie sensu do konceptu wymaga osobnego, zatwierdzonego SemanticBinding; wspólny zapis słowa lub tłumaczenie nie tworzy takiego mapowania.

Każdy rekord danych powinien zachować identyfikator źródła, jego wersję i licencję. Wartość licencji `unknown` jest jawnym brakiem ustalenia, nie zgodą na kopiowanie. Nie dodawaj tu starych goldów, expected translations ani automatycznych danych z flexLingua.

Na tym etapie katalog nie zawiera aktywnego słownika ani loadera. Pusta kategoria nie oznacza, że dane są fałszywe lub nieobsługiwane.

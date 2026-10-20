# Dane językowe

`languages/<kod>/` przechowuje dane i reguły konkretnego języka. Pakiety są zwykłymi danymi Go/JSON, nie kodem pluginów `.so`.

Pakiet językowy może opisywać MorphLexeme, LexicalSense, paradygmaty, wyjątki, walencję i konstrukcje. Jego manifest podaje `schema_version`, `dataset_id`, kod języka oraz źródła. `SourceRef` wymaga jawnej wersji i pola licencji; walidacja metadanych nie stanowi zgody na redystrybucję.

Definicje ConceptID pozostają w neutralnym CORE. Lexeme może być znany morfologii, choć jego sens nie ma zatwierdzonego mapowania do konceptu. Nie przenoś polskiej fleksji, angielskiego szyku ani innych reguł gramatycznych do CORE; nie zgaduj SenseID lub SemanticBinding z pisowni.

`pl/features.json` i `en/features.json` zawierają wersjonowane profile cech morfologicznych. Pakiet [EN](en.md) udostępnia mały seed generatora i analizatora przez Go API; nie ma komendy forms w CLI ani aktywnego pakietu morfologicznego PL. Profile samodzielnie nie deklarują obsługi języka w CLI.

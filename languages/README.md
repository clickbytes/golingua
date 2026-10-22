# Dane językowe

`languages/<kod>/` przechowuje dane i reguły konkretnego języka. Pakiety są zwykłymi danymi Go/JSON, nie kodem pluginów `.so`.

Pakiet językowy może opisywać MorphLexeme, LexicalSense, paradygmaty, wyjątki, walencję i konstrukcje. Jego manifest podaje `schema_version`, `dataset_id`, kod języka oraz źródła. `SourceRef` wymaga jawnej wersji i pola licencji; walidacja metadanych nie stanowi zgody na redystrybucję.

Definicje ConceptID pozostają w neutralnym CORE. Lexeme może być znany morfologii, choć jego sens nie ma zatwierdzonego mapowania do konceptu. Nie przenoś polskiej fleksji, angielskiego szyku ani innych reguł gramatycznych do CORE; nie zgaduj SenseID lub SemanticBinding z pisowni.

`pl/features.json` i `en/features.json` zawierają wersjonowane profile cech morfologicznych. Pakiety [EN](../docs/languages/en.md) i [PL](../docs/languages/pl.md) udostępniają wąskie seedy generatora i analizatora przez Go API; działająca komenda `forms --language pl|en ... --json` obsługuje tylko te wąskie seedy, a `catalog` raportuje oddzielnie stan znaczeń Go. Profile samodzielnie nie deklarują obsługi języka w CLI.

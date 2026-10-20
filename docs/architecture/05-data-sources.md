# 05 — Źródła danych i manifesty

Ten dokument określa małą, wersjonowaną otoczkę dla przyszłych ręcznie tworzonych plików danych. Nie definiuje jeszcze kompletnego schematu ConceptGraph, morfologii ani konstrukcji i nie dodaje loadera. Brak wpisu nie jest negacją, a brak mapowania nie upoważnia do zgadywania znaczenia.

## Granice zbiorów

- `data/core/` zawiera neutralne definicje ConceptID i relacje pojęć. Nie zawiera odmiany ani polskich lub angielskich reguł gramatycznych.
- `languages/pl/` i `languages/en/` zawierają dane właściwe językowi: MorphLexeme, LexicalSense, paradygmaty, konstrukcje i reguły.
- MorphLexeme może istnieć bez SemanticBinding do ConceptID. Mapowanie sensu do konceptu jest osobnym, jawnym i zatwierdzanym rekordem.
- ConceptID, LexemeID i SenseID są różnymi, stabilnymi identyfikatorami. Nie wyprowadzaj jednego z drugiego ani z zapisu słowa.

## Manifest

Każdy plik JSON deklaruje `schema_version` i stabilny `dataset_id`. Plik core nie ma pola `language`; plik językowy podaje kod języka, np. `pl` albo `en`. Manifest oddziela wersję kształtu JSON od wersji źródła danych.

Minimalny szkielet neutralnego CORE:

```json
{
  "schema_version": 1,
  "dataset_id": "core.example",
  "sources": [],
  "concepts": []
}
```

Minimalny szkielet danych polskich:

```json
{
  "schema_version": 1,
  "dataset_id": "pl.example",
  "language": "pl",
  "sources": [],
  "lexemes": [],
  "senses": [],
  "paradigms": [],
  "constructions": []
}
```

To puste, ilustracyjne manifesty, a nie istniejące zasoby językowe. Schemat pól poszczególnych rekordów będzie ustalany przez task implementujący dany typ; nie należy z tych szkieletów wywnioskować żadnych leksemów, faktów ani mapowań.

## Pochodzenie

Każdy użyty rekord wskazuje co najmniej jeden wpis `sources` przez `source_id` lub `source_ids`. Minimalny wpis ma stabilne `id`, dokładnie jedno `path` albo `uri`, wersję źródła i jawne pole licencji:

```json
{
  "id": "source.manual-1",
  "path": "sources/editorial-notes.md",
  "version": "owner-reviewed-edition-1",
  "license": "unknown"
}
```

Przykładowe wartości powyżej są znacznikami struktury, nie cytowaniem prawdziwego źródła. `license: "unknown"` oznacza brak ustalenia; nie jest zgodą na kopiowanie ani redystrybucję. `SourceRef.Validate` sprawdza wyłącznie kształt, ograniczenia długości oraz bezpieczną postać lokalnej ścieżki lub URI (`http`, `https`, `urn`, `doi`, `ark`). Nie pobiera źródła, nie ocenia licencji i nie zatwierdza danych do runtime.

Wersja źródła może być nazwą wydania, datą lub rewizją właściwą dla danego dostawcy. Nie musi być SHA. Nie wiąż identyfikatora ani wersji ręcznie utworzonego zbioru ze starą rewizją flexLingua. Stare źródło może być konsultowane wyłącznie jako opcjonalna pomoc; ten format nie zawiera pól gold, expected translation ani JS schema.

## Walidacja i aktywacja

Walidacja manifestu, kontrola odwołań między rekordami, zatwierdzanie danych i publikacja snapshotu należą do późniejszego loadera/kompilatora. Sama poprawna składnia JSON lub `SourceRef` nie dowodzi, że semantyka jest kompletna, licencja pozwala na użycie ani że dane mogą trafić do aktywnego runtime. Dane częściowe pozostają częściowe; nie uzupełniaj brakujących sensów, relacji lub źródeł na podstawie nazwy.

Nie ma automatycznego importu ani pobierania danych podczas obsługi tekstu. W tym tasku format jest dokumentowany, a `SourceRef` jest jedynym nowym typem wykonawczym.

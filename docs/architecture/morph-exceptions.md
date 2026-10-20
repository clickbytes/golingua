# Wyjątki morfologiczne

`internal/morph` kompiluje ograniczony zestaw jawnych reguł wyjątków dla jednego `LanguageFeatureProfile`. Reguła jest danymi, nie kodem. Każda wymaga wersji schematu, stabilnego ID, selektora, efektu, poprawnego `SourceRef` oraz recenzenta i czasu recenzji (`reviewed_by`, RFC3339 `reviewed_at`). Kompilator nie pobiera źródeł ani nie stwierdza, czy licencja zezwala na redystrybucję.

## Zakres selektora i precedencja

Zakres wymaga dokładnie odpowiadającego mu zestawu identyfikatorów:

| Zakres | Wymagane pola selektora | Specyficzność |
| --- | --- | ---: |
| `language` | język | 1 |
| `pos` | język, POS | 2 |
| `paradigm` | język, POS, paradigm ID | 3 |
| `lexeme` | język, POS, lexeme ID | 4 |
| `cell` | język, POS, paradigm ID i cell ID | 5 |
| `lexeme_cell` | język, POS, paradigm ID, lexeme ID i cell ID | 6 |

Selektor może dodatkowo zawierać ograniczenia do znanych wartości cech profilu. W ramach jednego zakresu większa liczba zgodnych ograniczeń cech jest bardziej specyficzna. Identyfikatory cech i wartości muszą istnieć w danym profilu językowym.

Bardziej specyficzna reguła nakładająca się na mniej specyficzną musi wymienić ją w `override_of`; jej `priority` musi być ściśle wyższe. Przy równej specyficzności rozstrzyga tylko jawne `override_of` z wyższym priorytetem. Priorytet sam w sobie nigdy nie rozstrzyga konfliktu. Wiele nakładających się `add_variant` tej samej specyficzności może współistnieć, jeżeli nie powtarza form. Każdy pozostały konflikt, nieistniejący cel override albo cykl override powoduje odrzucenie całego zestawu.

Kompilator kopiuje wejście i porządkuje skompilowany zestaw po ID, dlatego zmiana kolejności reguł w JSON nie zmienia wyboru ani wyniku.

## Efekty

- `block`: zwraca `blocked=true`, bez form. Konsument nie może użyć regularnego fallbacku; sygnał blokady dotyczy też akceptacji formy w analizatorze.
- `replace`: zwraca wyłącznie jedną jawnie podaną formę. Forma bazowa nie jest zachowywana.
- `add_variant`: zachowuje formy bazowe ze źródłem `regular` i dopisuje podane warianty ze źródłem `exception`.
- `stem_override`: zwraca jawne `StemOverrides`, `RequiresRegeneration=true`, źródło `exception` i ślad. Konsument generuje formę od nowa na nadpisanym temacie; nie używa wyniku policzonego ze starego tematu.
- `suppletion`: zwraca wyłącznie jedną jawną formę, bez fikcyjnego regularnego przepisu.

Każda wygenerowana forma wyjątkowa zachowuje `source=exception`, ID reguły oraz `SourceRef`. Ślad niesie ID reguły, efekt i identyfikator źródła. Nieznany selektor z brakującą/absent cechą daje `unknown`; nierozstrzygnięta cecha daje `incomplete`. Brak formy bazowej bez pasującego wyjątku daje `unknown`.

Pusta komórka paradygmatu, jawny zerowy wykładnik i brak dowodu na formę to trzy różne sytuacje. `block` jest jawną odmową dla celu, `replace` i `suppletion` wymagają jawnej formy, a brak reguły lub danych nie tworzy formy.

## Limity i granice

Kompilator ogranicza zestaw do 2 048 reguł, 8 192 zapisów form i 16 384 krawędzi override; jedna reguła może emitować najwyżej 64 formy. Formy są ograniczone do 1 024 bajtów i 512 punktów kodowych. Nie ma wykonań skryptów, fallbacku po `block` ani zależności od kolejności dokumentu. Obsługa wyjątków jest osobnym krokiem runtime; pełny `Generate`/`Analyze` jest realizowany w kolejnym issue G02.

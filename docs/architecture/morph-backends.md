# Generator i tabela relacji morfologicznej

`internal/morph.Generate` łączy jeden `MorphLexeme`, jawne `PrincipalParts`, językowy `FeatureBundle`, skompilowane paradygmaty i wyjątki oraz budżet procedury. Generuje tylko komórkę wybraną z profilu. Brak wymaganej formy diagnostycznej daje `unknown`; brak reguły komórki daje `unavailable`; wyjątek `block` daje `blocked` i nigdy nie uruchamia regularnego fallbacku. Błędne dane, niezgodny profil i limity pozostają typowanymi błędami.

Wynik zawiera powierzchnie wraz z leksymem, językiem, wersją profilu, POS, paradygmatem, komórką i wartościami cech. Źródło regularne oznacza plan procedury. Źródło wyjątkowe zachowuje ID reguły, provenance i trace; także `stem_override` wykonuje ponowne generowanie na jawnym nadpisanym temacie.

## Finite relation table

`GeneratedForm.Reading` konwertuje wygenerowany wynik do relacji wejściowej. `BuildFiniteRelationTable` waliduje i kanonizuje odczyty względem profilu, sortuje je deterministycznie i oblicza SHA-256 kanonicznego JSON. Duplikaty odczytów są błędem. Kodowanie zawiera digest, a dekoder odrzuca nieznane pola, nadmiarowe JSON i niezgodny digest.

Tabela ogranicza liczbę rekordów do 100 000 i rozmiar do 32 MiB; budowniczy przyjmuje niższe limity przez `FiniteTableBudget`. Każda powierzchnia ma ograniczoną długość. Normalizacja `Analyze` jest wyłącznie NFC. Nie zmienia wielkości liter ani diakrytyki i nie utożsamia transliteracji.

`BuildInverseIndex` buduje niemutowalny indeks powierzchnia→odczyty. `Analyze` zwraca wszystkie dopasowania w stabilnej kolejności, co zachowuje homografię i synkretyzm. Brak odczytu to `unknown`, nie negatywny fakt językowy. Indeks jest projekcją dostarczonej finite table; nie importuje ani nie zgaduje brakujących interpretacji.

Moduł jest biblioteką wewnętrzną, bez nowej komendy CLI i bez zależności wymaganych przez CI poza modułem Go do Unicode NFC. Wywołujący może nadal budować i uruchamiać bootstrap bez instalowania zewnętrznych narzędzi RTK lub CodeGraph.

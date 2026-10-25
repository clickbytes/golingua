# A1-M8 — liczebność a pełne zdanie egzystencjalne

Decyzja semantyczna koordynatora z 2026-10-25, po odrębnym przeglądzie kodu i kontraktu przez Luna Low. Akceptacja językowa właściciela: NOT_EVALUATED. Nie zmieniamy silnika ani danych CORE.

Dotychczasowy `QuantifierExactly(2) → cat(many)` jest legalnym grafem samej liczebności. Nie zawiera czasu. `TemporalFrame` jest dozwolone tylko na zdarzeniu/stanie; węzły kwantyfikatora i referenta nie mogą go otrzymać. CORE `existence` nie deklaruje roli wiążącej stan z liczonymi referentami. Utworzenie, utrzymywanie obecności lub lokalizacja nie zastępują czystego stwierdzenia istnienia.

Pełne zdanie „There are two cats.” ma znany teraźniejszy zakres czasowy. Nie można zgubić tej informacji, a potem oczekiwać od parsera tego samego nieczasowanego grafu. Kontrakt LIGHT-F1 wymaga samowystarczalnego znaczenia: [czas i wielość](../../architecture/meaning-attributes.md). Dowody strukturalne są w `internal/meaning/{types,validate,attributes}.go`, a brak roli istnienia w `internal/catalog/concepts_abstract.go`.

W partii lokalizacji/ilości pełne zdania dostaną jawny GAP_I z brakiem wiązania egzystencji, referenta/ilości i czasu. Starsze dwa rekordy kierunkowe o dwóch kotach wymagają tej samej korekty. Legalny komponent liczebności pozostaje osobno badany; nie jest pełnym wzorcem zdania ani zaliczeniem wymogu. To korekta semantyczna kontraktu, niezależna od outputu.

Dobór językowy również nie dodaje faktów przestrzennych: „is on” nie upoważnia do wprowadzenia leżenia; „in” nie przesądza pełnego zamknięcia pojemnika; przód/tył nie wybiera niepodanego punktu widzenia.

Dla liczebników zbiorowych zweryfikowano regułę pojedynczego orzeczenia w [materiale ZPE](https://zpe.gov.pl/a/przeczytaj/DVvjiEYAM) oraz użycie dwoje przy drzwi w [WSJP PAN](https://wsjp.pl/haslo/podglad/37518/dwoje). Stąd autorska forma „Jest dwoje drzwi.”, bez mnożenia drzwi przez „pary”. Źródła służą weryfikacji reguły; nie importujemy zdań ani kodu.

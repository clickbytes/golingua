# #425 — kontrakt jawnej korekty jednej komórki morfologicznej

**Status:** kontrakt zakresu implementacyjnego. Nie opisuje działającego dziś polecenia ani trwałego magazynu uczenia. Ocena agenta i akceptacja językowa właściciela są osobne; akceptacja właściciela pozostaje `NOT EVALUATED`, dopóki właściciel jej nie wyrazi.

## Cel i granica

Jedna jawna korekta ma zmienić wynik generowania **jednej komórki** jednego istniejącego leksemu. Adres komórki to pełny zestaw:

`language + profile_version + LexemeID + POS + ParadigmID + CellID + kompletny FeatureBundle`.

Korekta nie zmienia lematu, klasy odmiany, principal parts, reguły paradygmatu ani innych komórek. Nie uogólnia formy na inne liczby, przypadki, osoby, czasy, aspekty ani leksemy, nawet jeśli ich powierzchnia wydaje się podobna. Nie uczy reguły ortograficznej ani fonologicznej.

## Podstawa korekty i pochodzenie

Korekta zaczyna się od istniejącej, jednoznacznie wybranej analizy/generacji. Przed zmianą system zapisuje:

- wybrany `MorphLexeme.ID`, język, POS, `ParadigmID`, `CellID`, wersję profilu i pełny, zwalidowany zestaw cech;
- tożsamość aktywnego snapshotu danych: dataset, wersję i digest;
- oryginalny tekst lub formę, źródło zgłoszenia oraz dokładny odczyt, który wskazał leksem i komórkę; wieloznaczność zachowuje wszystkie odczyty i wymaga jawnego wyboru użytkownika;
- wynik bazowego `morph.Generate`: powierzchnię, stan, `GeneratedForm.Source`, ewentualny `ExceptionID` i ślad procedur/wyjątków;
- formę poprawioną jawnie podaną przez użytkownika albo upoważnionego agenta, czas zgłoszenia, identyfikator wpisu korekty i jawne działanie inicjatora zapisu.

Forma bazowa pochodzi z aktywnego snapshotu: z leksykalnego `MorphLexeme`, przypisanej komórki skompilowanego `Paradigm`, jawnych `PrincipalParts` tego leksemu i deklaratywnego planu procedur. Z samego lematu nie wyprowadza się nieregularnej formy. Jeśli generator nie zwróci znanej formy dla wskazanej komórki albo cel pozostaje niejednoznaczny, system odmawia uczenia; nie fabrykuje wartości bazowej.

Dla pierwszego pionowego wycinka bazowa generacja musi wskazać jedną formę dla wybranej komórki. Wiele form, nierozstrzygnięte cechy, `unknown`, `unavailable`, `blocked`, `incomplete` albo niezgodność profilu wymagają jawnego rozwiązania przed korektą. Wygenerowana propozycja modelu sama w sobie nie jest pochodzeniem korekty ani zgodą użytkownika.

Pochodzenie korekty pozostaje odrębne od źródła form bazowych. Zapis przechowuje oba: odnośnik do bazowego snapshotu i śladu oraz autora/źródło poprawionej formy. `reviewed_by` i `reviewed_at` wymagane przez obecny format `morph.ExceptionRule` nie mogą być wypełniane fikcyjną recenzją ani przypisywać decyzji właścicielowi bez jego działania.

## Lokalny overlay

Poprawiona forma jest osobnym, jawnie włączanym rekordem lokalnym. Nie edytuje źródłowych danych językowych i nie aktywuje globalnej reguły. Rekord jest przypięty do dokładnego snapshotu i pełnego adresu komórki powyżej. Zawiera co najmniej identyfikator korekty, formę przed i po, provenance, jawnego inicjatora zapisu, stan (`active`, `undone` lub `conflict`) i powiązania do wcześniejszej korekty, jeśli istnieje.

Runtime może odwzorować aktywną korektę na wąski wyjątek typu `lexeme_cell` dla tej samej komórki, z efektem `replace`. Obejmuje on wyłącznie poprawioną formę dla wybranego leksemu i komórki. Nie używa `stem_override`, zmiany `Paradigm` ani `add_variant`, bo mogłyby zmienić więcej niż jedną komórkę lub utrzymywać równolegle niepotwierdzony wariant. Brak obsługi dokładnego zakresu przez kompilator oznacza odmowę zastosowania korekty, a nie rozszerzenie jej zasięgu.

Po zapisaniu overlayu żądanie tej komórki zwraca skorygowaną formę ze śladem źródła `exception`, ID korekty i jej provenance. Pozostałe komórki nadal korzystają z bazowego paradygmatu i wyjątków. Gdy analizator powierzchni widzi formę poprawioną, zachowuje wszystkie jej odczyty; korekta nie scala homografów ani synkretycznych analiz innych leksemów.

## Konflikty

- Ponowne zgłoszenie identycznej korekty dla identycznego adresu i snapshotu jest idempotentne; nie tworzy drugiej aktywnej reguły.
- Różna forma dla tej samej komórki przy aktywnym overlayu jest konfliktem. Nie ma „ostatni zapis wygrywa”, sortowania po czasie ani automatycznego wyboru.
- Rozwiązanie konfliktu wymaga jawnego działania uprawnionego inicjatora wskazującego aktywny wpis do zastąpienia oraz nową formę. Historia zostaje zachowana, a nowy wpis jawnie wskazuje poprzedni. Jeśli bieżący mechanizm precedencji nie potrafi bezpiecznie zastosować takiego zastąpienia, korekta pozostaje nierozstrzygnięta.
- Zmiana aktywnego snapshotu, profilu, leksemu, paradygmatu, komórki lub wartości cech unieważnia przypięcie. Stary overlay nie przenosi się automatycznie; wymaga ponownej walidacji i jawnej ponownej decyzji albo pozostaje nieaktywny jako `conflict`.
- Istniejący, recenzowany wyjątek lub blokada dla tego samego celu nie może zostać po cichu nadpisana przez lokalną korektę. Konflikt zostaje zgłoszony; `block` nigdy nie uruchamia bazowego fallbacku.

## Cofnięcie i trwałość

Cofnięcie wskazuje konkretny identyfikator aktywnej korekty. Jest osobnym, trwałym zdarzeniem historii, które wyłącza tylko ten overlay; nie usuwa ani nie przepisuje danych bazowych. Przy tym samym snapshotcie generacja po cofnięciu musi odtworzyć poprzedni wynik bazowy. Jeśli snapshot zmienił się od korekty, system nie obiecuje identycznej powierzchni: pokazuje nową bazę i wymaga ponownej decyzji zamiast reaktywować starą korektę automatycznie.

Overlay i jego historia są lokalne dla użytkownika oraz trwale zapisywane poza bazowym snapshotem językowym. Zapis jest atomowy: po awarii użytkownik widzi albo poprzednią kompletną wersję, albo nową kompletną wersję, nigdy częściowy zestaw. Przy starcie system weryfikuje format, provenance, digest bazowy i konflikt selektorów. Niedostępny, uszkodzony, obcy lub niezgodny wpis powoduje jawną odmowę aktywacji całego overlayu z diagnostyką; żaden wpis nie staje się aktywną wiedzą bez pełnej walidacji.

Warunkiem stwierdzenia trwałości jest zamknięcie procesu i uruchomienie go ponownie z tym samym lokalnym magazynem: korekta ma być ponownie odczytana i zastosowana do tego samego adresu komórki. Jeżeli magazyn nie może potwierdzić trwałego zapisu, wynik nie może być zgłoszony jako „nauczony” ani „zapisany”. Ten kontrakt nie przesądza o nowym CLI ani o istniejącej implementacji magazynu.

## Zastosowanie poza przykładem korekty

Kryterium wykorzystania korekty to **inne zdanie lub inny kontekst** z tym samym istniejącym `LexemeID` i dokładnie tym samym `CellID` oraz kompletnym zestawem cech. Po restarcie realizator używa wtedy poprawionej formy, a analizator przypisuje tę powierzchnię do właściwego odczytu. To pokazuje użycie nabytego wpisu poza pierwotnym zgłoszeniem.

Kontrola negatywna sprawdza co najmniej sąsiednią komórkę tego leksemu i ten sam zestaw cech innego leksemu. Nie wolno stosować poprawionej formy do żadnej z nich, chyba że niezależne bazowe dane już ją przewidują. Pojedyncza korekta nie zmienia tematu, odmiany, całego paradygmatu ani reguł dla innych wyrazów.

## Odbiór tego kontraktu

Implementacja spełnia pionowy zakres dopiero po wykazaniu na jednym rzeczywistym leksemie i jednej komórce, że:

1. zgłoszenie zapisuje jednoznaczne wejście, rzeczywistą bazę wraz ze snapshotem i provenance oraz poprawioną formę z jawnym inicjatorem;
2. nakłada aktywny lokalny overlay wyłącznie na adresowaną komórkę i zachowuje pochodzenie w wyniku;
3. odmawia konfliktu dwóch różnych aktywnych form, blokady, niejednoznacznej komórki i niezgodnego snapshotu bez cichego wyboru;
4. po restarcie używa tej samej korekty w innym zdaniu, przy czym analiza/realizacja zachowuje dokładny leksem i cechy;
5. pokazuje, że inne komórki i inne leksemy pozostały przy bazowym wyniku;
6. cofnięcie korekty, także po restarcie, przywraca bazową generację i pozostawia audyt cofnięcia.

Są to kryteria wdrożenia i obserwacji, nie stwierdzenie, że taka ścieżka już działa. Techniczne wyniki agenta nie oznaczają akceptacji językowej właściciela.

## Aktualna podstawa w kodzie

`internal/morph.Generate` przyjmuje jawny `MorphLexeme`, `PrincipalParts`, `FeatureBundle`, skompilowane paradygmaty, skompilowane wyjątki i budżet. `GeneratedForm` zachowuje `LexemeID`, język, POS, `ParadigmID`, `CellID`, cechy, źródło, provenance i ślad. `ExceptionRule` ma selektory do poziomu `lexeme_cell`, jawny `replace`, provenance, reviewer metadata, ograniczone dekodowanie/kompilację i wykrywanie konfliktów. Te mechanizmy dostarczają typowanej podstawy dla ograniczonego override.

Obecny kod nie udostępnia przepływu uczenia poprawki, lokalnego trwałego overlayu ani jego ponownego wczytania po restarcie. Nie ma tu polecenia CLI do uczenia. Opisane wyżej wymagania muszą być zaimplementowane i rzeczywiście zaobserwowane przed ogłoszeniem funkcji; nie należy przedstawiać samego generatora wyjątków jako uczenia.

Powiązane kontrakty: [paradygmaty i generator](../../architecture/07-morphology-procedures-exceptions.md), [wyjątki](../../architecture/morph-exceptions.md), [pochodzenie danych](../../architecture/10-import-provenance.md), [decyzje o danych i snapshotach](../../architecture/03-data-decisions.md) oraz [LIGHT LM](README.md).

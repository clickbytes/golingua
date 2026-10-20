# 10 — Tworzenie danych i opcjonalny import

**PROPOSED, G04/G07.** Piszemy **nowy** system leksykalny i gramatyczny w Go. Nie ma obowiązku importowania obecnych słowników flexLingua ani jego testów. Celem jest około 1000 konceptów z sensami PL/EN, ale najpierw prosty działający model i ręcznie oceniane tłumaczenia.

## Format logiczny danych

1. `ConceptDefinition`: znaczenie, typ, rolę i relacje pojęć.
2. `MorphLexeme`: jeden leksem z lematem, POS i klasą odmiany; działa niezależnie od konceptu.
3. `LexicalSense`: oddzielna jednostka semantyczna, bo jedno słowo może znaczyć wiele rzeczy.
4. `SemanticBinding`: łącze sensu z ConceptID, różne jakości mapowania: exact, broader, narrower, related, candidate. Nie utożsamiaj ich.
5. `Paradigm / PrincipalParts / Exception`: przepis do generowania form.
6. `ValencyFrame / ConstructionBinding`: rekcja sensu i dopuszczone konstrukcje.
7. `ConstructionFamily / Variant / Operator`: deklaracje językowe wielokrotnego użytku.

## Importer to narzędzie, nie wyrocznia

Dopuszczamy w przyszłości import TSV/UniMorph, wybranych otwartych źródeł leksykalnych i ręcznych deklaracji JSON. Obce dane mają **provenance, wersję, licencję i źródło**. Etapy są proste:

`source → bounded parse → normalizacja → propozycja leksemu/sensu/paradygmatu → sprawdzenie spójności → decyzja czy aktywować → immutable snapshot`.

Import słowa bez mapy CORE nie jest błędem: może ono być dostępne w morfologii, a niedostępne w tłumaczeniu. Formy obserwowane ze źródła mogą pomóc zasugerować paradygmat i wyjątek, **ale nie muszą być kopiowane do zbioru „expected”**. Brak komórki w słowniku nie oznacza zakazanej formy. Dwie możliwe klasy odmiany pozostają propozycjami, dopóki nie rozstrzygniemy konfliktu.

Nie ma automatycznego zaakceptowania zdania albo jego tłumaczenia na podstawie danych importera. Właściciel może sam sprawdzić wynik Go i zdecydować o poprawce.

## Licencje i atomowość

Wersje źródeł i warunki redystrybucji są realnymi ograniczeniami. Nie kopiuj zewnętrznych tekstów/definicji bez prawa. Import nie może wykonywać kodu z danych, pisać poza staging root ani częściowo aktualizować aktywnego pakietu w razie awarii.

Do redagowania niewielkich zasobów wystarczy JSON i weryfikacja struktury. SQLite jest opcją późniejszego narzędzia edytorskiego, nie obowiązkiem ani runtime backendem.

## Koniec zadania G04

Rozróżnione typy danych, prosty konfigurowalny importer (gdy jest potrzebny), walidowane i powtarzalne generowanie niemutowalnego pakietu językowego. Nie wymagamy importu dawnych 1002 konceptów, 593 zdań, formalnych reviewer IDs lub test runnera. Językowe wyjścia kontrolujemy później ręcznie z użytkownikiem.

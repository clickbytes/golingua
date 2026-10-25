# A1-M5 #341 — zakres wykonania

Base: `645e9f9255c3ec65548f87a91343fa1566f078ea`; branch: `a1-m5-contrasts`. Zakres obejmuje dokładnie 16 par z `/tmp/golingua-matrix-expansion/m5-selected.json`, czyli 32 kierunkowe rekordy A1.2 o ID `A1.2-<domain>-201…232`.

Kontrasty obejmują cztery wymagania posiadania bez fikcyjnych grafów (GAP_I dla ograniczeń CORE), zdarzenia w liczbie mnogiej, negację, present/past, homografię `read`, przypisanie `black` do referenta obiektu, szyk OVS i niezależność rodzajników od grafu. W analizie angielskiego `The teachers read the message.` wymagane są osobno grafy present i past; realizacja PL dla tych grafów ma odrębne teksty. Polski input w czasie teraźniejszym wymaga tylko grafu present.

Nowe grafy I są zadeklarowane ręcznie z jawnych konceptów, ról, czasu, liczby, scope i referentów; istniejące Reading i Seeing są ponownie użyte tam, gdzie kontrakt jest ten sam. Liczebnik mnogi w angielskim nie otrzymuje `a`; warianty bare/the nie dodają kwantyfikatora `all`. Wszystkie rekordy zachowują źródło koordynatora i `OwnerAcceptance: NOT_EVALUATED`.

Zakres nie zmienia gramatyki, engine, CORE ani CI. Braki i niezgodności pozostają GAP/FAIL.

Korekta kontraktu po niezależnym review: rzeczownik pospolity w pozycji podmiotu gramatycznie ustala osobę trzecią, więc przypadki realizacji nie wymagają założenia `person` jako artefaktu adaptera. Pole `person` pozostaje dozwolone w ograniczonych trasach. Wymagania `determiner` pozostają tam, gdzie wybór angielskiego rodzajnika jest nierozstrzygnięty.

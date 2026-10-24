# LIGHT-F4: koszt współdzielonego runtime

Issue [#328](https://github.com/clickbytes/golingua/issues/328), base `57e4882e79eb6596caed5126108fa272177e6982`, owner koordynator. Zakres dotyczy dwóch miejsc potwierdzonych odczytem i profilem wcześniejszego audytu; najpierw powtarzamy pomiar na aktualnym base, potem zmieniamy kod.

## Zamknięte decyzje

`SelectTargets` nie potrzebuje pełnej kopii dokumentu leksykonu ani dwóch map z całego katalogu przy każdym żądaniu. Zachowuje dotychczasową kontrolę `Len`, budżety, kolejność, wszystkie mapowania i diagnostyki. `ByCoreConcept` dostarcza dotychczasowy defensywny zbiór wiązań; sens jest wyszukiwany nowym `BySenseID` przez binary search w posortowanych prywatnych danych, leksem przez istniejące `ByLexemeID`. Zwracane rekordy nadal mają defensywne kopie slice'ów. Brak nowej mapy/cache i mutowalnego dostępu do indeksu.

CORE i pack są niezmienne przez publiczne API. `Engine.New` oblicza raz zgodność inwentarza konceptów, z szybkim potwierdzeniem dla tego samego wskaźnika CORE; dla różnych wskaźników zachowuje istniejące DeepEqual danych. Prywatne pole przechowuje wynik. `Translate` sprawdza tę wartość w dotychczasowym miejscu i zachowuje `RUNTIME_SNAPSHOT_MISMATCH`, status invalid oraz wszystkie wcześniejsze warunki. Nie zmieniamy niezgodnego setupu w nowy rodzaj błędu konstruktora. Brak CORE/pack nadal daje dotychczasowy requestowy wynik. Walidacja każdego Meaning i Evidence, loader/digest, registry i budżety nie są usuwane ani przenoszone.

## Dowód

Ten sam Go, CPU, GOMAXPROCS=16 i dane dla base/after. Trzy oddzielne procesy: zmierzona inicjalizacja `matrix.NewRuntime` (rzeczywiste kompilowane adaptery i dane Go, bez analizy zdania), alokacje inicjalizacji i heap po GC; potem dla każdego z czterech stałych wejść 10 rozgrzewających i 100 mierzonych wywołań `Engine.Translate`. Wyniki: mediana/p95 czasu, bajty i liczba alokacji/żądanie, statusy i liczba kandydatur. To pomiar, nie procentowy próg ani benchmark w CI. Inicjalizacja nie jest czasem wczytania snapshotu z dysku lub całego procesu CLI; zimny CLI obserwujemy osobno. Osobny przebieg profilu CPU/allocs obejmie załadowane tłumaczenie i wyłączy analizę kosztu inicjalizacji.

Stałe wejścia: `The teacher sees the dog.`, `The teacher reads the message.`, `Nauczyciel nie widzi psa.`, `Czarny kot widzi psa.`. Pierwsze trzy pochodzą z wcześniejszego pomiaru, ostatnie sprawdza dostarczoną konstrukcję. Reprodukcja ma używać jednorazowej kopii bez `.git` i lokalnego tekstu programu pomiarowego; żaden runner ani zależność nie staje się częścią produktu.

Kontrole techniczne, lokalna selekcja I/języków/integracji, pełna macierz z dotychczasowymi36 SPEC_PENDING i40 wejść CLI327 mają zachować wynik. Małe bezpośrednie próby muszą wykazać defensywne kopie indeksu i zachowanie mismatch/missing runtime. Nie powiększamy korpusu, nie zmieniamy języków/Meaning ani CI. TUI nie jest zmieniane; istniejący przepływ obserwowany na #327. Wniosek dotyczy kosztu tego runtime na tych wejściach, nie energii ani porównania z LLM.

# LIGHT-F2: wspólne klasy odmiany

Zakres [#326](https://github.com/clickbytes/golingua/issues/326) zamknięto przed implementacją na base `1707590ead2892f16f065fe52eaaf935e9e9216e`. Reguły są danymi języka wykonywanymi przez istniejący interpreter procedur. Nie zmieniamy CORE, znaczeń, parsera, realizatora ani backendu morfologii.

## Polski

Jawne przypisania `pl.szkola.n`, `pl.ryba.n`, `pl.rzeka.n`, `pl.mewa.n` wskazują profil `feminine_a_partial`. Nie klasyfikujemy pozostałych słów przez rodzaj lub samą końcówkę. Sześć komórek wybiera principal part `sg_nominative`, usuwa dokładne końcowe `a` i dodaje wspólną końcówkę:

| Komórka | Końcówka | Przykład |
| --- | --- | --- |
| sg_accusative | ę | szkołę |
| sg_instrumental | ą | rybą |
| sg_vocative | o | mewo |
| pl_dative | om | rzekom |
| pl_instrumental | ami | mewami |
| pl_locative | ach | rzekach |

Pozostałe osiem komórek pozostaje jawne: mianownik, dopełniacz, celownik i miejscownik liczby pojedynczej oraz mianownik, dopełniacz, biernik i wołacz liczby mnogiej. W szczególności `szkół`, `szkole`, `rzece`, `rybie` i `mewie` nie wynikają z nowej reguły. To częściowy profil, nie kompletna polska deklinacja.

Usuwamy 24 powielone pełne formy z danych produktu. Pusty slot jest dozwolony tylko w sześciu komórkach przypisanej klasy; nie staje się pustym principal part. Brak wymaganej jawnej formy i podanie duplikatu formy proceduralnej są błędami danych. Operacja `strip_suffix` nadal odmawia przy braku dokładnego `a`.

`szkoła` i `ryba` przechodzą z ogólnego paradygmatu jawnych form żeńskich na wspólny `pl_noun_feminine_a_partial`, ze swoim dotychczasowym authored provenance. `rzeka` i `mewa` zachowują indywidualne ID paradygmatów/procedur oraz dotychczasowe źródła WSJP PAN z licencją `unknown`. Wszystkie ID leksemów i komórek pozostają bez zmian. Źródła poświadczają formy; wybór sześciu wspólnych operacji jest decyzją implementacyjną agenta. `mewa` pozostaje leksemem morfologicznym bez nowego bindingu CORE.

Przypisanie następnego odpowiedniego leksemu wymaga danych klasy i ośmiu jawnych form, a nie nowego warunku w interpreterze. Nie deklarujemy jeszcze, że osiem form można zawsze zredukować do samego lematu.

## Angielski

`en.look.v` używa istniejącej `en_verb_finite_regular`, tak jak `print`, `cook`, `clean`, `plant`, `harvest`, `play`, `earn` i `act`. Zostaje jedna część `base=look`; procedury dają `look`, `looks`, `looked` i `looking`. Zachowujemy wszystkie dziesięć dotychczasowych komórek finite/imperative/participle z identycznymi cechami, w tym osobne odczyty past i past participle dla `looked`. Zmieniają się ID przypisanego paradygmatu i ślady procedur. Nie zmieniamy osobnej klasy `en_verb_regular` ani nieregularnych `see/hear/get/read`.

## Odbiór

Koordynator porównuje rzeczywiste 56 form PL i 10 komórek EN z obserwacją base, sprawdza analizę wszystkich ich powierzchni i ślady nowych procedur. Kontrole obejmują istniejące `print/cook/cat/dog`, `pies/kot/ptak`, zdania korpusu S051 w obu kierunkach i negację trasy patrzenia. Dodatkowe zdania z nowymi kombinacjami istniejących rzeczowników są oznaczone jako próby agenta, nie wejścia właściciela ani pełne A1.

Obowiązują dotychczasowe lokalne testy macierzy, gofmt, vet, build, smoke oraz rzeczywiste CI PR i końcowego main. Pełna macierz zachowuje niezaliczone wymagania innych domen. Nie dodajemy test runnera, masowego korpusu ani bramki językowej do CI. Akceptacja językowa właściciela: NOT_EVALUATED.

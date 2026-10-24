# LIGHT-F3: referent i cecha w złożonym znaczeniu

Owner kontraktu: koordynator Codex, [#327](https://github.com/clickbytes/golingua/issues/327), base `11772b75729ac1ef1f7dc1f4abff51a5f7269363`. Decyzja po odczycie rzeczywistego parsera, kompozytora, konstrukcji i realizatora. Jeden PR obejmuje poniższy pionowy zakres.

## Identyfikacja

Istniejące `class_membership(member=r)` oraz `black(bearer=r)` nie wymagają `concept:cat`. Referent ma własny ConceptRef; jego tożsamość, cechy i kontekst muszą odpowiadać temu samemu węzłowi. Klasyfikacja nie staje się identycznością dwóch osobników ani faktem CORE. Zachowujemy ograniczenia present/one, zakres negacji i tryby proximal/demonstrative/unresolved.

Realizacja nadal wymaga zatwierdzonego wiązania leksemu, zgodnej klasy liczby i dostępnej komórki morfologicznej. Angielskie klasyfikujące `a/an` wymaga rzeczownika policzalnego; ogólność ConceptRef nie uprawnia do `a water`. Polski `czarny` ma obecnie wyłącznie mianownik liczby pojedynczej klasy masculine_animal. Nie dodajemy form innych rodzajów ani liczby mnogiej; brak zgodnej formy jest jawną odmową.

Obecne szablony angielskiej identyfikacji emitują stałe `a`, dlatego ta konkretna trasa wymaga także jawnego `indefinite_article=a`. Brak danych lub wymagane `an` nie jest licencją na błędny rodzajnik; nowy wariant identyfikacji z `an` pozostaje poza tym zakresem.

Obowiązkowe przykłady: `This is a dog` / `To jest pies.`, `This dog is black.` / `Ten pies jest czarny.`, `This is not a dog.` / `To nie jest pies.`, dotychczasowe koty oraz drugi zgodny znany rzeczownik. Nie dodajemy warunków specjalnych dla psa.

## Cecha uczestnika zdarzenia

Pierwsza złożona konstrukcja to dodatnie widzenie w teraźniejszości z jedną cechą podmiotu: `The black cat sees the dog.` / `Czarny kot widzi psa.`. Warianty zamieniają istniejące zgodne rzeczowniki (pies, ptak, kot); nie dodają słownika. Dwa miejsca referentów pozostają oddzielne również dla `The black cat sees the cat.` — równy ConceptID nie dowodzi tożsamości ani jej różności w świecie.

Wzorcowe I używa istniejących typów, bez zmiany Meaning schema:

```text
roots = [e, q]
e: event seeing, temporal_frame=present
   experiencer -> r1; stimulus -> r2
q: state black, temporal_frame niepodany
   bearer -> r1
r1: referent cat, multiplicity=one
r2: referent dog, multiplicity=one
```

Cecha nie ma własnego jawnego czasu w tej frazie. Nie dopisujemy jej czasu ani nie gubimy jawnego czasu dostarczonego przez inne I. Oba korzenie należą do jednego znaczenia wejścia; żaden nie jest osobną alternatywą tłumaczenia. Nie dodajemy aktu `statement`, trwałego faktu, płci, określonej tożsamości, deiktycznego wskazania ani koloru wszystkich kotów. Zmiana `bearer` na `r2` oznacza inne znaczenie; realizator nie może wtedy umieścić przymiotnika przy podmiocie.

Pierwsza trasa obsługuje dokładnie jedno zdarzenie seeing, jeden nieczasowany stan black, dwa miejsca referentów o multiplicity=one i dwa korzenie. Główne zdarzenie ma present. Przy negacji tej złożonej frazy, drugim modyfikatorze, czasowanym stanie cechy, cesze dopełnienia lub nieobsługiwanej zgodzie odmawiamy realizacji, zamiast pomijać treść. Dotychczasowa prosta negacja identyfikacji/widzenia pozostaje obsługiwana. Nie deklarujemy dowolnych zdań złożonych.

Kontekst złożonej trasy może być niepodany lub zawierać poprawne, nierozstrzygnięte wzmianki odpowiadające temu samemu I. Jawne wskazanie, rozstrzygnięta tożsamość lub sprzeczny graf kontekstu wymagają innej realizacji; nie wolno ich zignorować w tej ograniczonej konstrukcji.

## Minimalny kontrakt konstrukcji

`SlotSpec.ModifierBinding` jest opcjonalną deklaracją `{target_slot, semantic_role}`. W tym zakresie może wystąpić najwyżej raz w rodzinie: wymagany pojedynczy AP wskazuje wymagany pojedynczy NP z argument binding. Nie może wskazywać siebie, nieznanego slotu, VP, PP ani innego modyfikatora. Nie jest argumentem głównego zdarzenia.

Dla nowej rodziny AP ma `target_slot=subject`, `semantic_role=bearer`, a NP subject jest związany z rolą `experiencer`. Te role **nie są sobie równe**: `bearer` należy do stanu właściwości, `experiencer` do widzenia. Stan powstaje wyłącznie z zatwierdzonego równoważnego bindingu sensu AP do właściwości CORE z licencjonowaną rolą. Brak lub konkurencja sensów nie może zostać rozstrzygnięta kolejnością wpisów.

Głównym predykatem pozostaje jeden slot VP/AP bez ModifierBinding. Parser używa skompilowanej sekwencji `article AP NP VP article NP .` (EN) / `AP NP VP NP .` (PL), a nie rozpoznawania całych zdań. Kompozytor odczytuje powiązanie z dokładnej rodziny w registry, łączy stan z węzłem referenta wybranego dla wskazanego slotu i zachowuje jego dowody. Nie trzeba dodawać semantyki do `parser.SlotBinding` ani sklejać referentów po pojęciu.

Deklaracja musi być walidowana, głęboko kopiowana i uwzględniona w tożsamości planu/układu ról. Istniejące rodziny bez tego pola zachowują zachowanie i identyfikatory. Główny predykat jest odróżniany od modyfikatora na poziomie slotu, bez globalnego pozbawiania AP statusu kategorii predykatywnej.

Realizator wybiera leksykalizację stanu dla AP i dokładnego referenta dla NP. Powtarzające się licencje wariantów tego samego leksemu/sensu nie są odrębnymi wyborami leksykalnymi; różne leksemy lub sensy nadal wymagają jawnej niejednoznaczności. Generacja AP przechodzi przez zwykłą morfologię i linearizację skompilowanego planu. Cały plan musi pokrywać oba predykaty; samo zdanie bez koloru oraz sama atrybucja bez widzenia nie są poprawnymi powierzchniami tego I.

PL bierze zgodę AP z dokładnego NP podmiotu. EN zachowuje znaną określoność; bez niej pokazuje licencjonowane warianty `a/the black ...` i rodzajnika dopełnienia z jawnymi założeniami. Rodzajnik podmiotu odnosi się do początku frazy `black`, więc nie dziedziczy `an` od późniejszego rzeczownika.

## Odbiór

Niezależne, ręcznie zapisane I oraz testy analizy i realizacji obu języków mają sprawdzać oba predykaty, role, czas, wielość i brak rozstrzygnięcia tożsamości. Kontrprzykłady: zgubiona cecha, błędny bearer, zamiana ról, dwa nierozstrzygnięte referenty tego samego ConceptID, negacja i nieobsługiwane cechy. Bez parsera języka drugiego jako źródła wzorca. Testy językowe pozostają lokalne.

Rzeczywiste CLI i TUI mają pokazać nowe zdania i wcześniejszą identyfikację, znaczenia pośrednie, kandydatury, założenia oraz diagnostyki. Dotychczasowe kontrole techniczne i językowe muszą zachować wynik; pełna macierz nadal ujawnia SPEC_PENDING innych domen. Koordynator wykonuje odbiór, przegląd diffu, PR, merge i sprawdza rzeczywiste CI main. Akceptacja właściciela pozostaje NOT_EVALUATED.

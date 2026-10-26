# Wiązania leksykalne z typowanymi celami

Ten kontrakt dodaje jawny, ograniczony most między sensem leksykalnym i typowanym celem lub wyrażeniem. Wersja 1 używa `positive-facts-v1`; nie uruchamia parsera, generatora zdań ani ogólnego mapowania CORE.

## Pakiet językowy

Pakiet v1 podaje język, leksykon, sensy, ramy semantyczne i wiązania. Wiązanie wskazuje ramę i jawnie mapuje każdą używaną rolę semantyczną na port symbolu albo port celu. Odczyt `inspect` pokazuje zapis pakietu, ale zwraca `metadata_verified: false`. `validate` wymaga jawnie wybranego pakietu metadanych i sprawdza podpisy oraz bramkę autoryzacji osobno.

Role semantyczne ramy, jej funkcje gramatyczne (np. podmiot i dopełnienie), porty celu i role argumentów symbolu należą do odrębnych warstw. Ich zgodność wynika wyłącznie z wpisanych mapowań i typów. Nazwy, opisy i etykiety nie sterują wyborem ani nie tworzą aliasów. Pakiet może przechowywać cele oczekujące lub szerokie do inspekcji i walidacji; sama ich obecność nie daje prawa wykonania. Bramka instancjacji wymaga konkretnej zatwierdzonej deklaracji albo jawnie zatwierdzonego celu równoważnego. Nie można wykonać celu oczekującego, szerokiego ani nierozstrzygniętego.

Osobny typ wiązania `reference_predicate` mapuje jawny symbol jednoargumentowy do referencji bez ramy i bez tworzenia fikcyjnej walencji. Nadal wymaga on jawnego typu i identyfikatora portu referencji oraz przechodzi tę samą kontrolę pakietu i autoryzacji. Nie zmienia formatu ani bramki istniejących wiązań ramowych i `CoreRef`.

Wiązanie do symbolu tworzy typowany dokument: atom, gdy wszystkie jego porty są jawnie przypisane, albo funkcję z otwartymi portami. Port symbolu niepokryty rolą ramy można zamknąć osobną referencją z żądania; bez niej pozostaje typowanym argumentem funkcji. System nie wymyśla uczestników ani zdarzeń, a niezgodny typ referencji jest odrzucany. Wiązanie do celu zachowuje jego porty i ograniczenia.

Opcjonalny `binding_index` wskazuje pozycję w uporządkowanej liście wszystkich celów danego sensu, zanim wybrany cel zostanie sprawdzony względem ramy. Bez wyboru kilka celów pozostawia przypadek `ambiguous`; jawny indeks wybiera konkretną pozycję, a niezgodność wybranego celu z ramą jest odrzucana. Żądanie i wynik zachowują identyfikatory sensu, ramy, kontekstu oraz wybrany cel.

## Kontekst i kontrola celu

Cel bez referencji i bez przypiętego kontekstu można użyć jako szablonu: instancjacja podstawia jawny `context_id` żądania do konkretnego celu, nie zmieniając zachowanego szablonu. Jeśli szablon zawiera referencje albo własny kontekst, identyfikator kontekstu żądania musi się z nim zgadzać. Identyfikator kontekstu wyznacza granicę żądania; sam nie uwierzytelnia tożsamości ani nie dowodzi faktów.

Pakiet językowy jest ograniczony do 8 MiB, żądanie instancjacji i specyfikacja celu do 1 MiB, a certyfikat weryfikacji do 8 MiB. Dodatkowe limity weryfikacji ograniczają liczbę operacji i pracę; limity termu ograniczają jego węzły i głębokość. Przekroczenie limitu zwraca odmowę/limit, a nie zgodę na niepełny wynik.

Wynik symboliczny jest typowanym wyrażeniem. Wynik celu zawiera konkretny cel, konstrukcję, interfejs i certyfikat. `verify-goal` niezależnie sprawdza typ, doprecyzowanie celu i odtworzenie operacji. `goal_satisfied` oraz `certificate_valid` opisują wyłącznie sprawdzenie strukturalne profilu; nie oznaczają prawdy o świecie, dowodu istnienia faktów, wyłączności interpretacji ani jakości językowej. Bramkę autoryzacji formalnej i jej pochodzenie należy odróżniać od akceptacji językowej właściciela.

Format i autoryzacja istniejących wiązań `CoreRef` pozostają bez zmian. Nowy CLI nie ładuje CORE dla instancjacji `CoreRef`, więc ścieżka wymagająca takiego celu nie ma tu obsługiwanej ścieżki legacy ani automatycznego fallbacku. Nierozstrzygnięty cel pozostaje nierozstrzygnięty.

## Ograniczony pilot

`data/semantic/examples/lexical-bindings/` zawiera sześć przykładów: cztery ramy percepcji PL/EN oraz dwa oddzielne, techniczne użycia angielskiego `link`. Pakiety PL i EN wiążą `widzieć`/`see` z `pilot.vision_event`, a `słyszeć`/`hear` z `pilot.auditory_event`. Role `experiencer` i `stimulus` są mapowane jawnie na porty typu `Entity`; port `event` wymaga osobnej referencji typu `Event`. Bodziec typu `Event` jest odrzucany w tym pilocie.

Angielskie przykłady `link.direct` i `link.indirect` są nazwami neutralnej relacji matematycznej `pilot.r`, nie ogólną analizą naturalnego słowa „link”. Pierwszy cel opisuje jedną uporządkowaną krawędź, drugi jawnie dwukrawędziową ścieżkę z pośrednim bytem `Entity`. Przykłady nie dodają aksjomatów ani faktów.

Plik `source-records.json` przechowuje cztery wcześniejsze powiązania CORE i pochodzenie leksykalne/ram. Nie są one aktywnym celem nowych wiązań. Warto też zachować rozróżnienie w historii: wiersze `see` i `hear` w `data/en/1000-core-review.tsv` mają kategorię `EXISTING_PENDING`, choć ich opis zawiera notę o delegowanym przeglądzie i zatwierdzonym wiązaniu CORE. Pilotowe deklaracje formalne i wiązania są odrębnymi wpisami. Ta rozbieżność nie jest rozstrzygana przez samo uruchomienie CLI. Pilot formalny i nowe wiązania są oceną agenta dla tego ograniczonego przykładu; akceptacja językowa właściciela pozostaje **NOT EVALUATED**. Pakietów nie należy traktować jako zatwierdzenia innych sensów, rodzin konstrukcji, odmiany, parsowania lub tłumaczenia.

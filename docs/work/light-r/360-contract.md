# R6 child children wiek i potomek

Aktywny [#360](https://github.com/clickbytes/golingua/issues/360), [epik #354](https://github.com/clickbytes/golingua/issues/354). Base bb5614225694f881f4ef9b584d20c0e40b55f400. Jeden writer, małe delegowane zlecenia; koordynator integruje i ocenia. Akceptacja językowa właściciela NOT_EVALUATED.

Na istniejących trasach zdań EN child/children oraz PL dziecko/dzieci zachowują uzasadnione odczyty `concept:child` (wiek) oraz `concept:child_of_person` (potomek bez ograniczenia wieku). Osobne sensy i wiązania, bez dodawania węzła rodzica lub relacji do wymyślonej osoby. Wielość i role są cechami danego wystąpienia referenta; różne wystąpienia nie stają się jednym węzłem.

Samodzielny I każdego odczytu ma dać poprawne PL/EN, w liczbie pojedynczej i mnogiej tam gdzie obecna konstrukcja obsługuje tę liczbę. Odczyty source zostają jawnie alternatywne. Docelowa powierzchnia może używać istniejącego offspring dla potomka, bez nieuzasadnionego rozszerzania znaczenia age child. Dobór determinera i formy nie zgaduje liczby; synkretyzm offspring lub dzieci nie usuwa rzeczywistych odczytów.

Zakres: potrzebne lokalne dane lexical/morphology i wykazane błędy istniejących adapterów. Bez nowych domen, predykatów, WH/HAVE, rodzicielskiej konstrukcji dopełniaczowej, uczenia, redesignu ani publicznego API/schematu I. Inne pending senses pozostają bez automatycznej akceptacji.

Odbiór: focused kontrole istniejących autorskich I, niezależna analiza/realizacja, realny snapshot i CLI na istniejącym korpusie, gofmt/vet/build/help/version, pełna obserwacyjna macierz, przegląd diffu i rzeczywiste CI PR/main. Ocena agenta jest osobna od akceptacji właściciela.

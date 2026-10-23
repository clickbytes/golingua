# 08 — MeaningGraph, wiedza i docelowe rozmowy

**Częściowo wdrożone (G01-04/G01-05/G06).** `internal/meaning` definiuje bounded request-scoped Meaning IR, cross-CORE role validation i fail-closed asserted projection. G06 dodaje request FactSet, bounded reasoner, typowane pytanie CLI, ContentPlanner i ograniczoną realizację dodatniej odpowiedzi yes/no przez istniejące trasy PL/EN. Koncept to typ znaczenia, a MeaningGraph to konkretna interpretacja zdania.

## Oddzielne pojęcia

- `ConceptDefinition(event.read)`: definicja sensu i ról, nie zdarzenie w świecie.
- `LexicalSense`: znaczenie słowa w języku, związane z konceptem.
- `MeaningGraph(Event(concept:event.read,agent:person_1,theme:book_2))`: opis *tej wypowiedzi*.
- `DialogueAct`: czy użytkownik pyta, stwierdza, prosi o informację, odmawia itd.
- `FactSet`: fakty podane jawnie w request, nie pamięć długoterminowa, nie twierdzenie o całym świecie.

## Operator scope

Negacja, pytanie, możliwość i zdanie warunkowe **nie** są automatycznymi asercjami dodatnimi. `Not(Possible(A))` nie musi znaczyć `Possible(Not(A))`. Uzgodnienie powierzchni PL/EN nie może zmienić operatora semantycznego, referenta, czasu ani przypisania ról.

Osoba/zwierzę/obiekt `animal.dog` jest pojęciem, a `dog_1` i `dog_2` to dwie możliwe instancje. Identyczna nazwa leksemu lub koncept nie jest dowodem tożsamości. Nie domyślamy się bez danych, do kogo odnosi się zaimek.

## Dopasowanie znaczenia

Twarde typy i wymagane relacje mogą odrzucać niezgodny sens. Relacje `typical` mogą co najwyżej porządkować kandydatów; nie powinny tworzyć pewnego faktu ani usuwać innych legalnych sensów. Niepełne poszukiwanie ma pozostać niepełne.

Przy PL szyku OVS i SVO role uczestników mogą pozostać identyczne, gdy rozróżnia je fleksja. EN nie ma tych samych swobód. To trzeba uwzględnić przy budowaniu MeaningGraph.

## Reasoning i odpowiedzi

`Typed Question + jawnie dostarczony FactSet → mały ograniczony reasoner → ContentPlanner → MeaningGraph odpowiedzi → LanguageConstruction → opcjonalny tekst`.

Odpowiedź `true` tylko gdy wynika z jawnego faktu/dozwolonej reguły, `false` wymaga dowodu negatywnego, brak faktu daje `unknown`; sprzeczności ujawniają konflikt. Reasoner nie musi znać historii rozmowy ani wyszukiwać faktów w internecie. CLI `question` nie parsuje pytań PL/EN. Pozytywne odpowiedzi yes/no mogą korzystać z istniejących zatwierdzonych tras oznajmujących; brakujący czas i niektóre cechy powierzchni są jawnie raportowanymi założeniami. Negacja i samodzielne frazy NP pozostają bez tekstu, dopóki język nie dostarczy odpowiednich zatwierdzonych konstrukcji. Wynik `unsupported` zachowuje typową odpowiedź i dowody.

`ContentPlanner` wybiera, co powiedzieć; `ConstructionGraph` wybiera jak sformułować. Nie umieszczaj słów polskich/angielskich w neutralnym planowaniu znaczenia.

## Polityka ręcznej pracy

Nie twórz masowych danych sytuacyjnych, sztucznych pytań ani automatycznej macierzy reasoning. Właściciel delegował agentowi ograniczoną ocenę na dostępnych przykładach oraz jawnie oznaczonych faktach demonstracyjnych; nie są one danymi właściciela ani odbiorem właściciela. To nie zwalnia z walidacji operatorów i ograniczenia kosztu parsera/reasonera. Go v0 **nie posiada samouczenia ani pamięci**.

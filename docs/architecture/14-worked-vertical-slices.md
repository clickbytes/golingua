# 14 — Przykłady działania nowej architektury (ilustracje)

**Nie są to obowiązkowe testy ani przykłady, które należy automatycznie uruchamiać.** Właściciel poda własne wejścia w trakcie prac. Poniższe ilustracje służą wyłącznie wyjaśnieniu typów i kierunku przepływu.

## A. Przechodnia konstrukcja percepcji

`event.visual_perception` jest konceptem. `pl.widziec.1` to leksem; jego konkretny sens wiąże się z konceptem. `ValencyFrame` określa, że w dopuszczonej konstrukcji agent jest kontrolerem zgody, a temat jest dopełnieniem o wymaganym przypadku. `ConstructionFamily(pl.transitive.active)` potrafi złożyć SVO, a po dodatkowych warunkach też OVS.

`Dziewczyna widzi psa` i `Psa widzi dziewczyna` mogą mieć tego samego agenta i temat. `The dog sees the girl` nie jest automatycznie tym samym znaczeniem. To nie znaczy, że dzisiejszy bootstrap umie tłumaczyć te zdania.

## B. Nieregularne formy

EN `go` z żądanym `past` wybiera zatwierdzony temat/formę `went`, nie wymyśla `goed`. PL `pies` ma formy typu `psa`, które mogą wymagać jawnej alternacji tematu. W obu przypadkach ważne są paradygmat i wyjątek, nie osobne koncepty `PAST` pod słowem.

## C. Składanie NP/VP/PP

ConstructionGraph ma podplany: NP uwzględnia liczbę i przypadek, VP zgodę i czas, PP wybiera preposition/rection z walencji. Nie przechowujemy sztywnego zdania dla każdego czasownika i rzeczownika. `see` / `look at` albo `widzieć` / `patrzeć na` mogą mieć różne wymogi.

## D. Pytanie, negacja i czas

`Jutro` nie jest automatycznym czasem przyszłym w samym ConceptGraph. Pytanie o zdarzenie nie dowodzi, że zdarzenie zaszło. Negacja nie jest tylko powierzchniowym znakiem dodanym do tekstu.

## E. Gdy dojdzie reasoner

Jeśli użytkownik dostarczy fakt `dog_1 located_in room_3` i pyta „Gdzie jest pies?”, reasoner może zaproponować odpowiedź z referentem `room_3`, pod warunkiem zgodności jawnych faktów i reguł. Sam concept `animal.dog` nie zawiera lokalizacji konkretnego psa.

## Jak z nich korzystać

Nie dodawaj tych zdań jako masowego zestawu automatycznych sprawdzianów. Po wdrożeniu odpowiedniej funkcjonalności właściciel poda konkretne PL/EN wejście, uruchomimy aktualne Go CLI i ocenimy obserwowany rezultat. Jeśli potrzebna jest poprawka, modyfikujemy leksem, walencję, konstrukcję lub generator zamiast dodawać całe zapamiętane tłumaczenie.

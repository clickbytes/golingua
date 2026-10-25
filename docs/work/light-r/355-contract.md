# R1 negacja i przypadek dopełnienia

Issue #355, epik #354. Base e41b1d124fdd4b177c720964b88fc56a7493a664. Owner: koordynator Codex. Akceptacja językowa właściciela NOT_EVALUATED.

Seeing i hearing z bezpośrednim dopełnieniem w zatwierdzonych teraźniejszych trasach PL wymagają dopełniacza pod negacją. Dodatnie konstrukcje zachowują biernik. Patrzeć na zachowuje rząd na + biernik także z nie. Role experiencer/stimulus, liczba referentów i zakres negacji nie zmieniają się przy wyborze przypadka.

Ramy negatywne są osobnymi lokalnymi ramami tego samego sensu, z dokładnym bindingiem do istniejącej transitive_negative; wspólny BindFrame nadal odrzuca sprzeczne deklaracje. Nie dodajemy globalnego wyjątku od walidacji case ani reguły negacja→genitive dla każdego dopełnienia. Nowe ramy muszą być legalne w leksykonie, providerze, parserze i realizatorze.

Obowiązkowe kontrasty z istniejących form: Nauczyciel nie widzi książki (book one, nie acc plural), The teacher does not see the book → Nauczyciel nie widzi książki; Matka nie słyszy ptaków; dodatnie widzi książkę i ujemne nie patrzy na książkę. Pies/psa jest kontrolą synkretyzmu, nie jedynym dowodem rządu. Pending book.work pozostaje unresolved; nie zatwierdzamy sensu ani nie wymuszamy OK kandydata.

JOB BRIEF Luna Low: jedyny writer, languages/pl/valency.go i tests/languages/pl/A1/negation_case_test.go; mały zamknięty zakres, jeden przebieg i najwyżej jedna korekta. Przed zmianą dodatkowej ścieżki wykonawca zgłasza zależność koordynatorowi. Bez GitHub, commita, nowej gałęzi, runnera i innych danych. Przekazanie: diff, polecenia i rzeczywiste obserwacje. Koordynator prowadzi review, CLI, wymagane kontrole, PR i merge.

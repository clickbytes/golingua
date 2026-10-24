# Zlecenia LIGHT-F2

Issue [#326](https://github.com/clickbytes/golingua/issues/326), base `1707590ead2892f16f065fe52eaaf935e9e9216e`, gałąź `light-f2-procedures`. Owner: koordynator Codex. Przed edycją sprawdzono main, brak otwartych PR i DONE #325 z CI main SUCCESS; zakres zapisano w komentarzu issue. [Kontrakt](326-contract.md) posiada koordynator.

Dzisiejsza bramka OpenCode nie została spełniona, zgodnie z [rejestrem](README.md); użyto dopuszczonego Luna Low. Jeden writer naraz, pomocnik nie wykonuje commitów, GitHub ani zmian gałęzi. Każdy brief dopuszcza jeden przebieg implementacji i jedną nazwaną korektę po review; drugi błąd, niezamknięta semantyka lub potrzeba wyjścia poza dozwolone pliki zatrzymuje zadanie. Zakazane są zmiany oczekiwań, zakresu, CORE, parsera, realizatora, CI i danych spoza briefu; istniejący lokalny `store/` pozostaje poza pracą.

## Rozpoznanie i decyzja

Dwa ograniczone odczyty PL i EN wskazały konkretne komórki, źródła, dostępne operacje i polecenia odbioru. CLI: NOT RUN w rozpoznaniu. Koordynator zweryfikował pliki i skorygował dwie nieścisłości: szkoła/ryba mają authored provenance, a przypisana angielska klasa finite ma dziesięć komórek, nie pięć komórek osobnej starszej klasy. Nie kopiowano tych nieścisłości do kontraktu.

## A — EN look

Writer: Luna Low. Wyłącznie `languages/en/lexicon.go` (`englishParadigmByLexeme`) i `languages/en/morphology.go` (`inCodeMorphRecords`). Zmienić przypisanie `en.look.v` na `en_verb_finite_regular`, usunąć pięć jawnych principal parts, zachować powierzchnie i odczyty. Wejścia: `look`, `looks`, `looked`, `looking`. Wynik CLI: generacja past `looked`, analiza dwóch odrębnych odczytów `looked`, ślad wspólnej procedury.

Przekazanie: zmieniono tylko dwa dozwolone pliki, bez korekty. `gofmt`, `go build -o /tmp/golingua-light-f2-en ./cmd/golingua` i CLI forms PASS. Próba `verb_form=base` dała UNKNOWN_PARADIGM_CELL; poprawne zapytanie finite present dało `look`. Vet i testy NOT RUN przez workera, pozostawione odbiorowi koordynatora. Niezależny odczyt review potwierdził równość dziesięciu ID/cech komórek i brak konsumenta uzależnionego od starego paradigm ID.

## B — PL sześć wspólnych komórek

Writer: ten sam Luna Low, po zakończeniu A. Wyłącznie `languages/pl/morphology_attested_nouns.go`; opcjonalny nowy `morphology_noun_classes.go` nie był potrzebny. Symbole: dane przypisania klasy, mapa końcówek, `attestedNounForms`, rejestracja i budowa paradygmatów, `attestedNounRecords`. Dokładne cztery leksemy, sześć operacji i osiem zachowanych form określa kontrakt. Usunąć 24 literały, walidować brak jawnej formy/duplikat proceduralnej; zachować źródła i ID według kontraktu.

Przekazanie: jeden zmieniony plik, bez korekty. `gofmt`, `git diff --check`, `go build -o /tmp/golingua-light-f2 ./cmd/golingua` PASS. Generacja `szkołę`, `rybą`, `rzekach`, `mewami`, `szkół`, `rzece` i analiza `rzekach` PASS; formy proceduralne mają `select_stem → strip_suffix → append`. Vet i testy NOT RUN przez workera. Pełny odbiór, zakres diffu, zapis dowodów, commit, PR i merge należą do koordynatora.

# A1.9 — wybór 58 par pytań

Dobór koordynatora z 2026-10-25; propozycje wejść i przekładów, akceptacja właściciela `NOT_EVALUATED`. Dane wykonawcze zachowują odrębne kontrakty źródłowe EN/PL. Operator pytania obejmuje pełną propozycję; nie stwierdzamy odpowiedzi ani faktu. Ograniczenie reasonera nie dowodzi braku reprezentacji w Meaning.

| Para | PL | EN | Zjawisko | Specyfikacja |
|---|---|---|---|---|
| 1 | Czy to jest pies? | Is this a dog? | classification | dog:one:proximal_unresolved |
| 2 | Czy tamto jest okno? | Is that a window? | classification | window:one:distal |
| 3 | Czy te koty są czarne? | Are these cats black? | property | cat:many:proximal |
| 4 | Czy tamte książki są czarne? | Are those books black? | property | book:many:distal |
| 5 | Czy pies śpi? | Does the dog sleep? | event | sleeping:sleeper=dog:one:present |
| 6 | Czy kot jadł chleb? | Did the cat eat bread? | event | eating:eater=cat:one,food=bread:mass:past |
| 7 | Czy lekarz widzi ptaka? | Does the doctor see the bird? | event | seeing:experiencer=doctor:one,stimulus=bird:one:present |
| 8 | Czy nauczyciel nie słyszy telefonu? | Does the teacher not hear the phone? | negative_event | hearing:experiencer=teacher:one,stimulus=telephone_device:one:present |
| 9 | Czy dzieci nie śpią? | Do the children not sleep? | negative_event | sleeping:sleeper=child:many:present |
| 10 | Czy książka nie jest czarna? | Is the book not black? | negative_property | black:bearer=book:one:present |
| 11 | Kto widzi psa? | Who sees the dog? | wh_role | seeing:experiencer |
| 12 | Kogo widzi pies? | Whom does the dog see? | wh_role | seeing:stimulus:person_variable |
| 13 | Kto słyszy kota? | Who hears the cat? | wh_role | hearing:experiencer |
| 14 | Kogo słyszy matka? | Whom does the mother hear? | wh_role | hearing:stimulus:person_variable |
| 15 | Kto je jabłko? | Who eats the apple? | wh_role | eating:eater |
| 16 | Co je dziecko? | What does the child eat? | wh_role | eating:food |
| 17 | Kto pije wodę? | Who drinks water? | wh_role | drinking:drinker |
| 18 | Co pije pies? | What does the dog drink? | wh_role | drinking:liquid |
| 19 | Kto czyta książkę? | Who reads the book? | wh_role | reading:reader |
| 20 | Co czyta matka? | What does the mother read? | wh_role | reading:material |
| 21 | Kto daje jabłko dziecku? | Who gives the apple to the child? | wh_role | giving:giver |
| 22 | Komu matka daje książkę? | To whom does the mother give the book? | wh_role | giving:recipient |
| 23 | Co nauczyciel daje dziecku? | What does the teacher give the child? | wh_role | giving:theme |
| 24 | Kto pisze wiadomość? | Who writes a message? | wh_content | writing_content:writer |
| 25 | Komu nauczyciel pisze? | To whom does the teacher write? | wh_role | writing_message:recipient |
| 26 | Gdzie pies śpi? | Where does the dog sleep? | wh_location | sleeping:location |
| 27 | Gdzie jest książka? | Where is the book? | wh_location | located_in:location |
| 28 | Gdzie jest samochód? | Where is the car? | wh_location | located_in:location |
| 29 | Skąd pochodzi ten lekarz? | Where is this doctor from? | wh_origin | origin_from:source |
| 30 | Dokąd idzie dziecko? | Where is the child walking? | wh_destination | walking:destination |
| 31 | Ile kotów widzi matka? | How many cats does the mother see? | wh_count | seeing:stimulus |
| 32 | Ile jabłek je dziecko? | How many apples does the child eat? | wh_count | eating:food |
| 33 | Ile książek czyta nauczyciel? | How many books does the teacher read? | wh_count | reading:material |
| 34 | Ilu nauczycieli widzi pies? | How many teachers does the dog see? | wh_count | seeing:stimulus |
| 35 | Ile wody pije matka? | How much water does the mother drink? | wh_amount | drinking:liquid |
| 36 | Kiedy pies jadł? | When did the dog eat? | wh_time | eating:past |
| 37 | Kiedy nauczyciel czyta? | When does the teacher read? | wh_time | reading:present |
| 38 | Dlaczego dziecko płacze? | Why does the child cry? | wh_cause | crying_emotionally:reason |
| 39 | Jak pies biegnie? | How does the dog run? | wh_manner | running:manner |
| 40 | Czy pies ma książkę? | Does the dog have a book? | possession | have_unspecified |
| 41 | Czy matka ma kota? | Does the mother have a cat? | possession | have_unspecified |
| 42 | Co ma nauczyciel? | What does the teacher have? | wh_possession | have_unspecified:object |
| 43 | Kto ma klucz? | Who has the key? | wh_possession | have_unspecified:holder |
| 44 | Ile książek ma matka? | How many books does the mother have? | wh_possession_count | have_unspecified:object |
| 45 | Czy to jest ptak? | Is this a bird? | classification | bird:one:proximal_unresolved |
| 46 | Czy to jest książka? | Is that a book? | classification | book:one:distal_unresolved |
| 47 | Czy to są psy? | Are these dogs? | classification | dog:many:proximal_unresolved |
| 48 | Czy tamte psy są czarne? | Are those dogs black? | property | dog:many:distal |
| 49 | Czy to jest czarne? | Is this black? | property | entity:unspecified:proximal_unresolved |
| 50 | Czy tamto jest czarne? | Is that black? | property | entity:unspecified:distal |
| 51 | Czy te drzwi są otwarte? | Is this door open? | property_gap | open_state:door:one |
| 52 | Czy te okna są otwarte? | Are these windows open? | property_gap | open_state:window:many |
| 53 | Czy kot spał w domu? | Did the cat sleep in the house? | location_event_gap | sleeping:past:located_in |
| 54 | Czy nauczyciel nie czytał książki? | Did the teacher not read the book? | negative_event | reading:reader=teacher:one,material=book:one:past |
| 55 | Kto nie widzi kota? | Who does not see the cat? | negative_wh_role | seeing:experiencer |
| 56 | Kogo nie słyszy dziecko? | Whom does the child not hear? | negative_wh_role | hearing:stimulus:person_variable |
| 57 | Która książka jest czarna? | Which book is black? | wh_choice | black:bearer |
| 58 | Czy oba psy śpią? | Do both dogs sleep? | both_event | sleeping:sleeper=dog:many:present |

WH wymaga niewiadomej i jej wiązania, a nie tylko pominiętego argumentu. `this/that` i `to/ten/tamten` mają osobne tryby referencji; nie narzucamy bliskości polskiemu `ten`. `Is this black?` nie określa klasy ani liczby obiektu. `drzwi` jako plural_only nie rozstrzygają liczby realnych drzwi. Oba wejścia dla pary 30 jawnie opisują chodzenie; jej brakujący WH-binding miejsca docelowego pozostaje GAP_I. Braki pełnego I pozostają GAP_I, propozycje tekstów nie zastępują grafu.

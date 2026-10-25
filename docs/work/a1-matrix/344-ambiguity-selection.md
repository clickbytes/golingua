# #344 AMB43 — complete selected set

| Pair | Phenomenon | Polish source | English source | Authored semantic spec |
|---:|---|---|---|---|
| 01 | `determiners` | Nauczyciel widzi ptaka. | A teacher sees a bird. | `seeing:experiencer=teacher:one,stimulus=bird:one` |
| 02 | `determiners` | Lekarz słyszy psa. | The doctor hears a dog. | `hearing:experiencer=doctor:one,stimulus=dog:one` |
| 03 | `determiners` | Matka czyta książkę. | A mother reads the book. | `reading:reader=mother:one,material=book:one` |
| 04 | `determiners` | Ojciec widzi samochód. | The father sees a car. | `seeing:experiencer=father:one,stimulus=car:one` |
| 05 | `mass_reference` | Pies pije wodę. | The dog drinks water. | `drinking:drinker=dog:one,liquid=water:mass` |
| 06 | `mass_reference` | Kot je chleb. | A cat eats bread. | `eating:eater=cat:one,food=bread:mass` |
| 07 | `plural_reference` | Nauczyciele widzą ptaki. | Teachers see the birds. | `seeing:experiencer=teacher:many,stimulus=bird:many` |
| 08 | `plural_reference` | Lekarze słyszą koty. | The doctors hear cats. | `hearing:experiencer=doctor:many,stimulus=cat:many` |
| 09 | `read_homography` | Lekarze czytają książkę. | The doctors read the book. | `reading:reader=doctor:many,material=book:one` |
| 10 | `read_homography` | Matki czytają wiadomość. | The mothers read the message. | `reading:reader=mother:many,material=message:one` |
| 11 | `read_homography` | Ojcowie czytają książkę. | The fathers read the book. | `reading:reader=father:many,material=book:one` |
| 12 | `read_homography_child` | Dzieci czytają wiadomość. | The children read the message. | `reading:reader=child:many,material=message:one` |
| 13 | `read_homography` | Nauczycielki czytają list. | The teachers read the letter. | `reading:reader=teacher:many,material=letter_document:one` |
| 14 | `read_homography_context` | Czytamy wiadomość. | We read the message. | `reading:reader=person:many:speaker_group,material=message:one` |
| 15 | `child_polysemy` | Dziecko widzi kota. | The child sees the cat. | `seeing:experiencer=child:one,stimulus=cat:one` |
| 16 | `child_polysemy` | Pies widzi dziecko. | The dog sees the child. | `seeing:experiencer=dog:one,stimulus=child:one` |
| 17 | `child_polysemy` | Dzieci słyszą psa. | The children hear the dog. | `hearing:experiencer=child:many,stimulus=dog:one` |
| 18 | `child_polysemy` | Matka widzi dziecko. | The mother sees the child. | `seeing:experiencer=mother:one,stimulus=child:one` |
| 19 | `child_polysemy` | Nauczyciel daje dziecku jabłko. | The teacher gives the child an apple. | `giving:giver=teacher:one,recipient=child:one,theme=apple_fruit:one` |
| 20 | `child_polysemy` | Dziecko daje książkę matce. | The child gives the book to the mother. | `giving:giver=child:one,recipient=mother:one,theme=book:one` |
| 21 | `syncretic_roles` | Koty widzą psy. | The cats see the dogs. | `seeing:experiencer=cat:many,stimulus=dog:many` |
| 22 | `syncretic_roles` | Psy widzą koty. | The dogs see the cats. | `seeing:experiencer=dog:many,stimulus=cat:many` |
| 23 | `syncretic_roles` | Ptaki widzą koty. | The birds see the cats. | `seeing:experiencer=bird:many,stimulus=cat:many` |
| 24 | `syncretic_roles` | Koty widzą ptaki. | The cats see the birds. | `seeing:experiencer=cat:many,stimulus=bird:many` |
| 25 | `syncretic_roles` | Psy słyszą koty. | The dogs hear the cats. | `hearing:experiencer=dog:many,stimulus=cat:many` |
| 26 | `syncretic_roles` | Koty słyszą psy. | The cats hear the dogs. | `hearing:experiencer=cat:many,stimulus=dog:many` |
| 27 | `two_deictic_roles` | Ten pies widzi tamtego kota. | This dog sees that cat. | `seeing:experiencer=dog:one:proximal,stimulus=cat:one:distal` |
| 28 | `two_deictic_roles` | Tamten pies widzi tego kota. | That dog sees this cat. | `seeing:experiencer=dog:one:distal,stimulus=cat:one:proximal` |
| 29 | `two_deictic_roles` | Te psy widzą tamte koty. | These dogs see those cats. | `seeing:experiencer=dog:many:proximal,stimulus=cat:many:distal` |
| 30 | `two_deictic_roles` | Tamte psy widzą te koty. | Those dogs see these cats. | `seeing:experiencer=dog:many:distal,stimulus=cat:many:proximal` |
| 31 | `two_deictic_roles_child` | Ta matka słyszy tamto dziecko. | This mother hears that child. | `hearing:experiencer=mother:one:proximal,stimulus=child:one:distal` |
| 32 | `two_deictic_roles_child` | Tamto dziecko słyszy tę matkę. | That child hears this mother. | `hearing:experiencer=child:one:distal,stimulus=mother:one:proximal` |
| 33 | `quantifier_scope` | Każdy nauczyciel widzi psa. | Every teacher sees a dog. | `allteacher->somedog may vary perteacher` |
| 34 | `quantifier_scope` | Jakiś nauczyciel widzi każdego psa. | Some teacher sees every dog. | `someteacher->alldog sharedteacher` |
| 35 | `entity_count` | Nauczyciel widzi dwa koty. | The teacher sees two cats. | `seeing:experiencer=teacher:one,stimulus=cat:many:exactly2` |
| 36 | `entity_count` | Pies widzi dwóch nauczycieli. | The dog sees two teachers. | `seeing:experiencer=dog:one,stimulus=teacher:many:exactly2` |
| 37 | `absence_possession_ambiguity` | Nie ma kotów. | There are no cats. | `PLimpersonalabsenceorprodropthirdpersonnegativeHAVE;ENabsenceonly` |
| 38 | `PP_attachment` | Nauczyciel widzi matkę z dzieckiem. | The teacher sees the mother with the child. | `accompanimentattachesmotherorobserver;noinventedownership` |
| 39 | `PP_attachment` | Pies widzi kota na stole. | The dog sees the cat on the table. | `catlocationorobservationlocation;noaddedposture` |
| 40 | `PP_attachment` | Matka widzi nauczyciela z książką. | The mother sees the teacher with a book. | `associationofbooktoteacherorobserver;noownershiptitle` |
| 41 | `collective_numeral_child` | Dwoje dzieci widzi psa. | Two children see the dog. | `seeing:experiencer=child:many:exactly2,stimulus=dog:one` |
| 42 | `same_class_mentions` | Nauczyciel widzi nauczyciela. | The teacher sees the teacher. | `seeing:experiencer=teacher:one,stimulus=teacher:one;distinctmentionsnotassertdistinctidentities` |
| 43 | `kinship_binding` | Dziecko matki widzi psa. | The mother's child sees the dog. | `offspringsenseonly;motherrelationnotageassumption` |

All rows share the selected semantic caution: require every licensed I reading, avoid inferred identity, preserve role-bound anchors, and keep entity quantification distinct from event count. Exact per-row details for the number of interpretations are encoded in the I catalogue and cases. Full gaps: 33, 34, 37–40, 43. Row 37 PL requires both absence and negative-HAVE readings; its English absence proposal is illustrative and is not a complete oracle for that PL source.

## Coordinator-supplied English target vectors

For Polish syncretic rows 21–26, both the primary role assignment and `.swap` use the exact English SVO candidate vectors below; each surface carries independent experiencer/stimulus determiner assumptions. For Polish deictic rows 27–32, the source-anchor ambiguity is expressed by the listed English reference alternatives; only the uncertain role receives a reference assumption. Rows 31–32 have explicit offspring-sense vectors.

### Row 21: primary and `.swap`

**base:**
- `Cats see dogs.` — experiencer `bare`, stimulus `bare`
- `Cats see the dogs.` — experiencer `bare`, stimulus `definite`
- `The cats see dogs.` — experiencer `definite`, stimulus `bare`
- `The cats see the dogs.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Dogs see cats.` — experiencer `bare`, stimulus `bare`
- `Dogs see the cats.` — experiencer `bare`, stimulus `definite`
- `The dogs see cats.` — experiencer `definite`, stimulus `bare`
- `The dogs see the cats.` — experiencer `definite`, stimulus `definite`

### Row 22: primary and `.swap`

**base:**
- `Dogs see cats.` — experiencer `bare`, stimulus `bare`
- `Dogs see the cats.` — experiencer `bare`, stimulus `definite`
- `The dogs see cats.` — experiencer `definite`, stimulus `bare`
- `The dogs see the cats.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Cats see dogs.` — experiencer `bare`, stimulus `bare`
- `Cats see the dogs.` — experiencer `bare`, stimulus `definite`
- `The cats see dogs.` — experiencer `definite`, stimulus `bare`
- `The cats see the dogs.` — experiencer `definite`, stimulus `definite`

### Row 23: primary and `.swap`

**base:**
- `Birds see cats.` — experiencer `bare`, stimulus `bare`
- `Birds see the cats.` — experiencer `bare`, stimulus `definite`
- `The birds see cats.` — experiencer `definite`, stimulus `bare`
- `The birds see the cats.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Cats see birds.` — experiencer `bare`, stimulus `bare`
- `Cats see the birds.` — experiencer `bare`, stimulus `definite`
- `The cats see birds.` — experiencer `definite`, stimulus `bare`
- `The cats see the birds.` — experiencer `definite`, stimulus `definite`

### Row 24: primary and `.swap`

**base:**
- `Cats see birds.` — experiencer `bare`, stimulus `bare`
- `Cats see the birds.` — experiencer `bare`, stimulus `definite`
- `The cats see birds.` — experiencer `definite`, stimulus `bare`
- `The cats see the birds.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Birds see cats.` — experiencer `bare`, stimulus `bare`
- `Birds see the cats.` — experiencer `bare`, stimulus `definite`
- `The birds see cats.` — experiencer `definite`, stimulus `bare`
- `The birds see the cats.` — experiencer `definite`, stimulus `definite`

### Row 25: primary and `.swap`

**base:**
- `Dogs hear cats.` — experiencer `bare`, stimulus `bare`
- `Dogs hear the cats.` — experiencer `bare`, stimulus `definite`
- `The dogs hear cats.` — experiencer `definite`, stimulus `bare`
- `The dogs hear the cats.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Cats hear dogs.` — experiencer `bare`, stimulus `bare`
- `Cats hear the dogs.` — experiencer `bare`, stimulus `definite`
- `The cats hear dogs.` — experiencer `definite`, stimulus `bare`
- `The cats hear the dogs.` — experiencer `definite`, stimulus `definite`

### Row 26: primary and `.swap`

**base:**
- `Cats hear dogs.` — experiencer `bare`, stimulus `bare`
- `Cats hear the dogs.` — experiencer `bare`, stimulus `definite`
- `The cats hear dogs.` — experiencer `definite`, stimulus `bare`
- `The cats hear the dogs.` — experiencer `definite`, stimulus `definite`

**.swap:**
- `Dogs hear cats.` — experiencer `bare`, stimulus `bare`
- `Dogs hear the cats.` — experiencer `bare`, stimulus `definite`
- `The dogs hear cats.` — experiencer `definite`, stimulus `bare`
- `The dogs hear the cats.` — experiencer `definite`, stimulus `definite`

### Row 27: reference choices
- `This dog sees that cat.` — role `experiencer`, reference `this`
- `That dog sees that cat.` — role `experiencer`, reference `that`

### Row 28: reference choices
- `That dog sees this cat.` — role `stimulus`, reference `this`
- `That dog sees that cat.` — role `stimulus`, reference `that`

### Row 29: reference choices
- `These dogs see those cats.` — role `experiencer`, reference `this`
- `Those dogs see those cats.` — role `experiencer`, reference `that`

### Row 30: reference choices
- `Those dogs see these cats.` — role `stimulus`, reference `this`
- `Those dogs see those cats.` — role `stimulus`, reference `that`

### Row 31: reference choices
- `This mother hears that child.` — role `experiencer`, reference `this`
- `That mother hears that child.` — role `experiencer`, reference `that`
**Offspring meaning:**
- `This mother hears that offspring.` — role `experiencer`, reference `this`
- `That mother hears that offspring.` — role `experiencer`, reference `that`

### Row 32: reference choices
- `That child hears this mother.` — role `stimulus`, reference `this`
- `That child hears that mother.` — role `stimulus`, reference `that`
**Offspring meaning:**
- `That offspring hears this mother.` — role `stimulus`, reference `this`
- `That offspring hears that mother.` — role `stimulus`, reference `that`

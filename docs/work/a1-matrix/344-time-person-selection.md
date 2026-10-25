# #344 M8-TIME-PERSON — selection and phenomena

Selected 16 of 56 A1.10 time pairs for this serial subcomponent. Source and candidate text below is copied from the coordinator-curated selection file; semantic requirements remain distinct from observed implementation behavior. Owner acceptance: `NOT_EVALUATED`.

| Pair | English source | Polish source | Predicate / roles | Frame | PL present candidates | PL past candidates |
|---|---|---|---|---|---|---|
| a1.10-time-001 | I see the dog. | Widzę psa. | `seeing`, `experiencer=person:one:speaker,stimulus=dog:one` | `present` | Widzę psa. | — |
| a1.10-time-002 | You see the cat. | Widzisz kota. | `seeing`, `experiencer=person:one:addressee,stimulus=cat:one` | `present` | Widzisz kota.<br>Widzicie kota. | — |
| a1.10-time-003 | We see the bird. | Widzimy ptaka. | `seeing`, `experiencer=person:many:speaker_group,stimulus=bird:one` | `present` | Widzimy ptaka. | — |
| a1.10-time-004 | You see the book. | Widzicie książkę. | `seeing`, `experiencer=person:many:addressee_group,stimulus=book:one` | `present` | Widzisz książkę.<br>Widzicie książkę. | — |
| a1.10-time-005 | I hear the dog. | Słyszę psa. | `hearing`, `experiencer=person:one:speaker,stimulus=dog:one` | `present` | Słyszę psa. | — |
| a1.10-time-006 | You hear the bird. | Słyszysz ptaka. | `hearing`, `experiencer=person:one:addressee,stimulus=bird:one` | `present` | Słyszysz ptaka.<br>Słyszycie ptaka. | — |
| a1.10-time-007 | We hear the cat. | Słyszymy kota. | `hearing`, `experiencer=person:many:speaker_group,stimulus=cat:one` | `present` | Słyszymy kota. | — |
| a1.10-time-008 | You hear the child. | Słyszycie dziecko. | `hearing`, `experiencer=person:many:addressee_group,stimulus=child:one` | `present` | Słyszysz dziecko.<br>Słyszycie dziecko. | — |
| a1.10-time-009 | I read the book. | Czytam książkę. | `reading`, `reader=person:one:speaker,material=book:one` | `present` | Czytam książkę. | Czytałem książkę.<br>Czytałam książkę. |
| a1.10-time-010 | You read the message. | Czytasz wiadomość. | `reading`, `reader=person:one:addressee,material=message:one` | `present` | Czytasz wiadomość.<br>Czytacie wiadomość. | Czytałeś wiadomość.<br>Czytałaś wiadomość.<br>Czytaliście wiadomość.<br>Czytałyście wiadomość. |
| a1.10-time-011 | We read the letter. | Czytamy list. | `reading`, `reader=person:many:speaker_group,material=letter_document:one` | `present` | Czytamy list. | Czytaliśmy list.<br>Czytałyśmy list. |
| a1.10-time-012 | You read the book. | Czytacie książkę. | `reading`, `reader=person:many:addressee_group,material=book:one` | `present` | Czytasz książkę.<br>Czytacie książkę. | Czytałeś książkę.<br>Czytałaś książkę.<br>Czytaliście książkę.<br>Czytałyście książkę. |
| a1.10-time-013 | I drink water. | Piję wodę. | `drinking`, `drinker=person:one:speaker,liquid=water:mass` | `present` | Piję wodę. | — |
| a1.10-time-014 | You drink milk. | Pijesz mleko. | `drinking`, `drinker=person:one:addressee,liquid=milk_food:mass` | `present` | Pijesz mleko.<br>Pijecie mleko. | — |
| a1.10-time-015 | We drink coffee. | Pijemy kawę. | `drinking`, `drinker=person:many:speaker_group,liquid=coffee_beverage:mass` | `present` | Pijemy kawę. | — |
| a1.10-time-016 | You drink tea. | Pijecie herbatę. | `drinking`, `drinker=person:many:addressee_group,liquid=tea_beverage:mass` | `present` | Pijesz herbatę.<br>Pijecie herbatę. | — |

## Phenomena preserved

- Source fixture IDs are `a1.10.time.NNN.en|pl`. English bare *read* rows 9–12 require present and past I variants; Polish finite sources keep their authored frame.
- Child pair 8 keeps the age-child base and `.offspring` alternative for both languages. No `.age` alias is used.
- English *you* keeps unknown participant count and an addressee anchor. Polish targets list singular/plural forms with local `number` assumptions.
- Past Polish personal forms carry local `gender` assumptions (masculine/feminine for singular, masculine_personal/non_masculine_personal for plural); they do not alter I or infer biological sex.
- Count nouns retain singular/plural and determiner choices; mass nouns retain bare/definite choices. Source PL to target EN uses personal pronouns from participant context, not a guessed third-person noun.
- The remaining 40 authored rows stay queued within #344.

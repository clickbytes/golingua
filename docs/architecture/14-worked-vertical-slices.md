# 14 — Przekrojowe przebiegi i falsyfikacja architektury

**Status:** PROPOSED EXAMPLES, nie gotowe golds. Żadne poniższe zdanie nie staje się automatycznie zatwierdzonym oraclem testowym. W G02/G03/G05/G06 każdy przykład musi otrzymać niezależną recenzję, konkretny fixture ID i powiązany pack SHA.

## A. Zdanie przechodnie — pl → CORE → en

**Input (illustrative):** `Dziewczyna widzi psa.` (PL).

1. `Tokenizer` zachowuje source spans: `Dziewczyna`, `widzi`, `psa`.
2. `Morphology.Analyze`: `widzi` może wskazać `pl.widziec.1`, features `present/third/singular`; `psa` to prawidłowa forma określonego przypadku i liczby dla `pl.pies.1`; homografy, jeśli istnieją, pozostają alternatywami.
3. `LexicalSense` `pl.widziec.1.sense.1` → approved `event.visual_perception`; `pl.dziewczyna.1` → `person.girl`; `pl.pies.1` → `animal.dog`. To source-like IDs, a dokładny Go source binding wymaga review.
4. `ValencyFrame` przypisuje `agent→subject` z mianownikiem i `theme→object` z biernikiem dla danej konstrukcji; nie jest to globalny mapping wszystkich zdarzeń.
5. `ConstructionFamily(pl.transitive.active) + Variant(svo)` wiąże kolejność; warunek liczby/osoby steruje formą `widzi`. Prosta kolejność bez morfologii nie jest wystarczającym dowodem.
6. `MeaningGraph`: `event.visual_perception(agent: girl_1, theme: dog_2, tense:present, polarity:positive)`. Identity referent IDs pozostają lokalne dla tego dokumentu.
7. EN approved sense `en.see.1.sense.1`; `en.transitive.active.svo` i EN NP/article/VP policies generują **przykładowo** `The girl sees the dog.`, o ile definiteness/target determiner są licencjonowane i nie wprowadzają nieuzasadnionej semantycznej determinacji.
8. Target re-analysis must include the same semantic roles, tense, polarity and referent identities under declared translation context.

**Falsification:** `Psa widzi dziewczyna` może być legalnym PL OVS z tym samym MeaningGraph; `The dog sees the girl` ma inne role i nie jest automatycznym wariantem. Brak dowodu przypadków przy PL OVS → ambiguity/incomplete, nie guessed unique.

## B. Morfologia i nielegalny fallback

`en.go.1`, `past`, zaakceptowany principal part `went`. Generator dopuszcza `went`, lecz niedopuszczalne jest emitowanie `goed` jako „druga regularna alternatywa” bez jawnie zatwierdzonego wariantu. `Analyze("went")` zachowuje poprawne `LexemeID + past`.

Polski przykład: `pl.pies.1` nominative `pies` vs accusative/genitive `psa` wymaga odpowiedniego stem/allomorph. Nie dołączaj automatycznie `a` do `pies`, jeśli daje błędną powierzchnię.

Negatives: `block` jest zamieniony w `missing`; `replace` zostawia regularny wynik; `addVariant` usuwa regularny wynik; formy dwóch komórek synkretycznych sprowadzane do jednego reading. Każdy przypadek musi FAIL.

## C. Wieloznaczne słowo i MWE

English `read` może mieć taki sam surface dla past i base, więc analiza formy nie powinna od razu wybierać czasu. Ostateczny wybór wymaga składni/context/hard evidence. Jeżeli parser kończy z powodu limitu przy kilku możliwych gałęziach → incomplete nawet wtedy, kiedy pozostał jeden kandydat.

`break the ice` w interpretacji społecznej powinno prowadzić do leksykalnego sense konceptu `event.social_icebreak`, a w literalnej do `event.break_object` z fizycznym dopełnieniem. Polisemia nie uprawnia do twierdzenia, że ktoś fizycznie złamał lód, gdy chodziło o przełamanie pierwszych lodów.

Negative: Go wybiera sense o najwyższym „typical weight” i ukrywa drugi legalny → FAIL.

## D. Liczebniki i czas jako generator frazy, nie tabela zdań

`quantity: 5` w CORE pozostaje wartością semantyczną; plugin PL w określonej konstrukcji dobiera formę NP `pięć psów` i wymagane cechy składniowe, a EN może realizować `five dogs`. Sama znajomość `animal.dog` i `5` nie dowodzi całego zdania, a reguła zgodności czasownika jest konstrukcyjna i profile-specific.

`timeReference:tomorrow` nie zmienia `event.tense` na przyszły bez oddzielnego, zatwierdzonego wniosku. `calendarDate`/`clockTime` mają jawne zakresy; pominięcie ich w MeaningGraph comparison jest regresją semantyczną.

Negatives: system zamienia `5` na `singular` w CORE, usuwa `clockTime` podczas transferu lub generuje nieistniejącą liczbę bez licencjonowanej reguły → FAIL.

## E. Controlled answer, jawne fakty, zero pamięci

`FactSet(request)`:
- `dog_1 instanceOf animal.dog`;
- `dog_1 located_in room_3`.

`QuestionMeaning` „Gdzie jest ten pies?” z jawnie przypiętym referentem `dog_1`.

`Reasoner` rozpoznaje tylko zatwierdzony typ relacji i argumentów, zwraca `room_3` oraz `ProofStep` wskazujący input FactID. `ContentPlanner` wybiera typ odpowiedzi i wynikowy MeaningGraph (bez języka), target PL/EN plugin wybiera właściwą konstrukcję odpowiedzi.

**Nie wolno:** uznać `dog_2` za ten sam obiekt; odpowiedzieć `house_4` na podstawie niezatwierdzonej przechodniości `located_in`; zapisać `dog_1` do trwałej pamięci; zamienić brak faktu na „nie”; zwrócić pewny answer tylko z graph definition typów.

## F. Kontrakt jawnej niekompletności

Wejście zawiera 2 licencjonowane odczytania: A zrekonstruowany, B jeszcze niewyeksplorowany z powodu budżetu. Wynik z jedną pozycją w `alternatives` **nie może być `status:ok,complete:true`**. Zamiast tego `incomplete` lub `resource_limit + completeness:incomplete` z diagnostyką. „Pierwsza legalna realizacja” jest preview, nie dowodem semantycznej unikalności.

## Jak zamieniać te przykłady w testy

Dla każdej linii: niezależny recenzent → approved expected MeaningGraph/Surface/ResultClass → pozytywny fixture + semantyczny negatywny kontrprzykład → source and pack hash → Go CI/PR review. Nie kopiuj illustrative wire JSON jako gotowego kontraktu bez walidacji wersji.

# 08 — MeaningGraph, semantyka, instancje i mały reasoner

**Status:** PROPOSED. Issues: [G01 #2](https://github.com/clickbytes/golingua/issues/2), [G05-04 #47](https://github.com/clickbytes/golingua/issues/47), [G06 #7](https://github.com/clickbytes/golingua/issues/7). Source: `js/core/meaning-graph.js`, `ontology.js`, `typed-graph-match.js`, `semantic-composition.js`, `response-planner.js` oraz źródłowe ADRs.

## 1. Statyczna wiedza językowa vs znaczenie wypowiedzi vs fakty świata

- `ConceptDefinition(event.read)` definiuje znaczenie typu i wymagane role oraz role type signatures.
- `MeaningGraph(Event{concept:event.read,agent:girl_1,theme:book_2})` jest interpretacją pojedynczej konkretnej wypowiedzi; sam fakt obecności graph nie jest jeszcze twierdzeniem.
- `DialogueAct(Assert,MeaningGraph)` wyraża intencję asercji, którą można ewentualnie przyjąć jako fakt, jeżeli trust/provenance dopuszcza.
- `WorldFact` to źródłowo przypisane twierdzenie o referentach, czasie, zakresie i polaryzacji; w Go v0 podawane jawnie per request.
- `Proof` pokazuje dozwolone wnioskowanie z WorldFact i reguł, nie „podobieństwo” do konceptu.

Nie dodajemy `dog_1` jako nowego ConceptDefinition. Węzeł `animal.dog` to *typ*, referent to *instancja*. Dwa psy = dwa ReferentID, nawet jeśli mają wspólne cechy i lemma.

## 2. Algebra dokumentów znaczenia

Potrzebny bounded AST/tagged union: `Event, State, Participant, Referent, Question, Negation/Scope, Modality, Coordination, Conditional, Temporal, Quantifier, DialogueAct, Discourse`, ale tylko subset faktycznie potwierdzony Go fixtures jest IMPLEMENTED. Żaden rodzaj nie może być reprezentowany „pustym stringiem fallback”.

Przykład **Proposed, nie obecny wire format**:

```json
{
  "kind":"event",
  "concept":"event.visual_perception",
  "roles":{
    "agent":{"referent":"girl_1","concept":"person.girl"},
    "theme":{"referent":"dog_2","concept":"animal.dog"}
  },
  "features":{"tense":"present","polarity":"positive"}
}
```

Konkretne enumy, formaty JSON i wymagane pola muszą zostać potwierdzone w G01-04/G00-03; nie kopiuj z przykładów schematu bez walidatora. Na poziomie semantycznym **roles i referents muszą być zachowane** przy reordering i roundtrip.

## 3. Zakres operatora ma znaczenie logiczne

`Negation(Event)` to inny znaczeniowo graf niż `Event(polarity=positive)`. `Question(Event)` **nie** wprowadza dodatniej asercji. W `Conditional(if A, then B)` nie wolno utworzyć pewnej asercji ani dla A, ani dla B bez dodatkowej reguły. `Modality(possibility, A)` nie jest dowodem rzeczywistego A.

Zasięg i precedence operatorów muszą być jawnie reprezentowane, np. `Not(Possible(A))` ≠ `Possible(Not(A))`. Nie zakładać, że dwie permutacje operatorów są równoważne. W przypadku braku licencjonowanej konstrukcji do ich realizacji w EN/PL → `unsupported` lub ślad informationLoss, nie pomijanie znaku/argumentu.

Temporalia: odróżniaj `event.tense` od `timeReference(today/tomorrow)`, `clockTime` i `calendarDate` oraz jawnych kotwic czasu stwierdzenia. Nie wybieraj tense wyłącznie po „jutro” i nie pomijaj daty przy porównywaniu MeaningGraph.

Kwantyfikacja: liczba `exact N` jest znaczeniem, a forma NP, przypadek i zgoda są kwestią języka. Nie pozwalaj na niejawne `N=singular` w CORE, jeśli wynika to tylko z polskiej reguły zgodności.

## 4. Tożsamość wzmianki i niepewność

`MentionID` to pozycja/odniesienie w zdaniu; `ReferentID` oznacza tożsamość instancji w zdefiniowanym scope. Ponowne wystąpienie `pies` nie identyfikuje tego samego psa. Deiktyczne `speaker/addressee` wymagają uprawomocnionego kontekstu; anafora może pozostać unresolved. Brak resolved reference ≠ losowo wybrany najbliższy participant.

Source `referent-vs-mention.md` dopuszcza lokalne ID w dyskursie i jawne koreferencje. Go v0 nie utrwala tożsamości między niezależnymi requestami bez jawnego inputu.

## 5. Ontologia i matching

`parent/is_a` twardy i acykliczny; `definitionGraph` zawiera typed `required` lub `typical`. Typical w źródłowym ADR ma wagę salience, **nie prawdopodobieństwo prawdziwości**. Nie wykluczaj zgodnego sensu na podstawie braku typowej relacji; soft ranking nigdy nie zamienia kilku licencjonowanych odczytań w dowiedzione unique.

Wynik matching musi móc zwrócić:
- hard-compatible candidates;
- hard-violations z source path;
- unresolved wymagany edge przy braku obserwacji;
- `complete: false` po zakończeniu budżetu;
- alternatives z odrębnymi `MeaningGraph`.

Wszystkie przypadki muszą być typowane; nie dopasowywać tylko nazw etykiet i powierzchniowych stringów.

## 6. Request-scoped FactSet, proof i truth values

Fakty G06: `(FactID,subjectReferent,relation,objectReferent,polarity,validTime,source/assertionScope)` z walidacją subject/object types i trust. Brak danego fact oznacza `unknown`; odpowiedź false na yes/no wymaga jawnego przeciwnego faktu lub dozwolonej reguły negacji, nigdy open-world absence. Sprzeczne poparte facts dają `conflict/ambiguous`.

Proposed: `ProofStep{RuleID,PremiseFactIDs,ConclusionFactID/Proposition,Context,SourceRefs}`. Reguły jawne, monotoniczne/typowane w początkowym subset, z zamkniętym DSL i ograniczeniami rozgałęzień. Nie aktywować ogólnego wnioskowania `located_in` transitive, dopóki dokładnie ta relacja, jej czas i scope nie są do tego zatwierdzone.

**Pytanie przykładowe:**
- FactSet: `dog_1 located_in room_3` i `room_3 located_in house_4`.
- Query: `Is dog_1 in house_4?`
- Dopóki nie ma jawnie zatwierdzonej reguły właściwego zasięgu `located_in`, odpowiedź **unknown**, nie automatyczne true. Po dodaniu licencjonowanej reguły, wszystkie kroki inference mają zawierać ProofStep.

## 7. Planner treści, nie tylko generowanie tekstu

`QuestionMeaning + QueryResult + evidence → ContentPlan{Goal,DialogueAct,Proposition,SelectedFacts,ProofRefs,RequiredConstructionCapability}`. Cele v0: yes/no answer, ograniczone WH answer, clarify/refuse/unknown. PL/EN realization jest osobnym krokiem licencjonowanych konstrukcji.

Kontrole negatywne:
- Pytanie użytkownika samo w sobie wytwarza fakt → FAIL.
- Dostępny tylko `typical` z ontology i odpowiedź pewna „tak” → FAIL.
- Dwa różne `dog` referenty utożsamione z jednym → FAIL.
- Wiele konfliktowych faktów i planner wybiera pierwszy po kolejności JSON → FAIL.
- Niewspierany WH gap lub brak dowodu skutkuje pewnym wygenerowanym zdaniem → FAIL.
- Zmiana stylu/personality zmienia role/polarity/referent/date → FAIL.

## 8. Bez pamięci i uczenia

Go v0 nie ma session/episodic memory ani online learning. Source JS **ma ograniczoną SessionMemory i planner**, więc sformułowanie „flexLingua nie ma w ogóle pamięci” jest nieprecyzyjne. W nowym Go v0 te możliwości są świadomie **wyłączone**; G09 opisuje przyszłe rozszerzenia bez aktywacji i bez uczenia reguł w locie.

# Planowanie treści odpowiedzi

`internal/planner` wybiera cel dialogowy oraz buduje językowo niezależny `Meaning` dla odpowiedzi. Wejściem jest typowane pytanie, wynik z `internal/reason`, jego `Meaning` i przypięty snapshot CORE. Dane faktów są potrzebne, gdy odpowiedź wskazuje referent lub przytacza dowody.

Planner nie realizuje tekstu, nie dobiera leksemów i nie ocenia pewności liczbowo. Wynik zachowuje identyfikatory wybranych faktów, źródła i dowodów reguł. Nie tworzy nowych twierdzeń ani referentów, których nie ma w wejściowym kontekście.

## Wybór celu

- Ugruntowane pytanie tak/nie daje odpowiedź twierdzącą albo zaprzeczoną, zgodnie z polaryzacją wyniku.
- Pytanie WH z jednym ugruntowanym kandydatem daje odpowiedź wskazującą ten referent.
- Brakujący referent albo wybór między wieloma kandydatami prowadzi do doprecyzowania tylko wtedy, gdy wejście zawiera typowaną lukę do wyjaśnienia.
- Sprzeczne dowody prowadzą do aktu wyjaśnienia, który zachowuje obie polaryzacje.
- Wynik nieobsługiwany, nieznany bez jawnej luki albo bezpiecznie niemożliwy do zrealizowania prowadzi do odmowy.

Luka doprecyzowania jest sprawdzana względem roli predykatu oraz jawnego kontekstu wzmianki/referentów. Planner nie zgaduje, o co należy dopytać. Nierozpoznane wyniki, niespójny snapshot CORE, brakujące dowody i przekroczenia limitów kończą się typowanym błędem.

## Granice i odbiór

Cele odpowiedzi są `answer_yes_no`, `answer_wh`, `clarify`, `refuse` i `explain`. Akty odmowy i wyjaśnienia nie są asercjami; projektowany `Meaning` nie trafia przez to do projekcji asertywnych faktów.

Typowany CLI `question` udostępnia plan odpowiedzi oraz opcjonalną realizację PL/EN na aktywnym snapshotcie. Obecna realizacja tekstowa jest ograniczona do dodatnich odpowiedzi yes/no na zatwierdzonych trasach zdań; odpowiedź `incomplete` jawnie pokazuje założenia. False, WH, clarify, refusal i explain pozostają bez tekstu tam, gdzie brakuje zatwierdzonej konstrukcji. Ocena delegowana przez agenta i akceptacja językowa właściciela są osobnymi statusami; akceptacja właściciela pozostaje **NOT RUN**.

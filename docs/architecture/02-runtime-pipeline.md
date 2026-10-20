# 02 — Runtime i pełne ścieżki interpretacji/realizacji

**Status:** PROPOSED. Ta specyfikacja określa kontrakty przepływu, nie gotowe publiczne funkcje w Go. Issues: [G00-03](https://github.com/clickbytes/golingua/issues/13), [G03](https://github.com/clickbytes/golingua/issues/4), [G05](https://github.com/clickbytes/golingua/issues/6), [G06](https://github.com/clickbytes/golingua/issues/7).

## Ścieżka A — compile on data change, nie przy każdym tłumaczeniu

`source JSON/TSV → bounded parse → schema & type validation → staged candidates → license/provenance/review → approved snapshot → Concept/Language/Construction/Procedure compilers → immutable indexed pack + checksum + capability manifest`.

Walidacja **wszystkich** kandydatów w transakcji; naruszenie nie tworzy częściowo aktywnego pakietu. Kody procedur/rozwiązań, krawędzie grafu, rodziny konstrukcji i indeksy odwróconych form muszą być gotowe przed uruchomieniem żądania. Ten sam zatwierdzony input + wersja kompilatora daje ten sam logiczny pack digest.

## Ścieżka B — analyze: tekst → MeaningGraph

1. `validate input`: język, długość, UTF-8, schema i zasoby; zachowaj oryginał (source byte offsets).
2. `segment/tokenize`: możliwe segmentacje i MWE, bez wymuszania jednej przez zachłanny wybór.
3. `morphology.Analyze`: wszystkie licencjonowane odczytania form (leksem, POS, cechy, wyjątek/rule trace).
4. `lexical senses`: tylko zatwierdzone sensy mogą licencjonować semantic transfer; wolno zachować unknown-unmapped morph analysis w diagnostyce.
5. `construction parser`: dopasuj plan konstrukcji (sloty, wariant, przypadek, zgodność, pomocnicze i zakres) do lattice; współdziel poddrzewa i stosuj budżety.
6. `valency + semantic bindings`: sprawdź zgodność typu ról i ograniczeń selekcyjnych. `typical` może uszeregować, ale nie usuwa kandydata.
7. `MeaningGraph validate`: rodzaje event/state/question, operator scopes, referent IDs, polarity, temporal, quantity.
8. `result`: complete unique / complete ambiguous / incomplete / unsupported / invalid / resource_limit. **Jedna znana analiza plus przerwana druga gałąź NIE JEST UNIQUE**.

Wynik ma przenosić źródłowe `sourceSpans`, `SenseID`, `ValencyFrameID`, `ConstructionPlanID` i diagnostykę odrzuconych hard-violations. Nie wybieraj pierwszego kandydata w kolejności JSON jako „jedyny”.

## Ścieżka C — realize: MeaningGraph → tekst

1. Waliduj MeaningGraph i jawne żądane cechy prezentacji.
2. Szukaj *approved target* `SemanticBinding` i odpowiedniego `ValencyFrame`, sprawdź konstrukcje i pokrycie operacji pytania/negacji/strony.
3. `SelectConstructionPlan` tylko spośród licencjonowanych rodzin i wariantów; przeniesienie semantycznych ról na argumenty składniowe odbywa się tutaj.
4. Uzgodnij cechy, wybierz leksem i komórkę paradygmatu, wywołaj generator, zastosuj `block/replace/addVariant` zgodnie z kontraktem.
5. Linearizuj z zachowaniem obowiązkowego `@verb`, wszystkich ról, znaków interpunkcyjnych i source/target segment spans; nie dopowiadaj czasu/referenta/ilości.
6. Jeśli target nie oddaje cechy, wymagaj jawnego `informationLoss(path,reason,status)` lub refusal — nigdy nie pomijaj po cichu.
7. Tam, gdzie parser i realizer mają to samo zatwierdzone capability: powtórna analiza surface powinna zachować oryginalne znaczenie (nie musi być jedynym odczytaniem).

Ranking wariantów jest dopuszczalny **wyłącznie po filtrach semantycznej legalności**. Nie traktować preferencji stylistycznej jako dowodu znaczenia.

## Ścieżka D — rozumowanie/odpowiedź, bez pamięci

`Validated QuestionMeaning + request FactSet → type+reference check → bounded proof-based entailment → QueryResult (true/false/unknown/conflict/unsupported) → ContentPlanner (answer/clarify/refuse) → ResponseMeaningGraph → target licensed ConstructionPlan → response + proof refs`.

Fakty są referent- i time-scoped; brak danych nie pozwala uznać pytania za fałszywe. W G06 mówimy o ograniczonych jawnych regułach, nie ogólnej wiedzy o świecie. Nie korzystamy z historii czatu, jeśli request jawnie jej nie podał jako zatwierdzonych faktów.

## Kontrakt wyników — dwie niezależne osie

**Status próby:** `ok`, `ambiguous`, `incomplete`, `unsupported`, `invalid`, `resource_limit`, `infrastructure_error`. **Kompletność poszukiwania:** `complete` lub `incomplete`. Wartość semantyczna dla query: `true`, `false`, `unknown`, `conflict`. Te osie **nie są tym samym**. `absent` to brak pola, `unknown` to jawny brak wiedzy; `unsupported` oznacza poprawną, lecz niewspieraną konstrukcję, a `invalid` narusza kontrakt.

Przykładowa propozycja wire envelope (pola implementacyjne do zatwierdzenia w G00-03):

```json
{
  "version": "proposed-0",
  "status": "incomplete",
  "searchCompleteness": "incomplete",
  "meaning": null,
  "alternatives": [],
  "diagnostics": [{"code":"SOURCE_ANALYSIS_INCOMPLETE","path":"#/analysis","severity":"warning"}],
  "evidence": [],
  "informationLoss": []
}
```

Nie traktować tego jako już obsługiwanego JSON. Krytyczna zasada: incomplete search nie dowodzi, że lista `alternatives` jest pełna. `resource_limit` może być przyczyną incomplete, ale nie jest poprawnym tłumaczeniem.

## Budżety i deterministyka

Każda kosztowna operacja przyjmuje zwalidowany `Budget`: input length, token count, number of candidates, depth, nodes/edges, morphology cells, procedure steps, derivations and trace length. Twarde limity są jawne, wymuszane przed wzrostem alokacji; konkretne liczby ustali dany task po benchmarku (G00-03 / G01–G08), nie narzucamy obecnych JS pułapów przez przypadek. Przekroczenie → odmowa/niekompletność z powodem, nie najlepsza zgadywana wartość.

Wolno cache'ować tylko pure rezultaty oznaczone `packHash` i wszystkimi kontekstowymi parametrami, jeśli są istotne; brak cache wyników pomiędzy niezależnymi referentami/requestami. Snapshot pack pozostaje read-only i wielokrotnie współdzielony, per-request state jest odrębne.

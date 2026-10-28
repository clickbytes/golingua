# DSE — wyniki i zaobserwowane możliwości

**Stan:** techniczny obraz ograniczonego profilu na evidence #09-01 i #09-02. Odbiór #400 pozostaje `IN_PROGRESS`; akceptacja językowa właściciela: **NOT EVALUATED**.

## Co potwierdzają artefakty

- Pakiety metadanych są jawnie wybierane przez CLI. Dostępne polecenia obejmują sprawdzanie IR, budowanie i syntezę konstrukcji, ograniczone wyszukiwanie, niezależne `verify-goal`, walidację faktów, ewaluację, porównanie i realizację na wskazanym pakiecie językowym. Interfejs odczytano z `golingua --help` binarki użytej w #09-02; wywołania i pliki stdout zapisano w [manifeście komend evidence #09-02](evidence/09-02/commands.jsonl).
- Neutralny przykład #09-02 buduje anonimową funkcję ścieżki dwukrokowej z restrykcją unary dla jawnego sortu `route_node`. Konstruktor zwraca `ok` i `fully_validated=true`; niezależne `verify-goal` zwraca `ok`, `complete=true`, `goal_satisfied=true` i sprawdzone ograniczenia. Certyfikat potwierdza konstrukcję strukturalną, nie prawdziwość w świecie.
- Ewaluacja tego wyrażenia nad wskazanym, jawnym zestawem faktów zwraca `ok`, `knowledge=supported` i dwie odpowiedzi: `(route-hub-b, route-terminal)` oraz `(route-origin, route-terminal)`. Wynik dotyczy tylko przekazanych faktów i skończonej dziedziny referentów. Brak faktu oznacza `unknown`, nie `false`.
- Niezależna realizacja checked IR jest `ok` w obu językach. EN: “The relation from x to y holds exactly when there is an entity z such that x sees z and z sees y.” PL: “Relacja od x do y zachodzi wtedy i tylko wtedy, gdy istnieje taki element z, że x widzi z i z widzi y.” Coverage potwierdza kształt funkcji, wybrane metadata i porty. `grammatical_tense_interpreted=false`: ta realizacja nie stanowi dowodu czasu zdarzeń ani prawdziwości twierdzenia.
- Evidence #09-01 uruchamia tę samą binarkę przed i po przygotowaniu pakietów. Osiem konstrukcji w dwóch neutralnych domenach — oryginalnych i po przemianowaniu identyfikatorów — uzyskało `ok`, `fully_validated=true`, niezależne `verify-goal: ok` i ewaluację `ok`. Długości ścieżek 2 i 3 oraz restrykcje są jawne w celach. Eksport zwykły i `--no-cache` zachował ten sam `meaning_fingerprint`. To obserwacja tych konkretnych pakietów, nie dowód braku wszystkich możliwych zależności domenowych w całym produkcie.
- Próby kontrolne #09-01 odmawiają nieznanej roli (`UNKNOWN_RELATION_ROLE`), zmienionej restrykcji (`REFINEMENT_CONSTRAINTS_CHANGED`) i podmienionego certyfikatu (`GOAL_FINGERPRINT_MISMATCH`). Wyszukiwanie z wieloma dopuszczonymi długościami zachowuje dwa kandydaty; ograniczenie budżetu kończy się `resource_limit` i `incomplete`.

## Tłumaczenie WH

W evidence #09-02 oba kierunki pozostają niekompletne, mimo że zawierają tekstową kandydaturę:

| Kierunek | Tekst kandydata | Status | Zapisany powód / założenie |
| --- | --- | --- | --- |
| PL → EN, “Kto widzi psa?” | “Who sees the dog?” | `incomplete` | Założenia: liczba pojedyncza i 3. osoba dla podmiotu odpowiedzi oraz rodzajnik określony. Diagnostyka `EXPLICIT_REALIZATION_ASSUMPTION`. |
| EN → PL, “Who sees the dog?” | “Kto widzi psa?” | `incomplete` | Pozostała diagnostyka `REALIZATION_RESIDUALS_REMAIN`. |

Tekstu z niekompletnej kandydatury nie należy przedstawiać jako jednoznacznego, w pełni pokrytego tłumaczenia. Zwykłe wejścia w tych samych konfiguracjach również mają w tych zapisach status `incomplete`.

## Granice profilu

Profil `positive-facts-v1` konsumuje jawne dodatnie fakty dla ograniczonego zapytania. Nie wyprowadza negacji z braku faktu, nie rozstrzyga prawdy świata ani kompletności domeny. Manifest możliwości rozdziela reprezentację od ewaluacji: `Not` i kwantyfikatory można reprezentować, lecz nie są ewaluowane przez ten profil; `Ask` ma ograniczony profil dodatniego ciała. Czas względny i wiele wartości atrybutów referencji pozostają poza zakresem ewaluatora. Szczegóły: [synteza](../../architecture/semantic-algebra/synthesis.md), [fakty](../../architecture/semantic-algebra/facts.md), [zakresy](../../architecture/semantic-algebra/scopes.md) i [manifest capabilities](evidence/09-02/capabilities.json).

## Źródła obserwacji

- [Evidence #09-01: komendy, konstrukcje, certyfikaty i wyniki negatywne](evidence/09-01/README.md), [pełny manifest komend](evidence/09-01/commands.jsonl) i [audyt niezależności](independence-audit.md).
- [Evidence #09-02: manifest komend](evidence/09-02/commands.jsonl), [konstrukcja](evidence/09-02/construct.json), [weryfikacja](evidence/09-02/verify.json), [ewaluacja](evidence/09-02/evaluate.json), [realizacja EN](evidence/09-02/realize-en.json), [realizacja PL](evidence/09-02/realize-pl.json), [WH PL→EN](evidence/09-02/translate-wh-pl-en.json) i [WH EN→PL](evidence/09-02/translate-wh-en-pl.json).
- Stan zadań i SHA poprzedników: [issue map](issue-map.json). Odrębny status akceptacji właściciela pozostaje `NOT EVALUATED`.

# Recenzja PL leksykalizacji CORE

`data/pl/1000-core-review.tsv` jest redakcyjnym zapisem pracy nad issue
G07-03 (pełna polska warstwa językowa na 1000 CORE). Plik ma 1000 wierszy,
po jednym na każdy aktywny `ConceptID` z `internal/catalog/concepts.go`, w tej
samej kolejności co katalog. Jest to artefakt redakcyjny: **nie jest** czytany
przy starcie, nie generuje deklaracji Go i nie jest formą testu ani golda.
Wykonywalnym źródłem prawdy pozostaje typowany Go w `languages/pl/`.

Licencja metadanych: brak importu słowników zewnętrznych. `SOURCE_REF`
wskazuje na autorską decyzję redakcyjną projektu
(`golingua-pl-authored/unknown`), `REVIEWER` to agent wykonujący (codex),
`REVIEW_DATE` to 2026-10-21. Wartość licencji `unknown` pozostaje jawnym
brakiem ustalenia i nie daje prawa do kopiowania danych zewnętrznych.

## Statusy pracy (kolumna `PL_WORK_STATUS`)

Na 1000 wierszy:

- `REVIEWED_LEXICALIZED` — **448**; jest naturalny polski odpowiednik,
  zwykle jednowyrazowy (dla rzeczowników dopuszczalny lekki frazemat).
- `REVIEWED_BROADER_NARROWER` — **14**; polski wyraz jest trwale
  szerszy albo węższy niż definicja CORE (np. gatunkowe zawężenie).
- `REVIEWED_RELATED` — **2**; bez bezpośredniej formy, wiążący jest tylko
  pokrewny wyraz; jawnie bez bindingu do czasu decyzji właściciela.
- `REVIEWED_NO_DIRECT_FORM` — **15**; CORE concept bez jednowyrazowej
  polskiej formy (metakategorie, wyrażenia opisowe, nazwy relacji).
- `DEFERRED` — **521**; wiersz bez decyzji PL w tym batchu, z powodem
  `Redakcja PL przesunieta do kolejnego batcha G07-03`. DEFERRED nie znaczy
  „brak translacji"; znaczy „nie rozpatrzono jeszcze".

## Kolumny

- `CORE_ID`, `CORE_KIND`, `CORE_DEFINITION` — skopiowane z katalogu CORE.
- `DOMAIN` — domena z katalogu (dla 49 wierszy oryginalnych brak domeny
  dziedzinowej; widać to jako `Existing typed CORE definitions`).
- `PL_WORK_STATUS` — status z listy powyżej.
- `PL_LEMMAS` — `lemma:pos`, razem ze śladami homonimii POS (np. `kot:noun`).
- `LEXEME_SENSE_IDS` — pary `lexeme:<id>|sense:<id>` ze stabilnymi
  ASCII slugami (spacje w MWE → `_`, np. `pl.kropla_deszczu.n`).
  Wiersze bez PL lub bez bindingu mają `-`.
- `MAPPING_KIND` — `equivalent` / `broad` / `narrow` / `related` /
  puste (`none`). Nigdy nie jest to zgoda na approved binding.
- `BINDING_STATUS` — `pending` dla sensów z kandydaturą; `none` bez niej.
  **Nie ma żadnego `approved`.** Awans wymaga ręcznej akceptacji właściciela.
- `VALENCY_FRAME_IDS` — identyfikatory ram z `languages/pl/valency.go`
  dla czasowników z istniejącą ramą (`widzieć`, `patrzeć na`, `słyszeć`,
  `dostać`). Inne czasowniki wymagają ram przed dopisaniem do Go.
- `MORPHOLOGY_COVERAGE` — `seed morfologiczny: brak paradygmatu czasownika`
  w znanych przypadkach; pełna morfologia jest seeded tylko dla
  `kot` (rzeczownik męski, zwierzęcy) i patternów seedu.
- `CONSTRUCTION_COVERAGE` — zawsze `-`; seed nie definiuje konstrukcji.
- `SOURCE_REF/LICENSE/REVIEW_DATE/REVIEWER/DECISION_RATIONALE` —
  provenance decyzji + uzasadnienie po polsku (uzasadnienia zapisano
  bez diakrytyków, ASCII).

## Zakres implementacji Go w tym batchu

Go w `languages/pl/` obejmuje siedem grup domenowych: People (59 leksemów /
60 sensów), nature (95 leksemów; ręcznie dopisane sensy polisemii
`ryba.food` i `owoc.food`), body (40/41), places (osady/budynki/
pomieszczenia/trasy, 56), objects (obiekty domowe/narzędzia/maszyny, 79),
food (jedzenie/napoje/okrycia, 58) oraz basics (23: resztka oryginalnych 49,
pozycje osobowe `kurier`/`wolontariusz`, rzeczowniki właściwości).
Dla homonimu `staw` wyodrębniono osobny leksem `pl.staw_zbiornik.n`
(zbiornik wodny), a polisemiczny leksem `pl.skora.n` łączy sens `part`
(organ) i `material` (surowiec). Batch 4 autorsko dopisał ramy walencji i leksemy dla
siedmiu czasowników rdzeniowych (`pytać`, `odpowiadać`, `czytać`, `uczyć`,
`uczyć się`, `jeść`, `spać`) w `lexicon_core_verbs.go`. Batch 5 rozpatrzył
54 kolejne wiersze DEFERRED: 18 rzeczowników substancji i korzenia
(implementacja w grupie nature) oraz 36 czasowników lokomocji (decyzje
w worklist bez ram walencji). Pozostałe 467 wierszy DEFERRED czeka na
redakcję kolejnych batchów.

Znane przypadki jawne, do decyzji właściciela, bez wymuszonej decyzji
automatycznej:

- `pl.ksiazka.work` (drugi sens „książki", = dzieło) nie ma pasującego
  conceptu CORE i w Go pozostaje bez bindingu.
- `pl.dom.household`, `pl.szkola.institution`, `pl.reka.arm`,
  `pl.dziecko.kinship` mają binding `candidate`/`pending` do dopuszczalnych
  CORE conceptów (w tym `concept:child_of_person`), „candidate" nie uprawnia
  do tłumaczenia.

## Ocena ręczna

Ręczny przegląd wyników tłumaczeń i form na przykładach właściciela jest
**NOT RUN** dla tego batcha. Liczby w worklist opisują decyzje redakcyjne,
nie poprawność językową.

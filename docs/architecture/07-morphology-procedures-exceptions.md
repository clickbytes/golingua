# 07 — Proceduralna morfologia i formalne wyjątki

**Status:** PROPOSED. Issues: [G02 #3](https://github.com/clickbytes/golingua/issues/3), [G04-03 #40](https://github.com/clickbytes/golingua/issues/40), [G07-05 #62](https://github.com/clickbytes/golingua/issues/62). Source: `js/core/morphology/{engine,operations,compiler}.js`, `languages/{pl,en}/morphology.json`, pinned `docs/adr/morphology-backend-scaling.md`.

## 1. Zasada: nie zapisujemy regularnej formy, jeśli mamy dowód jej odtwarzalności

**Jednostki:** `MorphLexeme`, `Paradigm`, `FeatureBundle`, `StemSet`, `PrincipalParts`, `CellRecipe`, `ExceptionRule`, `FormVariant`, `MorphReading`. `MorphLexeme` istnieje **bez CORE mappingu**, bo można odmienić słowo, którego semantyka nie została zatwierdzona.

`Generate(lexemeID, featureBundle) → wszystkie licencjonowane powierzchnie`; `Analyze(surface) → wszystkie licencjonowane (lexemeID, featureBundle, variant, source)`. API nie zakłada jednego lematyzatora na string, a `inverse` nie rozstrzyga polisemi podczas odmiany.

**Przykład EN:** regularne `rock + past` może dać `rocked` z reguły; nieregularne `go + past` wymaga przechowywanego principal part `went`. To są **formy czasowników języka EN**, nie węzły ConceptGraph. Polisemia `rock` (skała/muzyka/kołysać) dotyczy SenseID, a nie obowiązku osobnego paradygmatu dla każdego sense.

## 2. Potok dla jednej komórki

`(MorphLexemeID, requestedFeatureBundle)`
→ validate profile/POS/bounds
→ resolve Paradigm and inheritance (acyclic)
→ identify legal cell and required principal parts
→ select stem/allomorph
→ execute typed steps (bounded; no eval)
→ apply morphological boundary/orthography policy
→ resolve explicit overrides
→ emit ALL licensed variants+trace
→ insert into inverse index.

Ważne: rozstrzygnięcie `block/replace` i konfliktów należy do kompilacji i wyboru ścieżki, aby zablokowana komórka **w ogóle nie uruchamiała fallback rule**. Niezbędne puste formy (`zero exponent`) są różne od `missing/unavailable cell`.

## 3. FeatureProfile i principal parts

- Wspólna techniczna unifikacja w `internal/features`, językowe profile niezależne. `case=gen/nom/acc` w PL nie jest cechą globalnej ontologii; EN może nie mieć analogicznego wymiaru dla zwykłych rzeczowników.
- Inherent features leksemu (rodzaj/żywotność/klasa) różnią się od requested inflectional features (liczba, przypadek, tense, person) i od semantycznych cech MeaningGraph.
- `PrincipalParts` to **minimalny wystarczający zestaw diagnostycznych form**, ale nie twierdzenie, że samo lemma zawsze go zastąpi. Przy EN `go` trzeba znać `went`, `gone` tylko jeżeli odpowiadające komórki należą do wspieranego profilu; przy PL `pies` nieregularne stem `ps-` wynika z jawnej alternacji albo dodatkowego tematu.
- Klasa odmiany może być proponowana na podstawie importowanych komórek, jednak kilka zgodnych klas = `ambiguous/pending`. Dobór klasy nie jest automatycznym zatwierdzeniem brakujących form.

## 4. Zamknięty zestaw operacji (przykładowe rodziny)

`append, prepend, stripSuffix, replaceSuffix, selectStem, appendStem, alternateStem, harmony, rootTemplate, reduplication, orthographyVariant`. W **G02-03** trzeba wybrać minimum PL/EN na podstawie goldów; nie obiecywać pełnej tureckiej harmonii lub arabskich broken plurals. Source JS ma już bounded `harmonicSuffix`, `rootTemplate` i inne, lecz nie pełną uniwersalną fonologię.

Plan jest tablicą typowanych operatorów; brak funkcji jako stringi, `eval`, regular expressions z importowanej konfiguracji, rekurencji bez budżetu lub I/O.

## 5. Normatywna semantyka wyjątków

| Effect | Co robi | Co dzieje się z regułą bazową |
| --- | --- | --- |
| `block` | usuwa komórkę/tryb z legalnych wyników | generator bazowy **nie** jest uruchamiany |
| `replace` | jawne wyjścia zamiast wyniku regularnego | wynik bazowy **nie** jest emitowany |
| `addVariant` | legalny dodatkowy wynik z style/region/register | bazowy pozostaje, chyba że osobno zablokowany |
| `stemOverride` | wybiera zatwierdzony temat w zakresie cech | pozostałe komórki bez zmian |
| `suppletion` | podaje nieprzewidywalną formę/temat | w danej komórce działa jak precyzyjne replace; bez fikcyjnych reguł |
| `constructionException` | ogranicza licencję ramy/voice/adposition | należy do Valency/Construction, **nie** do morph cell |

Minimalne pola: `ExceptionID`, `scope (lexeme|paradigm|cell)`, `featureSelector`, `effect`, `overrideOf`, `source/provenance`, `review`, `positiveWitness` i `negativeWitness`. W przypadku dziedziczenia pierwszeństwo nie zależy od kolejności pliku: najpierw jawna legalność i konkretność scope, potem weryfikowany `overrideOf`; konflikt dwóch `replace` bez rozwiązania → **invalid compilation**. Nie mieszać różnych standardów i dialektów bez jawnej adnotacji.

### Negative cases, które muszą FAIL-ować

- `replace(past:went)` dla go, ale generator nadal zwraca `goed` → FAIL.
- `block` genitive działa, a fallback regularny generuje niedozwoloną formę → FAIL.
- `addVariant` zastępuje bazowy wynik bez deklaracji replace → FAIL.
- Ten sam homograf ma dwa lub więcej legalnych odczytań, a analyzer zwraca tylko pierwszy → FAIL.
- Dwie niezgodne reguły wyjątku w tej samej komórce przetwarzane zależnie od kolejności rekordów → FAIL.
- Import wskazuje brak komórki i system traktuje go jako twardy dowód `block` → FAIL.

## 6. Formalna zgodność kierunków

Dla każdej zatwierdzonej komórki i każdej legalnej wersji formy:

`(lexemeID, featureBundle, variant) ∈ Analyze(Generate(lexemeID,featureBundle).surface)`.

To **inkluzja**, nie wymóg unikalności; jedna forma może reprezentować wiele przypadków, czasów, sensów. Oddzielnie walidujemy `external independent observed forms`, bo generator i analyzer mogą dzielić ten sam błędny przepis. Nie należy dopasowywać expected do aktualnej implementacji.

## 7. Precompute kontra lazy kontra FST

W JS `createMorphology` rozwija całe bounded inventories do inverse index. To już działa i jest punkt odniesienia, nie „ręczny słownik bez procedur”. Go v0 ma na start **finite relation table** (morph generator + inverse index), ale fizyczny nośnik (map/arrays/CSR/FST) pozostaje szczegółem backendu.

Porównywać:
- build/cold start i RSS;
- per-form `Generate/Analyze` wraz ze wszystkimi readings;
- rozmiar pack i diff po zmianie jednego wyjątku;
- pełne parity z zatwierdzonymi goldami i negatywami.

FST nie wybiera się, dopóki pomiar nie potwierdzi korzyści; „FST indeks słów” to nie to samo co dwukierunkowy **transduktor morfologiczny**. Nie ogłaszać 2×/10× bez raw data i tego samego hosta.

## 8. Dane i import

Źródło PL: 27 paradygmatów, 465 wiązań, **939** override declarations; EN: 13 paradygmatów, 462 wiązania, **65** override declarations — statyczny snapshot pinned SHA. Jeden wyjątek może dotyczyć jednego slotu, a licznik nie jest wskaźnikiem błędów czy liczby ręcznie opisanych słów.

W G04 importer dzieli źródłowe `forms` na `evidence(candidate)` i `rules(candidate)`; kompiluje dopiero zatwierdzone reguły. Powtarzalny import tych samych rekordów nie zmienia pakietu. Nie wolno zgubić explicit morphology evidence przy „optymalizacji” do paradygmatu.

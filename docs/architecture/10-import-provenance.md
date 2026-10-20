# 10 — Import leksykonów, conceptów i konstrukcji: staging/review/promote

**Status:** PROPOSED. EPIC [G04 #5](https://github.com/clickbytes/golingua/issues/5); docelowe 1000 conceptów [G07 #8](https://github.com/clickbytes/golingua/issues/8). Legacy baseline [flexLingua@79c4a958](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92).

## 1. Co importujemy osobno

| Strumień | Rekord staging | Po zatwierdzeniu |
| --- | --- | --- |
| CORE definicje | `ConceptCandidate` + typed relation edges | ConceptDefinition/ConceptGraph |
| Słownik PL/EN | `MorphLexemeCandidate`, `LexicalSenseCandidate` | LexicalGraph i MorphLexeme |
| Mapowanie sensu | `SemanticBindingCandidate` (equivalent/broader/narrower/related) | approved binding lub pending; **bez auto-equivalence** |
| Odmiana | `FormEvidence`, `ParadigmCandidate`, `Stem/ExceptionCandidate` | MorphologyGraph i finite relation table |
| Walencja | `ValencyFrameCandidate`, source frame/role constraints | ValencyGraph dla zatwierdzonego SenseID |
| Konstrukcja | `ConstructionFamily/VariantCandidate`, bridge mapping | ConstructionGraph i compiled plans |
| QA | `EvidenceRecord`, human-reviewed independent expected | evidence index, case audit, capability gating |

**Morfologia działa bez SemanticBinding.** Import nie może odrzucać prawidłowego lematu tylko dlatego, że CORE go jeszcze nie zna. I odwrotnie: żadna zatwierdzona forma nie jest dowodem konkretnego konceptu.

## 2. Siedem etapów importu (deterministycznie i atomowo)

1. **Snapshot source**: wskazane repo/URL, commit/revision, exact path i sha256 contents, źródło/autor, licencja i terms. Jeżeli brak praw do redystrybucji — staging-only / blocked. Rozróżnij własny source JS i external provider; każdy ma własne licencje.
2. **Parse + normalize**: strumieniowo, limity bytes/rows/codepoints, UTF-8/NFC, oryginalne tagi i linie, parser bez wykonywania kodu. Dysk zewnętrzny i ścieżki nie mogą przejść poza staging root.
3. **Map schema**: stable lexical IDs i źródłowe semantics. Zachowaj rozdzielenie forms/sense/valency/construction. Unknown field/unknown tag → reject candidate z diagnostyką, nie hidden delete.
4. **Suggest proceduralization**: form evidence vs current paradygmaty, candidate principal parts, dodatkowe tematy i wyjątki. Dopasuj wszystkie potwierdzone komórki, pokaż konflikty i kilka równie dobrych klas. **Suggest ≠ approve**.
5. **Validate cross refs**: symbol collisions, dangling SenseID/CoreRef, typeof expected role, language profiles, license conflicts, feature cardinality i malformed definitions.
6. **Review/approve**: zatwierdź konkretne candidate IDs i exact input hashes plus niezależne pozytywne/negatywne witness IDs. Zamknięty stan: `unverified → proposed → reviewed-approved / reviewed-rejected`. Review to zapis autora decyzji i treści, nie automatyczny podpis kryptograficzny; jeśli potrzebny cryptographic signature, wymaga osobnej, rzeczywistej implementacji key management.
7. **Promote/compile**: transakcja waliduje cały zbiór, dopiero potem tworzy nowy immutable pack snapshot i atomowo aktualizuje aktywny reference. Nie mutuj live graph in place. Failed promote leaves previous pack intact; run comparison report.

**Idempotentność:** ponowny import identycznych rekordów nie powiela leksemów ani zmienia referencji. Wersje źródeł są jawne, conflict nie jest rozstrzygany „ostatni import wygrywa”.

## 3. Algorytm przypisania paradygmatu: proposal, nie magiczne lematyzowanie

Dla `lemma+POS+observed forms`:
1. Znajdź kandydatów dozwolonych dla języka i POS w indeksie klas.
2. Sprawdź zgodność wszystkich zaobserwowanych komórek; brak komórki w źródle nie znaczy, że forma nie istnieje.
3. Wylicz minimalny jawny zestaw principal parts, które rozróżniają kandydatów; uwzględnij alternacje tematu.
4. Zaproponuj dozwolone `replace/addVariant/block` tylko z dowodem licencji. **Brak formy w UniMorph nie pozwala zaproponować block jako fakt.**
5. Pozostaw kilku kandydatów, jeśli dane nie rozstrzygają; ograniczenie automatyczne nie może wytworzyć pewności.
6. Przechowuj `FormEvidence` z provenance jako niezależną kontrolę po kompilacji i analizy odwrotnej.

Koszt/score dopasowania paradygmatu służy wyłącznie do priorytetyzacji pracy redakcyjnej, nie do auto-activation regularizacji nieznanych leksemów.

## 4. Źródła zewnętrzne i granica licencji

Rozważane źródła specjalistyczne: dla PL morfologia (SGJP/PoliMorf/Morfeusz), walencja (Walenty), semantyka (Słowosieć); EN morphology UniMorph, walencja VerbNet, sensy Open English WordNet. Sam fakt, że materiał jest dostępny publicznie, **nie daje automatycznie prawa importu/redistribucji**. Przed aktywacją wskazać konkretną wersję, licencję, atrybucję i zakres użycia; np. dane fleksyjne a pełny SGJP mogą mieć różne warunki. Nie wstawiaj żadnych zewnętrznych definicji słownikowych bez sprawdzenia praw. Zewnętrzne synsety to kandydaci na mapowanie, nie gwarantowane tożsame koncepty.

## 5. Import source pinned i rozliczenie liczebności

Statyczny anchor: source CORE 1002 IDs / PL 1005 / EN 1006 lexeme entries / PL 465 / EN 462 morphology bindings / 51 event grammar frames per language / 943 common lexically referenced concept IDs. Import raportuje:
- `sourceRead`, `staged`, `rejected`, `needsReview`, `approved`, `packCompiled`;
- count per language/POS/semantic equivalence kind;
- lost/unsupported source fields by exact source JSON pointer;
- side-by-side link źródłowe `forms`→paradigm/exception/evidence;
- constructor plan dedup only after structural and valency differences checked (PL structural shapes != EN).

**Nie** raportować przetworzenia 1002 CORE nodes jako 1002 approved bilingual concept mapping. Cel 1000 jest osobną bramką G07; nie wolno dosztukować 57 dowolnych concept nodes.

## 6. Negative tests

- Nieznane znaczenie słowa z poprawną fleksją zostaje w morfologicznym katalogu, ale nie może być tłumaczone jako losowy concept.
- Dwa synsety powiązane relation `related` nie uzyskują `equivalent`.
- Brak, konflikt lub nieważna licencja → rekord nie zostaje zatwierdzony.
- Import A zakończył się z błędem na ostatnim rekordzie → aktywny pack pozostaje poprzedni, bez fragmentów A.
- Nieznany tag UniMorph i brak odpowiadającego dimension → rejected/staged review; nie znika.
- Zmiana source checksum/revision musi skutkować odmową lub jawnym oddzielnym importem; nie hidden refresh.
- Dwa bridge frames podobne lecz o różnym case/PP/roleScope nie są zmergowane jako ten sam plan.

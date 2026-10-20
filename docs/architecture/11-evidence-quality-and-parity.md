# 11 — EvidenceGraph, niezależne testy semantyczne i benchmark

**Status:** PROPOSED; zadania [G00-04 #14](https://github.com/clickbytes/golingua/issues/14), [G05-08 #51](https://github.com/clickbytes/golingua/issues/51), [G06-06 #57](https://github.com/clickbytes/golingua/issues/57), [G07 #8](https://github.com/clickbytes/golingua/issues/8), [G08 #9](https://github.com/clickbytes/golingua/issues/9).

## 1. EvidenceGraph — typowane źródła i dowody, a nie dekoracja

`EvidenceRecord` powinien mieć: `EvidenceID, Kind, SubjectRef, SourceRef (file/revision/hash/span), ReviewStatus, Reviewer, ExpectedCaseID, RuntimeSHA/PackHash (dla wyników testów), TestResult (PASS/FAIL/INFRA/NOT_RUN), Scope, Notes`. To **proponowany logiczny kontrakt**; G04/G08 muszą go wdrożyć i zindeksować. Nie wpisuj reviewer ID człowieka, jeżeli weryfikację zrobił agent bez jawnej delegacji; autor testu ≠ recenzent oracla.

**Istotne rozdzielenie proof gates:** zatwierdzony `SemanticBinding` potrzebuje rzetelnego źródła, recenzji zgodności sense↔CORE i lexical witness; **nie** musi mieć gotowej analizy całego zdania. Natomiast ogłoszenie `sentenceAnalyze/sentenceRealize/dialogueActRealize: implemented` wymaga zatwierdzonego passing-case odpowiedniej konstrukcji/operacji na aktualnym pack SHA. Wymaganie 1000 dwujęzycznych sensów nie może po cichu oznaczać 1000 zdaniowych certyfikatów.

Krawędzie evidences:
- SourceForm ↔ MorphLexeme/Paradigm/Exception/FeatureBundle;
- Review ↔ SemanticBinding (approve/exactness);
- ApprovedCase ↔ ConstructionFamily+Variant / MeaningGraph / SourceText;
- TestOutcome ↔ specific runtime SHA / pack SHA / case ID;
- ProofStep ↔ source Fact/Rule/valid time/scope.

Źródłowy `language graph` i stara `manifest.capabilities` to tylko wejście migracyjne; **capability=true** wymaga skompilowanego provider, rzeczywistego testu i reviewed evidence. Green CI nie dowodzi nieprzetestowanego wariantu.

## 2. Legacy corpus — dokładny i uczciwy punkt odniesienia

Na pinned `flexLingua/main@79c4a958` katalog `tests/cases/**` zawiera **70 plików i 593 case records**, w tym **591 required+approved**, dwa pozostałe to proposed/exploratory. Deklarowane oczekiwane klasy obejmują **489** `ok`, **71** `unsupported`, **30** `ambiguous`, **2** `resource_limit` oraz **1** `invalid`. To **liczby oczekiwań**, nie wyniki własnego uruchomienia Go. Stary GitHub Actions `translation-matrix` miał PASS na przypiętym source SHA; nie jest to automatyczny dowód równoważności nowej implementacji.

Dla każdego migrowanego case zachowaj stable legacy case ID, version/source SHA, oryginalne expected fields (text accepted alternatives, meaning, roleSet, features, nested scopes, diagnostics, informationLoss, graph assertions, source spans, preview selectors). Jeżeli schema nie mieści się w nowym G00 format — `NOT_RUN / migration_schema_unsupported`, nie ignoruj wymagań.

## 3. Dwie rodziny testów, żadna nie zastępuje drugiej

**Independent sentence gold:** ręcznie określony sens **przed** obserwowanym wynikiem Go; zgodność tekstu z zaakceptowanym zbiorem wariantów i MeaningGraph po walidacji, dokładne role/polarity/time/scope/referents oraz allowed informationLoss. Dopuszczalna odmowa testowana oddzielnie od sukcesu. Nie można „naprawić” testu przez przepisanie expected do output z błędnego engine.

**Property/metamorphic tests:** konstrukcja/generator zgłasza warunek zachowania semantyki. Przykłady:
- PL SVO/OVS, gdy rola jest rozwiązywalna: role identyczne, surface różne.
- EN „The dog sees the girl” kontra „The girl sees the dog”: różne MeaningGraph.
- Zamiana `agent↔theme`, polarity, tense, `clockTime`, `calendarDate`, quantifier lub referenta **musi zostać wykryta** jako zmiana znaczenia.
- `Analyze(Generate(l,f))` zawiera zatwierdzoną parę `(l,f)` i wszystkie homografy.
- `block` brak illegal surface nawet przez fallback; `replace` nie dopuszcza regular wrong form; same rules in shuffled JSON produce same pack.
- Query with only question/hypothetical/typical edge cannot assert positive fact.
- No explicit world fact -> `unknown` not automatic `false`.
- Incomplete parser search or resource limit -> never unique complete.
- Capability without approved evidence -> fails activation.

## 4. Held-out compositionality bez leakage

Nie testować generalizacji poprzez losowy podział pojedynczych zdań ze wspólnym SenseID i niemal identyczną ramką. Dla ograniczonego modelu procedur:
- dzielić po **rodzinach konstrukcji**, po leksemach/sensach i po nowych kombinacjach już licencjonowanych cech;
- testować osobno legalne nowe kompozycje i pozornie gramatyczne, lecz niedozwolone kombinacje;
- trzymać testy **poza** zbiorem użytym do inferencji paradygmatów/wyboru wyjątków;
- raportować coverage denominators i false-positive/error severity, a nie tylko ogólne „accuracy”.

Wynik naturalności/generowania (płynność) nie zastępuje semantycznego checku. Wyrażenie może brzmieć naturalnie i jednocześnie kłamać co do roli lub referenta.

## 5. Ocena readiness per capability i per język

Każdy koncept/mapping ma **niezależne** poziomy: catalog → bilingual → morphology-ready → construction-ready → sentence-witnessed → generation-witnessed → dialogue-witnessed. Raport: exact numerators/denominators + invalid/pending/unreviewed. Nie rysować 100% na pustym denominatorze. Jeden paradygmat/lexeme nie liczy się jako 20 konceptów tylko dlatego, że ma 20 form.

Stabilność indeksów i deterministyczny pack digest weryfikować na kilku kolejnościach wejścia. Szacunki szybkości należy opierać na **ten sam corpus, to samo znaczenie i kompletność readings**, podobny host, repeat runs, build vs runtime odseparowane.

## 6. Benchmark JS kontra Go — reguły raportu

**Workload:** verified PL/EN morphological generate+analyze, construction compile, parse/realize sentences including refusals/ambiguities, selected reasoning requests; pinned source and pack SHA.

**Mierz i publikuj:** go/node version, OS/CPU, inputs hashes, repeats/warmup, cold-start, time per query P50/P95, allocations/op, max RSS/heap, output bytes, import/build time, status per case i semantic parity count. Nie ustalaj absolutnego SLA z benchmarku jednego laptopa. Przy porównaniu odmów i niekompletności Go, liczba poprawnie rozwiązanych zadań jest pierwszorzędna.

Decyzje FST, lazy generation, CSR specjalizowane, SQLite albo własny pack format uzależnić od **zmierzonego konkretnego problemu**, nie od nazwy technologii. Skrócenie runtime kosztem utraconej wieloznaczności jest regresją, a nie optymalizacją.

## 7. CI, statusy i granice dowodów

Jeden kanoniczny `go-ci` test/vet/build + przyszłe go fixtures gate. Przed PR: poprawnie opisany owner, SHA, expected i negative witness. PR CI PASS na head, review diff, merge, CI PASS na merged SHA; task zamknąć dopiero po realnym DoD. Brak lokalnego worktree dostępu = lokalne testy `NOT RUN`, nie „są na pewno OK”.

Raport rozróżnia: `PASS`, `FAIL`, `INFRA` i `NOT_RUN`; domenowo `ok/unsupported/ambiguous/incomplete/invalid/resource_limit/unknown(conflict)`. Kontrolowana odmowa może być poprawnym wynikiem **testu odmowy**, ale nie dowodem skutecznego tłumaczenia. Oczekiwane braki nie liczą się jako PASS funkcjonalnego pokrycia.

Równolegle mogą istnieć planowane issues, **nie** dwa aktywne PR/agent tasks. Epic zamykamy po spełnieniu wszystkich DoD z linkami do dowodów na SHA, nie po samym odhaczeniu children.

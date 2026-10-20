# Architektura golingua — kanoniczny suplement implementacyjny

**Status:** DESIGN SPECIFICATION / PROPOSED (chyba że wskazano inaczej). **Repo baseline:** `golingua/main@a968840b5669723bacf237d2aad58d3a284d9be6`. **Source baseline:** [flexLingua/main@79c4a9586f7f94b6065da12dc25c4cb272ebdc92](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92).

Nie jest to dokumentacja *gotowego* modelu Go. Na rozpoczęciu G00-02 Go implementuje jedynie CLI bootstrap i workflow `go-ci`. Następne wymienione typy i operacje są **docelowym kontraktem**, podzielonym na issues, i nie wolno deklarować ich capability, dopóki nie ma kodu, zatwierdzonych dowodów i PASS na właściwym SHA.

## Czytaj w tej kolejności

1. [00 — produkt i miara gotowości](00-product-contract.md): cel bounded light Graph Language Engine, 1000 konceptów, nie LLM ogólnego zastosowania.
2. [01 — granice grafów, identyfikatory, DAG pakietów](01-graph-layers.md): jedyne źródło odpowiedzialności za Concept/Lexeme/Sense/Valency/Construction/Procedure/Evidence/Meaning/Fact.
3. [02 — ścieżki wykonania i stany](02-runtime-pipeline.md): kompilacja, analysis, realization, answer, zakresy i braki dowodów.
4. [03 — format danych i wersjonowanie](03-data-decisions.md): human-review staging, immutable runtime, FST/DB decyzje odroczone.
5. [06 — konstrukcje i walencja](06-constructions-and-valency.md): pierwszorzędna jednostka granulacji, shared families, case/governance, generatory fraz.
6. [07 — morfologia i proceduralne wyjątki](07-morphology-procedures-exceptions.md): klasy, principal parts, odwracalność, precedence i brak nielegalnego fallbacku.
7. [08 — semantyka, wiedza i wnioskowanie](08-meaning-facts-reasoning.md): operatory, referenty, proof i brak halucynowanych faktów.
8. [09 — języki jako pakiety](09-language-plugin-contract.md): manifest, zdolności, źródła, brak natywnego dynamicznego pluginu.
9. [10 — import i provenance](10-import-provenance.md): source-to-new conversion, approval i promowanie atomowe.
10. [11 — testy, evidence, wydajność](11-evidence-quality-and-parity.md): kontrakty golds, coverage, held-out, benchmark.
11. [12 — decyzje i ryzyka](12-decisions-and-risks.md): accepted design direction / proposed detail / deferred; także decyzja o **odrzuceniu wymaganego API `ROCK.VERB.MORPH.PAST`**.
12. [13 — powiązanie dokumentacji z issues](13-issue-traceability.md): EPIC G00–G09, numery issues, bramki.
13. [14 — przypadki przekrojowe](14-worked-vertical-slices.md): PL/EN, wyjątek, ambig/negative, reasoning.

**Zarezerwowane ścieżki następnych zadań:** `04-results-and-limits.md` jest DoD [G00-03 #13](https://github.com/clickbytes/golingua/issues/13), a `05-parity-policy.md` jest DoD [G00-04 #14](https://github.com/clickbytes/golingua/issues/14). Nie interpretuj ich nieobecności jako gotowych implementacji.

## Kolejność prawdy

1. Jawna decyzja właściciela i zatwierdzone, niezależne oracles semantyczne.
2. Wersjonowany **zatwierdzony** kontrakt w tym katalogu (projekt, nie udowodniona funkcjonalność).
3. Faktyczne API/kod/fixtures w golingua na konkretnej rewizji.
4. GitHub CI i zaakceptowane wyniki testów na właściwym SHA (dowód wykonania).
5. Legacy flexLingua: punkt porównania, **nie wyrocznia tam, gdzie zachowanie źródła jest błędne**.
6. Issues opisują plan, nie zastępują powyższego kontraktu.

Jeżeli dokument i issue się różnią: **przerwij implementację danego punktu, opisz konflikt w issue i rozstrzygnij go przed kodowaniem**, zamiast potajemnie wybrać wygodniejszą interpretację. Obecna dokumentacja jest draftem z przyszłymi nazwami; stabilność wire formatów zostanie potwierdzona dopiero issue implementującym dany typ.

## Jedno miejsce decyzji

Single agent, jeden ACTIVE EPIC, jedno ACTIVE child issue, jeden otwarty implementation PR. Stan Epiców: [G00 ACTIVE](https://github.com/clickbytes/golingua/issues/1), G01–G08 QUEUED, G09 DEFERRED. Planowanie i poprawianie specyfikacji to nie równoległa implementacja. Każdy agent zaczyna od AGENTS.md, dokumentacji zakresu, issue i aktualnego main SHA.

## Statusy funkcjonalne

- **IMPLEMENTED:** kod + test/capability evidence + CI PASS na danym SHA.
- **PROPOSED:** niniejszy target contract; implementator ma dopiero go stworzyć.
- **DEFERRED:** poza v0, nie może być dodawane boczną ścieżką.
- **UNKNOWN/NOT RUN:** nie ma wystarczającego dowodu; nie jest to FAIL ani PASS.

W szczególności nazwa *Graph Language Model* jest celem badań nad **kontrolowanym symbolem, gramatyką i planowaniem**, a nie twierdzeniem o ogólnej inteligencji lub jakości neuronowych LLM.

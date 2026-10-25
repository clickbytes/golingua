# R3 czas przeszły widzenia i uzgodnienie

Issue #357, epik #354; base a27311ce0e9f5d4fb4c9ebca3680f58e60603b97. Wdrażamy dodatnie seeing w past, 3sg/3pl, z istniejącymi nominalnymi podmiotami. I zachowuje past, role i one/many; lexical gender PL służy odmianie i nie staje się płcią biologiczną w I. Owner acceptance NOT_EVALUATED.

EN saw korzysta z rodziny transitive_past, która nie wymusza zgody liczby na synkretycznej formie past. Nie naprawiamy braku number przez dopisanie pozornej cechy do saw. Wybór present/past family musi odpowiadać jawnej ramie I; brak czasu pozostaje odrębnie opisanym założeniem dotychczasowej trasy. Reading/receiving zachowują zachowanie.

PL widzieć ma imperfective past. Dodajemy potrzebne 3. osobowe komórki i jawne alternacje tematu; nie zmieniamy aspektu na receiving perfective. Podmiot teacher one daje widział, teacher many widzieli, mother one widziała, dog many widziały. Odmiana i analiza są proceduralne na zadeklarowanych częściach głównych. Nie dodajemy future, ogólnego past hearing ani negatywnego past.

Wymagane: pełne source→I→target i niezależna realizacja autora I, obecne czasy reading/receiving/present jako kontrole. Niepoprawna zgoda nie może być pełnym OK. Mały writer otrzymuje jawne ścieżki i jeden przebieg z najwyżej jedną korektą; koordynator odbiera CLI, focused testy, pełną macierz oraz rzeczywisty CI PR/main. Kontrakt i dowody bez zmiany expected pod output.

Po pełnym review koordynator otrzymał od wykonawcy ograniczoną propozycję dodania tense=past do generic EN transitive_past.svo. Bez tego present obchodził agreement i łamał istniejące REF-EN-001/002. Koordynator integruje tylko ten warunek w languages/en/constructions/transitive_past.json, bez zmiany expected i bez nowej pętli implementacji.

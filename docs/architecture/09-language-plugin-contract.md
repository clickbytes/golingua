# 09 — PL i EN jako niezależne pluginy danych

**Częściowo wdrożone.** Jeden silnik, dwa języki. W G02 dodano mały pakiet EN dla generowania i analizy morfologicznej; rejestr pluginów runtime, parser i realizacja zdań pozostają przyszłym zakresem. Nie pisz `TranslatePolishToEnglish` jako osobnego kompilatora w `core`.

## Kontrakt logiczny pluginu

Każdy pakiet języka dostarcza lokalne profile cech, tokenizację/ortografię, MorphLexemes, SenseBindings, paradygmaty i wyjątki, walencję, rodziny konstrukcji/operatory oraz możliwe do wywołania procedury.

`languages/en` obecnie dostarcza wersjonowany profil, manifest, ograniczone paradygmaty oraz seed MorphLexemes z principal parts i Go API `Generate`/`Analyze`. To pakiet morfologii, nie pełny plugin językowy: manifest deklaruje sentence analysis i meaning realization jako unsupported. PL ma profil cech, ale nie ma jeszcze aktywnego seed generatora.

Język udostępnia deklaracje możliwości dla:
- analizy morfologicznej,
- generowania form,
- analizy zdania,
- realizacji MeaningGraph,
- ewentualnie realizacji odpowiedzi.

Deklarujemy `implemented` dopiero, kiedy realna metoda istnieje i można ręcznie zobaczyć jej wynik. Nie potrzebujemy ID formalnych goldów ani zdawania automatycznej macierzy przykładów, aby włączyć funkcję; **status nie może jednak obiecywać więcej niż realizuje kod**.

## Typowana granica API

`LanguageID` i `FeatureBundle`, sensy, konstrukcje i procedury pozostają typowane. Runtime może mieć zwykłe `Analyze(language,text)`, `GenerateForm(language,lexeme,features)` oraz `Realize(language,meaning)`. Dokładne Go nazwy powstaną w odpowiednich taskach.

Brak modułu → jawne `unsupported`; nie ma fallback „udawaj tłumaczenie” w innym języku. Nie wykonuj arbitralnych callbacków z JSON. Pakiety nie są Go dynamic .so; można je kompilować z deklaracji do indeksów.

## Przykłady różnic językowych

Polski ma system przypadków, zgody i wariantów szyku, angielski ma własną politykę determinerów, szyku i pomocniczych. Cechy składniowe nie są globalną cechą ontologii. `semantic agent` nie znaczy globalnie `nominative subject`.

## Rozszerzenie zakresu

Obecnie tylko PL/EN. Dodanie kolejnego języka wymaga nowego katalogu i profili; nie zmieniamy CORE aby dopisać listę przypadków trzeciego języka. Rozdzielone mechanizmy pozwolą rozbudowywać system później bez przepisywania generatora od nowa.

## Manualna ocena

Właściciel podaje zdanie, język i oczekiwaną korektę. Realny CLI pokazuje tekst/MeaningGraph/diagnostykę (w miarę implementacji). Nie potrzebujemy approved golds, sentence acceptance runnera ani testów w CI. Kompilacja pakietu ma natomiast nadal odrzucać błędne identyfikatory, cykle i niekompletne definicje.

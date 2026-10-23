# 02 — Przepływy wykonania (nowy Go)

**Zaimplementowano ograniczone szlaki zdań PL/EN dla `seeing`, `hearing`, `looking` i `receiving`.**
Pierwsze trzy obsługują jawne formy teraźniejsze 3. osoby liczby pojedynczej;
`receiving` obsługuje angielskie `got` i polski czas przeszły 3. osoby liczby
pojedynczej z uzgodnieniem rodzaju. Inne sensy, formy i profile gramatyczne
pozostają poza zakresem.

`looking` wiąże angielskie `look at` z `at` + NP oraz polskie `patrzeć na` z
`na` + NP w bierniku. Szlak obsługuje tylko jawne formy 3. osoby liczby
pojedynczej i zapisane formy rzeczowników. Niejednoznaczne odczytania rzeczowników
pozostają widoczne; realizacja PL→EN zgłasza `incomplete`, gdy rodzajnik `the`
został przyjęty jako jawne założenie.

## Kompilacja wiedzy

`źródło danych JSON (zapisane reguły) → walidator → Concept/Sense/Paradigm/Valency/Construction/Procedure → skompilowane, niemutowalne plany i indeksy`.

Zamiast tworzyć tysiące zagnieżdżonych obiektów jak `ROCK.VERB.MORPH.PAST`, łączymy wspólne struktury numerycznymi ID. Materializujemy tylko to, co rzeczywiście potrzebne do działania. Typy danych muszą być prostsze niż w prototypowym JS.

## Analiza

`text → tokenizacja → wszystkie odczytania morfologiczne → kandydaci sensów → dopuszczone konstrukcje i walencja → MeaningGraph`.

Nie wybieramy kandydata wyłącznie według kolejności importu. Gdy brak słowa/cechy/konstrukcji — `unsupported`/`unknown`; gdy kilka znaczeń — `ambiguous`. Jeśli limit przerwie analizę, wynik ma być niekompletny, nie fałszywie pewny.

Jeżeli fraza ma co najmniej jeden zatwierdzony sens roli i konkurencyjny sens
bez zatwierdzonego bindingu, kompozycja zachowuje graf tylko dla
zatwierdzonych odczytań i dołącza nierozstrzygnięty sens do evidence. Runtime
może wtedy zrealizować tekst dla tego grafu, lecz kandydat i całe żądanie
pozostają `incomplete`; tekst nie staje się automatycznie wybranym wynikiem.
Pozostałe rodzaje niekompletnego źródła nie przechodzą przez ten wąski wyjątek,
a fraza bez żadnego zatwierdzonego ekwiwalentu nadal jest `unsupported`.

## Realizacja

`MeaningGraph → zatwierdzony lub spójny SemanticBinding → ValencyFrame → językowa ConstructionFamily/Variant → procedura NP/VP/PP → formy fleksyjne → tekst`.

Czas, negacja, liczba, rekcja, role i referenty nie mogą znikać po drodze tylko dlatego, że silnik wybrał łatwiejszy szablon. Zmiana szyku PL może dać różną powierzchnię przy tych samych rolach, ale EN SVO ma własne reguły.

## Odpowiedź w kolejnych etapach

`QuestionMeaning + jawne FactSet → Reasoner (ograniczone reguły) → ContentPlanner → ResponseMeaning → PL/EN ConstructionPlan → odpowiedź`.

Brak FactSet nie jest dowodem fałszu; świat nie jest zamkniętą tabelą. Nie planujemy pamięci ani uczenia teraz.

## CLI jako główny wskaźnik funkcjonalny

CLI udostępnia `golingua translate --from pl --to en --text "..." --runtime-config <path> --json`, typowane `question` na jawnych Meaning/FactSet oraz analizę i generację ograniczonych form przez `forms`. `question --language pl|en` może realizować dodatnią odpowiedź yes/no na istniejących trasach; nie parsuje zdań pytających w języku naturalnym. Właściciel może podać wejścia do ręcznego odbioru; wynik `ok` potwierdza wykonanie pipeline'u, nie poprawność językową.

**Nie ma skryptu sprawdzającego setki zdań, założeń procentowych ani porównania do starego JS.** Minimalne CI uruchamia formatowanie, vet, build i lekkie smoke CLI; nie dodajemy suite testów gramatycznych ani goldów. Ograniczenia runtime (maksymalna liczba kandydatów/węzłów) są cechą bezpieczeństwa, a nie celem generowania automatycznych testów.

## Jawna konfiguracja runtime

`translate` i `question` mogą przyjąć `--runtime-config <path>`. Wersjonowany plik JSON wskazuje istniejący aktywny dataset store i `dataset_id`; ścieżki względne liczone są od katalogu pliku konfiguracji. Store musi już zawierać aktywny snapshot; CLI nie importuje ani nie aktywuje danych. W skompilowanym CLI pakiety PL/EN powstają przez zaufaną fabrykę Go, więc `language_package_dirs` można pominąć.

```json
{
  "schema_version": 1,
  "dataset_store": "./store",
  "dataset_id": "golingua-authored"
}
```

Inne aplikacje mogą zamiast skompilowanej fabryki wybrać zewnętrzne katalogi pakietów; jeden runtime nie może użyć obu źródeł naraz.

CLI odrzuca pliki konfiguracyjne z nieznanymi lub powtórzonymi kluczami, dodatkowymi wartościami JSON i przekroczonym limitem rozmiaru. Pakiety są weryfikowane przez `internal/language`; kod providerów i adapterów musi być skompilowany w aplikacji. Interfejsy `language.Analyzer` i `language.Realizer` obsługują kierunkowe operacje pojedynczego providera w ramach zatwierdzonego bindingu. Interfejsy silnika `engine.SourceAnalyzer` i `engine.TargetRealizer` zachowują alternatywy znaczeń i łączą analizę zdań, składnię, kompozycję semantyczną, wybór leksykalizacji, planowanie i realizację. Oba poziomy mają odrębne kontrakty; `StageFactory` tworzy adaptery silnika osobno od providerów rejestrowanych w `language.Registry`. Loader podłącza je według języka po weryfikacji registry. Brak zatwierdzonych bindingów, danych lub adaptera pozostaje `unsupported`. Sama konfiguracja nie deklaruje capability i nie stanowi językowego odbioru.

`internal/runtime.SourcePipeline` łączy zaimplementowane etapy źródłowe: `BuildLattice → Parse → Compose → Disambiguate`. Wymaga skompilowanego `PhraseEvidenceBuilder`, który jawnie dostarcza kandydatów fraz i kontekst parsera dla lokalnej gramatyki języka. `SimplePhraseEvidenceBuilder` tworzy jednokenowe kandydaty z jawnej mapy `POS → kategorie` i zachowuje wszystkie znane odczyty oraz sensy; angielski wrapper przenosi określoność `the` do cech NP. MWE wymagają osobnego buildera. Opcjonalny `CompositionEvidenceBuilder` może dostarczyć jawne referencje i operatory dla konkretnego żądania. Adapter nie wyprowadza kategorii fraz, ról semantycznych ani tożsamości referentów z samych odczytów morfologicznych. Pipeline źródłowy dla silnika jest przekazywany przez `StageFactory` jako `engine.SourceAnalyzer`.

`internal/runtime.TargetPipeline` łączy `SelectTargets → BuildPlan → Execute → Linearize` i sprawdza końcowy wybór względem dokładnego, zatwierdzonego bindingu w zweryfikowanym `language.Registry`. Wymaga skompilowanego `TargetExecutionBuilder`, który podaje principal parts oraz jawne wartości cech dla slotów; adapter nie wyprowadza przypadku, zgody ani innych cech z ról semantycznych. Opcjonalny `TargetOperatorBuilder` dostarcza operatory wyłącznie dla konkretnego kandydata Meaning i jawnie oznacza kompletność pokrycia. Adapter może być przekazany jako `engine.TargetRealizer` w `runtime.Options.Realizers`. `SelectTargets.Result.Outcome` agreguje stan całej selekcji, natomiast każdy `TargetCandidate` zachowuje status i diagnostyki swojej własnej ścieżki. Dlatego kilka poprawnych tras może dać zbiorcze `ambiguous`, nie zmieniając statusu leksykalnych kandydatów ról ani poszczególnych tekstowych alternatyw.

`runtime.Options.LanguageFactory` kompiluje dokładne pakiety z aktywnego packa i wiąże je z zaufanymi providerami; `StageFactory` uruchamia się po weryfikacji `language.Registry`. `cmd/golingua/compiled.go` rejestruje oba elementy dla dokładnych szlaków `seeing`, `hearing`, `looking` i `receiving` w PL i EN. Pozostałe bindingi nie są deklarowane jako możliwości zdań. Powtórne przypisanie adaptera do tego samego języka jest błędem konfiguracji. Plik JSON nie może uruchamiać ani wybierać kodu providera.

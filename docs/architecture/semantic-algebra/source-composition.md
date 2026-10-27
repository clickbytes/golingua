# Składanie analizy źródłowej

`algebra analyze` jest jawną ścieżką analizy składniowej, która tworzy typowane fragmenty semantyczne. Przepływ zachowuje kolejne poziomy: lattice i morfologię, frazy, a następnie forest wszystkich alternatyw — nie wybiera jednego drzewa. Używa skonfigurowanego `runtime.SourcePipeline.AnalyzeSyntax`, a następnie `ComposeDSE`/`AnalyzeDSE`. CLI pobiera pipeline z aktywnego `EngineLoader` przez `SourceAnalyzerFor`; nie dodaje domyślnej translacji ani fallbacku do `CORECompose`.

```text
tekst → lattice/morfologia → frazy → forest wszystkich alternatyw
      → wiązanie ról/portów + jawny kontekst → kandydaci i standalone Document
```

## Jawny wybór i granice profilu

Polecenie wymaga `--language`, `--text`, `--runtime-config`, `--language-package`, `--source-plan`, `--context`, co najmniej jednego jawnego `--metadata` oraz `--json`. Powtarzalne `--source-limit name=value` przyjmuje `max_alternatives`, `max_work` i `max_output_bytes`; podanie limitu ogranicza wykonanie, a jego wyczerpanie nie jest brakiem interpretacji. Dane wejściowe wybierają dokładny sens, ramę, rodziny i dozwolone warianty; silnik sprawdza także zezwolenie parsera. Pusta lista `allowed_variant_ids` nie dodaje filtra wariantu wewnątrz niepustych rodzin. Niepusta lista jest dokładnym filtrem.

Pilot obejmuje skończone konstrukcje oznajmujące czasu teraźniejszego i przeczące z `widzieć/see` i `słyszeć/hear`, z podmiotem nauczycielem lub psem i dopełnieniem pies, oraz zamianę ról i dwa psy. Oddzielnie dodano cztery jawne warianty WH dla `see/widzieć`: pytanie o experiencer oraz stimulus w EN i PL. To ograniczone wejścia z przypiętym wariantem, typowanym gap i jawną referencją pozostałego argumentu; nie jest to twierdzenie o pełnym pokryciu WH, innych czasowników, odmianie ani modalnościach. Span luki pochodzi z dokładnego tokenu źródłowego i jest przenoszony do `role_sources[].source` oraz `role_sources[].wh_gap`. Composer kończy parserowy slot tylko wtedy, gdy jego slot/rola/span pasują do jawnego dowodu WH; konflikt referencji wywołującego dla tej samej roli daje wynik `incomplete`. Brak mapowania, nieznane słowo, nierozstrzygnięta referencja, nieobsługiwany operator i limity pozostają widoczne jako odmowa lub niepełny wynik.

Mapowania ról pochodzą z wybranego drzewa i jawnego planu. Identyfikacja uczestnika wymaga referencji kontekstu powiązanej z rzeczywistym kandydatem składniowym i potwierdzonej `ResolutionEvidence`. `role_references` to jawne założenie wywołującego używane jako fallback, nie dowód tożsamości. Dwa rzeczowniki tej samej klasy pozostają osobnymi ID. Zdarzenie jest jawnie przypisaną referencją `Event`; silnik nie tworzy jej przez zgadywanie.

Warunki cech mają ograniczony zakres: suma jawnych `feature_conditions[].feature_id` i `required_feature_ids`. Znane cechy w tym zakresie muszą mieć mapowanie albo pozostają resztą. Pozostała morfoskładnia jest śladem źródłowym, a nie semantycznym predykatem. Atrybuty `TemporalFrame` i `Multiplicity` są kopiowane z rzeczywistych atrybutów węzłów do referencji i kontekstu; ich sprzeczność powoduje odmowę. Cechy określoności EN są jawnymi warunkami kontekstu; plan PL ich nie zgaduje. Warunki i presupozycje pozostają nieasertywnymi danymi kontekstu, nie faktami.

`reference_predicate` jest osobnym, jednoargumentowym wiązaniem predykatu do referencji. Nie wymaga ramy ani sztucznej walencji, ale nadal wymaga jawnego typu, portu i autoryzowanej deklaracji. Bramka wykonania jest niezależna od tego, czy wpis istnieje w pakiecie.

## Wynik i kontrola

Wynik zawiera analizę źródłową, jawne mapowania, rozpoznane referencje, kandydatów, standalone dokumenty i diagnostykę. Wszystkie 12 przypadków oznajmujących/przeczących zwróciło `goal_specs: []`; ścieżka sprawdzania dowodu celu dla tych fragmentów **nie była uruchomiona (NOT RUN)**. Cztery oddzielne przykłady WH zwróciły `Ask` z jednym typowanym `Entity` answer i przechodziły niezależne `algebra check` jako `Question`; nie twierdzą nic o kompletności odpowiedzi ani o faktach. Odpowiedź nie ma restryktora `Person`, więc te przykłady nie kodują pełnej semantyki leksykalnej `who/whom/kto/kogo`. Ich parserowy forest zachowuje `incomplete` dla brakującego NP, a DSE uznaje tylko dokładne, jawne luki WH za obsłużone argumenty. Niezależne `algebra check` sprawdziło wyodrębnione dokumenty. Pełne zdanie, w tym zakres `Not`, zostało sprawdzone typowo, strukturalnie i względem kontekstu jako `type_scope_context_only`. Nie jest to dowód prawdy świata, kompletności interpretacji, faktów ani poprawności tłumaczenia.

Dla sześciu par PL/EN zachowano pełne referencje, warunki i atrybuty. Porównanie z pustym profilem reguł i niezależna weryfikacja certyfikatu zakończyły się `proved_equivalent`. Nie usuwano kontekstu ani atrybutów, by uzyskać zgodność. Te kontrole potwierdzają równość wyrażeń w ograniczonym profilu, a nie akceptację językową.

## Przykłady i obserwacje

Przykłady wejść znajdują się w [`data/semantic/examples/source-composition/`](../../../data/semantic/examples/source-composition/). Literalne wyniki końcowego uruchomienia i ich wejścia są zapisane w [evidence #395](../../work/dse/evidence/07-02/README.md). Ocena agenta dotyczy wyłącznie tego profilu. Akceptacja językowa właściciela: **NOT EVALUATED**.

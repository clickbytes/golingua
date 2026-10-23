# Mapa ekranów F1–F5

To jest kontrakt interfejsu G11. Ekrany opisują aktualne API, nie obiecują szerszej obsługi języka. TUI nie jest jeszcze zaimplementowany.

| Ekran | Cel | Źródło danych | Widoczne ograniczenia |
| --- | --- | --- | --- |
| F1 Translate | Uruchomić jedną translację PL↔EN i przejrzeć wynik | Wspólna usługa wywołująca `engine.Engine.Translate` | Zawsze pokaż status. Zachowaj wszystkie rzeczywiste kandydatury, ID, założenia i diagnostyki. `ambiguous` nie wybiera tekstu automatycznie. |
| F2 Forms | Analizować formę albo wygenerować formę leksemu | Wspólna usługa korzystająca z istniejących `pl.Load()` / `en.Load()` i `Analyze` / `Generate` | Tryb Analyze XOR Generate. Cechy zależą od języka i leksemu. Nieznana komórka nie jest zgadywana. |
| F3 Concepts | Znaleźć koncept i zobaczyć jego rzeczywisty rekord | Katalog i skompilowany `core.ConceptGraph` | Filtr po ConceptID; definicja/relacje tylko po prawdziwym `Lookup`. Pokrycie leksykalne pochodzi z `catalog.Inspect`. Bez edycji danych. |
| F4 Languages | Obejrzeć PL/EN i ich stan | `catalog.Inspect` oraz informacja o wybranym runtime | Liczniki katalogu są oddzielone od gotowości runtime'u i zdolności translacji. Bez procentowych ocen jakości i instalatora języków. |
| F5 Settings | Zmienić preferencje widoku i zarządzać snapshotem | Preferencje TUI oraz istniejące preview/apply stagingu | Poziom Simple/Normal/Technical zmienia wyłącznie prezentację. `Preview` nie aktywuje snapshotu. `Apply` wymaga osobnego potwierdzenia. |

## F1 — Translate

Pola: język źródłowy, język docelowy, tekst i akcja Translate. Widok wyniku pokazuje status silnika, wyjściowy tekst gdy istnieje, wszystkie kandydatury i ich statusy. Wybranie podglądu kandydata używa jego stabilnego ID i nie zmienia statusu wyniku ani kandydata. Założenia określoności lub cech są widoczne przy właściwej kandydaturze.

Gdy runtime nie jest wybrany, ekran informuje o jego braku. Nie udaje, że katalog języka oznacza gotowość do tłumaczenia.

## F2 — Forms

Użytkownik wybiera dokładnie jeden tryb:

- **Analyze:** język i forma powierzchniowa; wyświetl wszystkie zwrócone odczytania.
- **Generate:** język, istniejący LexemeID i wybrane jawne cechy; wyświetl wszystkie wygenerowane formy.

Lista cech pochodzi z aktualnego profilu/leksemu. Nie ma swobodnego formularza, który akceptuje nieznane cechy.

## F3 — Concepts

Filtr przyjmuje stabilny ConceptID lub tekstową frazę do filtrowania rzeczywistego katalogu. Szczegóły pochodzą z `ConceptGraph.Lookup`; pokrycie PL/EN z bieżącego inwentarza. Ekran odróżnia brak rekordu od braku zatwierdzonego bindingu i nie sugeruje tłumaczenia na podstawie samej nazwy.

## F4 — Languages

Pokazuje PL i EN, liczbę leksemów/sensów/bindingów oraz liczbę konceptów zatwierdzonych/pending z `catalog --summary --json`. Obok pokazuje, czy wybrano aktywny runtime. Nie przelicza tych wartości na „jakość” ani „procent tłumaczenia”.

## F5 — Settings

Sekcje: poziom wyświetlania, domyślny kierunek, runtime config i katalog snapshotu. `Save` zapisuje preferencje interfejsu. `Preview snapshot` wyświetla kandydaturę i różnicę bez aktywacji. `Apply snapshot` wymaga dodatkowego pytania z nazwą katalogu i jawnym potwierdzeniem. Ustawienia nie zmieniają znaczeń, bindingów, konstrukcji ani reguł morfologii.

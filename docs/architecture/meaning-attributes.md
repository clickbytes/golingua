# Czas odniesienia i wielość referenta

Kontrakt koordynatora dla [LIGHT-F1 #325](https://github.com/clickbytes/golingua/issues/325), zatwierdzony do implementacji 2026-10-24. Uzupełnia istniejące Meaning v1 bez migracji lub nowej wersji produktu. Gramatyka konkretnego języka pozostaje poza I.

## Pola znaczenia

`Node.TemporalFrame` (`temporal_frame`) występuje wyłącznie na event/state. Wartości `past` i `present` oznaczają przeszłą lub teraźniejszą ramę odniesienia wypowiedzi. Teraźniejsza rama dopuszcza odczyt zwyczajowy/ogólny i nie twierdzi, że pojedyncze zdarzenie trwa w chwili mówienia. Przeszłość nie oznacza dokonania ani zakończenia. Brak pola to brak rozstrzygnięcia; nie jest domyślną teraźniejszością. W tym zakresie analizujemy zwykłe, już licencjonowane odczyty obecnych tras; historyczny present i zaplanowana przyszłość nie otrzymują automatycznej interpretacji.

`Node.Multiplicity` (`multiplicity`) występuje wyłącznie na referent. `one` oznacza pojedynczy desygnat, `many` wielość bez konkretnej liczby. Brak oznacza nierozstrzygnięcie. Nie ustala tożsamości, płci, istnienia ani prawdziwości zdania; nadal działa we właściwym zakresie negacji. Kwantyfikatory i ich dokładne liczby pozostają osobnymi operatorami.

Jawne null, puste lub nieznane wartości nowych pól w JSON są błędem. Wartość zerowa w Go oznacza pominięte pole. Pola na niewłaściwym rodzaju węzła są błędem. Walidacja CORE, canonical JSON/hash, porównanie, klonowanie, sygnatury kandydatów i projekcja zachowują pola. Projekcja statement zachowuje ramę predykatu i opisy bezpośrednich referentów; nie zamienia ich w trwałe fakty.

## Własność mapowania

Adapter języka dostarcza jawne, związane z CandidateID wskazówki `semantic.NodeAttributesEvidence`; wspólny compositor przenosi je na właściwe węzły, nie czyta nazw końcówek lub cech PL/EN. Wskazówka jest ograniczona do istniejącego kandydata i zgodnego rodzaju; sprzeczne wskazówki nie są arbitralnie wybierane. Współodniesienie wymaga zgodnych rozstrzygniętych atrybutów. Nie łączymy referentów po ConceptID.

Dla obecnych finite past/present oraz licencjonowanych konstrukcji z kopułą źródło jawnie ustala ramę. Odczyty mnogiego EN `read` zachowują osobne ramy present/past. Liczba rzeczownika daje wielość tylko dla jawnie zadeklarowanego `number_behavior=count`; `mass` i `plural_only` nie pozwalają wnioskować one/many z morfologii. Ta klasyfikacja jest danymi języka, nie nową regułą ontologii. Zbiór count obejmuje istniejące rzeczowniki potrzebne w dotychczasowych trasach i odbiorze; inne pozostają nierozstrzygnięte do przeglądu.

Realizacja odczytuje ramę i wielość po NodeID z I. Evidence może pozostać dowodem pochodzenia i nieobsłużonych tu wyborów (np. określoności), ale nie zastępuje brakującej ramy/wielości ani nie nadpisuje obecnej. Usunięcie Evidence nie zmienia znanej ramy/wielości. Sprzeczne dane o tych atrybutach dają diagnostykę, nigdy cichy wybór. Brak I pozwala tylko na jawnie opisane założenie/brak, jak w istniejącym kontrakcie. Formę gramatyczną dla mass/plural_only ustala lokalna leksyka. Zgodę czasownika wyprowadza się z realizacji podmiotu; brak cechy nie zgaduje płci referenta.

Obecne trasy mass/plural_only nie wyrażają liczenia porcji ani kwot. Jawne one/many na takim referencie wymaga odmowy realizacji zamiast zignorowania pola. Kontrola dowodów liczby czasownika dotyczy podmiotu wyznaczonego przez ramę, nigdy dowolnego dopełnienia.

`reason.Plan` nie umie jeszcze mapować względnej ramy czasu na faktowe `At` ani odpowiadać o grupach. Znane temporal_frame i many na argumencie pytania dają odpowiednio `UNSUPPORTED_QUESTION_TEMPORAL_FRAME` i `UNSUPPORTED_QUESTION_MULTIPLICITY`. One jest zgodne z pojedynczym request ReferentID. To ochrona przed zgubieniem znaczenia, bez nowego wnioskowania temporalnego lub grupowego.

Nie rozszerzamy tras na nowe czasy, osoby, domeny A1 lub przyszłą pamięć. Nie przenosimy do I case, rodzajów fleksyjnych ani angielskich rodzajników w tym issue. Akceptacja właściciela pozostaje NOT EVALUATED.

## Odbiór

Istniejące Seeing/Identification/Negation dostają jawne atrybuty wynikające z zamkniętych znaczeń, nie z aktualnego outputu. Dodajemy ograniczone lokalne kontrole `MeaningAttributes`: odrębne grafy czasu i wielości, przygotowane I→PL/EN bez parsera drugiego języka, zachowana homografia read, negacja i zamiana ról. Przypadki masowe/plural_only nie otrzymują fałszywej liczebności. JSON i porównanie wykrywają zmianę/utratę pola. Rzeczywiste CLI/TUI oraz techniczne CI PR/main są obowiązkowe. Pełna macierz nadal ujawnia braki pozostałych domen.

# GoLingua TUI

Interfejs terminalowy używa tych samych typowanych operacji
`internal/app.Service` co CLI. Startuje przy bezargumentowym uruchomieniu,
gdy stdin i stdout są TTY, a `TERM` nie ma wartości `dumb`. Przy przekierowaniu
lub braku TTY CLI zachowuje tryb nieinteraktywny i nie emituje ekranu TUI ANSI.

## Ekrany i kontrolki

| Ekran | Kolejność fokusu i działanie |
| --- | --- |
| F1 Translate | `prompt`, `direction`, `candidates`, `preview`, `result`, `tabs`. Enter na promptcie tłumaczy; Alt+D zmienia kierunek. Up/Down wybiera rzeczywiste kandydatury; Enter otwiera szczegóły. Preview jawnie wywołuje Translate z wybranym zwróconym PreviewID. |
| F2 Forms | Analyze: `prompt`, `language`, `mode`, `run`, `result`, `tabs`. Generate wstawia `features` między `mode` i `run`. Enter wybiera PL/EN i Analyze/Generate; żądanie zawiera tylko wpisaną formę, LexemeID i pary cech. Simple pokazuje skrót; Enter na result rozwija go. |
| F3 Concepts | `prompt`, `list`, `result`, `tabs`. Lokalny filtr dopasowuje dosłowny, uwzględniający wielkość liter podciąg wyłącznie w ConceptID i zachowuje kolejność katalogu. Up/Down wybiera; Enter na list lub result wywołuje LookupConcept dla wybranego ID. |
| F4 Languages | `refresh`, `result`, `tabs`. Wejście przez F-key, taby lub paletę odświeża inwentarz katalogu i walidację runtime'u. Enter na Refresh ponawia oba odczyty. Inwentarz pozostaje widoczny, gdy walidacja runtime'u nie istnieje albo jest błędna. |
| F5 Settings | `prompt`, `validate`, `level`, `default_direction`, `store`, `snapshot_preview`, `snapshot_apply`, `save`, `cancel`, `result`, `tabs`. Runtime i preferencje są szkicami do udanego Save. Validate sprawdza szkic runtime'u bez jego wyboru. Cancel lub Esc resetuje szkice. |

Na F5 Snapshot Preview używa jawnej ścieżki szkicu store. Preview nie aktywuje
snapshotu, choć otwarcie store może utworzyć katalogi. Apply jest dostępne po
udanym podglądzie dla bieżącej ścieżki. Dostępny z klawiatury dialog pokazuje
ścieżkę i digest, domyślnie wybiera Cancel. Apply ponawia preview i przekazuje
potwierdzony digest do usługi; zmieniony digest wymaga nowego potwierdzenia.
Zmiana ścieżki store unieważnia preview. Apply nie wybiera runtime'u ani nie
zapisuje preferencji. Ctrl+C/Ctrl+D czeka na zakończenie rozpoczętego Apply.

Paleta nawigacji rejestruje pięć tras: `/translate`, `/forms`, `/concepts`,
`/languages` i `/settings`. Ctrl+P otwiera ją z każdego ekranu. Początkowy
ukośnik w promptcie F1–F3 także ją otwiera; ukośnik w polu ścieżki F5 jest
literalny. Paleta służy wyłącznie do nawigacji.

## Preferencje i prezentacja

Ustawienia zachowują schemat wersji 1. Starsze pliki bez `display_level`,
`default_from` lub `snapshot_store` otrzymują domyślne wartości w pamięci
(Normal, EN→PL i pustą ścieżkę store) i nie są przepisywane przy wczytaniu.
Save waliduje przechwycony szkic i atomowo zapisuje plik z uprawnieniami 0600;
preferencje stosuje tylko udany zapis.

Simple pokazuje zwięzłe podsumowanie. Normal pokazuje listy, ID, cechy i
założenia. Technical dodaje zwróconą odpowiedź jako JSON. Szczegóły kandydata
pozostają dostępne z klawiatury na każdym poziomie; Enter w Simple rozwija
odczytania/formy F2. Zmiana poziomu prezentacji nie wysyła ponownie żądania
usługi ani nie zmienia selekcji.

## Terminal i dokumenty projektu

Przy 40×12 lub większym oknie shell zawija i przycina treść do wymiarów; poniżej
40×12 pokazuje minimalny wymagany rozmiar. Resize zachowuje ekran, filtr,
selekcję, szkice i przewinięcie. F1–F5 przełączają ekrany; Tab/Shift+Tab
zmienia fokus, a Left/Right na pasku zmienia ekran. Fokus wyniku obsługuje
Up/Down i Page Up/Down. Ctrl+U czyści aktywny edytor, `?` pokazuje pomoc, Esc
zamyka pomoc/podwidok, a Ctrl+C/Ctrl+D kończy działanie i przywraca terminal.

[Mapa klawiszy](keymap.md), [bindingi usług](bindings.md),
[model stanu](state-model.md) i [routing](command-contract.md) opisują obecne
zachowanie. [Makieta globalnej ramy](mockups/global-shell.md) i
[szczegółów kandydatów F1](mockups/translate-candidates.md) to historyczne
projekty, nie bieżący output TUI. Bubble Tea v2 dostarcza zdarzenia i
renderowanie; semantyka usług GoLingua pozostaje w `internal/app`.

## Domknięcie G11

V1–V7 są scalone do main. Dokumentacja V8 i rzeczywiste obserwacje F1–F5,
TTY/non-TTY oraz resize są opisane w [issue #307](https://github.com/clickbytes/golingua/issues/307).
Wyniki końcowego CI i status zamknięcia są zapisane w tym issue oraz
[epiku G11](https://github.com/clickbytes/golingua/issues/254).
Ocena techniczna agenta i akceptacja językowa właściciela pozostają odrębne.

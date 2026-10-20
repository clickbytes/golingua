# Bieżące CLI

CLI w obecnym bootstrapie służy wyłącznie do sprawdzenia wersji, pomocy i deklarowanych capabilities. Nie ma parsera tekstu, translacji, generowania form ani aktywnych danych PL/EN.

## Działające polecenia

Uruchom z katalogu repozytorium:

```sh
go run ./cmd/golingua --version
go run ./cmd/golingua --help
go run ./cmd/golingua capabilities --json
```

`--version` zwraca wersję bootstrapu. `--help` pokazuje dostępne komendy. `capabilities --json` obecnie zwraca `status: "unsupported"`, puste tablice `languages` i `features`, oraz `ready: false`; sama komenda kończy się kodem 0.

## Translacja nie jest zaimplementowana

`translate` obecnie **nie analizuje argumentów tekstu** i nie może być użyte do ręcznej oceny PL/EN. Każde wywołanie tej komendy zwraca wynik `unsupported`, diagnostykę `LANGUAGE_RUNTIME_UNAVAILABLE` i exit code 3. Na przykład aktualne:

```sh
go run ./cmd/golingua translate --source pl --target en 'tekst właściciela'
```

kończy się `unsupported`; nie jest to wynik próby tłumaczenia. Pole `message` wyjaśnia, że translacja zdań nie została zaimplementowana w bootstrapie. Nie ma komendy `forms`.

Nieznana lub niepoprawna komenda zwraca `invalid` i exit code 2. Ogólne błędy procesu/serializacji mają osobne kody; `unsupported`, `invalid` i awaria programu nie są jedną kategorią.

## Przyszłe funkcje

Gdy komenda analizy, tłumaczenia lub generowania form rzeczywiście zostanie zaimplementowana, dopisz tu jej dokładne argumenty, kierunki, wyjście i kody statusu na podstawie obserwowalnego zachowania. Do tego czasu nie traktuj przykładowych przyszłych komend jako działających. Ręczna procedura oceny jest opisana w [manual review](manual-review.md).

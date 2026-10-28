# DSE — status odbioru #400

**Status zadania:** `IN_PROGRESS` na bazie zapisanej w [issue map](issue-map.json). Ten dokument podsumowuje istniejące obserwacje; nie oznacza zamknięcia #400.

**Ocena techniczna agenta:** przedstawione niżej przepływy działają w zapisanym, ograniczonym zakresie. **Akceptacja językowa właściciela: NOT EVALUATED.** Nie przypisuję właścicielowi oceny naturalności zdań, adekwatności formalizacji ani poprawności znaczeń.

## Macierz obserwacji

| Zakres | Obserwacja | Status odbioru |
| --- | --- | --- |
| Neutralna synteza 2-krokowa z unary restriction | Construct, niezależne `verify-goal` i ewaluacja jawnych faktów w evidence #09-02. Ewaluacja dała dwie wspierane odpowiedzi w wybranym kontekście. | Zaobserwowano `ok` dla tego wejścia; nie jest to dowód prawdy świata ani ogólnej kompletności. |
| Niezależność od danych domenowych | Dwie neutralne domeny, ścieżki 2/3, restrykcje i warianty po przemianowaniu symboli, ról, sortów, portów, referentów i faktów w evidence #09-01. | Osiem przypadków `ok`, niezależne sprawdzenie celów i ewaluacja; ograniczony audyt danych i wskazanych modułów. |
| Eksport z cache i bez cache | Zwykły eksport i `--no-cache` dla pakietów #09-01. | Zgodny `meaning_fingerprint` w zapisanych parach. |
| Kontrole błędów i limitu | Nieznana rola, usunięta restrykcja, podmieniony certyfikat, alternatywy długości ścieżki i niski budżet. | Odmowy, wielość kandydatów oraz `resource_limit`/`incomplete` zachowane w evidence. |
| Niezależna realizacja checked IR | Ten sam jawny anonimowy wyraz zrealizowany na osobno wybranych pakietach EN i PL w evidence #09-02. | Oba wywołania `ok`, z weryfikacją coverage. Nie interpretuje czasu gramatycznego jako czasu zdarzenia. |
| Tłumaczenie WH PL→EN | “Kto widzi psa?” → kandydat “Who sees the dog?” | `incomplete`; kandydatura niesie jawne założenia zgodności podmiotu i określoności. |
| Tłumaczenie WH EN→PL | “Who sees the dog?” → kandydat “Kto widzi psa?” | `incomplete`; pozostają residuals realizacji. |
| Akceptacja językowa | Właściciel nie ocenił tych propozycji w przytoczonych artefaktach. | `NOT EVALUATED`. |

## Zakres wniosku

Odbiór techniczny obejmuje tylko rzeczywiście zapisane wejścia, wybraną metadaną, fakty, budżety i CLI. `ok` konstruktora i niezależnego weryfikatora oznacza spełnienie strukturalnych ograniczeń celu oraz odtworzenie konstrukcji; nie stwierdza, że relacja zachodzi w świecie. Ewaluator obsługuje jawny profil dodatnich faktów: nieobecność dowodu to `unknown`, a nie fałsz. Reprezentowalne `Not` i kwantyfikatory nie są przez ten profil ewaluowane.

Realizacja IR→język nie zatwierdza jakości propozycji językowej ani czasu zdarzenia. Oba kierunki WH pozostają `incomplete`; ich tekstowych kandydatur nie uznaję za w pełni odebrane tłumaczenia. Należy utrzymać te statusy, założenia i residuals jako część wyniku.

## Artefakty

Podstawą są [wyniki i granice](results.md), [evidence #09-01](evidence/09-01/README.md) wraz z [manifestem komend](evidence/09-01/commands.jsonl) oraz pliki konstrukcji, weryfikacji, ewaluacji, realizacji i tłumaczeń w [evidence #09-02](evidence/09-02/commands.jsonl). Interfejs CLI i deklarowane granice profilu są zapisane w wyjściu `golingua --help` użytej w #09-02 oraz w [kontrakcie syntezy](../../architecture/semantic-algebra/synthesis.md), [kontrakcie faktów](../../architecture/semantic-algebra/facts.md) i [manifeście capabilities](evidence/09-02/capabilities.json).

Techniczny status agenta i akceptacja właściciela są odrębne. Nie należy oznaczać całego DSE jako zaakceptowanego językowo, dopóki właściciel nie dokona tej oceny.

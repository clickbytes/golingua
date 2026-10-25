# JOB BRIEF — A1-M2 #338

Base: `9bf866bc620ec1ad4ae994876f57bc47885aac71`. Wykonawca: Luna Low, jedyny writer infrastruktury. Wejścia: aktualny inwentarz, kontrakty AllowedTexts, pełny bazowy log, audyt właściciela.

Dozwolone: `tests/matrix/{cases,checks}.go`, `tests/translation/{cases,cases_test,lookup}.go`, reporter Python, `tests/README.md`; istniejące dane i lokalne wywołania realizacji PL/EN tylko dla zachowania dotychczas wymaganych wariantów. Ograniczone kontrole mechanizmu w `tests/`.

Rezultat: inwentarz bez limitu liczbowego, osobne AllowedTexts/RequiredTexts i warunki wariantów, raport także dodatkowych kontroli z jednoznaczną nazwą pakietu. Deklaracje pokrycia odróżnione od wykonania.

Zakazy: zmiany silnika/gramatyki, danych produktu, go-ci, masowe zdania, ukrywanie braków, akceptacja właściciela przez agenta, GitHub/branch/commit wykonawcy. Jeden przebieg i jedna korekta po review. Przekazanie: lista ścieżek, diff, rzeczywiste komendy/wyniki, ograniczenia. Koordynator sprawdza i integruje.

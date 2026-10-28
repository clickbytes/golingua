# Dane obserwacyjne #399

`commands.jsonl` zawiera literalne argumenty, exit code, stderr i plik stdout każdej próby. Polecenia uruchomiono jako odrębne procesy `/tmp/golingua-dse-399`, zbudowane przed utworzeniem pakietów z bazy `3643bdcd69298f85fff2826431502743ad4c3024`; SHA-256 przed i po: `2c792432f78f883cf5e546b4096b46e54b1b1f373bbc989342120b7a65feb4c5`. Pliki `*-candidate.json` i `*-certificate.json` wydzielono z rzeczywistej odpowiedzi konstruktora dla niezależnego CLI `verify-goal`; `*-export-request.json` analogicznie wydzielono do eksportu. Wejścia negatywne znajdują się obok wyników. Binarne ścieżki `/tmp` są lokalne; do odtworzenia zbuduj binarkę w tej lokalizacji z podanej bazy.

Porównanie odpowiedzi po przemianowaniu wykorzystuje `referent_ids` z właściwego `id-map.json`. Eksport `--no-cache` i zwykły eksport mają identyczny `meaning_fingerprint`; ich struktury węzłów mogą się różnić. Opis granic i wyników znajduje się w `../../independence-audit.md`.

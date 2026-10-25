# A1-M4 #340 — zakres wykonania

Base: `78bba8dbf9cec5d09981f0ae1d1c7d2026d5d393` (main CI dla poprzedniego source PR zielone). Branch: `a1-m4-corpus`.

Zakres obejmuje dokładnie 16 wskazanych par `S003, S010, S016, S023, S028, S034, S040, S043, S051, S055, S063, S074, S076, S081, S086, S099` z `data/sentences/100-pl-en.tsv`, czyli 32 rekordy kierunkowe A1.1. Dane źródłowe: `/tmp/golingua-matrix-expansion/m4-selected.json`.

Grafy I są niezależnie zapisanymi zdarzeniami, rolami, konceptami, czasem i liczbą pojedynczą. Nie pochodzą z parsera ani bieżącego wyjścia runtime. Czas przeszły dotyczy `receiving`, a teraźniejszy `seeing`, `hearing`, `looking`. Każdy rekord zachowuje ID źródłowe, `draft_unreviewed` i `OwnerAcceptance: NOT_EVALUATED`. Dla realizacji PL→EN kontrakt zawiera wszystkie cztery kombinacje określoności dwóch rzeczowników oraz jawne założenia rodzajników.

Zdania z `dziecko/child` w S023, S040, S074 i S099 mają dwa wymagane, niezależne odczytania: `concept:child` oraz `concept:child_of_person`. Grafy alternatywne zachowują to samo zdarzenie i pozostałe role. Teksty realizacji są przypisane do właściwego grafu, a wymaganie każdego odczytania jest sprawdzane osobno.

Brak zmian gramatyki i implementacji produktu. Rozbieżności wyników pozostają widoczne jako FAIL/GAP.

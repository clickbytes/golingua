# JOB BRIEF — A1-M3 #339

Base: `cb5d52f5ad45a5dccd9fd19b11f3405b19e940cd`, branch `a1-m3-contracts`, po merge M2 i zielonym CI main. Owner: koordynator Codex. Jeden Luna Low writer; koordynator review/verifies.

Cel: zamknąć niezależne wymagania dla istniejących 12 kierunków `LQ`, `PO`, `Q` bez dostrajania expected do outputu. Każdy rekord ma niepuste MeaningID, wymagane predykaty/asercje, role uczestników, osobę/referencję/liczebność/akt pytania zależnie od zjawiska, propozycje, dozwolone i wymagane powierzchnie, provenance i `OwnerAcceptance: NOT_EVALUATED`.

Dozwolone: `tests/interlingua/pending_domains.testdomain.go` i wąskie kontrole semantyczne; `tests/matrix/{cases,checks}.go`; `tests/languages/{en,pl}/A1/{location_quantity,possession,questions}.testdomain.go`; `scripts/report-language-matrix.py` wyłącznie aby `GAP_I` był statusem głównym zachowując szczegóły GAP_ANALYZE/REALIZE; `docs/work/a1-matrix/339-{brief,results}.md` i status w `README.md`.

Kontrakty: on-support nie jest równoznaczne z `located_in`; wymagane wsparcie/kontakt z rolami supported/support nie ma obecnie reprezentacji Meaning. Posiadanie tymczasowe jest KindRelation, a neutralny Meaning nie wiąże `I` z mówiącym ani `she` z referentem żeńskim. WH-location wymaga nieobsługiwanej relacji `located_in` (occupant/location); niewypełniane miejsce nie może być zgadywane. `There are two cats` ma niezależne I: dokładny kwantyfikator 2 nad referentem kotów o multiplicity many; domena wypowiedzi pozostaje nieokreślona. Yes/no klasyfikacja kota jest reprezentowana przez Meaning identyfikacji owinięty `Question` i `Discourse(ActQuestion)`, z kontekstem proximal lub unresolved.

W runnerach błędy reprezentacji są `GAP_I`, bez fikcyjnych grafów. Analysis i zintegrowane tłumaczenie wykonują produkcyjny adapter/pipeline i logują rezultat; realizacja bez samodzielnego grafu jawnie zapisuje `independent_realization_executed:false`. Unsupported nie jest PASS. Bez zmian engine/CORE/gramatyki, bez `t.Skip`, bez zmiany CI.

Przekazanie: pliki, dokładne komendy i wyniki, przyczyny GAP, ograniczenia. Koordynator wykonuje build, CLI, vet, PR i CI.

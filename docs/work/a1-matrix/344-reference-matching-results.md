# M8 reference matching observations

- `go test ./tests/matrix -run '^(TestContextMatches|TestMatchAllowedCandidate)' -count=1`: PASS (8 tests). Covers alpha-renamed nodes, reordered roots/nodes/mentions, wrong-role anchors with an unchanged anchor set, missing/extra/wrong anchors, resolved identity, and selecting the correctly anchored secondary variant when graphs are identical, selecting integrations by source candidate ID, rejecting wrong graphs/unknown or empty source IDs, and propagating invalid fixture and comparison-budget errors.
- `go test ./tests/interlingua -run '^TestIdentification' -count=1`: PASS (2 tests), retaining strict one-referent identification behavior.
- `go test ./tests/translation -run '^TestInventory$' -count=1`: PASS.
- Existing integration check `go test ./tests/translation/en-pl ./tests/translation/pl-en -run '^TestIdentification$' -count=1`: PASS (6 tests).
- `gofmt` and `git diff --check`: PASS.
- Full matrix, CLI, vet/build, and CI: NOT RUN (coordinator-owned).

No new time, ambiguity, or imperative sentence records were added. The generalized matcher is ready for authored fixtures whose loaders attach their graph and request context through `IdentificationMeaning`. Owner acceptance remains `NOT_EVALUATED`.

Coordinator full run source65602ee+working-tree:536records/1608checks;255PASS,66FAIL,442GAP_ANALYZE,221GAP_REALIZE,624GAPI,0SPEC_PENDING/INFRA/NOT RUN; exit1. One integration classification changed: A1.5-ID-553-EN-PL now FAIL because the output cannot bind to a validated OK source candidate; its source analysis already FAILed. Other record statuses are unchanged. Corrected independent local node IDs, same-graph/different-context variant selection and empty/unknown/duplicate/invalid-status source-ID guards after read-only review. Focused proofs and vet/build/help/version/diff passed. No new language inputs; existing CLI evidence remains applicable because product code and datasets are unchanged. Owner acceptance NOT_EVALUATED.

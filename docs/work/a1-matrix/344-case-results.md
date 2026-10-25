# #344 CASE42 — results

Implementation records 42 pairs / 84 directions in group `case_adjective_order`. Independent fixture IDs are `a1.11.case.NNN.en|pl`; child uses the base ID for age-child and `.offspring` for child-of-person. Distinct referent mentions are separate nodes. Black attributes are untimed state roots attached to the matching referent. Row 6 preserves source-specific door count: PL source empty, EN source one.

Explicit full I gaps: row 2 on-support (CORE lacks exact relation assertion); row 14 open-state (opening is an event, not this state); row 37 using_object predicate; row 38 cutting instrument role; row 39 writing with instrument without adding recipient/durable document; row 40 reciprocal conversation; row 41 accompaniment. No partial graph is returned for these meanings. OVS, dative-fronting, and accusative-fronting rows use legal event graphs. OwnerAcceptance: `NOT_EVALUATED`.

Observed focused checks and matrix run outcomes are recorded in the handoff; language realizations are expectations, not observed correctness approvals.

## Observed controls

- `go test ./tests/interlingua -run '^TestCaseOrderExpansion' -count=1`: PASS. Independent licensed graphs validate against CORE and references; row 6 verifies PL door multiplicity empty and EN one; black-state bindings and child/offspring graphs are checked.
- `go test ./tests/translation -run '^TestInventory$' -count=1`: PASS.
- `go test ./tests/languages/en/A1.11 ./tests/languages/pl/A1.11 ./tests/translation/en-pl ./tests/translation/pl-en -run 'TestCaseOrder|TestOtherDomains' -count=1 -v`: EXIT 1, expected language-contract failures. CASE42 showed the declared seven `GAP_I` requirements for each source/target direction and no CASE42 `SPEC_PENDING` or `INFRA` diagnostics. `TestOtherDomains` also executes previously registered domains, so this exit is not an isolated CASE42 result.
- `gofmt` on changed Go files and `git diff --check`: PASS. Full matrix, CLI, vet, and build: NOT RUN (outside this component's requested scope).

Final declared inventory after this addition: 424 pairs / 848 directions; `case_adjective_order` has 45 pairs. Language-owner acceptance remains `NOT_EVALUATED`.

Coordinator final run source351b1b1+working-tree:848records/2544checks;262PASS,137FAIL,806GAP_ANALYZE,379GAP_REALIZE,960GAP_I;0SPEC_PENDING/INFRA/NOT RUN, exit1. All84 CLI inputs RUN:70unsupported/exit3,12incomplete/exit7,2ambiguous/exit6, [responses](evidence/344-case-cli.json). Independent bounded review found no discrepancy in the84 directions, role/attribute bindings, source-specific door count or seven claimed CORE gaps. Vet/build/help/version/diff passed. New PASS checks are observed existing capabilities, not a claim that product code changed or the whole group works. Owner NOT_EVALUATED.

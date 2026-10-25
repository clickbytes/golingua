# M8-NEG45 implementation observations

Fixture and inventory checks after the bounded correction:

- `go test ./tests/interlingua -run '^TestNegationExpansion|^TestNoCatSentenceNeedsParticipantQuantification' -count=1`: PASS (3 focused tests).
- `go test ./tests/translation -run '^(TestInventory|TestA18NegationLinksHaveMatchingMeaningContracts)$' -count=1`: PASS (2 focused tests).
- `git diff --check`: PASS.

Focused execution of the new language and integration records:

- `go test ./tests/languages/en/A1.8 ./tests/languages/pl/A1.8 -count=1 -v`: exit 1 from observed language `FAIL`/`GAP_I`; zero `SPEC_PENDING`, zero `INFRA`.
- `go test ./tests/translation/en-pl ./tests/translation/pl-en -run 'TestOtherDomains/A1\.8-NEG-' -count=1 -v`: exit 1 from observed language `FAIL`/`GAP_I`; zero `SPEC_PENDING`, zero `INFRA`.
- Raw logs: `/tmp/neg-native-v.log`, `/tmp/neg-integrations-v.log`.
- Full matrix, CLI, vet/build, and CI: NOT RUN (coordinator-owned).

The runtime does not yet meet many authored negation and realization contracts. The nine declared representation gaps remain visible as `GAP_I`; observed failures were not used to revise expected semantics. Owner acceptance remains `NOT_EVALUATED`.

Coordinator review corrected all mismatched MeaningIDs before accepting this batch; the new inventory validates nonempty shared Meaning IDs and identical source/target semantic requirements. No-cat participant quantification is GAP_I, never a quantified-event substitute. Final full source965743b+working-tree:536records/1608checks, exit1;255PASS,65FAIL,442GAP_ANALYZE,222GAP_REALIZE,624GAP_I,0SPEC_PENDING/INFRA/NOT RUN. All90 CLI calls RUN:84unsupported/exit3,3incomplete/exit7,1ambiguous/exit6,2unknown/exit5; none counted as a boundary PASS. [Full CLI responses](evidence/344-negation-cli.json). Vet/build/help/version/diff passed. Owner acceptance NOT_EVALUATED.

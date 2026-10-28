# Bounded source coordination — #398, 2026-10-28

Source implementation: `9f25acb` on the existing #422 branch. This completes the bounded source subtask; the whole issue remains **PARTIAL**. Agent technical assessment and owner linguistic acceptance are separate: the latter remains **NOT EVALUATED**.

## Observed production flow

The final-prefixed `algebra analyze` commands return `ok`, `complete=true`, one checked `And` and two clause provenance records for:

- EN: `The teacher sees the dog, and the dog sees the teacher.`
- PL: `Nauczyciel widzi psa i pies widzi nauczyciela.`

E1 has experiencer A and stimulus B; E2 has experiencer B and stimulus A. The contexts explicitly bind every used source candidate through request-local mention/resolution evidence. Repeated nouns do not infer identity. A is the teacher, B the dog; both Event references are explicitly distinct. The public source trace now exposes each clause forest, exact plan/frame/sense/candidate IDs, global ranges, language connector and evidence. The legacy whole-input forest remains `unknown`, so the legacy source route cannot treat two clause roots as alternatives.

The syntax-only probe preserves the existing simple sentences and accepts two complete positive clause parses. Empty left/right clause, an extra connector, WH and negative clauses produce no coordination. The Polish connector for this new bounded proposal is `i`; this observation does not assert support for the earlier `a` proposal. Optional-comma observations establish parser behavior, not owner approval of punctuation.

Final negative observations: missing or ambiguous Event mapping gives `incomplete` with no complete candidate; reusing E1 for the second event gives `invalid / COORDINATION_EVENT_IDENTITY_COLLAPSED`; global role/port fallback maps are refused; work/output limits give `resource_limit`. A limit of one retained alternative still permits the single complete conjunction. The standalone conjunction helper also refuses global fallback maps as `unsupported`; the runtime adapter returns an explicit incomplete refusal before calling it.

## Independent formal comparison

`independent-two-events-declared.json` was constructed independently from the existing direct-event target probe: E1 A→B and E2 B→A, with no source text/Evidence. Its signature is the explicit Event/experiencer/stimulus declaration from the selected metadata, not copied parser output. `compare-*-expression-declared` returns **proved_equivalent** for EN and PL after explicitly projecting request context out of both documents. Reference types, identities, temporal/multiplicity declarations and the two event arguments remain included. This is an expression-only proof, not proof of full context equivalence or the lambda/existential anonymous relation goal.

Full-context comparison remains unproved: the attempted comparison is refused with `EQUIVALENCE_REQUEST_CONTEXT_MISMATCH`. The independent context and source contexts have different identities and provenance/conditions; none was silently removed from the preserved originals. Earlier expression attempts used the undeclared independent input and correctly returned `UNKNOWN_SYMBOL`; the declared comparison inputs fix the missing declaration without changing the expected expression. Both failed attempts and successful follow-ups remain in the log.

## Reproduction and limits

`commands.jsonl` records exact argv/cwd/time/exit, output hashes and source HEAD. `environment.json` identifies the immutable final binary and Go runtime. `final-*` outputs use committed code `9f25acb`. Earlier draft commands have recorded dirty-tree fingerprints and are not final reproducibility receipts. The temporary syntax probe source is preserved as `syntax-probe.go.txt`; it is a diagnostic, not a committed runner or acceptance gate. Inputs, contexts, projected comparison documents and raw outputs are beside this file.

Coordinator checks on the final source implementation: gofmt, `go vet ./...`, `go build ./cmd/golingua`, `--help`, `--version`, and `git diff --check` passed in the real workspace. No new tests, runner or CI gate were added. Source-only composition does not complete independent target realization, translation, round-trips, quantifier scope or full WH context coverage. Those remain separate required work; target execution of this new conjunction is not claimed here.

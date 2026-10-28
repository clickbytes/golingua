# Independent two-clause target realization — #398, 2026-10-28

Final implementation: `a9df534`. This is the bounded target subtask in existing draft PR #422; the complete #398 remains **PARTIAL**. Agent technical assessment is separate from owner language acceptance (**NOT EVALUATED**).

## Observed final behavior

The `final-*` commands use the immutable binary and source SHA in `environment.json`. The independent input is the committed `../source-coordination/independent-two-events.json`: two direct positive Event atoms under `And`, E1 A→B and E2 B→A, explicit teacher/dog classification, no source text, source mentions or Source Evidence. Selected metadata provides declarations; the separate declared comparison input is standalone and cannot be combined with selected metadata. Initial attempts that incorrectly mixed inline declarations and metadata are preserved as `INLINE_DECLARATIONS_WITH_METADATA` refusals.

- EN realizes both clauses, with two coverage entries and sixteen bounded alternatives. Example: `The teacher sees the dog and the dog sees the teacher.` The result is `ambiguous`; the alternatives retain explicit determiner assumptions, including alternatives for repeated mentions. No arbitrary article choice is accepted automatically.
- PL realizes `Nauczyciel widzi psa i pies widzi nauczyciela.` with two coverage entries. The result remains `incomplete`: a complete supported alternative coexists with an unsupported aggregate WH-construction selection route, reported as `PARTIAL_TARGET_REALIZATION`. The refusal is retained; it is not a failed clause-binding proof or owner approval.
- Each composite alternative has one language-local connector and one final period. Coverage preserves distinct E1/E2 and the shared Entity IDs/roles. The merged Meaning and exact request context are checked before per-clause execution; each clause requires its own registry-validated generated evidence record.
- Unsupported or unmapped second terms return no text, preserving the unsupported term. Repeated Event identity is refused with `REPEATED_EVENT_IDENTITY`. Missing Event context is invalid. There is no supported-first-clause fallback.
- Eighty-character reference IDs are retained in coverage while bounded ordinal local Meaning IDs permit realization. Role swap produces `The dog sees the teacher and the dog sees the teacher.` / `Pies widzi nauczyciela i pies widzi nauczyciela.`, with the changed roles visible in coverage.
- Singleton and both subject/object WH controls still produce text. WH remains `incomplete` with its meaningful context residuals; these were not dropped. The singleton control deliberately retains unused second-event context, so it also retains that residual.

The plan JSON in each language is independently decoded with schema, construction ID, language and provenance validation. Existing clause morphology/compiled construction plans execute both clauses; the shared runtime contains no EN/PL connector dispatch. Assumptions, losses, diagnostics and evidence survive the bounded cross-product. Cross-product exhaustion is refused in code; that resource branch was not artificially induced through this CLI configuration.

## Review and correction provenance

One Luna Low implementation pass and one named correction were followed by coordinator review. The second pass introduced a WH regression: its coordination query-data gate ran before checking the two-event shape, refusing ordinary WH in both languages. The actual failed `review-*-wh-subject-metadata` outputs are retained. `draft-environment.json` and `pre-integration-draft.patch` identify that historical code from base `f2c6560`; it is not the delivered code.

After the implementation budget, the worker supplied only `supplied-integration.patch`; the coordinator reviewed and integrated it. It moves coordination-only gates after exact shape matching and reuses the existing strict surface validator for singleton and composite outputs instead of maintaining a weaker copy. Coordinator editorial integration also retained diagnostics in the one-record proof views, removed unused imports and placed participant-slice copying at the actual remap. Fresh builds and final EN/PL WH controls confirmed the regression is fixed. No third implementation pass, tests, runner or CI gate was added.

Coordinator verification on final code: gofmt, `go vet ./...`, `go build ./cmd/golingua`, actual help/version smoke and `git diff --check` passed. `commands.jsonl` records literal commands, exits and raw output hashes. Final outputs are reproduced on committed `a9df534`; draft/review/integrated outputs are historical observations, not final acceptance receipts.

## Remaining #398 work

Configured DSE translation still has the previous single-Atom/Ask source-context bridge. Its new two-clause source candidate must be integrated separately; source→target translation and both complete round-trips are not claimed here. Anonymous relation alias/existential closure, both quantifier scope directions and full WH context coverage remain required separate work. These direct-Atom outputs do not prove realization of the generic lambda relation goal or any quantifier truth.

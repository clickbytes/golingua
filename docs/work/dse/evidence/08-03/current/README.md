# DSE-08-03 current CLI evidence

These are bounded observations from the active #398 branch, not acceptance of the three full language slices. The required six paths per slice remain **NOT RUN** as a complete set; owner language acceptance remains **NOT EVALUATED**.

## Build and controls

- Source HEAD: `dab9d23b48fe429ca4221c5efdf8e081bc6405b3` plus the uncommitted #398 evaluator change.
- Binary: `/tmp/golingua-dse/08-03-evaluator/golingua`, SHA-256 `b16653da5489fa0062b4cc674a0ea4ca0330668fc46ca5afc210091c663cf097`.
- `go vet ./...`: PASS.
- `go build -o /tmp/golingua-dse/08-03-evaluator/golingua ./cmd/golingua`: PASS.
- `golingua --help` and `golingua --version`: PASS; version `0.1.0-alpha.1`.
- `git diff --check` and `gofmt`: PASS. No Go tests were added or run.
- Read-only code review of the temporal evaluator change: NO_BLOCKERS. The reviewer confirmed Event-only, present-only, explicit-At behavior and retained refusals for past, reference-time links, and multiplicity `many`.

Commands and raw outputs are in `commands.jsonl`, `replay/commands.jsonl`, `slices/commands.jsonl`, and their referenced output files. `replay/` reruns the earlier source/translate/realize controls on the active branch. `slices/` contains CLI checks and comparisons for independent IR, GoalSpec synthesis, source analysis, target realization, translation, scope/role swaps, and the evaluator. The canonical source inputs include the signatures emitted in the configured source-analysis candidate; independently authored formal candidates are in `data/semantic/acceptance/dse-08-03/`. Earlier stripped projections are kept separately to document the reproduction discrepancy.

## Observed CLI behavior

The four configured source-analysis inputs (`Who sees the dog?`, `The teacher sees whom?`, `Kto widzi psa?`, `Kogo widzi nauczyciel?`) produced complete typed `Ask` documents. They preserve the expected experiencer/stimulus gap distinction for these configured examples. This does not establish neutral WH syntax or owner-approved language quality.

With the four canonical `Ask` documents in `inputs/`, the independent evaluator produced:

| Probe | Observed result |
| --- | --- |
| EN subject gap, active explicit fact | `supported`, answer `A`, proof cites `dse-08-03-explicit-see-A-B` |
| EN object gap, same fact | `supported`, answer `B`, proof cites the same explicit fact |
| PL subject gap, active explicit fact | `supported`, answer `A`, proof cites the same explicit fact |
| PL object gap, same fact | `supported`, answer `B`, proof cites the same explicit fact |
| Empty or expired fact set | `knowledge: unknown`, no answers |
| Present Event without a query instant | `unsupported`, `QUERY_INSTANT_REQUIRED` |
| Past Event with a query instant | `unsupported`, `UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE` |

The fact is a synthetic diagnostic probe with agent-authored provenance. It is not an owner-supplied world fact or a language gold. The evaluator consumes the typed Ask body and explicit fact; it does not validate the natural-language presuppositions as facts. This evidence covers query evaluation only.

The active branch rerun confirmed the remaining integration gaps: the production `translate` path does not accept these WH inputs; `algebra realize` refuses the independent Ask for both PL and EN with `UNSUPPORTED_EXPRESSION_PART`; the proposed scope example stops at `UNSUPPORTED_SOURCE_FOREST`. The two-hop relation’s six language routes were not implemented or evaluated.

The independent WH subject/object documents, relation GoalSpec/expression, and two scope trees all pass `algebra check`. The relation GoalSpec builds a two-atom `pilot.r` path with one shared binder; `algebra compare` proves the generated expression equivalent to the independent expression. `algebra compare` also proves each of four configured source WH outputs equivalent to its independent EN/PL Ask. Role-swap and scope-swap comparisons keep different normal forms and return `undetermined`; neither result is a proof of non-equivalence. Proposed two-hop and scope source strings were tried in both languages; all four source-analysis and translation routes return `UNSUPPORTED_SOURCE_FOREST`. Independent target realization attempts for WH, relation, and both scope trees in EN/PL return `UNSUPPORTED_EXPRESSION_PART`.

## Historical baseline discrepancy

The old `baseline/outputs/en-wh-evaluate.stdout.json` records `UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE`. Re-running the exact command with the preserved baseline binary (SHA-256 `2890fe56ea260a43213b6e2e65de0056f00d53213f5d85a2099fa51a4cbdd5b5`) and the preserved projected Ask/facts instead returns `UNKNOWN_SYMBOL` at `expression.body.symbol`; see `replay/commands.jsonl` entry `baseline-binary-en-wh-evaluate-reproduction`. The projected Ask omitted all six declared signatures, including `pilot.vision_event`.

The temporal-refusal record therefore cannot be reproduced from its preserved command, binary, and inputs. We do not treat that historical line as verified evidence. The corrected canonical documents in `inputs/` retain their emitted signatures and independently reproduce the bounded evaluator results above.

## Limits

This does not complete #398. Independent Meaning→EN/PL realization, all six routes, the relation-composition slice, and the scope-contrast slice remain open. The result is `PARTIAL`; the agent review and owner acceptance are separate records.

# DSE-08-03 current CLI evidence

This report records the active #398 branch observations. The WH slice now reaches the production CLI in both languages, but its realization and translation results remain `incomplete`; relation and scope remain unsupported in the language paths. Overall status is **PARTIAL**. Owner language acceptance is **NOT EVALUATED**.

## Build and controls

- Base/source HEAD recorded by the latest run: `3e9b7dcdbd389c54b65c02d8877a05dde2553b78`; the current working tree contains the #398 changes.
- Latest executable: `/tmp/golingua-dse-08-03-current`, SHA-256 `46758c5a47a3c031697b0ebdc0ccc5be6cded475b8eb2d99b1a7a43a141f545e`.
- `gofmt -l` on changed Go files: clean; `git diff --check`: PASS.
- `go vet ./...`: PASS; `go build -o /tmp/golingua-dse-08-03-current ./cmd/golingua`: PASS.
- `--help` and `--version`: PASS; version `0.1.0-alpha.1`. No new Go tests were added or run.
- A read-only final diff review found no blockers. The reviewer separately flagged the stale evidence text, which this update corrects.

The current run manifest is `slices/current-run-commands.jsonl`; raw stdout and stderr are under `slices/outputs/current-*`. The earlier failed comparison with full request contexts is preserved in `slices/current-run-commands.pre-context-projection.jsonl`. The working expression-only comparisons intentionally omit request context; they prove only the Ask expression, not equivalence of referents, assumptions, or presuppositions. An earlier evaluator run and the historical binary reproduction remain in `slices/commands.jsonl` and `replay/commands.jsonl`.

## Current WH observations

| Route | Current observation |
| --- | --- |
| EN/PL source analysis | All four subject/object inputs return `ok` and the expected typed `Ask` role gap. The four expression-only comparisons return `proved_equivalent` to their independently authored WH expressions. |
| Independent Meaning → EN/PL | All four realizations emit a surface candidate, but overall status is `incomplete`. Context referents and conditions not consumed by the selected expression remain explicit residuals. English also records caller-supplied person/definiteness assumptions where required. |
| Translation EN↔PL | The four observed pairs are `Who sees the dog?` → `Kto widzi psa?`, `Kto widzi psa?` → `Who sees the dog?`, `The teacher sees whom?` → `Kogo widzi nauczyciel?`, and `Kogo widzi nauczyciel?` → `Whom does the teacher see?`. Every result remains `incomplete`: Polish candidates retain context residuals; English candidates expose assumptions such as person or definite article. Candidate status is also `incomplete`. |
| Surface re-analysis | All four translated surface candidates re-analyze with status `ok`. Their expression-only comparisons return `proved_equivalent`; full-context comparison is not established because the translation context and independently authored context use different referent sets. |
| Negative agreement control | `Whom does the teachers see?` is refused with `UNSUPPORTED_SOURCE_FOREST`; the bounded do-support path requires an adjacent singular third-person NP. `forms --surface does` returns `UNKNOWN_SURFACE_FORM` because `does` is a construction terminal, while `forms --surface see` finds the base form. |

The exact output records are in `slices/outputs/current-analyze-*`, `current-realize-*`, `current-translate-*`, `current-roundtrip-analyze-*`, and the corresponding `*-expression-compare` files. The relation alias comparison is recorded in `slices/outputs/compare-two-hop-see-alias-proposal.*`; its proposal data and profile are in `data/semantic/acceptance/dse-08-03/relation/`. A previous round-trip attempt failed at `UNSUPPORTED_SOURCE_FOREST` for English object WH; the preserved pre-fix output is `slices/outputs/pre-fix-roundtrip-analyze-pl-en-object.*`. The current exact singular construction now parses `Whom does the teacher see?` as an object-gap Ask.

## Evaluator evidence

The bounded evaluator probe uses an explicit synthetic fact with agent-authored provenance; it is not an owner-supplied world fact or a language gold. Its preserved CLI evidence reports:

| Probe | Observed result |
| --- | --- |
| EN/PL subject gap with the explicit fact | `supported`, answer `A`, with proof citing `dse-08-03-explicit-see-A-B` |
| EN/PL object gap with the same fact | `supported`, answer `B`, with proof citing the same fact |
| Empty or expired fact set | `knowledge: unknown`, no answers |
| Present event without a query instant | `unsupported`, `QUERY_INSTANT_REQUIRED` |
| Past event with a query instant | `unsupported`, `UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE` |

This checks query evaluation only. It does not establish the natural-language presuppositions as facts, and the question is not inserted as an assertion.

## Relation and scope observations

- The independent `r∘r` GoalSpec synthesizes two `pilot.r` atoms with a shared binder. A separate agent-reviewed formal proposal defines `pilot.r(a,b)` as `exists e. pilot.vision_event(e,a,b)`; the current CLI comparison proves the path equivalent to two independently bound seeing events under the selected `alias_v1` profile. This is a formal proposal only, not a production lexical mapping or owner acceptance. The latest EN/PL source and translation attempts stop at `UNSUPPORTED_SOURCE_FOREST`; production target realization is `unsupported` with `NO_EXACT_REALIZATION_BINDING`.
- The independent scope trees pass `algebra check`. Scope swaps retain distinct normal forms, while comparison returns `undetermined` (not proof of non-equivalence). Both earlier proposed EN/PL surface pairs (`Not every...` / `Nie każdy...` and `Every...no...` / `Każdy...żadnego...`) were attempted; source and translation return `UNSUPPORTED_SOURCE_FOREST`. A language review found these forms do not establish a single scope reading reliably. More explicit paraphrases are recorded as unverified proposals in `docs/work/dse/vertical-slices.md`; they have not yet been run. Independent target realization is `unsupported` with `NO_EXACT_REALIZATION_BINDING`.
- These two slices do not have complete production language paths. Their candidate surface readings and mappings remain proposals, not owner-approved language data.

## Historical baseline discrepancy

The old `baseline/outputs/en-wh-evaluate.stdout.json` records `UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE`. Re-running the saved command with the preserved baseline binary (SHA-256 `2890fe56ea260a43213b6e2e65de0056f00d53213f5d85a2099fa51a4cbdd5b5`) and the same projected Ask/facts returns `UNKNOWN_SYMBOL` at `expression.body.symbol`; see `replay/commands.jsonl` entry `baseline-binary-en-wh-evaluate-reproduction`. The projected Ask omitted all six declared signatures, including `pilot.vision_event`. That historical temporal-refusal result is not treated as reproducible evidence. The canonical documents in `inputs/` retain signatures emitted by the configured source-analysis candidate.

## Limits and acceptance

WH source analysis and the two translated round-trip directions are observable, but the independent realization and all four translations remain `incomplete`. Relation and scope do not have supported production language paths. Therefore #398 is **PARTIAL**, not DONE. The agent review and owner language acceptance are separate; owner acceptance remains **NOT EVALUATED**.

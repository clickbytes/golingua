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

- The independent `r∘r` GoalSpec synthesizes two `pilot.r` atoms with a shared binder. A separate agent-reviewed formal proposal defines `pilot.r(a,b)` as `exists e. pilot.vision_event(e,a,b)`; the current CLI comparison proves the path equivalent to two independently bound seeing events under the selected `alias_v1` profile. This is a formal proposal only, not a production lexical mapping or owner acceptance. Syntax-stage diagnostics show the coordinated EN/PL inputs have `forest_outcome=unknown`, zero roots, and no licensed reading for `and` / `a`; single-clause controls have one transitive root. The production `algebra analyze` wrapper surfaces this as `UNSUPPORTED_SOURCE_FOREST`. Target-side code review confirms the algebra realizer handles a direct `Atom` or one direct atom in a root `And`, and does not expand aliases into two events. EN's direct `pilot.r` binding belongs to a different link sense; PL has no direct `pilot.r` binding. Independent target realization therefore remains `unsupported` with `NO_EXACT_REALIZATION_BINDING`.
- The independent scope trees pass `algebra check`. Scope swaps retain distinct normal forms, while comparison returns `undetermined` (not proof of non-equivalence). Both earlier proposed EN/PL surface pairs (`Not every...` / `Nie każdy...` and `Every...no...` / `Każdy...żadnego...`) were attempted; source syntax has zero roots because `every` / `każdy` and, in the All-Not inputs, `no` / `żadnego` have no licensed reading, and no start construction is available. The production `algebra analyze` wrapper reports `UNSUPPORTED_SOURCE_FOREST`. A language review found the earlier surfaces do not establish a single scope reading reliably. The four more explicit paraphrases in `docs/work/dse/vertical-slices.md` were subsequently probed: each has zero syntax roots and public `algebra analyze` returns `UNSUPPORTED_SOURCE_FOREST`; raw results and hashes are in `current/slices/scope-paraphrase-probe/`. They remain unverified proposals. A separate Luna Low language review finds the F1 proposals logically equivalent by an existential-counterexample paraphrase, not literal outer `Not(All(...))`; its explicit outer-negation alternative has not been probed. F2 was assessed as scope-transparent. Owner acceptance remains NOT EVALUATED. Target-side code review confirms negation and quantifier roots are outside the direct-atom realization profile; independent target realization remains `unsupported` with `NO_EXACT_REALIZATION_BINDING`.
- These two slices do not have complete production language paths. Their candidate surface readings and mappings remain proposals, not owner-approved language data.

The isolated syntax probe source is `slices/probes/source-syntax-probe.go.txt`; it calls `SourcePipeline.AnalyzeSyntax` without running semantic composition. Its exact EN/PL outputs are `slices/outputs/source-syntax-diagnostics-{en,pl}.stdout.jsonl`, with empty stderr files. The simple-clause controls parse as one `construction:{en,pl}.transitive` root; the relation and scope inputs do not produce a root. This identifies the earliest observed boundary as language-local lexical/morphological data and source construction coverage. The downstream `UNSUPPORTED_SOURCE_FOREST` diagnostic is not evidence that `ComposeDSE` itself lost a valid parse.

Target realization has a separate, later limitation: `internal/engine/algebra_realization.go` searches exact target bindings for a direct `Atom` or one direct atom within a root `And`, while DSE translation accepts only a direct positive atom or direct-atom WH `Ask`. The formal alias-comparison metadata is not part of that production target interface. The general Meaning target path also descends through a statement discourse only when it has exactly one content; `TargetPipeline.Realize` linearly returns separate event realizations, not one coordinated clause. Construction slots currently omit a clause category. Relation realization therefore needs a general, meaning-preserving composed-event route and language clause coordination; scope realization needs a typed DSE target path for quantifier scope plus language-local quantifier constructions. A direct `pilot.r` binding alone would not repair either limitation.

## Historical baseline discrepancy

The old `baseline/outputs/en-wh-evaluate.stdout.json` records `UNSUPPORTED_RELATIVE_TEMPORAL_REFERENCE`. Re-running the saved command with the preserved baseline binary (SHA-256 `2890fe56ea260a43213b6e2e65de0056f00d53213f5d85a2099fa51a4cbdd5b5`) and the same projected Ask/facts returns `UNKNOWN_SYMBOL` at `expression.body.symbol`; see `replay/commands.jsonl` entry `baseline-binary-en-wh-evaluate-reproduction`. The projected Ask omitted all six declared signatures, including `pilot.vision_event`. That historical temporal-refusal result is not treated as reproducible evidence. The canonical documents in `inputs/` retain signatures emitted by the configured source-analysis candidate.

## Limits and acceptance

WH source analysis and the two translated round-trip directions are observable, but the independent realization and all four translations remain `incomplete`. Relation and scope do not have supported production language paths. Therefore #398 is **PARTIAL**, not DONE. The agent review and owner language acceptance are separate; owner acceptance remains **NOT EVALUATED**.

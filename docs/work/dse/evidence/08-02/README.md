# Evidence #397 — runtime integration

This directory records the bounded DSE integration on branch `dse/08-02-runtime-integration`, based on `ce8c0653b3d98d9929663eddbf74a60683d735f6`, in [PR #421](https://github.com/clickbytes/golingua/pull/421). The captured binary, exact resource hashes, command arguments, raw outputs, and exit codes are in [`environment.json`](environment.json), [`runtime-config-positive.json`](runtime-config-positive.json), [`runtime-config-unsupported.json`](runtime-config-unsupported.json), [`inputs/`](inputs/), and [`outputs/`](outputs/).

The route is opt-in through `semantic_profile: "dse"`; the same input without that field remains on Meaning v1. The configured route validates the exact registry, source and target language packages, source plans, contexts, and active Pack before translating. There is no automatic retry through Meaning v1.

## Observed CLI behavior

All observations below were recaptured from the final coordinator build. Each `outputs/*.run.json` contains argv, the shell-safe command, binary SHA-256, UTC capture time, exit code, and stdout/stderr filenames. The neighboring stdout file is the unmodified CLI response.

| Command / input | Result | Evidence |
| --- | --- | --- |
| Configured DSE capabilities | `ok`, ready; representation and analysis/realization are reported for EN and PL, evaluation is empty | [`capabilities-dse-positive`](outputs/capabilities-dse-positive.stdout.json) |
| PL `Nauczyciel widzi psa.` → EN | `ambiguous`, exit 6; four explicit article alternatives remain, each `incomplete` with determiner assumptions | [`translate-pl-en`](outputs/translate-pl-en.stdout.json) |
| Same request with explicit preview ID | Top-level text is the caller-selected `The teacher sees the dog.` while overall status stays `ambiguous` and all four alternatives remain | [`translate-pl-en-explicit-preview`](outputs/translate-pl-en-explicit-preview.stdout.json) |
| EN `The teacher sees the dog.` → PL | `unsupported`, exit 3; two exact lexical descriptors are recorded as consumed, while two definite-reference conditions remain residuals. The strict DSE adapter refuses before target generation | [`translate-en-pl-definite`](outputs/translate-en-pl-definite.stdout.json) |
| PL → EN with an unsupported source context | `unsupported`, exit 3; two residual context conditions remain and no text is returned | [`translate-pl-en-context-constraints`](outputs/translate-pl-en-context-constraints.stdout.json) |
| EN → PL with the alternate unsupported context | `unsupported`, exit 3; the same pre-generation refusal boundary is observed | [`translate-en-pl-context-constraints`](outputs/translate-en-pl-context-constraints.stdout.json) |
| Missing registry, wrong plan version, or mismatched context pin | Each capabilities request is `invalid`, exit 2, before the engine becomes ready | [`missing registry`](outputs/capabilities-dse-missing-registry.stdout.json), [`wrong plan`](outputs/capabilities-dse-wrong-plan-version.stdout.json), [`context mismatch`](outputs/capabilities-dse-context-mismatch.stdout.json) |
| Existing v1 capabilities and the same PL → EN sentence | v1 remains available without `semantic_profile`; its separate result has no DSE trace | [`v1 capabilities`](outputs/capabilities-v1.stdout.json), [`v1 translation`](outputs/translate-v1-pl-en.stdout.json) |
| Standalone DSE analyze; standalone PL and EN algebra realization | PL analyze and PL realization are `ok`; EN realization is `ambiguous` with its licensed article alternatives | [`analyze`](outputs/algebra-analyze-pl.stdout.json), [`PL realize`](outputs/algebra-realize-pl.stdout.json), [`EN realize`](outputs/algebra-realize-en.stdout.json) |
| CLI help/version | Both exit 0; version is `0.1.0-alpha.1` | [`help`](outputs/help.stdout.txt), [`version`](outputs/version.stdout.txt) |

The positive PL → EN response records the checked canonical candidate, active Pack identity, registry fingerprint, exact source/target resource pins, source role-to-reference mappings, target coverage, and two `consumed_conditions` receipts. The receipts are limited to exact `lexical_descriptor` presuppositions whose source sense, approved reference-predicate binding, matching fragment evidence, and unique approved CORE transfer agree. The target rechecks its projection against separately pinned metadata. Source mention-resolution evidence remains in the canonical trace; it does not become a second target-side authority.

The EN → PL refusal is an intentional limitation of this profile: its source context contains definite-reference predicates that the current Polish target route cannot consume. Those conditions remain visible and prevent target generation. This delivery does not claim that they are equivalent to Polish surface morphology.

## Technical checks

| Check | Result |
| --- | --- |
| `gofmt -l` for all changed Go files | PASS; no output |
| `git diff --check` | PASS |
| `go vet ./...` | PASS, exit 0 |
| `go build -o /tmp/golingua-dse/397/golingua ./cmd/golingua` | PASS, exit 0 |
| Built binary `--help` / `--version` | PASS, exit 0 |
| `go test ./...` | NOT RUN; not required by the active issue instructions |
| PR CI | PENDING for PR #421 |
| Post-merge `main` CI | NOT RUN; pending PR merge |

OpenCode was unavailable for execution because this environment had no authenticated provider/model, as recorded in the #397 START comment. The bounded implementation was delegated to Luna Low; the coordinator integrated and verified it. A separate Luna Low read-only final review passed the strict pre-generation refusal and meaning/provenance checks. Coordinator technical verification and agent review are not owner language approval. Owner language acceptance: **NOT EVALUATED**.

The result is bounded to one positive direct Atom and the configured PL/EN pilot. It does not claim that target article choices are resolved for PL → EN, that all English source contexts are realizable in Polish, or that the runtime evaluates world facts. These limits remain explicit rather than being guessed away.

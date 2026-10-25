# Codex Sol Medium-led execution protocol for GoLingua

This complements root `AGENTS.md` without changing project architecture. Applies to all Go v0 issues.

For the owner-authorized Dynamic Semantic Expressions plan, use the [DSE issue map](work/dse/issue-map.json), [DSE execution runbook](work/dse/execution.md), and [DSE work index](work/dse/README.md). These links do not create a second execution policy; root `AGENTS.md` and the active leaf remain authoritative.

## Decision rights and allowed concurrency

| Actor | May decide scope/priority? | May write code? | May write GitHub? | May approve merge? |
| --- | --- | --- | --- | --- |
| Codex Sol Medium (main) | Yes, under active issue and owner constraints | Editorial/integration changes; delegate implementation | Yes | Yes, after review and real CI |
| Codex Luna Low scout / reviewer | No | No | No | No |
| Codex Luna Low micro-implementer | No | Yes, when sole writer with assigned paths | No | No |
| External OpenCode YOLO worker | No | Yes, in isolated disposable workspace, assigned paths only | No | No |

Owner update 2026-10-23: delegate small, precisely scoped implementation jobs to OpenCode; GPT Luna Low may assist. The coordinator owns editing, integration, review and verification. At most four helpers total may run concurrently, subject to actual tool limits; OpenCode is a separate process. At most one implementation writer, with no simultaneous coordinator edits to its assigned files.

## Per-issue state machine

`QUEUED -> AUTHORIZED -> SCOPED -> WRITING -> REVIEW -> VERIFY -> PR -> MERGED -> CLOSED`

At any gate: `BLOCKED` (with evidence and a specific blocker) or `PARTIAL` (with incomplete acceptance criteria). No automatic retries or automatic activation of a different child issue when blocked. Roles `issue_scout`, `diff_reviewer` and `micro_implementer` are registered explicitly in `.codex/config.toml`, referencing `.codex/agents/*.toml`; standalone role files without these registrations are not enough.

1. **AUTHORIZED:** Sol checks clean checkout, current main SHA, active epic and child status, previous issue/PR closure and current open PRs. Follow only ONE current child. G11-V1 was still QUEUED at the review snapshot.
2. **SCOPED:** Read active child issue and linked current contracts only, then trace implementation reality in code. Draft a 3-8 line issue comment with SHA, exact files/symbols, observable result, risks and verification. Do not turn an issue into a wall of planning text.
3. **DELEGATION:** Always split implementation into small closed jobs with an exact JOB BRIEF, delegating to OpenCode. GPT Luna Low can handle bounded analysis/review or small local jobs. The coordinator defines architecture/state contracts, redacts, integrates, reviews and verifies the result. Respect explicit OpenCode/isolation requirements; unavailable tools do not justify host YOLO.
4. **WRITING:** Exactly one writer. The worker gets the complete brief. Sol does not simultaneously edit its files. The worker returns a patch/diff and a structured handoff, not an unbounded follow-up plan.
5. **REVIEW:** Sol checks `git diff` against the exact scope, worker's claims vs actual code, existing CLI and app contract, any changed docs/tests and all changed paths. Optionally assign read-only Luna diff review; Luna does not decide acceptance.
6. **VERIFY:** For changed Go: run existing gofmt/vet/build and issue-relevant CLI/terminal observation. For docs-only: links and `git diff --check`, with actual CI on PR. Do not create new tests unless the active issue explicitly specifies them. Record PASS/FAIL/INFRA/NOT RUN as applicable; do not fabricate commands or outputs.
7. **PR:** By default Sol owns one branch, commits, issue-linked PR, CI inspection and decision to merge. After merging, confirm main CI. Then close/update child + epic in one concise transaction. A specific, explicit owner instruction can exceptionally authorize a direct commit to fresh `main` with a head-SHA lease and the same diff/CI/issue evidence; this does not waive the PR gate for future issues. Do not write GitHub from workers.

## Authority and version boundaries

Latest explicit owner decision and standing v0 invariants outrank older issues; the active issue defines authorized implementation, explicitly linked current contracts define how that part behaves, epic sets sequence, older docs are contextual. Code is evidence of current behavior, not authorization to implement other work. If conflicts would change public contract or violate v0, Sol must stop for owner clarification rather than rewrite issues or docs silently. `Bubble Tea v2` is a pre-approved third-party library version, not a GoLingua product v2.

Do not add product migrations, versioned compatibility adapters, future-proof abstractions, unsolicited test suites/fixtures, JS parity, unrelated cleanups, bulk data, or docs outside the issue. Do not write acceptance claims without observed evidence. Keep owner approval separate from agent technical assessment.

## Loop breakers

- One worker implementation pass. One correction only after Sol names a concrete defect in a new brief. No silent retries.
- After worker limits, Sol may only integrate/redact a small supplied correction within the issue; otherwise mark BLOCKED/PARTIAL with the exact missing dependency.
- OpenCode worker has `steps: 12` (V2 has no `doom_loop` action); Sol should also impose an external process timeout when it launches it. These controls are separate from proof of functional correctness.
- Never rerun an unchanged failing command repeatedly. Categorize failure as: CODE, SPEC, ENV, EXTERNAL, or OUT_OF_SCOPE. Each retry must state a new hypothesis and a different action.
- STOP on repeated identical failure, failed acceptance after allowed correction, off-scope file modifications, need for a migration/new test framework, credential requirement, ambiguous domain semantics, or issue/document authority conflict.
- Report `DONE` only for met acceptance criteria with evidence. `PARTIAL` lists unmet criteria; `BLOCKED` names blocking dependency. `NOT RUN` is never PASS. `go-ci` green does not imply linguistic acceptance.

## Example isolated OpenCode V2 handoff

The coordinator uses a disposable code snapshot without `.git` or credentials.
Create an explicit `opencode.json` in its root for each delegation; the agent
Markdown file alone does not prove configuration is loaded. Use V2 `permissions`,
`shell` and `subagent`, not V1 `permission`, `bash` or `task`. V2 has no
`doom_loop` permission action. `steps: 12` and an external timeout bound the run.

```sh
BASE=$(git rev-parse HEAD)
WORKER_DIR=$(mktemp -d -t golingua-worker-XXXXXX)
git archive "$BASE" | tar -x -C "$WORKER_DIR"
# Coordinator writes the complete JOB BRIEF and explicit opencode.json here.
# Required JSON fields: model, default_agent, agents (system/steps/permissions),
# and any provider settings needed by the verified free model.
# Deny by default; allow read/edit only for the exact assigned paths.
# Explicitly deny shell, subagent, webfetch, websearch and external_directory.
# Verify effective config/agent/model in this same environment before editing.
# Model catalogs can be empty before the location finishes initialization;
# inspect actual session/effective API state rather than assuming no models.
timeout 180 docker run --rm --cap-drop ALL --security-opt no-new-privileges \
  --pids-limit 128 --memory 1g --cpus 2 \
  -v "$WORKER_DIR:/workspace" -w /workspace \
  ghcr.io/anomalyco/opencode:2.0.15 run --standalone --auto \
  --agent golingua-worker --model opencode/glm-5.3-flash-free \
  --file /workspace/job-brief.txt 'Implement only the attached JOB BRIEF'
# Inspect every changed path against the snapshot and review the complete diff.
# Integrate only a reviewed patch; do not copy configuration or unrelated files.
```

Follow root `AGENTS.md` for daily model availability/cost/ranking review and
immediate refresh on failures or relevant changes. Before `--auto`, inspect
loaded config and actual permission behavior. A tool removed from the exposed
toolset is reported as unavailable, never as an executed denial probe. An
external read rejected by the guard is an actual observed denial. Never report
an unattempted tool call as executed. Unsupported restrictions, unexpected
permissions or unexplained missing agents stop delegation.

## Historical example: G11-V1 splitting (without more micro-PRs)

At the time of this example, Sol owned scope, public CLI/TUI startup decisions, the `internal/app.Service` contract and integration/review of delegated implementation jobs. Luna Low could map `cmd/golingua` and `internal/app` read-only; the implementation writer owned only a sharply bounded shell component. Sol integrated it and owned the G11-V1 PR. This is a process example, not the current queued-work status.

## Completion gate for user-visible flows

Required observable product behavior (for example, G11-V1 opening an actual TTY and executing F1 Translate) is NOT proven by a `NOT RUN` entry. Report PARTIAL/BLOCKED until it is actually verified, unless the owner explicitly waives that exact acceptance criterion. Owner linguistic acceptance can stay NOT EVALUATED without blocking otherwise verified technical work.

## OpenCode version and isolation gate

Before any `--auto` run, Sol verifies `opencode --version`, `opencode run --help`, the local model, and actual denial of shell/both shell and bash aliases, web/network, nested task execution and external-directory access. If any configured deny rule is unsupported, refuse YOLO and use Sol or Luna instead.

`steps: 12` caps model iterations, not wall-clock runtime, number of file writes or OS resource consumption. Add an external process timeout and run in an isolated disposable container/VM without secrets and GitHub credentials. A worktree alone is NOT security isolation; it shares repository metadata and often host credentials.

## G11 closeout owner decision (2026-10-23)

Deliver V4 → V5 → V6 → V7 → V8 sequentially, one child and PR per usable slice, after prior merge, passing main CI and child closure. The four-helper ceiling is subject to available tool slots and retains exactly one implementation writer. Optional roles: OpenCode implementation, Luna Low analysis, Luna Low review, OpenCode read-only evidence/document proposal. Select an actually available free OpenCode model through the documented daily review and per-job permission gates. If unavailable or a mandatory gate cannot be verified, use bounded Luna Low jobs; do not bypass isolation.

## Owner-authorized A1 matrix exception (2026-10-23)

The later owner decision authorizes the first small level-oriented PL/EN sample under `tests/A1/*.testdomain.go` (#309), superseding the older blanket matrix-directory prohibition for this scope. The first delivery is declarative domain data without a runner, `*_test.go` or a CI language gate. Future A1.1 batches require their own defined scope; do not create a backlog automatically. References are agent proposals, separate from actual observations and owner acceptance; never turn an existing unsupported result into the desired behavior. See [matrix rules](../tests/README.md). Serial delegation/review/verification rules remain in force.

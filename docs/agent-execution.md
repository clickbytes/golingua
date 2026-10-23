# Codex Sol Medium-led execution protocol for GoLingua

This complements root `AGENTS.md` without changing project architecture. Applies to all Go v0 issues.

## Decision rights and allowed concurrency

| Actor | May decide scope/priority? | May write code? | May write GitHub? | May approve merge? |
| --- | --- | --- | --- | --- |
| Codex Sol Medium (main) | Yes, under active issue and owner constraints | Yes, when no worker has writer lease | Yes | Yes, after review and real CI |
| Codex Luna Low scout / reviewer | No | No | No | No |
| Codex Luna Low micro-implementer | No | Yes, when sole writer with assigned paths | No | No |
| External OpenCode YOLO worker | No | Yes, in isolated disposable workspace, assigned paths only | No | No |

Codex max native threads is 2, but OpenCode is a separate process and is NOT included in that limit. The coordinator enforces at most 2 workers in total, no more than 1 implementation writer, and never edits the same files concurrently.

## Per-issue state machine

`QUEUED -> AUTHORIZED -> SCOPED -> WRITING -> REVIEW -> VERIFY -> PR -> MERGED -> CLOSED`

At any gate: `BLOCKED` (with evidence and a specific blocker) or `PARTIAL` (with incomplete acceptance criteria). No automatic retries or automatic activation of a different child issue when blocked. Roles `issue_scout`, `diff_reviewer` and `micro_implementer` are registered explicitly in `.codex/config.toml`, referencing `.codex/agents/*.toml`; standalone role files without these registrations are not enough.

1. **AUTHORIZED:** Sol checks clean checkout, current main SHA, active epic and child status, previous issue/PR closure and current open PRs. Follow only ONE current child. G11-V1 was still QUEUED at the review snapshot.
2. **SCOPED:** Read active child issue and linked current contracts only, then trace implementation reality in code. Draft a 3-8 line issue comment with SHA, exact files/symbols, observable result, risks and verification. Do not turn an issue into a wall of planning text.
3. **DELEGATION DECISION:** Delegate only if one coherent subtask fits a short JOB BRIEF, has reviewable output, and saves more effort than orchestration. Prefer Luna for exploration/diff checking. Prefer Sol for architecture, TUI state ownership, engine meaning/status, snapshot mutations and cross-module design. Use OpenCode only for low-risk mechanical implementation with explicit paths and no GitHub access. Delegation is optional, not automatic.
4. **WRITING:** Exactly one writer. The worker gets the complete brief. Sol does not simultaneously edit its files. The worker returns a patch/diff and a structured handoff, not an unbounded follow-up plan.
5. **REVIEW:** Sol checks `git diff` against the exact scope, worker's claims vs actual code, existing CLI and app contract, any changed docs/tests and all changed paths. Optionally assign read-only Luna diff review; Luna does not decide acceptance.
6. **VERIFY:** For changed Go: run existing gofmt/vet/build and issue-relevant CLI/terminal observation. For docs-only: links and `git diff --check`, with actual CI on PR. Do not create new tests unless the active issue explicitly specifies them. Record PASS/FAIL/INFRA/NOT RUN as applicable; do not fabricate commands or outputs.
7. **PR:** By default Sol owns one branch, commits, issue-linked PR, CI inspection and decision to merge. After merging, confirm main CI. Then close/update child + epic in one concise transaction. A specific, explicit owner instruction can exceptionally authorize a direct commit to fresh `main` with a head-SHA lease and the same diff/CI/issue evidence; this does not waive the PR gate for future issues. Do not write GitHub from workers.

## Authority and version boundaries

Latest explicit owner decision and standing v0 invariants outrank older issues; the active issue defines authorized implementation, explicitly linked current contracts define how that part behaves, epic sets sequence, older docs are contextual. Code is evidence of current behavior, not authorization to implement other work. If conflicts would change public contract or violate v0, Sol must stop for owner clarification rather than rewrite issues or docs silently. `Bubble Tea v2` is a pre-approved third-party library version, not a GoLingua product v2.

Do not add product migrations, versioned compatibility adapters, future-proof abstractions, unsolicited test suites/fixtures, JS parity, unrelated cleanups, bulk data, or docs outside the issue. Do not write acceptance claims without observed evidence. Keep owner approval separate from agent technical assessment.

## Loop breakers

- One worker implementation pass. One correction only after Sol names a concrete defect in a new brief. No silent retries.
- One direct Sol correction after worker limits only when small and within existing issue; otherwise mark BLOCKED/PARTIAL with next human decision or missing contract.
- OpenCode worker has `steps: 12` and `doom_loop: deny`; Sol should also impose an external process timeout when it launches it. These controls are separate from proof of functional correctness.
- Never rerun an unchanged failing command repeatedly. Categorize failure as: CODE, SPEC, ENV, EXTERNAL, or OUT_OF_SCOPE. Each retry must state a new hypothesis and a different action.
- STOP on repeated identical failure, failed acceptance after allowed correction, off-scope file modifications, need for a migration/new test framework, credential requirement, ambiguous domain semantics, or issue/document authority conflict.
- Report `DONE` only for met acceptance criteria with evidence. `PARTIAL` lists unmet criteria; `BLOCKED` names blocking dependency. `NOT RUN` is never PASS. `go-ci` green does not imply linguistic acceptance.

## Example OpenCode worktree handoff (POSIX demonstration)

> This is a reviewable diff workflow, NOT a security sandbox. A separate worktree shares Git metadata with the source repository and the process may retain host credentials. For genuinely untrusted YOLO mode, use a container/VM with no secrets or GitHub credentials and a disposable code snapshot. Sol must audit the patch before applying.

```sh
# Sol must ensure one writer and a clean, expected base SHA first.
BASE=$(git rev-parse HEAD)
WORKER_DIR=$(mktemp -d -t golingua-worker-XXXXXX)
git worktree add --detach "$WORKER_DIR" "$BASE"
# Supply a complete, issue-specific job brief as the prompt; no second planner.
# Replace ./job-brief.md with Sol's temporary filled brief.
opencode run --auto --agent golingua-worker --dir "$WORKER_DIR" "$(cat ./job-brief.md)"
# Run git operations as Sol, not as OpenCode.
git -C "$WORKER_DIR" add -N .   # include newly added files in the diff
git -C "$WORKER_DIR" diff --binary > ./worker-output.patch
git -C "$WORKER_DIR" diff --name-only
# Stop if any modified path is not explicitly allowed by the job brief.
git apply --check ./worker-output.patch
# Do not automatically git apply. Sol reviews the exact diff first.
# After recording evidence and applying only an accepted patch:
# git worktree remove --force "$WORKER_DIR"
```

Reviewers must verify that the worker did not change AGENTS.md, CI, GitHub metadata, docs or tests unless the brief expressly allowed it. Never paste credentials into the worker prompt. The sample command assumes OpenCode is installed, authenticated to a model provider and project agent files are present; otherwise fall back to a Sol or Luna-only workflow.

## Example G11-V1 splitting (without more micro-PRs)

Sol owns the overall G11-V1 TUI vertical slice, public CLI/TUI startup decision and `internal/app.Service` contract. Luna Low can map existing `cmd/golingua` and `internal/app` call paths read-only. The implementation writer may own one sharply bounded presentational component of the shell, if Sol has specified the exact types/interfaces; it may not invent state ownership or status semantics. Sol integrates, drives actual TTY and non-TTY probes and owns the sole G11-V1 PR. G11-V2 and later remain queued.

## Completion gate for user-visible flows

Required observable product behavior (for example, G11-V1 opening an actual TTY and executing F1 Translate) is NOT proven by a `NOT RUN` entry. Report PARTIAL/BLOCKED until it is actually verified, unless the owner explicitly waives that exact acceptance criterion. Owner linguistic acceptance can stay NOT EVALUATED without blocking otherwise verified technical work.

## OpenCode version and isolation gate

Before any `--auto` run, Sol verifies `opencode --version`, `opencode run --help`, the local model, and actual denial of shell/both shell and bash aliases, web/network, nested task execution and external-directory access. If any configured deny rule is unsupported, refuse YOLO and use Sol or Luna instead.

`steps: 12` caps model iterations, not wall-clock runtime, number of file writes or OS resource consumption. Add an external process timeout and run in an isolated disposable container/VM without secrets and GitHub credentials. A worktree alone is NOT security isolation; it shares repository metadata and often host credentials.

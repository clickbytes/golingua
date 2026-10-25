# DSE execution runbook

This runbook is for the coordinator running the serial DSE issue plan. It is a procedure, not an installed workflow or automation. The issue hierarchy, dependencies, and reserved branch names live in [`issue-map.json`](issue-map.json); the formal and compatibility rules live in the [semantic-algebra RFC index](../../architecture/semantic-algebra/README.md). Root [`AGENTS.md`](../../../AGENTS.md) and current repository policy take precedence.

## Authority and current record

- The coordinator owns issue activation, scope, branch/PR, integration, verification, merge authorization, and status updates. Helpers never write to GitHub.
- At most one DSE leaf is `ACTIVE`; at most one implementation writer edits the active checkout. Read-only review may run in parallel. A chapter's status is aggregate only and does not authorize another leaf.
- Parent links/checklists and the issue map are the hierarchy record. GitHub native Sub-issue or Development links are supplementary and may be recorded only after actually verifying them. Their absence does not block work and must not be represented as a link.
- A reserved `branch_name` is not a created branch. `branch_url` is populated only after the branch has actually been created or otherwise verified.
- A docs-only leaf has no runtime behavior to demonstrate. Runtime is `NOT RUN`, not `PASS`; that fact is not a runtime defect for a docs-only scope. Implementation leaves must meet their own observed CLI criteria.
- Keep technical status, formal/mathematical assessment, agent language assessment, and owner acceptance separate. Never infer owner approval from a merge or a green CI.

Current recorded state: #377 is `DONE` with actual base `29370224ba8e29d3471ff02db200373f40e130a0`, PR [#401](https://github.com/clickbytes/golingua/pull/401), merge `bd6b4d24a8a16f348672a98a4d1a9b09fb4e1dc3`, and three RFC files as evidence. #378 is the sole `ACTIVE` leaf, based on `46f40097c238912ede3a64c004782b7e103307fc`, on `dse/01-02-execution-runbook`; its PR and merge are not yet known. These values are an example of the receipt format, not defaults for later leaves.

## START gate

The coordinator performs every check before changing an issue to `ACTIVE`, creating a branch, or assigning a writer:

1. Read current `AGENTS.md`, current issue, linked RFCs/contracts, and its dependency rows in the issue map. Confirm the map has 9 chapter records and the expected leaf set; dependencies resolve and form an acyclic graph. Do not silently repair a mismatch during an implementation task.
2. Read the current `main` SHA, open PRs, and the current DSE epic/chapter/leaf states. Check for any conflicting `ACTIVE` epic/leaf, writer, or implementation PR. A chapter aggregate such as `IN_PROGRESS` is not itself a conflict; a second active leaf is.
3. For every dependency, inspect the actual merged commit on current `main`, successful required CI, and the predecessor's completion receipt. A `CLOSED` issue alone, an open PR, or a branch tip does not satisfy a dependency. Confirm the merged code and evidence match the receipt.
4. Confirm there is one authorized issue and one executable next leaf in the serial order. Check branch ownership/history before using a reserved name. Never reset or force-push a pre-existing branch to make it fit the reservation.
5. Record the actual `main` SHA as `actual_base_sha`; create the assigned branch from that fresh main and verify its URL. A base from an epic plan or an old issue is historical context, not a substitute.
6. Post the issue START comment with the coordinator, one writer, optional read-only reviewer, exact whitelist, concrete positive and negative observable results, actual base SHA, branch URL, verification commands, and known constraints. Then set only that leaf `ACTIVE`; chapters/epic remain aggregates.

### Paper START simulations

- **Second active leaf:** if another DSE leaf is already `ACTIVE`, refuse to start the requested leaf. Report the conflicting leaf and wait for coordinator resolution; do not mark both active or assume the chapter aggregate grants an exception.
- **Dependency has not merged:** if a required issue is `CLOSED` but its PR is open, merge is absent from current `main`, CI is not successful, or predecessor receipt is missing, refuse START. Do not build on its branch or treat closure text as proof.
- **Native links unavailable:** if Sub-issue/Development links cannot be inspected or are absent, continue using issue parent/checklist and map. Record native links as unverified/absent; do not invent them or block the valid dependency DAG.

## DELEGATE and write isolation

Issue work may be delegated only as a small, closed job. The coordinator remains responsible for scope and acceptance. One worker writes at a time; reviewers are read-only. The JOB BRIEF must state:

```text
Issue / leaf and base SHA:
Exact assigned files and symbols:
Inputs and current contracts to read:
Positive observable result:
Negative/refusal result:
Forbidden paths/actions (including GitHub writes, commits, branch/PR creation):
Allowed verification and work/time limit:
Required handoff (diff, commands, actual outputs, risks, unresolved items):
```

Before any OpenCode delegation, follow the current `AGENTS.md` gate: use the daily review for the same environment/work type only while its model, cost, version, and configuration remain current; select only a currently available free model. For each job, create a deliberate `opencode.json` in an isolated disposable snapshot, cap the agent at 12 steps, grant only assigned-path read/write permissions, and verify the effective config and agent in the worker's actual environment. Probe actual denials for shell, network, subagents, and access outside the workspace before `--auto`; a configured `deny` field alone is not evidence. Use an external timeout and a container/VM without `.git` or credentials. Stop the delegation if isolation, model, config, or refusal behavior cannot be verified; do not bypass it with host YOLO.

GPT Luna Low may perform bounded implementation or read-only review when allowed by current policy, including as a fallback when OpenCode is unavailable. It is not a GitHub writer. Do not run helpers merely to fill a concurrency allowance. Preserve all changes outside the exact assignment and report unexpected paths instead of reverting other work.

## REVIEW and VERIFY

The coordinator reviews the complete diff against the assigned whitelist and checks that claims in the handoff match the actual changes. Reviewers may report findings but do not change scope or approve language data on the owner's behalf. Allow at most the authorized first pass and one named correction; further failure requires a bounded coordinator decision to integrate a supplied fix or report `PARTIAL`/`BLOCKED`, not an automatic retry loop.

For changed Go, run the current required checks: `gofmt`, `go vet ./...`, `go build ./cmd/golingua`, help/version smoke, issue-specific existing checks, and real CLI observations. Do not add a new test runner, gold, percentage gate, JS parity check, or `*_test.go` unless current owner policy and active issue explicitly authorize that exact test. For docs-only changes, check links, formatting/whitespace, scope, and current CI on the PR; do not report a product CLI result as required or run it merely to imply runtime coverage.

Record each actual command, input, stdout/stderr, exit code, code SHA, relevant artifact hash, and limitations. Use `PASS`, `FAIL`, `INFRA`, or `NOT RUN` accurately. `NOT RUN` is never `PASS`. A required runtime flow not observed makes an implementation leaf `PARTIAL`/`BLOCKED` unless the owner explicitly waived that criterion. Docs-only runtime is marked `NOT RUN` because it is inapplicable, separately from its documentation checks.

The worker handoff contains:

- changed paths and concise diff summary;
- actual commands and outputs, not reconstructed or assumed output;
- unresolved requirements, risks, and exact reasons for any `NOT RUN`;
- no GitHub actions, branch creation, commit, or merge claimed by a helper.

## PR, merge, and receipts

The coordinator owns branch updates, commits, one PR per leaf, PR review, CI inspection, and the final merge decision. Default PR body: `Closes #<leaf>` and `Refs #<chapter>, #367`, with concise scope and evidence. A merge occurs only after coordinator review, required actual PR CI, and current explicit merge authorization; adding a checklist or getting a green check never auto-merges.

After merge, verify the merge commit is on current `main`, verify required main CI, and inspect the merged paths. Then update the leaf's receipt and aggregate chapter/epic status. Do not claim a receipt before verifying it.

### Issue-map snapshot timing

The map is committed as part of the leaf's PR, before that PR has a merge SHA. Therefore the active leaf's own map row cannot know its future merge SHA. It must keep `merge_sha: null` and `pr_url: null` until those facts exist; never guess or backfill them directly on `main` outside the next authorized work.

When the next leaf is authorized, its branch may update the predecessor's map row using the predecessor's verified GitHub issue completion receipt and current-main/CI check. The predecessor receipt records the actual PR, merge SHA, evidence paths, and limitations. The final #400 row follows the same rule: the prior map snapshot leaves its own merge SHA null; the canonical post-merge self-receipt is the #400 GitHub issue after final main/CI verification. Do not create an unplanned direct-main commit solely to fill the final self-reference. If no later authorized leaf exists to refresh an earlier map row, preserve the null and rely on its verified issue receipt rather than fabricate a value.

### Normal and final timelines

```text
Normal leaf:
START checks → branch + issue ACTIVE → one writer → review/verify → PR + actual CI
→ coordinator-authorized merge → verify main/CI → issue completion receipt
→ next authorized leaf may record predecessor receipt in its map update

Final leaf #400:
START checks → branch + issue ACTIVE → review/verify → PR + actual CI
→ coordinator-authorized merge → verify main/CI → #400 issue becomes canonical
  post-merge receipt; its pre-merge map row keeps its own merge_sha null
```

If a required check fails, distinguish a code, specification, environment, external, or out-of-scope blocker. Stop at the current gate and report it. Do not activate the next leaf automatically.

## END statuses

- **DONE:** every active-leaf acceptance criterion has evidence; required checks passed; receipt and current main state are accurate. Owner/language acceptance may still be `NOT_EVALUATED` and must remain separately labeled.
- **PARTIAL:** some work is useful, but a criterion or required observation is incomplete. Name it and its current status.
- **BLOCKED:** a specific dependency, permission, environment, contract, or review issue prevents required progress. Include what was attempted and the evidence.
- `NOT RUN` stays visible. A technical merge does not turn language acceptance into approval, and a docs-only runbook does not claim a runtime pass.

## Ready-to-copy issue templates

### START comment

```text
START — #<leaf> / <code>
Coordinator: <name>; writer: <one worker>; read-only reviewer: <optional>
Base: <verified main SHA> | branch: <created branch URL>
Dependencies: <each merged main SHA + CI + predecessor receipt>
Whitelist: <exact paths/symbols>
Positive result: <observable output/criterion>
Negative result: <refusal or preserved boundary>
Checks: <actual commands planned for this scope>
Limits: <runtime NOT RUN if docs-only; known acceptance gaps>
```

### Completion receipt

```text
Receipt — #<leaf> / <code>
Status: DONE | PARTIAL | BLOCKED
Base: <actual base SHA> | merged main SHA: <actual merge SHA, after verification>
PR: <actual URL> | CI: <actual checks and results>
Evidence: <paths/commands/outputs>
NOT RUN / limitations: <explicit list>
Agent technical/language assessment: <separate result>
Owner acceptance: NOT_EVALUATED | <actual owner decision only>
```

For #400, post this receipt on the issue after main verification; do not write an unverified self-merge SHA into the prior map snapshot.

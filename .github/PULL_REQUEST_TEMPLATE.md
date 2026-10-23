## Authority / issue scope

Closes #

- Parent epic and one ACTIVE child:
- Sole coordinator: Codex Sol Medium
- Base `main` SHA:
- Exact changed files and symbols:
- Positive observable outcome:
- Negative behavior / unsupported / remaining:
- Contracts and docs explicitly authorized to change:
- New tests authorized by this exact issue? NO / YES (cite)
- Product v2/migrations, JS parity and unrelated refactors avoided? YES / NO

## Delegation and boundaries

- Worker(s): NONE / Luna Low scout / Luna Low implementer / OpenCode YOLO
- Bounded JOB BRIEF / permitted paths:
- At most one writer at a time? YES / NO
- Worker handoff: DONE / PARTIAL / BLOCKED
- Sol inspected whole diff and newly created files? YES / NO
- Out-of-scope changes: NONE / explained

## Verification (actual observations)

- `gofmt`: NOT RUN
- `go vet ./...`: NOT RUN
- `go build ./cmd/golingua`: NOT RUN
- Existing CLI smoke: NOT RUN
- Mandatory issue-specific CLI/TTY/manual observations: NOT RUN / evidence
- GitHub `go-ci` run on PR head: PENDING (link)
- Owner linguistic acceptance: NOT EVALUATED (distinct from agent technical review)

## Final review by Sol

- [ ] Every mandatory acceptance criterion has evidence or is explicitly waived by owner; required NOT RUN is not DONE.
- [ ] No queued issue, speculative v2, migration, unrequested tests/docs or unrelated change.
- [ ] Reviewer findings classified and blockers resolved.
- [ ] Full diff, required checks and actual CI reviewed before merge.
- [ ] After merge, `main` SHA and CI confirmed before activating next child.

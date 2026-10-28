# Dynamic Semantic Expressions (DSE)

DSE is the owner-authorized, bounded work plan for domain-independent semantic expressions. This page is the entry point to its issue record, execution procedure, and formal contracts. It does not authorize a product redesign, hypothetical v2, migration, extra tests, or work outside the active leaf.

## Working documents

- [Issue map](issue-map.json): epic, chapter and leaf identities, dependencies, reserved branch names, current states, and verified execution receipts.
- [Execution runbook](execution.md): serial START gates, delegation, review, verification, PR/merge, and completion receipts.
- [Semantic algebra RFC index](../../architecture/semantic-algebra/README.md): [formal contract](../../architecture/semantic-algebra/formal-contract.md) and [Meaning v1 compatibility](../../architecture/semantic-algebra/compatibility.md).
- [Epic #367](https://github.com/clickbytes/golingua/issues/367): canonical issue hierarchy and sequence. Each leaf issue remains authoritative for its own implementation scope.

## Plan and current snapshot

The plan has nine aggregate chapters, #368–#376, and 24 executable leaves, #377–#400. The recommended work order is serial: #377–#378; #379–#385; #386–#390; #391–#393; #394–#398; then #399–#400. Dependencies listed on each leaf remain the actual START gate; the grouped order is not a substitute for checking them on current `main`.

At the #400 acceptance snapshot, leaves #377–#399 have merged results and evidence; #400 is the active acceptance leaf. The current register is [issue-map.json](issue-map.json). [Domain-independence audit](independence-audit.md), [results](results.md), [acceptance](acceptance.md), and [final observations](evidence/09-02/README.md) describe the supported profile and its limits. Owner language acceptance remains **NOT EVALUATED**.

The issue map is a pre-merge snapshot: an active leaf cannot know its own future PR or merge SHA. Keep unknown values null. The verified post-merge completion receipt belongs on that leaf's GitHub issue; a later authorized leaf may record that predecessor receipt in its map update. For final #400, its issue is the canonical post-merge self-receipt. Do not invent a SHA or make an unplanned direct-main backfill just to fill a self-reference.

## Relationship to earlier architecture chapters

The epic maps the earlier project outline onto the nine DSE chapters as follows:

| DSE chapter | Earlier chapter mapping |
| --- | --- |
| DSE-01 — formal contract and execution protocol | 1–3 |
| DSE-02 — metadata, explicit facts, catalog | 5 |
| DSE-03 — typed IR and generic operators | 3, 4, 6, 13 |
| DSE-04 — runtime synthesis | new runtime synthesis work |
| DSE-05 — normalization, execution, evidence | 9, 10 |
| DSE-06 — scope, questions, context | 7, 8 |
| DSE-07 — declarative lexicon and PL/EN analysis | 11 and part of 12 |
| DSE-08 — realization and production integration | 12, 14 |
| DSE-09 — domain independence and acceptance | 15, 16 |

The mapping describes organization; it does not reopen completed scopes or silently transfer old requirements. Active issue, current owner decisions, and root `AGENTS.md` control each change.

## Policy boundary

Follow the repository's current serial-work and delegation policy. One active DSE leaf and one implementation writer at a time; helpers do not write GitHub. Use exact assigned paths, current base SHA, observable positive and negative outcomes, and actual checks. The DSE docs do not create a test runner, golds, percentage gate, JS parity requirement, broad corpus, background automation, or new delegation/cost authority. Existing tests policy and explicit owner decisions continue to apply.

Keep runtime/technical completion, agent assessment, and owner language acceptance distinct. A docs-only task records runtime as `NOT RUN` because it is inapplicable; an implementation task must meet its issue's observable CLI requirements before it can be reported `DONE`.

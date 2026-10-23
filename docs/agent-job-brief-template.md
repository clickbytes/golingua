# Delegated worker job brief (copy and fill for each worker)

JOB_ID: <issue>-W1
COORDINATOR: Codex Sol Medium
WORKER: Codex Luna Low / OpenCode YOLO
ACTIVE_ISSUE: #<number> (must be OPEN + ACTIVE)
PARENT_EPIC: #<number>
BASE_SHA: <full SHA>
ASSIGNED_SCOPE: <one small result from the current issue>
ALLOWED_PATHS: <explicit existing file paths or narrowly bounded directory globs>
RELEVANT_SYMBOLS: <functions/types>
REFERENCE_DOCS: <only canonical docs linked from the issue>
ACCEPTANCE: <concrete positive and negative observable behavior>
NON_GOALS: <out-of-scope behavior, no future issue implementation>
TEST_POLICY: <NO NEW TESTS unless exact tests in the issue / authorized owner decision>
DO_NOT_CHANGE: GitHub, epics, issues, AGENTS.md, migration plans, product v2, unrelated files
MAX_PASSES: 1 implementation + 1 correction after Sol review (correction requires new brief)
STEP_BUDGET: 12 for OpenCode; 5 targeted lookups for read-only Luna scout
WORKSPACE: <isolated scratch for external worker; shared checkout only with exclusive writer lease>

Return only:
STATUS: DONE | PARTIAL | BLOCKED
FILES_CHANGED: ...
EVIDENCE: exact observations; PASS / FAIL / INFRA / NOT RUN
OPEN_RISKS: ...
STOP: do not continue or spawn further agents

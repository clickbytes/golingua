# Latest language matrix result

> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.

- Source SHA: `937f81eadf6af8dddbd93fd9a5154754b4f121a4`
- Generated (UTC): `2026-10-25T07:13:58Z`
- Command: `go test ./tests/... -count=1 -v`
- Command exit code: `1` (NONZERO)
- Registered records: 24; evaluated checks: 72

| Case ID | Direction | Analysis | Realization | Integration |
|---|---|---|---|---|
| `A1-AC-001-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-AC-002-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-AC-003-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-AC-004-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-ID-001-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-ID-002-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-ID-003-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-ID-004-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-LQ-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-LQ-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-LQ-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-LQ-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-NEG-001-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-NEG-002-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-NEG-003-EN-PL` | EN → PL | PASS | PASS | PASS |
| `A1-NEG-004-PL-EN` | PL → EN | PASS | PASS | PASS |
| `A1-PO-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-PO-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-PO-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-PO-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-Q-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-Q-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-Q-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |
| `A1-Q-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING |

## Counts

PASS: 36 | FAIL: 0 | GAP_ANALYZE: 0 | GAP_REALIZE: 0 | GAP_I: 0 | SPEC_PENDING: 36 | INFRA: 0 | NOT RUN: 0

`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching check result was observed. A nonzero command exit may also include failures outside these 24 sample records; consult the complete raw log artifact for every run detail.

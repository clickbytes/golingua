# #344 observational report results

Status: report implementation evidence only; owner acceptance remains `NOT_EVALUATED`.

The reporter retains the 1,000 declared directions and all 3,000 analysis/realization/integration check outcomes, plus 205 extra leaf checks, package summaries, and 32 separate refusal controls. Group counts report 500 unique PairIDs against planning targets; the targets are not pass criteria. It parses Go structured metadata by explicit marker and raw JSON decoding, including source semantics, declared IDs, tags, owner status, pack identity, and sense-to-lexeme catalogue associations. Catalogue associations are explicitly not morphology execution evidence.

For the final verbose log (`/tmp/golingua-matrix-expansion/m8-final-full.log`, command exit 1), adapter observations include 1,004 native-analysis calls, 1,000 integration calls, 746 executed independent realizations, and 362 independent realizations explicitly not executed with recorded reasons and IDs. Source gap probes and context source analyses remain separate stages. The report separately summarizes authored `meaning_input`, observed source/result graphs, trace family/variant/frame/outcome, observed feature triples, and case-linked diagnostics for declared representation gaps. Check NOT RUN and adapter NOT RUN remain separate.

The full generated report was written outside the repository to `/tmp/golingua-matrix-expansion/m8-report.md` for review. Its corrected check totals are PASS 297, FAIL 244, GAP_ANALYZE 946, GAP_REALIZE 433, and GAP_I 1,080, with SPEC_PENDING/INFRA/NOT RUN all zero; the command exit remains 1. It retains all 32 refusal checks as PASS and 205 other extra leaf checks. The full raw log remains the source for individual runtime detail.

## Validation

- `ast.parse` on `scripts/report-language-matrix.py`: PASS; no bytecode/cache emitted.
- Full report invocation using `--sha f11303577966a913709ac3008d99cbf4fab29c7b --timestamp 2026-10-25T12:00:00Z --exit-code 1 --log /tmp/golingua-matrix-expansion/m8-final-full.log`: PASS. Parsed 1,000/1,000 inventory rows, one snapshot, 2,023 catalogue association rows, and 149 direct `result=` JSON objects with zero parse issues. The rendered matrix has all 1,000 rows; check outcomes total 3,000. The test command itself returned exit 1, preserved in the report.
- Empty-log probe: PASS; retained all 1,000 source rows as NOT RUN, with zero observed graphs and zero inventory metadata records.
- Truncated-log probe: PASS; retained all 1,000 source rows, absent check results remained NOT RUN, and no observed graphs were fabricated.
- Priority probe with GAP_ANALYZE then GAP_I and later INFRA: PASS; final classification is INFRA. GAP_ANALYZE then GAP_I classifies as GAP_I. Inventory prose containing GAP_I or NOT RUN does not affect a Go PASS. The full log still reports TestInventory as PASS.
- Owner acceptance: `NOT_EVALUATED`.

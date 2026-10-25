#!/usr/bin/env python3
"""Render the latest registered PL/EN matrix checks from `go test -v` output."""

from __future__ import annotations

import argparse
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path
import re
import sys


CASE_RE = re.compile(r"A1-[A-Z]+-\d{3}-(?:EN-PL|PL-EN)")
REGISTERED_RE = re.compile(r'ID:\s*"(A1-[A-Z]+-\d{3}-(?:EN-PL|PL-EN))"')
RUN_RE = re.compile(r"^=== RUN\s+(\S+)")
RESULT_RE = re.compile(r"^\s*--- (PASS|FAIL):\s+(\S+)")


def registered_cases(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    found = sorted(set(REGISTERED_RE.findall(text)))
    if not found:
        raise ValueError(f"no stable A1 IDs found in {path}")
    return found


def check_kind(test_name: str) -> str | None:
    parent = test_name.split("/", 1)[0]
    if parent.endswith("Analysis"):
        return "analysis"
    if parent.endswith("Realization"):
        return "realization"
    if parent in {"TestSeeing", "TestIdentification", "TestNegation", "TestOtherDomains"}:
        return "integration"
    return None


def observed_status(go_result: str, lines: list[str]) -> str:
    content = "\n".join(lines)
    if "INFRA:" in content:
        return "INFRA"
    if "SPEC_PENDING" in content or "NOT RUN" in content:
        return "SPEC_PENDING"
    for category in ("GAP_ANALYZE", "GAP_REALIZE", "GAP_I"):
        if category in content:
            return category
    if re.search(r"(?:^|\W)FAIL:", content):
        return "FAIL"
    return "PASS" if go_result == "PASS" else "FAIL"


def parse_checks(log_text: str) -> dict[tuple[str, str], list[str]]:
    observations: dict[tuple[str, str], list[str]] = defaultdict(list)
    details: dict[str, list[str]] = defaultdict(list)
    current_name: str | None = None

    for line in log_text.splitlines():
        run_match = RUN_RE.match(line)
        if run_match:
            current_name = run_match.group(1) if CASE_RE.search(run_match.group(1)) else None
            continue

        if current_name is not None:
            details[current_name].append(line)

        result_match = RESULT_RE.match(line)
        if not result_match:
            continue
        result, test_name = result_match.groups()
        case_match = CASE_RE.search(test_name)
        kind = check_kind(test_name)
        if case_match and kind:
            # The name also ties the output markers to this exact subtest.
            body = details[test_name]
            observations[(case_match.group(0), kind)].append(observed_status(result, body))
        if current_name == test_name:
            current_name = None

    return observations


def render(sha: str, timestamp: str, exit_code: int, case_file: Path, log_file: Path) -> str:
    cases = registered_cases(case_file)
    observations = parse_checks(log_file.read_text(encoding="utf-8", errors="replace"))
    kinds = ("analysis", "realization", "integration")
    rows: list[tuple[str, list[str]]] = []
    counts: Counter[str] = Counter()

    for case_id in cases:
        statuses: list[str] = []
        for kind in kinds:
            seen = observations.get((case_id, kind), [])
            if len(seen) == 1:
                status = seen[0]
            elif not seen:
                status = "NOT RUN"
            else:
                status = "FAIL"
            statuses.append(status)
            counts[status] += 1
        direction = case_id.rsplit("-", 2)[-2] + " → " + case_id.rsplit("-", 1)[-1]
        rows.append((case_id, [direction, *statuses]))

    lines = [
        "# Latest language matrix result",
        "",
        "> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.",
        "",
        f"- Source SHA: `{sha}`",
        f"- Generated (UTC): `{timestamp}`",
        "- Command: `go test ./tests/... -count=1 -v`",
        f"- Command exit code: `{exit_code}` ({'PASS' if exit_code == 0 else 'NONZERO'})",
        f"- Registered records: {len(cases)}; evaluated checks: {len(cases) * len(kinds)}",
        "",
        "| Case ID | Direction | Analysis | Realization | Integration |",
        "|---|---|---|---|---|",
    ]
    for case_id, values in rows:
        lines.append(f"| `{case_id}` | {values[0]} | {values[1]} | {values[2]} | {values[3]} |")

    lines.extend(
        [
            "",
            "## Counts",
            "",
            " | ".join(f"{name}: {counts[name]}" for name in ("PASS", "FAIL", "GAP_ANALYZE", "GAP_REALIZE", "GAP_I", "SPEC_PENDING", "INFRA", "NOT RUN")),
            "",
            "`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching check result was observed. A nonzero command exit may also include failures outside these 24 sample records; consult the complete raw log artifact for every run detail.",
            "",
        ]
    )
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sha", required=True, help="source commit SHA")
    parser.add_argument("--exit-code", required=True, type=int, help="go test exit code")
    parser.add_argument("--timestamp", help="UTC timestamp; defaults to current UTC time")
    parser.add_argument("--cases", type=Path, default=Path("tests/translation/cases.go"))
    parser.add_argument("--log", required=True, type=Path, help="complete verbose go test log")
    args = parser.parse_args()
    timestamp = args.timestamp or datetime.now(timezone.utc).replace(microsecond=0).isoformat()
    try:
        sys.stdout.write(render(args.sha, timestamp, args.exit_code, args.cases, args.log))
    except (OSError, ValueError) as exc:
        parser.error(str(exc))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

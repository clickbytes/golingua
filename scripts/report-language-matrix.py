#!/usr/bin/env python3
"""Render declared PL/EN records and all observed checks from `go test -v`."""

from __future__ import annotations

import argparse
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path
import re
import sys


ID_PATTERN = r"[ABC][0-9]+(?:\.[0-9]+)*-[A-Z][A-Z0-9]*-[0-9]+-(?:EN-PL|PL-EN)"
CASE_RE = re.compile(ID_PATTERN)
REGISTERED_RE = re.compile(r'ID:\s*"(' + ID_PATTERN + r')"')
TAGS_RE = re.compile(r'CoverageTags:\s*\[\]string\{([^}]*)\},\s*ID:\s*"(' + ID_PATTERN + r')"')
RUN_RE = re.compile(r"^=== RUN\s+(\S+)")
RESULT_RE = re.compile(r"^\s*--- (PASS|FAIL):\s+(\S+)")
PACKAGE_RE = re.compile(r"^(?:ok|FAIL)\s+(github\.com/\S+)", re.MULTILINE)


def registered_cases(path: Path) -> list[str]:
    found = sorted(set(REGISTERED_RE.findall(path.read_text(encoding="utf-8"))))
    if not found:
        raise ValueError(f"no stable language matrix IDs found in {path}")
    return found


def declared_coverage(path: Path) -> dict[str, list[str]]:
    coverage: dict[str, list[str]] = defaultdict(list)
    for tags, case_id in TAGS_RE.findall(path.read_text(encoding="utf-8")):
        coverage[case_id].extend(re.findall(r'"([^"]+)"', tags))
    return coverage


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


def parse_checks(log_text: str) -> tuple[dict[tuple[str, str], list[str]], list[tuple[str, str, str]], list[tuple[str, str, int]]]:
    observations: dict[tuple[str, str], list[str]] = defaultdict(list)
    extras: list[tuple[str, str, str]] = []
    package_summaries: list[tuple[str, str, int]] = []
    # `go test -v` buffers output by package, then writes its ok/FAIL summary.
    # Keeping that boundary makes identically named tests in different packages distinct.
    chunks: list[tuple[str, list[str]]] = []
    pending: list[str] = []
    for line in log_text.splitlines():
        pending.append(line)
        match = PACKAGE_RE.match(line)
        if match:
            chunks.append((match.group(1), pending))
            pending = []
    if pending:
        chunks.append(("(package summary not observed)", pending))

    for package, lines in chunks:
        summary_line = next((line for line in reversed(lines) if PACKAGE_RE.match(line)), "")
        summary = "FAIL" if summary_line.startswith("FAIL") else "PASS" if summary_line.startswith("ok") else "NOT RUN"
        details: dict[str, list[str]] = defaultdict(list)
        current_name: str | None = None
        results: list[tuple[str, str]] = []
        for line in lines:
            run_match = RUN_RE.match(line)
            if run_match:
                current_name = run_match.group(1)
                continue
            if current_name is not None:
                details[current_name].append(line)
            result_match = RESULT_RE.match(line)
            if result_match:
                result, test_name = result_match.groups()
                results.append((result, test_name))
                if current_name == test_name:
                    current_name = None
        run_names = {match.group(1) for line in lines if (match := RUN_RE.match(line))}
        package_summaries.append((package, summary, len(run_names)))
        for result, test_name in results:
            if any(name.startswith(test_name + "/") for name in run_names):
                continue  # Report leaf checks; parent summaries duplicate their children.
            case_match = CASE_RE.search(test_name)
            kind = check_kind(test_name)
            body = details[test_name]
            state = observed_status(result, body)
            if case_match and kind:
                observations[(case_match.group(0), kind)].append(state)
            elif "/" in test_name or test_name.startswith("Test"):
                extras.append((package, test_name, state))
    return observations, extras, package_summaries


def render(sha: str, timestamp: str, exit_code: int, case_file: Path, log_file: Path) -> str:
    cases = registered_cases(case_file)
    coverage = declared_coverage(case_file)
    observations, extras, package_summaries = parse_checks(log_file.read_text(encoding="utf-8", errors="replace"))
    kinds = ("analysis", "realization", "integration")
    rows: list[tuple[str, list[str]]] = []
    counts: Counter[str] = Counter()

    for case_id in cases:
        statuses: list[str] = []
        for kind in kinds:
            seen = observations.get((case_id, kind), [])
            if len(seen) == 1:
                state = seen[0]
            elif not seen:
                state = "NOT RUN"
            else:
                state = "FAIL"
            statuses.append(state)
            counts[state] += 1
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
        f"- Registered records: {len(cases)}; record checks: {len(cases) * len(kinds)}",
        "- Additional observed checks: " + str(len(extras)),
        "",
        "| Case ID | Direction | Analysis | Realization | Integration | Declared coverage tags |",
        "|---|---|---|---|---|---|",
    ]
    for case_id, values in rows:
        tags = ", ".join(sorted(set(coverage.get(case_id, [])))) or "(none)"
        lines.append(f"| `{case_id}` | {values[0]} | {values[1]} | {values[2]} | {values[3]} | {tags} |")

    lines.extend(["", "## Additional observed checks", "", "| Package | Check | Status |", "|---|---|---|"])
    for package, test_name, state in extras:
        lines.append(f"| `{package}` | `{test_name}` | {state} |")

    lines.extend(["", "## Package summaries", "", "Packages with no observed test run are reported as infrastructure/build failures when Go returned `FAIL`; summaries do not replace per-check outcomes.", "", "| Package | Go summary | Observed test names |", "|---|---|---|"])
    for package, summary, run_count in package_summaries:
        state = "INFRA (no tests ran)" if summary == "FAIL" and run_count == 0 else summary
        lines.append(f"| `{package}` | {state} | {run_count} |")

    tag_counts: Counter[str] = Counter(tag for values in coverage.values() for tag in values)
    lines.extend(["", "## Declared coverage", "", "Declared tags are metadata attached to records; they do not prove that a construction executed successfully.", ""])
    lines.append(" | ".join(f"{tag}: {tag_counts[tag]} records" for tag in sorted(tag_counts)) or "No coverage tags declared.")
    lines.extend([
        "", "## Counts", "",
        " | ".join(f"{name}: {counts[name]}" for name in ("PASS", "FAIL", "GAP_ANALYZE", "GAP_REALIZE", "GAP_I", "SPEC_PENDING", "INFRA", "NOT RUN")),
        "",
        "`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching record check result was observed. Additional checks retain their own observed statuses above. Consult the full raw log artifact for every run detail.",
        "",
    ])
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

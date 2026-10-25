#!/usr/bin/env python3
"""Render declared PL/EN records and observed checks from `go test -v`."""

from __future__ import annotations

import argparse
from collections import Counter, defaultdict
from datetime import datetime, timezone
import json
from pathlib import Path
import re
import sys
from typing import Any, Iterable


ID_PATTERN = r"[ABC][0-9]+(?:\.[0-9]+)*-[A-Z][A-Z0-9]*-[0-9]+-(?:EN-PL|PL-EN)"
CASE_RE = re.compile(ID_PATTERN)
REGISTERED_RE = re.compile(r'ID:\s*"(' + ID_PATTERN + r')"')
CASE_META_RE = re.compile(
    r'\{PairID:\s*"([^"]+)",\s*PrimaryGroup:\s*"([^"]+)",\s*'
    r'CoverageTags:\s*\[\]string\{([^}]*)\},\s*ID:\s*"(' + ID_PATTERN + r')",\s*From:\s*"([a-z]+)",\s*To:\s*"([a-z]+)"'
)
RUN_RE = re.compile(r"^=== RUN\s+(\S+)")
RESULT_RE = re.compile(r"^\s*--- (PASS|FAIL):\s+(\S+)")
PACKAGE_RE = re.compile(r"^(?:ok|FAIL)\s+(github\.com/\S+)", re.MULTILINE)
JSON_DECODER = json.JSONDecoder()


def registered_cases(path: Path) -> list[str]:
    found = sorted(set(REGISTERED_RE.findall(path.read_text(encoding="utf-8"))))
    if not found:
        raise ValueError(f"no stable language matrix IDs found in {path}")
    return found


def declared_metadata(path: Path) -> dict[str, dict[str, Any]]:
    text = path.read_text(encoding="utf-8")
    result: dict[str, dict[str, Any]] = {}
    for pair_id, group, tags, case_id, source, target in CASE_META_RE.findall(text):
        result[case_id] = {
            "pair_id": pair_id,
            "primary_group": group,
            "coverage_tags": re.findall(r'"([^"]+)"', tags),
            "from": source,
            "to": target,
        }
    return result


def declared_coverage(path: Path) -> dict[str, list[str]]:
    return {case_id: value["coverage_tags"] for case_id, value in declared_metadata(path).items()}


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
    # Only explicit diagnostic prefixes indicate a check status. Inventory JSON,
    # semantic prose, and runtime adapter NOT RUN markers are not check failures.
    found: set[str] = set()
    for line in lines:
        message = re.match(r"^\s*[^\s].*?\.go:\d+:\s*(.*)$", line)
        text = message.group(1) if message else ""
        if text.startswith("INFRA:"):
            found.add("INFRA")
        elif text.startswith(("SPEC_PENDING:", "SPEC_PENDING / NOT RUN:")):
            found.add("SPEC_PENDING")
        elif text.startswith("GAP_I:"):
            found.add("GAP_I")
        elif text.startswith("GAP_ANALYZE:"):
            found.add("GAP_ANALYZE")
        elif text.startswith("GAP_REALIZE:"):
            found.add("GAP_REALIZE")
        elif text.startswith("FAIL:"):
            found.add("FAIL")
    for status in ("INFRA", "SPEC_PENDING", "GAP_I", "GAP_ANALYZE", "GAP_REALIZE", "FAIL"):
        if status in found:
            return status
    return "PASS" if go_result == "PASS" else "FAIL"


def parse_checks(log_text: str) -> tuple[dict[tuple[str, str], list[str]], list[tuple[str, str, str]], list[tuple[str, str, int]]]:
    observations: dict[tuple[str, str], list[str]] = defaultdict(list)
    extras: list[tuple[str, str, str]] = []
    package_summaries: list[tuple[str, str, int]] = []
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
                continue
            case_match = CASE_RE.search(test_name)
            kind = check_kind(test_name)
            state = observed_status(result, details[test_name])
            if case_match and kind:
                observations[(case_match.group(0), kind)].append(state)
            elif "/" in test_name or test_name.startswith("Test"):
                extras.append((package, test_name, state))
    return observations, extras, package_summaries


def marked_json_entries(log_text: str, marker: str) -> tuple[list[tuple[int, str, dict[str, Any]]], list[str]]:
    """Read JSON immediately after an explicit marker using raw_decode."""
    values: list[tuple[int, str, dict[str, Any]]] = []
    errors: list[str] = []
    for line_number, line in enumerate(log_text.splitlines(), 1):
        start = 0
        while True:
            index = line.find(marker, start)
            if index < 0:
                break
            payload = line[index + len(marker):].lstrip()
            try:
                value, consumed = JSON_DECODER.raw_decode(payload)
            except json.JSONDecodeError as exc:
                errors.append(f"line {line_number}: {marker} JSON: {exc.msg}")
                start = index + len(marker)
                continue
            if isinstance(value, dict):
                values.append((line_number, line, value))
            else:
                errors.append(f"line {line_number}: {marker} JSON is not an object")
            start = index + len(marker) + consumed
    return values, errors


def marked_json(log_text: str, marker: str) -> tuple[list[dict[str, Any]], list[str]]:
    entries, errors = marked_json_entries(log_text, marker)
    return [value for _line_number, _line, value in entries], errors


def strings(value: Any) -> list[str]:
    if isinstance(value, str):
        return [value]
    if isinstance(value, list):
        return [item for item in value if isinstance(item, str)]
    return []


def stage_key(observation: dict[str, Any]) -> str:
    return str(observation.get("stage") or "(stage not declared)")


def source_case_id(case_id: str) -> str:
    for suffix in (":analyze", ":realize"):
        if case_id.endswith(suffix):
            return case_id[:-len(suffix)]
    return case_id


def walk(value: Any) -> Iterable[dict[str, Any]]:
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from walk(child)
    elif isinstance(value, list):
        for child in value:
            yield from walk(child)


def graph_signature(graph: dict[str, Any]) -> str:
    return json.dumps(graph, sort_keys=True, ensure_ascii=False, separators=(",", ":"))


def collect_runtime(observations: list[dict[str, Any]]) -> dict[str, Any]:
    calls: Counter[tuple[str, str]] = Counter()
    false_realizations: dict[tuple[str, str], list[str]] = defaultdict(list)
    diagnostics_by_case: dict[str, set[str]] = defaultdict(set)
    authored_graphs: set[tuple[str, str, str, str]] = set()
    authored_node_kinds: Counter[str] = Counter()
    authored_roles: Counter[str] = Counter()
    authored_frames: Counter[str] = Counter()
    authored_multiplicities: Counter[str] = Counter()
    stage_outcomes: Counter[tuple[str, str]] = Counter()
    graphs: set[tuple[str, str, str]] = set()
    node_kinds: Counter[str] = Counter()
    predicates: Counter[tuple[str, str]] = Counter()
    roles: Counter[str] = Counter()
    frames: Counter[str] = Counter()
    multiplicities: Counter[str] = Counter()
    traces: set[tuple[str, str, str, str, str, str]] = set()
    feature_values: set[tuple[str, str, str, str, str]] = set()
    diagnostics: Counter[tuple[str, str]] = Counter()

    for observation in observations:
        stage = stage_key(observation)
        case_id = str(observation.get("case_id") or "(case id missing)")
        adapter = observation.get("adapter_executed")
        if adapter is not None or stage in {"native_analysis", "independent_realization", "integration"}:
            calls[(stage, str(adapter).lower() if adapter is not None else "unspecified")] += 1
        result = observation.get("result")
        if isinstance(result, dict):
            result_outcome = result.get("outcome", result.get("status", result.get("Outcome", "unknown")))
            stage_outcomes[(stage, str(result_outcome))] += 1
        if stage == "independent_realization" and observation.get("independent_realization_executed") is False:
            reason = str(observation.get("reason") or "reason not recorded")
            false_realizations[(reason, stage)].append(case_id)
        meaning_input = observation.get("meaning_input")
        meaning_id = str(observation.get("meaning_id") or "(meaning id missing)")
        if isinstance(meaning_input, dict) and isinstance(meaning_input.get("nodes"), list):
            signature = graph_signature(meaning_input)
            authored_key = (case_id, stage, meaning_id, signature)
            if authored_key not in authored_graphs:
                authored_graphs.add(authored_key)
                for node in meaning_input["nodes"]:
                    if not isinstance(node, dict):
                        continue
                    authored_node_kinds[str(node.get("kind") or "unknown")] += 1
                    authored_frames[str(node.get("temporal_frame") or "unknown")] += 1
                    if node.get("kind") == "referent":
                        authored_multiplicities[str(node.get("multiplicity") or "unknown")] += 1
                    for participant in node.get("participants", []) or []:
                        if isinstance(participant, dict):
                            authored_roles[str(participant.get("role") or "unknown")] += 1
        if not isinstance(result, (dict, list)):
            continue
        result_dict = result if isinstance(result, dict) else {"items": result}
        for obj in walk(result_dict):
            # Only Meaning graphs inside observed results; authored meaning_input is never traversed.
            if isinstance(obj.get("nodes"), list) and isinstance(obj.get("roots"), list):
                signature = graph_signature(obj)
                graph_key = (case_id, stage, signature)
                if graph_key not in graphs:
                    graphs.add(graph_key)
                    for node in obj["nodes"]:
                        if not isinstance(node, dict):
                            continue
                        kind = str(node.get("kind") or "unknown")
                        node_kinds[kind] += 1
                        concept = node.get("concept")
                        if concept and kind in {"event", "state"}:
                            predicates[(kind, str(concept))] += 1
                        frame = node.get("temporal_frame")
                        frames[str(frame) if frame else "unknown"] += 1
                        if kind == "referent":
                            mult = node.get("multiplicity")
                            multiplicities[str(mult) if mult else "unknown"] += 1
                        for participant in node.get("participants", []) or []:
                            if isinstance(participant, dict):
                                roles[str(participant.get("role") or "unknown")] += 1
            if any(key in obj for key in ("family_id", "variant_id", "frame_id")):
                traces.add((case_id, stage, str(obj.get("family_id") or "unknown"), str(obj.get("variant_id") or "unknown"), str(obj.get("frame_id") or "unknown"), str(obj.get("outcome") or "unknown")))
            feature_list = obj.get("features")
            if isinstance(feature_list, list):
                for feature in feature_list:
                    if isinstance(feature, dict) and "feature" in feature:
                        feature_values.add((case_id, stage, str(feature.get("feature") or "unknown"), str(feature.get("state") or "unknown"), str(feature.get("value") or "unknown")))
            code = obj.get("code")
            if code:
                diagnostics[(stage, str(code))] += 1
                diagnostics_by_case[source_case_id(case_id)].add(str(code))

    return {
        "calls": calls,
        "false_realizations": false_realizations,
        "stage_outcomes": stage_outcomes,
        "graph_count": len(graphs),
        "node_kinds": node_kinds,
        "predicates": predicates,
        "roles": roles,
        "frames": frames,
        "multiplicities": multiplicities,
        "traces": traces,
        "feature_values": feature_values,
        "diagnostics": diagnostics,
        "diagnostics_by_case": diagnostics_by_case,
        "authored_graph_count": len(authored_graphs),
        "authored_node_kinds": authored_node_kinds,
        "authored_roles": authored_roles,
        "authored_frames": authored_frames,
        "authored_multiplicities": authored_multiplicities,
    }


def render(sha: str, timestamp: str, exit_code: int, case_file: Path, log_file: Path, targets_file: Path) -> str:
    cases = registered_cases(case_file)
    declared = declared_metadata(case_file)
    coverage = {case_id: item["coverage_tags"] for case_id, item in declared.items()}
    log_text = log_file.read_text(encoding="utf-8", errors="replace") if log_file.exists() else ""
    check_observations, extras, package_summaries = parse_checks(log_text)
    inventory, inventory_errors = marked_json(log_text, "matrix_inventory=")
    associations, association_errors = marked_json(log_text, "matrix_sense_association=")
    snapshots, snapshot_errors = marked_json(log_text, "matrix_snapshot=")
    runtime_logs, observation_errors = marked_json(log_text, "observation=")
    direct_entries, direct_result_errors = marked_json_entries(log_text, "result=")
    direct_results = []
    for line_number, line, value in direct_entries:
        case_match = CASE_RE.search(line)
        direct_results.append({"case_id": case_match.group(0) if case_match else f"direct-result-line-{line_number}", "stage": "direct_result", "result": value})
    runtime = collect_runtime(runtime_logs + direct_results)
    inventory_by_id = {str(item.get("case_id")): item for item in inventory if item.get("case_id")}
    boundary_checks = [item for item in extras if item[1].startswith(("TestRefusalAnalysis/", "TestRefusalTranslation/"))]
    extras = [item for item in extras if item not in boundary_checks]
    kinds = ("analysis", "realization", "integration")
    rows: list[tuple[str, list[str]]] = []
    counts: Counter[str] = Counter()
    group_stage_counts: Counter[tuple[str, str, str, str]] = Counter()

    for case_id in cases:
        statuses: list[str] = []
        for kind in kinds:
            seen = check_observations.get((case_id, kind), [])
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
        group = declared.get(case_id, {}).get("primary_group")
        if group:
            meta = declared.get(case_id, {})
            direction = f"{meta.get('from', 'unknown')} → {meta.get('to', 'unknown')}"
            for stage, state in zip(kinds, statuses):
                group_stage_counts[(str(group), direction, stage, state)] += 1

    lines = [
        "# Latest language matrix result", "",
        "> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.", "",
        f"- Source SHA: `{sha}`", f"- Generated (UTC): `{timestamp}`",
        "- Command: `go test ./tests/... -count=1 -v`",
        f"- Command exit code: `{exit_code}` ({'PASS' if exit_code == 0 else 'NONZERO'})",
        f"- Registered records: {len(cases)}; record checks: {len(cases) * len(kinds)}",
        f"- Additional observed checks: {len(extras)}", "",
        "| Case ID | Direction | Analysis | Realization | Integration | Declared coverage tags |", "|---|---|---|---|---|---|",
    ]
    for case_id, values in rows:
        tags = ", ".join(sorted(set(coverage.get(case_id, [])))) or "(none)"
        lines.append(f"| `{case_id}` | {values[0]} | {values[1]} | {values[2]} | {values[3]} | {tags} |")

    boundary_counts: Counter[str] = Counter(state for _, _, state in boundary_checks)
    lines.extend(["", "## Boundary refusal checks", "", "These malformed or unsupported inputs are separate from normal registered translation records and do not contribute to record PASS totals.", "", f"- Total: {len(boundary_checks)}; " + ", ".join(f"{state}: {boundary_counts[state]}" for state in sorted(boundary_counts)), "", "| Category | Package | Check | Status |", "|---|---|---|---|"])
    for package, test_name, state in boundary_checks:
        category = "native analysis refusal" if test_name.startswith("TestRefusalAnalysis/") else "translation refusal"
        lines.append(f"| {category} | `{package}` | `{test_name}` | {state} |")

    lines.extend(["", "## Additional observed checks", "", "| Package | Check | Status |", "|---|---|---|"])
    for package, test_name, state in extras:
        lines.append(f"| `{package}` | `{test_name}` | {state} |")

    lines.extend(["", "## Package summaries", "", "Packages with no observed test run are reported as infrastructure/build failures when Go returned `FAIL`; summaries do not replace per-check outcomes.", "", "| Package | Go summary | Observed test names |", "|---|---|---|"])
    for package, summary, run_count in package_summaries:
        state = "INFRA (no tests ran)" if summary == "FAIL" and run_count == 0 else summary
        lines.append(f"| `{package}` | {state} | {run_count} |")

    tag_counts: Counter[str] = Counter(tag for values in coverage.values() for tag in values)
    lines.extend(["", "## Declared coverage", "", "Declared tags are authored metadata; they do not prove construction execution or correctness.", ""])
    lines.append(" | ".join(f"{tag}: {tag_counts[tag]} records" for tag in sorted(tag_counts)) or "No coverage tags declared.")
    lines.extend(["", "## Pair and group balance", "", "Pair/group counts describe sample distribution. Coverage targets are planning goals, not correctness gates.", ""])
    pair_groups: dict[str, set[str]] = defaultdict(set)
    group_pairs: dict[str, set[str]] = defaultdict(set)
    source_meta = dict(declared)
    source_meta.update({case_id: {**source_meta.get(case_id, {}), **{key: item.get(key) for key in ("pair_id", "primary_group", "coverage_tags")}} for case_id, item in inventory_by_id.items()})
    for case_id in cases:
        meta = source_meta.get(case_id, {})
        pair_id, group = meta.get("pair_id"), meta.get("primary_group")
        if pair_id and group:
            pair_groups[str(pair_id)].add(str(group))
            group_pairs[str(group)].add(str(pair_id))
    try:
        targets = json.loads(targets_file.read_text(encoding="utf-8")) if targets_file.exists() else {}
    except (OSError, json.JSONDecodeError):
        targets = {}
    goal_groups = targets.get("groups", {}) if isinstance(targets, dict) else {}
    lines.extend([f"- Declared unique pairs: {len(pair_groups)}; target total: {targets.get('target_pairs', 'not supplied') if isinstance(targets, dict) else 'not supplied'}", "", "| Group | Observed pairs | Planning target |", "|---|---:|---:|"])
    for group in sorted(set(group_pairs) | set(goal_groups)):
        lines.append(f"| {group} | {len(group_pairs.get(group, set()))} | {goal_groups.get(group, 'not supplied')} |")

    lines.extend(["", "### Direction outcomes by group and stage", "", "| Group | Direction | Stage | Check outcome | Records |", "|---|---|---|---|---:|"])
    for (group, direction, stage, outcome), count in sorted(group_stage_counts.items()):
        lines.append(f"| {group} | {direction} | {stage} | {outcome} | {count} |")

    source_sha_counts = Counter(str(item.get("source_sha", "unknown")) for item in inventory)
    semantic_predicates = Counter(str((item.get("source_semantics") or {}).get("predicate") or "(unspecified)") for item in inventory)
    semantic_roles = Counter(role for item in inventory for role in strings((item.get("source_semantics") or {}).get("roles")))
    allowed_id_count = sum(len(strings(item.get("allowed_meaning_ids"))) for item in inventory)
    required_id_count = sum(len(strings(item.get("required_meaning_ids"))) for item in inventory)
    lines.extend(["", "## Inventory and pack metadata", "", f"- Inventory records parsed: {len(inventory)} / {len(cases)} declared directions.", f"- Snapshot identity records: {len(snapshots)}.", f"- Sense catalogue associations parsed: {len(associations)} (not morphology execution coverage).", f"- Declared Meaning ID references: {allowed_id_count} allowed; {required_id_count} required.", "- Inventory fields preserve declared tags, source semantics, allowed/required Meaning IDs, and owner status separately from runtime observations."])
    for source_sha, count in sorted(source_sha_counts.items()):
        lines.append(f"- Inventory source SHA `{source_sha}`: {count} directions.")
    lines.extend(["", "### Declared source requirements", "", "Aggregates below summarize authored source requirements; they are not observed analyzer output.", "", "| Required predicate | Directions |", "|---|---:|"])
    for predicate, count in sorted(semantic_predicates.items()):
        lines.append(f"| `{predicate}` | {count} |")
    lines.extend(["", "Declared role mentions: " + (", ".join(f"`{role}` {count}" for role, count in sorted(semantic_roles.items())) or "none recorded") + "."])
    if snapshots:
        for snapshot in snapshots:
            lines.append(f"- Pack identity: `{json.dumps(snapshot.get('pack_identity', {}), sort_keys=True, ensure_ascii=False)}`; inventory source SHA: `{snapshot.get('source_sha', 'unknown')}`.")
    association_counts: Counter[tuple[str, str, str, str]] = Counter()
    for item in associations:
        lexeme = item.get("lexeme") if isinstance(item.get("lexeme"), dict) else {}
        association_counts[(str(item.get("association_status", "unknown")), str(lexeme.get("language", "unknown")), str(lexeme.get("pos", "unknown")), str(lexeme.get("paradigm_id", "unknown")))] += 1
    lines.extend(["", "Catalogue sense associations are not proof that a paradigm form or morphology procedure ran. Aggregates below describe declared pack associations only.", "", "| Association | Language | POS | Paradigm | Senses |", "|---|---|---|---|---:|"])
    for (state, language, pos, paradigm), count in sorted(association_counts.items()):
        lines.append(f"| {state} | {language} | {pos} | {paradigm} | {count} |")
    for label, errors in (("inventory", inventory_errors), ("association", association_errors), ("snapshot", snapshot_errors), ("observation", observation_errors), ("direct result", direct_result_errors)):
        if errors:
            lines.extend(["", f"{label.title()} JSON parse issues: {len(errors)} (records remain visible through source fallback)."])
            lines.extend(f"- {error}" for error in errors[:20])

    lines.extend(["", "## Runtime adapter observations", "", "Runtime calls are separate from check outcomes. A false adapter marker is runtime NOT RUN even when its corresponding GAPI check has a recorded result. Check NOT RUN remains in the record table above.", "", "| Stage | Adapter executed | Calls |", "|---|---|---:|"])
    for (stage, executed), count in sorted(runtime["calls"].items()):
        lines.append(f"| {stage} | {executed} | {count} |")
    false_call_count = sum(len(ids) for ids in runtime["false_realizations"].values())
    lines.extend(["", "### Independent realization runtime NOT RUN", "", f"- Calls not executed: {false_call_count}", "", "| Reason | Affected record IDs | Unexecuted calls |", "|---|---|---:|"])
    for (reason, _stage), ids in sorted(runtime["false_realizations"].items()):
        lines.append(f"| {reason.replace('|', '&#124;')} | {', '.join(f'`{value}`' for value in sorted(set(ids)))} | {len(ids)} |")
    lines.extend(["", "### Separate observed source probes", ""])
    for stage, executed in sorted(runtime["calls"]):
        if stage in {"source_gap_probe", "context_source_analysis"} or stage not in {"native_analysis", "independent_realization", "integration"}:
            lines.append(f"- {stage}, adapter executed `{executed}`: {runtime['calls'][(stage, executed)]}")

    lines.extend(["", "## Direct-result JSON observations", "", f"- Explicit `result=` objects decoded with `JSONDecoder.raw_decode`: {len(direct_entries)}. They remain separate from adapter-call counts.", "- Parse issues: " + str(len(direct_result_errors)), "", "## Observed result outcomes", "", "Observed top-level outcomes by runtime stage; these do not replace check statuses.", "", "| Stage | Result outcome | Count |", "|---|---|---:|"])
    for (stage, outcome), count in sorted(runtime["stage_outcomes"].items()):
        lines.append(f"| {stage} | {outcome} | {count} |")

    lines.extend(["", "## Authored independent realization Meaning inputs", "", "These are the authored graphs supplied to the realizer, distinct from source analyzer and runtime result graphs. Repeated identical inputs are deduplicated by case/stage/MeaningID.", "", f"- Unique authored Meaning inputs observed: {runtime['authored_graph_count']}", "", "| Authored graph property | Value | Count |", "|---|---|---:|"])
    for kind, count in sorted(runtime["authored_node_kinds"].items()):
        lines.append(f"| node kind | {kind} | {count} |")
    for role, count in sorted(runtime["authored_roles"].items()):
        lines.append(f"| participant role | {role} | {count} |")
    for frame, count in sorted(runtime["authored_frames"].items()):
        lines.append(f"| temporal_frame | {frame} | {count} |")
    for multiplicity, count in sorted(runtime["authored_multiplicities"].items()):
        lines.append(f"| referent multiplicity | {multiplicity} | {count} |")

    lines.extend(["", "## Observed source/result Meaning graphs", "", "Authored `meaning_input` supplied to independent realization is excluded from observed graphs below. Graph summaries deduplicate identical graphs within a case and stage. Missing temporal frames or multiplicities are counted as unknown.", "", f"- Unique observed graph snapshots: {runtime['graph_count']}", "", "| Graph property | Value | Count |", "|---|---|---:|"])
    for kind, count in sorted(runtime["node_kinds"].items()):
        lines.append(f"| node kind | {kind} | {count} |")
    for (kind, concept), count in sorted(runtime["predicates"].items()):
        lines.append(f"| predicate ({kind}) | `{concept}` | {count} |")
    for role, count in sorted(runtime["roles"].items()):
        lines.append(f"| participant role | {role} | {count} |")
    for frame, count in sorted(runtime["frames"].items()):
        lines.append(f"| temporal_frame | {frame} | {count} |")
    for multiplicity, count in sorted(runtime["multiplicities"].items()):
        lines.append(f"| referent multiplicity | {multiplicity} | {count} |")

    lines.extend(["", "## Observed trace and feature associations", "", "Trace rows are deduplicated by case/stage/family/variant/frame/outcome. Feature values are observed only when present in emitted result traces.", "", "| Family ID | Variant ID | Frame ID | Outcome | Trace records |", "|---|---|---|---|---:|"])
    trace_counts = Counter((family, variant, frame, outcome) for _case, _stage, family, variant, frame, outcome in runtime["traces"])
    for (family, variant, frame, outcome), count in sorted(trace_counts.items()):
        lines.append(f"| `{family}` | `{variant}` | `{frame}` | {outcome} | {count} |")
    lines.extend(["", "### Observed features", "", "| Feature | State | Value | Deduplicated observations |", "|---|---|---|---:|"])
    feature_counts = Counter((feature, state, value) for _case, _stage, feature, state, value in runtime["feature_values"])
    for (feature, state, value), count in sorted(feature_counts.items()):
        lines.append(f"| `{feature}` | {state} | {value} | {count} |")

    gap_groups: dict[tuple[str, str, str], set[str]] = defaultdict(set)
    for case_id in cases:
        metadata = inventory_by_id.get(case_id, {})
        semantics = metadata.get("source_semantics") or {}
        gap = str(semantics.get("gap") or "") if isinstance(semantics, dict) else ""
        if gap:
            predicate = str(semantics.get("predicate") or "unspecified")
            codes = sorted(runtime["diagnostics_by_case"].get(case_id, set()))
            gap_groups[(predicate, gap, ", ".join(codes) or "none observed")].add(case_id)
    lines.extend(["", "## Representation gap diagnosis", "", "Gaps are grouped by required predicate and authored missing-representation reason. Diagnostic codes shown are observed codes from the raw log; absence of a code is reported explicitly.", "", "| Layer | Required predicate | Missing representation / reason | Observed diagnostic codes | Records |", "|---|---|---|---|---|"])
    for (predicate, reason, codes), ids in sorted(gap_groups.items()):
        lines.append(f"| Meaning representation | `{predicate}` | {reason.replace('|', '&#124;')} | {codes} | {', '.join(f'`{value}`' for value in sorted(ids))} |")
    lines.extend(["", "### Diagnostic code counts", "", "| Stage | Diagnostic code | Occurrences |", "|---|---|---:|"])
    for (stage, code), count in sorted(runtime["diagnostics"].items()):
        lines.append(f"| {stage} | `{code}` | {count} |")

    lines.extend(["", "## Counts", "", " | ".join(f"{name}: {counts[name]}" for name in ("PASS", "FAIL", "GAP_ANALYZE", "GAP_REALIZE", "GAP_I", "SPEC_PENDING", "INFRA", "NOT RUN")), "", "`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching record check result was observed. Runtime adapter NOT RUN markers are counted separately above. Additional checks retain their own statuses. Consult the full raw log artifact for every individual observation.", ""])
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sha", required=True, help="source commit SHA")
    parser.add_argument("--exit-code", required=True, type=int, help="go test exit code")
    parser.add_argument("--timestamp", help="UTC timestamp; defaults to current UTC time")
    parser.add_argument("--cases", type=Path, default=Path("tests/translation/cases.go"))
    parser.add_argument("--log", required=True, type=Path, help="complete verbose go test log")
    parser.add_argument("--targets", type=Path, default=Path("tests/coverage-targets.json"), help="optional planning goals")
    args = parser.parse_args()
    timestamp = args.timestamp or datetime.now(timezone.utc).replace(microsecond=0).isoformat()
    try:
        sys.stdout.write(render(args.sha, timestamp, args.exit_code, args.cases, args.log, args.targets))
    except (OSError, ValueError) as exc:
        parser.error(str(exc))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

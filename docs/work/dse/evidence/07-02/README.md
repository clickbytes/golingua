# Evidence #395 — source composition

This directory records the final frozen runtime observation for the bounded #395 source-composition profile. It is evidence for reproducibility and technical behavior, not a quality score, broad language coverage claim, or owner approval.

## Frozen source and checks

- Binary: SHA-256 `c0499634e98b29259069e10949bb2ca7b0e0fc1fb40e4bad5a34f8a6d7659fa7` (`golingua-395-final`).
- Source hashes: [`reports/source-hashes.json`](reports/source-hashes.json), covering 11 Go files and 2 schemas. The only documentation-phase source change is two schema `$comment` clarifications; schema validation behavior is unchanged.
- [`reports/technical-observation.json`](reports/technical-observation.json): `gofmt -l` empty; `go vet ./...`, `go build ./cmd/golingua`, `--help`, and `--version` exit 0 for the frozen build.
- [`reports/schema-observation.json`](reports/schema-observation.json): 25 selected artifacts validated with no schema errors.

## Literal runtime records

[`outputs/main/observations.jsonl`](outputs/main/observations.jsonl) preserves 36 literal CLI calls and their raw stdout files: 12 PL/EN analyses, 12 independent checks of extracted canonical documents, six full-context comparisons, and six independent equivalence-certificate verifications. All 36 exited 0. [`reports/semantic-main.json`](reports/semantic-main.json) records the observed counts and properties.

[`outputs/extra/observations.jsonl`](outputs/extra/observations.jsonl) preserves 14 additional calls and raw outputs. Their exit codes were 0:3, 7:4, 4:2, 3:2, and 2:3. [`reports/semantic-extra.json`](reports/semantic-extra.json) summarizes renamed-symbol and feature-port controls, binding/operator refusals, resource limits, attribute conflicts, ambiguous references, and unknown/unmapped source cases. These are bounded controls, not aggregate accuracy metrics.

The JSONL ledgers preserve the literal `/tmp/golingua-dse/395-root-observations/` and `395-extra-observations/` paths used at execution time. Their stdout basenames are copied beside each ledger under `outputs/main/` and `outputs/extra/`. Main candidate documents and certificate inputs are copied into `outputs/main/`; changed control inputs whose original path was `/tmp/golingua-dse/395-extra-inputs/` are copied with the same basename under [`inputs/`](inputs/). Product package, plan, context, metadata, and source cases are linked below.

The combined 50 calls had 39 exit-0, 4 incomplete, 2 resource-limit, 2 unsupported, and 3 invalid results. Exit counts describe this selected probe set only.

The small reproducible documentation flow also ran on the same binary. Its scratch runtime config, profile, extracted documents, certificate and six outputs are under [`inputs/`](inputs/); [`reports/doc-flow-observation.json`](reports/doc-flow-observation.json) records output sizes, hashes, statuses and exits. Snapshot apply, EN and PL analysis, standalone check, full-context comparison and independent verification all exited 0. The certificate was extracted from `.result.certificate`.

## Inputs and provenance

The raw output directories also retain the candidate documents, profiles and certificates used for the 50 calls. Additional changed-input controls are under [`inputs/`](inputs/). The authored language packages, source plans, metadata, contexts and 12 source cases remain in [`data/semantic/examples/source-composition/`](../../../../../data/semantic/examples/source-composition/) and [`data/semantic/languages/`](../../../../../data/semantic/languages/); they are the actual inputs, not generated golds.

The final primary run included six full-context PL/EN pairs. It preserved request references, `TemporalFrame`, `Multiplicity`, context conditions and `Not` scope; it did not strip attributes or contexts to obtain equivalence. Each pair compared as `proved_equivalent` under the selected empty rewrite profile, and each certificate independently verified. This establishes equivalence only within that profile. All 12 source analyses returned `goal_specs: []`; the goal-fragment evidence-check path was **NOT RUN**. Whole-sentence scope was checked as `type_scope_context_only`; there is no world-truth, fact-existence, completeness, or general language-acceptance claim.

The two-dog case uses distinct B and C references with explicit caller-provided definite-reference conditions. Those conditions are context assumptions, not an inferred Polish definiteness rule. Mention identity uses the actual parser candidate IDs and corroborating resolution evidence. Role-reference fallback is caller-supplied and is not identity proof. Same-class mentions are not merged.

The pilot is limited to the selected present seeing/hearing and single negation routes. Unmapped/unknown forms, unresolved references, unauthorized mappings, unsupported operators, and resource caps remain visible as refusals or incomplete outcomes. The `reference_predicate` addition is an explicit unary reference mapping; it does not add a fake valency or alter existing frame bindings. Source provenance and pending-audit labels are retained; DSE review does not promote them. Agent assessment is bounded to this pilot. Owner language acceptance: **NOT EVALUATED**. PR and main CI receipts are pending.

Koordynator niezależnie uruchomił cały literalny blok CLI z dokumentacji: snapshot, analizy EN/PL, odczyt IR, pełne porównanie i weryfikacja certyfikatu — wszystkie 6 wywołań z kodem 0. [Obserwacja](root-docflow/observation.json), [dokładne wywołania](root-docflow/commands.jsonl).

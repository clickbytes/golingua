# DSE-02-01 metadata registry evidence

This is a bounded mathematical CLI observation record, not a language corpus, test runner or automated acceptance gate. Expected typing, admission and refusal requirements were specified independently of output; owner language acceptance remains `NOT_EVALUATED`.

## Inputs and replay

- [Neutral package](../../../../../data/semantic/examples/metadata/neutral.json): one nominal sort, approved primitive, approved closed two-hop alias, pending opaque record, structural constraints, provenance and no axioms. [Expression](../../../../../data/semantic/examples/metadata/expression.json) uses its primitive without inline signatures.
- [Post-build package](../../../../../data/semantic/examples/metadata/after-build.json) and [expression](../../../../../data/semantic/examples/metadata/after-build-expression.json) were created after the final executable build. Their fresh symbol, role and sort IDs derive from that executable's digest; no build occurred between their creation and successful inspect/check.
- [Control inputs](inputs/) retain label removal, consistent renaming, invalid structures, alias/dependency cycles, unavailable records and dependency-pin experiments. They are explicit input data, not a runner or runtime source code.
- [Literal observations](observations.jsonl) contain argv, cwd, stdout/stderr, exit codes and input/binary hashes. [Environment](environment.json) pins actual base SHA plus exact working-tree source fingerprints, compiler, module hashes, binary modification time and package creation order. [Invariant observations](invariants.json) record comparisons made from those actual results.

## Observed result and limits

The unchanged executable inspected, validated and checked the neutral and post-build packages. Structural checking returned `Prop`; it did not execute the relation, synthesize a term, prove a world fact or grant language approval. Package creation time followed binary build time, and the executable SHA-256 before and after the experiments is identical.

Removing labels/descriptions leaves formal denotation identity and type checking unchanged while changing resource identity. Renaming domain symbol/role/sort IDs consistently preserves the corresponding checked type and canonical bindings; literal hash equality is not assumed. A parent alias's formal ID is unchanged by an imported dependency's label edit (with corrected resource pin), but changes when that dependency's alias body changes under the same ID/signature.

Invalid fields, versions, record/role duplication, wrong structural constraints, missing sort/edge/dependency references, provenance, inline/external conflicts and alias cycles were refused. Axioms and unknown constraint kinds remain unsupported. Pending/opaque/alias records can be inspected but cannot masquerade as operational primitives. Oversize input is resource_limit. Failure of one package in an explicit set emits no partial package list. The cyclic dependency input also has invalid resource pins and stops first at fingerprint mismatch; the separate alias-cycle input yields ALIAS_CYCLE. The first missing-sort control also contradicts its port_type constraint; the isolated missing-sort control preserves that constraint and yields UNKNOWN_SORT. These diagnostics are retained as observed rather than relabeled.

The first worker smoke of the unchanged two-port alias exposed ALIAS_RESULT_TYPE. Its actual inferred type was correct; the expected-type loop pointed Arrow.To at the variable it overwrote. After the allowed writer correction budget, the coordinator integrated the writer's exact bounded fix (`previous := expected`, then `To: &previous`), and the complete original example validates. No alias requirement or semantic expectation was removed to fit output. The final binary/evidence includes that fix.

`gofmt`, `go vet ./...`, build, help/version, source reviews and the recorded CLI controls completed. CI PR/main and merge receipts are recorded on issue #380 only after they occur. SourceRef and review flags are attribution/admission data; they are not authenticated proof of world truth or owner acceptance. Synthesis, alias execution, fact evaluation and PL/EN integration remain outside this leaf.

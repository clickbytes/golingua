# Semantic algebra and DSE

This directory holds the DSE contracts. The runtime implements bounded portions of those contracts; actual capabilities and evidence are described in the [DSE results](../../work/dse/results.md) and [acceptance](../../work/dse/acceptance.md).

For the issue record and work procedure, see the [DSE work index](../../work/dse/README.md), [issue map](../../work/dse/issue-map.json), and [execution runbook](../../work/dse/execution.md).

## Contracts

- [Formal contract](formal-contract.md): typed expression forms, identity and scope, generic operators, bounded synthesis, independent verification claims, and the initial positive facts-only evaluation profile.
- [Meaning v1 compatibility](compatibility.md): explicit route selection, authority boundaries, partial conversion rules, context and identity preservation, and refusal behavior for lossy mappings.
- [Finite explicit composition](composition.md): the bounded algebra construction command implemented in #381, its operator limits and neutral examples; it does not implement the broader DSE synthesis and evaluation plans above.
- [Bounded normalization](normalization.md): the explicitly licensed rewrite command, strict certificate verification, finite budgets, and neutral #382 examples; it does not evaluate facts or synthesize proofs from goals.
- [Runtime artifact storage](runtime-storage.md): the bounded #383 export, structural inspection, and metadata-bound replay path, with neutral storage requests; artifacts preserve checked syntax and receipts, not facts or proofs.
- [Request-scoped n-ary facts](facts.md): the #384 bounded explicit-evidence view, strict request format, UTC selection, and neutral positive/conflict examples; it reports evidence and does not infer world truth.

The formal contract's initial executable profile allows structural alpha/beta reduction and positive matching against supplied facts. Its domain-axiom and rule allowlist is empty; absence of a matching fact remains `unknown`. The compatibility contract keeps existing Meaning v1 routes authoritative unless a future route explicitly opts into DSE. Neither contract introduces an API, schema migration, fallback, or language binding.

## Issue hierarchy and planned order

DSE is tracked under [epic #367](https://github.com/clickbytes/golingua/issues/367). Its chapters are #368–#376, with executable leaves #377–#400. The final acceptance leaf is [#400](https://github.com/clickbytes/golingua/issues/400); the [issue map](../../work/dse/issue-map.json) records verified predecessor results.

The epic's recommended serial leaf order is:

1. #377–#378: formal contract and execution runbook.
2. #379–#385: typed IR, registry, generic operators, rewrite kernel, runtime DAG, facts/context, catalog triage.
3. #386–#390: goal construction, bounded synthesis, verification, query evaluation, equivalence proofs.
4. #391–#393: scope, references, context, and semantic attributes.
5. #394–#398: lexical goals, source composition, target coverage, runtime integration, and PL/EN slices.
6. #399–#400: data-independence review and final acceptance.

The list describes the serial order, not a claim of general linguistic completeness.

## Status and evidence

Typed IR, explicit metadata and facts, bounded construction and search, verification, positive-fact evaluation, and limited PL/EN integration are implemented. See [final observations](../../work/dse/evidence/09-02/README.md) and the [independence audit](../../work/dse/independence-audit.md) for actual results and limits. Representation of an operator does not imply it can be evaluated. Formal review, technical verification, agent assessment, and owner language acceptance are distinct; owner acceptance remains **NOT EVALUATED**.

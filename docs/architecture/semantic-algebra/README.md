# Semantic algebra and DSE

This directory holds the design contracts for the proposed Dynamic Semantic Expressions (DSE) work. The RFCs define future behavior; they do not describe capabilities already present in the runtime.

For the issue record and work procedure, see the [DSE work index](../../work/dse/README.md), [issue map](../../work/dse/issue-map.json), and [execution runbook](../../work/dse/execution.md).

## Contracts

- [Formal contract](formal-contract.md): typed expression forms, identity and scope, generic operators, bounded synthesis, independent verification claims, and the initial positive facts-only evaluation profile.
- [Meaning v1 compatibility](compatibility.md): explicit route selection, authority boundaries, partial conversion rules, context and identity preservation, and refusal behavior for lossy mappings.
- [Finite explicit composition](composition.md): the bounded algebra construction command implemented in #381, its operator limits and neutral examples; it does not implement the broader DSE synthesis and evaluation plans above.

The formal contract's initial executable profile allows structural alpha/beta reduction and positive matching against supplied facts. Its domain-axiom and rule allowlist is empty; absence of a matching fact remains `unknown`. The compatibility contract keeps existing Meaning v1 routes authoritative unless a future route explicitly opts into DSE. Neither contract introduces an API, schema migration, fallback, or language binding.

## Issue hierarchy and planned order

DSE is tracked under [epic #367](https://github.com/clickbytes/golingua/issues/367). Its chapters are #368–#376, with executable leaves #377–#400. The current leaf, [#377](https://github.com/clickbytes/golingua/issues/377), establishes these RFCs; [#378](https://github.com/clickbytes/golingua/issues/378) is the execution runbook. Later planned chapters cover metadata and facts (#369), typed IR and operators (#370), synthesis (#371), evaluation and evidence (#372), scope and context (#373), lexical analysis (#374), realization and integration (#375), and domain-independence/acceptance (#376).

The epic's recommended serial leaf order is:

1. #377–#378: formal contract and execution runbook.
2. #379–#385: typed IR, registry, generic operators, rewrite kernel, runtime DAG, facts/context, catalog triage.
3. #386–#390: goal construction, bounded synthesis, verification, query evaluation, equivalence proofs.
4. #391–#393: scope, references, context, and semantic attributes.
5. #394–#398: lexical goals, source composition, target coverage, runtime integration, and PL/EN slices.
6. #399–#400: data-independence review and final acceptance.

This is planned work, not evidence that the leaves have started or that their proposed paths, commands, schemas, or runtime behavior exist. Start each leaf only after checking its current issue, dependencies, main SHA, and repository instructions.

## Status and evidence

The current work is documentation-only. DSE implementation/runtime status is **NOT RUN**; owner language acceptance is **NOT_EVALUATED**. Formal correctness review, implementation status, agent assessment, and owner acceptance are separate outcomes. A document review does not count as a runtime pass.

Future implementation must proceed through the serial issue hierarchy and report observable CLI results and limits for its active leaf. The initial facts-only profile does not authorize language-specific assumptions, domain axioms, or a closed-world inference. See the RFCs for detailed refusal and evidence requirements.
